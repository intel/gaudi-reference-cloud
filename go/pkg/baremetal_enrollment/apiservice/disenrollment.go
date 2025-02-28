// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func (e *BMaaSEnrollment) disenrollBMaaSNode(ctx *gin.Context, deviceInfo config.BMaaSEnrollmentData, region string) {
	log := log.FromContext(ctx).WithName("apiservice.enrollment.disenrollBMaaSNode")

	// delete a running enrollment job while keeping a failed or completed job for a record
	enrollmentJobList, err := e.getEnrollmentJob(ctx, deviceInfo)
	if err != nil {
		log.Error(err, "failed to list enrollment jobs", "device", deviceInfo)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(enrollmentJobList.Items) > 0 {
		if enrollmentJobList.Items[0].Status.Active > 0 {
			if err := e.deleteEnrollmentJob(ctx, deviceInfo); err != nil {
				log.Error(err, "failed to delete active enrollment job", "device", deviceInfo)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// create a new disenrollment job if not exists
	disenrollmentJobList, err := e.getDisenrollmentJob(ctx, deviceInfo)
	if err != nil {
		log.Error(err, "failed to list disenrollment jobs", "device", deviceInfo)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(disenrollmentJobList.Items) == 1 {
		if (disenrollmentJobList.Items[0].Status.Failed > 0) || (disenrollmentJobList.Items[0].Status.Succeeded > 0) {
			if err := e.deleteDisenrollmentJob(ctx, deviceInfo); err != nil {
				log.Error(err, "failed to delete failed or completed enrollment job", "device", deviceInfo.DeviceName,
					"site", deviceInfo.AvailabilityZone, "region", region)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			log.Info("a disenrollment job is already running", "device", deviceInfo.DeviceName,
				"site", deviceInfo.AvailabilityZone, "region", region)
			ctx.Status(http.StatusConflict)
			return
		}
	}
	disenrollmentSpec := e.JobSpecBuilder.CreateDisenrollmentJobSpec(deviceInfo)
	if err := e.launchJob(ctx, disenrollmentSpec); err != nil {
		log.Error(err, "failed to create disenrollment job",
			"device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusCreated)
	log.Info("created disenrollment job successfully", "device", deviceInfo.DeviceName, "site", deviceInfo.AvailabilityZone, "region", region)
}

func (e *BMaaSEnrollment) getDisenrollmentJob(ctx *gin.Context, deviceInfo config.BMaaSEnrollmentData) (batchv1.JobList, error) {
	label := fmt.Sprintf("%s=%s-%s", config.DeviceInfoLabelKey, config.DisenrollmentJobNamePrefix, deviceInfo.DeviceName)
	jobs := e.ClientSet.BatchV1().Jobs(config.EnrollmentNamespace)
	jobList, err := jobs.List(ctx, metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return batchv1.JobList{}, err
	}
	return *jobList, nil
}

func (e *BMaaSEnrollment) deleteDisenrollmentJob(ctx *gin.Context, deviceInfo config.BMaaSEnrollmentData) error {
	jobName := fmt.Sprintf("%s-%s", config.DisenrollmentJobNamePrefix, deviceInfo.DeviceName)
	return e.deleteJob(ctx, jobName)
}
