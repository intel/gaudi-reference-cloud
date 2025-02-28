// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tradecheckintel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

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

func (gts *GTSclient) CreateProduct(ctx context.Context, product Product) error {
	logger := log.FromContext(ctx)
	retries := 3

	token, err := gts.apgclient.GetCurrentToken(ctx)
	if err != nil {
		logger.Info("error encountered in getting token", "error", err)
		return fmt.Errorf("error encountered in getting Apigee token")
	}

	for try := 1; try <= retries; try++ {

		resp, err := gts.client.R().
			SetAuthToken(token).
			SetContentLength(true).
			SetBody(product).
			Post(gts.cfg.createProductURL)

		if err != nil {
			if try == retries {
				logger.Info("PUT", "trial", try, "error in calling create product", err)
				return err
			}
			logger.Info("PUT", "trial", try, "error in calling create product", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}

		err = processResponse(ctx, resp)
		if err != nil {
			if try == retries {
				logger.Info("PUT", "trial", try, "error", err)
				return err
			}
			logger.Info("PUT", "trial", try, "error", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	logger.Info("Product Created successfully", "Product Number", product.ProductHeader.ProductNumber)
	return nil
}

func processResponse(ctx context.Context, resp *resty.Response) error {
	logger := log.FromContext(ctx)
	response := CreateProductResponse{StatusCode: "", StatusText: ""}
	if resp.StatusCode() != http.StatusOK {
		logger.Info("error in response from gts", "response status code", resp.StatusCode(), "response body", resp.Body())
		return fmt.Errorf("error in creating the product, response status code: %v", resp.StatusCode())
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		logger.Info("error unmarshalling response from gts", "response status", resp.Status(), "response body", resp.Body())
		return fmt.Errorf("unmarshalling Create Product Response failed")
	}

	if response.StatusCode != "P" {
		logger.Info("GTS Create Product API returned error", "Status Code", response.StatusCode, "Status Text", response.StatusText)
		return fmt.Errorf("GTS Create Product API returned Status Code %v", response.StatusCode)
	}

	return nil
}

func (gts *GTSclient) CreateOrder(ctx context.Context, order Order) (CreateOrderResponse, error) {
	logger := log.FromContext(ctx)
	response := CreateOrderResponse{}
	token, err := gts.apgclient.GetCurrentToken(ctx)
	if err != nil {
		logger.Info("error encountered in getting token", "error", err)
		return response, fmt.Errorf("error encountered in getting Apigee token")
	}

	ts1 := time.Now()
	resp, err := gts.client.R().
		SetAuthToken(token).
		SetContentLength(true).
		SetBody(order).
		Post(gts.cfg.createOrderURL)

	ts2 := time.Now()
	logger.Info("time to make gts createOrder", "total time", ts2.Sub(ts1).Seconds())

	if err != nil {
		logger.Info("error in calling create order", "error", err)
		return response, fmt.Errorf("error in GTS Create Order API")
	}

	if resp.StatusCode() != http.StatusOK {
		// TODO : Unmarshal the error to errorResponse
		logger.Info("error in response from gts", "response status code", resp.StatusCode(), "response body", resp.Body())
		return response, fmt.Errorf("error in creating the order, response status code: %v", resp.StatusCode())
	}
	logger.Info("gts response for create order", "raw response", string(resp.Body()))
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		logger.Info("error unmarshalling response from gts", "response body", resp.Body(), "error", err)
		return response, fmt.Errorf("unmarshalling Create Order Response failed")
	}
	return response, nil
}

func (gts *GTSclient) IsOrderValid(ctx context.Context, productId, email, personId,
	countryCode string) (bool, error) {
	ctx, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("GTSclient.IsOrderValid").Start()
	logger.Info("GTS check invoked", "productId", productId, "personId", personId, "countryCode", countryCode)
	split := strings.Split(email, "@")
	if len(split) == 2 && split[1] == "intel.com" {
		logger.Info("Skipping check for intel user account", "personId", personId)
		return true, nil
	}
	item := OrderItem{
		ProductNumber: productId,
		ItemNumber:    "1",
		Quantity:      "1",
		Value:         "1",
		Currency:      "usd",
	}
	partner := Partner{
		PartnerName:    email,
		PartnerNumber:  personId,
		PartnerType:    "ShipTo",
		PartnerCountry: countryCode,
	}
	order := Order{
		OrderHeader: OrderHeader{
			PlantCode:           "IDC1",
			LegalEntity:         "LE036",
			DocumentReferenceNo: fmt.Sprintf("%s_%d", personId, time.Now().Unix()),
			DocumentType:        "P",
			SourceSystemDetails: []SystemDetails{{LogicalSystem: "IDC"}},
			OrderItems:          []OrderItem{item},
			CancelFlag:          "",
			Partners:            []Partner{partner},
		},
	}
	resp, err := gts.CreateOrder(ctx, order)
	if err != nil {
		logger.Error(err, "Error in Creating Order")
		return false, err
	}
	logger.Info("gts response for create order", "parsed response", resp)
	if resp.Orders.OverallStatusCode != "P" {
		logger.Info("Create Order Failed: Failure in Overall Status Code")
		return false, nil
	}
	if resp.Orders.ShipToSanctionPartyListStatus != "P" {
		logger.Info("Create Order Failed: Failure in ShipTo Sanction Party List Status")
		return false, nil
	}
	if resp.Orders.ShipToEmbargoedCheckStatus != "P" {
		logger.Info("Create Order Failed: Failure in ShipTo Embargoed Check Status")
		return false, nil
	}
	for _, item := range resp.Orders.ExportLicenseCheckStatus.Items {
		if item.StatusCode != "P" {
			logger.Info("Create Order Failed: Failure in Item Status Code")
			return false, nil
		}
	}
	logger.Info("Create Order completed successfully")
	return true, nil
}

func (gts *GTSclient) ScreenBusinessPartner(ctx context.Context, screenRequest ScreenRequest) (ScreenResponse, error) {
	logger := log.FromContext(ctx).WithName("GTSclient.SceenBusinessPartner")
	response := ScreenResponse{}
	if len(screenRequest.Partners) == 0 || screenRequest.Partners[0].EnterpriseID == "" ||
		screenRequest.Partners[0].Name == "" ||
		screenRequest.Partners[0].Country == "" {
		logger.Info("invalid screening information, skipping check")
		return response, fmt.Errorf("invalid inputs")
	}

	token, err := gts.apgclient.GetCurrentToken(ctx)
	if err != nil {
		logger.Info("error encountered in getting token", "error", err)
		return response, fmt.Errorf("error encountered in getting Apigee token")
	}

	logger.Info("making screenBusinessPartner request for ", "enterpriseID", screenRequest.Partners[0].EnterpriseID)
	resp, err := gts.client.R().
		SetAuthToken(token).
		SetContentLength(true).
		SetBody(screenRequest).
		Post(gts.cfg.screenBusinessPartnerURL)

	if err != nil {
		logger.Info("error in calling Screen Business Partner API", "error", err)
		return response, fmt.Errorf("error in GTS Screen Business Partner API")
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Info("error in response from gts", "response status code", resp.StatusCode(), "response body", resp.Body())
		return response, fmt.Errorf("error in screening the Business Partner, response status code: %v", resp.StatusCode())
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		logger.Info("error unmarshalling response from gts", "response status", resp.Status(), "response body", resp.Body())
		return response, fmt.Errorf("unmarshalling Screen Business Partner Response failed")
	}
	return response, nil
}
