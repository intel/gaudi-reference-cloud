// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestInterfaceShortToLongName(t *testing.T) {
	tests := []struct {
		inputportname string
		expect        string
		expectError   bool
	}{
		{"Et1", "Ethernet1", false},
		{"Et1/1", "Ethernet1/1", false},
		{"Et1/5", "Ethernet1/5", false},
		{"Et23", "Ethernet23", false},
		{"Et23/1", "Ethernet23/1", false},
		{"Et34/12", "Ethernet34/12", false},
		{"Et1/2/3", "Ethernet1/2/3", false},
		{"Po5", "Port-Channel5", false},
		{"Vx5", "Vxlan5", false},

		{"1", "", true},
		{"1/2", "", true},
		{"Et 1", "", true},
		{"Et", "", true},
		{"NotEt1", "", true},
		{"(Et 1)", "", true},
		{"Et1a", "", true},
		{"Et 1/1", "", true},
		{"Vx5/6", "", true},
		{"Po5/6", "", true},
	}

	for _, test := range tests {
		result, err := InterfaceShortToLongName(test.inputportname)
		if err != nil && test.expectError == false {
			t.Errorf("For input %s, expected success but got error: %v", test.inputportname, err)
			continue
		}

		if err == nil && test.expectError == true {
			t.Errorf("For input %s, expected error: %v, but got no error", test.inputportname, test.expect)
			continue
		}

		if result != test.expect {
			t.Errorf("For input %s, expected %s, but got: %s", test.inputportname, test.expect, result)
		}
	}
}

func TestExpandVlanRanges(t *testing.T) {
	tests := []struct {
		inputVlanRanges string
		expect          []int
		expectError     bool
	}{
		{"100,123", []int{100, 123}, false},
		{"100-105", []int{100, 101, 102, 103, 104, 105}, false},
		{"100-101", []int{100, 101}, false},
		{"100,101", []int{100, 101}, false},
		{"100,110-112,115", []int{100, 110, 111, 112, 115}, false},
		{"100-102,110-112", []int{100, 101, 102, 110, 111, 112}, false},
		{"123", []int{123}, false},

		{"100-", nil, true},
		{"abc", nil, true},
		{"100-88", nil, true},
		{"100-100", nil, true},
	}

	for _, test := range tests {
		result, err := ExpandVlanRanges(test.inputVlanRanges)
		if err != nil && test.expectError == false {
			t.Errorf("For input %s, expected success but got error: %v", test.inputVlanRanges, err)
			continue
		}

		if err == nil && test.expectError == true {
			t.Errorf("For input %s, expected error but got no error", test.inputVlanRanges)
			continue
		}

		if !reflect.DeepEqual(result, test.expect) {
			t.Errorf("For input %s, expected %v, but got: %v", test.inputVlanRanges, test.expect, result)
		}
	}
}

func TestValidateVlanValue(t *testing.T) {
	allowedVlans := []int{100, 200, 3998, 3999, 4008}

	tests := []struct {
		vlan    int
		wantErr bool
	}{
		{vlan: 100, wantErr: false},
		{vlan: 200, wantErr: false},
		{vlan: 3998, wantErr: false},
		{vlan: 3999, wantErr: false},
		{vlan: 4008, wantErr: false},
		{vlan: 99, wantErr: true},
		{vlan: 0, wantErr: true},
		{vlan: 4000, wantErr: true},
		{vlan: 4007, wantErr: true},
		{vlan: 4009, wantErr: true},
	}

	for _, test := range tests {
		err := ValidateVlanValue(test.vlan, allowedVlans)
		if (err != nil) != test.wantErr {
			t.Errorf("NewValidateVlanValue(%d, %v) error = %v, wantErr %v", test.vlan, allowedVlans, err, test.wantErr)
		}
	}
}

