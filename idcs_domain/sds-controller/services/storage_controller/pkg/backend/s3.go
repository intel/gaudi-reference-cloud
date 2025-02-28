// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package backend

import (
	"fmt"
	"net"
	"regexp"
	"slices"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
)

type S3IAMPolicy struct {
	Statement *[]S3IAMStatement `json:"Statement,omitempty"`
	Version   *string           `json:"Version,omitempty"`
}

type S3IAMStatement struct {
	Action    *[]string        `json:"Action,omitempty"`
	Effect    *string          `json:"Effect,omitempty"`
	Resource  *[]string        `json:"Resource,omitempty"`
	Sid       *string          `json:"Sid,omitempty"`
	Condition *S3IAMICondition `json:"Condition,omitempty"`
}

type S3IAMICondition struct {
	AllowSourceNets    *S3IAMIPNet `json:"IpAddress,omitempty"`
	DisallowSourceNets *S3IAMIPNet `json:"NotIpAddress,omitempty"`
}

type S3IAMIPNet struct {
	SourceIp []string `json:"aws:SourceIp,omitempty"`
}

const S3_POLICY_VERSION = "2012-10-17"

var s3Re = regexp.MustCompile(`arn:aws:s3:::([^\/]*)\/?([^\*]*)`)

func IntoIAMPolicies(content S3IAMPolicy) []*S3Policy {
	policiesMap := make(map[string]*S3Policy, 0)

	if content.Statement == nil {
		return make([]*S3Policy, 0)
	}

	for _, statement := range *content.Statement {
		if statement.Resource == nil || statement.Sid == nil || statement.Effect == nil {
			log.Warn().Any("policy", statement).Msg("Unable to parse policy statement")
			continue
		}
		var allowSourceNets []*net.IPNet
		var disallowSourceNets []*net.IPNet
		var errors []error

		if statement.Condition != nil {
			if statement.Condition.AllowSourceNets != nil {
				allowSourceNets, errors = ParseNets(statement.Condition.AllowSourceNets.SourceIp)
				log.Error().Bool("security", true).Errs("parseErrors", errors).Msg("Cannot parse generated CIDR for allow nets")
			}

			if statement.Condition.DisallowSourceNets != nil {
				disallowSourceNets, errors = ParseNets(statement.Condition.DisallowSourceNets.SourceIp)
				log.Error().Bool("security", true).Errs("parseErrors", errors).Msg("Cannot parse generated CIDR for disallow nets")
			}
		}
		for _, resource := range *statement.Resource {
			matches := s3Re.FindStringSubmatch(resource)
			if len(matches) != 3 {
				log.Warn().Str("resourceName", resource).Msg("Invalid resource name")
				continue
			}
			var policy *S3Policy
			key := fmt.Sprintf("%s%s", matches[1], matches[2])
			policy, exists := policiesMap[key]
			if !exists {
				policy = &S3Policy{
					BucketID:           matches[1],
					Prefix:             matches[2],
					AllowSourceNets:    allowSourceNets,
					DisallowSourceNets: disallowSourceNets,
				}
				policiesMap[key] = policy
			}

			if *statement.Sid == "read" {
				policy.Read = true
			} else if *statement.Sid == "write" {
				policy.Write = true
			} else if *statement.Sid == "delete" {
				policy.Delete = true
			} else if *statement.Sid == "bucket" && statement.Action != nil {
				policy.Actions = *statement.Action
			}
		}
	}

	return maps.Values(policiesMap)
}

func FromIAMPolicy(policies []*S3Policy) S3IAMPolicy {
	statementTemplates := make(map[string]S3IAMStatement)
	for _, policy := range policies {
		appendIAMPolicy(statementTemplates,
			"bucket",
			fmt.Sprintf("arn:aws:s3:::%s", policy.BucketID),
			policy.Actions,
			policy.AllowSourceNets,
			policy.DisallowSourceNets,
		)
		objectResource := fmt.Sprintf("arn:aws:s3:::%s/%s*", policy.BucketID, policy.Prefix)
		if policy.Read {
			appendIAMPolicy(statementTemplates,
				"read",
				objectResource,
				[]string{
					"s3:GetObject",
				},
				policy.AllowSourceNets,
				policy.DisallowSourceNets,
			)
		}
		if policy.Write {
			appendIAMPolicy(statementTemplates,
				"write",
				objectResource,
				[]string{
					"s3:AbortMultipartUpload",
					"s3:PutObject",
				},
				policy.AllowSourceNets,
				policy.DisallowSourceNets,
			)

		}
		if policy.Delete {
			appendIAMPolicy(statementTemplates,
				"delete",
				objectResource,
				[]string{
					"s3:DeleteObject",
				},
				policy.AllowSourceNets,
				policy.DisallowSourceNets,
			)
		}
	}

	version := S3_POLICY_VERSION
	statements := maps.Values(statementTemplates)
	return S3IAMPolicy{
		Version:   &version,
		Statement: &statements,
	}
}

func ParseNets(nets []string) ([]*net.IPNet, []error) {
	var cidrs []*net.IPNet
	var errors []error
	for _, ipNet := range nets {
		_, cidr, err := net.ParseCIDR(ipNet)
		if err != nil {
			errors = append(errors, err)
		} else if cidr != nil {
			cidrs = append(cidrs, cidr)
		}
	}

	return cidrs, errors
}

func appendIAMPolicy(statementTemplates map[string]S3IAMStatement,
	statementId string,
	resource string,
	actions []string,
	allowSourceNets []*net.IPNet,
	disallowSourceNets []*net.IPNet) {

	allow := "Allow"
	var statement S3IAMStatement
	statement, exists := statementTemplates[statementId]
	if !exists {
		var condition *S3IAMICondition
		if (len(allowSourceNets) + len(disallowSourceNets)) > 0 {
			condition = &S3IAMICondition{}
			if len(allowSourceNets) > 0 {
				condition.AllowSourceNets = formatCIDRs(allowSourceNets)
			}
			if len(disallowSourceNets) > 0 {
				condition.DisallowSourceNets = formatCIDRs(disallowSourceNets)
			}
		}
		statement = S3IAMStatement{
			Effect:    &allow,
			Action:    &actions,
			Resource:  &[]string{},
			Sid:       &statementId,
			Condition: condition,
		}
	}
	if !slices.Contains(*statement.Resource, resource) {
		*statement.Resource = append(*statement.Resource, resource)
	}
	statementTemplates[statementId] = statement
}

func formatCIDRs(nets []*net.IPNet) *S3IAMIPNet {
	var cidrs []string
	for _, net := range nets {
		cidr := net.String()
		cidrs = append(cidrs, cidr)
	}
	return &S3IAMIPNet{
		SourceIp: cidrs,
	}
}
