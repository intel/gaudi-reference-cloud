// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package switchclients

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"go.opentelemetry.io/otel/codes"

	"github.com/aristanetworks/goeapi"
	"github.com/aristanetworks/goeapi/module"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/strings/slices"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var eapiGetSwitchPortsCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiGetSwitchPortscounter",
		Help:        "Total attempts to get switch ports.",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiGetSwitchPortsInterfacesCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiGetSwitchPortsInterfacescounter",
		Help:        "Total attempts to get switch ports interfaces.",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiUpdateVlanCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiUpdateVlancounter",
		Help:        "Total attempts to update vlan",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiUpdateModeCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiUpdateModecounter",
		Help:        "Total attempts to update mode",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiUpdateDescriptionCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiUpdateDescriptioncounter",
		Help:        "Total attempts to update description",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiUpdateNativeVlanCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiUpdateNativeVlancounter",
		Help:        "Total attempts to update native vlan",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiUpdateBGPCommunityCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiUpdateBGPCommunitycounter",
		Help:        "Total attempts to update BGP community",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiGetBGPCommunityCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiGetBGPCommunitycounter",
		Help:        "Total attempts to get BGP community",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiUpdateTrunkGroupsCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiUpdateTrunkGroupscounter",
		Help:        "Total attempts to update trunk groups",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

var eapiGetVlansCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:        "eapiGetVlanscounter",
		Help:        "Total attempts to get vlans.",
		ConstLabels: map[string]string{"application": "sdn-controller"},
	},
)

type AristaClient struct {
	Name string

	// config
	host               string
	switchSecretsPath  string
	port               int
	transport          string
	connectionTimeout  time.Duration
	allowedTrunkGroups []string

	// eapi
	Node *goeapi.Node
	Sys  *module.SystemEntity

	ReadOnly             bool // If set to true, will only do "read" type requests.
	AllowedModes         []string
	AllowedVlanIds       []int
	AllowedNativeVlanIds []int
	ProvisioningVlanIds  []int
}

type ConnResult struct {
	Node *goeapi.Node
	Err  error
}

func init() {
	metrics.Registry.MustRegister(eapiGetSwitchPortsCounter)
	metrics.Registry.MustRegister(eapiGetSwitchPortsInterfacesCounter)
	metrics.Registry.MustRegister(eapiGetVlansCounter)
	metrics.Registry.MustRegister(eapiUpdateVlanCounter)
	metrics.Registry.MustRegister(eapiUpdateModeCounter)
	metrics.Registry.MustRegister(eapiUpdateDescriptionCounter)
	metrics.Registry.MustRegister(eapiUpdateNativeVlanCounter)
	metrics.Registry.MustRegister(eapiUpdateBGPCommunityCounter)
	metrics.Registry.MustRegister(eapiGetBGPCommunityCounter)
	metrics.Registry.MustRegister(eapiUpdateTrunkGroupsCounter)
}

func NewAristaClient(host string, switchSecretsPath string, port int, transport string, connectionTimeout time.Duration, readOnly bool, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, allowedTrunkGroups []string, provisioningVlanIds []int) (*AristaClient, error) {
	_, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("AristaClient.NewAristaClient").Start()
	defer span.End()
	secretFile, err := ioutil.ReadFile(switchSecretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read eAPISecretsPath file [%s] err: %v", switchSecretsPath, err)
	}
	eAPISecret := &idcnetworkv1alpha1.EAPISecret{}
	err = yaml.Unmarshal(secretFile, &eAPISecret)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal eAPISecretsPath file [%s] err: %v", switchSecretsPath, err)
	}

	var node *goeapi.Node
	resCh := make(chan *ConnResult)
	go func() {
		node, err = goeapi.Connect(transport, host, eAPISecret.Credentials.Username, eAPISecret.Credentials.Password, port)
		// Note: goeapi.Connect() wouldn't return error for the cases like "401 Unauthorized".
		// We perform a simple validation to make sure if the connection is actually working.
		err = validateEAPIConnection(node)

		resCh <- &ConnResult{
			Node: node,
			Err:  err,
		}
	}()
	select {
	case <-time.After(connectionTimeout):
		err = fmt.Errorf("eapi connection timeout")
	case res, ok := <-resCh:
		if !ok {
			return nil, fmt.Errorf("result channel is closed")
		}
		node = res.Node
		err = res.Err
	}
	if err != nil {
		return nil, fmt.Errorf("error connecting to switch: %v", err)
	}
	sys := module.System(node)

	client := &AristaClient{
		host:               host,
		switchSecretsPath:  switchSecretsPath,
		port:               port,
		transport:          transport,
		connectionTimeout:  connectionTimeout,
		allowedTrunkGroups: allowedTrunkGroups,

		Node: node,
		Sys:  sys,

		ReadOnly:             readOnly,
		AllowedVlanIds:       allowedVlanIds,
		AllowedNativeVlanIds: allowedNativeVlanIds,
		AllowedModes:         allowedModes,
		ProvisioningVlanIds:  provisioningVlanIds,
	}

	return client, nil
}

func (a *AristaClient) GetHost() (string, error) {
	return a.host, nil
}

func (a *AristaClient) RefreshConnection() error {
	_, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("AristaClient.RefreshConnection").WithValues(utils.LogFieldSwitchFQDN, a.host).Start()
	defer span.End()

	secretFile, err := ioutil.ReadFile(a.switchSecretsPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s err: %v", a.switchSecretsPath, err)
	}
	eAPISecret := &idcnetworkv1alpha1.EAPISecret{}
	err = yaml.Unmarshal(secretFile, &eAPISecret)
	if err != nil {
		return fmt.Errorf("failed unmarshal file %s err: %v", a.switchSecretsPath, err)
	}

	var node *goeapi.Node
	resCh := make(chan *ConnResult)
	go func() {
		node, err = goeapi.Connect(a.transport, a.host, eAPISecret.Credentials.Username, eAPISecret.Credentials.Password, a.port)
		// Note: goeapi.Connect() wouldn't return error for the cases like "401 Unauthorized".
		// We perform a simple validation to make sure if the connection is actually working.
		err = validateEAPIConnection(node)

		resCh <- &ConnResult{
			Node: node,
			Err:  err,
		}
	}()
	select {
	case <-time.After(a.connectionTimeout):
		err = fmt.Errorf("eapi connection timeout")
	case res, ok := <-resCh:
		if !ok {
			return fmt.Errorf("result channel is closed")
		}
		node = res.Node
		err = res.Err
	}
	if err != nil {
		return fmt.Errorf("error connecting to switch: %v", err)
	}

	a.Node = node
	a.Sys = module.System(node)
	return nil
}

func validateEAPIConnection(node *goeapi.Node) error {
	resp, err := node.RunCommands([]string{"show version"}, "json")
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("%s", resp.Error.Message)
	}
	return nil
}

// AristaPort
type AristaPort struct {
	Bandwidth           int             `json:"Bandwidth,omitempty"`
	InterfaceType       string          `json:"InterfaceType,omitempty"`
	Description         string          `json:"Description,omitempty"`
	AutoNegotiateActive bool            `json:"AutoNegotiateActive,omitempty"`
	Duplex              string          `json:"Duplex,omitempty"`
	LinkStatus          string          `json:"LinkStatus,omitempty"`
	LineProtocolStatus  string          `json:"LineProtocolStatus,omitempty"`
	VlanInformation     VlanInformation `json:"VlanInformation,omitempty"`
}

type VlanInformation struct {
	InterfaceMode            string `json:"InterfaceMode,omitempty"`
	VlanID                   int    `json:"VlanID,omitempty"`
	InterfaceForwardingModel string `json:"InterfaceForwardingModel,omitempty"`
	VlanExplanation          string `json:"vlanExplanation,omitempty"`
}

type ShowVlan struct {
	SourceDetail string          `json:"sourceDetail"`
	Vlans        map[string]Vlan `json:"vlans"`
}

func (s *ShowVlan) GetCmd() string {
	return "show vlan"
}

type TrunkGroupNames struct {
	Names []string `json:"names"`
}

type ShowVlanTrunkGroup struct {
	VlanTrunkGroups map[string]TrunkGroupNames `json:"trunkGroups"`
}

func (s *ShowVlanTrunkGroup) GetCmd() string {
	return "show vlan trunk group"
}

type Vlan struct {
	Status     string               `json:"status"`
	Name       string               `json:"name"`
	Interfaces map[string]Interface `json:"interfaces"`
	Dynamic    bool                 `json:"dynamic"`
}

type VlanWithTrunkGroups struct {
	VlanId         int      `json:"vlanId"`
	Status         string   `json:"status"`
	Name           string   `json:"name"`
	InterfaceNames []string `json:"interfaceNames"`
	Dynamic        bool     `json:"dynamic"`
	TrunkGroups    []string `json:"trunkGroups"`
}

type Interface struct {
	Annotation      string `json:"annotation"`
	PrivatePromoted bool   `json:"privatePromoted"`
}

type PortVlans struct {
	TaggedVlans  []int `json:"taggedVlans"`
	UntaggedVlan int   `json:"untaggedVlan"`
}

type InterfaceSwitchPort struct {
	Enabled        bool           `json:"enabled"`
	SwitchportInfo SwitchportInfo `json:"switchportInfo"`
}

type SwitchportInfo struct {
	Mode                       string `json:"mode"`
	PhoneTrunk                 bool   `json:"phoneTrunk"`
	MacLearning                bool   `json:"macLearning"`
	Tpid                       string `json:"tpid"`
	TpidStatus                 bool   `json:"tpidStatus"`
	Dot1QVlanTagRequired       bool   `json:"dot1qVlanTagRequired"`
	Dot1QVlanTagRequiredStatus bool   `json:"dot1qVlanTagRequiredStatus"`
	Dot1QVlanTagDisallowed     bool   `json:"dot1qVlanTagDisallowed"`
	AccessVlanId               int    `json:"accessVlanId"`
	AccessVlanName             string `json:"accessVlanName"`
	TrunkingNativeVlanId       int    `json:"trunkingNativeVlanId"`
	TrunkingNativeVlanName     string `json:"trunkingNativeVlanName"`
	TrunkAllowedVlans          string `json:"trunkAllowedVlans"`
	DynamicAllowedVlans        struct {
	} `json:"dynamicAllowedVlans"`
	DynamicBlockedVlans struct {
	} `json:"dynamicBlockedVlans"`
	StaticTrunkGroups    []string      `json:"staticTrunkGroups"`
	DynamicTrunkGroups   []interface{} `json:"dynamicTrunkGroups"`
	SourceportFilterMode string        `json:"sourceportFilterMode"`
	VlanForwardingMode   string        `json:"vlanForwardingMode"`
	PhoneVlan            int           `json:"phoneVlan"`
	PhoneTrunkUntagged   bool          `json:"phoneTrunkUntagged"`
	MbvaEnabled          bool          `json:"mbvaEnabled"`
}

type AristaCommunityList struct {
	Entries []AristaCommunityEntry `json:"entries,omitempty"`
}
type AristaCommunityEntry struct {
	FilterType      string `json:"FilterType,omitempty"`
	ListType        string `json:"ListType,omitempty"`
	CommunityValues []string
}

type InterfaceDetail struct {
	Name                      string  `json:"name"`
	LastStatusChangeTimestamp float64 `json:"lastStatusChangeTimestamp"`
}

type showInterfacesResponse struct {
	InterfaceName     string                     `json:"interfaceName"`
	InterfaceResponse map[string]InterfaceDetail `json:"interfaces"`
}

func (s *showInterfacesResponse) GetCmd() string {
	return "show interfaces " + s.InterfaceName
}

type showInterfacesStatus struct {
	InterfaceStatuses map[string]AristaPort
}

func (s *showInterfacesStatus) GetCmd() string {
	return "show interfaces status"
}

