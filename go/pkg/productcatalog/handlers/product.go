// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/apis/private.cloud/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

var whitelistEnabled, _ = strconv.ParseBool(os.Getenv("WHITELIST_PRODUCTS_ENABLED"))

type ProductCatalogService struct {
	pb.UnimplementedProductCatalogServiceServer
	restClient         *rest.RESTClient
	cloudAccountClient pb.CloudAccountServiceClient
	dbClient           *sql.DB
}

func NewProductCatalogService(restClient *rest.RESTClient, cloudAccountClient pb.CloudAccountServiceClient, db *sql.DB) (*ProductCatalogService, error) {
	if restClient == nil {
		return nil, fmt.Errorf("k8s client is required")
	}

	if cloudAccountClient == nil {
		return nil, fmt.Errorf("cloudaccount client is required")
	}

	return &ProductCatalogService{
		restClient:         restClient,
		cloudAccountClient: cloudAccountClient,
		dbClient:           db,
	}, nil
}

func (srv *ProductCatalogService) AdminRead(ctx context.Context, filter *pb.ProductFilter) (*pb.ProductResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.AdminRead").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	pList := v1alpha1.ProductList{}

	// Get products from the catalog crd
	err := getProductsFromCatalog(ctx, &pList, srv.restClient)

	if err != nil {
		logger.Error(err, "error reading productList crd", "pList", pList, "context", "getProductsFromCatalog")
		return nil, fmt.Errorf("error reading productList crd: %v", err)
	} else {
		if len(pList.Items) == 0 {
			logger.Info("no products found to retrieve for the given filter criteria", "filter", filter)
		}
	}

	// Filter products based on region and input filter for admin
	var accountType *pb.AccountType = nil
	if ok := filter.AccountType != nil; ok {
		accountType = filter.AccountType
	}
	filterProductResponse(ctx, &pList, filter, accountType, true)

	sort.Slice(pList.Items, func(i, j int) bool {
		return pList.Items[j].ObjectMeta.Name < pList.Items[i].ObjectMeta.Name
	})
	response, err := marshalProductResponse(pList)
	if err != nil {
		logger.Error(err, "error creating productList", "pList", pList, "context", "marshalProductResponse")
		return nil, fmt.Errorf("error creating productList: %v", err)
	}
	defer logger.Info("END")

	return response, nil
}

func (srv *ProductCatalogService) UserRead(ctx context.Context, filter *pb.ProductUserFilter) (*pb.ProductResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.UserRead").WithValues("cloudAccountId", filter.GetCloudaccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	pList := v1alpha1.ProductList{}

	// Validate the filter
	if err := validateProductUserFilter(filter); err != nil {
		logger.Error(err, "error: invalid filter")
		return nil, err
	}

	// Get products from the catalog crd
	err := getProductsFromCatalog(ctx, &pList, srv.restClient)
	if err != nil {
		logger.Error(err, "error reading productList crd", "pList", pList, "context", "getProductsFromCatalog")
		return nil, fmt.Errorf("error reading productList crd: %v", err)
	} else {
		if len(pList.Items) == 0 {
			logger.Info("no products found to retrieve for the given filter criteria", "filter", filter)
		}
	}

	cloudAccountId := filter.GetCloudaccountId()
	cloudAcct, err := srv.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "invalid cloud account", "cloudAccountId", cloudAccountId, "context", "getById")
		return &pb.ProductResponse{}, fmt.Errorf("invalid cloud account : %v", err)
	}
	// Check if ProductFilter is nil and handle it
	if filter.GetProductFilter() == nil {
		logger.Info("productFilter is not provided; using an empty filter")
		productFilter := &pb.ProductFilter{}
		filter.ProductFilter = productFilter
	}

	// Filter products based on region and input filter for user
	filterProductResponse(ctx, &pList, filter.GetProductFilter(), &cloudAcct.Type, false)

	if whitelistEnabled {
		// Filter products based on access granted (whitelisted cloudaccounts)
		err = filterAccessResponse(ctx, &pList, filter.GetCloudaccountId(), srv.dbClient)
		if err != nil {
			logger.Error(err, "error reading accessList db", "cloudAccountId", cloudAccountId, "pList", pList, "context", "filterAccessResponse")
			return nil, fmt.Errorf("error reading accessList db: %v", err)
		}
	}
	sort.Slice(pList.Items, func(i, j int) bool {
		return pList.Items[j].ObjectMeta.Name < pList.Items[i].ObjectMeta.Name
	})
	response, err := marshalProductResponse(pList)
	if err != nil {
		logger.Error(err, "error creating productList", "cloudAccountId", cloudAccountId, "pList", pList, "context", "marshalProductResponse")
		return nil, fmt.Errorf("error creating productList: %v", err)
	}
	defer logger.Info("END")

	return response, nil
}

