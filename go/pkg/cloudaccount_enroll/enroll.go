// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount_enroll

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	icp "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount_enroll/icpintel"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"k8s.io/utils/strings/slices"
)

type EnrollConfig struct {
	ListenPort uint16     `koanf:"listenPort"`
	ICPConfig  icp.Config `koanf:"icp"`
	Authz      struct {
		Enabled bool `koanf:"enabled"`
	} `koanf:"authz"`
}

func (cc *EnrollConfig) GetListenPort() uint16 {
	return cc.ListenPort
}

func (cc *EnrollConfig) SetListenPort(port uint16) {
	cc.ListenPort = port
}

type CloudAccountEnrollService struct {
	pb.UnimplementedCloudAccountEnrollServiceServer
	cfg *EnrollConfig
}

type enrollRequestData struct {
	Email        string
	Tid          string
	EnterpriseId string
	Groups       []string
	Idp          string
	CountryCode  string
}

const defaultCreditcouponCode = "defaultCredits"

// This function attempts to detect an enterprise user based on IDP claim in the token.
// Note: This function does not work for an intel account.
func (data *enrollRequestData) isEnterpriseUser(ctx context.Context) bool {
	logger := log.FromContext(ctx)
	//This is the list of all b2c domain names used by IDP.
	b2cIDP := []string{"intelcorpintb2c.onmicrosoft.com", "intelcorpb2c.onmicrosoft.com", "intelcorpdevb2c.onmicrosoft.com"}
	if slices.Contains(b2cIDP, data.Idp) {
		logger.Info("non-Enterprise user details", "oid", data.EnterpriseId, "idp", data.Idp)
		return false
	} else {
		//The Idp claim will be changed only after the enterpise user is approved in eRPM.
		logger.Info("enterprise user details", "oid", data.EnterpriseId, "idp", data.Idp)
		return true
	}
}

func isIntelUser(email string) bool {
	intelSuffs := []string{"@intel.com", "@corpint.intel.com"}
	for _, suff := range intelSuffs {
		if strings.HasSuffix(email, suff) {
			return true
		}
	}
	return false
}

func isMember(ctx context.Context, email string) (bool, error) {
	logger := log.FromContext(ctx)
	memberAccount, err := cloudacctMemberClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{UserName: email})
	logger.V(9).Info("enroll API invoked for ReadUserCloudAccounts to check if the user is member", "memberAccount", memberAccount)
	if err == nil && len(memberAccount.MemberAccount) != 0 {
		return true, err
	}
	return false, err
}

