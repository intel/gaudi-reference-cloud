//go:build Long || CreateRecordsParalelly || Metering || CreateParalell
// +build Long CreateRecordsParalelly Metering CreateParalell

package MeteringAPITest

import (
	"fmt"
	
	"goFramework/framework/library/financials/metering"
	"time"
	"goFramework/framework/common/logger"
	"github.com/google/uuid"
	"goFramework/utils"
	"github.com/stretchr/testify/assert"
	"unsafe"
	"math/rand"
	"os"
	"sync"
	
	
)
func (suite *MeteringAPITestSuite) TestCreateRecordsParalell() {
	var group sync.WaitGroup  
	var count int
	var pickCa string
	var pickRid string
	count=0
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	//timetaken:= []string{}
	cloudAccounts := []string{}
	cloudAccountToResMap := map[string][]string{}
	scriptstart := time.Now()
	file, err := os.OpenFile("outputparalell.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	file1, err := os.OpenFile("timesparalell.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file1.Close()
	for idx := 1; idx <= 100; idx++ {

		ca := utils.GenerateInt(12)
		
		cloudAccounts = append(cloudAccounts, ca)
		for idy := 1; idy <= 100; idy++ {

			rid := uuid.New().String()
			
			cloudAccountToResMap[ca] = append(cloudAccountToResMap[ca], rid)
		}
			
			 
			
	}
	fmt.Println(cloudAccountToResMap)
	for i := 1; i <= numRecords; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			randIdx := rand.Intn(7)
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
			data2 := fmt.Sprintf("%s\n",timeinstring)
			_, err2 := file1.WriteString(data2)
			if err2 != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			//timetaken = append(timetaken,timeinstring)
	        
			//logger.Logf.Info("Time taken for creation of one record ",val)
			logger.Logf.Info("Result of val is ",val)
			assert.Equal(suite.T(), val, true, "Successfully created the record")
			size := unsafe.Sizeof(data)
			
            if val == true {
					count=count+1
					logger.Logf.Info("Number of records created is ",count)
					logger.Logf.Info("Size of record is ", size, " bytes")
					logger.Logf.Info("Time taken for creation of one record ",elapsed)
			}
			if val == false{ 
					logger.Logf.Info("Failed to create records")

			}
		}(i)
	}
	group.Wait()
	scriptended := time.Since(scriptstart)
	//logger.Logf.Info("Times taken: ",timetaken)
	logger.Logf.Info("The script ran for: ",scriptended)
}