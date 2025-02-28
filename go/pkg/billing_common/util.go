// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

// Todo: This is a rather very small file right now, and if this file remains being so small, will move this method elsewhere.

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetSchedulerError(schedulerError string, err error) error {
	return fmt.Errorf("scheduler error:%s,error:%w", schedulerError, err)
}

func GetBillingError(billingApiError string, err error) error {
	return fmt.Errorf("billing api error:%s,service error:%v", billingApiError, err)
}

func GetBillingInternalError(billingInternalError string, err error) error {
	return fmt.Errorf("billing internal error:%s,service error:%v", billingInternalError, err)
}

func grpcConnect(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		log := log.FromContext(ctx)
		log.Error(err, "grpcutil.NewClient", "addr", addr)
		return nil, err
	}
	return conn, nil
}

func ToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func ToTime(ts *timestamppb.Timestamp) time.Time {
	return ts.AsTime()
}

// FNV32a hashes using fnv32a algorithm
func FNV32a(text string) uint32 {
	logger := log.FromContext(context.Background()).WithName("FNV32a")
	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(text))
	if err != nil {
		logger.Error(err, "Error while hashes using fnv32a algorithm", text)
		return 0
	}
	return algorithm.Sum32()
}
