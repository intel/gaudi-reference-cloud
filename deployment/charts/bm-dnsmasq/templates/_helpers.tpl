{{/*
Expand the name of the chart.
*/}}
{{- define "bm-dnsmasq.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bm-dnsmasq.fullname" -}}
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
{{- define "bm-dnsmasq.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bm-dnsmasq.labels" -}}
helm.sh/chart: {{ include "bm-dnsmasq.chart" . }}
{{ include "bm-dnsmasq.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bm-dnsmasq.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bm-dnsmasq.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bm-dnsmasq.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bm-dnsmasq.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
container image
*/}}
{{- define "bm-dnsmasq.image" -}}
{{- $registry := (.Values.image).registry | default "internal-placeholder.com" -}}
{{- $repository := (.Values.image).repository -}}
{{- $tag := (.Values.image).tag | default $.Chart.AppVersion -}}
{{- if not (.Values.image).registry }}
  {{- printf "%s:%s" $repository $tag -}}
{{- else }}
  {{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}
{{- end -}}