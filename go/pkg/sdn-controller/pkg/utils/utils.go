// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/utils/strings/slices"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"

	"k8s.io/client-go/rest"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idclog "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

const (
	LogFieldSwitchFQDN               = "switchFQDN"               // log field value example: internal-placeholder.com
	LogFieldSwitchPortName           = "switchPortName"           // log field value example: Ethernet27/1
	LogFieldSwitchPortCRName         = "switchPortCRName"         // log field value example: ethernet27-1.internal-placeholder.com
	LogFieldPortChannelCRName        = "portChannelCRName"        // log field value example: Po123.internal-placeholder.com
	LogFieldVlanID                   = "vlanID"                   // log field value example: 100
	LogFieldPortChannel              = "portChannel"              // log field value example: 12
	LogFieldPortChannelInterfaceName = "portChannelInterfaceName" // log field value example: Port-Channel12
	LogFieldResourceId               = "resourceId"               // log field value example: the resource id of a SwitchPortCR is ethernet27-1.internal-placeholder.com
	LogFieldController               = "controller"               // log field value example: SwitchPort
	LogFieldInterfaceName            = "interfaceName"            // eg. Ethernet1/1
	LogFieldIP                       = "ip"
	LogFieldBareMetalHost            = "bareMetalHost"
	LogFieldBareMetalHostState       = "bareMetalHostState"
	LogFieldDataCenter               = "dataCenter"
	LogFieldSwitchBackendMode        = "switchBackendMode"
	LogFieldNetworkNode              = "networkNode"
	LogFieldFabricType               = "fabricType"
	LogFieldNodeGroup                = "nodeGroup"
	LogFieldBGPCommunity             = "BGPCommunity" // log field value example: 100
	LogFieldTimeElapsed              = "timeElapsed"
	LogFieldPoolName                 = "pool"
	LogFieldNodeGroupName            = "nodeGroupName"
	LogFieldNodeGroupUpdateRequest   = "nodeGroupUpdateRequest"
	LogFieldSwitchIpToUse            = "switchIpTouse"
	LogFieldTrunkGroups              = "trunkGroups"
	LogFieldNativeVlan               = "nativeVlan"
	LogFieldMode                     = "mode"
	LogFieldNetboxClientVersion      = "netboxClientVersion"
	LogFieldDescription              = "description"
)

func NewK8SClient() client.Client {

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(idcv1alpha1.AddToScheme(scheme))
	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))

	kubeconfig := ctrl.GetConfigOrDie()
	// set the rate limiter
	// kubeconfig.QPS = 20   // the number of queries per second that are allowed to be sent to the API server
	// kubeconfig.Burst = 50 // the maximum number of queries that are allowed to be sent to the API server at once before the rate limiting mechanism kicks in
	kubeconfig.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter() // a fake rateLimiter that without rate limiting feature.

	kubeclient, err := client.New(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("create k8s client failed, %v", err)
		return nil
	}

	return kubeclient
}

func NewK8SClientWithScheme(scheme *runtime.Scheme) client.Client {
	kubeconfig := ctrl.GetConfigOrDie()

	// set the rate limiter
	kubeconfig.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter() // a fake rateLimiter that without rate limiting feature.

	kubeclient, err := client.New(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("create k8s client failed, %v", err)
		return nil
	}

	return kubeclient
}

func NewK8SWatchClientWithScheme(scheme *runtime.Scheme) client.WithWatch {
	kubeconfig := ctrl.GetConfigOrDie()

	kubeclient, err := client.NewWithWatch(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("create k8s client failed, %v", err)
		return nil
	}

	return kubeclient
}

func NewK8SClientFromConfAndScheme(ctx context.Context, kubeConfPath string, scheme *runtime.Scheme) client.Client {
	var kubeconfig *restclient.Config
	var err error
	if len(kubeConfPath) != 0 {
		kubeconfig, err = LoadKubeConfigFile(ctx, kubeConfPath)
	} else {
		kubeconfig = ctrl.GetConfigOrDie()
	}

	if kubeconfig == nil {
		fmt.Printf("create k8s client failed, %v", err)
		return nil
	}

	// disable the rate limiter.
	kubeconfig.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter()

	kubeclient, err := client.New(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("create k8s client failed, %v", err)
		return nil
	}

	return kubeclient
}

