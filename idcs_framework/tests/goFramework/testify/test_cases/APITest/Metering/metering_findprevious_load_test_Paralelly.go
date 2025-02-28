//go:build Load || FindPreviousRecords || Metering || FindPreviousParalelly || Regression
// +build Load FindPreviousRecords Metering FindPreviousParalelly Regression

package MeteringAPITest

import (
	//"fmt"
	
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/common/logger"
	"github.com/stretchr/testify/assert"
	//"os"
	//"strconv"
	"sync"
	"goFramework/utils"
	
	"time"
)

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsParalelly() {
	var group sync.WaitGroup
	var  count int
	count=0
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	start := time.Now()


	for i := 1; i <= numRecords; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()

			id, data := metering.Create_Record_and_Get_Id()
			logger.Logf.Info("Result of id is",id)
		    assert.NotEqual(suite.T(), id, "", "Successfully created the record")

			if id != "" {
		
				count=count+1
			    logger.Logf.Info("Number of records created",count)
				filter := metering.UsagePrevious{
					Id:       id,
					ResourceId: data.ResourceId,
				}
				ret,_ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 200, 1)
				assert.Equal(suite.T(), ret, false, "Test Failed: Metering Find Previous API Test")
				
				if ret!=true  {
					logger.Logf.Info("Failed to find previous  record with id as ", id)
	
				}
				if ret == true {
					logger.Logf.Info("Successfully found the record with id ", id)
					logger.Logf.Info("Number of previous records found  is ",count)
				}
			}
			if id == "" { 
				logger.Logf.Info("Failed to create records")
	
			}
			}(i)
	}
	group.Wait()
	elapsed := time.Since(start)
	logger.Logf.Info("exited at",elapsed)

	
}
	