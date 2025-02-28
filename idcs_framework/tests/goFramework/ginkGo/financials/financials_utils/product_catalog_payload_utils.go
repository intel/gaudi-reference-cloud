package financials_utils

import (
	"fmt"
	"regexp"
	"strings"
)

func EnrichProductPayload(rawpayload string, id string, name string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<id>>", id, 1)
	return enriched_payload
}

func GetServiceType(value string) []string {
	re := regexp.MustCompile(`(?i)\b\w*AsAService\b`)
	matches := re.FindAllString(value, -1)
	fmt.Println("Service types...", matches)
	return matches
}

func GetInstanceGroupSize(value string) string {
	re := regexp.MustCompile(`instanceGroupSize\s*==\s*"(\d+)"`)
	matches := re.FindStringSubmatch(value)
	fmt.Println("GROUPSIZE: ", matches)
	if len(matches) >= 1 {
		fmt.Println("InstanceGroupSize...", matches[1])
		return matches[1]
	}
	return "1"
}
