// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/sethvargo/go-password/password"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	K8S_AWS_ID_HEADER = "x-k8s-aws-id"
	TOKEN_PREFIX      = "k8s-aws-v1."

	KIND        = "ExecCredential"
	API_VERSION = "client.authentication.k8s.io/v1beta1"

	EXPIRE_PARAM       = "X-Amz-Expires"
	EXPIRE_PARAM_VALUE = "60"
)

type ExecCredential struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     map[string]string `json:"status"`
}

func getToken(ctx context.Context, client *sts.Client, clusterName string) (string, error) {
	presignSts := sts.NewPresignClient(client)
	req, err := presignSts.PresignGetCallerIdentity(
		ctx,
		&sts.GetCallerIdentityInput{},
		func(pso *sts.PresignOptions) {
			pso.ClientOptions = append(pso.ClientOptions, sts.WithAPIOptions(
				addEksHeader(clusterName),
			))
		},
	)
	if err != nil {
		return "", err
	}

	return TOKEN_PREFIX + base64.URLEncoding.
		WithPadding(base64.NoPadding).
		EncodeToString([]byte(req.URL)), nil
}
func QueryGeneratorIKS(metric string, accountid string, name_cluster string) (string, string, error) {

	query := ""
	if metric == "apiserver_requestsbycode" {
		query = `sum by (code) (rate(apiserver_request_total{job=~"kube-apiserver", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m0s]))`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "", nil
	} else if metric == "apiserver_requestsbyverb" {
		query = `sum by (verb) (rate(apiserver_request_total{job=~"kube-apiserver", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m0s]))`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "", nil
	} else if metric == "apiserver_latencybyhostname" {
		query = `1000 * sum(rate(apiserver_request_duration_seconds_sum{job=~"kube-apiserver", cloudaccount=~"%accountid%", clustername=~"%name_cluster%"}[2m0s])) by (hostname)/sum(rate(apiserver_request_duration_seconds_count{job=~"kube-apiserver", cloudaccount=~"%accountid%",clustername=~"%name_cluster%"}[2m0s])) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "ms", nil

	} else if metric == "apiserver_latencybyverb" {
		query = `1000 * sum(rate(apiserver_request_duration_seconds_sum{job=~"kube-apiserver", cloudaccount=~"%accountid%", clustername=~"%name_cluster%"}[2m0s])) by (verb)/sum(rate(apiserver_request_duration_seconds_count{job=~"kube-apiserver", cloudaccount=~"%accountid%",clustername=~"%name_cluster%"}[2m0s])) by (verb)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "ms", nil

	} else if metric == "apiserver_errorsbyhostname" {
		query = `sum by(hostname) (rate(apiserver_request_total{code=~"4..|5..", job=~"kube-apiserver", cloudaccount=~"%accountid%", clustername=~"%name_cluster%"}[2m0s]))
		/ sum by(hostname) (rate(apiserver_request_total{job=~"kube-apiserver", cloudaccount=~"%accountid%", clustername=~"%name_cluster%"}[2m0s]))`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "percentage", nil

	} else if metric == "apiserver_errorsbyverb" {
		query = `sum by(verb) (rate(apiserver_request_total{code=~"4..|5..",job=~"kube-apiserver", cloudaccount=~"%accountid%", clustername=~"%name_cluster%"}[2m0s]))
		/ sum by(verb) (rate(apiserver_request_total{job=~"kube-apiserver", cloudaccount=~"%accountid%", clustername=~"%name_cluster%"}[2m0s]))`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "percentage", nil

	} else if metric == "apiserver_httprequestsbyhostname" {
		query = `sum(rate(apiserver_request_total{job=~"kube-apiserver",clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m0s])) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "", nil
	} else if metric == "apiserver_cpu" {
		query = `sum(rate(process_cpu_seconds_total{job=~"kube-apiserver", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m0s])) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "percentage", nil
	} else if metric == "apiserver_memory" {
		query = `sum(process_resident_memory_bytes{job=~"kube-apiserver", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "bytes", nil

	} else if metric == "etcd_clienttrafficin" {
		query = `sum(rate(etcd_network_client_grpc_received_bytes_total{job=~"etcd",clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m]))by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "B/s", nil

	} else if metric == "etcd_clienttrafficout" {
		query = `sum(rate(etcd_network_client_grpc_sent_bytes_total{job=~"etcd",clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m]))by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "B/s", nil

	} else if metric == "etcd_peertrafficin" {
		query = `sum(rate(etcd_network_peer_received_bytes_total{job=~"etcd",clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m]))by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "B/s", nil
	} else if metric == "etcd_peertrafficout" {
		query = `sum(rate(etcd_network_peer_sent_bytes_total{job=~"etcd",clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m]))by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "B/s", nil
	} else if metric == "etcd_memory" {
		query = `sum(process_resident_memory_bytes{job=~"etcd", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "bytes", nil
	} else if metric == "etcd_dbsizeinuse" {
		query = `sum(etcd_mvcc_db_total_size_in_use_in_bytes{job=~"etcd", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "bytes", nil
	} else if metric == "etcd_cpu" {
		query = `sum(rate(process_cpu_seconds_total{job=~"etcd", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m0s])) by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "percentage", nil
	} else if metric == "etcd_heartbeatsendfailurestotal" {
		query = `sum(etcd_server_heartbeat_send_failures_total{job=~"etcd", clustername=~"%name_cluster%", cloudaccount=~"%accountid%"}[2m0s])by (hostname)`
		query = strings.Replace(query, "%name_cluster%", name_cluster, -1)
		query = strings.Replace(query, "%accountid%", accountid, -1)
		return query, "", nil
	} else {
		//fmt.Println("Unsupported Metric")
		return "", "", fmt.Errorf("unsupported metric")
	}

}

//bytes: MiB divide 10^6
//percentage: *100
// b/s: divide 10^6 Mb/s
// B/s: KB/s divide 10^3
// catch all decimal upto 2 labels
//ms - no changes
//io/s - no changes

func addEksHeader(cluster string) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Build.Add(middleware.BuildMiddlewareFunc("AddEKSId", func(
			ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler,
		) (middleware.BuildOutput, middleware.Metadata, error) {
			switch req := in.Request.(type) {
			case *smithyhttp.Request:
				query := req.URL.Query()
				query.Add(EXPIRE_PARAM, EXPIRE_PARAM_VALUE)
				req.URL.RawQuery = query.Encode()

				req.Header.Add(K8S_AWS_ID_HEADER, cluster)
			}
			return next.HandleBuild(ctx, in)
		}), middleware.Before)
	}
}

