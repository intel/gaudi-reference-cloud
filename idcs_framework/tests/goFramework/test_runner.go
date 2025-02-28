package main

import (
	"bytes"
	"flag"
	"fmt"
	"goFramework/framework/common/logger"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// usage - Ex: go run test_runner.go --framework "testify" --test_suite "APITest" --feature "Metering" --test_tag "MeteringRestSmokeSuite"

func main() {
	var framework string
	var testsuite string
	var testtag string
	var feature string

	// flags declaration using flag package
	flag.StringVar(&framework, "framework", "testify", "Specify framework. Default is testify")
	flag.StringVar(&testsuite, "test_suite", "APITest", "Specify test_suite. Default is REST API tests")
	flag.StringVar(&feature, "feature", "Metering", "Specify feature to be tested. Default is Metering")
	flag.StringVar(&testtag, "test_tag", "TinyVM", "Specify test_tag. Default is Functional")

	flag.Parse() // after declaring flags we need to call it

	// Build the test run comamnd for goFramework
	_, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Setenv("TestTag", testtag)
	test_path := filepath.Join("./", framework, "test_cases", testsuite, feature)
	tag_cmd := "--tags=" + testtag
	cmd := exec.Command("go", "test", "./"+test_path, "-v", tag_cmd, " -timeout 99999s")

	fmt.Println("Framework : ", framework)
	fmt.Println("Test Suite : ", testsuite)
	fmt.Println("Feature Being Tested : ", feature)
	fmt.Println("Test Tag : ", testtag)
	fmt.Println("Test Run Command Being Executed is : ", cmd)

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	// Execute the command
	if err := cmd.Run(); err != nil {
		logger.Logf.Error(err)
	}

}
