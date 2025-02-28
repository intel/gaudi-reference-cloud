// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tradecheck

type Token struct {
	tokenString string
	exp         int64
}

type getTokenResponse struct {
	Access_token string
}

type CreateProductResponse struct {
	StatusCode string `json:"StatusCode"`
	StatusText string `json:"StatusText"`
}

type Orders struct {
	OverallStatusCode             string                   `json:"OverallStatusCode"`
	OverallStatusText             string                   `json:"OverallStatusText"`
	DocumentType                  string                   `json:"DocumentType"`
	DocumentReference             string                   `json:"DocumentReference"`
	ExportLicenseCheckStatus      ExportLicenseCheckStatus `json:"ExportLicenseCheckStatus"`
	ShipToSanctionPartyListStatus string                   `json:"ShipToSanctionPartyListStatus"`
	ShipToEmbargoedCheckStatus    string                   `json:"ShipToEmbargoedCheckStatus"`
}

type CreateOrderResponse struct {
	Orders Orders `json:"Orders"`
}

type ExportLicenseCheckStatus struct {
	Items []Item `json:"Items"`
}

type Item struct {
	LineItemNumber string `json:"LineItemNumber"`
	ProductNumber  string `json:"ProductNumber"`
	StatusCode     string `json:"StatusCode"`
	StatusText     string `json:"StatusText"`
}

type ProductHeader struct {
	ProductNumber      string `json:"ProductNumber"`
	ProductDescription string `json:"ProductDescription"`
	LogicalSystem      string `json:"LogicalSystem,omitempty"`
	PCQ_ID             string `json:"PCQ_ID"`
	ECCN               string `json:"ECCN"`
}

type Product struct {
	ProductHeader ProductHeader `json:"ProductHeader"`
}

type OrderHeader struct {
	PlantCode           string          `json:"PlantCode"`
	LegalEntity         string          `json:"LegalEntity"`
	DocumentReferenceNo string          `json:"DocumentReferenceNo"`
	DocumentType        string          `json:"DocumentType"`
	SourceSystemDetails []SystemDetails `json:"SourceSystemDetails"`
	OrderItems          []OrderItem     `json:"OrderItems"`
	CancelFlag          string          `json:"CancelFlag"`
	Partners            []Partner       `json:"Partners"`
}

type Order struct {
	OrderHeader OrderHeader `json:"OrderHeader"`
}

type SystemDetails struct {
	LogicalSystem string `json:"LogicalSystem"`
}

type OrderItem struct {
	ProductNumber string `json:"ProductNumber"`
	ItemNumber    string `json:"ItemNumber"`
	Quantity      string `json:"Quantity"`
	Value         string `json:"Value"`
	Currency      string `json:"Currency"`
}

type Partner struct {
	PartnerName    string `json:"PartnerName"`
	PartnerNumber  string `json:"PartnerNumber"`
	PartnerType    string `json:"PartnerType"`
	PartnerCountry string `json:"PartnerCountry"`
}

type ScreenRequest struct {
	Partners []BusinessPartnerRequest `json:"BusinessPartner"`
}

type BusinessPartnerRequest struct {
	EnterpriseID string `json:"EnterpriseID"`
	Name         string `json:"Name1"`
	Country      string `json:"Country"`
}

type ScreenResponse struct {
	BusinessPartner BusinessPartnerResponse `json:"BusinessPartner"`
}

type BusinessPartnerResponse struct {
	Status Status `json:"Status"`
}

type Status struct {
	// BusinessPartner string `json:"BusinessPartner"`
	SPLStatus     string `json:"SPLStatus"`
	EmbargoStatus string `json:"EmbargoStatus"`
}
