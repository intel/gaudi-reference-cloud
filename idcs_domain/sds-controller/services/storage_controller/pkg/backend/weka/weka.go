// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/helpers"
	"github.com/rs/zerolog/log"
)

type Backend struct {
	config           *conf.Cluster
	client           v4.ClientWithResponsesInterface
	adminCredentials backend.AuthCreds
	lock             sync.Mutex
	lrLocks          map[string]*sync.Mutex
	orgNames         map[string]string // Used for cache org names by id
}

type Interface interface {
	CreateFilesystem(ctx context.Context, opts CreateFilesystemOpts) (*Filesystem, error)
	DeleteFilesystem(ctx context.Context, opts DeleteFilesystemOpts) error
	GetFilesystem(ctx context.Context, opts GetFilesystemOpts) (*Filesystem, error)
	ListFilesystems(ctx context.Context, opts ListFilesystemsOpts) ([]*Filesystem, error)
	UpdateFilesystem(ctx context.Context, opts UpdateFilesystemOpts) (*Filesystem, error)
}

type CreateFilesystemOpts struct {
	NamespaceID  string
	Name         string
	Quota        uint64
	Encrypted    bool
	AuthRequired bool
	AuthCreds    *backend.AuthCreds
}

type DeleteFilesystemOpts struct {
	NamespaceID  string
	FilesystemID string
	AuthCreds    *backend.AuthCreds
}

type GetFilesystemOpts struct {
	NamespaceID  string
	FilesystemID string
	AuthCreds    *backend.AuthCreds
}

type ListFilesystemsOpts struct {
	NamespaceID string
	Names       []string
	AuthCreds   *backend.AuthCreds
}

type UpdateFilesystemOpts struct {
	NamespaceID  string
	FilesystemID string
	Quota        *uint64
	Name         *string
	AuthRequired *bool
	AuthCreds    *backend.AuthCreds
}

type Filesystem struct {
	ID               string
	Name             string
	Encrypted        bool
	AuthRequired     bool
	FilesystemStatus FilesystemStatus
	AvailableBytes   uint64
	TotalBytes       uint64
	BackendFQDN      string
}

type FilesystemStatus int

const (
	Unknown  FilesystemStatus = 0
	Creating FilesystemStatus = 1
	Ready    FilesystemStatus = 2
	Removing FilesystemStatus = 3
)

