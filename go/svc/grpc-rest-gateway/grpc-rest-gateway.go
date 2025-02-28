// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flowchartsman/swaggerui"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpc_rest_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	maxReqBodySize  = 16 * 1024 * 1024
	numOfReqHeaders = 30
	// DefaultMaxHeaderBytes is the maximum permitted size of the headers in an HTTP request.
	// Default value of DefaultMaxHeaderBytes = 1MB
	maxReqHeaderSize = http.DefaultMaxHeaderBytes
	// maximum duration for reading the entire request
	readTimeout = 30 * time.Second
	// maximum duration before timing out writes of the response
	writeTimeout = 40 * time.Second
	// maximum amount of time to wait for the next request when keep-alives are enabled.
	// if idleTimeout is zero, the value of readTimeout is used
	idleTimeout = 10 * time.Minute
)

var config = grpc_rest_gateway.Config{
	ListenPort:     8080,
	TargetAddr:     "grpc-proxy:8080",
	Deployment:     "all",
	AllowedOrigins: []string{},
}

var (
	openApiPath       = "/openapiv2/"
	swaggerEnabled, _ = strconv.ParseBool(os.Getenv("SWAGGER_ENABLED"))
)

func loadSwagger(config *grpc_rest_gateway.Config) ([]byte, error) {
	fd, err := pb.SwaggerFs.Open(fmt.Sprintf("%s/%s.swagger.json", pb.SwaggerDir, config.Deployment))
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Chain http handlers.
func filterRequest(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enable content-type validation
		if r.Method != http.MethodGet && r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid Content-type", http.StatusUnsupportedMediaType)
			return
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read request body", http.StatusBadRequest)
			return
		}
		// set the body with the same data.
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		//Validate if body is a valid utf8.
		if !utf8.Valid(b) {
			http.Error(w, "Invalid UTF-8", http.StatusBadRequest)
			return
		}

		// Limit the maximum number of request headers
		if len(r.Header) > numOfReqHeaders {
			http.Error(w, "Too many headers", http.StatusBadRequest)
			return
		}

		// Limit the maximum size of request body
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxReqBodySize))

		next.ServeHTTP(w, r)
	})
}

func RunService(ctx context.Context, config *grpc_rest_gateway.Config) error {
	log := log.FromContext(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Enable keepalive on the gRPC-gateway
	opts := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    120 * time.Second, // Send pings every 120 seconds
		Timeout: 10 * time.Second,  // Consider a connection dead if no response within 30 seconds
	})

	conn, err := grpcutil.NewClient(ctx, config.TargetAddr, opts)
	if err != nil {
		return fmt.Errorf("TLS Client %v: %w", "grpc-proxy", err)
	}
	go func() {
		conn := conn
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			log.Error(err, "gRPC client close")
		}
	}()

	mux := http.NewServeMux()
	if swaggerEnabled {
		swaggerData, err := loadSwagger(config)
		if err != nil {
			return fmt.Errorf("error loading swagger data: %w", err)
		}
		mux.Handle(openApiPath, http.StripPrefix(openApiPath[:len(openApiPath)-1], swaggerui.Handler(swaggerData)))
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					// discard unknown field in the json to enable UI/API compatibility
					// for multiple regions enabled with diff version of APIs
					DiscardUnknown: true,
				},
			},
		}),
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			originalURL := req.URL.String()
			originalMethod := req.Method
			// Create the metadata with the "x-original-http-path" header set to the original URL
			return metadata.New(map[string]string{
				"x-original-http-path":   originalURL,
				"x-original-http-method": originalMethod,
			})
		}),
	)
	registerServers(ctx, conn, gwmux)
	mux.Handle("/", filterRequest(gwmux))

	tlsConfig, err := tlsutil.NewTlsProvider().ServerTlsConfig(ctx)
	if err != nil {
		return err
	}

	// Get list of allowed FQDNs
	allowedOrigins := config.AllowedOrigins

	srv := &http.Server{
		Addr: fmt.Sprintf(":%v", config.ListenPort),
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			log.Info("Handling request", "RemoteAddr", req.RemoteAddr, "URL", req.URL)
			origin := req.Header.Get("Origin")
			isOriginAllowed := false
			log.Info("Handling request", "RequestMethod", req.Method, "lengthAllowedOrigins:", len(allowedOrigins))
			if len(allowedOrigins) == 0 || req.Method == "GET" {
				isOriginAllowed = true
			} else {
				// Check if the origin is in the list of allowed FQDNs
				log.Info("Handling request", "RequestMethod", req.Method)
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						isOriginAllowed = true
						break
					}
				}
			}
			// Set the appropriate CORS headers
			if isOriginAllowed {
				resp.Header().Set("Access-Control-Allow-Origin", origin)
				if req.Method == "OPTIONS" && req.Header.Get("Access-Control-Request-Method") != "" {
					headers := []string{"Content-Type", "Accept", "Authorization"}
					resp.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
					methods := []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
					resp.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
					return
				}
			}
			mux.ServeHTTP(resp, req)
		}),
		TLSConfig:         tlsConfig,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		MaxHeaderBytes:    maxReqHeaderSize,
		IdleTimeout:       idleTimeout,
	}
	log.Info("Service running")
	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.Info("ListenAndServeTLS", "err", err)
	}
	return err
}

func main() {
	ctx := context.Background()

	// Parse command line.
	var configFile string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	if configFile != "" {
		if err := conf.LoadConfigFile(ctx, configFile, &config); err != nil {
			log.Error(err, "fatal error", "configFile", configFile)
			os.Exit(1)
		}
	}

	if err := RunService(ctx, &config); err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
