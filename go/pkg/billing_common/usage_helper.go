// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UsageValidationErrors struct {
	Usage            *pb.Usage
	ValidationErrors []error
}

type MappedUsageToProduct struct {
	Usage    *pb.Usage
	Products []*pb.Product
}

type UsageMappingToProductError struct {
	Usage   *pb.Usage
	Product *pb.Product
	Error   error
}

// To make sure everything we log w.r.t errors is standardized.
// Standardization needed to be able to filter log messages for reacting appropriately.
const (
	UsageAlreadyReported                   string = "usage validation: usage already reported"
	MissingCloudAccountId                  string = "usage validation: missing cloud account id"
	MissingResourceId                      string = "usage validation: missing resource id"
	MissingTransactionId                   string = "usage validation: missing transaction id"
	InvalidTimestamp                       string = "usage validation: invalid timestamp"
	MissingUsageProperties                 string = "usage validation: missing usage properties"
	MissingRunningSecoundInUsageProperties string = "usage validation: missing running second in usage properties"
	InvalidRunningSecoundInUsageProperties string = "usage validation: invalid running second in usage properties"
	InvalidCloudAccountId                  string = "usage validation: invalid cloud account id"
)

const (
	instanceTypePropertyKey = "instanceType"
	runningSecPropertyKey   = "runningSeconds"
	instanceNamePropertyKey = "instanceName"
	regionNamePropertyKey   = "region"
)

func DetermineUsageDates(ctx context.Context, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	log := log.FromContext(ctx).WithName("UsageHelper.DetermineUsageDates")
	//log.Info("UsageHelper.DetermineUsageDates: enter")
	//defer log.Info("UsageHelper.DetermineUsageDates: return")

	if startTime == nil && endTime == nil {
		currentTime := time.Now()
		start := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
		log.Info("UsageHelper.DetermineUsageDates:", "startTimeValue", start.Format("January 2, 2006"), "endTimeValue", end.Format("January 2, 2006"))
		return ToTimestamp(start), ToTimestamp(end), nil
	}

	if err := validateInputDates(ctx, startTime, endTime); err != nil {
		return nil, nil, err
	}

	return startTime, endTime, nil

}

func validateInputDates(ctx context.Context, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) error {
	//log := log.FromContext(ctx).WithName("UsageHelper.validateInputDates")

	// If either start or end is provided, input is considered invalid
	// UI has to provide either both or none
	if startTime == nil && endTime != nil {
		//log.Info("startTime is nil, endTime is given in input")
		return fmt.Errorf("please provide start time")
	}
	if startTime != nil && endTime == nil {
		//log.Info("endTime is nil, startTime is given in input")
		return fmt.Errorf("please provide end time")
	}

	// Validation when both are provided
	if startTime != nil && endTime != nil && endTime.AsTime().Before(startTime.AsTime()) {
		//log.Info("endTime is before the starttime", "startTime", startTime.AsTime().Format("January 2, 2006"), "endTime", endTime.AsTime().Format("January 2, 2006"))
		return fmt.Errorf("invalid time range")
	}

	// Usages do not display the information for future months.
	if isInFutureMonth(startTime) || isInFutureMonth(endTime) {
		//log.Info("start or end time cannot in future month.", "startTime", startTime.AsTime().Format("January 2, 2006"), "endTime", endTime.AsTime().Format("January 2, 2006"))
		return fmt.Errorf("invalid time range")
	}

	return nil
}

func isInFutureMonth(inputTime *timestamppb.Timestamp) bool {
	if inputTime != nil {
		currentTime := time.Now()
		inputDate := time.Unix(inputTime.Seconds, int64(inputTime.Nanos))
		if inputDate.Year() > currentTime.Year() || (inputDate.Year() == currentTime.Year() && inputDate.Month() > currentTime.Month()) {
			return true
		}
	}
	return false
}

// Add all helper methods that are needed across billing and drivers related to usage.

func ValidateUsageToBeReported(inUsage []*pb.Usage) ([]*pb.Usage, []*UsageValidationErrors, error) {
	if inUsage == nil {
		return nil, nil, errors.New("usages cannot be nil")
	}
	var validUsages []*pb.Usage
	var invalidUsages []*UsageValidationErrors
	for _, usage := range inUsage {
		// default to false to avoid false positives
		usageIsValid := false
		var validationErrors []error
		// Todo: This might not be the full list of validations.
		if usage.GetReported() {
			validationErrors = append(validationErrors, errors.New(UsageAlreadyReported))
		} else if usage.GetTransactionId() == "" {
			validationErrors = append(validationErrors, errors.New(MissingTransactionId))
		} else if usage.GetCloudAccountId() == "" {
			validationErrors = append(validationErrors, errors.New(MissingCloudAccountId))
		} else if usage.GetResourceId() == "" {
			validationErrors = append(validationErrors, errors.New(MissingResourceId))
		} else if usage.Timestamp == nil {
			validationErrors = append(validationErrors, errors.New(InvalidTimestamp))
		} else if usage.GetProperties() == nil {
			validationErrors = append(validationErrors, errors.New(MissingUsageProperties))
		} else {
			usageIsValid = true
		}
		if usage.GetProperties() != nil {
			val, found := usage.GetProperties()[runningSecPropertyKey]
			if !found {
				validationErrors = append(validationErrors, errors.New(MissingRunningSecoundInUsageProperties))
				usageIsValid = false
			} else {
				_, err := strconv.ParseFloat(val, 64)
				if err != nil {
					validationErrors = append(validationErrors, errors.New(InvalidRunningSecoundInUsageProperties))
					usageIsValid = false
				}
			}
		}
		if !IsValidCloudAccountId(usage.GetCloudAccountId()) {
			usageIsValid = false
			validationErrors = append(validationErrors, errors.New(InvalidCloudAccountId))
		}
		if usageIsValid {
			validUsages = append(validUsages, usage)
		} else {
			invalidUsages = append(invalidUsages, &UsageValidationErrors{Usage: usage, ValidationErrors: validationErrors})
		}
	}
	return validUsages, invalidUsages, nil
}

func IsValidCloudAccountId(id string) bool {
	if len(id) != 12 {
		return false
	}
	for _, ch := range id {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}
