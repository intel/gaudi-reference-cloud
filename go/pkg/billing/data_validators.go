// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import "regexp"

func ValidateCloudAcctId(cloudAcctId string) bool {
	return regexp.MustCompile(`^\d+$`).MatchString(cloudAcctId)
}

// todo: add validation against the db.
func ValidateCouponCode(couponCode string) bool {
	return regexp.MustCompile(`^\d+$`).MatchString(couponCode)
}