func TestValidateModeValue(t *testing.T) {
	allowedModes := []string{"access", "trunk"}

	tests := []struct {
		mode     string
		expected bool
	}{
		{"access", true},
		{"trunk", true},
		{"other", false},
		{"", false},
		{"ACCESS", false},
	}

	for _, test := range tests {
		err := ValidateModeValue(test.mode, allowedModes)
		if (err == nil) != test.expected {
			if test.expected {
				t.Errorf("NewValidateModeValue(%s, %v) = %v; want no error", test.mode, allowedModes, err)
			} else {
				t.Errorf("NewValidateModeValue(%s, %v) = nil; want error", test.mode, allowedModes)
			}
		}
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name           string
		vlan           int
		port           string
		mode           string
		allowedVlanIds []int
		allowedModes   []string
		wantErr        bool
	}{
		{
			name:           "valid port, vlan, and mode",
			vlan:           100,
			port:           "Ethernet27/1",
			mode:           "access",
			allowedVlanIds: []int{100, 200, 3999, 4008},
			allowedModes:   []string{"access", "trunk"},
			wantErr:        false,
		},
		{
			name:           "invalid vlan",
			vlan:           99,
			port:           "Ethernet27/1",
			mode:           "access",
			allowedVlanIds: []int{100, 200, 3999, 4008},
			allowedModes:   []string{"access", "trunk"},
			wantErr:        true,
		},
		{
			name:           "invalid mode",
			vlan:           100,
			port:           "Ethernet27/1",
			mode:           "invalid_mode",
			allowedVlanIds: []int{10, 200, 3999, 4008},
			allowedModes:   []string{"access", "trunk"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.vlan, tt.port, tt.mode, tt.allowedVlanIds, tt.allowedModes)
			if tt.wantErr && err == nil {
				t.Errorf("NewValidatePort() = %v, want error", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NewValidatePort() = %v, want no error", err)
			}
		})
	}
}

func TestGetInterfaceRange(t *testing.T) {
	tests := []struct {
		name               string
		ethernetInterfaces []string
		wantRange          string
		wantErr            bool
	}{
		{
			name:               "valid interfaces",
			ethernetInterfaces: []string{"Ethernet1", "Ethernet2", "Ethernet1/5"},
			wantRange:          "Ethernet1-Ethernet2",
			wantErr:            false,
		},
		{
			name:               "valid single interface",
			ethernetInterfaces: []string{"Ethernet1"},
			wantRange:          "Ethernet1",
			wantErr:            false,
		},
		{
			name:               "interfaces with multiple parts",
			ethernetInterfaces: []string{"Ethernet1/1", "Ethernet1/2", "Ethernet1/3"},
			wantRange:          "Ethernet1/1-Ethernet1/3",
			wantErr:            false,
		},
		{
			name:               "interfaces with multiple parts",
			ethernetInterfaces: []string{"Ethernet1/1", "Ethernet3/5", "Ethernet2/3", "Ethernet3/1"},
			wantRange:          "Ethernet1/1-Ethernet3/5",
			wantErr:            false,
		},
		{
			name:               "Mixed interfaces",
			ethernetInterfaces: []string{"Ethernet1", "Ethernet3/5", "Ethernet2/3", "Ethernet3/1"},
			wantRange:          "Ethernet1-Ethernet3/5",
			wantErr:            false,
		},
		{
			name:               "2-digit interfaces",
			ethernetInterfaces: []string{"Ethernet1", "Ethernet10", "Ethernet2", "Ethernet1/5"},
			wantRange:          "Ethernet1-Ethernet10",
			wantErr:            false,
		},
		{
			name:               "2-digit interface with slash",
			ethernetInterfaces: []string{"Ethernet1/12", "Ethernet10/12", "Ethernet2/12", "Ethernet1/5"},
			wantRange:          "Ethernet1/5-Ethernet10/12",
			wantErr:            false,
		},
		{
			name:               "invalid interface names",
			ethernetInterfaces: []string{"Invalid1", "Ethernet2"},
			wantRange:          "",
			wantErr:            true,
		},
		{
			name:               "empty interface list",
			ethernetInterfaces: []string{},
			wantRange:          "",
			wantErr:            true,
		},
		{
			name:               "mixed valid and invalid interfaces",
			ethernetInterfaces: []string{"Ethernet1", "Invalid2", "Ethernet3"},
			wantRange:          "",
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRange, err := GetInterfaceRange(tt.ethernetInterfaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInterfaceRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRange != tt.wantRange {
				t.Errorf("getInterfaceRange() = %v, want %v", gotRange, tt.wantRange)
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name   string
		ip     string
		expect error
	}{
		{"Valid IPv4", "192.168.1.1", nil},
		{"Valid IPv4", "0.0.0.0", nil},
		{"Invalid IPv4 too many octets", "192.168.1.1.1", errors.New("invalid IP address: 192.168.1.1.1")},
		{"Invalid IPv4 empty string", "", errors.New("invalid IP address: ")},
		{"Invalid IPv4 with letters", "192.168.1.a", errors.New("invalid IP address: 192.168.1.a")},
		{"Invalid IPv4 with octet > 255", "256.168.1.1", errors.New("invalid IP address: 256.168.1.1")},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", errors.New("invalid IPv4 address: 2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIP(tt.ip)
			if (err != nil && tt.expect == nil) || (err == nil && tt.expect != nil) || (err != nil && tt.expect != nil && err.Error() != tt.expect.Error()) {
				t.Errorf("ValidateIP(%s) = %v; expected %v", tt.ip, err, tt.expect)
			}
		})
	}
}

func TestValidateIpOverride(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		dc     string
		expect error
	}{
		{"Valid IP", "192.168.1.1", "", nil},
		{"Valid FQDN", "internal-placeholder.com", "", nil},
		{"Valid FQDN", "clab-frontendonly-frontend-leaf1", "clab", nil},
		{"Valid FQDN", "internal-placeholder.com", "fxhb3p3r", nil},
		{"Invalid IP", "256.168.1.1", "", errors.New("value is neither a valid IP address nor a valid hostname: 256.168.1.1")},
		{"Invalid FQDN", "fxhb3p3r.internal-placeholder.com", "", errors.New("value is neither a valid IP address nor a valid hostname: fxhb3p3r.internal-placeholder.com")},
		{"", "", "", errors.New("value is neither a valid IP address nor a valid hostname: ")},
		{"Invalid IP and FQDN", "invalid-value", "", errors.New("value is neither a valid IP address nor a valid hostname: invalid-value")},
		{"Empty string", "", "", errors.New("value is neither a valid IP address nor a valid hostname: ")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIpOverride(tt.value, tt.dc)
			if (err != nil && tt.expect == nil) || (err == nil && tt.expect != nil) || (err != nil && tt.expect != nil && err.Error() != tt.expect.Error()) {
				t.Errorf("ValidateIpOverride(%s) = %v; expected %v", tt.value, err, tt.expect)
			}
		})
	}
}

