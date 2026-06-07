package service

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type SecretOperator struct {
	Type string // admin_user|system
	ID   int64
	Name string
}

const (
	secretCenterNamespaceExternalService = "external_service"
	secretCenterNamespaceWebhook         = "webhook"
	secretCenterNamespaceCallback        = "callback"
	secretCenterNamespacePluginConfig    = "plugin_config"
)

func secretCenterKeyMissingError() *domain.APIError {
	return domain.NewPluginError("secret_center_encryption_key_missing", "当前系统未配置插件敏感配置加密密钥，因此无法保存敏感配置。").
		WithStatus(500).
		WithSuggestion(strings.TrimSpace(`
当前系统未配置插件敏感配置加密密钥，因此无法保存敏感配置。

需要配置以下任一方式后重启 DevHub：

方式一，推荐 JSON：
DEVHUB_PLUGIN_CONFIG_KEYS='[{"id":"local-v1","key":"base64-xxx","primary":true}]'

方式二，兼容单 key：
DEVHUB_PLUGIN_CONFIG_KEY_ID=local-v1
DEVHUB_PLUGIN_CONFIG_KEY=base64-xxx

注意：该密钥是启动级 root key，不能在后台页面中创建或保存。
`)).
		WithDiagnostic(map[string]any{
			"type":               "startup_key_missing",
			"key_status_api":     "/api/v1/admin/plugins/config-keys/status",
			"restart_required":   true,
			"env_keys_read_only": true,
		})
}

func secretCenterKeyMissingErrorForRead() *domain.APIError {
	// Keep code stable; only message differs.
	e := secretCenterKeyMissingError()
	e.Message = "缺少 SecretCenter 加密密钥，无法解密敏感值"
	return e
}

func (s *Service) secretCenterKeyringOrError(write bool) (*pluginregistry.PluginConfigKeyring, error) {
	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return nil, domain.NewPluginError("secret_center_encrypt_key_invalid", "SecretCenter 加密密钥配置不合法").WithStatus(500)
	}
	if !ok || kr == nil {
		// Keep behavior consistent with existing sensitive-config flows:
		// - MySQL/production-like environments should require explicit root key.
		// - Memory/E2E can fallback to a stable ephemeral key for local convenience.
		if strings.TrimSpace(os.Getenv("CMS_STORE")) == "memory" || strings.TrimSpace(os.Getenv("DEVHUB_E2E_TESTING")) == "1" {
			sum := sha256.Sum256([]byte("devhub-secret-center-test-key"))
			k := sum[:]
			return &pluginregistry.PluginConfigKeyring{
				CurrentKeyID:      "test-key",
				Keys:              map[string][]byte{"test-key": k},
				LegacyV1Supported: true,
			}, nil
		}
		if write {
			return nil, secretCenterKeyMissingError()
		}
		return nil, secretCenterKeyMissingErrorForRead()
	}
	return kr, nil
}

func (s *Service) CreateSecret(operator SecretOperator, req domain.SecretCreateRequest) (domain.SecretRefRecord, error) {
	ns := strings.TrimSpace(req.Namespace)
	name := strings.TrimSpace(req.Name)
	value := strings.TrimSpace(req.Value)
	if ns == "" || name == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "namespace/name 必填").WithStatus(http.StatusBadRequest)
	}
	if value == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "value 不能为空").WithStatus(http.StatusBadRequest)
	}
	ref, err := domain.BuildSecretRef(ns, name)
	if err != nil {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid_ref", "secret_ref 不合法").WithStatus(http.StatusBadRequest).WithDetail("ref", strings.TrimSpace(ref))
	}
	if _, ok := s.repo.SecretRefByRef(ref); ok {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_already_exists", "secret_ref 已存在").WithStatus(http.StatusConflict).WithDetail("ref", ref)
	}
	kr, err := s.secretCenterKeyringOrError(true)
	if err != nil {
		return domain.SecretRefRecord{}, err
	}
	keyID, key := kr.CurrentKey()
	cipher, err := pluginregistry.EncryptStringV2(keyID, key, value)
	if err != nil {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_encrypt_failed", "SecretCenter 加密失败").WithStatus(500)
	}
	now := Now()
	record := domain.SecretRefRecord{
		Ref:            ref,
		Namespace:      ns,
		Name:           name,
		KeyID:          keyID,
		EncryptedValue: cipher,
		Status:         domain.SecretRefStatusActive,
		Description:    strings.TrimSpace(req.Description),
		RotatedAt:      now,
		CreatedBy:      operator.ID,
		UpdatedBy:      operator.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	saved, err := s.repo.AppendSecretRef(record)
	if err != nil {
		return domain.SecretRefRecord{}, err
	}
	// Never expose ciphertext via API.
	saved.EncryptedValue = ""
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: firstNonEmpty(operator.Type, "admin_user"),
		ActorID:   operator.ID,
		Action:    "secret_center.secret.created",
		Target:    "secret_refs#" + ref,
		Metadata:  mustJSON(s.secretAuditMetadata(saved, map[string]any{"key_id": keyID, "status": saved.Status})),
		CreatedAt: Now(),
	})
	return saved, nil
}

func (s *Service) UpdateSecret(operator SecretOperator, req domain.SecretUpdateRequest) (domain.SecretRefRecord, error) {
	ref := strings.TrimSpace(req.Ref)
	value := strings.TrimSpace(req.Value)
	if ref == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "ref 必填").WithStatus(http.StatusBadRequest)
	}
	if value == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "value 不能为空").WithStatus(http.StatusBadRequest)
	}
	if _, _, err := domain.ParseSecretRef(ref); err != nil {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid_ref", "secret_ref 不合法").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	cur, ok := s.repo.SecretRefByRef(ref)
	if !ok || cur.ID == 0 {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_not_found", "secret_ref 不存在").WithStatus(http.StatusNotFound).WithDetail("ref", ref)
	}
	kr, err := s.secretCenterKeyringOrError(true)
	if err != nil {
		return domain.SecretRefRecord{}, err
	}
	keyID, key := kr.CurrentKey()
	cipher, err := pluginregistry.EncryptStringV2(keyID, key, value)
	if err != nil {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_encrypt_failed", "SecretCenter 加密失败").WithStatus(500)
	}
	now := Now()
	cur.KeyID = keyID
	cur.EncryptedValue = cipher
	cur.Status = domain.SecretRefStatusActive
	cur.RotatedAt = now
	cur.UpdatedBy = operator.ID
	cur.UpdatedAt = now
	saved, err := s.repo.SaveSecretRef(cur)
	if err != nil {
		return domain.SecretRefRecord{}, err
	}
	saved.EncryptedValue = ""
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: firstNonEmpty(operator.Type, "admin_user"),
		ActorID:   operator.ID,
		Action:    "secret_center.secret.updated",
		Target:    "secret_refs#" + ref,
		Metadata:  mustJSON(s.secretAuditMetadata(saved, map[string]any{"key_id": keyID, "status": saved.Status})),
		CreatedAt: Now(),
	})
	return saved, nil
}

