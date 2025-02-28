//go:build Performance || CreateRecordsParalelly || Metering
// +build Performance CreateRecordsParalelly Metering

package MeteringAPITest

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/library/financials/metering"
	"goFramework/testify/test_cases/testutils"
	"goFramework/utils"
	"time"

	//"unsafe"
	"bufio"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func (suite *MeteringAPITestSuite) TestPerformance() {
	var count int
	var pickCa string
	var pickRid string
	var timeperecord string
	var cId string
	var vId string
	var searchcount int
	searchcount = 0
	var rId string
	var urId string
	var recId string
	var uvId string
	var previouscount int
	var updatecount int
	count = 0
	previouscount = 0
	updatecount = 0
	getnumberofRecordsPerf := utils.Get_PerformanceInput()
	numRecordsPerf := int(getnumberofRecordsPerf)
	getnumberofCloudacc := utils.Get_CloudaccountInput()
	numCloud := int(getnumberofCloudacc)
	createtimetaken := []string{}
	searchtimetaken := []string{}
	findprevioustimetaken := []string{}
	updatetimetaken := []string{}
	cacc := []string{}
	cloudAccountToResMap := map[string][]string{}
	createscriptstart := time.Now()
	file, err := os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	timefile, err := os.OpenFile("times.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer timefile.Close()
	searchfile, err := os.OpenFile("searchtime.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer searchfile.Close()
	readfile, err := os.Open("output.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer readfile.Close()
	previousfile, err := os.OpenFile("findprevioustime.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer previousfile.Close()
	updatefile, err := os.OpenFile("updatetime.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer updatefile.Close()
	for idx := 1; idx <= numCloud; idx++ {

		name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
		get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
			true, false, "ACCOUNT_TYPE_INTEL", 200)
		assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a intel user cloud account")

		cacc = append(cacc, get_CAcc_id)
		for idy := 1; idy <= numCloud; idy++ {

			rid := uuid.New().String()

			cloudAccountToResMap[get_CAcc_id] = append(cloudAccountToResMap[get_CAcc_id], rid)
		}

	}
	fmt.Println(cloudAccountToResMap)
	for i := 1; i <= numRecordsPerf; i++ {
		randIdx := rand.Intn(10)
		pickCa = cacc[randIdx]
		rids := cloudAccountToResMap[pickCa]
		pickRid = rids[randIdx]

		startcreate := time.Now()

		val, data := metering.Create_Record_and_Get_Id_Performance_Testing(pickCa, pickRid)

		elapsed := time.Since(startcreate)
		createtimeinstring := elapsed.String()
		logger.Logf.Info("value of jsonstr", data)

		createtimetaken = append(createtimetaken, createtimeinstring)
		fmt.Printf("cloudaccount-id: %s id: %s resource-id: %s\n", pickCa, val, pickRid)
		data1 := fmt.Sprintf("cloudaccount-id: %s id: %s resource-id: %s\n", pickCa, val, pickRid)
		_, err := file.WriteString(data1)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		//logger.Logf.Info("Time taken for creation of one record ",val)
		logger.Logf.Info("Result of val is ", val)
		assert.NotEqual(suite.T(), val, "", "Successfully created the record")
		//size := unsafe.Sizeof(data)
		sdata := []byte(data)
		sizeInBytes := len(sdata)

		if val != "" {
			count = count + 1
			logger.Logf.Info("Number of records created is ", count)
			logger.Logf.Info("Size of record is ", sizeInBytes, " bytes")
			logger.Logf.Info("Time taken for creation of one record ", elapsed)
			//write the time into file
			data2 := fmt.Sprintf("%s\n", createtimeinstring)
			_, err2 := timefile.WriteString(data2)
			if err2 != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}
		if val == "" {
			logger.Logf.Info("Failed to create records")

		}

	}
	scriptended := time.Since(createscriptstart)
	average, err := utils.CalculateAverageTime(createtimetaken)
	if err != nil {
		fmt.Printf("Error calculating average: %v\n", err)
		return
	}
	logger.Logf.Info("Average time taken: ", average)

	logger.Logf.Info("The script ran for: ", scriptended)
	//search API
	searchscriptstart := time.Now()
	scanner := bufio.NewScanner(readfile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			cId = parts[1]
			recId = parts[3]

			fmt.Println("Read value:", cId)
			i, _ := strconv.ParseInt(recId, 10, 64)
			filter := metering.UsageFilter{
				//CloudAccountId: &cId,
				Id: testutils.GetInt64Pointer(i),
			}
			searchstart := time.Now()
			logger.Logf.Info(filter)
			ret, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
			searchelapsed := time.Since(searchstart)
			searchtimeinstring := searchelapsed.String()
			assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Search API Test with TransactionId filter")
			if ret != true {
				logger.Logf.Info("Failed to search record with id as ", cId)

			}
			if ret == true {
				searchcount = searchcount + 1

				logger.Logf.Info("Successfully found the record with id", cId)
				logger.Logf.Info("Number of records searched is", searchcount)
				logger.Logf.Info("Time taken for searching one filter ", searchtimeinstring)
				timeperecord = (searchelapsed / time.Duration(count)).String()
				searchtimetaken = append(searchtimetaken, timeperecord)
				data2 := fmt.Sprintf("total records found: %d time taken per record: %s\n", searchcount, timeperecord)
				_, err2 := searchfile.WriteString(data2)
				if err2 != nil {
					fmt.Println("Error writing to file:", err)
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	searchscriptended := time.Since(searchscriptstart)
	searchaverage, err := utils.CalculateAverageTime(searchtimetaken)
	if err != nil {
		fmt.Printf("Error calculating average: %v\n", err)
		return
	}
	logger.Logf.Info("Average time taken: ", searchaverage)
	logger.Logf.Info("The script ran for: ", searchscriptended)
	//findprevious api
	findpreviousscriptstart := time.Now()

	readfileprevious, err := os.Open("output.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer readfileprevious.Close()
	scanner1 := bufio.NewScanner(readfileprevious)
	for scanner1.Scan() {
		line1 := scanner1.Text()

		// Store the value in a variable
		parts := strings.Fields(line1)
		if len(parts) >= 2 {
			rId = parts[5]
			vId = parts[3]
			logger.Logf.Info("value of rid is ", rId)
			logger.Logf.Info("value of vid is ", vId)
			//cId = "717488864688"
			//rId = "3dffac95-8a91-41f3-9b3d-42e0a31aed6b"
			//vId = "3970"

			// Use the value as needed
			fmt.Println("Read value:", rId)
			filter := metering.UsagePrevious{
				ResourceId: rId,
				Id:         vId,
			}
			findpreviousstart := time.Now()
			ret, fcount := metering.Find_Usage_Record_with_id(filter, 200, 1)
			findpreviouselapsed := time.Since(findpreviousstart)
			previoustimeinstring := findpreviouselapsed.String()
			assert.Equal(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test")
			if ret != true {
				logger.Logf.Info("Failed to search record with rid as ", rId)

			}

			if ret == true && fcount == 1 {
				previouscount = previouscount + 1

				logger.Logf.Info("Successfully found the record with id", vId)
				logger.Logf.Info("Number of records searched is", previouscount)
				logger.Logf.Info("Time taken for searching one previous record ", previoustimeinstring)
				//timeperecord = (elapsed / time.Duration(count)).String()
				findprevioustimetaken = append(findprevioustimetaken, previoustimeinstring)
				data2 := fmt.Sprintf(" %s\n", previoustimeinstring)
				_, err2 := previousfile.WriteString(data2)
				if err2 != nil {
					fmt.Println("Error writing to file:", err)
					return
				}
			}
		}
	}
	if err := scanner1.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	findpreviousscriptended := time.Since(findpreviousscriptstart)
	//logger.Logf.Info("Times taken: ",timetaken)
	findpreviousaverage, err := utils.CalculateAverageTime(findprevioustimetaken)
	if err != nil {
		fmt.Printf("Error calculating average: %v\n", err)
		return
	}
	logger.Logf.Info("Average time taken for finding previous: ", findpreviousaverage)
	logger.Logf.Info("The Find previous script ran for: ", findpreviousscriptended)

	//updateAPI
	updatescriptstart := time.Now()
	updatereadfile, err := os.Open("output.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer updatereadfile.Close()
	scanner2 := bufio.NewScanner(updatereadfile)
	for scanner2.Scan() {
		line := scanner2.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			urId = parts[3]
			uvId = parts[1]
			logger.Logf.Info("value of rid is ", urId)
			logger.Logf.Info("value of vid is ", uvId)
			res, _ := strconv.ParseInt(vId, 10, 64)
			fmt.Println("Read value:", rId)
			filter := metering.UsageUpdate{

				Id:       []int64{res},
				Reported: true,
			}
			updatestart := time.Now()
			ret := metering.Update_Usage_Record_with_rid(filter, rId, testutils.GetInt64Pointer(res), 200, 1)
			updateelapsed := time.Since(updatestart)
			updatetimeinstring := updateelapsed.String()
			assert.Equal(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test")
			if ret != true {
				logger.Logf.Info("Failed to search record with rid as ", rId)

			}
			if ret == true {
				updatecount = updatecount + 1

				logger.Logf.Info("Successfully found the record with id", vId)
				logger.Logf.Info("Number of records updated is", count)
				logger.Logf.Info("Time taken for updating one record ", updatetimeinstring)
				//timeperecord = (elapsed / time.Duration(count)).String()
				updatetimetaken = append(updatetimetaken, updatetimeinstring)
				data2 := fmt.Sprintf("%s\n", updatetimeinstring)
				_, err2 := updatefile.WriteString(data2)
				if err2 != nil {
					fmt.Println("Error writing to file:", err)
					return
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	updatescriptended := time.Since(updatescriptstart)
	//logger.Logf.Info("Times taken: ",timetaken)
	updateaverage, err := utils.CalculateAverageTime(updatetimetaken)
	if err != nil {
		fmt.Printf("Error calculating average: %v\n", err)
		return
	}
	logger.Logf.Info("Average time taken for updating records : ", updateaverage)
	logger.Logf.Info("The update api script ran for: ", updatescriptended)
	logger.Logf.Info("Average time taken for creating records : ", average)
	logger.Logf.Info("Average time taken for searching records: ", searchaverage)
	logger.Logf.Info("Average time taken for updating records : ", updateaverage)
	logger.Logf.Info("Average time taken for finding previous: ", findpreviousaverage)

}
