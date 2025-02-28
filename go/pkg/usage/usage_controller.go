// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const usageUnitKey = "usage.unit"

type UsageController struct {
	cloudAccountClient billingCommon.CloudAccountSvcClient
	productClient      billingCommon.ProductClientInterface
	meteringClient     *billingCommon.MeteringClient
	usageData          *UsageData
	usageRecordData    *UsageRecordData
}

func NewUsageController(cloudAccountClient billingCommon.CloudAccountSvcClient,
	productClient billingCommon.ProductClientInterface, meteringClient *billingCommon.MeteringClient,
	usageData *UsageData, usageRecordData *UsageRecordData) *UsageController {
	return &UsageController{
		cloudAccountClient: cloudAccountClient,
		productClient:      productClient,
		meteringClient:     meteringClient,
		usageData:          usageData,
		usageRecordData:    usageRecordData,
	}
}

func (usageController *UsageController) markMeteringRecordsAsReported(ctx context.Context, meteringRecords []*pb.MeteringRecord) error {
	var usageIds []int64
	for _, meteringRecord := range meteringRecords {
		usageIds = append(usageIds, meteringRecord.Id)
	}
	err := usageController.meteringClient.UpdateUsagesAsReported(ctx, usageIds)
	return err
}

func (usageController *UsageController) getValidProducts(ctx context.Context) ([]*pb.Product, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.getValidProducts").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	vendors, err := usageController.productClient.GetProductCatalogVendors(ctx)
	// fail fast if cannot get vendors.
	if err != nil {
		logger.Error(err, "failed to get vendors from product catalog")
		return nil, err
	}
	products, err := usageController.productClient.GetProductCatalogProducts(ctx)
	// fail fast if cannot get products
	if err != nil {
		logger.Error(err, "failed to get products from product catalog")
		return nil, err
	}
	// check the products for the product family
	validProducts, productValidationErrors, err := billingCommon.ValidateProducts(ctx, vendors, products)

	for _, productValidationError := range productValidationErrors {
		logger.Error(productValidationError.Err, "product invalid", "product id", productValidationError.Product.GetId())
	}

	// fail if cannot validate products for product family
	if err != nil {
		logger.Error(err, "failed to validate products")
		return nil, err
	}

	return validProducts, nil
}

func (usageController *UsageController) printValidProducts(ctx context.Context, products []*pb.Product) {
	logger := log.FromContext(ctx).WithName("UsageController.printValidProducts")
	for _, product := range products {
		logger.Info("valid product found", "productId", product.Id)
		logger.Info("valid product found", "productName", product.Name)
	}
}

/*
*Calculate the usages for all resources.
*
 */
