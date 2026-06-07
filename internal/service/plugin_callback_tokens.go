package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	callbackTokenRefPrefix = "cbtk_"
	callbackTokenPrefix    = "cbsk_"
)

var allowedCallbackScopes = map[string]bool{
	"config.read": true,
	"audit.write": true,
	// planning only:
	// "notification.send":     true,
	// "webhook.delivery.read": true,
}

type CreatePluginCallbackTokenRequest struct {
	PluginCode       string   `json:"plugin_code"`
	Name             string   `json:"name"`
	Scopes           []string `json:"scopes"`
	CommunityScope   []int64  `json:"community_scope"`
	ExpiresAtRFC3339 string   `json:"expires_at,omitempty"`
}

type CreatePluginCallbackTokenResponse struct {
	TokenRecord domain.PluginCallbackToken `json:"token_record"`
	TokenRef    string                     `json:"token_ref"`
	Token       string                     `json:"token"` // plaintext, returned once
	Scopes      []string                   `json:"scopes"`
	Status      string                     `json:"status"`
}

type RotatePluginCallbackTokenResponse struct {
	OldTokenRef string                     `json:"old_token_ref"`
	NewTokenRef string                     `json:"new_token_ref"`
	TokenRecord domain.PluginCallbackToken `json:"token_record"`
	Token       string                     `json:"token"` // plaintext, returned once
}

type CallbackTokenOperator struct {
	Type string
	ID   int64
	Name string
}

func (s *Service) ListPluginCallbackTokens(filter domain.PluginCallbackTokenFilter) (domain.PluginCallbackTokenListResponse, error) {
	items, total, err := s.repo.PluginCallbackTokens(filter)
	if err != nil {
		return domain.PluginCallbackTokenListResponse{}, err
	}
	f := filter.Normalize()
	// Never expose token_hash via admin list; it is not plaintext but still sensitive.
	for i := range items {
		items[i].TokenHash = ""
	}
	return domain.PluginCallbackTokenListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     f.Page,
			PageSize: f.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginCallbackToken(id int64) (domain.PluginCallbackToken, error) {
	it, ok := s.repo.PluginCallbackTokenByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginCallbackToken{}, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404)
	}
	it.TokenHash = ""
	return it, nil
}

func (s *Service) CreatePluginCallbackToken(operator CallbackTokenOperator, req CreatePluginCallbackTokenRequest) (CreatePluginCallbackTokenResponse, error) {
	pluginCode := strings.TrimSpace(req.PluginCode)
	if pluginCode == "" {
		return CreatePluginCallbackTokenResponse{}, domain.NewPluginError("callback_token_invalid", "plugin_code 必填").WithStatus(400)
	}
	if _, ok := s.repo.PluginByCode(pluginCode); !ok {
		return CreatePluginCallbackTokenResponse{}, domain.NewPluginError("plugin_not_found", "插件不存在").WithStatus(404).WithDetail("plugin_code", pluginCode)
	}
	scopes, err := normalizeCallbackScopes(req.Scopes)
	if err != nil {
		return CreatePluginCallbackTokenResponse{}, err
	}
	communityScope := normalizeCommunityScope(req.CommunityScope)
	if len(communityScope) == 0 {
		// do not default to all communities
		return CreatePluginCallbackTokenResponse{}, domain.NewPluginError("callback_token_invalid", "community_scope 不能为空").WithStatus(400).
			WithSuggestion("请至少指定一个 community_id（例如 [1]），避免 token 默认全社区可用。")
	}

	var expiresAt string
	if strings.TrimSpace(req.ExpiresAtRFC3339) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAtRFC3339))
		if err != nil {
			return CreatePluginCallbackTokenResponse{}, domain.NewPluginError("callback_token_invalid", "expires_at 格式不合法").WithStatus(400)
		}
		expiresAt = t.Local().Format("2006-01-02 15:04:05")
	}

	tokenRef := callbackTokenRefPrefix + randomHex(18)
	plaintext := callbackTokenPrefix + randomHex(24)
	hash := sha256StringHex(plaintext)

	now := Now()
	record := domain.PluginCallbackToken{
		PluginCode:         pluginCode,
		TokenRef:           tokenRef,
		TokenHash:          hash,
		Name:               strings.TrimSpace(req.Name),
		Status:             domain.PluginCallbackTokenStatusActive,
		ScopesJSON:         mustJSON(scopes),
		CommunityScopeJSON: mustJSON(communityScope),
		ExpiresAt:          expiresAt,
		CreatedBy:          operator.ID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	saved, err := s.repo.AppendPluginCallbackToken(record)
	if err != nil {
		return CreatePluginCallbackTokenResponse{}, err
	}

	// Audit without plaintext/hash.
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.callback_token.created",
		Target:    "callback-tokens#" + saved.TokenRef,
		Metadata: mustJSON(map[string]any{
			"plugin_code":     pluginCode,
			"token_ref":       saved.TokenRef,
			"scopes":          scopes,
			"community_scope": communityScope,
			"expires_at":      saved.ExpiresAt,
			"status":          saved.Status,
		}),
		CreatedAt: Now(),
	})

	saved.TokenHash = ""
	return CreatePluginCallbackTokenResponse{
		TokenRecord: saved,
		TokenRef:    saved.TokenRef,
		Token:       plaintext,
		Scopes:      scopes,
		Status:      saved.Status,
	}, nil
}

