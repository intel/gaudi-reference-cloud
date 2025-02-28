#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -x

helm uninstall ingress-nginx -n ingress-nginx
istioctl uninstall -y --purge

kubectl delete ns ingress-nginx
kubectl delete ns istio-system
kubectl delete ns idcs-system
kubectl delete ns idcs-otel
kubectl delete ns idcs-prometheus
kubectl delete ns habana-system

helm uninstall vault-secrets-operator -n vault-secrets-operator-system
kubectl delete ns vault-secrets-operator-system

#kubectl delete ClusterRole ingress-nginx
#kubectl delete ClusterRole ingress-nginx-admission
#kubectl delete Clusterrolebindings ingress-nginx
#kubectl delete Clusterrolebindings ingress-nginx-admission

set +x