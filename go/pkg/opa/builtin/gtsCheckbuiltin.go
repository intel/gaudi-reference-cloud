// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	tradecheck "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tradecheck/tradecheckintel"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

type GTSCheckBuiltIn struct {
	client *tradecheck.GTSclient
}

func NewGTSCheckBuiltIn(client *tradecheck.GTSclient) *GTSCheckBuiltIn {
	return &GTSCheckBuiltIn{
		client: client,
	}
}

func (gc *GTSCheckBuiltIn) GetGTSclient() *tradecheck.GTSclient {
	return gc.client
}

var isGTSOrderValidDecl = &rego.Function{
	Name: "gts.isGTSOrderValid",
	Decl: types.NewFunction(
		types.Args(types.S, types.S, types.S, types.S),
		types.A),
	Memoize:          true,
	Nondeterministic: true,
}

func (gc *GTSCheckBuiltIn) isGTSOrderValidImpl(bctx rego.BuiltinContext, productId *ast.Term, email *ast.Term,
	personId *ast.Term, countryCode *ast.Term) (*ast.Term, error) {
	ctx, logger, _ := obs.LogAndSpanFromContextOrGlobal(bctx.Context).WithName("GTSCheckBuiltIn.isGTSOrderValidImpl").Start()
	var prodId string
	if err := ast.As(productId.Value, &prodId); err != nil {
		logger.Error(err, "failed to parse productId")
		return nil, err
	}
	var emailId string
	if err := ast.As(email.Value, &emailId); err != nil {
		logger.Error(err, "failed to parse emailId")
		return nil, err
	}
	var personID string
	if err := ast.As(personId.Value, &personID); err != nil {
		logger.Error(err, "failed to parse personId")
		return nil, err
	}
	var cCode string
	if err := ast.As(countryCode.Value, &cCode); err != nil {
		logger.Error(err, "failed to parse countryCode")
		return nil, err
	}

	logger.Info("isGTSOrderValidImpl", "productId", prodId, "personId", personID, "countryCode", cCode)
	isValid, err := gc.client.IsOrderValid(ctx, prodId, emailId, personID, cCode)
	if err != nil {
		logger.Info("Error: gts client returned an error, log it and pass the check", "productId", prodId, "personId", personID, "countryCode", cCode)
		// Return true since a connectivity issue with GTS will block all requests.
		isValid = true
	}

	val, err := ast.InterfaceToValue(isValid)
	if err != nil {
		return nil, err
	}
	return ast.NewTerm(val), nil
}

func (gc *GTSCheckBuiltIn) Register() {
	rego.RegisterBuiltin4(isGTSOrderValidDecl, gc.isGTSOrderValidImpl)
}
