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
	"os"
	"strconv"
	"strings"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/client/metadata"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/credentials"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/credentials/clicreds"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/defaults"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/endpoints"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/session"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/signer/byteplussign"
)

type SdkClient struct {
	Config  *byteplus.Config
	Session *session.Session
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
		creds                              *credentials.Credentials
		region, endpoint, endpointResolver string
		disableSSl, useDualStack           bool
	)
	if ctx == nil || ctx.fixedFlags == nil {
		return nil, fmt.Errorf("invalid context for creating sdk client")
	}

	var currentProfile *Profile
	profileName := ""
	if ctx.config != nil {
		profileName = ctx.config.Current
		overrideProfile := false
		if f := ctx.fixedFlags.GetByName("profile"); f != nil && f.GetValue() != "" {
			profileName = f.GetValue()
			overrideProfile = true
		}
		currentProfile = ctx.config.Profiles[profileName]
		if overrideProfile && currentProfile == nil {
			return nil, fmt.Errorf("profile %q not found", profileName)
		}
	}

	if currentProfile == nil {
		if os.Getenv("BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS") == "true" {
			return nil, fmt.Errorf("no profile configured and default credential chain is disabled (BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS=true)")
		}
		// 无 active profile 时交给 SDK 默认凭证链：Env -> OIDC -> CLI profile -> ECS role。
		creds = defaults.NewDefaultCredentialProvider()
		region = os.Getenv("BYTEPLUS_REGION")
		endpoint = os.Getenv("BYTEPLUS_ENDPOINT")
		endpointResolver = os.Getenv("BYTEPLUS_ENDPOINT_RESOLVER")
		ssl := os.Getenv("BYTEPLUS_DISABLE_SSL")
		if ssl == "true" || ssl == "false" {
			disableSSl, _ = strconv.ParseBool(ssl)
		}
		dualStack := os.Getenv("BYTEPLUS_USE_DUALSTACK")
		if dualStack == "true" || dualStack == "false" {
			useDualStack, _ = strconv.ParseBool(dualStack)
		}
	} else {
		mode := strings.ToLower(strings.TrimSpace(currentProfile.Mode))
		if mode == ModeSSO {
			// SSO 的 STS 缓存由 CLI 负责刷新并写回 config，再由 SDK CliProvider 读取。
			sso := &Sso{
				Profile:        currentProfile,
				SsoSessionName: currentProfile.SsoSessionName,
				Region:         currentProfile.Region,
			}
			if err := sso.EnsureValidStsToken(ctx); err != nil {
				return nil, err
			}
		}

		if mode == ModeConsoleLogin {
			// console-login 只由 CLI 负责刷新本地登录缓存；最终凭证统一交给 SDK CliProvider 读取。
			_, err := EnsureValidLoginToken(ctx.config, profileName)
			if err != nil {
				return nil, err
			}
		}
		creds = clicreds.NewCliCredentials("", profileName)

		region = currentProfile.Region
		endpoint = currentProfile.Endpoint
		endpointResolver = currentProfile.EndpointResolver
		if currentProfile.DisableSSL != nil {
			disableSSl = *currentProfile.DisableSSL
		}
		if currentProfile.UseDualStack != nil {
			useDualStack = *currentProfile.UseDualStack
		}
	}

	if f := ctx.fixedFlags.GetByName("region"); f != nil && f.GetValue() != "" {
		region = f.GetValue()
	}
	if region == "" {
		return nil, fmt.Errorf("region not set, please set it via profile, ---region flag, or BYTEPLUS_REGION environment variable")
	}

	config := byteplus.NewConfig().
		WithRegion(region).
		WithCredentials(creds).
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

	sess, _ := session.NewSession(config)

	return &SdkClient{
		Config:  config,
		Session: sess,
	}, nil
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
