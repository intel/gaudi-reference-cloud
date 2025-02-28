// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/hashicorp/go-getter/helper/url"
	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"golang.org/x/crypto/ssh"
)

type TaskStatus string

// Various states of the validator task
const (
	NOT_STARTED TaskStatus = "NotStarted"
	IN_PROGRESS TaskStatus = "InProgress"
	SUCCESS     TaskStatus = "Success"
	FAILED      TaskStatus = "Failed"
)

type TaskMeta struct {
	instanceType   string
	isCluster      bool
	clusterGroupId string
	region         string
	az             string
	bmhNamespace   string
	bmhName        string
	// This represents the validation task artifact name. e.g: validation.tar.gz or validation-0.0.1.tar.gz.
	// The version is extracted from the cloudv1alpha1.BmInstanceOperatorConfig
	clusterArtifact, instanceArtifact string
	// This represents the test configuration specified by the BMH label cloud.intel.com/validation-test-configuration
	testConfig string
}

// Common constants used in the module.
const (
	DEFAULT_SSH_PORT = 22
	CACHE_BASE_PATH  = "/tmp/"
	BASE_PATH        = "/tmp/validation/"
	DEFAULT_ARTIFACT = "/validation.tar.gz"
)

// Validator helps to trigger a validation on the specified instance by establishing a ssh connection to the instance via
// the jump hosts. The validator create the underlying connection once and reuses it for all its operations.
type Validator struct {
	ip               string
	signer           *ssh.Signer
	proxy            cloudv1alpha1.SshProxyTunnelStatus
	userName         string
	bastionClient    *ssh.Client
	instanceClient   *ssh.Client
	taskMeta         *TaskMeta // This will be used to determine the type of validation task to be executed.
	repo_url         string    // This is the repository from where the tests can be downloaded.
	s3BucketName     string    // S3 bucket name where the validation reports are stored
	s3key            string    // S3 access key
	s3Secret         string    // S3 secret
	huggingFaceToken string    // Huggingface access token
	httpsProxy       string    // proxy to be used by the validation script
	memberIPs        *[]string
	memberNames      *[]string // BMH names of all the members
}

/*
Repository structure:
	- Every folder, corresponding to an instance type, has a validation.tar.gz that contains the test to be executed.

validation_repository/
└── catalog/
    ├── bm-icp-gaudi2/
    │   └── validation.tar.gz
    |   └── cluster
    |         └── validation.tar.gz
    ├── bm-spr/
    └── bm-spr-pvc-1100-4/
*/

