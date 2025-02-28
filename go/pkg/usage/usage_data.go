// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const searchResourceUsagesQueryBase = `
		SELECT id, cloud_account_id, resource_id, resource_name, product_id, product_name, transaction_id, region, creation, quantity, unreported_quantity, 
		rate, usage_unit_type, start_time, end_time, reported FROM resource_usages
	`

const searchProductUsagesQueryBase = `
		SELECT id, cloud_account_id, product_id, product_name, region, creation, quantity, 
		rate, usage_unit_type, start_time, end_time FROM product_usages
	`

const insertResourceUsageQuery = `
		INSERT INTO resource_usages (id, cloud_account_id, resource_id, resource_name, product_id, product_name, transaction_id, region, creation, 
			expiration, quantity, unreported_quantity, rate, usage_unit_type, start_time, end_time, reported) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

const insertProductUsageQuery = `
	INSERT INTO product_usages (id, cloud_account_id, product_id, product_name, region, creation, 
		expiration, quantity, rate, usage_unit_type, start_time, end_time) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`

const insertProductUsageReportQuery = `
	INSERT INTO product_usages_report (id, product_usage_id, transaction_id, cloud_account_id, product_id, product_name, region, quantity,
		rate, unreported_quantity, usage_unit_type, timestamp, created_at, reported, start_time, end_time) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
`

const insertResourceMeteringRecord = `
		INSERT INTO resource_metering (id, resource_id, cloud_account_id, transaction_id, region, last_recorded) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`

type UsageData struct {
	session *sql.DB
}

func NewUsageData(session *sql.DB) *UsageData {
	return &UsageData{session: session}
}

func (ud UsageData) BulkUpload(ctx context.Context, bulkUploadResourceUsages *pb.BulkUploadResourceUsages) (*pb.BulkUploadResourceUsagesFailed, error) {
	logger := log.FromContext(ctx).WithName("UsageData.BulkUpload")
	logger.Info("BEGIN")
	defer logger.Info("END")

	resourceUsagesFailed := []*pb.ResourceUsageCreateFailed{}
	resourceUsagesFailedToUpload := &pb.BulkUploadResourceUsagesFailed{ResourceUsages: resourceUsagesFailed}
	creationTime := time.Now()
	for _, resourceUsageCreate := range bulkUploadResourceUsages.ResourceUsages {
		logger.Info("storing resource usage for resource",
			"id", resourceUsageCreate.ResourceId, "product id", resourceUsageCreate.ProductId)
		resourceUsageRecordCreated, err := ud.StoreResourceUsage(ctx, resourceUsageCreate, creationTime)
		if err != nil {
			logger.Error(err, "failed to store resource usage for resource",
				"id", resourceUsageCreate.ResourceId, "product id", resourceUsageCreate.ProductId)

			resourceUsagesFailedToUpload.ResourceUsages = append(resourceUsagesFailedToUpload.ResourceUsages,
				ud.getResourceUsageCreateFailed(resourceUsageCreate))
		} else {
			err = ud.StoreProductUsage(ctx, &ProductUsageCreate{
				Id:             uuid.NewString(),
				CloudAccountId: resourceUsageCreate.CloudAccountId,
				ProductId:      resourceUsageCreate.ProductId,
				ProductName:    resourceUsageCreate.ProductName,
				Region:         resourceUsageCreate.Region,
				Quantity:       resourceUsageCreate.Quantity,
				Rate:           resourceUsageCreate.Rate,
				UsageUnitType:  resourceUsageCreate.UsageUnitType,
				StartTime:      resourceUsageCreate.StartTime.AsTime(),
				EndTime:        resourceUsageCreate.EndTime.AsTime(),
			}, time.Now())

			if err != nil {
				logger.Error(err, "failed to create product usage for resource",
					"id", resourceUsageCreate.ResourceId, "product id", resourceUsageCreate.ProductId)

				err = ud.DeleteResourceUsage(ctx, resourceUsageRecordCreated.Id)

				if err != nil {
					logger.Error(err, "failed to delete resource usage upon failed creation of product usage for",
						"id", resourceUsageRecordCreated.Id)
				}

				resourceUsagesFailedToUpload.ResourceUsages = append(resourceUsagesFailedToUpload.ResourceUsages,
					ud.getResourceUsageCreateFailed(resourceUsageCreate))
			}
		}
	}

	if len(resourceUsagesFailedToUpload.ResourceUsages) != 0 {
		return resourceUsagesFailedToUpload, errors.New("failed to upload resource usages")
	}

	return resourceUsagesFailedToUpload, nil
}

func (ud UsageData) getResourceUsageCreateFailed(resourceUsageCreate *pb.ResourceUsageCreate) *pb.ResourceUsageCreateFailed {
	return &pb.ResourceUsageCreateFailed{
		CloudAccountId: resourceUsageCreate.CloudAccountId,
		ResourceId:     resourceUsageCreate.ResourceId,
		ProductId:      resourceUsageCreate.ProductId,
		ProductName:    resourceUsageCreate.ProductName,
		Region:         resourceUsageCreate.Region,
		TransactionId:  resourceUsageCreate.TransactionId,
	}
}

