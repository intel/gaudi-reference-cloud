// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

//go:generate mockgen -destination=../mocks/mock_k8s_client.go -package=mock_k8s_client sigs.k8s.io/controller-runtime/pkg/client Client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"encoding/json"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Constants used by the image_finder.
// Config map name used by IRONIC to store the Firmware versions in the namespace.
const CONFIG_MAP_NAME = "ironic-fw-update"

// Config map key used by IRONIC
const CONFIG_MAP_KEY = "ironic-fw-update.json"

// The Component type in Machine Image that signifies the version number.
const FIRMWARE_TYPE = "Firmware kit"

// Version represents the firmware and build version information.
type Version struct {
	BuildVersion string `json:"buildVersion"`
	SpiVersion   string `json:"spiVersion"`
	// FullFwVersion denotes complete firmware version used for GPU firmware check during instance validation.
	// It includes the instance type, build version, SPI version, and SVN version.
	// Example: hl-gaudi2-1.17.0-fw-51.2.0-sec-9
	FullFwVersion string `json:"fullFwVersion"`
}

// InstanceTypeFirmwareVersions is a wrapper for the version mapping.
type FirmwareConfig struct {
	InstanceTypeFirmwareVersions map[string]Version `json:"instanceTypeFirmwareVersions"`
	LastModifiedTime             time.Time
}

// Structure to store the Machine Image Name and the firmware version.
type MachineImage struct {
	Name string
	// supported firmware versions
	SupportedFwVersions []string
}

func (image *MachineImage) isSupported(firmwareVersion string) bool {
	//This method currently does an exact match.
	return slices.Contains(image.SupportedFwVersions, firmwareVersion)
}

type ImageFinder struct {
	ImageClient pb.MachineImageServiceClient
	client      client.Client
}

// NewImageFinder creates a new ImageFinder instance with initialized clients.
func NewImageFinder(imageClient pb.MachineImageServiceClient, k8sClient client.Client) *ImageFinder {
	return &ImageFinder{
		ImageClient: imageClient,
		client:      k8sClient,
	}
}

// Get the version map for the specified metal3 namespace
// This function will be used to get the desire firmware version for a given metal3 namespace.
func (f *ImageFinder) GetFirmwareVersionMap(ctx context.Context, namespace string) (*FirmwareConfig, error) {
	log := log.FromContext(ctx).WithName("ImageFinder.GetFirmwareVersionMap")

	configMap := &corev1.ConfigMap{}
	namespacedName := client.ObjectKey{
		Namespace: namespace,
		Name:      CONFIG_MAP_NAME,
	}
	// Read the ConfigMap from the cluster.
	err := f.client.Get(ctx, namespacedName, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Firmware version map is not present in configmap", "namespace", namespace)
			return nil, nil
		} else {
			return nil, err
		}
	}

	jsonData, ok := configMap.Data[CONFIG_MAP_KEY]
	if !ok {
		// No configuration present.
		log.Info("No firmware configuration present", "namespace", namespace)
		return nil, nil
	}
	// Unmarshal the JSON into the InstanceTypeFirmwareVersions struct
	var firmwareVersions FirmwareConfig
	if err := json.Unmarshal([]byte(jsonData), &firmwareVersions); err != nil {
		return nil, err
	}
	log.Info("Ironic firmware version map", "namespace", namespace, "firmware version mapping", firmwareVersions)
	firmwareVersions.LastModifiedTime = getLastModifiedTime(ctx, configMap)
	return &firmwareVersions, nil
}

// Store the last modified time of the configmap
func getLastModifiedTime(ctx context.Context, configMap *corev1.ConfigMap) time.Time {
	log := log.FromContext(ctx).WithName("ImageFinder.getLastModifiedTime")
	if len(configMap.ManagedFields) > 0 {
		lastOperation := configMap.ManagedFields[len(configMap.ManagedFields)-1]
		return lastOperation.Time.Time
	}
	// If `managedFields` is not available, fall back to the `metadata.creationTimestamp`
	log.Info("Returning creationTimestamp since managedFields not available", logkeys.LastUpdatedTime, configMap.CreationTimestamp.Time)
	return configMap.CreationTimestamp.Time
}

