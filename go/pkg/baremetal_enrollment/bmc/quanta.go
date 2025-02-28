// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"net/http"

	helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
)

type QuantaIPMIInterface struct {
	IpmiOverLAN     int `json:"ipmi_over_LAN"`
	IpmiOverKcs     int `json:"ipmi_over_kcs"`
	IsKcsIfcSupport int `json:"is_kcs_ifc_support"`
}

type QuantaUser struct {
	ID                           int    `json:"id"`
	Channel                      int    `json:"channel"`
	ChannelType                  int    `json:"channel_type"`
	Userid                       int    `json:"userid"`
	Name                         string `json:"name"`
	Access                       int    `json:"access"`
	AccessByChannel              string `json:"accessByChannel"`
	Kvm                          int    `json:"kvm"`
	Vmedia                       int    `json:"vmedia"`
	Snmp                         int    `json:"snmp"`
	PrevSnmp                     int    `json:"prev_snmp"`
	Privilege                    string `json:"privilege"`
	PrivilegeByChannel           string `json:"privilegeByChannel"`
	FixedUserCount               int    `json:"fixed_user_count"`
	SnmpAccess                   string `json:"snmp_access"`
	OEMProprietaryLevelPrivilege int    `json:"OEMProprietary_level_Privilege"`
	SnmpAuthenticationProtocol   string `json:"snmp_authentication_protocol"`
	SnmpPrivacyProtocol          string `json:"snmp_privacy_protocol"`
	EmailID                      string `json:"email_id"`
	EmailFormat                  string `json:"email_format"`
	SSHKey                       string `json:"ssh_key"`
	CreationTime                 int    `json:"creation_time"`
}

type QuantaNewUser struct {
	ID                           int    `json:"id"`
	Channel                      int    `json:"channel"`
	ChannelType                  int    `json:"channel_type"`
	Userid                       int    `json:"userid"`
	Name                         string `json:"name"`
	Access                       int    `json:"access"`
	AccessByChannel              string `json:"accessByChannel"`
	Kvm                          int    `json:"kvm"`
	Vmedia                       int    `json:"vmedia"`
	Snmp                         int    `json:"snmp"`
	PrevSnmp                     int    `json:"prev_snmp"`
	Privilege                    string `json:"privilege"`
	PrivilegeByChannel           string `json:"privilegeByChannel"`
	FixedUserCount               int    `json:"fixed_user_count"`
	SnmpAccess                   string `json:"snmp_access"`
	OEMProprietaryLevelPrivilege int    `json:"OEMProprietary_level_Privilege"`
	SnmpAuthenticationProtocol   string `json:"snmp_authentication_protocol"`
	SnmpPrivacyProtocol          string `json:"snmp_privacy_protocol"`
	EmailID                      string `json:"email_id"`
	EmailFormat                  string `json:"email_format"`
	SSHKey                       string `json:"ssh_key"`
	CreationTime                 int    `json:"creation_time"`
	Changepassword               int    `json:"changepassword"`
	UserOperation                int    `json:"UserOperation"`
	LoggedinPassword             string `json:"loggedin_password"`
	Password                     string `json:"password"`
	ConfirmPassword              string `json:"confirm_password"`
	PasswordSize                 string `json:"password_size"`
}

type UpdateUserCompletionCode struct {
	Cc int `json:"cc"`
}

const (
	quantaD54Q2U = `^QuantaGrid D54Q-2U`
)

var _ Interface = (*QuantaBMC)(nil)

type QuantaBMC struct {
	BMC
}

func (c *QuantaBMC) GetHostMACAddress(ctx context.Context) (string, error) {
	return "", nil
}

// GetHostCPU returns the current available CPU of the host
func (c *QuantaBMC) GetHostCPU(ctx context.Context) (*CPUInfo, error) {
	log := log.FromContext(ctx).WithName("BMC.GetHostCPU")
	log.Info("Getting host's CPU information")
	proccessors, err := c.getProcessors()
	if err != nil {
		return nil, err
	}
	cpuInfo, err := c.getHostCPUInfo(proccessors)
	if err != nil {
		return nil, err
	}
	// TODO, currently EMR CPU ID is set to 0x00000
	cpuInfo.CPUID = "0x606A8"
	return cpuInfo, nil
}

func (c *QuantaBMC) EnableKCS(ctx context.Context) (err error) {
	return c.setKcs(ctx, true)
}

func (c *QuantaBMC) DisableKCS(ctx context.Context) (err error) {
	return c.setKcs(ctx, false)
}

