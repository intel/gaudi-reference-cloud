// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Minio           *minio.Client
}

func (m *Minio) GetMinioClient() error {

	useSSL := true

	if m.AccessKeyID == "" || m.SecretAccessKey == "" || m.Endpoint == "" {
		return fmt.Errorf("insufficient input: AccessKey or SecretKey or Endpoint is missing")
	}

	m.Endpoint = strings.TrimPrefix(m.Endpoint, "https://")

	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}} // Custom HTTP client using the modified transport

	// Initialize minio client object.
	minioClient, err := minio.New(m.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(m.AccessKeyID, m.SecretAccessKey, ""),
		Secure:    useSSL,
		Transport: transport,
	})
	if err != nil {
		return err
	}

	log.Printf("%#v\n", minioClient)
	m.Minio = minioClient

	return nil
}

func (m *Minio) IsFolderExists(bucketName string, folderName string) (bool, error) {
	log.Println("Inside IsFolderExists")
	// Create context
	ctx := context.Background()

	// Check if folder exists
	folderExists := false
	objectCh := m.Minio.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    folderName,
		Recursive: false,
	})
	log.Println("Inside IsFolderExists, ListObjects completed ")

	for object := range objectCh {
		if object.Err != nil {
			return false, object.Err
		}
		folderExists = true
		break
	}
	log.Println("Inside IsFolderExists Completed. ")

	return folderExists, nil
}

func (m *Minio) CreateFolder(bucketName string, folderName string) error {
	log.Println("Inside Create Folder")
	folderExists, err := m.IsFolderExists(bucketName, folderName)
	if err != nil {
		return err
	}

	// If folder doesn't exist, create it
	if !folderExists {
		// Create a zero-byte object with folderName as the key (object name)
		log.Println("Starting to create the folder")
		_, err = m.Minio.PutObject(context.Background(), bucketName, folderName, nil, 0, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("Folder '%s' created successfully.\n", folderName)
	} else {
		fmt.Printf("Folder '%s' already exists.\n", folderName)
	}
	return nil
}

func (m *Minio) DeleteFolder(bucketName string, folderName string) error {
	// Create context
	ctx := context.Background()

	// List all objects in the folder
	objectCh := m.Minio.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    folderName,
		Recursive: true,
	})

	// Delete all objects in the folder
	objectsToDelete := []minio.ObjectInfo{}
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		objectsToDelete = append(objectsToDelete, object)
	}

	if len(objectsToDelete) == 0 {
		fmt.Printf("Folder '%s' does not exist or is already empty.\n", folderName)
	} else {
		fmt.Printf("Deleting folder '%s'...\n", folderName)

		// Delete objects one by one
		for _, object := range objectsToDelete {
			err := m.Minio.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Fatalln("Failed to delete object:", object.Key, err)
			}
			fmt.Printf("Deleted object: %s\n", object.Key)
		}

		fmt.Printf("Folder '%s' deleted successfully.\n", folderName)
	}
	return nil
}

func ExtractFolderPath(s3Path string) (string, error) {
	// Remove the "s3://" prefix
	trimmedPath := strings.TrimPrefix(s3Path, "s3://")

	// Split the remaining path into bucket and folder parts
	parts := strings.SplitN(trimmedPath, "/", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("folder path not found in %s", s3Path)
	}

	// Return the folder path
	return parts[1], nil
}