func updatePersonIdForMember(ctx context.Context, memberAccount []*pb.MemberCloudAccount, memberEmail string, enterpriseId string) error {
	logger := log.FromContext(ctx)
	logger.Info("update Person ID in the members table for given email")

	for _, member := range memberAccount {
		if member.Name == memberEmail {
			// memberEmail is the owner
			return nil
		} else {
			// memberEmail not the owner - fetch the personID ICP.
			personId, err := icpClient.GetPersonId(ctx, memberEmail, enterpriseId)
			// Update Members table with PersonID.
			cloudacctMemberClient.UpdatePersonId(ctx, &pb.MemberPersonId{PersonId: personId, MemberEmail: memberEmail})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (ces *CloudAccountEnrollService) Enroll(ctx context.Context, req *pb.CloudAccountEnrollRequest) (*pb.CloudAccountEnrollResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.Enroll").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var reqData *enrollRequestData

	// Extract info from JWT token
	reqData, err := extractRequestfromCtx(ctx)
	if err != nil {
		logger.Error(err, "failed to get required data from JWT token")
		return nil, status.Errorf(codes.InvalidArgument, "Failed to get required data from the JWT token")
	}
	span.SetAttributes(attribute.String("enterpriseId", reqData.EnterpriseId))

	logger = logger.WithValues("enterpriseId", reqData.EnterpriseId)
	logger.Info("enroll API invoked", "countryCode", reqData.CountryCode)
	resp := pb.CloudAccountEnrollResponse{
		Registered: true,
	}

	cloudAccountExists, err := cloudacctClient.CheckCloudAccountExists(ctx, &pb.CloudAccountName{Name: reqData.Email})

	if err != nil {
		logger.Error(err, "failed to check cloudAccount exists")
		return nil, err
	}

	logger.Info("check for cloudaccountexists completed", "cloudAccountExists", cloudAccountExists.Value)

	if !cloudAccountExists.Value && !req.TermsStatus {
		memberAccount, err := cloudacctMemberClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{UserName: reqData.Email})
		logger.V(9).Info("enroll API invoked for ReadUserCloudAccounts", "memberAccount", memberAccount)
		if err == nil && len(memberAccount.MemberAccount) != 0 {

			if ces.cfg.Authz.Enabled {
				logger.V(9).Info("authz enabled, assigning default cloud account roles")
				ces.AssignDefaultCloudAccountRole(ctx, memberAccount, logger)
			}

			resp := &pb.CloudAccountEnrollResponse{
				Registered:         true,
				HaveCloudAccount:   false,
				HaveBillingAccount: false,
				Enrolled:           true,
				Action:             pb.CloudAccountEnrollAction_ENROLL_ACTION_NONE,
				CloudAccountId:     "",
				CloudAccountType:   pb.AccountType_ACCOUNT_TYPE_MEMBER,
				IsMember:           true,
			}

			// Update Person ID in the members table for GTS check
			if err := updatePersonIdForMember(ctx, memberAccount.MemberAccount, reqData.Email, reqData.EnterpriseId); err != nil {
				logger.Error(err, "error updating personId for member", "enterpriseId", reqData.EnterpriseId)
				return nil, err
			}

			return resp, err
		}

	}
	currentAccDetails, err := cloudacctClient.GetByOid(ctx, &pb.CloudAccountOid{Oid: reqData.EnterpriseId})
	acctType := pb.AccountType_ACCOUNT_TYPE_STANDARD // default account type
	personId := ""
	if err == nil {
		// Cloud account already exists no need to fetch person id.
		// The account type needs to be changed only in the case of
		// 		a. enterprise approved or enterprise pending.
		//		b. Explicit request to upgrade to Premium has been disabled for now.
		acctType = currentAccDetails.Type
		personId = currentAccDetails.PersonId
	} else if status.Code(err) == codes.NotFound {
		// Cloud account does not exist
		// Notify IDC Console, if T&C not yet accepted
		// check if the account is member or not

		member, err := isMember(ctx, reqData.Email)
		if err != nil {
			logger.Error(err, "error validating reqData")
			return nil, err
		}
		if !req.TermsStatus && !member {
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_TC
			return &resp, nil
		}

		// Compute the acctType
		if isIntelUser(reqData.Email) {
			logger.Info("cloud account does not exist, intel user found", "oid", reqData.EnterpriseId)
			acctType = pb.AccountType_ACCOUNT_TYPE_INTEL
		} else {
			// Fetch the personId from ICP/eRPM
			logger.Info("cloud account does not exist, attempting to fetch the personId field from eRPM", "oid", reqData.EnterpriseId)
			isEnterprisePending, entId, err := enterprisePendingDetails(ctx, reqData)
			if err != nil {
				logger.Error(err, "failed eRPM APIs to check for pending enterprise enrollment", "enterpriseId", reqData.EnterpriseId)
				return nil, err
			}
			personId = entId
			if isEnterprisePending {
				acctType = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING
			}
		}
	} else {
		logger.Info("failed to fetch CloudAccount details ", "oid", reqData.EnterpriseId)
		return nil, err
	}

	// Create/Update the Cloud account.
	logger.Info("attempting to create/update cloud account ", "oid", reqData.EnterpriseId, "type", acctType, "personId", personId,
		"countryCode", reqData.CountryCode)
	currentAccDetails, err = cloudacctClient.Ensure(ctx, &pb.CloudAccountCreate{
		Tid:         reqData.Tid,
		Oid:         reqData.EnterpriseId,
		Name:        reqData.Email,
		Owner:       reqData.Email,
		Type:        acctType,
		PersonId:    personId,
		CountryCode: reqData.CountryCode,
	})
	if err != nil {
		logger.Error(err, "cloudaccount ensure failed", "oid", reqData.EnterpriseId)
		resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
		return &resp, nil
	}
	resp.HaveCloudAccount = true
	resp.CloudAccountId = currentAccDetails.Id
	resp.CloudAccountType = currentAccDetails.Type
	resp.IsMember = false

	memberAccount, err := cloudacctMemberClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{UserName: reqData.Email})
	logger.V(9).Info("enroll API invoked for ReadUserCloudAccounts", "memberAccount", memberAccount)
	// Check if the cloudaccount is member to other cloudaccount
	logger.V(9).Info("enroll API invoked for ReadUserCloudAccounts to check if the user is member", "memberAccount", memberAccount)
	if len(memberAccount.MemberAccount) > 1 {
		for _, mem := range memberAccount.MemberAccount {
			if mem.Name == reqData.Email {
				logger.Info("cloudaccount is a member", "cloudAccountId", resp.CloudAccountId)
				resp.IsMember = true
			}
		}
	} else if err != nil {
		logger.Error(err, "cloudaccount check failed", "cloudAccountId", resp.CloudAccountId)
		return nil, err
	}

	enrollSteps(ctx, reqData, currentAccDetails, &resp)
	logger.Info("update cloud account success after enroll steps ", "oid", reqData.EnterpriseId, "type", acctType, "personId", personId,
		"countryCode", reqData.CountryCode, "cloudAccountId", resp.CloudAccountId)
	_, err = cloudacctClient.Ensure(ctx, &pb.CloudAccountCreate{
		Tid:                   reqData.Tid,
		Oid:                   reqData.EnterpriseId,
		Name:                  reqData.Email,
		Owner:                 reqData.Email,
		Type:                  resp.CloudAccountType,
		PersonId:              personId,
		CountryCode:           reqData.CountryCode,
		Enrolled:              &resp.Enrolled,
		BillingAccountCreated: &resp.HaveBillingAccount,
	})

	if err != nil {
		logger.Error(err, "cloudaccount ensure failed", "oid", reqData.EnterpriseId, "cloudAccountId", resp.CloudAccountId)
		resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
		return &resp, nil
	}

	span.SetAttributes(attribute.String("cloudAccountId", resp.CloudAccountId))

	if ces.cfg.Authz.Enabled {
		logger.V(9).Info("authz enabled, assigning default cloud account roles")
		ces.AssignDefaultCloudAccountRole(ctx, memberAccount, logger)
	}

	logger.Info("enroll completed for ", "oid", reqData.EnterpriseId, "cloudAccountId", resp.CloudAccountId,
		"cloudAccountType", resp.CloudAccountType, "enrollAction", resp.Action)
	return &resp, nil

}

func (*CloudAccountEnrollService) AssignDefaultCloudAccountRole(ctx context.Context, memberAccount *pb.MemberAccount, logger logr.Logger) {
	// for each cloud account returned from the ReadUserCloudAccounts call, assign the default role if not yet assigned
	for _, account := range memberAccount.MemberAccount {
		assigned, err := authzClient.DefaultCloudAccountRoleAssigned(ctx, &pb.DefaultCloudAccountRoleAssignedRequest{CloudAccountId: account.Id})
		if err != nil {
			logger.Error(err, "unable to check if default role is assigned", "cloudAccountId", account.Id)
			continue
		}

		// should skip if the default role is already assigned
		if assigned {
			logger.Info("default role already assigned", "cloudAccountId", account.Id)
			continue
		}

		// get the active members for the cloud account, excluding the owner
		members, err := cloudacctMemberClient.ReadActiveMembers(ctx, &pb.CloudAccountId{Id: account.Id})

		if err != nil {
			logger.Error(err, "unable to get the members", "cloudAccountId", account.Id)
			continue
		}

		if members.Members == nil {
			members.Members = make([]string, 0)
		}

		// assign system role for admin and members and create the default cloud account roles with wilcard permissions to all resources
		_, err = authzClient.AssignDefaultCloudAccountRole(ctx, &pb.AssignDefaultCloudAccountRoleRequest{
			CloudAccountId: account.Id,
			Members:        members.Members,
			Admins:         []string{account.Owner},
		})

		if err != nil {
			logger.Error(err, "unable to assign the default role", "cloudAccountId", account.Id)
			continue
		}
	}
}

func extractRequestfromCtx(ctx context.Context) (*enrollRequestData, error) {
	// Attempt to extract the headers.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("unable to read the headers from the grpc request")
	}
	autheader := md.Get("authorization")
	if len(autheader) == 0 {
		return nil, errors.New("JWT token is not passed along with the request")
	} else {
		//NOTE: Token validation is perfomed here since it is already verified at the ingress.
		jwtToken := strings.Replace(autheader[0], "Bearer ", "", 1)
		token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
		if err != nil {
			return nil, errors.New("error while parsing the token")
		}
		return extractRequestFromToken(ctx, token)
	}
}

