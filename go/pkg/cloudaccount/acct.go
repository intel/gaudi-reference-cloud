// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"github.com/golang/protobuf/ptypes/empty"
	authz "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var fieldOpts []protodb.FieldOptions = []protodb.FieldOptions{
	{Name: "parentId", StoreEmptyStringAsNull: true},
	{Name: "personId", StoreEmptyStringAsNull: true},
	{Name: "countryCode", StoreEmptyStringAsNull: true},
	{Name: "adminName", StoreEmptyStringAsNull: true},
}

type CloudAccountService struct {
	cfg                                       *config.Config
	authzClient                               *authz.AuthzClient
	pb.UnimplementedCloudAccountServiceServer // Used for forward compatability
}

// non-const for testing purposes
var maxId int64 = 1_000_000_000_000

const (
	kErrUniqueViolation = "23505"
	kMaxIdRetries       = 10
)

// Public for testing purposes
func NewId() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(maxId))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%012d", intId), nil
}

func MustNewId() string {
	id, err := NewId()
	if err != nil {
		panic(err)
	}
	return id
}

func IsValidId(id string) bool {
	if len(id) != 12 {
		return false
	}

	for _, ch := range id {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

func CheckValidId(id string) error {
	if !IsValidId(id) {
		//CloudAccountId should be string with exactly 12 digits
		return status.Error(codes.InvalidArgument, "invalid CloudAccountId")
	}
	return nil
}

func (cs *CloudAccountService) tryCreate(ctx context.Context, obj *pb.
	CloudAccountCreate) (*pb.CloudAccountId, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.tryCreate").WithValues("oid", obj.Oid).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	id, err := NewId() // cloud account id.

	if err != nil {
		logger.Error(err, "error generating account id")
		return nil, fmt.Errorf("unable to create account id: %w", err)
	}

	span.SetAttributes(attribute.String("cloudAccountId", id))

	params := protodb.NewProtoToSql(obj, fieldOpts...)
	vals := params.GetValues()
	vals = append(vals, id)
	query := fmt.Sprintf("INSERT INTO cloud_accounts (%v, id) VALUES(%v, $%v)",
		params.GetNamesString(), params.GetParamsString(), len(vals))
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error starting db transaction")
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, query, vals...); err != nil {
		logger.Error(err, "error inserting account into db", "query", query)
		return nil, err
	}
	query = "INSERT INTO cloud_account_members (account_id, member) VALUES($1, $2)"
	if _, err := tx.ExecContext(ctx, query, id, obj.Name); err != nil {
		logger.Error(err, "error inserting members into db", "query", query)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing db transaction")
		return nil, err
	}

	// set authz system role for cloud_account_admin
	if cs.cfg.Authz.Enabled {
		if _, err := cs.authzClient.AssignSystemRole(ctx, &pb.RoleRequest{CloudAccountId: id, Subject: obj.Name, SystemRole: pb.SystemRole_cloud_account_admin.String()}); err != nil {
			logger.Error(err, "couldn't assign systemrole")
			return nil, status.Errorf(codes.Internal, "unable to create systemrole")
		}
	}

	return &pb.CloudAccountId{Id: id}, nil
}

func (cs *CloudAccountService) Create(ctx context.Context, obj *pb.
	CloudAccountCreate) (*pb.CloudAccountId, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.Create").WithValues("oid", obj.GetOid(), "name", obj.GetName()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.V(9).Info("Create CloudAccount invoked", "name", obj.Name, "type", obj.Type)

	// Input Validation
	if obj.Type == pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED {
		logger.V(9).Info("Cloud Account type is unspecified for ", "name", obj.Name)
		return nil, status.Error(codes.InvalidArgument, "Cloud Account Type is not specified")
	}

	var err error
	for ii := 0; ii < kMaxIdRetries; ii++ {
		var id *pb.CloudAccountId
		if id, err = cs.tryCreate(ctx, obj); err == nil {
			return id, nil
		}
		pgErr := &pgconn.PgError{}
		if !errors.As(err, &pgErr) ||
			pgErr.Code != kErrUniqueViolation {
			return nil, err
		}
		logger.Info("retrying Cloud account create after id collision", "cloudAccountId", id, "iter", ii)
	}

	logger.Error(err, "unable to find unique id", "maxIdRetries", kMaxIdRetries)
	return nil, status.Error(codes.Internal, fmt.Sprintf("error creating cloud account: %v", err))
}

func (cs *CloudAccountService) Ensure(ctx context.Context, obj *pb.
	CloudAccountCreate) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.Ensure").WithValues("oid", obj.GetOid()).Start()
	defer span.End()
	logger.Info("Ensure API invoked", "oid", obj.Oid, "type", obj.Type)
	acct, err := cs.GetByOid(ctx, &pb.CloudAccountOid{Oid: obj.Oid})
	if err == nil {
		logger.Info("Cloud account already exists", "oid", obj.Oid)
		updateReq := compare(ctx, acct, obj)
		if updateReq != nil {
			logger.Info("Update CloudAccount with the newer details", "oid", obj.Oid)
			_, err = cs.Update(ctx, updateReq)
			if err != nil {
				return nil, err
			}
		}
		return cs.GetById(ctx, &pb.CloudAccountId{Id: acct.Id})
	}
	if status.Code(err) != codes.NotFound {
		logger.Error(err, "Failed to get by oid", "oid", obj.Oid)
		return nil, err
	}
	// Cloud account does not exist, create one
	id, err := cs.Create(ctx, obj)
	if err != nil {
		return nil, err
	}
	return cs.GetById(ctx, id)
}

