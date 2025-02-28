// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func MapInstanceTypeFromPbToString(instType v1.MachineType) string {
	switch instType {
	case v1.MachineType_LRG_VM_TYPE:
		return "vm-spr-lrg"
	case v1.MachineType_MED_VM_TYPE:
		return "vm-spr-med"
	case v1.MachineType_SML_VM_TYPE:
		return "vm-spr-sml"
	case v1.MachineType_GAUDI_BM_TYPE:
		return "bm-icp-gaudi2"
	case v1.MachineType_PVC_BM_1100_4:
		return "bm-spr-pvc-1100-4"
	case v1.MachineType_PVC_BM_1100_8:
		return "bm-spr-pvc-1100-8"
	case v1.MachineType_PVC_BM_1550_8:
		return "bm-spr-pvc-1550-8"
	default:
		return ""
	}
}

func MapAccessModeFromTrainingToStorage(accessMode v1.StorageAccessModeType) v1.FilesystemAccessModes {
	switch accessMode {
	case v1.StorageAccessModeType_STORAGE_READ_WRITE:
		return v1.FilesystemAccessModes_ReadWrite
	case v1.StorageAccessModeType_STORAGE_READ_ONLY:
		return v1.FilesystemAccessModes_ReadOnly
	case v1.StorageAccessModeType_STORAGE_READ_WRITE_ONCE:
		return v1.FilesystemAccessModes_ReadWriteOnce
	default:
		// Default to Read Write Storage
		return v1.FilesystemAccessModes_ReadWrite
	}
}

func MapMountProtocolFromTrainingToStorage(mountProtocol v1.StorageMountType) v1.FilesystemMountProtocols {
	switch mountProtocol {
	case v1.StorageMountType_STORAGE_WEKA:
		return v1.FilesystemMountProtocols_Weka
	default:
		// Default to Weka Storage
		return v1.FilesystemMountProtocols_Weka
	}
}
