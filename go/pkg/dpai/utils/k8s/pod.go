// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (k *K8sClient) ListPods(namespace string) ([]corev1.Pod, error) {
	var allPods []corev1.Pod
	namespaces, err := k.ClientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing pods: %v", err)
	}
	allPods = append(allPods, namespaces.Items...)
	return allPods, nil
}

func (k *K8sClient) isPodExists(namespace, name string) (exists bool, pod *corev1.Pod) {

	pods, err := k.ListPods(namespace)
	if err != nil {
		log.Printf("error listing pods: %v", err)
		return false, nil
	}
	for _, pod := range pods {
		if pod.Name == name {
			log.Printf("The name matches: %s == %s", name, pod.Name)
			return true, &pod
		}
	}
	return false, nil
}

func (k *K8sClient) ExecInPod(namespace string, name string, command []string) error {
	exists, pod := k.isPodExists(namespace, name)
	if !exists {
		return fmt.Errorf("pod %s does not exist in namespace %s", name, namespace)
	}

	// Create exec request
	req := k.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec")

	// Set query parameters
	req.VersionedParams(&corev1.PodExecOptions{
		Container: pod.Spec.Containers[0].Name,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	// Create executor
	exec, err := remotecommand.NewSPDYExecutor(k.ClientConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating the executor: %v", err)
	}

	// Set up buffers for output
	var stdout, stderr bytes.Buffer

	// Execute command
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return fmt.Errorf("error: %v\n Stderr: %v", err, stderr.String())
	}

	fmt.Printf("Output: %v\n", stdout.String())

	return nil
}
