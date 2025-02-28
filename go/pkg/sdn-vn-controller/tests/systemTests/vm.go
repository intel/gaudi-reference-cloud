// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package systemTests

import (
	"fmt"
	"os/exec"
	"strings"
)

type VM struct {
	Name      string `json:"name"`
	Chassis   string `json:"chassis"`
	Interface string `json:"interface"` // Network interface
	DeviceID  int    `json:"device_id"`
	MAC       string `json:"MAC"`
	IPaddress string
	DefaultGW string
}

// executeCommand runs a command locally or remotely over SSH based on the VM's Server setting.
func (vm *VM) executeCommand(command string) error {
	server := config.ChassisMap[vm.Chassis]
	if server.IP == "localhost" {
		// Run locally
		return exec.Command("bash", "-c", command).Run()
	}
	// Escape the command for remote SSH execution
	escapedCommand := strings.ReplaceAll(command, "'", `'"'"'`)
	sshCommand := fmt.Sprintf(
		"ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s %s@%s '%s'",
		server.SSHKeyPath, server.SSHUser, server.IP, escapedCommand,
	)
	return exec.Command("bash", "-c", sshCommand).Run()
}

// executeCommandWithOutput runs a command and returns the output as a string.
func (vm *VM) executeCommandWithOutput(command string) (string, error) {
	server := config.ChassisMap[vm.Chassis]
	if server.IP == "localhost" {
		output, err := exec.Command("bash", "-c", command).CombinedOutput()
		return string(output), err
	}
	sshCommand := fmt.Sprintf("ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s %s@%s '%s'", server.SSHKeyPath, server.SSHUser, server.IP, command)
	output, err := exec.Command("bash", "-c", sshCommand).CombinedOutput()
	return string(output), err
}

// createNamespace creates a network namespace for the VM and configures the interface.
func (vm *VM) createNamespace() error {
	// Create the network namespace
	if err := vm.executeCommand(fmt.Sprintf("sudo ip netns add %s", vm.Name)); err != nil {
		return fmt.Errorf("failed to create namespace %s: %v", vm.Name, err)
	}

	// Move the interface to the namespace
	if err := vm.executeCommand(fmt.Sprintf("sudo ip link set %s netns %s", vm.Interface, vm.Name)); err != nil {
		return fmt.Errorf("failed to set interface %s for namespace %s: %v", vm.Interface, vm.Name, err)
	}

	// Bring up the interface within the namespace
	if err := vm.executeCommand(fmt.Sprintf("sudo ip netns exec %s ip link set %s up", vm.Name, vm.Interface)); err != nil {
		return fmt.Errorf("failed to bring up interface %s for namespace %s: %v", vm.Interface, vm.Name, err)
	}

	// Disable rx and tx offload on the interface because The UDP checksum is not directly calculated by Open vSwitch (OVS) and
	// should be calculated in software instead of offloading it to the NIC which in this case is a veth pair.
	server := config.ChassisMap[vm.Chassis]
	if server.IP == "localhost" {
		if err := exec.Command("sudo", "ip", "netns", "exec", vm.Name, "ethtool", "--offload", vm.Interface, "rx", "off", "tx", "off").Run(); err != nil {
			return fmt.Errorf("failed to disable offload on interface %s for namespace %s: %v", vm.Interface, vm.Name, err)
		}
	}

	return nil
}

// deleteNamespace removes the interface from the namespace and then deletes the namespace.
func (vm *VM) deleteNamespace() error {
	// Move the interface out of the namespace
	if err := vm.executeCommand(fmt.Sprintf("sudo ip netns exec %s ip link set %s netns 1", vm.Name, vm.Interface)); err != nil {
		return fmt.Errorf("failed to move interface %s out of namespace %s: %v", vm.Interface, vm.Name, err)
	}

	// Delete the network namespace
	if err := vm.executeCommand(fmt.Sprintf("sudo ip netns del %s", vm.Name)); err != nil {
		return fmt.Errorf("failed to delete namespace %s: %v", vm.Name, err)
	}

	return nil
}

// assignIP assigns an IP address to the VM's interface.
func (vm *VM) assignIP(ip string, prefix string) error {
	if vm.IPaddress == ip {
		return nil
	}
	command := fmt.Sprintf("sudo ip netns exec %s ip addr add %s dev %s", vm.Name, ip+prefix, vm.Interface)
	if err := vm.executeCommand(command); err != nil {
		return fmt.Errorf("failed to assign IP %s to interface %s in namespace %s: %v", ip, vm.Interface, vm.Name, err)
	}
	vm.IPaddress = ip
	return nil
}

