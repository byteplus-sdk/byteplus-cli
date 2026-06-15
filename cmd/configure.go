/*
 * // Copyright (c) 2024 Bytedance Ltd. and/or its affiliates
 * //
 * // Licensed under the Apache License, Version 2.0 (the "License");
 * // you may not use this file except in compliance with the License.
 * // You may obtain a copy of the License at
 * //
 * //	http://www.apache.org/licenses/LICENSE-2.0
 * //
 * // Unless required by applicable law or agreed to in writing, software
 * // distributed under the License is distributed on an "AS IS" BASIS,
 * // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * // See the License for the specific language governing permissions and
 * // limitations under the License.
 */

package cmd

// Copyright 2023 Byteplus.  All Rights Reserved.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/byteplus-sdk/byteplus-cli/util"
)

var (
	configFileMu sync.Mutex
	// configFileDirFunc 是配置目录获取函数的注入点。
	// 生产环境固定使用 util.GetConfigFileDir；单测会替换为临时目录，避免读写真实 ~/.byteplus。
	configFileDirFunc = util.GetConfigFileDir
)

const (
	ModeSSO          = "sso"
	ModeAK           = "ak"
	ModeConsoleLogin = "console-login"
	ModeRamRoleArn   = "ramrolearn"
	ModeOIDC         = "oidc"
	ModeEcsRole      = "ecsrole"

	ConfigFile = "config.json"
)

type Configure struct {
	Current     string                 `json:"current"`
	Profiles    map[string]*Profile    `json:"profiles"`
	EnableColor bool                   `json:"enableColor"`
	SsoSession  map[string]*SsoSession `json:"sso-session"`
}

type Profile struct {
	Name             string `json:"name"`
	Mode             string `json:"mode"`
	AccessKey        string `json:"access-key"`
	SecretKey        string `json:"secret-key"`
	Region           string `json:"region"`
	Endpoint         string `json:"endpoint"`
	EndpointResolver string `json:"endpoint-resolver,omitempty"`
	UseDualStack     *bool  `json:"use-dual-stack,omitempty"`
	SessionToken     string `json:"session-token"`
	DisableSSL       *bool  `json:"disable-ssl"`
	SsoSessionName   string `json:"sso-session-name,omitempty"`
	AccountId        string `json:"account-id"`
	RoleName         string `json:"role-name"`
	StsExpiration    int64  `json:"sts-expiration"`
	OidcTokenFile    string `json:"oidc-token-file,omitempty"`
	RoleTrn          string `json:"role-trn,omitempty"`
	LoginSession     string `json:"login-session,omitempty"`
}

type SsoSession struct {
	Name               string   `json:"name"`
	StartURL           string   `json:"start-url"`
	Region             string   `json:"region"`
	RegistrationScopes []string `json:"registration-scopes,omitempty"`
}

// LoadConfig from CONFIG_FILE_DIR(default ~/.byteplus)
func LoadConfig() *Configure {
	configFileMu.Lock()
	defer configFileMu.Unlock()

	configFileDir, err := configFileDirFunc()
	if err != nil {
		return nil
	}

	if err := os.MkdirAll(configFileDir, 0700); err != nil {
		return nil
	}
	_ = os.Chmod(configFileDir, 0700)

	configFilePath := filepath.Join(configFileDir, ConfigFile)
	file, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()
	_ = file.Chmod(0600)

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		return nil
	}

	cfg := &Configure{}
	err = json.Unmarshal(fileContent, cfg)
	if err != nil {
		return nil
	}

	return cfg
}

// runtimeConfig returns the in-memory config used by the current CLI process.
func runtimeConfig() *Configure {
	if ctx != nil && ctx.config != nil {
		return ctx.config
	}
	return config
}

// setRuntimeConfig keeps the global config references in sync after updates.
func setRuntimeConfig(cfg *Configure) {
	if ctx != nil {
		ctx.config = cfg
	}
	config = cfg
}

