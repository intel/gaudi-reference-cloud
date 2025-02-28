package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))
}

// TODO: combine the 32 and 8-node-per-group tool into one.
func main() {
	var action string
	var numGroup int
	var bmhKubeconfig string
	var nwcpKubeconfig string
	var resources string
	var fabrics string
	var idle bool
	flag.StringVar(&action, "action", "add", "add/delete")
	flag.BoolVar(&idle, "idle", true, "")
	flag.StringVar(&resources, "r", "bmh:sw", "select which resources to create/delete")
	flag.IntVar(&numGroup, "n", 32, "number of groups")
	flag.StringVar(&bmhKubeconfig, "bmhKubeconfig", "", "bmh kubeconfig file path")
	flag.StringVar(&nwcpKubeconfig, "nwcpKubeconfig", "", "nwcp kubeconfig file path")
	flag.StringVar(&fabrics, "fabrics", "FAS", "F: frontend, A: accelerator, S: storage")
	flag.Parse()

	ctx := context.Background()

	var handlebmh bool
	var handlesw bool
	for _, str := range strings.Split(resources, ":") {
		if str == "bmh" {
			handlebmh = true
		} else if str == "sw" {
			handlesw = true
		}
	}

	var bmhClusterClient rtclient.Client
	if handlebmh {
		if len(bmhKubeconfig) > 0 {
			fmt.Printf("Connecting to bmh k8s... \n")
			bmhClusterClient = utils.NewK8SClientFromConfAndScheme(ctx, bmhKubeconfig, scheme)
		} else {
			fmt.Printf("Connecting to local k8s... \n")
			bmhClusterClient = utils.NewK8SClient()
		}
	}

	var nwcpClusterClient rtclient.Client
	if handlesw {
		if len(nwcpKubeconfig) > 0 {
			fmt.Printf("Connecting to nwcp k8s... \n")
			nwcpClusterClient = utils.NewK8SClientFromConfAndScheme(ctx, nwcpKubeconfig, scheme)
		} else {
			fmt.Printf("Connecting to local k8s... \n")
			nwcpClusterClient = utils.NewK8SClient()
		}
	}

	// 8 frontend switches
	sw := generateFrontEndSwitches(numGroup)
	// 96 accel switches
	accSW := generateAccelSwitches(numGroup)
	// 8 frontend switches
	strgSw := generateStorageSwitches(numGroup)
	// 8 NodeGroups
	groups := generateNodeGroups(numGroup)
	// 256 nodes
	bmhs := generateBMHs(numGroup, fabrics, idle)

	if action == "add" {
		add(bmhClusterClient, nwcpClusterClient, groups, sw, accSW, strgSw, bmhs, handlebmh, handlesw)
	} else if action == "delete" {
		delete(bmhClusterClient, nwcpClusterClient, groups, sw, accSW, strgSw, bmhs, handlebmh, handlesw)
	}
}

func add(bmhClient rtclient.Client, nwcpClient rtclient.Client, groups []idcnetworkv1alpha1.NodeGroup,
	sws []idcnetworkv1alpha1.Switch, accSWs []idcnetworkv1alpha1.Switch, strgSWs []idcnetworkv1alpha1.Switch,
	bmhs []*baremetalv1alpha1.BareMetalHost, handlebmh, handlesw bool) {
	ctx := context.Background()

	if handlesw {
		for _, sw := range sws {
			err := nwcpClient.Create(ctx, &sw, &rtclient.CreateOptions{})
			if err != nil {
				fmt.Printf("Create front-end Switches failed: %v \n", err)
			}
		}

		for _, accsw := range accSWs {
			err := nwcpClient.Create(ctx, &accsw, &rtclient.CreateOptions{})
			if err != nil {
				fmt.Printf("Create Accel Switches failed: %v \n", err)
			}
		}

		for _, strgsw := range strgSWs {
			err := nwcpClient.Create(ctx, &strgsw, &rtclient.CreateOptions{})
			if err != nil {
				fmt.Printf("Create Storage Switches failed: %v \n", err)
			}
		}
	}

	if handlebmh {
		for _, bmh := range bmhs {
			bmhCopy := bmh.DeepCopy()
			err := bmhClient.Create(ctx, bmh, &rtclient.CreateOptions{})
			if err != nil {
				fmt.Printf("Create BMH failed: %v \n", err)
			}
			latestBMH := &baremetalv1alpha1.BareMetalHost{}
			key := types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}

			err = bmhClient.Get(ctx, key, latestBMH)
			if err != nil {
				fmt.Printf("Get BMH failed: %v \n", err)
			}

			latestBMH.Status = bmhCopy.Status
			err = bmhClient.Status().Update(ctx, latestBMH)
			if err != nil {
				fmt.Printf("BMH Status().Update() failed: %v \n", err)
			}
		}
	}

	return
}

