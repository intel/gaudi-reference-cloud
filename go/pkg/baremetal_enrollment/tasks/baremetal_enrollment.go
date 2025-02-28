// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ddi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipacmd"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/myssh"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/units"
)

const (
	bmcDeploymentSecretsPath   = "deployed/"
	bmcUserSecretsPrefix       = "user/"
	bmcBIOSSecretsPrefix       = "bios/"
	envBMCEnrollUsername       = "BMC_ENROLL_USERNAME"
	defaultBMCEnrollUsername   = "bmaas"
	Metal3NamespaceSelectorKey = "cloud.intel.com/bmaas-metal3-namespace"
	metal3NamespaceIronicIPKey = "ironicIP"

	menAndMiceBMCType          = "BMC"
	menAndMiceProvisioningType = "Provisioning"

	MenAndMiceUrlEnvVar           = "MEN_AND_MICE_URL"
	MenAndMiceServerAddressEnvVar = "MEN_AND_MICE_SERVER_ADDRESS"
	TftpServerIPEnvVar            = "TFTP_SERVER"
	IPXEProfileName               = "boot.ipxe"
	IronicHttpPortNb              = "6180"
	IPXEBinarayName               = "snponly.efi"

	SetBiosPasswordEnvVar = "SET_BIOS_PASSWORD"
	DhcpProxyUrlEnvVar    = "DHCP_PROXY_URL"

	BMCInterfaceName            = "BMC"
	HostInterfaceName           = "net0/0"
	StorageInterfaceName1       = "net1/0"
	ComputeApiServerAddrEnvVar  = "COMPUTE_API_SERVER_ADDRESS"
	TotalGaudiExternalPorts     = 24
	GaudiExternalPortMacPrefix  = "b0:fd:0b"
	GaudiExternalPortMacPrefix1 = "68:93:2e"
)

// Hardware specification labels
const (
	//
	// GPU Prefix Annotation
	GPUAnnotationPrefix    = "gpu.mac.cloud.intel.com"
	GPUIPsAnnotationPrefix = "gpu.ip.cloud.intel.com"
	// Storage Prefix Annotation
	StorageMACAnnotationPrefix = "storage.mac.cloud.intel.com"
	// CPUManufacturerLabel is the manufacturer of the CPU
	CPUManufacturerLabel = "cloud.intel.com/host-cpu-manufacturer"
	// CPUIDLabel is a unique number that defines a CPU version
	CPUIDLabel = "cloud.intel.com/host-cpu-id"
	// CPUModelLabel is the model name of the CPU
	CPUModelLabel = "cloud.intel.com/host-cpu-model"
	// CPUCountLabel is the total number of logical cores or hyperthreads available on host
	CPUCountLabel = "cloud.intel.com/host-cpu-count"
	// CPUSocketsLabel is total number of enabled CPU sockets on host
	CPUSocketsLabel = "cloud.intel.com/host-cpu-sockets"
	// CPUCoresLabel is the total number of physical cores available on host
	CPUCoresLabel = "cloud.intel.com/host-cpu-cores"
	// CPUThreadsLabel is the total number of CPU threads available on host
	CPUThreadsLabel = "cloud.intel.com/host-cpu-threads"
	// GPU Count Label
	GPUCountLabel = "cloud.intel.com/host-gpu-count"
	// GPU  Device ID Label
	GPUModelNameLabel = "cloud.intel.com/host-gpu-model"
	// HBM mode Label
	HBMModeLabel = "cloud.intel.com/hbm-mode"
	// latest associated resource
	LastAssociatedInstance = "last-associated-instance"
	// Label key applied to Instance to identify the instance group.
	ClusterGroup     = "instance-group"
	LastClusterGroup = "last-cluster-group"
	// Label key applied to the memory size label to match the memory of BM and InstanceTypes
	MemorySizeLabel = "cloud.intel.com/host-memory-size"
	// Label key applied to BareMetalHost to identify hosts that connect to the same cluster fabric.
	// Hosts with the same value should connect to the same cluster fabric and can be consumed by instances in
	// the same instance group.
	ClusterGroupID = "cloud.intel.com/instance-group-id"
	// Label key to identify the supercompute group.
	SuperComputeGroupID = "cloud.intel.com/supercompute-group-id"
	// Label key applied to BareMetalHost.
	// applied value define the number of nodes in the cluster
	ClusterSize = "cloud.intel.com/cluster-size"
	// Label key applied to Instance to label the Instance Type.
	// applied value define the type of instance type
	InstanceTypeLabel = "instance-type.cloud.intel.com/%s"
	// Label key applied to Instance to label the assigned compute node pool.
	// applied value assigns the host to 'general' pool
	ComputeNodePoolLabel = "pool.cloud.intel.com/%s"
	// compute node pool for all baremetalhosts with no specific node pool.
	NodePoolGeneral = "general"
	// Labels used by the validation operator
	// Label used to trigger the validation process by the validation operator.
	ReadyToTestLabel = "cloud.intel.com/ready-to-test"
	// Label set by the validation operator to indicate that the validation completed successfully.
	VerifiedLabel = "cloud.intel.com/verified"
	// Label that indicates that the checking/validation failed. The value will have type of validation that failed
	CheckingFailedLabel = "cloud.intel.com/validation-check-failed"
	// Label used by the scheduler to prevent node from being scheduled.
	UnschedulableLabel = "cloud.intel.com/unschedulable"
	// Label used to indicate the network mode of the node.
	NetworkModeLabel = "cloud.intel.com/network-mode"
	// network mode for non-spine-leaf accelerator fabric isolation type with VLAN.
	NetworkModeVVXStandalone = "VVX-standalone"
	// network mode for spine-leaf accelerator fabric isolation type with BGP.
	NetworkModeXBX = "XBX"
	// network mode for partitioned-leaf accelerator fabric isolation type with VLAN.
	NetworkModeVVV = "VVV"
	// network mode to ignore creation of accelarator fabric.
	NetworkModeIgnore = "IGNORE_XBX"
	// Label used to indicate the firmware versions compatible with the machine image.
	FWVersionLabel = "cloud.intel.com/firmware-version"
	// Label used pass configuration to the validation scripts.
	TestConfigurationLabel = "cloud.intel.com/validation-test-configuration"

	// Label which indicates the state of the validation process
	// Imaging is in progress for validation.
	ImagingLabel = "cloud.intel.com/validation-imaging"
	// Wait for all the instances in the Instance group to complete individual validation
	WaitForInstanceValidation = "cloud.intel.com/group-wait-for-InstanceValidation"
	// Image process has completed and the bmh is initialized.
	ImagingCompletedLabel = "cloud.intel.com/validation-imaging-completed"
	// Instance validation has completed for all bmhs in the instance group.
	InstanceValidationCompletedLabel = "cloud.intel.com/validation-instance-completed"
	// Checking/Validation process for InstanceGroup is in progress.
	CheckingGroupLabel = "cloud.intel.com/group-validation-checking"
	// Checking/Validation process is in progress.
	CheckingLabel = "cloud.intel.com/validation-checking"
	// Checking/Validation process has completed.
	CheckingCompletedLabel = "cloud.intel.com/validation-checking-completed"
	// Checking/Validation process for InstanceGroup has completed.
	CheckingCompletedGroupLabel = "cloud.intel.com/validation-checking-completed-group"
	// Deletion for validation process
	DeletionLabel = "cloud.intel.com/deletion-for-validation"
	// Label to indicate a node is a master node for cluster validation
	MasterNodeLabel = "cloud.intel.com/validation-master-node"
	// Label to gate cluster validation.
	GateValidationLabel = "cloud.intel.com/validation-gating"
	// Label to represent the validation id.
	ValidationIdLabel = "cloud.intel.com/validation-id"
	// Label to skip all validation (instance and group validation)
	SkipValidationLabel = "cloud.intel.com/skip-validation"
	// Label to skip group validation (Instance validation will happen)
	SkipGroupValidationLabel = "cloud.intel.com/skip-group-validation"
	// Label to indicate validation is happening for firware upgrade.
	FWVersionUpdateTriggerLabel = "cloud.intel.com/fw-update-trigger"
	// Label to skip BMH deprovision after validation completes.
	SkipDeprovisionLabel = "cloud.intel.com/skip-deprovisioning"
)

var bmHostGVR = baremetalv1alpha1.SchemeBuilder.GroupVersion.WithResource("baremetalhosts")

type DeviceData struct {
	Name    string
	ID      int64
	Rack    string
	Region  string
	Cluster string
}

type BMCData struct {
	URL        string
	Username   string
	Password   string
	MACAddress string
}

type EnrollmentTask struct {
	deviceData *DeviceData
	bmcData    *BMCData

	bmHostNamespace *corev1.Namespace
	bmhIpAddress    string

	netBox                    dcim.DCIM
	vault                     secrets.SecretManager
	bmc                       bmc.Interface
	clientSet                 kubernetes.Interface
	dynamicClient             dynamic.Interface
	menAndMice                ddi.DDI
	instanceTypeServiceClient pb.InstanceTypeServiceClient
}