func KubeConfigFromRESTConfig(token string, cluster string, ca string, server string) ([]byte, error) {

	context := cluster
	user := "user"

	clusters := make(map[string]*clientcmdapi.Cluster)
	// fmt.Println(ca)
	data, err := os.ReadFile(ca)

	if err != nil {
		return []byte(""), fmt.Errorf("Failed to load ca from vault %v", err)
	}
	//str := "-----BEGIN CERTIFICATE-----\nMIIDBTCCAe2gAwIBAgIICBTTN93ZPcYwDQYJKoZIhvcNAQELBQAwFTETMBEGA1UE\nAxMKa3ViZXJuZXRlczAeFw0yNDA0MDMwODUzNDdaFw0zNDA0MDEwODU4NDdaMBUx\nEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQCvmHKAQD50rrd1JP8C4R7LcWxR7B5Y3AMM793JHWxDDrsXBgMqqGfY+cfS\n/OuEnF2r9bBF5zR4rTdHjzvMImqw6ktK+LLOnVNpP86aD6zxRHwrEWc2BQpDJNiR\noo6CdSCiYbvzFApCNvvLjLovQq02oiQdp9LS75TPadzNIWcMCRfLfcEnSE6IfFye\nPGKt+2AZuSQ28ESepPNlK985XJc0NDKFdwii867j3bYo7mzp+ohKBMSI3hBsVmhA\nEMMnQNtubDmt+RVGqPR+7pgH2pI3+3L/ENyEFA1gmwggqaWUYn264/4dU4h3pZ4e\nj7rV/HMGD4vY0LLlLQCh2W0B0xUnAgMBAAGjWTBXMA4GA1UdDwEB/wQEAwICpDAP\nBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRCiXQVS/hTPIrjB7NTXcDW3sIlojAV\nBgNVHREEDjAMggprdWJlcm5ldGVzMA0GCSqGSIb3DQEBCwUAA4IBAQANBvDnfT2h\nG5gFc/neHc4YldVYbFobBt+UHhwWAVjomfl5SuyEks2CXumwmDymepo2lJYT2NPg\nwSWFZaBVepA59UK7gebDdZJGTTUahm+Na4uB0cIxGvok+ghZQLEYe/kx27q+TFgP\nn75BXFjn4+u/dSWk6Dpk1LBCSYxQArmOOAgOhkouuD/dV/utVaStATVCxtCEE7nA\nAOh2R9CVne0gWanDpWQJFv3HytczhqBk3+Kjp3xkO8sUkDZU/aXXhvyU7mk4htAX\nQ/44jaB0v4uTYFcoftdeOCLb7JC1ao827A1BiY/YGL8KTbdIg61yc8/Kre+gN44i\nc7PjEqZGGbAX\n-----END CERTIFICATE-----\n"

	//data := []byte(str)
	clusters[cluster] = &clientcmdapi.Cluster{
		Server:                   server,
		CertificateAuthorityData: data,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts[context] = &clientcmdapi.Context{
		Cluster:  cluster,
		AuthInfo: user,
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos[user] = &clientcmdapi.AuthInfo{
		Token: token,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: context,
		AuthInfos:      authinfos,
	}
	return clientcmd.Write(clientConfig)
}

//	type Metric struct {
//		Drive string `json:"drive"`
//	}
func CreateVMUserAWSAuth(ca string, clusterName string, cloudaccountID string, vmRecord string, password string, server string, region string, roleArn string) (string, error) {
	ctx := context.Background()
	awsaccesskeysecret, err := os.ReadFile("/vault/secrets/awsaccesskeysecret")
	if err != nil {
		return "", fmt.Errorf("Failed to load awsaccesskeysecret from vault %v", err)
	}
	awsaccesskey, err := os.ReadFile("/vault/secrets/awsaccesskey")
	if err != nil {
		return "", fmt.Errorf("Failed to load awsaccesskey from vault %v", err)
	}
	// fmt.Println(string(awsaccesskeysecret))
	// fmt.Println(string(awsaccesskey))

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(string(awsaccesskey), string(awsaccesskeysecret), "")),
	)

	if err != nil {
		return "", fmt.Errorf("Failed to load AWS config: %v", err)
	}

	client := sts.NewFromConfig(cfg)
	// fmt.Println(roleArn)
	if roleArn != "" {
		creds := stscreds.NewAssumeRoleProvider(client, roleArn)
		cfg.Credentials = aws.NewCredentialsCache(creds)
		client = sts.NewFromConfig(cfg)
	}

	token, err := getToken(ctx, client, clusterName)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch STS token: %v", err)
	}
	// fmt.Println(token)

	kubeConfigBytes, err := KubeConfigFromRESTConfig(token, clusterName, ca, server)

	if err != nil {
		return "", fmt.Errorf("Failed to create bytes kubeconfig: %v", err)
	}
	// fmt.Println(string(kubeConfigBytes))
	if err := os.WriteFile("/kubernetesfile/kubeconfig", kubeConfigBytes, 0600); err != nil {
		return "", fmt.Errorf("Failed to create kubeconfig: %v", err)
	}
	vmusers, err := GetVMUsers("/kubernetesfile/kubeconfig", cloudaccountID)
	if err != nil {
		fmt.Println("Error retrieving VMUser:", err)
		vmuser, err := CreateVMUser("/kubernetesfile/kubeconfig", cloudaccountID, password, vmRecord)
		if err != nil {
			return "", fmt.Errorf("Error craeting VMUser: %v", err)
		}

		fmt.Println("Vmuser created:", vmuser)
		return password, nil
	}
	return vmusers, nil

	// _=createKubeconfig(caCert,clusterName,username,serverURL,userToken,kubeconfigPath)
}