type ResourceMeteringCreate struct {
	ResourceId     string
	CloudAccountId string
	TransactionId  string
	Region         string
	LastRecorded   time.Time
}

type ResourceMeteringUpdate struct {
	TransactionId string
	LastRecorded  time.Time
}

type ResourceMetering struct {
	Id             string
	ResourceId     string
	CloudAccountId string
	TransactionId  string
	Region         string
	LastRecorded   time.Time
}

func (ud UsageData) StoreResourceMetering(ctx context.Context, resourceMeteringRecordCreate *ResourceMeteringCreate) (*ResourceMetering, error) {
	logger := log.FromContext(ctx).WithName("UsageData.StoreResourceMetering")
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")

	logger.V(1).Info("creating resource metering for resource",
		"id", resourceMeteringRecordCreate.ResourceId, "transaction id", resourceMeteringRecordCreate.TransactionId)
	id := uuid.NewString()

	_, err := ud.session.ExecContext(ctx,
		insertResourceMeteringRecord,
		id,
		resourceMeteringRecordCreate.ResourceId,
		resourceMeteringRecordCreate.CloudAccountId,
		resourceMeteringRecordCreate.TransactionId,
		resourceMeteringRecordCreate.Region,
		resourceMeteringRecordCreate.LastRecorded,
	)

	if err != nil {
		return nil, err
	}

	return &ResourceMetering{
		Id:             id,
		ResourceId:     resourceMeteringRecordCreate.ResourceId,
		CloudAccountId: resourceMeteringRecordCreate.CloudAccountId,
		TransactionId:  resourceMeteringRecordCreate.TransactionId,
		Region:         resourceMeteringRecordCreate.Region,
		LastRecorded:   resourceMeteringRecordCreate.LastRecorded,
	}, nil
}

func (ud UsageData) GetLastMeteringTimestampForResource(ctx context.Context, resourceId string) (time.Time, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetLastMeteringTimestampForResource")
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")

	const getLastMeteringRecordTimestamp = `
		SELECT last_recorded FROM resource_metering WHERE resource_id=$1
	`

	rows, err := ud.session.QueryContext(ctx, getLastMeteringRecordTimestamp, resourceId)
	if err != nil {
		logger.Error(err, "failed to get last metering timestamp for resource", "id", resourceId)
		return time.Time{}, err
	}

	defer rows.Close()

	lastMeteringTimestamp := time.Time{}

	if rows.Next() {
		if err := rows.Scan(&lastMeteringTimestamp); err != nil {
			logger.Error(err, "failed to get last metering timestamp for resource", "id", resourceId)
			return time.Time{}, err
		}
	} else {
		logger.Error(err, "failed to get last metering timestamp for resource", "id", resourceId)
		return time.Time{}, err
	}

	if rows.Next() {
		err := errors.New(InvalidInputArguments)
		logger.Error(err, "should have returned only one resource metering row")
		return time.Time{}, err
	}

	return lastMeteringTimestamp, nil
}

func (ud UsageData) GetMeteringForResource(ctx context.Context, resourceId string) (*ResourceMetering, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetMeteringForResource")
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")

	const getResourceMeteringRecord = `
		SELECT id, resource_id, cloud_account_id, transaction_id, region, last_recorded FROM resource_metering WHERE resource_id=$1
	`

	resourceMetering := &ResourceMetering{}

	rows, err := ud.session.QueryContext(ctx, getResourceMeteringRecord, resourceId)
	if err != nil {
		logger.Error(err, "failed to get metering for resource", "id", resourceId)
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&resourceMetering.Id, &resourceMetering.ResourceId,
			&resourceMetering.CloudAccountId, &resourceMetering.TransactionId, &resourceMetering.Region,
			&resourceMetering.LastRecorded); err != nil {
			logger.Error(err, "failed to get metering for resource", "id", resourceId)
			return nil, err
		}
	} else {
		logger.Error(err, "failed to get metering timestamp for resource", "id", resourceId)
		return nil, err
	}

	if rows.Next() {
		err := errors.New(InvalidInputArguments)
		logger.Error(err, "should have returned only one resource metering row")
		return nil, err
	}

	return resourceMetering, nil
}

func (ud UsageData) GetTotalResourceUsageQty(ctx context.Context, resourceId string) (float64, error) {

	logger := log.FromContext(ctx).WithName("UsageData.GetTotalResourceUsageQty")
	getTotalResourceUsageQtyQuery := `SELECT coalesce(SUM(quantity), 0.00) FROM resource_usages WHERE resource_id=$1`
	rows, err := ud.session.QueryContext(ctx, getTotalResourceUsageQtyQuery, resourceId)
	if err != nil {
		logger.Error(err, "failed to get sum of resource usages for resource", "id", resourceId)
		return 0, err
	}

	defer rows.Close()

	var totalResourceUsageQty float64

	if rows.Next() {
		if err := rows.Scan(&totalResourceUsageQty); err != nil {
			logger.Error(err, "failed to get sum of resource usages for resource", "id", resourceId)
			return 0, err
		}
	} else {
		logger.Error(err, "failed to get sum of resource usages for resource", "id", resourceId)
		return 0, err
	}

	if rows.Next() {
		err := errors.New(InvalidInputArguments)
		logger.Error(err, "should have returned only one row")
		return 0, err
	}

	return totalResourceUsageQty, nil
}

