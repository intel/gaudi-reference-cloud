// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// Generate a 32-byte AES encryption key (store securely!)
func generateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // 256-bit AES key
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func GetEncryptionKey(keyFile string) ([]byte, error) {
	excryptionKey, err := utils.ReadSecretFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption key file: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(*excryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}
	return key, nil
}

// Encrypt password using AES-GCM
func encryptPassword(plainText string, key []byte) (string, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, 12) // GCM standard nonce size
	_, err = rand.Read(nonce)
	if err != nil {
		return "", "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// Encrypt password
	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)

	// Encode to Base64 for easy storage
	return base64.StdEncoding.EncodeToString(cipherText), base64.StdEncoding.EncodeToString(nonce), nil
}

func EncryptPassword(sql *db.Queries, conf config.Config, plainText string) (int32, error) {
	key, err := GetEncryptionKey(conf.Encryption.KeyFile)
	if err != nil {
		return 0, err
	}

	password, nonce, err := encryptPassword(plainText, key)
	if err != nil {
		return 0, err
	}

	secret, err := sql.CreateSecret(context.Background(), db.CreateSecretParams{
		EncryptedPassword: password,
		Nonce:             nonce,
	})
	if err != nil {
		return 0, err
	}

	return secret.ID.Int32, nil
}

// Decrypt password using AES-GCM
func decryptPassword(encryptedText string, nonceText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	// Decode Base64
	cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(nonceText)
	if err != nil {
		return "", err
	}

	// Decrypt
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func DecryptedPassword(sql *db.Queries, conf config.Config, id int32) (string, error) {
	key, err := GetEncryptionKey(conf.Encryption.KeyFile)
	if err != nil {
		return "", err
	}

	secret, err := sql.GetSecret(context.Background(), pgtype.Int4{Int32: id})
	if err != nil {
		return "", err
	}

	password, err := decryptPassword(secret.EncryptedPassword, secret.Nonce, key)
	if err != nil {
		return "", err
	}

	return password, nil
}
