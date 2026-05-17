package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	webhookSecretRefPrefix          = "whsec_"
	webhookSecretBytesDefault       = 32
	webhookSecretGracePeriodDefault = 24 * time.Hour
	webhookSignatureAlgHMACSHA256   = "HMAC-SHA256"
)

type WebhookSecretOperator struct {
	ID   int64
	Name string
}

type CreateWebhookSecretRequest struct {
	PluginCode string `json:"plugin_code"`
	TargetURL  string `json:"target_url"`
}

type CreateWebhookSecretResponse struct {
	Secret          domain.PluginWebhookSecret `json:"secret_record"`
	SecretPlaintext string                     `json:"secret"` // plaintext, returned once
}

type RotateWebhookSecretResponse struct {
	OldSecretRef    string                     `json:"old_secret_ref"`
	NewSecretRef    string                     `json:"new_secret_ref"`
	GraceUntil      string                     `json:"grace_until"`
	NewSecretRecord domain.PluginWebhookSecret `json:"secret_record"`
	SecretPlaintext string                     `json:"secret"` // plaintext, returned once
}

func (s *Service) ListPluginWebhookSecrets(filter domain.PluginWebhookSecretFilter) (domain.PluginWebhookSecretListResponse, error) {
	items, total, err := s.repo.PluginWebhookSecrets(filter)
	if err != nil {
		return domain.PluginWebhookSecretListResponse{}, err
	}
	f := filter.Normalize()
	// Never expose secret_ciphertext in list.
	_ = s.MarkExpiredPluginWebhookSecrets(time.Now(), 50)
	for i := range items {
		items[i].SecretCiphertext = ""
	}
	return domain.PluginWebhookSecretListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     f.Page,
			PageSize: f.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginWebhookSecret(id int64) (domain.PluginWebhookSecret, error) {
	it, ok := s.repo.PluginWebhookSecretByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginWebhookSecret{}, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404)
	}
	_ = s.MarkExpiredPluginWebhookSecrets(time.Now(), 50)
	// Never expose ciphertext via API.
	it.SecretCiphertext = ""
	return it, nil
}

