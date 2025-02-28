# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
apiVersion: apps/v1
kind: Deployment
metadata:
  name: storage-controller
spec:
  template:
    spec:
      containers:
      - name: storage-controller
        image: {{ .Repository }}@{{ .Digest }}
