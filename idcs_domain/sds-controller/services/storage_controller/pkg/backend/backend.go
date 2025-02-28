// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package backend

import (
	"context"
	"fmt"
	"net"
	"strconv"
)

type Backend struct {
	Interface
}

type HealthInterface interface {
	GetStatus(ctx context.Context) (*ClusterStatus, error)
}

// This is abstraction layer between general backend interface and implementation
// This functions can be later factored down into separate services and controlled by CRD
type Interface interface {
	GetStatus(ctx context.Context) (*ClusterStatus, error)
}

type S3Ops interface {
	CreateBucket(ctx context.Context, opts CreateBucketOpts) (*Bucket, error)
	DeleteBucket(ctx context.Context, opts DeleteBucketOpts) error
	GetBucketPolicy(ctx context.Context, opts GetBucketPolicyOpts) (*AccessPolicy, error)
	ListBuckets(ctx context.Context, opts ListBucketsOpts) ([]*Bucket, error)
	UpdateBucketPolicy(ctx context.Context, opts UpdateBucketPolicyOpts) error

	CreateLifecycleRules(ctx context.Context, opts CreateLifecycleRulesOpts) ([]*LifecycleRule, error)
	DeleteLifecycleRules(ctx context.Context, opts DeleteLifecycleRulesOpts) error
	ListLifecycleRules(ctx context.Context, opts ListLifecycleRulesOpts) ([]*LifecycleRule, error)
	UpdateLifecycleRules(ctx context.Context, opts UpdateLifecycleRulesOpts) ([]*LifecycleRule, error)

	CreateS3Principal(ctx context.Context, opts CreateS3PrincipalOpts) (*S3Principal, error)
	DeleteS3Principal(ctx context.Context, opts DeleteS3PrincipalOpts) error
	GetS3Principal(ctx context.Context, opts GetS3PrincipalOpts) (*S3Principal, error)
	UpdateS3PrincipalPassword(ctx context.Context, opts UpdateS3PrincipalPasswordOpts) error
	UpdateS3PrincipalPolicies(ctx context.Context, opts UpdateS3PrincipalPoliciesOpts) (*S3Principal, error)
}

type NamespaceOps interface {
	CreateNamespace(ctx context.Context, opts CreateNamespaceOpts) (*Namespace, error)
	DeleteNamespace(ctx context.Context, opts DeleteNamespaceOpts) error
	GetNamespace(ctx context.Context, opts GetNamespaceOpts) (*Namespace, error)
	ListNamespaces(ctx context.Context, opts ListNamespacesOpts) ([]*Namespace, error)
	UpdateNamespace(ctx context.Context, opts UpdateNamespaceOpts) (*Namespace, error)
}

type StatefulClientOps interface {
	CreateStatefulClient(ctx context.Context, opts CreateStatefulClientOpts) (*StatefulClient, error)
	DeleteStatefulClient(ctx context.Context, opts DeleteStatefulClientOpts) error
	GetStatefulClient(ctx context.Context, opts GetStatefulClientOpts) (*StatefulClient, error)
	ListStatefulClients(ctx context.Context, opts ListStatefulClientsOpts) ([]*StatefulClient, error)
}

type UserOps interface {
	CreateUser(ctx context.Context, opts CreateUserOpts) (*User, error)
	DeleteUser(ctx context.Context, opts DeleteUserOpts) error
	GetUser(ctx context.Context, opts GetUserOpts) (*User, error)
	ListUsers(ctx context.Context, opts ListUsersOpts) ([]*User, error)
	UpdateUser(ctx context.Context, opts UpdateUserOpts) (*User, error)
	UpdateUserPassword(ctx context.Context, opts UpdateUserPasswordOpts) error
}

// This types can be reworked into CRD later, this why we not reuse protobuf definitions
type CreateNamespaceOpts struct {
	Name          string
	Quota         uint64
	AdminName     string
	AdminPassword string
	IPRanges      [][]string
}

type DeleteNamespaceOpts struct {
	NamespaceID string
}

type GetNamespaceOpts struct {
	NamespaceID string
}

type ListNamespacesOpts struct {
	Names []string
}

type UpdateNamespaceOpts struct {
	NamespaceID string
	Quota       uint64
	IPRanges    [][]string
}

type CreateUserOpts struct {
	NamespaceID string
	Name        string
	Password    string
	Role        UserRole
	AuthCreds   *AuthCreds
}

type DeleteUserOpts struct {
	NamespaceID string
	UserID      string
	AuthCreds   *AuthCreds
}

type GetUserOpts struct {
	NamespaceID string
	UserID      string
	AuthCreds   *AuthCreds
}

type ListUsersOpts struct {
	NamespaceID string
	Names       []string
	AuthCreds   *AuthCreds
}

type UpdateUserOpts struct {
	NamespaceID string
	UserID      string
	Role        UserRole
	AuthCreds   *AuthCreds
}

