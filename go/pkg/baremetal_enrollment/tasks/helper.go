// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tasks

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ddi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type Runnable interface {
	Run(ctx context.Context) error
}

func getDeviceData(ctx context.Context) (device *DeviceData, err error) {
	log := log.FromContext(ctx).WithName("getDeviceData")
	log.Info("Gathering device information")

	device = &DeviceData{}

	device.Name = os.Getenv(dcim.DeviceNameEnvVar)
	if device.Name == "" {
		return nil, fmt.Errorf("failed to get the NetBox device Name")
	}

	device.ID, err = strconv.ParseInt(os.Getenv(dcim.DeviceIdEnvVar), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to get the netbox device ID: %v", err)
	}

	device.Rack = os.Getenv(dcim.RackNameEnvVar)
	if device.Rack == "" {
		return nil, fmt.Errorf("failed to get the NetBox rack Name")
	}

	device.Region = os.Getenv(dcim.RegionEnvVar)
	if device.Region == "" {
		return nil, fmt.Errorf("failed to get the enrollment region Name")
	}

	device.Cluster = os.Getenv(dcim.ClusterNameEnvVar)
	if device.Cluster == "" {
		return nil, fmt.Errorf("failed to get the enrollment cluster Name")
	}
	log.Info("Cluster name", "cluster", device.Cluster)

	log.Info("Successfully gathered Vault client")
	return device, nil
}

func getK8sClients(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
	log := log.FromContext(ctx).WithName("getK8sClients")
	log.Info("Initializing K8s Clients")

	config, err := util.GetRESTConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get K8s REST config: %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get K8s ClientSet: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get K8s Dynamic Client: %v", err)
	}

	log.Info("Successfully initialized K8s client")
	return clientSet, dynamicClient, nil
}

func GetVaultClient(ctx context.Context) (secrets.SecretManager, error) {
	log := log.FromContext(ctx).WithName("getVaultClient")
	log.Info("Initializing Vault client")

	vault, err := secrets.NewVaultClient(ctx,
		secrets.VaultOptionRenewToken(true),
		secrets.VaultOptionValidateClient(true),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	log.Info("Successfully initialized Vault client")
	return vault, nil
}

func GetNetBoxClient(ctx context.Context, accessor secrets.NetBoxSecretAccessor, region string) (dcim.DCIM, error) {
	log := log.FromContext(ctx).WithName("getNetBoxClient")
	log.Info("Initializing NetBox client")

	if accessor == nil {
		return nil, fmt.Errorf("NetBox secret accessor has been initialized")
	}

	secretPath := fmt.Sprintf("%s/baremetal/enrollment/netbox", region)
	token, err := accessor.GetNetBoxAPIToken(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get NetBox API token: %v", err)
	}

	netBox, err := dcim.NewNetBoxClient(ctx, token, false)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client: %v", err)
	}

	log.Info("Successfully initialized NetBox client")
	return netBox, nil
}

func getBMCURL(ctx context.Context, deviceData *DeviceData, netBox dcim.DCIM, menAndMice ddi.DDI) (bmcURL string, err error) {
	log := log.FromContext(ctx).WithName("getBMCURL")

	if menAndMice == nil {
		log.Info("Getting BMC URL from NetBox")
		bmcURL, err = netBox.GetBMCURL(ctx, deviceData.Name)
		if err != nil {
			return "", fmt.Errorf("unable to get BMC URL with error:  %v", err)
		}
	} else {
		log.Info("Getting BMC URL from DHCP Reservation")
		bmcMACAddress, err := netBox.GetBMCMACAddress(ctx, deviceData.Name, BMCInterfaceName)
		if err != nil {
			return "", fmt.Errorf("unable to get BMC MACAddress with error:  %v", err)
		}

		// find Range for BMC network based on Type = BMC and RackNAme
		ipRange, err := menAndMice.GetRangeByName(ctx, deviceData.Rack, menAndMiceBMCType)
		if err != nil {
			return "", fmt.Errorf("failed to read Range: %v of Rack %s", err, deviceData.Rack)
		}
		if len(ipRange.DhcpScopes) < 1 {
			return "", fmt.Errorf("failed to find dhcp Scopes in Range: %s", ipRange.Ref)
		}

		dhcpReservation, err := menAndMice.GetDhcpReservationsByMacAddress(ctx, ipRange.DhcpScopes[0].Ref, bmcMACAddress)
		if err != nil {
			// TODO check Leases and convert to reservation
			// Currently delete Lease cause a failure in dhcp server
			return "", fmt.Errorf("failed to get dhcp reservation: %v", err)
		}
		if len(dhcpReservation.Addresses) < 1 {
			return "", fmt.Errorf("no addresses found for dhcpreservation: %s", dhcpReservation.Ref)
		}
		bmcURL = fmt.Sprintf("https://%s", dhcpReservation.Addresses[0])
	}

	log.Info("Found BMC URL", "bmcURL", bmcURL)
	return bmcURL, nil
}

