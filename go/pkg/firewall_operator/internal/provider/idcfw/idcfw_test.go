// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idcfw

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/provider"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ruleA = provider.Rule{
		DestIp:     "2.0.0.3",
		Port:       "80",
		Protocol:   "TCP",
		SourceIp:   "127.0.0.1",
		CustomerId: "123456789123",
	}

	ruleA_SourceIP2 = provider.Rule{
		DestIp:     "2.0.0.3",
		Port:       "80",
		Protocol:   "TCP",
		SourceIp:   "127.0.0.9",
		CustomerId: "123456789123",
	}

	ruleA_SourceIPChanged = provider.Rule{
		DestIp:     "2.0.0.3",
		Port:       "80",
		Protocol:   "TCP",
		SourceIp:   "127.0.0.2",
		CustomerId: "123456789123",
	}

	ruleB = provider.Rule{
		DestIp:     "2.0.0.3",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "127.0.0.1",
		CustomerId: "123456789123",
	}

	ruleB_SourceIPChanged = provider.Rule{
		DestIp:     "2.0.0.3",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "127.0.0.2",
		CustomerId: "123456789123",
	}

	ruleC_1 = provider.Rule{
		DestIp:     "2.0.0.4",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "23.93.211.124/32",
		CustomerId: "123456789123",
	}

	ruleC_2 = provider.Rule{
		DestIp:     "2.0.0.4",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "146.152.225.171/32",
		CustomerId: "123456789123",
	}

	ruleC_3 = provider.Rule{
		DestIp:     "2.0.0.4",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "146.152.225.0/24",
		CustomerId: "123456789123",
	}

	ruleC_SourceIP_1 = provider.Rule{
		DestIp:     "2.0.0.4",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "66.75.124.69/32",
		CustomerId: "123456789123",
	}

	ruleC_SourceIP_2 = provider.Rule{
		DestIp:     "2.0.0.4",
		Port:       "443",
		Protocol:   "TCP",
		SourceIp:   "69.181.169.121/32",
		CustomerId: "123456789123",
	}
)

func TestFlattenFWRules(t *testing.T) {
	tests := map[string]struct {
		fw             []v1alpha1.FirewallRule
		expectedResult []provider.Rule
	}{
		"simple": {
			fw: []v1alpha1.FirewallRule{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "123456789123",
				},
				Spec: v1alpha1.FirewallRuleSpec{
					SourceIPs: []string{
						"127.0.0.1",
					},
					DestinationIP: "2.0.0.3",
					Protocol:      "TCP",
					Port:          "80",
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
				Spec: v1alpha1.FirewallRuleSpec{
					SourceIPs: []string{
						"127.0.0.1",
					},
					DestinationIP: "2.0.0.3",
					Protocol:      "TCP",
					Port:          "443",
				},
			}},
			expectedResult: []provider.Rule{
				ruleA,
				ruleB,
			},
		},
		"multiple sources": {
			fw: []v1alpha1.FirewallRule{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "123456789123",
				},
				Spec: v1alpha1.FirewallRuleSpec{
					SourceIPs: []string{
						"127.0.0.1",
						"127.0.0.2",
					},
					DestinationIP: "2.0.0.3",
					Protocol:      "TCP",
					Port:          "80",
				},
			}},
			expectedResult: []provider.Rule{
				ruleA, ruleA_SourceIPChanged,
			},
		},
		"multiple sources - multiple ports": {
			fw: []v1alpha1.FirewallRule{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "123456789123",
				},
				Spec: v1alpha1.FirewallRuleSpec{
					SourceIPs: []string{
						"127.0.0.1",
						"127.0.0.2",
					},
					DestinationIP: "2.0.0.3",
					Protocol:      "TCP",
					Port:          "80",
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.FirewallRuleSpec{
					SourceIPs: []string{
						"127.0.0.1",
						"127.0.0.2",
					},
					DestinationIP: "2.0.0.3",
					Protocol:      "TCP",
					Port:          "443",
				},
			}},
			expectedResult: []provider.Rule{
				ruleA,
				ruleA_SourceIPChanged,
				ruleB,
				ruleB_SourceIPChanged,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{
				baseURL:     "",
				environment: "",
				region:      "",
				Client:      nil,
			}
			got := c.flattenFWRules(tt.fw, "123456789123")
			assert.ElementsMatch(t, tt.expectedResult, got)
		})
	}
}

