// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package instance

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// List instances as a stream.
// This returns all non-deleted instances as messages with WatchDeltaType=Updated,
// followed by a single WatchDeltaType=Bookmark with the last-seen resourceVersion.
// Used by Instance Replicator.
// This is modeled after https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes.
func (s *InstanceService) SearchStreamPrivate(req *pb.InstanceSearchStreamPrivateRequest, svc pb.InstancePrivateService_SearchStreamPrivateServer) error {
	ctx := svc.Context()
	log := log.FromContext(ctx).WithName("InstanceService.SearchStreamPrivate")
	err := func() error {
		log.Info("Request", logkeys.Request, req)

		maxResourceVersionAtStart, err := s.getMaximumResourceVersion(ctx)
		if err != nil {
			return err
		}
		log.Info("Starting", logkeys.MaxResourceVersionAtStart, maxResourceVersionAtStart)

		selectSql := fmt.Sprintf("select %s from instance", s.sqlTransformer.ColumnsForFromRow())
		// Do not send deleted records.
		query := selectSql + " where resource_version <= $1 and deleted_timestamp = $2"
		rows, err := s.db.QueryContext(ctx, query, maxResourceVersionAtStart, common.TimestampInfinityStr)
		if err != nil {
			return err
		}
		defer rows.Close()
		if err := s.watchSendRows(ctx, svc, rows); err != nil {
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if err := s.watchSendBookmark(ctx, svc, maxResourceVersionAtStart); err != nil {
			return err
		}
		return nil
	}()
	if err != nil && err != context.Canceled {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Completed")
	}
	return err
}

// Return a stream of changes to instances using messages with WatchDeltaType=Updated or Deleted.
// Messages with WatchDeltaType=Bookmark and the last-seen resourceVersion will be sent periodically.
// Used by Instance Replicator.
// This is modeled after https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes.
// This polls Postgres periodically to find records with a resource_version greater than the maximum from the previous iteration.
// Based on https://github.com/k3s-io/kine/blob/27bd5e740946e0f1e9faeb83d594fb854180a1d4/pkg/logstructured/sqllog/sql.go#L376-L482
func (s *InstanceService) Watch(req *pb.InstanceWatchRequest, svc pb.InstancePrivateService_WatchServer) error {
	ctx := svc.Context()
	log := log.FromContext(ctx).WithName("InstanceService.Watch")
	err := func() error {
		log.Info("Request", logkeys.Request, req)

		// Validate input.
		if req.ResourceVersion == "" {
			return status.Error(codes.InvalidArgument, "missing resource version")
		}

		// Resource version was converted by rowToInstancePrivate from an integer in the database to a string.
		// Convert it back to an integer so we can compare it.
		afterResourceVersion, err := strconv.ParseInt(req.ResourceVersion, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid resource version: %w", err)
		}

		pollInterval := 1 * time.Second
		wait := time.NewTicker(pollInterval)
		defer wait.Stop()

		selectSql := fmt.Sprintf("select %s from instance", s.sqlTransformer.ColumnsForFromRow())
		// Use a Prepared statement for repeated use. It avoid recreating the statement every time.
		// Note: A prepared statement is created on a connection with DB. So every execution will attempt
		// to use a same connection. If the connection is not present a new prepared statement is created.
		stmt, err := s.db.PrepareContext(ctx, selectSql+" where resource_version > $1 and resource_version <= $2")
		if err != nil {
			return err
		}
		defer stmt.Close()

		// For each iteration, send all records (including deleted records) updated since the last iteration.
		for {
			maxResourceVersion, err := s.getMaximumResourceVersion(ctx)
			if err != nil {
				return err
			}

			// Use an anonymous function to ensure rows.Close is invoked immediately.
			if queryError := func() error {
				rows, err := stmt.QueryContext(ctx, afterResourceVersion, maxResourceVersion)
				if err != nil {
					return err
				}
				defer rows.Close()
				if err := s.watchSendRows(ctx, svc, rows); err != nil {
					return err
				}
				if err := rows.Err(); err != nil {
					return err
				}
				return nil
			}(); queryError != nil {
				// return an error and break from the outer for loop
				return queryError
			}

			// Send a Bookmark message at every iteration, even if there were no new records.
			// This prevents the response stream from being detected as idle by the Instance Replicator.
			if err := s.watchSendBookmark(ctx, svc, maxResourceVersion); err != nil {
				return err
			}

			// Start next iteration beyond maxResourceVersion.
			afterResourceVersion = maxResourceVersion

			// Sleep for a short time or until context is cancelled.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-wait.C:
			}
		}
	}()
	if err != nil && err != context.Canceled {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Completed")
	}
	return err
}

func (s *InstanceService) getMaximumResourceVersion(ctx context.Context) (int64, error) {
	// Get max resourceVersion over entire table.
	var row *sql.Row
	if row = s.db.QueryRowContext(ctx, "select coalesce(max(resource_version), 0) from instance"); row.Err() != nil {
		return 0, row.Err()
	}
	var resourceVersion int64
	if err := row.Scan(&resourceVersion); err != nil {
		return 0, err
	}
	return resourceVersion, nil
}

// Read SQL rows and send to Watch client.
// Note: rows lifecycle is not not managed by this function.
func (s *InstanceService) watchSendRows(ctx context.Context, svc pb.InstancePrivateService_WatchServer, rows *sql.Rows) error {
	log := log.FromContext(ctx).WithName("InstanceService.watchSendRows")
	recordCount := int64(0)
	for rows.Next() {
		instance, err := s.sqlTransformer.FromRow(ctx, rows)
		if err != nil {
			return err
		}
		resp := pb.InstanceWatchResponse{
			Type:   pb.WatchDeltaType_Updated,
			Object: instance,
		}
		if instance.Metadata.DeletedTimestamp != nil {
			resp.Type = pb.WatchDeltaType_Deleted
		}
		if err := svc.Send(&resp); err != nil {
			return err
		}
		recordCount++
	}
	level := 0
	if recordCount == 0 {
		level = 9
	}
	log.V(level).Info("Statistics", logkeys.RecordCount, recordCount)
	return nil
}

func (s *InstanceService) watchSendBookmark(ctx context.Context, svc pb.InstancePrivateService_WatchServer, resourceVersion int64) error {
	resp := pb.InstanceWatchResponse{
		Type: pb.WatchDeltaType_Bookmark,
		Object: &pb.InstancePrivate{
			Metadata: &pb.InstanceMetadataPrivate{
				ResourceVersion: fmt.Sprintf("%d", resourceVersion),
			},
		},
	}
	if err := svc.Send(&resp); err != nil {
		return err
	}
	return nil
}
