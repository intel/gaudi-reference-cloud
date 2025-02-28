package testsetup

import (
	//"bytes"
	"fmt"
	//"goFramework/framework/library/financials/billing"
	"encoding/json"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/utils"

	"github.com/tidwall/gjson"

	//"io/ioutil"
	"sort"
	"strconv"
	"time"
)

// Cloud Accounts
var ariaAuth string
var ariaClientId string

func Validate_Enroll_Response(data []byte) bool {
	var structResponse CreateCAccEnrollResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func ValidateEnroll(accType string, responseBody string, responseCode int, userName string, baseUrl string, authToken string) error {
	cloud_account_created := gjson.Get(responseBody, "cloudAccountId").String()
	if cloud_account_created == "" {
		return fmt.Errorf("\nCloudAccountId can not be empty")
	}
	cloudaccount_type := gjson.Get(responseBody, "cloudAccountType").String()
	if cloudaccount_type != accType {
		return fmt.Errorf("\naccount Type did not match Expected : %s , Actual : %s", accType, cloudaccount_type)
	}

	// Validate response schema
	flag := Validate_Enroll_Response([]byte(responseBody))
	if !flag {
		return fmt.Errorf("\nresponse validation failed for enroll user, Actual Response : %s ", responseBody)
	}

	// Aria validation, will enable this valiadtion after complete testing, Disabled this code due to the bug : https://internal-placeholder.com/browse/TWC4724-307

	ariaClientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(cloud_account_created, ariaClientId, ariaAuth)
	if response_status != 200 {
		return fmt.Errorf("\naria call failed to validate user enroll for  user : %s, Got Response Code : %d, and Response : %s", userName, response_status, responseBody)
	}
	logger.Logf.Infof("ariaClientId", ariaClientId)
	logger.Logf.Infof("ariaAuth", ariaAuth)
	//_, responseBody = financials.GetAriaAccountDetailsAllForClientId(cloud_account_created, ariaclientId, ariaAuth)
	if accType == "ACCOUNT_TYPE_PREMIUM" || accType == "ACCOUNT_TYPE_ENTERPRISE" {
		var defCredits string
		if accType == "ACCOUNT_TYPE_PREMIUM" {
			defCredits = "2600"
		}

		if accType == "ACCOUNT_TYPE_ENTERPRISE" {
			defCredits = "2500"
		}
		// Check default cloud credits assignned
		credits, err := GetUnappliedCredits(userName, baseUrl, authToken)
		if err != nil {
			return fmt.Errorf("\nfailed to get unappiled credit for user : %s", userName)
		}
		if credits == defCredits {
			logger.Logf.Infof("'\n 2500 default credits successfully applied on enroll for user : %s " + userName)
		} else {
			return fmt.Errorf("\n'\n 2500 default credits not applied on enroll for user : %s", userName)
		}
		client_acc_id1 := gjson.Get(responseBody, "client_acct_id").String()
		if client_acc_id1 != cloud_account_created {
			return fmt.Errorf("\nenroll failed in Aria validation  for username: %s", userName)
		} else {
			logger.Logf.Infof("Aria account created for username : %s " + userName)
		}

	}
	if accType == "ACCOUNT_TYPE_STANDARD" || accType == "ACCOUNT_TYPE_INTEL" { //
		credits, err := GetUnappliedCredits(userName, baseUrl, authToken)
		if err != nil {
			return fmt.Errorf("\nfailed to get unappiled credit for user : %s", userName)
		}
		if credits == "0" {
			logger.Logf.Infof("'\n default credits is 0 for user : %s " + userName)
		} else {
			return fmt.Errorf("\n'\n default credits not equal to 0 for user : %s", userName)
		}
		err_msg := gjson.Get(responseBody, "error_msg").String()
		if err_msg == "account does not exist" { //
			fmt.Printf("Enroll passed in Aria validation for username : %s", userName)
		} else { //
			return fmt.Errorf("\nenroll failed in Aria validation  for username : %s", userName)
		}

	}
	return nil
}

func EnrollUsersOIDC(data string, userName string, accType string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	//oidcUrl := gjson.Get(ConfigData, "oidcUrl").String()
	token := Get_OIDC_Admin_Token()
	cloudaccount_enroll_url := baseUrl + "/v1/cloudaccounts/enroll"
	var premium bool
	var defaultCredits int
	logger.Log.Info("Starting Enrollment for User " + userName)
	if accType == "ACCOUNT_TYPE_PREMIUM" || accType == "ACCOUNT_TYPE_ENTERPRISE" {
		premium = true
		defaultCredits = 2500
	} else {
		premium = false
		defaultCredits = 0
	}
	enroll_payload := CreateCloudAccountEnrollStruct{
		Premium: premium,
	}
	out, _ := json.Marshal(enroll_payload)
	//authToken := "Bearer " + Get_OIDC_Enroll_Token(oidcUrl, tid, userName, eid, idp)
	logger.Logf.Infof("Account creation payload", string(out))
	logger.Logf.Infof("Auth TOken ", authToken)
	logger.Logf.Infof("Enroll Url ", cloudaccount_enroll_url)
	cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, authToken, string(out))
	// If account type is premium then add coupon to it

	if accType == "ACCOUNT_TYPE_PREMIUM" {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := baseUrl + "/v1/billing/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		if coupon_creation_status != 200 {
			return fmt.Errorf("\nenroll failed for username : %s, failed to create coupon for premium user", userName)
		}

		// Redeem coupon
		redeem_coupon_endpoint := baseUrl + "/v1/billing/coupons/redeem"
		cloudAccId, err := GetCloudAccountId(userName, baseUrl, token)
		if err != nil {
			return err
		}
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), cloudAccId)
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		if coupon_redeem_status != 200 {
			return fmt.Errorf("\nenroll failed for username : %s, failed to redeem coupon for premium user", userName)
		}
		cloudaccount_creation_status, cloudaccount_creation_body = financials.CreateCloudAccount(cloudaccount_enroll_url, authToken, string(out))
		defaultCredits = defaultCredits + 100
	}
	err := ValidateEnroll(accType, cloudaccount_creation_body, cloudaccount_creation_status, userName, baseUrl, token)
	if err != nil {
		return fmt.Errorf("\nenroll failed for username : %s", userName)
	}
	client_acct_id := gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	Userdata := UserData{
		AccType:        accType,
		CloudAccountId: client_acct_id,
		TotalCredits:   defaultCredits,
	}
	if Testdata == nil {
		Testdata = make(map[string]UserData)
	}
	Testdata[userName] = Userdata
	logger.Log.Info("Successfully Enrolled User " + userName)
	return nil
}

