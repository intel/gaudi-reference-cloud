// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type UserRequest struct {
	HashedEnterpriseId string `json:"hashedEnterpriseId"`
	RandomUserId       string `json:"randomUserId"`
	CouponExpiration   string `json:"couponExpiration"`
	SelectedTraining   string `json:"selectedTraining"`
}

var (
	db         *sql.DB
	ldapMutex  sync.Mutex
	slurmMutex sync.Mutex
)

func initDB() {
	// DB CONNECTION

	// read RegistrationDB info from mount
	secretsPath := "/etc/secrets/app"

	hostPath := filepath.Join(secretsPath, "host")
	host, err := os.ReadFile(hostPath)
	if err != nil {
		log.Fatalf("Failed to read secrets from %s: %v", hostPath, err)
	}

	portPath := filepath.Join(secretsPath, "port")
	port, err := os.ReadFile(portPath)
	if err != nil {
		log.Fatalf("Failed to read secrets from %s: %v", portPath, err)
	}

	userPath := filepath.Join(secretsPath, "user")
	user, err := os.ReadFile(userPath)
	if err != nil {
		log.Fatalf("Failed to read secrets from %s: %v", userPath, err)
	}

	passwordPath := filepath.Join(secretsPath, "password")
	password, err := os.ReadFile(passwordPath)
	if err != nil {
		log.Fatalf("Failed to read secrets from %s: %v", passwordPath, err)
	}

	dbnamePath := filepath.Join(secretsPath, "dbname")
	dbname, err := os.ReadFile(dbnamePath)
	if err != nil {
		log.Fatalf("Failed to read secrets from %s: %v", dbnamePath, err)
	}

	certsPath := "/etc/secrets/ca"
	cacertPath := filepath.Join(certsPath, "ca.crt")

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=verify-full sslrootcert=%s",
		string(host),
		string(port),
		string(user),
		string(password),
		string(dbname),
		cacertPath,
	)
	// below is a false positive from Coverity for hardcoded credentials
	// coverity[HARDCODED_CREDENTIALS:FALSE]
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
}

func main() {
	initDB()
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/user", registerUser).Methods("POST")

	// TODO: add TLS and authorization token

	server := &http.Server{
		Addr:    "localhost:8000",
		Handler: router,
	}
	log.Fatal(server.ListenAndServe())
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	var userRequest UserRequest
	err = json.Unmarshal(body, &userRequest)
	if err != nil {
		log.Printf("Error unmarshaling request body: %v", err)
		http.Error(w, "Error unmarshaling request body", http.StatusInternalServerError)
		return
	}

	log.Printf("Processing request: %+v", userRequest)

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	query := `
		INSERT INTO training_user
			(hashed_enterprise_id, random_user_id, launch_time, coupon_expiration, selected_training)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (hashed_enterprise_id)
		DO UPDATE SET coupon_expiration = $6, launch_time = $7, selected_training = $8;
	`

	launch_start_time := time.Now().UTC()
	coupon_expiration, err := time.Parse(time.RFC3339, userRequest.CouponExpiration)
	if err != nil {
		if userRequest.CouponExpiration == "" {
			coupon_expiration, err = time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
			if err != nil {
				log.Printf("Error parsing default coupon expiration from body: %v", err)
				http.Error(w, "Error parsing default coupon expiration from body", http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Error parsing coupon expiration from body: %v", err)
			http.Error(w, "Error parsing coupon expiration from body", http.StatusInternalServerError)
			return
		}
	}

	_, err = tx.Exec(
		query,
		userRequest.HashedEnterpriseId,
		userRequest.RandomUserId,
		launch_start_time,
		coupon_expiration,
		userRequest.SelectedTraining,
		coupon_expiration,
		launch_start_time,
		userRequest.SelectedTraining,
	)
	if err != nil {
		log.Printf(
			"Error inserting/updating user with hashed_enterprise_id '%s', random_user_id '%s', coupon_expiration '%s', launch_start_time '%s', selected_training '%s': %v",
			userRequest.HashedEnterpriseId,
			userRequest.RandomUserId,
			userRequest.CouponExpiration,
			launch_start_time.String(),
			userRequest.SelectedTraining,
			err,
		)
		http.Error(w, "Error adding or updating user", http.StatusInternalServerError)
		return
	}

	message := "User registered/updated successfully"
	log.Printf(
		"Successfully registered/updated user with hashed_enterprise_id '%s', random_user_id '%s', coupon_expiration '%s', launch_start_time '%s', selected_training '%s'",
		userRequest.HashedEnterpriseId,
		userRequest.RandomUserId,
		userRequest.CouponExpiration,
		launch_start_time.String(),
		userRequest.SelectedTraining,
	)
	response := map[string]string{
		"hashed_enterprise_id": userRequest.HashedEnterpriseId,
		"random_user_id":       userRequest.RandomUserId,
		"coupon_expiration":    userRequest.CouponExpiration,
		"launch_start_time":    launch_start_time.String(),
		"selected_training":    userRequest.SelectedTraining,
		"message":              message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		log.Printf("Error encoding JSON response: %v", jsonErr)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
	_, writeErr := w.Write(jsonResponse)
	if writeErr != nil {
		log.Printf("Error writing JSON response: %v", jsonErr)
		http.Error(w, "Error writing JSON response", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}
	return
}