func (s *Service) DisableSecret(operator SecretOperator, req domain.SecretStatusRequest) (domain.SecretRefRecord, error) {
	ref := strings.TrimSpace(req.Ref)
	if ref == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "ref 必填").WithStatus(http.StatusBadRequest)
	}
	if preview, err := s.SecretImpactPreview(ref, "disable"); err != nil {
		s.appendSecretStatusAudit(operator, "secret_center.secret.disable.failed", ref, "", err.Error())
		return domain.SecretRefRecord{}, err
	} else if !preview.Allowed {
		s.appendSecretStatusAudit(operator, "secret_center.secret.disable.failed", ref, preview.Status, preview.Message)
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_disable_blocked", firstNonEmpty(preview.Message, "当前 Secret 不允许禁用")).WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	if out, ok, err := s.disableSourceSecret(operator, ref); ok || err != nil {
		if err != nil {
			s.appendSecretStatusAudit(operator, "secret_center.secret.disable.failed", ref, out.Status, err.Error())
		} else {
			s.appendSecretStatusAudit(operator, "secret_center.secret.disabled", ref, out.Status, "")
		}
		return out, err
	}
	cur, ok := s.repo.SecretRefByRef(ref)
	if !ok || cur.ID == 0 {
		s.appendSecretStatusAudit(operator, "secret_center.secret.disable.failed", ref, "", "secret_ref 不存在")
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_not_found", "secret_ref 不存在").WithStatus(http.StatusNotFound).WithDetail("ref", ref)
	}
	if cur.Status != "" && cur.Status != domain.SecretRefStatusActive {
		s.appendSecretStatusAudit(operator, "secret_center.secret.disable.failed", ref, cur.Status, "只有 active Secret 可以禁用")
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_status_conflict", "只有 active Secret 可以禁用").WithStatus(http.StatusBadRequest).WithDetail("ref", ref).WithDetail("status", cur.Status)
	}
	cur.Status = domain.SecretRefStatusDisabled
	cur.UpdatedBy = operator.ID
	cur.UpdatedAt = Now()
	saved, err := s.repo.SaveSecretRef(cur)
	if err != nil {
		s.appendSecretStatusAudit(operator, "secret_center.secret.disable.failed", ref, cur.Status, err.Error())
		return domain.SecretRefRecord{}, err
	}
	saved.EncryptedValue = ""
	s.appendSecretStatusAudit(operator, "secret_center.secret.disabled", ref, saved.Status, "")
	return saved, nil
}

func (s *Service) RevokeSecret(operator SecretOperator, req domain.SecretStatusRequest) (domain.SecretRefRecord, error) {
	ref := strings.TrimSpace(req.Ref)
	if ref == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "ref 必填").WithStatus(http.StatusBadRequest)
	}
	if strings.TrimSpace(req.ConfirmRef) != ref && !req.StrongConfirm {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, "", "吊销 Secret 需要输入完整 ref 或勾选强确认")
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_revoke_confirm_required", "吊销 Secret 需要输入完整 ref 或勾选强确认").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	if preview, err := s.SecretImpactPreview(ref, "revoke"); err != nil {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, "", err.Error())
		return domain.SecretRefRecord{}, err
	} else if !preview.Allowed {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, preview.Status, preview.Message)
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_revoke_blocked", firstNonEmpty(preview.Message, "当前 Secret 不允许吊销")).WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	if out, ok, err := s.revokeSourceSecret(operator, ref); ok || err != nil {
		if err != nil {
			s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, out.Status, err.Error())
		} else {
			s.appendSecretStatusAudit(operator, "secret_center.secret.revoked", ref, out.Status, "")
		}
		return out, err
	}
	cur, ok := s.repo.SecretRefByRef(ref)
	if !ok || cur.ID == 0 {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, "", "secret_ref 不存在")
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_not_found", "secret_ref 不存在").WithStatus(http.StatusNotFound).WithDetail("ref", ref)
	}
	if cur.Status == domain.SecretRefStatusRevoked {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, cur.Status, "revoked Secret 不能重复吊销")
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_status_conflict", "revoked Secret 不能重复吊销").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	if cur.Status != "" && cur.Status != domain.SecretRefStatusActive && cur.Status != domain.SecretRefStatusDisabled {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, cur.Status, "只有 active 或 disabled Secret 可以吊销")
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_status_conflict", "只有 active 或 disabled Secret 可以吊销").WithStatus(http.StatusBadRequest).WithDetail("ref", ref).WithDetail("status", cur.Status)
	}
	cur.Status = domain.SecretRefStatusRevoked
	cur.UpdatedBy = operator.ID
	cur.UpdatedAt = Now()
	saved, err := s.repo.SaveSecretRef(cur)
	if err != nil {
		s.appendSecretStatusAudit(operator, "secret_center.secret.revoke.failed", ref, cur.Status, err.Error())
		return domain.SecretRefRecord{}, err
	}
	saved.EncryptedValue = ""
	s.appendSecretStatusAudit(operator, "secret_center.secret.revoked", ref, saved.Status, "")
	return saved, nil
}

func (s *Service) appendSecretStatusAudit(operator SecretOperator, action, ref, status, reason string) {
	ref = strings.TrimSpace(ref)
	meta := map[string]any{"ref": ref}
	if rec, err := s.GetSecretMetadata(ref); err == nil {
		meta = s.secretAuditMetadata(rec, meta)
	}
	if strings.TrimSpace(status) != "" {
		meta["status"] = strings.TrimSpace(status)
	}
	if strings.TrimSpace(reason) != "" {
		meta["reason"] = strings.TrimSpace(reason)
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: firstNonEmpty(operator.Type, "admin_user"),
		ActorID:   operator.ID,
		Action:    action,
		Target:    "secret_refs#" + ref,
		Metadata:  mustJSON(meta),
		CreatedAt: Now(),
	})
}

