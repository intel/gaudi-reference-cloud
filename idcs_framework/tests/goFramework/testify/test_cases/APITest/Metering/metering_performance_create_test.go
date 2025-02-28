//go:build Long || CreateRecordsParalelly || Metering || CreateL
// +build Long CreateRecordsParalelly Metering CreateL

package MeteringAPITest

import (
	"fmt"
	
	"goFramework/framework/library/financials/metering"
	"time"
	"goFramework/framework/common/logger"
	"github.com/google/uuid"
	"goFramework/utils"
	"github.com/stretchr/testify/assert"
	//"unsafe"
	"math/rand"
	"os"
	
	
)
func (suite *MeteringAPITestSuite) TestCreateRecords() {

	
	var  count int
	var pickCa string
	var pickRid string
	count=0
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	timetaken:= []string{}
	cloudAccounts := []string{}
	cloudAccountToResMap := map[string][]string{}
	scriptstart := time.Now()
	file, err := os.OpenFile("output1.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	file1, err := os.OpenFile("times.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file1.Close()
	for idx := 1; idx <= 1000; idx++ {

		ca := utils.GenerateInt(12)
		
		cloudAccounts = append(cloudAccounts, ca)
		for idy := 1; idy <= 1000; idy++ {

			rid := uuid.New().String()
			
			cloudAccountToResMap[ca] = append(cloudAccountToResMap[ca], rid)
		}
			
			 
			
	}
	fmt.Println(cloudAccountToResMap)
	for i := 1; i <= numRecords; i++ {
		    randIdx := rand.Intn(1000)
			pickCa = cloudAccounts[randIdx]
			rids := cloudAccountToResMap[pickCa]
			pickRid = rids[randIdx]
			fmt.Printf("cloud account: %s resource-id: %s\n", pickCa, pickRid)
			data1 := fmt.Sprintf("cloud account: %s resource-id: %s\n", pickCa, pickRid)
			_, err := file.WriteString(data1)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			

			start := time.Now()

		    val,data := metering.Create_CloudAccount_and_Get_Id(pickCa, pickRid)
			
			elapsed := time.Since(start)
			timeinstring := elapsed.String()
			logger.Logf.Info("value of jsonstr",data)
			timetaken = append(timetaken,timeinstring)
	        
			//logger.Logf.Info("Time taken for creation of one record ",val)
			logger.Logf.Info("Result of val is ",val)
			assert.Equal(suite.T(), val, true, "Successfully created the record")
			sdata := []byte(data)
			sizeInBytes := len(sdata)
			
            if val == true {
					count=count+1
					logger.Logf.Info("Number of records created is ",count)
					logger.Logf.Info("Size of record is ", sizeInBytes, " bytes")
					logger.Logf.Info("Time taken for creation of one record ",elapsed)
					//write the time into file
					data2 := fmt.Sprintf("%s\n",timeinstring)
					_, err2 := file1.WriteString(data2)
			        if err2 != nil {
				         fmt.Println("Error writing to file:", err)
				         return
			        }
			}
			if val == false{ 
					logger.Logf.Info("Failed to create records")

			}
			
	}
	scriptended := time.Since(scriptstart)
	average, err := utils.CalculateAverageTime(timetaken)
	if err != nil {
		fmt.Printf("Error calculating average: %v\n", err)
		return
	}
	logger.Logf.Info("Average time taken: ",average)
	
	logger.Logf.Info("The script ran for: ",scriptended)


}

