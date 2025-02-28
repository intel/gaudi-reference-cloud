// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"testing"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// clean up database by deleting all users
func cleanUserAccounting(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM user_accounting")
	return err
}

// clean up database by deleting all metrics
func cleanTrainingMetrics(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM training_metrics")
	return err
}

func setupDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", "postgres://oneapiint_so:6Z7AvB5X0IeLqU1@postgres5808-lb-fm-in.dbaas.intel.com:5432/oneapiint")
	if err != nil {
		t.Fatalf("failed DB init %v", err)
	}
	return db, func() {
		db.Close()
	}
}

func getLatestSSHAccess(db *sql.DB, enterpriseId string) (string, error) {
	var latestSSHAccess string
	query := `
		SELECT latest_ssh_access
		FROM user_accounting
		WHERE enterprise_id = $1
	`
	err := db.QueryRow(query, enterpriseId).Scan(&latestSSHAccess)
	return latestSSHAccess, err
}

func getLatestJupyterAccess(db *sql.DB, enterpriseId string) (string, error) {
	var latestJupyterAccess string
	query := `
		SELECT latest_jupyter_access
		FROM user_accounting
		WHERE enterprise_id = $1
	`
	err := db.QueryRow(query, enterpriseId).Scan(&latestJupyterAccess)
	return latestJupyterAccess, err
}

// Store User
func TestStoreUserRegistration(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("transaction init failed %v", err)
	}
	defer tx.Rollback()

	ctx := context.Background()

	ExpiryDate := "2023-10-04T00:00:00Z"
	accountType := v1.AccountType_ACCOUNT_TYPE_ENTERPRISE
	products := []*v1.Product{}
	enterpriseId := "test-enterprise-id"
	linuxUsername := "test-linux-username"
	userEmail := "test-user-email"
	countryCode := "test-country-code"

	t.Run("insert user, request ssh", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		in := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
			AccessType:     v1.AccessType_ACCESS_TYPE_SSH,
		}

		// Call the function under test
		err = StoreUserRegistration(ctx, tx, in, ExpiryDate, accountType, products, enterpriseId, linuxUsername, userEmail, countryCode)
		assert.NoError(t, err)
		err = tx.Commit()
		assert.NoError(t, err)

		// check the user was added
		query := `
			SELECT FROM user_accounting
			WHERE enterprise_id = $1
		`
		result, err := db.ExecContext(ctx, query, enterpriseId)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// check that the ssh access time was updated
		latestSSHAccess, err := getLatestSSHAccess(db, enterpriseId)
		assert.NoError(t, err)
		assert.NotNil(t, latestSSHAccess)

		// clean up
		err = cleanUserAccounting(db)
		assert.NoError(t, err)
		err = cleanTrainingMetrics(db)
		assert.NoError(t, err)
	})

	t.Run("insert user, request jupyter", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		in := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
			AccessType:     v1.AccessType_ACCESS_TYPE_JUPYTER,
		}

		// Call the function under test
		err = StoreUserRegistration(ctx, tx, in, ExpiryDate, accountType, products, enterpriseId, linuxUsername, userEmail, countryCode)
		assert.NoError(t, err)
		err = tx.Commit()
		assert.NoError(t, err)

		// check the user was added
		query := `
			SELECT FROM user_accounting
			WHERE enterprise_id = $1
		`
		result, err := db.ExecContext(ctx, query, enterpriseId)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// check that the jupyter access time was updated
		latestJupyterAccess, err := getLatestJupyterAccess(db, enterpriseId)
		assert.NoError(t, err)
		assert.NotNil(t, latestJupyterAccess)

		// clean up
		err = cleanUserAccounting(db)
		assert.NoError(t, err)
		err = cleanTrainingMetrics(db)
		assert.NoError(t, err)
	})

	t.Run("insert user and lookup product", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		products = []*v1.Product{{
			Name: "test-product",
		}}

		in := &v1.TrainingRegistrationRequest{
			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
			AccessType:     v1.AccessType_ACCESS_TYPE_SSH,
		}

		// Call the function under test
		err = StoreUserRegistration(ctx, tx, in, ExpiryDate, accountType, products, enterpriseId, linuxUsername, userEmail, countryCode)
		assert.NoError(t, err)
		err = tx.Commit()
		assert.NoError(t, err)

		// check the user was added
		query := `
			SELECT FROM user_accounting
			WHERE enterprise_id = $1
		`
		result, err := db.ExecContext(ctx, query, enterpriseId)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// check that the ssh access time was updated
		latestSSHAccess, err := getLatestSSHAccess(db, enterpriseId)
		assert.NoError(t, err)
		assert.NotNil(t, latestSSHAccess)

		// check that the paroduct access time was updated
		query = `
			SELECT FROM training_metrics
			WHERE enterprise_id = $1 AND training_id = $2
		`
		result, err = db.ExecContext(ctx, query, enterpriseId, in.TrainingId)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		rowsAffected, err = result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// clean up
		err = cleanUserAccounting(db)
		assert.NoError(t, err)
		err = cleanTrainingMetrics(db)
		assert.NoError(t, err)
	})

}

// Read Expiry
func TestReadExpiry(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("transaction init failed %v", err)
	}
	defer tx.Rollback()

	ctx := context.Background()

	ExpiryDate := "2023-10-04T00:00:00Z"
	accountType := v1.AccountType_ACCOUNT_TYPE_ENTERPRISE
	products := []*v1.Product{}
	enterpriseId := "test-enterprise-id"
	linuxUsername := "test-linux-username"
	userEmail := "test-user-email"
	countryCode := "test-country-code"

	t.Run("no user", func(t *testing.T) {
		// Call the function under test
		filter := v1.GetDataRequest{
			CloudAccountId: "test-account",
		}
		response, err := ReadExpiry(ctx, db, &filter, enterpriseId)
		assert.NoError(t, err)
		assert.Equal(t, &v1.GetDataResponse{}, response)
	})

	t.Run("user exists", func(t *testing.T) {
		//store mock user
		in := &v1.TrainingRegistrationRequest{

			CloudAccountId: "test-account",
			TrainingId:     "test-training",
			SshKeyNames:    []string{},
		}

		err = StoreUserRegistration(ctx, tx, in, ExpiryDate, accountType, products, enterpriseId, linuxUsername, userEmail, countryCode)
		assert.NoError(t, err)
		err = tx.Commit()
		assert.NoError(t, err)

		// Call the function under test
		filter := v1.GetDataRequest{
			CloudAccountId: "test-account",
		}
		response, err := ReadExpiry(ctx, db, &filter, enterpriseId)
		assert.NoError(t, err)
		expiry := response.GetExpiryDate()
		assert.Equal(t, ExpiryDate, expiry)

		// clean up
		err = cleanUserAccounting(db)
		assert.NoError(t, err)
		err = cleanTrainingMetrics(db)
		assert.NoError(t, err)
	})
}

func TestSliceToJson(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	slice := []string{"a", "b", "c"}
	json := sliceToJson(slice)
	assert.Equal(t, "[\"a\",\"b\",\"c\"]", json)
}
