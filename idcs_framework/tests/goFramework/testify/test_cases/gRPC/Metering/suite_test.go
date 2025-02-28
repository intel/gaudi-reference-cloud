package GRPCTest

import (
	"bytes"

	"goFramework/framework/common/logger"
	"goFramework/framework/common/report_portal"
	"goFramework/utils"
	_ "io"
	"os"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var rescueStdout *os.File
var r *os.File
var w *os.File
var Logger *zap.Logger

type GRPCTestSuite struct {
	suite.Suite
	buf       bytes.Buffer
	testdata  report_portal.TestData
	suitedata report_portal.SuiteData
	stepdata  report_portal.TestStep
}

func (suite *GRPCTestSuite) SetupSuite() {
	os.Setenv("Test_Suite", "testify")
	logger.InitializeZapCustomLogger()
	// rescueStdout = os.Stdout
	// r, w, _ = os.Pipe()
	// os.Stdout = w
	logger.Log.Info("Starting gRPC  test suite")
	logger.Log.Info("Starting Report Portal Setup")
	logger.Log.Info("-- Start Launch --")
	//report_portal.StartLaunch(report_portal.Launchdata)
	logger.Log.Info("-- Start Suite --")
	suite.suitedata.Name = "IDC gRPC Test Suite"
	suite.suitedata.Description = "Suite - IDC gRPC End Point Tests"
	//report_portal.StartSuite(report_portal.Launchdata, &suite.suitedata, &suite.testdata)
	suite.testdata.Name = suite.T().Name()
	logger.Log.Info("-- Start Test --")
	//log.SetOutput(&suite.buf)
	suite.testdata.Description = "Test - IDC gRPC Test"
	utils.Get_config_file_data()
	utils.Get_Metering_config_file_data()
	//report_portal.StartChildContainer(report_portal.Launchdata, &suite.testdata)

}

func (suite *GRPCTestSuite) TearDownSuite() {
	logger.Log.Info("Finishing Sample test suite")
	logger.Log.Info("-- Finish Test --")
	//report_portal.FinishChildContainer(&suite.testdata)
	logger.Log.Info("-- Finish Suite --")
	//report_portal.FinishSuite(report_portal.Launchdata, &suite.testdata)
	logger.Log.Info("-- Finish Launch --")
	//report_portal.FinishLaunch(report_portal.Launchdata)
}

func (suite *GRPCTestSuite) SetupTest() {
	logger.Log.Info("Starting IDC gRPC Functional test cases")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Executing Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	suite.stepdata.Name = suite.T().Name()
	//report_portal.StartChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
}

func (suite *GRPCTestSuite) TearDownTest() {
	logger.Log.Info("In teardown IDC gRPC Functional test cases")
	// w.Close()
	// var buf bytes.Buffer
	// io.Copy(&buf, r)
	// //out, _ := ioutil.ReadAll(r)
	// os.Stdout = rescueStdout
	// suite.testdata.TestLog = buf.String()
	logger.Log.Info("-- Finish Test Step --")
	if suite.T().Failed() == true {
		suite.stepdata.StepStatus = "FAILED"
	} else {
		suite.stepdata.StepStatus = "PASSED"
	}
	//report_portal.FinishChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Finished Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
}
