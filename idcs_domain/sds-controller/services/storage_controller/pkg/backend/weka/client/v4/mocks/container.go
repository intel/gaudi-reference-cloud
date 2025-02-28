// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) GetContainersWithResponse(ctx context.Context, reqEditors ...v4.RequestEditorFn) (*v4.GetContainersResponse, error) {
	cores := 1
	resp := v4.GetContainersResponse{
		JSON200: &struct {
			Data *[]v4.Container "json:\"data,omitempty\""
		}{
			Data: &[]v4.Container{
				{
					Hostname: w("sc"),
					Uid:      w("scId"),
					Status:   w("UP"),
					Mode:     w("client"),
					Cores:    &cores,
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetContainersWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetContainersWithResponse.empty") != nil {
		return &v4.GetContainersResponse{}, nil
	}

	if ctx.Value("testing.GetContainersWithResponse.incomplete") != nil {
		c := *resp.JSON200.Data
		c[0].Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetSingleContainerWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.GetSingleContainerResponse, error) {
	cores := 1
	resp := v4.GetSingleContainerResponse{
		JSON200: &struct {
			Data *v4.Container "json:\"data,omitempty\""
		}{
			Data: &v4.Container{
				Hostname: w("sc"),
				Uid:      w("scId"),
				Status:   w("UP"),
				Mode:     w("client"),
				Cores:    &cores,
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.empty") != nil {
		return &v4.GetSingleContainerResponse{}, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.incomplete_with_Uid_nil") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.incomplete_with_Hostname_nil") != nil {
		resp.JSON200.Data.Hostname = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.incomplete_with_Status_nil") != nil {
		resp.JSON200.Data.Status = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.incomplete_with_Mode_nil") != nil {
		resp.JSON200.Data.Mode = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.incomplete_with_Cores_nil") != nil {
		resp.JSON200.Data.Cores = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.statusdown") != nil {
		cores := 1
		resp := v4.GetSingleContainerResponse{
			JSON200: &struct {
				Data *v4.Container "json:\"data,omitempty\""
			}{
				Data: &v4.Container{
					Hostname: w("sc"),
					Uid:      w("scId"),
					Status:   w("DOWN"),
					Mode:     w("client"),
					Cores:    &cores,
				},
			},
			HTTPResponse: &http.Response{
				StatusCode: 200,
			},
		}
		return &resp, nil
	}

	if ctx.Value("testing.GetSingleContainerWithResponse.nonClientUidError") != nil {
		resp = v4.GetSingleContainerResponse{
			JSON200: &struct {
				Data *v4.Container "json:\"data,omitempty\""
			}{
				Data: &v4.Container{
					Hostname: w("sc"),
					Uid:      w("scId"),
					Status:   w("UP"),
					Mode:     w("backend"),
					Cores:    &cores,
				},
			},
			HTTPResponse: &http.Response{
				StatusCode: 200,
			},
		}
		return &resp, nil

	}

	return &resp, nil
}

func (*MockWekaClient) AddContainerWithResponse(ctx context.Context, body v4.AddContainerJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.AddContainerResponse, error) {
	cores := 1
	resp := v4.AddContainerResponse{
		JSON200: &struct {
			Data *v4.Container "json:\"data,omitempty\""
		}{
			Data: &v4.Container{
				Hostname: w("sc"),
				Uid:      w("scId"),
				Status:   w("UP"),
				Mode:     w("client"),
				Cores:    &cores,
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.AddContainerWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.AddContainerWithResponse.empty") != nil {
		return &v4.AddContainerResponse{}, nil
	}

	if ctx.Value("testing.AddContainerWithResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetProcessesWithResponse(ctx context.Context, reqEditors ...v4.RequestEditorFn) (*v4.GetProcessesResponse, error) {
	management_roles := []string{"MANAGEMENT"}
	frontend_roles := []string{"FRONTEND"}
	resp := v4.GetProcessesResponse{
		JSON200: &struct {
			Data *[]v4.Process "json:\"data,omitempty\""
		}{
			Data: &[]v4.Process{
				{
					Hostname: w("sc"),
					Uid:      w("scProcessId1"),
					Status:   w("UP"),
					Mode:     w("client"),
					Roles:    &management_roles,
				},
				{
					Hostname: w("sc"),
					Uid:      w("scProcessId2"),
					Status:   w("UP"),
					Mode:     w("client"),
					Roles:    &frontend_roles,
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetProcessesWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetProcessesWithResponse.empty") != nil {
		return &v4.GetProcessesResponse{}, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.incomplete_with_Uid_nil") != nil {
		c := *resp.JSON200.Data
		c[0].Uid = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.incomplete_with_Hostname_nil") != nil {
		c := *resp.JSON200.Data
		c[0].Hostname = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.incomplete_with_Status_nil") != nil {
		c := *resp.JSON200.Data
		c[0].Status = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.incomplete_with_Mode_nil") != nil {
		c := *resp.JSON200.Data
		c[0].Mode = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.incomplete_with_Roles_nil") != nil {
		c := *resp.JSON200.Data
		c[0].Roles = nil
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.incorrect_with_more_than_one_role") != nil {
		c := *resp.JSON200.Data
		roles := []string{"FRONTEND", "COMPUTE"}
		c[1].Roles = &roles
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.notready") != nil {
		management_roles := []string{"MANAGEMENT"}
		frontend_roles := []string{"FRONTEND"}
		resp := v4.GetProcessesResponse{
			JSON200: &struct {
				Data *[]v4.Process "json:\"data,omitempty\""
			}{
				Data: &[]v4.Process{
					{
						Hostname: w("sc2"),
						Uid:      w("scProcessId1"),
						Status:   w("UP"),
						Mode:     w("client"),
						Roles:    &management_roles,
					},
					{
						Hostname: w("sc2"),
						Uid:      w("scProcessId2"),
						Status:   w("UP"),
						Mode:     w("client"),
						Roles:    &frontend_roles,
					},
				},
			},
			HTTPResponse: &http.Response{
				StatusCode: 200,
			},
		}
		return &resp, nil
	}

	if ctx.Value("testing.GetProcessesWithResponse.down") != nil {
		management_roles := []string{"MANAGEMENT"}
		frontend_roles := []string{"FRONTEND"}
		resp := v4.GetProcessesResponse{
			JSON200: &struct {
				Data *[]v4.Process "json:\"data,omitempty\""
			}{
				Data: &[]v4.Process{
					{
						Hostname: w("sc"),
						Uid:      w("scProcessId1"),
						Status:   w("DOWN"),
						Mode:     w("client"),
						Roles:    &management_roles,
					},
					{
						Hostname: w("sc"),
						Uid:      w("scProcessId2"),
						Status:   w("DOWN"),
						Mode:     w("client"),
						Roles:    &frontend_roles,
					},
				},
			},
			HTTPResponse: &http.Response{
				StatusCode: 200,
			},
		}
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) UpdateContainerWithResponse(ctx context.Context, uid string, body v4.UpdateContainerJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.UpdateContainerResponse, error) {
	resp := v4.UpdateContainerResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.UpdateContainerWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.UpdateContainerWithResponse.empty") != nil {
		return &v4.UpdateContainerResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) CreateContainerNetworkWithResponse(ctx context.Context, uid string, body v4.CreateContainerNetworkJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.CreateContainerNetworkResponse, error) {
	resp := v4.CreateContainerNetworkResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.CreateContainerNetworkWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateContainerNetworkWithResponse.empty") != nil {
		return &v4.CreateContainerNetworkResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) ApplyContainerWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.ApplyContainerResponse, error) {
	resp := v4.ApplyContainerResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.ApplyContainerWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.ApplyContainerWithResponse.empty") != nil {
		return &v4.ApplyContainerResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) DeactivateContainerWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.DeactivateContainerResponse, error) {
	resp := v4.DeactivateContainerResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.DeactivateContainerWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DeactivateContainerWithResponse.empty") != nil {
		return &v4.DeactivateContainerResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) RemoveContainerWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.RemoveContainerResponse, error) {
	resp := v4.RemoveContainerResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.RemoveContainerWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.RemoveContainerWithResponse.empty") != nil {
		return &v4.RemoveContainerResponse{}, nil
	}

	return &resp, nil
}