func NewEnrollmentTask(ctx context.Context) (*EnrollmentTask, error) {
	log := log.FromContext(ctx).WithName("NewEnrollmentTask")
	log.Info("Initializing new enrollment task")

	deviceData, err := getDeviceData(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get device information: %v", err)
	}

	vault, err := GetVaultClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	netBox, err := GetNetBoxClient(ctx, vault, deviceData.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client: %v", err)
	}

	clientSet, dynamicClient, err := getK8sClients(ctx)
	if err != nil {
		err := fmt.Errorf("unable to initialize K8s Client: %v", err)
		updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
		return nil, err
	}

	menAndMice, err := GetMenAndMiceClient(ctx, deviceData.Region, vault)
	if err != nil {
		err := fmt.Errorf("unable to initialize MenAndMice client: %v", err)
		updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
		return nil, err
	}

	bmcData, err := getBMCData(ctx, deviceData, netBox, vault, menAndMice)
	if err != nil {
		err := fmt.Errorf("unable to get BMC data: %v", err)
		updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
		return nil, err
	}
	bmcInterface, err := getBMCInterface(ctx, deviceData, bmcData)
	if err != nil {
		err := fmt.Errorf("unable to initialize BMC Interface: %v", err)
		updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
		return nil, err
	}

	instancetypeClient, err := GetInstanceTypeServiceClient(ctx)
	if err != nil {
		err := fmt.Errorf("unable to initialize InstanceTypeServiceClient: %v", err)
		updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
		return nil, err
	}

	task := &EnrollmentTask{
		deviceData:                deviceData,
		bmcData:                   bmcData,
		vault:                     vault,
		netBox:                    netBox,
		bmc:                       bmcInterface,
		clientSet:                 clientSet,
		dynamicClient:             dynamicClient,
		instanceTypeServiceClient: instancetypeClient,
		menAndMice:                menAndMice,
	}

	if task.bmc.GetHwType() != bmc.Gaudi2Wiwynn {
		if err = task.verifyBMCCredentials(ctx); err != nil {
			err := fmt.Errorf("unable to verify BMC creds: %v", err)
			updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
			return nil, err
		}
	}

	if !task.bmc.IsVirtual() {
		bmcData.MACAddress, err = task.getBMCMacAddress(ctx)
		if err != nil {
			err := fmt.Errorf("unable to get BMC MAC address: %v", err)
			updateDeviceStatus(ctx, netBox, deviceData, dcim.BMEnrollmentFailed, err.Error())
			return nil, err
		}
	}

	return task, nil
}

func (t *EnrollmentTask) Run(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.Run")
	log.Info("Enrollment task started")

	if err := t.enroll(ctx); err != nil {
		log.Error(err, "unable to enroll device")
		if err := t.updateDeviceStatus(ctx, dcim.BMEnrollmentFailed, fmt.Sprintf("Enrollment failed: %v", err)); err != nil {
			return fmt.Errorf("unable to update NetBox device status: %v", err)
		}
		return fmt.Errorf("unable to enroll device: %v", err)
	}

	if err := t.updateDeviceStatus(ctx, dcim.BMEnrolled, "Enrollment is complete"); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}

	log.Info("Enrollment task completed!!")
	return nil
}

func (t *EnrollmentTask) enroll(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.Enroll")

	if err := t.updateDeviceStatus(ctx, dcim.BMEnrolling, "Enrollment is in progress"); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}

	// Enforce minimum firmware requirements
	hardwareHasMinimumSupportedFirmware, err := t.checkMinFWVersions(ctx)
	if err != nil {
		return fmt.Errorf("unable to verify hardware has minimum supported firmware. %v", err)
	}
	if !hardwareHasMinimumSupportedFirmware {
		return fmt.Errorf("hardware does not meet minimum supported firmware required")
	}

	// Check if host exists in BMH
	if err := t.checkIfHostExists(ctx); err != nil {
		return fmt.Errorf("cannot enroll due to this error: %v", err)
	}
	// TODO support account service via IPMI
	if t.bmc.IsVirtual() || t.bmc.GetHwType() == bmc.Gaudi2Wiwynn {
		log.Info("Can not change accounts on virtual BMC, ignoring BMC creds update.")
	} else {
		newUsername, newPassword, err := t.generateBMCCredentials(ctx)
		if err != nil {
			return fmt.Errorf("unable to generate new BMC creds: %v", err)
		}

		if err = t.storeUserBMCCredentialsInVault(ctx, newUsername, newPassword); err != nil {
			return fmt.Errorf("unable to write to Vault client: %v", err)
		}

		if createErr := t.updateBMCCredentialsInRedfish(ctx, newUsername, newPassword); createErr != nil {
			log.Info("Failed to update BMC Credentials, removing entry from Vault", "error", createErr)
			if err = t.deleteBMCCredentialsFromVault(ctx, bmcUserSecretsPrefix); err != nil {
				return fmt.Errorf("unable to write to Vault client: %v", err)
			}
			return fmt.Errorf("unable to update BMC creds: %v", createErr)
		}

		// recreate a new BMC interface using new credentials
		t.bmc, err = bmc.New(
			&mygofish.MyGoFishManager{},
			&bmc.Config{
				URL:      t.bmcData.URL,
				Username: t.bmcData.Username,
				Password: t.bmcData.Password,
			})
		if err != nil {
			return fmt.Errorf("unable to recreate BMC Interface with new user and password:  %v", err)
		}
		// Make sure the BMC updated to the new credentials
		// Testing with an invalid password resulting in updating the
		// username, but not the password
		if err = t.verifyBMCCredentials(ctx); err != nil {
			return fmt.Errorf("unable to reverify BMC creds after attempted change: %v", err)
		}

		// find Boot mac address here so we can identify the BIOS pxe boot regex
		_, err = t.getBootMacAddress(ctx)
		if err != nil {
			return fmt.Errorf("unable to get Boot MAC Address: %v", err)
		}

		if err = t.bmc.SanitizeBMCBootOrder(ctx); err != nil {
			return fmt.Errorf("unable to update BMC Boot Order: %v", err)
		}

		if err = t.bmc.ConfigureNTP(ctx); err != nil {
			return fmt.Errorf("unable to update BMC NTP: %v", err)
		}

		if err = t.bmc.VerifyPlatformFirmwareResilience(ctx); err != nil {
			return fmt.Errorf("unable to verify the Platform Firmware Resilience: %v", err)
		}
	}

	if t.bmHostNamespace, err = t.getBareMetalHostNamespace(ctx); err != nil {
		return fmt.Errorf("unable to get the namespace for baremetalhost: %v", err)
	}

	// Make sure KCS is enabled
	// KCS needs to be restricted in scheduling before we handed to customer
	if t.bmc.IsVirtual() {
		log.Info("Cannot Enable KCS on virtual BMC, skipping.")
	} else {
		if err = t.EnableKCS(ctx); err != nil {
			return fmt.Errorf("KCS Policy Control Mode was not set: %v", err)
		}
		log.Info("KCS is set to Provisioning.")
	}

	// Set Fan mode of SMC Gaudi3 to FullSpeed
	if t.bmc.GetHwType() == bmc.Smc822GANGR3IN001 {
		if err = t.bmc.SetFanSpeed(ctx); err != nil {
			return fmt.Errorf("failed to set fan mode: %v", err)
		}
	}

	if err = t.registerBareMetalHost(ctx); err != nil {
		return fmt.Errorf("unable to register baremetal: %v", err)
	}
	return nil
}

const (
	minFwConfigMapName    = "ironic-fw-update"
	BIOSKEY               = "BIOS"
	CURRENTBIOSFORVIRTUAL = "1.0.0" // current bios version reported for virtual machines
	BMCKEY                = "BMC"
	CURRENTBMCFORVIRTUAL  = "1.0.0" // current bmc version reported for virtual machines
)

// CurrentFirmwareVersions returns map of firmware present on bmhost
func (t *EnrollmentTask) CurrentFirmwareVersions() (map[string]string, error) {

	firmwareInfo := make(map[string]string)

	if t.bmc.IsVirtual() { // Virtual devices are treated special
		firmwareInfo[BIOSKEY] = CURRENTBIOSFORVIRTUAL
		firmwareInfo[BMCKEY] = CURRENTBMCFORVIRTUAL
		return firmwareInfo, nil
	}

	bmcClient := t.bmc.GetClient()
	if bmcClient == nil {
		return nil, fmt.Errorf("unable to get BMC client")
	}

	service := bmcClient.GetService()
	if service == nil {
		return nil, fmt.Errorf("unable to get BMC client service")
	}

	updateService, err := service.UpdateService()
	if err != nil || updateService == nil {
		return nil, fmt.Errorf("unable to get UpdateService: %v", err)
	}

	inventories, err := updateService.FirmwareInventories()
	if err != nil || inventories == nil {
		return nil, fmt.Errorf("unable to get FirmwareInventories: %v", err)
	}

	for _, inventory := range inventories {
		firmwareInfo[strings.ToUpper(inventory.ID)] = inventory.Version
	}

	return firmwareInfo, nil

}

type FwVersion struct {
	BIOS string `json:"BIOS"`
	BMC  string `json:"BMC"`
}

