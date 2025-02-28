// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

//go:generate mockgen -destination ../mocks/bmc.go -package mocks -mock_names Interface=MockBMCInterface github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc Interface

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipmilan"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	IntelCorporation  = "Intel Corporation"
	Supermicro        = "Supermicro"
	Wiwynn            = "WIWYNN"
	Dell              = "Dell Inc."
	Quanta            = "Quanta Cloud Technology Inc."
	SushyEmulator     = "Sushy Emulator"
	envBMCNTPServer1  = "BMC_NTP_SERVER1"
	envBMCNTPServer2  = "BMC_NTP_SERVER2"
	defaultNTPServer1 = "10.104.196.174"
	defaultNTPServer2 = "10.104.192.105"
	DefaultCPUID      = "0x00000"
)

type MemoryDeviceType string

const (
	DDR5 MemoryDeviceType = "DDR5"
	HBM  MemoryDeviceType = "HBM2"
)

type HBMMode string

const (
	HBMNone  = ""
	HBMOnly  = "HBM-only"
	HBMFlat  = "Flat-1LM"
	HBMCache = "Cache-2LM"
)

type HWType int64

// Supported hardware
const (
	DenaliPass HWType = iota
	CoyotePass
	Virtual
	Gaudi2Wiwynn
	Gaudi2Smc
	Smc521GeTNRT
	Smc821GVTNRT
	Smc621CTN12R
	Smc822GANGR3IN001
	Gaudi2Dell
	DellServer
	Gaudi3Dell
	QuantaD54Q2U
	QuantaGridD55Q2U
	Smc820GHTNR2
)

func (d HWType) String() string {
	return [...]string{"DenaliPass", "CoyotePass", "Virtual", "Gaudi2Wiwynn",
		"Gaudi2Smc", "Smc521GeTNRT", "Smc821GVTNRT", "Smc621CTN12R", "Smc822GANGR3IN001",
		"Gaudi2Dell", "DellServer", "Gaudi3Dell",
		"QuantaD54Q2U", "QuantaGridD55Q2U", "Smc820GHTNR2"}[d]
}

type PCIEVendor string

// Supported PCIE vendors
const (
	Mellanox PCIEVendor = "15b3h"
	Intel    PCIEVendor = "8086h"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrHCINotSupported = errors.New("host interface not supported")
	ErrKCSNotSupported = errors.New("KCS interface not supported")
)

// define supported GPUs
const NoGpuType = ""

var pcieToGPUTable = map[string]string{
	"0x8086:0x56c0": "GPU-Flex-170",
	"0x8086:0x0bda": "GPU-Max-1100",
	"0x1da3:0x1020": "HL-225",
	"0x1da3:0x1060": "HL-325",
	"0x10de:0x20b5": "A100",
}

// Odata
type BMCOdata struct {
	OdataContext string `json:"@odata.context"`
	OdataID      string `json:"@odata.id"`
	OdataType    string `json:"@odata.type"`
	Description  string `json:"Description"`
	Members      []struct {
		OdataID string `json:"@odata.id"`
	} `json:"Members"`
	MembersOdataCount int    `json:"Members@odata.count"`
	Name              string `json:"Name"`
}

// NetworkPort raw data
type NetworkPortRaw struct {
	OdataID                    string   `json:"@odata.id"`
	OdataType                  string   `json:"@odata.type"`
	AssociatedNetworkAddresses []string `json:"AssociatedNetworkAddresses"`
	CurrentLinkSpeedMbps       int      `json:"CurrentLinkSpeedMbps"`
	FlowControlConfiguration   string   `json:"FlowControlConfiguration"`
	ID                         string   `json:"Id"`
	Name                       string   `json:"Name"`
	Oem                        struct {
		OpenBmc struct {
			OdataType    string `json:"@odata.type"`
			DeviceID     string `json:"DeviceId"`
			MediaState   int    `json:"MediaState"`
			PCIClassCode int    `json:"PCIClassCode"`
			PortIndex    int    `json:"PortIndex"`
			SlotNumber   int    `json:"SlotNumber"`
		} `json:"OpenBmc"`
	} `json:"Oem"`
	PhysicalPortNumber string `json:"PhysicalPortNumber"`
	Status             struct {
		Health       string `json:"Health"`
		HealthRollup string `json:"HealthRollup"`
		State        string `json:"State"`
	} `json:"Status"`
	VendorID string `json:"VendorId"`
}

