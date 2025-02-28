// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func ConvertToYAMLString(data map[string]interface{}) (string, error) {
	// Marshal the data to YAML
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string
	yamlString := string(yamlBytes)
	return yamlString, nil
}

func ReadConfig() (map[string]interface{}, error) {
	log.Println("Reading the config file")

	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting the current working directory. Error message: %+v", err)
	}
	environment := os.Getenv("ENVIRONMENT")
	filepath := fmt.Sprintf("%s/../../dpai/config/%s.yaml", pwd, environment)
	var data map[string]interface{}

	log.Println(filepath)
	f, err := os.ReadFile(filepath)

	if err != nil {
		return nil, fmt.Errorf("error reading the config file %s. Error message: %+v", filepath, err)
	}

	err = yaml.Unmarshal(f, &data)

	if err != nil {
		return nil, fmt.Errorf("error marshalling the config file %s. Error message: %+v", filepath, err)
	}

	return data, nil
}

func GetK8sConfig(workspace_id string) (*rest.Config, error) {
	// write a function to get the kubeconfig from the secure vault
	// Provide the kubeconfig string
	kubeconfigString := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJek1URXlNekUyTURJd01sb1hEVE16TVRFeU1ERTJNREl3TWxvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBS290CmdpclV0cGFrK2tKWGNOd2l6TDlvZjkzcU1URlBsRnluYkFTNjE5cTRYelQ3Q2lMYVYzRCtMczkwQVNqRjJud2oKVndabVZTTDZMWC9aOVFodFBaWmNhd2xlciswNmdCWkV0QlJweTRRKzZkZlVUNkRoaFRaYThNd1cxblBQUlFpSwpxRGJjbUx3YS9jVHNkYXVueXFMUDdMSkpVM2NjeGk5NFh3ZFhHcG1ReDNvRFczMFJuSWtQM283SmRzU2IrM0wwCmJXdTA1SU9JeWlqMlE5NFlreEczK29WVjR4TFBTUWJEY1IrMVkycW9rblJuT2gwRUd1cEtWcWlEK2l3dW5nenMKa1hZM0ZMQUYyT3kwUk5OZWsxVzFhUENRSmcxQ2FZZkpQRjZKZ1lZbFJLQWg3SUZqR3QycUNNT05GZzdHTU9oeQphSk9FS3E0WnZ2Rzh6TzBtdURzQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZQSEppclJaSndmL3hUQm9VRkJyb3pNWUVpaFZNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBRVNWbjRxTlRKTlVaTkJKdVJRVAp2VWp1cGxBaEUyeUM5VmZuSyszSnprV3ZnUDRXN2k0WFBTS1NlRWVQZmIrbjJoQ0NEK0lnUWZzSTZiSUErazNTCkx3VnF2QnM1SEdwNy9DSVRhQThNRlFTd0VTNHd3emFjVlR0WUowUXpWY3ZYWWdHVXpxSk94OUxzZk0rNmV0ekUKT2dZbGtwUGlTWjVtbWpwQ1EwMEdHbXhXbDZDdjZFT2lhMGlpN1VBaXdPNm5BT2kzOHlscThEM1Z5ZE9GajNMOQpRcHlkSFpEN2w5TUJFMmMvZ2FPNjNzNDdpdHB1RXN3Nm1WWGUrSFlNQ3F0dmxtM0IyQlk4UVhZeEFvOE9RTTBnClpLS2FWNjhRM1YwNWpmV0ZWNm1DcWxBZ0FFN3Fxdy9hS1EzYjZIdU9XaVNMdEExMzVjQVhRbWN5NWxOdzAvV1IKTi9JPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://127.0.0.1:39099
  name: kind-lets-go
contexts:
- context:
    cluster: kind-lets-go
    user: kind-lets-go
  name: kind-lets-go
current-context: kind-lets-go
kind: Config
preferences: {}
users:
- name: kind-lets-go
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJVENDQWdtZ0F3SUJBZ0lJTWpWQ1JKL1ZEc1F3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TXpFeE1qTXhOakF5TURKYUZ3MHlOREV4TWpJeE5qQXlNRFJhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXU5OWgwNUxnVVQ2NGVHbHcKcGZJM2p4R3d6cmpPTC9nbzErUEMrbS92SDFjU1l6RGNCVmRMZmdpZXV3eTJ4UDhmZ0V5SzQvK3FHc1Vqdks1QwpsR2hQQ1c2U2VQdis1WlZibjY2NVlTdStDVm5JK213a2l5NlBGUDlCbmplQUJYMzBoKzBhWEZ1RWhDY3NRc3JBCmVXWWtrcDgyOXlNQVFKT2orNjlpbVdZbDY2U1krKzN6NGtYTUx4RTVoM3RDdlNNS0t5TzlJWEl0OUNFU1R3bGsKczk1Tlg2bmcwQXl3dldrZDA2ZVNGemxScmo3Rk9OV0F0WVBQZG13V1MvM3ZmYklWSUNFT0ZrcTFMcXkxWExraQo5VGZ6ZEhsb1dWZUVsZHpPS242THJCQk1LdkZRakNLWFl5cXBDdlI2TW04WVovdmIwWHpnUEdJRkk5TlRyWHNTCk1hUnQ2d0lEQVFBQm8xWXdWREFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFmQmdOVkhTTUVHREFXZ0JUeHlZcTBXU2NILzhVd2FGQlFhNk16R0JJbwpWVEFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBYlhEUHZ6THo3U2Z5SEFtNTZsNUw3eUlkWVc0UCtNWVZ6Q3dBCjc5Sk5wWWNOSkJjcHZtd09jU2RlZVdSOFl0WkdVQkZRSDhGa0NKaDMwTjNYWUVhNnpIMmVsSmVMckVWWHB1disKaWEwemZkTEkxQk8reGVDY3FQc2VpdVdHenpLUng3NFdlUmVDMyt0QzhWRnl4dUgxTUpLb3lVYkpMWElON2xlbgphby8wV0RtbU44OGJIVy9Tbmo3b2NReWFyek1oYnJXbStQVWU1ZE0vT3E5bDNYMEFYaEQ5bGcwZmxIR0RIb3k1Cmo5Q010TmVWSHI5TDRRa2U4Qzh1ZGlSOEJ2UEp2ZURMbXYwc3VxZ3RUR3krNzkzcjhLcWlMcUpPdUtPMUpiM1QKREF4TlpvVTJHTEUzOUVNcW1DMlZpd3lkRFZJUHVIWFFwejVWbmMvaDNZYmlCd05na1E9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBdTk5aDA1TGdVVDY0ZUdsd3BmSTNqeEd3enJqT0wvZ28xK1BDK20vdkgxY1NZekRjCkJWZExmZ2lldXd5MnhQOGZnRXlLNC8rcUdzVWp2SzVDbEdoUENXNlNlUHYrNVpWYm42NjVZU3UrQ1ZuSSttd2sKaXk2UEZQOUJuamVBQlgzMGgrMGFYRnVFaENjc1FzckFlV1lra3A4Mjl5TUFRSk9qKzY5aW1XWWw2NlNZKyszego0a1hNTHhFNWgzdEN2U01LS3lPOUlYSXQ5Q0VTVHdsa3M5NU5YNm5nMEF5d3ZXa2QwNmVTRnpsUnJqN0ZPTldBCnRZUFBkbXdXUy8zdmZiSVZJQ0VPRmtxMUxxeTFYTGtpOVRmemRIbG9XVmVFbGR6T0tuNkxyQkJNS3ZGUWpDS1gKWXlxcEN2UjZNbThZWi92YjBYemdQR0lGSTlOVHJYc1NNYVJ0NndJREFRQUJBb0lCQUVjekNEc0xqZEdjaUlLeAp5d3hJK0g5VEFBUElDL1FvQXlxV1diMVprSEQ3S2EvSHc0cU9vOENXK2JqL3Y0QjNBM1EzRGVnTWEyWUwwbGhlClhrTXFLTkgxUXJOeEpRL1RBODlIZzEvdEdPOG9STCtMSG1wVThjck9WZ3JsRTdLQklwd2s0bm1nQVYrb0RPRWYKUDhTQ1RsZjIyUGplSGVsYlNxbEd0WUpYTVFYVlM5MlVtYkd3R0ExaDNSOGNkWGNJZG05OHd2OURiM2hHZTdFTgpCdGx1dTRnNTZBTVprR3dzakxRWXBXenE3c3RjTmcwczJHdE9RR1ViMHFhenhPSlhkd2hOcVNhQ2tpT21qK05iCnhUbXM5VTNLaGY5TzdxZkpjTUxlWmR2MmlNMmVjckoyWlpJeVd6RVQwb0lCVHo0YVNvVHdMalBXRzQvS0xGNk4KMHZHcTN3RUNnWUVBejNneXM5WDdvU2pScmZMQ05HVGpGQUt2Sk9lb1lZTDFnUkFEREJiTXdmaW5vQnJ2N3AyegpranJZdUtmTGw1dnpQZlJwaU9yL2xXZXh0TW5BMWRFQkx1L2dLdnVMSC8wekFvU2QwL2JISUxIc3pzSlRKc0RkCjFodzRDWHMrTys3MXRmVVRRSnZpRWJTbnd0eExibGxzcGVDTUhBYTJFSGFzbUVBNDF6dDZ4MnNDZ1lFQTU5R3IKY2toSG9wcUo3L3ZSNFZGYTVsWkUyTCtpRm5ObG8zUVpkNCswVTVsLzBBb1VOL0s5TnhJNUp3aEtncFd6bGU0eApXRUV2emJiOXF6WEx6c3FtaEJaQVplUU5DT09yWFF3K2VJRzR2MFdJOFczOEhGKzFsaHB0RE14YUtZR0lncWhkClg5MEFaK042TDMxRWNFcUN5MTlsUkc1YnJkM3dJaHdtT3hjckU0RUNnWUI4aE9HNWVjYXF0VHNUc2MzSmNIdXgKdU9LeHZiQWJ4bVBlOEdORDBETllnQXJyVnYrR01YUmpZWnBOU2JneDZGR1hiNmdmbXh1RzZsT0g5RWNlalhDdQpZUjRqN2hGVVE3MW4xM3g2aVhSK05RdXlGSU9ma01ZVUNaL0FZUTI3SHZLdUlNZExVY1BmYUpQL3FnUkQyTHdTCjJyQjg2YkpBTTFubmp6UU9TZzZvdlFLQmdRQzlBdlJ5SXNMZHJubEhSYUF6d1h5SEQ5SGtsQ01aZ1E5ZS9WdUMKWTVZWkQ2L3B0VDRVL0FNUTh2Zk1mbk1hRStzd0ZOMEE0MjJSb21Fa0JxSGRRdGdRSElCM1UwRFpmYUFNSnQ2TQpyT0F3ZkFhVEt5VURFWHBpSDd2SVVwbzArMXVOaEs5VmFVRmFjMXY5ZW5xaHYrN3kyUjRkZEltbHNxNkhjVUl1CnJjNCtnUUtCZ1FDY0lVMC9tUUxFRk5FMTRYOXAxT2JiQkJZMTE0UmRKNXRNbnBCUkFGbW1VUlM1Ylg1Z1Z3N2cKY3pUK3ozc2VCQlJ2Q1VwM2N1WmFlOUd6djloL0k5VnIxREpFSlVDSHlyZ2NaUHdvMDE5VmZ5dUd4MElxVTFRYgo3em94VThHcmh2WExvZ0JqRW9KTDAyazQwVUhHU1czczZRNXltdkNzNFBTR245ZDBwbzNiZ0E9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigString))
	if err != nil {
		// Handle error appropriately
		log.Println("Error parsing kubeconfig:", err)
		return nil, err
	}
	return config, nil
}