func (s *Service) secretAuditMetadata(rec domain.SecretRefRecord, base map[string]any) map[string]any {
	if base == nil {
		base = map[string]any{}
	}
	rec = s.enrichSecretRefForDisplay(rec)
	ref := strings.TrimSpace(rec.Ref)
	if ref != "" {
		base["ref"] = ref
	}
	if strings.TrimSpace(rec.Namespace) != "" {
		base["namespace"] = strings.TrimSpace(rec.Namespace)
	}
	if strings.TrimSpace(rec.Name) != "" {
		base["name"] = strings.TrimSpace(rec.Name)
	}
	if strings.TrimSpace(rec.UsageType) != "" {
		base["usage_type"] = strings.TrimSpace(rec.UsageType)
	}
	source := s.secretSourceInfo(rec)
	if strings.TrimSpace(source.Type) != "" {
		base["source_type"] = strings.TrimSpace(source.Type)
	}
	if strings.TrimSpace(source.SourceID) != "" {
		base["source_id"] = strings.TrimSpace(source.SourceID)
	}
	if strings.TrimSpace(source.SourceCode) != "" {
		base["source_code"] = strings.TrimSpace(source.SourceCode)
	}
	if strings.TrimSpace(source.PluginCode) != "" {
		base["plugin_code"] = strings.TrimSpace(source.PluginCode)
	}
	if strings.TrimSpace(source.ConfigEntry) != "" {
		base["config_entry"] = strings.TrimSpace(source.ConfigEntry)
	}
	return base
}

func (s *Service) GetSecretMetadata(ref string) (domain.SecretRefRecord, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "ref 必填").WithStatus(http.StatusBadRequest)
	}
	it, ok := s.repo.SecretRefByRef(ref)
	if !ok || it.ID == 0 {
		if virtual, vok := s.virtualSecretRefByRef(ref); vok {
			return virtual, nil
		}
		return domain.SecretRefRecord{}, domain.NewPluginError("secret_center_not_found", "secret_ref 不存在").WithStatus(http.StatusNotFound).WithDetail("ref", ref)
	}
	return s.enrichSecretRefForDisplay(it), nil
}

func (s *Service) GetSecretDetail(ref string) (domain.SecretDetailResponse, error) {
	record, err := s.GetSecretMetadata(ref)
	if err != nil {
		return domain.SecretDetailResponse{}, err
	}
	source := s.secretSourceInfo(record)
	usages := s.SecretUsages(record.Ref)
	return domain.SecretDetailResponse{
		Record: record,
		Source: source,
		Usages: usages,
		SafetyNotes: []string{
			"该详情只展示 ref、状态、来源和脱敏元数据，不返回明文 token / secret / Authorization。",
			"如需轮换，请跳转到来源治理入口重新写入新凭据。",
		},
	}, nil
}

func (s *Service) ListSecretMetadata(filter domain.SecretRefFilter) (domain.SecretRefListResponse, error) {
	items, total, err := s.repo.SecretRefs(filter)
	if err != nil {
		return domain.SecretRefListResponse{}, err
	}
	f := filter.Normalize()
	for i := range items {
		items[i] = s.enrichSecretRefForDisplay(items[i])
	}
	if strings.TrimSpace(filter.Namespace) == "" {
		items = append(items, s.virtualWebhookSecretRefs()...)
		items = append(items, s.virtualCallbackTokenRefs()...)
		total = len(items)
	}
	return domain.SecretRefListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     f.Page,
			PageSize: f.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) enrichSecretRefForDisplay(rec domain.SecretRefRecord) domain.SecretRefRecord {
	rec.EncryptedValue = ""
	status := strings.TrimSpace(rec.Status)
	if status == "" {
		status = domain.SecretRefStatusActive
		rec.Status = status
	}
	rec.Available = status == domain.SecretRefStatusActive
	ns, name := rec.Namespace, rec.Name
	if parsedNS, parsedName, err := domain.ParseSecretRef(rec.Ref); err == nil {
		if strings.TrimSpace(ns) == "" {
			ns = parsedNS
			rec.Namespace = parsedNS
		}
		if strings.TrimSpace(name) == "" {
			name = parsedName
			rec.Name = parsedName
		}
	}
	rec.Type = secretRefDisplayType(ns, name)
	rec.AssociatedWith = secretRefAssociatedObject(ns, name)
	rec.DisplayName = secretRefDisplayName(ns, name, rec.Description)
	rec.Usage = secretRefUsage(ns, name)
	rec.UsageType = secretRefUsageType(ns, name)
	rec.SourceType = s.secretSourceType(rec)
	if rec.SourceType == "test" {
		rec.UsageType = "test_fixture_seed"
	}
	rec.SourceID, rec.SourceCode = secretRefSourceIdentifiers(rec)
	if masked, ok := s.maskSecretRefValue(rec.Ref); ok {
		rec.MaskedValue = masked
	}
	return rec
}

func secretRefUsageType(namespace, name string) string {
	ns := strings.ToLower(strings.TrimSpace(namespace))
	lowerName := strings.ToLower(strings.TrimSpace(name))
	switch {
	case ns == secretCenterNamespaceExternalService && strings.HasSuffix(lowerName, "/token"):
		return "external_service_token"
	case ns == secretCenterNamespaceWebhook:
		return "webhook_secret"
	case ns == secretCenterNamespaceCallback:
		return "callback_token"
	case ns == secretCenterNamespacePluginConfig:
		return "plugin_config_sensitive_field"
	default:
		return "other_unknown"
	}
}

func secretRefSourceIdentifiers(rec domain.SecretRefRecord) (string, string) {
	ns := strings.ToLower(strings.TrimSpace(rec.Namespace))
	if ns == "" && domain.IsSecretRef(rec.Ref) {
		ns, _, _ = domain.ParseSecretRef(rec.Ref)
		ns = strings.ToLower(strings.TrimSpace(ns))
	}
	associated := secretRefAssociatedObject(ns, rec.Name)
	switch ns {
	case secretCenterNamespaceExternalService, secretCenterNamespacePluginConfig:
		return associated, associated
	case secretCenterNamespaceWebhook, secretCenterNamespaceCallback:
		return strings.TrimSpace(rec.Ref), associated
	default:
		if strings.TrimSpace(rec.Ref) != "" {
			return strings.TrimSpace(rec.Ref), associated
		}
		return associated, associated
	}
}