func getBMCData(ctx context.Context, deviceData *DeviceData, netBox dcim.DCIM, vault secrets.SecretManager, menAndMice ddi.DDI) (*BMCData, error) {
	log := log.FromContext(ctx).WithName("getBMCData")
	log.Info("Gathering BMC data")

	bmcURL, err := getBMCURL(ctx, deviceData, netBox, menAndMice)
	if err != nil {
		return nil, fmt.Errorf("unable to get BMC URL:  %v", err)
	}

	var bmcMACAddress string
	if menAndMice != nil {
		bmcMACAddress, err = netBox.GetBMCMACAddress(ctx, deviceData.Name, BMCInterfaceName)
		if err != nil {
			return nil, fmt.Errorf("unable to get BMC MACAddress:  %v", err)
		}
	}

	// get BMC credentials from Vault
	var secretPath string
	if bmcMACAddress == "" {
		secretPath = fmt.Sprintf("%s/deployed/virtual/default", deviceData.Region)
	} else {
		secretPath = fmt.Sprintf("%s/deployed/%s/default", deviceData.Region, bmcMACAddress)
	}
	bmcUsername, bmcPassword, err := vault.GetBMCCredentials(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get BMC Credentials:  %v", err)
	}

	bmcData := &BMCData{
		MACAddress: bmcMACAddress,
		URL:        bmcURL,
		Username:   bmcUsername,
		Password:   bmcPassword,
	}

	log.Info("Successfully gathered BMC data", "bmcURL", bmcURL, "bmcMACAddress", bmcMACAddress)
	return bmcData, nil
}

func GetMenAndMiceClient(ctx context.Context, region string, vault secrets.DDISecretAccessor) (menAndMice ddi.DDI, err error) {
	log := log.FromContext(ctx).WithName("getMenAndMiceClient")
	log.Info("Initializing MenAndMice client")

	menAndMiceUrl, exists := os.LookupEnv(MenAndMiceUrlEnvVar)
	if !exists {
		log.Info("MenAndMice URL is not found in environment")
		return nil, nil
	}

	menAndMiceServerAddress, exists := os.LookupEnv(MenAndMiceServerAddressEnvVar)
	if !exists {
		log.Info("MenAndMice server address is not found in environment")
		return nil, nil
	}

	secretPath := fmt.Sprintf("%s/baremetal/enrollment/menandmice", region)
	ddiUsername, ddiPassword, err := vault.GetDDICredentials(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get Men and Mice Credentails:  %v", err)
	}

	menAndMice, err = ddi.NewMenAndMice(ctx, ddiUsername, ddiPassword, menAndMiceUrl, menAndMiceServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create men and mice object:  %v", err)
	}

	log.Info("Successfully initialized MenAndMice client")
	return menAndMice, nil
}

func getBMCInterface(ctx context.Context, deviceData *DeviceData, bmcData *BMCData) (bmc.Interface, error) {
	log := log.FromContext(ctx).WithName("getBMCInterface")
	log.Info("Initializing BMC interface")

	bmcInterface, err := bmc.New(&mygofish.MyGoFishManager{}, &bmc.Config{
		URL:      bmcData.URL,
		Username: bmcData.Username,
		Password: bmcData.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create BMC interface: %v", err)
	}

	log.Info("Successfully initialized BMC interface")
	return bmcInterface, nil
}

func getMetal3Namespaces(ctx context.Context, clientSet kubernetes.Interface) ([]corev1.Namespace, error) {
	log := log.FromContext(ctx).WithName("getMetal3Namespaces")
	log.Info("Getting Metal3 namespaces")

	selector := fmt.Sprintf("%s=true", Metal3NamespaceSelectorKey)
	namespaces, err := clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("unable to list metal3 namespaces: %v", err)
	}
	if len(namespaces.Items) == 0 {
		return nil, fmt.Errorf("no metal3 namespace found")
	}

	return namespaces.Items, nil
}

func GetInstanceTypeServiceClient(ctx context.Context) (pb.InstanceTypeServiceClient, error) {
	log := log.FromContext(ctx).WithName("getInstanceTypeServiceClient")
	log.Info("Initializing InstanceTypeService Client")

	computeApiServerAddr := os.Getenv(ComputeApiServerAddrEnvVar)
	if computeApiServerAddr == "" {
		return nil, fmt.Errorf("failed to get the compute api server Address")
	}

	computeApiServerClientConn, err := grpcutil.NewClient(ctx, computeApiServerAddr)
	if err != nil {
		return nil, fmt.Errorf("computeApiServerClientConn is not getting init %v", err)
	}

	return pb.NewInstanceTypeServiceClient(computeApiServerClientConn), nil
}

func updateDeviceStatus(ctx context.Context, netbox dcim.DCIM, deviceData *DeviceData, status dcim.BMEnrollmentStatus, comment string) error {
	if err := netbox.UpdateDeviceCustomFields(ctx, deviceData.Name, deviceData.ID, &dcim.DeviceCustomFields{
		BMEnrollmentStatus:  status,
		BMEnrollmentComment: comment,
	}); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}
	return nil
}

func GetBareMetalHostComputeNodePool(bmh *baremetalv1alpha1.BareMetalHost) (string, error) {
	for k := range bmh.Labels {
		if strings.HasPrefix(k, "pool.cloud.intel.com/") {
			substrings := strings.Split(k, "pool.cloud.intel.com/")
			if substrings[1] != "" {
				return substrings[1], nil
			} else {
				return "", fmt.Errorf("invalid compute node pool label ")
			}
		}
	}
	return "", nil
}
