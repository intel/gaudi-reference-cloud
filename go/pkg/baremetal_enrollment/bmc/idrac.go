// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipmilan"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stmcginnis/gofish/redfish"

	"encoding/json"
	"io"
	"net/http"
)

type IDRACAccounts struct {
	OdataContext string `json:"@odata.context"`
	OdataID      string `json:"@odata.id"`
	OdataType    string `json:"@odata.type"`
	Description  string `json:"Description"`
	Members      []struct {
		OdataID string `json:"@odata.id"`
	} `json:"Members"`
	MembersOdataCount int    `json:"Members@odata.count"`
	Name              string `json:"Name"`
}

type iDRACUser struct {
	OdataContext string   `json:"@odata.context"`
	OdataEtag    string   `json:"@odata.etag"`
	OdataID      string   `json:"@odata.id"`
	OdataType    string   `json:"@odata.type"`
	AccountTypes []string `json:"AccountTypes"`
	Description  string   `json:"Description"`
	Enabled      bool     `json:"Enabled"`
	ID           string   `json:"Id"`
	Keys         struct {
		OdataID string `json:"@odata.id"`
	} `json:"Keys"`
	Links struct {
		Role struct {
			OdataID string `json:"@odata.id"`
		} `json:"Role"`
	} `json:"Links"`
	Locked          bool     `json:"Locked"`
	Name            string   `json:"Name"`
	OEMAccountTypes []string `json:"OEMAccountTypes"`
	Oem             struct {
		Dell struct {
			OdataType               string `json:"@odata.type"`
			SNMPv3PassphraseEnabled string `json:"SNMPv3PassphraseEnabled"`
		} `json:"Dell"`
	} `json:"Oem"`
	Password               interface{} `json:"Password"`
	PasswordChangeRequired bool        `json:"PasswordChangeRequired"`
	PasswordExpiration     interface{} `json:"PasswordExpiration"`
	RoleID                 string      `json:"RoleId"`
	SNMP                   struct {
		AuthenticationKey      interface{} `json:"AuthenticationKey"`
		AuthenticationKeySet   bool        `json:"AuthenticationKeySet"`
		AuthenticationProtocol string      `json:"AuthenticationProtocol"`
		EncryptionKey          interface{} `json:"EncryptionKey"`
		EncryptionKeySet       bool        `json:"EncryptionKeySet"`
		EncryptionProtocol     string      `json:"EncryptionProtocol"`
	} `json:"SNMP"`
	StrictAccountTypes bool   `json:"StrictAccountTypes"`
	UserName           string `json:"UserName"`
}

const (
	iDRACXE9680Gaudi3Regex    = `^PowerEdge XE9680`
	iDRACKCSNetFunc           = 0x32
	iDRACKCSCmd               = 0x02
	iDRACOSPassthroughEnable  = "Enabled"
	iDRACOSPassthroughDisable = "Disabled"
)

var _ Interface = (*IdracBMC)(nil)

type IdracBMC struct {
	BMC
}

func (c *IdracBMC) GetHostMACAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (c *IdracBMC) GetHostBMCAddress() (string, error) {
	system, err := c.getSystem()
	if err != nil {
		return "", fmt.Errorf("unable to get the computing system: %v", err)
	}
	address := fmt.Sprintf("idrac-redfish+%s%s", c.config.URL, system.ODataID())
	return address, nil
}

func (c *IdracBMC) GPUDiscovery(ctx context.Context) (count int, gpuModel string, err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.GPUDiscovery")
	log.Info("Starting GPU Discovery")
	switch c.hwType {
	case Gaudi3Dell:
		return 8, pcieToGPUTable["0x1da3:0x1060"], nil
	case Gaudi2Dell:
		return 8, pcieToGPUTable["0x1da3:0x1020"], nil
	default:
		return 0, NoGpuType, nil
	}

}

func (c *IdracBMC) EnableKCS(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.EnableKCS")
	log.Info("Enable KCS")
	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	byteData := []byte{byte(ipmilan.IDRACAllow)}
	err = ipmiHelper.RunRawCommand(ctx, byteData, iDRACKCSNetFunc, iDRACKCSCmd)
	if err != nil {
		return fmt.Errorf("unable to Enable KCS: %v", err)
	}
	return nil
}

