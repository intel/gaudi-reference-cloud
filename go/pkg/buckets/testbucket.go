// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package buckets

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/testminio"
)

type TestBucket struct {
	testMinIO  *testminio.TestMinIO
	BucketName string
	Bucket     *Bucket
}

func (t *TestBucket) Stop(ctx context.Context) error {
	return t.testMinIO.Stop(ctx)
}

func NewTestBucket(ctx context.Context, bucketName string) (*TestBucket, error) {
	minio := testminio.New()
	url, err := minio.Start(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	bucket, err := New(ctx, &Config{
		URL:           url.String(),
		BucketName:    bucketName,
		DebugMode:     true,
		SigningRegion: "test-region",
	})
	if err != nil {
		return nil, err
	}
	return &TestBucket{
		testMinIO:  minio,
		BucketName: bucketName,
		Bucket:     bucket,
	}, nil
}