func (s *Service) CreatePluginWebhookSecret(operator WebhookSecretOperator, req CreateWebhookSecretRequest) (CreateWebhookSecretResponse, error) {
	pluginCode := strings.TrimSpace(req.PluginCode)
	targetURL := strings.TrimSpace(req.TargetURL)
	if pluginCode == "" || targetURL == "" {
		return CreateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_invalid", "plugin_code/target_url 必填").WithStatus(400)
	}
	if _, err := url.Parse(targetURL); err != nil {
		return CreateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_invalid", "target_url 不合法").WithStatus(400)
	}
	if _, ok := s.repo.PluginByCode(pluginCode); !ok {
		return CreateWebhookSecretResponse{}, domain.NewPluginError("plugin_not_found", "插件不存在").WithStatus(404).WithDetail("plugin_code", pluginCode)
	}

	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return CreateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_encrypt_key_invalid", "Secret 加密密钥配置不合法").WithStatus(500)
	}
	needEncrypt := true
	if needEncrypt && !ok {
		// Allow memory/e2e with ephemeral key, but never rely on this for production.
		if strings.TrimSpace(os.Getenv("CMS_STORE")) == "memory" || strings.TrimSpace(os.Getenv("DEVHUB_E2E_TESTING")) == "1" {
			sum := sha256.Sum256([]byte("devhub-webhook-secret-test-key"))
			k := sum[:]
			kr = &pluginregistry.PluginConfigKeyring{
				CurrentKeyID:      "test-key",
				Keys:              map[string][]byte{"test-key": k},
				LegacyV1Supported: true,
			}
			ok = true
		}
	}
	if needEncrypt && !ok {
		return CreateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_encryption_key_missing", "缺少 Webhook Secret 加密密钥，无法创建 Secret").
			WithStatus(500).
			WithSuggestion("请配置环境变量 DEVHUB_PLUGIN_CONFIG_KEYS（推荐 JSON 格式）或 DEVHUB_PLUGIN_CONFIG_KEY_ID/DEVHUB_PLUGIN_CONFIG_KEY 后重试。")
	}

	secretRef := webhookSecretRefPrefix + randomToken(18)
	plaintext := randomWebhookSecretPlaintext(webhookSecretBytesDefault)
	secretHash := sha256StringHex(plaintext)

	keyID, key := kr.CurrentKey()
	ciphertext, err := pluginregistry.EncryptStringV2(keyID, key, plaintext)
	if err != nil {
		return CreateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_encrypt_failed", "Secret 加密失败").WithStatus(500)
	}

	now := Now()
	record := domain.PluginWebhookSecret{
		PluginCode:       pluginCode,
		TargetURL:        targetURL,
		SecretRef:        secretRef,
		SecretCiphertext: ciphertext,
		SecretHash:       secretHash,
		Version:          1,
		Status:           domain.PluginWebhookSecretStatusActive,
		RotationGroup:    pluginCode + ":" + targetURL,
		ActiveFrom:       now,
		GraceUntil:       "",
		CreatedBy:        operator.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Ensure only one active per (plugin_code, target_url).
	if old, okOld := s.repo.ActivePluginWebhookSecret(pluginCode, targetURL); okOld && old.ID > 0 {
		old.Status = domain.PluginWebhookSecretStatusDisabled
		old.UpdatedAt = now
		_, _ = s.repo.SavePluginWebhookSecret(old)
	}

	saved, err := s.repo.AppendPluginWebhookSecret(record)
	if err != nil {
		return CreateWebhookSecretResponse{}, err
	}

	// Audit without plaintext/ciphertext.
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.webhook.secret.created",
		Target:    "webhook-secrets#" + saved.SecretRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": pluginCode, "target_url": targetURL, "secret_ref": saved.SecretRef, "status": saved.Status}),
		CreatedAt: Now(),
	})

	// Return once: strip ciphertext from record in response, but include plaintext.
	saved.SecretCiphertext = ""
	return CreateWebhookSecretResponse{Secret: saved, SecretPlaintext: plaintext}, nil
}

