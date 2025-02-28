package productcatalog

type StatusPayload struct {
	Error     string `json:"error"`
	FamilyId  string `json:"familyId"`
	ProductId string `json:"productId"`
	Status    string `json:"status"`
	VendorId  string `json:"vendorId"`
}

type SetStatusPayload struct {
	Status []StatusPayload `json:"status"`
}

var SETSTATUS_ENDPOINT = "proto.ProductCatalogService/SetStatus"
