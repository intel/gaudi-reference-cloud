// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	idcComputeSvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_compute"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewTrainingBatchUserService(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	// db setup
	db, err := sql.Open("postgres", "postgres://oneapiint_so:6Z7AvB5X0IeLqU1@postgres5808-lb-fm-in.dbaas.intel.com:5432/oneapiint")
	assert.NoError(t, err)
	defer db.Close()
	computeSvcClient := &idcComputeSvc.IDCServiceClient{}
	mockProductClient := NewMockProductClient(nil, nil, nil, nil, nil)
	mockCloudAccountSvcClient := NewMockCloudAccountSvcClient(nil, nil, nil)
	mockBillingClient := NewMockBillingClient(nil, nil)

	t.Run("nil db session", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", nil, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("nil compute service", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(nil, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("empty batch service", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "", "slurm-jupyterhub-service", "slurm-ssh-service", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("empty jupyterhub service", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "", "slurm-ssh-service", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("empty ssh service", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("nil product client and nil billing client", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, nil, mockCloudAccountSvcClient, nil)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("nil cloud account client and nil billing client", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, mockProductClient, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})

	t.Run("nil cloud account client and nil product client", func(t *testing.T) {
		// call function under test
		userService, err := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, nil, nil, mockBillingClient)
		assert.Error(t, err)
		assert.Nil(t, userService)
	})
}