func TestGetIp(t *testing.T) {
	tests := []struct {
		name             string
		sw               *idcnetworkv1alpha1.Switch
		dc               string
		expectedErrValue string
		expectedValue    string
	}{
		{
			name: "Valid IP override for valid IP",
			sw: &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      "switch-name",
					Namespace: "test-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN:       "example.com",
					IpOverride: "192.168.1.1",
				},
			},
			dc:               "",
			expectedErrValue: "",
			expectedValue:    "192.168.1.1",
		},

		{
			name: "Valid IP override for valid FQDN",
			sw: &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      "switch-name",
					Namespace: "test-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN:       "example.com",
					IpOverride: "internal-placeholder.com",
				},
			},
			dc:               "",
			expectedErrValue: "",
			expectedValue:    "internal-placeholder.com",
		},

		{
			name: "Invalid IP override for invalid FQDN",
			sw: &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      "switch-name",
					Namespace: "test-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN:       "example.com",
					IpOverride: "fxhb3p3r.internal-placeholder.com",
				},
			},
			dc:               "",
			expectedErrValue: "fxhb3p3r.internal-placeholder.com",
			expectedValue:    "",
		},

		{
			name: "Valid IP Override if override is not set",
			sw: &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      "switch-name",
					Namespace: "test-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN: "internal-placeholder.com",
				},
			},
			dc:               "",
			expectedErrValue: "",
			expectedValue:    "internal-placeholder.com",
		},

		{
			name: "Invalid IP Override if override is not set",
			sw: &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      "switch-name",
					Namespace: "test-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN: "fxhb3p3r.internal-placeholder.com",
				},
			},
			dc:               "fxhb3p3r",
			expectedErrValue: "fxhb3p3r.internal-placeholder.com",
			expectedValue:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := GetIp(tt.sw, tt.dc)
			expectErr := tt.expectedErrValue != ""

			if expectErr {
				if err == nil {
					t.Errorf("%s: expected error, got nil", tt.name)
				} else if !strings.Contains(err.Error(), tt.expectedErrValue) {
					t.Errorf("%s: expected error to contain %v, got error %v", tt.name, tt.expectedErrValue, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: expected no error, got error %v", tt.name, err)
				} else if value != tt.expectedValue {
					t.Errorf("%s: expected value %v, got value %v", tt.name, tt.expectedValue, value)
				}
			}
		})
	}
}