//TODO : This function might be required if Azure AD comes into picture
// Some testing is pending on this
// func EnrollUsersAzure() error {
// 	baseUrl := gjson.Get(ConfigData, "baseUrl").String()
// 	cloudaccount_enroll_url := baseUrl + "/v1/cloudaccounts/enroll"
// 	result := gjson.Get(ConfigData, "enrollUsers")
// 	var premium bool
// 	result.ForEach(func(key, value gjson.Result) bool {
// 		data := value.String()
// 		userName := gjson.Get(data, "email").String()
// 		accType := gjson.Get(data, "accType").String()
// 		if accType == "ACCOUNT_TYPE_PREMIUM" || accType == "ACCOUNT_TYPE_ENTERPRISE" {
// 			premium = true
// 		} else {
// 			premium = false
// 		}
// 		enroll_payload := CreateCloudAccountEnrollStruct{
// 			Premium: premium,
// 		}
// 		authToken := "Bearer " + Get_Token(userName)
// 		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, authToken, enroll_payload)
// 		err := ValidateEnroll(accType, cloudaccount_creation_body, cloudaccount_creation_status)
// 		if err != nil {
// 			fmt.Errorf("\nenroll failed for username : %s", userName)
// 		}
// 		client_acct_id := "idc." + cloud_account_created
// 		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map3["cloud_account_id"], ariaclientId, ariaAuth)
// 		if accType == "ACCOUNT_TYPE_PREMIUM" || accType == "ACCOUNT_TYPE_ENTERPRISE" {
// 			client_acct_id := "idc." + cloud_account_created
// 			response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map3["cloud_account_id"], ariaclientId, ariaAuth)
// 			client_acc_id1 := gjson.Get(responseBody, "client_acct_id").String()
// 			if client_acc_id1 != client_acct_id {
// 				fmt.Errorf("\nenroll failed in Aria validation  for username: %s", userName)
// 			}

// 		}
// 		if accType == "ACCOUNT_TYPE_STANDARD" || accType == "ACCOUNT_TYPE_INTEL" {
// 			err_msg := gjson.Get(responseBody, "error_msg").String()
// 			if err_msg == "account does not exist" {
// 				fmt.Printf("Enroll passed in Aria validation for username : %s", userName)
// 			} else {
// 				fmt.Errorf("\nenroll failed in Aria validation  for username : %s", userName)
// 			}

// 		}
// 		Testdata[userName] = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
// 		return true // keep iterating
// 	})
// 	return nil
// }

// Product catalog

func ApplyProductsAndVendors() error {
	cloneFlag := gjson.Get(ConfigData, "productCatalog.applyProducts").Bool()
	if cloneFlag {
		repoLink := gjson.Get(ConfigData, "productCatalog.pcRepo").String()
		cloneDir := gjson.Get(ConfigData, "productCatalog.pcCloneDir").String()
		ip := gjson.Get(ConfigData, "clusterDetails.hostIP").String()
		userName := gjson.Get(ConfigData, "clusterDetails.userName").String()
		keyFilePath := gjson.Get(ConfigData, "clusterDetails.sshKeyFile").String()
		tag := gjson.Get(ConfigData, "productCatalog.tag").String()
		cloneCmd := "git clone " + repoLink + " " + cloneDir
		logger.Logf.Infof("Command ", cloneCmd)
		logger.Logf.Infof("repoLink ", repoLink)
		logger.Logf.Infof("cloneDir ", cloneDir)
		logger.Logf.Infof("ip ", ip)
		logger.Logf.Infof("userName  " + userName)
		logger.Logf.Infof("keyFilePath ", keyFilePath)
		err := RemoteCommandExecution(ip, userName, keyFilePath, cloneCmd)
		if err != nil {
			return err
		}

		// Apply Products and Vendors for specific tag
		dirName := cloneDir + "/" + tag
		logger.Logf.Infof("dirName ", dirName)
		productApplyCmd := "kubectl apply -f " + dirName + "/products"
		vendorApplyCommand := "kubectl apply -f " + dirName + "/products"
		err = RemoteCommandExecution(ip, userName, keyFilePath, productApplyCmd)
		if err != nil {
			return err
		}
		err = RemoteCommandExecution(ip, userName, keyFilePath, vendorApplyCommand)
		if err != nil {
			return err
		}

		//Delete the cloned dir

		// cmd := "rm -rf " + cloneDir
		// err = RemoteCommandExecution(ip, userName, keyFilePath, cmd)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}

func Validate_Coupon_Creation_Response(responseBody string) error {
	var structResponse CreateCouponResponse
	flag := utils.CompareJSONToStruct([]byte(responseBody), structResponse)
	if !flag {
		return fmt.Errorf("\nschema Validation Failed for Coupon Create response : %s", &responseBody)
	}
	return nil
}

func Validate_Coupon_Redemption(responseBody string, couponCode string, userName string, numUses int, baseUrl string, authToken string) error {
	// Get on coupon

	coupon_get_body, err1 := GetCoupon(couponCode, userName, baseUrl, authToken)
	if err1 != nil {
		return fmt.Errorf("\nget coupon failed with error : %s", err1.Error())
	}
	noOfUses := gjson.Get(coupon_get_body, "result.numRedeemed").Int()
	if int(noOfUses) != numUses {
		return fmt.Errorf("\nredeem coupon did not update numUses Properly Expected: %d, Actual :%d ", numUses, noOfUses)
	}
	flag := false
	result := gjson.Get(responseBody, "redemptions")
	var err error
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		cloudAccountId := gjson.Get(data, "cloudAccountId").String()
		if cloudAccountId == cloudAccId {
			code := gjson.Get(data, "code").String()
			if code != couponCode {
				flag = true
				err = fmt.Errorf("\nredeem coupon did not update coupon code  Properly Expected: %s, Actual :%s ", couponCode, code)
				return false

			}
			installed := gjson.Get(data, "installed").String()
			if installed != "true" {
				flag = true
				err = fmt.Errorf("\nredeem coupon did not update installed value Properly Expected: %s, Actual :%s ", "true", installed)
				return false
			}
		}
		return true
	})
	if flag {
		return fmt.Errorf("\nredeem coupon validation failed responseBody : %s with error :%s", responseBody, err.Error())
	}
	return nil
}

