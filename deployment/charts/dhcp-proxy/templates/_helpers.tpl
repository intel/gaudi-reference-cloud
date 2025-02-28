{{/*
Expand the name of the chart.
*/}}
{{- define "dhcp-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "dhcp-proxy.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "dhcp-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "dhcp-proxy.labels" -}}
helm.sh/chart: {{ include "dhcp-proxy.chart" . }}
{{ include "dhcp-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "dhcp-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dhcp-proxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "dhcp-proxy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "dhcp-proxy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
container image
*/}}
{{- define "dhcp-proxy.image" -}}
{{- $registry := (.Values.image).registry | default "internal-placeholder.com" -}}
{{- $repository := (.Values.image).repository -}}
{{- $tag := (.Values.image).tag | default $.Chart.AppVersion -}}
{{- if not (.Values.image).registry }}
  {{- printf "%s:%s" $repository $tag -}}
{{- else }}
  {{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}
{{- end -}}