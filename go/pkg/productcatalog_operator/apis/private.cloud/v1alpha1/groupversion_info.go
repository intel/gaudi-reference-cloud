// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Package v1alpha1 contains API Schema definitions for the cloud v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=private.cloud.intel.com
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const GroupName = "private.cloud.intel.com"

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "private.cloud.intel.com", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return GroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}