func secretRefDisplayType(namespace, name string) string {
	ns := strings.ToLower(strings.TrimSpace(namespace))
	lowerName := strings.ToLower(strings.TrimSpace(name))
	switch {
	case ns == secretCenterNamespaceExternalService && strings.HasSuffix(lowerName, "/token"):
		return "external_service token"
	case ns == secretCenterNamespaceExternalService:
		return "external_service secret"
	case ns == secretCenterNamespaceWebhook:
		return "Webhook Secret"
	case ns == secretCenterNamespaceCallback:
		return "Callback Token"
	case ns == secretCenterNamespacePluginConfig:
		return "plugin config sensitive field"
	default:
		return "secret"
	}
}

func secretRefAssociatedObject(namespace, name string) string {
	ns := strings.ToLower(strings.TrimSpace(namespace))
	parts := strings.Split(strings.Trim(strings.TrimSpace(name), "/"), "/")
	switch {
	case ns == secretCenterNamespaceExternalService && len(parts) > 0:
		return parts[0]
	case ns == secretCenterNamespacePluginConfig && len(parts) > 0:
		return parts[0]
	default:
		return strings.TrimSpace(namespace)
	}
}

func secretRefDisplayName(namespace, name, description string) string {
	if strings.TrimSpace(description) != "" {
		return strings.TrimSpace(description)
	}
	ns := strings.ToLower(strings.TrimSpace(namespace))
	associated := secretRefAssociatedObject(namespace, name)
	switch {
	case ns == secretCenterNamespaceExternalService && associated != "":
		return associated + " 外部服务 Token"
	case ns == secretCenterNamespaceWebhook:
		return "Webhook Secret"
	case ns == secretCenterNamespaceCallback:
		return "Callback Token"
	case ns == secretCenterNamespacePluginConfig && associated != "":
		return associated + " 插件配置敏感字段"
	default:
		return firstNonEmpty(strings.TrimSpace(name), strings.TrimSpace(namespace), "Secret")
	}
}

func secretRefUsage(namespace, name string) string {
	ns := strings.ToLower(strings.TrimSpace(namespace))
	associated := secretRefAssociatedObject(namespace, name)
	switch {
	case ns == secretCenterNamespaceExternalService && associated != "":
		return associated + " external_service 投递鉴权"
	case ns == secretCenterNamespaceWebhook:
		return "DevHub 向插件服务发送 Webhook 时签名鉴权"
	case ns == secretCenterNamespaceCallback:
		return "插件服务回调 DevHub Core API 鉴权"
	case ns == secretCenterNamespacePluginConfig && associated != "":
		return associated + " 插件配置敏感字段"
	default:
		return "运行时敏感配置引用"
	}
}

func (s *Service) maskSecretRefValue(ref string) (string, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", false
	}
	rec, ok := s.repo.SecretRefByRef(ref)
	if !ok || rec.ID == 0 || strings.TrimSpace(rec.EncryptedValue) == "" {
		return "", false
	}
	kr, err := s.secretCenterKeyringOrError(false)
	if err != nil {
		return "", false
	}
	plain, _, derr := pluginregistry.DecryptStringWithKeyring(kr, rec.EncryptedValue)
	if derr != nil || strings.TrimSpace(plain) == "" {
		return "", false
	}
	runes := []rune(strings.TrimSpace(plain))
	if len(runes) <= 4 {
		return "******", true
	}
	return "******" + string(runes[len(runes)-4:]), true
}

func (s *Service) virtualSecretRefByRef(ref string) (domain.SecretRefRecord, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return domain.SecretRefRecord{}, false
	}
	if wh, ok := s.repo.PluginWebhookSecretByRef(ref); ok && wh.ID > 0 {
		return s.virtualWebhookSecretRef(wh), true
	}
	if cb, ok := s.repo.PluginCallbackTokenByRef(ref); ok && cb.ID > 0 {
		return s.virtualCallbackTokenRef(cb), true
	}
	return domain.SecretRefRecord{}, false
}

func (s *Service) virtualWebhookSecretRefs() []domain.SecretRefRecord {
	items, _, err := s.repo.PluginWebhookSecrets(domain.PluginWebhookSecretFilter{Page: 1, PageSize: 100})
	if err != nil {
		return nil
	}
	out := make([]domain.SecretRefRecord, 0, len(items))
	for _, it := range items {
		out = append(out, s.virtualWebhookSecretRef(it))
	}
	return out
}

func (s *Service) virtualWebhookSecretRef(it domain.PluginWebhookSecret) domain.SecretRefRecord {
	status := strings.TrimSpace(it.Status)
	if status == "" {
		status = domain.PluginWebhookSecretStatusActive
	}
	rec := domain.SecretRefRecord{
		ID:             it.ID,
		Ref:            strings.TrimSpace(it.SecretRef),
		Namespace:      secretCenterNamespaceWebhook,
		Name:           strings.TrimSpace(it.SecretRef),
		DisplayName:    firstNonEmpty(strings.TrimSpace(it.PluginCode)+" Webhook Secret", "Webhook Secret"),
		Type:           "Webhook Secret",
		Usage:          "DevHub 向插件服务发送 Webhook 时签名鉴权",
		UsageType:      "webhook_secret",
		SourceType:     secretCenterNamespaceWebhook,
		SourceID:       strings.TrimSpace(it.SecretRef),
		SourceCode:     strings.TrimSpace(it.PluginCode),
		AssociatedWith: firstNonEmpty(strings.TrimSpace(it.PluginCode), strings.TrimSpace(it.TargetURL)),
		Status:         status,
		Description:    redactSensitiveURLForSecretCenter(it.TargetURL),
		LastUsedAt:     it.LastUsedAt,
		RotatedAt:      it.RotatedAt,
		CreatedBy:      it.CreatedBy,
		CreatedAt:      it.CreatedAt,
		UpdatedAt:      it.UpdatedAt,
	}
	rec.Available = status == domain.PluginWebhookSecretStatusActive || status == domain.PluginWebhookSecretStatusPrevious
	rec.MaskedValue = maskOpaqueRefSuffix(rec.Ref)
	return rec
}