var shortEtRegex = regexp.MustCompile("^Et([0-9]+(/[0-9]+)*)$")
var shortVxRegex = regexp.MustCompile("^Vx([0-9]+)$")
var shortPoRegex = regexp.MustCompile("^Po([0-9]+)$")

// Et1 => Ethernet1
// Et2/1 => Ethernet2/1
// Vx1 => Vxlan1
func InterfaceShortToLongName(shortIntName string) (string, error) {
	if shortEtRegex.MatchString(shortIntName) {
		matches := shortEtRegex.FindStringSubmatch(shortIntName)
		return "Ethernet" + matches[1], nil
	} else if shortVxRegex.MatchString(shortIntName) {
		matches := shortVxRegex.FindStringSubmatch(shortIntName)
		return "Vxlan" + matches[1], nil
	} else if shortPoRegex.MatchString(shortIntName) {
		matches := shortPoRegex.FindStringSubmatch(shortIntName)
		return "Port-Channel" + matches[1], nil
	} else {
		return "", fmt.Errorf("shortIntName \"%s\" did not match knwon patterns EtX, VxX, PoX", shortIntName)
	}
}

const (
	DefaultAllowedVlanIdsStr       = "100-3999,4008"
	DefaultAllowedNativeVlanIdsStr = "1,55"
)

var intRangeRegex = regexp.MustCompile("^([0-9]+)-([0-9]+)$")

// 100-103 => []int{100,101,102,103}
// 100,110-112 => []int{100,110,111,112}
func ExpandVlanRanges(intSeries string) ([]int, error) {
	vlanStrings := strings.Split(intSeries, ",")
	var vlans []int
	for _, vlanString := range vlanStrings {
		if intRangeRegex.MatchString(vlanString) { // a range
			rangeParams := intRangeRegex.FindStringSubmatch(vlanString)
			rangeStart, err := strconv.Atoi(rangeParams[1])
			if err != nil {
				return nil, err
			}
			rangeEnd, err := strconv.Atoi(rangeParams[2])
			if err != nil {
				return nil, err
			}
			if rangeStart >= rangeEnd {
				return nil, fmt.Errorf("got a range with start >= end: %s", vlanString)
			}
			for i := rangeStart; i <= rangeEnd; i++ {
				vlans = append(vlans, i)
			}
		} else { // a single number
			vlanInt, err := strconv.Atoi(vlanString)
			if err != nil {
				return nil, err
			}
			vlans = append(vlans, vlanInt)
		}
	}
	return vlans, nil
}

func ValidateVlanValue(vlan int, allowedVlans []int) error {
	if len(allowedVlans) == 0 {
		return fmt.Errorf("no allowed VLANs provided")
	}

	for _, v := range allowedVlans {
		if v == vlan {
			return nil
		}
	}

	return fmt.Errorf("VLAN %d not in allowed VLANs given in config", vlan)
}

// Function to validate and sanitize the description
func ValidateAndSanitizeDescription(description string) (string, error) {
	// Define a regular expression for allowed characters (alphanumeric and spaces)
	re := regexp.MustCompile(`^[a-zA-Z0-9\s_\-/,]+$`)

	// Check if the description matches the allowed pattern
	if !re.MatchString(description) {
		return "", fmt.Errorf("description contains invalid characters")
	}

	// Trim leading and trailing spaces
	sanitizedDescription := strings.TrimSpace(description)

	// Check the length of the description (e.g., max 100 characters)
	if len(sanitizedDescription) > 100 {
		return "", fmt.Errorf("description is too long")
	}

	return sanitizedDescription, nil
}

