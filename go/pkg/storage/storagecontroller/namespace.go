// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	storageControllerApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
)

type NamespaceMetadata struct {
	ClusterId string
	Name      string
	User      string
	Password  string
	UUID      string
	Id        string
}

type NamespaceProperties struct {
	Quota     string
	IPFilters []IPFilter
}
type IPFilter struct {
	Start string
	End   string
}
type Namespace struct {
	Metadata   NamespaceMetadata
	Properties NamespaceProperties
}

// Checks if the namespace with given name exists. There is no `head` method, so we will try to get
// the namespace and check the failed message. Namespace lookup might failed because of:
// - grpc connection error
// - invalid credentials
// error is populated with right message in those cases
func (client *StorageControllerClient) IsNamespaceExists(ctx context.Context, queryParams NamespaceMetadata) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.IsNamespaceExists").Start()
	defer span.End()
	logger.Info("checking if namespace exists")
	clusterUuid := queryParams.UUID

	// START to be removed after ID is saved
	_, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.Name)
	// END to be removed after ID is saved
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return exists, err
	}
	return exists, nil
}

// Creates a namespace with given quota and basic auth for access control
func (client *StorageControllerClient) CreateNamespace(ctx context.Context, namespace Namespace) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.CreateNamespace").Start()
	defer span.End()
	logger.Info("starting namespace creation")
	clusterUuid := namespace.Metadata.UUID

	// Create a new NamespaceRequest object.
	size, err := strconv.ParseInt(namespace.Properties.Quota, 10, 64)
	if err != nil || size < 0 {
		logger.Error(err, "error in creating a namespace due to parsing")
		return err
	}
	logger.Info("making a call to storage client of weka")
	// Make a gRPC call to add a new namespace.
	logger.Info("creating a namespace", "payload: ", namespace)
	// Extract IP filters
	ipFilters := make([]IPFilter, 0)
	for _, ipFilter := range namespace.Properties.IPFilters {
		ipFilters = append(ipFilters, IPFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		})
	}

	// Convert IP filters to []*api.Namespace_IpFilter
	apiIpFilters := make([]*storageControllerApi.Namespace_IpFilter, len(ipFilters))
	for i, ipFilter := range ipFilters {
		apiIpFilters[i] = &storageControllerApi.Namespace_IpFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		}
	}
	_, err = client.NamespaceSvcClient.CreateNamespace(ctx,
		&storageControllerApi.CreateNamespaceRequest{
			ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUuid},
			Name:      namespace.Metadata.Name,
			AdminUser: &storageControllerApi.CreateNamespaceRequest_AdminUser{
				Name:     namespace.Metadata.User,
				Password: namespace.Metadata.Password,
			},
			Quota: &storageControllerApi.Namespace_Quota{
				TotalBytes: uint64(size),
			},
			IpFilters: apiIpFilters,
		})
	if err != nil {
		logger.Error(err, "error in creating a namespace ")
		return err
	}
	return nil
}

func intoNamespace(ctx context.Context, nsResponse *storageControllerApi.Namespace) Namespace {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.intoNamespace").Start()
	defer span.End()
	logger.Info("intoNamespace conversion function")
	logger.Info("nsResponse is", "payload: ", nsResponse)
	quota := strconv.FormatUint(nsResponse.GetQuota().GetTotalBytes(), 10)
	logger.Info("quota is", "payload: ", quota)

	namespace := Namespace{
		Metadata: NamespaceMetadata{
			Name: nsResponse.Name,
			Id:   nsResponse.Id.GetId(),
		},
		Properties: NamespaceProperties{
			Quota: quota,
		},
	}
	logger.Info("namespace returned is", "namespace: ", namespace)

	return namespace
}

func intoVastNamespace(ns *storageControllerApi.Namespace) Namespace {
	quota := strconv.FormatUint(ns.GetQuota().GetTotalBytes(), 10)
	ipFilters := make([]IPFilter, 0)
	for _, ipFilter := range ns.GetIpFilters() {
		ipFilters = append(ipFilters, IPFilter{
			Start: ipFilter.GetStart(),
			End:   ipFilter.GetEnd(),
		})
	}
	namespace := Namespace{
		Metadata: NamespaceMetadata{
			Name: ns.Name,
			Id:   ns.Id.GetId(),
		},
		Properties: NamespaceProperties{
			Quota:     quota,
			IPFilters: ipFilters,
		},
	}
	return namespace
}

