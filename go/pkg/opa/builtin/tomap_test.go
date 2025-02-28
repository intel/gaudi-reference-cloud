// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestToMap(t *testing.T) {
	prod := pb.Product{
		Name: "abc",
		Rates: []*pb.Rate{
			{
				AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
				Rate:        "0.10",
			},
		},
	}
	prodMap := ProtoMessageToMap(&prod)
	if prodMap["name"] != "abc" {
		t.Errorf("expected name, got %v", prodMap["name"])
	}
	rates, ok := prodMap["rates"].([]any)
	if !ok {
		t.Fatal("prodmap[\"rates\"] is not a slice")
	}
	for _, elem := range rates {
		rate, ok := elem.(map[string]any)
		if !ok {
			t.Fatal("non-rate in rate slice")
		}
		typ, ok := rate["accountType"].(protoreflect.EnumNumber)
		if !ok {
			t.Fatal("rate has non-int for accountType")
		}
		if typ != pb.AccountType_ACCOUNT_TYPE_PREMIUM.Number() {
			t.Error("rate has wrong accountType")
		}
		amount, ok := rate["rate"].(string)
		if !ok {
			t.Fatal("rate has non-string rate")
		}
		if amount != "0.10" {
			t.Errorf("rate has wrong rate %v", amount)
		}
	}
}