// Extracts the numeric parts from an Ethernet interface name.
func ExtractNumbers(interfaceName string) (int, int, error) {
	parts := strings.Split(interfaceName, "/")
	firstPart, err := strconv.Atoi(strings.TrimPrefix(parts[0], "Ethernet"))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to extract first part from %s: %v", interfaceName, err)
	}
	secondPart := 0
	if len(parts) > 1 {
		secondPart, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("failed to extract second part from %s: %v", interfaceName, err)
		}
	}
	return firstPart, secondPart, nil
}

// Sorts the list of interfaces and returns the interface range
func GetInterfaceRange(ethernetInterfaces []string) (string, error) {
	// Extract numeric parts and handle errors before sorting
	type ethernetInterface struct {
		name       string
		firstPart  int
		secondPart int
	}

	var interfaces []ethernetInterface
	for _, iface := range ethernetInterfaces {
		firstPart, secondPart, err := ExtractNumbers(iface)
		if err != nil {
			return "", err
		}
		interfaces = append(interfaces, ethernetInterface{
			name:       iface,
			firstPart:  firstPart,
			secondPart: secondPart,
		})
	}

	// Sort the list of Ethernet interfaces based on the numeric parts
	sort.Slice(interfaces, func(i, j int) bool {
		if interfaces[i].firstPart == interfaces[j].firstPart {
			return interfaces[i].secondPart < interfaces[j].secondPart
		}
		return interfaces[i].firstPart < interfaces[j].firstPart
	})

	// Handle the case where there are no interfaces
	if len(interfaces) == 0 {
		return "", fmt.Errorf("no Ethernet interfaces found")
	}

	// Handle the case where there is only one interface
	if len(interfaces) == 1 {
		return interfaces[0].name, nil
	}

	// Find the first and last elements
	firstEthernetInterface := interfaces[0].name
	lastEthernetInterface := interfaces[len(interfaces)-1].name

	// Construct the interface range
	interfaceRange := fmt.Sprintf("%s-%s", firstEthernetInterface, lastEthernetInterface)

	return interfaceRange, nil
}

func ShouldUpdateSwitchPortVlan(switchPortCR idcv1alpha1.SwitchPort) bool {
	return switchPortCR.Spec.VlanId != idcv1alpha1.NOOPVlanID &&
		switchPortCR.Spec.VlanId != 0 &&
		(switchPortCR.Status.VlanId != 0 || (switchPortCR.Spec.Mode == "access" && (switchPortCR.Spec.PortChannel == 0 || switchPortCR.Status.PortChannel == 0))) &&
		switchPortCR.Spec.VlanId != switchPortCR.Status.VlanId
}

func ShouldUpdatePortChannelVlan(portChannelCR idcv1alpha1.PortChannel) bool {
	return portChannelCR.Spec.VlanId != idcv1alpha1.NOOPVlanID &&
		portChannelCR.Spec.VlanId != 0 &&
		(portChannelCR.Status.VlanId != 0 || portChannelCR.Spec.Mode == "access") &&
		portChannelCR.Spec.VlanId != portChannelCR.Status.VlanId
}

// Ethernet27/1 and fxhb3p3r-zal0112a.idcmgt.intel.com => ethernet27-1.fxhb3p3r-zal0112a.idcmgt.intel.com
// Ethernet5 and fxhb3p3r-zal0112a.idcmgt.intel.com => ethernet5.fxhb3p3r-zal0112a.idcmgt.intel.com
func GeneratePortFullName(switchFQDN, portName string) string {
	return strings.Replace(fmt.Sprintf("%s.%s", strings.ToLower(portName), strings.ToLower(switchFQDN)), "/", "-", -1)
}

// ethernet27-1.internal-placeholder.com => Ethernet27/1 and internal-placeholder.com
func PortFullNameToPortNameAndSwitchFQDN(switchPortCRName string) (string, string) {
	idx := strings.Index(switchPortCRName, ".")
	if idx < 0 {
		return "", ""
	}
	portShortName := strings.Title(strings.Replace(switchPortCRName[:idx], "-", "/", -1))
	switchFQDN := switchPortCRName[idx+1:]
	return portShortName, switchFQDN
}