func TestRegisterUser(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	// setup
	db, err := sql.Open("postgres", "postgres://oneapiint_so:6Z7AvB5X0IeLqU1@postgres5808-lb-fm-in.dbaas.intel.com:5432/oneapiint")
	assert.NoError(t, err)
	defer db.Close()
	ctx := context.Background()
	computeSvcClient := &idcComputeSvc.IDCServiceClient{}
	mockProductClient := NewMockProductClient(nil, nil, nil, nil, nil)
	mockCloudAccountSvcClient := NewMockCloudAccountSvcClient(nil, nil, nil)
	mockBillingClient := NewMockBillingClient(nil, nil)
	userService, _ := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)

	t.Run("nil db session", func(t *testing.T) {
		// make db session nil
		userService.session = nil

		// call the function under test
		request := &v1.TrainingRegistrationRequest{

			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)

		// restore db session
		userService.session = db
	})

	t.Run("nil product client", func(t *testing.T) {
		// make product client nil
		userService.productClient = nil

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)

		// restore product client
		userService.productClient = mockProductClient
	})

	t.Run("nil cloud account client", func(t *testing.T) {
		// make cloud account client nil
		userService.cloudAccountClient = nil

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)

		// restore cloud account client
		userService.cloudAccountClient = mockCloudAccountSvcClient
	})

	t.Run("successful execution", func(t *testing.T) {
		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockExpiryDate := "2023-10-04T00:00:00Z"
		mockUserId := "test-user"
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			bodyString := fmt.Sprintf(`{"expiry_date": "%s", "user_id": "%s"}`, mockExpiryDate, mockUserId)
			return 200, []byte(bodyString), nil
		}

		// mock store user registration
		oldStoreUserRegistration := StoreUserRegistration
		defer func() { StoreUserRegistration = oldStoreUserRegistration }()
		storeUserRegistrationCalls := 0
		StoreUserRegistration = func(
			ctx context.Context,
			tx *sql.Tx,
			in *v1.TrainingRegistrationRequest,
			ExpiryDate string,
			cloudAccountType v1.AccountType,
			products []*v1.Product,
			enterpriseId string,
			linuxUsername string,
			userEmail string,
			countryCode string,
		) error {
			storeUserRegistrationCalls++
			return nil
		}

		// mock get product catalog products
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockGetProductCatalogProductsWithFilter := func(ctx context.Context, filter *v1.ProductFilter) ([]*v1.Product, error) {
			return []*v1.Product{&mockProduct}, nil
		}
		userService.productClient = NewMockProductClient(nil, mockGetProductCatalogProductsWithFilter, nil, nil, nil)

		// mock get cloud account
		mockAccount := v1.CloudAccount{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		mockGetCloudAccount := func(ctx context.Context, accountId *v1.CloudAccountId) (*v1.CloudAccount, error) {
			return &mockAccount, nil
		}
		userService.cloudAccountClient = NewMockCloudAccountSvcClient(mockGetCloudAccount, nil, nil)

		// mock get GetCloudAccountCredits
		lastUpdatedTimestamp := &timestamppb.Timestamp{Seconds: 1711060625, Nanos: 884908000}
		expirationTimestamp := &timestamppb.Timestamp{Seconds: 1711060699}

		CreatedAtTimestamp := &timestamppb.Timestamp{Seconds: 1711060625, Nanos: 882938000}
		expirationTimestampBillingCredit := &timestamppb.Timestamp{Seconds: 1711060699}

		mockBillingCredit := &v1.BillingCredit{
			Created:         CreatedAtTimestamp,
			Expiration:      expirationTimestampBillingCredit,
			CloudAccountId:  "test-account-id",
			Reason:          v1.BillingCreditReason_CREDIT_INITIAL,
			OriginalAmount:  1.0,
			RemainingAmount: 1.0,
			CouponCode:      "test-coupon-code",
		}

		mockBillingCredits := []*v1.BillingCredit{mockBillingCredit}

		mockGetCloudAccountCredits := func(ctx context.Context, in *v1.BillingCreditFilter, opts ...grpc.CallOption) (*v1.BillingCreditResponse, error) {
			return &v1.BillingCreditResponse{
				TotalRemainingAmount: 1.0,
				TotalUsedAmount:      0.0,
				TotalUnAppliedAmount: 0.0,
				LastUpdated:          lastUpdatedTimestamp,
				ExpirationDate:       expirationTimestamp,
				Credits:              mockBillingCredits,
			}, nil
		}

		mockBillingCreditServiceClient := NewMockBillingCreditServiceClient(mockGetCloudAccountCredits)

		// Mock GetCouponExpiry
		CreatedAtTimestampCoupon := &timestamppb.Timestamp{Seconds: 1711060129}
		startTimestamp := &timestamppb.Timestamp{Seconds: 1710964973}
		expiresTimestamp := &timestamppb.Timestamp{Seconds: 1742500400}

		mockBillingCoupon := &v1.BillingCoupon{
			Code:        "test-coupon-code",
			Creator:     "test-id",
			Created:     CreatedAtTimestampCoupon,
			Start:       startTimestamp,
			Expires:     expiresTimestamp,
			Amount:      1.0,
			NumUses:     100,
			NumRedeemed: 1,
			IsStandard:  nil,
		}

		mockBillingCoupons := []*v1.BillingCoupon{mockBillingCoupon}

		mockGetCouponExpiry := func(ctx context.Context, in *v1.BillingCouponFilter, opts ...grpc.CallOption) (*v1.BillingCouponResponse, error) {
			return &v1.BillingCouponResponse{
				Coupons: mockBillingCoupons,
			}, nil
		}

		mockBillingCouponServiceClient := NewMockBillingCouponServiceClient(mockGetCouponExpiry)

		userService.billingClient = &BillingClient{
			BillingCreditServiceClient: mockBillingCreditServiceClient,
			BillingCouponServiceClient: mockBillingCouponServiceClient,
		}

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, makePOSTAPICallCalls)
		assert.Equal(t, 1, storeUserRegistrationCalls)
		assert.Equal(t, mockExpiryDate, response.ExpiryDate)
	})

	t.Run("request jupyter access", func(t *testing.T) {
		// mock get product catalog products
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockGetProductCatalogProductsWithFilter := func(ctx context.Context, filter *v1.ProductFilter) ([]*v1.Product, error) {
			return []*v1.Product{&mockProduct}, nil
		}
		userService.productClient = NewMockProductClient(nil, mockGetProductCatalogProductsWithFilter, nil, nil, nil)

		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockExpiryDate := "2023-10-04T00:00:00Z"
		mockUserId := "test-user"
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			bodyString := fmt.Sprintf(`{"expiry_date": "%s", "user_id": "%s"}`, mockExpiryDate, mockUserId)
			return 200, []byte(bodyString), nil
		}

		// mock store user registration
		// mock store user registration
		oldStoreUserRegistration := StoreUserRegistration
		defer func() { StoreUserRegistration = oldStoreUserRegistration }()
		storeUserRegistrationCalls := 0
		StoreUserRegistration = func(
			ctx context.Context,
			tx *sql.Tx,
			in *v1.TrainingRegistrationRequest,
			ExpiryDate string,
			cloudAccountType v1.AccountType,
			products []*v1.Product,
			enterpriseId string,
			linuxUsername string,
			userEmail string,
			countryCode string,
		) error {
			storeUserRegistrationCalls++
			return nil
		}

		// mock get cloud account
		mockAccount := v1.CloudAccount{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		mockGetCloudAccount := func(ctx context.Context, accountId *v1.CloudAccountId) (*v1.CloudAccount, error) {
			return &mockAccount, nil
		}
		userService.cloudAccountClient = NewMockCloudAccountSvcClient(mockGetCloudAccount, nil, nil)

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
			AccessType:     v1.AccessType_ACCESS_TYPE_JUPYTER,
		}
		response, err := userService.Register(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, makePOSTAPICallCalls)
		assert.Equal(t, 1, storeUserRegistrationCalls)
		assert.Equal(t, mockExpiryDate, response.ExpiryDate)
		assert.NotNil(t, response.JupyterLoginInfo)
		assert.GreaterOrEqual(t, len(*response.JupyterLoginInfo), 1)
	})

	t.Run("createSlurmBatchUser error", func(t *testing.T) {
		// mock get product catalog products
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockGetProductCatalogProductsWithFilter := func(ctx context.Context, filter *v1.ProductFilter) ([]*v1.Product, error) {
			return []*v1.Product{&mockProduct}, nil
		}
		userService.productClient = NewMockProductClient(nil, mockGetProductCatalogProductsWithFilter, nil, nil, nil)

		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			return 500, nil, errors.New("error making api call")
		}

		// mock store user registration
		oldStoreUserRegistration := StoreUserRegistration
		defer func() { StoreUserRegistration = oldStoreUserRegistration }()
		storeUserRegistrationCalls := 0
		StoreUserRegistration = func(
			ctx context.Context,
			tx *sql.Tx,
			in *v1.TrainingRegistrationRequest,
			ExpiryDate string,
			cloudAccountType v1.AccountType,
			products []*v1.Product,
			enterpriseId string,
			linuxUsername string,
			userEmail string,
			countryCode string,
		) error {
			storeUserRegistrationCalls++
			return nil
		}

		// mock get cloud account
		mockAccount := v1.CloudAccount{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		mockGetCloudAccount := func(ctx context.Context, accountId *v1.CloudAccountId) (*v1.CloudAccount, error) {
			return &mockAccount, nil
		}
		userService.cloudAccountClient = NewMockCloudAccountSvcClient(mockGetCloudAccount, nil, nil)

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, 1, makePOSTAPICallCalls)
		assert.Equal(t, 0, storeUserRegistrationCalls)
	})

	t.Run("StoreUserRegistration error", func(t *testing.T) {
		// mock get product catalog products
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockGetProductCatalogProductsWithFilter := func(ctx context.Context, filter *v1.ProductFilter) ([]*v1.Product, error) {
			return []*v1.Product{&mockProduct}, nil
		}
		userService.productClient = NewMockProductClient(nil, mockGetProductCatalogProductsWithFilter, nil, nil, nil)

		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockExpiryDate := "2023-10-04T00:00:00Z"
		mockUserId := "test-user"
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			bodyString := fmt.Sprintf(`{"expiry_date": "%s", "user_id": "%s"}`, mockExpiryDate, mockUserId)
			return 200, []byte(bodyString), nil
		}

		// mock store user registration
		oldStoreUserRegistration := StoreUserRegistration
		defer func() { StoreUserRegistration = oldStoreUserRegistration }()
		storeUserRegistrationCalls := 0
		StoreUserRegistration = func(
			ctx context.Context,
			tx *sql.Tx,
			in *v1.TrainingRegistrationRequest,
			ExpiryDate string,
			cloudAccountType v1.AccountType,
			products []*v1.Product,
			enterpriseId string,
			linuxUsername string,
			userEmail string,
			countryCode string,
		) error {
			storeUserRegistrationCalls++
			return errors.New("error storing user registration")
		}

		// mock get cloud account
		mockAccount := v1.CloudAccount{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		mockGetCloudAccount := func(ctx context.Context, accountId *v1.CloudAccountId) (*v1.CloudAccount, error) {
			return &mockAccount, nil
		}
		userService.cloudAccountClient = NewMockCloudAccountSvcClient(mockGetCloudAccount, nil, nil)

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, 1, makePOSTAPICallCalls)
		assert.Equal(t, 1, storeUserRegistrationCalls)
	})

	t.Run("GetCloudAccount error", func(t *testing.T) {
		// mock get product catalog products
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockGetProductCatalogProductsWithFilter := func(ctx context.Context, filter *v1.ProductFilter) ([]*v1.Product, error) {
			return []*v1.Product{&mockProduct}, nil
		}
		userService.productClient = NewMockProductClient(nil, mockGetProductCatalogProductsWithFilter, nil, nil, nil)

		// mock get cloud account error
		mockGetCloudAccount := func(ctx context.Context, accountId *v1.CloudAccountId) (*v1.CloudAccount, error) {
			return nil, errors.New("error getting cloud account")
		}
		userService.cloudAccountClient = NewMockCloudAccountSvcClient(mockGetCloudAccount, nil, nil)

		// call the function under test
		request := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}
		response, err := userService.Register(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestGetExpiryTimeById(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	// setup
	db, err := sql.Open("postgres", "postgres://oneapiint_so:6Z7AvB5X0IeLqU1@postgres5808-lb-fm-in.dbaas.intel.com:5432/oneapiint")
	assert.NoError(t, err)
	defer db.Close()
	ctx := context.Background()
	computeSvcClient := &idcComputeSvc.IDCServiceClient{}
	mockProductClient := NewMockProductClient(nil, nil, nil, nil, nil)
	mockCloudAccountSvcClient := NewMockCloudAccountSvcClient(nil, nil, nil)
	mockBillingClient := NewMockBillingClient(nil, nil)
	userService, _ := NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)

	t.Run("nil db session", func(t *testing.T) {
		// make db session nil
		userService.session = nil

		// call the function under test
		request := &v1.GetDataRequest{
			CloudAccountId: "test-account",
		}
		response, err := userService.GetExpiryTimeById(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)

		// restore db session
		userService, _ = NewTrainingBatchUserService(computeSvcClient, "slurm-batch-service", "slurm-jupyterhub-service", "slurm-ssh-service", db, mockProductClient, mockCloudAccountSvcClient, mockBillingClient)
	})

	t.Run("ReadExpiry error", func(t *testing.T) {
		// inject mock read expiry
		oldReadExpiry := ReadExpiry
		defer func() { ReadExpiry = oldReadExpiry }()
		readExpiryCalls := 0
		ReadExpiry = func(ctx context.Context, db *sql.DB, filter *v1.GetDataRequest, enterprise_id string) (*v1.GetDataResponse, error) {
			readExpiryCalls++
			return nil, errors.New("error reading expiry")
		}

		// call the function under test
		request := &v1.GetDataRequest{
			CloudAccountId: "test-account",
		}
		response, err := userService.GetExpiryTimeById(ctx, request)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, 1, readExpiryCalls)
	})

	t.Run("successful operation", func(t *testing.T) {
		exampleTimestamp := "2023-10-04T00:00:00Z"
		expectedResponse := "2023-10-04" // output date is formatted to be YYYY-MM-DD format

		// inject mock read expiry
		oldReadExpiry := ReadExpiry
		defer func() { ReadExpiry = oldReadExpiry }()
		readExpiryCalls := 0
		ReadExpiry = func(ctx context.Context, db *sql.DB, filter *v1.GetDataRequest, enterprise_id string) (*v1.GetDataResponse, error) {
			readExpiryCalls++
			return &v1.GetDataResponse{ExpiryDate: exampleTimestamp}, nil
		}

		// call the function under test
		request := &v1.GetDataRequest{
			CloudAccountId: "test-account",
		}
		response, err := userService.GetExpiryTimeById(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, readExpiryCalls)
		assert.Equal(t, expectedResponse, response.ExpiryDate)
	})
}

