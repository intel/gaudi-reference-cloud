// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/common/httpclient"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/database/query"
	idcComputeSvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_compute"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
)

// injectible helper functions for mocking during unit tests
var (
	ReadExpiry            = query.ReadExpiry
	StoreUserRegistration = query.StoreUserRegistration
	MakePOSTAPICall       = httpclient.MakePOSTAPICall
)

// TrainingBatchUserService is used to implement helloworld.GreeterServer.
type TrainingBatchUserService struct {
	pb.UnimplementedTrainingBatchUserServiceServer
	IDCComputeServiceClient *idcComputeSvc.IDCServiceClient
	SlurmBatchService       string
	SlurmJupyterhubService  string
	SlurmSSHService         string
	session                 *sql.DB
	productClient           ProductClientInterface
	cloudAccountClient      CloudAccountSvcClientInterface
	billingClient           BillingClientInterface
}

type slurmUserResponse struct {
	ExpiryDate       string `json:"expiry_date"`
	UserId           string `json:"user_id"`
	CouponExpiration string `json:"coupon_expiration"`
	LaunchStartTime  string `json:"launch_start_time"`
}

const (
	createUserURI = "/user"
)

func NewTrainingBatchUserService(computeSvcClient *idcComputeSvc.IDCServiceClient, slurmBatchService string, slurmJupyterhubService string, slurmSSHService string, session *sql.DB, productClient ProductClientInterface, cloudAccountClient CloudAccountSvcClientInterface, billingClient BillingClientInterface) (*TrainingBatchUserService, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	if computeSvcClient == nil {
		return nil, fmt.Errorf("compute service client is required")
	}
	if slurmBatchService == "" {
		return nil, fmt.Errorf("slurm batch service endpoint is required")
	}
	if slurmJupyterhubService == "" {
		return nil, fmt.Errorf("slurm JupyterHub service endpoint is required")
	}
	if slurmSSHService == "" {
		return nil, fmt.Errorf("slurm SSH service endpoint is required")
	}
	if productClient == nil {
		return nil, fmt.Errorf("product client is required")
	}
	if cloudAccountClient == nil {
		return nil, fmt.Errorf("cloud account client is required")
	}
	if billingClient == nil {
		return nil, fmt.Errorf("billing account client is required")
	}
	return &TrainingBatchUserService{
		IDCComputeServiceClient: computeSvcClient,
		SlurmBatchService:       slurmBatchService,
		SlurmJupyterhubService:  slurmJupyterhubService,
		SlurmSSHService:         slurmSSHService,
		session:                 session,
		productClient:           productClient,
		cloudAccountClient:      cloudAccountClient,
		billingClient:           billingClient,
	}, nil
}

