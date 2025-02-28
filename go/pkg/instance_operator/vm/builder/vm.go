// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builder

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	DefaultVMGenerateName = "harv-"
	DefaultVMNamespace    = "default"

	DefaultVMCPUCores = 1
	DefaultVMMemory   = "256Mi"

	HarvesterAPIGroup                                     = "harvesterhci.io"
	LabelAnnotationPrefixHarvester                        = HarvesterAPIGroup + "/"
	LabelKeyVirtualMachineCreator                         = LabelAnnotationPrefixHarvester + "creator"
	LabelKeyVirtualMachineName                            = LabelAnnotationPrefixHarvester + "vmName"
	AnnotationKeyVirtualMachineSSHNames                   = LabelAnnotationPrefixHarvester + "sshNames"
	AnnotationKeyVirtualMachineWaitForLeaseInterfaceNames = LabelAnnotationPrefixHarvester + "waitForLeaseInterfaceNames"
	AnnotationKeyVirtualMachineDiskNames                  = LabelAnnotationPrefixHarvester + "diskNames"
	AnnotationKeyImageID                                  = LabelAnnotationPrefixHarvester + "imageId"

	AnnotationPrefixCattleField = "field.cattle.io/"
	LabelPrefixHarvesterTag     = "tag.harvesterhci.io/"
	AnnotationKeyDescription    = AnnotationPrefixCattleField + "description"
	AnnotationKeyReservedMemory = LabelAnnotationPrefixHarvester + "reservedMemory"
)

type VMBuilder struct {
	VirtualMachine             *kubevirtv1.VirtualMachine
	SSHNames                   []string
	WaitForLeaseInterfaceNames []string
}

func NewVMBuilder(creator string) *VMBuilder {
	vmLabels := map[string]string{
		LabelKeyVirtualMachineCreator: creator,
	}
	objectMeta := metav1.ObjectMeta{
		Namespace:    DefaultVMNamespace,
		GenerateName: DefaultVMGenerateName,
		Labels:       vmLabels,
		Annotations:  map[string]string{},
	}
	runStrategy := kubevirtv1.RunStrategyHalted
	cpu := &kubevirtv1.CPU{
		Cores: DefaultVMCPUCores,
	}
	resources := kubevirtv1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(DefaultVMMemory),
			corev1.ResourceCPU:    *resource.NewQuantity(DefaultVMCPUCores, resource.DecimalSI),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(DefaultVMMemory),
			corev1.ResourceCPU:    *resource.NewQuantity(DefaultVMCPUCores, resource.DecimalSI),
		},
	}
	template := &kubevirtv1.VirtualMachineInstanceTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: vmLabels,
		},
		Spec: kubevirtv1.VirtualMachineInstanceSpec{
			Domain: kubevirtv1.DomainSpec{
				CPU: cpu,
				Devices: kubevirtv1.Devices{
					Disks:       []kubevirtv1.Disk{},
					Interfaces:  []kubevirtv1.Interface{},
					HostDevices: []kubevirtv1.HostDevice{},
				},
				Resources: resources,
			},
			Affinity: &corev1.Affinity{},
			Networks: []kubevirtv1.Network{},
			Volumes:  []kubevirtv1.Volume{},
		},
	}

	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: objectMeta,
		Spec: kubevirtv1.VirtualMachineSpec{
			RunStrategy: &runStrategy,
			Template:    template,
		},
	}
	return &VMBuilder{
		VirtualMachine:             vm,
		SSHNames:                   []string{},
		WaitForLeaseInterfaceNames: []string{},
	}
}

func (v *VMBuilder) Name(name string) *VMBuilder {
	v.VirtualMachine.ObjectMeta.Name = name
	v.VirtualMachine.ObjectMeta.GenerateName = ""
	v.VirtualMachine.Spec.Template.ObjectMeta.Labels[LabelKeyVirtualMachineName] = name
	return v
}

func (v *VMBuilder) Namespace(namespace string) *VMBuilder {
	v.VirtualMachine.ObjectMeta.Namespace = namespace
	return v
}

