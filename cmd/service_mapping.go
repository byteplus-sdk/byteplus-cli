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

var (
	serviceMapping = map[string]string{
		//"rds_mysql_v2": "rds_mysql",
	}

	//svcVersionMapping = map[string]map[string]string{
	//	"rds_mysql": {
	//		"2022-01-01": "rds_mysql_v2",
	//	},
	//}
)

func SetServiceMapping(s1, s2 string) {
	serviceMapping[s1] = s2
}

func GetServiceMapping(s string) (string, bool) {
	if v, ok := serviceMapping[s]; ok {
		return v, true
	}
	return s, false
}

//func GetSvcVersionMapping(svc, version string) (string, bool) {
//	if v, ok := svcVersionMapping[svc]; ok {
//		if v1, ok1 := v[version]; ok1 {
//			return v1, true
//		} else {
//			return svc, false
//		}
//	}
//	return svc, false
//}
