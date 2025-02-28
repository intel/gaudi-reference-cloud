// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	mock "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/aria_mock_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/accts"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/plans"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gopkg.in/yaml.v3"
)

const (
	ACCOUNT_TYPE_PREMIUM            = "premium"
	ACCOUNT_TYPE_ENTERPRISE         = "enterprise"
	ACCOUNT_TYPE_ENTERPRISE_PENDING = "enterprise_pending"
	CURRENCY_CD                     = "usd"
	ISDEFAULT                       = 1
	USAGE_BASED                     = "Usage-Based"
	GLCD                            = "1"
	TIERED_PRICING                  = "Tiered Pricing"
	//usage type constant
	USAGE_TYPE_NAME                = "minutes"
	USAGE_TYPE_DESC                = "minutes"
	USAGE_TYPE_MINUTES_CODE_SUFFIX = "minutes"
	USAGE_UNIT_TYPE_MINUTE_DESC    = "minute"

	TIERED_SCHEDULE_FROM_UNITS = 1
	// plan constant
	PLANSUPPFIELDSYNC                           = "SYNC TO SAP GTS"
	PLANSUPPFIELDPCQID                          = "PCQ ID"
	PLANSUPPFIELDSYNCVALUE                      = "Yes"
	PLAN_SUPP_FIELD_BUSINESS_UNIT               = "Business Unit"
	PLAN_SUPP_FIELD_BUSINESS_UNIT_KEY           = "businessUnit"
	PLAN_SUPP_FIELD_BUSINESS_UNIT_VALUE         = "Intel Developer Cloud"
	PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT       = "Business Unit Contact"
	PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT_KEY   = "businessUnitContact"
	PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT_VALUE = "Harmen Van der Linde"
	PLAN_SUPP_FIELD_GL_NUMBER                   = "GL Number"
	PLAN_SUPP_FIELD_GL_NUMBER_KEY               = "glnumber"
	PLAN_SUPP_FIELD_GL_NUMBER_VALUE             = "431510"
	PLAN_SUPP_FIELD_LEGAL_ENTITY                = "Legal Entity"
	PLAN_SUPP_FIELD_LEGAL_ENTITY_KEY            = "legalEntity"
	PLAN_SUPP_FIELD_LEGAL_ENTITY_VALUE          = "LE036"
	PLAN_SUPP_FIELD_PRODUCT_LINE                = "Product Line"
	PLAN_SUPP_FIELD_PRODUCT_LINE_KEY            = "productLine"
	PLAN_SUPP_FIELD_PRODUCT_LINE_VALUE          = "Intel Developer Cloud"
	PLAN_SUPP_FIELD_PROFIT_CENTER               = "Profit Center"
	PLAN_SUPP_FIELD_PROFIT_CENTER_KEY           = "profitCenter"
	PLAN_SUPP_FIELD_PROFIT_CENTER_VALUE         = "3639"
	PLAN_SUPP_FIELD_SUPER_GROUP                 = "Super Group"
	PLAN_SUPP_FIELD_SUPER_GROUP_KEY             = "superGroup"
	PLAN_SUPP_FIELD_SUPER_GROUP_VALUE           = "SATG"
	PLAN_TIERED_PRICING_RULE                    = 1
	PLAN_INSTANCE_INDEX                         = 1
	PLAN_INSTANCE_STATUS                        = 1
	PLAN_INSTANCE_UNITS                         = 1
	BILL_LAG_DAYS                               = 7
	ALT_BILL_DAY                                = 0
	PLAN_EDIT_DIRECTIVE                         = 2
	//TODO: replace with config value
	BILLING_GROUP_NOTIFY_METHOD               = 10
	BILLING_GROUP_INDEX                       = 1
	DUNNING_GROUP_INDEX                       = 1
	ACCOUNT_NOTIFY_METHOD                     = 10
	ACCOUNT_CURRENCY                          = "USD"
	ACCOUNT_COUNTRY                           = "US"
	ACCOUNT_FUNC_GROUP_ID                     = "LE036"
	ACCOUNT_INVOICE_OPTION                    = 1
	ClIENT_DUNNING_PROCESS_ID                 = "Credit_Card_B2C"
	ENTERPRISE_ClIENT_DUNNING_PROCESS_ID      = "Standard_Dunning"
	SUPP_FIELD_NAME_WORKFLOW                  = "Auto approval by workflow"
	SUPP_FIELD_VALUE_WORKFLOW                 = "TRUE"
	SUPP_FIELD_NAME_COMPANY_CODE              = "Company Code"
	SUPP_FIELD_VALUE_COMPANY_CODE             = "036"
	SUPP_FIELD_NAME_INTEL_ADDRESS             = "Intel Address"
	SUPP_FIELD_VALUE_INTEL_ADDRESS            = "Intel Services Divisions LLC\n2200 Mission College Blvd\nSanta Clara,CA 95054"
	SUPP_FIELD_NAME_ACCOUNT_TYPE              = "IDC Account Type"
	SUPP_FIELD_VALUE_ACCOUNT_TYPE_PREMIUM     = "Premium"
	SUPP_FIELD_VALUE_ACCOUNT_TYPE_ENT_PENDING = "Enterprise Child Pending"
	SUPP_FIELD_VALUE_ACCOUNT_TYPE_ENTERPRISE  = "Enterprise Child"
	SUPP_FIELD_VALUE_ACCOUNT_TYPE_UNKNOWN     = "Unknown"
	ALT_CALLER_ID                             = "IDC Tester"
	USAGES_RATING_TIMING                      = 1
	STATUS_UNTIL_ALT_START                    = 1
	DATETIME_LAYOUT                           = "2006-01-02 15:04:05"
	MONTHYEAR_LAYOUT                          = "January, 2006"
	DATE_LAYOUT                               = "2006-01-02"
	TIME_LAYOUT                               = "15:04:05"
	ENTERPRISE_ACCOUNT_ASSIGNMENT_DIRECTIVE   = 4
	PREMIUM_ACCOUNT_ASSIGNMENT_DIRECTIVE      = 4
	PREMIUM_ACCOUNT_PLAN_INVOICING_OPTION     = 2
	PRORATE_FIRST_INVOICE                     = 2
	PARENT_PAY_FOR_CHILD_ACCOUNT              = 2
	CONTACT_INDICATOR                         = 1
)