type showInterfacesVlans struct {
	Output     string // Only populated if we use "text" encoding (required for old EOS versions compatibility)
	Interfaces map[string]PortVlans
}

func (s *showInterfacesVlans) GetCmd() string {
	return "show interfaces vlans"
}

type showInterfacesSwitchport struct {
	Switchports map[string]InterfaceSwitchPort
}

func (s *showInterfacesSwitchport) GetCmd() string {
	return "show interfaces switchport"
}

type enableCmd map[string]interface{}

func (s *enableCmd) GetCmd() string {
	return "enable"
}

type configCmd map[string]interface{}

func (s *configCmd) GetCmd() string {
	return "configure"
}

type lldpDisableCmd map[string]interface{}

func (s *lldpDisableCmd) GetCmd() string {
	return "no lldp transmit"
}

type lldpEnableCmd map[string]interface{}

func (s *lldpEnableCmd) GetCmd() string {
	return "lldp transmit"
}

type selectInterfaceCmd struct {
	interfaceName string
}

func (s *selectInterfaceCmd) GetCmd() string {
	return fmt.Sprintf("interface %s", s.interfaceName)
}

type setDescriptionCmd struct {
	description string
}

func (s *setDescriptionCmd) GetCmd() string {
	return fmt.Sprintf("description %s", s.description)
}

type switchportAccessVlanCmd struct {
	vlanId int
}

func (s *switchportAccessVlanCmd) GetCmd() string {
	return fmt.Sprintf("switchport access vlan %d", s.vlanId)
}

type noSwitchportAccessVlanCmd map[string]interface{}

func (s *noSwitchportAccessVlanCmd) GetCmd() string {
	return fmt.Sprintf("no switchport access vlan")
}

type spanningTreePortfastCmd map[string]interface{}

func (s *spanningTreePortfastCmd) GetCmd() string {
	return fmt.Sprintf("spanning-tree portfast")
}

type spanningTreeBpduGuardCmd map[string]interface{}

func (s *spanningTreeBpduGuardCmd) GetCmd() string {
	return fmt.Sprintf("spanning-tree bpduguard enable")
}

type modeAccessCmd map[string]interface{}

func (s *modeAccessCmd) GetCmd() string {
	return "switchport mode access"
}

type modeTrunkCmd map[string]interface{}

func (s *modeTrunkCmd) GetCmd() string {
	return "switchport mode trunk"
}

type addTrunkGroupCmd struct {
	trunkGroup string
}

func (s *addTrunkGroupCmd) GetCmd() string {
	return fmt.Sprintf("switchport trunk group %s", s.trunkGroup)
}

type noTrunkGroupCmd map[string]interface{}

func (s *noTrunkGroupCmd) GetCmd() string {
	return fmt.Sprintf("no switchport trunk group")
}

type trunkNativeVlanCmd struct {
	nativeVlan int
}

func (s *trunkNativeVlanCmd) GetCmd() string {
	return fmt.Sprintf("switchport trunk native vlan %d", s.nativeVlan)
}

type noTrunkNativeVlanCmd map[string]interface{}

func (s *noTrunkNativeVlanCmd) GetCmd() string {
	return fmt.Sprintf("no switchport trunk native vlan")
}

type selectAdvertiseRouteMapCmd struct {
}

func (s *selectAdvertiseRouteMapCmd) GetCmd() string {
	return fmt.Sprintf("route-map adv-set-comm permit 10")
}

type removeAdvertiseCommunityCmd struct {
}

func (s *removeAdvertiseCommunityCmd) GetCmd() string {
	return "no set community"
}

type addAdvertiseCommunityCmd struct {
	BGPCommunity int
}

func (s *addAdvertiseCommunityCmd) GetCmd() string {
	return fmt.Sprintf("set community 101:%d", s.BGPCommunity)
}

type exitCmd struct {
}

func (s *exitCmd) GetCmd() string {
	return "exit"
}

type removeIncomingCommunityCmd struct {
	BGPCommunityGroupName string
}

func (s *removeIncomingCommunityCmd) GetCmd() string {
	return fmt.Sprintf("no ip community-list %s", s.BGPCommunityGroupName)
}

type addIncomingCommunityCmd struct {
	BGPCommunity          int
	BGPCommunityGroupName string
}

func (s *addIncomingCommunityCmd) GetCmd() string {
	return fmt.Sprintf("ip community-list %s permit 101:%d", s.BGPCommunityGroupName, s.BGPCommunity)
}

type showBGPCommunityStatus struct {
	BGPCommunityGroupName string
	IpCommunityLists      map[string]AristaCommunityList
}

func (s *showBGPCommunityStatus) GetCmd() string {
	return fmt.Sprintf("show ip community-list %s", s.BGPCommunityGroupName)
}

type showRunningConfig struct {
	Output string
}

func (s *showRunningConfig) GetCmd() string {
	return "show running-config"
}

type showStartupConfig struct {
	Output string
}

func (s *showStartupConfig) GetCmd() string {
	return "show startup-config"
}

type copyRunningConfigAsStartupConfig struct {
	Output string
}

func (s *copyRunningConfigAsStartupConfig) GetCmd() string {
	return "copy running-config startup-config"
}

type reload struct {
	Output string
}

func (s *reload) GetCmd() string {
	return "reload"
}

type restoreRunningConfigFromStartupConfig struct {
	Output string
}

func (s *restoreRunningConfigFromStartupConfig) GetCmd() string {
	return "configure replace startup-config"
}

// example: "copy running-config my-config-backup"
type saveRunningConfigWithName struct {
	ConfigName string
	Output     string
}

func (s *saveRunningConfigWithName) GetCmd() string {
	return fmt.Sprintf("copy running-config %s", s.ConfigName)
}

type setStartupConfigWithName struct {
	ConfigName string
	Output     string
}

// example: "copy my-config-backup startup-config"
func (s *setStartupConfigWithName) GetCmd() string {
	return fmt.Sprintf("copy %s startup-config", s.ConfigName)
}