func (s *Service) RotatePluginWebhookSecret(operator WebhookSecretOperator, id int64) (RotateWebhookSecretResponse, error) {
	cur, ok := s.repo.PluginWebhookSecretByID(id)
	if !ok || cur.ID == 0 {
		return RotateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404)
	}
	if cur.Status == domain.PluginWebhookSecretStatusRevoked {
		return RotateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_revoked", "revoked Secret 不能轮换").WithStatus(400)
	}

	pluginCode := strings.TrimSpace(cur.PluginCode)
	targetURL := strings.TrimSpace(cur.TargetURL)

	kr, okKR, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return RotateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_encrypt_key_invalid", "Secret 加密密钥配置不合法").WithStatus(500)
	}
	if !okKR {
		if strings.TrimSpace(os.Getenv("CMS_STORE")) == "memory" || strings.TrimSpace(os.Getenv("DEVHUB_E2E_TESTING")) == "1" {
			sum := sha256.Sum256([]byte("devhub-webhook-secret-test-key"))
			k := sum[:]
			kr = &pluginregistry.PluginConfigKeyring{
				CurrentKeyID:      "test-key",
				Keys:              map[string][]byte{"test-key": k},
				LegacyV1Supported: true,
			}
			okKR = true
		}
	}
	if !okKR {
		return RotateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_encryption_key_missing", "缺少 Webhook Secret 加密密钥，无法轮换 Secret").
			WithStatus(500).
			WithSuggestion("请配置环境变量 DEVHUB_PLUGIN_CONFIG_KEYS（推荐 JSON 格式）或 DEVHUB_PLUGIN_CONFIG_KEY_ID/DEVHUB_PLUGIN_CONFIG_KEY 后重试。")
	}

	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05")
	graceUntil := now.Add(webhookSecretGracePeriodDefault).Format("2006-01-02 15:04:05")

	// Create new active secret.
	newRef := webhookSecretRefPrefix + randomToken(18)
	plaintext := randomWebhookSecretPlaintext(webhookSecretBytesDefault)
	secretHash := sha256StringHex(plaintext)
	keyID, key := kr.CurrentKey()
	ciphertext, err := pluginregistry.EncryptStringV2(keyID, key, plaintext)
	if err != nil {
		return RotateWebhookSecretResponse{}, domain.NewPluginError("webhook_secret_encrypt_failed", "Secret 加密失败").WithStatus(500)
	}
	newRec := domain.PluginWebhookSecret{
		PluginCode:        pluginCode,
		TargetURL:         targetURL,
		SecretRef:         newRef,
		SecretCiphertext:  ciphertext,
		SecretHash:        secretHash,
		Version:           cur.Version + 1,
		Status:            domain.PluginWebhookSecretStatusActive,
		RotationGroup:     firstNonEmpty(strings.TrimSpace(cur.RotationGroup), pluginCode+":"+targetURL),
		PreviousSecretRef: strings.TrimSpace(cur.SecretRef),
		ActiveFrom:        nowStr,
		GraceUntil:        "",
		CreatedBy:         operator.ID,
		CreatedAt:         nowStr,
		RotatedAt:         nowStr,
		UpdatedAt:         nowStr,
	}
	savedNew, err := s.repo.AppendPluginWebhookSecret(newRec)
	if err != nil {
		return RotateWebhookSecretResponse{}, err
	}

	// Mark old active as previous and set grace window.
	if strings.TrimSpace(cur.Status) == domain.PluginWebhookSecretStatusActive {
		cur.Status = domain.PluginWebhookSecretStatusPrevious
		cur.GraceUntil = graceUntil
		cur.RotatedAt = nowStr
		cur.UpdatedAt = nowStr
		_, _ = s.repo.SavePluginWebhookSecret(cur)
	}

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.webhook.secret.rotated",
		Target:    "webhook-secrets#" + savedNew.SecretRef,
		Metadata: mustJSON(map[string]any{
			"plugin_code":       pluginCode,
			"target_url":        targetURL,
			"old_secret_ref":    cur.SecretRef,
			"new_secret_ref":    savedNew.SecretRef,
			"grace_until":       graceUntil,
			"previous_status":   cur.Status,
			"new_secret_status": savedNew.Status,
		}),
		CreatedAt: Now(),
	})

	// Return once: strip ciphertext, return plaintext.
	savedNew.SecretCiphertext = ""
	return RotateWebhookSecretResponse{
		OldSecretRef:    cur.SecretRef,
		NewSecretRef:    savedNew.SecretRef,
		GraceUntil:      graceUntil,
		NewSecretRecord: savedNew,
		SecretPlaintext: plaintext,
	}, nil
}

func (s *Service) DisablePluginWebhookSecret(operator WebhookSecretOperator, id int64) (domain.PluginWebhookSecret, error) {
	it, ok := s.repo.PluginWebhookSecretByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginWebhookSecret{}, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404)
	}
	if it.Status == domain.PluginWebhookSecretStatusRevoked {
		return domain.PluginWebhookSecret{}, domain.NewPluginError("webhook_secret_revoked", "revoked Secret 不能禁用").WithStatus(400)
	}
	now := Now()
	it.Status = domain.PluginWebhookSecretStatusDisabled
	it.UpdatedAt = now
	out, err := s.repo.SavePluginWebhookSecret(it)
	if err != nil {
		return domain.PluginWebhookSecret{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.webhook.secret.disabled",
		Target:    "webhook-secrets#" + out.SecretRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": out.PluginCode, "target_url": out.TargetURL, "secret_ref": out.SecretRef, "status": out.Status}),
		CreatedAt: Now(),
	})
	out.SecretCiphertext = ""
	return out, nil
}

