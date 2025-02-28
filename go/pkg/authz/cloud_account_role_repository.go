// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CloudAccountRoleRepository struct {
	db *sql.DB
}

func NewCloudAccountRoleRepository(db *sql.DB) (*CloudAccountRoleRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &CloudAccountRoleRepository{
		db: db,
	}, nil
}

func (p CloudAccountRoleRepository) GetCloudAccountRoles(ctx context.Context, decisionId string, cloudAccountId string, subject string, resourceType string, resourceId string, action string) ([]*pb.CloudAccountRole, error) {
	var rows *sql.Rows
	var err error
	actionJSON := json.RawMessage(`"` + action + `"`)
	subjectJSON := json.RawMessage(`"` + subject + `"`)

	query := `
	SELECT c.id,effect
	FROM cloud_account_roles c
	LEFT JOIN permissions p
	ON c.id = p.cloud_account_role_id
	WHERE c.cloud_account_id = $1 
	 AND p.resource_type = $2
	 AND (p.resource_id = $3 OR p.resource_id = '*')
	 AND p.actions @> $4::jsonb
	 AND users @> $5::jsonb;`

	rows, err = p.db.Query(query, cloudAccountId, resourceType, resourceId, actionJSON, subjectJSON)
	if err != nil {
		logger.Error(err, "error executing db query", "query", query)
		return nil, err
	}

	defer rows.Close()

	cloudAccountRoles := []*pb.CloudAccountRole{}
	for rows.Next() {

		resp := pb.CloudAccountRole{}
		var effect string
		if err := rows.Scan(&resp.Id, &effect); err != nil {
			logger.Error(err, "error scanning rows")
			return nil, err
		}

		resp.Effect = pb.CloudAccountRole_Effect(pb.CloudAccountRole_Effect_value[effect])
		if resp.Effect.String() == pb.CloudAccountRole_deny.Enum().String() {
			logger.V(9).Info("found deny permission", "permissionId", resp.Id)
			return nil, err
		}
		cloudAccountRoles = append(cloudAccountRoles, &resp)
	}
	return cloudAccountRoles, nil
}

func (p CloudAccountRoleRepository) CreateCloudAccountRole(ctx context.Context, cloudAccountRole *pb.CloudAccountRole) (*pb.CloudAccountRole, error) {
	tx, err := db.Begin()
	if err != nil {
		logger.Error(err, "error initializing db transaction")
		return nil, status.Errorf(codes.Internal, "error, while creating cloudAccountRole")
	}

	query := "INSERT INTO cloud_account_roles (cloud_account_id, alias, effect, users) VALUES ($1, $2, $3, $4) RETURNING id"
	var id string
	if cloudAccountRole.Users == nil {
		cloudAccountRole.Users = []string{}
	}
	err = tx.QueryRowContext(ctx, query, cloudAccountRole.CloudAccountId, cloudAccountRole.Alias, cloudAccountRole.Effect.String(), cloudAccountRole.Users).Scan(&id)
	if err != nil {
		logger.Error(err, "error inserting cloudAccountRole", "cloudAccountRole", cloudAccountRole)
		rollbackError := tx.Rollback()
		if rollbackError != nil {
			logger.Error(rollbackError, "there was an error performing the rollback")
		}
		if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" {
			logger.Error(err, "error duplicate alias")
			return nil, status.Error(codes.AlreadyExists, "error, creating cloudAccountRole (alias) already exist")
		}
		return nil, status.Errorf(codes.Internal, "error, while creating cloudAccountRole")
	}
	cloudAccountRole.Id = id

	// creating permissions for the cloudAccountRole
	for _, permission := range cloudAccountRole.Permissions {
		query := "INSERT INTO permissions (cloud_account_role_id, cloud_account_id, resource_type, resource_id, actions) VALUES ($1, $2, $3, $4, $5) RETURNING id"
		var permissionId string
		if permission.Actions == nil {
			permission.Actions = []string{}
		}

		err = tx.QueryRowContext(ctx, query, id, cloudAccountRole.CloudAccountId, permission.ResourceType, permission.ResourceId, permission.Actions).Scan(&permissionId)
		if err != nil {
			logger.Error(err, "error inserting permission", "permission", permission)
			rollbackError := tx.Rollback()
			if rollbackError != nil {
				logger.Error(rollbackError, "there was an error performing the rollback")
			}
			if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" {
				logger.Error(err, "error duplicate permission")
				return nil, status.Error(codes.AlreadyExists, "error, creating permission (id,type) already exist")
			}
			return nil, status.Errorf(codes.Internal, "error, while creating permission")
		}
		permission.Id = permissionId
	}
	err = tx.Commit()
	if err != nil {
		logger.Error(err, "error while creating cloudAccountRole ", "cloudAccountRole", cloudAccountRole)
		return nil, status.Errorf(codes.Internal, "error, while creating cloudAccountRole")
	}
	logger.V(9).Info("cloudAccountRole created successfully", "id", id)
	return cloudAccountRole, nil
}