func (a *AristaClient) UpdateMode(ctx context.Context, req UpdateModeRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdateMode").WithValues(utils.LogFieldMode, req.Mode)
	startTime := time.Now().UTC()

	//enable
	//configure
	//interface ethernet27/1

	err := utils.ValidatePortValue(req.PortName)
	if err != nil {
		return fmt.Errorf("ValidatePortValue failed, error: %v", err)
	}

	err = utils.ValidateModeValue(req.Mode, a.AllowedModes)
	if err != nil {
		return fmt.Errorf("ValidateModeValue failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.PortName,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}

	if req.Mode == "access" {
		// switchport mode access
		// no switchport trunk group
		// no switchport trunk native vlan
		modeAccessCmdRsp := &modeAccessCmd{}
		err = handle.AddCommand(modeAccessCmdRsp)
		if err != nil {
			return fmt.Errorf("modeAccessCmd command failed, %v", err)
		}

		// noTrunkGroupCmdRsp := &noTrunkGroupCmd{}
		// handle.AddCommand(noTrunkGroupCmdRsp)
		// if err != nil {
		// 	return fmt.Errorf("noTrunkGroupCmd command failed, %v", err)
		// }

		// noTrunkNativeVlanCmdRsp := &noTrunkNativeVlanCmd{}
		// handle.AddCommand(noTrunkNativeVlanCmdRsp)
		// if err != nil {
		// 	return fmt.Errorf("noTrunkNativeVlanCmd command failed, %v", err)
		// }

	} else if req.Mode == "trunk" {
		// switchport mode trunk
		// no switchport access vlan
		modeTrunkCmdRsp := &modeTrunkCmd{}
		err = handle.AddCommand(modeTrunkCmdRsp)
		if err != nil {
			return fmt.Errorf("modeTrunkCmdRsp command failed, %v", err)
		}

		noSwitchportAccessVlanCmdRsp := &noSwitchportAccessVlanCmd{}
		err = handle.AddCommand(noSwitchportAccessVlanCmdRsp)
		if err != nil {
			return fmt.Errorf("noSwitchportAccessVlanCmd command failed, %v", err)
		}
	}

	if !a.ReadOnly {
		eapiUpdateModeCounter.Add(1)
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		timeElapsed := time.Since(startTime)
		logger.V(1).Info("AristaClient.UpdateMode success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}

func (a *AristaClient) UpdateVlan(ctx context.Context, req UpdateVlanRequest) error {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AristaClient.UpdateVlan").WithValues(utils.LogFieldSwitchFQDN, a.host, utils.LogFieldVlanID, req.Vlan).Start()
	defer span.End()

	startTime := time.Now().UTC()

	//enable
	//configure
	//interface ethernet27/1
	//switchport access vlan 21
	//no lldp transmit or lldp transmit

	err := utils.ValidatePortValue(req.PortName)
	if err != nil {
		errfmt := fmt.Errorf("ValidatePortValue failed, error: %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	err = utils.ValidateVlanValue(int(req.Vlan), a.AllowedVlanIds)
	if err != nil {
		errfmt := fmt.Errorf("ValidateVlanValue failed, error: %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	entries, err := a.ListVlans(ctx, ListVlansParamsRequest{SwitchFQDN: a.host})
	if err != nil {
		errfmt := fmt.Errorf("ListVlans failed, error: %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	found := false
	for _, vlan := range entries {
		if vlan.VlanId == int(req.Vlan) {
			found = true
			break
		}
	}
	if !found {
		errfmt := fmt.Errorf("Requested Vlan entry not found on the switch\n")
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		errfmt := fmt.Errorf("node.GetHandle failed, error: %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		errfmt := fmt.Errorf("enableCmdRsp command failed, %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		errfmt := fmt.Errorf("configCmdRsp command failed, %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.PortName,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		errfmt := fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	switchportAccessVlanCmdRsp := &switchportAccessVlanCmd{
		vlanId: int(req.Vlan),
	}
	err = handle.AddCommand(switchportAccessVlanCmdRsp)
	if err != nil {
		errfmt := fmt.Errorf("switchportAccessVlanCmdRsp command failed, %v", err)
		span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	if req.UpdateLLDP == true {
		provisioningVlan := false
		for _, vlan := range a.ProvisioningVlanIds {
			if int(req.Vlan) == vlan {
				provisioningVlan = true
				break
			}
		}
		if !provisioningVlan {
			lldpDisableCmdRsp := &lldpDisableCmd{}
			err = handle.AddCommand(lldpDisableCmdRsp)
			if err != nil {
				errfmt := fmt.Errorf("lldpDisableCmdRsp command failed, %v", err)
				span.SetStatus(codes.Error, errfmt.Error())
				return errfmt
			}
		} else {
			lldpEnableCmdRsp := &lldpEnableCmd{}
			err = handle.AddCommand(lldpEnableCmdRsp)
			if err != nil {
				errfmt := fmt.Errorf("lldpEnableCmdRsp command failed, %v", err)
				span.SetStatus(codes.Error, errfmt.Error())
				return errfmt
			}
		}
	}

	if !a.ReadOnly {
		eapiUpdateVlanCounter.Add(1)
		err = handle.Call()
		if err != nil {
			errfmt := fmt.Errorf("handle.Call failed, %v", err)
			span.SetStatus(codes.Error, errfmt.Error())
			return errfmt
		}
		timeElapsed := time.Since(startTime)
		logger.Info("AristaClient.UpdateVlan success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}

func (a *AristaClient) UpdateDescription(ctx context.Context, req UpdateDescriptionRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdateDescription").WithValues(utils.LogFieldDescription, req.Description)
	startTime := time.Now().UTC()

	err := utils.ValidatePortValue(req.PortName)
	if err != nil {
		return fmt.Errorf("ValidatePortValue failed, error: %v", err)
	}

	sanitizedDescription, err := utils.ValidateAndSanitizeDescription(req.Description)
	if err != nil {
		return fmt.Errorf("ValidateAndSanitizeDescription failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.PortName,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}
	setDescriptionCmdRsp := &setDescriptionCmd{
		description: sanitizedDescription,
	}
	err = handle.AddCommand(setDescriptionCmdRsp)
	if err != nil {
		errfmt := fmt.Errorf("setDescriptionCmdRsp command failed, %v", err)
		//span.SetStatus(codes.Error, errfmt.Error())
		return errfmt
	}

	if !a.ReadOnly {
		eapiUpdateDescriptionCounter.Add(1)
		err = handle.Call()
		if err != nil {
			errfmt := fmt.Errorf("handle.Call failed, %v", err)
			return errfmt
		}
		timeElapsed := time.Since(startTime)
		logger.Info("AristaClient.UpdateDescription success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}

func (a *AristaClient) UpdateTrunkGroups(ctx context.Context, req UpdateTrunkGroupsRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdateTrunkGroups").WithValues(utils.LogFieldTrunkGroups, req.TrunkGroups)
	startTime := time.Now().UTC()

	// enable
	// configure
	// interface ethernet27/1
	// switchport trunk group Provider_Nets
	// switchport trunk group Tenant_Nets

	err := utils.ValidatePortValue(req.PortName)
	if err != nil {
		return fmt.Errorf("validate PortName failed, error: %v", err)
	}
	err = utils.ValidateTrunkGroups(req.TrunkGroups, a.allowedTrunkGroups)
	if err != nil {
		return fmt.Errorf("ValidateTrunkGroups failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmdRsp command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.PortName,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}

	// Remove all trunk-groups, then add back only the specified ones.
	noTrunkGroupCmdRsp := &noTrunkGroupCmd{}
	err = handle.AddCommand(noTrunkGroupCmdRsp)
	if err != nil {
		return fmt.Errorf("noTrunkGroupCmdRsp command failed, %v", err)
	}

	for _, trunkGroup := range req.TrunkGroups {
		trunkNativeVlanCmdRsp := &addTrunkGroupCmd{
			trunkGroup: trunkGroup,
		}
		err = handle.AddCommand(trunkNativeVlanCmdRsp)
		if err != nil {
			return fmt.Errorf("addTrunkGroupCmd command failed, %v", err)
		}
	}

	if !a.ReadOnly {
		eapiUpdateTrunkGroupsCounter.Add(1)
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		timeElapsed := time.Since(startTime)
		logger.Info("AristaClient.UpdateTrunkGroups success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}

func (a *AristaClient) UpdateNativeVlan(ctx context.Context, req UpdateNativeVlanRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdateNativeVlan").WithValues(utils.LogFieldNativeVlan, req.NativeVlan)
	startTime := time.Now().UTC()

	// enable
	// configure
	// interface ethernet27/1
	// switchport trunk native vlan 100

	err := utils.ValidatePortValue(req.PortName)
	if err != nil {
		return fmt.Errorf("validate PortName failed, error: %v", err)
	}
	err = utils.ValidateVlanValue(int(req.NativeVlan), a.AllowedNativeVlanIds)
	if err != nil {
		return fmt.Errorf("ValidateVlanValue NativeVlan failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmdRsp command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.PortName,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}

	trunkNativeVlanCmdRsp := &trunkNativeVlanCmd{
		nativeVlan: int(req.NativeVlan),
	}
	err = handle.AddCommand(trunkNativeVlanCmdRsp)
	if err != nil {
		return fmt.Errorf("trunkNativeVlanCmd command failed, %v", err)
	}

	if !a.ReadOnly {
		eapiUpdateNativeVlanCounter.Add(1)
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		timeElapsed := time.Since(startTime)
		logger.Info("AristaClient.UpdateNativeVlan success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}

func (a *AristaClient) UpdateBGPCommunity(ctx context.Context, req UpdateBGPCommunityRequest) error {
	_, logger, bgpSpan := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AristaClient.UpdateBGPCommunity").WithValues(utils.LogFieldSwitchFQDN, a.host, utils.LogFieldBGPCommunity, req.BGPCommunity).Start()
	defer bgpSpan.End()

	err := utils.ValidateBGPCommunityValue(req.BGPCommunity)
	if err != nil {
		return fmt.Errorf("ValidateBGPCommunityValue failed, error: %v", err)
	}
	err = utils.ValidateBGPCommunityGroupName(req.BGPCommunityIncomingGroupName)
	if err != nil {
		return fmt.Errorf("ValidateBGPCommunityGroupName failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmd command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	selectAdvertiseRouteMapCmdRsp := &selectAdvertiseRouteMapCmd{}
	err = handle.AddCommand(selectAdvertiseRouteMapCmdRsp)
	if err != nil {
		return fmt.Errorf("selectAdvertiseRouteMapCmd command failed, %v", err)
	}

	removeAdvertiseRouteMapCmdRsp := &removeAdvertiseCommunityCmd{}
	err = handle.AddCommand(removeAdvertiseRouteMapCmdRsp)
	if err != nil {
		return fmt.Errorf("removeAdvertiseRouteMapCmd command failed, %v", err)
	}

	addAdvertiseCommunityCmdRsp := &addAdvertiseCommunityCmd{
		BGPCommunity: int(req.BGPCommunity),
	}
	err = handle.AddCommand(addAdvertiseCommunityCmdRsp)
	if err != nil {
		return fmt.Errorf("addAdvertiseCommunityCmd command failed, %v", err)
	}

	exitCmdRsp := &exitCmd{}
	err = handle.AddCommand(exitCmdRsp)
	if err != nil {
		return fmt.Errorf("exitCmdRsp command failed, %v", err)
	}

	removeIncomingCommunityCmdRsp := &removeIncomingCommunityCmd{
		BGPCommunityGroupName: req.BGPCommunityIncomingGroupName,
	}
	err = handle.AddCommand(removeIncomingCommunityCmdRsp)
	if err != nil {
		return fmt.Errorf("removeIncomingCommunityCmd command failed, %v", err)
	}

	addIncomingCommunityCmdRsp := &addIncomingCommunityCmd{
		BGPCommunityGroupName: req.BGPCommunityIncomingGroupName,
		BGPCommunity:          int(req.BGPCommunity),
	}
	err = handle.AddCommand(addIncomingCommunityCmdRsp)
	if err != nil {
		return fmt.Errorf("addIncomingCommunityCmd command failed, %v", err)
	}

	if !a.ReadOnly {
		eapiUpdateBGPCommunityCounter.Add(1)
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		logger.Info("AristaClient.UpdateBGPCommunity success!")
	}

	return nil
}

func (a *AristaClient) GetBGPCommunity(ctx context.Context, req GetBGPCommunityRequest) (int, error) {
	_, _, bgpSpan := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AristaClient.GetBGPCommunity").WithValues(utils.LogFieldSwitchFQDN, a.host).Start()
	defer bgpSpan.End()

	var incomingGroupName = req.BGPCommunityIncomingGroupName
	err := utils.ValidateBGPCommunityGroupName(req.BGPCommunityIncomingGroupName)
	if err != nil {
		return 0, fmt.Errorf("ValidateBGPCommunityGroupName failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return 0, err
	}

	showIntBGPCommunityRsp := &showBGPCommunityStatus{
		BGPCommunityGroupName: incomingGroupName,
	}
	err = handle.AddCommand(showIntBGPCommunityRsp)
	if err != nil {
		return 0, fmt.Errorf("eapi showBGPCommunityStatus failed, %v", err)
	}

	eapiGetBGPCommunityCounter.Add(1)
	err = handle.Call()
	if err != nil {
		return 0, fmt.Errorf("eapi handle.Call failed, %v", err)
	}

	if showIntBGPCommunityRsp.IpCommunityLists == nil {
		return 0, fmt.Errorf("showIntBGPCommunityRsp.IpCommunityLists was nil")
	}

	incomingGroup, ok := showIntBGPCommunityRsp.IpCommunityLists[incomingGroupName]
	if !ok {
		return 0, fmt.Errorf("did not find incoming group %s in response from switch", incomingGroupName)
	}

	if incomingGroup.Entries == nil {
		return 0, fmt.Errorf("incoming group %s.Entries was nil", incomingGroupName)
	}

	if len(incomingGroup.Entries) != 1 {
		return 0, fmt.Errorf("incoming group %s on switch did not have exactly one entry", incomingGroupName)
	}

	firstEntry := incomingGroup.Entries[0]
	if len(firstEntry.CommunityValues) != 1 {
		return 0, fmt.Errorf("incoming group %s did not have exactly one communityValue", incomingGroupName)
	}

	firstCommunityValueString := firstEntry.CommunityValues[0]
	communityValue, err := utils.BGPCommunityStringToValue(firstCommunityValueString)

	if err != nil {
		return 0, fmt.Errorf("could not parse communityValue: %v", err)
	}

	return communityValue, nil

}

func (a *AristaClient) GetSwitchPorts(ctx context.Context, req GetSwitchPortsRequest) (map[string]*idcnetworkv1alpha1.SwitchPortStatus, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AristaClient.GetSwitchPorts").WithValues(utils.LogFieldSwitchFQDN, a.host).Start()
	defer span.End()

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return nil, err
	}

	shIntRsp := &showInterfacesStatus{}
	shIntVlansRsp := &showInterfacesVlans{}
	shIntSwitchportsRsp := &showInterfacesSwitchport{}
	err = handle.AddCommand(shIntRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfacesStatus failed, %v", err)
	}
	err = handle.AddCommand(shIntVlansRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfacesVlans failed, %v", err)
	}
	err = handle.AddCommand(shIntSwitchportsRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfacesSwitchport failed, %v", err)
	}

	eapiGetSwitchPortsCounter.Add(1)
	err = handle.Call()
	if err != nil {
		if strings.Contains(err.Error(), "unconverted command") && strings.Contains(err.Error(), "show interfaces vlans") {
			// Fall back to getting via text, not json.
			return a.GetSwitchPortsTextVlans(ctx, req)
		} else {
			return nil, fmt.Errorf("eapi handle.Call failed, %v", err)
		}
	}

	// Generate initial switch ports data
	res := generateSwitchPorts(ctx, shIntRsp, shIntVlansRsp, shIntSwitchportsRsp)
	var ethernetInterfaces []string

	for interfaceName := range res {
		if strings.HasPrefix(interfaceName, "Ethernet") {
			err := utils.ValidatePortValue(interfaceName)
			if err != nil {
				logger.Info("port from switch not valid. Ignoring!'", utils.LogFieldSwitchPortName, interfaceName)
				continue
			} else {
				ethernetInterfaces = append(ethernetInterfaces, interfaceName)
			}
		}
	}

	interfaceRange, err := utils.GetInterfaceRange(ethernetInterfaces)
	if err != nil {
		return nil, fmt.Errorf("getInterfaceRange failed, %v", err)
	}

	shIntResponseRsp := &showInterfacesResponse{
		InterfaceName:     interfaceRange,
		InterfaceResponse: make(map[string]InterfaceDetail),
	}

	err = handle.AddCommand(shIntResponseRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfaces %s failed, %v", interfaceRange, err)
	}

	eapiGetSwitchPortsInterfacesCounter.Add(1)
	// Call the handle again to execute the new command
	err = handle.Call()
	if err != nil {
		return nil, fmt.Errorf("eapi handle.Call for individual interfaces failed, %v", err)
	}
	// update each interface status with SwitchSideLastStatusChangeTimestamp
	for _, interfaceName := range ethernetInterfaces {
		if detail, exists := shIntResponseRsp.InterfaceResponse[interfaceName]; exists {
			switchPortStatus := res[interfaceName]
			switchPortStatus.SwitchSideLastStatusChangeTimestamp = int64(detail.LastStatusChangeTimestamp)
		} else {
			return nil, fmt.Errorf("no detailed response for interface %s", interfaceName)
		}
	}
	return res, nil
}

func (a *AristaClient) GetSwitchPortsTextVlans(ctx context.Context, req GetSwitchPortsRequest) (map[string]*idcnetworkv1alpha1.SwitchPortStatus, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AristaClient.GetSwitchPortsTextEncoding").WithValues(utils.LogFieldSwitchFQDN, a.host).Start()
	defer span.End()

	logger.Info("Falling back to GetSwitchPorts using TextEncoding for vlans")
	node := a.Node

	// show interfaces status (using json, because that is supported)
	jsonHandle, err := node.GetHandle("json")
	if err != nil {
		return nil, err
	}
	shIntRsp := &showInterfacesStatus{}
	shIntSwitchportsRsp := &showInterfacesSwitchport{}
	err = jsonHandle.AddCommand(shIntRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfacesStatus failed, %v", err)
	}
	err = jsonHandle.AddCommand(shIntSwitchportsRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfacesSwitchport failed, %v", err)
	}
	err = jsonHandle.Call()
	if err != nil {
		return nil, fmt.Errorf("eapi jsonHandle.Call failed, %v", err)
	}

	// show vlans status using text format
	textHandle, err := node.GetHandle("text")
	if err != nil {
		return nil, err
	}
	shIntVlansRsp := &showInterfacesVlans{}
	err = textHandle.AddCommand(shIntVlansRsp)
	if err != nil {
		return nil, fmt.Errorf("eapi showInterfacesVlans failed, %v", err)
	}
	err = textHandle.Call()
	if err != nil {
		return nil, fmt.Errorf("eapi textHandle.Call failed, %v", err)
	}

	shIntVlansParsed, err := parseTextVlansToPortVlans(*shIntVlansRsp)
	if err != nil {
		return nil, fmt.Errorf("parseTextVlansToPortVlans failed, %v", err)
	}

	res := generateSwitchPorts(ctx, shIntRsp, shIntVlansParsed, shIntSwitchportsRsp)
	return res, nil
}

var spaceRegex = regexp.MustCompile(" +")

func parseTextVlansToPortVlans(textResponse showInterfacesVlans) (*showInterfacesVlans, error) {
	var ret = &showInterfacesVlans{
		Interfaces: make(map[string]PortVlans, 0),
	}

	interfaceName := ""
	lines := strings.Split(textResponse.Output, "\n")
	for linenum, line := range lines {
		if linenum == 0 { // Title row
			continue
		}
		if len(line) == 0 { // Final row
			continue
		}
		line = spaceRegex.ReplaceAllString(line, " ") // Collapse "       " to " "
		columns := strings.Split(line, " ")
		portVlans := &PortVlans{}
		for columnnum, cell := range columns {
			if columnnum == 0 && len(columns) == 3 {
				var err error
				interfaceName, err = utils.InterfaceShortToLongName(cell)
				if err != nil {
					return nil, fmt.Errorf("could not convert interface name \"%s\" to long name: %v", cell, err)
				}
			} else if columnnum == 1 && len(columns) == 3 {
				if cell == "None" {
					// Do nothing.
				} else {
					untaggedVlan, err := strconv.Atoi(cell)
					if err != nil {
						return nil, fmt.Errorf("could not convert untagged vlan \"%s\" to integer: %v", cell, err)
					}
					portVlans.UntaggedVlan = untaggedVlan
				}
			} else if (columnnum == 2 && len(columns) == 3) || (columnnum == 1 && len(columns) == 2) {
				var err error
				var taggedVlansInt []int
				if cell != "-" {
					taggedVlansInt, err = utils.ExpandVlanRanges(cell)
					if err != nil {
						return nil, fmt.Errorf("could not convert tagged vlan \"%s\" to vlan ranges: %v", cell, err)
					}
				}
				portVlans.TaggedVlans = append(ret.Interfaces[interfaceName].TaggedVlans, taggedVlansInt...)
			} else {
				//fmt.Printf("found columnnum %d with contents '%s' \n", columnnum, cell)
			}
		}
		ret.Interfaces[interfaceName] = *portVlans
	}

	return ret, nil
}

const (
	ModeBridged  = "bridged"
	ModeInactive = "inactive"
	ModeTrunk    = "trunk"
	ModeRouted   = "routed"
)

func generateSwitchPorts(ctx context.Context, intStatus *showInterfacesStatus, intVlans *showInterfacesVlans, intSwitchports *showInterfacesSwitchport) map[string]*idcnetworkv1alpha1.SwitchPortStatus {
	logger := log.FromContext(ctx).WithName("AristaClient.generateSwitchPorts")

	switchPorts := make(map[string]*idcnetworkv1alpha1.SwitchPortStatus)
	for interfaceName, interf := range intStatus.InterfaceStatuses {

		var portChannel int
		var err error
		re := regexp.MustCompile(`^in Po([0-9]+)$`)
		portChannelStrs := re.FindStringSubmatch(interf.VlanInformation.VlanExplanation)
		if len(portChannelStrs) > 1 {
			portChannel, err = strconv.Atoi(portChannelStrs[1])
			if err != nil {
				logger.Error(err, "Failed to parse integer portChannel from VlanExplanation. Got: '%s'", interf.VlanInformation.VlanExplanation)
			}
		}

		switchPort := &idcnetworkv1alpha1.SwitchPortStatus{
			Name:               interfaceName,
			VlanId:             int64(interf.VlanInformation.VlanID),
			Description:        interf.Description,
			LinkStatus:         interf.LinkStatus,
			Bandwidth:          interf.Bandwidth,
			InterfaceType:      interf.InterfaceType,
			LineProtocolStatus: interf.LineProtocolStatus,
			PortChannel:        int64(portChannel),
			Duplex:             interf.Duplex,
		}

		//interfaceVlans, found := intVlans.Interfaces[interfaceName]
		//if found {
		//	// Only makes sense to add NativeVlan field to trunked ports (For access ports it's just the vlanID)
		//	if switchPort.Mode == "trunk" {
		//		switchPort.NativeVlan = int64(interfaceVlans.UntaggedVlan)
		//	}
		//	switchPort.UntaggedVlan = int64(interfaceVlans.UntaggedVlan)
		//}

		interfaceSwitchport, found := intSwitchports.Switchports[interfaceName]
		if found {
			if len(interfaceSwitchport.SwitchportInfo.StaticTrunkGroups) >= 1 {
				sort.Strings(interfaceSwitchport.SwitchportInfo.StaticTrunkGroups)
				switchPort.TrunkGroups = interfaceSwitchport.SwitchportInfo.StaticTrunkGroups
			}
			switchPort.NativeVlan = int64(interfaceSwitchport.SwitchportInfo.TrunkingNativeVlanId)
			switchPort.Mode = interfaceSwitchport.SwitchportInfo.Mode
		} else { // not a switchport.
			if interf.VlanInformation.InterfaceMode == ModeRouted {
				switchPort.Mode = "routed"
			} else {
				// Do we ever get here? If it's not a switchport, surely it's routed? Log to make sure, if no logs then we can remove this else.
				logger.Info("port wasn't in 'show interfaces switchport' output, but also isn't 'routed'", utils.LogFieldMode, interf.VlanInformation.InterfaceMode, utils.LogFieldSwitchPortName, interfaceName)

				//"bridged" or "inactive" value indicates this port is an "access" port (?????)
				if interf.VlanInformation.InterfaceMode == ModeBridged || interf.VlanInformation.InterfaceMode == ModeInactive {
					switchPort.Mode = "access"
				} else if interf.VlanInformation.InterfaceMode == ModeTrunk {
					switchPort.Mode = "trunk"
				}
			}
		}

		switchPorts[interfaceName] = switchPort

	}
	return switchPorts
}

func (a *AristaClient) GetRunningConfig(ctx context.Context) (string, error) {
	node := a.Node
	handle, err := node.GetHandle("text")
	if err != nil {
		return "", err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return "", fmt.Errorf("enableCmd command failed, %v", err)
	}
	shRunRsp := &showRunningConfig{}
	err = handle.AddCommand(shRunRsp)
	if err != nil {
		return "", fmt.Errorf("showRunningConfig command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return "", fmt.Errorf("handle.Call failed, %v", err)
	}

	return shRunRsp.Output, nil
}

func (a *AristaClient) GetStartupConfig(ctx context.Context) (string, error) {
	node := a.Node
	handle, err := node.GetHandle("text")
	if err != nil {
		return "", err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return "", fmt.Errorf("enableCmd command failed, %v", err)
	}
	shStartupResp := &showStartupConfig{}
	err = handle.AddCommand(shStartupResp)
	if err != nil {
		return "", fmt.Errorf("showStartupConfig command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return "", fmt.Errorf("handle.Call failed, %v", err)
	}

	return shStartupResp.Output, nil
}

func (a *AristaClient) SaveRunningConfigAsStartupConfig(ctx context.Context) (string, error) {
	node := a.Node
	handle, err := node.GetHandle("text")
	if err != nil {
		return "", err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return "", fmt.Errorf("enableCmd command failed, %v", err)
	}
	saveStartupConfig := &copyRunningConfigAsStartupConfig{}
	err = handle.AddCommand(saveStartupConfig)
	if err != nil {
		return "", fmt.Errorf("saveStartupConfig command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return "", fmt.Errorf("AristaClient.SaveRunningConfigAsStartupConfig handle.Call failed, %v", err)
		}
	}

	return saveStartupConfig.Output, nil
}

func (a *AristaClient) Reload(ctx context.Context) error {
	node := a.Node
	handle, err := node.GetHandle("text")
	if err != nil {
		return err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmd command failed, %v", err)
	}
	reload := &reload{}
	err = handle.AddCommand(reload)
	if err != nil {
		return fmt.Errorf("reload command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
	}

	return nil
}

// RestoreRunningConfigFromStartupConfig copy the config from startup-config, mostly used for resetting in testing when Reload is not supported.
func (a *AristaClient) RestoreRunningConfigFromStartupConfig(ctx context.Context) (string, error) {
	node := a.Node
	handle, err := node.GetHandle("text")
	if err != nil {
		return "", err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return "", fmt.Errorf("enableCmd command failed, %v", err)
	}
	restoreRunningConfig := &restoreRunningConfigFromStartupConfig{}
	err = handle.AddCommand(restoreRunningConfig)
	if err != nil {
		return "", fmt.Errorf("restoreRunningConfig command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return "", fmt.Errorf("handle.Call failed, %v", err)
		}
	}

	return restoreRunningConfig.Output, nil
}

func (a *AristaClient) ValidateConnection() error {
	return validateEAPIConnection(a.Node)
}

type showMacAddressTable struct {
	MacUnicastTable *MacUnicastTable `json:"unicastTable,omitempty"`
}

type MacUnicastTable struct {
	MacTableEntries []*MacAddressTableEntry `json:"tableEntries,omitempty"`
}

func (s *showMacAddressTable) GetCmd() string {
	return "show mac address-table dynamic"
}

type MacAddressTableEntry struct {
	Interface  string `json:"interface,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
	VlanId     int    `json:"vlanId,omitempty"`
}

type ResMacAddressTableEntry struct {
	Interface  string `json:"interface"`
	MacAddress string `json:"mac_address"`
	VlanTag    int    `json:"vlanId"`
}

type MacUcastTableEntries struct {
	MacEntries map[string]interface{}
}

type showPortChannelDenseTable struct {
	NumberOfAggregators   int                      `json:"numberOfAggregators,omitempty"`
	NumberOfChannelsInUse int                      `json:"numberOfChannelsInUse,omitempty"`
	PortChannels          map[string]*PortChannels `json:"portChannels,omitempty"`
}

func (s *showPortChannelDenseTable) GetCmd() string {
	return "show port-channel dense"
}

type PortChannels struct {
	PortChannelEntries map[string]*PortChannelEntry `json:"ports,omitempty"`
}

type PortChannelEntry struct {
	Ports map[string]Ports
}

type Ports struct {
	Intf string `json:"intf,omitempty"`
}

func (a *AristaClient) GetMacAddressTable(ctx context.Context, req ListMacAddressTableRequest) ([]ResMacAddressTableEntry, error) {

	node := a.Node
	entries := []ResMacAddressTableEntry{}
	handle, err := node.GetHandle("json")
	if err != nil {
		return entries, err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("enableCmd command failed, %v", err)
	}

	shMacCmdRsp := &showMacAddressTable{}
	err = handle.AddCommand(shMacCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("showMacAddressTable command failed, %v", err)
	}

	shPOCmdRsp := &showPortChannelDenseTable{}
	err = handle.AddCommand(shPOCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("showPOAddressTable command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return entries, fmt.Errorf("handle.Call failed, %v", err)
	}

	// ------------------------------
	// show mac address-table dynamic
	if shMacCmdRsp.MacUnicastTable != nil && len(shMacCmdRsp.MacUnicastTable.MacTableEntries) > 0 {
		for _, item := range shMacCmdRsp.MacUnicastTable.MacTableEntries {
			// fmt.Printf("%+v \n", item)
			entries = append(entries, ResMacAddressTableEntry{Interface: item.Interface, MacAddress: item.MacAddress, VlanTag: item.VlanId})
		}
	}

	// ------------------------------
	// show port-channel dense
	portChannelMap := make(map[string][]string)
	if shPOCmdRsp.PortChannels != nil && len(shPOCmdRsp.PortChannels) > 0 {
		for name, entries := range shPOCmdRsp.PortChannels {
			for intf, _ := range entries.PortChannelEntries {
				if intf != "Peer" {
					portChannelMap[name] = append(portChannelMap[name], intf)
				}
			}
		}
	}

	// ------------------------------
	// exclude uplink ports based on the switch model (using "show version" to get switch model)
	resp, err := node.RunCommands([]string{"show version"}, "json")
	if err != nil {
		return entries, fmt.Errorf("showVersion command failed, %v", err)
	}

	excPorts := []string{"Vxlan1"}

	modelName := resp.Result[0]["modelName"]
	switch modelName {
	case "DCS-7050CX3-32":
		excPorts = append(excPorts, "Port-Channel33")
	case "DCS-7050SX3-48":
		excPorts = append(excPorts, "Port-Channel47", "Port-Channel551")
	case "CCS-720XP-96ZC2":
		excPorts = append(excPorts, "CCS-720XP-96ZC2")
	case "DCS-7010T":
		excPorts = append(excPorts, "Port-Channel49")
	}

	// ------------------------------
	// Filter
	filteredEntries := []ResMacAddressTableEntry{}
	for _, entry := range entries {
		if slices.Contains(excPorts, entry.Interface) {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	// ------------------------------
	// Returns only entries used by port-channels
	toReturn := []ResMacAddressTableEntry{}
	for _, entry := range filteredEntries {
		intf := entry.Interface
		_, ok := portChannelMap[intf]
		if ok {
			// Consider port-channel as interface as well
			list := portChannelMap[intf]
			intf = strings.Join(list[:], "")
		}
		entry.Interface = intf
		toReturn = append(toReturn, entry)
	}

	// ------------------------------
	return toReturn, nil
}

// for LLDP all ports request
type showLLDPNeighbors struct {
	LldpPorts map[string]*LldpPortInfo `json:"lldpNeighbors,omitempty"`
}

func (s *showLLDPNeighbors) GetCmd() string {
	return "show lldp neighbors detail"
}

type LldpPortInfo struct {
	LldpInfo []*NeighborEntry `json:"lldpNeighborInfo,omitempty"`
}

type NeighborEntry struct {
	SystemName    string          `json:"systemName,omitempty"`
	NeighIntfInfo NeighIntfInfoId `json:"neighborInterfaceInfo,omitempty"`
}

type NeighIntfInfoId struct {
	InterfaceId string `json:"interfaceId,omitempty"`
}

type ResLLDPNeighbors struct {
	Interface          string               `json:"interface,omitempty"`
	ResLDPIntNeighbors []ResLDPIntNeighbors `json:"lldp_neighbors,omitempty"`
}

type ResLDPIntNeighbors struct {
	NeighborInterfaceId string `json:"neighbor_interface_id,omitempty"`
	NeighborName        string `json:"neighbor_name,omitempty"`
}

// for LLDP single port request
type showLLDPPortPNeighbors struct {
	Interface string                          `json:"interface,omitempty"`
	LldpPorts map[string]*LldpPSinglePortInfo `json:"lldpNeighbors,omitempty"`
}

func (s *showLLDPPortPNeighbors) GetCmd() string {
	return fmt.Sprintf("show lldp neighbors ethernet %s detail", s.Interface)
}

type LldpPSinglePortInfo struct {
	LldpDetailInfo []*PortNeighborEntry `json:"lldpNeighborInfo,omitempty"`
}

type PortNeighborEntry struct {
	NeighborName     string               `json:"systemName,omitempty"`
	SystemDescr      string               `json:"systemDescription,omitempty"`
	NeighborChassId  string               `json:"chassisId,omitempty"`
	MgmtAddresses    []*MgmtAddressEntry  `json:"managementAddresses,omitempty"`
	NeighborIntfInfo PortNeighborIntfInfo `json:"neighborInterfaceInfo,omitempty"`
}

type MgmtAddressEntry struct {
	InterfaceNum int    `json:"interfaceNum,omitempty"`
	Address      string `json:"address,omitempty"`
}

type PortNeighborIntfInfo struct {
	NeighborIntfDescr  string            `json:"interfaceDescription,omitempty"`
	NeighborIntfId     string            `json:"interfaceId,omitempty"`
	NeighborIntfType   string            `json:"interfaceIdType,omitempty"`
	NeighborVlanName   map[string]string `json:"vlanNames,omitempty"`
	NeighborLinkAggr   string            `json:"linkAggregation8023Status,omitempty"`
	NeighborLinkAggrId int               `json:"linkAggregation8023InterfaceId"`
}

type ResLLDPPortNeighbors struct {
	Interface          string                `json:"interface,omitempty"`
	ResLDPIntNeighbors []ResLLDPPortNeighbor `json:"lldp_neighbors,omitempty"`
}

type ResLLDPPortNeighbor struct {
	NeighborName        string            `json:"neighbor_name,omitempty"`
	NeighborSystemDescr string            `json:"neighbor_system_description,omitempty"`
	NeighborMgmtIP      string            `json:"neighbor_management_ip,omitempty"`
	NeighborChassId     string            `json:"neighbor_chassis_id,omitempty"`
	NeighborIntfDescr   string            `json:"neighbor_interface_description,omitempty"`
	NeighborIntfId      string            `json:"neighbor_interface_id,omitempty"`
	NeighborIntfType    string            `json:"neighbor_interface_id_type,omitempty"`
	NeighborVlanName    map[string]string `json:"neighbor_vlan_name"`
	NeighborLinkAggr    string            `json:"neighbor_link_aggregation,omitempty"`
	NeighborLinkAggrId  int               `json:"neighbor_link_aggregation_id"`
}

func (a *AristaClient) GetLLDPNeighbors(ctx context.Context, req PortParamsRequest) ([]ResLLDPNeighbors, error) {

	node := a.Node
	entries := []ResLLDPNeighbors{}
	handle, err := node.GetHandle("json")
	if err != nil {
		return entries, err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("enableCmd command failed, %v", err)
	}

	// ------------------------------
	err = utils.ValidateSwitchFQDN(req.SwitchFQDN, "")
	if err != nil {
		return entries, fmt.Errorf("BadRequest: %v", err)
	}

	shLLDPCmdRsp := &showLLDPNeighbors{}
	err = handle.AddCommand(shLLDPCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("showLLDPNeighbors command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return entries, fmt.Errorf("handle.Call failed, %v", err)
	}

	// ------------------------------
	if len(shLLDPCmdRsp.LldpPorts) > 0 {
		// loop through each port
		for port, info := range shLLDPCmdRsp.LldpPorts {
			if len(info.LldpInfo) > 0 {
				lldpNeighbors := []ResLDPIntNeighbors{}
				// port has a list of neighbor info
				for _, neighbor := range info.LldpInfo {
					neighborName := neighbor.SystemName
					neighborInfo := neighbor.NeighIntfInfo
					// neighbor info contains key-value pair
					intfId := neighborInfo.InterfaceId
					lldpNeighbors = append(lldpNeighbors, ResLDPIntNeighbors{
						NeighborInterfaceId: intfId,
						NeighborName:        neighborName,
					})
				}
				// Add this interface with found neighbors
				entries = append(entries, ResLLDPNeighbors{
					Interface:          port,
					ResLDPIntNeighbors: lldpNeighbors,
				})
			}
		}
	}

	// ------------------------------
	return entries, nil
}

func (a *AristaClient) GetLLDPPortNeighbors(ctx context.Context, req PortParamsRequest) ([]ResLLDPPortNeighbors, error) {

	node := a.Node
	entries := []ResLLDPPortNeighbors{}
	handle, err := node.GetHandle("json")
	if err != nil {
		return entries, err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("enableCmd command failed, %v", err)
	}

	// ------------------------------
	err = utils.ValidateSwitchFQDN(req.SwitchFQDN, "")
	if err != nil {
		return entries, fmt.Errorf("BadRequest: %v", err)
	}

	if req.SwitchPort != "none" {
		// Port number should be <num>[/<num>] format (no "Ethernet" prefix)
		err := utils.ValidatePortNumber(req.SwitchPort)
		if err != nil {
			return entries, fmt.Errorf("BadRequest: %v", err)
		}

		shLLDPCmdRsp := &showLLDPPortPNeighbors{
			Interface: req.SwitchPort,
		}
		err = handle.AddCommand(shLLDPCmdRsp)
		if err != nil {
			return entries, fmt.Errorf("showLLDPPortNeighbors command failed, %v", err)
		}

		err = handle.Call()
		if err != nil {
			return entries, fmt.Errorf("handle.Call failed, %v", err)
		}

		// ------------------------------
		if len(shLLDPCmdRsp.LldpPorts) > 0 {
			// only 1 port will be shown if found
			for port, info := range shLLDPCmdRsp.LldpPorts {
				if len(info.LldpDetailInfo) > 0 {
					lldpNeighbors := []ResLLDPPortNeighbor{}
					// port has a list of neighbor info
					for _, neighbor := range info.LldpDetailInfo {
						neighborName := neighbor.NeighborName
						systemDescr := neighbor.SystemDescr
						neighborChassId := neighbor.NeighborChassId
						mgmtAddreses := neighbor.MgmtAddresses
						intfInfo := neighbor.NeighborIntfInfo

						neighborIP := "no data found"
						if mgmtAddreses[0].Address != "" {
							// Use this address to fill neighborName if neighborName is empty
							neighborIP = mgmtAddreses[0].Address
						}

						if neighborName == "" {
							neighborName = neighborIP
						}

						// intfId := value.(string)
						lldpNeighbors = append(lldpNeighbors, ResLLDPPortNeighbor{
							NeighborName:        neighborName,
							NeighborSystemDescr: systemDescr,
							NeighborMgmtIP:      neighborIP,
							NeighborChassId:     neighborChassId,
							NeighborIntfDescr:   intfInfo.NeighborIntfDescr,
							NeighborIntfId:      intfInfo.NeighborIntfId,
							NeighborIntfType:    intfInfo.NeighborIntfType,
							NeighborVlanName:    intfInfo.NeighborVlanName,
							NeighborLinkAggr:    intfInfo.NeighborLinkAggr,
							NeighborLinkAggrId:  intfInfo.NeighborLinkAggrId,
						})
					}
					// Add this interface with found neighbors
					entries = append(entries, ResLLDPPortNeighbors{
						Interface:          port,
						ResLDPIntNeighbors: lldpNeighbors,
					})
				} else {
					entries = append(entries, ResLLDPPortNeighbors{
						Interface:          port,
						ResLDPIntNeighbors: []ResLLDPPortNeighbor{},
					})
				}
			}
		}
	}

	// ------------------------------
	return entries, nil
}

func (a *AristaClient) SaveConfig(ctx context.Context, fqdn string) (string, error) {

	err := utils.ValidateSwitchFQDN(fqdn, "")
	if err != nil {
		return "ValidateSwitchFQDN failed", err
	}

	// ------------------------------
	// This is using Arista switch command
	return a.SaveRunningConfigAsStartupConfig(ctx)
}

// --------------------
type showIpArpVrfAll struct {
	IpArpInfo *IpArpInfo `json:"vrfs,omitempty"`
}

type IpArpInfo struct {
	P051          P051Entry          `json:"P051,omitempty"`
	ProviderInfra ProviderInfraEntry `json:"ProviderInfra,omitempty"`
	Tenants       TenantEntry        `json:"Tenants,omitempty"`
}

func (s *showIpArpVrfAll) GetCmd() string {
	return "show ip arp vrf all"
}

type showIntfPhysical struct {
	PhysStatuses map[string]interface{} `json:"interfacePhyStatuses,omitempty"`
}

func (s *showIntfPhysical) GetCmd() string {
	return "show interfaces phy"
}

type showMacAddressInterface struct {
	PortRange       string           `json:"portRange,omitempty"`
	MacUnicastTable *MacUnicastTable `json:"unicastTable,omitempty"`
}

func (s *showMacAddressInterface) GetCmd() string {
	return fmt.Sprintf("show mac address-table interface ethernet %s", s.PortRange)
}

type showMacAddrDynPOInterface struct {
	MacUnicastTable *MacUnicastTable `json:"unicastTable,omitempty"`
}

func (s *showMacAddrDynPOInterface) GetCmd() string {
	return "show mac address-table dynamic interface port-Channel 1-281"
}

type P051Entry struct {
	IpV4Neighbors []*IpV4Neighbor `json:"ipV4Neighbors,omitempty"`
}

type ProviderInfraEntry struct {
	IpV4Neighbors []*IpV4Neighbor `json:"ipV4Neighbors,omitempty"`
}

type TenantEntry struct {
	IpV4Neighbors []*IpV4Neighbor `json:"ipV4Neighbors,omitempty"`
}

type IpV4Neighbor struct {
	Address   string `json:"address,omitempty"`
	HwAddress string `json:"hwAddress,omitempty"`
	Interface string `json:"interface,omitempty"`
}

type ResIpMacInfo struct {
	Interface  string `json:"interface"`
	IpAddress  string `json:"ip_address"`
	MacAddress string `json:"mac_address"`
	VlanNum    int    `json:"vlan_tag"`
}

func (a *AristaClient) GetIpMacInfo(ctx context.Context, req ParamsRequest) ([]ResIpMacInfo, error) {

	// ------------------------------
	entries := []ResIpMacInfo{}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return entries, err
	}

	// -----------------------------------------------------------
	// find port-range
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("enableCmd command failed, %v", err)
	}

	shIntfPhysicalRsp := &showIntfPhysical{}
	err = handle.AddCommand(shIntfPhysicalRsp)
	if err != nil {
		return entries, fmt.Errorf("showIntfPhysical command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return entries, fmt.Errorf("handle.Call failed, %v", err)
	}

	allIntfs := []string{}
	for intf := range shIntfPhysicalRsp.PhysStatuses {
		intf = strings.Replace(intf, "Ethernet", "", -1)
		port := strings.Split(intf, "/")
		if len(port[0]) == 1 {
			// Prepend "0" so sort will be correct
			intf = fmt.Sprintf("0%s", intf)
		}
		allIntfs = append(allIntfs, intf)
	}

	if len(allIntfs) == 0 {
		// no ports found, just return
		return entries, nil
	}

	sort.Strings(allIntfs)
	first := allIntfs[0]
	if string(first[0]) == "0" {
		first = first[1:]
	}
	last := allIntfs[len(allIntfs)-1]

	// Set port-range
	portRange := fmt.Sprintf("%s-%s", first, last)

	// -----------------------------------------------------------
	// Get mac ip info
	enableCmdRsp = &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return entries, fmt.Errorf("enableCmd command failed, %v", err)
	}

	shIpArpVrfAllRsp := &showIpArpVrfAll{}
	err = handle.AddCommand(shIpArpVrfAllRsp)
	if err != nil {
		return entries, fmt.Errorf("showIpArpVrfAll command failed, %v", err)
	}

	shMacAddrIntfRsp := &showMacAddressInterface{
		PortRange: portRange,
	}
	err = handle.AddCommand(shMacAddrIntfRsp)
	if err != nil {
		return entries, fmt.Errorf("showMacAddressInterface command failed, %v", err)
	}

	shMacAddrIntfPORsp := &showMacAddrDynPOInterface{}
	err = handle.AddCommand(shMacAddrIntfPORsp)
	if err != nil {
		return entries, fmt.Errorf("showMacAddressInterfacePO command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return entries, fmt.Errorf("handle.Call failed, %v", err)
	}

	// ------------------------------
	// show mac address-table interface ethernet *
	//    & show mac address-table dynamic interface port-Channel 1-281
	AllMacEntries := []*MacAddressTableEntry{}
	if len(shMacAddrIntfRsp.MacUnicastTable.MacTableEntries) > 0 {
		AllMacEntries = append(AllMacEntries, shMacAddrIntfRsp.MacUnicastTable.MacTableEntries...)
	}

	if len(shMacAddrIntfPORsp.MacUnicastTable.MacTableEntries) > 0 {
		AllMacEntries = append(AllMacEntries, shMacAddrIntfPORsp.MacUnicastTable.MacTableEntries...)
	}

	// ------------------------------
	// show ip arp vrf all"
	AllIpV4Neighbor := []IpV4Neighbor{}
	if shIpArpVrfAllRsp.IpArpInfo != nil && len(shIpArpVrfAllRsp.IpArpInfo.P051.IpV4Neighbors) > 0 {
		for _, item := range shIpArpVrfAllRsp.IpArpInfo.P051.IpV4Neighbors {
			AllIpV4Neighbor = append(AllIpV4Neighbor, *item)
		}
	}

	if shIpArpVrfAllRsp.IpArpInfo != nil && len(shIpArpVrfAllRsp.IpArpInfo.ProviderInfra.IpV4Neighbors) > 0 {
		for _, item := range shIpArpVrfAllRsp.IpArpInfo.ProviderInfra.IpV4Neighbors {
			AllIpV4Neighbor = append(AllIpV4Neighbor, *item)
		}
	}

	if shIpArpVrfAllRsp.IpArpInfo != nil && len(shIpArpVrfAllRsp.IpArpInfo.Tenants.IpV4Neighbors) > 0 {
		for _, item := range shIpArpVrfAllRsp.IpArpInfo.Tenants.IpV4Neighbors {
			AllIpV4Neighbor = append(AllIpV4Neighbor, *item)
		}
	}

	ArpData := []ResIpMacInfo{}
	for _, item := range AllIpV4Neighbor {
		intfs := strings.Split(item.Interface, ",")
		// Ignore Vxlan1 entries, since they are not local to the switch
		if len(intfs) == 2 && !slices.Contains(intfs, "Vxlan1") {
			// ip_address and mac_address are used later
			// Include vlan_tag and interface variables for debugging
			// fmt.Printf("%+v \n", item)
			vlanNum := strings.Replace(intfs[0], "Vlan", "", 1)
			i, err := strconv.Atoi(vlanNum)
			if err == nil {
				ArpData = append(ArpData, ResIpMacInfo{
					Interface:  intfs[1],
					IpAddress:  item.Address,
					MacAddress: utils.ConvertMac(item.HwAddress),
					VlanNum:    i,
				})
			}
		}
	}

	// ------------------------------
	// Create mapping for MAC data
	MacData := []ResIpMacInfo{}
	for _, mac := range AllMacEntries {
		if mac.Interface != "Vxlan1" {
			MacData = append(MacData, ResIpMacInfo{
				Interface:  mac.Interface,
				IpAddress:  "unknown",
				MacAddress: mac.MacAddress,
				VlanNum:    mac.VlanId,
			})
		}
	}

	// Update MAC data with IP address from ARP table
	for _, macEntry := range MacData {
		// For each MAC entry check the ARP table. Nested for loop will scale ok
		isFound := false
		for _, arpEntry := range ArpData {
			// If entry in ARP table matches an entry in the MAC table
			if macEntry.MacAddress == arpEntry.MacAddress {
				entries = append(entries, ResIpMacInfo{
					Interface:  macEntry.Interface,
					IpAddress:  arpEntry.IpAddress,
					MacAddress: macEntry.MacAddress,
					VlanNum:    macEntry.VlanNum,
				})
				isFound = true
				break
			}
		}
		if !isFound {
			entries = append(entries, macEntry)
		}
	}

	// ------------------------------
	return entries, nil
}

func (a *AristaClient) GetPortRunningConfig(ctx context.Context, req PortParamsRequest) ([]string, error) {

	node := a.Node
	entries := []string{}

	var runCmd string
	if req.SwitchPort != "" {
		err := utils.ValidatePortNumber(req.SwitchPort)
		if err != nil {
			return nil, err
		}
		runCmd = fmt.Sprintf("show running-config interfaces Ethernet%s", req.SwitchPort)
	} else if req.PortChannel != 0 {
		err := utils.ValidatePortChannelNumber(req.PortChannel)
		if err != nil {
			return nil, err
		}
		portChannelName, err := utils.PortChannelNumberToInterfaceName(req.PortChannel)
		if err != nil {
			return nil, err
		}
		runCmd = fmt.Sprintf("show running-config interfaces %s", portChannelName)
	} else {
		return entries, fmt.Errorf("must specify switchport or portchannel")
	}

	showRunningPortCfg, err := node.RunCommands([]string{runCmd}, "text")
	if err != nil {
		return entries, fmt.Errorf("error during 'show running-config interfaces': %v", err)
	}

	interfaces := showRunningPortCfg.Result[0]
	for key, info := range interfaces {
		if key == "output" {
			lines := strings.Split(info.(string), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					entries = append(entries, line)
				}
			}
		}
	}

	// ------------------------------
	return entries, nil
}

type clearMacAddrTable struct {
	Output string
}

func (s *clearMacAddrTable) GetCmd() string {
	return "clear mac address-table dynamic"
}

func (a *AristaClient) ClearMacAddressTable(ctx context.Context, fqdn string) (string, error) {

	node := a.Node
	message := ""
	handle, err := node.GetHandle("json")
	if err != nil {
		return message, err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return message, fmt.Errorf("enableCmd command failed, %v", err)
	}

	// ------------------------------
	err = utils.ValidateSwitchFQDN(fqdn, "")
	if err != nil {
		return message, fmt.Errorf("BadRequest: %v", err)
	}

	clearCmdRsp := &clearMacAddrTable{}
	err = handle.AddCommand(clearCmdRsp)
	if err != nil {
		return message, fmt.Errorf("clearMacAddrTable command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return message, fmt.Errorf("handle.Call failed, %v", err)
	}

	// ------------------------------
	return message, nil
}

type AssignSwitchPortToPortChannelCmd struct {
	TargetPortChannel int64
}

// if the port channel already exist, nothing will happen.
// if the port channel doesn't exist, it will create a new one.
func (s *AssignSwitchPortToPortChannelCmd) GetCmd() string {
	return fmt.Sprintf("channel-group %v mode active", s.TargetPortChannel)
}

type SetLacpTimerFastCmd struct {
}

func (s *SetLacpTimerFastCmd) GetCmd() string {
	return fmt.Sprintf("lacp timer fast")
}

func (a *AristaClient) AssignSwitchPortToPortChannel(ctx context.Context, req AssignSwitchPortToPortChannelRequest) error {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("AristaClient.AssignSwitchPortToPortChannel").Start()
	defer span.End()
	logger.Info("Starting AssignSwitchPortToPortChannel", "req", req)

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmd command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.SwitchPort,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}

	assignSPtoPCCmdRsp := &AssignSwitchPortToPortChannelCmd{
		TargetPortChannel: req.TargetPortChannel,
	}
	err = handle.AddCommand(assignSPtoPCCmdRsp)
	if err != nil {
		return fmt.Errorf("AssignSwitchPortToPortChannel command failed, %v", err)
	}

	setLacpTimerFastCmdRsp := &SetLacpTimerFastCmd{}
	err = handle.AddCommand(setLacpTimerFastCmdRsp)
	if err != nil {
		return fmt.Errorf("setLacpTimerFast command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
	}

	return nil
}

type RemoveSwitchPortFromPortChannel struct {
}

func (s *RemoveSwitchPortFromPortChannel) GetCmd() string {
	return fmt.Sprintf("no channel-group")
}

func (a *AristaClient) RemoveSwitchPortFromPortChannel(ctx context.Context, req RemoveSwitchPortFromPortChannelRequest) error {
	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmd command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: req.SwitchPort,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}

	removeSPFromPCCmdRsp := &RemoveSwitchPortFromPortChannel{}
	err = handle.AddCommand(removeSPFromPCCmdRsp)
	if err != nil {
		return fmt.Errorf("RemoveSwitchPortFromPortChannel command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
	}

	return nil
}

/*
{
  "jsonrpc": "2.0",
  "id": "EapiExplorer-1",
  "result": [
    {
      "numberOfChannelsInUse": 0,
      "numberOfAggregators": 0,
      "portChannels": {
        "Port-Channel261": {
          "lacpMode": "active",
          "protocol": "lacp",
          "linkState": "down",
          "fallback": {
            "status": "fallbackIndividual"
          },
          "ports": {
            "Ethernet26/1": {
              "intf": "Ethernet26/1",
              "lagMember": false,
              "linkDown": true,
              "suspended": false,
              "lacpMisconfig": {
                "status": "noAgg"
              },
              "lacpState": {
                "activity": true,
                "timeout": true,
                "aggregation": true,
                "synchronization": false,
                "collecting": false,
                "distributing": false,
                "defaulted": true,
                "expired": false
              },
              "staticLag": false,
              "exceedsMaxWeight": false
            }
          },
          "inactive": false
        }
      }
    }
  ]
}
*/

type ShowPortChannelDense struct {
	NumberOfAggregators   int64                  `json:"numberOfAggregators"`
	NumberOfChannelsInUse int64                  `json:"numberOfChannelsInUse"`
	PortChannels          map[string]PortChannel `json:"portChannels"`
}

func (s *ShowPortChannelDense) GetCmd() string {
	return "show port-channel dense"
}

type PortChannel struct {
	LacpMode  string                           `json:"lacpMode"`
	Protocol  string                           `json:"protocol"`
	LinkState string                           `json:"linkState"`
	Ports     map[string]PortChannelPortMember `json:"ports"`
}

type PortChannelPortMember struct {
	Intf             string         `json:"intf"`
	LagMember        *bool          `json:"lagMember"`
	LinkDown         *bool          `json:"linkDown"`
	Suspended        *bool          `json:"suspended"`
	LacpMisconfig    *LacpMisconfig `json:"lacpMisconfig"`
	LacpState        *LacpState     `json:"lacpState"`
	StaticLag        *bool          `json:"staticLag"`
	ExceedsMaxWeight *bool          `json:"exceedsMaxWeight"`
}

type LacpMisconfig struct {
	Status string `json:"status"`
}

type LacpState struct {
	Activity        *bool `json:"activity"`
	Timeout         *bool `json:"timeout"`
	Aggregation     *bool `json:"aggregation"`
	Synchronization *bool `json:"synchronization"`
	Collecting      *bool `json:"collecting"`
	Distributing    *bool `json:"distributing"`
	Defaulted       *bool `json:"defaulted"`
	Expired         *bool `json:"expired"`
}

func (a *AristaClient) GetPortChannels(ctx context.Context, req GetPortChannelsRequest) (map[string]PortChannel, error) {

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return nil, err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return nil, fmt.Errorf("enableCmd command failed, %v", err)
	}

	showPCCmdRsp := &ShowPortChannelDense{}
	err = handle.AddCommand(showPCCmdRsp)
	if err != nil {
		return nil, fmt.Errorf("ShowPortChannelDense command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return nil, fmt.Errorf("handle.Call failed, %v", err)
	}

	return showPCCmdRsp.PortChannels, nil
}

type CreatePortChannelCmd struct {
	PortChannel int64
}

// if the port channel already exist, nothing will happen.
// if the port channel doesn't exist, it will create a new one.
func (s *CreatePortChannelCmd) GetCmd() string {
	return fmt.Sprintf("interface port-channel %v", s.PortChannel)
}

func (a *AristaClient) CreatePortChannel(ctx context.Context, req CreatePortChannelRequest) error {

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmd command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	createPCCmdRsp := &CreatePortChannelCmd{
		PortChannel: req.PortChannel,
	}
	err = handle.AddCommand(createPCCmdRsp)
	if err != nil {
		return fmt.Errorf("CreatePortChannel command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
	}

	return nil
}

type NoInterfacePortChannelCmd struct {
	TargetPortChannel int64
}

// if the port channel already exist, nothing will happen.
// if the port channel doesn't exist, it will create a new one.
func (s *NoInterfacePortChannelCmd) GetCmd() string {
	return fmt.Sprintf("no interface port-channel %v", s.TargetPortChannel)
}

func (a *AristaClient) DeletePortChannel(ctx context.Context, req DeletePortChannelRequest) error {
	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return err
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmd command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	deletePCCmdRsp := &NoInterfacePortChannelCmd{
		TargetPortChannel: req.TargetPortChannel,
	}
	err = handle.AddCommand(deletePCCmdRsp)
	if err != nil {
		return fmt.Errorf("DeletePortChannel command failed, %v", err)
	}

	err = handle.Call()
	if err != nil {
		return fmt.Errorf("handle.Call failed, %v", err)
	}

	return nil
}

/*
type selectPortChannelCmd struct {
	PortChannel int32
}

func (s *selectPortChannelCmd) GetCmd() string {
	return fmt.Sprintf("interface Port-Channel %v", s.PortChannel)
}


func (a *AristaClient) UpdatePortChannelVlan(ctx context.Context, req UpdatePortChannelVlanRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdatePortChannelVlan").WithValues(utils.LogFieldVlanID, req.Vlan)
	err := a.performVlanUpdate(ctx, req.PortChannel, req.Vlan, req.Description)
	if err != nil {
		logger.Error(err, "UpdatePortChannelVlan failed")
		return err
	}
	return nil
}

func (a *AristaClient) performVlanUpdate(ctx context.Context, portChannel int32, vlan int32, description string) error {
	logger := log.FromContext(ctx).WithName("AristaClient.performVlanUpdate").WithValues(utils.LogFieldVlanID, vlan)
	startTime := time.Now().UTC()
	//enable
	//configure
	//interface Port-Channel 27
	//switchport access vlan 21

	// portName := fmt.Sprintf("Port-Channel%v", portChannel)
	// err := utils.ValidatePortValue(portName)
	// if err != nil {
	// 	return fmt.Errorf("ValidatePortValue failed, error: %v", err)
	// }

	err := utils.ValidateVlanValue(vlan, a.TenantValidation)
	if err != nil {
		return fmt.Errorf("ValidateVlanValue failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}

	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmdRsp command failed, %v", err)
	}

	// TODO this doesn't work.
	// I tried to use "Port-Channel111" as interface name in the GetCmd() but failed somwhow, need to further debug the root case.
	// selectInterfaceCmdRsp := &selectInterfaceCmd{
	// 	interfaceName: portName,
	// }
	// err = handle.AddCommand(selectInterfaceCmdRsp)
	// if err != nil {
	// 	return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	// }

	selectPortChannelCmdRsp := &selectPortChannelCmd{
		PortChannel: portChannel,
	}
	err = handle.AddCommand(selectPortChannelCmdRsp)
	if err != nil {
		return fmt.Errorf("selectPortChannelCmd command failed, %v", err)
	}

	// only update description if provided
	if len(description) > 0 {
		setDescriptionCmdRsp := &setDescriptionCmd{
			description: description,
		}
		err = handle.AddCommand(setDescriptionCmdRsp)
		if err != nil {
			return fmt.Errorf("setDescriptionCmdRsp command failed, %v", err)
		}
	}

	switchportAccessVlanCmdRsp := &switchportAccessVlanCmd{
		vlanId: int(vlan),
	}
	err = handle.AddCommand(switchportAccessVlanCmdRsp)
	if err != nil {
		return fmt.Errorf("switchportAccessVlanCmdRsp command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		timeElapsed := time.Since(startTime)
		logger.Info("UpdateVlan success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}
*/

/*
func (a *AristaClient) UpdatePortChannelMode(ctx context.Context, req UpdatePortChannelModeRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdatePortChannelMode").WithValues(utils.LogFieldVlanID, req.Mode)
	err := a.performPortChannelModeUpdate(ctx, req.PortChannel, req.Mode)
	if err != nil {
		logger.Error(err, "performPortChannelModeUpdate failed")
		return err
	}
	return nil
}

func (a *AristaClient) performPortChannelModeUpdate(ctx context.Context, portChannel int32, mode string) error {
	logger := log.FromContext(ctx).WithName("AristaClient.performPortChannelModeUpdate").WithValues(utils.LogFieldMode, mode)
	startTime := time.Now().UTC()
	//enable
	//configure
	//interface po12

	// portName := fmt.Sprintf("Port-Channel%v", portChannel)
	// err := utils.ValidatePortValue(portChannel)
	// if err != nil {
	// 	return fmt.Errorf("ValidatePortValue failed, error: %v", err)
	// }

	err := utils.ValidateModeValue(mode, a.TenantValidation)
	if err != nil {
		return fmt.Errorf("ValidateModeValue failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmd command failed, %v", err)
	}

	// TODO this doesn't work.
	// selectInterfaceCmdRsp := &selectInterfaceCmd{
	// 	interfaceName: portName,
	// }
	// err = handle.AddCommand(selectInterfaceCmdRsp)
	// if err != nil {
	// 	return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	// }

	selectPortChannelCmdRsp := &selectPortChannelCmd{
		PortChannel: portChannel,
	}
	err = handle.AddCommand(selectPortChannelCmdRsp)
	if err != nil {
		return fmt.Errorf("selectPortChannelCmd command failed, %v", err)
	}

	if mode == "access" {
		// switchport mode access
		modeAccessCmdRsp := &modeAccessCmd{}
		handle.AddCommand(modeAccessCmdRsp)
		if err != nil {
			return fmt.Errorf("modeAccessCmd command failed, %v", err)
		}

	} else if mode == "trunk" {
		// switchport mode trunk
		// no switchport access vlan
		modeTrunkCmdRsp := &modeTrunkCmd{}
		handle.AddCommand(modeTrunkCmdRsp)
		if err != nil {
			return fmt.Errorf("modeTrunkCmdRsp command failed, %v", err)
		}

		noSwitchportAccessVlanCmdRsp := &noSwitchportAccessVlanCmd{}
		handle.AddCommand(noSwitchportAccessVlanCmdRsp)
		if err != nil {
			return fmt.Errorf("noSwitchportAccessVlanCmd command failed, %v", err)
		}
	}

	spanningTreePortfastCmdRsp := &spanningTreePortfastCmd{}
	handle.AddCommand(spanningTreePortfastCmdRsp)
	if err != nil {
		return fmt.Errorf("spanningTreePortfastCmd command failed, %v", err)
	}

	spanningTreeBpduGuardCmdRsp := &spanningTreeBpduGuardCmd{}
	handle.AddCommand(spanningTreeBpduGuardCmdRsp)
	if err != nil {
		return fmt.Errorf("spanningTreeBpduGuardCmd command failed, %v", err)
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		timeElapsed := time.Since(startTime)
		logger.V(1).Info("performPortChannelModeUpdate success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}
*/

/*
func (a *AristaClient) UpdatePortChannelTrunkGroup(ctx context.Context, req UpdatePortChannelTrunkGroupRequest) error {
	logger := log.FromContext(ctx).WithName("AristaClient.UpdatePortChannelTrunkGroup").WithValues(utils.LogFieldTrunkGroups, req.TrunkGroups)
	err := a.performPortChannelTrunkGroupUpdate(ctx, req.PortName, req.TrunkGroups, req.Description, req.NativeVlan)
	if err != nil {
		logger.Error(err, "performPortChannelTrunkGroupUpdate failed")
		return err
	}
	return nil
}

func (a *AristaClient) performPortChannelTrunkGroupUpdate(ctx context.Context, PortName string, TrunkGroups []string, Description string, NativeVlan int32) error {
	logger := log.FromContext(ctx).WithName("AristaClient.performPortChannelTrunkGroupUpdate").WithValues(utils.LogFieldTrunkGroups, TrunkGroups, utils.LogFieldNativeVlan, NativeVlan)
	startTime := time.Now().UTC()
	// enable
	// configure
	// interface ethernet27/1
	// switchport trunk native vlan 21
	// switchport trunk group Provider_Nets
	// switchport trunk group Tenant_Nets

	err := utils.ValidatePortValue(PortName)
	if err != nil {
		return fmt.Errorf("validate PortName failed, error: %v", err)
	}
	err = utils.ValidateTrunkGroups(TrunkGroups, a.TenantValidation)
	if err != nil {
		return fmt.Errorf("ValidateTrunkGroups failed, error: %v", err)
	}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return fmt.Errorf("node.GetHandle failed, error: %v", err)
	}
	enableCmdRsp := &enableCmd{}
	err = handle.AddCommand(enableCmdRsp)
	if err != nil {
		return fmt.Errorf("enableCmdRsp command failed, %v", err)
	}

	configCmdRsp := &configCmd{}
	err = handle.AddCommand(configCmdRsp)
	if err != nil {
		return fmt.Errorf("configCmdRsp command failed, %v", err)
	}

	selectInterfaceCmdRsp := &selectInterfaceCmd{
		interfaceName: PortName,
	}
	err = handle.AddCommand(selectInterfaceCmdRsp)
	if err != nil {
		return fmt.Errorf("selectInterfaceCmdRsp command failed, %v", err)
	}

	// only update description if provided
	if len(Description) > 0 {
		setDescriptionCmdRsp := &setDescriptionCmd{
			description: Description,
		}
		err = handle.AddCommand(setDescriptionCmdRsp)
		if err != nil {
			return fmt.Errorf("setDescriptionCmdRsp command failed, %v", err)
		}
	}

	if NativeVlan != 0 {
		err = utils.ValidateVlanValue(NativeVlan, a.TenantValidation)
		if err != nil {
			return fmt.Errorf("ValidateVlanValue NativeVlan failed, error: %v", err)
		}

		trunkNativeVlanCmdRsp := &trunkNativeVlanCmd{
			nativeVlan: int(NativeVlan),
		}
		err = handle.AddCommand(trunkNativeVlanCmdRsp)
		if err != nil {
			return fmt.Errorf("trunkNativeVlanCmd command failed, %v", err)
		}
	}

	// Remove all trunk-groups, then add back only the specified ones.
	noTrunkGroupCmdRsp := &noTrunkGroupCmd{}
	err = handle.AddCommand(noTrunkGroupCmdRsp)
	if err != nil {
		return fmt.Errorf("noTrunkGroupCmdRsp command failed, %v", err)
	}

	for _, trunkGroup := range TrunkGroups {
		trunkNativeVlanCmdRsp := &addTrunkGroupCmd{
			trunkGroup: trunkGroup,
		}
		err = handle.AddCommand(trunkNativeVlanCmdRsp)
		if err != nil {
			return fmt.Errorf("addTrunkGroupCmd command failed, %v", err)
		}
	}

	if !a.ReadOnly {
		err = handle.Call()
		if err != nil {
			return fmt.Errorf("handle.Call failed, %v", err)
		}
		timeElapsed := time.Since(startTime)
		logger.Info("UpdateTrunkGroups success!", utils.LogFieldTimeElapsed, timeElapsed)
	}
	return nil
}*/

type showIntfStatus struct {
	Interface string               `json:"interface,omitempty"`
	PortCfgs  map[string]*PortInfo `json:"interfaceStatuses,omitempty"`
}

func (s *showIntfStatus) GetCmd() string {
	return fmt.Sprintf("show interfaces %s status", s.Interface)
}

type showIntfSwitchport struct {
	Interface   string `json:"interface,omitempty"`
	Switchports map[string]InterfaceSwitchPort
}

func (s *showIntfSwitchport) GetCmd() string {
	return fmt.Sprintf("show interfaces %s switchport", s.Interface)
}

type PortInfo struct {
	LinkStatus                          string   `json:"linkStatus,omitempty"`
	Description                         string   `json:"description,omitempty"`
	Bandwidth                           int64    `json:"bandwidth,omitempty"`
	Duplex                              string   `json:"duplex,omitempty"`
	AutoNegotiateActive                 bool     `json:"autoNegotiateActive,omitempty"`
	InterfaceType                       string   `json:"interfaceType,omitempty"`
	VlanInformation                     VlanInfo `json:"vlanInformation,omitempty"`
	LineProtocolStatus                  string   `json:"lineProtocolStatus,omitempty"`
	SwitchSideLastStatusChangeTimestamp int64    `json:"switchSideLastStatusChangeTimestamp,omitempty"`
}

type VlanInfo struct {
	InterfaceForwardingModel string `json:"interfaceForwardingModel,omitempty"`
	VlanId                   uint16 `json:"vlanId,omitempty"`
	InterfaceMode            string `json:"interfaceMode,omitempty"`
}

type ResPortInfo struct {
	Description                         string   `json:"description,omitempty"`
	LinkAggregation                     bool     `json:"link_aggregation,omitempty"`
	LinkStatus                          string   `json:"link_status,omitempty"`
	Mode                                string   `json:"mode,omitempty"`
	Bandwidth                           int64    `json:"bandwidth,omitempty"`
	Duplex                              string   `json:"duplex,omitempty"`
	VlanId                              int      `json:"vlan_tag,omitempty"`
	NativeVlan                          int      `json:"native_vlan,omitempty"`
	Port                                string   `json:"port,omitempty"`
	InterfaceName                       string   `json:"interface_name,omitempty"`
	TrunkGroups                         []string `json:"trunk_groups,omitempty"`
	UntaggedVlan                        int      `json:"untagged_vlan,omitempty"` // This is an alias of either NativeVlan or VlanId depending on current mode.
	PortChannel                         string   `json:"port_channel,omitempty"`
	SwitchSideLastStatusChangeTimestamp int64    `json:"switchSideLastStatusChangeTimestamp,omitempty"`
}

func (a *AristaClient) GetPortDetails(ctx context.Context, req PortParamsRequest) (ResPortInfo, error) {
	entry := ResPortInfo{}

	err := utils.ValidateSwitchFQDN(req.SwitchFQDN, "")
	if err != nil {
		return entry, fmt.Errorf("BadRequest: %v", err)
	}

	interfaceName := ""
	if req.SwitchPort != "" {
		// Port number should be <num>[/<num>] format (no "Ethernet" prefix)
		err = utils.ValidatePortNumber(req.SwitchPort)
		if err != nil {
			return entry, fmt.Errorf("BadRequest: %v", err)
		}
		interfaceName = fmt.Sprintf("Ethernet%s", req.SwitchPort)
	} else if req.PortChannel != 0 {
		err = utils.ValidatePortChannelNumber(req.PortChannel)
		if err != nil {
			return entry, fmt.Errorf("BadRequest: %v", err)
		}
		interfaceName = fmt.Sprintf("Port-Channel%d", req.PortChannel)
	} else {
		return entry, fmt.Errorf("BadRequest: SwitchPort or PortChannel must be specified")
	}

	allInterfacesStatus, err := a.GetSwitchPorts(ctx, GetSwitchPortsRequest{
		SwitchFQDN: req.SwitchFQDN,
	})
	if err != nil {
		return entry, err
	}

	interf, found := allInterfacesStatus[interfaceName]
	if !found || interf == nil {
		return entry, fmt.Errorf("BadRequest: Interface %s not found", interfaceName)
	} else {
		return convertSwitchPortStatusToResPortInfo(interf), nil
	}
}

func (a *AristaClient) ListPortsDetails(ctx context.Context, req ListPortParamsRequest) ([]ResPortInfo, error) {
	list := []ResPortInfo{}

	err := utils.ValidateSwitchFQDN(req.SwitchFQDN, "")
	if err != nil {
		return list, fmt.Errorf("BadRequest: %v", err)
	}

	allInterfacesStatus, err := a.GetSwitchPorts(ctx, GetSwitchPortsRequest{
		SwitchFQDN: req.SwitchFQDN,
	})
	if err != nil {
		return list, err
	}

	for _, interf := range allInterfacesStatus {
		entry := convertSwitchPortStatusToResPortInfo(interf)
		list = append(list, entry)
	}

	return list, nil
}

func (a *AristaClient) ListVlans(ctx context.Context, req ListVlansParamsRequest) ([]VlanWithTrunkGroups, error) {
	entries := []VlanWithTrunkGroups{}

	node := a.Node
	handle, err := node.GetHandle("json")
	if err != nil {
		return entries, err
	}

	// -----------------------------------------------------------
	// find port-range
	showVlanRsp := &ShowVlan{}
	err = handle.AddCommand(showVlanRsp)
	if err != nil {
		return entries, fmt.Errorf("showVlan command failed, %v", err)
	}

	showVlanTrunkGroupRsp := &ShowVlanTrunkGroup{}
	err = handle.AddCommand(showVlanTrunkGroupRsp)
	if err != nil {
		return entries, fmt.Errorf("ShowVlanTrunkGroup command failed, %v", err)
	}

	eapiGetVlansCounter.Add(1)
	err = handle.Call()
	if err != nil {
		return entries, fmt.Errorf("handle.Call failed, %v", err)
	}

	for vlanId, vlanInfo := range showVlanRsp.Vlans {
		vlanIdInt, err := strconv.Atoi(vlanId)
		if err != nil {
			return entries, fmt.Errorf("got non-integer vlanId from switch, %v", vlanId)
		}

		entry := VlanWithTrunkGroups{
			VlanId:         vlanIdInt,
			Name:           vlanInfo.Name,
			InterfaceNames: maps.Keys(vlanInfo.Interfaces),
			Status:         vlanInfo.Status,
			Dynamic:        vlanInfo.Dynamic,
		}

		// Add trunkGroupNames for this vlan.
		trunkGroups, found := showVlanTrunkGroupRsp.VlanTrunkGroups[vlanId]
		if found {
			entry.TrunkGroups = trunkGroups.Names
		}

		sort.Strings(entry.InterfaceNames)
		sort.Strings(entry.TrunkGroups)
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].VlanId < entries[j].VlanId
	})
	return entries, nil
}
