// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (k *K8sClient) MakeDefaultStorageClass(storageClassName string) error {
	patchData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
	}

	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return fmt.Errorf("error marshalling patch data: %+v", err)
	}

	// Patch the StorageClass
	patchedStorageClass, err := k.ClientSet.StorageV1().StorageClasses().Patch(context.TODO(),
		storageClassName,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("error patching StorageClass: %+v", err)
	}

	fmt.Printf("Patched StorageClass: %v as the default storageClass\n", patchedStorageClass.Name)

	return nil
}
