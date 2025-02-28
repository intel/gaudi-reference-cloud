// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// generate .rego code for user authz policies in OPA
package main

import (
	"bytes"
	"flag"
	"fmt"
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
authz := {"scheme":scheme, "token":payload} {
 	[scheme, encoded] := split(http.headers.authorization, " ")
 	# The header and signature are ignored. The JWT has already been validated by
 	# the ingress. We are using the JWT for information about the caller, not for
 	# authentication.
 	[_, payload, _] := io.jwt.decode(encoded)
}
     
clientId := authz.token.client_id

# Get current cloudaccount associated with access_token
cloudAccount := cloudaccount.getAppClientCloudAccount(clientId)
cloudAccountId := cloudAccount["id"]
 
# getAppClientCloudAccount returns owner_email - if owner's token
# getAppClientCloudAccount returns member_email - if member's token
email := cloudAccount["name"]

# countryCode of the user (owner or member)
countryCode := cloudAccount["countryCode"]

# get current user associated CloudAccounts
response := cloudaccount.getRelatedCloudAccounts(email)
relatedCloudAccounts := response["relatedAccounts"]

result["allowed"] := allow

result["body"] := "User is restricted" {
	method_ok
	not user_ok
	not allow
}

# product_ok and gts_ok enabled only in prod and staging
{{ print "{{-"}} if not (eq $.Values.environmentName "kind-idc-global") {{ print "}}"}}
result["body"] := "product not found" {
	product_need_custom_body
	prod == {}
}
result["body"] := "paid service not allowed" {
	product_access_ok
	product_need_custom_body
	prod != {}
}
result["body"] := "product access not allowed" {
	need_product_match
	not product_access_ok
	prod != {}
}
result["body"] := "gts check failed" {
	need_gts_match
	method_ok
	user_ok
	product_ok
	product_access_ok
	not gts_ok
}

product_need_custom_body {
	need_product_match
	method_ok
	user_ok
	not product_ok
}
{{ print "{{- end }}"}}

default allow = false
default prod = {}
default need_product_match = false
default need_gts_match = false
default method_ok = false
default user_ok = false
default product_ok = false
default product_access_ok = false

# users allowed (owner or member)
allow {
	email != ""
	method_ok
	user_ok
{{ print "{{-"}} if not (eq $.Values.environmentName "kind-idc-global") {{ print "}}"}}
    product_ok
	product_access_ok
	gts_ok
{{ print "{{- end }}"}}
}

# gts_ok by default for intel users
gts_ok {
	endswith(email, "@intel.com")
}

{{- range .Methods }}
{{ if $.ConfigMap }}
{{ print "{{-"}} if or (eq $.Values.deployment "all") (eq $.Values.deployment "{{.Deploy}}") {{ print "}}"}}
{{- end}}

{{- if .AuthzCheck }}
{{ print "{{-"}} if $.Values.authz.enabled {{ print "}}"}}
{{- if .AuthzCheckField }}
{{ .Service }}_{{ .Method }}_authZCloudAccountId := input.parsed_body.{{ .AuthzCheckField }}
{{- end }}
{{- if not .AuthzCheckField }}
{{ .Service }}_{{ .Method }}_authZCloudAccountId := "*"
{{- end }}
{{ .Service }}_{{ .Method }}_authZResult := authzService.check(email, clientId,{{ .Service }}_{{ .Method }}_authZCloudAccountId,input.attributes.request.http.headers["x-original-http-path"],input.attributes.request.http.headers["x-original-http-method"],input.parsed_body)
{{ print "{{- end }}"}}
{{- end }}

method_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .CloudAccountField }}
	# enforce access to cloudaccount-specific resources
	some _, relatedCloudAccount in relatedCloudAccounts
	input.parsed_body.{{ .CloudAccountField }} == relatedCloudAccount["id"]
{{- end }}
{{- if .UserNameField }}
	input.parsed_body.{{ .UserNameField }} == email
{{- end }}
{{- if .AuthzCheck }}
{{ print "{{-"}} if $.Values.authz.enabled {{ print "}}"}}
	{{ .Service }}_{{ .Method }}_authZResult  == true
{{ print "{{- end }}"}}
{{- end }}
}