func (p CloudAccountRoleRepository) ListCloudAccountRoles(ctx context.Context, cloudAccountRoleQuery *pb.CloudAccountRoleQuery, size uint32, offset uint32) (cloudAccountRoles []*pb.CloudAccountRole, err error) {
	query := `
	SELECT c.id, c.cloud_account_id,
	alias, effect, users, COALESCE(json_agg(json_build_object('id', p.id, 'resourceId', p.resource_id, 'resourceType', p.resource_type, 'actions', p.actions, 'createdAt', p.created_at, 'updatedAt', p.updated_at)) FILTER (WHERE p.id IS NOT NULL), '[]') AS permissions
	FROM cloud_account_roles c
	LEFT JOIN
	permissions p
	ON
	c.id = p.cloud_account_role_id
	WHERE c.cloud_account_id = $1
	`
	args := []interface{}{cloudAccountRoleQuery.CloudAccountId}
	paramIndex := 2

	if cloudAccountRoleQuery.ResourceType != nil {
		query += " AND p.resource_type = $" + strconv.Itoa(paramIndex)
		args = append(args, *cloudAccountRoleQuery.ResourceType)
		paramIndex++
	}

	if cloudAccountRoleQuery.UserId != nil {
		query += " AND c.users @> $" + strconv.Itoa(paramIndex) + "::jsonb"
		args = append(args, `"`+*cloudAccountRoleQuery.UserId+`"`)
		paramIndex++
	}

	query += " GROUP BY c.id ORDER BY c.created_at DESC"

	if size > 0 {
		query += " LIMIT $" + strconv.Itoa(paramIndex) + " OFFSET $" + strconv.Itoa(paramIndex+1)
		args = append(args, size, offset)
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error(err, "error executing db query", "query", query)
		return nil, status.Error(codes.Internal, "failed to list cloudAccountRoles")
	}
	defer rows.Close()

	for rows.Next() {
		cloudAccountRole := pb.CloudAccountRole{}
		var permissionsJson []byte
		var usersJson []byte

		var effect string
		if err := rows.Scan(&cloudAccountRole.Id, &cloudAccountRole.CloudAccountId, &cloudAccountRole.Alias, &effect, &usersJson, &permissionsJson); err != nil {
			logger.Error(err, "error scanning rows")
			return nil, status.Errorf(codes.Internal, "error scanning rows")
		}

		cloudAccountRole.Effect = pb.CloudAccountRole_Effect(pb.CloudAccountRole_Effect_value[effect])

		if usersJson != nil {
			cloudAccountRole.Users = []string{}
			if err := json.Unmarshal(usersJson, &cloudAccountRole.Users); err != nil {
				logger.Error(err, "error unmarshalling users json")
				return nil, status.Error(codes.Internal, "failed to process user data")
			}
		}

		if permissionsJson != nil {
			cloudAccountRole.Permissions = []*pb.CloudAccountRole_Permission{}
			if err := json.Unmarshal(permissionsJson, &cloudAccountRole.Permissions); err != nil {
				logger.Error(err, "error unmarshalling permissions json")
				return nil, status.Error(codes.Internal, "failed to process permission data")
			}
		}
		cloudAccountRoles = append(cloudAccountRoles, &cloudAccountRole)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err, "error iterating over rows")
		return nil, status.Error(codes.Internal, "failed to process cloudAccountRole list")
	}
	logger.V(9).Info("cloudAccountRoles listed successfully", "count", len(cloudAccountRoles))
	return cloudAccountRoles, nil
}

