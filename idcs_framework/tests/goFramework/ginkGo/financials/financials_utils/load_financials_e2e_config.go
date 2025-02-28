package financials_utils

import (
	"fmt"
	"goFramework/ginkGo/compute/compute_utils"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tidwall/gjson"
)

var ariaUrl string
var ariaClientId string
var ariaAuth string
var createCouponPayload string
var createCouponPayloadStandard string
var redeemCouponPayload string
var disableCouponPayload string
var startSchedulerPayload string
var creditcardPayload1 string
var region1Name string
var region2Name string
var usageTime time.Duration
var usageTimeEnterprise int
var createInvitation string
var resendInvitation string
var revokeInvitation string
var validateInviteCode string
var rejectInvitation string
var removeMember string
var createOTP string
var verifyOTP string
var resendOTP string
var cloudaccount_pg_password string
var expirationDate string
var invitationMessage string
var mailSlurpKey string
var inboxIdEnterprise string
var inboxIdPremium string
var inboxIdStandard string
var upgradewithCoupon string
var visaCardData string
var masterCardData string
var discoverCardData string
var amexCardData string
var creditMigrate string
var ariaCertPath string
var ariaKeyPath string
var upgradewithoutCoupon string
var globalgRPCUrl string
var cognitoUserPool string
var cognitoClientId string
var cognitoUrl string
var cognitoClientSecret string
var intelsmlvmrate float64
var intelmedvmrate float64
var intellrgvmrate float64
var intelbmvmrate float64
var intelgaudi2rate float64
var intelimagerate float64
var inteltextrate float64
var credit_reason string
var pg_password string
var create_maas_usage string
var premiumsmlvmrate float64
var premiummedvmrate float64
var premiumlrgvmrate float64
var premiumbmvmrate float64
var premiumgaudi2rate float64
var premiumimagerate float64
var premiumtextrate float64
var standardsmlvmrate float64
var standardmedvmrate float64
var standardlrgvmrate float64
var standardbmvmrate float64
var standardgaudi2rate float64
var standardimagerate float64
var standardtextrate float64
var createCreditsPayload string

