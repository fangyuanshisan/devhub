package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

// ValidateConfigJSON validates config_json against a plugin's simplified config_schema.
//
// This is intentionally NOT a full JSON Schema implementation. We implement the minimum
// governance constraints used by DevHub plugin platform:
// - type: object/string/number/integer/boolean/array
// - properties (object)
// - required (object)
// - enum
// - min/max (number/integer)
// - items (array)
// - additionalProperties (object)
//
// Unknown fields are rejected by default unless additionalProperties=true is explicitly set.
func ValidateConfigJSON(def domain.Plugin, configJSON string) error {
	configJSON = strings.TrimSpace(configJSON)
	if configJSON == "" {
		// Empty means default config; always valid.
		return nil
	}
	if !json.Valid([]byte(configJSON)) {
		return errors.New("config_json 必须是合法 JSON")
	}

	var value any
	if err := json.Unmarshal([]byte(configJSON), &value); err != nil {
		return errors.New("config_json 解析失败")
	}
	schema, ok := def.ConfigSchema.(map[string]any)
	if !ok || len(schema) == 0 {
		// No schema declared => only JSON validity is enforced.
		return nil
	}
	return validateValueAgainstSchema(value, schema, "$")
}

func validateValueAgainstSchema(value any, schema map[string]any, path string) error {
	wantType := strings.TrimSpace(asString(schema["type"]))
	if wantType == "" {
		// No type constraint.
		return nil
	}
	switch wantType {
	case "object":
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s 必须是 object", path)
		}
		required := asStringSlice(schema["required"])
		sort.Strings(required)
		for _, key := range required {
			if _, ok := obj[key]; !ok {
				return fmt.Errorf("%s 缺少必填字段 %s", path, key)
			}
		}

		props, _ := schema["properties"].(map[string]any)
		additional := asBool(schema["additionalProperties"], false)

		// Reject unknown fields by default.
		if !additional && props != nil {
			for key := range obj {
				if _, ok := props[key]; !ok {
					return fmt.Errorf("%s 存在未声明字段 %s", path, key)
				}
			}
		}
		for key, rawSchema := range props {
			subSchema, ok := rawSchema.(map[string]any)
			if !ok {
				continue
			}
			if v, ok := obj[key]; ok {
				if err := validateValueAgainstSchema(v, subSchema, path+"."+key); err != nil {
					return err
				}
			}
		}
		return nil
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s 必须是 string", path)
		}
		return validateEnum(value, schema, path)
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s 必须是 boolean", path)
		}
		return nil
	case "integer":
		num, ok := value.(float64)
		if !ok || math.Trunc(num) != num {
			return fmt.Errorf("%s 必须是 integer", path)
		}
		if err := validateMinMax(num, schema, path); err != nil {
			return err
		}
		return validateEnum(num, schema, path)
	case "number":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%s 必须是 number", path)
		}
		if err := validateMinMax(num, schema, path); err != nil {
			return err
		}
		return validateEnum(num, schema, path)
	case "array":
		arr, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s 必须是 array", path)
		}
		itemsSchema, _ := schema["items"].(map[string]any)
		if itemsSchema == nil {
			return nil
		}
		for i, v := range arr {
			if err := validateValueAgainstSchema(v, itemsSchema, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
		return nil
	default:
		// Unsupported type keyword => ignore.
		return nil
	}
}

func validateEnum(value any, schema map[string]any, path string) error {
	raw, ok := schema["enum"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	for _, allowed := range arr {
		if fmt.Sprintf("%v", allowed) == fmt.Sprintf("%v", value) {
			return nil
		}
	}
	return fmt.Errorf("%s 值不在允许范围 enum 内", path)
}

func validateMinMax(num float64, schema map[string]any, path string) error {
	if minRaw, ok := schema["min"]; ok {
		if min, ok := minRaw.(float64); ok && num < min {
			return fmt.Errorf("%s 不能小于 %v", path, min)
		}
	}
	if maxRaw, ok := schema["max"]; ok {
		if max, ok := maxRaw.(float64); ok && num > max {
			return fmt.Errorf("%s 不能大于 %v", path, max)
		}
	}
	return nil
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asBool(v any, def bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func asStringSlice(v any) []string {
	items, ok := v.([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if s, ok := it.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}
