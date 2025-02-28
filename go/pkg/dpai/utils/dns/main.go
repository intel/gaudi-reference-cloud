package dns

// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
)

const (
	DPAI_ZONE      = "dpai.cloudworkspace.io."
	REALM_INTERNAL = "Internal"
	REALM_EXTERNAL = "External"
)

// DNS mount path for vault sidecar secrets
const (
	DNS_IPAM_USERNAME = "/vault/secrets/dns_username"
	DNS_IPAM_PASSWORD = "/vault/secrets/dns_password"
)

const (
	TYPE_CNAME = "CNAME"
	TYPE_A     = "A"
	TYPE_AAAA  = "AAAA"
	TYPE_SOA   = "SOA"
	TYPE_TXT   = "TXT"
	TYPE_MX    = "MX"
)

// ipam menmice utils for DNS config
// TODO: Read from config map
const (
	DNS_MEN_MICE_IPAM_URL = "https://ipami.idcstage.intel.com/mmws/api/v2" // load from config map from DPAI helm
)

const (
	DNS_DOMAIN_TTL = 1080
)

// IMenMice Service Provider
type IMenMice interface {
	// auth dns zone
	SetApiToken()
	GetBaseUrl() string

	// get dns zone api's
	FetchDNSZones() (*DNSZoneResponse, error)
	FetchDNSZoneByName(string) (map[string][]DNSZone, error)
	FetchAllDnsRecords() error
	FetchDNSServers() (*DNSServerResponse, error)

	FetchDnsRecordsInZone(map[string][]DNSZone) (*DnsRecordsResponse, error)
	FetchDnsRecordsInZoneByDnsRef(map[string][]DNSZone, string) (*DNSRecord, error)
	FetchDnsRecordsInZoneByDataRef(map[string][]DNSZone, string) (*DNSRecord, error)

	// create record in zone
	CreateDnsRecordInZone(map[string][]DNSZone, string, string, string) (*DnsRecordRequestSuccessResponse, error)

	DeleteDnsRecord(string) error
	DeleteDnsRecordsInZoneByDataRef(zoneInfo map[string][]DNSZone, dataRef string) error
}

type DNSIpanLoginStruct struct {
	username string
	password string
}

type MenMiceService struct {
	ctx                *context.Context
	ctxCancel          context.CancelFunc
	BaseURL            string
	HttpClient         *http.Client
	APIBasicAuthHeader string
	Logger             *log.Logger
}

func generateContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// generics can be used here in future for go 1.19+
func getDNSToken() *DNSIpanLoginStruct {
	log.Println("Fetching the Login Basic Auth Header")

	// username
	_, err := os.Stat(DNS_IPAM_USERNAME)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic(fmt.Errorf("File %s does not exist", DNS_IPAM_USERNAME))
		}
		panic(err.Error())
	}

	dns_username, err := os.ReadFile(DNS_IPAM_USERNAME)
	if err != nil {
		panic(fmt.Errorf("Error reading the file %s", DNS_IPAM_USERNAME))
	}

	_, err = os.Stat(DNS_IPAM_PASSWORD)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic(fmt.Errorf("File %s does not exist", DNS_IPAM_PASSWORD))
		}
		panic(err.Error())
	}

	dns_password, err := os.ReadFile(DNS_IPAM_PASSWORD)
	if err != nil {
		panic(fmt.Errorf("Error reading the file %s", DNS_IPAM_PASSWORD))
	}

	dns_username_string := strings.Trim(string(dns_username), "\n")
	dns_password_string := strings.Trim(string(dns_password), "\n")

	return &DNSIpanLoginStruct{
		username: dns_username_string,
		password: dns_password_string,
	}
}

// reads the config mounted via dpai helm configmap
func getDnsConfig() *config.DnsConfig {
	config, err := config.ReadConfig()
	if err != nil {
		log.Printf("The DNS util service cannot start unless valid DNS config are provided %+v", err)
		panic(err.Error())
	}
	return &config.Dns
}

// Service root factory
// NewMenMiceService creates a new instance (a factory method) of the MenMiceService struct with the provided baseUrl.
func NewMenMiceService(baseUrl string) IMenMice {
	requestCancelContext, canceFunc := context.WithCancel(context.Background())
	return &MenMiceService{
		BaseURL: baseUrl,
		HttpClient: &http.Client{
			Timeout: time.Second * 45,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
		},
		ctx:       &requestCancelContext,
		ctxCancel: canceFunc,
	}
}

func getCancelContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func createHttpRequest(method string, url string, addonOptions *bytes.Reader,
	authHeader string) http.Request {
	var req *http.Request
	var err error
	if addonOptions == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, addonOptions)
	}
	if err != nil {
		fmt.Printf("Error in forming a base Http Request Header %+v", err)
		panic(err.Error())
	}

	// this is done with the base auth header by internall using base64 for encode / decode
	req.Header.Add("Authorization", "Basic "+authHeader)
	return *req
}

