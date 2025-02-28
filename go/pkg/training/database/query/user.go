// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	insertUserInfo = `
		INSERT INTO user_accounting (enterprise_id, cloud_account_id, creation_date, expiration_date, latest_ssh_access, latest_jupyter_access, cloud_account_type, linux_user_id, user_email, country_code)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3, NULL, NULL, $4, $5, $6, $7)
		ON CONFLICT (enterprise_id) DO UPDATE SET expiration_date = $3, cloud_account_Type = $4, linux_user_id = $5, user_email = $6;
	`

	insertTrainingMetrics = `
		INSERT INTO training_metrics (enterprise_id, cloud_account_id, training_id, training_name, access_time)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP);
	`

	selectExpirationbyEnterpriseID = `
		SELECT expiration_date
		FROM user_accounting
		WHERE enterprise_id = $1;
	`

	updateUserSSH = `
		UPDATE user_accounting SET latest_ssh_access = CURRENT_TIMESTAMP
		WHERE enterprise_id = $1;
	`

	updateUserJupyter = `
		UPDATE user_accounting SET latest_jupyter_access = CURRENT_TIMESTAMP
		WHERE enterprise_id = $1;
	`
)

func StoreUserRegistration(ctx context.Context, tx *sql.Tx, in *v1.TrainingRegistrationRequest, ExpiryDate string, AccountType v1.AccountType, ProductLookup []*v1.Product, enterprise_id string, linux_username string, user_email string, country_code string) error {
	log := log.FromContext(ctx).WithName("StoreUserRegistration")
	if len(ProductLookup) == 1 {
		_, product := ProductLookup[0], ProductLookup[0]
		productName := product.GetName()
		_, err := tx.ExecContext(ctx, insertTrainingMetrics, enterprise_id, in.CloudAccountId,
			in.TrainingId, productName)
		if err != nil {
			log.Error(err, "error creating record in Training Metrics", "enterpriseId:", enterprise_id, "accountID: ", in.CloudAccountId)
			return errors.New("record insertion failed")
		}
	}
	_, err := tx.ExecContext(ctx, insertUserInfo, enterprise_id, in.CloudAccountId, ExpiryDate, AccountType, linux_username, user_email, country_code)
	if err != nil {
		log.Error(err, "error creating record in Users", "enterpriseId:", enterprise_id, "accountID: ", in.CloudAccountId)
		return errors.New("record insertion failed")
	}

	if in.AccessType == v1.AccessType_ACCESS_TYPE_JUPYTER {
		_, err := tx.ExecContext(ctx, updateUserJupyter, enterprise_id)
		if err != nil {
			log.Error(err, "error updating user type access", "enterpriseId:", enterprise_id, "accountID: ", in.CloudAccountId)
			return errors.New("record insertion failed")
		}
	} else {
		_, err := tx.ExecContext(ctx, updateUserSSH, enterprise_id)
		if err != nil {
			log.Error(err, "error updating user type access", "enterpriseId:", enterprise_id, "accountID: ", in.CloudAccountId)
			return errors.New("record insertion failed")
		}
	}
	return nil
}

func ReadExpiry(ctx context.Context, dbconn *sql.DB, filter *v1.GetDataRequest, enterprise_id string) (*v1.GetDataResponse, error) {
	log := log.FromContext(ctx).WithName("ReadExpiry")
	resp := v1.GetDataResponse{}
	row := dbconn.QueryRowContext(ctx, selectExpirationbyEnterpriseID, enterprise_id)
	switch err := row.Scan(&resp.ExpiryDate); err {
	case sql.ErrNoRows:
		log.Info("no records found", "enterpriseId:", enterprise_id, "accountID: ", filter.CloudAccountId)
		return &resp, nil
	case nil:
		log.Info("record found", "enterpriseId:", enterprise_id, "accountID: ", filter.CloudAccountId)
		return &resp, nil
	default:
		log.Error(err, "error searching User record in db")
		return &resp, status.Errorf(codes.Internal, "training user record find failed")
	}
}

func sliceToJson(slice []string) string {
	json := "["
	for _, item := range slice {
		json += "\"" + item + "\","
	}
	if len(slice) > 0 {
		json = json[:len(json)-1]
	}
	json += "]"
	return json
}