func (ud UsageData) UpdateResourceMetering(ctx context.Context, resourceId string, resourceMeteringUpdate *ResourceMeteringUpdate) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateResourceMeteringQuery := "UPDATE resource_metering SET transaction_id=$1, last_recorded=$2 WHERE resource_id=$3"
	_, err = tx.ExecContext(ctx, updateResourceMeteringQuery,
		resourceMeteringUpdate.TransactionId,
		resourceMeteringUpdate.LastRecorded,
		resourceId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ud UsageData) StoreResourceUsage(ctx context.Context,
	resourceUsageCreate *pb.ResourceUsageCreate, creationTime time.Time) (*pb.ResourceUsage, error) {

	logger := log.FromContext(ctx).WithName("UsageData.StoreResourceUsage")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("creating resource usage for resource",
		"id", resourceUsageCreate.ResourceId, "product id", resourceUsageCreate.ProductId)
	id := uuid.NewString()
	expirationTime := creationTime.AddDate(1, 0, 0)

	_, err := ud.session.ExecContext(ctx,
		insertResourceUsageQuery,
		id,
		resourceUsageCreate.CloudAccountId,
		resourceUsageCreate.ResourceId,
		resourceUsageCreate.ResourceName,
		resourceUsageCreate.ProductId,
		resourceUsageCreate.ProductName,
		resourceUsageCreate.TransactionId,
		resourceUsageCreate.Region,
		creationTime,
		// set the expiration for one year from now.
		expirationTime,
		resourceUsageCreate.Quantity,
		resourceUsageCreate.Quantity,
		resourceUsageCreate.Rate,
		resourceUsageCreate.UsageUnitType,
		resourceUsageCreate.StartTime.AsTime(),
		resourceUsageCreate.EndTime.AsTime(),
		false,
	)

	if err != nil {
		return nil, err
	}

	return &pb.ResourceUsage{
		Id:                 id,
		CloudAccountId:     resourceUsageCreate.CloudAccountId,
		ResourceId:         resourceUsageCreate.ResourceId,
		ResourceName:       resourceUsageCreate.ResourceName,
		ProductId:          resourceUsageCreate.ProductId,
		ProductName:        resourceUsageCreate.ProductName,
		Region:             resourceUsageCreate.Region,
		Quantity:           resourceUsageCreate.Quantity,
		UnReportedQuantity: resourceUsageCreate.Quantity,
		Rate:               resourceUsageCreate.Rate,
		UsageUnitType:      resourceUsageCreate.UsageUnitType,
		Timestamp:          timestamppb.New(creationTime),
		StartTime:          resourceUsageCreate.StartTime,
		EndTime:            resourceUsageCreate.EndTime,
		Reported:           false,
	}, nil
}

func (ud UsageData) DeleteResourceUsage(ctx context.Context, resourceUsageRecordId string) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteResourceUsageQuery := "DELETE from resource_usages WHERE id=$1"

	_, err = tx.ExecContext(ctx, deleteResourceUsageQuery, resourceUsageRecordId)

	if err != nil {
		return err
	}

	return tx.Commit()
}

type ResourceUsageUpdate struct {
	UnReportedQuantity float64
}