type MinFirmwareVersion struct {
	HwTypeMinFirmwareVersions map[string]FwVersion `json:"hwTypeMinFirmware"`
}

// MinFwVersionSupported returns  minimum supported firmware on the namespace for the bmhost
func (t *EnrollmentTask) MinFwVersionSupported(ctx context.Context) (map[string]string, error) {
	// Returns empty list if host hardware type is not found in the list of supported firmware

	currentHwType := t.bmc.GetHwType().String()

	if t.bmHostNamespace == nil {
		nameSpace, err := t.getBareMetalHostNamespace(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get the namespace for baremetalhost: %v", err)
		}
		t.bmHostNamespace = nameSpace
	}

	configMap, err := t.clientSet.CoreV1().ConfigMaps(t.bmHostNamespace.Name).Get(ctx, minFwConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	jsondata := configMap.Data[minFwConfigMapName+".json"]
	var data MinFirmwareVersion
	err = json.Unmarshal([]byte(jsondata), &data)
	if err != nil {
		return nil, err
	}

	minFwSupported := make(map[string]string)
	for supportedHwType, fwVersions := range data.HwTypeMinFirmwareVersions {
		if strings.EqualFold(supportedHwType, currentHwType) {
			minFwSupported[BIOSKEY] = fwVersions.BIOS
			minFwSupported[BMCKEY] = fwVersions.BMC
		}
	}

	return minFwSupported, nil
}

func removeLeadingZeros(version string) string {
	vNumbers := strings.Split(version, ".")
	versionArray := make([]string, len(vNumbers))
	for i, v := range vNumbers {
		vInt, err := strconv.Atoi(v)
		if err != nil {
			vInt = 0
		}
		versionArray[i] = strconv.Itoa(vInt)

	}
	return strings.Join(versionArray, ".")
}

func convertToSemver(value string) string {
	// semver package expects the version  strings to be of the form v##.##.##
	// Leading zeros are not allowed and only take the numbers at end of version string
	// for example "Date 10/12/23 01.04.05" should result in v1.4.5
	n := value[strings.LastIndex(value, " ")+1:]
	n = strings.TrimLeft(n, "vV")
	n = removeLeadingZeros(n)
	return "v" + n
}

// oldestVersion returns the oldest version of the two versions strings passed in
func oldestVersion(v, w string) string {

	vSemver := convertToSemver(v)
	wSemver := convertToSemver(w)

	result := semver.Compare(vSemver, wSemver)
	if result < 0 {
		return v
	}
	return w
}

func meetsMiniumFWRequirements(ctx context.Context, firmwarePresent, minimumFirmwareSupported map[string]string) bool {
	// Only return false if it is confirmed that at least one present firmware is below minimum
	//
	// Returns true if all present firmware listed under supported firmware are equal or higher
	// Returns true if none of the present firmware are listed under supported firmware
	//
	// Returns false if at least one of the present firmware is below the listed supported firmware version
	// Returns false if the list of present firmware is empty

	logger := log.FromContext(ctx).WithName("EnrollmentTask.meetsMiniumFWRequirements")
	if len(firmwarePresent) < 1 {
		logger.Info("list of present firmware is empty")
		return false
	}

	for fwID, minVersion := range minimumFirmwareSupported {
		if fwVersion, found := firmwarePresent[fwID]; found {
			if oldestVersion(fwVersion, minVersion) != minVersion { // the minVersion should be the oldest version
				logger.Info("firmware version below minimum supported version ", fwID, fwVersion, "minimum supported", minVersion)
				return false
			}
		} else {
			logger.Info("could not find present version for %v", fwID)
		}
	}
	return true
}

func (t *EnrollmentTask) checkMinFWVersions(ctx context.Context) (bool, error) {
	// Returns true if no minimum requirements found for hardware type
	// Returns true if present firmware meets minimum requirements
	// Returns false if present firmware does not meet minimum requirements
	// Returns false on all errors

	logger := log.FromContext(ctx).WithName("EnrollmentTask.checkMinFWVersions")
	logger.Info("Checking minimum firmware requirements", "device", t.deviceData.Name)

	minimumFirmwareSupported, err := t.MinFwVersionSupported(ctx) // Get a list of hardware types with their min firmware
	if err != nil {
		return false, err
	}
	if len(minimumFirmwareSupported) < 1 {
		nameSpace := &corev1.Namespace{}
		if t.bmHostNamespace == nil {
			nameSpace, err = t.getBareMetalHostNamespace(ctx)
			if err != nil {
				return false, fmt.Errorf("unable to get the namespace for baremetalhost: %v", err)
			}
		} else {
			nameSpace = t.bmHostNamespace
		}
		if err != nil {
			return false, err
		}
		logger.Info("No minimum firmware requirements found for hardware", "hardware type", t.bmc.GetHwType().String(), "namespace", nameSpace.Name)
		return true, nil
	}

	firmwarePresent, err := t.CurrentFirmwareVersions() // Get a list of current firmwares in bmhost
	if err != nil {
		return false, err
	}

	if meetsMiniumFWRequirements(ctx, firmwarePresent, minimumFirmwareSupported) {
		logger.Info("device meets minimum firmware versions ", "device", t.deviceData.Name, "versions present", firmwarePresent)
		return true, nil
	}

	logger.Info("device does not meet firmware versions minimums",
		"device", t.deviceData.Name, "versions present", firmwarePresent,
		"minimum version supported", minimumFirmwareSupported)
	return false, nil

}

func (t *EnrollmentTask) updateDeviceStatus(ctx context.Context, status dcim.BMEnrollmentStatus, comment string) error {
	return updateDeviceStatus(ctx, t.netBox, t.deviceData, status, comment)
}

func (t *EnrollmentTask) getBootMacAddress(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.getBootMacAddress")
	log.Info("Getting Boot MAC Address")

	hostMAC, err := t.bmc.GetHostMACAddress(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get the host MAC address: %v", err)
	}
	if hostMAC == "" {
		// Try to extract it from Netbox
		hostMAC, err = t.netBox.GetBMCMACAddress(ctx, t.deviceData.Name, HostInterfaceName)
		if err != nil {
			return "", fmt.Errorf("unable to get Host MACAddress with error:  %v", err)
		}
		if hostMAC == "" {
			return "", fmt.Errorf("host MACAddress in Netbox is empty")
		}
	}

	normalizedMAC := helper.NormalizeMACAddress(hostMAC)
	log.Info("Normalizing MAC Address", "Original", hostMAC, "Normalized", normalizedMAC)

	return normalizedMAC, nil
}

func (t *EnrollmentTask) getBMCMacAddress(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.getBMCMacAddress")
	log.Info("Getting BMC MAC Address")

	bmcMACAddress, err := t.netBox.GetBMCMACAddress(ctx, t.deviceData.Name, BMCInterfaceName)
	if err != nil {
		return "", fmt.Errorf("unable to get BMC MACAddress with error:  %v", err)
	}

	return bmcMACAddress, nil
}

func (t *EnrollmentTask) verifyBMCCredentials(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.verifyBMCCredentials")
	log.Info("Verifying BMC default credentials")

	// Retrieve the service root
	service := t.bmc.GetClient().GetService()

	//get a list of systems
	systems, err := service.Systems()
	if err != nil {
		return fmt.Errorf("failed to get Systems for BMC URL '%s': '%s'", t.bmcData.URL, err)
	}

	//iterate over the systems and print their details
	for _, system := range systems {
		log.Info("BMC", "System ID", system.ID(), "Name", system.Name(), "Power State", system.PowerState())
	}

	return nil
}

func (t *EnrollmentTask) generateBMCCredentials(ctx context.Context) (string, string, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.generateBMCCredentials")
	log.Info("Generating BMC credentials")

	newBMCUsername := helper.GetEnv(envBMCEnrollUsername, defaultBMCEnrollUsername)
	log.Info("New Username", "newBMCUsername", newBMCUsername)

	// Customize the list of symbols of password generator.
	// Dell Gaudi doesn't following symbols in the password: &*_\\\"<>./
	passwordGenerator, err := password.NewGenerator(&password.GeneratorInput{
		// Supported symbols
		Symbols: "@%{}|[]?(),#+-=^~`",
	})
	if err != nil {
		return "", "", fmt.Errorf("could not generate a random password generator: %v", err)
	}
	// Generate a password that is 16 characters long with 5 digits, 5 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	randomBMCPassword, err := passwordGenerator.Generate(16, 5, 5, false, false)
	if err != nil {
		return "", "", fmt.Errorf("could not generate a random password: %v", err)
	}
	return newBMCUsername, randomBMCPassword, nil
}

func (t *EnrollmentTask) getBMCSecretPath(prefix string) (string, error) {
	if len(t.bmcData.MACAddress) < 12 { // Could be ff:ff:ff:ff:ff:ff, ff-ff-ff-ff-ff-ff, ffffffffffff
		return "", fmt.Errorf("BMC MAC address too short: %s", t.bmcData.MACAddress)
	}
	return fmt.Sprintf("%s/%s%s/%s", t.deviceData.Region, bmcDeploymentSecretsPath, t.bmcData.MACAddress, prefix), nil
}

func (t *EnrollmentTask) storeBMCCredentialsInVault(ctx context.Context, prefix string, secretData map[string]interface{}) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.storeBMCCredentialsInVault")
	log.Info("Storing new BMC deployed credentials")

	path, err := t.getBMCSecretPath(prefix)
	if err != nil {
		return fmt.Errorf("unable to get BMC '%s': %v", t.bmcData.MACAddress, err)
	}
	log.Info("Storing credentials for MAC address", "bmcMACAddress", t.bmcData.MACAddress, "path", path)

	secret, err := t.vault.PutBMCSecrets(ctx, path, secretData)
	if err != nil {
		return fmt.Errorf("failed to write secrets from vault: %v", err)
	}

	log.Info("Returned data from write", "secret", secret)

	return nil
}