// Helper method to create a Validator
func CreateValidator(ctx context.Context, instance *cloudv1alpha1.Instance, bmh *baremetalv1alpha1.BareMetalHost,
	signer *ssh.Signer, cfg *cloudv1alpha1.BmInstanceOperatorConfig) (*Validator, error) {
	if len(instance.Status.Interfaces) == 0 || len(instance.Status.Interfaces[0].Addresses) == 0 {
		return nil, fmt.Errorf("createValidator: Instance does not have an IP address; instance.Status.Interfaces=%v",
			instance.Status.Interfaces)
	}

	taskMeta := getTaskMeta(instance.Spec.InstanceType, cfg, bmh)
	// read s3 access key
	accessKeyBytes, err := os.ReadFile(cfg.ValidationReportS3Config.S3AccessKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read s3Access key file %v", err)
	}

	// read s3 secret access key
	secretAccessKeyBytes, err := os.ReadFile(cfg.ValidationReportS3Config.S3SecretAccessKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read s3 Secret Access key file %v", err)
	}

	huggingFaceTokenBytes, err := os.ReadFile(cfg.EnvConfiguration.HuggingFaceTokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Huggingface Token key file %v", err)
	}

	v := &Validator{
		ip:               instance.Status.Interfaces[0].Addresses[0],
		signer:           signer,
		proxy:            instance.Status.SshProxy,
		userName:         instance.Status.UserName,
		taskMeta:         taskMeta,
		repo_url:         cfg.ValidationTaskRepositoryURL,
		s3BucketName:     cfg.ValidationReportS3Config.BucketName,
		s3key:            strings.TrimSuffix(string(accessKeyBytes), "\n"),
		s3Secret:         strings.TrimSuffix(string(secretAccessKeyBytes), "\n"),
		httpsProxy:       cfg.ValidationReportS3Config.HttpsProxy,
		huggingFaceToken: strings.TrimSuffix(string(huggingFaceTokenBytes), "\n"),
	}
	// initialize the ssh connections to the jump host and instances
	if err := v.init(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

func CreateValidatorForGroup(ctx context.Context, instance *cloudv1alpha1.Instance, bmh *baremetalv1alpha1.BareMetalHost,
	signer *ssh.Signer, cfg *cloudv1alpha1.BmInstanceOperatorConfig, memberIPs, memberNames *[]string) (*Validator, error) {
	validator, err := CreateValidator(ctx, instance, bmh, signer, cfg)
	if err == nil {
		validator.memberIPs = memberIPs
		validator.memberNames = memberNames
	}
	return validator, err
}

func getTaskMeta(instanceType string, cfg *cloudv1alpha1.BmInstanceOperatorConfig,
	bmh *baremetalv1alpha1.BareMetalHost) *TaskMeta {
	taskMeta := &TaskMeta{
		instanceType: instanceType,
		region:       cfg.EnvConfiguration.Region,
		az:           cfg.EnvConfiguration.AvailabilityZone,
		bmhNamespace: bmh.Namespace,
		bmhName:      bmh.Name,
	}
	clusterGroupId := GetLabel(bmenroll.ClusterGroupID, bmh)
	// Mark an instance as cluster only if feature flag is enabled
	if clusterGroupId != "" && cfg.FeatureFlags.GroupValidation && exists(cfg.FeatureFlags.EnabledGroupInstanceTypes, instanceType) {
		taskMeta.isCluster = true
		taskMeta.clusterGroupId = clusterGroupId
	} else {
		taskMeta.isCluster = false //not a cluster
	}
	taskMeta.clusterArtifact = getClusterArtifact(instanceType, cfg)
	taskMeta.instanceArtifact = getInstanceArtifact(instanceType, cfg)
	taskMeta.testConfig = GetLabel(bmenroll.TestConfigurationLabel, bmh)
	return taskMeta
}

func getClusterArtifact(instanceType string, cfg *cloudv1alpha1.BmInstanceOperatorConfig) string {
	artifact := DEFAULT_ARTIFACT
	if cfg.ValidationTaskVersion.ClusterVersionMap != nil && cfg.ValidationTaskVersion.ClusterVersionMap[instanceType] != "" {
		artifact = "/validation-" + cfg.ValidationTaskVersion.ClusterVersionMap[instanceType] + ".tar.gz"
	}
	return artifact
}

func getInstanceArtifact(instanceType string, cfg *cloudv1alpha1.BmInstanceOperatorConfig) string {
	artifact := DEFAULT_ARTIFACT
	if cfg.ValidationTaskVersion.InstanceVersionMap != nil || cfg.ValidationTaskVersion.InstanceVersionMap[instanceType] != "" {
		artifact = "/validation-" + cfg.ValidationTaskVersion.InstanceVersionMap[instanceType] + ".tar.gz"
	}
	return artifact
}

func (meta *TaskMeta) getTaskRepoPath(isCluster bool) string {
	if isCluster {
		return meta.instanceType + "/cluster"
	} else {
		return meta.instanceType
	}
}

func (meta *TaskMeta) getArtifact(isCluster bool) string {
	if isCluster {
		return meta.clusterArtifact
	} else {
		return meta.instanceArtifact
	}
}

// path where the report/logs need to be uploaded
func (meta *TaskMeta) getReportPath(isCluster bool) string {
	t := time.Now()
	if isCluster {
		return path.Join("/", meta.region, meta.az, meta.bmhNamespace, meta.bmhName+t.Format("-20060102150405")+"-group")
	} else {
		return path.Join("/", meta.region, meta.az, meta.bmhNamespace, meta.bmhName+t.Format("-20060102150405"))
	}
}

// Initialize the validator by creating connections to bastion and instances.
func (v *Validator) init(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Validator.init")
	log.Info("Initializing validator, Downloading validation artifact from S3 repository if no cached copy is available",
		logkeys.ValidatorIp, v.ip, logkeys.TaskMeta, v.taskMeta)
	// Parse the private key
	bastionhost := formatHost(v.proxy.ProxyAddress, v.proxy.ProxyPort)
	destinationHost := formatHost(v.ip, DEFAULT_SSH_PORT)

	bastionConfig := getSSHConfig(v.proxy.ProxyUser, v.signer)
	// Connect to the bastion host
	bClient, err := ssh.Dial("tcp", bastionhost, bastionConfig)
	if err != nil {
		return fmt.Errorf("failed to dial to jump host: %w", err)
	}
	v.bastionClient = bClient

	// Dial a connection to the instance, from the bastion
	conn, err := bClient.Dial("tcp", destinationHost)
	if err != nil {
		return fmt.Errorf("failed to connect to service host: %w", err)
	}

	destinationConfig := getSSHConfig(v.userName, v.signer)
	// Establish an ssh connection to the instance using the underlying connection.
	ncc, chans, reqs, err := ssh.NewClientConn(conn, destinationHost, destinationConfig)
	if err != nil {
		return fmt.Errorf("failed to establish ssh connection to service host: %w", err)
	}
	log.V(9).Info("Established ssh connection with instance.", logkeys.DestinationHostIp, destinationHost)
	sClient := ssh.NewClient(ncc, chans, reqs)
	v.instanceClient = sClient
	// Download both the instance and cluster validation artifacts incase of cluster type
	if v.taskMeta.isCluster {
		if err := v.downloadFile(ctx, v.buildRepoPath(true), true); err != nil {
			log.Error(err, "failed to download and cache the cluster validation archive", logkeys.InstanceType,
				v.taskMeta.instanceType)
			return NonRetryableError(fmt.Sprintf("failed to cache the cluster validation archive for instanceType %s",
				v.taskMeta.instanceType))
		}
	}
	if err := v.downloadFile(ctx, v.buildRepoPath(false), false); err != nil {
		log.Error(err, "failed to download and cache the instance validation archive", logkeys.InstanceType,
			v.taskMeta.instanceType)
		return NonRetryableError(fmt.Sprintf("failed to cache the instance validation archive for instanceType %s",
			v.taskMeta.instanceType))
	}
	// no errors.
	return nil
}

// close all open connections.
func (v *Validator) Close(ctx context.Context) {
	log := log.FromContext(ctx).WithName("Validator.Close")
	if v.instanceClient != nil {
		if err := v.instanceClient.Close(); err != nil {
			log.Error(err, "Failed to close instanceClient")
		}
	}
	if v.bastionClient != nil {
		if err := v.bastionClient.Close(); err != nil {
			log.Error(err, "Failed to close bastionClient")
		}
	}
}

func (v *Validator) startGroupValidationTask(ctx context.Context, desiredFwVersion *Version) (bool, error) {
	log := log.FromContext(ctx).WithName("Validator.startGroupValidationTask")
	s := *v.memberIPs
	names := *v.memberNames
	env := &map[string]string{
		"MASTER_IP":    v.ip,
		"MASTER_NAME":  v.taskMeta.bmhName,
		"MEMBER_IPS":   strings.Join(s[:], ","),
		"MEMBER_NAMES": strings.Join(names[:], ","),
	}
	log.Info("Starting group validation task", logkeys.EnvDetails, env)
	// Invoke the standard validation task.
	return v.startValidationTask(ctx, env, true, desiredFwVersion)
}

func (v *Validator) startInstanceValidationTask(ctx context.Context, desiredFwVersion *Version) (bool, error) {
	return v.startValidationTask(ctx, nil, false, desiredFwVersion)
}

// Start a validation task, ensure it is idempotent. If the task is already started return true.
func (v *Validator) startValidationTask(ctx context.Context, customEnv *map[string]string, isCluster bool, desiredFwVersion *Version) (bool, error) {
	log := log.FromContext(ctx).WithName("Validator.startValidationTask")
	_, err := v.executeCmd("mkdir -p " + BASE_PATH)
	if err != nil {
		return false, err
	}
	currentStatus, _, err := v.getValidationTaskStatus(ctx)
	if err != nil {
		return false, err
	}
	if currentStatus != NOT_STARTED {
		taskType := "Instance Validation"
		if isCluster {
			taskType = "Group Validation"
		}
		log.Info("Validation task has already started", logkeys.Task, taskType)
		return true, nil //do not start the task again
	}
	path, err := v.copyWrapperFile(ctx)
	if err != nil {
		return false, err
	}
	_, err = v.copyValidationArchive(ctx, isCluster)
	if err != nil {
		return false, err
	}
	env := map[string]string{
		"uploadPath":         v.taskMeta.getReportPath(isCluster),
		"s3Key":              v.s3key,
		"s3Secret":           v.s3Secret,
		"huggingFaceToken":   v.huggingFaceToken,
		"bucket":             v.s3BucketName,
		"https_proxy":        v.httpsProxy,
		"bmhName":            v.taskMeta.bmhName, // bmhName on which validation is running will be logged
		"instanceType":       v.taskMeta.instanceType,
		"region":             v.taskMeta.region,
		"az":                 v.taskMeta.az,
		"TEST_CONFIGURATION": v.taskMeta.testConfig,
		"BUILD_VERSION":      desiredFwVersion.BuildVersion,
		"SPI_VERSION":        desiredFwVersion.SpiVersion,
		"FULL_FW_VERSION":    desiredFwVersion.FullFwVersion,
	}
	if v.taskMeta.isCluster {
		env["clusterGroupId"] = v.taskMeta.clusterGroupId
	}
	// append custom env
	if customEnv != nil {
		for k, v := range *customEnv {
			env[k] = v
		}
	}
	result, err := v.executeCmdWithEnv("nohup "+path+" >/tmp/validation/validation.out 2>&1 &", env)
	if err != nil {
		return false, err
	}
	log.Info("Validation task has started", logkeys.Artifact,
		v.taskMeta.getTaskRepoPath(isCluster)+v.taskMeta.getArtifact(isCluster))
	log.V(9).Info("Start validation command output", logkeys.ValidationResult, result)
	return false, nil
}

func (v *Validator) buildRepoPath(isCluster bool) string {
	repo_path := v.repo_url + "/validation_repository/catalog/" +
		v.taskMeta.getTaskRepoPath(isCluster) + v.taskMeta.getArtifact(isCluster)
	return repo_path
}

func (v *Validator) downloadFile(ctx context.Context, src string, isCluster bool) error {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Validator.downloadfile").Start()
	defer span.End()
	dst := CACHE_BASE_PATH + v.taskMeta.getTaskRepoPath(isCluster) + "/"
	u, err := url.Parse(src)
	if err != nil {
		return err
	}
	file := path.Base(u.Path)
	targetFilePath := filepath.Join(dst, file)

	// Check if the file has been already downloaded.
	_, err = os.Stat(targetFilePath)

	//TODO: Improve logic based on checksum
	if !os.IsNotExist(err) {
		// File already exists.
		return nil
	}

	// Check if repo file exists.
	err = httpFileExists(src)
	if err != nil {
		return fmt.Errorf("validation archive does not exist at %s; Details: %w", src, err)
	}

	// Create a cache directory
	err = os.MkdirAll(dst, os.FileMode(0700))
	if err != nil {
		return err
	}
	// Fetch the file.
	resp, err := http.Get(src)
	if err != nil {
		return fmt.Errorf("error while downloading validation archive %w", err)
	}

	// cleanup
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err, "Error closing connection to remote data source")
		}
	}()

	// Create a cache file.
	cacheFile, err := os.Create(targetFilePath)
	if err != nil {
		return fmt.Errorf("failed to create a cache file at %s; Details %w", targetFilePath, err)
	}
	// cleanup
	defer func(localFile *os.File) {
		err := localFile.Close()
		if err != nil {
			log.Error(err, "Error closing file")
		}
	}(cacheFile)

	//Read response and save it to file
	_, err = cacheFile.ReadFrom(resp.Body)
	return err
}

