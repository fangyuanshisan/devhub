package service

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type ConfigEncryptResult struct {
	EncryptedJSON string
	RedactedJSON  string
	ChangedKeys   []string
}

// encryptPluginConfigJSON encrypts sensitive fields for storage.
// It keeps non-sensitive fields as-is. It supports placeholder retention and clearing.
//
// Inputs:
// - plugin: used for schema
// - storedJSON: current stored config_json (may be ciphertext for sensitive fields)
// - submittedJSON: user's submitted config JSON string (plaintext for sensitive fields)
//
// Output:
// - EncryptedJSON: to be persisted
// - RedactedJSON: for API return / audit / history diff (never contains plaintext secrets or ciphertext enc:v1)
func (s *Service) encryptPluginConfigJSON(plugin domain.Plugin, storedJSON string, submittedJSON string) (ConfigEncryptResult, error) {
	out := ConfigEncryptResult{
		EncryptedJSON: strings.TrimSpace(submittedJSON),
		RedactedJSON:  "{}",
		ChangedKeys:   []string{},
	}
	submittedJSON = strings.TrimSpace(submittedJSON)
	storedJSON = strings.TrimSpace(storedJSON)
	if submittedJSON == "" {
		// treat as clear all config (no encryption needed)
		out.EncryptedJSON = ""
		out.RedactedJSON = "{}"
		return out, nil
	}

	// Parse stored/submitted JSON objects.
	var submitted any
	if err := json.Unmarshal([]byte(submittedJSON), &submitted); err != nil {
		return out, err
	}
	var stored any
	if storedJSON != "" {
		_ = json.Unmarshal([]byte(storedJSON), &stored)
	}

	// Merge placeholders: keep stored ciphertext when user submits placeholder/omits.
	merged := pluginregistry.MergeSensitivePlaceholders(plugin.ConfigSchema, stored, submitted)

	// Encrypt merged.
	key, ok, err := pluginregistry.LoadPluginConfigKey()
	if err != nil {
		return out, err
	}
	// No key: reject if any sensitive plaintext exists.
	needEncrypt := hasSensitivePlaintext(plugin.ConfigSchema, "", merged)
	if needEncrypt && !ok {
		// Allow in memory/E2E environments with an ephemeral key, but never rely on this for production.
		if strings.TrimSpace(os.Getenv("CMS_STORE")) == "memory" || strings.TrimSpace(os.Getenv("DEVHUB_E2E_TESTING")) == "1" {
			sum := sha256.Sum256([]byte("devhub-plugin-config-test-key"))
			key = sum[:]
			ok = true
		}
	}
	if needEncrypt && !ok {
		return out, domain.NewPluginError("plugin_config_encryption_key_missing", "缺少插件配置加密密钥，无法保存敏感字段").
			WithStatus(500).
			WithSuggestion("请配置环境变量 DEVHUB_PLUGIN_CONFIG_KEY（推荐 base64 32 bytes）后重试。")
	}
	encrypted := merged
	if needEncrypt {
		encrypted, err = encryptSensitiveObject(key, plugin.ConfigSchema, "", merged)
		if err != nil {
			return out, domain.NewPluginError("plugin_config_encrypt_failed", "敏感配置加密失败").
				WithStatus(500).
				WithDetail("plugin_code", plugin.Code).
				WithSuggestion("请检查密钥配置是否正确后重试。")
		}
	}

	encBytes, err := json.Marshal(encrypted)
	if err != nil {
		return out, err
	}
	out.EncryptedJSON = string(encBytes)

	// Redact for API/audit/history: never expose ciphertext.
	out.RedactedJSON = marshalRedacted(plugin.ConfigSchema, encrypted)
	out.ChangedKeys, _ = pluginregistry.DiffPluginConfig(plugin.ConfigSchema, storedJSON, out.EncryptedJSON)
	return out, nil
}

func marshalRedacted(schema any, encrypted any) string {
	raw, _ := json.Marshal(encrypted)
	redacted := pluginregistry.RedactConfig(schema, string(raw))
	buf, _ := json.Marshal(redacted)
	return string(buf)
}

func hasSensitivePlaintext(schema any, path string, v any) bool {
	obj, ok := v.(map[string]any)
	if !ok {
		return false
	}
	props := pluginregistrySensitiveSchemaProps(schema)
	for k, vv := range obj {
		subPath := k
		if path != "" {
			subPath = path + "." + k
		}
		subSchema := props[k]
		if pluginregistry.IsSensitiveField(subSchema, k, subPath) {
			// If value is empty string, it's a "clear" and doesn't require encryption.
			if s, ok := vv.(string); ok {
				if strings.TrimSpace(s) == "" {
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(s), pluginregistry.EncryptedPrefixV1) {
					continue
				}
				// placeholder should have been merged to stored, but double-check.
				if pluginregistry.IsSensitivePlaceholder(s) {
					continue
				}
			}
			return true
		}
		if hasSensitivePlaintext(subSchema, subPath, vv) {
			return true
		}
	}
	return false
}

func encryptSensitiveObject(key []byte, schema any, path string, v any) (any, error) {
	obj, ok := v.(map[string]any)
	if !ok {
		return v, nil
	}
	props := pluginregistrySensitiveSchemaProps(schema)
	out := map[string]any{}
	for k, vv := range obj {
		subPath := k
		if path != "" {
			subPath = path + "." + k
		}
		subSchema := props[k]
		if pluginregistry.IsSensitiveField(subSchema, k, subPath) {
			// clear: keep empty string
			if s, ok := vv.(string); ok && strings.TrimSpace(s) == "" {
				out[k] = ""
				continue
			}
			// keep ciphertext
			if pluginregistry.IsEncryptedValue(vv) {
				out[k] = vv
				continue
			}
			// encrypt string only; non-string treated as invalid and stringified.
			plain := ""
			if s, ok := vv.(string); ok {
				plain = s
			} else {
				b, _ := json.Marshal(vv)
				plain = string(b)
			}
			enc, err := pluginregistry.EncryptStringV1(key, plain)
			if err != nil {
				return nil, err
			}
			out[k] = enc
			continue
		}
		nv, err := encryptSensitiveObject(key, subSchema, subPath, vv)
		if err != nil {
			return nil, err
		}
		out[k] = nv
	}
	return out, nil
}

func pluginregistrySensitiveSchemaProps(schema any) map[string]any {
	// local copy to avoid cross-package unexported helpers.
	m, ok := schema.(map[string]any)
	if !ok || len(m) == 0 {
		return map[string]any{}
	}
	props, _ := m["properties"].(map[string]any)
	out := map[string]any{}
	for k, v := range props {
		out[k] = v
	}
	return out
}