func GetCoupon(code string, userName string, baseUrl string, authToken string) (string, error) {
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	get_coupon_end_point := baseUrl + "/v1/billing/coupons"
	//authToken := Get_OIDC_Admin_Token()
	coupon_get_status, coupon_get_body := financials.GetCouponsByCode(get_coupon_end_point, authToken, code)
	if coupon_get_status != 200 {
		return "", fmt.Errorf("\nget on Coupon Failed for user : %s, Got Response Code : %d, and Response : %s", userName, coupon_get_status, coupon_get_body)
	}
	return coupon_get_body, nil

}

func RedeemCoupon(userName string, code string, couponAmount int, numUses int, baseUrl string, authToken string) error {
	// Redeem coupon
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	beforeCredits, err := GetUnappliedCredits(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	beforeCreds, _ := strconv.ParseInt(beforeCredits, 10, 64)
	redeem_coupon_endpoint := baseUrl + "/v1/billing/coupons/redeem"
	redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), code, cloudAccId)
	coupon_redeem_status, coupon_redeem_body := financials.RedeemCoupon(redeem_coupon_endpoint, authToken, redeem_payload)
	if coupon_redeem_status != 200 {
		return fmt.Errorf("\nredeem Coupon Failed for user : %s, Got Response Code : %d, and Response : %s", userName, coupon_redeem_status, coupon_redeem_body)
	}
	// Validate Coupon Redemption

	Validate_Coupon_Redemption(coupon_redeem_body, code, userName, numUses, baseUrl, authToken)
	AfterCredits, err := GetUnappliedCredits(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	AfterCreds, _ := strconv.ParseInt(AfterCredits, 10, 64)
	totalCreds := int(beforeCreds) + couponAmount
	if totalCreds != int(AfterCreds) {
		return fmt.Errorf("\nredeem Coupon: Amount not updated in credits for user : %s, Actual : %d, Expected : %d", userName, AfterCreds, totalCreds)
	}
	logger.Log.Info("Successfully Validated Coupon creation and redemption for user " + userName)
	return nil

}

func CreateandRedeemCoupon(data string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	userName := gjson.Get(data, "userName").String()
	payload := gjson.Get(data, "payload").String()
	numUses := gjson.Get(data, "payload.numUses").Int()
	couponAmount := gjson.Get(data, "payload.amount").Int() / gjson.Get(data, "payload.numUses").Int()
	logger.Log.Info("Create and Redeem coupon for users " + userName)
	create_coupon_endpoint := baseUrl + "/v1/billing/coupons"
	//authToken := Get_OIDC_Admin_Token()
	coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, authToken, payload)
	if coupon_creation_status != 200 {
		return fmt.Errorf("\ncoupon Creation Failed for user : %s, Got Response Code : %d, and Response : %s", userName, coupon_creation_status, coupon_creation_body)
	}
	// Validate response
	err := Validate_Coupon_Creation_Response(coupon_creation_body)
	if err != nil {
		return err
	}
	couponData := Coupons{
		Code:   gjson.Get(coupon_creation_body, "code").String(),
		Amount: int(gjson.Get(coupon_creation_body, "amount").Int()),
	}
	if val, ok := Testdata[userName]; ok {
		val.Coupons = append(val.Coupons, couponData)
		Testdata[userName] = val
	}

	// Redeem coupon
	var users []string
	results := gjson.Get(data, "userName")
	results.ForEach(func(key, value gjson.Result) bool {
		println(value.String())
		users = append(users, value.String())
		return true // keep iterating
	})
	count := 0
	for i := 0; i < len(users); i++ {
		if count >= int(numUses) {
			return fmt.Errorf("\ncan not redeem coupon for user : %s, numUses exceeded ", userName)
		}
		userName = users[i]
		cloud_account_created, err := GetCloudAccountId(userName, baseUrl, authToken)
		if err != nil {
			return err
		}
		beforeCredits, error := GetUnappliedCredits(userName, baseUrl, authToken)
		if error != nil {
			// return error
			beforeCredits = "0"
		}
		beforeCreds, _ := strconv.ParseInt(beforeCredits, 10, 64)
		redeem_coupon_endpoint := baseUrl + "/v1/billing/coupons/redeem"
		authToken = Get_OIDC_Admin_Token()
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), cloud_account_created)
		coupon_redeem_status, coupon_redeem_body := financials.RedeemCoupon(redeem_coupon_endpoint, authToken, redeem_payload)
		if coupon_redeem_status != 200 {
			return fmt.Errorf("\nredeem Coupon Failed for user : %s, Got Response Code : %d, and Response : %s", userName, coupon_redeem_status, coupon_redeem_body)
		}
		// Validate Coupon Redemption
		Validate_Coupon_Redemption(coupon_redeem_body, gjson.Get(coupon_creation_body, "code").String(), userName, count, baseUrl, authToken)
		time.Sleep(20 * time.Second)
		AfterCredits, err := GetUnappliedCredits(userName, baseUrl, authToken)
		if err != nil {
			return err
		}
		AfterCreds, _ := strconv.ParseInt(AfterCredits, 10, 64)
		logger.Logf.Infof("Before Credits", beforeCredits)
		logger.Logf.Infof("Coupon AMount Credits", couponAmount)
		totalCreds := beforeCreds + couponAmount
		if totalCreds != AfterCreds {
			return fmt.Errorf("\nredeem Coupon: Amount not updated in credits for user : %s, Actual : %d, Expected : %d", userName, AfterCreds, totalCreds)
		}
		logger.Logf.Infof("\n\nSuccessfully Validated Coupon creation and redemption for user " + userName)
		count = count + 1
		if val, ok := Testdata[userName]; ok {
			val.TotalCredits = val.TotalCredits + int(couponAmount)
			Testdata[userName] = val
		}
	}

	return nil
}

