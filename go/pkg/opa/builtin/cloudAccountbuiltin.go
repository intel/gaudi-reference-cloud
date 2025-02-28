// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"google.golang.org/grpc"
)

type CloudAccountBuiltIn struct {
	connection               *grpc.ClientConn
	cloudAccountClient       pb.CloudAccountServiceClient
	cloudAccountMemberClient pb.CloudAccountMemberServiceClient
	cloudAccountAppClient    pb.CloudAccountAppClientServiceClient
}

func NewCloudAccountBuiltIn(connection *grpc.ClientConn) *CloudAccountBuiltIn {
	caClient := pb.NewCloudAccountServiceClient(connection)
	camClient := pb.NewCloudAccountMemberServiceClient(connection)
	caAppClient := pb.NewCloudAccountAppClientServiceClient(connection)
	return &CloudAccountBuiltIn{
		connection:               connection,
		cloudAccountClient:       caClient,
		cloudAccountMemberClient: camClient,
		cloudAccountAppClient:    caAppClient,
	}
}

func (builtIn *CloudAccountBuiltIn) GetCloudAccountClient() pb.CloudAccountServiceClient {
	return builtIn.cloudAccountClient
}

func (builtIn *CloudAccountBuiltIn) Register() {
	rego.RegisterBuiltin1(getCloudAccountByIdDecl, builtIn.getCloudAccountByIdImpl)
	rego.RegisterBuiltin1(getCloudAccountByNameDecl, builtIn.getCloudAccountByNameImpl)
	rego.RegisterBuiltin1(checkCloudAccountExistsByNameDecl, builtIn.checkCloudAccountExistsByNameImpl)
	rego.RegisterBuiltin1(getUserCloudAccountsDecl, builtIn.getUserCloudAccountsImpl)
	rego.RegisterBuiltin1(getAppClientCloudAccount, builtIn.getAppClientCloudAccountsImpl)
	rego.RegisterBuiltin2(getMemberPersonId, builtIn.getMemberPersonIdImpl)
}

var getCloudAccountByIdDecl = &rego.Function{
	Name: "cloudaccount.getById",
	Decl: types.NewFunction(
		types.Args(types.S), // Cloud account Id
		types.A),            // Returns account object
	Memoize:          true, // enable memoization across multiple calls in the same query.
	Nondeterministic: true, // indicates that the results are non-deterministic based on network conditions.
}

var checkCloudAccountExistsByNameDecl = &rego.Function{
	Name: "cloudaccount.exists",
	Decl: types.NewFunction(
		types.Args(types.S), // Cloud account Id
		types.A),            // Returns true if the cloudaccount exists
	Memoize:          true, // enable memoization across multiple calls in the same query.
	Nondeterministic: true, // indicates that the results are non-deterministic based on network conditions.
}

var getCloudAccountByNameDecl = &rego.Function{
	Name: "cloudaccount.getByName",
	Decl: types.NewFunction(
		types.Args(types.S), // Cloud Account Name
		types.A),            // Returns account object
	Memoize:          true, // enable memoization across multiple calls in the same query.
	Nondeterministic: true, // indicates that the results are non-deterministic based on network conditions.
}

var getUserCloudAccountsDecl = &rego.Function{
	Name: "cloudaccount.getRelatedCloudAccounts",
	Decl: types.NewFunction(
		types.Args(types.S), // Cloud Account Name (Member or Owner Email)
		types.A),            // Returns account object
	Memoize:          true, // enable memoization across multiple calls in the same query.
	Nondeterministic: true, // indicates that the results are non-deterministic based on network conditions.
}

var getAppClientCloudAccount = &rego.Function{
	Name: "cloudaccount.getAppClientCloudAccount",
	Decl: types.NewFunction(
		types.Args(types.S), // Cloud Account Client Id
		types.A),            // Returns account object
	Memoize:          true, // enable memoization across multiple calls in the same query.
	Nondeterministic: true, // indicates that the results are non-deterministic based on network conditions.
}

var getMemberPersonId = &rego.Function{
	Name: "cloudaccount.getMemberPersonId",
	Decl: types.NewFunction(
		types.Args(types.S, types.S), // Cloud Account Id, Member Email
		types.A),                     // Returns account object
	Memoize:          true, // enable memoization across multiple calls in the same query.
	Nondeterministic: true, // indicates that the results are non-deterministic based on network conditions.
}

func (builtIn *CloudAccountBuiltIn) checkCloudAccountExistsByNameImpl(bctx rego.BuiltinContext, b *ast.Term) (*ast.Term, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("CloudAccountBuiltIn.checkCloudAccountExistsByNameImpl").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var name string
	if err := ast.As(b.Value, &name); err != nil {
		logger.Error(err, "getting name")
		return nil, err
	}
	logger = logger.WithValues("name", name)
	exists, err := builtIn.cloudAccountClient.CheckCloudAccountExists(ctx, &pb.CloudAccountName{Name: name})
	if err != nil {
		logger.Info("failed to get cloud account by name", "error", err)
		// ignore this error and return undefined.
		return nil, nil // undefined
	}

	logger.Info("account", "exists", exists)
	v, err := ast.InterfaceToValue(exists.Value)
	logger.Info("accountV", "v", ast.NewTerm(v))
	return ast.NewTerm(v), nil
}

