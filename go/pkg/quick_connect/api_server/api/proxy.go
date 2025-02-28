// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
)

const (
	networkTimeoutSeconds = 30
	TargetHostKey         = "x-target-host"
)

type QuickConnectProxyTarget struct {
	URL       *url.URL
	Transport *http.Transport
}

type QuickConnectProxy struct {
	computeApiServerClient *resty.Client
	vaultClient            secrets.SecretManager
	marshaler              *runtime.JSONPb
	// targetTransport is a cache of the client transport used to communicate with the service
	// in the instance (e.g. JupyterLab). The targetTransport is cached to encourage connection
	// reuse and avoid issuing client certificates more than necessary.
	targetTransport             *http.Transport
	targetPort                  string
	targetClientCertRefreshTime time.Time
	targetClientCertCommonName  string
	targetClientCertTTL         time.Duration
	mut                         sync.Mutex
}

func NewQuickConnectProxy(ctx context.Context, cfg *Config, tlsClientConfig *tls.Config, vaultClient *secrets.Vault) (*QuickConnectProxy, error) {
	proxy := &QuickConnectProxy{
		targetPort:  fmt.Sprintf("%d", cfg.TargetPort),
		vaultClient: vaultClient,
		computeApiServerClient: resty.New().
			SetTLSClientConfig(tlsClientConfig).
			SetBaseURL(fmt.Sprintf("https://%s", cfg.ComputeApiServerAddress)),
		marshaler: &runtime.JSONPb{
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
		targetClientCertCommonName: cfg.ClientCertificate.CommonName,
		targetClientCertTTL:        cfg.ClientCertificate.TTL,
	}
	return proxy, nil
}

func (p *QuickConnectProxy) AddRoutes(router *gin.Engine) error {
	router.Use(customRecoveryMiddleware())

	// Helper functions to add route handlers.
	roMethods := []string{
		http.MethodGet,
		http.MethodHead,
	}
	rwMethods := append([]string{
		http.MethodDelete,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
	}, roMethods...)
	handle := func(methods []string, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
		for _, method := range methods {
			router.Handle(method, relativePath, handlers...)
		}
		return router
	}

	handle(rwMethods, "/v1/connect/:cloudaccountid/:instanceid/*path", p.ParameterValidator, p.CloudAccountAuthorizer, p.PassthroughHandler)

	return nil
}

func (p *QuickConnectProxy) ParameterValidator(ctx *gin.Context) {
	if !cloudaccount.IsValidId(ctx.Param("cloudaccountid")) {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	if !utils.IsValidResourceID(ctx.Param("instanceid")) {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
}

// Middleware to ensure that the user is authorized for the cloud account in the request.
//
// Access is granted when the authenticated token provided by the user can be used to
// lookup the requested cloud account and instance and the spec indicates that quick connect
// is enabled for the instance.
func (p *QuickConnectProxy) CloudAccountAuthorizer(ctx *gin.Context) {
	log := log.FromContext(ctx)
	cloudAccountId := ctx.Param("cloudaccountid")
	instanceId := ctx.Param("instanceid")
	err := func() error {
		auth := strings.Split(ctx.Request.Header.Get("Authorization"), " ")
		if len(auth) != 2 || auth[0] != "Bearer" {
			return errors.New("missing token")
		}
		token := auth[1]

		resp, err := p.computeApiServerClient.R().
			SetAuthToken(token).
			Get(fmt.Sprintf("/v1/cloudaccounts/%s/instances/id/%s", cloudAccountId, instanceId))
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("failed to get instance, received status %v - %v", resp.Status(), string(resp.Body()))
		}
		var instance *pb.Instance
		if err := p.marshaler.Unmarshal(resp.Body(), &instance); err != nil {
			return err
		}
		if instance.Spec.QuickConnectEnabled != pb.TriState_True {
			return errors.New("quick connect not enabled")
		}

		if len(instance.Status.Interfaces) == 0 || len(instance.Status.Interfaces[0].Addresses) == 0 {
			return fmt.Errorf("CloudAccountAuthorizer: Instance does not have an IP address; instance.Status.Interfaces=%v", instance.Status.Interfaces)
		}
		host := instance.Status.Interfaces[0].Addresses[0]
		log.Info("getTargetForRequest", "cloudAccountId", cloudAccountId, "instanceId", instanceId, "host", host)
		// Now that the request is authorized, forward the host to the next handler in the
		// chain which will actually serve the request.
		ctx.Set(TargetHostKey, host)
		return nil
	}()
	if err != nil {
		log.Info("Unauthorized", "cloudAccountId", cloudAccountId, "instanceId", instanceId, "err", err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	log.V(9).Info("Authorized", "cloudAccountId", cloudAccountId, "instanceId", instanceId, "err", err)
}

func (p *QuickConnectProxy) getTransportForRequest(ctx *gin.Context) (*http.Transport, error) {
	p.mut.Lock()
	defer p.mut.Unlock()

	if p.targetTransport == nil || time.Now().After(p.targetClientCertRefreshTime) {
		cert, err := p.vaultClient.IssueQuickConnectCertificate(ctx, p.targetClientCertCommonName, p.targetClientCertTTL)
		if err != nil {
			return nil, fmt.Errorf("unable to issue Quick Connect client certificate: %w", err)
		}
		// Refresh the client certificate when we get near the expiration time. The current delta
		// is 30s and this may need to be tuned further.
		p.targetClientCertRefreshTime = time.Now().Add(p.targetClientCertTTL - 30*time.Second)
		p.targetTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   networkTimeoutSeconds * time.Second,
				KeepAlive: networkTimeoutSeconds * time.Second,
			}).Dial,
			TLSHandshakeTimeout: networkTimeoutSeconds * time.Second,
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*cert},
				// Instance server certificates are self-signed, verification must be skipped.
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS13,
			},
		}
	}

	return p.targetTransport, nil
}