func (v *VMBuilder) MachineType(machineType string) *VMBuilder {
	v.VirtualMachine.Spec.Template.Spec.Domain.Machine = &kubevirtv1.Machine{
		Type: machineType,
	}
	return v
}

func (v *VMBuilder) HostName(hostname string) *VMBuilder {
	v.VirtualMachine.Spec.Template.Spec.Hostname = hostname
	return v
}

func (v *VMBuilder) Description(description string) *VMBuilder {
	if v.VirtualMachine.ObjectMeta.Annotations == nil {
		v.VirtualMachine.ObjectMeta.Annotations = map[string]string{}
	}
	v.VirtualMachine.ObjectMeta.Annotations[AnnotationKeyDescription] = description
	return v
}

func (v *VMBuilder) Labels(labels map[string]string) *VMBuilder {
	if v.VirtualMachine.ObjectMeta.Labels == nil {
		v.VirtualMachine.ObjectMeta.Labels = labels
	}
	for key, value := range labels {
		v.VirtualMachine.ObjectMeta.Labels[key] = value
	}
	return v
}

func (v *VMBuilder) HostDevices(ctx context.Context, instanceTypeSpec cloudv1alpha1.InstanceTypeSpec) *VMBuilder {
	log := log.FromContext(ctx).WithName("VMBuilder.HostDevices")
	log.Info("BEGIN")
	defer log.Info("END")
	var deviceName string
	var hostDeviceId string
	gpuCount := int(instanceTypeSpec.Gpu.Count)

	log.Info("configuring HostDevices", logkeys.GpuModelName, instanceTypeSpec.Gpu.ModelName, logkeys.GpuCount, gpuCount)

	//TODO: Fetch hardcoded values from configurations
	switch instanceTypeSpec.Gpu.ModelName {
	case "HL-225":
		deviceName = "habana.com/GAUDI2_AI_TRAINING_ACCELERATOR"
		hostDeviceId = "GAUDI2_AI_TRAINING_ACCELERATOR"
	case "gpu-max-1100":
		deviceName = "intel.com/PONTE_VECCHIO_XT_1_TILE_DATA_CENTER_GPU_MAX_1100"
		hostDeviceId = "PONTE_VECCHIO_XT_1_TILE_DATA_CENTER_GPU_MAX_1100"
	default:
		log.Info("Unsupported GPU model", logkeys.GpuModelName, instanceTypeSpec.Gpu.ModelName)
		return v
	}

	for i := 1; i <= gpuCount; i++ {
		hostDevice := kubevirtv1.HostDevice{
			Name:       hostDeviceId + strconv.Itoa(i),
			DeviceName: deviceName,
		}
		v.VirtualMachine.Spec.Template.Spec.Domain.Devices.HostDevices = append(v.VirtualMachine.Spec.Template.Spec.Domain.Devices.HostDevices, hostDevice)
	}

	return v
}

func (v *VMBuilder) Annotations(annotations map[string]string) *VMBuilder {
	if v.VirtualMachine.ObjectMeta.Annotations == nil {
		v.VirtualMachine.ObjectMeta.Annotations = annotations
	}
	for key, value := range annotations {
		v.VirtualMachine.ObjectMeta.Annotations[key] = value
	}
	return v
}

func (v *VMBuilder) Memory(ctx context.Context, memory string, gpuCount int32) *VMBuilder {
	if v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits == nil {
		v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits = corev1.ResourceList{}
	}

	memoryQty := resource.MustParse(memory)
	if vmOverheadMemoryQty := util.GetVmOverheadMemory(ctx, memoryQty, gpuCount); vmOverheadMemoryQty.Value() > 0 {
		// Extend VM memory with overhead memory only for non-GPU instance types
		if gpuCount == 0 {
			memoryQty.Add(vmOverheadMemoryQty)
		}
		// Set the reserved memory annotation for all instances
		v.VirtualMachine.ObjectMeta.Annotations[AnnotationKeyReservedMemory] = vmOverheadMemoryQty.String()
	}

	v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits[corev1.ResourceMemory] = memoryQty
	return v
}

