// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ddi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
)

const (
	InsecureSkipVerifyEnvVar = "MEN_AND_MICE_INSECURE_SKIP_VERIFY"
)

type RangesResult struct {
	Result struct {
		Ranges       []Range `json:"ranges"`
		TotalResults int     `json:"totalResults"`
	} `json:"result"`
}

type DhcpScopeResult struct {
	Result struct {
		RangeDhcpScope DhcpScope `json:"dhcpScope"`
	} `json:"result"`
}

type DhcpReservationsResult struct {
	Result struct {
		DhcpReservations []DhcpReservation `json:"dhcpReservations"`
		TotalResults     int               `json:"totalResults"`
	} `json:"result"`
}

type DhcpLeasesResult struct {
	Result struct {
		DhcpLeases   []DhcpLease `json:"dhcpLeases"`
		TotalResults int         `json:"totalResults"`
	} `json:"result"`
}

type IpAddressResult struct {
	Result struct {
		Address string `json:"address"`
	} `json:"result"`
}

type Range struct {
	Ref         string `json:"ref"`
	Name        string `json:"name"`
	From        string `json:"from"`
	To          string `json:"to"`
	ParentRef   string `json:"parentRef"`
	ChildRanges []any  `json:"childRanges"`
	DhcpScopes  []struct {
		Ref     string `json:"ref"`
		ObjType string `json:"objType"`
		Name    string `json:"name"`
	} `json:"dhcpScopes"`
	Authority struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		Sources []struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Ref     string `json:"ref"`
			Enabled bool   `json:"enabled"`
		} `json:"sources"`
	} `json:"authority"`
	Subnet           bool `json:"subnet"`
	Locked           bool `json:"locked"`
	AutoAssign       bool `json:"autoAssign"`
	HasSchedule      bool `json:"hasSchedule"`
	HasMonitor       bool `json:"hasMonitor"`
	CustomProperties struct {
		Title       string `json:"Title"`
		Description string `json:"Description"`
		Rack        string `json:"Rack"`
		Type        string `json:"Type"`
	} `json:"customProperties"`
	InheritAccess         bool   `json:"inheritAccess"`
	IsContainer           bool   `json:"isContainer"`
	UtilizationPercentage int    `json:"utilizationPercentage"`
	HasRogueAddresses     bool   `json:"hasRogueAddresses"`
	Created               string `json:"created"`
	LastModified          string `json:"lastModified"`
}

type DhcpScope struct {
	Ref           string `json:"ref"`
	Name          string `json:"name"`
	RangeRef      string `json:"rangeRef"`
	DhcpServerRef string `json:"dhcpServerRef"`
	Superscope    string `json:"superscope"`
	Description   string `json:"description"`
	Available     int    `json:"available"`
	Enabled       bool   `json:"enabled"`
}

