// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretType struct {
	Opaque         bool
	DockerRegistry bool
}

func (k *K8sClient) ListSecrets(namespace string) []corev1.Secret {
	var allSecrets []corev1.Secret
	secrets, _ := k.ClientSet.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	allSecrets = append(allSecrets, secrets.Items...)
	return allSecrets
}

func (k *K8sClient) IsSecretExists(namespace string, name string) (exists bool, secret *corev1.Secret) {

	secrets := k.ListSecrets(namespace)
	for _, s := range secrets {
		if s.Name == name {
			log.Printf("The name matches: %s == %s", name, s.Name)
			return true, &s
		}
	}
	return false, nil
}

func (k *K8sClient) CreateSecret(namespace string, name string, secretData map[string][]byte) error {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace, // Adjust namespace if needed
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	result, err := k.ClientSet.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Falied to create new secret %s under the namespace %s. Error message: %+v", name, namespace, err)
		return err
	}

	log.Printf("Secret created successfully: %s\n", result.GetObjectMeta().GetName())

	return nil
}

type DockerConfigSecretData struct {
	Server   string
	UserName string
	Password string
	Email    string
}

// encodeAuth encodes Docker registry credentials for the "auth" field
func encodeAuth(username, password string) string {
	auth := username + ":" + password
	return encodeBase64(auth)
}

// encodeBase64 encodes a string in base64
func encodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func (k *K8sClient) CreateSecretDockerConfigJson(namespace string, name string, secretData DockerConfigSecretData) error {

	secretValue := map[string][]byte{
		".dockerconfigjson": []byte(
			fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"%s","auth":"%s"}}}`,
				secretData.Server, secretData.UserName, secretData.Password, secretData.Email, encodeAuth(secretData.UserName, secretData.Password))),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace, // Adjust namespace if needed
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: secretValue,
	}

	result, err := k.ClientSet.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Falied to create new secret %s under the namespace %s. Error message: %+v", name, namespace, err)
		return err
	}

	log.Printf("Secret created successfully: %s\n", result.GetObjectMeta().GetName())

	return nil
}

func (k *K8sClient) GetSecret(namespace string, name string) (map[string][]byte, error) {
	secret, err := k.ClientSet.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secretData := secret.Data
	return secretData, nil
}

func (k *K8sClient) GetSecretValue(namespace string, name string, key string) (string, error) {
	secretData, err := k.GetSecret(namespace, name)
	if err != nil {
		return "", err
	}
	secret := string(secretData[key])
	return secret, nil
}

func (k *K8sClient) DeleteSecret(namespace string, name string, ignoreIfNotExists bool) error {

	ok, _ := k.IsSecretExists(namespace, name)

	if !ok && ignoreIfNotExists {
		log.Printf("secret %s does not exists in namespace %s", name, namespace)
		return nil
	} else if !ok && !ignoreIfNotExists {
		return fmt.Errorf("secret %s does not exists in namespace %s", name, namespace)
	} else {
		err := k.ClientSet.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Deletion of secret %s from the namespace: %s failed with the error: %+v", name, namespace, err)
			return err
		}
		log.Printf("Secret %s deleted successfully\n", name)
		return nil
	}

}
