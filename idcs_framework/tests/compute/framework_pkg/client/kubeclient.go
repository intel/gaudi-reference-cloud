package client

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset
var namespace string
var kubeConfigPath string

// var dynamicClient *kubernetes.Clientset

func SetUpKubeClient(kubeconfig string, testEnv string) {
	namespace = "idcs-system"
	if testEnv == "kind" {
		//config := ctrl.GetConfigOrDie()
		//clientset, _ = kubernetes.NewForConfig(config)
		config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
		clientset, _ = kubernetes.NewForConfig(config)
	} else {
		config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
		clientset, _ = kubernetes.NewForConfig(config)
	}

	kubeConfigPath = "--kubeconfig=" + kubeconfig
}

func GetConfigMap(configMapName string) {
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		logInstance.Println("Retrieval of configmap failed : ", zap.Error(err))
	}
	logInstance.Println("Config map retrival response is : ", zap.Any("", configMap))
}

// GetNodes retrieves and returns a list of node names in the cluster
func GetNodes() ([]string, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logInstance.Println("Failed to retrieve nodes: ", zap.Error(err))
		return nil, err
	}

	var nodeNames []string
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
		//logger.Log.Info("Node: ", zap.String("name", node.Name))
	}

	return nodeNames, nil
}

// GetPods retrieves and returns a list of pod names in the specified namespace
func GetPods() ([]string, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logInstance.Println("Failed to retrieve pods: ", zap.Error(err))
		return nil, err
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
		//logger.Log.Info("Pod: ", zap.String("name", pod.Name))
	}

	return podNames, nil
}

// GetPodsInNodeWithNumericNamespace retrieves and returns a list of pod names
// running on the specified node and in namespaces that contain only digits
func GetPodsInNode(nodeName string) ([]string, error) {
	// Get all pods across all namespaces on the specified node
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		logInstance.Println("Failed to retrieve pods: ", zap.Error(err))
		return nil, err
	}

	var podNames []string
	numericNamespaceRegex := regexp.MustCompile(`^\d+$`) // Regex to match namespaces with only digits

	for _, pod := range pods.Items {
		if numericNamespaceRegex.MatchString(pod.Namespace) {
			podNames = append(podNames, pod.Name)
			logInstance.Println("Pod found", zap.String("name", pod.Name), zap.String("namespace", pod.Namespace))
		}
	}

	if len(podNames) == 0 {
		logInstance.Println("No pods found in namespaces with only digits")
	}

	return podNames, nil
}

func GetNodeNameFromPod(podPattern, namespace string) (string, error) {
	// List all pods in the specified namespace
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		//logger.Log.Error("Failed to list pods: ", zap.Error(err))
		return "", err
	}

	// Iterate through the pod list to find a pod that matches the pattern
	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, podPattern) {
			// Found a matching pod, now get its node name
			nodeName := pod.Spec.NodeName
			if nodeName == "" {
				//logger.Log.Info("Pod is not assigned to any node yet", zap.String("pod", pod.Name))
				return "", nil
			}

			//logger.Log.Info("Node name retrieved from pod", zap.String("pod", pod.Name), zap.String("node", nodeName))
			return nodeName, nil
		}
	}

	// If no matching pod was found
	logInstance.Println("No pod matching the pattern found", zap.String("pattern", podPattern))
	return "", fmt.Errorf("no pod matching the pattern '%s' found in namespace '%s'", podPattern, namespace)
}

