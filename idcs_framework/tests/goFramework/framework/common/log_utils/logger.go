package log_utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

// GeneralLogger exported
var Logger *log.Logger

// ErrorLogger exported
var ErrorLogger *log.Logger

// Test Logger

var TestLogger *log.Logger

//Report Portal Logger

var RpLogger *log.Logger



func init() {
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println("File path", filename)
	logfilepath := path.Join(path.Dir(filename), "../../../results")
	fmt.Println("LOG File path", logfilepath)
	testLogPath:= logfilepath+"/test-log.log"
	e := os.Remove(testLogPath)
	if e != nil {
		fmt.Println("Log file does not exists")
	}
	singleTestLogPath:=logfilepath+"/singletest-log.log"
	e = os.Remove(singleTestLogPath)
	if e != nil {
		fmt.Println("Log file does not exists")
	}
	generalLog, err := os.OpenFile(logfilepath+"/test-log.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	testLog, err := os.OpenFile(logfilepath+"/singletest-log.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	Logger = log.New(generalLog, "Info:\t", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(generalLog, "Error:\t", log.Ldate|log.Ltime|log.Lshortfile)
	TestLogger = log.New(testLog, "Info:\t", log.Ldate|log.Ltime|log.Lshortfile)
}
