package authorization

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/grpc_client"

	"github.com/tidwall/gjson"
)

func GetCoupons(host string) bool {
	fmt.Println("Host 2 ", host)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequestTLS("", "-insecure", host, GET_COUPONS_ENDPOINT)
	fmt.Println("Json ", sejsonStr)
	fmt.Println("Json out", outStr)
	if outStr != "" {
		fmt.Println("Failed to read coupons", sejsonStr)
		return false
	}
	result := gjson.Parse(sejsonStr)
	fmt.Println("Get response is: ", result)
	return true
}

func CreateCoupon(host string, amount int, creator string, expires string, isStandard bool, numUses int, start string) bool {
	coupon_payload := fmt.Sprintf(`{
		"amount": %d,
		"creator":  "%s",
		"expires": "%s",
		"isStandard": %t,
		"numUses": %d,
		"start": "%s"
	}`, amount, creator, expires, isStandard, numUses, start)
	data, _ := json.Marshal(coupon_payload)
	jsonPayload := string(data)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequestTLS(jsonPayload, "-insecure", host, CREATE_COUPON_ENDPOINT)
	if outStr != "" {
		fmt.Println("Failed to create coupon ", sejsonStr)
		return false
	}
	result := gjson.Parse(sejsonStr)
	fmt.Println("Post response is: ", result)
	return true
}

func CreateMeteringRecord(host string, amount int, creator string, expires string, isStandard bool, numUses int, start string) bool {
	coupon_payload := fmt.Sprintf(`{
		"amount": %d,
		"creator":  "%s",
		"expires": "%s",
		"isStandard": %t,
		"numUses": %d,
		"start": "%s"
	}`, amount, creator, expires, isStandard, numUses, start)
	data, _ := json.Marshal(coupon_payload)
	jsonPayload := string(data)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequestTLS(jsonPayload, "-insecure", host, CREATE_COUPON_ENDPOINT)
	if outStr != "" {
		fmt.Println("Failed to create coupon ", sejsonStr)
		return false
	}
	result := gjson.Parse(sejsonStr)
	fmt.Println("Post response is: ", result)
	return true
}
