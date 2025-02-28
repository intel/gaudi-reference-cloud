// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ip_resource_manager

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"github.com/praserx/ipconv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

type IpResourceManagerService struct {
	pb.UnimplementedIpResourceManagerServiceServer
	db *sql.DB
}

const (
	serializeUpdateError = "40001"
)

var dbRetryBackoff = wait.Backoff{
	Steps:    10,
	Duration: 10 * time.Millisecond,
	Factor:   1.1,
	Jitter:   0.5,
}

func NewIpResourceManagerService(db *sql.DB) (*IpResourceManagerService, error) {
	if db == nil {
		panic("db is required")
	}
	return &IpResourceManagerService{
		db: db,
	}, nil
}

// Create or update subnet and address records.
func (s *IpResourceManagerService) PutSubnet(ctx context.Context, req *pb.CreateSubnetRequest) (*pb.CreateSubnetResponse, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.PutSubnet")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.CreateSubnetResponse, error) {
		if err := validateAndNormalizeCreateSubnetRequest(req); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		// Run in a transaction to ensure that the subnet is not visible until all addresses have been added.
		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			return nil, fmt.Errorf("BeginTx: %w", err)
		}
		defer tx.Rollback()

		// Search for existing subnet.
		var subnetId int
		query := `
			select subnet_id
			from   subnet
			where  region = $1
			  and  availability_zone = $2
			  and  address_space = $3
			  and  subnet = $4
			  and  vlan_domain = $5
		`
		args := []any{req.Region, req.AvailabilityZone, req.AddressSpace, req.Subnet, req.VlanDomain}
		logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)
		err = tx.QueryRowContext(ctx, query, args...).Scan(&subnetId)
		if err == sql.ErrNoRows {
			// Insert new subnet into database.
			if err := tx.QueryRowContext(ctx, `select nextval('subnet_id_seq')`).Scan(&subnetId); err != nil {
				return nil, fmt.Errorf("select nextval: %w", err)
			}
			query := `
				insert into subnet (subnet_id, region, availability_zone, address_space, subnet, prefix_length, gateway, vlan_id, vlan_domain)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`
			_, err = tx.ExecContext(ctx, query, subnetId, req.Region, req.AvailabilityZone, req.AddressSpace, req.Subnet, req.PrefixLength, req.Gateway, req.VlanId, req.VlanDomain)
			if err != nil {
				return nil, fmt.Errorf("insert into subnet: %w", err)
			}
			logger.Info("Created new subnet", logkeys.SubnetId, subnetId, logkeys.Subnet, req.Subnet)
		} else if err != nil {
			return nil, fmt.Errorf("select subnet_id: %w", err)
		} else {
			// Update existing subnet.
			logger.Info("Updating existing subnet", logkeys.SubnetId, subnetId)
			query := `
				update subnet
				set    prefix_length = $2,
				       gateway = $3,
					   vlan_id = $4,
					   vlan_domain = $5
				where  subnet_id = $1
			`
			_, err = tx.ExecContext(ctx, query, subnetId, req.PrefixLength, req.Gateway, req.VlanId, req.VlanDomain)
			if err != nil {
				return nil, fmt.Errorf("insert into subnet: %w", err)
			}
		}

		// Insert addresses into database.
		var insertPlaceholders []string
		var insertValues []any
		for index, address := range req.Address {
			insertPlaceholders = append(insertPlaceholders, fmt.Sprintf("($%d,$%d)", 1+index*2, 2+index*2))
			insertValues = append(insertValues, subnetId, address)
		}
		query = fmt.Sprintf(`
			insert into address (subnet_id, address) values %s on conflict do nothing
		`, strings.Join(insertPlaceholders, ","))
		sqlResult, err := tx.ExecContext(ctx, query, insertValues...)
		if err != nil {
			return nil, fmt.Errorf("insert into address: %w", err)
		}
		insertedAddresses, err := sqlResult.RowsAffected()
		if err != nil {
			return nil, err
		}

		// Delete addresses from database.
		query = `
			delete from address where subnet_id = $1 and not (address = any($2))
		`
		sqlResult, err = tx.ExecContext(ctx, query, subnetId, req.Address)
		if err != nil {
			return nil, fmt.Errorf("delete from address: %w", err)
		}
		deletedAddresses, err := sqlResult.RowsAffected()
		if err != nil {
			return nil, err
		}

		logger.Info("Updated addresses", logkeys.SubnetId, subnetId, logkeys.InsertedAddresses, insertedAddresses, logkeys.DeletedAddresses, deletedAddresses)

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		resp := &pb.CreateSubnetResponse{}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Delete subnet and address records.
func (s *IpResourceManagerService) DeleteSubnet(ctx context.Context, req *pb.DeleteSubnetRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.DeleteSubnet")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if err := validateAndNormalizeDeleteSubnetRequest(req); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			return nil, fmt.Errorf("BeginTx: %w", err)
		}
		defer tx.Rollback()

		// Search for existing subnet.
		var subnetId int
		query := `
			select subnet_id
			from   subnet
			where  region = $1
			  and  availability_zone = $2
			  and  address_space = $3
			  and  subnet = $4
		`
		args := []any{req.Region, req.AvailabilityZone, req.AddressSpace, req.Subnet}
		logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)
		err = tx.QueryRowContext(ctx, query, args...).Scan(&subnetId)
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "DeleteSubnet: subnet not found")
		} else if err != nil {
			return nil, fmt.Errorf("DeleteSubnet: select subnet_id: %w", err)
		}

		logger.Info("Deleting subnet", logkeys.SubnetId, subnetId)

		// Delete addresses from database.
		query = `delete from address where subnet_id = $1`
		_, err = tx.ExecContext(ctx, query, subnetId)
		if err != nil {
			return nil, fmt.Errorf("DeleteSubnet: delete from address: %w", err)
		}

		// Delete subnet from database.
		query = `delete from subnet where subnet_id = $1`
		_, err = tx.ExecContext(ctx, query, subnetId)
		if err != nil {
			return nil, fmt.Errorf("DeleteSubnet: delete from subnet: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		resp := &emptypb.Empty{}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *IpResourceManagerService) SearchSubnetStream(req *pb.SearchSubnetRequest, svc pb.IpResourceManagerService_SearchSubnetStreamServer) error {
	ctx := svc.Context()
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.SearchSubnetStream")
	logger.Info("BEGIN", logkeys.Request, req)
	defer logger.Info("END")
	err := func() error {
		query := `
			select region, availability_zone, address_space, subnet, prefix_length, gateway, vlan_id, subnet_consumer_id, vlan_domain
			from   subnet
		`
		rows, err := s.db.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			resp := pb.Subnet{}
			var subnetConsumerId *string
			if err := rows.Scan(&resp.Region, &resp.AvailabilityZone, &resp.AddressSpace, &resp.Subnet, &resp.PrefixLength, &resp.Gateway, &resp.VlanId, &subnetConsumerId, &resp.VlanDomain); err != nil {
				return err
			}
			if subnetConsumerId != nil {
				resp.SubnetConsumerId = *subnetConsumerId
			}
			if err := svc.Send(&resp); err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		logger.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return err
}

// Try to find subnet record with same SubnetConsumerId.
// If not found, find subnet record with empty subnetConsumerId but same parameters, then set SubnetConsumerId.
func (s *IpResourceManagerService) ReserveSubnet(ctx context.Context, req *pb.ReserveSubnetRequest) (*pb.Subnet, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.ReserveSubnet")
	logger.Info("Request", logkeys.Request, req)
	var responseSubnet *pb.Subnet

	resp, err := func() (*pb.Subnet, error) {
		err := retry.OnError(dbRetryBackoff, retriable, func() error {
			tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
			})
			if err != nil {
				return err
			}
			defer tx.Rollback()
			// Try to find an existing reserved subnet with the same SubnetConsumerId.
			_, subnet, err := s.findSubnetWithConsumerId(ctx, tx, req.SubnetReference.SubnetConsumerId)
			if err == nil {
				// Existing reserved subnet found.
				if err := tx.Commit(); err != nil {
					return err
				}

				responseSubnet = subnet
				return nil
			}
			if status.Code(err) != codes.NotFound {
				return err
			}
			// Reserve an unreserved subnet with matching parameters.
			query := `
			update subnet
			set    subnet_consumer_id = $1
			where subnet_id in (
				select subnet_id
				from subnet
				where  region = $2
				  and  availability_zone = $3
				  and  vlan_domain = $4
				  and  address_space = $5
				  and  prefix_length <= $6
				  and  subnet is not null
				  and  gateway is not null
				  and  vlan_id is not null
				  and  subnet_consumer_id is null
				order by prefix_length desc, random()
				limit 1
			)
		`
			args := []any{req.SubnetReference.SubnetConsumerId, req.Spec.Region, req.Spec.AvailabilityZone, req.Spec.VlanDomain, req.Spec.AddressSpace, req.Spec.PrefixLengthHint}
			logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)

			sqlResult, err := tx.ExecContext(ctx, query, args...)

			if err != nil {
				return err
			}

			rowsAffected, err := sqlResult.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected < 1 {
				return status.Error(codes.ResourceExhausted, "no subnets available")
			}
			if rowsAffected > 1 {
				// Should never happen.
				return fmt.Errorf("more than one row affected")
			}
			_, subnet, err = s.findSubnetWithConsumerId(ctx, tx, req.SubnetReference.SubnetConsumerId)
			if err != nil {
				return err
			}
			if err := tx.Commit(); err != nil {
				return err
			}
			responseSubnet = subnet

			return nil
		})

		if err != nil {
			return nil, err
		}

		return responseSubnet, nil
	}()

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *IpResourceManagerService) ReleaseSubnet(ctx context.Context, req *pb.ReleaseSubnetRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.ReleaseSubnet")
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*emptypb.Empty, error) {
		// Serialization isolation level may be overkill but we'll use it to be safe.
		err := retry.OnError(dbRetryBackoff, retriable, func() error {
			tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
			})
			if err != nil {
				return err
			}
			defer tx.Rollback()
			_, _, err = s.findSubnetWithConsumerId(ctx, tx, req.SubnetReference.SubnetConsumerId)
			if err != nil {
				// If subnet was already released, this will return a NotFound error.
				return err
			}
			// Unreserve subnet, but only if there are no consumed addresses.
			query := `
			update subnet
			set    subnet_consumer_id = null
			where  subnet_consumer_id = $1
			  and  not exists (select address
				               from   address
							   where  address.subnet_id = subnet.subnet_id
							     and  address_consumer_id is not null)
		`
			args := []any{req.SubnetReference.SubnetConsumerId}
			logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)

			sqlResult, err := tx.ExecContext(ctx, query, args...)
			if err != nil {
				return err
			}

			rowsAffected, err := sqlResult.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected < 1 {
				return status.Error(codes.FailedPrecondition, "subnet has consumed addresses")
			}
			if rowsAffected > 1 {
				// Should never happen.
				return fmt.Errorf("more than one row affected")
			}
			if err := tx.Commit(); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		return &emptypb.Empty{}, nil
	}()

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *IpResourceManagerService) ReserveAddress(ctx context.Context, req *pb.ReserveAddressRequest) (*pb.ReserveAddressResponse, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.ReserveAddress")
	logger.Info("Request", logkeys.Request, req)
	var address string

	resp, err := func() (*pb.ReserveAddressResponse, error) {
		if err := validateAndNormalizeReserveAddressRequest(req); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		err := retry.OnError(dbRetryBackoff, retriable, func() error {
			tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
			})
			if err != nil {
				return err
			}
			defer tx.Rollback()

			// Find subnet.
			subnetId, _, err := s.findSubnetWithConsumerId(ctx, tx, req.SubnetReference.SubnetConsumerId)
			if err != nil {
				return err
			}

			// Try to find an existing reserved address with the same AddressConsumerId.
			address, err = s.findAddressWithConsumerId(ctx, tx, subnetId, req.AddressReference.AddressConsumerId, req.AddressReference.Address)
			if err == nil {
				// Existing reserved address found.
				if err := tx.Commit(); err != nil {
					return err
				}
				return nil
			}
			if status.Code(err) != codes.NotFound {
				return err
			}
			// Reserve an unreserved address in the subnet.
			var (
				query string
				args  []any
			)
			if req.AddressReference.Address == "" {
				query = `
				update address
				set    address_consumer_id = $2
				where subnet_id = $1
				  and address in (
					select address
					from   address
					where  subnet_id = $1
					  and  address_consumer_id is null
					order by random()
					limit 1
				)
			`
				args = []any{subnetId, req.AddressReference.AddressConsumerId}
			} else {
				query = `
				update address
				set    address_consumer_id = $2
				where  subnet_id = $1
				  and  address_consumer_id is null
				  and  address = $3
			`
				args = []any{subnetId, req.AddressReference.AddressConsumerId, req.AddressReference.Address}

			}
			logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)

			sqlResult, err := tx.ExecContext(ctx, query, args...)
			if err != nil {
				return err
			}

			rowsAffected, err := sqlResult.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected < 1 {
				return status.Error(codes.ResourceExhausted, "no addresses available")
			}
			if rowsAffected > 1 {
				// Should never happen.
				return fmt.Errorf("more than one row affected")
			}
			address, err = s.findAddressWithConsumerId(ctx, tx, subnetId, req.AddressReference.AddressConsumerId, req.AddressReference.Address)
			if err != nil {
				return err
			}
			if err := tx.Commit(); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		return &pb.ReserveAddressResponse{
			Address: address,
		}, nil
	}()

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Returns NotFound if subnet with this SubnetConsumerId does not exist.
// Returns NotFound if address with this AddressConsumerId does not exist in the subnet.
func (s *IpResourceManagerService) ReleaseAddress(ctx context.Context, req *pb.ReleaseAddressRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.ReleaseAddress")
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*emptypb.Empty, error) {
		if req.AddressReference.Address != "" {
			return nil, status.Error(codes.InvalidArgument, "address must be empty")
		}
		err := retry.OnError(dbRetryBackoff, retriable, func() error {
			tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
			})
			if err != nil {
				return err
			}
			defer tx.Rollback()

			// Find subnet.
			subnetId, _, err := s.findSubnetWithConsumerId(ctx, tx, req.SubnetReference.SubnetConsumerId)
			if err != nil {
				// If subnet was already released, this will return a NotFound error.
				return err
			}

			// Unreserve address.
			query := `
			update address
			set    address_consumer_id = null
			where  subnet_id = $1
			  and  address_consumer_id = $2
		`
			args := []any{subnetId, req.AddressReference.AddressConsumerId}
			logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)

			sqlResult, err := tx.ExecContext(ctx, query, args...)
			if err != nil {
				return err
			}

			rowsAffected, err := sqlResult.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected < 1 {
				return status.Error(codes.NotFound, "address not found")
			}
			if err := tx.Commit(); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *IpResourceManagerService) GetSubnetStatistics(ctx context.Context, req *emptypb.Empty) (*pb.GetSubnetStatisticsResponse, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.GetSubnetStatistics")
	logger.Info("BEGIN")
	defer logger.Info("END")

	resp, err := func() (*pb.GetSubnetStatisticsResponse, error) {
		query := `
			select 
				region,
				availability_zone,
				coalesce(address_space, '') as address_space,
				coalesce(vlan_domain, '') as vlan_domain,
				prefix_length,
				count(subnet) as total_subnets,
				sum(case when subnet_consumer_id is not null then 1 else 0 end) as total_consumed_subnets
			from subnet
			group by region, availability_zone, address_space, vlan_domain, prefix_length
		`

		rows, err := s.db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var subnetStatistics []*pb.SubnetStatisticsRecord
		for rows.Next() {
			subnetStatistic := &pb.SubnetStatisticsRecord{}
			if err := rows.Scan(&subnetStatistic.Region, &subnetStatistic.AvailabilityZone, &subnetStatistic.AddressSpace, &subnetStatistic.VlanDomain, &subnetStatistic.PrefixLength, &subnetStatistic.TotalSubnets, &subnetStatistic.TotalConsumedSubnets); err != nil {
				return nil, err
			}
			subnetStatistics = append(subnetStatistics, subnetStatistic)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		resp := &pb.GetSubnetStatisticsResponse{
			SubnetStatistics: subnetStatistics,
		}
		logger.Info("GetSubnetStatisticsResponse", logkeys.Response, resp)
		return resp, nil
	}()
	log.LogResponseOrError(logger, nil, resp, err)
	return resp, err
}

func (s *IpResourceManagerService) findSubnetWithConsumerId(ctx context.Context, tx *sql.Tx, subnetConsumerId string) (int, *pb.Subnet, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.findSubnetWithConsumerId")
	query := `
		select subnet_id, region, availability_zone, host(subnet) as subnet, prefix_length, gateway, vlan_id, vlan_domain, subnet_consumer_id
		from   subnet
		where  subnet_consumer_id = $1
	`
	args := []any{subnetConsumerId}
	logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)
	subnetId := int(0)
	resp := &pb.Subnet{}
	err := tx.QueryRowContext(ctx, query, args...).
		Scan(&subnetId, &resp.Region, &resp.AvailabilityZone, &resp.Subnet, &resp.PrefixLength, &resp.Gateway, &resp.VlanId, &resp.VlanDomain, &resp.SubnetConsumerId)
	if err == sql.ErrNoRows {
		return 0, nil, status.Error(codes.NotFound, "findSubnetWithConsumerId: subnet not found")
	} else if err != nil {
		return 0, nil, err
	}
	return subnetId, resp, nil
}

