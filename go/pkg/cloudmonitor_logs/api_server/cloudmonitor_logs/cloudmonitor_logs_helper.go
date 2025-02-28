// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudmonitor_logs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"
)

func sendRequestToOpenSearch(useProxy bool, insecureSkipVerify bool, requestURL string, requestMethod string, query string) ([]byte, int, error) {
	// read OpenSearch user name from vault
	userNameFilePath := "/vault/secrets/opensearchusername"
	userName, err := os.ReadFile(userNameFilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("unable reading OpenSearch username from vault: %v", err)
	}
	openSearchUserName := string(userName)

	// read OpenSearch password from vault
	passwordFilePath := "/vault/secrets/opensearchpassword"
	password, err := os.ReadFile(passwordFilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("unable reading OpenSearch password from vault: %v", err)
	}
	openSearchPassword := string(password)

	client := &http.Client{}

	if useProxy && insecureSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		fmt.Println("Proxy and Insecure Connection is used!!")
	} else {
		// caCert, err := os.ReadFile("/vault/secrets/rootca")
		// if err != nil {
		// 	fmt.Println("Unable to read CA for OpenSearch")
		// 	return []byte(""), fmt.Errorf("unable to read CA for OpenSearch: %v", err)
		// }
		// caCertPool := x509.NewCertPool()
		// caCertPool.AppendCertsFromPEM(caCert)
		// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: caCertPool}
	}

	req, err := http.NewRequest(requestMethod, requestURL, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, 0, fmt.Errorf("error creating HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(openSearchUserName, openSearchPassword)

	// Send HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error sending HTTP request to OpenSearch server: %v", err)
	}
	defer resp.Body.Close()

	// check if response code is 200 for GET requests
	if requestMethod == "GET" && resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("received non-200 response for GET request: %d", resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response body from OpenSearch")
	}

	// for non HEAD request methods
	if requestMethod != "HEAD" && len(body) == 0 {
		return nil, 0, fmt.Errorf("no data received in response body")
	}

	return body, resp.StatusCode, nil
}

func ConvertEpochToUTC(epoch int64) string {

	// Convert epoch to time.Time
	t := time.Unix(0, epoch*int64(time.Millisecond)).UTC()

	// Format time.Time as ISO 8601 date string in UTC with miliseconds
	isoString := t.Format("2006-01-02T15:04:05.000Z")

	// fmt.Println("ISO String: " + isoString)
	return isoString
}

func validateTimestamp(startTimeEpoch string, endTimeEpoch string) (int64, int64) {
	// Convert string epoch to int64
	startTime, err := strconv.ParseInt(startTimeEpoch, 10, 64)
	if err != nil {
		fmt.Printf("Error converting string to int64: %v\n", err)
	}
	endTime, err := strconv.ParseInt(endTimeEpoch, 10, 64)
	if err != nil {
		fmt.Printf("Error converting string to int64: %v\n", err)
	}

	// Get the current time in epoch
	now := time.Now().UnixMilli()

	// Calculate the retention period time in epoch
	retentionPeriodDays := int64(7)
	retentionPeriod := now - retentionPeriodDays*24*60*60*1000

	// Adjust the startTime if it's older than logs retentionPeriod i.e 7 days
	if startTime < retentionPeriod {
		startTime = retentionPeriod
	}

	// Adjust the endTime if it's in the future
	if endTime > now {
		endTime = now
	}

	// Ensure the endTime is not before the startTime
	if endTime < startTime {
		endTime = now
	}

	return startTime, endTime
}

func BuildQuery(fieldNames []string, values []string, cloudAccountId string, clusterRegion string, clusterId string, size int32, from int32, startTime int64, endTime int64) (string, error) {

	var mustClauses []interface{}

	// Add multi-phrase match clauses
	for i := 0; i < len(fieldNames); i++ {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match_phrase": map[string]interface{}{
				fieldNames[i]: values[i],
			},
		})
	}

	// add cloudAccountId as match phrase
	mustClauses = append(mustClauses, map[string]interface{}{
		"term": map[string]interface{}{
			"cloudaccount_id": cloudAccountId,
		},
	})

	// add clusterRegion as match phrase
	mustClauses = append(mustClauses, map[string]interface{}{
		"term": map[string]interface{}{
			"cluster_region": clusterRegion,
		},
	})

	// add clusterId as match phrase
	mustClauses = append(mustClauses, map[string]interface{}{
		"term": map[string]interface{}{
			"cluster_id": clusterId,
		},
	})

	// Add range clause for timestamp
	mustClauses = append(mustClauses, map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]interface{}{
				"gte":    startTime,
				"lte":    endTime,
				"format": "epoch_millis",
			},
		},
	})

	// Build the final query Data Structure
	queryStruct := map[string]interface{}{
		"from": from,
		"size": size,
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
	}

	// Convert the data structure to a JSON encoded byte slice
	jsonData, err := json.Marshal(queryStruct)
	if err != nil {
		return "", fmt.Errorf("error marshaling queryStruct to JSON")
	}

	// JSON string
	query := string(jsonData)
	return query, nil
}

