package plugins

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type PluginConfigKeyring struct {
	CurrentKeyID string
	Keys         map[string][]byte // key_id -> 32 bytes
	// LegacyV1Supported indicates the keyring is able to decrypt enc:v1 values (no key_id).
	LegacyV1Supported bool
}

type pluginConfigKeysEnv struct {
	Current string            `json:"current"`
	Keys    map[string]string `json:"keys"`
}

type pluginConfigKeysArrayItem struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Primary bool   `json:"primary"`
}

// LoadPluginConfigKeyring loads keyring from env.
//
// Supported env formats:
//  1. JSON form:
//     DEVHUB_PLUGIN_CONFIG_KEYS='{"current":"key-2026-01","keys":{"key-2026-01":"...","key-2025-12":"..."}}'
//     or array form:
//     DEVHUB_PLUGIN_CONFIG_KEYS='[{"id":"key-2026-01","key":"...","primary":true},{"id":"key-2025-12","key":"..."}]'
//  2. Split form:
//     DEVHUB_PLUGIN_CONFIG_KEY_ID=key-2026-01
//     DEVHUB_PLUGIN_CONFIG_KEY=base64-key
//     DEVHUB_PLUGIN_CONFIG_OLD_KEYS='{"key-2025-12":"base64-key-old"}'
//
// Backward compat:
// - DEVHUB_PLUGIN_CONFIG_KEY (without KEY_ID) is treated as a single current key with id "legacy".
//
// It returns (keyring, ok, err). ok=false means no key env configured.
func LoadPluginConfigKeyring() (*PluginConfigKeyring, bool, error) {
	// Prefer JSON form.
	if raw := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_CONFIG_KEYS")); raw != "" {
		kr := &PluginConfigKeyring{Keys: map[string][]byte{}}

		// Support both object form and array form.
		if strings.HasPrefix(raw, "[") {
			var arr []pluginConfigKeysArrayItem
			if err := json.Unmarshal([]byte(raw), &arr); err != nil {
				return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEYS JSON 不合法: %w", err)
			}
			primaryID := ""
			for _, item := range arr {
				id := strings.TrimSpace(item.ID)
				material := strings.TrimSpace(item.Key)
				if id == "" {
					return nil, true, errors.New("DEVHUB_PLUGIN_CONFIG_KEYS 存在空 id")
				}
				if _, exists := kr.Keys[id]; exists {
					return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEYS 存在重复 id: %s", id)
				}
				key, err := parse32ByteKey(material)
				if err != nil {
					return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEYS[%s] 无效: %w", id, err)
				}
				kr.Keys[id] = key
				if item.Primary {
					primaryID = id
				}
			}
			if primaryID == "" && len(arr) > 0 {
				primaryID = strings.TrimSpace(arr[0].ID)
			}
			kr.CurrentKeyID = primaryID
		} else {
			var cfg pluginConfigKeysEnv
			if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
				return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEYS JSON 不合法: %w", err)
			}
			kr.CurrentKeyID = strings.TrimSpace(cfg.Current)
			for id, material := range cfg.Keys {
				id = strings.TrimSpace(id)
				material = strings.TrimSpace(material)
				if id == "" {
					return nil, true, errors.New("DEVHUB_PLUGIN_CONFIG_KEYS.keys 存在空 key_id")
				}
				if _, exists := kr.Keys[id]; exists {
					return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEYS.keys 存在重复 key_id: %s", id)
				}
				key, err := parse32ByteKey(material)
				if err != nil {
					return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEYS.keys[%s] 无效: %w", id, err)
				}
				kr.Keys[id] = key
			}
		}
		if err := ValidatePluginConfigKeyring(kr); err != nil {
			return nil, true, err
		}
		kr.LegacyV1Supported = len(kr.Keys) > 0
		return kr, true, nil
	}

	// Split form.
	id := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_CONFIG_KEY_ID"))
	keyRaw := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_CONFIG_KEY"))
	oldRaw := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_CONFIG_OLD_KEYS"))

	// Backward compat: only DEVHUB_PLUGIN_CONFIG_KEY set.
	if id == "" && keyRaw != "" && oldRaw == "" {
		key, err := parse32ByteKey(keyRaw)
		if err != nil {
			return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEY 无效: %w", err)
		}
		kr := &PluginConfigKeyring{
			CurrentKeyID:      "legacy",
			Keys:              map[string][]byte{"legacy": key},
			LegacyV1Supported: true,
		}
		return kr, true, nil
	}

	if id == "" && keyRaw == "" && oldRaw == "" {
		return nil, false, nil
	}

	kr := &PluginConfigKeyring{CurrentKeyID: id, Keys: map[string][]byte{}}
	if keyRaw != "" {
		key, err := parse32ByteKey(keyRaw)
		if err != nil {
			return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEY 无效: %w", err)
		}
		if id == "" {
			return nil, true, errors.New("缺少 DEVHUB_PLUGIN_CONFIG_KEY_ID")
		}
		kr.Keys[id] = key
	}
	if oldRaw != "" {
		var olds map[string]string
		if err := json.Unmarshal([]byte(oldRaw), &olds); err != nil {
			return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_OLD_KEYS JSON 不合法: %w", err)
		}
		for oldID, material := range olds {
			oldID = strings.TrimSpace(oldID)
			material = strings.TrimSpace(material)
			if oldID == "" {
				return nil, true, errors.New("DEVHUB_PLUGIN_CONFIG_OLD_KEYS 存在空 key_id")
			}
			if _, exists := kr.Keys[oldID]; exists {
				return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_OLD_KEYS 存在重复 key_id: %s", oldID)
			}
			key, err := parse32ByteKey(material)
			if err != nil {
				return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_OLD_KEYS[%s] 无效: %w", oldID, err)
			}
			kr.Keys[oldID] = key
		}
	}
	if err := ValidatePluginConfigKeyring(kr); err != nil {
		return nil, true, err
	}
	kr.LegacyV1Supported = len(kr.Keys) > 0
	return kr, true, nil
}