func ValidatePortSpec(port idcv1alpha1.SwitchPortSpec, allowedVlanIds []int, allowedModes []string) error {
	return ValidatePort(int(port.VlanId), port.Name, port.Mode, allowedVlanIds, allowedModes)
}

func ValidatePort(vlan int, port string, mode string, allowedVlanIds []int, allowedModes []string) error {
	// Validation checks - to prevent putting tenant port on a provider vlan.
	err := ValidatePortValue(port)
	if err != nil {
		return err
	}

	err = ValidateVlanValue(vlan, allowedVlanIds)
	if err != nil {
		return err
	}

	err = ValidateModeValue(mode, allowedModes)
	if err != nil {
		return err
	}

	return nil
}

func ValidateModeValue(mode string, allowedModes []string) error {
	for _, allowedMode := range allowedModes {
		if mode == allowedMode {
			return nil // The mode is valid
		}
	}
	return fmt.Errorf("Mode %s is not allowed mode value", mode)
}

func NewSwitchPortTemplate(switchFQDN, portName string, vlanID int64, portChannel int64, labels map[string]string) *idcv1alpha1.SwitchPort {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["switch_fqdn"] = switchFQDN

	sp := &idcv1alpha1.SwitchPort{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Namespace: idcv1alpha1.SDNControllerNamespace,
			Name:      GeneratePortFullName(switchFQDN, portName),
		},
		Spec: idcv1alpha1.SwitchPortSpec{
			Name:        portName,
			Mode:        "",
			VlanId:      vlanID,
			PortChannel: portChannel,
		},
	}
	return sp
}

func NewSwitchTemplate(switchFQDN string, ip string) *idcv1alpha1.Switch {
	sp := &idcv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: idcv1alpha1.SDNControllerNamespace,
			Name:      switchFQDN,
		},
		Spec: idcv1alpha1.SwitchSpec{
			FQDN: switchFQDN,
			Ip:   ip,
			EAPIConf: &idcv1alpha1.EAPIConf{
				Port:      443,
				Transport: "https",
			},
			BGP: &idcv1alpha1.BGPConfig{},
		},
	}
	return sp
}

func GetIp(sw *idcv1alpha1.Switch, dc string) (string, error) {
	if sw == nil {
		return "", fmt.Errorf("sw is nil")
	}

	if len(sw.Spec.IpOverride) != 0 {
		err := ValidateIpOverride(sw.Spec.IpOverride, dc)
		if err == nil {
			return sw.Spec.IpOverride, nil
		} else {
			return "", fmt.Errorf("%s: is not a valid IP address or hostname, %s", sw.Spec.IpOverride, err.Error())
		}
	} else {
		err := ValidateSwitchFQDN(sw.Spec.FQDN, dc)
		if err != nil {
			return "", fmt.Errorf("%s: is not a valid switch FQDN, %s", sw.Spec.FQDN, err.Error())
		} else {
			return sw.Spec.FQDN, nil
		}
	}
}

func ValidateIpOverride(value string, dc string) error {
	ipErr := ValidateIP(value)
	if ipErr == nil {
		return nil
	}

	fqdnErr := ValidateSwitchFQDN(value, dc)
	if fqdnErr == nil {
		return nil
	}

	return fmt.Errorf("value is neither a valid IP address nor a valid hostname: %s", value)
}

func ValidateIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	if parsedIP.To4() == nil {
		return fmt.Errorf("invalid IPv4 address: %s", ip)
	}
	return nil
}

