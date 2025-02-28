//go:build Load || MultiOperations|| Metering  
// +build Load MultiOperations Metering 

package MeteringAPITest

import (
	//"fmt"
	//"flag"
	//"github.com/google/uuid"
	
	"goFramework/framework/library/financials/metering"
	"github.com/stretchr/testify/assert"
	//"os"
	"strconv"
	"sync"
	"time"
	"goFramework/framework/common/logger"
	"goFramework/testify/test_cases/testutils"
	"goFramework/utils"
	
)
func (suite *MeteringAPITestSuite) TestMultiOperations() {
	var group sync.WaitGroup
	var  count int
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
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
				filter1 := metering.UsagePrevious{
					Id:       id,
					ResourceId: data.ResourceId,
				}
				ret1,_ := metering.Find_Usage_Record_with_dynamic_filter(filter1, data.Timestamp, 200, 1)
				assert.Equal(suite.T(), ret1, false, "Test Failed: Metering Find Previous API Test")
				
				
				if ret1!=true  {
					logger.Logf.Info("Failed to find previous  record with id as ", id)
	
				}
				if ret1 == true {
					logger.Logf.Info("Successfully found the  usage previousrecord with id ", id)
					logger.Logf.Info("Number of previous records found  is ",count)
				}
				n, _ := strconv.ParseInt(id, 10, 64)
			    filter2 := metering.UsageUpdate{
				Id:       []int64{n},
				Reported: true,
			}
			ret2 := metering.Update_Usage_Record_with_dynamic_filter(filter2, id, data, 200, 1)
			assert.Equal(suite.T(), ret2, true, "Test Failed: Metering Update API Test")
			if ret2!= true {
				logger.Logf.Info("Failed to search record with id as ", id)

			}
				
			    
			    if ret2 == true {
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
