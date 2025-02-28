// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package processor

import (
	"context"
	"fmt"
	"reflect"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Processor stores the current configuration for all Loadbalancers in the system
type Processor struct {
	client             client.Client
	LBProvider         provider.Provider
	Scheme             *runtime.Scheme
	regionId           string
	availabilityZoneId string
}

func NewProcessor(client client.Client, LBProvider provider.Provider, scheme *runtime.Scheme, regionId string, availabilityZoneId string) *Processor {
	return &Processor{
		client:             client,
		LBProvider:         LBProvider,
		Scheme:             scheme,
		regionId:           regionId,
		availabilityZoneId: availabilityZoneId,
	}
}

// ProcessDeleteLoadbalancer is responsible for deleting a loadbalancer and its corresponding firewall rules.
//
//	Workflow:
//	 - Delete any associated Firewall Rules. Marking them as deleted will cause the Firewall Oeprator
//	   to remove the rules from the IDC FW API
//	 - If the FW rules still exist return and wait for them to be fully terminated
//	 - Once the rules are removed then remove the Loadbalancer Finalizer from the FW Rules
//	 - Remove the Loadbalancer Finalizer from any Instances that are attached to this LB
//	 - Remove the LB resources from the API (Virtual Server(s), Pool(s), Pool Assocation(s))
func (p *Processor) ProcessDeleteLoadbalancer(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer) (bool, error) {
	log := log.FromContext(ctx)

	log.Info("The deletion Time stamp Exist", "Deletion timestamp", loadbalancer.DeletionTimestamp.GoString())

	// Set the status to deleting
	loadbalancer.Status.State = loadbalancerv1alpha1.DELETING
	loadbalancer.Status.Message = "Deleting listeners and firewall rules"

	// Remove the associated firewall rules
	var fwRules firewallv1alpha1.FirewallRuleList
	err := p.client.List(ctx, &fwRules, client.InNamespace(loadbalancer.Namespace))
	if err != nil {
		log.Error(err, "error looking up firewall rules")
		return false, err
	}

	log.Info("Found fwRules in Namespace", "Rules", len(fwRules.Items))

	allFirewallRulesTerminated := true
	for _, fwRule := range fwRules.Items {

		// Only process rules that match this LB being deleted
		if HasControllerReference(&fwRule, loadbalancer.Name) {

			// If the rule hasn't been deleted, remove it and then wait for it
			// to be processed by the Firewall Operator
			if fwRule.DeletionTimestamp.IsZero() {

				log.Info("deleting fw rule", "rule", fwRule.Name, "namespace", fwRule.Namespace)

				// Delete the object
				err = p.client.Delete(ctx, &fwRule)
				if err != nil {
					log.Error(err, "error deleting firewall rule")
					return false, err
				}
			}

			// Check the status of the rule, if it's not terminated yet then return and
			// wait for the status to update properly.
			// Wait to remove the finalizer until the rule is fully removed via the Firewall Operator
			if meta.IsStatusConditionFalse(fwRule.Status.Conditions, "Terminated") {
				allFirewallRulesTerminated = false
				continue
			}

			// Remove the Loadbalanacer finalizer from the firewall rule.
			if err := p.PersistFirewallRuleFinalizer(ctx, remove, fwRule); err != nil {
				return false, err
			}
		}
	}

	if !allFirewallRulesTerminated {
		log.Info("not processing since firewall rule is not yet removed")
		loadbalancer.Status.Message = "Removing firewall rules"
		return false, nil
	}

	// Delete the load balancer
	updatedStatusMessage, err := p.LBProvider.ProcessFinalizers(ctx, loadbalancer)
	if err != nil {
		log.Error(err, "error processing finalizers: "+updatedStatusMessage)
		return false, err
	}

	// Iterate over each listener
	for _, listener := range loadbalancer.Spec.Listeners {

		// Get the members of the load balancer
		poolMembers, err := p.GetLoadbalancerInstances(ctx, loadbalancer.Namespace, listener)
		if err != nil {
			return false, err
		}

		for _, member := range poolMembers {
			if err := p.PersistInstanceFinalizer(ctx, remove, member); err != nil {
				return false, err
			}
		}
	}

	// remove the finalizer if successful
	loadbalancer.Status.State = loadbalancerv1alpha1.DELETED
	loadbalancer.Status.Message = "Load balancer successfully deleted"
	if err := p.PersistLoadbalancerFinalizer(ctx, remove, *loadbalancer); err != nil {
		return false, err
	}

	log.Info("Successfully removed the finalizers")
	return false, nil
}

// ProcessLoadbalancer handles an update to a Loadbalancer CR. This could be a new object created or an update to an existing.
// This method reconciles the spec defined against the LB & FW APIs.
func (p *Processor) ProcessLoadbalancer(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer) (bool, error) {

	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("LoadbalancerReconciler.processLoadBalancer").WithValues(logkeys.LoadBalancerName, loadbalancer.Name).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Add the finalizer to the LB CRD if it doesn't exist
	if !controllerutil.ContainsFinalizer(loadbalancer, constants.LoadbalancerFinalizer) {
		if err := p.PersistLoadbalancerFinalizer(ctx, add, *loadbalancer); err != nil {
			return false, err
		}

		log.Info("add finalizer: requeuing lb")
		return true, nil
	}

	// Lookup the FirewallRules to calculate status, we can have one per listener in the LB spec.
	fwRules, err := p.getFirewallRules(ctx, loadbalancer)
	if err != nil {
		return false, err
	}

	// Get the overall status of the Loadbalancer, this will be used to inform what needs to be provisioned or updated.
	err = p.LBProvider.GetStatus(ctx, loadbalancer, fwRules)
	if err != nil {
		return false, err
	}

	// Process Firewall Rules
	err = p.provisionFirewallRule(ctx, loadbalancer, fwRules)
	if err != nil {
		return false, err
	}

	// Check if a VIP exists.
	primaryVIPExisting := false
	if loadbalancer.Status.Vip != "" {
		primaryVIPExisting = true
	}

	var allInstancesAssigned []loadbalancerv1alpha1.PoolStatusMember

	// Iterate over each listener
	for _, listener := range loadbalancer.Spec.Listeners {

		listenerStatusConditions := &loadbalancerv1alpha1.ConditionsListenerStatus{}
		listenerStatusConditionsIndex := -1
		for i, l := range loadbalancer.Status.Conditions.Listeners {
			if l.Port == listener.VIP.Port {
				listenerStatusConditions = &l
				listenerStatusConditionsIndex = i
				break
			}
		}

		var listenerStatus *loadbalancerv1alpha1.ListenerStatus
		listenerStatusIndex := -1
		for i, l := range loadbalancer.Status.Listeners {
			if l.Port == listener.VIP.Port {
				listenerStatus = &l
				listenerStatusIndex = i
				break
			}
		}

		if listenerStatus == nil || listenerStatusConditions == nil {
			log.Info("skipping listener process due to nil condition", "listener", listener.VIP.Port)
			continue
		}

		// Process this listener, creating the VIP if needed, creating loadbalancer pool members, attaching sslProfiles,
		// associating them with the VIP, as well as managing Finalizers on Instances
		result, err := p.processListener(ctx, loadbalancer, listener, primaryVIPExisting,
			listenerStatusConditions, listenerStatus)
		if err != nil {
			return false, err
		}

		if len(listenerStatus.PoolMembers) == 0 {
			listenerStatus.Message = "No instances assigned to listener"
		}

		// Update status from processListeners setting it on the root Loadbalancer object.
		loadbalancer.Status.Conditions.Listeners[listenerStatusConditionsIndex] = *listenerStatusConditions
		loadbalancer.Status.Listeners[listenerStatusIndex] = *listenerStatus

		allInstancesAssigned = append(allInstancesAssigned, listenerStatus.PoolMembers...)

		// TODO (SAS): Remove the ILB logic of requeuing the same resource.
		if result.Requeue {
			return true, nil
		}
	}

	// Reconcile the final set of desired Instances, removing finalizers from any that are no longer
	// part of a LB Pool.
	if err := p.ReconcileInstanceFinalizers(ctx, loadbalancer.Namespace, allInstancesAssigned); err != nil {
		return false, err
	}

	// Remove any listeners that might have been deleted. This will determine which Listener has been removed,
	// and update the resource in by removing it. It returns a set of ports which map to the
	// listeners which have been removed.
	removedPorts, err := p.LBProvider.ReconcileListeners(ctx, loadbalancer.Name, loadbalancer.Status.Vip, loadbalancer.Spec.Listeners)
	if err != nil {
		log.Error(err, "error reconciling listeners")
		return false, err
	}

	// Interate over each port removing the FirewallRule.
	for port := range removedPorts {

		// Remove the Firewall Rule
		if err := p.removeFirewallRule(ctx, loadbalancer, port); err != nil {
			return false, err
		}
	}

	return false, nil
}

func (p *Processor) ReconcileInstanceFinalizers(ctx context.Context, namespace string, assignedInstances []loadbalancerv1alpha1.PoolStatusMember) error {

	log := log.FromContext(ctx)

	var allInstances cloudv1alpha1.InstanceList
	if err := p.client.List(ctx, &allInstances, client.InNamespace(namespace)); err != nil {
		log.Error(err, "error looking up instance list for namespace")
		return err
	}

	// Iterate over each existing instance, if it has the LB finalizer, check it's in the
	// list of Instances associated with a LB, otherwise remove the finalizer.
	for _, currentInstance := range allInstances.Items {

		found := false

		for _, assignedInstance := range assignedInstances {
			// Look for the instance by matching on IP
			if len(currentInstance.Status.Interfaces) > 0 {
				if len(currentInstance.Status.Interfaces[0].Addresses) > 0 {
					if currentInstance.Status.Interfaces[0].Addresses[0] == assignedInstance.IPAddress {
						found = true
						break
					}
				}
			}
		}

		// If the instance wasn't found in the current set of Instances, then it's not assigned to any
		// Loadbalancer pools and should have it's finalizer removed.
		if !found {
			// Instance found, check if the Instance contains the LB Finalizer, if so remove it.
			if err := p.PersistInstanceFinalizer(ctx, remove, currentInstance); err != nil {
				return err
			}
			log.Info("Instance finalizer removed", "name", currentInstance.Name, "namespace", currentInstance.Namespace)
		}
	}

	return nil
}

// GetLoadbalancerInstances is called when a Load Balancer is changed, the pool members for
// this load balancer listener are recalculated.
func (p *Processor) GetLoadbalancerInstances(ctx context.Context, namespace string, listener loadbalancerv1alpha1.LoadbalancerListener) ([]cloudv1alpha1.Instance, error) {

	instances, err := p.CalculatePoolMembers(ctx, namespace, listener)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// GetLoadbalancers is called when an instance that is watched changes, created, updated, or deleted.
// When this happens, the pool of instances needs to be updated.
func (p *Processor) GetLoadbalancers(ctx context.Context, instance *cloudv1alpha1.Instance) ([]*loadbalancerv1alpha1.Loadbalancer, error) {

	loadbalancersMap := make(map[string]*loadbalancerv1alpha1.Loadbalancer)
	var loadbalancers []*loadbalancerv1alpha1.Loadbalancer

	// Find all the load balancers in this namespace
	lbList := &loadbalancerv1alpha1.LoadbalancerList{}
	err := p.client.List(ctx, lbList, client.InNamespace(instance.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancers: %v", err)
	}

	for _, lb := range lbList.Items {
		for _, listener := range lb.Spec.Listeners {
			// Determine the load balancer pool type (static or selector)
			if len(listener.Pool.Members) > 0 {
				for _, instancePoolMember := range listener.Pool.Members {
					if instancePoolMember.InstanceResourceId == instance.Name {
						if _, found := loadbalancersMap[lb.Name]; !found {
							loadbalancersMap[lb.Name] = &lb
						}
						break
					}
				}
			} else if len(listener.Pool.InstanceSelectors) > 0 {
				pairsFound := 0

				// Iterate over all the keys in the instance selector
				for k, v := range listener.Pool.InstanceSelectors {
					// Does this key exist in the instance map? If so validate the value matches as well.
					if val, found := instance.Spec.Labels[k]; found {
						if val == v {
							pairsFound++
						}
					}
				}

				// If the number of k/v pairs found match the number of instance selectors,
				// then the instance is part of this lb.
				if pairsFound == len(listener.Pool.InstanceSelectors) {
					if _, found := loadbalancersMap[lb.Name]; !found {
						loadbalancersMap[lb.Name] = &lb
					}
				}
			}
		}
	}

	for _, lb := range loadbalancersMap {
		loadbalancers = append(loadbalancers, lb)
	}

	return loadbalancers, nil
}

func (p *Processor) CalculatePoolMembers(ctx context.Context, namespace string,
	listener loadbalancerv1alpha1.LoadbalancerListener) ([]cloudv1alpha1.Instance, error) {
	instances := make(map[types.NamespacedName]cloudv1alpha1.Instance)

	// Determine the load balancer pool type (static or selector)
	if len(listener.Pool.Members) > 0 {
		// Find all the members of the pool
		instanceList := &cloudv1alpha1.InstanceList{}
		err := p.client.List(ctx, instanceList, client.InNamespace(namespace))
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %v", err)
		}

		for _, instance := range instanceList.Items {
			for _, poolMember := range listener.Pool.Members {

				// Store the instance if it matches a static pool entry
				if instance.Name == poolMember.InstanceResourceId {
					instances[parseNsName(&instance)] = instance
					break
				}
			}
		}
	} else if len(listener.Pool.InstanceSelectors) > 0 {

		// Find all the members of the pool
		instanceList := &cloudv1alpha1.InstanceList{}
		err := p.client.List(ctx, instanceList, client.InNamespace(namespace))
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %v", err)
		}

		for _, instance := range instanceList.Items {
			// If the instance matches keep it as a pool member

			matchingLabels := 0

			// Check if the instance has the labels matching the lb selector
			for k, v := range listener.Pool.InstanceSelectors {
				if val, found := instance.Spec.Labels[k]; found {
					if v == val {
						matchingLabels++
					}
				}
			}

			if matchingLabels == len(listener.Pool.InstanceSelectors) {
				instances[parseNsName(&instance)] = instance
			}
		}
	} else {
		return nil, fmt.Errorf("no pool configuration found in loadbalancer spec")
	}

	instanceList := []cloudv1alpha1.Instance{}
	for _, instance := range instances {
		instanceList = append(instanceList, instance)
	}

	return instanceList, nil
}

// parseNsName extracts the namespace and name from a given Kubernetes object
// and returns them as a NamespacedName.
//
// Parameters:
//
//	obj - The Kubernetes object from which to extract the namespace and name.
//
// Returns:
//
//	A NamespacedName containing the namespace and name of the given object.
func parseNsName(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

// PersistLoadbalancerStatusUpdate updates the status of the given load balancer in the Kubernetes cluster.
// It retrieves the latest version of the load balancer and compares its status with the provided load balancer's status.
// If there is a mismatch, it updates the status of the latest load balancer with the provided status.
// The function uses retry logic to handle conflicts during the update process.
//
// Parameters:
//   - ctx: The context for the operation.
//   - loadbalancer: The load balancer whose status needs to be updated.
//   - namespacedName: The namespaced name of the load balancer.
//
// Returns:
//   - error: An error if the status update fails, otherwise nil.
func (p *Processor) PersistLoadbalancerStatusUpdate(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer, namespacedName types.NamespacedName) error {
	log := log.FromContext(ctx).WithName("LoadbalancerReconciler.PersistLoadbalancerStatusUpdate")
	log.Info("BEGIN")

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestLoadbalancer, err := p.GetLoadbalancer(ctx, namespacedName)
		if err != nil {
			return fmt.Errorf("failed to get the loadbalancer: %+v. error:%w", loadbalancer, err)
		}
		if latestLoadbalancer == nil {
			log.Info("loadbalancer not found", logkeys.LoadBalancer, loadbalancer)
			return nil
		}

		if !equality.Semantic.DeepEqual(loadbalancer.Status, latestLoadbalancer.Status) {
			log.Info("loadbalancer status mismatch", logkeys.LoadBalancerStatus, loadbalancer.Status, logkeys.LatestLoadbalancerStatus, latestLoadbalancer.Status)
			// update latest instance status
			loadbalancer.Status.DeepCopyInto(&latestLoadbalancer.Status)
			if err := p.client.Status().Update(ctx, latestLoadbalancer); err != nil {
				return fmt.Errorf("PersistLoadbalancerStatusUpdate: %w", err)
			}
		} else {
			log.Info("loadbalancer status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update loadbalancer status: %w", err)
	}
	log.Info("END")
	return nil
}

// PersistFirewallRuleStatusUpdate updates the status of a given FirewallRule resource in the Kubernetes cluster.
// It retrieves the latest version of the FirewallRule from the API server and compares its status with the provided
// FirewallRule's status. If there is a mismatch, it updates the status of the latest FirewallRule instance.
// The function uses retry logic to handle conflicts during the update process.
//
// Parameters:
// - ctx: The context for the request.
// - firewallRule: The FirewallRule resource whose status needs to be updated.
// - namespacedName: The namespaced name of the FirewallRule resource.
//
// Returns:
// - error: An error if the status update fails, otherwise nil.
func (p *Processor) PersistFirewallRuleStatusUpdate(ctx context.Context, firewallRule *firewallv1alpha1.FirewallRule, namespacedName types.NamespacedName) error {
	log := log.FromContext(ctx).WithName("LoadbalancerReconciler.PersistLoadbalancerStatusUpdate")
	log.Info("BEGIN")

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestFirewallRule, err := p.GetFirewallRule(ctx, namespacedName)
		if err != nil {
			return fmt.Errorf("failed to get the firewallRule: %+v. error:%w", firewallRule, err)
		}

		// In the event the Firewallrule isn't found, return a "Conflict" error. This will retry
		// until the resource is found or the DefaultRetry is met. This can happen when a FirewallRule
		// is first created, but is not yet fully committed to the k8s API server.
		if latestFirewallRule == nil {
			log.Info("firewall rule not found", logkeys.LoadBalancer, firewallRule)
			return errors.NewConflict(schema.GroupResource{Group: firewallv1alpha1.GroupVersion.Group, Resource: firewallv1alpha1.GroupVersion.Version}, firewallRule.Name, err)
		}

		if !equality.Semantic.DeepEqual(firewallRule.Status, latestFirewallRule.Status) {
			log.Info("firewall rule status mismatch", logkeys.LoadBalancerStatus, firewallRule.Status, logkeys.LatestLoadbalancerStatus, latestFirewallRule.Status)
			// update latest instance status
			firewallRule.Status.DeepCopyInto(&latestFirewallRule.Status)
			if err := p.client.Status().Update(ctx, latestFirewallRule); err != nil {
				return fmt.Errorf("PersistFirewallRuleStatusUpdate: %w", err)
			}
		} else {
			log.Info("firewall rule status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update firewall rule status: %w", err)
	}
	log.Info("END")
	return nil
}

type finalizerOperation string

const (
	add    finalizerOperation = "add"
	remove finalizerOperation = "remove"
)

// PersistInstanceFinalizer ensures that the specified finalizer is added or removed from the given instance.
// It retries the operation on conflict errors.
//
// Parameters:
//   - ctx: The context for the operation.
//   - operation: The finalizer operation to perform (add or remove).
//   - instance: The instance to which the finalizer should be added or removed.
//
// Returns:
//   - error: An error if the operation fails, or nil if it succeeds.
func (p *Processor) PersistInstanceFinalizer(ctx context.Context, operation finalizerOperation, instance cloudv1alpha1.Instance) error {
	log := log.FromContext(ctx).WithName("Processor.PersistInstanceFinalizer")
	log.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		latestInstance, err := p.GetInstance(ctx, types.NamespacedName{
			Name: instance.Name, Namespace: instance.Namespace,
		})
		// instance is deleted
		if latestInstance == nil {
			log.Info("instance not found", "instance", instance)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the instance: %+v. error:%w", instance, err)
		}

		switch operation {
		case add:
			// Check if the finalizer already exists, if it does return.
			if controllerutil.ContainsFinalizer(latestInstance, constants.LoadbalancerFinalizer) {
				return nil
			}
			controllerutil.AddFinalizer(latestInstance, constants.LoadbalancerFinalizer)
		case remove:
			// Check if the finalizer doesn't exist, if it doesn't return.
			if !controllerutil.ContainsFinalizer(latestInstance, constants.LoadbalancerFinalizer) {
				return nil
			}
			controllerutil.RemoveFinalizer(latestInstance, constants.LoadbalancerFinalizer)
		}

		// Use the Client to update since this resource exists in the compute az cluster.
		if err := p.client.Update(ctx, latestInstance); err != nil {
			return fmt.Errorf("PersistInstanceFinalizer: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}
	log.Info("END")
	return nil
}

// PersistLoadbalancerFinalizer ensures that the specified finalizer is added or removed
// from the given Loadbalancer resource. It retries the operation on conflict errors.
//
// Parameters:
//   - ctx: The context for the operation.
//   - operation: The finalizer operation to perform (add or remove).
//   - loadbalancer: The Loadbalancer resource to update.
//
// Returns:
//   - error: An error if the operation fails, otherwise nil.
func (p *Processor) PersistLoadbalancerFinalizer(ctx context.Context, operation finalizerOperation, loadbalancer loadbalancerv1alpha1.Loadbalancer) error {
	log := log.FromContext(ctx).WithName("Processor.PersistLoadbalancerFinalizer")
	log.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		latestLoadbalancer, err := p.GetLoadbalancer(ctx, types.NamespacedName{
			Name: loadbalancer.Name, Namespace: loadbalancer.Namespace,
		})
		// instance is deleted
		if latestLoadbalancer == nil {
			log.Info("loadbalancer not found", "loadbalancer", loadbalancer)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the loadbalancer: %+v. error:%w", loadbalancer, err)
		}

		switch operation {
		case add:
			if controllerutil.ContainsFinalizer(latestLoadbalancer, constants.LoadbalancerFinalizer) {
				return nil
			}
			controllerutil.AddFinalizer(latestLoadbalancer, constants.LoadbalancerFinalizer)
		case remove:
			if !controllerutil.ContainsFinalizer(latestLoadbalancer, constants.LoadbalancerFinalizer) {
				return nil
			}
			controllerutil.RemoveFinalizer(latestLoadbalancer, constants.LoadbalancerFinalizer)
		}

		// Use the Client to update since this resource exists in the compute az cluster.
		if err := p.client.Update(ctx, latestLoadbalancer); err != nil {
			return fmt.Errorf("PersistLoadbalancerFinalizer: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update loadbalancer: %w", err)
	}
	log.Info("END")
	return nil
}

// PersistFirewallRuleFinalizer ensures that the specified finalizer is added or removed
// from the given FirewallRule resource. It retries the operation on conflict errors.
//
// Parameters:
//   - ctx: The context for the operation.
//   - operation: The finalizer operation to perform (add or remove).
//   - fwRule: The FirewallRule resource to update.
//
// Returns:
//   - error: An error if the operation fails, otherwise nil.
func (p *Processor) PersistFirewallRuleFinalizer(ctx context.Context, operation finalizerOperation, fwRule firewallv1alpha1.FirewallRule) error {
	log := log.FromContext(ctx).WithName("Processor.PersistFirewallRuleFinalizer")
	log.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		latestFWRule, err := p.GetFirewallRule(ctx, types.NamespacedName{
			Name: fwRule.Name, Namespace: fwRule.Namespace,
		})
		// instance is deleted
		if latestFWRule == nil {
			log.Info("firewall rule not found", "firewall rule", fwRule)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the firewall rule: %+v. error:%w", fwRule, err)
		}

		switch operation {
		case add:
			if controllerutil.ContainsFinalizer(latestFWRule, constants.LoadbalancerFinalizer) {
				return nil
			}
			controllerutil.AddFinalizer(latestFWRule, constants.LoadbalancerFinalizer)
		case remove:
			if !controllerutil.ContainsFinalizer(latestFWRule, constants.LoadbalancerFinalizer) {
				return nil
			}
			controllerutil.RemoveFinalizer(latestFWRule, constants.LoadbalancerFinalizer)
		}

		// Use the Client to update since this resource exists in the compute az cluster.
		if err := p.client.Update(ctx, latestFWRule); err != nil {
			return fmt.Errorf("PersistFirewallRuleFinalizer: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update firewall rule: %w", err)
	}
	log.Info("END")
	return nil
}

// Get loadbalancer from K8s.
// Returns (nil, nil) if not found.
func (p *Processor) GetLoadbalancer(ctx context.Context, namespacedName types.NamespacedName) (*loadbalancerv1alpha1.Loadbalancer, error) {
	lb := &loadbalancerv1alpha1.Loadbalancer{}
	err := p.client.Get(ctx, namespacedName, lb)
	if errors.IsNotFound(err) || reflect.ValueOf(lb).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetLoadbalancer: %w", err)
	}
	return lb, nil
}

// Get instance from K8s.
// Returns (nil, nil) if not found.
func (p *Processor) GetInstance(ctx context.Context, namespacedName types.NamespacedName) (*cloudv1alpha1.Instance, error) {
	instance := &cloudv1alpha1.Instance{}
	err := p.client.Get(ctx, namespacedName, instance)
	if errors.IsNotFound(err) || reflect.ValueOf(instance).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetInstance: %w", err)
	}
	return instance, nil
}

// Get firewall rule from K8s.
// Returns (nil, nil) if not found.
func (p *Processor) GetFirewallRule(ctx context.Context, namespacedName types.NamespacedName) (*firewallv1alpha1.FirewallRule, error) {
	fwRule := &firewallv1alpha1.FirewallRule{}
	err := p.client.Get(ctx, namespacedName, fwRule)
	if errors.IsNotFound(err) || reflect.ValueOf(fwRule).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetFirewallRule: %w", err)
	}
	return fwRule, nil
}