func LoadE2EConfig(filepath string, filename string) {
	configData, _ := compute_utils.ConvertFileToString(filepath, filename)
	//fmt.Println("config data is " + configData)
	ariaUrl = gjson.Get(configData, "ariaBaseUrl").String()
	ariaClientId = gjson.Get(configData, "ariaConfig.client_no").String()
	ariaAuth = gjson.Get(configData, "ariaConfig.auth_key").String()
	createCouponPayload = gjson.Get(configData, "createCoupon").String()
	createCouponPayloadStandard = gjson.Get(configData, "createStandardCoupon").String()
	redeemCouponPayload = gjson.Get(configData, "redeemCoupon").String()
	createCreditsPayload = gjson.Get(configData, "createCredits").String()
	disableCouponPayload = gjson.Get(configData, "Coupon").String()
	startSchedulerPayload = gjson.Get(configData, "startScheduler").String()
	region1Name = gjson.Get(configData, "region1Name").String()
	region2Name = gjson.Get(configData, "region1Name").String()
	creditcardPayload1 = gjson.Get(configData, "creditcardPayload1").String()
	creditcardPayload1 = gjson.Get(configData, "createInvitation").String()
	createInvitation = gjson.Get(configData, "resendInvitation").String()
	resendInvitation = gjson.Get(configData, "creditcardPayload1").String()
	revokeInvitation = gjson.Get(configData, "revokeInvitation").String()
	validateInviteCode = gjson.Get(configData, "validateInviteCode").String()
	rejectInvitation = gjson.Get(configData, "rejectInvitation").String()
	removeMember = gjson.Get(configData, "removeMember").String()
	createOTP = gjson.Get(configData, "createOTP").String()
	verifyOTP = gjson.Get(configData, "verifyOTP").String()
	resendOTP = gjson.Get(configData, "resendOTP").String()
	cloudaccount_pg_password = gjson.Get(configData, "cloudaccount_pg_password").String()
	usageTime = 1
	usageTimeEnterprise = int(gjson.Get(configData, "usageTimeEnterprise").Int())
	expirationDate = gjson.Get(configData, "expirationDate").String()
	invitationMessage = gjson.Get(configData, "invitationMessage").String()
	mailSlurpKey = gjson.Get(configData, "mailSlurpKey").String()
	inboxIdEnterprise = gjson.Get(configData, "inboxIdEnterprise").String()
	inboxIdPremium = gjson.Get(configData, "inboxIdPremium").String()
	inboxIdStandard = gjson.Get(configData, "inboxIdStandard").String()
	upgradewithCoupon = gjson.Get(configData, "upgradeWithCoupon").String()
	upgradewithoutCoupon = gjson.Get(configData, "upgradeWithoutCoupon").String()
	visaCardData = gjson.Get(configData, "creditCards.visaCard").String()
	masterCardData = gjson.Get(configData, "creditCards.masterCard").String()
	discoverCardData = gjson.Get(configData, "creditCards.discoverCard").String()
	amexCardData = gjson.Get(configData, "creditCards.amexCard").String()
	creditMigrate = gjson.Get(configData, "creditMigrate").String()
	globalgRPCUrl = gjson.Get(configData, "globalgRPCUrl").String()
	cognitoUserPool = gjson.Get(configData, "cognitoUserPool").String()
	cognitoClientId = gjson.Get(configData, "cognitoClientId").String()
	cognitoUrl = gjson.Get(configData, "cognitoUrl").String()
	cognitoClientSecret = gjson.Get(configData, "cognitoClientSecret").String()
	intelsmlvmrate = gjson.Get(configData, "rates.intel.sml").Float()
	intelmedvmrate = gjson.Get(configData, "rates.intel.med").Float()
	intellrgvmrate = gjson.Get(configData, "rates.intel.lrg").Float()
	intelbmvmrate = gjson.Get(configData, "rates.intel.bm").Float()
	intelgaudi2rate = gjson.Get(configData, "rates.intel.gaudi2").Float()
	inteltextrate = gjson.Get(configData, "rates.intel.text").Float()
	intelimagerate = gjson.Get(configData, "rates.intel.image").Float()
	credit_reason = gjson.Get(configData, "credit_reason").String()
	pg_password = gjson.Get(configData, "pg_password").String()
	create_maas_usage = gjson.Get(configData, "create_maas_usage").String()
	premiumsmlvmrate = gjson.Get(configData, "rates.premium.sml").Float()
	premiummedvmrate = gjson.Get(configData, "rates.premium.med").Float()
	premiumlrgvmrate = gjson.Get(configData, "rates.premium.lrg").Float()
	premiumbmvmrate = gjson.Get(configData, "rates.premium.bm").Float()
	premiumgaudi2rate = gjson.Get(configData, "rates.premium.gaudi2").Float()
	premiumtextrate = gjson.Get(configData, "rates.premium.text").Float()
	premiumimagerate = gjson.Get(configData, "rates.premium.image").Float()
	standardsmlvmrate = gjson.Get(configData, "rates.standard.sml").Float()
	standardmedvmrate = gjson.Get(configData, "rates.standard.med").Float()
	standardlrgvmrate = gjson.Get(configData, "rates.standard.lrg").Float()
	standardbmvmrate = gjson.Get(configData, "rates.standard.bm").Float()
	standardgaudi2rate = gjson.Get(configData, "rates.standard.gaudi2").Float()
	standardtextrate = gjson.Get(configData, "rates.standard.text").Float()
	standardimagerate = gjson.Get(configData, "rates.standard.image").Float()
}

func GetAriaBaseUrl() string {
	return ariaUrl
}

func GetCreditReason() string {
	return credit_reason
}

func GetUsageTime() time.Duration {
	return usageTime
}

func GetUsageTimeEnterprise() time.Duration {
	return time.Duration(usageTimeEnterprise)
}

func GetAriaClientNo() string {
	return ariaClientId
}

func GetariaAuthKey() string {
	return ariaAuth
}

func GetCreateCouponPayload() string {
	return createCouponPayload
}

func GetCreateCouponPayloadStandard() string {
	return createCouponPayloadStandard
}

func GetDisableCouponPayload() string {
	return disableCouponPayload
}

func GetRedeemCouponPayload() string {
	return redeemCouponPayload
}

func GetCreateCreditPayload() string {
	return createCreditsPayload
}

func GetStartSchedulerPayload() string {
	return startSchedulerPayload
}

func GetCreditCardPayload() string {
	return creditcardPayload1
}

func GetRegion1Name() string {
	return region1Name
}

func GetRegion2Name() string {
	return region2Name
}

func GetCreateInvitationPayload() string {
	return createInvitation
}

func GetResendInvitationPayload() string {
	return resendInvitation
}

func GetValidateInviteCodePayload() string {
	return validateInviteCode
}

func GetRejectInvitationPayload() string {
	return rejectInvitation
}