func ValidatePluginConfigKeyring(kr *PluginConfigKeyring) error {
	if kr == nil {
		return errors.New("plugin config keyring 为空")
	}
	kr.CurrentKeyID = strings.TrimSpace(kr.CurrentKeyID)
	if kr.CurrentKeyID == "" {
		return errors.New("缺少 current key_id")
	}
	if kr.Keys == nil || len(kr.Keys) == 0 {
		return errors.New("缺少 keys")
	}
	k, ok := kr.Keys[kr.CurrentKeyID]
	if !ok || len(k) != 32 {
		return fmt.Errorf("current key_id 不存在或无效: %s", kr.CurrentKeyID)
	}
	return nil
}

func parse32ByteKey(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("key 为空")
	}
	// base64 32 bytes is recommended.
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil {
		if len(b) != 32 {
			return nil, fmt.Errorf("base64 解码后长度必须为 32 bytes，当前为 %d", len(b))
		}
		return b, nil
	}
	// Fallback derive via sha256 for dev/testing convenience.
	sum := sha256.Sum256([]byte(raw))
	out := make([]byte, 32)
	copy(out, sum[:])
	return out, nil
}

func (kr *PluginConfigKeyring) CurrentKey() (string, []byte) {
	if kr == nil {
		return "", nil
	}
	return kr.CurrentKeyID, kr.Keys[kr.CurrentKeyID]
}

func (kr *PluginConfigKeyring) ResolveKey(keyID string) ([]byte, bool) {
	if kr == nil {
		return nil, false
	}
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return nil, false
	}
	k, ok := kr.Keys[keyID]
	return k, ok
}

func (kr *PluginConfigKeyring) AllKeyIDs() []string {
	if kr == nil || len(kr.Keys) == 0 {
		return []string{}
	}
	ids := make([]string, 0, len(kr.Keys))
	for id := range kr.Keys {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
