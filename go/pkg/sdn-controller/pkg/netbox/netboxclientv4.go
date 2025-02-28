package netboxclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"github.com/jerryzhen01/go-netbox/v4"
)

type NetBoxClientV4 struct {
	client *netbox.APIClient
}

func (n *NetBoxClientV4) Version() string {
	return "4"
}

func (n *NetBoxClientV4) ListDevices(ctx context.Context, req ListDevicesRequest) ([]Device, error) {
	// TODO: do NOT provide invalid or non-existing roles values to the filter, Execute() will return 400 error.
	// need to further investigate if only role field has this issue and provide a solution.
	res, httpResp, err := n.client.DcimAPI.DcimDevicesList(ctx).
		Name(req.Filter.DeviceName).
		Id(req.Filter.DeviceID).
		// RegionId(req.Filter.RegionID). // this is not supported in Netbox v4
		Site(req.Filter.Site).
		SiteId(req.Filter.SiteID).
		Role(req.Filter.Role).
		RoleId(req.Filter.RoleID).
		DeviceType(req.Filter.DeviceType).
		DeviceTypeId(req.Filter.DeviceTypeID).
		Status(req.Filter.Status).
		Tag(req.Filter.Tag).
		Limit(0).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("DcimDevicesList failed. Request URL: %v, Error: %v, Response body: %v", readRequestURL(httpResp), err, readStartOfBody(httpResp))
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DcimDevicesList got unexpected http status code %v. Request URL: %v Response body: %v", readResponseStatusCode(httpResp), readRequestURL(httpResp), readStartOfBody(httpResp))
	}
	return generateV4DeviceResults(res.Results)
}

func generateV4DeviceResults(devices []netbox.DeviceWithConfigContext) ([]Device, error) {
	res := make([]Device, 0)
	for _, device := range devices {
		d := Device{
			Id:         device.GetId(),
			Url:        device.Url,
			Display:    device.Display,
			Name:       device.GetName(),
			DeviceType: device.GetDeviceType().Slug,
			Role:       device.GetRole().Slug,
			// DeviceRole:    device.GetDeviceRole().Slug, // not available in v4
			Site:          device.GetSite().Slug,
			Status:        string(*device.GetStatus().Value),
			Description:   device.GetDescription(),
			ConfigContext: device.GetConfigContext(),
			CustomFields:  device.GetCustomFields(),
		}
		res = append(res, d)
	}
	return res, nil
}

func (n *NetBoxClientV4) ListInterfaces(ctx context.Context, req ListInterfaceRequest) ([]Interface, error) {
	request := n.client.DcimAPI.DcimInterfacesList(ctx).
		Name(req.Filter.InterfaceName).
		Device(req.Filter.Device).
		DeviceId(req.Filter.DeviceID).
		DeviceRole(req.Filter.DeviceRole).
		Label(req.Filter.Label).
		Tag(req.Filter.Tag).
		Limit(0)

	if req.Filter.Cabled != nil {
		request = request.Cabled(*req.Filter.Cabled)
	}
	if req.Filter.Connected != nil {
		request = request.Cabled(*req.Filter.Connected)
	}

	request = request.NameN(req.Filter.InterfaceNameN)

	listRes, httpResp, err := request.Execute()
	if err != nil {
		return nil, fmt.Errorf("DcimInterfacesList failed. Request URL: %v, Error: %v, Response body: %v", readRequestURL(httpResp), err, readStartOfBody(httpResp))
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DcimInterfacesList got unexpected http status code %v. Request URL: %v, Response body: %v", readResponseStatusCode(httpResp), readRequestURL(httpResp), readStartOfBody(httpResp))
	}

	// perform extra filtering
	result := make([]netbox.Interface, 0)
	for i := range listRes.Results {
		if len(req.Filter.InterfaceNameRegex) > 0 && !utils.RegexMatch(req.Filter.InterfaceNameRegex, listRes.Results[i].Name) {
			continue
		}
		result = append(result, listRes.Results[i])
	}

	return generateV4InterfaceResults(result)
}

func generateV4InterfaceResults(networkInterfaces []netbox.Interface) ([]Interface, error) {
	res := make([]Interface, 0)
	for _, intf := range networkInterfaces {
		// parse linkpeer
		linkPeers := make([]LinkPeer, 0)
		for i := range intf.LinkPeers {
			linkPeer := &LinkPeer{}
			// Serialize the src to JSON.
			jsonData, err := json.Marshal(intf.LinkPeers[i])
			if err != nil {
				logger.Error(err, "failed to Marshal LinkPeers")
				continue
			}

			// Deserialize the JSON into the target.
			err = json.Unmarshal(jsonData, linkPeer)
			if err != nil {
				logger.Error(err, "failed to Unmarshal linkPeer")
				continue
			}
			linkPeers = append(linkPeers, *linkPeer)
		}

		i := Interface{
			Id:                   intf.GetId(),
			Url:                  intf.GetUrl(),
			Display:              intf.GetDisplay(),
			Name:                 intf.GetName(),
			Enabled:              intf.Enabled,
			Description:          intf.Description,
			LinkPeers:            linkPeers,
			AdditionalProperties: intf.AdditionalProperties,
		}
		res = append(res, i)
	}
	return res, nil
}

func (n *NetBoxClientV4) ListIPAddresses(ctx context.Context, req ListIPAddressesRequest) ([]IPAddress, error) {
	res, httpResp, err := n.client.IpamAPI.IpamIpAddressesList(ctx).
		Id(req.Filter.DeviceID).
		Device(req.Filter.DeviceName).
		DeviceId(req.Filter.DeviceID).
		Interface_(req.Filter.Interfaces).
		InterfaceId(req.Filter.InterfacesId).
		Vrf(req.Filter.VRF).
		Role(req.Filter.Role).
		Status(req.Filter.Status).
		Tag(req.Filter.Tag).
		Limit(0).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("IpamIpAddressesList failed. Request URL: %v, Error: %v, Response body: %v", readRequestURL(httpResp), err, readStartOfBody(httpResp))
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IpamIpAddressesList got unexpected http status code %v. Request URL: %v, Response body: %v", readResponseStatusCode(httpResp), readRequestURL(httpResp), readStartOfBody(httpResp))
	}
	return generateV4IPAddressResults(res.Results)
}

func generateV4IPAddressResults(ips []netbox.IPAddress) ([]IPAddress, error) {
	res := make([]IPAddress, 0)
	for _, ip := range ips {
		i := IPAddress{
			Id:      ip.GetId(),
			Url:     ip.GetUrl(),
			Display: ip.GetDisplay(),
			Family:  int32(*ip.GetFamily().Value),
			Address: ip.GetAddress(),
			// DnsName:              ip.DnsName, // not available in v4
			Description: ip.Description,
			// CustomFields:         ip.CustomFields, // not available in v4
			AdditionalProperties: ip.AdditionalProperties,
		}
		res = append(res, i)
	}
	return res, nil
}
