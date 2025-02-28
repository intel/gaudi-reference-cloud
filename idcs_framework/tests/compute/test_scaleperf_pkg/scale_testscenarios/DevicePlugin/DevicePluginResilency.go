package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

// Function to connect to the node and reboot it
func rebootNode(sshHost, sshUser, sshPassword string) error {
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	client, err := ssh.Dial("tcp", sshHost+":22", config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	if err := session.Run("sudo reboot &"); err != nil {
		return fmt.Errorf("failed to run reboot command: %w", err)
	}

	fmt.Println("Reboot command executed successfully on the node.")
	return nil
}

func fetchDevicePluginLogs(podName, harvesterKubeconfigPath string) {
	outputFile, err := os.OpenFile("output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer outputFile.Close()

	cmd := exec.Command("kubectl", "logs", podName, "-n", "kube-system", fmt.Sprintf("--kubeconfig=%s", harvesterKubeconfigPath), "--since=1h")
	cmd.Stdout = outputFile
	cmd.Stderr = outputFile

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		return
	}

	// Write a line of dashes after the logs
	if _, err := outputFile.WriteString("------------------\n"); err != nil {
		fmt.Printf("Error writing separator line: %v\n", err)
		return
	}
}

func main() {

	sshHostPointer := flag.String("sshHost", "<<default-host>>", "Harvester Host IP")
	sshUserPointer := flag.String("sshUser", "<<default-user>>", "Username")
	sshPasswordPointer := flag.String("sshPassword", "<<default-pwd>>", "password")
	podNamePointer := flag.String("podName", "<<default-pod-name>>", "Device plugin pod name")
	harvesterKubeconfigPointer := flag.String("harvesterKubeconfigPath", "<<default-path>>", "Harvester kubeconfig path")

	flag.Parse()

	restartInterval := 10   // restart interval in minutes
	totalExecutionTime := 1 // total execution time in hours
	sshHost := *sshHostPointer
	sshUser := *sshUserPointer
	sshPassword := *sshPasswordPointer
	podName := *podNamePointer
	harvesterKubeconfigPath := *harvesterKubeconfigPointer

	var executionStartTime = time.Now().UnixMilli()

	for (time.Now().UnixMilli() - executionStartTime) < (int64(totalExecutionTime) * 60 * 60 * 1000) {
		// reboot the node
		if err := rebootNode(sshHost, sshUser, sshPassword); err != nil {
			log.Fatalf("Error: %v", err)
		}

		// wait for the node to be up
		time.Sleep(time.Duration(restartInterval) * time.Minute)

		// fetch the logs
		fetchDevicePluginLogs(podName, harvesterKubeconfigPath)
	}
}