func intoNamespaces(nsResponse []*storageControllerApi.Namespace) []Namespace {
	namespaceArray := make([]Namespace, 0)
	for _, ns := range nsResponse {
		quota := strconv.FormatUint(ns.GetQuota().GetTotalBytes(), 10)

		// Extract IP filters
		ipFilters := make([]IPFilter, 0)
		for _, ipFilter := range ns.GetIpFilters() {
			ipFilters = append(ipFilters, IPFilter{
				Start: ipFilter.GetStart(),
				End:   ipFilter.GetEnd(),
			})
		}

		namespace := Namespace{
			Metadata: NamespaceMetadata{
				Name: ns.Name,
				Id:   ns.Id.GetId(),
			},
			Properties: NamespaceProperties{
				Quota:     quota,
				IPFilters: ipFilters,
			},
		}
		namespaceArray = append(namespaceArray, namespace)
	}
	return namespaceArray
}

// Gets namespace object from the storage controller
func (client *StorageControllerClient) GetNamespace(ctx context.Context, queryParams NamespaceMetadata) (Namespace, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.GetNamespace").Start()
	defer span.End()

	namespaceObj := Namespace{}
	clusterUuid := queryParams.UUID

	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.Name)
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return namespaceObj, err
	}

	if !exists {
		return namespaceObj, nil
	}

	if ns != nil {
		logger.Info("namespace details:", logkeys.Namespace, ns)
		namespaceObj = intoNamespace(ctx, ns)
		return namespaceObj, nil
	} else {
		return Namespace{}, nil
	}
}

// Gets namespace object from the storage controller
func (client *StorageControllerClient) GetVastNamespace(ctx context.Context, queryParams NamespaceMetadata) (Namespace, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.GetNamespace").Start()
	defer span.End()

	namespaceObj := Namespace{}
	clusterUuid := queryParams.UUID

	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.Name)
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return namespaceObj, err
	}
	if !exists {
		return namespaceObj, nil
	}

	if ns != nil {
		logger.Info("namespace details:", logkeys.Namespace, ns)
		namespaceObj = intoVastNamespace(ns)
		return namespaceObj, nil
	} else {
		return Namespace{}, nil
	}
}

func (client *StorageControllerClient) GetAllFileSystemOrgs(ctx context.Context, clusterId string) ([]Namespace, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetAllFileSystemOrgs")

	ns, exists, err := client.getNamespaces(ctx, clusterId)
	if err != nil {
		logger.Error(err, "error in finding ns")
		return nil, exists, err
	}
	if !exists {
		logger.Info("namespace does not exists for this cluster id", logkeys.ClusterId, clusterId)
		return nil, exists, nil
	}
	nsList := intoNamespaces(ns)
	logger.Info("namespace array", logkeys.NamespaceList, nsList)
	return nsList, exists, nil
}

// Deletes namespace object from the storage controller
func (client *StorageControllerClient) DeleteNamespace(ctx context.Context, queryParams NamespaceMetadata) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.DeleteNamespace").Start()
	defer span.End()

	clusterUuid := queryParams.UUID

	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.Name)
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return err
	}
	if !exists {
		logger.Info("could not find namespace by name", logkeys.Namespace, queryParams.Name)
		return nil
	}
	logger.Info("making a call to storage delete client of weka")
	// Make a gRPC call to add a new namespace.
	logger.Info("DeleteNamespace sds input params", logkeys.Namespace, ns.Id)
	_, err = client.NamespaceSvcClient.DeleteNamespace(ctx, &storageControllerApi.DeleteNamespaceRequest{
		NamespaceId: ns.Id,
	})
	if err != nil {
		logger.Error(err, "error in deleting a namespace")
		return err
	}
	return nil
}

func (client *StorageControllerClient) ModifyNamespace(ctx context.Context, namespace Namespace, addFlag bool, extraSize string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.ModifyNamespace").Start()
	defer span.End()

	clusterUuid := namespace.Metadata.UUID
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, namespace.Metadata.Name)
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return err
	}
	if !exists {
		logger.Info("could not find namespace by name", logkeys.Namespace, namespace.Metadata.Name)
		return nil
	}
	// flag is used to see if we want to reduce namespace size or increase namespace size (depends on create/delete call)
	// Create a new NamespaceRequest object.
	existingSize, err := strconv.ParseInt(namespace.Properties.Quota, 10, 64)
	if err != nil {
		logger.Error(err, "error in converting existingSize to integer")
		return err
	}
	additionalSize, err := strconv.ParseInt(extraSize, 10, 64)
	if err != nil {
		logger.Error(err, "error in converting additionalSize to integer")
		return err
	}
	logger.Info("existing quota", logkeys.ExistingSize, existingSize)
	logger.Info("additional size quota", logkeys.AdditionalSize, additionalSize)

	finalQuota := existingSize
	if addFlag {
		finalQuota = finalQuota + additionalSize
	} else {
		finalQuota = finalQuota - additionalSize
	}

	logger.Info("final quota", logkeys.FinalQuota, finalQuota)
	logger.Info("making a call to storage client of weka")
	// Make a gRPC call to add a new namespace.
	logger.Info("UpdateNamespace sds input params", logkeys.Namespace, ns.Id, logkeys.Size, finalQuota)
	_, err = client.NamespaceSvcClient.UpdateNamespace(ctx, &storageControllerApi.UpdateNamespaceRequest{
		NamespaceId: ns.Id,
		Quota: &storageControllerApi.Namespace_Quota{
			TotalBytes: uint64(finalQuota),
		},
	})
	if err != nil {
		logger.Error(err, "error in modify a namespace")
		return err
	}
	return nil
}