func valsEqual(left interface{}, right interface{}) bool {
	if ts, ok := left.(*timestamppb.Timestamp); ok {
		rightTs := right.(*timestamppb.Timestamp)
		return timestampsCloseEnough(rightTs, ts)
	}

	return left == right
}

func timestampsCloseEnough(left *timestamppb.Timestamp, right *timestamppb.Timestamp) bool {
	// postgres has less precision in nanoseconds than golang
	return right.AsTime().Sub(left.AsTime()).Milliseconds() == 0
}

func ifFromVal(fd protoreflect.FieldDescriptor, value protoreflect.Value) any {
	if fd.Kind() == protoreflect.MessageKind {
		return value.Message().Interface()
	}
	return value.Interface()
}

// helper function to compare the existing cloud account meta with the request.
// It return a CloudAccountUpdate incase there are updates.
// A nil is returned incase there are no updates.
// Note: OID is not allowed to be changed here.
func compare(ctx context.Context, current *pb.CloudAccount, new *pb.CloudAccountCreate) *pb.CloudAccountUpdate {
	logger := log.FromContext(ctx)
	isChanged := false
	update := pb.CloudAccountUpdate{Id: current.Id}

	newMsg := new.ProtoReflect()
	newMsgFields := newMsg.Descriptor().Fields()

	curMsg := current.ProtoReflect()
	curMsgFields := curMsg.Descriptor().Fields()

	updateMsg := update.ProtoReflect()
	updateMsgFields := updateMsg.Descriptor().Fields()

	for ii := 0; ii < newMsgFields.Len(); ii++ {
		newFd := newMsgFields.Get(ii)
		if !newMsg.Has(newFd) {
			continue
		}
		updateFd := updateMsgFields.ByName(newFd.Name())
		if updateFd == nil {
			continue
		}
		curFd := curMsgFields.ByName(newFd.Name())
		newVal := newMsg.Get(newFd)
		curVal := curMsg.Get(curFd)
		if !valsEqual(ifFromVal(newFd, newVal), ifFromVal(curFd, curVal)) {
			updateMsg.Set(updateFd, newVal)
			isChanged = true
		}
	}

	if isChanged {
		logger.V(9).Info("CloudAccountCreate fields has updates", "oid", current.GetOid(), "cloudAccountId", current.GetId())
		return &update
	} else {
		logger.V(9).Info("No fields need to be updated.", "oid", current.GetOid(), "cloudAccountId", current.GetId())
		return nil
	}
}

func (cs *CloudAccountService) get(ctx context.Context, argName string,
	arg interface{}) (*pb.CloudAccount, error) {

	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.get").
		WithValues("argName", argName, "argValue", arg).Start()
	defer span.End()

	log.V(9).Info("Get CloudAccount invoked for ", "argName", argName, "argValue", arg)
	obj := pb.CloudAccount{}
	readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	query := fmt.Sprintf("SELECT %v from cloud_accounts WHERE %v = $1",
		readParams.GetNamesString(), argName)

	rows, err := db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, status.Errorf(codes.NotFound, "cloud account %v not found", arg)
	}
	if err = readParams.Scan(rows); err != nil {
		log.Error(err, "Error observed while fetching Cloud Account", "argName", argName, "argValue", arg)
		return nil, err
	}
	return &obj, nil
}

func (cs *CloudAccountService) GetById(ctx context.Context,
	cloudAccountId *pb.CloudAccountId) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.GetById").WithValues("cloudAccountId", cloudAccountId.GetId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := CheckValidId(cloudAccountId.Id); err != nil {
		return nil, err
	}
	return cs.get(ctx, "id", cloudAccountId.GetId())
}

func (cs *CloudAccountService) GetByName(ctx context.Context,
	cloudAccountName *pb.CloudAccountName) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.GetByName").WithValues("cloudAccountName", cloudAccountName.GetName()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	return cs.get(ctx, "name", cloudAccountName.GetName())
}

