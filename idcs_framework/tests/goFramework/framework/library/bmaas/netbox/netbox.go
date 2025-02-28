package netbox

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	vault "goFramework/framework/library/bmaas/vault"

	netbox "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	runtimeclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	NetBoxTransportProtocolEnv = "NETBOX_TRANSPORT_PROTOCOL"
	TransportDebug             = false
	pollingInterval            = 1 * time.Second
)

var once sync.Once
var netboxClient *client.NetBoxAPI
var TransportScheme = util.GetEnv(NetBoxTransportProtocolEnv, "https")
var transport *httptransport.Runtime

func GetNetboxClient() *client.NetBoxAPI {
	once.Do(func() {
		var err error
		netboxClient, err = initNetboxClient()
		if err != nil {
			log.WithError(err).Error("Failed to get Netbox client")
		}
	})
	return netboxClient
}

func initNetboxClient() (*client.NetBoxAPI, error) {
	// get netbox host
	netboxHost := os.Getenv("NETBOX_HOST")
	if netboxHost == "" {
		fmt.Printf("failed to get the netbox host")
		return nil, fmt.Errorf("failed to get the netbox host")
	}

	// get vault client token
	vaultClient, err := vault.GetVaultClient()
	if err != nil {
		return nil, fmt.Errorf("falied to get Vault client: %v", err)
	}

	// netbox region
	// From IDC_ENV=kind-jenkins
	region := "us-dev-1"
	secretPath := fmt.Sprintf("%s/baremetal/enrollment/netbox", region)
	// get netbox token
	netboxToken, err := vault.GetNetBoxSecret(vaultClient, secretPath)
	if err != nil {
		return nil, fmt.Errorf("falied to get Netbox token: %v", err)
	}

	// netbox connection
	transport = httptransport.New(netboxHost, client.DefaultBasePath, []string{TransportScheme})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+netboxToken)
	if TransportDebug {
		transport.SetDebug(true)
	}
	httpTransport := transport.Transport.(*http.Transport)
	httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	netboxClient = client.New(transport, nil)

	return netboxClient, nil
}

func GetDevicesList() (*dcim.DcimDevicesListOK, error) {
	req := dcim.NewDcimDevicesListParams()
	devicesList, err := GetNetboxClient().Dcim.DcimDevicesList(req, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get devices list: %v", err)
	}
	return devicesList, nil
}

func GetBMCInterfaceForDevice(deviceName string) (*models.Interface, error) {
	bmcInterfaceName := "BMC"
	req := dcim.NewDcimInterfacesListParams()
	req.SetDevice(&deviceName)
	req.SetName(&bmcInterfaceName)
	interfacesList, err := GetNetboxClient().Dcim.DcimInterfacesList(req, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get %s interfaces list for device %s: %v", bmcInterfaceName, deviceName, err)
	}

	if *interfacesList.Payload.Count == 0 {
		return nil, fmt.Errorf("cannot get the %s interface", bmcInterfaceName)
	}

	return interfacesList.Payload.Results[0], nil
}

func GetDeviceByName(deviceName string) (*models.DeviceWithConfigContext, error) {
	deviceParam := dcim.NewDcimDevicesListParams()
	deviceParam.Name = &deviceName

	deviceList, err := GetNetboxClient().Dcim.DcimDevicesList(deviceParam, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot get devices list: %v", err)
	}

	if *deviceList.Payload.Count == 0 {
		return nil, fmt.Errorf("cannot get the %s device", deviceName)
	}

	return deviceList.Payload.Results[0], nil
}

func UpdateInterfaceLabel(deviceInterface *models.Interface, label string) error {
	updateParams := dcim.NewDcimInterfacesUpdateParams()
	data := models.WritableInterface{}
	data.Device = &deviceInterface.Device.ID
	data.Name = deviceInterface.Name
	data.Label = label
	data.ID = deviceInterface.ID
	data.Type = deviceInterface.Type.Value
	data.TaggedVlans = []int64{}
	data.WirelessLans = []int64{}
	data.Vdcs = []int64{}

	updateParams.ID = deviceInterface.ID
	updateParams.WithData(&data)
	log.Info("updateData", "data", data)
	res, err := GetNetboxClient().Dcim.DcimInterfacesUpdate(updateParams, nil)
	if err != nil {
		return fmt.Errorf("failed to update interface label: %v", err)
	}
	if res.IsSuccess() {
		return nil
	} else {
		return fmt.Errorf("failed to update interface label: %v", res)
	}
}

