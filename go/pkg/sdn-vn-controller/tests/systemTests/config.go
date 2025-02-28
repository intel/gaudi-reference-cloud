// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package systemTests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Chassis struct {
	ChassisID  string `json:"chassis_id"`
	IP         string `json:"ip"`
	SSHUser    string `json:"ssh_user"`
	SSHKeyPath string `json:"ssh_key_path"`
}

type NetworkConfig struct {
	SdnController string     `json:"sdnController"`
	CPDeploy      []CPDeploy `json:"cp_deploy"`
	Chassis       []Chassis  `json:"chassis"`
	VMs           []VM       `json:"vms"`
	ChassisMap    map[string]Chassis
}

type CPDeploy struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	File        string `json:"file"`
	Profile     string `json:"profile"`
	OvnCentalIP string `json:"ovnCentalIP"`
}

var config *NetworkConfig
var client v1.OvnnetClient

// loadConfig reads the configuration file, parses it into NetworkConfig, and sets up the gRPC client.
func loadConfig(filename string) (*NetworkConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %v", filename, err)
	}

	// Temporary struct to parse JSON with chassis as an array
	var rawConfig struct {
		SdnController string     `json:"sdnController"`
		CPDeploy      []CPDeploy `json:"cp_deploy"`
		Chassis       []Chassis  `json:"chassis"` // Temporarily holds chassis as an array
		VMs           []VM       `json:"vms"`
	}

	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Convert the chassis array to a map
	chassisMap := make(map[string]Chassis)
	for _, chassis := range rawConfig.Chassis {
		chassisMap[chassis.ChassisID] = chassis
	}

	// Populate NetworkConfig
	config = &NetworkConfig{
		SdnController: rawConfig.SdnController,
		CPDeploy:      rawConfig.CPDeploy,
		Chassis:       rawConfig.Chassis,
		ChassisMap:    chassisMap,
		VMs:           rawConfig.VMs,
	}

	if err := deploy(); err != nil {
		return nil, fmt.Errorf("failed to deploy: %v", err)
	}

	// Initialize namespaces for each VM
	for i := range config.VMs {
		if err := config.VMs[i].createNamespace(); err != nil {
			return nil, fmt.Errorf("failed to initialize VM %s: %v", config.VMs[i].Name, err)
		}
	}

	// Set up the gRPC client to connect to the SDN Controller
	addr := config.SdnController + ":50051" // Assume the gRPC service is on port 50051
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to SDN Controller at %s: %v", addr, err)
		return nil, err
	}

	client = v1.NewOvnnetClient(conn)
	return config, nil
}

// cleanupConfig deletes namespaces for each VM in the configuration.
func cleanupConfig() error {
	for i := range config.VMs {
		if err := config.VMs[i].deleteNamespace(); err != nil {
			return fmt.Errorf("failed to clean up VM %s: %v", config.VMs[i].Name, err)
		}
	}
	if err := undeploy(); err != nil {
		return fmt.Errorf("failed to undeploy: %v", err)
	}
	return nil
}

func deploy() error {
	dockerCmd, err := getDockerCommand()
	if err != nil {
		return fmt.Errorf("failed to get Docker command: %v", err)
	}

	for _, cp := range config.CPDeploy {
		if cp.Type == "docker-compose" {
			// Build
			if err := runDockerComposeCommand(dockerCmd, &cp, "build"); err != nil {
				return fmt.Errorf("failed to build with Docker Compose for %s: %v", cp.Name, err)
			}

			// Up (detached)
			if err := runDockerComposeCommand(dockerCmd, &cp, "up", "--detach"); err != nil {
				return fmt.Errorf("failed to start with Docker Compose for %s: %v", cp.Name, err)
			}
		}
	}

	return nil
}

func undeploy() error {
	dockerCmd, err := getDockerCommand()
	if err != nil {
		return fmt.Errorf("failed to get Docker command: %v", err)
	}

	for _, cp := range config.CPDeploy {
		if cp.Type == "docker-compose" {
			// Down
			if err := runDockerComposeCommand(dockerCmd, &cp, "down"); err != nil {
				return fmt.Errorf("failed to stop with Docker Compose for %s: %v", cp.Name, err)
			}
		}
	}

	return nil
}

func getDockerCommand() (string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("failed to open /etc/os-release: %v", err)
	}
	defer file.Close()

	var versionID string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.TrimPrefix(line, "VERSION_ID=")
			versionID = strings.Trim(versionID, `"`)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading /etc/os-release: %v", err)
	}

	if versionID == "24.04" {
		return "docker compose", nil
	}
	return "docker-compose", nil
}

func runDockerComposeCommand(dockerCmd string, cp *CPDeploy, args ...string) error {
	profileArg := ""
	if cp.Profile != "" && cp.Profile != "none" {
		profileArg = "--profile " + cp.Profile
	}

	cmdStr := fmt.Sprintf("%s -f %s %s %s", dockerCmd, cp.File, profileArg, strings.Join(args, " "))

	if cp.OvnCentalIP != "" && cp.OvnCentalIP != "none" {
		cmdStr = fmt.Sprintf("OVN_IP=%s ", cp.OvnCentalIP) + cmdStr
	}
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Running: %s\n", cmdStr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %q: %v", cmdStr, err)
	}
	return nil
}