func mapKeysToString(in map[any]any) map[string]any {
	out := map[string]any{}
	for key, val := range in {
		valMap, ok := val.(map[any]any)
		if ok {
			val = mapKeysToString(valMap)
		}
		out[fmt.Sprintf("%v", key)] = val
	}
	return out
}

func GetErrorForStatusCode(ariaApi string, statusCode int) error {
	return fmt.Errorf("aria api:%s,status code:%d", ariaApi, statusCode)
}

func GetErrorForParamter(ariaApi string, parameter any, err error) error {
	return fmt.Errorf("aria api:%s, paramter :%v error %v", ariaApi, parameter, err)
}

func GetDriverError(driverApiError string, err error) error {
	return fmt.Errorf("aria driver api error:%s,aria controller error:%v", driverApiError, err)
}

func GetErrorForAriaErrorCode(idcError string, ariaApi string, ariaApiErrorCode int64, ariaApiErrorMessage string) error {
	return fmt.Errorf(":context:%s,aria api:%s,aria error code:%d,aria error message:%s",
		idcError, ariaApi, ariaApiErrorCode, ariaApiErrorMessage)
}

func DebugLogJsonPayload(ctx context.Context, msg string, payload []byte) {
	logger := log.FromContext(ctx).V(1)
	if !logger.Enabled() {
		return
	}
	data := map[string]any{}
	if err := json.Unmarshal(payload, &data); err != nil {
		logger.Error(err, "error unmarshaling payload")
		return
	}
	if data["auth_key"] != "" {
		data["auth_key"] = "[REDACTED]"
	}
	redacted, err := json.Marshal(data)
	if err != nil {
		logger.Error(err, "error marshaling redacted payload")
		return
	}
	logger.Info(msg, "payload", string(redacted))
}