func (t *EnrollmentTask) storeUserBMCCredentialsInVault(ctx context.Context, newUsername, newPassword string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.storeUserBMCCredentialsInVault")
	log.Info("Storing new User BMC deployed credentials")

	secretData := map[string]interface{}{
		"username": newUsername,
		"password": newPassword,
	}
	if err := t.storeBMCCredentialsInVault(ctx, bmcUserSecretsPrefix, secretData); err != nil {
		return fmt.Errorf("unable to write to Vault client: %v", err)
	}

	return nil
}

func (t *EnrollmentTask) updateBMCCredentialsInRedfish(ctx context.Context, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.updateBMCCredentialsInRedfish")
	log.Info("Updating BMC credentials")

	updateErr := t.bmc.UpdateAccount(ctx, newUserName, newPassword)
	if updateErr == bmc.ErrAccountNotFound {
		if t.bmc.IsVirtual() {
			log.Info("Skip creating an admin account on virtual BMC")
			return nil
		}
		if createdErr := t.bmc.CreateAccount(ctx, newUserName, newPassword); createdErr != nil {
			return fmt.Errorf("unable to create Admin account's credentials: %v", createdErr)
		}
	} else if updateErr != nil {
		return fmt.Errorf("unable to update Admin account's credentials: %v", updateErr)
	}

	t.bmcData.Username = newUserName
	t.bmcData.Password = newPassword

	return nil
}

func (t *EnrollmentTask) deleteBMCCredentialsFromVault(ctx context.Context, prefix string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.deleteBMCCredentialsFromVault")
	log.Info("Deleting new BMC deployed credentials")

	path, err := t.getBMCSecretPath(prefix)
	if err != nil {
		return fmt.Errorf("unable to get BMC '%s': %v", t.bmcData.MACAddress, err)
	}

	log.Info("Deleting credentials for MAC address", "bmcMACAddress", t.bmcData.MACAddress, "path", path)
	err = t.vault.DeleteBMCSecrets(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secrets from vault: %v", err)
	}

	log.Info("Deleted secrete from path", "path", path)

	return nil
}

func (t *EnrollmentTask) GetExtraEthernetMacAnnotations(ctx context.Context, annotations map[string]string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.GetExtraEthernetMacAnnotations")
	log.Info("getting extra ethernet information")
	sshManager := &myssh.MySSHManager{}
	hahabna, err := ipacmd.NewIpaCmdHelper(ctx, t.vault, sshManager, t.bmhIpAddress, t.deviceData.Region)
	if err != nil {
		return fmt.Errorf("unable to initialize IpmCmd helper: %v", err)
	}
	habanabuses, err := hahabna.HabanaEthernetBusInfo(ctx)
	if err != nil {
		return fmt.Errorf("unable to get Habana Ethernet Bus Info: %v", err)
	}
	if err = hahabna.HabanaEthernetMacAddress(ctx, habanabuses); err != nil {
		return fmt.Errorf("unable to get Habana Ethernet MacAddresses: %v", err)
	}
	for _, item := range habanabuses {
		if len(item.MacAddresses) > 0 {
			for i, _ := range item.MacAddresses {
				annotations[fmt.Sprintf("%s/gpu-eth%s-%d", GPUAnnotationPrefix, item.ModuleID, i)] = item.MacAddresses[i]
			}
		}
	}
	return nil
}

func (t *EnrollmentTask) GetGPUIPAddressesAnnotations(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, annotations map[string]string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.GetGPUIPAddressesAnnotations")
	log.Info("getting GPU IP addresses information")

	nicInformation := host.Status.HardwareDetails.NIC
	if len(nicInformation) < 1 {
		return fmt.Errorf("failed to find nic information in host %s", host.Name)
	}

	var assignedIpAddress int
	for i := range nicInformation {
		if nicInformation[i].Name != "" {
			if nicInformation[i].MAC[:8] == GaudiExternalPortMacPrefix || nicInformation[i].MAC[:8] == GaudiExternalPortMacPrefix1 {
				switchName := strings.Split(nicInformation[i].LLDP.SwitchSystemName, ".")[0]
				switchInterface := nicInformation[i].LLDP.SwitchPortId
				switchIpAddress, err := t.netBox.GetDeviceSwitchIPAddress(ctx, t.deviceData.Name, switchName, switchInterface)
				if err != nil {
					return fmt.Errorf("unable to fetch GPU IP address: %v", err)
				}

				macAddress := util.NormalizeMACAddress(nicInformation[i].MAC)
				ipAddr, subnet, err := net.ParseCIDR(switchIpAddress)
				if err != nil {
					return fmt.Errorf("unable to parse switch IP CIDR: %v", err)
				}

				// subtract 1 from switch IP Address to obtain GPU IP Address
				gpuIpAddress, err := convertSwitchIpToGpuIp(ipAddr, subnet)
				if err != nil {
					return fmt.Errorf("unable to calculate GPU IP address from switch IP address: %v", err)
				}

				gpuIpAddressSubnet := strings.Split(subnet.String(), "/")[1]
				for annotation, mac := range annotations {
					if mac == macAddress {
						annotations[GPUIPsAnnotationPrefix+"/"+strings.Split(annotation, "/")[1]] = gpuIpAddress.String() + "/" + gpuIpAddressSubnet
						assignedIpAddress++
					}
				}
			}
		}
	}

	if assignedIpAddress != TotalGaudiExternalPorts {
		return fmt.Errorf("unable to add GPU IPs annotations; MAC Address information is missing")
	}

	return nil
}

func (t *EnrollmentTask) getBIOSPasswordFromVault(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.getBIOSPasswordFromVault")
	log.Info("Check for existing BIOS password")

	path, err := t.getBMCSecretPath(bmcBIOSSecretsPrefix)
	if err != nil {
		return "", fmt.Errorf("unable to get BMC Secret path '%s': %v", t.bmcData.MACAddress, err)
	}
	log.Info("requesting BIOS Password for MAC address", "bmcMACAddress", t.bmcData.MACAddress, "path", path)

	// Request the password from Vault
	biosPassword, err := t.vault.GetBMCBIOSPassword(ctx, path)
	if err != nil {
		return "", fmt.Errorf("unable to read from Vault client: %v", err)
	}

	return biosPassword, nil
}

func (t *EnrollmentTask) generateBIOSPassword(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.generateBIOSPassword")
	log.Info("Generating BIOS password")

	// BIOS/syscfg password rules
	// The password must have a length of 8–14 characters.
	// The password can have alphanumeric characters (a-z, A-Z, 0–9) and the following special characters:
	// ! @ # $ % ^ *( ) - _ + = ? '
	// Use two double quotes (“”) to represent a null password.
	passwordGenerator, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "!@#$%^*()-_+=?",
	})
	if err != nil {
		return "", fmt.Errorf("Could not create password generator: %v", err)
	}

	randomBIOSPassword, err := passwordGenerator.Generate(14, 3, 2, false, false)
	if err != nil {
		return "", fmt.Errorf("Could not generate a random password: %v", err)
	}

	return randomBIOSPassword, nil
}

func (t *EnrollmentTask) storeBIOSPasswordInVault(ctx context.Context, newPassword string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.storeBIOSPasswordInVault")
	log.Info("Storing new BIOS password")

	secretData := map[string]interface{}{
		"password": newPassword,
	}
	if err := t.storeBMCCredentialsInVault(ctx, bmcBIOSSecretsPrefix, secretData); err != nil {
		return fmt.Errorf("unable to write to Vault client: %v", err)
	}

	return nil
}

func (t *EnrollmentTask) EnableKCS(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.ipmiEnableKCS")
	log.Info("Enabling KCS")
	err := t.bmc.EnableKCS(ctx)
	if err != nil {
		return fmt.Errorf("unable to Enable KCS: %v", err)
	}
	return nil
}