func delete(bmhClient rtclient.Client, nwcpClient rtclient.Client, groups []idcnetworkv1alpha1.NodeGroup,
	sws []idcnetworkv1alpha1.Switch, accSWs []idcnetworkv1alpha1.Switch, strgSWs []idcnetworkv1alpha1.Switch,
	bmhs []*baremetalv1alpha1.BareMetalHost, handlebmh, handlesw bool) {
	ctx := context.Background()
	if handlebmh {
		for _, bmh := range bmhs {

			// try to remove the finalizer
			bmhCR := &baremetalv1alpha1.BareMetalHost{}
			key := types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}
			err := bmhClient.Get(ctx, key, bmhCR)
			if err != nil {
				fmt.Printf("get NetworkNode failed: %v \n", err)
				continue
			}

			if len(bmhCR.Finalizers) > 0 {
				bmhCRCopy := bmhCR.DeepCopy()
				patch := client.MergeFrom(bmhCR)
				bmhCRCopy.Finalizers = nil
				if err := bmhClient.Patch(ctx, bmhCRCopy, patch); err != nil {
					fmt.Printf("Patch NetworkNode failed: %v \n", err)
					continue
				}
			}

			err = bmhClient.Delete(ctx, bmh, &rtclient.DeleteOptions{})
			if err != nil {
				fmt.Printf("Delete NetworkNode failed: %v \n", err)
			}
		}

	}

	if handlesw {
		for _, group := range groups {
			err := nwcpClient.Delete(ctx, &group, &rtclient.DeleteOptions{})
			if err != nil {
				fmt.Printf("Delete NodeGroup failed: %v \n", err)
			}
		}

		for _, sw := range sws {
			err := nwcpClient.Delete(ctx, &sw, &rtclient.DeleteOptions{})
			if err != nil {
				fmt.Printf("Delete front-end Switches failed: %v \n", err)
			}
		}

		for _, accsw := range accSWs {
			err := nwcpClient.Delete(ctx, &accsw, &rtclient.DeleteOptions{})
			if err != nil {
				fmt.Printf("Delete Accel Switches failed: %v \n", err)
			}
		}

		for _, strgsw := range strgSWs {
			err := nwcpClient.Delete(ctx, &strgsw, &rtclient.DeleteOptions{})
			if err != nil {
				fmt.Printf("Delete Storage Switches failed: %v \n", err)
			}
		}
	}

	return
}

func generateNodeGroups(numGroup int) []idcnetworkv1alpha1.NodeGroup {
	// groups := make([]idcnetworkv1alpha1.NodeGroup, 8)
	groups := make([]idcnetworkv1alpha1.NodeGroup, 0)
	for g := 1; g <= numGroup; g++ {
		groupName := fmt.Sprintf("group-%v", g)

		var group idcnetworkv1alpha1.NodeGroup
		labels := make(map[string]string)
		group = idcnetworkv1alpha1.NodeGroup{
			ObjectMeta: v1.ObjectMeta{
				Name:      groupName,
				Namespace: "idcs-system",
				Labels:    labels,
			},
			Spec: idcnetworkv1alpha1.NodeGroupSpec{},
		}
		groups = append(groups, group)
	}
	return groups
}

