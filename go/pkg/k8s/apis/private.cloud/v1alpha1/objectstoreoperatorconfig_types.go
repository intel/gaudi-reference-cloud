package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// ObjectStoreOperatorConfig is the Schema for the Object Storeoperatorconfigs API.
// It stores the configuration for the Obj store Operator.
type ObjectStoreOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`
	// Controller addr returns the adrress for weka controllers
	StorageControllerServerAddr string `json:"storageControllerServerAddr"`
	// Flag to indicate if mtls is use or not
	StorageControllerServerUseMtls bool `json:"storageControllerServerUseMtls"`
}

func init() {
	SchemeBuilder.Register(&ObjectStoreOperatorConfig{})
}
