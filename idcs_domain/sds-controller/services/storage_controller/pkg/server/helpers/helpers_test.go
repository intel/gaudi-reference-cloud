// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package helpers

import (
	"testing"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

func TestIntoAuthCreds(t *testing.T) {
	type args struct {
		ctx *v1.AuthenticationContext
	}
	tests := []struct {
		name string
		args args
		want *backend.AuthCreds
	}{
		{
			name: "basic context",
			args: args{
				ctx: &v1.AuthenticationContext{
					Scheme: &v1.AuthenticationContext_Basic_{
						Basic: &v1.AuthenticationContext_Basic{
							Principal:   "user",
							Credentials: "password",
						},
					},
				},
			},
			want: &backend.AuthCreds{
				Scheme:      backend.Basic,
				Principal:   "user",
				Credentials: "password",
			},
		},
		{
			name: "bearer context",
			args: args{
				ctx: &v1.AuthenticationContext{
					Scheme: &v1.AuthenticationContext_Bearer_{
						Bearer: &v1.AuthenticationContext_Bearer{
							Token: "token",
						},
					},
				},
			},
			want: &backend.AuthCreds{
				Scheme:      backend.Bearer,
				Credentials: "token",
			},
		},
		{
			name: "nil on nil",
			args: args{},
			want: nil,
		},
		{
			name: "nil on nil scheme",
			args: args{
				ctx: &v1.AuthenticationContext{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IntoAuthCreds(tt.args.ctx))
		})
	}
}

func TestValueOrNilInt(t *testing.T) {
	i := func(i int) *int {
		return &i
	}
	type args struct {
		value *int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "number", args: args{value: i(12)}, want: 12},
		{name: "negative number", args: args{value: i(-500)}, want: -500},
		{name: "zero", args: args{value: i(0)}, want: 0},
		{name: "zero on nil", args: args{}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValueOrNil(tt.args.value))
		})
	}
}

func TestValueOrNilBool(t *testing.T) {
	b := func(i bool) *bool {
		return &i
	}
	type args struct {
		value *bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "bool true", args: args{value: b(true)}, want: true},
		{name: "bool false", args: args{value: b(false)}, want: false},
		{name: "false on nil", args: args{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValueOrNil(tt.args.value))
		})
	}
}

func TestValueOrNilString(t *testing.T) {
	s := func(s string) *string {
		return &s
	}
	type args struct {
		value *string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "string exists", args: args{value: s("value")}, want: "value"},
		{name: "string on nil", args: args{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValueOrNil(tt.args.value))
		})
	}
}

func TestIsAllKeyValueExists(t *testing.T) {
	type args struct {
		m      map[string]string
		search map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "empty", args: args{m: map[string]string{}, search: map[string]string{}}, want: true},
		{name: "empty and search", args: args{m: map[string]string{}, search: map[string]string{"key": "value"}}, want: false},
		{name: "m and empty", args: args{m: map[string]string{"key": "value"}, search: map[string]string{}}, want: true},
		{name: "single value", args: args{m: map[string]string{"key": "value"}, search: map[string]string{"key": "value"}}, want: true},
		{name: "multiple m", args: args{m: map[string]string{"key": "value", "key2": "value2"}, search: map[string]string{"key": "value"}}, want: true},
		{name: "multiple match", args: args{m: map[string]string{"key": "value", "key2": "value2"}, search: map[string]string{"key": "value", "key2": "value2"}}, want: true},
		{name: "multiple mismatch", args: args{m: map[string]string{"key": "value", "key2": "value2"}, search: map[string]string{"key": "value", "key2": "value"}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsAllKeyValueExists(tt.args.m, tt.args.search))
		})
	}
}
