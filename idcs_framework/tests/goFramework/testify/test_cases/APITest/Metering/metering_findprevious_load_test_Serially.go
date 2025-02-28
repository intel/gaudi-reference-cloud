//go:build Load || FindPreviousRecords || Metering || FindPreviousSerially || Regression
// +build Load FindPreviousRecords Metering FindPreviousSerially Regression

package MeteringAPITest

import (
	//"fmt"
	
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/common/logger"
	"github.com/stretchr/testify/assert"
	"goFramework/utils"
	//"os"
	//"strconv"
	//"sync"
	
	"time"
)

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsSerially() {
	start := time.Now()
	var  count int
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	count=0
	for i := 1; i <= numRecords; i++ {
		logger.Logf.Info("Time started is",start)
		id, data := metering.Create_Record_and_Get_Id()
		logger.Logf.Info("Result of id is",id)
		assert.NotEqual(suite.T(), id, "", "Successfully created the record")

		if id != "" {
            count=count+1
			logger.Logf.Info("Number of records created is",count)
			//i, _ := strconv.ParseInt(id, 10, 64)
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
	}
			
	
}

