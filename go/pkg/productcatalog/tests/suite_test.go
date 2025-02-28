// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"database/sql/driver"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Product Catalog Test Suite")
}

func paramsToDriverValues(params *protodb.ProtoToSql) []driver.Value {
	var vals []driver.Value
	for _, item := range params.GetValues() {
		vals = append(vals, item)
	}
	return vals
}
