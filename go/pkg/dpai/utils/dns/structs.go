package dns

// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

// result response
type DNSZoneResponse struct {
	Result DNSRecordZonesList `json:"result"`
}

// the zone information list
type DNSRecordZonesList struct {
	DnsZones []DNSZone `json:"dnsZones"`
}

// individual Zone Information for the dns m&mice zones
type DNSZone struct {
	Ref              string                 `json:"ref"`
	Name             string                 `json:"name"`
	Dynamic          bool                   `json:"dynamic"`
	AdIntegrated     bool                   `json:"adIntegrated"`
	DNSViewRef       string                 `json:"dnsViewRef"`
	SourceZoneRef    string                 `json:"sourceZoneRef"`
	Authority        string                 `json:"authority"`
	Type             string                 `json:"type"`
	DNSSecSigned     bool                   `json:"dnssecSigned"`
	KSKIDs           string                 `json:"kskIDs"`
	ZSKIDs           string                 `json:"zskIDs"`
	CustomProperties map[string]interface{} `json:"customProperties,omitempty"`
	Created          string                 `json:"created"`
	LastModified     string                 `json:"lastModified"`
	DisplayName      string                 `json:"displayName"`
}

// dns server utils struct information

type DNSServerResponse struct {
	Result DNSServerList `json:"result"`
}

// the dnsServer Information list
type DNSServerList struct {
	DnsServers []DnsServers `json:"dnsServers"`
}

type DnsServers struct {
	Ref              string      `json:"ref"`
	Name             string      `json:"name"`
	Address          string      `json:"address"`
	ResolvedAddress  string      `json:"resolvedAddress"`
	Port             int         `json:"port"`
	Type             string      `json:"type"`
	State            string      `json:"state"`
	CustomProperties interface{} `json:"customProperties"`
	Subtype          string      `json:"subtype"`
	Enabled          bool        `json:"enabled"`
}

// dns Records Response Information
type DnsRecordsResponse struct {
	Result DnsRecordsList `json:"result"`
}

// list of all the DNS Records
type DnsRecordsList struct {
	DNSRecords []DNSRecord `json:"dnsRecords"`
}

// dns Record Information taken from a zone
type DNSRecord struct {
	Ref         string            `json:"ref,omitempty"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	TTL         string            `json:"ttl"`
	Data        string            `json:"data"`
	Comment     string            `json:"comment"`
	Enabled     bool              `json:"enabled"`
	Aging       int               `json:"aging,omitempty"`
	DNSZoneRef  string            `json:"dnsZoneRef"`
	CustomProps map[string]string `json:"customProperties,omitempty"`
}

// create dns Record in the zone
type DnsRecordRequest struct {
	DNSRecord                          DNSRecord `json:"dnsRecord"`
	SaveComment                        string    `json:"saveComment"`
	PlaceAfterRef                      string    `json:"placeAfterRef,omitempty"`
	AutoAssignRangeRef                 string    `json:"autoAssignRangeRef,omitempty"`
	ForceOverrideOfNamingConflictCheck bool      `json:"forceOverrideOfNamingConflictCheck"`
}

// dns record add in a zone sucessresponse
type DnsRecordRequestSuccessResponse struct {
	Result struct {
		Ref string `json:"ref"`
	} `json:"result"`
}