func (t *EnrollmentTask) getBareMetalHostNamespace(ctx context.Context) (*corev1.Namespace, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.getBareMetalHostNamespace")
	log.Info("Finding BareMetalHost namespace")

	selector := fmt.Sprintf("%s=true", Metal3NamespaceSelectorKey)
	namespaces, err := t.clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("unable to list metal3 namespaces: %v", err)
	}
	if len(namespaces.Items) == 0 {
		return nil, fmt.Errorf("no metal3 namespace found")
	}

	// use device's namespace if found
	deviceNamespace, err := t.netBox.GetDeviceNamespace(ctx, t.deviceData.Name)
	if err != nil {
		return nil, fmt.Errorf("unable to get device namespace: %v", err)
	}
	if deviceNamespace != "" {
		for _, ns := range namespaces.Items {
			if ns.Name == deviceNamespace {
				log.Info(fmt.Sprintf("Using %q namespace for the new host", ns.Name))
				return &ns, nil
			}
		}
		return nil, fmt.Errorf("unable to find %q namespace", deviceNamespace)
	}

	// find the namespace with the least number of baremetalhosts
	targetNamespace := &namespaces.Items[0]
	lowest := math.MaxInt64

	for i, ns := range namespaces.Items {
		hostList, err := t.dynamicClient.Resource(bmHostGVR).Namespace(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to list baremetalhosts: %v", err)
		}

		current := len(hostList.Items)

		log.Info("Metal3 Namespace", "name", ns.Name, "host found", current)

		if current < lowest {
			lowest = current
			targetNamespace = &namespaces.Items[i]
		}
	}

	log.Info(fmt.Sprintf("Using %q namespace for the new host", targetNamespace.Name))
	return targetNamespace, nil
}

func (t *EnrollmentTask) checkIfHostExists(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.checkIfHostExists")
	log.Info("checking if host exist")

	selector := fmt.Sprintf("%s=true", Metal3NamespaceSelectorKey)
	namespaces, err := t.clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("unable to list metal3 namespaces: %v", err)
	}
	if len(namespaces.Items) == 0 {
		return fmt.Errorf("no metal3 namespace found")
	}

	for _, ns := range namespaces.Items {
		hostList, err := t.dynamicClient.Resource(bmHostGVR).Namespace(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("unable to list baremetalhosts: %v", err)
		}
		// check host name
		for _, host := range hostList.Items {
			if host.GetName() == t.deviceData.Name {
				return fmt.Errorf("host is already enrolled in metal3 inventory")
			}
		}
	}
	return nil
}

func (t *EnrollmentTask) createPxeRecord(ctx context.Context, ironicIp string, mac string) error {
	// Using only in virtual environment in kind
	dhcpProxyUrl, present := os.LookupEnv(DhcpProxyUrlEnvVar)
	if present {
		_, err := http.Post(dhcpProxyUrl+"/record?mac="+mac+"&ironicip="+ironicIp, "Application/json", nil)
		if err != nil {
			return err
		}
	} else { // check if men and mice url is present
		if t.menAndMice != nil {
			var ipRange *ddi.Range
			ipRange, err := t.menAndMice.GetRangeByName(ctx, t.deviceData.Rack, menAndMiceProvisioningType)
			if err != nil {
				return fmt.Errorf("failed to read Range: %v of Rack %s", err, t.deviceData.Rack)
			}
			if len(ipRange.DhcpScopes) < 1 {
				return fmt.Errorf("failed to find dhcp Scopes in Range: %s", ipRange.Ref)
			}
			var dhcpReservation *ddi.DhcpReservation
			var dhcpReservationRef string
			filename := fmt.Sprintf("http://%s:%s/%s", ironicIp, IronicHttpPortNb, IPXEProfileName)
			tftpServer := helper.GetEnv(TftpServerIPEnvVar, ironicIp)
			dhcpReservation, err = t.menAndMice.GetDhcpReservationsByMacAddress(ctx, ipRange.DhcpScopes[0].Ref, mac)
			if err != nil {
				//get available IP address
				ipaddress, err := t.menAndMice.GetAvailableIp(ctx, ipRange)
				if err != nil {
					return fmt.Errorf("failed to get an IP %v", err)
				}
				// Create a reservation
				dhcpReservationRef, err = t.menAndMice.SetDhcpReservationByScope(ctx, ipRange, mac, ipaddress, t.deviceData.Name, filename, tftpServer)
				if err != nil {
					return fmt.Errorf("failed Set DhcpReservation %v", err)
				}
			} else {
				dhcpReservationRef = dhcpReservation.Ref
			}
			err = t.menAndMice.UpdateDhcpReservationOptions(ctx, ipRange, t.deviceData.Name, filename, tftpServer, dhcpReservationRef, IPXEBinarayName)
			if err != nil {
				return fmt.Errorf("failed Update DhcpReservation ByMacAddress %v", err)
			}
		}
	}
	return nil
}

// newBareMetalHostSecret returns a new secret that contain BMC credentials for BareMetalHost
func (t *EnrollmentTask) newBareMetalHostSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-bmc-secret", t.deviceData.Name),
			Namespace: t.bmHostNamespace.Name,
		},
		Data: map[string][]byte{
			"username": []byte(t.bmcData.Username),
			"password": []byte(t.bmcData.Password),
		},
	}
}

// createBareMetalHostSecret creates a secret that is owned by BareMetalHost
func (t *EnrollmentTask) createBareMetalHostSecret(ctx context.Context, secret *corev1.Secret, hostName string, hostUID types.UID) (*corev1.Secret, error) {
	secret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: baremetalv1alpha1.GroupVersion.String(),
			Kind:       "BareMetalHost",
			Name:       hostName,
			UID:        hostUID,
		},
	}
	newSecret, err := t.clientSet.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create BMC Secret %q: %v", secret.Name, err)
	}

	return newSecret, nil
}

// newBareMetalHost returns a new BareMetalHost with default spec
func (t *EnrollmentTask) newBareMetalHost() *baremetalv1alpha1.BareMetalHost {
	rootDeviceHints := &baremetalv1alpha1.RootDeviceHints{}
	bootMode := baremetalv1alpha1.UEFI
	if t.bmc.IsVirtual() {
		rootDeviceHints = &baremetalv1alpha1.RootDeviceHints{
			DeviceName: "/dev/vda",
		}
		bootMode = baremetalv1alpha1.Legacy
	}

	bmHost := &baremetalv1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "metal3.io/v1alpha1",
			Kind:       "BareMetalHost",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.deviceData.Name,
			Namespace: t.bmHostNamespace.Name,
		},
		Spec: baremetalv1alpha1.BareMetalHostSpec{
			Online:         true,
			BootMode:       bootMode,
			BootMACAddress: "",
			BMC: baremetalv1alpha1.BMCDetails{
				Address:                        "",
				CredentialsName:                "",
				DisableCertificateVerification: true,
			},
			RootDeviceHints: rootDeviceHints,
		},
	}

	return bmHost
}

