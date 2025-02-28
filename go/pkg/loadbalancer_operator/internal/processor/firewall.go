// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package processor

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// getFirewallRules returns all existing Firewall Rules for a tenant which match a specific
// load balancer's set of listeners.
func (p *Processor) getFirewallRules(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer) ([]firewallv1alpha1.FirewallRule, error) {
	log := log.FromContext(ctx)

	fwRules := []firewallv1alpha1.FirewallRule{}

	var allFWRules firewallv1alpha1.FirewallRuleList
	if err := p.client.List(ctx, &allFWRules, client.InNamespace(loadbalancer.Namespace)); err != nil {

		// requeue the result since the firewall rule was found but encountered an error
		log.Info("error getting firewall rules", "error", err)
		return fwRules, err
	}

	// Filter on only the firewall rules for this LB
	for _, rule := range allFWRules.Items {
		if !HasControllerReference(&rule, loadbalancer.Name) {
			log.Info("skipping firewall rule missing controller reference: ", "rule", rule.Name)
			continue
		}
		fwRules = append(fwRules, rule)
	}

	return fwRules, nil
}

// provisionFirewallRule manages firewallrules for a load balancer
func (p *Processor) provisionFirewallRule(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer,
	fwRules []firewallv1alpha1.FirewallRule) error {
	log := log.FromContext(ctx)

	// Check status of Load Balancer, must have a VIP provisioned
	if loadbalancer.Status.Vip == "" {
		log.Info("unable to provision firewall rule due to missing VIP")
		return nil
	}

	clusterRules := make(map[string]firewallv1alpha1.FirewallRule)
	for _, fwRule := range fwRules {
		clusterRules[fwRule.Name] = fwRule
	}

	for _, listener := range loadbalancer.Spec.Listeners {

		// Format the firewallrule name. It follows the pattern of `loadbalancerName-port`
		fwName := fmt.Sprintf("%s-%d", loadbalancer.Name, listener.VIP.Port)

		// Check if a rule already exists, if it does not then create. If it does
		// then update the rule if needed.
		var fwRule firewallv1alpha1.FirewallRule
		if err := p.client.Get(ctx, types.NamespacedName{
			Name:      fwName,
			Namespace: loadbalancer.Namespace,
		}, &fwRule); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}

			log.Info("firewall rule not found, creating")

			// Build new FirewallRule
			fwRule := &firewallv1alpha1.FirewallRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fwName,
					Namespace: loadbalancer.Namespace,
					Labels: map[string]string{
						"control-plane":                "controller-manager",
						"app.kubernetes.io/name":       fwName,
						"app.kubernetes.io/component":  "loadbalancer",
						"app.kubernetes.io/created-by": "loadbalancer-operator",
						"app.kubernetes.io/part-of":    "firewall-operator",
						"app.kubernetes.io/managed-by": "firewall-operator",
						"regionId":                     p.regionId,
						"availabiltyZoneId":            p.availabilityZoneId,
					},
					Finalizers: []string{
						constants.LoadbalancerFinalizer,
					},
				},
				Spec: firewallv1alpha1.FirewallRuleSpec{
					SourceIPs:     loadbalancer.Spec.Security.Sourceips,
					DestinationIP: loadbalancer.Status.Vip,
					Protocol:      firewallv1alpha1.Protocol(listener.VIP.IPProtocol),
					Port:          strconv.Itoa(listener.VIP.Port),
				},
			}

			// Set the controller ownerReferences reference to be the loadbalancer-operator.
			// This allows the LB Operator to watch for changes in any FirewallRule.
			if err := controllerutil.SetControllerReference(loadbalancer, fwRule, p.Scheme); err != nil {
				return err
			}

			// Create the rule
			if err := p.client.Create(ctx, fwRule); err != nil {
				log.Error(err, "error creating firewall rule")
				return err
			}

			// Set the status of the firewall rule
			fwRule.Status = firewallv1alpha1.FirewallRuleStatus{
				State: firewallv1alpha1.RECONCILING,
				Conditions: []metav1.Condition{
					{
						Type:               "Reconciling",
						Status:             metav1.ConditionTrue,
						LastTransitionTime: metav1.Now(),
						Message:            "Firewall Rule is reconciling",
						Reason:             "Reconciling",
					},
					{
						Type:               "Ready",
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             "Ready",
						Message:            "Firewall rules are applied",
					},
					{
						Type:               "Terminated",
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             "Terminated",
						Message:            "Firewall rules are terminated",
					},
				},
			}

			// Set the status on the resource
			if err := p.PersistFirewallRuleStatusUpdate(ctx, fwRule, types.NamespacedName{
				Name: fwName, Namespace: loadbalancer.Namespace}); err != nil {
				log.Error(err, "error updating status on firewall rule")
				return err
			}

		} else {
			// Update the FW rule
			updatedfwRule := fwRule.DeepCopy()
			updatedfwRule.Spec = firewallv1alpha1.FirewallRuleSpec{
				SourceIPs:     loadbalancer.Spec.Security.Sourceips,
				DestinationIP: loadbalancer.Status.Vip,
				Protocol:      firewallv1alpha1.Protocol(listener.VIP.IPProtocol),
				Port:          strconv.Itoa(listener.VIP.Port),
			}

			// Compare the specs and determine if an update is required
			if !reflect.DeepEqual(fwRule.Spec, updatedfwRule.Spec) {
				// Update the rule
				log.Info("updating firewall rule due to changes")
				if err := p.client.Update(ctx, updatedfwRule); err != nil {
					log.Error(err, "error updating firewall rules")
					return err
				}

				// Set the status of the firewall rule
				updatedfwRule.Status = firewallv1alpha1.FirewallRuleStatus{
					State: firewallv1alpha1.RECONCILING,
					Conditions: []metav1.Condition{
						{
							Type:               "Reconciling",
							Status:             metav1.ConditionTrue,
							LastTransitionTime: metav1.Now(),
							Message:            "Firewall Rule is reconciling",
							Reason:             "Reconciling",
						},
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Now(),
							Reason:             "Ready",
							Message:            "Firewall rules are applied",
						},
						{
							Type:               "Terminated",
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Now(),
							Reason:             "Terminated",
							Message:            "Firewall rules are terminated",
						},
					},
				}

				// Set the status on the resource
				if err := p.PersistFirewallRuleStatusUpdate(ctx, updatedfwRule, types.NamespacedName{
					Name: updatedfwRule.Name, Namespace: updatedfwRule.Namespace}); err != nil {
					log.Error(err, "error updating status on firewall rule")
					return err
				}
			}
		}

		// Remove the fwRule from the list marking it as processed
		delete(clusterRules, fwName)
	}

	// Any remaining rules in the list should be deleted
	for _, fwRule := range clusterRules {

		// Delete the object
		if err := p.client.Delete(ctx, &fwRule); err != nil {
			log.Error(err, "error deleting firewall rule")
			return err
		}

		// Remove the finalizer from the firewall rule
		if err := p.PersistFirewallRuleFinalizer(ctx, remove, fwRule); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) removeFirewallRule(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer, port int) error {
	log := log.FromContext(ctx)

	var fwRule firewallv1alpha1.FirewallRule
	err := p.client.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%d", loadbalancer.Name, port)}, &fwRule)
	if err != nil {
		log.Error(err, "error looking up firewall rule")
		return err
	}

	// Delete the object
	err = p.client.Delete(ctx, &fwRule)
	if err != nil {
		log.Error(err, "error deleting firewall rule")
		return err
	}

	if controllerutil.ContainsFinalizer(&fwRule, constants.LoadbalancerFinalizer) {

		// Remove the finalizer from the firewall rule
		controllerutil.RemoveFinalizer(&fwRule, constants.LoadbalancerFinalizer)
		if err := p.client.Update(ctx, &fwRule); err != nil {
			return err
		}
	}

	return err
}

// HasControllerReference returns true if the object
// has an owner ref with controller equal to true
func HasControllerReference(object metav1.Object, loadbalancerName string) bool {
	owners := object.GetOwnerReferences()
	for _, owner := range owners {
		isTrue := owner.Controller
		if owner.Controller != nil &&
			*isTrue &&
			owner.Kind == "Loadbalancer" &&
			owner.Name == loadbalancerName {
			return true
		}
	}

	return false
}
