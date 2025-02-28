package report_portal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/avarabyeu/goRP/gorp"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
)

// NewBase creates a new base formatter.
func NewBase(suite string, out io.Writer) *Base {
	return &Base{
		suiteName: suite,
		indent:    2,
		out:       out,
		Lock:      new(sync.Mutex),
	}
}

// Base is a base formatter.
type Base struct {
	suiteName string
	out       io.Writer
	indent    int
	Lock      *sync.Mutex
}

type User1 struct {
	Name string `json:"name"`
	Job  string `json:"job"`
}

type SampleTestSuite struct {
	suite.Suite
}

// The sequence of type structs are used to marshall the json object.
type rpComment struct {
	Value string `json:"value"`
	Line  int    `json:"line"`
}

type launchData struct {
	launchId     string
	launchName   string
	suiteId      string
	suiteStatus  string
	launchStatus string
}

type TestData struct {
	TestId        string
	CurrentStepID string
	CurrentTestID string
	TestStatus    string
	Name          string
	Description   string
	TestUUID      *uuid.UUID
	TestLog       string
}

type SuiteData struct {
	Name        string
	Description string
}

type TestStep struct {
	Name          string
	Description   string
	StepStatus    string
	CurrentStepID string
	ErrorMessage  string
	LogLevel      string
	LogMessage    string
}
type rpDocstring struct {
	Value       string `json:"value"`
	ContentType string `json:"content_type"`
	Line        int    `json:"line"`
}

type rpTag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

type rpResult struct {
	Status   string `json:"status"`
	Error    string `json:"error_message,omitempty"`
	Duration *int   `json:"duration,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type rpMatch struct {
	Location string `json:"location"`
}

type Data struct {
	Rpendpoint        string
	Rpuuid            string
	Rpproject         string
	Rplaunch          string
	Launchdescription string
	Attributes        []*gorp.Attribute
}

var payload Data

func RpFunc() *gorp.Client {
	// Read Config.json
	_, filename, _, _ := runtime.Caller(0)
	suite_path := "../../../" + os.Getenv("Test_Suite") + "/test_config/config.json"
	fmt.Println("Suite Path", suite_path)
	filepath := path.Join(path.Dir(filename), suite_path)
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`

	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	// Start RP Connection
	goRP := gorp.NewClient(
		payload.Rpendpoint,
		payload.Rpproject,
		payload.Rpuuid,
	)

	return goRP
}

var Launchdata = &launchData{}

// StartLaunch - Start Launch in RP
func StartLaunch(ld *launchData) {
	fmt.Println("Start launch")
	f := RpFunc()
	launchRqData := &gorp.StartLaunchRQ{
		StartRQ: gorp.StartRQ{
			Name:        payload.Rplaunch,
			Attributes:  payload.Attributes,
			StartTime:   gorp.Timestamp{time.Now()},
			Description: payload.Launchdescription,
		},
		Mode: "Default",
	}
	launchRsData, err := f.StartLaunch(launchRqData)
	if err != nil {
		fmt.Println("Error no adding launch to RP")
		fmt.Println(err.Error())
	}
	currentLaunchID := launchRsData.ID
	ld.launchId = launchRsData.ID
	fmt.Println("Start launch", currentLaunchID)
}

// FinishLaunch - Finish Launch in RP
func FinishLaunch(ld *launchData) {
	f := RpFunc()
	fmt.Println("Finish Launch", ld.launchId)
	launchRqData := &gorp.FinishExecutionRQ{
		EndTime: gorp.Timestamp{time.Now()},
		Status:  "PASSED",
	}
	_, err := f.FinishLaunch(ld.launchId, launchRqData)
	if err != nil {
		fmt.Println("Error no finishing launch to RP")
		fmt.Println(err.Error())
	}
}

// StartSuite - Start Feature
func StartSuite(ld *launchData, sd *SuiteData, td *TestData) {
	f := RpFunc()
	testRqData := &gorp.StartTestRQ{
		StartRQ: gorp.StartRQ{
			Name:        sd.Name,
			Description: sd.Description,
			StartTime:   gorp.Timestamp{time.Now()},
		},
		LaunchID: ld.launchId,
		Type:     "SUITE",
	}
	testRsData, err := f.StartTest(testRqData)
	if err != nil {
		fmt.Println("Error on adding suite (feature) to RP")
		fmt.Println(err.Error())
	}
	currentSuiteID := testRsData.ID
	ld.suiteId = testRsData.ID
	fmt.Println("Start Suite", currentSuiteID)

	// Save logs

	logTest := &gorp.SaveLogRQ{
		ItemID:  ld.suiteId,
		Level:   "Info",
		LogTime: gorp.Timestamp{time.Now()},
		Message: td.TestLog,
	}
	logData, err := f.SaveLog(logTest)
	if err != nil {
		fmt.Println("Error on saving log to RP")
		fmt.Println(err.Error())
	}
	logId := logData.ID
	fmt.Println("Log id", logId)
}

