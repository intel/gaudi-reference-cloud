// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks_integration_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"
	"github.com/rancher/lasso/pkg/log"

	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
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
	RunSecRulesTests                bool   `yaml:"run_security_rules_tests"`
}

type CloudAccountDetails struct {
	AccountID string `json:"id"`
	Email     string `json:"name"`
	Enrolled  bool   `json:"enrolled"`
}

// Define the struct for the firewall response
type FirewallResponse struct {
	SourceIP      []string `json:"sourceip"`
	State         string   `json:"state"`
	DestinationIP string   `json:"destinationip"`
	Port          int      `json:"port"`
	VIPID         int      `json:"vipid"`
	VIPName       string   `json:"vipname"`
	VIPType       string   `json:"viptype"`
	Protocol      []string `json:"protocol"`
	InternalPort  int      `json:"internalport"`
}

// Define the struct for the top-level response
type GetFirewallResponse struct {
	GetFirewallResponse []FirewallResponse `json:"getfirewallresponse"`
}

type createVIPType struct {
	Description string `json:"description"`
	Viptype     string `json:"viptype"`
	Name        string `json:"name"`
	Port        int32  `json:"port"`
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
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxy), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
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

// Get all security rules for specific cluster
func GetSecRules(clusterUUID string) ([]byte, error) {
	now := time.Now()
	currTimeinMS := now.UnixNano() / int64(time.Millisecond)

	getSecRuleUrlUpdated := fmt.Sprintf(getSecRuleUrl, clusterUUID, currTimeinMS)

	data, err := SendHttpRequest(getSecRuleUrlUpdated, getMethod, nil, bearerToken)
	return data, err
}

// Check that the security rules for specified LB have the correct sourceIPs and protocol
func CheckSecRuleExists(clusterUUID, vipName, vipIp string, expectedSourceIps, expectedProtocols []string, port int32) (bool, error) {
	log.Infof("Checking if security rule exists for clusterUUID: %s, vipName: %s", clusterUUID, vipName)

	resp, err := GetSecRules(clusterUUID)
	if err != nil {
		return false, err
	}

	log.Infof("Security Rules retrieved from cluster:")
	log.Infof(string(resp))

	// convert response into array
	var secRules GetFirewallResponse
	err = json.Unmarshal(resp, &secRules)
	if err != nil {
		return false, err
	}

	// Sort the expected arrays for easier comparison
	sort.Strings(expectedSourceIps)
	sort.Strings(expectedProtocols)

	// search through security rules for matching source IPs and protocol
	for j := 0; j < len(secRules.GetFirewallResponse); j++ {
		if secRules.GetFirewallResponse[j].VIPName == vipName {
			// compare the source IPs and protocols
			sort.Strings(secRules.GetFirewallResponse[j].SourceIP)
			log.Infof("j = %d, source IP - Expected: %s, Actual: %s", j, expectedSourceIps, secRules.GetFirewallResponse[j].SourceIP)

			sort.Strings(secRules.GetFirewallResponse[j].Protocol)
			log.Infof("j = %d, protocols - Expected: %s, Actual: %s", j, expectedProtocols, secRules.GetFirewallResponse[j].Protocol)

			if reflect.DeepEqual(secRules.GetFirewallResponse[j].SourceIP, expectedSourceIps) && reflect.DeepEqual(secRules.GetFirewallResponse[j].Protocol, expectedProtocols) {
				log.Infof("source IPs and protocols match")
				return true, nil
			} else {
				log.Infof("source IPs and/or protocols do not match")
				return false, nil
			}
		}
	}

	// VIP with matching name was not found, so return false
	return false, nil
}

