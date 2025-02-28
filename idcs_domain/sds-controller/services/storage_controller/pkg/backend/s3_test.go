// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package backend

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackend_TestIntoIAMPolicies(t *testing.T) {
	read := "read"
	write := "write"
	delete := "delete"
	bucket := "bucket"
	allow := "Allow"
	ipNetAllow := "127.0.0.0/28"
	ipNetAllow2 := "192.0.0.0/8"
	ipNetDisallow := "127.0.0.1/32"
	_, netAllow, _ := net.ParseCIDR(ipNetAllow)
	_, netAllow2, _ := net.ParseCIDR(ipNetAllow2)
	_, netDisallow, _ := net.ParseCIDR(ipNetDisallow)
	ipNetInvalid := "127.0.0.0/44"

	tests := []struct {
		name string
		arg  S3IAMPolicy
		want []*S3Policy
	}{
		{
			name: "full with prefix",
			arg: S3IAMPolicy{
				Statement: &[]S3IAMStatement{
					{
						Action:   &[]string{"action"},
						Sid:      &bucket,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket"},
						Condition: &S3IAMICondition{
							AllowSourceNets: &S3IAMIPNet{
								SourceIp: []string{ipNetAllow, ipNetAllow2},
							},
							DisallowSourceNets: &S3IAMIPNet{
								SourceIp: []string{ipNetDisallow},
							},
						},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &read,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/pre*"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &write,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/pre*"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &delete,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/pre*"},
					},
				},
			},
			want: []*S3Policy{
				{
					BucketID: "bucket",
					Read:     false,
					Write:    false,
					Delete:   false,
					Actions:  []string{"action"},
					AllowSourceNets: []*net.IPNet{
						netAllow,
						netAllow2,
					},
					DisallowSourceNets: []*net.IPNet{
						netDisallow,
					},
				},
				{
					BucketID: "bucket",
					Prefix:   "pre",
					Read:     true,
					Write:    true,
					Delete:   true,
				},
			},
		},
		{
			name: "full without prefix",
			arg: S3IAMPolicy{
				Statement: &[]S3IAMStatement{
					{
						Action:   &[]string{"action"},
						Sid:      &bucket,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &read,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/*"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &write,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/*"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &delete,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/*"},
					},
				},
			},
			want: []*S3Policy{
				{
					BucketID: "bucket",
					Read:     true,
					Write:    true,
					Delete:   true,
					Actions:  []string{"action"},
				},
			},
		},
		{
			name: "invalid cidr",
			arg: S3IAMPolicy{
				Statement: &[]S3IAMStatement{
					{
						Action:   &[]string{"action"},
						Sid:      &bucket,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket"},
						Condition: &S3IAMICondition{
							AllowSourceNets: &S3IAMIPNet{
								SourceIp: []string{ipNetInvalid},
							},
						},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &read,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/*"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &write,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/*"},
					},
					{
						Action:   &[]string{"action"},
						Sid:      &delete,
						Effect:   &allow,
						Resource: &[]string{"arn:aws:s3:::bucket/*"},
					},
				},
			},
			want: []*S3Policy{
				{
					BucketID: "bucket",
					Read:     true,
					Write:    true,
					Delete:   true,
					Actions:  []string{"action"},
				},
			},
		},
		{
			name: "empty statements",
			arg: S3IAMPolicy{
				Statement: &[]S3IAMStatement{},
			},
			want: []*S3Policy{},
		},
		{
			name: "empty",
			arg:  S3IAMPolicy{},
			want: []*S3Policy{},
		},
		{
			name: "invalid resource name",
			arg: S3IAMPolicy{
				Statement: &[]S3IAMStatement{
					{
						Action:   &[]string{},
						Sid:      &bucket,
						Effect:   &allow,
						Resource: &[]string{"something"},
					},
				},
			},
			want: []*S3Policy{},
		},
		{
			name: "empty sid",
			arg: S3IAMPolicy{
				Statement: &[]S3IAMStatement{
					{
						Action:   &[]string{},
						Effect:   &allow,
						Resource: &[]string{"something"},
					},
				},
			},
			want: []*S3Policy{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, IntoIAMPolicies(tt.arg))
		})
	}
}

func TestBackend_TestAppendIAMPolicy(t *testing.T) {
	allow := "Allow"
	bucket := "bucket"
	ipNetAllow := "127.0.0.0/28"
	ipNetDisallow := "127.0.0.1/32"
	_, netAllow, _ := net.ParseCIDR(ipNetAllow)
	_, netDisallow, _ := net.ParseCIDR(ipNetDisallow)
	type args struct {
		statementTemplates map[string]S3IAMStatement
		statementId        string
		resource           string
		actions            []string
		allowSourceNets    []*net.IPNet
		disallowSourceNets []*net.IPNet
	}
	tests := []struct {
		name string
		args args
		want map[string]S3IAMStatement
	}{
		{
			name: "add to empty",
			args: args{
				statementTemplates: make(map[string]S3IAMStatement),
				statementId:        "bucket",
				resource:           "res1",
				actions:            []string{"action1", "action2"},
				allowSourceNets:    []*net.IPNet{netAllow},
				disallowSourceNets: []*net.IPNet{netDisallow},
			},
			want: map[string]S3IAMStatement{
				"bucket": {
					Action:   &[]string{"action1", "action2"},
					Effect:   &allow,
					Resource: &[]string{"res1"},
					Sid:      &bucket,
					Condition: &S3IAMICondition{
						AllowSourceNets: &S3IAMIPNet{
							SourceIp: []string{ipNetAllow},
						},
						DisallowSourceNets: &S3IAMIPNet{
							SourceIp: []string{ipNetDisallow},
						},
					},
				},
			},
		},
		{
			name: "add to existent",
			args: args{
				statementTemplates: map[string]S3IAMStatement{
					"bucket": {
						Action:   &[]string{"action1", "action2"},
						Effect:   &allow,
						Resource: &[]string{"res1"},
						Sid:      &bucket,
					},
				},
				statementId: "bucket",
				resource:    "res2",
				actions:     []string{"action1", "action2"},
			},
			want: map[string]S3IAMStatement{
				"bucket": {
					Action:   &[]string{"action1", "action2"},
					Effect:   &allow,
					Resource: &[]string{"res1", "res2"},
					Sid:      &bucket,
				},
			},
		},
		{
			name: "ignore duplicate",
			args: args{
				statementTemplates: map[string]S3IAMStatement{
					"bucket": {
						Action:   &[]string{"action1", "action2"},
						Effect:   &allow,
						Sid:      &bucket,
						Resource: &[]string{"res1"},
					},
				},
				statementId: "bucket",
				resource:    "res1",
				actions:     []string{"action1", "action2"},
			},
			want: map[string]S3IAMStatement{
				"bucket": {
					Action:   &[]string{"action1", "action2"},
					Effect:   &allow,
					Resource: &[]string{"res1"},
					Sid:      &bucket,
				},
			},
		},
		{
			name: "ignore new actions on add",
			args: args{
				statementTemplates: map[string]S3IAMStatement{
					"bucket": {
						Action:   &[]string{"action1", "action2"},
						Effect:   &allow,
						Resource: &[]string{"res1"},
						Sid:      &bucket,
					},
				},
				statementId: "bucket",
				resource:    "res2",
				actions:     []string{"action4", "action3"},
			},
			want: map[string]S3IAMStatement{
				"bucket": {
					Action:   &[]string{"action1", "action2"},
					Effect:   &allow,
					Resource: &[]string{"res1", "res2"},
					Sid:      &bucket,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appendIAMPolicy(tt.args.statementTemplates,
				tt.args.statementId,
				tt.args.resource,
				tt.args.actions,
				tt.args.allowSourceNets,
				tt.args.disallowSourceNets)
			assert.Equal(t, tt.want, tt.args.statementTemplates)
		})
	}
}