// createBareMetalHost create a BareMetalHost
func (t *EnrollmentTask) createBareMetalHost(ctx context.Context, host *baremetalv1alpha1.BareMetalHost) (*baremetalv1alpha1.BareMetalHost, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(host)
	if err != nil {
		return nil, fmt.Errorf("unable to convert %s to unstructured object: %v", host.Name, err)
	}

	newHost, err := t.dynamicClient.
		Resource(bmHostGVR).
		Namespace(host.Namespace).
		Create(ctx, &unstructured.Unstructured{Object: obj}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create BareMetalHost %q: %v", host.Name, err)
	}

	bmh := &baremetalv1alpha1.BareMetalHost{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(newHost.UnstructuredContent(), bmh); err != nil {
		return nil, fmt.Errorf("unable to decode BareMetalHost object")
	}

	return bmh, nil
}

func (t *EnrollmentTask) registerBareMetalHost(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.registerBareMetalHost")
	log.Info("Registering BareMetalHost with Metal3", "name", t.deviceData.Name, "namespace", t.bmHostNamespace)

	// get host's boot MAC address
	bootMACAddress, err := t.getBootMacAddress(ctx)
	if err != nil {
		return fmt.Errorf("unable to get Boot MAC Address: %v", err)
	}

	// create pxe record
	ironicIP, exists := t.bmHostNamespace.ObjectMeta.Labels[metal3NamespaceIronicIPKey]
	if !exists {
		return fmt.Errorf("unable to extract ironic IP from metal3 namespace")
	}
	err = t.createPxeRecord(ctx, ironicIP, bootMACAddress)
	if err != nil {
		return fmt.Errorf("unable to create mac record in dhcp server: %v", err)
	}

	// create new secret for BMC access
	bmcSecret := t.newBareMetalHostSecret()

	// get host's BMC address
	bmcAddress, err := t.bmc.GetHostBMCAddress()
	if err != nil {
		return fmt.Errorf("unable to get host's BMC address: %v", err)
	}

	// create a new BareMetalHost
	bmHost := t.newBareMetalHost()
	bmHost.Spec.BootMACAddress = bootMACAddress
	bmHost.Spec.BMC.Address = bmcAddress
	bmHost.Spec.BMC.CredentialsName = bmcSecret.Name
	// Add testing image
	bmHost.Spec.Image = &baremetalv1alpha1.Image{
		URL:      fmt.Sprintf("http://%s:%s/images/cirros-disk.img", ironicIP, IronicHttpPortNb),
		Checksum: fmt.Sprintf("http://%s:%s/images/cirros-disk.img.md5sum", ironicIP, IronicHttpPortNb),
	}

	log.Info("Creating BareMetalHost", "name", bmHost.Name, "bmcAddress", bmHost.Spec.BMC.Address)
	newHost, err := t.createBareMetalHost(ctx, bmHost)
	if err != nil {
		return fmt.Errorf("unable to create BareMetalHost %q: %v", bmHost.Name, err)
	}

	// create BareMetalHost's secret
	log.Info("Creating BMC Secret", "name", bmcSecret.Name)
	if _, err = t.createBareMetalHostSecret(ctx, bmcSecret, newHost.GetName(), newHost.GetUID()); err != nil {
		return fmt.Errorf("unable to create BMC Secret %q: %v", bmcSecret.Name, err)
	}
	log.Info("Device details", "Device", t.deviceData)
	if err := t.updateDeviceStatus(ctx, dcim.BMEnrolling, "Enrollment is in progress: testing host's ability to provision"); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}
	// find env variable for provisioning timeout
	provisioningTimeout, present := os.LookupEnv(helper.ProvisioningTimeoutVar)
	if !present {
		return fmt.Errorf("error finding the expected provisioning timeout variable")
	}
	deprovisioningTimeout, present := os.LookupEnv(helper.DeprovisionTimeoutVar)
	if !present {
		return fmt.Errorf("error finding the expected deprovisioning timeout variable")
	}
	provisioningTimeoutInt, err := strconv.Atoi(provisioningTimeout)
	if err != nil {
		return fmt.Errorf("failed to convert provsioning timeout to int %v", err)
	}
	deprovisioningTimeoutInt, err := strconv.Atoi(deprovisioningTimeout)
	if err != nil {
		return fmt.Errorf("failed to convert deprovsioning timeout to int %v", err)
	}
	provisionedHost, err := t.waitForAvailableHost(ctx, newHost, baremetalv1alpha1.StateProvisioned, int64(provisioningTimeoutInt))
	if err != nil {
		return fmt.Errorf("error waiting for host to be become provisioned: %v", err)
	}

	// trigger Deprovisioning
	patch := []byte(`{"spec":{"image": null}}`)

	_, err = t.dynamicClient.Resource(bmHostGVR).
		Namespace(provisionedHost.GetNamespace()).
		Patch(ctx, provisionedHost.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	if err := t.updateDeviceStatus(ctx, dcim.BMEnrolling, "Enrollment is in progress: waiting host to become available"); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}

	availableHost, err := t.waitForAvailableHost(ctx, provisionedHost, baremetalv1alpha1.StateAvailable, int64(deprovisioningTimeoutInt))
	if err != nil {
		return fmt.Errorf("error waiting for host to be become available: %v", err)
	}

	// find BMH IP
	nicInformation := availableHost.Status.HardwareDetails.NIC
	for i := range nicInformation {
		if helper.NormalizeMACAddress(nicInformation[i].MAC) == helper.NormalizeMACAddress(availableHost.Spec.BootMACAddress) {
			t.bmhIpAddress = nicInformation[i].IP
			break
		}
	}
	if t.bmhIpAddress == "" {
		return fmt.Errorf("failed to fetch Host IP from BMH %s", bmHost.Name)
	}
	log.Info("MacAddresses", "BootMacAddress", availableHost.Spec.BootMACAddress, "BMCMacAddress", t.bmcData.MACAddress)

	if err := t.updateDeviceStatus(ctx, dcim.BMEnrolling, "Enrollment is in progress: updating hardware specification labels"); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}

	var storageMac string
	var isStorageNicPresent bool
	storageIntfToMacMap := make(map[string]string)
	if !t.bmc.IsVirtual() {
		// Check for net1/0 which is the storage interface.
		isStorageNicPresent, storageMac, err = t.getStorageNetworkDetails(ctx, StorageInterfaceName1)
		if err != nil {
			return fmt.Errorf("failed to fetch Storage MAC of interface %s from Netbox %s, %v", StorageInterfaceName1, bmHost.Name, err)
		}
		storageIntfToMacMap[StorageInterfaceName1] = storageMac
	} else {
		for _, nic := range nicInformation {
			normalized_mac := helper.NormalizeMACAddress(nic.MAC)
			if normalized_mac != helper.NormalizeMACAddress(t.bmcData.MACAddress) &&
				normalized_mac != helper.NormalizeMACAddress(availableHost.Spec.BootMACAddress) {
				log.Info("Found storage nic", "nic", nic)
				isStorageNicPresent = true
				storageMac = normalized_mac
				break
			}
		}
		storageIntfToMacMap[StorageInterfaceName1] = storageMac
	}
	if isStorageNicPresent {
		if err := t.updateStorageLabel(ctx, availableHost, storageIntfToMacMap); err != nil {
			return fmt.Errorf("error observed while updating StorageLabel: %v", err)
		}
	} else {
		log.Info("No Storage NIC Present for BMH")
	}

	// Verify LLDP data for Boot MAC Address and external accelerator ports for Gaudi devices
	if t.bmc.GetHwType() != bmc.Virtual {
		if err := t.verifyHostLLDPData(ctx, availableHost); err != nil {
			return fmt.Errorf("unable to verify LLDP data for BareMetalHost: %v", err)
		}
	}

	if err := t.updateHostHardwareLabels(ctx, availableHost, isStorageNicPresent); err != nil {
		return fmt.Errorf("unable to update hardware labels: %v", err)
	}

	return nil
}

func (t *EnrollmentTask) verifyHostLLDPData(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.verifyHostLLDPData")
	log.Info("Verifying LLDP data for BareMetalHost")

	// Fetch NIC information
	nicInformation := bmh.Status.HardwareDetails.NIC
	if len(nicInformation) < 1 {
		return fmt.Errorf("failed to find NIC information for BareMetalHost %s", bmh.Name)
	}

	// Verify LLDP Data for boot MAC address
	bootMACAddress := bmh.Spec.BootMACAddress
	for i := range nicInformation {
		if nicInformation[i].Name != "" {
			if nicInformation[i].MAC == bootMACAddress {
				if reflect.ValueOf(nicInformation[i].LLDP).IsZero() {
					log.Info("failed to find LLDP data for Boot MAC address", "BootMACAddress", bootMACAddress, "NICs", nicInformation[i])
					return fmt.Errorf("failed to verify LLDP data for tenant Boot MAC address")
				}
			}
		}
	}

	log.Info("Successfully verified LLDP data for Boot MAC address", "BootMACAddress", bootMACAddress)

	// Verify LLDP data for accelerator external ports
	switch t.bmc.GetHwType() {
	case bmc.Gaudi2Smc, bmc.Gaudi2Wiwynn, bmc.Gaudi2Dell, bmc.Gaudi3Dell, bmc.Smc822GANGR3IN001:
		if err := t.verifyAcceleratorLLDPData(ctx, bmh); err != nil {
			return fmt.Errorf("failed to verify LLDP data for accelerator external ports %v: ", err)
		}
	default:
		break
	}

	return nil
}

func (t *EnrollmentTask) verifyAcceleratorLLDPData(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.verifyAcceleratorLLDPData")
	log.Info("Verifying LLDP data on accelerator external ports for BareMetalHost")

	// Fetch NIC information
	nicInformation := bmh.Status.HardwareDetails.NIC
	if len(nicInformation) < 1 {
		return fmt.Errorf("failed to find NIC information for BareMetalHost %s", bmh.Name)
	}

	verifiedExternalPorts := 0
	absentLLDPDataPorts := make(map[string]string)
	presentLLDPDataPorts := make(map[string]string)

	for i := range nicInformation {
		if nicInformation[i].Name != "" {
			if nicInformation[i].MAC[:8] == GaudiExternalPortMacPrefix || nicInformation[i].MAC[:8] == GaudiExternalPortMacPrefix1 {
				if reflect.ValueOf(nicInformation[i].LLDP).IsZero() {
					absentLLDPDataPorts[nicInformation[i].Name] = nicInformation[i].MAC
				} else {
					presentLLDPDataPorts[nicInformation[i].Name] = nicInformation[i].MAC
				}
				verifiedExternalPorts++
			}
		}
	}

	// 24 external ports must be present
	if verifiedExternalPorts != TotalGaudiExternalPorts {
		return fmt.Errorf("failed to verify external ports, some ports are missing. expected: %d, found: %d", TotalGaudiExternalPorts, verifiedExternalPorts)
	}

	log.Info("Found all external ports", "TotalGaudiExternalPorts", TotalGaudiExternalPorts, "verifiedExternalPorts", verifiedExternalPorts)

	// LLDP data must be absent for single nodes
	if t.deviceData.Cluster == "None" && len(presentLLDPDataPorts) > 0 {
		return fmt.Errorf("Unexpected: LLDP data present on %d external ports; presentLLDPDataPorts: %s. All accelerator ports must be down for single Gaudi node. LLDP data not expected on any external port.", len(presentLLDPDataPorts), presentLLDPDataPorts)
	}

	// LLDP data must be present for cluster nodes
	if t.deviceData.Cluster != "None" && len(absentLLDPDataPorts) > 0 {
		log.Info("")
		return fmt.Errorf("Unexpected: LLDP data absent on %d ports; absentLLDPDataPorts: %s. All accelerator ports must be up for clustered Gaudi node. LLDP data expected on all external ports.", len(absentLLDPDataPorts), absentLLDPDataPorts)
	}

	log.Info("Successfully verified LLDP data on all external accelerator ports")
	return nil
}