func TestGeneratePortFullName(t *testing.T) {
	type args struct {
		switchFQDN string
		portName   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				switchFQDN: "internal-placeholder.com",
				portName:   "Ethernet27/1",
			},
			want: "ethernet27-1.internal-placeholder.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GeneratePortFullName(tt.args.switchFQDN, tt.args.portName); got != tt.want {
				t.Errorf("GeneratePortFullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePortValue(t *testing.T) {
	tests := []struct {
		inputportname string
		expectError   bool
	}{
		{"Ethernet1", false},
		{"Ethernet1/1", false},
		{"Ethernet1/5", false},
		{"Ethernet23", false},
		{"Ethernet10", false},
		{"Ethernet23/1", false},
		{"Ethernet34/12", false},
		{"Ethernet1/2/3", false},
		{"Port-Channel12", false},

		{"1", true},
		{"1/2", true},
		{"Ethernet 1", true},
		{"Ethernet", true},
		{"Ethernet12345", true},
		{"NotEthernet1", true},
		{"(Ethernet 1)", true},
		{"Ethernet1a", true},
		{"Ethernet 1/1", true},
	}

	for _, test := range tests {
		err := ValidatePortValue(test.inputportname)

		if test.expectError == false && err != nil {
			t.Errorf("For input (%s), expected success but got error: %v", test.inputportname, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%s), expected error: %v, but got no error: %v", test.inputportname, test.expectError, err)
		}
	}
}

func TestValidateSwitchFQDN(t *testing.T) {
	tests := []struct {
		inputfqdn       string
		inputdatacenter string
		expect          error
	}{
		{"internal-placeholder.com", "", nil},
		{"internal-placeholder.com", "", nil},
		{"internal-placeholder.com", "", nil},
		{"internal-placeholder.com", "fxhb3p3r", nil},
		{"internal-placeholder.com", "azs1101pe", nil},
		{"fxhb3p3r.internal-placeholder.com", "", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a.idcmgt.intel.net", "", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a.idcmgt.mydomain.com", "", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a.idcabc.intel.com", "", errors.New("invalid FQDN format")},
		{"fxhb3p3ab-zal0113a.idcmgt.intel.net", "", errors.New("invalid FQDN format")},
		{"abc-fxhb3p3r-zal0113a.idcmgt.intel.net", "", errors.New("invalid FQDN format")},
		{"internal-placeholder.com", "", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a-internal-placeholder.com", "", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a.idcmgt.mydomain.com", "", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a.idcmgt.mydomain.com.extra", "", errors.New("invalid FQDN format")},
		{"internal-placeholder.com", "", errors.New("invalid FQDN format")},
		{"", "", errors.New("switch FQDN was empty")},
		{"", "fxhb3p3s", errors.New("switch FQDN was empty")},
		{"", "fxhb3p3s:fxhb3p3r", errors.New("switch FQDN was empty")},
		{"internal-placeholder.com", "", errors.New("invalid FQDN format")},
		{"internal-placeholder.com", "fxhb3p3r", errors.New("invalid FQDN format")},
		{"fxhb3p3r-zal0113a", "", errors.New("invalid FQDN format")},
		{"internal-placeholder.com", "fxhb3p3s", errors.New("the switch internal-placeholder.com doesn't belong to the data center/s [fxhb3p3s]")},
		{"internal-placeholder.com", "fxhb", errors.New("the switch internal-placeholder.com doesn't belong to the data center/s [fxhb]")},
		{"internal-placeholder.com", "fxhb3p3s:fxhb3p3r", nil},
		{"internal-placeholder.com", "fxhb3p3a:fxhb3p3b", errors.New("the switch internal-placeholder.com doesn't belong to the data center/s [fxhb3p3a:fxhb3p3b]")},
		{"internal-placeholder.com", "", errors.New("invalid FQDN format")}, // Cannot use new hostname format "pdx05-c01-acsw001" with old "idcmgt" suffix.
		{"pdx11-c01-acsw001.us-staging-3.cloud.intel.com", "", nil},
		{"pdx11-a00-acsw001.us-dev-1a.cloud.intel.com", "", nil},
		{"pdx11-s01-acsw001.us-region-1.cloud.intel.com", "", nil},
		{"clab-lacp-frontend-leaf1a", "", nil},
		{"clab-this is not valid", "", errors.New("invalid FQDN format")},
		{"pdx11-a00-acsw001.us-dev-1a.fakecloud.intel.com", "", nil},
	}

	for _, test := range tests {
		err := ValidateSwitchFQDN(test.inputfqdn, test.inputdatacenter)
		if err != test.expect && test.expect == nil {
			t.Errorf("For input (%s, %s), expected success but got error: %v", test.inputfqdn, test.inputdatacenter, err)
		}

		if err != test.expect && err.Error() != test.expect.Error() {
			t.Errorf("For input (%s, %s), expected error: %v, but got: %v", test.inputfqdn, test.inputdatacenter, test.expect, err)
		}
	}
}

func TestValidatePortNumber(t *testing.T) {
	tests := []struct {
		inputportnumberstring string
		expectError           bool
	}{
		{"5", false},
		{"3", false},
		{"22", false},
		{"23/1", false},
		{"34/12", false},
		{"1/2/3", false},
		{"10", false},
		{"1/10", false},

		{"0", true},
		{"00", true},
		{"1/0", true},
		{"1/0/5", true},
		{"1.0", true},
		{"1/2.0", true},
		{"Ethernet 1", true},
		{"Ethernet", true},
		{"1e7", true},
		{"1,000", true},
		{"1234", true},
		{"12/3456", true},
		{"-6", true},
		{"1.0/1", true},
		{"1%23QNAN", true},
		{"1%23INF", true},
		{"1#INF", true},
	}

	for _, test := range tests {
		err := ValidatePortNumber(test.inputportnumberstring)

		if test.expectError == false && err != nil {
			t.Errorf("For input (%s), expected success but got error: %v", test.inputportnumberstring, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%s), expected error: %v, but got no error: %v", test.inputportnumberstring, test.expectError, err)
		}
	}
}

func TestBGPCommunityStringToValue(t *testing.T) {
	tests := []struct {
		inputString string
		expectValue int
		expectError bool
	}{
		{"101:123", 123, false},
		{"101:100", 100, false},
		{"101:9999", 9999, false},
		{"101:6", 6, false},

		{"101:999999999", 0, true},
		{"555:123", 0, true},
		{"fdjnsafjn234diedsfsa", 0, true},
		{"fdjnsafjn101:234diedsfsa", 0, true},
		{":123", 0, true},
		{"101:-123", 0, true},
	}

	for _, test := range tests {
		returned, err := BGPCommunityStringToValue(test.inputString)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%s), expected success but got error: %v", test.inputString, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%s), expected error: %v, but got no error: %v", test.inputString, test.expectError, err)
		}

		if returned != test.expectValue {
			t.Errorf("For input (%s), expected value: %d, but got: %d", test.inputString, test.expectValue, returned)
		}
	}
}

