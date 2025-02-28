// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package protodb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoJsonTable provides CRUD operations for Protobuf messages stored in a Postgres table.
// Protobuf messages are serialized as JSON and stored in a jsonb column.
type ProtoJsonTable struct {
	Db *sql.DB
	// The name of the table.
	TableName string
	// The list of columns that make up a unique key, usually the primary key.
	KeyColumns []string
	// The list of columns that make up a secondary unique key.
	SecondaryKeyColumns []string
	// The column that contains the Protobuf message serialized as JSON.
	JsonDocumentColumn string
	// A column name that is a primary key or has a unique constraint or has a unique index. It cannot be other non-unique columns
	// https://www.postgresql.org/docs/current/sql-insert.html#SQL-ON-CONFLICT
	ConflictTarget string
	// An empty Protobuf message of the type that will be stored in the table.
	EmptyMessage proto.Message
	// A function that returns the key values in a Protobuf message, in the same order as KeyColumns.
	GetKeyValuesFunc func(m proto.Message) ([]any, error)
	// A function that returns the key values in a Protobuf message, in the same order as SecondaryKeyColumns.
	GetSecondaryKeyValuesFunc func(m proto.Message) ([]any, error)
	// A function that returns the column name/value pairs for a Protobuf message passed to the Search function.
	// If nil or the returned Flattened struct is empty, Search will be unfiltered.
	SearchFilterFunc func(m proto.Message) (Flattened, error)
}

// Put (insert or update) a Protobuf message into the table.
func (p *ProtoJsonTable) Put(ctx context.Context, req proto.Message) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.Put")
	jsonBytes, err := p.getMarshaler().Marshal(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Put: unable to serialize to json: %w", err)
	}
	keyValues, err := p.GetKeyValuesFunc(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Put: GetKeyValuesFunc: %w", err)
	}
	keys := &Flattened{
		Columns: append([]string{}, p.KeyColumns...),
		Values:  append([]any{}, keyValues...),
	}
	nonKeys := &Flattened{
		Columns: []string{p.JsonDocumentColumn},
		Values:  []any{string(jsonBytes)},
	}

	if len(p.SecondaryKeyColumns) != 0 {
		secondaryKeyValues, err := p.GetSecondaryKeyValuesFunc(req)
		if err != nil {
			return fmt.Errorf("ProtoJsonTable.Put: GetSecondaryKeyValuesFunc: %w", err)
		}
		// Remove common columns
		secondaryColumns := RemoveCommonColumns(p.KeyColumns, p.SecondaryKeyColumns)
		valuesMap := make(map[string]any)
		for index, keyCol := range p.KeyColumns {
			valuesMap[keyCol] = keyValues[index]
		}
		for index, secKeyCol := range p.SecondaryKeyColumns {
			valuesMap[secKeyCol] = secondaryKeyValues[index]
		}

		for _, col := range secondaryColumns {
			keys.Add(col, valuesMap[col])
		}
	}

	// Use Postgres upsert to insert or update.
	query := fmt.Sprintf(`
	insert into %s (%s, %s)
	values (%s, %s)
	on conflict (%s)
	do update set %s
	`,
		p.TableName, keys.GetColumnsString(), nonKeys.GetColumnsString(),
		keys.GetInsertValuesString(1), nonKeys.GetInsertValuesString(len(keys.Columns)+1),
		p.ConflictTarget,
		nonKeys.GetUpdateSetString(len(keys.Columns)+1),
	)
	args := append(append([]any{}, keys.Values...), nonKeys.Values...)
	log.Info("Executing query", "query", query, "args", args)
	if _, err := p.Db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("ProtoJsonTable.Put: insert: %w", err)
	}
	return nil
}

