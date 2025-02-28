// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// To make sure everything we log w.r.t errors is standardized.
// Standardization needed to be able to filter log messages for reacting appropriately.
const (
	ProductSyncInvalidProductFamily     = "product sync: invalid product family"
	ProductSyncFailedToCheckPlanExists  = "product sync: failed to check if plan exists"
	ProductSyncFailedToCreatePlan       = "product sync: failed to create plan"
	ProductSyncFailedToDeactivatePlan   = "product sync: failed to deactivate plan"
	ProductSyncFailedToEditPlan         = "product sync: failed to edit plan"
	ProductSyncFailedToGetUsageUnitType = "product sync: failed to get usage unit type"
	ERROR_CODE_OBJECT_DOES_NOT_EXIST    = 1013
	instanceTypeKey                     = "instanceType"
	processingTypeKey                   = "processingType"
	inferenceProcessingType             = "image"
	tokenProcessingType                 = "text"
	UsageUnitTypeUnavailable            = "usage unit type unavailable"
	UsageUnitKey                        = "usage.quantity.unit"
	MinuteUsageUnitValue                = "min"
)

type ProductController struct {
	ariaPlanClient      *client.AriaPlanClient
	ariaUsageTypeClient *client.AriaUsageTypeClient
	ariaPromoClient     *client.PromoClient
	ariaServiceClient   *client.AriaServiceClient
	ariaController      *AriaController
	productClient       billingCommon.ProductClientInterface
}

type ProductSyncError struct {
	Product *pb.Product
	Err     error
}

type UsageUnitType struct {
	minutesUsageType   *data.UsageType
	storageUsageType   *data.UsageType
	tokenUsageType     *data.UsageType
	inferenceUsageType *data.UsageType
}

var usageUnitTypes = UsageUnitType{}

var AccountPremiumTypes = []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_PREMIUM, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE}

func NewProductController(ariaClient *client.AriaClient, ariaAdminClient *client.AriaAdminClient, ariaCredentials *client.AriaCredentials, productClient billingCommon.ProductClientInterface, ariaController *AriaController) *ProductController {
	logger := log.FromContext(context.Background()).WithName("NewProductController")
	ariaPlanClient := client.NewAriaPlanClient(config.Cfg, ariaAdminClient, ariaClient, ariaCredentials)
	ariaUsageClient := client.NewAriaUsageTypeClient(ariaAdminClient, ariaCredentials)
	ariaPromoClient := client.NewPromoClient(ariaAdminClient, ariaCredentials)
	ariaServiceClient, err := client.NewAriaServiceClient(ariaAdminClient, ariaCredentials)
	if err != nil {
		logger.Error(err, "Error crating new Aria Service Client")
		return &ProductController{}
	}
	return &ProductController{ariaPlanClient: ariaPlanClient, ariaUsageTypeClient: ariaUsageClient,
		productClient: productClient, ariaController: ariaController, ariaPromoClient: ariaPromoClient,
		ariaServiceClient: ariaServiceClient}
}

func (productController *ProductController) GetPlanServices(ctx context.Context, resp *response.GetPlanDetailResponse, clientPlanId string) ([]data.PlanService, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.GetPlanServices").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	planServices := make([]data.PlanService, 0, len(resp.Services))
	var err error
	for _, service := range resp.Services {
		//Map service to rates
		// Note: This is different than design - Atanu to address.
		clientPlanServiceRates, err := productController.ariaPlanClient.GetClientPlanServiceRates(ctx, clientPlanId, service.ClientServiceId)
		if err != nil {
			return nil, err
		}

		dService := data.PlanService{
			ServiceNo:        int64(service.ServiceNo),
			ClientServiceId:  service.ClientServiceId,
			PlanServiceRates: clientPlanServiceRates.PlanServiceRates,
		}
		planServices = append(planServices, dService)
	}
	return planServices, err
}

