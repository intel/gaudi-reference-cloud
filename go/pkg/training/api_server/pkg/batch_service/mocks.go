// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"errors"
	"io"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// mock of go/pkg/grpcutil Resolver for testing
type MockResolver struct {
	MockResolve MockResolve
}

func (m *MockResolver) Resolve(ctx context.Context, s string) (string, error) {
	return m.MockResolve(ctx, s)
}

type MockResolve func(context.Context, string) (string, error)

func NewMockResolver(r MockResolve) *MockResolver {
	return &MockResolver{
		MockResolve: r,
	}
}

// mock of go/pkg/pb ProductCatalogServiceClient for testing
type MockProductCatalogServiceClient struct {
	MockAdminRead        MockProductCatalogServiceClientAdminRead
	MockUserRead         MockProductCatalogServiceClientUserRead
	MockUserReadExternal MockProductCatalogServiceClientUserReadExternal
	MockSetStatus        MockProductCatalogServiceClientSetStatus
	MockPing             MockProductCatalogServiceClientPing
}

func (m *MockProductCatalogServiceClient) AdminRead(ctx context.Context, in *pb.ProductFilter, opts ...grpc.CallOption) (*pb.ProductResponse, error) {
	return m.MockAdminRead(ctx, in, opts...)
}

func (m *MockProductCatalogServiceClient) UserRead(ctx context.Context, in *pb.ProductUserFilter, opts ...grpc.CallOption) (*pb.ProductResponse, error) {
	return m.MockUserRead(ctx, in, opts...)
}

func (m *MockProductCatalogServiceClient) UserReadExternal(ctx context.Context, in *pb.ProductUserFilter, opts ...grpc.CallOption) (*pb.ProductResponse, error) {
	return m.MockUserReadExternal(ctx, in, opts...)
}

func (m *MockProductCatalogServiceClient) SetStatus(ctx context.Context, in *pb.SetProductStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.MockSetStatus(ctx, in, opts...)
}

func (m *MockProductCatalogServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.MockPing(ctx, in, opts...)
}

type MockProductCatalogServiceClientAdminRead func(context.Context, *pb.ProductFilter, ...grpc.CallOption) (*pb.ProductResponse, error)
type MockProductCatalogServiceClientUserRead func(context.Context, *pb.ProductUserFilter, ...grpc.CallOption) (*pb.ProductResponse, error)
type MockProductCatalogServiceClientUserReadExternal func(context.Context, *pb.ProductUserFilter, ...grpc.CallOption) (*pb.ProductResponse, error)
type MockProductCatalogServiceClientSetStatus func(context.Context, *pb.SetProductStatusRequest, ...grpc.CallOption) (*emptypb.Empty, error)
type MockProductCatalogServiceClientPing func(context.Context, *emptypb.Empty, ...grpc.CallOption) (*emptypb.Empty, error)

func NewMockProductCatalogServiceClient(
	ar MockProductCatalogServiceClientAdminRead,
	ur MockProductCatalogServiceClientUserRead,
	ure MockProductCatalogServiceClientUserReadExternal,
	s MockProductCatalogServiceClientSetStatus,
	p MockProductCatalogServiceClientPing,
) *MockProductCatalogServiceClient {
	return &MockProductCatalogServiceClient{
		MockAdminRead:        ar,
		MockUserRead:         ur,
		MockUserReadExternal: ure,
		MockSetStatus:        s,
		MockPing:             p,
	}
}

// mock of go/pkg/pb ProductVendorServiceClient for testing
type MockProductVendorServiceClient struct {
	MockRead MockProductVendorServiceClientRead
}

func (m *MockProductVendorServiceClient) Read(ctx context.Context, in *pb.VendorFilter, opts ...grpc.CallOption) (*pb.VendorResponse, error) {
	return m.MockRead(ctx, in, opts...)
}

type MockProductVendorServiceClientRead func(context.Context, *pb.VendorFilter, ...grpc.CallOption) (*pb.VendorResponse, error)

func NewMockProductVendorServiceClient(
	r MockProductVendorServiceClientRead,
) *MockProductVendorServiceClient {
	return &MockProductVendorServiceClient{
		MockRead: r,
	}
}

// mock of product_client.go ProductClient for testing
// implements product_client_interface.go ProductClientInterface
type MockProductClient struct {
	MockGetProductCatalogProducts                MockGetProductCatalogProducts
	MockGetProductCatalogProductsWithFilter      MockGetProductCatalogProductsWithFilter
	MockGetProductCatalogProductsForAccountTypes MockGetProductCatalogProductsForAccountTypes
	MockSetProductStatus                         MockSetProductStatus
	MockGetProductCatalogVendors                 MockGetProductCatalogVendors
}

