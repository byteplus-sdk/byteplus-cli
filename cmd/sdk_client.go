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
		ak, sk, sessionToken, region, endpoint string
		disableSSl                             bool
	)

	// first try to get ak/sk/region from config file
	var currentProfile *Profile
	if ctx.config != nil {
		if currentProfile = ctx.config.Profiles[ctx.config.Current]; currentProfile != nil {
			ak = currentProfile.AccessKey
			sk = currentProfile.SecretKey
			region = currentProfile.Region
			endpoint = currentProfile.Endpoint
			sessionToken = currentProfile.SessionToken
			disableSSl = *currentProfile.DisableSSL

			if ak == "" {
				return nil, fmt.Errorf("profile AccessKey not set")
			}
			if sk == "" {
				return nil, fmt.Errorf("profile SecretKey not set")
			}
			if region == "" {
				return nil, fmt.Errorf("profile Region not set")
			}
		}
	}

	// if cannot get from config file, try to get from export variable
	if currentProfile == nil {
		ak = os.Getenv("BYTEPLUS_ACCESS_KEY")
		sk = os.Getenv("BYTEPLUS_SECRET_KEY")
		region = os.Getenv("BYTEPLUS_REGION")
		endpoint = os.Getenv("BYTEPLUS_ENDPOINT")
		sessionToken = os.Getenv("BYTEPLUS_SESSION_TOKEN")
		ssl := os.Getenv("BYTEPLUS_DISABLE_SSL")
		if ssl == "true" || ssl == "false" {
			disableSSl, _ = strconv.ParseBool(ssl)
		}

		if ak == "" {
			return nil, fmt.Errorf("BYTEPLUS_ACCESS_KEY not set")
		}
		if sk == "" {
			return nil, fmt.Errorf("BYTEPLUS_SECRET_KEY not set")
		}
		if region == "" {
			return nil, fmt.Errorf("BYTEPLUS_REGION not set")
		}
	}

	config := byteplus.NewConfig().
		WithRegion(region).
		WithCredentials(credentials.NewStaticCredentials(ak, sk, sessionToken)).
		WithDisableSSL(disableSSl)

	if endpoint != "" {
		config.WithEndpoint(endpoint)
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
