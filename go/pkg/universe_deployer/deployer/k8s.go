// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/manifests_generator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"github.com/phayes/freeport"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *Deployer) InitializeK8sClients(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("InitializeK8sClients")
	if !e.K8sApiEnabled {
		return nil
	}
	if e.InitializeK8sClientsComplete {
		// Already initialized.
		return nil
	}
	kubeContexts, err := e.GetKubeContexts()
	if err != nil {
		return err
	}
	log.Info("Loading KubeConfig for Kubernetes clusters", "kubeContexts", kubeContexts)
	restConfigs, err := k8s.LoadKubeConfigContexts(ctx, kubeContexts)
	if err != nil {
		return err
	}
	k8sClients := map[string]k8sclient.Client{}
	for kubeContext, restConfig := range restConfigs {
		k8sClient, err := k8sclient.New(restConfig, k8sclient.Options{})
		if err != nil {
			return err
		}
		k8sClients[kubeContext] = k8sClient
	}
	e.RestConfigs = restConfigs
	e.K8sClients = k8sClients
	log.Info("Testing connection to Kubernetes clusters", "kubeContexts", kubeContexts)
	if _, err := e.ListDeployments(ctx); err != nil {
		return err
	}
	log.Info("Connected to Kubernetes clusters", "kubeContexts", kubeContexts)
	e.InitializeK8sClientsComplete = true
	return nil
}

// Return all distinct KubeContexts in EnvConfig.
func (e *Deployer) GetKubeContexts() ([]string, error) {
	m := map[string]bool{}
	m[e.Options.EnvConfig.Values.Global.KubeContext] = true
	for _, region := range e.Options.EnvConfig.Values.Regions {
		m[region.KubeContext] = true
		for _, availabilityZone := range region.AvailabilityZones {
			m[availabilityZone.KubeContext] = true
			m[availabilityZone.NetworkCluster.KubeContext] = true
			m[availabilityZone.QuickConnect.KubeContext] = true
		}
	}
	var kubeContexts []string
	for kubeContext := range m {
		if kubeContext != "" {
			kubeContexts = append(kubeContexts, kubeContext)
		}
	}
	return kubeContexts, nil
}

func (e *Deployer) GetGlobalK8sRestConfig() (*restclient.Config, error) {
	restConfig, ok := e.RestConfigs[e.Options.EnvConfig.Values.Global.KubeContext]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return restConfig, nil
}

