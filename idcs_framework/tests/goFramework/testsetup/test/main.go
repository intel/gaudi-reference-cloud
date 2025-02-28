package main

import (
	"fmt"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/framework/common/logger"
	// "github.com/tidwall/gjson"
	// "time"
)

func main() {	

	logger.InitializeZapCustomLogger()
	testsetup.Get_config_file_data("../config.json")
	compute_utils.LoadE2EConfig("../../ginkGo/compute/data", "vmaas_input.json")
	financials_utils.LoadE2EConfig("../../ginkGo/compute/data", "billing.json")
	testsetup.EnrollUsers()
	// testsetup.CreateandRedeemCoupons()
	// testsetup.CreateVents()
	// testsetup.CreateSSHKeys()
	// testsetup.CreateVmInstances()
	// testsetup.WriteTestDatatoJson()
	// fmt.Println("Waiting for the instances to comeup  ....\n")
	// time.Sleep(3 * time.Minute)

	// baseUrl := gjson.Get(testsetup.ConfigData, "baseUrl").String()
	// regionalUrl := gjson.Get(testsetup.ConfigData, "regionalUrl").String()

	// //userToken := testsetup.Get_User_Token("varsha.karinje@intel.com", "ACCOUNT_TYPE_INTEL", "varshaeid-01")
	// authToken := testsetup.Get_OIDC_Admin_Token()
	// resourceInfo, _ := testsetup.GetInstanceDetails("varsha.karinje@intel.com", baseUrl, authToken, regionalUrl)
    // time.Sleep(3 * time.Minute)
	// usageData, _ := testsetup.GetUsageAndValidateTotalUsage("varsha.karinje@intel.com", resourceInfo, baseUrl, authToken, regionalUrl)
    // fmt.Println("Usage Data for cloud accountId ", usageData)    

	// Sample test calls
	// runningSec, _ := testsetup.GetRunningSeconds("testuser1@intel.com", "211547377448", "6d707d7b-76ac-41a3-ba7d-098dfc80b359", baseUrl, authToken)
	// total_amt, err := testsetup.GetUsageAndValidateTotalUsage("varsha.karinje@intel.com", resourceInfo, baseUrl, authToken, regionalUrl)
	// if err != nil {
	// 	fmt.Println("Wrong usage reported")
	// }
	// fmt.Println("Overall Usage", testsetup.ProductUsageData)
	// fmt.Println("Total Usage", total_amt)

	//testsetup.RemoteCommandExecution("10.54.69.26", "visaliva", "C:\\Users\\visaliva\\.ssh\\windows.key", "ls -ltr")
	// err := testsetup.ApplyProductsAndVendors()
	// if err != nil {
	// 	fmt.Println("PC Failed ", err)
	// }
	// err := testsetup.RestartService("billing")
	// if err != nil {
	// 	fmt.Println("PC Failed ", err)
	// }

	// testsetup.GetUnappliedCredits("varsha.karinje@intel.com")
	// err := testsetup.ValidateCreditsDeduction("varsha.karinje@intel.com", "vm-spr-tny")
	// if err != nil {
	// 	fmt.Println("Error in credit depletion ", err)
	// }
	// // //Cleanup


	fmt.Println("Waiting for one minute before performing the cleanup  ....\n")
	// time.Sleep(1 * time.Minute)
	// testsetup.DeleteSSHKeys()
	// testsetup.DeleteVNets()
	// testsetup.DeleteInstances()
	testsetup.DeleteCloudAccounts()

}
