// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
)

type FirewallProvider interface {
	// SyncFirewallRules compares the current state of the firewall and the desired state and updates
	// the firewall accordingly by creating & removing rules from the Firewall.
	SyncFirewallRules(ctx context.Context, rules []v1alpha1.FirewallRule, existingRules []Rule, vip, cloudAccountId string) error

	// Get the current rules for a customer VIP (i.e. destination IP) that exist in the firewall. This is then compared against
	// the desired rules defined as FirewallRule CRs.
	GetExistingCustomerAccess(ctx context.Context, customerId string, vip string) ([]Rule, error)

	// RemoveAccess removes an existing rule
	RemoveAccess(ctx context.Context, rule v1alpha1.FirewallRule) (*RequestResponse, error)
}

type RequestResponse struct {
	Result string `json:"result"`
}

// EnvironmentZone defines model for Environments
type EnvironmentZone string

// Rule defines model for Rule.
type Rule struct {
	CustomerId  string          `json:"customer_id"`
	DestIp      string          `json:"dest_ip"`
	Environment EnvironmentZone `json:"environment"`
	Port        string          `json:"port"`
	Protocol    string          `json:"protocol"`
	Region      string          `json:"region"`
	SourceIp    string          `json:"source_ip"`
}

// GetCloudAccountId gets the customer cloud account ID from the firewall rule passed in.
// It first looks to see what namespace the rule is deployed into, but for other applciations,
// the rule is in the 'default' namespace, then it looks for an annoation to exist to determine
// the customer id.
func GetCloudAccountId(object v1alpha1.FirewallRule) (string, error) {

	// label key for IDC cloud account id
	const cloudAccountIdLabel = "cloud-account-id"

	// Handle the 'default' namespace edge case. IKS clusters only deploy to a default namespace.
	// All other logic in the controllers use the namespace as the cloud-account-id. Since the default
	// namespace for IKS clusters won't have the proper account-id, read the namespace from the standard
	// label on the object. Should the label be empty then return an error.
	if err := cloudaccount.CheckValidId(object.GetNamespace()); err != nil {
		cloudAcctID, found := object.Labels[cloudAccountIdLabel]
		if !found {
			return "", fmt.Errorf("could not determine cloud account id from namespace or label")
		}

		// Verify the cloud account id is valid on the label
		if err := cloudaccount.CheckValidId(cloudAcctID); err != nil {
			return "", fmt.Errorf("could not determine cloud account id, invalid label value for key \\'cloud-account-id\\'")
		}

		return cloudAcctID, nil
	}

	return object.GetNamespace(), nil
}
