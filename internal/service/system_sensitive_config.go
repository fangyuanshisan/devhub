package service

import (
	"strings"

	"devhub-gin-backend/internal/domain"
)

// SystemSensitiveConfigStatus is a Core-only, read-only status view for sensitive config governance.
// It must never expose plaintext secrets/tokens/keys.
func (s *Service) SystemSensitiveConfigStatus() (domain.SystemSensitiveConfigStatusResponse, error) {
	keyStatus, _ := s.PluginConfigKeyStatus()
	secretStatus, _ := s.SecretCenterStatus()

	allowlist, _ := s.ExternalServiceHTTPAllowlist()
	httpPolicy := allowlist.Policy

	notes := []string{
		"启动密钥只从环境变量读取（或由 KMS/Vault/容器 Secret 注入到环境变量）。DevHub 后台不会保存或生成 root key；修改后需要重启生效。",
		"external_service HTTP allowlist 由系统默认 localhost、启动环境变量和后台受控配置共同组成；后台配置会立即参与运行时校验，环境变量修改通常需要重启。",
	}

	// Keep notes compact and safe.
	for i := range notes {
		notes[i] = strings.TrimSpace(notes[i])
	}

	return domain.SystemSensitiveConfigStatusResponse{
		PluginConfigKeyring: keyStatus,
		ExternalServiceHTTP: httpPolicy,
		SecretCenter:        secretStatus,
		Notes:               notes,
	}, nil
}