func GetVMUserAWSAuth(ca string, clusterName string, cloudaccountID string, server string, region string, roleArn string) (string, error) {
	ctx := context.Background()
	awsaccesskeysecret, err := os.ReadFile("/vault/secrets/awsaccesskeysecret")
	if err != nil {
		return "", fmt.Errorf("Failed to load awsaccesskeysecret from vault %v", err)
	}
	awsaccesskey, err := os.ReadFile("/vault/secrets/awsaccesskey")
	if err != nil {
		return "", fmt.Errorf("Failed to load awsaccesskey from vault %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(string(awsaccesskey), string(awsaccesskeysecret), "")),
	)

	if err != nil {
		return "", fmt.Errorf("Failed to load AWS config: %v", err)
	}

	client := sts.NewFromConfig(cfg)
	//fmt.Println(roleArn)
	if roleArn != "" {
		creds := stscreds.NewAssumeRoleProvider(client, roleArn)
		cfg.Credentials = aws.NewCredentialsCache(creds)
		client = sts.NewFromConfig(cfg)
	}

	token, err := getToken(ctx, client, clusterName)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch STS token: %v", err)
	}
	//fmt.Println(token)

	kubeConfigBytes, err := KubeConfigFromRESTConfig(token, clusterName, ca, server)

	if err != nil {
		return "", fmt.Errorf("Failed to create bytes kubeconfig: %v", err)
	}
	//fmt.Println(string(kubeConfigBytes))
	if err := os.WriteFile("/kubernetesfile/kubeconfig", kubeConfigBytes, 0600); err != nil {
		return "", fmt.Errorf("Failed to create kubeconfig: %v", err)
	}
	vmusers, err := GetVMUsers("/kubernetesfile/kubeconfig", cloudaccountID)
	if err != nil {
		fmt.Println("Error retrieving VMUser:", err)
		return "", fmt.Errorf("Unable to get vmuser: %v", err)
	}
	return vmusers, nil

	// _=createKubeconfig(caCert,clusterName,username,serverURL,userToken,kubeconfigPath)
}

