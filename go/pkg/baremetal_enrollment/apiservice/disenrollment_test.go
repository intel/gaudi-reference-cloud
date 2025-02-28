// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
)

func newDisenrollDeviceData() *config.BMaaSEnrollmentData {
	return &config.BMaaSEnrollmentData{
		DeviceID:         12345,
		DeviceName:       "compute-node",
		AvailabilityZone: "us-dev-1a",
		RackName:         "someRack",
		ClusterName:      "1",
		EnrollmentStatus: string(dcim.BMDisenroll),
	}
}

func newTestHandler() *BMaaSEnrollment {
	clientset := fake.NewSimpleClientset()
	return &BMaaSEnrollment{
		K8SClient:      fakeClient{},
		JobSpecBuilder: testSpecConfig,
		Region:         testRegion{},
		ClientSet:      clientset,
	}
}

func serveRequest(data config.BMaaSEnrollmentData, handler *BMaaSEnrollment) *httptest.ResponseRecorder {
	router := SetUpRouter()
	_ = handler.AddRoutes(router, username, password)
	jsonValue, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/api/v1/enroll", bytes.NewBuffer(jsonValue))
	req.SetBasicAuth(username, password)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestDisenrollment(t *testing.T) {
	deviceData := newDisenrollDeviceData()
	deviceData.DeviceName = "test-happy-disenrollment"
	handler := newTestHandler()

	response := serveRequest(*deviceData, handler)
	assert.Equal(t, http.StatusCreated, response.Code)
}

func TestDisenrollmentWithExistingActiveJob(t *testing.T) {
	deviceData := newDisenrollDeviceData()
	deviceData.DeviceName = "test-existing-active-job"
	handler := newTestHandler()

	// add an active disenrollment job
	job := handler.JobSpecBuilder.CreateDisenrollmentJobSpec(*deviceData)
	job.Status.Active = 1
	_, err := handler.ClientSet.BatchV1().Jobs(config.EnrollmentDefaultNamespace).
		Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	// it should return a conflict error
	response := serveRequest(*deviceData, handler)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestDisenrollmentWithExistingFailedJob(t *testing.T) {
	deviceData := newDisenrollDeviceData()
	deviceData.DeviceName = "test-existing-failed-job"
	handler := newTestHandler()

	// add a failed disenrollment job
	job := handler.JobSpecBuilder.CreateDisenrollmentJobSpec(*deviceData)
	job.Status.Failed = 1
	_, err := handler.ClientSet.BatchV1().Jobs(config.EnrollmentDefaultNamespace).
		Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	// it should delete the existing job and create a new job
	response := serveRequest(*deviceData, handler)
	assert.Equal(t, http.StatusCreated, response.Code)
	newJob, err := handler.ClientSet.BatchV1().Jobs(config.EnrollmentDefaultNamespace).
		Get(context.TODO(), job.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	assert.Equal(t, int32(0), newJob.Status.Failed)
}