func (s *Service) EnablePluginWebhookSecret(operator WebhookSecretOperator, id int64) (domain.PluginWebhookSecret, error) {
	it, ok := s.repo.PluginWebhookSecretByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginWebhookSecret{}, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404)
	}
	if it.Status != domain.PluginWebhookSecretStatusDisabled {
		return domain.PluginWebhookSecret{}, domain.NewPluginError("webhook_secret_enable_invalid", "仅 disabled Secret 可恢复").WithStatus(400)
	}
	now := Now()
	it.Status = domain.PluginWebhookSecretStatusActive
	it.ActiveFrom = now
	it.UpdatedAt = now
	// Disable any existing active for same target.
	if old, okOld := s.repo.ActivePluginWebhookSecret(it.PluginCode, it.TargetURL); okOld && old.ID > 0 && old.ID != it.ID {
		old.Status = domain.PluginWebhookSecretStatusDisabled
		old.UpdatedAt = now
		_, _ = s.repo.SavePluginWebhookSecret(old)
	}
	out, err := s.repo.SavePluginWebhookSecret(it)
	if err != nil {
		return domain.PluginWebhookSecret{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.webhook.secret.enabled",
		Target:    "webhook-secrets#" + out.SecretRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": out.PluginCode, "target_url": out.TargetURL, "secret_ref": out.SecretRef, "status": out.Status}),
		CreatedAt: Now(),
	})
	out.SecretCiphertext = ""
	return out, nil
}

func (s *Service) RevokePluginWebhookSecret(operator WebhookSecretOperator, id int64) (domain.PluginWebhookSecret, error) {
	it, ok := s.repo.PluginWebhookSecretByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginWebhookSecret{}, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404)
	}
	now := Now()
	it.Status = domain.PluginWebhookSecretStatusRevoked
	it.RevokedAt = now
	it.GraceUntil = ""
	it.UpdatedAt = now
	out, err := s.repo.SavePluginWebhookSecret(it)
	if err != nil {
		return domain.PluginWebhookSecret{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.webhook.secret.revoked",
		Target:    "webhook-secrets#" + out.SecretRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": out.PluginCode, "target_url": out.TargetURL, "secret_ref": out.SecretRef, "status": out.Status}),
		CreatedAt: Now(),
	})
	out.SecretCiphertext = ""
	return out, nil
}