func getExecAuth(token string) (string, error) {
	execAuth := ExecCredential{
		ApiVersion: API_VERSION,
		Kind:       KIND,
		Status:     map[string]string{"token": token},
	}
	encoded, err := json.Marshal(execAuth)
	return string(encoded), err

}

func GetVMUsers(kubeconfig string, cloudaccount string) (string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return "", fmt.Errorf("Error building kubeconfig from flags %v", err)
	}

	// Create a dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("Error creating dynamic client %v", err)
	}
	result, err := client.Resource(schema.GroupVersionResource{
		Group:    "operator.victoriametrics.com",
		Version:  "v1beta1",
		Resource: "vmusers",
	}).Namespace("cloudmonitor").Get(context.TODO(), "vmuser-"+cloudaccount, metav1.GetOptions{})
	//members, found, err := unstructured.NestedInt64(mdb.UnstructuredContent(), "spec", "members")
	if err != nil {
		return "", fmt.Errorf("Error creating dynamic client %v", err)
	}
	//NestedString(myunstruct.Object, "status", "message")
	fieldValue, found, err := unstructured.NestedString(result.Object, "spec", "bearerToken")
	if err != nil {
		return "", fmt.Errorf("Error getting field %v", err)
	}
	if !found {
		fmt.Println("Field not found")
		return "", fmt.Errorf("Error getting field %v", err)

	}
	return fieldValue, nil

}