func (s *IpResourceManagerService) findAddressWithConsumerId(ctx context.Context, tx *sql.Tx, subnetId int, addressConsumerId, address string) (string, error) {
	logger := log.FromContext(ctx).WithName("IpResourceManagerService.findAddressWithConsumerId")
	var (
		query string
		args  []any
	)
	if address == "" {
		query = `
			select address
			from   address
			where  subnet_id = $1
			  and  address_consumer_id = $2
		`
		args = []any{subnetId, addressConsumerId}
	} else {
		query = `
			select address
			from   address
			where  subnet_id = $1
			  and  address_consumer_id = $2
			  and  address = $3
		`
		args = []any{subnetId, addressConsumerId, address}
	}
	logger.Info("Executing query", logkeys.Query, query, logkeys.Args, args)
	var foundAddress string
	err := tx.QueryRowContext(ctx, query, args...).Scan(&foundAddress)
	if err == sql.ErrNoRows {
		return "", status.Error(codes.NotFound, "findAddressWithConsumerId: address not found")
	} else if err != nil {
		return "", err
	}
	return foundAddress, nil
}

// Normalize subnet and prefix.
func normalizeSubnetAndPrefixLength(subnet string, prefixLength int) (*net.IPNet, int, error) {
	subnetCidr := subnet
	// Try to parse subnet in CIDR format 1.2.3.4/24.
	if _, _, err := net.ParseCIDR(subnet); err != nil {
		// Subnet is not in CIDR format 1.2.3.4/24.
		// Build CIDR format by adding PrefixLength field.
		subnetCidr = fmt.Sprintf("%s/%d", subnet, prefixLength)
	}
	subnetIp, ipNet, err := net.ParseCIDR(subnetCidr)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot parse subnet %q: %s", subnetCidr, err)
	}
	if !subnetIp.Equal(ipNet.IP) {
		return nil, 0, fmt.Errorf("subnet %q has non-zero host bits", subnetCidr)
	}
	prefixLength, _ = ipNet.Mask.Size()
	return ipNet, prefixLength, nil
}

