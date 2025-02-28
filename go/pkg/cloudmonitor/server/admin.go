// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"

	empty "github.com/golang/protobuf/ptypes/empty"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// type Server struct {
// 	pb.UnimplementedCloudMonitorServiceServer
// 	session *sql.DB
// 	cfg     config.Config
// }

func (c *Server) GetResourceCategories(ctx context.Context, req *empty.Empty) (*pb.GetResourceCategoriesResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudMonitorService.GetResourceCategories").Start()
	defer span.End()

	logger.Info("list resources invoked")

	returnResponse := &pb.GetResourceCategoriesResponse{
		Resourcetypes: []*pb.Resourcetypes{},
	}

	query := `
		select resourcetype, resourcetype_id from resourcetypes	
	`
	rows, err := c.session.QueryContext(ctx, query)
	if err != nil {
		logger.Error(err, "error while quering database")
		return &pb.GetResourceCategoriesResponse{}, fmt.Errorf("unable to retrieve records")
	}

	for rows.Next() {
		resourcetype := pb.Resourcetypes{}
		err = rows.Scan(&resourcetype.Resourcetypes, &resourcetype.ResourceId)
		if err != nil {
			logger.Error(err, "failed to retrieve table row")
			return &pb.GetResourceCategoriesResponse{}, fmt.Errorf("unable to retrieve records")
		}
		returnResponse.Resourcetypes = append(returnResponse.Resourcetypes, &resourcetype)
	}

	return returnResponse, nil
}

func (c *Server) GetIntervals(ctx context.Context, req *pb.GetIntervalsRequest) (*pb.GetIntervalsResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudMonitorService.GetIntervals").Start()
	defer span.End()

	logger.Info("list resources invoked")

	returnResponse := &pb.GetIntervalsResponse{
		Interval: []*pb.Intervals{},
	}

	query := `
		select intervals.interval as interval
		from intervals inner join resourcetypes on intervals.resourcetype_id=resourcetypes.resourcetype_id where
		resourcetypes.resourcetype = $1
	`
	args := []any{req.Category}

	// rows, err := c.session.QueryContext(ctx, query, args...)
	rows, err := c.session.QueryContext(ctx, query, args...)

	if err != nil {
		logger.Error(err, "error while quering database")
		return &pb.GetIntervalsResponse{}, fmt.Errorf("unable to retrieve records")
	}

	for rows.Next() {
		intervals := pb.Intervals{}
		err = rows.Scan(&intervals.Interval)
		if err != nil {
			logger.Error(err, "failed to retrieve table row")
			return &pb.GetIntervalsResponse{}, fmt.Errorf("unable to retrieve records")
		}
		returnResponse.Interval = append(returnResponse.Interval, &intervals)
	}

	return returnResponse, nil
}

func (c *Server) GetMetricTypes(ctx context.Context, req *pb.GetMetricTypesRequest) (*pb.GetMetricTypesResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudMonitorService.GetIntervals").Start()
	defer span.End()

	logger.Info("list resources invoked")

	returnResponse := &pb.GetMetricTypesResponse{
		Metric: []*pb.Metrics{},
	}

	query := `
		select metricstypes.metricstype as metricstype 
		from metricstypes inner join resourcetypes on metricstypes.resourcetype_id=resourcetypes.resourcetype_id where resourcetypes.resourcetype = $1
	`
	args := []any{req.Category}

	// rows, err := c.session.QueryContext(ctx, query, args...)
	rows, err := c.session.QueryContext(ctx, query, args...)

	if err != nil {
		logger.Error(err, "error while quering database")
		return &pb.GetMetricTypesResponse{}, fmt.Errorf("unable to retrieve records")
	}

	for rows.Next() {
		metrics := pb.Metrics{}
		err = rows.Scan(&metrics.Metric)
		if err != nil {
			logger.Error(err, "failed to retrieve table row")
			return &pb.GetMetricTypesResponse{}, fmt.Errorf("unable to retrieve records")
		}
		returnResponse.Metric = append(returnResponse.Metric, &metrics)
	}

	return returnResponse, nil
}
