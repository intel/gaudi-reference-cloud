//go:build Functional || JWTToken || VMaaS
// +build Functional JWTToken VMaaS

package APITest

import (
	"goFramework/framework/common/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"goFramework/framework/common/http_client"
	"goFramework/utils"
)

func (suite *APITestSuite) TestGetInstancesWithJWT() {
	logger.Log.Info("Test Get Instance with JWT")
	url := utils.Get_Base_Url() + "instances"
	StatusCode := http_client.Get_Response(url)
	assert.Equal(suite.T(), StatusCode == 200, true, "Test Failed to Validate JWT Token")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMediumVMAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
