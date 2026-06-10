package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
