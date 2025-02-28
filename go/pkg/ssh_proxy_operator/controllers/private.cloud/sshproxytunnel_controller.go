// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package privatecloud

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	idcclientset "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcscheme "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned/scheme"
	idcinformer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions/private.cloud/v1alpha1"
	idclister "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/listers/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type SshProxyController struct {
	// clientset for SshProxyTunnel custom resource
	SshProxyClientSet idcclientset.Interface
	// SshProxyTunnel cache has synced
	SshProxySynced cache.InformerSynced
	// lister : will be used to get and list objects avoiding calling api server directly
	SshProxyLister idclister.SshProxyTunnelLister
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	Workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder
	// configValues defines the config values passed to the the controller
	ConfigValues SshProxyTunnelConfig
	// mockScpTargets are scp targets used to mock failure recovery test
	MockScpTargets map[string]error
	// mockScpTargetsMutex is the mutex used for avoiding race condition
	MockScpTargetsMutex *sync.Mutex
}

type SshProxyTunnelConfig struct {
	AuthorizedKeysFilePath   string
	ProxyUser                string
	ProxyAddress             string
	ProxyPort                int
	AuthorizedKeysScpTargets []string
	PrivateKey               string
	PublicKey                string
	HostPublicKey            string
}

const SshAuthorizedKeysOptions = "command=\"exit\",no-user-rc,no-agent-forwarding,no-pty,no-X11-forwarding"

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=sshproxytunnels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=sshproxytunnels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;get;list;update;patch;delete

func NewSshProxyController(ctx context.Context, kubeClientSet kubernetes.Interface, sshProxyClientSet idcclientset.Interface,
	sshProxyInformer idcinformer.SshProxyTunnelInformer, sshProxyConfig SshProxyTunnelConfig) (*SshProxyController, error) {
	log := log.FromContext(ctx).WithName("NewSshProxyController")

	// Add sshproxy-controller types to the default Kubernetes Scheme so Events can be logged for sshproxy-controller types.
	utilruntime.Must(idcscheme.AddToScheme(scheme.Scheme))

	// Create event broadcaster
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "sshproxy-controller"})

	sshProxyController := &SshProxyController{
		SshProxyClientSet:   sshProxyClientSet,
		SshProxySynced:      sshProxyInformer.Informer().HasSynced,
		SshProxyLister:      sshProxyInformer.Lister(),
		Workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SshProxy"),
		Recorder:            recorder,
		ConfigValues:        sshProxyConfig,
		MockScpTargets:      make(map[string]error),
		MockScpTargetsMutex: new(sync.Mutex),
	}

	// Registering Event Handlers to the informer
	sshProxyInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: sshProxyController.handleObject,
			UpdateFunc: func(old, new interface{}) {
				newSshProxyTunnel := new.(*cloudv1alpha1.SshProxyTunnel)
				oldSshProxyTunnel := old.(*cloudv1alpha1.SshProxyTunnel)

				// To avoid triggering of update each time status is updated
				if newSshProxyTunnel.Generation == oldSshProxyTunnel.Generation {
					// Periodic resync will send update events for all known SshProxyTunnel Objects.
					// Two different versions of the same SshProxyTunnel Object will always have different Generation.
					return
				}
				sshProxyController.handleObject(new)
			},
			DeleteFunc: sshProxyController.handleObject,
		},
	)

	// Create directories with the permissions required by sshd
	if err := sshProxyController.createDirectories(ctx); err != nil {
		log.Error(err, "Failed to create directory")
		return &SshProxyController{}, err
	}

	return sshProxyController, nil
}

func (c *SshProxyController) handleObject(obj interface{}) {
	c.Workqueue.Add(obj)
}

// Run will set up the event handlers for types we are interested in, as well as syncing informer caches and starting workers.
// It will block until stopCh is closed, at which point it will shutdown the workqueue and wait for workers to finish
// processing their current work items.
func (c *SshProxyController) Run(ctx context.Context, stopCh <-chan struct{}) error {
	log := log.FromContext(ctx).WithName("SshProxyController.Run").
		WithValues(logkeys.Controller, "sshProxy", logkeys.ControllerGroup, "private.cloud.intel.com", logkeys.ControllerKind, "SshProxyTunnel")

	defer utilruntime.HandleCrash()

	// Let the workers stop when we are done
	defer c.Workqueue.ShutDown()

	log.Info("Starting Controller")

	// Wait for all involved caches to be synced, before processing items from the queue is started
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.SshProxySynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")

	// Launch the worker to process the SSHProxyTunnel resources
	go wait.Until(c.worker, time.Second, stopCh)

	<-stopCh

	log.Info("All workers finished")

	return nil
}

