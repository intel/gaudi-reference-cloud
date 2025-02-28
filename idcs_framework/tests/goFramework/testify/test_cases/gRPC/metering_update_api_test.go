//go:build Functional || UpdateUsageRecord || Metering
// +build Functional UpdateUsageRecord Metering

package GRPCTest

import (
	"context"
	"errors"
	_ "fmt"
	"goFramework/framework/common/logger"
	_ "goFramework/framework/library"
	pb "goFramework/framework/library/grpc/metering/pkg/gen/pb/metering/v1"
	"goFramework/testify/test_cases/testutils"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testUpdate GRPCTestSuite

func get_update_rec_id() int64 {
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

	for scenario, testcase := range tests {
		logger.Logf.Infof("Scenario", scenario)
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

func (suite *GRPCTestSuite) TestUpdateRecords() {
	logger.Logf.Infof("Starting Metering Update API Test")
	ctx := context.Background()
	id := get_update_rec_id()
	// // Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageUpdate
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsageUpdate{
				Id:       []int64{id},
				Reported: true},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// // Run tests

	for _, testcase := range tests {
		out, err := MeteringClient.Update(ctx, testcase.in)
		logger.Logf.Infof("Output from Update Usage Record API --> %s\n", out)
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Update Record")
		assert.Equal(suite.T(), err, nil, "Test Failed to Update Record")

		//Verify the change by reading the data

		readtests := map[string]struct {
			in       *pb.UsageFilter
			expected expectation
		}{

			"ReadApi_CloudId_TransactionId_Filter": {
				in: &pb.UsageFilter{
					Reported: testutils.GetBoolPointer(true),
				},
				expected: expectation{
					out: RECORDS_COUNT_ONE,
					err: nil,
				},
			},
		}
		//

		for scenario, testcase := range readtests {
			logger.Logf.Infof("Scenario", scenario)
			out, _ := MeteringClient.Search(ctx, testcase.in)
			logger.Logf.Infof("Output of search is", out)
			var outs []*pb.Usage
			for {
				o, err := out.Recv()
				logger.Logf.Infof("Search Output record ", o)
				if errors.Is(err, io.EOF) {
					break
				}
				outs = append(outs, o)
			}
			logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
			logger.Logf.Infof("Output from DB", outs)
			assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Update Record with Reported")

		}

	}
}

func (suite *GRPCTestSuite) TestUpdateRecords1() {
	logger.Logf.Infof("Starting Metering Update API Test")
	ctx := context.Background()
	// Define TestCase inputs and expectations
	_, err := testutils.UpdateMeteringRecord(MeteringClient, ctx)
	assert.Equal(suite.T(), err, nil, "Test Failed to Update Record")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithOutId() {
	logger.Logf.Infof("Starting Metering Update API Test without Id")

	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageUpdate
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsageUpdate{
				Reported: true},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	for scenario, testcase := range tests {
		logger.Logf.Infof("Scenario", scenario)
		_, err := MeteringClient.Update(ctx, testcase.in)
		logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
		assert.Equal(suite.T(), err, nil, "Test Failed to Update Record")
	}

}

func (suite *GRPCTestSuite) TestUpdateRecordsWithOutReported() {
	logger.Logf.Infof("Starting Metering Update API Test without Reported")
	id := get_update_rec_id()
	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageUpdate
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsageUpdate{
				Id: []int64{id},
			},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	//err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record updation")
	for _, testcase := range tests {
		logger.Logf.Infof("Search Metering Record with payload : %s\n", testcase.in)
		_, err := MeteringClient.Update(ctx, testcase.in)
		logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
		assert.Equal(suite.T(), err, nil, "Test Failed to Update Record without Reported Field")
	}

}

func (suite *GRPCTestSuite) TestUpdateRecordsWithWrongId() {
	logger.Logf.Infof("Starting Metering Update API Test with Wrong Id")

	ctx := context.Background()
	// Define TestCase inputs and expectations
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageUpdate
		expected expectation
	}{

		"Update_Usage_Record": {
			in: &pb.UsageUpdate{
				Id:       []int64{99999},
				Reported: *testutils.GetBoolPointer(true),
			},

			expected: expectation{
				out: RECORDS_COUNT_ONE,
				err: nil,
			},
		},
	}

	// Run tests
	//err_str := errors.New("rpc error: code = Unknown desc = invalid input arguments, ignoring record updation")
	for _, testcase := range tests {
		logger.Logf.Infof("Search Metering Record with payload : %s\n", testcase.in)
		_, err := MeteringClient.Update(ctx, testcase.in)
		logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
		assert.Equal(suite.T(), err, nil, "Test Failed to Update Record With Wrong Id")
	}

}

func (suite *GRPCTestSuite) TestUpdateRecordsWithInvalidId() {
	logger.Logf.Infof("Starting Metering Update API Test")
	payload := testutils.Get_Invalid_Value("update", "id")
	out, err := testutils.Execute_Update_Usage_Record_Gcurl(payload)
	assert.Contains(suite.T(), err, "invalid syntax", "Test Failed to Update Record")
	assert.Contains(suite.T(), out, "", "Test Failed to Update Record")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithInvalidreported() {
	logger.Logf.Infof("Starting Metering Update API Test With Invalid Reported")
	payload := testutils.Get_Invalid_Value("update", "reported")
	out, err := testutils.Execute_Update_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed to Update Record with Invalid Reported")
	assert.Contains(suite.T(), out, "", "Test Failed to Update Record with Invalid Reported")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithMissingId() {
	logger.Logf.Infof("Starting Metering Update API Test with Missing Id Field")
	payload := testutils.Get_Missing_Value("update", "id")
	_, err := testutils.Execute_Update_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Update Record with Missing Id Field")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithMissingreported() {
	logger.Logf.Infof("Starting Metering Update API Test with missing Reported")
	payload := testutils.Get_Missing_Value("update", "reported")
	_, err := testutils.Execute_Update_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Update Record with missing Reported")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Metering Update API Test with empty payload")
	payload := testutils.Get_Missing_Value("update", "empty")
	_, err := testutils.Execute_Update_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Update Record with empty payload")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithExtraField() {
	logger.Logf.Infof("Starting Metering Update API Test with extra field")
	payload := testutils.Get_Invalid_Value("update", "extraField")
	_, err := testutils.Execute_Update_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "message type idc_metering.UsageUpdate has no known field named extraField", "Test Failed to Update Record with Extra Field")
}

// In order for 'go test' to run this suite, we need to create
//
//	a normal test function and pass our suite to suite.Run
func TestUpdateUsageSuite(t *testing.T) {
	testUpdate = GRPCTestSuite{}
	//Single client used for testing the APIs
	suite.Run(t, new(GRPCTestSuite))

}