// WriteConfigToFile store config
func WriteConfigToFile(config *Configure) error {
	configFileMu.Lock()
	defer configFileMu.Unlock()

	configFileDir, err := configFileDirFunc()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configFileDir, 0700); err != nil {
		return err
	}
	_ = os.Chmod(configFileDir, 0700)

	targetPath := filepath.Join(configFileDir, ConfigFile)

	dir := filepath.Dir(targetPath)
	tempFile, err := os.CreateTemp(dir, ".tmp-config-*")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempName)
	}()
	_ = tempFile.Chmod(0600)

	data, err := marshalConfig(config)
	if err != nil {
		return err
	}
	if _, err := tempFile.Write(data); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempName, targetPath); err != nil {
		_ = os.Remove(targetPath)
		if err2 := os.Rename(tempName, targetPath); err2 != nil {
			return err2
		}
	}
	_ = os.Chmod(targetPath, 0600)
	return nil
}

// marshalConfig 使用稳定缩进格式写出配置，便于用户排查 profile 与凭证链配置。
func marshalConfig(config *Configure) ([]byte, error) {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func (config *Configure) SetRandomCurrentProfile() {
	if config == nil {
		return
	}

	if config.Profiles == nil || len(config.Profiles) == 0 {
		config.Current = ""
		return
	}

	config.Current = ""
	for key := range config.Profiles {
		if config.Current == "" {
			config.Current = key
			break
		}
	}
}

func setConfigProfile(profile *Profile) error {
	var (
		exist          bool
		currentProfile *Profile
		cfg            *Configure
	)

	// if config not exist, create an empty config
	if cfg = ctx.config; cfg == nil {
		cfg = &Configure{
			Profiles: make(map[string]*Profile),
		}
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
	}

	// check if the target profileFlags already exists
	// otherwise create a new profileFlags
	if currentProfile, exist = cfg.Profiles[profile.Name]; !exist {
		currentProfile = &Profile{
			Name:         profile.Name,
			Mode:         ModeAK,
			DisableSSL:   new(bool),
			UseDualStack: new(bool),
		}
		*currentProfile.DisableSSL = false
		*currentProfile.UseDualStack = false
	}

	nextProfile := mergeProfile(currentProfile, profile)
	if err := validateProfileMode(nextProfile); err != nil {
		return err
	}

	cfg.Profiles[nextProfile.Name] = nextProfile
	cfg.Current = nextProfile.Name
	return WriteConfigToFile(cfg)
}

// mergeProfile 只合并用户显式传入的字段，避免局部更新 profile 时清空旧凭证或开关。
func mergeProfile(base *Profile, input *Profile) *Profile {
	merged := cloneProfile(base)
	if merged == nil {
		merged = &Profile{}
	}
	if input == nil {
		return merged
	}

	if input.Name != "" {
		merged.Name = input.Name
	}
	if input.AccessKey != "" {
		merged.AccessKey = input.AccessKey
	}
	if input.SecretKey != "" {
		merged.SecretKey = input.SecretKey
	}
	if input.Region != "" {
		merged.Region = input.Region
	}
	if input.Endpoint != "" {
		merged.Endpoint = input.Endpoint
	}
	if input.EndpointResolver != "" {
		merged.EndpointResolver = input.EndpointResolver
	}
	if input.SessionToken != "" {
		merged.SessionToken = input.SessionToken
	}
	if input.DisableSSL != nil {
		if merged.DisableSSL == nil {
			merged.DisableSSL = new(bool)
		}
		*merged.DisableSSL = *input.DisableSSL
	}
	if input.UseDualStack != nil {
		if merged.UseDualStack == nil {
			merged.UseDualStack = new(bool)
		}
		*merged.UseDualStack = *input.UseDualStack
	}
	if input.SsoSessionName != "" {
		merged.SsoSessionName = input.SsoSessionName
	}
	if input.AccountId != "" {
		merged.AccountId = input.AccountId
	}
	if input.RoleName != "" {
		merged.RoleName = input.RoleName
	}
	if input.OidcTokenFile != "" {
		merged.OidcTokenFile = input.OidcTokenFile
	}
	if input.RoleTrn != "" {
		merged.RoleTrn = input.RoleTrn
	}
	if input.Mode != "" {
		merged.Mode = input.Mode
	}
	if base == nil && merged.Mode == "" {
		merged.Mode = ModeAK
	}

	return merged
}

// cloneProfile 深拷贝含指针的 profile 字段，避免 merge 时意外修改调用方对象。
func cloneProfile(profile *Profile) *Profile {
	if profile == nil {
		return nil
	}
	clone := *profile
	if profile.DisableSSL != nil {
		clone.DisableSSL = new(bool)
		*clone.DisableSSL = *profile.DisableSSL
	}
	if profile.UseDualStack != nil {
		clone.UseDualStack = new(bool)
		*clone.UseDualStack = *profile.UseDualStack
	}
	return &clone
}

func getConfigProfile(profileName string) error {
	var (
		exist          bool
		currentProfile *Profile
		cfg            *Configure
	)

	// if config not exist, return
	if cfg = ctx.config; cfg == nil {
		fmt.Println("no profile created")
		return nil
	}

	if profileName == "" {
		fmt.Printf("no profile name specified, show current profile: [%v]\n", cfg.Current)
		profileName = cfg.Current
	}

	// check if the target profile already exists, otherwise print an empty profileFlags
	if currentProfile, exist = cfg.Profiles[profileName]; !exist || currentProfile == nil {
		currentProfile = &Profile{}
	}

	if config == nil || !config.EnableColor {
		util.ShowJson(currentProfile.ToMap(), false)
	} else {
		util.ShowJson(currentProfile.ToMap(), true)
	}
	return nil
}

func listConfigProfiles() error {
	var (
		cfg *Configure
	)

	// if config not exist, return
	if cfg = ctx.config; cfg == nil {
		fmt.Println("no profile created")
		return nil
	}

	fmt.Printf("*** current profile: %v ***\n", ctx.config.Current)
	for _, profile := range ctx.config.Profiles {
		util.ShowJson(profile.ToMap(), config.EnableColor)
	}
	return nil
}

func deleteConfigProfile(profileName string) error {
	var (
		exist bool
		cfg   *Configure
	)

	// if config not exist, return error
	if cfg = ctx.config; cfg == nil {
		return fmt.Errorf("configuration profile %v not found", profileName)
	}

	// check if the target profileFlags exists
	if _, exist = cfg.Profiles[profileName]; !exist {
		return fmt.Errorf("configuration profile %v not found", profileName)
	}

	// delete profileFlags and write change to config file
	delete(cfg.Profiles, profileName)
	if profileName == cfg.Current {
		cfg.SetRandomCurrentProfile()
		fmt.Printf("delete current profile, set new current profile to [%v]\n", cfg.Current)
	}

	return WriteConfigToFile(cfg)
}

func changeConfigProfile(profileName string) error {
	var (
		exist bool
		cfg   *Configure
	)

	// if config not exist, return error
	if cfg = ctx.config; cfg == nil {
		return fmt.Errorf("configuration profile %v not found", profileName)
	}

	// check if the target profileFlags exists
	if _, exist = cfg.Profiles[profileName]; !exist {
		return fmt.Errorf("configuration profile %v not found", profileName)
	}

	// if not change,skip it
	if profileName == cfg.Current {
		return nil
	}

	// change current
	cfg.Current = profileName
	return WriteConfigToFile(cfg)
}

func (p *Profile) ToMap() map[string]interface{} {
	data, _ := json.Marshal(p)
	m := make(map[string]interface{})
	json.Unmarshal(data, &m)

	return m
}

func (p *Profile) String() string {
	b, _ := json.MarshalIndent(p, "", "    ")
	return string(b)
}

func setSsoSession(session *SsoSession) error {
	var (
		cfg *Configure
	)
	scopes, err := normalizeRegistrationScopes(session.RegistrationScopes)
	if err != nil {
		return err
	}

	if cfg = ctx.config; cfg == nil {
		cfg = &Configure{
			Profiles:   make(map[string]*Profile),
			SsoSession: make(map[string]*SsoSession),
		}
	}

	if cfg.SsoSession == nil {
		cfg.SsoSession = make(map[string]*SsoSession)
	}

	newSession := &SsoSession{
		Name:               session.Name,
		StartURL:           session.StartURL,
		Region:             session.Region,
		RegistrationScopes: scopes,
	}

	cfg.SsoSession[session.Name] = newSession

	return WriteConfigToFile(cfg)
}