func TestBGPCommunityValueToString(t *testing.T) {
	tests := []struct {
		inputValue  int
		expect      string
		expectError bool
	}{
		{123, "101:123", false},
		{100, "101:100", false},
		{9999, "101:9999", false},
		{6, "101:6", false},

		{999999999, "", true},
		{-123, "", true},
	}

	for _, test := range tests {
		returned, err := BGPCommunityValueToString(test.inputValue)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%d), expected success but got error: %v", test.inputValue, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%d), expected error: %v, but got no error: %v", test.inputValue, test.expectError, err)
		}

		if returned != test.expect {
			t.Errorf("For input (%d), expected value: %s, but got: %s", test.inputValue, test.expect, returned)
		}
	}
}

func TestValidateBGPCommunityGroupName(t *testing.T) {
	tests := []struct {
		inputValue  string
		expectError bool
	}{
		{"incoming_group", false},
		{"my_group", false},
		{"comm_group_101", false},
		{"dashed-group", false},
		{"CamelCaseGroup", false},

		{"", true},
		{"invalid spaces", true},
		{"verylonggroupnamethatshouldnotbevalid", true},
	}

	for _, test := range tests {
		err := ValidateBGPCommunityGroupName(test.inputValue)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%s), expected success but got error: %v", test.inputValue, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%s), expected error: %v, but got no error: %v", test.inputValue, test.expectError, err)
		}
	}
}

func TestPortFullNameToPortNameAndSwitchFQDN(t *testing.T) {
	type args struct {
		switchPortCRName string
	}
	tests := []struct {
		name              string
		args              args
		wantShortPortName string
		wantSwitchFQDN    string
	}{
		{
			name: "case 1",
			args: args{
				switchPortCRName: "ethernet27-1.internal-placeholder.com",
			},
			wantShortPortName: "Ethernet27/1",
			wantSwitchFQDN:    "internal-placeholder.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShortPortName, gotSwitchFQDN := PortFullNameToPortNameAndSwitchFQDN(tt.args.switchPortCRName)
			if gotShortPortName != tt.wantShortPortName {
				t.Errorf("PortFullNameToPortNameAndSwitchFQDN() got = %v, want %v", gotShortPortName, tt.wantShortPortName)
			}
			if gotSwitchFQDN != tt.wantSwitchFQDN {
				t.Errorf("PortFullNameToPortNameAndSwitchFQDN() got1 = %v, want %v", gotSwitchFQDN, tt.wantSwitchFQDN)
			}
		})
	}
}

