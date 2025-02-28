// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/go-logr/logr"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
)

var (
	Logger      logr.Logger
	vpcMutexMap sync.Map
)

const (
	InfoLevel  = 0
	DebugLevel = 1
)

type UUID string

type OvnBasedHandler struct {
	OvnClient libovsdbclient.Client
	DbClient  *sql.DB
	v1.UnimplementedOvnnetServer
}

func GetVpcMutex(key string) *sync.Mutex {
	logger := Logger.WithName("GetVpcMutex")
	muInterface, loaded := vpcMutexMap.LoadOrStore(key, &sync.Mutex{})
	if !loaded {
		logger.V(DebugLevel).Info(fmt.Sprintf("Mutex created for VPC %s", key))
	} else {
		logger.V(DebugLevel).Info(fmt.Sprintf("Mutex called for VPC %s", key))
	}
	return muInterface.(*sync.Mutex)
}

func RemoveVpcMutex(key string) {
	logger := Logger.WithName("RemoveVpcMutex")
	vpcMutexMap.Delete(key)
	logger.V(DebugLevel).Info(fmt.Sprintf("Mutex removed for VPC %s", key))
}

func LockVpcMutex(mu *sync.Mutex, key string) {
	logger := Logger.WithName("LockVpcMutex")
	mu.Lock()
	logger.V(DebugLevel).Info(fmt.Sprintf("Mutex locked for VPC %s", key))
}

func UnlockVpcMutex(mu *sync.Mutex, key string) {
	logger := Logger.WithName("UnlockVpcMutex")
	logger.V(DebugLevel).Info(fmt.Sprintf("Mutex unlocked for VPC %s", key))
	mu.Unlock()
}