func extractRequestFromToken(ctx context.Context, token_p *jwt.Token) (*enrollRequestData, error) {
	claims, ok := token_p.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("unable to read claims from the JWT token")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email is not part of the JWT token")
	}
	tid, ok := claims["tid"].(string)
	if !ok || tid == "" {
		return nil, fmt.Errorf("TID is not part of the JWT token with email %v ", email)
	}
	id, ok := claims["enterpriseId"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("enterpriseId is not part of the JWT token with email %v ", email)
	}
	countryCode, ok := claims["countryCode"].(string)
	// Checking only for the external customers if the country code is null
	if !isIntelUser(email) {
		if !ok || countryCode == "" {
			return nil, fmt.Errorf("countryCode is not part of the JWT token with email %v ", email)
		}
	} else {
		countryCode = ""
	}
	idp, ok := claims["idp"].(string)
	if !ok || idp == "" {
		return nil, fmt.Errorf("idp is not part of the JWT token with email %v ", email)
	}
	aInterface, ok := claims["groups"].([]interface{})
	roles := make([]string, len(aInterface))
	for i, v := range aInterface {
		roles[i] = v.(string)
	}
	return &enrollRequestData{
		Tid:          tid,
		Email:        email,
		EnterpriseId: id,
		Groups:       roles,
		Idp:          idp,
		CountryCode:  countryCode,
	}, nil
}