// Check if Validation task has completed execution.
func (v *Validator) isValidationTaskCompleted(ctx context.Context) (bool, TaskStatus, error) {
	log := log.FromContext(ctx).WithName("Validator.isValidationTaskCompleted")
	currentStatus, _, err := v.getValidationTaskStatus(ctx)
	if err != nil {
		return false, "", err
	}
	if currentStatus == NOT_STARTED {
		// This is unexpected, a system reboot can be one of the causes.
		log.Error(nil, "Validation task is not found, please check if system rebooted during validation")
		// marking the validation as failed as there is we are not able to access the pid file for the validation task.
		return true, FAILED, nil
	} else if currentStatus == SUCCESS || currentStatus == FAILED {
		return true, currentStatus, nil
	} else {
		return false, currentStatus, nil
	}
}

func (v *Validator) clearTestData(ctx context.Context) {
	log := log.FromContext(ctx).WithName("Validator.clearTestData")
	t := time.Now()
	cmd := "mv -f " + BASE_PATH + " /tmp/validationbkup-" + t.Format("20060102150405")

	_, err := v.executeCmd(cmd)
	if err != nil {
		log.Info("Err observed while attempting to backup the validation folder, ignoring it")
	}
}

// Fetch the current Status of the Task and additional details
// for e.g: this can return the error details of the task
func (v *Validator) getValidationTaskStatus(ctx context.Context) (TaskStatus, string, error) {
	log := log.FromContext(ctx).WithName("Validator.getValidationTaskStatus")
	result, err := v.executeCmd("[[ ! -f /tmp/validation/validation.pid ]] && echo \"NotStarted\" || ps -o pid= -q `cat /tmp/validation/validation.pid`")
	if err != nil {
		log.V(9).Info("Err returned since the pid does not exist, indicating the at the task has completed")
	}
	if result == "" {
		log.Info("Task has completed, check the exit code ")
		exitCode, err := v.getValidationTaskExitCode(ctx)
		if err != nil {
			return "", "", err
		}
		if exitCode == "0" {
			return SUCCESS, "validation task completed successfully", nil
		} else {
			return FAILED, "validation task failed check /tmp/validation/validation.out file for logs", nil
		}
	} else if result == string(NOT_STARTED) {
		log.Info("Task has never been started")
		return NOT_STARTED, "validation has not started", nil
	} else {
		log.Info("Task is in progress")
		return IN_PROGRESS, "validation completed successfully", nil
	}
}