func (c *IdracBMC) DisableKCS(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.DisableKCS")
	log.Info("Disable KCS")
	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	byteData := []byte{byte(ipmilan.IDRACDisable)}
	err = ipmiHelper.RunRawCommand(ctx, byteData, iDRACKCSNetFunc, iDRACKCSCmd)
	if err != nil {
		return fmt.Errorf("unable to Disable KCS: %v", err)
	}
	return nil
}

func (c *IdracBMC) CreateAccount(ctx context.Context, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("BMC.CreateBMCCredentials")
	log.Info("Dell Creating BMC account")
	// find an available account ID
	accountID, err := c.getAvailableUserID()
	if err != nil {
		return err
	}
	// Need to add the new Admin Account
	payload := map[string]interface{}{
		"UserName": newUserName,
		"Password": newPassword,
		"RoleId":   "Administrator",
		"Enabled":  true,
	}

	// POST the request to the Redfish API
	url := fmt.Sprintf("/redfish/v1/Managers/iDRAC.Embedded.1/Accounts/%s", accountID)
	response, err := c.GetClient().Patch(url, payload)
	if err != nil {
		return fmt.Errorf("failed to create new Admin account %q on BMC URL %q: %v", newUserName, c.config.URL, err)
	}
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		log.Info("BMC Admin account created! Updating runtime values.", "newUserName", newUserName)
	} else {
		log.Info("BMC account creation failed", "bmcUsername", newUserName)
		return fmt.Errorf("failed to create new Admin account %q on BMC URL %q: %d: %q",
			newUserName, c.config.URL, response.StatusCode, http.StatusText(response.StatusCode))
	}
	// Enable IPMI oveLan
	payload = map[string]interface{}{
		fmt.Sprintf("Users.%s.IpmiLanPrivilege", accountID):    "Administrator",
		fmt.Sprintf("Users.%s.IpmiSerialPrivilege", accountID): "Administrator",
		fmt.Sprintf("Users.%s.SolEnable", accountID):           "Enabled",
	}
	attributes := map[string]interface{}{
		"Attributes": payload,
	}
	response, err = c.GetClient().Patch("/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", attributes)
	if err != nil {
		return fmt.Errorf("failed to update new Admin account %q with IPMI access on BMC URL %q: %v", newUserName, c.config.URL, err)
	}
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		log.Info("BMC Admin Added IPMI over lan access", "newUserName", newUserName)
	} else {
		log.Info("BMC account Attributes Updates failed", "bmcUsername", newUserName)
		return fmt.Errorf("failed to update new Admin account IPMI over Lan %q on BMC URL %q: %d: %q",
			newUserName, c.config.URL, response.StatusCode, http.StatusText(response.StatusCode))
	}
	return nil
}

func (c *IdracBMC) getAvailableUserID() (string, error) {
	//Check first for available ID
	response, err := c.GetClient().Get("/redfish/v1/Managers/iDRAC.Embedded.1/Accounts")
	if err != nil {
		return "", fmt.Errorf("failed to get Accounts  %v", err)
	}
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Get Accounts BMC System response: %v", err)
	}
	accounts := IDRACAccounts{}
	err = json.Unmarshal(resBody, &accounts)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal open BMC pcieFunction: %v", err)
	}
	for _, odataID := range accounts.Members {
		res, err := c.GetClient().Get(odataID.OdataID)
		if err != nil {
			return "", fmt.Errorf("failed to get account members  %v", err)
		}
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read Get single account, response: %v", err)
		}
		bmcUser := iDRACUser{}
		err = json.Unmarshal(resBody, &bmcUser)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal open BMC pcieFunction: %v", err)
		}
		if bmcUser.UserName == "" && bmcUser.ID != "1" {
			return bmcUser.ID, nil
		}
	}
	return "", fmt.Errorf("failed to find an available user ID")
}