func (productController *ProductController) GetProductClientPlanMap(ctx context.Context, clientPlansDetail []data.AllClientPlanDtl, products []*pb.Product) (map[string]*client.ProductClientPlan, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.GetProductClientPlanMap").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("getting product client plan map")

	planMap := client.CreateProductClientPlanMap(ctx, clientPlansDetail)
	for _, product := range products {
		clientPlanId := client.GetPlanClientId(product.GetId())
		logger.Info("client plan", "clientPlanId", clientPlanId)
		if clientPlanDetail, ok := planMap[clientPlanId]; ok {
			logger.V(1).Info("client plan detail", "clientPlanDetail", clientPlanDetail, "productId", product.Id, "productName", product.Name)
			clientPlanDetail.UpdateRequired = true
			clientPlanDetail.Product = product
			clientPlanDetail.IsActive = true
		} else {
			planDetailResp, err := productController.ariaController.GetClientPlanDetails(ctx, clientPlanId)
			logger.Info("client plan detail response", "planDetailResp", planDetailResp)
			if err != nil || planDetailResp.GetErrorCode() != 0 {
				planMap[clientPlanId] = &client.ProductClientPlan{UpdateRequired: false, IsActive: true, ClientPlanDetail: data.AllClientPlanDtl{}, Product: product}
			} else if planDetailResp.GetErrorCode() == 0 && planDetailResp.ActiveInd == 0 && strings.EqualFold(planDetailResp.ClientPlanId, clientPlanId) {
				planServices, err := productController.GetPlanServices(ctx, planDetailResp, clientPlanId)
				if err != nil {
					return nil, err
				}
				clientPlanDetailData := client.MapResponseToClientPlanDetail(planDetailResp, planServices)
				planMap[clientPlanId] = &client.ProductClientPlan{UpdateRequired: true, IsActive: true, ClientPlanDetail: *clientPlanDetailData, Product: product}
			}
		}
	}
	return planMap, nil
}