type RetryTestObject struct {
	shouldReturnErr error
	retryCnt        int
}

func (r *RetryTestObject) execute() error {
	fmt.Printf("performing execute()....\n")
	return r.shouldReturnErr
}
func (r *RetryTestObject) retryAction() error {
	fmt.Printf("performing retryAction....\n")
	r.shouldReturnErr = nil
	return nil
}

func shouldRetry(err error) bool {
	if strings.Contains(err.Error(), "connection lost") {
		return true
	}
	if strings.Contains(err.Error(), "connection refused") {
		return true
	}
	if strings.Contains(err.Error(), "i/o timeout") {
		return true
	}
	return false
}

func TestConnectionRetry(t *testing.T) {

	tests := []struct {
		obj              *RetryTestObject
		expectedRetryCnt int
	}{
		{
			obj: &RetryTestObject{
				shouldReturnErr: fmt.Errorf("connection lost"),
			},
			expectedRetryCnt: 1,
		},
		{
			obj: &RetryTestObject{
				shouldReturnErr: fmt.Errorf("connection refused"),
			},
			expectedRetryCnt: 1,
		},
		{
			obj: &RetryTestObject{
				shouldReturnErr: fmt.Errorf("i/o timeout"),
			},
			expectedRetryCnt: 1,
		},
		{
			obj: &RetryTestObject{
				shouldReturnErr: nil,
			},
			expectedRetryCnt: 0,
		},
		{
			obj: &RetryTestObject{
				shouldReturnErr: fmt.Errorf("some error"),
			},
			expectedRetryCnt: 0,
		},
	}

	for _, test := range tests {
		ExecuteWithRetry(
			func() error {
				return test.obj.execute()
			},
			shouldRetry,
			func() error {
				test.obj.retryCnt++
				return test.obj.retryAction()
			},
			3)

		if test.obj.retryCnt != test.expectedRetryCnt {
			t.Errorf("test failed, expected %v retries, but got %v ", test.expectedRetryCnt, test.obj.retryCnt)
		}
	}

}

func TestValidateTrunkGroups(t *testing.T) {

	var allowedTenantTrunkGroups = []string{
		"Tenant_Nets",
		"Other_Tenant_Nets",
		"Tenant_Nets_One",
		"Tenant_Nets_Two",
	}

	tests := []struct {
		inputTrunkGroups         []string
		allowedTenantTrunkGroups []string
		expectError              bool
	}{
		{[]string{}, allowedTenantTrunkGroups, false},
		{[]string{"Tenant_Nets"}, allowedTenantTrunkGroups, false},
		{[]string{"Other_Tenant_Nets"}, allowedTenantTrunkGroups, false},
		{[]string{"Tenant_Nets_One", "Tenant_Nets_Two"}, allowedTenantTrunkGroups, false},

		{[]string{"Provider_Nets"}, nil, false},
		{[]string{"Provider_Nets"}, []string{}, false}, // Should treat [] same as ""
		{[]string{"Provider_Nets", "More_Provider_Nets"}, nil, false},

		// Error cases:
		{[]string{"x"}, allowedTenantTrunkGroups, true},
		{[]string{""}, allowedTenantTrunkGroups, true},
		{[]string{"Provider_Nets"}, nil, false},
		{[]string{"Tenant_Nets", "Provider_Nets", "Other_Tenant_Nets"}, allowedTenantTrunkGroups, true},
		{[]string{"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"}, allowedTenantTrunkGroups, true},

		{[]string{"abc", "", "cde"}, nil, true},
	}

	for _, test := range tests {
		err := ValidateTrunkGroups(test.inputTrunkGroups, test.allowedTenantTrunkGroups)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%v, %v), expected success but got error: %v", test.inputTrunkGroups, test.allowedTenantTrunkGroups, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%v, %v), expected error: %v, but got no error: %v", test.inputTrunkGroups, test.allowedTenantTrunkGroups, test.expectError, err)
		}
	}
}

