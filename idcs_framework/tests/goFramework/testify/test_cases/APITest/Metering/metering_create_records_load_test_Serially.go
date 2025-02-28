//go:build Load || CreateRecordsSerially || Metering || CreateS 
// +build Load CreateRecordsSerially Metering CreateS 

package MeteringAPITest

import (
	//"fmt"
	//"flag"
	//"github.com/google/uuid"
	
	"goFramework/framework/library/financials/metering"
	//"os"
	//"strconv"
	//"sync"
	"time"
	"goFramework/framework/common/logger"
	"goFramework/utils"
	"github.com/stretchr/testify/assert"
	
)

func (suite *MeteringAPITestSuite) TestCreateRecordsSerially() {

	start := time.Now()
	var  count int
	
	count=0
	getnumberofRecords := utils.Get_numofRecords()
    numRecords:= int(getnumberofRecords)
	
	for i := 1; i <= numRecords; i++ {
		
		
		logger.Logf.Info("Time started is",start)
		                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    

		id, _ := metering.Create_Record_and_Get_Id()
		logger.Logf.Info("Result of id is",id)
		assert.NotEqual(suite.T(), id, "", "Successfully created the record")
		if id != "" {
            count=count+1
			logger.Logf.Info("Number of records created is",count)
		}
		if id == ""{ 
			logger.Logf.Info("Failed to create records")

		}
	}
	
	elapsed := time.Since(start)
	logger.Logf.Info("exited at",elapsed)


}