// Wait for security rule for specific load balancer to get to Active state
func WaitForSecRuleState(clusterId, vipName, vipIp string, port int32, targetState string) (bool, error) {

	log.Infof("Waiting for security rules for load balancer %s to get to [%s] state.", vipName, targetState)

	i := 0
	maxWaitTimeInMinutes := 60
	for i = 0; i < maxWaitTimeInMinutes; i++ {
		resp, err := GetSecRules(clusterId)
		if err != nil {
			return false, err
		}

		// convert response into array
		var secRules GetFirewallResponse
		err = json.Unmarshal(resp, &secRules)
		if err != nil {
			return false, err
		}

		// search through security rules for matching vip name
		vipFound := false
		for j := 0; j < len(secRules.GetFirewallResponse); j++ {
			if secRules.GetFirewallResponse[j].VIPName == vipName {
				vipFound = true
				log.Infof("Current state: %s, Expected state: %s", secRules.GetFirewallResponse[j].State, targetState)
				if secRules.GetFirewallResponse[j].State == targetState {
					return true, nil
				}
			}
		}

		if vipFound == false {
			return false, errors.New("VIP not found")
		}

		log.Infof("Iteration: %d, Sleeping for 60 seconds to allow security rule state to update", i)
		time.Sleep(60 * time.Second)
	}

	return false, errors.New(fmt.Sprintf("Timed out waiting for security rule state to reach %s", targetState))
}

// Update a security rule
// Note: The internalIp and vipPort are used to identify the target load balancer.
func UpdateSecRule(clusterUUID, internalIp string, sourceIps []string, vipPort int32, protocols []string) error {
	now := time.Now()
	currTimeinMS := now.UnixNano() / int64(time.Millisecond)

	// Update the cluster Id and VIP UUID in updateSecRuleUrl
	updateSecRuleUrlUpdated := fmt.Sprintf(updateSecRuleUrl, clusterUUID, currTimeinMS)
	requestData, err := ReadRequestData("requests/update_security_rule_request.json")

	type updateSecRuleType struct {
		SourceIP   []string `json:"sourceip"`
		InternalIp string   `json:"internalip"`
		Port       int32    `json:"port"`
		Protocol   []string `json:"protocol"`
	}
	var updateSecRuleDetails updateSecRuleType
	json.Unmarshal(requestData, &updateSecRuleDetails)

	updateSecRuleDetails.SourceIP = sourceIps
	updateSecRuleDetails.InternalIp = internalIp
	updateSecRuleDetails.Port = vipPort
	updateSecRuleDetails.Protocol = protocols

	req, err := json.Marshal(updateSecRuleDetails)
	if err != nil {
		return err
	}

	body := bytes.NewReader(req)

	_, err = SendHttpRequest(updateSecRuleUrlUpdated, putMethod, body, bearerToken)
	return err
}

func DeleteSecRule(clusterUUID string, vipID int32) ([]byte, error) {
	now := time.Now()
	currTimeinMS := now.UnixNano() / int64(time.Millisecond)

	delSecRuleUrlUpdated := fmt.Sprintf(deleteSecRuleUrl, clusterUUID, vipID, currTimeinMS)
	data, err := SendHttpRequest(delSecRuleUrlUpdated, deleteMethod, nil, bearerToken)

	return data, err
}

// Given the security rule name, this method will return the internal IP and port
// If security rule name is not found then it will return an error
// Returns Internal IP, VIP ID, Port and error
func GetSecRulesInfoByName(clusterId, vipName string) (string, int32, int32, error) {

	log.Infof("Search for VIP with name: %s", vipName)

	resp, err := GetSecRules(clusterId)
	if err != nil {
		return "", 0, 0, err
	}

	// convert response into array
	var secRules GetFirewallResponse
	err = json.Unmarshal(resp, &secRules)
	if err != nil {
		return "", 0, 0, err
	}

	// search through security rules for matching vip name
	for j := 0; j < len(secRules.GetFirewallResponse); j++ {
		if secRules.GetFirewallResponse[j].VIPName == vipName {
			return secRules.GetFirewallResponse[j].DestinationIP, int32(secRules.GetFirewallResponse[j].VIPID), int32(secRules.GetFirewallResponse[j].Port), nil
		}
	}

	return "", 0, 0, errors.New("VIP not found")
}

