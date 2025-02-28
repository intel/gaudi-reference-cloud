// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	mock_client "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/baremetal-validation-operator/controllers/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const IMAGE_NAME = "ubuntu-22.04-server-cloudimg-amd64-latest"

func TestMachineImage_isSupported(t *testing.T) {
	type fields struct {
		SupportedFwVersions []string
	}

	tests := []struct {
		name           string
		fields         fields
		inputFwVersion string
		want           bool
	}{
		{
			name: "test-01",
			fields: fields{
				SupportedFwVersions: []string{"1.16.0", "1.17.0"},
			},
			inputFwVersion: "1.16.0",
			want:           true,
		},
		{
			name: "test-02",
			fields: fields{
				SupportedFwVersions: []string{"1.16.0", "1.17.0"},
			},
			inputFwVersion: "1.15.0",
			want:           false,
		},
		{
			name: "test-03",
			fields: fields{
				SupportedFwVersions: []string{"1.16.0", "1.17.0"},
			},
			inputFwVersion: "1.16.1",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image := &MachineImage{
				Name:                tt.name,
				SupportedFwVersions: tt.fields.SupportedFwVersions,
			}
			if got := image.isSupported(tt.inputFwVersion); got != tt.want {
				t.Errorf("MachineImage.isSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFinder_GetFirmwareVersionMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	ctx := context.Background()
	configMapNamespace := "metal3-1"
	t.Run("ConfigMap exists with data", func(t *testing.T) {
		instanceTypeVerMap := map[string]Version{
			"bm-icp-gaudi2": {
				BuildVersion: "1.0.0",
				SpiVersion:   "2.0.0",
			},
		}
		expectedConfigMap := &corev1.ConfigMap{
			Data: map[string]string{
				CONFIG_MAP_KEY: create_ironic_fw_update_json(t, instanceTypeVerMap),
			},
		}
		mockClient.EXPECT().
			Get(ctx, types.NamespacedName{Name: CONFIG_MAP_NAME, Namespace: configMapNamespace}, gomock.Any()).
			SetArg(2, *expectedConfigMap).
			Return(nil)
		imgFinder := NewImageFinder(nil, mockClient)
		fwCfg, err := imgFinder.GetFirmwareVersionMap(ctx, configMapNamespace)
		if err != nil {
			t.Fatalf("GetFirmwareVersionMap returned an unexpected error %v", err)
		}
		if !reflect.DeepEqual(fwCfg.InstanceTypeFirmwareVersions, instanceTypeVerMap) {
			t.Fatalf("GetFirmwareVersionMap returned an unexpected error %v", err)
		}
	})

	t.Run("ConfigMap not present, NotFound error is returned", func(t *testing.T) {
		mockClient.EXPECT().
			Get(ctx, types.NamespacedName{Name: CONFIG_MAP_NAME, Namespace: configMapNamespace}, gomock.Any()).
			Return(errors.NewNotFound(corev1.Resource("configmap"), CONFIG_MAP_NAME))

		imgFinder := NewImageFinder(nil, mockClient)
		fwCfg, err := imgFinder.GetFirmwareVersionMap(ctx, configMapNamespace)
		if err != nil {
			t.Fatalf("GetFirmwareVersionMap returned an unexpected error %v", err)
		}
		if fwCfg != nil {
			t.Fatalf("GetFirmwareVersionMap should return nil")
		}
	})

	t.Run("ConfigMap exists but mapping is missing", func(t *testing.T) {

		expectedConfigMap := &corev1.ConfigMap{
			Data: map[string]string{
				"FIRMWARE_IMAGE_SERVER": "http://1.1.1.1:50001",
			},
		}
		mockClient.EXPECT().
			Get(ctx, types.NamespacedName{Name: CONFIG_MAP_NAME, Namespace: configMapNamespace}, gomock.Any()).
			SetArg(2, *expectedConfigMap).
			Return(nil)
		imgFinder := NewImageFinder(nil, mockClient)
		fwCfg, err := imgFinder.GetFirmwareVersionMap(ctx, configMapNamespace)
		if err != nil {
			t.Fatalf("GetFirmwareVersionMap returned an unexpected error %v", err)
		}
		if fwCfg != nil {
			t.Fatalf("GetFirmwareVersionMap should return nil since instanceTypeFirmwareVersions entry is missing")
		}
	})
}

func TestImageFinder_GetLatestImage(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_client.NewMockClient(ctrl)
	ctx := context.Background()
	t.Run("GetLatestImage when image is present.", func(t *testing.T) {
		mockImageClient := NewMockImageClient(ctx, ctrl, "1.16.0")
		imgFinder := NewImageFinder(mockImageClient, mockClient)
		imageName, err := imgFinder.GetLatestImage(ctx, "bm-icp-gaudi2", &Version{
			BuildVersion: "1.16.0",
			SpiVersion:   "2.0.0",
		})
		if err != nil {
			t.Fatalf("GetLatestImage returned an unexpected error %v", err)
		}
		if imageName != IMAGE_NAME {
			t.Fatalf("Incorrct image name returned")
		}
	})
	t.Run("GetLatestImage when image is present with backward compatability.", func(t *testing.T) {
		mockImageClient := NewMockImageClient(ctx, ctrl, "1.16.0", "1.15.0")
		imgFinder := NewImageFinder(mockImageClient, mockClient)
		imageName, err := imgFinder.GetLatestImage(ctx, "bm-icp-gaudi2", &Version{
			BuildVersion: "1.15.0",
			SpiVersion:   "2.0.1",
		})
		if err != nil {
			t.Fatalf("GetLatestImage returned an unexpected error %v", err)
		}
		if imageName != IMAGE_NAME {
			t.Fatalf("Incorrct image name returned")
		}
	})

	t.Run("GetLatestImage: no matching fw version.", func(t *testing.T) {
		mockImageClient := NewMockImageClient(ctx, ctrl, "1.16.0", "1.15.0")
		imgFinder := NewImageFinder(mockImageClient, mockClient)
		_, err := imgFinder.GetLatestImage(ctx, "bm-icp-gaudi2", &Version{
			BuildVersion: "1.14.0",
			SpiVersion:   "2.0.1",
		})
		if err == nil {
			t.Fatalf("GetLatestImage should return an error")
		}
		if !IsRetryable(err) {
			t.Fatalf("Expected a non-retryable error")
		}
	})

	t.Run("GetLatestImage: multipleImages.", func(t *testing.T) {
		mockImageClient := NewMockMultiImageClient(ctx, ctrl, "imageA-20240810", "imageA-20240819")
		imgFinder := NewImageFinder(mockImageClient, mockClient)
		imageName, err := imgFinder.GetLatestImage(ctx, "bm-icp-gaudi2", &Version{
			BuildVersion: "1.17.0",
			SpiVersion:   "2.0.1",
		})
		if err != nil {
			t.Fatalf("GetLatestImage returned an unexpected error %v", err)
		}
		if imageName != "imageA-20240819" {
			t.Fatalf("Image name returned is incorrect")
		}
	})
}

func create_ironic_fw_update_json(t *testing.T, versionMap map[string]Version) string {
	cfg := FirmwareConfig{
		InstanceTypeFirmwareVersions: versionMap,
	}
	userJson, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal("Error while marshalling FirmwareConfig")
	}
	return string(userJson)
}

func NewMockImageClient(ctx context.Context, mockController *gomock.Controller, firmwareKitVersions ...string) pb.MachineImageServiceClient {
	machineImageClient := pb.NewMockMachineImageServiceClient(mockController)
	var components []*pb.MachineImageComponent

	for _, ver := range firmwareKitVersions {
		component := &pb.MachineImageComponent{
			Name:    "Intel Gaudi SW",
			Type:    FIRMWARE_TYPE,
			Version: ver,
		}
		components = append(components, component)
	}

	resp := &pb.MachineImageSearchResponse{
		Items: []*pb.MachineImage{
			{
				Metadata: &pb.MachineImage_Metadata{
					Name: IMAGE_NAME,
				},
				Spec: &pb.MachineImageSpec{
					InstanceTypes: []string{"bm-icp-gaudi2"},
					Components:    components,
				},
			},
		},
	}
	machineImageClient.EXPECT().Search(gomock.Any(), gomock.Any()).Return(resp, nil)

	return machineImageClient
}

func NewMockMultiImageClient(ctx context.Context, mockController *gomock.Controller, imageNames ...string) pb.MachineImageServiceClient {
	machineImageClient := pb.NewMockMachineImageServiceClient(mockController)
	var images []*pb.MachineImage

	for _, imageName := range imageNames {
		image := &pb.MachineImage{
			Metadata: &pb.MachineImage_Metadata{
				Name: imageName,
			},
			Spec: &pb.MachineImageSpec{
				InstanceTypes: []string{"bm-icp-gaudi2"},
				Components: []*pb.MachineImageComponent{
					{
						Name:    "Intel Gaudi SW",
						Type:    FIRMWARE_TYPE,
						Version: "1.17.0",
					},
				},
			},
		}
		images = append(images, image)
	}

	resp := &pb.MachineImageSearchResponse{
		Items: images,
	}
	machineImageClient.EXPECT().Search(gomock.Any(), gomock.Any()).Return(resp, nil)

	return machineImageClient
}