func TestFlattenCurrentRules(t *testing.T) {
	tests := map[string]struct {
		response       *CurrentRulesResult
		expectedResult []provider.Rule
	}{
		"simple": {
			response: &CurrentRulesResult{
				CurrentRules: []RuleResult{{
					CustomerId:  "123456789123",
					Region:      "Dev",
					Environment: "dev",
					RuleName:    "Flex_IDCAPI_Prod_Inbound_112233445566_2",
					SourceAddress: []string{
						"134.134.137.84",
					},
					DestAddress: []string{
						"7.8.9.0",
					},
					Port: []string{"TCP_6789"},
				}, {
					CustomerId:  "123456789123",
					Region:      "Dev",
					Environment: "dev",
					RuleName:    "Flex_IDCAPI_Prod_Inbound_112233445566_1",
					SourceAddress: []string{
						"134.134.137.83",
					},
					DestAddress: []string{
						"7.8.9.0",
					},
					Port: []string{"TCP_6789"},
				}},
			},
			expectedResult: []provider.Rule{
				{
					CustomerId:  "123456789123",
					DestIp:      "7.8.9.0",
					Environment: "dev",
					Port:        "6789",
					Protocol:    "TCP",
					Region:      "us-dev-1",
					SourceIp:    "134.134.137.84",
				},
				{
					CustomerId:  "123456789123",
					DestIp:      "7.8.9.0",
					Environment: "dev",
					Port:        "6789",
					Protocol:    "TCP",
					Region:      "us-dev-1",
					SourceIp:    "134.134.137.83",
				},
			},
		},
		"multiple ports same source/dest": {
			response: &CurrentRulesResult{
				CurrentRules: []RuleResult{{
					CustomerId:  "123456789123",
					Region:      "us-dev-1",
					Environment: "dev",
					RuleName:    "Flex_IDCAPI_Prod_Inbound_112233445566_2",
					SourceAddress: []string{
						"134.134.137.84",
					},
					DestAddress: []string{
						"7.8.9.0",
					},
					Port: []string{"TCP_6789"},
				}, {
					CustomerId:  "123456789123",
					Region:      "us-dev-1",
					Environment: "dev",
					RuleName:    "Flex_IDCAPI_Prod_Inbound_112233445566_1",
					SourceAddress: []string{
						"134.134.137.84",
					},
					DestAddress: []string{
						"7.8.9.0",
					},
					Port: []string{"TCP_6790"},
				}},
			},
			expectedResult: []provider.Rule{{
				CustomerId:  "123456789123",
				DestIp:      "7.8.9.0",
				Environment: "dev",
				Port:        "6789",
				Protocol:    "TCP",
				Region:      "us-dev-1",
				SourceIp:    "134.134.137.84",
			}, {
				CustomerId:  "123456789123",
				DestIp:      "7.8.9.0",
				Environment: "dev",
				Port:        "6790",
				Protocol:    "TCP",
				Region:      "us-dev-1",
				SourceIp:    "134.134.137.84",
			}},
		},
		"sourceIP with /32": {
			response: &CurrentRulesResult{
				CurrentRules: []RuleResult{{
					CustomerId: "123456789123",
					Region:     "Dev",
					RuleName:   "Flex_IDCAPI_Prod_Inbound_112233445566_2",
					SourceAddress: []string{
						"134.134.137.84",
						"134.134.137.85/32",
					},
					DestAddress: []string{
						"7.8.9.0",
					},
					Port: []string{"TCP_6789"},
				}, {
					CustomerId: "123456789123",
					Region:     "Dev",
					RuleName:   "Flex_IDCAPI_Prod_Inbound_112233445566_1",
					SourceAddress: []string{
						"134.134.137.83",
					},
					DestAddress: []string{
						"7.8.9.0",
					},
					Port: []string{"TCP_6789"},
				}},
			},
			expectedResult: []provider.Rule{{
				CustomerId:  "123456789123",
				DestIp:      "7.8.9.0",
				Environment: "dev",
				Port:        "6789",
				Protocol:    "TCP",
				Region:      "us-dev-1",
				SourceIp:    "134.134.137.84",
			}, {
				CustomerId:  "123456789123",
				DestIp:      "7.8.9.0",
				Environment: "dev",
				Port:        "6789",
				Protocol:    "TCP",
				Region:      "us-dev-1",
				SourceIp:    "134.134.137.83",
			}, {
				CustomerId:  "123456789123",
				DestIp:      "7.8.9.0",
				Environment: "dev",
				Port:        "6789",
				Protocol:    "TCP",
				Region:      "us-dev-1",
				SourceIp:    "134.134.137.85",
			}},
		},
	}
	for name, tt := range tests {

		c := Client{
			baseURL:     "",
			environment: "dev",
			region:      "us-dev-1",
			Client:      nil,
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			got := c.flattenCurrentRules(ctx, tt.response, "7.8.9.0")
			assert.ElementsMatch(t, tt.expectedResult, got, "FlattenCurrentRules(%v)", tt.response)
		})
	}
}