func generateBMHs(numGroup int, fabrics string, idle bool) []*baremetalv1alpha1.BareMetalHost {
	bmhs := make([]*baremetalv1alpha1.BareMetalHost, 0)
	cnt := 0
	for g := 1; g <= numGroup; g++ {
		for nn := 1; nn <= 8; nn++ {
			cnt++
			groupID := fmt.Sprintf("group-%v", g)
			nodeName := fmt.Sprintf("g%vn%v", g, nn)

			spID := (cnt-1)%32 + 1
			var feSPShortName, feSWName, mac string
			if strings.Contains(fabrics, "F") {
				// front end switch port
				feSwitchID := (cnt-1)/32 + 1
				feSPName := fmt.Sprintf("ethernet%v-1.fxhb3p3r-zal%internal-placeholder.com", spID, feSwitchID)
				feSPShortName, feSWName = utils.ExtractSwitchAndPortNameFromSwitchPortCRName(feSPName)
				mac = generateRandomMAC()
			}

			// accel switches port
			bmh := &baremetalv1alpha1.BareMetalHost{
				ObjectMeta: v1.ObjectMeta{
					Name:        nodeName,
					Namespace:   "metal3-1",
					Annotations: make(map[string]string),
					Labels: map[string]string{
						idcnetworkv1alpha1.LabelBMHGroupID: groupID,
					},
				},
				Spec: baremetalv1alpha1.BareMetalHostSpec{
					BootMACAddress: mac,
				},
				Status: baremetalv1alpha1.BareMetalHostStatus{
					HardwareDetails: &baremetalv1alpha1.HardwareDetails{
						NIC: []baremetalv1alpha1.NIC{
							baremetalv1alpha1.NIC{
								Name: feSPShortName,
								MAC:  mac,
								LLDP: baremetalv1alpha1.LLDP{
									SwitchPortId:     feSPShortName,
									SwitchSystemName: feSWName,
								},
							},
						},
					},
					OperationalStatus: baremetalv1alpha1.OperationalStatusOK,
				},
			}

			if strings.Contains(fabrics, "A") {
				accelSPCnt := 0
				for ply := 1; ply <= 3; ply++ {
					sw := g
					swfqdn := fmt.Sprintf("fxhb3p3r-zasp%v%internal-placeholder.com", ply, sw)
					for sp := 1; sp <= 8; sp++ {
						accelSPCnt++
						// (8 * 8): each switch has 64 ports that connected - 8 nodes, each node 8 ports.
						spNum := ((nn-1)*8+sp-1)%(8*8) + 1
						switchPort := fmt.Sprintf("ethernet%v-1.%v", spNum, swfqdn)

						accelSPShortName, accelSWName := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(switchPort)
						mac := generateRandomMAC()
						gpu := fmt.Sprintf("gpu.mac.cloud.intel.com/gpu-%v", accelSPCnt)
						bmh.Annotations[gpu] = mac
						bmh.Status.HardwareDetails.NIC = append(bmh.Status.HardwareDetails.NIC, baremetalv1alpha1.NIC{
							Name: accelSPShortName,
							MAC:  mac,
							LLDP: baremetalv1alpha1.LLDP{
								SwitchPortId:     accelSPShortName,
								SwitchSystemName: accelSWName,
							},
						})
					}
				}
			}

			if strings.Contains(fabrics, "S") {
				// storage sp
				strgSwitchID := (cnt-1)/32 + 1
				strgSPName := fmt.Sprintf("ethernet%v-1.fxhb3p3r-zals%internal-placeholder.com", spID, strgSwitchID)
				strgSPShortName, strgSWName := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(strgSPName)
				strgmac := generateRandomMAC()
				strgAnnotation := fmt.Sprintf("storage.mac.cloud.intel.com/storage-eth-%v", 1)
				bmh.Annotations[strgAnnotation] = strgmac
				bmh.Status.HardwareDetails.NIC = append(bmh.Status.HardwareDetails.NIC, baremetalv1alpha1.NIC{
					Name: strgSPShortName,
					MAC:  strgmac,
					LLDP: baremetalv1alpha1.LLDP{
						SwitchPortId:     strgSPShortName,
						SwitchSystemName: strgSWName,
					},
				})
			}

			//
			if !idle {
				bmh.Spec.ConsumerRef = &corev1.ObjectReference{
					Name: "some-instance",
				}
			}

			bmhs = append(bmhs, bmh)
		}

	}
	return bmhs
}

func generateRandomMAC() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("b0:fd:0b:%02x:%02x:%02x",
		rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func generateAccelSwitches(numGroup int) []idcnetworkv1alpha1.Switch {
	accelSwitches := make([]idcnetworkv1alpha1.Switch, 0)
	for ply := 1; ply <= 3; ply++ {
		for i := 1; i <= numGroup; i++ {
			swfqdn := fmt.Sprintf("fxhb3p3r-zasp%v%internal-placeholder.com", ply, i)
			ip := fmt.Sprintf("192.168.%v.%v", ply, i)
			sw := idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      swfqdn,
					Namespace: "idcs-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN: swfqdn,
					Ip:   ip,
					BGP:  &idcnetworkv1alpha1.BGPConfig{},
				},
			}
			accelSwitches = append(accelSwitches, sw)
		}
	}
	return accelSwitches
}

func generateFrontEndSwitches(numGroup int) []idcnetworkv1alpha1.Switch {
	feSwitches := make([]idcnetworkv1alpha1.Switch, 0)
	for i := 1; i <= numGroup; i++ {
		swfqdn := fmt.Sprintf("fxhb3p3r-zal%internal-placeholder.com", i)
		ip := fmt.Sprintf("192.168.10.%v", i)
		sw := idcnetworkv1alpha1.Switch{
			ObjectMeta: v1.ObjectMeta{
				Name:      swfqdn,
				Namespace: "idcs-system",
			},
			Spec: idcnetworkv1alpha1.SwitchSpec{
				FQDN: swfqdn,
				Ip:   ip,
				BGP:  &idcnetworkv1alpha1.BGPConfig{},
			},
		}
		feSwitches = append(feSwitches, sw)
	}
	return feSwitches
}

func generateStorageSwitches(numGroup int) []idcnetworkv1alpha1.Switch {
	strgSwitches := make([]idcnetworkv1alpha1.Switch, 0)
	for i := 1; i <= numGroup; i++ {
		swfqdn := fmt.Sprintf("fxhb3p3r-zals%internal-placeholder.com", i)
		ip := fmt.Sprintf("192.168.11.%v", i)
		sw := idcnetworkv1alpha1.Switch{
			ObjectMeta: v1.ObjectMeta{
				Name:      swfqdn,
				Namespace: "idcs-system",
			},
			Spec: idcnetworkv1alpha1.SwitchSpec{
				FQDN: swfqdn,
				Ip:   ip,
				BGP:  &idcnetworkv1alpha1.BGPConfig{},
			},
		}
		strgSwitches = append(strgSwitches, sw)
	}
	return strgSwitches
}