// open bmc system raw data
type OpenBMCSystem struct {
	OdataID   string `json:"@odata.id"`
	OdataType string `json:"@odata.type"`
	Actions   struct {
		ComputerSystemReset struct {
			RedfishActionInfo string `json:"@Redfish.ActionInfo"`
			Target            string `json:"target"`
		} `json:"#ComputerSystem.Reset"`
	} `json:"Actions"`
	AssetTag string `json:"AssetTag"`
	Bios     struct {
		OdataID string `json:"@odata.id"`
	} `json:"Bios"`
	BiosVersion string `json:"BiosVersion"`
	Boot        struct {
		AutomaticRetryAttempts                         int      `json:"AutomaticRetryAttempts"`
		AutomaticRetryConfig                           string   `json:"AutomaticRetryConfig"`
		AutomaticRetryConfigRedfishAllowableValues     []string `json:"AutomaticRetryConfig@Redfish.AllowableValues"`
		BootOrder                                      []string `json:"BootOrder"`
		BootSourceOverrideEnabled                      string   `json:"BootSourceOverrideEnabled"`
		BootSourceOverrideMode                         string   `json:"BootSourceOverrideMode"`
		BootSourceOverrideModeRedfishAllowableValues   []string `json:"BootSourceOverrideMode@Redfish.AllowableValues"`
		BootSourceOverrideTarget                       string   `json:"BootSourceOverrideTarget"`
		BootSourceOverrideTargetRedfishAllowableValues []string `json:"BootSourceOverrideTarget@Redfish.AllowableValues"`
		TrustedModuleRequiredToBoot                    string   `json:"TrustedModuleRequiredToBoot"`
	} `json:"Boot"`
	Description      string `json:"Description"`
	GraphicalConsole struct {
		ConnectTypesSupported []string `json:"ConnectTypesSupported"`
		MaxConcurrentSessions int      `json:"MaxConcurrentSessions"`
		ServiceEnabled        bool     `json:"ServiceEnabled"`
	} `json:"GraphicalConsole"`
	HostWatchdogTimer struct {
		FunctionEnabled bool `json:"FunctionEnabled"`
		Status          struct {
			State string `json:"State"`
		} `json:"Status"`
		TimeoutAction string `json:"TimeoutAction"`
	} `json:"HostWatchdogTimer"`
	ID            string    `json:"Id"`
	IndicatorLED  string    `json:"IndicatorLED"`
	LastResetTime time.Time `json:"LastResetTime"`
	Links         struct {
		Chassis []struct {
			OdataID string `json:"@odata.id"`
		} `json:"Chassis"`
		ManagedBy []struct {
			OdataID string `json:"@odata.id"`
		} `json:"ManagedBy"`
	} `json:"Links"`
	LocationIndicatorActive bool `json:"LocationIndicatorActive"`
	LogServices             struct {
		OdataID string `json:"@odata.id"`
	} `json:"LogServices"`
	Manufacturer string `json:"Manufacturer"`
	Memory       struct {
		OdataID string `json:"@odata.id"`
	} `json:"Memory"`
	MemorySummary struct {
		Status struct {
			Health       string `json:"Health"`
			HealthRollup string `json:"HealthRollup"`
			State        string `json:"State"`
		} `json:"Status"`
		TotalSystemMemoryGiB int `json:"TotalSystemMemoryGiB"`
	} `json:"MemorySummary"`
	Model             string `json:"Model"`
	Name              string `json:"Name"`
	NetworkInterfaces struct {
		OdataID string `json:"@odata.id"`
	} `json:"NetworkInterfaces"`
	Oem struct {
		OpenBmc struct {
			OdataType            string `json:"@odata.type"`
			FirmwareProvisioning struct {
				OdataType          string `json:"@odata.type"`
				ProvisioningStatus string `json:"ProvisioningStatus"`
			} `json:"FirmwareProvisioning"`
			KcsPolicyControlMode struct {
				OdataType string `json:"@odata.type"`
				Value     string `json:"Value"`
			} `json:"KcsPolicyControlMode"`
			KcsPolicyControlModeRedfishAllowableValues []string `json:"KcsPolicyControlMode@Redfish.AllowableValues"`
			PhysicalLED                                struct {
				OdataType string `json:"@odata.type"`
				AmberLED  string `json:"AmberLED"`
				GreenLED  string `json:"GreenLED"`
				SusackLED string `json:"SusackLED"`
			} `json:"PhysicalLED"`
		} `json:"OpenBmc"`
		VoltageRegulators struct {
			OdataID   string `json:"@odata.id"`
			OdataType string `json:"@odata.type"`
		} `json:"VoltageRegulators"`
	} `json:"Oem"`
	PCIeDevices []struct {
		OdataID string `json:"@odata.id"`
	} `json:"PCIeDevices"`
	PCIeDevicesOdataCount int    `json:"PCIeDevices@odata.count"`
	PartNumber            string `json:"PartNumber"`
	PowerRestorePolicy    string `json:"PowerRestorePolicy"`
	PowerState            string `json:"PowerState"`
	ProcessorSummary      struct {
		CoreCount int    `json:"CoreCount"`
		Count     int    `json:"Count"`
		Model     string `json:"Model"`
		Status    struct {
			Health       string `json:"Health"`
			HealthRollup string `json:"HealthRollup"`
			State        string `json:"State"`
		} `json:"Status"`
	} `json:"ProcessorSummary"`
	Processors struct {
		OdataID string `json:"@odata.id"`
	} `json:"Processors"`
	SerialConsole struct {
		Ipmi struct {
			ServiceEnabled bool `json:"ServiceEnabled"`
		} `json:"IPMI"`
		MaxConcurrentSessions int `json:"MaxConcurrentSessions"`
		SSH                   struct {
			HotKeySequenceDisplay string `json:"HotKeySequenceDisplay"`
			Port                  int    `json:"Port"`
			ServiceEnabled        bool   `json:"ServiceEnabled"`
		} `json:"SSH"`
	} `json:"SerialConsole"`
	SerialNumber string `json:"SerialNumber"`
	Status       struct {
		Health       string `json:"Health"`
		HealthRollup string `json:"HealthRollup"`
		State        string `json:"State"`
	} `json:"Status"`
	Storage struct {
		OdataID string `json:"@odata.id"`
	} `json:"Storage"`
	SystemType   string `json:"SystemType"`
	UUID         string `json:"UUID"`
	VirtualMedia struct {
		OdataID string `json:"@odata.id"`
	} `json:"VirtualMedia"`
	VirtualMediaConfig struct {
		ServiceEnabled bool `json:"ServiceEnabled"`
	} `json:"VirtualMediaConfig"`
}

