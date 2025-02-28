// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v4

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/google/uuid"
	auth "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/ci/k6/modules/weka/auth"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwt"
)

type OrganizationContainer struct {
	uid       string
	fsNames   map[string]string
	usernames map[string]string
	passwords map[string]string
}

// simulate the container configuration info
type ContainerInfo struct {
	uid string
}

type SimulatedWekaAPI struct {
	lock               sync.Mutex
	fa                 *auth.FakeAuthenticator
	organizations      map[string]*Organization
	filesystems        map[string]*FileSystem
	fsGroups           map[string]*FileSystemGroup
	users              map[string]*User
	containers         map[string]*Container
	processes          map[string]*Process
	fsGroupNames       map[string]bool
	orgNames           map[string]*OrganizationContainer
	containerHostNames map[string]*ContainerInfo
	buckets            *S3Bucket
	lifecycleRules     map[string]*S3LifecycleRule
	polices            map[string]*S3IAMPolicy
	policyMapping      map[string]string
	ServerInterface
}

func (s *SimulatedWekaAPI) Login(ctx v4.Context) error {
	var req LoginJSONBody

	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	var userID string
	var orgName string
	if req.Org == nil || *req.Org == "" {
		orgName = "Root"
	} else {
		orgName = *req.Org
	}

	org, exists := s.orgNames[orgName]

	if !exists {
		return ctx.JSON(http.StatusUnauthorized, N401{})
	}

	userID, exists = org.usernames[req.Username]
	if !exists || org.passwords[userID] != req.Password {
		userID = ""
	}

	if userID == "" {
		return ctx.JSON(http.StatusUnauthorized, N401{})
	}

	readerJWS, err := s.fa.CreateJWSWithClaims(userID, org.uid, []string{})
	if err != nil {
		return ctx.JSON(http.StatusOK, N403{})
	}

	tokenValue := string(readerJWS)
	refreshToken := "refresh"
	expiresIn := 3000
	changeRequired := false
	tokenType := "Bearer"
	var token interface{} = Tokens{
		AccessToken:            &tokenValue,
		ExpiresIn:              &expiresIn,
		PasswordChangeRequired: &changeRequired,
		RefreshToken:           &refreshToken,
		TokenType:              &tokenType,
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &token,
	})
}

