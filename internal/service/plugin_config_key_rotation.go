package service

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

func (s *Service) PluginConfigKeyStatus() (domain.PluginConfigKeyStatusResponse, error) {
	examples := []string{
		`DEVHUB_PLUGIN_CONFIG_KEYS='[{"id":"local-v1","key":"base64-xxx","primary":true}]'`,
		"DEVHUB_PLUGIN_CONFIG_KEY_ID=local-v1\nDEVHUB_PLUGIN_CONFIG_KEY=base64-xxx",
	}
	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return domain.PluginConfigKeyStatusResponse{
			Status:   "blocked",
			Warnings: []string{err.Error()},
			Source:   "env",
			// Changing startup keys always requires a restart for the running process.
			RestartRequired: true,
			EnvExamples:     examples,
		}, nil
	}
	if !ok || kr == nil {
		return domain.PluginConfigKeyStatusResponse{
			Status:          "blocked",
			Warnings:        []string{"缺少启动加密密钥：无法创建/解密运行时敏感配置（Webhook Secret、external_service token 等）。该密钥只能在启动环境变量中配置，DevHub 后台不会保存或生成。"},
			Source:          "env",
			RestartRequired: true,
			EnvExamples:     examples,
		}, nil
	}
	return domain.PluginConfigKeyStatusResponse{
		CurrentKeyID:      kr.CurrentKeyID,
		LoadedKeyIDs:      kr.AllKeyIDs(),
		LegacyV1Supported: kr.LegacyV1Supported,
		KeyCount:          len(kr.Keys),
		Status:            "ok",
		Warnings:          []string{},
		Source:            "env",
		RestartRequired:   true,
		EnvExamples:       examples,
	}, nil
}