func (usageController *UsageController) CalculateUsages(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.CalculateUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("calculating usages across cloud accounts")

	validProducts, err := usageController.getValidProducts(ctx)
	//logger.Info("calculating usages across cloud accounts", "validProducts", validProducts)

	if err != nil {
		logger.Error(err, "getting valid products for calculation of usages failed")
		return err
	}

	usageController.printValidProducts(ctx, validProducts)

	// get unreported metering records
	logger.Info("getting unreported resource metering records")
	unreportedResourceMeteringRecords, err := usageController.meteringClient.SearchUnreportedResourceMeteringRecords(ctx)
	if err != nil {
		logger.Error(err, "failed to get unreported resource metering records")
		return err
	}

	if len(unreportedResourceMeteringRecords.ResourceMeteringRecordsList) == 0 {
		logger.Info("no unreported resource metering records")
		return nil
	}

	for _, resourceMeteringRecords := range unreportedResourceMeteringRecords.ResourceMeteringRecordsList {

		lengthOfResourceMeteringRecords := len(resourceMeteringRecords.MeteringRecords)
		// there are no metering records - continue
		if lengthOfResourceMeteringRecords == 0 {
			logger.Info("no unreported metering records for resource", "id", resourceMeteringRecords.ResourceId)
			continue
		}

		logger.Info("processing metering records for the resource", "id", resourceMeteringRecords.ResourceId, "cloudAccountId",
			resourceMeteringRecords.CloudAccountId)

		// check if the table already has a entry for the resource usage
		resourceUsages, err := usageController.usageData.GetResourceUsagesForResource(ctx, resourceMeteringRecords.ResourceId)
		if err != nil {
			logger.Error(err, "failed to get resource usages for resource", "id", resourceMeteringRecords.ResourceId)
			continue
		}

		startTime := resourceMeteringRecords.MeteringRecords[lengthOfResourceMeteringRecords-1].Timestamp
		endTime := resourceMeteringRecords.MeteringRecords[0].Timestamp

		lengthOfResourceUsages := len(resourceUsages.ResourceUsages)

		var previousUsageQuantity float64 = 0
		// this is the first entry for the resource
		if lengthOfResourceUsages == 0 {
			logger.Info("no existing resource usages, and adding for resource", "id", resourceMeteringRecords.ResourceId, "cloudAccountId",
				resourceMeteringRecords.CloudAccountId)

			err = usageController.addResourceUsage(ctx, resourceMeteringRecords.ResourceId, resourceMeteringRecords.CloudAccountId,
				resourceMeteringRecords.ResourceName, resourceMeteringRecords.Region,
				resourceMeteringRecords.MeteringRecords, validProducts, previousUsageQuantity, startTime.AsTime(), endTime.AsTime())
			if err != nil {
				logger.Error(err, "failed to add resource usages for resource", "id", resourceMeteringRecords.ResourceId)
				continue
			} else {
				err = usageController.markMeteringRecordsAsReported(ctx, resourceMeteringRecords.MeteringRecords)
				if err != nil {
					logger.Error(err, "failed to update metering records as reported for resource", "id", resourceMeteringRecords.ResourceId)
					continue
				}
			}
		} else {
			logger.Info("found existing resource usages, and adding for resource", "id", resourceMeteringRecords.ResourceId, "cloudAccountId",
				resourceMeteringRecords.CloudAccountId)
			sumOfPreviousQuantity, err := usageController.usageData.GetTotalResourceUsageQty(ctx, resourceMeteringRecords.ResourceId)

			if err != nil {
				logger.Error(err, "failed to get sum of usage quantity for resource", "id", resourceMeteringRecords.ResourceId)
				continue
			}

			previousMeteringRecord, err := usageController.meteringClient.FindPreviousUsage(ctx, &pb.Usage{
				Id:         resourceMeteringRecords.MeteringRecords[len(resourceMeteringRecords.MeteringRecords)-1].Id,
				ResourceId: resourceMeteringRecords.ResourceId,
			})

			meteringRecords := resourceMeteringRecords.MeteringRecords

			// consider the previous reported metering records for doing calculation for linear historical metering records.
			if err == nil && (previousMeteringRecord != nil) && previousMeteringRecord.Reported {
				// todo: add the resource name and region but it is fine.
				meteringRecords = append(meteringRecords, &pb.MeteringRecord{
					Id:             previousMeteringRecord.Id,
					TransactionId:  previousMeteringRecord.TransactionId,
					ResourceId:     previousMeteringRecord.ResourceId,
					CloudAccountId: previousMeteringRecord.CloudAccountId,
					Timestamp:      previousMeteringRecord.Timestamp,
					Properties:     previousMeteringRecord.Properties,
					Reported:       previousMeteringRecord.Reported,
				})
			}

			// this is update of the entry for a resource.
			err = usageController.updateResourceUsage(ctx, resourceMeteringRecords.ResourceId, resourceMeteringRecords.CloudAccountId,
				resourceMeteringRecords.ResourceName, resourceMeteringRecords.Region,
				meteringRecords, resourceUsages.ResourceUsages[0], sumOfPreviousQuantity,
				startTime.AsTime(), endTime.AsTime())
			if err != nil {
				logger.Error(err, "failed to update resource usages for resource", "id", resourceMeteringRecords.ResourceId)
				continue
			} else {
				err = usageController.markMeteringRecordsAsReported(ctx, resourceMeteringRecords.MeteringRecords)
				if err != nil {
					logger.Error(err, "failed to update metering records as reported for resource", "id", resourceMeteringRecords.ResourceId)
					continue
				}
			}
		}
	}

	return nil
}

func GetProductsForProperties(ctx context.Context,
	properties map[string]string, products []*pb.Product) ([]*pb.Product, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.GetProductsForMeteringRecord").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	mappedProducts := []*pb.Product{}

	for _, product := range products {
		productMatches, err := CheckIfProductMapsToProperties(product, properties)
		if err != nil {
			logger.Error(err, "failed to match product to metering for product", "id", product.Id)
		}
		if productMatches {
			mappedProducts = append(mappedProducts, product)
		}
	}

	return mappedProducts, nil
}

