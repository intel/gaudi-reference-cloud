package financials

// Util Functions

import (
	"time"

	"github.com/nsf/jsondiff"
)

type GetProductsResponse struct {
	Products []struct {
		Created     time.Time `json:"created"`
		Description string    `json:"description"`
		Eccn        string    `json:"eccn"`
		FamilyID    string    `json:"familyId"`
		ID          string    `json:"id"`
		MatchExpr   string    `json:"matchExpr"`
		Metadata    struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"metadata"`
		Name  string `json:"name"`
		Pcq   string `json:"pcq"`
		Rates []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
		VendorID string `json:"vendorId"`
	} `json:"products"`
}

type GetVendorsResponse struct {
	Vendors []struct {
		Created     time.Time `json:"created"`
		Description string    `json:"description"`
		Families    []struct {
			Created     time.Time `json:"created"`
			Description string    `json:"description"`
			ID          string    `json:"id"`
			Name        string    `json:"name"`
		} `json:"families"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"vendors"`
}

func Validate_Get_Response(data []byte, jsonValidateData string, jsonResponseData string) bool {
	opts := jsondiff.DefaultConsoleOptions()
	result, _ := jsondiff.Compare([]byte(jsonResponseData), []byte(jsonValidateData), &opts)
	return result == jsondiff.SupersetMatch
}
