// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package admin_query

import (
	"context"
	"database/sql"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"

	empty "github.com/golang/protobuf/ptypes/empty"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	deleteimiquery = `
	DELETE FROM public.osimageinstance where osimageinstance_name = $1`

	deleteinstanceTypequery = `
	DELETE FROM public.instancetype where instancetype_name = $1`

	imiexistsquery = `
	SELECT count(*) FROM public.osimageinstance
		WHERE osimageinstance_name = $1
		`

	instanceTypeexistsquery = `
	SELECT count(*) FROM public.instancetype
		WHERE instancetype_name = $1
		`

	getInstanceTypefromNodegroupquery = `
	SELECT count(*) FROM public.nodegroup where instancetype_name = $1`

	getOsinstancefromNodegroupquery = `
	SELECT count(*) FROM public.nodegroup where osimageinstance_name = $1`

	getOsinstancefromComponentquery = `
	SELECT count(*) FROM public.osimageinstancecomponent where osimageinstance_name = $1`

	k8sVersionExistsquery = `SELECT count(*) FROM public.k8sversion WHERE k8sversion_name = $1`

	getK8sVersionfromNodegroupquery = `SELECT count(*) FROM public.nodegroup where k8sversion_name = $1`

	getK8sVersionfromImis = `SELECT count(*) FROM public.osimageinstance where k8sversion_name = $1`

	deleteVersionFromCompatibilityQuery = `DELETE FROM public.k8scompatibility where k8sversion_name = $1`

	deleteVersionQuery = `DELETE FROM public.k8sversion where k8sversion_name = $1`

	deleteInstanceTypeFromCompatibilityQuery = `DELETE FROM public.k8scompatibility where instancetype_name = $1`

	deletecomponentquery = `DELETE FROM public.osimageinstancecomponent where osimageinstance_name = $1`
)

func DeleteInstanceType(ctx context.Context, dbconn *sql.DB, record *pb.DeleteInstanceTypeRequest) (*empty.Empty, error) {
	friendlyMessage := "DeleteInstanceType.UnexpectedError"
	failedFunction := "DeleteInstanceType."
	emptyReturn := &empty.Empty{}

	// Instance Type Existance Check
	var count int32
	err := dbconn.QueryRowContext(ctx, instanceTypeexistsquery,
		record.Name,
	).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"imiexistancequery", friendlyMessage+err.Error())
	}
	if count != 1 {
		return emptyReturn, grpc_status.Error(codes.NotFound, "Instance Type not found")
	}

	// Instance Type In Use by Nodegroups Check
	err = dbconn.QueryRowContext(ctx, getInstanceTypefromNodegroupquery, record.Name).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getInstanceTypefromNodegroupquery", friendlyMessage+err.Error())
	}
	if count >= 1 {
		return emptyReturn, grpc_status.Error(codes.FailedPrecondition, "Instance Type in use by nodegroups, can not be deleted")
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	_, err = tx.QueryContext(ctx, deleteInstanceTypeFromCompatibilityQuery, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteInstanceTypeFromCompatibilityQuery", friendlyMessage+err.Error())
	}

	_, err = tx.QueryContext(ctx, deleteinstanceTypequery, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteinstanceTypequery", friendlyMessage+err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}
	return emptyReturn, nil
}

func DeleteIMI(ctx context.Context, dbconn *sql.DB, record *pb.DeleteIMIRequest) (*empty.Empty, error) {
	friendlyMessage := "DeleteImi.UnexpectedError"
	failedFunction := "DeleteImi."
	emptyReturn := &empty.Empty{}

	// IMI Duplicate Existance Check
	var count int32
	err := dbconn.QueryRowContext(ctx, imiexistsquery,
		record.Name,
	).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"imiexistancequery", friendlyMessage+err.Error())
	}
	if count != 1 {
		return emptyReturn, grpc_status.Error(codes.NotFound, "IMI not found")
	}

	// IMI In Use by Nodegroups Check
	err = dbconn.QueryRowContext(ctx, getOsinstancefromNodegroupquery, record.Name).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getOsinstancefromNodegroupquery", friendlyMessage+err.Error())
	}
	if count >= 1 {
		return emptyReturn, grpc_status.Error(codes.FailedPrecondition, "IMI in use with nodegroups, can not be deleted")
	}

	// IMI In Use by Components Check
	err = dbconn.QueryRowContext(ctx, getOsinstancefromComponentquery, record.Name).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getOsinstancefromNodegroupquery", friendlyMessage+err.Error())
	}
	if count >= 1 {
		return emptyReturn, grpc_status.Error(codes.FailedPrecondition, "IMI in use with components, can not be deleted")
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}
	_, err = tx.QueryContext(ctx, deleteimiquery, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteimiquery", friendlyMessage+err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}
	return emptyReturn, nil
}

func DeleteK8SVersion(ctx context.Context, dbconn *sql.DB, record *pb.GetK8SRequest) (*empty.Empty, error) {
	friendlyMessage := "DeleteImi.UnexpectedError"
	failedFunction := "DeleteImi."
	emptyReturn := &empty.Empty{}

	// Existance Check
	var count int32
	err := dbconn.QueryRowContext(ctx, k8sVersionExistsquery,
		record.Name,
	).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"k8sVersionExistsquery", friendlyMessage+err.Error())
	}
	if count != 1 {
		return emptyReturn, grpc_status.Error(codes.NotFound, "K8s Version not found")
	}

	// K8sVersion in use by Nodegroups
	err = dbconn.QueryRowContext(ctx, getK8sVersionfromNodegroupquery, record.Name).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getK8sVersionfromNodegroupquery", friendlyMessage+err.Error())
	}
	if count >= 1 {
		return emptyReturn, grpc_status.Error(codes.FailedPrecondition, "K8s Version in use with nodegroups, can not be deleted")
	}

	// K8sVersion in use by IMIs
	err = dbconn.QueryRowContext(ctx, getK8sVersionfromImis, record.Name).Scan(&count)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getK8sVersionfromImis", friendlyMessage+err.Error())
	}
	if count >= 1 {
		return emptyReturn, grpc_status.Error(codes.FailedPrecondition, "K8s Version in use with IMIs, can not be deleted")
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}
	_, err = tx.QueryContext(ctx, deleteVersionFromCompatibilityQuery, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteVersionFromCompatabilityQuery", friendlyMessage+err.Error())
	}

	_, err = tx.QueryContext(ctx, deleteVersionQuery, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteVersionQuery", friendlyMessage+err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage+err.Error())
	}
	return emptyReturn, nil
}