// This function returns true only if user is enterprise pending.
func enterprisePendingDetails(ctx context.Context, obj *enrollRequestData) (bool, string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.GetICPData").WithValues("enterpriseId", obj.EnterpriseId).Start()
	defer span.End()

	// This function calls eRPM APIs to check for pending enterprise enrollment
	isPending, personid, err := icpClient.IsEnterprisePending(ctx, obj.Email, obj.EnterpriseId)
	if err != nil {
		//log the error.
		logger.Error(err, "failed to get user information  (isPending and personId)  icpClient.IsEnterprisePending ")
		return false, "", err
	}
	return isPending, personid, nil
}

func enrollSteps(ctx context.Context, obj *enrollRequestData,
	acct *pb.CloudAccount, resp *pb.CloudAccountEnrollResponse) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("enrollSteps").WithValues("cloudAccountId", acct.Id).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	// Perform conditional billing account creation
	if !acct.BillingAccountCreated {
		_, err := billingAcctClient.Create(ctx, &pb.BillingAccount{
			CloudAccountId: acct.GetId(),
		})
		if err != nil {
			logger.Error(err, "billing account create failed", "oid", obj.EnterpriseId,
				"cloudAccountId", acct.GetId())
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
			return
		}
		logger.Info("billing Account created for cloudAccountId", "cloudAccountId", acct.GetId())
	}
	resp.HaveBillingAccount = true

	if !acct.Enrolled && acct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		creds, err := cloudCreditsCreditServiceClient.ReadInternal(ctx, &pb.Account{CloudAccountId: acct.GetId()})
		if err != nil {
			logger.Error(err, "read credits failed", "oid", obj.EnterpriseId,
				"cloudAccountId", acct.GetId())
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
			return
		}

		// check if the credit was acquired as a result of coupon redemption
		for {
			out, err := creds.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				logger.Error(err, "recv error in Read")
				resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
				return
			}
			if couponCode := out.CouponCode; couponCode != "" && couponCode != defaultCreditcouponCode {
				resp.Enrolled = true
				break
			}

			logger.Info("skipping", "couponCode", out.CouponCode)
		}

		// check if the BillingOption was setup
		if !resp.Enrolled {
			billingOption, err := billingOptionsClient.Read(ctx, &pb.BillingOptionFilter{CloudAccountId: &acct.Id})
			if err != nil {
				logger.Error(err, "read billing options failed", "oid", obj.EnterpriseId,
					"cloudAccountId", acct.GetId())
				resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
				return
			}

			if len(billingOption.CloudAccountId) > 0 {
				resp.Enrolled = true
			}
		}

		if !resp.Enrolled {
			logger.Error(fmt.Errorf("error Premium account with no payment method enrolled false"), "Premium account with no payment method", "cloudAccountId", resp.CloudAccountId)
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_COUPON_OR_CREDIT_CARD
			return
		}
	}
	resp.Enrolled = true
}

