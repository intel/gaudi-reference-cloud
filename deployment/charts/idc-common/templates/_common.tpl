{{/*
Expand the name of the chart.
*/}}
{{- define "idc-common.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "idc-common.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}


{{/*
TLS PKI certificate common name
*/}}
{{- define "idc-common.tlsCommonName" -}}
{{- with (.Values.tls).commonName }}
{{- . }}
{{- else }}
{{- include "idc-common.fullname" . -}}.idcs-system.svc.cluster.local
{{- end }}
{{- end }}


{{/*
The namespace to install a release to
*/}}
{{- define "idc-common.namespace" -}}
{{- .Values.namespaceOverride | default .Release.Namespace -}}
{{- end -}}


{{/*
container image
*/}}
{{- define "idc-common.image" -}}
{{- $registry := (.Values.image).registry | default "internal-placeholder.com" -}}
{{- $repository := (.Values.image).repository -}}
{{- $tag := (.Values.image).tag | default $.Chart.AppVersion -}}
{{- if not (.Values.image).registry }}
  {{- printf "%s:%s" $repository $tag -}}
{{- else }}
  {{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}
{{- end -}}


{{/*
zap logging parameters
  --zap-encoder: Zap log encoding (one of 'json' or 'console')

  --zap-log-level: Zap Level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', or any integer value > 0 which corresponds to custom debug levels of increasing verbosity.

  --zap-stacktrace-level: Zap Level at and above which stacktraces are captured (one of 'info', 'error', 'panic')
*/}}
{{- define "idc-common.logArgs" -}}
- "--zap-encoder={{ (.Values.log).encoder | default "json" }}"
- "--zap-log-level={{ (.Values.log).level | default "debug" }}"
- "--zap-stacktrace-level={{ (.Values.log).stacktraceLevel | default "error" }}"
{{- end -}}


{{- define "idc-common.proxyEnv" -}}
- name: http_proxy
  value: {{ (.Values.proxy).http_proxy | quote }}
- name: HTTP_PROXY
  value: {{ (.Values.proxy).http_proxy | quote }}
- name: https_proxy
  value: {{ (.Values.proxy).https_proxy | quote }}
- name: HTTPS_PROXY
  value: {{ (.Values.proxy).https_proxy | quote }}
- name: no_proxy
  value: {{ (.Values.proxy).no_proxy | quote }}
- name: NO_PROXY
  value: {{ (.Values.proxy).no_proxy | quote }}
{{- end -}}


{{- define "idc-common.commonEnv" -}}
- name: IDC_SERVER_TLS_ENABLED
  value: {{ ((.Values.tls).server).enabled | quote }}
- name: IDC_CLIENT_TLS_ENABLED
  value: {{ ((.Values.tls).client).enabled | quote }}
- name: IDC_GRPC_TLS_AUTHZ_ENABLED
  value: {{ ((.Values.tls).grpcTlsAuthz).enabled | quote }}
- name: IDC_GRPC_TLS_AUTHZ_ALLOWED_CALLERS
  value: {{ ((.Values.tls).grpcTlsAuthz).allowedCallers | quote }}
- name: IDC_REQUIRE_CLIENT_CERTIFICATE
  value: {{ ((.Values.tls).server).requireClientCertificate | quote }}
- name: IDC_INSECURE_SKIP_VERIFY
  value: {{ ((.Values.tls).client).insecureSkipVerify | quote }}
- name: GRPC_GO_LOG_SEVERITY_LEVEL
  value: {{ (.Values.log).grpcGoLogSeverityLevel | quote }}
- name: GRPC_GO_LOG_FORMATTER
  value: {{ (.Values.log).grpcGoLogEncoding | quote }}
- name: GRPC_GO_LOG_VERBOSITY_LEVEL
  value: {{ (.Values.log).grpcGoLogVerbosityLevel | quote }}
{{- end -}}

{{- define "idc-common.cognitoEnv" -}}
- name: IDC_COGNITO_ENABLED
  value: {{ ((.Values).cognito).enabled | default false | quote }}
- name: IDC_COGNITO_ENDPOINT
  value: {{ ((.Values).cognito).endpoint | quote }}
{{- end -}}

{{- define "idc-common.gtsEnv" -}}
- name: gts_get_token_url
  value: {{ (.Values.gts).get_token_url | quote }}
- name: gts_create_product_url
  value: {{ (.Values.gts).create_product_url | quote }}
- name: gts_create_order_url
  value: {{ (.Values.gts).create_order_url | quote }}
- name: gts_business_screen_url
  value: {{ (.Values.gts).business_screen_url | quote }}
{{- end -}}

{{/*
Define OTEL environment variables that will be used for all IDC service pods.
This will only used for tracing.
For logs in development environments, see helmfile-telemetry.yaml.
*/}}
{{- define "idc-common.otelEnv" -}}

  {{- $resourceAttributes := (dict
        "deployment.environment" ((.Values.otel).deployment).environment
        "service.namespace" (include "idc-common.namespace" .)
        "service.group" ((.Values.otel).deployment).group) }}
  {{- $extraResourceAttributes := (.Values.otel).extraResourceAttributes | default (dict) }}
  {{- $resourceAttributes = mustMergeOverwrite $resourceAttributes (deepCopy $extraResourceAttributes) }}

  {{- $otelResourceAttributesValue := "" }}
  {{- range $key, $value := $resourceAttributes }}
    {{- if $otelResourceAttributesValue }}
      {{- $otelResourceAttributesValue = print $otelResourceAttributesValue "," }}
    {{- end }}
    {{- $otelResourceAttributesValue = print $otelResourceAttributesValue $key "=" $value }}
  {{- end }}

- name: OTEL_RESOURCE_ATTRIBUTES
  value: {{ $otelResourceAttributesValue | quote }}
- name: OTEL_SERVICE_NAME
  value: "{{ include "idc-common.fullname" . }}"
- name: OTEL_EXPORTER_OTLP_CERTIFICATE
  value: {{ (((.Values.otel).exporter).otlp).certificate | default "/vault/secrets/otel_ca_pem" | quote }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ (((.Values.otel).exporter).otlp).endpoint | default "internal-placeholder.com:443" | quote }}
{{- end -}}

{{- define "idc-common.otelAnnotations" -}}
vault.hashicorp.com/agent-inject-secret-otel_ca_pem: public/data/otel
vault.hashicorp.com/agent-inject-template-otel_ca_pem: |-
  {{`{{- with secret `}}"{{ "public/data/otel" }}"{{` -}}`}}
  {{`{{ .Data.data.otel_ca_pem }}`}}
  {{`{{- end }}`}}
{{- end -}}

{{- define "idc-common.globalDBSSLCerts" -}}
vault.hashicorp.com/agent-inject-secret-db_sslcert: public/data/globaldbssl
vault.hashicorp.com/agent-inject-template-db_sslpem: |-
  {{`{{- with secret `}}"{{ "public/data/globaldbssl" }}"{{` -}}`}}
  {{`{{ .Data.data.sslcert }}`}}
  {{`{{- end }}`}}
{{- end -}}

{{- define "idc-common.vaultAnnotations" -}}
vault.hashicorp.com/agent-init-first: "true"
vault.hashicorp.com/agent-inject: "true"
vault.hashicorp.com/agent-inject-status: "update"
vault.hashicorp.com/role: "{{ include "idc-common.serviceAccountName" . }}-role"
vault.hashicorp.com/service: {{ (.Values.vault).service | quote }}
vault.hashicorp.com/proxy-address: {{ (.Values.vault).proxy | default "" | quote }}
vault.hashicorp.com/auth-path: {{ (.Values.vault).authPath | quote }}
{{- end -}}


{{- define "idc-common.ironicIP" -}}
{{ default "10.10.10.1" .Values.ironicEnvConfigMap.ironicIP }}
{{- end }}


{{- define "idc-common.vaultPkiAnnotations" -}}
# Inject certificate and private-key issued by PKI-Engine
vault.hashicorp.com/agent-inject-secret-certkey.pem: ""
# Service certificate expiration will be the lower of PKI_ROLE_TTL in deployment/common/vault/configure.sh and the ttl value below.
vault.hashicorp.com/agent-inject-template-certkey.pem: |
  {{`{{- with pkiCert`}} "{{ (.Values.tls).issueCa }}/issue/{{ include "idc-common.fullname" . }}" "common_name={{ include "idc-common.tlsCommonName" . }}" "ttl=24h" {{` -}}`}}
  {{`{{ .Data.Cert }}`}}
  {{`{{ .Data.CA }}`}}
  {{`{{ .Data.Key }}`}}
  {{`{{ .Data.Cert | writeToFile "/vault/secrets/cert.pem" "vault" "vault" "0644" }}`}}
  {{`{{ .Data.CA | writeToFile "/vault/secrets/cert.pem" "vault" "vault" "0644" "append" }}`}}
  {{`{{ .Data.Key | writeToFile "/vault/secrets/cert.key" "vault" "vault" "0644" }}`}}
  {{`{{- end }}`}}
# Inject ca_chain certificates
vault.hashicorp.com/agent-inject-secret-ca.pem: "{{ ((.Values.tls).client).rootCa }}/cert/ca_chain"
vault.hashicorp.com/agent-inject-template-ca.pem: |
  {{`{{- with secret`}} "{{ ((.Values.tls).client).rootCa }}/cert/ca_chain" {{` -}}`}}
  {{`{{ .Data.certificate }}`}}
  {{`{{- end }}`}}
{{- end -}}

# Inject the AWS Cognito Credentials to access tokens
{{- define "idc-common.vaultCognitoAnnotations" -}}
vault.hashicorp.com/agent-inject-secret-client_id: "{{ .Values.cognito.vaultCredentialsPath }}"
vault.hashicorp.com/agent-inject-template-client_id: |-
  {{`{{- with secret `}}"{{ .Values.cognito.vaultCredentialsPath }}"{{` -}}`}}
  {{`{{ .Data.data.client_id }}`}}
  {{`{{- end }}`}}
vault.hashicorp.com/agent-inject-secret-client_secret: "{{ .Values.cognito.vaultCredentialsPath }}"
vault.hashicorp.com/agent-inject-template-client_secret: |-
  {{`{{- with secret `}}"{{ .Values.cognito.vaultCredentialsPath }}"{{` -}}`}}
  {{`{{ .Data.data.client_secret }}`}}
  {{`{{- end }}`}}
{{- end -}}

{{- define "idc-common.vaultBaremetalIronicPkiAnnotations" -}}
vault.hashicorp.com/role: "{{ include "idc-common.serviceAccountName" . }}-ironic-role"
# Inject ironic certificate and private-key issued by PKI-Engine
vault.hashicorp.com/agent-inject-secret-ironic: ""
vault.hashicorp.com/secret-volume-path-ironic: "/certs/ironic"
vault.hashicorp.com/agent-inject-file-ironic: 'ironic'
vault.hashicorp.com/agent-inject-template-ironic: |
  {{`{{- with pkiCert`}} "{{ (.Values.tls).issueCa }}/issue/{{ include "idc-common.fullname" . }}" "common_name={{ include "idc-common.ironicIP" . }}:6385" "ttl=300m" "ip_sans=127.0.0.1,{{ include "idc-common.ironicIP" . }}" -}}`}}
  {{`{{ .Data.Cert | writeToFile "/certs/ironic/tls.crt" "vault" "vault" "0600" }}`}}
  {{`{{ .Data.CA | writeToFile "/certs/ironic/tls.crt" "vault" "vault" "0600" "append" }}`}}
  {{`{{ .Data.Key | writeToFile "/certs/ironic/tls.key" "vault" "vault" "0600" }}`}}
  {{`{{- end }}`}}
# Inject inspector certificate and private-key issued by PKI-Engine
vault.hashicorp.com/agent-inject-secret-ironic-inspector: ""
vault.hashicorp.com/secret-volume-path-ironic-inspector: "/certs/ironic-inspector"
vault.hashicorp.com/agent-inject-file-ironic-inspector: 'ironic-inspector'
vault.hashicorp.com/agent-inject-template-ironic-inspector: |
  {{`{{- with pkiCert`}} "{{ (.Values.tls).issueCa }}/issue/{{ include "idc-common.fullname" . }}" "common_name={{ include "idc-common.ironicIP" . }}:5050" "ttl=300m" "ip_sans=127.0.0.1,{{ include "idc-common.ironicIP" . }}" -}}`}}
  {{`{{ .Data.Cert | writeToFile "/certs/ironic-inspector/tls.crt" "vault" "vault" "0600" }}`}}
  {{`{{ .Data.CA | writeToFile "/certs/ironic-inspector/tls.crt" "vault" "vault" "0600" "append" }}`}}
  {{`{{ .Data.Key | writeToFile "/certs/ironic-inspector/tls.key" "vault" "vault" "0600" }}`}}
  {{`{{- end }}`}}
