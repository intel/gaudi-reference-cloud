// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/apis/private.cloud/v1alpha1"
	"k8s.io/client-go/rest"
)

type ProductVendorService struct {
	pb.UnimplementedProductVendorServiceServer
	restClient *rest.RESTClient
}

func NewProductVendorService(restClient *rest.RESTClient) (*ProductVendorService, error) {
	if restClient == nil {
		return nil, fmt.Errorf("client is required")
	}
	return &ProductVendorService{
		restClient: restClient,
	}, nil
}

func (srv *ProductVendorService) Read(ctx context.Context, filter *pb.VendorFilter) (*pb.VendorResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductVendorService.Read").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	vList := v1alpha1.VendorList{}

	err := srv.restClient.
		Get().
		Resource("vendors").
		Do(ctx).
		Into(&vList)

	if err != nil {
		logger.Error(err, "error reading vendorList crd", "filter", filter, "context", "Get()")
		return nil, fmt.Errorf("error reading vendorList crd: %v", err)
	}
	filterVendorResponse(&vList, filter)
	response, err := marshalVendorResponse(vList)

	if err != nil {
		logger.Error(err, "error creating vendorList", "filter", filter, "context", "marshalVendorResponse")
		return nil, fmt.Errorf("error creating vendorList: %v", err)
	}

	return response, nil
}

func marshalVendorResponse(vList v1alpha1.VendorList) (*pb.VendorResponse, error) {
	res := pb.VendorResponse{}

	for _, v := range vList.Items {
		ts, err := ptypes.TimestampProto(v.CreationTimestamp.Time)
		if err != nil {
			return nil, err
		}

		vr := pb.Vendor{
			Id:          v.Spec.ID,
			Description: v.Spec.Description,
			Name:        v.ObjectMeta.Name,
			Created:     ts,
		}

		families := []*pb.ProductFamily{}
		for _, f := range v.Spec.Families {
			fr := pb.ProductFamily{
				Name:        f.Name,
				Id:          f.ID,
				Description: f.Description,
				Created:     ts,
			}
			families = append(families, &fr)
		}
		vr.Families = families
		res.Vendors = append(res.Vendors, &vr)
	}
	return &res, nil
}

func filterVendorResponse(vList *v1alpha1.VendorList, filter *pb.VendorFilter) {

	insertIdx := 0
	for _, vendor := range vList.Items {
		skip := false
		if filter.Id != nil {
			if !strings.EqualFold(filter.GetId(), vendor.Spec.ID) {
				skip = true
			}
		}
		if filter.Name != nil {
			if !strings.EqualFold(filter.GetName(), vendor.ObjectMeta.Name) {
				skip = true
			}
		}
		if !skip {
			vList.Items[insertIdx] = vendor
			insertIdx++
		}
	}
	vList.Items = vList.Items[:insertIdx]
}