func (s *OvnBasedHandler) ListSubnets(ctx context.Context, r *v1.ListSubnetsRequest) (*v1.ListSubnetsResponse, error) {
	return ListSubnetsHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) CreateSubnet(ctx context.Context, r *v1.CreateSubnetRequest) (*v1.CreateSubnetResponse, error) {
	return CreateSubnetHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeleteSubnet(ctx context.Context, r *v1.DeleteSubnetRequest) (*v1.DeleteSubnetResponse, error) {
	return DeleteSubnetHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetSubnet(ctx context.Context, r *v1.GetSubnetRequest) (*v1.GetSubnetResponse, error) {
	return GetSubnetHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) CreatePort(ctx context.Context, r *v1.CreatePortRequest) (*v1.CreatePortResponse, error) {
	return CreatePortHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) ListPorts(ctx context.Context, r *v1.ListPortsRequest) (*v1.ListPortsResponse, error) {
	return ListPortsHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetPort(ctx context.Context, r *v1.GetPortRequest) (*v1.GetPortResponse, error) {
	return GetPortHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeletePort(ctx context.Context, r *v1.DeletePortRequest) (*v1.DeletePortResponse, error) {
	return DeletePortHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) UpdatePort(ctx context.Context, r *v1.UpdatePortRequest) (*v1.UpdatePortResponse, error) {
	return UpdatePortHandler(s.DbClient, s.OvnClient, r)
}
func (s *OvnBasedHandler) ListRouters(ctx context.Context, r *v1.ListRoutersRequest) (*v1.ListRoutersResponse, error) {
	return ListRoutersHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) CreateRouter(ctx context.Context, r *v1.CreateRouterRequest) (*v1.CreateRouterResponse, error) {
	return CreateRouterHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeleteRouter(ctx context.Context, r *v1.DeleteRouterRequest) (*v1.DeleteRouterResponse, error) {
	return DeleteRouterHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetRouter(ctx context.Context, r *v1.GetRouterRequest) (*v1.GetRouterResponse, error) {
	return GetRouterHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) ListRouterInterfaces(ctx context.Context, r *v1.ListRouterInterfacesRequest) (*v1.ListRouterInterfacesResponse, error) {
	return ListRouterInterfacesHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetRouterInterface(ctx context.Context, r *v1.GetRouterInterfaceRequest) (*v1.GetRouterInterfaceResponse, error) {
	return GetRouterInterfaceHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) CreateRouterInterface(ctx context.Context, r *v1.CreateRouterInterfaceRequest) (*v1.CreateRouterInterfaceResponse, error) {
	return CreateRouterInterfaceHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeleteRouterInterface(ctx context.Context, r *v1.DeleteRouterInterfaceRequest) (*v1.DeleteRouterInterfaceResponse, error) {
	return DeleteRouterInterfaceHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) ListStaticRoutes(ctx context.Context, r *v1.ListStaticRoutesRequest) (*v1.ListStaticRoutesResponse, error) {
	return ListStaticRoutesHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetStaticRoute(ctx context.Context, r *v1.GetStaticRouteRequest) (*v1.GetStaticRouteResponse, error) {
	return GetStaticRouteHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) CreateStaticRoute(ctx context.Context, r *v1.CreateStaticRouteRequest) (*v1.CreateStaticRouteResponse, error) {
	return CreateStaticRouteHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeleteStaticRoute(ctx context.Context, r *v1.DeleteStaticRouteRequest) (*v1.DeleteStaticRouteResponse, error) {
	return DeleteStaticRouteHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) ListVPCs(ctx context.Context, r *v1.ListVPCsRequest) (*v1.ListVPCsResponse, error) {
	return ListVPCsHandler(s.DbClient, r)
}

func (s *OvnBasedHandler) GetVPC(ctx context.Context, r *v1.GetVPCRequest) (*v1.GetVPCResponse, error) {
	return GetVPCHandler(s.DbClient, r)
}

func (s *OvnBasedHandler) CreateVPC(ctx context.Context, r *v1.CreateVPCRequest) (*v1.CreateVPCResponse, error) {
	return CreateVPCHandler(s.DbClient, r)
}

func (s *OvnBasedHandler) DeleteVPC(ctx context.Context, r *v1.DeleteVPCRequest) (*v1.DeleteVPCResponse, error) {
	return DeleteVPCHandler(s.DbClient, r)
}

func (s *OvnBasedHandler) CreateSecurityRule(ctx context.Context, r *v1.CreateSecurityRuleRequest) (*v1.CreateSecurityRuleResponse, error) {
	return CreateSecurityRuleHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeleteSecurityRule(ctx context.Context, r *v1.DeleteSecurityRuleRequest) (*v1.DeleteSecurityRuleResponse, error) {
	return DeleteSecurityRuleHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetSecurityRule(ctx context.Context, r *v1.GetSecurityRuleRequest) (*v1.GetSecurityRuleResponse, error) {
	return GetSecurityRuleHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) UpdateSecurityRule(ctx context.Context, r *v1.UpdateSecurityRuleRequest) (*v1.UpdateSecurityRuleResponse, error) {
	return UpdateSecurityRuleHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) ListSecurityRules(ctx context.Context, r *v1.ListSecurityRulesRequest) (*v1.ListSecurityRulesResponse, error) {
	return ListSecurityRulesHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) CreateSecurityGroup(ctx context.Context, r *v1.CreateSecurityGroupRequest) (*v1.CreateSecurityGroupResponse, error) {
	return CreateSecurityGroupHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) DeleteSecurityGroup(ctx context.Context, r *v1.DeleteSecurityGroupRequest) (*v1.DeleteSecurityGroupResponse, error) {
	return DeleteSecurityGroupHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) GetSecurityGroup(ctx context.Context, r *v1.GetSecurityGroupRequest) (*v1.GetSecurityGroupResponse, error) {
	return GetSecurityGroupHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) ListSecurityGroups(ctx context.Context, r *v1.ListSecurityGroupsRequest) (*v1.ListSecurityGroupsResponse, error) {
	return ListSecurityGroupsHandler(s.DbClient, s.OvnClient, r)
}

func (s *OvnBasedHandler) UpdateSecurityGroup(ctx context.Context, r *v1.UpdateSecurityGroupRequest) (*v1.UpdateSecurityGroupResponse, error) {
	return UpdateSecurityGroupHandler(s.DbClient, s.OvnClient, r)
}
