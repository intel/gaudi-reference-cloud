// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iprm

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/core/cache"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	portv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
	sdnv1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/sdn"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/atomicduration"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	toolscache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// SubnetReconciler reconciles Subnet objects from the Network API Server to K8s objects of kind subnet.private.cloud.intel.com.
// It copies lb Status and deletion confirmation (removal of finalizer) from K8s to the Compute API Server.
// See https://docs.bitnami.com/tutorials/kubewatch-an-example-of-kubernetes-custom-controller/
// See https://komodor.com/learn/controller-manager/
type Reconciler struct {
	informer                 toolscache.SharedIndexInformer
	cache                    *cache.Cache
	k8sClient                k8sclient.Client
	iprmClient               pb.IPRMPrivateServiceClient
	sdnClient                sdnv1.OvnnetClient
	durationSinceLastSuccess *atomicduration.AtomicDuration
	marshaler                *grpcruntime.JSONPb
}

func NewReconciler(ctx context.Context, mgr ctrl.Manager, grpcClient pb.IPRMPrivateServiceClient,
	sdnClient sdnv1.OvnnetClient) (*Reconciler, error) {
	durationSinceLastSuccess := atomicduration.New()
	// Create source that reads from GRPC IPRMPrivateServiceClient.
	lw := NewListerWatcher(grpcClient, 60*time.Second)
	// Whenever the ListerWatcher receives a Watch response, reset durationSinceLastSuccess so that
	// the health check can detect idleness.
	lw.OnWatchSuccess = durationSinceLastSuccess.Reset
	informer := toolscache.NewSharedIndexInformer(lw, &portv1alpha1.Port{}, 0, toolscache.Indexers{})
	cache := &cache.Cache{
		Informer: informer,
	}
	// Create replicator.
	r := &Reconciler{
		informer:                 informer,
		cache:                    cache,
		k8sClient:                mgr.GetClient(),
		iprmClient:               grpcClient,
		sdnClient:                sdnClient,
		durationSinceLastSuccess: durationSinceLastSuccess,
		marshaler: &grpcruntime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				// Force fields with default values, including for enums.
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	}
	// Create controller.
	controllerOptions := controller.Options{
		Reconciler: r,
	}
	c, err := controller.New("iprm_reconciler", mgr, controllerOptions)
	if err != nil {
		return nil, err
	}
	// Connect sources to manager.
	src := source.Kind(cache, &portv1alpha1.Port{})
	if err := c.Watch(src, &handler.EnqueueRequestForObject{},
		predicate.Or(predicate.ResourceVersionChangedPredicate{}, predicate.AnnotationChangedPredicate{})); err != nil {
		return nil, err
	}
	// Ensure manager runs informer.
	err = mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		return r.Run(ctx)
	}))
	if err != nil {
		return nil, err
	}
	if err := mgr.AddHealthzCheck("healthz", r.Healthz); err != nil {
		return r, fmt.Errorf("unable to set up health check: %w", err)
	}
	return r, nil
}

func (r *Reconciler) Run(ctx context.Context) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMReconciler.Run").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	defer utilruntime.HandleCrash()
	log.Info("Service running")
	return r.cache.Start(ctx)
}

// Liveness check. Returns success (nil) if the service recently received a Watch response.
func (r *Reconciler) Healthz(req *http.Request) error {
	ctx := req.Context()
	log := log.FromContext(ctx).WithName("IPRMReconciler.Healthz")
	lastSuccessAge := r.durationSinceLastSuccess.SinceReset()
	log.Info("Checking health", "lastSuccessAge", lastSuccessAge)
	if lastSuccessAge > 10*time.Second {
		return fmt.Errorf("last success was %s ago", lastSuccessAge)
	}
	return nil
}

// Reconcile is called by the controller runtime when an create/update/delete event occurs
// in the Network API Server or K8s.
// req contains only the namespace (CloudAccountId) and name (ResourceId).
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IPRMReconciler.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("Reconciling iprm ")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the port from the Compute API Server (Postgres).
		port, err := r.getSourcePort(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if port == nil {
			log.Info("Ignoring reconcile request because source port was not found in cache")
			return ctrl.Result{}, nil
		}

		result, processErr := func() (ctrl.Result, error) {
			if port.Metadata.DeletionTimestamp == nil {
				return r.processPort(ctx, port)
			} else {
				return r.processDeletePort(ctx, port)
			}
		}()

		// Update the status of the resource
		if err := r.updateStatus(ctx, port); err != nil {
			return ctrl.Result{}, err
		}

		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling port")
	}

	log.Info("IPRMReconciler.Reconcile: Completed", logkeys.Result, result, logkeys.Error, reconcileErr)
	//If an error occurs, the controller runtime will schedule a retry.
	return result, reconcileErr
}