func (s *Service) virtualCallbackTokenRefs() []domain.SecretRefRecord {
	items, _, err := s.repo.PluginCallbackTokens(domain.PluginCallbackTokenFilter{Page: 1, PageSize: 100})
	if err != nil {
		return nil
	}
	out := make([]domain.SecretRefRecord, 0, len(items))
	for _, it := range items {
		out = append(out, s.virtualCallbackTokenRef(it))
	}
	return out
}

func (s *Service) virtualCallbackTokenRef(it domain.PluginCallbackToken) domain.SecretRefRecord {
	status := strings.TrimSpace(it.Status)
	if status == "" {
		status = domain.PluginCallbackTokenStatusActive
	}
	rec := domain.SecretRefRecord{
		ID:             it.ID,
		Ref:            strings.TrimSpace(it.TokenRef),
		Namespace:      secretCenterNamespaceCallback,
		Name:           strings.TrimSpace(it.TokenRef),
		DisplayName:    firstNonEmpty(strings.TrimSpace(it.Name), strings.TrimSpace(it.PluginCode)+" Callback Token", "Callback Token"),
		Type:           "Callback Token",
		Usage:          "插件服务回调 DevHub Core API 鉴权",
		UsageType:      "callback_token",
		SourceType:     secretCenterNamespaceCallback,
		SourceID:       strings.TrimSpace(it.TokenRef),
		SourceCode:     strings.TrimSpace(it.PluginCode),
		AssociatedWith: strings.TrimSpace(it.PluginCode),
		Status:         status,
		Description:    strings.TrimSpace(it.Name),
		LastUsedAt:     it.LastUsedAt,
		RotatedAt:      it.RotatedAt,
		CreatedBy:      it.CreatedBy,
		CreatedAt:      it.CreatedAt,
		UpdatedAt:      it.UpdatedAt,
	}
	rec.Available = status == domain.PluginCallbackTokenStatusActive
	rec.MaskedValue = maskOpaqueRefSuffix(rec.Ref)
	return rec
}

func maskOpaqueRefSuffix(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	runes := []rune(ref)
	if len(runes) <= 4 {
		return "******"
	}
	return "******" + string(runes[len(runes)-4:])
}

func redactSensitiveURLForSecretCenter(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		u.User = url.User("redacted")
	}
	q := u.Query()
	for key := range q {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "authorization") || strings.Contains(lower, "password") || strings.Contains(lower, "credential") {
			q.Set(key, "******")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Service) secretSourceInfo(rec domain.SecretRefRecord) domain.SecretSourceInfo {
	category := s.secretSourceType(rec)
	source := domain.SecretSourceInfo{Type: category, Label: "未知来源", TestData: s.isTestSecretRecord(rec)}
	if source.TestData {
		source.Label = "测试数据"
		source.JumpDisabledReason = "测试/fixture/seed Secret 默认不允许在 SecretCenter 执行危险操作；请到测试数据清理流程处理。"
		source.RotationDisabledNote = "测试数据不支持在此处轮换。"
		return source
	}
	switch category {
	case secretCenterNamespaceExternalService:
		pluginCode := externalServicePluginCodeFromSecret(rec)
		source.Label = "external_service"
		source.SourceID = pluginCode
		source.SourceCode = pluginCode
		source.PluginCode = pluginCode
		source.ConfigEntry = "plugin_external_services.token_ref"
		source.ManagementPage = "/admin-next/plugins/overview?tab=list&plugin_code=" + pluginCode + "&detail_tab=runtime"
		source.ManagementQueryHint = "detail_tab=runtime"
		source.RotationTarget = "插件 external_service 配置"
		source.CanJump = pluginCode != ""
		if p, ok := s.PluginByCode(pluginCode); ok {
			source.PluginName = p.Name
		}
	case secretCenterNamespaceWebhook:
		if wh, ok := s.repo.PluginWebhookSecretByRef(rec.Ref); ok && wh.ID > 0 {
			source.Label = "Webhook Secret"
			source.SourceID = wh.SecretRef
			source.SourceCode = wh.PluginCode
			source.PluginCode = wh.PluginCode
			source.ConfigEntry = "plugin_webhook_secrets.secret_ref"
			source.ManagementPage = "/admin-next/plugins/webhooks?tab=secrets&sec_ref=" + wh.SecretRef + "&sec_plugin_code=" + wh.PluginCode
			source.ManagementQueryHint = "tab=secrets"
			source.RotationTarget = "Webhook 密钥轮换流程"
			source.CanJump = true
			if p, ok := s.PluginByCode(wh.PluginCode); ok {
				source.PluginName = p.Name
			}
		}
	case secretCenterNamespaceCallback:
		if cb, ok := s.repo.PluginCallbackTokenByRef(rec.Ref); ok && cb.ID > 0 {
			source.Label = "Callback Token"
			source.SourceID = cb.TokenRef
			source.SourceCode = cb.PluginCode
			source.PluginCode = cb.PluginCode
			source.ConfigEntry = "plugin_callback_tokens.token_ref"
			source.ManagementPage = "/admin-next/plugins/webhooks?tab=callback_tokens&cbtk_plugin_code=" + cb.PluginCode
			source.ManagementQueryHint = "tab=callback_tokens"
			source.RotationTarget = "Callback Token 轮换流程"
			source.CanJump = true
			if p, ok := s.PluginByCode(cb.PluginCode); ok {
				source.PluginName = p.Name
			}
		}
	case secretCenterNamespacePluginConfig:
		pluginCode := externalServicePluginCodeFromSecret(rec)
		source.Label = "插件配置敏感字段"
		source.SourceID = pluginCode
		source.SourceCode = pluginCode
		source.PluginCode = pluginCode
		source.ConfigEntry = "plugin_configs.config_json sensitive field"
		source.ManagementPage = "/admin-next/plugins/overview?tab=config&plugin_code=" + pluginCode
		source.ManagementQueryHint = "tab=config"
		source.RotationTarget = "插件配置页"
		source.CanJump = pluginCode != ""
		if p, ok := s.PluginByCode(pluginCode); ok {
			source.PluginName = p.Name
		}
	default:
		source.JumpDisabledReason = "暂无法定位来源配置。"
		source.RotationDisabledNote = "来源未知，无法安全轮换。"
	}
	return source
}