func DebugLogQueryPayload(ctx context.Context, msg string, payload string) {
	logger := log.FromContext(ctx).V(1)
	if !logger.Enabled() {
		return
	}
	data := []byte(payload)
	authKey := bytes.Index(data, []byte("auth_key="))
	if authKey < 0 {
		logger.Info("error finding auth_key in payload")
		return
	}
	start := bytes.IndexByte(data[authKey:], '=')
	if start < 0 {
		logger.Info("error parsing start of auth_key in payload")
		return
	}
	start += authKey + 1
	end := bytes.IndexByte(data[start:], '&')
	if end < 0 {
		end = len(data[start:])
	}
	end += start
	buf := bytes.Buffer{}
	if _, err := buf.Write(data[:start]); err != nil {
		logger.Info("error writting to buffer")
		return
	}
	if _, err := buf.WriteString("[REDACTED]"); err != nil {
		logger.Info("error writting string to buffer")
		return
	}
	if _, err := buf.Write(data[end:]); err != nil {
		logger.Info("error writting to buffer")
		return
	}
	data = buf.Bytes()
	logger.Info(msg, "payload", string(data))
}

func IsPayloadEmpty(payload interface{}) bool {
	if payload == nil {
		return true
	} else if payload == "" {
		return true
	}
	if reflect.ValueOf(payload).Kind() == reflect.Struct {
		empty := reflect.New(reflect.TypeOf(payload)).Elem().Interface()
		if reflect.DeepEqual(payload, empty) {
			return true
		} else {
			return false
		}
	}
	return false
}

func ConvertYamlToJson(ctx context.Context, createRequestDefaults []byte, requestCategory string) ([]byte, error) {
	logger := log.FromContext(ctx).WithName("AriaClientUtil.ConvertYamlToJson")
	mm := map[any]any{}

	if err := yaml.Unmarshal(createRequestDefaults, &mm); err != nil {
		logger.Error(err, "failed to unmarshal YAML default values for create ", requestCategory, " request: ", string(createRequestDefaults))
		return nil, err
	}
	createRequestJson, err := json.Marshal(mapKeysToString(mm))
	if err != nil {
		logger.Error(err, "failed to JSON marshal default values for create ", requestCategory, " request: ", string(createRequestDefaults))
		return nil, err
	}
	return createRequestJson, nil
}

func genValue(key string, val any, builder *strings.Builder) error {
	logger := log.FromContext(context.Background()).WithName("genValue")
	if builder.Len() > 0 {
		err := builder.WriteByte('&')
		if err != nil {
			logger.Error(err, "Error executing WriteByte")
			return err
		}
	}
	if _, err := builder.WriteString(key); err != nil {
		logger.Error(err, "Error executing WriteString", key)
		return err
	}
	if _, err := builder.WriteRune('='); err != nil {
		logger.Error(err, "Error executing WriteRune")
		return err
	}
	strVal := ""
	switch value := val.(type) {
	case float64:
		bb, err := json.Marshal(value)
		if err != nil {
			logger.Error(err, "Error executing Marshal")
			return err
		}
		strVal = string(bb)
	case string:
		strVal = value
	default:
		strVal = fmt.Sprintf("%v", value)
	}
	_, err := builder.WriteString(strVal)
	if err != nil {
		logger.Error(err, "Error executing WriteString", strVal)
		return err
	}
	return nil
}

func genQueryAny(prefix string, key any, val any, builder *strings.Builder) error {
	kk := ""
	if prefix == "" {
		kk = fmt.Sprintf("%v", key)
	} else {
		kk = fmt.Sprintf("%v[%v]", prefix, key)
	}

	switch value := val.(type) {
	case []any:
		return genQueryArray(kk, value, builder)
	case map[string]any:
		return genQueryMap(kk, value, builder)
	default:
		return genValue(kk, value, builder)
	}
}

func genQueryArray(prefix string, vals []any, builder *strings.Builder) error {
	for ii, val := range vals {
		if err := genQueryAny(prefix, ii, val, builder); err != nil {
			return err
		}
	}
	return nil
}

func genQueryMap(prefix string, vals map[string]any, builder *strings.Builder) error {
	for key, val := range vals {
		if err := genQueryAny(prefix, key, val, builder); err != nil {
			return err
		}
	}
	return nil
}

