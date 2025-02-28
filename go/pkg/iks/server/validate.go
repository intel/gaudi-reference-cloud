// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package server

import (
	"errors"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var AlphanumericHyphenUnderscoreRegex = regexp.MustCompile("^[a-zA-Z0-9-_]+$")
var AlphanumericHyphenUnderscoreSpacesRegex = regexp.MustCompile("^[a-zA-Z0-9-_\\s]+$")
var ValidIPCIDR = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
var SemanticVersionRegex = regexp.MustCompile(`^v?(\d+\.)?(\d+\.)?(\*|\d+)?$`)

func validateSourceIps(sourceips []string) error {
	result := true
	l := len(sourceips)
	uniqueValues := make(map[string]bool)
	for _, i := range sourceips {
		ipv4Regex := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\/(3[0-2]|[1-2]?[0-9])$|^any$`
		if i == "any" && l > 1 {
			result = false
		}
		if regexp.MustCompile(ipv4Regex).MatchString(i) != true {
			result = false
		}
		if _, ok := uniqueValues[i]; ok {
			result = false
		} else {
			uniqueValues[i] = true
		}
		if strings.Contains(i, "/") {
			_, ipv4Net, err := net.ParseCIDR(i)
			if err != nil {
				result = false
			}
			if ipv4Net.String() != i {
				result = false
			}
		} else {
			if i != "any" && net.ParseIP(i) == nil {
				result = false
			}
		}
	}
	if result == false {
		return fmt.Errorf("security rule only accepts unique, valid source ips and subnets or any")
	}
	return nil

}

// validatePort validates port
func validatePort(port int32) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port should be between 1 and 65535")
	}
	return nil
}

// validateClusterRequest validates cluster request
func validateClusterRequest(req *pb.ClusterRequest) error {
	var resultErr error

	if err := validateString(req.Name, "cluster name", stringMaxLength(63), stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Description != nil && *req.Description != "" {
		if err := validateString(*req.Description, "description", stringMaxLength(253), stringAlphaNumericWithSpaces()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if err := validateString(req.K8Sversionname, "k8s version name", stringSemanticVersion()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Runtimename, "runtime name", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Network != nil {
		if err := validateNetwork(req.Network); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if err := validateAnnotations(req.Annotations); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateKeyValuePair(req.Tags); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Clustertype != nil && *req.Clustertype != "" {
		if err := validateString(*req.Clustertype, "cluster type", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if req.InstanceType != "" {
		if err := validateString(req.InstanceType, "instance type", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	return resultErr
}

// validateClusterID validates cluster id
func validateClusterID(req *pb.ClusterID) error {
	var resultErr error

	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if req.Clustertype != nil && *req.Clustertype != "" {
		if err := validateString(*req.Clustertype, "cluster type", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	return resultErr
}

// validateIksCloudAccountId validates iks cloud account id
func validateIksCloudAccountId(req *pb.IksCloudAccountId) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

// validateGetClusterRequest validates get cluster request
func validateGetNodeGroupsRequest(req *pb.GetNodeGroupsRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

// validateGetClusterRequest validates get cluster request
func validateGetNodeGroupRequest(req *pb.GetNodeGroupRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Nodegroupuuid, "nodegroup uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

// validateNodeGroupidRequest validates node group id request
func validateNodeGroupidRequest(req *pb.NodeGroupid) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Nodegroupuuid, "nodegroup uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Nodegrouptype != nil && *req.Nodegrouptype != "" {
		if err := validateString(*req.Nodegrouptype, "nodegroup type", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	return resultErr
}

// validateDeleteNodeGroupRequest validates delete nodegroup request
func validateDeleteNodeGroupInstanceRequest(req *pb.DeleteNodeGroupInstanceRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Nodegroupuuid, "nodegroup uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.InstanceName, "instance name", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

// validateUpdateClusterRequest validates update cluster request
func validateUpdateClusterRequest(req *pb.UpdateClusterRequest) error {
	var resultErr error

	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Name != nil && *req.Name != "" {
		if err := validateString(*req.Name, "cluster name", stringAlphaNumeric(), stringMaxLength(63)); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if req.Description != nil {
		if err := validateString(*req.Description, "description", stringAlphaNumericWithSpaces(), stringMaxLength(253)); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if err := validateAnnotations(req.Annotations); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateKeyValuePair(req.Tags); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	return resultErr
}

// validateCreateNodeGroupRequest validates create nodegroup request
func validateCreateNodeGroupRequest(req *pb.CreateNodeGroupRequest) error {
	var resultErr error

	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Name, "nodegroup name", stringMaxLength(63), stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Description != nil && *req.Description != "" {
		if err := validateString(*req.Description, "description", stringMaxLength(253), stringAlphaNumericWithSpaces()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if err := validateString(req.Instancetypeid, "instance type id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateVnets(req.Vnets); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateNodeCount(req.Count, 1, 10); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateSshKey(req.Sshkeyname); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if err := validateKeyValuePair(req.Tags); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateUpgradeStrategy(req.Upgradestrategy); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateAnnotations(req.Annotations); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateUserdataUrl(req.Userdataurl); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Nodegrouptype != nil {
		if err := validateString(*req.Nodegrouptype, "nodegroup type", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	return resultErr
}

func validateUpgradeClusterRequest(req *pb.UpgradeClusterRequest) error {
	var resultErr error

	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.K8Sversionname != nil && *req.K8Sversionname != "" {
		if err := validateString(*req.K8Sversionname, "k8s version name", stringSemanticVersion()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	return resultErr
}

func validateClusterStorageRequest(req *pb.ClusterStorageRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Storagesize, "storage size ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	return resultErr
}

func validateClusterStorageUpdateRequest(req *pb.ClusterStorageUpdateRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Storagesize, "storage size ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	return resultErr
}

func validateNodeGroupid(req *pb.NodeGroupid) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Nodegroupuuid, "nodegroup uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Nodegrouptype != nil && *req.Nodegrouptype != "" {
		if err := validateString(*req.Nodegrouptype, "nodegroup type ", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	return resultErr
}

func validateVipCreateRequest(req *pb.VipCreateRequest) error {
	var resultErr error
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Name, "vip name", stringAlphaNumeric(), stringMaxLength(63)); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if req.Description != "" {
		if err := validateString(req.Description, "vip description", stringAlphaNumericWithSpaces(), stringMaxLength(253)); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if err := validatePort(req.Port); err != nil {
		resultErr = errors.Join(resultErr, fmt.Errorf("port should be between 1 and 65535"))
	}
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	return resultErr
}

func validateUpdateFirewallRuleRequest(req *pb.UpdateFirewallRuleRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateSourceIps(req.Sourceip); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Internalip, "internal ip", stringValidIp()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validatePort(req.Port); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateProtocols(req.Protocol); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr

}

func validateGetKubeconfigRequest(req *pb.GetKubeconfigRequest) error {
	var resultErr error
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

func validateProtocols(req []string) error {
	var resultErr error
	for _, protocol := range req {
		if protocol != "TCP" && protocol != "UDP" {
			resultErr = errors.Join(resultErr, fmt.Errorf("protocol should be either TCP or UDP"))
		}
	}
	return resultErr
}

func validateVipId(req *pb.VipId) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

func validateDeleteFirewallRuleRequest(req *pb.DeleteFirewallRuleRequest) error {
	var resultErr error
	if err := validateString(req.CloudAccountId, "cloud account id", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	return resultErr
}

func validateNetwork(network *v1.Network) error {
	var resultErr error

	if err := validateString(network.Region, "network region", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}
	if network.Servicecidr != nil {
		if err := validateString(*network.Servicecidr, "network service cidr", stringValidIpCidr(), stringMaxLength(253)); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if network.Clustercidr != nil {
		if err := validateString(*network.Clustercidr, "network cluster cidr", stringValidIpCidr(), stringMaxLength(253)); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	if network.Clusterdns != nil {
		if err := validateString(*network.Clusterdns, "network cluster dns", stringAlphaNumericOrIP()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	return resultErr
}

// validateUpdateNodeGroupRequest validates update nodegroup request
func validateUpdateNodeGroupRequest(req *pb.UpdateNodeGroupRequest) error {
	var resultErr error

	if err := validateString(req.Clusteruuid, "cluster uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if req.Name != nil && *req.Name != "" {
		if err := validateString(*req.Name, "nodegroup name ", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	if err := validateString(req.Nodegroupuuid, "nodegroup uuid ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if req.Description != nil {
		if err := validateString(*req.Description, "description ", stringAlphaNumericWithSpaces(), stringMaxLength(253)); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	if req.Count != nil {
		if err := validateNodeCount(*req.Count, 0, 10); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	if err := validateKeyValuePair(req.Tags); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if req.Upgradestrategy != nil {
		if err := validateUpgradeStrategy(req.Upgradestrategy); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	if err := validateAnnotations(req.Annotations); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if err := validateString(req.CloudAccountId, "cloud account id ", stringAlphaNumeric()); err != nil {
		resultErr = errors.Join(resultErr, err)
	}

	if req.Nodegrouptype != nil && *req.Nodegrouptype != "" {
		if err := validateString(*req.Nodegrouptype, "nodegroup type ", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	return resultErr
}

type StringValidator func(string) error

func stringSemanticVersion() StringValidator {
	return func(input string) error {
		if !SemanticVersionRegex.MatchString(input) {
			return fmt.Errorf("should be valid semantic version")
		}
		return nil
	}
}

func stringAlphaNumeric() StringValidator {
	return func(input string) error {
		if !AlphanumericHyphenUnderscoreRegex.MatchString(input) {
			return fmt.Errorf("should be alphanumeric and can contain only - and _")
		}
		return nil
	}
}

func stringAlphaNumericWithSpaces() StringValidator {
	return func(input string) error {
		if !AlphanumericHyphenUnderscoreSpacesRegex.MatchString(input) {
			return fmt.Errorf("should be alphanumeric and can contain only - and _ and spaces")
		}
		return nil
	}
}

func stringValidIp() StringValidator {
	return func(input string) error {
		if net.ParseIP(input) == nil {
			return fmt.Errorf("should be valid ip")
		}
		return nil
	}
}

func stringValidIpCidr() StringValidator {
	return func(input string) error {
		if !ValidIPCIDR.MatchString(input) {
			return fmt.Errorf("should be valid ip cidr")
		}
		return nil
	}
}

func stringAlphaNumericOrIP() StringValidator {
	return func(input string) error {
		if !AlphanumericHyphenUnderscoreRegex.MatchString(input) && net.ParseIP(input) == nil {
			return fmt.Errorf("should be alphanumeric or valid ip")
		}
		return nil
	}
}

func stringMaxLength(maxLength int) StringValidator {
	return func(input string) error {
		if len(input) > maxLength {
			return fmt.Errorf("cannot exceed %d bytes", maxLength)
		}
		return nil
	}
}

func stringMinLength(minLength int) StringValidator {
	return func(input string) error {
		if len(input) < minLength {
			return fmt.Errorf("should have atleast %d characters", minLength)
		}
		return nil
	}
}

func validateString(input string, paramName string, validators ...StringValidator) error {
	var resultErr error
	for _, validator := range validators {
		if err := validator(input); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("%s: %v", paramName, err))
		}
	}
	return resultErr
}

// validateUserdataUrl validates userdata URL
func validateUserdataUrl(userdataUrl *string) error {
	var resultErr error
	if userdataUrl != nil && *userdataUrl != "" {
		_, err := url.ParseRequestURI(*userdataUrl)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("invalid userdata URL"))
		}
	}
	return resultErr
}

// validateUpgradeStrategy validates upgrade strategy
func validateUpgradeStrategy(upgradestrategy *v1.UpgradeStrategy) error {
	var resultErr error
	if upgradestrategy != nil {
		if upgradestrategy.Maxunavailablepercentage < 0 || upgradestrategy.Maxunavailablepercentage > 100 {
			resultErr = errors.Join(resultErr, fmt.Errorf("maxunavailablepercentage unavailable should be between 0 and 100"))
		}
	}
	return resultErr
}

func validateNodeCount(count, min, max int32) error {
	if count < 0 || count > 10 {
		return fmt.Errorf("nodegroup count should be between %d and %d", min, max)
	}
	return nil
}

// validateKeyValuePair validates key value pair
func validateKeyValuePair(keyValuePairs []*v1.KeyValuePair) error {
	var resultErr error
	for _, keyValuePair := range keyValuePairs {
		if err := validateString(keyValuePair.Key, "key", stringMaxLength(63), stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}

		if err := validateString(keyValuePair.Value, "value", stringMaxLength(253), stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	return resultErr
}

func validateAnnotations(annotations []*v1.Annotations) error {
	var resultErr error
	if annotations == nil || len(annotations) == 0 {
		return nil
	}

	for _, a := range annotations {
		if err := validateString(a.Key, "annotation key", stringMaxLength(63), stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
		if err := validateString(a.Value, "annotation value", stringMaxLength(253), stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}

	return resultErr
}

// validateSshKey validates ssh key
func validateSshKey(sshKey []*v1.SshKey) error {
	var resultErr error

	if sshKey == nil || len(sshKey) == 0 {
		return nil
	}
	for _, key := range sshKey {
		if key.Sshkey == "" || !AlphanumericHyphenUnderscoreRegex.MatchString(key.Sshkey) || len(key.Sshkey) > 63 {
			resultErr = errors.Join(resultErr, fmt.Errorf("ssh key should be alphanumeric and can contain only - and _, cannot exceed 63 bytes"))
		}
	}

	return resultErr
}

// validateVnets validates vnets
func validateVnets(vnets []*pb.Vnet) error {
	var resultErr error
	if len(vnets) == 0 {
		return fmt.Errorf("vnets cannot be empty")
	}
	for _, vnet := range vnets {
		if err := validateString(vnet.Networkinterfacevnetname, "vnet name", stringMaxLength(63), stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
		if err := validateString(vnet.Availabilityzonename, "availability zone name", stringAlphaNumeric()); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}
	return resultErr
}