func (v *Validator) getValidationTaskExitCode(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("Validator.getValidationTaskExitCode")
	result, err := v.executeCmd("cat /tmp/validation/validation_exitcode")

	if err != nil {
		return "", err
	}
	exitCode := string(result)
	log.Info("Validation task exit code ", logkeys.ValidationTaskExitCode, exitCode)
	return exitCode, nil
}

func (v *Validator) getValidationResultMeta(ctx context.Context) (map[string]string, error) {
	log := log.FromContext(ctx).WithName("Validator.getValidationResultMeta")
	result, err := v.executeCmd("cat /tmp/validation_result.meta")

	if err != nil {
		return nil, err
	}
	resultMeta := convertToMap(string(result))
	log.Info("Validation result meta", logkeys.ValidationResultEntrycount, len(resultMeta))
	return resultMeta, nil
}

// Execute command.
func (v *Validator) executeCmd(cmd string) (string, error) {
	session, err := v.instanceClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to establish session %w", err)
	}
	defer session.Close()
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("error while executing command %s Details: %w", cmd, err)
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}

func (v *Validator) executeCmdWithEnv(cmd string, env map[string]string) (string, error) {
	if env != nil {
		envString := ""
		for name, value := range env {
			envString = envString + name + "=" + value + " "
		}
		completeCmd := envString + cmd
		fmt.Printf("Command being executed %v", completeCmd)
		return v.executeCmd(completeCmd)
	} else {
		return v.executeCmd(cmd)
	}
}

