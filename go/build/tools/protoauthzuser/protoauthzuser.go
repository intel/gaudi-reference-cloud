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
     
email := authz.token.email
enterpriseId := authz.token.enterpriseId

response := cloudaccount.getRelatedCloudAccounts(email)
relatedCloudAccounts := response["relatedAccounts"]

countryCode := authz.token.countryCode

result["allowed"] := allow

result["body"] := "User is restricted" {
	method_ok
    not user_ok
	not allow
}

{{ print "{{-"}} if not $.Values.adminJwtOpaRules.insecureAlwaysAllowAdminToken {{ print "}}"}}
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

result["body"] := "permission to access the resource is denied" {
	not authz_ok
	needs_authz_match
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
default needs_authz_match = false

# users allowed with or without cloudaccount
allow {
	email != ""
	method_ok
	user_ok
{{ print "{{-"}} if not $.Values.adminJwtOpaRules.insecureAlwaysAllowAdminToken {{ print "}}"}}
    product_ok
	product_access_ok
	authz_ok
	gts_ok
{{ print "{{- end }}"}}
}

allow {
	some _, role in authz.token.groups
	role == "IDC.Admin"
	endswith(email, "@intel.com")
{{ print "{{-"}} if not $.Values.adminJwtOpaRules.insecureAlwaysAllowAdminToken {{ print "}}"}}
	valid_admin_endpoint
{{ print "{{- end }}"}}
}

allow {
	some _, role in authz.token.roles
	role == "IDC.Admin"
	endswith(authz.token.email, "@intel.com")
{{ print "{{-"}} if not $.Values.adminJwtOpaRules.insecureAlwaysAllowAdminToken {{ print "}}"}}
	valid_admin_endpoint
{{ print "{{- end }}"}}
}

admin_allowed_actions = {
	# AuthzService admin endpoint
	"/proto.AuthzService/AddCloudAccountRolesToUser": ["GET", "POST"],
	"/proto.AuthzService/RemoveCloudAccountRolesFromUser": ["GET", "POST"],
	"/proto.AuthzService/AssignSystemRole": ["GET", "POST"],
	"/proto.AuthzService/UnassignSystemRole": ["GET", "POST"],
	"/proto.AuthzService/ListUsersByCloudAccount": ["GET", "POST"],
	"/proto.AuthzService/CreatePolicy": ["GET", "POST"],
	"/proto.AuthzService/RemovePolicy": ["GET", "POST"],
	"/proto.AuthzService/Check": ["GET", "POST"],
	"/proto.AuthzService/Lookup": ["GET", "POST"],
	# Billing admin endpoints
	"/proto.BillingAccountService/Create": ["GET", "POST"],
	"/proto.BillingUsageService/Read": ["GET", "POST"],
	"/proto.BillingCouponService/Read": ["GET", "POST"],
	"/proto.BillingCouponService/Create": ["GET", "POST"],
	"/proto.BillingCouponService/Disable": ["GET", "POST"],
	"/proto.BillingOptionService/Read": ["GET", "POST"],
	"/proto.BillingCreditService/Read": ["GET", "POST"],
	"/proto.BillingCreditService/Create": ["GET", "POST"],
	"/proto.BillingCreditService/ReadUnappliedCreditBalance": ["GET", "POST"],
	"/proto.BillingDeactivateInstancesService/GetDeactivateInstances": ["GET", "POST"],
	#Cloudaccount admin endpoints
	"/proto.CloudAccountService/GetByName": ["GET", "POST"],
	"/proto.CloudAccountService/Search": ["GET", "POST"],
	"/proto.CloudAccountService/Delete": ["GET", "POST"],
	"/proto.CloudAccountService/GetById": ["GET", "POST"],
	"/proto.CloudAccountService/Update": ["GET", "POST"],
	"/proto.CloudAccountService/Create": ["GET", "POST"],
	"/proto.CloudAccountService/Ensure": ["GET", "POST"],
	# Cloudcredits admin endpoints
	"/proto.CloudCreditsCouponService/Read": ["GET", "POST"], 
	"/proto.CloudCreditsCouponService/Create": ["GET", "POST"],
	"/proto.CloudCreditsCouponService/ReadCredits": ["GET", "POST"],
	"/proto.CloudCreditsCouponService/Redeem": ["GET", "POST"],
	"/proto.CloudCreditsCouponService/Disable": ["GET", "POST"],
	"/proto.CloudCreditsCreditService/CreditMigrate": ["GET", "POST"],
	"/proto.CloudCreditsCreditService/Create": ["GET", "POST"],
	"/proto.CloudCreditsCreditService/ReadCredits": ["GET", "POST"],
	# Fleetadmin admin endpoints
	"/proto.FleetAdminService/Ping": ["GET", "POST"],
	"/proto.FleetAdminUIService/SearchNodes": ["GET", "POST"],
	"/proto.FleetAdminUIService/SearchComputeNodePoolsForPoolAccountManager": ["GET", "POST"],
	"/proto.FleetAdminUIService/SearchComputeNodePoolsForNodeAdmin": ["GET", "POST"],
	"/proto.FleetAdminUIService/PutComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminUIService/SearchCloudAccountsForComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminUIService/DeleteCloudAccountFromComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminUIService/UpdateNode": ["GET", "POST"],
	"/proto.FleetAdminService/SearchComputeNodePoolsForInstanceScheduling": ["GET", "POST"],
	"/proto.FleetAdminService/UpdateComputeNodePoolsForCloudAccount": ["GET", "POST"],	
	"/proto.FleetAdminUIService/AddCloudAccountToComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminUIService/SearchInstanceTypeStatsForNode": ["GET", "POST"],
	# Deprecated Fleetadmin endpoints
	"/proto.FleetAdminService/SearchNodes": ["GET", "POST"],
	"/proto.FleetAdminService/SearchComputeNodePoolsForPoolAccountManager": ["GET", "POST"],
	"/proto.FleetAdminService/SearchComputeNodePoolsForNodeAdmin": ["GET", "POST"],
	"/proto.FleetAdminService/PutComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminService/SearchCloudAccountsForComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminService/DeleteCloudAccountFromComputeNodePool": ["GET", "POST"],
	"/proto.FleetAdminService/UpdateNode": ["GET", "POST"],
	"/proto.FleetAdminService/AddCloudAccountToComputeNodePool": ["GET", "POST"],
	# IKS admin endpoints
	"/proto.IksAdmin/GetFirewallRule": ["GET", "POST"],
	"/proto.IksAdmin/AuthenticateIKSAdminUser": ["GET", "POST"],
	"/proto.IksAdmin/ClusterRecreate": ["GET", "POST"],
	"/proto.IksAdmin/ClusterSnapshot": ["GET", "POST"],
	"/proto.IksAdmin/CreateIMI": ["GET", "POST"],
	"/proto.IksAdmin/CreateInstanceTypes": ["GET", "POST"],
	"/proto.IksAdmin/CreateK8SVersion": ["GET", "POST"],
	"/proto.IksAdmin/CreateNewAddOn": ["GET", "POST"],
	"/proto.IksAdmin/DeleteAddOn": ["GET", "POST"],
	"/proto.IksAdmin/DeleteIMI": ["GET", "POST"],
	"/proto.IksAdmin/DeleteInstanceType": ["GET", "POST"],
	"/proto.IksAdmin/DeleteK8SVersion": ["GET", "POST"],
	"/proto.IksAdmin/DeleteLoadBalancer": ["GET", "POST"],
	"/proto.IksAdmin/GetAddOn": ["GET", "POST"],
	"/proto.IksAdmin/GetAddOns": ["GET", "POST"],
	"/proto.IksAdmin/GetCloudAccountApproveList": ["GET", "POST"],
	"/proto.IksAdmin/GetCluster": ["GET", "POST"],
	"/proto.IksAdmin/GetClusters": ["GET", "POST"],
	"/proto.IksAdmin/GetControlPlaneSSHKeys": ["GET", "POST"],
	"/proto.IksAdmin/GetEvents": ["GET", "POST"],
	"/proto.IksAdmin/GetIMI": ["GET", "POST"],
	"/proto.IksAdmin/GetIMIs": ["GET", "POST"],
	"/proto.IksAdmin/GetIMIsInfo": ["GET", "POST"],
	"/proto.IksAdmin/GetInstanceType": ["GET", "POST"],
	"/proto.IksAdmin/GetInstanceTypeInfo": ["GET", "POST"],
	"/proto.IksAdmin/GetInstanceTypes": ["GET", "POST"],
	"/proto.IksAdmin/GetK8SVersion": ["GET", "POST"],
	"/proto.IksAdmin/GetLoadBalancer": ["GET", "POST"],
	"/proto.IksAdmin/GetLoadBalancers": ["GET", "POST"],
	"/proto.IksAdmin/PostCloudAccountApproveList": ["GET", "POST"],
	"/proto.IksAdmin/PostLoadBalancer": ["GET", "POST"],
	"/proto.IksAdmin/PutAddOn": ["GET", "POST"],
	"/proto.IksAdmin/PutCPNodegroup": ["GET", "POST"],
	"/proto.IksAdmin/PutCloudAccountApproveList": ["GET", "POST"],
	"/proto.IksAdmin/PutIMI": ["GET", "POST"],
	"/proto.IksAdmin/PutInstanceType": ["GET", "POST"],
	"/proto.IksAdmin/PutK8SVersion": ["GET", "POST"],
	"/proto.IksAdmin/PutLoadBalancer": ["GET", "POST"],
	"/proto.IksAdmin/UpdateIMIInstanceTypeToK8sCompatibility": ["GET", "POST"],
	"/proto.IksAdmin/UpdateInstanceTypeIMIToK8sCompatibility": ["GET", "POST"],
	"/proto.IksAdmin/UpgradeClusterControlPlane": ["GET", "POST"],
	# Instance admin endpoints
	"/proto.InstanceService/Search": ["GET", "POST"],
	"/proto.InstanceService/Search2": ["GET", "POST"],
	"/proto.InstanceService/Delete": ["GET", "POST"],
	"/proto.InstanceService/Delete2": ["GET", "POST"],
	"/proto.InstanceGroupService/Search": ["GET", "POST"],
	"/proto.InstanceGroupService/Delete": ["GET", "POST"],
	# Metering admin endpoints
	"/proto.MeteringService/Update": ["GET", "POST"],
	"/proto.MeteringService/Search": ["GET", "POST"],
	"/proto.MeteringService/SearchInvalid": ["GET", "POST"],
	"/proto.MeteringService/Create": ["GET", "POST"],
	"/proto.MeteringService/FindPrevious": ["GET", "POST"],
	"/proto.MeteringService/CreateInvalidRecords": ["GET", "POST"],
	"/proto.MeteringService/IsMeteringRecordAvailable": ["GET", "POST"],
	"/proto.SecurityInsights/CompareReleaseVulnerabilities": ["GET", "POST"],
	# Security Storage
	"/proto.SecurityInsights/GetRelease": ["GET", "POST"],
	"/proto.SecurityInsights/GetReleaseSBOM": ["GET", "POST"],
	"/proto.SecurityInsights/GetSummary": ["GET", "POST"],
	"/proto.SecurityInsights/GetReleaseComponent": ["GET", "POST"],
	"/proto.SecurityInsights/GetReleaseVulnerabilities": ["GET", "POST"],
	"/proto.SecurityInsights/GetAllReleases": ["GET", "POST"],
	"/proto.SecurityInsights/GetAllComponents": ["GET", "POST"],
	# Storage admin endpoints
	"/proto.StorageAdminService/GetResourceUsage": ["GET", "POST"],
	"/proto.StorageAdminService/InsertStorageQuotaByAccount": ["GET", "POST"],
	"/proto.StorageAdminService/UpdateStorageQuotaByAccount": ["GET", "POST"],
	"/proto.StorageAdminService/DeleteStorageQuotaByAccount": ["GET", "POST"],
	"/proto.StorageAdminService/GetStorageQuotaByAccount": ["GET", "POST"],
	"/proto.StorageAdminService/GetAllStorageQuota": ["GET", "POST"],
	"/proto.FileStorageService/FileStorageService": ["GET", "POST"],
	"/proto.ObjectStorageService/DeleteBucket": ["GET", "POST"],	
	"/proto.ObjectStorageService/DeleteBucketLifecycleRule": ["GET", "POST"],
	"/proto.ObjectStorageService/DeleteObjectUser": ["GET", "POST"],
	"/proto.S3Service/DeleteBucket": ["GET", "POST"],
	"/proto.S3Service/DeleteLifecycleRules": ["GET", "POST"],
	"/proto.S3Service/DeleteS3Principal": ["GET", "POST"],
	"/proto.QuotaManagementService/Register": ["GET", "POST"],
	"/proto.QuotaManagementService/GetServiceResource": ["GET", "POST"],
	"/proto.QuotaManagementService/UpdateServiceRegistration": ["GET", "POST"],
	"/proto.QuotaManagementService/CreateServiceQuota": ["GET", "POST"],
	"/proto.QuotaManagementService/GetServiceQuotaResource": ["GET", "POST"],
	"/proto.QuotaManagementService/UpdateServiceQuotaResource": ["GET", "POST"],
	"/proto.QuotaManagementService/DeleteServiceQuotaResource": ["GET", "POST"],
	"/proto.QuotaManagementService/DeleteService": ["GET", "POST"],
	"/proto.QuotaManagementService/ListServiceQuota": ["GET", "POST"],
	"/proto.QuotaManagementService/ListAllServiceQuotas": ["GET", "POST"],
	"/proto.QuotaManagementService/ListRegisteredServices": ["GET", "POST"],
	"/proto.QuotaManagementService/ListServiceResources": ["GET", "POST"],
	# Product Catalog admin endpoints
	"/proto.ProductAccessService/ReadAccess": ["GET", "POST"],
	"/proto.ProductAccessService/CheckProductAccess": ["GET", "POST"],
	"/proto.ProductAccessService/AddAccess": ["GET", "POST"],
	"/proto.ProductAccessService/RemoveAccess": ["GET", "POST"],
	# PC
	"/proto.ProductCatalogService/AdminRead": ["GET", "POST"],
	"/proto.ProductCatalogService/UserRead": ["GET", "POST"],
	"/proto.ProductCatalogService/UserReadExternal": ["GET", "POST"],
	"/proto.RegionService/Add": ["GET", "POST"],
	"/proto.RegionService/AdminRead": ["GET", "POST"],
	"/proto.RegionService/UserRead": ["GET", "POST"],
	"/proto.RegionService/Update": ["GET", "POST"],
	"/proto.RegionService/Delete": ["GET", "POST"],
	"/proto.ProductFamilyService/Add": ["GET", "POST"],
	"/proto.ProductFamilyService/Read": ["GET", "POST"],
	"/proto.ProductFamilyService/Update": ["GET", "POST"],
	"/proto.ProductFamilyService/Delete": ["GET", "POST"],
	"/proto.ProductVendorService/Add": ["GET", "POST"],
	"/proto.ProductVendorService/Read": ["GET", "POST"],
	"/proto.ProductVendorService/Update": ["GET", "POST"],
	"/proto.ProductVendorService/Delete": ["GET", "POST"],
    # Usage Admin endpoints
	"/proto.UsageService/SearchProductUsages": ["GET", "POST"],
	"/proto.UsageService/SearchResourceUsages": ["GET", "POST"],
	"/proto.UsageService/StreamSearchProductUsages": ["GET", "POST"],
	"/proto.UsageService/StreamSearchResourceUsages": ["GET", "POST"],
	"/proto.UsageRecordService/CreateProductUsageRecord": ["GET", "POST"],
	"/proto.UsageRecordService/SearchProductUsageRecords": ["GET", "POST"],
	"/proto.UsageRecordService/SearchInvalidProductUsageRecords": ["GET", "POST"],
	# Cloud Account Region Access admin endpoints
	"/proto.RegionAccessService/ReadAccess": ["GET", "POST"],
	"/proto.RegionAccessService/AddAccess": ["GET", "POST"],
	"/proto.RegionAccessService/RemoveAccess": ["GET", "POST"],
	"/proto.RegionAccessService/CheckRegionAccess": ["GET", "POST"],

}

valid_admin_endpoint {
	admin_allowed_actions[input.attributes.request.http.path][_] == input.attributes.request.http.method
}

is_reflect {
	input.parsed_path[0] == "grpc.reflection.v1alpha.ServerReflection"
}

method_ok {
	is_reflect
}

product_ok {
	is_reflect
	endswith(authz.token.email, "@intel.com")
}

gts_ok {
	is_reflect
	endswith(authz.token.email, "@intel.com")
}

gts_ok {
	endswith(authz.token.email, "@intel.com")
}

{{- range .Methods }}
{{ if $.ConfigMap }}
{{ print "{{-"}} if or (eq $.Values.deployment "all") (eq $.Values.deployment "{{.Deploy}}") {{ print "}}"}}
{{- end}}
{{- if .AuthzCheck }}

{{ .Service }}{{ .Method }}_authz_ok {
{{- if .AuthzCheckField }}
	{{ .Service }}_{{ .Method }}_AuthzCloudAccountId := input.parsed_body.{{ .AuthzCheckField }}
{{- end }}
{{- if not .AuthzCheckField }}
	{{ .Service }}_{{ .Method }}_AuthzCloudAccountId := "*"
{{- end }}
	authzCheck := authzService.check(email, enterpriseId,{{ .Service }}_{{ .Method }}_AuthzCloudAccountId, input.attributes.request.http.headers["x-original-http-path"],input.attributes.request.http.headers["x-original-http-method"],input.parsed_body)
	authzCheck == true
}
{{- end }}

authz_ok {
    input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .AuthzCheck }}
	{{ .Service }}{{ .Method }}_authz_ok
   {{- end }}
}

