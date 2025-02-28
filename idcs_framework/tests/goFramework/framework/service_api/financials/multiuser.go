package financials

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/frisby_client"
	"os/exec"
	"strings"
	"time"
)

// OTP

func getOTP(cloudAccId string, memberEmail string, pg_password string) (string, string) {
	// To be implemented, this function will fetch OTP from the cluster's cloud account database
	// Hence exporting kubeconfig before executing this is mandatory
	otp_command := fmt.Sprintf(`kubectl exec cloudaccount-db-postgresql-0 -n idcs-system -- bash -c "PGPASSWORD=%s  psql -U postgres -qtAX -d cloudaccount -c  \"select otp_code from admin_otp where member_email ='%s' and cloud_account_id ='%s' ORDER BY id DESC LIMIT 1;\""`, pg_password, memberEmail, cloudAccId)
	cmd := exec.Command("/bin/sh", "-c", otp_command)
	//cmd := exec.Command("kubectl exec -it cloudaccount-db-postgresql-0 -n idcs-system -- bash -c "+"\"PGPASSWORD="+pg_password, " psql -U postgres -d cloudaccount -c ", "\"select invitation_code from members where member_email ='"+memberEmail+"' and admin_account_id ='"+cloudAccId+"';\"")
	logger.Logf.Infof("Get OTP command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	outStr1 := strings.TrimSpace(outStr)
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr1, errStr

}

func deleteOTP(cloudAccId string, memberEmail string, pg_password string) (string, string) {
	// To be implemented, this function will fetch OTP from the cluster's cloud account database
	// Hence exporting kubeconfig before executing this is mandatory
	otp_command := fmt.Sprintf(`kubectl exec cloudaccount-db-postgresql-0 -n idcs-system -- bash -c "PGPASSWORD=%s  psql -U postgres -qtAX -d cloudaccount -c  \"delete from admin_otp where member_email ='%s' and cloud_account_id ='%s';\""`, pg_password, memberEmail, cloudAccId)
	cmd := exec.Command("/bin/sh", "-c", otp_command)
	//cmd := exec.Command("kubectl exec -it cloudaccount-db-postgresql-0 -n idcs-system -- bash -c "+"\"PGPASSWORD="+pg_password, " psql -U postgres -d cloudaccount -c ", "\"select invitation_code from members where member_email ='"+memberEmail+"' and admin_account_id ='"+cloudAccId+"';\"")
	logger.Logf.Infof("Get OTP command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	outStr1 := strings.TrimSpace(outStr)
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr1, errStr

}

func createOTP(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func verifyOTP(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}
func resendOTP(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func resendInvitation(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func rejectInvitation(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func revokeInvitation(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func removeInvitation(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getMembers(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

// Invite

func getInviteCode(cloudAccId string, memberEmail string, pg_password string) (string, string) {
	// To be implemented, this function will fetch OTP from the cluster's cloud account database
	// Hence exporting kubeconfig before executing this is mandatory
	otp_command := fmt.Sprintf(`kubectl exec cloudaccount-db-postgresql-0 -n idcs-system -- bash -c "PGPASSWORD=%s  psql -U postgres -qtAX -d cloudaccount -c  \"select invitation_code from members where member_email ='%s' and admin_account_id ='%s' ORDER BY id DESC LIMIT 1;\""`, pg_password, memberEmail, cloudAccId)
	cmd := exec.Command("/bin/sh", "-c", otp_command)
	//cmd := exec.Command("kubectl exec cloudaccount-db-postgresql-0 -n idcs-system -- bash -c", "\"PGPASSWORD=", pg_password, " psql -U postgres -d cloudaccount -c ", "\"select otp_code from admin_otp where member_email ='", memberEmail+"' and cloud_account_id ='"+cloudAccId+"';\"")
	logger.Logf.Infof("GetInvite command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	outStr1 := strings.TrimSpace(outStr)
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr1, errStr

}

func deleteMember(url string, token string, memberemail string) (int, string) {
	payload := fmt.Sprintf(`{
		"members": [
               "%s"
              ]
	}`, memberemail)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	frisby_response := frisby_client.Delete_With_Json(url, token, jsonMap)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func readInvite(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createInviteCode(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func verifyInviteCode(url string, token string, payload map[string]interface{}) (int, string) {
	fmt.Println("url", url)
	fmt.Println("payload", payload)
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func sendInviteCode(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Credits response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateOTP(base_url string, token string, api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := createOTP(base_url, token, jsonMap)
	return response_status, response_body
}

func VerifyOTP(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/v1/otp/verify"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := verifyOTP(post_url, token, jsonMap)
	return response_status, response_body
}

func ResendOTP(base_url string, token string, api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := resendOTP(base_url, token, jsonMap)
	return response_status, response_body
}

func Resendinvitation(base_url string, token string, api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := resendInvitation(base_url, token, jsonMap)
	return response_status, response_body
}

func Rejectinvitation(base_url string, token string, admin_cloudaccount_id string, invitation_stage string, member_cloudaccount_email string) (int, string) {
	payload := fmt.Sprintf(`{
		"adminAccountId": "%s",
		"invitationState": "%s",
		"memberEmail": "%s"
	}`, admin_cloudaccount_id, invitation_stage, member_cloudaccount_email)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	url := base_url + "/v1/cloudaccounts/invitations/member/reject"
	response_status, response_body := rejectInvitation(url, token, jsonMap)
	return response_status, response_body
}

func RevokeInvitation(base_url string, token string, admin_account_id string, member_email string) (int, string) {
	var jsonMap map[string]interface{}
	payload := fmt.Sprintf(`{
		"adminAccountId": "%s",
		"invitationState": "%s",
		"memberEmail": "%s"
	}`, admin_account_id, "INVITE_STATE_PENDING_ACCEPT", member_email)
	fmt.Println("Payload: ", payload)
	json.Unmarshal([]byte(payload), &jsonMap)
	url := base_url + "/v1/cloudaccounts/invitations/revoke"
	response_status, response_body := revokeInvitation(url, token, jsonMap)
	return response_status, response_body
}

func RemoveInvitation(base_url string, token string, admin_account_id string, member_cloudaccount_email string) (int, string) {
	payload := fmt.Sprintf(`{
		"adminAccountId": "%s",
		"invitationState": "%s",
		"memberEmail": "%s"
	}`, admin_account_id, "INVITE_STATE_ACCEPTED", member_cloudaccount_email)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	url := base_url + "/v1/cloudaccounts/invitations/remove"
	response_status, response_body := removeInvitation(url, token, jsonMap)
	return response_status, response_body
}

func GetOTP(cloudAccId string, memberEmail string, pg_password string) (string, string) {
	otp, errStr := getOTP(cloudAccId, memberEmail, pg_password)
	return otp, errStr
}

func DeleteOTP(cloudAccId string, memberEmail string, pg_password string) (string, string) {
	otp, errStr := deleteOTP(cloudAccId, memberEmail, pg_password)
	return otp, errStr
}

func DeleteMember(base_url string, token string, cloudAccid string, memberemail string) (int, string) {
	var delete_url = base_url + "/v1/cloudaccounts/id/" + cloudAccid + "/members/delete"
	delete_response_byid_body, delete_response_byid_status := deleteMember(delete_url, token, memberemail)
	return delete_response_byid_body, delete_response_byid_status
}

func CreateInviteCode(base_url string, token string, admin_cloudaccount_id string, member_email string) (int, string) {
	now := time.Now().UTC()
	current_time_formatted := now.Add(72 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	// fmt.Println("Time is :", current_time_formatted)

	// current_time := time.Now().Add(72 * time.Hour)
	// current_time_formatted := current_time.Format("2024-10-18T00:00:00.000Z")
	logger.Logf.Infof("Expiration time to be set: %s ", current_time_formatted)
	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"invites": [{ "expiry": "%s", "memberEmail": "%s", "note":"%s"}]
	}`, admin_cloudaccount_id, current_time_formatted, member_email, "Note")
	logger.Logf.Infof("Create Invite Code Url :%s", base_url)
	logger.Logf.Infof("Create Invite Code Payload :%s", payload)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createInviteCode(base_url, token, jsonMap)
	logger.Logf.Infof("Token Used :%s", token)
	logger.Logf.Infof("Create Invite Response Code :%s", response_status)
	logger.Logf.Infof("Create Invite Response Body :%s", response_body)
	return response_status, response_body
}

func CreateExpiredInviteCode(base_url string, token string, admin_cloudaccount_id string, member_email string) (int, string) {
	current_time := time.Now().Add(-10 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Expiration time to be set: ", current_time_d)
	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"invites": [{ "expiry": "%s", "memberEmail": "%s", "note":"%s"}]
	}`, admin_cloudaccount_id, current_time_d, member_email, "Note...")
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createInviteCode(base_url, token, jsonMap)
	return response_status, response_body
}

func GetInviteCode(cloudAccId string, memberEmail string, pg_password string) (string, string) {
	inviteCode, errStr := getInviteCode(cloudAccId, memberEmail, pg_password)
	return inviteCode, errStr
}

func ReadInvitations(base_url string, token string, cloudAccountId string, history string) (int, string) {
	var get_url = base_url + "/v1/cloudaccounts/invitations/read?adminAccountId=" + cloudAccountId
	logger.Logf.Infof("ReadInvitations Url :%s", get_url)
	get_response_status, get_response_body := readInvite(get_url, token)
	return get_response_status, get_response_body
}

func GetMembers(base_url string, token string, cloudAccountName string, history string) (int, string) {
	var get_url = base_url + "/name/" + cloudAccountName + "/members"
	fmt.Println("get_url", get_url)
	get_response_status, get_response_body := getMembers(get_url, token)
	return get_response_status, get_response_body
}

func GetActiveMembers(base_url string, token string, cloudAccountName string) (int, string) {
	var get_url = base_url + "/v1/cloudaccounts/name/" + cloudAccountName + "/members?onlyActive=true"
	logger.Logf.Infof("GetActiveMembers Url :%s", get_url)
	get_response_status, get_response_body := getMembers(get_url, token)
	return get_response_status, get_response_body
}

func VerifyInviteCode(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/v1/cloudaccounts/validateinvitecode"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := verifyInviteCode(post_url, token, jsonMap)
	return response_status, response_body
}

func SendInviteCode(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/v1/cloudaccounts/sendinvitecode"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := sendInviteCode(post_url, token, jsonMap)
	logger.Logf.Infof("SendInviteCode Url :%s", post_url)
	logger.Logf.Infof("SendInviteCode Payload :%s", jsonMap)
	logger.Logf.Infof("SendInviteCode Response Code :%s", response_status)
	logger.Logf.Infof("SendInviteCode Response Body :%s", response_body)
	return response_status, response_body
}
