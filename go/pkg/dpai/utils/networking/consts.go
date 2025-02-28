// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

const (
	IstioNamespace       = "istio-system"
	CertManagerNamesapce = "cert-manager"
	OpenEbsNamespace     = "openebs"
	SecretNamespace      = "secrets"
	DockerRegistrySecret = "docker-registry-secret"
)

// cert manager config for the gateway
const (
	IstioRootClusterIssuer            = "istio-self-sign-issuer"
	IstioRootClusterIssuerCertificate = "istio-ca"
	IstioK8SCSR                       = "istio-ca-issuer"
	CLUSTER_ISSUER                    = "ClusterIssuer"   // apiVersion Kind
	CLUSTER_ISSUER_API_VERSION        = "cert-manager.io" // api version for the cert manager
	CLUSTER_ROOT_SECRET               = "root-secret"     // used for the gateway configuration
)

// ingreess lb operator vm instances operator reference key
const (
	IngressIKSNodeLabelKey = "nodegroupName"
)

// node selector deployment for pod affinities
const (
	NodeSelectorLabelKey = "kubernetes.io/hostname"
)

// LoadBalancer operator SSL Profile configuration
const (
	DPAI_HIGHWIRE_SSL_PROFILE = "lbauto-dpai-cloudworkspace-io-ssl"
	DPAI_HIGHWIRE_LB_NAME     = "dpai-tls-ingress-lb"
	DPAI_HIGHWIRE_POOL_PORT   = 443 // 443 is the default port for all tls traffic inbound to any DPAI service
)
