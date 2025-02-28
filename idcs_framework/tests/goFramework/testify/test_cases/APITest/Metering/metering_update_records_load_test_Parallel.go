//go:build Load || UpdateRecordsParalelly || Metering || Updaterecord || Regression
// +build Load UpdateRecordsParalelly Metering Updaterecord Regression

package MeteringAPITest

import (
	//"fmt"
	
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/common/logger"
	"github.com/stretchr/testify/assert"
	"sync"
	//"os"
	"strconv"
	"goFramework/utils"
	
	"time"
)
func (suite *MeteringAPITestSuite) TestUpdateRecordsParalelly() {
	var group sync.WaitGroup
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	var  count int
	count=0
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
				i, _ := strconv.ParseInt(id, 10, 64)
			    filter := metering.UsageUpdate{
				Id:       []int64{i},
				Reported: true,
			}
			ret := metering.Update_Usage_Record_with_dynamic_filter(filter, id, data, 200, 1)
			assert.Equal(suite.T(), ret, true, "Test Failed: Metering Update API Test")

			if ret!= true {
				logger.Logf.Info("Failed to search record with id as ", id)

			}
				
			    
			    if ret == true {
				logger.Logf.Info("Successfully found the record with id", id)
			    logger.Logf.Info("Number of records updated is",count)
			    }

			}

			if id =="" {
				logger.Logf.Info("Failed to create records")
			}
			
		}(i)
	}
	group.Wait()
	elapsed := time.Since(start)
	logger.Logf.Info("exited at",elapsed)

	
}
