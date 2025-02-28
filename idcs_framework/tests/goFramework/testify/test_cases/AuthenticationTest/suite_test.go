package AuthenticationTest

import (
	"bytes"
	"goFramework/framework/common/logger"
	"goFramework/framework/common/report_portal"
	"goFramework/utils"
	"io"
	"log"
	"os"
	_ "path/filepath"
	_ "runtime"

	"github.com/stretchr/testify/suite"
)

var rescueStdout *os.File
var r *os.File
var w *os.File

type authNTestSuite struct {
	suite.Suite
	buf       bytes.Buffer
	testdata  report_portal.TestData
	suitedata report_portal.SuiteData
	stepdata  report_portal.TestStep
}

func (suite *authNTestSuite) SetupSuite() {
	os.Setenv("Test_Suite", "testify")
	logger.InitializeZapCustomLogger()
	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	logger.Log.Info("Starting AuthN API test suite")
	logger.Log.Info("Starting Report Portal Setup")
	logger.Log.Info("-- Start Launch --")
	report_portal.StartLaunch(report_portal.Launchdata)
	logger.Log.Info("-- Start Suite --")
	suite.suitedata.Name = "IDC AuthN API Test Suite"
	suite.suitedata.Description = "Suite - IDC AuthN API End Point Tests"
	report_portal.StartSuite(report_portal.Launchdata, &suite.suitedata, &suite.testdata)
	suite.testdata.Name = suite.T().Name()
	logger.Log.Info("-- Start Test --")
	log.SetOutput(&suite.buf)
	suite.testdata.Description = "Test - IDC API Test"
	report_portal.StartChildContainer(report_portal.Launchdata, &suite.testdata)

	/*
		TODO
		Can be moved to suite level or function level (before method or after method).
		Keeping it here locally till we sort out framework level changes
	*/
	utils.Get_config_file_data()
	utils.LoadAuthNConfig("../../test_config/authentication_resources/authn_api_input.json")

}

func (suite *authNTestSuite) TearDownSuite() {
	logger.Log.Info("-- Finish Test --")
	report_portal.FinishChildContainer(&suite.testdata)
	logger.Log.Info("-- Finish Suite --")
	report_portal.FinishSuite(report_portal.Launchdata, &suite.testdata)
	logger.Log.Info("-- Finish Launch --")
	report_portal.FinishLaunch(report_portal.Launchdata)
}

func (suite *authNTestSuite) SetupTest() {
	logger.Log.Info("Starting AuthN API Functional test cases")
	suite.stepdata.Name = suite.T().Name()
	report_portal.StartChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Executing Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
}

func (suite *authNTestSuite) TearDownTest() {
	logger.Log.Info("In teardown AuthN API Functional test cases")
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
