// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/casbin/casbin/v2/util"
)

func KeyMatchAuthzFunc(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(2, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "KeyMatchAuthzFunc", err)
	}

	key1 := removeQueryParameters(args[0].(string)) // remove query parameters from r.obj / path
	key2 := args[1].(string)

	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`:[^/]+`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")

	return bool(RegexMatch(key1, "^"+key2+"$")), nil
}

func KeyGetAuthzFunc(args ...interface{}) (interface{}, error) {
	if err := validateVariadicArgs(3, args...); err != nil {
		return false, fmt.Errorf("%s: %s", "KeyGetAuthzFunc", err)
	}

	key1, ok1 := args[0].(string)
	key2, ok2 := args[1].(string)
	key3, ok3 := args[2].(string)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("%s: %s", "KeyGetAuthzFunc", "arguments are not strings")
	}

	// Parse the URL to ignore the query string
	parsedURL, err := url.Parse(key1)
	if err != nil {
		return nil, fmt.Errorf("%s: %s : %s", "KeyGetAuthzFunc", "invalid URL", key1)
	}

	// Use the path without the query string
	key1Path := parsedURL.Path

	return util.KeyGet2(key1Path, key2, key3), nil
}

func RegexMatch(key1 string, key2 string) bool {
	res, err := regexp.MatchString(key2, key1)
	if err != nil {
		panic(err)
	}
	return res
}

func validateVariadicArgs(expectedLen int, args ...interface{}) error {
	if len(args) != expectedLen {
		return fmt.Errorf("expected %d arguments, but got %d", expectedLen, len(args))
	}

	for _, p := range args {
		_, ok := p.(string)
		if !ok {
			return errors.New("argument must be a string")
		}
	}

	return nil
}

func removeQueryParameters(path string) string {
	parsedURL, err := url.Parse(path)
	if err != nil {
		return path
	}
	parsedURL.RawQuery = ""
	return parsedURL.String()
}
