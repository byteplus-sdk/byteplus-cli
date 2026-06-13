package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// expandFlatToJSON 将 dotted-key 参数展开为 JSON body。
// 数字路径段表示 1-based 数组下标；叶子节点按 metadata 转成对应 JSON 类型。
func expandFlatToJSON(flat map[string]string, apiMeta *ApiMeta) (map[string]interface{}, error) {
	tree := map[string]interface{}{}

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		segs := strings.Split(key, ".")
		if err := validateIndexSegments(key, segs); err != nil {
			return nil, err
		}
		leaf, err := convertLeaf(apiMeta, key, flat[key])
		if err != nil {
			return nil, err
		}
		if err := insertLeaf(tree, segs, key, leaf); err != nil {
			return nil, err
		}
	}

	out, err := collapseNode(tree, "")
	if err != nil {
		return nil, err
	}
	m, _ := out.(map[string]interface{})
	return m, nil
}

// validateIndexSegments 提前拒绝 0、负数、带符号数字等非法数组下标。
func validateIndexSegments(fullKey string, segs []string) error {
	for _, seg := range segs {
		n, err := strconv.Atoi(seg)
		if err != nil {
			continue
		}
		if !isNumericSeg(seg) || n < 1 {
			return fmt.Errorf("parameter %q: invalid array index %q (array indices must be positive 1-based integers)", fullKey, seg)
		}
	}
	return nil
}

// convertLeaf 根据 metadata 将单个叶子值转换为 JSON 标量或复合值。
func convertLeaf(apiMeta *ApiMeta, fullKey, raw string) (interface{}, error) {
	mt, matchedKey, ok := resolveRequestMetaType(apiMeta, fullKey)
	if !ok {
		return raw, nil
	}
	tn := mt.TypeName

	// array 的 indexed element 只转换单个元素，不能按整个数组解析。
	if isIndexedStringArrayElement(matchedKey) && isArrayType(tn) {
		return convertScalar(fullKey, raw, arrayElemType(mt))
	}

	switch {
	case tn == "object" || tn == "map" || isArrayType(tn):
		var v interface{}
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return nil, fmt.Errorf("parameter %q: expected JSON for %s, got %q", fullKey, tn, raw)
		}
		return v, nil
	default:
		return convertScalar(fullKey, raw, tn)
	}
}

// resolveRequestMetaType 兼容顶层 MetaTypes 和嵌套 ChildMetas 两种 metadata 结构。
func resolveRequestMetaType(apiMeta *ApiMeta, name string) (*MetaType, string, bool) {
	if mt, matched, ok := getRequestMetaType(apiMeta, name); ok {
		return mt, matched, true
	}
	if apiMeta == nil || apiMeta.Request == nil {
		return nil, "", false
	}

	segs := strings.Split(name, ".")
	meta := apiMeta.Request
	matched := make([]string, 0, len(segs))
	for i := 0; i < len(segs); i++ {
		if meta == nil || meta.MetaTypes == nil {
			return nil, "", false
		}
		seg := segs[i]
		mt, ok := meta.MetaTypes[seg]
		if !ok {
			return nil, "", false
		}
		matched = append(matched, seg)

		if isArrayType(mt.TypeName) && i+1 < len(segs) && isNumericSeg(segs[i+1]) {
			matched = append(matched, "N")
			if i+1 == len(segs)-1 {
				return mt, strings.Join(matched, "."), true
			}
			if meta.ChildMetas == nil {
				return nil, "", false
			}
			meta = meta.ChildMetas[seg]
			i++
			continue
		}

		if i == len(segs)-1 {
			return mt, strings.Join(matched, "."), true
		}
		if meta.ChildMetas == nil {
			return nil, "", false
		}
		meta = meta.ChildMetas[seg]
	}
	return nil, "", false
}

// convertScalar 将字符串转换为 metadata 声明的 JSON 标量类型。
func convertScalar(fullKey, raw, typeName string) (interface{}, error) {
	switch typeName {
	case "integer", "long":
		n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: expected %s, got %q", fullKey, typeName, raw)
		}
		return n, nil
	case "number", "float", "double":
		f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: expected %s, got %q", fullKey, typeName, raw)
		}
		return f, nil
	case "boolean":
		b, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("parameter %q: expected boolean, got %q", fullKey, raw)
		}
		return b, nil
	default:
		return raw, nil
	}
}

func isArrayType(typeName string) bool {
	return typeName == "array" || strings.HasPrefix(typeName, "array[")
}

func arrayElemType(mt *MetaType) string {
	if mt.TypeOf != "" {
		return mt.TypeOf
	}
	tn := mt.TypeName
	if strings.HasPrefix(tn, "array[") && strings.HasSuffix(tn, "]") {
		return tn[len("array[") : len(tn)-1]
	}
	return "string"
}

func insertLeaf(tree map[string]interface{}, segs []string, fullKey string, leaf interface{}) error {
	cur := tree
	for i := 0; i < len(segs)-1; i++ {
		seg := segs[i]
		switch child := cur[seg].(type) {
		case nil:
			next := map[string]interface{}{}
			cur[seg] = next
			cur = next
		case map[string]interface{}:
			cur = child
		default:
			return fmt.Errorf("parameter %q: conflicting paths at segment %q", fullKey, seg)
		}
	}
	last := segs[len(segs)-1]
	if _, exists := cur[last]; exists {
		return fmt.Errorf("parameter %q: conflicting paths at segment %q", fullKey, last)
	}
	cur[last] = leaf
	return nil
}

func isNumericSeg(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// collapseNode 将所有纯数字 key 的 map 折叠为 slice，并校验下标连续。
func collapseNode(v interface{}, path string) (interface{}, error) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v, nil
	}
	if len(m) == 0 {
		return m, nil
	}

	numeric := 0
	for k := range m {
		if isNumericSeg(k) {
			numeric++
		}
	}

	for k := range m {
		childPath := k
		if path != "" {
			childPath = path + "." + k
		}
		nc, err := collapseNode(m[k], childPath)
		if err != nil {
			return nil, err
		}
		m[k] = nc
	}

	if numeric == 0 {
		return m, nil
	}
	if numeric != len(m) {
		return nil, fmt.Errorf("parameter path %q: mixes object fields and array indices", path)
	}

	arr := make([]interface{}, len(m))
	seen := make([]bool, len(m))
	for k := range m {
		idx, err := strconv.Atoi(k)
		if err != nil || idx < 1 || idx > len(m) {
			return nil, fmt.Errorf("parameter path %q: array indices must be 1-based and contiguous", path)
		}
		if seen[idx-1] {
			return nil, fmt.Errorf("parameter path %q: duplicate array index %d", path, idx)
		}
		seen[idx-1] = true
		arr[idx-1] = m[k]
	}
	return arr, nil
}