func (m *MenMiceService) GetBaseUrl() string { return m.BaseURL }

// init the base struct with the Api token to auth with the DNS men and mice Api's
func (m *MenMiceService) SetApiToken() {
	var loginBody DNSIpanLoginStruct = *getDNSToken()
	header := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", loginBody.username, loginBody.password)))
	m.APIBasicAuthHeader = header
}

// FetchDNSZones fetches all the DNS zones from the Men & Mice API and returns a DNSZoneResponse struct containing the zones.
func (m *MenMiceService) FetchDNSZones() (*DNSZoneResponse, error) {

	log.Println("Fetching All the DNS Zones")
	req := createHttpRequest("GET", fmt.Sprintf("%s/DNSZones", m.BaseURL), nil, m.APIBasicAuthHeader)

	resp, err := m.HttpClient.Do(&req)

	if err != nil {
		fmt.Println("Error in fetching dns Zones ")
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error fetching the zones %+v", err)
		m.ctxCancel()
	}

	var dnsZones DNSZoneResponse

	if err := json.Unmarshal(data, &dnsZones); err != nil {
		log.Println("Error in marshall the dns Zone respone")
		return nil, err
	}

	return &dnsZones, nil
}

// FetchDNSZoneByName fetches the DNS zones for the given zone name and returns a map of zone type to a slice of DNSZone pointers.
func (m *MenMiceService) FetchDNSZoneByName(zoneName string) (map[string][]DNSZone, error) {
	dns_zones, err := m.FetchDNSZones()
	zone := *dns_zones

	log.Printf("Fetching the DnzZone By Name %s", zoneName)
	if err != nil {
		return nil, err
	}

	var dpai_zoneMap map[string][]DNSZone = make(map[string][]DNSZone)
	for _, zone := range zone.Result.DnsZones {
		if zone.Name == zoneName {
			if zone.Type == "Primary" {
				log.Println("Fetched information for the Primary Zones type ", zone.Type, zone.CustomProperties, zone.Name, zone.Ref)
				dpai_zoneMap[zone.Type] = append(dpai_zoneMap[zone.Type], zone)
			}
		}
	}

	if len(dpai_zoneMap) == 0 {
		return nil, fmt.Errorf("No Matching Dns Zone found for the request")
	}

	log.Println("zone info ", dpai_zoneMap)

	return dpai_zoneMap, nil
}

func (m *MenMiceService) FetchAllDnsRecords() error {
	return nil
}

// getRealms takes a slice of DNSZone pointers and returns the internal and external realms from the zone information.
func getRealms(primaryZone []DNSZone) (*DNSZone, *DNSZone, error) {
	log.Println("Fetching the Realms from DnsZones Map")

	var external_realm *DNSZone
	var internal_realm *DNSZone

	for _, realm := range primaryZone {
		realmProps, fd := realm.CustomProperties["Realm"]
		if fd {
			if realmProps == REALM_EXTERNAL {
				external_realm = &realm
			} else if realmProps == REALM_INTERNAL {
				internal_realm = &realm
			}
		} else {

		}
	}
	if external_realm == nil || internal_realm == nil {
		return nil, nil, fmt.Errorf("Error while fetching info for realm since configured as a primary server in zone")
	}
	return internal_realm, external_realm, nil
}

// FetchDnsRecordsInZone fetches all the DNS records in a specific zone.
func (m *MenMiceService) FetchDnsRecordsInZone(zoneInfo map[string][]DNSZone) (*DnsRecordsResponse, error) {
	// assume we will use primary for use relying on dns server for sync replication

	primaryZone := zoneInfo["Primary"]
	log.Println("the Zone information is ", primaryZone)
	internal_realm, _, err := getRealms(primaryZone)
	if err != nil {
		fmt.Println("error is ", err)
		return nil, err
	}
	refList := strings.Split(internal_realm.Ref, "/")
	ref := refList[len(refList)-1]
	req := createHttpRequest("GET", fmt.Sprintf("%s/dnsZones/%s/dnsRecords", m.BaseURL, ref), nil, m.APIBasicAuthHeader)

	resp, err := m.HttpClient.Do(&req)
	if err != nil {
		fmt.Println("Error in fetching the DNS Records for the zone with name", internal_realm.Name)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		log.Printf("Error in fetching the DNS Records for the zone with name %s for request : %s", internal_realm.Name, string(body))
		_, reasons := processStatusCode(resp.StatusCode, body)
		return nil, fmt.Errorf("%s", reasons)
	}

	data, err := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	var dnsRecords DnsRecordsResponse
	if err := json.Unmarshal(data, &dnsRecords); err != nil {
		log.Println("Error Unmarshalling thr dns Payload Information")
		return nil, err
	}

	return &dnsRecords, nil
}

