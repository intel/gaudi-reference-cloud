{{- define "vm-machine-image-resources.VirtualMachineImage" }}
{{- $os := "" }}
{{- range $component := .spec.components }}
  {{- if (eq "OS" $component.type) }}
    {{- if (lower $component.name | contains "ubuntu") }}
      {{- $os = "ubuntu" }}
    {{- end }}
  {{- end }}
{{- end }}
apiVersion: harvesterhci.io/v1beta1
kind: VirtualMachineImage
metadata:
  labels:
    harvesterhci.io/image-type: raw_qcow2
    harvesterhci.io/imageDisplayName: {{ .metadata.name }}
    {{- if not (empty $os) }}
    harvesterhci.io/os-type: {{ $os }}
    {{- end }}
  name: {{ .metadata.name }}
  namespace: default
spec:
  {{- if .spec.sha512sum }}
  checksum: {{ .spec.sha512sum }}
  {{- end }}
  displayName: {{ .metadata.name }}
  sourceType: download
  url: {{ .urlPrefix }}/{{ .metadata.name }}.qcow2
{{- end }}

{{- define "vm-machine-image-resources.DataVolume" }}
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: {{ .metadata.name }}
  namespace: default
spec:
  source:
    http:
      url: {{ .urlPrefix }}/{{ .metadata.name }}.qcow2
  storage:
    accessModes:
    - ReadOnlyMany
    resources:
      requests:
        storage: {{ .spec.virtualSizeBytes }}
{{- end }}