func (m *MockProductClient) GetProductCatalogProducts(ctx context.Context) ([]*pb.Product, error) {
	return m.MockGetProductCatalogProducts(ctx)
}

func (m *MockProductClient) GetProductCatalogProductsWithFilter(ctx context.Context, filter *pb.ProductFilter) ([]*pb.Product, error) {
	return m.MockGetProductCatalogProductsWithFilter(ctx, filter)
}

func (m *MockProductClient) GetProductCatalogProductsForAccountTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.Product, error) {
	return m.MockGetProductCatalogProductsForAccountTypes(ctx, accountTypes)
}

func (m *MockProductClient) SetProductStatus(ctx context.Context, productStatus *pb.SetProductStatusRequest) error {
	return m.MockSetProductStatus(ctx, productStatus)
}

func (m *MockProductClient) GetProductCatalogVendors(ctx context.Context) ([]*pb.Vendor, error) {
	return m.MockGetProductCatalogVendors(ctx)
}

type MockGetProductCatalogProducts func(context.Context) ([]*pb.Product, error)
type MockGetProductCatalogProductsWithFilter func(context.Context, *pb.ProductFilter) ([]*pb.Product, error)
type MockGetProductCatalogProductsForAccountTypes func(context.Context, []pb.AccountType) ([]*pb.Product, error)
type MockSetProductStatus func(context.Context, *pb.SetProductStatusRequest) error
type MockGetProductCatalogVendors func(context.Context) ([]*pb.Vendor, error)

func NewMockProductClient(
	g MockGetProductCatalogProducts,
	f MockGetProductCatalogProductsWithFilter,
	a MockGetProductCatalogProductsForAccountTypes,
	s MockSetProductStatus,
	v MockGetProductCatalogVendors,
) *MockProductClient {
	return &MockProductClient{
		MockGetProductCatalogProducts:                g,
		MockGetProductCatalogProductsWithFilter:      f,
		MockGetProductCatalogProductsForAccountTypes: a,
		MockSetProductStatus:                         s,
		MockGetProductCatalogVendors:                 v,
	}
}

// mock of go/pkg/pb CloudAccountServiceClient for testing
type MockCloudAccountServiceClient struct {
	MockCreate        MockCloudAccountServiceClientCreate
	MockEnsure        MockCloudAccountServiceClientEnsure
	MockGetById       MockCloudAccountServiceClientGetById
	MockGetByOid      MockCloudAccountServiceClientGetByOid
	MockGetByName     MockCloudAccountServiceClientGetByName
	MockGetByPersonId MockCloudAccountServiceClientGetByPersonId
	MockSearch        MockCloudAccountServiceClientSearch
	MockUpdate        MockCloudAccountServiceClientUpdate
	MockExists        MockCloudAccountServiceClientCheckCloudAccountExists
	MockDelete        MockCloudAccountServiceClientDelete
	MockPing          MockCloudAccountServiceClientPing
}

