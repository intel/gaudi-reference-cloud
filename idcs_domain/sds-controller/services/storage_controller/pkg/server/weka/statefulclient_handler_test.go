// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package weka

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	weka "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/weka"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockStatefulClientBackend struct {
	backend.Interface
}

// MockStatefulClientBackendStatus defines parameters for StatefulClientBackendStatus.
type MockStatefulClientBackendStatus string

// Defines values for MockStatefulClientBackendStatus.
const (
	MockStatefulClientBackendStatusUP       MockStatefulClientBackendStatus = "UP"
	MockStatefulClientBackendStatusDEGRADED MockStatefulClientBackendStatus = "DEGRADED"
	MockStatefulClientBackendStatusDOWN     MockStatefulClientBackendStatus = "DOWN"
	MockStatefulClientBackendStatusCUSTOM   MockStatefulClientBackendStatus = "CUSTOM"
)

// MockStatefulClientBackendID defines parameters for StatefulClientBackendID.
type MockStatefulClientBackendID string

// Defines values for MockStatefulClientBackendID.
const (
	MockStatefulClientBackendIDUP       MockStatefulClientBackendID = "SCIDUP"
	MockStatefulClientBackendIDDEGRADED MockStatefulClientBackendID = "SCIDDEGRADED"
	MockStatefulClientBackendIDDOWN     MockStatefulClientBackendID = "SCIDDOWN"
	MockStatefulClientBackendIDCUSTOM   MockStatefulClientBackendID = "SCIDCUSTOM"
)

// MockStatefulClientBackendName defines parameters for StatefulClientBackendName.
type MockStatefulClientBackendName string

// Defines values for MockStatefulClientBackendName.
const (
	MockStatefulClientBackendNameUP       MockStatefulClientBackendName = "SCNAMEUP"
	MockStatefulClientBackendNameDEGRADED MockStatefulClientBackendName = "SCNAMEDEGRADED"
	MockStatefulClientBackendNameDOWN     MockStatefulClientBackendName = "SCNAMEDOWN"
	MockStatefulClientBackendNameCUSTOM   MockStatefulClientBackendName = "SCNAMECUSTOM"
)

var bScUP = backend.StatefulClient{
	ID:     string(MockStatefulClientBackendIDUP),
	Name:   string(MockStatefulClientBackendNameUP),
	Status: string(MockStatefulClientBackendStatusUP),
}

var bScDegraded = backend.StatefulClient{
	ID:     string(MockStatefulClientBackendIDDEGRADED),
	Name:   string(MockStatefulClientBackendNameDEGRADED),
	Status: string(MockStatefulClientBackendStatusDEGRADED),
}

var bScDown = backend.StatefulClient{
	ID:     string(MockStatefulClientBackendIDDOWN),
	Name:   string(MockStatefulClientBackendNameDOWN),
	Status: string(MockStatefulClientBackendStatusDOWN),
}

var bScCustom = backend.StatefulClient{
	ID:     string(MockStatefulClientBackendIDCUSTOM),
	Name:   string(MockStatefulClientBackendNameCUSTOM),
	Status: string(MockStatefulClientBackendStatusCUSTOM),
}

var scResponseWithUpStatus = weka.StatefulClient{
	Id: &weka.StatefulClientIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bScUP.ID,
	},
	Name:   bScUP.Name,
	Status: &weka.StatefulClient_PredefinedStatus{PredefinedStatus: weka.StatefulClient_STATUS_UP},
}

var scResponseWithDegradedStatus = weka.StatefulClient{
	Id: &weka.StatefulClientIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bScDegraded.ID,
	},
	Name:   bScDegraded.Name,
	Status: &weka.StatefulClient_PredefinedStatus{PredefinedStatus: weka.StatefulClient_STATUS_DEGRADED_UNSPECIFIED},
}

var scResponseWithDownStatus = weka.StatefulClient{
	Id: &weka.StatefulClientIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bScDown.ID,
	},
	Name:   bScDown.Name,
	Status: &weka.StatefulClient_PredefinedStatus{PredefinedStatus: weka.StatefulClient_STATUS_DOWN},
}

var scResponseWithCustomStatus = weka.StatefulClient{
	Id: &weka.StatefulClientIdentifier{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Id: bScCustom.ID,
	},
	Name:   bScCustom.Name,
	Status: &weka.StatefulClient_CustomStatus{CustomStatus: string(MockStatefulClientBackendStatusCUSTOM)},
}

// Make sure it implements StatefulClientOps
var _ backend.StatefulClientOps = &MockStatefulClientBackend{}

type MockNoStatefulClientsBackend struct {
	backend.Interface
}

func (*MockStatefulClientBackend) CreateStatefulClient(ctx context.Context, opts backend.CreateStatefulClientOpts) (*backend.StatefulClient, error) {
	if opts.Name == "error1" {
		return nil, errors.New("something wrong")
	}

	var status string
	var Id string
	switch opts.Name {
	case bScUP.Name:
		status = bScUP.Status
		Id = bScUP.ID
	case bScDegraded.Name:
		status = bScDegraded.Status
		Id = bScDegraded.ID
	case bScDown.Name:
		status = bScDown.Status
		Id = bScDown.ID
	default:
		status = bScCustom.Status
		Id = bScCustom.ID
	}

	return &backend.StatefulClient{
		ID:     Id,
		Name:   opts.Name,
		Status: status,
	}, nil
}

