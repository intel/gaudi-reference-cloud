// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"testing"

	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getInstanceType(t *testing.T) {
	type args struct {
		bmh *baremetalv1alpha1.BareMetalHost
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				bmh: createBMHwithLabel("instance-type.cloud.intel.com/bm-virtual"),
			},
			want:    "bm-virtual",
			wantErr: false,
		},
		{
			name: "invalidLabel",
			args: args{
				bmh: createBMHwithLabel("instance-type.cloud.intel.com/"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "missingLabel",
			args: args{
				bmh: createBMHwithLabel(""),
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInstanceType(tt.args.bmh)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInstanceType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getInstanceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// helper methods
func createBMHwithLabel(key string) *baremetalv1alpha1.BareMetalHost {
	return &baremetalv1alpha1.BareMetalHost{
		ObjectMeta: v1.ObjectMeta{
			Name: "testBMH",
			Labels: map[string]string{
				key: "true",
			},
		},
	}
}