# Inject mariadb certificate and private-key issued by PKI-Engine
vault.hashicorp.com/agent-inject-secret-mariadb: ""
vault.hashicorp.com/secret-volume-path-mariadb: "/certs/mariadb"
vault.hashicorp.com/agent-inject-file-mariadb: 'mariadb'
vault.hashicorp.com/agent-inject-template-mariadb: |
  {{`{{- with pkiCert`}} "{{ (.Values.tls).issueCa }}/issue/{{ include "idc-common.fullname" . }}" "common_name=127.0.0.1" "ttl=300m" "ip_sans=127.0.0.1" -}}`}}
  {{`{{ .Data.Cert | writeToFile "/certs/mariadb/tls.crt" "vault" "vault" "0600" }}`}}
  {{`{{ .Data.CA | writeToFile "/certs/mariadb/tls.crt" "vault" "vault" "0600" "append" }}`}}
  {{`{{ .Data.Key | writeToFile "/certs/mariadb/tls.key" "vault" "vault" "0600" }}`}}
  {{`{{- end }}`}}
{{- $featureList := list "ironic" "ironic-inspector" "mariadb" }}
{{- range $feature := $featureList }}
# Inject ironic ca_chain certificates
vault.hashicorp.com/agent-inject-secret-{{ $feature }}-ca.crt: "{{ (.Values.tls).issueCa }}/cert/ca_chain"
vault.hashicorp.com/secret-volume-path-{{ $feature }}-ca.crt: "/certs/ca/{{ $feature }}"
vault.hashicorp.com/agent-inject-file-{{ $feature }}-ca.crt: 'tls.crt'
vault.hashicorp.com/agent-inject-template-{{ $feature }}-ca.crt: |
  {{`{{- with secret "{{ (.Values.tls).issueCa }}/cert/ca_chain" -}}`}}
  {{`{{ .Data.certificate }}`}}
  {{`{{- end }}`}}
{{- end }}  
{{- end -}}


{{- define "idc-common.vaultBaremetalOperatorPkiAnnotations" -}}
# Inject ironic certificate and private-key issued by PKI-Engine
vault.hashicorp.com/agent-inject-secret-ironic: ""
vault.hashicorp.com/secret-volume-path-ironic: "/opt/metal3/certs/client"
vault.hashicorp.com/agent-inject-file-ironic: 'ironic'
vault.hashicorp.com/agent-inject-template-ironic: |
  {{`{{- with pkiCert`}} "{{ (.Values.tls).issueCa }}/issue/{{ include "idc-common.fullname" . }}" "common_name={{ include "idc-common.ironicIP" . }}:6385" "ttl=300m" "ip_sans=127.0.0.1,{{ include "idc-common.ironicIP" . }}" -}}`}}
  {{`{{ .Data.Cert | writeToFile "/opt/metal3/certs/client/tls.crt" "vault" "vault" "0644" }}`}}
  {{`{{ .Data.CA | writeToFile "/opt/metal3/certs/client/tls.crt" "vault" "vault" "0644" "append" }}`}}
  {{`{{ .Data.Key | writeToFile "/opt/metal3/certs/client/tls.key" "vault" "vault" "0644" }}`}}
  {{`{{- end }}`}}