// FetchDnsRecordsInZoneByDnsRef fetches the DNS record with the given name from the provided zone information.
func (m *MenMiceService) FetchDnsRecordsInZoneByDnsRef(zoneInfo map[string][]DNSZone, dnsRecordName string) (*DNSRecord, error) {

	dnsRecords, err := m.FetchDnsRecordsInZone(zoneInfo)
	if err != nil {
		fmt.Println("Error in fetching all the required dns records")
		return nil, err
	}

	var dnsRecord DNSRecord
	for _, record := range dnsRecords.Result.DNSRecords {
		if record.Name == dnsRecordName {
			dnsRecord = record
			break
		}
	}

	return &dnsRecord, nil
}

// FetchDnsRecordsInZoneByDataRef fetches all the DNS records in the provided zone information that have a matching Data field and a CNAME record type.
func (m *MenMiceService) FetchDnsRecordsInZoneByDataRef(zoneInfo map[string][]DNSZone, dataRef string) (*DNSRecord, error) {

	dnsRecords, err := m.FetchDnsRecordsInZone(zoneInfo)

	if err != nil {
		log.Println("Error in fetching all the required dns records")
		return nil, err
	}

	// appends all the zone ref matching the current data ref
	// for now only use for dpai case to fetch and delete all the CNAME based records
	for _, record := range dnsRecords.Result.DNSRecords {
		if strings.TrimSpace(record.Name) == dataRef && record.Type == "CNAME" {
			return &record, nil
		}
	}

	// the length of the dnsRecord slice will be always 1 since only one unique subdomain per name is allowed
	return nil, nil
}

// CreateDnsRecordInZone creates a new DNS record in the specified zone with the given record name and target zone domain.
func (m *MenMiceService) CreateDnsRecordInZone(zoneInfo map[string][]DNSZone, recordName string,
	targetZoneDomain string, realmInfo string) (*DnsRecordRequestSuccessResponse, error) {
	log.Printf("Inside Create dns Record in Zone for %s", recordName)

	primaryZone := zoneInfo["Primary"]
	internal_realm, external_realm, err := getRealms(primaryZone)

	if err != nil {
		return nil, fmt.Errorf("Error in fetching the realm information for the zone")
	}

	var realm *DNSZone = nil
	if realmInfo == REALM_EXTERNAL {
		realm = external_realm
	} else {
		realm = internal_realm
	}

	refList := strings.Split(realm.Ref, "/")
	ref := refList[len(refList)-1]
	var dnsRecordZoneRequest DnsRecordRequest = DnsRecordRequest{
		DNSRecord: DNSRecord{
			Name:       recordName,
			Type:       TYPE_CNAME,
			TTL:        strconv.Itoa(DNS_DOMAIN_TTL),
			Comment:    "Created DNS Records By DPAI for DPAI  Service for Airflow with record Name " + recordName,
			Enabled:    true,
			Data:       targetZoneDomain,
			DNSZoneRef: realm.Ref,
		},
		SaveComment:                        "Created by DPAI Services for Airflow with record Name " + recordName,
		ForceOverrideOfNamingConflictCheck: true,
	}

	payload, err := json.Marshal(dnsRecordZoneRequest)
	if err != nil {
		log.Printf("Error in Marshalling the Dns Record Request for the zone with name %s", realm.Name)
		return nil, err
	}

	reader := bytes.NewReader(payload)
	req := createHttpRequest("POST", fmt.Sprintf("%s/dnsZones/%s/dnsRecords", m.BaseURL, ref), reader,
		m.APIBasicAuthHeader)

	resp, err := m.HttpClient.Do(&req)
	if err != nil {
		log.Println("Error in fetching the DNS Records for the zone with name", realm.Name)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error in Reading the Response Body")
		}
		log.Println("err in creating a Dns Record in zone with name ", targetZoneDomain, recordName, string(body))
		msg, reasons := processStatusCode(resp.StatusCode, body)
		log.Println(msg, reasons)
		return nil, fmt.Errorf("%s", reasons)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading the response body %+v", err)
	}
	defer resp.Body.Close()

	log.Println("Dns Record susccessfull Created with Response", string(data))

	var dnsRecordCreattionSuccessResponse DnsRecordRequestSuccessResponse
	err = json.Unmarshal(data, &dnsRecordCreattionSuccessResponse)
	if err != nil {
		log.Println("Error in parsing the response ")
		return nil, err
	}

	return &dnsRecordCreattionSuccessResponse, nil
}