func (s *Service) SecretUsages(ref string) []domain.SecretUsageRelationship {
	rec, err := s.GetSecretMetadata(ref)
	if err != nil {
		return []domain.SecretUsageRelationship{{
			Type:              "other",
			Label:             "未知来源",
			SecretRef:         strings.TrimSpace(ref),
			Unresolved:        true,
			UnresolvedMessage: "暂无法定位来源配置，请复制 ref 后查看审计。",
		}}
	}
	category := s.secretSourceType(rec)
	switch category {
	case secretCenterNamespaceExternalService:
		return s.externalServiceSecretUsages(rec)
	case secretCenterNamespaceWebhook:
		return s.webhookSecretUsages(rec)
	case secretCenterNamespaceCallback:
		return s.callbackTokenUsages(rec)
	case secretCenterNamespacePluginConfig:
		return s.pluginConfigSecretUsages(rec)
	case "test":
		return []domain.SecretUsageRelationship{{
			Type:        "test_fixture_seed",
			Label:       "测试 / fixture / seed",
			UsageType:   "test_fixture_seed",
			SourceType:  "test",
			SourceID:    rec.Ref,
			SourceCode:  secretRefAssociatedObject(rec.Namespace, rec.Name),
			Status:      rec.Status,
			SecretRef:   rec.Ref,
			ConfigEntry: "test/fixture/seed",
		}}
	}
	return []domain.SecretUsageRelationship{{
		Type:              "other",
		Label:             "未知来源",
		SecretRef:         rec.Ref,
		Unresolved:        true,
		UnresolvedMessage: "暂无法定位来源配置，请复制 ref 后查看审计。",
	}}
}

func (s *Service) externalServiceSecretUsages(rec domain.SecretRefRecord) []domain.SecretUsageRelationship {
	pluginCode := externalServicePluginCodeFromSecret(rec)
	cfg, ok := s.PluginExternalServiceConfig(pluginCode)
	if !ok || strings.TrimSpace(cfg.TokenRef) != strings.TrimSpace(rec.Ref) {
		return []domain.SecretUsageRelationship{{
			Type:              secretCenterNamespaceExternalService,
			Label:             "external_service",
			PluginCode:        pluginCode,
			SecretRef:         rec.Ref,
			Unresolved:        true,
			UnresolvedMessage: "未找到绑定该 token_ref 的 external_service 运行配置。",
		}}
	}
	rel := domain.SecretUsageRelationship{
		Type:          secretCenterNamespaceExternalService,
		Label:         "external_service token",
		UsageType:     "external_service_token",
		SourceType:    secretCenterNamespaceExternalService,
		SourceID:      pluginCode,
		SourceCode:    pluginCode,
		PluginCode:    pluginCode,
		ServiceName:   "external_service",
		EndpointURL:   cfg.EndpointURL,
		Enabled:       cfg.Enabled,
		CurrentHealth: firstNonEmpty(cfg.LastHealthStatus, cfg.Status, "unknown"),
		LastSuccessAt: cfg.LastSuccessAt,
		LastFailureAt: cfg.LastFailureAt,
		ConfigEntry:   "plugin_external_services.token_ref",
		ManagementPage: "/admin-next/plugins/overview?tab=list&plugin_code=" + pluginCode +
			"&detail_tab=runtime",
		Status:    cfg.Status,
		SecretRef: rec.Ref,
		TokenRef:  cfg.TokenRef,
	}
	if p, ok := s.PluginByCode(pluginCode); ok {
		rel.PluginName = p.Name
	}
	return []domain.SecretUsageRelationship{rel}
}

func (s *Service) webhookSecretUsages(rec domain.SecretRefRecord) []domain.SecretUsageRelationship {
	wh, ok := s.repo.PluginWebhookSecretByRef(rec.Ref)
	if !ok || wh.ID == 0 {
		return []domain.SecretUsageRelationship{{
			Type:              secretCenterNamespaceWebhook,
			Label:             "Webhook Secret",
			SecretRef:         rec.Ref,
			Unresolved:        true,
			UnresolvedMessage: "未找到对应 Webhook Secret 记录。",
		}}
	}
	rel := domain.SecretUsageRelationship{
		Type:           secretCenterNamespaceWebhook,
		Label:          "Webhook Secret",
		UsageType:      "webhook_secret",
		SourceType:     secretCenterNamespaceWebhook,
		SourceID:       wh.SecretRef,
		SourceCode:     wh.PluginCode,
		PluginCode:     wh.PluginCode,
		TargetURL:      redactSensitiveURLForSecretCenter(wh.TargetURL),
		Status:         wh.Status,
		SecretRef:      wh.SecretRef,
		ConfigEntry:    "plugin_webhook_secrets.secret_ref",
		ManagementPage: "/admin-next/plugins/webhooks?tab=secrets&sec_ref=" + wh.SecretRef + "&sec_plugin_code=" + wh.PluginCode,
	}
	if p, ok := s.PluginByCode(wh.PluginCode); ok {
		rel.PluginName = p.Name
	}
	return []domain.SecretUsageRelationship{rel}
}

func (s *Service) callbackTokenUsages(rec domain.SecretRefRecord) []domain.SecretUsageRelationship {
	cb, ok := s.repo.PluginCallbackTokenByRef(rec.Ref)
	if !ok || cb.ID == 0 {
		return []domain.SecretUsageRelationship{{
			Type:              secretCenterNamespaceCallback,
			Label:             "Callback Token",
			SecretRef:         rec.Ref,
			Unresolved:        true,
			UnresolvedMessage: "未找到对应 Callback Token 记录。",
		}}
	}
	rel := domain.SecretUsageRelationship{
		Type:           secretCenterNamespaceCallback,
		Label:          "Callback Token",
		UsageType:      "callback_token",
		SourceType:     secretCenterNamespaceCallback,
		SourceID:       cb.TokenRef,
		SourceCode:     cb.PluginCode,
		PluginCode:     cb.PluginCode,
		Status:         cb.Status,
		SecretRef:      cb.TokenRef,
		ConfigEntry:    "plugin_callback_tokens.token_ref",
		ManagementPage: "/admin-next/plugins/webhooks?tab=callback_tokens&cbtk_plugin_code=" + cb.PluginCode,
	}
	if p, ok := s.PluginByCode(cb.PluginCode); ok {
		rel.PluginName = p.Name
	}
	return []domain.SecretUsageRelationship{rel}
}

