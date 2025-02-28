//go:build Long || CreateRecordsParalelly || Metering || CreateId
// +build Long CreateRecordsParalelly Metering CreateId

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
func (suite *MeteringAPITestSuite) TestCreateRecordsSerially() {

	
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
	file, err := os.OpenFile("outputgetid.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	file1, err := os.OpenFile("timesgetid.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
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
		    randIdx := rand.Intn(10)
			pickCa = cloudAccounts[randIdx]
			rids := cloudAccountToResMap[pickCa]
			pickRid = rids[randIdx]
			
			

			start := time.Now()

		    val,data := metering.Create_Record_and_Get_Id_Performance_Testing(pickCa, pickRid)
			
			elapsed := time.Since(start)
			timeinstring := elapsed.String()
			logger.Logf.Info("value of jsonstr",data)
			
			// _, err2 := file1.WriteString(data2)
			// if err2 != nil {
			// 	fmt.Println("Error writing to file:", err)
			// 	return
			// }
			timetaken = append(timetaken,timeinstring)
			fmt.Printf("id: %s resource-id: %s\n", val, pickRid)
			data1 := fmt.Sprintf("id: %s resource-id: %s\n", val, pickRid)
			_, err := file.WriteString(data1)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
	        
			//logger.Logf.Info("Time taken for creation of one record ",val)
			logger.Logf.Info("Result of val is ",val)
			assert.NotEqual(suite.T(), val, "", "Successfully created the record")
			//size := unsafe.Sizeof(data)
			sdata := []byte(data)
			sizeInBytes := len(sdata)
			
            if val != "" {
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
			if val == ""{ 
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