// Fetch the latest image for a desired Firmware version.
// This will be used by the validation operator to image the BM and run validation.
func (f *ImageFinder) GetLatestImage(ctx context.Context, instanceType string, desiredFwVersion *Version) (string, error) {
	log := log.FromContext(ctx).WithName("ImageFinder.GetLatestImage")
	machineImages, err := f.getImages(ctx, instanceType)
	if err != nil {
		return "", err
	}
	if len(machineImages) == 0 {
		log.Info("No images were found", "instanceType", instanceType,
			"desiredFwVersion", desiredFwVersion)
		return "", NonRetryableError(fmt.Sprintf("no machine images for %s found", instanceType))
	}
	var filteredImages []*string
	for _, machineImage := range machineImages {
		log.Info("Going through image", "Image", machineImage.Name, "supportedFwVersions", machineImage.SupportedFwVersions)
		imageName := machineImage.Name
		if desiredFwVersion.BuildVersion == "" || machineImage.isSupported(desiredFwVersion.BuildVersion) {
			// If desired firmware version is empty short list all the images and choose the latest one
			filteredImages = append(filteredImages, &imageName)
		}
	}
	var sortedImages []*string
	if len(filteredImages) == 0 {
		err := RetryableError(fmt.Sprintf("no images found with the desired Firmware version, desiredVersion: %s", desiredFwVersion))
		log.Error(err, "No Images found with the desired Firmware version, retrying ...", "instanceType", instanceType, "desiredFwVersion", desiredFwVersion)
		return "", err
	} else {
		log.Info("Images with the desired Firmware version found, finding the latest one if there is more than one", "instanceType",
			instanceType, "desiredFwVersion", desiredFwVersion, "image names", filteredImages)
		// This function sorts the images and fetches the latest image with the most recent timestamp.
		sortedImages = reverseSuffixSort(filteredImages, "-v")
	}

	return *sortedImages[0], nil
}

// Get the firmware version given the image name.
func (f *ImageFinder) GetImageFwVersion(ctx context.Context, imageName string) (string, error) {
	log := log.FromContext(ctx).WithName("ImageFinder.GetImageFwVersion")
	log.V(9).Info("Invoking image client to fetch image meta", logkeys.Name, imageName)
	machineImage, err := f.ImageClient.Get(ctx, &pb.MachineImageGetRequest{
		Metadata: &pb.MachineImageGetRequest_Metadata{
			Name: imageName,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to get image details for image name %s", imageName)
	}
	versions := getFwVersionsFromImage(ctx, machineImage)
	// Join the versions with `_` to ensure the return value is a valid Kubernetes label value
	return strings.Join(versions, "_"), nil
}

// Return machine images that can be used for a given instance Type.
func (f *ImageFinder) getImages(ctx context.Context, instanceType string) ([]MachineImage, error) {
	log := log.FromContext(ctx).WithName("ImageFinder.getImages")
	log.V(9).Info("Invoking image client to fetch image name", logkeys.InstanceType, instanceType)
	imageList, err := f.ImageClient.Search(ctx, &pb.MachineImageSearchRequest{
		Metadata: &pb.MachineImageSearchRequest_Metadata{
			InstanceType: instanceType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query to fetch image failed %w", err)
	}
	if len(imageList.Items) == 0 {
		return nil, fmt.Errorf("no images found for the instance type %s", instanceType)
	}
	var list []MachineImage
	for _, machineImage := range imageList.Items {
		list = append(list, MachineImage{
			Name:                machineImage.Metadata.Name,
			SupportedFwVersions: getFwVersionsFromImage(ctx, machineImage),
		})
	}
	return list, nil
}

func getFwVersionsFromImage(ctx context.Context, machineImage *pb.MachineImage) []string {
	log := log.FromContext(ctx).WithName("ImageFinder.getFwVersionsFromImage")
	log.Info("getting fw version from ", "image", machineImage)
	var fwVersions []string

	if machineImage.Spec == nil || machineImage.Spec.Components == nil {
		return fwVersions
	}
	for _, component := range machineImage.Spec.Components {
		if component.Type == FIRMWARE_TYPE {
			fwVersions = append(fwVersions, component.Version)
		}
	}
	if len(fwVersions) == 0 {
		log.Info("No firmware version found in machine image", "ImageName", machineImage.Metadata.Name)
	}
	return fwVersions
}