func GetRunningSecondsFoAllInstances(userName string, productType string, cloudAccId string, baseUrl string, authToken string) (float64, error) {
	var metering_record_runningSec float64
	var temp float64
	//var instanceId string
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	logger.Log.Info("Get Metring Usage data for cloud account " + userName)

	// Get all resource Ids
	search_payload := financials_utils.EnrichMmeteringSearchPayloadCloudAcc(compute_utils.GetMeteringSearchPayloadCloudAcc(),
		cloudAccId)
	//authToken := Get_OIDC_Admin_Token()
	response_status, response_body := financials.SearchAllMeteringRecords(metering_api_base_url, authToken, search_payload)
	if response_status != 200 {
		return temp, fmt.Errorf("\nmetering Search Failed  : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
	}
	result := gjson.Parse(response_body)
	arr := gjson.Get(result.String(), "..#.result.resourceId")
	var resourceArray []string
	for _, val := range arr.Array() {
		resourceArray = append(resourceArray, val.String())
	}

	totalRunSeconds := 0.0

	if val, ok := ProductUsageData[userName]; ok {
		// Get usages from ProductUsageDta
		usageData := val.Usage
		for i := 0; i < len(usageData); i++ {
			if usageData[i].ProductType == productType {
				for i := 0; i < len(usageData); i++ {
					resourceArray = append(resourceArray, usageData[i].ResourceIds...)

				}
			}
		}

	}

	resourceArray = RemoveDuplicateValues(resourceArray)
	logger.Logf.Infof("Resource Array", resourceArray)
	for _, resourceId := range resourceArray {
		search_payload = financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(),
			resourceId, cloudAccId)
		//authToken := Get_OIDC_Admin_Token()
		response_status, response_body = financials.SearchAllMeteringRecords(metering_api_base_url, authToken, search_payload)
		result = gjson.Parse(response_body)
		arr = gjson.Get(result.String(), "..#.result.properties.runningSeconds")
		var runnSecArray []float64
		for _, val := range arr.Array() {
			runnSecArray = append(runnSecArray, val.Float())
		}
		// Sort array
		sort.Float64s(runnSecArray)
		if len(runnSecArray) <= 0 {
			return temp, fmt.Errorf("\nmetering doesnt have data for resource : %s", resourceId)

		} else {
			arrLen := len(runnSecArray) - 1
			lastEle := runnSecArray[arrLen]
			metering_record_runningSec = lastEle
		}

		logger.Logf.Infof("metering_record_runningSec %s for resource : %s", metering_record_runningSec, resourceId)
		if response_status != 200 {
			return temp, fmt.Errorf("\nmetering Search Failed  : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
		}
		logger.Logf.Infof("")
		totalRunSeconds = totalRunSeconds + metering_record_runningSec
	}
	logger.Logf.Infof("Total Run sec", totalRunSeconds)
	return totalRunSeconds, nil
}

// Validate instance data in metering

func GetRunningSeconds(userName string, cloudAccId string, resourceId string, baseUrl string, authToken string) (float64, error) {
	var metering_record_runningSec float64
	var temp float64
	var totalRunSeconds float64
	//var instanceId string
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	logger.Log.Info("Get Metring Usage data for cloud account " + userName)
	logger.Logf.Infof("Entered function GetRunningSeconds")

	search_payload := financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(),
		resourceId, cloudAccId)
	//authToken := Get_OIDC_Admin_Token()
	response_status, response_body := financials.SearchAllMeteringRecords(metering_api_base_url, authToken, search_payload)
	result := gjson.Parse(response_body)
	arr := gjson.Get(result.String(), "..#.result.properties.runningSeconds")
	var runnSecArray []float64
	for _, val := range arr.Array() {
		runnSecArray = append(runnSecArray, val.Float())
	}
	// Sort array
	sort.Float64s(runnSecArray)
	if len(runnSecArray) <= 0 {
		return temp, fmt.Errorf("metering doesnt have data for resource : %s", resourceId)
	} else {
		arrLen := len(runnSecArray) - 1
		lastEle := runnSecArray[arrLen]
		metering_record_runningSec = lastEle
	}

	logger.Logf.Infof("metering_record_runningSec %x for resource : %s", metering_record_runningSec, resourceId)
	if response_status != 200 {
		return temp, fmt.Errorf("\nmetering Search Failed  : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
	}
	logger.Logf.Infof("")
	totalRunSeconds = totalRunSeconds + metering_record_runningSec

	// Get all resource Ids
	// search_payload := financials_utils.EnrichMmeteringSearchPayloadCloudAcc(compute_utils.GetMeteringSearchPayloadCloudAcc(),
	// 	cloudAccId)
	// //authToken := Get_OIDC_Admin_Token()
	// response_status, response_body := financials.SearchAllMeteringRecords(metering_api_base_url, authToken, search_payload)
	// if response_status != 200 {
	// 	return temp, fmt.Errorf("\nmetering Search Failed  : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
	// }
	// result := gjson.Parse(response_body)
	// arr := gjson.Get(result.String(), "..#.result.resourceId")
	// var resourceArray []string
	// for _, val := range arr.Array() {
	// 	resourceArray = append(resourceArray, val.String())
	// }
	// resourceArray = RemoveDuplicateValues(resourceArray)
	// totalRunSeconds := 0.0
	// logger.Logf.Infof("Resource Array", resourceArray)
	// for _, resourceId := range resourceArray {
	// 	// search_payload = financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(),
	// 	// 	resourceId, cloudAccId)
	// 	// //authToken := Get_OIDC_Admin_Token()
	// 	// response_status, response_body = financials.SearchAllMeteringRecords(metering_api_base_url, authToken, search_payload)
	// 	// result = gjson.Parse(response_body)
	// 	// arr = gjson.Get(result.String(), "..#.result.properties.runningSeconds")
	// 	// var runnSecArray []float64
	// 	// for _, val := range arr.Array() {
	// 	// 	runnSecArray = append(runnSecArray, val.Float())
	// 	// }
	// 	// // Sort array
	// 	// sort.Float64s(runnSecArray)
	// 	// if len(runnSecArray) <= 0 {
	// 	// 	return temp, fmt.Errorf("\nmetering doesnt have data for resource : %s", resourceId)

	// 	// } else {
	// 	// 	arrLen := len(runnSecArray) - 1
	// 	// 	lastEle := runnSecArray[arrLen]
	// 	// 	metering_record_runningSec = lastEle
	// 	// }

	// 	// logger.Logf.Infof("metering_record_runningSec %s for resource : %s", metering_record_runningSec, resourceId)
	// 	// if response_status != 200 {
	// 	// 	return temp, fmt.Errorf("\nmetering Search Failed  : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
	// 	// }
	// 	// logger.Logf.Infof("")
	// 	// totalRunSeconds = totalRunSeconds + metering_record_runningSec
	// }

	logger.Logf.Infof("Total Run seconds %x", totalRunSeconds)
	return totalRunSeconds, nil
}

//Get Usage and Verify

func GetUsageAndValidate(userName string, productType string, baseUrl string, authToken string, computeUrl string) error {
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	//computeUrl := gjson.Get(ConfigData, "regionalUrl").String()
	//Get cloud account Id from userName
	logger.Logf.Infof("Validate the usage for product ", productType)
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, computeUrl, authToken)
	if err != nil {
		return err
	}
	usage_url := baseUrl + "/v1/billing/usages?cloudAccountId=" + cloudAccId

	// Get instances running for cloud account Id and product name
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	response_status, response_body := frisby.GetAllInstance(instance_endpoint, authToken)
	if response_status != 200 {
		return fmt.Errorf("\nfailed to retrieve VM instance,  Error: %s", response_body)
	}
	result := gjson.Parse(response_body)
	arr := gjson.Get(result.String(), "items")
	runningSecs := 0.0
	arr.ForEach(func(key, value gjson.Result) bool {
		inst_data := value.String()
		logger.Logf.Infof("Instance data", inst_data)
		if gjson.Get(inst_data, "spec.instanceType").String() == productType {
			resourceId := gjson.Get(inst_data, "metadata.resourceId").String()
			logger.Logf.Infof("Entered function GetRunningSeconds : 1 block")
			runTime, _ := GetRunningSeconds(userName, cloudAccId, resourceId, baseUrl, authToken)
			runningSecs = runningSecs + runTime
		}
		return true // keep iterating
	})
	flag := 0
	calcAmount := -1.0
	actualAMount := -1.0
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	if usage_response_status != 200 {
		return fmt.Errorf("\nerror Fetching Usage details  : %s, Got Response Code : %d, and Response : %s", userName, usage_response_status, usage_response_body)
	}
	// Get Usage specific to product
	result = gjson.Parse(usage_response_body)
	arr = gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Dta", data)
		if gjson.Get(data, "productType").String() == productType {
			logger.Logf.Infof("Running time", runningSecs)
			AMount := gjson.Get(data, "amount").String()
			actualAMount, _ = strconv.ParseFloat(AMount, 64)
			usage := runningSecs / 60
			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			calcAmount = usage * rateFactor
			logger.Logf.Infof("Calculated usage", calcAmount)
			logger.Logf.Infof("Actual Usage ", actualAMount)
			if calcAmount != actualAMount {
				flag = 1
			}
		}
		return true // keep iterating
	})

	if flag == 1 {
		return fmt.Errorf("\nerror validating usage amount for product : %s, Actual Amount :%f, Expected Amount :%g", productType, actualAMount, calcAmount)
	}
	return nil

}

