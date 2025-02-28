// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"net"
	"regexp"
	"strings"
)

const (
	minNetworkPrefixLen  = 16
	maxNetworkPrefixLen  = 28
	IPv4LinkLocalAddress = "169.254.0.0/16"
	subnetNameMaxLength  = 63
)

var _, IPv4LinkLocal, _ = net.ParseCIDR(IPv4LinkLocalAddress)

var subnetNamePattern = regexp.MustCompile("^[a-zA-Z0-9_\\-.]+$")

// The CIDR field is mandatory for VPCs and subnets.
// It specifies an IPv4 network in the format: "x.x.x.x/mask".
// Restrictions:
// - Must be a valid IPv4 address.
// - Mask length must be between 16 and 28 (inclusive).
// - Subnets must not overlap with the local-link range (169.254.0.0/16)
// See: https://www.rfc-editor.org/rfc/rfc4632
// See: https://www.rfc-editor.org/rfc/rfc5735
func ValidateCIDR(cidrBlock string) error {
	_, ipNet, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return fmt.Errorf("Invalid CIDR: %v", err)
	}

	// Validate IPv4.
	if ipNet.IP.To4() == nil {
		return fmt.Errorf("Invalid CIDR: should be in IPv4 format")
	}

	// Validate that the mask is in canonical form (contiguous 1s followed by 0s) and that its length is valid
	numOfOnes, numOfBits := ipNet.Mask.Size()
	if (numOfOnes == 0 && numOfBits == 0) || numOfOnes < minNetworkPrefixLen || numOfOnes > maxNetworkPrefixLen {
		return fmt.Errorf("Invalid CIDR: netmask must be between %d and %d, got: %v", minNetworkPrefixLen, maxNetworkPrefixLen, cidrBlock)
	}

	// Is it link-local?
	if isOverlap, err := isCIDROverlap(ipNet, IPv4LinkLocal); err != nil {
		return fmt.Errorf("Invalid CIDR: Error checking CIDR overlap: %v", err)
	} else if isOverlap {
		return fmt.Errorf("Invalid CIDR: overlap with local link address: %v", cidrBlock)
	}

	return nil
}

// Subnet name is not mandatory.
// It cant be just spaces.
// It can't contain leading or trailing spaces.
func ValidateSubnetName(name string) error {
	// no name are allowed
	nameLen := len(name)
	if nameLen == 0 {
		return nil
	}

	if nameLen > subnetNameMaxLength {
		return fmt.Errorf("Invalid name: must be less than %d characters", subnetNameMaxLength)
	}

	if name != strings.TrimSpace(name) {
		return fmt.Errorf("Invalid name: ensure it doesn't have leading or trailing spaces.")
	}

	if matched := subnetNamePattern.MatchString(name); !matched {
		return fmt.Errorf("Invalid name: contains invalid characters")
	}

	return nil
}

func isCIDROverlap(ipNet1, ipNet2 *net.IPNet) (bool, error) {

	net1FirstIP, err := ipToInt(ipNet1.IP)
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, ipNet1.IP)
	}

	net1LastIP, err := ipToInt(lastIP(ipNet1))
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, lastIP(ipNet1))
	}

	net2FirstIP, err := ipToInt(ipNet2.IP)
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, ipNet2.IP)
	}

	net2LastIP, err := ipToInt(lastIP(ipNet2))
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, lastIP(ipNet2))
	}

	return net1LastIP >= net2FirstIP && net1FirstIP <= net2LastIP, nil
}

// Returns true if the childCIDR is within the parentCIDR. false otherwise.
func IsCIDRWithinCIDR(parentCIDR, childCIDR string) (bool, error) {
	_, parentIPNet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return false, fmt.Errorf("failed to parse parentCIDR: %v", err)
	}

	_, childIPNet, err := net.ParseCIDR(childCIDR)
	if err != nil {
		return false, fmt.Errorf("failed to parse childCIDR: %v", err)
	}

	parentFirstIP, err := ipToInt(parentIPNet.IP)
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, parentIPNet.IP)
	}

	parentLastIP, err := ipToInt(lastIP(parentIPNet))
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, lastIP(parentIPNet))
	}

	childFirstIP, err := ipToInt(childIPNet.IP)
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, childIPNet.IP)
	}

	childLastIP, err := ipToInt(lastIP(childIPNet))
	if err != nil {
		return false, fmt.Errorf("failed to transform ip to int. %v ip: %v", err, lastIP(childIPNet))
	}

	return parentFirstIP <= childFirstIP && parentLastIP >= childLastIP, nil
}

func ipToInt(ip net.IP) (uint32, error) {
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, errors.New("Not an IPv4 address")
	}
	return binary.BigEndian.Uint32(ip4), nil
}

func lastIP(network *net.IPNet) net.IP {
	ip := network.IP
	mask := network.Mask

	// Create a copy of the IP to avoid modifying the original
	lastIP := make(net.IP, len(ip))
	copy(lastIP, ip)

	// Bitwise OR the inverted mask with the network IP
	for i := 0; i < len(ip); i++ {
		lastIP[i] |= ^mask[i]
	}

	return lastIP
}

func OverlapsExistingCIDRs(ctx context.Context, db *sql.Tx, subnet *pb.SubnetPrivate) (bool, error) {
	query := `
		select count(*)
		from   subnet
		where  cloud_account_id = $1
		  and  value->'spec'->>'vpcId' = $2
		  and  (value->'spec'->>'cidrBlock')::inet && $3::inet
		  and  deleted_timestamp = $4
	`

	var count int

	err := db.QueryRowContext(ctx, query, subnet.Metadata.CloudAccountId, subnet.Spec.VpcId, subnet.Spec.CidrBlock, common.TimestampInfinityStr).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to execute query: %v", err)
	}

	return count > 0, nil
}
