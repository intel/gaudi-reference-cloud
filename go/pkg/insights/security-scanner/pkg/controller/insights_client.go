// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/actions/vulns"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InsightsClient struct {
	insightServiceClient pb.SecurityInsightsClient
}

func NewInsightClient(ctx context.Context, insightsServerAddr string) (*InsightsClient, error) {
	logger := log.FromContext(ctx).WithName("insightsClient.NewInsightClient")

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}
	insightsConn, err := grpcutil.NewClient(ctx, insightsServerAddr, dialOptions...)
	if err != nil {
		logger.Error(err, "unable to obtain connection for security insights", logkeys.ServerAddr, insightsServerAddr)
		return nil, fmt.Errorf("security insights server grpc dial failed")
	}

	insightsClient := pb.NewSecurityInsightsClient(insightsConn)

	return &InsightsClient{insightServiceClient: insightsClient}, nil
}

type releaseImage struct {
	ReleaseId string
	Name      string
	Version   string
	Sha256    string
}

func (apiclient *InsightsClient) GetAllImages(ctx context.Context) ([]releaseImage, error) {
	logger := log.FromContext(ctx).WithName("insightsClient.GetAllImages")
	imageList := []releaseImage{}
	logger.Info("retreiving all release images")

	releases, err := apiclient.insightServiceClient.GetAllReleases(ctx, &pb.ReleaseFilter{Top: 0})
	if err != nil {
		logger.Info("error fetching all releases", "error", err)
		return nil, fmt.Errorf("error getting release images")
	}

	for _, release := range releases.Releases {
		for _, comp := range release.Components {
			if comp.Type == pb.ComponentType_OCI_IMAGE {
				imageList = append(imageList, releaseImage{
					ReleaseId: release.ReleaseId,
					Name:      comp.Name,
					Version:   comp.Version,
					Sha256:    comp.Sha256,
				})
			}
		}
	}
	logger.Info("returing from release images discovery", "total images", len(imageList))

	return imageList, nil
}

func (apiclient *InsightsClient) StoreVulnerabilityReport(ctx context.Context, image releaseImage, report vulns.VulnerabilityReport) error {
	logger := log.FromContext(ctx).WithName("InsightsClient.StoreVulnerabilityReport")

	logger.Info("storing vulnerability report ", "releaseId", image.ReleaseId)

	in := pb.VulnerabilityReport{}
	in.ReleaseId = image.ReleaseId
	in.ComponentName = image.Name
	in.ComponentSHA256 = image.Sha256
	in.ComponentVersion = image.Version
	in.ScanTimestamp = timestamppb.New(time.Now())

	for _, v := range report.Vulnerabilities {
		vPb := pb.Vulnerability{
			Id:               v.Id,
			Description:      v.Title,
			AffectedPackage:  v.PkgName,
			AffectedVersions: v.InstalledVersion,
			Severity:         v.Severity,
			PublishedAt:      timestamppb.New(v.PublishedAt),
			FixedVersion:     v.FixedVersion,
		}
		in.Vulnerabilities = append(in.Vulnerabilities, &vPb)
	}

	_, err := apiclient.insightServiceClient.StoreReleaseVulnerabilityReport(ctx, &in)
	if err != nil {
		logger.Error(err, "error storing vulnerability report")
		return fmt.Errorf("error storing vulnerability report")
	}

	return nil
}
