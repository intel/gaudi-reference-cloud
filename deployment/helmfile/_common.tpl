{{/*
Ensure that provided name is a valid Helm release name.
The Argo CD Application name will be generated with the pattern "kubeContext-releaseName",
and this must not exceed 63 characters.
If the resulting string would exceed this length, this will fail rendering.
l.
*/}}
{{- define "idc-common.toReleaseName" -}}
  {{- $kubeContext := "xxxxx-xxx-xxxx-" }}
  {{- $name := . }}
  {{- $maxResourceLabelLength := 63 }}
  {{- if (kindIs "map" . ) }}
    {{- $kubeContext = .kubeContext }}
    {{- $name = .name }}
  {{- end }}
  {{- $appName := print $kubeContext "-" $name }}
  {{- if gt (len $appName) $maxResourceLabelLength }}
    {{- fail (print "the application name '" $appName "' (which includes the kubeContext) exceeds " $maxResourceLabelLength " characters") }}
  {{- end }}
  {{- $name }}
{{- end }}


{{/*
Returns the concatenation of .name and .suffix.
But if the resulting string exceeds maxLength characters,
name is truncated and a hashLength-sized hash is appended so that
the total is maxLength characters long.
*/}}
{{- define "idc-common.truncateWithHash" -}}
  {{- $suffix := . | get "suffix" "" }}
  {{- $maxLength := . | get "maxLength" 63 }}
  {{- $hashLength := . | get "hashLength" 7 }}
  {{- $resultWithoutTrunc := print .name $suffix}}
  {{- if gt (len $resultWithoutTrunc) $maxLength }}
    {{- $suffixLength := len $suffix }}
    {{- $hash := .name | sha256sum | trunc $hashLength }}
    {{- $prefixLength := sub $maxLength (add (add1 $hashLength) $suffixLength) | int }}
    {{- if gt $prefixLength 0 }}
      {{- $prefix := .name | trunc $prefixLength }}
      {{- print $prefix "-" $hash $suffix }}
    {{- else }}
      {{- print $hash $suffix }}
    {{- end }}
  {{- else }}
    {{- $resultWithoutTrunc }}
  {{- end }}
{{- end }}
