/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pools

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/codes"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	PoolMappingChangeInProgressRequeueTimeInSec = 5
	PoolMappingChangeErrorRequeueTimeInSec      = 10
	PoolMappingChangeResyncRequeueTimeInSec     = 300

	NodeGroupPoolChangeTimeOutInSec = 600
)

// NodeGroupToPoolMappingReconciler reconciles a NodeGroupToPoolMapping object
type NodeGroupToPoolMappingReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	EventRecorder    record.EventRecorder
	MappingEventChan chan MappingEvent
}

//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=nodegrouptopoolmappings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=nodegrouptopoolmappings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=nodegrouptopoolmappings/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NodeGroupToPoolMappingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NodeGroupToPoolMappingReconciler.Reconcile").WithValues(utils.LogFieldResourceId, req.Name).Start()
	defer span.End()
	result, reconcileErr := func() (ctrl.Result, error) {
		mapping := &idcnetworkv1alpha1.NodeGroupToPoolMapping{}
		err := r.Get(ctx, req.NamespacedName, mapping)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("ignoring reconcile request because source CR was not found")
				// the object has already been removed, do an extra check to make sure everything is removed.
				deleteRes, delErr := r.handleDelete(ctx, req.Name)
				if delErr != nil {
					return deleteRes, delErr
				}
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}

		if mapping.ObjectMeta.DeletionTimestamp.IsZero() {
			return r.handleCreateOrUpdate(ctx, mapping)
		} else {
			// the object has been marked to be deleted, but object not yet removed.
			return r.handleDelete(ctx, mapping.Name)
		}
	}()

	if reconcileErr != nil {
		span.SetStatus(codes.Error, reconcileErr.Error())
		logger.Error(reconcileErr, "NodeGroupToPoolMappingReconciler.Reconcile: error reconciling NodeGroupToPoolMapping")
	}

	return result, reconcileErr
}

func (r *NodeGroupToPoolMappingReconciler) handleCreateOrUpdate(ctx context.Context, mapping *idcnetworkv1alpha1.NodeGroupToPoolMapping) (ctrl.Result, error) {
	return r.process(ctx, MAPPING_EVENT_UPDATE, mapping, mapping.Name, mapping.Spec.Pool)
}

func (r *NodeGroupToPoolMappingReconciler) handleDelete(ctx context.Context, nodeGroupName string) (ctrl.Result, error) {
	return r.process(ctx, MAPPING_EVENT_DELETE, nil, nodeGroupName, "")
}

