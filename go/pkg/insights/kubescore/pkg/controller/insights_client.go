// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	shared "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InsightsClient struct {
	insightServiceClient pb.SecurityInsightsClient
}

func NewInsightClient(ctx context.Context, insightsServerAddr string) (*InsightsClient, error) {
	logger := log.FromContext(ctx).WithName("insightsClient.NewInsightClient")
	logger.Info("connecting to insights client")
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

func (insightClient *InsightsClient) StoreReleaseMetadata(ctx context.Context, releaseMD common.ReleaseReport) error {
	log := log.FromContext(ctx).WithName("InsightsClient.StoreReleaseMetadata")
	rIn := v1.K8SReleaseMD{}
	rIn.ReleaseId = releaseMD.ReleaseMD.Tag
	rIn.Purl = "https://github.com/kubernetes/kubernetes/"
	rIn.ReleaseTimestamp = timestamppb.New(releaseMD.ReleaseMD.CreatedAt)
	rIn.Vendor = v1.ValidVendors_OSS_KUBE_VENDOR
	rIn.License = "Apache License 2.0" //TODO: Fix me

	rIn.EolTimestamp = timestamppb.New(releaseMD.SupportMD.EOLTime)
	rIn.EosTimestamp = timestamppb.New(releaseMD.SupportMD.EOSTime)
	for _, img := range releaseMD.Images {
		comp := v1.ReleaseComponents{}
		imgMD := strings.Split(img.URL, ":")
		if len(imgMD) == 2 {
			comp.Version = imgMD[1]
			comp.Name = imgMD[0]
		} else {
			comp.Version = img.Digest
			comp.Name = img.URL
		}
		comp.Purl = img.URL
		comp.Sha256 = img.Digest
		comp.License = "Apache License 2.0" //TODO: Fix me
		comp.ReleaseTime = timestamppb.New(releaseMD.ReleaseMD.CreatedAt)
		comp.Type = shared.MapComponentTypeToPB(shared.ComponentTypeImage)
		rIn.Components = append(rIn.Components, &comp)
	}

	resp, err := insightClient.insightServiceClient.CreateRelease(ctx, &rIn)
	if err != nil {
		log.Error(err, "error storing release md to insights")
		return err
	}
	log.Info("insight", "k8s md stored successfully", resp)

	return nil
}

func (insightClient *InsightsClient) StoreComponentVulnerability(ctx context.Context, vuln common.VulnerabilityData) error {
	return nil
}

func (insightClient *InsightsClient) StoreReleaseComponent(ctx context.Context, componentName string, releaseMD []common.ReleaseComponentMD) error {
	log := log.FromContext(ctx).WithName("InsightsClient.StoreReleaseComponent")
	for _, rd := range releaseMD {
		rIn := v1.ReleaseComponents{
			ReleaseId:   rd.ReleaseId,
			Name:        componentName,
			Version:     rd.ComponentVersion,
			Vendor:      v1.ValidVendors_OSS_KUBE_VENDOR,
			Purl:        rd.Purl,
			License:     rd.License,
			ReleaseTime: timestamppb.New(rd.ReleaseTime),
			Type:        shared.MapComponentTypeToPB(rd.Type),
		}
		_, err := insightClient.insightServiceClient.StoreReleaseComponent(ctx, &rIn)
		if err != nil {
			log.Error(err, "error storing release component md to insights")
			return err
		}
		log.Info("insight", "k8s release md stored successfully", componentName, "version", rd.ReleaseId)
	}

	return nil
}

func (insightClient *InsightsClient) GetLastReleaseTimestamp(ctx context.Context) (*time.Time, error) {
	log := log.FromContext(ctx).WithName("InsightsClient.GetLastReleaseTimestamp")
	rIn := pb.ReleaseFilter{
		Top: 1,
	}
	releases, err := insightClient.insightServiceClient.GetAllReleases(ctx, &rIn)
	if err != nil {
		log.Error(err, "error searching top release component")
		return nil, err
	}
	release := releases.Releases[len(releases.Releases)-1]

	if release != nil {
		ts := release.ReleaseTimestamp.AsTime()
		return &(ts), nil
	}
	return nil, fmt.Errorf("error reading latest release ")
}

func (insightClient *InsightsClient) StoreReleaseSBOM(ctx context.Context, releaseId, format string, creationTs time.Time, sbom []byte) error {
	logger := log.FromContext(ctx).WithName("InsightsClient.StoreReleaseSBOM")
	sbomIn := pb.ReleaseSBOM{}

	type sbomStruct struct {
		CreationInfo struct {
			CreationTime string `json:"created"`
		} `json:"creationInfo"`
	}

	sbomTmp := sbomStruct{}
	if err := json.Unmarshal(sbom, &sbomTmp); err != nil {
		logger.Error(err, "error unmarshalling sbom")
	}

	// dateString := "2021-16-06T13:52:43Z"
	date, err := time.Parse(time.RFC3339, sbomTmp.CreationInfo.CreationTime)
	if err != nil {
		logger.Error(err, "error parsing creation time from sbom")
	}

	logger.Info("entering release sbom store", "releaseId", releaseId)
	sbomIn.ReleaseId = releaseId
	sbomIn.Sbom = &pb.SBOM{}
	sbomIn.Sbom.CreateTimestamp = timestamppb.New(date)
	sbomIn.Sbom.Sbom = string(sbom)
	sbomIn.Sbom.Format = mapSBOMFormat(string(format))

	if _, err := insightClient.insightServiceClient.StoreReleaseSBOM(ctx, &sbomIn); err != nil {
		logger.Error(err, "error storing release sbom", "releaseId", releaseId)
		return err
	}
	logger.Info("release sbom stored successfully", "releaseId", releaseId)
	return nil
}

func mapSBOMFormat(format string) pb.ValidSBOMFormats {
	switch format {
	case "spdx":
		return pb.ValidSBOMFormats_SPDX_FORMAT
	case "cdx":
		return pb.ValidSBOMFormats_CYCLONEDX_FORMAT
	default:
		return pb.ValidSBOMFormats_UNSPECIFIED_FORMAT
	}
}