func (ud UsageData) UpdateResourceUsage(ctx context.Context, resourceUsageRecordId string, resourceUsageUpdate *ResourceUsageUpdate) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateResourceUsageQuery := "UPDATE resource_usages SET unreported_quantity=$1 WHERE id=$2"

	_, err = tx.ExecContext(ctx, updateResourceUsageQuery,
		resourceUsageUpdate.UnReportedQuantity,
		resourceUsageRecordId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ud UsageData) MarkResourceUsageAsReported(ctx context.Context, resourceUsageRecordId string) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateResourceUsageQuery := "UPDATE resource_usages SET reported=$1, unreported_quantity=$2 WHERE id=$3"

	_, err = tx.ExecContext(ctx, updateResourceUsageQuery,
		true, 0, resourceUsageRecordId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ud UsageData) MarkProductUsageAsReported(ctx context.Context, productUsageReportId string) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateResourceUsageQuery := "UPDATE product_usages_report SET reported=$1, unreported_quantity=$2 WHERE id=$3"

	_, err = tx.ExecContext(ctx, updateResourceUsageQuery,
		true, 0, productUsageReportId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

type ProductUsageCreate struct {
	Id             string
	CloudAccountId string
	ProductId      string
	ProductName    string
	Region         string
	Quantity       float64
	Rate           float64
	UsageUnitType  string
	StartTime      time.Time
	EndTime        time.Time
}

func (ud UsageData) StoreProductUsage(ctx context.Context, productUsageCreate *ProductUsageCreate, creationTime time.Time) error {
	logger := log.FromContext(ctx).WithName("UsageData.StoreProductUsageRecord")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("creating product usage for cloud account",
		"id", productUsageCreate.CloudAccountId, "product id", productUsageCreate.ProductId)
	_, err := ud.session.ExecContext(ctx,
		insertProductUsageQuery,
		productUsageCreate.Id,
		productUsageCreate.CloudAccountId,
		productUsageCreate.ProductId,
		productUsageCreate.ProductName,
		productUsageCreate.Region,
		creationTime,
		// set the expiration for one year from now.
		creationTime.AddDate(1, 0, 0),
		productUsageCreate.Quantity,
		productUsageCreate.Rate,
		productUsageCreate.UsageUnitType,
		productUsageCreate.StartTime,
		productUsageCreate.EndTime,
	)

	if err != nil {
		return err
	}

	return nil
}

func (ud UsageData) StoreProductUsageReport(ctx context.Context, productUsageCreate *ProductUsageCreate, productUsageId string) error {
	logger := log.FromContext(ctx).WithName("UsageData.StoreProductUsageReport")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	/**
	INSERT INTO product_usages_report (id, product_usage_id, transaction_id, cloud_account_id, product_id, product_name, region, quantity,
	rate, unreported_quantity, usage_unit_type, timestamp, created_at, reported, start_time, end_time)**/
	logger.Info("creating product usage report for cloud account",
		"id", productUsageCreate.CloudAccountId, "product id", productUsageCreate.ProductId)
	_, err := ud.session.ExecContext(ctx,
		insertProductUsageReportQuery,
		uuid.NewString(),
		productUsageId,
		uuid.NewString(),
		productUsageCreate.CloudAccountId,
		productUsageCreate.ProductId,
		productUsageCreate.ProductName,
		productUsageCreate.Region,
		productUsageCreate.Quantity,
		productUsageCreate.Rate,
		productUsageCreate.Quantity,
		productUsageCreate.UsageUnitType,
		time.Now(),
		time.Now(),
		false,
		productUsageCreate.StartTime,
		productUsageCreate.EndTime,
	)

	if err != nil {
		return err
	}

	return nil
}

type ProductUsageUpdate struct {
	CreationTime time.Time
	Quantity     float64
	StartTime    time.Time
	EndTime      time.Time
}

func (ud UsageData) UpdateProductUsage(ctx context.Context, productUsageId string, productUsageUpdate *ProductUsageUpdate) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateProductUsageQuery := "UPDATE product_usages SET creation=$1, expiration=$2, quantity=$3, start_time=$4 , end_time=$5 WHERE id=$6"

	_, err = tx.ExecContext(ctx, updateProductUsageQuery,
		productUsageUpdate.CreationTime,
		// set the expiration for one year from now.
		productUsageUpdate.CreationTime.AddDate(1, 0, 0),
		productUsageUpdate.Quantity,
		productUsageUpdate.StartTime,
		productUsageUpdate.EndTime,
		productUsageId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ud UsageData) UpdateProductUsageReport(ctx context.Context, reportProductUsageUpdate *pb.ReportProductUsageUpdate) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateProductUsageReportQuery := "UPDATE product_usages_report SET unreported_quantity=$1 WHERE id=$2"

	_, err = tx.ExecContext(ctx, updateProductUsageReportQuery,
		reportProductUsageUpdate.UnReportedQuantity,
		reportProductUsageUpdate.ProductUsageReportId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ud UsageData) DeleteProductUsage(ctx context.Context, productUsageId string) error {

	tx, err := ud.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteProductUsageQuery := "DELETE from product_usages WHERE id=$1"

	_, err = tx.ExecContext(ctx, deleteProductUsageQuery, productUsageId)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ud UsageData) GetResourceUsageById(ctx context.Context, resourceUsageId string) (*pb.ResourceUsage, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetResourceUsageById")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("getting resource usage by", "id", resourceUsageId)

	const getResourceUsageQuery = `
		SELECT id, cloud_account_id, resource_id, resource_name, product_id, product_name, transaction_id, region, creation, quantity, unreported_quantity, 
		rate, usage_unit_type, start_time, end_time, reported FROM resource_usages WHERE id=$1
	`

	rows, err := ud.session.QueryContext(ctx, getResourceUsageQuery, resourceUsageId)
	if err != nil {
		logger.Error(err, "failed to search resource usages using", "id", resourceUsageId)
		return nil, err
	}

	defer rows.Close()

	resourceUsage := &pb.ResourceUsage{}

	timestamp := time.Time{}
	startTime := time.Time{}
	endTime := time.Time{}

	if rows.Next() {
		if err := rows.Scan(&resourceUsage.Id, &resourceUsage.CloudAccountId,
			&resourceUsage.ResourceId, &resourceUsage.ResourceName, &resourceUsage.ProductId,
			&resourceUsage.ProductName, &resourceUsage.TransactionId, &resourceUsage.Region, &timestamp,
			&resourceUsage.Quantity, &resourceUsage.UnReportedQuantity, &resourceUsage.Rate,
			&resourceUsage.UsageUnitType, &startTime,
			&endTime, &resourceUsage.Reported); err != nil {

			logger.Error(err, "failed to read resource usage row")
			return nil, err
		}
	} else {
		logger.Error(err, "failed to read resource usage row")
		return nil, err
	}

	resourceUsage.Timestamp = timestamppb.New(timestamp)
	resourceUsage.StartTime = timestamppb.New(startTime)
	resourceUsage.EndTime = timestamppb.New(endTime)

	if rows.Next() {
		err := errors.New(InvalidInputArguments)
		logger.Error(err, "should have returned only one usage row")
		return nil, err
	}

	resourceUsage.Timestamp = timestamppb.New(timestamp)

	return resourceUsage, nil
}

// Return it as time series with the last one first.
func (ud UsageData) GetResourceUsagesForResource(ctx context.Context, resourceId string) (*pb.ResourceUsages, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetResourceUsagesForResource")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("getting resource usage by resource", "id", resourceId)

	var queryBuilder strings.Builder

	_, err := queryBuilder.WriteString(searchResourceUsagesQueryBase)
	if err != nil {
		logger.Error(err, "failed to build query for searching resource usages")
		return nil, err
	}

	_, err = queryBuilder.WriteString(" WHERE resource_id=$1 ORDER BY creation DESC")
	if err != nil {
		logger.Error(err, "failed to append resource id to search query for searching resource usages")
		return nil, err
	}

	rows, err := ud.session.QueryContext(ctx, queryBuilder.String(), resourceId)
	if err != nil {
		logger.Error(err, "failed to search resource usages for resource id", "id", resourceId)
		return nil, err
	}

	defer rows.Close()

	resourceUsages := &pb.ResourceUsages{}

	for rows.Next() {
		resourceUsage := &pb.ResourceUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}
		if err := rows.Scan(&resourceUsage.Id, &resourceUsage.CloudAccountId,
			&resourceUsage.ResourceId, &resourceUsage.ResourceName, &resourceUsage.ProductId,
			&resourceUsage.ProductName, &resourceUsage.TransactionId, &resourceUsage.Region, &timestamp,
			&resourceUsage.Quantity, &resourceUsage.UnReportedQuantity, &resourceUsage.Rate,
			&resourceUsage.UsageUnitType, &startTime,
			&endTime, &resourceUsage.Reported); err != nil {

			logger.Error(err, "failed to read resource usage row")
			continue

		}

		resourceUsage.Timestamp = timestamppb.New(timestamp)
		resourceUsage.StartTime = timestamppb.New(startTime)
		resourceUsage.EndTime = timestamppb.New(endTime)
		resourceUsages.ResourceUsages = append(resourceUsages.ResourceUsages, resourceUsage)
	}

	return resourceUsages, nil
}

