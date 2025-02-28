package loadbalancer

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type loadBalancerType string
type loadBalancerMonitorType string

const (
	LBTypeExternal loadBalancerType = "external"
	LBTypeInternal loadBalancerType = "internal"
)

const (
	LBMonitorTypeTCP   = "tcp"
	LBMonitorTypeHTTP  = "http"
	LBMonitorTypeHTTPS = "https"
)

// validateLoadBalancerName is valid when name is starting and ending with lowercase alphanumeric
// and contains lowercase alphanumeric, '-' characters and should have at most 63 characters
func validateLoadBalancerName(name string) error {
	re := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	matches := re.FindAllString(name, -1)
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid load balancer name")
	}
	return nil
}

// validateInstanceSelector verifies the instance selector is valid and defined.
func validateInstanceSelector(listener *pb.LoadBalancerListener) error {

	if listener.Pool == nil {
		return status.Error(codes.InvalidArgument, "missing pool configuration")
	}

	selectors := listener.Pool.InstanceSelectors
	instances := listener.Pool.InstanceResourceIds

	// Validate that either InstanceSelectors is defined or Instances
	if len(selectors) == 0 && len(instances) == 0 {
		return status.Error(codes.InvalidArgument, "pool configuration of InstanceSelector or Instances is required")
	}

	// Validate that both InstanceSelectors & Instances are not defined
	if len(selectors) > 0 && len(instances) > 0 {
		return status.Error(codes.InvalidArgument, "pool configuration only supports one of InstanceSelector or Instances")
	}

	if err := utils.ValidateLabels(selectors); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate the port is valid.
	if err := validateLBPort(listener.Pool.Port); err != nil {
		return err
	}

	return nil
}

// Validate the load balancer monitoring type is tcp , http, or https, error if anything else.
func validateLBMonitoringType(monitorType pb.LoadBalancerMonitorType) error {
	switch monitorType {
	case pb.LoadBalancerMonitorType_tcp, pb.LoadBalancerMonitorType_http, pb.LoadBalancerMonitorType_https:
		return nil
	}
	return status.Error(codes.InvalidArgument, "a valid load balancer monitoring type is required")
}

// Validate the load balancer monitoring type is tcp , http, or https, error if anything else.
func validateLBMModeType(mode pb.LoadBalancingMode) error {
	switch mode {
	case pb.LoadBalancingMode_leastConnectionsMember, pb.LoadBalancingMode_roundRobin:
		return nil
	}
	return status.Error(codes.InvalidArgument, "a valid load balancer mode type is required")
}

// Validate the load balancer port is valid.
func validateLBPort(port int32) error {
	if port <= 0 || port > 65535 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("port \"%d\" is out of range", port))
	}
	return nil
}

// Validate the members of the pool are valid and map to Instances in the tenant account
func (lb *Service) validateListenerPoolMembers(ctx context.Context, cloudAccountId string, lbSpec *pb.LoadBalancerSpecPrivate) error {

	// Validate the instanceID is a valid GUID
	for _, listener := range lbSpec.Listeners {
		for _, resourceId := range listener.Pool.InstanceResourceIds {
			if !utils.IsValidUUID(resourceId) {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid instance resourceid %s in listener pool", resourceId))
			}

			// Validate the instanceId passed in are part of this cloud account. A user should not be able
			// to assign an instance from another account.
			_, err := lb.instanceService.Get(ctx, &pb.InstanceGetRequest{
				Metadata: &pb.InstanceMetadataReference{
					CloudAccountId: cloudAccountId,
					NameOrId: &pb.InstanceMetadataReference_ResourceId{
						ResourceId: resourceId,
					},
				},
			})
			if err != nil {
				// If the Instance doesn't exist, the get will return a NotFound error.
				if errors.Is(err, status.Error(codes.NotFound, "resource not found")) {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid instance resourceid %s in listener pool", resourceId))
				}
				return err
			}
		}
	}

	return nil
}

// Validate the source IPs are valid.
func validateSourceIPs(sourceips []string) error {

	for _, ip := range sourceips {
		if strings.ToLower(ip) == "any" {
			// It's only valid to have "any" OR any number of source IPs.
			// If an "any" is encountered, validate the rest of the list is only size 1,
			// meaning it can only have a single "any" entry.
			if len(sourceips) > 1 {
				return status.Error(codes.InvalidArgument, "source ip \"any\" is only valid when no other source ips are defined")
			}
			return nil
		}
		if strings.Contains(ip, "/") {
			_, ipv4Net, err := net.ParseCIDR(ip)
			if err != nil || ipv4Net.String() != ip {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("cidr format %q is invalid", ip))
			}
		} else {
			if net.ParseIP(ip) == nil {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("source ip %q is invalid", ip))
			}
		}
	}

	return nil
}