func (v *VMBuilder) CPU(cores int) *VMBuilder {
	if len(v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits) == 0 {
		v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits = corev1.ResourceList{}
	}
	v.VirtualMachine.Spec.Template.Spec.Domain.Resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(int64(cores), resource.DecimalSI)
	return v
}

func (v *VMBuilder) EvictionStrategy(liveMigrate bool) *VMBuilder {
	if liveMigrate {
		evictionStrategy := kubevirtv1.EvictionStrategyLiveMigrate
		v.VirtualMachine.Spec.Template.Spec.EvictionStrategy = &evictionStrategy
	}
	return v
}

func (v *VMBuilder) Affinity(affinity *corev1.Affinity) *VMBuilder {
	if affinity == nil {
		return v.DefaultPodAntiAffinity()
	}

	v.VirtualMachine.Spec.Template.Spec.Affinity = affinity
	return v
}

func (v *VMBuilder) DefaultPodAntiAffinity() *VMBuilder {
	podAffinityTerm := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      LabelKeyVirtualMachineCreator,
					Operator: metav1.LabelSelectorOpExists,
				},
			},
		},
		TopologyKey: corev1.LabelHostname,
	}
	return v.PodAntiAffinity(podAffinityTerm, true, 100)
}

func (v *VMBuilder) PodAntiAffinity(podAffinityTerm corev1.PodAffinityTerm, soft bool, weight int32) *VMBuilder {
	podAffinity := &corev1.PodAntiAffinity{}
	if soft {
		podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
			{
				Weight:          weight,
				PodAffinityTerm: podAffinityTerm,
			},
		}
	} else {
		podAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []corev1.PodAffinityTerm{
			podAffinityTerm,
		}
	}
	v.VirtualMachine.Spec.Template.Spec.Affinity.PodAntiAffinity = podAffinity
	return v
}

func (v *VMBuilder) TopologySpreadConstraints(topologySpreadConstraints []corev1.TopologySpreadConstraint) *VMBuilder {
	v.VirtualMachine.Spec.Template.Spec.TopologySpreadConstraints = topologySpreadConstraints
	return v
}

func (v *VMBuilder) Run(start bool) *VMBuilder {
	runStrategy := kubevirtv1.RunStrategyHalted
	if start {
		// The expectation here is that the runStrategy is RerunOnFailure, however this can lead to a rare issue between
		// the scheduler's cached view of the allocatable resources and the actual allocatable resources. In the event of
		// a failure it's possible for two Pods to overlap in execution (one is being deleted, one is being created). When
		// this occurs the cache incorrectly sees the two Pods as one and miscalculates the allocatable resources. This is
		// described in more detail in https://internal-placeholder.com/browse/TWC4727-1796.
		//
		// To avoid the issue, set the run strategy to once to avoid the possibly problematic restart.
		runStrategy = kubevirtv1.RunStrategyOnce
	}
	v.VirtualMachine.Spec.RunStrategy = &runStrategy
	return v
}

func (v *VMBuilder) RunStrategy(runStrategy kubevirtv1.VirtualMachineRunStrategy) *VMBuilder {
	v.VirtualMachine.Spec.RunStrategy = &runStrategy
	return v
}

func (v *VMBuilder) Build() (*kubevirtv1.VirtualMachine, error) {
	if v.VirtualMachine.Spec.Template.ObjectMeta.Annotations == nil {
		v.VirtualMachine.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	sshNames, err := json.Marshal(v.SSHNames)
	if err != nil {
		return nil, err
	}
	v.VirtualMachine.Spec.Template.ObjectMeta.Annotations[AnnotationKeyVirtualMachineSSHNames] = string(sshNames)

	waitForLeaseInterfaceNames, err := json.Marshal(v.WaitForLeaseInterfaceNames)
	if err != nil {
		return nil, err
	}
	v.VirtualMachine.Spec.Template.ObjectMeta.Annotations[AnnotationKeyVirtualMachineWaitForLeaseInterfaceNames] = string(waitForLeaseInterfaceNames)

	return v.VirtualMachine, nil
}

func (v *VMBuilder) Update(vm *kubevirtv1.VirtualMachine) *VMBuilder {
	v.VirtualMachine = vm
	return v
}