func (srv *ProductCatalogService) UserReadExternal(ctx context.Context, filter *pb.ProductUserFilter) (*pb.ProductResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.UserReadExternal").WithValues("cloudAccountId", filter.GetCloudaccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	response, err := srv.UserRead(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, product := range response.Products {
		if nodesCountStr, exists := product.Metadata["nodesCount"]; exists {
			nodesCountInt, err := strconv.Atoi(nodesCountStr)
			if err != nil {
				logger.Error(err, "error parsing nodesCount", "cloudAccountId", filter.GetCloudaccountId(), "context", "UserReadExternal")
				return nil, fmt.Errorf("error parsing nodesCount: %v", err)
			}
			if nodesCountInt > 1 {
				for i, rate := range product.Rates {
					rateFloat, err := strconv.ParseFloat(rate.Rate, 64)
					if err != nil {
						logger.Error(err, "error parsing rate", "cloudAccountId", filter.GetCloudaccountId(), "context", "UserReadExternal")
						return nil, fmt.Errorf("error parsing rate: %v", err)
					}
					totalRate := rateFloat * float64(nodesCountInt)
					newRate := strconv.FormatFloat(totalRate, 'f', -1, 64)
					product.Rates[i].Rate = newRate
				}
			}
		}
	}
	return response, nil
}

func getProductsFromCatalog(ctx context.Context, pList *v1alpha1.ProductList, restClient *rest.RESTClient) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.getProductsFromCatalog").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// This function should return true if the error is transient.
	// We can check for specific error types here. For now, we're retrying on all errors.
	isRetryable := func(err error) bool { return true }

	return retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		return restClient.
			Get().
			Resource("products").
			Do(ctx).
			Into(pList)
	})
}

func marshalProductResponse(pList v1alpha1.ProductList) (*pb.ProductResponse, error) {
	res := pb.ProductResponse{}

	for _, p := range pList.Items {
		ts, err := ptypes.TimestampProto(p.CreationTimestamp.Time)
		if err != nil {
			return nil, err
		}

		var access string
		for _, meta := range p.Spec.Metadata {
			if meta.Key == "access" {
				access = meta.Value
				break
			}
		}

		pr := pb.Product{
			Id:          p.Spec.ID,
			VendorId:    p.Spec.VendorID,
			FamilyId:    p.Spec.FamilyID,
			Description: p.Spec.Description,
			MatchExpr:   p.Spec.MatchExpr,
			Name:        p.ObjectMeta.Name,
			Eccn:        p.Spec.ECCN,
			Pcq:         p.Spec.PCQ,
			Access:      access,
			Created:     ts,
			Status:      getProductStatus(p),
		}
		rates := []*pb.Rate{}

		for _, r := range p.Spec.Rates {
			rb := pb.Rate{
				AccountType: marshalAccountTypeFromCRToPB(r.AccountType),
				Unit:        marshalRateUnit(r.Unit),
				UsageExpr:   r.UsageExpr,
				Rate:        r.Rate,
			}
			rates = append(rates, &rb)
		}
		pr.Rates = rates

		props := map[string]string{}
		productMeta := p.Spec.Metadata
		for _, pm := range productMeta {
			props[pm.Key] = pm.Value
		}
		pr.Metadata = props

		res.Products = append(res.Products, &pr)
	}
	return &res, nil
}