type Service struct {
	icpConfig *icp.ICPConfig
}

var (
	billingAcctClient               pb.BillingAccountServiceClient
	billingOptionsClient            pb.BillingOptionServiceClient
	billingCreditServiceClient      pb.BillingCreditServiceClient
	billingCouponServiceClient      pb.BillingCouponServiceClient
	cloudacctClient                 pb.CloudAccountServiceClient
	cloudacctMemberClient           pb.CloudAccountMemberServiceClient
	cloudCreditsCouponServiceClient pb.CloudCreditsCouponServiceClient
	cloudCreditsCreditServiceClient pb.CloudCreditsCreditServiceClient
	icpClient                       icp.ICPClient
	authzClient                     *authz.AuthzClient
)

func (svc *Service) Init(ctx context.Context, cfg *EnrollConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	// Load the ICP configuration.
	if svc.icpConfig == nil {
		var err error
		svc.icpConfig, err = icp.New(ctx, &cfg.ICPConfig)
		if err != nil {
			return err
		}
	}

	// Instantiate the icpClient
	icpClient = icp.CreateICPClient(svc.icpConfig)
	if err := svc.initEnrollService(ctx, icpClient, resolver, grpcServer, cfg); err != nil {
		return err
	}
	return nil

}

func (svc *Service) setICPClient(client icp.ICPClient) {
	icpClient = client
}

func (svc *Service) initEnrollService(ctx context.Context, icpClient icp.ICPClient, resolver grpcutil.Resolver, grpcServer *grpc.Server, cfg *EnrollConfig) error {
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}

	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}

	if cfg.Authz.Enabled {
		authzClient, err = authz.NewAuthzClient(ctx, resolver)
		if err != nil {
			return err
		}
	}

	cloudacctClient = pb.NewCloudAccountServiceClient(conn)
	cloudacctMemberClient = pb.NewCloudAccountMemberServiceClient(conn)
	addr, err = resolver.Resolve(ctx, "billing")
	if err != nil {
		return err
	}
	conn, err = grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	billingAcctClient = pb.NewBillingAccountServiceClient(conn)
	billingOptionsClient = pb.NewBillingOptionServiceClient(conn)
	billingCreditServiceClient = pb.NewBillingCreditServiceClient(conn)
	billingCouponServiceClient = pb.NewBillingCouponServiceClient(conn)
	// set authz client

	addr, err = resolver.Resolve(ctx, "cloudcredits")
	if err != nil {
		return err
	}
	conn, err = grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	cloudCreditsCouponServiceClient = pb.NewCloudCreditsCouponServiceClient(conn)
	cloudCreditsCreditServiceClient = pb.NewCloudCreditsCreditServiceClient(conn)

	pb.RegisterCloudAccountEnrollServiceServer(grpcServer, &CloudAccountEnrollService{cfg: cfg})
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "cloudaccount-enroll"
}
