package plugins

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	EncryptedPrefixV1      = "enc:v1:"
	EncryptedPrefixV2      = "enc:v2:"
	SensitivePlaceholderV1 = "[ENCRYPTED]"
)

// SensitiveKeyWords defines the default heuristic keywords for sensitive fields.
// It is used as a fallback when schema is missing.
var SensitiveKeyWords = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"access_token",
	"refresh_token",
	"api_key",
	"apikey",
	"key",
	"credential",
	"app_secret",
	"aes_key",
	"encoding_aes_key",
	"private_key",
	"client_secret",
}

func IsSensitiveField(schema any, key string, path string) bool {
	key = strings.TrimSpace(key)
	path = strings.TrimSpace(path)

	// Schema explicit markers have higher priority.
	if schema != nil {
		if m, ok := schema.(map[string]any); ok {
			if b, ok := m["x-sensitive"].(bool); ok && b {
				return true
			}
			if b, ok := m["writeOnly"].(bool); ok && b {
				return true
			}
			if format, ok := m["format"].(string); ok && strings.EqualFold(strings.TrimSpace(format), "password") {
				return true
			}
		}
	}

	// Fallback heuristic by key/path.
	return isSensitiveKey(key) || isSensitiveKey(path)
}

// NOTE: isSensitiveKey and schemaProperties are implemented in config_diff.go

func IsEncryptedValue(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, EncryptedPrefixV1) || strings.HasPrefix(s, EncryptedPrefixV2)
}

func IsSensitivePlaceholder(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	return s == SensitivePlaceholderV1 || s == "******"
}

// LoadPluginConfigKey reads DEVHUB_PLUGIN_CONFIG_KEY and derives a 32-byte key.
//
// Accepted formats:
// - base64-encoded 32 bytes
// - any non-empty string, derived via sha256
//
// It returns (key, ok, err). ok=false means missing.
func LoadPluginConfigKey() ([]byte, bool, error) {
	raw := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_CONFIG_KEY"))
	if raw == "" {
		return nil, false, nil
	}
	// Try base64 first.
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil {
		if len(b) != 32 {
			return nil, true, fmt.Errorf("DEVHUB_PLUGIN_CONFIG_KEY base64 解码后长度必须为 32 bytes，当前为 %d", len(b))
		}
		return b, true, nil
	}
	// Derive from string.
	sum := sha256.Sum256([]byte(raw))
	out := make([]byte, 32)
	copy(out, sum[:])
	return out, true, nil
}