func GetRateForProduct(ctx context.Context, cloudAccountClient billingCommon.CloudAccountSvcClient, product *pb.Product, cloudAccountId string) (string, string, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.GetRateForProduct").WithValues("cloudAccountId", cloudAccountId, "productId", product.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccountType, err := cloudAccountClient.GetCloudAccountType(ctx, &pb.CloudAccountId{Id: cloudAccountId})

	if err != nil {
		return "", "", err
	}

	for _, rate := range product.Rates {
		if rate.AccountType == cloudAccountType {
			return rate.Unit.Enum().String(), rate.Rate, nil
		}
	}

	err = errors.New("failed to get rate for product id: " + product.Id)
	logger.Error(err, "failed to get rate for product", "productId", product.Id)

	return "", "", err
}

/*
*
Add the usage for a resource.
*
*/
func (usageController *UsageController) addResourceUsage(ctx context.Context, resourceId string, cloudAccountId string, resourceName string,
	region string, meteringRecords []*pb.MeteringRecord, products []*pb.Product, previousQuantity float64, startTime time.Time, endTime time.Time) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.addResourceUsage").WithValues("cloudAccountId", cloudAccountId, "resourceId", resourceId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("adding a resource usage entry for resource", "id", resourceId, "cloudAccountId", cloudAccountId)

	lengthOfMeteringRecords := len(meteringRecords)

	// nothing to do..
	if lengthOfMeteringRecords == 0 {
		return nil
	}

	mappedProducts, err := GetProductsForProperties(ctx, meteringRecords[0].Properties, products)
	if err != nil {
		usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_NO_MATCHING_PRODUCT)
		logger.Error(err, "failed to map metering record to products for resource", "id", resourceId)
		return err
	}

	if len(mappedProducts) == 0 {
		usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_NO_MATCHING_PRODUCT)
		logger.Error(err, "no mapped products found for resource", "id", resourceId)
		return err
	}

	// we only care about the last metering record.
	usageQuantity, err := GetUsageQuantity(meteringRecords)
	if err != nil {
		usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_FAILED_TO_CALCULATE_QTY)
		logger.Error(err, "failed to get metering quantity", "id", meteringRecords[0].Id)
		return err
	}

	usageQuantity = usageQuantity - previousQuantity

	for _, mappedProduct := range mappedProducts {
		logger.Info("found mapped product for resource", "id", resourceId, "cloudAccountId", cloudAccountId, "product", mappedProduct.Name)
		// the usage unit type is wrong!! - it needs to be mins and not dollars per min..
		usageUnitType, rate, err := GetRateForProduct(ctx, usageController.cloudAccountClient, mappedProduct, cloudAccountId)
		if err != nil {
			usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_FAILED_TO_GET_PRODUCT_RATE)
			logger.Error(err, "failed to get usage unit type and rate for product", "id", mappedProduct.Id)
			return err
		}

		productRate, err := strconv.ParseFloat(rate, 64)
		if err != nil {
			usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_FAILED_TO_GET_PRODUCT_RATE)
			logger.Error(err, "invalid rate for product", "id", mappedProduct.Id)
			return err
		}
		if usageUnit, ok := mappedProduct.Metadata[usageUnitKey]; ok {
			usageUnitType = usageUnit
		}
		err = usageController.addUsages(ctx, cloudAccountId,
			resourceId, resourceName, mappedProduct.Id, mappedProduct.Name,
			region, usageQuantity, productRate, usageUnitType, startTime, endTime)

		if err != nil {
			logger.Error(err, "failed to add resource usage entries for resource", "id", resourceId)
			return err
		}
	}

	// add the resource metering record to keep track of the last metering record handled.
	// this will be used to check if a metering record received is after the last metering record handled.
	// if yes, then it will be ignored.
	// if the last metering record received is earlier than the last metering record handled,
	// then we ignore the entire metering records list.
	// something will add in the next increment of changes.
	resourceMeteringCreate := &ResourceMeteringCreate{
		ResourceId:     resourceId,
		CloudAccountId: cloudAccountId,
		TransactionId:  meteringRecords[0].TransactionId,
		Region:         region,
		LastRecorded:   meteringRecords[0].Timestamp.AsTime(),
	}

	_, err = usageController.usageData.StoreResourceMetering(ctx, resourceMeteringCreate)

	if err != nil {
		logger.Error(err, "failed to store the metering for the resource")
		return err
	}

	return nil
}

func (usageController *UsageController) invalidateMeteringRecords(ctx context.Context, meteringRecords []*pb.MeteringRecord,
	meteringInvalidityReason pb.MeteringRecordInvalidityReason) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.invalidateMeteringRecords").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	usageController.meteringClient.InvalidateMeteringRecords(ctx, meteringRecords, meteringInvalidityReason)
	return nil
}

