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

var configFileMu sync.Mutex

const (
	ModeSSO = "sso"
	ModeAK  = "ak"

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
	AccountId        string `json:"account-id,omitempty"`
	RoleName         string `json:"role-name,omitempty"`
	StsExpiration    int64  `json:"sts-expiration,omitempty"`
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

	configFileDir, err := util.GetConfigFileDir()
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

// WriteConfigToFile store config
func WriteConfigToFile(config *Configure) error {
	configFileMu.Lock()
	defer configFileMu.Unlock()

	configFileDir, err := util.GetConfigFileDir()
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

	if err := json.NewEncoder(tempFile).Encode(config); err != nil {
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

	// check if the target profileFlags already exists
	// otherwise create a new profileFlags
	if currentProfile, exist = cfg.Profiles[profile.Name]; !exist {
		currentProfile = &Profile{
			Name:         profile.Name,
			Mode:         "AK",
			DisableSSL:   new(bool),
			UseDualStack: new(bool),
		}
		*currentProfile.DisableSSL = false
		*currentProfile.UseDualStack = false
	}

	if profile.AccessKey != "" {
		currentProfile.AccessKey = profile.AccessKey
	}
	if profile.SecretKey != "" {
		currentProfile.SecretKey = profile.SecretKey
	}
	if profile.Region != "" {
		currentProfile.Region = profile.Region
	}
	if profile.Endpoint != "" {
		currentProfile.Endpoint = profile.Endpoint
	}
	if profile.EndpointResolver != "" {
		currentProfile.EndpointResolver = profile.EndpointResolver
	}
	if profile.SessionToken != "" {
		currentProfile.SessionToken = profile.SessionToken
	}
	if profile.DisableSSL != nil {
		*currentProfile.DisableSSL = *profile.DisableSSL
	}
	if profile.UseDualStack != nil {
		if currentProfile.UseDualStack == nil {
			currentProfile.UseDualStack = new(bool)
		}
		*currentProfile.UseDualStack = *profile.UseDualStack
	}
	if profile.SsoSessionName != "" {
		currentProfile.SsoSessionName = profile.SsoSessionName
	}

	cfg.Profiles[currentProfile.Name] = currentProfile
	cfg.Current = currentProfile.Name
	return WriteConfigToFile(cfg)
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
