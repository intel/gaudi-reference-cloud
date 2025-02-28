package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"io/ioutil"
	"net/http"
	"os"

	"time"

	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var couponData string
var isStandard bool

type CreateCouponStruct struct {
	Amount     int64  `json:"amount"`
	Creator    string `json:"creator"`
	Expires    string `json:"expires"`
	NumUses    int64  `json:"numUses"`
	Start      string `json:"start"`
	IsStandard bool   `json:"isStandard"`
}

// Post Coupon

func Post(url string, data *bytes.Buffer, expected_status_code int, bearer_token string) (string, int) {
	var jsonStr string
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPost, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err1 := http.DefaultClient.Do(request)
	if err1 != nil {
		return jsonStr, resp.StatusCode
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		return jsonStr, resp.StatusCode
	}

	//Convert bytes to String and print
	jsonStr = string(body)
	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusCreated || expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusCreated || expected_status_code == http.StatusOK {
			fmt.Println("Response: ", jsonStr)

		} else {
			//The status is not Created. print the error.
			return jsonStr, resp.StatusCode

		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

// Load the configuration from the provided yaml file.
func LoadConfig(filePath string) (string, error) {
	fmt.Println("Config file path here", filePath)
	configData, err := os.ReadFile(filePath) // if we os.Open returns an error then handle it

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(configData), nil
}

func GetConfigFileData() {
	pwd, _ := os.Getwd()
	couponData, _ = LoadConfig(pwd + "/data.json")
	fmt.Println("Config Data", couponData)

}

func CreateCoupon(couponPayload []byte, endPoint string, token string) (string, int) {
	url := endPoint + "/v1/billing/coupons"
	//jsonPayload, _ := json.Marshal(couponPayload)
	//req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(couponPayload)
	fmt.Println("couponPayload", reqBody)
	fmt.Println("url", url)
	respBody, respCode := Post(url, reqBody, 200, token)
	return respBody, respCode

}

func main() {
	//pwd, _ := os.Getwd()
	// Load config data
	f, _ := os.Create("coupon.txt")

	defer f.Close()

	GetConfigFileData()
	logger.InitializeZapCustomLogger()
	auth.Get_config_file_data("./env.json")
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	os.Setenv("fetch_admin_role_token_from_env", token)
	endPoint := gjson.Get(couponData, "endPoint").String()
	userType := gjson.Get(couponData, "userTier").String()
	noOfCoupons := gjson.Get(couponData, "numberofCoupons").Int()
	Amount := gjson.Get(couponData, "amount").Int()
	expirationtime := gjson.Get(couponData, "expiry").String()
	if userType == "Standard" {
		isStandard = true
	} else {
		isStandard = false
	}

	// Set creation time

	time.Sleep(1 * time.Second)
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(1 * time.Minute)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	createCoupon := CreateCouponStruct{
		Amount:     Amount,
		Creator:    "IDC Admin",
		Expires:    expirationtime,
		Start:      creation_timestamp,
		NumUses:    1,
		IsStandard: isStandard,
	}
	fmt.Fprintf(f, "End Point  : %v\n", endPoint)
	fmt.Fprintf(f, "User Type : %v\n", userType)
	fmt.Fprintf(f, "Number of Coupons : %v\n", noOfCoupons)
	fmt.Fprintf(f, "Coupon Amount : %v\n", Amount)
	fmt.Fprintf(f, "Coupon Expiry : %v\n\n\n", expirationtime)
	fmt.Fprintf(f, "Coupons List:\n")
	fmt.Fprintf(f, "===============\n")

	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	for i := 0; i < int(noOfCoupons); i++ {
		respBody, _ := CreateCoupon(req, endPoint, token)
		couponCode := gjson.Get(respBody, "code").String()
		fmt.Fprintf(f, "%v\n", couponCode)
	}

}