func TestCalculateRulesToAddRemove(t *testing.T) {

	tests := map[string]struct {
		desired              []provider.Rule
		existing             []provider.Rule
		expectedRulesAdded   map[RuleRequest]SourceIPs
		expectedRulesRemoved map[RuleRequest]SourceIPs
	}{
		"no change": {
			desired:              []provider.Rule{ruleA, ruleB},
			existing:             []provider.Rule{ruleA, ruleB},
			expectedRulesAdded:   make(map[RuleRequest]SourceIPs),
			expectedRulesRemoved: make(map[RuleRequest]SourceIPs),
		},
		"empty set, 2 rules added": {
			desired: []provider.Rule{
				ruleA,
				ruleA_SourceIP2,
			},
			existing: []provider.Rule{},
			expectedRulesAdded: map[RuleRequest]SourceIPs{
				ruleRequestFromRule(ruleA): []string{ruleA.SourceIp, ruleA_SourceIP2.SourceIp},
			},
			expectedRulesRemoved: map[RuleRequest]SourceIPs{},
		},
		"add sourceIP to existing rule": {
			desired: []provider.Rule{
				ruleA,
				ruleA_SourceIP2,
			},
			existing: []provider.Rule{ruleA},
			expectedRulesAdded: map[RuleRequest]SourceIPs{
				ruleRequestFromRule(ruleA): []string{ruleA_SourceIP2.SourceIp},
			},
			expectedRulesRemoved: map[RuleRequest]SourceIPs{},
		},
		"add sourceIP to existing rule - large": {
			desired: []provider.Rule{
				ruleC_1, ruleC_2, ruleC_3,
				ruleC_SourceIP_1, ruleC_SourceIP_2,
			},
			existing: []provider.Rule{ruleC_1, ruleC_2, ruleC_3},
			expectedRulesAdded: map[RuleRequest]SourceIPs{
				ruleRequestFromRule(ruleC_1): []string{ruleC_SourceIP_1.SourceIp, ruleC_SourceIP_2.SourceIp},
			},
			expectedRulesRemoved: map[RuleRequest]SourceIPs{},
		},
		"change port": {
			desired: []provider.Rule{
				ruleB,
			},
			existing: []provider.Rule{ruleA},
			expectedRulesAdded: map[RuleRequest]SourceIPs{
				ruleRequestFromRule(ruleB): []string{ruleB.SourceIp},
			},
			expectedRulesRemoved: map[RuleRequest]SourceIPs{
				ruleRequestFromRule(ruleA): []string{ruleA.SourceIp},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			gotAdd, gotRemove := calculateRulesToAddRemove(ctx, tt.desired, tt.existing, "123456789123")

			assert.Equal(t, tt.expectedRulesAdded, gotAdd, "added")
			assert.Equal(t, tt.expectedRulesRemoved, gotRemove, "removed")
		})
	}
}

func ruleRequestFromRule(rule provider.Rule) RuleRequest {
	return RuleRequest{
		DestIp:     rule.DestIp,
		Port:       rule.Port,
		Protocol:   rule.Protocol,
		CustomerId: "123456789123",
	}
}
