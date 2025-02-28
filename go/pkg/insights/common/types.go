// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import (
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	ComponentTypeImage            = "image"
	ComponentTypeGitrepo          = "gitrepo"
	ComponentTypeFile             = "file"
	ComponentTypeCompressedBundle = "compressed_bundle"
	ComponentTypeUnspecified      = "unspecified"

	KubeVendorOSS         = "oss"
	KubeVendorRancher     = "rancher"
	KubeVendorUnspecified = "unspecified"
)

func MapComponentTypeToPB(strType string) v1.ComponentType {
	switch strType {
	case ComponentTypeImage:
		return v1.ComponentType_OCI_IMAGE
	case ComponentTypeCompressedBundle:
		return v1.ComponentType_COMPRESSED_BUNDLE
	case ComponentTypeFile:
		return v1.ComponentType_FILE
	case ComponentTypeGitrepo:
		return v1.ComponentType_GIT_REPO
	case ComponentTypeUnspecified:
		return v1.ComponentType_UNSPECIFIED_TYPE
	default:
		return v1.ComponentType_UNSPECIFIED_TYPE
	}
}

func MapComponentTypeToSQL(pbT v1.ComponentType) string {
	switch pbT {
	case v1.ComponentType_OCI_IMAGE:
		return ComponentTypeImage
	case v1.ComponentType_COMPRESSED_BUNDLE:
		return ComponentTypeCompressedBundle
	case v1.ComponentType_FILE:
		return ComponentTypeFile
	case v1.ComponentType_GIT_REPO:
		return ComponentTypeGitrepo
	case v1.ComponentType_UNSPECIFIED_TYPE:
		return ComponentTypeUnspecified
	default:
		return ComponentTypeUnspecified
	}
}

func MapOSSVendorTypeToPB(v string) v1.ValidVendors {
	if v == KubeVendorOSS {
		return v1.ValidVendors_OSS_KUBE_VENDOR
	} else if v == KubeVendorRancher {
		return v1.ValidVendors_RANCHER_KUBE_VENDOR
	}
	return v1.ValidVendors_UNSPECIFIED_KUBE_VENDOR
}

func MapOSSVendorTypeToSQL(v v1.ValidVendors) string {
	if v == v1.ValidVendors_OSS_KUBE_VENDOR {
		return KubeVendorOSS
	} else if v == v1.ValidVendors_RANCHER_KUBE_VENDOR {
		return KubeVendorRancher
	}
	return KubeVendorUnspecified
}
