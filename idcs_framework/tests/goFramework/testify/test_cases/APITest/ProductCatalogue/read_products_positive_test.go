//go:build Functional || Products || Regression || Positive
// +build Functional Products Regression Positive

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_value bool

func (suite *PcAPITestSuite) TestGetproducts() {
	logger.Log.Info("Starting Test Get Products with No filters")
	ret_value = financials.Get_Products("noFilters", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestGetProductsWithNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	ret_value = financials.Get_Products("filterbyName", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Name Filter")
}

func (suite *PcAPITestSuite) TestGetProductsWithIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	ret_value = financials.Get_Products("filterbyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Filter")
}

func (suite *PcAPITestSuite) TestGetProductsWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Products By Name and Id Filter ")
	ret_value = financials.Get_Products("filterbyIdName", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Name and Id Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByVendorId() {
	logger.Log.Info("Starting Test Get Products by VendorId Filter ")
	ret_value = financials.Get_Products("filterbyVendorId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Vendor Id Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyId() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter")
	ret_value = financials.Get_Products("filterbyfamilyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Family Id Filter")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByDescription() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
// 	ret_value = financials.Get_Products("filterbydescription", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Description Filter")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByMetadata() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
	ret_value = financials.Get_Products("filterbymetadata", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByeccn() {
	logger.Log.Info("Starting Test Get Products by eccn Filter ")
	ret_value = financials.Get_Products("filterbyeccn", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with eccn Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBypcq() {
	logger.Log.Info("Starting Test Get Products by pcq Filter ")
	ret_value = financials.Get_Products("filterbypcq", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with pcq Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByMatchExp() {
	logger.Log.Info("Starting Test Get Products by Match Expression Filter ")
	ret_value = financials.Get_Products("filterbymatchExpr", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Match Expression  Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByIdname() {
	logger.Log.Info("Starting Test Get Products by Id Name ")
	ret_value = financials.Get_Products("filterbyIdName", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Name")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBynameVendorId() {
	logger.Log.Info("Starting Test Get Products by Name VendorId")
	ret_value = financials.Get_Products("filterbyNameVendorId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name VendorId")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBynameFamilyId() {
	logger.Log.Info("Starting Test Get Products by Name VendorId")
	ret_value = financials.Get_Products("filterbyNameFamilyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name FamilyId")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterBynameDescription() {
// 	logger.Log.Info("Starting Test Get Products by Name Description")
// 	ret_value = financials.Get_Products("filterbyNameDescription", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name Description")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterBynameMetadata() {
	logger.Log.Info("Starting Test Get Products by Name Metadata")
	ret_value = financials.Get_Products("filterbyNameMetadata", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name Metadata")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBynameEccn() {
	logger.Log.Info("Starting Test Get Products by Name eccn")
	ret_value = financials.Get_Products("filterbyNameeccn", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name eccn")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterBynamepqn() {
// 	logger.Log.Info("Starting Test Get Products by Name pqn")
// 	ret_value = financials.Get_Products("filterbyNampqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name pqn")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterBynameMatchExp() {
	logger.Log.Info("Starting Test Get Products by Name Match Expression ")
	ret_value = financials.Get_Products("filterbyNamMatchExp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name Match Expression")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorId() {
	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid ")
	ret_value = financials.Get_Products("filterbyNamIdVendorId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId")
	ret_value = financials.Get_Products("filterbyNamIdVendorIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId")
}

func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadata() {
	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadata")
	ret_value = financials.Get_Products("filterbyNamIdVendorIdFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDesc() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadata")
// 	ret_value = financials.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDescEccn() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadat, Eccn")
// 	ret_value = financials.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description, Eccn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDescEccnPqn() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadat, Eccn, Pqn")
// 	ret_value = financials.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description, Eccn, Pqn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterBynameIdVendorIdFamilyIdMetadataDescEccnPqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadat, Eccn, Pqn, Match Exp")
// 	ret_value = financials.Get_Products("filterbyNamIdVendorIdFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description, Eccn, Pqn, Match Exp")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterById() {
	logger.Log.Info("Starting Test Get Products by Id  ")
	ret_value = financials.Get_Products("filterbyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id ")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Id FamilyId ")
	ret_value = financials.Get_Products("filterbyIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id FamilyId")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByIdDescription() {
// 	logger.Log.Info("Starting Test Get Products by Id Description ")
// 	ret_value = financials.Get_Products("filterbyIdDescription", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Description")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByIdMetadata() {
	logger.Log.Info("Starting Test Get Products by Id Metadata ")
	ret_value = financials.Get_Products("filterbyIdMetadata", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id Metadata")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByIdeccn() {
	logger.Log.Info("Starting Test Get Products by Id eccn ")
	ret_value = financials.Get_Products("filterbyIdeccn", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id eccn")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByIdpqn() {
// 	logger.Log.Info("Starting Test Get Products by Id pqn ")
// 	ret_value = financials.Get_Products("filterbyIdpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id pqn")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by Id MatchExp ")
	ret_value = financials.Get_Products("filterbyIdMatchExp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id MatchExp")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorId() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId ")
	ret_value = financials.Get_Products("filterbyIdIdVendorId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId")

}

func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId ")
	ret_value = financials.Get_Products("filterbyIdIdVendorIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta")
	ret_value = financials.Get_Products("filterbyIdIdVendorIdFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesc() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description")
// 	ret_value = financials.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn")
// 	ret_value = financials.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn pqn")
// 	ret_value = financials.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByIdVendorIdFamilyIdMetaDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn pqn Match Exp")
// 	ret_value = financials.Get_Products("filterbyIdIdVendorIdFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn pqn Match Exp")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by VendorId FamilyId ")
	ret_value = financials.Get_Products("filterbyVendorIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  FamilyId")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdDescription() {
// 	logger.Log.Info("Starting Test Get Products by VendorId Description ")
// 	ret_value = financials.Get_Products("filterbyVendorIdDescription", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Description")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdMetadata() {
	logger.Log.Info("Starting Test Get Products by VendorId Metadata ")
	ret_value = financials.Get_Products("filterbyVendorIdMetadata", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Metadata")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdeccn() {
	logger.Log.Info("Starting Test Get Products by VendorId eccn ")
	ret_value = financials.Get_Products("filterbyVendorIdeccn", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  eccn")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdpqn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId pqn ")
// 	ret_value = financials.Get_Products("filterbyVendorIdpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  pqn")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by VendorId MatchExp ")
	ret_value = financials.Get_Products("filterbyVendorIdMatchExp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  MatchExp")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta")
	ret_value = financials.Get_Products("filterbyVendorIdFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesc() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description")
// 	ret_value = financials.Get_Products("filterbyVendorIdFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn")
// 	ret_value = financials.Get_Products("filterbyVendorIdFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn pqn")
// 	ret_value = financials.Get_Products("filterbyVendorIdFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByVendorIdFamilyIdMetaDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn pqn Match Exp")
// 	ret_value = financials.Get_Products("filterbyVendorIdFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdDescription() {
// 	logger.Log.Info("Starting Test Get Products by VendorId Description ")
// 	ret_value = financials.Get_Products("filterbyFamilyIdDescription", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Description")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMetadata() {
	logger.Log.Info("Starting Test Get Products by VendorId Metadata ")
	ret_value = financials.Get_Products("filterbyFamilyIdMetadata", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  Metadata")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdeccn() {
	logger.Log.Info("Starting Test Get Products by VendorId eccn ")
	ret_value = financials.Get_Products("filterbyFamilyIdeccn", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  eccn")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdpqn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId pqn ")
// 	ret_value = financials.Get_Products("filterbyFamilyIdpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  pqn")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by VendorId MatchExp ")
	ret_value = financials.Get_Products("filterbyFamilyIdMatchExp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with VendorId  MatchExp")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta")
	ret_value = financials.Get_Products("filterbyFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMetaDesc() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description")
// 	ret_value = financials.Get_Products("filterbyFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMetaDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn")
// 	ret_value = financials.Get_Products("filterbyFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMetaDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn pqn")
// 	ret_value = financials.Get_Products("filterbyFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByFamilyIdMetaDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn pqn Match Exp")
// 	ret_value = financials.Get_Products("filterbyFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByMetadataDescription() {
// 	logger.Log.Info("Starting Test Get Products byMetadata Description ")
// 	ret_value = financials.Get_Products("filterbyMetaDescription", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  Description")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByMetadataeccn() {
	logger.Log.Info("Starting Test Get Products byMetadata eccn ")
	ret_value = financials.Get_Products("filterbyMetaeccn", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  eccn")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByMetadatapqn() {
// 	logger.Log.Info("Starting Test Get Products byMetadata pqn ")
// 	ret_value = financials.Get_Products("filterbyMetapqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  pqn")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByMetadataMatchExp() {
	logger.Log.Info("Starting Test Get Products byMetadata MatchExp ")
	ret_value = financials.Get_Products("filterbyMetaMatchExp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Metadata  MatchExp")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterByMetadataDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn")
// 	ret_value = financials.Get_Products("filterbyMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByMetadataDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn pqn")
// 	ret_value = financials.Get_Products("filterbyMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByMetadataDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn pqn Match Exp")
// 	ret_value = financials.Get_Products("filterbyMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Metadata Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByDescriptioneccn() {
// 	logger.Log.Info("Starting Test Get Products by Description eccn ")
// 	ret_value = financials.Get_Products("filterbyDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByDescriptionpqn() {
// 	logger.Log.Info("Starting Test Get Products by Description pqn ")
// 	ret_value = financials.Get_Products("filterbyDescpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Description pqn")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByDescriptionMatchExp() {
// 	logger.Log.Info("Starting Test Get Products by Description MatchExp ")
// 	ret_value = financials.Get_Products("filterbyDescMatchExp", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with Description MatchExp")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByDescriptioneccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by Description eccn pqn Match Exp")
// 	ret_value = financials.Get_Products("filterbyDescDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetProductsFilterByeccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by eccn pqn ")
// 	ret_value = financials.Get_Products("filterbyeccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with eccn pqn")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByeccnMatchExp() {
	logger.Log.Info("Starting Test Get Products by eccn MatchExp ")
	ret_value = financials.Get_Products("filterbyeccnMatchExp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with eccn MatchExp")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByeccneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by eccn eccn pqn Match Exp")
	ret_value = financials.Get_Products("filterbyeccneccnpqnMatch", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with eccn eccn pqn Match Exp")
}

// func (suite *PcAPITestSuite) TestGetProductsFilterBypqnMatchExp() {
// 	logger.Log.Info("Starting Test Get Products by pqn MatchExp ")
// 	ret_value = financials.Get_Products("filterbypqnMatchExp", 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors  with pqn MatchExp")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterBypqneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by  eccn pqn Match Exp")
	ret_value = financials.Get_Products("filterbypqneccnpqnMatch", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to Get Vendors with eccn pqn Match Exp")
}

func (suite *PcAPITestSuite) TestGetProductsCreationTime() {
	logger.Log.Info("Starting Test Get Products and check creation time is not null")
	ret_value = financials.Check_Creation_Time("noFilters", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed to check creation time in products")
}
