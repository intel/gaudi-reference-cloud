// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"strconv"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	getCloudAccountsForDeactivationErr         = "failed to get cloud accounts for checking deactivation of paid instances"
	getTradeResCloudAccountsForDeactivationErr = "failed to get trade restricted cloud accounts for checking deactivation of instances"
	getPaidProductsForDeactivationErr          = "failed to get paid products from product catalog service"
	getAllProductsForDeactivationErr           = "failed to get all products from product catalog service"
	getProductsForProductFamilyErr             = "failed to get products from product catalog service for product family"
	getCloudAccountsForServiceDeactivationErr  = "failed to get cloud accounts for service deactivation"
	sendCloudAccountsForServiceDeactivationErr = "failed to send cloud accounts for service deactivation"
	defaultCleanupThreshold                    = 30 //days
	defaultTimeout                             = 5  //seconds
)

type BillingDeactivateInstancesService struct {
	cloudAccountClient *billingCommon.CloudAccountSvcClient
	pb.UnimplementedBillingDeactivateInstancesServiceServer
}

func (svc *BillingDeactivateInstancesService) getInstanceDeactivationList(ctx context.Context, accountTypes []pb.AccountType, tradeRestricted bool) ([]*pb.DeactivateInstances, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.getInstanceDeactivationList").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	if tradeRestricted {
		log.Info("getting the list of trade restricted cloud accounts for deactivation of instances")
	} else {
		log.Info("getting the list of cloud accounts for deactivation of paid instances")
	}
	terminatePaidServices := true
	accounts, err := svc.getCloudAccountsForDeactivation(ctx, accountTypes, tradeRestricted, terminatePaidServices)
	if err != nil {
		if tradeRestricted {
			log.Error(err, getTradeResCloudAccountsForDeactivationErr)
			return nil, status.Errorf(codes.Internal, "%v: %v", getAllProductsForDeactivationErr, err)
		} else {
			log.Error(err, getCloudAccountsForDeactivationErr)
			return nil, status.Errorf(codes.Internal, "%v: %v", getPaidProductsForDeactivationErr, err)
		}
	}

	// in case of retriving trade restricted accounts, list of all products is required
	// otherwise we need to get a lit of only the paid products
	productMap, err := getProductsMap(ctx, accountTypes, !tradeRestricted, "")
	if err != nil {
		if tradeRestricted {
			log.Error(err, getAllProductsForDeactivationErr)
			return nil, status.Errorf(codes.Internal, "%v: %v", getAllProductsForDeactivationErr, err)
		} else {
			log.Error(err, getPaidProductsForDeactivationErr)
			return nil, status.Errorf(codes.Internal, "%v: %v", getPaidProductsForDeactivationErr, err)
		}

	}

	// create the list of instance quotas for each account type based on the product map.
	instanceQuotas := make(map[pb.AccountType][]*pb.InstanceQuotas)
	for accountType, products := range productMap {
		for _, product := range products {
			instanceQuotas[accountType] = append(instanceQuotas[accountType], &pb.InstanceQuotas{
				InstanceType: product.GetName(),
				Limit:        0,
			})
		}
	}

	// create the list of accounts with their corresponding instance quotas.
	var deactivationList []*pb.DeactivateInstances
	for _, account := range accounts {
		if quotas, found := instanceQuotas[account.Type]; found {
			deactivationList = append(deactivationList, &pb.DeactivateInstances{
				CloudAccountId: account.GetId(),
				Quotas:         quotas,
			})
		}
	}

	return deactivationList, nil
}

func (svc *BillingDeactivateInstancesService) GetDeactivateInstances(ctx context.Context, _ *emptypb.Empty) (*pb.DeactivateInstancesResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.GetDeactivateInstances").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	servicesTerminationAccountTypes := Cfg.GetServicesTerminationAccountTypes()
	accountTypes := GetAccountTypes(ctx, servicesTerminationAccountTypes)
	if len(accountTypes) == 0 {
		return nil, status.Errorf(codes.Internal, "%v for account types %v", getCloudAccountsForDeactivationErr, servicesTerminationAccountTypes)
	}
	deactivationList, err := svc.getInstanceDeactivationList(ctx, accountTypes, false)
	if err != nil {
		log.Error(err, getCloudAccountsForDeactivationErr)
		return nil, status.Errorf(codes.Internal, "%v: %v", getPaidProductsForDeactivationErr, err)
	}

	accountTypes = []pb.AccountType{
		pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
		pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING,
		pb.AccountType_ACCOUNT_TYPE_INTEL,
		pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}
	tradeResDeactivationList, err := svc.getInstanceDeactivationList(ctx, accountTypes, true)
	if err != nil {
		log.Error(err, getCloudAccountsForDeactivationErr)
		return nil, status.Errorf(codes.Internal, "%v: %v", getPaidProductsForDeactivationErr, err)
	}
	deactivationList = append(deactivationList, tradeResDeactivationList...)

	response := &pb.DeactivateInstancesResponse{DeactivationList: deactivationList}

	return response, nil
}

