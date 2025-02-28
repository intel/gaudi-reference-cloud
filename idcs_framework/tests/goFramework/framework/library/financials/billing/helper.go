package billing

//Structs to store responses

const (
	ServiceNamePrefix string = "IDC."
	AriaClientId      string = "IDC.driver"
)

type CreateCloudAccountStruct struct {
	BillingAccountCreated  bool   `json:"billingAccountCreated"`
	CreditsDepleted        string `json:"creditsDepleted"`
	Enrolled               bool   `json:"enrolled"`
	LowCredits             bool   `json:"lowCredits"`
	Name                   string `json:"name"`
	Oid                    string `json:"oid"`
	Owner                  string `json:"owner"`
	ParentID               string `json:"parentId"`
	TerminateMessageQueued bool   `json:"terminateMessageQueued"`
	TerminatePaidServices  bool   `json:"terminatePaidServices"`
	Tid                    string `json:"tid"`
	Type                   string `json:"type"`
}

type CreateCloudAccount1Response struct {
	Id string `json:"id"`
}

type CreateBillingAccountWithCloudAccIdStruct struct {
	CloudAccountID string `json:"cloudAccountId"`
}

type AriaUnappliedCredits struct {
	AcctNo                  string `json:"acct_no"`
	ClientAcctID            string `json:"client_acct_id"`
	Userid                  string `json:"userid"`
	StatusCd                int    `json:"status_cd"`
	NotifyMethod            int    `json:"notify_method"`
	TestAcctInd             int    `json:"test_acct_ind"`
	AcctStartDate           string `json:"acct_start_date"`
	SeqFuncGroupNo          int    `json:"seq_func_group_no"`
	InvoiceApprovalRequired int    `json:"invoice_approval_required"`
	FunctionalAcctGroup     []struct {
		FunctionalAcctGroupNo       int    `json:"functional_acct_group_no"`
		ClientFunctionalAcctGroupID string `json:"client_functional_acct_group_id"`
	} `json:"functional_acct_group"`
	SuppField []struct {
		SuppFieldName  string `json:"supp_field_name"`
		SuppFieldValue string `json:"supp_field_value"`
	} `json:"supp_field"`
	AcctCurrency    string `json:"acct_currency"`
	MasterPlanCount int    `json:"master_plan_count"`
	SuppPlanCount   int    `json:"supp_plan_count"`
	MasterPlansInfo []struct {
		MasterPlanInstanceNo       int    `json:"master_plan_instance_no"`
		ClientMasterPlanInstanceID string `json:"client_master_plan_instance_id"`
		ClientMasterPlanID         string `json:"client_master_plan_id"`
		MasterPlanNo               int    `json:"master_plan_no"`
		DunningGroupNo             int    `json:"dunning_group_no"`
		ClientDunningGroupID       string `json:"client_dunning_group_id"`
		DunningGroupName           string `json:"dunning_group_name"`
		DunningProcessNo           int    `json:"dunning_process_no"`
		ClientDunningProcessID     string `json:"client_dunning_process_id"`
		BillingGroupNo             int    `json:"billing_group_no"`
		ClientBillingGroupID       string `json:"client_billing_group_id"`
		MasterPlanInstanceStatus   int    `json:"master_plan_instance_status"`
		MpInstanceStatusLabel      string `json:"mp_instance_status_label"`
		MasterPlanUnits            int    `json:"master_plan_units"`
		RespLevelCd                int    `json:"resp_level_cd"`
		AltRateScheduleNo          int    `json:"alt_rate_schedule_no"`
		ClientAltRateScheduleID    string `json:"client_alt_rate_schedule_id"`
		BillDay                    int    `json:"bill_day"`
		LastBillThruDate           string `json:"last_bill_thru_date"`
		NextBillDate               string `json:"next_bill_date"`
		PlanDate                   string `json:"plan_date"`
		StatusDate                 string `json:"status_date"`
		RecurringBillingInterval   int    `json:"recurring_billing_interval"`
		UsageBillingInterval       int    `json:"usage_billing_interval"`
		RecurringBillingPeriodType int    `json:"recurring_billing_period_type"`
		UsageBillingPeriodType     int    `json:"usage_billing_period_type"`
		InitialPlanStatus          int    `json:"initial_plan_status"`
		DunningState               int    `json:"dunning_state"`
		MasterPlanProductFields    []struct {
			FieldName  string `json:"field_name"`
			FieldValue string `json:"field_value"`
		} `json:"master_plan_product_fields"`
		MasterPlansServices []struct {
			ServiceNo       int    `json:"service_no"`
			ClientServiceID string `json:"client_service_id"`
			TaxInclusiveInd int    `json:"tax_inclusive_ind"`
		} `json:"master_plans_services"`
	} `json:"master_plans_info"`
	ConsumerAcctInd int    `json:"consumer_acct_ind"`
	AcctLocaleNo    string `json:"acct_locale_no"`
	AcctLocaleName  string `json:"acct_locale_name"`
	ErrorCode       int    `json:"error_code"`
	ErrorMsg        string `json:"error_msg"`
	AcctNo2         int    `json:"acct_no_2"`
	AcctLocaleNo2   int    `json:"acct_locale_no_2"`
	ChiefAcctInfo   []struct {
		ChiefAcctNo       int    `json:"chief_acct_no"`
		ChiefAcctUserID   string `json:"chief_acct_user_id"`
		ChiefClientAcctID string `json:"chief_client_acct_id"`
	} `json:"chief_acct_info"`
}

