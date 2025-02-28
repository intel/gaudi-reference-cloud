package utils

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func GetRandomString() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 6+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : 6+2]
}

func GetRandomStringWithLimit(n int) string {
	var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, n)
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

func GetUnixTime(timestamp string) int64 {
	layout := time.RFC3339
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		logInstance.Println(err)
	}

	return t.UnixMilli()
}

func ValidateTimeStamp(i, min, max int64) bool {
	if (i >= min) && (i <= max) {
		return true
	} else {
		return false
	}
}

func EnrichInventoryData(rawData string, proxyIp string, proxyUser string, machineIp string, privateKeyPath string) string {
	enrichedData := rawData
	enrichedData = strings.ReplaceAll(enrichedData, "<<proxy_machine_ip>>", proxyIp)
	enrichedData = strings.ReplaceAll(enrichedData, "<<proxy_machine_user>>", proxyUser)
	enrichedData = strings.ReplaceAll(enrichedData, "<<machine_ip>>", machineIp)
	enrichedData = strings.ReplaceAll(enrichedData, "<<authorised_key_path>>", privateKeyPath)
	return enrichedData
}

func RunCommand(sshCommand []string) (*bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(sshCommand[0], sshCommand[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Command failed with error: %v\n", err)
		if stderr.Len() > 0 {
			fmt.Printf("Error: %s\n", stderr.String())
		}
		return nil, err
	}
	return &stdout, nil
}

func ReadFileFromSSHEndpoint(proxyUser, proxyIp, user, ip, path string) (string, error) {
	var buf bytes.Buffer

	sshCommand := []string{"ssh", "-J", proxyUser + "@" + proxyIp, "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", user + "@" + ip, "cat", path}
	cmd := exec.Command(sshCommand[0], sshCommand[1:]...)
	cmd.Stdout = &buf

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ConvertFileToString(filePath string, filename string) (string, error) {
	//fmt.Println("Config file path", filePath)
	wd, _ := os.Getwd()
	wd = filepath.Clean(filepath.Join(wd, filePath))
	configData, err := os.ReadFile(wd + "/" + filename)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(configData), nil
}

func WriteStringToFile(filePath string, filename string, content string) {

	f, err := os.Create(filePath + "/" + filename)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(content)
	if err2 != nil {
		fmt.Println(err2)
	}
}

func GetRESTConfig(customKubeconfigPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", customKubeconfigPath)
	if err != nil {
		return rest.InClusterConfig()
	}
	return config, nil
}

func GetAvailabilityZone(region string, enabled bool) string {
	if region[:7] == "us-dev-" {
		if enabled {
			return region + "b"
		} else {
			return region + "a"
		}
	} else {
		return region + "a"
	}
}

func GetkeyFilePath(filePath string, filename string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	keyPath := filepath.Join(currentUser.HomeDir, filePath, filename)

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", err
	}
	return keyPath, nil
}

func ReadFileAsString(authorizedKeysPath string) (string, error) {
	content, err := ioutil.ReadFile(authorizedKeysPath)
	if err != nil {
		return "", err
	}
	contentStr := strings.TrimSpace(string(content))
	return contentStr, nil
}

// ExpandPath expands the ~ to the user's home directory if present
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir := usr.HomeDir
		path = filepath.Join(homeDir, path[1:])
	}
	return path, nil
}

// ReadPublicKey reads the public key from a specified file path
func ReadPublicKey(filePath string) (string, error) {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		return "", err
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	keyBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	publicKeyValue := strings.TrimSpace(string(keyBytes))
	return publicKeyValue, nil
}

func IsArrayInIncreasingOrder(arr []float64) bool {
	if len(arr) == 1 {
		return true
	} else {
		for i := 1; i < len(arr); i++ {
			if arr[i] < arr[i-1] {
				return false
			}
		}
		return true
	}
}

func PickRandomItemFromList(list []string) string {
	// Seed the random number generator to ensure different results on each run
	rand.Seed(time.Now().UnixNano())

	randomIndex := rand.Intn(len(list))
	randomObject := list[randomIndex]

	return randomObject
}

func JsonToMap(json gjson.Result) map[string][]string {
	result := make(map[string][]string)

	// Iterate over each key-value pair in the GJSON result
	json.ForEach(func(key, value gjson.Result) bool {
		// Create a slice to hold the associated string values
		var names []string
		// Iterate over the array of names
		value.ForEach(func(_, v gjson.Result) bool {
			names = append(names, v.String())
			return true
		})
		result[key.String()] = names // Assign the names slice to the key
		return true
	})

	return result
}

func CopyFile(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, input, 0644)
	return err
}
