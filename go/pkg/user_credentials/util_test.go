// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package user_credentials

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestGetEmailExclusionRegex(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetEmailExclusionRegex")
	logger.Info("BEGIN")
	defer logger.Info("End")
	re := GetEmailExclusionRegex()
	testCases := []struct {
		input    string
		expected bool
	}{
		{"hello@example.com", false},
		{"user!example@domain.com", true},
		{"user@example.com#", true},
		{"user@example.com/plain", true},
		{"plain_text@domain.com", false},
		{"user<>example@domain.com", true},
		{"user<abc>example@domain.com", true},
		{"no_special_chars@domain.com", false},
		{"example+email@domain.com", true},
		{"user.test.last@example.com", false},
	}
	for _, tc := range testCases {
		result := re.MatchString(tc.input)
		if result != tc.expected {
			t.Errorf("for input '%s', expected %v but got %v", tc.input, tc.expected, result)
		}
	}
}

func TestValidateEmailRequestField(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestValidateEmailRequestField")
	testCases := []struct {
		input string
	}{
		{"hello@example.com"},
		{"user!example@domain.com"},
		{"user@example.com#"},
		{"user@example.com/plain"},
		{"plain_text@domain.com"},
		{"user<>example@domain.com"},
		{"user<abc>example@domain.com"},
		{"user.example@domain.com"},
		{"no_special_chars@domain.com"},
		{"example+email@domain.com"},
	}
	for _, tc := range testCases {
		result, err := ValidateEmailRequestField(tc.input, "member email")
		if err != nil {
			logger.Info("validate email request", "input", tc.input, "result", result, "err", err)
		}
		if len(result) != 0 {
			logger.Info("validate email request", "input", tc.input, "result", result)
		} else {
			logger.Info("validate email request", "input", tc.input)
		}
	}
}
