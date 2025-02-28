// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package dcim

//go:generate mockgen -destination ../mocks/dcim.go -package mocks github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim DCIM

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"crypto/tls"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
	"github.com/netbox-community/go-netbox/netbox/client/virtualization"
	"github.com/netbox-community/go-netbox/netbox/models"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	NetBoxTransportProtocolEnv = "NETBOX_TRANSPORT_PROTOCOL"
	TransportDebug             = false
	Timeout                    = 30 * time.Second
	NetBoxAddressEnvVar        = "NETBOX_HOST"
	DeviceNameEnvVar           = "DEVICE_NAME"
	DeviceIdEnvVar             = "DEVICE_ID"
	RackNameEnvVar             = "RACK_NAME"
	ClusterNameEnvVar          = "CLUSTER_NAME"
	RegionEnvVar               = "ENROLLMENT_REGION"
	AvailabilityZoneEnvVar     = "ENROLLMENT_AZ"
	InsecureSkipVerifyEnvVar   = "NETBOX_INSECURE_SKIP_VERIFY"
)

var (
	TransportScheme = util.GetEnv(NetBoxTransportProtocolEnv, "https")
)

type Devices struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Device    `json:"results"`
}

type Device struct {
	ID         int64  `json:"id"`
	URL        string `json:"url"`
	Display    string `json:"display"`
	Name       string `json:"name"`
	DeviceType struct {
		ID           int    `json:"id"`
		URL          string `json:"url"`
		Display      string `json:"display"`
		Manufacturer struct {
			ID      int    `json:"id"`
			URL     string `json:"url"`
			Display string `json:"display"`
			Name    string `json:"name"`
			Slug    string `json:"slug"`
		} `json:"manufacturer"`
		Model string `json:"model"`
		Slug  string `json:"slug"`
	} `json:"device_type"`
	Role struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Display string `json:"display"`
		Name    string `json:"name"`
		Slug    string `json:"slug"`
	} `json:"role"`
	DeviceRole struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Display string `json:"display"`
		Name    string `json:"name"`
		Slug    string `json:"slug"`
	} `json:"device_role"`
	Tenant   interface{} `json:"tenant"`
	Platform interface{} `json:"platform"`
	Serial   string      `json:"serial"`
	AssetTag interface{} `json:"asset_tag"`
	Site     struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Display string `json:"display"`
		Name    string `json:"name"`
		Slug    string `json:"slug"`
	} `json:"site"`
	Location interface{} `json:"location"`
	Rack     struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Display string `json:"display"`
		Name    string `json:"name"`
	} `json:"rack"`
	Position     interface{} `json:"position"`
	Face         interface{} `json:"face"`
	Latitude     interface{} `json:"latitude"`
	Longitude    interface{} `json:"longitude"`
	ParentDevice interface{} `json:"parent_device"`
	Status       struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"status"`
	Airflow        interface{} `json:"airflow"`
	PrimaryIP      interface{} `json:"primary_ip"`
	PrimaryIP4     interface{} `json:"primary_ip4"`
	PrimaryIP6     interface{} `json:"primary_ip6"`
	OobIP          interface{} `json:"oob_ip"`
	Cluster        interface{} `json:"cluster"`
	VirtualChassis interface{} `json:"virtual_chassis"`
	VcPosition     interface{} `json:"vc_position"`
	VcPriority     interface{} `json:"vc_priority"`
	Description    string      `json:"description"`
	Comments       string      `json:"comments"`
	ConfigTemplate interface{} `json:"config_template"`
	ConfigContext  struct {
	} `json:"config_context"`
	LocalContextData       interface{}        `json:"local_context_data"`
	Tags                   []interface{}      `json:"tags"`
	CustomFields           DeviceCustomFields `json:"custom_fields"`
	Created                time.Time          `json:"created"`
	LastUpdated            time.Time          `json:"last_updated"`
	ConsolePortCount       int                `json:"console_port_count"`
	ConsoleServerPortCount int                `json:"console_server_port_count"`
	PowerPortCount         int                `json:"power_port_count"`
	PowerOutletCount       int                `json:"power_outlet_count"`
	InterfaceCount         int                `json:"interface_count"`
	FrontPortCount         int                `json:"front_port_count"`
	RearPortCount          int                `json:"rear_port_count"`
	DeviceBayCount         int                `json:"device_bay_count"`
	ModuleBayCount         int                `json:"module_bay_count"`
	InventoryItemCount     int                `json:"inventory_item_count"`
}

type DeviceCustomFields struct {
	BMEnrollmentComment   string             `json:"bm_enrollment_comment,omitempty"`
	BMEnrollmentNamespace string             `json:"bm_enrollment_namespace,omitempty"`
	BMEnrollmentStatus    BMEnrollmentStatus `json:"bm_enrollment_status,omitempty"`
	BMValidationReportURL string             `json:"bm_validation_report_url,omitempty"`
	BMValidationStatus    string             `json:"bm_validation_status,omitempty"`
}

type BMEnrollmentStatus string

const (
	BMEnroll              BMEnrollmentStatus = "enroll"
	BMEnrolling           BMEnrollmentStatus = "enrolling"
	BMEnrollmentFailed    BMEnrollmentStatus = "enrollment-failed"
	BMEnrolled            BMEnrollmentStatus = "enrolled"
	BMDisenroll           BMEnrollmentStatus = "disenroll"
	BMDisenrolling        BMEnrollmentStatus = "disenrolling"
	BMDisenrollmentFailed BMEnrollmentStatus = "disenrollment-failed"
	BMDisenrolled         BMEnrollmentStatus = "disenrolled"
)

// Helper method to ensure the merge changes between the current custom fields with the newer changes.
func (custom *DeviceCustomFields) merge(currentCustomFields map[string]interface{}) {
	if custom.BMEnrollmentStatus == "" {
		custom.BMEnrollmentStatus = BMEnrollmentStatus(fmt.Sprintf("%v", currentCustomFields["bm_enrollment_status"]))
	}
	if custom.BMEnrollmentComment == "" {
		custom.BMEnrollmentComment = fmt.Sprintf("%v", currentCustomFields["bm_enrollment_comment"])
	}
	if custom.BMValidationStatus == "" {
		custom.BMValidationStatus = fmt.Sprintf("%v", currentCustomFields["bm_validation_status"])
	}
	if custom.BMValidationReportURL == "" {
		custom.BMValidationReportURL = fmt.Sprintf("%v", currentCustomFields["bm_validation_report_url"])
	}
}

type DCIM interface {
	GetBMCURL(ctx context.Context, deviceName string) (string, error)
	GetBMCMACAddress(ctx context.Context, deviceName string, bmcInterfaceName string) (string, error)
	GetDeviceRegionName(ctx context.Context, siteName string) (string, error)
	GetDeviceNamespace(ctx context.Context, deviceName string) (string, error)
	GetClusterSize(ctx context.Context, clusterName string, availabilityZone string) (int64, error)
	GetClusterNetworkMode(ctx context.Context, clusterName string, siteName string) (string, error)
	UpdateDeviceStatus(ctx context.Context, deviceName string, deviceId int64, status string, comments string) error
	UpdateDeviceCustomFields(ctx context.Context, deviceName string, deviceId int64, customfields interface{}) error
	GetDeviceId(ctx context.Context, deviceName string) (int64, error)
	GetDeviceSwitchIPAddress(ctx context.Context, deviceName string, deviceSwitchName string, switchInterface string) (string, error)
	NetBoxCustomGetRequest(ctx context.Context, path string, queryMap map[string]string) ([]byte, error)
	NetBoxCustomPatchRequest(ctx context.Context, path string, data []byte) error
	GetCustomField(ctx context.Context, path string, customField string, customFieldValue string, region string, site string) ([]byte, error)
	PatchDeviceCustomFields(ctx context.Context, path string, id int64, name string, customFields DeviceCustomFields) error
	UpdateBMValidationStatus(ctx context.Context, id int64, name string, customFields DeviceCustomFields) error
}

type NetBox struct {
	client      *client.NetBoxAPI
	httpClient  http.Client
	httpRequest *http.Request
}

var _ DCIM = (*NetBox)(nil)

func NewNetBoxClient(ctx context.Context, token string, custom bool) (*NetBox, error) {
	// get netbox host
	netboxHost := os.Getenv(NetBoxAddressEnvVar)
	if netboxHost == "" {
		return nil, fmt.Errorf("failed to get the netbox host")
	}

	insecureSkipVerify, err := strconv.ParseBool(util.GetEnv(InsecureSkipVerifyEnvVar, "true"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env InsecureSkipVerifyEnvVar")
	}
	// netbox connection
	transport := httptransport.New(netboxHost, client.DefaultBasePath, []string{TransportScheme})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+token)
	httpTransport := transport.Transport.(*http.Transport)
	httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	if TransportDebug {
		transport.SetDebug(true)
	}

	nb := &NetBox{
		client: client.New(transport, nil),
	}

	// Add a custom client and request if custom is set to true.
	if custom {
		defaultTransport := http.DefaultTransport.(*http.Transport)

		// Create new Transport that ignores self-signed SSL
		customTransport := &http.Transport{
			Proxy:                 defaultTransport.Proxy,
			DialContext:           defaultTransport.DialContext,
			MaxIdleConns:          200,
			IdleConnTimeout:       defaultTransport.IdleConnTimeout,
			ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
			TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecureSkipVerify},
		}

		customClient := http.Client{Timeout: 60 * time.Second, Transport: customTransport}

		req, err := http.NewRequest("", "", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create new HTTP request: %v", err)
		}
		req.URL.Host = netboxHost
		req.URL.Scheme = TransportScheme
		authHeaderValue := fmt.Sprintf("TOKEN %s", token)
		req.Header.Add("Authorization", authHeaderValue)
		req.Header.Add("Content-Type", "application/json")

		nb.httpClient = customClient
		nb.httpRequest = req
	}

	return nb, nil
}

func (n *NetBox) NetBoxCustomGetRequest(ctx context.Context, path string, queryMap map[string]string) ([]byte, error) {

	httpRequest := n.httpRequest.Clone(ctx)

	// set request method to Patch
	httpRequest.Method = http.MethodGet

	// set request path
	if path != "" {
		httpRequest.URL.Path = path
	} else {
		httpRequest.URL.Path = "/"
	}

	// set request query
	if queryMap != nil {
		queryValues := httpRequest.URL.Query()
		for key, value := range queryMap {
			queryValues.Add(key, value)
		}
		httpRequest.URL.RawQuery = queryValues.Encode()
	} else {
		httpRequest.URL.RawQuery = ""
	}

	// print request if debug is set to true
	if TransportDebug {
		res, err := httputil.DumpRequest(httpRequest, true)
		if err != nil {
			fmt.Printf("failed to print http request: %v", err)
		}
		fmt.Printf("\nRequest: \n%s\n", string(res))
	}

	// netbox get request
	response, err := func(req *http.Request) (*http.Response, error) {
		httpRetryAttempt := 3
		var response *http.Response
		var err error
		for httpRetryAttempt > 0 {
			response, err = n.httpClient.Do(httpRequest)
			if err != nil {
				httpRetryAttempt -= 1
				fmt.Printf("netbox connection error. Retrying. error: %v", err)
				continue
			}
			if response == nil || response.StatusCode != http.StatusOK {
				httpRetryAttempt -= 1
				fmt.Printf("failed to get a successful response from the netbox. Retrying. response_status: %s, error: %v", response.Status, err)
				continue
			}
			return response, nil
		}
		return nil, fmt.Errorf("Failed to get connection or response from the netbox")
	}(httpRequest)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return body, nil
}

func (n *NetBox) NetBoxCustomPatchRequest(ctx context.Context, path string, data []byte) error {

	httpRequest := n.httpRequest.Clone(ctx)

	// set request method to Patch
	httpRequest.Method = http.MethodPatch

	//set content length
	httpRequest.Header.Add("Content-Length", strconv.FormatInt(int64(len(data)), 10))

	// Data
	httpRequest.Body = io.NopCloser(bytes.NewReader(data))
	defer httpRequest.Body.Close()

	// set request path
	if path != "" {
		httpRequest.URL.Path = path
	} else {
		httpRequest.URL.Path = "/"
	}

	// print request if debug is set to true
	if TransportDebug {
		res, err := httputil.DumpRequest(httpRequest, true)
		if err != nil {
			fmt.Printf("failed to print http patch request: %v", err)
		}
		fmt.Printf("\nRequest: \n%s\n", string(res))
	}

	// netbox get request
	err := func(req *http.Request) error {
		httpRetryAttempt := 3
		var response *http.Response
		var err error
		for httpRetryAttempt > 0 {
			response, err = n.httpClient.Do(httpRequest)
			if err != nil {
				httpRetryAttempt -= 1
				fmt.Printf("netbox connection error during patching. Retrying. error: %v", err)
				continue
			}
			if response == nil || response.StatusCode != http.StatusOK {
				httpRetryAttempt -= 1
				fmt.Printf("failed to get a successful response from the netbox during patching. Retrying. response_status: %s, error: %v", response.Status, err)
				continue
			}
			break
		}
		return nil
	}(httpRequest)

	if err != nil {
		return err
	} else {
		return nil
	}
}

func (n *NetBox) GetCustomField(ctx context.Context, path string, customField string, customFieldValue string, region string, site string) ([]byte, error) {

	// prepend custom field with cf
	customField = fmt.Sprintf("cf_%s", customField)

	// query parameters
	queryMap := make(map[string]string)
	queryMap[customField] = customFieldValue
	queryMap["region"] = region
	queryMap["site"] = site

	// get devices
	devicesBytesData, err := n.NetBoxCustomGetRequest(ctx, path, queryMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get netbox data: %v", err)
	}
	return devicesBytesData, nil
}

func (n *NetBox) PatchDeviceCustomFields(ctx context.Context, path string, id int64, name string, customFields DeviceCustomFields) error {

	data, err := json.Marshal(
		struct {
			ID           int64              `json:"id"`
			Name         string             `json:"name"`
			CustomFields DeviceCustomFields `json:"custom_fields"`
		}{
			ID:           id,
			Name:         name,
			CustomFields: customFields,
		})
	if err != nil {
		fmt.Printf("failed to marshal custom fields: %v", err)
	}

	// get devices
	err = n.NetBoxCustomPatchRequest(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to patch custom fields: %v", err)
	}

	return nil
}

func (n *NetBox) UpdateBMValidationStatus(ctx context.Context, id int64, name string, customFields DeviceCustomFields) error {

	// request path
	path := fmt.Sprintf("/api/dcim/devices/%d/", id)

	err := n.PatchDeviceCustomFields(ctx, path, id, name, customFields)
	if err != nil {
		return fmt.Errorf("failed to update the baremetal validation status: %v", err)
	}

	return nil
}

func (n *NetBox) GetBMCURL(ctx context.Context, deviceName string) (string, error) {
	log := log.FromContext(ctx).WithName("NetBox.GetBMCURL")
	log.Info("Getting BMC URL", "deviceName", deviceName)

	bmcInterfaceName := "BMC"

	// list interface filter parameters
	listParams := dcim.NewDcimInterfacesListParams()
	listParams.SetTimeout(Timeout)
	listParams.SetDevice(&deviceName)
	listParams.SetName(&bmcInterfaceName)

	// get BMC interface
	interfaceList, err := n.client.Dcim.DcimInterfacesList(listParams, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get the interface list. error: %v", err)
	}

	if *interfaceList.Payload.Count == 0 {
		return "", fmt.Errorf("failed to get the bmc interface")
	}

	if interfaceList.Payload.Results[0].Label == "" {
		return "", fmt.Errorf("failed to get a valid bmc interface URL label")
	}

	bmcURL := interfaceList.Payload.Results[0].Label
	log.Info("Found BMC URL", "bmcUrl", bmcURL)
	return bmcURL, nil
}

func (n *NetBox) UpdateDeviceCustomFields(ctx context.Context, deviceName string, deviceID int64, customfields interface{}) error {
	log := log.FromContext(ctx).WithName("NetBox.UpdateDeviceCustomFields")
	log.Info("Getting device", "deviceID", deviceID, "deviceName", &deviceName)
	device, err := n.getDevice(ctx, deviceName, deviceID)
	if err != nil {
		return err
	}
	log.Info("Found device", "device", device)
	newCustomFields, ok := customfields.(*DeviceCustomFields)
	if !ok {
		fmt.Println("failed to convert to DeviceCustomFields")
		return err
	}

	if device.CustomFields != nil {
		currentCustomFields, ok := device.CustomFields.(map[string]interface{})
		if !ok {
			fmt.Println("failed to read the current CustomFields")
			return err
		}
		//Ensure we merge the old custom fields with the newer changes.
		newCustomFields.merge(currentCustomFields)

	}
	updateParams := dcim.NewDcimDevicesPartialUpdateParams().
		WithID(device.ID).
		WithData(&models.WritableDeviceWithConfigContext{
			ID:           device.ID,
			Name:         device.Name,
			DeviceType:   &device.DeviceType.ID,
			Site:         &device.Site.ID,
			CustomFields: newCustomFields,
		})

	log.Info("Updating device", "data", updateParams.Data)

	res, err := n.client.Dcim.DcimDevicesPartialUpdate(updateParams, nil)
	if err != nil {
		return fmt.Errorf("failed to update device's customfields': %v", err)
	}
	if res.IsSuccess() {
		return nil
	} else {
		return fmt.Errorf("failed to update device's customfields': %v", res)
	}
}

func (n *NetBox) UpdateDeviceStatus(ctx context.Context, deviceName string, deviceID int64, status string, comments string) error {
	log := log.FromContext(ctx).WithName("NetBox.updateDeviceStatus")
	log.Info("Getting device", "deviceID", deviceID, "deviceName", &deviceName)

	device, err := n.getDevice(ctx, deviceName, deviceID)
	if err != nil {
		return err
	}

	log.Info("Found device", "device", device)

	updateParams := dcim.NewDcimDevicesPartialUpdateParams().
		WithID(device.ID).
		WithData(&models.WritableDeviceWithConfigContext{
			ID:         device.ID,
			Name:       device.Name,
			DeviceType: &device.DeviceType.ID,
			Rack:       &device.Rack.ID,
			Site:       &device.Site.ID,
			Status:     status,
			Comments:   fmt.Sprintf("%s  -- last run: kubectl logs -f %s -n idcs-enrollment", comments, os.Getenv("HOSTNAME")),
		})

	log.Info("Updating device", "data", updateParams.Data)

	res, err := n.client.Dcim.DcimDevicesPartialUpdate(updateParams, nil)
	if err != nil {
		return fmt.Errorf("failed to update device %s status to %q: %v", deviceName, status, err)
	}
	if res.IsSuccess() {
		return nil
	} else {
		return fmt.Errorf("failed to update device %s status to %q: %v", deviceName, status, res)
	}
}

func (n *NetBox) GetDeviceRegionName(ctx context.Context, siteName string) (string, error) {
	log := log.FromContext(ctx).WithName("NetBox.GetDeviceRegionName")
	log.Info("Getting device's region name", "siteName", siteName)

	// list interface filter parameters
	listParams := dcim.NewDcimSitesListParams()
	listParams.SetTimeout(Timeout)
	listParams.SetName(&siteName)

	// get BMC interface
	sitesList, err := n.client.Dcim.DcimSitesList(listParams, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get the sites list. error: %v", err)
	}
	if *sitesList.Payload.Count == 0 {
		return "", fmt.Errorf("failed to get the sites")
	}

	if sitesList.Payload.Results[0].Region.Name == nil {
		return "", fmt.Errorf("failed to get a valid region name")
	}

	region := *sitesList.Payload.Results[0].Region.Name
	log.Info("Found Device's region", "region", region)
	return region, nil
}

func (n *NetBox) GetDeviceNamespace(ctx context.Context, deviceName string) (string, error) {
	log := log.FromContext(ctx).WithName("NetBox.GetDeviceNamespace")
	log.Info("Getting device's GetDeviceNamespace", "deviceName", deviceName)

	listParams := dcim.NewDcimDevicesListParams()
	listParams.SetTimeout(Timeout)
	listParams.SetName(&deviceName)

	deviceList, err := n.client.Dcim.DcimDevicesList(listParams, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get the devices list. error: %v", err)
	}
	if *deviceList.Payload.Count == 0 {
		return "", fmt.Errorf("failed to get the device")
	}
	if deviceList.Payload.Results[0].CustomFields == nil {
		return "", fmt.Errorf("failed to get the custom fields")
	}
	customFields, found := deviceList.Payload.Results[0].CustomFields.(map[string]interface{})
	if !found {
		return "", fmt.Errorf("failed to read the custom fields")
	}

	value, found := customFields["bm_enrollment_namespace"]
	if !found {
		return "", fmt.Errorf("missing bm_enrollment_namespace in custom_fields: %+v", customFields)
	}
	if value == nil {
		return "", nil
	}
	namespace, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("failed to read the namespace")
	}

	return namespace, nil
}

func (n *NetBox) GetClusterSize(ctx context.Context, clusterName string, siteName string) (int64, error) {
	log := log.FromContext(ctx).WithName("NetBox.GetClusterSize")
	log.Info("Getting cluster name", "clusterName", clusterName)

	// list interface filter parameters
	listParams := virtualization.NewVirtualizationClustersListParams()
	listParams.SetName(&clusterName)
	listParams.SetSite(&siteName)
	listParams.SetTimeout(Timeout)

	// get BMC interface
	clustersList, err := n.client.Virtualization.VirtualizationClustersList(listParams, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get the clusters list. error: %v", err)
	}
	if *clustersList.Payload.Count == 0 {
		return 0, fmt.Errorf("failed to get cluster with name %s in availability zone %s", clusterName, siteName)
	}

	if *clustersList.Payload.Count > 1 {
		return 0, fmt.Errorf("more than 1 cluster with the same name %s in availability zone %s", clusterName, siteName)
	}

	if clustersList.Payload.Results[0].DeviceCount == 0 {
		return 0, fmt.Errorf("device count is 0 for cluster %s in availability zone %s", clusterName, siteName)
	}

	clusterSize := clustersList.Payload.Results[0].DeviceCount
	log.Info("cluster size", "cluster_size", clusterSize, "cluster_name", clusterName, "availability_zone", siteName)
	return clusterSize, nil
}

func (n *NetBox) GetClusterNetworkMode(ctx context.Context, clusterName string, siteName string) (string, error) {
	log := log.FromContext(ctx).WithName("NetBox.GetClusterNetworkMode")
	log.Info("Getting cluster name", "clusterName", clusterName, "siteName", siteName)

	listParams := virtualization.NewVirtualizationClustersListParams()
	listParams.SetName(&clusterName)
	listParams.SetSite(&siteName)
	listParams.SetTimeout(Timeout)

	// get cluster name
	clustersList, err := n.client.Virtualization.VirtualizationClustersList(listParams, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get the clusters list. error: %v", err)
	}

	if *clustersList.Payload.Count == 0 {
		return "", fmt.Errorf("failed to get cluster with name %s in availability zone %s", clusterName, siteName)
	}

	if *clustersList.Payload.Count > 1 {
		return "", fmt.Errorf("more than 1 cluster with the same name %s in availability zone %s", clusterName, siteName)
	}

	if clustersList.Payload.Results[0].CustomFields == nil {
		return "", fmt.Errorf("missing custom fields for cluster %s in availability zone %s", clusterName, siteName)
	}

	customFields, found := clustersList.Payload.Results[0].CustomFields.(map[string]interface{})
	if !found {
		return "", fmt.Errorf("failed to read the custom fields for cluster %s in availability zone %s", clusterName, siteName)
	}

	value, found := customFields["bm_network_mode"]
	if !found {
		return "", fmt.Errorf("missing bm_network_mode custom field for cluster %+v in availability zone %s", customFields, siteName)
	}
	if value == nil {
		return "", nil
	}

	networkMode, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("failed to read the network mode for cluster %s in availability zone %s", clusterName, siteName)
	}

	return networkMode, nil
}

func (n *NetBox) GetBMCMACAddress(ctx context.Context, deviceName string, interfaceName string) (string, error) {
	log := log.FromContext(ctx).WithName("NetBox.GetBMCMACAddress")
	log.Info("Getting BMC MAC address", "deviceName", deviceName, "interfaceName", interfaceName)

	// list interface filter parameters
	listParams := dcim.NewDcimInterfacesListParams()
	listParams.SetTimeout(Timeout)
	listParams.SetDevice(&deviceName)
	listParams.SetName(&interfaceName)

	// get BMC interface
	interfaceList, err := n.client.Dcim.DcimInterfacesList(listParams, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get the interface list. error: %v", err)
	}

	if *interfaceList.Payload.Count == 0 {
		return "", fmt.Errorf("failed to get the %s interface", interfaceName)
	}

	if interfaceList.Payload.Results[0].MacAddress == nil {
		return "", fmt.Errorf("missing MAC Address for the %s interface", interfaceName)
	}

	if len(*interfaceList.Payload.Results[0].MacAddress) < 12 { // Could be ff:ff:ff:ff:ff:ff, ff-ff-ff-ff-ff-ff, ffffffffffff
		return "", fmt.Errorf("missing or invalid MAC Address for the %s interface", interfaceName)
	}

	macAddress := *interfaceList.Payload.Results[0].MacAddress
	log.Info("Found BMC MAC address", "MacAddress", macAddress)

	return util.NormalizeMACAddress(macAddress), nil
}

func (n *NetBox) GetDeviceId(ctx context.Context, deviceName string) (int64, error) {
	deviceListParams := dcim.NewDcimDevicesListParams().WithName(&deviceName)
	deviceList, err := n.client.Dcim.DcimDevicesList(deviceListParams, nil)
	if err != nil {
		return -1, fmt.Errorf("unable to get devices list: %v", err)
	}
	if *deviceList.Payload.Count == 0 {
		return -1, fmt.Errorf("no matching devices found")
	}
	if *deviceList.Payload.Count > 1 {
		return -1, fmt.Errorf("more than one matching device found")
	}
	return deviceList.Payload.Results[0].ID, nil
}

func (n *NetBox) getDevice(ctx context.Context, deviceName string, deviceID int64) (*models.DeviceWithConfigContext, error) {
	deviceListParams := dcim.NewDcimDevicesListParams().WithName(&deviceName)
	deviceList, err := n.client.Dcim.DcimDevicesList(deviceListParams, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get devices list: %v", err)
	}
	if *deviceList.Payload.Count == 0 {
		return nil, fmt.Errorf("no matching devices found")
	}
	found := deviceList.Payload.Results[0]
	if deviceID != found.ID {
		return nil, fmt.Errorf("unable to find a device %q with ID %q", deviceName, deviceID)
	}

	return found, nil
}

func (n *NetBox) GetDeviceSwitchIPAddress(ctx context.Context, deviceName string, deviceSwitchName string, switchInterface string) (string, error) {
	log := log.FromContext(ctx).WithName("Netbox.GetDeviceSwitchIPAddress")
	log.Info("Getting switch IP addresses", "deviceName", deviceName, "deviceSwitchName", deviceSwitchName)

	// get switch
	switchNameParams := dcim.NewDcimDevicesListParams().WithName(&deviceSwitchName)
	switchNameParams.SetTimeout(Timeout)
	switchNameList, err := n.client.Dcim.DcimDevicesList(switchNameParams, nil)
	if err != nil {
		return "", fmt.Errorf("unable to get switches list: %v", err)
	}

	if *switchNameList.Payload.Count == 0 {
		return "", fmt.Errorf("no matching device switch found")
	}

	if *switchNameList.Payload.Count > 1 {
		return "", fmt.Errorf("more than 1 matching switches found for %s and switch interface %s", deviceName, switchInterface)
	}

	deviceSwitch := switchNameList.Payload.Results[0]
	log.Info("Found device switch", "deviceSwitch", deviceSwitch.Name, "switchInteface", switchInterface)

	// get switch IP
	ipAddressListParams := ipam.NewIpamIPAddressesListParams().WithDevice(deviceSwitch.Name)
	ipAddressListParams.SetInterface(&switchInterface)
	ipAddressListParams.SetTimeout(Timeout)
	ipAddressList, err := n.client.Ipam.IpamIPAddressesList(ipAddressListParams, nil)
	if err != nil {
		return "", fmt.Errorf("unable to get IP address list: %v", err)
	}

	if *ipAddressList.Payload.Count == 0 {
		return "", fmt.Errorf("no IP address found for switch %s and switch interface %s", *deviceSwitch.Name, switchInterface)
	}

	switchIPAddress := ipAddressList.Payload.Results[0]
	log.Info("Found device switch IP Address", "switchIPAddress.Address", switchIPAddress.Address, "switchInterface", switchInterface)
	return *switchIPAddress.Address, nil
}