// Get All the DNS Servers to return list of DNS servers from the MenMice DDI.
func (m *MenMiceService) FetchDNSServers() (*DNSServerResponse, error) {
	req := createHttpRequest("GET", fmt.Sprintf("%s/dnsServers", m.BaseURL), nil, m.APIBasicAuthHeader)

	resp, err := m.HttpClient.Do(&req)

	payload, err := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		log.Println("Error fetching the dns Servers Panic")
		return nil, err
	}

	var dnsServers DNSServerResponse

	if err := json.Unmarshal(payload, &dnsServers); err != nil {
		log.Println("Error Marshalling the dns Servers Panic")
		return nil, err
	}

	return &dnsServers, nil
}

// DeleteDnsRecord deletes a DNS record with the given dnsRecordRef using /dnsRecords/{dnsRecordRef} endpoint,.
func (m *MenMiceService) DeleteDnsRecord(dnsRecordRef string) error {
	log.Println("Deleting the dnsRecordRef ", dnsRecordRef)
	req := createHttpRequest("DELETE", fmt.Sprintf("%s/dnsRecords/%s", m.BaseURL, dnsRecordRef), nil, m.APIBasicAuthHeader)

	log.Println("url invoked to Delete DNS Data ref", fmt.Sprintf("%s/dnsRecords/%s", m.BaseURL, dnsRecordRef), dnsRecordRef)
	resp, err := m.HttpClient.Do(&req)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		log.Println("err in deleting a Dns Record in zone with name ", dnsRecordRef, string(body))
		_, reasons := processStatusCode(resp.StatusCode, body)
		return fmt.Errorf("%s", reasons)
	}

	fmt.Printf("The Required dnsRecordRef successfully deleted %s and unlinked from all DnsZones", dnsRecordRef)

	return nil
}

/*
Delete DNS Record use data Ref
Uses the target and cascadingly deletes all the DNS Records which matches the provided target data ref
*/
func (m *MenMiceService) DeleteDnsRecordsInZoneByDataRef(zoneInfo map[string][]DNSZone, dataRef string) error {

	dnsRecordsbyDataRef, err := m.FetchDnsRecordsInZoneByDataRef(zoneInfo, dataRef)

	if err != nil {
		return err
	}

	if dnsRecordsbyDataRef == nil {
		log.Println("No matching dns records found for the dataRef ", dataRef)
		return nil
	}

	log.Println("record fetched to Delete ", dnsRecordsbyDataRef.Name, dnsRecordsbyDataRef.Ref)
	err = m.DeleteDnsRecord(dnsRecordsbyDataRef.Ref)
	if err != nil {
		fmt.Println("Error in deleting the name and record ref ", dnsRecordsbyDataRef.Name, dnsRecordsbyDataRef.Ref)
	}

	log.Println("Successfully deleted all the dns Record Ref for the following dataRef", dnsRecordsbyDataRef.Name, dataRef)
	return nil
}

func processStatusCode(c int, b []byte) (string, string) {

	// provider specific error payload
	type message struct {
		message string
	}

	//this is how errors will be returned by the provider
	type issues struct {
		Messages []message `json:"messages"`
	}

	//Set of messages associated with the status codes in case of issues from this provider
	const (
		BAD_REQUEST  string = "Bad Request: Invalid parameters or incorrect values for env type"
		UNAUTHORIZED string = "Unauthorized: Missing, Expired, or Invalid apiToken"
		FORBIDDEN    string = "Forbidden: User/User Group does not have access to given object"
		NOT_FOUND    string = "Not Found: Object or supporting object not found"
		CONFLICT     string = "Conflict: Object already exists with name/ip/... or object is in use by another object"
		ERROR        string = "Error Occured"
	)

	//Get the error messages and return with status codes
	var ret issues
	var m, reasons string
	err := json.Unmarshal(b, &ret)
	if err != nil {
		return ERROR, ""
	}
	//get all the issues appended. We have seen one message only.
	for _, msg := range ret.Messages {
		reasons += msg.message
	}

	//See what code we have received and assign the message
	switch c {
	case 400:
		m = BAD_REQUEST
	case 401:
		m = UNAUTHORIZED
	case 403:
		m = FORBIDDEN
	case 404:
		m = NOT_FOUND
	case 409:
		m = CONFLICT
	default:
		m = ERROR
	}

	return m, reasons
}

// used to generate domain fqdn from highwire ip and associated env highwire is linked to in IDC
func GenerateFQDNfromIP(ip string, env string) string {
	ip = strings.Join(strings.Split(ip, "."), "-")
	return fmt.Sprintf("lbauto-%s.%s.intel.com.", ip, env)
}

// used to load DPAI IPAM IDC endpoint from configuration
func ConfigureIpamApiEndpoint() string {
	val, fd := os.LookupEnv("DPAI_API_ENV")
	if !fd {
		panic(fmt.Errorf("No Required DNS IPAM Endpoint configured"))
	}
	return val
}
