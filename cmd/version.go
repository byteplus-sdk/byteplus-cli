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
	"runtime"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
)

var clientVersionAndUserAgentHandler = request.NamedHandler{
	Name: "ByteplusCliUserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler(clientName, clientVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH),
}

const clientName = "byteplus-cli"
const clientVersion = "1.0.3"
