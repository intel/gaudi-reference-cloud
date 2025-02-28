package netboxclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-logr/logr"
	netboxv4 "github.com/jerryzhen01/go-netbox/v4"
	netboxv3 "github.com/netbox-community/go-netbox/v3"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
)

var (
	logger logr.Logger
)

func init() {
	logger = log.FromContext(context.Background()).WithName("NetBoxClient")
}

const (
	authHeaderName          = "Authorization"
	authHeaderFormat        = "Token %v"
	errorBodyTruncateLength = 500

	netboxVersion3 = "3"
	netboxVersion4 = "4"
)

type SDNNetboxClient interface {
	Version() string
	ListDevices(context.Context, ListDevicesRequest) ([]Device, error)
	ListInterfaces(context.Context, ListInterfaceRequest) ([]Interface, error)
	ListIPAddresses(context.Context, ListIPAddressesRequest) ([]IPAddress, error)
}

// Device represents a simplified structure for a device in Netbox
type Device struct {
	Id                   int32                  `json:"id"`
	Url                  string                 `json:"url"`
	Display              string                 `json:"display"`
	Name                 string                 `json:"name,omitempty"`
	DeviceType           string                 `json:"device_type"`
	Role                 string                 `json:"role"`
	DeviceRole           string                 `json:"device_role"`
	Site                 string                 `json:"site"`
	Status               string                 `json:"status,omitempty"`
	Description          string                 `json:"description,omitempty"`
	ConfigContext        interface{}            `json:"config_context"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

func (d *Device) GetId() int32 {
	return d.Id
}

func (d *Device) GetName() string {
	return d.Name
}

// Interface represents a simplified structure for a network interface in Netbox
type Interface struct {
	Id                   int32      `json:"id"`
	Url                  string     `json:"url"`
	Display              string     `json:"display"`
	Device               Device     `json:"device"`
	Name                 string     `json:"name"`
	Label                *string    `json:"label,omitempty"`
	Enabled              *bool      `json:"enabled,omitempty"`
	Description          *string    `json:"description,omitempty"`
	LinkPeers            []LinkPeer `json:"link_peers"`
	AdditionalProperties map[string]interface{}
}

func (d *Interface) GetId() int32 {
	return d.Id
}

func (d *Interface) GetName() string {
	return d.Name
}

type LinkPeer struct {
	ID      int32           `json:"id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Display string          `json:"display,omitempty"`
	Device  *LinkPeerDevice `json:"device,omitempty"`
	// Cable is an integer in v3 but interface in v4, which is causing parsing issue.
	// Since we are NOT using this field, commenting it out is the simplest work around.
	// If we need field like this in the future, we can create customized json unmarshal function for each of them.
	// Cable   int32           `json:"cable,omitempty"`
}

type LinkPeerDevice struct {
	ID      int32  `json:"id,omitempty"`
	Display string `json:"display,omitempty"`
	Name    string `json:"name,omitempty"`
}