func StartSchedulers(baseUrl string, authToken string, action string) error {
	scheduler_endpoint := baseUrl + "/v1/billing/ops/actions/scheduler"
	logger.Logf.Infof("Scheduler end point : ", scheduler_endpoint)
	scheduler_payload := financials_utils.EnrichStartSchedulerPayload(financials_utils.GetStartSchedulerPayload(), action)
	logger.Logf.Infof("Scheduler Payload : ", scheduler_payload)
	response_code, response_body := financials.StartScheduler(scheduler_endpoint, authToken, scheduler_payload)
	if response_code != 200 {
		return fmt.Errorf("\nSchedular start failed for action: %s, response : %s", action, response_body)
	}
	return nil

}

func GetUnappliedCredits(userName string, baseUrl string, authToken string) (string, error) {
	// TODO : Aria validation
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	//Get cloud account Id from userName
	logger.Log.Info("Get unapplied credits for user  " + userName)
	// err := StartSchedulers(baseUrl, authToken, "START_CREDIT_USAGE_SCHEDULER")
	// if err != nil {
	// 	logger.Logf.Errorf("Starting scheduler for credit usage failed")
	// }
	// err = StartSchedulers(baseUrl, authToken, "START_CREDIT_EXPIRY_SCHEDULER")
	// if err != nil {
	// 	logger.Logf.Errorf("Starting scheduler for credit expiry failed")
	// }
	time.Sleep(10 * time.Second)
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return "", err
	}

	credit_url := baseUrl + "/v1/cloudcredits/credit/"
	response_status, response_body := financials.GetUnappliedCredits(credit_url, authToken, cloudAccId)
	if response_status != 200 {
		return "", fmt.Errorf("\nerror Fetching Usage details  : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
	}
	unappliedAmount := gjson.Get(response_body, "unappliedAmount").String()
	return unappliedAmount, nil

}