func (svc *BillingDeactivateInstancesService) GetDeactivateInstancesStream(req *emptypb.Empty, stream pb.BillingDeactivateInstancesService_GetDeactivateInstancesStreamServer) error {
	ctx, log, span := obs.LogAndSpanFromContext(stream.Context()).WithName("BillingDeactivateInstancesService.GetDeactivateInstances").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	servicesTerminationAccountTypes := Cfg.GetServicesTerminationAccountTypes()
	accountTypes := GetAccountTypes(ctx, servicesTerminationAccountTypes)
	if len(accountTypes) == 0 {
		return status.Errorf(codes.Internal, "%v for account types %v", getCloudAccountsForDeactivationErr, servicesTerminationAccountTypes)
	}

	deactivationLists, err := svc.getDeactivationLists(ctx, accountTypes)
	if err != nil {
		log.Error(err, getCloudAccountsForDeactivationErr)
		return status.Errorf(codes.Internal, "%v: %v", getPaidProductsForDeactivationErr, err)
	}
	const batchSize = 100

	for i := 0; i < len(deactivationLists); i += batchSize {
		end := i + batchSize
		if end > len(deactivationLists) {
			end = len(deactivationLists)
		}
		batch := deactivationLists[i:end]
		log.Info("Sending Stream batch", "BatchSize", len(batch))
		if err := svc.InvokeSend(stream.Context(), func() error {
			return stream.Send(&pb.DeactivateInstancesResponse{DeactivationList: batch})
		}); err != nil {
			log.Error(err, sendCloudAccountsForServiceDeactivationErr)
			return status.Errorf(codes.Internal, "%v: %v", sendCloudAccountsForServiceDeactivationErr, err)
		}

	}

	return nil
}

func (svc *BillingDeactivateInstancesService) getDeactivationLists(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.DeactivateInstances, error) {

	deactivationList, err := svc.getInstanceDeactivationList(ctx, accountTypes, false)
	if err != nil {
		return nil, err
	}

	accountTypes = []pb.AccountType{
		pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
		pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING,
		pb.AccountType_ACCOUNT_TYPE_INTEL,
		pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}
	tradeResDeactivationList, err := svc.getInstanceDeactivationList(ctx, accountTypes, true)
	if err != nil {
		return nil, err
	}
	deactivationList = append(deactivationList, tradeResDeactivationList...)

	return deactivationList, nil
}

func (svc *BillingDeactivateInstancesService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("BillingDeactivateInstancesService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func getProductsMap(ctx context.Context, accountTypes []pb.AccountType, getPaidProducts bool, productFamily string) (map[pb.AccountType][]*pb.Product, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.getPaidProductMap").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")

	products, err := productClient.GetProductCatalogProductsForAccountTypes(ctx, accountTypes)
	if err != nil {
		log.Error(err, "failed to get paid products for deactivation")
		return nil, err
	}

	products_map := make(map[pb.AccountType][]*pb.Product)

	// find products associated with each account type
	for _, product := range products {
		for _, rate := range product.GetRates() {
			billingRate, err := strconv.ParseFloat(rate.Rate, 64)
			if err != nil {
				log.Error(err, "error converting string to float64")
				return nil, err
			}
			if len(productFamily) > 0 {
				if tProductFamily, ok := product.Metadata[productFamilyDescriptionKey]; ok {
					if productFamily == tProductFamily {
						products_map[rate.GetAccountType()] = append(products_map[rate.GetAccountType()], product)
					}
				} else {
					log.V(9).Info("missing product family from product metadata", "productId", product.Id, "metadata", product.Metadata, "productFamily", productFamily)
				}
				continue
			}
			if !getPaidProducts {
				products_map[rate.GetAccountType()] = append(products_map[rate.GetAccountType()], product)
			}
			if getPaidProducts && billingRate > 0 {
				products_map[rate.GetAccountType()] = append(products_map[rate.GetAccountType()], product)
			}
		}
	}

	return products_map, nil
}

