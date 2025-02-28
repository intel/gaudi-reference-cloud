package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	nc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/netbox"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"k8s.io/apimachinery/pkg/types"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NetboxController pulls the latest devices information from Netbox and create/update the SDN CRDs.
type NetboxController struct {
	sync.Mutex
	cfg              idcnetworkv1alpha1.SDNControllerConfig
	networkK8sClient client.Client

	netboxClient nc.SDNNetboxClient
	netboxToken  string
}

var netboxSwitchErrorCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "netbox_switch_error_counter",
		Help:        "Failed attempts to fetch switches from Netbox.",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var netboxSwitchPortsErrorCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "netbox_switch_ports_error_counter",
		Help:        "Failed attempts to fetch provider servers from Netbox.",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

func init() {
	metrics.Registry.MustRegister(netboxSwitchErrorCounter)
	metrics.Registry.MustRegister(netboxSwitchPortsErrorCounter)
}

func NewNetboxController(cfg idcnetworkv1alpha1.SDNControllerConfig, nwcpK8sClient client.Client) (*NetboxController, error) {
	if nwcpK8sClient == nil {
		return nil, fmt.Errorf("nwcp k8s client is not provided")
	}

	if len(cfg.ControllerConfig.NetboxTokenPath) == 0 {
		return nil, fmt.Errorf("NetboxTokenPath is not provided")
	}

	if len(cfg.ControllerConfig.NetboxServer) == 0 {
		return nil, fmt.Errorf("NetboxServer is not provided")
	}

	// read netbox key
	netboxTokenBytes, err := os.ReadFile(cfg.ControllerConfig.NetboxTokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Netbox key file %s", err.Error())
	}

	netboxClient, err := nc.NewSDNNetBoxClient(cfg.ControllerConfig.NetboxServer, string(netboxTokenBytes), cfg.ControllerConfig.NetboxClientInsecureSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client %s", err.Error())
	}

	return &NetboxController{
		cfg:              cfg,
		netboxClient:     netboxClient,
		netboxToken:      string(netboxTokenBytes),
		networkK8sClient: nwcpK8sClient,
	}, nil
}

func (d *NetboxController) ResetNetboxClientIfNeeded(err error) error {
	if strings.Contains(err.Error(), "cannot unmarshal") || strings.Contains(err.Error(), "no value given for required property") {
		return d.ResetNetboxClient()
	}
	return nil
}

func (d *NetboxController) ResetNetboxClient() error {
	d.Lock()
	defer d.Unlock()
	netboxClient, err := nc.NewSDNNetBoxClient(d.cfg.ControllerConfig.NetboxServer, d.netboxToken, d.cfg.ControllerConfig.NetboxClientInsecureSkipVerify)
	if err != nil {
		return fmt.Errorf("unable to create NetBox client: %s", err.Error())
	}
	d.netboxClient = netboxClient
	return nil
}

