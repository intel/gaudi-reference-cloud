package testsetup

import (
	"time"
)

//User structs

type Product struct {
	Highlight    string `json:"highlight"`
	Disks        string `json:"disks"`
	InstanceType string `json:"instanceType"`
	Memory       string `json:"memory"`
	Region       string `json:"region"`
	Rate         string `json:"rate"`
	UsageExp     string `json:"usageExp"`
	Unit         string `json:"unit"`
}

type VmStruct struct {
	InstanceName   string  `json:"instanceName"`
	InstanceId     string  `json:"instanceId"`
	SshKeyname     string  `json:"sshKeyname"`
	Vnet           string  `json:"vnet"`
	CloudAccountId string  `json:"cloudAccountId"`
	CreationTime   string  `json:"creationTime"`
	MachineIp      string  `json:"machineIp"`
	Product        Product `json:"product"`
}

type Coupons struct {
	Code   string `json:"code"`
	Amount int    `json:"amount"`
}

type UserData struct {
	AccType        string     `json:"accType"`
	CloudAccountId string     `json:"cloudAccountId"`
	TotalCredits   int        `json:"totalCredits"`
	Vms            []VmStruct `json:"vms"`
	Coupons        []Coupons  `json:"coupons"`
}

// Cloud Accounts

type CreateCloudAccountEnrollStruct struct {
	Premium bool `json:"premium"`
}

type CreateCAccEnrollResponse struct {
	Registered         bool   `json:"registered"`
	HaveCloudAccount   bool   `json:"haveCloudAccount"`
	HaveBillingAccount bool   `json:"haveBillingAccount"`
	Enrolled           bool   `json:"enrolled"`
	Action             string `json:"action"`
	Id                 string `json:"cloudAccountId"`
	CAccType           string `json:"cloudAccountType"`
}

// Instance

type CreateInstance struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone  string   `json:"availabilityZone"`
		InstanceType      string   `json:"instanceType"`
		MachineImage      string   `json:"machineImage"`
		RunStrategy       string   `json:"runStrategy"`
		SSHPublicKeyNames []string `json:"sshPublicKeyNames"`
		Interfaces        []struct {
			Name string `json:"name"`
			VNet string `json:"vNet"`
		} `json:"interfaces"`
	} `json:"spec"`
}

type InstanceCreateResponseStruct struct {
	Metadata struct {
		CloudAccountID    string `json:"cloudAccountId"`
		CreationTimestamp string `json:"creationTimestamp"`
		DeletionTimestamp string `json:"deletionTimestamp"`
		Labels            struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"labels"`
		Name            string `json:"name"`
		ResourceID      string `json:"resourceId"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone string `json:"availabilityZone"`
		InstanceType     string `json:"instanceType"`
		Interfaces       []struct {
			Name string `json:"name"`
			VNet string `json:"vNet"`
		} `json:"interfaces"`
		MachineImage      string   `json:"machineImage"`
		RunStrategy       string   `json:"runStrategy"`
		SSHPublicKeyNames []string `json:"sshPublicKeyNames"`
	} `json:"spec"`
	Status struct {
		Interfaces []struct {
			Addresses    []string `json:"addresses"`
			DNSName      string   `json:"dnsName"`
			Gateway      string   `json:"gateway"`
			Name         string   `json:"name"`
			PrefixLength int      `json:"prefixLength"`
			Subnet       string   `json:"subnet"`
			VNet         string   `json:"vNet"`
		} `json:"interfaces"`
		Message  string `json:"message"`
		Phase    string `json:"phase"`
		SSHProxy struct {
			ProxyAddress string `json:"proxyAddress"`
			ProxyPort    int    `json:"proxyPort"`
			ProxyUser    string `json:"proxyUser"`
		} `json:"sshProxy"`
		UserName string `json:"userName"`
	} `json:"status"`
}

// Vnet

type VnetCreationResponse struct {
	Metadata struct {
		CloudAccountID string `json:"cloudAccountId"`
		Name           string `json:"name"`
		ResourceID     string `json:"resourceId"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone string `json:"availabilityZone"`
		Gateway          string `json:"gateway"`
		PrefixLength     int    `json:"prefixLength"`
		Region           string `json:"region"`
		Subnet           string `json:"subnet"`
	} `json:"spec"`
}

//SSH Key

type SSHKeyCreationResponse struct {
	Metadata struct {
		CloudAccountID string `json:"cloudAccountId"`
		Name           string `json:"name"`
		ResourceID     string `json:"resourceId"`
		Labels         struct {
		} `json:"labels"`
		CreationTimestamp string `json:"creationTimestamp"`
	} `json:"metadata"`
	Spec struct {
		SSHPublicKey string `json:"sshPublicKey"`
	} `json:"spec"`
}

// Coupons

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

type RedeemCouponPayload struct {
	CloudAccountID string `json:"cloudAccountId"`
	Code           string `json:"code"`
}

//Product Catalog

type ProductFilter struct {
	Name string `json:"name"`
}

type GetProductsResponse struct {
	Products []struct {
		Name        string    `json:"name"`
		ID          string    `json:"id"`
		Created     time.Time `json:"created"`
		VendorID    string    `json:"vendorId"`
		FamilyID    string    `json:"familyId"`
		Description string    `json:"description"`
		Metadata    struct {
			Category     string `json:"category"`
			Disks        string `json:"disks.size"`
			DisplayName  string `json:"displayName"`
			Desc         string `json:"family.displayDescription"`
			DispName     string `json:"family.displayName"`
			Highlight    string `json:"highlight"`
			Information  string `json:"information"`
			InstanceType string `json:"instanceType"`
			Memory       string `json:"memory.size"`
			Processor    string `json:"processor"`
			Region       string `json:"region"`
			Service      string `json:"service"`
		} `json:"metadata"`
		Eccn      string `json:"eccn"`
		Pcq       string `json:"pcq"`
		MatchExpr string `json:"matchExpr"`
		Rates     []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
	} `json:"products"`
}

type UsageData struct {
	Usage []ProductRunTime
}
type ProductRunTime struct {
	Rate        float64
	RunningSecs float64
	Amount      float64
	ResourceIds []string
	ProductType string
}

type Resources struct {
	ResourceId  string
	ProductName string
}

type ResourcesInfo struct {
	ProductData []Resources
}

type SchedulerStart struct {
	Action string `json:"action"`
}
