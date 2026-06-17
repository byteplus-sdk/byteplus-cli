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

import (
	"fmt"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/endpoints"
	"os"
	"strconv"
	"strings"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client/metadata"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/credentials"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/session"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/signer/byteplussign"
)

type SdkClient struct {
	Config      *byteplus.Config
	Session     *session.Session
	DebugLogger *DebugLogger
}

type SdkClientInfo struct {
	ServiceName string
	Action      string
	Version     string
	Method      string
	ContentType string
}

func NewSimpleClient(ctx *Context) (*SdkClient, error) {
	var (
		ak, sk, sessionToken, region, endpoint, endpointResolver string
		httpProxy, httpsProxy                                    string
		disableSSl, useDualStack                                 bool
	)
	if ctx == nil || ctx.fixedFlags == nil {
		return nil, fmt.Errorf("invalid context for creating sdk client")
	}

	// first try to get ak/sk/region from config file
	var currentProfile *Profile
	profileName := ""
	profileSource := "env:BYTEPLUS_*"
	if ctx.config != nil {
		profileName, profileSource = defaultProfileNameWithSource(ctx.config)
		overrideProfile := false
		if f := ctx.fixedFlags.GetByName("profile"); f != nil && f.GetValue() != "" {
			profileName = f.GetValue()
			profileSource = "flag"
			overrideProfile = true
		}
		currentProfile = ctx.config.Profiles[profileName]
		if overrideProfile && currentProfile == nil {
			return nil, fmt.Errorf("profile %q not found", profileName)
		}
		if currentProfile != nil {
			mode := strings.ToLower(strings.TrimSpace(currentProfile.Mode))
			switch mode {
			case ModeSSO:
				sso := &Sso{
					Profile:        currentProfile,
					SsoSessionName: currentProfile.SsoSessionName,
					Region:         currentProfile.Region,
				}
				if err := sso.EnsureValidStsToken(ctx); err != nil {
					return nil, err
				}
				fallthrough
			case ModeAK, "":
				ak = currentProfile.AccessKey
				sk = currentProfile.SecretKey
				region = currentProfile.Region
				if region == "" {
					region = os.Getenv("BYTEPLUS_REGION")
				}
				endpoint = currentProfile.Endpoint
				if endpoint == "" {
					endpoint = os.Getenv("BYTEPLUS_ENDPOINT")
				}
				endpointResolver = currentProfile.EndpointResolver
				if endpointResolver == "" {
					endpointResolver = os.Getenv("BYTEPLUS_ENDPOINT_RESOLVER")
				}
				sessionToken = currentProfile.SessionToken
				httpProxy = currentProfile.HTTPProxy
				httpsProxy = currentProfile.HTTPSProxy
				if currentProfile.DisableSSL != nil {
					disableSSl = *currentProfile.DisableSSL
				}
				if currentProfile.UseDualStack != nil {
					useDualStack = *currentProfile.UseDualStack
				}

				if ak == "" {
					return nil, fmt.Errorf("profile AccessKey not set")
				}
				if sk == "" {
					return nil, fmt.Errorf("profile SecretKey not set")
				}
			case ModeConsoleLogin:
				creds, err := EnsureValidLoginToken(ctx.config, profileName)
				if err != nil {
					return nil, err
				}
				ak = creds.AccessKeyID
				sk = creds.SecretAccessKey
				sessionToken = creds.SessionToken
				region = currentProfile.Region
				if region == "" {
					region = os.Getenv("BYTEPLUS_REGION")
				}
				endpoint = currentProfile.Endpoint
				if endpoint == "" {
					endpoint = os.Getenv("BYTEPLUS_ENDPOINT")
				}
				endpointResolver = currentProfile.EndpointResolver
				if endpointResolver == "" {
					endpointResolver = os.Getenv("BYTEPLUS_ENDPOINT_RESOLVER")
				}
				httpProxy = currentProfile.HTTPProxy
				httpsProxy = currentProfile.HTTPSProxy
				if currentProfile.DisableSSL != nil {
					disableSSl = *currentProfile.DisableSSL
				}
				if currentProfile.UseDualStack != nil {
					useDualStack = *currentProfile.UseDualStack
				}

			default:
				return nil, fmt.Errorf("unsupported credential mode %q, supported modes: ak, sso, console-login", currentProfile.Mode)
			}
		}
	}

	// if cannot get from config file, try to get from export variable
	if currentProfile == nil {
		ak = os.Getenv("BYTEPLUS_ACCESS_KEY")
		sk = os.Getenv("BYTEPLUS_SECRET_KEY")
		region = os.Getenv("BYTEPLUS_REGION")
		endpoint = os.Getenv("BYTEPLUS_ENDPOINT")
		endpointResolver = os.Getenv("BYTEPLUS_ENDPOINT_RESOLVER")
		sessionToken = os.Getenv("BYTEPLUS_SESSION_TOKEN")
		ssl := os.Getenv("BYTEPLUS_DISABLE_SSL")
		if ssl == "true" || ssl == "false" {
			disableSSl, _ = strconv.ParseBool(ssl)
		}
		dualStack := os.Getenv("BYTEPLUS_USE_DUALSTACK")
		if dualStack == "true" || dualStack == "false" {
			useDualStack, _ = strconv.ParseBool(dualStack)
		}

		if ak == "" {
			return nil, fmt.Errorf("BYTEPLUS_ACCESS_KEY not set")
		}
		if sk == "" {
			return nil, fmt.Errorf("BYTEPLUS_SECRET_KEY not set")
		}
	}

	if f := ctx.fixedFlags.GetByName("region"); f != nil && f.GetValue() != "" {
		region = f.GetValue()
	}

	if f := ctx.fixedFlags.GetByName("endpoint"); f != nil && f.GetValue() != "" {
		endpoint = f.GetValue()
		endpointResolver = ""
	}

	if region == "" {
		return nil, fmt.Errorf("region not set, please set it via profile, ---region flag, or BYTEPLUS_REGION environment variable")
	}

	config := byteplus.NewConfig().
		WithRegion(region).
		WithCredentials(credentials.NewStaticCredentials(ak, sk, sessionToken)).
		WithDisableSSL(disableSSl)

	resolverValue := strings.ToLower(strings.TrimSpace(endpointResolver))
	switch resolverValue {
	case "standard":
		config.WithEndpointResolver(endpoints.NewStandardEndpointResolver())
	default:
		if endpoint != "" {
			if strings.ToLower(strings.TrimSpace(endpoint)) == "auto-addressing" {
				config.WithEndpointResolver(endpoints.NewStandardEndpointResolver())
			} else {
				config.WithEndpoint(endpoint)
			}
		}
	}

	if useDualStack {
		config.WithUseDualStack(true)
	}
	if httpProxy != "" {
		config.WithHTTPProxy(httpProxy)
	}
	if httpsProxy != "" {
		config.WithHTTPSProxy(httpsProxy)
	}

	debugLogClientConfig(ctx, debugClientConfig{
		ProfileName:          profileName,
		ProfileSource:        profileSource,
		CredentialMode:       debugCredentialMode(currentProfile),
		Region:               region,
		Endpoint:             endpoint,
		EndpointResolver:     endpointResolver,
		DisableSSL:           disableSSl,
		UseDualStack:         useDualStack,
		HTTPProxyConfigured:  httpProxy != "",
		HTTPSProxyConfigured: httpsProxy != "",
	})

	sess, _ := session.NewSession(config)

	return &SdkClient{
		Config:      config,
		Session:     sess,
		DebugLogger: debugLoggerFromContext(ctx),
	}, nil
}

