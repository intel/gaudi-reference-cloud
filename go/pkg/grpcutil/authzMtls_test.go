// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"testing"

	"crypto/x509/pkix"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestGrpcAuthzServerInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		authzEnabled   bool
		allowedCallers string
		peerCertOUs    []string
		expectedError  codes.Code
	}{
		{
			name:           "AuthorizedCaller",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{"grpc-proxy-external"},
			expectedError:  codes.OK,
		},
		{
			name:           "UnauthorizedCaller",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{"cloudaccount"},
			expectedError:  codes.PermissionDenied,
		},
		{
			name:           "NoClientCertificate",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{},
			expectedError:  codes.InvalidArgument,
		},
		{
			name:           "AuthzDisabled",
			authzEnabled:   false,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{"cloudaccount-enroll"},
			expectedError:  codes.OK,
		},
		{
			name:           "MultipleAuthorizedCallers",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external,grpc-proxy-internal",
			peerCertOUs:    []string{"grpc-proxy-external", "grpc-proxy-internal"},
			expectedError:  codes.OK,
		},
		{
			name:           "MultipleUnauthorizedCallers",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{"cloudaccount", "cloudaccount-enroll"},
			expectedError:  codes.PermissionDenied,
		},
		{
			name:           "MixedAuthorizedAndUnauthorizedCallers",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{"grpc-proxy-external", "cloudaccount"},
			expectedError:  codes.PermissionDenied,
		},
		{
			name:           "EmptyAllowedCallers",
			authzEnabled:   true,
			allowedCallers: "",
			peerCertOUs:    []string{"cloudaccount-enroll"},
			expectedError:  codes.PermissionDenied,
		},
		{
			name:           "NilPeerInformation",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    nil,
			expectedError:  codes.NotFound,
		},
		{
			name:           "InvalidPeerTransportCredentials",
			authzEnabled:   true,
			allowedCallers: "grpc-proxy-external",
			peerCertOUs:    []string{"grpc-proxy-external"},
			expectedError:  codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize with mock values
			initialize(tt.authzEnabled, tt.allowedCallers)

			// Create a mock context with peer information
			ctx := context.Background()
			if tt.name == "NilPeerInformation" {
				// Do not set peer information
			} else if tt.name == "InvalidPeerTransportCredentials" {
				p := &peer.Peer{
					AuthInfo: nil, // Invalid transport credentials
				}
				ctx = peer.NewContext(ctx, p)
			} else {
				peerCerts := []*x509.Certificate{}
				for _, ou := range tt.peerCertOUs {
					peerCerts = append(peerCerts, &x509.Certificate{
						Subject: pkix.Name{
							OrganizationalUnit: []string{ou},
						},
					})
				}
				tlsInfo := credentials.TLSInfo{
					State: tls.ConnectionState{
						PeerCertificates: peerCerts,
					},
				}
				p := &peer.Peer{
					AuthInfo: tlsInfo,
				}
				ctx = peer.NewContext(ctx, p)
			}

			// Create a mock handler
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			}

			// Call the interceptor
			interceptor := GrpcAuthzServerInterceptor()
			_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			// Check the error code
			if status.Code(err) != tt.expectedError {
				t.Errorf("expected error code %v, got %v", tt.expectedError, status.Code(err))
			}
		})
	}
}
