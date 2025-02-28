// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package buckets

import "fmt"

func IsValidPartNumber(partNum int32) error {
	if partNum < MinPartNum || partNum > MaxPartNum {
		return fmt.Errorf("PartNumber %d is invalid. Valid range is between %d-%d", partNum, MinPartNum, MaxPartNum)
	}
	return nil
}