func (t *EnrollmentTask) getStorageNetworkDetails(ctx context.Context, storageInterfaceName string) (bool, string, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.getStorageNetworkDetails")
	// Fetch the Storage MAC from netbox
	log.Info("Fetching Storage Network from netbox")
	storageMacAddress, err := t.netBox.GetBMCMACAddress(ctx, t.deviceData.Name, storageInterfaceName)
	if err != nil {
		log.Info("Error observed while attempting to fetch storage interface", "error", err)
		// log the error and set the storage as false
		return false, "", nil
	}
	if storageMacAddress == "" {
		return false, "", fmt.Errorf("host MACAddress of interface %s in Netbox is empty", storageInterfaceName)
	}
	return true, storageMacAddress, nil
}

func (t *EnrollmentTask) deleteBmh(ctx context.Context, host *baremetalv1alpha1.BareMetalHost) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.deleteBmh")
	//start by skipping clean mode to complete this task faster
	patch := []byte(`{"spec":{"automatedCleaningMode": "disabled"}}`)
	_, err := t.dynamicClient.Resource(bmHostGVR).
		Namespace(host.GetNamespace()).
		Patch(ctx, host.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	if err = t.dynamicClient.
		Resource(bmHostGVR).
		Namespace(host.GetNamespace()).
		Delete(ctx, host.GetName(), metav1.DeleteOptions{}); err != nil {
		log.Error(err, "unable to delete the failed BareMetalHost")
	}
	return nil
}

func (t *EnrollmentTask) waitForAvailableHost(ctx context.Context, host *baremetalv1alpha1.BareMetalHost, desiredState baremetalv1alpha1.ProvisioningState, timeout int64) (*baremetalv1alpha1.BareMetalHost, error) {
	log := log.FromContext(ctx).WithName("EnrollmentTask.waitForAvailableHost")

	w, err := t.dynamicClient.
		Resource(bmHostGVR).
		Namespace(host.GetNamespace()).
		Watch(ctx, metav1.ListOptions{TimeoutSeconds: &timeout})
	if err != nil {
		return nil, err
	}
	defer w.Stop()
	log.Info("BareMetalHost", "Provisioning Desired State", desiredState)
	for {
		select {
		case <-time.After(time.Second * time.Duration(timeout)):
			log.Info("Cleaning up BareMetalHost")
			err = t.deleteBmh(ctx, host)
			if err != nil {
				log.Error(err, "failed to delete failed bmh")
			}
			return nil, fmt.Errorf("timeout waiting for host to become %s", desiredState)
		case event, ok := <-w.ResultChan():
			if !ok {
				// trigger Delete
				err = t.deleteBmh(ctx, host)
				if err != nil {
					log.Error(err, "failed to delete failed bmh")
				}
				return nil, fmt.Errorf("error watching for BareMetalHost events")
			}

			bmh := &baremetalv1alpha1.BareMetalHost{}

			switch receivedObj := event.Object.(type) {
			case *baremetalv1alpha1.BareMetalHost:
				bmh = receivedObj
			case *unstructured.Unstructured:
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(receivedObj.UnstructuredContent(), bmh); err != nil {
					return nil, fmt.Errorf("unable to decode BareMetalHost object")
				}
			default:
				continue
			}

			if bmh.GetName() != host.GetName() {
				continue
			}

			switch event.Type {
			case watch.Modified:
				log.Info("BareMetalHost update", " Provisioning Status", bmh.Status.Provisioning, "Operational Status", bmh.Status.OperationalStatus)
				if bmh.Status.Provisioning.State == desiredState {
					log.Info(fmt.Sprintf("Host has reached the desired state %q", desiredState))
					return bmh, nil
				}
				if bmh.Status.OperationalStatus == baremetalv1alpha1.OperationalStatusError {
					log.Info("BareMetalHost operational error", "type", bmh.Status.ErrorType, "message", bmh.Status.ErrorMessage)
					if bmh.Status.ErrorCount > 0 {
						log.Info("Cleaning up BareMetalHost")
						err = t.deleteBmh(ctx, host)
						if err != nil {
							log.Error(err, "failed to delete failed bmh")
						}
					}
					return nil, fmt.Errorf("error Type: %v, Error Message: %v", bmh.Status.ErrorType, bmh.Status.ErrorMessage)
				}
			case watch.Deleted:
				return nil, fmt.Errorf("host has been deleted")
			}
		}
	}
}

// Upate storage label for BMH
func (t *EnrollmentTask) updateStorageLabel(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost, storageIntfToMacMap map[string]string) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.updateStorageLabel")
	for intf, storageMacAddress := range storageIntfToMacMap {
		normalizedStorageMac := helper.NormalizeMACAddress(storageMacAddress)
		// Verify if storage MAC is different from boot MAC address
		if helper.NormalizeMACAddress(bmh.Spec.BootMACAddress) == normalizedStorageMac {
			return fmt.Errorf("storage MAC %v from netbox matches BootMACAddress", storageMacAddress)
		}
		nicInformation := bmh.Status.HardwareDetails.NIC
		var found bool
		// Verify if the storage nic is present in the BMH after inspection
		for _, nic := range nicInformation {
			normalized_mac := helper.NormalizeMACAddress(nic.MAC)
			if normalized_mac == normalizedStorageMac {
				found = true
			}
		}
		if !found {
			log.Info("Storage MAC in netbox is not present on the host", "storageMAC", storageMacAddress)
			return fmt.Errorf("storage MAC %v from netbox does not match BMH after inspection", storageMacAddress)
		}

		if bmh.Annotations == nil {
			bmh.Annotations = make(map[string]string)
		}
		intfIndex := strings.Replace(strings.Trim(intf, "net"), "/", "-", -1)
		bmh.Annotations[fmt.Sprintf("%s/eth%s", StorageMACAnnotationPrefix, intfIndex)] = storageMacAddress
	}
	return nil
}

// updateHostHardwareLabels updates host's labels with hardware details from Ironic inspection and BMC data.
func (t *EnrollmentTask) updateHostHardwareLabels(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost, isStorageNicPresent bool) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.updateHostHardwareLabels")
	log.Info("Updating host's hardware labels")

	if bmh.Status.HardwareDetails == nil {
		return fmt.Errorf("hardware details not found")
	}

	bmhMemorySize := (bmh.Status.HardwareDetails.RAMMebibytes) / (units.GiB / units.MiB)

	cpu, err := t.bmc.GetHostCPU(ctx)
	if err != nil {
		return fmt.Errorf("unable to get CPU information from BMC: %v", err)
	}

	gpuCount, gpuModel, err := t.bmc.GPUDiscovery(ctx)
	if err != nil {
		return fmt.Errorf("unable to get GPU information from BMC: %v", err)
	}

	hbmMode, err := t.bmc.HBMDiscovery(ctx)
	if err != nil {
		return fmt.Errorf("unable to get HBM information from BMC: %v", err)
	}

	hostLabels := bmh.GetLabels()
	if hostLabels == nil {
		hostLabels = make(map[string]string)
	}
	hostLabels[CPUIDLabel] = strings.ToLower(cpu.CPUID)
	hostLabels[CPUCountLabel] = strconv.Itoa(bmh.Status.HardwareDetails.CPU.Count)
	hostLabels[CPUSocketsLabel] = strconv.Itoa(cpu.Sockets)
	hostLabels[CPUCoresLabel] = strconv.Itoa(cpu.Cores)
	hostLabels[CPUThreadsLabel] = strconv.Itoa(cpu.Threads)
	hostLabels[CPUManufacturerLabel] = formatLabelValue(cpu.Manufacturer)
	hostLabels[GPUModelNameLabel] = strings.ToLower(gpuModel)
	hostLabels[GPUCountLabel] = strconv.Itoa(gpuCount)
	hostLabels[HBMModeLabel] = strings.ToLower(hbmMode)
	hostLabels[MemorySizeLabel] = strconv.Itoa(bmhMemorySize) + "Gi"
	hostLabels[NetworkModeLabel] = ""

	assignInstanceTypeLabelFailed := false
	var assignInstanceTypeErr error
	// function to assign InstanceTypeLabel to the host as a part of the enrollment
	if err := t.assignInstanceTypeLabel(ctx, hostLabels, isStorageNicPresent); err != nil {
		log.Error(err, "failed to assign InstanceTypeLabel to BareMetalHost")
		assignInstanceTypeLabelFailed = true
		assignInstanceTypeErr = err
	}

	if !assignInstanceTypeLabelFailed {
		// trigger validation
		hostLabels[ReadyToTestLabel] = "true"
		// set instance group ID and cluster size
		if t.deviceData.Cluster != "None" {
			hostLabels[ClusterGroupID] = t.deviceData.Cluster
			// Set cluster size if cluster name is not "None"
			siteName := os.Getenv(dcim.AvailabilityZoneEnvVar)
			if siteName == "" {
				return fmt.Errorf("failed to get the availability zone(site name) of the device")
			}
			clusterSize, err := t.netBox.GetClusterSize(ctx, t.deviceData.Cluster, siteName)
			if err != nil {
				return fmt.Errorf("failed to get the cluster size. error: %s", err.Error())
			}
			hostLabels[ClusterSize] = strconv.FormatInt(clusterSize, 10)
			// set network mode
			clusterNetworkMode, err := t.netBox.GetClusterNetworkMode(ctx, t.deviceData.Cluster, siteName)
			if err != nil {
				return fmt.Errorf("failed to get the cluster network mode. error: %s", err.Error())
			}
			hostLabels[NetworkModeLabel] = clusterNetworkMode
		}
	}

	if err := t.assignComputeNodePoolLabel(ctx, hostLabels, bmh); err != nil {
		return fmt.Errorf("failed to assign ComputeNodePoolLabel to BareMetalHost. Error: %v", err)
	}

	bmh.SetLabels(hostLabels)

	hostAnnotations := bmh.GetAnnotations()
	if hostAnnotations == nil {
		hostAnnotations = make(map[string]string)
	}
	// add annotation if system is gaudi
	switch t.bmc.GetHwType() {
	case bmc.Gaudi2Smc, bmc.Gaudi2Wiwynn, bmc.Gaudi2Dell, bmc.Gaudi3Dell, bmc.Smc822GANGR3IN001:
		if err = t.GetExtraEthernetMacAnnotations(ctx, hostAnnotations); err != nil {
			return fmt.Errorf("unable to habana GPUs mac addresses: %v", err)
		}

		// GPU IP Assignment
		if hostLabels[NetworkModeLabel] == NetworkModeXBX {
			if err := t.GetGPUIPAddressesAnnotations(ctx, bmh, hostAnnotations); err != nil {
				return fmt.Errorf("unable to annotate GPU IP addresses %v", err)
			}
		}
	default:
		break
	}

	// add model to annotation instead because its value format is not suitable for label
	hostAnnotations[CPUModelLabel] = bmh.Status.HardwareDetails.CPU.Model
	bmh.SetAnnotations(hostAnnotations)

	labelStr, err := json.Marshal(bmh.GetLabels())
	if err != nil {
		return err
	}
	annotationsStr, err := json.Marshal(bmh.GetAnnotations())
	if err != nil {
		return err
	}

	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":%s, "annotations":%s}}`, labelStr, annotationsStr))

	_, err = t.dynamicClient.Resource(bmHostGVR).
		Namespace(bmh.GetNamespace()).
		Patch(ctx, bmh.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	if assignInstanceTypeLabelFailed {
		return fmt.Errorf("unable to assign InstanceType Label to BareMetalHost: %v", assignInstanceTypeErr)
	}

	return nil
}

