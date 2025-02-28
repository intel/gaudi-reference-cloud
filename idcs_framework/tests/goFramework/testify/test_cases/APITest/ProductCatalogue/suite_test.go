package ProductCatalogue

import (
	"bytes"
	"goFramework/framework/common/logger"
	"goFramework/framework/common/report_portal"
	"goFramework/framework/library/auth"
	"goFramework/utils"
	"io"
	"log"
	"os"
	_ "path/filepath"
	_ "runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

var rescueStdout *os.File
var r *os.File
var w *os.File

type PcAPITestSuite struct {
	suite.Suite
	buf       bytes.Buffer
	testdata  report_portal.TestData
	suitedata report_portal.SuiteData
	stepdata  report_portal.TestStep
}

func (suite *PcAPITestSuite) SetupSuite() {
	os.Setenv("Test_Suite", "testify")
	os.Setenv("fetch_admin_role_token", "True")
	auth.Get_config_file_data("../../../test_config/config.json")
	logger.InitializeZapCustomLogger()
	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	logger.Log.Info("Starting API  test suite")
	logger.Log.Info("Starting Report Portal Setup")
	logger.Log.Info("-- Start Launch --")
	report_portal.StartLaunch(report_portal.Launchdata)
	logger.Log.Info("-- Start Suite --")
	suite.suitedata.Name = "IDC API Test Suite"
	suite.suitedata.Description = "Suite - IDC API End Point Tests"
	report_portal.StartSuite(report_portal.Launchdata, &suite.suitedata, &suite.testdata)
	suite.testdata.Name = suite.T().Name()
	logger.Log.Info("-- Start Test --")
	log.SetOutput(&suite.buf)
	suite.testdata.Description = "Test - IDC API Test"
	report_portal.StartChildContainer(report_portal.Launchdata, &suite.testdata)
	// Add Generate token method here
	utils.Get_PC_config_file_path()
	utils.Get_PC_config_file_data()
	utils.LoadCAConfig("../../../test_config/product_catalog.json")

}

func (suite *PcAPITestSuite) TearDownSuite() {
	logger.Log.Info("Finishing Sample test suite")
	logger.Log.Info("-- Finish Test --")
	report_portal.FinishChildContainer(&suite.testdata)
	logger.Log.Info("-- Finish Suite --")
	report_portal.FinishSuite(report_portal.Launchdata, &suite.testdata)
	logger.Log.Info("-- Finish Launch --")
	report_portal.FinishLaunch(report_portal.Launchdata)
}

func (suite *PcAPITestSuite) SetupTest() {
	logger.Log.Info("Starting IDC API Functional test cases")
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

func (suite *PcAPITestSuite) TearDownTest() {
	logger.Log.Info("In teardown IDC API Functional test cases")
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProductsAPITestSuite(t *testing.T) {
	suite.Run(t, new(PcAPITestSuite))
}
