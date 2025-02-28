package testutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareConfigs(t *testing.T) {
	testCases := []struct {
		name     string
		config1  string
		config2  string
		expected string
	}{
		{
			name: "No Differences",
			config1: `
!
interface Ethernet5
   switchport access vlan 101
   switchport
!
interface Ethernet6
   switchport access vlan 101
   switchport
!
interface Loopback0
   ip address A.B.C.D/32`,
			config2: `
!
interface Ethernet5
   switchport access vlan 101
   switchport
!
interface Ethernet6
   switchport access vlan 101
   switchport
!
interface Loopback0
   ip address A.B.C.D/32`,
			expected: "",
		},
		{
			name: "Single Line Difference",
			config1: `
!
interface Ethernet5
   switchport access vlan 101
   switchport
!
interface Ethernet6
   switchport access vlan 101
   switchport
!
interface Loopback0
   ip address A.B.C.D/32`,
			config2: `
!
interface Ethernet5
   switchport access vlan 101
   switchport
!
interface Ethernet6
   switchport access vlan 102
   switchport
!
interface Loopback0
   ip address A.B.C.D/32`,
			expected: `interface Ethernet6
   switchport access vlan 10
- 1
+ 2
   switchport
!`,
		},
		{
			name: "Multiple Line Difference",
			config1: `
!
interface Ethernet5
  switchport access vlan 101
  switchport
!
interface Ethernet6
  switchport access vlan 101
  switchport
!
interface Loopback0
  ip address A.B.C.D/32`,
			config2: `
!
interface Ethernet5
  switchport access vlan 103
  switchport
!
interface Ethernet6
  switchport access vlan 102
  switchport
!
interface Loopback0
  ip address A.B.C.D/32`,
			expected: `interface Ethernet5
  switchport access vlan 10
- 1
+ 3
  switchport
!
interface Ethernet6
  switchport access vlan 10
- 1
+ 2
  switchport
!`,
		},
		{
			name: "Multiple Line Difference with skipped content",
			config1: `
!
interface Ethernet5
  switchport access vlan 101
  switchport
!
interface Ethernet1
  switchport access vlan 100
  switchport
!
interface Ethernet2
  switchport access vlan 100
  switchport
!
interface Ethernet6
  switchport access vlan 101
  switchport
!
interface Loopback0
  ip address A.B.C.D/32`,
			config2: `
!
interface Ethernet5
  switchport access vlan 103
  switchport
!
interface Ethernet1
  switchport access vlan 100
  switchport
!
interface Ethernet2
  switchport access vlan 100
  switchport
!
interface Ethernet6
  switchport access vlan 102
  switchport
!
interface Loopback0
  ip address A.B.C.D/32`,
			expected: `interface Ethernet5
  switchport access vlan 10
- 1
+ 3
  switchport
!
---
interface Ethernet6
  switchport access vlan 10
- 1
+ 2
  switchport
!`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareConfigs(tc.config1, tc.config2)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDiff(t *testing.T) {
	testCases := []struct {
		name     string
		config1  string
		config2  string
		expected string
	}{
		{
			name: "csae 1",
			config1: `
!
interface Ethernet5
   switchport access vlan 100
   switchport
!
interface Ethernet6
   switchport access vlan 100
   switchport
!
interface Loopback0
   ip address A.B.C.D/32
`,
			config2: `
!
interface Ethernet5
   switchport access vlan 100
   switchport
!
interface Ethernet6
   switchport access vlan 101
   switchport
!
interface Loopback0
   ip address A.B.C.D/32
`,
			expected: `@@ @@
 !
 interface Ethernet6
-   switchport access vlan 100
+   switchport access vlan 101
    switchport
 !
`,
		},
		{
			name: "Single Line Difference",
			config1: `
!
interface Ethernet5
   switchport access vlan 101
   switchport
!
interface Ethernet6
   switchport access vlan 101
   switchport
!
interface Loopback0
   ip address A.B.C.D/32
`,
			config2: `
!
interface Ethernet5
   switchport access vlan 101
   switchport
!
interface Ethernet6
   switchport access vlan 102
   switchport
!
interface Loopback0
   ip address A.B.C.D/32
`,
			expected: `@@ @@
 !
 interface Ethernet6
-   switchport access vlan 101
+   switchport access vlan 102
    switchport
 !
`,
		},
		{
			name: "Multiple Line Difference",
			config1: `
!
interface Ethernet5
  switchport access vlan 101
  switchport
!
interface Ethernet6
  switchport access vlan 101
  switchport
!
interface Loopback0
  ip address A.B.C.D/32
`,
			config2: `
!
interface Ethernet5
  switchport access vlan 103
  switchport
!
interface Ethernet6
  switchport access vlan 102
  switchport
!
interface Loopback0
  ip address A.B.C.D/32
`,
			expected: `@@ @@
 !
 interface Ethernet5
-  switchport access vlan 101
+  switchport access vlan 103
   switchport
 !
 interface Ethernet6
-  switchport access vlan 101
+  switchport access vlan 102
   switchport
 !
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Diff(tc.config1, tc.config2, 2)
			if err != nil {
				fmt.Println(err)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