func (c *SshProxyController) worker() {
	var object struct{}
	ctx := context.Background()

	// Start span here.
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SshProxyController.worker").Start()
	defer span.End()

	err := func() error {
		scpTargetsCount := len(c.ConfigValues.AuthorizedKeysScpTargets)

		// Declaring atomic variable
		var scpErrorCount int32
		// Get the first item from the queue
		obj, shutdown := c.Workqueue.Get()
		if shutdown {
			return fmt.Errorf("failed to get the first item from the queue")
		}

		// We have at least one SshProxyTunnel to reconcile.
		c.Workqueue.Forget(obj)
		c.Workqueue.Done(obj)

		// Getting the remaining items from the queue
		currentNumberOfItemsInQueue := c.Workqueue.Len()
		for i := 0; i < currentNumberOfItemsInQueue; i++ {
			obj, shutdown := c.Workqueue.Get()
			if shutdown {
				return fmt.Errorf("failed to get the item from the queue")
			}
			c.Workqueue.Forget(obj)
			c.Workqueue.Done(obj)
		}

		// Fetch all the SshProxyTunnel Instances from all the namespaces from cache using Lister
		sshProxyTunnelList, err := c.SshProxyLister.SshProxyTunnels("").List(labels.NewSelector())
		if err != nil {
			err = fmt.Errorf("error reading the list of SshProxyTunnel Objects: %w", err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		err = c.writeAuthorizedKeysFile(ctx, sshProxyTunnelList)
		if err != nil {
			return fmt.Errorf("failed to update authorized_keys file %s: %w", c.ConfigValues.AuthorizedKeysFilePath, err)
		}

		g := new(errgroup.Group)
		for _, authorizedKeysScpTarget := range c.ConfigValues.AuthorizedKeysScpTargets {
			authorizedKeysScpTarget := authorizedKeysScpTarget
			g.Go(func() error {
				err := c.scpAuthorizedKeysFile(ctx, authorizedKeysScpTarget)
				if err != nil {
					atomic.AddInt32(&scpErrorCount, 1)
					log.Error(err, "failed to scp the authorized_keys file to the proxy server", logkeys.Target, authorizedKeysScpTarget)
				}
				return nil
			})
		}

		// Any error is logged in goroutines which increment scpErrorCount.
		_ = g.Wait()

		// if scp to all the proxy servers in `authorizedKeysScpTargets` is failed
		if int(scpErrorCount) == scpTargetsCount {
			err = fmt.Errorf("all SCP targets failed")
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		// Updating the status for all the objects
		for _, sshProxyTunnel := range sshProxyTunnelList {
			state, err := c.updateSshProxyTunnelStatus(ctx, sshProxyTunnel)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("Cannot update the status of SshProxyTunnel object. Ignoring since tunnel object must be deleted", logkeys.Namespace, sshProxyTunnel.Namespace, logkeys.Name, sshProxyTunnel.Name)
				} else {
					err = fmt.Errorf("error occured while updating the status of SshProxyTunnel object %s/%s: %w", sshProxyTunnel.Namespace, sshProxyTunnel.Name, err)
					span.SetStatus(codes.Error, err.Error())
					return err
				}
			} else if err == nil && state == "created" {
				c.Recorder.Event(sshProxyTunnel, corev1.EventTypeNormal, "CreationSuccessful", "SshProxyTunnel instance updated successfully")
			}
		}

		// Retry scp if there is any failure
		if int(scpErrorCount) > 0 {
			c.Workqueue.Add(object)
		}
		return nil
	}()

	if err != nil {
		log.Error(err, "worker encountered an error")
		c.Workqueue.Add(object)
	}
}

func (c *SshProxyController) updateSshProxyTunnelStatus(ctx context.Context, sshProxyTunnel *cloudv1alpha1.SshProxyTunnel) (string, error) {
	// NOTE: Never modify objects from the store(informer cache). It's a read-only, local cache.
	// Use DeepCopy() to make a deep copy of original object and modify this copy Or create a copy manually for better performance
	sshProxyTunnelCopy := sshProxyTunnel.DeepCopy()

	if !(sshProxyTunnelCopy.Status.ProxyUser == c.ConfigValues.ProxyUser && sshProxyTunnelCopy.Status.ProxyAddress == c.ConfigValues.ProxyAddress && sshProxyTunnelCopy.Status.ProxyPort == c.ConfigValues.ProxyPort) {
		ctx, _, span := obs.LogAndSpanFromContextOrGlobal(ctx).
			WithName("SshProxyController.updateSshProxyTunnelStatus").
			WithValues(logkeys.ResourceId, sshProxyTunnel.ObjectMeta.Name).Start()
		defer span.End()
		sshProxyTunnelCopy.Status.ProxyUser = c.ConfigValues.ProxyUser
		sshProxyTunnelCopy.Status.ProxyAddress = c.ConfigValues.ProxyAddress
		sshProxyTunnelCopy.Status.ProxyPort = c.ConfigValues.ProxyPort
		// NOTE: If the CustomResourceSubresources feature gate is not enabled, we must use Update instead of UpdateStatus to update
		// the Status block of the resource. UpdateStatus will not allow changes to the Spec of the resource,
		// which is ideal for ensuring nothing other than resource status has been updated.
		_, errUpdate := c.SshProxyClientSet.PrivateV1alpha1().SshProxyTunnels(sshProxyTunnel.Namespace).UpdateStatus(ctx, sshProxyTunnelCopy, metav1.UpdateOptions{})
		if errUpdate != nil {
			span.SetStatus(codes.Error, errUpdate.Error())
		}
		return "created", errUpdate
	}
	return "", nil
}

func (c *SshProxyController) scpAuthorizedKeysFile(ctx context.Context, authorizedKeysScpTarget string) error {
	log := log.FromContext(ctx).WithName("SshProxyController.scpAuthorizedKeysFile")

	// This block of code is for mocking failure recovery test
	c.MockScpTargetsMutex.Lock()
	mockErr, isMock := c.MockScpTargets[authorizedKeysScpTarget]
	c.MockScpTargetsMutex.Unlock()
	if isMock {
		return mockErr
	}

	targetUrl, err := url.Parse(authorizedKeysScpTarget)
	if err != nil {
		return err
	}
	if targetUrl.Scheme != "scp" {
		return fmt.Errorf("unsupported scheme for %s", targetUrl)
	}
	log.Info("SCP Authorized Keys File", logkeys.TargetUrl, targetUrl, logkeys.TargetUrlUserName, targetUrl.User.Username(), logkeys.TargetUrl, targetUrl.Host, logkeys.TargetUrlPath, targetUrl.Path)

	// Parse private key from string
	parsedPrivateKey, err := ssh.ParsePrivateKey([]byte(c.ConfigValues.PrivateKey))
	if err != nil {
		return err
	}

	hostPublicKey, err := util.GetHostPublicKey(c.ConfigValues.HostPublicKey)
	if err != nil {
		return err
	}

	clientConfig := ssh.ClientConfig{
		User: targetUrl.User.Username(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(parsedPrivateKey),
		},
		HostKeyAlgorithms: util.GetSupportedHostKeyAlgorithms(),
		HostKeyCallback:   ssh.FixedHostKey(hostPublicKey),
		Timeout:           5 * time.Second,
	}

	// Create a new SCP client
	client := scp.NewClient(targetUrl.Host, &clientConfig)
	defer client.Close()

	// Connect to the remote server
	if err := client.Connect(); err != nil {
		return fmt.Errorf("couldn't establish a connection to the remote server: %w", err)
	}

	f, err := os.Open(c.ConfigValues.AuthorizedKeysFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Copy the file to the remote server. Make sure the path exists at the remote server
	// For Example - If we are copying the file /tmp/.ssh/authorized_keys.tmp
	remoteTempFileName := targetUrl.Path + ".tmp"

	if err := client.CopyFromFile(ctx, *f, remoteTempFileName, "0600"); err != nil {
		return fmt.Errorf("error while copying file: %w", err)
	}

	newclient, err := ssh.Dial("tcp", targetUrl.Host, &clientConfig)
	if err != nil {
		return fmt.Errorf("error while creating the new connection to the remote server: %w", err)
	}
	defer newclient.Close()

	// Create a new session. (A session is a remote execution of a program).
	session, err := newclient.NewSession()
	if err != nil {
		return fmt.Errorf("error while open a new Session for the new client: %w", err)
	}
	defer session.Close()

	// Executing remote ssh command in the session
	cmd := fmt.Sprintf("mv %s %s", remoteTempFileName, targetUrl.Path)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("error while running the command on the remote server: %w", err)
	}

	log.Info("scpAuthorizedKeysFile completed successfully", logkeys.TargetUrl, targetUrl)
	return nil
}

func (c *SshProxyController) writeAuthorizedKeysFile(ctx context.Context, sshtlist []*cloudv1alpha1.SshProxyTunnel) error {

	specMap, err := c.getSpecMap(ctx, sshtlist)
	if err != nil {
		return err
	}
	if err := c.writeAuthorizedKeysFileFromSpecMap(ctx, specMap); err != nil {
		return fmt.Errorf("generateNewAuthorizedKeysFile failed: %w", err)
	}
	return nil
}

func (c *SshProxyController) getSpecMap(ctx context.Context, sshtlist []*cloudv1alpha1.SshProxyTunnel) (map[string][]cloudv1alpha1.SshProxyTunnelSpec, error) {
	log := log.FromContext(ctx).WithName("SshProxyController.getSpecMap")
	specMap := make(map[string][]cloudv1alpha1.SshProxyTunnelSpec)
	for _, ssht := range sshtlist {
		for _, sshPublicKey := range ssht.Spec.SshPublicKeys {
			publicKey, err := CleanPublicKey(ctx, sshPublicKey)
			if err == nil {
				specMap[publicKey] = append(specMap[publicKey], ssht.Spec)
			} else {
				log.Error(err, "Ignoring invalid SSH public keys", logkeys.PublicKeys, ssht.Spec.SshPublicKeys)
			}
		}
	}
	log.Info("Spec map statistics", logkeys.UniquePublicKeyCount, len(specMap), logkeys.ResourceCount, len(sshtlist))

	return specMap, nil
}

// Remove the last field from a sshkey public key.
// Input SSH Public key: "ssh-rsa AADD#$R%%@$$== testuser@intel.com"
// Expected sshkey: "ssh-rsa AADD#$R%%@$$=="
func CleanPublicKey(ctx context.Context, publicKey string) (string, error) {
	if publicKey == "" {
		// nothing to clean
		return publicKey, nil
	}

	sshPublicKeyFields := strings.Fields(publicKey)
	if len(sshPublicKeyFields) < 2 {
		return "", fmt.Errorf("invalid public key '%s'", publicKey)
	}
	return fmt.Sprintf("%s %s", sshPublicKeyFields[0], sshPublicKeyFields[1]), nil
}

func (c *SshProxyController) writeAuthorizedKeysFileFromSpecMap(
	ctx context.Context, allSpecsForAnSSHKey map[string][]cloudv1alpha1.SshProxyTunnelSpec) (err error) {

	authorizedKeysContent, err := c.getAuthorizedKeysContentsFromSpecMap(ctx, allSpecsForAnSSHKey)
	if err != nil {
		return err
	}

	err = c.writeTextFileAtomically(ctx, c.ConfigValues.AuthorizedKeysFilePath, authorizedKeysContent)
	if err != nil {
		return err
	}

	return nil
}

func (c *SshProxyController) getAuthorizedKeysContentsFromSpecMap(
	ctx context.Context, specMap map[string][]cloudv1alpha1.SshProxyTunnelSpec) (string, error) {

	// Sort SSH public keys so that the file is deterministic. This makes testing and troubleshooting easier.
	sshPublicKeys := make([]string, 0)
	for sshPublicKey := range specMap {
		sshPublicKeys = append(sshPublicKeys, sshPublicKey)
	}
	sort.Strings(sshPublicKeys)

	var b strings.Builder

	// Adding the content of id_rsa.pub of the operator to the authorized_keys file in order to scp/ssh to the proxy server via operator
	_, err := b.WriteString(c.ConfigValues.PublicKey)
	if err != nil {
		return "", err
	}

	for _, sshPublicKey := range sshPublicKeys {
		sshProxyTunnelSpecs := specMap[sshPublicKey]
		ips := make([]string, 0)

		for _, sshProxyTunnelSpec := range sshProxyTunnelSpecs {
			for _, targetAddress := range sshProxyTunnelSpec.TargetAddresses {
				for _, targetPort := range sshProxyTunnelSpec.TargetPorts {
					ip := fmt.Sprintf("permitopen=\"%s:%d\"", targetAddress, targetPort)
					ips = append(ips, ip)
				}
			}
		}
		sort.Strings(ips)

		ipsFormatted := strings.Join(ips, ",")
		line := fmt.Sprintf("%s,%s %s\n", ipsFormatted, SshAuthorizedKeysOptions, sshPublicKey)
		_, err := b.WriteString(line)
		if err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

func (c *SshProxyController) writeTextFileAtomically(ctx context.Context, filePath string, contents string) error {
	tmpFilePath := fmt.Sprintf("%s.tmp", filePath)
	f, err := os.OpenFile(tmpFilePath, os.O_WRONLY+os.O_CREATE+os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if _, err = w.WriteString(contents); err != nil {
		f.Close()
		return err
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}
	return nil
}

// Create directories with the permissions required by sshd.
func (c *SshProxyController) createDirectories(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("SshProxyController.createDirectories")

	sshPath := filepath.Dir(c.ConfigValues.AuthorizedKeysFilePath)
	homePath := filepath.Dir(sshPath)
	log.Info("createDirectories", logkeys.SshPath, sshPath, logkeys.HomePath, homePath, logkeys.KeyPath, c.ConfigValues.AuthorizedKeysFilePath)

	if err := os.MkdirAll(homePath, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		return err
	}
	return nil
}