type OpenBMCPcieFunction struct {
	OdataID      string `json:"@odata.id"`
	OdataType    string `json:"@odata.type"`
	ClassCode    string `json:"ClassCode"`
	DeviceClass  string `json:"DeviceClass"`
	DeviceID     string `json:"DeviceId"`
	FunctionID   int    `json:"FunctionId"`
	FunctionType string `json:"FunctionType"`
	ID           string `json:"Id"`
	Links        struct {
		PCIeDevice struct {
			OdataID string `json:"@odata.id"`
		} `json:"PCIeDevice"`
	} `json:"Links"`
	Name              string `json:"Name"`
	RevisionID        string `json:"RevisionId"`
	SubsystemID       string `json:"SubsystemId"`
	SubsystemVendorID string `json:"SubsystemVendorId"`
	VendorID          string `json:"VendorId"`
}

// Interface provides an interface for standard and custom Redfish APIs
type Interface interface {
	GetClient() mygofish.GoFishClientAccessor
	IsVirtual() bool
	UpdateAccount(ctx context.Context, newUserName, newPassword string) error
	CreateAccount(ctx context.Context, newUserName, newPassword string) error
	GetHostCPU(ctx context.Context) (*CPUInfo, error)
	GetHostBMCAddress() (string, error)
	GetHostMACAddress(ctx context.Context) (string, error)
	GetBMCPowerState(ctx context.Context) (redfish.PowerState, error)
	PowerOnBMC(ctx context.Context) error
	PowerOffBMC(ctx context.Context) error
	SanitizeBMCBootOrder(ctx context.Context) error
	ConfigureNTP(ctx context.Context) error
	VerifyPlatformFirmwareResilience(ctx context.Context) error
	GPUDiscovery(ctx context.Context) (count int, gpuType string, err error)
	HBMDiscovery(ctx context.Context) (hbmMode string, err error)
	EnableKCS(ctx context.Context) error
	DisableKCS(ctx context.Context) error
	EnableHCI(ctx context.Context) error
	DisableHCI(ctx context.Context) error
	SetFanSpeed(ctx context.Context) error
	IsIntelPlatform() bool
	GetHwType() HWType
}

