// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package subnet

import (
	"context"
	"fmt"
	"net/http"
	"time"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/core/cache"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	subnetv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
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
	subnetClient             pb.SubnetPrivateServiceClient
	sdnClient                sdnv1.OvnnetClient
	durationSinceLastSuccess *atomicduration.AtomicDuration
	marshaler                *grpcruntime.JSONPb
}

func NewReconciler(ctx context.Context, mgr ctrl.Manager, grpcClient pb.SubnetPrivateServiceClient,
	sdnClient sdnv1.OvnnetClient) (*Reconciler, error) {

	durationSinceLastSuccess := atomicduration.New()
	// Create source that reads from GRPC SubnetPrivateServiceClient.
	lw := NewListerWatcher(grpcClient, 60*time.Second)
	// Whenever the ListerWatcher receives a Watch response, reset durationSinceLastSuccess so that
	// the health check can detect idleness.
	lw.OnWatchSuccess = durationSinceLastSuccess.Reset
	informer := toolscache.NewSharedIndexInformer(lw, &subnetv1alpha1.Subnet{}, 0, toolscache.Indexers{})
	cache := &cache.Cache{
		Informer: informer,
	}
	// Create replicator.
	r := &Reconciler{
		informer:                 informer,
		cache:                    cache,
		k8sClient:                mgr.GetClient(),
		subnetClient:             grpcClient,
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
	c, err := controller.New("subnet_reconciler", mgr, controllerOptions)
	if err != nil {
		return nil, err
	}
	// Connect sources to manager.
	src := source.Kind(cache, &subnetv1alpha1.Subnet{})
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
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetReconciler.Run").Start()
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
	log := log.FromContext(ctx).WithName("SubnetReconciler.Healthz")
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
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SubnetReconciler.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("Reconciling subnet")
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the subnet from the Compute API Server (Postgres).
		subnet, err := r.getSourceSubnet(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if subnet == nil {
			log.Info("Ignoring reconcile request because source subnet was not found in cache")
			return ctrl.Result{}, nil
		}

		result, processErr := func() (ctrl.Result, error) {
			if subnet.Metadata.DeletionTimestamp == nil {
				return r.processSubnet(ctx, subnet)
			} else {
				return r.processDeleteSubnet(ctx, subnet)
			}
		}()

		// Update the status of the resource
		if err := r.updateStatus(ctx, subnet); err != nil {
			return ctrl.Result{}, err
		}

		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling subnet")
	}
	log.Info("SubnetReconciler.Reconcile: Completed", logkeys.Result, result, logkeys.Error, reconcileErr)
	// If an error occurs, the controller runtime will schedule a retry.
	return result, reconcileErr
}

func (r *Reconciler) updateStatus(ctx context.Context, subnet *pb.SubnetPrivate) error {

	subnetOrig, err := r.subnetClient.GetPrivate(ctx, &pb.SubnetGetPrivateRequest{
		Metadata: &pb.SubnetMetadataReference{
			NameOrId: &pb.SubnetMetadataReference_ResourceId{
				ResourceId: subnet.Metadata.ResourceId,
			},
			CloudAccountId:  subnet.Metadata.CloudAccountId,
			ResourceVersion: subnet.Metadata.ResourceVersion,
		},
	})
	if err != nil {
		return err
	}

	// Check if the current status is different than the one in the DB, if so, then update the status.
	if subnet.Status.Message != subnetOrig.Status.Message ||
		subnet.Status.Phase != subnetOrig.Status.Phase {

		// Update the status of the resource
		_, err = r.subnetClient.UpdateStatus(ctx, &pb.SubnetUpdateStatusRequest{
			Metadata: &pb.SubnetIdReference{
				CloudAccountId:   subnet.Metadata.CloudAccountId,
				ResourceId:       subnet.Metadata.ResourceId,
				ResourceVersion:  subnet.Metadata.ResourceVersion,
				DeletedTimestamp: subnet.Metadata.DeletedTimestamp,
			},
			Status: subnet.Status,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) processSubnet(ctx context.Context, subnet *pb.SubnetPrivate) (reconcile.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SubnetReconciler.processSubnet").WithValues(logkeys.ResourceId, subnet.Metadata.Name).Start()
	defer span.End()

	// Set status to Provisioning since it's an update.
	subnet.Status.Message = "Provisioning"
	subnet.Status.Phase = pb.SubnetPhase_SubnetPhase_Provisioning

	// Determine current subnet status
	subnetGetRequest := &sdnv1.GetSubnetRequest{
		SubnetId: &sdnv1.SubnetId{
			Uuid: subnet.Metadata.ResourceId,
		},
	}
	upstreamSubnet, err := r.sdnClient.GetSubnet(ctx, subnetGetRequest)
	// Check if the subnet exists
	if status.Code(err) == codes.NotFound {

		log.Info("subnet does not exist, create", logkeys.Subnet, subnet.Metadata.Name)

		// Send request to SDN controller
		subnetRequest := &sdnv1.CreateSubnetRequest{
			VpcId: &sdnv1.VPCId{
				Uuid: subnet.Spec.VpcId,
			},
			Name:             subnet.Metadata.Name,
			Cidr:             subnet.Spec.CidrBlock,
			AvailabilityZone: subnet.Spec.AvailabilityZone,
			SubnetId:         &sdnv1.SubnetId{Uuid: subnet.Metadata.ResourceId},
		}
		_, err := r.sdnClient.CreateSubnet(ctx, subnetRequest)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("could not create subnet: %v", err)
		}
		// Set status to ready
		subnet.Status.Message = "Subnet ready"
		subnet.Status.Phase = pb.SubnetPhase_SubnetPhase_Ready

		return ctrl.Result{}, err
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not status subnet: %v", err)
	}

	// If the SDN controller found the existing subnet, update any changes
	if upstreamSubnet != nil {
		log.Info("subnet exists, updating", logkeys.Subnet, subnet.Metadata.Name)
		// Process any updates
		// -------------------
	}

	// Set status to ready
	subnet.Status.Message = "Subnet ready"
	subnet.Status.Phase = pb.SubnetPhase_SubnetPhase_Ready

	return ctrl.Result{}, nil
}

func (r *Reconciler) processDeleteSubnet(ctx context.Context, subnet *pb.SubnetPrivate) (reconcile.Result, error) {
	// TODO: remove any assigned security groups
	// TODO: remove router interface.

	// remove subnet from sdn.
	_, err := r.sdnClient.DeleteSubnet(ctx, &sdnv1.DeleteSubnetRequest{
		SubnetId: &sdnv1.SubnetId{
			Uuid: subnet.Metadata.ResourceId,
		},
	})

	if err != nil {
		// If the subnet is not found, it is already deleted in sdn.
		// therefor, ignore the error and continue.
		if status.Code(err) != codes.NotFound {
			return ctrl.Result{}, fmt.Errorf("could not delete subnet: %v", err)
		}
	}

	// Update the subnet in the db too
	subnet.Metadata.DeletedTimestamp = timestamppb.Now()
	subnet.Status.Message = "Deleted"
	subnet.Status.Phase = pb.SubnetPhase_SubnetPhase_Deleted

	return ctrl.Result{}, err
}

// Get Subnet from Network API Server.
// Returns (nil, nil) if not found.
func (r *Reconciler) getSourceSubnet(ctx context.Context, req ctrl.Request) (*pb.SubnetPrivate, error) {

	cachedObject, exists, err := r.informer.GetStore().GetByKey(req.NamespacedName.String())
	if err != nil {
		return nil, fmt.Errorf("getSourceSubnet error: %w", err)
	}
	if !exists {
		return nil, nil
	}
	subnet, ok := cachedObject.(*subnetv1alpha1.Subnet)
	if !ok {
		return nil, fmt.Errorf("getSourceSubnet error: unexpected type of cached object")
	}

	spec := &pb.SubnetSpecPrivate{}
	if err := r.marshaler.Unmarshal([]byte(subnet.Spec), spec); err != nil {
		return nil, fmt.Errorf("SubnetReconciler.getSourceSubnet: Spec: %w", err)
	}

	status := &pb.SubnetStatusPrivate{}
	if len(subnet.Status) > 0 {
		if err := r.marshaler.Unmarshal([]byte(subnet.Status), status); err != nil {
			return nil, fmt.Errorf("SubnetReconciler.getSourceSubnet: Status: %w", err)
		}
	}

	subnetPB := &pb.SubnetPrivate{
		Metadata: &pb.SubnetMetadataPrivate{
			CloudAccountId:    subnet.ObjectMeta.Labels["cloud-account-id"],
			ResourceId:        subnet.ObjectMeta.Name,
			Name:              subnet.ObjectMeta.Name, //TODO (SAS): Name is the same as resourceId
			Labels:            subnet.ObjectMeta.Labels,
			CreationTimestamp: fromK8sTimestamp(&subnet.ObjectMeta.CreationTimestamp),
			DeletionTimestamp: fromK8sTimestamp(subnet.ObjectMeta.DeletionTimestamp),
		},
		Spec:   spec,
		Status: status,
	}

	return subnetPB, nil
}

// TODO: (SAS) Move to central spot
func fromK8sTimestamp(t *metav1.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.Time)
}