type UpdateUserPasswordOpts struct {
	NamespaceID string
	UserID      string
	Password    string
	AuthCreds   *AuthCreds
}

type CreateBucketOpts struct {
	Name         string
	AccessPolicy AccessPolicy
	Versioned    bool
	QuotaBytes   uint64
}

type DeleteBucketOpts struct {
	ID          string
	ForceDelete bool
}

type GetBucketPolicyOpts struct {
	ID string
}

type ListBucketsOpts struct {
	Names []string
}

type UpdateBucketPolicyOpts struct {
	ID           string
	AccessPolicy AccessPolicy
}

type CreateLifecycleRulesOpts struct {
	BucketID       string
	LifecycleRules []LifecycleRule
}

type DeleteLifecycleRulesOpts struct {
	BucketID string
}

type ListLifecycleRulesOpts struct {
	BucketID string
}

type UpdateLifecycleRulesOpts struct {
	BucketID       string
	LifecycleRules []LifecycleRule
}

type CreateS3PrincipalOpts struct {
	Name        string
	Credentials string
}

type DeleteS3PrincipalOpts struct {
	PrincipalID string
}

type GetS3PrincipalOpts struct {
	PrincipalID string
}

type UpdateS3PrincipalPasswordOpts struct {
	PrincipalID string
	Credentials string
}

type UpdateS3PrincipalPoliciesOpts struct {
	PrincipalID string
	Policies    []*S3Policy
}

type ClusterStatus struct {
	AvailableBytes      uint64
	TotalBytes          uint64
	NamespacesLimit     int32
	NamespacesAvailable int32
	HealthStatus        HealthStatus
	Labels              map[string]string
}

type HealthStatus int

const (
	Healthy   HealthStatus = 1
	Degraded  HealthStatus = 2
	Unhealthy HealthStatus = 3
)

func (h HealthStatus) String() string {
	switch h {
	case Healthy:
		return "Healty"
	case Degraded:
		return "Degraded"
	case Unhealthy:
		return "Unhealthy"
	}
	return fmt.Sprintf("Unknown(%s)", strconv.Itoa(int(h)))
}

type Namespace struct {
	ID         string
	Name       string
	QuotaTotal uint64
	IPRanges   [][]string
}

type User struct {
	ID   string
	Name string
	Role UserRole
}

type UserRole int

const (
	Admin   UserRole = 1
	Regular UserRole = 2
	CSI     UserRole = 3
)

type AuthCreds struct {
	Scheme      AuthScheme
	Principal   string
	Credentials string
}

type AuthScheme int

const (
	Basic  AuthScheme = 1
	Bearer AuthScheme = 2
)

type Bucket struct {
	ID             string
	Name           string
	AccessPolicy   AccessPolicy
	Versioned      bool
	QuotaBytes     uint64
	AvailableBytes uint64
	EndpointURL    string
}

type AccessPolicy int

const (
	None      AccessPolicy = 1
	Read      AccessPolicy = 2
	ReadWrite AccessPolicy = 3
)

type LifecycleRule struct {
	ID                   string
	Prefix               string
	ExpireDays           uint32
	NoncurrentExpireDays uint32
	DeleteMarker         bool
}

type S3Principal struct {
	ID       string
	Name     string
	Policies []*S3Policy
}

type S3Policy struct {
	BucketID           string
	Prefix             string
	Read               bool
	Write              bool
	Delete             bool
	Actions            []string
	AllowSourceNets    []*net.IPNet
	DisallowSourceNets []*net.IPNet
}

type CreateStatefulClientOpts struct {
	Name string
	Ip   string
}

type DeleteStatefulClientOpts struct {
	StatefulClientID string
}

type GetStatefulClientOpts struct {
	StatefulClientID string
}

type ListStatefulClientsOpts struct {
	Names []string
}

type StatefulClient struct {
	ID string
	// StatefulClient name, arbitrary string
	Name   string
	Status string
	Mode   string
	Cores  int
}

type Process struct {
	ID string
	// StatefulClient name, arbitrary string
	Hostname string
	Status   string
	Role     string
	Mode     string
}

type ContainerMode string
type ContainerStatus string
type ContainerRole string

const (
	ContainerModeClient           ContainerMode   = "client"
	ContainerStatusUP             ContainerStatus = "UP"
	ContainerStatusDegraded       ContainerStatus = "DEGRADED"
	ContainerStatusDown           ContainerStatus = "DOWN"
	ContainerStatusProcessesNotUP ContainerStatus = "PROCESSESNOTUP"
	ContainerRoleFrontend         ContainerRole   = "FRONTEND"
	ContainerRoleManagement       ContainerRole   = "MANAGEMENT"
)

func ResponseAsErr(msg string, status int, body []byte) error {
	return fmt.Errorf("%s, status: %s, body: %v", msg, strconv.Itoa(status), string(body))
}
