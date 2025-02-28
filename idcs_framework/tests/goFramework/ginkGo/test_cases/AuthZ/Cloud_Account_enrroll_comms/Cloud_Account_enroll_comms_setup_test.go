package cloud_account_enrroll_comms_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount_enroll"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"k8s.io/utils/strings/slices"
)

type EnrollConfig struct {
	ListenPort uint16                     `koanf:"listenPort"`
	ICPConfig  cloudaccount_enroll.Config `koanf:"icp"`
}

func (cc *EnrollConfig) GetListenPort() uint16 {
	return cc.ListenPort
}

func (cc *EnrollConfig) SetListenPort(port uint16) {
	cc.ListenPort = port
}

type CloudAccountEnrollService struct {
	pb.UnimplementedCloudAccountEnrollServiceServer
}

type enrollRequestData struct {
	Email        string
	Tid          string
	EnterpriseId string
	Groups       []string
	Idp          string
}

// This function attempts to detect an enterprise user based on IDP claim in the token.
// Note: This function does not work for an intel account.
func (data *enrollRequestData) isEnterprise(ctx context.Context) bool {
	logger := log.FromContext(ctx)
	//This is the list of all b2c domain names used by IDP.
	b2cIDP := []string{"intelcorpintb2c.onmicrosoft.com", "intelcorpb2c.onmicrosoft.com", "intelcorpdevb2c.onmicrosoft.com"}
	if slices.Contains(b2cIDP, data.Idp) {
		logger.Info("Non-Enterprise user details", "oid", data.EnterpriseId, "idp", data.Idp)
		return false
	} else {
		//The Idp claim will be changed only after the enterpise user is approved in eRPM.
		logger.Info("Enterprise user details", "oid", data.EnterpriseId, "idp", data.Idp)
		return true
	}
}

func isIntelEmail(email string) bool {
	intelSuffs := []string{"@intel.com", "@corpint.intel.com"}
	for _, suff := range intelSuffs {
		if strings.HasSuffix(email, suff) {
			return true
		}
	}
	return false
}

func validate(ctx context.Context, reqData *enrollRequestData, enrollRequest *pb.CloudAccountEnrollRequest) error {
	logger := log.FromContext(ctx)
	// Intel user does not have group
	// Validate if the Groups contains DevCloud Console Standard other than intel users.
	if !isIntelEmail(reqData.Email) && !slices.Contains(reqData.Groups, "DevCloud Console Standard") {
		logger.V(9).Info("DevCloud Console Standard group is missing")
		return status.Error(codes.InvalidArgument, "DevCloud Console Standard group is missing from the JWT token")
	}

	if isIntelEmail(reqData.Email) && enrollRequest.Premium {
		logger.V(9).Info("Intel email id should not have Premium flag")
		return status.Error(codes.InvalidArgument, "Intel email id should not have Premium flag")
	}
	return nil
}

func (ces *CloudAccountEnrollService) Enroll(ctx context.Context, req *pb.CloudAccountEnrollRequest) (*pb.CloudAccountEnrollResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.Enroll").Start()
	defer span.End()
	var reqData *enrollRequestData

	//attempt to extract from JWT token.
	reqData, err := extractRequestfromCtx(ctx)
	if err != nil {
		logger.Error(err, "failed to get required data from JWT token")
		return nil, status.Errorf(codes.InvalidArgument, "Failed to get required data from the JWT token")
	}
	logger.Info("Enroll API invoked", "oid", reqData.EnterpriseId)

	if err := validate(ctx, reqData, req); err != nil {
		return nil, err
	}

	resp := pb.CloudAccountEnrollResponse{
		Registered: true,
	}
	acctType := pb.AccountType_ACCOUNT_TYPE_STANDARD // default account type
	personId := ""
	currentAccDetails, err := cloudacctClient.GetByOid(ctx, &pb.CloudAccountOid{Oid: reqData.EnterpriseId})
	if err == nil {
		// Cloud account already exists no need to fetch person id.
		// The account type needs to be changed only in the case of
		// 		a. enterprise approved or enterprise pending.
		//		b. Explicit request to upgrade to Premium.
		acctType = currentAccDetails.Type
		personId = currentAccDetails.PersonId
		if currentAccDetails.Type != pb.AccountType_ACCOUNT_TYPE_INTEL && reqData.isEnterprise(ctx) {
			// this handles the upgrade scenario.
			acctType = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
		} else if currentAccDetails.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE && !reqData.isEnterprise(ctx) {
			//demotion usecase, change account to standard.
			acctType = pb.AccountType_ACCOUNT_TYPE_STANDARD
		} else if currentAccDetails.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD {
			if req.Premium {
				logger.Info("Request to upgrade to Premium", "oid", reqData.EnterpriseId, "persondId", currentAccDetails.PersonId)
				acctType = pb.AccountType_ACCOUNT_TYPE_PREMIUM
			} else {
				// check if the account is enterprise pending.
				isEnterprisePending, _, err := enterprisePendingDetails(ctx, reqData)
				if err == nil && isEnterprisePending {
					acctType = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING
				}
			}
		}
	} else if status.Code(err) == codes.NotFound {
		// Cloud account does not exist
		// Fetch the personId from ICP/eRPM
		logger.Info("Cloud account does not exist, attempting to fetch the personId field from eRPM", "oid", reqData.EnterpriseId)
		isEnterprisePending, entId, err := enterprisePendingDetails(ctx, reqData)
		if err != nil {
			return nil, err
		}
		personId = entId

		// Compute the acctType
		if isIntelEmail(reqData.Email) {
			acctType = pb.AccountType_ACCOUNT_TYPE_INTEL
		} else if reqData.isEnterprise(ctx) {
			acctType = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
		} else if req.Premium {
			acctType = pb.AccountType_ACCOUNT_TYPE_PREMIUM
		} else if isEnterprisePending {
			acctType = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING
		}
	} else {
		logger.Info("Failed to fetch CloudAccount details ", "oid", reqData.EnterpriseId)
		return nil, err
	}

	// Create/Update the Cloud account.
	logger.Info("Attempting to create/update cloud account ", "oid", reqData.EnterpriseId, "type", acctType, "personId", personId)
	currentAccDetails, err = cloudacctClient.Ensure(ctx, &pb.CloudAccountCreate{
		Tid:      reqData.Tid,
		Oid:      reqData.EnterpriseId,
		Name:     reqData.Email,
		Owner:    reqData.Email,
		Type:     acctType,
		PersonId: personId,
	})
	if err != nil {
		logger.Error(err, "cloudaccount ensure failed", "oid", reqData.EnterpriseId)
		resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
		return &resp, nil
	}
	resp.HaveCloudAccount = true
	resp.CloudAccountId = currentAccDetails.Id
	resp.CloudAccountType = currentAccDetails.Type

	enrollSteps(ctx, reqData, currentAccDetails, &resp)
	logger.Info("Enroll completed for ", "oid", reqData.EnterpriseId, "cloudAccountId", resp.CloudAccountId,
		"CloudAccountType", resp.CloudAccountType, "enrollAction", resp.Action)
	return &resp, nil
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

	tid, ok := claims["tid"].(string)
	if !ok || tid == "" {
		return nil, errors.New("TID is not part of the JWT token")
	}
	id, ok := claims["enterpriseId"].(string)
	if !ok || id == "" {
		return nil, errors.New("enterpriseId is not part of the JWT token")
	}
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email is not part of the JWT token")
	}
	idp, ok := claims["idp"].(string)
	if !ok || idp == "" {
		return nil, errors.New("idp is not part of the JWT token")
	}
	aInterface, ok := claims["groups"].([]interface{})
	if !ok {
		return nil, errors.New("groups not part of the JWT token")
	}
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
	}, nil
}