func (ud UsageData) SearchResourceUsages(ctx context.Context, resourceUsagesFilter *pb.ResourceUsagesFilter) (*pb.ResourceUsages, error) {
	logger := log.FromContext(ctx).WithName("UsageData.SearchResourceUsages")
	//logger.V(9).Info("BEGIN")
	//defer logger.Info("END")

	queryBuilder, params, err := ud.GetResourceUsageQueryAndParams(ctx, resourceUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get query and parameters")
		return nil, err
	}

	rows, err := ud.session.QueryContext(ctx, queryBuilder.String(), params...)
	if err != nil {
		logger.Error(err, "failed to search resource usages")
		return nil, err
	}

	defer rows.Close()

	resourceUsages := &pb.ResourceUsages{}

	for rows.Next() {
		resourceUsage := &pb.ResourceUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&resourceUsage.Id, &resourceUsage.CloudAccountId,
			&resourceUsage.ResourceId, &resourceUsage.ResourceName, &resourceUsage.ProductId,
			&resourceUsage.ProductName, &resourceUsage.TransactionId, &resourceUsage.Region, &timestamp,
			&resourceUsage.Quantity, &resourceUsage.UnReportedQuantity, &resourceUsage.Rate,
			&resourceUsage.UsageUnitType, &startTime,
			&endTime, &resourceUsage.Reported); err != nil {

			logger.Error(err, "failed to read resource usage row")
			continue

		}

		resourceUsage.Timestamp = timestamppb.New(timestamp)
		resourceUsage.StartTime = timestamppb.New(startTime)
		resourceUsage.EndTime = timestamppb.New(endTime)
		resourceUsages.ResourceUsages = append(resourceUsages.ResourceUsages, resourceUsage)
	}

	return resourceUsages, nil
}