func ValidateCreditsDeduction(userName string, resourcesInfo ResourcesInfo, baseUrl string, authToken string, computeurl string) error {
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	Load_Test_Setup_Data("testdata.json")
	//var totalCredits float64
	// Get  credits using get credits API
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	getCreditsUrl := baseUrl + "/v1/cloudcredits/credit"
	get_response_status, get_response_body := financials.GetCredits(getCreditsUrl, authToken, cloudAccId)
	if get_response_status != 200 {
		return fmt.Errorf("\nerror Fetching Ccredit details  : %s, Got Response Code : %d, and Response : %s", userName, get_response_status, get_response_body)
	}
	totalUsage, err := GetUsageAndValidateTotalUsage(userName, resourcesInfo, baseUrl, authToken, computeurl)
	if err != nil {
		logger.Logf.Infof("Usage amount mismatch")
		//return err
	}
	logger.Logf.Infof("Overall Usage", ProductUsageData)
	logger.Logf.Infof("Total Usage Amount Calculated : ", totalUsage)

	//totalRemainingAmount := gjson.Get(get_response_body, "totalRemainingAmount").Float()
	totalUsedAmount := gjson.Get(get_response_body, "totalUsedAmount").Float()
	//TODO : Will Need to test this flow e2e
	// if TestSetupData != "" {
	// 	path := userName + "." + "totalCredits"
	// 	totalCredits = gjson.Get(TestSetupData, path).Float()
	// 	creds := totalRemainingAmount + totalUsedAmount
	// 	if totalCredits != creds {
	// 		return fmt.Errorf("\ntotal credits not matching for user %s,, Credits from testdata.json  : %g, and credits from get credits api : %g", userName, totalCredits, creds)
	// 	}
	// }
	// Calculate credit depletion

	//Validate
	if totalUsage != totalUsedAmount {
		return fmt.Errorf("\ntotal credits used not matching for user %s, Credits used from calculation  : %g, and credits used from get credits api : %g", userName, totalUsage, totalUsedAmount)
	}
	return nil

}

func ValidateUsageAfterInstanceDeletion(userName string, productType string, baseUrl string, authToken string, computeUrl string) (float64, error) {
	// err := StartSchedulers(baseUrl, authToken, "START_POST_USAGE_SCHEDULER")
	// if err != nil {
	// 	logger.Logf.Errorf("Starting scheduler for usage failed")
	// }
	// err = StartSchedulers(baseUrl, authToken, "START_CREDIT_USAGE_SCHEDULER")
	// if err != nil {
	// 	logger.Logf.Errorf("Starting scheduler for credit usage failed")
	// }
	var temp float64
	//Get cloud account Id from userName
	logger.Log.Info("Validate total usage for user " + userName)
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return temp, err
	}
	usage_url := baseUrl + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	logger.Log.Info("Usage UURL" + usage_url)

	// Get instances running for cloud account Id and product name

	//runningSecs := 0.0
	if ProductUsageData == nil {
		ProductUsageData = make(map[string]UsageData)
	}
	logger.Logf.Infof("Entered function GetRunningSeconds : 12block")
	runSecs, _ := GetRunningSecondsFoAllInstances(userName, productType, cloudAccId, baseUrl, authToken)
	logger.Logf.Infof("Running Sec from Metering : %f", runSecs)

	logger.Logf.Infof("Product Name :", productType)
	if val, ok := ProductUsageData[userName]; ok {
		logger.Logf.Infof("In first loop", runSecs)
		//runTime, _ := strconv.ParseFloat(runSecs, 64)
		runningSecs := runSecs
		logger.Logf.Infof("Total run secs now ", runningSecs)
		// Get usages from ProductUsageDta
		usageData := val.Usage
		for i := 0; i < len(usageData); i++ {
			if usageData[i].ProductType == productType {
				usageData[i].RunningSecs = runSecs
			}
		}

	}
	jsonContent1, _ := json.MarshalIndent(ProductUsageData, "", "    ")
	logger.Logf.Infof("Product usage data map  : ", string(jsonContent1))
	logger.Logf.Infof("Instance Running time from Metering : ", runSecs)
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	if usage_response_status != 200 {
		return temp, fmt.Errorf("\nerror Fetching Usage details  : %s, Got Response Code : %d, and Response : %s", userName, usage_response_status, usage_response_body)

	}
	// Get Usage specific to product
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		var tmpcalcAmount float64
		var tmpactualAMount float64
		data := value.String()
		logger.Logf.Infof("Usage Data", data)
		product_type := gjson.Get(data, "productType").String()
		AMount := gjson.Get(data, "amount").String()
		tmpactualAMount, _ = strconv.ParseFloat(AMount, 64)
		prod_run_sec := GetProductRunSec(userName, product_type)
		usage := prod_run_sec / 60
		rate := gjson.Get(data, "rate").String()
		rateFactor, _ := strconv.ParseFloat(rate, 64)
		logger.Logf.Infof("Rate Factor of product %v", rateFactor)
		logger.Logf.Infof("Running Seconds product %v", prod_run_sec)
		tmpcalcAmount = usage * rateFactor
		//tmpcalcAmount, _ = strconv.ParseFloat(tmpcalcAmount, 64)
		if val, ok := ProductUsageData[userName]; ok {
			// Get usages from ProductUsageDta
			usageData := val.Usage
			for i := 0; i < len(usageData); i++ {
				if usageData[i].ProductType == productType {
					usageData[i].RunningSecs = prod_run_sec
					usageData[i].Amount = tmpcalcAmount
					usageData[i].Rate = rateFactor

				}
			}

		}

		logger.Logf.Infof("Calculated amount from usage %f", tmpcalcAmount)
		logger.Logf.Infof("Actual  amount from usage %f", tmpactualAMount)
		if tmpcalcAmount != tmpactualAMount {
			_ = fmt.Errorf("\nerror validating usage amount for product : %s, Actual Amount :%f, Expected Amount :%f", productType, tmpactualAMount, tmpcalcAmount)
			return false
		}
		return true // keep iterating
	})
	totalAmount := GetTotalUsageAmount(userName)
	jsonContent, _ := json.MarshalIndent(ProductUsageData, "", "    ")
	logger.Logf.Infof("Product usage data map  : ", string(jsonContent))
	// Validate usage for the customer
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	logger.Logf.Infof("Calculated total amount from usage %v", totalAmount)
	logger.Logf.Infof("Actual  total amount from Api %v", total_amount_from_response)
	if totalAmount != total_amount_from_response {
		return totalAmount, fmt.Errorf("\nerror validating total usage amount for product : %s, Actual Amount :%g, Expected Amount :%g", productType, total_amount_from_response, totalAmount)
	}
	return totalAmount, nil
}

