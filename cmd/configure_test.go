package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/credentials/clicreds"
)

func resetProfileFlagsForTest(t *testing.T) {
	t.Helper()
	old := profileFlags
	profileFlags = Profile{}
	t.Cleanup(func() {
		profileFlags = old
	})
}

func withTestCtxConfig(t *testing.T, cfg *Configure) {
	t.Helper()
	oldCtx := ctx
	oldConfig := config
	ctx = NewContext()
	ctx.SetConfig(cfg)
	config = cfg
	t.Cleanup(func() {
		ctx = oldCtx
		config = oldConfig
	})
}

func readConfigFileAsMap(t *testing.T, dir string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ConfigFile))
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal config file: %v", err)
	}
	return out
}

func writeTestConfig(t *testing.T, cfg *Configure) string {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ConfigFile)
	data, err := marshalConfig(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
}

func TestMarshalConfigUsesIndentedJSON(t *testing.T) {
	data, err := marshalConfig(&Configure{
		Current: "test",
		Profiles: map[string]*Profile{
			"test": {
				Name:      "test",
				Mode:      ModeAK,
				Region:    "ap-southeast-1",
				AccessKey: "ak",
				SecretKey: "sk",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshalConfig returned error: %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("marshalConfig returned invalid json: %s", string(data))
	}
	if !strings.Contains(string(data), "\n    \"profiles\":") {
		t.Fatalf("marshalConfig should indent top-level fields, got: %s", string(data))
	}
}

func TestMarshalConfigKeepsCredentialProfileShapeAlignedWithSDK(t *testing.T) {
	data, err := marshalConfig(&Configure{
		Current: "test",
		Profiles: map[string]*Profile{
			"test": {
				Name:      "test",
				Mode:      ModeAK,
				AccessKey: "ak",
				SecretKey: "sk",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshalConfig returned error: %v", err)
	}
	text := string(data)
	for _, key := range []string{`"account-id": ""`, `"role-name": ""`, `"sts-expiration": 0`} {
		if !strings.Contains(text, key) {
			t.Fatalf("marshalConfig should keep %s in profile json, got: %s", key, text)
		}
	}
}

func TestConfigureSetSupportsOIDCModeFields(t *testing.T) {
	dir := withTestConfigDir(t)
	resetProfileFlagsForTest(t)
	withTestCtxConfig(t, &Configure{Profiles: map[string]*Profile{}})

	setCmd := newConfigureSetCmd()
	setCmd.SetArgs([]string{
		"--profile", "oidc",
		"--mode", "oidc",
		"--region", "ap-southeast-1",
		"--oidc-token-file", "/var/run/secrets/oidc-token",
		"--role-trn", "trn:iam::2100000000:role/CIRole",
	})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("configure set oidc mode returned error: %v", err)
	}

	raw := readConfigFileAsMap(t, dir)
	profiles := raw["profiles"].(map[string]interface{})
	profile := profiles["oidc"].(map[string]interface{})
	if profile["mode"] != "oidc" {
		t.Fatalf("mode = %v, want oidc", profile["mode"])
	}
	if profile["oidc-token-file"] != "/var/run/secrets/oidc-token" {
		t.Fatalf("oidc-token-file = %v", profile["oidc-token-file"])
	}
	if profile["role-trn"] != "trn:iam::2100000000:role/CIRole" {
		t.Fatalf("role-trn = %v", profile["role-trn"])
	}
}

func TestConfigureSetSupportsRamRoleArnModeFields(t *testing.T) {
	dir := withTestConfigDir(t)
	resetProfileFlagsForTest(t)
	withTestCtxConfig(t, &Configure{Profiles: map[string]*Profile{}})

	setCmd := newConfigureSetCmd()
	setCmd.SetArgs([]string{
		"--profile", "ram",
		"--mode", "ramrolearn",
		"--region", "ap-southeast-1",
		"--access-key", "ak",
		"--secret-key", "sk",
		"--account-id", "2100000000",
		"--role-name", "AdminRole",
	})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("configure set ramrolearn mode returned error: %v", err)
	}

	raw := readConfigFileAsMap(t, dir)
	profiles := raw["profiles"].(map[string]interface{})
	profile := profiles["ram"].(map[string]interface{})
	if profile["mode"] != "ramrolearn" {
		t.Fatalf("mode = %v, want ramrolearn", profile["mode"])
	}
	if profile["account-id"] != "2100000000" {
		t.Fatalf("account-id = %v, want 2100000000", profile["account-id"])
	}
	if profile["role-name"] != "AdminRole" {
		t.Fatalf("role-name = %v, want AdminRole", profile["role-name"])
	}
}

func TestConfigureSetSupportsEcsRoleModeFields(t *testing.T) {
	dir := withTestConfigDir(t)
	resetProfileFlagsForTest(t)
	withTestCtxConfig(t, &Configure{Profiles: map[string]*Profile{}})

	setCmd := newConfigureSetCmd()
	setCmd.SetArgs([]string{
		"--profile", "ecs",
		"--mode", "ecsrole",
		"--region", "ap-southeast-1",
		"--role-name", "EcsRole",
	})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("configure set ecsrole mode returned error: %v", err)
	}

	raw := readConfigFileAsMap(t, dir)
	profiles := raw["profiles"].(map[string]interface{})
	profile := profiles["ecs"].(map[string]interface{})
	if profile["mode"] != "ecsrole" {
		t.Fatalf("mode = %v, want ecsrole", profile["mode"])
	}
	if profile["role-name"] != "EcsRole" {
		t.Fatalf("role-name = %v, want EcsRole", profile["role-name"])
	}
}

func TestConfigureSetPreservesPointerFlagsWhenNotPassed(t *testing.T) {
	withTestConfigDir(t)
	resetProfileFlagsForTest(t)

	trueVal := true
	withTestCtxConfig(t, &Configure{
		Current: "p1",
		Profiles: map[string]*Profile{
			"p1": {
				Name:         "p1",
				Mode:         ModeAK,
				AccessKey:    "old-ak",
				SecretKey:    "old-sk",
				Region:       "cn-beijing",
				DisableSSL:   &trueVal,
				UseDualStack: &trueVal,
			},
		},
	})

	setCmd := newConfigureSetCmd()
	setCmd.SetArgs([]string{"--profile", "p1", "--region", "cn-shanghai"})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("configure set returned error: %v", err)
	}

	cfg := LoadConfig()
	profile := cfg.Profiles["p1"]
	if profile.Region != "cn-shanghai" {
		t.Fatalf("region = %q, want cn-shanghai", profile.Region)
	}
	if profile.DisableSSL == nil || !*profile.DisableSSL {
		t.Fatalf("DisableSSL should remain true when --disable-ssl is not passed, got %v", profile.DisableSSL)
	}
	if profile.UseDualStack == nil || !*profile.UseDualStack {
		t.Fatalf("UseDualStack should remain true when --use-dual-stack is not passed, got %v", profile.UseDualStack)
	}
}

func TestConfigureSetExplicitFalseOverridesPointerFlags(t *testing.T) {
	withTestConfigDir(t)
	resetProfileFlagsForTest(t)

	trueVal := true
	withTestCtxConfig(t, &Configure{
		Current: "p1",
		Profiles: map[string]*Profile{
			"p1": {
				Name:         "p1",
				Mode:         ModeAK,
				AccessKey:    "old-ak",
				SecretKey:    "old-sk",
				Region:       "ap-southeast-1",
				DisableSSL:   &trueVal,
				UseDualStack: &trueVal,
			},
		},
	})

	setCmd := newConfigureSetCmd()
	setCmd.SetArgs([]string{"--profile", "p1", "--disable-ssl=false", "--use-dual-stack=false"})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("configure set returned error: %v", err)
	}

	cfg := LoadConfig()
	profile := cfg.Profiles["p1"]
	if profile.DisableSSL == nil || *profile.DisableSSL {
		t.Fatalf("DisableSSL should be explicit false, got %v", profile.DisableSSL)
	}
	if profile.UseDualStack == nil || *profile.UseDualStack {
		t.Fatalf("UseDualStack should be explicit false, got %v", profile.UseDualStack)
	}
}

func TestConfigureSetInitializesPointerFlagsForNewProfile(t *testing.T) {
	withTestConfigDir(t)
	resetProfileFlagsForTest(t)
	withTestCtxConfig(t, &Configure{Profiles: map[string]*Profile{}})

	setCmd := newConfigureSetCmd()
	setCmd.SetArgs([]string{
		"--profile", "fresh",
		"--access-key", "ak",
		"--secret-key", "sk",
	})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("configure set returned error: %v", err)
	}

	cfg := LoadConfig()
	profile := cfg.Profiles["fresh"]
	if profile.Mode != ModeAK {
		t.Fatalf("mode = %q, want %q", profile.Mode, ModeAK)
	}
	if profile.DisableSSL == nil || *profile.DisableSSL {
		t.Fatalf("DisableSSL should be non-nil false, got %v", profile.DisableSSL)
	}
	if profile.UseDualStack == nil || *profile.UseDualStack {
		t.Fatalf("UseDualStack should be non-nil false, got %v", profile.UseDualStack)
	}
}

func TestValidateProfileModeRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
		wantErr string
	}{
		{
			name:    "invalid mode",
			profile: &Profile{Name: "p", Mode: "unknown"},
			wantErr: "unsupported mode",
		},
		{
			name:    "ak missing secret",
			profile: &Profile{Name: "p", Mode: ModeAK, AccessKey: "ak"},
			wantErr: "--secret-key",
		},
		{
			name:    "ram role arn missing account",
			profile: &Profile{Name: "p", Mode: ModeRamRoleArn, AccessKey: "ak", SecretKey: "sk", RoleName: "role"},
			wantErr: "--account-id",
		},
		{
			name:    "oidc missing token file",
			profile: &Profile{Name: "p", Mode: ModeOIDC, RoleTrn: "trn:iam::123:role/r"},
			wantErr: "--oidc-token-file",
		},
		{
			name:    "ecs role missing role name",
			profile: &Profile{Name: "p", Mode: ModeEcsRole},
			wantErr: "--role-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfileMode(tt.profile)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateProfileModeAcceptsValidModes(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
	}{
		{
			name:    "ak",
			profile: &Profile{Name: "p", Mode: ModeAK, AccessKey: "ak", SecretKey: "sk"},
		},
		{
			name:    "sso",
			profile: &Profile{Name: "p", Mode: ModeSSO},
		},
		{
			name:    "console-login",
			profile: &Profile{Name: "p", Mode: ModeConsoleLogin, LoginSession: "login-session"},
		},
		{
			name:    "ramrolearn",
			profile: &Profile{Name: "p", Mode: ModeRamRoleArn, AccessKey: "ak", SecretKey: "sk", RoleName: "role", AccountId: "2100000000"},
		},
		{
			name:    "oidc",
			profile: &Profile{Name: "p", Mode: ModeOIDC, OidcTokenFile: "/tmp/token", RoleTrn: "trn:iam::2100000000:role/CIRole"},
		},
		{
			name:    "ecsrole",
			profile: &Profile{Name: "p", Mode: ModeEcsRole, RoleName: "role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateProfileMode(tt.profile); err != nil {
				t.Fatalf("validateProfileMode returned error: %v", err)
			}
		})
	}
}