func (s *Service) PluginConfigKeyRotationDryRun(req domain.PluginConfigKeyRotationDryRunRequest) (domain.PluginConfigKeyRotationDryRunResponse, error) {
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "all"
	}
	if req.IncludeConfigVersions {
		return domain.PluginConfigKeyRotationDryRunResponse{
			Status:   "blocked",
			Warnings: []string{},
			Errors: []domain.APIError{
				*domain.NewPluginError("plugin_config_rotation_history_unsupported", "本版本暂不支持扫描并轮换配置历史版本").WithStatus(400).
					WithSuggestion("请先使用 include_config_versions=false 轮换当前配置；历史轮换将于后续版本补齐。"),
			},
		}, nil
	}

	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return domain.PluginConfigKeyRotationDryRunResponse{
			Status: "blocked",
			Errors: []domain.APIError{*domain.NewPluginError("plugin_config_key_invalid", "插件配置加密密钥配置不合法").WithStatus(500).WithSuggestion("请检查 DEVHUB_PLUGIN_CONFIG_KEYS 或 split 环境变量格式。")},
		}, nil
	}
	if !ok || kr == nil {
		return domain.PluginConfigKeyRotationDryRunResponse{
			Status: "blocked",
			Errors: []domain.APIError{*domain.NewPluginError("plugin_config_key_current_missing", "缺少 current 插件配置加密密钥").WithStatus(500).WithSuggestion("请配置 DEVHUB_PLUGIN_CONFIG_KEYS.current 和 keys，或 DEVHUB_PLUGIN_CONFIG_KEY_ID/DEVHUB_PLUGIN_CONFIG_KEY。")},
		}, nil
	}

	items := []domain.PluginConfigKeyRotationDryRunItem{}
	summary := domain.PluginConfigKeyRotationDryRunSummary{}

	switch scope {
	case "all":
		for _, p := range s.repo.Plugins() {
			if strings.TrimSpace(p.Code) == "" {
				continue
			}
			scanKeyRotationForConfig(kr, p.Code, domain.PluginConfigScopeGlobal, 0, p.ConfigSchema, p.ConfigJSON, &items, &summary)
		}
		for _, comm := range s.repo.Communities() {
			plist, err := s.repo.CommunityPlugins(comm.ID)
			if err != nil {
				continue
			}
			for _, p := range plist {
				if strings.TrimSpace(p.Code) == "" {
					continue
				}
				scanKeyRotationForConfig(kr, p.Code, domain.PluginConfigScopeCommunity, comm.ID, p.ConfigSchema, p.ConfigJSON, &items, &summary)
			}
		}
	case "plugin":
		code := strings.TrimSpace(req.PluginCode)
		if code == "" {
			return domain.PluginConfigKeyRotationDryRunResponse{
				Status: "blocked",
				Errors: []domain.APIError{*domain.NewPluginError("plugin_config_rotation_dry_run_blocked", "scope=plugin 时 plugin_code 必填").WithStatus(400)},
			}, nil
		}
		p, ok := s.repo.PluginByCode(code)
		if ok {
			scanKeyRotationForConfig(kr, p.Code, domain.PluginConfigScopeGlobal, 0, p.ConfigSchema, p.ConfigJSON, &items, &summary)
		}
		for _, comm := range s.repo.Communities() {
			plist, err := s.repo.CommunityPlugins(comm.ID)
			if err != nil {
				continue
			}
			for _, it := range plist {
				if it.Code == code {
					scanKeyRotationForConfig(kr, it.Code, domain.PluginConfigScopeCommunity, comm.ID, it.ConfigSchema, it.ConfigJSON, &items, &summary)
				}
			}
		}
	case "community":
		cid := req.CommunityID
		if cid <= 0 {
			return domain.PluginConfigKeyRotationDryRunResponse{
				Status: "blocked",
				Errors: []domain.APIError{*domain.NewPluginError("plugin_config_rotation_dry_run_blocked", "scope=community 时 community_id 必填").WithStatus(400)},
			}, nil
		}
		plist, err := s.repo.CommunityPlugins(cid)
		if err != nil {
			return domain.PluginConfigKeyRotationDryRunResponse{
				Status: "blocked",
				Errors: []domain.APIError{*domain.NewPluginError("plugin_config_rotation_dry_run_blocked", "获取子站插件配置失败").WithStatus(500)},
			}, nil
		}
		for _, p := range plist {
			scanKeyRotationForConfig(kr, p.Code, domain.PluginConfigScopeCommunity, cid, p.ConfigSchema, p.ConfigJSON, &items, &summary)
		}
	default:
		return domain.PluginConfigKeyRotationDryRunResponse{
			Status: "blocked",
			Errors: []domain.APIError{*domain.NewPluginError("plugin_config_rotation_dry_run_blocked", "scope 不合法").WithStatus(400).WithDetail("scope", scope)},
		}, nil
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].PluginCode != items[j].PluginCode {
			return items[i].PluginCode < items[j].PluginCode
		}
		if items[i].Scope != items[j].Scope {
			return items[i].Scope < items[j].Scope
		}
		if items[i].CommunityID != items[j].CommunityID {
			return items[i].CommunityID < items[j].CommunityID
		}
		return items[i].FieldPath < items[j].FieldPath
	})

	status := "ok"
	warnings := []string{}
	errors := []domain.APIError{}
	if summary.DecryptFailed > 0 || summary.MissingKey > 0 {
		status = "blocked"
		errors = append(errors, *domain.NewPluginError("plugin_config_rotation_dry_run_blocked", "密钥轮换预检失败：存在无法解密的敏感配置").WithStatus(400).
			WithDetail("decrypt_failed", summary.DecryptFailed).
			WithDetail("missing_key", summary.MissingKey).
			WithSuggestion("请补齐 old keys 或修复密文后重试。"))
	} else if summary.NeedsReencrypt > 0 || summary.LegacyV1 > 0 {
		status = "warning"
		warnings = append(warnings, "存在需要重新加密的敏感配置，建议执行 re-encrypt。")
	}

	return domain.PluginConfigKeyRotationDryRunResponse{
		Status:       status,
		CurrentKeyID: kr.CurrentKeyID,
		Summary:      summary,
		Items:        items,
		Warnings:     warnings,
		Errors:       errors,
	}, nil
}

