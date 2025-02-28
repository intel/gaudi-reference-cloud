{{/*
Split sourceMap into chunks of chunkSize and return number of chunks
*/}}
{{- define "git-to-grpc-synchronizer.numberOfChunks" -}}
{{- $sourceMap := .Values.sourceMap }}
{{- $chunkSize := int (.Values.chunkSize) }}
{{- $counter := 0 }}
{{- $currentMap := dict }}
{{- $listOfMaps := list }}
{{- range $filename, $contents := $sourceMap }}
  {{- if eq $counter $chunkSize }}
    {{- $listOfMaps = append $listOfMaps $currentMap }}
    {{- $currentMap = dict }}
    {{- $counter = 0 }}
  {{- end }}
  {{- $currentMap = set $currentMap $filename $contents }}
  {{- $counter = add $counter 1}}
{{- end }}

{{- if gt $counter 0 }}
  {{- $listOfMaps = append $listOfMaps $currentMap }}
{{- end }}

{{- len $listOfMaps }}
{{- end }}


{{/*
Return the Job spec.
*/}}
{{- define "git-to-grpc-synchronizer.jobSpec" -}}
{{- $numberOfChunks := int (include "git-to-grpc-synchronizer.numberOfChunks" .) }}
spec:
  backoffLimit: 20
  ttlSecondsAfterFinished: {{ .Values.ttlSecondsAfterFinished }}
  template:
    metadata:
      annotations:
      {{- with .Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
        {{- include "idc-common.vaultAnnotations" . | nindent 8 }}
        {{- include "idc-common.otelAnnotations" . | nindent 8 }}
        {{- include "idc-common.vaultPkiAnnotations" . | nindent 8 }}
        vault.hashicorp.com/agent-pre-populate-only: "true"
        # Add checksum to ensure a new job is created if the configmap changes.
        checksum/configmap: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
    spec:
      restartPolicy: OnFailure
      containers:
        - name: main
          image: {{ include "idc-common.image" . }}
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          args:
            - "--kind"
            - {{ .Values.kind | quote }}
            - "--dir"
            - "/source"
            - "--target"
            - {{ .Values.target | quote }}
            {{- include "idc-common.logArgs" . | nindent 12 }}
          env:
            {{- include "idc-common.proxyEnv" . | nindent 12 }}
            {{- include "idc-common.commonEnv" . | nindent 12 }}
            {{- include "idc-common.otelEnv" . | nindent 12 }}
          volumeMounts:
          {{- range $i := until $numberOfChunks }}
          - mountPath: /source/source-{{ $i }}
            name: source-{{ $i }}
          {{- end }}
      securityContext:
        runAsNonRoot: true
      serviceAccountName: {{ include "idc-common.fullname" . }}
      volumes:
      {{- range $i := until $numberOfChunks }}
        - configMap:
            name: {{ include "idc-common.fullname" $ }}-{{ $i }}
          name: source-{{ $i }}
      {{- end }}
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
