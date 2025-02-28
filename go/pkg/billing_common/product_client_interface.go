// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Product catalog works with Product catalog operator.
// Product catalog operator is for deploying CRDs using a chosen YAML description for the catalog.
// Product catalog then reads the catalog information from a cluster with such deployed CRDs.
// For billing to be able to get the product catalog, we have to either change the behavior of product catalog to have a intermediate DB,
// such that the DB is responsible for storing and returning product catalog information or
// billing has to get this data from a deployed cluster.
// We do not wish to change the behavior of product catalog.
// We cannot deploy a entire cluster for testing billing.
// Hence, we introduce a interface which can provide the needed data from product catalog.
// Such data will represent a product catalog behavior for integration and unit testing of billing.
// However, when we run tests against a full deployment, the behavior will be as is.
// We will invest in both unit testing, and integration testing without a full deployment and with a full deployment.
// The interface allows for having such test specific behavior and having similar to a full deployment behavior.
// This does not mean that there will be broken integration points because the behavior is purely for retrieving data.
// Data as epxected by billing which breaks if expectations are broken.
// If and when expectations are broken, they will be verified by tests which run against a full deployment.
// At the same time, the tests against test implementation of the interface will help assessing the integration is broken because
// of change in expectations and not change in billing.

type ProductClientInterface interface {
	GetProductCatalogProducts(ctx context.Context) ([]*pb.Product, error)
	GetProductCatalogProductsWithFilter(ctx context.Context, filter *pb.ProductFilter) ([]*pb.Product, error)
	GetProductCatalogProductsForAccountTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.Product, error)
	SetProductStatus(ctx context.Context, productStatus *pb.SetProductStatusRequest) error
	GetProductCatalogVendors(ctx context.Context) ([]*pb.Vendor, error)
}
