// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This is here to show test code that uses the embedded metering
// service
func TestMetering(t *testing.T) {
	usage := pb.UsageCreate{
		TransactionId:  "2023-04-07T10",
		ResourceId:     uuid.New().String(),
		CloudAccountId: "012345678901",
		Timestamp:      timestamppb.Now(),
		Properties: map[string]string{
			"serviceType":    "ComputeAsAService",
			"runningSeconds": "40",
			"region":         "us-west-1",
		},
	}

	resp, err := AriaService.meteringClient.Create(context.Background(), &usage)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("resp=%v\n", resp)
}

func TestMain(m *testing.M) {
	log.SetDefaultLogger()
	ctx := context.Background()
	EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()
	m.Run()
}
