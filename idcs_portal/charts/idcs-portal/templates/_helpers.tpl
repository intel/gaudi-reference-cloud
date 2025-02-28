{{/*
Expand the name of the chart.
*/}}
{{- define "idcs-portal.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "idcs-portal.fullname" -}}
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
The namespace to install a release to
*/}}
{{- define "idcs-portal.namespace" -}}
{{- .Values.namespaceOverride | default .Release.Namespace -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "idcs-portal.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "idcs-portal.labels" -}}
helm.sh/chart: {{ include "idcs-portal.chart" . }}
{{ include "idcs-portal.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "idcs-portal.selectorLabels" -}}
app.kubernetes.io/name: {{ include "idcs-portal.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "idcs-portal.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "idcs-portal.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
This resolves the image with the global image registry with parent and child charts if applicable
Usage: {{ include "common.image" (dict "root" $ "image" .Values.path.to.the.image) }}
*/}}
{{- define "common.image" -}}
{{- $ := .root -}}
{{- $globalRegistry := $.Values.global.imageRegistry -}}
{{- $registry := .image.registry -}}
{{- $repository := .image.repository -}}
{{- $tag := .image.tag | default $.Chart.AppVersion -}}
{{- if $globalRegistry }}
  {{- printf "%s/%s:%s" $globalRegistry $repository $tag -}}
{{- else if not .image.registry }}
  {{- printf "%s:%s" $repository $tag -}}
{{- else }}
  {{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}
{{- end -}}

{{/*
Web portal image
*/}}
{{- define "idcs-portal.image" -}}
{{ include "common.image" (dict "root" $ "image" .Values.image) }}
{{- end -}}
