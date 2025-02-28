package pools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
)

// LocalPoolMapping read the pool mapping from a local file (or configMap file injected into a Pod)
type LocalPoolMapping struct {
	groupToPoolMappingFilePath string
}

func NewConfigmapPoolMappingReader(conf *idcnetworkv1alpha1.SDNControllerConfig) (*LocalPoolMapping, error) {
	// get the pool configuration
	poolByteValue, err := os.ReadFile(conf.ControllerConfig.PoolsConfigFilePath)
	if err != nil {
		return nil, err
	}
	var poolList idcnetworkv1alpha1.PoolList
	err = json.Unmarshal(poolByteValue, &poolList)
	if err != nil {
		return nil, err
	}
	pools := make(map[string]*idcnetworkv1alpha1.Pool)
	for i := range poolList.Items {
		pools[poolList.Items[i].Name] = poolList.Items[i]
	}

	return &LocalPoolMapping{
		groupToPoolMappingFilePath: conf.ControllerConfig.NodeGroupToPoolMappingConfigFilePath,
	}, nil
}

// GetPoolByGroupName
func (l *LocalPoolMapping) GetPoolByGroupName(ctx context.Context, groupName string) (string, error) {
	if len(l.groupToPoolMappingFilePath) == 0 {
		return "", fmt.Errorf("pools mapping file path is not provided")
	}

	mappingByteValue, err := os.ReadFile(l.groupToPoolMappingFilePath)
	if err != nil {
		return "", err
	}

	var mapping *idcnetworkv1alpha1.NodeGroupToPoolMap
	err = json.Unmarshal(mappingByteValue, &mapping)
	if err != nil {
		return "", err
	}

	poolName, found := mapping.NodeGroupToPoolMap[groupName]
	if !found {
		// it's not an error if there is no mapping for a Group.
		return "", nil
	}

	return poolName, nil
}

// GetGroupToPoolMappings
func (l *LocalPoolMapping) GetGroupToPoolMappings(ctx context.Context) (map[string]string, error) {
	if len(l.groupToPoolMappingFilePath) == 0 {
		return nil, fmt.Errorf("pools mapping file path is not provided")
	}

	mappingByteValue, err := os.ReadFile(l.groupToPoolMappingFilePath)
	if err != nil {
		return nil, err
	}

	var mapping *idcnetworkv1alpha1.NodeGroupToPoolMap
	err = json.Unmarshal(mappingByteValue, &mapping)
	if err != nil {
		return nil, err
	}

	return mapping.NodeGroupToPoolMap, nil
}

func (l *LocalPoolMapping) WatchGroupToPoolMappings() (chan MappingEvent, error) {
	return nil, nil
}