// MarkExpiredPluginWebhookSecrets transitions previous->expired when grace_until passed.
// It is a lightweight sweep used by admin list/detail; later can be moved to cron/worker.
func (s *Service) MarkExpiredPluginWebhookSecrets(now time.Time, limit int) int {
	if s == nil || s.repo == nil {
		return 0
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	filter := domain.PluginWebhookSecretFilter{Status: domain.PluginWebhookSecretStatusPrevious, Page: 1, PageSize: limit}
	items, _, err := s.repo.PluginWebhookSecrets(filter)
	if err != nil {
		return 0
	}
	changed := 0
	for _, it := range items {
		if it.ID <= 0 {
			continue
		}
		if strings.TrimSpace(it.GraceUntil) == "" {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(it.GraceUntil), time.Local)
		if err != nil {
			continue
		}
		if t.After(now) {
			continue
		}
		it.Status = domain.PluginWebhookSecretStatusExpired
		it.UpdatedAt = Now()
		if _, err := s.repo.SavePluginWebhookSecret(it); err == nil {
			changed++
			s.repo.AppendAdminLog(domain.AdminLog{
				Site:      "admin",
				Actor:     "system",
				ActorType: "system",
				ActorID:   0,
				Action:    "plugin.webhook.secret.expired",
				Target:    "webhook-secrets#" + it.SecretRef,
				Metadata:  mustJSON(map[string]any{"plugin_code": it.PluginCode, "target_url": it.TargetURL, "secret_ref": it.SecretRef, "status": it.Status}),
				CreatedAt: Now(),
			})
		}
	}
	return changed
}

type resolvedWebhookSecret struct {
	Plaintext string
	Record    domain.PluginWebhookSecret
}

func (s *Service) resolveActiveWebhookSecret(pluginCode, targetURL string) (resolvedWebhookSecret, string, error) {
	sec, ok := s.repo.ActivePluginWebhookSecret(pluginCode, targetURL)
	if !ok || sec.ID == 0 {
		// For better operator feedback: if there is a secret record but it's not active,
		// return a more precise status than "missing".
		if latest, okLatest := s.repo.LatestPluginWebhookSecretForTarget(pluginCode, targetURL); okLatest && latest.ID > 0 {
			switch latest.Status {
			case domain.PluginWebhookSecretStatusDisabled:
				return resolvedWebhookSecret{Record: latest}, "secret_disabled", errors.New("secret disabled")
			case domain.PluginWebhookSecretStatusRevoked:
				return resolvedWebhookSecret{Record: latest}, "secret_revoked", errors.New("secret revoked")
			case domain.PluginWebhookSecretStatusExpired:
				return resolvedWebhookSecret{Record: latest}, "secret_expired", errors.New("secret expired")
			default:
				return resolvedWebhookSecret{Record: latest}, "secret_missing", errors.New("missing active secret")
			}
		}
		return resolvedWebhookSecret{}, "secret_missing", errors.New("missing active secret")
	}
	switch sec.Status {
	case domain.PluginWebhookSecretStatusActive:
	default:
		if sec.Status == domain.PluginWebhookSecretStatusDisabled {
			return resolvedWebhookSecret{Record: sec}, "secret_disabled", errors.New("secret disabled")
		}
		if sec.Status == domain.PluginWebhookSecretStatusRevoked {
			return resolvedWebhookSecret{Record: sec}, "secret_revoked", errors.New("secret revoked")
		}
		if sec.Status == domain.PluginWebhookSecretStatusExpired {
			return resolvedWebhookSecret{Record: sec}, "secret_expired", errors.New("secret expired")
		}
		return resolvedWebhookSecret{Record: sec}, "secret_missing", errors.New("secret not active")
	}

	kr, okKR, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return resolvedWebhookSecret{Record: sec}, "sign_failed", err
	}
	if !okKR {
		if strings.TrimSpace(os.Getenv("CMS_STORE")) == "memory" || strings.TrimSpace(os.Getenv("DEVHUB_E2E_TESTING")) == "1" {
			sum := sha256.Sum256([]byte("devhub-webhook-secret-test-key"))
			k := sum[:]
			kr = &pluginregistry.PluginConfigKeyring{
				CurrentKeyID:      "test-key",
				Keys:              map[string][]byte{"test-key": k},
				LegacyV1Supported: true,
			}
			okKR = true
		}
	}
	if !okKR {
		return resolvedWebhookSecret{Record: sec}, "sign_failed", errors.New("missing encryption keyring")
	}
	plain, _, err := pluginregistry.DecryptStringWithKeyring(kr, sec.SecretCiphertext)
	if err != nil {
		return resolvedWebhookSecret{Record: sec}, "sign_failed", err
	}
	return resolvedWebhookSecret{Plaintext: plain, Record: sec}, "signed", nil
}

func signWebhookHMACSHA256(secret []byte, signingString string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signingString))
	return hex.EncodeToString(mac.Sum(nil))
}

func randomWebhookSecretPlaintext(nbytes int) string {
	if nbytes <= 0 {
		nbytes = webhookSecretBytesDefault
	}
	buf := make([]byte, nbytes)
	_, _ = rand.Read(buf)
	// base64 for easy copy/paste; plugin side should treat as opaque bytes.
	return base64.StdEncoding.EncodeToString(buf)
}

func randomToken(n int) string {
	if n <= 0 {
		n = 16
	}
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func sha256StringHex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