func ValidateSwitchFQDN(switchFQDN string, datacenter string) error {
	if switchFQDN == "" {
		return errors.New("switch FQDN was empty")
	}
	// if datacenter is provided, then we will check if the switchFQDN's datacenter prefix falls into the list.
	if datacenter != "" {
		res := false
		dcList := strings.Split(datacenter, ":")
		dashIndex := strings.IndexByte(switchFQDN, '-')
		if dashIndex == -1 {
			return errors.New("invalid FQDN format")
		}
		dcPrefix := switchFQDN[:dashIndex]
		for _, dc := range dcList {
			if dc == dcPrefix {
				res = true
			}
		}
		if !res {
			return fmt.Errorf("the switch %v doesn't belong to the data center/s [%v]", switchFQDN, datacenter)
		}
	}

	// eg. internal-placeholder.com
	pattern := `^[a-zA-Z0-9]{8,9}-(zal|zas)[a-zA-Z0-9]{5}\.(fake)?idcmgt\.intel\.com$`
	re := regexp.MustCompile(pattern)
	isMatch := re.MatchString(switchFQDN)
	if isMatch {
		return nil
	}

	// eg. pdx05-c01-acsw001.us-staging-3.cloud.intel.com
	pattern = `^[a-zA-Z]{3}\d{2}-[a-zA-Z]{1}\d{2}-[a-zA-Z0-9]{0,10}\.[a-zA-Z]{2}-(dev|staging|region)-\d{1,2}[a-zA-Z]?\.(fake)?cloud\.intel\.com$`
	re = regexp.MustCompile(pattern)
	isMatch = re.MatchString(switchFQDN)
	if isMatch {
		return nil
	}

	// eg. clab-allscfabrics-accply2-leaf2 (containerlab)
	pattern = `^clab-[A-Za-z0-9-]+$`
	re = regexp.MustCompile(pattern)
	isMatch = re.MatchString(switchFQDN)
	if isMatch {
		return nil
	}

	return errors.New("invalid FQDN format")
}

// ValidatePortValue validates things like Ethernet1/2 or Ethernet1.
func ValidatePortValue(port string) error {
	pattern := `^Ethernet[1-9][0-9]{0,2}(/[1-9][0-9]{0,2})*$`
	validInterfaceRegex := regexp.MustCompile(pattern)
	if validInterfaceRegex.MatchString(port) {
		return nil
	}

	if err := ValidatePortChannelName(port); err == nil {
		return nil
	}

	return fmt.Errorf("invalid interface name: %s", port)
}

// ValidatePortNumber validates things like 12/3 or 2
func ValidatePortNumber(port string) error {
	pattern := `^[1-9][0-9]{0,2}(/[1-9][0-9]{0,2})*?$`
	validInterfaceRegex := regexp.MustCompile(pattern)
	if !validInterfaceRegex.MatchString(port) {
		return fmt.Errorf("invalid port number: %s", port)
	}
	return nil
}

func ValidateBGPCommunityGroupName(bgpGroupName string) error {
	if len(bgpGroupName) == 0 || len(bgpGroupName) > 32 {
		return fmt.Errorf("BGPCommunityGroupName must be between 1 and 32 characters")
	}

	pattern := `^[a-zA-Z0-9_-]*$`
	re := regexp.MustCompile(pattern)
	isMatch := re.MatchString(bgpGroupName)
	if !isMatch {
		return fmt.Errorf("BGPCommunityGroupName must be alphanumeric with - and _")
	}

	return nil
}

func ValidateBGPCommunityValue(community int32) error {
	if community < 0 || community > 65535 {
		return fmt.Errorf("invalid BGP Community. Requested: %d", community)
	}
	return nil
}

func BGPCommunityStringToValue(communityString string) (int, error) {
	bgpCommunityRegex := regexp.MustCompile("^101:([0-9]{1,5})$")
	match := bgpCommunityRegex.FindStringSubmatch(communityString)

	if match == nil {
		return 0, fmt.Errorf("wrong format for community string \"%s\"", communityString)
	}

	// Cast to int
	communityValue, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}

	err = ValidateBGPCommunityValue(int32(communityValue))
	if err != nil {
		return 0, err
	}

	return communityValue, nil
}

func BGPCommunityValueToString(communityValue int) (string, error) {

	err := ValidateBGPCommunityValue(int32(communityValue))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("101:%d", communityValue), nil
}