func GetProductRunSec(userName string, product_type string) float64 {
	var tmp float64
	if val, ok := ProductUsageData[userName]; ok {
		// Get usages from ProductUsageDta
		usageData := val.Usage
		for i := 0; i < len(usageData); i++ {
			if usageData[i].ProductType == product_type {
				return usageData[i].RunningSecs
			}
		}

	}
	return tmp
}

func GetProductUsage(userName string, product_type string) float64 {
	var tmp float64
	if val, ok := ProductUsageData[userName]; ok {
		// Get usages from ProductUsageDta
		usageData := val.Usage
		for i := 0; i < len(usageData); i++ {
			if usageData[i].ProductType == product_type {
				return usageData[i].Amount
			}
		}

	}
	return tmp
}

func GetTotalUsageAmount(userName string) float64 {
	var tmp float64
	totalUsage := 0.0
	if val, ok := ProductUsageData[userName]; ok {
		// Get usages from ProductUsageDta
		usageData := val.Usage
		for i := 0; i < len(usageData); i++ {
			totalUsage = totalUsage + usageData[i].Amount
		}
		return totalUsage

	}
	return tmp
}

func GetInstanceDetails(userName string, baseUrl string, authToken string, computeUrl string) (ResourcesInfo, error) {
	logger.Log.Info("Get Instance details for user " + userName)
	var allRes []Resources
	var resourcesInfo ResourcesInfo
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return resourcesInfo, err
	}
	usage_url := baseUrl + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	logger.Log.Info("Usage URL " + usage_url)

	// Get instances running for cloud account Id and product name
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	logger.Log.Info("Instance Url" + instance_endpoint)
	response_status, response_body := frisby.GetAllInstance(instance_endpoint, authToken)
	//logger.Log.Info("Instance Response body" + response_body)
	if response_status != 200 {
		return resourcesInfo, fmt.Errorf("\nfailed to retrieve VM instance,  Error: %s", response_body)
	}
	result := gjson.Parse(response_body)
	arr := gjson.Get(result.String(), "items")

	arr.ForEach(func(key, value gjson.Result) bool {
		inst_data := value.String()
		logger.Logf.Infof("Instance data", inst_data)
		if ProductUsageData == nil {
			ProductUsageData = make(map[string]UsageData)
		}
		productName := gjson.Get(inst_data, "spec.instanceType").String()
		resourceId := gjson.Get(inst_data, "metadata.resourceId").String()
		resourceData := Resources{
			ResourceId:  resourceId,
			ProductName: productName,
		}
		allRes = append(allRes, resourceData)
		return true
	})
	resourcesInfo = ResourcesInfo{
		ProductData: allRes,
	}
	return resourcesInfo, nil
}