type CreateBillingAccountStruct struct {
	Registered         bool   `json:"registered"`
	HaveCloudAccount   bool   `json:"haveCloudAccount"`
	HaveBillingAccount bool   `json:"haveBillingAccount"`
	Enrolled           bool   `json:"enrolled"`
	Action             string `json:"action"`
	CloudAccountID     string `json:"cloudAccountId"`
}

type CreateCloudAccountResponse struct {
	BillingAccountCreated  bool   `json:"billingAccountCreated"`
	CreditsDepleted        string `json:"creditsDepleted"`
	Delinquent             bool   `json:"delinquent"`
	Enrolled               bool   `json:"enrolled"`
	LowCredits             bool   `json:"lowCredits"`
	Name                   string `json:"name"`
	Oid                    string `json:"oid"`
	Owner                  string `json:"owner"`
	ParentID               string `json:"parentId"`
	TerminateMessageQueued bool   `json:"terminateMessageQueued"`
	TerminatePaidServices  bool   `json:"terminatePaidServices"`
	PaidServicesAllowed    bool   `json:"paidServicesAllowed"`
	Tid                    string `json:"tid"`
	Type                   string `json:"type"`
}

type CreateCloudCreditsStruct struct {
	AmountUsed      int    `json:"amountUsed"`
	CloudAccountID  string `json:"cloudAccountId"`
	CouponCode      string `json:"couponCode"`
	Created         string `json:"created"`
	Expiration      string `json:"expiration"`
	OriginalAmount  int    `json:"originalAmount"`
	Reason          string `json:"reason"`
	RemainingAmount int    `json:"remainingAmount"`
}

//Aria
type AriaClient struct {
	serverUrl string
	apiPrefix string
	// Todo: Fix for production.
	insecureSsl bool
}

type AriaCredentials struct {
	clientNo int64
	authKey  string
}

func NewAriaCredentials(clientNo int64, authKey string) *AriaCredentials {
	return &AriaCredentials{
		clientNo: clientNo,
		authKey:  authKey,
	}
}

// Coupon structs

type CreateCouponStruct struct {
	Amount  int64  `json:"amount"`
	Creator string `json:"creator"`
	Expires string `json:"expires"`
	NumUses int64  `json:"numUses"`
	Start   string `json:"start"`
}

