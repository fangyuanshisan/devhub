package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

const RedactedValue = "[REDACTED]"

// ConfigHash computes a stable hash for JSON objects by decoding and re-encoding with stable key order.
// Invalid/empty JSON returns empty hash.
func ConfigHash(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return ""
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return ""
	}
	canonical, err := marshalCanonicalJSON(v)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}

func marshalCanonicalJSON(v any) ([]byte, error) {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		obj := make(map[string]any, len(x))
		for _, k := range keys {
			obj[k] = x[k]
		}
		// Recurse to normalize nested maps.
		for _, k := range keys {
			if m, ok := obj[k].(map[string]any); ok {
				buf, err := marshalCanonicalJSON(m)
				if err != nil {
					return nil, err
				}
				var out any
				_ = json.Unmarshal(buf, &out)
				obj[k] = out
			}
		}
		return json.Marshal(obj)
	case []any:
		// Arrays: keep order; normalize nested maps.
		out := make([]any, 0, len(x))
		for _, it := range x {
			if m, ok := it.(map[string]any); ok {
				buf, err := marshalCanonicalJSON(m)
				if err != nil {
					return nil, err
				}
				var vv any
				_ = json.Unmarshal(buf, &vv)
				out = append(out, vv)
				continue
			}
			out = append(out, it)
		}
		return json.Marshal(out)
	default:
		return json.Marshal(v)
	}
}

// RedactConfig returns a redacted config object for display purposes.
// It never returns secrets in plain text based on schema sensitivity hints and key heuristics.
func RedactConfig(schema any, configRaw string) any {
	configRaw = strings.TrimSpace(configRaw)
	if configRaw == "" || !json.Valid([]byte(configRaw)) {
		return map[string]any{}
	}
	var v any
	if err := json.Unmarshal([]byte(configRaw), &v); err != nil {
		return map[string]any{}
	}
	return redactValue(schema, "", v)
}

func redactValue(schema any, path string, value any) any {
	// Only object paths can be traversed with schema properties.
	obj, ok := value.(map[string]any)
	if !ok {
		return value
	}
	props := schemaProperties(schema)
	out := map[string]any{}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		subSchema := props[k]
		subPath := joinPath(path, k)
		if isSensitiveField(k, subSchema) || isSensitivePath(subPath) {
			out[k] = RedactedValue
			continue
		}
		out[k] = redactValue(subSchema, subPath, obj[k])
	}
	return out
}

func schemaProperties(schema any) map[string]any {
	m, ok := schema.(map[string]any)
	if !ok || len(m) == 0 {
		return map[string]any{}
	}
	props, _ := m["properties"].(map[string]any)
	out := map[string]any{}
	for k, v := range props {
		if mm, ok := v.(map[string]any); ok {
			out[k] = mm
		} else {
			out[k] = v
		}
	}
	return out
}

func isSensitiveField(key string, schema any) bool {
	m, ok := schema.(map[string]any)
	if ok {
		if b, ok := m["x-sensitive"].(bool); ok && b {
			return true
		}
		if b, ok := m["writeOnly"].(bool); ok && b {
			return true
		}
		if format, ok := m["format"].(string); ok && strings.ToLower(strings.TrimSpace(format)) == "password" {
			return true
		}
	}
	return isSensitiveKey(key)
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return false
	}
	for _, token := range []string{"password", "passwd", "secret", "token", "credential", "app_secret", "aes_key", "private_key", "access_key", "api_key", "apikey"} {
		if strings.Contains(key, token) {
			return true
		}
	}
	if strings.HasSuffix(key, "_key") || strings.HasSuffix(key, "_secret") || strings.HasSuffix(key, "_token") {
		return true
	}
	return false
}

func isSensitivePath(path string) bool {
	// best-effort nested path heuristic.
	return isSensitiveKey(path)
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// DiffPluginConfig builds a stable, redacted diff between two JSON configs.
// It supports nested object paths; arrays are diffed as a whole value.
func DiffPluginConfig(schema any, beforeRaw, afterRaw string) ([]string, []domain.PluginConfigDiffItem) {
	beforeObj := decodeJSONObject(beforeRaw)
	afterObj := decodeJSONObject(afterRaw)

	changedKeys := topLevelChangedKeys(beforeObj, afterObj)
	items := []domain.PluginConfigDiffItem{}

	props := schemaProperties(schema)
	visitDiffObject(props, "", beforeObj, afterObj, &items)
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return changedKeys, items
}

func decodeJSONObject(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return map[string]any{}
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return map[string]any{}
	}
	return v
}

func topLevelChangedKeys(before, after map[string]any) []string {
	seen := map[string]bool{}
	keys := []string{}
	for k := range before {
		seen[k] = true
		keys = append(keys, k)
	}
	for k := range after {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	out := []string{}
	for _, k := range keys {
		if !deepEqualJSON(before[k], after[k]) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func visitDiffObject(schemaProps map[string]any, prefix string, before, after map[string]any, out *[]domain.PluginConfigDiffItem) {
	seen := map[string]bool{}
	keys := []string{}
	for k := range before {
		seen[k] = true
		keys = append(keys, k)
	}
	for k := range after {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		path := joinPath(prefix, k)
		subSchema := schemaProps[k]
		sensitive := isSensitiveField(k, subSchema) || isSensitivePath(path)

		bv, bok := before[k]
		av, aok := after[k]

		// Recurse only when both are objects and schema suggests object properties.
		if bm, ok1 := bv.(map[string]any); ok1 {
			if am, ok2 := av.(map[string]any); ok2 {
				subProps := schemaProperties(subSchema)
				visitDiffObject(subProps, path, bm, am, out)
				continue
			}
		}

		item := domain.PluginConfigDiffItem{Path: path}
		switch {
		case !bok && aok:
			item.Type = "added"
			item.After = redactMaybe(sensitive, av)
		case bok && !aok:
			item.Type = "removed"
			item.Before = redactMaybe(sensitive, bv)
		default:
			if deepEqualJSON(bv, av) {
				item.Type = "unchanged"
			} else {
				item.Type = "changed"
			}
			item.Before = redactMaybe(sensitive, bv)
			item.After = redactMaybe(sensitive, av)
		}
		*out = append(*out, item)
	}
}

func redactMaybe(sensitive bool, v any) any {
	if sensitive {
		if v == nil {
			return nil
		}
		return RedactedValue
	}
	return v
}

func deepEqualJSON(a, b any) bool {
	// json marshaling as a stable compare.
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(ab) == string(bb)
}