// FinishSuite - Finish BDD Feature
func FinishSuite(ld *launchData, td *TestData) {
	f := RpFunc()
	// Save logs

	logTest := &gorp.SaveLogRQ{
		ItemID:  ld.suiteId,
		Level:   "Info",
		LogTime: gorp.Timestamp{time.Now()},
		Message: td.TestLog,
	}
	logData, err := f.SaveLog(logTest)
	if err != nil {
		fmt.Println("Error on saving log to RP")
		fmt.Println(err.Error())
	}
	logId := logData.ID
	fmt.Println("Log id", logId)

	fmt.Println("Finish Suite", ld.suiteId)
	testRqData := &gorp.FinishTestRQ{
		FinishExecutionRQ: gorp.FinishExecutionRQ{
			EndTime: gorp.Timestamp{time.Now()},
		},
	}
	finTest, err := f.FinishTest(ld.suiteId, testRqData)
	if err != nil {
		fmt.Println("Error on finishing suite (feature) to RP", finTest)
		fmt.Println(err.Error())
	}
	fmt.Println("Result of test", finTest)
}

// StartChildContainer - Start BDD Scenario
func StartChildContainer(ld *launchData, td *TestData) {
	f := RpFunc()
	testRqData := &gorp.StartTestRQ{
		StartRQ: gorp.StartRQ{
			Name:        td.Name,
			Description: td.Description,
			StartTime:   gorp.Timestamp{time.Now()},
		},
		LaunchID: ld.launchId,
		Type:     "TEST",
	}
	testRsData, err := f.StartChildTest(ld.suiteId, testRqData)
	if err != nil {
		fmt.Println("Error on adding child container (scenario) to RP")
		fmt.Println(err.Error())
	}
	td.CurrentTestID = testRsData.ID
}

// FinishChildContainer - Finish BDD Scenario
func FinishChildContainer(td *TestData) {
	fmt.Println("Finish Child Container")
	f := RpFunc()
	testRqData := &gorp.FinishTestRQ{
		FinishExecutionRQ: gorp.FinishExecutionRQ{
			EndTime: gorp.Timestamp{time.Now()},
			Status:  "PASSED",
		},
	}
	_, err := f.FinishTest(td.CurrentTestID, testRqData)
	if err != nil {
		fmt.Println("Error on finishing child container (scenario) to RP")
		fmt.Println(err.Error())
	}
}

// StartChildStep - Start BDD Step
func StartChildStep(ld *launchData, ts *TestStep, td *TestData) {
	f := RpFunc()
	testRqData := &gorp.StartTestRQ{
		StartRQ: gorp.StartRQ{
			Name:        ts.Name,
			Description: ts.Description,
			StartTime:   gorp.Timestamp{time.Now()},
		},
		LaunchID: ld.launchId,
		Type:     "STEP",
	}
	testRsData, err := f.StartChildTest(td.CurrentTestID, testRqData)
	if err != nil {
		fmt.Println("Error on adding step to RP")
		fmt.Println(err.Error())
	}
	ts.CurrentStepID = testRsData.ID
}

// FinishChildStep - Finish BDD Step
func FinishChildStep(ld *launchData, ts *TestStep, td *TestData) {
	f := RpFunc()
	// Send log Error if there is an error in step
	if ts.StepStatus == "FAILED" {
		logTest := &gorp.SaveLogRQ{
			ItemID:  ts.CurrentStepID,
			LogTime: gorp.Timestamp{time.Now()},
			Message: td.TestLog,
			Level:   "Error",
		}
		_, err := f.SaveLog(logTest)
		if err != nil {
			fmt.Println("Error on saving log to RP")
			fmt.Println(err.Error())
		}
	}
	if ts.StepStatus != "FAILED" {
		logTest := &gorp.SaveLogRQ{
			ItemID:  ts.CurrentStepID,
			LogTime: gorp.Timestamp{time.Now()},
			Message: td.TestLog,
			Level:   "Info",
		}
		_, err := f.SaveLog(logTest)
		if err != nil {
			fmt.Println("Error on saving log to RP")
			fmt.Println(err.Error())
		}
	}

	// Finish Step
	testRqData := &gorp.FinishTestRQ{
		FinishExecutionRQ: gorp.FinishExecutionRQ{
			EndTime: gorp.Timestamp{time.Now()},
			Status:  ts.StepStatus,
		},
	}
	_, err := f.FinishTest(ts.CurrentStepID, testRqData)
	if err != nil {
		fmt.Println("Error on finishing step to RP")
		fmt.Println(err.Error())
	}
}
