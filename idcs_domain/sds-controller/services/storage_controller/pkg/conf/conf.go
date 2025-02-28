// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func LoadStorageConfig(configLocation string) (*Config, error) {
	data, err := os.Open(configLocation)
	if err != nil {
		return nil, fmt.Errorf("error in loading conf %w", err)
	}

	cn := Config{}
	decoder := yaml.NewDecoder(data)
	decoder.KnownFields(true)

	err = decoder.Decode(&cn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse storage config file %w", err)
	}

	log.Info().Str("configLocation", configLocation).Any("config", cn).Msg("Config loaded from file")

	return &cn, nil
}

func ReadCredentials(auth Auth) (*AuthCreds, error) {
	var creds *AuthCreds
	var err error

	if auth.Env != "" {
		data, exists := os.LookupEnv(auth.Env)
		if !exists {
			return nil, fmt.Errorf("env variable %s provided in config, but not set", auth.Env)
		}
		creds, err = getCreds(data, auth.Scheme)
	} else if auth.File != "" {
		data, ferr := os.ReadFile(auth.File)
		if ferr != nil {
			return nil, fmt.Errorf("could not read credentials from file %w", ferr)
		}
		creds, err = getCreds(string(data), auth.Scheme)
	} else if auth.Secret != "" {
		return nil, errors.New("secret auth is unsupported")
	} else if auth.VaultFile != "" {
		data, ferr := os.ReadFile(auth.VaultFile)
		if ferr != nil {
			return nil, fmt.Errorf("could not read credentials from file %w", ferr)
		}
		var vaultCreds VaultSecret
		err = json.Unmarshal(data, &vaultCreds)
		if err == nil {
			creds = &AuthCreds{
				Scheme:      auth.Scheme,
				Principal:   vaultCreds.Data.Username,
				Credentials: vaultCreds.Data.Password,
			}
		}
	} else {
		return nil, errors.New("no credentials location were specified, must have at least one [env,file,secret]")
	}

	if err != nil {
		return nil, fmt.Errorf("could not obtain creds for %v with error: %w", auth, err)
	}

	return creds, nil
}

func getCreds(data string, scheme AuthScheme) (*AuthCreds, error) {
	var creds AuthCreds

	if scheme == Basic {
		pair := strings.Split(data, ":")
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid basic credentials provided %s != 2", strconv.Itoa(len(pair)))
		}

		creds = AuthCreds{
			Scheme:      scheme,
			Principal:   pair[0],
			Credentials: pair[1],
		}
	} else if scheme == Bearer || scheme == Digest {
		creds = AuthCreds{
			Scheme:      scheme,
			Credentials: data,
		}
	}

	return &creds, nil
}
