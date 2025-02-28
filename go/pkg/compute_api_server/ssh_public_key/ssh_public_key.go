// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ssh_public_key

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SshPublicKeyService struct {
	pb.UnimplementedSshPublicKeyServiceServer
	cloudAccountAppClientServiceClient pb.CloudAccountAppClientServiceClient
	db                                 *sql.DB
}

func NewSshPublicKeyService(cloudAccountAppClientServiceClient pb.CloudAccountAppClientServiceClient, db *sql.DB) (*SshPublicKeyService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &SshPublicKeyService{
		cloudAccountAppClientServiceClient: cloudAccountAppClientServiceClient,
		db:                                 db,
	}, nil
}

func validateSshPublicKey(sshPublicKey string) error {
	if sshPublicKey == "" {
		return fmt.Errorf("invalid SSH public key")
	}

	keyFields := strings.Fields(sshPublicKey)

	if len(keyFields) < 2 {
		return fmt.Errorf("SshPublicKey should have at least algorithm and publickey")
	}

	// Decode SSH key from base64 to bytes
	decodePublicKey, err := base64.StdEncoding.DecodeString(keyFields[1])
	if err != nil {
		return fmt.Errorf("could not decode sshpublickey: %w", err)
	}

	algoToMinBits := map[string]int{
		"ssh-rsa":                            3072,
		"ecdsa-sha2-nistp256":                256,
		"sk-ecdsa-sha2-nistp256@openssh.com": 256,
		"ssh-ed25519":                        256,
		"sk-ssh-ed25519@openssh.com":         256,
		"ecdsa-sha2-nistp384":                384,
		"ecdsa-sha2-nistp521":                384,
	}

	if minBits, ok := algoToMinBits[keyFields[0]]; ok {
		// Convert bytes into bits
		keyLengthBits := len(decodePublicKey) * 8
		if keyLengthBits < minBits {
			return fmt.Errorf("required minimum bits for ssh algorithm %s is %d", keyFields[0], minBits)
		}
	} else {
		return fmt.Errorf("provided unsupported algorithm")
	}

	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sshPublicKey))
	if err != nil {
		return fmt.Errorf("could not parse sshpublickey spec: %w", err)
	}
	// check if the supported ssh algorithm and hash key are mismatching
	if keyFields[0] != publicKey.Type() {
		return fmt.Errorf("ssh algorithm and key are mismatching")
	}

	return nil
}