func (usageController *UsageController) updateResourceUsage(ctx context.Context, resourceId string, cloudAccountId string, resourceName string,
	region string, meteringRecords []*pb.MeteringRecord, lastResourceUsage *pb.ResourceUsage, previousQuantity float64,
	startTime time.Time, endTime time.Time) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.updateResourceUsage").WithValues("cloudAccountId", cloudAccountId, "resourceId", resourceId, "resourceName", resourceName).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("updating resource usage entries for resource", "id", resourceId, "cloudAccountId", cloudAccountId)
	logger.Info("using these values from the last resource usage", "id", lastResourceUsage.Id, "quantity", lastResourceUsage.Quantity)

	usageQuantity, err := GetUsageQuantity(meteringRecords)
	if err != nil {
		usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_FAILED_TO_CALCULATE_QTY)
		logger.Error(err, "failed to get metering quantity")
		return err
	}

	// we will default to using previous metering records for calculation once hr is a default usage metric
	// also performance needs to be improved.
	if _, foundServiceTypeKey := meteringRecords[0].Properties[meteringServiceTypeKey]; foundServiceTypeKey {
		serviceType := meteringRecords[0].Properties[meteringServiceTypeKey]
		if slices.Contains(Cfg.StorageServiceTypes, serviceType) {
			previousQuantity = 0
		}
	}

	if usageQuantity < previousQuantity {
		usageController.invalidateMeteringRecords(ctx, meteringRecords, pb.MeteringRecordInvalidityReason_INVALID_METERING_QTY)
		logger.Error(err, "invalid quantity in the metering record")
		return err
	}

	// we will use a update path if the usages have not been reported.
	// however, in the existing reporting mechanism a usage will be marked as reported after some processing.
	// if during this processing time, we update the amount, then it does not work.
	// hence, we will add the update path when we change the reporting code in the next iteration.
	// reporting code needs to be changed to mark as reported as a first thing and then retry to report.
	// delayed marking of as reported can leave to inconsistent states and hence not using update path for now.
	// Not using update paths will mean that we will end up with more records than we need which is ok.

	logger.Info("adding record for", "quantity", usageQuantity, "previousQuantity", previousQuantity)
	quantityToUpdate := usageQuantity - previousQuantity
	err = usageController.addUsages(ctx, cloudAccountId,
		resourceId, resourceName, lastResourceUsage.ProductId, lastResourceUsage.ProductName,
		region, quantityToUpdate, lastResourceUsage.Rate, lastResourceUsage.UsageUnitType,
		startTime, endTime)

	if err != nil {
		logger.Error(err, "failed to update resource usage entries for resource", "id", resourceId)
		return err
	}

	resourceMeteringRecordUpdate := &ResourceMeteringUpdate{
		TransactionId: meteringRecords[0].TransactionId,
		LastRecorded:  meteringRecords[0].Timestamp.AsTime(),
	}

	err = usageController.usageData.UpdateResourceMetering(ctx, resourceId, resourceMeteringRecordUpdate)

	if err != nil {
		logger.Error(err, "failed to update	the metering for the resource")
		return err
	}

	return nil
}

func (usageController *UsageController) addUsages(ctx context.Context, cloudAccountId string,
	resourceId string, resourceName string, productId string, productName string,
	region string, quantity float64, rate float64, usageUnitType string, startTime time.Time, endTime time.Time) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.addUsages").WithValues("cloudAccountId", cloudAccountId, "resourceId", resourceId, "resourceName", resourceName).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// store the resource usage.
	resourceUsageRecord := &pb.ResourceUsageCreate{
		CloudAccountId: cloudAccountId,
		ResourceId:     resourceId,
		// resource name needs to be standardized across all services
		ResourceName:  resourceName,
		ProductId:     productId,
		ProductName:   productName,
		TransactionId: uuid.NewString(),
		Region:        region,
		Quantity:      quantity,
		Rate:          rate,
		UsageUnitType: usageUnitType,
		StartTime:     timestamppb.New(startTime),
		EndTime:       timestamppb.New(endTime),
	}

	resourceUsageRecordCreated, err := usageController.usageData.StoreResourceUsage(ctx, resourceUsageRecord, time.Now())
	if err != nil {
		logger.Error(err, "failed to store resource usage record for resource", "id", resourceId)
		return err
	}

	productUsageRecordCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: cloudAccountId,
		ProductId:      productId,
		ProductName:    productName,
		Region:         region,
		Quantity:       quantity,
		Rate:           rate,
		UsageUnitType:  usageUnitType,
		StartTime:      startTime,
		EndTime:        endTime,
	}

	// store the product usage record.
	err = usageController.usageData.StoreProductUsage(ctx, productUsageRecordCreate, time.Now())
	if err != nil {
		logger.Error(err, "failed to store product usage for product", "id", productId)
		err = usageController.usageData.DeleteResourceUsage(ctx, resourceUsageRecordCreated.Id)
		if err != nil {
			logger.Error(err, "failed to delete resource usage for resource", "id", resourceId)
		}
		return err
	}
	return nil
}