func (UsageData) GetResourceUsageQueryAndParams(ctx context.Context, resourceUsagesFilter *pb.ResourceUsagesFilter) (strings.Builder, []interface{}, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetResourceUsageQueryAndParams")
	var queryBuilder strings.Builder

	_, err := queryBuilder.WriteString(searchResourceUsagesQueryBase)
	if err != nil {
		logger.Error(err, "failed to build query for searching resource usages")
		return strings.Builder{}, nil, err
	}

	if _, err := queryBuilder.WriteString(" WHERE 1=1"); err != nil {
		logger.Error(err, "failed to build query for searching resource usages")
		return strings.Builder{}, nil, err
	}

	params := []interface{}{}

	if resourceUsagesFilter.ResourceId != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND resource_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append resource id to search query for searching resource usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceUsagesFilter.ResourceId)
	}

	if resourceUsagesFilter.CloudAccountId != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append cloud account id to search query for searching resource usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceUsagesFilter.CloudAccountId)
	}

	if resourceUsagesFilter.Region != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND region=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append region to search query for searching resource usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceUsagesFilter.Region)
	}

	if resourceUsagesFilter.Reported != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND reported=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append reported to search query for searching resource usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceUsagesFilter.Reported)
	}

	if resourceUsagesFilter.StartTime != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf("  AND start_time>=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append start time to search query for searching resource usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceUsagesFilter.StartTime.AsTime())
	}

	if resourceUsagesFilter.EndTime != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf("  AND end_time<=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append end time to search query for searching resource usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceUsagesFilter.EndTime.AsTime())
	}

	_, err = queryBuilder.WriteString(" ORDER BY creation ASC")
	if err != nil {
		logger.Error(err, "failed to append order by to search query for searching resource usages")
		return strings.Builder{}, nil, err
	}
	return queryBuilder, params, nil
}

func (ud UsageData) SendResourceUsages(ctx context.Context, resourceUsagesFilter *pb.ResourceUsagesFilter, resourceUsagesStream pb.UsageService_StreamSearchResourceUsagesServer) error {
	logger := log.FromContext(ctx).WithName("UsageData.SendResourceUsages")
	queryBuilder, params, err := ud.GetResourceUsageQueryAndParams(ctx, resourceUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get query and parameters")
		return err
	}

	rows, err := ud.session.QueryContext(ctx, queryBuilder.String(), params...)
	if err != nil {
		logger.Error(err, "failed to search resource usages")
		return err
	}

	defer rows.Close()
	for rows.Next() {
		resourceUsage := pb.ResourceUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&resourceUsage.Id, &resourceUsage.CloudAccountId,
			&resourceUsage.ResourceId, &resourceUsage.ResourceName, &resourceUsage.ProductId,
			&resourceUsage.ProductName, &resourceUsage.TransactionId, &resourceUsage.Region, &timestamp,
			&resourceUsage.Quantity, &resourceUsage.UnReportedQuantity, &resourceUsage.Rate,
			&resourceUsage.UsageUnitType, &startTime,
			&endTime, &resourceUsage.Reported); err != nil {

			logger.Error(err, "failed to read resource usage row")
			continue

		}

		resourceUsage.Timestamp = timestamppb.New(timestamp)
		resourceUsage.StartTime = timestamppb.New(startTime)
		resourceUsage.EndTime = timestamppb.New(endTime)
		if err := resourceUsagesStream.Send(&resourceUsage); err != nil {
			logger.Error(err, "error sending resource usage")
			return err
		}

	}
	return nil

}

func (ud UsageData) GetProductUsageById(ctx context.Context, productUsageId string) (*pb.ProductUsage, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetProductUsageById")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("getting product usage by", "id", productUsageId)

	getProductUsageQuery := `
		SELECT id, cloud_account_id, product_id, product_name, region, creation, quantity,  
		rate, usage_unit_type, start_time, end_time FROM product_usages WHERE id=$1
	`

	rows, err := ud.session.QueryContext(ctx, getProductUsageQuery, productUsageId)
	if err != nil {
		logger.Error(err, "failed to search product usages using", "id", productUsageId)
		return nil, err
	}

	defer rows.Close()

	productUsage := &pb.ProductUsage{}

	timestamp := time.Time{}
	startTime := time.Time{}
	endTime := time.Time{}

	if rows.Next() {
		if err := rows.Scan(&productUsage.Id, &productUsage.CloudAccountId,
			&productUsage.ProductId,
			&productUsage.ProductName, &productUsage.Region, &timestamp,
			&productUsage.Quantity, &productUsage.Rate,
			&productUsage.UsageUnitType, &startTime, &endTime); err != nil {

			logger.Error(err, "failed to read product usage row")
			return nil, err
		}
	} else {
		logger.Error(err, "failed to read product usage row")
		return nil, err
	}

	if rows.Next() {
		err := errors.New(InvalidInputArguments)
		logger.Error(err, "should have returned only one product usage row")
		return nil, err
	}

	productUsage.Timestamp = timestamppb.New(timestamp)
	productUsage.StartTime = timestamppb.New(startTime)
	productUsage.EndTime = timestamppb.New(endTime)

	return productUsage, nil
}