func (p CloudAccountRoleRepository) GetCloudAccountRole(ctx context.Context, cloudAccountRoleId *pb.CloudAccountRoleId) (*pb.CloudAccountRole, error) {
	query := `SELECT  id, cloud_account_id, alias, effect, users 
	FROM cloud_account_roles
	WHERE id = $1 AND cloud_account_id = $2`

	// Execute the query with the provided cloudAccountRoleId and cloudAccountId.
	row := p.db.QueryRowContext(ctx, query, cloudAccountRoleId.Id, cloudAccountRoleId.CloudAccountId)

	cloudAccountRole := pb.CloudAccountRole{}
	var effect string
	var usersJson []byte

	if err := row.Scan(&cloudAccountRole.Id, &cloudAccountRole.CloudAccountId, &cloudAccountRole.Alias, &effect, &usersJson); err != nil {
		if err == sql.ErrNoRows {
			logger.Error(err, "cloud account role not found")
			return nil, status.Errorf(codes.NotFound, "cloud account role not found")
		}
		logger.Error(err, "error while getting cloud account role")
		return nil, status.Errorf(codes.Internal, "error while getting cloud account role")
	}

	cloudAccountRole.Effect = pb.CloudAccountRole_Effect(pb.CloudAccountRole_Effect_value[effect])

	if usersJson != nil {
		cloudAccountRole.Users = []string{}
		if err := json.Unmarshal(usersJson, &cloudAccountRole.Users); err != nil {
			logger.Error(err, "error unmarshalling users json", "usersJson", string(usersJson))
			return nil, status.Errorf(codes.Internal, "error unmarshalling users json")
		}
	}

	return &cloudAccountRole, nil
}

func (p CloudAccountRoleRepository) LookupPermission(ctx context.Context, cloudAccountId string, subject string, resourceType string, resourcesId []string, action string, returnType string) (map[string]string, error) {
	var rows *sql.Rows
	var query string
	if returnType == "resources" {
		query = "SELECT p.resource_id,effect"
	} else {
		query = "SELECT jsonb_array_elements_text(p.actions),effect"
	}
	query = query + `
	FROM cloud_account_roles c
	LEFT JOIN
    permissions p
	ON
    c.id = p.cloud_account_role_id
	WHERE c.cloud_account_id = $1
	AND p.resource_type = $2
	AND (
		p.resource_id = ANY($3)
		OR
		p.resource_id = '*'
	)
	AND (c.users @> $4::jsonb OR c.users @> '"*"'::jsonb)
	AND ( $5 = '""'  OR p.actions @> $5::jsonb )`

	rows, err := p.db.Query(query, cloudAccountId, resourceType, pq.Array(resourcesId), `"`+subject+`"`, `"`+action+`"`)
	if err != nil {
		logger.Error(err, "error executing db query")
		return nil, status.Error(codes.Internal, "failed to execute query")
	}
	defer rows.Close()
	allowedObjectsMap := make(map[string]string)
	for rows.Next() {
		var allowed string
		var effect string
		if err := rows.Scan(&allowed, &effect); err != nil {
			logger.Error(err, "error scanning rows")
			return nil, err
		}

		logger.V(9).Info("found allow permission", "allowed", allowed, effect)

		if _, ok := allowedObjectsMap[allowed]; !ok {
			allowedObjectsMap[allowed] = effect
		} else {
			if effect == pb.CloudAccountRole_deny.Enum().String() {
				allowedObjectsMap[allowed] = effect
			}
		}
	}
	logger.V(9).Info("allowed map", "allowedMap", allowedObjectsMap)
	return allowedObjectsMap, nil
}