func (svc *BillingDeactivateInstancesService) getCloudAccountsForDeactivation(ctx context.Context, accountTypes []pb.AccountType, getTradeResAccounts bool, terminatePaidServices bool) ([]*pb.CloudAccount, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.getCloudAccountsForDeactivation").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")

	cloudAccounts := []*pb.CloudAccount{}

	for _, accountType := range accountTypes {
		filter := &pb.CloudAccountFilter{}
		if getTradeResAccounts {
			tradeRestricted := true
			filter = &pb.CloudAccountFilter{
				Type:            &accountType,
				TradeRestricted: &tradeRestricted,
			}
		} else {
			paidServicesAllowed := false
			filter = &pb.CloudAccountFilter{
				Type:                &accountType,
				PaidServicesAllowed: &paidServicesAllowed,
			}
			if terminatePaidServices {
				filter.TerminatePaidServices = &terminatePaidServices
			}
		}

		results, err := svc.cloudAccountClient.SearchCloudAccounts(ctx, filter)
		if err != nil {
			return nil, err
		}
		cloudAccounts = append(cloudAccounts, results...)
	}

	for _, cloudAccount := range cloudAccounts {
		if cloudAccount.TerminatePaidServices {
			log.V(9).Info("the paid services for cloud account will be terminated due to credits depletion", "id", cloudAccount.Id)
		}
		if cloudAccount.TradeRestricted {
			log.V(9).Info("the services for cloud account will be terminated due to trade restrictions", "id", cloudAccount.Id)
		}

	}
	return cloudAccounts, nil
}

func (svc *BillingDeactivateInstancesService) GetDeactivatedServiceAccounts(req *pb.GetDeactivatedAccountsRequest, stream pb.BillingDeactivateInstancesService_GetDeactivatedServiceAccountsServer) error {
	ctx, log, span := obs.LogAndSpanFromContext(stream.Context()).WithName("BillingDeactivateInstancesService.GetDeactivatedServiceAccounts").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	servicesTerminationAccountTypes := Cfg.GetServicesTerminationAccountTypes()
	accountTypes := GetAccountTypes(ctx, servicesTerminationAccountTypes)
	if len(accountTypes) == 0 {
		return status.Errorf(codes.Internal, "%v for account types %v", getCloudAccountsForServiceDeactivationErr, servicesTerminationAccountTypes)
	}
	log.V(9).Info("GetDeactivatedServiceAccounts request", "req", req)
	deactivationAccountsMap, err := svc.getDeactivationAccounts(ctx, accountTypes, req)
	if err != nil {
		log.Error(err, getCloudAccountsForServiceDeactivationErr)
		return status.Errorf(codes.Internal, "%v: %v", getCloudAccountsForServiceDeactivationErr, err)
	}

	log.Info("GetDeactivatedServiceAccounts count", "deactivationAccount count", len(deactivationAccountsMap))
	for _, deactivationAccount := range deactivationAccountsMap {
		select {
		case <-stream.Context().Done():
			log.Info("client canceled the stream")
			err := stream.Context().Err()
			if err == context.Canceled {
				log.Error(err, "error client context was canceled")
			} else if err == context.DeadlineExceeded {
				log.Error(err, "context client deadline exceeded")
			}
			log.V(9).Info("GetDeactivatedServiceAccounts", "deactivationAccount", deactivationAccount)
			return status.Errorf(codes.Canceled, "stream context canceled")
		default:
			if err := svc.InvokeSend(stream.Context(), func() error { return stream.Send(deactivationAccount) }); err != nil {
				log.Error(err, sendCloudAccountsForServiceDeactivationErr)
				return status.Errorf(codes.Internal, "%v: %v", sendCloudAccountsForServiceDeactivationErr, err)
			}
		}
	}

	return nil
}

func (svc *BillingDeactivateInstancesService) InvokeSend(ctx context.Context, f func() error) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.InvokeSend").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	errChan := make(chan error, 1)
	go func() {
		errChan <- f()
		close(errChan)
	}()
	t := time.NewTimer(defaultTimeout * time.Second)
	select {
	case <-ctx.Done():
		log.Error(ctx.Err(), "context error")
		return ctx.Err()
	case err := <-errChan:
		if !t.Stop() {
			<-t.C
		}
		return err
	}
}

