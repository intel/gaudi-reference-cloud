// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package artifactory

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

// Upload and download files in Artifactory.
// Caller should provide a Token, or both UserName and Password.
// None can be provided when using a fake Artifactory with no authentication.
type Artifactory struct {
	Token    string
	UserName string
	Password string
	// Retention days must be set for file uploads.
	// See https://internal-placeholder.com/x/mBqmTQ.
	RetentionDays int
}

type ErrNotFound struct {
	Resource string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *ErrNotFound) Is(target error) bool {
	_, ok := target.(*ErrNotFound)
	return ok
}

func New(ctx context.Context) (*Artifactory, error) {
	return &Artifactory{
		Token:    os.Getenv("ARTIFACTORY_TOKEN"),
		UserName: os.Getenv("ARTIFACTORY_CREDS_USR"),
		Password: os.Getenv("ARTIFACTORY_CREDS_PSW"),
	}, nil
}

func NewFromSecretsDir(ctx context.Context, secretsDir string) (*Artifactory, error) {
	return &Artifactory{
		Token:    string(readFileOrEmpty(ctx, filepath.Join(secretsDir, "ARTIFACTORY_TOKEN"))),
		UserName: string(readFileOrEmpty(ctx, filepath.Join(secretsDir, "ARTIFACTORY_CREDS_USR"))),
		Password: string(readFileOrEmpty(ctx, filepath.Join(secretsDir, "ARTIFACTORY_CREDS_PSW"))),
	}, nil
}

// Download from Artifactory to a file.
// It is safe to call this function concurrently with the same parameters.
func (a Artifactory) Download(ctx context.Context, artifactUrl url.URL, filename string) error {
	log := log.FromContext(ctx).WithName("Download")

	if artifactUrl.Host == "" {
		return fmt.Errorf("url not specified")
	}
	if filename == "" {
		return fmt.Errorf("filename not specified")
	}

	log.Info("Downloading", "url", artifactUrl.String(), "filename", filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, artifactUrl.String(), nil)
	if err != nil {
		return err
	}
	if err := a.prepareRequest(ctx, req); err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &ErrNotFound{Resource: req.URL.String()}
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to get artifact from %s: status %s", req.URL.String(), resp.Status)
	}

	tempFileName := filename + "." + uuid.NewString() + ".tmp"
	out, err := os.Create(tempFileName)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("unable to get artifact from %s: %w", req.URL.String(), err)
	}

	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempFileName, filename); err != nil {
		return err
	}

	log.Info("Download complete", "url", artifactUrl.String(), "filename", filename)
	return nil
}

// Upload a file to Artifactory.
func (a Artifactory) Upload(ctx context.Context, filename string, artifactUrl url.URL) error {
	log := log.FromContext(ctx).WithName("Upload")

	if artifactUrl.Host == "" {
		return fmt.Errorf("url not specified")
	}
	if filename == "" {
		return fmt.Errorf("filename not specified")
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	uploadUrl := artifactUrl

	// Retention days must be set for file uploads.
	// See https://internal-placeholder.com/x/mBqmTQ.
	if a.RetentionDays != 0 {
		uploadUrl.Path = fmt.Sprintf("%s;retention.days=%d", uploadUrl.Path, a.RetentionDays)
	}

	log.Info("Uploading", "url", uploadUrl.String(), "filename", filename)

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadUrl.String(), file)
	if err != nil {
		return err
	}
	if err := a.prepareRequest(ctx, req); err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("unable to put artifact to %s: status %s", req.URL.String(), resp.Status)
		log.Error(err, "Upload failed", "url", uploadUrl.String(), "filename", filename, "Response", respBody.String())
		return err
	}
	log.Info("Upload complete", "url", uploadUrl.String(), "filename", filename, "Response", respBody.String())
	return nil
}

func (a Artifactory) prepareRequest(ctx context.Context, req *http.Request) error {
	log := log.FromContext(ctx).WithName("Artifactory")
	if a.Token != "" {
		// Used by Jenkins.
		log.Info("Using ARTIFACTORY_TOKEN")
		req.Header.Add("X-JFrog-Art-Api", a.Token)
	} else if a.UserName != "" && a.Password != "" {
		// Used by developers to connect to Artifactory.
		log.Info("Using ARTIFACTORY_CREDS_USR and ARTIFACTORY_CREDS_PSW")
		req.SetBasicAuth(a.UserName, a.Password)
	} else {
		// Used to connect to hack/fake-artifactory.sh or a test server.
		log.Info("Using no authentication")
	}
	return nil
}

func readFileOrEmpty(ctx context.Context, filename string) []byte {
	log := log.FromContext(ctx).WithName("Artifactory")
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Info("Unable to read file", "filename", filename, "error", err)
		return nil
	}
	if len(fileBytes) == 0 {
		log.Info("File is empty", "filename", filename)
	}
	return fileBytes
}
