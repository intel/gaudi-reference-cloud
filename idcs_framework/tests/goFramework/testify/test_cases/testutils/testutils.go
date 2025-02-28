package testutils

import (
	"bytes"
	"goFramework/framework/common/logger"
	"os/exec"
)

func GetIntPointer(value int) *int {
	return &value
}

func GetStringPointer(value string) *string {
	return &value
}

func GetUintPointer(value uint) *uint {
	return &value
}

func GetUint64Pointer(value uint64) *uint64 {
	return &value
}

func GetBoolPointer(value bool) *bool {
	return &value
}

func GetInt64Pointer(value int64) *int64 {
	return &value
}

func Execute_Command(commandString string) (string, string) {

	cmd := exec.Command(commandString)
	logger.Logf.Infof("Command to execute : %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("Error: ", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr, errStr

}