func (svc *BillingDeactivateInstancesService) getDeactivationAccounts(ctx context.Context, accountTypes []pb.AccountType, req *pb.GetDeactivatedAccountsRequest) (map[string]*pb.DeactivateAccounts, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.getDeactivationAccounts").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	log.V(1).Info("get cloud accounts for service deactivation")
	deactivationAccounts := make(map[string]*pb.DeactivateAccounts)
	err := svc.getDeactivateAccounts(ctx, accountTypes, req, false, deactivationAccounts)
	if err != nil {
		log.Error(err, getCloudAccountsForServiceDeactivationErr)
		return nil, status.Errorf(codes.Internal, "%v: %v", getPaidProductsForDeactivationErr, err)

	}
	tradeRestricted := true
	if req.TradeRestricted != nil {
		tradeRestricted = *req.TradeRestricted
	}
	if tradeRestricted {
		log.V(1).Info("get trade restricted cloud accounts for service deactivation")
		accountTypes = []pb.AccountType{
			pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
			pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING,
			pb.AccountType_ACCOUNT_TYPE_INTEL,
			pb.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		err := svc.getDeactivateAccounts(ctx, accountTypes, req, tradeRestricted, deactivationAccounts)
		if err != nil {
			log.Error(err, getTradeResCloudAccountsForDeactivationErr)
			return nil, status.Errorf(codes.Internal, "%v: %v", getAllProductsForDeactivationErr, err)
		}
	}

	return deactivationAccounts, nil
}

func (svc *BillingDeactivateInstancesService) getDeactivateAccounts(ctx context.Context, accountTypes []pb.AccountType, req *pb.GetDeactivatedAccountsRequest, tradeRestrictedAccounts bool, deactivationAccounts map[string]*pb.DeactivateAccounts) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingDeactivateInstancesService.getDeactivateAccounts").Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	terminatePaidServices := false
	cloudAccts, err := svc.getCloudAccountsForDeactivation(ctx, accountTypes, tradeRestrictedAccounts, terminatePaidServices)
	if err != nil {
		log.Error(err, getTradeResCloudAccountsForDeactivationErr)
		return status.Errorf(codes.Internal, "%v: %v", getAllProductsForDeactivationErr, err)
	}
	productFamily := ""
	if req.ProductFamily != nil {
		productFamily = *req.ProductFamily
	}
	cleanupThreshold := defaultCleanupThreshold
	if req.CleanupThreshold != nil {
		cleanupThreshold = svc.cleanupThresholdInt(*req.CleanupThreshold)
	}
	// get product family product
	productMap, err := getProductsMap(ctx, accountTypes, tradeRestrictedAccounts, productFamily)
	if err != nil {
		log.Error(err, getAllProductsForDeactivationErr)
		return status.Errorf(codes.Internal, "%v: %v", getAllProductsForDeactivationErr, err)
	}
	if len(productMap) == 0 {
		log.V(9).Info(getProductsForProductFamilyErr, "productFamily", productFamily)
		return status.Errorf(codes.NotFound, "%v", getProductsForProductFamilyErr)
	}
	for _, cloudAcct := range cloudAccts {
		if _, found := productMap[cloudAcct.Type]; found {
			if tradeRestrictedAccounts {
				deactivationAccounts[cloudAcct.GetId()] = &pb.DeactivateAccounts{
					CloudAccountId:  cloudAcct.GetId(),
					Email:           cloudAcct.GetOwner(),
					CreditsDepleted: cloudAcct.GetCreditsDepleted(),
				}
			} else {
				if cloudAcct.CreditsDepleted != nil && cloudAcct.CreditsDepleted.AsTime().Unix() != 0 && !cloudAcct.PaidServicesAllowed {
					cleanupThresholdTime := cloudAcct.CreditsDepleted.AsTime().AddDate(0, 0, cleanupThreshold)
					creditsDepletedCheck := time.Now().After(cleanupThresholdTime)
					log.V(9).Info("credit depleted check for cloud account", "cloudAccountId", cloudAcct.GetId(), "creditsDepleted", cloudAcct.CreditsDepleted.AsTime(), "currentTime", time.Now(), "paidServicesAllowed", cloudAcct.PaidServicesAllowed, "creditsDepletedCheck", creditsDepletedCheck)
					if creditsDepletedCheck {
						deactivationAccounts[cloudAcct.GetId()] = &pb.DeactivateAccounts{
							CloudAccountId:  cloudAcct.GetId(),
							Email:           cloudAcct.GetOwner(),
							CreditsDepleted: cloudAcct.GetCreditsDepleted(),
						}
					}
				}
			}
		}
	}
	return nil
}

func (svc *BillingDeactivateInstancesService) cleanupThresholdInt(cleanupThreshold string) int {
	icleanupThreshold, err := strconv.Atoi(cleanupThreshold)
	if err != nil {
		return defaultCleanupThreshold
	}
	return icleanupThreshold
}