// CPUInfo contains the overall CPU information on host
type CPUInfo struct {
	// CPUID is a unique number that defines a CPU version
	CPUID string
	// Sockets is the total number of CPU sockets
	Sockets int
	// Cores is the total number of CPU cores
	Cores int
	// Threads is the total number of CPU threads
	Threads int
	// Manufacturer is the manufacturer of the CPU
	Manufacturer string
}

// Config contains BMC configuration
type Config struct {
	URL      string
	Username string
	Password string
}

// BMC provides standard Redfish API
type BMC struct {
	config       *Config
	isVirtual    bool
	name         string
	pxeBootRegex string
	APIClient    mygofish.GoFishClientAccessor
	hwType       HWType
	manufacturer string
}

// New create a new BMC interface with the specified config
func New(gofishManager mygofish.GoFishManagerAccessor, cfg *Config) (Interface, error) {
	client, err := gofishManager.Connect(gofish.ClientConfig{
		Endpoint:  cfg.URL,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Insecure:  true,
		BasicAuth: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BMC service: %v", err)
	}

	bmc := &BMC{
		APIClient: client,
		config:    cfg,
	}

	system, err := bmc.getSystem()
	if err != nil {
		impiErr := bmc.ipmiGetManufacturer()
		if impiErr != nil {
			return nil, fmt.Errorf("%v: unable to get the manufacturer using IPMI: %v", err, impiErr)
		}
		if bmc.manufacturer == "" {
			return nil, fmt.Errorf("unable to get the computing system: %v", err)
		}
	} else {
		bmc.manufacturer = system.Manufacturer()
	}

	if bmc.manufacturer == "" {
		return nil, fmt.Errorf("System.Manufacturer is missing")
	}

	if strings.EqualFold(bmc.manufacturer, SushyEmulator) {
		bmc.isVirtual = true
		bmc.name = "VirtualBMC"
		bmc.hwType = Virtual
		return &VirtualBMC{BMC: *bmc}, nil
	}

	if strings.EqualFold(bmc.manufacturer, Wiwynn) {
		bmc.name = "Gaudi2"
		bmc.hwType = Gaudi2Wiwynn
		return &WiwynnBMC{BMC: *bmc}, nil
	}

	model := system.Model()
	if model == "" {
		return nil, fmt.Errorf("System.Model is missing")
	}

	if strings.EqualFold(bmc.manufacturer, Dell) {
		switch {
		case regexp.MustCompile(iDRACXE9680Gaudi3Regex).MatchString(model):
			bmc.name = "DellGaudi3"
			bmc.hwType = Gaudi3Dell
			return &IdracBMC{BMC: *bmc}, nil
		default:
			bmc.name = "DellServer"
			bmc.hwType = DellServer
			return &IdracBMC{BMC: *bmc}, nil
		}
	}

	if strings.EqualFold(bmc.manufacturer, IntelCorporation) {
		switch {
		case regexp.MustCompile(denaliPassModelRegex).MatchString(model):
			bmc.name = "DenaliPassBMC"
			bmc.hwType = DenaliPass
			return &DenaliPassBMC{BMC: *bmc}, nil
		case regexp.MustCompile(coyotePassModelRegex).MatchString(model):
			bmc.name = "CoyotePassBMC"
			bmc.hwType = CoyotePass
			return &CoyotePassBMC{BMC: *bmc}, nil
		}
	}
	if strings.EqualFold(bmc.manufacturer, Supermicro) {
		switch {
		case regexp.MustCompile(smcGaudi2PassModelRegex).MatchString(model):
			bmc.name = "Gaudi2"
			bmc.hwType = Gaudi2Smc
			return &SmcBMC{BMC: *bmc}, nil
		case regexp.MustCompile(smcSys521GeTNRT).MatchString(model):
			bmc.name = "SYS-521GE-TNRT"
			bmc.hwType = Smc521GeTNRT
			return &SmcBMC{BMC: *bmc}, nil
		case regexp.MustCompile(smcSys821GVTNRT).MatchString(model):
			bmc.name = "SYS-821GV-TNRT"
			bmc.hwType = Smc821GVTNRT
			return &SmcBMC{BMC: *bmc}, nil
		case regexp.MustCompile(smcSys621CTN12R).MatchString(model):
			bmc.name = "SYS-621C-TN12R"
			bmc.hwType = Smc621CTN12R
			return &SmcBMC{BMC: *bmc}, nil
		case regexp.MustCompile(smcSys822GANGR3IN001).MatchString(model):
			bmc.name = "SYS-822GA-NGR3-IN001"
			bmc.hwType = Smc822GANGR3IN001
			return &SmcBMC{BMC: *bmc}, nil
		}
	}
	if strings.EqualFold(bmc.manufacturer, Quanta) {
		switch {
		case regexp.MustCompile(quantaD54Q2U).MatchString(model):
			bmc.name = "QuantaD54Q2U"
			bmc.hwType = QuantaD54Q2U
			return &QuantaBMC{BMC: *bmc}, nil
		case regexp.MustCompile(quantaGridD55Q2U).MatchString(model):
			bmc.name = "QuantaGridD55Q2U"
			bmc.hwType = QuantaGridD55Q2U
			return &QuantaV2BMC{BMC: *bmc}, nil
		}
	}

	return nil, fmt.Errorf("failed to determine BMC supported module")
}

func (c *BMC) GetHwType() HWType {
	return c.hwType
}

// GetClient returns a Redfish client
func (c *BMC) GetClient() mygofish.GoFishClientAccessor {
	return c.APIClient
}

// IsVirtual return true if the BMC is virtual
func (c *BMC) IsVirtual() bool {
	return c.isVirtual
}

// IsIntel return true if the manufacturer is Intel
func (c *BMC) IsIntelPlatform() bool {
	return c.manufacturer == IntelCorporation
}

func (c *BMC) getProcessors() ([]*redfish.Processor, error) {
	system, err := c.getSystem()
	if err != nil {
		return nil, err
	}

	processors, err := system.Processors()
	if err != nil {
		return nil, fmt.Errorf("unable to list processors: %v", err)
	}
	if len(processors) == 0 {
		return nil, fmt.Errorf("no processors found")
	}

	return processors, nil
}

// GetHostCPU returns the current available CPU of the host
func (c *BMC) GetHostCPU(ctx context.Context) (*CPUInfo, error) {
	log := log.FromContext(ctx).WithName("BMC.GetHostCPU")
	log.Info("Getting host's CPU information")
	proccessors, err := c.getProcessors()
	if err != nil {
		return nil, err
	}
	return c.getHostCPUInfo(proccessors)
}

// GetHostBMCAddress returns the URL for accessing BMC on the network
func (c *BMC) GetHostBMCAddress() (string, error) {
	system, err := c.getSystem()
	if err != nil {
		return "", fmt.Errorf("unable to get the computing system: %v", err)
	}
	address := fmt.Sprintf("redfish+%s%s", c.config.URL, system.ODataID())

	return address, nil
}

// GetHostMACAddress returns the MAC address of the host system managed by BMC
func (c *BMC) GetHostMACAddress(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("BMC.getEthernetInterface").WithValues("HWType", c.hwType)
	log.Info("Getting Ethernet Interface")

	system, err := c.getSystem()
	if err != nil {
		return "", err
	}

	ethernetInterfaces, err := system.EthernetInterfaces()
	if err != nil {
		return "", fmt.Errorf("unable to get the ethernet interface: %v", err)
	}
	var macAddress []string
	for _, eth := range ethernetInterfaces {
		switch c.hwType {
		case Virtual:
			if eth.Status.State == common.EnabledState &&
				eth.Status.Health == common.OKHealth {
				log.Info("Host interface will have this data", "mac", eth.MACAddress, "details", eth)
				macAddress = append(macAddress, eth.MACAddress)
			}
		default:
			if eth.LinkStatus == redfish.LinkUpLinkStatus {
				macAddress = append(macAddress, eth.MACAddress)
			}
		}
	}
	if len(macAddress) == 0 {
		return "", fmt.Errorf("no available ethernet interface")
	} else {
		//TODO: it is an assumption on the order of eth interfaces...
		log.Info("MAC addresses observed, the first MAC is treated as host MAC addresses", "MACs", macAddress)
		return macAddress[0], nil
	}
}

func (c *BMC) SanitizeBMCBootOrder(ctx context.Context) (err error) {
	// not implemented for standard BMC
	return nil
}

func (c *BMC) GPUDiscovery(ctx context.Context) (count int, gpuType string, err error) {
	// not implemented for standard BMC
	return 0, NoGpuType, nil
}

func (c *BMC) HBMDiscovery(ctx context.Context) (hbmMode string, err error) {
	// not implemented for standard BMC
	return string(HBMNone), nil
}

func (c *BMC) EnableKCS(ctx context.Context) (err error) {
	// not implemented for standard BMC
	log := log.FromContext(ctx).WithName("BMC.EnableKCS")
	log.Info("Enabling KCS is not supported", "platform", c.name)
	return ErrKCSNotSupported
}

func (c *BMC) DisableKCS(ctx context.Context) (err error) {
	// not implemented for standard BMC
	log := log.FromContext(ctx).WithName("BMC.DisableKCS")
	log.Info("Disabling KCS is not supported", "platform", c.name)
	return ErrKCSNotSupported
}

func (c *BMC) EnableHCI(ctx context.Context) (err error) {
	// not implemented for standard BMC
	log := log.FromContext(ctx).WithName("BMC.EnableHCI")
	log.Info("Host interface is not supported", "platform", c.name)
	return ErrHCINotSupported
}

func (c *BMC) DisableHCI(ctx context.Context) (err error) {
	// not implemented for standard BMC
	log := log.FromContext(ctx).WithName("BMC.DisableHCI")
	log.Info("Host interface is not supported", "platform", c.name)
	return ErrHCINotSupported
}

func (c *BMC) SetFanSpeed(ctx context.Context) (err error) {
	// not implemented for standard BMC
	log := log.FromContext(ctx).WithName("BMC.SetFanSpeed")
	log.Info("Setting fan speed is not supported", "platform", c.name)
	return nil
}

func (c *BMC) sanitizeBMCBootOrder(ctx context.Context, bootRegex string) (err error) {
	log := log.FromContext(ctx).WithName("BMC.sanitizeBMCBootOrder")
	log.Info("Checking BMC Boot Order and updating as needed")

	system, err := c.getSystem()
	if err != nil {
		return fmt.Errorf("unable to get the computing system: %v", err)
	}
	log.Info(fmt.Sprintf(".ODataID %v", system.ODataID()))

	origOrder := system.Boot().BootOrder()
	log.Info("Current Boot Order", "model", system.Model(), "bootOrder", origOrder)

	log.Info("Changing Boot Order based on", "model", system.Model(), "regex", bootRegex)

	newOrder, err := util.MoveMatchingStringsToStart(system.Boot().BootOrder(), bootRegex)
	if err != nil {
		return fmt.Errorf("unable to move matching strings to start: %v", err)
	}
	log.Info("Planned new Boot Order", "model", system.Model(), "bootOrder", newOrder)
	if reflect.DeepEqual(origOrder, newOrder) {
		log.Info("Boot Order already correct, No change needed", "model", system.Model(), "bootOrder", newOrder)
		return nil
	}

	newBoot := system.Boot()
	if err := newBoot.SetBootOrder(newOrder); err != nil {
		return fmt.Errorf("failed to set new Boot Order for BMC URL '%s': '%s'", c.config.URL, err)
	}
	if err := system.SetBoot(newBoot); err != nil {
		return fmt.Errorf("failed to change the Boot Order for BMC URL '%s': '%s'", c.config.URL, err)
	}

	return nil
}

func (c *BMC) getSystem() (mygofish.GoFishComputerSystemAccessor, error) {
	systems, err := c.GetClient().GetService().Systems()
	if err != nil {
		return nil, fmt.Errorf("unable to get the computing system: %v", err)
	}

	if len(systems) == 0 {
		return nil, fmt.Errorf("no system found for BMC under Services")
	}

	return systems[0], nil
}

func (c *BMC) VerifyPlatformFirmwareResilience(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.VerifyPlatformFirmwareResilience")
	log.Info("Platform Firmware Resilience", c.name, "is not supported")
	return nil
}

func (c *BMC) GetBMCPowerState(ctx context.Context) (redfish.PowerState, error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.GetBMCPowerState")
	log.Info("Getting Power State of the BMC ")

	var state redfish.PowerState
	state = "Unknown"

	systems, err := c.GetClient().GetService().Systems()
	if err != nil {
		return state, fmt.Errorf("unable to get the computing system: %v", err)
	}

	if len(systems) < 1 {
		return state, fmt.Errorf("no system found for BMC under Services")
	}
	log.Info(fmt.Sprintf(".ODataID %v", systems[0].ODataID()))

	system := systems[0]
	state = system.PowerState()
	log.Info("Found Power State", "model", system.Model(), "PowerState", state)

	return state, nil
}

func (c *BMC) PowerOffBMC(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.PowerOffBMC")
	log.Info("Forcing Power Off and monitoring BMC ")

	return c.setBMCPowerTo(ctx, redfish.OffPowerState, redfish.ForceOffResetType)
}

func (c *BMC) PowerOnBMC(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.PowerOnBMC")
	log.Info("Forcing Power On and monitoring BMC ")

	return c.setBMCPowerTo(ctx, redfish.OnPowerState, redfish.OnResetType)
}

func (c *BMC) setBMCPowerTo(ctx context.Context, powerState redfish.PowerState, resetType redfish.ResetType) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.setBMCPowerTo")
	log.Info("Forcing Power to state", "powerState", powerState)

	systems, err := c.GetClient().GetService().Systems()
	if err != nil {
		return fmt.Errorf("unable to get the computing system: %v", err)
	}

	if len(systems) < 1 {
		return fmt.Errorf("no system found for BMC under Services")
	}
	log.Info(fmt.Sprintf(".ODataID %v", systems[0].ODataID()))

	system := systems[0]
	currentState := system.PowerState()
	log.Info("Initial Power State", "model", system.Model(), "currentState", currentState)
	if currentState == powerState {
		log.Info("Initial Power State matches the request", "model", system.Model(), "currentState", currentState, "desiredState", powerState)
		return nil
	}

	if err := system.Reset(resetType); err != nil {
		return fmt.Errorf("failed to %s for BMC URL '%s': '%s'", resetType, c.config.URL, err)
	}

	timeout := time.NewTimer(4 * time.Second)
	defer timeout.Stop()

	for ever := true; ever; {
		select {
		case <-timeout.C:
			return fmt.Errorf("timeout waiting for model '%s' to transition to Power State '%s", system.Model(), powerState)
		default:
			systems, err = c.GetClient().GetService().Systems()
			if err != nil {
				return fmt.Errorf("unable to get the computing system: %v", err)
			}

			if len(systems) < 1 {
				return fmt.Errorf("no system found for BMC under Services")
			}
			system = systems[0]
			currentState = system.PowerState()

			log.Info("Current Power State", "model", system.Model(), "currentState", currentState, "TargetState", powerState)

			if currentState != powerState {
				time.Sleep(2 * time.Second)
			} else {
				ever = false
			}
		}
	}

	log.Info("Reached Target Power State", "model", system.Model(), "PowerState", currentState, "TargetState", powerState)

	return nil
}