# Inject ironic ca_chain certificates
vault.hashicorp.com/agent-inject-secret-ironic-ca.crt: "{{ (.Values.tls).issueCa }}/cert/ca_chain"
vault.hashicorp.com/secret-volume-path-ironic-ca.crt: "/opt/metal3/certs/ca"
vault.hashicorp.com/agent-inject-file-ironic-ca.crt: 'tls.crt'
vault.hashicorp.com/agent-inject-template-ironic-ca.crt: |
  {{`{{- with secret "{{ (.Values.tls).issueCa }}/cert/ca_chain" -}}`}}
  {{`{{ .Data.certificate }}`}}
  {{`{{- end }}`}}
{{- end -}}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "idc-common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Common labels
*/}}
{{- define "idc-common.labels" -}}
helm.sh/chart: {{ include "idc-common.chart" . }}
{{ include "idc-common.selectorLabels" . }}

{{- if .Values.gitCommit }}
app.kubernetes.io/version: {{ .Values.gitCommit | quote }}
{{- else if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

{{- if .Values.configCommit }}
cloud.intel.com/config-commit: {{ .Values.configCommit | quote }}
{{- end }}

app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{/*
Selector labels
*/}}
{{- define "idc-common.selectorLabels" -}}
app.kubernetes.io/name: {{ include "idc-common.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{- define "idc-common.listenPort" -}}
{{ default 8443 .Values.listenPort }}
{{- end }}


{{- define "idc-common.nodePort" -}}
{{ default 0 .Values.nodePort }}
{{- end }}


{{- define "idc-common.targetPort" -}}
{{ default 0 .Values.targetPort }}
{{- end }}


{{/*
IDC grpc ports
*/}}
{{- define "idc-common.grpcPorts" -}}
- name: grpc
  protocol: TCP
  containerPort: {{ include "idc-common.listenPort" . }}
{{- end }}


{{/*
IDC grpc service ports
*/}}
{{- define "idc-common.grpcServicePorts" -}}
- name: grpc
  protocol: TCP
  port: {{ include "idc-common.listenPort" . }}
  nodePort: {{ include "idc-common.nodePort" . }}
  targetPort: {{ include "idc-common.targetPort" . }}
  # Required for Istio. See: https://istio.io/latest/docs/ops/configuration/traffic-management/protocol-selection/
  appProtocol: grpc
{{- end }}

{{/*
Default IDC securityContext
*/}}
{{- define "idc-common.securityContext" -}}
readOnlyRootFilesystem: true
runAsNonRoot: true
runAsUser: 65534
runAsGroup: 65534
allowPrivilegeEscalation: false
capabilities:
  drop:
  - ALL
{{- end }}

{{/*
IDC securityContext with capabilities
*/}}
{{- define "idc-common.securityContextWithCapabilities" -}}
readOnlyRootFilesystem: true
runAsNonRoot: true
runAsUser: 65534
runAsGroup: 65534
capabilities:
  add:
  - NET_ADMIN
  - NET_RAW
{{- end }}
  


{{/*
Create the name of the service account to use
*/}}
{{- define "idc-common.serviceAccountName" -}}
{{- default (include "idc-common.fullname" . ) .Values.serviceAccountName }}
{{- end }}


{{/*
Database URL
If database.service does not contain a dot, it will be suffixed with .namespace.svc.cluster.local.
*/}}
{{- define "idc-common.db-url" -}}

{{- $svc := required "value 'database.service' is required" (.Values.database).service -}}
{{- if not (contains "." $svc) }}
  {{- $namespace := .Values.namespaceOverride | default .Release.Namespace -}}
  {{- $svc := printf "%s.%s.svc.cluster.local" $svc $namespace }}
{{- end }}

{{- $port := default "5432" (.Values.database).port -}}
{{- $name := default "postgres" (.Values.database).name -}}
{{- $arg := default "sslmode=verify-full" (.Values.database).arg -}}
{{ printf "postgres://%s:%s/%s?%s" $svc $port $name $arg }}

{{- end -}}

{{/*
Additional Database URL
If database.service does not contain a dot, it will be suffixed with .namespace.svc.cluster.local.
*/}}
{{- define "idc-common.add-db-url" -}}

{{- if .Values.addDatabase.enabled }}
{{- $svc := required "value 'addDatabase.service' is required" (.Values.addDatabase).service -}}
{{- if not (contains "." $svc) }}
  {{- $namespace := .Values.namespaceOverride | default .Release.Namespace -}}
  {{- $svc := printf "%s.%s.svc.cluster.local" $svc $namespace }}
{{- end }}

{{- $port := default "5432" (.Values.addDatabase).port -}}
{{- $name := default "postgres" (.Values.addDatabase).name -}}
{{- $arg := default "sslmode=verify-full" (.Values.addDatabase).arg -}}
{{ printf "postgres://%s:%s/%s?%s" $svc $port $name $arg }}
{{- end }}

{{- end -}}
