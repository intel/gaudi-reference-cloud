// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"bytes"
	"flag"
	"html/template"
	"io"
	"log"
	"os"
	"path"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/build/tools/util"
)

var allowTempl = `package envoy.authz

import future.keywords.in
import input.attributes.request.http

default allow = false 
result["allowed"] := allow

authz := {"scheme":scheme, "token":payload} {
	 [scheme, encoded] := split(http.headers["authorization"], " ")
	 [_, payload, _] := io.jwt.decode(encoded)
 }

scopes := split(authz.token.scope, " ")
	
# For tracability/debugging
result["scopes"] := scopes

# TODO: move to .proto file definitions
external_client_path_scopes_map := {
	"/proto.CloudCreditsCouponService/Read": "/coupons:view",
	"/proto.BillingCouponService/Read": "/coupons:view",
	"/proto.CloudCreditsCouponService/Create": "/coupons:create",
	"/proto.BillingCouponService/Create": "/coupons:create",
}

internal_client_scopes := [
	"/bucket-metering-monitor",
	"/compute-api-server", 
	"/iks-api-server",
	"/storage-admin-api-server",
	"/storage-api-server", 
	"/storage-metering-monitor", 
	"/storage-resource-cleaner",
	"/grpc-proxy", 
	"/write", # aria
	"/compute-metering-monitor",
	"/training-api-server",
	"internal/read"
]

allow {
	some scope in scopes
	some internal_client_scope in internal_client_scopes
	contains(scope, internal_client_scope)
}

allow {
	required_scope := external_client_path_scopes_map[http.path]
	some i
	contains(scopes[i], required_scope)
}
`

type tData struct {
	ConfigMap     bool
	HasSameRegion bool
	Server        string
}

var (
	configmap  string
	includeDir string
)

func init() {
	flag.StringVar(&configmap, "configmap", "", "configmap output file")
	flag.StringVar(&includeDir, "I", "", "include dir")
}

var tool string

func main() {
	tool = path.Base(os.Args[0])
	flag.Parse()
	if configmap == "" {
		log.Fatal("--configmap is required")
	}

	data := tData{}

	tmpl, err := template.New("authz").Parse(allowTempl)
	if err != nil {
		log.Fatal(err)
	}

	outputConfigMap(tmpl, &data)
}

func outputConfigMap(tmpl *template.Template, data *tData) {
	if configmap == "" {
		return
	}
	data.ConfigMap = true
	data.HasSameRegion = true
	defer func() { data.ConfigMap = false }()
	if err := util.WriteFileAtomically(configmap,
		func(outf io.Writer) error {
			buf := bytes.Buffer{}
			if err := tmpl.Execute(&buf, data); err != nil {
				return err
			}
			return util.WriteConfigMap(outf, tool, "authzservice", buf.Bytes())
		}); err != nil {
		log.Fatal(err)
	}
}
