// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Based on https://github.com/kubernetes/apimachinery/blob/master/pkg/util/validation/validation.go.

package utils

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

const MaxNumberOfLabels int = 20

const qnameCharFmt string = "[A-Za-z0-9]"
const qnameExtCharFmt string = "[-A-Za-z0-9_.]"
const qualifiedNameFmt string = "(" + qnameCharFmt + qnameExtCharFmt + "*)?" + qnameCharFmt
const qualifiedNameErrMsg string = "must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character"
const qualifiedNameMaxLength int = 63

var qualifiedNameRegexp = regexp.MustCompile("^" + qualifiedNameFmt + "$")

const labelValueFmt string = "(" + qualifiedNameFmt + ")?"
const labelValueErrMsg string = "a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character"

// LabelValueMaxLength is a label's max length
const LabelValueMaxLength int = 63

var labelValueRegexp = regexp.MustCompile("^" + labelValueFmt + "$")

// If the value is not valid, a list of error strings is returned.
// Otherwise an empty list (or nil) is returned.
func IsValidLabelKey(value string) []string {
	var errs []string
	parts := strings.Split(value, "/")
	var name string
	switch len(parts) {
	case 1:
		name = parts[0]
	case 2:
		var prefix string
		prefix, name = parts[0], parts[1]
		if len(prefix) == 0 {
			errs = append(errs, "prefix part "+validation.EmptyError())
		} else if msgs := IsDNS1123Subdomain(prefix); len(msgs) != 0 {
			errs = append(errs, prefixEach(msgs, "prefix part ")...)
		}
	default:
		return append(errs, "a qualified name "+validation.RegexError(qualifiedNameErrMsg, qualifiedNameFmt, "MyName", "my.name", "123-abc")+
			" with an optional DNS subdomain prefix and '/' (e.g. 'example.com/MyName')")
	}
	if len(name) == 0 {
		errs = append(errs, "name part "+validation.EmptyError())
	} else if len(name) > qualifiedNameMaxLength {
		errs = append(errs, "name part "+validation.MaxLenError(qualifiedNameMaxLength))
	}
	if !qualifiedNameRegexp.MatchString(name) {
		errs = append(errs, "name part "+validation.RegexError(qualifiedNameErrMsg, qualifiedNameFmt, "MyName", "my.name", "123-abc"))
	}
	return errs
}

func prefixEach(msgs []string, prefix string) []string {
	for i := range msgs {
		msgs[i] = prefix + msgs[i]
	}
	return msgs
}

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const dns1123SubdomainFmt string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"
const dns1123SubdomainErrorMsg string = "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"

// DNS1123SubdomainMaxLength is a subdomain's max length in DNS (RFC 1123)
const DNS1123SubdomainMaxLength int = 253

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

// IsDNS1123Subdomain tests for a string that conforms to the definition of a
// subdomain in DNS (RFC 1123).
func IsDNS1123Subdomain(value string) []string {
	var errs []string
	if len(value) > DNS1123SubdomainMaxLength {
		errs = append(errs, validation.MaxLenError(DNS1123SubdomainMaxLength))
	}
	if !dns1123SubdomainRegexp.MatchString(value) {
		errs = append(errs, validation.RegexError(dns1123SubdomainErrorMsg, dns1123SubdomainFmt, "example.com"))
	}
	return errs
}

// IsValidLabelValue tests whether the value passed is a valid label value.  If
// the value is not valid, a list of error strings is returned.  Otherwise an
// empty list (or nil) is returned.
func IsValidLabelValue(value string) []string {
	var errs []string
	if len(value) > LabelValueMaxLength {
		errs = append(errs, validation.MaxLenError(LabelValueMaxLength))
	}
	if !labelValueRegexp.MatchString(value) {
		errs = append(errs, validation.RegexError(labelValueErrMsg, labelValueFmt, "MyValue", "my_value", "12345"))
	}
	return errs
}

// The labels field is an optional set of key/value pairs.
// Keys and values must be 63 characters or less, beginning and ending with an alphanumeric character ([a-z0-9A-Z])
// with dashes (-), underscores (_), dots (.), and alphanumerics between.
// A maximum of 20 key/value pairs is allowed.
// See https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set.
// See https://github.com/kubernetes/apimachinery/blob/master/pkg/util/validation/validation.go.
func ValidateLabels(labels map[string]string) error {
	if labels == nil {
		return nil
	}
	if len(labels) > MaxNumberOfLabels {
		return fmt.Errorf("the number of labels must not exceed %d", MaxNumberOfLabels)
	}
	var errs []string
	for k, v := range labels {
		errs = append(errs, IsValidLabelKey(k)...)
		errs = append(errs, IsValidLabelValue(v)...)
	}
	if len(errs) > 0 {
		return fmt.Errorf("labels are invalid: %v", strings.Join(errs, "; "))
	}
	return nil
}