func (c *BMC) ConfigureNTP(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.ConfigureNTP")
	log.Info("ConfigureNTP Not Supported by default")

	// not implemented for standard BMC
	return nil
}

func (c *BMC) openBMCConfigureNTP(ctx context.Context, redfishPath string) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.openBMCConfigureNTP")

	log.Info("Setting the NTP server list via OpenBMC standard BMC Manager Network Protocol")

	ntpServer1 := util.GetEnv(envBMCNTPServer1, defaultNTPServer1)
	ntpServer2 := util.GetEnv(envBMCNTPServer2, defaultNTPServer2)
	// {"NTP":{"NTPServers":["10.104.196.174", "10.104.192.105"]}}
	serversPayload := map[string]interface{}{
		"NTP": map[string][]string{
			"NTPServers": {ntpServer1, ntpServer2}},
	}

	// Patch NTP Server list using the request to the Redfish API
	response, err := c.GetClient().Patch(redfishPath, serversPayload)
	if err != nil {
		return fmt.Errorf("failed to set NTP Servers '%v': '%v'", serversPayload, err)
	}
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		log.Info("BMC NTP Servers set, but not active yet", "ntpServer1", ntpServer1, "ntpServer2", ntpServer2)
	} else {
		log.Info("Failed to set BMC NTP Servers", "ntpServer1", ntpServer1, "ntpServer2", ntpServer2)
		return fmt.Errorf("failed to set NTP Servers '%s', '%s' : '%d': '%s'", ntpServer1, ntpServer2,
			response.StatusCode, http.StatusText(response.StatusCode))
	}

	log.Info("Enabling the NTP servers via OpenBMC standard BMC Manager Network Protocol")

	// {"NTP":{"ProtocolEnabled": true}}
	enablePayload := map[string]interface{}{
		"NTP": map[string]bool{
			"ProtocolEnabled": true},
	}

	// Patch NTP Server list using the request to the Redfish API
	response, err = c.GetClient().Patch(redfishPath, enablePayload)
	if err != nil {
		return fmt.Errorf("failed to enable NTP Servers '%v': '%v'", enablePayload, err)
	}
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		log.Info("Enabled BMC NTP Servers")
	} else {
		log.Info("Failed to Enable BMC NTP Servers")
		return fmt.Errorf("failed to Enable NTP Servers: '%d': '%s'",
			response.StatusCode, http.StatusText(response.StatusCode))
	}

	return nil
}

