package PU_CredentialsAsAService_test

import (
	"flag"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	compute_url = compute_utils.GetComputeUrl()
	financials_utils.LoadE2EConfig("../../financials/data", "billing.json")
	auth.Get_config_file_data("../../../data/config.json")
	logger.InitializeZapCustomLogger()
}

func TestPU_CredentialsAsAService_test(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "PUCredentialsAsAServiceTest Suite")
}
