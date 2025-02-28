// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package reconciler_query

import (
	"context"
	"database/sql"

	"fmt"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	GetLatestRev = `
	SELECT cluster.clusterstate_name, 
	clusterrevVariantTable.clusterrev_id, clusterrevVariantTable.desiredspec_json
	FROM cluster 
	INNER JOIN( SELECT t.*
	FROM (
		SELECT *, ROW_NUMBER() OVER (PARTITION BY  cluster_id ORDER BY timestamp DESC) 
		AS row_num
		FROM clusterrev
	) AS t
	WHERE t.row_num = 1) AS clusterrevVariantTable 
	ON cluster.cluster_id = clusterrevVariantTable.cluster_id
	WHERE cluster.clusterstate_name = 'DeletePending' 
	OR cluster.clusterstate_name = 'Pending' 
	`
	GetLatestRevWithOption = `
	SELECT cluster.clusterstate_name, 
	clusterrevVariantTable.clusterrev_id, clusterrevVariantTable.desiredspec_json
	FROM cluster 
	INNER JOIN( SELECT t.*
	FROM (
		SELECT *, ROW_NUMBER() OVER (PARTITION BY  cluster_id ORDER BY timestamp DESC) 
		AS row_num
		FROM clusterrev
	) AS t
	WHERE t.row_num = 1) AS clusterrevVariantTable 
	ON cluster.cluster_id = clusterrevVariantTable.cluster_id
	WHERE cluster.clusterstate_name = '%s'
	`
)

func GetReconcilerClusters(ctx context.Context, dbconn *sql.DB, req *pb.ClusterReconcilerRequest) (*pb.ClusterReconcilerResponse, error) {
	friendlyMessage := "GetReconcilerClusters.UnexpectedError"
	failedFunction := "GetReconcilerClusters."
	returnError := &pb.ClusterReconcilerResponse{}
	query := GetLatestRev
	if req.State != "" {
		// Request is not empty
		query = fmt.Sprintf(GetLatestRevWithOption, req.State)
	}
	var clusters []*pb.ClusterReconcilerResponseCluster

	rows, err := dbconn.QueryContext(ctx, query)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestRevWithOption", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		cluster := &pb.ClusterReconcilerResponseCluster{
			State:        "",
			ClusterRevId: 0,
			DesiredJson:  "",
		}
		err = rows.Scan(&cluster.State, &cluster.ClusterRevId, &cluster.DesiredJson)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestRevWithOption.rows.scan", friendlyMessage+err.Error())
		}
		clusters = append(clusters, cluster)
	}

	return &pb.ClusterReconcilerResponse{Clusters: clusters}, nil
}