func SplitAndValidateTrunkGroupsString(trunkGroupsStr string, allowedTrunkGroups []string) ([]string, error) {
	if strings.TrimSpace(trunkGroupsStr) == "" {
		return nil, fmt.Errorf("empty trunkGroup")
	}

	trunkGroups := strings.Split(strings.TrimSpace(trunkGroupsStr), ",")
	trunkGroupsTrimmed := make([]string, 0)
	for _, trunkGroup := range trunkGroups {
		trunkGroupsTrimmed = append(trunkGroupsTrimmed, strings.TrimSpace(trunkGroup))
	}
	sort.Strings(trunkGroupsTrimmed)

	return trunkGroupsTrimmed, ValidateTrunkGroups(trunkGroupsTrimmed, allowedTrunkGroups)
}

func ValidateTrunkGroups(trunkGroups []string, allowedTrunkGroups []string) error {

	trunkGroupRegex := regexp.MustCompile("^[A-Za-z0-9_-]+$")
	for _, trunkGroup := range trunkGroups {

		if len(trunkGroup) > 32 {
			return fmt.Errorf("trunkGroup name must be under 32 characters: %s", trunkGroup)
		}

		if !trunkGroupRegex.MatchString(trunkGroup) {
			return fmt.Errorf("invalid trunkGroup name: %s", trunkGroup)
		}

		// Validate against whitelist of trunkGroups given in config.
		if len(allowedTrunkGroups) > 0 {
			if !slices.Contains(allowedTrunkGroups, trunkGroup) {
				return fmt.Errorf("trunkGroup %s not in allowedTrunkGroups given in config %v", trunkGroup, allowedTrunkGroups)
			}
		}
	}

	return nil
}

func LoadKubeConfigFile(ctx context.Context, filePath string) (*restclient.Config, error) {
	logger := idclog.FromContext(ctx).WithName("sdn-controller.utils.LoadKubeConfigFile")
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("LoadKubeConfigFile: ReadFile: unable to read %s: %w", filePath, err)
	}

	config, err := clientcmd.Load(configBytes)
	if err != nil {
		return nil, fmt.Errorf("LoadKubeConfigFile: clientcmd.Load: unable to load %s: %w", filePath, err)
	}

	// Loop over each kubecontext in the kubeconfig file until we find one we can connect to.
	var i = 0
	for kubecontextname, _ := range config.Contexts {
		i++
		logger.Info("Trying to connect to kubecontext", "kubecontextname", kubecontextname, "contextNumber", i, "totalContexts", len(config.Contexts))
		clientConfig := clientcmd.NewNonInteractiveClientConfig(*config, kubecontextname, nil, nil)
		restConfig, err := clientConfig.ClientConfig()

		if err != nil {
			logger.Error(err, "Unable to create clientConfig for kubecontext. Will try next context.", "kubecontextname", kubecontextname)
			continue
		}

		_, err = client.New(restConfig, client.Options{})
		if err != nil {
			logger.Error(err, "Unable to create k8sClient from restConfig.  Will try next context.", "kubecontextname", kubecontextname)
			continue
		}

		logger.Info("Successfully loaded kubeconfig file.", "kubecontextname", kubecontextname)

		return restConfig, nil
	}

	return nil, fmt.Errorf("LoadKubeConfigFile: failed to connect to any of the kubecontexts within the kubeconfig file %s", filePath)
}

// "Ethernet27/1" -> "27/1"
func ConvertStandardPortNameToRavenFormat(portName string) string {
	return strings.Replace(portName, "Ethernet", "", -1)
}

// "27/1" -> "Ethernet27/1"
func ConvertRavenPortNameToStandardFormat(portName string) string {
	return fmt.Sprintf("Ethernet%s", portName)
}

func FormatErrorWithHttpBody(methodName string, httpResp *http.Response, err error) error {
	var responseString []byte
	var err2 error
	if httpResp != nil {
		responseString, err2 = io.ReadAll(httpResp.Body)
	}
	if err2 == nil && responseString != nil {
		return fmt.Errorf("error when calling `%s`: %v  Response body: %s", methodName, err, responseString)
	} else {
		return fmt.Errorf("error when calling `%s`: %v", methodName, err)
	}
}

