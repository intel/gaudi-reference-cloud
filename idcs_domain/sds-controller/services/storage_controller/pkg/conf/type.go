// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package conf

type Config struct {
	ListenPort int      `yaml:"listenPort"`
	GrpcTLS    *GrpcTLS `yaml:"grpcTls,omitempty"`

	Clusters []*Cluster `yaml:"clusters"`
	HealthInterval int  `yaml:"healthInterval,omitempty"`

}

type GrpcTLS struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type Cluster struct {
	Name        string            `yaml:"name"`
	UUID        string            `yaml:"uuid"`
	Type        ClusterType       `yaml:"type"`
	Location    string            `yaml:"location"`
	API         *API              `yaml:"api"`
	Labels      map[string]string `yaml:"labels"`
	Auth        *Auth             `yaml:"auth"`
	WekaConfig  *WekaConfig       `yaml:"wekaConfig"`
	MinioConfig *MinioConfig      `yaml:"minioConfig"`
	VastConfig  *VastConfig       `yaml:"vastConfig"`
	SupportsAPI []SupportsAPI     `yaml:"supportsApi"`
}

type API struct {
	Type       APIType `yaml:"type"`
	URL        string  `yaml:"url"`
	CaCertFile string  `yaml:"caCertFile,omitempty"`
}

type Auth struct {
	Scheme    AuthScheme `yaml:"scheme"`
	File      string     `yaml:"file,omitempty"`
	Env       string     `yaml:"env,omitempty"`
	Secret    string     `yaml:"secret,omitempty"`
	VaultFile string     `yaml:"vaultFile,omitempty"`
}

type ClusterType string

const (
	Weka  ClusterType = "Weka"
	MinIO ClusterType = "MinIO"
	Vast  ClusterType = "Vast"
)

type APIType string

const (
	REST APIType = "REST"
	GRPC APIType = "GRPC"
)

type SupportsAPI string

const (
	WekaFilesystem SupportsAPI = "WekaFilesystem"
	ObjectStore    SupportsAPI = "ObjectStore"
	VastView       SupportsAPI = "VastView"
)

type AuthScheme string

const (
	Basic  AuthScheme = "Basic"
	Bearer AuthScheme = "Bearer"
	Digest AuthScheme = "Digest"
)

type AuthCreds struct {
	Scheme      AuthScheme
	Principal   string
	Credentials string
}

type WekaConfig struct {
	ProtectedOrgIds      []string `yaml:"protectedOrgIds"`
	TenantFsGroupName    string   `yaml:"tenantFsGroupName"`
	FileSystemDeleteWait int      `yaml:"fileSystemDeleteWait"`
	BackendFQDN          string   `yaml:"backendFqdn"`
}

type MinioConfig struct {
	KESKey string `yaml:"kesKey"`
}

type VaultSecret struct {
	Data VaultSecretData `json:"data"`
}

type VaultSecretData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VastConfig struct {
	VipPool         string   `yaml:"vipPool"`
	ProtectedOrgIds []string `yaml:"protectedOrgIds"`
}
