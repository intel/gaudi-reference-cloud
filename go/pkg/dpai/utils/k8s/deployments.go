// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *K8sClient) ListDeployments(namespace string) ([]appsv1.Deployment, error) {
	var allDeployments []appsv1.Deployment
	client := k.ClientSet.AppsV1().Deployments(namespace)
	deployments, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	allDeployments = append(allDeployments, deployments.Items...)
	return allDeployments, nil
}

func (k *K8sClient) waitForRollingRestart(namespace, deploymentName string, timeout time.Duration) error {

	startTime := time.Now()
	deadline := startTime.Add(timeout)

	for time.Now().Before(deadline) {
		deployment, err := k.ClientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas && deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
			// All replicas have been updated and are available, rolling restart is complete
			return nil
		}

		// Wait for a short interval before checking again
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timeout: Rolling restart did not complete within %s", timeout)

}

func (k *K8sClient) RestartDeployment(namespace string, name string, timeout time.Duration) error {

	client := k.ClientSet.AppsV1().Deployments(namespace)
	deployment, err := client.Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		log.Printf("Not able to get the deployment: %s in the namespace: %s", name, namespace)
		return err
	}

	// Trigger a rolling restart by updating the annotations
	deployment.Annotations["dapi.idcservice.net/restartedAt"] = time.Now().Format(time.RFC3339)
	_, err = client.Update(context.TODO(), deployment, metav1.UpdateOptions{})

	if err != nil {
		log.Printf("Error performing rolling restart: %s\n", err)
		return err
	}

	log.Printf("Rolling restart initiated for Deployment %s in Namespace %s\n", name, namespace)
	if err = k.waitForRollingRestart(namespace, name, timeout); err != nil {
		return err
	}
	return nil
}

func (k *K8sClient) RestartAllDeployments(namespace string, timeoutPerDeployment time.Duration) error {

	deployments, err := k.ListDeployments(namespace)

	if err != nil {
		return err
	}

	for _, d := range deployments {
		if err := k.RestartDeployment(namespace, d.Name, timeoutPerDeployment); err != nil {
			return err
		}
	}
	return nil
}