func validateAndNormalizeCreateSubnetRequest(req *pb.CreateSubnetRequest) error {
	// Validate and normalize subnet and prefix.
	ipNet, prefixLength, err := normalizeSubnetAndPrefixLength(req.Subnet, int(req.PrefixLength))
	if err != nil {
		return err
	}
	req.Subnet = ipNet.String()
	req.PrefixLength = int32(prefixLength)

	// Generate addresses and gateway if requested.
	if req.GenerateAddressesMethod == pb.GenerateAddressesMethod_Auto {
		if len(req.Address) == 0 {
			req.GenerateAddressesMethod = pb.GenerateAddressesMethod_GenerateStandardHostAddresses
		} else {
			req.GenerateAddressesMethod = pb.GenerateAddressesMethod_NoGeneration
		}
	}
	if req.GenerateAddressesMethod == pb.GenerateAddressesMethod_GenerateStandardHostAddresses {
		gateway, addresses, err := generateStandardAddresses(ipNet.String())
		if err != nil {
			return err
		}
		req.Address = addresses
		if req.Gateway == "" {
			req.Gateway = gateway
		}
	}

	// Validate and normalize gateway.
	gatewayIp := net.ParseIP(req.Gateway)
	if gatewayIp == nil {
		return fmt.Errorf("cannot parse gateway %q", req.Gateway)
	}
	if !ipNet.Contains(gatewayIp) {
		return fmt.Errorf("gateway %q is not contained in subnet %q", gatewayIp, ipNet)
	}
	req.Gateway = gatewayIp.String()

	// Validate and normalize addresses.
	var normalizedAddresses []string
	for _, address := range req.Address {
		ip := net.ParseIP(address)
		if ip == nil {
			return fmt.Errorf("cannot parse address %q", address)
		}
		if !ipNet.Contains(ip) {
			return fmt.Errorf("address %q is not contained in subnet %q", address, ipNet)
		}
		normalizedAddresses = append(normalizedAddresses, ip.String())
	}
	req.Address = normalizedAddresses
	return nil
}