func TestConvertMac(t *testing.T) {
	type args struct {
		mac string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				mac: "10:70:fd:b3:5d:3e",
			},
			want: "10:70:fd:b3:5d:3e",
		},
		{
			name: "case 2",
			args: args{
				mac: "001b.21ea.66c6",
			},
			want: "00:1b:21:ea:66:c6",
		},
		{
			name: "case 3",
			args: args{
				mac: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertMac(tt.args.mac); got != tt.want {
				t.Errorf("ConvertMac() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPortChannelNumberToInterfaceName(t *testing.T) {
	tests := []struct {
		inputPortNumber int
		expectValue     string
		expectError     bool
	}{
		{5, "Port-Channel5", false},
		{25, "Port-Channel25", false},
		{999999, "Port-Channel999999", false},

		{1000000, "", true},
		{99900002, "", true},
		{0, "", true},
		{-5, "", true},
		{-10000000, "", true},
	}

	for _, test := range tests {
		returned, err := PortChannelNumberToInterfaceName(test.inputPortNumber)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%d), expected success but got error: %v", test.inputPortNumber, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%d), expected error: %v, but got no error: %v", test.inputPortNumber, test.expectError, err)
		}

		if returned != test.expectValue {
			t.Errorf("For input (%d), expected value: %s, but got: %s", test.inputPortNumber, test.expectValue, returned)
		}
	}
}

func TestPortChannelInterfaceNameToNumber(t *testing.T) {
	tests := []struct {
		inputInterfaceName string
		expectValue        int
		expectError        bool
	}{
		{"Port-Channel5", 5, false},
		{"Port-Channel25", 25, false},
		{"Port-Channel999999", 999999, false},

		{"NotAPort-Channel100", 0, true},
		{"Channel100", 0, true},
		{"Po100", 0, true},
		{"Pc100", 0, true},
		{"", 0, true},
		{"23456", 0, true},
		{"Ethernet12", 0, true},
		{"Ethernet2/3", 0, true},
	}

	for _, test := range tests {
		returned, err := PortChannelInterfaceNameToNumber(test.inputInterfaceName)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%s), expected success but got error: %v", test.inputInterfaceName, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%s), expected error: %v, but got no error: %v", test.inputInterfaceName, test.expectError, err)
		}

		if returned != test.expectValue {
			t.Errorf("For input (%s), expected value: %d, but got: %d", test.inputInterfaceName, test.expectValue, returned)
		}
	}
}

func TestPortChannelNumberAndSwitchFQDNToCRName(t *testing.T) {
	tests := []struct {
		inputChannelNumber int
		inputSwitchFQDN    string
		expectValue        string
		expectError        bool
	}{
		{5, "internal-placeholder.com", "po5.internal-placeholder.com", false},
		{25, "internal-placeholder.com", "po25.internal-placeholder.com", false},

		{0, "internal-placeholder.com", "", true},
		{1, "internal-placeholder.com", "", true},
	}

	for _, test := range tests {
		returned, err := PortChannelNumberAndSwitchFQDNToCRName(test.inputChannelNumber, test.inputSwitchFQDN)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%d, %s), expected success but got error: %v", test.inputChannelNumber, test.inputSwitchFQDN, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%d, %s), expected error: %v, but got no error: %v", test.inputChannelNumber, test.inputSwitchFQDN, test.expectError, err)
		}

		if returned != test.expectValue {
			t.Errorf("For input (%d, %s), expected value: %s, but got: %s", test.inputChannelNumber, test.inputSwitchFQDN, test.expectValue, returned)
		}
	}
}

func TestPortChannelInterfaceNameAndSwitchFQDNToCRName(t *testing.T) {
	tests := []struct {
		inputChannelName string
		inputSwitchFQDN  string
		expectValue      string
		expectError      bool
	}{
		{"Port-Channel5", "internal-placeholder.com", "po5.internal-placeholder.com", false},
		{"Port-Channel25", "internal-placeholder.com", "po25.internal-placeholder.com", false},

		{"Port-Channel0", "internal-placeholder.com", "", true},
		{"Port-Channel1", "internal-placeholder.com", "", true},
	}

	for _, test := range tests {
		returned, err := PortChannelInterfaceNameAndSwitchFQDNToCRName(test.inputChannelName, test.inputSwitchFQDN)
		if test.expectError == false && err != nil {
			t.Errorf("For input (%s, %s), expected success but got error: %v", test.inputChannelName, test.inputSwitchFQDN, err)
		}

		if test.expectError == true && err == nil {
			t.Errorf("For input (%s, %s), expected error: %v, but got no error: %v", test.inputChannelName, test.inputSwitchFQDN, test.expectError, err)
		}

		if returned != test.expectValue {
			t.Errorf("For input (%s, %s), expected value: %s, but got: %s", test.inputChannelName, test.inputSwitchFQDN, test.expectValue, returned)
		}
	}
}

