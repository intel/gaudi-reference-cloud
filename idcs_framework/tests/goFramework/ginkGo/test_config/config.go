/**
 * Configurations used by api_gateway.
 *
 **/

package test_config

import (
	"os"

)

// Load the configuration from the provided yaml file.
func LoadConfig(filePath string) (string, error) {
	configData, err := os.ReadFile(filePath)	// if we os.Open returns an error then handle it
	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	return string(configData), nil
}
