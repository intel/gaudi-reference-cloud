package PU_GTS_CountryCode_token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/testsetup"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/zitadel/oidc/pkg/crypto"
	"gopkg.in/square/go-jose.v2"
)

var instanceType string
var sshPublicKey string
var proxyIp string
var sshPrivateKeyPath string
var gtsRegions []string
var staas_payload string
var place_holder_map = make(map[string]string)
var compute_url string
var token string
var instance_id_created string
var iks_version string
var instancesTypes string
var clusterId string
var ssh_publickey_name_created string
var cloud_account_created string

func init() {
	flag.StringVar(&instanceType, "instanceType", "vm-spr-sml", "")
	flag.StringVar(&sshPublicKey, "sshPublicKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCbnYTaUpeo7bsx6pgWhf1IDymryvU7UYVOqfxv5F0Lbpg4osD8bpuopOvVvrMGEdzhxv8Kzr5VaXi8689lPaAXf4Ltx6SFYAQwx+ZdzyqLA73vWftBoTuFvqVe+/IkZ/Y7Vyw+FFZqGOJSc40dLTQ7JfoKV896x5sllP5cO995T+a6R0krX/t00f+VjAypiiK5zWIVQwbGCq4x8upgB4RPeHyQUxMeRzLZAgbEqJBTr5pnZTZjLsPKrhvDp8FRdoUhGMwr6k+pfKL4ZV9T99c0pols5xZBMnreiugPDPt6/w2zXoE/Wa3vXawYZBnt1T0iW5SFJCab85bP/8PkLPRHWtGTttZat9zKWrztVcuG/AaonNi7xtHA6AyAnNs2FnpQOx5VTaMgP2f2l95b8Gg1RN2+5NnT2yxPIN2uCIiTHHjLHxLlpg6rQUs5wUzazpjsj0vfMYTb48d8lOimBJJaMb1iz6DhNtIC9nm8mYRrMgXytMHvUSg+/pxXROaS/zMYdE1dH/PUlnWSW5P3phZ++z1RPVZc7U8k7bNOdHLqXfrAqP+hs6o+9CLfRxOGGQP0jDiYe0S+K8TAt+iiZMxGw5xOa9yItYapZj+zzSaHfLudSvzFjlC+4PY9hIkHvqU08XFzgAEtZh4fkL9HN69Ubt2NrR2Xje8YFYvb0q0d1w== sdp@internal-placeholder.com", "")
	flag.StringVar(&proxyIp, "proxyIp", "10.165.62.252", "")
	flag.StringVar(&sshPrivateKeyPath, "sshPrivateKeyPath", "../../ansible/id-rsa.pub", "")
	testsetup.Get_config_file_data("../../../data/gts.json")
	compute_utils.LoadE2EConfig("../../../data", "staas_input.json")
	compute_url = compute_utils.GetComputeUrl()
	fmt.Println("Stass payload:", staas_payload)
	gtsRegions = testsetup.GetRegions()
	logger.InitializeZapCustomLogger()
}

func createMockGTSToken(region string) string {
	return "Bearer " + generateJWT(map[string]string{}, region)
}

func generateJWT(input map[string]string, region string) string {
	var kArrayClaims = map[string]bool{"amr": true, "roles": true, "wids": true, "groups": true}
	claims := map[string]any{}
	claims["iss"] = "http://issuer:port"
	now := time.Now().UTC()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims["exp"] = now.Add(5 * time.Minute).Unix()
	claims["country"] = region
	for key, val := range input {
		if kArrayClaims[key] {
			claims[key] = []string{val}
		} else {
			claims[key] = val
		}
	}

	rsaTestKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: rsaTestKey}, nil)
	token, _ := crypto.Sign(claims, signer)
	return token
}

func TestPUGTSCountryCodeToken(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "PUGTSCountryCodeToken Suite")
}

type GetProductsResponse struct {
	Products []struct {
		Name        string    `json:"name"`
		ID          string    `json:"id"`
		Created     time.Time `json:"created"`
		VendorID    string    `json:"vendorId"`
		FamilyID    string    `json:"familyId"`
		Description string    `json:"description"`
		Metadata    struct {
			Category     string `json:"category"`
			Disks        string `json:"disks.size"`
			DisplayName  string `json:"displayName"`
			Desc         string `json:"family.displayDescription"`
			DispName     string `json:"family.displayName"`
			Highlight    string `json:"highlight"`
			Information  string `json:"information"`
			InstanceType string `json:"instanceType"`
			Memory       string `json:"memory.size"`
			Processor    string `json:"processor"`
			Region       string `json:"region"`
			Service      string `json:"service"`
		} `json:"metadata"`
		Eccn      string `json:"eccn"`
		Pcq       string `json:"pcq"`
		MatchExpr string `json:"matchExpr"`
		Rates     []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
	} `json:"products"`
}

// Delete volume created....
var _ = AfterSuite(func() {
	//DELETE THE INSTANCE CREATED
	By("Delete FileSystem Created...")
	if place_holder_map["resource_id"] != "" {
		code, body := financials.DeleteFileSystemByResourceId(compute_url, token, place_holder_map["cloud_account_id"], place_holder_map["resource_id"])
		fmt.Println("DeleteFileSystemByResourceId Response: ", code, body)
		Expect(code).To(Equal(200), "Failed to delete FileSystem")
	}
	//DELETE THE INSTANCE CREATED
	By("Deleted instance created...")
	instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
	time.Sleep(10 * time.Second)
	// delete the instance created
	delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, instance_id_created)
	Expect(delete_response_status).To(Or(Equal(200), Equal(404)), "Failed to delete VM instance")
	time.Sleep(5 * time.Second)
	// validate the deletion
	// Adding a sleep because it seems to take some time to reflect the deletion status
	time.Sleep(1 * time.Minute)
	get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
	Expect(get_response_status).To(Equal(404), "Resource shouldn't be found")
	place_holder_map["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	// DELETE SSH KEYS
	By("Delete SSH keys...")
	logger.Logf.Info("Delete SSH keys...")
	fmt.Println("Delete the SSH-Public-Key Created above...")
	ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
	delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
	Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")

})