// This pattern to repeat for all resources that are a part of product catalog sync:
// Check if a resource has been created - If fails to check, don't act on the resource.
// If resource has not been created - create it - If fails to create, don't act on the resource further.
// If resource has been created - get values for the resource - If fails to get values, don't act on the resource further.
// If values are not as expected - update the resource - If fails to update, don't act on the resource further.
// If failed to update - don't act on the resource further.
// For every failure, build a correlation between the resource and cause of failure.
// For all such correlations of resource and cause of failure, notify appropriately on the resource and cause of failure.
// If such resource to cause of failure correlation needs to be stored as a part of the resource representation:
// In this case for product, every failure to sync, needs to be stored as a part of product representation in product catalog -
// Update the representation of the product with all failures associated with syncing the product.
// Such representation of a product with individual failures to sync, needs to be addressed by Ops.
// However, retrials will be performed when the scheduled sync happens again or when ops forces a sync.
// When syncing again, we could do dedicated steps based on a previously recorded set of failures with sync - To be decided.
func (productController *ProductController) SyncProducts(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.SyncProducts").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("syncing products")

	// InitAria ensures the master plan, the promo code, and the plan sets
	// are configured
	err := productController.ariaController.InitAria(ctx)
	if err != nil {
		logger.Error(err, "error initializing Aria")
		return err
	}

	vendors, err := productController.productClient.GetProductCatalogVendors(ctx)
	// fail fast if cannot get vendors.
	if err != nil {
		logger.Error(err, "failed to get vendors from product catalog")
		return err
	}
	products, err := productController.productClient.GetProductCatalogProductsForAccountTypes(ctx, AccountPremiumTypes)
	// fail fast if cannot get products
	if err != nil {
		logger.Error(err, "failed to get products from product catalog")
		return err
	}
	if len(vendors) == 0 && len(products) == 0 {
		logger.Info("product catalog data ", "products", products, "vendors", vendors)
		return nil
	}
	validProducts, invalidProducts, err := billingCommon.ValidateProductsForProductFamily(vendors, products)
	// fail fast if cannot validate products
	if err != nil {
		logger.Error(err, "failed to validate products")
		return err
	}
	var productSyncErrors []ProductSyncError
	// populate errors for invalid products
	for _, invalidProduct := range invalidProducts {
		productSyncErrors = append(productSyncErrors, ProductSyncError{
			Product: invalidProduct,
			Err:     errors.New(ProductSyncInvalidProductFamily)})
	}

	// This vendor map is populated as a part of checking for validity as well.
	// To avoid repopulating here, need to ask for a API which returns a product family for a product or
	// product needs to have a representation of product family ideally.
	vendorMap := make(map[string]*pb.ProductFamily)
	for _, vendor := range vendors {
		for _, productFamily := range vendor.GetFamilies() {
			vendorMap[productFamily.GetId()] = productFamily
		}
	}

	createdProducts := []string{}

	// Iterate through valid products
	productsStatus := make([]*pb.SetProductStatus, 0, len(validProducts))
	productsClientPlanMap := map[string]*client.ProductClientPlan{}
	clientPlanDtls, err := productController.ariaPlanClient.GetAllClientPlansForPromoCode(ctx, client.GetPromoCode())
	if err != nil {
		productSyncErrors = productController.SyncError(ctx, validProducts, err, productSyncErrors)
	} else {
		productsClientPlanMap, err = productController.GetProductClientPlanMap(ctx, clientPlanDtls.AllClientPlanDtls, validProducts)
		if err != nil {
			productSyncErrors = productController.SyncError(ctx, validProducts, err, productSyncErrors)
		}
	}
	productStatusChange := false
	for clientPlanId, productClientPlan := range productsClientPlanMap {
		logger.Info("client plan id", "clientPlanId", clientPlanId)
		product := productClientPlan.Product
		usageType, err := productController.GetUsageUnitType(ctx, product)
		if err != nil {
			productSyncErrors = append(productSyncErrors, ProductSyncError{Product: product, Err: errors.New(ProductSyncFailedToGetUsageUnitType)})
			continue
		}
		if !productClientPlan.UpdateRequired && productClientPlan.IsActive {
			logger.Info("client plan does not exist, creating", "clientPlanId", clientPlanId)
			_, err := productController.ariaPlanClient.CreatePlan(ctx, product, vendorMap[product.GetFamilyId()], usageType)
			if err != nil {
				productSyncErrors = append(productSyncErrors, ProductSyncError{Product: product, Err: errors.New(ProductSyncFailedToCreatePlan)})
				continue
			}
			productStatusChange = true
			createdProducts = append(createdProducts, client.GetPlanClientId(product.GetId()))
		} else if productClientPlan.UpdateRequired && productClientPlan.IsActive {
			logger.Info("client plan exist, updating", "clientPlanId", clientPlanId)
			logger.V(1).Info("client plan details", "clientPlanId", clientPlanId, "productClientPlan", productClientPlan.ClientPlanDetail)
			productStatusChange, err = productController.UpdatePlan(ctx, productClientPlan.ClientPlanDetail, product, vendorMap[product.GetFamilyId()], usageType)
			if err != nil {
				productSyncErrors = append(productSyncErrors, ProductSyncError{Product: product, Err: errors.New(ProductSyncFailedToEditPlan)})
				continue
			}

		} else if !productClientPlan.UpdateRequired && !productClientPlan.IsActive {
			logger.Info("client plan exist, missing in product catalog deactivating", "clientPlanId", clientPlanId)
			if !strings.EqualFold(clientPlanId, client.GetDefaultPlanClientId()) {
				_, err := productController.ariaPlanClient.DeactivatePlanFromClientPlanDetail(ctx, &productClientPlan.ClientPlanDetail, usageType)
				if err != nil {
					productSyncErrors = append(productSyncErrors, ProductSyncError{Product: product, Err: errors.New(ProductSyncFailedToDeactivatePlan)})
					continue
				}
				productStatusChange = true
			}
		}
		if productStatusChange {
			status := *pb.ProductStatus_PRODUCT_STATUS_READY.Enum()
			productSyncErr := ""
			if err != nil {
				status = *pb.ProductStatus_PRODUCT_STATUS_ERROR.Enum()
				productSyncErr = err.Error()
				logger.Error(err, "product catalog sync to aria systems error")

			}
			productStatus := pb.SetProductStatus{VendorId: product.GetVendorId(), FamilyId: product.GetFamilyId(), ProductId: product.GetId(), Status: status, Error: productSyncErr}
			productsStatus = append(productsStatus, &productStatus)
		}
	}
	logger.Info("product catatlog request status ", "productsStatus", productsStatus)
	if len(productsStatus) > 0 {
		productStatusRequest := pb.SetProductStatusRequest{Status: productsStatus}
		err = productController.productClient.SetProductStatus(ctx, &productStatusRequest)
		if err != nil {
			logger.Error(err, "product catalog set status error")
		}
	}

	if err = productController.ariaPromoClient.AddPlansToPromo(ctx, createdProducts); err != nil {
		// When this happens the plans have been created but they are missing
		// from the promo plan set. The next time sync runs, the plans will not
		// be returned by GetAllClientPlansForPromoCode (which we're not
		// calling yet). This will lead to CreatePlan failing. This means that
		// when CreatePlan fails this code needs to look for the existing plan
		// that's not in the promo plan set.
		logger.Error(err, "error adding plans to promo plan set")

		// we can keep going. The plans are usable for usage reporting,
		// they're just not wired up for GetAllClientPlansForPromoCode
	}

	// Notify of all errors with product sync for the same product.
	// Errors with syncing are independent of each other and need to be reported individually.
	for _, productSyncError := range productSyncErrors {
		billingCommon.NotifyInvalidProductEntryForSync(ctx, productSyncError.Product, productSyncError.Err)
	}
	return nil
}