func UpdateDeviceCustomFields(deviceName string, customfields interface{}) error {
	device, err := GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	updateParams := dcim.NewDcimDevicesPartialUpdateParams().
		WithID(device.ID).
		WithData(&models.WritableDeviceWithConfigContext{
			ID:           device.ID,
			Name:         device.Name,
			DeviceType:   &device.DeviceType.ID,
			Site:         &device.Site.ID,
			CustomFields: customfields,
		})

	res, err := GetNetboxClient().Dcim.DcimDevicesPartialUpdate(updateParams, nil)
	if err != nil {
		return fmt.Errorf("failed to update device's customfields': %v", err)
	}
	if res.IsSuccess() {
		return nil
	} else {
		return fmt.Errorf("failed to update device's customfields': %v", res)
	}
}

func UpdateDeviceEnrollmentStatus(deviceName string, status netbox.BMEnrollmentStatus) error {
	return UpdateDeviceCustomFields(deviceName, &netbox.DeviceCustomFields{
		BMEnrollmentStatus: status,
	})
}

func UpdateDeviceStatus(deviceName string, status string) error {
	nbclient := GetNetboxClient()
	deviceParam := dcim.NewDcimDevicesListParams()
	deviceParam.Name = &deviceName
	deviceList, err := nbclient.Dcim.DcimDevicesList(deviceParam, nil)
	if err != nil {
		return fmt.Errorf("cannot get devices list: %v", err)
	}
	updateParams := dcim.NewDcimDevicesPartialUpdateParams()
	data := models.WritableDeviceWithConfigContext{}
	data.Name = &deviceName
	data.Status = status
	data.Comments = deviceList.Payload.Results[0].Comments
	data.ID = deviceList.Payload.Results[0].ID
	data.Rack = &deviceList.Payload.Results[0].Rack.ID

	data.Site = &deviceList.Payload.Results[0].Site.ID
	data.DeviceType = &deviceList.Payload.Results[0].DeviceType.ID

	updateParams.ID = deviceList.Payload.Results[0].ID
	updateParams.WithData(&data)
	res, err := nbclient.Dcim.DcimDevicesPartialUpdate(updateParams, nil)
	if err != nil {
		return fmt.Errorf("failed to update device %s status to %q: %v", deviceName, status, err)
	}
	if res.IsSuccess() {
		return nil
	} else {
		return fmt.Errorf("failed to update device %s status to %q: %v", deviceName, status, res)
	}
}

func UpdateDeviceRole(deviceName string, role int64) error {
	nbclient := GetNetboxClient()
	deviceParam := dcim.NewDcimDevicesListParams()
	deviceParam.Name = &deviceName

	deviceList, err := nbclient.Dcim.DcimDevicesList(deviceParam, nil)
	if err != nil {
		return fmt.Errorf("cannot get devices list: %v", err)
	}

	// override the device update params
	// use dcim.NewDcimDevicesUpdateParams() once the netbox go client library has full support for v3.6
	updateParams := &DcimDevicesUpdateParams{}

	//data := models.WritableDeviceWithConfigContext{}
	data := DcimDeviceData{}
	data.Name = &deviceName
	data.Status = *deviceList.Payload.Results[0].Status.Value
	data.Comments = deviceList.Payload.Results[0].Comments
	data.ID = deviceList.Payload.Results[0].ID
	data.DeviceRole = &role
	data.Role = &role
	data.Site = &deviceList.Payload.Results[0].Site.ID
	data.DeviceType = &deviceList.Payload.Results[0].DeviceType.ID
	updateParams.ID = deviceList.Payload.Results[0].ID
	data.Rack = &deviceList.Payload.Results[0].Rack.ID
	updateParams.WithData(&data)

	//res, err := nbclient.Dcim.DcimDevicesUpdate(updateParams, nil)
	res, err := DcimDevicesUpdate(updateParams)
	if err != nil {
		return fmt.Errorf("failed to update device %s role to roleID value %q: %v", deviceName, role, err)
	}
	if res.IsSuccess() {
		return nil
	} else {
		return fmt.Errorf("failed to update device %s role to roleID value %q: %v", deviceName, role, err)
	}
}