func (r *NodeGroupToPoolMappingReconciler) process(ctx context.Context, eventType MappingEventType, mapping *idcnetworkv1alpha1.NodeGroupToPoolMapping, nodeGroupName string, targetPoolName string) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("NodeGroupToPoolMappingReconciler.process").WithValues(utils.LogFieldNodeGroupName, nodeGroupName)

	logger.V(1).Info(fmt.Sprintf("processing reconciliation for mapping. NodeGroup %v, Pool: %v", nodeGroupName, targetPoolName))

	resCh := make(chan MappingHandleResult)
	if r.MappingEventChan == nil {
		return ctrl.Result{}, fmt.Errorf("MappingEventChan is nil")
	}
	r.MappingEventChan <- MappingEvent{Type: eventType, NodeGroup: nodeGroupName, Pool: targetPoolName, ResCh: resCh}
	select {
	case res, ok := <-resCh:
		if !ok {
			return ctrl.Result{}, fmt.Errorf("channel is closed")
		}

		if res.ResultStatus == MAPPING_EVENT_PROCESS_IN_PROGRESS {
			return ctrl.Result{RequeueAfter: time.Duration(PoolMappingChangeInProgressRequeueTimeInSec) * time.Second}, nil
		} else if res.ResultStatus == MAPPING_EVENT_PROCESS_FAILED {
			logger.Error(fmt.Errorf("mapping process failed"), res.ErrorMessage)
			return ctrl.Result{RequeueAfter: time.Duration(PoolMappingChangeErrorRequeueTimeInSec) * time.Second}, nil
		} else if res.ResultStatus == MAPPING_EVENT_PROCESS_SUCCESS {
			// get the latest mapping CR
			if mapping != nil {
				mappingCopy := mapping.DeepCopy()
				mappingCopy.Status.LastChangeTime = metav1.Now()
				statusUpdateErr := r.Status().Update(ctx, mappingCopy)
				if statusUpdateErr != nil {
					logger.Info(fmt.Sprintf("failed to update status for mapping %v", mapping.Name))
				}
				r.EventRecorder.Event(mapping, corev1.EventTypeNormal, "nodeGroup-to-pool mapping changed", fmt.Sprintf("nodeGroup %v has been moved to pool [%v]", mapping.Name, targetPoolName))
			}
			return ctrl.Result{}, nil
		} else if res.ResultStatus == MAPPING_EVENT_PROCESS_NOOP {
			return ctrl.Result{RequeueAfter: time.Duration(PoolMappingChangeResyncRequeueTimeInSec) * time.Second}, nil
		}
	case <-time.After(NodeGroupPoolChangeTimeOutInSec * time.Second):
		return ctrl.Result{}, fmt.Errorf("timeout waiting for mapping handling, nodeGroup: %v, targetPool: %v", nodeGroupName, targetPoolName)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeGroupToPoolMappingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&idcnetworkv1alpha1.NodeGroupToPoolMapping{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Watches(&idcnetworkv1alpha1.NodeGroup{},
			&nodeGroupEventHandler{}).
		Complete(r)
}

type nodeGroupEventHandler struct{}

func (e *nodeGroupEventHandler) Create(ctx context.Context, evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	q.Add(mapToMappingEvent(evt.Object))
}

func (e *nodeGroupEventHandler) Update(ctx context.Context, evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if !reflect.DeepEqual(evt.ObjectOld.GetLabels(), evt.ObjectNew.GetLabels()) {
		q.Add(mapToMappingEvent(evt.ObjectNew))
	}
}

func (e *nodeGroupEventHandler) Delete(ctx context.Context, evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
}

func (e *nodeGroupEventHandler) Generic(ctx context.Context, evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}

func mapToMappingEvent(obj client.Object) reconcile.Request {
	req := reconcile.Request{}
	nodeGroup, ok := obj.(*idcnetworkv1alpha1.NodeGroup)
	if !ok {
		return req
	}
	key := types.NamespacedName{
		Name:      nodeGroup.Name,
		Namespace: nodeGroup.Namespace,
	}
	req.NamespacedName = key
	return req
}

// GetPoolByGroupName
func (r *NodeGroupToPoolMappingReconciler) GetPoolByGroupName(ctx context.Context, groupName string) (string, error) {
	mapping := &idcnetworkv1alpha1.NodeGroupToPoolMapping{}
	key := types.NamespacedName{Name: groupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}

	err := r.Get(ctx, key, mapping)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			// got a real problem when talking to K8s API.
			return "", err
		} else {
			// Object not found
			return "", nil
		}
	}
	return mapping.Spec.Pool, nil
}

// GetGroupToPoolMappings
func (r *NodeGroupToPoolMappingReconciler) GetGroupToPoolMappings(ctx context.Context) (map[string]string, error) {
	mappings := &idcnetworkv1alpha1.NodeGroupToPoolMappingList{}
	err := r.List(ctx, mappings)
	if err != nil {
		return nil, fmt.Errorf("list mappings failed, %v", err)
	}
	res := make(map[string]string)
	for _, mapping := range mappings.Items {
		res[mapping.Name] = mapping.Spec.Pool
	}

	return res, nil
}

func (r *NodeGroupToPoolMappingReconciler) WatchGroupToPoolMappings() (chan MappingEvent, error) {
	if r.MappingEventChan == nil {
		return nil, fmt.Errorf("mapping event channel is nil")
	}
	return r.MappingEventChan, nil
}
