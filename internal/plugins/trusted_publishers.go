package plugins

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devhub-gin-backend/internal/domain"
)

const defaultTrustedPublishersPath = "storage/plugins/trusted_publishers.json"

type TrustedPublisherStatus string

const (
	TrustedPublisherTrusted TrustedPublisherStatus = "trusted"
	TrustedPublisherUnknown TrustedPublisherStatus = "unknown"
	TrustedPublisherBlocked TrustedPublisherStatus = "blocked"
	TrustedPublisherRevoked TrustedPublisherStatus = "revoked"
)

type TrustedPublisher struct {
	PublisherID        string `json:"publisher_id"`
	Name               string `json:"name,omitempty"`
	PublicKeyID        string `json:"public_key_id,omitempty"`
	PublicKeyAlgorithm string `json:"public_key_algorithm,omitempty"`
	PublicKey          string `json:"public_key,omitempty"`
	Status             string `json:"status,omitempty"`
	Notes              string `json:"notes,omitempty"`
}

type TrustedPublishersConfig struct {
	Publishers []TrustedPublisher `json:"publishers"`
}

func TrustedPublishersConfigFromDomain(items []domain.PluginTrustedPublisher) TrustedPublishersConfig {
	out := TrustedPublishersConfig{Publishers: make([]TrustedPublisher, 0, len(items))}
	for _, it := range items {
		out.Publishers = append(out.Publishers, TrustedPublisher{
			PublisherID:        it.PublisherID,
			Name:               it.Name,
			PublicKeyID:        it.PublicKeyID,
			PublicKeyAlgorithm: it.PublicKeyAlgorithm,
			PublicKey:          it.PublicKey,
			Status:             it.Status,
			Notes:              it.Notes,
		})
	}
	return out
}

func DomainPublishersFromConfig(cfg TrustedPublishersConfig) []domain.PluginTrustedPublisher {
	out := make([]domain.PluginTrustedPublisher, 0, len(cfg.Publishers))
	for _, it := range cfg.Publishers {
		out = append(out, domain.PluginTrustedPublisher{
			PublisherID:        strings.TrimSpace(it.PublisherID),
			Name:               strings.TrimSpace(it.Name),
			PublicKeyID:        strings.TrimSpace(it.PublicKeyID),
			PublicKeyAlgorithm: strings.TrimSpace(it.PublicKeyAlgorithm),
			PublicKey:          strings.TrimSpace(it.PublicKey),
			Fingerprint:        FingerprintTrustedPublisherPublicKey(it.PublicKey),
			Status:             strings.TrimSpace(it.Status),
			Notes:              strings.TrimSpace(it.Notes),
		})
	}
	return out
}

func FingerprintTrustedPublisherPublicKey(publicKey string) string {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(publicKey))
	if err != nil || len(raw) == 0 {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + base64.RawURLEncoding.EncodeToString(sum[:])
}

type TrustedPublisherMatch struct {
	Found     bool
	Publisher TrustedPublisher
}

// LoadTrustedPublishers loads local trusted publishers config from a fixed path under project root.
//
// Notes:
// - This config is the only trust source in v1.5 P2; plugin package publisher.json is just a declaration.
// - Missing file is not blocking; callers should treat it as "unavailable" and keep trust_status=unknown.
func LoadTrustedPublishers() (cfg TrustedPublishersConfig, found bool, err error) {
	workdir, werr := os.Getwd()
	if werr != nil {
		return TrustedPublishersConfig{}, false, fmt.Errorf("读取工作目录失败：%w", werr)
	}
	root, rerr := findProjectRoot(workdir)
	if rerr != nil {
		return TrustedPublishersConfig{}, false, fmt.Errorf("读取项目根目录失败：%w", rerr)
	}
	path := filepath.Join(root, defaultTrustedPublishersPath)
	raw, rerr2 := os.ReadFile(path)
	if rerr2 != nil {
		if os.IsNotExist(rerr2) {
			return TrustedPublishersConfig{}, false, nil
		}
		return TrustedPublishersConfig{}, true, domain.NewPluginError("plugin_package_trusted_publishers_unavailable", "读取 trusted_publishers 配置失败").
			WithStatus(500).
			WithDetail("path", defaultTrustedPublishersPath).
			WithDetail("reason", strings.TrimSpace(rerr2.Error())).
			WithSuggestion("请检查 storage/plugins/trusted_publishers.json 是否存在且可读。")
	}
	var wire TrustedPublishersConfig
	if err := json.Unmarshal(raw, &wire); err != nil {
		return TrustedPublishersConfig{}, true, domain.NewPluginError("plugin_package_trusted_publishers_unavailable", "trusted_publishers 不是合法 JSON").
			WithStatus(400).
			WithDetail("path", defaultTrustedPublishersPath).
			WithDetail("reason", strings.TrimSpace(err.Error())).
			WithSuggestion("请修复 storage/plugins/trusted_publishers.json 后重试。")
	}
	return wire, true, nil
}

func FindTrustedPublisher(cfg TrustedPublishersConfig, publisherID string, publicKeyID string) TrustedPublisherMatch {
	publisherID = strings.TrimSpace(publisherID)
	publicKeyID = strings.TrimSpace(publicKeyID)
	for _, it := range cfg.Publishers {
		if strings.TrimSpace(it.PublisherID) == "" || strings.TrimSpace(it.PublicKeyID) == "" {
			continue
		}
		if strings.TrimSpace(it.PublisherID) == publisherID && strings.TrimSpace(it.PublicKeyID) == publicKeyID {
			return TrustedPublisherMatch{Found: true, Publisher: it}
		}
	}
	return TrustedPublisherMatch{Found: false}
}

func (m TrustedPublisherMatch) NormalizedStatus() string {
	if !m.Found {
		return string(TrustedPublisherUnknown)
	}
	v := strings.ToLower(strings.TrimSpace(m.Publisher.Status))
	switch v {
	case string(TrustedPublisherTrusted), string(TrustedPublisherBlocked), string(TrustedPublisherRevoked):
		return v
	default:
		return string(TrustedPublisherUnknown)
	}
}

func (m TrustedPublisherMatch) PublicKeyBytes() ([]byte, error) {
	if !m.Found {
		return nil, fmt.Errorf("publisher not found")
	}
	key := strings.TrimSpace(m.Publisher.PublicKey)
	if key == "" {
		return nil, fmt.Errorf("missing public key")
	}
	raw, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("invalid public key base64")
	}
	return raw, nil
}
