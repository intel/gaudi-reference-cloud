// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// Keys for client provided userdata map access
const (
	APT                      string = "apt"
	FQDN                     string = "fqdn"
	HOSTNAME                 string = "hostname"
	MANAGE_ETC_HOSTS         string = "manage_etc_hosts"
	PACKAGES                 string = "packages"
	PACKAGE_UPDATE           string = "package_update"
	RUNCMD                   string = "runcmd"
	SSH_AUTHORIZED_KEYS      string = "ssh_authorized_keys"
	USERS                    string = "users"
	WRITE_FILES              string = "write_files"
	USER_NAME                string = "name"
	USER_GROUPS              string = "groups"
	USER_SHELL               string = "shell"
	USER_SUDO                string = "sudo"
	USER_SSH_AUTHORIZED_KEYS string = "ssh_authorized_keys"

	defaultOS    string = "ubuntu"
	defaultShell string = "/bin/bash"
	defaultSudo  string = "ALL=(ALL) NOPASSWD:ALL"

	defaultPowerStateDelay string = "now"
	defaultPowerStateMode  string = "reboot"
	defaultUserGroups      string = "docker,render"
)

type WriteFile struct {
	Content     string `yaml:"content"`
	Path        string `yaml:"path"`
	Append      bool   `yaml:"append,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Encoding    string `yaml:"encoding,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Defer       bool   `yaml:"defer,omitempty"`
}