func (cs *CloudAccountService) GetByOid(ctx context.Context,
	cloudAccountOid *pb.CloudAccountOid) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.GetByOid").WithValues("oid", cloudAccountOid.GetOid()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	return cs.get(ctx, "oid", cloudAccountOid.GetOid())
}

func (cs *CloudAccountService) GetByPersonId(ctx context.Context,
	cloudAccountPersonId *pb.CloudAccountPersonId) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.GetByOid").WithValues("personId", cloudAccountPersonId.GetPersonid()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	return cs.get(ctx, "person_id", cloudAccountPersonId.GetPersonid())
}

func (cs *CloudAccountService) Search(filter *pb.CloudAccountFilter,
	svc pb.CloudAccountService_SearchServer) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(svc.Context()).WithName("CloudAccount.Search").WithValues("cloudAccountId", filter.GetId(), "oid", filter.GetOid()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	filterParams := protodb.NewProtoToSql(filter, fieldOpts...)
	obj := pb.CloudAccount{}
	readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	query := fmt.Sprintf("SELECT %v FROM cloud_accounts %v", readParams.GetNamesString(), filterParams.GetFilter())

	rows, err := db.QueryContext(ctx, query, filterParams.GetValues()...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err = readParams.Scan(rows); err != nil {
			return err
		}
		if err := svc.Send(&obj); err != nil {
			return err
		}
	}
	return nil
}

func (cs *CloudAccountService) Update(ctx context.Context,
	obj *pb.CloudAccountUpdate) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.Update").
		WithValues("cloudAccountId", obj.GetId()).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	proto := protodb.NewProtoToSql(obj, fieldOpts...)
	params := proto.GetParams()
	vals := proto.GetValues()

	query := strings.Builder{}
	_, err := query.WriteString("UPDATE cloud_accounts SET ")
	if err != nil {
		log.Error(err, "failed to update cloud account")
		return nil, err
	}

	idParam := ""
	creationTime := timestamppb.Now()
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)

	first := true
	for ii, name := range proto.GetNames() {
		if name == "id" {
			idParam = params[ii]
			continue
		}

		if first {
			first = false
		} else {
			_, err := query.WriteString(",")
			if err != nil {
				log.Error(err, "failed to write query string")
				return nil, err
			}
		}
		_, err := query.WriteString(name)
		if err != nil {
			return nil, err
		}

		_, err = query.WriteString("=")
		if err != nil {
			return nil, err
		}
		_, err = query.WriteString(params[ii])
		if name == "restricted" {
			queryT := `
			UPDATE cloud_accounts SET access_limited_timestamp=$1 WHERE id = $2 
			`
			_, err = db.ExecContext(ctx, queryT, creation_timestamp, obj.Id)
			if err != nil {
				log.Error(err, "failed to execute query for restricting account")
				return nil, err
			}
		}
		if err != nil {
			log.Error(err, "failed to write query string params")
			return nil, err
		}
	}
	if idParam == "" {
		err = status.Errorf(codes.Internal, "id column missing")
		log.Error(err, "failed to write query string id param")
		return nil, err
	}
	_, err = query.WriteString(fmt.Sprintf(" WHERE id = %v", idParam))
	if err != nil {
		log.Error(err, "failed to write query string where clause")
		return nil, err
	}
	_, err = db.ExecContext(ctx, query.String(), vals...)
	if err != nil {
		log.Error(err, "failed to update cloud account")
	}
	return &emptypb.Empty{}, err
}

func (cs *CloudAccountService) CheckCloudAccountExists(ctx context.Context, cloudAccountName *pb.CloudAccountName) (*wrapperspb.BoolValue, error) {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.Exists").
		WithValues("name", cloudAccountName.Name).Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	// CHECK DB
	count := 0
	err := db.QueryRow("SELECT COUNT(*) FROM cloud_accounts WHERE name = $1", cloudAccountName.Name).Scan(&count)
	if count != 1 {
		return &wrapperspb.BoolValue{Value: false}, err
	}
	return &wrapperspb.BoolValue{Value: true}, err
}

func (cs *CloudAccountService) Delete(ctx context.Context,
	obj *pb.CloudAccountId) (*empty.Empty, error) {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccount.Delete").
		WithValues("cloudAccountId", obj.Id).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	if err := CheckValidId(obj.Id); err != nil {
		return nil, err
	}
	_, err := db.Exec("DELETE FROM cloud_accounts WHERE id = $1", obj.Id)
	if err != nil {
		log.Error(err, "Failed to delete Cloud Account", "cloudAccountId", obj.Id)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (cs *CloudAccountService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	_, log, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountService.Ping").Start()
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}
