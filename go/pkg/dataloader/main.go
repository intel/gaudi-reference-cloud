package main

import (
	"context"
	"fmt"
	"os"
	"time"

	metal3ClientSet "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/clientset/versioned"
	metal3Informerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/informers/externalversions"
	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/listers/metal3.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig *string
var clusterClient *kubernetes.Clientset
var metal3Lister v1alpha1.BareMetalHostLister
var selector labels.Selector

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("kubeconfigpdx"))
	selector = labels.Everything()

	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}
	clusterClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}
	metal3client, err := metal3ClientSet.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Metal3 client: %v\n", err)
		os.Exit(1)
	}

	metal3InformerFactory := metal3Informerfactory.NewSharedInformerFactory(metal3client, time.Second*30)
	metal3Informer := metal3InformerFactory.Metal3().V1alpha1().BareMetalHosts().Informer()
	metal3Lister = metal3InformerFactory.Metal3().V1alpha1().BareMetalHosts().Lister()

	metal3Informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    handleBMHAdd,
		DeleteFunc: handleBMHDelete,
		UpdateFunc: handleBMHUpdate,
	})
}

func handleBMHAdd(obj interface{}) {
	logger := log.FromContext(context.Background()).WithName("BMHController.handleBMHAdd")
	logger.Info("BMH Add event observed")
	listBmhs(clusterClient, metal3Lister, selector)
}

func handleBMHDelete(obj interface{}) {
	logger := log.FromContext(context.Background()).WithName("BMHController.handleBMHDelete")
	logger.Info("BMH Delete event observed")
	listBmhs(clusterClient, metal3Lister, selector)
}

func handleBMHUpdate(obj interface{}, newObj interface{}) {
	logger := log.FromContext(context.Background()).WithName("BMHController.handleBMHUpdate")
	logger.Info("BMH Update event observed")
	listBmhs(clusterClient, metal3Lister, selector)
}

func listBmhs(clusterClient *kubernetes.Clientset, metal3Lister v1alpha1.BareMetalHostLister, selector labels.Selector) {
	nsList, err := clusterClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing namespaces: %v\n", err)
	}

	for _, n := range nsList.Items {
		fmt.Printf("Namespace: %s\n", n.Name)

		bmhInstances, err := metal3Lister.BareMetalHosts(n.Name).List(selector)

		if err != nil {
			fmt.Printf("Error listing BMH instances: %v\n", err)
			os.Exit(1)
		}

		for _, instance := range bmhInstances {
			fmt.Printf("BMH Instance Name: %s\n", instance.Name)
		}
	}
}
