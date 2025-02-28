package compute_utils

import (
	"fmt"
	"os"
	"path/filepath"
)

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