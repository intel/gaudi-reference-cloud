// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/grpc/backoff"

	"github.com/friendsofgo/errors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	addrDevHost       = "dispatcher.intel.com:443" // my dev cluster
	addrDevHostIP     = "146.152.224.83:443"       // my dev cluster
	addrPreProdHostIP = "146.152.225.162:443"      // public preprod
	addrProdHostIP    = "146.152.224.101:443"      // public prod
	addrProdHost      = "maas-dispatcher.idcmgt.intel.com:443"
	addrPreProdHost   = "maas-dispatcher.idcstage.intel.com:443"

	addrStgNewHostIP = "146.152.227.169"
)

func NewRemoteGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remotegen [target json]",
		Short: "runs generate prompt remotely",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			target := &targetConfig{
				Addr:             addrStgNewHostIP,
				CertificatesPath: "go/pkg/infaas-dispatcher/deployment/",
				MaxTokens:        30,
				Prompt:           "what is the meaning of life?",
				Model:            "meta-llama/Meta-Llama-3.1-8B-Instruct",
			}

			if len(args) != 0 && args[0] != "" {
				if err := json.Unmarshal([]byte(args[0]), target); err != nil {
					return errors.Wrap(err, "failed to read target def")
				}
			}
			fmt.Printf("---> Using target config: %+v\n", target)

			tlsCreds, err := loadTLSConfig(target)
			if err != nil {
				return errors.Wrap(err, "failed to load certificates")
			}

			//tlsCreds := insecure.NewCredentials()
			//tlsCreds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

			fmt.Println("dialing...")
			conn, err := grpc.NewClient(
				target.Addr,
				grpc.WithTransportCredentials(tlsCreds),
				grpc.WithMaxCallAttempts(5),
				grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.DefaultConfig}),
			)
			if err != nil {
				return errors.Wrap(err, "failed to connect to the service dispatcher")
			}

			client := pb.NewDispatcherClient(conn)
			req := &pb.DispatcherRequest{
				RequestID: "remote-request-test",
				Model:     target.Model,
				Request: &pb.GenerateStreamRequest{
					Prompt: target.Prompt,
					Params: &pb.GenerateRequestParameters{
						MaxNewTokens: target.MaxTokens,
					}},
			}

			fmt.Println("inferring...")
			respStream, err := client.GenerateStream(clientCtx, req /*, grpc.WaitForReady(true)*/)
			if err != nil {
				return errors.Wrap(err, "failed to call Generate()")
			}
			for {
				resp, err := respStream.Recv()
				if err != nil {
					if err == io.EOF {
						fmt.Println("End of response stream!")
						return nil
					}
					return errors.Wrap(err, "failed to receive response chunk")
				}
				fmt.Printf("%+v\n", resp.Response.Token)
			}
		},
	}

	return cmd
}

func loadTLSConfig(target *targetConfig) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(target.clientCertFile(), target.clientKeyFile())
	if err != nil {
		return nil, errors.Wrap(err, "failed to load client certification")
	}

	ca, err := os.ReadFile(target.caCertFile())
	if err != nil {
		return nil, errors.Wrap(err, "faild to read CA certificate")
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, errors.Wrap(err, "faild to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		RootCAs:            capool,
		InsecureSkipVerify: true,
		ServerName:         target.Addr,
	}

	return credentials.NewTLS(tlsConfig), nil
}

type targetConfig struct {
	Addr             string
	CertificatesPath string
	MaxTokens        uint32
	Prompt           string
	Model            string
}

func (t *targetConfig) clientCertFile() string {
	return t.CertificatesPath + "cert.pem"
}

func (t *targetConfig) caCertFile() string {
	return t.CertificatesPath + "ca.pem"
}

func (t *targetConfig) clientKeyFile() string {
	return t.CertificatesPath + "cert.key"
}
