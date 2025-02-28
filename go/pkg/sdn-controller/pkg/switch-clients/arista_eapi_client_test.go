// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package switchclients

import (
	"fmt"
	"os"
	"testing"

	"github.com/aristanetworks/goeapi"
	"github.com/aristanetworks/goeapi/module"
)

/*
   "Ethernet27/1": {
     "Bandwidth": 1215752192,
     "InterfaceType": "Not Present",
     "Description": "Tenant_Trunk",
     "AutoNegotiateActive": false,
     "Duplex": "duplexFull",
     "LinkStatus": "notconnect",
     "LineProtocolStatus": "notPresent",
     "VlanInformation": {
       "InterfaceMode": "trunk",
       "VlanID": 0,
       "InterfaceForwardingModel": "bridged"
     }
   },
*/

func TestGetSwitch(t *testing.T) {
	err := os.Setenv("EAPI_CONF", "eapi.conf")
	if err != nil {
		fmt.Printf("error setting eAPI env var")
	}
	node, err := goeapi.ConnectTo("zal0112a")
	if err != nil {
		fmt.Printf("error connecting to switch: %v", err)
	}
	sys := module.System(node)

	_ = sys
}