func GetK8sClientSet(workspace_id string) (*kubernetes.Clientset, error) {
	// write a function to get the kubeconfig from the secure vault
	config, err := GetK8sConfig(workspace_id)
	if err != nil {
		// Handle error appropriately
		log.Printf("Not able to fetch the kubeconfig for the workspace id: %s. Error message: %+v", workspace_id, err)
		return nil, err
	}

	// Create a clientset for interacting with Kubernetes resources
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// Handle error appropriately
		fmt.Println("Error creating clientset:", err)
		return nil, err
	}
	return clientset, nil
}

func GenerateSshKey(comment string) ([]byte, []byte, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %+v", err)
	}

	// Create the private key PEM block
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Encode the private key to PEM format
	privateKeyPEMBytes := pem.EncodeToMemory(privateKeyPEM)

	// Generate the public key in the OpenSSH format
	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %+v", err)
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(pubKey)
	publicKey := []byte(fmt.Sprintf("%s %s", strings.TrimSpace(string(publicKeyBytes)), comment))

	// Print the keys to the console
	fmt.Printf("Private Key:\n%s\n", privateKeyPEMBytes)
	fmt.Printf("Public Key:\n%s\n", publicKey)

	return privateKeyPEMBytes, publicKey, nil
}

func ReadSecretFile(path string) (*string, error) {
	password, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read password file %s: %+v", path, err)
	}
	passwordStr := string(password)
	return &passwordStr, nil
}

func GenerateDpaiSshKeySecretName(clusterName string) string {
	return fmt.Sprintf("ssh-keys-%s", clusterName)
}