type DhcpServer struct {
	Ref              string `json:"ref"`
	Name             string `json:"name"`
	Address          string `json:"address"`
	ResolvedAddress  string `json:"resolvedAddress"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Type             string `json:"type"`
	State            string `json:"state"`
	Security         string `json:"security"`
	CustomProperties struct {
	} `json:"customProperties"`
	Enabled bool `json:"enabled"`
	Dhcpv6  bool `json:"dhcpv6"`
}

type DhcpReservationRequest struct {
	Name              string   `json:"name"`
	ClientIdentifier  string   `json:"clientIdentifier"`
	ReservationMethod string   `json:"reservationMethod"`
	Addresses         []string `json:"addresses"`
	Type              string   `json:"type"`
}

type DhcpOptions struct {
	Option string `json:"option"`
	Value  string `json:"value"`
}

type DhcpOptionsRequest struct {
	ObjType        string        `json:"objType"`
	DhcpOptions    []DhcpOptions `json:"dhcpOptions"`
	DhcpPolicyName string        `json:"dhcpPolicyName"`
	SaveComment    string        `json:"saveComment"`
}

type DhcpReservationUpdate struct {
	Ref               string `json:"ref"`
	ObjType           string `json:"objType"`
	SaveComment       string `json:"saveComment"`
	DeleteUnspecified bool   `json:"deleteUnspecified"`
	Properties        struct {
		Name string `json:"name"`
	} `json:"properties"`
}

type DhcpReservation struct {
	Ref               string   `json:"ref"`
	Name              string   `json:"name"`
	ClientIdentifier  string   `json:"clientIdentifier"`
	ReservationMethod string   `json:"reservationMethod"`
	Addresses         []string `json:"addresses"`
	OwnerRef          string   `json:"ownerRef"`
}

type ddiResult struct {
	Ref string `json:"ref"`
}

type DhcpReservationResult struct {
	Result ddiResult `json:"result"`
}
type DhcpLease struct {
	Name         string `json:"name"`
	Mac          string `json:"mac"`
	Address      string `json:"address"`
	Lease        string `json:"lease"`
	State        string `json:"state"`
	DhcpScopeRef string `json:"dhcpScopeRef"`
}

type DDI interface {
	GetRangeByName(ctx context.Context, rack string, rangeType string) (*Range, error)
	GetDhcpReservationsByMacAddress(ctx context.Context, dhcpScopeName string, macAddress string) (*DhcpReservation, error)
	GetAvailableIp(ctx context.Context, rangeRef *Range) (string, error)
	SetDhcpReservationByScope(ctx context.Context, rangeRef *Range, macAddress string, address string, name string, filename string, nextServer string) (string, error)
	UpdateDhcpReservationOptions(ctx context.Context, rangeRef *Range, name string, filename string, nextServer string, dhcpReservationRef string, iPXEBinaryName string) error
	GetDhcpLeasesByScope(ctx context.Context, dhcpScopeName string, macAddress string) (*DhcpLease, error)
	DeleteDhcpReservation(ctx context.Context, dhcpReservationRef string) (int, error)
	DeleteDhcpLease(ctx context.Context, dhcpLeaseAddress string, dhcpScopeName string) (int, error)
	SetClient(client HTTPClient)
}

var _ DDI = (*MenAndMice)(nil)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MenAndMice struct {
	Username      string
	Password      string
	Url           string
	ServerAddress string
	client        HTTPClient
}

func NewMenAndMice(ctx context.Context, username string, password string, menAndMiceUrl string, serverAddress string) (*MenAndMice, error) {

	insecureSkipVerify, err := strconv.ParseBool(helper.GetEnv(InsecureSkipVerifyEnvVar, "false"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env InsecureSkipVerifyEnvVar")
	}

	// get vault secrets
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	return &MenAndMice{
		Username:      username,
		Password:      password,
		Url:           menAndMiceUrl,
		ServerAddress: serverAddress,
		client:        &http.Client{Timeout: 5 * time.Second, Transport: tr},
	}, nil
}

func (m *MenAndMice) SetClient(client HTTPClient) {
	m.client = client
}

func (m *MenAndMice) GetRangeByName(ctx context.Context, rack string, rangeType string) (*Range, error) {
	// first, lets find ranges that matches the type first
	// In staging, we have Racks sharing switches, which means filtering by
	// rack name will return empty
	data, _, err := m.httpRequest(ctx, "ranges?server="+m.ServerAddress+"&filter=Type="+rangeType, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get range: %v", err)
	}
	var ddi_ranges RangesResult
	err = json.Unmarshal(data, &ddi_ranges)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshall range: %v", err)
	}
	if len(ddi_ranges.Result.Ranges) < 1 {
		return nil, fmt.Errorf("failed to find ranges with type  %s", rangeType)
	}
	// find matching Rack Name
	for i, ddiRange := range ddi_ranges.Result.Ranges {
		// check if CustomProperties Rack exist
		if ddiRange.CustomProperties.Rack != "" {
			if strings.Contains(ddiRange.CustomProperties.Rack, rack) {
				return &ddi_ranges.Result.Ranges[i], nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find Requested Range with Rack %s and Type %s", rack, rangeType)
}

func (m *MenAndMice) GetDhcpReservationsByMacAddress(ctx context.Context, dhcpScopeName string, macAddress string) (*DhcpReservation, error) {
	data, _, err := m.httpRequest(ctx, dhcpScopeName+"/DHCPReservations?server="+m.ServerAddress, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get DHCPScopes: %v", err)
	}
	var dhcpReservationsResults DhcpReservationsResult
	err = json.Unmarshal(data, &dhcpReservationsResults)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshall dhcpReservationsResults: %v", err)
	}
	if len(dhcpReservationsResults.Result.DhcpReservations) < 1 {
		return nil, fmt.Errorf("failed to find DhcpReservations with %s", dhcpScopeName)
	}
	// find matching Rack Name
	for i, dhcpReservation := range dhcpReservationsResults.Result.DhcpReservations {
		// Check if Mac address in the reservation
		if dhcpReservation.ReservationMethod == "HardwareAddress" {
			if helper.NormalizeMACAddress(dhcpReservation.ClientIdentifier) == helper.NormalizeMACAddress(macAddress) {
				return &dhcpReservationsResults.Result.DhcpReservations[i], nil
			}
		}
	}
	return nil, fmt.Errorf("could not dhcp reservation in %s for %s", dhcpScopeName, macAddress)
}

func (m *MenAndMice) GetDhcpLeasesByScope(ctx context.Context, dhcpScopeName string, macAddress string) (*DhcpLease, error) {
	data, _, err := m.httpRequest(ctx, dhcpScopeName+"/Leases?server="+m.ServerAddress, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get DhcpScopes: %v", err)
	}
	var dhcpLeasesResults DhcpLeasesResult
	err = json.Unmarshal(data, &dhcpLeasesResults)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshall dhcpLeasesResults: %v", err)
	}
	if len(dhcpLeasesResults.Result.DhcpLeases) < 1 {
		return nil, fmt.Errorf("failed to find DHCPLeases with %s", dhcpScopeName)
	}
	// find matching Rack Name
	for i, dhcpLease := range dhcpLeasesResults.Result.DhcpLeases {
		// Check if Mac address in the reservation
		if helper.NormalizeMACAddress(dhcpLease.Mac) == helper.NormalizeMACAddress(macAddress) {
			return &dhcpLeasesResults.Result.DhcpLeases[i], nil
		}
	}
	return nil, fmt.Errorf("could not dhcp lease in %s for %s", dhcpScopeName, macAddress)
}

func (m *MenAndMice) DeleteDhcpLease(ctx context.Context, dhcpLeaseAddress string, dhcpScopeName string) (int, error) {
	var statusCode int
	_, statusCode, err := m.httpRequest(ctx, dhcpScopeName+"/Leases/"+dhcpLeaseAddress+"?server="+m.ServerAddress, http.MethodDelete, nil)
	if err != nil {
		return 500, fmt.Errorf("%v", err)
	}
	return statusCode, nil
}

func (m *MenAndMice) DeleteDhcpLeasesByRangeRef(ctx context.Context, rangeRef *Range, macAddress string) error {
	for i, dhcpScope := range rangeRef.DhcpScopes {
		var dhcpLease *DhcpLease
		dhcpLease, err := m.GetDhcpLeasesByScope(ctx, rangeRef.DhcpScopes[i].Ref, macAddress)
		if err != nil {
			continue
		}
		statusCode, err := m.DeleteDhcpLease(ctx, dhcpLease.Address, dhcpScope.Ref)
		if err != nil {
			continue
		}
		if statusCode >= 400 {
			continue
		}
	}
	return nil
}

func (m *MenAndMice) DeleteDhcpReservationsByRangeRef(ctx context.Context, rangeRef *Range, macAddress string) error {
	for i := range rangeRef.DhcpScopes {
		var dhcpReservation *DhcpReservation
		dhcpReservation, err := m.GetDhcpReservationsByMacAddress(ctx, rangeRef.DhcpScopes[i].Ref, macAddress)
		if err != nil {
			continue
		}
		statusCode, err := m.DeleteDhcpReservation(ctx, dhcpReservation.Ref)
		if err != nil {
			continue
		}
		if statusCode >= 400 {
			continue
		}

	}
	return nil
}

func (m *MenAndMice) DeleteDhcpReservation(ctx context.Context, dhcpReservationRef string) (int, error) {
	_, statusCode, err := m.httpRequest(ctx, dhcpReservationRef+"?server="+m.ServerAddress, http.MethodDelete, nil)
	if err != nil {
		return 500, fmt.Errorf(" %v", err)
	}
	return statusCode, nil
}

func (m *MenAndMice) SetDhcpReservationByScope(ctx context.Context, rangeRef *Range, macAddress string,
	address string, name string, filename string, nextServer string) (string, error) {
	dhcpScope := rangeRef.DhcpScopes[0]
	var dhcpReservation DhcpReservationRequest
	dhcpReservation.ReservationMethod = "HardwareAddress"
	dhcpReservation.Addresses = append(dhcpReservation.Addresses, address)
	dhcpReservation.ClientIdentifier = macAddress
	dhcpReservation.Name = name
	dhcpReservation.Type = "DHCP"
	var body struct {
		DhcpReservation DhcpReservationRequest `json:"dhcpReservation"`
	}
	body.DhcpReservation = dhcpReservation
	httpBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf(" %v", err)
	}
	var resultRef DhcpReservationResult
	httpResult, statusCode, err := m.httpRequest(ctx, dhcpScope.Ref+"/DHCPReservations?server="+m.ServerAddress, http.MethodPost, httpBody)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	err = json.Unmarshal(httpResult, &resultRef)
	if err != nil {
		return "", fmt.Errorf("failed unmarshall ipAddressResult: %v", err)
	}
	if statusCode >= 400 {
		return "", fmt.Errorf("failed to set dhcpReservation %s %d %s", dhcpReservation.Name, statusCode, string(httpResult))
	}
	return resultRef.Result.Ref, nil
}

func (m *MenAndMice) UpdateDhcpReservationOptions(ctx context.Context, rangeRef *Range, name string, filename string, nextServer string, dhcpReservationRef string, iPXEBinaryName string) error {

	var body DhcpOptionsRequest
	// file server option
	var fileServer DhcpOptions
	fileServer.Option = "::66"
	fileServer.Value = nextServer
	// standard filename option
	var fileNameStd DhcpOptions
	fileNameStd.Option = "::67"
	fileNameStd.Value = iPXEBinaryName
	// iPXE  filename option
	var fileNameIpxe DhcpOptions
	fileNameIpxe.Option = ":iPXE:67"
	fileNameIpxe.Value = filename

	body.ObjType = "Unknown"
	body.DhcpOptions = append(body.DhcpOptions, fileServer, fileNameStd, fileNameIpxe)

	httpBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	data, statusCode, err := m.httpRequest(ctx, dhcpReservationRef+"/Options?server="+m.ServerAddress, http.MethodPut, httpBody)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if statusCode >= 400 {
		return fmt.Errorf("failed to update dhcpReservation %s %d %s", dhcpReservationRef, statusCode, string(data))
	}
	return nil
}

func (m *MenAndMice) GetAvailableIp(ctx context.Context, rangeRef *Range) (string, error) {
	// TODO remove this code; DDi should mark first 3 addresses allocated
	// Start address skip first 3 IPs
	startIpFrom := rangeRef.From
	s := strings.Split(startIpFrom, ".")
	s[len(s)-1] = "4"
	startIpFrom = strings.Join(s[:], ".")
	// Call Get next ipaddress and hold it for 15 seconds
	data, _, err := m.httpRequest(ctx, rangeRef.Ref+"/NextFreeAddress?server="+m.ServerAddress+"&excludeDHCP=true&temporaryClaimTime=15&startAddress="+startIpFrom, http.MethodGet, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get NextFreeAddress: %v", err)
	}
	var ipAddressResult IpAddressResult
	err = json.Unmarshal(data, &ipAddressResult)
	if err != nil {
		return "", fmt.Errorf("failed unmarshall ipAddressResult: %v %s", err, string(data))
	}
	return ipAddressResult.Result.Address, nil
}

func (m *MenAndMice) httpRequest(ctx context.Context, api string, method string, body []byte) ([]byte, int, error) {

	req, err := http.NewRequest(method, fmt.Sprintf("%s/mmws/api/%s", m.Url, api), bytes.NewBuffer(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	if err != nil {
		return nil, 500, fmt.Errorf("failed to create http request: %v", err)
	}

	req.SetBasicAuth(m.Username, m.Password)

	res, err := m.client.Do(req)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to send http request: %v", err)
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to read http response: %v", err)
	}
	return resBody, res.StatusCode, nil
}