func (s *Service) PluginConfigKeyRotationReencrypt(req domain.PluginConfigKeyRotationReencryptRequest, operator PluginConfigVersionOperator) (domain.PluginConfigKeyRotationReencryptResponse, error) {
	// Always dry-run first (do not trust caller).
	dry, err := s.PluginConfigKeyRotationDryRun(domain.PluginConfigKeyRotationDryRunRequest{
		Scope:                 req.Scope,
		PluginCode:            req.PluginCode,
		CommunityID:           req.CommunityID,
		IncludeConfigVersions: req.IncludeConfigVersions,
	})
	if err != nil {
		return domain.PluginConfigKeyRotationReencryptResponse{}, err
	}
	if dry.Status == "blocked" {
		return domain.PluginConfigKeyRotationReencryptResponse{
				Status:  "blocked",
				Summary: dry.Summary,
				Errors:  dry.Errors,
			}, domain.NewPluginError("plugin_config_rotation_dry_run_blocked", "密钥轮换预检 blocked，禁止执行 re-encrypt").WithStatus(400).
				WithSuggestion("请先修复 dry-run 阻断原因后重试。")
	}

	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil || !ok || kr == nil {
		return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_config_key_current_missing", "缺少 current 插件配置加密密钥").WithStatus(500)
	}
	if strings.TrimSpace(req.ConfirmCurrentKeyID) == "" || strings.TrimSpace(req.ConfirmCurrentKeyID) != strings.TrimSpace(kr.CurrentKeyID) {
		return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_config_rotation_confirm_key_mismatch", "confirm_current_key_id 与当前密钥不匹配").WithStatus(400).
			WithDetail("current_key_id", kr.CurrentKeyID).
			WithSuggestion("请刷新页面后重新确认当前 key_id。")
	}

	updated := 0
	seenConfig := map[string]bool{}
	for _, it := range dry.Items {
		if it.Status != "needs_reencrypt" && it.Status != "cipher_invalid" {
			continue
		}
		key := it.PluginCode + "|" + it.Scope + "|" + int64ToString(it.CommunityID)
		seenConfig[key] = true
	}

	for key := range seenConfig {
		parts := strings.Split(key, "|")
		if len(parts) != 3 {
			continue
		}
		code := parts[0]
		scope := parts[1]
		cid := parseInt64(parts[2])

		if scope == domain.PluginConfigScopeGlobal {
			before, _ := s.repo.PluginByCode(code)
			afterJSON, changed, err := reencryptSensitiveConfigJSON(kr, before.ConfigSchema, before.ConfigJSON)
			if err != nil {
				return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_config_rotation_reencrypt_failed", "重新加密失败").WithStatus(500).WithDetail("plugin_code", code).
					WithSuggestion("请检查密钥配置与密文格式后重试。")
			}
			if changed {
				if _, err := s.repo.SetPluginConfig(code, afterJSON); err != nil {
					return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_config_rotation_reencrypt_failed", "写入插件配置失败").WithStatus(500).WithDetail("plugin_code", code)
				}
				if err := s.refreshPluginRegistry(pluginRegistryRefreshEvent{
					Trigger:    "after_config_change",
					PluginCode: code,
					ActorType:  "admin_user",
					ActorID:    operator.ID,
					ActorName:  firstNonEmpty(operator.Name, "system"),
				}); err != nil {
					return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_registry_reload_failed", "插件配置已轮换，但运行态刷新失败").WithStatus(500).WithDetail("plugin_code", code)
				}
				// Record a new config version for traceability (ciphertext changes).
				_, _, _ = s.RecordPluginConfigVersion(code, domain.PluginConfigScopeGlobal, 0, before.ConfigJSON, afterJSON, "key_rotation", operator)
				updated++
			}
			continue
		}
		if scope == domain.PluginConfigScopeCommunity && cid > 0 {
			items, _ := s.repo.CommunityPlugins(cid)
			var before domain.Plugin
			for _, p := range items {
				if p.Code == code {
					before = p
					break
				}
			}
			afterJSON, changed, err := reencryptSensitiveConfigJSON(kr, before.ConfigSchema, before.ConfigJSON)
			if err != nil {
				return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_config_rotation_reencrypt_failed", "重新加密失败").WithStatus(500).
					WithDetail("plugin_code", code).WithDetail("community_id", cid)
			}
			if changed {
				if _, err := s.repo.SetCommunityPluginConfig(cid, code, afterJSON); err != nil {
					return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_config_rotation_reencrypt_failed", "写入子站插件配置失败").WithStatus(500).
						WithDetail("plugin_code", code).WithDetail("community_id", cid)
				}
				if err := s.refreshPluginRegistry(pluginRegistryRefreshEvent{
					Trigger:     "after_config_change",
					PluginCode:  code,
					CommunityID: cid,
					ActorType:   "admin_user",
					ActorID:     operator.ID,
					ActorName:   firstNonEmpty(operator.Name, "system"),
				}); err != nil {
					return domain.PluginConfigKeyRotationReencryptResponse{}, domain.NewPluginError("plugin_registry_reload_failed", "子站插件配置已轮换，但运行态刷新失败").WithStatus(500).
						WithDetail("plugin_code", code).WithDetail("community_id", cid)
				}
				_, _, _ = s.RecordPluginConfigVersion(code, domain.PluginConfigScopeCommunity, cid, before.ConfigJSON, afterJSON, "key_rotation", operator)
				updated++
			}
		}
	}

	out := domain.PluginConfigKeyRotationReencryptResponse{
		Status:       "ok",
		CurrentKeyID: kr.CurrentKeyID,
		Summary:      dry.Summary,
		UpdatedCount: updated,
		Warnings:     []string{},
		Errors:       []domain.APIError{},
	}
	if updated > 0 {
		out.Status = "ok"
	}
	return out, nil
}

