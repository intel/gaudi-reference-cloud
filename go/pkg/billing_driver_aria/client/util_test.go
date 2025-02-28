// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"encoding/json"
	"reflect"
	"testing"
)

var js1 = []byte(`
{
    "strKey": "val",
    "intKey": 1,
    "floatKey": 1.0,
    "mapKey": {
        "strKey": "val",
        "intKey": 1,
        "floatKey": 1.0
    },
    "nestedArr": [ ["a", "b", "c" ], ["c", "d", "e" ]],
    "arrayKey": [
        {
            "strKey": "val",
            "intKey": 1,
            "floatKey": 1.0,
            "mapKey": {
                "strkey": "val",
                "intKey": 1,
                "floatKey": 1.0
            }
        },
        {
            "strKey": "val",
            "intKey": 1,
            "floatKey": 1.0,
            "mapKey": {
                "strkey": "val",
                "intKey": 1,
                "floatKey": 1.0
            },
            "arrayKey": [
                {
                    "strKey": "val",
                    "intKey": 1,
                    "floatKey": 1.0,
                    "mapKey": {
                        "strkey": "val",
                        "intKey": 1,
                        "floatKey": 1.0
                    }
                },
                {
                    "strKey": "val",
                    "intKey": 1,
                    "floatKey": 1.0,
                    "mapKey": {
                        "strkey": "val",
                        "intKey": 1,
                        "floatKey": 1.0
                    },
                    "arrayKey": [
                        {
                            "strKey": "val",
                            "intKey": 1,
                            "floatKey": 1.0,
                            "mapKey": {
                                "strkey": "val",
                                "intKey": 1,
                                "floatKey": 1.0
                            }
                        },
                        {
                            "strKey": "val",
                            "intKey": 1,
                            "floatKey": 1.0,
                            "mapKey": {
                                "strkey": "val",
                                "intKey": 1,
                                "floatKey": 1.0
                            }
                        }
                    ]
                }
            ]
        }
    ]
}
`)

var js2 = []byte(`
{"rest_call":"create_new_plan_m","output_format":"json","client_no":5025576,"auth_key":"[REDACTED]","alt_caller_id":"","client_plan_id":".IDC.Master","plan_name":"Test Product/Plan Name","plan_type":"Master Recurring Plan","currency":"usd","active":1,"schedule":[{"schedule_name":"ACCOUNT_TYPE_PREMIUM","currency_cd":"usd","client_rate_schedule_id":".ACCOUNT_TYPE_PREMIUM","is_default":1},{"schedule_name":"ACCOUNT_TYPE_ENTERPRISE","currency_cd":"usd","client_rate_schedule_id":".ACCOUNT_TYPE_ENTERPRISE"}],"service":[{"name":"Test Service Plan Name","client_service_id":".Test Service Plan Id","service_type":"Usage-Based","gl_cd":1,"rate_type":"Tiered Pricing","tier":[{"schedule":[{"from":1,"amount":0},{"from":1,"amount":0}]}]}],"supplemental_obj_field":[{"field_name":"SYNC TO SAP GTS","field_value":["Yes"]},{"field_name":"PCQ ID","field_value":["Test PCQ"]}]}`)

func testQuery(t *testing.T, label string, jsonStr []byte) {
	str, err := GenerateQuery(jsonStr)
	if err != nil {
		t.Error(err)
		return
	}
	parsedQuery, err := ParseQuery(str)
	if err != nil {
		t.Error(err)
		return
	}
	var parsedJson any
	err = json.Unmarshal(jsonStr, &parsedJson)
	if err != nil {
		t.Error("unable to json parse js1")
		return
	}
	if !reflect.DeepEqual(parsedQuery, parsedJson) {
		t.Errorf("ParseQuery(GenerateQuery(%v) != Unmarshal(%v)", label, label)
		t.Logf("query: %v\n", str)

		pq, _ := json.Marshal(parsedQuery)
		t.Logf("parsedQuery: %v\n", string(pq))

		jq, _ := json.Marshal(parsedJson)
		t.Logf("parsedJson: %v\n", string(jq))
	}
}

func TestGenerate(t *testing.T) {
	testQuery(t, "js1", js1)
	testQuery(t, "js2", js2)
}

func TestArrayToMap(t *testing.T) {
	query := "a[0]=0&a[str]=1"
	parsedQuery, err := ParseQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	val := parsedQuery["a"]
	hash, ok := val.(map[string]any)
	if !ok {
		t.Fatal("expecting a map for \"a\"")
	}
	if hash["0"].(float64) != 0 {
		t.Errorf("wrong value %v for 0", hash["0"])
	}
	if hash["str"].(float64) != 1 {
		t.Errorf("wrong value %v for str", hash["str"])
	}
}
