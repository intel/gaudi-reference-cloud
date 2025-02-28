// Boilerplate code taken from k8s apiextensions library
// TODO: check feasibility of unstructured yaml to json conversion to remove boilerplate code 
// reference - "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CustomResourceDefinitionExtV1 struct {
	metav1.TypeMeta `yaml:",inline"`
	// Standard object's metadata
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec describes how the user wants the resources to appear
	Spec CustomResourceDefinitionSpec `yaml:"spec" protobuf:"bytes,2,opt,name=spec"`
	// status indicates the actual state of the CustomResourceDefinition
	// +optional
	Status CustomResourceDefinitionStatus `yaml:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// CustomResourceDefinitionSpec describes how a user wants their resource to appear
type CustomResourceDefinitionSpec struct {
	// group is the API group of the defined custom resource.
	// The custom resources are served under `/apis/<group>/...`.
	// Must match the name of the CustomResourceDefinition (in the form `<names.plural>.<group>`).
	Group string `yaml:"group" protobuf:"bytes,1,opt,name=group"`
	// names specify the resource and kind names for the custom resource.
	Names CustomResourceDefinitionNames `yaml:"names" protobuf:"bytes,3,opt,name=names"`
	// scope indicates whether the defined custom resource is cluster- or namespace-scoped.
	// Allowed values are `Cluster` and `Namespaced`.
	Scope ResourceScope `yaml:"scope" protobuf:"bytes,4,opt,name=scope,casttype=ResourceScope"`
	// versions is the list of all API versions of the defined custom resource.
	// Version names are used to compute the order in which served versions are listed in API discovery.
	// If the version string is "kube-like", it will sort above non "kube-like" version strings, which are ordered
	// lexicographically. "Kube-like" versions start with a "v", then are followed by a number (the major version),
	// then optionally the string "alpha" or "beta" and another number (the minor version). These are sorted first
	// by GA > beta > alpha (where GA is a version with no suffix such as beta or alpha), and then by comparing
	// major version, then minor version. An example sorted list of versions:
	// v10, v2, v1, v11beta2, v10beta3, v3beta1, v12alpha1, v11alpha2, foo1, foo10.
	Versions []CustomResourceDefinitionVersion `yaml:"versions" protobuf:"bytes,7,rep,name=versions"`

	// conversion defines conversion settings for the CRD.
	// +optional
	Conversion *CustomResourceConversion `yaml:"conversion,omitempty" protobuf:"bytes,9,opt,name=conversion"`

	// preserveUnknownFields indicates that object fields which are not specified
	// in the OpenAPI schema should be preserved when persisting to storage.
	// apiVersion, kind, metadata and known fields inside metadata are always preserved.
	// This field is deprecated in favor of setting `x-preserve-unknown-fields` to true in `spec.versions[*].schema.openAPIV3Schema`.
	// See https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#pruning-versus-preserving-unknown-fields for details.
	// +optional
	PreserveUnknownFields bool `yaml:"preserveUnknownFields,omitempty" protobuf:"varint,10,opt,name=preserveUnknownFields"`
}

// CustomResourceDefinitionStatus indicates the state of the CustomResourceDefinition
type CustomResourceDefinitionStatus struct {
	// conditions indicate state for particular aspects of a CustomResourceDefinition
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []CustomResourceDefinitionCondition `yaml:"conditions" protobuf:"bytes,1,opt,name=conditions"`

	// acceptedNames are the names that are actually being used to serve discovery.
	// They may be different than the names in spec.
	// +optional
	AcceptedNames CustomResourceDefinitionNames `yaml:"acceptedNames" protobuf:"bytes,2,opt,name=acceptedNames"`

	// storedVersions lists all versions of CustomResources that were ever persisted. Tracking these
	// versions allows a migration path for stored versions in etcd. The field is mutable
	// so a migration controller can finish a migration to another version (ensuring
	// no old objects are left in storage), and then remove the rest of the
	// versions from this list.
	// Versions may not be removed from `spec.versions` while they exist in this list.
	// +optional
	StoredVersions []string `yaml:"storedVersions" protobuf:"bytes,3,rep,name=storedVersions"`
}

// CustomResourceDefinitionNames indicates the names to serve this CustomResourceDefinition
type CustomResourceDefinitionNames struct {
	// plural is the plural name of the resource to serve.
	// The custom resources are served under `/apis/<group>/<version>/.../<plural>`.
	// Must match the name of the CustomResourceDefinition (in the form `<names.plural>.<group>`).
	// Must be all lowercase.
	Plural string `yaml:"plural" protobuf:"bytes,1,opt,name=plural"`
	// singular is the singular name of the resource. It must be all lowercase. Defaults to lowercased `kind`.
	// +optional
	Singular string `yaml:"singular,omitempty" protobuf:"bytes,2,opt,name=singular"`
	// shortNames are short names for the resource, exposed in API discovery documents,
	// and used by clients to support invocations like `kubectl get <shortname>`.
	// It must be all lowercase.
	// +optional
	ShortNames []string `yaml:"shortNames,omitempty" protobuf:"bytes,3,opt,name=shortNames"`
	// kind is the serialized kind of the resource. It is normally CamelCase and singular.
	// Custom resource instances will use this value as the `kind` attribute in API calls.
	Kind string `yaml:"kind" protobuf:"bytes,4,opt,name=kind"`
	// listKind is the serialized kind of the list for this resource. Defaults to "`kind`List".
	// +optional
	ListKind string `yaml:"listKind,omitempty" protobuf:"bytes,5,opt,name=listKind"`
	// categories is a list of grouped resources this custom resource belongs to (e.g. 'all').
	// This is published in API discovery documents, and used by clients to support invocations like
	// `kubectl get all`.
	// +optional
	Categories []string `yaml:"categories,omitempty" protobuf:"bytes,6,rep,name=categories"`
}

// ResourceScope is an enum defining the different scopes available to a custom resource
type ResourceScope string
const (
	ClusterScoped   ResourceScope = "Cluster"
	NamespaceScoped ResourceScope = "Namespaced"
)

// CustomResourceDefinitionVersion describes a version for CRD.
type CustomResourceDefinitionVersion struct {
	// name is the version name, e.g. “v1”, “v2beta1”, etc.
	// The custom resources are served under this version at `/apis/<group>/<version>/...` if `served` is true.
	Name string `yaml:"name" protobuf:"bytes,1,opt,name=name"`
	// served is a flag enabling/disabling this version from being served via REST APIs
	Served bool `yaml:"served" protobuf:"varint,2,opt,name=served"`
	// storage indicates this version should be used when persisting custom resources to storage.
	// There must be exactly one version with storage=true.
	Storage bool `yaml:"storage" protobuf:"varint,3,opt,name=storage"`
	// deprecated indicates this version of the custom resource API is deprecated.
	// When set to true, API requests to this version receive a warning header in the server response.
	// Defaults to false.
	// +optional
	Deprecated bool `yaml:"deprecated,omitempty" protobuf:"varint,7,opt,name=deprecated"`
	// deprecationWarning overrides the default warning returned to API clients.
	// May only be set when `deprecated` is true.
	// The default warning indicates this version is deprecated and recommends use
	// of the newest served version of equal or greater stability, if one exists.
	// +optional
	DeprecationWarning *string `yaml:"deprecationWarning,omitempty" protobuf:"bytes,8,opt,name=deprecationWarning"`
	// schema describes the schema used for validation, pruning, and defaulting of this version of the custom resource.
	// +optional
	Schema *CustomResourceValidation `yaml:"schema,omitempty" protobuf:"bytes,4,opt,name=schema"`
	// subresources specify what subresources this version of the defined custom resource have.
	// +optional
	Subresources *CustomResourceSubresources `yaml:"subresources,omitempty" protobuf:"bytes,5,opt,name=subresources"`
	// additionalPrinterColumns specifies additional columns returned in Table output.
	// See https://kubernetes.io/docs/reference/using-api/api-concepts/#receiving-resources-as-tables for details.
	// If no columns are specified, a single column displaying the age of the custom resource is used.
	// +optional
	AdditionalPrinterColumns []CustomResourceColumnDefinition `yaml:"additionalPrinterColumns,omitempty" protobuf:"bytes,6,rep,name=additionalPrinterColumns"`
}

// CustomResourceConversion describes how to convert different versions of a CR.
type CustomResourceConversion struct {
	// strategy specifies how custom resources are converted between versions. Allowed values are:
	// - `None`: The converter only change the apiVersion and would not touch any other field in the custom resource.
	// - `Webhook`: API Server will call to an external webhook to do the conversion. Additional information
	//   is needed for this option. This requires spec.preserveUnknownFields to be false, and spec.conversion.webhook to be set.
	Strategy ConversionStrategyType `yaml:"strategy" protobuf:"bytes,1,name=strategy"`

	// webhook describes how to call the conversion webhook. Required when `strategy` is set to `Webhook`.
	// +optional
	Webhook *WebhookConversion `yaml:"webhook,omitempty" protobuf:"bytes,2,opt,name=webhook"`
}

// CustomResourceDefinitionCondition contains details for the current condition of this pod.
type CustomResourceDefinitionCondition struct {
	// type is the type of the condition. Types include Established, NamesAccepted and Terminating.
	Type CustomResourceDefinitionConditionType `yaml:"type" protobuf:"bytes,1,opt,name=type,casttype=CustomResourceDefinitionConditionType"`
	// status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `yaml:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// lastTransitionTime last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `yaml:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// reason is a unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `yaml:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// message is a human-readable message indicating details about last transition.
	// +optional
	Message string `yaml:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// CustomResourceValidation is a list of validation methods for CustomResources.
type CustomResourceValidation struct {
	// openAPIV3Schema is the OpenAPI v3 schema to use for validation and pruning.
	// +optional
	OpenAPIV3Schema *JSONSchemaProps `yaml:"openAPIV3Schema,omitempty" protobuf:"bytes,1,opt,name=openAPIV3Schema"`
}

// CustomResourceSubresources defines the status and scale subresources for CustomResources.
type CustomResourceSubresources struct {
	// status indicates the custom resource should serve a `/status` subresource.
	// When enabled:
	// 1. requests to the custom resource primary endpoint ignore changes to the `status` stanza of the object.
	// 2. requests to the custom resource `/status` subresource ignore changes to anything other than the `status` stanza of the object.
	// +optional
	Status *CustomResourceSubresourceStatus `yaml:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
	// scale indicates the custom resource should serve a `/scale` subresource that returns an `autoscaling/v1` Scale object.
	// +optional
	Scale *CustomResourceSubresourceScale `yaml:"scale,omitempty" protobuf:"bytes,2,opt,name=scale"`
}

