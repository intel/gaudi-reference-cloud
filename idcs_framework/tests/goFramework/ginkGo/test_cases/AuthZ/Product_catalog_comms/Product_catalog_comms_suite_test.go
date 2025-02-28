package product_catalog_service_comms_test

import (
	"bytes"
	"context"
	"goFramework/framework/common/logger"
	"goFramework/framework/common/report_portal"
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mTLSTestSuite struct {
	buf       bytes.Buffer
	testdata  report_portal.TestData
	suitedata report_portal.SuiteData
	stepdata  report_portal.TestStep
}

var (
	ctx    context.Context
	cancel context.CancelFunc
	suite  = &mTLSTestSuite{}
)

func TestProductCatalogComms(t *testing.T) {
	if os.Getenv("MULTI_RUNNER") != "" {
		t.Skip("Skipping not suitable for multi runner container")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProductCatalogComms Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
	logger.Log.Info("Starting mTLS  test suite")
	logger.Log.Info("Starting Report Portal Setup")
	ctx, cancel = context.WithCancel(context.Background())
	logger.Log.Info("-- Start Launch --")
	report_portal.StartLaunch(report_portal.Launchdata)
	logger.Log.Info("-- Start Suite --")
	suite.suitedata.Name = "IDC mTLS Test Suite"
	suite.suitedata.Description = "Suite - IDC mTLS Tests"
	report_portal.StartSuite(report_portal.Launchdata, &suite.suitedata, &suite.testdata)
	suite.testdata.Name = "Test Suite"
	logger.Log.Info("-- Start Test --")
	log.SetOutput(&suite.buf)
	suite.testdata.Description = "Test - IDC mTLS Test"
	report_portal.StartChildContainer(report_portal.Launchdata, &suite.testdata)
})

var _ = AfterSuite(func() {
	logger.Log.Info("Finishing Sample test suite")
	logger.Log.Info("-- Finish Test --")
	report_portal.FinishChildContainer(&suite.testdata)
	logger.Log.Info("-- Finish Suite --")
	report_portal.FinishSuite(report_portal.Launchdata, &suite.testdata)
	logger.Log.Info("-- Finish Launch --")
	report_portal.FinishLaunch(report_portal.Launchdata)
})

var _ = BeforeEach(func() {
	logger.Log.Info("Starting IDC mTLS Functional test cases")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Starting Test  Test ------------ %s ", GinkgoT().Name())
	logger.Log.Info("################################################################################################################")
	suite.stepdata.Name = GinkgoT().Name()
	report_portal.StartChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
})

var _ = AfterEach(func() {
	logger.Log.Info("In teardown IDC mTLS Functional test cases")
	suite.testdata.TestLog = CurrentSpecReport().CapturedGinkgoWriterOutput
	logger.Log.Info("-- Finish Test Step --")
	if CurrentSpecReport().Failed() {
		suite.stepdata.StepStatus = "FAILED"
	} else {
		suite.stepdata.StepStatus = "PASSED"
	}

	report_portal.FinishChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Finished Test   ------------ %s ", GinkgoT().Name())
	logger.Log.Info("################################################################################################################")
})
