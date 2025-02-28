// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"crypto/x509"
	"strings"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	GrpcTlsAuthzEnabled        bool
	GrpcTlsAuthzAllowedCallers map[string]struct{}
)

func init() {
	// Initialize with the default TlsProvider
	tlsProvider := tlsutil.NewTlsProvider()
	initialize(tlsProvider.GrpcTlsAuthzEnabled(), tlsProvider.GrpcTlsAuthzAllowedCallers())
}

func initialize(authzEnabled bool, allowedCallersEnv string) {
	GrpcTlsAuthzEnabled = authzEnabled
	GrpcTlsAuthzAllowedCallers = initializeAllowedCallers(allowedCallersEnv)
}

func initializeAllowedCallers(allowedCallersEnv string) map[string]struct{} {
	allowedCallers := make(map[string]struct{})
	callers := strings.Split(allowedCallersEnv, ",")
	for _, caller := range callers {
		allowedCallers[caller] = struct{}{}
	}
	return allowedCallers
}

// GrpcAuthzServerInterceptor intercepts and extracts and verify the OU from the client certificate
func GrpcAuthzServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if GrpcTlsAuthzEnabled {
			ctx, log, _ := obs.LogAndSpanFromContext(ctx).WithName("grpcutil.GrpcAuthzServerInterceptor").Start()
			// Extract the OU from the peer certificate
			p, ok := peer.FromContext(ctx)
			if !ok {
				err := status.Errorf(codes.NotFound, "no peer found")
				log.Error(err, "no peer found")
				return nil, err
			}

			tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
			if !ok {
				err := status.Errorf(codes.FailedPrecondition, "unexpected peer transport credentials")
				log.Error(err, "unexpected peer transport credentials")
				return nil, err
			}

			if len(tlsInfo.State.PeerCertificates) == 0 {
				err := status.Errorf(codes.InvalidArgument, "no client certificate provided")
				log.Error(err, "no client certificate provided")
				return nil, err
			}

			clientOUs := extractClientOUs(tlsInfo.State.PeerCertificates)
			// If authorization is enabled, check each client OU individually
			for ou := range clientOUs {
				if !isAuthorized(ou) {
					err := status.Errorf(codes.PermissionDenied, "client: %s, is not authorized", ou)
					log.Error(err, "client is not authorized", "client", ou)
					return nil, err
				}
			}
		}
		return handler(ctx, req)
	}
}

// extractClientOUs extracts the Organizational Units (OUs) from the client certificates
func extractClientOUs(peerCertificates []*x509.Certificate) map[string]struct{} {
	clientOUs := make(map[string]struct{})
	for _, clientCert := range peerCertificates {
		for _, ou := range clientCert.Subject.OrganizationalUnit {
			clientOUs[ou] = struct{}{}
		}
	}
	return clientOUs
}

// isAuthorized checks if the OU is authorized
func isAuthorized(ou string) bool {
	_, ok := GrpcTlsAuthzAllowedCallers[ou]
	return ok
}