func (*MockStatefulClientBackend) DeleteStatefulClient(ctx context.Context, opts backend.DeleteStatefulClientOpts) error {
	if opts.StatefulClientID == "error2" {
		return errors.New("something wrong")
	}

	return nil
}

func (*MockStatefulClientBackend) GetStatefulClient(ctx context.Context, opts backend.GetStatefulClientOpts) (*backend.StatefulClient, error) {
	if opts.StatefulClientID == "error3" {
		return nil, errors.New("something wrong")
	}

	var status string
	var name string
	switch opts.StatefulClientID {
	case bScUP.ID:
		status = bScUP.Status
		name = bScUP.Name
	case bScDegraded.ID:
		status = string(MockStatefulClientBackendStatusDEGRADED)
		name = string(MockStatefulClientBackendNameDEGRADED)
	case bScDown.ID:
		status = string(MockStatefulClientBackendStatusDOWN)
		name = string(MockStatefulClientBackendNameDOWN)
	default:
		status = string(MockStatefulClientBackendStatusCUSTOM)
		name = string(MockStatefulClientBackendNameCUSTOM)

	}

	return &backend.StatefulClient{
		ID:     opts.StatefulClientID,
		Name:   name,
		Status: status,
	}, nil
}

func (*MockStatefulClientBackend) ListStatefulClients(ctx context.Context, opts backend.ListStatefulClientsOpts) ([]*backend.StatefulClient, error) {
	if slices.Contains(opts.Names, "error4") {
		return nil, errors.New("something wrong")
	}

	scs := make([]*backend.StatefulClient, 0)
	scs = append(scs, &bScUP)
	scs = append(scs, &bScDegraded)
	scs = append(scs, &bScDown)
	scs = append(scs, &bScCustom)
	for i := len(scs) - 1; i >= 0; i-- {
		if len(opts.Names) > 0 && !slices.Contains(opts.Names, scs[i].Name) {
			// If names filtering is applied and the name is not in opts.Names, remove the item from the slice.
			scs = append(scs[:i], scs[i+1:]...)
		}
	}

	return scs, nil
}

