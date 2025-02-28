package financials_utils

import (
	"fmt"
	"goFramework/framework/common/logger"
	"strings"
	"time"

	"github.com/tidwall/sjson"
)

func EnrichCreateCouponPayload(rawpayload string, amount float64, user string, numUses int) string {
	var enriched_payload = rawpayload
	// Get the current time
	start := time.Now().UTC()
	// Calculate the expiration time (2 days from now)
	expires := start.Add(48 * time.Hour)
	// Format the times as strings in the ISO 8601 format
	startStr := start.Format(time.RFC3339)
	expiresStr := expires.Format(time.RFC3339)
	//creator := "fnf_automations@intel.com"
	enriched_payload, _ = sjson.Set(enriched_payload, "amount", amount)
	enriched_payload, _ = sjson.Set(enriched_payload, "creator", user)
	enriched_payload, _ = sjson.Set(enriched_payload, "numUses", numUses)
	enriched_payload, _ = sjson.Set(enriched_payload, "start", startStr)
	enriched_payload, _ = sjson.Set(enriched_payload, "expires", expiresStr)

	return enriched_payload
}

func EnrichDisableCouponPayload(rawpayload string, code string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<code>>", code, 1)
	fmt.Println("enriched_payload", enriched_payload)
	return enriched_payload
}

func EnrichRedeemCouponPayload(rawpayload string, code string, cloudAccountId string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<code>>", code, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	return enriched_payload
}

func EnrichCreateCreditPayload(rawpayload string, amountUsed int, cloudAccountId string, couponCode string, created string, expiration string, originalAmount int, reason string, remainingAmount int) string {
	enriched_payload := rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "amountUsed", amountUsed)
	enriched_payload, _ = sjson.Set(enriched_payload, "cloudAccountId", cloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "couponCode", couponCode)
	enriched_payload, _ = sjson.Set(enriched_payload, "created", created)
	enriched_payload, _ = sjson.Set(enriched_payload, "expiration", expiration)
	enriched_payload, _ = sjson.Set(enriched_payload, "originalAmount", originalAmount)
	enriched_payload, _ = sjson.Set(enriched_payload, "reason", reason)
	enriched_payload, _ = sjson.Set(enriched_payload, "remainingAmount", remainingAmount)
	return enriched_payload
}

func EnrichCreditMigratePayload(rawpayload string, cloudAccountId string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	return enriched_payload
}

func EnrichStartSchedulerPayload(rawpayload string, action string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<action>>", action, 1)
	return enriched_payload
}

func EnrichUpgradeCouponPayload(rawpayload string, cloudAccountId string, cloudAccountUpgradeToType string, code string) string {
	logger.Logf.Infof("Upgrade coupon payload : %s ", rawpayload)
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountUpgradeToType>>", cloudAccountUpgradeToType, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<code>>", code, 1)
	logger.Logf.Infof("Enriched Payload : %s ", enriched_payload)
	return enriched_payload
}

func EnrichUpgradeWithoutCouponPayload(rawpayload string, cloudAccountId string, cloudAccountUpgradeToType string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountUpgradeToType>>", cloudAccountUpgradeToType, 1)
	return enriched_payload
}

// Multi User

func EnrichCreateInvitationPayload(rawpayload string, cloudAccountId string, expiry string, memberEmail string, note string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "cloudAccountId", cloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "expiry", expiry)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	enriched_payload, _ = sjson.Set(enriched_payload, "note", note)
	return enriched_payload
}

func EnrichResendInvitationPayload(rawpayload string, adminAccountId string, memberEmail string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "adminAccountId", adminAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	return enriched_payload
}

func EnrichRevokeInvitationPayload(rawpayload string, adminAccountId string, memberEmail string, invitationState string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "adminAccountId", adminAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	enriched_payload, _ = sjson.Set(enriched_payload, "invitationState", invitationState)
	return enriched_payload
}

func EnrichValidateInvitationCodePayload(rawpayload string, adminCloudAccountId string, memberEmail string, inviteCode string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "adminCloudAccountId", adminCloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	enriched_payload, _ = sjson.Set(enriched_payload, "inviteCode", inviteCode)
	return enriched_payload
}

func EnrichRejectInvitationPayload(rawpayload string, adminAccountId string, memberEmail string, invitationState string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "adminAccountId", adminAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	enriched_payload, _ = sjson.Set(enriched_payload, "invitationState", invitationState)
	return enriched_payload
}

func EnrichRemoveMemberPayload(rawpayload string, adminAccountId string, memberEmail string, invitation_state string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "adminAccountId", adminAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	enriched_payload, _ = sjson.Set(enriched_payload, "invitation_state", invitation_state)
	return enriched_payload
}

func EnrichCreateOTPPayload(rawpayload string, cloudAccountId string, memberEmail string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "cloudAccountId", cloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	return enriched_payload
}

func EnrichVerifyOTPPayload(rawpayload string, cloudAccountId string, memberEmail string, otpCode string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "cloudAccountId", cloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	enriched_payload, _ = sjson.Set(enriched_payload, "otpCode", otpCode)
	return enriched_payload
}

func EnrichResendOTPPayload(rawpayload string, cloudAccountId string, memberEmail string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "cloudAccountId", cloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "memberEmail", memberEmail)
	return enriched_payload
}

func EnrichCreateMaasUsagePayload(rawpayload string, cloudAccountId string, endTime string, processingType string, quantity string, region string, startTime string, timestamp string, transactionId string) string {
	var enriched_payload = rawpayload
	enriched_payload, _ = sjson.Set(enriched_payload, "cloudAccountId", cloudAccountId)
	enriched_payload, _ = sjson.Set(enriched_payload, "endTime", endTime)
	enriched_payload, _ = sjson.Set(enriched_payload, "processingType", processingType)
	enriched_payload, _ = sjson.Set(enriched_payload, "quantity", quantity)
	enriched_payload, _ = sjson.Set(enriched_payload, "region", region)
	enriched_payload, _ = sjson.Set(enriched_payload, "startTime", startTime)
	enriched_payload, _ = sjson.Set(enriched_payload, "timestamp", timestamp)
	enriched_payload, _ = sjson.Set(enriched_payload, "transactionId", transactionId)
	return enriched_payload
}
