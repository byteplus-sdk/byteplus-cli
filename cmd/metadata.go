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
	"strconv"
	"strings"
)

// Copyright 2023 Byteplus.  All Rights Reserved.

type ByteplusMeta struct {
	ApiInfo  *ApiInfo
	Request  *MetaInfo
	Response *MetaInfo
}

type MetaInfo struct {
	Basic     *[]string
	Structure *map[string]MetaInfo
}

type ApiInfo struct {
	Method      string
	ContentType string
	ServiceName string
	ParamTypes  map[string]string
	// int float64
	// [], {}
}

type StructInfo struct {
	PkgName     string
	ServiceName string
	Version     string
}

type param struct {
	key      string
	typeName string
	required bool
}

func formatParamsHelpUsage(params []param) []string {
	maxKeyLen := -1
	maxTypeNameLen := -1
	for _, p := range params {
		if len(p.key) > maxKeyLen {
			maxKeyLen = len(p.key)
		}
		if len(p.typeName) > maxTypeNameLen {
			maxTypeNameLen = len(p.typeName)
		}
	}

	maxKeyLen++
	maxTypeNameLen++

	// TODO: not print required field now
	//formatString := "%-" + strconv.Itoa(maxKeyLen) + "v%-" + strconv.Itoa(maxTypeNameLen) + "v %v"
	formatString := "%-" + strconv.Itoa(maxKeyLen) + "v%-" + strconv.Itoa(maxTypeNameLen) + "v"

	var paramStrings []string
	for _, p := range params {
		//paramStrings = append(paramStrings, fmt.Sprintf(formatString, p.key, p.typeName, formatRequired(p.required)))
		paramStrings = append(paramStrings, fmt.Sprintf(formatString, p.key, p.typeName))
	}

	return paramStrings
}

func formatRequired(required bool) string {
	if required {
		return "Required"
	}
	return "Optional"
}

func (meta *ByteplusMeta) GetRequestParams(apiMeta *ApiMeta) (params []param) {
	var s []string
	var traverse func(MetaInfo)

	traverse = func(meta MetaInfo) {
		if meta.Basic != nil {
			for _, v := range *meta.Basic {
				s = append(s, v)
				if apiMeta == nil {
					paramKey := strings.Join(s, ".")
					params = append(params, param{
						key:      paramKey,
						typeName: "",
						required: false,
					})
				} else {
					paramKey := strings.Join(s, ".")
					params = append(params, param{
						key:      paramKey,
						typeName: apiMeta.GetReqTypeName(paramKey),
						required: apiMeta.GetReqRequired(paramKey),
					})
				}
				s = s[:len(s)-1]
			}
		}

		if meta.Structure != nil {
			for k2, v2 := range *meta.Structure {
				s = append(s, k2)
				traverse(v2)
				s = s[:len(s)-1]
			}
		}
	}

	traverse(*meta.Request)
	return
}
