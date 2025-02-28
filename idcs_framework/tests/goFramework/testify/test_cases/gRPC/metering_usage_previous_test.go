//go:build Functional || FindPreviousUsageRecord || Metering
// +build Functional FindPreviousUsageRecord Metering

package GRPCTest

import (
	"context"
	"errors"
	_ "fmt"
	"goFramework/framework/common/logger"
	pb "goFramework/framework/library/grpc/metering/pkg/gen/pb/metering/v1"
	"goFramework/testify/test_cases/testutils"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testFindPrevious GRPCTestSuite
var resourceId string
var Id int64

func _create_records_1(ctx context.Context, resource_id string) {

	for i := 0; i < 5; i++ {
		u := pb.UsageCreate{
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1" + strconv.Itoa(i),
			CloudAccountId: 123456,
			ResourceId:     resource_id,
			Properties: map[string]string{
				"instance": "small",
			},
			Timestamp: timestamppb.Now(),
		}
		logger.Logf.Infof("Creating record", &u)
		_, err := MeteringClient.Create(ctx, &u)
		if err != nil {
			logger.Logf.Infof("Create Record error: ", err)
		}
	}
}

func Get_id() int64 {
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
		logger.Logf.Infof("Output of search is", out)
		var outs []*pb.Usage
		for {
			o, err := out.Recv()
			logger.Logf.Infof("Read Output", o)
			if errors.Is(err, io.EOF) {
				break
			}
			outs = append(outs, o)
			return o.Id
		}

	}
	return 0

}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecords() {
	logger.Logf.Infof("Starting Find Previous API Test")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)

	//  Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}
	_create_records_1(ctx, req.ResourceId)
	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{
		"Update_Usage_Record": {
			in: &pb.UsagePrevious{
				ResourceId: req.ResourceId},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}
	// // Run tests
	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		resourceId = out.GetResourceId()
		Id = out.GetId()
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), out.ResourceId, req.ResourceId, "Test Failed to Read previous usage on resource")
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Read previous usage on resource")
	}
}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecordsWithOutId() {
	logger.Logf.Infof("Starting Find Previous API Test without Id")

	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsagePrevious{
				ResourceId: resourceId},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	//err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record updation")
	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Read previous usage on resource without Id")
	}

}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecordsWithOutResourceId() {
	logger.Logf.Infof("Starting Find Previous API Test Without ResourceId")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsagePrevious{
				Id: Id},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring find request")
	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed to Create Record without ResourceId")
	}

}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Find Previous API Test With Empty Payload")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsagePrevious{},
			expected: expectation{
				out: 0,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring find request")
	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed to Create Record with empty payload")
	}

}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecordsWithWrongId() {
	logger.Logf.Infof("Starting Find Previous API Test with wrong Id")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsagePrevious{
				Id: *testutils.GetInt64Pointer(123456),
			},

			expected: expectation{
				out: 0,
				err: nil,
			},
		},
	}

	// Run tests
	err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring find request")
	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), err.Error(), err_str.Error(), "Test Failed to Create Record with wrong id")
	}

}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecordsWithResourceId() {
	logger.Logf.Infof("Starting Find Previous API Test with ResourceId")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsagePrevious{
				ResourceId: *testutils.GetStringPointer("InvalidResourceId"),
			},

			expected: expectation{
				out: 0,
				err: nil,
			},
		},
	}

	// Run tests
	//err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring find request")
	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Create Record with ResourceId")
	}

}

func (suite *GRPCTestSuite) TestFindPreviousUsageRecordsWithWrongPayload() {
	logger.Logf.Infof("Starting Find Previous API Test with Wrong Payload")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsagePrevious
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsagePrevious{
				ResourceId: *testutils.GetStringPointer("InvalidResourceId1"),
				Id:         *testutils.GetInt64Pointer(654321),
			},

			expected: expectation{
				out: 0,
				err: nil,
			},
		},
	}

	// Run tests

	for _, testcase := range tests {
		out, err := MeteringClient.FindPrevious(ctx, testcase.in)
		logger.Logf.Infof("Output from Find Previous Record API  --> %s\n", out)
		logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Create Record with wrong payload")
	}

}

func (suite *GRPCTestSuite) TestPreviousUsageRecordsWithInvalidId() {
	logger.Logf.Infof("Starting Metering FInd Previous Usage  API Test with Invalid Id")
	payload := testutils.Get_Invalid_Value("findPrevious", "id")
	out, err := testutils.Execute_FindPrevious_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "invalid syntax", "Test Failed to FInd Previous Usage  Record with Invalid Id")
	assert.Contains(suite.T(), out, "", "Test Failed to FInd Previous Usage  Record with Invalid Id")
}

func (suite *GRPCTestSuite) TestPreviousUsageRecordsWithInvalidresourceId() {
	logger.Logf.Infof("Starting Metering FInd Previous Usage  API Testwith invalid resourceId")
	payload := testutils.Get_Invalid_Value("findPrevious", "resourceId")
	out, err := testutils.Execute_FindPrevious_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed to FInd Previous Usage  Record with invalid resourceId")
	assert.Contains(suite.T(), out, "", "Test Failed to FInd Previous Usage  Record with invalid resourceId")
}

func (suite *GRPCTestSuite) TestPreviousUsageRecordsWithMissingId() {
	logger.Logf.Infof("Starting Metering FInd Previous Usage  API Test with missing Id")
	payload := testutils.Get_Missing_Value("findPrevious", "id")
	_, err := testutils.Execute_FindPrevious_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to FInd Previous Usage  Record with missing Id")
}

func (suite *GRPCTestSuite) TestPreviousUsageRecordsWithMissingresourceId() {
	logger.Logf.Infof("Starting Metering FInd Previous Usage  API Test with missing resourceId")
	payload := testutils.Get_Missing_Value("findPrevious", "resourceId")
	_, err := testutils.Execute_FindPrevious_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to FInd Previous Usage  Record with missing resourceId")
}

func (suite *GRPCTestSuite) TestPreviousUsageRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Metering FInd Previous Usage  API Test with Empty Payload")
	payload := testutils.Get_Missing_Value("findPrevious", "empty")
	_, err := testutils.Execute_FindPrevious_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to FInd Previous Usage  Record with Empty Payload")
}

func (suite *GRPCTestSuite) TestPreviousUsageRecordsWithExtraField() {
	logger.Logf.Infof("Starting Metering FInd Previous Usage  API Test with Extra Field")
	payload := testutils.Get_Invalid_Value("findPrevious", "extraField")
	_, err := testutils.Execute_FindPrevious_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Find Previous  Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "message type idc_metering.UsagePrevious has no known field named extraField", "Test Failed to FInd Previous Usage  Record with Extra Field")
}

// // In order for 'go test' to run this suite, we need to create
// //
// //	a normal test function and pass our suite to suite.Run
func TestFindPreviousSuite(t *testing.T) {
	testFindPrevious = GRPCTestSuite{}
	//Single client used for testing the APIs
	suite.Run(t, new(GRPCTestSuite))

}