func (s *Service) DisablePluginCallbackToken(operator CallbackTokenOperator, id int64) (domain.PluginCallbackToken, error) {
	it, ok := s.repo.PluginCallbackTokenByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginCallbackToken{}, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404)
	}
	if it.Status == domain.PluginCallbackTokenStatusRevoked {
		return domain.PluginCallbackToken{}, domain.NewPluginError("callback_token_revoked", "revoked token 不能禁用").WithStatus(400)
	}
	it.Status = domain.PluginCallbackTokenStatusDisabled
	it.UpdatedAt = Now()
	out, err := s.repo.SavePluginCallbackToken(it)
	if err != nil {
		return domain.PluginCallbackToken{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.callback_token.disabled",
		Target:    "callback-tokens#" + out.TokenRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": out.PluginCode, "token_ref": out.TokenRef, "status": out.Status}),
		CreatedAt: Now(),
	})
	out.TokenHash = ""
	return out, nil
}

func (s *Service) EnablePluginCallbackToken(operator CallbackTokenOperator, id int64) (domain.PluginCallbackToken, error) {
	it, ok := s.repo.PluginCallbackTokenByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginCallbackToken{}, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404)
	}
	if it.Status != domain.PluginCallbackTokenStatusDisabled {
		return domain.PluginCallbackToken{}, domain.NewPluginError("callback_token_enable_invalid", "仅 disabled token 可恢复").WithStatus(400)
	}
	it.Status = domain.PluginCallbackTokenStatusActive
	it.UpdatedAt = Now()
	out, err := s.repo.SavePluginCallbackToken(it)
	if err != nil {
		return domain.PluginCallbackToken{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.callback_token.enabled",
		Target:    "callback-tokens#" + out.TokenRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": out.PluginCode, "token_ref": out.TokenRef, "status": out.Status}),
		CreatedAt: Now(),
	})
	out.TokenHash = ""
	return out, nil
}

func (s *Service) RevokePluginCallbackToken(operator CallbackTokenOperator, id int64, reason string) (domain.PluginCallbackToken, error) {
	it, ok := s.repo.PluginCallbackTokenByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginCallbackToken{}, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404)
	}
	it.Status = domain.PluginCallbackTokenStatusRevoked
	it.RevokedAt = Now()
	it.RevokedReason = truncateString(reason, 500)
	it.UpdatedAt = it.RevokedAt
	out, err := s.repo.SavePluginCallbackToken(it)
	if err != nil {
		return domain.PluginCallbackToken{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.callback_token.revoked",
		Target:    "callback-tokens#" + out.TokenRef,
		Metadata:  mustJSON(map[string]any{"plugin_code": out.PluginCode, "token_ref": out.TokenRef, "status": out.Status, "reason": out.RevokedReason}),
		CreatedAt: Now(),
	})
	out.TokenHash = ""
	return out, nil
}

