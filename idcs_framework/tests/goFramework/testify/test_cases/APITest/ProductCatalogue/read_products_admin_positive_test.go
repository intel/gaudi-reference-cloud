//go:build Functional || Products || Regression || Positive || Admin
// +build Functional Products Regression Positive Admin

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_value_admin bool

func (suite *PcAPITestSuite) TestGetproductsAdmin() {
	logger.Log.Info("Starting Test Get Products with No filters")
	ret_value_admin = financials.Get_Products_Admin("noFilters", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestGetproductsAdminWithNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	ret_value_admin = financials.Get_Products_Admin("filterbyName", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Name Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminWithIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	ret_value_admin = financials.Get_Products_Admin("filterbyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Products By Name and Id Filter ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdName", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Name and Id Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorId() {
	logger.Log.Info("Starting Test Get Products by VendorId Filter ")
	ret_value_admin = financials.Get_Products_Admin("filterbyVendorId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Vendor Id Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyId() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter")
	ret_value_admin = financials.Get_Products_Admin("filterbyfamilyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Family Id Filter")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByDescription() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbydescription", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Description Filter")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadata() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
	ret_value_admin = financials.Get_Products_Admin("filterbymetadata", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Metadata Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByeccn() {
	logger.Log.Info("Starting Test Get Products by eccn Filter ")
	ret_value_admin = financials.Get_Products_Admin("filterbyeccn", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with eccn Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBypcq() {
	logger.Log.Info("Starting Test Get Products by pcq Filter ")
	ret_value_admin = financials.Get_Products_Admin("filterbypcq", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with pcq Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMatchExp() {
	logger.Log.Info("Starting Test Get Products by Match Expression Filter ")
	ret_value_admin = financials.Get_Products_Admin("filterbymatchExpr", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Match Expression  Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdname() {
	logger.Log.Info("Starting Test Get Products by Id Name ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdName", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id Name")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameVendorId() {
	logger.Log.Info("Starting Test Get Products by Name VendorId")
	ret_value_admin = financials.Get_Products_Admin("filterbyNameVendorId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name VendorId")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameFamilyId() {
	logger.Log.Info("Starting Test Get Products by Name VendorId")
	ret_value_admin = financials.Get_Products_Admin("filterbyNameFamilyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name FamilyId")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameDescription() {
// 	logger.Log.Info("Starting Test Get Products by Name Description")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyNameDescription", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name Description")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameMetadata() {
	logger.Log.Info("Starting Test Get Products by Name Metadata")
	ret_value_admin = financials.Get_Products_Admin("filterbyNameMetadata", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name Metadata")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameEccn() {
	logger.Log.Info("Starting Test Get Products by Name eccn")
	ret_value_admin = financials.Get_Products_Admin("filterbyNameeccn", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name eccn")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynamepqn() {
// 	logger.Log.Info("Starting Test Get Products by Name pqn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyNampqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name pqn")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameMatchExp() {
	logger.Log.Info("Starting Test Get Products by Name Match Expression ")
	ret_value_admin = financials.Get_Products_Admin("filterbyNamMatchExp", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name Match Expression")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorId() {
	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid ")
	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId")
	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorIdFamilyIdMetadata() {
	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadata")
	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorIdFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorIdFamilyIdMetadataDesc() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadata")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorIdFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorIdFamilyIdMetadataDescEccn() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadat, Eccn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorIdFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description, Eccn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorIdFamilyIdMetadataDescEccnPqn() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadat, Eccn, Pqn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorIdFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description, Eccn, Pqn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBynameIdVendorIdFamilyIdMetadataDescEccnPqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by Name, Id, Vendorid, FamilyId, Metadat, Eccn, Pqn, Match Exp")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyNamIdVendorIdFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors by Name, Id, VendorId, familyId, metaData, description, Eccn, Pqn, Match Exp")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterById() {
	logger.Log.Info("Starting Test Get Products by Id  ")
	ret_value_admin = financials.Get_Products_Admin("filterbyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id ")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Id FamilyId ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id FamilyId")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdDescription() {
// 	logger.Log.Info("Starting Test Get Products by Id Description ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyIdDescription", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id Description")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdMetadata() {
	logger.Log.Info("Starting Test Get Products by Id Metadata ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdMetadata", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id Metadata")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdeccn() {
	logger.Log.Info("Starting Test Get Products by Id eccn ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdeccn", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id eccn")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdpqn() {
// 	logger.Log.Info("Starting Test Get Products by Id pqn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyIdpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id pqn")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by Id MatchExp ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdMatchExp", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id MatchExp")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorId() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId")

}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId ")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId FamilyId")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorIdFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta")
	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorIdFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorIdFamilyIdMetaDesc() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorIdFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorIdFamilyIdMetaDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorIdFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorIdFamilyIdMetaDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn pqn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorIdFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByIdVendorIdFamilyIdMetaDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by Id  VendorId FamilyId MetaDta Description eccn pqn Match Exp")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyIdIdVendorIdFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Id  VendorId FamilyId Metadata Description eccn pqn Match Exp")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdFamilyId() {
	logger.Log.Info("Starting Test Get Products by VendorId FamilyId ")
	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdFamilyId", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  FamilyId")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdDescription() {
// 	logger.Log.Info("Starting Test Get Products by VendorId Description ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdDescription", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  Description")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdMetadata() {
	logger.Log.Info("Starting Test Get Products by VendorId Metadata ")
	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdMetadata", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  Metadata")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdeccn() {
	logger.Log.Info("Starting Test Get Products by VendorId eccn ")
	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdeccn", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  eccn")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdpqn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId pqn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  pqn")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by VendorId MatchExp ")
	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdMatchExp", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  MatchExp")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta")
	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdFamilyIdMetaDesc() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdFamilyIdMetaDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdFamilyIdMetaDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn pqn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByVendorIdFamilyIdMetaDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by VendorId  VendorId FamilyId MetaDta Description eccn pqn Match Exp")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyVendorIdFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with VendorId FamilyId Metadata Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdDescription() {
// 	logger.Log.Info("Starting Test Get Products by VendorId Description ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdDescription", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  Description")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMetadata() {
	logger.Log.Info("Starting Test Get Products by VendorId Metadata ")
	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMetadata", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  Metadata")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdeccn() {
	logger.Log.Info("Starting Test Get Products by VendorId eccn ")
	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdeccn", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  eccn")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdpqn() {
// 	logger.Log.Info("Starting Test Get Products by VendorId pqn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  pqn")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMatchExp() {
	logger.Log.Info("Starting Test Get Products by VendorId MatchExp ")
	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMatchExp", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with VendorId  MatchExp")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMeta() {
	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta")
	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMeta", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with FamilyId Metadata")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMetaDesc() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMetaDesc", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with FamilyId Metadata Description")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMetaDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMetaDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn pqn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByFamilyIdMetaDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by FamilyId MetaDta Description eccn pqn Match Exp")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyFamilyIdMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with FamilyId Metadata Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadataDescription() {
// 	logger.Log.Info("Starting Test Get Products byMetadata Description ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyMetaDescription", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Metadata  Description")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadataeccn() {
	logger.Log.Info("Starting Test Get Products byMetadata eccn ")
	ret_value_admin = financials.Get_Products_Admin("filterbyMetaeccn", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Metadata  eccn")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadatapqn() {
// 	logger.Log.Info("Starting Test Get Products byMetadata pqn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyMetapqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Metadata  pqn")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadataMatchExp() {
	logger.Log.Info("Starting Test Get Products byMetadata MatchExp ")
	ret_value_admin = financials.Get_Products_Admin("filterbyMetaMatchExp", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Metadata  MatchExp")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadataDesceccn() {
// 	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyMetaDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Metadata Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadataDesceccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn pqn")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyMetaDesceccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Metadata Description eccn pqn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByMetadataDesceccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by MetaDta Description eccn pqn Match Exp")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyMetaDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Metadata Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByDescriptioneccn() {
// 	logger.Log.Info("Starting Test Get Products by Description eccn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyDesceccn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Description eccn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByDescriptionpqn() {
// 	logger.Log.Info("Starting Test Get Products by Description pqn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyDescpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Description pqn")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByDescriptionMatchExp() {
// 	logger.Log.Info("Starting Test Get Products by Description MatchExp ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyDescMatchExp", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with Description MatchExp")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByDescriptioneccnpqnMatch() {
// 	logger.Log.Info("Starting Test Get Products by Description eccn pqn Match Exp")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyDescDesceccnpqnMatch", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with Description eccn pqn Match Exp")
// }

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByeccnpqn() {
// 	logger.Log.Info("Starting Test Get Products by eccn pqn ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbyeccnpqn", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with eccn pqn")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByeccnMatchExp() {
	logger.Log.Info("Starting Test Get Products by eccn MatchExp ")
	ret_value_admin = financials.Get_Products_Admin("filterbyeccnMatchExp", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with eccn MatchExp")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByeccneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by eccn eccn pqn Match Exp")
	ret_value_admin = financials.Get_Products_Admin("filterbyeccneccnpqnMatch", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with eccn eccn pqn Match Exp")
}

// func (suite *PcAPITestSuite) TestGetproductsAdminFilterBypqnMatchExp() {
// 	logger.Log.Info("Starting Test Get Products by pqn MatchExp ")
// 	ret_value_admin = financials.Get_Products_Admin("filterbypqnMatchExp", 200)
// 	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors  with pqn MatchExp")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterBypqneccnpqnMatch() {
	logger.Log.Info("Starting Test Get Products by  eccn pqn Match Exp")
	ret_value_admin = financials.Get_Products_Admin("filterbypqneccnpqnMatch", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to Get Vendors with eccn pqn Match Exp")
}

func (suite *PcAPITestSuite) TestGetproductsAdminCreationTime() {
	logger.Log.Info("Starting Test Get Products and check creation time is not null")
	ret_value_admin = financials.Check_Creation_Time("noFilters", 200)
	assert.Equal(suite.T(), ret_value_admin, true, "Test Failed to check creation time in products")
}