// IPAddress represents a simplified structure for an IP address in Netbox
type IPAddress struct {
	Id                   int32                  `json:"id"`
	Url                  string                 `json:"url"`
	Display              string                 `json:"display"`
	Family               int32                  `json:"family"`
	Address              string                 `json:"address"`
	DnsName              *string                `json:"dns_name,omitempty"`
	Description          *string                `json:"description,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	AdditionalProperties map[string]interface{}
}

func (d *IPAddress) GetId() int32 {
	return d.Id
}

type GetVersionResponse struct {
	Version string `json:"netbox-version"`
}

func NewSDNNetBoxClient(netboxServerURL string, token string, insecureSkipVerify bool) (SDNNetboxClient, error) {
	token = strings.TrimSpace(token)
	transportScheme, host, err := utils.ExtractTransportProtocol(netboxServerURL)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}

	httpClient := &http.Client{
		Transport: httpTransport,
		Timeout:   30 * time.Second,
	}

	// get the Netbox version
	versionRequestURL := fmt.Sprintf("%s/api/status/", netboxServerURL)
	// Create a new HTTP request with headers
	req, err := http.NewRequest("GET", versionRequestURL, nil)
	if err != nil {
		return nil, err
	}

	// Set the necessary headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get version from %v, reason: %v", versionRequestURL, err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get a successful response from netbox /status. response_status: %s, body: %s", resp.Status, body)
	}

	// Parse the JSON response to extract the version
	var netboxFullVersionResp GetVersionResponse
	if err := json.Unmarshal(body, &netboxFullVersionResp); err != nil {
		return nil, err
	}
	majorVersion := extractMajorVersion(netboxFullVersionResp)

	if majorVersion == netboxVersion3 {
		// create v3 netbox client
		cfgv3 := netboxv3.NewConfiguration()
		cfgv3.HTTPClient = httpClient
		cfgv3.Scheme = transportScheme
		cfgv3.Host = host
		cfgv3.DefaultHeader = map[string]string{
			"Authorization": "Token " + token,
		}
		cv3 := netboxv3.NewAPIClient(cfgv3)
		return &NetBoxClientV3{
			client: cv3,
		}, nil

	} else if majorVersion == netboxVersion4 {
		// create v4 netbox client
		cfgv4 := netboxv4.NewConfiguration()
		cfgv4.HTTPClient = httpClient
		cfgv4.Scheme = transportScheme
		cfgv4.Host = host
		cfgv4.DefaultHeader = map[string]string{
			"Authorization": "Token " + token,
		}
		cv4 := netboxv4.NewAPIClient(cfgv4)
		return &NetBoxClientV4{
			client: cv4,
		}, nil
	} else {
		return nil, fmt.Errorf("unrecognized Netbox version: %s. Full response from /status: %s", majorVersion, body)
	}
}

func extractMajorVersion(resp GetVersionResponse) string {
	// Return 0 if the version string is empty
	if resp.Version == "" {
		return "0"
	}

	// Split the version string by "."
	parts := strings.Split(resp.Version, ".")

	// Check if the first part exists
	if len(parts) > 0 {
		majorVersion := parts[0]
		return majorVersion
	}

	// Return 0 if extraction fails
	return "0"
}

type DevicesFilter struct {
	/* Locations Fields */
	RegionID    []int32  `json:"regionID,omitempty"`
	Site        []string `json:"site,omitempty"`
	SiteID      []int32  `json:"siteID,omitempty"`
	SiteGroupID []int32  `json:"siteGroupID,omitempty"`
	LocationID  []int32  `json:"locationID,omitempty"`
	RackID      []int32  `json:"rackID,omitempty"`

	/* Device Fields */
	DeviceName   []string `json:"deviceName,omitempty"`
	DeviceID     []int32  `json:"deviceID,omitempty"`
	DeviceType   []string `json:"deviceType,omitempty"`
	DeviceTypeID []int32  `json:"deviceTypeID,omitempty"`

	/* Management Fields */
	Status []string `json:"status,omitempty"`
	Role   []string `json:"role,omitempty"`
	RoleID []int32  `json:"roleID,omitempty"`

	/* tags */
	Tag []string `json:"tag,omitempty"`
}

type ListDevicesRequest struct {
	Filter *DevicesFilter
}

type InterfacesFilter struct {
	InterfaceName  []string          `json:"interfaceName,omitempty"`
	InterfaceNameN []string          `json:"interfaceNameN,omitempty"`
	Device         []*string         `json:"device,omitempty"`
	DeviceID       []int32           `json:"deviceID,omitempty"`
	DeviceRole     []string          `json:"deviceRole,omitempty"`
	Cabled         *bool             `json:"cabled,omitempty"`
	Connected      *bool             `json:"connected,omitempty"`
	Label          []string          `json:"string,omitempty"`
	Tag            []string          `json:"tags,omitempty"`
	CustomFields   map[string]string `json:"customFields,omitempty"`

	// extra filter fields
	InterfaceNameRegex string `json:"interfaceNameRegex,omitempty"`
}

type ListInterfaceRequest struct {
	Filter *InterfacesFilter
}

type IPAddressesFilter struct {
	/* Device Fields */
	DeviceName []string `json:"deviceName,omitempty"`
	DeviceID   []int32  `json:"deviceID,omitempty"`

	Interfaces   []string  `json:"interfaces,omitempty"`
	InterfacesId []int32   `json:"interfacesId,omitempty"`
	VRF          []*string `json:"vrf,omitempty"`

	/* Management Fields */
	Status []string `json:"status,omitempty"`
	Role   []string `json:"role,omitempty"`
	RoleID []int32  `json:"roleID,omitempty"`

	/* tags */
	Tag []string `json:"tag,omitempty"`
}

type ListIPAddressesRequest struct {
	Filter *IPAddressesFilter
}

type ClusterFilter struct {
	RegionID       []int32  `json:"regionID,omitempty"`
	Site           []string `json:"site,omitempty"`
	SiteID         []*int32 `json:"siteID,omitempty"`
	ClusterName    []string `json:"clusterName,omitempty"`
	ClusterID      []int32  `json:"clusterID,omitempty"`
	ClusterGroup   []string `json:"clusterGroup,omitempty"`
	ClusterGroupID []*int32 `json:"clusterGroupID,omitempty"`
	Type           []string `json:"type,omitempty"`
	TypeID         []int32  `json:"typeID,omitempty"`
	Tenant         []string `json:"tenant,omitempty"`
	Status         []string `json:"status,omitempty"`
	Tag            []string `json:"tag,omitempty"`
}

type ListClusterRequest struct {
	Filter *ClusterFilter
}

func readStartOfBody(httpResp *http.Response) string {
	if httpResp == nil || httpResp.Body == nil {
		return ""
	}
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		// If we can't read the body, return an empty string
		return ""
	}
	bodyStr := string(body)
	if len(bodyStr) > errorBodyTruncateLength {
		bodyStr = bodyStr[0:errorBodyTruncateLength] + "..."
	}
	return bodyStr
}

func readResponseBody(httpResp *http.Response) string {
	if httpResp == nil || httpResp.Body == nil {
		return ""
	}
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		// If we can't read the body, return an empty string
		return ""
	}
	return string(body)
}

func readRequestURL(httpResp *http.Response) *url.URL {
	if httpResp == nil || httpResp.Request == nil {
		return nil
	}
	return httpResp.Request.URL
}

func readResponseStatusCode(httpResp *http.Response) int {
	if httpResp == nil {
		return 0
	}
	return httpResp.StatusCode
}