func filterProductResponse(ctx context.Context, pList *v1alpha1.ProductList, filter *pb.ProductFilter, acctType *pb.AccountType, admin bool) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.filterProductResponse").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	insertIdx := 0
	for _, prod := range pList.Items {
		skip := false
		mdMap := map[string]string{}

		// create a map of all product metadata
		if prod.Spec.Metadata != nil {
			mds := prod.Spec.Metadata
			for _, md := range mds {
				mdMap[md.Key] = md.Value
			}
		}

		// first check if product has any disable flag by account type
		if val, found := mdMap["disableForAccountTypes"]; found {
			if val != "" {
				disabledAccounts := strings.Split(val, ",")
				for _, dAcc := range disabledAccounts {
					if acctType != nil && *acctType == mapAccountTypeToPb(dAcc) {
						skip = true
						break
					}
				}
			}
		}

		// product disabled for acctType
		if skip {
			continue
		}

		// check the region in filter for backward compatibility
		if filter.GetRegion() == "" {
			skip = false
		} else {
			// check if filtering for user
			if !admin {
				// check if product has any enabled regions
				if val, found := mdMap["region"]; found {
					if val != "" {
						enabledRegions := strings.Split(val, ",")
						if enabledRegions[0] == "global" {
							// check if the product is globally enabled
							skip = false
						} else {

							rgMap := map[string]bool{}
							for _, rg := range enabledRegions {
								rgMap[rg] = true
							}
							// check if the product is enabled for the given region
							if _, found := rgMap[filter.GetRegion()]; !found {
								skip = true
							}
						}
					} else {
						skip = true
					}
				}
			} else { //filtering for admin
				skip = false
			}
		}

		if skip {
			continue
		}
		filteredRates := []v1alpha1.ProductRate{}
		if filter.Id != nil {
			if !strings.EqualFold(filter.GetId(), prod.Spec.ID) {
				logger.Info("product skipped due to ID mismatch", "productId", prod.Spec.ID, "filterId", filter.GetId())
				skip = true
			}
		}
		if filter.FamilyId != nil {
			if !strings.EqualFold(filter.GetFamilyId(), prod.Spec.FamilyID) {
				logger.Info("product skipped due to familyId mismatch", "productId", prod.Spec.ID, "filterFamilyId", filter.GetFamilyId())
				skip = true
			}
		}
		if filter.MatchExpr != nil {
			if !strings.EqualFold(filter.GetMatchExpr(), prod.Spec.MatchExpr) {
				logger.Info("product skipped due to matchExpr mismatch", "productId", prod.Spec.ID, "filterMatchExpr", filter.GetMatchExpr())
				skip = true
			}
		}
		if filter.VendorId != nil {
			if !strings.EqualFold(filter.GetVendorId(), prod.Spec.VendorID) {
				logger.Info("product skipped due to vendorID mismatch", "productId", prod.Spec.ID, "filterVendorId", filter.GetVendorId())
				skip = true
			}
		}
		if filter.Name != nil {
			if !strings.EqualFold(filter.GetName(), prod.ObjectMeta.Name) {
				logger.Info("product skipped due to Name mismatch", "productId", prod.Spec.ID, "filterName", filter.GetName())
				skip = true
			}
		}
		if filter.Eccn != nil {
			if !strings.EqualFold(filter.GetEccn(), prod.Spec.ECCN) {
				logger.Info("product skipped due to ECCN mismatch", "productId", prod.Spec.ID, "filterECCN", filter.GetEccn())
				skip = true
			}
		}
		if filter.Pcq != nil {
			if !strings.EqualFold(filter.GetPcq(), prod.Spec.PCQ) {
				logger.Info("product skipped due to PCQ mismatch", "productId", prod.Spec.ID, "filterPCQ", filter.GetPcq())
				skip = true
			}
		}
		if filter.Metadata != nil {
			for k, v := range filter.Metadata {
				if mapV, found := mdMap[k]; found {
					if !strings.EqualFold(v, mapV) {
						logger.Info("product skipped due to metadata value mismatch", "productId", prod.Spec.ID, "metadataKey", k, "expectedValue", v, "actualValue", mapV)
						skip = true
					}
				} else {
					logger.Info("product skipped due to missing metadata key", "productId", prod.Spec.ID, "missingMetadataKey", k, "expectedValue", v)
					skip = true
				}
			}
		}

		// this filter could be empty because user didn't specified it
		// or user specified it to be empty
		if acctType != nil && *acctType != pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED {
			// For enterprise pending user, show premium rates
			if *acctType == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING {
				*acctType = pb.AccountType_ACCOUNT_TYPE_PREMIUM
			}

			for _, r := range prod.Spec.Rates {
				if marshalAccountTypeFromCRToPB(r.AccountType) == *acctType {
					filteredRates = append(filteredRates, r)
				}
			}
			if len(filteredRates) == 0 {
				logger.Info("no rates specified for the product tier", "accountType", *acctType, "product", prod.ObjectMeta.Name)
			}
			// Do not update skip flag
		}

		if !skip {
			prod.Spec.Rates = filteredRates
			pList.Items[insertIdx] = prod
			insertIdx++
		}
	}
	pList.Items = pList.Items[:insertIdx]
}

