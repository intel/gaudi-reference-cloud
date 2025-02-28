// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gorilla/mux"
	devicesmanager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	"gopkg.in/yaml.v3"

	// "crypto/tls"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	mineralRiver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/mineral-river"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	//+kubebuilder:scaffold:imports
	swClient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// TODO: Move out to handlers.go or types.go.
type SaveConfigPayload struct {
	SwitchFQDN string `json:"switch_fqdn"`
}

type AccessPortPayload struct {
	SwitchFQDN  string `json:"switch_fqdn"`
	SwitchPort  string `json:"switch_port"`
	Description string `json:"description"`
	VlanTag     int    `json:"vlan_tag"`
}

type UnmaintainedPortPayload struct {
	SwitchFQDN string `json:"switch_fqdn"`
	SwitchPort string `json:"switch_port"`
}

type AccessPortChannelPayload struct {
	SwitchFQDN  string `json:"switch_fqdn"`
	PortChannel int    `json:"port_channel"`
	Description string `json:"description"`
	VlanTag     int    `json:"vlan_tag"`
}

type UnmaintainedPortChannelPayload struct {
	SwitchFQDN  string `json:"switch_fqdn"`
	PortChannel int    `json:"port_channel"`
}

type TrunkPortPayload struct {
	SwitchFQDN  string  `json:"switch_fqdn"`
	SwitchPort  string  `json:"switch_port"`
	Description string  `json:"description"`
	NativeVlan  int     `json:"native_vlan"`
	TrunkGroup  *string `json:"trunk_group"`
}

type TrunkPortChannelPayload struct {
	SwitchFQDN  string  `json:"switch_fqdn"`
	PortChannel int     `json:"port_channel"`
	Description string  `json:"description"`
	NativeVlan  int     `json:"native_vlan"`
	TrunkGroup  *string `json:"trunk_group"`
}

type CreatePortChannelPayload struct {
	SwitchFQDN  string `json:"switch_fqdn"`
	PortChannel int    `json:"port_channel"`
}
type DeletePortChannelPayload struct {
	SwitchFQDN  string `json:"switch_fqdn"`
	PortChannel int    `json:"port_channel"`
}

type SetSwitchportToPortChannel struct {
	SwitchFQDN  string `json:"switch_fqdn"`
	SwitchPort  string `json:"switch_port"`
	PortChannel int    `json:"port_channel"`
	Description string `json:"description"`
}

func respondWithJsonError(ctx context.Context, w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	err2 := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	if err2 != nil {
		log.FromContext(ctx).Error(err2, "Got another error while encoding error to json")
	}
}

func respondWithJsonSuccessMinimal(ctx context.Context, w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(map[string]interface{}{"status": true})
	if err != nil {
		log.FromContext(ctx).Error(err, "Got an error while encoding response to json")
	}
}

func respondWithJsonSuccessSimple(ctx context.Context, w http.ResponseWriter, responseKey string, response interface{}) {
	err := json.NewEncoder(w).Encode(map[string]interface{}{responseKey: response, "status": true})
	if err != nil {
		log.FromContext(ctx).Error(err, "Got an error while encoding response to json")
	}
}

