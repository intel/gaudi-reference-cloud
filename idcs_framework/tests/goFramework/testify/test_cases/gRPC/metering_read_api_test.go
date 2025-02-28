//go:build Functional || SearchUsageRecord || Metering
// +build Functional SearchUsageRecord Metering

package GRPCTest

import (
	"context"
	"errors"
	_ "fmt"
	"goFramework/framework/common/logger"
	_ "goFramework/framework/library"
	_ "goFramework/framework/library/grpc/metering/pkg/db/model"

	pb "goFramework/framework/library/grpc/metering/pkg/gen/pb/metering/v1"
	"goFramework/testify/test_cases/testutils"
	"io"
	_ "log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var met_ret bool
var test GRPCTestSuite
var searchId int64
var starttime timestamppb.Timestamp
var endtimestamp timestamppb.Timestamp

func Get_Search_id() int64 {
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

func _create_records(ctx context.Context) {

	for i := 0; i < RECORDS_COUNT_ALL; i++ {
		u := pb.UsageCreate{
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1" + strconv.Itoa(i),
			CloudAccountId: 123456,
			ResourceId:     "test",
			Properties: map[string]string{
				"instance": "small",
			},
			Timestamp: timestamppb.Now(),
		}
		logger.Logf.Infof("Creating Usage Record : %d\n", &u)
		_, err := MeteringClient.Create(ctx, &u)
		if err != nil {
			logger.Logf.Infof("Create Record error: %s\n", err)
		}

	}
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithNoFilter() {
	logger.Log.Info("Starting Metering Read API Test Without Filter")
	ctx := context.Background()
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_NoFilter": {
			in: &pb.UsageFilter{},
			expected: expectation{
				out: RECORDS_COUNT_ALL,
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
				logger.Log.Info(" Validate Error : Want: " + testcase.expected.err.Error() + " Got: " + err.Error())
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with No filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with No filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with No filter")
	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithTransactionIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with TransactionId filter")
	ctx := context.Background()
	_create_records(ctx)
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
				TransactionId:  &req.TransactionId,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with TransactionId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with TransactionId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with TransactionId filter")
	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with cloudId filter")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_CloudId_Filter": {
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with cloudId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with cloudId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with cloudId filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithResourceIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with ResourceId Filter")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_ResourceId_Filter": {
			in: &pb.UsageFilter{
				ResourceId: &req.ResourceId,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with resourceId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with resourceId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with resourceId filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdResourceId() {
	logger.Log.Info("Starting Metering Read API Test with cloudid and resourceId")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_cloudId_resourceId_Filter": {

			in: &pb.UsageFilter{
				CloudAccountId: &req.CloudAccountId,
				ResourceId:     &req.ResourceId,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with cloudId and resourceId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with cloudId and resourceId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with cloudId and resourceId filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdTransactionId() {
	logger.Log.Info("Starting Metering Read API Test")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)
	_create_records(ctx)

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
				TransactionId:  &req.TransactionId,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with cloudId and transactionId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with cloudId and transactionId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with cloudId and transactionId filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithReportedResourceId() {
	logger.Log.Info("Starting Metering Read API Test with Reported and ResourceId")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_Reported_ResourceId_Filter": {

			in: &pb.UsageFilter{
				Reported:   testutils.GetBoolPointer(false),
				ResourceId: &req.ResourceId,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with reported and ResourceId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with reported and ResourceId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with reported and ResourceId filter")

	}

}

// func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithTimestamp() {
// 	logger.Log.Info("Starting Metering Read API Test")
// 	ctx := context.Background()
// 	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
// 	time.Sleep(20)
// 	_create_records(ctx)

// 	type expectation struct {
// 		out int
// 		err error
// 	}
// 	tests := map[string]struct {
// 		in       *pb.UsageFilter
// 		expected expectation
// 	}{

// 		"ReadApi_TransactionId_Filter": {

// 			in: &pb.UsageFilter{
// 				StartTime: req.Timestamp,
// 			},

// 			expected: expectation{
// 				out: RECORDS_COUNT_ONE,
// 				err: nil,
// 			},
// 		},
// 	}

// 	//

// 	for scenario, testcase := range tests {
// 		logger.Log.Info("Scenario", scenario)
// 		out, err := MeteringClient.Search(ctx, testcase.in)
// 		var outs []*pb.Usage
// 		for {
// 			o, err := out.Recv()
// 			logger.Log.Info("Read Output", o)
// 			if errors.Is(err, io.EOF) {
// 				break
// 			}
// 			outs = append(outs, o)
// 		}
// 		if err != nil {
// 			if testcase.expected.err.Error() != err.Error() {
// 				logger.Log.Info("Err -> Want: Got:\n", testcase.expected.err, err)
// 			}
// 		} else {
// 			if len(outs) != testcase.expected.out {
// 				logger.Log.Info("Out -> Want: Got : \n", testcase.expected.out, len(outs))
// 			}
// 		}
// 		assert.Equal(suite.T(), err, nil, "Test Failed to Search Record")
// 	}

// }

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithReported() {
	logger.Log.Info("Starting Metering Read API Test with Reported")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)
	_create_records(ctx)
	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_Reported_Filter": {
			in: &pb.UsageFilter{
				Reported:       testutils.GetBoolPointer(false),
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with reported  filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with reported filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with reported filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithAllFilters() {
	logger.Log.Info("Starting Metering Read API Test with all Filters")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_All_Filter": {
			in: &pb.UsageFilter{
				Reported:       testutils.GetBoolPointer(false),
				CloudAccountId: &req.CloudAccountId,
				TransactionId:  &req.TransactionId,
				ResourceId:     &req.ResourceId,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with all filters")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with all filters")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with all filters")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithId() {
	searchId = Get_Search_id()
	logger.Log.Info("Starting Metering Read API Test with Id")
	ctx := context.Background()
	_create_records(ctx)

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

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithOneCorrectFilter() {
	logger.Log.Info("Starting Metering Read API Test with Non Existing Values in Filters")
	ctx := context.Background()
	req, _ := testutils.CreateMeteringRecord(MeteringClient, ctx)
	time.Sleep(5 * time.Second)
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_NonExisting_Values_Filter": {

			in: &pb.UsageFilter{
				CloudAccountId: &req.CloudAccountId,
				TransactionId:  testutils.GetStringPointer("qqqqqq"),
				ResourceId:     testutils.GetStringPointer("zzzz"),
			},

			expected: expectation{
				out: 0,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in  filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with Non existing values in  filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in  filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingCloudAccountId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing Cloud Account Id")
	ctx := context.Background()
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{

		"ReadApi_nonexiting_cloudAccountId_Filter": {
			in: &pb.UsageFilter{
				Reported:       testutils.GetBoolPointer(false),
				CloudAccountId: testutils.GetUint64Pointer(888888),
			},

			expected: expectation{
				out: 0,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in cloudAccountID filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with Non existing values in cloudAccountID filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in cloudAccountID filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingResourceId() {
	logger.Log.Info("Starting Metering Read API Test Non existing ResourceId")
	ctx := context.Background()
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{
		"ReadApi_Non_Existing_ResourceId_Filter": {
			in: &pb.UsageFilter{
				ResourceId: testutils.GetStringPointer("abcd"),
			},

			expected: expectation{
				out: 0,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in cloudAccountID filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with Non existing values in cloudAccountID filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in cloudAccountID filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingTransactionId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing TransactionId")
	ctx := context.Background()
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{
		"ReadApi_Non_Existing_TransactionId_Filter": {
			in: &pb.UsageFilter{
				TransactionId: testutils.GetStringPointer("abc123d"),
			},

			expected: expectation{
				out: 0,
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
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in transactionId filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with Non existing values in transactionId filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in transactionId filter")

	}

}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing Id")
	ctx := context.Background()
	_create_records(ctx)

	type expectation struct {
		out int
		err error
	}

	tests := map[string]struct {
		in       *pb.UsageFilter
		expected expectation
	}{
		"ReadApi_Non_Existing_Id_Filter": {
			in: &pb.UsageFilter{
				Id: testutils.GetInt64Pointer(13500),
			},

			expected: expectation{
				out: 0,
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
			logger.Logf.Infof("Search Output record ", o)
			if errors.Is(err, io.EOF) {
				break
			}
			outs = append(outs, o)
		}
		if err != nil {
			if testcase.expected.err.Error() != err.Error() {
				logger.Logf.Infof("Error -> Want: %s\nGot: %s\n", testcase.expected.err, err)
				assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in Id filter")
			}
		} else {
			if len(outs) != testcase.expected.out {
				logger.Logf.Infof("Output -> Want: %d Got : %d\n", testcase.expected.out, len(outs))
				logger.Logf.Infof("Output from DB", outs)
				assert.Equal(suite.T(), testcase.expected.out, len(outs), "Test Failed to Search Record with Non existing values in Id filter")
			}
		}
		assert.Equal(suite.T(), testcase.expected.err, err, "Test Failed to Search Record with Non existing values in Id filter")

	}

}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidcloudAccountId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid CloudAccountId")
	payload := testutils.Get_Invalid_Value("search", "cloudAccountId")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "invalid syntax", "Test Failed to Search Record with Invalid CloudAccountId")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with Invalid CloudAccountId")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Id")
	payload := testutils.Get_Invalid_Value("search", "id")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "invalid syntax", "Test Failed to Search Record with Invalid Id")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with Invalid Id")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidResourceId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Resource Id")
	payload := testutils.Get_Invalid_Value("search", "resourceId")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed to Search Record with Invalid Resource Id")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with Invalid Resource Id")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidTransactionId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid TransactionID")
	payload := testutils.Get_Invalid_Value("search", "transactionId")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed to Search Record with Invalid TransactionID")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with Invalid TransactionID")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidstarttime() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Start Time")
	payload := testutils.Get_Invalid_Value("search", "startTime")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "invalid input arguments", "Test Failed to Search Record with Invalid Start Time")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with Invalid Start Time")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidendtime() {
	logger.Log.Info("Starting Metering Search API Test with invalid End Time")
	payload := testutils.Get_Invalid_Value("search", "endTime")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> ", err)
	assert.Contains(suite.T(), err, "invalid input arguments", "Test Failed to Search Record with invalid End Time")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with invalid End Time")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidreported() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Reported")
	payload := testutils.Get_Invalid_Value("search", "reported")
	out, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "bad input: expecting", "Test Failed to Search Record with Invalid Reported")
	assert.Contains(suite.T(), out, "", "Test Failed to Search Record with Invalid Reported")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingcloudAccountId() {
	logger.Log.Info("Starting Metering Search API Test with Missing cloudAccountId")
	payload := testutils.Get_Missing_Value("search", "cloudAccountId")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with Missing cloudAccountId")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingResourceId() {
	logger.Log.Info("Starting Metering Search API Test with Missing ResourceId")
	payload := testutils.Get_Missing_Value("search", "resourceId")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingId() {
	logger.Log.Info("Starting Metering Search API Test with missing Id")
	payload := testutils.Get_Missing_Value("search", "id")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with missing Id")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingTransactionId() {
	logger.Log.Info("Starting Metering Search API Test with missing TransactionId")
	payload := testutils.Get_Missing_Value("search", "transactionId")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with missing TransactionId")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingstartTime() {
	logger.Log.Info("Starting Metering Search API Test with missing Start Time")
	payload := testutils.Get_Missing_Value("search", "startTime")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with missing startTime")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingendTime() {
	logger.Log.Info("Starting Metering Search API Test with missing endTime")
	payload := testutils.Get_Missing_Value("search", "endTime")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with missing endTime")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithMissingreported() {
	logger.Log.Info("Starting Metering Search API Test with missing reported")
	payload := testutils.Get_Missing_Value("search", "reported")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with missing Reported")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithEmptyPayload() {
	logger.Log.Info("Starting Metering Search API Test with Empty Payload")
	payload := testutils.Get_Missing_Value("search", "empty")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "", "Test Failed to Search Record with Empty Payload")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithExtraField() {
	logger.Log.Info("Starting Metering Search API Test with Extra Field")
	payload := testutils.Get_Invalid_Value("search", "extraField")
	_, err := testutils.Execute_Read_Usage_Record_Gcurl(payload)
	logger.Logf.Infof("Output from Search Usage Record API Error --> %s\n", err)
	assert.Contains(suite.T(), err, "message type idc_metering.UsageFilter has no known field named extraField", "Test Failed to Search Record with Extra Field")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMeteringSuite(t *testing.T) {
	test = GRPCTestSuite{}
	suite.Run(t, new(GRPCTestSuite))

}