func (c *QuantaBMC) UpdateAccount(ctx context.Context, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.UpdateAccounts")
	log.Info("quanta update accounts")
	accounts, err := c.getAccounts(ctx)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.Name == newUserName {
			// Update account
			updateUser := QuantaNewUser{}
			updateUser.Name = newUserName
			updateUser.Password = newPassword
			updateUser.ConfirmPassword = newPassword
			updateUser.ID = account.ID
			err = c.patchAccount(ctx, &updateUser)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return ErrAccountNotFound
}

func (c *QuantaBMC) CreateAccount(ctx context.Context, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.createAccount")
	log.Info("quanta create accounts")
	accounts, err := c.getAccounts(ctx)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.Access == 0 && account.Name == "" {
			// Assume this is an empty spot
			newUser := QuantaNewUser{}
			newUser.Name = newUserName
			newUser.Password = newPassword
			newUser.ConfirmPassword = newPassword
			newUser.ID = account.ID
			err = c.patchAccount(ctx, &newUser)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("failed to find an free account ID")
}

func (c *QuantaBMC) getAccounts(ctx context.Context) ([]QuantaUser, error) {

	log := log.FromContext(ctx).WithName("BaremetalEnrollment.getAccounts")
	log.Info("settings/users")
	loginSession := helper.ApiSession{}
	cookies, err := helper.HttpLogin(ctx, &loginSession, c.config.Username, c.config.Password, c.config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to login %v", err)
	}
	defer helper.HttpLogout(ctx, loginSession.CSRFToken, cookies, c.config.URL)
	httpResult, _, err := helper.HttpRequest(ctx, "settings/users", http.MethodGet, nil, loginSession.CSRFToken, cookies, c.config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts %v", err)
	}
	// Unmarshal response
	log.Info(string(httpResult))
	resultUsers := []QuantaUser{}
	err = json.Unmarshal(httpResult, &resultUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get  QuantaUsers: %v", err)
	}
	return resultUsers, nil
}

func (c *QuantaBMC) patchAccount(ctx context.Context, account *QuantaNewUser) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.patchAccount")
	account.ChannelType = 4
	account.PasswordSize = "bytes_16"
	account.LoggedinPassword = ""
	account.Access = 1
	account.Channel = 1
	account.AccessByChannel = "(1,0)"
	account.Kvm = 0
	account.PrevSnmp = 0
	account.Snmp = 0
	account.Privilege = "administrator"
	account.PrivilegeByChannel = "(administrator,none)"
	account.SnmpAccess = "read_only"
	account.OEMProprietaryLevelPrivilege = 0
	account.SnmpAuthenticationProtocol = "sha256"
	account.FixedUserCount = 2
	account.SnmpPrivacyProtocol = "des"
	account.EmailID = ""
	account.EmailFormat = "ami_format"
	account.SSHKey = "Not Available"
	account.CreationTime = 0
	account.Changepassword = 1
	account.UserOperation = 0
	// Patch user
	log.Info("quanta patch accounts", "settings/users/", account.ID)
	loginSession := helper.ApiSession{}
	cookies, err := helper.HttpLogin(ctx, &loginSession, c.config.Username, c.config.Password, c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to login %v", err)
	}
	defer helper.HttpLogout(ctx, loginSession.CSRFToken, cookies, c.config.URL)
	httpBody, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf(" %v", err)
	}
	httpResult, _, err := helper.HttpRequest(ctx, fmt.Sprintf("settings/users/%d", account.ID), http.MethodPut, httpBody, loginSession.CSRFToken, cookies, c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to send update account ID %v", err)
	}
	log.Info(string(httpResult))
	results := UpdateUserCompletionCode{}
	err = json.Unmarshal(httpResult, &results)
	if err != nil {
		return fmt.Errorf("failed to patch account: %v", err)
	}
	return nil
}

func (c *QuantaBMC) setKcs(ctx context.Context, enableKcs bool) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.setKcs")
	log.Info("setKcs")
	loginSession := helper.ApiSession{}
	cookies, err := helper.HttpLogin(ctx, &loginSession, c.config.Username, c.config.Password, c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to login %v", err)
	}
	defer helper.HttpLogout(ctx, loginSession.CSRFToken, cookies, c.config.URL)
	ipmiOverKcs := QuantaIPMIInterface{}
	switch enableKcs {
	case true:
		ipmiOverKcs.IpmiOverKcs = 1
	default:
		ipmiOverKcs.IpmiOverKcs = 0
	}
	ipmiOverKcs.IpmiOverLAN = 1
	ipmiOverKcs.IsKcsIfcSupport = 1
	// convert json payload to bytes
	httpBody, err := json.Marshal(ipmiOverKcs)
	if err != nil {
		return fmt.Errorf(" %v", err)
	}
	httpResult, _, err := helper.HttpRequest(ctx, "settings/ipmi_disable_interfaces", http.MethodPost, httpBody, loginSession.CSRFToken, cookies, c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to send ipmi_disable_interfaces %v", err)
	}
	// Unmarshal response
	log.Info(string(httpResult))
	resultIpmiOverKcs := QuantaIPMIInterface{}
	err = json.Unmarshal(httpResult, &resultIpmiOverKcs)
	if err != nil {
		return fmt.Errorf("failed unmarshall resultIpmiOverKcs: %v", err)
	}
	if resultIpmiOverKcs.IpmiOverKcs != ipmiOverKcs.IpmiOverKcs {
		return fmt.Errorf("failed to set KCS, values don't match expected %d vs result %d", ipmiOverKcs.IpmiOverKcs, resultIpmiOverKcs.IpmiOverKcs)
	}
	return nil
}
