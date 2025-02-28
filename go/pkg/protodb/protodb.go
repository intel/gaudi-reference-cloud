// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Package protodb facilitates using protobuf messages
// with SQL code.
//
// protodb is an alternative to using gorm that works well with
// protobuf messages and gives the caller direct control over
// the SQL queries used.
//
// To work with this code, the names of the fields in a protobuf
// messsage must correlate to names of columns in an SQL table.
//
// protodb converts camelCase protobuf fields name into snake_case
// SQL column names. The reason to convert to snake_case is
// Postgres has case insensitive column names. Using camelCase
// in Postgres results in loss of separation between words in an
// column names.
package protodb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FieldOptions is for customizing protodb's handling of fields.
type FieldOptions struct {
	// Name of the field this option is for
	Name string

	// If StoreEmptyStringAsNull is true, protodb maps null in the database to
	// the empty string in the struct field and vice versa.
	StoreEmptyStringAsNull bool
}

type tNameParamValue struct {
	names  []string
	params []string
	values []any
	// nullStrings is derived from FieldOptions. nullStrings is a map of the
	// fields where the empty string is stored as null.
	nullStrings map[string]bool
}

func (nvp *tNameParamValue) setOptions(opts []FieldOptions) {
	if len(opts) == 0 {
		return
	}
	nvp.nullStrings = map[string]bool{}
	for _, opt := range opts {
		if opt.StoreEmptyStringAsNull {
			nvp.nullStrings[opt.Name] = true
		}
	}
}

func (nvp *tNameParamValue) addParam(name string, value any) {
	nvp.names = append(nvp.names, name)
	nvp.params = append(nvp.params, fmt.Sprintf("$%v", len(nvp.params)+1))
	nvp.values = append(nvp.values, value)
}

// ProtoToSql is for constructing SQL queries and parameters.
// ProtoToSql holds three slices:
//
// names: accessible by GetNames(), names holds the SQL
// column names corresponding to fields that were received over
// the wire into a protobuf message.
//
// params: accessible by GetParama(), params holds the SQL
// parameters that should go into a query. For example, GetParams()
// for a protobuf message with three fields containing data
// would be []string{"$1", "$2", "$3"}.
//
// protodb currently uses postgres syntax for parameters
//
// values: Accessible by GetValues(), values holds the values
// for an SQL query. These values should be passed to sql.Exec
// as arguments after the query.
//
// The names, params, and values slices all have the same number
// of elements and the elements correspond to one another.
//
// For example:
//
//	params := protodb.NewProtoToSql(obj)
//	query := fmt.Sprintf("INSERT INTO cloud_accounts (%v) VALUES(%v) RETURNING id",
//	params.GetNamesString(), params.GetParamsString())
//	if _, err := tx.Exec(query, params.GetValues()...); err != nil {
//		return nil, err
//	}
//
// In the example code above, the query string looks like:
// "INSERT INTO cloud_accounts (name,owner,type) VALUES($1,$2,$3) RETURNING id"
type ProtoToSql struct {
	tNameParamValue
}

func (npvs *tNameParamValue) GetNames() []string {
	return append([]string{}, npvs.names...)
}

func (params *ProtoToSql) GetFilter() string {
	names := params.GetNames()
	if len(names) == 0 {
		return ""
	}
	buf := strings.Builder{}
	buf.WriteString("WHERE")
	args := params.GetParams()
	for ii, vv := range names {
		if ii > 0 {
			buf.WriteString(" AND")
		}
		buf.WriteString(fmt.Sprintf(" (%v) IN (%v)", vv, args[ii]))
	}
	return buf.String()
}

func (params *ProtoToSql) GetParams() []string {
	return append([]string{}, params.params...)
}

func (params *ProtoToSql) GetValues() []any {
	return append([]any{}, params.values...)
}

func (nvps *tNameParamValue) GetNamesString() string {
	return strings.Join(nvps.names, ",")
}

func (params *ProtoToSql) GetParamsString() string {
	return strings.Join(params.params, ",")
}

// Create a ProtoToSql for storing a protobuf message into an
// SQL database.
func NewProtoToSql(obj protoreflect.ProtoMessage, opts ...FieldOptions) *ProtoToSql {
	params := ProtoToSql{}
	params.setOptions(opts)

	msg := obj.ProtoReflect()
	msg.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		name := toSnakeCase(string(fd.Name()))
		var value interface{}
		if fd.Kind() != protoreflect.MessageKind {
			value = val.Interface()
			if params.nullStrings[string(fd.Name())] && fd.Kind() == protoreflect.StringKind {
				str, ok := value.(string)
				if ok && str == "" {
					value = nil
				}
			}
		} else {
			innerMsg := val.Message().Interface()
			if ts, ok := innerMsg.(*timestamppb.Timestamp); ok {
				value = ts.AsTime()
			} else {
				panic("can't handle non-timestamp nested interface")
			}
		}
		params.addParam(name, value)
		return true
	})
	return &params
}

