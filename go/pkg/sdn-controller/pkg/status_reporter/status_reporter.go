package statusreporter

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-test/deep"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"go.opentelemetry.io/otel/codes"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	devicesmanager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StatusReporter manages StatusReporterWorkers. It will spawn a worker instance per switch.
// StatusReporterWorker fetches the config of a switch and update the status of the SwitchPort and Switch CRD.
type StatusReporter struct {
	sync.Mutex
	reportIntervalInSec            int
	reportAcceleratedIntervalInSec int
	deviceAccessManager            devicesmanager.DevicesAccessManager
	networkK8sClient               client.Client
	statusReportWorkers            map[string]*statusReportWorker
	ctx                            context.Context
	statusReporterRecorder         record.EventRecorder
	BGPCommunityIncomingGroupName  string
	portChannelsEnabled            bool
}

type StatusReporterConfig struct {
	ReportIntervalInSec            int
	ReportAcceleratedIntervalInSec int
	DeviceAccessManager            devicesmanager.DevicesAccessManager
	NetworkK8sClient               client.Client
	Ctx                            context.Context
	StatusReporterRecorder         record.EventRecorder
	BGPCommunityIncomingGroupName  string
	PortChannelsEnabled            bool
}

func NewStatusReporter(conf StatusReporterConfig) *StatusReporter {
	statusReporter := &StatusReporter{
		reportIntervalInSec:            conf.ReportIntervalInSec,
		reportAcceleratedIntervalInSec: conf.ReportAcceleratedIntervalInSec,
		networkK8sClient:               conf.NetworkK8sClient,
		deviceAccessManager:            conf.DeviceAccessManager,
		statusReportWorkers:            make(map[string]*statusReportWorker),
		ctx:                            conf.Ctx,
		statusReporterRecorder:         conf.StatusReporterRecorder,
		BGPCommunityIncomingGroupName:  conf.BGPCommunityIncomingGroupName,
		portChannelsEnabled:            conf.PortChannelsEnabled,
	}

	return statusReporter
}

func (sr *StatusReporter) AddSwitch(switchFQDN string) error {
	logger := log.FromContext(sr.ctx).WithName("StatusReporter.AddSwitch")

	sr.Lock()
	defer sr.Unlock()
	// if we already have a worker, return directly.
	_, found := sr.statusReportWorkers[switchFQDN]
	if found {
		return nil
	}

	wCtx, wCancel := context.WithCancel(sr.ctx)
	// spawn a worker for monitoring the given switch
	statusReportWorker := &statusReportWorker{
		switchFQDN:               switchFQDN,
		networkK8sClient:         sr.networkK8sClient,
		deviceAccessManager:      sr.deviceAccessManager,
		reportIntervalInSec:      sr.reportIntervalInSec,
		acceleratedIntervalInSec: sr.reportAcceleratedIntervalInSec,
		ctx:                      wCtx,
		cancel:                   wCancel,
		statusReporterRecorder:   sr.statusReporterRecorder,
		BGPCommunityGroupName:    sr.BGPCommunityIncomingGroupName,
		portChannelsEnabled:      sr.portChannelsEnabled,
	}
	go statusReportWorker.Start()
	sr.statusReportWorkers[switchFQDN] = statusReportWorker
	logger.Info(fmt.Sprintf("successfully added status worker for %v", switchFQDN))
	return nil
}

func (sr *StatusReporter) RemoveSwitch(switchFQDN string) error {
	logger := log.FromContext(sr.ctx).WithName("StatusReporter.RemoveSwitch")
	sr.Lock()
	defer sr.Unlock()
	// stop and delete the worker
	statueReporter := sr.statusReportWorkers[switchFQDN]
	if statueReporter != nil {
		statueReporter.Stop()
	}
	delete(sr.statusReportWorkers, switchFQDN)
	logger.Info(fmt.Sprintf("successfully removed status worker for %v", switchFQDN))
	return nil
}

