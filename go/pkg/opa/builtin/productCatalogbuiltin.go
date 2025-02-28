// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"google.golang.org/grpc"
)

type ProductCatalogBuiltIn struct {
	conn                *grpc.ClientConn
	productClient       pb.ProductCatalogServiceClient
	productAccessClient pb.ProductAccessServiceClient
}

func NewProductCatalogBuiltIn(conn *grpc.ClientConn) *ProductCatalogBuiltIn {
	return &ProductCatalogBuiltIn{
		conn:                conn,
		productClient:       pb.NewProductCatalogServiceClient(conn),
		productAccessClient: pb.NewProductAccessServiceClient(conn),
	}
}

func (pc *ProductCatalogBuiltIn) GetProductCatalogClient() pb.ProductCatalogServiceClient {
	return pc.productClient
}

var getProductByNameDecl = &rego.Function{
	Name: "productcatalog.getProductByName",
	Decl: types.NewFunction(
		types.Args(types.S, types.N),
		types.A),
	Memoize:          true,
	Nondeterministic: true,
}

var checkProductAccessDecl = &rego.Function{
	Name: "productcatalog.checkProductAccess",
	Decl: types.NewFunction(
		types.Args(types.S, types.S),
		types.A),
	Memoize:          true,
	Nondeterministic: true,
}

func (pc *ProductCatalogBuiltIn) getProductByNameImpl(bctx rego.BuiltinContext, prodIdArg *ast.Term, acctTypeArg *ast.Term) (*ast.Term, error) {
	ctx, logger, _ := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("ProductCatalogBuiltIn.getProductByNameImpl").Start()
	name := ""
	if err := ast.As(prodIdArg.Value, &name); err != nil {
		logger.Error(err, "failed to parse name")
		return nil, err
	}
	var typ pb.AccountType
	if err := ast.As(acctTypeArg.Value, &typ); err != nil {
		logger.Error(err, "failed to parse type")
		return nil, err
	}
	logger.WithValues("productName", name, "accountType", typ)
	logger.Info("getProductByName args", "name", name, "typ", typ)
	resp, err := pc.productClient.AdminRead(ctx, &pb.ProductFilter{
		Name:        &name,
		AccountType: &typ,
	})
	if err != nil {
		logger.Error(err, "error getting product by name")
		return nil, err
	}
	if len(resp.Products) != 1 {
		logger.Info("products length is different than 1")
		return nil, nil
	}
	prodMap := ProtoMessageToMap(resp.Products[0])
	logger.Info("product", "prodMap", prodMap)
	val, err := ast.InterfaceToValue(prodMap)
	if err != nil {
		logger.Error(err, "error converting interface to value")
		return nil, err
	}
	return ast.NewTerm(val), nil
}

func (pc *ProductCatalogBuiltIn) checkProductAccessImpl(bctx rego.BuiltinContext, productIdArg *ast.Term, cloudAccountIdArg *ast.Term) (*ast.Term, error) {
	ctx, logger, _ := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("ProductCatalogBuiltIn.checkProductAccessImpl").Start()
	defer logger.V(9).Info("end")

	productId := ""
	if err := ast.As(productIdArg.Value, &productId); err != nil {
		logger.Error(err, "failed to parse productid")
		return nil, err
	}

	cloudAccountId := ""
	if err := ast.As(cloudAccountIdArg.Value, &cloudAccountId); err != nil {
		logger.Error(err, "failed to parse cloudaccountid")
		return nil, err
	}

	logger.WithValues("productId", productId, "cloudAccountId", cloudAccountId)
	logger.Info("checking product access", "productId", productId, "cloudAccountId", cloudAccountId)

	// Check if the cloudAccountId has access to the productId
	hasAccess, err := pc.productAccessClient.CheckProductAccess(ctx, &pb.ProductAccessCheckRequest{
		CloudaccountId: cloudAccountId,
		ProductId:      productId,
	})

	if err != nil {
		logger.Error(err, "error checking product access")
		return nil, err
	}

	if !hasAccess.GetValue() {
		logger.Info("access denied", "cloudAccountId", cloudAccountId, "productId", productId)
		return nil, nil
	}

	logger.Info("access granted", "cloudAccountId", cloudAccountId, "productId", productId)

	// Return a successful response indicating access is granted
	return ast.NewTerm(ast.Boolean(true)), nil
}

func (pc *ProductCatalogBuiltIn) Register() {
	rego.RegisterBuiltin2(getProductByNameDecl, pc.getProductByNameImpl)

	rego.RegisterBuiltin2(checkProductAccessDecl, pc.checkProductAccessImpl)
}
