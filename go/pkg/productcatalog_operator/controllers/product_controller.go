// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controllers

import (
	"context"
	"fmt"
	"os"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	tradecheck "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tradecheck/tradecheckintel"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/apis/private.cloud/v1alpha1"
)

// ProductReconciler reconciles a Product object
type ProductReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	BillingSyncClient pb.BillingProductCatalogSyncServiceClient
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=products,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=products/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=products/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Product object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ProductReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductReconciler.Reconcile").Start()
	defer span.End()
	log.Info("Reconciling Product: Started")

	result, err := func() (ctrl.Result, error) {
		product := &cloudv1alpha1.Product{}
		var productStatus cloudv1alpha1.ProductStatus
		ctxErr := r.Get(context.TODO(), req.NamespacedName, product)
		if ctxErr != nil {
			// if product is already deleted, no need to update status
			if errors.IsNotFound(ctxErr) {
				log.Info("product seems to be deleted", "namespace", req.Namespace, "name", req.Name)
				// nothing to do here
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, fmt.Errorf("error encountered in fetching product context: %v", ctxErr)
		}

		p := tradecheck.Product{
			ProductHeader: tradecheck.ProductHeader{
				ProductNumber:      product.Spec.ID,
				ProductDescription: product.Spec.Description,
				LogicalSystem:      "IDC",
				PCQ_ID:             product.Spec.PCQ,
				ECCN:               product.Spec.ECCN,
			},
		}

		usernamefile, exists := os.LookupEnv("usernameFile")
		if !exists {
			return ctrl.Result{}, fmt.Errorf("usernamefile name not found")
		}
		passwordfile, exists := os.LookupEnv("passwordFile")
		if !exists {
			return ctrl.Result{}, fmt.Errorf("passwordfile name not found")
		}
		token_url, exists := os.LookupEnv("gts_get_token_url")
		if !exists {
			return ctrl.Result{}, fmt.Errorf("GTS Get Token URL not found")
		}
		create_product_url, exists := os.LookupEnv("gts_create_product_url")
		if !exists {
			return ctrl.Result{}, fmt.Errorf("GTS Create Product URL not found")
		}

		cfg, err := tradecheck.CreateConfig(usernamefile, passwordfile, token_url, create_product_url, "", "")
		if err != nil {
			return ctrl.Result{}, err
		}

		client, err := tradecheck.NewClient(cfg)
		if err != nil {
			return ctrl.Result{}, err
		}

		err = client.CreateProduct(ctx, p)
		if err != nil {
			productStatus.State = cloudv1alpha1.ProductStateError
			errStatus := updateProductStatus(ctx, r, product, productStatus)
			if errStatus != nil {
				log.Error(errStatus, "Error in updating status of the product")
				return ctrl.Result{}, fmt.Errorf("Error in updating status of the product: %v", errStatus)
			}
			log.Error(err, "Error encountered in Create Product, product state set to error")
			return ctrl.Result{}, fmt.Errorf("GTS Create Product failed: %v", err)
		} else {
			productStatus.State = cloudv1alpha1.ProductStateReady
		}

		// invoke sync API here if billing sync client is non nil
		// billing sync method returns  (*emptypb.Empty, error)
		if r.BillingSyncClient != nil {
			_, syncErr := r.BillingSyncClient.Sync(ctx, &emptypb.Empty{})
			if syncErr != nil {
				log.Error(syncErr, "Error encountered in billing productcatalog sync, product state set to errored")
				return ctrl.Result{}, fmt.Errorf("error encountered in billing productcatalog sync: %v", syncErr)
			}
		}

		err = updateProductStatus(ctx, r, product, productStatus)
		if err != nil {
			log.Error(err, "Error in updating status of the product")
			return ctrl.Result{}, fmt.Errorf("Error in updating status of the product: %v", err)
		}

		log.Info("Reconciling Product: Completed")
		return ctrl.Result{}, nil
	}()
	if err != nil {
		log.Error(err, "Error Reconciling Product")
		return result, err
	}

	return result, nil
}

func updateProductStatus(ctx context.Context, r *ProductReconciler,
	product *cloudv1alpha1.Product,
	productStatus cloudv1alpha1.ProductStatus) error {

	log := log.FromContext(ctx)

	log.Info("updating the status for the product")
	product.Status = productStatus

	if err := r.Status().Update(ctx, product); err != nil {
		return fmt.Errorf("updateStatusCondition error: %v", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProductReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// ref: https://github.com/kubernetes-sigs/kubebuilder/issues/618#issuecomment-895027532
		For(&cloudv1alpha1.Product{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