// SqlToProto is used for reading protobuf messages from a table in
// an SQL database. SqlToProto has the same names, params, and values
// as ProtoToSql.
//
// The Scan() function wraps sql.Rows.Scan, storing the values
// from a row in the database into a protobuf message.
//
// SqlToProto may be used by itself to read protobuf messages from
// a database:
//
//	obj := pb.CloudAccount{}
//	readParams := protodb.NewSqlToProto(&obj)
//
//	query := fmt.Sprintf("SELECT %v from cloud_accounts WHERE %v = $1",
//		readParams.GetNamesString(), argName)
//
//	rows, err := db.Query(query, arg)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	while rows.Next() {
//		return nil, status.Errorf(codes.NotFound, "cloud account %v not found", arg)
//		if err = readParams.Scan(rows); err != nil {
//			return nil, err
//		}
//		...
//	}
//
// SqlToProto may be used in conjunction with ProtoToSql to provide
// query parameters for filtering results and processing the resulting
// rows:
//
//	filterParams := protodb.NewProtoToSql(filter)
//	obj := pb.CloudAccount{}
//	readParams := protodb.NewSqlToProto(&obj)
//
//	query := fmt.Sprintf("SELECT %v FROM cloud_accounts", readParams.GetNamesString())
//	names := filterParams.GetNamesString()
//	if len(names) > 0 {
//		query += fmt.Sprintf(" WHERE (%v) IN (%v)", names,
//			filterParams.GetParamsString())
//	}
//	rows, err := db.Query(query, filterParams.GetValues()...)
//	if err != nil {
//		return err
//	}
//	defer rows.Close()
//	for rows.Next() {
//		if err := readParams.Scan(rows); err != nil {
//			return err
//		}
//		...
//	}
type SqlToProto struct {
	tNameParamValue
}

type tProtoScanner struct {
	params *tNameParamValue
	fd     protoreflect.FieldDescriptor
	msg    protoreflect.Message
}

func (ss *tProtoScanner) Scan(val any) error {
	switch ss.fd.Kind() {
	case protoreflect.MessageKind:
		ts := val.(time.Time)
		tsVal := timestamppb.New(ts).ProtoReflect()
		ss.msg.Set(ss.fd, protoreflect.ValueOf(tsVal))
	case protoreflect.EnumKind:
		ss.msg.Set(ss.fd, protoreflect.ValueOfEnum(protoreflect.EnumNumber(val.(int64))))
	case protoreflect.StringKind:
		if val == nil && ss.params.nullStrings[string(ss.fd.Name())] {
			val = ""
		}
		fallthrough
	default:
		ss.msg.Set(ss.fd, protoreflect.ValueOf(val))
	}
	return nil
}

// Create an SqlToProto for storing values from a database into
// fields of a protobuf message
func NewSqlToProto(obj protoreflect.ProtoMessage, opts ...FieldOptions) *SqlToProto {
	params := SqlToProto{}
	params.setOptions(opts)
	msg := obj.ProtoReflect()
	fds := msg.Descriptor().Fields()
	for ii := 0; ii < fds.Len(); ii++ {
		fd := fds.Get(ii)
		name := toSnakeCase(string(fd.Name()))
		value := &tProtoScanner{params: &params.tNameParamValue, fd: fd, msg: msg}
		params.addParam(name, value)
	}
	return &params
}

// Call rows.Scan, storing the data into the fields of the protobuf
// message that was passed to NewSqlToProto
func (params *SqlToProto) Scan(rows *sql.Rows) error {
	if err := rows.Scan(params.values...); err != nil {
		return err
	}
	return nil
}

// Convert to snake case for case-insensitive Postgres column names
func toSnakeCase(str string) string {
	from := []byte(str)
	firstCap := false
	numCaps := 0

	for ii, bb := range from {
		if 'A' <= bb && bb <= 'Z' {
			numCaps++
			if ii == 0 {
				firstCap = true
			}
		}
	}

	if numCaps == 0 {
		return str
	}

	length := len(str) + numCaps
	if firstCap {
		length--
	}
	to := make([]byte, len(str)+numCaps)
	jj := 0
	for ii, bb := range from {
		if 'A' <= bb && bb <= 'Z' {
			if ii > 0 {
				to[jj] = '_'
				jj++
			}
			bb = bb - 'A' + 'a'
		}
		to[jj] = bb
		jj++
	}
	return string(to)
}