// AccelerateStatusUpdate does a one-time status update for the given switch, in a short time. For when we expect an imminent change.
// It will "debounce" calls, so this can be called multiple times in a short time, and only the first one will be executed.
// It handles errors if the accelerate can't be completed and logs them using the logger, but does not fail/return an error.
func (sr *StatusReporter) AccelerateStatusUpdate(switchFQDN string) {
	logger := log.FromContext(sr.ctx).WithName("statusReporter.AccelerateStatusUpdate").WithValues(utils.LogFieldSwitchFQDN, switchFQDN)
	srw, found := sr.statusReportWorkers[switchFQDN]
	if !found || srw == nil {
		logger.Error(fmt.Errorf("status worker not found for %v", switchFQDN), "status worker not found")
		return
	}
	srw.AccelerateTick()
}

type statusReportWorker struct {
	sync.Mutex
	switchFQDN               string
	networkK8sClient         client.Client
	deviceAccessManager      devicesmanager.DevicesAccessManager
	reportIntervalInSec      int
	ctx                      context.Context
	cancel                   context.CancelFunc
	acceleratedIntervalInSec int
	acceleratedTimer         *time.Timer
	statusReporterRecorder   record.EventRecorder
	BGPCommunityGroupName    string
	portChannelsEnabled      bool
}

func (w *statusReportWorker) Start() {
	time.Sleep(time.Duration(rand.Intn(w.reportIntervalInSec)) * time.Second) // Wait for a short time to avoid all switches updating at the same time
	logger := log.FromContext(w.ctx).WithName("statusReportWorker.Start").WithValues(utils.LogFieldSwitchFQDN, w.switchFQDN)
	ticker := time.NewTicker(time.Duration(w.reportIntervalInSec) * time.Second)
	defer ticker.Stop()

	for {

		err := w.performStatusUpdate()
		if err != nil {
			logger.Error(err, "update status failed for Switch", utils.LogFieldSwitchFQDN, w.switchFQDN)
		}

		logger.V(1).Info(fmt.Sprintf("finished updating status for [%v]", w.switchFQDN))

		select {
		case <-ticker.C:
			continue
		case <-w.ctx.Done():
			logger.Info(fmt.Sprintf("context is done, stopping the reporter, err: %v", w.ctx.Err()))
			return
		}
	}
}

func (w *statusReportWorker) AccelerateTick() {
	logger := log.FromContext(w.ctx).WithName("statusReportWorker.AccelerateTick").WithValues(utils.LogFieldSwitchFQDN, w.switchFQDN)

	// Prevents case where 2x AccelerateTick() calls in quick succession could get "inside" the if statement before w.acceleratedTimer gets set to a non-nil value.
	w.Lock()
	defer w.Unlock()

	if w.acceleratedTimer == nil {
		accelerateTime := time.Duration(math.Max(1, float64(w.acceleratedIntervalInSec))) * time.Second
		logger.V(1).Info(fmt.Sprintf("Accelerating status update for Switch. Will check status in %v", accelerateTime))
		w.acceleratedTimer = time.AfterFunc(accelerateTime, func() {
			logger.V(1).Info(fmt.Sprintf("Accelerated update firing for %v", w.switchFQDN))
			w.acceleratedTimer = nil // Reset so later changes can trigger another accelerated tick
			err := w.performStatusUpdate()
			if err != nil {
				logger.Error(err, "update status failed for Switch", utils.LogFieldSwitchFQDN, w.switchFQDN)
			}
		})
	} else { // Already accelerated, skip
		logger.V(1).Info(fmt.Sprintf("Skipping acceleration because one already is in progress."))
	}
}