func (s *Service) pluginConfigSecretUsages(rec domain.SecretRefRecord) []domain.SecretUsageRelationship {
	pluginCode := externalServicePluginCodeFromSecret(rec)
	rel := domain.SecretUsageRelationship{
		Type:           "plugin_config_sensitive_field",
		Label:          "插件配置敏感字段",
		UsageType:      "plugin_config_sensitive_field",
		SourceType:     secretCenterNamespacePluginConfig,
		SourceID:       pluginCode,
		SourceCode:     pluginCode,
		PluginCode:     pluginCode,
		Status:         rec.Status,
		SecretRef:      rec.Ref,
		ConfigEntry:    "plugin_configs.config_json sensitive field",
		ManagementPage: "/admin-next/plugins/overview?tab=config&plugin_code=" + pluginCode,
	}
	if p, ok := s.PluginByCode(pluginCode); ok {
		rel.PluginName = p.Name
	}
	if pluginCode == "" {
		rel.Unresolved = true
		rel.UnresolvedMessage = "暂无法从 secret_ref 解析插件编码，请查看审计或技术详情。"
	}
	return []domain.SecretUsageRelationship{rel}
}

func (s *Service) SecretImpactPreview(ref, action string) (domain.SecretImpactPreview, error) {
	ref = strings.TrimSpace(ref)
	action = strings.TrimSpace(action)
	rec, err := s.GetSecretMetadata(ref)
	if err != nil {
		return domain.SecretImpactPreview{}, err
	}
	usages := s.SecretUsages(ref)
	preview := domain.SecretImpactPreview{
		Ref:              ref,
		Action:           action,
		Status:           rec.Status,
		Allowed:          true,
		AffectedBusiness: usages,
		Notes: []string{
			"预览基于当前存储配置和运行记录生成；执行时会重新检查状态。",
			"不会读取或返回敏感明文。",
		},
	}
	pluginSeen := map[string]bool{}
	for _, u := range usages {
		if u.PluginCode != "" {
			pluginSeen[u.PluginCode] = true
		}
		switch u.Type {
		case secretCenterNamespaceExternalService:
			if !u.Unresolved {
				preview.AffectedExternalServices++
				preview.PossibleFailedHealthChecks++
				preview.PossibleFailedDeliveries++
			}
		case secretCenterNamespaceWebhook:
			if !u.Unresolved {
				preview.AffectedWebhooks++
				preview.PossibleFailedDeliveries++
			}
		case secretCenterNamespaceCallback:
			if !u.Unresolved {
				preview.AffectedCallbacks++
			}
		}
	}
	preview.AffectedPlugins = len(pluginSeen)
	preview.UsageCountTotal = rec.UsageCount
	preview.UsageCountLast24h, preview.UsageCountLast7d = s.secretRecentUsageCounts(rec, usages)
	if s.isTestSecretRecord(rec) {
		preview.Allowed = false
		preview.Message = "测试/fixture/seed Secret 默认不允许在 SecretCenter 执行危险操作。"
		return preview, nil
	}
	switch action {
	case "disable":
		preview.Warning = "禁用后运行时解析该 Secret 会失败，相关健康检查、投递或回调可能失败。"
		if rec.Status != "" && rec.Status != domain.SecretRefStatusActive {
			preview.Allowed = false
			preview.Message = "只有 active Secret 可以禁用。"
		}
	case "revoke":
		preview.RequiresStrongConfirmation = true
		preview.ConfirmationText = ref
		preview.Warning = "吊销不是直接可恢复操作；相关来源需要通过轮换流程重新写入新凭据。"
		if rec.Status == domain.SecretRefStatusRevoked {
			preview.Allowed = false
			preview.Message = "revoked Secret 不能重复吊销。"
		}
		if rec.Status != "" && rec.Status != domain.SecretRefStatusActive && rec.Status != domain.SecretRefStatusDisabled {
			preview.Allowed = false
			preview.Message = "只有 active 或 disabled Secret 可以吊销。"
		}
	default:
		preview.Allowed = false
		preview.Message = "不支持的 Secret 操作。"
	}
	return preview, nil
}

func (s *Service) secretRecentUsageCounts(rec domain.SecretRefRecord, usages []domain.SecretUsageRelationship) (int, int) {
	if s.secretSourceType(rec) != secretCenterNamespaceExternalService {
		return 0, 0
	}
	pluginCode := ""
	for _, u := range usages {
		if u.Type == secretCenterNamespaceExternalService {
			pluginCode = u.PluginCode
			break
		}
	}
	if pluginCode == "" {
		return 0, 0
	}
	start24h := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	start7d := time.Now().Add(-7 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	_, count24h, _ := s.repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: pluginCode, ServiceType: externalServiceTypeExternal, StartTime: start24h, Page: 1, PageSize: 1})
	_, count7d, _ := s.repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: pluginCode, ServiceType: externalServiceTypeExternal, StartTime: start7d, Page: 1, PageSize: 1})
	return count24h, count7d
}

func (s *Service) secretSourceType(rec domain.SecretRefRecord) string {
	if s.isTestSecretRecord(rec) {
		return "test"
	}
	ns := strings.ToLower(strings.TrimSpace(rec.Namespace))
	if ns == "" && domain.IsSecretRef(rec.Ref) {
		ns, _, _ = domain.ParseSecretRef(rec.Ref)
		ns = strings.ToLower(strings.TrimSpace(ns))
	}
	switch ns {
	case secretCenterNamespaceExternalService, secretCenterNamespaceWebhook, secretCenterNamespaceCallback, secretCenterNamespacePluginConfig:
		return ns
	}
	if strings.HasPrefix(strings.TrimSpace(rec.Ref), webhookSecretRefPrefix) {
		return secretCenterNamespaceWebhook
	}
	if strings.HasPrefix(strings.TrimSpace(rec.Ref), callbackTokenRefPrefix) {
		return secretCenterNamespaceCallback
	}
	return "other"
}

func externalServicePluginCodeFromSecret(rec domain.SecretRefRecord) string {
	name := strings.Trim(strings.TrimSpace(rec.Name), "/")
	if name == "" && domain.IsSecretRef(rec.Ref) {
		_, parsedName, _ := domain.ParseSecretRef(rec.Ref)
		name = strings.Trim(parsedName, "/")
	}
	return strings.Split(name, "/")[0]
}