func CreateVMUser(kubeconfig string, cloudaccount string, token string, vmRecord string) (string, error) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.victoriametrics.com/v1beta1",
			"kind":       "VMUser",
			"metadata": map[string]interface{}{
				"name": "vmuser-" + cloudaccount,
				"labels": map[string]interface{}{
					"configured-by": "cloudmonitor",
					"tenant":        "cloudmonitortools",
				},
			},
			"spec": map[string]interface{}{
				"bearerToken": token,
				"targetRefs": []map[string]interface{}{{
					"static": map[string]interface{}{
						"url": "http://vmcluster-victoria-metrics-cluster-vminsert:8480/",
					},
					"paths":                      []string{"/api/v1/write"},
					"drop_src_path_prefix_parts": 4,
					"target_path_suffix":         "insert/" + vmRecord + "/prometheus",
				}, {
					"static": map[string]interface{}{
						"url": "http://vmcluster-victoria-metrics-cluster-vmselect:8481/",
					},
					"paths":              []string{"/api/v1/query_range"},
					"target_path_suffix": "select/" + vmRecord + "/prometheus",
				}},
			},
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operator.victoriametrics.com",
		Version: "v1beta1",
		Kind:    "VMUser",
	})

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return "", fmt.Errorf("Error building kubeconfig from flags %v", err)
	}

	// Create a dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("Error creating dynamic client %v", err)
	}
	result, err := client.Resource(schema.GroupVersionResource{
		Group:    "operator.victoriametrics.com",
		Version:  "v1beta1",
		Resource: "vmusers",
	}).Namespace("cloudmonitor").Create(context.TODO(), obj, metav1.CreateOptions{})

	if err != nil {
		return "", fmt.Errorf("Error creating custom crd %v", err)
	} else {
		fmt.Println("Custom resource created successfully ", result)
	}

	return "VMUser", nil

}

func GenerateRandomPassword() (string, error) {
	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	// Restrict symbols to only those defined in the Symbols list
	gen, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "%*&#",
	})
	randomPassword, err := gen.Generate(20, 6, 4, false, false)
	if err != nil {
		return "", fmt.Errorf("could not generate a random password: %v", err)
	}
	return "k" + randomPassword, nil
}

func CreateRecord(cloud_account_id string, resource_id string, jobname string, status string, c *Server, ctx context.Context) error {

	fmt.Println("Creating new OptIn record")

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudMonitor.CreateRecord").WithValues("cloudAccountId", cloud_account_id).Start()
	defer span.End()

	//params := protodb.NewProtoToSql(obj, fieldOpts...)
	// vals := params.GetValues()
	// vals = append(vals, id)

	// query := fmt.Sprintf("INSERT INTO kfaas_deployments (%v, deployment_id) VALUES(%v, $%v)",
	// 	params.GetNamesString(), params.GetParamsString(), len(vals))
	// // fmt.Println(query)
	tx, err := c.session.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error initiating db transaction")
	}

	query := `insert into cloudmonitorbmmapping (cloud_account_id, resource_id, job_name, opt_in_status, created_at) values ($1, $2, $3, $4, $5)`
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, query, cloud_account_id, resource_id, jobname, status, time.Now()); err != nil {
		logger.Error(err, "error inserting cloudmonitorbmmapping", "query", query)
		return fmt.Errorf("unable to create optin record")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing db transaction")
		return fmt.Errorf("unable to create opt in  record")

	}
	fmt.Println("Inserted opt  in record")
	return nil

}

