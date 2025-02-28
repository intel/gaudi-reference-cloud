//go:build Functional || Vendors || Smoke || gRPC
// +build Functional Vendors Smoke gRPC

package ProductCatalogue

// import (
// 	"goFramework/framework/common/logger"
// 	"goFramework/framework/library/grpc/productcatalog"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/suite"
// )

// var val bool

// func (suite *PcGrpcTestSuite) TestGetVendorsWithName() {
// 	logger.Log.Info("Starting Test Get Vendors ")
// 	val, _ = productcatalog.Get_Vendors("filterbyName", "getvendorGrpc")
// 	assert.Equal(suite.T(), val, true, "Test Failed to Get Vendors with No filters")
// }

// func (suite *PcGrpcTestSuite) TestGetVendorsWithId() {
// 	logger.Log.Info("Starting Test Get Vendors ")
// 	val, _ = productcatalog.Get_Vendors("filterbyId", "getvendorGrpc")
// 	assert.Equal(suite.T(), val, true, "Test Failed to Get Vendors with No filters")
// }

// // In order for 'go test' to run this suite, we need to create
// // a normal test function and pass our suite to suite.Run
// func TestVendorsGrpcSmokeTestSuite(t *testing.T) {
// 	suite.Run(t, new(PcGrpcTestSuite))
// }