func (*ProductController) SyncError(ctx context.Context, validProducts []*pb.Product, err error, productSyncErrors []ProductSyncError) []ProductSyncError {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.SyncError").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for _, product := range validProducts {
		logger.Error(err, "failed to check if plan exists", "clientPlanId", client.GetPlanClientId(product.GetId()))
		productSyncErrors = append(productSyncErrors, ProductSyncError{Product: product, Err: errors.New(ProductSyncFailedToCheckPlanExists)})
	}
	return productSyncErrors
}

/*
AllClientPlanDtl: client plan details is correlated and mapped for
plan_name, plan_desc, plan_services and plan_supp_fields with product, product family & usage type
plan_services: plan service is parsed for service_desc(same as service name) and plan_service_rates
plan_service_rates: is parsed for client_rate_schedule_id, rate_per_unit and from_unit for active schedule
then edit_plan_m  and update_service_m
*/
func (productController *ProductController) UpdatePlan(ctx context.Context, clientPlanDtl data.AllClientPlanDtl, product *pb.Product, productFamily *pb.ProductFamily, usageType *data.UsageType) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.UpdatePlan").Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	edited := false
	var err error
	hasDiffPlanDetail := false
	hasDiffPlanService := false
	hasDiffPlanServiceRate := false
	hasdiffPlanSupplField := false

	logger.V(1).Info("ClientPlanDtls", "clientPlanDtl", clientPlanDtl, "productId", product.Id, "productName", product.Name, "usageType", usageType)
	hasDiffPlanDetail = client.HasDiffPlanDetail(ctx, clientPlanDtl, product)

	for _, planService := range clientPlanDtl.PlanServices {
		logger.V(1).Info("ClientPlanDtls.PlanService", "planService", planService, "productId", product.Id, "productName", product.Name)
		hasDiffPlanService = client.HasDiffPlanService(ctx, planService, usageType, product)
		if hasDiffPlanService {
			break
		}
		logger.V(1).Info("planService.PlanServiceRates", "PlanServiceRates", planService.PlanServiceRates)
		hasDiffPlanServiceRate = client.HasDiffPlanServiceRate(ctx, planService.PlanServiceRates, product)
		if hasDiffPlanServiceRate {
			break
		}
	}

	if hasDiffPlanService {
		_, err = productController.ariaServiceClient.UpdateAriaService(
			ctx, client.GetServiceClientId(product.Id), product.Metadata["displayName"],
			usageType.UsageTypeNo)
		edited = true
	}

	logger.V(1).Info("client plan supp fields", "PlanSuppFields", clientPlanDtl.PlanSuppFields)
	hasdiffPlanSupplField = client.HasDiffPlanSupplField(ctx, clientPlanDtl.PlanSuppFields, product)

	if hasDiffPlanDetail || hasDiffPlanServiceRate || hasdiffPlanSupplField {
		logger.Info("update client plan", "hasDiffPlanDetail", hasDiffPlanDetail, "hasDiffPlanService", hasDiffPlanService, "hasDiffPlanServiceRate", hasDiffPlanServiceRate, "hasdiffPlanSupplField", hasdiffPlanSupplField)
		_, err = productController.ariaPlanClient.EditPlan(ctx, product, productFamily, usageType)
		if err != nil {
			logger.Error(err, "failed to update plan", "product", product)
			return false, err
		}
		edited = true
	}
	return edited, err
}