func (w *statusReportWorker) performStatusUpdate() error {
	spanCtx, logger, span := obs.LogAndSpanFromContextOrGlobal(w.ctx).WithName("statusReportWorker.performStatusUpdate").WithValues(utils.LogFieldSwitchFQDN, w.switchFQDN).Start()
	defer span.End()

	// get the latest switch CR
	sw := &idcnetworkv1alpha1.Switch{}
	key := types.NamespacedName{Name: w.switchFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := w.networkK8sClient.Get(spanCtx, key, sw)
	if err != nil {
		logger.Error(err, "unable to fetch Switch CR", utils.LogFieldSwitchFQDN, w.switchFQDN)
		span.SetStatus(codes.Error, "unable to fetch Switch CR")
		return err

	}

	// get the switch client
	switchClient, err := w.deviceAccessManager.GetSwitchClient(devicesmanager.GetOption{SwitchFQDN: w.switchFQDN})
	if err != nil {
		logger.Error(err, "deviceAccessManager.GetSwitchClient failed")
		span.SetStatus(codes.Error, "deviceAccessManager.GetSwitchClient failed")
		w.statusReporterRecorder.Event(sw, corev1.EventTypeWarning, "deviceAccessManager.GetSwitchClient failed", err.Error())
		return err
	}

	// Use a bool to avoid multiple calls to w.AccelerateTick() which could result in an exponential number of calls to performStatusUpdate() if rate-limiting somehow failed.
	var shouldAccelerateSwitchStatusCheck = false

	//////////////////////////////
	// update Switch CR status
	//////////////////////////////
	func() {
		// fetch the switch BGP data
		req := sc.GetBGPCommunityRequest{
			BGPCommunityIncomingGroupName: w.BGPCommunityGroupName,
		}
		bgp, err := switchClient.GetBGPCommunity(spanCtx, req)
		if err != nil {
			logger.V(1).Info(fmt.Sprintf("switchClient.GetBGPCommunity failed, req: %v, reason: %v", req, err))

			// If it failed because we got we couldn't parse the response, or it's not a BGP switch, then we "successfully checked" the switch, so still want to update the CR's status.
			// If it failed because we couldn't connect to the switch, do not update the "last checked time".
			// TODO: Better way of determining whether we could connect to switch or not than parsing the messsage.
			if !(strings.Contains(err.Error(), w.BGPCommunityGroupName) || strings.Contains(err.Error(), "communityValue")) {
				w.statusReporterRecorder.Event(sw, corev1.EventTypeWarning, "switchClient.GetBGPCommunity failed", err.Error())
				return
			}
		}

		if sw.Status.SwitchBGPConfigStatus == nil {
			sw.Status.SwitchBGPConfigStatus = &idcnetworkv1alpha1.SwitchBGPConfigStatus{}
		}
		if bgp != 0 {
			sw.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity = int64(bgp)

			if sw.Spec.BGP != nil && sw.Spec.BGP.BGPCommunity != -1 && sw.Spec.BGP.BGPCommunity != sw.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity {
				logger.V(1).Info("BGP desired not in sync with actual, accelerating status check")
				shouldAccelerateSwitchStatusCheck = true // If it's still not in-sync, do another fast update because it's likely to change soon.
			}
		}
		sw.Status.LastStatusUpdateTime = metav1.Now()

		err = w.networkK8sClient.Status().Update(spanCtx, sw)
		if err != nil {
			logger.Error(err, "switch CR status update failed", utils.LogFieldSwitchFQDN, w.switchFQDN)
		}
	}()

	//////////////////////////////
	// update SwitchPorts & PortChannel CR status
	//////////////////////////////
	func() {
		// Get the actual switch config from the switch.
		allInterfacesStatus, err := switchClient.GetSwitchPorts(spanCtx, sc.GetSwitchPortsRequest{
			SwitchFQDN: w.switchFQDN,
		})
		if err != nil {
			logger.Error(err, "switchClient.GetSwitchPorts failed", utils.LogFieldSwitchFQDN, w.switchFQDN)
			w.statusReporterRecorder.Event(sw, corev1.EventTypeWarning, "switchClient.GetSwitchPorts failed", err.Error())
			span.SetStatus(codes.Error, "switchClient.GetSwitchPorts failed")
			return
		}

		// get all switch ports crs for this switch
		labelSelector := labels.SelectorFromSet(labels.Set{
			idcnetworkv1alpha1.LabelNameSwitchFQDN: w.switchFQDN,
		})

		listOpts := &client.ListOptions{
			LabelSelector: labelSelector,
		}
		allSwitchPortCRs := &idcnetworkv1alpha1.SwitchPortList{}
		err = w.networkK8sClient.List(spanCtx, allSwitchPortCRs, listOpts)
		if err != nil {
			logger.Error(err, "networkK8sClient.List failed", utils.LogFieldSwitchFQDN, w.switchFQDN)
			span.SetStatus(codes.Error, "networkK8sClient.List failed")
			return
		}

		for _, spCR := range allSwitchPortCRs.Items {
			actualPortConf, found := allInterfacesStatus[spCR.Spec.Name]

			if !found || actualPortConf == nil {
				logger.Error(err, "cannot find switch port config in response from switch", utils.LogFieldSwitchPortCRName, spCR.Name)
				continue
			}

			lastVlan := spCR.Status.VlanId

			spCRForStatusUpdate := spCR.DeepCopy()

			spCRForStatusUpdate.Status.Name = actualPortConf.Name
			// when a port is set to trunk mode, the vlan id will be 0.
			spCRForStatusUpdate.Status.VlanId = actualPortConf.VlanId
			spCRForStatusUpdate.Status.Mode = actualPortConf.Mode
			spCRForStatusUpdate.Status.NativeVlan = actualPortConf.NativeVlan
			spCRForStatusUpdate.Status.LinkStatus = actualPortConf.LinkStatus
			spCRForStatusUpdate.Status.LineProtocolStatus = actualPortConf.LineProtocolStatus
			spCRForStatusUpdate.Status.PortChannel = actualPortConf.PortChannel
			spCRForStatusUpdate.Status.TrunkGroups = actualPortConf.TrunkGroups
			spCRForStatusUpdate.Status.Description = actualPortConf.Description
			spCRForStatusUpdate.Status.Bandwidth = actualPortConf.Bandwidth
			spCRForStatusUpdate.Status.Duplex = actualPortConf.Duplex
			spCRForStatusUpdate.Status.SwitchSideLastStatusChangeTimestamp = actualPortConf.SwitchSideLastStatusChangeTimestamp

			// only update status if there is any changes since last update.
			if !reflect.DeepEqual(spCR.Status, spCRForStatusUpdate.Status) {
				logger.V(1).Info("Observed a change on switchport, updating status & accelerating", utils.LogFieldSwitchPortCRName, spCR.Name, "Diff", deep.Equal(spCR.Status, spCRForStatusUpdate.Status))

				spCRForStatusUpdate.Status.LastStatusChangeTime = metav1.Now()
				err = w.networkK8sClient.Status().Update(spanCtx, spCRForStatusUpdate)
				if err != nil {
					logger.Error(err, "update status failed for SwitchPort CR", utils.LogFieldSwitchPortCRName, spCR.Name)
				}
				shouldAccelerateSwitchStatusCheck = true // If something changed, keep doing fast updates because something is going on (server restart, vlan change, etc.).
			}

			// If it's still not in-sync, do another fast update because it's likely to change soon.
			if utils.UpdateSwitchConfRequired(spCRForStatusUpdate) && actualPortConf.VlanId != 0 {
				logger.V(1).Info(fmt.Sprintf("desired vlanId %d doesn't match actual %d for %s, accelerating", spCR.Spec.VlanId, actualPortConf.VlanId, spCR.Name))
				shouldAccelerateSwitchStatusCheck = true
			}

			// record an event if vlan has unexpectedly changed on the switch
			if spCR.Spec.VlanId == lastVlan && spCR.Spec.VlanId != actualPortConf.VlanId && actualPortConf.VlanId != 0 {
				logger.Info(fmt.Sprintf("unexpected VLAN value [%v] is observed for switch port [%v]. Last saw [%v]", actualPortConf.VlanId, spCR.Spec.Name, lastVlan))
				w.statusReporterRecorder.Event(&spCR, corev1.EventTypeWarning, "unexpected VLAN value", fmt.Sprintf("unexpected VLAN value [%v] is observed. Last saw [%v]", actualPortConf.VlanId, lastVlan))
			}
		}

		if w.portChannelsEnabled {
			notControlledPortChannels := CreateNotControlledPortChannelsMap(allInterfacesStatus, allSwitchPortCRs)

			// Port Channels come back in the same ""GetSwitchPorts"" call, so update those here too.
			allPortChannelCRs := &idcnetworkv1alpha1.PortChannelList{}
			err = w.networkK8sClient.List(w.ctx, allPortChannelCRs, listOpts)
			if err != nil {
				logger.Error(err, "networkK8sClient.List portchannels failed", utils.LogFieldSwitchFQDN, w.switchFQDN)
				return
			}

			// If we discovered new PortChannels that exist on the switch, but don't have a CR, create them.
			for _, actualPc := range allInterfacesStatus {
				if !strings.Contains(actualPc.Name, "Port-Channel") {
					continue
				}

				var existingPcCR idcnetworkv1alpha1.PortChannel
				var found = false

				// Find the CR for this PortChannel if it exists
				for _, pcCR := range allPortChannelCRs.Items {
					if pcCR.Spec.Name == actualPc.Name {
						existingPcCR = pcCR
						found = true
						break
					}
				}
				if !found {
					// Only create PortChannel CRs for those that should be modified by the Provider-SDN Controller (eg. do not import spine-link PortChannels)
					pcNumber, err := utils.PortChannelInterfaceNameToNumber(actualPc.Name)
					if err != nil {
						logger.Error(err, "PortChannelInterfaceNameToNumber failed", utils.LogFieldSwitchFQDN, w.switchFQDN, utils.LogFieldPortChannelInterfaceName, actualPc.Name)
						continue
					}
					nonControlledSwitchPortMember, shouldNotBeControlled := notControlledPortChannels[pcNumber]
					if shouldNotBeControlled {
						logger.V(1).Info(fmt.Sprintf("not creating portchannel CR for %v because it contains an interface that is not owned by SDN controller, %v ", actualPc.Name, nonControlledSwitchPortMember))
						continue
					}

					// create new PortChannel CR
					newPortChannelCRName, err := utils.PortChannelInterfaceNameAndSwitchFQDNToCRName(actualPc.Name, w.switchFQDN)
					if err != nil {
						logger.Error(err, "PortChannelInterfaceNameAndSwitchFQDNToCRName failed", utils.LogFieldSwitchFQDN, w.switchFQDN, utils.LogFieldPortChannelCRName, newPortChannelCRName)
						continue
					}

					newPcCR := &idcnetworkv1alpha1.PortChannel{
						ObjectMeta: metav1.ObjectMeta{
							Name:      newPortChannelCRName,
							Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
							Labels: map[string]string{
								idcnetworkv1alpha1.LabelNameSwitchFQDN: w.switchFQDN,
							},
						},
						Spec: idcnetworkv1alpha1.PortChannelSpec{
							Name: actualPc.Name,
						},
					}
					err = w.networkK8sClient.Create(w.ctx, newPcCR)
					if err != nil {
						logger.Error(err, "create PortChannel CR failed", utils.LogFieldSwitchFQDN, w.switchFQDN, utils.LogFieldPortChannelCRName, newPortChannelCRName)
						continue
					}

					logger.Info("created new PortChannel CR. Scheduling accelerated check to populate its status", utils.LogFieldSwitchFQDN, w.switchFQDN, utils.LogFieldPortChannelCRName, newPortChannelCRName)
					shouldAccelerateSwitchStatusCheck = true

				} else {
					pcCRForStatusUpdate := existingPcCR.DeepCopy()

					// Update PortChannel CRs status
					pcCRForStatusUpdate.Status.Name = actualPc.Name
					pcCRForStatusUpdate.Status.VlanId = actualPc.VlanId
					pcCRForStatusUpdate.Status.Mode = actualPc.Mode
					pcCRForStatusUpdate.Status.NativeVlan = actualPc.NativeVlan
					pcCRForStatusUpdate.Status.LinkStatus = actualPc.LinkStatus
					pcCRForStatusUpdate.Status.TrunkGroups = actualPc.TrunkGroups
					pcCRForStatusUpdate.Status.Description = actualPc.Description

					// only update status if there is any changes.
					if !reflect.DeepEqual(existingPcCR.Status, pcCRForStatusUpdate.Status) {
						logger.Info("Portchannel config change observed on switch. Updating PortChannel CR's status. ", utils.LogFieldSwitchFQDN, w.switchFQDN, utils.LogFieldPortChannelCRName, existingPcCR.Name)
						pcCRForStatusUpdate.Status.LastStatusChangeTime = metav1.Now()
						err = w.networkK8sClient.Status().Update(w.ctx, pcCRForStatusUpdate)
						if err != nil {
							logger.Error(err, "update status failed for PortChannel CR", utils.LogFieldSwitchPortCRName, existingPcCR.Name)
						}
					}
				}
			}

			// Check the reverse direction - Portchannel CRs that no longer exist on the switch so we can update their observed status (to empty).
			for _, portChannelCR := range allPortChannelCRs.Items {

				// Find the actual PortChannel from the switch.
				foundOnSwitch := false
				for _, actualPc := range allInterfacesStatus {
					if actualPc.Name == portChannelCR.Spec.Name {
						foundOnSwitch = true
						break
					}
				}

				if !foundOnSwitch {
					logger.Info("Did not find portchannel on switch. Clearing status for PortChannel CR", utils.LogFieldSwitchFQDN, w.switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Name)
					// If we didn't find the actual PortChannel on the switch, update the CR status to empty.
					pcCRForStatusUpdate := portChannelCR.DeepCopy()
					pcCRForStatusUpdate.Status.Name = ""
					pcCRForStatusUpdate.Status.VlanId = 0
					pcCRForStatusUpdate.Status.Mode = ""
					pcCRForStatusUpdate.Status.NativeVlan = 0
					pcCRForStatusUpdate.Status.LinkStatus = ""
					pcCRForStatusUpdate.Status.TrunkGroups = nil
					pcCRForStatusUpdate.Status.Description = ""

					// only update status if there is any changes.
					if !reflect.DeepEqual(portChannelCR.Status, pcCRForStatusUpdate.Status) {
						pcCRForStatusUpdate.Status.LastStatusChangeTime = metav1.Now()
						err = w.networkK8sClient.Status().Update(w.ctx, pcCRForStatusUpdate)
						if err != nil {
							logger.Error(err, "clearing CR status failed for deleted portchannel", utils.LogFieldSwitchPortCRName, portChannelCR.Name)
						}
					}
				}
			}

		}

	}()

	if shouldAccelerateSwitchStatusCheck {
		w.AccelerateTick()
	}

	return nil
}

func CreateNotControlledPortChannelsMap(allInterfacesStatus map[string]*idcnetworkv1alpha1.SwitchPortStatus, allSwitchPortCRs *idcnetworkv1alpha1.SwitchPortList) map[int]string {
	// If a portchannel contains a switchport that is NOT owned by the SDNController (eg. a link to a spine), then do not create that portchannel CR - we do not want to control it.
	// We DO want to control portchannels that contain switchports that are owned by the SDNController, or empty portchannels that contain NO interfaces.
	// Make a list of all portchannels that contain switchports we do NOT control (interfaces returned from the switch, but that we do NOT have a CR for).
	notControlledPortChannels := make(map[int]string)
	for _, actualInterface := range allInterfacesStatus {
		if !strings.Contains(actualInterface.Name, "Ethernet") {
			continue
		}

		// Identify if this interface is controlled by the SDNController
		var controlled = false
		for _, spCR := range allSwitchPortCRs.Items {
			if spCR.Spec.Name == actualInterface.Name {
				// This interface is controlled by the SDNController
				controlled = true
				break
			}
		}

		if !controlled && actualInterface.PortChannel != 0 {
			notControlledPortChannels[int(actualInterface.PortChannel)] = actualInterface.Name
		}
	}
	return notControlledPortChannels
}

func (w *statusReportWorker) Stop() {
	w.cancel()
}