func scanKeyRotationForConfig(kr *pluginregistry.PluginConfigKeyring, pluginCode, scope string, communityID int64, schema any, configJSON string, out *[]domain.PluginConfigKeyRotationDryRunItem, summary *domain.PluginConfigKeyRotationDryRunSummary) {
	configJSON = strings.TrimSpace(configJSON)
	if configJSON == "" || !json.Valid([]byte(configJSON)) {
		return
	}
	var v any
	if err := json.Unmarshal([]byte(configJSON), &v); err != nil {
		return
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return
	}
	props := pluginregistrySchemaProps(schema)
	scanSensitiveFields(kr, pluginCode, scope, communityID, props, "", obj, out, summary)
}

func scanSensitiveFields(kr *pluginregistry.PluginConfigKeyring, pluginCode, scope string, communityID int64, schemaProps map[string]any, prefix string, obj map[string]any, out *[]domain.PluginConfigKeyRotationDryRunItem, summary *domain.PluginConfigKeyRotationDryRunSummary) {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		subSchema := schemaProps[k]
		val := obj[k]

		// Recurse for nested objects.
		if m, ok := val.(map[string]any); ok {
			subProps := pluginregistrySchemaProps(subSchema)
			scanSensitiveFields(kr, pluginCode, scope, communityID, subProps, path, m, out, summary)
			continue
		}

		if !pluginregistry.IsSensitiveField(subSchema, k, path) {
			continue
		}

		summary.TotalSensitiveValues++

		// Empty = no-op.
		if val == nil {
			continue
		}
		s, ok := val.(string)
		if !ok {
			*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
				PluginCode:    pluginCode,
				Scope:         scope,
				CommunityID:   communityID,
				FieldPath:     path,
				CipherVersion: "invalid",
				KeyID:         "",
				Status:        "cipher_invalid",
				Message:       "敏感字段值类型不合法，预期 string",
			})
			summary.DecryptFailed++
			continue
		}
		if strings.TrimSpace(s) == "" {
			// cleared
			summary.TotalSensitiveValues--
			continue
		}

		info := pluginregistry.DetectCiphertextVersion(s)
		switch info.Version {
		case "v2":
			if info.KeyID == "" {
				*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
					PluginCode:    pluginCode,
					Scope:         scope,
					CommunityID:   communityID,
					FieldPath:     path,
					CipherVersion: "v2",
					KeyID:         "",
					Status:        "missing_key",
					Message:       "enc:v2 密文缺少 key_id",
				})
				summary.MissingKey++
				continue
			}
			if strings.TrimSpace(info.KeyID) == strings.TrimSpace(kr.CurrentKeyID) {
				*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
					PluginCode:    pluginCode,
					Scope:         scope,
					CommunityID:   communityID,
					FieldPath:     path,
					CipherVersion: "v2",
					KeyID:         info.KeyID,
					Status:        "already_current",
					Message:       "已使用 current key 加密",
				})
				summary.AlreadyCurrent++
				continue
			}
			if _, ok := kr.ResolveKey(info.KeyID); !ok {
				*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
					PluginCode:    pluginCode,
					Scope:         scope,
					CommunityID:   communityID,
					FieldPath:     path,
					CipherVersion: "v2",
					KeyID:         info.KeyID,
					Status:        "missing_key",
					Message:       "缺少对应 key_id 的 old key，无法解密",
				})
				summary.MissingKey++
				continue
			}
			if _, _, err := pluginregistry.DecryptStringWithKeyring(kr, s); err != nil {
				*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
					PluginCode:    pluginCode,
					Scope:         scope,
					CommunityID:   communityID,
					FieldPath:     path,
					CipherVersion: "v2",
					KeyID:         info.KeyID,
					Status:        "decrypt_failed",
					Message:       "解密失败，可能密钥不匹配或密文损坏",
				})
				summary.DecryptFailed++
				continue
			}
			*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
				PluginCode:    pluginCode,
				Scope:         scope,
				CommunityID:   communityID,
				FieldPath:     path,
				CipherVersion: "v2",
				KeyID:         info.KeyID,
				Status:        "needs_reencrypt",
				Message:       "旧 key_id 密文可解密，建议重新加密为 current key",
			})
			summary.NeedsReencrypt++
		case "v1":
			if _, _, err := pluginregistry.DecryptStringWithKeyring(kr, s); err != nil {
				*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
					PluginCode:    pluginCode,
					Scope:         scope,
					CommunityID:   communityID,
					FieldPath:     path,
					CipherVersion: "v1",
					KeyID:         "",
					Status:        "decrypt_failed",
					Message:       "旧 v1 密文解密失败，可能缺少 legacy/old key",
				})
				summary.DecryptFailed++
				continue
			}
			*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
				PluginCode:    pluginCode,
				Scope:         scope,
				CommunityID:   communityID,
				FieldPath:     path,
				CipherVersion: "v1",
				KeyID:         "",
				Status:        "needs_reencrypt",
				Message:       "旧 v1 密文可解密，建议重新加密为 enc:v2 current key",
			})
			summary.LegacyV1++
			summary.NeedsReencrypt++
		default:
			*out = append(*out, domain.PluginConfigKeyRotationDryRunItem{
				PluginCode:    pluginCode,
				Scope:         scope,
				CommunityID:   communityID,
				FieldPath:     path,
				CipherVersion: "plain",
				KeyID:         "",
				Status:        "decrypt_failed",
				Message:       "发现敏感字段明文或未知格式，禁止轮换",
			})
			summary.DecryptFailed++
		}
	}
}

