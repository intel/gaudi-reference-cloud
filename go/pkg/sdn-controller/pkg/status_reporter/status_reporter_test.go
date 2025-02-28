package statusreporter

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateNotControlledPortChannelsMap(t *testing.T) {

	var allInterfacesFromSwitch = make(map[string]*v1alpha1.SwitchPortStatus)

	// Spine link
	allInterfacesFromSwitch["Ethernet27"] = &v1alpha1.SwitchPortStatus{
		Name:        "Ethernet27",
		PortChannel: 271,
	}

	// SDN-Controlled link in port channel
	allInterfacesFromSwitch["Ethernet5"] = &v1alpha1.SwitchPortStatus{
		Name:        "Ethernet5",
		PortChannel: 51,
	}

	// SDN-Controlled link NOT in port channel
	allInterfacesFromSwitch["Ethernet6"] = &v1alpha1.SwitchPortStatus{
		Name: "Ethernet6",
	}

	allInterfacesFromSwitch["Port-Channel271"] = &v1alpha1.SwitchPortStatus{ // Port-Channel to spine (should NOT be controlled)
		Name: "Port-Channel271",
	}
	allInterfacesFromSwitch["Port-Channel51"] = &v1alpha1.SwitchPortStatus{ // should be controlled because it contains a controlled switchport
		Name: "Port-Channel51",
	}
	allInterfacesFromSwitch["Port-Channel81"] = &v1alpha1.SwitchPortStatus{ // Port-channel that is empty
		Name: "Port-Channel81",
	}
	allInterfacesFromSwitch["Port-Channel61"] = &v1alpha1.SwitchPortStatus{ // Also empty, but shares a name with a switch port.
		Name: "Port-Channel61",
	}

	controlledSwitchPortsFromK8s := &v1alpha1.SwitchPortList{
		Items: []v1alpha1.SwitchPort{
			{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec: v1alpha1.SwitchPortSpec{
					Name:        "Ethernet5",
					Mode:        "access",
					VlanId:      4008,
					NativeVlan:  1,
					Description: "server1-1_8",
					TrunkGroups: nil,
					PortChannel: 51,
				},
			},
		},
	}

	result := CreateNotControlledPortChannelsMap(allInterfacesFromSwitch, controlledSwitchPortsFromK8s)

	// Should have added 271 to the map because we do NOT control Ethernet27.
	member, shouldNotBeControlled := result[271]
	if shouldNotBeControlled != true {
		t.Errorf("Port-Channel271 should not be controlled, but it is.")
	}
	if member != "Ethernet27" {
		t.Errorf("Port-Channel271 should have member Ethernet27.")
	}

	// Should not have added 51 to the map because we do control Ethernet5.
	_, shouldNotBeControlled = result[51]
	if shouldNotBeControlled == true {
		t.Errorf("Port-Channel5 should be controlled because we control its switchport member, but it is not.")
	}

	// Should not have added 81 to the map because it is empty (no members)
	_, shouldNotBeControlled = result[81]
	if shouldNotBeControlled == true {
		t.Errorf("Port-Channel81 should be controlled because it is empty, but it is not.")
	}

	// Should not have added 61 to the map because it is empty (no members)
	_, shouldNotBeControlled = result[61]
	if shouldNotBeControlled == true {
		t.Errorf("Port-Channel61 should be controlled because it is empty, but it is not.")
	}

}
