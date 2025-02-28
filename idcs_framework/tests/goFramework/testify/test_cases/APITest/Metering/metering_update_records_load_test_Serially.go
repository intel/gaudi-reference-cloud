//go:build Load || UpdateRecordsSerially || Metering || Updaterecord || Regression
// +build Load UpdateRecordsSerially Metering Updaterecord Regression

package MeteringAPITest

import (
	//"fmt"
	
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/common/logger"
	"github.com/stretchr/testify/assert"
	//"sync"
	//"os"
	"strconv"
	"goFramework/utils"
	
	"time"
)

func (suite *MeteringAPITestSuite) TestUpdateRecordsSerially() {
	start := time.Now()
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	var  count int
	count=0
	for i := 1; i <=numRecords; i++ {
		logger.Logf.Info("Time started is",start)
		id, data := metering.Create_Record_and_Get_Id()
		logger.Logf.Info("Result of id is",id)
		assert.NotEqual(suite.T(), id, "", "Successfully created the record")

		if id != "" {
            count=count+1
			logger.Logf.Info("Number of records created is",count)
			i, _ := strconv.ParseInt(id, 10, 64)
			filter := metering.UsageUpdate{
				Id:       []int64{i},
				Reported: true,
			}
			ret := metering.Update_Usage_Record_with_dynamic_filter(filter, id, data, 200, 1)
			assert.Equal(suite.T(), ret, true, "Test Failed: Metering Update API Test")

			
			if ret!=true  {
				logger.Logf.Info("Failed to update record with id as ", id)

			}
			if ret == true {
				logger.Logf.Info("Successfully found the record with id ", id)
			    logger.Logf.Info("Number of records updated is ",count)
			}
		}
		if id == "" { 
			logger.Logf.Info("Failed to create records")

		}
	}
	
	
}



