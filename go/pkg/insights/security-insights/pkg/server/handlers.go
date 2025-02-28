// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-insights/pkg/db/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type SecurityInsightsServer struct {
	v1.UnimplementedSecurityInsightsServer
	session *sql.DB
}

func NewSecurityInsightsService(session *sql.DB) (*SecurityInsightsServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &SecurityInsightsServer{
		session: session,
	}, nil
}

func (srv *SecurityInsightsServer) CreateRelease(ctx context.Context, in *v1.K8SReleaseMD) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.CreateRelease")

	log.Info("CreateRelease", "entering create release for ", in.ReleaseId)
	defer log.Info("CreateRelease", "returning from create release for ", in.ReleaseId)

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	insertIdx, err := query.StoreReleaseMetadata(ctx, dbSession, in)
	if err != nil {
		return nil, err
	}
	log.Info("record inserted successfully", "insertIdx", insertIdx)
	return &emptypb.Empty{}, nil
}

func (srv *SecurityInsightsServer) GetRelease(ctx context.Context, in *v1.GetReleaseRequest) (*v1.K8SReleaseMD, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetRelease")
	log.Info("GetRelease", "entering get release for ", in.ReleaseId)
	defer log.Info("GetRelease", "returning from get release for ", in.ReleaseId)
	//validate input
	if err := validateGetRelease(in); err != nil {
		return nil, err
	}
	//check valid releaseId for component
	if valid := srv.checkReleaseId(ctx, in.ReleaseId, in.Component.String()); !valid {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid ReleaseId")
	}

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	res, err := query.GetReleaseMetadata(ctx, dbSession, in)
	if err != nil {
		log.Error(err, "error reading release metadata")
		return nil, fmt.Errorf("error reading release metadata")
	}
	return res, nil
}

func (srv *SecurityInsightsServer) GetAllReleases(ctx context.Context, in *v1.ReleaseFilter) (*v1.K8SReleaseMDList, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetAllReleases")
	log.Info("GetAllReleases", "entering GetAllReleases for ", in)
	defer log.Info("GetAllReleases", "returning from GetAllReleases for", in)

	//validate input
	if err := in.Validate(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid Component")
	}
	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	topK := 0
	if in.Top != 0 {
		topK = int(in.Top)
	}
	releases, err := query.GetAllReleases(ctx, topK, in.Component, dbSession)
	if err != nil {
		log.Error(err, "error fetching releases from db")
		return nil, status.Errorf(codes.Internal, "failed to retrieve release details")
	}
	releaseList := &v1.K8SReleaseMDList{
		Releases: releases,
	}
	return releaseList, nil
}

func (srv *SecurityInsightsServer) StoreReleaseSBOM(ctx context.Context, in *v1.ReleaseSBOM) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.StoreReleaseSBOM")
	log.Info("StoreReleaseSBOM", "entering store release sbom for ", in.ReleaseId)
	defer log.Info("StoreReleaseSBOM", "returning from store release sbom", in.ReleaseId)

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	if err := query.StoreSBOM(ctx, dbSession, in); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (srv *SecurityInsightsServer) StoreReleaseComponent(ctx context.Context, in *v1.ReleaseComponents) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.StoreReleaseComponent")

	log.Info("StoreReleaseComponent", "entering store release component for ", in.ReleaseId)
	defer log.Info("StoreReleaseComponent", "returning from store release component for ", in.ReleaseId)

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	insertIdx, err := query.StoreReleaseComponentsMetadata(ctx, dbSession, in)
	if err != nil {
		return nil, err
	}
	log.Info("record inserted successfully", "insertIdx", insertIdx)
	return &emptypb.Empty{}, nil

}

func (srv *SecurityInsightsServer) GetReleaseComponent(in *v1.GetReleaseRequest, rs v1.SecurityInsights_GetReleaseComponentServer) error {
	return nil
}

