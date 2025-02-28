// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

// import (
// 	"context"
// 	"fmt"
// 	"reflect"
// 	"testing"

// 	"sigs.k8s.io/controller-runtime/pkg/client"
// )

// TODO: it take a local kubeocnfig for the test, comment this out for now, and will fix it later.
// func TestSDNClient_UpdateVlan(t *testing.T) {
// 	type fields struct {
// 		// dynamicClient       dynamic.Interface
// 		k8sClient           client.Client
// 		watchTimeoutSeconds int
// 	}
// 	type args struct {
// 		ctx         context.Context
// 		switch_fqdn string
// 		port        string
// 		vlan        int64
// 		description string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr error
// 	}{
// 		{
// 			name:   "case1",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        0,
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf("ValidateVlanValue failed, tenant port must be set to vlan between 100 and 3999, or provisioning vlan 4008. Use Raven to set provider ports. Requested: %d", 0),
// 		},
// 		{
// 			name:   "case2",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        -1,
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf("ValidateVlanValue failed, tenant port must be set to vlan between 100 and 3999, or provisioning vlan 4008. Use Raven to set provider ports. Requested: %d", -1),
// 		},
// 		{
// 			name:   "case3",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        0x100,
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf(`k8sClient.Get SwitchPort failed, switchports.idcnetwork.intel.com "ethernet27-1.internal-placeholder.com" not found`), // this error in this case is OK, as we provide a nil k8s client.
// 		},
// 		{
// 			name:   "case4",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        0x7fffffff, // 2147483647
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf("ValidateVlanValue failed, tenant port must be set to vlan between 100 and 3999, or provisioning vlan 4008. Use Raven to set provider ports. Requested: %d", 2147483647),
// 		},
// 		{
// 			name:   "case5",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        0x80000000, // 2147483648
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf("ValidateVlanValue failed, tenant port must be set to vlan between 100 and 3999, or provisioning vlan 4008. Use Raven to set provider ports. Requested: %d", -2147483648),
// 		},
// 		{
// 			name:   "case6",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        99,
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf("ValidateVlanValue failed, tenant port must be set to vlan between 100 and 3999, or provisioning vlan 4008. Use Raven to set provider ports. Requested: %v", 99),
// 		},
// 		{
// 			name:   "case7",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        4000,
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf("ValidateVlanValue failed, tenant port must be set to vlan between 100 and 3999, or provisioning vlan 4008. Use Raven to set provider ports. Requested: %v", 4000),
// 		},
// 		{
// 			name:   "case8",
// 			fields: fields{},
// 			args: args{
// 				ctx:         context.Background(),
// 				port:        "Ethernet27/1",
// 				vlan:        4008,
// 				switch_fqdn: "internal-placeholder.com",
// 			},
// 			wantErr: fmt.Errorf(`k8sClient.Get SwitchPort failed, switchports.idcnetwork.intel.com "ethernet27-1.internal-placeholder.com" not found`), // this error in this case is OK, as we provide a nil k8s client.
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &SDNClient{
// 				// k8sClient:           tt.fields.k8sClient,
// 				watchTimeoutSeconds: tt.fields.watchTimeoutSeconds,
// 			}
// 			err := c.UpdateVlan(tt.args.ctx, tt.args.switch_fqdn, tt.args.port, tt.args.vlan, tt.args.description)
// 			if tt.wantErr == nil && err != nil {
// 				t.Errorf("test %v failed, error is not expected: %v", tt.name, err)
// 			}
// 			if tt.wantErr != nil && !reflect.DeepEqual(tt.wantErr, err) {
// 				t.Errorf("test %v failed, expect error [%v], but got [%v]", tt.name, tt.wantErr, err)
// 			}

// 		})
// 	}
// }