// Insert a Protobuf message into the table.
func (p *ProtoJsonTable) Create(ctx context.Context, req proto.Message) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.Create")
	jsonBytes, err := p.getMarshaler().Marshal(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Create: unable to serialize to json: %w", err)
	}
	keyValues, err := p.GetKeyValuesFunc(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Create: GetKeyValuesFunc: %w", err)
	}
	keys := &Flattened{
		Columns: append([]string{}, p.KeyColumns...),
		Values:  append([]any{}, keyValues...),
	}
	nonKeys := &Flattened{
		Columns: []string{p.JsonDocumentColumn},
		Values:  []any{string(jsonBytes)},
	}

	if len(p.SecondaryKeyColumns) != 0 {
		secondaryKeyValues, err := p.GetSecondaryKeyValuesFunc(req)
		if err != nil {
			return fmt.Errorf("ProtoJsonTable.Create: GetSecondaryKeyValuesFunc: %w", err)
		}
		// Remove common columns
		secondaryColumns := RemoveCommonColumns(p.KeyColumns, p.SecondaryKeyColumns)
		valuesMap := make(map[string]any)
		for index, keyCol := range p.KeyColumns {
			valuesMap[keyCol] = keyValues[index]
		}
		for index, secKeyCol := range p.SecondaryKeyColumns {
			valuesMap[secKeyCol] = secondaryKeyValues[index]
		}

		for _, col := range secondaryColumns {
			keys.Add(col, valuesMap[col])
		}
	}

	query := fmt.Sprintf(`
	insert into %s (%s, %s)
	values (%s, %s)
	`,
		p.TableName, keys.GetColumnsString(), nonKeys.GetColumnsString(),
		keys.GetInsertValuesString(1), nonKeys.GetInsertValuesString(len(keys.Columns)+1),
	)
	args := append(append([]any{}, keys.Values...), nonKeys.Values...)
	log.Info("Executing query", "query", query, "args", args)
	if _, err := p.Db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("ProtoJsonTable.Create: insert: %w", err)
	}
	return nil
}

// Update a Protobuf message in the table.
func (p *ProtoJsonTable) Update(ctx context.Context, req proto.Message) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.Update")
	jsonBytes, err := p.getMarshaler().Marshal(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Update: unable to serialize to json: %w", err)
	}
	keyValues, err := p.GetKeyValuesFunc(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Update: GetKeyValuesFunc: %w", err)
	}
	keys := &Flattened{
		Columns: append([]string{}, p.KeyColumns...),
		Values:  append([]any{}, keyValues...),
	}
	nonKeys := &Flattened{
		Columns: []string{p.JsonDocumentColumn},
		Values:  []any{string(jsonBytes)},
	}

	if len(p.SecondaryKeyColumns) != 0 {
		secondaryKeyValues, err := p.GetSecondaryKeyValuesFunc(req)
		if err != nil {
			return fmt.Errorf("ProtoJsonTable.Update: GetSecondaryKeyValuesFunc: %w", err)
		}
		// Remove common columns
		secondaryColumns := RemoveCommonColumns(p.KeyColumns, p.SecondaryKeyColumns)
		valuesMap := make(map[string]any)
		for index, keyCol := range p.KeyColumns {
			valuesMap[keyCol] = keyValues[index]
		}
		for index, secKeyCol := range p.SecondaryKeyColumns {
			valuesMap[secKeyCol] = secondaryKeyValues[index]
		}

		for _, col := range secondaryColumns {
			nonKeys.Add(col, valuesMap[col])
		}
	}

	query := fmt.Sprintf(`
	update %s
	set %s
	where %s
	`,
		p.TableName,
		nonKeys.GetUpdateSetString(len(keys.Columns)+1),
		keys.GetWhereString(1),
	)
	args := append(append([]any{}, keys.Values...), nonKeys.Values...)
	log.Info("Executing query", "query", query, "args", args)
	if _, err := p.Db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("ProtoJsonTable.Update: update: %w", err)
	}
	return nil
}