func (srv *SecurityInsightsServer) GetReleaseSBOM(ctx context.Context, in *v1.GetReleaseRequest) (*v1.SBOM, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetReleaseSBOM")
	log.Info("GetReleaseSBOM", "entering read release sbom for ", in.ReleaseId)
	defer log.Info("GetReleaseSBOM", "returning from read release sbom", in.ReleaseId)

	//validate input
	if err := validateGetRelease(in); err != nil {
		return nil, err
	}
	//check valid releaseId for component
	if valid := srv.checkReleaseId(ctx, in.ReleaseId, in.Component.String()); !valid {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid ReleaseId")
	}
	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	return query.ReadSBOM(ctx, dbSession, in)
}

func (srv *SecurityInsightsServer) StoreReleaseVulnerabilityReport(ctx context.Context, in *v1.VulnerabilityReport) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.StoreReleaseVulnerabilityReport")
	log.Info("StoreReleaseVulnerabilityReport", "store vulnerability report ", in.ReleaseId)
	defer log.Info("StoreReleaseVulnerabilityReport", "store vulnerability report ", in.ReleaseId)

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	if err := query.StoreReleaseVulnerability(ctx, dbSession, in); err != nil {
		return nil, status.Errorf(codes.Internal, "no database connection found.")
	}
	return &emptypb.Empty{}, nil
}

