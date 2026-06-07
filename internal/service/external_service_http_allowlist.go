package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

const (
	externalServiceAllowlistNamespace = "external_service"
	externalServiceAllowlistKey       = "http_allowlist"

	externalServiceAllowlistSourceDefault = "system_default"
	externalServiceAllowlistSourceEnv     = "env"
	externalServiceAllowlistSourceAdmin   = "admin"
)

type storedExternalServiceHTTPAllowlistOrigin struct {
	ID        string `json:"id"`
	Origin    string `json:"origin"`
	Usage     string `json:"usage,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func (s *Service) ExternalServiceHTTPAllowlist() (domain.PluginExternalServiceHTTPAllowlistResponse, error) {
	admin, err := s.externalServiceAdminHTTPAllowlist()
	if err != nil {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, err
	}
	return buildExternalServiceHTTPAllowlistResponse(admin), nil
}

func (s *Service) AddExternalServiceHTTPAllowlistOrigin(operator PluginExternalServiceOperator, req domain.PluginExternalServiceHTTPAllowlistUpdateRequest) (domain.PluginExternalServiceHTTPAllowlistResponse, error) {
	if !req.RiskConfirmed {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, domain.NewPluginError("external_service_http_allowlist_confirm_required", "请先确认该 HTTP origin 的联调风险").WithStatus(400)
	}
	origin, err := validateExternalServiceHTTPAllowlistOrigin(req.Origin, false)
	if err != nil {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, err
	}
	usage := strings.TrimSpace(req.Usage)
	if usage == "" {
		usage = "本地开发或受控内网联调"
	}
	if len([]rune(usage)) > 200 {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, domain.NewPluginError("external_service_http_allowlist_usage_too_long", "用途说明不能超过 200 字").WithStatus(400)
	}
	admin, err := s.externalServiceAdminHTTPAllowlist()
	if err != nil {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, err
	}
	now := Now()
	found := false
	for i := range admin {
		if admin[i].Origin == origin {
			admin[i].Usage = usage
			admin[i].UpdatedAt = now
			found = true
			break
		}
	}
	if !found {
		admin = append(admin, storedExternalServiceHTTPAllowlistOrigin{
			ID:        externalServiceHTTPAllowlistID(origin),
			Origin:    origin,
			Usage:     usage,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	if err := s.saveExternalServiceAdminHTTPAllowlist(admin); err != nil {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Type:      "system",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "system.external_service.http_allowlist.created",
		Target:    "external_service:http_allowlist#" + externalServiceHTTPAllowlistID(origin),
		Metadata: mustJSON(map[string]any{
			"origin":       origin,
			"usage":        usage,
			"operator":     firstNonEmpty(operator.Name, "system"),
			"operated_at":  now,
			"source":       externalServiceAllowlistSourceAdmin,
			"confirm_risk": true,
		}),
		CreatedAt: now,
	})
	return buildExternalServiceHTTPAllowlistResponse(admin), nil
}

func (s *Service) DeleteExternalServiceHTTPAllowlistOrigin(operator PluginExternalServiceOperator, id string) (domain.PluginExternalServiceHTTPAllowlistResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, domain.NewPluginError("external_service_http_allowlist_id_required", "allowlist id 不能为空").WithStatus(400)
	}
	admin, err := s.externalServiceAdminHTTPAllowlist()
	if err != nil {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, err
	}
	next := make([]storedExternalServiceHTTPAllowlistOrigin, 0, len(admin))
	var deleted storedExternalServiceHTTPAllowlistOrigin
	for _, item := range admin {
		if item.ID == id {
			deleted = item
			continue
		}
		next = append(next, item)
	}
	if deleted.ID == "" {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, domain.NewPluginError("external_service_http_allowlist_not_found", "后台 allowlist origin 不存在或不可删除").WithStatus(404)
	}
	if err := s.saveExternalServiceAdminHTTPAllowlist(next); err != nil {
		return domain.PluginExternalServiceHTTPAllowlistResponse{}, err
	}
	now := Now()
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Type:      "system",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "system.external_service.http_allowlist.deleted",
		Target:    "external_service:http_allowlist#" + deleted.ID,
		Metadata: mustJSON(map[string]any{
			"origin":      deleted.Origin,
			"usage":       deleted.Usage,
			"operator":    firstNonEmpty(operator.Name, "system"),
			"operated_at": now,
			"source":      externalServiceAllowlistSourceAdmin,
		}),
		CreatedAt: now,
	})
	return buildExternalServiceHTTPAllowlistResponse(next), nil
}

func (s *Service) externalServiceAdminHTTPAllowlist() ([]storedExternalServiceHTTPAllowlistOrigin, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	setting, ok := s.repo.SystemSetting(externalServiceAllowlistNamespace, externalServiceAllowlistKey)
	if !ok || strings.TrimSpace(setting.Value) == "" {
		return []storedExternalServiceHTTPAllowlistOrigin{}, nil
	}
	var items []storedExternalServiceHTTPAllowlistOrigin
	if err := json.Unmarshal([]byte(setting.Value), &items); err != nil {
		return nil, domain.NewPluginError("external_service_http_allowlist_storage_invalid", "后台 HTTP allowlist 存储格式异常").WithStatus(500)
	}
	out := make([]storedExternalServiceHTTPAllowlistOrigin, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		origin, err := validateExternalServiceHTTPAllowlistOrigin(item.Origin, false)
		if err != nil || seen[origin] {
			continue
		}
		item.Origin = origin
		item.ID = firstNonEmpty(strings.TrimSpace(item.ID), externalServiceHTTPAllowlistID(origin))
		seen[origin] = true
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Origin < out[j].Origin })
	return out, nil
}

func (s *Service) saveExternalServiceAdminHTTPAllowlist(items []storedExternalServiceHTTPAllowlistOrigin) error {
	sort.Slice(items, func(i, j int) bool { return items[i].Origin < items[j].Origin })
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = s.repo.SaveSystemSetting(domain.SystemSetting{
		Namespace: externalServiceAllowlistNamespace,
		Key:       externalServiceAllowlistKey,
		Value:     string(raw),
	})
	return err
}

func buildExternalServiceHTTPAllowlistResponse(admin []storedExternalServiceHTTPAllowlistOrigin) domain.PluginExternalServiceHTTPAllowlistResponse {
	defaults := defaultExternalServiceHTTPAllowlistEntries()
	env := envExternalServiceHTTPAllowlistEntries()
	adminEntries := make([]domain.PluginExternalServiceHTTPAllowlistEntry, 0, len(admin))
	for _, item := range admin {
		adminEntries = append(adminEntries, domain.PluginExternalServiceHTTPAllowlistEntry{
			ID:        item.ID,
			Origin:    item.Origin,
			Source:    externalServiceAllowlistSourceAdmin,
			Usage:     firstNonEmpty(item.Usage, "本地开发或受控内网联调"),
			Status:    "active",
			Deletable: true,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	effective := mergeExternalServiceHTTPAllowlistEntries(defaults, env, adminEntries)
	policy := domain.PluginExternalServiceHTTPPolicy{
		HTTPSAllowed:               true,
		LocalhostHTTPAllowed:       true,
		NonLocalHTTPNeedsAllowlist: true,
		AllowlistEnv:               "DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST",
		AllowlistOrigins:           originsFromEntries(effective),
		Defaults:                   originsFromEntries(defaults),
		EnvAllowlist:               originsFromEntries(env),
		AdminAllowlist:             originsFromEntries(adminEntries),
		EffectiveAllowlist:         originsFromEntries(effective),
	}
	return domain.PluginExternalServiceHTTPAllowlistResponse{
		Defaults:           defaults,
		EnvAllowlist:       env,
		AdminAllowlist:     adminEntries,
		EffectiveAllowlist: effective,
		Policy:             policy,
	}
}

func defaultExternalServiceHTTPAllowlistEntries() []domain.PluginExternalServiceHTTPAllowlistEntry {
	return []domain.PluginExternalServiceHTTPAllowlistEntry{
		{ID: "default_localhost", Origin: "localhost", Source: externalServiceAllowlistSourceDefault, Usage: "系统默认允许 localhost HTTP（任意端口）", Status: "active"},
		{ID: "default_127_0_0_1", Origin: "127.0.0.1", Source: externalServiceAllowlistSourceDefault, Usage: "系统默认允许 127.0.0.1 HTTP（任意端口）", Status: "active"},
		{ID: "default_ipv6_loopback", Origin: "::1", Source: externalServiceAllowlistSourceDefault, Usage: "系统默认允许 ::1 HTTP（任意端口）", Status: "active"},
	}
}

func envExternalServiceHTTPAllowlistEntries() []domain.PluginExternalServiceHTTPAllowlistEntry {
	items := splitExternalServiceAllowlist(os.Getenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST"))
	out := make([]domain.PluginExternalServiceHTTPAllowlistEntry, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		origin, err := validateExternalServiceHTTPAllowlistOrigin(item, true)
		if err != nil || seen[origin] {
			continue
		}
		seen[origin] = true
		out = append(out, domain.PluginExternalServiceHTTPAllowlistEntry{
			ID:        "env_" + externalServiceHTTPAllowlistID(origin),
			Origin:    origin,
			Source:    externalServiceAllowlistSourceEnv,
			Usage:     "启动环境变量 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST",
			Status:    "active",
			Deletable: false,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Origin < out[j].Origin })
	return out
}

func mergeExternalServiceHTTPAllowlistEntries(groups ...[]domain.PluginExternalServiceHTTPAllowlistEntry) []domain.PluginExternalServiceHTTPAllowlistEntry {
	out := []domain.PluginExternalServiceHTTPAllowlistEntry{}
	seen := map[string]bool{}
	for _, group := range groups {
		for _, item := range group {
			key := item.Source + "\x00" + item.Origin
			if item.Source != externalServiceAllowlistSourceDefault {
				key = item.Origin
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, item)
		}
	}
	return out
}

func originsFromEntries(items []domain.PluginExternalServiceHTTPAllowlistEntry) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Origin)
	}
	return out
}

func validateExternalServiceHTTPAllowlistOrigin(raw string, allowNoScheme bool) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_empty", "Origin 不能为空").WithStatus(400)
	}
	if strings.Contains(raw, "*") {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_wildcard", "HTTP allowlist 不允许 wildcard").WithStatus(400)
	}
	if strings.Contains(raw, "\\") {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_invalid", "Origin 格式不合法").WithStatus(400)
	}
	if !strings.Contains(raw, "://") {
		if !allowNoScheme {
			return "", domain.NewPluginError("external_service_http_allowlist_origin_scheme_required", "Origin 必须包含 http:// scheme").WithStatus(400)
		}
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_invalid", "Origin 格式不合法").WithStatus(400)
	}
	if !strings.EqualFold(u.Scheme, "http") {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_scheme_invalid", "HTTP allowlist 只允许 http:// scheme").WithStatus(400)
	}
	if u.User != nil {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_userinfo_forbidden", "Origin 不允许包含 userinfo").WithStatus(400)
	}
	if strings.TrimSpace(u.Hostname()) == "" {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_host_required", "Origin 必须包含 host").WithStatus(400)
	}
	if u.EscapedPath() != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_exact_required", "HTTP allowlist 只允许 exact origin，不允许 path、query 或 fragment").WithStatus(400)
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "0.0.0.0" || raw == "0.0.0.0/0" {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_unspecified_forbidden", "HTTP allowlist 不允许 0.0.0.0 或 0.0.0.0/0").WithStatus(400)
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_unspecified_forbidden", "HTTP allowlist 不允许 0.0.0.0 或未指定地址").WithStatus(400)
	}
	if strings.Contains(host, "/") || strings.Contains(raw, "/0") || strings.Contains(raw, "/16") || strings.Contains(raw, "/24") {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_cidr_forbidden", "HTTP allowlist 不允许 CIDR").WithStatus(400)
	}
	port := strings.TrimSpace(u.Port())
	if port == "" {
		port = "80"
	}
	for _, ch := range port {
		if ch < '0' || ch > '9' {
			return "", domain.NewPluginError("external_service_http_allowlist_origin_port_invalid", "Origin port 不合法").WithStatus(400)
		}
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return "", domain.NewPluginError("external_service_http_allowlist_origin_port_invalid", "Origin port 不合法").WithStatus(400)
	}
	return normalizeExternalServiceEndpointOrigin(u), nil
}

func externalServiceHTTPAllowlistID(origin string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(origin)))
	return "allow_" + hex.EncodeToString(sum[:])[:16]
}