func validateFiledNames(allowedFilterFieldNames []string, reqFieldNames []string, reqFieldValues []string) ([]string, []string) {
	fieldNames := []string{}
	fieldValues := []string{}

	for i := 0; i < len(reqFieldNames); i++ {
		if slices.Contains(allowedFilterFieldNames, reqFieldNames[i]) {
			fieldNames = append(fieldNames, reqFieldNames[i])
			fieldValues = append(fieldValues, reqFieldValues[i])
		}
	}

	return fieldNames, fieldValues
}

func CreateTemplate(cloudAccountId string, openSearchURL string) (string, string, error) {
	templateName := fmt.Sprintf("cm_logs_%s_tmplt", cloudAccountId)
	url := fmt.Sprintf("%s/_index_template/%s", openSearchURL, templateName)

	templateBodyMap := map[string]interface{}{
		"index_patterns": []string{fmt.Sprintf("cm_logs_%s*", cloudAccountId)},
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"plugins.index_state_management.rollover_alias": fmt.Sprintf("cm_logs_%s_wrt_als", cloudAccountId),
				"index": map[string]interface{}{
					"number_of_shards":   "1",
					"number_of_replicas": "1",
					"refresh_interval":   "30s",
					"sort.field":         "@timestamp",
					"sort.order":         "desc",
				},
			},
			"mappings": map[string]interface{}{
				"dynamic_templates": []map[string]interface{}{
					{"strings": map[string]interface{}{"match_mapping_type": "string", "mapping": map[string]interface{}{"type": "keyword"}}},
					{"wholes": map[string]interface{}{"match_mapping_type": "long", "mapping": map[string]interface{}{"type": "keyword"}}},
					{"fractionals": map[string]interface{}{"match_mapping_type": "double", "mapping": map[string]interface{}{"type": "keyword"}}},
					{"dates": map[string]interface{}{"match_mapping_type": "date", "mapping": map[string]interface{}{"type": "keyword"}}},
					{"booleans": map[string]interface{}{"match_mapping_type": "boolean", "mapping": map[string]interface{}{"type": "keyword"}}},
					{"strings": map[string]interface{}{"match_mapping_type": "string", "mapping": map[string]interface{}{"type": "keyword"}}},
				},
				"properties": map[string]interface{}{
					"@timestamp":      map[string]interface{}{"type": "date", "ignore_malformed": true},
					"log":             map[string]interface{}{"type": "text"},
					"cloudaccount_id": map[string]interface{}{"type": "keyword"},
				},
			},
		},
	}

	templateJson, err := json.Marshal(templateBodyMap)
	if err != nil {
		return "", "", fmt.Errorf("error marshalling template body")
	}

	// JSON string
	templateBody := string(templateJson)

	return url, templateBody, nil
}

func CreateIndex(cloudAccountId string, openSearchURL string) (string, string, error) {
	indexName := fmt.Sprintf("cm_logs_%s-000001", cloudAccountId)
	url := fmt.Sprintf("%s/%s", openSearchURL, indexName)

	aliasName := fmt.Sprintf("cm_logs_%s_wrt_als", cloudAccountId)
	indexBodyMap := map[string]interface{}{
		"aliases": map[string]interface{}{
			aliasName: map[string]interface{}{
				"is_write_index": true,
			},
		},
	}

	indexJson, err := json.Marshal(indexBodyMap)
	if err != nil {
		return "", "", fmt.Errorf("error marshalling index body")
	}

	// JSON string
	indexString := string(indexJson)

	return url, indexString, nil
}
