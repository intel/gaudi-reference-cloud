// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	"os"
	"strconv"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/otel/attribute"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type AuthzBuiltIn struct {
	connection  *grpc.ClientConn
	authzClient pb.AuthzServiceClient
}

func NewAuthzBuiltIn(connection *grpc.ClientConn) *AuthzBuiltIn {
	authzClient := pb.NewAuthzServiceClient(connection)

	return &AuthzBuiltIn{
		connection:  connection,
		authzClient: authzClient,
	}
}

func (builtIn *AuthzBuiltIn) GetAuthzClient() pb.AuthzServiceClient {
	return builtIn.authzClient
}

func (builtIn *AuthzBuiltIn) Register() {
	rego.RegisterBuiltinDyn(checkAuthz, builtIn.authCheck)
}

var checkAuthz = &rego.Function{
	Name: "authzService.check",
	Decl: types.NewFunction(
		types.Args(
			types.S, // email
			types.S, // enterpriseId
			types.S, // cloudAccountId
			types.S, // path
			types.S, // verb
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.A)), // payload as a JSON object
		),
		types.A),
	Memoize:          true,
	Nondeterministic: true,
}

func (builtIn *AuthzBuiltIn) authCheck(bctx rego.BuiltinContext, operands []*ast.Term) (*ast.Term, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("AuthzBuiltIn.authCheck").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	var email, enterpriseId, cloudAccountId, path, verb string
	var payload *structpb.Struct

	if err := ast.As(operands[0].Value, &email); err != nil {
		logger.Error(err, "failed to get email")
		return nil, err
	}

	if err := ast.As(operands[1].Value, &enterpriseId); err != nil {
		logger.Error(err, "failed to get enterpriseId")
		return nil, err
	}

	if err := ast.As(operands[2].Value, &cloudAccountId); err != nil {
		logger.Error(err, "failed to get cloudAccountId")
		return nil, err
	}

	span.SetAttributes(attribute.String("cloudAccountId", cloudAccountId))

	if err := ast.As(operands[3].Value, &path); err != nil {
		logger.Error(err, "failed to get path")
		return nil, err
	}

	if err := ast.As(operands[4].Value, &verb); err != nil {
		logger.Error(err, "failed to get verb")
		return nil, err
	}

	if err := ast.As(operands[5].Value, &payload); err != nil {
		logger.Error(err, "failed to get payload")
		return nil, err
	}

	logger.V(9).Info("AuthzBuiltIn", "cloudAccountId", cloudAccountId, "path", path, "verb", verb, "payload", payload, "enterpriseId", enterpriseId)

	authzEnabled, _ := strconv.ParseBool(os.Getenv("AUTHZ_ENABLED"))
	checkResult := false

	if authzEnabled {
		logger.V(9).Info("authz is enabled executing check call")
		check, err := builtIn.authzClient.CheckInternal(ctx, &pb.AuthorizationRequestInternal{CloudAccountId: cloudAccountId, Path: path, Verb: verb, Payload: payload, User: &pb.UserIdentification{Email: email, EnterpriseId: enterpriseId}})
		if err != nil {
			logger.Error(err, "failed to call check internal")
			return nil, nil // undefined
		}
		checkResult = check.Allowed
	} else {
		logger.V(9).Info("authz is disabled returning true")
		checkResult = true
	}

	v, err := ast.InterfaceToValue(checkResult)

	if err != nil {
		logger.Error(err, "failed to convert authorization check result")
		return nil, nil
	}

	logger.V(9).Info("result", "check", ast.NewTerm(v))

	return ast.NewTerm(v), nil
}
