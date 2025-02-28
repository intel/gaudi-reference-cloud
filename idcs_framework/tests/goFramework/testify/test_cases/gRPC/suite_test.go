package GRPCTest

import (
	"bytes"
	"context"
	"goFramework/framework/common/grpc_client"
	"goFramework/framework/common/logger"
	"goFramework/framework/common/report_portal"
	"io"
	"log"
	"os"
	_ "path/filepath"
	_ "runtime"

	"go.uber.org/zap"

	pb "goFramework/framework/library/grpc/metering/pkg/gen/pb/metering/v1"

	"github.com/stretchr/testify/suite"
)

const (
	RECORDS_COUNT_ALL       int    = 10
	RECORDS_COUNT_ONE       int    = 1
	RECORDS_READAFTER_ID    uint64 = 10
	RECORDS_READAFTER_COUNT int    = 990
)

var rescueStdout *os.File
var r *os.File
var w *os.File
var MeteringClient pb.MeteringServiceClient
var Logger *zap.Logger

type GRPCTestSuite struct {
	suite.Suite
	buf        bytes.Buffer
	testdata   report_portal.TestData
	suitedata  report_portal.SuiteData
	stepdata   report_portal.TestStep
	TestStruct grpc_client.Test
}

func (suite *GRPCTestSuite) SetupSuite() {
	os.Setenv("Test_Suite", "testify")
	logger.InitializeZapCustomLogger()
	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	logger.Log.Info("Starting gRPC  test suite")
	logger.Log.Info("Starting Report Portal Setup")
	logger.Log.Info("-- Start Launch --")
	report_portal.StartLaunch(report_portal.Launchdata)
	logger.Log.Info("-- Start Suite --")
	suite.suitedata.Name = "IDC gRPC Test Suite"
	suite.suitedata.Description = "Suite - IDC gRPC End Point Tests"
	report_portal.StartSuite(report_portal.Launchdata, &suite.suitedata, &suite.testdata)
	suite.testdata.Name = suite.T().Name()
	logger.Log.Info("-- Start Test --")
	log.SetOutput(&suite.buf)
	suite.testdata.Description = "Test - IDC gRPC Test"
	report_portal.StartChildContainer(report_portal.Launchdata, &suite.testdata)
	suite.TestStruct = grpc_client.Test{}
	logger.Log.Info("Starting Grpc")
	suite.TestStruct.GrpcInit()
	logger.Logf.Infof("Setting up Client connection", suite.TestStruct.ClientConn)
	MeteringClient = pb.NewMeteringServiceClient(suite.TestStruct.ClientConn)
	logger.Logf.Infof("Metering  Client connection", MeteringClient)
}

func (suite *GRPCTestSuite) TearDownSuite() {
	logger.Log.Info("Finishing Sample test suite")
	logger.Log.Info("-- Finish Test --")
	report_portal.FinishChildContainer(&suite.testdata)
	logger.Log.Info("-- Finish Suite --")
	report_portal.FinishSuite(report_portal.Launchdata, &suite.testdata)
	logger.Log.Info("-- Finish Launch --")
	report_portal.FinishLaunch(report_portal.Launchdata)
	suite.TestStruct.GrpcDone()
}

func (suite *GRPCTestSuite) SetupTest() {
	logger.Log.Info("Starting IDC gRPC Functional test cases")
	ctx := context.Background()
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Executing Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	suite.TestStruct.DropUsageRecordTable(ctx)
	suite.stepdata.Name = suite.T().Name()
	report_portal.StartChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)

}

func (suite *GRPCTestSuite) TearDownTest() {
	logger.Log.Info("In teardown IDC gRPC Functional test cases")
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	//out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	suite.testdata.TestLog = buf.String()
	logger.Log.Info("-- Finish Test Step --")
	if suite.T().Failed() == true {
		suite.stepdata.StepStatus = "FAILED"
	} else {
		suite.stepdata.StepStatus = "PASSED"
	}
	report_portal.FinishChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Finished Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
}