func TestValidateProfileModeRequiresOIDCFieldsFromProfile(t *testing.T) {
	t.Setenv("BYTEPLUS_OIDC_TOKEN_FILE", "/tmp/token")
	t.Setenv("BYTEPLUS_OIDC_ROLE_TRN", "trn:iam::2100000000:role/CIRole")

	err := validateProfileMode(&Profile{Name: "p", Mode: ModeOIDC})
	if err == nil {
		t.Fatal("expected oidc mode to require profile fields even when env vars exist")
	}
}

func TestMergeProfilePreservesAndOverridesPointerBools(t *testing.T) {
	trueVal := true
	falseVal := false
	base := &Profile{
		Name:         "p1",
		Mode:         ModeAK,
		AccessKey:    "old-ak",
		SecretKey:    "old-sk",
		Region:       "cn-beijing",
		DisableSSL:   &trueVal,
		UseDualStack: &trueVal,
	}

	merged := mergeProfile(base, &Profile{
		Name:       "p1",
		Region:     "ap-southeast-1",
		DisableSSL: &falseVal,
	})

	if merged.Region != "ap-southeast-1" {
		t.Fatalf("region = %q, want ap-southeast-1", merged.Region)
	}
	if merged.DisableSSL == nil || *merged.DisableSSL {
		t.Fatalf("DisableSSL = %v, want explicit false override", merged.DisableSSL)
	}
	if merged.UseDualStack == nil || !*merged.UseDualStack {
		t.Fatalf("UseDualStack = %v, want preserved true", merged.UseDualStack)
	}
	if base.DisableSSL == nil || !*base.DisableSSL {
		t.Fatalf("base DisableSSL was mutated, got %v", base.DisableSSL)
	}

	*merged.UseDualStack = false
	if base.UseDualStack == nil || !*base.UseDualStack {
		t.Fatalf("base UseDualStack should not share pointer with merged profile")
	}
}

