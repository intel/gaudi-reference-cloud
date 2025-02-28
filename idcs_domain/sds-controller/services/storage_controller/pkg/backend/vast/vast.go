// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/helpers"
	"github.com/rs/zerolog/log"
)

type Backend struct {
	config           *conf.Cluster
	client           client.ClientWithResponsesInterface
	adminCredentials backend.AuthCreds
	lock             sync.Mutex
}

type Interface interface {
	CreateView(ctx context.Context, opts CreateViewOpts) (*View, error)
	DeleteView(ctx context.Context, opts DeleteViewOpts) error
	GetView(ctx context.Context, opts GetViewOpts) (*View, error)
	ListViews(ctx context.Context, opts ListViewsOpts) ([]*View, error)
	UpdateView(ctx context.Context, opts UpdateViewOpts) (*View, error)
}

type CreateViewOpts struct {
	NamespaceID string
	Name        string
	Path        string
	Protocols   []Protocol
	Quota       uint64
}

type DeleteViewOpts struct {
	NamespaceID string
	ViewID      string
}

type GetViewOpts struct {
	NamespaceID string
	ViewID      string
}

type ListQuotasOpts struct {
	NamespaceID string
	QuotaID     *string
	Paths       []string
}

type ListViewsOpts struct {
	NamespaceID string
	Names       []string
}

type UpdateViewOpts struct {
	NamespaceID string
	ViewID      string
	Name        *string
	Quota       *uint64
	Protocols   []Protocol
}

type View struct {
	ID             string
	Name           string
	Path           string
	Protocols      []Protocol
	PolicyID       int
	AvailableBytes uint64
	TotalBytes     uint64
}

type Quota struct {
	ID    int
	Name  string
	Path  string
	Quota uint64
}

type Protocol int

const (
	NFSV3 Protocol = 1
	NFSV4 Protocol = 2
	SMB   Protocol = 3
)

func (b *Backend) login(ctx context.Context, creds backend.AuthCreds) (*securityprovider.SecurityProviderBearerToken, error) {
	resp, err := b.client.TokenCreateWithResponse(ctx, client.TokenCreateJSONRequestBody{
		Username: creds.Principal,
		Password: creds.Credentials,
	})

	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Access == nil {
		return nil, fmt.Errorf("login unsuccessful, status: %s, body: %v, cluster-id: %v", strconv.Itoa(resp.StatusCode()), string(resp.Body), b.config.UUID)
	}

	return securityprovider.NewSecurityProviderBearerToken(*resp.JSON200.Access)
}

func NewBackend(config *conf.Cluster) (*Backend, error) {
	if config.VastConfig == nil || config.VastConfig.VipPool == "" {
		return nil, fmt.Errorf("vast config could not be empty: %s", config.UUID)
	}

	if config.Auth == nil {
		return nil, fmt.Errorf("auth field in the config cannot be nil for cluster: %s", config.UUID)
	}

	if config.Auth.Scheme != conf.Basic {
		return nil, fmt.Errorf("vast support only basic credentials, provided: %v", config.Auth.Scheme)
	}

	creds, err := conf.ReadCredentials(*config.Auth)
	if err != nil {
		return nil, err
	}

	tr, err := backend.CreateHTTPTransport(config, true)
	if err != nil {
		return nil, err
	}

	client, err := client.NewClientWithResponses(config.API.URL, func(c *client.Client) error {
		c.Client = &http.Client{
			Transport: tr,
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not create Vast client: %v", err)
	}

	return &Backend{
		config: config,
		client: client,
		adminCredentials: backend.AuthCreds{
			Principal:   creds.Principal,
			Credentials: creds.Credentials,
		},
	}, nil
}

func (b *Backend) listQuotas(ctx context.Context, opts ListQuotasOpts, token *securityprovider.SecurityProviderBearerToken) ([]Quota, error) {
	tenantId, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenantId %s: %w", opts.NamespaceID, err)
	}
	var quotaId *int = nil
	if opts.QuotaID != nil {
		var quotaIdParsed int
		quotaIdParsed, err = strconv.Atoi(*opts.QuotaID)
		if err != nil {
			return nil, fmt.Errorf("invalid quotaId %s: %w", *opts.QuotaID, err)
		}
		quotaId = &quotaIdParsed
	}
	request := &client.QuotasListParams{
		TenantId: &tenantId,
		Id:       quotaId,
	}
	if len(opts.Paths) == 1 {
		request.Path = &opts.Paths[0]
	} else if len(opts.Paths) > 1 {
		allPaths := strings.Join(opts.Paths, ",")
		request.PathIn = &allPaths
	}

	quotas, err := b.client.QuotasListWithResponse(ctx, request, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return nil, err
	} else {
		if quotas.JSON200 == nil {
			optsStr, _ := json.Marshal(opts)
			return nil, backend.ResponseAsErr(fmt.Sprintf("can't find quotas by params %v", string(optsStr)), quotas.StatusCode(), quotas.Body)
		} else {
			result := make([]Quota, 0)
			emptyQuota := uint64(0)
			for _, quota := range *quotas.JSON200 {
				// Set quota to zero
				if quota.HardLimit == nil {
					quota.HardLimit = &emptyQuota
				}
				result = append(result, Quota{
					ID:    *quota.Id,
					Name:  quota.Name,
					Path:  quota.Path,
					Quota: *quota.HardLimit,
				})
			}
			return result, nil
		}
	}
}

func intoNamespace(tenant client.Tenant, quota uint64) (*backend.Namespace, error) {
	if tenant.Id == nil {
		return nil, errors.New("tenant does not contain id")
	}

	var ranges [][]string
	if tenant.ClientIpRanges != nil && len(*tenant.ClientIpRanges) == 1 {
		ranges = *tenant.ClientIpRanges
	}

	return &backend.Namespace{
		ID:         strconv.Itoa(*tenant.Id),
		Name:       tenant.Name,
		IPRanges:   ranges,
		QuotaTotal: quota,
	}, nil
}

func intoView(view client.View, quota uint64) (*View, error) {
	if view.Id == nil {
		return nil, errors.New("view does not contain id")
	}

	return &View{
		ID:         strconv.Itoa(*view.Id),
		Name:       helpers.ValueOrNil(view.Name),
		Path:       view.Path,
		Protocols:  stringsToProtocols(view.Protocols),
		PolicyID:   view.PolicyId,
		TotalBytes: quota,
	}, nil
}

func intoUser(user client.ManagersCreateResponse) (*backend.User, error) {
	if user.JSON201 == nil {
		return nil, errors.New("user does not contain id or name")
	}

	id_str := ""
	if user.JSON201.Id != nil {
		id_str = strconv.Itoa(*user.JSON201.Id)
	}

	return &backend.User{
		ID:   id_str,
		Name: user.JSON201.Username,
		Role: backend.CSI,
	}, nil
}