func (c *BMC) UpdateAccount(ctx context.Context, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("BMC.UpdateBMCCredentials")
	log.Info("Updating BMC credentials")

	// Retrieve the service root
	service := c.GetClient().GetService()
	if service == nil {
		return fmt.Errorf("AccountService is nil")
	}

	// Query the AccountService using the session token
	accountService, err := service.AccountService()
	if err != nil {
		return fmt.Errorf("failed to get AccountService for BMC URL %q: %v", c.config.URL, err)
	}

	// Get list of accounts
	accounts, err := accountService.Accounts()
	if err != nil {
		return fmt.Errorf("failed to get Accounts for BMC URL %q: %v", c.config.URL, err)
	}

	// Iterate over accounts to check for the new user account
	for _, account := range accounts {
		if account.UserName == newUserName {
			log.Info("Updating password on existing Admin account on BMC", "UserName", account.UserName)
			account.Password = newPassword
			if err := account.Update(); err != nil {
				return fmt.Errorf("failed to update password for %q on BMC URL %q: '%v'", newUserName, c.config.URL, err)
			}
			log.Info("BMC password updated! Updating runtime values.", "UserName", account.UserName)

			return nil
		}
	}

	return ErrAccountNotFound
}

func (c *BMC) CreateAccount(ctx context.Context, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("BMC.CreateBMCCredentials")
	log.Info("Creating BMC account")

	// Need to add the new Admin Account
	payload := map[string]interface{}{
		"UserName": newUserName,
		"Password": newPassword,
		"RoleId":   "Administrator",
		"Enabled":  true,
	}

	// POST the request to the Redfish API
	response, err := c.GetClient().Post("/redfish/v1/AccountService/Accounts", payload)
	if err != nil {
		return fmt.Errorf("failed to create new Admin account %q on BMC URL %q: %v", newUserName, c.config.URL, err)
	}
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		log.Info("BMC Admin account created! Updating runtime values.", "newUserName", newUserName)
	} else {
		log.Info("BMC account creation failed", "bmcUsername", newUserName)
		return fmt.Errorf("failed to create new Admin account %q on BMC URL %q: %d: %q",
			newUserName, c.config.URL, response.StatusCode, http.StatusText(response.StatusCode))
	}

	return nil
}