func (r *Reconciler) updateStatus(ctx context.Context, port *pb.PortPrivate) error {
	portOrig, err := r.iprmClient.GetPortPrivate(ctx, &pb.GetPortPrivateRequest{
		Metadata: &pb.PortMetadataReference{
			CloudAccountId: port.Metadata.CloudAccountId,
			ResourceId:     port.Metadata.ResourceId,
		},
	})

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(port.Status, portOrig.Status) {

		// Update the status of the resource
		_, err = r.iprmClient.UpdateStatus(ctx, &pb.PortUpdateStatusRequest{
			Metadata: &pb.PortMetadataReference{
				CloudAccountId: port.Metadata.CloudAccountId,
				ResourceId:     port.Metadata.ResourceId,
			},
			Status: port.Status,
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) processPort(ctx context.Context, port *pb.PortPrivate) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IPRMReconciler.processSubnet").WithValues(logkeys.ResourceId, port.Metadata.ResourceId).Start()
	defer span.End()

	// Determine current port status
	getPortRequest := &sdnv1.GetPortRequest{
		PortId: &sdnv1.PortId{
			Uuid: port.Metadata.ResourceId,
		},
	}
	upstreamPort, err := r.sdnClient.GetPort(ctx, getPortRequest)

	// Check if the subnet exists
	if status.Code(err) == codes.NotFound {
		log.Info("port does not exist, create", logkeys.PORT, port.Metadata.ResourceId)

		// Convert chassisId from string to int.
		// TODO: remove conversion when sdn will change chassisId to string
		chassisId, _ := strconv.Atoi(port.Spec.ChassisId)

		// Send request to SDN controller
		portRequest := &sdnv1.CreatePortRequest{
			PortId: &sdnv1.PortId{
				Uuid: port.Metadata.ResourceId,
			},
			SubnetId: &sdnv1.SubnetId{
				Uuid: port.Spec.SubnetId,
			},
			// TODO: add the other params
			ChassisId: uint32(chassisId),
			//DeviceId:   port.Spec.DeviceId,
			MACAddress: port.Spec.MacAddress,
		}

		_, err := r.sdnClient.CreatePort(ctx, portRequest)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("could not create port: %v", err)
		}
		// Set status to ready
		port.Status.Message = "Port ready"
		port.Status.Phase = pb.PortPhase_PortPhase_Ready

		return ctrl.Result{}, err
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not status subnet: %v", err)
	}
	// If the SDN controller found the existing subnet, update any changes
	if upstreamPort != nil {
		log.Info("port exists, updating", logkeys.PORT, port.Metadata.ResourceId)
		// Process any updates
		// -------------------
	}

	// Set status to ready
	port.Status.Message = "Port ready"
	port.Status.Phase = pb.PortPhase_PortPhase_Ready

	return ctrl.Result{}, nil
}

func (r *Reconciler) processDeletePort(ctx context.Context, port *pb.PortPrivate) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IPRMReconciler.processDeletePort").WithValues(logkeys.ResourceId, port.Metadata.ResourceId).Start()
	defer span.End()

	// TODO: remove any assigned security groups
	// TODO: remove router interface.

	// remove the port from the sdn.
	_, err := r.sdnClient.DeletePort(ctx, &sdnv1.DeletePortRequest{
		PortId: &sdnv1.PortId{
			Uuid: port.Metadata.ResourceId,
		},
	})

	if err != nil {
		// If the port is not found, it is already deleted in sdn.
		// therefore, ignore the error and continue.
		log.Info("port is not exists in sdn ", logkeys.PORT, port.Metadata.ResourceId)
		if status.Code(err) != codes.NotFound {
			return ctrl.Result{}, fmt.Errorf("could not delete port - %s: %w", port.Metadata.ResourceId, err)
		}
	}

	// Update the subnet in the db too
	port.Metadata.DeletedTimestamp = timestamppb.Now()
	port.Status.Message = "Deleted"
	port.Status.Phase = pb.PortPhase_PortPhase_Deleted

	return ctrl.Result{}, err
}

// Get Port from Network API Server.
// func (r *Reconciler) getSourceSubnet(ctx context.Context, req ctrl.Request) (*pb.SubnetPrivate, error) {
func (r *Reconciler) getSourcePort(ctx context.Context, req ctrl.Request) (*pb.PortPrivate, error) {
	cachedObject, exists, err := r.informer.GetStore().GetByKey(req.NamespacedName.String())
	if err != nil {
		return nil, fmt.Errorf("getSourceSubnet error: %w", err)
	}
	if !exists {
		return nil, nil
	}
	port, ok := cachedObject.(*portv1alpha1.Port)
	if !ok {
		return nil, fmt.Errorf("getSourcePort error: unexpected type of cached object")
	}

	spec := &pb.PortSpecPrivate{}
	if err := r.marshaler.Unmarshal([]byte(port.Spec), spec); err != nil {
		return nil, fmt.Errorf("IPRMReconciler.getSourceSubnet: Spec: %w", err)
	}

	status := &pb.PortStatusPrivate{}
	if len(port.Status) > 0 {
		if err := r.marshaler.Unmarshal([]byte(port.Status), status); err != nil {
			return nil, fmt.Errorf("IPRMReconciler.getSourceport: Status: %w", err)
		}
	}

	portPB := &pb.PortPrivate{
		Metadata: &pb.PortMetadataPrivate{
			CloudAccountId:    port.ObjectMeta.Labels["cloud-account-id"],
			ResourceId:        port.ObjectMeta.Name,
			Name:              port.ObjectMeta.Name, //TODO (SAS): Name is the same as resourceId
			CreationTimestamp: fromK8sTimestamp(&port.ObjectMeta.CreationTimestamp),
			DeletionTimestamp: fromK8sTimestamp(port.ObjectMeta.DeletionTimestamp),
		},
		Spec:   spec,
		Status: status,
	}

	return portPB, nil
}

// TODO: (SAS) Move to central spot
func fromK8sTimestamp(t *metav1.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.Time)
}
