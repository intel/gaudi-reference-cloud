package cloudAccounts

type CreateCloudAccountStruct struct {
	BillingAccountCreated  bool   `json:"billingAccountCreated"`
	CreditsDepleted        string `json:"creditsDepleted"`
	CountryCode            string `json:"countryCode"`
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
	Restricted             bool   `json:"restricted"`
	TradeRestricted        bool   `json:"tradeRestricted"`
	PersonID               string `json:"personId"`
}

type CreateCloudAccountOIDCStruct struct {
	BillingAccountCreated  bool   `json:"billingAccountCreated"`
	CountryCode            string `json:"countryCode"`
	CreditsDepleted        string `json:"creditsDepleted"`
	Delinquent             bool   `json:"delinquent"`
	Enrolled               bool   `json:"enrolled"`
	LowCredits             bool   `json:"lowCredits"`
	Name                   string `json:"name"`
	Oid                    string `json:"oid"`
	Owner                  string `json:"owner"`
	ParentID               string `json:"parentId"`
	PersonID               string `json:"personId"`
	TerminateMessageQueued bool   `json:"terminateMessageQueued"`
	TerminatePaidServices  bool   `json:"terminatePaidServices"`
	PaidServicesAllowed    bool   `json:"paidServicesAllowed"`
	Premium                bool   `json:"premium"`
	Tid                    string `json:"tid"`
	Type                   string `json:"type"`
}

type DeleteCloudAccountStruct struct {
	Id string `json:"id"`
}

type MembersCAccStruct struct {
	Members []string `json:"members"`
}

type GetCloudAccountResponse struct {
	ID                     string `json:"id"`
	Tid                    string `json:"tid"`
	Oid                    string `json:"oid"`
	ParentID               string `json:"parentId"`
	Created                string `json:"created"`
	Name                   string `json:"name"`
	Owner                  string `json:"owner"`
	Type                   string `json:"type"`
	BillingAccountCreated  bool   `json:"billingAccountCreated"`
	Enrolled               bool   `json:"enrolled"`
	LowCredits             bool   `json:"lowCredits"`
	CreditsDepleted        string `json:"creditsDepleted"`
	TerminatePaidServices  bool   `json:"terminatePaidServices"`
	TerminateMessageQueued bool   `json:"terminateMessageQueued"`
	Delinquent             bool   `json:"delinquent"`
	PaidServicesAllowed    bool   `json:"paidServicesAllowed"`
	PersonID               string `json:"personId"`
	CountryCode            string `json:"countryCode"`
	Restricted             bool   `json:"restricted"`
	AdminName              string `json:"adminName"`
	AccessLimitedTimestamp string `json:"accessLimitedTimestamp"`
	TradeRestricted        bool   `json:"tradeRestricted"`
	UpgradedToPremium      string `json:"upgradedToPremium"`
	UpgradedToEnterprise   string `json:"upgradedToEnterprise"`
}

type GetCAccResponse2 struct {
	Result struct {
		Id                     string `json:"id"`
		Tid                    string `json:"tid"`
		Oid                    string `json:"oid"`
		ParentID               string `json:"parentId"`
		Created                string `json:"created"`
		Name                   string `json:"name"`
		Owner                  string `json:"owner"`
		Type                   string `json:"type"`
		BillingAccountCreated  bool   `json:"billingAccountCreated"`
		Enrolled               bool   `json:"enrolled"`
		LowCredits             bool   `json:"lowCredits"`
		CreditsDepleted        string `json:"creditsDepleted"`
		TerminateMessageQueued bool   `json:"terminateMessageQueued"`
		TerminatePaidServices  bool   `json:"terminatePaidServices"`
		Delinquent             bool   `json:"delinquent"`
		PaidServicesAllowed    bool   `json:"paidServicesAllowed"`
	} `json:"result"`
}

type CreateCloudAccountResponse struct {
	Id string `json:"id"`
}

type GetMemberByIdResponse struct {
	CloudAccountId string   `json:"cloudAccountId"`
	Members        []string `json:"members"`
}

type CreateCloudAccountEnrollStruct struct {
	Premium     bool `json:"premium"`
	TermsStatus bool `json:"termsStatus"`
}

type CreateCAccEnrollResponse struct {
	Registered         bool   `json:"registered"`
	HaveCloudAccount   bool   `json:"haveCloudAccount"`
	HaveBillingAccount bool   `json:"haveBillingAccount"`
	Enrolled           bool   `json:"enrolled"`
	Action             string `json:"action"`
	CloudAccountID     string `json:"cloudAccountId"`
	CloudAccountType   string `json:"cloudAccountType"`
	IsMember           bool   `json:"isMember"`
}

type DuplicateOidStruct struct {
	Oid string `json:"oid"`
}