func validateAndNormalizeDeleteSubnetRequest(req *pb.DeleteSubnetRequest) error {
	// Validate and normalize subnet and prefix.
	ipNet, prefixLength, err := normalizeSubnetAndPrefixLength(req.Subnet, int(req.PrefixLength))
	if err != nil {
		return err
	}
	req.Subnet = ipNet.String()
	req.PrefixLength = int32(prefixLength)
	return nil
}

func validateAndNormalizeReserveAddressRequest(req *pb.ReserveAddressRequest) error {
	// Validate and normalize address.
	if req.AddressReference.Address != "" {
		addressIp := net.ParseIP(req.AddressReference.Address)
		if addressIp == nil {
			return fmt.Errorf("cannot parse address %q", req.AddressReference.Address)
		}
		// Skip checking if address is within subnet: if it isn't then it won't be found later.
		req.AddressReference.Address = addressIp.String()
	}
	return nil
}

func generateStandardAddresses(subnetCidr string) (string, []string, error) {
	subnetIp, ipNet, err := net.ParseCIDR(subnetCidr)
	if err != nil {
		return "", nil, err
	}
	subnetIpInt, err := ipconv.IPv4ToInt(subnetIp)
	if err != nil {
		return "", nil, err
	}
	gateway := ipconv.IntToIPv4(subnetIpInt + 1).String()
	// Skip network and reserved host addresses
	firstOffset := 1 + utils.GetReservedHostCount()
	ones, bits := ipNet.Mask.Size()
	numAddresses := 1 << (bits - ones)
	// Skip broadcast (.255)
	lastOffset := numAddresses - 2
	var addresses []string
	for i := firstOffset; i <= lastOffset; i++ {
		addresses = append(addresses, ipconv.IntToIPv4(subnetIpInt+uint32(i)).String())
	}
	return gateway, addresses, nil
}

// retriable function for the retry.OnError()
func retriable(err error) bool {
	// Check for error code 40001 in the err
	if err != nil {
		if strings.Contains(err.Error(), serializeUpdateError) {
			return true
		}
	}

	return false
}