// Delete a Protobuf message from the table.
func (p *ProtoJsonTable) Delete(ctx context.Context, req proto.Message) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.Delete")
	keyValues, err := p.GetKeyValuesFunc(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Delete: GetKeyValuesFunc: %w", err)
	}
	keys := &Flattened{
		Columns: append([]string{}, p.KeyColumns...),
		Values:  append([]any{}, keyValues...),
	}
	query := fmt.Sprintf(`delete from %s where %s`,
		p.TableName, keys.GetWhereString(1),
	)
	args := keys.Values
	log.Info("Executing query", "query", query, "args", args)
	if _, err := p.Db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("ProtoJsonTable.Delete: delete: %w", err)
	}
	return nil
}

// Delete a Protobuf message from the table by secondary key
func (p *ProtoJsonTable) DeleteBySecondaryKey(ctx context.Context, req proto.Message) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.DeleteBySecondaryKey")
	secondaryKeyValues, err := p.GetSecondaryKeyValuesFunc(req)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.DeleteBySecondaryKey: GetSecondaryKeyValuesFunc: %w", err)
	}
	secondaryKeys := &Flattened{
		Columns: append([]string{}, p.SecondaryKeyColumns...),
		Values:  append([]any{}, secondaryKeyValues...),
	}
	query := fmt.Sprintf(`delete from %s where %s`,
		p.TableName, secondaryKeys.GetWhereString(1),
	)
	args := secondaryKeys.Values
	log.Info("Executing query", "query", query, "args", args)
	if _, err := p.Db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("ProtoJsonTable.DeleteBySecondaryKey: delete: %w", err)
	}
	return nil
}

// Get a Protobuf message from the table.
func (p *ProtoJsonTable) Get(ctx context.Context, req proto.Message) (proto.Message, error) {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.Get")
	keyValues, err := p.GetKeyValuesFunc(req)
	if err != nil {
		return nil, fmt.Errorf("ProtoJsonTable.Get: GetKeyValuesFunc: %w", err)
	}
	keys := &Flattened{
		Columns: append([]string{}, p.KeyColumns...),
		Values:  append([]any{}, keyValues...),
	}
	query := fmt.Sprintf(`select %s from %s where %s`,
		p.JsonDocumentColumn, p.TableName, keys.GetWhereString(1))
	args := keys.Values
	log.Info("Executing query", "query", query, "args", args)
	rows, err := p.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ProtoJsonTable.Get: select: %w", err)
	}
	defer rows.Close()
	var resp proto.Message
	handlerFunc := func(m proto.Message) error {
		resp = m
		return nil
	}
	if err := p.scanRows(ctx, rows, handlerFunc); err != nil {
		return nil, fmt.Errorf("ProtoJsonTable.Get: %w", err)
	}
	if resp == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return resp, nil
}

func (p *ProtoJsonTable) GetBySecondaryKey(ctx context.Context, req proto.Message) (proto.Message, error) {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.GetBySecondaryKey")
	secondaryKeyValues, err := p.GetSecondaryKeyValuesFunc(req)
	if err != nil {
		return nil, fmt.Errorf("ProtoJsonTable.GetBySecondaryKey: GetSecondaryKeyValuesFunc: %w", err)
	}
	secondaryKeys := &Flattened{
		Columns: append([]string{}, p.SecondaryKeyColumns...),
		Values:  append([]any{}, secondaryKeyValues...),
	}
	query := fmt.Sprintf(`select %s from %s where %s`,
		p.JsonDocumentColumn, p.TableName, secondaryKeys.GetWhereString(1))
	args := secondaryKeys.Values
	log.Info("Executing query", "query", query, "args", args)
	rows, err := p.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ProtoJsonTable.GetBySecondaryKey: select: %w", err)
	}
	defer rows.Close()
	var resp proto.Message
	handlerFunc := func(m proto.Message) error {
		resp = m
		return nil
	}
	if err := p.scanRows(ctx, rows, handlerFunc); err != nil {
		return nil, fmt.Errorf("ProtoJsonTable.GetBySecondaryKey: %w", err)
	}
	if resp == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return resp, nil
}

