// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
)

var (
	testEnrollmentData = config.BMaaSEnrollmentData{
		DeviceID:         12345,
		DeviceName:       "compute-node",
		AvailabilityZone: "us-dev-1a",
		RackName:         "someRack",
		ClusterName:      "1",
		EnrollmentStatus: string(dcim.BMEnroll),
	}
	testSpecConfig = config.EnrollmentJobConfig{
		PlaybookImage:          "ubuntu:latest",
		Backofflimit:           2,
		ProvisioningDuration:   1,
		DeprovisioningDuration: 1,
	}
	username = "admin"
	password = "secret"
)

type fakeClient struct{}

func (f fakeClient) GetK8SClientset(
	ctx context.Context, region string, availabilityZone string, objects ...runtime.Object) (kubernetes.Interface, error) {
	return fake.NewSimpleClientset(objects...), nil
}

type fakeJobSpecBuilder struct{}

// Incorrect job spec to fail the job
func (s fakeJobSpecBuilder) CreateEnrollmentJobSpec(deviceInfo config.BMaaSEnrollmentData) *batchv1.Job {
	jobName := fmt.Sprintf("%s-%s", config.EnrollmentJobNamePrefix, deviceInfo.DeviceName)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "default",
			Labels: map[string]string{
				config.DeviceInfoLabelKey: jobName,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Command: []string{"/bin/sh", "-c", "sleep 120"},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
}

func (s fakeJobSpecBuilder) CreateDisenrollmentJobSpec(deviceInfo config.BMaaSEnrollmentData) *batchv1.Job {
	// not implemented
	return nil
}

func jobReactor(deviceName string, activeJobCount, succeededJobCount, failedJobCount int32) testcore.ReactionFunc {
	jobName := fmt.Sprintf("bmaas-enrollment-%v", deviceName)
	jobMetadata := metav1.ObjectMeta{
		Name:      jobName,
		Namespace: config.EnrollmentNamespace,
		Labels: map[string]string{
			config.DeviceInfoLabelKey: jobName,
		},
	}
	var activeDeadline int64
	activeDeadline = int64(testSpecConfig.ProvisioningDuration + testSpecConfig.DeprovisioningDuration)
	job := &batchv1.Job{
		ObjectMeta: jobMetadata,
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            config.EnrollmentJobNamePrefix,
							Image:           testSpecConfig.PlaybookImage,
							ImagePullPolicy: v1.PullAlways,
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "enrollment",
				},
			},
			BackoffLimit:          &testSpecConfig.Backofflimit,
			ActiveDeadlineSeconds: &activeDeadline,
		},
		Status: batchv1.JobStatus{
			Active:    activeJobCount,
			Succeeded: succeededJobCount,
			Failed:    failedJobCount,
		},
	}

	return func(action testcore.Action) (handled bool, ret runtime.Object, err error) {
		jobList := &batchv1.JobList{
			Items: []batchv1.Job{*job},
		}
		return true, jobList, nil
	}
}

func podReactor(deviceName string, podPhase v1.PodPhase) testcore.ReactionFunc {
	podName := fmt.Sprintf("bmaas-enrollment-%v", deviceName)
	podMetadata := metav1.ObjectMeta{
		Name:      podName,
		Namespace: config.EnrollmentNamespace,
		Labels: map[string]string{
			"job-name": podName,
		},
	}
	pod := &v1.Pod{
		ObjectMeta: podMetadata,
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            config.EnrollmentJobNamePrefix,
					Image:           testSpecConfig.PlaybookImage,
					ImagePullPolicy: v1.PullAlways,
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: "enrollment",
		},
		Status: v1.PodStatus{
			Phase: podPhase,
		},
	}
	return func(action testcore.Action) (handled bool, ret runtime.Object, err error) {
		podList := &v1.PodList{
			Items: []v1.Pod{*pod},
		}
		return true, podList, nil
	}
}

type testRegion struct{}

func (r testRegion) GetClusterRegion(deviceInfo config.BMaaSEnrollmentData) (string, error) {
	return "us-dev-1", nil
}

type emptyRegion struct{}

func (r emptyRegion) GetClusterRegion(deviceInfo config.BMaaSEnrollmentData) (string, error) {
	return "", nil
}

type invalidClient struct{}

func (f invalidClient) GetK8SClientset(
	ctx context.Context, region string, availabilityZone string, objects ...runtime.Object) (kubernetes.Interface, error) {
	return nil, nil
}

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	return router
}