func filterAccessResponse(ctx context.Context, pList *v1alpha1.ProductList, cloudaccountID string, db *sql.DB) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.filterAccessResponse").WithValues("cloudAccountId", cloudaccountID).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// Lookup cloud_accounts_product_access for permitted products
	accessMap, err := getProductAccessList(ctx, db, cloudaccountID)
	if err != nil {
		logger.Error(err, "error getting product access list", "cloudAccountId", cloudaccountID, "context", "getProductAccessList")
		return err
	}

	// Filter out the input list agaist the access map
	insertIdx := 0
	for _, prod := range pList.Items {

		mdMap := map[string]string{}
		// create a map of all product metadata
		if prod.Spec.Metadata != nil {
			mds := prod.Spec.Metadata
			for _, md := range mds {
				mdMap[md.Key] = md.Value
			}
		}

		// first check if product has any access flag
		if val, found := mdMap["access"]; found {
			if val == "controlled" {
				// check the product against the whitelist of the cloudaccount
				if found := accessMap[prod.Spec.VendorID+prod.Spec.FamilyID+prod.Spec.ID]; found {
					pList.Items[insertIdx] = prod
					insertIdx++
				}
			} else {
				pList.Items[insertIdx] = prod
				insertIdx++
			}
		}
	}
	pList.Items = pList.Items[:insertIdx]
	return nil
}

func getProductAccessList(ctx context.Context, db *sql.DB, arg string) (map[string]bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.getProductAccessList").WithValues("cloudAccountId", arg).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// Create a map from the access list for a cloudaccount
	acl := map[string]bool{}

	query := `
		SELECT cloudaccount_id, product_id, family_id, vendor_id, admin_name 
		from cloud_accounts_product_access 
		WHERE cloudaccount_id = $1`

	rows, err := db.QueryContext(ctx, query, arg)
	if err != nil {
		logger.Error(err, "error getting product access list by cloud account")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var obj pb.ProductAccessRequest
		if err := rows.Scan(&obj.CloudaccountId, &obj.ProductId, &obj.FamilyId, &obj.VendorId, &obj.AdminName); err != nil {
			logger.Error(err, "error observed while fetching Cloud Account", "arg", arg)
			return acl, err
		}
		logger.Info("result", "cloudAccountId", obj.CloudaccountId, "vendorId", obj.VendorId, "familyId", obj.FamilyId, "productId", obj.ProductId, "adminName", obj.AdminName)
		acl[obj.VendorId+obj.FamilyId+obj.ProductId] = true
	}
	return acl, nil
}

func getProductStatus(p v1alpha1.Product) string {
	if !reflect.ValueOf(p.Status).IsZero() && p.Status.State != "" {
		return string(p.Status.State)
	}
	return string(v1alpha1.ProductStateUndetermined)
}