func (productController *ProductController) GetUsageUnitType(ctx context.Context, product *pb.Product) (*data.UsageType, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.GetUsageUnitType").Start()
	defer span.End()
	logger.V(1).Info("product catalog", "product name", product.Name)
	logger.V(1).Info("product metadata", "metadata", product.GetMetadata(), "processingTypeKey", processingTypeKey)
	if _, foundProcessingType := product.GetMetadata()[processingTypeKey]; foundProcessingType {
		processingType := product.GetMetadata()[processingTypeKey]
		if inferenceProcessingType == processingType {
			if usageUnitTypes.inferenceUsageType == nil {
				usageType, err := productController.ariaUsageTypeClient.GetInferenceUsageType(ctx)
				if err != nil {
					logger.Error(err, "failed to get inference usage type ", "usageType", usageType)
					return nil, err
				}
				usageUnitTypes.inferenceUsageType = usageType
			}
			logger.V(1).Info("product usage unit type", "product name", product.Name, "inferenceUsageType", usageUnitTypes.inferenceUsageType)
			return usageUnitTypes.inferenceUsageType, nil
		} else if tokenProcessingType == processingType {
			if usageUnitTypes.tokenUsageType == nil {
				usageType, err := productController.ariaUsageTypeClient.GetTokenUsageType(ctx)
				if err != nil {
					logger.Error(err, "failed to get token usage type ", "usageType", usageType)
					return nil, err
				}
				usageUnitTypes.tokenUsageType = usageType
			}
			logger.V(1).Info("product usage unit type", "product name", product.Name, "tokenUsageType", usageUnitTypes.tokenUsageType)
			return usageUnitTypes.tokenUsageType, nil
		}
	}

	logger.V(1).Info("product metadata", "metadata", product.GetMetadata(), "instanceTypeKey", instanceTypeKey)
	if _, foundInstanceType := product.GetMetadata()[instanceTypeKey]; foundInstanceType {
		instanceType := product.GetMetadata()[instanceTypeKey]
		logger.V(1).Info("product instance type", "storageInstanceType", config.Cfg.GetStorageInstanceTypes(), "instanceType", instanceType)
		if slices.Contains(config.Cfg.GetStorageInstanceTypes(), instanceType) {
			if usageUnitTypes.storageUsageType == nil {
				usageType, err := productController.ariaUsageTypeClient.GetStorageUsageType(ctx)
				if err != nil {
					logger.Error(err, "failed to get stroage usage type ", "usageType", usageType)
					return nil, err
				}
				usageUnitTypes.storageUsageType = usageType
			}
			logger.V(1).Info("product usage unit type", "product name", product.Name, "storageUsageType", usageUnitTypes.storageUsageType)
			return usageUnitTypes.storageUsageType, nil
		}
	}

	if usageUnitValue, ok := product.Metadata[UsageUnitKey]; ok {
		if usageUnitValue == MinuteUsageUnitValue {
			if usageUnitTypes.minutesUsageType == nil {
				usageType, err := productController.ariaUsageTypeClient.GetMinutesUsageType(ctx)
				if err != nil {
					logger.Error(err, "failed to get minute usage type ", "usageType", usageType)
					return nil, err
				}
				usageUnitTypes.minutesUsageType = usageType
			}
			logger.V(1).Info("product usage unit type", "product name", product.Name, "minutesUsageType", usageUnitTypes.minutesUsageType)
			return usageUnitTypes.minutesUsageType, nil
		}
	}
	err := errors.New(UsageUnitTypeUnavailable)
	span.RecordError(fmt.Errorf("failed to get usage unit type for product %v %v", product.Name, err))
	return nil, err
}