func (m *MockCloudAccountServiceClient) Create(ctx context.Context, in *pb.CloudAccountCreate, opts ...grpc.CallOption) (*pb.CloudAccountId, error) {
	return m.MockCreate(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) Ensure(ctx context.Context, in *pb.CloudAccountCreate, opts ...grpc.CallOption) (*pb.CloudAccount, error) {
	return m.MockEnsure(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) GetById(ctx context.Context, in *pb.CloudAccountId, opts ...grpc.CallOption) (*pb.CloudAccount, error) {
	return m.MockGetById(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) GetByOid(ctx context.Context, in *pb.CloudAccountOid, opts ...grpc.CallOption) (*pb.CloudAccount, error) {
	return m.MockGetByOid(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) GetByName(ctx context.Context, in *pb.CloudAccountName, opts ...grpc.CallOption) (*pb.CloudAccount, error) {
	return m.MockGetByName(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) GetByPersonId(ctx context.Context, in *pb.CloudAccountPersonId, opts ...grpc.CallOption) (*pb.CloudAccount, error) {
	return m.MockGetByPersonId(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) Search(ctx context.Context, in *pb.CloudAccountFilter, opts ...grpc.CallOption) (pb.CloudAccountService_SearchClient, error) {
	return m.MockSearch(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) Update(ctx context.Context, in *pb.CloudAccountUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.MockUpdate(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) CheckCloudAccountExists(ctx context.Context, in *pb.CloudAccountName, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error) {
	return m.MockExists(ctx, in, opts...)
}
func (m *MockCloudAccountServiceClient) Delete(ctx context.Context, in *pb.CloudAccountId, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.MockDelete(ctx, in, opts...)
}

func (m *MockCloudAccountServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.MockPing(ctx, in, opts...)
}

type MockCloudAccountServiceClientCreate func(context.Context, *pb.CloudAccountCreate, ...grpc.CallOption) (*pb.CloudAccountId, error)
type MockCloudAccountServiceClientEnsure func(context.Context, *pb.CloudAccountCreate, ...grpc.CallOption) (*pb.CloudAccount, error)
type MockCloudAccountServiceClientGetById func(context.Context, *pb.CloudAccountId, ...grpc.CallOption) (*pb.CloudAccount, error)
type MockCloudAccountServiceClientGetByOid func(context.Context, *pb.CloudAccountOid, ...grpc.CallOption) (*pb.CloudAccount, error)
type MockCloudAccountServiceClientGetByName func(context.Context, *pb.CloudAccountName, ...grpc.CallOption) (*pb.CloudAccount, error)
type MockCloudAccountServiceClientGetByPersonId func(context.Context, *pb.CloudAccountPersonId, ...grpc.CallOption) (*pb.CloudAccount, error)
type MockCloudAccountServiceClientSearch func(context.Context, *pb.CloudAccountFilter, ...grpc.CallOption) (pb.CloudAccountService_SearchClient, error)
type MockCloudAccountServiceClientUpdate func(context.Context, *pb.CloudAccountUpdate, ...grpc.CallOption) (*emptypb.Empty, error)
type MockCloudAccountServiceClientCheckCloudAccountExists func(context.Context, *pb.CloudAccountName, ...grpc.CallOption) (*wrapperspb.BoolValue, error)
type MockCloudAccountServiceClientDelete func(context.Context, *pb.CloudAccountId, ...grpc.CallOption) (*emptypb.Empty, error)
type MockCloudAccountServiceClientPing func(context.Context, *emptypb.Empty, ...grpc.CallOption) (*emptypb.Empty, error)

func NewMockCloudAccountServiceClient(
	c MockCloudAccountServiceClientCreate,
	e MockCloudAccountServiceClientEnsure,
	g MockCloudAccountServiceClientGetById,
	o MockCloudAccountServiceClientGetByOid,
	n MockCloudAccountServiceClientGetByName,
	p MockCloudAccountServiceClientGetByPersonId,
	s MockCloudAccountServiceClientSearch,
	u MockCloudAccountServiceClientUpdate,
	t MockCloudAccountServiceClientCheckCloudAccountExists,
	d MockCloudAccountServiceClientDelete,
	i MockCloudAccountServiceClientPing,
) *MockCloudAccountServiceClient {
	return &MockCloudAccountServiceClient{
		MockCreate:        c,
		MockEnsure:        e,
		MockGetById:       g,
		MockGetByOid:      o,
		MockGetByName:     n,
		MockGetByPersonId: p,
		MockSearch:        s,
		MockUpdate:        u,
		MockExists:        t,
		MockDelete:        d,
		MockPing:          i,
	}
}

// mock of go/pkg/pb CloudAccountService_SearchClient for testing
type MockCloudAccountServiceSearchClient struct {
	cloudAccounts []*pb.CloudAccount
	generateError bool
}

func (m *MockCloudAccountServiceSearchClient) Recv() (*pb.CloudAccount, error) {
	if m.generateError {
		return nil, errors.New("mock error")
	}
	if len(m.cloudAccounts) == 0 {
		return nil, io.EOF
	}
	nextCloudAccount := m.cloudAccounts[0]
	m.cloudAccounts = m.cloudAccounts[1:]
	return nextCloudAccount, nil
}

// functions to implement the grpc.ClientStream interface, these aren't run ever
func (m *MockCloudAccountServiceSearchClient) Header() (metadata.MD, error) {
	return nil, nil
}
func (m *MockCloudAccountServiceSearchClient) Trailer() metadata.MD {
	return nil
}
func (m *MockCloudAccountServiceSearchClient) CloseSend() error {
	return nil
}
func (m *MockCloudAccountServiceSearchClient) Context() context.Context {
	return context.Background()
}
func (m *MockCloudAccountServiceSearchClient) SendMsg(m_ interface{}) error {
	return nil
}
func (m *MockCloudAccountServiceSearchClient) RecvMsg(m_ interface{}) error {
	return nil
}

func NewMockCloudAccountServiceSearchClient(generateError bool, cloudAccounts []*pb.CloudAccount) *MockCloudAccountServiceSearchClient {
	return &MockCloudAccountServiceSearchClient{
		cloudAccounts: cloudAccounts,
		generateError: generateError,
	}
}

// mock of cloud_account_client.go CloudAccountSvcClient for testing
// implements cloud_account_client_interface.go CloudAccountClientInterface
type MockCloudAccountSvcClient struct {
	MockGetCloudAccount     MockGetCloudAccount
	MockGetCloudAccountType MockGetCloudAccountType
	MockGetAllCloudAccount  MockGetAllCloudAccount
}

func (m *MockCloudAccountSvcClient) GetCloudAccount(ctx context.Context, accountId *pb.CloudAccountId) (*pb.CloudAccount, error) {
	return m.MockGetCloudAccount(ctx, accountId)
}

func (m *MockCloudAccountSvcClient) GetCloudAccountType(ctx context.Context, accountId *pb.CloudAccountId) (pb.AccountType, error) {
	return m.MockGetCloudAccountType(ctx, accountId)
}

func (m *MockCloudAccountSvcClient) GetAllCloudAccount(ctx context.Context) ([]*pb.CloudAccount, error) {
	return m.MockGetAllCloudAccount(ctx)
}

type MockGetCloudAccount func(context.Context, *pb.CloudAccountId) (*pb.CloudAccount, error)
type MockGetCloudAccountType func(context.Context, *pb.CloudAccountId) (pb.AccountType, error)
type MockGetAllCloudAccount func(context.Context) ([]*pb.CloudAccount, error)

func NewMockCloudAccountSvcClient(c MockGetCloudAccount, g MockGetCloudAccountType, a MockGetAllCloudAccount) *MockCloudAccountSvcClient {
	return &MockCloudAccountSvcClient{
		MockGetCloudAccount:     c,
		MockGetCloudAccountType: g,
		MockGetAllCloudAccount:  a,
	}
}

// mock of go/pkg/pb BillingCreditServiceClient for testing
type MockBillingCreditServiceClient struct {
	pb.BillingCreditServiceClient
	MockCreditRead MockBillingCreditServiceClientRead
}

func (m *MockBillingCreditServiceClient) Read(ctx context.Context, in *pb.BillingCreditFilter, opts ...grpc.CallOption) (*pb.BillingCreditResponse, error) {
	return m.MockCreditRead(ctx, in, opts...)
}

type MockBillingCreditServiceClientRead func(context.Context, *pb.BillingCreditFilter, ...grpc.CallOption) (*pb.BillingCreditResponse, error)

func NewMockBillingCreditServiceClient(r MockBillingCreditServiceClientRead) *MockBillingCreditServiceClient {
	return &MockBillingCreditServiceClient{
		MockCreditRead: r,
	}
}

// mock of go/pkg/pb BillingCouponServiceClient for testing
type MockBillingCouponServiceClient struct {
	pb.BillingCouponServiceClient
	MockCouponRead MockBillingCouponServiceClientRead
}

func (m *MockBillingCouponServiceClient) Read(ctx context.Context, in *pb.BillingCouponFilter, opts ...grpc.CallOption) (*pb.BillingCouponResponse, error) {
	return m.MockCouponRead(ctx, in, opts...)
}

type MockBillingCouponServiceClientRead func(context.Context, *pb.BillingCouponFilter, ...grpc.CallOption) (*pb.BillingCouponResponse, error)

func NewMockBillingCouponServiceClient(r MockBillingCouponServiceClientRead) *MockBillingCouponServiceClient {
	return &MockBillingCouponServiceClient{
		MockCouponRead: r,
	}
}

// mock of billing_client.go BillingClient for testing
// implements billing_client_interface.go BillingClientInterface
type MockBillingClient struct {
	MockGetCloudAccountCredits MockGetCloudAccountCredits
	MockGetCouponExpiry        MockGetCouponExpiry
}

func (m *MockBillingClient) GetCloudAccountCredits(ctx context.Context, accountId string, history bool) ([]*pb.BillingCredit, error) {
	return m.MockGetCloudAccountCredits(ctx, accountId, history)
}

func (m *MockBillingClient) GetCouponExpiry(ctx context.Context, couponCode string) ([]*pb.BillingCoupon, error) {
	return m.MockGetCouponExpiry(ctx, couponCode)
}

type MockGetCloudAccountCredits func(context.Context, string, bool) ([]*pb.BillingCredit, error)
type MockGetCouponExpiry func(context.Context, string) ([]*pb.BillingCoupon, error)

func NewMockBillingClient(c MockGetCloudAccountCredits, e MockGetCouponExpiry) *MockBillingClient {
	return &MockBillingClient{
		MockGetCloudAccountCredits: c,
		MockGetCouponExpiry:        e,
	}
}