// addDefaultGateway sets the default gateway in the VM's namespace.
func (vm *VM) addDefaultGateway(gateway string) error {
	if vm.DefaultGW == gateway {
		return nil
	}
	command := fmt.Sprintf("sudo ip netns exec %s ip route add default via %s dev %s", vm.Name, gateway, vm.Interface)
	if err := vm.executeCommand(command); err != nil {
		return fmt.Errorf("failed set default gw %s to interface %s in namespace %s: %v", gateway, vm.Interface, vm.Name, err)
	}
	vm.DefaultGW = gateway
	return nil
}

// ping sends a ping to the specified IP address from the VM's namespace.
func (vm *VM) ping(ip string) (string, error) {
	timeout := 5
	command := fmt.Sprintf("sudo ip netns exec %s ping -c 4 -i 0.1 -w %d %s", vm.Name, timeout, ip)
	return vm.executeCommandWithOutput(command)
}

// testPing tests the connectivity from the VM to the specified IP address.
func (vm *VM) testPing(targetIP string) error {
	fmt.Printf("Testing ping from %s to %s\n", vm.Name, targetIP)
	output, err := vm.ping(targetIP)
	if err != nil {
		return fmt.Errorf("ping from %s to IP %s failed: %v", vm.Name, targetIP, err)
	}

	if !containsExpectedResponse(output, "4 received") {
		return fmt.Errorf("ping from %s to IP %s failed: unexpected response: %s", vm.Name, targetIP, output)
	}

	return nil
}

// testPingVM tests the connectivity from the VM to the target VM using its IPaddress.
func (vm *VM) testPingVM(target *VM) error {
	fmt.Printf("Testing ping from %s to %s\n", vm.Name, target.Name)
	if target.IPaddress == "" {
		return fmt.Errorf("target VM %s has no IPaddress set", target.Name)
	}
	return vm.testPing(target.IPaddress)
}

func (vm *VM) startServer(port int, protocol string) error {
	if protocol == "tcp" {
		command := fmt.Sprintf(
			"sudo ip netns exec %s nohup bash -c 'while true; do echo \"Hello from %s\" | socat - TCP-LISTEN:%d,reuseaddr; done' >/dev/null 2>&1 &",
			vm.Name, vm.Name, port,
		)
		return vm.executeCommand(command)
	} else if protocol == "udp" {
		command := fmt.Sprintf(
			"sudo ip netns exec %s nohup bash -c 'socat -v -T 5 UDP-RECVFROM:%d,fork EXEC:\"echo \\\"Hello from %s\\\"\"' >/dev/null 2>&1 &",
			vm.Name, port, vm.Name,
		)
		return vm.executeCommand(command)
	}
	return fmt.Errorf("unsupported protocol: %s", protocol)
}

func (vm *VM) testConnection(target *VM, port int, protocol string) error {
	if target.IPaddress == "" {
		return fmt.Errorf("target VM %s has no IP address set", target.Name)
	}

	expectedResponse := fmt.Sprintf("Hello from %s", target.Name)
	fmt.Printf("Testing %s connection from %s to %s \n", protocol, vm.Name, target.Name)

	var command string
	if protocol == "tcp" {
		command = fmt.Sprintf("sudo ip netns exec %s bash -c \"echo 'Hello' | socat - TCP:%s:%d,connect-timeout=5\"", vm.Name, target.IPaddress, port)
	} else if protocol == "udp" {
		command = fmt.Sprintf(
			"sudo ip netns exec %s bash -c \"echo 'Hello' | socat - UDP:%s:%d\"",
			vm.Name, target.IPaddress, port,
		)
	} else {
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	output, err := vm.executeCommandWithOutput(command)
	if err != nil {
		return fmt.Errorf("%s connection test from %s to %s:%d failed: %v", strings.ToUpper(protocol), vm.Name, target.IPaddress, port, err)
	}

	if !strings.Contains(strings.TrimSpace(output), expectedResponse) {
		return fmt.Errorf("unexpected response from %s: got %q, want %q", target.Name, output, expectedResponse)
	}

	return nil
}

// stopServers stops any iperf server processes in the VM's namespace.
func (vm *VM) stopServers() error {
	command := fmt.Sprintf("sudo ip netns exec %s pkill -f 'socat.*LISTEN'", vm.Name)
	return vm.executeCommand(command)
}

// containsExpectedResponse checks if the command output contains the expected response.
func containsExpectedResponse(output string, expected string) bool {
	if !strings.Contains(output, expected) {
		return false
	}

	return true
}
