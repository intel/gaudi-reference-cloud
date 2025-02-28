package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
)

func setupSSHAgent() (string, string, error) {
	cmd := exec.Command("ssh-agent", "-s")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error starting ssh-agent: %v\n%s", err, output)
	}
	reSock := regexp.MustCompile(`SSH_AUTH_SOCK=(.*?);`)
	rePID := regexp.MustCompile(`SSH_AGENT_PID=(.*?);`)
	matchesSock := reSock.FindStringSubmatch(string(output))
	matchesPID := rePID.FindStringSubmatch(string(output))
	if len(matchesSock) < 2 || len(matchesPID) < 2 {
		return "", "", fmt.Errorf("unexpected output from ssh-agent: %s", output)
	}
	sshAuthSock := matchesSock[1]
	sshAgentPID := matchesPID[1]
	return sshAuthSock, sshAgentPID, nil
}

func SSHIntoInstanceViaJumpHost(ctx context.Context, proxyIp, proxyGuestUser, instanceIp, instanceUser, instancePrivateKeyPath string) (string, error) {
	logger.Logf.Info("SSHIntoInstanceViaJumpHost")
	sshAuthSock, sshAgentPID, err := setupSSHAgent()
	if err != nil {
		return "", fmt.Errorf("error setting up ssh-agent: %v", err)
	}
	// Set environment variables for ssh-agent
	os.Setenv("SSH_AUTH_SOCK", sshAuthSock)
	os.Setenv("SSH_AGENT_PID", sshAgentPID)
	os.Setenv("BAZEL_SSH_OPTIONS", "-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o HostKeyAlgorithms=ecdsa-sha2-nistp256,ed25519")

	cmdAddKey := exec.Command("ssh-add", instancePrivateKeyPath)
	output, err := cmdAddKey.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error adding key to ssh-agent: %v %s", err, output)
	}

	sshCommand := fmt.Sprintf(
		`ssh -i %s %s -J %s@%s %s@%s 'ls -l' <<{{EOF}}
		yes
		{{EOF}}`,
		instancePrivateKeyPath,
		os.Getenv("BAZEL_SSH_OPTIONS"),
		proxyGuestUser,
		proxyIp,
		instanceUser,
		instanceIp,
	)
	yescmd := exec.Command("bash", "-c", sshCommand)
	yescmd.Env = os.Environ()

	// Connect the command's stdout and stderr to the current process's stdout and stderr
	var stdoutBuf bytes.Buffer
	yescmd.Stdout = &stdoutBuf
	yescmd.Stderr = os.Stderr

	// Start and wait for the command to finish
	logger.Logf.Info("running ssh command")
	err = yescmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running ssh command: %v", err)
	}
	stdoutStr := stdoutBuf.String()
	return stdoutStr, nil
}
