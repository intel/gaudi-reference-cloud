// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
)

var (
	ErrRetryable    = errors.New("retryable error")
	ErrNonRetryable = errors.New("non-retryable error")
)

func RetryableError(msg string) error {
	return errors.Wrap(ErrRetryable, msg)
}

func IsRetryable(err error) bool {
	return errors.Is(err, ErrRetryable)
}

func NonRetryableError(msg string) error {
	return errors.Wrap(ErrNonRetryable, msg)
}

func IsNonRetryable(err error) bool {
	return errors.Is(err, ErrNonRetryable)
}

// commonly used methods

// Return a Label value , if label does not exist return empty string
func GetLabel(key string, bmh *baremetalv1alpha1.BareMetalHost) string {
	labelValue := ""
	if bmh.Labels != nil {
		value, ok := bmh.Labels[key]
		if ok {
			labelValue = value
		}
	}
	return labelValue
}

// Return true if label exists
func CheckLabelExists(key string, bmh *baremetalv1alpha1.BareMetalHost) bool {
	exists := false
	if bmh.Labels != nil {
		_, ok := bmh.Labels[key]
		if ok {
			exists = true
		}
	}
	return exists
}

// method to sort an array in reverse alphabetical order. It uses the substring after the last key in the string to sort.
// e.g: array ["z-key01", "a-key02"] with key as "-key" returns ["a-key02", "z-key01"]
func reverseSuffixSort(arr []*string, key string) []*string {
	comparator := func(a, b string) bool {
		aIndex := strings.LastIndex(a, key)
		bIndex := strings.LastIndex(b, key)
		if aIndex == -1 && bIndex == -1 {
			return a > b // ensure the higher/latest one is chosen.
		} else if aIndex == -1 {
			return false // ensure b is chosen
		} else if bIndex == -1 {
			return true // ensure a is chosen.
		} else {
			substringA := a[aIndex+len(key):]
			substringB := b[bIndex+len(key):]
			return substringA > substringB
		}
	}
	sort.Slice(arr, func(i, j int) bool {
		return comparator(*arr[i], *arr[j])
	})
	return arr
}

func convertToMap(kvpString string) map[string]string {
	result := map[string]string{}
	for _, kvp := range strings.Split(kvpString, "\n") {
		if kvp != "" {
			// only the first seperator is used
			kv := strings.SplitN(kvp, "=", 2)
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func generateRandom8Digit() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 10000000
	max := 99999999
	randomID := min + r.Intn(max-min)
	return strconv.Itoa(randomID)
}

func isTimedOut(startTime time.Time, minutes int) bool {
	// Calculate the time difference in minutes
	now := time.Now().UTC()
	timeDiff := now.Sub(startTime).Minutes()

	// Check if the current time is greater than the specified number of minutes
	return timeDiff > float64(minutes)
}

// Check if entry exists
func exists(array []string, key string) bool {
	for _, e := range array {
		if key == e {
			return true
		}
	}
	return false
}

func getInstanceType(bmh *baremetalv1alpha1.BareMetalHost) (string, error) {
	for k := range bmh.Labels {
		if strings.HasPrefix(k, "instance-type.cloud.intel.com/") {
			substrings := strings.Split(k, "instance-type.cloud.intel.com/")
			if substrings[1] != "" {
				return substrings[1], nil
			} else {
				return "", fmt.Errorf("invalid instance-type label ")
			}
		}
	}
	return "", nil
}