func (ud UsageData) GetProductUsageByProduct(ctx context.Context, productId string) (*pb.ProductUsages, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetProductUsageByProduct")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("getting product usage by product", "id", productId)

	getProductUsageQuery := `
		SELECT id, cloud_account_id, product_id, product_name, region, creation, quantity,  
		rate, usage_unit_type, start_time, end_time FROM product_usages WHERE product_id=$1 ORDER BY creation DESC
	`

	rows, err := ud.session.QueryContext(ctx, getProductUsageQuery, productId)
	if err != nil {
		logger.Error(err, "failed to search product usages using product", "id", productId)
		return nil, err
	}

	defer rows.Close()

	productUsages := &pb.ProductUsages{}

	for rows.Next() {
		productUsage := &pb.ProductUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&productUsage.Id, &productUsage.CloudAccountId,
			&productUsage.ProductId,
			&productUsage.ProductName, &productUsage.Region, &timestamp,
			&productUsage.Quantity, &productUsage.Rate,
			&productUsage.UsageUnitType, &startTime, &endTime); err != nil {

			logger.Error(err, "failed to read product usage row")
			continue

		}

		productUsage.Timestamp = timestamppb.New(timestamp)
		productUsage.StartTime = timestamppb.New(startTime)
		productUsage.EndTime = timestamppb.New(endTime)
		productUsages.ProductUsages = append(productUsages.ProductUsages, productUsage)
	}

	return productUsages, nil
}

func (ud UsageData) SearchProductUsages(ctx context.Context, productUsagesFilter *pb.ProductUsagesFilter) (*pb.ProductUsages, error) {
	logger := log.FromContext(ctx).WithName("UsageData.SearchProductUsages")
	//logger.V(9).Info("BEGIN")
	//defer logger.Info("END")

	queryBuilder, params, err := ud.GetProductUsageQueryAndParams(ctx, productUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get product usages query and parameters")
		return nil, err
	}

	rows, err := ud.session.QueryContext(ctx, queryBuilder.String(), params...)
	if err != nil {
		logger.Error(err, "failed to search product usages")
		return nil, err
	}

	defer rows.Close()

	productUsages := &pb.ProductUsages{}

	for rows.Next() {
		productUsage := &pb.ProductUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&productUsage.Id, &productUsage.CloudAccountId,
			&productUsage.ProductId,
			&productUsage.ProductName, &productUsage.Region, &timestamp,
			&productUsage.Quantity, &productUsage.Rate,
			&productUsage.UsageUnitType, &startTime, &endTime); err != nil {

			logger.Error(err, "failed to read product usage row")
			continue

		}

		productUsage.Timestamp = timestamppb.New(timestamp)
		productUsage.StartTime = timestamppb.New(startTime)
		productUsage.EndTime = timestamppb.New(endTime)
		productUsages.ProductUsages = append(productUsages.ProductUsages, productUsage)
	}

	return productUsages, nil
}

func (UsageData) GetProductUsageQueryAndParams(ctx context.Context, productUsagesFilter *pb.ProductUsagesFilter) (strings.Builder, []interface{}, error) {
	logger := log.FromContext(ctx).WithName("UsageData.GetProductUsageQueryAndParams")
	var queryBuilder strings.Builder

	_, err := queryBuilder.WriteString(searchProductUsagesQueryBase)
	if err != nil {
		logger.Error(err, "failed to build query for searching product usages")
		return strings.Builder{}, nil, err
	}

	if _, err := queryBuilder.WriteString(" WHERE 1=1"); err != nil {
		logger.Error(err, "failed to build query for searching product usages")
		return strings.Builder{}, nil, err
	}

	params := []interface{}{}

	if productUsagesFilter.ProductId != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND product_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append product id to search query for searching product usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, productUsagesFilter.ProductId)
	}

	if productUsagesFilter.CloudAccountId != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append cloud account id to search query for searching product usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, productUsagesFilter.CloudAccountId)
	}

	if productUsagesFilter.Region != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND region=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append region to search query for searching product usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, productUsagesFilter.Region)
	}

	if productUsagesFilter.StartTime != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf("  AND start_time>=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append start time to search query for searching product usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, productUsagesFilter.StartTime.AsTime())
	}

	if productUsagesFilter.EndTime != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf("  AND end_time<=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append end time to search query for searching product usages")
			return strings.Builder{}, nil, err
		}
		params = append(params, productUsagesFilter.EndTime.AsTime())
	}

	_, err = queryBuilder.WriteString(" ORDER BY creation ASC")
	if err != nil {
		logger.Error(err, "failed to append order by to search query for searching product usages")
		return strings.Builder{}, nil, err
	}
	return queryBuilder, params, nil
}

func (ud UsageData) SendProductUsages(ctx context.Context, productUsagesFilter *pb.ProductUsagesFilter, productUsagesStream pb.UsageService_StreamSearchProductUsagesServer) error {
	logger := log.FromContext(ctx).WithName("UsageData.SearchProductUsages")

	queryBuilder, params, err := ud.GetProductUsageQueryAndParams(ctx, productUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get product usages query and parameters")
		return err
	}

	rows, err := ud.session.QueryContext(ctx, queryBuilder.String(), params...)
	if err != nil {
		logger.Error(err, "failed to search product usages")
		return err
	}

	defer rows.Close()

	for rows.Next() {
		productUsage := pb.ProductUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&productUsage.Id, &productUsage.CloudAccountId,
			&productUsage.ProductId,
			&productUsage.ProductName, &productUsage.Region, &timestamp,
			&productUsage.Quantity, &productUsage.Rate,
			&productUsage.UsageUnitType, &startTime, &endTime); err != nil {

			logger.Error(err, "failed to read product usage row")
			continue

		}

		productUsage.Timestamp = timestamppb.New(timestamp)
		productUsage.StartTime = timestamppb.New(startTime)
		productUsage.EndTime = timestamppb.New(endTime)
		if err := productUsagesStream.Send(&productUsage); err != nil {
			logger.Error(err, "error sending product usage")
			return err
		}

	}

	return nil
}

