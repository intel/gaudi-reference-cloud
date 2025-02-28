// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"fmt"
	"strings"
	"testing"
)

func TestKeyMatchAuthzFunc(t *testing.T) {
	tests := []struct {
		key1     string
		key2     string
		expected bool
	}{
		{"/resource1", "/resource1", true},
		{"/resource1", "/resource2", false},
		{"/resource1/sub", "/resource1/*", true},
		{"/resource1?query=123", "/resource1", true},
		{"/resource1/sub", "/resource1/:id", true},
		{"/resource1", "", false},
		{"/v1/cloudaccounts/602030948063/filesystems?t=1723588117721", "/v1/cloudaccounts/:cloud_account_id/filesystems", true},
		{"/v1/cloudaccounts/602030948063/filesystem", "/v1/cloudaccounts/:cloud_account_id/filesystems", false},
		{"/v1/cloudaccounts/602030948063/filesystem", "/v1/cloudaccounts/602030948063/filesystems", false},
	}

	for _, test := range tests {
		result, err := KeyMatchAuthzFunc(test.key1, test.key2)
		if err != nil {
			t.Errorf("Unexpected error for inputs (%s, %s): %s", test.key1, test.key2, err)
		}
		if result != test.expected {
			t.Errorf("KeyMatchAuthzFunc(%s, %s) = %v, expected %v", test.key1, test.key2, result, test.expected)
		}

	}
}

func TestKeyGetAuthzFunc(t *testing.T) {
	bigstring := strings.Repeat("a", 1000)

	tests := []struct {
		key1     string
		key2     string
		key3     string
		expected string
	}{
		{"/resource1/aaaa", "/resource1/:keyGet", "keyGet", "aaaa"},
		{"/resource1/bbb?t=12345", "/resource1/:keyGet", "keyGet", "bbb"},
		{"/resource1/foo?metadata=0000&t=12345", "/resource1/:bar", "bar", "foo"},
		{fmt.Sprintf("/resource1/%s?metadata=0000&t=12345", bigstring), "/resource1/:bar", "bar", bigstring},
	}

	for _, test := range tests {
		result, err := KeyGetAuthzFunc(test.key1, test.key2, test.key3)
		if err != nil {
			t.Errorf("Unexpected error for inputs (%s, %s): %s", test.key1, test.key2, err)
		}
		if result != test.expected {
			t.Errorf("KeyGetAuthzFunc(%s, %s) = %v, expected %v", test.key1, test.key2, result, test.expected)
		}

	}
}

func TestValidateVariadicArgs(t *testing.T) {
	tests := []struct {
		expectedLen int
		args        []interface{}
		expectErr   bool
	}{
		{2, []interface{}{"arg1", "arg2"}, false},
		{2, []interface{}{"arg1"}, true},
		{2, []interface{}{"arg1", 123}, true},
	}

	for _, test := range tests {
		err := validateVariadicArgs(test.expectedLen, test.args...)
		if test.expectErr {
			if err == nil {
				t.Errorf("Expected an error for inputs %v", test.args)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for inputs %v: %s", test.args, err)
			}
		}
	}
}

func TestRemoveQueryParameters(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/resource1?query=123", "/resource1"},
		{"/resource1", "/resource1"},
		{"http://example.com/resource1?query=123", "http://example.com/resource1"},
	}

	for _, test := range tests {
		result := removeQueryParameters(test.path)
		if result != test.expected {
			t.Errorf("removeQueryParameters(%s) = %s, expected %s", test.path, result, test.expected)
		}
	}
}