func CheckDeviceEnrollmentStatus(deviceName string, desired netbox.BMEnrollmentStatus, timeout time.Duration) (bool, error) {
	err := wait.PollImmediate(pollingInterval, timeout, func() (bool, error) {
		device, err := GetDeviceByName(deviceName)
		if err != nil {
			fmt.Printf("Unable to get BMH: %s, waiting... \n", deviceName)
			return false, nil
		}

		customfields, ok := device.CustomFields.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unable to get device %q customfields", deviceName)
		}
		status, ok := customfields["bm_enrollment_status"].(string)
		if !ok {
			return false, fmt.Errorf("unable to get device %q customfields.bm_enrollment_status", deviceName)
		}
		fmt.Printf("->-> Device: %s is %q, waiting to reach %q state\n", deviceName, status, desired)

		if status == string(desired) {
			fmt.Printf("***Device: %s is in status %q\n", deviceName, status)
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		fmt.Println("unable to verify the enrollment status:", err)
		return false, err
	}

	return true, nil
}

func CheckDeviceStatus(deviceName string, status string, timeout int) (bool, error) {
	statusReached := make(chan bool)

	var to time.Duration = (time.Duration(timeout) * time.Second)
	var reached bool
	var err error

	go func() {
		checkStatus(statusReached, deviceName, status)
		<-statusReached
	}()
	select {
	case res := <-statusReached:
		if res {
			reached = true
			err = nil
		}
	case <-time.After(to):
		reached = false
		err = fmt.Errorf("timeout waiting for %s to reach status: %s", deviceName, status)
	}

	return reached, err
}

func checkStatus(done chan bool, deviceName string, status string) {
	for {
		device, err := GetDeviceByName(deviceName)
		if err != nil {
			fmt.Printf("Unable to get BMH: %s, waiting... \n", deviceName)
			time.Sleep(time.Duration(pollingInterval))
			continue
		}
		currState := fmt.Sprint(*device.Status.Value)
		fmt.Printf("->-> Device: %s is %s, waiting to reach %s state\n", deviceName, currState, status)
		if fmt.Sprint(*device.Status.Value) == status {
			fmt.Printf("***Device: %s is in status %s\n", deviceName, status)
			break
		}
		time.Sleep(time.Duration(pollingInterval))
	}
	done <- true
}

// DcimDeviceData overrides a device data model to support a device update for NetBox v3.6
type DcimDeviceData struct {
	models.WritableDeviceWithConfigContext

	// Device role
	// Required: true
	Role *int64 `json:"role"`
}

type DcimDevicesUpdateParams struct {
	dcim.DcimDevicesUpdateParams
	Data *DcimDeviceData
}

func (o *DcimDevicesUpdateParams) WithData(data *DcimDeviceData) *DcimDevicesUpdateParams {
	o.SetData(data)
	return o
}

func (o *DcimDevicesUpdateParams) SetData(data *DcimDeviceData) {
	o.Data = data
}

func (o *DcimDevicesUpdateParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {
	if err := r.SetTimeout(runtimeclient.DefaultTimeout); err != nil {
		return err
	}
	var res []error
	if o.Data != nil {
		if err := r.SetBodyParam(o.Data); err != nil {
			return err
		}
	}

	// path param id
	if err := r.SetPathParam("id", swag.FormatInt64(o.ID)); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func DcimDevicesUpdate(params *DcimDevicesUpdateParams) (*dcim.DcimDevicesUpdateOK, error) {
	if params == nil {
		params = &DcimDevicesUpdateParams{}
	}
	op := &runtime.ClientOperation{
		ID:                 "dcim_devices_update",
		Method:             "PUT",
		PathPattern:        "/dcim/devices/{id}/",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &dcim.DcimDevicesUpdateReader{},
		AuthInfo:           nil,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}

	result, err := transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*dcim.DcimDevicesUpdateOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*dcim.DcimDevicesUpdateDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}
