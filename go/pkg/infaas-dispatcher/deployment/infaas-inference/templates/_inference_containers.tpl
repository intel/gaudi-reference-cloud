{# INTEL CONFIDENTIAL #}
{# Copyright (C) 2023 Intel Corporation #}


{{- define "tgi-server-containers" }}
{{- $modelName := .Values.deployedModel | required ".Values.deployedModel is required." -}}
{{ with $model := (index .Values.models $modelName) | required ".Values.deployedModel map content is required" }}
        {{ $shortModelName := (index $model "shortName") }}
        {{ $startupProbeInitialDelaySeconds := (index $model "startupProbeInitialDelaySeconds") }}

        - name: "tgi-server-{{$shortModelName}}" # vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
          securityContext:
            capabilities:
              add: ["SYS_NICE"]
          image: {{ $.Values.image.tgiServer.repository }}:{{ $.Values.image.tgiServer.tag }}
          imagePullPolicy: {{ $.Values.image.tgiServer.pullPolicy }}
          env:
            {{- $model.env | toYaml | nindent 12}}
            - name: MODEL_ID
              value: "/usr/src/models/llm"
            {{- if eq $.Values.precision "fp8" }}
            - name: QUANT_CONFIG
              value: "/usr/src/quantization_config/maxabs_quant.json"
            - name: WARMUP_ENABLED
              value: "false"
            {{- end }}
          args:
            {{- $model.args | toYaml | nindent 12}}
          startupProbe:
            httpGet:
              port: 8080
              path: /health
            initialDelaySeconds: {{ $startupProbeInitialDelaySeconds }}
            periodSeconds: 5
            timeoutSeconds: 2
            failureThreshold: 48
          livenessProbe:
            httpGet:
              port: 8080
              path: /health
            periodSeconds: 5
            timeoutSeconds: 2
            failureThreshold: 12
          resources:
            {{- $model.resources | toYaml | nindent 12 }}
          volumeMounts:
            - name: fp8-config-volume
              mountPath: /usr/src/quantization_config
              subPath: quantization_config
              readOnly: true
            - name: fp8-config-volume
              mountPath: /usr/src/hqt_output
              subPath: hqt_output
              readOnly: true
            - name: models
              mountPath: /usr/src/models
              readOnly: true
        # ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

        - name: "tgi-proxy-{{ $shortModelName }}" # vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
          image: "{{ $.Values.image.tgiProxy.repository }}:{{ $.Values.image.tgiProxy.tag }}"
          imagePullPolicy: {{ $.Values.image.tgiProxy.pullPolicy }}
          env:
            - name: base_url
              value: http://127.0.0.1:8080
            - name: safeguard_url
              value: http://infaas-safeguard:9000
            - name: server_config
              value: /server.yaml
          ports:
            - name: grpc
              containerPort: 50051
              protocol: TCP
          startupProbe:
            grpc:
              port: 50051
              service: startup
            initialDelaySeconds: {{ $startupProbeInitialDelaySeconds }}
            periodSeconds: 5
            timeoutSeconds: 2
            failureThreshold: 50
          livenessProbe:
            grpc:
              port: 50051
              service: liveness
            periodSeconds: 2
            timeoutSeconds: 2
            failureThreshold: 10
          readinessProbe:
            grpc:
              port: 50051
              service: readiness
            initialDelaySeconds: 5
            periodSeconds: 2
            timeoutSeconds: 2
            failureThreshold: 5
          resources:
            requests:
              cpu: "4"
            limits:
              cpu: "4"
          volumeMounts:
            - name: models
              mountPath: /app/models
              readOnly: true
            - name: server
              mountPath: /server.yaml
              subPath: server.yaml
        # ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

{{- end }}
{{- end }}