func (t *EnrollmentTask) assignComputeNodePoolLabel(ctx context.Context, hostLabels map[string]string, host *baremetalv1alpha1.BareMetalHost) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.assignComputeNodePoolLabel")
	log.Info("Assigning ComputeNodePoolLabel to BareMetalHost")

	isClusterNode := (t.deviceData.Cluster != "None")

	// Add single nodes to 'General' pool
	if !isClusterNode || hostLabels[NetworkModeLabel] != NetworkModeXBX {
		hostLabels[fmt.Sprintf(ComputeNodePoolLabel, NodePoolGeneral)] = "true"
		return nil
	}

	// For XBX network mode:
	// Set a 'general' node pool label if no other members have a specific node pool.
	// Set a specific node pool label if other cluster members have a specific node pool.

	// Get cluster members
	siteName := os.Getenv(dcim.AvailabilityZoneEnvVar)
	if siteName == "" {
		return fmt.Errorf("failed to get the availability zone(site name) of the device")
	}

	selector := fmt.Sprintf("%s=%s", ClusterGroupID, t.deviceData.Cluster)
	clusterHostsItems, err := t.dynamicClient.Resource(bmHostGVR).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return fmt.Errorf("failed to get the list of hosts in cluster. error: %s", err.Error())
	}

	// Store frequency of pool assigned to cluster members
	nodePoolCount := make(map[string]int)
	for _, item := range clusterHostsItems.Items {
		clusterHost := &baremetalv1alpha1.BareMetalHost{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.UnstructuredContent(), clusterHost); err != nil {
			return fmt.Errorf("unable to decode BareMetalHost object")
		}
		if clusterHost.Name == host.Name {
			continue
		}

		computeNodePool, err := GetBareMetalHostComputeNodePool(clusterHost)
		if err != nil {
			return fmt.Errorf("failed to get ComputeNodePoolLabel on BareMetalHost %s. Error: %v", clusterHost.Name, err)
		}
		nodePoolCount[computeNodePool]++
	}

	// Set default pool label
	maxHostsPool := NodePoolGeneral
	mostAssignedPoolCount := 0

	for pool, count := range nodePoolCount {
		if count > mostAssignedPoolCount && pool != NodePoolGeneral {
			maxHostsPool = pool
			mostAssignedPoolCount = count
		}
	}

	hostLabels[fmt.Sprintf(ComputeNodePoolLabel, maxHostsPool)] = "true"
	return nil
}

func (t *EnrollmentTask) assignInstanceTypeLabel(ctx context.Context, hostLabels map[string]string, isStorageNicPresent bool) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.assignInstanceTypeLabel")
	log.Info("Assigning InstanceTypeLabel to host")

	instanceTypesSearchResponse, err := t.instanceTypeServiceClient.Search(ctx, &pb.InstanceTypeSearchRequest{})
	if err != nil {
		return fmt.Errorf("unable to get instanceTypeList information from InstanceTypeClient: %v", err)
	}

	// Map to store instanceTypes specifications
	instanceTypeSpecsLabels := make(map[string]string)

	// Map to store for host specifications
	hostSpecsLabels := make(map[string]string)

	hostSpecsLabels[CPUIDLabel] = hostLabels[CPUIDLabel]
	hostSpecsLabels[CPUCountLabel] = hostLabels[CPUCountLabel]
	hostSpecsLabels[GPUModelNameLabel] = hostLabels[GPUModelNameLabel]
	hostSpecsLabels[GPUCountLabel] = hostLabels[GPUCountLabel]
	hostSpecsLabels[HBMModeLabel] = hostLabels[HBMModeLabel]
	hostSpecsLabels[MemorySizeLabel] = hostLabels[MemorySizeLabel]
	hostSpecsLabels[CPUThreadsLabel] = hostLabels[CPUThreadsLabel]

	var instanceTypeLabel string
	instanceTypeLabelAssigned := false

	for _, instanceType := range instanceTypesSearchResponse.Items {
		if instanceType.Spec.InstanceCategory == pb.InstanceCategory_BareMetalHost {
			instanceTypeSpecsLabels[CPUIDLabel] = strings.ToLower(instanceType.Spec.Cpu.Id)
			instanceTypeSpecsLabels[CPUCountLabel] = strconv.Itoa(calculateTotalCpuCount(instanceType))
			instanceTypeSpecsLabels[GPUModelNameLabel] = strings.ToLower(instanceType.Spec.Gpu.ModelName)
			instanceTypeSpecsLabels[GPUCountLabel] = strconv.Itoa(int(instanceType.Spec.Gpu.Count))
			instanceTypeSpecsLabels[HBMModeLabel] = strings.ToLower(instanceType.Spec.HbmMode)
			instanceTypeSpecsLabels[MemorySizeLabel] = instanceType.Spec.Memory.Size
			instanceTypeSpecsLabels[CPUThreadsLabel] = strconv.Itoa(int(instanceType.Spec.Cpu.Threads))

			// match the specs
			specMatch := reflect.DeepEqual(instanceTypeSpecsLabels, hostSpecsLabels)
			if specMatch {
				instanceTypeName := instanceType.Spec.Name
				hostLabels[fmt.Sprintf(InstanceTypeLabel, instanceTypeName)] = "true"
				instanceTypeLabelAssigned = true

				if !strings.Contains(instanceTypeName, "sc") {
					instanceTypeLabel = fmt.Sprintf(InstanceTypeLabel, instanceTypeName)
				}
			}
		}
	}

	// Update the instance type label for SC support
	if isStorageNicPresent {
		delete(hostLabels, instanceTypeLabel)
	} else {
		delete(hostLabels, instanceTypeLabel+"-sc")
	}

	if !instanceTypeLabelAssigned {
		return errors.New("failed to assign InstanceType Label")
	}

	return nil
}

// Helper Methods

func calculateTotalCpuCount(instance *pb.InstanceType) int {
	cpu := instance.Spec.Cpu
	return int(cpu.Cores * cpu.Sockets * cpu.Threads)
}

// formatLabelValue returns a label value with invalid characters removed
func formatLabelValue(value string) string {
	r := strings.NewReplacer(
		"(R)", "",
		"(", "",
		")", "",
		"@", "",
	)
	return strings.Join(strings.Fields(r.Replace(value)), "_")
}

func convertSwitchIpToGpuIp(ip net.IP, subnet *net.IPNet) (net.IP, error) {
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address format")
	}

	for i := 3; i >= 0; i-- {
		if ip[i] > 0 {
			ip[i]--
			break
		}
		ip[i] = 255
	}

	// resulting IP should be within the same subnet
	if !subnet.Contains(ip) {
		return nil, fmt.Errorf("converted IP %s is out of subnet range", ip)
	}

	return ip, nil
}