// Validate SSHPublicKeyName.
// SshPublicKeyName is valid when name is lowercase alphanumeric, contains '.', '-', '@'
func validateSshPublicKeyName(name string) error {
	re := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9\-\.@]{0,61}[a-z0-9])?$`)
	matches := re.FindAllString(name, -1)
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid sshpublickey name")
	}
	return nil
}

// Extract user email from token and/or user-credentials
func (s *SshPublicKeyService) extractEmailFromRequest(ctx context.Context) (string, error) {
	// mTLS calls will not have JWT in the context. So passing jwtRequired as false
	userEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		if err == grpcutil.ErrClaimNotFound {
			// request coming from app-client with client_id in the context
			clientId, err1 := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.ClientIdClaim)
			if err1 != nil {
				return "", status.Error(codes.InvalidArgument, err.Error())
			}
			// lookup email from cloudaccount user-credentials
			cloudAccount, err := s.cloudAccountAppClientServiceClient.GetAppClientCloudAccount(ctx, &pb.AccountClient{ClientId: clientId})
			if err != nil {
				return "", status.Error(codes.InvalidArgument, err.Error())
			}
			if cloudAccount == nil {
				return "", status.Error(codes.NotFound, errors.New("AppClient associated cloudAccount not found").Error())
			}
			userEmail = cloudAccount.Name
		} else {
			return "", status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return userEmail, nil
}

func (s *SshPublicKeyService) Create(ctx context.Context, req *pb.SshPublicKeyCreateRequest) (*pb.SshPublicKey, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SshPublicKeyService.Create").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN", logkeys.Request, req)
	defer log.Info("END")
	resp, err := func() (*pb.SshPublicKey, error) {
		// Extract user email.
		userEmail, err := s.extractEmailFromRequest(ctx)
		if err != nil {
			return nil, err
		}

		// Validate input.
		if err := validateSshPublicKey(req.Spec.SshPublicKey); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		// Calculate resourceId.
		resourceId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		resourceIdStr := resourceId.String()

		// Calculate name.
		var name string
		if req.Metadata.Name == "" {
			name = resourceIdStr
		} else {
			name = req.Metadata.Name
		}

		if err := validateSshPublicKeyName(name); err != nil {
			return nil, err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		creationTimestamp := timestamppb.Now()

		// Convert the timestamppb.Timestamp to time.Time
		timestamp := creationTimestamp.AsTime()

		// Insert into database.
		query := `insert into ssh_public_key (resource_id, cloud_account_id, name, ssh_public_key, creation_timestamp, owner_email) values ($1, $2, $3, $4, $5, $6)`
		if _, err := s.db.ExecContext(ctx, query, resourceIdStr, req.Metadata.CloudAccountId, name, req.Spec.SshPublicKey, pq.FormatTimestamp(timestamp), userEmail); err != nil {
			// if error code matches unique_violation then return appropriate code.
			// ref: https://github.com/lib/pq/blob/3d613208bca2e74f2a20e04126ed30bcb5c4cc27/error.go#L178
			pgErr := &pgconn.PgError{}
			if errors.As(err, &pgErr) && pgErr.Code == manageddb.KErrUniqueViolation {
				return nil, status.Error(codes.AlreadyExists, "insert: ssh_public_key "+name+" already exists")
			}
			return nil, err
		}

		// Build response.
		resp := pb.SshPublicKey{
			Metadata: &pb.ResourceMetadata{
				CloudAccountId:    cloudAccountId,
				Name:              name,
				ResourceId:        resourceIdStr,
				CreationTimestamp: creationTimestamp,
				AllowDelete:       true,
			},
			Spec: req.Spec,
		}
		resp.Spec.OwnerEmail = userEmail
		return &resp, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return resp, utils.SanitizeError(err)
}

func resourceNameOrId(ref *pb.ResourceMetadataReference) (string, string, error) {
	if x, ok := ref.GetNameOrId().(*pb.ResourceMetadataReference_ResourceId); ok {
		return "resource_id", x.ResourceId, nil
	}
	if x, ok := ref.GetNameOrId().(*pb.ResourceMetadataReference_Name); ok {
		return "name", x.Name, nil
	}
	return "", "", status.Error(codes.InvalidArgument, "either ResourceId or Name must be provided")
}

func (s *SshPublicKeyService) Get(ctx context.Context, req *pb.SshPublicKeyGetRequest) (*pb.SshPublicKey, error) {
	log := log.FromContext(ctx).WithName("SshPublicKeyService.Get")
	log.Info("BEGIN", logkeys.Request, req)
	defer log.Info("END")
	resp, err := func() (*pb.SshPublicKey, error) {

		// Extract user email.
		userEmail, err := s.extractEmailFromRequest(ctx)
		if err != nil {
			return nil, err
		}

		argName, arg, err := resourceNameOrId(req.Metadata)
		if err != nil {
			return nil, err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		query := fmt.Sprintf(`
			select cloud_account_id, resource_id, name, ssh_public_key, creation_timestamp, owner_email
			from   ssh_public_key
			where  cloud_account_id = $1
			  and  %s = $2
		`, argName)
		rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if !rows.Next() {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		resp := pb.SshPublicKey{
			Metadata: &pb.ResourceMetadata{},
			Spec:     &pb.SshPublicKeySpec{},
		}

		var creationTimestamp time.Time
		if err := rows.Scan(&resp.Metadata.CloudAccountId, &resp.Metadata.ResourceId, &resp.Metadata.Name, &resp.Spec.SshPublicKey, &creationTimestamp, &resp.Spec.OwnerEmail); err != nil {
			return nil, err
		}

		// Convert the time.Time value to *timestamppb.Timestamp
		pbTimestamp := timestamppb.New(creationTimestamp)
		resp.Metadata.CreationTimestamp = pbTimestamp
		resp.Metadata.AllowDelete = allowDelete(userEmail, resp.Spec.OwnerEmail)

		return &resp, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return resp, utils.SanitizeError(err)
}

func (s *SshPublicKeyService) Search(ctx context.Context, req *pb.SshPublicKeySearchRequest) (*pb.SshPublicKeySearchResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SshPublicKeyService.Search").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN", logkeys.Request, req)
	defer log.Info("END")

	resp, err := func() (*pb.SshPublicKeySearchResponse, error) {

		// Extract user email.
		userEmail, err := s.extractEmailFromRequest(ctx)
		if err != nil {
			return nil, err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		query := `
			select cloud_account_id, resource_id, name, ssh_public_key, creation_timestamp, owner_email
			from   ssh_public_key
			where  cloud_account_id = $1
			order by name
		`
		rows, err := s.db.QueryContext(ctx, query, cloudAccountId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []*pb.SshPublicKey
		var creationTimestamp time.Time
		for rows.Next() {
			item := &pb.SshPublicKey{
				Metadata: &pb.ResourceMetadata{},
				Spec:     &pb.SshPublicKeySpec{},
			}
			if err := rows.Scan(&item.Metadata.CloudAccountId, &item.Metadata.ResourceId, &item.Metadata.Name, &item.Spec.SshPublicKey, &creationTimestamp, &item.Spec.OwnerEmail); err != nil {
				return nil, err
			}

			// Convert the time.Time value to *timestamppb.Timestamp
			pbTimestamp := timestamppb.New(creationTimestamp)
			item.Metadata.CreationTimestamp = pbTimestamp
			item.Metadata.AllowDelete = allowDelete(userEmail, item.Spec.OwnerEmail)

			items = append(items, item)
		}
		resp := &pb.SshPublicKeySearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return resp, utils.SanitizeError(err)
}

func (s *SshPublicKeyService) SearchStream(req *pb.SshPublicKeySearchRequest, svc pb.SshPublicKeyService_SearchStreamServer) error {
	ctx := svc.Context()
	log := log.FromContext(ctx).WithName("SshPublicKeyService.SearchStream")
	log.Info("BEGIN", logkeys.Request, req)
	defer log.Info("END")

	err := func() error {

		// Extract user email.
		userEmail, err := s.extractEmailFromRequest(ctx)
		if err != nil {
			return err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return err
		}

		query := `
			select cloud_account_id, resource_id, name, ssh_public_key, owner_email
			from   ssh_public_key
			where  cloud_account_id = $1
			order by name
		`
		rows, err := s.db.QueryContext(ctx, query, cloudAccountId)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			resp := pb.SshPublicKey{
				Metadata: &pb.ResourceMetadata{},
				Spec:     &pb.SshPublicKeySpec{},
			}
			if err := rows.Scan(&resp.Metadata.CloudAccountId, &resp.Metadata.ResourceId, &resp.Metadata.Name, &resp.Spec.SshPublicKey, &resp.Spec.OwnerEmail); err != nil {
				return err
			}

			resp.Metadata.AllowDelete = allowDelete(userEmail, resp.Spec.OwnerEmail)

			if err := svc.Send(&resp); err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return utils.SanitizeError(err)
}

func (s *SshPublicKeyService) Delete(ctx context.Context, req *pb.SshPublicKeyDeleteRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SshPublicKeyService.Delete").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	log.Info("BEGIN", logkeys.Request, req)
	defer log.Info("END")
	resp, err := func() (*emptypb.Empty, error) {
		argName, arg, err := resourceNameOrId(req.Metadata)
		if err != nil {
			return nil, err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if err := s.allowDelete(ctx, cloudAccountId, argName, arg); err != nil {
			return nil, err
		}

		query := fmt.Sprintf(`
			delete
			from   ssh_public_key
			where  cloud_account_id = $1
			  and  %s = $2
		`, argName)
		result, err := s.db.ExecContext(ctx, query, cloudAccountId, arg)
		if err != nil {
			return nil, err
		}
		noOfRowsDeleted, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if noOfRowsDeleted == 0 {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return resp, utils.SanitizeError(err)
}

func (s *SshPublicKeyService) allowDelete(ctx context.Context, cloudAccountId string, argName string, arg string) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SshPublicKeyService.allowDelete").WithValues(logkeys.CloudAccountId, cloudAccountId).Start()
	defer span.End()

	log.Info("BEGIN", logkeys.CloudAccountId, cloudAccountId)
	defer log.Info("END")

	// Extract user email.
	userEmail, err := s.extractEmailFromRequest(ctx)
	if err != nil {
		return err
	}

	// Allow deletion if userEmail is empty
	if len(userEmail) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		select name, owner_email
		from   ssh_public_key
		where  cloud_account_id = $1
		  and  %s = $2
	`, argName)
	rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return status.Error(codes.NotFound, "resource not found")
	}
	resp := pb.SshPublicKey{
		Metadata: &pb.ResourceMetadata{},
		Spec:     &pb.SshPublicKeySpec{},
	}

	if err := rows.Scan(&resp.Metadata.Name, &resp.Spec.OwnerEmail); err != nil {
		return err
	}

	if resp.Spec.OwnerEmail != "" && resp.Spec.OwnerEmail != userEmail {
		return status.Error(codes.PermissionDenied, "only the owner is allowed to delete this SSH public key")
	}
	return nil
}

func (s *SshPublicKeyService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SshPublicKeyService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func allowDelete(userEmail string, ownerEmail string) bool {
	return userEmail == "" || ownerEmail == "" || ownerEmail == userEmail
}
