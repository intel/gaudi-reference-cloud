//go:build Long || CreateRecordsParalelly || Metering || UpdateL
// +build Long CreateRecordsParalelly Metering UpdateL

package MeteringAPITest

import (
	"fmt"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"
	"goFramework/testify/test_cases/testutils"
	"goFramework/utils"
	"time"
	//"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	//"unsafe"
	//"math/rand"
	"bufio"
	"os"
	"strings"
	"strconv"
	
)

func (suite *MeteringAPITestSuite) TestUpdateRecords() {
	//var timeperecord string
	timetaken := []string{}
	var rId string
	var vId string
	var count int
	file3, err := os.OpenFile("updatetime.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
			fmt.Println("Error opening file:", err)
			return
	}
	defer file3.Close()
	scriptstart := time.Now()

	
		file, err := os.Open("outputgetid.txt")
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
				rId = parts[3]
				vId = parts[1]
				logger.Logf.Info("value of rid is ",rId)
				logger.Logf.Info("value of vid is ",vId)
				//cId = "717488864688"
				//rId = "3dffac95-8a91-41f3-9b3d-42e0a31aed6b"
				//vId = "4074"
				res, _:= strconv.ParseInt(vId, 10, 64)

			


				// Use the value as needed
				fmt.Println("Read value:", rId)
				filter := metering.UsageUpdate{
					
					Id : []int64{res},
					Reported: true,
					
				}
				start := time.Now()
				ret := metering.Update_Usage_Record_with_rid(filter,rId,testutils.GetInt64Pointer(res),200,1)
				elapsed := time.Since(start)
				timeinstring := elapsed.String()
				assert.Equal(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test")
				if ret != true {
					logger.Logf.Info("Failed to search record with rid as ", rId)
		
				}
				if ret == true {
					count=count+1
		
					logger.Logf.Info("Successfully found the record with id", vId)
					logger.Logf.Info("Number of records searched is", count)
					logger.Logf.Info("Time taken for searching one filter ", timeinstring)
					//timeperecord = (elapsed / time.Duration(count)).String()
					timetaken = append(timetaken, timeinstring)
					data2 := fmt.Sprintf("%s\n",timeinstring)
					_, err2 := file3.WriteString(data2)
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
	logger.Logf.Info("Average time taken: ",average)
	logger.Logf.Info("The script ran for: ", scriptended)

}