func TestMergeProfileDefaultsNewProfileToAK(t *testing.T) {
	merged := mergeProfile(nil, &Profile{
		Name:      "p1",
		AccessKey: "ak",
		SecretKey: "sk",
	})

	if merged.Mode != ModeAK {
		t.Fatalf("mode = %q, want %q", merged.Mode, ModeAK)
	}
}

func TestMergeProfilePreservesNonAKModeOnUpdate(t *testing.T) {
	falseVal := false
	base := &Profile{
		Name:       "ecs",
		Mode:       ModeEcsRole,
		RoleName:   "old-role",
		Region:     "ap-southeast-1",
		DisableSSL: &falseVal,
	}

	merged := mergeProfile(base, &Profile{Name: "ecs", Region: "cn-beijing"})

	if merged.Mode != ModeEcsRole {
		t.Fatalf("mode = %q, want %q", merged.Mode, ModeEcsRole)
	}
	if merged.RoleName != "old-role" {
		t.Fatalf("role-name = %q, want old-role", merged.RoleName)
	}
	if merged.Region != "cn-beijing" {
		t.Fatalf("region = %q, want cn-beijing", merged.Region)
	}
}

func TestMergeProfileNilInputReturnsClone(t *testing.T) {
	falseVal := false
	base := &Profile{
		Name:       "p1",
		Mode:       ModeAK,
		AccessKey:  "ak",
		SecretKey:  "sk",
		DisableSSL: &falseVal,
	}

	merged := mergeProfile(base, nil)
	*merged.DisableSSL = true

	if base.DisableSSL == nil || *base.DisableSSL {
		t.Fatalf("base DisableSSL should not be mutated, got %v", base.DisableSSL)
	}
}

func TestNewSimpleClientProfileOverrideNotFound(t *testing.T) {
	falseVal := false
	testCtx := NewContext()
	testCtx.SetConfig(&Configure{
		Current: "default",
		Profiles: map[string]*Profile{
			"default": {
				Name:       "default",
				Mode:       ModeAK,
				AccessKey:  "ak",
				SecretKey:  "sk",
				Region:     "cn-beijing",
				DisableSSL: &falseVal,
			},
		},
	})
	flag, _ := testCtx.fixedFlags.AddByName("profile")
	flag.SetValue("missing")

	_, err := NewSimpleClient(testCtx)
	if err == nil {
		t.Fatal("expected error for missing ---profile override")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %q, want profile not found", err.Error())
	}
}

func TestNewSimpleClientProfileOverrideValid(t *testing.T) {
	falseVal := false
	testCtx := NewContext()
	testCtx.SetConfig(&Configure{
		Current: "default",
		Profiles: map[string]*Profile{
			"default": {
				Name:       "default",
				Mode:       ModeAK,
				AccessKey:  "default-ak",
				SecretKey:  "default-sk",
				Region:     "ap-southeast-1",
				DisableSSL: &falseVal,
			},
			"prod": {
				Name:       "prod",
				Mode:       ModeAK,
				AccessKey:  "prod-ak",
				SecretKey:  "prod-sk",
				Region:     "cn-beijing",
				DisableSSL: &falseVal,
			},
		},
	})
	flag, _ := testCtx.fixedFlags.AddByName("profile")
	flag.SetValue("prod")

	client, err := NewSimpleClient(testCtx)
	if err != nil {
		t.Fatalf("NewSimpleClient returned error: %v", err)
	}
	if client == nil || client.Config == nil || client.Config.Region == nil || *client.Config.Region != "cn-beijing" {
		t.Fatalf("region = %v, want cn-beijing", client.Config.Region)
	}
}