func (p CloudAccountRoleRepository) UpdateCloudAccountRole(ctx context.Context, cloudAccountRoleId string, cloudAccountId string, alias string, effect string, users []string, permissions []*pb.CloudAccountRoleUpdate_Permission) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "Failed to begin transaction")
		return err
	}

	// Ensure the transaction is rolled back if any error occurs
	defer func() {
		if p := recover(); p != nil {
			rollbackError := tx.Rollback()
			if rollbackError != nil {
				logger.Error(rollbackError, "there was an error performing the rollback")
			}
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			rollbackError := tx.Rollback() // err is non-nil; don't change it
			if rollbackError != nil {
				logger.Error(rollbackError, "there was an error performing the rollback")
			}
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err with commit err
		}
	}()

	queryUpdate := `
		UPDATE cloud_account_roles 
		SET alias = $3, effect = $4, updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{cloudAccountRoleId, cloudAccountId, alias, effect}

	if users != nil {
		queryUpdate += `, users = $5`
		args = append(args, users)
	}

	queryUpdate += ` WHERE id = $1 and cloud_account_id = $2`

	result, err := tx.ExecContext(ctx, queryUpdate, args...)

	if err != nil {
		logger.Error(err, "error encountered while updating the cloudAccountRole record")
		return status.Error(codes.Internal, "failed to update cloudAccountRole")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		logger.Error(err, "error getting the number of rows affected")
		return status.Error(codes.Internal, "error getting the number of rows affected")

	}

	if rowsAffected < 1 {
		logger.V(9).Info("cloud account role record not found")
		err = status.Error(codes.NotFound, "cloud account role not found")
		return err
	}

	// if permissions were not informed, it should not delete any
	if permissions != nil {

		permissionIds := []string{}

		for _, permission := range permissions {
			if permission.Id != nil {
				permissionIds = append(permissionIds, *permission.Id)
			}
		}

		query := `SELECT id FROM permissions WHERE cloud_account_role_id = $1 AND cloud_account_id = $2 AND id != ALL($3)`

		rows, err := tx.QueryContext(ctx, query, cloudAccountRoleId, cloudAccountId, pq.Array(permissionIds))
		if err != nil {
			logger.Error(err, "error fetching permissions")
			return status.Error(codes.Internal, "error fetching permissions")
		}

		var permissionIdsToBeRemoved []string

		for rows.Next() {
			var permissionId string
			if err := rows.Scan(&permissionId); err != nil {
				logger.Error(err, "error scanning permission id")
				return status.Error(codes.Internal, "error scanning permission id")
			}
			permissionIdsToBeRemoved = append(permissionIdsToBeRemoved, permissionId)
		}

		for _, permissionId := range permissionIdsToBeRemoved {
			err := p.RemovePermissionFromCloudAccountRole(ctx, tx, cloudAccountId, cloudAccountRoleId, permissionId)
			if err != nil {
				logger.Error(err, "there was an error removing a permission from the cloud account role")
			}
		}

	}

	for _, permission := range permissions {

		if permission.Id != nil {
			convertedPermission := &pb.CloudAccountRole_Permission{
				Id:           *permission.Id,
				ResourceType: permission.ResourceType,
				ResourceId:   permission.ResourceId,
				Actions:      permission.Actions,
			}

			err = p.UpdatePermissionCloudAccountRole(ctx, tx, cloudAccountId, cloudAccountRoleId, convertedPermission)

			if err != nil {
				logger.Error(err, "error updating permission")
				return err
			}
		} else {
			convertedPermission := &pb.CloudAccountRole_Permission{
				ResourceType: permission.ResourceType,
				ResourceId:   permission.ResourceId,
				Actions:      permission.Actions,
			}
			err = p.AddPermissionToCloudAccountRole(ctx, tx, cloudAccountId, cloudAccountRoleId, convertedPermission)

			if err != nil {
				logger.Error(err, "error adding permission")
				return err
			}
		}

	}

	logger.V(9).Info("cloud account role record updated successfully")

	return nil
}

func (p CloudAccountRoleRepository) RemoveCloudAccountRole(ctx context.Context, cloudAccountRoleId string, cloudAccountId string) error {
	tx, err := db.Begin()
	if err != nil {
		logger.Error(err, "error encountered while creating tx db")
		return status.Error(codes.Internal, "failed to remove cloud account role")
	}

	queryDelete := `delete from permissions where cloud_account_role_id = $1 and cloud_account_id = $2`

	_, err = tx.ExecContext(ctx, queryDelete, cloudAccountRoleId, cloudAccountId)
	if err != nil {
		rollbackError := tx.Rollback()
		if rollbackError != nil {
			logger.Error(rollbackError, "there was an error performing the rollback")
		}
		logger.Error(err, "error encountered while deleting the permission record")
		return status.Error(codes.Internal, "failed to remove permission")
	}

	queryDelete = `
		delete from cloud_account_roles where id = $1 and cloud_account_id = $2`

	result, err := tx.ExecContext(ctx, queryDelete, cloudAccountRoleId, cloudAccountId)
	if err != nil {
		rollbackError := tx.Rollback()
		if rollbackError != nil {
			logger.Error(rollbackError, "there was an error performing the rollback")
		}
		logger.Error(err, "error encountered while deleting the cloud account role record")
		return status.Error(codes.Internal, "failed to remove cloud account role")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err, "error getting the number of rows affected")
		return status.Error(codes.Internal, "failed to remove cloud account role")
	}
	if rowsAffected < 1 {
		logger.V(9).Info("cloud account role record not found")
		return status.Error(codes.NotFound, "cloud account role not found")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction for deletetion")
		return status.Error(codes.Internal, "failed to remove cloud account role")
	}

	logger.V(9).Info("cloud account role record removed successfully", "rowsAffected", rowsAffected)
	return nil
}

func (p CloudAccountRoleRepository) GetPermissionsUsingWildcardCount(ctx context.Context, cloudAccountId string, id string, resourceType string) (count uint32, err error) {
	count = 0

	query := "SELECT count(id) FROM permissions WHERE cloud_account_id = $1 AND id = $2 AND ids ? '*' AND type = $3 "

	row := p.db.QueryRow(query, cloudAccountId, id, resourceType)
	err = row.Scan(&count)

	if err == sql.ErrNoRows {
		logger.V(9).Info("no existing wildcard resource found")
		return 0, nil
	} else if err != nil {
		logger.Error(err, "error getting current permission")
		return 0, status.Error(codes.Internal, "failed to retrieve current permission")
	}
	return count, nil
}

func (p CloudAccountRoleRepository) AddUserToCloudAccountRole(ctx context.Context, cloudAccountId string, id string, userId string) error {
	update := `
	UPDATE cloud_account_roles SET users = users || to_jsonb($1::text), 
	updated_at = CURRENT_TIMESTAMP 
	WHERE cloud_account_id = $2 AND id = $3 
	AND NOT(users ? $1 )
	`

	_, err := p.db.ExecContext(ctx, update, userId, cloudAccountId, id)
	if err != nil {
		logger.Error(err, "error adding the user to cloud account role")
		return status.Error(codes.Internal, "failed to add user to cloud account role")
	}

	return nil
}

func (p CloudAccountRoleRepository) AddPermissionToCloudAccountRole(ctx context.Context, tx *sql.Tx, cloudAccountId string, cloudAccountRoleid string, permission *pb.CloudAccountRole_Permission) error {
	query := "INSERT INTO permissions (cloud_account_id, cloud_account_role_id, resource_type, resource_id, actions) VALUES ($1, $2, $3, $4, $5)"

	if permission.Actions == nil {
		permission.Actions = []string{}
	}

	args := []interface{}{cloudAccountId, cloudAccountRoleid, permission.ResourceType, permission.ResourceId, permission.Actions}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = p.db.ExecContext(ctx, query, args...)
	}

	if err != nil {
		logger.Error(err, "error inserting permission", "cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleid, "permission", permission)
		if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" {
			logger.Error(err, "error duplicate permission type / cloudAccountRoleId", "cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleid, "permission", permission)
			return status.Error(codes.AlreadyExists, "error, adding permission to cloud account role (type/cloudAccountRoleId) already exist")
		}
		return status.Errorf(codes.Internal, "error, while adding permission to cloud account role")
	}

	return nil
}

func (p CloudAccountRoleRepository) RemovePermissionFromCloudAccountRole(ctx context.Context, tx *sql.Tx, cloudAccountId string, cloudAccountRoleid string, permissionId string) error {
	queryDelete := `DELETE FROM permissions WHERE id = $1 AND cloud_account_role_id = $2 AND cloud_account_id = $3`

	args := []interface{}{permissionId, cloudAccountRoleid, cloudAccountId}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, queryDelete, args...)
	} else {
		_, err = p.db.ExecContext(ctx, queryDelete, args...)
	}

	if err != nil {
		logger.Error(err, "error deleting permission", "cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleid, "permissionId", permissionId)
		return status.Errorf(codes.Internal, "error, while removing permission from cloud account role")
	}

	return nil
}

func (p CloudAccountRoleRepository) UpdatePermissionCloudAccountRole(ctx context.Context, tx *sql.Tx, cloudAccountId string, cloudAccountRoleId string, permission *pb.CloudAccountRole_Permission) error {
	updateQuery := `
	UPDATE permissions 
	SET resource_type = $4, resource_id = $5, actions = $6, updated_at = CURRENT_TIMESTAMP 
	WHERE cloud_account_id = $1 AND cloud_account_role_id = $2 AND id = $3`

	args := []interface{}{cloudAccountId, cloudAccountRoleId, permission.Id, permission.ResourceType, permission.ResourceId, permission.Actions}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, updateQuery, args...)
	} else {
		_, err = p.db.ExecContext(ctx, updateQuery, args...)
	}

	if err != nil {
		logger.Error(err, "error updating permission", "cloudAccountId", cloudAccountId, "cloudAccountRoleId", cloudAccountRoleId, "permissionId", permission.Id)
		return status.Errorf(codes.Internal, "error updating permission: %v", err)
	}

	return nil
}

func (p CloudAccountRoleRepository) RemoveResourceFromCloudAccountRole(ctx context.Context, cloudAccountId string, id string, resourceId string, resourceType string) ([]string, error) {
	var update string
	query := `
		delete from permissions 
		WHERE cloud_account_id = $2 AND ($3::text = '' OR cloud_account_role_id = $3::uuid) AND resource_type = $4 AND resource_id = $1
		RETURNING id
		`

	rows, err := p.db.QueryContext(ctx, query, resourceId, cloudAccountId, id, resourceType)
	if err != nil {
		logger.Error(err, "error removing the resource from the cloud account role", "update", update)
		return nil, status.Errorf(codes.Internal, "error removing the resource from the cloud account role")
	}

	var affectedIDs []string
	for rows.Next() {
		var affectedID string
		if err := rows.Scan(&affectedID); err != nil {
			logger.Error(err, "error scanning the affected ID")
			return nil, status.Errorf(codes.Internal, "error scanning the affected ID")
		}
		affectedIDs = append(affectedIDs, affectedID)
	}

	return affectedIDs, nil
}

func (p CloudAccountRoleRepository) RemoveUserFromCloudAccountRole(ctx context.Context, cloudAccountId string, id string, userId string) ([]string, error) {
	update := `
	UPDATE cloud_account_roles SET users = users - $1, 
	updated_at = CURRENT_TIMESTAMP 
	WHERE cloud_account_id = $2 AND ($3::text = '' OR id = $3::uuid) 
	AND (users ? $1 )
	RETURNING id
	`
	rows, err := p.db.QueryContext(ctx, update, userId, cloudAccountId, id)
	if err != nil {
		logger.Error(err, "error removing the user from the cloud account role", "update", update)
		return nil, status.Errorf(codes.Internal, "error removing the user from the cloud account role")
	}
	var affectedIDs []string
	for rows.Next() {
		var affectedID string
		if err := rows.Scan(&affectedID); err != nil {
			logger.Error(err, "error scanning the affected ID")
			return nil, status.Errorf(codes.Internal, "error scanning the affected ID")
		}
		affectedIDs = append(affectedIDs, affectedID)
	}
	return affectedIDs, nil
}

func (p CloudAccountRoleRepository) ListPermissions(ctx context.Context, cloudAccountRoleId string) (permissions []*pb.CloudAccountRole_Permission, err error) {
	query := `
	SELECT  id, resource_type, resource_id, actions
	FROM permissions WHERE cloud_account_role_id = $1
	`
	rows, err := p.db.Query(query, cloudAccountRoleId)
	if err != nil {
		logger.Error(err, "error executing db query", "query", query)
		return nil, status.Error(codes.Internal, "failed to fetch permissions")
	}
	defer rows.Close()

	for rows.Next() {
		permission := pb.CloudAccountRole_Permission{}
		var actionsJson []byte

		if err := rows.Scan(&permission.Id, &permission.ResourceType, &permission.ResourceId, &actionsJson); err != nil {
			logger.Error(err, "error scanning rows")
			return nil, status.Errorf(codes.Internal, "error scanning rows")
		}
		if actionsJson != nil {
			permission.Actions = []string{}
			if err := json.Unmarshal(actionsJson, &permission.Actions); err != nil {
				logger.Error(err, "error unmarshalling actions json")
				return nil, status.Error(codes.Internal, "failed to process user data")

			}
		}

		permissions = append(permissions, &permission)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err, "error iterating over rows")
		return nil, status.Error(codes.Internal, "failed to process permissions")
	}
	logger.V(9).Info("permissions fetch successfully", "count", len(permissions))
	return permissions, nil
}

func (p CloudAccountRoleRepository) CountCloudAccountRole(ctx context.Context, cloudAccountId string) (uint32, error) {
	var count uint32
	query := "SELECT count(id) FROM cloud_account_roles WHERE cloud_account_id = $1"
	row := p.db.QueryRowContext(ctx, query, cloudAccountId)
	err := row.Scan(&count)
	if err != nil {
		logger.Error(err, "error getting count of cloud account roles")
		return 0, status.Error(codes.Internal, "failed to get count of cloud account roles")
	}
	return count, nil
}

func (p CloudAccountRoleRepository) CountPermission(ctx context.Context, cloudAccountId string, cloudAccountRoleId string) (uint32, error) {
	var count uint32
	query := "SELECT count(id) FROM permissions WHERE cloud_account_id = $1 AND cloud_account_role_id = $2"
	row := p.db.QueryRowContext(ctx, query, cloudAccountId, cloudAccountRoleId)
	err := row.Scan(&count)
	if err != nil {
		logger.Error(err, "error getting count of permissions")
		return 0, status.Error(codes.Internal, "failed to get count of permissions")
	}
	return count, nil
}