// CustomResourceColumnDefinition specifies a column for server side printing.
type CustomResourceColumnDefinition struct {
	// name is a human readable name for the column.
	Name string `yaml:"name" protobuf:"bytes,1,opt,name=name"`
	// type is an OpenAPI type definition for this column.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for details.
	Type string `yaml:"type" protobuf:"bytes,2,opt,name=type"`
	// format is an optional OpenAPI type definition for this column. The 'name' format is applied
	// to the primary identifier column to assist in clients identifying column is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for details.
	// +optional
	Format string `yaml:"format,omitempty" protobuf:"bytes,3,opt,name=format"`
	// description is a human readable description of this column.
	// +optional
	Description string `yaml:"description,omitempty" protobuf:"bytes,4,opt,name=description"`
	// priority is an integer defining the relative importance of this column compared to others. Lower
	// numbers are considered higher priority. Columns that may be omitted in limited space scenarios
	// should be given a priority greater than 0.
	// +optional
	Priority int32 `yaml:"priority,omitempty" protobuf:"bytes,5,opt,name=priority"`
	// jsonPath is a simple JSON path (i.e. with array notation) which is evaluated against
	// each custom resource to produce the value for this column.
	JSONPath string `yaml:"jsonPath" protobuf:"bytes,6,opt,name=jsonPath"`
}