func (v *Validator) copyWrapperFile(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("Validator.copyWrapperFile")
	scpClient, err := scp.NewClientBySSH(v.instanceClient)
	file_path := "/tmp/validation/wrapper_script.sh"
	if err != nil {
		return "", fmt.Errorf("scp: failed to create a new session using existing connection %w", err)
	}

	f, err := os.Open("wrapper_script.sh")
	if err != nil {
		return "", fmt.Errorf("failed to open file %w", err)
	}
	defer f.Close()
	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = scpClient.CopyFromFile(ctx, *f, file_path, "0777")
	if err != nil {
		return "", fmt.Errorf("failed to copy file %w", err)
	}
	log.Info("Completed copy of Validation wrapper file to BMH", logkeys.Artifact, f.Name())
	return file_path, nil
}

func (v *Validator) copyValidationArchive(ctx context.Context, isCluster bool) (string, error) {
	log := log.FromContext(ctx).WithName("Validator.copyValidationArchive")
	scpClient, err := scp.NewClientBySSH(v.instanceClient)
	dst_path := BASE_PATH + DEFAULT_ARTIFACT
	src_path := CACHE_BASE_PATH + v.taskMeta.getTaskRepoPath(isCluster) + v.taskMeta.getArtifact(isCluster)
	if err != nil {
		return "", fmt.Errorf("scp: failed to create a new session using existing connection %w", err)
	}

	f, err := os.Open(src_path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %w", err)
	}
	defer f.Close()
	// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
	err = scpClient.CopyFromFile(ctx, *f, dst_path, "0666")
	if err != nil {
		return "", fmt.Errorf("failed to copy file %w", err)
	}
	log.Info("Completed copy of Validation archive to BMH", logkeys.Artifact, src_path)
	return dst_path, nil
}

// Helper methods
func formatHost(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func getSSHConfig(user string, signer *ssh.Signer) *ssh.ClientConfig {
	// Set up the SSH config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(*signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	return config
}

// Check if the file exists without having to download the large file.
func httpFileExists(path string) error {
	head, err := http.Head(path)
	if err != nil {
		return err
	}
	statusOK := head.StatusCode >= 200 && head.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("%s does not exists: head request get http code %d", path, head.StatusCode)
	}
	defer func() {
		_ = head.Body.Close()
	}()
	return err
}