func TestNewEnrollment(t *testing.T) {
	deviceName := "compute-node-1"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()
	newEnrollment, _ := NewBMaaSEnrollment(fakeClient{}, testSpecConfig, testRegion{})
	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentWithExistingJob(t *testing.T) {
	deviceName := "compute-node-2"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()
	testClient := fakeClient{}
	// Create a job using fake client
	clientSet, _ := testClient.GetK8SClientset(context.TODO(), "", "", testSpecConfig.CreateEnrollmentJobSpec(testEnrollmentData))
	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestEnrollmentWithExistingFailedJob(t *testing.T) {
	deviceName := "compute-node-3"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	jobSpec := testSpecConfig.CreateEnrollmentJobSpec(testEnrollmentData)

	clientSet.BatchV1().Jobs(config.EnrollmentNamespace).Create(context.TODO(), jobSpec, metav1.CreateOptions{})

	// Add a reactor to fail the job
	clientSet.PrependReactor("list", "*", jobReactor(deviceName, 0, 0, 1))

	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentWithExistingSucceededJob(t *testing.T) {
	deviceName := "compute-node-4"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	jobSpec := testSpecConfig.CreateEnrollmentJobSpec(testEnrollmentData)

	clientSet.BatchV1().Jobs(config.EnrollmentNamespace).Create(context.TODO(), jobSpec, metav1.CreateOptions{})

	// Add a reactor to succeed the job
	clientSet.PrependReactor("list", "*", jobReactor(deviceName, 0, 1, 0))

	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentWithPodRunning(t *testing.T) {
	deviceName := "compute-node-5"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(1 * time.Second)
	// Add a reactor to list jobs with jobStatus active == 1
	clientSet.PrependReactor("list", "jobs", jobReactor(deviceName, 1, 0, 0))
	// Add a reactor with pod status running
	clientSet.PrependReactor("list", "pods", podReactor(deviceName, v1.PodRunning))

	time.Sleep(6 * time.Second)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentWithPodPending(t *testing.T) {
	deviceName := "compute-node-6"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(1 * time.Second)
	// Add a reactor to list jobs with jobStatus active == 1
	clientSet.PrependReactor("list", "jobs", jobReactor(deviceName, 1, 0, 0))
	// Add a reactor with pod status pending
	clientSet.PrependReactor("list", "pods", podReactor(deviceName, v1.PodPending))

	time.Sleep(6 * time.Second)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentWithPodMissing(t *testing.T) {
	deviceName := "compute-node-7"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(1 * time.Second)
	// Add a reactor to list jobs with jobStatus active == 1
	clientSet.PrependReactor("list", "jobs", jobReactor(deviceName, 1, 0, 0))

	time.Sleep(6 * time.Second)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentFailedActiveJobs(t *testing.T) {
	deviceName := "compute-node-8"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	time.Sleep(5 * time.Second)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnrollmentWithFailedJobCreation(t *testing.T) {
	deviceName := "compute-node-9"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()

	newEnrollment := BMaaSEnrollment{
		K8SClient:      fakeClient{},
		JobSpecBuilder: fakeJobSpecBuilder{},
		Region:         testRegion{},
	}
	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEnrollmentWithBadRequest(t *testing.T) {
	deviceName := "compute-node-10"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()

	newEnrollment := BMaaSEnrollment{
		K8SClient:      fakeClient{},
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
	}
	_ = newEnrollment.AddRoutes(router, username, password)
	incorrectEnrollmentData := struct {
		Name string
	}{
		Name: "bmaas-compute-node-flex-lab",
	}
	jsonValue, _ := json.Marshal(incorrectEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvalidURI(t *testing.T) {
	deviceName := "compute-node-11"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()
	newEnrollment, _ := NewBMaaSEnrollment(fakeClient{}, testSpecConfig, testRegion{})
	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/test", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEmptyRegion(t *testing.T) {
	deviceName := "compute-node-12"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()
	newEnrollment, _ := NewBMaaSEnrollment(fakeClient{}, testSpecConfig, emptyRegion{})
	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInvalidK8SClientset(t *testing.T) {
	deviceName := "compute-node-13"
	testEnrollmentData.DeviceName = deviceName
	router := SetUpRouter()
	newEnrollment, _ := NewBMaaSEnrollment(invalidClient{}, testSpecConfig, emptyRegion{})
	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
func TestEnrollmentWithPodFailed(t *testing.T) {
	deviceName := "compute-node-14"
	testEnrollmentData.DeviceName = deviceName
	maxElapsedTime = 3 * time.Second
	router := SetUpRouter()
	testClient := fakeClient{}

	// Create a job using the fake client
	clientSet := fake.NewSimpleClientset()
	newEnrollment := BMaaSEnrollment{
		K8SClient:      testClient,
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientSet,
	}

	_ = newEnrollment.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(testEnrollmentData)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(1 * time.Second)
	// Add a reactor to list jobs with jobStatus active == 1
	clientSet.PrependReactor("list", "jobs", jobReactor(deviceName, 1, 0, 0))
	// Add a reactor with pod status failed
	clientSet.PrependReactor("list", "pods", podReactor(deviceName, v1.PodFailed))

	time.Sleep(6 * time.Second)
	assert.Equal(t, http.StatusCreated, w.Code)
}