func CreateVIP(clusterUUID, vipName, vipType string, port int32, description string) ([]byte, error) {
	log.Infof("Creating load balancer - Name: %s, Type: %s, Port: %d", vipName, vipType, port)
	requestData, err := ReadRequestData("requests/create_load_balancer_request.json")
	if err != nil {
		return nil, err
	}

	var createVIPDetails createVIPType
	json.Unmarshal(requestData, &createVIPDetails)

	createVIPDetails.Name = vipName
	createVIPDetails.Viptype = vipType
	createVIPDetails.Port = port
	createVIPDetails.Description = description

	req, err := json.Marshal(createVIPDetails)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(req)
	createVIPSUrlUpdated := fmt.Sprintf(createVIPSUrl, clusterUUID)

	data, err := SendHttpRequest(createVIPSUrlUpdated, postMethod, body, bearerToken)
	if err != nil {
		return nil, err
	}
	return data, err
}

func DeleteVIP(clusterUUID string, vipId int32) error {
	deleteVIPByIdUrlUpdated := fmt.Sprintf(deleteVIPByIdUrl, clusterUUID, strconv.Itoa(int(vipId)))
	log.Infof("deleteVIPByIdUrlUpdated: %s", deleteVIPByIdUrlUpdated)
	deleteVIPRequest := `{}`

	var deleteRequestBuffer bytes.Buffer
	deleteRequestBuffer.WriteString(deleteVIPRequest)
	_, err := SendHttpRequest(deleteVIPByIdUrlUpdated, deleteMethod, &deleteRequestBuffer, bearerToken)

	return err
}

func WaitForVIPState(clusterUUID string, vipId int32, targetState string) (bool, error) {
	log.Infof("Waiting for VIP %d to get to state [%s]", vipId, targetState)

	getVIPByIDurlUpdated := fmt.Sprintf(getVIPByIDurl, clusterUUID, strconv.Itoa(int(vipId)))

	maxWaitTimeInMin := 60
	for i := 0; i < maxWaitTimeInMin; i++ {
		data, err := SendHttpRequest(getVIPByIDurlUpdated, getMethod, nil, bearerToken)

		if err != nil {
			return false, err
		}

		var vipResponse pb.GetVipResponse
		json.Unmarshal(data, &vipResponse)

		if vipResponse.Vipstate == targetState {
			return true, nil
		}

		log.Infof("Iteration %d, Sleeping for 60 sec to allow cluster VIP state to update", i)
		time.Sleep(60 * time.Second)
	}

	return false, errors.New(fmt.Sprintf("Timed out waiting for VIP state to reach %s", targetState))
}

func ReportGeneration(report Report) {
	f, _ := os.Create("report.html")
	for index, specReport := range report.SpecReports {
		var tmpl = `<tr style="color: green"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`

		if index == 0 {
			fmt.Fprintf(f, `<html><head><style>table, th, td {border: 1px solid black;}</style></head><body>
			<u><b><p>Test suite Details   : `+strings.Join(strings.Split(report.SuiteDescription, ""), "")+`</p></b></u>
			<p>Execution-Start-Time : `+report.StartTime.UTC().String()+`</p>
			<p>Execution-End-Time   : `+report.EndTime.UTC().String()+`</p>
			<p>Total-Execution-Time : `+report.RunTime.String()+`</p>
			<table><tr><th>Sl.no</th><th>Testcase-Name</th><th>Execution-Status</th><th>Failure-Reason</th></tr>`)
		}
		tc_name := strings.Join(specReport.ContainerHierarchyTexts, "-")
		tc_status := specReport.State.String()
		var failure_reason string

		if tc_status == "failed" {
			tmpl = `<tr style="color: red"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failure_reason = specReport.FailureMessage()
		} else if tc_status == "skipped" {
			tmpl = `<tr style="color: orange"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failure_reason = "NA"
		} else {
			failure_reason = "NA"
		}

		if tc_name != "" {
			output := fmt.Sprintf(tmpl, strconv.Itoa(index), tc_name, tc_status, failure_reason)
			fmt.Fprintf(f, output)
		}

	}
	fmt.Fprintf(f, `</table></body></html>`)
	f.Close()
}
