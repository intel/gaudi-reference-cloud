// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sethvargo/go-password/password"
)

func GenerateRandomPassword() (string, error) {
	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	// Restrict symbols to only those defined in the Symbols list
	gen, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "~!@#$%^*()_+`-={}|[]:?,./",
	})
	randomPassword, err := gen.Generate(20, 5, 5, false, false)
	if err != nil {
		return "", fmt.Errorf("could not generate a random password: %v", err)
	}
	return randomPassword, nil
}

func ConvertToBytes(size string) string {
	var unit string
	zeros := "000000000"
	if strings.HasSuffix(size, "GB") {
		unit = "GB"
	} else if strings.HasSuffix(size, "TB") {
		unit = "TB"
		zeros = zeros + "000"
	} else {
		return ""
	}
	strs := strings.Split(size, unit)
	if len(strs) != 2 {
		return ""
	}
	return strs[0] + zeros
}

func GenerateKMSPath(cloudAccount, clusterUUID string, isUser bool) string {
	if isUser {
		return fmt.Sprintf("staas/tenant/%s/%s/u%s", cloudAccount, clusterUUID, cloudAccount)
	}
	return fmt.Sprintf("staas/tenant/%s/%s/ns%s", cloudAccount, clusterUUID, cloudAccount)
}

func GenerateFilesystemUser(cloudAccount string) string {
	return fmt.Sprintf("u%s", cloudAccount)
}

func GenerateFilesystemNamespaceUser(cloudAccount string) string {
	return fmt.Sprintf("ns%s", cloudAccount)
}

func GenerateVASTNamespaceName(cloudAccount string) string {
	return fmt.Sprintf("vastns-%s", cloudAccount)
}

func GenerateVASTVolumePath(cloudAccount, name string) string {
	// if volPath == "" {
	// 	return fmt.Sprintf("/%s", cloudAccount)
	// }
	// if strings.HasPrefix(volPath, "/") {
	// 	return fmt.Sprintf("/%s%s", cloudAccount, volPath)
	// } else {
	// 	return fmt.Sprintf("/%s/%s", cloudAccount, volPath)
	// }
	return fmt.Sprintf("/%s", name)
}

func GenerateVASTFilesystemName(cloudAccount, name string) string {
	// if volPath == "" {
	// 	return fmt.Sprintf("/%s", cloudAccount)
	// }
	// if strings.HasPrefix(volPath, "/") {
	// 	return fmt.Sprintf("/%s%s", cloudAccount, volPath)
	// } else {
	// 	return fmt.Sprintf("/%s/%s", cloudAccount, volPath)
	// }
	return fmt.Sprintf("%s-%s", cloudAccount, name)
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
