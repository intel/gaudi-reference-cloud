// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package authz

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"embed"
	"fmt"
	fa "io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Test struct {
	Service
	clientConn *grpc.ClientConn
	testDb     manageddb.TestDb
	cfg        *config.Config
}

var test Test

func ClientConn() *grpc.ClientConn {
	return test.clientConn
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*config.Config](&test, &config.Config{})
}

func (test *Test) Init(ctx context.Context, cfg *config.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	var err error
	test.cfg = cfg
	test.mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}

	// Load casbin model
	confFilePath := getPathConfigFile("test/data/", "model.conf")
	_, err = os.Stat(confFilePath)
	if err == nil {
		cfg.ModelFilePath = confFilePath
	} else {
		logger.Error(err, "configuration file set not found, policies not loaded", "policyFilePath", confFilePath)
		return err
	}

	// Synchronize the policies from the CSV file to the database
	confFilePath = getPathConfigFile("test/data/", "policy.csv")
	_, err = os.Stat(confFilePath)
	if err == nil {
		cfg.PolicyFilePath = confFilePath
	} else {
		logger.Error(err, "configuration file set not found, policies not loaded", "policyFilePath", confFilePath)
		return err
	}

	// Synchronize the groups from the CSV file to the database
	confFilePath = getPathConfigFile("test/data/", "groups.csv")
	_, err = os.Stat(confFilePath)
	if err == nil {
		cfg.GroupFilePath = confFilePath
	} else {
		logger.Error(err, "configuration file set not found, groups not loaded", "groupsFilePath", confFilePath)
		return err
	}

	cfg.Features.PoliciesStartupSync = true
	cfg.Features.Watcher = true
	test.casbinEngine, err = NewSyncedEnforcer(test.mdb, cfg)
	if err != nil {
		logger.Error(err, "error initializing casbin engine")
		return err
	}

	// setup resources repository
	confFilePath = getPathConfigFile("test/data/", "resources.yaml")
	test.resourceRepository, err = NewResourceRepository(confFilePath)
	if err != nil {
		logger.Error(err, "configuration file set not found, resources not loaded", "resourcesFilePath", confFilePath)
		return fmt.Errorf("error getPathConfigFile: %m", err)
	}

	cfg.Limits = config.Limits{
		MaxCloudAccountRoles: 100,
		MaxPermissions:       100,
	}

	if err := test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "authz")
	if err != nil {
		return err
	}
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	return nil
}

func (test *Test) Done() error {
	grpcutil.ServiceDone[*config.Config](&test.Service)
	err := test.testDb.Stop(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (test *Test) ClientConn() *grpc.ClientConn {
	return test.clientConn
}

//go:embed test/data/*
var embeddedData embed.FS

type CustomClaims struct {
	jwt.Claims
	Email        string `json:"email"`
	EnterpriseId string `json:"enterpriseId"`
	CountryCode  string `json:"countryCode"`
}

func CreateTokenJWT(email string) (error, string) {
	signerOpts := &jose.SignerOptions{}
	signerOpts.WithType("JWT")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("error generating the key: %v", err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, signerOpts)
	if err != nil {
		log.Fatalf("error creating signer: %v", err)
		return err, ""
	}

	customClaims := CustomClaims{
		Claims: jwt.Claims{
			Issuer:    "your-issuer",
			Subject:   "user-id",
			Audience:  jwt.Audience{"your-audience"},
			Expiry:    jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email:        email,
		EnterpriseId: "enterprise-id",
		CountryCode:  "country-code",
	}

	jwtBuilder := jwt.Signed(signer)
	jwtBuilder = jwtBuilder.Claims(customClaims)
	rawToken, err := jwtBuilder.CompactSerialize()
	if err != nil {
		log.Fatalf("error on compact serialize: %v", err)
		return err, ""
	}
	return nil, rawToken
}

func CreateContextWithToken(email string) context.Context {
	err, rawToken := CreateTokenJWT(email)
	if err != nil {
		log.Fatalf("error creating jwt token: %v", err)
	}

	md := metadata.Pairs(
		"Authorization", "Bearer "+rawToken,
	)
	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}

func getPathConfigFile(folder string, filename string) string {
	// Extract the embedded file to the file system
	filePath := folder + filename
	fileData, err := fa.ReadFile(embeddedData, filePath)
	if err != nil {
		log.Printf("failed to read embedded file: %v", err)
	}
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Printf("failed to create temp dir: %v", err)
	}
	tempFilePath := filepath.Join(tempDir, filename)
	err = ioutil.WriteFile(tempFilePath, fileData, 0644)
	if err != nil {
		log.Printf("failed to write file to temp dir: %v", err)
	}
	return tempFilePath
}