// This function returns true only if user is enterprise pending.
func enterprisePendingDetails(ctx context.Context, obj *enrollRequestData) (bool, string, error) {
	logger := log.FromContext(ctx)
	// This function calls eRPM APIs to check for pending enterprise enrollment
	isPending, personid, err := icpClient.IsEnterprisePending(ctx, obj.Email, obj.EnterpriseId)
	if err != nil {
		//log the error.
		logger.Error(err, "Failed to get information if user is enterprise pending and personId", "oid", obj.EnterpriseId)
		return false, "", err
	}
	return isPending, personid, nil
}

func enrollSteps(ctx context.Context, obj *enrollRequestData,
	acct *pb.CloudAccount, resp *pb.CloudAccountEnrollResponse) {
	logger := log.FromContext(ctx)

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
	}
	resp.HaveBillingAccount = true

	if !acct.Enrolled && acct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		creds, err := billingCreditServiceClient.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: acct.GetId()})
		if err != nil {
			logger.Error(err, "read billing credits failed", "oid", obj.EnterpriseId,
				"cloudAccountId", acct.GetId())
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
			return
		}
		_, err = creds.Recv()
		if err == nil {
			resp.Enrolled = true
		} else if !errors.Is(err, io.EOF) {
			logger.Error(err, "billing credit service Client returned an error", "oid", obj.EnterpriseId, "cloudAccountId", acct.GetId())
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
			return
		}

		if !resp.Enrolled {
			opts, err := billingOptionsClient.Read(ctx, &pb.BillingOptionFilter{CloudAccountId: &acct.Id})
			if err != nil {
				logger.Error(err, "read billing options failed", "oid", obj.EnterpriseId,
					"cloudAccountId", acct.GetId())
				resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_RETRY
				return
			}
			fmt.Print(opts)
		}

		if !resp.Enrolled {
			resp.Action = pb.CloudAccountEnrollAction_ENROLL_ACTION_COUPON_OR_CREDIT_CARD
			return
		}
	}
	resp.Enrolled = true
}

type Service struct {
	icpConfig *cloudaccount_enroll.ICPConfig
}

var (
	billingAcctClient          pb.BillingAccountServiceClient
	billingOptionsClient       pb.BillingOptionServiceClient
	billingCreditServiceClient pb.BillingCreditServiceClient
	cloudacctClient            pb.CloudAccountServiceClient
	icpClient                  cloudaccount_enroll.ICPClient
)

func (svc *Service) Init(ctx context.Context, cfg *EnrollConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	// Load the ICP configuration.
	if svc.icpConfig == nil {
		var err error
		svc.icpConfig, err = cloudaccount_enroll.New(ctx, &cfg.ICPConfig)
		if err != nil {
			return err
		}
	}

	// Instantiate the icpClient
	icpClient = cloudaccount_enroll.CreateICPClient(svc.icpConfig)
	if err := svc.initEnrollService(ctx, icpClient, resolver, grpcServer); err != nil {
		return err
	}
	return nil

}

func (svc *Service) setICPClient(client cloudaccount_enroll.ICPClient) {
	icpClient = client
}

func (svc *Service) initEnrollService(ctx context.Context, icpClient cloudaccount_enroll.ICPClient, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}

	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	cloudacctClient = pb.NewCloudAccountServiceClient(conn)

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

	pb.RegisterCloudAccountEnrollServiceServer(grpcServer, &CloudAccountEnrollService{})
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "cloudaccount-enroll"
}