func ExecuteKubectlCommand(command string) (string, error) {
	cmd := exec.Command("kubectl", strings.Split(command, " ")...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func GetInstanceCustomResource(instanceName string, instanceNamespace string, kubeFilePath string) (*cloudv1alpha1.Instance, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeFilePath)
	if err != nil {
		return nil, err
	}

	kubeclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	resource := cloudv1alpha1.SchemeBuilder.GroupVersion.WithResource("instances")
	instance := &cloudv1alpha1.Instance{}

	host, err := kubeclient.Resource(resource).Namespace(instanceNamespace).Get(context.Background(), instanceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable find instance %s", instanceName)
	} else {
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(host.UnstructuredContent(), instance); err != nil {
			return nil, fmt.Errorf("Unable to decode instance object")
		} else {
			return instance, nil
		}
	}
}

// TODO: convert this into generic function (specific to metering)
func UpdateMeteringConfigMap(configMapName string, key string) bool {
	configMap, getErr := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if getErr != nil {
		logInstance.Println("Updation of configmap failed : ", zap.Error(getErr))
	}
	matcher := regexp.MustCompile(`maxUsageRecordSendInterval(.*)`)
	updatedValue := matcher.ReplaceAllString(configMap.Data[key], `maxUsageRecordSendInterval: "2m"`)
	configMap.Data[key] = updatedValue
	updated, updateErr := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if updateErr == nil {
		logInstance.Println("Config map updation response is : ", zap.Any("", updated))
		return true
	} else {
		logInstance.Println("Updation failed")
		return false
	}
}

func RevertMeteringConfigMap(configMapName string, key string) bool {
	configMap, getErr := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if getErr != nil {
		logInstance.Println("Updation of configmap failed : ", zap.Error(getErr))
	}
	matcher := regexp.MustCompile(`maxUsageRecordSendInterval(.*)`)
	updatedValue := matcher.ReplaceAllString(configMap.Data[key], `maxUsageRecordSendInterval: "60m"`)
	configMap.Data[key] = updatedValue
	updated, updateErr := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if updateErr == nil {
		logInstance.Println("Config map updation response is : ", zap.Any("", updated))
		return true
	} else {
		logInstance.Println("Updation failed")
		return false
	}
}

// TODO: convert this into generic function (specific to quota)
func UpdateQuotaConfigMap(configMapName string, key string) bool {
	configMap, getErr := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if getErr != nil {
		logInstance.Println("Updation of configmap failed : ", zap.Error(getErr))
	}

	matcher := regexp.MustCompile(`cloudAccountQuota(.*)`)
	updatedValue := matcher.ReplaceAllString(configMap.Data[key], `cloudAccountQuota: {"cloudAccounts":{"ENTERPRISE":{"instanceQuota":{"bm-icp-gaudi2":5,"bm-spr":10,"bm-spr-gaudi2":0,"bm-spr-pl":10,"bm-spr-pvc-1100-4":10,"bm-spr-pvc-1100-8":5,"bm-spr-pvc-1550-8":0,"bm-virtual":5,"vm-icp-gaudi2-1":8,"vm-spr-lrg":50,"vm-spr-med":50,"vm-spr-sml":50,"vm-spr-tny":5},"storageQuota":{"filesystems":2,"totalSizeGB":1000}},"ENTERPRISE_PENDING":{"instanceQuota":{"bm-icp-gaudi2":0,"bm-spr":1,"bm-spr-gaudi2":0,"bm-spr-pl":1,"bm-spr-pvc-1100-4":1,"bm-spr-pvc-1550-8":0,"bm-virtual":5,"vm-icp-gaudi2-1":0,"vm-spr-lrg":0,"vm-spr-med":0,"vm-spr-sml":2,"vm-spr-tny":5}},"INTEL":{"instanceQuota":{"bm-icp-gaudi2":2,"bm-spr":2,"bm-spr-gaudi2":2,"bm-spr-pl":2,"bm-spr-pvc-1100-4":2,"bm-spr-pvc-1100-8":2,"bm-spr-pvc-1550-8":2,"bm-virtual":1,"vm-icp-gaudi2-1":2,"vm-spr-lrg":2,"vm-spr-med":2,"vm-spr-sml":2,"vm-spr-tny":2},"storageQuota":{"filesystems":2,"totalSizeGB":1000}},"PREMIUM":{"instanceQuota":{"bm-icp-gaudi2":2,"bm-spr":2,"bm-spr-gaudi2":2,"bm-spr-pl":2,"bm-spr-pvc-1100-4":2,"bm-spr-pvc-1100-8":2,"bm-spr-pvc-1550-8":2,"bm-virtual":1,"vm-icp-gaudi2-1":2,"vm-spr-lrg":2,"vm-spr-med":2,"vm-spr-sml":2,"vm-spr-tny":2},"storageQuota":{"filesystems":2,"totalSizeGB":1000}},"STANDARD":{"instanceQuota":{"bm-icp-gaudi2":1,"bm-spr":1,"bm-spr-pl":1,"bm-spr-pvc-1100-4":1,"bm-virtual":1,"vm-icp-gaudi2-1":1,"vm-spr-lrg":1,"vm-spr-med":1,"vm-spr-sml":1,"vm-spr-tny":1},"storageQuota":{"filesystems":1,"totalSizeGB":100}}}}`)
	configMap.Data[key] = updatedValue
	updated, updateErr := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if updateErr == nil {
		logInstance.Println("Config map updation response is : ", zap.Any("", updated))
		return true
	} else {
		logInstance.Println("Updation failed")
		return false
	}
}

func RevertQuotaConfigMap(configMapName string, key string) bool {
	configMap, getErr := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if getErr != nil {
		logInstance.Println("Updation of configmap failed : ", zap.Error(getErr))
	}

	matcher := regexp.MustCompile(`cloudAccountQuota(.*)`)
	updatedValue := matcher.ReplaceAllString(configMap.Data[key], `cloudAccountQuota: {"cloudAccounts":{"ENTERPRISE":{"instanceQuota":{"bm-icp-gaudi2":5,"bm-spr":10,"bm-spr-gaudi2":0,"bm-spr-pl":10,"bm-spr-pvc-1100-4":10,"bm-spr-pvc-1100-8":5,"bm-spr-pvc-1550-8":0,"bm-virtual":5,"vm-icp-gaudi2-1":8,"vm-spr-lrg":50,"vm-spr-med":50,"vm-spr-sml":50,"vm-spr-tny":5},"storageQuota":{"filesystems":2,"totalSizeGB":1000}},"ENTERPRISE_PENDING":{"instanceQuota":{"bm-icp-gaudi2":0,"bm-spr":1,"bm-spr-gaudi2":0,"bm-spr-pl":1,"bm-spr-pvc-1100-4":1,"bm-spr-pvc-1550-8":0,"bm-virtual":5,"vm-icp-gaudi2-1":0,"vm-spr-lrg":0,"vm-spr-med":0,"vm-spr-sml":2,"vm-spr-tny":5}},"INTEL":{"instanceQuota":{"bm-icp-gaudi2":2,"bm-spr":2,"bm-spr-gaudi2":2,"bm-spr-pl":2,"bm-spr-pvc-1100-4":2,"bm-spr-pvc-1100-8":2,"bm-spr-pvc-1550-8":2,"bm-virtual":20,"vm-icp-gaudi2-1":5,"vm-spr-lrg":4,"vm-spr-med":5,"vm-spr-sml":10,"vm-spr-tny":20},"storageQuota":{"filesystems":2,"totalSizeGB":1000}},"PREMIUM":{"instanceQuota":{"bm-icp-gaudi2":2,"bm-spr":2,"bm-spr-gaudi2":2,"bm-spr-pl":2,"bm-spr-pvc-1100-4":2,"bm-spr-pvc-1100-8":2,"bm-spr-pvc-1550-8":2,"bm-virtual":10,"vm-icp-gaudi2-1":2,"vm-spr-lrg":3,"vm-spr-med":5,"vm-spr-sml":5,"vm-spr-tny":10},"storageQuota":{"filesystems":2,"totalSizeGB":1000}},"STANDARD":{"instanceQuota":{"bm-icp-gaudi2":1,"bm-spr":1,"bm-spr-pl":1,"bm-spr-pvc-1100-4":1,"bm-virtual":5,"vm-icp-gaudi2-1":1,"vm-spr-lrg":2,"vm-spr-med":3,"vm-spr-sml":4,"vm-spr-tny":8},"storageQuota":{"filesystems":1,"totalSizeGB":100}}}}`)
	configMap.Data[key] = updatedValue
	updated, updateErr := clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if updateErr == nil {
		logInstance.Println("Config map updation response is : ", zap.Any("", updated))
		return true
	} else {
		logInstance.Println("Updation failed")
		return false
	}
}

func RestartPod(podNameRegex string) bool {
	podList, _ := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	var podName string
	for _, eachPod := range podList.Items {
		if strings.Contains(eachPod.Name, podNameRegex) {
			podName = eachPod.Name
		}
	}
	logInstance.Println("Pod name is : ", zap.String("", podName))
	podDeletion := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	logInstance.Println("Config map updation response is : ", zap.Any("", podDeletion))

	time.Sleep(10 * time.Second)
	// validate after deletion
	podListAfterDelete, _ := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	for _, eachPod := range podListAfterDelete.Items {
		if strings.Contains(eachPod.Name, podNameRegex) {
			logInstance.Println("Pod is restarted successfully")
			return true
		}
	}
	return false
}