needs_authz_match {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
}

method_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .CloudAccountField }}
    some _, relatedCloudAccount in relatedCloudAccounts
	input.parsed_body.{{ .CloudAccountField }} == relatedCloudAccount["id"]
{{- end }}
{{- if .UserNameField }}
	input.parsed_body.{{ .UserNameField }} == authz.token.email
{{- end }}
}

{{- if .CloudAccountField }}
{{ .Service }}_{{ .Method }}_cloudAccount := cloudaccount.getById(input.parsed_body.{{ .CloudAccountField }})

personId := {{ .Service }}_{{ .Method }}_cloudAccount["personId"] if {
	email == {{ .Service }}_{{ .Method }}_cloudAccount["name"]
} else := cloudaccount.getMemberPersonId(email, {{ .Service }}_{{ .Method }}_cloudAccount["id"])
{{- end }}

user_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .CloudAccountField }}
	not {{ .Service }}_{{ .Method }}_cloudAccount["restricted"]
{{- end }}
{{- if .OwnerCheckField }}
	email == {{ .Service }}_{{ .Method }}_cloudAccount["name"]
{{- end }}
}

product_ok {
	input.parsed_path == ["proto.{{ .Service }}", "{{ .Method }}"]
{{- if .ProductNameField }}
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
			return util.WriteConfigMap(outf, tool, "authzuser", buf.Bytes())
		}); err != nil {
		log.Fatal(err)
	}
}