func (d *NetboxController) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("NetboxController.Start")
	logger.Info("", utils.LogFieldNetboxClientVersion, d.netboxClient.Version())

	ticker := time.NewTicker(time.Duration(d.cfg.ControllerConfig.SwitchImportPeriodInSec) * time.Second)
	var err error
	defer ticker.Stop()

	var switchesFilter *nc.DevicesFilter
	var providerServersFilter *nc.DevicesFilter
	var providerInterfacesFilter *nc.InterfacesFilter

	if d.cfg.ControllerConfig.SwitchImportSource == idcnetworkv1alpha1.SwitchImportSourceNetbox {
		switchesFilterBytes, err := os.ReadFile(d.cfg.ControllerConfig.NetboxSwitchesFilterFilePath)
		if err != nil {
			return fmt.Errorf("failed to read NetboxSwitchesFilterFilePath, %v", err)
		}
		switchesFilter = &nc.DevicesFilter{}
		err = json.Unmarshal(switchesFilterBytes, switchesFilter)
		if err != nil {
			return fmt.Errorf("unmarshal netboxDeviceFilter failed, %v", err)
		}
		logger.V(1).Info("switch filters", "filters", switchesFilter)
	}

	if d.cfg.ControllerConfig.SwitchPortImportSource == idcnetworkv1alpha1.SwitchPortImportSourceNetbox {
		// provider servers filters
		providerServersFilterBytes, err := os.ReadFile(d.cfg.ControllerConfig.NetboxProviderServersFilterFilePath)
		if err != nil {
			return fmt.Errorf("failed to read NetboxProviderServersFilterFilePath, %v", err)
		}
		providerServersFilter = &nc.DevicesFilter{}
		err = json.Unmarshal(providerServersFilterBytes, providerServersFilter)
		if err != nil {
			return fmt.Errorf("unmarshal netboxDeviceFilter failed, %v", err)
		}
		logger.V(1).Info("provider server filters", "filters", providerServersFilter)

		// provider interfaces filters (interfaces that connect to the provider servers)
		providerInterfacesFilterBytes, err := os.ReadFile(d.cfg.ControllerConfig.NetboxProviderInterfacesFilterFilePath)
		if err != nil {
			return fmt.Errorf("failed to read netboxProviderInterfacesFilterFilePath, %v", err)
		}
		providerInterfacesFilter = &nc.InterfacesFilter{}
		err = json.Unmarshal(providerInterfacesFilterBytes, providerInterfacesFilter)
		if err != nil {
			return fmt.Errorf("unmarshal netboxInterfacesFilter failed, %v", err)
		}
		logger.V(1).Info("provider interface filters", "filters", providerInterfacesFilter)
	}

	for {
		//////////////////////////////
		// sync the switches
		//////////////////////////////

		if d.cfg.ControllerConfig.SwitchImportSource == idcnetworkv1alpha1.SwitchImportSourceNetbox {
			logger.V(1).Info("start syncing Switch data with Netbox")
			// Note: tenant and provider may share switches. Make sure don't set the BGP community value for the switches that the global SDN maintain.
			err = d.syncSwitches(ctx, switchesFilter)
			if err != nil {
				netboxSwitchErrorCounter.Add(1)
				logger.Error(err, "failed to fetch switches from Netbox")
			}
		}

		//////////////////////////////
		// sync the provider switch ports
		//////////////////////////////
		if d.cfg.ControllerConfig.SwitchPortImportSource == idcnetworkv1alpha1.SwitchPortImportSourceNetbox {
			logger.V(1).Info("start syncing SwitchPort data with Netbox")
			err = d.syncSwitchPorts(ctx, providerServersFilter, providerInterfacesFilter, switchesFilter)
			if err != nil {
				netboxSwitchPortsErrorCounter.Add(1)
				logger.Error(err, "failed to fetch provider switch interfaces")
			}
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return fmt.Errorf("context is done, err: %v", ctx.Err())
		}
	}
}

func (d *NetboxController) syncSwitches(ctx context.Context, switchesFilter *nc.DevicesFilter) error {
	logger := log.FromContext(ctx).WithName("NetboxController.syncSwitches")
	switchesListReq := nc.ListDevicesRequest{
		Filter: switchesFilter,
	}
	switches, err := d.netboxClient.ListDevices(ctx, switchesListReq)
	if err != nil {
		logger.V(1).Info(fmt.Sprintf("failed to fetch switches from Netbox, %v", err), "filters", switchesListReq.Filter)
		// reset netbox client if needed
		resetErr := d.ResetNetboxClientIfNeeded(err)
		if resetErr != nil {
			logger.Info(fmt.Sprintf("failed to reset Netbox client, %v", err))
		}

		// return the actual error
		return err
	}

	filteredSwitches := make([]nc.Device, 0)
	for _, sw := range switches {
		if len(sw.Name) > 0 {
			filteredSwitches = append(filteredSwitches, sw)
		}
	}

	logger.V(1).Info(fmt.Sprintf("listDevices returned %v switches (%v after filtering)", len(switches), len(filteredSwitches)))
	return d.addOrUpdateSwitches(ctx, filteredSwitches)
}

type configContext struct {
	Search []string `json:"search"`
}