func TestShouldUpdateSwitchPortVlan(t *testing.T) {
	tests := []struct {
		name         string
		switchPortCR idcnetworkv1alpha1.SwitchPort
		expectValue  bool
	}{
		{
			name: "VLAN should be updated - different vlan ids",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      10,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      5,
					PortChannel: 0,
				},
			},
			expectValue: true,
		},
		{
			name: "VLAN should not be updated - NOOPVlanID",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      idcnetworkv1alpha1.NOOPVlanID,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      5,
					PortChannel: 0,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should not be updated - when no vlan update is requested)",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      0,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      5,
					PortChannel: 0,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should not be updated - Status.VlanId is 0 and Mode is not access",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      10,
					Mode:        "trunk",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					PortChannel: 0,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should not be updated - same VlanId",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      10,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      10,
					PortChannel: 0,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should be updated - mode is access & status vlan is 0",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      20,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					PortChannel: 0,
				},
			},
			expectValue: true,
		},
		{
			name: "VLAN should be updated - removing port out from portchannel",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      30,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					PortChannel: 1,
				},
			},
			expectValue: true,
		},
		{
			name: "VLAN should not be updated -  mode is access and port is in PortChannel",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      40,
					Mode:        "access",
					PortChannel: 1,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					PortChannel: 1,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should not be updated -  mode is trunk and port is in PortChannel",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      40,
					Mode:        "trunk",
					PortChannel: 1,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					PortChannel: 1,
					Mode:        "trunk",
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should be updated - mode is access and port is not in PortChannel",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      50,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      1,
					PortChannel: 0,
				},
			},
			expectValue: true,
		},
		{
			name: "VLAN should not be updated - mode on port changed from access to trunk",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      60,
					Mode:        "trunk",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      10,
					Mode:        "access",
					PortChannel: 0,
				},
			},
			// Vlan update happens just once.
			expectValue: true,
		},
		{
			name: "VLAN should not be updated - mode on port is trunk",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      70,
					Mode:        "trunk",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					Mode:        "trunk",
					PortChannel: 0,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should be updated - mode on port changed from trunk to access",
			switchPortCR: idcnetworkv1alpha1.SwitchPort{
				Spec: idcnetworkv1alpha1.SwitchPortSpec{
					VlanId:      80,
					Mode:        "access",
					PortChannel: 0,
				},
				Status: idcnetworkv1alpha1.SwitchPortStatus{
					VlanId:      0,
					Mode:        "trunk",
					PortChannel: 0,
				},
			},
			expectValue: true,
		},
	}

	for _, test := range tests {
		returned := ShouldUpdateSwitchPortVlan(test.switchPortCR)
		if returned != test.expectValue {
			t.Errorf("For input (%v), expected value: %v, but got: %v", test.switchPortCR, test.expectValue, returned)
		}
	}
}

func TestShouldUpdatePortChannelVlan(t *testing.T) {
	tests := []struct {
		name          string
		portChannelCR idcnetworkv1alpha1.PortChannel
		expectValue   bool
	}{
		{
			name: "VLAN should be updated - different VLAN IDs",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: 10,
					Mode:   "access",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 5,
				},
			},
			expectValue: true,
		},
		{
			name: "VLAN should not be updated - same VLAN IDs",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: 10,
					Mode:   "access",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 10,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should not be updated - NOOPVlanID",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: idcnetworkv1alpha1.NOOPVlanID,
					Mode:   "access",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 5,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should not be updated - VlanID to update is 0",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: 0,
					Mode:   "access",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 5,
				},
			},
			expectValue: false,
		},
		{
			name: "VLAN should be updated - portchannel moved from access to trunk mode",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: 20,
					Mode:   "trunk",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 10,
					Mode:   "access",
				},
			},
			// Vlan update happens just once.
			expectValue: true,
		},
		{
			name: "VLAN should be updated - portchannel moved from trunk to access mode",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: 10,
					Mode:   "access",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 0,
					Mode:   "trunk",
				},
			},
			expectValue: true,
		},
		{
			name: "VLAN should not be updated - portchannel mode is trunk",
			portChannelCR: idcnetworkv1alpha1.PortChannel{
				Spec: idcnetworkv1alpha1.PortChannelSpec{
					VlanId: 30,
					Mode:   "trunk",
				},
				Status: idcnetworkv1alpha1.PortChannelStatus{
					VlanId: 0,
					Mode:   "trunk",
				},
			},
			expectValue: false,
		},
	}

	for _, test := range tests {
		returned := ShouldUpdatePortChannelVlan(test.portChannelCR)
		if returned != test.expectValue {
			t.Errorf("For input (%v), expected value: %v, but got: %v", test.portChannelCR, test.expectValue, returned)
		}
	}
}
