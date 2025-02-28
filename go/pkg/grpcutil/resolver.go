// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"fmt"
	"net"
)

// Resolver is used by services to look up the hostname/port of other
// services it wants to call.
//
// When running as a stand-alone service, resolver looks up the SRV records
// for a service to determine the hostname and port to use.
//
// When running multiple services embedded together in a test program,
// a test resolver is used that maps from each service's name to localhost
// and the port that got randomly assigned to it.
type Resolver interface {
	Resolve(ctx context.Context, name string) (string, error)
}

type DnsResolver struct{}

// Lookup the port in the SRV record for the service named "name"
func (*DnsResolver) Resolve(ctx context.Context, name string) (string, error) {
	_, srvs, err := net.LookupSRV("grpc", "tcp", name)
	if err != nil {
		return "", err
	}
	if len(srvs) == 0 {
		return "", fmt.Errorf("%v not found", name)
	}
	// Return the FQDN and port. Don't return the IP address because
	// it can change when the pod gets redeployed. Returning FQDN
	// enables the hostname to match the no_proxy environment variable
	// to avoid using an external proxy
	return fmt.Sprintf("%v.idcs-system.svc.cluster.local:%v", name, srvs[0].Port), nil
}