func (d *NetboxController) addOrUpdateSwitches(ctx context.Context, switches []nc.Device) error {
	logger := log.FromContext(ctx).WithName("NetboxController.addOrUpdateSwitches")

	switchesFailedToInit := make([]string, 0)
	for _, device := range switches {
		swName := strings.ToLower(device.GetName())
		logger.V(1).Info(fmt.Sprintf("add or update switch [%v]", swName))

		err, fqdn := netboxSwNameToFqdn(d.cfg.ControllerConfig.NetboxSwitchFQDNDomainName, d.cfg.ControllerConfig.DataCenter, swName, &device)
		if err != nil {
			logger.Error(err, "Failed to convert switch name to FQDN", "switchName", swName)
			continue
		}

		// Only add switches that match the expected format / datacenter.
		err = utils.ValidateSwitchFQDN(fqdn, d.cfg.ControllerConfig.DataCenter)
		if err != nil {
			logger.Error(err, "switch FQDN is invalid. Ignoring. ", utils.LogFieldSwitchFQDN, fqdn)
			switchesFailedToInit = append(switchesFailedToInit, fqdn)
			continue
		}

		// find the IP for this switch
		ip, err := d.getSwitchManagementIP(ctx, device.GetName())
		if err != nil || len(ip) == 0 {
			logger.V(1).Info(fmt.Sprintf("failed to get the switch IP for %v, will use FQDN %v instead, err %v", device.GetName(), fqdn, err))
			// if we cannot get the IP, use the fqdn instead
			ip = fqdn
		} else {
			logger.V(1).Info(fmt.Sprintf("found a management IP [%v] for switch [%v]", swName, ip))
		}

		existingSwitchCR := &idcnetworkv1alpha1.Switch{}
		key := types.NamespacedName{Name: fqdn, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		objectExist := true
		err = d.networkK8sClient.Get(ctx, key, existingSwitchCR)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				logger.Error(err, "failed to get switch CR", utils.LogFieldSwitchFQDN, fqdn)
				switchesFailedToInit = append(switchesFailedToInit, fqdn)
				continue
			}
			objectExist = false
		}

		if objectExist {
			// if the switch CR exists, check if the IP has been changed and update it when needed.
			if existingSwitchCR.Spec.Ip != ip {
				logger.Info("the existing switch CR's IP is different from Netbox, updating it...", utils.LogFieldSwitchFQDN, fqdn)
				// patch update the switch CR
				newSwitchCR := existingSwitchCR.DeepCopy()
				newSwitchCR.Spec.Ip = ip
				patch := client.MergeFrom(existingSwitchCR)
				if err := d.networkK8sClient.Patch(ctx, newSwitchCR, patch); err != nil {
					logger.Error(err, "patch update Switch CR failed")
				}
			}
		} else {
			switchCR := utils.NewSwitchTemplate(fqdn, ip)
			switchCR.Spec.EAPIConf.CredentialPath = d.cfg.ControllerConfig.SwitchSecretsPath
			err = d.networkK8sClient.Create(ctx, switchCR, &client.CreateOptions{})
			if err != nil {
				logger.Error(err, "create switch CR failed", utils.LogFieldSwitchFQDN, switchCR.Name)
				switchesFailedToInit = append(switchesFailedToInit, fqdn)
			}
		}
	}

	if len(switchesFailedToInit) > 0 {
		return fmt.Errorf("%v/%v of the switch CR failed to initialize. switch list: %v", len(switchesFailedToInit), len(switches), switchesFailedToInit)
	}

	return nil
}

func netboxSwNameToFqdn(cfgNetboxSwitchFQDNDomainName string, dataCenter string, swName string, switchDevice *nc.Device) (error, string) {
	var fqdn string
	var swNameLowercase = strings.ToLower(swName)
	if cfgNetboxSwitchFQDNDomainName == "configContextSearch" {
		// Older environments (pdx05, pdx09, pdx04, phx04) don't have the correct configContext.Search in Netbox, to resolve FQDN in DNS we need to use .internal-placeholder.com suffix.
		pattern := `^[a-zA-Z0-9]{8,9}-(zal|zas)[a-zA-Z0-9]{5}$` //eg. fxhb3p3r-zal0118a
		re := regexp.MustCompile(pattern)
		isMatch := re.MatchString(swNameLowercase)
		if isMatch {
			fqdn = fmt.Sprintf("%s.%s", swNameLowercase, "internal-placeholder.com")
		} else {
			// Newer environment, style of switch name is like "pdx11-s01-acsw001".
			if switchDevice == nil {
				return fmt.Errorf("switchDevice is nil"), ""
			}

			var configCtx configContext
			configCtxJson, err := json.Marshal(switchDevice.ConfigContext)
			if err != nil {
				//can't marshal configContext to JSON.
				return err, ""
			}
			err = json.Unmarshal(configCtxJson, &configCtx)
			if err != nil {
				// can't unmarshal configContext from JSON.
				return err, ""
			}
			for _, searchDomain := range configCtx.Search {
				fqdn = fmt.Sprintf("%s.%s", swNameLowercase, searchDomain)

				err := utils.ValidateSwitchFQDN(fqdn, dataCenter)
				if err == nil { // Found a searchDomain that creates a valid FQDN.
					break
				}
			}
		}
	} else {
		fqdn = fmt.Sprintf("%s.%s", swNameLowercase, cfgNetboxSwitchFQDNDomainName)
	}
	return nil, fqdn
}