func defaultProfileName(cfg *Configure) string {
	name, _ := defaultProfileNameWithSource(cfg)
	return name
}

func defaultProfileNameWithSource(cfg *Configure) (string, string) {
	if cfg != nil && cfg.Current != "" {
		return cfg.Current, "current"
	}
	if profile := os.Getenv("BYTEPLUS_PROFILE"); profile != "" {
		return profile, "env:BYTEPLUS_PROFILE"
	}
	if profile := os.Getenv("BYTEPLUS_CLI_PROFILE"); profile != "" {
		return profile, "env:BYTEPLUS_CLI_PROFILE"
	}
	return "", "env:BYTEPLUS_*"
}

type debugClientConfig struct {
	ProfileName          string
	ProfileSource        string
	CredentialMode       string
	Region               string
	Endpoint             string
	EndpointResolver     string
	DisableSSL           bool
	UseDualStack         bool
	HTTPProxyConfigured  bool
	HTTPSProxyConfigured bool
}

func debugCredentialMode(profile *Profile) string {
	if profile == nil {
		return "env"
	}

	mode := strings.ToLower(strings.TrimSpace(profile.Mode))
	if mode == "" {
		return ModeAK
	}
	return mode
}

func debugLogClientConfig(ctx *Context, info debugClientConfig) {
	logger := debugLoggerFromContext(ctx)
	if logger == nil || !logger.Enabled() {
		return
	}
	logger.Printf("client_config profile_source=%s profile=%s credential_mode=%s region=%s endpoint=%s endpoint_resolver=%s disable_ssl=%t use_dual_stack=%t http_proxy_configured=%t https_proxy_configured=%t",
		info.ProfileSource,
		info.ProfileName,
		info.CredentialMode,
		info.Region,
		info.Endpoint,
		info.EndpointResolver,
		info.DisableSSL,
		info.UseDualStack,
		info.HTTPProxyConfigured,
		info.HTTPSProxyConfigured,
	)
}

func (s *SdkClient) initClient(svc string, version string) *client.Client {
	config := s.Session.ClientConfig(svc)
	c := client.New(
		*config.Config,
		metadata.ClientInfo{
			ServiceName:   svc,
			ServiceID:     svc,
			SigningName:   config.SigningName,
			SigningRegion: config.SigningRegion,
			Endpoint:      config.Endpoint,
			APIVersion:    version,
		},
		config.Handlers,
	)

	c.Handlers.Build.PushBackNamed(clientVersionAndUserAgentHandler)
	c.Handlers.Sign.PushBackNamed(byteplussign.SignRequestHandler)
	c.Handlers.Build.PushBackNamed(byteplusquery.BuildHandler)
	c.Handlers.Unmarshal.PushBackNamed(byteplusquery.UnmarshalHandler)
	c.Handlers.UnmarshalMeta.PushBackNamed(byteplusquery.UnmarshalMetaHandler)
	c.Handlers.UnmarshalError.PushBackNamed(byteplusquery.UnmarshalErrorHandler)
	s.addDebugRequestAttemptHandler(c)

	return c
}

func (s *SdkClient) CallSdk(info SdkClientInfo, input interface{}) (output *map[string]interface{}, err error) {
	c := s.initClient(info.ServiceName, info.Version)
	op := &request.Operation{
		Name:       info.Action,
		HTTPMethod: strings.ToUpper(info.Method),
		HTTPPath:   "/",
	}
	if input == nil {
		input = &map[string]interface{}{}
	}
	output = &map[string]interface{}{}
	req := c.NewRequest(op, input, output)
	if strings.ToLower(info.ContentType) == "application/json" {
		req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
	} else if info.ContentType != "" {
		req.HTTPRequest.Header.Set("Content-Type", info.ContentType)
	}
	err = req.Send()
	return output, err
}
