// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"errors"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sClient) ListNamespaces() []corev1.Namespace {
	var allNamespaces []corev1.Namespace
	namespaces, _ := k.ClientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	allNamespaces = append(allNamespaces, namespaces.Items...)
	return allNamespaces
}

func (k *K8sClient) isNamespaceExists(name string) (exists bool, namespace *corev1.Namespace) {

	namespaces := k.ListNamespaces()
	for _, ns := range namespaces {
		if ns.Name == name {
			log.Printf("The name matches: %s == %s", name, ns.Name)
			return true, &ns
		}
	}
	return false, nil
}

func (k *K8sClient) CreateNamespace(name string, ignoreIfExists bool, labels map[string]string) (*corev1.Namespace, error) {

	exists, ns := k.isNamespaceExists(name)
	if exists && !ignoreIfExists {
		return ns, errors.New("namespace already exists")
	} else if exists && ignoreIfExists {
		log.Printf("The namespace %s already exists. Skipping createNamespace", name)
		return ns, nil
	} else {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}

		namespace, err := k.ClientSet.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})

		if err != nil {
			log.Printf("Namespace creation failed with the error: %+v", err)
			return nil, err
		}
		log.Printf("Created the namespace. %+v", namespace)

		if len(labels) != 0 {
			// Update the labels
			if namespace.ObjectMeta.Labels == nil {
				namespace.ObjectMeta.Labels = make(map[string]string)
			}
			for key, value := range labels {
				namespace.ObjectMeta.Labels[key] = value
			}

			// Apply the update
			namespace, err = k.ClientSet.CoreV1().Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("error updating namespace labels: %+v", err)
			}

			fmt.Printf("Namespace %s updated successfully with labels %v\n", namespace.Name, labels)

		}

		return namespace, nil
	}

}

func (k *K8sClient) DeleteNamespace(name string, ignoreIfNotExists bool) (bool, error) {

	exists, _ := k.isNamespaceExists(name)
	if !exists && !ignoreIfNotExists {
		return false, errors.New("namespace doesnot exists")
	} else if !exists && ignoreIfNotExists {
		log.Println("The namespace doesnot exists. Skipping DeleteNamespace")
		return true, nil
	} else {
		err := k.ClientSet.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Namespace deletion failed with the error: %+v", err)
			return false, err
		}
		log.Printf("Deleted the namespace. %s", name)

		return true, nil
	}

}
