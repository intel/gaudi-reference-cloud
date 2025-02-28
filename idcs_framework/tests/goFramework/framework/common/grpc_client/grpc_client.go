package grpc_client

import (
	"bytes"
	"fmt"
	"goFramework/framework/common/logger"
	"os/exec"
)

func ExecuteGrpcCurlRequest(payload string, grpc_host string, endpoint string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", grpc_host, endpoint)
	logger.Logf.Infof("Executing command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	logger.Logf.Infof("Command output:\n%s\n error:\n%s\n", outStr, errStr)
	return outStr, errStr
}

func ExecuteGrpcCurlRequestTLS(payload string, token string, grpc_host string, endpoint string) (string, string) {
	var cmd *exec.Cmd

	// Construct the Authorization header correctly
	authHeader := fmt.Sprintf("Authorization: Bearer %s", token)

	if payload != "" {
		cmd = exec.Command("grpcurl", "-d", payload, "-H", authHeader, grpc_host, endpoint)
	} else {
		fmt.Println("No payload")
		cmd = exec.Command("grpcurl", "-H", authHeader, grpc_host, endpoint)
	}

	fmt.Println("Executing command: ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd.Run() failed with ", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	fmt.Println("Command output: ", outStr, " error: ", errStr)
	return outStr, errStr
}