// "ethernet27-1.internal-placeholder.com" -> "Ethernet27/1" and "internal-placeholder.com"
func ExtractSwitchAndPortNameFromSwitchPortCRName(spName string) (string, string) {
	idx := strings.Index(spName, ".")
	if idx < 0 {
		return "", ""
	}

	portName := strings.Title(strings.Replace(spName[:idx], "-", "/", -1))
	swName := spName[idx+1:]
	return portName, swName
}

// ValidatePortChannelName allows values like "Port-Channel67"
func ValidatePortChannelName(port string) error {
	pattern := `^Port-Channel\d+$`
	validInterfaceRegex := regexp.MustCompile(pattern)
	if validInterfaceRegex.MatchString(port) {
		return nil
	}

	return fmt.Errorf("invalid Port-Channel name: %s", port)
}

// ValidatePortChannelNumber allows values like 67 or 1234
func ValidatePortChannelNumber(portChannelNumber int) error {
	if portChannelNumber <= 0 || portChannelNumber > 999999 {
		return fmt.Errorf("invalid port channel number: %d", portChannelNumber)
	}

	return nil
}

// "Port-Channel24" -> "24"
func PortChannelInterfaceNameToNumber(portName string) (int, error) {
	if err := ValidatePortChannelName(portName); err != nil {
		return 0, err
	}
	pattern := `^Port-Channel(\d+)$`
	portChannelRegex := regexp.MustCompile(pattern)
	portChannelNumbers := portChannelRegex.FindStringSubmatch(portName)
	portChannelNumber, err := strconv.Atoi(portChannelNumbers[1])
	if err != nil {
		return 0, err
	}

	return portChannelNumber, nil
}

// "24" -> "Port-Channel24"
func PortChannelNumberToInterfaceName(portChannelNumber int) (string, error) {
	if err := ValidatePortChannelNumber(portChannelNumber); err != nil {
		return "", err
	}

	return fmt.Sprintf("Port-Channel%d", portChannelNumber), nil
}

func PortChannelNumberAndSwitchFQDNToCRName(portChannelNumber int, switchFQDN string) (string, error) {
	if err := ValidatePortChannelNumber(portChannelNumber); err != nil {
		return "", err
	}

	if err := ValidateSwitchFQDN(switchFQDN, ""); err != nil {
		return "", err
	}

	return fmt.Sprintf("po%d.%s", portChannelNumber, switchFQDN), nil // po is the prefix for PortChannel. Used because "Port-ChannelX.{switchfqdn}" is not a valid CR name.
}

func PortChannelInterfaceNameAndSwitchFQDNToCRName(portChannelInterfaceName string, switchFQDN string) (string, error) {
	pcNumber, err := PortChannelInterfaceNameToNumber(portChannelInterfaceName)
	if err != nil {
		return "", err
	}

	if err := ValidateSwitchFQDN(switchFQDN, ""); err != nil {
		return "", err
	}

	return PortChannelNumberAndSwitchFQDNToCRName(pcNumber, switchFQDN)
}

func GetPools(path string) (map[string]*idcv1alpha1.Pool, error) {
	byteValue, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pools idcv1alpha1.PoolList

	err = json.Unmarshal(byteValue, &pools)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*idcv1alpha1.Pool)
	for i := range pools.Items {
		res[pools.Items[i].Name] = pools.Items[i]
	}
	return res, nil
}

func GetNodeGroupToPoolMapping(path string) (*idcv1alpha1.NodeGroupToPoolMap, error) {
	byteValue, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var NodeGroupToPoolMapping *idcv1alpha1.NodeGroupToPoolMap

	err = json.Unmarshal(byteValue, &NodeGroupToPoolMapping)
	if err != nil {
		return nil, err
	}
	return NodeGroupToPoolMapping, nil
}

