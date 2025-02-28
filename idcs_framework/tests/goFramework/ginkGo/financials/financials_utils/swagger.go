package financials_utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

type Swagger struct {
	Paths map[string]map[string]interface{} `json:"paths"`
}

func ParseSwaggerFile(filePath string) (*Swagger, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var swagger Swagger
	err = json.Unmarshal(bytes, &swagger)
	if err != nil {
		return nil, err
	}

	return &swagger, nil
}

func replaceDefaultParams(endpoint string) string {
	newValue := "testparams"

	// Compile the regex pattern
	re := regexp.MustCompile(`\{[^}]+\}`)

	// Replace the matched patterns with the new value
	newEndpoint := re.ReplaceAllString(endpoint, newValue)

	fmt.Println("New URL", newEndpoint)

	return newEndpoint
}

func ExtractSpecificEndpoints(swagger *Swagger, urls []string) []string {
	var endpoints []string
	for _, url := range urls {
		for path, methods := range swagger.Paths {
			if path == url {
				for method := range methods {
					path = replaceDefaultParams(path)
					endpoints = append(endpoints, fmt.Sprintf("%s %s", method, path))
				}
			}
		}
	}
	return endpoints
}

// extractEndpointsByOperationID extracts endpoints and methods filtered by operationId
func ExtractEndpointsByOperationID(swagger *Swagger, actions []string) map[string]string {
	endpoints := make(map[string]string)
	for path, methods := range swagger.Paths {
		for method, operation := range methods {
			operation := operation.(map[string]interface{})
			for _, action := range actions {
				if action == operation["operationId"] {
					endpoints[fmt.Sprintf("%s", path)] = fmt.Sprintf("%s", method)
				}
			}
		}
	}
	return endpoints
}
