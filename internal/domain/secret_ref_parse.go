package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const secretRefPrefix = "secret://"

var (
	secretNamespaceRe = regexp.MustCompile(`^[a-z0-9_-]+$`)
	secretNameRe      = regexp.MustCompile(`^[a-z0-9_-]+(?:/[a-z0-9_-]+)*$`)
)

// ParseSecretRef parses ref in the format:
//
//	secret://{namespace}/{name}
//
// Validation rules:
// - namespace: lowercase letters, digits, '_' and '-'
// - name: lowercase letters, digits, '_' and '-', with optional '/' separators
// - no empty namespace/name
// - no path traversal ("..") or backslashes
func ParseSecretRef(ref string) (namespace string, name string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", errors.New("secret_ref 不能为空")
	}
	if !strings.HasPrefix(ref, secretRefPrefix) {
		return "", "", fmt.Errorf("secret_ref 必须以 %s 开头", secretRefPrefix)
	}
	rest := strings.TrimPrefix(ref, secretRefPrefix)
	rest = strings.TrimSpace(rest)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.New("secret_ref 格式不合法，应为 secret://{namespace}/{name}")
	}
	ns := strings.TrimSpace(parts[0])
	nm := strings.TrimSpace(parts[1])
	if ns == "" || nm == "" {
		return "", "", errors.New("secret_ref namespace/name 不能为空")
	}
	if strings.Contains(nm, "\\") {
		return "", "", errors.New("secret_ref name 不允许包含反斜杠")
	}
	if strings.Contains(nm, "..") {
		return "", "", errors.New("secret_ref name 不允许包含 ..")
	}
	if !secretNamespaceRe.MatchString(ns) {
		return "", "", errors.New("secret_ref namespace 不合法（仅允许小写字母/数字/_/-）")
	}
	if !secretNameRe.MatchString(nm) {
		return "", "", errors.New("secret_ref name 不合法（仅允许小写字母/数字/_/-，可包含 / 分段）")
	}
	return ns, nm, nil
}

// BuildSecretRef builds a normalized secret_ref string from namespace/name.
func BuildSecretRef(namespace string, name string) (string, error) {
	ns := strings.TrimSpace(namespace)
	nm := strings.TrimSpace(name)
	ref := secretRefPrefix + ns + "/" + nm
	if _, _, err := ParseSecretRef(ref); err != nil {
		return "", err
	}
	return ref, nil
}

func IsSecretRef(v string) bool {
	_, _, err := ParseSecretRef(v)
	return err == nil
}