func (c *IdracBMC) GetHostCPU(ctx context.Context) (*CPUInfo, error) {
	log := log.FromContext(ctx).WithName("BMC.GetHostCPU")
	log.Info("Getting host's CPU information")
	proccessors, err := c.getProcessors()
	if err != nil {
		return nil, err
	}
	return c.getHostCPUInfo(proccessors)
}

func (c *IdracBMC) getProcessors() ([]*redfish.Processor, error) {
	system, err := c.getSystem()
	if err != nil {
		return nil, err
	}

	// get Processors
	response, err := c.GetClient().Get(fmt.Sprintf("%s/Processors", system.ODataID()))
	if err != nil {
		return nil, fmt.Errorf("failed to get Processors %v", err)
	}
	var oData BMCOdata
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Processors %v", err)
	}
	err = json.Unmarshal(resBody, &oData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BMC processors Odata: %v", err)
	}

	// Check memebers for CPU only
	members := oData.Members
	var processors []*redfish.Processor
	for _, processor := range members {
		if !strings.Contains(processor.OdataID, "CPU.Socket") {
			continue
		}
		response, err := c.GetClient().Get(processor.OdataID)
		if err != nil {
			return nil, fmt.Errorf("failed to get processor %v", err)
		}
		var processor redfish.Processor
		defer response.Body.Close()
		resBody, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read Processors %v", err)
		}
		err = json.Unmarshal(resBody, &processor)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal BMC processor Odata: %v", err)
		}
		processors = append(processors, &processor)
	}
	return processors, nil
}

// Enable/Disable OS to Idrac passthrough
func (c *IdracBMC) EnableHCI(ctx context.Context) (err error) {
	return c.setOSToIDRACPassthroughAdminState(ctx, iDRACOSPassthroughEnable)
}

func (c *IdracBMC) DisableHCI(ctx context.Context) (err error) {
	return c.setOSToIDRACPassthroughAdminState(ctx, iDRACOSPassthroughDisable)
}

func (c *IdracBMC) setOSToIDRACPassthroughAdminState(ctx context.Context, state string) (err error) {
	log := log.FromContext(ctx).WithName("BMC.IDRAC.setOSToIDRACPassthroughAdminState")
	log.Info("set OS to IDRAC passthrough admin state")
	serversPayload := map[string]interface{}{
		"Attributes": map[string]string{
			"OS-BMC.1.AdminState": state,
		},
	}

	resp, err := c.GetClient().Patch("/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", serversPayload)
	if err != nil {
		return fmt.Errorf("failed to set  OS to IDRAC Passthrough admin state to %s with err %v", state, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code %d when setting the OS to IDRAC Passthrough admin state to %s. expected status code is 200", resp.StatusCode, state)
	}
	//validate if the change is applied
	time.Sleep(2 * time.Second)

	resp, err = c.GetClient().Get("/redfish/v1/Managers/iDRAC.Embedded.1/Attributes")
	if err != nil {
		return fmt.Errorf("failed to get IDRAC Attributes with err %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code %d when getting IDRAC Attributes. expected status code is 200", resp.StatusCode)
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read idrac attributes response: %v", err)
	}
	var IDRACAttributes map[string]interface{}
	err = json.Unmarshal(resBody, &IDRACAttributes)
	if err != nil {
		return fmt.Errorf("error unmarshalling IDRAC attributes response JSON: %v", err)
	}
	// Access a specific value from the dynamic result
	if attributes, ok := IDRACAttributes["Attributes"].(map[string]interface{}); ok {
		if osBMCAdminStateValue, exists := attributes["OS-BMC.1.AdminState"].(string); exists {
			log.Info("OS-BMC.1.AdminState", "OS-BMC.1.AdminState", osBMCAdminStateValue)
			if osBMCAdminStateValue == state {
				log.Info("OS-BMC.1.AdminState is updated", "currentState", osBMCAdminStateValue)
			} else {
				return fmt.Errorf("failed to update OS-BMC.1.AdminState value. currentState: %s, expectedState: %s", osBMCAdminStateValue, state)
			}
		} else {
			return fmt.Errorf("OS-BMC.1.AdminState key not found in response")
		}
	} else {
		return fmt.Errorf("'Attributes' key not found in IDRAC attributes response")
	}
	return nil
}
