// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks_integration_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type HostConfig struct {
	Host                            string `yaml:"host"`
	Environment                     string `yaml:"environment"`
	GlobalHost                      string `yaml:"global_host"`
	DefaultAccount                  string `yaml:"default_account"`
	BearerToken                     string `yaml:"bearer_token"`
	AdminToken                      string `yaml:"admin_token"`
	BearerTokenExpiry               int64  `yaml:"bearer_token_expiry"`
	AdminTokenExpiry                int64  `yaml:"admin_token_expiry"`
	CreateCluster                   bool   `yaml:"create_cluster"`
	CreateNodeGroup                 bool   `yaml:"create_node_group"`
	AddNodeToNodeGroup              bool   `yaml:"add_node_to_node_group"`
	DeleteNodeFromNodeGroup         bool   `yaml:"delete_node_from_node_group"`
	DeleteSpecificNodeGroup         bool   `yaml:"delete_specific_node_group"`
	DeleteSpecificVIP               bool   `yaml:"delete_specific_vip"`
	DeleteCluster                   bool   `yaml:"delete_cluster"`
	CreateVIP                       bool   `yaml:"create_vip"`
	DownloadKubeConfig              bool   `yaml:"download_kubeconfig"`
	CreateClusterTimeOutInMinutes   int    `yaml:"create_cluster_timeout_in_min"`
	CreateNodeGroupTimeOutInMinutes int    `yaml:"create_node_group_timeout_in_min"`
	CreateILBTimeOutInMinutes       int    `yaml:"create_vip_timeout_in_min"`
	RunStorageTests                 bool   `yaml:"run_storage_tests"`
}

type CloudAccountDetails struct {
	AccountID string `json:"id"`
	Email     string `json:"name"`
	Enrolled  bool   `json:"enrolled"`
}

func ReadHostConfig(filename string) (HostConfig, error) {
	var hostConfig HostConfig
	file, err := os.Open(filename)
	if err != nil {
		return hostConfig, errors.Wrap(err, "failed to open config file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return hostConfig, errors.Wrap(err, "failed to read config file")
	}

	err = yaml.Unmarshal(data, &hostConfig)
	if err != nil {
		return hostConfig, errors.Wrap(err, "failed to unmarshal config file")
	}

	return hostConfig, nil
}

func ReadRequestData(filename string) ([]byte, error) {
	var requestData []byte
	file, err := os.Open(filename)
	if err != nil {
		return requestData, errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return requestData, errors.Wrap(err, "failed to read file")
	}

	return data, nil
}

func SendHttpRequest(hostUrl string, method string, payload io.Reader, bearerToken string) ([]byte, error) {

	proxyURL := os.Getenv("https_proxy")

	client := &http.Client{}

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxy), TLSClientConfig: &tls.Config{InsecureSkipVerify: false}}
	}

	req, err := http.NewRequest(method, hostUrl, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", bearerToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func makeGRPCCall(requestBody string, serverAddress string, packageName string, methodName string, caCertPath, clientCertPath, clientKeyPath string) (string, error) {
	cmd := exec.Command("grpcurl",
		"-d", fmt.Sprintf("'%s'", requestBody),
		"--cacert", caCertPath,
		"--cert", clientCertPath,
		"--key", clientKeyPath,
		serverAddress,
		fmt.Sprintf("%s/%s", packageName, methodName),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to make GRPC Call: %s", string(output))
	}

	return string(output), nil
}

func randomize(name string) string {
	return name + "-" + fmt.Sprint(rand.Intn(10000))
}

func isTokenExpired(expiry int64) bool {

	currentTime := time.Now().Unix()

	return expiry <= currentTime
}