func (ud UsageData) DeleteAllUsages(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("UsageData.DeleteAllUsages")
	//logger.V(9).Info("BEGIN")
	//defer logger.Info("END")

	const deleteResourceUsages = `DELETE FROM resource_usages`
	const deleteResourceMetering = `DELETE FROM resource_metering`
	const deleteProductUsages = `DELETE FROM product_usages`
	const deleteProductUsagesReport = `DELETE FROM product_usages_report`

	_, err := ud.session.ExecContext(ctx, deleteResourceUsages)

	if err != nil {
		logger.Error(err, "error deleting resource_usages")
		return err
	}

	_, err = ud.session.ExecContext(ctx, deleteProductUsages)

	if err != nil {
		logger.Error(err, "error deleting product_usages")
		return err
	}

	_, err = ud.session.ExecContext(ctx, deleteResourceMetering)

	if err != nil {
		logger.Error(err, "error deleting resource metering records")
		return err
	}

	_, err = ud.session.ExecContext(ctx, deleteProductUsagesReport)

	if err != nil {
		logger.Error(err, "error deleting product usages report")
		return err
	}

	return nil
}

func (ud UsageData) SearchProductUsagesReport(productUsagesReportFilter *pb.ProductUsagesReportFilter, rs pb.UsageService_SearchProductUsagesReportServer) error {

	logger := log.FromContext(rs.Context()).WithName("UsageData.SearchProductUsagesReport")

	getProductUsageUnreportedQueryBase := `
		SELECT id, product_usage_id, cloud_account_id, product_id, product_name, quantity, rate,  
		unreported_quantity, usage_unit_type, timestamp, start_time, end_time,
		reported FROM product_usages_report 
	`

	var queryBuilder strings.Builder

	_, err := queryBuilder.WriteString(getProductUsageUnreportedQueryBase)
	if err != nil {
		logger.Error(err, "failed to build query for searching product usage report")
		return err
	}

	if _, err := queryBuilder.WriteString(" WHERE 1=1"); err != nil {
		logger.Error(err, "failed to build query for searching product usage report")
		return err
	}

	params := []interface{}{}

	if productUsagesReportFilter.ProductId != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND product_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append product id to search query for searching product usage report")
			return err
		}
		params = append(params, productUsagesReportFilter.ProductId)
	}

	if productUsagesReportFilter.CloudAccountId != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append cloud account id to search query for searching product usage report")
			return nil
		}
		params = append(params, productUsagesReportFilter.CloudAccountId)
	}

	if productUsagesReportFilter.Region != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND region=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append region to search query for searching product usages report")
			return err
		}
		params = append(params, productUsagesReportFilter.Region)
	}

	if productUsagesReportFilter.StartTime != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf("  AND start_time>=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append start time to search query for searching product usages report")
			return err
		}
		params = append(params, productUsagesReportFilter.StartTime.AsTime())
	}

	if productUsagesReportFilter.EndTime != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf("  AND end_time<=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append end time to search query for searching product usages report")
			return err
		}
		params = append(params, productUsagesReportFilter.EndTime.AsTime())
	}

	if productUsagesReportFilter.Reported != nil {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND reported=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append reported to search query for searching product usages report")
			return err
		}
		params = append(params, productUsagesReportFilter.Reported)
	}

	_, err = queryBuilder.WriteString(" ORDER BY created_at ASC")
	if err != nil {
		logger.Error(err, "failed to append order by to search query for searching product usages report")
		return err
	}

	var query string
	query = queryBuilder.String()
	logger.Info(query)
	rows, err := ud.session.QueryContext(rs.Context(), queryBuilder.String(), params...)
	if err != nil {
		logger.Error(err, "failed to search product usages report")
		return err
	}

	defer rows.Close()

	for rows.Next() {
		reportProductUsage := &pb.ReportProductUsage{}

		timestamp := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&reportProductUsage.Id, &reportProductUsage.ProductUsageId, &reportProductUsage.CloudAccountId,
			&reportProductUsage.ProductId,
			&reportProductUsage.ProductName, &reportProductUsage.Quantity, &reportProductUsage.Rate, &reportProductUsage.UnReportedQuantity,
			&reportProductUsage.UsageUnitType, &timestamp,
			&startTime, &endTime,
			&reportProductUsage.Reported); err != nil {

			logger.Error(err, "failed to read product usage row for searching unreported product usages")
			continue

		}

		reportProductUsage.Timestamp = timestamppb.New(timestamp)
		reportProductUsage.StartTime = timestamppb.New(startTime)
		reportProductUsage.EndTime = timestamppb.New(endTime)

		if err := rs.Send(reportProductUsage); err != nil {
			logger.Error(err, "error sending product usage for searching unreported product usages")
		}
	}

	return nil
}