func TestCreateSlurmBatchUser(t *testing.T) {
	t.Run("successful api call", func(t *testing.T) {
		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockExpiryDate := "2023-10-04T00:00:00Z"
		mockUserId := "test-user"
		mockCouponExpirationDate := time.Now()
		mockLaunchStartTime := time.Now()
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			bodyString := fmt.Sprintf(`{"expiry_date": "%s", "user_id": "%s", "coupon_expiration": "%s", "launch_start_time": "%s"}`, mockExpiryDate, mockUserId, mockCouponExpirationDate.Format(time.RFC3339), mockLaunchStartTime.Format(time.RFC3339))
			return 200, []byte(bodyString), nil
		}

		// call the function under test
		ctx := context.Background()
		slurmUserResponse, err := createSlurmBatchUser(ctx, "slurm-batch-service", "test-account", [][]byte{[]byte("test-pubkey")}, mockCouponExpirationDate)
		assert.NoError(t, err)
		assert.NotNil(t, slurmUserResponse)
		assert.Equal(t, 1, makePOSTAPICallCalls)
		assert.Equal(t, mockExpiryDate, slurmUserResponse.ExpiryDate)
		assert.Equal(t, mockUserId, slurmUserResponse.UserId)
		assert.Equal(t, mockCouponExpirationDate.Format(time.RFC3339), slurmUserResponse.CouponExpiration)
		assert.Equal(t, mockLaunchStartTime.Format(time.RFC3339), slurmUserResponse.LaunchStartTime)
	})

	t.Run("api call error", func(t *testing.T) {
		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockCouponExpirationDate := time.Now()
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			return 500, nil, errors.New("error making api call")
		}

		// call the function under test
		ctx := context.Background()
		slurmUserResponse, err := createSlurmBatchUser(ctx, "slurm-batch-service", "test-account", [][]byte{[]byte("test-pubkey")}, mockCouponExpirationDate)
		assert.Error(t, err)
		assert.Nil(t, slurmUserResponse)
		assert.Equal(t, 1, makePOSTAPICallCalls)
	})

	t.Run("non-200 response", func(t *testing.T) {
		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockCouponExpirationDate := time.Now()
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			return 304, nil, nil
		}

		// call the function under test
		ctx := context.Background()
		slurmUserResponse, err := createSlurmBatchUser(ctx, "slurm-batch-service", "test-account", [][]byte{[]byte("test-pubkey")}, mockCouponExpirationDate)
		assert.Error(t, err)
		assert.Nil(t, slurmUserResponse)
		assert.Equal(t, 1, makePOSTAPICallCalls)
	})

	t.Run("nil response", func(t *testing.T) {
		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockCouponExpirationDate := time.Now()
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			return 200, nil, nil
		}

		// call the function under test
		ctx := context.Background()
		slurmUserResponse, err := createSlurmBatchUser(ctx, "slurm-batch-service", "test-account", [][]byte{[]byte("test-pubkey")}, mockCouponExpirationDate)
		assert.Error(t, err)
		assert.Nil(t, slurmUserResponse)
		assert.Equal(t, 1, makePOSTAPICallCalls)
	})

	t.Run("json unmarshal error", func(t *testing.T) {
		// mock make post api call
		oldMakePOSTAPICall := MakePOSTAPICall
		defer func() { MakePOSTAPICall = oldMakePOSTAPICall }()
		makePOSTAPICallCalls := 0
		mockCouponExpirationDate := time.Now()
		MakePOSTAPICall = func(ctx context.Context, server string, uri string, auth string, payload []byte) (int, []byte, error) {
			makePOSTAPICallCalls++
			bodyString := `"this is not json"`
			return 200, []byte(bodyString), nil
		}

		// call the function under test
		ctx := context.Background()
		slurmUserResponse, err := createSlurmBatchUser(ctx, "slurm-batch-service", "test-account", [][]byte{[]byte("test-pubkey")}, mockCouponExpirationDate)
		assert.Error(t, err)
		assert.Nil(t, slurmUserResponse)
		assert.Equal(t, 1, makePOSTAPICallCalls)
	})
}
