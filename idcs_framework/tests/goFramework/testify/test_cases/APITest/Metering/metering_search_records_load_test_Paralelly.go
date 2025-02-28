//go:build Load || SearchRecordsParalelly || Metering || Searchrecord || Regression
// +build Load SearchRecordsParalelly Metering Searchrecord Regression

package MeteringAPITest

import (
	//"fmt"
	
	
	//"os"
	"strconv"
	"sync"
	
	"time"
	
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"
	"goFramework/testify/test_cases/testutils"
	"github.com/stretchr/testify/assert"
	"goFramework/utils"
)
func (suite *MeteringAPITestSuite) TestSearchRecordsParallel() {

	var group sync.WaitGroup
	var  count int
	count=0
	start := time.Now()
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)


	for i := 1; i <= numRecords; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()

			id, _ := metering.Create_Record_and_Get_Id()
			logger.Logf.Info("Result of id is",id)
		    assert.NotEqual(suite.T(), id, "", "Successfully created the record")


			if id != "" {

				count=count+1
			    logger.Logf.Info("Number of records created",count)
				i, _ := strconv.ParseInt(id, 10, 64)
			    filter := metering.UsageFilter{
				Id:       testutils.GetInt64Pointer(i),
			}
			ret,_ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
			assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Search API Test with TransactionId filter")
			if ret!= true {
				logger.Logf.Info("Failed to search record with id as ", id)

			}
				
			    
			    if ret == true {
				logger.Logf.Info("Successfully found the record with id", id)
			    logger.Logf.Info("Number of records searched is",count)
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