// ConversionStrategyType describes different conversion types.
type ConversionStrategyType string

const (
	// KubeAPIApprovedAnnotation is an annotation that must be set to create a CRD for the k8s.io, *.k8s.io, kubernetes.io, or *.kubernetes.io namespaces.
	// The value should be a link to a URL where the current spec was approved, so updates to the spec should also update the URL.
	// If the API is unapproved, you may set the annotation to a string starting with `"unapproved"`.  For instance, `"unapproved, temporarily squatting"` or `"unapproved, experimental-only"`.  This is discouraged.
	KubeAPIApprovedAnnotation = "api-approved.kubernetes.io"

	// NoneConverter is a converter that only sets apiversion of the CR and leave everything else unchanged.
	NoneConverter ConversionStrategyType = "None"
	// WebhookConverter is a converter that calls to an external webhook to convert the CR.
	WebhookConverter ConversionStrategyType = "Webhook"
)

// WebhookConversion describes how to call a conversion webhook
type WebhookConversion struct {
	// clientConfig is the instructions for how to call the webhook if strategy is `Webhook`.
	// +optional
	ClientConfig *WebhookClientConfig `yaml:"clientConfig,omitempty" protobuf:"bytes,2,name=clientConfig"`

	// conversionReviewVersions is an ordered list of preferred `ConversionReview`
	// versions the Webhook expects. The API server will use the first version in
	// the list which it supports. If none of the versions specified in this list
	// are supported by API server, conversion will fail for the custom resource.
	// If a persisted Webhook configuration specifies allowed versions and does not
	// include any versions known to the API Server, calls to the webhook will fail.
	ConversionReviewVersions []string `yaml:"conversionReviewVersions" protobuf:"bytes,3,rep,name=conversionReviewVersions"`
}

// CustomResourceDefinitionConditionType is a valid value for CustomResourceDefinitionCondition.Type
type CustomResourceDefinitionConditionType string