func (e *Deployer) StartPortForward(ctx context.Context, namespace string, service string, remotePort string) (int, error) {
	log := log.FromContext(ctx).WithName("StartPortForward")
	localPort, err := freeport.GetFreePort()
	if err != nil {
		return 0, err
	}
	e.BackgroundProcessesWaitGroup.Add(1)
	go func() {
		defer e.BackgroundProcessesWaitGroup.Done()
		// Keep retrying port forward until context is cancelled.
		for {
			cmd := exec.CommandContext(
				e.BackgroundProcessesContext,
				e.KubectlBinary,
				"port-forward",
				"--namespace", namespace,
				service,
				fmt.Sprintf("%d:%s", localPort, remotePort),
			)
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "HOME="+e.HomeDir)
			cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeConfig)
			err := util.RunCmd(ctx, cmd)
			if e.BackgroundProcessesContext.Err() == context.Canceled {
				log.V(2).Info("kubectl port-forward canceled")
				return
			}
			if err != nil {
				log.Error(err, "kubectl port-forward")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return localPort, nil
}

// List deployments in all Kubernetes clusters.
func (e *Deployer) ListDeployments(ctx context.Context) (*appsv1.DeploymentList, error) {
	all := appsv1.DeploymentList{}
	for kubeContext, k8sClient := range e.K8sClients {
		items := appsv1.DeploymentList{}
		if err := k8sClient.List(ctx, &items, &k8sclient.ListOptions{}); err != nil {
			return nil, fmt.Errorf("%s: %w", kubeContext, err)
		}
		all.Items = append(all.Items, items.Items...)
	}
	return &all, nil
}

// List daemonsets in all Kubernetes clusters.
func (e *Deployer) ListDaemonSets(ctx context.Context) (*appsv1.DaemonSetList, error) {
	all := appsv1.DaemonSetList{}
	for kubeContext, k8sClient := range e.K8sClients {
		items := appsv1.DaemonSetList{}
		if err := k8sClient.List(ctx, &items, &k8sclient.ListOptions{}); err != nil {
			return nil, fmt.Errorf("%s: %w", kubeContext, err)
		}
		all.Items = append(all.Items, items.Items...)
	}
	return &all, nil
}

// List jobs in all Kubernetes clusters.
func (e *Deployer) ListJobs(ctx context.Context) (*batchv1.JobList, error) {
	all := batchv1.JobList{}
	for kubeContext, k8sClient := range e.K8sClients {
		items := batchv1.JobList{}
		if err := k8sClient.List(ctx, &items, &k8sclient.ListOptions{}); err != nil {
			return nil, fmt.Errorf("%s: %w", kubeContext, err)
		}
		all.Items = append(all.Items, items.Items...)
	}
	return &all, nil
}

type ManifestStatus struct {
	manifests_generator.Manifest
	Healthy    bool
	Reason     error
	DaemonSet  *appsv1.DaemonSet
	Deployment *appsv1.Deployment
	Job        *batchv1.Job
}

// For each manifest produced by manifests generator, try to find a matching deployment or job in any Kubernetes cluser,
// then get the deployment or job health.
func (e *Deployer) GetManifestsStatus(ctx context.Context) ([]ManifestStatus, error) {
	deploymentList, err := e.ListDeployments(ctx)
	if err != nil {
		return nil, err
	}
	daemonSetList, err := e.ListDaemonSets(ctx)
	if err != nil {
		return nil, err
	}

	jobList, err := e.ListJobs(ctx)
	if err != nil {
		return nil, err
	}
	var manifestsStatus []ManifestStatus
	for _, manifest := range e.Manifests.Manifests {
		found := false
		manifestStatus := ManifestStatus{
			Manifest: manifest,
			Healthy:  false,
		}
		// Check if this manifest should be assumed to be healthy.
		for _, nn := range e.Options.EnvConfig.Values.UniverseDeployer.IgnoreHealthCheckFor {
			if nn.Namespace == manifest.Namespace() {
				re, err := regexp.Compile(nn.Name)
				if err != nil {
					return nil, err
				}
				if re.MatchString(manifest.ReleaseName()) {
					// Do not check health.
					manifestStatus.Healthy = true
					break
				}
			}
		}
		// Try to find a matching Deployment.
		if !manifestStatus.Healthy && !found {
			for _, deployment := range deploymentList.Items {
				if isMatchingInstance(manifest, deployment.ObjectMeta) {
					if err := ensureMatchingCommit(manifest, deployment.ObjectMeta); err != nil {
						manifestStatus.Reason = err
					} else {
						condFound := true
						for _, cond := range deployment.Status.Conditions {
							if cond.Type == appsv1.DeploymentAvailable {
								if cond.Status == corev1.ConditionTrue {
									manifestStatus.Healthy = true
								} else {
									manifestStatus.Reason = fmt.Errorf("deployment not available: %s", cond.Reason)
								}
								condFound = true
								break
							}
						}
						if !condFound {
							manifestStatus.Reason = fmt.Errorf("deployment condition not found")
						}
					}
					manifestStatus.Deployment = &deployment
					found = true
					break
				}
			}
		}
		// Try to find a matching DaemonSet.
		if !manifestStatus.Healthy && !found {
			for _, daemonSet := range daemonSetList.Items {
				if isMatchingInstance(manifest, daemonSet.ObjectMeta) {
					if err := ensureMatchingCommit(manifest, daemonSet.ObjectMeta); err != nil {
						manifestStatus.Reason = err
					} else {
						desiredCount := daemonSet.Status.DesiredNumberScheduled
						readyCount := daemonSet.Status.NumberReady
						if desiredCount == readyCount {
							manifestStatus.Healthy = true
						} else {
							manifestStatus.Reason = fmt.Errorf("daemonSet not ready: %d/%d replicas are healthy", readyCount, desiredCount)
						}
					}
					manifestStatus.DaemonSet = &daemonSet
					found = true
					break
				}
			}
		}
		// Try to find a matching Job.
		if !manifestStatus.Healthy && !found {
			for _, job := range jobList.Items {
				if isMatchingInstance(manifest, job.ObjectMeta) {
					if err := ensureMatchingCommit(manifest, job.ObjectMeta); err != nil {
						manifestStatus.Reason = err
					} else {
						condFound := false
						for _, cond := range job.Status.Conditions {
							if cond.Type == batchv1.JobComplete {
								if cond.Status == corev1.ConditionTrue {
									manifestStatus.Healthy = true
								} else {
									manifestStatus.Reason = fmt.Errorf("job not complete: %s", cond.Reason)
								}
								condFound = true
								break
							}
						}
						if !condFound {
							manifestStatus.Reason = fmt.Errorf("job condition not found")
						}
					}
					manifestStatus.Job = &job
					found = true
					break
				}
			}
		}
		if !found {
			manifestStatus.Reason = fmt.Errorf("not found")
		}
		manifestsStatus = append(manifestsStatus, manifestStatus)
	}
	return manifestsStatus, nil
}

func isMatchingInstance(manifest manifests_generator.Manifest, objectMeta metav1.ObjectMeta) bool {
	expectedInstanceValue := fmt.Sprintf("%s-%s", manifest.KubeContext, manifest.ReleaseName())
	foundInstanceValue := objectMeta.Labels["app.kubernetes.io/instance"]
	return foundInstanceValue == expectedInstanceValue
}

func ensureMatchingCommit(manifest manifests_generator.Manifest, objectMeta metav1.ObjectMeta) error {
	foundVersion := objectMeta.Labels["app.kubernetes.io/version"]
	if !util.IsGitCommit(foundVersion) {
		// foundVersion is not a Git commit. This is not a standard IDC chart.
		// This function cannot ensure a matching commit so assume it matches.
		return nil
	}
	if foundVersion != manifest.GitCommit {
		return fmt.Errorf("resource has Git commit %s but expected %s", foundVersion, manifest.GitCommit)
	}

	foundConfigCommit := objectMeta.Labels[util.ConfigCommitLabel]
	if !util.IsGitCommit(foundConfigCommit) {
		// foundConfigCommit is not a Git commit. This is legacy IDC chart.
		// This function cannot ensure a matching commit so assume it matches.
		return nil
	}
	if foundConfigCommit != manifest.ConfigCommit {
		return fmt.Errorf("resource has config commit %s but expected %s", foundConfigCommit, manifest.ConfigCommit)
	}
	return nil
}

func (e *Deployer) WaitForK8sResources(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("WaitForK8sResources")

	logHelpMessage := "At least one application (Helm release) was deployed but is not yet healthy.\n" +
		"If the Helm release does not define a Deployment nor a Job, then add it to " +
		"universeDeployer.ignoreHealthCheckFor in deployment/helmfile/defaults.yaml.gotmpl."
	var manifestsStatus []ManifestStatus
	firstHelpMessage := 3 * time.Minute
	timeLastHelpMessage := time.Now().Add(firstHelpMessage)
	timeout := 15 * time.Minute
	backoff := LinearBackoff(timeout, 5*time.Second)
	if err := retry.OnError(backoff, func(error) bool { return true }, func() error {
		var err error
		manifestsStatus, err = e.GetManifestsStatus(ctx)
		if err != nil {
			return err
		}
		unhealthyCount := 0
		for _, manifest := range manifestsStatus {
			if !manifest.Healthy {
				unhealthyCount += 1
				log.Info("Unhealthy application", "name", manifest.ReleaseName(), "namespace", manifest.Namespace(), "reason", manifest.Reason)
			}
		}
		totalCount := len(manifestsStatus)
		if unhealthyCount > 0 {
			log.Info("Waiting for applications to become healthy", "unhealthyCount", unhealthyCount, "totalCount", totalCount)

			// Periodically, collect logs and show a troubleshooting help message.
			if time.Since(timeLastHelpMessage) > 1*time.Minute {
				if err := e.CollectK8sLogs(ctx); err != nil {
					log.Error(err, "unable to collect logs")
				}
				log.Error(nil, logHelpMessage+
					"\nThis will continue to wait for a total of "+fmt.Sprintf("%v", timeout)+".")
				timeLastHelpMessage = time.Now()
			}

			return fmt.Errorf("waiting for %d of %d applications to become healthy", unhealthyCount, totalCount)
		}
		return nil
	}); err != nil {
		log.Error(nil, logHelpMessage)
		return err
	}

	log.Info("All applications are healthy", "totalCount", len(manifestsStatus))
	return nil
}

func (e *Deployer) DeployK8sTlsSecrets(ctx context.Context) error {
	// Create wildcard-tls secret for ingress.
	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "hack/deploy-k8s-tls-secrets.sh"))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeConfig)
	cmd.Env = append(cmd.Env, "KUBECTL="+e.KubectlBinary)
	cmd.Env = append(cmd.Env, "SECRETS_DIR="+e.SecretsDir)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}
	return nil
}

func (e *Deployer) CollectK8sLogs(ctx context.Context) error {
	if e.Options.BuildArtifactsDir == "" {
		// Log collection is disabled.
		return nil
	}
	kubeContexts, err := e.GetKubeContexts()
	if err != nil {
		return err
	}
	for _, kubeContext := range kubeContexts {
		cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "go/pkg/universe_deployer/deployer/collect_k8s_logs.sh"))
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeConfig)
		cmd.Env = append(cmd.Env, "KUBECONTEXT="+kubeContext)
		cmd.Env = append(cmd.Env, "KUBECTL="+e.KubectlBinary)
		cmd.Env = append(cmd.Env, "LOGDIR="+filepath.Join(e.Options.BuildArtifactsDir, "k8s-logs", kubeContext))
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}