func GetRemoveMemberPayload() string {
	return removeMember
}

func GetCreateOTPPayload() string {
	return createOTP
}

func GetVerifyOTPPayload() string {
	return verifyOTP
}

func GetResendOTPPayload() string {
	return resendOTP
}

func GetCloudAccDBPAssword() string {
	if os.Getenv("PG_PASSWORD") != "" {
		cloudaccount_pg_password = os.Getenv("PG_PASSWORD")
	}

	fmt.Println(cloudaccount_pg_password)
	return cloudaccount_pg_password
}

func GetCognitoClientId() string {
	if os.Getenv("COGNITO_CLIEND_ID") != "" {
		cognitoClientId = os.Getenv("COGNITO_CLIEND_ID")
	}

	fmt.Println(cognitoClientId)
	return cognitoClientId
}

func GetCognitoUrl() string {
	return cognitoUrl
}

func GetgRPCGlobalUrl() string {
	return globalgRPCUrl
}

func GetcognitoUserPool() string {
	return cognitoUserPool
}

func GetcognitoClientSecret() string {
	return cognitoClientSecret
}

func GetExpirationDate() string {
	return expirationDate
}

func GetInvitationMessage() string {
	return invitationMessage
}

func GetMailSlurpKey() string {
	return mailSlurpKey
}

func GetInboxIdEnterprise() string {
	return inboxIdEnterprise
}

func GetInboxIdPremium() string {
	return inboxIdPremium
}

func GetInboxIdStandard() string {
	return inboxIdStandard
}

func GetMasterCardPayload() string {
	return masterCardData
}

func GetVisaCardPayload() string {
	return visaCardData
}

func GetDiscoverCardPayload() string {
	return discoverCardData
}

func GetAmexCardPayload() string {
	return amexCardData
}

func GetUpgradeCouponPayload() string {
	return upgradewithCoupon
}

func GetUpgradeWithoutCouponPayload() string {
	return upgradewithoutCoupon
}

func GetCreditMigratePayload() string {
	return creditMigrate
}

func GetIntelSmlVmRate() float64 {
	return intelsmlvmrate
}

func GetIntelMedVmRate() float64 {
	return intelmedvmrate
}

func GetIntelLrgVmRate() float64 {
	return intellrgvmrate
}

func GetIntelBMRate() float64 {
	return intelbmvmrate
}

func GetIntelGaudi2Rate() float64 {
	return intelgaudi2rate
}

func GetPremiumSmlVmRate() float64 {
	return premiumsmlvmrate
}

func GetPremiumMedVmRate() float64 {
	return premiummedvmrate
}

func GetPremiumLrgVmRate() float64 {
	return premiumlrgvmrate
}

func GetPremiumBMRate() float64 {
	return premiumbmvmrate
}

func GetPremiumGaudi2Rate() float64 {
	return premiumgaudi2rate
}

func GetStandardSmlVmRate() float64 {
	return standardsmlvmrate
}

func GetStandardMedVmRate() float64 {
	return standardmedvmrate
}

func GetStandardLrgVmRate() float64 {
	return standardlrgvmrate
}

func GetStandardBMRate() float64 {
	return standardbmvmrate
}

func GetStandardGaudi2Rate() float64 {
	return standardgaudi2rate
}

func GetIntelTextRate() float64 {
	return inteltextrate
}

func GetIntelImageRate() float64 {
	return intelimagerate
}

func GetPremiumTextRate() float64 {
	return premiumtextrate
}

func GetPremiumImageRate() float64 {
	return premiumimagerate
}

func GetStandardTextRate() float64 {
	return standardtextrate
}

func GetStandardImageRate() float64 {
	return standardimagerate
}

func GetCreateMaasUsagePayload() string {
	return create_maas_usage
}

func GetPGPassword() string {
	if os.Getenv("PG_PASSWORD") != "" {
		cloudaccount_pg_password = os.Getenv("PG_PASSWORD")
	}

	fmt.Println(cloudaccount_pg_password)
	return cloudaccount_pg_password
}

func GetAriaCertKeyFilePath() (string, string) {
	//wd, _ := os.Getwd()
	_, filename, _, _ := runtime.Caller(1)
	wd := path.Dir(filename)

	wd = filepath.Clean(filepath.Join(wd, "../../../ginkGo/financials/data"))
	certFileName := wd + "/IDC_ARIA_API_CRT.crt"
	keyFileName := wd + "/IDC_ARIA_API_KEY.key"
	return certFileName, keyFileName
}