// A request from a user will be mapped to at most one target.
func (p *QuickConnectProxy) getTargetForRequest(ctx *gin.Context) (*QuickConnectProxyTarget, error) {
	host := ctx.GetString(TargetHostKey)
	url := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(host, p.targetPort),
	}
	transport, err := p.getTransportForRequest(ctx)
	if err != nil {
		return nil, err
	}
	return &QuickConnectProxyTarget{
		URL:       url,
		Transport: transport,
	}, nil
}

func (p *QuickConnectProxy) PassthroughHandler(ctx *gin.Context) {
	log := log.FromContext(ctx)
	err := func() error {
		target, err := p.getTargetForRequest(ctx)
		if err != nil {
			return err
		}

		proxy := httputil.NewSingleHostReverseProxy(target.URL)
		proxy.Transport = target.Transport
		proxy.Director = func(req *http.Request) {
			log.V(9).Info("PassthroughHandler request IN", "method", req.Method, "url", req.URL.String(), "header", req.Header)
			// Mutate request for target
			sanitizeRequest(ctx, req)
			req.Host = target.URL.Host
			req.URL.Scheme = target.URL.Scheme
			req.URL.Host = target.URL.Host
			log.V(9).Info("PassthroughHandler request OUT", "method", req.Method, "url", req.URL.String(), "header", req.Header)
		}
		// The Instance may be reported as Ready before JupyterLab is up and running. Rather than
		// return 502 Bad Gateway, request the browser to refresh.
		//
		// Note that if the downstream Envoy container times out a pending request, the user will
		// see a 504 Gateway Timeout and the err reported here will be "context cancelled". The
		// ErrorHandler will have no effect on what the user sees as Envoy has already sent a response.
		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			log.Info("ErrorHandler", "URL", req.URL.String(), "err", err)
			rw.Write([]byte(`<!DOCTYPE html>
<meta http-equiv="refresh" content="10"/>
<html lang="en">
	<body>
		Waiting for JupyterLab to be ready...
	</body>
</html>`))
		}
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
		return nil
	}()
	if err != nil {
		log.Error(err, "Internal Server Error")
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

// This prevents logging of errors when clients disconnect while watching resources.
// Based on https://github.com/golang/go/issues/56228#issuecomment-1290730592
func customRecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if p := recover(); p != nil {
				if err, ok := p.(error); ok {
					// Ignore panic abort handler for text/event-stream server-side-events (SSE).
					if errors.Is(err, http.ErrAbortHandler) {
						return
					}
				}
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}

func sanitizeRequest(ctx *gin.Context, req *http.Request) {
	sanitizeHeaders(ctx, req)
	// TODO: Ensure that all other request fields are sanitized.
}

// This removes all headers except for those listed below.
// Only these headers will be passed-through unmodified from the caller to the target server.
func sanitizeHeaders(ctx *gin.Context, req *http.Request) {
	log := log.FromContext(ctx)
	originalHeader := req.Header
	log.V(9).Info("sanitizeHeaders", "originalHeader", originalHeader)
	req.Header = make(http.Header)
	allowedHeaders := []string{
		"Accept",
		"Content-Type",
		"X-Xsrftoken",
		// The headers below are the forbidden request headers: these are under full control
		// of the user agent. See https://fetch.spec.whatwg.org/#forbidden-request-header.
		"Accept-Charset",
		"Accept-Encoding",
		"Access-Control-Request-Headers",
		"Access-Control-Request-Method",
		"Connection",
		"Content-Length",
		"Cookie",
		"Cookie2",
		"Date",
		"DNT",
		"Expect",
		"Host",
		"Keep-Alive",
		"Origin",
		"Referer",
		// Set-Cookie is handled by the http.Request separate from the Header field
		"TE",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
		"Via",
	}
	for _, allowedHeader := range allowedHeaders {
		headerValues := originalHeader.Values(allowedHeader)
		if len(headerValues) > 0 {
			req.Header[allowedHeader] = headerValues
		}
	}
	// The list of forbidden headers also includes any with prefix of Proxy- or Sec-.
	for header := range originalHeader {
		if strings.HasPrefix(header, "Proxy-") || strings.HasPrefix(header, "Sec-") {
			req.Header[header] = originalHeader.Values(header)
		}
	}
	log.V(9).Info("sanitizeHeaders", "newHeader", req.Header)
}
