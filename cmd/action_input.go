package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/byteplus-sdk/byteplus-cli/util"
)

// buildActionInput 根据 API 的 Content-Type 构造 SDK 入参。
// JSON API 支持两种互斥输入：--body 传完整 JSON，或通过扁平参数自动展开为 JSON body。
func buildActionInput(flags []*Flag, apiMeta *ApiMeta, jsonBody bool) (interface{}, bool, error) {
	hasBody := false
	hasFlat := false
	var bodyVal string
	flat := make(map[string]string)

	for _, f := range flags {
		if f.Name == "body" {
			hasBody = true
			bodyVal = f.value
			continue
		}
		hasFlat = true
		flat[f.Name] = f.value
	}

	if hasBody && hasFlat {
		return nil, false, fmt.Errorf("--body cannot be used together with flattened parameters")
	}

	if hasBody {
		parsed, err := parseJSONBody(bodyVal)
		if err != nil {
			return nil, false, err
		}
		return parsed, true, nil
	}

	if jsonBody {
		nested, err := expandFlatToJSON(flat, apiMeta)
		if err != nil {
			return nil, false, err
		}
		return nested, false, nil
	}

	// 非 JSON API 保持历史 dotted-key 行为，服务端会继续按原规则处理参数。
	input := make(map[string]interface{})
	for name, val := range flat {
		if isStringParam(apiMeta, name) {
			input[name] = val
		} else if a, success := util.ParseToJsonArrayOrObject(strings.TrimSpace(val)); success {
			input[name] = a
		} else {
			input[name] = val
		}
	}
	return input, false, nil
}

// parseJSONBody 只接受 JSON object 或 array，避免把普通字符串误当作 JSON body 发送。
func parseJSONBody(body string) (interface{}, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &m); err == nil {
		return &m, nil
	}

	var a []interface{}
	if err := json.Unmarshal([]byte(body), &a); err == nil {
		return &a, nil
	}

	return nil, fmt.Errorf("json format error")
}