func main() {
	var metricsAddr string
	var probeAddr string
	var configFile string
	var eapiSecretPath string
	var nok8s bool
	flag.StringVar(&configFile, "configFile", "", "The application will load its configuration from this file.")
	flag.StringVar(&eapiSecretPath, "eapiSecretPath", "", "Eapi secret file path")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&nok8s, "insecure-no-k8s", false, "If set to true, will allow management of any switch, not just those in k8s. Setting this to true disables all k8s operations.")
	log.BindFlags()
	flag.Parse()

	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx)

	var err error

	// read configuration from config file
	cfg := &idcnetworkv1alpha1.SDNControllerRestConfig{}
	if configFile == "" {
		logger.Error(fmt.Errorf("unable to read configuration"), "configuration file is not provided")
		os.Exit(1)
	}
	cfgBytes, err := os.ReadFile(configFile)
	if err != nil {
		logger.Error(fmt.Errorf("failed to read configFile %s err: %v", configFile, err), "failed to read configFile")
		os.Exit(1)
	}
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		logger.Error(fmt.Errorf("failed to unmarshal configFile %s err: %v", configFile, err), "failed to unmarshal configFile")
	}
	logger.Info("configuration", "cfg", cfg)
	err = validateConfig(cfg)
	if err != nil {
		logger.Error(err, "unable to validate the config file")
		os.Exit(1)
	}

	logger.V(1).Info("Debug logging enabled")

	var allowedVlanIds []int
	var allowedNativeVlanIds []int
	allowedVlanIds, err = utils.ExpandVlanRanges(cfg.RestConfig.AllowedVlanIds)
	if err != nil {
		logger.Error(err, "Error expanding valid VLAN range")
	}

	allowedNativeVlanIds, err = utils.ExpandVlanRanges(cfg.RestConfig.AllowedNativeVlanIds)
	if err != nil {
		logger.Error(err, "Error expanding valid Native VLAN ids")
	}

	// Initialize monitoring
	mr := mineralRiver.New(
		mineralRiver.WithLogLevel("Debug"),
	)
	tracerProvider := mr.InitTracer(context.Background())
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Error(err, "Error shutting down tracer provider")
		}
	}()

	var devicemngr devicesmanager.DevicesAccessManager
	var k8sClient client.Client
	var deviceMgrControllerCfg = devicesmanager.DeviceManagerControllerConfig{
		SwitchBackendMode:    "eapi", // REST-API only supports eapi backend.
		SwitchSecretsPath:    eapiSecretPath,
		AllowedTrunkGroups:   cfg.RestConfig.AllowedTrunkGroups,
		AllowedVlanIds:       allowedVlanIds,
		AllowedNativeVlanIds: allowedNativeVlanIds,
		Datacenter:           cfg.RestConfig.DataCenter,
	}
	var deviceMgrCfg = devicesmanager.DeviceManagerConfig{
		ControllerConfig: deviceMgrControllerCfg,
	}

	if nok8s {
		devicemngr = devicesmanager.NewDeviceAccessManager(nil, deviceMgrCfg)
		devicemngr.InsecurelyDisableFQDNValidation()
	} else {
		k8sClient = utils.NewK8SClientWithScheme(scheme)
		devicemngr = devicesmanager.NewDeviceAccessManager(k8sClient, deviceMgrCfg)
	}

	router := mux.NewRouter()

	// TODO: Move these to routes.go or handlers.go or similar
	// Get Mac Address-Table
	router.HandleFunc("/devcloud/v4/list/mac_address_table", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadOnlyOU(ctx, w, r) {
			return
		}

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			// If there's an error parsing, abort. The alternative, r.URL.Query() ignores errors and sets values to "".
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		req := swClient.ListMacAddressTableRequest{SwitchFQDN: fqdn}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get mac address-table
		entries, err := conn.GetMacAddressTable(ctx, req)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "mac_address_table", entries)
		}
	}).Methods("GET")

	// Get LLDP Neighbors
	router.HandleFunc("/devcloud/v4/list/lldp_neighbors", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadOnlyOU(ctx, w, r) {
			return
		}

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		portName := queryParams.Get("switch_port")
		if portName == "" {
			portName = "none"
		} else {
			err = utils.ValidatePortNumber(portName)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		req := swClient.PortParamsRequest{SwitchFQDN: fqdn, SwitchPort: portName}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get lldp neighbors
		if portName == "none" {
			// Get all ports
			entries, err := conn.GetLLDPNeighbors(ctx, req)
			if err != nil {
				if strings.Contains(err.Error(), "BadRequest") {
					// Return client error
					respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				} else {
					// Return server error
					respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
				}
			} else {
				respondWithJsonSuccessSimple(ctx, w, "lldp_neighbors", entries)
			}
		} else {
			// Get a single port (return value is different than get all ports above)
			entries, err := conn.GetLLDPPortNeighbors(ctx, req)
			if err != nil {
				if strings.Contains(err.Error(), "BadRequest") {
					// Return client error
					respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				} else {
					// Return server error
					respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
				}
			} else {
				respondWithJsonSuccessSimple(ctx, w, "lldp_neighbors", entries)
			}
		}

	}).Methods("GET")

	// Save Config
	router.HandleFunc("/devcloud/v4/save/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload SaveConfigPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		fqdn := payload.SwitchFQDN
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to save configuration
		_, err = conn.SaveConfig(ctx, fqdn)
		if err != nil {
			logger.Error(err, "Error saving config")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "save_config", "success")
		}
	}).Methods("POST")

	// Get Running Port Config
	router.HandleFunc("/devcloud/v4/port/running_config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadOnlyOU(ctx, w, r) {
			return
		}

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		portName := queryParams.Get("switch_port")
		if portName != "" {
			err = utils.ValidatePortNumber(portName)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		portChannelStr := queryParams.Get("port_channel")
		var portChannel = 0
		if portChannelStr != "" {
			portChannel, err = strconv.Atoi(portChannelStr)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
			err = utils.ValidatePortChannelNumber(portChannel)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		if portName == "" && portChannel == 0 {
			respondWithJsonError(ctx, w, fmt.Errorf("must specify switchport or portchannel"), http.StatusBadRequest)
			return
		}

		req := swClient.PortParamsRequest{
			SwitchFQDN:  fqdn,
			SwitchPort:  portName,
			PortChannel: portChannel,
		}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get running config
		entries, err := conn.GetPortRunningConfig(ctx, req)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "interface_config", entries)
		}
	}).Methods("GET")

	// Get Port Details
	router.HandleFunc("/devcloud/v4/port/details", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadOnlyOU(ctx, w, r) {
			return
		}

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		portNum := queryParams.Get("switch_port")
		portChannelNumStr := queryParams.Get("port_channel")
		req := swClient.PortParamsRequest{SwitchFQDN: fqdn}
		if portNum != "" {
			err = utils.ValidatePortNumber(portNum)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
			req.SwitchPort = portNum
		} else if portChannelNumStr != "" {
			portChannelNum, err := strconv.Atoi(portChannelNumStr)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
			err = utils.ValidatePortChannelNumber(portChannelNum)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
			req.PortChannel = portChannelNum
		} else {
			respondWithJsonError(ctx, w, fmt.Errorf("switch_port or port_channel must be specified"), http.StatusBadRequest)
			return
		}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get single port details
		entries, err := conn.GetPortDetails(ctx, req)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "port_information", entries)
		}
	}).Methods("GET")

	// Get ALL ports from switch (like /port/details, but all ports)
	router.HandleFunc("/devcloud/v4/list/ports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadOnlyOU(ctx, w, r) {
			return
		}

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get port details
		entries, err := conn.ListPortsDetails(ctx, swClient.ListPortParamsRequest{SwitchFQDN: fqdn})
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "port_list", entries)
		}

	}).Methods("GET")

	router.HandleFunc("/devcloud/v4/list/vlans", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadOnlyOU(ctx, w, r) {
			return
		}

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get port details
		entries, err := conn.ListVlans(ctx, swClient.ListVlansParamsRequest{SwitchFQDN: fqdn})
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "vlans", entries)
		}

	}).Methods("GET")

	// Get IP Mac Info
	router.HandleFunc("/devcloud/v4/list/ip_mac_info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		fqdn := queryParams.Get("switch_fqdn")
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		req := swClient.ParamsRequest{SwitchFQDN: fqdn}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to get ip mac info
		entries, err := conn.GetIpMacInfo(ctx, req)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "ip_mac_info", entries)
		}
	}).Methods("GET")

	// Clear MAC Address Table
	router.HandleFunc("/devcloud/v4/clear/mac_address_table", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload SaveConfigPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		fqdn := payload.SwitchFQDN
		err = utils.ValidateSwitchFQDN(fqdn, cfg.RestConfig.DataCenter)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		// Make switch connection
		conn, err := devicemngr.GetOrCreateSwitchClient(ctx, devicesmanager.GetOption{SwitchFQDN: fqdn})
		if err != nil {
			logger.Error(err, "unable to get SwitchClient")
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// Call to clear mac address-table
		_, err = conn.ClearMacAddressTable(ctx, fqdn)
		if err != nil {
			logger.Error(err, "Error clear mac address-table")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
		} else {
			respondWithJsonSuccessSimple(ctx, w, "clear_mac_table", "success")
		}
	}).Methods("POST")

	router.HandleFunc("/devcloud/v4/configure/port/access", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload AccessPortPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		var sanitizedDescription string
		if payload.Description != "" {
			sanitizedDescription, err = utils.ValidateAndSanitizeDescription(payload.Description)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		err = utils.ValidatePortNumber(payload.SwitchPort)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		switchportName := utils.GeneratePortFullName(payload.SwitchFQDN, "Ethernet"+payload.SwitchPort)

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingSwitchPort := &idcnetworkv1alpha1.SwitchPort{}
		key := types.NamespacedName{Name: switchportName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingSwitchPort)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		if existingSwitchPort.Spec.Mode != idcnetworkv1alpha1.AccessMode && payload.VlanTag == 0 {
			respondWithJsonError(ctx, w, fmt.Errorf("port not in the access mode, vlan_tag must be provided in the payload"), http.StatusBadRequest)
			return
		}
		if payload.VlanTag != 0 {
			err = utils.ValidateVlanValue(int(payload.VlanTag), allowedVlanIds)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		// patch update the switch CR
		newSwitchPortCR := existingSwitchPort.DeepCopy()
		newSwitchPortCR.Spec.Mode = idcnetworkv1alpha1.AccessMode
		if payload.VlanTag != 0 {
			newSwitchPortCR.Spec.VlanId = int64(payload.VlanTag)
		}
		newSwitchPortCR.Spec.TrunkGroups = &[]string{}
		newSwitchPortCR.Spec.NativeVlan = 1
		newSwitchPortCR.Spec.PortChannel = 0
		if payload.Description != "" {
			newSwitchPortCR.Spec.Description = sanitizedDescription
		}
		patch := client.MergeFrom(existingSwitchPort)
		if err := k8sClient.Patch(ctx, newSwitchPortCR, patch); err != nil {
			logger.Error(err, "update SwitchPort CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the switchport status has been updated, ie. the change has actually happened on the server (For backwards compatibility with Raven)
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, key, existingSwitchPort)
				if err != nil {
					resCh <- &err
				}

				modeInSync := existingSwitchPort.Status.Mode == existingSwitchPort.Spec.Mode && existingSwitchPort.Spec.Mode == idcnetworkv1alpha1.AccessMode
				vlanTagInSync := payload.VlanTag == 0 || (payload.VlanTag != 0 && existingSwitchPort.Status.VlanId == existingSwitchPort.Spec.VlanId && existingSwitchPort.Status.VlanId == int64(payload.VlanTag))
				descriptionInSync := payload.Description == "" || (payload.Description != "" && existingSwitchPort.Status.Description == existingSwitchPort.Spec.Description && existingSwitchPort.Spec.Description == sanitizedDescription)
				portChannelInSync := existingSwitchPort.Status.PortChannel == existingSwitchPort.Spec.PortChannel
				// spec == status == desired
				if modeInSync && descriptionInSync && vlanTagInSync && portChannelInSync {
					// Complete.
					resCh <- nil
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for update to be applied, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for update to be applied")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)

	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/configure/port/unmaintained", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload UnmaintainedPortPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		err = utils.ValidatePortNumber(payload.SwitchPort)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		switchportName := utils.GeneratePortFullName(payload.SwitchFQDN, "Ethernet"+payload.SwitchPort)

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingSwitchPort := &idcnetworkv1alpha1.SwitchPort{}
		key := types.NamespacedName{Name: switchportName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingSwitchPort)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// patch update the switch CR
		newSwitchPortCR := existingSwitchPort.DeepCopy()
		newSwitchPortCR.Spec.Mode = ""
		newSwitchPortCR.Spec.VlanId = -1
		newSwitchPortCR.Spec.TrunkGroups = nil
		newSwitchPortCR.Spec.NativeVlan = -1
		newSwitchPortCR.Spec.PortChannel = -1
		newSwitchPortCR.Spec.Description = ""
		patch := client.MergeFrom(existingSwitchPort)
		if err := k8sClient.Patch(ctx, newSwitchPortCR, patch); err != nil {
			logger.Error(err, "update SwitchPort CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)

	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/configure/portchannel/access", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !cfg.RestConfig.PortChannelsEnabled {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel support is disabled in config"), http.StatusBadRequest)
			return
		}

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload AccessPortChannelPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		var sanitizedDescription string
		if payload.Description != "" {
			sanitizedDescription, err = utils.ValidateAndSanitizeDescription(payload.Description)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		err = utils.ValidatePortChannelNumber(payload.PortChannel)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		portChannelName, err := utils.PortChannelNumberAndSwitchFQDNToCRName(payload.PortChannel, payload.SwitchFQDN)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingPortChannel := &idcnetworkv1alpha1.PortChannel{}
		key := types.NamespacedName{Name: portChannelName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingPortChannel)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		if existingPortChannel.Spec.Mode != idcnetworkv1alpha1.AccessMode && payload.VlanTag == 0 {
			respondWithJsonError(ctx, w, fmt.Errorf("portChannel not in the access mode, vlan_tag must be provided in the payload"), http.StatusBadRequest)
			return
		}

		if payload.VlanTag != 0 {
			err = utils.ValidateVlanValue(int(payload.VlanTag), allowedVlanIds)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		// patch update the switch CR
		newPortChannelCR := existingPortChannel.DeepCopy()
		newPortChannelCR.Spec.Mode = idcnetworkv1alpha1.AccessMode
		if payload.VlanTag != 0 {
			newPortChannelCR.Spec.VlanId = int64(payload.VlanTag)
		}
		// explicitly REMOVE the trunk settings from the switch
		newPortChannelCR.Spec.TrunkGroups = &[]string{}
		newPortChannelCR.Spec.NativeVlan = 1
		if payload.Description != "" {
			newPortChannelCR.Spec.Description = sanitizedDescription
		}
		patch := client.MergeFrom(existingPortChannel)
		if err := k8sClient.Patch(ctx, newPortChannelCR, patch); err != nil {
			logger.Error(err, "patch update PortChannel CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the portChannel status has been updated, ie. the change has actually happened on the server (For backwards compatibility with Raven)
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, key, existingPortChannel)
				if err != nil {
					resCh <- &err
				}
				modeInSync := existingPortChannel.Status.Mode == existingPortChannel.Spec.Mode && existingPortChannel.Spec.Mode == idcnetworkv1alpha1.AccessMode
				vlanTagInSync := payload.VlanTag == 0 || (payload.VlanTag != 0 && existingPortChannel.Status.VlanId == existingPortChannel.Spec.VlanId && existingPortChannel.Status.VlanId == int64(payload.VlanTag))
				descriptionInSync := payload.Description == "" || (payload.Description != "" && existingPortChannel.Status.Description == existingPortChannel.Spec.Description && existingPortChannel.Spec.Description == sanitizedDescription)
				// spec == status == desired
				if modeInSync && descriptionInSync && vlanTagInSync {
					// Complete.
					resCh <- nil
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for update to be applied, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for update to be applied")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)

	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/configure/portchannel/unmaintained", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !cfg.RestConfig.PortChannelsEnabled {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel support is disabled in config"), http.StatusBadRequest)
			return
		}

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload UnmaintainedPortChannelPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidatePortChannelNumber(payload.PortChannel)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		portChannelName, err := utils.PortChannelNumberAndSwitchFQDNToCRName(payload.PortChannel, payload.SwitchFQDN)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingPortChannel := &idcnetworkv1alpha1.PortChannel{}
		key := types.NamespacedName{Name: portChannelName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingPortChannel)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// patch update the switch CR
		newPortChannelCR := existingPortChannel.DeepCopy()
		newPortChannelCR.Spec.Mode = ""
		newPortChannelCR.Spec.VlanId = -1
		newPortChannelCR.Spec.TrunkGroups = nil
		newPortChannelCR.Spec.NativeVlan = -1
		newPortChannelCR.Spec.Description = ""
		patch := client.MergeFrom(existingPortChannel)
		if err := k8sClient.Patch(ctx, newPortChannelCR, patch); err != nil {
			logger.Error(err, "patch update PortChannel CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)

	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/configure/port/trunk", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload TrunkPortPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		err = utils.ValidatePortNumber(payload.SwitchPort)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		var trunkGroups []string
		if payload.TrunkGroup != nil {
			trunkGroups, err = utils.SplitAndValidateTrunkGroupsString(*payload.TrunkGroup, cfg.RestConfig.AllowedTrunkGroups)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		var sanitizedDescription string
		if payload.Description != "" {
			sanitizedDescription, err = utils.ValidateAndSanitizeDescription(payload.Description)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		switchportName := utils.GeneratePortFullName(payload.SwitchFQDN, "Ethernet"+payload.SwitchPort)

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingSwitchPort := &idcnetworkv1alpha1.SwitchPort{}
		key := types.NamespacedName{Name: switchportName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingSwitchPort)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		if existingSwitchPort.Spec.Mode != idcnetworkv1alpha1.TrunkMode && payload.NativeVlan == 0 {
			respondWithJsonError(ctx, w, fmt.Errorf("port not in the trunk mode, native_vlan must be provided in the payload"), http.StatusBadRequest)
			return
		}

		if payload.NativeVlan != 0 {
			err = utils.ValidateVlanValue(int(payload.NativeVlan), allowedNativeVlanIds)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		// patch update the switchport CR
		newSwitchPortCR := existingSwitchPort.DeepCopy()

		newSwitchPortCR.Spec.Mode = idcnetworkv1alpha1.TrunkMode
		if payload.NativeVlan != 0 {
			newSwitchPortCR.Spec.NativeVlan = int64(payload.NativeVlan)
		}
		newSwitchPortCR.Spec.PortChannel = 0
		if payload.Description != "" {
			newSwitchPortCR.Spec.Description = sanitizedDescription
		}
		if trunkGroups != nil {
			newSwitchPortCR.Spec.TrunkGroups = &trunkGroups
		}
		newSwitchPortCR.Spec.VlanId = 1

		patch := client.MergeFrom(existingSwitchPort)
		if err := k8sClient.Patch(ctx, newSwitchPortCR, patch); err != nil {
			logger.Error(err, "patch update SwitchPort CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the switchport status has been updated, ie. the change has actually happened on the server (For backwards compatibility with Raven)
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, key, existingSwitchPort)
				if err != nil {
					resCh <- &err
				}

				trunkGroupsInSync := true
				if trunkGroups != nil {

					var specTrunkGroups []string
					if existingSwitchPort.Spec.TrunkGroups != nil {
						specTrunkGroups = *existingSwitchPort.Spec.TrunkGroups
					}
					sort.Strings(specTrunkGroups)
					sort.Strings(existingSwitchPort.Status.TrunkGroups)
					sort.Strings(trunkGroups)

					trunkGroupsInSync = slices.Equal(existingSwitchPort.Status.TrunkGroups, specTrunkGroups) && slices.Equal(existingSwitchPort.Status.TrunkGroups, trunkGroups)
				}

				modeInSync := existingSwitchPort.Status.Mode == existingSwitchPort.Spec.Mode && existingSwitchPort.Spec.Mode == idcnetworkv1alpha1.TrunkMode
				nativeVlanInSync := payload.NativeVlan == 0 || (payload.NativeVlan != 0 && existingSwitchPort.Status.NativeVlan == existingSwitchPort.Spec.NativeVlan && existingSwitchPort.Status.NativeVlan == int64(payload.NativeVlan))
				descriptionInSync := payload.Description == "" || (payload.Description != "" && existingSwitchPort.Status.Description == existingSwitchPort.Spec.Description && existingSwitchPort.Spec.Description == sanitizedDescription)
				portChannelInSync := existingSwitchPort.Status.PortChannel == existingSwitchPort.Spec.PortChannel
				// spec == status == desired
				if modeInSync && trunkGroupsInSync && descriptionInSync && nativeVlanInSync && portChannelInSync {
					// Complete.
					resCh <- nil
					return
				}
				//fmt.Printf("Still waiting... %v %v %v %v %v %v",
				//	existingSwitchPort.Status.Mode == existingSwitchPort.Spec.Mode,
				//	existingSwitchPort.Spec.Mode == idcnetworkv1alpha1.TrunkMode,
				//	existingSwitchPort.Status.NativeVlan == existingSwitchPort.Spec.NativeVlan,
				//	existingSwitchPort.Spec.NativeVlan == int64(payload.NativeVlan),
				//	slices.Equal(existingSwitchPort.Status.TrunkGroups, existingSwitchPort.Spec.TrunkGroups),
				//	slices.Equal(existingSwitchPort.Status.TrunkGroups, trunkGroups))
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for update to be applied, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for update to be applied")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)

	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/configure/portchannel/trunk", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !cfg.RestConfig.PortChannelsEnabled {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel support is disabled in config"), http.StatusBadRequest)
		}

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload TrunkPortChannelPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		err = utils.ValidatePortChannelNumber(payload.PortChannel)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		var trunkGroups []string
		if payload.TrunkGroup != nil {
			trunkGroups, err = utils.SplitAndValidateTrunkGroupsString(*payload.TrunkGroup, cfg.RestConfig.AllowedTrunkGroups)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		var sanitizedDescription string
		if payload.Description != "" {
			sanitizedDescription, err = utils.ValidateAndSanitizeDescription(payload.Description)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		portChannelCRName, err := utils.PortChannelNumberAndSwitchFQDNToCRName(payload.PortChannel, payload.SwitchFQDN)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingPortChannel := &idcnetworkv1alpha1.PortChannel{}
		key := types.NamespacedName{Name: portChannelCRName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingPortChannel)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		if existingPortChannel.Spec.Mode != idcnetworkv1alpha1.TrunkMode && payload.NativeVlan == 0 {
			respondWithJsonError(ctx, w, fmt.Errorf("portchannel not in the trunk mode, native_vlan must be provided in the payload"), http.StatusBadRequest)
			return
		}
		if payload.NativeVlan != 0 {
			err = utils.ValidateVlanValue(int(payload.NativeVlan), allowedNativeVlanIds)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}
		// patch update the PortChannel CR
		newPortChannelCR := existingPortChannel.DeepCopy()
		newPortChannelCR.Spec.Mode = idcnetworkv1alpha1.TrunkMode
		if payload.NativeVlan != 0 {
			newPortChannelCR.Spec.NativeVlan = int64(payload.NativeVlan)
		}
		if payload.Description != "" {
			newPortChannelCR.Spec.Description = sanitizedDescription
		}
		if trunkGroups != nil {
			newPortChannelCR.Spec.TrunkGroups = &trunkGroups
		}
		newPortChannelCR.Spec.VlanId = 1

		patch := client.MergeFrom(existingPortChannel)
		if err := k8sClient.Patch(ctx, newPortChannelCR, patch); err != nil {
			logger.Error(err, "patch update PortChannel CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the portChannel status has been updated, ie. the change has actually happened on the server (For backwards compatibility with Raven)
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, key, existingPortChannel)
				if err != nil {
					resCh <- &err
				}

				trunkGroupsInSync := true
				if trunkGroups != nil {
					var specTrunkGroups []string
					if existingPortChannel.Spec.TrunkGroups != nil {
						specTrunkGroups = *existingPortChannel.Spec.TrunkGroups
					}
					sort.Strings(specTrunkGroups)
					sort.Strings(existingPortChannel.Status.TrunkGroups)
					sort.Strings(trunkGroups)

					trunkGroupsInSync = slices.Equal(existingPortChannel.Status.TrunkGroups, specTrunkGroups) && slices.Equal(existingPortChannel.Status.TrunkGroups, trunkGroups)
				}

				modeInSync := existingPortChannel.Status.Mode == existingPortChannel.Spec.Mode && existingPortChannel.Spec.Mode == idcnetworkv1alpha1.TrunkMode
				nativeVlanInSync := payload.NativeVlan == 0 || (payload.NativeVlan != 0 && existingPortChannel.Status.NativeVlan == existingPortChannel.Spec.NativeVlan && existingPortChannel.Status.NativeVlan == int64(payload.NativeVlan))
				descriptionInSync := payload.Description == "" || (payload.Description != "" && existingPortChannel.Status.Description == existingPortChannel.Spec.Description && existingPortChannel.Spec.Description == sanitizedDescription)
				// spec == status == desired
				if modeInSync && trunkGroupsInSync && descriptionInSync && nativeVlanInSync {
					// Complete.
					resCh <- nil
					return
				}
				//fmt.Printf("Still waiting... %v %v %v %v %v %v %v %v \n",
				//	existingPortChannel.Status.Mode == existingPortChannel.Spec.Mode,
				//	existingPortChannel.Spec.Mode == idcnetworkv1alpha1.TrunkMode,
				//	existingPortChannel.Status.NativeVlan == existingPortChannel.Spec.NativeVlan,
				//	existingPortChannel.Spec.NativeVlan == int64(payload.NativeVlan),
				//	slices.Equal(existingPortChannel.Status.TrunkGroups, existingPortChannel.Spec.TrunkGroups),
				//	slices.Equal(existingPortChannel.Status.TrunkGroups, trunkGroups),
				//	existingPortChannel.Status.TrunkGroups,
				//	trunkGroups,
				//)
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for update to be applied, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for update to be applied")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)
	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/configure/port/portchannel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !cfg.RestConfig.PortChannelsEnabled {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel support is disabled in config"), http.StatusBadRequest)
			return
		}

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload SetSwitchportToPortChannel
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		err = utils.ValidatePortNumber(payload.SwitchPort)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		switchportName := utils.GeneratePortFullName(payload.SwitchFQDN, "Ethernet"+payload.SwitchPort)

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		existingSwitchPort := &idcnetworkv1alpha1.SwitchPort{}
		switchportKey := types.NamespacedName{Name: switchportName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, switchportKey, existingSwitchPort)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		if (existingSwitchPort.Spec.PortChannel == 0 || existingSwitchPort.Spec.PortChannel == -1) && payload.PortChannel == 0 {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel must be provided in the payload"), http.StatusBadRequest)
			return
		}

		if payload.PortChannel != 0 {
			err = utils.ValidatePortChannelNumber(payload.PortChannel)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		var sanitizedDescription string
		if payload.Description != "" {
			sanitizedDescription, err = utils.ValidateAndSanitizeDescription(payload.Description)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		var portChannelCRName string
		if payload.PortChannel != 0 {
			portChannelCRName, err = utils.PortChannelNumberAndSwitchFQDNToCRName(payload.PortChannel, payload.SwitchFQDN)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		} else {
			portChannelCRName, err = utils.PortChannelNumberAndSwitchFQDNToCRName(int(existingSwitchPort.Spec.PortChannel), payload.SwitchFQDN)
			if err != nil {
				respondWithJsonError(ctx, w, err, http.StatusBadRequest)
				return
			}
		}

		// Check if PortChannel already exists
		existingPortChannel := &idcnetworkv1alpha1.PortChannel{}
		key := types.NamespacedName{Name: portChannelCRName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingPortChannel)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, err, http.StatusNotFound)
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			}
			return
		}

		// patch update the SwitchPort CR
		patchedSwitchportCR := existingSwitchPort.DeepCopy()

		if payload.PortChannel != 0 {
			patchedSwitchportCR.Spec.PortChannel = int64(payload.PortChannel)
		}

		if payload.Description != "" {
			patchedSwitchportCR.Spec.Description = sanitizedDescription
		}
		// Explicitly remove access & trunk settings from the switchport because it is in a portchannel and these will be ignored by the switch once it's put into the portchannel.
		patchedSwitchportCR.Spec.Mode = "" // TODO ???
		patchedSwitchportCR.Spec.NativeVlan = 1
		patchedSwitchportCR.Spec.TrunkGroups = &[]string{}
		patchedSwitchportCR.Spec.VlanId = 1

		patch := client.MergeFrom(existingSwitchPort)
		if err := k8sClient.Patch(ctx, patchedSwitchportCR, patch); err != nil {
			logger.Error(err, "patch update SwitchPort CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the portChannel status has been updated, ie. the change has actually happened on the server (For backwards compatibility with Raven)
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, switchportKey, existingSwitchPort)
				if err != nil {
					resCh <- &err
				}

				// spec == status == desired
				portChannelInSync := payload.PortChannel == 0 || (payload.PortChannel != 0 && existingSwitchPort.Status.PortChannel == existingSwitchPort.Spec.PortChannel && existingSwitchPort.Status.PortChannel == int64(payload.PortChannel))
				descriptionInSync := payload.Description == "" || (payload.Description != "" && existingSwitchPort.Status.Description == existingSwitchPort.Spec.Description && existingSwitchPort.Spec.Description == sanitizedDescription)
				if portChannelInSync && descriptionInSync {
					// Complete.
					resCh <- nil
					return
				}

				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for update to be applied, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for update to be applied")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)
	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/create/portchannel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !cfg.RestConfig.PortChannelsEnabled {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel support is disabled in config"), http.StatusBadRequest)
			return
		}

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload CreatePortChannelPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		err = utils.ValidatePortChannelNumber(payload.PortChannel)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		portChannelCRName, err := utils.PortChannelNumberAndSwitchFQDNToCRName(payload.PortChannel, payload.SwitchFQDN)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		portChannelInterfaceName, err := utils.PortChannelNumberToInterfaceName(payload.PortChannel)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		// Check if PortChannel already exists
		existingPortChannelCR := &idcnetworkv1alpha1.PortChannel{}
		key := types.NamespacedName{Name: portChannelCRName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingPortChannelCR)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Expected behaviour
			} else {
				// There was another sort of error
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
				return
			}
		}
		if existingPortChannelCR.Name != "" {
			w.WriteHeader(http.StatusOK) // 200 because there is no "409 conflict" (because we don't pass any additional parameters in this call).
			respondWithJsonSuccessSimple(ctx, w, "message", "PortChannel already exists")
			return
		}

		// Create new PortChannel
		newPortChannelCR := &idcnetworkv1alpha1.PortChannel{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					idcnetworkv1alpha1.LabelNameSwitchFQDN: payload.SwitchFQDN,
				},
				Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
				Name:      portChannelCRName,
			},
			Spec: idcnetworkv1alpha1.PortChannelSpec{
				Name: portChannelInterfaceName,
				Mode: "access", // Must specify a mode so that it is "SDN controlled". Otherwise, creating and then deleting a PortChannel via the REST API won't actually delete it from the switch.
			},
		}

		if err := k8sClient.Create(ctx, newPortChannelCR); err != nil {
			logger.Error(err, "create PortChannel CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the switchport status has been updated, ie. the change has actually happened on the switch (For backwards compatibility with Raven)
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, key, newPortChannelCR)
				if err != nil {
					resCh <- &err
				}

				if !newPortChannelCR.Status.LastStatusChangeTime.IsZero() {
					// Complete.
					resCh <- nil
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for portchannel to be created, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for portchannel to be created")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		respondWithJsonSuccessMinimal(ctx, w)
	}).Methods("PUT")

	router.HandleFunc("/devcloud/v4/delete/portchannel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !cfg.RestConfig.PortChannelsEnabled {
			respondWithJsonError(ctx, w, fmt.Errorf("PortChannel support is disabled in config"), http.StatusBadRequest)
			return
		}

		if !requireReadWriteOU(ctx, w, r) {
			return
		}

		rawBody := json.NewDecoder(r.Body)
		var payload DeletePortChannelPayload
		err := rawBody.Decode(&payload)
		if err != nil {
			// Return client error
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		err = utils.ValidateSwitchFQDN(payload.SwitchFQDN, "")
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}
		err = utils.ValidatePortChannelNumber(payload.PortChannel)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		portChannelCRName, err := utils.PortChannelNumberAndSwitchFQDNToCRName(payload.PortChannel, payload.SwitchFQDN)
		if err != nil {
			respondWithJsonError(ctx, w, err, http.StatusBadRequest)
			return
		}

		if k8sClient == nil {
			respondWithJsonError(ctx, w, fmt.Errorf("this endpoint requires kubernetes access"), http.StatusInternalServerError)
			return
		}

		// Check if PortChannel already exists
		existingPortChannelCR := &idcnetworkv1alpha1.PortChannel{}
		key := types.NamespacedName{Name: portChannelCRName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = k8sClient.Get(ctx, key, existingPortChannelCR)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondWithJsonError(ctx, w, fmt.Errorf("PortChannel not found"), http.StatusBadRequest)
				return
			} else {
				respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
				return
			}
		}

		// Update the portchannel mode so that deletion of the CR really removes it from the switch (SDN ignores deletions of PortChannels that have all fields empty to prevent removing portchannels LEARNED from the switch)
		existingPortChannelCR.Spec.Mode = "delete"
		if err := k8sClient.Update(ctx, existingPortChannelCR); err != nil {
			logger.Error(err, "update PortChannel CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Delete PortChannel
		if err := k8sClient.Delete(ctx, existingPortChannelCR); err != nil {
			logger.Error(err, "delete PortChannel CR failed")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		// Wait until the switchport CR has really been deleted.
		resCh := make(chan *error)
		go func() {
			startWaitTime := time.Now()
			for true {
				// Prevent call from continuing to check, even once JSON error has been returned to the user in case it's never going to succeed (eg. it was already changed to something else).
				if time.Since(startWaitTime) > 30*time.Second {
					return
				}

				err = k8sClient.Get(ctx, key, existingPortChannelCR)
				if err != nil {
					if apierrors.IsNotFound(err) {
						// Complete.
						err = nil
						resCh <- nil
						return
					} else {
						respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
						return
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			err = fmt.Errorf("timeout waiting for portchannel to be deleted, desired change has NOT been rolled-back")
		case err2, ok := <-resCh:
			if !ok {
				respondWithJsonError(ctx, w, fmt.Errorf("result channel is closed"), http.StatusInternalServerError)
			}
			if err2 != nil {
				err = *err2
			}
		}
		if err != nil {
			logger.Error(err, "error while waiting for portchannel to be deleted")
			respondWithJsonError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		respondWithJsonSuccessMinimal(ctx, w)
	}).Methods("DELETE")

	healthRouter := mux.NewRouter()
	healthRouter.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		respondWithJsonSuccessMinimal(ctx, w)
	})
	healthRouter.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		respondWithJsonSuccessMinimal(ctx, w)
	})

	port := strconv.FormatUint(uint64(cfg.RestConfig.ListenPort), 10)
	healthPort := strconv.FormatUint(uint64(cfg.RestConfig.HealthPort), 10)
	addr := "0.0.0.0:" + port
	healthAddr := "0.0.0.0:" + healthPort

	// Defaults are secure. Can be overridden by ENV vars.
	tlsConfig, err := tlsutil.NewTlsProvider().ServerTlsConfig(ctx)
	if err != nil {
		logger.Error(err, "Error while generating Server TLS config")
	}

	// Disallow TLS 1.0 and 1.1
	tlsConfig.MinVersion = tls.VersionTLS12

	// As per CT-35: https://readthedocs.intel.com/cryptoteam/crypto_bkms/tls.html#id3
	tlsConfig.CipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}

	// liveness / readiness probes don't support clientCerts, so start a server with almost identical config, but
	// without mTLS for health endpoints only.
	// It won't check the health of the real "srv", but will at least fail if the go process exits.
	healthTlsConfig := tlsConfig.Clone()
	healthTlsConfig.ClientAuth = tls.NoClientCert

	srv := &http.Server{
		Handler:      LoggerMiddleware(router),
		Addr:         addr,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
		TLSConfig:    tlsConfig,
	}

	healthSrv := &http.Server{
		Handler:      healthRouter,
		Addr:         healthAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		TLSConfig:    healthTlsConfig,
	}

	go func() {
		logger.Info(fmt.Sprintf("Health srv listening on %s", healthAddr))
		err = healthSrv.ListenAndServeTLS("", "") // empty values here aren't used - overridden by tlsConfig above.
		if err != nil {
			logger.Error(err, "Error while running SDN-Controller REST Health server")
			os.Exit(1)
		}
	}()

	logger.Info(fmt.Sprintf("Listening on %s", addr))
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		logger.Error(err, "Error while running SDN-Controller REST server")
		os.Exit(1)
	}
	<-ctx.Done()

	logger.Error(err, "terminating server")
}

func requireReadOnlyOU(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	if r.TLS != nil {
		if b, err := strconv.ParseBool(os.Getenv("IDC_REQUIRE_CLIENT_CERTIFICATE")); err == nil && !b {
			// INSECURELY OVERRIDE & ALLOW - Do not use in production.
			return true
		}

		certificates := r.TLS.PeerCertificates
		if len(certificates) > 0 {
			cert := certificates[0]

			var hasReadOnlyPermission = false
			for _, ou := range cert.Subject.OrganizationalUnit {
				if ou == "provider-sdn-readwrite" || ou == "provider-sdn-readonly" {
					hasReadOnlyPermission = true
				}
			}
			if !hasReadOnlyPermission {
				respondWithJsonError(ctx, w, fmt.Errorf("provider-sdn-readonly permission required to call this endpoint"), http.StatusUnauthorized)
				return false
			} else {
				return true
			}
		} else {
			respondWithJsonError(ctx, w, fmt.Errorf("no client-certificate found in TLS request"), http.StatusUnauthorized)
			return false
		}
	}
	respondWithJsonError(ctx, w, fmt.Errorf("TLS is not enabled"), http.StatusUnauthorized)
	return false
}

func requireReadWriteOU(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	if r.TLS != nil {
		if b, err := strconv.ParseBool(os.Getenv("IDC_REQUIRE_CLIENT_CERTIFICATE")); err == nil && !b {
			// INSECURELY OVERRIDE & ALLOW - Do not use in production.
			return true
		}

		certificates := r.TLS.PeerCertificates
		if len(certificates) > 0 {
			cert := certificates[0]

			var hasReadWritePermission = false
			for _, ou := range cert.Subject.OrganizationalUnit {
				if ou == "provider-sdn-readwrite" {
					hasReadWritePermission = true
				}
			}
			if !hasReadWritePermission {
				respondWithJsonError(ctx, w, fmt.Errorf("provider-sdn-readwrite permission required to call this endpoint"), http.StatusUnauthorized)
				return false
			} else {
				return true
			}
		} else {
			respondWithJsonError(ctx, w, fmt.Errorf("no client-certificate found in TLS request"), http.StatusUnauthorized)
			return false
		}
	}
	respondWithJsonError(ctx, w, fmt.Errorf("TLS is not enabled"), http.StatusUnauthorized)
	return false
}

func validateConfig(cfg *idcnetworkv1alpha1.SDNControllerRestConfig) error {
	var errMsgs []string
	if cfg.RestConfig.ListenPort == 0 {
		errMsgs = append(errMsgs, "ListenPort is not provided")
	}
	if cfg.RestConfig.DataCenter == "" {
		errMsgs = append(errMsgs, "DataCenter is not provided")
	}

	if len(errMsgs) == 0 {
		return nil
	}
	errMsg := "config file validation failed"
	for _, msg := range errMsgs {
		errMsg += ", " + msg
	}
	return fmt.Errorf(errMsg)
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	if rw.status == 0 {
		return http.StatusOK // Default
	}
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now().UTC()

			// Read body and copy because it can't be read twice.
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.FromContext(r.Context()).WithName("LoggerMiddleWare").Error(err, "failed to read body")
			}
			err = r.Body.Close()
			if err != nil {
				log.FromContext(r.Context()).WithName("LoggerMiddleWare").Error(err, "failed to close body")
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(r.Context()).WithName(fmt.Sprintf("%v %v", r.Method, r.URL.Path)).WithValues("method", r.Method, "requestURI", r.RequestURI, "body", string(bodyBytes)).Start()
			logger.Info(fmt.Sprintf("Serving request %v %v", r.Method, r.RequestURI))

			// Wrap the writer so we can intercept the status for logging and pass the context down so logs inside will be associated with this span.
			wrappedWriter := wrapResponseWriter(w)
			wrappedReader := r.WithContext(ctx)

			next.ServeHTTP(wrappedWriter, wrappedReader)

			logger.Info(fmt.Sprintf("Finished serving request %v with status code %v in %v", r.RequestURI, wrappedWriter.Status(), time.Since(startTime)),
				"status", wrappedWriter.Status(),
				"duration", time.Since(startTime),
			)

			span.End()
		},
	)
}
