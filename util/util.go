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

package util

// Copyright 2023 Byteplus.  All Rights Reserved.

import (
	"encoding/json"
	"os/user"
	"strings"
)

func IsRepeatedField(f string) bool {
	return strings.Contains(f, ".N")
}

func IsJsonArray(value string) bool {
	return len(value) >= 2 && value[0] == '[' && value[len(value)-1] == ']'
}

// ParseToJsonArrayOrObject try to parse string to json array or json object
func ParseToJsonArrayOrObject(s string) (interface{}, bool) {
	if !json.Valid([]byte(s)) || len(s) < 2 {
		return nil, false
	}

	var a interface{}
	if (s[0] == '[' && s[len(s)-1] == ']') || (s[0] == '{' && s[len(s)-1] == '}') {
		if err := json.Unmarshal([]byte(s), &a); err != nil {
			return err, false
		} else {
			return a, true
		}
	}
	return nil, false
}

func GetConfigFileDir() (string, error) {
	var (
		err     error
		homeDir string
	)

	if homeDir, err = getHomeDir(); err != nil {
		return "", err
	}

	return homeDir + "/.byteplus/", nil
}

func getHomeDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return user.HomeDir, nil
}