func TestNewSimpleClientRegionOverride(t *testing.T) {
	falseVal := false
	testCtx := NewContext()
	testCtx.SetConfig(&Configure{
		Current: "default",
		Profiles: map[string]*Profile{
			"default": {
				Name:       "default",
				Mode:       ModeAK,
				AccessKey:  "ak",
				SecretKey:  "sk",
				Region:     "ap-southeast-1",
				DisableSSL: &falseVal,
			},
		},
	})
	flag, _ := testCtx.fixedFlags.AddByName("region")
	flag.SetValue("cn-beijing")

	client, err := NewSimpleClient(testCtx)
	if err != nil {
		t.Fatalf("NewSimpleClient returned error: %v", err)
	}
	if client == nil || client.Config == nil || client.Config.Region == nil || *client.Config.Region != "cn-beijing" {
		t.Fatalf("region = %v, want cn-beijing", client.Config.Region)
	}
}

func TestNewSimpleClientRegionOverrideFixesEmptyProfileRegion(t *testing.T) {
	falseVal := false
	testCtx := NewContext()
	testCtx.SetConfig(&Configure{
		Current: "default",
		Profiles: map[string]*Profile{
			"default": {
				Name:       "default",
				Mode:       ModeAK,
				AccessKey:  "ak",
				SecretKey:  "sk",
				DisableSSL: &falseVal,
			},
		},
	})
	flag, _ := testCtx.fixedFlags.AddByName("region")
	flag.SetValue("ap-southeast-1")

	client, err := NewSimpleClient(testCtx)
	if err != nil {
		t.Fatalf("NewSimpleClient returned error: %v", err)
	}
	if client == nil || client.Config == nil || client.Config.Region == nil || *client.Config.Region != "ap-southeast-1" {
		t.Fatalf("region = %v, want ap-southeast-1", client.Config.Region)
	}
}

func TestNewSimpleClientRegionOverrideFixesEmptyEnvRegion(t *testing.T) {
	t.Setenv("BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS", "")
	t.Setenv("BYTEPLUS_ACCESS_KEY", "env-ak")
	t.Setenv("BYTEPLUS_SECRET_KEY", "env-sk")
	t.Setenv("BYTEPLUS_REGION", "")

	testCtx := NewContext()
	testCtx.SetConfig(&Configure{Profiles: map[string]*Profile{}})
	flag, _ := testCtx.fixedFlags.AddByName("region")
	flag.SetValue("ap-southeast-1")

	client, err := NewSimpleClient(testCtx)
	if err != nil {
		t.Fatalf("NewSimpleClient returned error: %v", err)
	}
	if client == nil || client.Config == nil || client.Config.Region == nil || *client.Config.Region != "ap-southeast-1" {
		t.Fatalf("region = %v, want ap-southeast-1", client.Config.Region)
	}
}

func TestNewSimpleClientUsesDefaultCredentialChainWithEnvRegion(t *testing.T) {
	t.Setenv("BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS", "")
	t.Setenv("BYTEPLUS_ACCESS_KEY", "env-ak")
	t.Setenv("BYTEPLUS_SECRET_KEY", "env-sk")
	t.Setenv("BYTEPLUS_REGION", "ap-southeast-1")

	testCtx := NewContext()
	testCtx.SetConfig(&Configure{Profiles: map[string]*Profile{}})

	client, err := NewSimpleClient(testCtx)
	if err != nil {
		t.Fatalf("NewSimpleClient returned error: %v", err)
	}
	if client == nil || client.Config == nil || client.Config.Region == nil || *client.Config.Region != "ap-southeast-1" {
		t.Fatalf("region = %v, want ap-southeast-1", client.Config.Region)
	}
	if client.Config.Credentials == nil {
		t.Fatal("expected default credential chain to be configured")
	}
}