type PowerState struct {
	Delay   string `yaml:"delay"`
	Mode    string `yaml:"mode"`
	Message string `yaml:"message,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

// Convert between Protobuf messages.
type CloudConfig struct {
	cloudConfig cloudConfig
	userDataMap map[string]interface{}
	writeFiles  []WriteFile
	runCmds     []string
	packages    []string
}

type User struct {
	Name              string   `yaml:"name"`
	Groups            string   `yaml:"groups"`
	Shell             string   `yaml:"shell,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys" mapstructure:"ssh_authorized_keys"`
	Sudo              string   `yaml:"sudo,omitempty"`
}

type cloudConfig interface {
	AddUser(user *User)
	SetAttr(key string, value interface{})
	SetSSHAuthorizedKeys(keys string)
	SetSystemUpdate(update bool)
	RenderYAML() ([]byte, error)
}

// userYAML exists to add additional marshalled fields (lock_passwd) and to ensure the ordering matches
// the previous implementation
type userYAML struct {
	Groups            []string `yaml:"groups"`
	LockPassword      bool     `yaml:"lock_passwd,omitempty"`
	Name              string   `yaml:"name"`
	Shell             string   `yaml:"shell,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys" mapstructure:"ssh_authorized_keys"`
	Sudo              string   `yaml:"sudo,omitempty"`
}

type cloudConfigYAML struct {
	config map[string]interface{}
}

func newCloudConfigYAML(osname string) cloudConfig {
	return &cloudConfigYAML{
		config: make(map[string]interface{}),
	}
}

func (c *cloudConfigYAML) AddUser(user *User) {
	yaml := userYAML{
		Groups:            strings.Split(user.Groups, ","),
		LockPassword:      true,
		Name:              user.Name,
		Shell:             user.Shell,
		SSHAuthorizedKeys: user.SSHAuthorizedKeys,
		Sudo:              user.Sudo,
	}
	if c.config[USERS] == nil {
		c.config[USERS] = make([]userYAML, 0)
	}
	users := c.config[USERS].([]userYAML)
	for i, u := range users {
		if u.Name == yaml.Name {
			users[i] = yaml
			return
		}
	}
	c.config[USERS] = append(users, yaml)
}

func (c *cloudConfigYAML) SetAttr(key string, value interface{}) {
	c.config[key] = value
}

func (c *cloudConfigYAML) SetSSHAuthorizedKeys(keys string) {
	// TODO Investigate why yaml.Marshal as used in RenderYAML() does not remove empty strings
	ss := make([]string, 0, len(keys))
	for _, s := range strings.Split(keys, "\n") {
		if s != "" {
			ss = append(ss, s)
		}
	}
	c.config[SSH_AUTHORIZED_KEYS] = ss
}

func (c *cloudConfigYAML) SetSystemUpdate(update bool) {
	c.SetAttr(PACKAGE_UPDATE, update)
}

func (c *cloudConfigYAML) RenderYAML() ([]byte, error) {
	var bs bytes.Buffer
	w := io.Writer(&bs)
	if _, err := fmt.Fprintf(w, "#cloud-config\n"); err != nil {
		return nil, err
	}

	// The previous implementation rendered the keys in sorted order
	// Preserve that behavior to avoid any unforseen problems
	sortedKeys := make([]string, 0, len(c.config))
	for k := range c.config {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		item := map[string]interface{}{k: c.config[k]}
		b, err := yaml.Marshal(item)
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintf(w, "%s", string(b)); err != nil {
			return nil, err
		}
	}
	return bs.Bytes(), nil
}

// Parse the users list from cloudinit yaml
func parseUsers(config cloudConfig, users []interface{}) error {

	var ssh_keys string
	var newUser User
	for _, userInt := range users {
		item := userInt.(map[interface{}]interface{})
		err := mapstructure.Decode(item, &newUser)
		if err != nil {
			return err
		}
		// SSH auth key string
		for _, v := range newUser.SSHAuthorizedKeys {
			ssh_keys += v + "\n"
		}
		config.AddUser(&newUser)
	}
	return nil
}

// Parse the write_files list instance userData
func parseWriteFiles(write_files []interface{}) ([]WriteFile, error) {

	var writeFiles []WriteFile
	var newFile WriteFile

	for _, wFilesInt := range write_files {
		item := wFilesInt.(map[interface{}]interface{})
		err := mapstructure.Decode(item, &newFile)
		if err != nil {
			return nil, err
		}
		writeFiles = append(writeFiles, newFile)
	}
	return writeFiles, nil
}

func parseRunCmds(commands []interface{}) ([]string, error) {

	var runCmds []string
	var runCmd string

	for _, cmdInt := range commands {
		item := cmdInt.(interface{})
		err := mapstructure.Decode(item, &runCmd)
		if err != nil {
			return nil, err
		}
		runCmds = append(runCmds, runCmd)
	}
	return runCmds, nil
}

func NewEmptyCloudConfig(osname string) (*CloudConfig, error) {
	newConfig := newCloudConfigYAML(defaultOS)
	return &CloudConfig{
		cloudConfig: newConfig,
		userDataMap: map[string]interface{}{},
		writeFiles:  []WriteFile{},
		runCmds:     []string{},
		packages:    []string{},
	}, nil
}

func NewCloudConfig(osname string, userData string) (*CloudConfig, error) {
	var writeFiles []WriteFile
	var runCmds []string
	// convert userData to userDataMap
	cfg := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(userData), &cfg)
	if err != nil {
		return nil, err
	}

	newConfig := newCloudConfigYAML(defaultOS)

	// default fields
	newConfig.SetSystemUpdate(false)

	// override default values with client userData
	for key, value := range cfg {
		if key == USERS {
			err = parseUsers(newConfig, value.([]interface{}))
			if err != nil {
				return nil, err
			}
		} else if key == WRITE_FILES {
			writeFiles, err = parseWriteFiles(value.([]interface{}))
			if err != nil {
				return nil, err
			}
		} else if key == RUNCMD {
			runCmds, err = parseRunCmds(value.([]interface{}))
			if err != nil {
				return nil, err
			}
		} else {
			newConfig.SetAttr(key, value)
		}
	}

	return &CloudConfig{
		cloudConfig: newConfig,
		userDataMap: cfg,
		writeFiles:  writeFiles,
		runCmds:     runCmds,
		packages:    []string{},
	}, nil
}

func (c *CloudConfig) SetManageEtcHosts(mngEtcHosts string) {
	c.cloudConfig.SetAttr("manage_etc_hosts", mngEtcHosts)
}

func (c *CloudConfig) SetHostName(hostname string) {
	c.cloudConfig.SetAttr("hostname", hostname)
}

func (c *CloudConfig) SetFqdn(fqdn string) {
	c.cloudConfig.SetAttr("fqdn", fqdn)
}

func (c *CloudConfig) SetSshPublicKeys(ssh_public_keys []string) {

	// prepare user provided keys
	var user_keys string
	if _, exists := c.userDataMap[SSH_AUTHORIZED_KEYS]; exists {
		keys := c.userDataMap[SSH_AUTHORIZED_KEYS].([]interface{})
		for _, key := range keys {
			user_keys += key.(string) + "\n"
		}
	}

	// prepare system keys
	system_keys := strings.Join(ssh_public_keys, "\n") + "\n"

	// append both sets in the cloud config
	c.cloudConfig.SetSSHAuthorizedKeys(system_keys + user_keys)
}

func (c *CloudConfig) SetDefaultUserGroup(name string, ssh_authorized_keys []string) {
	c.cloudConfig.AddUser(&User{
		Name:              name,
		Groups:            defaultUserGroups,
		Shell:             defaultShell,
		SSHAuthorizedKeys: ssh_authorized_keys,
		Sudo:              defaultSudo,
	})
}

func (c *CloudConfig) AddWriteFile(path string, content string) {
	new_file := WriteFile{
		Path:    path,
		Content: content,
	}
	c.writeFiles = append(c.writeFiles, new_file)
}

func (c *CloudConfig) AddWriteFileWithPermissions(path string, content string, permissions string) {
	new_file := WriteFile{
		Path:        path,
		Content:     content,
		Permissions: permissions,
	}
	c.writeFiles = append(c.writeFiles, new_file)
}

func (c *CloudConfig) AddRunBinaryFile(path string, content []byte, permissions uint) {
	c.AddRunCmd(fmt.Sprintf("install -D -m %o /dev/null '%s'", permissions, path))
	c.AddRunCmd(fmt.Sprintf("echo -n %s | base64 -d > '%s'", base64.StdEncoding.EncodeToString(content), path))
}

func (c *CloudConfig) AddRunCmd(commands ...string) {
	c.runCmds = append(c.runCmds, commands...)
}

func (c *CloudConfig) SetRunCmd() {
	if len(c.runCmds) > 0 {
		c.cloudConfig.SetAttr(RUNCMD, c.runCmds)
	}
}

func (c *CloudConfig) AddPackage(packages ...string) {
	c.packages = append(c.packages, packages...)
}

func (c *CloudConfig) SetPackages() {
	if len(c.packages) > 0 {
		c.cloudConfig.SetAttr(PACKAGES, c.packages)
	}
}

func (c *CloudConfig) SetWriteFile() {
	if len(c.writeFiles) > 0 {
		c.cloudConfig.SetAttr(WRITE_FILES, c.writeFiles)
	}
}

func (c *CloudConfig) SetDefaultPowerState() {
	new_power_state := PowerState{
		Delay: defaultPowerStateDelay,
		Mode:  defaultPowerStateMode,
	}
	c.cloudConfig.SetAttr("power_state", new_power_state)
}

func (c *CloudConfig) RenderYAML() ([]byte, error) {
	// render CloudConfig to yaml format
	data, err := c.cloudConfig.RenderYAML()
	if err != nil {
		return nil, err
	}
	return data, nil
}