func GenerateQuery(jsonStr []byte) (string, error) {
	logger := log.FromContext(context.Background()).WithName("GenerateQuery")
	var data map[string]interface{}
	if err := json.Unmarshal(jsonStr, &data); err != nil {
		logger.Error(err, "Error executing json Unmarshal")
		return "", err
	}

	builder := strings.Builder{}
	if err := genQueryMap("", data, &builder); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func parseBrackets(str string) (string, string, error) {
	startBrack := strings.Index(str, "[")
	if startBrack == -1 {
		// No brackets. Simple string
		return str, "", nil
	}
	if startBrack > 0 {
		// Something before the first bracket. Top level map item
		return str[:startBrack], str[startBrack:], nil
	}
	subKey := str[startBrack+1:]
	endBrack := strings.Index(subKey, "]")
	if endBrack == -1 {
		return "", "", fmt.Errorf("missing close bracket in key %v", str)
	}

	// nextBrack is whatever comes after the end bracket. It should be either nothing
	// or another bracket
	nextBrack := subKey[endBrack+1:]
	if len(nextBrack) > 0 && nextBrack[0] != '[' {
		return "", "", fmt.Errorf("non bracket key found in key %v", str)
	}

	// inBrack is a numeric index if we're parsing an array, or a string if
	// we're parsing a map.
	subKey = subKey[:endBrack]

	return subKey, nextBrack, nil
}

// Make a slice big enough to hold an entry at ind
func ensureIndex(arr []any, ind int) []any {
	for jj := len(arr); jj <= int(ind); jj++ {
		arr = append(arr, nil)
	}
	return arr
}

// Attempt to interpret "key" as an int. Returns -1 if key is not an int
func keyToInt(key string) int {
	ii, err := strconv.ParseInt(key, 10, 32)
	if err != nil {
		return -1
	}
	return int(ii)
}

// If we're parsing what we think is an array and we encounter a non int
// key, we convert the array to a map
func arrayToMap(arr []any) map[string]any {
	hash := map[string]any{}
	for ii, val := range arr {
		hash[fmt.Sprintf("%v", ii)] = val
	}
	return hash
}

// Get the container to use for this key.
// cont can be either []any, map[string]any, or nil
func getContainerForKey(cont any, key string) (any, error) {
	if hash, ok := cont.(map[string]any); ok {
		// container is already a hash, so use it
		return hash, nil
	}
	ii := keyToInt(key)
	arr, ok := cont.([]any)
	if (ok || cont == nil) && ii != -1 {
		// container is a slice or nil, the new key is an int. Expand the slice
		// to be big enough or create a new slice.
		return ensureIndex(arr, ii), nil
	}
	if ok {
		// convert array to map since we encountered a non-int index
		return arrayToMap(arr), nil
	}

	if cont != nil {
		// cont is neither a map nor an array. It must be a scalar value we've
		// already set, so it's a duplicate
		return nil, fmt.Errorf("duplicate value for %v", key)
	}

	// Create a new map
	return map[string]any{}, nil
}

// Set the key=val in cont. The caller has already called getContainerForKey so we
// don't have to check for error scenarios.
func setContainerVal(cont any, key string, val any) {
	if hash, ok := cont.(map[string]any); ok {
		// cont is a map
		hash[key] = val
		return
	}
	// cont is a slice
	arr := cont.([]any)
	arr[keyToInt(key)] = val
}

// Get the value for key from a map or array. The value is either
// map[string]any, []any, or nil and is the next level container
// Caller already called getContainerForKey so we don't have to check for error
// scenarios
func getContainerVal(cont any, key string) any {
	if hash, ok := cont.(map[string]any); ok {
		return hash[key]
	}
	arr := cont.([]any)
	return arr[keyToInt(key)]
}

// Walk through key, creating arrays and maps corresponding to
// bracketed array indexes and map keys as necessary.
//
// Example key:
// service[0][tier][0][schedule][1][from]=1
//
// result:
//
//	map[string]any{
//		service: []any{
//			map[string]any{
//				tier: []any{
//					nil,
//					map[string]any {
//						schedule: []any {
//							map[string]any {
//								from: float64(1)
//							}
//						}
//					}
//				}
//			}
//		}
//	}
func setQueryParam(cont any, key string, val any) (any, error) {
	subKey, nextBrack, err := parseBrackets(key)
	if err != nil {
		return nil, err
	}
	cont, err = getContainerForKey(cont, subKey)
	if err != nil {
		return nil, err
	}
	if nextBrack == "" {
		// No more brackets means the actual value goes into the current
		// container. No more recursive calls for this key.
		setContainerVal(cont, subKey, val)
		return cont, nil
	}

	// Create the next level of container and make a recursive call
	// to parse the rest of the key
	contVal, err := setQueryParam(getContainerVal(cont, subKey), nextBrack, val)
	if err != nil {
		return nil, err
	}

	setContainerVal(cont, subKey, contVal)
	return cont, nil
}

// Get json unmarshaled version of string. This gets the types right
// for numeric and bool values
func getJsonValue(str string) any {
	var val any
	err := json.Unmarshal([]byte(str), &val)
	if err != nil {
		return str
	}
	return val
}

// Parse a query string for the Aria admin API. The returned
// map[string]any is suitable for passing to json.Marshal. The
// resulting json can be unmarshaled into an Aria API struct.
//
// It would be nice to parse the query directly into an Aria API
// struct and skip the intermediate JSON. A project for another day...
func ParseQuery(query string) (map[string]any, error) {
	result := map[string]any{}

	params := strings.Split(query, "&")
	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("missing equals in param %v", param)
		}
		if _, err := setQueryParam(result, kv[0], getJsonValue(kv[1])); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func GetPlanRequestData(ctx context.Context, planRequest *request.PlanRequest, product *pb.Product, productFamily *pb.ProductFamily, usageType *data.UsageType, cfg *config.Config) error {

	logger := log.FromContext(ctx).WithName("AriaClientUtil.GetPlanRequestData")
	var err error

	planRequest.PlanName = product.Name
	planRequest.PlanDescription = product.Description
	planRequest.ClientPlanId = GetPlanClientId(product.Id)

	planRequest.Service = make([]plans.Service, 1)
	planRequest.Service[0].Tier = make([]plans.Tier, 1)

	// For Multiple product rates (Premium, Enterprise) which needs to be mapped to CreatePlan's schedule and service
	// TODO: Multiple Rate schedule tier needs to be handeled
	for _, rate := range product.Rates {
		schedName := GetAccountType(rate)
		if schedName == "" {
			continue
		}
		sched := plans.Schedule{}
		sched.ScheduleName = schedName
		sched.ClientRateScheduleId = fmt.Sprintf("%s.%s.%s", cfg.ClientIdPrefix, product.Id, schedName)
		sched.CurrencyCd = CURRENCY_CD

		tierSched := plans.TierSchedule{}
		// Service tier schedule `from` is required during create plan but not for edit/deactivate plan

		tierSched.From = 1
		if rate.AccountType == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
			sched.IsDefault = ISDEFAULT
		}

		planRequest.Schedule = append(planRequest.Schedule, sched)

		tierSched.Amount, err = strconv.ParseFloat(rate.Rate, 64)
		if err != nil {
			logger.Error(err, "error converting string to float64")
			return err
		}

		planRequest.Service[0].Tier[0].Schedule = append(planRequest.Service[0].Tier[0].Schedule, tierSched)
	}

	planRequest.Service[0].ClientServiceId = GetServiceClientId(product.Id)
	planRequest.Service[0].Name = product.Metadata["displayName"]
	planRequest.Service[0].ServiceType = USAGE_BASED
	planRequest.Service[0].GlCd = GLCD
	planRequest.Service[0].RateType = TIERED_PRICING
	planRequest.Service[0].UsageType = usageType.UsageTypeNo
	planRequest.Service[0].TaxableInd = 1
	planRequest.Service[0].ClientTaxGroupId = cfg.GetAriaSystemClientTaxGroupId()
	planRequest.Service[0].TaxGroup = cfg.GetAriaSystemTaxGroup()
	return err
}