{{- if .CloudAccountField }}
# Store Service.Method.cloudAccount based on CloudAccountField
{{ .Service }}_{{ .Method }}_cloudAccount := cloudaccount.getById(input.parsed_body.{{ .CloudAccountField }})

# Store personId for gts-check if owner/member
personId := {{ .Service }}_{{ .Method }}_cloudAccount["personId"] if {
	# email belongs to an owner
	email == {{ .Service }}_{{ .Method }}_cloudAccount["name"]
} else := cloudaccount.getMemberPersonId(email, {{ .Service }}_{{ .Method }}_cloudAccount["id"])
{{- end }}

user_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .CloudAccountField }}
	# admin will set restricted to true if this user needs to be restricted
	# use Service.Method.cloudAccount
	not {{ .Service }}_{{ .Method }}_cloudAccount["restricted"]
{{- end }}

{{- if .OwnerCheckField }}
	# check that the email belongs to the owner of the cloudaccount
	email == {{ .Service }}_{{ .Method }}_cloudAccount["name"]
{{- end }}
}

product_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .ProductNameField }}	
	# ProductNameField validation requires CloudAccountField
	prod["name"] != ""
	{{ .Service }}{{ .Method }}_product_ok
{{- end }}
}

{{ if .ProductAccessField }}
{{ .Service }}{{ .Method }}_product_access_ok {
	prod["access"] == "open"
}

{{ .Service }}{{ .Method }}_product_access_ok {
	checkProductAccess := productcatalog.checkProductAccess(input.parsed_body.{{ .ProductAccessField }}, {{ .Service }}_{{ .Method }}_cloudAccount["id"])
	checkProductAccess == true
}
{{ end }}

product_access_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
	{{- if .ProductAccessField }}
		{{ .Service }}{{ .Method }}_product_access_ok
	{{- end }}
}

{{- if .ProductNameField }}

prod := pp {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
	pp := productcatalog.getProductByName(input.parsed_body.{{.ProductNameField}}, {{ .Service }}_{{ .Method }}_cloudAccount["type"])
}

need_product_match {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
}

{{ .Service }}{{ .Method }}_product_ok {
	some ii
	to_number(prod.rates[ii].rate) == 0
}

{{ .Service }}{{ .Method }}_product_ok {
	{{ .Service }}_{{ .Method }}_cloudAccount["paidServicesAllowed"]
}

{{- end }}

gts_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .GTSCheckNameField }}
	# GTSCheckNameField validation requires CloudAccountField
	personId != ""
	countryCode != ""
	prodData["id"] != ""
	gts.isGTSOrderValid(prodData["id"], email, personId, countryCode)
{{- end }}
}

{{- if .GTSCheckNameField }}

prodData := pp {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
	pp := productcatalog.getProductByName(input.parsed_body.{{.GTSCheckNameField}}, {{ .Service }}_{{ .Method }}_cloudAccount["type"])
}

need_gts_match {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
}
{{- end }}


