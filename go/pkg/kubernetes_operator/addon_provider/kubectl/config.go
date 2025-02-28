// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package kubectl

type S3AddonConfig struct {
	URL        string `json:"url"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	UseSSL     bool   `json:"useSSL"`
	BucketName string `json:"bucketName"`
	S3Path     string `json:"s3Path"`
}