func GetPlanRequestDataFromClientPlanDetail(ctx context.Context, planRequest *request.PlanRequest, clientPlanDetail *data.AllClientPlanDtl, usageType *data.UsageType, cfg *config.Config) {
	planRequest.PlanName = clientPlanDetail.PlanName
	planRequest.PlanDescription = clientPlanDetail.PlanDesc
	planRequest.ClientPlanId = clientPlanDetail.ClientPlanId

	planRequest.Service = make([]plans.Service, 0, len(clientPlanDetail.PlanServices))

	for _, service := range clientPlanDetail.PlanServices {
		serviceReq := plans.Service{
			ClientServiceId: service.ClientServiceId,
			Name:            service.ServiceDesc,
			ServiceType:     USAGE_BASED,
			GlCd:            GLCD,
			RateType:        TIERED_PRICING,
			UsageType:       usageType.UsageTypeNo,
			Tier:            make([]plans.Tier, 1),
		}

		for _, rate := range service.PlanServiceRates {
			schedName := GetRateType(rate.ClientRateScheduleId, len(clientPlanDetail.ClientPlanId))
			if schedName != "" {
				if (strings.ToLower(schedName) != strings.ToLower(ACCOUNT_TYPE_PREMIUM)) ||
					(strings.ToLower(schedName) != strings.ToLower(ACCOUNT_TYPE_ENTERPRISE)) {
					schedName = ACCOUNT_TYPE_PREMIUM
				}
				sched := plans.Schedule{}
				sched.ScheduleName = schedName
				sched.ClientRateScheduleId = rate.ClientRateScheduleId
				sched.CurrencyCd = CURRENCY_CD
				tierSched := plans.TierSchedule{}
				tierSched.From = 1
				if strings.EqualFold(schedName, ACCOUNT_TYPE_PREMIUM) {
					sched.IsDefault = ISDEFAULT
				}
				planRequest.Schedule = append(planRequest.Schedule, sched)
				tierSched.Amount = rate.RatePerUnit
				serviceReq.Tier[0].Schedule = append(serviceReq.Tier[0].Schedule, tierSched)
			}
		}
		planRequest.Service = append(planRequest.Service, serviceReq)
	}
}

