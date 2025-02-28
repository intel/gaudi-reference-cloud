// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package switchclients

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseTextVlansToPortVlans(t *testing.T) {
	tests := []struct {
		input       string
		expected    *showInterfacesVlans
		expectError bool
	}{
		{ // simple
			"Port       Untagged Tagged\nEt1        1      -\nEt2        123      1,456\n",
			&showInterfacesVlans{
				Interfaces: map[string]PortVlans{
					"Ethernet1": {TaggedVlans: []int(nil), UntaggedVlan: 1},
					"Ethernet2": {TaggedVlans: []int{1, 456}, UntaggedVlan: 123},
				},
			},
			false,
		},

		{ // with ranges, from real RnD switch
			"Port       Untagged Tagged\nEt1/1      1        2,4032,4066,4074\nEt2/1      20       -\nEt3/1      4033     -\nEt4/1      4033     -\nEt6/1      4033     -\nEt7/1      4033     -\nEt8/1      4033     -\nEt10/1     4033     -\nEt11/1     4033     -\nEt12/1     4033     -\nEt14/1     1        2,4032,4066,4074\nEt15/1     1        2,4032,4066,4074\nEt16/1     1        2\nEt17/1     1        -\nEt18/1     4033     -\nEt19/1     1        -\nEt22/1     4033     -\nEt23/1     4033     -\nEt24/1     4033     -\nEt25/1     4033     -\nEt32/1     1        -\nEt33       1        2,4066-4067\n",
			&showInterfacesVlans{
				Interfaces: map[string]PortVlans{
					"Ethernet1/1":  {TaggedVlans: []int{2, 4032, 4066, 4074}, UntaggedVlan: 1},
					"Ethernet10/1": {UntaggedVlan: 4033},
					"Ethernet11/1": {UntaggedVlan: 4033},
					"Ethernet12/1": {UntaggedVlan: 4033},
					"Ethernet14/1": {TaggedVlans: []int{2, 4032, 4066, 4074}, UntaggedVlan: 1},
					"Ethernet15/1": {TaggedVlans: []int{2, 4032, 4066, 4074}, UntaggedVlan: 1},
					"Ethernet16/1": {TaggedVlans: []int{2}, UntaggedVlan: 1},
					"Ethernet17/1": {UntaggedVlan: 1},
					"Ethernet18/1": {UntaggedVlan: 4033},
					"Ethernet19/1": {UntaggedVlan: 1},
					"Ethernet2/1":  {UntaggedVlan: 20},
					"Ethernet22/1": {UntaggedVlan: 4033},
					"Ethernet23/1": {UntaggedVlan: 4033},
					"Ethernet24/1": {UntaggedVlan: 4033},
					"Ethernet25/1": {UntaggedVlan: 4033},
					"Ethernet3/1":  {UntaggedVlan: 4033},
					"Ethernet32/1": {UntaggedVlan: 1},
					"Ethernet33":   {TaggedVlans: []int{2, 4066, 4067}, UntaggedVlan: 1},
					"Ethernet4/1":  {UntaggedVlan: 4033},
					"Ethernet6/1":  {UntaggedVlan: 4033},
					"Ethernet7/1":  {UntaggedVlan: 4033},
					"Ethernet8/1":  {UntaggedVlan: 4033},
				},
			},
			false,
		},

		{ // Vxlan interface
			"Port       Untagged Tagged\nVx1        None     1017-1019,4031-4033,4054-4059,4066-4067,4074\n",
			&showInterfacesVlans{
				Interfaces: map[string]PortVlans{
					"Vxlan1": {TaggedVlans: []int{1017, 1018, 1019, 4031, 4032, 4033, 4054, 4055, 4056, 4057, 4058, 4059, 4066, 4067, 4074}},
				},
			},
			false,
		},

		{ // With a wrapped line
			"Port       Untagged Tagged\nVx1        None     1017-1019,4031-4033,4054-4059,4066-4067,4074\n                    4091-4094\n",
			&showInterfacesVlans{
				Interfaces: map[string]PortVlans{
					"Vxlan1": {TaggedVlans: []int{1017, 1018, 1019, 4031, 4032, 4033, 4054, 4055, 4056, 4057, 4058, 4059, 4066, 4067, 4074, 4091, 4092, 4093, 4094}},
				},
			},
			false,
		},

		{ // With two wrapped lines
			"Port       Untagged Tagged\nVx1        None     1017-1019,2031-2033,2054-2059,2066-2067,2074\n                    3091,3092,3093,3094,3095,3096,3097,3098,3099\n                     4000,4001\n",
			&showInterfacesVlans{
				Interfaces: map[string]PortVlans{
					"Vxlan1": {TaggedVlans: []int{1017, 1018, 1019, 2031, 2032, 2033, 2054, 2055, 2056, 2057, 2058, 2059, 2066, 2067, 2074, 3091, 3092, 3093, 3094, 3095, 3096, 3097, 3098, 3099, 4000, 4001}},
				},
			},
			false,
		},
	}

	for _, test := range tests {
		result, err := parseTextVlansToPortVlans(showInterfacesVlans{
			Output: test.input,
		})
		if err != nil && test.expectError == false {
			t.Errorf("For input:\n%s\n, got unexpected error: %v", test.input, err)
		}
		if err == nil && test.expectError == true {
			t.Errorf("For input\n%s\n, expected error but got none", test.input)
		}
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("For input\n%s\n, result did not match expected: %v", test.input, cmp.Diff(test.expected, result))
		}
	}
}
