package utils

import (
	"github.com/tidwall/gjson"
	"strings"
)

var configData string

func LoadConfig(filepath string, filename string) error {
	configData, _ = ConvertFileToString(filepath, filename)
	return nil
}

func GetJsonValue(key string) string {
	return gjson.Get(configData, key).String()
}

func GetJsonObject(key string) gjson.Result {
	return gjson.Get(string(configData), key)
}

func GetJsonArray(key string) []gjson.Result {
	return gjson.Get(configData, key).Array()
}

func GetMachineImagesList(key string) []string {
	machineImages := GetJsonArray(key)
	var machineImagesList []string
	for _, item := range machineImages {
		machineImagesList = append(machineImagesList, item.String())
	}
	return machineImagesList
}

func GetInstanceTypesList(key string) []string {
	instanceTypes := GetJsonArray(key)
	var instanceTypesList []string
	for _, item := range instanceTypes {
		instanceTypesList = append(instanceTypesList, item.String())
	}
	return instanceTypesList
}

func GetBUAllImagesMapping() (map[string][]string, error) {
	allimagesMapping := GetJsonArray("AllImagesMapping")
	imageMapping := make(map[string][]string)
	for _, result := range allimagesMapping {
		instanceType := result.Get("instanceType").String()
		images := result.Get("machineImage").String()                // Get the images as a string
		imageList := strings.Split(strings.Trim(images, "[]"), ", ") // Remove outer brackets and split
		imageMapping[instanceType] = imageList
	}

	return imageMapping, nil
}
