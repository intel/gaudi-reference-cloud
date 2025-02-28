package financials

// Util Functions

import (
	"github.com/nsf/jsondiff"
)

func Validate_Get_Response(data []byte, jsonValidateData string, jsonResponseData string) bool {
	opts := jsondiff.DefaultConsoleOptions()
	result, _ := jsondiff.Compare([]byte(jsonResponseData), []byte(jsonValidateData), &opts)
	return result == jsondiff.FullMatch
}