func TestNewSimpleClientNilConfigUsesDefaultCredentialChain(t *testing.T) {
	t.Setenv("BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS", "")
	t.Setenv("BYTEPLUS_ACCESS_KEY", "env-ak")
	t.Setenv("BYTEPLUS_SECRET_KEY", "env-sk")
	t.Setenv("BYTEPLUS_REGION", "ap-southeast-1")

	testCtx := NewContext()
	testCtx.SetConfig(nil)

	client, err := NewSimpleClient(testCtx)
	if err != nil {
		t.Fatalf("NewSimpleClient returned error: %v", err)
	}
	if client == nil || client.Config == nil || client.Config.Credentials == nil {
		t.Fatal("expected client with default credential chain")
	}
}

func TestNewSimpleClientDisableDefaultCredentials(t *testing.T) {
	t.Setenv("BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS", "true")
	t.Setenv("BYTEPLUS_ACCESS_KEY", "env-ak")
	t.Setenv("BYTEPLUS_SECRET_KEY", "env-sk")
	t.Setenv("BYTEPLUS_REGION", "ap-southeast-1")

	testCtx := NewContext()
	testCtx.SetConfig(&Configure{Profiles: map[string]*Profile{}})

	_, err := NewSimpleClient(testCtx)
	if err == nil {
		t.Fatal("expected error when default credential chain is disabled")
	}
	if !strings.Contains(err.Error(), "BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS=true") {
		t.Fatalf("error = %q, want disable default credentials guidance", err.Error())
	}
}

func TestNewSimpleClientRequiresRegion(t *testing.T) {
	t.Setenv("BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS", "")
	t.Setenv("BYTEPLUS_ACCESS_KEY", "env-ak")
	t.Setenv("BYTEPLUS_SECRET_KEY", "env-sk")
	t.Setenv("BYTEPLUS_REGION", "")

	testCtx := NewContext()
	testCtx.SetConfig(&Configure{Profiles: map[string]*Profile{}})

	_, err := NewSimpleClient(testCtx)
	if err == nil {
		t.Fatal("expected error when region is missing")
	}
	if !strings.Contains(err.Error(), "region not set") {
		t.Fatalf("error = %q, want missing region guidance", err.Error())
	}
}

func TestCliProviderContractAKMode(t *testing.T) {
	configPath := writeTestConfig(t, &Configure{
		Current: "test",
		Profiles: map[string]*Profile{
			"test": {
				Name:      "test",
				Mode:      ModeAK,
				AccessKey: "ak",
				SecretKey: "sk",
				Region:    "ap-southeast-1",
			},
		},
	})

	creds := clicreds.NewCliCredentials(configPath, "test")
	value, err := creds.Get()
	if err != nil {
		t.Fatalf("CliProvider returned error: %v", err)
	}
	if value.AccessKeyID != "ak" {
		t.Fatalf("AccessKeyID = %q, want ak", value.AccessKeyID)
	}
	if value.SecretAccessKey != "sk" {
		t.Fatalf("SecretAccessKey = %q, want sk", value.SecretAccessKey)
	}
}

func TestCliProviderContractProfileSelection(t *testing.T) {
	configPath := writeTestConfig(t, &Configure{
		Current: "default",
		Profiles: map[string]*Profile{
			"default": {
				Name:      "default",
				Mode:      ModeAK,
				AccessKey: "default-ak",
				SecretKey: "default-sk",
			},
			"prod": {
				Name:      "prod",
				Mode:      ModeAK,
				AccessKey: "prod-ak",
				SecretKey: "prod-sk",
			},
		},
	})

	creds := clicreds.NewCliCredentials(configPath, "prod")
	value, err := creds.Get()
	if err != nil {
		t.Fatalf("CliProvider returned error: %v", err)
	}
	if value.AccessKeyID != "prod-ak" {
		t.Fatalf("AccessKeyID = %q, want prod-ak", value.AccessKeyID)
	}
}

func TestCliProviderContractProfileNotFound(t *testing.T) {
	configPath := writeTestConfig(t, &Configure{
		Current: "default",
		Profiles: map[string]*Profile{
			"default": {
				Name:      "default",
				Mode:      ModeAK,
				AccessKey: "ak",
				SecretKey: "sk",
			},
		},
	})

	creds := clicreds.NewCliCredentials(configPath, "missing")
	if _, err := creds.Get(); err == nil {
		t.Fatal("expected error when profile does not exist")
	}
}