func (client *StorageControllerClient) ModifyVastNamespace(ctx context.Context, namespace Namespace, addFlag bool, extraSize string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.ModifyNamespace").Start()
	defer span.End()

	clusterUuid := namespace.Metadata.UUID
	nsId := namespace.Metadata.Id
	// flag is used to see if we want to reduce namespace size or increase namespace size (depends on create/delete call)
	// Create a new NamespaceRequest object.
	existingSize, err := strconv.ParseInt(namespace.Properties.Quota, 10, 64)
	if err != nil {
		logger.Error(err, "error in converting existingSize to integer")
		return err
	}
	logger.Info("extra size is", logkeys.ExistingSize, extraSize)
	logger.Info("existing size is ", logkeys.ExistingSize, existingSize)
	additionalSize, err := strconv.ParseInt(extraSize, 10, 64)
	if err != nil {
		logger.Error(err, "error in converting additionalSize to integer")
		return err
	}

	logger.Info("additional size quota", logkeys.AdditionalSize, additionalSize)

	finalQuota := existingSize
	if addFlag {
		finalQuota = finalQuota + additionalSize
	} else {
		finalQuota = finalQuota - additionalSize
	}

	logger.Info("final quota", logkeys.FinalQuota, finalQuota)
	logger.Info("making a call to storage client of weka")
	// Make a gRPC call to add a new namespace.

	ipFilters := make([]IPFilter, 0)
	for _, ipFilter := range namespace.Properties.IPFilters {
		ipFilters = append(ipFilters, IPFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		})
	}

	// Convert IP filters to []*api.Namespace_IpFilter
	apiIpFilters := make([]*storageControllerApi.Namespace_IpFilter, len(ipFilters))
	for i, ipFilter := range ipFilters {
		apiIpFilters[i] = &storageControllerApi.Namespace_IpFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		}
	}
	_, err = client.NamespaceSvcClient.UpdateNamespace(ctx, &storageControllerApi.UpdateNamespaceRequest{
		NamespaceId: &storageControllerApi.NamespaceIdentifier{
			ClusterId: &storageControllerApi.ClusterIdentifier{
				Uuid: clusterUuid,
			},
			Id: nsId,
		},
		Quota: &storageControllerApi.Namespace_Quota{
			TotalBytes: uint64(finalQuota),
		},
		IpFilters: apiIpFilters,
	})
	if err != nil {
		logger.Error(err, "error in modify avast namespace")
		return err
	}
	return nil
}

func (client *StorageControllerClient) UpdateVastIPFilyers(ctx context.Context, namespace Namespace) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.UpdateVastIPFilyers").Start()
	defer span.End()

	clusterUuid := namespace.Metadata.UUID
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, namespace.Metadata.Name)
	if err != nil {
		logger.Error(err, "error in finding vast namespace by name")
		return err
	}
	if !exists {
		logger.Info("could not find vast namespace by name", logkeys.Namespace, namespace.Metadata.Name)
		return nil
	}
	ipFilters := make([]IPFilter, 0)
	for _, ipFilter := range namespace.Properties.IPFilters {
		ipFilters = append(ipFilters, IPFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		})
	}

	// Convert IP filters to []*api.Namespace_IpFilter
	apiIpFilters := make([]*storageControllerApi.Namespace_IpFilter, len(ipFilters))
	for i, ipFilter := range ipFilters {
		apiIpFilters[i] = &storageControllerApi.Namespace_IpFilter{
			Start: ipFilter.Start,
			End:   ipFilter.End,
		}
	}
	size, err := strconv.ParseInt(namespace.Properties.Quota, 10, 64)
	if err != nil || size < 0 {
		logger.Error(err, "error in creating a namespace due to parsing")
		return err
	}
	logger.Info("making a call to storage client of weka", "ipFilters is : ", apiIpFilters)
	_, err = client.NamespaceSvcClient.UpdateNamespace(ctx, &storageControllerApi.UpdateNamespaceRequest{
		NamespaceId: &storageControllerApi.NamespaceIdentifier{
			ClusterId: &storageControllerApi.ClusterIdentifier{
				Uuid: clusterUuid,
			},
			Id: ns.Id.Id,
		},
		Quota: &storageControllerApi.Namespace_Quota{
			TotalBytes: uint64(size),
		},
		IpFilters: apiIpFilters, // Directly assign apiIpFilters here
	})
	if err != nil {
		logger.Error(err, "error in modify avast namespace")
		return err
	}
	return nil
}