func GetUsageAndValidateTotalUsage(userName string, resourceInfo ResourcesInfo, baseUrl string, authToken string, computeUrl string) (float64, error) {
	// err := StartSchedulers(baseUrl, authToken, "START_POST_USAGE_SCHEDULER")
	// if err != nil {
	// 	logger.Logf.Errorf("Starting scheduler for usage failed with error : ", err)
	// }
	// err = StartSchedulers(baseUrl, authToken, "START_CREDIT_USAGE_SCHEDULER")
	// if err != nil {
	// 	logger.Logf.Errorf("Starting scheduler for credit usage failed with error : ", err)
	// }

	var temp float64
	//Get cloud account Id from userName
	logger.Log.Info("Validate total usage for user  " + userName)
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return temp, err
	}
	runningSecs := 0.0
	logger.Logf.Infof("Initial run seconds ", runningSecs)
	usage_url := baseUrl + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	logger.Log.Info("Usage URL" + usage_url)

	InstanceInfo := resourceInfo.ProductData
	for i := 0; i < len(InstanceInfo); i++ {
		if ProductUsageData == nil {
			ProductUsageData = make(map[string]UsageData)
		}
		productName := InstanceInfo[i].ProductName
		resourceId := InstanceInfo[i].ResourceId

		runSecs, _ := GetRunningSeconds(userName, cloudAccId, resourceId, baseUrl, authToken)

		if val, ok := ProductUsageData[userName]; ok {
			//runTime, _ := strconv.ParseFloat(runSecs, 64)
			runningSecs = runSecs
			// Get usages from ProductUsageDta
			usageData := val.Usage
			prodFound := false
			for i := 0; i < len(usageData); i++ {
				if usageData[i].ProductType == productName {
					prodFound = true
					resourceIds := usageData[i].ResourceIds
					ret := SearchSlice(resourceIds, resourceId)
					if !ret {
						usageData[i].ResourceIds = append(usageData[i].ResourceIds, resourceId)
					}
					usageData[i].RunningSecs = runSecs

				}
			}
			if !prodFound {
				var resources []string
				resources = append(resources, resourceId)
				prod_data := ProductRunTime{
					RunningSecs: runSecs,
					ResourceIds: resources,
					ProductType: productName,
				}
				usageData = append(usageData, prod_data)
				usageD := UsageData{
					Usage: usageData,
				}
				ProductUsageData[userName] = usageD

			}

		} else {
			//runTime, _ := strconv.ParseFloat(runSecs, 64)

			var resources []string
			resources = append(resources, resourceId)
			prod_data := ProductRunTime{
				RunningSecs: runSecs,
				ResourceIds: resources,
				ProductType: productName,
			}
			var vall []ProductRunTime
			vall = append(vall, prod_data)
			usageD := UsageData{
				Usage: vall,
			}
			ProductUsageData[userName] = usageD

		}

	}

	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	if usage_response_status != 200 {
		return temp, fmt.Errorf("\nerror Fetching Usage details  : %s, Got Response Code : %d, and Response : %s", userName, usage_response_status, usage_response_body)

	}
	fmt.Println("Usage Response body", usage_response_body)
	// Get Usage specific to product
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		var tmpcalcAmount float64
		var tmpactualAMount float64
		data := value.String()
		logger.Logf.Infof("Usage Data", data)
		product_type := gjson.Get(data, "productType").String()
		prod_run_sec := GetProductRunSec(userName, product_type)
		if prod_run_sec == 0 {
			_ = fmt.Errorf("\n Total run seconds of the product %s is zero, it should be greater than zero", product_type)
			return false
		}
		AMount := gjson.Get(data, "amount").String()
		tmpactualAMount, _ = strconv.ParseFloat(AMount, 64)
		usage := prod_run_sec / 60
		rate := gjson.Get(data, "rate").String()
		rateFactor, _ := strconv.ParseFloat(rate, 64)
		tmpcalcAmount = usage * rateFactor
		//tmpcalcAmount, _ = strconv.ParseFloat(tmpcalcAmount, 64)
		if val, ok := ProductUsageData[userName]; ok {
			// Get usages from ProductUsageDta
			usageData := val.Usage
			for i := 0; i < len(usageData); i++ {
				if usageData[i].ProductType == product_type {
					usageData[i].RunningSecs = prod_run_sec
					usageData[i].Amount = tmpcalcAmount
					usageData[i].Rate = rateFactor

				}
			}

		}
		if tmpcalcAmount != tmpactualAMount {
			_ = fmt.Errorf("\nerror validating usage amount  Actual Amount :%f, Expected Amount :%f", tmpactualAMount, tmpcalcAmount)
			return false
		}
		return true // keep iterating
	})

	totalAmount := GetTotalUsageAmount(userName)
	jsonContent, _ := json.MarshalIndent(ProductUsageData, "", "    ")
	logger.Log.Info("Product usage data map  : " + string(jsonContent))
	// Validate usage for the customer
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	logger.Logf.Infof("Calculated total amount from usage %v", totalAmount)
	logger.Logf.Infof("Actual  total amount from Api %v", total_amount_from_response)
	total_amount_from_response = RoundFloat(total_amount_from_response, 8)
	totalAmount = RoundFloat(totalAmount, 8)
	fmt.Println("Total Amount from response", total_amount_from_response)
	fmt.Println("Total Amount from calculation", totalAmount)
	if totalAmount != total_amount_from_response {
		return totalAmount, fmt.Errorf("\nerror validating total usage amount  Actual Amount :%g, Expected Amount :%g", total_amount_from_response, totalAmount)
	}
	return totalAmount, nil
}

// Delete setup

// Delete created data

func DeleteCloudAccount(userName string, baseUrl string, authToken string) error {
	logger.Logf.Infof("Delete cloud account")
	//baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	//authToken := Get_OIDC_Admin_Token()
	logger.Log.Info("Delete cloud account for user  " + userName)
	url := baseUrl + "/v1/cloudaccounts"

	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	delete_Cacc, _ := financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	if delete_Cacc != 200 {
		return fmt.Errorf("\ndeletion of cloud account failed for user  : %s", userName)
	}
	logger.Log.Info("Successfully Deleted cloud account for user " + userName)
	return nil
}

// Restart Service

func RestartService(serviceName string) error {
	ip := gjson.Get(ConfigData, "clusterDetails.hostIP").String()
	userName := gjson.Get(ConfigData, "clusterDetails.userName").String()
	keyFilePath := gjson.Get(ConfigData, "clusterDetails.sshKeyFile").String()
	serviceStopCmd := "kubectl -n idcs-system scale --replicas=0 deployment " + serviceName
	err := RemoteCommandExecution(ip, userName, keyFilePath, serviceStopCmd)
	if err != nil {
		return err
	}

	serviceStartCmd := "kubectl -n idcs-system scale --replicas=1 deployment " + serviceName
	err = RemoteCommandExecution(ip, userName, keyFilePath, serviceStartCmd)
	if err != nil {
		return err
	}
	return nil
}
