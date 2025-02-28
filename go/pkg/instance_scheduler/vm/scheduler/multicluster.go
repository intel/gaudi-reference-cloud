// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This supports gathering node allocatable resources and workloads (pods) from multiple Kubernetes (Harvester) clusters.
// This is done by attaching informers to each configured Kubernetes cluster.
// When an informer calls an event handler (add pod, add node, etc.), it modifies the node name by prepending the clusterId.
// kube-scheduler then simply performs scheduling as if all nodes were part of a single Kubernetes cluster.
// At the end of the scheduling process, the clusterId is extracted from the recommended node name and this is used when
// writing the Instance to the Compute Database.
package scheduler

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	metal3Informerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/informers/externalversions"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

type Cluster struct {
	ClusterId        string
	InformerFactory  interface{}
	informersStarted chan struct{}
}

// Wait until scheduler cache has been updated with all nodes and pods from all clusters.
// TODO: Allow a subset of clusters to be down.
func (sched *Scheduler) WaitForCacheSync(ctx context.Context) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.WaitForCacheSync").Start()
	defer span.End()

	log.Info("BEGIN")
	defer log.Info("END")
	if len(sched.clusters) == 0 {
		return fmt.Errorf("Scheduler.WaitForCacheSync: no clusters")
	}
	for _, cluster := range sched.clusters {
		log.Info("Waiting for informer to start", logkeys.ClusterId, cluster.ClusterId)
		<-cluster.informersStarted
		log.Info("Waiting for informer cache to sync", logkeys.ClusterId, cluster.ClusterId)
		var syncResult map[reflect.Type]bool
		if cluster.ClusterId == BmaasLocalCluster {
			syncResult = cluster.InformerFactory.(metal3Informerfactory.SharedInformerFactory).WaitForCacheSync(ctx.Done())
		} else {
			syncResult = cluster.InformerFactory.(informers.SharedInformerFactory).WaitForCacheSync(ctx.Done())
		}

		log.Info("syncResult", logkeys.Result, fmt.Sprintf("%#v", syncResult), logkeys.ClusterId, cluster.ClusterId)
		for k, v := range syncResult {
			if !v {
				return fmt.Errorf("unable to synchronize %v from cluster %v", k, cluster.ClusterId)
			}
		}
	}
	return nil
}

const nodeIdSeparator = "/"

// Ensure nodeName has "clusterId/" prefix.
func nodeNameWithClusterId(clusterId string, nodeName string) string {
	if _, _, err := extractClusterFromNodeName(nodeName); err != nil {
		return clusterId + nodeIdSeparator + nodeName
	}
	return nodeName
}

func addClusterIdToNode(clusterId string, node *corev1.Node) {
	node.Name = nodeNameWithClusterId(clusterId, node.Name)
}

func addClusterIdToPod(clusterId string, pod *corev1.Pod) {
	pod.Spec.NodeName = nodeNameWithClusterId(clusterId, pod.Spec.NodeName)
}

func extractClusterFromNodeName(nodeName string) (clusterId string, nodeId string, err error) {
	components := strings.SplitN(nodeName, nodeIdSeparator, 2)
	if len(components) != 2 {
		err = fmt.Errorf("extractClusterFromNodeName: Unable to parse node name %q", nodeName)
		return
	}
	clusterId = components[0]
	nodeId = components[1]
	return
}