func CreateVMRecord(cloud_account_id string, c *Server, ctx context.Context) (string, error) {

	fmt.Println("Creating VictoriaMetrics ID")

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudMonitor.CreateVMRecord").WithValues("cloudAccountId", cloud_account_id).Start()
	defer span.End()

	fmt.Println("Inserting VM record")
	//params := protodb.NewProtoToSql(obj, fieldOpts...)
	// vals := params.GetValues()
	// vals = append(vals, id)

	// query := fmt.Sprintf("INSERT INTO kfaas_deployments (%v, deployment_id) VALUES(%v, $%v)",
	// 	params.GetNamesString(), params.GetParamsString(), len(vals))
	// // fmt.Println(query)
	tx, err := c.session.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error initiating db transaction")
	}

	query := `insert into cloudmonitorcloudaccountmapping (cloud_account_id,created_at) values ($1, $2)`
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, query, cloud_account_id, time.Now()); err != nil {
		logger.Error(err, "error inserting cloudmonitorcloudaccountmapping", "query", query)
		return "", fmt.Errorf("unable to create KF deployment")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing db transaction")
		return "", fmt.Errorf("unable to create KF deployment")

	}
	fmt.Println("Inserted VM id mapping record")
	return "", nil

}

func (c *Server) EnableMonitor(ctx context.Context, req *pb.EnableMonitorRequest) (*pb.EnableMonitorResponse, error) {
	//first(ctx, req)
	if !c.cfg.EnableMetricsBM {
		return &pb.EnableMonitorResponse{}, fmt.Errorf("Metrics is not enabled for BM resource")
	}

	returnValue := &pb.EnableMonitorResponse{
		Config: "",
		Token:  "",
	}
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorService.EnableMonitor").WithValues("cloudAccountId", req.CloudAccountId).Start()

	cluster := c.cfg.VMClusterName  //"ilens-dev"
	role := c.cfg.IamRole           //"arn:aws:iam::390677890188:role/cloudmonitor-describe-cluster-role"
	server := c.cfg.ClusterEndpoint //"https://E9403701F90A90B408CFF9ABCC4144C5.gr7.us-west-2.eks.amazonaws.com"
	region := c.cfg.AwsVMClusterRegion
	defer span.End()

	//fmt.Println("Request", "req", req)

	// filePathAccessId := "/vault/secrets/accessid"
	// filePathAccessSecret := "/vault/secrets/accesssecret"
	caFile := "/vault/secrets/eksserverca"

	// Create a Kubernetes clientset
	vmRecord, err := GetVMRecord(req.CloudAccountId, c, ctx)
	if err != nil {
		// fmt.Println("Error creating request:", err)
		log.Error(err, "Unable to get VM ID")
		return &pb.EnableMonitorResponse{}, fmt.Errorf("unable to enable monitor")
	}
	if vmRecord == "" {
		newVmRecord, err := CreateVMRecord(req.CloudAccountId, c, ctx)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to get VM ID")
			return &pb.EnableMonitorResponse{}, fmt.Errorf("unable to enable monitor")
		}
		vmRecord, err = GetVMRecord(req.CloudAccountId, c, ctx)
		fmt.Println("Record created", newVmRecord, vmRecord)
	}

	password, err := GenerateRandomPassword()
	if err != nil {
		// fmt.Println("Error creating request:", err)
		log.Error(err, "Unable to get random password")
		return &pb.EnableMonitorResponse{}, fmt.Errorf("unable to enable monitor")
	}
	passwordUpdated, err := CreateVMUserAWSAuth(caFile, cluster, req.CloudAccountId, vmRecord, password, server, region, role)

	if err != nil {
		// fmt.Println("Error creating request:", err)
		log.Error(err, "Unable to create vmuser")
		return &pb.EnableMonitorResponse{}, fmt.Errorf("unable to enable monitor")
	}
	otelconfig, err := os.ReadFile("/vault/secrets/otelconfig")
	if err != nil {
		log.Error(err, "Unable to read otel config from vault for baremetals")
		return &pb.EnableMonitorResponse{}, fmt.Errorf("unable to enable monitor")
	}
	otelconfigString := string(otelconfig)
	//otelconfigString1 := string(otelconfig)
	otelconfigString = strings.Replace(otelconfigString, "%remotewriteendpoint%", c.cfg.RemoteWriteBMAddr, -1)
	otelconfigString = strings.Replace(otelconfigString, "%remotewritetoken%", passwordUpdated, -1)
	returnValue.Config = otelconfigString

	returnValue.Token = passwordUpdated
	return returnValue, nil

}

