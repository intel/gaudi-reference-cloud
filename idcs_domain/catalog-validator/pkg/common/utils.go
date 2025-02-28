package common

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

//IsYAMLFile :
func isYAMLFile(filepath string) bool {
	isYaml := false
	r, err := regexp.MatchString(".yaml", filepath)
	if err == nil && r {
		isYaml = true
	}
	r, err = regexp.MatchString(".yml", filepath)
	if err == nil && r {
		isYaml = true
	}
	return isYaml
}

func GetYamlContent(filePath string) []byte {

	if isYAMLFile(filePath) {
		if filebuf, err := ioutil.ReadFile(filePath); err == nil {
			fileAsString := string(filebuf[:])
			if fileAsString == "\n" || fileAsString == "" {
				fmt.Println("Empty file: ", filePath)
				return nil
			}
			return filebuf
		} else {
			fmt.Printf("error parsing yaml file: %s, Error: %v", filePath, err)
			return nil
		}

	} else {
		fmt.Println("Invalid YAML file: ", filePath)
		return nil
	}

}