func EncryptStringV1(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("加密密钥长度必须为 32 bytes")
	}
	if strings.HasPrefix(strings.TrimSpace(plaintext), EncryptedPrefixV1) {
		return strings.TrimSpace(plaintext), nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return EncryptedPrefixV1 + base64.StdEncoding.EncodeToString(nonce) + ":" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func EncryptStringV2(keyID string, key []byte, plaintext string) (string, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return "", errors.New("key_id 不能为空")
	}
	if len(key) != 32 {
		return "", errors.New("加密密钥长度必须为 32 bytes")
	}
	plain := strings.TrimSpace(plaintext)
	if strings.HasPrefix(plain, EncryptedPrefixV2) || strings.HasPrefix(plain, EncryptedPrefixV1) {
		// do not double-encrypt ciphertext; caller decides whether to re-encrypt.
		return plain, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return EncryptedPrefixV2 + keyID + ":" + base64.StdEncoding.EncodeToString(nonce) + ":" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptStringV1(key []byte, enc string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("解密密钥长度必须为 32 bytes")
	}
	enc = strings.TrimSpace(enc)
	if !strings.HasPrefix(enc, EncryptedPrefixV1) {
		return enc, nil
	}
	parts := strings.Split(strings.TrimPrefix(enc, EncryptedPrefixV1), ":")
	if len(parts) != 2 {
		return "", errors.New("密文格式不合法")
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("nonce 解码失败")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("ciphertext 解码失败")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(nonce) != gcm.NonceSize() {
		return "", errors.New("nonce 长度不合法")
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("解密失败")
	}
	return string(plain), nil
}

type CipherInfo struct {
	Version string // v1|v2|plain
	KeyID   string // v2 only
}

func DetectCiphertextVersion(v string) CipherInfo {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, EncryptedPrefixV2) {
		parts := strings.Split(strings.TrimPrefix(v, EncryptedPrefixV2), ":")
		if len(parts) >= 3 {
			return CipherInfo{Version: "v2", KeyID: strings.TrimSpace(parts[0])}
		}
		return CipherInfo{Version: "v2", KeyID: ""}
	}
	if strings.HasPrefix(v, EncryptedPrefixV1) {
		return CipherInfo{Version: "v1", KeyID: ""}
	}
	return CipherInfo{Version: "plain", KeyID: ""}
}

func DecryptStringWithKeyring(kr *PluginConfigKeyring, enc string) (string, CipherInfo, error) {
	info := DetectCiphertextVersion(enc)
	switch info.Version {
	case "plain":
		return enc, info, nil
	case "v2":
		keyID := info.KeyID
		if keyID == "" {
			return "", info, errors.New("密文 key_id 缺失")
		}
		key, ok := kr.ResolveKey(keyID)
		if !ok {
			return "", info, errors.New("缺少对应解密密钥")
		}
		plain, err := decryptStringV2WithKey(key, keyID, enc)
		if err != nil {
			return "", info, err
		}
		return plain, info, nil
	case "v1":
		if kr == nil || len(kr.Keys) == 0 {
			return "", info, errors.New("缺少解密密钥")
		}
		// v1 has no key_id, try current first, then others.
		if curID, curKey := kr.CurrentKey(); curID != "" && len(curKey) == 32 {
			if plain, err := DecryptStringV1(curKey, enc); err == nil {
				return plain, info, nil
			}
		}
		for _, id := range kr.AllKeyIDs() {
			if id == kr.CurrentKeyID {
				continue
			}
			k, _ := kr.ResolveKey(id)
			if len(k) != 32 {
				continue
			}
			if plain, err := DecryptStringV1(k, enc); err == nil {
				return plain, info, nil
			}
		}
		return "", info, errors.New("解密失败")
	default:
		return "", info, errors.New("不支持的密文版本")
	}
}

func decryptStringV2WithKey(key []byte, keyID string, enc string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("解密密钥长度必须为 32 bytes")
	}
	enc = strings.TrimSpace(enc)
	if !strings.HasPrefix(enc, EncryptedPrefixV2) {
		return enc, nil
	}
	parts := strings.Split(strings.TrimPrefix(enc, EncryptedPrefixV2), ":")
	if len(parts) != 3 {
		return "", errors.New("密文格式不合法")
	}
	if strings.TrimSpace(parts[0]) != strings.TrimSpace(keyID) {
		return "", errors.New("密文 key_id 不匹配")
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("nonce 解码失败")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", errors.New("ciphertext 解码失败")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(nonce) != gcm.NonceSize() {
		return "", errors.New("nonce 长度不合法")
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("解密失败")
	}
	return string(plain), nil
}

// MergeSensitivePlaceholders merges the user's submitted config with the stored config.
//
// Rules:
// - If user omits a sensitive field: keep stored value.
// - If user submits placeholder ([ENCRYPTED]/******): keep stored value.
// - If user submits empty string: clear (set empty string).
// - Otherwise: use user value (to be encrypted later).
//
// The merge is best-effort and supports nested objects. Arrays are treated as whole values.
func MergeSensitivePlaceholders(schema any, stored any, submitted any) any {
	switch s := submitted.(type) {
	case map[string]any:
		out := map[string]any{}
		storedMap, _ := stored.(map[string]any)
		keys := make([]string, 0, len(s))
		for k := range s {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out[k] = mergeSensitiveValue(schema, k, k, storedMap[k], s[k])
		}
		// Keep stored keys not present in submitted.
		for k, v := range storedMap {
			if _, ok := s[k]; ok {
				continue
			}
			out[k] = v
		}
		return out
	default:
		return submitted
	}
}

func mergeSensitiveValue(schema any, key string, path string, stored any, submitted any) any {
	props := schemaProperties(schema)
	subSchema := props[key]
	if IsSensitiveField(subSchema, key, path) {
		if submitted == nil {
			return stored
		}
		if IsSensitivePlaceholder(submitted) {
			return stored
		}
		if s, ok := submitted.(string); ok && strings.TrimSpace(s) == "" {
			return ""
		}
		return submitted
	}

	// Recurse on objects.
	sm, ok1 := stored.(map[string]any)
	subm, ok2 := submitted.(map[string]any)
	if ok1 && ok2 {
		// For nested, schema should be subSchema.
		return MergeSensitivePlaceholders(subSchema, sm, subm)
	}
	// Non-object, keep submitted.
	return submitted
}