func reencryptSensitiveConfigJSON(kr *pluginregistry.PluginConfigKeyring, schema any, raw string) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, nil
	}
	if !json.Valid([]byte(raw)) {
		return raw, false, nil
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw, false, err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return raw, false, nil
	}
	props := pluginregistrySchemaProps(schema)
	changed := false
	out, err := reencryptSensitiveObject(kr, props, "", obj, &changed)
	if err != nil {
		return raw, false, err
	}
	buf, err := json.Marshal(out)
	if err != nil {
		return raw, false, err
	}
	return string(buf), changed, nil
}

func reencryptSensitiveObject(kr *pluginregistry.PluginConfigKeyring, schemaProps map[string]any, prefix string, obj map[string]any, changed *bool) (map[string]any, error) {
	out := map[string]any{}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		subSchema := schemaProps[k]
		val := obj[k]
		if m, ok := val.(map[string]any); ok {
			subProps := pluginregistrySchemaProps(subSchema)
			nested, err := reencryptSensitiveObject(kr, subProps, path, m, changed)
			if err != nil {
				return nil, err
			}
			out[k] = nested
			continue
		}
		if !pluginregistry.IsSensitiveField(subSchema, k, path) {
			out[k] = val
			continue
		}
		s, ok := val.(string)
		if !ok {
			out[k] = val
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" {
			out[k] = ""
			continue
		}
		info := pluginregistry.DetectCiphertextVersion(s)
		if info.Version == "v2" && strings.TrimSpace(info.KeyID) == strings.TrimSpace(kr.CurrentKeyID) {
			out[k] = s
			continue
		}
		if info.Version != "v1" && info.Version != "v2" {
			// do not try to encrypt plaintext here.
			out[k] = s
			continue
		}
		plain, _, err := pluginregistry.DecryptStringWithKeyring(kr, s)
		if err != nil {
			return nil, err
		}
		keyID, key := kr.CurrentKey()
		enc, err := pluginregistry.EncryptStringV2(keyID, key, plain)
		if err != nil {
			return nil, err
		}
		out[k] = enc
		*changed = true
	}
	return out, nil
}

func pluginregistrySchemaProps(schema any) map[string]any {
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

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

func parseInt64(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, _ := strconv.ParseInt(raw, 10, 64)
	return n
}
