// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"context"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	sql "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/dns"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"

	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/rest"
)

type Network struct {
	/* An Unique Identifier used to create DNS with other unique aspects as FQDN to expose the service */
	ServiceType string // used for generaic DNS Record creation
	/* This is needs to be the namespace the platform service is deployed for Istio to do service discovery upstream across nodegroups */
	ServiceNamespace string

	// an unique ID, which identifies the service deployment.
	ServiceId string

	/*  pass the components of data platform need to be exposed
	For example, for airflow pass {flower : 5555, webserver: 8080 } which are individual k8s clusterIP services and clusterIP port
	*/
	ExposedK8sServiceNames map[string]int

	CloudAccountID  string
	WorkspaceId     string
	IksClusterUUID  string
	K8sClientConfig *rest.Config
	SqlModel        *sql.Queries
	DnsService      dns.IMenMice
	K8sClient       *k8s.K8sClient
	ServiceConfig   *config.Config
	IstioClientSet  *k8s.IstioClientSet
}

func (n *Network) InitDnsService(ctx context.Context) {
	dnsService := dns.NewMenMiceService(n.ServiceConfig.Dns.DnsIpamApiEndpoint)
	dnsService.SetApiToken()
	n.DnsService = dnsService
}

func (n *Network) InitIstioClient(ctx context.Context, clientConfig *rest.Config) {

	log.Println("Initializing the Istio ClientSet with Iks Tenant Cluster clientConfig ", clientConfig)

	istio, err := versionedclient.NewForConfig(clientConfig)
	if err != nil {
		log.Fatalf("Error getting Istio client set: %+v", err)
	}

	log.Println("Configuring the Istio ClientSet")

	n.IstioClientSet = &k8s.IstioClientSet{
		ClientSet: istio,
		Ctx:       context.Background(),
	}
}

// Please use this to verify network errors enforce type errors.New() if required
// TODO: Need to add more custom enforced, errors
func IsNetworkError(err error) bool {
	if err != nil {
		return true
	}
	return false
}

// cache the result across different delete task to prevent requeiring database
type NetworkRecordsOutput struct {
	GatewayIstioName             string
	GatewayIstioLabelName        string
	GatewayIstioSecretName       string
	DnsCnameRecord               string
	LBHwDnsARecord               string
	WorkspaceLoadBalancerARecord string
}