func (usageController *UsageController) CalculateProductUsages(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageController.CalculateProductUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("calculating product usages across cloud accounts")

	validProducts, err := usageController.getValidProducts(ctx)

	if err != nil {
		logger.Error(err, "getting valid products for calculation of usages failed")
		return err
	}

	usageController.printValidProducts(ctx, validProducts)

	// get unreported product usage records
	logger.Info("getting unreported product usage records")
	unreportedProductUsageRecords, err := usageController.usageRecordData.GetUnreportedProductUsageRecords(ctx)
	if err != nil {
		logger.Error(err, "failed to get unreported product usage records")
		return err
	}

	if len(unreportedProductUsageRecords) == 0 {
		logger.Info("no unreported product usage records")
		return nil
	}

	for _, productUsageRecord := range unreportedProductUsageRecords {

		mappedProducts, err := GetProductsForProperties(ctx, productUsageRecord.Properties, validProducts)
		if err != nil {
			logger.Error(err, "failed to map product usage record to products for product usage record", "id", productUsageRecord.Id)
			continue
		}

		countOfProductsAdded := 0
		// todo: this needs to be a single product.
		for _, mappedProduct := range mappedProducts {
			logger.Info("found mapped product for product usage record", "id", productUsageRecord.Id, "cloudAccountId",
				productUsageRecord.CloudAccountId, "product", mappedProduct.Name)

			//todo: some of this can move to common code..
			usageUnitType, rate, err := GetRateForProduct(ctx, usageController.cloudAccountClient, mappedProduct, productUsageRecord.CloudAccountId)

			if err != nil {
				logger.Error(err, "failed to get usage unit type and rate for product", "id", mappedProduct.Id)
				continue
			}

			productRate, err := strconv.ParseFloat(rate, 64)
			if err != nil {
				logger.Error(err, "invalid rate for product", "id", mappedProduct.Id)
				continue
			}

			var startTime time.Time
			if productUsageRecord.StartTime == nil {
				startTime = productUsageRecord.Timestamp.AsTime()
			} else {
				startTime = productUsageRecord.StartTime.AsTime()
			}

			var endTime time.Time
			if productUsageRecord.EndTime == nil {
				endTime = productUsageRecord.Timestamp.AsTime()
			} else {
				endTime = productUsageRecord.EndTime.AsTime()
			}

			productUsageId := uuid.NewString()
			//todo: add timestamp to product usage create.
			productUsageCreate := &ProductUsageCreate{
				Id:             productUsageId,
				CloudAccountId: productUsageRecord.CloudAccountId,
				ProductId:      mappedProduct.Id,
				ProductName:    mappedProduct.Name,
				Region:         productUsageRecord.Region,
				Quantity:       productUsageRecord.Quantity,
				Rate:           productRate,
				UsageUnitType:  usageUnitType,
				StartTime:      startTime,
				EndTime:        endTime,
			}

			err = usageController.usageData.StoreProductUsage(ctx, productUsageCreate, time.Now())

			if err != nil {
				logger.Error(err, "failed to add product usage entry for product usage record", "id", productUsageRecord.Id)
				continue
			}

			err = usageController.usageData.StoreProductUsageReport(ctx, productUsageCreate, productUsageId)

			if err != nil {
				logger.Error(err, "failed to add product usage report entry for product usage record", "id", productUsageRecord.Id)
				err = usageController.usageData.DeleteProductUsage(ctx, productUsageId)

				if err != nil {
					logger.Error(err, "failed to revert product usage report entry for product usage record", "id", productUsageRecord.Id)
				}

				continue
			}

			countOfProductsAdded++
		}

		if countOfProductsAdded == len(mappedProducts) {
			err = usageController.usageRecordData.MarkProductUsageRecordAsReported(ctx, productUsageRecord.Id)
			if err != nil {
				logger.Error(err, "failed update as reported for product usage record", "id", productUsageRecord.Id)
				continue
			}
		}

	}

	return nil
}