// Search through all Protobuf messages in the table.
// For each message, the provided handlerFunc is called with the message.
// This can be easily adapted for streaming and non-streaming use cases.
func (p *ProtoJsonTable) Search(ctx context.Context, req proto.Message, handlerFunc func(proto.Message) error) error {
	where := Flattened{}
	if p.SearchFilterFunc != nil {
		var err error
		where, err = p.SearchFilterFunc(req)
		if err != nil {
			return fmt.Errorf("ProtoJsonTable.Search: SearchFilterFunc: %w", err)
		}
	}
	query := fmt.Sprintf(`select %s from %s where %s`, p.JsonDocumentColumn, p.TableName, where.GetWhereString(1))
	args := where.Values
	return p.executeQuery(ctx, query, args, handlerFunc)
}

// Search through all Protobuf messages in the table.
// For each message, the provided handlerFunc is called with the message.
// This can be easily adapted for streaming and non-streaming use cases.
func (p *ProtoJsonTable) SearchContains(ctx context.Context, req proto.Message, handlerFunc func(proto.Message) error) error {
	where := Flattened{}
	if p.SearchFilterFunc != nil {
		var err error
		where, err = p.SearchFilterFunc(req)
		if err != nil {
			return fmt.Errorf("ProtoJsonTable.Search: SearchFilterFunc: %w", err)
		}
	}
	// Ensure determinstic order, newest first if the first column contains timestamp
	query := fmt.Sprintf(`select %s from %s where %s order by  %s desc`, p.JsonDocumentColumn,
		p.TableName, where.GetWhereContains(1), p.KeyColumns[0])
	args := where.Values
	return p.executeQuery(ctx, query, args, handlerFunc)
}

func (p *ProtoJsonTable) executeQuery(ctx context.Context, query string, args []any,
	handlerFunc func(protoreflect.ProtoMessage) error) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.executeQuery")
	log.Info("Executing query", "query", query, "args", args)
	rows, err := p.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ProtoJsonTable.Search: select: %w", err)
	}
	defer rows.Close()
	return p.scanRows(ctx, rows, handlerFunc)
}

func (p *ProtoJsonTable) scanRows(ctx context.Context, rows *sql.Rows, handlerFunc func(proto.Message) error) error {
	log := log.FromContext(ctx).WithName("ProtoJsonTable.scanRows")
	for rows.Next() {
		var jsonBytes []byte
		if err := rows.Scan(&jsonBytes); err != nil {
			return fmt.Errorf("ProtoJsonTable.scanRows: Scan: %w", err)
		}
		log.V(9).Info("scanned", "jsonBytes", string(jsonBytes))
		message := proto.Clone(p.EmptyMessage)
		err := p.getMarshaler().Unmarshal(jsonBytes, message)
		if err != nil {
			return fmt.Errorf("ProtoJsonTable.scanRows: unable to deserialize from json: %w", err)
		}
		if err := handlerFunc(message); err != nil {
			return fmt.Errorf("ProtoJsonTable.scanRows: handlerFunc: %w", err)
		}
	}
	return nil
}

func (p *ProtoJsonTable) getMarshaler() *runtime.JSONPb {
	return &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			// When writing JSON, emit fields that have default values, including for enums.
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			// When reading JSON, ignore fields with unknown names.
			DiscardUnknown: true,
		},
	}
}

func RemoveCommonColumns(keyColumns, secondaryKeyColumns []string) []string {
	keyColumnsMap := make(map[string]bool)
	updatedSecondaryKeyColumns := make([]string, 0)

	for _, col := range keyColumns {
		keyColumnsMap[col] = true
	}

	// Remove duplicate items from secondaryKeyColumns using the map
	for _, secondaryKeyCol := range secondaryKeyColumns {
		if _, ok := keyColumnsMap[secondaryKeyCol]; !ok {
			updatedSecondaryKeyColumns = append(updatedSecondaryKeyColumns, secondaryKeyCol)
		}
	}
	return updatedSecondaryKeyColumns
}