func (s *SimulatedWekaAPI) GetOrganizations(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	orgs := make([]Organization, 0, len(s.organizations))
	for _, val := range s.organizations {
		orgs = append(orgs, *val)
	}
	var answer interface{} = orgs
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) CreateOrganization(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req CreateOrganizationJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}
	_, exists := s.orgNames[req.Name]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	newOrg := createOrganization(req.Name, req.TotalQuota, req.SsdQuota)
	s.organizations[*newOrg.Uid] = &newOrg
	orgContainer := OrganizationContainer{
		uid:       *newOrg.Uid,
		fsNames:   make(map[string]string),
		passwords: make(map[string]string),
		usernames: make(map[string]string),
	}
	adminUser := createUser(req.Username, *newOrg.Id, "OrgAdminUser")
	orgContainer.passwords[*adminUser.Uid] = req.Password
	orgContainer.usernames[*adminUser.Username] = *adminUser.Uid

	s.orgNames[req.Name] = &orgContainer
	s.users[*adminUser.Uid] = &adminUser

	var answer interface{} = newOrg

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetMultipleOrgExist(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var answer interface{}
	if len(s.organizations) > 1 {
		answer = true
	} else {
		answer = false
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) DeleteOrganization(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	org, exists := s.organizations[uid]
	if exists {
		delete(s.organizations, uid)
		for _, fsUID := range s.orgNames[*org.Name].fsNames {
			delete(s.filesystems, fsUID)
		}
		for _, userUID := range s.orgNames[*org.Name].usernames {
			delete(s.users, userUID)
		}
		delete(s.orgNames, *org.Name)
		return ctx.JSON(http.StatusOK, N200{
			Data: nil,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) GetOrganization(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	org, exists := s.organizations[uid]
	if exists {
		var answer interface{} = org
		return ctx.JSON(http.StatusOK, N200{
			Data: &answer,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) UpdateOrganization(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req UpdateOrganizationJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	org, exists := s.organizations[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}
	_, exists = s.orgNames[*req.NewName]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}
	s.orgNames[*req.NewName] = s.orgNames[*org.Name]
	org.Name = req.NewName
	delete(s.orgNames, *org.Name)

	var answer interface{} = org

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) SetOrganizationLimit(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req SetOrganizationLimitJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	org, exists := s.organizations[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}

	org.TotalQuota = req.TotalQuota
	org.SsdQuota = req.SsdQuota

	var answer interface{} = org

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetUsers(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	users := make([]User, 0)
	for _, val := range s.orgNames[*org.Name].usernames {
		user, exists := s.users[val]
		if exists {
			users = append(users, *user)
		}
	}
	var answer interface{} = users

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) CreateUser(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req CreateUserJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	_, exists = s.orgNames[*org.Name].usernames[*req.Username]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	newUser := createUser(*req.Username, *org.Id, string(*req.Role))
	s.users[*newUser.Uid] = &newUser
	s.orgNames[*org.Name].passwords[*newUser.Uid] = *req.Password
	s.orgNames[*org.Name].usernames[*req.Username] = *newUser.Uid
	var answer interface{} = newUser

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) UpdateUserPassword(ctx v4.Context) error {
	var req UpdateUserPasswordJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	user := ctx.Get("user").(jwt.Token)
	subject := user.Subject()
	s.orgNames[*org.Name].passwords[subject] = req.NewPassword

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) WhoAmI(ctx v4.Context) error {
	user := ctx.Get("user").(jwt.Token)
	subject := user.Subject()
	var answer interface{} = s.users[subject]
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) DeleteUser(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	user, exists := s.users[uid]
	if exists {
		delete(s.users, uid)
		delete(s.orgNames[*org.Name].passwords, uid)
		delete(s.orgNames[*org.Name].usernames, *user.Username)
		return ctx.JSON(http.StatusOK, N200{
			Data: nil,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) UpdateUser(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req UpdateUserJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	user, exists := s.users[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}

	user.Role = (*string)(req.Role)

	var answer interface{} = user

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) SetUserPassword(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req SetUserPasswordJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	user, exists := s.users[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	s.orgNames[*org.Name].passwords[uid] = req.Password

	var answer interface{} = user

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetFileSystemGroups(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	fsGroups := make([]FileSystemGroup, 0, len(s.users))
	for _, val := range s.fsGroups {
		fsGroups = append(fsGroups, *val)
	}
	var answer interface{} = fsGroups

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) CreateFileSystemGroup(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req CreateFileSystemGroupJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}
	_, exists := s.fsGroupNames[req.Name]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}
	uid := genUID()
	id := fmt.Sprintf("FSGroupId<%d>", rand.Int31())
	newFsGroup := FileSystemGroup{
		Uid:                &uid,
		Name:               &req.Name,
		StartDemote:        req.StartDemote,
		TargetSsdRetention: req.TargetSsdRetention,
		Id:                 &id,
	}
	s.fsGroups[uid] = &newFsGroup
	s.fsGroupNames[req.Name] = true
	var answer interface{} = newFsGroup

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) DeleteFileSystemGroup(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	fsGroup, exists := s.fsGroups[uid]
	if exists {
		delete(s.fsGroups, uid)
		delete(s.fsGroupNames, *fsGroup.Name)
		return ctx.JSON(http.StatusOK, N200{
			Data: nil,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) GetFileSystemGroup(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	fsGroup, exists := s.fsGroups[uid]
	if exists {
		var answer interface{} = fsGroup
		return ctx.JSON(http.StatusOK, N200{
			Data: &answer,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) UpdateFileSystemGroup(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req UpdateFileSystemJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	fsGroup, exists := s.fsGroups[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}
	_, exists = s.fsGroupNames[*req.NewName]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}
	fsGroup.Name = req.NewName
	s.fsGroupNames[*req.NewName] = true
	delete(s.fsGroupNames, *fsGroup.Name)
	var answer interface{} = fsGroup

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetFileSystems(ctx v4.Context, _ GetFileSystemsParams) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	filesystems := make([]FileSystem, 0)
	for _, val := range s.orgNames[*org.Name].fsNames {
		filesystems = append(filesystems, *s.filesystems[val])
	}
	var answer interface{} = filesystems

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) CreateFileSystem(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req CreateFileSystemJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	_, exists = s.orgNames[*org.Name].fsNames[req.Name]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}
	uid := genUID()
	id := fmt.Sprintf("Filesystem<%d>", rand.Int31())
	newFilesystem := FileSystem{
		Uid:          &uid,
		Name:         &req.Name,
		AuthRequired: req.AuthRequired,
		IsEncrypted:  req.Encrypted,
		GroupName:    &req.GroupName,
		SsdBudget:    req.SsdCapacity,
		TotalBudget:  &req.TotalCapacity,
		Id:           &id,
	}
	s.filesystems[uid] = &newFilesystem
	s.orgNames[*org.Name].fsNames[req.Name] = uid
	var answer interface{} = newFilesystem

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) DeleteFileSystem(ctx v4.Context, uid string, _ DeleteFileSystemParams) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	fs, exists := s.filesystems[uid]
	if exists {
		delete(s.filesystems, uid)
		delete(s.orgNames[*org.Name].fsNames, *fs.Name)
		return ctx.JSON(http.StatusOK, N200{
			Data: nil,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) GetFileSystem(ctx v4.Context, uid string, _ GetFileSystemParams) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	filesystem, exists := s.filesystems[uid]
	if exists {
		var answer interface{} = filesystem
		return ctx.JSON(http.StatusOK, N200{
			Data: &answer,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) UpdateFileSystem(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req UpdateFileSystemJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	filesystem, exists := s.filesystems[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}

	org, exists := getOrgFromToken(ctx, s.organizations)
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	_, exists = s.orgNames[*org.Name].fsNames[*req.NewName]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	if req.NewName != nil {
		filesystem.Name = req.NewName
	}
	if req.AuthRequired != nil {
		filesystem.AuthRequired = req.AuthRequired
	}
	if req.TotalCapacity != nil {
		filesystem.TotalBudget = req.TotalCapacity
	}
	if req.SsdCapacity != nil {
		filesystem.SsdBudget = req.SsdCapacity
	}

	s.orgNames[*org.Name].fsNames[*req.NewName] = uid
	delete(s.orgNames[*org.Name].fsNames, *filesystem.Name)
	var answer interface{} = filesystem

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetS3Buckets(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var answer interface{} = s.buckets

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) CreateS3Bucket(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req CreateS3BucketJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	var quota uint64
	if req.HardQuota != nil {
		quota, _ = strconv.ParseUint((*req.HardQuota)[:len(*req.HardQuota)-1], 10, 64)
	}

	var used uint64 = 0
	*s.buckets.Buckets = append(*s.buckets.Buckets,
		struct {
			HardLimitBytes *uint64 "json:\"hard_limit_bytes,omitempty\""
			Name           *string "json:\"name,omitempty\""
			Path           *string "json:\"path,omitempty\""
			UsedBytes      *uint64 "json:\"used_bytes,omitempty\""
		}{
			HardLimitBytes: &quota,
			UsedBytes:      &used,
			Name:           &req.BucketName,
			Path:           req.Policy, // reuse to store policy
		})

	return ctx.JSON(http.StatusOK, N200{
		Data: &bucket,
	})
}

func (s *SimulatedWekaAPI) DestroyS3Bucket(ctx v4.Context, bucketName string, _ DestroyS3BucketParams) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	newBuckets := make([]struct {
		HardLimitBytes *uint64 `json:"hard_limit_bytes,omitempty"`
		Name           *string `json:"name,omitempty"`
		Path           *string `json:"path,omitempty"`
		UsedBytes      *uint64 `json:"used_bytes,omitempty"`
	}, 0)

	for _, b := range *s.buckets.Buckets {
		if *b.Name != bucketName {
			newBuckets = append(newBuckets, b)
		}
	}
	s.buckets.Buckets = &newBuckets

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) S3ListAllLifecycleRules(ctx v4.Context, bucketName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	rule, exists := s.lifecycleRules[bucketName]
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N404{})
	}

	var answer interface{} = rule
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) S3CreateLifecycleRule(ctx v4.Context, bucketName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req S3CreateLifecycleRuleJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	rules, exists := s.lifecycleRules[bucketName]
	uid := genUID()
	rule := struct {
		Enabled    *bool   "json:\"enabled,omitempty\""
		ExpiryDays *string "json:\"expiry_days,omitempty\""
		Id         *string "json:\"id,omitempty\""
		Prefix     *string "json:\"prefix,omitempty\""
	}{
		ExpiryDays: &req.ExpiryDays,
		Id:         &uid,
		Prefix:     req.Prefix,
	}

	if exists {
		*rules.Rules = append(*rules.Rules, rule)
	} else {
		s.lifecycleRules[bucketName] = &S3LifecycleRule{
			Bucket: &bucketName,
			Rules: &[]struct {
				Enabled    *bool   "json:\"enabled,omitempty\""
				ExpiryDays *string "json:\"expiry_days,omitempty\""
				Id         *string "json:\"id,omitempty\""
				Prefix     *string "json:\"prefix,omitempty\""
			}{rule},
		}
	}

	var answer interface{} = rule

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) S3DeleteLifecycleRule(ctx v4.Context, bucketName string, lrId string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	rules, exists := s.lifecycleRules[bucketName]
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N404{})
	}

	newRules := make([]struct {
		Enabled    *bool   "json:\"enabled,omitempty\""
		ExpiryDays *string "json:\"expiry_days,omitempty\""
		Id         *string "json:\"id,omitempty\""
		Prefix     *string "json:\"prefix,omitempty\""
	}, 0)

	for _, lr := range *rules.Rules {
		if *lr.Id != lrId {
			newRules = append(newRules, lr)
		}
	}
	rules.Rules = &newRules

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) S3DeleteAllLifecycleRules(ctx v4.Context, bucketName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, exists := s.lifecycleRules[bucketName]
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N404{})
	}

	delete(s.lifecycleRules, bucketName)

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) GetS3BucketPolicy(ctx v4.Context, bucketName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	var answer interface{}

	for _, b := range *s.buckets.Buckets {
		if *b.Name == bucketName {
			answer = map[string]interface{}{
				"policy": b.Path,
			}
		}
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) SetS3BucketPolicy(ctx v4.Context, bucketName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req SetS3BucketPolicyJSONBody
	err := ctx.Bind(&req)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	for _, b := range *s.buckets.Buckets {
		if *b.Name == bucketName {
			*b.Path = req.BucketPolicy
		}
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) CreateS3Policy(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req CreateS3PolicyJSONBody
	err := ctx.Bind(&req)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	var answer interface{} = S3Policy{
		Policy: &struct {
			Content *S3IAMPolicy "json:\"content,omitempty\""
			Name    *string      "json:\"name,omitempty\""
		}{
			Content: &req.PolicyFileContent,
			Name:    &req.PolicyName,
		},
	}
	s.polices[req.PolicyName] = &req.PolicyFileContent

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) AttachS3Policy(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req AttachS3PolicyJSONBody
	err := ctx.Bind(&req)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	s.policyMapping[req.UserName] = req.PolicyName

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) DetachS3Policy(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req AttachS3PolicyJSONBody
	err := ctx.Bind(&req)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	delete(s.policyMapping, req.UserName)

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) DeleteS3Policy(ctx v4.Context, policyName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.policyMapping, policyName)

	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *SimulatedWekaAPI) GetS3Policy(ctx v4.Context, policyName string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	policy, exists := s.polices[policyName]
	if !exists {
		return ctx.JSON(http.StatusBadRequest, N404{})
	}

	var answer interface{} = S3Policy{
		Policy: &struct {
			Content *S3IAMPolicy "json:\"content,omitempty\""
			Name    *string      "json:\"name,omitempty\""
		}{
			Content: policy,
			Name:    &policyName,
		},
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func NewSimulatedWekaAPI(fa *auth.FakeAuthenticator) *SimulatedWekaAPI {
	users := make(map[string]*User)
	usernames := make(map[string]string)
	passwords := make(map[string]string)
	orgNames := make(map[string]*OrganizationContainer)
	containerHostNames := make(map[string]*ContainerInfo)
	organizations := make(map[string]*Organization)
	containers := make(map[string]*Container)
	processes := make(map[string]*Process)

	// Create cluster admin
	adminUser := createUser("admin", 0, "ClusterAdmin")
	users[*adminUser.Uid] = &adminUser
	passwords[*adminUser.Uid] = "adminPassword"

	// Create root Org
	var totalQuota uint64 = 100000
	rootOrg := createOrganization("Root", &totalQuota, &totalQuota)
	rootOrgUUID := "00000000-0000-0000-0000-000000000000"
	rootOrg.Uid = &rootOrgUUID
	organizations[*rootOrg.Uid] = &rootOrg

	users[*adminUser.Uid] = &adminUser
	passwords[*adminUser.Uid] = "adminPassword"
	usernames[*adminUser.Username] = *adminUser.Uid
	orgNames[*rootOrg.Name] = &OrganizationContainer{
		uid:       *rootOrg.Uid,
		fsNames:   make(map[string]string),
		usernames: usernames,
		passwords: passwords,
	}

	return &SimulatedWekaAPI{
		fa:                 fa,
		filesystems:        make(map[string]*FileSystem),
		fsGroups:           make(map[string]*FileSystemGroup),
		organizations:      organizations,
		containers:         containers,
		processes:          processes,
		users:              users,
		orgNames:           orgNames,
		containerHostNames: containerHostNames,
		fsGroupNames:       make(map[string]bool),
		buckets: &S3Bucket{
			Buckets: &[]struct {
				HardLimitBytes *uint64 "json:\"hard_limit_bytes,omitempty\""
				Name           *string "json:\"name,omitempty\""
				Path           *string "json:\"path,omitempty\""
				UsedBytes      *uint64 "json:\"used_bytes,omitempty\""
			}{},
		},
		lifecycleRules: make(map[string]*S3LifecycleRule),
		polices:        make(map[string]*S3IAMPolicy),
		policyMapping:  make(map[string]string),
		lock:           sync.Mutex{},
	}
}

func genUID() string {
	uid, err := uuid.NewUUID()
	if err != nil {
		panic("Cannot generate UUID")
	}
	return uid.String()
}

func createUser(name string, orgID int, role string) User {
	uid := genUID()
	source := "Internal"
	return User{
		Uid:      &uid,
		Username: &name,
		Role:     &role,
		Source:   &source,
		OrgId:    &orgID,
	}
}

func createOrganization(name string, totalQuota *uint64, ssdQuota *uint64) Organization {
	uid := genUID()
	id := rand.Int()
	return Organization{
		Uid:        &uid,
		Name:       &name,
		TotalQuota: totalQuota,
		SsdQuota:   ssdQuota,
		Id:         &id,
	}
}

func getOrgFromToken(ctx v4.Context, orgs map[string]*Organization) (*Organization, bool) {
	user := ctx.Get(auth.JWTClaimsContextKey).(jwt.Token)
	orgID, exists := user.Get(auth.OrganizationClaim)
	if !exists {
		return nil, false
	}

	org, exists := orgs[orgID.(string)]
	if !exists {
		return nil, false
	}

	return org, true
}

func (s *SimulatedWekaAPI) AddContainer(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var req AddContainerJSONBody
	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	_, exists := s.containerHostNames[req.ContainerName]
	if exists {
		return ctx.JSON(http.StatusBadRequest, N400{})
	}

	newContainer := addContainer(req.ContainerName, req.Ip, req.NoWait)
	s.containers[*newContainer.Uid] = &newContainer
	container := ContainerInfo{
		uid: *newContainer.Uid,
	}

	if newContainer.Cores == nil {
		return errors.New("container does not contain required \"Cores\" fields")
	}

	if newContainer.Hostname == nil {
		return errors.New("container does not contain required \"Hostname\" fields")
	}

	for i := 0; i < *newContainer.Cores+1; i++ {
		var role string
		if i == 0 {
			role = string(backend.ContainerRoleManagement)
		} else {
			role = string(backend.ContainerRoleFrontend)
		}
		newProcess := addProcess(*newContainer.Hostname, role)
		if newProcess.Hostname == nil {
			return errors.New("container process does not contain required \"hostname\" fields")
		}
		key := *newProcess.Hostname + fmt.Sprint(i)
		s.processes[key] = &newProcess
	}

	s.containerHostNames[req.ContainerName] = &container

	var answer interface{} = newContainer

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func addProcess(name string, role string) Process {
	uid := genUID()
	status := string(backend.ContainerStatusUP)
	mode := string(backend.ContainerModeClient)
	roles := []string{role}
	return Process{
		Uid:      &uid,
		Hostname: &name,
		Status:   &status,
		Roles:    &roles,
		Mode:     &mode,
	}
}

func addContainer(name string, ip *string, nowait *bool) Container {
	uid := genUID()
	status := string(backend.ContainerStatusUP)
	mode := string(backend.ContainerModeClient)
	cores := 1
	containerHostIps := []string{*ip}
	return Container{
		Uid:      &uid,
		Hostname: &name,
		Ips:      &containerHostIps,
		Status:   &status,
		Mode:     &mode,
		Cores:    &cores,
	}
}

func (s *SimulatedWekaAPI) GetSingleContainer(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	c, exists := s.containers[uid]
	if !exists {
		return ctx.JSON(http.StatusNotFound, N404{})
	}

	var answer interface{} = c

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetProcesses(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	ps := make([]Process, 0, len(s.processes))
	for _, val := range s.processes {
		ps = append(ps, *val)
	}
	var answer interface{} = ps
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) GetContainers(ctx v4.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	cs := make([]Container, 0, len(s.containers))
	for _, val := range s.containers {
		cs = append(cs, *val)
	}
	var answer interface{} = cs
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (s *SimulatedWekaAPI) DeactivateContainer(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, exists := s.containers[uid]
	if exists {
		status := "DOWN"
		s.containers[uid].Status = &status
		return ctx.JSON(http.StatusOK, N200{})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}

func (s *SimulatedWekaAPI) RemoveContainer(ctx v4.Context, uid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	cs, exists := s.containers[uid]
	if exists {
		for i := 0; i < *cs.Cores+1; i++ {
			key := *cs.Hostname + fmt.Sprint(i)
			delete(s.processes, key)
		}
		delete(s.containers, uid)
		delete(s.containerHostNames, *cs.Hostname)
		return ctx.JSON(http.StatusOK, N200{
			Data: nil,
		})
	}

	return ctx.JSON(http.StatusNotFound, N404{})
}