func GetRateType(clientRateId string, planIdLen int) string {
	if len(clientRateId) > planIdLen {
		return clientRateId[planIdLen+1:]
	}
	return ""
}

func GetAccountType(rate *pb.Rate) string {
	accountType := ""
	switch rate.AccountType {
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		accountType = ACCOUNT_TYPE_PREMIUM
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		accountType = ACCOUNT_TYPE_ENTERPRISE
	}
	return accountType
}

func GetMasterPlanDetail(clientPlanId string) accts.MasterPlansDetail {
	masterPlanDetail := accts.MasterPlansDetail{
		ClientPlanId:       clientPlanId,
		PlanInstanceIdx:    PLAN_INSTANCE_INDEX,
		PlanInstanceUnits:  PLAN_INSTANCE_UNITS,
		PlanInstanceStatus: PLAN_INSTANCE_STATUS,
		BillingGroupIdx:    BILLING_GROUP_INDEX,
		DunningGroupIdx:    DUNNING_GROUP_INDEX,
	}
	return masterPlanDetail
}

func GetFunctionalAccountGroup() accts.FunctionalAcctGroup {
	functionalAcctGroup := accts.FunctionalAcctGroup{
		ClientFunctionalAcctGroupId: ACCOUNT_FUNC_GROUP_ID,
	}
	return functionalAcctGroup
}

func GetBillingGroup(clientAccountId string, clientPlanId string) accts.BillingGroup {
	billingGroup := accts.BillingGroup{
		ClientBillingGroupId: GetBillingGroupId(clientAccountId),
		BillingGroupName:     GetBillingGroupName(clientPlanId),
		BillingGroupIdx:      BILLING_GROUP_INDEX,
		NotifyMethod:         BILLING_GROUP_NOTIFY_METHOD,
	}
	return billingGroup
}
func GetDunningGroup(clientAccountId string, clientPlanId string) accts.DunningGroup {
	dunningGroup := accts.DunningGroup{
		ClientDunningGroupId:   GetDunningGroupId(clientAccountId),
		DunningGroupName:       GetDunningGroupName(clientPlanId),
		DunningGroupIdx:        DUNNING_GROUP_INDEX,
		ClientDunningProcessId: ClIENT_DUNNING_PROCESS_ID,
	}
	return dunningGroup
}

func GetSuppFiled() []accts.SuppField {
	suppFields := []accts.SuppField{}
	suppFieldWorkflow := accts.SuppField{
		SuppFieldName:  SUPP_FIELD_NAME_WORKFLOW,
		SuppFieldValue: SUPP_FIELD_VALUE_WORKFLOW,
	}
	suppFields = append(suppFields, suppFieldWorkflow)
	suppField := accts.SuppField{
		SuppFieldName:  SUPP_FIELD_NAME_COMPANY_CODE,
		SuppFieldValue: SUPP_FIELD_VALUE_COMPANY_CODE,
	}
	suppFields = append(suppFields, suppField)
	//TODO: Change after review with ESPRIT Team
	suppFieldIntelAddress := accts.SuppField{
		SuppFieldName:  SUPP_FIELD_NAME_INTEL_ADDRESS,
		SuppFieldValue: SUPP_FIELD_VALUE_INTEL_ADDRESS,
	}
	suppFields = append(suppFields, suppFieldIntelAddress)
	return suppFields
}