const (
	// Established means that the resource has become active. A resource is established when all names are
	// accepted without a conflict for the first time. A resource stays established until deleted, even during
	// a later NamesAccepted due to changed names. Note that not all names can be changed.
	Established CustomResourceDefinitionConditionType = "Established"
	// NamesAccepted means the names chosen for this CustomResourceDefinition do not conflict with others in
	// the group and are therefore accepted.
	NamesAccepted CustomResourceDefinitionConditionType = "NamesAccepted"
	// NonStructuralSchema means that one or more OpenAPI schema is not structural.
	//
	// A schema is structural if it specifies types for all values, with the only exceptions of those with
	// - x-kubernetes-int-or-string: true — for fields which can be integer or string
	// - x-kubernetes-preserve-unknown-fields: true — for raw, unspecified JSON values
	// and there is no type, additionalProperties, default, nullable or x-kubernetes-* vendor extenions
	// specified under allOf, anyOf, oneOf or not.
	//
	// Non-structural schemas will not be allowed anymore in v1 API groups. Moreover, new features will not be
	// available for non-structural CRDs:
	// - pruning
	// - defaulting
	// - read-only
	// - OpenAPI publishing
	// - webhook conversion
	NonStructuralSchema CustomResourceDefinitionConditionType = "NonStructuralSchema"
	// Terminating means that the CustomResourceDefinition has been deleted and is cleaning up.
	Terminating CustomResourceDefinitionConditionType = "Terminating"
	// KubernetesAPIApprovalPolicyConformant indicates that an API in *.k8s.io or *.kubernetes.io is or is not approved.  For CRDs
	// outside those groups, this condition will not be set.  For CRDs inside those groups, the condition will
	// be true if .metadata.annotations["api-approved.kubernetes.io"] is set to a URL, otherwise it will be false.
	// See https://github.com/kubernetes/enhancements/pull/1111 for more details.
	KubernetesAPIApprovalPolicyConformant CustomResourceDefinitionConditionType = "KubernetesAPIApprovalPolicyConformant"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// CustomResourceSubresourceStatus defines how to serve the status subresource for CustomResources.
// Status is represented by the `.status` JSON path inside of a CustomResource. When set,
// * exposes a /status subresource for the custom resource
// * PUT requests to the /status subresource take a custom resource object, and ignore changes to anything except the status stanza
// * PUT/POST/PATCH requests to the custom resource ignore changes to the status stanza
type CustomResourceSubresourceStatus struct{}

// CustomResourceSubresourceScale defines how to serve the scale subresource for CustomResources.
type CustomResourceSubresourceScale struct {
	// specReplicasPath defines the JSON path inside of a custom resource that corresponds to Scale `spec.replicas`.
	// Only JSON paths without the array notation are allowed.
	// Must be a JSON Path under `.spec`.
	// If there is no value under the given path in the custom resource, the `/scale` subresource will return an error on GET.
	SpecReplicasPath string `yaml:"specReplicasPath" protobuf:"bytes,1,name=specReplicasPath"`
	// statusReplicasPath defines the JSON path inside of a custom resource that corresponds to Scale `status.replicas`.
	// Only JSON paths without the array notation are allowed.
	// Must be a JSON Path under `.status`.
	// If there is no value under the given path in the custom resource, the `status.replicas` value in the `/scale` subresource
	// will default to 0.
	StatusReplicasPath string `yaml:"statusReplicasPath" protobuf:"bytes,2,opt,name=statusReplicasPath"`
	// labelSelectorPath defines the JSON path inside of a custom resource that corresponds to Scale `status.selector`.
	// Only JSON paths without the array notation are allowed.
	// Must be a JSON Path under `.status` or `.spec`.
	// Must be set to work with HorizontalPodAutoscaler.
	// The field pointed by this JSON path must be a string field (not a complex selector struct)
	// which contains a serialized label selector in string form.
	// More info: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions#scale-subresource
	// If there is no value under the given path in the custom resource, the `status.selector` value in the `/scale`
	// subresource will default to the empty string.
	// +optional
	LabelSelectorPath *string `yaml:"labelSelectorPath,omitempty" protobuf:"bytes,3,opt,name=labelSelectorPath"`
}

// WebhookClientConfig contains the information to make a TLS connection with the webhook.
type WebhookClientConfig struct {
	// url gives the location of the webhook, in standard URL form
	// (`scheme://host:port/path`). Exactly one of `url` or `service`
	// must be specified.
	//
	// The `host` should not refer to a service running in the cluster; use
	// the `service` field instead. The host might be resolved via external
	// DNS in some apiservers (e.g., `kube-apiserver` cannot resolve
	// in-cluster DNS as that would be a layering violation). `host` may
	// also be an IP address.
	//
	// Please note that using `localhost` or `127.0.0.1` as a `host` is
	// risky unless you take great care to run this webhook on all hosts
	// which run an apiserver which might need to make calls to this
	// webhook. Such installs are likely to be non-portable, i.e., not easy
	// to turn up in a new cluster.
	//
	// The scheme must be "https"; the URL must begin with "https://".
	//
	// A path is optional, and if present may be any string permissible in
	// a URL. You may use the path to pass an arbitrary string to the
	// webhook, for example, a cluster identifier.
	//
	// Attempting to use a user or basic auth e.g. "user:password@" is not
	// allowed. Fragments ("#...") and query parameters ("?...") are not
	// allowed, either.
	//
	// +optional
	URL *string `json:"url,omitempty" protobuf:"bytes,3,opt,name=url"`

	// service is a reference to the service for this webhook. Either
	// service or url must be specified.
	//
	// If the webhook is running within the cluster, then you should use `service`.
	//
	// +optional
	Service *ServiceReference `yaml:"service,omitempty" protobuf:"bytes,1,opt,name=service"`

	// caBundle is a PEM encoded CA bundle which will be used to validate the webhook's server certificate.
	// If unspecified, system trust roots on the apiserver are used.
	// +optional
	CABundle []byte `yaml:"caBundle,omitempty" protobuf:"bytes,2,opt,name=caBundle"`
}

// ServiceReference holds a reference to Service.legacy.k8s.io
type ServiceReference struct {
	// namespace is the namespace of the service.
	// Required
	Namespace string `yaml:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	// name is the name of the service.
	// Required
	Name string `yaml:"name" protobuf:"bytes,2,opt,name=name"`

	// path is an optional URL path at which the webhook will be contacted.
	// +optional
	Path *string `yaml:"path,omitempty" protobuf:"bytes,3,opt,name=path"`

	// port is an optional service port at which the webhook will be contacted.
	// `port` should be a valid port number (1-65535, inclusive).
	// Defaults to 443 for backward compatibility.
	// +optional
	Port *int32 `yaml:"port,omitempty" protobuf:"varint,4,opt,name=port"`
}

// JSONSchemaProps is a JSON-Schema following Specification Draft 4 (http://json-schema.org/).
type JSONSchemaProps struct {
	ID          string        `json:"id,omitempty" protobuf:"bytes,1,opt,name=id"`
	Schema      JSONSchemaURL `json:"$schema,omitempty" protobuf:"bytes,2,opt,name=schema"`
	Ref         *string       `json:"$ref,omitempty" protobuf:"bytes,3,opt,name=ref"`
	Description string        `json:"description,omitempty" protobuf:"bytes,4,opt,name=description"`
	Type        string        `json:"type,omitempty" protobuf:"bytes,5,opt,name=type"`

	// format is an OpenAPI v3 format string. Unknown formats are ignored. The following formats are validated:
	//
	// - bsonobjectid: a bson object ID, i.e. a 24 characters hex string
	// - uri: an URI as parsed by Golang net/url.ParseRequestURI
	// - email: an email address as parsed by Golang net/mail.ParseAddress
	// - hostname: a valid representation for an Internet host name, as defined by RFC 1034, section 3.1 [RFC1034].
	// - ipv4: an IPv4 IP as parsed by Golang net.ParseIP
	// - ipv6: an IPv6 IP as parsed by Golang net.ParseIP
	// - cidr: a CIDR as parsed by Golang net.ParseCIDR
	// - mac: a MAC address as parsed by Golang net.ParseMAC
	// - uuid: an UUID that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$
	// - uuid3: an UUID3 that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?3[0-9a-f]{3}-?[0-9a-f]{4}-?[0-9a-f]{12}$
	// - uuid4: an UUID4 that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?4[0-9a-f]{3}-?[89ab][0-9a-f]{3}-?[0-9a-f]{12}$
	// - uuid5: an UUID5 that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?5[0-9a-f]{3}-?[89ab][0-9a-f]{3}-?[0-9a-f]{12}$
	// - isbn: an ISBN10 or ISBN13 number string like "0321751043" or "978-0321751041"
	// - isbn10: an ISBN10 number string like "0321751043"
	// - isbn13: an ISBN13 number string like "978-0321751041"
	// - creditcard: a credit card number defined by the regex ^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})$ with any non digit characters mixed in
	// - ssn: a U.S. social security number following the regex ^\\d{3}[- ]?\\d{2}[- ]?\\d{4}$
	// - hexcolor: an hexadecimal color code like "#FFFFFF: following the regex ^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$
	// - rgbcolor: an RGB color code like rgb like "rgb(255,255,2559"
	// - byte: base64 encoded binary data
	// - password: any kind of string
	// - date: a date string like "2006-01-02" as defined by full-date in RFC3339
	// - duration: a duration string like "22 ns" as parsed by Golang time.ParseDuration or compatible with Scala duration format
	// - datetime: a date time string like "2014-12-15T19:30:20.000Z" as defined by date-time in RFC3339.
	Format string `json:"format,omitempty" protobuf:"bytes,6,opt,name=format"`

	Title string `json:"title,omitempty" protobuf:"bytes,7,opt,name=title"`
	// default is a default value for undefined object fields.
	// Defaulting is a beta feature under the CustomResourceDefaulting feature gate.
	// Defaulting requires spec.preserveUnknownFields to be false.
	Default              *JSON                      `json:"default,omitempty" protobuf:"bytes,8,opt,name=default"`
	Maximum              *float64                   `json:"maximum,omitempty" protobuf:"bytes,9,opt,name=maximum"`
	ExclusiveMaximum     bool                       `json:"exclusiveMaximum,omitempty" protobuf:"bytes,10,opt,name=exclusiveMaximum"`
	Minimum              *float64                   `json:"minimum,omitempty" protobuf:"bytes,11,opt,name=minimum"`
	ExclusiveMinimum     bool                       `json:"exclusiveMinimum,omitempty" protobuf:"bytes,12,opt,name=exclusiveMinimum"`
	MaxLength            *int64                     `json:"maxLength,omitempty" protobuf:"bytes,13,opt,name=maxLength"`
	MinLength            *int64                     `json:"minLength,omitempty" protobuf:"bytes,14,opt,name=minLength"`
	Pattern              string                     `json:"pattern,omitempty" protobuf:"bytes,15,opt,name=pattern"`
	MaxItems             *int64                     `json:"maxItems,omitempty" protobuf:"bytes,16,opt,name=maxItems"`
	MinItems             *int64                     `json:"minItems,omitempty" protobuf:"bytes,17,opt,name=minItems"`
	UniqueItems          bool                       `json:"uniqueItems,omitempty" protobuf:"bytes,18,opt,name=uniqueItems"`
	MultipleOf           *float64                   `json:"multipleOf,omitempty" protobuf:"bytes,19,opt,name=multipleOf"`
	Enum                 []JSON                     `json:"enum,omitempty" protobuf:"bytes,20,rep,name=enum"`
	MaxProperties        *int64                     `json:"maxProperties,omitempty" protobuf:"bytes,21,opt,name=maxProperties"`
	MinProperties        *int64                     `json:"minProperties,omitempty" protobuf:"bytes,22,opt,name=minProperties"`
	Required             []string                   `json:"required,omitempty" protobuf:"bytes,23,rep,name=required"`
	Items                *JSONSchemaPropsOrArray    `json:"items,omitempty" protobuf:"bytes,24,opt,name=items"`
	AllOf                []JSONSchemaProps          `json:"allOf,omitempty" protobuf:"bytes,25,rep,name=allOf"`
	OneOf                []JSONSchemaProps          `json:"oneOf,omitempty" protobuf:"bytes,26,rep,name=oneOf"`
	AnyOf                []JSONSchemaProps          `json:"anyOf,omitempty" protobuf:"bytes,27,rep,name=anyOf"`
	Not                  *JSONSchemaProps           `json:"not,omitempty" protobuf:"bytes,28,opt,name=not"`
	Properties           map[string]JSONSchemaProps `json:"properties,omitempty" protobuf:"bytes,29,rep,name=properties"`
	AdditionalProperties *JSONSchemaPropsOrBool     `json:"additionalProperties,omitempty" protobuf:"bytes,30,opt,name=additionalProperties"`
	PatternProperties    map[string]JSONSchemaProps `json:"patternProperties,omitempty" protobuf:"bytes,31,rep,name=patternProperties"`
	Dependencies         JSONSchemaDependencies     `json:"dependencies,omitempty" protobuf:"bytes,32,opt,name=dependencies"`
	AdditionalItems      *JSONSchemaPropsOrBool     `json:"additionalItems,omitempty" protobuf:"bytes,33,opt,name=additionalItems"`
	Definitions          JSONSchemaDefinitions      `json:"definitions,omitempty" protobuf:"bytes,34,opt,name=definitions"`
	ExternalDocs         *ExternalDocumentation     `json:"externalDocs,omitempty" protobuf:"bytes,35,opt,name=externalDocs"`
	Example              *JSON                      `json:"example,omitempty" protobuf:"bytes,36,opt,name=example"`
	Nullable             bool                       `json:"nullable,omitempty" protobuf:"bytes,37,opt,name=nullable"`

	// x-kubernetes-preserve-unknown-fields stops the API server
	// decoding step from pruning fields which are not specified
	// in the validation schema. This affects fields recursively,
	// but switches back to normal pruning behaviour if nested
	// properties or additionalProperties are specified in the schema.
	// This can either be true or undefined. False is forbidden.
	XPreserveUnknownFields *bool `json:"x-kubernetes-preserve-unknown-fields,omitempty" protobuf:"bytes,38,opt,name=xKubernetesPreserveUnknownFields"`

	// x-kubernetes-embedded-resource defines that the value is an
	// embedded Kubernetes runtime.Object, with TypeMeta and
	// ObjectMeta. The type must be object. It is allowed to further
	// restrict the embedded object. kind, apiVersion and metadata
	// are validated automatically. x-kubernetes-preserve-unknown-fields
	// is allowed to be true, but does not have to be if the object
	// is fully specified (up to kind, apiVersion, metadata).
	XEmbeddedResource bool `json:"x-kubernetes-embedded-resource,omitempty" protobuf:"bytes,39,opt,name=xKubernetesEmbeddedResource"`

	// x-kubernetes-int-or-string specifies that this value is
	// either an integer or a string. If this is true, an empty
	// type is allowed and type as child of anyOf is permitted
	// if following one of the following patterns:
	//
	// 1) anyOf:
	//    - type: integer
	//    - type: string
	// 2) allOf:
	//    - anyOf:
	//      - type: integer
	//      - type: string
	//    - ... zero or more
	XIntOrString bool `json:"x-kubernetes-int-or-string,omitempty" protobuf:"bytes,40,opt,name=xKubernetesIntOrString"`

	// x-kubernetes-list-map-keys annotates an array with the x-kubernetes-list-type `map` by specifying the keys used
	// as the index of the map.
	//
	// This tag MUST only be used on lists that have the "x-kubernetes-list-type"
	// extension set to "map". Also, the values specified for this attribute must
	// be a scalar typed field of the child structure (no nesting is supported).
	//
	// The properties specified must either be required or have a default value,
	// to ensure those properties are present for all list items.
	//
	// +optional
	XListMapKeys []string `json:"x-kubernetes-list-map-keys,omitempty" protobuf:"bytes,41,rep,name=xKubernetesListMapKeys"`

	// x-kubernetes-list-type annotates an array to further describe its topology.
	// This extension must only be used on lists and may have 3 possible values:
	//
	// 1) `atomic`: the list is treated as a single entity, like a scalar.
	//      Atomic lists will be entirely replaced when updated. This extension
	//      may be used on any type of list (struct, scalar, ...).
	// 2) `set`:
	//      Sets are lists that must not have multiple items with the same value. Each
	//      value must be a scalar, an object with x-kubernetes-map-type `atomic` or an
	//      array with x-kubernetes-list-type `atomic`.
	// 3) `map`:
	//      These lists are like maps in that their elements have a non-index key
	//      used to identify them. Order is preserved upon merge. The map tag
	//      must only be used on a list with elements of type object.
	// Defaults to atomic for arrays.
	// +optional
	XListType *string `json:"x-kubernetes-list-type,omitempty" protobuf:"bytes,42,opt,name=xKubernetesListType"`

	// x-kubernetes-map-type annotates an object to further describe its topology.
	// This extension must only be used when type is object and may have 2 possible values:
	//
	// 1) `granular`:
	//      These maps are actual maps (key-value pairs) and each fields are independent
	//      from each other (they can each be manipulated by separate actors). This is
	//      the default behaviour for all maps.
	// 2) `atomic`: the list is treated as a single entity, like a scalar.
	//      Atomic maps will be entirely replaced when updated.
	// +optional
	XMapType *string `json:"x-kubernetes-map-type,omitempty" protobuf:"bytes,43,opt,name=xKubernetesMapType"`

	// x-kubernetes-validations describes a list of validation rules written in the CEL expression language.
	// This field is an alpha-level. Using this field requires the feature gate `CustomResourceValidationExpressions` to be enabled.
	// +patchMergeKey=rule
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=rule
	XValidations ValidationRules `json:"x-kubernetes-validations,omitempty" patchStrategy:"merge" patchMergeKey:"rule" protobuf:"bytes,44,rep,name=xKubernetesValidations"`
}

// ValidationRules describes a list of validation rules written in the CEL expression language.
type ValidationRules []ValidationRule

// ValidationRule describes a validation rule written in the CEL expression language.
type ValidationRule struct {
	// Rule represents the expression which will be evaluated by CEL.
	// ref: https://github.com/google/cel-spec
	// The Rule is scoped to the location of the x-kubernetes-validations extension in the schema.
	// The `self` variable in the CEL expression is bound to the scoped value.
	// Example:
	// - Rule scoped to the root of a resource with a status subresource: {"rule": "self.status.actual <= self.spec.maxDesired"}
	//
	// If the Rule is scoped to an object with properties, the accessible properties of the object are field selectable
	// via `self.field` and field presence can be checked via `has(self.field)`. Null valued fields are treated as
	// absent fields in CEL expressions.
	// If the Rule is scoped to an object with additionalProperties (i.e. a map) the value of the map
	// are accessible via `self[mapKey]`, map containment can be checked via `mapKey in self` and all entries of the map
	// are accessible via CEL macros and functions such as `self.all(...)`.
	// If the Rule is scoped to an array, the elements of the array are accessible via `self[i]` and also by macros and
	// functions.
	// If the Rule is scoped to a scalar, `self` is bound to the scalar value.
	// Examples:
	// - Rule scoped to a map of objects: {"rule": "self.components['Widget'].priority < 10"}
	// - Rule scoped to a list of integers: {"rule": "self.values.all(value, value >= 0 && value < 100)"}
	// - Rule scoped to a string value: {"rule": "self.startsWith('kube')"}
	//
	// The `apiVersion`, `kind`, `metadata.name` and `metadata.generateName` are always accessible from the root of the
	// object and from any x-kubernetes-embedded-resource annotated objects. No other metadata properties are accessible.
	//
	// Unknown data preserved in custom resources via x-kubernetes-preserve-unknown-fields is not accessible in CEL
	// expressions. This includes:
	// - Unknown field values that are preserved by object schemas with x-kubernetes-preserve-unknown-fields.
	// - Object properties where the property schema is of an "unknown type". An "unknown type" is recursively defined as:
	//   - A schema with no type and x-kubernetes-preserve-unknown-fields set to true
	//   - An array where the items schema is of an "unknown type"
	//   - An object where the additionalProperties schema is of an "unknown type"
	//
	// Only property names of the form `[a-zA-Z_.-/][a-zA-Z0-9_.-/]*` are accessible.
	// Accessible property names are escaped according to the following rules when accessed in the expression:
	// - '__' escapes to '__underscores__'
	// - '.' escapes to '__dot__'
	// - '-' escapes to '__dash__'
	// - '/' escapes to '__slash__'
	// - Property names that exactly match a CEL RESERVED keyword escape to '__{keyword}__'. The keywords are:
	//	  "true", "false", "null", "in", "as", "break", "const", "continue", "else", "for", "function", "if",
	//	  "import", "let", "loop", "package", "namespace", "return".
	// Examples:
	//   - Rule accessing a property named "namespace": {"rule": "self.__namespace__ > 0"}
	//   - Rule accessing a property named "x-prop": {"rule": "self.x__dash__prop > 0"}
	//   - Rule accessing a property named "redact__d": {"rule": "self.redact__underscores__d > 0"}
	//
	// Equality on arrays with x-kubernetes-list-type of 'set' or 'map' ignores element order, i.e. [1, 2] == [2, 1].
	// Concatenation on arrays with x-kubernetes-list-type use the semantics of the list type:
	//   - 'set': `X + Y` performs a union where the array positions of all elements in `X` are preserved and
	//     non-intersecting elements in `Y` are appended, retaining their partial order.
	//   - 'map': `X + Y` performs a merge where the array positions of all keys in `X` are preserved but the values
	//     are overwritten by values in `Y` when the key sets of `X` and `Y` intersect. Elements in `Y` with
	//     non-intersecting keys are appended, retaining their partial order.
	Rule string `json:"rule" protobuf:"bytes,1,opt,name=rule"`
	// Message represents the message displayed when validation fails. The message is required if the Rule contains
	// line breaks. The message must not contain line breaks.
	// If unset, the message is "failed rule: {Rule}".
	// e.g. "must be a URL with the host matching spec.host"
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
}

// JSON represents any valid JSON value.
// These types are supported: bool, int64, float64, string, []interface{}, map[string]interface{} and nil.
type JSON struct {
	Raw []byte `protobuf:"bytes,1,opt,name=raw"`
}

// JSONSchemaURL represents a schema url.
type JSONSchemaURL string

// JSONSchemaPropsOrArray represents a value that can either be a JSONSchemaProps
// or an array of JSONSchemaProps. Mainly here for serialization purposes.
type JSONSchemaPropsOrArray struct {
	Schema      *JSONSchemaProps  `protobuf:"bytes,1,opt,name=schema"`
	JSONSchemas []JSONSchemaProps `protobuf:"bytes,2,rep,name=jSONSchemas"`
}

// JSONSchemaPropsOrBool represents JSONSchemaProps or a boolean value.
// Defaults to true for the boolean property.
type JSONSchemaPropsOrBool struct {
	Allows bool             `protobuf:"varint,1,opt,name=allows"`
	Schema *JSONSchemaProps `protobuf:"bytes,2,opt,name=schema"`
}

// JSONSchemaDependencies represent a dependencies property.
type JSONSchemaDependencies map[string]JSONSchemaPropsOrStringArray

// JSONSchemaPropsOrStringArray represents a JSONSchemaProps or a string array.
type JSONSchemaPropsOrStringArray struct {
	Schema   *JSONSchemaProps `protobuf:"bytes,1,opt,name=schema"`
	Property []string         `protobuf:"bytes,2,rep,name=property"`
}


// JSONSchemaDefinitions contains the models explicitly defined in this spec.
type JSONSchemaDefinitions map[string]JSONSchemaProps

// ExternalDocumentation allows referencing an external resource for extended documentation.
type ExternalDocumentation struct {
	Description string `json:"description,omitempty" protobuf:"bytes,1,opt,name=description"`
	URL         string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
}