func (builtIn *CloudAccountBuiltIn) getCloudAccountByIdImpl(bctx rego.BuiltinContext, b *ast.Term) (*ast.Term, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("CloudAccountBuiltIn.getCloudAccountByIdImpl").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var id string
	if err := ast.As(b.Value, &id); err != nil {
		logger.Error(err, "getting id")
		return nil, err
	}
	logger = logger.WithValues("cloudAccountId", id)
	cloudAccount, err := builtIn.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: id})

	if err != nil {
		logger.Info("failed to get cloud account by id", "error", err)
		// ignore this error and return undefined.
		return nil, nil // undefined
	}

	if !cloudAccount.PaidServicesAllowed {
		logger.Info("paid services not allowed for this cloudAccountId", "cloudAccountId", id)
	}

	acctMap := ProtoMessageToMap(cloudAccount)
	logger.Info("account", "acctMap", acctMap)
	v, err := ast.InterfaceToValue(acctMap)
	if err != nil {
		return nil, nil
	}
	return ast.NewTerm(v), nil
}

func (builtIn *CloudAccountBuiltIn) getCloudAccountByNameImpl(bctx rego.BuiltinContext, b *ast.Term) (*ast.Term, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("CloudAccountBuiltIn.getCloudAccountByNameImpl").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var name string
	if err := ast.As(b.Value, &name); err != nil {
		logger.Error(err, "getting name")
		return nil, err
	}
	logger = logger.WithValues("name", name)
	cloudAccount, err := builtIn.cloudAccountClient.GetByName(ctx, &pb.CloudAccountName{Name: name})
	if err != nil {
		// The stack trace has been intentionally removed as this masks the actual error in the opa logs.
		logger.Info("Failed to get cloud account by name", "error", err)
		// ignore this error and return undefined.
		return nil, nil // undefined
	}

	acctMap := ProtoMessageToMap(cloudAccount)
	logger.Info("account", "acctMap", acctMap)
	v, err := ast.InterfaceToValue(acctMap)
	if err != nil {
		logger.Error(err, "convert to map")
		return nil, nil
	}
	return ast.NewTerm(v), nil
}

func (builtIn *CloudAccountBuiltIn) getUserCloudAccountsImpl(bctx rego.BuiltinContext, b *ast.Term) (*ast.Term, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("CloudAccountBuiltIn.getUserCloudAccountsImpl").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var name string
	if err := ast.As(b.Value, &name); err != nil {
		logger.Error(err, "getting name")
		return nil, err
	}
	logger = logger.WithValues("name", name)
	relatedAccounts, err := builtIn.cloudAccountMemberClient.GetCloudAccountsForOpa(ctx, &pb.AccountUser{UserName: name})
	if err != nil {
		// The stack trace has been intentionally removed as this masks the actual error in the opa logs.
		logger.Info("Failed to get cloud account by name", "error", err)
		// ignore this error and return undefined.
		return nil, nil // undefined
	}

	acctList := ProtoMessageToMap(relatedAccounts)
	logger.Info("accounts", "acctsList", acctList)
	v, err := ast.InterfaceToValue(acctList)
	if err != nil {
		logger.Error(err, "convert to map")
		return nil, nil
	}
	return ast.NewTerm(v), nil
}

func (builtIn *CloudAccountBuiltIn) getMemberPersonIdImpl(bctx rego.BuiltinContext, a *ast.Term, b *ast.Term) (*ast.Term, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("CloudAccountBuiltIn.getMemberPersonIdImpl").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var email string
	if err := ast.As(a.Value, &email); err != nil {
		logger.Error(err, "error getting email")
		return nil, err
	}

	var cloudaccount_id string
	if err := ast.As(b.Value, &cloudaccount_id); err != nil {
		logger.Error(err, "getting cloudaccount_id")
		return nil, err
	}
	logger = logger.WithValues("cloudAccountId", cloudaccount_id)
	accountPerson, err := builtIn.cloudAccountMemberClient.GetMemberPersonId(ctx, &pb.AccountUser{UserName: email, CloudAccountId: &cloudaccount_id})
	if err != nil {
		// The stack trace has been intentionally removed as this masks the actual error in the opa logs.
		logger.Info("Failed to get cloud account by name", "error", err)
		// ignore this error and return undefined.
		return nil, nil // undefined
	}

	acctPerson := ProtoMessageToMap(accountPerson)
	logger.Info("accounts", "acctPerson", acctPerson)
	v, err := ast.InterfaceToValue(acctPerson)
	if err != nil {
		logger.Error(err, "convert to map")
		return nil, nil
	}
	return ast.NewTerm(v), nil
}

func (builtIn *CloudAccountBuiltIn) getAppClientCloudAccountsImpl(bctx rego.BuiltinContext, b *ast.Term) (*ast.Term, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("CloudAccountBuiltIn.getAppClientCloudAccountsImpl").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	var clientId string
	if err := ast.As(b.Value, &clientId); err != nil {
		logger.Error(err, "getting clientId")
		return nil, err
	}
	logger = logger.WithValues("clientId", clientId)
	cloudAccount, err := builtIn.cloudAccountAppClient.GetAppClientCloudAccount(ctx, &pb.AccountClient{ClientId: clientId})
	if err != nil {
		// The stack trace has been intentionally removed as this masks the actual error in the opa logs.
		logger.Info("Failed to get cloud account by clientid", "error", err)
		// ignore this error and return undefined.
		return nil, nil // undefined
	}

	logger.Info("RESPOSE", "name: ", cloudAccount.Name, "countryCode: ", cloudAccount.CountryCode, "cloudAccountId (associated): ", cloudAccount.Id)

	acctMap := ProtoMessageToMap(cloudAccount)
	logger.Info("account", "acctMap", acctMap)
	v, err := ast.InterfaceToValue(acctMap)
	if err != nil {
		logger.Error(err, "convert to map")
		return nil, nil
	}
	return ast.NewTerm(v), nil
}