func ConvertMac(mac string) string {
	// Convert Mac to a valid expected format (xx:xx:xx:xx:xx:xx).  Return empty string if invalid
	validMacAddr := ""
	if len(mac) == 14 {
		// Convert xxxx.xxxx.xxxx to xx:xx:xx:xx:xx:xx
		re := regexp.MustCompile(`[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}`)
		if re.MatchString(mac) {
			grps := strings.Split(mac, ".") // split by period
			vals := []string{}
			for _, grp := range grps {
				for i := 0; i < len(grp); i += 2 {
					vals = append(vals, grp[i:i+2])
				}
			}
			validMacAddr = strings.Join(vals, ":")
		}
	}
	if len(mac) == 17 {
		// Check for correct mac format
		re := regexp.MustCompile(`([0-9a-fA-F]{2}\:){5}[0-9a-fA-F]{2}`)
		if re.MatchString(mac) {
			validMacAddr = mac
		}
	}
	return validMacAddr
}

func RegexMatch(pattern string, input string) bool {
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString(input) {
		return false
	}
	return true
}

// "192.168.1.100/24" -> "192.168.1.100"
func ConvertIPFormat(input string) string {
	parts := strings.Split(input, "/")
	return parts[0]
}

func ListFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func ExtractTransportProtocol(rawURL string) (string, string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	return parsedURL.Scheme, parsedURL.Host, nil
}

// RetryableFunc is the function we are going to execute
type RetryableFunc func() error

// ShouldRetryFunc is the condition function that determine if it should trigger a retry
type ShouldRetryFunc func(error) bool

// ActionFunc is the action should be taken when we need a retry
type ActionFunc func() error

func ExecuteWithRetry(retryableFunc RetryableFunc, shouldRetry ShouldRetryFunc, action ActionFunc, maxRetries int) error {
	logger := log.FromContext(context.Background()).WithName("ExecuteWithRetry")
	var err error
	var i int
	// if we set maxRetries to 3, then this will run 4 rounds, as the first round is not considered as retry.
	for i = 0; i <= maxRetries; i++ {
		err = retryableFunc()
		if err != nil {
			logger.Info(fmt.Sprintf("retryableFunc failed on retry number %d, %v", i, err))
			if shouldRetry(err) {
				actionErr := action()
				if actionErr != nil {
					// if it returns error after running the action function, we will also make it trigger a retry.
					logger.Info(fmt.Sprintf("action function failed on retry number %d, %v\n", i, actionErr))
				}
				time.Sleep(3 * time.Second)
				continue
			}
			return err
		}
		break
	}
	if err != nil {
		return fmt.Errorf("%v, retries: %v/%v", err, i-1, maxRetries)
	}
	return nil
}

func UpdateSwitchConfRequired(switchPort *idcv1alpha1.SwitchPort) bool {
	return switchPort.Spec.VlanId != idcv1alpha1.NOOPVlanID &&
		switchPort.Spec.VlanId != 0 &&
		switchPort.Spec.VlanId != switchPort.Status.VlanId &&
		switchPort.Status.Mode != "routed" // No need to update if mode is "routed" because the vlan is not used in this case
}

// GetServersFromKubeconfig extracts all server URLs from a kubeconfig file
func GetServersFromKubeconfig(kubeconfigPath string) ([]string, error) {
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
	}

	serverURLs := make([]string, 0)
	for _, cluster := range config.Clusters {
		serverURLs = append(serverURLs, cluster.Server)
	}
	return serverURLs, nil
}

func GenerateServersKey(kubeconfigPath string) string {
	servers, err := GetServersFromKubeconfig(kubeconfigPath)
	if err != nil {
		return ""
	}
	sort.Strings(servers)
	return strings.Join(servers, ";")
}

func ContainsFinalizer(finalizers []string, s string) bool {
	for _, item := range finalizers {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveFinalizer(finalizers []string, s string) []string {
	var result []string
	for _, item := range finalizers {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// NewClientWithoutRateLimiter creates a k8s client with a disbaled rate limiter.
// Note: The reason we disabled the rate limiter is that in the previous benchmark test, it blocked the k8s api call when we scaled up the number of CRs.
// future improvement: try to evaluation and estimate the QPS and provide a more specific rate limiter like the below example:
// config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(50, 100)
func NewClientWithoutRateLimiter(config *rest.Config, options client.Options) (c client.Client, err error) {
	config.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter()
	return client.New(config, options)
}
