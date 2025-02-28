//go:build Functional || Products || Positive || gRPC
// +build Functional Products Positive gRPC

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/grpc/productcatalog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var ret_value bool

func (suite *PcGrpcTestSuite) TestGetproducts() {
	logger.Log.Info("Starting Test Get Products with No filters")
	ret_value, _ = productcatalog.Get_Products("noFilter", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcGrpcTestSuite) TestGetProductsWithNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	ret_value, _ = productcatalog.Get_Products("filterbyName", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Name Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsWithIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	ret_value, _ = productcatalog.Get_Products("filterbyId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Products By Name and Id Filter ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdName", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Name and Id Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameId() {
	logger.Log.Info("Starting Test Get Products by Name, Id,  ")
	ret_value, _ = productcatalog.Get_Products("filterbyNamId", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, ")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorId() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid ")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorId", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid, FamilyId")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorIdFamilyId", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId, familyId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadata() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid, FamilyId, Metadata")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorIdFamilyIdMeta", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId, familyId, metaData")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDesc() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid, FamilyId, Metadata")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesc", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId, familyId, metaData, description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDescEccn() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid, FamilyId, Metadat, Eccn")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesceccn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId, familyId, metaData, description, Eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDescEccnPqn() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid, FamilyId, Metadat, Eccn, Pqn")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesceccnpqn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId, familyId, metaData, description, Eccn, Pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDescEccnPqnMatch() {
	logger.Log.Info("Starting Test Get Products by Name, Id, , Vendorid, FamilyId, Metadat, Eccn, Pqn, Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesceccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, , VendorId, familyId, metaData, description, Eccn, Pqn, Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterById() {
	logger.Log.Info("Starting Test Get Products by Id  ")
	ret_value, _ = productcatalog.Get_Products("filterbyId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id ")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorId() {
	logger.Log.Info("Starting Test Get Products by Id VendorId ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdVendorId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id VendorId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Id FamilyId ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdFamilyId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id FamilyId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdDescription() {
	logger.Log.Info("Starting Test Get Products by Id Description ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdDescription", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdMetadata() {
	logger.Log.Info("Starting Test Get Products by Id Metadata ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdMetadata", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdeccn() {
	logger.Log.Info("Starting Test Get Products by Id eccn ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdeccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdpqn() {
	logger.Log.Info("Starting Test Get Products by Id pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by Id MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId ")
	ret_value, _ = productcatalog.Get_Products("filterbyIdIdVendorIdFamilyId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta")
	ret_value, _ = productcatalog.Get_Products("filterbyIdIdVendorIdFamilyIdMeta", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesc() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description")
	ret_value, _ = productcatalog.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesc", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesceccn() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn")
	ret_value, _ = productcatalog.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesceccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesceccnpqn() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn pqn")
	ret_value, _ = productcatalog.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesceccnpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesceccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesceccnpqnMatch", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn pqn Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyId() {
	logger.Log.Info("Starting Test Get Products by  FamilyId ")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with  FamilyId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByDescription() {
	logger.Log.Info("Starting Test Get Products by  Description ")
	ret_value, _ = productcatalog.Get_Products("filterbyDescription", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with  Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadata() {
	logger.Log.Info("Starting Test Get Products by  Metadata ")
	ret_value, _ = productcatalog.Get_Products("filterbyMetadata", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with  Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByeccn() {
	logger.Log.Info("Starting Test Get Products by  eccn ")
	ret_value, _ = productcatalog.Get_Products("filterbyeccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with  eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBypqn() {
	logger.Log.Info("Starting Test Get Products by  pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbypqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with  pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMatchExp() {
	logger.Log.Info("Starting Test Get Products by  MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyMatchExp", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with  MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorId() {
	logger.Log.Info("Starting Test Get Products by   VendorId ")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorId", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with   VendorId")

}
func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdDescription() {
	logger.Log.Info("Starting Test Get Products by VendorId Description ")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdDescription", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdMetadata() {
	logger.Log.Info("Starting Test Get Products by VendorId Metadata ")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdMetadata", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdeccn() {
	logger.Log.Info("Starting Test Get Products by VendorId eccn ")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdeccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdpqn() {
	logger.Log.Info("Starting Test Get Products by VendorId pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by VendorId MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdFamilyIdMeta", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesc() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdFamilyIdMetaDesc", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesceccn() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdFamilyIdMetaDesceccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesceccnpqn() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn pqn")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdFamilyIdMetaDesceccnpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesceccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyVendorIdFamilyIdMetaDesceccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn pqn Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdDescription() {
	logger.Log.Info("Starting Test Get Products by VendorId Description ")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdDescription", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMetadata() {
	logger.Log.Info("Starting Test Get Products by VendorId Metadata ")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMetadata", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdeccn() {
	logger.Log.Info("Starting Test Get Products by VendorId eccn ")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdeccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdpqn() {
	logger.Log.Info("Starting Test Get Products by VendorId pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by VendorId MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMeta", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMetaDesc() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMetaDesc", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMetaDesceccn() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMetaDesceccn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMetaDesceccnpqn() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn pqn")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMetaDesceccnpqn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdMetaDesceccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyFamilyIdMetaDesceccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn pqn Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataDescription() {
	logger.Log.Info("Starting Test Get Products byMetadata Description ")
	ret_value, _ = productcatalog.Get_Products("filterbyMetaDescription", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataeccn() {
	logger.Log.Info("Starting Test Get Products byMetadata eccn ")
	ret_value, _ = productcatalog.Get_Products("filterbyMetaeccn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadatapqn() {
	logger.Log.Info("Starting Test Get Products byMetadata pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbyMetapqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataMatchExp() {
	logger.Log.Info("Starting Test Get Products byMetadata MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyMetaMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataDesceccn() {
	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn")
	ret_value, _ = productcatalog.Get_Products("filterbyMetaDesceccn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Description eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataDesceccnpqn() {
	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn pqn")
	ret_value, _ = productcatalog.Get_Products("filterbyMetaDesceccnpqn", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Description eccn pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataDesceccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyMetaDesceccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Description eccn pqn Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByDescriptioneccn() {
	logger.Log.Info("Starting Test Get Products by Description eccn ")
	ret_value, _ = productcatalog.Get_Products("filterbyDesceccn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Description eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByDescriptionpqn() {
	logger.Log.Info("Starting Test Get Products by Description pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbyDescpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Description pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByDescriptionMatchExp() {
	logger.Log.Info("Starting Test Get Products by Description MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyDescMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Description MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByDescriptioneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by Description eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyDescDesceccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Description eccn pqn Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByeccnpqn() {
	logger.Log.Info("Starting Test Get Products by eccn pqn ")
	ret_value, _ = productcatalog.Get_Products("filterbyeccnpqn", "getproductAll")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with eccn pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByeccnMatchExp() {
	logger.Log.Info("Starting Test Get Products by eccn MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbyeccnMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with eccn MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByeccneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by eccn eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbyeccneccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with eccn eccn pqn Match Exp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBypqnMatchExp() {
	logger.Log.Info("Starting Test Get Products by pqn MatchExp ")
	ret_value, _ = productcatalog.Get_Products("filterbypqnMatchExp", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with pqn MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBypqneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by  eccn pqn Match Exp")
	ret_value, _ = productcatalog.Get_Products("filterbypqneccnpqnMatch", "getproduct1")
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with eccn pqn Match Exp")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProductsGrpcTestSuite(t *testing.T) {
	suite.Run(t, new(PcGrpcTestSuite))
}
