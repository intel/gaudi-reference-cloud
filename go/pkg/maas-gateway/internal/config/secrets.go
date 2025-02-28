// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

func readFile(directoryName, filename string, log logr.Logger) (string, error) {
	file, err := os.Open(fmt.Sprintf("%s/%s", directoryName, filename))
	if err != nil {
		return "", errors.Wrap(err, "failed to open file")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err, "failed to close file")
		}
	}(file) // Don't forget to close the file

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Wrap(err, "failed to read file")
	}
	return strings.TrimSpace(string(fileBytes)), nil
}

func ReadSecret(secretsDir string, log logr.Logger) (username, password string, err error) {
	username, err = readFile(secretsDir, "username", log)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read username")
	}

	password, err = readFile(secretsDir, "password", log)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to read password")
	}

	return username, password, nil
}