func (d *NetboxController) syncSwitchPorts(ctx context.Context, providerServersFilter *nc.DevicesFilter, providerInterfacesFilter *nc.InterfacesFilter, switchesFilter *nc.DevicesFilter) error {
	logger := log.FromContext(ctx).WithName("NetboxController.syncSwitchPorts")
	providerServerListReq := nc.ListDevicesRequest{
		Filter: providerServersFilter,
	}

	providerServers, err := d.netboxClient.ListDevices(ctx, providerServerListReq)
	if err != nil {
		logger.V(1).Info(fmt.Sprintf("failed to fetch provider servers from Netbox, %v", err), "filters", providerServerListReq.Filter)
		// reset netbox client if needed. Call ResetNetboxClientIfNeeded once in this function should be enough.
		refreshErr := d.ResetNetboxClientIfNeeded(err)
		if refreshErr != nil {
			logger.Info(fmt.Sprintf("failed to reset Netbox client, %v", err))
		}
		return err
	}
	logger.V(1).Info(fmt.Sprintf("listDevices return %v provider servers", len(providerServers)))

	// If NetboxSwitchFQDNDomainName is dynamically generated, we need to know the FULL fqdn of each switch, for which we need the configContext for that device.
	var switchMap map[int32]*nc.Device = make(map[int32]*nc.Device)
	if d.cfg.ControllerConfig.NetboxSwitchFQDNDomainName == "configContextSearch" {
		switchesListReq := nc.ListDevicesRequest{
			Filter: switchesFilter,
		}
		switches, err := d.netboxClient.ListDevices(ctx, switchesListReq)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to fetch switches, needed for configContext \n"))
		}
		for i := range switches {
			switchMap[switches[i].GetId()] = &switches[i]
		}
	}

	// fetch the interfaces for each switch and create the k8s CRs.
	for i := range providerServers {
		providerServerID := providerServers[i].GetId()
		providerInterfaceFilterWithDeviceID := *providerInterfacesFilter
		// add device ID to the filter
		providerInterfaceFilterWithDeviceID.DeviceID = append(providerInterfaceFilterWithDeviceID.DeviceID, providerServerID)

		intListReq := nc.ListInterfaceRequest{
			Filter: &providerInterfaceFilterWithDeviceID,
		}

		interfaces, err := d.netboxClient.ListInterfaces(ctx, intListReq)
		if err != nil {
			logger.Error(err, fmt.Sprintf("ListInterfaces failed, device id: %+v \n", providerInterfaceFilterWithDeviceID.DeviceID))
			continue
		}
		logger.V(1).Info(fmt.Sprintf("listInterfaces return %v interfaces for provider server %v", len(interfaces), providerServers[i].GetName()))

		for j := range interfaces {
			err = d.addOrUpdateSwitchPort(ctx, interfaces[j], switchMap)
			if err != nil {
				logger.Error(err, fmt.Sprintf("addOrUpdateSwitchPort failed, interface name: %+v \n", interfaces[j].GetName()))
				continue
			}
		}
	}
	return nil
}