func (s *Service) RotatePluginCallbackToken(operator CallbackTokenOperator, id int64) (RotatePluginCallbackTokenResponse, error) {
	cur, ok := s.repo.PluginCallbackTokenByID(id)
	if !ok || cur.ID == 0 {
		return RotatePluginCallbackTokenResponse{}, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404)
	}
	if cur.Status == domain.PluginCallbackTokenStatusRevoked {
		return RotatePluginCallbackTokenResponse{}, domain.NewPluginError("callback_token_revoked", "revoked token 不能轮换").WithStatus(400)
	}

	now := Now()
	newRef := callbackTokenRefPrefix + randomHex(18)
	plaintext := callbackTokenPrefix + randomHex(24)
	hash := sha256StringHex(plaintext)

	newRec := domain.PluginCallbackToken{
		PluginCode:           cur.PluginCode,
		PluginInstallationID: cur.PluginInstallationID,
		PublisherID:          cur.PublisherID,
		TokenRef:             newRef,
		TokenHash:            hash,
		Name:                 firstNonEmpty(strings.TrimSpace(cur.Name), "rotated"),
		Status:               domain.PluginCallbackTokenStatusActive,
		ScopesJSON:           cur.ScopesJSON,
		CommunityScopeJSON:   cur.CommunityScopeJSON,
		ExpiresAt:            cur.ExpiresAt,
		CreatedBy:            operator.ID,
		CreatedAt:            now,
		RotatedAt:            now,
		UpdatedAt:            now,
	}
	savedNew, err := s.repo.AppendPluginCallbackToken(newRec)
	if err != nil {
		return RotatePluginCallbackTokenResponse{}, err
	}

	// Revoke old token immediately by default.
	cur.Status = domain.PluginCallbackTokenStatusRevoked
	cur.RevokedAt = now
	cur.RevokedReason = "rotated"
	cur.RotatedAt = now
	cur.UpdatedAt = now
	_, _ = s.repo.SavePluginCallbackToken(cur)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.callback_token.rotated",
		Target:    "callback-tokens#" + savedNew.TokenRef,
		Metadata: mustJSON(map[string]any{
			"plugin_code":   cur.PluginCode,
			"old_token_ref": cur.TokenRef,
			"new_token_ref": savedNew.TokenRef,
		}),
		CreatedAt: Now(),
	})

	savedNew.TokenHash = ""
	return RotatePluginCallbackTokenResponse{
		OldTokenRef: cur.TokenRef,
		NewTokenRef: savedNew.TokenRef,
		TokenRecord: savedNew,
		Token:       plaintext,
	}, nil
}

func (s *Service) TouchPluginCallbackTokenUsage(id int64, ip string) error {
	ip = strings.TrimSpace(ip)
	it, ok := s.repo.PluginCallbackTokenByID(id)
	if !ok || it.ID == 0 {
		return domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404)
	}
	// token_hash is required for Save. Keep it as-is and do NOT expose to caller.
	it.LastUsedAt = Now()
	it.LastUsedIP = parseIPFromAddr(ip)
	it.UpdatedAt = Now()
	_, err := s.repo.SavePluginCallbackToken(it)
	return err
}

func normalizeCallbackScopes(in []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !allowedCallbackScopes[s] {
			return nil, domain.NewPluginError("callback_scope_invalid", "scope 不合法").WithStatus(400).WithDetail("scope", s)
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil, domain.NewPluginError("callback_scope_empty", "scopes 不能为空").WithStatus(400)
	}
	return out, nil
}

func normalizeCommunityScope(in []int64) []int64 {
	seen := map[int64]bool{}
	out := []int64{}
	for _, id := range in {
		if id <= 0 {
			continue
		}
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func randomHex(nbytes int) string {
	if nbytes <= 0 {
		nbytes = 16
	}
	buf := make([]byte, nbytes)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func hashBearerToken(token string) string {
	token = strings.TrimSpace(token)
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func parseIPFromAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		return host
	}
	return addr
}

func (s *Service) ValidateCallbackScope(scope string) bool {
	return allowedCallbackScopes[strings.TrimSpace(scope)]
}

func (s *Service) HashCallbackBearerToken(token string) string {
	return hashBearerToken(token)
}

func (s *Service) ShouldAllowCallbackForPlugin(pluginCode string) error {
	p, ok := s.repo.PluginByCode(pluginCode)
	if !ok || p.Code == "" {
		return domain.NewPluginError("plugin_not_found", "插件不存在").WithStatus(404)
	}
	// Global status must be enabled to allow callback channel.
	if strings.TrimSpace(p.Status) != pluginregistry.StatusEnabled {
		return domain.NewPluginError("plugin_disabled", "插件未启用").WithStatus(403).WithDetail("plugin_code", pluginCode)
	}
	return nil
}