func (s *Service) isTestSecretRecord(rec domain.SecretRefRecord) bool {
	values := []string{rec.Ref, rec.Namespace, rec.Name, rec.AssociatedWith}
	prefixes := []string{"s15smoke", "e2e", "fixture", "test", "demo", "seed"}
	for _, raw := range values {
		value := strings.ToLower(strings.TrimSpace(raw))
		for _, prefix := range prefixes {
			if value == prefix || strings.HasPrefix(value, prefix+"_") || strings.HasPrefix(value, prefix+"-") || strings.Contains(value, "/"+prefix) {
				return true
			}
		}
	}
	return false
}

func (s *Service) disableSourceSecret(operator SecretOperator, ref string) (domain.SecretRefRecord, bool, error) {
	if wh, ok := s.repo.PluginWebhookSecretByRef(ref); ok && wh.ID > 0 {
		out, err := s.DisablePluginWebhookSecret(WebhookSecretOperator{ID: operator.ID, Name: operator.Name}, wh.ID)
		if err != nil {
			return domain.SecretRefRecord{}, true, err
		}
		return s.virtualWebhookSecretRef(out), true, nil
	}
	if cb, ok := s.repo.PluginCallbackTokenByRef(ref); ok && cb.ID > 0 {
		out, err := s.DisablePluginCallbackToken(CallbackTokenOperator{Type: operator.Type, ID: operator.ID, Name: operator.Name}, cb.ID)
		if err != nil {
			return domain.SecretRefRecord{}, true, err
		}
		return s.virtualCallbackTokenRef(out), true, nil
	}
	return domain.SecretRefRecord{}, false, nil
}

func (s *Service) revokeSourceSecret(operator SecretOperator, ref string) (domain.SecretRefRecord, bool, error) {
	if wh, ok := s.repo.PluginWebhookSecretByRef(ref); ok && wh.ID > 0 {
		out, err := s.RevokePluginWebhookSecret(WebhookSecretOperator{ID: operator.ID, Name: operator.Name}, wh.ID)
		if err != nil {
			return domain.SecretRefRecord{}, true, err
		}
		return s.virtualWebhookSecretRef(out), true, nil
	}
	if cb, ok := s.repo.PluginCallbackTokenByRef(ref); ok && cb.ID > 0 {
		out, err := s.RevokePluginCallbackToken(CallbackTokenOperator{Type: operator.Type, ID: operator.ID, Name: operator.Name}, cb.ID, "revoked from SecretCenter")
		if err != nil {
			return domain.SecretRefRecord{}, true, err
		}
		return s.virtualCallbackTokenRef(out), true, nil
	}
	return domain.SecretRefRecord{}, false, nil
}

// ResolveSecretInternal resolves secret plaintext for internal runtime use only.
//
// This must never be exposed via admin APIs.
func (s *Service) ResolveSecretInternal(ref string) (string, domain.SecretRefRecord, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid", "ref 必填").WithStatus(http.StatusBadRequest)
	}
	if _, _, err := domain.ParseSecretRef(ref); err != nil {
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_center_invalid_ref", "secret_ref 不合法").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	rec, ok := s.repo.SecretRefByRef(ref)
	if !ok || rec.ID == 0 {
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_ref_not_found", "secret_ref 不存在").WithStatus(http.StatusNotFound).WithDetail("ref", ref)
	}
	switch strings.TrimSpace(rec.Status) {
	case "", domain.SecretRefStatusActive:
		// ok
	case domain.SecretRefStatusDisabled:
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_ref_disabled", "secret_ref 已停用").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	case domain.SecretRefStatusRevoked:
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_ref_revoked", "secret_ref 已吊销").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	default:
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_ref_invalid_status", "secret_ref 状态不合法").WithStatus(http.StatusBadRequest).WithDetail("ref", ref).WithDetail("status", rec.Status)
	}
	if strings.TrimSpace(rec.EncryptedValue) == "" {
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_ref_value_missing", "secret_ref 缺少密文值").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	kr, err := s.secretCenterKeyringOrError(false)
	if err != nil {
		return "", domain.SecretRefRecord{}, err
	}
	plain, _, derr := pluginregistry.DecryptStringWithKeyring(kr, rec.EncryptedValue)
	if derr != nil || strings.TrimSpace(plain) == "" {
		return "", domain.SecretRefRecord{}, domain.NewPluginError("secret_ref_decrypt_failed", "secret_ref 解密失败").WithStatus(http.StatusBadRequest).WithDetail("ref", ref)
	}
	// Best-effort usage tracking; never fail the resolve solely due to tracking.
	_ = s.repo.TouchSecretRef(ref, Now())
	// Never return ciphertext in metadata.
	rec.EncryptedValue = ""
	return plain, rec, nil
}

func (s *Service) SecretCenterStatus() (domain.SecretCenterStatusResponse, error) {
	keyStatus, _ := s.PluginConfigKeyStatus()
	status := keyStatus.Status
	if status == "" {
		status = "warning"
	}
	counts, err := s.repo.SecretRefNamespaceCounts()
	if err != nil {
		return domain.SecretCenterStatusResponse{}, err
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	notes := []string{
		"SecretCenter 的 root key 只从环境变量读取；后台不会保存或生成 root key，修改后需要重启生效。",
		"当前 SecretCenter 只落地最小 secret_ref 引用层：明文仅在 create/update 请求中出现一次，列表/审计/执行记录不会回显明文。",
	}
	for i := range notes {
		notes[i] = strings.TrimSpace(notes[i])
	}
	// Stable output ordering is handled by UI; still keep counts map deterministic if needed.
	if counts == nil {
		counts = map[string]int{}
	}
	// Ensure external_service namespace always present in response for UI.
	if _, ok := counts[secretCenterNamespaceExternalService]; !ok {
		counts[secretCenterNamespaceExternalService] = 0
	}
	// Remove empty key to keep response clean.
	delete(counts, "")
	return domain.SecretCenterStatusResponse{
		Status:          status,
		SecretRefCount:  total,
		NamespaceCounts: counts,
		Notes:           notes,
	}, nil
}

func (s *Service) ExternalServiceTokenRef(pluginCode string) (string, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" {
		return "", errors.New("plugin_code 不能为空")
	}
	return domain.BuildSecretRef(secretCenterNamespaceExternalService, pluginCode+"/token")
}
