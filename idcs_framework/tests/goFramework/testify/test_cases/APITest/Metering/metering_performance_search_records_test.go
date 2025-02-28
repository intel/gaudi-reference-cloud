//go:build Long || CreateRecordsParalelly || Metering || SearchL
// +build Long CreateRecordsParalelly Metering SearchL

package MeteringAPITest

import (
	"fmt"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"
	"goFramework/utils"
	"time"
	//"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	//"unsafe"
	//"math/rand"
	"bufio"
	"os"
	"strings"
)

func (suite *MeteringAPITestSuite) TestSearchRecords() {

	//var count int
	//var startT string
	//var endT string
	var timeperecord string
	//count = 0
	
	timetaken := []string{}
	var cId string
	file2, err := os.OpenFile("searchtime.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file2.Close()
	scriptstart := time.Now()

	file, err := os.Open("output1.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Store the value in a variable
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			cId = parts[2]
			//cId = "717488864688"

			// Use the value as needed
			fmt.Println("Read value:", cId)
			filter := metering.UsageFilter{
				CloudAccountId: &cId,
				//StartTime:      &startT,
				//EndTime:        &endT,
			}
			start := time.Now()
			ret, count := metering.Search_Usage_Record_with_Cloud_AccountId(filter, 200)
			elapsed := time.Since(start)
			timeinstring := elapsed.String()

			assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Search API Test with TransactionId filter")
			if ret != true {
				logger.Logf.Info("Failed to search record with id as ", cId)

			}
			if ret == true {
				//count=count+1

				logger.Logf.Info("Successfully found the record with id", cId)
				logger.Logf.Info("Number of records searched is", count)
				logger.Logf.Info("Time taken for searching one filter ", timeinstring)
				timeperecord = (elapsed / time.Duration(count)).String()
				timetaken = append(timetaken, timeperecord)
				data2 := fmt.Sprintf("total records found: %d time taken per record: %s\n", count, timeperecord)
				_, err2 := file2.WriteString(data2)
				if err2 != nil {
					fmt.Println("Error writing to file:", err)
					return
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	scriptended := time.Since(scriptstart)
	//logger.Logf.Info("Times taken: ",timetaken)
	average, err := utils.CalculateAverageTime(timetaken)
	if err != nil {
		fmt.Printf("Error calculating average: %v\n", err)
		return
	}
	logger.Logf.Info("Average time taken: ", average)
	logger.Logf.Info("The script ran for: ", scriptended)

}
