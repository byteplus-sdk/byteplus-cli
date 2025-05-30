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
	"strings"
)

// Copyright 2023 Byteplus.  All Rights Reserved.

type MetaType struct {
	TypeName string `json:"TypeName,omitempty"`
	TypeOf   string `json:"TypeOf,omitempty"`
	Required bool   `json:"Required,omitempty"`
}

type Meta struct {
	MetaTypes  map[string]*MetaType `json:"MetaTypes,omitempty"`
	ChildMetas map[string]*Meta     `json:"ChildMetas,omitempty"`
}

type ApiMeta struct {
	Request  *Meta
	Response *Meta
}

func (m *Meta) getDefaultValue(s string) interface{} {
	var r interface{}
	switch s {
	case "string":
		r = "string"
	case "boolean":
		r = false
	case "integer":
		r = 0
	}
	return r
}

func (m *Meta) GetReqBody() map[string]interface{} {
	r := make(map[string]interface{})
	for k, v := range m.MetaTypes {
		switch v.TypeName {
		case "object":
			if len(m.ChildMetas) > 0 {
				if _, ok := m.ChildMetas[k]; ok {
					r[k] = m.ChildMetas[k].GetReqBody()
				}
			}
		case "array":
			if v.TypeOf != "object" {
				r[k] = v.TypeName
			} else {
				if len(m.ChildMetas) > 0 {
					if _, ok := m.ChildMetas[k]; ok {
						r[k] = []interface{}{
							m.ChildMetas[k].GetReqBody(),
						}
					}
				}
			}
		case "map":
			if v.TypeOf != "object" {
				r1 := map[string]interface{}{
					"string": m.getDefaultValue(v.TypeOf),
				}
				r[k] = r1
			} else {
				if len(m.ChildMetas) > 0 {
					if _, ok := m.ChildMetas[k]; ok {
						r1 := map[string]interface{}{
							"string": m.ChildMetas[k].GetReqBody(),
						}
						r[k] = r1
					}
				}
			}
		default:
			r[k] = m.getDefaultValue(v.TypeName)
		}

	}
	return r
}

func (m *ApiMeta) GetReqTypeName(pattern string) string {
	p := strings.Split(pattern, ".")
	var result string
	meta := m.Request

	if v, ok := meta.MetaTypes[pattern]; ok {
		return v.TypeName
	}

	var index int
	for _, _p := range p {
		index++
		metaTypes := meta.MetaTypes
		if _, ok := metaTypes[_p]; ok {
			result = " " + metaTypes[_p].TypeName
		} else {
			result = ""
		}
		if index < len(p) {
			if _, ok := meta.ChildMetas[_p]; ok {
				meta = meta.ChildMetas[_p]
			} else {
				break
			}
		}
	}
	return result
}

func (m *ApiMeta) GetReqRequired(pattern string) bool {
	p := strings.Split(pattern, ".")
	var result bool
	meta := m.Request

	if v, ok := meta.MetaTypes[pattern]; ok {
		return v.Required
	}

	var index int
	for _, _p := range p {
		index++
		metaTypes := meta.MetaTypes
		if _, ok := metaTypes[_p]; ok {
			result = metaTypes[_p].Required
		} else {
			result = false
		}
		if index < len(p) {
			if _, ok := meta.ChildMetas[_p]; ok {
				meta = meta.ChildMetas[_p]
			} else {
				break
			}
		}
	}
	return result
}
