// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/core/cache"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	vpcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
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
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler reconciles VPC objects from the Network API Server to K8s objects of kind vpc.private.cloud.intel.com.
// It copies lb Status and deletion confirmation (removal of finalizer) from K8s to the Compute API Server.
// See https://docs.bitnami.com/tutorials/kubewatch-an-example-of-kubernetes-custom-controller/
// See https://komodor.com/learn/controller-manager/
type Reconciler struct {
	informer                 toolscache.SharedIndexInformer
	cache                    *cache.Cache
	k8sClient                k8sclient.Client
	vpcClient                pb.VPCPrivateServiceClient
	sdnClient                sdnv1.OvnnetClient
	durationSinceLastSuccess *atomicduration.AtomicDuration
	marshaler                *grpcruntime.JSONPb
	idcRegion                string
}

func NewReconciler(ctx context.Context, mgr ctrl.Manager, grpcClient pb.VPCPrivateServiceClient, sdnClient sdnv1.OvnnetClient, idcRegion string) (*Reconciler, error) {

	durationSinceLastSuccess := atomicduration.New()
	// Create source that reads from GRPC VPCPrivateServiceClient.
	lw := NewVPCListerWatcher(grpcClient, 60*time.Second)
	// Whenever the ListerWatcher receives a Watch response, reset durationSinceLastSuccess so that
	// the health check can detect idleness.
	lw.OnWatchSuccess = durationSinceLastSuccess.Reset
	informer := toolscache.NewSharedIndexInformer(lw, &vpcv1alpha1.VPC{}, 0, toolscache.Indexers{})
	cache := &cache.Cache{
		Informer: informer,
	}
	// Create reconciler.
	r := &Reconciler{
		informer:                 informer,
		cache:                    cache,
		k8sClient:                mgr.GetClient(),
		vpcClient:                grpcClient,
		sdnClient:                sdnClient,
		durationSinceLastSuccess: durationSinceLastSuccess,
		idcRegion:                idcRegion,
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
	c, err := controller.New("vpc_reconciler", mgr, controllerOptions)
	if err != nil {
		return nil, err
	}
	// Connect sources to manager.
	src := source.Kind(cache, &vpcv1alpha1.VPC{})
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
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("VPCReconciler.Run").Start()
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
	log := log.FromContext(ctx).WithName("VPCReconciler.Healthz")
	lastSuccessAge := r.durationSinceLastSuccess.SinceReset()
	log.Info("Checking health", "lastSuccessAge", lastSuccessAge)
	if lastSuccessAge > 10*time.Second {
		return fmt.Errorf("last success was %s ago", lastSuccessAge)
	}
	return nil
}

// Reconcile is called by the controller runtime when a create/update/delete event occurs
// in the Network API Server.
// req contains only the namespace (CloudAccountId) and name (ResourceId).
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("VPCReconciler.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("Reconciling vpc")

	result, reconcileErr := func() (ctrl.Result, error) {

		// Fetch the vpc from the Network API Server (Postgres).
		vpc, err := r.getSourceVPC(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if vpc == nil {
			log.Info("Ignoring reconcile request because source vpc was not found in cache")
			return ctrl.Result{}, nil
		}

		result, processErr := func() (ctrl.Result, error) {
			if vpc.Metadata.DeletionTimestamp == nil {
				return r.processVPC(ctx, vpc)
			} else {
				return r.processDeleteVPC(ctx, vpc)
			}
		}()

		// Update the status of the resource
		if err := r.updateStatus(ctx, vpc); err != nil {
			return ctrl.Result{}, err
		}

		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "error reconciling vpc")
	}
	log.Info("VPCReconciler.Reconcile: Completed", logkeys.Result, result, logkeys.Error, reconcileErr)
	// If an error occurs, the controller runtime will schedule a retry.
	return result, reconcileErr
}

func (r *Reconciler) updateStatus(ctx context.Context, vpc *pb.VPCPrivate) error {
	vpcOrig, err := r.vpcClient.GetPrivate(ctx, &pb.VPCGetPrivateRequest{
		Metadata: &pb.VPCMetadataReference{
			NameOrId: &pb.VPCMetadataReference_ResourceId{
				ResourceId: vpc.Metadata.ResourceId,
			},
			CloudAccountId:  vpc.Metadata.CloudAccountId,
			ResourceVersion: vpc.Metadata.ResourceVersion,
		},
	})
	if err != nil {
		return err
	}

	// Check if the current status is different than the one in the DB, if so, then update the status.
	if vpc.Status.Message != vpcOrig.Status.Message ||
		vpc.Status.Phase != vpcOrig.Status.Phase {

		// Update the status of the resource
		_, err = r.vpcClient.UpdateStatus(ctx, &pb.VPCUpdateStatusRequest{
			Metadata: &pb.VPCIdReference{
				CloudAccountId:  vpc.Metadata.CloudAccountId,
				ResourceId:      vpc.Metadata.ResourceId,
				ResourceVersion: vpc.Metadata.ResourceVersion,
			},
			Status: vpc.Status,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) processVPC(ctx context.Context, vpc *pb.VPCPrivate) (reconcile.Result, error) {
	// Determine current VPC status
	vpcGetRequest := &sdnv1.GetVPCRequest{
		VpcId: &sdnv1.VPCId{
			Uuid: vpc.Metadata.ResourceId,
		},
	}
	_, err := r.sdnClient.GetVPC(ctx, vpcGetRequest)

	// Check if the VPC
	if status.Code(err) == codes.NotFound {

		// Send request to SDN controller
		vpcCreateRequest := &sdnv1.CreateVPCRequest{
			VpcId: &sdnv1.VPCId{
				Uuid: vpc.Metadata.ResourceId,
			},
			Name:     vpc.Metadata.Name,
			TenantId: vpc.Metadata.CloudAccountId,
			RegionId: r.idcRegion,
		}
		_, err = r.sdnClient.CreateVPC(ctx, vpcCreateRequest)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("could not create vpc: %v", err)
		}

		// Set status to ready
		vpc.Status.Message = "VPC ready"
		vpc.Status.Phase = pb.VPCPhase_VPCPhase_Ready

		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not status vpc: %v", err)
	}

	// Process any updates
	// -------------------

	// Set status to ready
	vpc.Status.Message = "VPC ready"
	vpc.Status.Phase = pb.VPCPhase_VPCPhase_Ready

	return ctrl.Result{}, nil
}

func (r *Reconciler) processDeleteVPC(ctx context.Context, vpc *pb.VPCPrivate) (reconcile.Result, error) {
	return ctrl.Result{}, nil
}

// Get VPC from Network API Server.
// Returns (nil, nil) if not found.
func (r *Reconciler) getSourceVPC(ctx context.Context, req ctrl.Request) (*pb.VPCPrivate, error) {

	cachedObject, exists, err := r.informer.GetStore().GetByKey(req.NamespacedName.String())
	if err != nil {
		return nil, fmt.Errorf("getSourceVPC error: %w", err)
	}
	if !exists {
		return nil, nil
	}
	vpc, ok := cachedObject.(*vpcv1alpha1.VPC)
	if !ok {
		return nil, fmt.Errorf("getSourceVPC error: unexpected type of cached object")
	}

	spec := &pb.VPCSpecPrivate{}
	if err := r.marshaler.Unmarshal([]byte(vpc.Spec), spec); err != nil {
		return nil, fmt.Errorf("VPCReconciler.getSourceVPC: Spec: %w", err)
	}

	status := &pb.VPCStatusPrivate{}
	if len(vpc.Status) > 0 {
		if err := r.marshaler.Unmarshal([]byte(vpc.Status), status); err != nil {
			return nil, fmt.Errorf("VPCReconciler.getSourceVPC: Status: %w", err)
		}
	}

	vpcPB := &pb.VPCPrivate{
		Metadata: &pb.VPCMetadataPrivate{
			ResourceId:        vpc.ObjectMeta.Name,
			CloudAccountId:    vpc.ObjectMeta.Labels["cloud-account-id"],
			Name:              vpc.ObjectMeta.Name,
			ResourceVersion:   vpc.ObjectMeta.ResourceVersion,
			Labels:            vpc.ObjectMeta.Labels,
			CreationTimestamp: fromK8sTimestamp(&vpc.ObjectMeta.CreationTimestamp),
			DeletionTimestamp: fromK8sTimestamp(vpc.ObjectMeta.DeletionTimestamp),
		},
		Spec:   spec,
		Status: status,
	}

	return vpcPB, nil
}

// Convert generic struct to JSON then to Protobuf.
func (r *Reconciler) toPb(source any, target any) error {
	jsonBytes, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("VPCReconciler.toPb: unable to serialize to json: %w", err)
	}
	if err := r.marshaler.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("VPCReconciler.toPb: unable to deserialize from json: %w", err)
	}
	return nil
}

// TODO: (SAS) Move to central spot
func fromK8sTimestamp(t *metav1.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.Time)
}
