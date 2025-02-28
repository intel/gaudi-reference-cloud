package financials

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/frisby_client"
	"goFramework/ginkGo/financials/financials_utils"
	"log"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/playwright-community/playwright-go"
	"github.com/tidwall/gjson"
)

var timeout float64 = 60

type AddPaymentMethodResponse struct {
	ErrorCode          int    `json:"error_code"`
	ErrorMsg           string `json:"error_msg"`
	BillingGroupNo     string `json:"billing_group_no"`
	BillingContactInfo []struct {
		PaymentMethodNo int `json:"payment_method_no"`
	} `json:"billing_contact_info"`
	BillingGroupNo2 int `json:"billing_group_no_2"`
}

func assertErrorToNilf(message string, err error) error {
	if err != nil {
		log.Fatalf(message, err)
		return err
	}
	return nil

}

func assertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("%v does not equal %v", actual, expected))
	}
}

// Coupon

func getCredits(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createCredits(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getUnappliedCredits(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateCredits(Credits_api_base_url string, token string, Credits_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(Credits_api_payload), &jsonMap)
	response_status, response_body := createCredits(Credits_api_base_url, token, jsonMap)
	return response_status, response_body
}

func GetCreditsByHistory(Credits_api_base_url string, token string, cloudAccountId string, history string) (int, string) {
	var get_Credits_url = Credits_api_base_url + "?cloudAccountId=" + cloudAccountId + "&history=" + history
	logger.Logf.Infof("get_Credits_url : %s ", get_Credits_url)
	get_response_status, get_response_body := getCredits(get_Credits_url, token)
	logger.Logf.Infof("get_Credits_response", get_response_body)
	return get_response_status, get_response_body
}

func GetCredits(Credits_api_base_url string, token string, cloudAccountId string) (int, string) {
	var get_Credits_url = Credits_api_base_url + "?cloudAccountId=" + cloudAccountId + "&history=true"
	logger.Logf.Infof("get_Credits_url : %s ", get_Credits_url)
	get_response_status, get_response_body := getCredits(get_Credits_url, token)
	return get_response_status, get_response_body
}

func GetCreditsByRegion(Credits_api_base_url string, token string, cloudAccountId string, region string) (int, string) {
	var get_Credits_url = Credits_api_base_url + "?cloudAccountId=" + cloudAccountId + "&regionName=" + region
	logger.Logf.Infof("get_Credits_url : %s ", get_Credits_url)
	get_response_status, get_response_body := getCredits(get_Credits_url, token)
	return get_response_status, get_response_body
}

func GetCreditsByDate(Credits_api_base_url string, token string, cloudAccountId string, startDate string, endDate string) (int, string) {
	var get_Credits_url = Credits_api_base_url + "/invoices?cloudAccountId=" + cloudAccountId + "&searchStart=" + startDate + "&searchEnd=" + endDate
	logger.Logf.Infof("get_Credits_url : %s ", get_Credits_url)
	get_response_status, get_response_body := getCredits(get_Credits_url, token)
	return get_response_status, get_response_body
}

func GetUnappliedCredits(Credits_api_base_url string, token string, cloudAccountId string) (int, string) {
	var get_Credits_url = Credits_api_base_url + "/unapplied" + "?cloudAccountId=" + cloudAccountId
	logger.Logf.Infof("get_Credits_url : %s ", get_Credits_url)
	get_response_status, get_response_body := getUnappliedCredits(get_Credits_url, token)
	return get_response_status, get_response_body
}

//Coupons

func getCoupons(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Coupons response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createCoupon(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func redeemCoupon(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func disableCoupon(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func startScheduler(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func GetCoupons(url string, token string) (int, string) {
	var get_Coupon_url = url
	fmt.Println("get_Coupon_url", get_Coupon_url)
	get_response_status, get_response_body := getCoupons(get_Coupon_url, token)
	return get_response_status, get_response_body
}

func GetCouponsByCode(url string, token string, code string) (int, string) {
	var get_Coupon_url = url + "?code=" + code
	fmt.Println("get_Coupon_url", get_Coupon_url)
	get_response_status, get_response_body := getCoupons(get_Coupon_url, token)
	return get_response_status, get_response_body
}

func CreateCoupon(Coupon_api_base_url string, token string, Coupon_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(Coupon_api_payload), &jsonMap)
	response_status, response_body := createCoupon(Coupon_api_base_url, token, jsonMap)
	return response_status, response_body
}

func RedeemCoupon(Coupon_api_base_url string, token string, Coupon_api_payload string) (int, string) {
	fmt.Println("Waiting for coupon to be available.")
	time.Sleep(1 * time.Minute)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(Coupon_api_payload), &jsonMap)
	response_status, response_body := redeemCoupon(Coupon_api_base_url, token, jsonMap)
	return response_status, response_body
}

func DisableCoupon(Coupon_api_base_url string, token string, Coupon_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(Coupon_api_payload), &jsonMap)
	response_status, response_body := disableCoupon(Coupon_api_base_url, token, jsonMap)
	return response_status, response_body
}

//Usage

func getUsage(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Coupons response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getNotifications(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Coupons response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getOptions(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Coupons response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetUsage(url string, token string) (int, string) {
	var get_Usage_url = url
	fmt.Println("get_Usage_url", get_Usage_url)
	get_response_status, get_response_body := getUsage(get_Usage_url, token)
	return get_response_status, get_response_body
}

func GetUsageTimeRange(url string, token string, cloudAccountId string, starttime string, endtime string) (int, string) {
	var get_Usage_url = url + "/usages" + "?cloudAccountId=" + cloudAccountId + "&searchStart=" + starttime + "&searchEnd=" + endtime
	fmt.Println("get_Usage_url", get_Usage_url)
	get_response_status, get_response_body := getUsage(get_Usage_url, token)
	return get_response_status, get_response_body
}

func StartScheduler(url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := startScheduler(url, token, jsonMap)
	return response_status, response_body
}

func GetNotificationsShortPoll(url string, token string, cloudAccountId string) (int, string) {
	var get_notifications_url = url + "/v1/billing/events/all" + "?cloudAccountId=" + cloudAccountId
	fmt.Println("get_notifications_url", get_notifications_url)
	get_response_status, get_response_body := getNotifications(get_notifications_url, token)
	return get_response_status, get_response_body
}

func GetBillingOptions(url string, token string, cloudAccountId string) (int, string) {
	var get_options_url = url + "/v1/billing/options" + "?cloudAccountId=" + cloudAccountId
	fmt.Println("get_options_url", get_options_url)
	get_response_status, get_response_body := getOptions(get_options_url, token)
	return get_response_status, get_response_body
}

func EnrollPremiumUserWithCreditCard(creditCardPayload string, userName string, password string, consoleurl string, replaceUrl string) error {
	time.Sleep(2 * time.Minute)
	pw, err := playwright.Run()
	enrollurl := consoleurl + "/premium"
	assertErr := assertErrorToNilf("could not launch playwright: %w", err)
	if assertErr != nil {
		return assertErr
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})

	assertErr = assertErrorToNilf("could not launch Browser: %w", err)
	context, err := browser.NewContext()
	assertErr = assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErr = assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(enrollurl, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErr = assertErrorToNilf("could not goto: %w", err)
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(10 * time.Second)
	assertErr = assertErrorToNilf("could not click: %v", page.Click("#signInName"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type("#signInName", userName))
	assertErr = assertErrorToNilf("could not press: %v", page.Click("#continue"))
	time.Sleep(10 * time.Second)
	assertErr = assertErrorToNilf("could not click: %v", page.Click("#password"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type("#password", password))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not press: %v", page.Click("#continue"))
	time.Sleep(2 * time.Minute)

	// Test reaches add credit card page, enter credit card details here
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".creditCard-all-images"))
	if assertErr != nil {
		return assertErr
	}
	//assertErr = assertErrorToNilf("could not click: %v", page.Click("#CardnumberInput]"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".creditCard-all-images", gjson.Get(creditCardPayload, "CardnumberInput").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-2:nth-child(1) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-2:nth-child(1) .form-control", gjson.Get(creditCardPayload, "MonthInput").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-2:nth-child(2) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-2:nth-child(2) .form-control", gjson.Get(creditCardPayload, "Year").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".offset-md-1 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".offset-md-1 .form-control", gjson.Get(creditCardPayload, "Cvc").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(9) > .col-xl-4 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(9) > .col-xl-4 .form-control", gjson.Get(creditCardPayload, "FirstName").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(9) > .col-xl-5 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(9) > .col-xl-5 .form-control", gjson.Get(creditCardPayload, "LastName").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".mb-3 > .form-group .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".mb-3 > .form-group .form-control", gjson.Get(creditCardPayload, "Email").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(11) > .col-xl-5 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(11) > .col-xl-5 .form-control", gjson.Get(creditCardPayload, "Phone").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-9 .form-select"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-9 .form-select", gjson.Get(creditCardPayload, "Country").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-9 .form-select"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(13) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(13) .form-control", gjson.Get(creditCardPayload, "AddressLine").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(15) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(15) .form-control", gjson.Get(creditCardPayload, "City").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-4 .form-select"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-4 .form-select", gjson.Get(creditCardPayload, "State").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-4 .form-select"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(16) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(16) .form-control", gjson.Get(creditCardPayload, "ZipCode").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".me-3"))
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(3 * time.Minute)

	// Get Payment methods and validate credit card is added

	pageUrl := page.URL()
	getConsoleDom, _ := url.Parse(consoleurl)
	fmt.Println("pageUrl", pageUrl)
	u, _ := url.Parse(pageUrl)

	u.Host = getConsoleDom.Hostname()
	logger.Log.Info("Console new URL : " + u.String())
	pageUrl = strings.Replace(pageUrl, replaceUrl, consoleurl, 1)
	u, _ = url.Parse(pageUrl)
	_, err = page.Goto(u.String(), playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErr = assertErrorToNilf("could not goto: %w", err)
	pageUrl = page.URL()
	if !strings.Contains(pageUrl, consoleurl) {
		return fmt.Errorf("failed to add credit card to the premium account")
	}
	time.Sleep(2 * time.Minute)

	logger.Log.Info("Successfully added credit card to the Users account")

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

	return nil
}

func ChangeCreditCard(creditCardPayload string, userName string, password string, consoleurl string, replaceUrl string) error {
	logger.Log.Info("Starting to Change credit card in user account")
	pw, err := playwright.Run()
	enrollurl := consoleurl + "/billing/managePaymentMethods"
	assertErr := assertErrorToNilf("could not launch playwright: %w", err)
	if assertErr != nil {
		return assertErr
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})

	assertErr = assertErrorToNilf("could not launch Browser: %w", err)
	if assertErr != nil {
		return assertErr
	}
	context, err := browser.NewContext()
	assertErr = assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErr = assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(enrollurl, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErr = assertErrorToNilf("could not goto: %w", err)
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(10 * time.Second)
	assertErr = assertErrorToNilf("could not click: %v", page.Click("#signInName"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type("#signInName", userName))
	assertErr = assertErrorToNilf("could not press: %v", page.Click("#continue"))
	time.Sleep(10 * time.Second)
	assertErr = assertErrorToNilf("could not click: %v", page.Click("#password"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type("#password", password))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not press: %v", page.Click("#continue"))
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(2 * time.Minute)

	// assertErr = assertErrorToNilf("could not click: %v", page.Click("svg:nth-child(2)"))
	// assertErr = assertErrorToNilf("could not click: %v", page.Click(".list-unstyled > .dropdown-item:nth-child(4)"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click("[intc-id=btn-managepayment-changecard]"))
	if assertErr != nil {
		return assertErr
	}

	// Test reaches add credit card page, enter credit card details here
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".creditCard-all-images"))
	if assertErr != nil {
		return assertErr
	}
	//assertErr = assertErrorToNilf("could not click: %v", page.Click("#CardnumberInput]"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".creditCard-all-images", gjson.Get(creditCardPayload, "CardnumberInput").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-2:nth-child(1) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-2:nth-child(1) .form-control", gjson.Get(creditCardPayload, "MonthInput").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-2:nth-child(2) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-2:nth-child(2) .form-control", gjson.Get(creditCardPayload, "Year").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".offset-md-1 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".offset-md-1 .form-control", gjson.Get(creditCardPayload, "Cvc").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(6) > .col-xl-4 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(6) > .col-xl-4 .form-control", gjson.Get(creditCardPayload, "FirstName").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(6) > .col-xl-5 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(6) > .col-xl-5 .form-control", gjson.Get(creditCardPayload, "LastName").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".mb-3 > .form-group .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".mb-3 > .form-group .form-control", gjson.Get(creditCardPayload, "Email").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(8) > .col-xl-5 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(8) > .col-xl-5 .form-control", gjson.Get(creditCardPayload, "Phone").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-9 .form-select"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-9 .form-select", gjson.Get(creditCardPayload, "Country").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-9 .form-select"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(10) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(10) .form-control", gjson.Get(creditCardPayload, "AddressLine").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(12) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(12) .form-control", gjson.Get(creditCardPayload, "City").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-4 .form-select"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-4 .form-select", gjson.Get(creditCardPayload, "State").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-4 .form-select"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(13) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(13) .form-control", gjson.Get(creditCardPayload, "ZipCode").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".btn-primary"))
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(3 * time.Minute)

	// Get Payment methods and validate credit card is added

	pageUrl := page.URL()
	getConsoleDom, _ := url.Parse(consoleurl)
	fmt.Println("pageUrl", pageUrl)
	u, _ := url.Parse(pageUrl)

	u.Host = getConsoleDom.Hostname()
	logger.Log.Info("Console new URL : " + u.String())
	pageUrl = strings.Replace(pageUrl, replaceUrl, consoleurl, 1)
	u, _ = url.Parse(pageUrl)
	_, err = page.Goto(u.String(), playwright.PageGotoOptions{Timeout: playwright.Float(0)})

	assertErr = assertErrorToNilf("could not goto: %w", err)
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(1 * time.Minute)
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	return nil
}

// Invoice

func getInvoice(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

func GetInvoice(url string, token string, cloudAccountId string) (int, string) {
	// func GetInvoice(url string, token string, cloudAccountId string, starttime string, endtime string) (int, string) {
	// starttime := "2023-07-01T09:59:40.062Z"
	// endtime := "2023-07-30T09:59:40.062Z"
	// var get_Invoice_url = url + "?cloudAccountId=" + cloudAccountId + "&searchStart=" + starttime + "&searchEnd=" + endtime
	var get_Invoice_url = url + "?cloudAccountId=" + cloudAccountId
	fmt.Println("get_Usage_url", get_Invoice_url)
	get_response_status, get_response_body := getInvoice(get_Invoice_url, token)
	return get_response_status, get_response_body
}

// Invoice with invoiceId

func GetInvoicewithInvoiceId(url string, token string, cloudAccountId string, invoiceId string) (int, string) {
	var get_Invoice_url = url + "?cloudAccountId=" + cloudAccountId + "&invoiceId=" + invoiceId
	fmt.Println("get_Usage_url", get_Invoice_url)
	get_response_status, get_response_body := getInvoice(get_Invoice_url, token)
	return get_response_status, get_response_body
}

// Instance Termination

func getDeactivationList(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

func GetInstanceDeactivationList(url string, token string) (int, string) {
	var get_deactivation_url = url + "/v1/billing/instances/deactivate"
	fmt.Println("get_Usage_url", get_deactivation_url)
	get_response_status, get_response_body := getDeactivationList(get_deactivation_url, token)
	return get_response_status, get_response_body
}

func GetStaaSDeactivationList(url string, token string) (int, string) {
	var get_deactivation_url = url + "/v1/billing/service/deactivate"
	fmt.Println("get_Usage_url", get_deactivation_url)
	get_response_status, get_response_body := getDeactivationList(get_deactivation_url, token)
	return get_response_status, get_response_body
}

func GetProductRate(url string, token string, cloudAccId string, product string) (int, string) {
	var get_deactivation_url = url + "/v1/billing/rates?cloudAccountId=" + cloudAccId + "&productId=" + product
	fmt.Println("get_Usage_url", get_deactivation_url)
	get_response_status, get_response_body := getDeactivationList(get_deactivation_url, token)
	return get_response_status, get_response_body
}

func CheckCloudAccInDeactivationList(url string, token string, cloudAccountId string) (int, bool) {
	get_response_status, get_response_body := GetInstanceDeactivationList(url, token)
	return get_response_status, strings.Contains(get_response_body, cloudAccountId)

}

// Standard to premium upgrade libraries

func upgradewithCoupon(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func UpgradeWithCoupon(Coupon_api_base_url string, token string, Coupon_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(Coupon_api_payload), &jsonMap)
	fmt.Println("Waiting for coupon to be available.")
	time.Sleep(1 * time.Minute)
	fmt.Println("Premium upgrade payload", Coupon_api_payload)
	response_status, response_body := upgradewithCoupon(Coupon_api_base_url, token, jsonMap)
	return response_status, response_body
}

func UpgradeThroughCreditCard(creditCardPayload string, userName string, password string, consoleurl string, replaceUrl string) error {
	logger.Log.Info("Starting to Upgrade using credit card")
	pw, err := playwright.Run()
	enrollurl := consoleurl + "/upgradeaccount"
	assertErr := assertErrorToNilf("could not launch playwright: %w", err)
	if assertErr != nil {
		return assertErr
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})

	assertErr = assertErrorToNilf("could not launch Browser: %w", err)
	if assertErr != nil {
		return assertErr
	}
	context, err := browser.NewContext()
	assertErr = assertErrorToNilf("could not create context: %w", err)
	if assertErr != nil {
		return assertErr
	}
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErr = assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(enrollurl, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not goto: %w", err)
	if assertErr != nil {
		return assertErr
	}
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(10 * time.Second)
	assertErr = assertErrorToNilf("could not click: %v", page.Click("#signInName"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type("#signInName", userName))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not press: %v", page.Click("#continue"))
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(10 * time.Second)
	assertErr = assertErrorToNilf("could not click: %v", page.Click("#password"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type("#password", password))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not press: %v", page.Click("#continue"))
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(2 * time.Minute)

	// Test reaches add credit card page, enter credit card details here
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".creditCard-all-images"))
	if assertErr != nil {
		return assertErr
	}
	//assertErr = assertErrorToNilf("could not click: %v", page.Click("#CardnumberInput]"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".creditCard-all-images", gjson.Get(creditCardPayload, "CardnumberInput").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-2:nth-child(1) .form-control"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-2:nth-child(1) .form-control", gjson.Get(creditCardPayload, "MonthInput").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-2:nth-child(2) .form-control"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-2:nth-child(2) .form-control", gjson.Get(creditCardPayload, "Year").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".offset-md-1 .form-control"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".offset-md-1 .form-control", gjson.Get(creditCardPayload, "Cvc").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(9) > .col-xl-4 .form-control"))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(9) > .col-xl-4 .form-control", gjson.Get(creditCardPayload, "FirstName").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(9) > .col-xl-5 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(9) > .col-xl-5 .form-control", gjson.Get(creditCardPayload, "LastName").String()))
	if assertErr != nil {
		return assertErr
	}
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".mb-3 > .form-group .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".mb-3 > .form-group .form-control", gjson.Get(creditCardPayload, "Email").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(11) > .col-xl-5 .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(11) > .col-xl-5 .form-control", gjson.Get(creditCardPayload, "Phone").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-9 .form-select"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-9 .form-select", gjson.Get(creditCardPayload, "Country").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-9 .form-select"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(13) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(13) .form-control", gjson.Get(creditCardPayload, "AddressLine").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(15) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(15) .form-control", gjson.Get(creditCardPayload, "City").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-4 .form-select"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".col-xl-4 .form-select", gjson.Get(creditCardPayload, "State").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".col-xl-4 .form-select"))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".row:nth-child(16) .form-control"))
	assertErr = assertErrorToNilf("could not type: %v", page.Type(".row:nth-child(16) .form-control", gjson.Get(creditCardPayload, "ZipCode").String()))
	assertErr = assertErrorToNilf("could not click: %v", page.Click(".me-3"))
	if assertErr != nil {
		return assertErr
	}
	time.Sleep(3 * time.Minute)

	// Get Payment methods and validate credit card is added

	pageUrl := page.URL()
	getConsoleDom, _ := url.Parse(consoleurl)
	fmt.Println("pageUrl", pageUrl)
	u, _ := url.Parse(pageUrl)

	u.Host = getConsoleDom.Hostname()
	logger.Log.Info("Console new URL : " + u.String())
	pageUrl = strings.Replace(pageUrl, replaceUrl, consoleurl, 1)
	u, _ = url.Parse(pageUrl)
	_, err = page.Goto(u.String(), playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErr = assertErrorToNilf("could not goto: %w", err)
	if assertErr != nil {
		return assertErr
	}

	time.Sleep(1 * time.Minute)
	page.Reload()

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	return nil
}

func CreditMigrate(base_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createCoupon(base_url, token, jsonMap)
	return response_status, response_body
}

// Credit Card

func prePaymentMethod(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "OPTIONS API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func PrePaymentMethod(base_url string, token string, cloudAccId string) (int, string) {
	base_url = base_url + "/v1/billing/payments/prepayment?cloudAccountId=" + cloudAccId
	response_status, response_body := prePaymentMethod(base_url, token)
	return response_status, response_body
}

func postPaymentMethod(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "OPTIONS API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func UpgradeAccount(Coupon_api_base_url string, token string, Coupon_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(Coupon_api_payload), &jsonMap)
	response_status, response_body := upgradeAccount(Coupon_api_base_url, token, jsonMap)
	return response_status, response_body
}

func upgradeAccount(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func PostpaymentMethod(base_url string, token string, cloudAccId string, paymentMethodNum string) (int, string) {
	base_url = base_url + "/v1/billing/payments/postpayment?cloudAccountId=" + cloudAccId + "&primaryPaymentMethodNo=" + paymentMethodNum
	response_status, response_body := postPaymentMethod(base_url, token)
	return response_status, response_body
}

func AddCreditCardToAccount(base_url string, cloudAccId string, token string, creditCardDetails CreditCardDetails, ccType string, paymentMethodNum string) error {
	_, details := GetAriaAccountDetails(cloudAccId)
	logger.Logf.Infof("Details ", details)
	payMethodType := 1
	// Add Prepayment method
	rescode, resbody := PrePaymentMethod(base_url, token, cloudAccId)
	if rescode != 200 {
		return fmt.Errorf("prepayment method for cloud acc id %s failed with error %s", cloudAccId, resbody)
	}
	logger.Logf.Infof(" Prepayment response code", rescode)
	logger.Logf.Infof(" Prepayment response body", resbody)

	clientBillingGroupId := "idc." + cloudAccId + ".billing_group"

	clientPaymentMethodId := uuid.New().String()
	ariacode, ariaoutput := AddAccountPaymentMethod(cloudAccId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	data := AddPaymentMethodResponse{}
	err := json.Unmarshal([]byte(ariaoutput), &data)
	if err != nil {
		return err
	}
	if data.ErrorCode != 0 {
		return fmt.Errorf("add payment method failed for cloud account %s with error %s and error code %d", cloudAccId, data.ErrorMsg, ariacode)
	}

	if data.ErrorMsg != "OK" {
		return fmt.Errorf("add payment method failed for cloud account %s with error %s and error code %d", cloudAccId, data.ErrorMsg, ariacode)
	}

	rescode, resbody = PostpaymentMethod(base_url, token, cloudAccId, paymentMethodNum)
	if rescode != 200 {
		return fmt.Errorf("postpayment method for cloud acc id %s failed with error %s", cloudAccId, resbody)
	}

	logger.Logf.Infof(" Postpayment response code", rescode)
	logger.Logf.Infof(" Postpayment response body", resbody)

	time.Sleep(10 * time.Second)
	// Validate Credit card being added

	resCode, resBody := GetBillingOptions(base_url, token, cloudAccId)
	logger.Log.Info("resBody" + resBody)
	logger.Logf.Infof("bilingOptions", resCode)
	s := fmt.Sprint(creditCardDetails.CCNumber)
	expiration := fmt.Sprint(creditCardDetails.CCExpireMonth) + "/" + fmt.Sprint(creditCardDetails.CCExpireYear)
	suffix := s[len(s)-4:]
	if resCode != 200 {
		return fmt.Errorf("validating billing options for cloud account %s failed with response code %d and reponse body %s", cloudAccId, resCode, resBody)
	}

	if gjson.Get(resBody, "creditCard.suffix").String() != suffix || gjson.Get(resBody, "creditCard.expiration").String() != expiration ||
		gjson.Get(resBody, "creditCard.type").String() != ccType || gjson.Get(resBody, "paymentType").String() != "PAYMENT_CREDIT_CARD" ||
		gjson.Get(resBody, "cloudAccountId").String() != cloudAccId {
		return fmt.Errorf("validating billing options for cloud account %s failed with response code %d and reponse body %s", cloudAccId, resCode, resBody)
	}

	return nil

}

func UpgradeAccountWithCreditCard(base_url string, cloudAccId string, token string, creditCardDetails CreditCardDetails, ccType string, paymentMethodNum string, cloudAccountUpgradeToType string) error {

	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	upgrade_url := base_url + "/v1/cloudaccounts/upgrade"
	Coupon_api_payload := financials_utils.EnrichUpgradeWithoutCouponPayload(financials_utils.GetUpgradeWithoutCouponPayload(), cloudAccId, cloudAccountUpgradeToType)
	logger.Logf.Infof("Upgrade with credit card  payload", Coupon_api_payload)
	response_status, responseBody := UpgradeAccount(upgrade_url, token, Coupon_api_payload)
	if response_status != 200 {
		return fmt.Errorf("failed to upgrade cloud account : %s, error : %s", cloudAccId, responseBody)
	}
	time.Sleep(20 * time.Second)
	_, details := GetAriaAccountDetails(cloudAccId)
	logger.Logf.Infof("Details ", details)
	payMethodType := 1
	// Add Prepayment method
	rescode, resbody := PrePaymentMethod(base_url, token, cloudAccId)
	if rescode != 200 {
		return fmt.Errorf("prepayment method for cloud acc id %s failed with error %s", cloudAccId, resbody)
	}
	logger.Logf.Infof(" Prepayment response code", rescode)
	logger.Logf.Infof(" Prepayment response body", resbody)

	clientBillingGroupId := "idc." + cloudAccId + ".billing_group"

	clientPaymentMethodId := uuid.New().String()
	ariacode, ariaoutput := AddAccountPaymentMethod(cloudAccId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	data := AddPaymentMethodResponse{}
	err := json.Unmarshal([]byte(ariaoutput), &data)
	if err != nil {
		return err
	}
	if data.ErrorCode != 0 {
		return fmt.Errorf("add payment method failed for cloud account %s with error %s and error code %d", cloudAccId, data.ErrorMsg, ariacode)
	}

	if data.ErrorMsg != "OK" {
		return fmt.Errorf("add payment method failed for cloud account %s with error %s and error code %d", cloudAccId, data.ErrorMsg, ariacode)
	}

	// Upgrade account

	rescode, resbody = PostpaymentMethod(base_url, token, cloudAccId, paymentMethodNum)
	if rescode != 200 {
		return fmt.Errorf("postpayment method for cloud acc id %s failed with error %s", cloudAccId, resbody)
	}

	logger.Logf.Infof(" Postpayment response code", rescode)
	logger.Logf.Infof(" Postpayment response body", resbody)

	time.Sleep(10 * time.Second)
	// Validate Credit card being added

	resCode, resBody := GetBillingOptions(base_url, token, cloudAccId)
	logger.Log.Info("resBody" + resBody)
	logger.Logf.Infof("bilingOptions", resCode)
	s := fmt.Sprint(creditCardDetails.CCNumber)
	expiration := fmt.Sprint(creditCardDetails.CCExpireMonth) + "/" + fmt.Sprint(creditCardDetails.CCExpireYear)
	suffix := s[len(s)-4:]
	if resCode != 200 {
		return fmt.Errorf("validating billing options for cloud account %s failed with response code %d and reponse body %s", cloudAccId, resCode, resBody)
	}

	if gjson.Get(resBody, "creditCard.suffix").String() != suffix || gjson.Get(resBody, "creditCard.expiration").String() != expiration ||
		gjson.Get(resBody, "creditCard.type").String() != ccType || gjson.Get(resBody, "paymentType").String() != "PAYMENT_CREDIT_CARD" ||
		gjson.Get(resBody, "cloudAccountId").String() != cloudAccId {
		return fmt.Errorf("validating billing options for cloud account %s failed with response code %d and reponse body %s", cloudAccId, resCode, resBody)
	}

	return nil

}

//Maas libraries

func CreateMaasUsageRecords(base_url string, token string, api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := createMaasUsageRecords(base_url, token, jsonMap)
	return response_status, response_body
}

func createMaasUsageRecords(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func SearchMaasUsageRecords(base_url string, token string, api_payload string) (int, string) {
	logger.Logf.Infof("Search Url : %s", base_url)
	logger.Logf.Infof("Search Payload : %s", api_payload)

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := searchMaasUsageRecords(base_url, token, jsonMap)
	return response_status, response_body
}

func searchMaasUsageRecords(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	logger.Logf.Infof("Search Response : %s", responseBody)
	//Coupon response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}
