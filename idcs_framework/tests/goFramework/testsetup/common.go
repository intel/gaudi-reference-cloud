package testsetup

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"github.com/melbahja/goph"
	"goFramework/framework/common/logger"
	"os"
)

func WriteTestDatatoJson() {
	// open a file for writing
	logger.Log.Info("Writing Created Map to testdata.json")
	
	file, err := os.Create("../testdata.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// create a JSON encoder
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// write the data to the file using the encoder
	err = encoder.Encode(Testdata)
	if err != nil {
		fmt.Println(err)
		return
	}
	logger.Log.Info("Map data written to testdata.json")
	
}

func RemoteCommandExecution(ip string, username string, keyPath string, command string) error {
	// Start new ssh connection with private key.
	auth, err := goph.Key(keyPath, "")
	if err != nil {
		return fmt.Errorf("\nerror : %s", err.Error())
	}

	client, err := goph.New(username, ip, auth)
	if err != nil {
		return fmt.Errorf("\nerror : %s", err.Error())
	}

	// Defer closing the network connection.
	defer client.Close()

	// Execute your command.
	out, err := client.Run(command)

	if err != nil {
		return fmt.Errorf("\nerror %s", err.Error())
	}

	// Get your output as []byte.
	fmt.Println(string(out))
	return nil
}

func RemoveDuplicateValues(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
