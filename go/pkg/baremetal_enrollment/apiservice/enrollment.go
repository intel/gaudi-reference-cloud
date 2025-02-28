// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var (
	maxElapsedTime = 10 * time.Minute
)

var errEnrollmentPod = errors.New("enrollment pod failed")

type K8SClient interface {
	GetK8SClientset(ctx context.Context, region string, availabilityZone string,
		objects ...kuberuntime.Object) (kubernetes.Interface, error)
}

type JobSpecBuilder interface {
	CreateEnrollmentJobSpec(deviceInfo config.BMaaSEnrollmentData) *batchv1.Job
	CreateDisenrollmentJobSpec(deviceInfo config.BMaaSEnrollmentData) *batchv1.Job
}

type ClusterRegion interface {
	GetClusterRegion(deviceInfo config.BMaaSEnrollmentData) (string, error)
}

type BMaaSEnrollment struct {
	K8SClient      K8SClient
	JobSpecBuilder JobSpecBuilder
	Region         ClusterRegion
	ClientSet      kubernetes.Interface
}

func NewBMaaSEnrollment(client K8SClient, jobConfig JobSpecBuilder,
	region ClusterRegion) (BMaaSEnrollment, error) {

	enrollment := BMaaSEnrollment{
		K8SClient:      client,
		JobSpecBuilder: jobConfig,
		Region:         region,
		ClientSet:      nil,
	}
	return enrollment, nil
}

func (e *BMaaSEnrollment) AddRoutes(router *gin.Engine, username string, password string) error {
	// logger init
	log.SetDefaultLogger()
	v1Routes := router.Group("/api/v1")
	v1Routes.POST("/enroll", gin.BasicAuth((gin.Accounts{username: password})), e.handle)
	return nil
}

