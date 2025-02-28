//go:build Load || CreateRecordsParalelly || Metering || CreateP
// +build Load CreateRecordsParalelly Metering CreateP

package MeteringAPITest

import (
	//"fmt"
	//"github.com/google/uuid"
	
	"goFramework/framework/library/financials/metering"
	//"os"
	//"strconv"
	"sync"
	"time"
	"goFramework/framework/common/logger"
	"goFramework/utils"
	"github.com/stretchr/testify/assert"
	
	
)

func (suite *MeteringAPITestSuite) TestCreateRecordsParallel() {

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

			id, _ := metering.Create_Record_and_Get_Id()
			logger.Logf.Info("Result of id is",id)
		    assert.NotEqual(suite.T(), id, "", "Successfully created the record")

			if id != "" {
		

				count=count+1
			    logger.Logf.Info("Number of records created",count)
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