func TestStatefulClientHandler_CreateStatefulClient(t *testing.T) {
	createReq := weka.CreateStatefulClientRequest{
		ClusterId: &v1.ClusterIdentifier{
			Uuid: testClusterUuid,
		},
		Name: "sc",
		// IP address range 192.0.2.0/24, also known as TEST-NET-1,
		// is reserved for use in documentation and example code.
		// More info RFC 5737
		Ip: "192.0.2.0",
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.CreateStatefulClientRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.CreateStatefulClientResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "create sc wit UP status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.CreateStatefulClientRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: string(MockStatefulClientBackendNameUP),
					// IP address range 192.0.2.0/24, also known as TEST-NET-1,
					// is reserved for use in documentation and example code.
					// More info RFC 5737
					Ip: "192.0.2.0",
				},
			},
			want: &weka.CreateStatefulClientResponse{
				StatefulClient: &scResponseWithUpStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "create sc wit DEGRADED status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.CreateStatefulClientRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: string(MockStatefulClientBackendNameDEGRADED),
					// IP address range 192.0.2.0/24, also known as TEST-NET-1,
					// is reserved for use in documentation and example code.
					// More info RFC 5737
					Ip: "192.0.2.0",
				},
			},
			want: &weka.CreateStatefulClientResponse{
				StatefulClient: &scResponseWithDegradedStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "create sc wit DOWN status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.CreateStatefulClientRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: string(MockStatefulClientBackendNameDOWN),
					// IP address range 192.0.2.0/24, also known as TEST-NET-1,
					// is reserved for use in documentation and example code.
					// More info RFC 5737
					Ip: "192.0.2.0",
				},
			},
			want: &weka.CreateStatefulClientResponse{
				StatefulClient: &scResponseWithDownStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "create sc wit CUSTOM status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.CreateStatefulClientRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: string(MockStatefulClientBackendNameCUSTOM),
					// IP address range 192.0.2.0/24, also known as TEST-NET-1,
					// is reserved for use in documentation and example code.
					// More info RFC 5737
					Ip: "192.0.2.0",
				},
			},
			want: &weka.CreateStatefulClientResponse{
				StatefulClient: &scResponseWithCustomStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.CreateStatefulClientRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Name: "error1",
					// IP address range 192.0.2.0/24, also known as TEST-NET-1,
					// is reserved for use in documentation and example code.
					// More info RFC 5737
					Ip: "192.0.2.0",
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &createReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.CreateStatefulClientRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoStatefulClientsBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.CreateStatefulClientRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatefulClientHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.CreateStatefulClient(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatefulClientHandler_DeleteStatefulClient(t *testing.T) {
	deleteReqWithoutWait := weka.DeleteStatefulClientRequest{
		StatefulClientId: scResponseWithUpStatus.GetId(),
	}
	deleteReq := weka.DeleteStatefulClientRequest{
		StatefulClientId: scResponseWithUpStatus.GetId(),
	}
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.DeleteStatefulClientRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.DeleteStatefulClientResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "delete sc async",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReqWithoutWait,
			},
			want:      &weka.DeleteStatefulClientResponse{},
			assertion: assert.NoError,
		},
		{
			name: "delete sc",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      &weka.DeleteStatefulClientResponse{},
			assertion: assert.NoError,
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &deleteReq,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.DeleteStatefulClientRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoStatefulClientsBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.DeleteStatefulClientRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatefulClientHandler{
				Backends: tt.fields.Backends,
			}

			got, err := h.DeleteStatefulClient(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatefulClientHandler_GetStatefulClient(t *testing.T) {
	getReqwithUpStatus := weka.GetStatefulClientRequest{
		StatefulClientId: scResponseWithUpStatus.GetId(),
	}

	getReqwithDegradedStatus := weka.GetStatefulClientRequest{
		StatefulClientId: scResponseWithDegradedStatus.GetId(),
	}

	getReqwithDownStatus := weka.GetStatefulClientRequest{
		StatefulClientId: scResponseWithDownStatus.GetId(),
	}

	getReqwithCustomStatus := weka.GetStatefulClientRequest{
		StatefulClientId: scResponseWithCustomStatus.GetId(),
	}

	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.GetStatefulClientRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.GetStatefulClientResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "get sc with UP Status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReqwithUpStatus,
			},
			want: &weka.GetStatefulClientResponse{
				StatefulClient: &scResponseWithUpStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "get sc with DEGRADED Status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReqwithDegradedStatus,
			},
			want: &weka.GetStatefulClientResponse{
				StatefulClient: &scResponseWithDegradedStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "get sc with DOWN Status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReqwithDownStatus,
			},
			want: &weka.GetStatefulClientResponse{
				StatefulClient: &scResponseWithDownStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "get sc with Custom Status",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReqwithCustomStatus,
			},
			want: &weka.GetStatefulClientResponse{
				StatefulClient: &scResponseWithCustomStatus,
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.GetStatefulClientRequest{
					StatefulClientId: &weka.StatefulClientIdentifier{
						ClusterId: &v1.ClusterIdentifier{
							Uuid: testClusterUuid,
						},
						Id: "error3",
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReqwithUpStatus,
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.GetStatefulClientRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoStatefulClientsBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &getReqwithUpStatus,
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatefulClientHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.GetStatefulClient(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatefulClientHandler_ListStatefulClients(t *testing.T) {
	type fields struct {
		Backends map[string]backend.Interface
	}
	type args struct {
		ctx context.Context
		r   *weka.ListStatefulClientsRequest
	}

	var scs []*weka.StatefulClient
	scs = append(scs, &scResponseWithUpStatus)
	scs = append(scs, &scResponseWithDegradedStatus)
	scs = append(scs, &scResponseWithDownStatus)
	scs = append(scs, &scResponseWithCustomStatus)

	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *weka.ListStatefulClientsResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "list sc",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListStatefulClientsRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want: &weka.ListStatefulClientsResponse{
				StatefulClients: scs,
			},
			assertion: assert.NoError,
		},
		{
			name: "apply filter",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListStatefulClientsRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Filter: &weka.ListStatefulClientsRequest_Filter{
						Names: []string{bScUP.Name},
					},
				},
			},
			want: &weka.ListStatefulClientsResponse{
				StatefulClients: []*weka.StatefulClient{&scResponseWithUpStatus},
			},
			assertion: assert.NoError,
		},
		{
			name: "fail on error",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListStatefulClientsRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
					Filter: &weka.ListStatefulClientsRequest_Filter{
						Names: []string{"error4"},
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.Unknown),
		},
		{
			name: "fail on not found",
			fields: fields{
				Backends: map[string]backend.Interface{
					"00000000-0000-0000-0000-000000000001": &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListStatefulClientsRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on no uuid",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockStatefulClientBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r:   &weka.ListStatefulClientsRequest{},
			},
			want:      nil,
			assertion: AssertStatus(codes.NotFound),
		},
		{
			name: "fail on unsupported",
			fields: fields{
				Backends: map[string]backend.Interface{
					testClusterUuid: &MockNoStatefulClientsBackend{},
				},
			},
			args: args{
				ctx: context.Background(),
				r: &weka.ListStatefulClientsRequest{
					ClusterId: &v1.ClusterIdentifier{
						Uuid: testClusterUuid,
					},
				},
			},
			want:      nil,
			assertion: AssertStatus(codes.FailedPrecondition),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatefulClientHandler{
				Backends: tt.fields.Backends,
			}
			got, err := h.ListStatefulClients(tt.args.ctx, tt.args.r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func AssertStatus(c codes.Code) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
		status, _ := status.FromError(err)
		if c != status.Code() {
			return assert.Fail(t, fmt.Sprintf("Unexpected gRPC status: \n"+
				"expected: %s\n"+
				"actual  : %s", c.String(), status.Code().String()), msgAndArgs...)
		}
		return true
	}
}
