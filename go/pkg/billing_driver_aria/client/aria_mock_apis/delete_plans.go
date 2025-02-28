// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
)

// Delete Plan Client
// This function has only client_id and auth_key as required parameters
var planNos = []int{
	//11306378,
}

//	func toJSON(m any) string {
//		js, err := json.Marshal(m)
//		if err != nil {
//			log.Fatal(err)
//		}
//		return strings.ReplaceAll(string(js), ",", ", ")
//	}
//
//	func InterfaceToSlice[T comparable](v interface{}) []T {
//		retSlice := make([]T, 0)
//		fmt.Println(reflect.TypeOf(v))
//		if reflect.TypeOf(v).Kind() == reflect.Slice {
//			s := reflect.ValueOf(v)
//			fmt.Println(s)
//			for i := 0; i < s.Len(); i++ {
//				sp := s.Interface()
//				spi := sp.(T)
//				retSlice = append(retSlice, spi)
//			}
//		}
//		fmt.Println(retSlice)
//		return retSlice
//	}
func DeletePlanClientHandler(w http.ResponseWriter, req map[string]any) {

	reqStruct := requestAriaAdminMock[*request.DeletePlans](w, req)
	if !lookupService(reqStruct.PlanNos[0], planNos) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Plan No doesn't exist or is deleted already!",
		})
		return
	}
	resp := response.DeletePlansResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		PlanNos: []int{
			planNos[len(planNos)-1],
		},
	}
	planName = planName[:len(planName)-1]
	planNos = planNos[:len(planNos)-1]
	MockResponseWriter(w, resp)
}