type StandardCreateCouponStruct struct {
	Amount     int64  `json:"amount"`
	Creator    string `json:"creator"`
	Expires    string `json:"expires"`
	NumUses    int64  `json:"numUses"`
	Start      string `json:"start"`
	IsStandard bool   `json:"isStandard"`
}

type CreateCouponResponse struct {
	Amount      int    `json:"amount"`
	Code        string `json:"code"`
	Created     string `json:"created"`
	Creator     string `json:"creator"`
	Disabled    string `json:"disabled"`
	Expires     string `json:"expires"`
	NumRedeemed int    `json:"numRedeemed"`
	NumUses     int    `json:"numUses"`
	Redemptions []struct {
		CloudAccountID string `json:"cloudAccountId"`
		Code           string `json:"code"`
		Installed      bool   `json:"installed"`
		Redeemed       string `json:"redeemed"`
	} `json:"redemptions"`
	Start string `json:"start"`
}

type CreateCouponErrorResponse struct {
	Code    int `json:"code"`
	Details []struct {
		Type            string `json:"@type"`
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"details"`
	Message string `json:"message"`
}

type DisableCouponStruct struct {
	Code     string `json:"code"`
	Disabled string `json:"disabled"`
}

type DisableCouponErrorResponse struct {
	Code    int `json:"code"`
	Details []struct {
		Type            string `json:"@type"`
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"details"`
	Message string `json:"message"`
}

type RedeemCouponStruct struct {
	CloudAccountID string `json:"cloudAccountId"`
	Code           string `json:"code"`
}

type RedeemCouponErrorResponse struct {
	Code    int `json:"code"`
	Details []struct {
		Type            string `json:"@type"`
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"details"`
	Message string `json:"message"`
}

type GetCouponResponse struct {
	Error struct {
		Code    int `json:"code"`
		Details []struct {
			Type            string `json:"@type"`
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"details"`
		Message string `json:"message"`
	} `json:"error"`
	Result struct {
		Amount      int    `json:"amount"`
		Code        string `json:"code"`
		Created     string `json:"created"`
		Creator     string `json:"creator"`
		Disabled    string `json:"disabled"`
		Expires     string `json:"expires"`
		NumRedeemed int    `json:"numRedeemed"`
		NumUses     int    `json:"numUses"`
		Redemptions []struct {
			CloudAccountID string `json:"cloudAccountId"`
			Code           string `json:"code"`
			Installed      bool   `json:"installed"`
			Redeemed       string `json:"redeemed"`
		} `json:"redemptions"`
		Start string `json:"start"`
	} `json:"result"`
}

type GetCouponErrorResponse struct {
	Code    int `json:"code"`
	Details []struct {
		Type            string `json:"@type"`
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"details"`
	Message string `json:"message"`
}

type InvoiceResponse struct {
	LastUpdated string `json:"lastUpdated"`
	Invoices    []struct {
		CloudAccountID string      `json:"cloudAccountId"`
		ID             string      `json:"id"`
		Total          int         `json:"total"`
		Paid           int         `json:"paid"`
		Due            int         `json:"due"`
		Start          string      `json:"start"`
		End            string      `json:"end"`
		InvoiceDate    interface{} `json:"invoiceDate"`
		DueDate        interface{} `json:"dueDate"`
		NotifyDate     interface{} `json:"notifyDate"`
		PaidDate       interface{} `json:"paidDate"`
		BillingPeriod  string      `json:"billingPeriod"`
		Status         string      `json:"status"`
		StatementLink  string      `json:"statementLink"`
		DownloadLink   string      `json:"downloadLink"`
	} `json:"invoices"`
}

type UsageResponse struct {
	TotalAmount int           `json:"totalAmount"`
	LastUpdated string        `json:"lastUpdated"`
	DownloadURL string        `json:"downloadUrl"`
	Period      string        `json:"period"`
	Usages      []interface{} `json:"usages"`
	TotalUsage  int           `json:"totalUsage"`
}
