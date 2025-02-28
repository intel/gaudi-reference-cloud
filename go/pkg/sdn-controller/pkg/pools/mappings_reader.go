package pools

import (
	"context"
	"fmt"
)

// PoolMappingReader knows which NodeGroup belongs to which Pool and the configuration of a Pool.
type PoolMappingReader interface {
	GetPoolByGroupName(ctx context.Context, groupName string) (string, error)
	GetGroupToPoolMappings(ctx context.Context) (map[string]string, error)
	WatchGroupToPoolMappings() (chan MappingEvent, error)
}

type MappingEvent struct {
	Type      MappingEventType
	NodeGroup string
	Pool      string
	ResCh     chan MappingHandleResult
}

type MappingEventType string

const (
	MAPPING_EVENT_UPDATE MappingEventType = "update"
	MAPPING_EVENT_DELETE MappingEventType = "delete"
)

type MappingHandleResult struct {
	ResultStatus MappingEventProcessStatus
	ErrorMessage string
}

type MappingEventProcessStatus string

const (
	MAPPING_EVENT_PROCESS_UNKNOWN     MappingEventProcessStatus = "unknown"
	MAPPING_EVENT_PROCESS_NOOP        MappingEventProcessStatus = "noop"
	MAPPING_EVENT_PROCESS_IN_PROGRESS MappingEventProcessStatus = "inProgress"
	MAPPING_EVENT_PROCESS_FAILED      MappingEventProcessStatus = "failed"
	MAPPING_EVENT_PROCESS_SUCCESS     MappingEventProcessStatus = "success"
)

const (
	// "local" will be used when we load the mapping from a file(when deployed as a Pod, the file can be injected into the Pod from a ConfigMap)
	GroupPoolMappingSourceLocal = "local"
	// "mappingcrd" should be used when SDN read the mappings from the NodeGroupToPoolMapping CRs
	GroupPoolMappingSourceCRD = "crd"
)

func GetGroupPoolMappingReader(pmConf *PoolResourceManagerConf) (PoolMappingReader, error) {
	// logger := log.FromContext(context.Background()).WithName("GetGroupPoolMappingReader")
	var err error
	var mappingsReader PoolMappingReader
	if pmConf.CtrlConf.ControllerConfig.NodeGroupToPoolMappingSource == GroupPoolMappingSourceLocal {
		mappingsReader, err = NewConfigmapPoolMappingReader(pmConf.CtrlConf)
		if err != nil {
			return nil, err
		}
	} else if pmConf.CtrlConf.ControllerConfig.NodeGroupToPoolMappingSource == GroupPoolMappingSourceCRD {
		mappingEventRecorder := pmConf.Mgr.GetEventRecorderFor("mapping-controller")

		mappingsController := &NodeGroupToPoolMappingReconciler{
			Client:           pmConf.Mgr.GetClient(),
			Scheme:           pmConf.Mgr.GetScheme(),
			EventRecorder:    mappingEventRecorder,
			MappingEventChan: make(chan MappingEvent),
		}
		err = mappingsController.SetupWithManager(pmConf.Mgr)
		if err != nil {
			return nil, fmt.Errorf("unable to create NodeGroupToPoolMapping controller")
		}
		return mappingsController, nil
	} else {
		return nil, fmt.Errorf("unable to find a mapping reader implementation for [%v]", pmConf.CtrlConf.ControllerConfig.NodeGroupToPoolMappingSource)
	}

	return mappingsReader, nil
}
