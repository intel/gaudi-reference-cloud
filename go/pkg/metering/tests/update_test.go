// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"io"
	"strconv"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapSetCopy(src map[string]string, key string, val string) map[string]string {
	dst := map[string]string{}
	for kk, vv := range src {
		dst[kk] = vv
	}
	dst[key] = val
	return dst
}

func TestUpdateValidRecord(t *testing.T) {

	// create some records
	cloudAccountId := cloudaccount.MustNewId()
	resourceID := "3bc52387-da79-4947-a562-ab7a88c38e1998"
	props := map[string]string{
		"availabilityZone":    "us-dev-1a",
		"clusterId":           "harvester1",
		"deleted":             "false",
		"firstReadyTimestamp": "2023-02-16T14:53:29Z",
		"instanceType":        "tiny",
		"region":              "us-dev-1",
	}

	// Test slightly more than MAX_UPDATE_IDS to test the splitting of the
	// large slice
	for ii := 0; ii < query.MAX_UPDATE_IDS+10; ii++ {
		iiStr := strconv.FormatInt(int64(ii), 10)
		_, err := client.Create(context.Background(),
			&pb.UsageCreate{
				CloudAccountId: cloudAccountId,
				TransactionId:  iiStr,
				ResourceId:     resourceID,
				Properties:     mapSetCopy(props, "runningSeconds", iiStr),
				Timestamp:      timestamppb.Now(),
			})
		if err != nil {
			t.Fatalf("create metering record: %v", err)
		}
	}

	filter := &pb.UsageFilter{
		ResourceId: &resourceID,
	}
	stream, err := client.Search(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}

	ids := []int64{}
	for {
		usage, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected errort %v", err)
		}
		if usage.Reported {
			t.Errorf("unexpected Reported=true for usage %v", usage.GetId())
		}
		ids = append(ids, usage.GetId())
	}

	_, err = client.Update(context.Background(), &pb.UsageUpdate{
		Id:       ids,
		Reported: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	stream, err = client.Search(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}

	for {
		usage, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected errort %v", err)
		}

		if !usage.Reported {
			t.Errorf("unexpected usage=false for usage %v", usage.GetId())
		}
	}
}