func MockServer() error {
	logger := log.FromContext(context.Background()).WithName("MockServerInit")
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Addr: "localhost:" + fmt.Sprint(port)}

	url := fmt.Sprintf("http://localhost:%d", port)
	config.Cfg = config.NewDefaultConfig()
	config.Cfg.AriaSystem.Server.CoreApiUrl = url + "/v1/core"
	config.Cfg.AriaSystem.Server.AdminApiUrl = url + "/admin"
	config.Cfg.AriaSystem.ClientNo = mock.ClientNo
	config.Cfg.AriaSystem.AuthKey = mock.AuthKey

	http.HandleFunc("/v1/core", RestCallAssignerWrapper)
	http.HandleFunc("/admin", RestCallAssignerWrapperAdmin)
	go func() {
		logger.Info("Mock Server has started", "addr", srv.Addr)
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			logger.Error(err, "Error in Listen and Server function:")
		}
	}()
	return nil
}

func ParseDateAndTime(date string, timeOfDay string) time.Time {
	logger := log.FromContext(context.Background()).WithName("ParseDateAndTime")
	dateStr := fmt.Sprintf("%s %s", date, timeOfDay)
	// If ParseDateTime fails it will use 0001-01-01 00:00:00 +0000 UTC
	// as datetime to sort the data.
	t, err := ParseDateTime(dateStr)
	if err != nil {
		logger.Error(err, "Error while parsing the date", dateStr)
	}
	return t
}

func ParseDateTime(dateTime string) (time.Time, error) {
	t, err := time.Parse(DATETIME_LAYOUT, dateTime)
	return t, err
}

func ParseAriaDate(date string) (*timestamppb.Timestamp, error) {
	t, err := time.Parse(DATE_LAYOUT, date)
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}

func TimeToMonthYearFormat(dateTime time.Time) string {
	return dateTime.Format(MONTHYEAR_LAYOUT)
}

func TimestampToAriaFormat(dateTime *timestamppb.Timestamp) string {
	return DateTimeToAriaFormat(dateTime.AsTime())
}

func DateTimeToAriaFormat(dateTime time.Time) string {
	return dateTime.Format(DATETIME_LAYOUT)
}

func TimestampToAriaDateFormat(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return DateToAriaFormat(ts.AsTime())
}

func SplitDateTimeToAriaFormat(dateTime time.Time) (string, string) {
	return dateTime.Format(DATE_LAYOUT), dateTime.Format(TIME_LAYOUT)
}

func DateToAriaFormat(date time.Time) string {
	return date.Format(DATE_LAYOUT)
}

func GetClientNotificationTemplateGroupId(typ pb.AccountType, country string) string {
	notificationTmpltGrpId := GetDefaultNotificationTemplateGroupId()
	if typ == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		if country == "US" {
			notificationTmpltGrpId = "B2C_US_Statement_Template"
		} else {
			notificationTmpltGrpId = "B2C_VAT_Statement_Template"
		}
	} else if typ == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
		if country == "US" {
			notificationTmpltGrpId = "US_Statement_Template"
		} else {
			notificationTmpltGrpId = "VAT_Statement_Template"
		}
	}
	return notificationTmpltGrpId
}

func ContainsNotificationTemplateGroupId(notificationDetails []data.AccountNotificationDetail, notificationTemplateGroupId string) bool {
	for _, element := range notificationDetails {
		if element.NotifyTmpltGrpId == notificationTemplateGroupId {
			return true
		}
	}
	return false
}

func GetLimitedDecimalUsageUnit(usageUnit float64) float64 {
	usageUnitStr := strconv.FormatFloat(usageUnit, 'f', 2, 64)
	fmtUsageUnit, err := strconv.ParseFloat(usageUnitStr, 64)
	if err != nil {
		return usageUnit
	}
	return fmtUsageUnit
}