func (srv *SecurityInsightsServer) GetReleaseVulnerabilities(ctx context.Context, in *v1.GetReleaseRequest) (*v1.VulnerabilitiesResult, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetReleaseVulnerabilities")
	log.Info("GetReleaseVulnerabilities", "component", in.Component, "release", in.ReleaseId)
	defer log.Info("GetReleaseVulnerabilities", "component", in.Component, "release ", in.ReleaseId)
	//validate input
	if err := validateGetRelease(in); err != nil {
		return nil, err
	}
	//check valid releaseId for component
	if valid := srv.checkReleaseId(ctx, in.ReleaseId, in.Component.String()); !valid {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid ReleaseId")
	}

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	res, err := query.ReadReleaseVulnerability(ctx, dbSession, in.Component, in.ReleaseId)
	if err != nil {
		log.Error(err, "failed to retrieve release details from db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return res, nil
}

func (srv *SecurityInsightsServer) UpdateRecommendationPolicy(ctx context.Context, in *v1.RecommendationPolicy) (*emptypb.Empty, error) {
	return nil, nil
}

func (srv *SecurityInsightsServer) GetRecommendationPolicies(ctx context.Context, in *v1.PolicyFilters) (*v1.AllPoliciesResponse, error) {
	return nil, nil
}

func (srv *SecurityInsightsServer) GetRecommendationPolicy(ctx context.Context, in *v1.PolicyId) (*v1.PolicyDetails, error) {
	return nil, nil
}

func (srv *SecurityInsightsServer) StoreCISReport(ctx context.Context, in *v1.CISReport) (*emptypb.Empty, error) {
	return nil, nil
}

func (srv *SecurityInsightsServer) GetCISReport(ctx context.Context, in *v1.GetCISRequest) (*v1.CISReport, error) {
	return nil, nil
}

func (srv *SecurityInsightsServer) GetSummary(ctx context.Context, in *v1.GetReleaseRequest) (*v1.SecuritySummary, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetSummary")
	log.Info("GetSummary", "get summary by releaseId ", in.ReleaseId)
	defer log.Info("GetSummary", "returning from get summary by releaseId ", in.ReleaseId)
	//validate input
	if err := validateGetRelease(in); err != nil {
		return nil, err
	}
	//check valid releaseId for component
	if valid := srv.checkReleaseId(ctx, in.ReleaseId, in.Component.String()); !valid {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid ReleaseId")
	}

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	// Retrieve report from db
	vulns, err := query.ReadReleaseVulnerability(ctx, dbSession, in.Component, in.ReleaseId)
	if err != nil {
		log.Error(err, "error fetching vulns from db")
		return nil, status.Errorf(codes.FailedPrecondition, "db record retrieval failed")
	}
	vulnSummary := []*v1.VulnerabilitySummary{}
	for _, vuln := range vulns.Report {
		countMap := make(map[string]uint32)
		// Count number of severity
		for _, value := range vuln.Vulnerabilities {
			if _, exists := countMap[value.Severity]; exists {
				countMap[value.Severity]++
			} else {
				countMap[value.Severity] = 1
			}
		}
		report := &v1.VulnerabilitySummary{
			ReleaseId:          vuln.ReleaseId,
			ComponentName:      vuln.ComponentName,
			ComponentVersion:   vuln.ComponentVersion,
			ComponentSHA256:    vuln.ComponentSHA256,
			ScanTimestamp:      vuln.ScanTimestamp,
			ScanTool:           vuln.ScanTool,
			VulnerabilityCount: countMap,
		}
		vulnSummary = append(vulnSummary, report)
	}

	summary := &v1.SecuritySummary{
		ReleaseId:       in.ReleaseId,
		Vulnerabilities: vulnSummary,
	}

	return summary, nil
}
func (srv *SecurityInsightsServer) GetUpdateRecommendation(ctx context.Context, in *v1.RecommendationFilter) (*v1.UpdateRecommendations, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetUpdateRecommendation")

	log.Info("GetUpdateRecommendation", "get recommendation by policy ", in.PolicyId)
	defer log.Info("GetUpdateRecommendation", "returning from recommendation by policy ", in.PolicyId)

	reco := v1.UpdateRecommendations{}

	reco.PolicyId = in.PolicyId
	reco.CurrentVersion = in.CurrentVersion
	reco.Vendor = in.Vendor

	var ver1, ver2, ver3 v1.RecommendedVersion
	ver1.Version = "v1.25.7"
	ver1.CreatedAt = "2022-12-16T15:34:00.000Z"

	ver2.Version = "v1.25.8"
	ver2.CreatedAt = "2023-01-16T15:34:00.000Z"

	ver3.Version = "v1.26.1"
	ver3.CreatedAt = "2023-02-16T15:34:00.000Z"

	reco.Versions = append(reco.Versions, &ver1)
	reco.Versions = append(reco.Versions, &ver2)
	reco.Versions = append(reco.Versions, &ver3)

	return &reco, nil
}

func (srv *SecurityInsightsServer) GetAllComponents(ctx context.Context, in *v1.ReleaseFilter) (*v1.ComponentList, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.GetAllComponents")
	log.Info("entering search GetAllComponents")
	defer log.Info("returning from GetAllComponents")
	//validate input
	if err := in.Validate(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid Component")
	}

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	topK := 0
	//kubernetes
	releases, err := query.GetAllReleases(ctx, topK, pb.ReleaseComponent_UNSPECIFIED_COMPONENT, dbSession)
	if err != nil {
		log.Error(err, "error fetching k8s releases from db")
		return nil, status.Errorf(codes.Internal, "failed to retrieve release details")
	}
	k8sList := &v1.Releases{}
	calicoList := &v1.Releases{}
	for _, r := range releases {
		if strings.HasPrefix(r.ReleaseId, "v1") {
			k8sList.Id = append(k8sList.Id, r.ReleaseId)
		}
		if strings.HasPrefix(r.ReleaseId, "v3") {
			calicoList.Id = append(calicoList.Id, r.ReleaseId)
		}
	}
	components := make(map[string]*pb.Releases)
	components["KUBERNETES"] = k8sList
	components["CALICO"] = calicoList
	releaseList := &v1.ComponentList{Components: components}
	return releaseList, nil
}

func (srv *SecurityInsightsServer) CompareReleaseVulnerabilities(ctx context.Context, in *v1.ReleaseComparisonFilter) (*pb.VulnerabilityComparisonReport, error) {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.CompareReleaseVulnerabilities")
	log.Info("entering CompareReleaseVulnerabilities")
	defer log.Info("returning from CompareReleaseVulnerabilities")
	// validate input
	if err := validateComparisonReleases(in); err != nil {
		log.Error(err, "validation failed")
		return nil, err
	} else {
		log.Info("validation passed")
	}
	//check valid releaseId for component
	validCurr := srv.checkReleaseId(ctx, in.CurrReleaseId, in.ComponentName.String())
	validNew := srv.checkReleaseId(ctx, in.NewReleaseId, in.ComponentName.String())
	if !validCurr || !validNew {
		return nil, status.Errorf(codes.FailedPrecondition, "Invalid ReleaseId")
	}

	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	// fetch curr vulns
	currVulns, err := query.ReadReleaseVulnerability(ctx, dbSession, in.ComponentName, in.CurrReleaseId)
	if err != nil {
		log.Error(err, "failed to retrieve release details from db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	// fetch new vulns
	newVulns, err := query.ReadReleaseVulnerability(ctx, dbSession, in.ComponentName, in.NewReleaseId)
	if err != nil {
		log.Error(err, "failed to retrieve release details from db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	currReport := make(map[string]*pb.VulnerabilityReport)
	newReport := make(map[string]*pb.VulnerabilityReport)
	// make set of all curr vulns
	for _, component := range currVulns.Report {
		currReport[component.ComponentName] = component
	} // make set of all new vulns
	for _, component := range newVulns.Report {
		newReport[component.ComponentName] = component
	}
	var report []*pb.VulnerabilityComparisonSummary
	var summary []*pb.VulnerabilityDiscovery

	// Iterate through the reports by component name
	for componentName, curr := range currReport {
		// Find new report by the same componentName
		new, exists := newReport[componentName]
		if exists {
			// Compare vulnerabilities between current and new report
			common, uniqueToCurr, uniqueToNew := compareVulnerabilities(curr.Vulnerabilities, new.Vulnerabilities)
			log.Info("component report", "name", componentName, "fixed", len(uniqueToCurr), "common", len(common), "new", len(uniqueToNew))
			summary = append(summary, &pb.VulnerabilityDiscovery{
				Component: componentName,
				Fixed:     uint32(len(uniqueToCurr)),
				Common:    uint32(len(common)),
				New:       uint32(len(uniqueToNew)),
			})
			// Append the comparison result for the component
			currList := &pb.VulnerabilitiesWrapper{
				Vulns: uniqueToCurr,
			}
			commonList := &pb.VulnerabilitiesWrapper{
				Vulns: common,
			}
			newList := &pb.VulnerabilitiesWrapper{
				Vulns: uniqueToNew,
			}
			vulns := &v1.VulnerabilityComparisonSummary{
				ComponentName:       componentName,
				CurrReleaseId:       in.CurrReleaseId,
				NewReleaseId:        in.NewReleaseId,
				Fixed:               currList,
				Common:              commonList,
				New:                 newList,
				CurrScanTimestamp:   curr.ScanTimestamp,
				NewScanTimestamp:    new.ScanTimestamp,
				CurrComponentSHA256: curr.ComponentSHA256,
				NewComponentSHA256:  new.ComponentSHA256,
				ScanTool:            "trivy",
			}
			report = append(report, vulns)
		}
	}
	finalReport := &pb.VulnerabilityComparisonReport{
		Summary: summary,
		Report:  report,
	}
	return finalReport, nil
}

// private function to validate releaseId and component
func (srv *SecurityInsightsServer) checkReleaseId(ctx context.Context, releaseId string, component string) bool {
	log := log.FromContext(ctx).WithName("SecurityInsightsServer.checkReleaseId")
	res, err := srv.GetAllComponents(ctx, &v1.ReleaseFilter{})
	if err != nil {
		return false
	}
	for _, id := range res.Components[component].Id {
		if releaseId == id {
			return true
		}
	}
	log.Info("Invalid releaseId")
	return false
}
