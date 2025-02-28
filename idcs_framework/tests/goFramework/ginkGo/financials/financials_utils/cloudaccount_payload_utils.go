package financials_utils

import (
	"fmt"
	"strings"
)

func EnrichEnrollTokenPayload(rawpayload string, tid string, enterpriseId string, email string, groups string, vnet_name string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<tid>>", tid, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<enterpriseId>>", enterpriseId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<email>>", email, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<groups>>", groups, 1)
	return enriched_payload
}

func EnrichEnrollPayload(rawpayload string, premium string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<premium>>", premium, 1)
	fmt.Println("enriched_payload", enriched_payload)
	enriched_payload = `{"premium":false}`
	return enriched_payload
}

// func EnrichDisableCouponPayload(rawpayload string, code string) string {
// 	enriched_payload := rawpayload
// 	enriched_payload = strings.Replace(enriched_payload, "<<code>>", code, 1)
// 	fmt.Println("enriched_payload", enriched_payload)
// 	return enriched_payload
// }

// func EnrichRedeemCouponPayload(rawpayload string, code string, cloudAccountId string) string {
// 	enriched_payload := rawpayload
// 	enriched_payload = strings.Replace(enriched_payload, "<<code>>", code, 1)
// 	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
// 	return enriched_payload
// }