func (e *BMaaSEnrollment) handle(ctx *gin.Context) {
	log := log.FromContext(ctx).WithName("apiservice.enrollment.handle")
	var deviceInfo config.BMaaSEnrollmentData
	if err := ctx.ShouldBindJSON(&deviceInfo); err != nil {
		log.Error(err, "failed to bind data in the enrollment request", "device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone, "enrollment_status", deviceInfo.EnrollmentStatus)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	region, err := e.Region.GetClusterRegion(deviceInfo)
	if err != nil || region == "" {
		log.Error(err, "failed to get the device region", "device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Info("device region", "device", deviceInfo.DeviceName, "region", region)

	// get the K8S clientset
	// Check for fakeclientset used for testing.
	if e.ClientSet == nil {
		clientSet, err := e.K8SClient.GetK8SClientset(ctx, region, deviceInfo.AvailabilityZone)
		// return internal server error if failed to get the clientset
		if err != nil {
			log.Error(err, "failed to get the kubeconfig clientset",
				"device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		e.ClientSet = clientSet
	}

	switch status := dcim.BMEnrollmentStatus(deviceInfo.EnrollmentStatus); status {
	case dcim.BMEnroll:
		log.Info("new enrollment request", "device", deviceInfo.DeviceName)
		e.enrollBMaaSNode(ctx, deviceInfo, region)
	case dcim.BMDisenroll:
		log.Info("new disenrollment request", "device", deviceInfo)
		e.disenrollBMaaSNode(ctx, deviceInfo, region)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported request for BM status: %s", status)})
		return
	}
}

func (e *BMaaSEnrollment) enrollBMaaSNode(ctx *gin.Context, deviceInfo config.BMaaSEnrollmentData, region string) {
	log := log.FromContext(ctx).WithName("apiservice.enrollment.enrollBMaaSNode")

	// Check if a job is already running for the device
	// search with the label device-info=bmaas-enrollment-<DeviceName>
	jobList, err := e.getEnrollmentJob(ctx, deviceInfo)
	if err != nil {
		log.Error(err, "failed to list enrollment jobs", "device", deviceInfo.DeviceName,
			"site", deviceInfo.AvailabilityZone)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		e.ClientSet = nil
		return
	}

	if len(jobList.Items) == 1 {
		// check for failed or completed jobs
		if (jobList.Items[0].Status.Failed > 0) || (jobList.Items[0].Status.Succeeded > 0) {
			err = e.deleteEnrollmentJob(ctx, deviceInfo)
			if err != nil {
				log.Error(err, "failed to delete enrollment job", "device", deviceInfo.DeviceName,
					"site", deviceInfo.AvailabilityZone)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				e.ClientSet = nil
				return
			}
			time.Sleep(5 * time.Second)
		} else {
			log.Info("an enrollment job is already running", "device", deviceInfo.DeviceName,
				"site", deviceInfo.AvailabilityZone, "region", region)
			ctx.Status(http.StatusConflict)
			e.ClientSet = nil
			return
		}

	}
	// get the Job spec for the enrollment Job
	enrollmentJobSpec := e.getJobSpec(deviceInfo)

	// create K8s job with the enrollment tasks
	err = e.launchJob(ctx, enrollmentJobSpec)
	if err != nil {
		log.Error(err, "failed to create enrollment job",
			"device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		e.ClientSet = nil
		return
	}
	ctx.Status(http.StatusCreated)
	log.Info("created enrollment job successfully", "device", deviceInfo.DeviceName,
		"site", deviceInfo.AvailabilityZone, "region", region)

	// verify enrollment job and pod
	go verifyEnrollmentJob(ctx, e.ClientSet, e.K8SClient, deviceInfo, region)

	e.ClientSet = nil
}

func (e *BMaaSEnrollment) getJobSpec(deviceInfo config.BMaaSEnrollmentData) *batchv1.Job {
	return e.JobSpecBuilder.CreateEnrollmentJobSpec(deviceInfo)
}

func (e *BMaaSEnrollment) getEnrollmentJob(ctx *gin.Context,
	deviceInfo config.BMaaSEnrollmentData) (batchv1.JobList, error) {

	label := fmt.Sprintf("%s=%s-%s", config.DeviceInfoLabelKey, config.EnrollmentJobNamePrefix, deviceInfo.DeviceName)
	jobs := e.ClientSet.BatchV1().Jobs(config.EnrollmentNamespace)
	jobList, err := jobs.List(ctx, metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return batchv1.JobList{}, err
	}
	return *jobList, nil
}

func (e *BMaaSEnrollment) launchJob(ctx *gin.Context, enrollmentJobSpec *batchv1.Job) error {
	jobs := e.ClientSet.BatchV1().Jobs(config.EnrollmentNamespace)
	_, err := jobs.Create(ctx, enrollmentJobSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (e *BMaaSEnrollment) deleteJob(ctx *gin.Context, jobName string) error {
	jobs := e.ClientSet.BatchV1().Jobs(config.EnrollmentNamespace)
	// delete child pods when deleting the job
	propagationPoloicy := metav1.DeletePropagationBackground
	err := jobs.Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPoloicy})
	if err != nil {
		return err
	}
	return nil
}

func (e *BMaaSEnrollment) deleteEnrollmentJob(ctx *gin.Context, deviceInfo config.BMaaSEnrollmentData) error {
	jobName := fmt.Sprintf("%s-%s", config.EnrollmentJobNamePrefix, deviceInfo.DeviceName)
	return e.deleteJob(ctx, jobName)
}

func verifyEnrollmentJob(ctx *gin.Context, clientSet kubernetes.Interface,
	k8sClient K8SClient, deviceInfo config.BMaaSEnrollmentData, region string) {
	enrollmentError := false
	activeJobs := false
	log := log.FromContext(ctx).WithName("apiservice.enrollment.verifyEnrollmentJob")
	label := fmt.Sprintf("%s=%s-%s", config.DeviceInfoLabelKey, config.EnrollmentJobNamePrefix, deviceInfo.DeviceName)
	backoffTimer := newExponentialBackoff()
	jobs := clientSet.BatchV1().Jobs(config.EnrollmentNamespace)
	for start := time.Now(); time.Since(start) < maxElapsedTime; {
		jobList, err := jobs.List(ctx, metav1.ListOptions{LabelSelector: label})
		if err != nil || len(jobList.Items) < 1 {
			log.Error(err, "failed to list the enrollment job. retrying...", "device",
				deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone, "region", region)
			enrollmentError = true
			time.Sleep(backoffTimer.NextBackOff())
			continue
		}

		enrollmentError = false
		for _, job := range jobList.Items {
			if job.Status.Active > 0 {
				activeJobs = true
				log.Info("enrollment job started running", "device", deviceInfo.DeviceName,
					"site", deviceInfo.AvailabilityZone, "region", region)
				error := verifyEnrollmentPod(ctx, clientSet, deviceInfo, region)
				if error != nil {
					switch {
					// log error and exit if enrollment pod is failed
					// failed enrollment pod updates the netbox with error message
					case errors.Is(error, errEnrollmentPod):
						log.Error(err, "enrollment pod failed", "device", deviceInfo.DeviceName, "site",
							deviceInfo.AvailabilityZone, "region", region)
						runtime.Goexit()
					default:
						log.Error(err, "failed to verify the the status of enrollment pod",
							"device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone,
							"region", region)
						enrollmentError = true
					}
				} else {
					log.Info("enrollment pod started running", "device", deviceInfo.DeviceName,
						"site", deviceInfo.AvailabilityZone, "region", region)
					enrollmentError = false
				}
			}
		}
		if !activeJobs {
			log.Info("enrollment job doesn't have any active pods. retrying...", "device",
				deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone, "region", region)
			enrollmentError = true
			time.Sleep(backoffTimer.NextBackOff())
			continue
		} else {
			break
		}
	}
	// update netbox job failure
	if enrollmentError {
		func(i interface{}) {
			switch v := i.(type) {
			case Client:
				log.Info("updating netbox about the enrollment job/pod failure", "device",
					deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone, "region", region)
				comment := "This system has failed to enroll into BMaaS"
				if err := v.netBox.UpdateDeviceCustomFields(ctx, deviceInfo.DeviceName, deviceInfo.DeviceID, &dcim.DeviceCustomFields{
					BMEnrollmentStatus:  dcim.BMEnrollmentFailed,
					BMEnrollmentComment: comment,
				}); err != nil {
					log.Error(err, "failed to update device status to 'failed' in netbox", "device",
						deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone, "region", region)
				}
			}
		}(k8sClient)
	}
	runtime.Goexit()
}

func verifyEnrollmentPod(ctx *gin.Context, clientSet kubernetes.Interface,
	deviceInfo config.BMaaSEnrollmentData, region string) error {
	log := log.FromContext(ctx).WithName("apiservice.enrollment.verifyEnrollmentPod")
	label := fmt.Sprintf("job-name=bmaas-enrollment-%v", deviceInfo.DeviceName)
	backoffTimer := newExponentialBackoff()
	pods := clientSet.CoreV1().Pods(config.EnrollmentNamespace)
	for start := time.Now(); time.Since(start) < maxElapsedTime; {
		podList, err := pods.List(ctx, metav1.ListOptions{LabelSelector: label})
		if err != nil || len(podList.Items) < 1 {
			log.Error(err, "failed to list the enrollment pod", "site",
				deviceInfo.AvailabilityZone, "region", region)
			time.Sleep(backoffTimer.NextBackOff())
			continue
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodFailed {
				log.Error(errEnrollmentPod, "enrollment pod failed",
					"device", deviceInfo.DeviceName,
					"site", deviceInfo.AvailabilityZone, "region", region)
				return errEnrollmentPod
			} else if pod.Status.Phase == corev1.PodPending {
				log.Info("enrollment pod is pending",
					"device", deviceInfo.DeviceName,
					"site", deviceInfo.AvailabilityZone, "region", region)
			} else if pod.Status.Phase == corev1.PodRunning {
				return nil
			} else {
				log.Info("checking the status of the enrollment pod",
					"device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone,
					"region", region)
			}
			time.Sleep(backoffTimer.NextBackOff())

		}

	}
	return fmt.Errorf("failed to verify the the status of enrollment pod")
}

func newExponentialBackoff() *backoff.ExponentialBackOff {
	return backoff.NewExponentialBackOff()
}
