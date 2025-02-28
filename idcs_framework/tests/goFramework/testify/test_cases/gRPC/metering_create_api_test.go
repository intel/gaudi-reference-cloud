//go:build Functional || CreateUsageRecord || Metering
// +build Functional CreateUsageRecord Metering

package GRPCTest

import (
	"context"
	"errors"
	_ "fmt"
	pb "goFramework/framework/library/grpc/metering/pkg/gen/pb/metering/v1"
	"goFramework/testify/test_cases/testutils"
	"io"
	"testing"

	"goFramework/framework/common/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// var met_ret bool
var testCreate GRPCTestSuite

func Get_Create_id() int64 {
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_CloudId_TransactionId_Filter": {
			in: &pb.UsageFilter{
				CloudAccountId: &req.CloudAccountId,
			},
			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}
	//

	for _, testcase := range tests {
		out, _ := MeteringClient.Search(ctx, testcase.in)
		var outs []*pb.Usage
		for {
			o, err := out.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			outs = append(outs, o)
			return o.Id
		}

	}
	return 0

}

func (suite *GRPCTestSuite) TestCreateRecords() {
	logger.Log.Info("Starting Metering Create API Test")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageCreate
		expected expectation
	}{

		"Create_Usage_Record": {
			in: &pb.UsageCreate{
				CloudAccountId: 123456,
				TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1999",
				ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
				Properties: map[string]string{
					"instance": "small",
				},
				Timestamp: timestamppb.Now()},

			expected: expectation{
				out: 1,
				err: nil,
			},
		},
	}

	// Run tests

	for _, testcase := range tests {
		logger.Log.Info("Payload for Create Usage Record:  " + testcase.in.String())
		_, err := MeteringClient.Create(ctx, testcase.in)
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed:  Create Record")
	}

}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutTransActionId() {
	logger.Log.Info("Starting Metering Create API Test without Transaction Id")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageCreate
		expected expectation
	}{
		"Create_Usage_Record": {
			in: &pb.UsageCreate{
				CloudAccountId: 123456,
				ResourceId:     "test",
				Properties: map[string]string{
					"instance": "small",
				},
				Timestamp: timestamppb.Now()},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record creation")
	for _, testcase := range tests {
		logger.Log.Info("Payload for Create Usage Record:  " + testcase.in.String())
		_, err := MeteringClient.Create(ctx, testcase.in)
		logger.Log.Info("Output from Create Usage Record API --> " + err.Error())
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed:  Create Record")
	}

}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutCloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test without CloudAccountId")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageCreate
		expected expectation
	}{
		"Create_Usage_Record": {
			in: &pb.UsageCreate{
				TransactionId: "3bc52387-da79-4947-a562-ab7a88c38e1999",
				ResourceId:    "test",
				Properties: map[string]string{
					"instance": "small",
				},
				Timestamp: timestamppb.Now()},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record creation")
	for _, testcase := range tests {
		logger.Log.Info("Payload for Create Usage Record: %s\n " + testcase.in.String())
		_, err := MeteringClient.Create(ctx, testcase.in)
		logger.Log.Info("Output from Create Usage Record API --> " + err.Error())
		logger.Log.Info(" Validate Error : Want:" + err_str.Error() + "Got: " + err.Error())
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed:  Create Record without CloudAccountId")
	}
}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutResourceId() {
	logger.Log.Info("Starting Metering Create API Test without Resource Id")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageCreate
		expected expectation
	}{

		"Create_Usage_Record": {
			in: &pb.UsageCreate{
				CloudAccountId: 123456,
				TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1999",
				Properties: map[string]string{
					"instance": "small",
				},
				Timestamp: timestamppb.Now()},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record creation")
	for _, testcase := range tests {
		_, err := MeteringClient.Create(ctx, testcase.in)
		logger.Log.Info("Output from Create Usage Record API --> " + err.Error())
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed:  Create Record without Resource Id")
	}

}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutProperties() {
	logger.Log.Info("Starting Metering Create API Test Without Properties")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageCreate
		expected expectation
	}{

		"Create_Usage_Record": {
			in: &pb.UsageCreate{
				CloudAccountId: 123456,
				TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1999",
				ResourceId:     "test",
				Timestamp:      timestamppb.Now()},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	for _, testcase := range tests {
		logger.Log.Info("Payload for Create Usage Record:  " + testcase.in.String())
		_, err := MeteringClient.Create(ctx, testcase.in)
		assert.Equal(suite.T(), err, nil, "Test Failed: Create Record Without Properties")
	}

}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutTimeStamp() {
	logger.Log.Info("Starting Metering Create API Test Without Timestamp")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageCreate
		expected expectation
	}{

		"Create_Usage_Record": {
			in: &pb.UsageCreate{
				CloudAccountId: 123456,
				TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1999",
				ResourceId:     "test",
				Properties: map[string]string{
					"instance": "small",
				},
			},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record creation")
	for _, testcase := range tests {
		logger.Log.Info("Payload for Create Usage Record:  " + testcase.in.String())
		_, err := MeteringClient.Create(ctx, testcase.in)
		logger.Log.Info("Output from Create Usage Record API --> " + err.Error())
		logger.Log.Info(" Validate Error : Want: " + err_str.Error() + "Got:" + err.Error())
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed: Create Record Without Timestamp")
	}

}

func (suite *GRPCTestSuite) TestCreateRecordsandValidateCreation() {
	searchId := Get_Create_id()
	logger.Log.Info("Starting Metering Read API Test with Id")
	ctx := context.Background()

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_Id_Filter": {

			in: &pb.UsageFilter{
				Id: &searchId,
			},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	//

	for _, testcase := range tests {
		logger.Logf.Infof("Search Metering Record with payload : %s\n", testcase.in)
		out, err := MeteringClient.Search(ctx, testcase.in)
		var outs []*pb.Usage
		for {
			o, err := out.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			outs = append(outs, o)
		}
		if err != nil {
			if testcase.expected.err.Error() != err.Error() {
				logger.Logf.Infof("Error -> Want: %s\nGot: %s\n", testcase.expected.err, err)
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Id filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with Id filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Id filter")

	}
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidcloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test Invalid Cloud Account Id")
	payload := testutils.Get_Invalid_Value("create", "cloudAccountId")
	out, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Response Body --> " + out + "and Error " + err)
	assert.Contains(suite.T(), err, "invalid syntax", "Test Failed:  Create Record with Invalid Cloud Account Id")
	assert.Contains(suite.T(), out, "", "Test Failed:  Create Record with Invalid Cloud Account Id")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidResourceId() {
	logger.Log.Info("Starting Metering Create API Test with Invalid Resource Id")
	payload := testutils.Get_Invalid_Value("create", "resourceId")
	out, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Response Body --> " + out + "and Error " + err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed:  Create Record with Invalid Resource Id")
	assert.Contains(suite.T(), out, "", "Test Failed - Create Record with Invalid Resource Id")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidTransactionId() {
	logger.Log.Info("Starting Metering Create API Test with Invalid transactionId")
	payload := testutils.Get_Invalid_Value("create", "transactionId")
	out, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Response Body --> " + out + "and Error " + err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed:  Create Record with Invalid transactionId")
	assert.Contains(suite.T(), out, "", "Test Failed:  Create Record with Invalid transactionId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidtimestamp() {
	logger.Log.Info("Starting Metering Create API Test with Invalid timestamp")
	payload := testutils.Get_Invalid_Value("create", "timeStamp")
	out, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Response Body --> " + out + "and Error " + err)
	assert.Contains(suite.T(), err, "invalid input arguments", "Test Failed:  Create Record with Invalid timestamp")
	assert.Contains(suite.T(), out, "", "Test Failed:  Create Record with Invalid timestamp")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidproperties() {
	logger.Log.Info("Starting Metering Create API Test with Invalid properties")
	payload := testutils.Get_Invalid_Value("create", "properties")
	out, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Response Body --> " + out + "and Error " + err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed:  Create Record with Invalid properties")
	assert.Contains(suite.T(), out, "", "Test Failed:  Create Record with Invalid properties")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingcloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test with missing cloudAccountId")
	payload := testutils.Get_Missing_Value("create", "cloudAccountId")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "", "Test Failed:  Create Record with missing cloudAccountId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingResourceId() {
	logger.Log.Info("Starting Metering Create API Test with missing resourceId")
	payload := testutils.Get_Missing_Value("create", "resourceId")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "", "Test Failed:  Create Record with missing resourceId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingTransactionId() {
	logger.Log.Info("Starting Metering Create API Test with missing transactionId")
	payload := testutils.Get_Missing_Value("create", "transactionId")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "", "Test Failed:  Create Record with missing transactionid")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingtimestamp() {
	logger.Log.Info("Starting Metering Create API Test with missing timestamp")
	payload := testutils.Get_Missing_Value("create", "timeStamp")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "", "Test Failed:  Create Record with missing timestamp")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingproperties() {
	logger.Log.Info("Starting Metering Create API Test with missing properties")
	payload := testutils.Get_Missing_Value("create", "properties")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "", "Test Failed:  Create Record with missing properties")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithEmptyPayload() {
	logger.Log.Info("Starting Metering Create API Test with empty payload")
	payload := testutils.Get_Missing_Value("create", "empty")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "", "Test Failed:  Create Record with empty payload")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithExtraField() {
	logger.Log.Info("Starting Metering Create API Test with extra field in the payload")
	payload := testutils.Get_Invalid_Value("create", "extraField")
	_, err := testutils.Execute_Create_Usage_Record_Gcurl(payload)
	logger.Log.Info("Output from Create Usage Record API Error --> " + err)
	assert.Contains(suite.T(), err, "message type idc_metering.UsageCreate has no known field named extraField", "Test Failed:  Create Record with extra field in the payload")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateUsageSuite(t *testing.T) {
	testCreate = GRPCTestSuite{}
	// Single client used for testing the APIs
	suite.Run(t, new(GRPCTestSuite))

}