func (batchSrv *TrainingBatchUserService) Register(ctx context.Context, in *v1.TrainingRegistrationRequest) (*v1.TrainingRegistrationResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TrainingBatchUserService.Register").WithValues("cloudAccountId", in.CloudAccountId, "trainingId", in.TrainingId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := batchSrv.session
	if dbSession == nil {
		logger.Info("no database connection found")
		return nil, fmt.Errorf("Our services are currently unable to connect to the database. Please try your request again later.")
	}

	catalogConnection := batchSrv.productClient
	if catalogConnection == nil {
		logger.Info("no product catalog connection found")
		return nil, fmt.Errorf("Our services are currently unable to locate your training. Please try again later.")
	}

	cloudAccountConnection := batchSrv.cloudAccountClient
	if cloudAccountConnection == nil {
		logger.Info("no cloud account service connection found")
		return nil, fmt.Errorf("Our services are currently unable to find your account. Please try again later.")
	}

	pubKeys := [][]byte{}
	for _, keyname := range in.SshKeyNames {
		pubKey, err := batchSrv.IDCComputeServiceClient.GetPublicKey(ctx, in.CloudAccountId, keyname)
		if err != nil || pubKey == nil {
			logger.Info("ssh verification failed", "error", err)
			return nil, fmt.Errorf("Our services are currently unable to find your SSH key. Please check your SSH keys.")
		}
		pubKeys = append(pubKeys, pubKey)
	}

	logger.V(9).Info("debug", "pubkeys", pubKeys)

	productLookup := ""
	testLookup, err := batchSrv.productClient.GetProductCatalogProductsWithFilter(ctx, &pb.ProductFilter{
		Id: &in.TrainingId,
	})

	if err != nil {
		logger.Error(err, "error getting product catalog products with filter")
	}

	if len(testLookup) == 0 {
		productLookup = ""
	} else if len(testLookup) == 1 {
		_, product := testLookup[0], testLookup[0]
		dirPath, exists := (product.GetMetadata())["launch.dirPath"]
		if exists {
			productLookup = dirPath
		} else {
			productLookup = ""
		}
	} else {
		productLookup = ""
	}

	enterprise_id, countryCode, err := grpcutil.ExtractEnterpriseIDAndCountryCodefromCtx(ctx, false)
	if err != nil {
		logger.Info("enterpriseId and/or country code missing from jwt token", "error", err)
		return nil, fmt.Errorf("Our services are currently unable to verify your account information. Please try again later.")
	}

	user_email, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		logger.Info("user email missing from jwt token", "error", err)
		return nil, fmt.Errorf("Our services are currently unable to verify your account information. Please try again later.")
	}

	cloudAccount, err := cloudAccountConnection.GetCloudAccount(ctx, &pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		logger.Info("error finding cloud account", "error", err)
		return nil, fmt.Errorf("Our services are currently unable to verify your account information. Please try again later.")
	}

	var couponExpirationTime time.Time
	defaultTime := "1970-01-01T00:00:00Z"
	cloudAccountType := cloudAccount.GetType()

	// Call billing service to get credits using cloud account
	cloudAccountCreditsRes, err := batchSrv.billingClient.GetCloudAccountCredits(ctx, in.CloudAccountId, true)
	if err != nil {
		logger.Info("error finding cloud_account_credits from cloud account", "error", err)
		return nil, fmt.Errorf("Our services are currently unable to verify your account information. Please try again later.")
	}

	for _, cloudAccountCredits := range cloudAccountCreditsRes {

		// skip credit if the coupon has been cancelled manually in the backend
		// CouponCode becomes a description message and doesn't look like a coupon if it's been cancelled
		validCouponRegex, _ := regexp.Compile("[0-9A-Z]{4}-[0-9A-Z]{4}-[0-9A-Z]{4}")
		if !validCouponRegex.MatchString(cloudAccountCredits.CouponCode) {
			continue
		}

		if couponExpirationTime.Before(time.Unix(cloudAccountCredits.GetExpiration().Seconds, 0)) {
			couponExpirationTime = time.Unix(cloudAccountCredits.GetExpiration().Seconds, 0)
		}

	}

	if (couponExpirationTime == time.Time{} || couponExpirationTime.Before(time.Now())) {
		if cloudAccountType == v1.AccountType_ACCOUNT_TYPE_INTEL || cloudAccountType == v1.AccountType_ACCOUNT_TYPE_ENTERPRISE || cloudAccountType == v1.AccountType_ACCOUNT_TYPE_MEMBER {
			logger.Info("Intel/Enterpise/Member account found, adding extra time")
			couponExpirationTime = time.Now().Add(30 * 24 * time.Hour)

		} else if (couponExpirationTime == time.Time{}) {
			logger.Info("no valid coupon found")
			parsedDefaultTime, _ := time.Parse(time.RFC3339, defaultTime)
			couponExpirationTime = parsedDefaultTime

		}
	}

	logger.Info("final coupon expiration time", "couponExpirationTime", couponExpirationTime)

	hashed := fmt.Sprintf("%x", sha256.Sum256([]byte(enterprise_id)))

	// Local testing values
	// hashed := fmt.Sprintf("%x", sha256.Sum256([]byte("test_user")))
	// cloudAccountType := pb.AccountType_ACCOUNT_TYPE_PREMIUM

	userResp, err := createSlurmBatchUser(ctx, batchSrv.SlurmBatchService, hashed[:31], pubKeys, couponExpirationTime)
	if err != nil {
		logger.Info("error creating user")
		return nil, fmt.Errorf("Due to a high volume of requests, our services are currently unable to create your user. Please try again later.")
	}

	sshLoginInfo := fmt.Sprintf("ssh %s@%s", userResp.UserId, batchSrv.SlurmSSHService)
	jupyterLink := fmt.Sprintf("%s/hub/login?next=/user-redirect/lab/tree%s", batchSrv.SlurmJupyterhubService, productLookup)
	if productLookup == "" {
		jupyterLink = fmt.Sprintf("%s/hub/login", batchSrv.SlurmJupyterhubService)
	}
	message := fmt.Sprintf("customize message for training %s", in.TrainingId)
	logger.Info(sshLoginInfo)

	linux_username := userResp.UserId

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		logger.Info("error logging transaction data for database")
		return nil, fmt.Errorf("Our services are currently unable to register your account. Please try again later.")

	}
	defer tx.Rollback()
	if err := StoreUserRegistration(ctx, tx, in, userResp.CouponExpiration, cloudAccountType, testLookup, enterprise_id, linux_username, user_email, countryCode); err != nil {
		logger.Info("error adding user data to database")
		return nil, fmt.Errorf("Our services are currently unable to store your account. Please try again later.")

	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error in commiting transaction")
	}

	responseExpiryDate := userResp.CouponExpiration
	if userResp.CouponExpiration == defaultTime {
		responseExpiryDate = "no-expiration-found"
	}

	if in.AccessType == pb.AccessType_ACCESS_TYPE_JUPYTER {
		return &v1.TrainingRegistrationResponse{
			JupyterLoginInfo: &jupyterLink,
			Message:          message,
			ExpiryDate:       responseExpiryDate,
		}, nil
	}
	return &v1.TrainingRegistrationResponse{
		SshLoginInfo: &sshLoginInfo,
		Message:      message,
		ExpiryDate:   responseExpiryDate,
	}, nil
}