func (d *NetboxController) addOrUpdateSwitchPort(ctx context.Context, providerServerInterface nc.Interface, switchMap map[int32]*nc.Device) error {
	logger := log.FromContext(ctx).WithName("NetboxController.addOrUpdateSwitchPort")

	// note: will there be multiple link peers??
	for i := range providerServerInterface.LinkPeers {
		linkPeer := providerServerInterface.LinkPeers[i]

		// get the switch FQDN
		if linkPeer.Device == nil || len(linkPeer.Device.Name) == 0 {
			logger.Error(fmt.Errorf("linkPeer.Device.Name is empty"), "")
			continue
		}

		linkPeerDevice := switchMap[linkPeer.Device.ID]
		err, swFQDN := netboxSwNameToFqdn(d.cfg.ControllerConfig.NetboxSwitchFQDNDomainName, d.cfg.ControllerConfig.DataCenter, linkPeer.Device.Name, linkPeerDevice)
		if err != nil {
			logger.Error(err, "Failed to convert switch name to FQDN", "switchName", linkPeer.Device.Name)
			continue
		}

		err = utils.ValidateSwitchFQDN(swFQDN, d.cfg.ControllerConfig.DataCenter)
		if err != nil {
			continue
		}

		// get the switch port name
		if len(linkPeer.Name) == 0 {
			logger.Error(fmt.Errorf("linkPeer.Name is empty, provider server: %v, provider server interface: %v", providerServerInterface.Device, providerServerInterface.Name), "")
			continue
		}

		spName := linkPeer.Name
		err = utils.ValidatePortValue(spName)
		if err != nil {
			logger.Error(fmt.Errorf("ValidatePortValue failed: %v, provider server: %v, provider server interface: %v", err, providerServerInterface.Device, spName), "")
			continue
		}

		spCRName := utils.GeneratePortFullName(swFQDN, spName)

		// return error if switch CR doesn't exist
		switchCR := &idcnetworkv1alpha1.Switch{}
		swkey := types.NamespacedName{Name: swFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = d.networkK8sClient.Get(ctx, swkey, switchCR)
		if err != nil {
			return fmt.Errorf("skip creating the SwitchPort CR %v, as we cannot find the Switch CR %v", spCRName, swFQDN)
		}

		// check if we already have this SwitchPort CR created
		frontEndSwitchPortCR := &idcnetworkv1alpha1.SwitchPort{}
		key := types.NamespacedName{Name: spCRName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		objectExists := true
		err = d.networkK8sClient.Get(ctx, key, frontEndSwitchPortCR)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			} else {
				// Object not found
				objectExists = false
			}
		}

		// SP doesn't exist, create it.
		if !objectExists {
			//
			labels := make(map[string]string, 0)
			// note: we don't actually create the NN, just add this label to better identify which provider server it belongs to.
			labels[idcnetworkv1alpha1.LabelNameNetworkNode] = providerServerInterface.Device.GetName()
			switchPortCR := utils.NewSwitchPortTemplate(swFQDN, spName, idcnetworkv1alpha1.NOOPVlanID, idcnetworkv1alpha1.NOOPPortChannel, labels)
			err = d.networkK8sClient.Create(ctx, switchPortCR, &client.CreateOptions{})
			if err != nil {
				return fmt.Errorf("create switchPort CR %v failed, reason: %v", switchPortCR.Name, err)
			}
		}
	}

	logger.V(1).Info("addOrUpdateSwitchPort completed")
	return nil
}

const (
	ManagementInterfaceName  = "Management 1"
	ManagementInterfaceName2 = "Management1"
)

func (d *NetboxController) getSwitchManagementIP(ctx context.Context, switchName string) (string, error) {

	filter := &nc.InterfacesFilter{
		InterfaceName: []string{ManagementInterfaceName, ManagementInterfaceName2},
		Device:        []*string{&switchName},
	}
	intListReq := nc.ListInterfaceRequest{
		Filter: filter,
	}

	interfaces, err := d.netboxClient.ListInterfaces(ctx, intListReq)
	if err != nil {
		return "", fmt.Errorf("ListInterfaces failed, %v", err)
	}
	// there should be only one management interface
	if len(interfaces) < 1 {
		return "", fmt.Errorf("no management interface found")
	}
	mgmt := interfaces[0]

	ipRequest := nc.ListIPAddressesRequest{
		Filter: &nc.IPAddressesFilter{
			InterfacesId: []int32{mgmt.GetId()},
		},
	}

	ips, err := d.netboxClient.ListIPAddresses(ctx, ipRequest)
	if err != nil {
		return "", fmt.Errorf("ListIPAddresses failed, %v", err)
	}
	if len(ips) < 1 {
		return "", fmt.Errorf("no ip address found for interface [%v]", ipRequest.Filter.InterfacesId)
	}

	return utils.ConvertIPFormat(ips[0].Address), nil
}