// getCPUID returns a five-digit Hex string that represents a CPUID
func getCPUID(str string) string {
	if str == "" {
		return DefaultCPUID
	}

	id := strings.ReplaceAll(str, "-", "")
	if len(str) > 5 {
		id = id[len(id)-5:]
	}
	if !strings.HasPrefix(id, "0x") {
		id = fmt.Sprintf("0x%s", id)
	}
	return id
}

// get Manufacturer using IPMI
func (c *BMC) ipmiGetManufacturer() error {

	ipmiHelper, err := ipmilan.NewIpmiLanHelper(context.TODO(), c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	c.manufacturer, err = ipmiHelper.GetManufacturer(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to GetManufacturer via IPMI: %v", err)
	}
	return nil
}

func (c *BMC) getHostCPUInfo(processors []*redfish.Processor) (*CPUInfo, error) {
	cpuInfo := &CPUInfo{}

	var availableCPUs []*redfish.Processor
	for _, p := range processors {
		if p.ProcessorType != redfish.CPUProcessorType {
			continue
		}
		if p.Status.State != common.EnabledState {
			continue
		}
		availableCPUs = append(availableCPUs, p)
	}
	if len(availableCPUs) == 0 {
		return nil, fmt.Errorf("no available CPU information")
	}

	cpuInfo.Manufacturer = availableCPUs[0].Manufacturer
	cpuInfo.CPUID = getCPUID(availableCPUs[0].ProcessorID.IdentificationRegisters)

	for _, cpu := range availableCPUs {
		cpuInfo.Sockets += 1
		cpuInfo.Cores += cpu.TotalCores
		cpuInfo.Threads += cpu.TotalThreads
	}
	if cpuInfo.Cores == 0 {
		return nil, fmt.Errorf("no available CPU resource: %+v", cpuInfo)
	}
	// calculate threads per core
	cpuInfo.Threads = cpuInfo.Threads / cpuInfo.Cores
	return cpuInfo, nil
}