func (batchSrv *TrainingBatchUserService) GetExpiryTimeById(ctx context.Context, in *v1.GetDataRequest) (*v1.GetDataResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TrainingBatchUserService.GetExpiryTimeById").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := batchSrv.session
	if dbSession == nil {
		logger.Info("database connection error")
		return nil, fmt.Errorf("Our services are currently unable to connect to the database. Please try again later.")
	}

	enterprise_id, _, err := grpcutil.ExtractEnterpriseIDAndCountryCodefromCtx(ctx, false)
	if err != nil {
		logger.Info("error finding enterprise_id")
		return nil, fmt.Errorf("Our services are currently unable to verify your account information. Please try again later.")
	}

	expiryDate, err := ReadExpiry(ctx, dbSession, in, enterprise_id)
	if err != nil {
		logger.Error(err, "error reading expiry")
		return nil, fmt.Errorf("Our services are currently unable to read your account details. Please try again later.")
	}

	if expiryDate.GetExpiryDate() == "" {
		return expiryDate, fmt.Errorf("Our services are currently experiencing an interruption. We apologize for the inconvenience. Please try your request again later.")
	}

	defaultTime := "1970-01-01T00:00:00Z"
	if expiryDate.GetExpiryDate() == defaultTime {
		return &v1.GetDataResponse{
			ExpiryDate: "no-expiration-found",
		}, nil
	}

	newDate, err := time.Parse(time.RFC3339, expiryDate.GetExpiryDate())
	return &v1.GetDataResponse{
		ExpiryDate: newDate.Format("2006-01-02"),
	}, nil

}

func createSlurmBatchUser(ctx context.Context, slurmBatchSvc, cloudAccountId string, pubKeys [][]byte, couponExpirationTimeArg time.Time) (*slurmUserResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TrainingBatchUserService.createSlurmBatchUser").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	sshKeysList := ""
	for _, key := range pubKeys {
		sshKeysList += string(key)
		sshKeysList += "\n"
	}

	type slurmUserCreateParams struct {
		CloudAccountId   string `json:"cloudAccountId"`
		SSHKey           string `json:"sshPublicKey"`
		CouponExpiration string `json:"couponExpiration"`
	}

	userParams := slurmUserCreateParams{
		CloudAccountId:   cloudAccountId,
		SSHKey:           base64.StdEncoding.EncodeToString([]byte(sshKeysList)),
		CouponExpiration: couponExpirationTimeArg.Format(time.RFC3339),
	}

	serializePayload, err := json.MarshalIndent(userParams, "", "    ")
	if err != nil {
		logger.Info("error serializing payload")
		return nil, fmt.Errorf("Our services are currently unable to create your training account. Please try again later.")
	}
	logger.Info("creating user with slurm batch user service", "payload", string(serializePayload))

	retCode, response, err := MakePOSTAPICall(ctx, slurmBatchSvc, createUserURI, "", serializePayload)
	if err != nil {
		logger.Info("error making post call")
		return nil, fmt.Errorf("Due to a high volume of requests, our services are currently unable to process your training request. Please try again later.")
	}
	if retCode != http.StatusOK {
		logger.Info("un-expected error code from slurm batch user service")
		return nil, fmt.Errorf("Due to a high volume of requests, our services are currently unable to process your training request. Please try again later.")
	}
	if response == nil {
		logger.Info("empty response from slurm batch user service")
		return nil, fmt.Errorf("Due to a high volume of requests, our services are currently unable to process your training request. Please try again later.")
	}

	batchUserResp := slurmUserResponse{}

	if err := json.Unmarshal(response, &batchUserResp); err != nil {
		logger.Error(err, "error unmarshaling batch service backend response")
		return nil, fmt.Errorf("Due to a high volume of requests, our services are currently unable to process your training request. Please try again later.")
	}

	return &batchUserResp, nil
}