{{- if $.ConfigMap }}
{{ print "{{- end }}"}}
{{- end}}
{{- end }}
`

type tMethod struct {
	Deploy             string
	Service            string
	Method             string
	CloudAccountField  string
	AuthzCheck         bool
	AuthzCheckField    string
	ProductNameField   string
	ProductAccessField string
	GTSCheckNameField  string
	UserNameField      string
	OwnerCheckField    string
}

type tData struct {
	ConfigMap bool
	Methods   []tMethod
}

var (
	regoDir    string
	configmap  string
	includeDir string
)

func init() {
	flag.StringVar(&regoDir, "rego-dir", "", "rego output directory")
	flag.StringVar(&configmap, "configmap", "", "configmap output file")
	flag.StringVar(&includeDir, "I", "", "include dir")
}

var tool string

func main() {
	tool = path.Base(os.Args[0])
	flag.Parse()
	if regoDir == "" && configmap == "" ||
		regoDir != "" && configmap != "" {
		log.Fatal("specify --rego-dir or --configmap but not both")
	}
	tmplData := tData{}
	err := util.ForEachMethod(flag.Args(),
		func(info *util.MethInfo) error {
			if info.MethodOptions == nil || info.MethodOptions.Authz == nil {
				return nil
			}
			// skip the method if the minimum tagging on proto.service.function is not present
			if !info.MethodOptions.Authz.AppClientAccess {
				return nil
			}
			if !info.MethodOptions.Authz.User && !info.MethodOptions.Authz.CloudAccount && !info.MethodOptions.Authz.AuthzCheck {
				return nil
			}
			serviceName := info.Service.GetName()
			methName := info.Method.GetName()
			if info.MethodOptions.Authz.User && info.MethodOptions.Authz.CloudAccount {
				log.Fatalf("%v: %v/%v: authz.user and authz.cloudAccount can't both be true", info.FileName, serviceName, methName)
			}
			methData := tMethod{
				Deploy:     info.Deploy,
				Service:    serviceName,
				Method:     methName,
				AuthzCheck: info.MethodOptions.Authz.AuthzCheck,
			}
			setOpt := func(msgType string, optName string, flag *string) {
				fds := util.FindOptField(msgType, optName)
				if fds == nil {
					log.Fatalf("%v: %v/%v: %v does not have a field with %v option", info.FileName, serviceName, methName, msgType, optName)
				}
				*flag = util.JoinFieldNames(fds, util.FieldTextName)
			}
			if info.MethodOptions.Authz.CloudAccount {
				setOpt(info.Method.GetInputType(), "cloudAccount", &methData.CloudAccountField)
			}
			if info.MethodOptions.Authz.Product {
				setOpt(info.Method.GetInputType(), "product", &methData.ProductNameField)
			}
			if info.MethodOptions.Authz.ProductAccess {
				setOpt(info.Method.GetInputType(), "productAccess", &methData.ProductAccessField)
			}
			if info.MethodOptions.Authz.OwnerCheck {
				setOpt(info.Method.GetInputType(), "ownerCheck", &methData.OwnerCheckField)
			}
			if info.MethodOptions.Authz.GtsCheck {
				setOpt(info.Method.GetInputType(), "gtsCheck", &methData.GTSCheckNameField)
			}
			if info.MethodOptions.Authz.UserName {
				setOpt(info.Method.GetInputType(), "userName", &methData.UserNameField)
			}
			if info.MethodOptions.Authz.AuthzCheck {
				setOpt(info.Method.GetInputType(), "authzCheck", &methData.AuthzCheckField)
			}

			tmplData.Methods = append(tmplData.Methods, methData)
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("authz").Parse(allowTempl)
	if err != nil {
		log.Fatal(err)
	}

	if configmap != "" {
		outputConfigMap(tmpl, &tmplData)
	}
	if regoDir != "" {
		outputRegoDir(tmpl, &tmplData)
	}
}

func outputRegoDir(tmpl *template.Template, tmplData *tData) {
	deploys := map[string]*tData{"all": {}}

	err := os.MkdirAll(regoDir, 0770)
	if err != nil {
		log.Fatal(err)
	}

	for _, methData := range tmplData.Methods {
		data, ok := deploys[methData.Deploy]
		if !ok {
			data = &tData{}
			deploys[methData.Deploy] = data
		}
		data.Methods = append(data.Methods, methData)
		all := deploys["all"]
		all.Methods = append(all.Methods, methData)
	}

	for deploy, data := range deploys {
		fileName := fmt.Sprintf("%s/%s.rego", regoDir, deploy)
		if err := util.WriteFileAtomically(fileName,
			func(outf io.Writer) error {
				if err := util.WriteGenCommentGo(outf, tool); err != nil {
					return err
				}
				return tmpl.Execute(outf, data)
			}); err != nil {
			log.Fatal(err)
		}
	}
}

func outputConfigMap(tmpl *template.Template, tmplData *tData) {
	tmplData.ConfigMap = true
	defer func() { tmplData.ConfigMap = false }()
	if err := util.WriteFileAtomically(configmap,
		func(outf io.Writer) error {
			buf := bytes.Buffer{}
			if err := tmpl.Execute(&buf, tmplData); err != nil {
				return err
			}
			return util.WriteConfigMap(outf, tool, "authzappclient", buf.Bytes())
		}); err != nil {
		log.Fatal(err)
	}
}
