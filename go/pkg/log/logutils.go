// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package log

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var reEmailAddress = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func sensitiveFields() []string {
	return []string{
		"email",
		"userdata",
		"sshkey",
		"sshkeyname",
	}
}

func potentiallySensitiveFields() []string {
	return []string{
		"name",
		"owner",
	}
}

func ContainsValidEmailIdValue(val string) bool {
	return reEmailAddress.MatchString(val)
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func TrimObjectForLogs(v reflect.Value) {
	t := v.Type()
	switch t.Kind() {
	// strip pointer if its a pointer
	case reflect.Ptr:
		if !v.IsNil() {
			TrimObjectForLogs(v.Elem())
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			TrimObjectForLogs(v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := v.Field(i)
			if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Struct {
				TrimObjectForLogs(fv)
			} else if f.Type.Kind() == reflect.Ptr && !fv.IsNil() {
				TrimObjectForLogs(fv.Elem())
			} else if f.Type.Kind() == reflect.String {
				redactRequired := Contains(sensitiveFields(), strings.ToLower(f.Name)) ||
					(Contains(potentiallySensitiveFields(), strings.ToLower(f.Name)) && ContainsValidEmailIdValue(fv.String()))
				if redactRequired {
					fv.SetString("[redacted]")
				} else if fv.Len() > 1000 {
					fv.SetString(fmt.Sprintf("Large string of size %d omitted from logs", fv.Len()))
				}
			}
		}
	}
}
