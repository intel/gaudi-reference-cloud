// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package logkeys

/*
For constants, the convention in Go is to use MixedCaps or mixedCaps rather than underscores to write multiword names.
We follow MixedCaps here since all keys are exported.
Ref: https://go.dev/doc/effective_go#formatting

*/

const (

	//Request, Response, Others
	Request            = "req"
	Response           = "resp"
	Error              = "err"
	ErrorType          = "errType"
	ResponseStatusCode = "respStatusCode"
	Issues             = "issues"
	Input              = "input"
	Result             = "result"
	Message            = "message"
	Reason             = "reason"
	Success            = "success"
	Timeout            = "timeout"
	Prefix             = "prefix"
	Output             = "output"
	LastUpdatedTime    = "lastUpdatedTime"
	LastSuccessAge     = "lastSuccessAge"
	Status             = "status"
	StatusPhase        = "statusPhase"
	StatusMessage      = "statusMessage"
	StatusConditions   = "statusConditions"
	OldStatus          = "oldStatus"
	NewStatus          = "newStatus"
	State              = "state"
	Versions           = "versions"
	UserName           = "userName"
	Name               = "name"
	Version            = "version"
	Namespace          = "namespace"
	NamespaceList      = "namespaceList"
	ClientId           = "clientId"
	Size               = "size"
	RequestSpec        = "requestSpec"
	Resource           = "resource"
	Usage              = "usage"
	Ratio              = "ratio"
	NameOrId           = "nameOrId"
	ProvisioningState  = "provisioningState"
	AvailabilityZone   = "zone"
	Label              = "label"
	Labels             = "labels"
	LabelsToBeDeleted  = "labelsToBeDeleted"
	UpdatedLabels      = "updatedLabels"
	Endpoint           = "endpoint"
	StoragePath        = "storagePath"
	ViewPolicy         = "viewPolicy"

	// Configuration
	ConfigFile    = "configFile"
	Configuration = "config"

	// Server, Controller
	ListenAddr        = "listenAddr"
	ListenPort        = "listenPort"
	ServerAddr        = "serverAddr"
	Controller        = "controller"
	ControllerKind    = "controllerKind"
	ControllerGroup   = "controllerGroup"
	Address           = "address"
	EnvDetails        = "envDetails"
	AddressConsumerId = "addressConsumerId"

	// SSH, Auth
	CognitoToken           = "cognitoToken"
	Token                  = "token"
	TokenTTL               = "tokenTTL"
	TTL                    = "ttl"
	RenewalTime            = "renewalTime"
	TimeSinceRenewalTime   = "timeSinceRenewalTime"
	SSHKeyName             = "sshKeyName"
	CommonName             = "commonName"
	Subject                = "subject"
	DNSNames               = "dnsNames"
	Secret                 = "secret"
	SecretsPath            = "secretsPath"
	KeyPath                = "keyPath"
	SecretPathNamespace    = "secretPathNamespace"
	PrincipalId            = "principalId"
	NsCredsPath            = "nsCredsPath"
	UserCredsPath          = "userCredsPath"
	SshPath                = "sshPath"
	HomePath               = "homePath"
	PublicKeys             = "publicKeys"
	HostKey                = "hostKey"
	SshProxyTunnelStatus   = "sshProxyTunnelStatus"
	SshProxyTunnelInstance = "sshProxyTunnelInstance"
	UniquePublicKeyCount   = "uniquePublicKeyCount"
	BmcDataSecretName      = "bmcDataSecretName"

	// Cloud Account, User
	CloudAccountId   = "cloudAccountId"
	CloudAccount     = "cloudAccount"
	CloudAccountType = "cloudAccountType"
	UserData         = "userData"
	SubscriptionName = "subscriptionName"

	// Resource, Instance, VM/BM Hosts
	ResourceId                = "resourceId"
	ResourceJson              = "resourceJson"
	ResourceName              = "resourceName"
	ResourceVersion           = "resourceVersion"
	ResourceCount             = "resourceCount"
	InstanceType              = "instanceType"
	Instance                  = "instance"
	InstanceName              = "instanceName"
	InstanceNamespace         = "instanceNamespace"
	InstanceCondition         = "instanceCondition"
	InstanceConditionType     = "instanceConditionType"
	InstanceMetadata          = "instanceMetadata"
	InstanceSpec              = "instanceSpec"
	InstanceStatus            = "instanceStatus"
	InstanceNodeId            = "instanceNodeId"
	InstancePrivateMetadata   = "instancePrivateMetadata"
	InstancePrivateSpec       = "instancePrivateSpec"
	InstanceIndex             = "instanceIndex"
	InstanceCount             = "instanceCount"
	RequestedInstanceCount    = "requestedInstanceCount"
	CurrentInstanceCount      = "currentInstanceCount"
	CurrentInstanceFinalizers = "currentInstanceFinalizers"
	LatestInstanceFinalizers  = "latestInstanceFinalizers"
	HostName                  = "hostName"
	HostNamespace             = "hostNamespace"
	Hosts                     = "hosts"
	DestinationHostIp         = "destinationHostIp"
	GpuCount                  = "gpuCount"
	GpuModelName              = "gpuModelName"
	TargetInstance            = "targetInstance"
	DeviceName                = "deviceName"
	DeviceId                  = "deviceId"
	DeviceHealth              = "deviceHealth"
	PvcName                   = "pvcName"
	VirtualMachine            = "virtualMachine"
	VmPrintableStatus         = "vmPrintableStatus"
	BmcType                   = "bmcType"
	DataDisksList             = "dataDisksList"
	RootDeviceHint            = "rootDeviceHint"
	RootDisksList             = "rootDisksList"
	PCIAddress                = "pciAddress"
	NumOfPCIDevices           = "numOfPCIDevices"
	Event                     = "event"
	RetryInterval             = "retryInterval"
	IOMMULink                 = "iommuLink"
	DriverLink                = "driverLink"
	NumaNodePath              = "numaNodePath"
	NumaNodeString            = "numaNodeString"
	MachineImage              = "machineImage"

	// Deactivated instances cleanup
	NumOfInstancesDeleted            = "numOfInstancesDeleted"
	NumOfDeactivatedInstancesDeleted = "numOfDeactivatedInstancesDeleted"
	DeactivationListSize             = "deactivationListSize"

	// instance quota limits and validation
	AllowedQuota                   = "allowedQuota"
	TotalQuota                     = "totalQuota"
	InputInstance                  = "inputInstance"
	AccountCacheAtReturnForBuckets = "accountCacheAtReturnForBuckets"
	QuotaBuckets                   = "quotaBuckets"
	QuotaFilesystem                = "quotaFilesystem"
	QuotaSize                      = "quotaSize"
	QuotaCache                     = "quotaCache"
	QuotaCacheTtl                  = "quotaCacheTtl"
	AccountCacheAtReturn           = "accountCacheAtReturn"
	ValidationId                   = "validationId"
	ExpectedValidationId           = "expectedValidationId"
	ObservedValidationId           = "observedValidationId"
	ValidationStatus               = "validationStatus"
	ValidationS3Bucket             = "validationS3Bucket"
	ValidationS3URL                = "validationS3URL"
	ValidatorIp                    = "validatorIp"
	ValidationResult               = "validationResult"
	ValidationTaskExitCode         = "validationTaskExitCode"
	ValidationResultEntrycount     = "validationResultEntrycount"

	// Instance group
	InstanceGroupSize        = "instanceGroupSize"
	InstanceGroupName        = "instanceGroupName"
	DesiredInstanceGroupSize = "desiredInstanceGroupSize"
	AssignedGroupIds         = "assignedGroupIds"

	// Network
	VNet                  = "vNet"
	VNetName              = "vNetName"
	VNetMetadata          = "vNetMetadata"
	Interfaces            = "interfaces"
	ComputeNodePools      = "computeNodePools"
	Subnet                = "subnet"
	SubnetId              = "subnetId"
	SubnetName            = "subnetName"
	SubnetResp            = "subnetResp"
	NumSubnetEvents       = "numSubnetEvents"
	MaximumPrefixLength   = "maximumPrefixLength"
	SpecPrefixLength      = "specPrefixLength"
	InsertedAddresses     = "insertedAddresses"
	DeletedAddresses      = "deletedAddresses"
	NetworkMode           = "networkMode"
	NetworkDataSerialized = "networkDataSerialized"
	NetworkBackendType    = "networkBackendType"
	NetworkNodeName       = "networkNodeName"

	// Load Balancer
	MaxResourceVersionAtStart = "maxResourceVersionAtStart"
	LoadBalancer              = "loadBalancer"
	LoadBalancerName          = "loadBalancerName"
	LoadBalancerNamespace     = "loadBalancerNamespace"
	LoadBalancerStatus        = "loadBalancerStatus"
	LoadBalancerSpec          = "loadBalancerSpec"
	LoadBalancerVIP           = "vip"
	LatestLoadbalancerStatus  = "latestLoadBalancerStatus"
	CurrentLoadBalancerCount  = "currentLoadBalancerCount"
	Target                    = "target"
	TargetUrl                 = "targetUrl"
	TargetUrlUserName         = "targetUrlUserName"
	TargetUrlHost             = "targetUrlHost"
	TargetUrlPath             = "targetUrlPath"
	TargetLoadBalancer        = "targetLoadBalancer"
	TargetLoadBalancerSpec    = "targetLoadBalancerSpec"
	Tenant                    = "tenant"
	Provider                  = "provider"
	ProviderType              = "providerType"
	ProviderURL               = "providerURL"
	NumOfLoadBalancers        = "numOfLoadBalancers"
	DeletionTimestamp         = "deletionTimestamp"
	NumOfFirewallRules        = "numOfFWRules"
	FirewallRuleName          = "fwRuleName"
	FirewallRuleNamespace     = "fwRuleNamespace"
	Pool                      = "pool"
	PoolId                    = "poolId"
	PoolURL                   = "poolURL"
	NewPool                   = "newPool"
	VirtualServer             = "virtualServer"
	VirtualServerName         = "virtualServerName"
	Members                   = "members"
	Member                    = "member"
	MemberId                  = "memberId"
	MemberName                = "memberName"
	MemberPeerURL             = "memberPeerURL"
	Duration                  = "duration"
	Expiry                    = "expiry"
	IlbName                   = "ilbName"
	Options                   = "options"

	// NaaS
	VPC                 = "vpc"
	SUBNET              = "subnet"
	IPRM                = "iprm"
	PORT                = "port"
	RESOURCE            = "resource"
	ADDRESS_TRANSLATION = "addressTranslation"

	// Framework
	Extension = "extension"
	Plugin    = "plugin"

	// DB
	DbMaxIdleConnectionCount = "dbMaxIdleConnectionCount"
	Query                    = "query"
	Args                     = "args"
	RecordCount              = "recordCount"
	Total                    = "total"
	TotalCount               = "totalCount"
	NumAffectedRows          = "numAffectedRows"

	// Metering
	MeteringRecord = "meteringRecord"

	//UUID
	UUIDType      = "uuidType"
	GeneratedUuid = "generatedUuid"

	// IKS, Kfaas, kubernetes, Control Plane
	Attempt                    = "attempt"
	Function                   = "function"
	ChangeApplied              = "changeApplied"
	ChartName                  = "chartName"
	RepoURL                    = "repoURL"
	ChartVersionURL            = "chartVersionURL"
	FileName                   = "fileName"
	Release                    = "release"
	AddonsURL                  = "addonsURL"
	AddonState                 = "addonState"
	Artifact                   = "artifact"
	DesiredArtifact            = "desiredArtifact"
	CurrentArtifact            = "currentArtifact"
	ControlplaneNodegroupState = "cpNodegroupState"
	CurrentInstanceIMI         = "currentInstanceIMI"
	DesiredInstanceIMI         = "desiredInstanceIMI"
	EtcdIlbIp                  = "etcdIlbIp"
	EtcdIlbPort                = "etcdIlbPort"
	EtcdSnapshotsPath          = "etcdSnapshotsPath"
	UploadInfoSize             = "uploadInfoSize"
	TimeInUpdatingStateSeconds = "timeInUpdatingStateSeconds"
	AutoRepairDisabledLabelKey = "autoRepairDisabledLabelKey"
	AutoRepairValue            = "autoRepairValue"
	KubeflowDeploymentId       = "kubeflowDeploymentId"
	NumCores                   = "numCores"
	Mode                       = "mode"
	IksCount                   = "iksCount"
	PatchHelper                = "patchHelper"
	StorageEnabled             = "storageEnabled"

	// Cluster
	ClusterName           = "clusterName"
	ClusterNamespace      = "clusterNamespace"
	Cluster               = "cluster"
	Clusters              = "clusters"
	ClusterId             = "clusterId"
	ClusterStatus         = "clusterStatus"
	ClusterType           = "clusterType"
	ClusterSpaceTotal     = "clusterSpaceTotal"
	ClusterSpaceAvailable = "clusterSpaceAvailable"
	ClusterEndpoint       = "clusterEndpoint"
	ClusterRevId          = "clusterRevId"
	ClusterGroup          = "clusterGroup"
	CurrentClusterState   = "currentClusterState"

	// Node
	Node                     = "node"
	NodeName                 = "nodeName"
	NodeState                = "nodeState"
	NodeStatus               = "nodeStatus"
	NodeIp                   = "nodeIp"
	NodeGroupName            = "nodeGroupName"
	NodeGroupId              = "nodeGroupId"
	NodeGroupStatus          = "nodeGroupStatus"
	NodeGroupResourceVersion = "nodeGroupResourceVersion"
	NodegroupGeneration      = "nodeGroupGeneration"
	NodeScore                = "nodeScore"
	CurrentNodeCount         = "currentNodeCount"
	DesiredNodeCount         = "desiredNodeCount"
	EvaluatedNodes           = "evaluatedNodes"
	FeasibleNodes            = "feasibleNodes"

	// Vip
	VipName            = "vipName"
	VipId              = "vipId"
	VipURL             = "vipURL"
	EtcdVIP            = "etcdVIP"
	ApiServerVIP       = "apiServerVIP"
	PublicApiServerVIP = "publicApiServerVIP"
	KonnectivityVIP    = "konnectivityVIP"

	// Scheduler resources
	Pod                      = "pod"
	ResourceToWeightMap      = "resourceToWeightMap"
	ResourceAllocationScorer = "resourceAllocationScorer"
	AllocatableResource      = "allocatableResource"
	RequestedResource        = "requestedResource"
	ResourceScore            = "resourceScore"
	TopologyKey              = "topologyKey"
	ProfileName              = "profileName"
	SchedulerName            = "schedulerName"
	Quantity                 = "quantity"

	// Storage, Filesystem, Weka
	AccessPolicy               = "policy"
	Storage                    = "storage"
	StorageClass               = "storageClass"
	StorageProvider            = "storageProvider"
	StorageSize                = "storageSize"
	StorageNamespaceState      = "storageNamespaceState"
	StorageState               = "storageState"
	TimeInSecondsSinceCreation = "timeInSecondsSinceCreation"
	TimeInSecondsSinceUpdate   = "timeInSecondsSinceUpdate"
	FsOrgName                  = "fsOrgName"
	CloudAccountQuota          = "cloudAccountQuota"
	BucketId                   = "bucketId"
	BucketName                 = "bucketName"
	BucketCount                = "bucketCount"
	BucketCapacity             = "bucketCapacity"
	BucketSpec                 = "bucketSpec"
	BucketMetadata             = "bucketMetadata"
	BucketUserId               = "bucketUserId"
	BucketUsage                = "bucketUsage"
	BucketInfo                 = "bucketInfo"
	Task                       = "task"
	TaskStatus                 = "taskStatus"
	TaskMeta                   = "taskMeta"
	Object                     = "object"
	ObjectName                 = "objectName"
	ObjectStatus               = "objectStatus"
	OldObject                  = "OldObject"
	NewObject                  = "newObject"
	LatestObjectStatus         = "latestObjectStatus"
	CurrentObjectFinalizer     = "currentObjectFinalizer"
	LatestObjectFinalizer      = "latestObjectFinalizer"
	CurrentStorageFinalizer    = "currentObjectFinalizer"
	LatestStorageFinalizer     = "latestObjectFinalizer"
	RoleId                     = "roleId"
	VaultSecretEngineStorage   = "vaultSecretEngineStorage"
	DefaultSecretEngineStorage = "defaultSecretEngineStorage"
	CurrentStorageStatus       = "currentStorageStatus"
	LatestStorageStatus        = "latestStorageStatus"
	Filesystem                 = "filesystem"
	FilesystemList             = "filesystemList"
	FilesystemName             = "filesystemName"
	FilesystemMetadata         = "filesystemMetadata"
	FilesystemCapacity         = "filesystemCapacity"
	FilesystemCount            = "filesystemCount"
	TotalDeletedQuota          = "totalDeletedQuota"
	FilesystemTypeFilter       = "filesystemTypeFilter"
	NetSize                    = "netSize"
	ExistingSize               = "existingSize"
	AdditionalSize             = "additionalSize"
	FinalQuota                 = "finalQuota"
	TotalEvents                = "totalEvents"
	EventType                  = "eventType"
	Filters                    = "filters"
	FSScheduler                = "fsScheduler"
	RuleId                     = "ruleId"
)