func GetSupplementalObjField(product *pb.Product) []plans.SupplementalObjectField {
	suppOpjFields := []plans.SupplementalObjectField{}
	suppFieldPCQId := plans.SupplementalObjectField{
		FieldName:  PLANSUPPFIELDPCQID,
		FieldValue: []string{product.Pcq},
	}
	suppOpjFields = append(suppOpjFields, suppFieldPCQId)
	businessUnit := GetBusinessUnit(product)
	suppFieldBusinessUnit := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_BUSINESS_UNIT,
		FieldValue: []string{businessUnit},
	}
	suppOpjFields = append(suppOpjFields, suppFieldBusinessUnit)
	businessUnitContact := GetBusinessUnitContact(product)
	suppFieldBusinessUnitContact := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT,
		FieldValue: []string{businessUnitContact},
	}
	suppOpjFields = append(suppOpjFields, suppFieldBusinessUnitContact)
	glnumber := GetGLNumber(product)
	suppFieldGLNumber := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_GL_NUMBER,
		FieldValue: []string{glnumber},
	}
	suppOpjFields = append(suppOpjFields, suppFieldGLNumber)
	legalEntity := GetLegalEntity(product)
	suppFieldLegalEntity := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_LEGAL_ENTITY,
		FieldValue: []string{legalEntity},
	}
	suppOpjFields = append(suppOpjFields, suppFieldLegalEntity)
	productLine := GetProductLine(product)
	suppFieldProductLine := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_PRODUCT_LINE,
		FieldValue: []string{productLine},
	}
	suppOpjFields = append(suppOpjFields, suppFieldProductLine)
	profitCenter := GetProfitCenter(product)
	suppFieldProfitCenter := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_PROFIT_CENTER,
		FieldValue: []string{profitCenter},
	}
	suppOpjFields = append(suppOpjFields, suppFieldProfitCenter)
	suberGroup := GetSuberGroup(product)
	suppFieldSuberGroup := plans.SupplementalObjectField{
		FieldName:  PLAN_SUPP_FIELD_SUPER_GROUP,
		FieldValue: []string{suberGroup},
	}
	suppOpjFields = append(suppOpjFields, suppFieldSuberGroup)
	return suppOpjFields
}

func GetSuberGroup(product *pb.Product) string {
	suberGroup := PLAN_SUPP_FIELD_SUPER_GROUP_VALUE
	if suberGroupVal, ok := product.Metadata[PLAN_SUPP_FIELD_SUPER_GROUP_KEY]; ok {
		suberGroup = suberGroupVal
	}
	return suberGroup
}

func GetProfitCenter(product *pb.Product) string {
	profitCenter := PLAN_SUPP_FIELD_PROFIT_CENTER_VALUE
	if profitCenterVal, ok := product.Metadata[PLAN_SUPP_FIELD_PROFIT_CENTER_KEY]; ok {
		profitCenter = profitCenterVal
	}
	return profitCenter
}

func GetProductLine(product *pb.Product) string {
	productLine := PLAN_SUPP_FIELD_PRODUCT_LINE_VALUE
	if productLineVal, ok := product.Metadata[PLAN_SUPP_FIELD_PRODUCT_LINE_KEY]; ok {
		productLine = productLineVal
	}
	return productLine
}

func GetLegalEntity(product *pb.Product) string {
	legalEntity := PLAN_SUPP_FIELD_LEGAL_ENTITY_VALUE
	if legalEntityVal, ok := product.Metadata[PLAN_SUPP_FIELD_LEGAL_ENTITY_KEY]; ok {
		legalEntity = legalEntityVal
	}
	return legalEntity
}

func GetGLNumber(product *pb.Product) string {
	glnumber := PLAN_SUPP_FIELD_GL_NUMBER_VALUE
	if glVal, ok := product.Metadata[PLAN_SUPP_FIELD_GL_NUMBER_KEY]; ok {
		glnumber = glVal
	}
	return glnumber
}

func GetBusinessUnitContact(product *pb.Product) string {
	businessUnitContact := PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT_VALUE
	if buContactVal, ok := product.Metadata[PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT_KEY]; ok {
		businessUnitContact = buContactVal
	}
	return businessUnitContact
}

func GetBusinessUnit(product *pb.Product) string {
	businessUnit := PLAN_SUPP_FIELD_BUSINESS_UNIT_VALUE
	if buVal, ok := product.Metadata[PLAN_SUPP_FIELD_BUSINESS_UNIT_KEY]; ok {
		businessUnit = buVal
	}
	return businessUnit
}
