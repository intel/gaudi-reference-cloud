package testutils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateRandomMAC() string {
	return fmt.Sprintf("b0:fd:0b:%02x:%02x:%02x",
		rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func GetCurrentFolder(filePath string) (string, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(dir)
	return base, nil
}

// "eth1" -> "Ethernet1"
func ConvertShortPortNameToLongFormat(portName string) (string, error) {
	after, found := strings.CutPrefix(portName, "eth")
	if !found {
		return "", fmt.Errorf("invalid port name")
	}
	return fmt.Sprintf("Ethernet%s", after), nil
}

// convert the sw short name to fqdn, frontend-leaf1 -> clab-allscfabrics-frontend-leaf1
func ConvertSWShortNameToFQDN(nodeName string, topologyName string) string {
	return fmt.Sprintf("clab-%s-%s", topologyName, nodeName)
}

func DeleteClabTmpFolder(rootDir string) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current path is a directory and its name starts with "clab-"
		if info.IsDir() && strings.HasPrefix(info.Name(), "clab-") {
			err := os.RemoveAll(path)
			if err != nil {
				fmt.Printf("Failed to delete folder %s: %v\n", path, err)
				return err
			}
			fmt.Printf("Folder %s deleted successfully\n", path)
		}
		return nil
	})

	return err
}

func RunCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func RunCommandStdOut(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.Output()
	return string(output), err
}

func GetPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return pwd
}

func CheckPortOpenWithTimeout(ip string, port int, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", ip, port)

	resCh := make(chan struct{})
	go func() {
		for {
			conn, err := net.DialTimeout("tcp", address, 10*time.Second)
			if err == nil {
				conn.Close()
				resCh <- struct{}{} // Port is open and reachable
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	select {
	case <-time.After(30 * time.Second):
		return fmt.Errorf("DialTimeout for %v", address)
	case _, ok := <-resCh:
		if !ok {
			return fmt.Errorf("result channel is closed")
		}
		fmt.Printf("[%v] is ready\n", address)
		break
	}
	return nil

}

func Timeout(timeoutDuration time.Duration, fn func(ctx context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	resultCh := make(chan error, 1)

	go func() {
		resultCh <- fn(timeoutCtx)
	}()

	select {
	case res := <-resultCh:
		return res
	case <-timeoutCtx.Done():
		return errors.New("function execution timed out")
	}
}

func CompareConfigs(config1, config2 string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(config1, config2, false)
	// Convert the diffs to git diff format
	var result strings.Builder

	for i, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			result.WriteString(fmt.Sprintf("\n+ %s", diff.Text))
		case diffmatchpatch.DiffDelete:
			result.WriteString(fmt.Sprintf("\n- %s", diff.Text))
		case diffmatchpatch.DiffEqual:
			// handling the prefix "\n"
			text := diff.Text
			if i-1 >= 0 && (diffs[i-1].Type == diffmatchpatch.DiffInsert || diffs[i-1].Type == diffmatchpatch.DiffDelete) && strings.HasPrefix(text, "\n") {
				text = strings.TrimPrefix(text, "\n")
			}
			result.WriteString(fmt.Sprintf("\n%s", text))
			//result.WriteString(fmt.Sprintf("\n  %s", diff.Text))
		}
	}
	all := result.String()
	allLines := strings.Split(all, "\n")
	n := len(allLines)
	selectedLines := make([]string, 0)
	keepContextLines := 2
	lastSelectedLine := -1
	for i := 0; i < n; i++ {
		currentLine := allLines[i]
		if strings.HasPrefix(currentLine, "+") || strings.HasPrefix(currentLine, "-") {
			selectedLines = append(selectedLines, currentLine)
			lastSelectedLine = i
		} else {
			shouldAddLine := false
			for j := 1; j <= keepContextLines; j++ {
				if (i-j >= 0 && (strings.HasPrefix(allLines[i-j], "+") || strings.HasPrefix(allLines[i-j], "-"))) || (i+j < n && (strings.HasPrefix(allLines[i+j], "+") || strings.HasPrefix(allLines[i+j], "-"))) {
					shouldAddLine = true
					break
				}

			}
			if shouldAddLine {
				if i-lastSelectedLine > 1 && lastSelectedLine != -1 {
					selectedLines = append(selectedLines, "---")
				}

				selectedLines = append(selectedLines, currentLine)
				lastSelectedLine = i
			}

		}
	}
	return strings.Join(selectedLines, "\n")
}

func WriteStringToFile(content, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not open or create file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("could not write to file: %v", err)
	}

	return nil
}

func Diff(str1, str2 string, surroundingLines int) (string, error) {
	// create temporary file for the input strings
	file1, err := os.CreateTemp("", "str1-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for str1: %s", err.Error())
	}
	defer os.Remove(file1.Name())

	file2, err := os.CreateTemp("", "str2-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for str2: %s", err.Error())
	}
	defer os.Remove(file2.Name())

	_, err = file1.WriteString(str1)
	if err != nil {
		return "", fmt.Errorf("failed to write to file1: %s", err.Error())
	}

	_, err = file2.WriteString(str2)
	if err != nil {
		return "", fmt.Errorf("failed to write to file1: %s", err.Error())
	}

	file1.Close()
	file2.Close()

	output, err := RunCommandStdOut(fmt.Sprintf("diff -U %d %s %s", surroundingLines, file1.Name(), file2.Name()))
	// "exit status 1" means files are different
	if err != nil && err.Error() != "exit status 1" {
		return "", fmt.Errorf("failed to run diff command: %s", err.Error())
	}

	// Remove the initial lines that contain the filenames, eg:
	//--- /tmp/str1-78177558.txt	2024-09-20 10:46:55.651953240 +0000
	//+++ /tmp/str2-1965453811.txt	2024-09-20 10:46:55.651953240 +0000
	initialLinesRegex := regexp.MustCompile("(?m)^(---|\\+\\+\\+) .*/str[12]-.*$[\n]+")
	output = initialLinesRegex.ReplaceAllString(output, "")

	// Remove the lines that contain the line-number, so we can change the switch-config in future, without having to change all the diffs. eg.
	// @@ -1,3 +1,3 @@
	lineNumberRegex := regexp.MustCompile("(?m)^@@.*@@$")
	output = lineNumberRegex.ReplaceAllString(output, "@@ @@")

	return output, nil
}