func NewBackend(config *conf.Cluster) (*Backend, error) {
	if config.WekaConfig == nil || config.WekaConfig.TenantFsGroupName == "" {
		return nil, fmt.Errorf("weka config should contain tenantFsGroupName value for cluster: %s", config.UUID)
	}

	if config.WekaConfig.BackendFQDN == "" {
		return nil, fmt.Errorf("weka config should contain backendFqdn value for cluster: %s", config.UUID)
	}

	if config.Auth == nil {
		return nil, fmt.Errorf("auth field in the config cannot be nil for cluster: %s", config.UUID)
	}

	if config.Auth.Scheme != conf.Basic {
		return nil, fmt.Errorf("weka support only basic credentials, provided: %v", config.Auth.Scheme)
	}

	if config.WekaConfig.FileSystemDeleteWait == 0 {
		log.Warn().Msg("Zero fileSystemDeleteWait value for weka backend, defaulting to 5")
		config.WekaConfig.FileSystemDeleteWait = 5
	}

	creds, err := conf.ReadCredentials(*config.Auth)
	if err != nil {
		return nil, err
	}

	tr, err := backend.CreateHTTPTransport(config, true)
	if err != nil {
		return nil, err
	}

	client, err := v4.NewClientWithResponses(config.API.URL, func(c *v4.Client) error {
		c.Client = &http.Client{
			Transport: tr,
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not create weka client: %v", err)
	}

	return &Backend{
		config: config,
		client: client,
		adminCredentials: backend.AuthCreds{
			Principal:   creds.Principal,
			Credentials: creds.Credentials,
		},
		orgNames: make(map[string]string),
		lrLocks:  make(map[string]*sync.Mutex),
	}, nil
}

func (b *Backend) login(ctx context.Context, creds backend.AuthCreds, org string) (*securityprovider.SecurityProviderBearerToken, error) {
	resp, err := b.client.LoginWithResponse(ctx, v4.LoginJSONRequestBody{
		Username: creds.Principal,
		Password: creds.Credentials,
		Org:      &org,
	})

	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil || resp.JSON200.Data.AccessToken == nil {
		return nil, fmt.Errorf("login unsuccessful, status: %s, body: %v, org-name: %v, cluster-id: %v", strconv.Itoa(resp.StatusCode()), string(resp.Body), org, b.config.UUID)
	}

	return securityprovider.NewSecurityProviderBearerToken(*resp.JSON200.Data.AccessToken)
}

func (b *Backend) loginUser(ctx context.Context, creds backend.AuthCreds, namespaceID string) (*securityprovider.SecurityProviderBearerToken, error) {
	b.lock.Lock()
	orgName, found := b.orgNames[namespaceID]
	b.lock.Unlock()
	if !found {
		// Try to refresh cache
		_, _ = b.ListNamespaces(ctx, backend.ListNamespacesOpts{})
		orgName, found = b.orgNames[namespaceID]
	}

	if !found {
		return nil, fmt.Errorf("could not find org name for id %s", namespaceID)
	}

	return b.login(ctx, creds, orgName)
}

func intoNamespace(org v4.Organization) (*backend.Namespace, error) {
	if org.Uid == nil || org.Name == nil {
		return nil, errors.New("org does not contain uid or name")
	}

	return &backend.Namespace{
		ID:         *org.Uid,
		Name:       *org.Name,
		QuotaTotal: helpers.ValueOrNil(org.TotalQuota),
	}, nil
}

func intoUser(user v4.User) (*backend.User, error) {
	if user.Uid == nil || user.Username == nil {
		return nil, errors.New("user does not contain uid or name")
	}

	var role backend.UserRole

	if user.Role != nil {
		switch *user.Role {
		case "OrgAdmin":
			role = backend.Admin
		case "Regular":
			role = backend.Regular
		}
	}

	return &backend.User{
		ID:   *user.Uid,
		Name: *user.Username,
		Role: role,
	}, nil
}

func intoFilesystem(fs v4.FileSystem, backendFqdn string) (*Filesystem, error) {
	if fs.Uid == nil || fs.Name == nil {
		return nil, errors.New("filesystem does not contain uid or name")
	}

	var status FilesystemStatus

	if fs.Status != nil {
		switch *fs.Status {
		case "CREATING":
			status = Creating
		case "READY":
			status = Ready
		case "REMOVING":
			status = Removing
		default:
			status = Unknown
		}
	}

	return &Filesystem{
		ID:               *fs.Uid,
		Name:             *fs.Name,
		Encrypted:        helpers.ValueOrNil(fs.IsEncrypted),
		AuthRequired:     helpers.ValueOrNil(fs.AuthRequired),
		AvailableBytes:   helpers.ValueOrNil(fs.AvailableTotal),
		TotalBytes:       helpers.ValueOrNil(fs.TotalBudget),
		FilesystemStatus: status,
		BackendFQDN:      backendFqdn,
	}, nil
}

func intoS3Principal(user v4.User) (*backend.S3Principal, error) {
	if user.Uid == nil || user.Username == nil {
		return nil, errors.New("user does not contain uid or name")
	}

	return &backend.S3Principal{
		ID:   *user.Uid,
		Name: *user.Username,
	}, nil
}

func intoStatefulClient(sc v4.Container) (*backend.StatefulClient, error) {
	if sc.Uid == nil {
		return nil, errors.New("container does not contain required \"Uid\" fields")
	}

	if sc.Hostname == nil {
		return nil, errors.New("container does not contain required \"Hostname\" fields")
	}

	if sc.Status == nil {
		return nil, errors.New("container does not contain required \"status\" fields")
	}

	if sc.Mode == nil {
		return nil, errors.New("container does not contain required \"mode\" fields")
	}

	if sc.Cores == nil {
		return nil, errors.New("container does not contain required \"cores\" fields")
	}

	return &backend.StatefulClient{
		ID:     *sc.Uid,
		Name:   *sc.Hostname,
		Status: *sc.Status,
		Mode:   *sc.Mode,
		Cores:  *sc.Cores,
	}, nil
}

func intoStatefulClientProcess(sc v4.Process) (*backend.Process, error) {
	if sc.Uid == nil {
		return nil, errors.New("container process does not contain required \"Uid\" fields")
	}

	if sc.Hostname == nil {
		return nil, errors.New("container process does not contain required \"Hostname\" fields")
	}

	if sc.Status == nil {
		return nil, errors.New("container process  does not contain required \"status\" fields")
	}

	if sc.Mode == nil {
		return nil, errors.New("container process  does not contain required \"mode\" fields")
	}

	if sc.Roles == nil {
		return nil, errors.New("container process  does not contain required \"roles\" fields")
	}

	roles := *sc.Roles
	if len(roles) != 1 {
		return nil, errors.New("statefulclient container process does not have more than 1 role")
	}

	return &backend.Process{
		ID:       *sc.Uid,
		Hostname: *sc.Hostname,
		Status:   *sc.Status,
		Role:     roles[0],
		Mode:     *sc.Mode,
	}, nil
}