func (srv *ProductCatalogService) SetStatus(ctx context.Context, in *pb.SetProductStatusRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.SetStatus").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	pList := v1alpha1.ProductList{}
	err := srv.restClient.
		Get().
		Resource("products").
		Do(ctx).
		Into(&pList)

	if err != nil {
		return nil, fmt.Errorf("error reading productList crd: %v", err)
	}
	for _, ps := range in.GetStatus() {
		filter := pb.ProductFilter{}
		filter.FamilyId = &ps.FamilyId
		filter.Id = &ps.ProductId
		filter.VendorId = &ps.VendorId
		filterProductResponse(ctx, &pList, &filter, filter.AccountType, true)
		if len(pList.Items) == 0 {
			logger.Info("no matching product found",
				"productId", ps.ProductId,
				"vendorId", ps.VendorId,
				"familyId", ps.FamilyId,
			)
			return &emptypb.Empty{}, fmt.Errorf("no matching products found")
		}
		if len(pList.Items) > 1 {
			logger.Info("ambigious input filters for products",
				"productId", ps.ProductId,
				"vendorId", ps.VendorId,
				"familyId", ps.FamilyId,
			)
			return &emptypb.Empty{}, fmt.Errorf("ambigious input filters for products")
		}
		product := pList.Items[0]
		product.Status = v1alpha1.ProductStatus{
			State: getProductStateMapping(ps.GetStatus()),
		}
		prodBuf, err := json.Marshal(product)
		if err != nil {
			logger.Error(err, "error in marshalling product", "product", product.Spec.ID, "context", "Marshal")
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "error in marshalling product")
		}
		_, err = srv.restClient.Put().Namespace("default").Resource("products").Name(product.ObjectMeta.Name).SubResource("status").Body(bytes.NewReader(prodBuf)).DoRaw(ctx)
		if err != nil {
			logger.Error(err, "error updating product status", "product", product.Spec.ID, "context", "DoRaw")
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "error updating product status")
		}
	}
	return &emptypb.Empty{}, nil
}

func (srv *ProductCatalogService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductCatalogService.Ping").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("Ping")
	return &emptypb.Empty{}, nil
}

func getProductStateMapping(pbState pb.ProductStatus) v1alpha1.ProductState {

	if pbState == pb.ProductStatus_PRODUCT_STATUS_ERROR {
		return v1alpha1.ProductStateError
	} else if pbState == pb.ProductStatus_PRODUCT_STATUS_PROVISIONING {
		return v1alpha1.ProductStateProvisioning
	} else if pbState == pb.ProductStatus_PRODUCT_STATUS_READY {
		return v1alpha1.ProductStateReady
	} else if pbState == pb.ProductStatus_PRODUCT_STATUS_UNSPECIFIED {
		return v1alpha1.ProductStateUndetermined
	}
	return v1alpha1.ProductStateUndetermined
}

func mapAccountTypeToPb(accountStr string) pb.AccountType {
	switch accountStr {
	case "standard":
		return pb.AccountType_ACCOUNT_TYPE_STANDARD
	case "premium":
		return pb.AccountType_ACCOUNT_TYPE_PREMIUM
	case "enterprise":
		return pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
	case "intel":
		return pb.AccountType_ACCOUNT_TYPE_INTEL
	default:
		return pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func marshalAccountTypeFromCRToPB(acc v1alpha1.IDCAccountType) pb.AccountType {
	switch acc {
	case v1alpha1.EnterpriseAccountType:
		return pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
	case v1alpha1.StandardAccountType:
		return pb.AccountType_ACCOUNT_TYPE_STANDARD
	case v1alpha1.PremiumAccountType:
		return pb.AccountType_ACCOUNT_TYPE_PREMIUM
	case v1alpha1.IntelAccountType:
		return pb.AccountType_ACCOUNT_TYPE_INTEL
	}
	return pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED
}

func marshalRateUnit(unit string) pb.RateUnit {
	switch unit {
	case "dollarsPerMinute":
		return pb.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE
	case "dollarsPerTBPerHour":
		return pb.RateUnit_RATE_UNIT_DOLLARS_PER_TB_PER_HOUR
	case "dollarsPerInference":
		return pb.RateUnit_RATE_UNIT_DOLLARS_PER_INFERENCE
	case "dollarsPerMillionTokens":
		return pb.RateUnit_RATE_UNIT_DOLLARS_PER_MILLION_TOKENS
	}
	return pb.RateUnit_RATE_UNIT_UNSPECIFIED
}