func ExecCommandInContainer(containerName string, command []string) (string, error) {
	args := append([]string{"exec", containerName}, command...)
	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	return string(output), err
}

// GetOutboundIP Gets the preferred outbound ip of this machine - easier than trying to figure out which IP assigned to various interfaces to use.
// From https://stackoverflow.com/a/37382208
func GetOutboundIP() (error, net.IP) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return err, nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return nil, localAddr.IP
}

func GetMetrics(metricNames []string) (error, map[string]int) {
	retMetrics := make(map[string]int)

	// Run a pod that connects in to the SDN-controller's /metrics endpoint, through the kube-rbac-proxy.
	// Must be run inside the cluster because it needs a serviceAccount token to authenticate through kube-rbac-proxy, and because routing into kind & service is tricky from outside.
	output, err := RunCommand("kubectl delete --ignore-not-found -f sdnMetricsCheckerJob.yaml")
	if err != nil {
		return fmt.Errorf("error deleting old sdnMetricsCheckerJob. Output: %s. Error: %v", output, err), nil
	}

	output, err = RunCommand("kubectl apply -f sdnMetricsCheckerJob.yaml")
	if err != nil {
		return fmt.Errorf("error running sdnMetricsCheckerJob. Output: %s. Error: %v", output, err), nil
	}

	output, err = RunCommand("kubectl wait --for=condition=complete --timeout=100s -n idcs-system job/sdn-e2e-metrics-checker")
	if err != nil {
		return fmt.Errorf("error waiting for sdnMetricsCheckerJob to complete. Output: %s. Error: %v", output, err), nil
	}

	metricsOutput, err := RunCommand("kubectl logs -n idcs-system -l job-name=\"sdn-e2e-metrics-checker\" --tail=-1")
	if err != nil {
		return fmt.Errorf("error getting logs from sdn-e2e-metrics-checker. Output: %s. Error: %v", metricsOutput, err), nil
	}

	for _, metricName := range metricNames {
		vlanMetricRegex := regexp.MustCompile(fmt.Sprintf("(?m)^%s ([0-9]*)$", metricName))
		metricFromRegex := vlanMetricRegex.FindStringSubmatch(metricsOutput)
		if metricFromRegex == nil {
			return fmt.Errorf("metric %s not found", metricName), nil
		}

		namedMetric, err := strconv.Atoi(metricFromRegex[1])
		if err != nil {
			return fmt.Errorf("failed to parse metric to integer: %s", metricFromRegex[1]), nil
		}

		retMetrics[metricName] = namedMetric
	}

	return nil, retMetrics
}