func QueryGeneratorBM(metric string, cloudaccountid string, jobname string) (string, string, error) {

	query := ""
	if metric == "memory" {
		query = `(
			1 - (
			  (
				node_memory_MemFree_bytes{job="%jobname%"} 
			  ) /
			  node_memory_MemTotal_bytes{job="%jobname%"}
			)
		  )`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "percentage", nil
	} else if metric == "network_receive_bytes" {
		query = `irate(node_network_receive_bytes_total{job = "%jobname%"} [2m])* 8`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "b/s", nil

	} else if metric == "disk" {
		query = `1 - ((node_filesystem_avail_bytes{job="%jobname%",device!~"rootfs"}) / node_filesystem_size_bytes{job="%jobname%",device!~"rootfs"})`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "percentage", nil

	} else if metric == "cpu" {
		query = `sum(rate(node_cpu_seconds_total{mode!="idle", job="%jobname%"}[2m])) by (mode)/ scalar(count(count(node_cpu_seconds_total{job="%jobname%"})by (cpu)))`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "percentage", nil

	} else if metric == "network_transmit_bytes" {
		query = `irate(node_network_transmit_bytes_total{job="%jobname%"}[2m])*8`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "b/s", nil

	} else if metric == "io_traffic_read" {
		query = `irate(node_disk_read_bytes_total{job="%jobname%"}[2m])`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "B/s", nil

	} else if metric == "io_traffic_write" {
		query = `irate(node_disk_written_bytes_total{job="%jobname%"}[2m])`
		query = strings.Replace(query, "%jobname%", jobname, -1)
		return query, "B/s", nil

	} else {
		// log.Info("Unsupported Metric")
		return "", "", fmt.Errorf("unsupported metric")
	}

}

func GetVMRecord(cloud_account_id string, c *Server, ctx context.Context) (string, error) {

	fmt.Println("List Victoria Metrics ID")
	//returnError := &pb.ListKubeFlowDeploymentResponse{}
	victoriaMetricsId := ""
	query := `
		select cloudmonitorid 
		from cloudmonitorcloudaccountmapping 
		where  cloud_account_id = $1
	`
	args := []any{cloud_account_id}
	//readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	// fmt.Println(query)

	rows, err := c.session.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	for rows.Next() {

		err = rows.Scan(&victoriaMetricsId)

		if err != nil {
			//logger.Error(err, "error starting db transaction")
			return "", err
		}

	}

	return victoriaMetricsId, nil

}
func GetOptInRecord(cloud_account_id string, resource_id string, c *Server, ctx context.Context) (string, error) {

	fmt.Println("List Victoria Metrics ID")
	//returnError := &pb.ListKubeFlowDeploymentResponse{}
	job_name := ""
	cloudaccountid := ""
	resourceid := ""
	opt_in_status := ""
	query := `
		select job_name,cloud_account_id,resource_id,opt_in_status
		from cloudmonitorbmmapping 
		where  cloud_account_id = $1 and resource_id= $2 and opt_in_status=$3
	`
	args := []any{cloud_account_id, resource_id, "opt-in"}
	//readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	// fmt.Println(query)

	rows, err := c.session.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	for rows.Next() {
		fmt.Println("1")

		//fmt.Println(&job_name)
		err = rows.Scan(&job_name, &cloudaccountid, &resourceid, &opt_in_status)

		if err != nil {
			//logger.Error(err, "error starting db transaction")
			return "", err
		}

	}
	fmt.Println(resource_id, ";", cloud_account_id, ";", opt_in_status, ";", job_name)
	return job_name, nil

}
