// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/sha3"

	models "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+<>?"

func GenerateUniqueId() string {
	uuid := uuid.NewString()
	buffer := []byte(uuid)

	hash := make([]byte, 8)

	sha3.ShakeSum256(hash, buffer)
	// Append as string to prefix
	paddedId := base32.StdEncoding.EncodeToString(hash)
	id := strings.ToLower(strings.Trim(paddedId, "="))
	return id
}

func GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("password length must be a positive integer")
	}
	if len(charset) == 0 {
		return "", fmt.Errorf("charset cannot be empty")
	}
	password := make([]byte, length)
	for i := range password {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[charIndex.Int64()]
	}
	return string(password), nil
}

func GenerateUniqueServiceId(sql_model *models.Queries, service_type pb.DpaiServiceType, is_deployment bool) (string, error) {

	id := GenerateUniqueId()

	var count int64 = 0
	var err error
	var deployment_prefix = "dep-"
	if !is_deployment {
		deployment_prefix = ""
	}
	switch service_type {
	case pb.DpaiServiceType_DPAI_WORKSPACE:
		id = fmt.Sprintf("%sws-%s", deployment_prefix, id)
		count, err = sql_model.CheckUniqueWorkspaceId(context.Background(), pgtype.Text{String: id, Valid: true})
	case pb.DpaiServiceType_DPAI_AIRFLOW:
		id = fmt.Sprintf("%saf-%s", deployment_prefix, id)
		count, err = sql_model.CheckUniqueAirflowId(context.Background(), pgtype.Text{String: id, Valid: true})
	case pb.DpaiServiceType_DPAI_POSTGRES:
		id = fmt.Sprintf("%spg-%s", deployment_prefix, id)
		count, err = sql_model.CheckUniquePostgresId(context.Background(), pgtype.Text{String: id, Valid: true})
	default:
		return "", fmt.Errorf("invalid ServiceType value: %s", service_type)
	}

	if err != nil {
		return "", fmt.Errorf("unable to check the uniqueness of the id")
	}

	if count != 0 {
		fmt.Printf("Id: %s is already in use. Trying to generate new value", id)
		id = fmt.Sprintf("%s-%d", id, count)
	}

	return id, err
}

func GenerateUniqueDeploymentId(sql_model *models.Queries, deploymentId string) (string, error) {

	id := GenerateUniqueId()

	var count int64 = 0
	var err error
	var prefix = "dep-"
	if deploymentId != "" {
		id = fmt.Sprintf("tas-%s-%s", strings.TrimPrefix(deploymentId, prefix), id)
		count, err = sql_model.CheckUniqueDeploymentTaskId(context.Background(), pgtype.Text{String: id, Valid: true})
	} else {
		id = fmt.Sprintf("%s%s", prefix, id)
		count, err = sql_model.CheckUniqueDeploymentId(context.Background(), pgtype.Text{String: id, Valid: true})
	}

	if err != nil {
		return "", fmt.Errorf("unable to check the uniqueness of the id")
	}

	if count != 0 {
		fmt.Printf("Id: %s is already in use. Trying to generate new value", id)
		id = fmt.Sprintf("%s-%d", id, count)
	}

	return id, err
}
