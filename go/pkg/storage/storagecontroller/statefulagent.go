package storagecontroller

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
)

type StatefulClient struct {
	ClusterId        string
	ClientId         string
	Name             string
	CustomStatus     string
	PredefinedStatus string
}

type CreateClientRequest struct {
	ClusterId string
	Name      string
	IPAddr    string
}

type NetworkConfig struct {
	Name    string
	IPAddrs []string
	Gateway string
	Netmask int32
}

func (client *StorageControllerClient) RegisterClient(ctx context.Context, request *CreateClientRequest) (*StatefulClient, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.RegisterClient").Start()
	defer span.End()

	logger.Info("handling register stateful client request", logkeys.ClusterId, request.ClusterId, logkeys.Name, request.Name, logkeys.Address, request.IPAddr)

	params := stcnt_api.CreateStatefulClientRequest{
		ClusterId: &api.ClusterIdentifier{
			Uuid: request.ClusterId,
		},
		Name: request.Name,
		Ip:   request.IPAddr,
	}
	statefulClientResp, err := client.StatefulSvcClient.CreateStatefulClient(ctx, &params)
	if err != nil {
		return nil, err
	}
	logger.Info("statefulclient registration response", logkeys.Response, statefulClientResp)

	return &StatefulClient{
		ClusterId:        request.ClusterId,
		ClientId:         statefulClientResp.StatefulClient.Id.Id,
		Name:             statefulClientResp.StatefulClient.Name,
		CustomStatus:     statefulClientResp.StatefulClient.GetCustomStatus(),
		PredefinedStatus: mapStatusCodeToString(statefulClientResp.StatefulClient.GetPredefinedStatus()),
	}, nil
}

func (client *StorageControllerClient) DeRegisterClient(ctx context.Context, clusterId, clientId string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.DeRegisterClient").Start()
	defer span.End()

	logger.Info("handling de-register stateful client request", logkeys.ClusterId, clusterId, logkeys.ClientId, clientId)
	resp, err := client.StatefulSvcClient.DeleteStatefulClient(ctx, &stcnt_api.DeleteStatefulClientRequest{
		StatefulClientId: &stcnt_api.StatefulClientIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: clusterId,
			},
			Id: clientId,
		},
	})
	if err != nil {
		logger.Error(err, "error deleteing stateful client", logkeys.Error, resp)
		return err
	}

	return nil
}

func (client *StorageControllerClient) GetClient(ctx context.Context, clusterId, clientId string) (*StatefulClient, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.GetClient").Start()
	defer span.End()

	logger.Info("handling get stateful client request", logkeys.ClusterId, clusterId, logkeys.ClientId, clientId)
	statefulClientResp, err := client.StatefulSvcClient.GetStatefulClient(ctx, &stcnt_api.GetStatefulClientRequest{
		StatefulClientId: &stcnt_api.StatefulClientIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: clusterId,
			},
			Id: clientId,
		},
	})
	if err != nil {
		logger.Error(err, "error getting stateful client", logkeys.Error, statefulClientResp)
		return nil, err
	}

	return &StatefulClient{
		ClusterId:        clusterId,
		ClientId:         statefulClientResp.StatefulClient.Id.Id,
		Name:             statefulClientResp.StatefulClient.Name,
		CustomStatus:     statefulClientResp.StatefulClient.GetCustomStatus(),
		PredefinedStatus: mapStatusCodeToString(statefulClientResp.StatefulClient.GetPredefinedStatus()),
	}, nil
}

func (client *StorageControllerClient) ListClients(ctx context.Context, clusterId string, names []string) ([]*StatefulClient, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.RegisterClient").Start()
	defer span.End()

	logger.Info("handling list stateful client request", logkeys.ClusterId, clusterId, logkeys.Filters, names)
	statefulClientsResp, err := client.StatefulSvcClient.ListStatefulClients(ctx, &stcnt_api.ListStatefulClientsRequest{
		ClusterId: &api.ClusterIdentifier{
			Uuid: clusterId,
		},
		Filter: &stcnt_api.ListStatefulClientsRequest_Filter{
			Names: names,
		},
	})
	if err != nil {
		logger.Error(err, "error listing stateful client", logkeys.Error, statefulClientsResp)
		return nil, err
	}

	retVal := []*StatefulClient{}
	for _, cl := range statefulClientsResp.StatefulClients {
		retVal = append(retVal, &StatefulClient{
			ClusterId:        clusterId,
			ClientId:         cl.Id.Id,
			Name:             cl.Name,
			CustomStatus:     cl.GetCustomStatus(),
			PredefinedStatus: mapStatusCodeToString(cl.GetPredefinedStatus()),
		})
	}
	return retVal, nil
}

func mapStatusCodeToString(status stcnt_api.StatefulClient_Status) string {
	switch status {
	case stcnt_api.StatefulClient_STATUS_DEGRADED_UNSPECIFIED:
		return "STATUS_DEGRADED_UNSPECIFIED"
	case stcnt_api.StatefulClient_STATUS_DOWN:
		return "STATUS_DOWN"
	case stcnt_api.StatefulClient_STATUS_UP:
		return "STATUS_UP"
	default:
		return ""
	}
}
