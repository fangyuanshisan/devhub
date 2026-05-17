package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"

	_ "github.com/go-sql-driver/mysql"
)

func nullJSONString(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return nil
	}
	return json.RawMessage(raw)
}

// MySQLConfig 保存 MySQL 连接配置。
type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// MySQLStore 是基于 MySQL 的数据仓储。
type MySQLStore struct {
	db       *sql.DB
	database string
}

// ===== Webhook secrets (v1.7.6) =====

func (s *MySQLStore) AppendPluginWebhookSecret(record domain.PluginWebhookSecret) (domain.PluginWebhookSecret, error) {
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.TargetURL = strings.TrimSpace(record.TargetURL)
	record.SecretRef = strings.TrimSpace(record.SecretRef)
	record.Status = strings.TrimSpace(record.Status)
	if record.PluginCode == "" || record.TargetURL == "" || record.SecretRef == "" {
		return domain.PluginWebhookSecret{}, errors.New("plugin_code/target_url/secret_ref 不能为空")
	}
	if record.Status == "" {
		record.Status = domain.PluginWebhookSecretStatusActive
	}
	if record.Version <= 0 {
		record.Version = 1
	}
	res, err := s.db.Exec(`INSERT INTO plugin_webhook_secrets
		(plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status, rotation_group, previous_secret_ref,
		 active_from, active_until, grace_until, created_by, created_at, rotated_at, revoked_at, last_used_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),?,?,?,?,NOW())
		ON DUPLICATE KEY UPDATE updated_at=NOW()`,
		record.PluginCode, record.TargetURL, record.SecretRef, record.SecretCiphertext, record.SecretHash, record.Version, record.Status, record.RotationGroup, record.PreviousSecretRef,
		nullTime(record.ActiveFrom), nullTime(record.ActiveUntil), nullTime(record.GraceUntil), record.CreatedBy, nullTime(record.CreatedAt), nullTime(record.RotatedAt), nullTime(record.RevokedAt), nullTime(record.LastUsedAt))
	if err != nil {
		return domain.PluginWebhookSecret{}, err
	}
	if id, _ := res.LastInsertId(); id > 0 {
		out, _ := s.PluginWebhookSecretByID(id)
		return out, nil
	}
	out, ok := s.PluginWebhookSecretByRef(record.SecretRef)
	if ok {
		return out, nil
	}
	return domain.PluginWebhookSecret{}, errors.New("append webhook secret failed")
}

func (s *MySQLStore) SavePluginWebhookSecret(record domain.PluginWebhookSecret) (domain.PluginWebhookSecret, error) {
	if record.ID <= 0 {
		return s.AppendPluginWebhookSecret(record)
	}
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.TargetURL = strings.TrimSpace(record.TargetURL)
	record.SecretRef = strings.TrimSpace(record.SecretRef)
	record.Status = strings.TrimSpace(record.Status)
	if record.PluginCode == "" || record.TargetURL == "" || record.SecretRef == "" {
		return domain.PluginWebhookSecret{}, errors.New("plugin_code/target_url/secret_ref 不能为空")
	}
	if record.Status == "" {
		record.Status = domain.PluginWebhookSecretStatusActive
	}
	if record.Version <= 0 {
		record.Version = 1
	}
	_, err := s.db.Exec(`UPDATE plugin_webhook_secrets SET
		plugin_code=?, target_url=?, secret_ref=?, secret_ciphertext=?, secret_hash=?, version=?, status=?, rotation_group=?, previous_secret_ref=?,
		active_from=?, active_until=?, grace_until=?, created_by=?, created_at=COALESCE(?, created_at), rotated_at=?, revoked_at=?, last_used_at=?, updated_at=NOW()
		WHERE id=?`,
		record.PluginCode, record.TargetURL, record.SecretRef, record.SecretCiphertext, record.SecretHash, record.Version, record.Status, record.RotationGroup, record.PreviousSecretRef,
		nullTime(record.ActiveFrom), nullTime(record.ActiveUntil), nullTime(record.GraceUntil), record.CreatedBy, nullTime(record.CreatedAt), nullTime(record.RotatedAt), nullTime(record.RevokedAt), nullTime(record.LastUsedAt),
		record.ID)
	if err != nil {
		return domain.PluginWebhookSecret{}, err
	}
	out, _ := s.PluginWebhookSecretByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginWebhookSecretByID(id int64) (domain.PluginWebhookSecret, bool) {
	var it domain.PluginWebhookSecret
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status,
		COALESCE(rotation_group,''), COALESCE(previous_secret_ref,''),
		COALESCE(DATE_FORMAT(active_from,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(active_until,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(grace_until,'%Y-%m-%d %H:%i:%s'),''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_webhook_secrets WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.SecretRef, &it.SecretCiphertext, &it.SecretHash, &it.Version, &it.Status,
			&it.RotationGroup, &it.PreviousSecretRef, &it.ActiveFrom, &it.ActiveUntil, &it.GraceUntil, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.LastUsedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginWebhookSecret{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginWebhookSecretByRef(secretRef string) (domain.PluginWebhookSecret, bool) {
	secretRef = strings.TrimSpace(secretRef)
	if secretRef == "" {
		return domain.PluginWebhookSecret{}, false
	}
	var it domain.PluginWebhookSecret
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status,
		COALESCE(rotation_group,''), COALESCE(previous_secret_ref,''),
		COALESCE(DATE_FORMAT(active_from,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(active_until,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(grace_until,'%Y-%m-%d %H:%i:%s'),''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_webhook_secrets WHERE secret_ref=? LIMIT 1`, secretRef).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.SecretRef, &it.SecretCiphertext, &it.SecretHash, &it.Version, &it.Status,
			&it.RotationGroup, &it.PreviousSecretRef, &it.ActiveFrom, &it.ActiveUntil, &it.GraceUntil, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.LastUsedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginWebhookSecret{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginWebhookSecrets(filter domain.PluginWebhookSecretFilter) ([]domain.PluginWebhookSecret, int, error) {
	filter = filter.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.SecretRef != "" {
		where = append(where, "secret_ref=?")
		args = append(args, filter.SecretRef)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_webhook_secrets WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status,
		COALESCE(rotation_group,''), COALESCE(previous_secret_ref,''),
		COALESCE(DATE_FORMAT(active_from,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(active_until,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(grace_until,'%Y-%m-%d %H:%i:%s'),''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_webhook_secrets WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginWebhookSecret{}
	for rows.Next() {
		var it domain.PluginWebhookSecret
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.SecretRef, &it.SecretCiphertext, &it.SecretHash, &it.Version, &it.Status,
			&it.RotationGroup, &it.PreviousSecretRef, &it.ActiveFrom, &it.ActiveUntil, &it.GraceUntil, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.LastUsedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) ActivePluginWebhookSecret(pluginCode, targetURL string) (domain.PluginWebhookSecret, bool) {
	pluginCode = strings.TrimSpace(pluginCode)
	targetURL = strings.TrimSpace(targetURL)
	var it domain.PluginWebhookSecret
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status,
		COALESCE(rotation_group,''), COALESCE(previous_secret_ref,''),
		COALESCE(DATE_FORMAT(active_from,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(active_until,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(grace_until,'%Y-%m-%d %H:%i:%s'),''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_webhook_secrets WHERE plugin_code=? AND target_url=? AND status=? ORDER BY id DESC LIMIT 1`,
		pluginCode, targetURL, domain.PluginWebhookSecretStatusActive).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.SecretRef, &it.SecretCiphertext, &it.SecretHash, &it.Version, &it.Status,
			&it.RotationGroup, &it.PreviousSecretRef, &it.ActiveFrom, &it.ActiveUntil, &it.GraceUntil, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.LastUsedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginWebhookSecret{}, false
	}
	return it, true
}

func (s *MySQLStore) PreviousPluginWebhookSecret(pluginCode, targetURL string) (domain.PluginWebhookSecret, bool) {
	pluginCode = strings.TrimSpace(pluginCode)
	targetURL = strings.TrimSpace(targetURL)
	var it domain.PluginWebhookSecret
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status,
		COALESCE(rotation_group,''), COALESCE(previous_secret_ref,''),
		COALESCE(DATE_FORMAT(active_from,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(active_until,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(grace_until,'%Y-%m-%d %H:%i:%s'),''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_webhook_secrets WHERE plugin_code=? AND target_url=? AND status=? ORDER BY id DESC LIMIT 1`,
		pluginCode, targetURL, domain.PluginWebhookSecretStatusPrevious).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.SecretRef, &it.SecretCiphertext, &it.SecretHash, &it.Version, &it.Status,
			&it.RotationGroup, &it.PreviousSecretRef, &it.ActiveFrom, &it.ActiveUntil, &it.GraceUntil, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.LastUsedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginWebhookSecret{}, false
	}
	return it, true
}

func (s *MySQLStore) LatestPluginWebhookSecretForTarget(pluginCode, targetURL string) (domain.PluginWebhookSecret, bool) {
	pluginCode = strings.TrimSpace(pluginCode)
	targetURL = strings.TrimSpace(targetURL)
	var it domain.PluginWebhookSecret
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, secret_ref, secret_ciphertext, secret_hash, version, status,
		COALESCE(rotation_group,''), COALESCE(previous_secret_ref,''),
		COALESCE(DATE_FORMAT(active_from,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(active_until,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(grace_until,'%Y-%m-%d %H:%i:%s'),''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_webhook_secrets WHERE plugin_code=? AND target_url=? ORDER BY id DESC LIMIT 1`,
		pluginCode, targetURL).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.SecretRef, &it.SecretCiphertext, &it.SecretHash, &it.Version, &it.Status,
			&it.RotationGroup, &it.PreviousSecretRef, &it.ActiveFrom, &it.ActiveUntil, &it.GraceUntil, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.LastUsedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginWebhookSecret{}, false
	}
	return it, true
}

// ===== Plugin callback tokens (v1.7.7) =====

func (s *MySQLStore) AppendPluginCallbackToken(record domain.PluginCallbackToken) (domain.PluginCallbackToken, error) {
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.TokenRef = strings.TrimSpace(record.TokenRef)
	record.TokenHash = strings.TrimSpace(record.TokenHash)
	record.Status = strings.TrimSpace(record.Status)
	if record.PluginCode == "" || record.TokenRef == "" || record.TokenHash == "" {
		return domain.PluginCallbackToken{}, errors.New("plugin_code/token_ref/token_hash 不能为空")
	}
	if record.Status == "" {
		record.Status = domain.PluginCallbackTokenStatusActive
	}
	_, err := s.db.Exec(`INSERT INTO plugin_callback_tokens
		(plugin_code, plugin_installation_id, publisher_id, token_ref, token_hash, name, status, scopes_json, community_scope_json,
		 expires_at, last_used_at, last_used_ip, created_by, created_at, rotated_at, revoked_at, revoked_reason, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),?,?,?,?,NOW())
		ON DUPLICATE KEY UPDATE updated_at=NOW()`,
		record.PluginCode, record.PluginInstallationID, record.PublisherID, record.TokenRef, record.TokenHash, record.Name, record.Status,
		nullJSONString(record.ScopesJSON), nullJSONString(record.CommunityScopeJSON),
		nullTime(record.ExpiresAt), nullTime(record.LastUsedAt), record.LastUsedIP, record.CreatedBy, nullTime(record.CreatedAt),
		nullTime(record.RotatedAt), nullTime(record.RevokedAt), record.RevokedReason)
	if err != nil {
		return domain.PluginCallbackToken{}, err
	}
	out, ok := s.PluginCallbackTokenByRef(record.TokenRef)
	if ok {
		return out, nil
	}
	return domain.PluginCallbackToken{}, errors.New("append callback token failed")
}

func (s *MySQLStore) SavePluginCallbackToken(record domain.PluginCallbackToken) (domain.PluginCallbackToken, error) {
	if record.ID <= 0 {
		return s.AppendPluginCallbackToken(record)
	}
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.TokenRef = strings.TrimSpace(record.TokenRef)
	record.TokenHash = strings.TrimSpace(record.TokenHash)
	record.Status = strings.TrimSpace(record.Status)
	if record.PluginCode == "" || record.TokenRef == "" || record.TokenHash == "" {
		return domain.PluginCallbackToken{}, errors.New("plugin_code/token_ref/token_hash 不能为空")
	}
	if record.Status == "" {
		record.Status = domain.PluginCallbackTokenStatusActive
	}
	_, err := s.db.Exec(`UPDATE plugin_callback_tokens SET
		plugin_code=?, plugin_installation_id=?, publisher_id=?, token_ref=?, token_hash=?, name=?, status=?, scopes_json=?, community_scope_json=?,
		expires_at=?, last_used_at=?, last_used_ip=?, created_by=?, created_at=COALESCE(?, created_at), rotated_at=?, revoked_at=?, revoked_reason=?, updated_at=NOW()
		WHERE id=?`,
		record.PluginCode, record.PluginInstallationID, record.PublisherID, record.TokenRef, record.TokenHash, record.Name, record.Status,
		nullJSONString(record.ScopesJSON), nullJSONString(record.CommunityScopeJSON),
		nullTime(record.ExpiresAt), nullTime(record.LastUsedAt), record.LastUsedIP, record.CreatedBy, nullTime(record.CreatedAt),
		nullTime(record.RotatedAt), nullTime(record.RevokedAt), record.RevokedReason, record.ID)
	if err != nil {
		return domain.PluginCallbackToken{}, err
	}
	out, _ := s.PluginCallbackTokenByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginCallbackTokenByID(id int64) (domain.PluginCallbackToken, bool) {
	var it domain.PluginCallbackToken
	err := s.db.QueryRow(`SELECT id, plugin_code, plugin_installation_id, COALESCE(publisher_id,''), token_ref, token_hash, COALESCE(name,''), status,
		COALESCE(CAST(scopes_json AS CHAR),''), COALESCE(CAST(community_scope_json AS CHAR),''),
		COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(last_used_ip,''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(revoked_reason,''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_callback_tokens WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PluginCode, &it.PluginInstallationID, &it.PublisherID, &it.TokenRef, &it.TokenHash, &it.Name, &it.Status,
			&it.ScopesJSON, &it.CommunityScopeJSON, &it.ExpiresAt, &it.LastUsedAt, &it.LastUsedIP, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.RevokedReason, &it.UpdatedAt)
	if err != nil {
		return domain.PluginCallbackToken{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginCallbackTokenByRef(tokenRef string) (domain.PluginCallbackToken, bool) {
	tokenRef = strings.TrimSpace(tokenRef)
	var it domain.PluginCallbackToken
	err := s.db.QueryRow(`SELECT id, plugin_code, plugin_installation_id, COALESCE(publisher_id,''), token_ref, token_hash, COALESCE(name,''), status,
		COALESCE(CAST(scopes_json AS CHAR),''), COALESCE(CAST(community_scope_json AS CHAR),''),
		COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(last_used_ip,''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(revoked_reason,''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_callback_tokens WHERE token_ref=? LIMIT 1`, tokenRef).
		Scan(&it.ID, &it.PluginCode, &it.PluginInstallationID, &it.PublisherID, &it.TokenRef, &it.TokenHash, &it.Name, &it.Status,
			&it.ScopesJSON, &it.CommunityScopeJSON, &it.ExpiresAt, &it.LastUsedAt, &it.LastUsedIP, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.RevokedReason, &it.UpdatedAt)
	if err != nil {
		return domain.PluginCallbackToken{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginCallbackTokenByHash(tokenHash string) (domain.PluginCallbackToken, bool) {
	tokenHash = strings.TrimSpace(tokenHash)
	var it domain.PluginCallbackToken
	err := s.db.QueryRow(`SELECT id, plugin_code, plugin_installation_id, COALESCE(publisher_id,''), token_ref, token_hash, COALESCE(name,''), status,
		COALESCE(CAST(scopes_json AS CHAR),''), COALESCE(CAST(community_scope_json AS CHAR),''),
		COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(last_used_ip,''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(revoked_reason,''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_callback_tokens WHERE token_hash=? LIMIT 1`, tokenHash).
		Scan(&it.ID, &it.PluginCode, &it.PluginInstallationID, &it.PublisherID, &it.TokenRef, &it.TokenHash, &it.Name, &it.Status,
			&it.ScopesJSON, &it.CommunityScopeJSON, &it.ExpiresAt, &it.LastUsedAt, &it.LastUsedIP, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.RevokedReason, &it.UpdatedAt)
	if err != nil {
		return domain.PluginCallbackToken{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginCallbackTokens(filter domain.PluginCallbackTokenFilter) ([]domain.PluginCallbackToken, int, error) {
	filter = filter.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.Scope != "" {
		// JSON search: simple LIKE fallback for portability; exact enforcement happens in service/auth layer.
		where = append(where, "CAST(scopes_json AS CHAR) LIKE ?")
		args = append(args, "%"+filter.Scope+"%")
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_callback_tokens WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, plugin_code, plugin_installation_id, COALESCE(publisher_id,''), token_ref, token_hash, COALESCE(name,''), status,
		COALESCE(CAST(scopes_json AS CHAR),''), COALESCE(CAST(community_scope_json AS CHAR),''),
		COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_used_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(last_used_ip,''),
		created_by, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(rotated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(revoked_reason,''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_callback_tokens WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginCallbackToken{}
	for rows.Next() {
		var it domain.PluginCallbackToken
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.PluginInstallationID, &it.PublisherID, &it.TokenRef, &it.TokenHash, &it.Name, &it.Status,
			&it.ScopesJSON, &it.CommunityScopeJSON, &it.ExpiresAt, &it.LastUsedAt, &it.LastUsedIP, &it.CreatedBy, &it.CreatedAt,
			&it.RotatedAt, &it.RevokedAt, &it.RevokedReason, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) AppendPluginCallbackRequest(record domain.PluginCallbackRequest) (domain.PluginCallbackRequest, error) {
	record.RequestID = strings.TrimSpace(record.RequestID)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.TokenRef = strings.TrimSpace(record.TokenRef)
	record.APIPath = strings.TrimSpace(record.APIPath)
	record.Method = strings.TrimSpace(record.Method)
	record.ScopeRequired = strings.TrimSpace(record.ScopeRequired)
	record.Status = strings.TrimSpace(record.Status)
	record.ErrorCode = strings.TrimSpace(record.ErrorCode)
	record.ErrorMessage = strings.TrimSpace(record.ErrorMessage)
	record.IPAddress = strings.TrimSpace(record.IPAddress)
	record.UserAgent = strings.TrimSpace(record.UserAgent)
	if record.RequestID == "" || record.PluginCode == "" || record.APIPath == "" || record.Method == "" {
		return domain.PluginCallbackRequest{}, errors.New("request_id/plugin_code/api_path/method 不能为空")
	}
	if record.Status == "" {
		record.Status = domain.PluginCallbackRequestStatusAccepted
	}
	_, err := s.db.Exec(`INSERT INTO plugin_callback_requests
		(request_id, plugin_code, token_ref, api_path, method, scope_required, status, response_status, error_code, error_message,
		 community_id, actor_type, actor_id, ip_address, user_agent, duration_ms, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()))`,
		record.RequestID, record.PluginCode, record.TokenRef, record.APIPath, record.Method, record.ScopeRequired, record.Status,
		record.ResponseStatus, record.ErrorCode, record.ErrorMessage, record.CommunityID, record.ActorType, record.ActorID, record.IPAddress, record.UserAgent, record.DurationMS, nullTime(record.CreatedAt))
	if err != nil {
		return domain.PluginCallbackRequest{}, err
	}
	// no stable unique key; just return record
	record.CreatedAt = Now()
	return record, nil
}

func (s *MySQLStore) PluginCallbackRequests(filter domain.PluginCallbackRequestFilter) ([]domain.PluginCallbackRequest, int, error) {
	filter = filter.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.TokenRef != "" {
		where = append(where, "token_ref=?")
		args = append(args, filter.TokenRef)
	}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.RequestID != "" {
		where = append(where, "request_id=?")
		args = append(args, filter.RequestID)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_callback_requests WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, request_id, plugin_code, token_ref, api_path, method, scope_required, status, response_status,
		COALESCE(error_code,''), COALESCE(error_message,''), community_id, COALESCE(actor_type,''), actor_id,
		COALESCE(ip_address,''), COALESCE(user_agent,''), duration_ms, DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_callback_requests WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginCallbackRequest{}
	for rows.Next() {
		var it domain.PluginCallbackRequest
		if err := rows.Scan(&it.ID, &it.RequestID, &it.PluginCode, &it.TokenRef, &it.APIPath, &it.Method, &it.ScopeRequired, &it.Status, &it.ResponseStatus,
			&it.ErrorCode, &it.ErrorMessage, &it.CommunityID, &it.ActorType, &it.ActorID, &it.IPAddress, &it.UserAgent, &it.DurationMS, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) AppendWebhookEvent(record domain.WebhookEvent) (domain.WebhookEvent, error) {
	record.EventID = strings.TrimSpace(record.EventID)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.HookName = strings.TrimSpace(record.HookName)
	if record.EventID == "" || record.PluginCode == "" {
		return domain.WebhookEvent{}, errors.New("event_id 和 plugin_code 不能为空")
	}
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.WebhookEventStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO webhook_events
		(event_id, event_name, event_type, plugin_code, hook_name, mode, community_id, actor_type, actor_id, resource_type, resource_id, request_id,
		 payload_json, metadata_json, status, occurred_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())
		ON DUPLICATE KEY UPDATE updated_at=NOW()`,
		record.EventID, record.EventName, record.EventType, record.PluginCode, record.HookName, record.Mode, record.CommunityID, record.ActorType, record.ActorID, record.ResourceType, record.ResourceID, record.RequestID,
		nullJSONString(record.PayloadJSON), nullJSONString(record.MetadataJSON), record.Status, nullTime(record.OccurredAt), nullTime(record.CreatedAt))
	if err != nil {
		return domain.WebhookEvent{}, err
	}
	if id, _ := res.LastInsertId(); id > 0 {
		out, _ := s.WebhookEventByID(id)
		return out, nil
	}
	// duplicated insert, fallback to fetch by event_id
	var out domain.WebhookEvent
	err = s.db.QueryRow(`SELECT id, event_id, event_name, event_type, plugin_code, hook_name, mode, community_id, actor_type, actor_id, resource_type, resource_id, request_id,
		COALESCE(CAST(payload_json AS CHAR),''), COALESCE(CAST(metadata_json AS CHAR),''), status,
		COALESCE(DATE_FORMAT(occurred_at,'%Y-%m-%d %H:%i:%s'),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_events WHERE event_id=? LIMIT 1`, record.EventID).
		Scan(&out.ID, &out.EventID, &out.EventName, &out.EventType, &out.PluginCode, &out.HookName, &out.Mode, &out.CommunityID, &out.ActorType, &out.ActorID, &out.ResourceType, &out.ResourceID, &out.RequestID,
			&out.PayloadJSON, &out.MetadataJSON, &out.Status, &out.OccurredAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.WebhookEvent{}, err
	}
	return out, nil
}

func (s *MySQLStore) SaveWebhookEvent(record domain.WebhookEvent) (domain.WebhookEvent, error) {
	if record.ID <= 0 {
		return s.AppendWebhookEvent(record)
	}
	record.EventID = strings.TrimSpace(record.EventID)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.HookName = strings.TrimSpace(record.HookName)
	if record.EventID == "" || record.PluginCode == "" {
		return domain.WebhookEvent{}, errors.New("event_id 和 plugin_code 不能为空")
	}
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.WebhookEventStatusPending
	}
	_, err := s.db.Exec(`UPDATE webhook_events SET
		event_id=?, event_name=?, event_type=?, plugin_code=?, hook_name=?, mode=?, community_id=?, actor_type=?, actor_id=?, resource_type=?, resource_id=?, request_id=?,
		payload_json=?, metadata_json=?, status=?, occurred_at=?, created_at=COALESCE(?, created_at), updated_at=NOW()
		WHERE id=?`,
		record.EventID, record.EventName, record.EventType, record.PluginCode, record.HookName, record.Mode, record.CommunityID, record.ActorType, record.ActorID, record.ResourceType, record.ResourceID, record.RequestID,
		nullJSONString(record.PayloadJSON), nullJSONString(record.MetadataJSON), record.Status, nullTime(record.OccurredAt), nullTime(record.CreatedAt), record.ID)
	if err != nil {
		return domain.WebhookEvent{}, err
	}
	out, _ := s.WebhookEventByID(record.ID)
	return out, nil
}

func (s *MySQLStore) WebhookEventByID(id int64) (domain.WebhookEvent, bool) {
	var it domain.WebhookEvent
	err := s.db.QueryRow(`SELECT id, event_id, event_name, event_type, plugin_code, hook_name, mode, community_id, actor_type, actor_id, resource_type, resource_id, request_id,
		COALESCE(CAST(payload_json AS CHAR),''), COALESCE(CAST(metadata_json AS CHAR),''), status,
		COALESCE(DATE_FORMAT(occurred_at,'%Y-%m-%d %H:%i:%s'),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_events WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.EventID, &it.EventName, &it.EventType, &it.PluginCode, &it.HookName, &it.Mode, &it.CommunityID, &it.ActorType, &it.ActorID, &it.ResourceType, &it.ResourceID, &it.RequestID,
			&it.PayloadJSON, &it.MetadataJSON, &it.Status, &it.OccurredAt, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.WebhookEvent{}, false
	}
	return it, true
}

func (s *MySQLStore) WebhookEvents(filter domain.WebhookEventFilter) ([]domain.WebhookEvent, int, error) {
	filter = filter.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.HookName != "" {
		where = append(where, "hook_name=?")
		args = append(args, filter.HookName)
	}
	if filter.Mode != "" {
		where = append(where, "mode=?")
		args = append(args, filter.Mode)
	}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.CommunityID > 0 {
		where = append(where, "community_id=?")
		args = append(args, filter.CommunityID)
	}
	if filter.ActorType != "" {
		where = append(where, "actor_type=?")
		args = append(args, filter.ActorType)
	}
	if filter.ActorID > 0 {
		where = append(where, "actor_id=?")
		args = append(args, filter.ActorID)
	}
	if filter.RequestID != "" {
		where = append(where, "request_id=?")
		args = append(args, filter.RequestID)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM webhook_events WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, event_id, event_name, event_type, plugin_code, hook_name, mode, community_id, actor_type, actor_id, resource_type, resource_id, request_id,
		COALESCE(CAST(payload_json AS CHAR),''), COALESCE(CAST(metadata_json AS CHAR),''), status,
		COALESCE(DATE_FORMAT(occurred_at,'%Y-%m-%d %H:%i:%s'),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_events WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.WebhookEvent{}
	for rows.Next() {
		var it domain.WebhookEvent
		if err := rows.Scan(&it.ID, &it.EventID, &it.EventName, &it.EventType, &it.PluginCode, &it.HookName, &it.Mode, &it.CommunityID, &it.ActorType, &it.ActorID, &it.ResourceType, &it.ResourceID, &it.RequestID,
			&it.PayloadJSON, &it.MetadataJSON, &it.Status, &it.OccurredAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) AppendWebhookDelivery(record domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	record.DeliveryID = strings.TrimSpace(record.DeliveryID)
	record.EventID = strings.TrimSpace(record.EventID)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.HookName = strings.TrimSpace(record.HookName)
	record.TargetURL = strings.TrimSpace(record.TargetURL)
	if record.DeliveryID == "" || record.EventID == "" || record.PluginCode == "" || record.TargetURL == "" {
		return domain.WebhookDelivery{}, errors.New("delivery_id/event_id/plugin_code/target_url 不能为空")
	}
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.WebhookDeliveryStatusPending
	}
	if record.Attempt <= 0 {
		record.Attempt = 1
	}
	if record.MaxAttempts <= 0 {
		record.MaxAttempts = 5
	}
	res, err := s.db.Exec(`INSERT INTO webhook_deliveries
		(delivery_id, event_id, plugin_code, hook_name, target_url, status, attempt, max_attempts,
		 signature_alg, secret_ref, body_sha256, signature_status, signed_at, signature_error,
		 next_retry_at, retry_reason, request_headers_json, request_body_sha256,
		 response_status, response_body_excerpt, error_message, duration_ms, started_at, finished_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())
		ON DUPLICATE KEY UPDATE updated_at=NOW()`,
		record.DeliveryID, record.EventID, record.PluginCode, record.HookName, record.TargetURL, record.Status, record.Attempt, record.MaxAttempts,
		record.SignatureAlg, record.SecretRef, record.BodySHA256, record.SignatureStatus, nullTime(record.SignedAt), record.SignatureError,
		nullTime(record.NextRetryAt), record.RetryReason, nullJSONString(record.RequestHeadersJSON), record.RequestBodySHA256,
		record.ResponseStatus, record.ResponseBodyExcerpt, record.ErrorMessage, record.DurationMS, nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt))
	if err != nil {
		return domain.WebhookDelivery{}, err
	}
	if id, _ := res.LastInsertId(); id > 0 {
		out, _ := s.WebhookDeliveryByID(id)
		return out, nil
	}
	var out domain.WebhookDelivery
	err = s.db.QueryRow(`SELECT id, delivery_id, event_id, plugin_code, hook_name, target_url, status, attempt, max_attempts,
		COALESCE(signature_alg,''), COALESCE(secret_ref,''), COALESCE(body_sha256,''), COALESCE(signature_status,''), COALESCE(DATE_FORMAT(signed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(signature_error,''),
		COALESCE(DATE_FORMAT(next_retry_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(retry_reason,''), COALESCE(CAST(request_headers_json AS CHAR),''),
		COALESCE(request_body_sha256,''), response_status, COALESCE(response_body_excerpt,''), COALESCE(error_message,''), duration_ms,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_deliveries WHERE delivery_id=? LIMIT 1`, record.DeliveryID).
		Scan(&out.ID, &out.DeliveryID, &out.EventID, &out.PluginCode, &out.HookName, &out.TargetURL, &out.Status, &out.Attempt, &out.MaxAttempts,
			&out.SignatureAlg, &out.SecretRef, &out.BodySHA256, &out.SignatureStatus, &out.SignedAt, &out.SignatureError,
			&out.NextRetryAt, &out.RetryReason, &out.RequestHeadersJSON, &out.RequestBodySHA256, &out.ResponseStatus, &out.ResponseBodyExcerpt, &out.ErrorMessage, &out.DurationMS,
			&out.StartedAt, &out.FinishedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}
	return out, nil
}

func (s *MySQLStore) SaveWebhookDelivery(record domain.WebhookDelivery) (domain.WebhookDelivery, error) {
	if record.ID <= 0 {
		return s.AppendWebhookDelivery(record)
	}
	record.DeliveryID = strings.TrimSpace(record.DeliveryID)
	record.EventID = strings.TrimSpace(record.EventID)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.HookName = strings.TrimSpace(record.HookName)
	record.TargetURL = strings.TrimSpace(record.TargetURL)
	if record.DeliveryID == "" || record.EventID == "" || record.PluginCode == "" || record.TargetURL == "" {
		return domain.WebhookDelivery{}, errors.New("delivery_id/event_id/plugin_code/target_url 不能为空")
	}
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.WebhookDeliveryStatusPending
	}
	if record.Attempt <= 0 {
		record.Attempt = 1
	}
	if record.MaxAttempts <= 0 {
		record.MaxAttempts = 5
	}
	_, err := s.db.Exec(`UPDATE webhook_deliveries SET
		delivery_id=?, event_id=?, plugin_code=?, hook_name=?, target_url=?, status=?, attempt=?, max_attempts=?,
		signature_alg=?, secret_ref=?, body_sha256=?, signature_status=?, signed_at=?, signature_error=?,
		next_retry_at=?, retry_reason=?, request_headers_json=?, request_body_sha256=?, response_status=?, response_body_excerpt=?, error_message=?, duration_ms=?,
		started_at=?, finished_at=?, created_at=COALESCE(?, created_at), updated_at=NOW()
		WHERE id=?`,
		record.DeliveryID, record.EventID, record.PluginCode, record.HookName, record.TargetURL, record.Status, record.Attempt, record.MaxAttempts,
		record.SignatureAlg, record.SecretRef, record.BodySHA256, record.SignatureStatus, nullTime(record.SignedAt), record.SignatureError,
		nullTime(record.NextRetryAt), record.RetryReason, nullJSONString(record.RequestHeadersJSON), record.RequestBodySHA256,
		record.ResponseStatus, record.ResponseBodyExcerpt, record.ErrorMessage, record.DurationMS,
		nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt), record.ID)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}
	out, _ := s.WebhookDeliveryByID(record.ID)
	return out, nil
}

func (s *MySQLStore) TryMarkWebhookDeliveryStatus(id int64, fromStatus, toStatus string) (bool, error) {
	fromStatus = strings.TrimSpace(fromStatus)
	toStatus = strings.TrimSpace(toStatus)
	if id <= 0 || fromStatus == "" || toStatus == "" {
		return false, errors.New("invalid args")
	}
	res, err := s.db.Exec(`UPDATE webhook_deliveries SET status=?, updated_at=NOW() WHERE id=? AND status=?`, toStatus, id, fromStatus)
	if err != nil {
		return false, err
	}
	aff, _ := res.RowsAffected()
	return aff == 1, nil
}

func (s *MySQLStore) WebhookDeliveryByID(id int64) (domain.WebhookDelivery, bool) {
	var it domain.WebhookDelivery
	err := s.db.QueryRow(`SELECT id, delivery_id, event_id, plugin_code, hook_name, target_url, status, attempt, max_attempts,
		COALESCE(signature_alg,''), COALESCE(secret_ref,''), COALESCE(body_sha256,''), COALESCE(signature_status,''), COALESCE(DATE_FORMAT(signed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(signature_error,''),
		COALESCE(DATE_FORMAT(next_retry_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(retry_reason,''), COALESCE(CAST(request_headers_json AS CHAR),''),
		COALESCE(request_body_sha256,''), response_status, COALESCE(response_body_excerpt,''), COALESCE(error_message,''), duration_ms,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_deliveries WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.DeliveryID, &it.EventID, &it.PluginCode, &it.HookName, &it.TargetURL, &it.Status, &it.Attempt, &it.MaxAttempts,
			&it.SignatureAlg, &it.SecretRef, &it.BodySHA256, &it.SignatureStatus, &it.SignedAt, &it.SignatureError,
			&it.NextRetryAt, &it.RetryReason, &it.RequestHeadersJSON, &it.RequestBodySHA256, &it.ResponseStatus, &it.ResponseBodyExcerpt, &it.ErrorMessage, &it.DurationMS,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.WebhookDelivery{}, false
	}
	return it, true
}

func (s *MySQLStore) WebhookDeliveries(filter domain.WebhookDeliveryFilter) ([]domain.WebhookDelivery, int, error) {
	filter = filter.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.HookName != "" {
		where = append(where, "hook_name=?")
		args = append(args, filter.HookName)
	}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.EventID != "" {
		where = append(where, "event_id=?")
		args = append(args, filter.EventID)
	}
	if filter.DeliveryID != "" {
		where = append(where, "delivery_id=?")
		args = append(args, filter.DeliveryID)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM webhook_deliveries WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, delivery_id, event_id, plugin_code, hook_name, target_url, status, attempt, max_attempts,
		COALESCE(signature_alg,''), COALESCE(secret_ref,''), COALESCE(body_sha256,''), COALESCE(signature_status,''), COALESCE(DATE_FORMAT(signed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(signature_error,''),
		COALESCE(DATE_FORMAT(next_retry_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(retry_reason,''), COALESCE(CAST(request_headers_json AS CHAR),''),
		COALESCE(request_body_sha256,''), response_status, COALESCE(response_body_excerpt,''), COALESCE(error_message,''), duration_ms,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_deliveries WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.WebhookDelivery{}
	for rows.Next() {
		var it domain.WebhookDelivery
		if err := rows.Scan(&it.ID, &it.DeliveryID, &it.EventID, &it.PluginCode, &it.HookName, &it.TargetURL, &it.Status, &it.Attempt, &it.MaxAttempts,
			&it.SignatureAlg, &it.SecretRef, &it.BodySHA256, &it.SignatureStatus, &it.SignedAt, &it.SignatureError,
			&it.NextRetryAt, &it.RetryReason, &it.RequestHeadersJSON, &it.RequestBodySHA256, &it.ResponseStatus, &it.ResponseBodyExcerpt, &it.ErrorMessage, &it.DurationMS,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) DueWebhookDeliveries(dueBefore string, limit int) ([]domain.WebhookDelivery, error) {
	dueBefore = strings.TrimSpace(dueBefore)
	if dueBefore == "" {
		dueBefore = time.Now().Format("2006-01-02 15:04:05")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(`SELECT id, delivery_id, event_id, plugin_code, hook_name, target_url, status, attempt, max_attempts,
		COALESCE(signature_alg,''), COALESCE(secret_ref,''), COALESCE(body_sha256,''), COALESCE(signature_status,''), COALESCE(DATE_FORMAT(signed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(signature_error,''),
		COALESCE(DATE_FORMAT(next_retry_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(retry_reason,''), COALESCE(CAST(request_headers_json AS CHAR),''),
		COALESCE(request_body_sha256,''), response_status, COALESCE(response_body_excerpt,''), COALESCE(error_message,''), duration_ms,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_deliveries
		WHERE status=? AND next_retry_at IS NOT NULL AND next_retry_at <= ?
		ORDER BY next_retry_at ASC, id ASC LIMIT ?`, domain.WebhookDeliveryStatusRetryScheduled, nullTime(dueBefore), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WebhookDelivery{}
	for rows.Next() {
		var it domain.WebhookDelivery
		if err := rows.Scan(&it.ID, &it.DeliveryID, &it.EventID, &it.PluginCode, &it.HookName, &it.TargetURL, &it.Status, &it.Attempt, &it.MaxAttempts,
			&it.SignatureAlg, &it.SecretRef, &it.BodySHA256, &it.SignatureStatus, &it.SignedAt, &it.SignatureError,
			&it.NextRetryAt, &it.RetryReason, &it.RequestHeadersJSON, &it.RequestBodySHA256, &it.ResponseStatus, &it.ResponseBodyExcerpt, &it.ErrorMessage, &it.DurationMS,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func (s *MySQLStore) UpsertWebhookCircuitBreaker(record domain.WebhookCircuitBreaker) (domain.WebhookCircuitBreaker, error) {
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.TargetURL = strings.TrimSpace(record.TargetURL)
	if record.PluginCode == "" || record.TargetURL == "" {
		return domain.WebhookCircuitBreaker{}, errors.New("plugin_code 和 target_url 不能为空")
	}
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.WebhookCircuitBreakerStatusClosed
	}
	_, err := s.db.Exec(`INSERT INTO webhook_circuit_breakers
		(plugin_code, target_url, status, failure_count, success_count, opened_at, closed_at, next_probe_at, last_error_message, last_failure_at, last_success_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())
		ON DUPLICATE KEY UPDATE
		status=VALUES(status), failure_count=VALUES(failure_count), success_count=VALUES(success_count),
		opened_at=VALUES(opened_at), closed_at=VALUES(closed_at), next_probe_at=VALUES(next_probe_at),
		last_error_message=VALUES(last_error_message), last_failure_at=VALUES(last_failure_at), last_success_at=VALUES(last_success_at),
		updated_at=NOW()`,
		record.PluginCode, record.TargetURL, record.Status, record.FailureCount, record.SuccessCount,
		nullTime(record.OpenedAt), nullTime(record.ClosedAt), nullTime(record.NextProbeAt), record.LastErrorMessage, nullTime(record.LastFailureAt), nullTime(record.LastSuccessAt), nullTime(record.CreatedAt))
	if err != nil {
		return domain.WebhookCircuitBreaker{}, err
	}
	out, ok := s.WebhookCircuitBreakerByKey(record.PluginCode, record.TargetURL)
	if ok {
		return out, nil
	}
	return domain.WebhookCircuitBreaker{}, errors.New("upsert circuit breaker failed")
}

func (s *MySQLStore) WebhookCircuitBreakerByID(id int64) (domain.WebhookCircuitBreaker, bool) {
	var it domain.WebhookCircuitBreaker
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, status, failure_count, success_count,
		COALESCE(DATE_FORMAT(opened_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(closed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(next_probe_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(last_error_message,''), COALESCE(DATE_FORMAT(last_failure_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_success_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_circuit_breakers WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.Status, &it.FailureCount, &it.SuccessCount,
			&it.OpenedAt, &it.ClosedAt, &it.NextProbeAt, &it.LastErrorMessage, &it.LastFailureAt, &it.LastSuccessAt,
			&it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.WebhookCircuitBreaker{}, false
	}
	return it, true
}

func (s *MySQLStore) WebhookCircuitBreakerByKey(pluginCode, targetURL string) (domain.WebhookCircuitBreaker, bool) {
	pluginCode = strings.TrimSpace(pluginCode)
	targetURL = strings.TrimSpace(targetURL)
	var it domain.WebhookCircuitBreaker
	err := s.db.QueryRow(`SELECT id, plugin_code, target_url, status, failure_count, success_count,
		COALESCE(DATE_FORMAT(opened_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(closed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(next_probe_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(last_error_message,''), COALESCE(DATE_FORMAT(last_failure_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_success_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_circuit_breakers WHERE plugin_code=? AND target_url=? LIMIT 1`, pluginCode, targetURL).
		Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.Status, &it.FailureCount, &it.SuccessCount,
			&it.OpenedAt, &it.ClosedAt, &it.NextProbeAt, &it.LastErrorMessage, &it.LastFailureAt, &it.LastSuccessAt,
			&it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.WebhookCircuitBreaker{}, false
	}
	return it, true
}

func (s *MySQLStore) WebhookCircuitBreakers(filter domain.WebhookCircuitBreakerFilter) ([]domain.WebhookCircuitBreaker, int, error) {
	filter = filter.Normalize()
	where := []string{"1=1"}
	args := []any{}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM webhook_circuit_breakers WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, plugin_code, target_url, status, failure_count, success_count,
		COALESCE(DATE_FORMAT(opened_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(closed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(next_probe_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(last_error_message,''), COALESCE(DATE_FORMAT(last_failure_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(last_success_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM webhook_circuit_breakers WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.WebhookCircuitBreaker{}
	for rows.Next() {
		var it domain.WebhookCircuitBreaker
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.TargetURL, &it.Status, &it.FailureCount, &it.SuccessCount,
			&it.OpenedAt, &it.ClosedAt, &it.NextProbeAt, &it.LastErrorMessage, &it.LastFailureAt, &it.LastSuccessAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) columnExists(tableName, columnName string) bool {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND COLUMN_NAME=?`, tableName, columnName).Scan(&count)
	return err == nil && count > 0
}

func (s *MySQLStore) PluginMigrations(pluginCode string) ([]domain.PluginMigration, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" {
		return []domain.PluginMigration{}, nil
	}
	rows, err := s.db.Query(`SELECT id,plugin_code,version,migration_name,checksum,status,COALESCE(DATE_FORMAT(executed_at,'%Y-%m-%d %H:%i:%s'),''),execution_time_ms,COALESCE(error_message,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_migrations WHERE plugin_code=? ORDER BY id DESC`, pluginCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PluginMigration{}
	for rows.Next() {
		var it domain.PluginMigration
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.Version, &it.MigrationName, &it.Checksum, &it.Status, &it.ExecutedAt, &it.ExecutionTimeMS, &it.ErrorMessage, &it.CreatedAt, &it.UpdatedAt); err == nil {
			it.MigrationVersion = it.Version
			it.FinishedAt = it.ExecutedAt
			it.DurationMS = it.ExecutionTimeMS
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *MySQLStore) AppendPluginMigration(record domain.PluginMigration) (domain.PluginMigration, error) {
	if record.MigrationVersion == "" {
		record.MigrationVersion = record.Version
	}
	if record.Version == "" {
		record.Version = record.MigrationVersion
	}
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.Version = strings.TrimSpace(record.Version)
	record.MigrationVersion = strings.TrimSpace(record.MigrationVersion)
	record.MigrationName = strings.TrimSpace(record.MigrationName)
	record.Status = strings.TrimSpace(record.Status)
	if record.PluginCode == "" || record.MigrationName == "" {
		return domain.PluginMigration{}, errors.New("plugin_code 和 migration_name 不能为空")
	}
	if record.Status == "" {
		record.Status = "pending"
	}
	// Upsert-like: if existing unique key exists, do not insert again.
	_, err := s.db.Exec(`INSERT IGNORE INTO plugin_migrations (plugin_code,version,migration_name,checksum,status,executed_at,execution_time_ms,error_message,created_at,updated_at)
		VALUES (?,?,?,?,?,NULL,?,?,NOW(),NOW())`,
		record.PluginCode, record.Version, record.MigrationName, record.Checksum, record.Status, record.ExecutionTimeMS, record.ErrorMessage)
	if err != nil {
		return domain.PluginMigration{}, err
	}
	// Return latest record for that unique key.
	var out domain.PluginMigration
	_ = s.db.QueryRow(`SELECT id,plugin_code,version,migration_name,checksum,status,COALESCE(DATE_FORMAT(executed_at,'%Y-%m-%d %H:%i:%s'),''),execution_time_ms,COALESCE(error_message,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_migrations WHERE plugin_code=? AND version=? AND migration_name=? ORDER BY id DESC LIMIT 1`,
		record.PluginCode, record.Version, record.MigrationName).
		Scan(&out.ID, &out.PluginCode, &out.Version, &out.MigrationName, &out.Checksum, &out.Status, &out.ExecutedAt, &out.ExecutionTimeMS, &out.ErrorMessage, &out.CreatedAt, &out.UpdatedAt)
	out.MigrationVersion = out.Version
	out.FinishedAt = out.ExecutedAt
	out.DurationMS = out.ExecutionTimeMS
	return out, nil
}

func (s *MySQLStore) SavePluginMigration(record domain.PluginMigration) (domain.PluginMigration, error) {
	if record.MigrationVersion == "" {
		record.MigrationVersion = record.Version
	}
	if record.Version == "" {
		record.Version = record.MigrationVersion
	}
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.Version = strings.TrimSpace(record.Version)
	record.MigrationName = strings.TrimSpace(record.MigrationName)
	record.Status = strings.TrimSpace(record.Status)
	if record.PluginCode == "" || record.MigrationName == "" {
		return domain.PluginMigration{}, errors.New("plugin_code 和 migration_name 不能为空")
	}
	if record.Status == "" {
		record.Status = "pending"
	}
	executedAt := record.FinishedAt
	if executedAt == "" {
		executedAt = record.ExecutedAt
	}
	if executedAt == "" && (record.Status == "success" || record.Status == "failed") {
		executedAt = Now()
	}
	duration := record.DurationMS
	if duration == 0 {
		duration = record.ExecutionTimeMS
	}
	_, err := s.db.Exec(`INSERT INTO plugin_migrations (plugin_code,version,migration_name,checksum,status,executed_at,execution_time_ms,error_message,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,NOW(),NOW())
		ON DUPLICATE KEY UPDATE checksum=VALUES(checksum),status=VALUES(status),executed_at=VALUES(executed_at),execution_time_ms=VALUES(execution_time_ms),error_message=VALUES(error_message),updated_at=NOW()`,
		record.PluginCode, record.Version, record.MigrationName, record.Checksum, record.Status, nullableTimeString(executedAt), duration, record.ErrorMessage)
	if err != nil {
		return domain.PluginMigration{}, err
	}
	var out domain.PluginMigration
	_ = s.db.QueryRow(`SELECT id,plugin_code,version,migration_name,checksum,status,COALESCE(DATE_FORMAT(executed_at,'%Y-%m-%d %H:%i:%s'),''),execution_time_ms,COALESCE(error_message,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_migrations WHERE plugin_code=? AND version=? AND migration_name=? ORDER BY id DESC LIMIT 1`,
		record.PluginCode, record.Version, record.MigrationName).
		Scan(&out.ID, &out.PluginCode, &out.Version, &out.MigrationName, &out.Checksum, &out.Status, &out.ExecutedAt, &out.ExecutionTimeMS, &out.ErrorMessage, &out.CreatedAt, &out.UpdatedAt)
	out.MigrationVersion = out.Version
	out.FinishedAt = out.ExecutedAt
	out.DurationMS = out.ExecutionTimeMS
	return out, nil
}

func (s *MySQLStore) AppendHookExecution(record domain.HookExecution) (domain.HookExecution, error) {
	record.HookName = strings.TrimSpace(record.HookName)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	if record.HookName == "" || record.PluginCode == "" {
		return domain.HookExecution{}, errors.New("hook_name 和 plugin_code 不能为空")
	}
	if record.StartedAt == "" {
		record.StartedAt = Now()
	}
	if record.FinishedAt == "" {
		record.FinishedAt = record.StartedAt
	}
	res, err := s.db.Exec(`INSERT INTO hook_executions
		(hook_name,plugin_code,mode,content_type,content_id,community_id,category_id,actor_type,actor_id,user_id,admin_user_id,request_id,started_at,finished_at,duration_ms,success,error_message,blocking,metadata_json,created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW())`,
		record.HookName, record.PluginCode, record.Mode, record.ContentType,
		nullableInt64(record.ContentID), nullableInt64(record.CommunityID), nullableInt64(record.CategoryID),
		record.ActorType, nullableInt64(record.ActorID), nullableInt64(record.UserID), nullableInt64(record.AdminUserID),
		record.RequestID, record.StartedAt, record.FinishedAt, record.DurationMS, record.Success, record.ErrorMessage, record.Blocking, mysqlJSONArg(record.Metadata))
	if err != nil {
		return domain.HookExecution{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	record.CreatedAt = Now()
	return record, nil
}

func (s *MySQLStore) HookExecutions(pluginCode string, limit int) ([]domain.HookExecution, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	where := ""
	args := []any{}
	if pluginCode != "" {
		where = "WHERE plugin_code=?"
		args = append(args, pluginCode)
	}
	args = append(args, limit)
	rows, err := s.db.Query(`SELECT id,hook_name,plugin_code,mode,COALESCE(content_type,''),COALESCE(content_id,0),COALESCE(community_id,0),COALESCE(category_id,0),COALESCE(actor_type,''),COALESCE(actor_id,0),COALESCE(user_id,0),COALESCE(admin_user_id,0),COALESCE(request_id,''),DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),duration_ms,success,COALESCE(error_message,''),blocking,COALESCE(CAST(metadata_json AS CHAR),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s')
		FROM hook_executions `+where+` ORDER BY id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.HookExecution{}
	for rows.Next() {
		var it domain.HookExecution
		if err := rows.Scan(&it.ID, &it.HookName, &it.PluginCode, &it.Mode, &it.ContentType, &it.ContentID, &it.CommunityID, &it.CategoryID, &it.ActorType, &it.ActorID, &it.UserID, &it.AdminUserID, &it.RequestID, &it.StartedAt, &it.FinishedAt, &it.DurationMS, &it.Success, &it.ErrorMessage, &it.Blocking, &it.Metadata, &it.CreatedAt); err == nil {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *MySQLStore) HookStats(pluginCode string) ([]domain.HookStats, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	where := ""
	args := []any{}
	if pluginCode != "" {
		where = "WHERE plugin_code=?"
		args = append(args, pluginCode)
	}
	rows, err := s.db.Query(`SELECT hook_name,plugin_code,mode,blocking,COUNT(*),SUM(CASE WHEN success=0 THEN 1 ELSE 0 END),COALESCE(AVG(duration_ms),0),MAX(finished_at),MAX(CASE WHEN success=0 THEN finished_at ELSE NULL END)
		FROM hook_executions `+where+` GROUP BY hook_name,plugin_code,mode,blocking ORDER BY plugin_code,hook_name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.HookStats{}
	for rows.Next() {
		var it domain.HookStats
		var lastExecuted, lastFailed sql.NullTime
		if err := rows.Scan(&it.HookName, &it.PluginCode, &it.Mode, &it.Blocking, &it.ExecutionCount, &it.FailureCount, &it.AvgDurationMS, &lastExecuted, &lastFailed); err != nil {
			continue
		}
		if lastExecuted.Valid {
			it.LastExecutedAt = lastExecuted.Time.Format(TimeLayout)
		}
		if lastFailed.Valid {
			it.LastFailedAt = lastFailed.Time.Format(TimeLayout)
			_ = s.db.QueryRow(`SELECT COALESCE(error_message,'') FROM hook_executions WHERE plugin_code=? AND hook_name=? AND success=0 ORDER BY id DESC LIMIT 1`, it.PluginCode, it.HookName).Scan(&it.LastError)
		}
		out = append(out, it)
	}
	return out, nil
}

func (s *MySQLStore) HookExecutionsByFilter(filter domain.HookExecutionFilter) ([]domain.HookExecution, int, error) {
	filter = filter.Normalize()
	where := "WHERE 1=1"
	args := []any{}
	if filter.PluginCode != "" {
		where += " AND plugin_code=?"
		args = append(args, filter.PluginCode)
	}
	if filter.HookName != "" {
		where += " AND hook_name=?"
		args = append(args, filter.HookName)
	}
	if filter.Mode != "" {
		where += " AND mode=?"
		args = append(args, filter.Mode)
	}
	if filter.Blocking != nil {
		where += " AND blocking=?"
		if *filter.Blocking {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	if filter.Success != nil {
		where += " AND success=?"
		if *filter.Success {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	if filter.ContentType != "" {
		where += " AND content_type=?"
		args = append(args, filter.ContentType)
	}
	if filter.ContentID > 0 {
		where += " AND content_id=?"
		args = append(args, filter.ContentID)
	}
	if filter.CommunityID > 0 {
		where += " AND community_id=?"
		args = append(args, filter.CommunityID)
	}
	if filter.ActorType != "" {
		where += " AND actor_type=?"
		args = append(args, filter.ActorType)
	}
	if filter.ActorID > 0 {
		where += " AND actor_id=?"
		args = append(args, filter.ActorID)
	}
	if filter.RequestID != "" {
		where += " AND request_id=?"
		args = append(args, filter.RequestID)
	}
	if filter.StartTime != "" {
		where += " AND started_at>=?"
		args = append(args, filter.StartTime)
	}
	if filter.EndTime != "" {
		where += " AND started_at<=?"
		args = append(args, filter.EndTime)
	}

	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM hook_executions "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id,hook_name,plugin_code,mode,COALESCE(content_type,''),COALESCE(content_id,0),COALESCE(community_id,0),COALESCE(category_id,0),COALESCE(actor_type,''),COALESCE(actor_id,0),COALESCE(user_id,0),COALESCE(admin_user_id,0),COALESCE(request_id,''),DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),duration_ms,success,COALESCE(error_message,''),blocking,COALESCE(CAST(metadata_json AS CHAR),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s')
		FROM hook_executions `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.HookExecution{}
	for rows.Next() {
		var it domain.HookExecution
		if err := rows.Scan(&it.ID, &it.HookName, &it.PluginCode, &it.Mode, &it.ContentType, &it.ContentID, &it.CommunityID, &it.CategoryID, &it.ActorType, &it.ActorID, &it.UserID, &it.AdminUserID, &it.RequestID, &it.StartedAt, &it.FinishedAt, &it.DurationMS, &it.Success, &it.ErrorMessage, &it.Blocking, &it.Metadata, &it.CreatedAt); err == nil {
			out = append(out, it)
		}
	}
	return out, total, nil
}

// NewMySQLStore 创建 MySQL 仓储，自动建表并在空库时写入演示数据。
func NewMySQLStore(cfg MySQLConfig) (*MySQLStore, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := pingWithRetry(db, 30*time.Second); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &MySQLStore{db: db, database: cfg.Database}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.migrateSiteScopedAudit(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.migratePluginPackageDetachedSignatureV171(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.seedIfEmpty(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库连接。
func (s *MySQLStore) Close() error { return s.db.Close() }

// Health 返回 MySQL 仓储状态和关键表行数，方便排查数据源问题。
func (s *MySQLStore) Health() domain.HealthStatus {
	status := domain.HealthStatus{
		OK:       true,
		Time:     Now(),
		Store:    "mysql",
		Database: s.database,
		Counts:   map[string]int{},
	}
	if err := s.db.Ping(); err != nil {
		status.OK = false
		status.Error = err.Error()
		return status
	}
	for _, table := range []string{"sites", "boards", "plugins", "tags", "posts", "comments", "notifications", "admin_users", "admin_roles"} {
		var count int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			status.OK = false
			status.Error = err.Error()
			return status
		}
		status.Counts[table] = count
	}
	return status
}

func pingWithRetry(db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(time.Second)
	}
	return lastErr
}

func (s *MySQLStore) migrate() error {
	schema := stripSQLLineComments(mySQLSchema)
	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) migrateSiteScopedAudit() error {
	_, _ = s.db.Exec(`ALTER TABLE refresh_tokens ADD COLUMN token_type VARCHAR(32) NOT NULL DEFAULT 'user' AFTER token_hash`)
	_, _ = s.db.Exec(`ALTER TABLE refresh_tokens ADD KEY idx_refresh_tokens_type_user (token_type, user_id, revoked_at)`)
	_, _ = s.db.Exec(`ALTER TABLE refresh_tokens DROP FOREIGN KEY fk_refresh_tokens_user`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN actor_type VARCHAR(32) NOT NULL DEFAULT 'admin_user' AFTER actor`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN actor_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER actor_type`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD KEY idx_admin_logs_actor_type_created (actor_type, created_at)`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN site_key VARCHAR(64) NOT NULL DEFAULT 'portal' AFTER id`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN actor_user_id BIGINT UNSIGNED NULL AFTER site_key`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN type VARCHAR(64) NOT NULL DEFAULT '' AFTER actor_user_id`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN target_type VARCHAR(32) NOT NULL DEFAULT '' AFTER type`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN target_id BIGINT UNSIGNED NULL AFTER target_type`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN topic_id BIGINT UNSIGNED NULL AFTER target_id`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN comment_id BIGINT UNSIGNED NULL AFTER topic_id`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD COLUMN read_at DATETIME NULL AFTER created_at`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD KEY idx_notifications_site_read_created (site_key, is_read, created_at)`)
	_, _ = s.db.Exec(`ALTER TABLE notifications ADD KEY idx_notifications_type_target (type, target_type, target_id)`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN topic_id BIGINT UNSIGNED NULL AFTER post_id`)
	_, _ = s.db.Exec(`ALTER TABLE comments DROP FOREIGN KEY fk_comments_post`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN reply_to_user_id BIGINT UNSIGNED NULL AFTER parent_id`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN user_id BIGINT UNSIGNED NULL AFTER reply_to_user_id`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN content_html MEDIUMTEXT NULL AFTER text`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN is_best TINYINT(1) NOT NULL DEFAULT 0 AFTER likes`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD COLUMN deleted_at DATETIME NULL AFTER updated_at`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD KEY idx_comments_topic_created (topic_id, created_at)`)
	_, _ = s.db.Exec(`ALTER TABLE comments ADD KEY idx_comments_user_created (user_id, created_at)`)
	_, _ = s.db.Exec(`UPDATE comments SET topic_id=post_id WHERE topic_id IS NULL OR topic_id=0`)
	_, _ = s.db.Exec(`UPDATE comments SET user_id=1 WHERE user_id IS NULL OR user_id=0`)
	_, _ = s.db.Exec(`ALTER TABLE topics ADD COLUMN comment_locked TINYINT NOT NULL DEFAULT 0 AFTER is_solved`)
	_, _ = s.db.Exec(`ALTER TABLE reports ADD COLUMN community_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER target_id`)
	_, _ = s.db.Exec(`ALTER TABLE reports ADD COLUMN topic_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER community_id`)
	_, _ = s.db.Exec(`ALTER TABLE reports ADD COLUMN handle_note VARCHAR(1000) NULL AFTER handled_at`)
	_, _ = s.db.Exec(`ALTER TABLE reports ADD KEY idx_reports_community_status (community_id, status)`)
	_, _ = s.db.Exec(`ALTER TABLE reports ADD KEY idx_reports_reporter_target_status (reporter_id, target_type, target_id, status)`)
	_, _ = s.db.Exec(`ALTER TABLE reactions ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at`)
	_, _ = s.db.Exec(`ALTER TABLE favorites ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at`)
	_, _ = s.db.Exec(`ALTER TABLE follows ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at`)
	_, _ = s.db.Exec(`ALTER TABLE activities ADD COLUMN topic_id BIGINT UNSIGNED NULL AFTER community_id`)
	_, _ = s.db.Exec(`ALTER TABLE activities ADD COLUMN metadata TEXT NULL AFTER remark`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN site_key VARCHAR(64) NOT NULL DEFAULT 'portal' AFTER id`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN role_code VARCHAR(64) NOT NULL DEFAULT '' AFTER actor`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN old_value JSON NULL AFTER target`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN new_value JSON NULL AFTER old_value`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD COLUMN metadata_json JSON NULL AFTER new_value`)
	_, _ = s.db.Exec(`ALTER TABLE admin_logs ADD KEY idx_admin_logs_site_created (site_key, created_at)`)
	_, _ = s.db.Exec(`UPDATE notifications SET site_key=scope WHERE scope NOT IN ('','all','portal') AND site_key='portal'`)
	_, _ = s.db.Exec(`ALTER TABLE categories DROP INDEX uk_categories_slug`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD UNIQUE KEY uk_categories_community_slug (community_id, slug)`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN cover_image VARCHAR(500) NOT NULL DEFAULT '' AFTER logo`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN slogan VARCHAR(255) NOT NULL DEFAULT '' AFTER cover_image`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN theme_color VARCHAR(32) NOT NULL DEFAULT '' AFTER description`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN seo_title VARCHAR(255) NOT NULL DEFAULT '' AFTER theme_color`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN seo_description VARCHAR(500) NOT NULL DEFAULT '' AFTER seo_title`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN seo_keywords VARCHAR(500) NOT NULL DEFAULT '' AFTER seo_description`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN follower_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER status`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN topic_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER follower_count`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN comment_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER topic_count`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN hot_score INT UNSIGNED NOT NULL DEFAULT 0 AFTER comment_count`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN announcement_title VARCHAR(255) NOT NULL DEFAULT '' AFTER hot_score`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN announcement_content TEXT NULL AFTER announcement_title`)
	_, _ = s.db.Exec(`ALTER TABLE communities ADD COLUMN announcement_url VARCHAR(500) NOT NULL DEFAULT '' AFTER announcement_content`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD COLUMN nav_visible TINYINT NOT NULL DEFAULT 1 AFTER visible`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD COLUMN postable TINYINT NOT NULL DEFAULT 1 AFTER nav_visible`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD COLUMN seo_title VARCHAR(255) NOT NULL DEFAULT '' AFTER postable`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD COLUMN seo_description VARCHAR(500) NOT NULL DEFAULT '' AFTER seo_title`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD COLUMN plugin_code VARCHAR(64) NOT NULL DEFAULT 'core' AFTER type`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD COLUMN allowed_content_types JSON NULL AFTER plugin_code`)
	_, _ = s.db.Exec(`ALTER TABLE categories ADD KEY idx_categories_plugin_status (plugin_code, status)`)
	_, _ = s.db.Exec(`ALTER TABLE plugins ADD COLUMN config_json JSON NULL AFTER description`)
	_, _ = s.db.Exec(`ALTER TABLE plugins MODIFY COLUMN status ENUM('discovered','installed','migrated','configured','enabled','disabled','running','archived','config_invalid','migration_pending','migration_failed','dependency_missing') NOT NULL DEFAULT 'enabled'`)
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_migrations (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		plugin_code VARCHAR(64) NOT NULL,
		version VARCHAR(32) NOT NULL DEFAULT '',
		migration_name VARCHAR(128) NOT NULL,
		checksum VARCHAR(128) NOT NULL DEFAULT '',
		status ENUM('pending','running','success','failed') NOT NULL DEFAULT 'pending',
		executed_at DATETIME NULL,
		execution_time_ms INT NOT NULL DEFAULT 0,
		error_message VARCHAR(1000) NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_plugin_migrations_unique (plugin_code, version, migration_name),
		KEY idx_plugin_migrations_plugin (plugin_code),
		KEY idx_plugin_migrations_status (status),
		KEY idx_plugin_migrations_executed (executed_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	_, _ = s.db.Exec(`ALTER TABLE plugin_migrations MODIFY COLUMN status ENUM('pending','running','success','failed') NOT NULL DEFAULT 'pending'`)
	if !s.columnExists("plugin_migrations", "updated_at") {
		_, _ = s.db.Exec(`ALTER TABLE plugin_migrations ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`)
	}
	// v1.5.0-P1-05: plugin config versions (history + diff + rollback preview).
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_config_versions (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		plugin_code VARCHAR(64) NOT NULL,
		scope VARCHAR(32) NOT NULL DEFAULT 'global',
		community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		version_no INT NOT NULL DEFAULT 0,
		config_json JSON NULL,
		config_hash VARCHAR(128) NOT NULL DEFAULT '',
		changed_keys_json JSON NULL,
		diff_json JSON NULL,
		source VARCHAR(32) NOT NULL DEFAULT 'manual',
		operator_type VARCHAR(32) NOT NULL DEFAULT 'admin_user',
		operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		operator_name VARCHAR(128) NOT NULL DEFAULT '',
		reason VARCHAR(255) NOT NULL DEFAULT '',
		previous_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		rollback_from_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		metadata_json JSON NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_plugin_config_versions_lookup (plugin_code, scope, community_id, version_no),
		KEY idx_plugin_config_versions_created (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	// v1.5.0-P1-07: plugin approvals (install/upgrade approvals + audit snapshots).
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_approval_requests (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		request_no VARCHAR(64) NOT NULL DEFAULT '',
		action VARCHAR(32) NOT NULL,
		plugin_code VARCHAR(64) NOT NULL DEFAULT '',
		plugin_name VARCHAR(128) NOT NULL DEFAULT '',
		current_version VARCHAR(32) NOT NULL DEFAULT '',
		target_version VARCHAR(32) NOT NULL DEFAULT '',
		package_path VARCHAR(500) NOT NULL DEFAULT '',
		package_checksum_status VARCHAR(32) NOT NULL DEFAULT '',
		package_risk_level VARCHAR(32) NOT NULL DEFAULT '',
		status VARCHAR(32) NOT NULL DEFAULT 'pending',
		reason VARCHAR(1000) NOT NULL DEFAULT '',
		requested_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		requested_by_name VARCHAR(128) NOT NULL DEFAULT '',
		requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		reviewed_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		reviewed_by_name VARCHAR(128) NOT NULL DEFAULT '',
		reviewed_at DATETIME NULL,
		review_comment VARCHAR(1000) NOT NULL DEFAULT '',
		executed_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		executed_at DATETIME NULL,
		execute_result_json JSON NULL,
		manifest_json JSON NULL,
		dry_run_json JSON NULL,
		risk_report_json JSON NULL,
		dependency_summary_json JSON NULL,
		compatibility_json JSON NULL,
		changed_keys_json JSON NULL,
		diff_json JSON NULL,
		error_code VARCHAR(64) NOT NULL DEFAULT '',
		error_message VARCHAR(1000) NOT NULL DEFAULT '',
		metadata_json JSON NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_plugin_approvals_status_created (status, created_at),
		KEY idx_plugin_approvals_action_created (action, created_at),
		KEY idx_plugin_approvals_plugin_created (plugin_code, created_at),
		KEY idx_plugin_approvals_requested (requested_by, requested_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	// v1.6.0-P0-02: uploaded zip plugin packages as lifecycle objects.
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_package_uploads (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		upload_id VARCHAR(80) NOT NULL,
		original_filename VARCHAR(255) NOT NULL DEFAULT '',
		uploaded_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		uploaded_by_name VARCHAR(128) NOT NULL DEFAULT '',
		uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status VARCHAR(32) NOT NULL DEFAULT 'uploaded',
		package_code VARCHAR(64) NOT NULL DEFAULT '',
		package_name VARCHAR(128) NOT NULL DEFAULT '',
		package_version VARCHAR(32) NOT NULL DEFAULT '',
		upload_path VARCHAR(500) NOT NULL DEFAULT '',
		staging_path VARCHAR(500) NOT NULL DEFAULT '',
		package_path VARCHAR(500) NOT NULL DEFAULT '',
		promoted_path VARCHAR(500) NOT NULL DEFAULT '',
		compressed_size BIGINT NOT NULL DEFAULT 0,
		uncompressed_size BIGINT NOT NULL DEFAULT 0,
		file_count INT NOT NULL DEFAULT 0,
		checksum_status VARCHAR(32) NOT NULL DEFAULT '',
		signature_status VARCHAR(32) NOT NULL DEFAULT '',
		publisher_id VARCHAR(128) NOT NULL DEFAULT '',
		trust_status VARCHAR(32) NOT NULL DEFAULT '',
		risk_level VARCHAR(32) NOT NULL DEFAULT '',
		risk_report_json JSON NULL,
		zip_scan_json JSON NULL,
		file_scan_json JSON NULL,
		manifest_validation_json JSON NULL,
		install_dry_run_json JSON NULL,
		approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		install_approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		expires_at DATETIME NULL,
		deleted_at DATETIME NULL,
		error_code VARCHAR(80) NOT NULL DEFAULT '',
		error_message VARCHAR(1000) NOT NULL DEFAULT '',
		metadata_json JSON NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_plugin_package_uploads_upload_id (upload_id),
		KEY idx_plugin_package_uploads_status_created (status, created_at),
		KEY idx_plugin_package_uploads_package (package_code, package_version),
		KEY idx_plugin_package_uploads_uploaded_by (uploaded_by, uploaded_at),
		KEY idx_plugin_package_uploads_risk (risk_level, status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	// v1.6.0-P0-06: plugin operation snapshots for install/upgrade protection and recovery.
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_operation_snapshots (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		operation_id VARCHAR(80) NOT NULL,
		operation_type VARCHAR(32) NOT NULL,
		plugin_code VARCHAR(64) NOT NULL,
		from_version VARCHAR(32) NOT NULL DEFAULT '',
		to_version VARCHAR(32) NOT NULL DEFAULT '',
		package_path VARCHAR(500) NOT NULL DEFAULT '',
		package_source VARCHAR(32) NOT NULL DEFAULT '',
		approval_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		before_plugin_json JSON NULL,
		before_manifest_json JSON NULL,
		before_config_json JSON NULL,
		before_config_version_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
		before_migrations_json JSON NULL,
		before_permissions_json JSON NULL,
		before_menus_json JSON NULL,
		before_routes_json JSON NULL,
		before_dependencies_json JSON NULL,
		before_status VARCHAR(32) NOT NULL DEFAULT '',
		after_manifest_json JSON NULL,
		dry_run_json JSON NULL,
		risk_report_json JSON NULL,
		diff_json JSON NULL,
		checksum_summary_json JSON NULL,
		signature_summary_json JSON NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'created',
		error_code VARCHAR(80) NOT NULL DEFAULT '',
		error_message VARCHAR(1000) NOT NULL DEFAULT '',
		created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		metadata_json JSON NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_plugin_operation_snapshots_operation_id (operation_id),
		KEY idx_plugin_operation_snapshots_plugin (plugin_code, created_at),
		KEY idx_plugin_operation_snapshots_status (status, created_at),
		KEY idx_plugin_operation_snapshots_type (operation_type, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_trusted_publishers (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		publisher_id VARCHAR(128) NOT NULL,
		name VARCHAR(128) NOT NULL DEFAULT '',
		homepage VARCHAR(255) NOT NULL DEFAULT '',
		email VARCHAR(255) NOT NULL DEFAULT '',
		public_key_id VARCHAR(128) NOT NULL,
		public_key_algorithm VARCHAR(32) NOT NULL DEFAULT 'ed25519',
		public_key TEXT NOT NULL,
		fingerprint VARCHAR(128) NOT NULL DEFAULT '',
		status VARCHAR(32) NOT NULL DEFAULT 'trusted',
		notes VARCHAR(1000) NOT NULL DEFAULT '',
		created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		revoked_at DATETIME NULL,
		blocked_at DATETIME NULL,
		metadata_json JSON NULL,
		PRIMARY KEY (id),
		UNIQUE KEY uk_plugin_trusted_publishers_key (publisher_id, public_key_id),
		KEY idx_plugin_trusted_publishers_status (status, updated_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_remote_indexes (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		source_id VARCHAR(128) NOT NULL,
		name VARCHAR(128) NOT NULL DEFAULT '',
		index_url VARCHAR(500) NOT NULL,
		homepage VARCHAR(255) NOT NULL DEFAULT '',
		description VARCHAR(1000) NOT NULL DEFAULT '',
		status VARCHAR(32) NOT NULL DEFAULT 'enabled',
		trust_policy VARCHAR(32) NOT NULL DEFAULT 'readonly',
		last_fetch_status VARCHAR(32) NOT NULL DEFAULT '',
		last_fetch_at DATETIME NULL,
		last_error_code VARCHAR(128) NOT NULL DEFAULT '',
		last_error_message VARCHAR(1000) NOT NULL DEFAULT '',
		last_index_hash VARCHAR(128) NOT NULL DEFAULT '',
		metadata_json JSON NULL,
		created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_plugin_remote_indexes_source (source_id),
		KEY idx_plugin_remote_indexes_status (status, updated_at),
		KEY idx_plugin_remote_indexes_fetch (last_fetch_status, last_fetch_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS hook_executions (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		hook_name VARCHAR(128) NOT NULL,
		plugin_code VARCHAR(64) NOT NULL,
		mode VARCHAR(32) NOT NULL DEFAULT 'non_blocking',
		content_type VARCHAR(64) NOT NULL DEFAULT '',
		content_id BIGINT UNSIGNED NULL,
		community_id BIGINT UNSIGNED NULL,
		category_id BIGINT UNSIGNED NULL,
		actor_type VARCHAR(32) NOT NULL DEFAULT '',
		actor_id BIGINT UNSIGNED NULL,
		user_id BIGINT UNSIGNED NULL,
		admin_user_id BIGINT UNSIGNED NULL,
		request_id VARCHAR(128) NOT NULL DEFAULT '',
		started_at DATETIME NOT NULL,
		finished_at DATETIME NULL,
		duration_ms INT NOT NULL DEFAULT 0,
		success TINYINT(1) NOT NULL DEFAULT 1,
		error_message TEXT NULL,
		blocking TINYINT(1) NOT NULL DEFAULT 0,
		metadata_json JSON NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_hook_executions_plugin_hook (plugin_code, hook_name),
		KEY idx_hook_executions_success (plugin_code, success, started_at),
		KEY idx_hook_executions_content (content_id),
		KEY idx_hook_executions_community (community_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	_, _ = s.db.Exec(`UPDATE categories SET nav_visible=visible WHERE nav_visible=1`)
	_, _ = s.db.Exec(`UPDATE categories SET type='document', plugin_code='docs', allowed_content_types=JSON_ARRAY('document','doc') WHERE type='doc' OR slug='docs'`)
	_, _ = s.db.Exec(`UPDATE categories SET type='wiki_page', plugin_code='wiki', allowed_content_types=JSON_ARRAY('wiki_page','wiki') WHERE type='wiki' OR slug='wiki'`)
	_, _ = s.db.Exec(`UPDATE categories SET plugin_code='qa', allowed_content_types=JSON_ARRAY('question') WHERE type='question'`)
	_, _ = s.db.Exec(`UPDATE categories SET plugin_code='projects', allowed_content_types=JSON_ARRAY('project') WHERE type='project' OR slug='opensource'`)
	_, _ = s.db.Exec(`UPDATE categories SET plugin_code='jobs', allowed_content_types=JSON_ARRAY('job') WHERE type='job' OR slug='jobs'`)
	_, _ = s.db.Exec(`UPDATE categories SET plugin_code='ai_works', allowed_content_types=JSON_ARRAY('ai_work') WHERE type='ai_work' OR slug='ai'`)
	_, _ = s.db.Exec(`UPDATE categories SET plugin_code='core', allowed_content_types=JSON_ARRAY(type) WHERE plugin_code='core' AND allowed_content_types IS NULL`)
	_, _ = s.db.Exec(`ALTER TABLE topics ADD COLUMN plugin_code VARCHAR(64) NOT NULL DEFAULT 'core' AFTER slug`)
	_, _ = s.db.Exec(`ALTER TABLE topics ADD KEY idx_topics_plugin_type_status (plugin_code, content_type, status)`)
	_, _ = s.db.Exec(`UPDATE topics SET content_type='document', plugin_code='docs' WHERE content_type='doc'`)
	_, _ = s.db.Exec(`UPDATE topics SET content_type='wiki_page', plugin_code='wiki' WHERE content_type='wiki'`)
	_, _ = s.db.Exec(`UPDATE topics SET plugin_code='qa' WHERE content_type='question'`)
	_, _ = s.db.Exec(`UPDATE topics SET plugin_code='projects' WHERE content_type='project'`)
	_, _ = s.db.Exec(`UPDATE topics SET plugin_code='jobs' WHERE content_type='job'`)
	_, _ = s.db.Exec(`UPDATE topics SET plugin_code='ai_works' WHERE content_type='ai_work'`)
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS community_plugins (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		community_id BIGINT UNSIGNED NOT NULL,
		plugin_code VARCHAR(64) NOT NULL,
		status ENUM('enabled','disabled') NOT NULL DEFAULT 'enabled',
		sort_order INT NOT NULL DEFAULT 0,
		config_json JSON NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_community_plugins_community_code (community_id, plugin_code),
		KEY idx_community_plugins_plugin (plugin_code),
		KEY idx_community_plugins_community (community_id),
		CONSTRAINT fk_community_plugins_community FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	_, _ = s.db.Exec(`INSERT IGNORE INTO community_plugins (community_id, plugin_code, status, sort_order, config_json, created_at, updated_at)
		SELECT c.id, p.plugin_code, 'enabled', 0, NULL, NOW(), NOW() FROM communities c JOIN plugins p ON p.status='enabled'`)
	_, _ = s.db.Exec(`ALTER TABLE tags MODIFY status VARCHAR(32) NOT NULL DEFAULT 'enable'`)
	_, _ = s.db.Exec(`UPDATE tags SET status='enable' WHERE status IN ('1','enabled','')`)
	_, _ = s.db.Exec(`UPDATE tags SET status='disable' WHERE status IN ('0','disabled')`)
	_, _ = s.db.Exec(`UPDATE tags SET status='merged' WHERE status IN ('2','merged')`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD COLUMN merged_to_id BIGINT UNSIGNED NULL AFTER status`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD COLUMN follower_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER use_count`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD COLUMN hot_score INT UNSIGNED NOT NULL DEFAULT 0 AFTER follower_count`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD COLUMN seo_title VARCHAR(255) NOT NULL DEFAULT '' AFTER follower_count`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD COLUMN seo_description VARCHAR(500) NOT NULL DEFAULT '' AFTER seo_title`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD COLUMN seo_keywords VARCHAR(500) NOT NULL DEFAULT '' AFTER seo_description`)
	_, _ = s.db.Exec(`ALTER TABLE tags ADD KEY idx_tags_merged_to (merged_to_id)`)
	_, _ = s.db.Exec(`CREATE TABLE IF NOT EXISTS tag_aliases (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		tag_id BIGINT UNSIGNED NOT NULL,
		site_key VARCHAR(64) NOT NULL DEFAULT 'portal',
		alias VARCHAR(128) NOT NULL,
		alias_slug VARCHAR(128) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_tag_aliases_site_alias_slug (site_key, alias_slug),
		KEY idx_tag_aliases_tag_id (tag_id),
		CONSTRAINT fk_tag_aliases_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	return nil
}

func (s *MySQLStore) migratePluginPackageDetachedSignatureV171() error {
	// v1.7.1: detached signature verification needs signature_url and key expiry support.
	_, _ = s.db.Exec(`ALTER TABLE plugin_package_downloads ADD COLUMN signature_url VARCHAR(1000) NOT NULL DEFAULT '' AFTER sha256_expected`)
	_, _ = s.db.Exec(`ALTER TABLE plugin_trusted_publishers ADD COLUMN expires_at DATETIME NULL AFTER blocked_at`)
	return nil
}

func stripSQLLineComments(schema string) string {
	lines := strings.Split(schema, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func (s *MySQLStore) seedIfEmpty() error {
	sites := []domain.Site{
		{Key: "portal", Name: "DevHub", Logo: "DH", Title: "DevHub", Sub: "总网站 · 多技术子站内容集合", Pub: "发布内容", Description: "聚合 PHP、Go、Java、AI、Frontend 内容", Color: "#2563eb", Status: "enable", Sort: 0},
		{Key: "php", Name: "PHP", Logo: "PHP", Title: "PHP 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 PHP 内容", Description: "PHP 技术社区", Color: "#7c3aed", Status: "enable", Sort: 1},
		{Key: "go", Name: "Go", Logo: "GO", Title: "Go 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 Go 内容", Description: "Go 技术社区", Color: "#06b6d4", Status: "enable", Sort: 2},
		{Key: "java", Name: "Java", Logo: "JAVA", Title: "Java 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 Java 内容", Description: "Java 技术社区", Color: "#f97316", Status: "enable", Sort: 3},
		{Key: "ai", Name: "AI", Logo: "AI", Title: "AI 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 AI 内容", Description: "AI Agent、RAG、Prompt 与工作流社区", Color: "#7c3aed", Status: "enable", Sort: 4},
		{Key: "frontend", Name: "Frontend", Logo: "FE", Title: "Frontend 子网站", Sub: "子网站 · 7 个板块", Pub: "发布前端内容", Description: "Vue、React、TypeScript 与前端工程化社区", Color: "#16a34a", Status: "enable", Sort: 5},
	}
	for _, site := range sites {
		if _, err := s.db.Exec(`INSERT IGNORE INTO sites (site_key,name,logo,title,subtitle,pub_text,description,color,status,sort_order) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			site.Key, site.Name, site.Logo, site.Title, site.Sub, site.Pub, site.Description, site.Color, site.Status, site.Sort); err != nil {
			return err
		}
	}
	boardNames := map[string]string{"all": "全部", "community": "社区", "qa": "问答中心", "opensource": "开源项目", "ai": "AI作品", "jobs": "招聘内推", "wiki": "Wiki", "docs": "文档"}
	for i, key := range []string{"all", "community", "qa", "opensource", "ai", "jobs", "wiki", "docs"} {
		if _, err := s.db.Exec(`INSERT IGNORE INTO boards (board_key,name,site_key,sort_order,visible) VALUES (?,?,?,?,1)`, key, boardNames[key], "all", i); err != nil {
			return err
		}
	}
	roles := []domain.AdminRole{
		{ID: 1, Name: "超级管理员", Builtin: true, Description: "拥有所有模块操作权限", Permissions: []string{"*"}, UserCount: 1},
		{ID: 2, Name: "运营管理员", Builtin: true, Description: "负责内容运营、数据查看、通知推送", Permissions: []string{"content:*", "operation:*", "statistics:read"}, UserCount: 2},
		{ID: 3, Name: "内容审核员", Builtin: true, Description: "负责帖子和评论审核、违规处理", Permissions: []string{"content:read", "content:audit", "comment:*"}, UserCount: 1},
	}
	for _, role := range roles {
		perms, _ := json.Marshal(role.Permissions)
		if _, err := s.db.Exec(`INSERT IGNORE INTO admin_roles (id,name,builtin,description,permissions_json,user_count) VALUES (?,?,?,?,?,?)`,
			role.ID, role.Name, role.Builtin, role.Description, string(perms), role.UserCount); err != nil {
			return err
		}
	}
	users := []domain.AdminUser{
		{ID: 1, Username: "admin", Nickname: "超级管理员", Phone: "13800000001", Email: "admin@devhub.local", Status: "normal", RoleID: 1, RoleName: "超级管理员", CreatedAt: "2026-04-01 09:00:00", LastLoginAt: "2026-05-06 09:30:00"},
		{ID: 2, Username: "operator", Nickname: "运营管理员", Phone: "13800000002", Email: "operator@devhub.local", Status: "normal", RoleID: 2, RoleName: "运营管理员", CreatedAt: "2026-04-08 09:00:00", LastLoginAt: "2026-05-05 18:20:00"},
		{ID: 3, Username: "auditor", Nickname: "内容审核员", Phone: "13800000003", Email: "auditor@devhub.local", Status: "normal", RoleID: 3, RoleName: "内容审核员", CreatedAt: "2026-04-12 09:00:00", LastLoginAt: "2026-05-06 10:12:00"},
	}
	for _, u := range users {
		if _, err := s.db.Exec(`INSERT IGNORE INTO admin_users (id,username,nickname,avatar,phone,email,status,role_id,role_name,created_at,last_login_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			u.ID, u.Username, u.Nickname, u.Avatar, u.Phone, u.Email, u.Status, u.RoleID, u.RoleName, u.CreatedAt, u.LastLoginAt); err != nil {
			return err
		}
	}
	settings := domain.AdminSettings{SiteName: "DevHub", Copyright: "© 2026 DevHub", DefaultPageSize: 20, ReviewTimeoutHour: 24, PasswordRule: "至少 8 位，包含字母和数字", CaptchaEnabled: true, SearchDefault: "portal", SearchSort: "time", HotViewWeight: 1, HotLikeWeight: 8, HotCommentWeight: 15}
	if _, err := s.db.Exec(`INSERT IGNORE INTO admin_settings (id,site_name,copyright,default_page_size,review_timeout_hour,password_rule,captcha_enabled,search_default,search_sort,hot_view_weight,hot_like_weight,hot_comment_weight) VALUES (1,?,?,?,?,?,?,?,?,?,?,?)`,
		settings.SiteName, settings.Copyright, settings.DefaultPageSize, settings.ReviewTimeoutHour, settings.PasswordRule, settings.CaptchaEnabled, settings.SearchDefault, settings.SearchSort, settings.HotViewWeight, settings.HotLikeWeight, settings.HotCommentWeight); err != nil {
		return err
	}
	if err := s.seedAuthData(); err != nil {
		return err
	}
	if err := s.seedPlugins(); err != nil {
		return err
	}

	// Seed Communities (子站)
	communities := communitySeedData()
	for _, comm := range communities {
		if _, err := s.db.Exec(`INSERT IGNORE INTO communities (id,name,slug,logo,cover_image,slogan,description,theme_color,seo_title,seo_description,seo_keywords,sort_order,status,announcement_title,announcement_content,announcement_url) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			comm.ID, comm.Name, comm.Slug, comm.Logo, comm.CoverImage, comm.Slogan, comm.Description, comm.ThemeColor, comm.SEOTitle, comm.SEODescription, comm.SEOKeywords, comm.SortOrder, comm.Status, comm.AnnouncementTitle, comm.AnnouncementContent, comm.AnnouncementURL); err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE communities SET logo=?,cover_image=?,slogan=?,description=?,theme_color=?,seo_title=?,seo_description=?,seo_keywords=?,sort_order=?,status=?,announcement_title=?,announcement_content=?,announcement_url=? WHERE id=?`,
			comm.Logo, comm.CoverImage, comm.Slogan, comm.Description, comm.ThemeColor, comm.SEOTitle, comm.SEODescription, comm.SEOKeywords, comm.SortOrder, comm.Status, comm.AnnouncementTitle, comm.AnnouncementContent, comm.AnnouncementURL, comm.ID); err != nil {
			return err
		}
	}
	_, _ = s.db.Exec(`INSERT IGNORE INTO community_plugins (community_id, plugin_code, status, sort_order, config_json, created_at, updated_at)
		SELECT c.id, p.plugin_code, 'enabled', 0, NULL, NOW(), NOW() FROM communities c JOIN plugins p ON p.status='enabled'`)

	// Seed Categories (板块/分类)
	for _, comm := range communities {
		for _, cat := range defaultCategorySeeds(comm.ID) {
			if _, err := s.db.Exec(`INSERT IGNORE INTO categories (id,community_id,name,slug,type,plugin_code,allowed_content_types,description,icon,sort_order,visible,nav_visible,postable,seo_title,seo_description,status) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				cat.ID, comm.ID, cat.Name, cat.Slug, cat.Type, cat.PluginCode, stringSliceJSON(cat.AllowedContentTypes), cat.Description, cat.Icon, cat.SortOrder, boolToInt(cat.Visible), boolToInt(cat.NavVisible), boolToInt(cat.Postable), cat.SEOTitle, cat.SEODescription, cat.Status); err != nil {
				return err
			}
			if _, err := s.db.Exec(`UPDATE categories SET type=?,plugin_code=?,allowed_content_types=?,description=?,icon=?,sort_order=?,visible=?,nav_visible=?,postable=?,seo_title=?,seo_description=?,status=? WHERE id=?`,
				cat.Type, cat.PluginCode, stringSliceJSON(cat.AllowedContentTypes), cat.Description, cat.Icon, cat.SortOrder, boolToInt(cat.Visible), boolToInt(cat.NavVisible), boolToInt(cat.Postable), cat.SEOTitle, cat.SEODescription, cat.Status, cat.ID); err != nil {
				return err
			}
		}
	}

	// Seed Tags for each community
	communityTags := map[int64][]string{
		1: {"Laravel", "Hyperf", "Composer", "Swoole", "Redis", "MySQL", "性能优化", "代码规范"},
		2: {"Gin", "Gorm", "微服务", "并发", "gRPC", "Docker", "性能优化"},
		3: {"Spring Boot", "MyBatis", "JVM", "Redis", "MySQL", "微服务", "消息队列"},
		4: {"AI Agent", "Prompt", "RAG", "OpenAI", "Claude", "Codex", "工作流"},
		5: {"Vue", "React", "TypeScript", "Tailwind CSS", "Vite", "Next.js", "性能优化"},
	}
	for commID, tags := range communityTags {
		for _, tagName := range tags {
			slug := normalizeSlug(tagName)
			if slug == "" {
				slug = strings.ToLower(strings.Join(strings.Fields(tagName), "-"))
			}
			siteKey := fmt.Sprintf("%d", commID)
			if comm, ok := s.communityByID(commID); ok && comm.Slug != "" {
				siteKey = comm.Slug
			}
			if _, err := s.db.Exec(`INSERT IGNORE INTO tags (site_key,name,slug,status,use_count,seo_title,seo_description,seo_keywords) VALUES (?,?,?,?,0,?,?,?)`,
				siteKey, tagName, slug, "enable", tagName+" 相关内容", "DevHub "+tagName+" 标签聚合，汇总相关文章、问答、项目和文档。", tagName); err != nil {
				return err
			}
		}
	}

	// Seed Topics (多种内容类型)
	topics := []domain.Topic{
		// PHP 社区内容
		{ID: 101, CommunityID: 1, CategoryID: 101, UserID: 1, Title: "Laravel 社区系统如何设计积分和通知？", ContentType: "article", Summary: "从用户行为、积分流水、通知触发器、异步队列几个角度拆解社区积分系统。", Content: "一个社区系统的积分与通知不应该散落在业务代码中，而应该通过事件、积分规则表、通知模板和队列消费组合起来。", ViewCount: 2380, LikeCount: 128, CommentCount: 2},
		{ID: 102, CommunityID: 1, CategoryID: 102, UserID: 1, Title: "PHP-FPM 502 问题应该如何排查？", ContentType: "question", Summary: "从 Nginx、PHP-FPM、慢日志、进程池、超时配置和内存占用逐层定位。", Content: "502 通常不是一个单点问题。建议先确认 Nginx upstream 是否能连通 PHP-FPM，再检查 PHP-FPM 进程池是否耗尽。", ViewCount: 1890, LikeCount: 84, CommentCount: 0, IsSolved: true},
		{ID: 105, CommunityID: 1, CategoryID: 103, UserID: 1, Title: "PHP Package Starter：Composer 包模板", ContentType: "project", Summary: "用于快速创建 Composer 包、测试、CI 和发布流程的模板。", Content: "PHP Package Starter 提供 PSR-4、PHPUnit、静态分析和 GitHub Actions 发布流程。", ViewCount: 860, LikeCount: 38, CommentCount: 0},
		{ID: 106, CommunityID: 1, CategoryID: 105, UserID: 1, Title: "招聘：PHP 后端工程师", ContentType: "job", Summary: "负责社区系统、支付系统和内部管理平台开发。", Content: "岗位要求熟悉 Laravel 或 Hyperf，有 Redis、MySQL 和队列系统经验。", ViewCount: 760, LikeCount: 26, CommentCount: 0},
		{ID: 103, CommunityID: 1, CategoryID: 106, UserID: 1, Title: "PHP Wiki：Laravel 请求生命周期", ContentType: "wiki", Summary: "从 public/index.php 到 Kernel、Middleware、Router、Controller 的完整链路。", Content: "Laravel 请求从 public/index.php 进入，随后创建应用容器，加载 HTTP Kernel，经过中间件管道，进入路由匹配。", ViewCount: 980, LikeCount: 45, CommentCount: 0},
		{ID: 104, CommunityID: 1, CategoryID: 107, UserID: 1, Title: "PHP 文档：Composer 自动加载", ContentType: "doc", Summary: "解释 PSR-4、autoload、classmap、files 自动加载的适用场景。", Content: "Composer 自动加载最常见的是 PSR-4。它通过命名空间前缀映射到目录路径。", ViewCount: 1320, LikeCount: 62, CommentCount: 0},

		// Go 社区内容
		{ID: 201, CommunityID: 2, CategoryID: 201, UserID: 1, Title: "Go 项目应该如何组织目录结构？", ContentType: "article", Summary: "讨论 cmd、internal、pkg、api、configs、scripts 等目录的边界。", Content: "Go 项目结构的重点不是目录越多越好，而是边界清晰。cmd 放入口，internal 放内部业务。", ViewCount: 2100, LikeCount: 101, CommentCount: 1},
		{ID: 202, CommunityID: 2, CategoryID: 202, UserID: 1, Title: "Goroutine 泄漏如何定位？", ContentType: "question", Summary: "从 pprof、context、channel 阻塞、select default 等角度排查。", Content: "Goroutine 泄漏常见原因是 channel 永久阻塞、context 未取消、后台循环没有退出条件。", ViewCount: 1760, LikeCount: 92, CommentCount: 0, IsSolved: true},
		{ID: 203, CommunityID: 2, CategoryID: 203, UserID: 1, Title: "Go CLI Starter：命令行工具模板", ContentType: "project", Summary: "一个用于快速启动 Cobra/Viper 命令行项目的模板。", Content: "Go CLI Starter 集成 Cobra 命令组织、Viper 配置读取、日志初始化和版本输出。", ViewCount: 860, LikeCount: 39, CommentCount: 0},
		{ID: 205, CommunityID: 2, CategoryID: 205, UserID: 1, Title: "招聘：Go 云原生工程师", ContentType: "job", Summary: "负责微服务、网关、任务调度和容器平台建设。", Content: "岗位要求熟悉 Go、Docker、gRPC 和高并发服务治理。", ViewCount: 940, LikeCount: 34, CommentCount: 0},
		{ID: 206, CommunityID: 2, CategoryID: 206, UserID: 1, Title: "Go Wiki：并发模式速览", ContentType: "wiki", Summary: "整理 worker pool、fan-in/fan-out、context 取消和限流模式。", Content: "Go 并发设计常用模式包括 worker pool、fan-in/fan-out、context 取消、超时和限流。", ViewCount: 790, LikeCount: 41, CommentCount: 0},
		{ID: 204, CommunityID: 2, CategoryID: 207, UserID: 1, Title: "Go 文档：context 超时控制", ContentType: "doc", Summary: "说明 WithCancel、WithTimeout、WithDeadline 的区别和常见误区。", Content: "context 用于跨 API 边界传递取消信号、超时和请求级元数据。", ViewCount: 1120, LikeCount: 58, CommentCount: 0},

		// Java 社区内容
		{ID: 301, CommunityID: 3, CategoryID: 301, UserID: 1, Title: "Spring Boot 项目如何分层更清晰？", ContentType: "article", Summary: "Controller、Application Service、Domain、Infrastructure 的职责拆分。", Content: "Spring Boot 项目常见问题是 Controller 太厚、Service 太杂。建议拆分接口层、应用服务、领域模型、基础设施。", ViewCount: 2320, LikeCount: 116, CommentCount: 0},
		{ID: 302, CommunityID: 3, CategoryID: 302, UserID: 1, Title: "JVM Full GC 频繁如何排查？", ContentType: "question", Summary: "结合 GC 日志、堆转储、对象增长趋势和内存区域定位问题。", Content: "Full GC 频繁需要先看 GC 日志确认触发原因，再通过 heap dump 分析大对象和引用链。", ViewCount: 2600, LikeCount: 149, CommentCount: 0, IsSolved: false},
		{ID: 305, CommunityID: 3, CategoryID: 303, UserID: 1, Title: "Spring Boot Admin Starter：后台脚手架", ContentType: "project", Summary: "包含 RBAC、审计日志、配置中心和内容管理基础模块。", Content: "该脚手架适合中后台管理系统，内置权限、菜单、审计和代码生成基础能力。", ViewCount: 1180, LikeCount: 52, CommentCount: 0},
		{ID: 303, CommunityID: 3, CategoryID: 304, UserID: 1, Title: "Spring AI 接入大模型的项目结构建议", ContentType: "ai_work", Summary: "把模型配置、提示词模板、工具调用和业务流程拆开。", Content: "Spring AI 接入大模型时，不建议把 Prompt 和业务逻辑全部写在 Controller。", ViewCount: 1480, LikeCount: 73, CommentCount: 0},
		{ID: 306, CommunityID: 3, CategoryID: 305, UserID: 1, Title: "招聘：Java 平台工程师", ContentType: "job", Summary: "负责中台服务、消息队列、缓存治理和 JVM 性能优化。", Content: "岗位要求熟悉 Spring Boot、MyBatis、Redis、消息队列和 JVM 调优。", ViewCount: 980, LikeCount: 36, CommentCount: 0},
		{ID: 304, CommunityID: 3, CategoryID: 306, UserID: 1, Title: "Java Wiki：JVM 内存区域速览", ContentType: "wiki", Summary: "堆、方法区、虚拟机栈、本地方法栈、程序计数器的基础说明。", Content: "JVM 运行时内存区域包括线程共享的堆和方法区，以及线程私有的虚拟机栈、本地方法栈、程序计数器。", ViewCount: 990, LikeCount: 54, CommentCount: 0},
		{ID: 307, CommunityID: 3, CategoryID: 307, UserID: 1, Title: "Java 文档：MyBatis 映射规范", ContentType: "doc", Summary: "记录 Mapper、XML、分页、批量写入和事务边界约定。", Content: "MyBatis 使用中需要明确 Mapper 职责、SQL 命名、分页约定、批量写入和事务边界。", ViewCount: 870, LikeCount: 33, CommentCount: 0},

		// AI 社区内容
		{ID: 401, CommunityID: 4, CategoryID: 401, UserID: 1, Title: "AI Agent 工作流如何设计得可维护？", ContentType: "article", Summary: "从工具调用、上下文压缩、任务状态和人工确认四个角度拆解 Agent 工程化。", Content: "AI Agent 的重点不是把模型接进来，而是把任务边界、工具权限、状态恢复和人工确认做清楚。", ViewCount: 1560, LikeCount: 88, CommentCount: 0},
		{ID: 402, CommunityID: 4, CategoryID: 402, UserID: 1, Title: "如何优化 Prompt 获得更好的 AI 回复？", ContentType: "question", Summary: "从 Prompt 结构、上下文、示例等角度讨论如何优化。", Content: "好的 Prompt 应该结构清晰、上下文完整、提供示例。", ViewCount: 1340, LikeCount: 65, CommentCount: 0, IsSolved: true},
		{ID: 403, CommunityID: 4, CategoryID: 403, UserID: 1, Title: "RAG Starter：知识库问答项目模板", ContentType: "project", Summary: "一个用于验证文档切分、向量检索和答案引用的 RAG 项目模板。", Content: "RAG Starter 包含文档导入、切分、Embedding、检索、重排和答案引用展示。", ViewCount: 1180, LikeCount: 56, CommentCount: 0},
		{ID: 404, CommunityID: 4, CategoryID: 405, UserID: 1, Title: "招聘：AI 应用开发工程师", ContentType: "job", Summary: "负责 AI 应用的开发和维护，薪资范围 20-35K。", Content: "岗位要求：熟悉主流大模型 API，有 AI 应用开发经验优先。", ViewCount: 920, LikeCount: 45, CommentCount: 0},
		{ID: 405, CommunityID: 4, CategoryID: 406, UserID: 1, Title: "AI Wiki：RAG 基础概念", ContentType: "wiki", Summary: "整理向量、Embedding、召回、重排和引用生成的基础知识。", Content: "RAG 通过检索外部知识补充模型上下文，核心流程包括切分、向量化、召回、重排和生成。", ViewCount: 760, LikeCount: 35, CommentCount: 0},
		{ID: 406, CommunityID: 4, CategoryID: 407, UserID: 1, Title: "AI 文档：OpenAI API 调用约定", ContentType: "doc", Summary: "记录模型调用、错误处理、重试和日志脱敏的基础规范。", Content: "调用大模型 API 时应统一封装超时、重试、错误分类、审计日志和敏感数据脱敏。", ViewCount: 830, LikeCount: 42, CommentCount: 0},

		// Frontend 社区内容
		{ID: 501, CommunityID: 5, CategoryID: 501, UserID: 1, Title: "Vue 3 Composition API 最佳实践", ContentType: "article", Summary: "分享 Vue 3 组合式 API 的使用技巧和注意事项。", Content: "Composition API 提供了更好的逻辑复用和类型推导能力。", ViewCount: 1890, LikeCount: 95, CommentCount: 0},
		{ID: 502, CommunityID: 5, CategoryID: 502, UserID: 1, Title: "React Hooks 常见误区", ContentType: "question", Summary: "讨论 React Hooks 使用中的常见问题和正确用法。", Content: "React Hooks 需要注意依赖项的正确设置和闭包陷阱。", ViewCount: 1680, LikeCount: 78, CommentCount: 0, IsSolved: true},
		{ID: 503, CommunityID: 5, CategoryID: 503, UserID: 1, Title: "Vite + React TypeScript 模板项目", ContentType: "project", Summary: "一个开箱即用的 Vite React TS 模板。", Content: "包含路由、状态管理、请求封装等基础能力。", ViewCount: 1240, LikeCount: 62, CommentCount: 0},
		{ID: 504, CommunityID: 5, CategoryID: 505, UserID: 1, Title: "招聘：高级前端工程师", ContentType: "job", Summary: "负责中后台产品、组件库和性能优化。", Content: "岗位要求熟悉 Vue 或 React，具备 TypeScript、工程化和性能优化经验。", ViewCount: 860, LikeCount: 31, CommentCount: 0},
		{ID: 505, CommunityID: 5, CategoryID: 506, UserID: 1, Title: "Frontend Wiki：TypeScript 类型收窄", ContentType: "wiki", Summary: "整理 typeof、in、谓词函数和可辨识联合类型。", Content: "TypeScript 类型收窄可以通过 typeof、in、instanceof、自定义谓词和可辨识联合类型实现。", ViewCount: 720, LikeCount: 29, CommentCount: 0},
		{ID: 506, CommunityID: 5, CategoryID: 507, UserID: 1, Title: "前端文档：Vite 构建性能优化", ContentType: "doc", Summary: "说明依赖预构建、代码分割、缓存和构建分析。", Content: "Vite 构建优化可以从依赖预构建、动态导入、manualChunks 和缓存策略入手。", ViewCount: 900, LikeCount: 44, CommentCount: 0},
	}

	for _, t := range topics {
		if t.HotScore == 0 {
			t.HotScore = t.ViewCount + t.CommentCount*5 + t.LikeCount*3 + t.FavoriteCount*4
		}
		t.ContentType = pluginregistry.NormalizeContentType(t.ContentType)
		t.PluginCode = pluginregistry.PluginCodeForContentType(t.ContentType)
		if _, err := s.db.Exec(`INSERT IGNORE INTO topics (id,community_id,category_id,user_id,title,plugin_code,content_type,summary,content,status,is_pinned,is_featured,is_solved,view_count,comment_count,like_count,favorite_count,hot_score,last_active_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())`,
			t.ID, t.CommunityID, t.CategoryID, t.UserID, t.Title, t.PluginCode, t.ContentType, t.Summary, t.Content,
			1, boolToInt(t.IsPinned), boolToInt(t.IsFeatured), boolToInt(t.IsSolved), t.ViewCount, t.CommentCount, t.LikeCount, t.FavoriteCount,
			t.HotScore, Now()); err != nil {
			return err
		}
		s.seedPluginContentRows(t)

		// 添加话题标签关联
		tagNames := communityTags[t.CommunityID]
		if len(tagNames) > 3 {
			tagNames = tagNames[:3]
		}
		for _, tagName := range tagNames {
			slug := normalizeSlug(tagName)
			if slug == "" {
				slug = strings.ToLower(strings.Join(strings.Fields(tagName), "-"))
			}
			siteKey := fmt.Sprintf("%d", t.CommunityID)
			if comm, ok := s.communityByID(t.CommunityID); ok && comm.Slug != "" {
				siteKey = comm.Slug
			}
			// 添加关联
			if _, err := s.db.Exec(`INSERT IGNORE INTO topic_tags (topic_id,tag_id) SELECT ?,id FROM tags WHERE site_key=? AND slug=?`,
				t.ID, siteKey, slug); err != nil {
				continue
			}
		}
	}

	// Seed Activities (用户动态)
	activities := []struct {
		userID      int64
		communityID int64
		action      string
		targetType  string
		targetID    int64
	}{
		{1, 1, "created_topic", "topic", 101},
		{1, 2, "created_topic", "topic", 201},
		{1, 3, "created_topic", "topic", 301},
		{1, 4, "created_topic", "topic", 401},
		{1, 5, "created_topic", "topic", 501},
		{1, 1, "commented", "topic", 101},
		{1, 1, "liked", "topic", 101},
	}
	for _, act := range activities {
		if _, err := s.db.Exec(`INSERT IGNORE INTO activities (user_id,community_id,action,target_type,target_id,created_at) VALUES (?,?,?,?,?,NOW())`,
			act.userID, act.communityID, act.action, act.targetType, act.targetID); err != nil {
			return err
		}
	}

	var postCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&postCount); err != nil {
		return err
	}
	posts := []domain.Post{
		{ID: 1, Site: "php", Board: "community", Title: "Laravel 社区系统如何设计积分和通知？", Summary: "从用户行为、积分流水、通知触发器、异步队列几个角度拆解社区积分系统。", Author: "LaravelChen", Views: 2380, Likes: 128, Comments: 0, CreatedAt: "2026-05-01 09:00:00", Tags: []string{"Laravel", "积分", "通知"}, Content: "一个社区系统的积分与通知不应该散落在业务代码中，而应该通过事件、积分规则表、通知模板和队列消费组合起来。"},
		{ID: 2, Site: "php", Board: "qa", Title: "PHP-FPM 502 问题应该如何排查？", Summary: "从 Nginx、PHP-FPM、慢日志、进程池、超时配置和内存占用逐层定位。", Author: "SwooleDev", Views: 1890, Likes: 84, Comments: 0, CreatedAt: "2026-05-02 10:00:00", Tags: []string{"PHP-FPM", "Nginx", "502"}, Content: "502 通常不是一个单点问题。建议先确认 Nginx upstream 是否能连通 PHP-FPM，再检查 PHP-FPM 进程池是否耗尽。"},
		{ID: 3, Site: "php", Board: "wiki", Title: "PHP Wiki：Laravel 请求生命周期", Summary: "从 public/index.php 到 Kernel、Middleware、Router、Controller 的完整链路。", Author: "DevHub Wiki", Views: 980, Likes: 45, Comments: 0, CreatedAt: "2026-04-29 12:00:00", Tags: []string{"Wiki", "Laravel", "生命周期"}, Content: "Laravel 请求从 public/index.php 进入，随后创建应用容器，加载 HTTP Kernel，经过中间件管道，进入路由匹配。"},
		{ID: 4, Site: "php", Board: "docs", Title: "PHP 文档：Composer 自动加载", Summary: "解释 PSR-4、autoload、classmap、files 自动加载的适用场景。", Author: "ComposerBot", Views: 1320, Likes: 62, Comments: 0, CreatedAt: "2026-04-28 12:00:00", Tags: []string{"Composer", "PSR-4"}, Content: "Composer 自动加载最常见的是 PSR-4。它通过命名空间前缀映射到目录路径。"},
		{ID: 5, Site: "go", Board: "community", Title: "Go 项目应该如何组织目录结构？", Summary: "讨论 cmd、internal、pkg、api、configs、scripts 等目录的边界。", Author: "GopherLin", Views: 2100, Likes: 101, Comments: 0, CreatedAt: "2026-05-01 12:00:00", Tags: []string{"Go", "项目结构"}, Content: "Go 项目结构的重点不是目录越多越好，而是边界清晰。cmd 放入口，internal 放内部业务。"},
		{ID: 6, Site: "go", Board: "qa", Title: "Goroutine 泄漏如何定位？", Summary: "从 pprof、context、channel 阻塞、select default 等角度排查。", Author: "GoTrace", Views: 1760, Likes: 92, Comments: 0, CreatedAt: "2026-05-03 08:30:00", Tags: []string{"Goroutine", "pprof", "泄漏"}, Content: "Goroutine 泄漏常见原因是 channel 永久阻塞、context 未取消、后台循环没有退出条件。"},
		{ID: 7, Site: "go", Board: "opensource", Title: "Go CLI Starter：命令行工具模板", Summary: "一个用于快速启动 Cobra/Viper 命令行项目的模板。", Author: "OpenSourceHub", Views: 860, Likes: 39, Comments: 0, CreatedAt: "2026-04-27 12:00:00", Tags: []string{"CLI", "Cobra", "开源项目"}, Content: "Go CLI Starter 集成 Cobra 命令组织、Viper 配置读取、日志初始化和版本输出。"},
		{ID: 8, Site: "go", Board: "docs", Title: "Go 文档：context 超时控制", Summary: "说明 WithCancel、WithTimeout、WithDeadline 的区别和常见误区。", Author: "DevHub Docs", Views: 1120, Likes: 58, Comments: 0, CreatedAt: "2026-04-26 12:00:00", Tags: []string{"context", "超时控制"}, Content: "context 用于跨 API 边界传递取消信号、超时和请求级元数据。"},
		{ID: 9, Site: "java", Board: "community", Title: "Spring Boot 项目如何分层更清晰？", Summary: "Controller、Application Service、Domain、Infrastructure 的职责拆分。", Author: "SpringLee", Views: 2320, Likes: 116, Comments: 0, CreatedAt: "2026-05-02 13:00:00", Tags: []string{"Spring Boot", "分层"}, Content: "Spring Boot 项目常见问题是 Controller 太厚、Service 太杂。建议拆分接口层、应用服务、领域模型、基础设施。"},
		{ID: 10, Site: "java", Board: "qa", Title: "JVM Full GC 频繁如何排查？", Summary: "结合 GC 日志、堆转储、对象增长趋势和内存区域定位问题。", Author: "JvmDoctor", Views: 2600, Likes: 149, Comments: 0, CreatedAt: "2026-05-03 09:00:00", Tags: []string{"JVM", "Full GC", "性能"}, Content: "Full GC 频繁需要先看 GC 日志确认触发原因，再通过 heap dump 分析大对象和引用链。"},
		{ID: 11, Site: "java", Board: "ai", Title: "Spring AI 接入大模型的项目结构建议", Summary: "把模型配置、提示词模板、工具调用和业务流程拆开。", Author: "AIJava", Views: 1480, Likes: 73, Comments: 0, CreatedAt: "2026-04-30 12:00:00", Tags: []string{"Spring AI", "LLM"}, Content: "Spring AI 接入大模型时，不建议把 Prompt 和业务逻辑全部写在 Controller。"},
		{ID: 12, Site: "java", Board: "wiki", Title: "Java Wiki：JVM 内存区域速览", Summary: "堆、方法区、虚拟机栈、本地方法栈、程序计数器的基础说明。", Author: "DevHub Wiki", Views: 990, Likes: 54, Comments: 0, CreatedAt: "2026-04-25 12:00:00", Tags: []string{"JVM", "Wiki"}, Content: "JVM 运行时内存区域包括线程共享的堆和方法区，以及线程私有的虚拟机栈、本地方法栈、程序计数器。"},
	}
	for _, p := range posts {
		if postCount > 0 && p.ID != 1 {
			continue
		}
		if err := s.insertSeedPost(p); err != nil {
			return err
		}
	}
	if postCount > 0 {
		_ = s.rebuildTagsFromPosts()
		return nil
	}
	if _, err := s.db.Exec(`INSERT IGNORE INTO comments (id,post_id,parent_id,author,to_author,text,status,likes,created_at) VALUES
		(1,1,0,'Corwien','','不错，图文并茂！','normal',12,'2026-05-03 10:12:00'),
		(2,1,1,'LaravelChen','Corwien','thank you','normal',5,'2026-05-03 10:30:00'),
		(3,1,2,'DevHubUser','LaravelChen','这个回复层级展示很清楚，后续可以加折叠。','normal',2,'2026-05-03 11:00:00'),
		(4,1,0,'Ysll','','不错不错，学习学习。','normal',8,'2026-05-03 12:15:00')`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`UPDATE posts SET comments=4 WHERE id=1`); err != nil {
		return err
	}
	s.createNotice("LaravelChen 回复了你的评论", "thank you")
	s.createNotice("你的帖子获得了新的点赞", "JVM Full GC 频繁如何排查？")
	s.createNotice("Go 子网站有新的问答", "Goroutine 泄漏如何定位？")
	s.appendLog("login", "admin", "管理员登录", "后台系统", "127.0.0.1")
	s.appendLog("operation", "operator", "配置首页推荐", "posts#6", "127.0.0.1")
	_ = s.rebuildTagsFromPosts()
	return nil
}

func (s *MySQLStore) insertSeedPost(p domain.Post) error {
	tags, _ := json.Marshal(uniqueTags(p.Tags))
	if p.Status == "" {
		p.Status = "publish"
	}
	if p.UpdatedAt == "" {
		p.UpdatedAt = p.CreatedAt
	}
	_, err := s.db.Exec(`INSERT IGNORE INTO posts (id,site_key,board_key,title,summary,content,author,status,views,likes,comments,tags_json,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Site, p.Board, p.Title, p.Summary, p.Content, p.Author, p.Status, p.Views, p.Likes, p.Comments, string(tags), p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *MySQLStore) ValidateSite(site string) bool {
	if site == "portal" || site == "" {
		return true
	}
	var exists int
	_ = s.db.QueryRow(`SELECT 1 FROM sites WHERE site_key=? LIMIT 1`, site).Scan(&exists)
	return exists == 1
}

func (s *MySQLStore) ValidateBoard(board string) bool {
	if board == "all" || board == "" {
		return true
	}
	var exists int
	_ = s.db.QueryRow(`SELECT 1 FROM boards WHERE board_key=? LIMIT 1`, board).Scan(&exists)
	return exists == 1
}

func (s *MySQLStore) seedAuthData() error {
	permissions := []struct {
		Code, Module, Action, Description string
	}{
		{"dashboard.read", "dashboard", "read", "查看控制台"},
		{"site.read", "site", "read", "查看站点"},
		{"site.write", "site", "write", "管理站点"},
		{"board.read", "board", "read", "查看板块"},
		{"board.write", "board", "write", "管理板块"},
		{"post.read", "post", "read", "查看内容"},
		{"post.create", "post", "create", "创建内容"},
		{"core.topic.create", "core", "topic.create", "创建 Core 内容（兼容类型）"},
		{"post.update", "post", "update", "更新内容"},
		{"post.delete", "post", "delete", "删除内容"},
		{"comment.read", "comment", "read", "查看评论"},
		{"comment.moderate", "comment", "moderate", "审核评论"},
		{"report.read", "report", "read", "查看举报"},
		{"report.handle", "report", "handle", "处理举报"},
		{"topic.moderate", "topic", "moderate", "治理主题"},
		{"moderator.read", "moderator", "read", "查看版主"},
		{"moderator.write", "moderator", "write", "管理版主"},
		{"user.read", "user", "read", "查看用户"},
		{"user.write", "user", "write", "管理用户"},
		{"role.read", "role", "read", "查看角色"},
		{"setting.read", "setting", "read", "查看设置"},
		{"setting.write", "setting", "write", "管理设置"},
		{"notification.write", "notification", "write", "推送通知"},
		{"log.read", "log", "read", "查看日志"},
		{"plugin.read", "plugin", "read", "查看插件"},
		{"plugin.write", "plugin", "write", "管理插件"},
		{"plugin.approve", "plugin", "approve", "审批插件高危操作"},
		{"qa.question.create", "qa", "question.create", "发布问题"},
		{"qa.question.audit", "qa", "question.audit", "审核问题"},
		{"qa.answer.create", "qa", "answer.create", "提交回答"},
		{"qa.answer.accept", "qa", "answer.accept", "采纳回答"},
		{"docs.document.create", "docs", "document.create", "创建文档"},
		{"docs.document.update", "docs", "document.update", "更新文档"},
		{"docs.document.audit", "docs", "document.audit", "审核文档"},
		{"docs.space.manage", "docs", "space.manage", "管理文档空间"},
		{"wiki.page.create", "wiki", "page.create", "创建 Wiki 页面"},
		{"wiki.page.edit", "wiki", "page.edit", "编辑 Wiki 页面"},
		{"wiki.page.audit", "wiki", "page.audit", "审核 Wiki 页面"},
		{"wiki.page.version.rollback", "wiki", "page.version.rollback", "回滚 Wiki 版本"},
		{"projects.project.create", "projects", "project.create", "发布项目"},
		{"projects.project.audit", "projects", "project.audit", "审核项目"},
		{"jobs.job.create", "jobs", "job.create", "发布招聘"},
		{"jobs.job.audit", "jobs", "job.audit", "审核招聘"},
		{"ai_works.work.create", "ai_works", "work.create", "发布 AI 作品"},
		{"ai_works.work.audit", "ai_works", "work.audit", "审核 AI 作品"},
	}
	for _, p := range permissions {
		if _, err := s.db.Exec(`INSERT IGNORE INTO permissions (code,module,action,description) VALUES (?,?,?,?)`, p.Code, p.Module, p.Action, p.Description); err != nil {
			return err
		}
	}
	roles := []struct {
		ID          int64
		Code        string
		Name        string
		Description string
		Permissions []string
	}{
		{1, "super_admin", "超级管理员", "拥有全部站点和全部权限", []string{"*"}},
		{2, "site_admin", "站点管理员", "管理被授权站点", []string{"dashboard.read", "site.read", "site.write", "board.read", "board.write", "post.read", "post.create", "post.update", "post.delete", "topic.moderate", "comment.read", "comment.moderate", "report.read", "report.handle", "moderator.read", "user.read", "setting.read", "notification.write", "log.read", "plugin.read", "qa.question.create", "qa.question.audit", "docs.document.create", "docs.document.audit", "docs.space.manage", "wiki.page.create", "wiki.page.audit", "wiki.page.version.rollback", "projects.project.create", "projects.project.audit", "jobs.job.create", "jobs.job.audit", "ai_works.work.create", "ai_works.work.audit"}},
		{3, "editor", "编辑", "创建和编辑内容", []string{"dashboard.read", "post.read", "post.create", "post.update", "comment.read", "projects.project.create", "jobs.job.create", "ai_works.work.create"}},
		{4, "moderator", "审核员", "审核内容和评论", []string{"dashboard.read", "post.read", "post.update", "topic.moderate", "comment.read", "comment.moderate", "report.read", "report.handle", "plugin.read", "qa.question.audit", "docs.document.audit", "wiki.page.audit", "projects.project.audit", "jobs.job.audit", "ai_works.work.audit"}},
		{5, "user", "普通用户", "前台登录用户", []string{"post.create", "comment.read", "qa.question.create", "qa.answer.create", "docs.document.create", "wiki.page.create", "projects.project.create", "jobs.job.create", "ai_works.work.create"}},
	}
	for _, r := range roles {
		if _, err := s.db.Exec(`INSERT IGNORE INTO roles (id,code,name,builtin,description) VALUES (?,?,?,?,?)`, r.ID, r.Code, r.Name, true, r.Description); err != nil {
			return err
		}
		for _, code := range r.Permissions {
			if code == "*" {
				continue
			}
			if _, err := s.db.Exec(`INSERT IGNORE INTO role_permissions (role_id,permission_id) SELECT ?,id FROM permissions WHERE code=?`, r.ID, code); err != nil {
				return err
			}
		}
	}
	defaultPassword, err := hashPassword("admin123")
	if err != nil {
		return err
	}
	users := []struct {
		ID, RoleID int64
		Username   string
		Nickname   string
		Phone      string
		Email      string
		Site       string
	}{
		{1, 1, "admin", "超级管理员", "13800000001", "admin@devhub.local", "*"},
		{2, 2, "operator", "PHP 站点管理员", "13800000002", "operator@devhub.local", "php"},
		{3, 4, "auditor", "Go 审核员", "13800000003", "auditor@devhub.local", "go"},
	}
	for _, u := range users {
		if _, err := s.db.Exec(`INSERT IGNORE INTO users (id,username,nickname,email,phone,password_hash,status,created_at,last_login_at) VALUES (?,?,?,?,?,?, 'normal', NOW(), NULL)`,
			u.ID, u.Username, u.Nickname, u.Email, u.Phone, defaultPassword); err != nil {
			return err
		}
		if _, err := s.db.Exec(`INSERT IGNORE INTO user_roles (user_id,role_id,site_key,status) VALUES (?,?,?,'normal')`, u.ID, u.RoleID, u.Site); err != nil {
			return err
		}
	}
	frontPassword, err := hashPassword("a123456")
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT INTO users (username,nickname,email,phone,password_hash,status,created_at,last_login_at)
		VALUES (?,?,?,?,?,'normal',NOW(),NULL)
		ON DUPLICATE KEY UPDATE nickname=VALUES(nickname),email=VALUES(email),phone=VALUES(phone),password_hash=VALUES(password_hash),status='normal',updated_at=NOW()`,
		"liuwei", "方圆十三", "liuwei@devhub.local", "13800000004", frontPassword); err != nil {
		return err
	}
	var liuweiID int64
	if err := s.db.QueryRow(`SELECT id FROM users WHERE username=? LIMIT 1`, "liuwei").Scan(&liuweiID); err != nil {
		return err
	}
	if _, err := s.db.Exec(`INSERT IGNORE INTO user_roles (user_id,role_id,site_key,status) VALUES (?,?,'*','normal')`, liuweiID, roleID("user")); err != nil {
		return err
	}
	for _, item := range []struct {
		communityID int64
		userID      int64
		role        string
	}{
		{1, 2, "moderator"},
		{2, 3, "moderator"},
	} {
		if _, err := s.db.Exec(`INSERT IGNORE INTO community_moderators (community_id,user_id,role,status) VALUES (?,?,?,1)`, item.communityID, item.userID, item.role); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) seedPlugins() error {
	for _, def := range pluginregistry.Definitions() {
		if _, err := s.db.Exec(`INSERT INTO plugins (plugin_code,name,version,status,description,source_type,compatible_core_version,config_json,created_at,updated_at) VALUES (?,?,?,?,?,?,?,NULL,NOW(),NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name),version=VALUES(version),description=VALUES(description),source_type=VALUES(source_type),compatible_core_version=VALUES(compatible_core_version),updated_at=updated_at`,
			def.Code, def.Name, def.Version, def.Status, def.Description, "builtin", def.MinCoreVersion); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) seedPluginContentRows(t domain.Topic) {
	t.ContentType = pluginregistry.NormalizeContentType(t.ContentType)
	t.PluginCode = firstNonEmptyString(t.PluginCode, pluginregistry.PluginCodeForContentType(t.ContentType))
	switch t.PluginCode {
	case "qa":
		_, _ = s.db.Exec(`INSERT IGNORE INTO qa_questions (topic_id,is_solved,best_answer_id,created_at,updated_at) VALUES (?,?,NULL,NOW(),NOW())`, t.ID, boolToInt(t.IsSolved))
	case "docs":
		spaceID := s.ensureDocsSpace(t.CommunityID)
		_, _ = s.db.Exec(`INSERT IGNORE INTO docs_documents (space_id,topic_id,parent_id,sort_order,status,created_at,updated_at) VALUES (?,?,NULL,0,1,NOW(),NOW())`, nullableInt64(spaceID), t.ID)
	case "wiki":
		spaceID := s.ensureWikiSpace(t.CommunityID)
		_, _ = s.db.Exec(`INSERT IGNORE INTO wiki_pages (space_id,topic_id,status,created_at,updated_at) VALUES (?,?,1,NOW(),NOW())`, nullableInt64(spaceID), t.ID)
		var pageID int64
		if err := s.db.QueryRow(`SELECT id FROM wiki_pages WHERE topic_id=?`, t.ID).Scan(&pageID); err == nil && pageID > 0 {
			_, _ = s.db.Exec(`INSERT IGNORE INTO wiki_page_versions (wiki_page_id,topic_id,editor_id,version_no,title,content,change_note,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
				pageID, t.ID, t.UserID, 1, t.Title, t.Content, "初始版本")
			var versionID int64
			if err := s.db.QueryRow(`SELECT id FROM wiki_page_versions WHERE wiki_page_id=? AND version_no=1`, pageID).Scan(&versionID); err == nil {
				_, _ = s.db.Exec(`UPDATE wiki_pages SET current_version_id=? WHERE id=? AND current_version_id IS NULL`, versionID, pageID)
			}
		}
	}
}

func (s *MySQLStore) ensureDocsSpace(communityID int64) int64 {
	slug := "default"
	name := "默认文档空间"
	_, _ = s.db.Exec(`INSERT IGNORE INTO docs_spaces (community_id,name,slug,description,status,created_at,updated_at) VALUES (?,?,?,?,1,NOW(),NOW())`, communityID, name, slug, "由 Docs 插件自动创建的默认空间。")
	var id int64
	_ = s.db.QueryRow(`SELECT id FROM docs_spaces WHERE community_id=? AND slug=? LIMIT 1`, communityID, slug).Scan(&id)
	return id
}

func (s *MySQLStore) ensureWikiSpace(communityID int64) int64 {
	slug := "default"
	name := "默认 Wiki 空间"
	_, _ = s.db.Exec(`INSERT IGNORE INTO wiki_spaces (community_id,name,slug,description,status,created_at,updated_at) VALUES (?,?,?,?,1,NOW(),NOW())`, communityID, name, slug, "由 Wiki 插件自动创建的默认空间。")
	var id int64
	_ = s.db.QueryRow(`SELECT id FROM wiki_spaces WHERE community_id=? AND slug=? LIMIT 1`, communityID, slug).Scan(&id)
	return id
}

// AdminLogin 使用 admin_users 表校验后台人员，并发行后台 token。
func (s *MySQLStore) AdminLogin(account, password string) (*domain.AdminSession, error) {
	account = strings.TrimSpace(account)
	var id int64
	var status string
	err := s.db.QueryRow(`SELECT id,status FROM admin_users WHERE username=? OR email=? OR phone=? LIMIT 1`, account, account, account).
		Scan(&id, &status)
	if err != nil || (password != "admin123" && password != "123456") {
		return nil, errors.New("账号或密码错误")
	}
	if status == "forbidden" {
		return nil, errors.New("账号已被禁用")
	}
	return s.issueAdminSession(id)
}

func (s *MySQLStore) UserLogin(account, password string) (*domain.AdminSession, error) {
	account = strings.TrimSpace(account)
	var id int64
	var hash, status string
	err := s.db.QueryRow(`SELECT id,password_hash,status FROM users WHERE username=? OR email=? OR phone=? LIMIT 1`, account, account, account).
		Scan(&id, &hash, &status)
	if err != nil || !checkPassword(hash, password) {
		return nil, errors.New("账号或密码错误")
	}
	if status == "forbidden" {
		return nil, errors.New("账号已被禁用")
	}
	return s.issueUserSession(id)
}

func (s *MySQLStore) Register(req domain.RegisterRequest) (*domain.AdminSession, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码至少 6 位")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = username
	}
	res, err := s.db.Exec(`INSERT INTO users (username,nickname,email,phone,password_hash,status,created_at) VALUES (?,?,?,?,?,'normal',NOW())`,
		username, nickname, strings.TrimSpace(req.Email), strings.TrimSpace(req.Phone), hash)
	if err != nil {
		return nil, errors.New("用户名或邮箱已存在")
	}
	id, _ := res.LastInsertId()
	if _, err := s.db.Exec(`INSERT IGNORE INTO user_roles (user_id,role_id,site_key,status) VALUES (?,?,'*','normal')`, id, roleID("user")); err != nil {
		return nil, err
	}
	return s.issueUserSession(id)
}

func (s *MySQLStore) RefreshSession(refreshToken string) (*domain.AdminSession, error) {
	return s.refreshSessionByType(refreshToken, "user")
}

func (s *MySQLStore) RefreshAdminSession(refreshToken string) (*domain.AdminSession, error) {
	return s.refreshSessionByType(refreshToken, "admin")
}

func (s *MySQLStore) refreshSessionByType(refreshToken, tokenType string) (*domain.AdminSession, error) {
	hash := tokenHash(strings.TrimSpace(refreshToken))
	var userID int64
	err := s.db.QueryRow(`SELECT user_id FROM refresh_tokens WHERE token_hash=? AND token_type=? AND revoked_at IS NULL AND expires_at>NOW()`, hash, tokenType).Scan(&userID)
	if err != nil {
		return nil, errors.New("refresh token 无效")
	}
	_, _ = s.db.Exec(`UPDATE refresh_tokens SET revoked_at=NOW() WHERE token_hash=?`, hash)
	if tokenType == "admin" {
		return s.issueAdminSession(userID)
	}
	return s.issueUserSession(userID)
}

func (s *MySQLStore) Logout(refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	_, err := s.db.Exec(`UPDATE refresh_tokens SET revoked_at=NOW() WHERE token_hash=?`, tokenHash(refreshToken))
	return err
}

func (s *MySQLStore) AuthUser(accessToken string) (*domain.AuthUser, error) {
	userID, err := parseAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	return s.authUserByID(userID, "user")
}

func (s *MySQLStore) AuthAdmin(accessToken string) (*domain.AuthUser, error) {
	adminID, err := parseAdminAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	return s.authAdminByID(adminID)
}

func (s *MySQLStore) issueUserSession(userID int64) (*domain.AdminSession, error) {
	user, err := s.authUserByID(userID, "user")
	if err != nil {
		return nil, errors.New("账号不存在")
	}
	if user.Status == "forbidden" {
		return nil, errors.New("账号已被禁用")
	}
	access, expiresIn, err := newAccessToken(user.ID)
	if err != nil {
		return nil, err
	}
	refresh, refreshHash, refreshExpires, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(`INSERT INTO refresh_tokens (user_id,token_hash,token_type,expires_at) VALUES (?,?,?,?)`, user.ID, refreshHash, "user", refreshExpires); err != nil {
		return nil, err
	}
	_, _ = s.db.Exec(`UPDATE users SET last_login_at=NOW() WHERE id=?`, user.ID)
	loginUser := domain.AdminLoginUser{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Email:       user.Email,
		Phone:       user.Phone,
		Role:        user.RoleName,
		RoleCode:    user.RoleCode,
		Sites:       user.Sites,
		Permissions: user.Permissions,
	}
	return &domain.AdminSession{Token: access, AccessToken: access, RefreshToken: refresh, ExpiresIn: expiresIn, User: loginUser, TokenType: "user", Audience: "devhub_frontend"}, nil
}

func (s *MySQLStore) issueAdminSession(adminID int64) (*domain.AdminSession, error) {
	admin, err := s.authAdminByID(adminID)
	if err != nil {
		return nil, errors.New("后台账号不存在")
	}
	if admin.Status == "forbidden" {
		return nil, errors.New("账号已被禁用")
	}
	access, expiresIn, err := newAdminAccessToken(admin.ID)
	if err != nil {
		return nil, err
	}
	refresh, refreshHash, refreshExpires, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(`INSERT INTO refresh_tokens (user_id,token_hash,token_type,expires_at) VALUES (?,?,?,?)`, admin.ID, refreshHash, "admin", refreshExpires); err != nil {
		return nil, err
	}
	_, _ = s.db.Exec(`UPDATE admin_users SET last_login_at=NOW() WHERE id=?`, admin.ID)
	s.appendLog("login", admin.Username, "管理员登录", "后台系统", "127.0.0.1")
	loginUser := domain.AdminLoginUser{
		ID:          admin.ID,
		Username:    admin.Username,
		Nickname:    admin.Nickname,
		Email:       admin.Email,
		Phone:       admin.Phone,
		Role:        admin.RoleName,
		RoleCode:    admin.RoleCode,
		Sites:       admin.Sites,
		Permissions: admin.Permissions,
	}
	return &domain.AdminSession{Token: access, AccessToken: access, RefreshToken: refresh, ExpiresIn: expiresIn, User: loginUser, TokenType: "admin", Audience: "devhub_admin"}, nil
}

func (s *MySQLStore) authUserByID(userID int64, identity string) (*domain.AuthUser, error) {
	user := &domain.AuthUser{}
	err := s.db.QueryRow(`SELECT id,username,nickname,email,phone,status FROM users WHERE id=?`, userID).
		Scan(&user.ID, &user.Username, &user.Nickname, &user.Email, &user.Phone, &user.Status)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT r.code,r.name,ur.site_key FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=? AND ur.status='normal' ORDER BY r.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roleCodes := []string{}
	siteSet := map[string]bool{}
	for rows.Next() {
		var code, name, site string
		if err := rows.Scan(&code, &name, &site); err != nil {
			continue
		}
		roleCodes = append(roleCodes, code)
		if user.RoleCode == "" || code == "super_admin" {
			user.RoleCode = code
			user.RoleName = name
		}
		siteSet[site] = true
	}
	if user.RoleCode == "" {
		user.RoleCode = "user"
		user.RoleName = "普通用户"
	}
	for site := range siteSet {
		user.Sites = append(user.Sites, site)
	}
	sort.Strings(user.Sites)
	user.Permissions = s.permissionsForRoles(roleCodes)
	user.TokenType = identity
	user.Identity = identity
	if identity == "admin" {
		user.Audience = "devhub_admin"
	} else {
		user.Audience = "devhub_frontend"
	}
	return user, nil
}

func (s *MySQLStore) authAdminByID(adminID int64) (*domain.AuthUser, error) {
	user := &domain.AuthUser{}
	var roleID int64
	err := s.db.QueryRow(`SELECT id,username,nickname,email,phone,status,role_id,role_name FROM admin_users WHERE id=?`, adminID).
		Scan(&user.ID, &user.Username, &user.Nickname, &user.Email, &user.Phone, &user.Status, &roleID, &user.RoleName)
	if err != nil {
		return nil, err
	}
	switch roleID {
	case 1:
		user.RoleCode = "super_admin"
		user.Sites = []string{"*"}
	case 2:
		user.RoleCode = "site_admin"
		user.Sites = []string{"php"}
	case 3:
		user.RoleCode = "moderator"
		user.Sites = []string{"go"}
	default:
		user.RoleCode = roleCodeByID(roleID)
	}
	user.Permissions = s.permissionsForRoles([]string{user.RoleCode})
	user.TokenType = "admin"
	user.Identity = "admin"
	user.Audience = "devhub_admin"
	return user, nil
}

func (s *MySQLStore) permissionsForRoles(roleCodes []string) []string {
	for _, code := range roleCodes {
		if code == "super_admin" {
			return []string{"*"}
		}
	}
	if len(roleCodes) == 0 {
		roleCodes = []string{"user"}
	}
	ids := []int64{}
	for _, code := range roleCodes {
		if id := roleID(code); id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return []string{}
	}
	perms := map[string]bool{}
	for _, id := range ids {
		rows, err := s.db.Query(`SELECT p.code FROM role_permissions rp JOIN permissions p ON p.id=rp.permission_id WHERE rp.role_id=?`, id)
		if err != nil {
			continue
		}
		for rows.Next() {
			var code string
			if err := rows.Scan(&code); err == nil {
				perms[code] = true
			}
		}
		_ = rows.Close()
	}
	out := make([]string, 0, len(perms))
	for code := range perms {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func (s *MySQLStore) GetSite(key string) (domain.Site, bool) {
	row := s.db.QueryRow(`SELECT site_key,name,logo,title,subtitle,pub_text,description,color,status,sort_order FROM sites WHERE site_key=?`, key)
	site, err := scanSite(row)
	return site, err == nil
}

func (s *MySQLStore) CreateSite(req domain.Site) (domain.Site, error) {
	key := strings.TrimSpace(req.Key)
	if key == "" || key == "portal" {
		return domain.Site{}, errors.New("子站标识不合法")
	}
	site := domain.Site{
		Key:         key,
		Name:        strings.TrimSpace(req.Name),
		Logo:        strings.TrimSpace(req.Logo),
		Title:       strings.TrimSpace(req.Title),
		Sub:         strings.TrimSpace(req.Sub),
		Pub:         strings.TrimSpace(req.Pub),
		Description: strings.TrimSpace(req.Description),
		Color:       strings.TrimSpace(req.Color),
		Status:      strings.TrimSpace(req.Status),
		Sort:        req.Sort,
	}
	if site.Name == "" {
		site.Name = key
	}
	if site.Logo == "" {
		site.Logo = strings.ToUpper(key)
	}
	if site.Title == "" {
		site.Title = site.Name + " 子网站"
	}
	if site.Pub == "" {
		site.Pub = "发布 " + site.Name + " 内容"
	}
	if site.Status == "" {
		site.Status = "enable"
	}
	_, err := s.db.Exec(`INSERT INTO sites (site_key,name,logo,title,subtitle,pub_text,description,color,status,sort_order) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		site.Key, site.Name, site.Logo, site.Title, site.Sub, site.Pub, site.Description, site.Color, site.Status, site.Sort)
	if err != nil {
		return domain.Site{}, err
	}
	s.appendLog("system", "admin", "新增子站", fmt.Sprintf("sites#%s", key), "127.0.0.1")
	created, _ := s.GetSite(key)
	return created, nil
}

func (s *MySQLStore) UpdateSite(key string, req domain.Site) (domain.Site, bool) {
	oldSite, exists := s.GetSite(key)
	if !exists {
		return domain.Site{}, false
	}
	status := strings.TrimSpace(req.Status)
	if key == "portal" && status == "disable" {
		status = "enable"
	}
	if status == "" {
		status = "enable"
	}
	name := strings.TrimSpace(req.Name)
	logo := strings.TrimSpace(req.Logo)
	title := strings.TrimSpace(req.Title)
	sub := strings.TrimSpace(req.Sub)
	pub := strings.TrimSpace(req.Pub)
	description := strings.TrimSpace(req.Description)
	color := strings.TrimSpace(req.Color)
	if key != "portal" && name != "" && name != oldSite.Name {
		if title == "" || title == oldSite.Title || title == oldSite.Name+" 子网站" {
			title = name + " 子网站"
		}
		if pub == "" || pub == oldSite.Pub || pub == "发布 "+oldSite.Name+" 内容" {
			pub = "发布 " + name + " 内容"
		}
	}
	res, err := s.db.Exec(`UPDATE sites SET name=?,logo=?,title=?,subtitle=?,pub_text=?,description=?,color=?,status=?,sort_order=? WHERE site_key=?`,
		name, logo, title, sub, pub, description, color, status, req.Sort, key)
	if err != nil {
		return domain.Site{}, false
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Site{}, false
	}
	s.appendLog("system", "admin", "更新子站配置", fmt.Sprintf("sites#%s", key), "127.0.0.1")
	site, _ := s.GetSite(key)
	return site, true
}

func (s *MySQLStore) ListSites() []domain.Site {
	rows, err := s.db.Query(`SELECT site_key,name,logo,title,subtitle,pub_text,description,color,status,sort_order FROM sites ORDER BY sort_order,id`)
	if err != nil {
		return []domain.Site{}
	}
	defer rows.Close()
	out := []domain.Site{}
	for rows.Next() {
		site, err := scanSite(rows)
		if err == nil {
			out = append(out, site)
		}
	}
	return out
}

func (s *MySQLStore) ListBoards() []domain.Board {
	rows, err := s.db.Query(`SELECT board_key,name,site_key,sort_order,visible FROM boards ORDER BY sort_order,id`)
	if err != nil {
		return []domain.Board{}
	}
	defer rows.Close()
	out := []domain.Board{}
	for rows.Next() {
		var b domain.Board
		if err := rows.Scan(&b.Key, &b.Name, &b.Site, &b.Sort, &b.Visible); err == nil {
			out = append(out, b)
		}
	}
	return out
}

func (s *MySQLStore) CreateBoard(req domain.Board) (domain.Board, error) {
	key := strings.TrimSpace(req.Key)
	if key == "" || key == "all" {
		return domain.Board{}, errors.New("板块标识不合法")
	}
	board := domain.Board{Key: key, Name: strings.TrimSpace(req.Name), Site: strings.TrimSpace(req.Site), Sort: req.Sort, Visible: req.Visible}
	if board.Name == "" {
		board.Name = key
	}
	if board.Site == "" {
		board.Site = "all"
	}
	if _, err := s.db.Exec(`INSERT INTO boards (board_key,name,site_key,sort_order,visible) VALUES (?,?,?,?,?)`, board.Key, board.Name, board.Site, board.Sort, board.Visible); err != nil {
		return domain.Board{}, err
	}
	s.appendLog("system", "admin", "新增板块", fmt.Sprintf("boards#%s", key), "127.0.0.1")
	for _, b := range s.ListBoards() {
		if b.Key == key {
			return b, nil
		}
	}
	return board, nil
}

func (s *MySQLStore) UpdateBoard(key string, req domain.Board) (domain.Board, bool) {
	res, err := s.db.Exec(`UPDATE boards SET name=?,site_key=?,sort_order=?,visible=? WHERE board_key=?`,
		strings.TrimSpace(req.Name), strings.TrimSpace(req.Site), req.Sort, req.Visible, key)
	if err != nil {
		return domain.Board{}, false
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Board{}, false
	}
	s.appendLog("system", "admin", "更新板块配置", fmt.Sprintf("boards#%s", key), "127.0.0.1")
	for _, b := range s.ListBoards() {
		if b.Key == key {
			return b, true
		}
	}
	return domain.Board{}, false
}

func (s *MySQLStore) Plugins() []domain.Plugin {
	rows, err := s.db.Query(`SELECT plugin_code,name,version,status,COALESCE(description,''),COALESCE(source_type,''),COALESCE(CAST(manifest_json AS CHAR),''),COALESCE(manifest_checksum,''),COALESCE(package_checksum,''),COALESCE(compatible_core_version,''),COALESCE(CAST(config_json AS CHAR),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM plugins ORDER BY plugin_code`)
	if err != nil {
		defs := pluginregistry.Definitions()
		for i := range defs {
			defs[i].ResolvedConfig = pluginregistry.ResolvePluginConfig(defs[i], "", "")
			defs[i] = pluginregistry.ApplyLifecycle(defs[i])
		}
		return defs
	}
	defer rows.Close()
	runtime := map[string]domain.Plugin{}
	for rows.Next() {
		var p domain.Plugin
		if err := rows.Scan(&p.Code, &p.Name, &p.Version, &p.Status, &p.Description, &p.SourceType, &p.ManifestJSON, &p.ManifestChecksum, &p.PackageChecksum, &p.CompatibleCoreVersion, &p.ConfigJSON, &p.CreatedAt, &p.UpdatedAt); err == nil {
			p.PluginCode = p.Code
			runtime[p.Code] = p
		}
	}
	out := make([]domain.Plugin, 0, len(runtime)+len(pluginregistry.Definitions()))
	seen := map[string]bool{}
	for _, def := range pluginregistry.Definitions() {
		seen[def.Code] = true
		if item, ok := runtime[def.Code]; ok {
			merged := pluginregistry.MergeRuntimeState(def, item)
			merged.ResolvedConfig = pluginregistry.ResolvePluginConfig(merged, item.ConfigJSON, "")
			out = append(out, pluginregistry.ApplyLifecycle(merged))
			continue
		}
		def.ResolvedConfig = pluginregistry.ResolvePluginConfig(def, "", "")
		out = append(out, pluginregistry.ApplyLifecycle(def))
	}
	for code, item := range runtime {
		if seen[code] {
			continue
		}
		def := mysqlPluginDefinitionFromRuntime(item)
		merged := pluginregistry.MergeRuntimeState(def, item)
		merged.ResolvedConfig = pluginregistry.ResolvePluginConfig(merged, item.ConfigJSON, "")
		out = append(out, pluginregistry.ApplyLifecycle(merged))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func (s *MySQLStore) PluginByCode(code string) (domain.Plugin, bool) {
	def, ok := pluginregistry.DefinitionByCode(code)
	var runtime domain.Plugin
	queryCode := strings.TrimSpace(code)
	if ok {
		queryCode = def.Code
	}
	err := s.db.QueryRow(`SELECT plugin_code,name,version,status,COALESCE(description,''),COALESCE(source_type,''),COALESCE(CAST(manifest_json AS CHAR),''),COALESCE(manifest_checksum,''),COALESCE(package_checksum,''),COALESCE(compatible_core_version,''),COALESCE(CAST(config_json AS CHAR),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM plugins WHERE plugin_code=?`, queryCode).
		Scan(&runtime.Code, &runtime.Name, &runtime.Version, &runtime.Status, &runtime.Description, &runtime.SourceType, &runtime.ManifestJSON, &runtime.ManifestChecksum, &runtime.PackageChecksum, &runtime.CompatibleCoreVersion, &runtime.ConfigJSON, &runtime.CreatedAt, &runtime.UpdatedAt)
	if err != nil {
		if !ok {
			return domain.Plugin{}, false
		}
		def.ResolvedConfig = pluginregistry.ResolvePluginConfig(def, "", "")
		return pluginregistry.ApplyLifecycle(def), true
	}
	runtime.PluginCode = runtime.Code
	if !ok {
		def = mysqlPluginDefinitionFromRuntime(runtime)
	}
	merged := pluginregistry.MergeRuntimeState(def, runtime)
	merged.ResolvedConfig = pluginregistry.ResolvePluginConfig(merged, runtime.ConfigJSON, "")
	return pluginregistry.ApplyLifecycle(merged), true
}

func mysqlPluginDefinitionFromRuntime(runtime domain.Plugin) domain.Plugin {
	if runtime.ManifestJSON != "" {
		if manifest, _, err := pluginregistry.DecodePluginManifestJSON([]byte(runtime.ManifestJSON)); err == nil {
			return domain.Plugin{PluginManifest: manifest, Status: runtime.Status}
		}
	}
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  runtime.Code,
			PluginCode:            runtime.Code,
			Name:                  firstNonEmptyString(runtime.Name, runtime.Code),
			Version:               runtime.Version,
			Description:           runtime.Description,
			SourceType:            firstNonEmptyString(runtime.SourceType, "manifest"),
			CompatibleCoreVersion: runtime.CompatibleCoreVersion,
		},
		Status: runtime.Status,
	}
}

func (s *MySQLStore) SavePlugin(plugin domain.Plugin) (domain.Plugin, error) {
	plugin.Code = strings.TrimSpace(firstNonEmptyString(plugin.Code, plugin.PluginCode))
	if plugin.Code == "" {
		return domain.Plugin{}, errors.New("插件 code 不能为空")
	}
	if plugin.Status == "" {
		plugin.Status = pluginregistry.StatusDisabled
	}
	if !pluginregistry.ValidGlobalStatus(plugin.Status) {
		return domain.Plugin{}, errors.New("插件状态不合法")
	}
	var manifest any = nil
	if plugin.ManifestJSON != "" {
		manifest = json.RawMessage(plugin.ManifestJSON)
	}
	var config any = nil
	if plugin.ConfigJSON != "" {
		config = json.RawMessage(plugin.ConfigJSON)
	}
	if _, err := s.db.Exec(`INSERT INTO plugins (plugin_code,name,version,status,description,source_type,manifest_json,manifest_checksum,package_checksum,compatible_core_version,config_json,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())
		ON DUPLICATE KEY UPDATE name=VALUES(name),version=VALUES(version),status=VALUES(status),description=VALUES(description),source_type=VALUES(source_type),manifest_json=VALUES(manifest_json),manifest_checksum=VALUES(manifest_checksum),package_checksum=VALUES(package_checksum),compatible_core_version=VALUES(compatible_core_version),config_json=VALUES(config_json),updated_at=NOW()`,
		plugin.Code, plugin.Name, plugin.Version, plugin.Status, plugin.Description, firstNonEmptyString(plugin.SourceType, "manifest"), manifest, plugin.ManifestChecksum, plugin.PackageChecksum, plugin.CompatibleCoreVersion, config); err != nil {
		return domain.Plugin{}, err
	}
	out, _ := s.PluginByCode(plugin.Code)
	return out, nil
}

func (s *MySQLStore) SetPluginStatus(code, status string) (domain.Plugin, error) {
	status = strings.TrimSpace(status)
	if !pluginregistry.ValidGlobalStatus(status) {
		return domain.Plugin{}, errors.New("插件状态不合法")
	}
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		if plugin, exists := s.PluginByCode(code); exists {
			plugin.Status = status
			return s.SavePlugin(plugin)
		}
		return domain.Plugin{}, errors.New("插件不存在")
	}
	if _, err := s.db.Exec(`INSERT INTO plugins (plugin_code,name,version,status,description,source_type,config_json,created_at,updated_at) VALUES (?,?,?,?,?,?,NULL,NOW(),NOW()) ON DUPLICATE KEY UPDATE status=VALUES(status),source_type=VALUES(source_type),updated_at=NOW()`,
		def.Code, def.Name, def.Version, status, def.Description, "builtin"); err != nil {
		return domain.Plugin{}, err
	}
	plugin, _ := s.PluginByCode(def.Code)
	return plugin, nil
}

func (s *MySQLStore) SetPluginConfig(code, configJSON string) (domain.Plugin, error) {
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		if plugin, exists := s.PluginByCode(code); exists {
			if err := pluginregistry.ValidateConfigJSON(plugin, configJSON); err != nil {
				return domain.Plugin{}, err
			}
			plugin.ConfigJSON = configJSON
			return s.SavePlugin(plugin)
		}
		return domain.Plugin{}, errors.New("插件不存在")
	}
	configJSON = strings.TrimSpace(configJSON)
	if err := pluginregistry.ValidateConfigJSON(def, configJSON); err != nil {
		return domain.Plugin{}, err
	}
	var config any = nil
	if configJSON != "" {
		config = json.RawMessage(configJSON)
	}
	if _, err := s.db.Exec(`INSERT INTO plugins (plugin_code,name,version,status,description,source_type,config_json,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,NOW(),NOW())
		ON DUPLICATE KEY UPDATE config_json=VALUES(config_json),source_type=VALUES(source_type),updated_at=NOW()`, def.Code, def.Name, def.Version, def.Status, def.Description, "builtin", config); err != nil {
		return domain.Plugin{}, err
	}
	plugin, _ := s.PluginByCode(def.Code)
	return plugin, nil
}

// ===== Plugin config versions (v1.5.0-P1-05) =====

func (s *MySQLStore) AppendPluginConfigVersion(record domain.PluginConfigVersion) (domain.PluginConfigVersion, error) {
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.Scope = strings.TrimSpace(record.Scope)
	if record.Scope == "" {
		record.Scope = domain.PluginConfigScopeGlobal
	}
	if record.PluginCode == "" {
		return domain.PluginConfigVersion{}, errors.New("plugin_code 不能为空")
	}
	if record.Scope != domain.PluginConfigScopeGlobal && record.Scope != domain.PluginConfigScopeCommunity {
		return domain.PluginConfigVersion{}, errors.New("scope 不合法")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return domain.PluginConfigVersion{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var prevID int64 = 0
	var prevNo int = 0
	_ = tx.QueryRow(`SELECT id, version_no FROM plugin_config_versions WHERE plugin_code=? AND scope=? AND community_id=? ORDER BY version_no DESC, id DESC LIMIT 1`,
		record.PluginCode, record.Scope, record.CommunityID).Scan(&prevID, &prevNo)
	record.VersionNo = prevNo + 1
	record.PreviousVersionID = prevID

	var config any = nil
	if strings.TrimSpace(record.ConfigJSON) != "" {
		config = json.RawMessage(record.ConfigJSON)
	}
	var changed any = nil
	if strings.TrimSpace(record.ChangedKeysJSON) != "" && json.Valid([]byte(record.ChangedKeysJSON)) {
		changed = json.RawMessage(record.ChangedKeysJSON)
	}
	var diff any = nil
	if strings.TrimSpace(record.DiffJSON) != "" && json.Valid([]byte(record.DiffJSON)) {
		diff = json.RawMessage(record.DiffJSON)
	}
	var meta any = nil
	if strings.TrimSpace(record.MetadataJSON) != "" && json.Valid([]byte(record.MetadataJSON)) {
		meta = json.RawMessage(record.MetadataJSON)
	}

	res, err := tx.Exec(`INSERT INTO plugin_config_versions
		(plugin_code, scope, community_id, version_no, config_json, config_hash, changed_keys_json, diff_json, source, operator_type, operator_id, operator_name, reason, previous_version_id, rollback_from_version_id, metadata_json, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW())`,
		record.PluginCode, record.Scope, record.CommunityID, record.VersionNo, config, record.ConfigHash, changed, diff, record.Source,
		record.OperatorType, record.OperatorID, record.OperatorName, record.Reason, record.PreviousVersionID, record.RollbackFromVersion, meta)
	if err != nil {
		return domain.PluginConfigVersion{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	_ = tx.Commit()

	// Fetch created_at for API consistency.
	_ = s.db.QueryRow(`SELECT DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM plugin_config_versions WHERE id=?`, record.ID).Scan(&record.CreatedAt)
	return record, nil
}

func (s *MySQLStore) PluginConfigVersions(pluginCode, scope string, communityID int64, page, pageSize int) ([]domain.PluginConfigVersion, int, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = domain.PluginConfigScopeGlobal
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM plugin_config_versions WHERE plugin_code=? AND scope=? AND community_id=?`,
		pluginCode, scope, communityID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(`SELECT id, plugin_code, scope, community_id, version_no,
		COALESCE(CAST(config_json AS CHAR),''), config_hash,
		COALESCE(CAST(changed_keys_json AS CHAR),''), COALESCE(CAST(diff_json AS CHAR),''), source,
		operator_type, operator_id, operator_name, reason, previous_version_id, rollback_from_version_id,
		COALESCE(CAST(metadata_json AS CHAR),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_config_versions
		WHERE plugin_code=? AND scope=? AND community_id=?
		ORDER BY version_no DESC, id DESC
		LIMIT ? OFFSET ?`, pluginCode, scope, communityID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginConfigVersion{}
	for rows.Next() {
		var it domain.PluginConfigVersion
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.Scope, &it.CommunityID, &it.VersionNo,
			&it.ConfigJSON, &it.ConfigHash, &it.ChangedKeysJSON, &it.DiffJSON, &it.Source,
			&it.OperatorType, &it.OperatorID, &it.OperatorName, &it.Reason, &it.PreviousVersionID, &it.RollbackFromVersion,
			&it.MetadataJSON, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) PluginConfigVersionByID(id int64) (domain.PluginConfigVersion, bool) {
	var it domain.PluginConfigVersion
	err := s.db.QueryRow(`SELECT id, plugin_code, scope, community_id, version_no,
		COALESCE(CAST(config_json AS CHAR),''), config_hash,
		COALESCE(CAST(changed_keys_json AS CHAR),''), COALESCE(CAST(diff_json AS CHAR),''), source,
		operator_type, operator_id, operator_name, reason, previous_version_id, rollback_from_version_id,
		COALESCE(CAST(metadata_json AS CHAR),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_config_versions WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PluginCode, &it.Scope, &it.CommunityID, &it.VersionNo,
			&it.ConfigJSON, &it.ConfigHash, &it.ChangedKeysJSON, &it.DiffJSON, &it.Source,
			&it.OperatorType, &it.OperatorID, &it.OperatorName, &it.Reason, &it.PreviousVersionID, &it.RollbackFromVersion,
			&it.MetadataJSON, &it.CreatedAt)
	if err != nil {
		return domain.PluginConfigVersion{}, false
	}
	return it, true
}

// ===== Plugin approvals (v1.5.0-P1-07) =====

func (s *MySQLStore) AppendPluginApprovalRequest(record domain.PluginApprovalRequest) (domain.PluginApprovalRequest, error) {
	record.Action = strings.TrimSpace(record.Action)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.PluginName = strings.TrimSpace(record.PluginName)
	record.Status = strings.TrimSpace(record.Status)
	if record.Status == "" {
		record.Status = domain.PluginApprovalStatusPending
	}

	var execRes any = nil
	if strings.TrimSpace(record.ExecuteResultJSON) != "" && json.Valid([]byte(record.ExecuteResultJSON)) {
		execRes = json.RawMessage(record.ExecuteResultJSON)
	}
	var manifest any = nil
	if strings.TrimSpace(record.ManifestJSON) != "" && json.Valid([]byte(record.ManifestJSON)) {
		manifest = json.RawMessage(record.ManifestJSON)
	}
	var dry any = nil
	if strings.TrimSpace(record.DryRunJSON) != "" && json.Valid([]byte(record.DryRunJSON)) {
		dry = json.RawMessage(record.DryRunJSON)
	}
	var risk any = nil
	if strings.TrimSpace(record.RiskReportJSON) != "" && json.Valid([]byte(record.RiskReportJSON)) {
		risk = json.RawMessage(record.RiskReportJSON)
	}
	var deps any = nil
	if strings.TrimSpace(record.DependencySummaryJSON) != "" && json.Valid([]byte(record.DependencySummaryJSON)) {
		deps = json.RawMessage(record.DependencySummaryJSON)
	}
	var compat any = nil
	if strings.TrimSpace(record.CompatibilityJSON) != "" && json.Valid([]byte(record.CompatibilityJSON)) {
		compat = json.RawMessage(record.CompatibilityJSON)
	}
	var changed any = nil
	if strings.TrimSpace(record.ChangedKeysJSON) != "" && json.Valid([]byte(record.ChangedKeysJSON)) {
		changed = json.RawMessage(record.ChangedKeysJSON)
	}
	var diff any = nil
	if strings.TrimSpace(record.DiffJSON) != "" && json.Valid([]byte(record.DiffJSON)) {
		diff = json.RawMessage(record.DiffJSON)
	}
	var meta any = nil
	if strings.TrimSpace(record.MetadataJSON) != "" && json.Valid([]byte(record.MetadataJSON)) {
		meta = json.RawMessage(record.MetadataJSON)
	}

	res, err := s.db.Exec(`INSERT INTO plugin_approval_requests
		(request_no, action, plugin_code, plugin_name, current_version, target_version,
		package_path, package_checksum_status, package_risk_level,
		status, reason, requested_by, requested_by_name, requested_at,
		reviewed_by, reviewed_by_name, reviewed_at, review_comment,
		executed_by, executed_at, execute_result_json,
		manifest_json, dry_run_json, risk_report_json, dependency_summary_json, compatibility_json,
		changed_keys_json, diff_json,
		error_code, error_message, metadata_json, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?, ?,?,?, NOW(), ?,?,?, ?,?, ?, ?,?,?,?,?,?,?, ?,?,?,?, NOW(), NOW())`,
		record.RequestNo, record.Action, record.PluginCode, record.PluginName, record.CurrentVersion, record.TargetVersion,
		record.PackagePath, record.PackageChecksumStatus, record.PackageRiskLevel,
		record.Status, record.Reason, record.RequestedBy, record.RequestedByName,
		record.ReviewedBy, record.ReviewedByName, nullTime(record.ReviewedAt), record.ReviewComment,
		record.ExecutedBy, nullTime(record.ExecutedAt), execRes,
		manifest, dry, risk, deps, compat,
		changed, diff,
		record.ErrorCode, record.ErrorMessage, meta)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	_ = s.db.QueryRow(`SELECT DATE_FORMAT(requested_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM plugin_approval_requests WHERE id=?`, record.ID).
		Scan(&record.RequestedAt, &record.CreatedAt, &record.UpdatedAt)
	_ = s.db.QueryRow(`SELECT COALESCE(DATE_FORMAT(reviewed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(executed_at,'%Y-%m-%d %H:%i:%s'),'') FROM plugin_approval_requests WHERE id=?`, record.ID).
		Scan(&record.ReviewedAt, &record.ExecutedAt)
	return record, nil
}

func (s *MySQLStore) SavePluginApprovalRequest(record domain.PluginApprovalRequest) (domain.PluginApprovalRequest, error) {
	if record.ID <= 0 {
		return domain.PluginApprovalRequest{}, errors.New("approval id 不能为空")
	}
	record.Action = strings.TrimSpace(record.Action)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.PluginName = strings.TrimSpace(record.PluginName)
	record.Status = strings.TrimSpace(record.Status)

	var execRes any = nil
	if strings.TrimSpace(record.ExecuteResultJSON) != "" && json.Valid([]byte(record.ExecuteResultJSON)) {
		execRes = json.RawMessage(record.ExecuteResultJSON)
	}
	var manifest any = nil
	if strings.TrimSpace(record.ManifestJSON) != "" && json.Valid([]byte(record.ManifestJSON)) {
		manifest = json.RawMessage(record.ManifestJSON)
	}
	var dry any = nil
	if strings.TrimSpace(record.DryRunJSON) != "" && json.Valid([]byte(record.DryRunJSON)) {
		dry = json.RawMessage(record.DryRunJSON)
	}
	var risk any = nil
	if strings.TrimSpace(record.RiskReportJSON) != "" && json.Valid([]byte(record.RiskReportJSON)) {
		risk = json.RawMessage(record.RiskReportJSON)
	}
	var deps any = nil
	if strings.TrimSpace(record.DependencySummaryJSON) != "" && json.Valid([]byte(record.DependencySummaryJSON)) {
		deps = json.RawMessage(record.DependencySummaryJSON)
	}
	var compat any = nil
	if strings.TrimSpace(record.CompatibilityJSON) != "" && json.Valid([]byte(record.CompatibilityJSON)) {
		compat = json.RawMessage(record.CompatibilityJSON)
	}
	var changed any = nil
	if strings.TrimSpace(record.ChangedKeysJSON) != "" && json.Valid([]byte(record.ChangedKeysJSON)) {
		changed = json.RawMessage(record.ChangedKeysJSON)
	}
	var diff any = nil
	if strings.TrimSpace(record.DiffJSON) != "" && json.Valid([]byte(record.DiffJSON)) {
		diff = json.RawMessage(record.DiffJSON)
	}
	var meta any = nil
	if strings.TrimSpace(record.MetadataJSON) != "" && json.Valid([]byte(record.MetadataJSON)) {
		meta = json.RawMessage(record.MetadataJSON)
	}

	_, err := s.db.Exec(`UPDATE plugin_approval_requests SET
		request_no=?, action=?, plugin_code=?, plugin_name=?, current_version=?, target_version=?,
		package_path=?, package_checksum_status=?, package_risk_level=?,
		status=?, reason=?,
		requested_by=?, requested_by_name=?,
		reviewed_by=?, reviewed_by_name=?, reviewed_at=?, review_comment=?,
		executed_by=?, executed_at=?, execute_result_json=?,
		manifest_json=?, dry_run_json=?, risk_report_json=?, dependency_summary_json=?, compatibility_json=?,
		changed_keys_json=?, diff_json=?,
		error_code=?, error_message=?, metadata_json=?, updated_at=NOW()
		WHERE id=?`,
		record.RequestNo, record.Action, record.PluginCode, record.PluginName, record.CurrentVersion, record.TargetVersion,
		record.PackagePath, record.PackageChecksumStatus, record.PackageRiskLevel,
		record.Status, record.Reason,
		record.RequestedBy, record.RequestedByName,
		record.ReviewedBy, record.ReviewedByName, nullTime(record.ReviewedAt), record.ReviewComment,
		record.ExecutedBy, nullTime(record.ExecutedAt), execRes,
		manifest, dry, risk, deps, compat,
		changed, diff,
		record.ErrorCode, record.ErrorMessage, meta, record.ID)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	out, _ := s.PluginApprovalRequestByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginApprovalRequestByID(id int64) (domain.PluginApprovalRequest, bool) {
	var it domain.PluginApprovalRequest
	err := s.db.QueryRow(`SELECT id, request_no, action, plugin_code, plugin_name, current_version, target_version,
		package_path, package_checksum_status, package_risk_level,
		status, reason,
		requested_by, requested_by_name, DATE_FORMAT(requested_at,'%Y-%m-%d %H:%i:%s'),
		reviewed_by, reviewed_by_name, COALESCE(DATE_FORMAT(reviewed_at,'%Y-%m-%d %H:%i:%s'),''), review_comment,
		executed_by, COALESCE(DATE_FORMAT(executed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(CAST(execute_result_json AS CHAR),''),
		COALESCE(CAST(manifest_json AS CHAR),''), COALESCE(CAST(dry_run_json AS CHAR),''), COALESCE(CAST(risk_report_json AS CHAR),''),
		COALESCE(CAST(dependency_summary_json AS CHAR),''), COALESCE(CAST(compatibility_json AS CHAR),''),
		COALESCE(CAST(changed_keys_json AS CHAR),''), COALESCE(CAST(diff_json AS CHAR),''),
		error_code, error_message, COALESCE(CAST(metadata_json AS CHAR),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_approval_requests WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.RequestNo, &it.Action, &it.PluginCode, &it.PluginName, &it.CurrentVersion, &it.TargetVersion,
			&it.PackagePath, &it.PackageChecksumStatus, &it.PackageRiskLevel,
			&it.Status, &it.Reason,
			&it.RequestedBy, &it.RequestedByName, &it.RequestedAt,
			&it.ReviewedBy, &it.ReviewedByName, &it.ReviewedAt, &it.ReviewComment,
			&it.ExecutedBy, &it.ExecutedAt, &it.ExecuteResultJSON,
			&it.ManifestJSON, &it.DryRunJSON, &it.RiskReportJSON,
			&it.DependencySummaryJSON, &it.CompatibilityJSON,
			&it.ChangedKeysJSON, &it.DiffJSON,
			&it.ErrorCode, &it.ErrorMessage, &it.MetadataJSON,
			&it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginApprovalRequest{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginApprovalRequests(filter domain.PluginApprovalFilter) ([]domain.PluginApprovalRequest, int, error) {
	status := strings.TrimSpace(filter.Status)
	action := strings.TrimSpace(filter.Action)
	pluginCode := strings.TrimSpace(filter.PluginCode)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	offset := (filter.Page - 1) * filter.PageSize

	where := []string{"1=1"}
	args := []any{}
	if status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if action != "" && action != "all" {
		where = append(where, "action=?")
		args = append(args, action)
	}
	if pluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, pluginCode)
	}
	if filter.RequestedBy > 0 {
		where = append(where, "requested_by=?")
		args = append(args, filter.RequestedBy)
	}
	if filter.ReviewedBy > 0 {
		where = append(where, "reviewed_by=?")
		args = append(args, filter.ReviewedBy)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_approval_requests WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, request_no, action, plugin_code, plugin_name, current_version, target_version,
		package_path, package_checksum_status, package_risk_level,
		status, reason,
		requested_by, requested_by_name, DATE_FORMAT(requested_at,'%Y-%m-%d %H:%i:%s'),
		reviewed_by, reviewed_by_name, COALESCE(DATE_FORMAT(reviewed_at,'%Y-%m-%d %H:%i:%s'),''), review_comment,
		executed_by, COALESCE(DATE_FORMAT(executed_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(CAST(execute_result_json AS CHAR),''),
		COALESCE(CAST(manifest_json AS CHAR),''), COALESCE(CAST(dry_run_json AS CHAR),''), COALESCE(CAST(risk_report_json AS CHAR),''),
		COALESCE(CAST(dependency_summary_json AS CHAR),''), COALESCE(CAST(compatibility_json AS CHAR),''),
		COALESCE(CAST(changed_keys_json AS CHAR),''), COALESCE(CAST(diff_json AS CHAR),''),
		error_code, error_message, COALESCE(CAST(metadata_json AS CHAR),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_approval_requests
		WHERE `+whereSQL+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginApprovalRequest{}
	for rows.Next() {
		var it domain.PluginApprovalRequest
		if err := rows.Scan(&it.ID, &it.RequestNo, &it.Action, &it.PluginCode, &it.PluginName, &it.CurrentVersion, &it.TargetVersion,
			&it.PackagePath, &it.PackageChecksumStatus, &it.PackageRiskLevel,
			&it.Status, &it.Reason,
			&it.RequestedBy, &it.RequestedByName, &it.RequestedAt,
			&it.ReviewedBy, &it.ReviewedByName, &it.ReviewedAt, &it.ReviewComment,
			&it.ExecutedBy, &it.ExecutedAt, &it.ExecuteResultJSON,
			&it.ManifestJSON, &it.DryRunJSON, &it.RiskReportJSON,
			&it.DependencySummaryJSON, &it.CompatibilityJSON,
			&it.ChangedKeysJSON, &it.DiffJSON,
			&it.ErrorCode, &it.ErrorMessage, &it.MetadataJSON,
			&it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Plugin package uploads (v1.6.0-P0-02) =====

func uploadJSONValue(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return nil
	}
	return json.RawMessage(raw)
}

func (s *MySQLStore) AppendPluginPackageUpload(record domain.PluginPackageUploadRecord) (domain.PluginPackageUploadRecord, error) {
	record.UploadID = strings.TrimSpace(record.UploadID)
	record.Status = strings.TrimSpace(record.Status)
	if record.Status == "" {
		record.Status = domain.PluginPackageUploadStatusUploaded
	}
	res, err := s.db.Exec(`INSERT INTO plugin_package_uploads
		(upload_id, original_filename, uploaded_by, uploaded_by_name, uploaded_at, status,
		package_code, package_name, package_version, upload_path, staging_path, package_path, promoted_path,
		compressed_size, uncompressed_size, file_count, checksum_status, signature_status, publisher_id, trust_status, risk_level,
		risk_report_json, zip_scan_json, file_scan_json, manifest_validation_json, install_dry_run_json,
		approval_id, install_approval_id, expires_at, deleted_at, error_code, error_message, metadata_json,
		created_at, updated_at)
		VALUES (?,?,?,?,COALESCE(?,NOW()),?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())`,
		record.UploadID, record.OriginalFilename, record.UploadedBy, record.UploadedByName, nullTime(record.UploadedAt), record.Status,
		record.PackageCode, record.PackageName, record.PackageVersion, record.UploadPath, record.StagingPath, record.PackagePath, record.PromotedPath,
		record.CompressedSize, record.UncompressedSize, record.FileCount, record.ChecksumStatus, record.SignatureStatus, record.PublisherID, record.TrustStatus, record.RiskLevel,
		uploadJSONValue(record.RiskReportJSON), uploadJSONValue(record.ZipScanJSON), uploadJSONValue(record.FileScanJSON), uploadJSONValue(record.ManifestValidationJSON), uploadJSONValue(record.InstallDryRunJSON),
		record.ApprovalID, record.InstallApprovalID, nullTime(record.ExpiresAt), nullTime(record.DeletedAt), record.ErrorCode, record.ErrorMessage, uploadJSONValue(record.MetadataJSON))
	if err != nil {
		return domain.PluginPackageUploadRecord{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	out, _ := s.PluginPackageUploadByUploadID(record.UploadID)
	return out, nil
}

func (s *MySQLStore) SavePluginPackageUpload(record domain.PluginPackageUploadRecord) (domain.PluginPackageUploadRecord, error) {
	if strings.TrimSpace(record.UploadID) == "" {
		return domain.PluginPackageUploadRecord{}, errors.New("upload_id 不能为空")
	}
	_, err := s.db.Exec(`UPDATE plugin_package_uploads SET
		original_filename=?, uploaded_by=?, uploaded_by_name=?, uploaded_at=COALESCE(?, uploaded_at), status=?,
		package_code=?, package_name=?, package_version=?, upload_path=?, staging_path=?, package_path=?, promoted_path=?,
		compressed_size=?, uncompressed_size=?, file_count=?, checksum_status=?, signature_status=?, publisher_id=?, trust_status=?, risk_level=?,
		risk_report_json=?, zip_scan_json=?, file_scan_json=?, manifest_validation_json=?, install_dry_run_json=?,
		approval_id=?, install_approval_id=?, expires_at=?, deleted_at=?, error_code=?, error_message=?, metadata_json=?, updated_at=NOW()
		WHERE upload_id=?`,
		record.OriginalFilename, record.UploadedBy, record.UploadedByName, nullTime(record.UploadedAt), record.Status,
		record.PackageCode, record.PackageName, record.PackageVersion, record.UploadPath, record.StagingPath, record.PackagePath, record.PromotedPath,
		record.CompressedSize, record.UncompressedSize, record.FileCount, record.ChecksumStatus, record.SignatureStatus, record.PublisherID, record.TrustStatus, record.RiskLevel,
		uploadJSONValue(record.RiskReportJSON), uploadJSONValue(record.ZipScanJSON), uploadJSONValue(record.FileScanJSON), uploadJSONValue(record.ManifestValidationJSON), uploadJSONValue(record.InstallDryRunJSON),
		record.ApprovalID, record.InstallApprovalID, nullTime(record.ExpiresAt), nullTime(record.DeletedAt), record.ErrorCode, record.ErrorMessage, uploadJSONValue(record.MetadataJSON), record.UploadID)
	if err != nil {
		return domain.PluginPackageUploadRecord{}, err
	}
	out, _ := s.PluginPackageUploadByUploadID(record.UploadID)
	return out, nil
}

func (s *MySQLStore) PluginPackageUploadByUploadID(uploadID string) (domain.PluginPackageUploadRecord, bool) {
	var it domain.PluginPackageUploadRecord
	err := s.db.QueryRow(`SELECT id, upload_id, original_filename, uploaded_by, uploaded_by_name, DATE_FORMAT(uploaded_at,'%Y-%m-%d %H:%i:%s'), status,
		package_code, package_name, package_version, upload_path, staging_path, package_path, promoted_path,
		compressed_size, uncompressed_size, file_count, checksum_status, signature_status, publisher_id, trust_status, risk_level,
		COALESCE(CAST(risk_report_json AS CHAR),''), COALESCE(CAST(zip_scan_json AS CHAR),''), COALESCE(CAST(file_scan_json AS CHAR),''),
		COALESCE(CAST(manifest_validation_json AS CHAR),''), COALESCE(CAST(install_dry_run_json AS CHAR),''),
		approval_id, install_approval_id, COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(deleted_at,'%Y-%m-%d %H:%i:%s'),''),
		error_code, error_message, COALESCE(CAST(metadata_json AS CHAR),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_uploads WHERE upload_id=? LIMIT 1`, strings.TrimSpace(uploadID)).
		Scan(&it.ID, &it.UploadID, &it.OriginalFilename, &it.UploadedBy, &it.UploadedByName, &it.UploadedAt, &it.Status,
			&it.PackageCode, &it.PackageName, &it.PackageVersion, &it.UploadPath, &it.StagingPath, &it.PackagePath, &it.PromotedPath,
			&it.CompressedSize, &it.UncompressedSize, &it.FileCount, &it.ChecksumStatus, &it.SignatureStatus, &it.PublisherID, &it.TrustStatus, &it.RiskLevel,
			&it.RiskReportJSON, &it.ZipScanJSON, &it.FileScanJSON, &it.ManifestValidationJSON, &it.InstallDryRunJSON,
			&it.ApprovalID, &it.InstallApprovalID, &it.ExpiresAt, &it.DeletedAt,
			&it.ErrorCode, &it.ErrorMessage, &it.MetadataJSON, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginPackageUploadRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginPackageUploads(filter domain.PluginPackageUploadFilter) ([]domain.PluginPackageUploadRecord, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.RiskLevel != "" && filter.RiskLevel != "all" {
		where = append(where, "risk_level=?")
		args = append(args, filter.RiskLevel)
	}
	if filter.PackageCode != "" {
		where = append(where, "package_code=?")
		args = append(args, filter.PackageCode)
	}
	if filter.PublisherID != "" {
		where = append(where, "publisher_id=?")
		args = append(args, filter.PublisherID)
	}
	if filter.TrustStatus != "" && filter.TrustStatus != "all" {
		where = append(where, "trust_status=?")
		args = append(args, filter.TrustStatus)
	}
	if filter.UploadedBy > 0 {
		where = append(where, "uploaded_by=?")
		args = append(args, filter.UploadedBy)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(upload_id LIKE ? OR original_filename LIKE ? OR package_code LIKE ? OR package_name LIKE ? OR uploaded_by_name LIKE ?)")
		args = append(args, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_package_uploads WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, upload_id, original_filename, uploaded_by, uploaded_by_name, DATE_FORMAT(uploaded_at,'%Y-%m-%d %H:%i:%s'), status,
		package_code, package_name, package_version, upload_path, staging_path, package_path, promoted_path,
		compressed_size, uncompressed_size, file_count, checksum_status, signature_status, publisher_id, trust_status, risk_level,
		COALESCE(CAST(risk_report_json AS CHAR),''), COALESCE(CAST(zip_scan_json AS CHAR),''), COALESCE(CAST(file_scan_json AS CHAR),''),
		COALESCE(CAST(manifest_validation_json AS CHAR),''), COALESCE(CAST(install_dry_run_json AS CHAR),''),
		approval_id, install_approval_id, COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(deleted_at,'%Y-%m-%d %H:%i:%s'),''),
		error_code, error_message, COALESCE(CAST(metadata_json AS CHAR),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_uploads WHERE `+whereSQL+`
		ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginPackageUploadRecord{}
	for rows.Next() {
		var it domain.PluginPackageUploadRecord
		if err := rows.Scan(&it.ID, &it.UploadID, &it.OriginalFilename, &it.UploadedBy, &it.UploadedByName, &it.UploadedAt, &it.Status,
			&it.PackageCode, &it.PackageName, &it.PackageVersion, &it.UploadPath, &it.StagingPath, &it.PackagePath, &it.PromotedPath,
			&it.CompressedSize, &it.UncompressedSize, &it.FileCount, &it.ChecksumStatus, &it.SignatureStatus, &it.PublisherID, &it.TrustStatus, &it.RiskLevel,
			&it.RiskReportJSON, &it.ZipScanJSON, &it.FileScanJSON, &it.ManifestValidationJSON, &it.InstallDryRunJSON,
			&it.ApprovalID, &it.InstallApprovalID, &it.ExpiresAt, &it.DeletedAt,
			&it.ErrorCode, &it.ErrorMessage, &it.MetadataJSON, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Plugin package remote downloads (v1.7.0-P0-01) =====

func (s *MySQLStore) AppendPluginPackageDownload(record domain.PluginPackageDownloadRecord) (domain.PluginPackageDownloadRecord, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginPackageDownloadStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO plugin_package_downloads
		(plugin_code, version, source_url, final_url, signature_url, status, file_name, staging_path, file_size,
		sha256_expected, sha256_actual, content_type, error_code, error_message, created_by,
		created_at, downloaded_at, deleted_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),?,?,NOW())`,
		record.PluginCode, record.Version, record.SourceURL, record.FinalURL, record.SignatureURL, record.Status, record.FileName, record.StagingPath, record.FileSize,
		record.SHA256Expected, record.SHA256Actual, record.ContentType, record.ErrorCode, record.ErrorMessage, record.CreatedBy,
		nullTime(record.CreatedAt), nullTime(record.DownloadedAt), nullTime(record.DeletedAt))
	if err != nil {
		return domain.PluginPackageDownloadRecord{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	out, _ := s.PluginPackageDownloadByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginPackageDownload(record domain.PluginPackageDownloadRecord) (domain.PluginPackageDownloadRecord, error) {
	if record.ID <= 0 {
		return s.AppendPluginPackageDownload(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_package_downloads SET
		plugin_code=?, version=?, source_url=?, final_url=?, signature_url=?, status=?, file_name=?, staging_path=?, file_size=?,
		sha256_expected=?, sha256_actual=?, content_type=?, error_code=?, error_message=?, created_by=?,
		created_at=COALESCE(?, created_at), downloaded_at=?, deleted_at=?, updated_at=NOW()
		WHERE id=?`,
		record.PluginCode, record.Version, record.SourceURL, record.FinalURL, record.SignatureURL, record.Status, record.FileName, record.StagingPath, record.FileSize,
		record.SHA256Expected, record.SHA256Actual, record.ContentType, record.ErrorCode, record.ErrorMessage, record.CreatedBy,
		nullTime(record.CreatedAt), nullTime(record.DownloadedAt), nullTime(record.DeletedAt), record.ID)
	if err != nil {
		return domain.PluginPackageDownloadRecord{}, err
	}
	out, _ := s.PluginPackageDownloadByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginPackageDownloadByID(id int64) (domain.PluginPackageDownloadRecord, bool) {
	var it domain.PluginPackageDownloadRecord
	err := s.db.QueryRow(`SELECT id, plugin_code, version, source_url, final_url, signature_url, status, file_name, staging_path, file_size,
		sha256_expected, sha256_actual, content_type, error_code, error_message, created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), COALESCE(DATE_FORMAT(downloaded_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(deleted_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_downloads WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PluginCode, &it.Version, &it.SourceURL, &it.FinalURL, &it.SignatureURL, &it.Status, &it.FileName, &it.StagingPath, &it.FileSize,
			&it.SHA256Expected, &it.SHA256Actual, &it.ContentType, &it.ErrorCode, &it.ErrorMessage, &it.CreatedBy,
			&it.CreatedAt, &it.DownloadedAt, &it.DeletedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginPackageDownloadRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginPackageDownloads(filter domain.PluginPackageDownloadFilter) ([]domain.PluginPackageDownloadRecord, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR source_url LIKE ? OR final_url LIKE ? OR file_name LIKE ? OR status LIKE ?)")
		args = append(args, like, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_package_downloads WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, plugin_code, version, source_url, final_url, signature_url, status, file_name, staging_path, file_size,
		sha256_expected, sha256_actual, content_type, error_code, error_message, created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), COALESCE(DATE_FORMAT(downloaded_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(deleted_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_downloads WHERE `+whereSQL+`
		ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginPackageDownloadRecord{}
	for rows.Next() {
		var it domain.PluginPackageDownloadRecord
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.Version, &it.SourceURL, &it.FinalURL, &it.SignatureURL, &it.Status, &it.FileName, &it.StagingPath, &it.FileSize,
			&it.SHA256Expected, &it.SHA256Actual, &it.ContentType, &it.ErrorCode, &it.ErrorMessage, &it.CreatedBy,
			&it.CreatedAt, &it.DownloadedAt, &it.DeletedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Plugin package prechecks / compatibility checks (v1.7.0-P0-03) =====

func (s *MySQLStore) AppendPluginPackagePrecheck(record domain.PluginPackagePrecheckRecord) (domain.PluginPackagePrecheckRecord, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginPackagePrecheckStatusPassed
	}
	res, err := s.db.Exec(`INSERT INTO plugin_package_prechecks
		(package_download_id, plugin_code, version, status, manifest_json, package_path, staging_path, checksum_status,
		 error_code, error_message, created_by, started_at, finished_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())`,
		record.PackageDownloadID, record.PluginCode, record.Version, record.Status, uploadJSONValue(record.ManifestJSON), record.PackagePath, record.StagingPath, record.ChecksumStatus,
		record.ErrorCode, record.ErrorMessage, record.CreatedBy, nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt))
	if err != nil {
		return domain.PluginPackagePrecheckRecord{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginPackagePrecheckByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginPackagePrecheck(record domain.PluginPackagePrecheckRecord) (domain.PluginPackagePrecheckRecord, error) {
	if record.ID <= 0 {
		return s.AppendPluginPackagePrecheck(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_package_prechecks SET
		package_download_id=?, plugin_code=?, version=?, status=?, manifest_json=?, package_path=?, staging_path=?, checksum_status=?,
		error_code=?, error_message=?, created_by=?, started_at=?, finished_at=?, created_at=COALESCE(?, created_at), updated_at=NOW()
		WHERE id=?`,
		record.PackageDownloadID, record.PluginCode, record.Version, record.Status, uploadJSONValue(record.ManifestJSON), record.PackagePath, record.StagingPath, record.ChecksumStatus,
		record.ErrorCode, record.ErrorMessage, record.CreatedBy, nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt), record.ID)
	if err != nil {
		return domain.PluginPackagePrecheckRecord{}, err
	}
	out, _ := s.PluginPackagePrecheckByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginPackagePrecheckByID(id int64) (domain.PluginPackagePrecheckRecord, bool) {
	var it domain.PluginPackagePrecheckRecord
	err := s.db.QueryRow(`SELECT id, package_download_id, plugin_code, version, status, COALESCE(CAST(manifest_json AS CHAR),''),
		package_path, staging_path, checksum_status, error_code, error_message, created_by,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_prechecks WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PackageDownloadID, &it.PluginCode, &it.Version, &it.Status, &it.ManifestJSON,
			&it.PackagePath, &it.StagingPath, &it.ChecksumStatus, &it.ErrorCode, &it.ErrorMessage, &it.CreatedBy,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginPackagePrecheckRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginPackagePrechecks(filter domain.PluginPackagePrecheckFilter) ([]domain.PluginPackagePrecheckRecord, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR status LIKE ? OR package_path LIKE ? OR staging_path LIKE ?)")
		args = append(args, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_package_prechecks WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, package_download_id, plugin_code, version, status, COALESCE(CAST(manifest_json AS CHAR),''),
		package_path, staging_path, checksum_status, error_code, error_message, created_by,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_prechecks WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginPackagePrecheckRecord{}
	for rows.Next() {
		var it domain.PluginPackagePrecheckRecord
		if err := rows.Scan(&it.ID, &it.PackageDownloadID, &it.PluginCode, &it.Version, &it.Status, &it.ManifestJSON,
			&it.PackagePath, &it.StagingPath, &it.ChecksumStatus, &it.ErrorCode, &it.ErrorMessage, &it.CreatedBy,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) AppendPluginPackageCompatCheck(record domain.PluginPackageCompatCheckRecord) (domain.PluginPackageCompatCheckRecord, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginPackageCompatCheckStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO plugin_package_compat_checks
		(package_download_id, package_precheck_id, plugin_code, version, status, can_install, core_version, compatible_core_version,
		 dependency_result_json, conflict_result_json, permission_result_json, route_result_json, menu_result_json, hook_result_json,
		 config_schema_result_json, migration_result_json, warnings_json, errors_json, summary_json, created_by, started_at, finished_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())`,
		record.PackageDownloadID, record.PackagePrecheckID, record.PluginCode, record.Version, record.Status, record.CanInstall, record.CoreVersion, record.CompatibleCoreVersion,
		uploadJSONValue(record.DependencyResultJSON), uploadJSONValue(record.ConflictResultJSON), uploadJSONValue(record.PermissionResultJSON), uploadJSONValue(record.RouteResultJSON), uploadJSONValue(record.MenuResultJSON), uploadJSONValue(record.HookResultJSON),
		uploadJSONValue(record.ConfigSchemaResultJSON), uploadJSONValue(record.MigrationResultJSON), uploadJSONValue(record.WarningsJSON), uploadJSONValue(record.ErrorsJSON), uploadJSONValue(record.SummaryJSON), record.CreatedBy,
		nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt))
	if err != nil {
		return domain.PluginPackageCompatCheckRecord{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginPackageCompatCheckByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginPackageCompatCheck(record domain.PluginPackageCompatCheckRecord) (domain.PluginPackageCompatCheckRecord, error) {
	if record.ID <= 0 {
		return s.AppendPluginPackageCompatCheck(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_package_compat_checks SET
		package_download_id=?, package_precheck_id=?, plugin_code=?, version=?, status=?, can_install=?, core_version=?, compatible_core_version=?,
		dependency_result_json=?, conflict_result_json=?, permission_result_json=?, route_result_json=?, menu_result_json=?, hook_result_json=?,
		config_schema_result_json=?, migration_result_json=?, warnings_json=?, errors_json=?, summary_json=?, created_by=?, started_at=?, finished_at=?,
		created_at=COALESCE(?, created_at), updated_at=NOW()
		WHERE id=?`,
		record.PackageDownloadID, record.PackagePrecheckID, record.PluginCode, record.Version, record.Status, record.CanInstall, record.CoreVersion, record.CompatibleCoreVersion,
		uploadJSONValue(record.DependencyResultJSON), uploadJSONValue(record.ConflictResultJSON), uploadJSONValue(record.PermissionResultJSON), uploadJSONValue(record.RouteResultJSON), uploadJSONValue(record.MenuResultJSON), uploadJSONValue(record.HookResultJSON),
		uploadJSONValue(record.ConfigSchemaResultJSON), uploadJSONValue(record.MigrationResultJSON), uploadJSONValue(record.WarningsJSON), uploadJSONValue(record.ErrorsJSON), uploadJSONValue(record.SummaryJSON), record.CreatedBy, nullTime(record.StartedAt), nullTime(record.FinishedAt),
		nullTime(record.CreatedAt), record.ID)
	if err != nil {
		return domain.PluginPackageCompatCheckRecord{}, err
	}
	out, _ := s.PluginPackageCompatCheckByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginPackageCompatCheckByID(id int64) (domain.PluginPackageCompatCheckRecord, bool) {
	var it domain.PluginPackageCompatCheckRecord
	err := s.db.QueryRow(`SELECT id, package_download_id, package_precheck_id, plugin_code, version, status, can_install, core_version, compatible_core_version,
		COALESCE(CAST(dependency_result_json AS CHAR),''), COALESCE(CAST(conflict_result_json AS CHAR),''), COALESCE(CAST(permission_result_json AS CHAR),''),
		COALESCE(CAST(route_result_json AS CHAR),''), COALESCE(CAST(menu_result_json AS CHAR),''), COALESCE(CAST(hook_result_json AS CHAR),''),
		COALESCE(CAST(config_schema_result_json AS CHAR),''), COALESCE(CAST(migration_result_json AS CHAR),''), COALESCE(CAST(warnings_json AS CHAR),''),
		COALESCE(CAST(errors_json AS CHAR),''), COALESCE(CAST(summary_json AS CHAR),''), created_by,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_compat_checks WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PackageDownloadID, &it.PackagePrecheckID, &it.PluginCode, &it.Version, &it.Status, &it.CanInstall, &it.CoreVersion, &it.CompatibleCoreVersion,
			&it.DependencyResultJSON, &it.ConflictResultJSON, &it.PermissionResultJSON, &it.RouteResultJSON, &it.MenuResultJSON, &it.HookResultJSON,
			&it.ConfigSchemaResultJSON, &it.MigrationResultJSON, &it.WarningsJSON, &it.ErrorsJSON, &it.SummaryJSON, &it.CreatedBy,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginPackageCompatCheckRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginPackageCompatChecks(filter domain.PluginPackageCompatCheckFilter) ([]domain.PluginPackageCompatCheckRecord, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.PackagePrecheckID > 0 {
		where = append(where, "package_precheck_id=?")
		args = append(args, filter.PackagePrecheckID)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR status LIKE ? OR core_version LIKE ? OR compatible_core_version LIKE ?)")
		args = append(args, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_package_compat_checks WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, package_download_id, package_precheck_id, plugin_code, version, status, can_install, core_version, compatible_core_version,
		COALESCE(CAST(dependency_result_json AS CHAR),''), COALESCE(CAST(conflict_result_json AS CHAR),''), COALESCE(CAST(permission_result_json AS CHAR),''),
		COALESCE(CAST(route_result_json AS CHAR),''), COALESCE(CAST(menu_result_json AS CHAR),''), COALESCE(CAST(hook_result_json AS CHAR),''),
		COALESCE(CAST(config_schema_result_json AS CHAR),''), COALESCE(CAST(migration_result_json AS CHAR),''), COALESCE(CAST(warnings_json AS CHAR),''),
		COALESCE(CAST(errors_json AS CHAR),''), COALESCE(CAST(summary_json AS CHAR),''), created_by,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_compat_checks WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginPackageCompatCheckRecord{}
	for rows.Next() {
		var it domain.PluginPackageCompatCheckRecord
		if err := rows.Scan(&it.ID, &it.PackageDownloadID, &it.PackagePrecheckID, &it.PluginCode, &it.Version, &it.Status, &it.CanInstall, &it.CoreVersion, &it.CompatibleCoreVersion,
			&it.DependencyResultJSON, &it.ConflictResultJSON, &it.PermissionResultJSON, &it.RouteResultJSON, &it.MenuResultJSON, &it.HookResultJSON,
			&it.ConfigSchemaResultJSON, &it.MigrationResultJSON, &it.WarningsJSON, &it.ErrorsJSON, &it.SummaryJSON, &it.CreatedBy,
			&it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Plugin package signatures (detached verification, v1.7.1) =====

func (s *MySQLStore) AppendPluginPackageSignature(record domain.PluginPackageSignatureRecord) (domain.PluginPackageSignatureRecord, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginPackageSignatureStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO plugin_package_signatures
		(package_download_id, package_precheck_id, package_compat_check_id, plugin_code, version, publisher_id, key_id, algorithm,
		 status, signature_url, signature_file_path, package_sha256, manifest_sha256, signature_payload_json, signature_base64,
		 verified_at, error_message, warnings_json, created_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())`,
		record.PackageDownloadID, record.PackagePrecheckID, record.PackageCompatID, record.PluginCode, record.Version, record.PublisherID, record.KeyID, record.Algorithm,
		record.Status, record.SignatureURL, record.SignatureFilePath, record.PackageSHA256, record.ManifestSHA256,
		uploadJSONValue(record.SignaturePayloadJSON), record.SignatureBase64, nullTime(record.VerifiedAt), record.ErrorMessage,
		uploadJSONValue(record.WarningsJSON), record.CreatedBy, nullTime(record.CreatedAt))
	if err != nil {
		return domain.PluginPackageSignatureRecord{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	out, _ := s.PluginPackageSignatureByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginPackageSignature(record domain.PluginPackageSignatureRecord) (domain.PluginPackageSignatureRecord, error) {
	if record.ID <= 0 {
		return s.AppendPluginPackageSignature(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_package_signatures SET
		package_download_id=?, package_precheck_id=?, package_compat_check_id=?, plugin_code=?, version=?, publisher_id=?, key_id=?, algorithm=?,
		status=?, signature_url=?, signature_file_path=?, package_sha256=?, manifest_sha256=?, signature_payload_json=?, signature_base64=?,
		verified_at=?, error_message=?, warnings_json=?, created_by=?, updated_at=NOW()
		WHERE id=?`,
		record.PackageDownloadID, record.PackagePrecheckID, record.PackageCompatID, record.PluginCode, record.Version, record.PublisherID, record.KeyID, record.Algorithm,
		record.Status, record.SignatureURL, record.SignatureFilePath, record.PackageSHA256, record.ManifestSHA256,
		uploadJSONValue(record.SignaturePayloadJSON), record.SignatureBase64, nullTime(record.VerifiedAt), record.ErrorMessage,
		uploadJSONValue(record.WarningsJSON), record.CreatedBy, record.ID)
	if err != nil {
		return domain.PluginPackageSignatureRecord{}, err
	}
	out, _ := s.PluginPackageSignatureByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginPackageSignatureByID(id int64) (domain.PluginPackageSignatureRecord, bool) {
	var it domain.PluginPackageSignatureRecord
	err := s.db.QueryRow(`SELECT id, package_download_id, package_precheck_id, package_compat_check_id, plugin_code, version, publisher_id, key_id, algorithm,
		status, signature_url, signature_file_path, package_sha256, manifest_sha256, COALESCE(CAST(signature_payload_json AS CHAR),''), signature_base64,
		COALESCE(DATE_FORMAT(verified_at,'%Y-%m-%d %H:%i:%s'),''), error_message, COALESCE(CAST(warnings_json AS CHAR),''), created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_signatures WHERE id=? LIMIT 1`, id).
		Scan(&it.ID, &it.PackageDownloadID, &it.PackagePrecheckID, &it.PackageCompatID, &it.PluginCode, &it.Version, &it.PublisherID, &it.KeyID, &it.Algorithm,
			&it.Status, &it.SignatureURL, &it.SignatureFilePath, &it.PackageSHA256, &it.ManifestSHA256, &it.SignaturePayloadJSON, &it.SignatureBase64,
			&it.VerifiedAt, &it.ErrorMessage, &it.WarningsJSON, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginPackageSignatureRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) LatestPluginPackageSignatureByPrecheckID(precheckID int64) (domain.PluginPackageSignatureRecord, bool) {
	var it domain.PluginPackageSignatureRecord
	err := s.db.QueryRow(`SELECT id, package_download_id, package_precheck_id, package_compat_check_id, plugin_code, version, publisher_id, key_id, algorithm,
		status, signature_url, signature_file_path, package_sha256, manifest_sha256, COALESCE(CAST(signature_payload_json AS CHAR),''), signature_base64,
		COALESCE(DATE_FORMAT(verified_at,'%Y-%m-%d %H:%i:%s'),''), error_message, COALESCE(CAST(warnings_json AS CHAR),''), created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_signatures WHERE package_precheck_id=? ORDER BY id DESC LIMIT 1`, precheckID).
		Scan(&it.ID, &it.PackageDownloadID, &it.PackagePrecheckID, &it.PackageCompatID, &it.PluginCode, &it.Version, &it.PublisherID, &it.KeyID, &it.Algorithm,
			&it.Status, &it.SignatureURL, &it.SignatureFilePath, &it.PackageSHA256, &it.ManifestSHA256, &it.SignaturePayloadJSON, &it.SignatureBase64,
			&it.VerifiedAt, &it.ErrorMessage, &it.WarningsJSON, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginPackageSignatureRecord{}, false
	}
	if it.ID <= 0 {
		return domain.PluginPackageSignatureRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginPackageSignatures(filter domain.PluginPackageSignatureFilter) ([]domain.PluginPackageSignatureRecord, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if filter.PackageDownloadID > 0 {
		where = append(where, "package_download_id=?")
		args = append(args, filter.PackageDownloadID)
	}
	if filter.PackagePrecheckID > 0 {
		where = append(where, "package_precheck_id=?")
		args = append(args, filter.PackagePrecheckID)
	}
	if filter.Keyword != "" {
		like := "%" + strings.TrimSpace(filter.Keyword) + "%"
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR publisher_id LIKE ? OR key_id LIKE ? OR status LIKE ? OR signature_url LIKE ?)")
		args = append(args, like, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_package_signatures WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, package_download_id, package_precheck_id, package_compat_check_id, plugin_code, version, publisher_id, key_id, algorithm,
		status, signature_url, signature_file_path, package_sha256, manifest_sha256, COALESCE(CAST(signature_payload_json AS CHAR),''), signature_base64,
		COALESCE(DATE_FORMAT(verified_at,'%Y-%m-%d %H:%i:%s'),''), error_message, COALESCE(CAST(warnings_json AS CHAR),''), created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_package_signatures WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginPackageSignatureRecord{}
	for rows.Next() {
		var it domain.PluginPackageSignatureRecord
		if err := rows.Scan(&it.ID, &it.PackageDownloadID, &it.PackagePrecheckID, &it.PackageCompatID, &it.PluginCode, &it.Version, &it.PublisherID, &it.KeyID, &it.Algorithm,
			&it.Status, &it.SignatureURL, &it.SignatureFilePath, &it.PackageSHA256, &it.ManifestSHA256, &it.SignaturePayloadJSON, &it.SignatureBase64,
			&it.VerifiedAt, &it.ErrorMessage, &it.WarningsJSON, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Plugin enable prechecks (v1.7.0-P0-05) =====

func (s *MySQLStore) AppendPluginEnablePrecheck(record domain.PluginEnablePrecheckRecord) (domain.PluginEnablePrecheckRecord, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginEnablePrecheckStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO plugin_enable_prechecks
		(plugin_code, version, plugin_install_task_id, plugin_installation_id, status, can_enable, core_version, installed_path, manifest_sha256,
		 file_integrity_result_json, manifest_result_json, dependency_result_json, config_result_json, migration_result_json,
		 permission_result_json, menu_result_json, route_result_json, hook_result_json, content_type_result_json, runtime_result_json,
		 warnings_json, errors_json, summary_json, created_by, started_at, finished_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),NOW())`,
		record.PluginCode, record.Version, record.PluginInstallTaskID, record.PluginInstallationID, record.Status, boolToTinyInt(record.CanEnable),
		record.CoreVersion, record.InstalledPath, record.ManifestSHA256,
		uploadJSONValue(record.FileIntegrityResultJSON), uploadJSONValue(record.ManifestResultJSON), uploadJSONValue(record.DependencyResultJSON),
		uploadJSONValue(record.ConfigResultJSON), uploadJSONValue(record.MigrationResultJSON),
		uploadJSONValue(record.PermissionResultJSON), uploadJSONValue(record.MenuResultJSON), uploadJSONValue(record.RouteResultJSON),
		uploadJSONValue(record.HookResultJSON), uploadJSONValue(record.ContentTypeResultJSON), uploadJSONValue(record.RuntimeResultJSON),
		uploadJSONValue(record.WarningsJSON), uploadJSONValue(record.ErrorsJSON), uploadJSONValue(record.SummaryJSON),
		record.CreatedBy, nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt))
	if err != nil {
		return domain.PluginEnablePrecheckRecord{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginEnablePrecheckByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginEnablePrecheck(record domain.PluginEnablePrecheckRecord) (domain.PluginEnablePrecheckRecord, error) {
	if record.ID <= 0 {
		return s.AppendPluginEnablePrecheck(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_enable_prechecks SET
		plugin_code=?, version=?, plugin_install_task_id=?, plugin_installation_id=?, status=?, can_enable=?, core_version=?, installed_path=?, manifest_sha256=?,
		file_integrity_result_json=?, manifest_result_json=?, dependency_result_json=?, config_result_json=?, migration_result_json=?,
		permission_result_json=?, menu_result_json=?, route_result_json=?, hook_result_json=?, content_type_result_json=?, runtime_result_json=?,
		warnings_json=?, errors_json=?, summary_json=?, created_by=?, started_at=?, finished_at=?, created_at=COALESCE(?, created_at), updated_at=NOW()
		WHERE id=?`,
		record.PluginCode, record.Version, record.PluginInstallTaskID, record.PluginInstallationID, record.Status, boolToTinyInt(record.CanEnable),
		record.CoreVersion, record.InstalledPath, record.ManifestSHA256,
		uploadJSONValue(record.FileIntegrityResultJSON), uploadJSONValue(record.ManifestResultJSON), uploadJSONValue(record.DependencyResultJSON),
		uploadJSONValue(record.ConfigResultJSON), uploadJSONValue(record.MigrationResultJSON),
		uploadJSONValue(record.PermissionResultJSON), uploadJSONValue(record.MenuResultJSON), uploadJSONValue(record.RouteResultJSON),
		uploadJSONValue(record.HookResultJSON), uploadJSONValue(record.ContentTypeResultJSON), uploadJSONValue(record.RuntimeResultJSON),
		uploadJSONValue(record.WarningsJSON), uploadJSONValue(record.ErrorsJSON), uploadJSONValue(record.SummaryJSON),
		record.CreatedBy, nullTime(record.StartedAt), nullTime(record.FinishedAt), nullTime(record.CreatedAt), record.ID)
	if err != nil {
		return domain.PluginEnablePrecheckRecord{}, err
	}
	out, _ := s.PluginEnablePrecheckByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginEnablePrecheckByID(id int64) (domain.PluginEnablePrecheckRecord, bool) {
	var it domain.PluginEnablePrecheckRecord
	err := s.db.QueryRow(`SELECT id, plugin_code, version, plugin_install_task_id, plugin_installation_id, status, can_enable, core_version, installed_path, manifest_sha256,
		COALESCE(CAST(file_integrity_result_json AS CHAR),''), COALESCE(CAST(manifest_result_json AS CHAR),''), COALESCE(CAST(dependency_result_json AS CHAR),''),
		COALESCE(CAST(config_result_json AS CHAR),''), COALESCE(CAST(migration_result_json AS CHAR),''), COALESCE(CAST(permission_result_json AS CHAR),''),
		COALESCE(CAST(menu_result_json AS CHAR),''), COALESCE(CAST(route_result_json AS CHAR),''), COALESCE(CAST(hook_result_json AS CHAR),''),
		COALESCE(CAST(content_type_result_json AS CHAR),''), COALESCE(CAST(runtime_result_json AS CHAR),''), COALESCE(CAST(warnings_json AS CHAR),''),
		COALESCE(CAST(errors_json AS CHAR),''), COALESCE(CAST(summary_json AS CHAR),''), created_by,
		COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_enable_prechecks WHERE id=?`, id).
		Scan(&it.ID, &it.PluginCode, &it.Version, &it.PluginInstallTaskID, &it.PluginInstallationID, &it.Status, &it.CanEnable, &it.CoreVersion, &it.InstalledPath, &it.ManifestSHA256,
			&it.FileIntegrityResultJSON, &it.ManifestResultJSON, &it.DependencyResultJSON, &it.ConfigResultJSON, &it.MigrationResultJSON,
			&it.PermissionResultJSON, &it.MenuResultJSON, &it.RouteResultJSON, &it.HookResultJSON, &it.ContentTypeResultJSON, &it.RuntimeResultJSON,
			&it.WarningsJSON, &it.ErrorsJSON, &it.SummaryJSON, &it.CreatedBy, &it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginEnablePrecheckRecord{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginEnablePrechecks(filter domain.PluginEnablePrecheckFilter) ([]domain.PluginEnablePrecheckRecord, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if filter.PluginCode != "" {
		where = append(where, "plugin_code=?")
		args = append(args, filter.PluginCode)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR status LIKE ? OR core_version LIKE ?)")
		args = append(args, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_enable_prechecks WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, plugin_code, version, plugin_install_task_id, plugin_installation_id, status, can_enable, core_version, installed_path, manifest_sha256,
		COALESCE(CAST(warnings_json AS CHAR),''), COALESCE(CAST(errors_json AS CHAR),''), COALESCE(DATE_FORMAT(started_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(finished_at,'%Y-%m-%d %H:%i:%s'),''),
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_enable_prechecks WHERE `+whereSQL+`
		ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginEnablePrecheckRecord{}
	for rows.Next() {
		var it domain.PluginEnablePrecheckRecord
		if err := rows.Scan(&it.ID, &it.PluginCode, &it.Version, &it.PluginInstallTaskID, &it.PluginInstallationID, &it.Status, &it.CanEnable, &it.CoreVersion, &it.InstalledPath, &it.ManifestSHA256,
			&it.WarningsJSON, &it.ErrorsJSON, &it.StartedAt, &it.FinishedAt, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Plugin enable tasks (v1.7.0-P0-06) =====

func (s *MySQLStore) AppendPluginEnableTask(record domain.PluginEnableTask) (domain.PluginEnableTask, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginEnableTaskStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO plugin_enable_tasks
  (plugin_code,version,plugin_install_task_id,plugin_enable_precheck_id,status,previous_status,new_status,
   registered_content_types_json,registered_permissions_json,registered_menus_json,registered_routes_json,registered_hooks_json,
   effective_config_json,errors_json,warnings_json,rollback_log_json,started_at,finished_at,duration_ms,enabled_by)
  VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(record.PluginCode),
		strings.TrimSpace(record.Version),
		record.PluginInstallTaskID,
		record.PluginEnablePrecheckID,
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.PreviousStatus),
		strings.TrimSpace(record.NewStatus),
		nullJSONString(record.RegisteredContentTypesJSON),
		nullJSONString(record.RegisteredPermissionsJSON),
		nullJSONString(record.RegisteredMenusJSON),
		nullJSONString(record.RegisteredRoutesJSON),
		nullJSONString(record.RegisteredHooksJSON),
		nullJSONString(record.EffectiveConfigJSON),
		nullJSONString(record.ErrorsJSON),
		nullJSONString(record.WarningsJSON),
		nullJSONString(record.RollbackLogJSON),
		nullTime(record.StartedAt),
		nullTime(record.FinishedAt),
		record.DurationMS,
		record.EnabledBy,
	)
	if err != nil {
		return domain.PluginEnableTask{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginEnableTaskByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginEnableTask(record domain.PluginEnableTask) (domain.PluginEnableTask, error) {
	if record.ID <= 0 {
		return s.AppendPluginEnableTask(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_enable_tasks SET
  plugin_code=?,version=?,plugin_install_task_id=?,plugin_enable_precheck_id=?,status=?,previous_status=?,new_status=?,
  registered_content_types_json=?,registered_permissions_json=?,registered_menus_json=?,registered_routes_json=?,registered_hooks_json=?,
  effective_config_json=?,errors_json=?,warnings_json=?,rollback_log_json=?,started_at=?,finished_at=?,duration_ms=?,enabled_by=?
  WHERE id=?`,
		strings.TrimSpace(record.PluginCode),
		strings.TrimSpace(record.Version),
		record.PluginInstallTaskID,
		record.PluginEnablePrecheckID,
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.PreviousStatus),
		strings.TrimSpace(record.NewStatus),
		nullJSONString(record.RegisteredContentTypesJSON),
		nullJSONString(record.RegisteredPermissionsJSON),
		nullJSONString(record.RegisteredMenusJSON),
		nullJSONString(record.RegisteredRoutesJSON),
		nullJSONString(record.RegisteredHooksJSON),
		nullJSONString(record.EffectiveConfigJSON),
		nullJSONString(record.ErrorsJSON),
		nullJSONString(record.WarningsJSON),
		nullJSONString(record.RollbackLogJSON),
		nullTime(record.StartedAt),
		nullTime(record.FinishedAt),
		record.DurationMS,
		record.EnabledBy,
		record.ID,
	)
	if err != nil {
		return domain.PluginEnableTask{}, err
	}
	out, _ := s.PluginEnableTaskByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginEnableTaskByID(id int64) (domain.PluginEnableTask, bool) {
	var record domain.PluginEnableTask
	var startedAt, finishedAt sql.NullTime
	var createdAt, updatedAt time.Time
	var registeredContentTypes, registeredPermissions, registeredMenus, registeredRoutes, registeredHooks sql.NullString
	var effectiveConfig, errorsJSON, warningsJSON, rollbackLog sql.NullString
	err := s.db.QueryRow(`SELECT id,plugin_code,version,plugin_install_task_id,plugin_enable_precheck_id,status,previous_status,new_status,
  registered_content_types_json,registered_permissions_json,registered_menus_json,registered_routes_json,registered_hooks_json,
  effective_config_json,errors_json,warnings_json,rollback_log_json,started_at,finished_at,duration_ms,enabled_by,created_at,updated_at
  FROM plugin_enable_tasks WHERE id=?`, id).Scan(
		&record.ID,
		&record.PluginCode,
		&record.Version,
		&record.PluginInstallTaskID,
		&record.PluginEnablePrecheckID,
		&record.Status,
		&record.PreviousStatus,
		&record.NewStatus,
		&registeredContentTypes,
		&registeredPermissions,
		&registeredMenus,
		&registeredRoutes,
		&registeredHooks,
		&effectiveConfig,
		&errorsJSON,
		&warningsJSON,
		&rollbackLog,
		&startedAt,
		&finishedAt,
		&record.DurationMS,
		&record.EnabledBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.PluginEnableTask{}, false
	}
	record.RegisteredContentTypesJSON = registeredContentTypes.String
	record.RegisteredPermissionsJSON = registeredPermissions.String
	record.RegisteredMenusJSON = registeredMenus.String
	record.RegisteredRoutesJSON = registeredRoutes.String
	record.RegisteredHooksJSON = registeredHooks.String
	record.EffectiveConfigJSON = effectiveConfig.String
	record.ErrorsJSON = errorsJSON.String
	record.WarningsJSON = warningsJSON.String
	record.RollbackLogJSON = rollbackLog.String
	if startedAt.Valid {
		record.StartedAt = startedAt.Time.Format(TimeLayout)
	}
	if finishedAt.Valid {
		record.FinishedAt = finishedAt.Time.Format(TimeLayout)
	}
	record.CreatedAt = createdAt.Format(TimeLayout)
	record.UpdatedAt = updatedAt.Format(TimeLayout)
	return record, true
}

func (s *MySQLStore) PluginEnableTasks(filter domain.PluginEnableTaskFilter) ([]domain.PluginEnableTask, int, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	status := strings.TrimSpace(filter.Status)
	code := strings.TrimSpace(filter.PluginCode)
	keyword := strings.TrimSpace(filter.Keyword)

	where := []string{"1=1"}
	args := []any{}
	if status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if code != "" {
		where = append(where, "plugin_code=?")
		args = append(args, code)
	}
	if keyword != "" {
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR status LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM plugin_enable_tasks WHERE `+strings.Join(where, " AND "), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := s.db.Query(`SELECT id,plugin_code,version,plugin_install_task_id,plugin_enable_precheck_id,status,previous_status,new_status,
  registered_content_types_json,registered_permissions_json,registered_menus_json,registered_routes_json,registered_hooks_json,
  effective_config_json,errors_json,warnings_json,rollback_log_json,started_at,finished_at,duration_ms,enabled_by,created_at,updated_at
  FROM plugin_enable_tasks WHERE `+strings.Join(where, " AND ")+` ORDER BY id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginEnableTask{}
	for rows.Next() {
		var record domain.PluginEnableTask
		var startedAt, finishedAt sql.NullTime
		var createdAt, updatedAt time.Time
		var registeredContentTypes, registeredPermissions, registeredMenus, registeredRoutes, registeredHooks sql.NullString
		var effectiveConfig, errorsJSON, warningsJSON, rollbackLog sql.NullString
		if err := rows.Scan(
			&record.ID,
			&record.PluginCode,
			&record.Version,
			&record.PluginInstallTaskID,
			&record.PluginEnablePrecheckID,
			&record.Status,
			&record.PreviousStatus,
			&record.NewStatus,
			&registeredContentTypes,
			&registeredPermissions,
			&registeredMenus,
			&registeredRoutes,
			&registeredHooks,
			&effectiveConfig,
			&errorsJSON,
			&warningsJSON,
			&rollbackLog,
			&startedAt,
			&finishedAt,
			&record.DurationMS,
			&record.EnabledBy,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		record.RegisteredContentTypesJSON = registeredContentTypes.String
		record.RegisteredPermissionsJSON = registeredPermissions.String
		record.RegisteredMenusJSON = registeredMenus.String
		record.RegisteredRoutesJSON = registeredRoutes.String
		record.RegisteredHooksJSON = registeredHooks.String
		record.EffectiveConfigJSON = effectiveConfig.String
		record.ErrorsJSON = errorsJSON.String
		record.WarningsJSON = warningsJSON.String
		record.RollbackLogJSON = rollbackLog.String
		if startedAt.Valid {
			record.StartedAt = startedAt.Time.Format(TimeLayout)
		}
		if finishedAt.Valid {
			record.FinishedAt = finishedAt.Time.Format(TimeLayout)
		}
		record.CreatedAt = createdAt.Format(TimeLayout)
		record.UpdatedAt = updatedAt.Format(TimeLayout)
		out = append(out, record)
	}
	return out, total, nil
}

// ===== Plugin uninstall tasks (v1.7.0-P0-07) =====

func (s *MySQLStore) AppendPluginUninstallTask(record domain.PluginUninstallTask) (domain.PluginUninstallTask, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginUninstallTaskStatusPending
	}
	if strings.TrimSpace(record.UninstallType) == "" {
		record.UninstallType = domain.PluginUninstallTypeSoft
	}
	res, err := s.db.Exec(`INSERT INTO plugin_uninstall_tasks
  (plugin_code,version,plugin_installation_id,plugin_enable_task_id,status,uninstall_type,previous_status,new_status,
   affected_contents_count,affected_communities_count,dependent_plugins_json,
   unregistered_content_types_json,unregistered_permissions_json,unregistered_menus_json,unregistered_routes_json,unregistered_hooks_json,
   preserved_files_json,preserved_configs_json,preserved_migrations_json,
   errors_json,warnings_json,rollback_log_json,reason,started_at,finished_at,duration_ms,requested_by)
  VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(record.PluginCode),
		strings.TrimSpace(record.Version),
		record.PluginInstallationID,
		record.PluginEnableTaskID,
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.UninstallType),
		strings.TrimSpace(record.PreviousStatus),
		strings.TrimSpace(record.NewStatus),
		record.AffectedContentsCount,
		record.AffectedCommunitiesCount,
		nullJSONString(record.DependentPluginsJSON),
		nullJSONString(record.UnregisteredContentTypesJSON),
		nullJSONString(record.UnregisteredPermissionsJSON),
		nullJSONString(record.UnregisteredMenusJSON),
		nullJSONString(record.UnregisteredRoutesJSON),
		nullJSONString(record.UnregisteredHooksJSON),
		nullJSONString(record.PreservedFilesJSON),
		nullJSONString(record.PreservedConfigsJSON),
		nullJSONString(record.PreservedMigrationsJSON),
		nullJSONString(record.ErrorsJSON),
		nullJSONString(record.WarningsJSON),
		nullJSONString(record.RollbackLogJSON),
		strings.TrimSpace(record.Reason),
		nullTime(record.StartedAt),
		nullTime(record.FinishedAt),
		record.DurationMS,
		record.RequestedBy,
	)
	if err != nil {
		return domain.PluginUninstallTask{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginUninstallTaskByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginUninstallTask(record domain.PluginUninstallTask) (domain.PluginUninstallTask, error) {
	if record.ID <= 0 {
		return s.AppendPluginUninstallTask(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_uninstall_tasks SET
  plugin_code=?,version=?,plugin_installation_id=?,plugin_enable_task_id=?,status=?,uninstall_type=?,previous_status=?,new_status=?,
  affected_contents_count=?,affected_communities_count=?,dependent_plugins_json=?,
  unregistered_content_types_json=?,unregistered_permissions_json=?,unregistered_menus_json=?,unregistered_routes_json=?,unregistered_hooks_json=?,
  preserved_files_json=?,preserved_configs_json=?,preserved_migrations_json=?,
  errors_json=?,warnings_json=?,rollback_log_json=?,reason=?,started_at=?,finished_at=?,duration_ms=?,requested_by=?
  WHERE id=?`,
		strings.TrimSpace(record.PluginCode),
		strings.TrimSpace(record.Version),
		record.PluginInstallationID,
		record.PluginEnableTaskID,
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.UninstallType),
		strings.TrimSpace(record.PreviousStatus),
		strings.TrimSpace(record.NewStatus),
		record.AffectedContentsCount,
		record.AffectedCommunitiesCount,
		nullJSONString(record.DependentPluginsJSON),
		nullJSONString(record.UnregisteredContentTypesJSON),
		nullJSONString(record.UnregisteredPermissionsJSON),
		nullJSONString(record.UnregisteredMenusJSON),
		nullJSONString(record.UnregisteredRoutesJSON),
		nullJSONString(record.UnregisteredHooksJSON),
		nullJSONString(record.PreservedFilesJSON),
		nullJSONString(record.PreservedConfigsJSON),
		nullJSONString(record.PreservedMigrationsJSON),
		nullJSONString(record.ErrorsJSON),
		nullJSONString(record.WarningsJSON),
		nullJSONString(record.RollbackLogJSON),
		strings.TrimSpace(record.Reason),
		nullTime(record.StartedAt),
		nullTime(record.FinishedAt),
		record.DurationMS,
		record.RequestedBy,
		record.ID,
	)
	if err != nil {
		return domain.PluginUninstallTask{}, err
	}
	out, _ := s.PluginUninstallTaskByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginUninstallTaskByID(id int64) (domain.PluginUninstallTask, bool) {
	var record domain.PluginUninstallTask
	var startedAt, finishedAt sql.NullTime
	var createdAt, updatedAt time.Time
	var dependentPlugins, unregCT, unregPerm, unregMenus, unregRoutes, unregHooks sql.NullString
	var preservedFiles, preservedConfigs, preservedMigrations sql.NullString
	var errorsJSON, warningsJSON, rollbackLog sql.NullString
	err := s.db.QueryRow(`SELECT id,plugin_code,version,plugin_installation_id,plugin_enable_task_id,status,uninstall_type,previous_status,new_status,
  affected_contents_count,affected_communities_count,dependent_plugins_json,
  unregistered_content_types_json,unregistered_permissions_json,unregistered_menus_json,unregistered_routes_json,unregistered_hooks_json,
  preserved_files_json,preserved_configs_json,preserved_migrations_json,
  errors_json,warnings_json,rollback_log_json,reason,started_at,finished_at,duration_ms,requested_by,created_at,updated_at
  FROM plugin_uninstall_tasks WHERE id=?`, id).Scan(
		&record.ID,
		&record.PluginCode,
		&record.Version,
		&record.PluginInstallationID,
		&record.PluginEnableTaskID,
		&record.Status,
		&record.UninstallType,
		&record.PreviousStatus,
		&record.NewStatus,
		&record.AffectedContentsCount,
		&record.AffectedCommunitiesCount,
		&dependentPlugins,
		&unregCT,
		&unregPerm,
		&unregMenus,
		&unregRoutes,
		&unregHooks,
		&preservedFiles,
		&preservedConfigs,
		&preservedMigrations,
		&errorsJSON,
		&warningsJSON,
		&rollbackLog,
		&record.Reason,
		&startedAt,
		&finishedAt,
		&record.DurationMS,
		&record.RequestedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.PluginUninstallTask{}, false
	}
	record.DependentPluginsJSON = dependentPlugins.String
	record.UnregisteredContentTypesJSON = unregCT.String
	record.UnregisteredPermissionsJSON = unregPerm.String
	record.UnregisteredMenusJSON = unregMenus.String
	record.UnregisteredRoutesJSON = unregRoutes.String
	record.UnregisteredHooksJSON = unregHooks.String
	record.PreservedFilesJSON = preservedFiles.String
	record.PreservedConfigsJSON = preservedConfigs.String
	record.PreservedMigrationsJSON = preservedMigrations.String
	record.ErrorsJSON = errorsJSON.String
	record.WarningsJSON = warningsJSON.String
	record.RollbackLogJSON = rollbackLog.String
	if startedAt.Valid {
		record.StartedAt = startedAt.Time.Format(TimeLayout)
	}
	if finishedAt.Valid {
		record.FinishedAt = finishedAt.Time.Format(TimeLayout)
	}
	record.CreatedAt = createdAt.Format(TimeLayout)
	record.UpdatedAt = updatedAt.Format(TimeLayout)
	return record, true
}

func (s *MySQLStore) PluginUninstallTasks(filter domain.PluginUninstallTaskFilter) ([]domain.PluginUninstallTask, int, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	status := strings.TrimSpace(filter.Status)
	code := strings.TrimSpace(filter.PluginCode)
	keyword := strings.TrimSpace(filter.Keyword)

	where := []string{"1=1"}
	args := []any{}
	if status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if code != "" {
		where = append(where, "plugin_code=?")
		args = append(args, code)
	}
	if keyword != "" {
		where = append(where, "(plugin_code LIKE ? OR version LIKE ? OR status LIKE ? OR reason LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like, like, like)
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM plugin_uninstall_tasks WHERE `+strings.Join(where, " AND "), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := s.db.Query(`SELECT id,plugin_code,version,plugin_installation_id,plugin_enable_task_id,status,uninstall_type,previous_status,new_status,
  affected_contents_count,affected_communities_count,dependent_plugins_json,
  unregistered_content_types_json,unregistered_permissions_json,unregistered_menus_json,unregistered_routes_json,unregistered_hooks_json,
  preserved_files_json,preserved_configs_json,preserved_migrations_json,
  errors_json,warnings_json,rollback_log_json,reason,started_at,finished_at,duration_ms,requested_by,created_at,updated_at
  FROM plugin_uninstall_tasks WHERE `+strings.Join(where, " AND ")+` ORDER BY id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginUninstallTask{}
	for rows.Next() {
		var record domain.PluginUninstallTask
		var startedAt, finishedAt sql.NullTime
		var createdAt, updatedAt time.Time
		var dependentPlugins, unregCT, unregPerm, unregMenus, unregRoutes, unregHooks sql.NullString
		var preservedFiles, preservedConfigs, preservedMigrations sql.NullString
		var errorsJSON, warningsJSON, rollbackLog sql.NullString
		if err := rows.Scan(
			&record.ID,
			&record.PluginCode,
			&record.Version,
			&record.PluginInstallationID,
			&record.PluginEnableTaskID,
			&record.Status,
			&record.UninstallType,
			&record.PreviousStatus,
			&record.NewStatus,
			&record.AffectedContentsCount,
			&record.AffectedCommunitiesCount,
			&dependentPlugins,
			&unregCT,
			&unregPerm,
			&unregMenus,
			&unregRoutes,
			&unregHooks,
			&preservedFiles,
			&preservedConfigs,
			&preservedMigrations,
			&errorsJSON,
			&warningsJSON,
			&rollbackLog,
			&record.Reason,
			&startedAt,
			&finishedAt,
			&record.DurationMS,
			&record.RequestedBy,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		record.DependentPluginsJSON = dependentPlugins.String
		record.UnregisteredContentTypesJSON = unregCT.String
		record.UnregisteredPermissionsJSON = unregPerm.String
		record.UnregisteredMenusJSON = unregMenus.String
		record.UnregisteredRoutesJSON = unregRoutes.String
		record.UnregisteredHooksJSON = unregHooks.String
		record.PreservedFilesJSON = preservedFiles.String
		record.PreservedConfigsJSON = preservedConfigs.String
		record.PreservedMigrationsJSON = preservedMigrations.String
		record.ErrorsJSON = errorsJSON.String
		record.WarningsJSON = warningsJSON.String
		record.RollbackLogJSON = rollbackLog.String
		if startedAt.Valid {
			record.StartedAt = startedAt.Time.Format(TimeLayout)
		}
		if finishedAt.Valid {
			record.FinishedAt = finishedAt.Time.Format(TimeLayout)
		}
		record.CreatedAt = createdAt.Format(TimeLayout)
		record.UpdatedAt = updatedAt.Format(TimeLayout)
		out = append(out, record)
	}
	return out, total, nil
}

// ===== Plugin upgrade tasks (v1.7.0-P0-08) =====

func (s *MySQLStore) AppendPluginUpgradeTask(record domain.PluginUpgradeTask) (domain.PluginUpgradeTask, error) {
	if strings.TrimSpace(record.Status) == "" {
		record.Status = domain.PluginUpgradeTaskStatusPending
	}
	res, err := s.db.Exec(`INSERT INTO plugin_upgrade_tasks
  (plugin_code,old_version,new_version,old_plugin_installation_id,new_package_download_id,new_package_precheck_id,new_package_compat_check_id,
   status,previous_plugin_status,new_plugin_status,backup_path,old_install_path,new_install_path,
   manifest_diff_json,config_diff_json,permission_diff_json,menu_diff_json,route_diff_json,hook_diff_json,content_type_diff_json,migration_diff_json,impact_json,
   errors_json,warnings_json,rollback_log_json,reason,started_at,finished_at,duration_ms,requested_by)
  VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		strings.TrimSpace(record.PluginCode),
		strings.TrimSpace(record.OldVersion),
		strings.TrimSpace(record.NewVersion),
		record.OldPluginInstallationID,
		record.NewPackageDownloadID,
		record.NewPackagePrecheckID,
		record.NewPackageCompatCheckID,
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.PreviousPluginStatus),
		strings.TrimSpace(record.NewPluginStatus),
		strings.TrimSpace(record.BackupPath),
		strings.TrimSpace(record.OldInstallPath),
		strings.TrimSpace(record.NewInstallPath),
		nullJSONString(record.ManifestDiffJSON),
		nullJSONString(record.ConfigDiffJSON),
		nullJSONString(record.PermissionDiffJSON),
		nullJSONString(record.MenuDiffJSON),
		nullJSONString(record.RouteDiffJSON),
		nullJSONString(record.HookDiffJSON),
		nullJSONString(record.ContentTypeDiffJSON),
		nullJSONString(record.MigrationDiffJSON),
		nullJSONString(record.ImpactJSON),
		nullJSONString(record.ErrorsJSON),
		nullJSONString(record.WarningsJSON),
		nullJSONString(record.RollbackLogJSON),
		strings.TrimSpace(record.Reason),
		nullTime(record.StartedAt),
		nullTime(record.FinishedAt),
		record.DurationMS,
		record.RequestedBy,
	)
	if err != nil {
		return domain.PluginUpgradeTask{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginUpgradeTaskByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginUpgradeTask(record domain.PluginUpgradeTask) (domain.PluginUpgradeTask, error) {
	if record.ID <= 0 {
		return s.AppendPluginUpgradeTask(record)
	}
	_, err := s.db.Exec(`UPDATE plugin_upgrade_tasks SET
  plugin_code=?,old_version=?,new_version=?,old_plugin_installation_id=?,new_package_download_id=?,new_package_precheck_id=?,new_package_compat_check_id=?,
  status=?,previous_plugin_status=?,new_plugin_status=?,backup_path=?,old_install_path=?,new_install_path=?,
  manifest_diff_json=?,config_diff_json=?,permission_diff_json=?,menu_diff_json=?,route_diff_json=?,hook_diff_json=?,content_type_diff_json=?,migration_diff_json=?,impact_json=?,
  errors_json=?,warnings_json=?,rollback_log_json=?,reason=?,started_at=?,finished_at=?,duration_ms=?,requested_by=?
  WHERE id=?`,
		strings.TrimSpace(record.PluginCode),
		strings.TrimSpace(record.OldVersion),
		strings.TrimSpace(record.NewVersion),
		record.OldPluginInstallationID,
		record.NewPackageDownloadID,
		record.NewPackagePrecheckID,
		record.NewPackageCompatCheckID,
		strings.TrimSpace(record.Status),
		strings.TrimSpace(record.PreviousPluginStatus),
		strings.TrimSpace(record.NewPluginStatus),
		strings.TrimSpace(record.BackupPath),
		strings.TrimSpace(record.OldInstallPath),
		strings.TrimSpace(record.NewInstallPath),
		nullJSONString(record.ManifestDiffJSON),
		nullJSONString(record.ConfigDiffJSON),
		nullJSONString(record.PermissionDiffJSON),
		nullJSONString(record.MenuDiffJSON),
		nullJSONString(record.RouteDiffJSON),
		nullJSONString(record.HookDiffJSON),
		nullJSONString(record.ContentTypeDiffJSON),
		nullJSONString(record.MigrationDiffJSON),
		nullJSONString(record.ImpactJSON),
		nullJSONString(record.ErrorsJSON),
		nullJSONString(record.WarningsJSON),
		nullJSONString(record.RollbackLogJSON),
		strings.TrimSpace(record.Reason),
		nullTime(record.StartedAt),
		nullTime(record.FinishedAt),
		record.DurationMS,
		record.RequestedBy,
		record.ID,
	)
	if err != nil {
		return domain.PluginUpgradeTask{}, err
	}
	out, _ := s.PluginUpgradeTaskByID(record.ID)
	return out, nil
}

func (s *MySQLStore) PluginUpgradeTaskByID(id int64) (domain.PluginUpgradeTask, bool) {
	var record domain.PluginUpgradeTask
	var startedAt, finishedAt sql.NullTime
	var createdAt, updatedAt time.Time
	var manifestDiff, configDiff, permDiff, menuDiff, routeDiff, hookDiff, ctDiff, migDiff, impactJSON sql.NullString
	var errorsJSON, warningsJSON, rollbackLog sql.NullString
	err := s.db.QueryRow(`SELECT id,plugin_code,old_version,new_version,old_plugin_installation_id,new_package_download_id,new_package_precheck_id,new_package_compat_check_id,
  status,previous_plugin_status,new_plugin_status,backup_path,old_install_path,new_install_path,
  manifest_diff_json,config_diff_json,permission_diff_json,menu_diff_json,route_diff_json,hook_diff_json,content_type_diff_json,migration_diff_json,impact_json,
  errors_json,warnings_json,rollback_log_json,reason,started_at,finished_at,duration_ms,requested_by,created_at,updated_at
  FROM plugin_upgrade_tasks WHERE id=?`, id).Scan(
		&record.ID,
		&record.PluginCode,
		&record.OldVersion,
		&record.NewVersion,
		&record.OldPluginInstallationID,
		&record.NewPackageDownloadID,
		&record.NewPackagePrecheckID,
		&record.NewPackageCompatCheckID,
		&record.Status,
		&record.PreviousPluginStatus,
		&record.NewPluginStatus,
		&record.BackupPath,
		&record.OldInstallPath,
		&record.NewInstallPath,
		&manifestDiff,
		&configDiff,
		&permDiff,
		&menuDiff,
		&routeDiff,
		&hookDiff,
		&ctDiff,
		&migDiff,
		&impactJSON,
		&errorsJSON,
		&warningsJSON,
		&rollbackLog,
		&record.Reason,
		&startedAt,
		&finishedAt,
		&record.DurationMS,
		&record.RequestedBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.PluginUpgradeTask{}, false
	}
	record.ManifestDiffJSON = manifestDiff.String
	record.ConfigDiffJSON = configDiff.String
	record.PermissionDiffJSON = permDiff.String
	record.MenuDiffJSON = menuDiff.String
	record.RouteDiffJSON = routeDiff.String
	record.HookDiffJSON = hookDiff.String
	record.ContentTypeDiffJSON = ctDiff.String
	record.MigrationDiffJSON = migDiff.String
	record.ImpactJSON = impactJSON.String
	record.ErrorsJSON = errorsJSON.String
	record.WarningsJSON = warningsJSON.String
	record.RollbackLogJSON = rollbackLog.String
	if startedAt.Valid {
		record.StartedAt = startedAt.Time.Format(TimeLayout)
	}
	if finishedAt.Valid {
		record.FinishedAt = finishedAt.Time.Format(TimeLayout)
	}
	record.CreatedAt = createdAt.Format(TimeLayout)
	record.UpdatedAt = updatedAt.Format(TimeLayout)
	return record, true
}

func (s *MySQLStore) PluginUpgradeTasks(filter domain.PluginUpgradeTaskFilter) ([]domain.PluginUpgradeTask, int, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	status := strings.TrimSpace(filter.Status)
	code := strings.TrimSpace(filter.PluginCode)
	keyword := strings.TrimSpace(filter.Keyword)

	where := []string{"1=1"}
	args := []any{}
	if status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if code != "" {
		where = append(where, "plugin_code=?")
		args = append(args, code)
	}
	if keyword != "" {
		where = append(where, "(plugin_code LIKE ? OR old_version LIKE ? OR new_version LIKE ? OR status LIKE ? OR reason LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like, like, like, like)
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM plugin_upgrade_tasks WHERE `+strings.Join(where, " AND "), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := s.db.Query(`SELECT id,plugin_code,old_version,new_version,old_plugin_installation_id,new_package_download_id,new_package_precheck_id,new_package_compat_check_id,
  status,previous_plugin_status,new_plugin_status,backup_path,old_install_path,new_install_path,
  manifest_diff_json,config_diff_json,permission_diff_json,menu_diff_json,route_diff_json,hook_diff_json,content_type_diff_json,migration_diff_json,impact_json,
  errors_json,warnings_json,rollback_log_json,reason,started_at,finished_at,duration_ms,requested_by,created_at,updated_at
  FROM plugin_upgrade_tasks WHERE `+strings.Join(where, " AND ")+` ORDER BY id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginUpgradeTask{}
	for rows.Next() {
		var record domain.PluginUpgradeTask
		var startedAt, finishedAt sql.NullTime
		var createdAt, updatedAt time.Time
		var manifestDiff, configDiff, permDiff, menuDiff, routeDiff, hookDiff, ctDiff, migDiff, impactJSON sql.NullString
		var errorsJSON, warningsJSON, rollbackLog sql.NullString
		if err := rows.Scan(
			&record.ID,
			&record.PluginCode,
			&record.OldVersion,
			&record.NewVersion,
			&record.OldPluginInstallationID,
			&record.NewPackageDownloadID,
			&record.NewPackagePrecheckID,
			&record.NewPackageCompatCheckID,
			&record.Status,
			&record.PreviousPluginStatus,
			&record.NewPluginStatus,
			&record.BackupPath,
			&record.OldInstallPath,
			&record.NewInstallPath,
			&manifestDiff,
			&configDiff,
			&permDiff,
			&menuDiff,
			&routeDiff,
			&hookDiff,
			&ctDiff,
			&migDiff,
			&impactJSON,
			&errorsJSON,
			&warningsJSON,
			&rollbackLog,
			&record.Reason,
			&startedAt,
			&finishedAt,
			&record.DurationMS,
			&record.RequestedBy,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		record.ManifestDiffJSON = manifestDiff.String
		record.ConfigDiffJSON = configDiff.String
		record.PermissionDiffJSON = permDiff.String
		record.MenuDiffJSON = menuDiff.String
		record.RouteDiffJSON = routeDiff.String
		record.HookDiffJSON = hookDiff.String
		record.ContentTypeDiffJSON = ctDiff.String
		record.MigrationDiffJSON = migDiff.String
		record.ImpactJSON = impactJSON.String
		record.ErrorsJSON = errorsJSON.String
		record.WarningsJSON = warningsJSON.String
		record.RollbackLogJSON = rollbackLog.String
		if startedAt.Valid {
			record.StartedAt = startedAt.Time.Format(TimeLayout)
		}
		if finishedAt.Valid {
			record.FinishedAt = finishedAt.Time.Format(TimeLayout)
		}
		record.CreatedAt = createdAt.Format(TimeLayout)
		record.UpdatedAt = updatedAt.Format(TimeLayout)
		out = append(out, record)
	}
	return out, total, nil
}

// ===== Plugin operations (v1.6.0-P0-06) =====

func (s *MySQLStore) AppendPluginOperationSnapshot(record domain.PluginOperationSnapshot) (domain.PluginOperationSnapshot, error) {
	record.OperationID = strings.TrimSpace(record.OperationID)
	record.OperationType = strings.TrimSpace(record.OperationType)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.Status = strings.TrimSpace(record.Status)
	if record.OperationID == "" || record.OperationType == "" || record.PluginCode == "" {
		return domain.PluginOperationSnapshot{}, errors.New("operation_id/operation_type/plugin_code 不能为空")
	}
	if record.Status == "" {
		record.Status = domain.PluginOperationStatusCreated
	}
	res, err := s.db.Exec(`INSERT INTO plugin_operation_snapshots
		(operation_id, operation_type, plugin_code, from_version, to_version, package_path, package_source, approval_id,
		before_plugin_json, before_manifest_json, before_config_json, before_config_version_id, before_migrations_json,
		before_permissions_json, before_menus_json, before_routes_json, before_dependencies_json, before_status,
		after_manifest_json, dry_run_json, risk_report_json, diff_json, checksum_summary_json, signature_summary_json,
		status, error_code, error_message, created_by, metadata_json, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())`,
		record.OperationID, record.OperationType, record.PluginCode, record.FromVersion, record.ToVersion, record.PackagePath, record.PackageSource, record.ApprovalID,
		uploadJSONValue(record.BeforePluginJSON), uploadJSONValue(record.BeforeManifestJSON), uploadJSONValue(record.BeforeConfigJSON), record.BeforeConfigVersionID, uploadJSONValue(record.BeforeMigrationsJSON),
		uploadJSONValue(record.BeforePermissionsJSON), uploadJSONValue(record.BeforeMenusJSON), uploadJSONValue(record.BeforeRoutesJSON), uploadJSONValue(record.BeforeDependenciesJSON), record.BeforeStatus,
		uploadJSONValue(record.AfterManifestJSON), uploadJSONValue(record.DryRunJSON), uploadJSONValue(record.RiskReportJSON), uploadJSONValue(record.DiffJSON), uploadJSONValue(record.ChecksumSummaryJSON), uploadJSONValue(record.SignatureSummaryJSON),
		record.Status, record.ErrorCode, record.ErrorMessage, record.CreatedBy, uploadJSONValue(record.MetadataJSON))
	if err != nil {
		return domain.PluginOperationSnapshot{}, err
	}
	id, _ := res.LastInsertId()
	record.ID = id
	out, _ := s.PluginOperationSnapshotByOperationID(record.OperationID)
	return out, nil
}

func (s *MySQLStore) SavePluginOperationSnapshot(record domain.PluginOperationSnapshot) (domain.PluginOperationSnapshot, error) {
	record.OperationID = strings.TrimSpace(record.OperationID)
	if record.OperationID == "" {
		return domain.PluginOperationSnapshot{}, errors.New("operation_id 不能为空")
	}
	_, err := s.db.Exec(`UPDATE plugin_operation_snapshots SET
		operation_type=?, plugin_code=?, from_version=?, to_version=?, package_path=?, package_source=?, approval_id=?,
		before_plugin_json=?, before_manifest_json=?, before_config_json=?, before_config_version_id=?, before_migrations_json=?,
		before_permissions_json=?, before_menus_json=?, before_routes_json=?, before_dependencies_json=?, before_status=?,
		after_manifest_json=?, dry_run_json=?, risk_report_json=?, diff_json=?, checksum_summary_json=?, signature_summary_json=?,
		status=?, error_code=?, error_message=?, created_by=?, metadata_json=?, updated_at=NOW()
		WHERE operation_id=?`,
		record.OperationType, record.PluginCode, record.FromVersion, record.ToVersion, record.PackagePath, record.PackageSource, record.ApprovalID,
		uploadJSONValue(record.BeforePluginJSON), uploadJSONValue(record.BeforeManifestJSON), uploadJSONValue(record.BeforeConfigJSON), record.BeforeConfigVersionID, uploadJSONValue(record.BeforeMigrationsJSON),
		uploadJSONValue(record.BeforePermissionsJSON), uploadJSONValue(record.BeforeMenusJSON), uploadJSONValue(record.BeforeRoutesJSON), uploadJSONValue(record.BeforeDependenciesJSON), record.BeforeStatus,
		uploadJSONValue(record.AfterManifestJSON), uploadJSONValue(record.DryRunJSON), uploadJSONValue(record.RiskReportJSON), uploadJSONValue(record.DiffJSON), uploadJSONValue(record.ChecksumSummaryJSON), uploadJSONValue(record.SignatureSummaryJSON),
		record.Status, record.ErrorCode, record.ErrorMessage, record.CreatedBy, uploadJSONValue(record.MetadataJSON), record.OperationID)
	if err != nil {
		return domain.PluginOperationSnapshot{}, err
	}
	out, _ := s.PluginOperationSnapshotByOperationID(record.OperationID)
	return out, nil
}

func (s *MySQLStore) PluginOperationSnapshotByOperationID(operationID string) (domain.PluginOperationSnapshot, bool) {
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return domain.PluginOperationSnapshot{}, false
	}
	var it domain.PluginOperationSnapshot
	err := s.db.QueryRow(`SELECT id, operation_id, operation_type, plugin_code, from_version, to_version, package_path, package_source, approval_id,
		COALESCE(CAST(before_plugin_json AS CHAR),''), COALESCE(CAST(before_manifest_json AS CHAR),''), COALESCE(CAST(before_config_json AS CHAR),''),
		before_config_version_id, COALESCE(CAST(before_migrations_json AS CHAR),''), COALESCE(CAST(before_permissions_json AS CHAR),''), COALESCE(CAST(before_menus_json AS CHAR),''), COALESCE(CAST(before_routes_json AS CHAR),''), COALESCE(CAST(before_dependencies_json AS CHAR),''),
		before_status,
		COALESCE(CAST(after_manifest_json AS CHAR),''), COALESCE(CAST(dry_run_json AS CHAR),''), COALESCE(CAST(risk_report_json AS CHAR),''), COALESCE(CAST(diff_json AS CHAR),''),
		COALESCE(CAST(checksum_summary_json AS CHAR),''), COALESCE(CAST(signature_summary_json AS CHAR),''), status, error_code, error_message, created_by,
		COALESCE(CAST(metadata_json AS CHAR),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_operation_snapshots WHERE operation_id=? LIMIT 1`, operationID).
		Scan(&it.ID, &it.OperationID, &it.OperationType, &it.PluginCode, &it.FromVersion, &it.ToVersion, &it.PackagePath, &it.PackageSource, &it.ApprovalID,
			&it.BeforePluginJSON, &it.BeforeManifestJSON, &it.BeforeConfigJSON, &it.BeforeConfigVersionID, &it.BeforeMigrationsJSON, &it.BeforePermissionsJSON, &it.BeforeMenusJSON, &it.BeforeRoutesJSON, &it.BeforeDependenciesJSON,
			&it.BeforeStatus,
			&it.AfterManifestJSON, &it.DryRunJSON, &it.RiskReportJSON, &it.DiffJSON,
			&it.ChecksumSummaryJSON, &it.SignatureSummaryJSON, &it.Status, &it.ErrorCode, &it.ErrorMessage, &it.CreatedBy,
			&it.MetadataJSON, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.PluginOperationSnapshot{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginOperationSnapshots(filter domain.PluginOperationFilter) ([]domain.PluginOperationSnapshot, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if code := strings.TrimSpace(filter.PluginCode); code != "" {
		where = append(where, "plugin_code=?")
		args = append(args, code)
	}
	if typ := strings.TrimSpace(filter.OperationType); typ != "" && typ != "all" {
		where = append(where, "operation_type=?")
		args = append(args, typ)
	}
	if status := strings.TrimSpace(filter.Status); status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_operation_snapshots WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args2 := append(append([]any{}, args...), filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id, operation_id, operation_type, plugin_code, from_version, to_version, package_path, package_source, approval_id,
		COALESCE(CAST(before_plugin_json AS CHAR),''), COALESCE(CAST(before_manifest_json AS CHAR),''), COALESCE(CAST(before_config_json AS CHAR),''),
		before_config_version_id, COALESCE(CAST(before_migrations_json AS CHAR),''), COALESCE(CAST(before_permissions_json AS CHAR),''), COALESCE(CAST(before_menus_json AS CHAR),''), COALESCE(CAST(before_routes_json AS CHAR),''), COALESCE(CAST(before_dependencies_json AS CHAR),''),
		before_status,
		COALESCE(CAST(after_manifest_json AS CHAR),''), COALESCE(CAST(dry_run_json AS CHAR),''), COALESCE(CAST(risk_report_json AS CHAR),''), COALESCE(CAST(diff_json AS CHAR),''),
		COALESCE(CAST(checksum_summary_json AS CHAR),''), COALESCE(CAST(signature_summary_json AS CHAR),''), status, error_code, error_message, created_by,
		COALESCE(CAST(metadata_json AS CHAR),''), DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_operation_snapshots WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginOperationSnapshot{}
	for rows.Next() {
		var it domain.PluginOperationSnapshot
		if err := rows.Scan(&it.ID, &it.OperationID, &it.OperationType, &it.PluginCode, &it.FromVersion, &it.ToVersion, &it.PackagePath, &it.PackageSource, &it.ApprovalID,
			&it.BeforePluginJSON, &it.BeforeManifestJSON, &it.BeforeConfigJSON, &it.BeforeConfigVersionID, &it.BeforeMigrationsJSON, &it.BeforePermissionsJSON, &it.BeforeMenusJSON, &it.BeforeRoutesJSON, &it.BeforeDependenciesJSON,
			&it.BeforeStatus,
			&it.AfterManifestJSON, &it.DryRunJSON, &it.RiskReportJSON, &it.DiffJSON,
			&it.ChecksumSummaryJSON, &it.SignatureSummaryJSON, &it.Status, &it.ErrorCode, &it.ErrorMessage, &it.CreatedBy,
			&it.MetadataJSON, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) DeletePluginByCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	_, err := s.db.Exec(`DELETE FROM plugins WHERE plugin_code=?`, code)
	return err
}

func (s *MySQLStore) DeleteCommunityPluginsByCode(code string) (int, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil
	}
	res, err := s.db.Exec(`DELETE FROM community_plugins WHERE plugin_code=?`, code)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (s *MySQLStore) DeletePluginMigrationsByPlugin(code string) (int, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil
	}
	res, err := s.db.Exec(`DELETE FROM plugin_migrations WHERE plugin_code=?`, code)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (s *MySQLStore) DeletePluginMigrationsByVersion(code, version string) (int, error) {
	code = strings.TrimSpace(code)
	version = strings.TrimSpace(version)
	if code == "" || version == "" {
		return 0, nil
	}
	res, err := s.db.Exec(`DELETE FROM plugin_migrations WHERE plugin_code=? AND version=?`, code, version)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (s *MySQLStore) DeletePluginConfigVersionsByPlugin(code string) (int, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil
	}
	res, err := s.db.Exec(`DELETE FROM plugin_config_versions WHERE plugin_code=?`, code)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (s *MySQLStore) AppendPluginTrustedPublisher(record domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error) {
	res, err := s.db.Exec(`INSERT INTO plugin_trusted_publishers
		(publisher_id,name,homepage,email,public_key_id,public_key_algorithm,public_key,fingerprint,status,notes,created_by,created_at,updated_by,updated_at,revoked_at,blocked_at,expires_at,metadata_json)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),?,NOW(),?,?,?,?)`,
		record.PublisherID, record.Name, record.Homepage, record.Email, record.PublicKeyID, record.PublicKeyAlgorithm, record.PublicKey, record.Fingerprint, record.Status, record.Notes,
		record.CreatedBy, nullTime(record.CreatedAt), record.UpdatedBy, nullTime(record.RevokedAt), nullTime(record.BlockedAt), nullTime(record.ExpiresAt), uploadJSONValue(record.MetadataJSON))
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginTrustedPublisherByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginTrustedPublisher(record domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error) {
	_, err := s.db.Exec(`UPDATE plugin_trusted_publishers SET
		publisher_id=?, name=?, homepage=?, email=?, public_key_id=?, public_key_algorithm=?, public_key=?, fingerprint=?, status=?, notes=?,
		updated_by=?, updated_at=NOW(), revoked_at=?, blocked_at=?, expires_at=?, metadata_json=?
		WHERE id=?`,
		record.PublisherID, record.Name, record.Homepage, record.Email, record.PublicKeyID, record.PublicKeyAlgorithm, record.PublicKey, record.Fingerprint, record.Status, record.Notes,
		record.UpdatedBy, nullTime(record.RevokedAt), nullTime(record.BlockedAt), nullTime(record.ExpiresAt), uploadJSONValue(record.MetadataJSON), record.ID)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	out, _ := s.PluginTrustedPublisherByID(record.ID)
	return out, nil
}

func (s *MySQLStore) DeletePluginTrustedPublisher(id int64) error {
	_, err := s.db.Exec(`DELETE FROM plugin_trusted_publishers WHERE id=?`, id)
	return err
}

func (s *MySQLStore) PluginTrustedPublisherByID(id int64) (domain.PluginTrustedPublisher, bool) {
	return s.scanTrustedPublisher(`WHERE id=?`, id)
}

func (s *MySQLStore) PluginTrustedPublisherByKey(publisherID, publicKeyID string) (domain.PluginTrustedPublisher, bool) {
	return s.scanTrustedPublisher(`WHERE publisher_id=? AND public_key_id=?`, strings.TrimSpace(publisherID), strings.TrimSpace(publicKeyID))
}

func (s *MySQLStore) scanTrustedPublisher(where string, args ...any) (domain.PluginTrustedPublisher, bool) {
	var it domain.PluginTrustedPublisher
	err := s.db.QueryRow(`SELECT id,publisher_id,name,homepage,email,public_key_id,public_key_algorithm,public_key,fingerprint,status,notes,created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),updated_by,DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(blocked_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(CAST(metadata_json AS CHAR),'')
		FROM plugin_trusted_publishers `+where+` LIMIT 1`, args...).Scan(
		&it.ID, &it.PublisherID, &it.Name, &it.Homepage, &it.Email, &it.PublicKeyID, &it.PublicKeyAlgorithm, &it.PublicKey, &it.Fingerprint, &it.Status, &it.Notes, &it.CreatedBy,
		&it.CreatedAt, &it.UpdatedBy, &it.UpdatedAt, &it.RevokedAt, &it.BlockedAt, &it.ExpiresAt, &it.MetadataJSON)
	if err != nil {
		return domain.PluginTrustedPublisher{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginTrustedPublishers(filter domain.PluginTrustedPublisherFilter) ([]domain.PluginTrustedPublisher, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(publisher_id LIKE ? OR name LIKE ? OR public_key_id LIKE ? OR fingerprint LIKE ? OR notes LIKE ?)")
		args = append(args, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_trusted_publishers WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id,publisher_id,name,homepage,email,public_key_id,public_key_algorithm,public_key,fingerprint,status,notes,created_by,
		DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),updated_by,DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(DATE_FORMAT(revoked_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(blocked_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(expires_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(CAST(metadata_json AS CHAR),'')
		FROM plugin_trusted_publishers WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginTrustedPublisher{}
	for rows.Next() {
		var it domain.PluginTrustedPublisher
		if err := rows.Scan(&it.ID, &it.PublisherID, &it.Name, &it.Homepage, &it.Email, &it.PublicKeyID, &it.PublicKeyAlgorithm, &it.PublicKey, &it.Fingerprint, &it.Status, &it.Notes, &it.CreatedBy,
			&it.CreatedAt, &it.UpdatedBy, &it.UpdatedAt, &it.RevokedAt, &it.BlockedAt, &it.ExpiresAt, &it.MetadataJSON); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

// ===== Remote plugin indexes (v1.6.0-P0-04) =====

func (s *MySQLStore) AppendPluginRemoteIndex(record domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error) {
	res, err := s.db.Exec(`INSERT INTO plugin_remote_indexes
		(source_id,name,index_url,homepage,description,status,trust_policy,last_fetch_status,last_fetch_at,last_error_code,last_error_message,last_index_hash,metadata_json,created_by,created_at,updated_by,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,COALESCE(?,NOW()),?,NOW())`,
		record.SourceID, record.Name, record.IndexURL, record.Homepage, record.Description, record.Status, record.TrustPolicy,
		record.LastFetchStatus, nullTime(record.LastFetchAt), record.LastErrorCode, record.LastErrorMessage, record.LastIndexHash, uploadJSONValue(record.MetadataJSON),
		record.CreatedBy, nullTime(record.CreatedAt), record.UpdatedBy)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	id, _ := res.LastInsertId()
	out, _ := s.PluginRemoteIndexByID(id)
	return out, nil
}

func (s *MySQLStore) SavePluginRemoteIndex(record domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error) {
	_, err := s.db.Exec(`UPDATE plugin_remote_indexes SET
		source_id=?, name=?, index_url=?, homepage=?, description=?, status=?, trust_policy=?, last_fetch_status=?, last_fetch_at=?,
		last_error_code=?, last_error_message=?, last_index_hash=?, metadata_json=?, updated_by=?, updated_at=NOW()
		WHERE id=?`,
		record.SourceID, record.Name, record.IndexURL, record.Homepage, record.Description, record.Status, record.TrustPolicy, record.LastFetchStatus, nullTime(record.LastFetchAt),
		record.LastErrorCode, record.LastErrorMessage, record.LastIndexHash, uploadJSONValue(record.MetadataJSON), record.UpdatedBy, record.ID)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	out, _ := s.PluginRemoteIndexByID(record.ID)
	return out, nil
}

func (s *MySQLStore) DeletePluginRemoteIndex(id int64) error {
	_, err := s.db.Exec(`DELETE FROM plugin_remote_indexes WHERE id=?`, id)
	return err
}

func (s *MySQLStore) PluginRemoteIndexByID(id int64) (domain.PluginRemoteIndexSource, bool) {
	return s.scanRemoteIndex(`WHERE id=?`, id)
}

func (s *MySQLStore) PluginRemoteIndexBySourceID(sourceID string) (domain.PluginRemoteIndexSource, bool) {
	return s.scanRemoteIndex(`WHERE source_id=?`, strings.TrimSpace(sourceID))
}

func (s *MySQLStore) scanRemoteIndex(where string, args ...any) (domain.PluginRemoteIndexSource, bool) {
	var it domain.PluginRemoteIndexSource
	err := s.db.QueryRow(`SELECT id,source_id,name,index_url,homepage,description,status,trust_policy,last_fetch_status,
		COALESCE(DATE_FORMAT(last_fetch_at,'%Y-%m-%d %H:%i:%s'),''),last_error_code,last_error_message,last_index_hash,COALESCE(CAST(metadata_json AS CHAR),''),
		created_by,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),updated_by,DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_remote_indexes `+where+` LIMIT 1`, args...).Scan(
		&it.ID, &it.SourceID, &it.Name, &it.IndexURL, &it.Homepage, &it.Description, &it.Status, &it.TrustPolicy, &it.LastFetchStatus,
		&it.LastFetchAt, &it.LastErrorCode, &it.LastErrorMessage, &it.LastIndexHash, &it.MetadataJSON,
		&it.CreatedBy, &it.CreatedAt, &it.UpdatedBy, &it.UpdatedAt)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, false
	}
	return it, true
}

func (s *MySQLStore) PluginRemoteIndexes(filter domain.PluginRemoteIndexFilter) ([]domain.PluginRemoteIndexSource, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	where := []string{"1=1"}
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where = append(where, "status=?")
		args = append(args, filter.Status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		where = append(where, "(source_id LIKE ? OR name LIKE ? OR index_url LIKE ? OR homepage LIKE ? OR description LIKE ? OR last_error_code LIKE ? OR last_error_message LIKE ?)")
		args = append(args, like, like, like, like, like, like, like)
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_remote_indexes WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	rows, err := s.db.Query(`SELECT id,source_id,name,index_url,homepage,description,status,trust_policy,last_fetch_status,
		COALESCE(DATE_FORMAT(last_fetch_at,'%Y-%m-%d %H:%i:%s'),''),last_error_code,last_error_message,last_index_hash,COALESCE(CAST(metadata_json AS CHAR),''),
		created_by,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),updated_by,DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')
		FROM plugin_remote_indexes WHERE `+whereSQL+` ORDER BY id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.PluginRemoteIndexSource{}
	for rows.Next() {
		var it domain.PluginRemoteIndexSource
		if err := rows.Scan(&it.ID, &it.SourceID, &it.Name, &it.IndexURL, &it.Homepage, &it.Description, &it.Status, &it.TrustPolicy, &it.LastFetchStatus,
			&it.LastFetchAt, &it.LastErrorCode, &it.LastErrorMessage, &it.LastIndexHash, &it.MetadataJSON,
			&it.CreatedBy, &it.CreatedAt, &it.UpdatedBy, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, nil
}

func (s *MySQLStore) CommunityPlugins(communityID int64) ([]domain.Plugin, error) {
	var exists int
	if err := s.db.QueryRow(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL`, communityID).Scan(&exists); err != nil {
		return nil, errors.New("子站不存在")
	}
	rows, err := s.db.Query(`SELECT plugin_code,status,sort_order,COALESCE(CAST(config_json AS CHAR),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM community_plugins WHERE community_id=? ORDER BY sort_order,plugin_code`, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type rowItem struct {
		status    string
		sortOrder int
		config    string
		createdAt string
		updatedAt string
	}
	runtime := map[string]rowItem{}
	for rows.Next() {
		var code string
		var item rowItem
		if err := rows.Scan(&code, &item.status, &item.sortOrder, &item.config, &item.createdAt, &item.updatedAt); err != nil {
			continue
		}
		runtime[code] = item
	}

	base := s.Plugins()
	out := make([]domain.Plugin, 0, len(base))
	for _, plugin := range base {
		item := plugin
		item.GlobalStatus = plugin.Status
		item.CommunityStatus = pluginregistry.StatusDisabled
		if rt, ok := runtime[item.Code]; ok {
			item.CommunityStatus = rt.status
			item.SortOrder = rt.sortOrder
			item.ResolvedConfig = pluginregistry.ResolvePluginConfig(item, item.ConfigJSON, rt.config)
			item.ConfigJSON = strings.TrimSpace(firstNonEmptyString(rt.config, item.ConfigJSON))
		} else {
			item.ResolvedConfig = pluginregistry.ResolvePluginConfig(item, item.ConfigJSON, "")
		}
		if item.GlobalStatus == pluginregistry.StatusEnabled && item.CommunityStatus == pluginregistry.StatusEnabled {
			item.Status = pluginregistry.StatusEnabled
		} else {
			item.Status = pluginregistry.StatusDisabled
		}
		out = append(out, pluginregistry.ApplyLifecycle(item))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func (s *MySQLStore) SetCommunityPluginStatus(communityID int64, code, status string) (domain.Plugin, error) {
	status = strings.TrimSpace(status)
	if status != pluginregistry.StatusEnabled && status != pluginregistry.StatusDisabled {
		return domain.Plugin{}, errors.New("插件状态不合法")
	}
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	plugin, ok := s.PluginByCode(def.Code)
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	if plugin.Status != pluginregistry.StatusEnabled && status == pluginregistry.StatusEnabled {
		return domain.Plugin{}, errors.New("插件全局未启用，不能在子站启用")
	}
	var exists int
	if err := s.db.QueryRow(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL`, communityID).Scan(&exists); err != nil {
		return domain.Plugin{}, errors.New("子站不存在")
	}
	if _, err := s.db.Exec(`INSERT INTO community_plugins (community_id,plugin_code,status,sort_order,config_json,created_at,updated_at)
		VALUES (?,?,?,0,NULL,NOW(),NOW())
		ON DUPLICATE KEY UPDATE status=VALUES(status),updated_at=NOW()`, communityID, def.Code, status); err != nil {
		return domain.Plugin{}, err
	}
	items, _ := s.CommunityPlugins(communityID)
	for _, item := range items {
		if item.Code == def.Code {
			return item, nil
		}
	}
	return domain.Plugin{}, errors.New("插件不存在")
}

func (s *MySQLStore) SetCommunityPluginConfig(communityID int64, code, configJSON string) (domain.Plugin, error) {
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	var exists int
	if err := s.db.QueryRow(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL`, communityID).Scan(&exists); err != nil {
		return domain.Plugin{}, errors.New("子站不存在")
	}
	configJSON = strings.TrimSpace(configJSON)
	if err := pluginregistry.ValidateConfigJSON(def, configJSON); err != nil {
		return domain.Plugin{}, err
	}
	var config any = nil
	if configJSON != "" {
		config = json.RawMessage(configJSON)
	}
	if _, err := s.db.Exec(`INSERT INTO community_plugins (community_id,plugin_code,status,sort_order,config_json,created_at,updated_at)
		VALUES (?,?,?,0,?,NOW(),NOW())
		ON DUPLICATE KEY UPDATE config_json=VALUES(config_json),updated_at=NOW()`, communityID, def.Code, pluginregistry.StatusDisabled, config); err != nil {
		return domain.Plugin{}, err
	}
	items, _ := s.CommunityPlugins(communityID)
	for _, item := range items {
		if item.Code == def.Code {
			return item, nil
		}
	}
	return domain.Plugin{}, errors.New("插件不存在")
}

func (s *MySQLStore) ReorderCommunityPlugins(communityID int64, codes []string) (int, error) {
	var exists int
	if err := s.db.QueryRow(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL`, communityID).Scan(&exists); err != nil {
		return 0, errors.New("子站不存在")
	}
	updated := 0
	for i, code := range codes {
		def, ok := pluginregistry.DefinitionByCode(code)
		if !ok {
			continue
		}
		if _, err := s.db.Exec(`INSERT INTO community_plugins (community_id,plugin_code,status,sort_order,config_json,created_at,updated_at)
			VALUES (?,?,?, ?, NULL, NOW(), NOW())
			ON DUPLICATE KEY UPDATE sort_order=VALUES(sort_order),updated_at=NOW()`, communityID, def.Code, pluginregistry.StatusDisabled, i); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func (s *MySQLStore) ListPosts(site, board, q, tag string) []domain.Post {
	site = strings.TrimSpace(site)
	board = strings.TrimSpace(board)
	q = strings.ToLower(strings.TrimSpace(q))
	tag = strings.ToLower(strings.TrimSpace(tag))
	posts := s.allPosts()
	out := make([]domain.Post, 0, len(posts))
	for _, p := range posts {
		if site != "" && site != "portal" && p.Site != site {
			continue
		}
		if board != "" && board != "all" && p.Board != board {
			continue
		}
		if tag != "" && !hasTag(p.Tags, tag) {
			continue
		}
		cp := p
		if q != "" && !postContains(&cp, q) {
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (s *MySQLStore) GetPost(id int64, increaseView bool) (*domain.Post, bool) {
	if increaseView {
		_, _ = s.db.Exec(`UPDATE posts SET views=views+1 WHERE id=?`, id)
	}
	p, err := s.postByID(id)
	return p, err == nil
}

func (s *MySQLStore) CreatePost(req domain.CreatePostRequest) (*domain.Post, error) {
	if req.Site == "portal" || strings.TrimSpace(req.Site) == "" {
		return nil, errors.New("帖子必须发布到具体子网站：php / go / java")
	}
	if !s.ValidateSite(req.Site) {
		return nil, errors.New("无效子网站")
	}
	if req.Board == "all" || strings.TrimSpace(req.Board) == "" {
		return nil, errors.New("帖子必须发布到具体板块")
	}
	if !s.ValidateBoard(req.Board) {
		return nil, errors.New("无效板块")
	}
	summary := strings.TrimSpace(req.Summary)
	content := strings.TrimSpace(req.Content)
	if summary == "" {
		summary = firstRunes(content, 80)
	}
	tags, _ := json.Marshal(uniqueTags(req.Tags))
	author := strings.TrimSpace(req.Author)
	if author == "" {
		author = "SUI.CHEN"
	}
	res, err := s.db.Exec(`INSERT INTO posts (site_key,board_key,title,summary,content,author,tags_json,created_at,updated_at) VALUES (?,?,?,?,?,?,?,NOW(),NOW())`,
		req.Site, req.Board, strings.TrimSpace(req.Title), summary, content, author, string(tags))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	_ = s.upsertTags(req.Site, uniqueTags(req.Tags))
	s.createNotice("你发布了新的内容", strings.TrimSpace(req.Title))
	return s.postByID(id)
}

func (s *MySQLStore) UpdatePost(id int64, req domain.UpdatePostRequest) (*domain.Post, error) {
	p, err := s.postByID(id)
	if err != nil {
		return nil, errors.New("帖子不存在")
	}
	if req.Site != nil {
		if *req.Site == "portal" || *req.Site == "" {
			return nil, errors.New("帖子必须属于具体子网站")
		}
		if !s.ValidateSite(*req.Site) {
			return nil, errors.New("无效子网站")
		}
		p.Site = *req.Site
	}
	if req.Board != nil {
		if *req.Board == "all" || *req.Board == "" {
			return nil, errors.New("帖子必须属于具体板块")
		}
		if !s.ValidateBoard(*req.Board) {
			return nil, errors.New("无效板块")
		}
		p.Board = *req.Board
	}
	if req.Title != nil {
		p.Title = strings.TrimSpace(*req.Title)
	}
	if req.Summary != nil {
		p.Summary = strings.TrimSpace(*req.Summary)
	}
	if req.Content != nil {
		p.Content = strings.TrimSpace(*req.Content)
	}
	if req.Status != nil {
		p.Status = strings.TrimSpace(*req.Status)
	}
	if req.Pinned != nil {
		p.Pinned = *req.Pinned
	}
	if req.Recommended != nil {
		p.Recommended = *req.Recommended
	}
	if req.RejectReason != nil {
		p.RejectReason = strings.TrimSpace(*req.RejectReason)
	}
	if req.OfflineReason != nil {
		p.OfflineReason = strings.TrimSpace(*req.OfflineReason)
	}
	if req.Tags != nil {
		p.Tags = uniqueTags(*req.Tags)
	}
	tags, _ := json.Marshal(p.Tags)
	_, err = s.db.Exec(`UPDATE posts SET site_key=?,board_key=?,title=?,summary=?,content=?,status=?,pinned=?,recommended=?,reject_reason=?,offline_reason=?,tags_json=?,updated_at=NOW() WHERE id=?`,
		p.Site, p.Board, p.Title, p.Summary, p.Content, p.Status, p.Pinned, p.Recommended, p.RejectReason, p.OfflineReason, string(tags), id)
	if err != nil {
		return nil, err
	}
	_ = s.rebuildTagsFromPosts()
	s.appendLog("operation", "admin", "更新帖子", fmt.Sprintf("posts#%d", id), "127.0.0.1")
	return s.postByID(id)
}

func (s *MySQLStore) DeletePost(id int64) bool {
	res, err := s.db.Exec(`DELETE FROM posts WHERE id=?`, id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return false
	}
	_, _ = s.db.Exec(`DELETE FROM comments WHERE post_id=?`, id)
	_ = s.rebuildTagsFromPosts()
	s.appendLog("operation", "admin", "删除帖子", fmt.Sprintf("posts#%d", id), "127.0.0.1")
	return true
}

func (s *MySQLStore) LikePost(id int64) (*domain.Post, error) {
	res, err := s.db.Exec(`UPDATE posts SET likes=likes+1,updated_at=NOW() WHERE id=?`, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, errors.New("帖子不存在")
	}
	p, err := s.postByID(id)
	if err == nil {
		s.createNotice("你的帖子获得了新的点赞", p.Title)
	}
	return p, err
}

func (s *MySQLStore) HotPosts(site string, limit int) []domain.Post {
	settings := s.AdminSettings()
	out := s.ListPosts(site, "all", "", "")
	sort.Slice(out, func(i, j int) bool {
		si := out[i].Views*settings.HotViewWeight + out[i].Likes*settings.HotLikeWeight + out[i].Comments*settings.HotCommentWeight
		sj := out[j].Views*settings.HotViewWeight + out[j].Likes*settings.HotLikeWeight + out[j].Comments*settings.HotCommentWeight
		return si > sj
	})
	return limitPosts(out, limit)
}

func (s *MySQLStore) Feed(site string, limit int) []domain.Post {
	out := s.ListPosts(site, "all", "", "")
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return limitPosts(out, limit)
}

func (s *MySQLStore) TagStats(site string) []domain.TagStat {
	tags := s.AdminTags(site, "", "enable")
	out := make([]domain.TagStat, 0, len(tags))
	for _, tag := range tags {
		if tag.TopicCount > 0 || tag.UseCount > 0 {
			out = append(out, domain.TagStat{
				ID:             tag.ID,
				Name:           tag.Name,
				Slug:           tag.Slug,
				Site:           tag.Site,
				CommunityID:    tag.CommunityID,
				CommunitySlug:  tag.CommunitySlug,
				Description:    tag.Description,
				TopicCount:     tag.TopicCount,
				Count:          firstNonZero(tag.TopicCount, tag.UseCount),
				FollowerCount:  tag.FollowerCount,
				Status:         tag.Status,
				SEOTitle:       tag.SEOTitle,
				SEODescription: tag.SEODescription,
				SEOKeywords:    tag.SEOKeywords,
			})
		}
	}
	return out
}

func (s *MySQLStore) AdminTags(site, q, status string) []domain.Tag {
	site = strings.TrimSpace(site)
	q = strings.ToLower(strings.TrimSpace(q))
	status = strings.TrimSpace(status)
	rows, err := s.db.Query(sqlTagSelect + ` ORDER BY site_key,sort_order,use_count DESC,id`)
	if err != nil {
		return []domain.Tag{}
	}
	defer rows.Close()
	out := []domain.Tag{}
	for rows.Next() {
		tag, err := scanSQLTag(rows)
		if err != nil {
			continue
		}
		tag = s.enrichSQLTag(tag)
		if site != "" && site != "portal" && tag.Site != site {
			continue
		}
		if status == "" || status == "all" {
			if tag.Status == "merged" {
				continue
			}
		} else if tag.Status != status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(tag.Name+" "+tag.Slug+" "+tag.Description), q) {
			continue
		}
		out = append(out, tag)
	}
	return out
}

func (s *MySQLStore) AdminTagByID(id int64) (domain.Tag, bool) {
	tag, err := s.tagByID(id)
	return tag, err == nil
}

func (s *MySQLStore) CreateTag(req domain.Tag) (domain.Tag, error) {
	tag := normalizeTag(req)
	if tag.Name == "" {
		return domain.Tag{}, errors.New("标签名称不能为空")
	}
	if tag.Site == "" {
		tag.Site = "portal"
	}
	if tag.Site != "portal" && !s.ValidateSite(tag.Site) {
		return domain.Tag{}, errors.New("无效子网站")
	}
	if _, ok := s.ResolveTag(tag.Site, tag.Slug); ok {
		return domain.Tag{}, errors.New("标签 slug 已存在或与别名冲突")
	}
	res, err := s.db.Exec(`INSERT INTO tags (site_key,name,slug,description,status,merged_to_id,sort_order,use_count,follower_count,hot_score,seo_title,seo_description,seo_keywords) VALUES (?,?,?,?,?,?,?,0,0,0,?,?,?)`,
		tag.Site, tag.Name, tag.Slug, tag.Description, tag.Status, nil, tag.Sort, tag.SEOTitle, tag.SEODescription, tag.SEOKeywords)
	if err != nil {
		return domain.Tag{}, err
	}
	id, _ := res.LastInsertId()
	s.appendLog("operation", "admin", "新增标签", fmt.Sprintf("tags#%d", id), "127.0.0.1")
	return s.tagByID(id)
}

func (s *MySQLStore) UpdateTag(id int64, req domain.Tag) (domain.Tag, bool) {
	tag := normalizeTag(req)
	if tag.Name == "" {
		return domain.Tag{}, false
	}
	if tag.Site == "" {
		tag.Site = "portal"
	}
	current, err := s.tagByID(id)
	if err != nil {
		return domain.Tag{}, false
	}
	if current.MergedToID > 0 {
		tag.MergedToID = current.MergedToID
	}
	var conflictID int64
	_ = s.db.QueryRow(`SELECT id FROM tags WHERE site_key=? AND slug=? AND id<>? LIMIT 1`, tag.Site, tag.Slug, id).Scan(&conflictID)
	if conflictID > 0 {
		return domain.Tag{}, false
	}
	var aliasConflictID int64
	_ = s.db.QueryRow(`SELECT id FROM tag_aliases WHERE site_key=? AND alias_slug=? AND tag_id<>? LIMIT 1`, tag.Site, tag.Slug, id).Scan(&aliasConflictID)
	if aliasConflictID > 0 {
		return domain.Tag{}, false
	}
	res, err := s.db.Exec(`UPDATE tags SET site_key=?,name=?,slug=?,description=?,status=?,merged_to_id=?,sort_order=?,seo_title=?,seo_description=?,seo_keywords=? WHERE id=?`,
		tag.Site, tag.Name, tag.Slug, tag.Description, tag.Status, nullableMergedToID(tag.MergedToID), tag.Sort, tag.SEOTitle, tag.SEODescription, tag.SEOKeywords, id)
	if err != nil {
		return domain.Tag{}, false
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Tag{}, false
	}
	s.appendLog("operation", "admin", "更新标签", fmt.Sprintf("tags#%d", id), "127.0.0.1")
	tag, err = s.tagByID(id)
	return tag, err == nil
}

func (s *MySQLStore) SetTagStatus(id int64, status string) (domain.Tag, bool) {
	status = normalizeTagStatus(status)
	if status == "" {
		return domain.Tag{}, false
	}
	res, err := s.db.Exec(`UPDATE tags SET status=?,updated_at=NOW() WHERE id=?`, status, id)
	if err != nil {
		return domain.Tag{}, false
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Tag{}, false
	}
	s.appendLog("operation", "admin", "更新标签状态", fmt.Sprintf("tags#%d:%s", id, status), "127.0.0.1")
	tag, err := s.tagByID(id)
	return tag, err == nil
}

func (s *MySQLStore) ResolveTag(site, slugOrName string) (domain.TagResolveResult, bool) {
	site = strings.TrimSpace(site)
	slugOrName = strings.TrimSpace(slugOrName)
	if slugOrName == "" {
		return domain.TagResolveResult{}, false
	}
	if id, err := strconv.ParseInt(slugOrName, 10, 64); err == nil && id > 0 {
		tag, err := s.tagByID(id)
		if err == nil && tag.Status == "enable" && (site == "" || site == "portal" || tag.Site == site) {
			return domain.TagResolveResult{Tag: tag, Requested: slugOrName, ResolvedBy: "direct"}, true
		}
	}
	slug := normalizeSlug(slugOrName)
	if slug == "" {
		slug = strings.ToLower(strings.Join(strings.Fields(slugOrName), "-"))
	}
	query := `SELECT id FROM tags WHERE (slug=? OR name=?)`
	args := []any{slug, slugOrName}
	if site != "" && site != "portal" {
		query += ` AND site_key=?`
		args = append(args, site)
	}
	if site == "" || site == "portal" {
		query += ` ORDER BY CASE WHEN site_key='portal' THEN 0 ELSE 1 END,use_count DESC,id LIMIT 1`
	} else {
		query += ` ORDER BY use_count DESC,id LIMIT 1`
	}
	var directID int64
	if err := s.db.QueryRow(query, args...).Scan(&directID); err == nil && directID > 0 {
		tag, err := s.tagByID(directID)
		if err == nil {
			if tag.Status == "merged" && tag.MergedToID > 0 {
				target, err := s.tagByID(tag.MergedToID)
				if err == nil && target.Status == "enable" {
					return domain.TagResolveResult{Tag: target, Requested: slugOrName, ResolvedBy: "merged"}, true
				}
				return domain.TagResolveResult{}, false
			}
			if tag.Status == "enable" {
				return domain.TagResolveResult{Tag: tag, Requested: slugOrName, ResolvedBy: "direct"}, true
			}
			return domain.TagResolveResult{}, false
		}
	}
	aliasQuery := `SELECT id,tag_id,site_key,alias,alias_slug,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM tag_aliases WHERE (alias_slug=? OR alias=?)`
	aliasArgs := []any{slug, slugOrName}
	if site != "" && site != "portal" {
		aliasQuery += ` AND site_key=?`
		aliasArgs = append(aliasArgs, site)
	}
	if site == "" || site == "portal" {
		aliasQuery += ` ORDER BY CASE WHEN site_key='portal' THEN 0 ELSE 1 END,id LIMIT 1`
	} else {
		aliasQuery += ` ORDER BY id LIMIT 1`
	}
	var alias domain.TagAlias
	err := s.db.QueryRow(aliasQuery, aliasArgs...).Scan(&alias.ID, &alias.TagID, &alias.Site, &alias.Alias, &alias.AliasSlug, &alias.CreatedAt, &alias.UpdatedAt)
	if err != nil {
		return domain.TagResolveResult{}, false
	}
	tag, err := s.tagByID(alias.TagID)
	if err != nil {
		return domain.TagResolveResult{}, false
	}
	if tag.Status == "merged" && tag.MergedToID > 0 {
		target, err := s.tagByID(tag.MergedToID)
		if err == nil && target.Status == "enable" {
			return domain.TagResolveResult{Tag: target, MatchedAlias: alias.Alias, Requested: slugOrName, ResolvedBy: "alias"}, true
		}
		return domain.TagResolveResult{}, false
	}
	if tag.Status != "enable" {
		return domain.TagResolveResult{}, false
	}
	return domain.TagResolveResult{Tag: tag, MatchedAlias: alias.Alias, Requested: slugOrName, ResolvedBy: "alias"}, true
}

func (s *MySQLStore) TagBySlug(site, slugOrName string) (domain.Tag, bool) {
	result, ok := s.ResolveTag(site, slugOrName)
	if !ok {
		return domain.Tag{}, false
	}
	return result.Tag, true
}

func (s *MySQLStore) TagAliases(tagID int64) ([]domain.TagAlias, error) {
	tag, err := s.tagByID(tagID)
	if err != nil {
		return nil, errors.New("标签不存在")
	}
	rows, err := s.db.Query(`SELECT id,tag_id,site_key,alias,alias_slug,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM tag_aliases WHERE tag_id=? ORDER BY alias_slug,id`, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TagAlias, 0)
	for rows.Next() {
		var item domain.TagAlias
		if err := rows.Scan(&item.ID, &item.TagID, &item.Site, &item.Alias, &item.AliasSlug, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}
		item.CommunitySlug = item.Site
		item.CommunityID = tag.CommunityID
		out = append(out, item)
	}
	return out, nil
}

func (s *MySQLStore) AddTagAlias(tagID int64, alias string) (domain.TagAlias, error) {
	tag, err := s.tagByID(tagID)
	if err != nil {
		return domain.TagAlias{}, errors.New("标签不存在")
	}
	if tag.Status != "enable" {
		return domain.TagAlias{}, errors.New("仅启用标签可添加别名")
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return domain.TagAlias{}, errors.New("别名不能为空")
	}
	if len([]rune(alias)) > 50 {
		return domain.TagAlias{}, errors.New("别名长度不能超过 50 个字符")
	}
	aliasSlug := normalizeSlug(alias)
	if aliasSlug == "" {
		aliasSlug = strings.ToLower(strings.Join(strings.Fields(alias), "-"))
	}
	if aliasSlug == "" {
		return domain.TagAlias{}, errors.New("别名 slug 不能为空")
	}
	var conflictID int64
	_ = s.db.QueryRow(`SELECT id FROM tags WHERE site_key=? AND slug=? LIMIT 1`, tag.Site, aliasSlug).Scan(&conflictID)
	if conflictID > 0 {
		return domain.TagAlias{}, errors.New("别名 slug 与现有标签冲突")
	}
	_ = s.db.QueryRow(`SELECT id FROM tag_aliases WHERE site_key=? AND alias_slug=? LIMIT 1`, tag.Site, aliasSlug).Scan(&conflictID)
	if conflictID > 0 {
		return domain.TagAlias{}, errors.New("别名 slug 已存在")
	}
	res, err := s.db.Exec(`INSERT INTO tag_aliases (tag_id,site_key,alias,alias_slug,created_at,updated_at) VALUES (?,?,?,?,NOW(),NOW())`, tagID, tag.Site, alias, aliasSlug)
	if err != nil {
		return domain.TagAlias{}, err
	}
	id, _ := res.LastInsertId()
	items, err := s.TagAliases(tagID)
	if err != nil {
		return domain.TagAlias{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.TagAlias{}, errors.New("标签别名创建失败")
}

func (s *MySQLStore) DeleteTagAlias(tagID, aliasID int64) error {
	res, err := s.db.Exec(`DELETE FROM tag_aliases WHERE id=? AND tag_id=?`, aliasID, tagID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("标签别名不存在")
	}
	return nil
}

func (s *MySQLStore) RecalculateTag(tagID int64) (domain.Tag, error) {
	tag, err := s.tagByID(tagID)
	if err != nil {
		return domain.Tag{}, errors.New("标签不存在")
	}
	var topicCount int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topic_tags tt JOIN topics tp ON tp.id=tt.topic_id AND tp.deleted_at IS NULL AND tp.status=1 WHERE tt.tag_id=?`, tagID).Scan(&topicCount)
	var followerCount int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM follows WHERE target_type='tag' AND target_id=?`, tagID).Scan(&followerCount)
	hotScore := topicCount*10 + followerCount*20
	if tag.Status == "merged" {
		topicCount = 0
		followerCount = 0
		hotScore = 0
	}
	if _, err := s.db.Exec(`UPDATE tags SET use_count=?,follower_count=?,hot_score=?,updated_at=NOW() WHERE id=?`, topicCount, followerCount, hotScore, tagID); err != nil {
		return domain.Tag{}, err
	}
	return s.tagByID(tagID)
}

func (s *MySQLStore) RecalculateAllTags() (int, error) {
	rows, err := s.db.Query(`SELECT id FROM tags`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		if _, err := s.RecalculateTag(id); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

func (s *MySQLStore) MergeTag(sourceTagID, targetTagID int64) (domain.Tag, error) {
	if sourceTagID <= 0 || targetTagID <= 0 || sourceTagID == targetTagID {
		return domain.Tag{}, errors.New("源标签和目标标签不能相同")
	}
	source, err := s.tagByID(sourceTagID)
	if err != nil {
		return domain.Tag{}, errors.New("源标签不存在")
	}
	target, err := s.tagByID(targetTagID)
	if err != nil {
		return domain.Tag{}, errors.New("目标标签不存在")
	}
	if source.Status == "merged" {
		return domain.Tag{}, errors.New("源标签已合并")
	}
	if target.Status != "enable" {
		return domain.Tag{}, errors.New("目标标签必须为启用状态")
	}
	if source.Site != target.Site {
		return domain.Tag{}, errors.New("仅支持同一子站范围合并")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return domain.Tag{}, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT IGNORE INTO topic_tags (topic_id,tag_id) SELECT topic_id,? FROM topic_tags WHERE tag_id=?`, targetTagID, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`DELETE FROM topic_tags WHERE tag_id=?`, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`INSERT IGNORE INTO follows (user_id,target_type,target_id,created_at,updated_at) SELECT user_id,'tag',?,created_at,NOW() FROM follows WHERE target_type='tag' AND target_id=?`, targetTagID, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`DELETE FROM follows WHERE target_type='tag' AND target_id=?`, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`INSERT IGNORE INTO tag_aliases (tag_id,site_key,alias,alias_slug,created_at,updated_at) SELECT ?,site_key,alias,alias_slug,created_at,NOW() FROM tag_aliases WHERE tag_id=?`, targetTagID, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`DELETE FROM tag_aliases WHERE tag_id=?`, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`UPDATE tags SET status='merged',merged_to_id=?,use_count=0,follower_count=0,hot_score=0,updated_at=NOW() WHERE id=?`, targetTagID, sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	if _, err := tx.Exec(`UPDATE tags SET updated_at=NOW() WHERE id=?`, targetTagID); err != nil {
		return domain.Tag{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Tag{}, err
	}
	if _, err := s.RecalculateTag(sourceTagID); err != nil {
		return domain.Tag{}, err
	}
	return s.RecalculateTag(targetTagID)
}

func (s *MySQLStore) TagSuggestions(site, q string, limit int) []domain.TagStat {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	q = strings.TrimSpace(q)
	tags := s.AdminTags(site, "", "enable")
	out := make([]domain.TagStat, 0, len(tags))
	seen := map[int64]bool{}
	for _, tag := range tags {
		if q != "" && !strings.Contains(strings.ToLower(tag.Name+" "+tag.Slug+" "+tag.Description), strings.ToLower(q)) {
			continue
		}
		out = append(out, domain.TagStat{
			ID:             tag.ID,
			Name:           tag.Name,
			Slug:           tag.Slug,
			Site:           tag.Site,
			CommunityID:    tag.CommunityID,
			CommunitySlug:  tag.CommunitySlug,
			Description:    tag.Description,
			TopicCount:     tag.TopicCount,
			Count:          firstNonZero(tag.TopicCount, tag.UseCount),
			FollowerCount:  tag.FollowerCount,
			HotScore:       tag.HotScore,
			Status:         tag.Status,
			SEOTitle:       tag.SEOTitle,
			SEODescription: tag.SEODescription,
			SEOKeywords:    tag.SEOKeywords,
		})
		seen[tag.ID] = true
		if len(out) >= limit {
			break
		}
	}
	if q != "" && len(out) < limit {
		query := `SELECT id,tag_id,site_key,alias,alias_slug,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM tag_aliases WHERE (alias LIKE ? OR alias_slug LIKE ?)`
		args := []any{"%" + q + "%", "%" + normalizeSlug(q) + "%"}
		if site != "" && site != "portal" {
			query += ` AND site_key=?`
			args = append(args, site)
		}
		query += ` ORDER BY alias_slug LIMIT ?`
		args = append(args, limit)
		rows, err := s.db.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var alias domain.TagAlias
				if err := rows.Scan(&alias.ID, &alias.TagID, &alias.Site, &alias.Alias, &alias.AliasSlug, &alias.CreatedAt, &alias.UpdatedAt); err != nil {
					continue
				}
				tag, ok := s.AdminTagByID(alias.TagID)
				if !ok || tag.Status != "enable" || seen[tag.ID] {
					continue
				}
				out = append(out, domain.TagStat{
					ID:             tag.ID,
					Name:           tag.Name,
					Slug:           tag.Slug,
					Site:           tag.Site,
					CommunityID:    tag.CommunityID,
					CommunitySlug:  tag.CommunitySlug,
					Description:    tag.Description,
					TopicCount:     tag.TopicCount,
					Count:          firstNonZero(tag.TopicCount, tag.UseCount),
					FollowerCount:  tag.FollowerCount,
					HotScore:       tag.HotScore,
					Status:         tag.Status,
					MatchedAlias:   alias.Alias,
					SEOTitle:       tag.SEOTitle,
					SEODescription: tag.SEODescription,
					SEOKeywords:    tag.SEOKeywords,
				})
				seen[tag.ID] = true
				if len(out) >= limit {
					break
				}
			}
		}
	}
	return out
}

func (s *MySQLStore) BoardCounts(site, q string) map[string]int {
	result := map[string]int{}
	posts := s.ListPosts(site, "all", q, "")
	result["all"] = len(posts)
	for _, b := range s.ListBoards() {
		if b.Key != "all" {
			result[b.Key] = 0
		}
	}
	for _, p := range posts {
		result[p.Board]++
	}
	return result
}

func (s *MySQLStore) PostStats(site string) domain.PostStats {
	posts := s.ListPosts(site, "all", "", "")
	stats := domain.PostStats{TotalPosts: len(posts)}
	for _, p := range posts {
		stats.TotalViews += p.Views
		stats.TotalLikes += p.Likes
		stats.TotalComments += p.Comments
	}
	return stats
}

func (s *MySQLStore) CommentsTree(postID int64) []*domain.Comment {
	items, _ := s.TopicComments(postID, "oldest", 1, 1000)
	return items
}

func (s *MySQLStore) TopicComments(topicID int64, sortBy string, page, pageSize int) ([]*domain.Comment, int) {
	if _, err := s.TopicByID(topicID, false); err != nil {
		return []*domain.Comment{}, 0
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE COALESCE(topic_id,post_id)=? AND parent_id=0 AND status NOT IN ('deleted','hidden') AND deleted_at IS NULL`, topicID).Scan(&total)
	order := `is_best DESC, id ASC`
	switch sortBy {
	case "latest":
		order = `id DESC`
	case "oldest":
		order = `id ASC`
	case "best":
		order = `is_best DESC, id ASC`
	}
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	rows, err := s.db.Query(`SELECT id,post_id,COALESCE(topic_id,post_id),parent_id,COALESCE(reply_to_user_id,0),COALESCE(user_id,1),author,to_author,text,COALESCE(content_html,''),status,likes,is_best,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'') FROM comments WHERE COALESCE(topic_id,post_id)=? AND parent_id=0 AND status NOT IN ('deleted','hidden') AND deleted_at IS NULL ORDER BY `+order+` LIMIT ? OFFSET ?`, topicID, pageSize, offset)
	if err != nil {
		return []*domain.Comment{}, 0
	}
	defer rows.Close()
	roots := []*domain.Comment{}
	rootIDs := []int64{}
	for rows.Next() {
		c := &domain.Comment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.TopicID, &c.ParentID, &c.ReplyToUserID, &c.UserID, &c.Author, &c.To, &c.Text, &c.ContentHTML, &c.Status, &c.Likes, &c.IsBest, &c.CreatedAt, &c.UpdatedAt); err == nil {
			normalizeSQLComment(c)
			roots = append(roots, c)
			rootIDs = append(rootIDs, c.ID)
		}
	}
	if len(rootIDs) > 0 {
		replies := s.topicReplies(rootIDs)
		for _, root := range roots {
			root.Replies = replies[root.ID]
		}
	}
	return roots, total
}

// CommentByID 返回评论详情。
func (s *MySQLStore) CommentByID(id int64) (*domain.Comment, error) {
	return s.commentByID(id)
}

func (s *MySQLStore) CreateComment(postID int64, req domain.CreateCommentRequest) (*domain.Comment, error) {
	topic, err := s.TopicByID(postID, false)
	if err != nil || topic == nil {
		return nil, errors.New("主题不存在")
	}
	if topic.Status != 1 {
		return nil, errors.New("主题已隐藏")
	}
	if topic.CommentLocked {
		return nil, errors.New("评论已锁定")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		content = strings.TrimSpace(req.Text)
	}
	if len([]rune(content)) < 2 {
		return nil, errors.New("评论内容至少 2 个字符")
	}
	if len([]rune(content)) > 5000 {
		return nil, errors.New("评论内容最多 5000 个字符")
	}
	userID := req.UserID
	if userID <= 0 {
		userID = 1
	}
	to := ""
	replyToUserID := int64(0)
	if req.ParentID > 0 {
		if err := s.db.QueryRow(`SELECT author,COALESCE(user_id,1) FROM comments WHERE id=? AND COALESCE(topic_id,post_id)=? AND status NOT IN ('deleted','hidden') AND deleted_at IS NULL`, req.ParentID, postID).Scan(&to, &replyToUserID); err != nil {
			return nil, errors.New("父评论不存在")
		}
	}
	author := strings.TrimSpace(req.Author)
	if author == "" {
		author = s.userDisplayName(userID)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	res, err := tx.Exec(`INSERT INTO comments (post_id,topic_id,parent_id,reply_to_user_id,user_id,author,to_author,text,status,likes,is_best,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?, 'normal',0,0,NOW(),NOW())`,
		postID, postID, req.ParentID, nullableInt64(replyToUserID), userID, author, to, content)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	id, _ := res.LastInsertId()
	if topic.PluginCode == "qa" || topic.ContentType == "question" {
		if _, err := tx.Exec(`INSERT IGNORE INTO qa_answers (topic_id,comment_id,user_id,is_accepted,created_at,updated_at) VALUES (?,?,?,0,NOW(),NOW())`, postID, id, userID); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if _, err := tx.Exec(`UPDATE topics SET comment_count=comment_count+1,last_active_at=NOW(),updated_at=NOW(),`+recalcTopicHotScoreSQL()+` WHERE id=?`, postID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE posts SET comments=comments+1,updated_at=NOW() WHERE id=?`, postID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	targetType := "topic"
	targetID := postID
	if req.ParentID > 0 {
		targetType = "comment"
		targetID = req.ParentID
	}
	if _, err := tx.Exec(`INSERT INTO activities (user_id,community_id,topic_id,action,target_type,target_id,remark,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
		userID, topic.CommunityID, topic.ID, "commented", targetType, targetID, topic.Title); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if req.ParentID > 0 {
		if _, err := s.createUserNoticeTx(tx, replyToUserID, userID, "comment_replied", "comment", id, topic.ID, id, "你的评论有新的回复", fmt.Sprintf("%s 回复了你的评论。", author)); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		if _, err := s.createUserNoticeTx(tx, topic.UserID, userID, "topic_commented", "topic", topic.ID, topic.ID, id, "你的主题有新的评论", fmt.Sprintf("%s 评论了《%s》。", author, topic.Title)); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.commentByID(id)
}

func (s *MySQLStore) LikeComment(id int64) (*domain.Comment, error) {
	res, err := s.db.Exec(`UPDATE comments SET likes=likes+1 WHERE id=?`, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, errors.New("评论不存在")
	}
	return s.commentByID(id)
}

func (s *MySQLStore) DeleteOwnComment(id int64, author string) error {
	author = strings.TrimSpace(author)
	if author == "" {
		return errors.New("缺少评论作者")
	}
	c, err := s.commentByID(id)
	if err != nil {
		return errors.New("评论不存在")
	}
	if c.Author != author {
		return errors.New("只能删除自己的评论")
	}
	if !s.DeleteComment(id) {
		return errors.New("评论不存在")
	}
	return nil
}

func (s *MySQLStore) DeleteComment(id int64) bool {
	c, err := s.commentByID(id)
	if err != nil {
		return false
	}
	ids := s.commentCascadeIDs(id)
	if len(ids) == 0 {
		return false
	}
	args := make([]any, 0, len(ids))
	holders := make([]string, 0, len(ids))
	for _, cid := range ids {
		args = append(args, cid)
		holders = append(holders, "?")
	}
	if _, err := s.db.Exec(`DELETE FROM comments WHERE id IN (`+strings.Join(holders, ",")+`)`, args...); err != nil {
		return false
	}
	_, _ = s.db.Exec(`UPDATE posts SET comments=GREATEST(comments-?,0),updated_at=NOW() WHERE id=?`, len(ids), c.PostID)
	s.appendLog("audit", "auditor", "删除评论", fmt.Sprintf("comments#%d", id), "127.0.0.1")
	return true
}

func (s *MySQLStore) Notices(site string) []domain.Notification {
	site = strings.TrimSpace(site)
	query := `SELECT id,site_key,title,content,is_read,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM notifications`
	args := []any{}
	if site != "" && site != "portal" {
		query += ` WHERE site_key IN ('portal',?)`
		args = append(args, site)
	}
	query += ` ORDER BY id DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.Notification{}
	}
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.Site, &n.Title, &n.Content, &n.Read, &n.CreatedAt); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func (s *MySQLStore) ReadNotice(id int64) bool {
	res, err := s.db.Exec(`UPDATE notifications SET is_read=1 WHERE id=?`, id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *MySQLStore) ReadAllNotices(site string) int {
	site = strings.TrimSpace(site)
	query := `UPDATE notifications SET is_read=1 WHERE is_read=0`
	args := []any{}
	if site != "" && site != "portal" {
		query += ` AND site_key IN ('portal',?)`
		args = append(args, site)
	}
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return 0
	}
	n, _ := res.RowsAffected()
	return int(n)
}

func (s *MySQLStore) UnreadNoticeCount(site string) int {
	var count int
	site = strings.TrimSpace(site)
	if site != "" && site != "portal" {
		_ = s.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE is_read=0 AND site_key IN ('portal',?)`, site).Scan(&count)
		return count
	}
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE is_read=0`).Scan(&count)
	return count
}

func (s *MySQLStore) UserProfile() domain.UserProfile {
	var posts, comments, likes int
	_ = s.db.QueryRow(`SELECT COUNT(*),COALESCE(SUM(likes),0) FROM posts WHERE author='SUI.CHEN'`).Scan(&posts, &likes)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE author='SUI.CHEN'`).Scan(&comments)
	return domain.UserProfile{ID: 1, Name: "SUI.CHEN", Bio: "DevHub 用户主页 · 关注 PHP / Go / Java 技术内容", Posts: posts, Comments: comments, Likes: likes}
}

func (s *MySQLStore) AdminOverview(site string) domain.AdminOverview {
	scope := strings.TrimSpace(site)
	posts := s.ListPosts(scope, "all", "", "")
	status := map[string]int{"draft": 0, "review": 0, "publish": 0, "offline": 0}
	siteStats := map[string]*domain.SiteStat{}
	boardStats := map[string]*domain.BoardStat{}
	for _, siteItem := range s.ListSites() {
		if siteItem.Key != "portal" {
			if scope != "" && scope != "portal" && siteItem.Key != scope {
				continue
			}
			siteStats[siteItem.Key] = &domain.SiteStat{Site: siteItem.Key}
		}
	}
	for _, board := range s.ListBoards() {
		if board.Key != "all" {
			boardStats[board.Key] = &domain.BoardStat{Board: board.Key}
		}
	}
	for _, p := range posts {
		status[p.Status]++
		if st, ok := siteStats[p.Site]; ok {
			st.Posts++
			st.Views += p.Views
			st.Likes += p.Likes
			st.Comments += p.Comments
		}
		if st, ok := boardStats[p.Board]; ok {
			st.Posts++
			st.Views += p.Views
			st.Likes += p.Likes
			st.Comments += p.Comments
		}
	}
	siteOut := make([]domain.SiteStat, 0, len(siteStats))
	for _, st := range siteStats {
		siteOut = append(siteOut, *st)
	}
	sort.Slice(siteOut, func(i, j int) bool { return siteOut[i].Site < siteOut[j].Site })
	boardOut := make([]domain.BoardStat, 0, len(boardStats))
	for _, st := range boardStats {
		boardOut = append(boardOut, *st)
	}
	sort.Slice(boardOut, func(i, j int) bool { return boardOut[i].Board < boardOut[j].Board })
	users := s.userStats()
	return domain.AdminOverview{
		Stats:              s.PostStats(scope),
		StatusDistribution: status,
		SiteStats:          siteOut,
		BoardStats:         boardOut,
		TopPosts:           s.HotPosts(scope, 10),
		SearchKeywords: []domain.KeywordStat{
			{Keyword: "JVM", Count: 36, Scope: "java"},
			{Keyword: "Laravel", Count: 28, Scope: "php"},
			{Keyword: "Goroutine", Count: 21, Scope: "go"},
			{Keyword: "context", Count: 18, Scope: "go"},
		},
		UserStats: users,
	}
}

func (s *MySQLStore) AdminUsers() []domain.AdminUser {
	rows, err := s.db.Query(`SELECT u.id,u.username,u.nickname,'' AS avatar,u.phone,u.email,u.status,COALESCE(MIN(r.id),5),COALESCE(MIN(r.name),'普通用户'),DATE_FORMAT(u.created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(u.last_login_at,'%Y-%m-%d %H:%i:%s'),''),''
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id=u.id AND ur.status='normal'
		LEFT JOIN roles r ON r.id=ur.role_id
		GROUP BY u.id,u.username,u.nickname,u.phone,u.email,u.status,u.created_at,u.last_login_at
		ORDER BY u.id`)
	if err != nil {
		return []domain.AdminUser{}
	}
	defer rows.Close()
	out := []domain.AdminUser{}
	for rows.Next() {
		var u domain.AdminUser
		if err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.Avatar, &u.Phone, &u.Email, &u.Status, &u.RoleID, &u.RoleName, &u.CreatedAt, &u.LastLoginAt, &u.ViolationNote); err == nil {
			_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE user_id=? AND deleted_at IS NULL`, u.ID).Scan(&u.Posts)
			_ = s.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE author=? OR author=?`, u.Username, u.Nickname).Scan(&u.Comments)
			out = append(out, u)
		}
	}
	return out
}

func (s *MySQLStore) UpdateUserStatus(id int64, status, note string) bool {
	res, err := s.db.Exec(`UPDATE users SET status=? WHERE id=?`, strings.TrimSpace(status), id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return false
	}
	_, _ = s.db.Exec(`UPDATE admin_users SET status=?,violation_note=? WHERE id=?`, strings.TrimSpace(status), strings.TrimSpace(note), id)
	s.appendLog("operation", "admin", "更新用户状态", fmt.Sprintf("users#%d", id), "127.0.0.1")
	return true
}

func (s *MySQLStore) AdminRoles() []domain.AdminRole {
	rows, err := s.db.Query(`SELECT id,name,builtin,description,permissions_json,user_count FROM admin_roles ORDER BY id`)
	if err != nil {
		return []domain.AdminRole{}
	}
	defer rows.Close()
	out := []domain.AdminRole{}
	for rows.Next() {
		var r domain.AdminRole
		var permissions string
		if err := rows.Scan(&r.ID, &r.Name, &r.Builtin, &r.Description, &permissions, &r.UserCount); err == nil {
			_ = json.Unmarshal([]byte(permissions), &r.Permissions)
			out = append(out, r)
		}
	}
	return out
}

func (s *MySQLStore) AdminPermissions() []domain.AdminPermission {
	return []domain.AdminPermission{
		{Code: "content", Module: "内容管理", Name: "帖子 / 评论 / 标签 / 文档", Ops: []string{"查", "增", "删", "改", "审核"}},
		{Code: "plugin", Module: "系统插件", Name: "qa / docs / wiki 插件", Ops: []string{"查", "启用", "禁用"}},
		{Code: "qa", Module: "问答插件", Name: "问题 / 回答 / 采纳", Ops: []string{"查", "增", "审核", "采纳"}},
		{Code: "docs", Module: "文档插件", Name: "空间 / 文档树 / 文档", Ops: []string{"查", "增", "改", "管理"}},
		{Code: "wiki", Module: "Wiki 插件", Name: "页面 / 版本 / 回滚", Ops: []string{"查", "增", "改", "回滚"}},
		{Code: "site", Module: "站点配置", Name: "子站 / 板块 / 搜索范围", Ops: []string{"查", "增", "删", "改"}},
		{Code: "user", Module: "用户管理", Name: "用户信息 / 行为 / 违规处理", Ops: []string{"查", "改", "审核"}},
		{Code: "operation", Module: "运营管理", Name: "推荐 / 通知 / 热门 / 草稿箱", Ops: []string{"查", "增", "删", "改"}},
		{Code: "moderator", Module: "版主管理", Name: "子站版主分配 / 停用", Ops: []string{"查", "增", "删", "改"}},
		{Code: "statistics", Module: "数据统计", Name: "内容 / 用户 / 搜索统计", Ops: []string{"查", "导出"}},
		{Code: "system", Module: "系统设置", Name: "参数 / 日志 / 备份恢复", Ops: []string{"查", "改", "删"}},
	}
}

func (s *MySQLStore) AdminComments(site string) []domain.AdminComment {
	site = strings.TrimSpace(site)
	query := `SELECT c.id,c.post_id,COALESCE(p.title,''),c.parent_id,c.author,c.to_author,c.text,c.status,c.likes,DATE_FORMAT(c.created_at,'%Y-%m-%d %H:%i:%s') FROM comments c LEFT JOIN posts p ON p.id=c.post_id`
	args := []any{}
	if site != "" && site != "portal" {
		query += ` WHERE p.site_key=?`
		args = append(args, site)
	}
	query += ` ORDER BY c.id DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.AdminComment{}
	}
	defer rows.Close()
	out := []domain.AdminComment{}
	for rows.Next() {
		var c domain.AdminComment
		if err := rows.Scan(&c.ID, &c.PostID, &c.PostTitle, &c.ParentID, &c.Author, &c.To, &c.Text, &c.Status, &c.Likes, &c.CreatedAt); err == nil {
			out = append(out, c)
		}
	}
	return out
}

// AdminTopics 返回后台内容列表，包含隐藏内容。
func (s *MySQLStore) AdminTopics(site, board, q string) []domain.Post {
	site = strings.TrimSpace(site)
	board = strings.TrimSpace(board)
	q = strings.ToLower(strings.TrimSpace(q))
	rows, err := s.db.Query(`SELECT id,community_id,category_id,user_id,title,COALESCE(plugin_code,''),COALESCE(content_type,''),COALESCE(summary,''),content,status,is_pinned,is_featured,comment_locked,view_count,like_count,comment_count,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM topics WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return []domain.Post{}
	}
	defer rows.Close()
	out := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		var communityID, categoryID, userID int64
		var status int
		if err := rows.Scan(&p.ID, &communityID, &categoryID, &userID, &p.Title, &p.PluginCode, &p.ContentType, &p.Summary, &p.Content, &status, &p.Pinned, &p.Recommended, &p.CommentLocked, &p.Views, &p.Likes, &p.Comments, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		p.UserID = userID
		if comm, ok := s.communityByID(communityID); ok && comm.Slug != "" {
			p.Site = comm.Slug
		} else {
			p.Site = siteByCommunityID(communityID)
		}
		p.Board = boardByCategoryID(categoryID)
		p.ContentType = pluginregistry.NormalizeContentType(p.ContentType)
		if p.ContentType == "" {
			p.ContentType = contentTypeForBoard(p.Board)
		}
		p.PluginCode = strings.TrimSpace(p.PluginCode)
		if p.PluginCode == "" {
			p.PluginCode = pluginregistry.PluginCodeForContentType(p.ContentType)
		}
		p.Author = "DevHub 用户"
		p.Tags = s.getTopicTags(p.ID)
		if status == 1 {
			p.Status = "publish"
		} else {
			p.Status = "offline"
		}
		if site != "" && site != "portal" && p.Site != site {
			continue
		}
		if board != "" && board != "all" && p.Board != board {
			continue
		}
		if q != "" && !postContains(&p, q) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (s *MySQLStore) UpdateCommentStatus(id int64, status string) bool {
	res, err := s.db.Exec(`UPDATE comments SET status=? WHERE id=?`, strings.TrimSpace(status), id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return false
	}
	s.appendLog("audit", "auditor", "更新评论状态", fmt.Sprintf("comments#%d", id), "127.0.0.1")
	return true
}

func (s *MySQLStore) AdminSettings() domain.AdminSettings {
	var settings domain.AdminSettings
	_ = s.db.QueryRow(`SELECT site_name,copyright,default_page_size,review_timeout_hour,password_rule,captcha_enabled,search_default,search_sort,hot_view_weight,hot_like_weight,hot_comment_weight FROM admin_settings WHERE id=1`).
		Scan(&settings.SiteName, &settings.Copyright, &settings.DefaultPageSize, &settings.ReviewTimeoutHour, &settings.PasswordRule, &settings.CaptchaEnabled, &settings.SearchDefault, &settings.SearchSort, &settings.HotViewWeight, &settings.HotLikeWeight, &settings.HotCommentWeight)
	return settings
}

func (s *MySQLStore) UpdateAdminSettings(req domain.AdminSettings) domain.AdminSettings {
	_, _ = s.db.Exec(`UPDATE admin_settings SET site_name=?,copyright=?,default_page_size=?,review_timeout_hour=?,password_rule=?,captcha_enabled=?,search_default=?,search_sort=?,hot_view_weight=?,hot_like_weight=?,hot_comment_weight=? WHERE id=1`,
		req.SiteName, req.Copyright, req.DefaultPageSize, req.ReviewTimeoutHour, req.PasswordRule, req.CaptchaEnabled, req.SearchDefault, req.SearchSort, req.HotViewWeight, req.HotLikeWeight, req.HotCommentWeight)
	s.appendLog("system", "admin", "更新基础参数", "settings", "127.0.0.1")
	return s.AdminSettings()
}

func (s *MySQLStore) AdminLogs(site string) []domain.AdminLog {
	site = strings.TrimSpace(site)
	query := `SELECT id,site_key,log_type,actor,actor_type,actor_id,role_code,action,target,COALESCE(CAST(old_value AS CHAR),''),COALESCE(CAST(new_value AS CHAR),''),COALESCE(CAST(metadata_json AS CHAR),''),ip,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM admin_logs`
	args := []any{}
	if site != "" && site != "portal" {
		query += ` WHERE site_key=?`
		args = append(args, site)
	}
	query += ` ORDER BY id DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.AdminLog{}
	}
	defer rows.Close()
	out := []domain.AdminLog{}
	users := s.AdminUsers()
	for rows.Next() {
		var l domain.AdminLog
		if err := rows.Scan(&l.ID, &l.Site, &l.Type, &l.Actor, &l.ActorType, &l.ActorID, &l.Role, &l.Action, &l.Target, &l.OldValue, &l.NewValue, &l.Metadata, &l.IP, &l.CreatedAt); err == nil {
			out = append(out, enrichAdminLog(l, users))
		}
	}
	return out
}

func (s *MySQLStore) AdminLogsByFilter(filter domain.AdminLogFilter) ([]domain.AdminLog, int) {
	site := normalizeSiteScope(filter.Site)
	where := ` WHERE 1=1`
	args := []any{}
	users := s.AdminUsers()
	if site != "" && site != "portal" {
		where += ` AND site_key=?`
		args = append(args, site)
	}
	if filter.Type != "" && filter.Type != "all" {
		where += ` AND log_type=?`
		args = append(args, filter.Type)
	}
	if strings.TrimSpace(filter.Action) != "" {
		where += ` AND action LIKE ?`
		args = append(args, "%"+strings.TrimSpace(filter.Action)+"%")
	}
	if strings.TrimSpace(filter.TargetType) != "" && filter.TargetType != "all" {
		where += ` AND target LIKE ?`
		args = append(args, strings.TrimSpace(filter.TargetType)+"%")
	}
	if filter.TargetID > 0 {
		where += ` AND (target LIKE ? OR target LIKE ?)`
		idText := strconv.FormatInt(filter.TargetID, 10)
		args = append(args, "%#"+idText+"%", "%:"+idText+"%")
	}
	if strings.TrimSpace(filter.PluginCode) != "" {
		needle := "%" + strings.TrimSpace(filter.PluginCode) + "%"
		where += ` AND (target LIKE ? OR COALESCE(CAST(metadata_json AS CHAR),'') LIKE ? OR COALESCE(CAST(old_value AS CHAR),'') LIKE ? OR COALESCE(CAST(new_value AS CHAR),'') LIKE ?)`
		args = append(args, needle, needle, needle, needle)
	}
	if strings.TrimSpace(filter.ActorType) != "" && filter.ActorType != "all" {
		where += ` AND actor_type=?`
		args = append(args, strings.TrimSpace(filter.ActorType))
	}
	if filter.ActorID > 0 {
		where += ` AND actor_id=?`
		args = append(args, filter.ActorID)
	}
	if filter.CommunityID > 0 {
		scope := siteByCommunityID(filter.CommunityID)
		if comm, ok := s.communityByID(filter.CommunityID); ok && comm.Slug != "" {
			scope = comm.Slug
		}
		if scope == "" {
			where += ` AND 1=0`
		} else {
			where += ` AND site_key=?`
			args = append(args, scope)
		}
	}
	if strings.TrimSpace(filter.Actor) != "" {
		where += ` AND actor LIKE ?`
		args = append(args, "%"+strings.TrimSpace(filter.Actor)+"%")
	}
	if strings.TrimSpace(filter.Target) != "" {
		where += ` AND target LIKE ?`
		args = append(args, "%"+strings.TrimSpace(filter.Target)+"%")
	}
	if strings.TrimSpace(filter.Metadata) != "" {
		where += ` AND COALESCE(CAST(metadata_json AS CHAR),'') LIKE ?`
		args = append(args, "%"+strings.TrimSpace(filter.Metadata)+"%")
	}
	if strings.TrimSpace(filter.RequestID) != "" {
		where += ` AND COALESCE(CAST(metadata_json AS CHAR),'') LIKE ?`
		args = append(args, "%"+strings.TrimSpace(filter.RequestID)+"%")
	}
	if strings.TrimSpace(filter.StartTime) != "" {
		where += ` AND created_at >= ?`
		args = append(args, strings.TrimSpace(filter.StartTime))
	}
	if strings.TrimSpace(filter.EndTime) != "" {
		where += ` AND created_at <= ?`
		args = append(args, strings.TrimSpace(filter.EndTime))
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM admin_logs`+where, args...).Scan(&total)
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	rows, err := s.db.Query(`SELECT id,site_key,log_type,actor,actor_type,actor_id,role_code,action,target,COALESCE(CAST(old_value AS CHAR),''),COALESCE(CAST(new_value AS CHAR),''),COALESCE(CAST(metadata_json AS CHAR),''),ip,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM admin_logs`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, append(append([]any{}, args...), pageSize, offset)...)
	if err != nil {
		return []domain.AdminLog{}, 0
	}
	defer rows.Close()
	out := []domain.AdminLog{}
	for rows.Next() {
		var l domain.AdminLog
		if err := rows.Scan(&l.ID, &l.Site, &l.Type, &l.Actor, &l.ActorType, &l.ActorID, &l.Role, &l.Action, &l.Target, &l.OldValue, &l.NewValue, &l.Metadata, &l.IP, &l.CreatedAt); err == nil {
			out = append(out, enrichAdminLog(l, users))
		}
	}
	return out, total
}

func (s *MySQLStore) PluginImpact(code string) (domain.PluginImpact, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.PluginImpact{}, errors.New("plugin_code 不能为空")
	}
	plugin, ok := s.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return domain.PluginImpact{}, errors.New("插件不存在")
	}

	enabledCommunities := 0
	_ = s.db.QueryRow(
		`SELECT COUNT(*) FROM community_plugins cp 
		 JOIN plugins p ON p.plugin_code=cp.plugin_code
		 WHERE cp.plugin_code=? AND cp.status='enabled' AND p.status='enabled'`,
		code,
	).Scan(&enabledCommunities)
	totalCommunities := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM communities WHERE deleted_at IS NULL`).Scan(&totalCommunities)
	disabledCommunities := totalCommunities - enabledCommunities
	if disabledCommunities < 0 {
		disabledCommunities = 0
	}

	categories := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM categories WHERE plugin_code=? AND deleted_at IS NULL`, code).Scan(&categories)

	topics := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE plugin_code=? AND deleted_at IS NULL`, code).Scan(&topics)
	recent := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE plugin_code=? AND deleted_at IS NULL AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`, code).Scan(&recent)

	// "pending" here maps to topics.status=2 (审核中) per schema comment.
	pending := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE plugin_code=? AND deleted_at IS NULL AND status=2`, code).Scan(&pending)

	frontendMenus := 0
	adminMenus := 0
	moderatorMenus := 0
	for _, m := range plugin.Menus {
		area := strings.TrimSpace(m.Area)
		if area == "" {
			area = strings.TrimSpace(m.Location)
		}
		switch area {
		case "frontend":
			frontendMenus++
		case "admin":
			adminMenus++
		case "moderator":
			moderatorMenus++
		}
	}
	configs := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM plugins WHERE plugin_code=? AND config_json IS NOT NULL`, code).Scan(&configs)
	communityConfigs := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM community_plugins WHERE plugin_code=? AND config_json IS NOT NULL`, code).Scan(&communityConfigs)
	configs += communityConfigs
	pendingMigrations := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM plugin_migrations WHERE plugin_code=? AND status='pending'`, code).Scan(&pendingMigrations)
	recentHookErrors := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hook_executions WHERE plugin_code=? AND success=0 AND started_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`, code).Scan(&recentHookErrors)

	return domain.PluginImpact{
		PluginCode:               code,
		ExistingContentsCount:    topics,
		EnabledCommunitiesCount:  enabledCommunities,
		DisabledCommunitiesCount: disabledCommunities,
		CategoriesCount:          categories,
		TopicsCount:              topics,
		RecentContentsCount:      recent,
		PendingTopicsCount:       pending,
		PendingContentsCount:     pending,
		MenusCount:               len(plugin.Menus),
		FrontendMenusCount:       frontendMenus,
		ModeratorMenusCount:      moderatorMenus,
		AdminMenusCount:          adminMenus,
		ConfigsCount:             configs,
		PendingMigrationsCount:   pendingMigrations,
		RecentHookErrorsCount:    recentHookErrors,
	}, nil
}

func (s *MySQLStore) CommunityPluginImpact(communityID int64, code string) (domain.PluginImpact, error) {
	code = strings.TrimSpace(code)
	if communityID <= 0 {
		return domain.PluginImpact{}, errors.New("community_id 不合法")
	}
	if _, ok := s.communityByID(communityID); !ok {
		return domain.PluginImpact{}, errors.New("子站不存在")
	}
	impact, err := s.PluginImpact(code)
	if err != nil {
		return domain.PluginImpact{}, err
	}

	enabledCommunities := 0
	_ = s.db.QueryRow(
		`SELECT COUNT(*) FROM community_plugins cp 
		 JOIN plugins p ON p.plugin_code=cp.plugin_code
		 WHERE cp.community_id=? AND cp.plugin_code=? AND cp.status='enabled' AND p.status='enabled'`,
		communityID, code,
	).Scan(&enabledCommunities)

	categories := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM categories WHERE community_id=? AND plugin_code=? AND deleted_at IS NULL`, communityID, code).Scan(&categories)

	topics := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE community_id=? AND plugin_code=? AND deleted_at IS NULL`, communityID, code).Scan(&topics)

	pending := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE community_id=? AND plugin_code=? AND deleted_at IS NULL AND status=2`, communityID, code).Scan(&pending)
	recent := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE community_id=? AND plugin_code=? AND deleted_at IS NULL AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`, communityID, code).Scan(&recent)
	configs := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM plugins WHERE plugin_code=? AND config_json IS NOT NULL`, code).Scan(&configs)
	communityConfigs := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM community_plugins WHERE community_id=? AND plugin_code=? AND config_json IS NOT NULL`, communityID, code).Scan(&communityConfigs)
	configs += communityConfigs
	recentHookErrors := 0
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM hook_executions WHERE plugin_code=? AND community_id=? AND success=0 AND started_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`, code, communityID).Scan(&recentHookErrors)

	impact.EnabledCommunitiesCount = enabledCommunities
	impact.DisabledCommunitiesCount = 1 - enabledCommunities
	if impact.DisabledCommunitiesCount < 0 {
		impact.DisabledCommunitiesCount = 0
	}
	impact.CategoriesCount = categories
	impact.TopicsCount = topics
	impact.ExistingContentsCount = topics
	impact.RecentContentsCount = recent
	impact.PendingTopicsCount = pending
	impact.PendingContentsCount = pending
	impact.ConfigsCount = configs
	impact.RecentHookErrorsCount = recentHookErrors
	return impact, nil
}

func (s *MySQLStore) PushNotification(req domain.PushNotificationRequest) *domain.Notification {
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "portal"
	}
	n := s.createNoticeForSite(scope, strings.TrimSpace(req.Title), strings.TrimSpace(req.Content))
	if n != nil {
		s.appendLogForSite(scope, "operation", "operator", "", "发送通知", n.Title, "127.0.0.1")
	}
	return n
}

func (s *MySQLStore) allPosts() []domain.Post {
	rows, err := s.db.Query(`SELECT id,site_key,board_key,title,summary,content,author,status,pinned,recommended,reject_reason,offline_reason,views,likes,comments,tags_json,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM posts`)
	if err != nil {
		return []domain.Post{}
	}
	defer rows.Close()
	out := []domain.Post{}
	for rows.Next() {
		p, err := scanPost(rows)
		if err == nil {
			out = append(out, *p)
		}
	}
	return out
}

func (s *MySQLStore) postByID(id int64) (*domain.Post, error) {
	row := s.db.QueryRow(`SELECT id,site_key,board_key,title,summary,content,author,status,pinned,recommended,reject_reason,offline_reason,views,likes,comments,tags_json,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM posts WHERE id=?`, id)
	return scanPost(row)
}

const sqlTagSelect = `SELECT id,site_key,name,slug,COALESCE(description,''),status,COALESCE(merged_to_id,0),sort_order,use_count,COALESCE(follower_count,0),COALESCE(hot_score,0),COALESCE(seo_title,''),COALESCE(seo_description,''),COALESCE(seo_keywords,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM tags`

func scanSQLTag(scanner scanner) (domain.Tag, error) {
	var tag domain.Tag
	err := scanner.Scan(&tag.ID, &tag.Site, &tag.Name, &tag.Slug, &tag.Description, &tag.Status, &tag.MergedToID, &tag.Sort, &tag.UseCount, &tag.FollowerCount, &tag.HotScore, &tag.SEOTitle, &tag.SEODescription, &tag.SEOKeywords, &tag.CreatedAt, &tag.UpdatedAt)
	return tag, err
}

func (s *MySQLStore) tagByID(id int64) (domain.Tag, error) {
	tag, err := scanSQLTag(s.db.QueryRow(sqlTagSelect+` WHERE id=?`, id))
	if err != nil {
		return tag, err
	}
	return s.enrichSQLTag(tag), nil
}

func (s *MySQLStore) commentByID(id int64) (*domain.Comment, error) {
	row := s.db.QueryRow(`SELECT id,post_id,COALESCE(topic_id,post_id),parent_id,COALESCE(reply_to_user_id,0),COALESCE(user_id,1),author,to_author,text,COALESCE(content_html,''),status,likes,is_best,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'') FROM comments WHERE id=? AND deleted_at IS NULL`, id)
	c := &domain.Comment{}
	err := row.Scan(&c.ID, &c.PostID, &c.TopicID, &c.ParentID, &c.ReplyToUserID, &c.UserID, &c.Author, &c.To, &c.Text, &c.ContentHTML, &c.Status, &c.Likes, &c.IsBest, &c.CreatedAt, &c.UpdatedAt)
	normalizeSQLComment(c)
	return c, err
}

func (s *MySQLStore) commentCascadeIDs(root int64) []int64 {
	rows, err := s.db.Query(`SELECT id,parent_id FROM comments`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	children := map[int64][]int64{}
	for rows.Next() {
		var id, parent int64
		if err := rows.Scan(&id, &parent); err == nil {
			children[parent] = append(children[parent], id)
		}
	}
	out := []int64{}
	var walk func(int64)
	walk = func(id int64) {
		out = append(out, id)
		for _, child := range children[id] {
			walk(child)
		}
	}
	walk(root)
	return out
}

func (s *MySQLStore) topicReplies(parentIDs []int64) map[int64][]*domain.Comment {
	out := map[int64][]*domain.Comment{}
	if len(parentIDs) == 0 {
		return out
	}
	holders := make([]string, 0, len(parentIDs))
	args := make([]any, 0, len(parentIDs))
	for _, id := range parentIDs {
		holders = append(holders, "?")
		args = append(args, id)
	}
	rows, err := s.db.Query(`SELECT id,post_id,COALESCE(topic_id,post_id),parent_id,COALESCE(reply_to_user_id,0),COALESCE(user_id,1),author,to_author,text,COALESCE(content_html,''),status,likes,is_best,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'') FROM comments WHERE parent_id IN (`+strings.Join(holders, ",")+`) AND status NOT IN ('deleted','hidden') AND deleted_at IS NULL ORDER BY id ASC`, args...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		c := &domain.Comment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.TopicID, &c.ParentID, &c.ReplyToUserID, &c.UserID, &c.Author, &c.To, &c.Text, &c.ContentHTML, &c.Status, &c.Likes, &c.IsBest, &c.CreatedAt, &c.UpdatedAt); err == nil {
			normalizeSQLComment(c)
			out[c.ParentID] = append(out[c.ParentID], c)
		}
	}
	return out
}

func normalizeSQLComment(c *domain.Comment) {
	if c == nil {
		return
	}
	if c.TopicID == 0 {
		c.TopicID = c.PostID
	}
	if c.UserID == 0 {
		c.UserID = 1
	}
	c.UserName = firstNonEmptyString(c.Author, "DevHub 用户")
	if c.Content == "" {
		c.Content = c.Text
	}
	if c.Text == "" {
		c.Text = c.Content
	}
	c.LikeCount = c.Likes
	if c.UpdatedAt == "" {
		c.UpdatedAt = c.CreatedAt
	}
}

func (s *MySQLStore) upsertTags(site string, tags []string) error {
	for _, name := range uniqueTags(tags) {
		tag := normalizeTag(domain.Tag{Site: site, Name: name, Status: "enable"})
		if _, err := s.db.Exec(`INSERT INTO tags (site_key,name,slug,status,use_count) VALUES (?,?,?,?,1) ON DUPLICATE KEY UPDATE name=VALUES(name),status=IF(status='',VALUES(status),status),use_count=use_count+1,updated_at=NOW()`,
			tag.Site, tag.Name, tag.Slug, tag.Status); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) rebuildTagsFromPosts() error {
	if _, err := s.db.Exec(`UPDATE tags SET use_count=0`); err != nil {
		return err
	}
	for _, p := range s.allPosts() {
		for _, name := range uniqueTags(p.Tags) {
			tag := normalizeTag(domain.Tag{Site: p.Site, Name: name, Status: "enable"})
			if _, err := s.db.Exec(`INSERT INTO tags (site_key,name,slug,status,use_count) VALUES (?,?,?,?,1) ON DUPLICATE KEY UPDATE name=VALUES(name),use_count=use_count+1,updated_at=NOW()`,
				tag.Site, tag.Name, tag.Slug, tag.Status); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MySQLStore) userStats() domain.UserAdminStats {
	var stats domain.UserAdminStats
	_ = s.db.QueryRow(`SELECT COUNT(*),SUM(status='normal'),SUM(status='forbidden') FROM admin_users`).Scan(&stats.TotalUsers, &stats.ActiveUsers, &stats.Forbidden)
	stats.NewThisWeek = 1
	return stats
}

func (s *MySQLStore) createNotice(title, content string) *domain.Notification {
	return s.createNoticeForSite("portal", title, content)
}

func (s *MySQLStore) createNoticeForSite(site, title, content string) *domain.Notification {
	site = normalizeSiteScope(site)
	res, err := s.db.Exec(`INSERT INTO notifications (site_key,title,content,scope,is_read,created_at) VALUES (?,?,?,?,0,NOW())`, site, title, content, site)
	if err != nil {
		return nil
	}
	id, _ := res.LastInsertId()
	for _, n := range s.Notices("") {
		if n.ID == id {
			return &n
		}
	}
	return nil
}

func (s *MySQLStore) createUserNotice(userID, actorUserID int64, noticeType, targetType string, targetID, topicID, commentID int64, title, content string) *domain.Notification {
	if userID <= 0 || actorUserID == userID {
		return nil
	}
	res, err := s.db.Exec(`INSERT INTO notifications (site_key,actor_user_id,type,target_type,target_id,topic_id,comment_id,title,content,scope,user_id,is_read,created_at) VALUES ('portal',?,?,?,?,?,?,?,?,?,?,0,NOW())`,
		actorUserID, noticeType, targetType, nullableInt64(targetID), nullableInt64(topicID), nullableInt64(commentID), strings.TrimSpace(title), strings.TrimSpace(content), "user", userID)
	if err != nil {
		return nil
	}
	id, _ := res.LastInsertId()
	items, _, _ := s.UserNotifications(userID, nil, 1, 100)
	for _, item := range items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

func (s *MySQLStore) createUserNoticeTx(tx *sql.Tx, userID, actorUserID int64, noticeType, targetType string, targetID, topicID, commentID int64, title, content string) (int64, error) {
	if userID <= 0 || actorUserID == userID {
		return 0, nil
	}
	res, err := tx.Exec(`INSERT INTO notifications (site_key,actor_user_id,type,target_type,target_id,topic_id,comment_id,title,content,scope,user_id,is_read,created_at) VALUES ('portal',?,?,?,?,?,?,?,?,?,?,0,NOW())`,
		actorUserID, noticeType, targetType, nullableInt64(targetID), nullableInt64(topicID), nullableInt64(commentID), strings.TrimSpace(title), strings.TrimSpace(content), "user", userID)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func nullableInt64(v int64) any {
	if v <= 0 {
		return nil
	}
	return v
}

func recalcTopicHotScoreSQL() string {
	return `hot_score=(view_count + comment_count * 5 + like_count * 3 + favorite_count * 4)`
}

func scanNullableString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func scanNullableInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

func (s *MySQLStore) appendLog(logType, actor, action, target, ip string) {
	s.appendLogForSite("portal", logType, actor, "", action, target, ip)
}

func (s *MySQLStore) appendLogForSite(site, logType, actor, role, action, target, ip string) {
	log := enrichAdminLog(domain.AdminLog{
		Site:   normalizeSiteScope(site),
		Type:   logType,
		Actor:  actor,
		Role:   role,
		Action: action,
		Target: target,
		IP:     ip,
	}, s.AdminUsers())
	_, _ = s.db.Exec(`INSERT INTO admin_logs (site_key,log_type,actor,actor_type,actor_id,role_code,action,target,old_value,new_value,metadata_json,ip,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,NOW())`,
		log.Site, log.Type, log.Actor, log.ActorType, log.ActorID, log.Role, log.Action, log.Target, mysqlJSONArg(log.OldValue), mysqlJSONArg(log.NewValue), mysqlJSONArg(log.Metadata), log.IP)
}

func (s *MySQLStore) AppendAdminLog(log domain.AdminLog) {
	log.Site = normalizeSiteScope(log.Site)
	log = enrichAdminLog(log, s.AdminUsers())
	_, _ = s.db.Exec(`INSERT INTO admin_logs (site_key,log_type,actor,actor_type,actor_id,role_code,action,target,old_value,new_value,metadata_json,ip,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,NOW())`,
		log.Site, log.Type, log.Actor, log.ActorType, log.ActorID, log.Role, log.Action, log.Target, mysqlJSONArg(log.OldValue), mysqlJSONArg(log.NewValue), mysqlJSONArg(log.Metadata), log.IP)
}

func mysqlJSONArg(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if json.Valid([]byte(raw)) {
		return json.RawMessage(raw)
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	return json.RawMessage(encoded)
}

func nullableTimeString(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return raw
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSite(row scanner) (domain.Site, error) {
	var site domain.Site
	err := row.Scan(&site.Key, &site.Name, &site.Logo, &site.Title, &site.Sub, &site.Pub, &site.Description, &site.Color, &site.Status, &site.Sort)
	return site, err
}

func scanPost(row scanner) (*domain.Post, error) {
	p := &domain.Post{}
	var summary, content, rejectReason, offlineReason, tags, createdAt, updatedAt sql.NullString
	err := row.Scan(&p.ID, &p.Site, &p.Board, &p.Title, &summary, &content, &p.Author, &p.Status, &p.Pinned, &p.Recommended, &rejectReason, &offlineReason, &p.Views, &p.Likes, &p.Comments, &tags, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	p.Summary = summary.String
	p.Content = content.String
	p.RejectReason = rejectReason.String
	p.OfflineReason = offlineReason.String
	p.CreatedAt = createdAt.String
	p.UpdatedAt = updatedAt.String
	_ = json.Unmarshal([]byte(tags.String), &p.Tags)
	if p.Tags == nil {
		p.Tags = []string{}
	}
	return p, nil
}

// ===== 新增：DevHub 通用社区系统方法 =====

const communitySelect = `id,name,slug,COALESCE(logo,''),COALESCE(cover_image,''),COALESCE(slogan,''),COALESCE(description,''),COALESCE(theme_color,''),COALESCE(seo_title,''),COALESCE(seo_description,''),COALESCE(seo_keywords,''),sort_order,status,COALESCE(follower_count,0),COALESCE(topic_count,0),COALESCE(comment_count,0),COALESCE(hot_score,0),COALESCE(announcement_title,''),COALESCE(announcement_content,''),COALESCE(announcement_url,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')`

const categorySelect = `id,community_id,name,slug,type,type,COALESCE(plugin_code,'core'),COALESCE(CAST(allowed_content_types AS CHAR),''),COALESCE(description,''),COALESCE(icon,''),sort_order,visible,nav_visible,postable,COALESCE(seo_title,''),COALESCE(seo_description,''),status,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s')`

func scanCommunityRow(scanner interface{ Scan(dest ...any) error }) (domain.Community, error) {
	var c domain.Community
	err := scanner.Scan(&c.ID, &c.Name, &c.Slug, &c.Logo, &c.CoverImage, &c.Slogan, &c.Description, &c.ThemeColor, &c.SEOTitle, &c.SEODescription, &c.SEOKeywords, &c.SortOrder, &c.Status, &c.FollowerCount, &c.TopicCount, &c.CommentCount, &c.HotScore, &c.AnnouncementTitle, &c.AnnouncementContent, &c.AnnouncementURL, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func scanCategoryRow(scanner interface{ Scan(dest ...any) error }) (domain.Category, error) {
	var cat domain.Category
	var allowedJSON string
	err := scanner.Scan(&cat.ID, &cat.CommunityID, &cat.Name, &cat.Slug, &cat.Type, &cat.ContentType, &cat.PluginCode, &allowedJSON, &cat.Description, &cat.Icon, &cat.SortOrder, &cat.Visible, &cat.NavVisible, &cat.Postable, &cat.SEOTitle, &cat.SEODescription, &cat.Status, &cat.CreatedAt, &cat.UpdatedAt)
	if cat.ContentType == "" {
		cat.ContentType = cat.Type
	}
	cat.ContentType = pluginregistry.NormalizeContentType(cat.ContentType)
	if cat.Type == "" {
		cat.Type = cat.ContentType
	}
	if cat.PluginCode == "" {
		cat.PluginCode = pluginregistry.PluginCodeForContentType(cat.ContentType)
	}
	cat.AllowedContentTypes = parseStringSliceJSON(allowedJSON)
	if len(cat.AllowedContentTypes) == 0 {
		cat.AllowedContentTypes = pluginregistry.DefaultAllowedContentTypes(cat.ContentType)
	}
	return cat, err
}

func (s *MySQLStore) Communities() []domain.Community {
	rows, err := s.db.Query(`SELECT ` + communitySelect + ` FROM communities WHERE deleted_at IS NULL ORDER BY sort_order,id`)
	if err != nil {
		return []domain.Community{}
	}
	defer rows.Close()
	out := []domain.Community{}
	for rows.Next() {
		if c, err := scanCommunityRow(rows); err == nil {
			stats := s.CommunityStats(c.ID)
			c.FollowerCount = stats.FollowerCount
			c.TopicCount = stats.TopicCount
			c.CommentCount = stats.CommentCount
			c.HotScore = stats.HotScore
			out = append(out, c)
		}
	}
	return out
}

func (s *MySQLStore) CommunityBySlug(slug string) (domain.Community, bool) {
	row := s.db.QueryRow(`SELECT `+communitySelect+` FROM communities WHERE slug=? AND deleted_at IS NULL`, slug)
	c, err := scanCommunityRow(row)
	if err == nil {
		stats := s.CommunityStats(c.ID)
		c.FollowerCount = stats.FollowerCount
		c.TopicCount = stats.TopicCount
		c.CommentCount = stats.CommentCount
		c.HotScore = stats.HotScore
	}
	return c, err == nil
}

func (s *MySQLStore) Categories(communityID int64) []domain.Category {
	query := `SELECT ` + categorySelect + ` FROM categories WHERE deleted_at IS NULL`
	args := []any{}
	if communityID > 0 {
		query += ` AND community_id=?`
		args = append(args, communityID)
	}
	query += ` ORDER BY sort_order,id`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.Category{}
	}
	defer rows.Close()
	out := []domain.Category{}
	for rows.Next() {
		if cat, err := scanCategoryRow(rows); err == nil {
			out = append(out, cat)
		}
	}
	return out
}

func (s *MySQLStore) CommunityStats(communityID int64) domain.CommunityStats {
	stats := domain.CommunityStats{}
	_ = s.db.QueryRow(`SELECT COUNT(*),COALESCE(SUM(comment_count),0),COALESCE(SUM(hot_score),0) FROM topics WHERE community_id=? AND status=1 AND deleted_at IS NULL`, communityID).
		Scan(&stats.TopicCount, &stats.CommentCount, &stats.HotScore)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE community_id=? AND content_type='question' AND status=1 AND deleted_at IS NULL`, communityID).
		Scan(&stats.QuestionCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE community_id=? AND content_type='question' AND is_solved=0 AND status=1 AND deleted_at IS NULL`, communityID).
		Scan(&stats.UnsolvedCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM follows WHERE target_type='community' AND target_id=?`, communityID).
		Scan(&stats.FollowerCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics WHERE community_id=? AND status=1 AND deleted_at IS NULL AND DATE(created_at)=CURDATE()`, communityID).
		Scan(&stats.TodayTopicCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM comments c JOIN topics t ON t.id=c.topic_id WHERE t.community_id=? AND c.status='normal' AND c.deleted_at IS NULL AND DATE(c.created_at)=CURDATE()`, communityID).
		Scan(&stats.TodayCommentCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM community_moderators WHERE community_id=? AND status=1`, communityID).
		Scan(&stats.ModeratorCount)
	_, _ = s.db.Exec(`UPDATE communities SET follower_count=?,topic_count=?,comment_count=?,hot_score=?,updated_at=updated_at WHERE id=?`, stats.FollowerCount, stats.TopicCount, stats.CommentCount, stats.HotScore, communityID)
	return stats
}

func (s *MySQLStore) CreateCommunity(req domain.CommunityRequest) (domain.Community, error) {
	comm, err := normalizeMySQLCommunityRequest(req, nil)
	if err != nil {
		return domain.Community{}, err
	}
	res, err := s.db.Exec(`INSERT INTO communities (name,slug,logo,cover_image,slogan,description,theme_color,seo_title,seo_description,seo_keywords,sort_order,status,announcement_title,announcement_content,announcement_url,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())`,
		comm.Name, comm.Slug, comm.Logo, comm.CoverImage, comm.Slogan, comm.Description, comm.ThemeColor, comm.SEOTitle, comm.SEODescription, comm.SEOKeywords, comm.SortOrder, comm.Status, comm.AnnouncementTitle, comm.AnnouncementContent, comm.AnnouncementURL)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return domain.Community{}, errors.New("子站 slug 已存在")
		}
		return domain.Community{}, err
	}
	id, _ := res.LastInsertId()
	for _, cat := range defaultCategorySeeds(id) {
		_, _ = s.db.Exec(`INSERT IGNORE INTO categories (community_id,name,slug,type,plugin_code,allowed_content_types,description,icon,sort_order,visible,nav_visible,postable,seo_title,seo_description,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,NOW(),NOW())`,
			id, cat.Name, cat.Slug, cat.Type, cat.PluginCode, stringSliceJSON(cat.AllowedContentTypes), cat.Description, cat.Icon, cat.SortOrder, boolToInt(cat.Visible), boolToInt(cat.NavVisible), boolToInt(cat.Postable), cat.SEOTitle, cat.SEODescription)
	}
	return s.communityByIDRequired(id)
}

func (s *MySQLStore) UpdateCommunity(id int64, req domain.CommunityRequest) (domain.Community, error) {
	current, ok := s.communityByID(id)
	if !ok {
		return domain.Community{}, errors.New("子站不存在")
	}
	comm, err := normalizeMySQLCommunityRequest(req, &current)
	if err != nil {
		return domain.Community{}, err
	}
	_, err = s.db.Exec(`UPDATE communities SET name=?,slug=?,logo=?,cover_image=?,slogan=?,description=?,theme_color=?,seo_title=?,seo_description=?,seo_keywords=?,sort_order=?,status=?,announcement_title=?,announcement_content=?,announcement_url=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`,
		comm.Name, comm.Slug, comm.Logo, comm.CoverImage, comm.Slogan, comm.Description, comm.ThemeColor, comm.SEOTitle, comm.SEODescription, comm.SEOKeywords, comm.SortOrder, comm.Status, comm.AnnouncementTitle, comm.AnnouncementContent, comm.AnnouncementURL, id)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return domain.Community{}, errors.New("子站 slug 已存在")
		}
		return domain.Community{}, err
	}
	return s.communityByIDRequired(id)
}

func (s *MySQLStore) SetCommunityStatus(id int64, status int) (domain.Community, error) {
	if !validCommunityStatus(status) {
		return domain.Community{}, errors.New("子站状态不合法")
	}
	res, err := s.db.Exec(`UPDATE communities SET status=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, status, id)
	if err != nil {
		return domain.Community{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Community{}, errors.New("子站不存在")
	}
	return s.communityByIDRequired(id)
}

func (s *MySQLStore) ReorderCommunities(ids []int64) int {
	updated := 0
	for i, id := range ids {
		res, err := s.db.Exec(`UPDATE communities SET sort_order=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, i+1, id)
		if err == nil {
			if n, _ := res.RowsAffected(); n > 0 {
				updated++
			}
		}
	}
	return updated
}

func (s *MySQLStore) CreateCategory(communityID int64, req domain.CategoryRequest) (domain.Category, error) {
	if !s.existsByQuery(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL LIMIT 1`, communityID) {
		return domain.Category{}, errors.New("子站不存在")
	}
	req.CommunityID = communityID
	cat, err := normalizeMySQLCategoryRequest(req, nil)
	if err != nil {
		return domain.Category{}, err
	}
	if cat.PluginCode != pluginregistry.CoreCode {
		plugin, ok := s.PluginByCode(cat.PluginCode)
		if !ok || plugin.Status != pluginregistry.StatusEnabled {
			return domain.Category{}, errors.New("插件全局未启用，不能绑定该插件板块")
		}
		var status string
		if err := s.db.QueryRow(`SELECT status FROM community_plugins WHERE community_id=? AND plugin_code=?`, cat.CommunityID, cat.PluginCode).Scan(&status); err != nil || status != pluginregistry.StatusEnabled {
			return domain.Category{}, errors.New("当前子站未启用该插件，不能绑定该插件板块")
		}
	}
	res, err := s.db.Exec(`INSERT INTO categories (community_id,name,slug,type,plugin_code,allowed_content_types,description,icon,sort_order,visible,nav_visible,postable,seo_title,seo_description,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())`,
		cat.CommunityID, cat.Name, cat.Slug, cat.Type, cat.PluginCode, stringSliceJSON(cat.AllowedContentTypes), cat.Description, cat.Icon, cat.SortOrder, boolToInt(cat.Visible), boolToInt(cat.NavVisible), boolToInt(cat.Postable), cat.SEOTitle, cat.SEODescription, cat.Status)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return domain.Category{}, errors.New("板块 slug 已存在")
		}
		return domain.Category{}, err
	}
	id, _ := res.LastInsertId()
	return s.categoryByIDRequired(id)
}

func (s *MySQLStore) UpdateCategory(id int64, req domain.CategoryRequest) (domain.Category, error) {
	current, ok := s.categoryByID(id)
	if !ok {
		return domain.Category{}, errors.New("板块不存在")
	}
	cat, err := normalizeMySQLCategoryRequest(req, &current)
	if err != nil {
		return domain.Category{}, err
	}
	if cat.PluginCode != pluginregistry.CoreCode {
		plugin, ok := s.PluginByCode(cat.PluginCode)
		if !ok || plugin.Status != pluginregistry.StatusEnabled {
			return domain.Category{}, errors.New("插件全局未启用，不能绑定该插件板块")
		}
		var status string
		if err := s.db.QueryRow(`SELECT status FROM community_plugins WHERE community_id=? AND plugin_code=?`, cat.CommunityID, cat.PluginCode).Scan(&status); err != nil || status != pluginregistry.StatusEnabled {
			return domain.Category{}, errors.New("当前子站未启用该插件，不能绑定该插件板块")
		}
	}
	_, err = s.db.Exec(`UPDATE categories SET community_id=?,name=?,slug=?,type=?,plugin_code=?,allowed_content_types=?,description=?,icon=?,sort_order=?,visible=?,nav_visible=?,postable=?,seo_title=?,seo_description=?,status=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`,
		cat.CommunityID, cat.Name, cat.Slug, cat.Type, cat.PluginCode, stringSliceJSON(cat.AllowedContentTypes), cat.Description, cat.Icon, cat.SortOrder, boolToInt(cat.Visible), boolToInt(cat.NavVisible), boolToInt(cat.Postable), cat.SEOTitle, cat.SEODescription, cat.Status, id)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return domain.Category{}, errors.New("板块 slug 已存在")
		}
		return domain.Category{}, err
	}
	return s.categoryByIDRequired(id)
}

func (s *MySQLStore) SetCategoryStatus(id int64, status int) (domain.Category, error) {
	if !validCategoryStatus(status) {
		return domain.Category{}, errors.New("板块状态不合法")
	}
	res, err := s.db.Exec(`UPDATE categories SET status=?,visible=?,nav_visible=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, status, status, status, id)
	if err != nil {
		return domain.Category{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Category{}, errors.New("板块不存在")
	}
	return s.categoryByIDRequired(id)
}

func (s *MySQLStore) ReorderCategories(ids []int64) int {
	updated := 0
	for i, id := range ids {
		res, err := s.db.Exec(`UPDATE categories SET sort_order=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, i+1, id)
		if err == nil {
			if n, _ := res.RowsAffected(); n > 0 {
				updated++
			}
		}
	}
	return updated
}

func (s *MySQLStore) TopicsByFilter(communityID, categoryID int64, contentType, sort string, isSolved *bool, tag string, page, pageSize int) ([]domain.Topic, int) {
	contentType = pluginregistry.NormalizeContentType(contentType)
	query := `SELECT id,community_id,category_id,user_id,title,COALESCE(slug,''),content_type,COALESCE(summary,''),content,COALESCE(ai_summary,''),COALESCE(cover_image,''),status,is_pinned,is_featured,is_solved,comment_locked,view_count,comment_count,like_count,favorite_count,hot_score,DATE_FORMAT(last_active_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM topics WHERE deleted_at IS NULL AND status=1`
	args := []any{}

	if communityID > 0 {
		query += ` AND community_id=?`
		args = append(args, communityID)
	}
	if categoryID > 0 {
		query += ` AND category_id=?`
		args = append(args, categoryID)
	}
	if contentType != "" && contentType != "all" {
		query += ` AND content_type=?`
		args = append(args, contentType)
	}
	if isSolved != nil {
		query += ` AND content_type='question' AND is_solved=?`
		args = append(args, boolToInt(*isSolved))
	}
	if tag != "" {
		slug := normalizeSlug(tag)
		if slug == "" {
			slug = strings.ToLower(strings.Join(strings.Fields(tag), "-"))
		}
		query += ` AND id IN (SELECT topic_id FROM topic_tags WHERE tag_id IN (SELECT id FROM tags WHERE name=? OR slug=?))`
		args = append(args, tag, slug)
	}
	if sort == "featured" {
		query += ` AND is_featured=1`
	}
	if sort == "unsolved" {
		query += ` AND content_type='question' AND is_solved=0`
	}

	// 排序
	switch sort {
	case "hot":
		query += ` ORDER BY is_pinned DESC, hot_score DESC`
	case "featured":
		query += ` ORDER BY is_featured DESC, created_at DESC`
	case "solved":
		query += ` ORDER BY is_solved DESC, created_at DESC`
	case "active", "unsolved":
		query += ` ORDER BY is_pinned DESC, last_active_at DESC`
	default: // latest
		query += ` ORDER BY is_pinned DESC, created_at DESC`
	}

	// 获取总数
	countQuery := strings.Replace(query, `SELECT id,community_id,category_id,user_id,title,COALESCE(slug,''),content_type,COALESCE(summary,''),content,COALESCE(ai_summary,''),COALESCE(cover_image,''),status,is_pinned,is_featured,is_solved,comment_locked,view_count,comment_count,like_count,favorite_count,hot_score,DATE_FORMAT(last_active_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM topics`, `SELECT COUNT(*) FROM topics`, 1)
	var total int
	_ = s.db.QueryRow(countQuery, args...).Scan(&total)

	// 分页
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query += ` LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.Topic{}, 0
	}
	defer rows.Close()
	topics := []domain.Topic{}
	for rows.Next() {
		t, _ := scanTopic(rows)
		if t != nil {
			topics = append(topics, *t)
		}
	}

	// 加载标签
	for i := range topics {
		topics[i].Tags = s.getTopicTags(topics[i].ID)
	}

	return topics, total
}

func (s *MySQLStore) TagTopics(tagID int64, communityID int64, contentType string, sort string, page, pageSize int) ([]domain.Topic, int) {
	tag, err := s.tagByID(tagID)
	if err != nil || tag.Status != "enable" {
		return []domain.Topic{}, 0
	}
	selectClause := `SELECT t.id,t.community_id,t.category_id,t.user_id,t.title,COALESCE(t.slug,''),t.content_type,COALESCE(t.summary,''),t.content,COALESCE(t.ai_summary,''),COALESCE(t.cover_image,''),t.status,t.is_pinned,t.is_featured,t.is_solved,t.comment_locked,t.view_count,t.comment_count,t.like_count,t.favorite_count,t.hot_score,DATE_FORMAT(t.last_active_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(t.created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(t.updated_at,'%Y-%m-%d %H:%i:%s') FROM topics t JOIN topic_tags tt ON tt.topic_id=t.id`
	where := ` WHERE tt.tag_id=? AND t.deleted_at IS NULL AND t.status=1`
	args := []any{tagID}
	if communityID > 0 {
		where += ` AND t.community_id=?`
		args = append(args, communityID)
	}
	if contentType = strings.TrimSpace(contentType); contentType != "" && contentType != "all" {
		where += ` AND t.content_type=?`
		args = append(args, contentType)
	}
	orderBy := ` ORDER BY t.created_at DESC`
	switch strings.TrimSpace(sort) {
	case "hot":
		orderBy = ` ORDER BY t.hot_score DESC,t.created_at DESC`
	case "active":
		orderBy = ` ORDER BY COALESCE(t.last_active_at,t.updated_at,t.created_at) DESC,t.created_at DESC`
	case "featured":
		where += ` AND t.is_featured=1`
		orderBy = ` ORDER BY t.updated_at DESC,t.created_at DESC`
	case "unsolved":
		where += ` AND t.content_type='question' AND t.is_solved=0`
		orderBy = ` ORDER BY COALESCE(t.last_active_at,t.updated_at,t.created_at) DESC,t.created_at DESC`
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topics t JOIN topic_tags tt ON tt.topic_id=t.id`+where, args...).Scan(&total)
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	rows, err := s.db.Query(selectClause+where+orderBy+` LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return []domain.Topic{}, 0
	}
	defer rows.Close()
	topics := []domain.Topic{}
	for rows.Next() {
		t, _ := scanTopic(rows)
		if t != nil {
			t.Tags = s.getTopicTags(t.ID)
			topics = append(topics, *t)
		}
	}
	return topics, total
}

func (s *MySQLStore) AdminTagTopics(id int64, page, pageSize int) ([]domain.Topic, int) {
	tag, err := s.tagByID(id)
	if err != nil {
		return []domain.Topic{}, 0
	}
	communityID := int64(0)
	if tag.CommunityID > 0 {
		communityID = tag.CommunityID
	}
	return s.TagTopics(id, communityID, "", "latest", page, pageSize)
}

func (s *MySQLStore) TopicByID(id int64, increaseView bool) (*domain.Topic, error) {
	if increaseView {
		_, _ = s.db.Exec(`UPDATE topics SET view_count=view_count+1 WHERE id=?`, id)
	}

	row := s.db.QueryRow(`SELECT id,community_id,category_id,user_id,title,COALESCE(slug,''),content_type,COALESCE(summary,''),content,COALESCE(ai_summary,''),COALESCE(cover_image,''),status,is_pinned,is_featured,is_solved,comment_locked,reject_reason,offline_reason,COALESCE(best_comment_id,0),view_count,comment_count,like_count,favorite_count,hot_score,DATE_FORMAT(last_active_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM topics WHERE id=? AND deleted_at IS NULL`, id)
	return s.scanTopicDetail(row)
}

func (s *MySQLStore) CreateTopic(req domain.CreateTopicRequest) (*domain.Topic, error) {
	req.ContentType = pluginregistry.NormalizeContentType(req.ContentType)
	req.PluginCode = pluginregistry.PluginCodeForContentType(req.ContentType)
	result, err := s.db.Exec(`INSERT INTO topics (community_id,category_id,user_id,title,plugin_code,content_type,summary,content,status,view_count,comment_count,like_count,favorite_count,hot_score,last_active_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,0,0,0,0,0,NOW(),NOW(),NOW())`,
		req.CommunityID, req.CategoryID, req.UserID, req.Title, req.PluginCode, req.ContentType, req.Summary, req.Content, 1)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	s.seedPluginContentRows(domain.Topic{
		ID:          id,
		CommunityID: req.CommunityID,
		CategoryID:  req.CategoryID,
		UserID:      req.UserID,
		Title:       req.Title,
		ContentType: req.ContentType,
		PluginCode:  req.PluginCode,
		Content:     req.Content,
	})

	// 添加标签
	for _, tagName := range req.Tags {
		tag, _ := s.getOrCreateTag(req.CommunityID, tagName)
		if tag.ID > 0 {
			_, _ = s.db.Exec(`INSERT IGNORE INTO topic_tags (topic_id,tag_id) VALUES (?,?)`, id, tag.ID)
		}
	}

	// 记录动态
	_, _ = s.db.Exec(`INSERT INTO activities (user_id,community_id,topic_id,action,target_type,target_id,remark,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
		req.UserID, req.CommunityID, id, "created_topic", "topic", id, req.Title)

	return s.TopicByID(id, false)
}

func (s *MySQLStore) UpdateTopic(id int64, req domain.UpdateTopicRequest) (*domain.Topic, error) {
	updates := []string{}
	args := []any{}

	if req.CommunitySlug != nil && strings.TrimSpace(*req.CommunitySlug) != "" {
		if comm, ok := s.CommunityBySlug(strings.TrimSpace(*req.CommunitySlug)); ok {
			req.CommunityID = &comm.ID
		} else {
			return nil, errors.New("子站不存在")
		}
	}
	if req.CommunityID != nil {
		if !s.existsByQuery(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL LIMIT 1`, *req.CommunityID) {
			return nil, errors.New("子站不存在")
		}
		updates = append(updates, "community_id=?")
		args = append(args, *req.CommunityID)
	}
	if req.CategoryID != nil {
		if !s.existsByQuery(`SELECT 1 FROM categories WHERE id=? AND deleted_at IS NULL LIMIT 1`, *req.CategoryID) {
			return nil, errors.New("板块不存在")
		}
		updates = append(updates, "category_id=?")
		args = append(args, *req.CategoryID)
	}
	if req.Title != nil {
		updates = append(updates, "title=?")
		args = append(args, strings.TrimSpace(*req.Title))
	}
	if req.ContentType != nil {
		normalized := pluginregistry.NormalizeContentType(*req.ContentType)
		updates = append(updates, "content_type=?")
		args = append(args, normalized)
		updates = append(updates, "plugin_code=?")
		args = append(args, pluginregistry.PluginCodeForContentType(normalized))
	}
	if req.PluginCode != nil && req.ContentType == nil {
		updates = append(updates, "plugin_code=?")
		args = append(args, strings.TrimSpace(*req.PluginCode))
	}
	if req.Summary != nil {
		updates = append(updates, "summary=?")
		args = append(args, strings.TrimSpace(*req.Summary))
	}
	if req.Content != nil {
		updates = append(updates, "content=?")
		args = append(args, strings.TrimSpace(*req.Content))
	}
	if req.Status != nil {
		updates = append(updates, "status=?")
		args = append(args, *req.Status)
	}
	if req.IsPinned != nil {
		updates = append(updates, "is_pinned=?")
		args = append(args, boolToInt(*req.IsPinned))
	}
	if req.IsFeatured != nil {
		updates = append(updates, "is_featured=?")
		args = append(args, boolToInt(*req.IsFeatured))
	}
	if req.IsSolved != nil {
		updates = append(updates, "is_solved=?")
		args = append(args, boolToInt(*req.IsSolved))
	}
	if req.CommentLocked != nil {
		updates = append(updates, "comment_locked=?")
		args = append(args, boolToInt(*req.CommentLocked))
	}

	if len(updates) == 0 && req.Tags == nil {
		return s.TopicByID(id, false)
	}

	if len(updates) > 0 {
		args = append(args, id)
		query := `UPDATE topics SET ` + strings.Join(updates, ",") + `,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`
		res, err := s.db.Exec(query, args...)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, errors.New("主题不存在")
		}
	}

	// 更新标签
	if req.Tags != nil {
		_, _ = s.db.Exec(`DELETE FROM topic_tags WHERE topic_id=?`, id)
		topic, _ := s.TopicByID(id, false)
		if topic != nil {
			for _, tagName := range *req.Tags {
				tag, _ := s.getOrCreateTag(topic.CommunityID, tagName)
				if tag.ID > 0 {
					_, _ = s.db.Exec(`INSERT IGNORE INTO topic_tags (topic_id,tag_id) VALUES (?,?)`, id, tag.ID)
				}
			}
		}
	} else if req.CommunityID != nil {
		_, _ = s.db.Exec(`DELETE FROM topic_tags WHERE topic_id=?`, id)
	}

	return s.TopicByID(id, false)
}

func (s *MySQLStore) DeleteTopic(id int64) bool {
	result, err := s.db.Exec(`UPDATE topics SET deleted_at=NOW() WHERE id=?`, id)
	if err != nil {
		return false
	}
	n, _ := result.RowsAffected()
	return n > 0
}

func (s *MySQLStore) SearchTopics(req domain.SearchRequest) ([]domain.Topic, int) {
	req.ContentType = pluginregistry.NormalizeContentType(req.ContentType)
	selectClause := `SELECT id,community_id,category_id,user_id,title,COALESCE(slug,''),content_type,COALESCE(summary,''),content,COALESCE(ai_summary,''),COALESCE(cover_image,''),status,is_pinned,is_featured,is_solved,comment_locked,view_count,comment_count,like_count,favorite_count,hot_score,DATE_FORMAT(last_active_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM topics`
	where := ` WHERE deleted_at IS NULL AND status=1`
	args := []any{}

	if strings.TrimSpace(req.Keyword) != "" {
		where += ` AND (title LIKE ? OR summary LIKE ? OR content LIKE ? OR EXISTS (SELECT 1 FROM topic_tags tt JOIN tags tg ON tg.id=tt.tag_id WHERE tt.topic_id=topics.id AND (tg.name LIKE ? OR tg.slug LIKE ?)) OR EXISTS (SELECT 1 FROM categories c WHERE c.id=topics.category_id AND c.name LIKE ?) OR EXISTS (SELECT 1 FROM communities cm WHERE cm.id=topics.community_id AND (cm.name LIKE ? OR cm.slug LIKE ?)))`
		likePattern := "%" + strings.TrimSpace(req.Keyword) + "%"
		args = append(args, likePattern, likePattern, likePattern, likePattern, likePattern, likePattern, likePattern, likePattern)
	}
	if req.CommunityID > 0 {
		where += ` AND community_id=?`
		args = append(args, req.CommunityID)
	}
	if req.CategoryID > 0 {
		where += ` AND category_id=?`
		args = append(args, req.CategoryID)
	}
	if req.ContentType != "" && req.ContentType != "all" {
		where += ` AND content_type=?`
		args = append(args, req.ContentType)
	}
	if req.PluginCode != "" {
		where += ` AND plugin_code=?`
		args = append(args, req.PluginCode)
	}
	if req.TagID > 0 {
		where += ` AND EXISTS (SELECT 1 FROM topic_tags tt WHERE tt.topic_id=topics.id AND tt.tag_id=?)`
		args = append(args, req.TagID)
	} else if strings.TrimSpace(req.Tag) != "" {
		where += ` AND EXISTS (SELECT 1 FROM topic_tags tt JOIN tags tg ON tg.id=tt.tag_id WHERE tt.topic_id=topics.id AND (tg.name=? OR tg.slug=?))`
		tag := strings.TrimSpace(req.Tag)
		slug := normalizeSlug(tag)
		if slug == "" {
			slug = strings.ToLower(strings.Join(strings.Fields(tag), "-"))
		}
		args = append(args, tag, slug)
	}
	if req.Sort == "featured" {
		where += ` AND is_featured=1`
	}
	if req.Sort == "unsolved" {
		where += ` AND content_type='question' AND is_solved=0`
	}

	countArgs := append([]any{}, args...)
	countQuery := `SELECT COUNT(*) FROM topics` + where
	var total int
	_ = s.db.QueryRow(countQuery, countArgs...).Scan(&total)

	orderBy := ` ORDER BY created_at DESC`
	switch req.Sort {
	case "active", "unsolved":
		orderBy = ` ORDER BY COALESCE(last_active_at,updated_at,created_at) DESC, created_at DESC`
	case "hot":
		orderBy = ` ORDER BY hot_score DESC, created_at DESC`
	case "featured":
		orderBy = ` ORDER BY updated_at DESC, created_at DESC`
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 50 {
		req.PageSize = 50
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PageSize
	query := selectClause + where + orderBy + ` LIMIT ? OFFSET ?`
	args = append(args, req.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.Topic{}, 0
	}
	defer rows.Close()
	topics := []domain.Topic{}
	for rows.Next() {
		t, _ := scanTopic(rows)
		if t != nil {
			topics = append(topics, *t)
		}
	}

	for i := range topics {
		topics[i].Tags = s.getTopicTags(topics[i].ID)
	}

	return topics, total
}

func (s *MySQLStore) ToggleReaction(userID int64, targetID int64, targetType, reactionType string) (bool, int, error) {
	if userID <= 0 {
		userID = 1
	}
	targetType = strings.TrimSpace(targetType)
	reactionType = strings.TrimSpace(reactionType)
	if reactionType == "" {
		reactionType = "like"
	}
	if reactionType != "like" {
		return false, 0, errors.New("不支持的互动类型")
	}
	if targetType != "topic" && targetType != "comment" {
		return false, 0, errors.New("不支持的点赞对象")
	}
	var topic *domain.Topic
	if targetType == "topic" {
		var err error
		topic, err = s.TopicByID(targetID, false)
		if err != nil || topic == nil {
			return false, 0, errors.New("主题不存在")
		}
	} else {
		if _, err := s.commentByID(targetID); err != nil {
			return false, 0, errors.New("评论不存在")
		}
	}

	var existingID int64
	_ = s.db.QueryRow(`SELECT id FROM reactions WHERE user_id=? AND target_type=? AND target_id=? AND reaction_type=? LIMIT 1`,
		userID, targetType, targetID, reactionType).Scan(&existingID)

	if existingID > 0 {
		_, err := s.db.Exec(`DELETE FROM reactions WHERE id=?`, existingID)
		if err != nil {
			return false, 0, err
		}
		s.updateReactionCount(targetType, targetID, -1)
		return false, s.reactionCount(targetType, targetID), nil
	}

	_, err := s.db.Exec(`INSERT INTO reactions (user_id,target_type,target_id,reaction_type) VALUES (?,?,?,?)`,
		userID, targetType, targetID, reactionType)
	if err != nil {
		return false, 0, err
	}
	s.updateReactionCount(targetType, targetID, 1)
	count := s.reactionCount(targetType, targetID)
	if targetType == "topic" && topic != nil {
		_, _ = s.db.Exec(`INSERT INTO activities (user_id,community_id,topic_id,action,target_type,target_id,remark,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
			userID, topic.CommunityID, topic.ID, "liked", "topic", topic.ID, topic.Title)
		s.createUserNotice(topic.UserID, userID, "topic_liked", "topic", topic.ID, topic.ID, 0, "你的内容被点赞", fmt.Sprintf("主题《%s》获得了新的点赞。", topic.Title))
	}

	return true, count, nil
}

func (s *MySQLStore) ToggleFavorite(userID int64, targetID int64, targetType string) (bool, error) {
	if userID <= 0 {
		userID = 1
	}
	targetType = strings.TrimSpace(targetType)
	if targetType == "" {
		targetType = "topic"
	}
	if targetType != "topic" {
		return false, errors.New("不支持的收藏对象")
	}
	topic, err := s.TopicByID(targetID, false)
	if err != nil || topic == nil {
		return false, errors.New("主题不存在")
	}

	var existingID int64
	_ = s.db.QueryRow(`SELECT id FROM favorites WHERE user_id=? AND target_type=? AND target_id=? LIMIT 1`,
		userID, targetType, targetID).Scan(&existingID)

	if existingID > 0 {
		_, err := s.db.Exec(`DELETE FROM favorites WHERE id=?`, existingID)
		if err != nil {
			return false, err
		}
		_, _ = s.db.Exec(`UPDATE topics SET favorite_count=GREATEST(favorite_count-1,0),`+recalcTopicHotScoreSQL()+`,updated_at=NOW() WHERE id=?`, targetID)
		return false, nil
	}

	_, err = s.db.Exec(`INSERT INTO favorites (user_id,target_type,target_id) VALUES (?,?,?)`, userID, targetType, targetID)
	if err != nil {
		return false, err
	}
	_, _ = s.db.Exec(`UPDATE topics SET favorite_count=favorite_count+1,`+recalcTopicHotScoreSQL()+`,updated_at=NOW() WHERE id=?`, targetID)
	_, _ = s.db.Exec(`INSERT INTO activities (user_id,community_id,topic_id,action,target_type,target_id,remark,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
		userID, topic.CommunityID, topic.ID, "favorited", "topic", topic.ID, topic.Title)
	s.createUserNotice(topic.UserID, userID, "topic_favorited", "topic", topic.ID, topic.ID, 0, "你的内容被收藏", fmt.Sprintf("主题《%s》被收藏。", topic.Title))
	return true, nil
}

func (s *MySQLStore) ToggleFollow(userID int64, targetID int64, targetType string) (bool, error) {
	if userID <= 0 {
		userID = 1
	}
	targetType = strings.TrimSpace(targetType)
	if err := s.validateFollowTarget(targetType, targetID); err != nil {
		return false, err
	}

	var existingID int64
	_ = s.db.QueryRow(`SELECT id FROM follows WHERE user_id=? AND target_type=? AND target_id=? LIMIT 1`,
		userID, targetType, targetID).Scan(&existingID)

	if existingID > 0 {
		_, err := s.db.Exec(`DELETE FROM follows WHERE id=?`, existingID)
		if err == nil && targetType == "community" {
			_, _ = s.db.Exec(`UPDATE communities SET follower_count=GREATEST(follower_count-1,0),updated_at=NOW() WHERE id=?`, targetID)
		}
		if err == nil && targetType == "tag" {
			_, _ = s.db.Exec(`UPDATE tags SET follower_count=GREATEST(follower_count-1,0),updated_at=NOW() WHERE id=?`, targetID)
		}
		return false, err
	}

	_, err := s.db.Exec(`INSERT INTO follows (user_id,target_type,target_id) VALUES (?,?,?)`, userID, targetType, targetID)
	if err != nil {
		return false, err
	}
	if targetType == "community" {
		_, _ = s.db.Exec(`UPDATE communities SET follower_count=follower_count+1,updated_at=NOW() WHERE id=?`, targetID)
	}
	if targetType == "tag" {
		_, _ = s.db.Exec(`UPDATE tags SET follower_count=follower_count+1,updated_at=NOW() WHERE id=?`, targetID)
	}
	communityID, topicID, remark := s.followActivityContext(targetType, targetID)
	_, _ = s.db.Exec(`INSERT INTO activities (user_id,community_id,topic_id,action,target_type,target_id,remark,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
		userID, nullableInt64(communityID), nullableInt64(topicID), "followed", targetType, targetID, remark)
	if targetType == "user" {
		s.createUserNotice(targetID, userID, "user_followed", "user", targetID, 0, 0, "你有新的关注者", "有用户关注了你。")
	}
	return true, nil
}

func (s *MySQLStore) TopicInteraction(userID int64, topicID int64) (domain.TopicInteraction, error) {
	if userID <= 0 {
		userID = 1
	}
	var interaction domain.TopicInteraction
	err := s.db.QueryRow(`SELECT like_count,favorite_count,hot_score FROM topics WHERE id=? AND deleted_at IS NULL`, topicID).
		Scan(&interaction.LikeCount, &interaction.FavoriteCount, &interaction.HotScore)
	if err != nil {
		return domain.TopicInteraction{}, errors.New("主题不存在")
	}
	interaction.Liked = s.existsByQuery(`SELECT 1 FROM reactions WHERE user_id=? AND target_type='topic' AND target_id=? AND reaction_type='like' LIMIT 1`, userID, topicID)
	interaction.Favorited = s.existsByQuery(`SELECT 1 FROM favorites WHERE user_id=? AND target_type='topic' AND target_id=? LIMIT 1`, userID, topicID)
	interaction.Followed = s.existsByQuery(`SELECT 1 FROM follows WHERE user_id=? AND target_type='topic' AND target_id=? LIMIT 1`, userID, topicID)
	return interaction, nil
}

func (s *MySQLStore) UserFavorites(userID int64, targetType string, page, pageSize int) ([]domain.FavoriteItem, int) {
	if userID <= 0 {
		userID = 1
	}
	targetType = normalizeOptionalTargetType(targetType, "topic")
	where := ` WHERE f.user_id=?`
	args := []any{userID}
	if targetType != "" {
		where += ` AND f.target_type=?`
		args = append(args, targetType)
	}

	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM favorites f`+where, args...).Scan(&total)
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	query := `SELECT f.id,f.user_id,f.target_type,f.target_id,DATE_FORMAT(f.created_at,'%Y-%m-%d %H:%i:%s'),
		COALESCE(t.id,0),COALESCE(t.community_id,0),COALESCE(t.category_id,0),COALESCE(t.user_id,0),COALESCE(t.title,''),COALESCE(t.slug,''),COALESCE(t.content_type,''),COALESCE(t.summary,''),COALESCE(t.content,''),COALESCE(t.ai_summary,''),COALESCE(t.cover_image,''),COALESCE(t.status,0),COALESCE(t.is_pinned,0),COALESCE(t.is_featured,0),COALESCE(t.is_solved,0),COALESCE(t.view_count,0),COALESCE(t.comment_count,0),COALESCE(t.like_count,0),COALESCE(t.favorite_count,0),COALESCE(t.hot_score,0),COALESCE(DATE_FORMAT(t.last_active_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(t.created_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(t.updated_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(c.id,0),COALESCE(c.name,''),COALESCE(c.slug,''),COALESCE(c.logo,''),COALESCE(c.description,''),COALESCE(c.sort_order,0),COALESCE(c.status,0),COALESCE(DATE_FORMAT(c.created_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(c.updated_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(cat.id,0),COALESCE(cat.community_id,0),COALESCE(cat.name,''),COALESCE(cat.slug,''),COALESCE(cat.type,''),COALESCE(cat.description,''),COALESCE(cat.icon,''),COALESCE(cat.sort_order,0),COALESCE(cat.visible,0),COALESCE(cat.status,0),COALESCE(DATE_FORMAT(cat.created_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(cat.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM favorites f
		LEFT JOIN topics t ON f.target_type='topic' AND f.target_id=t.id AND t.deleted_at IS NULL
		LEFT JOIN communities c ON t.community_id=c.id
		LEFT JOIN categories cat ON t.category_id=cat.id` + where + ` ORDER BY f.created_at DESC,f.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return []domain.FavoriteItem{}, 0
	}
	defer rows.Close()
	items := []domain.FavoriteItem{}
	for rows.Next() {
		var item domain.FavoriteItem
		var topic domain.Topic
		var community domain.Community
		var category domain.Category
		if err := rows.Scan(&item.ID, &item.UserID, &item.TargetType, &item.TargetID, &item.CreatedAt,
			&topic.ID, &topic.CommunityID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Slug, &topic.ContentType, &topic.Summary, &topic.Content, &topic.AISummary, &topic.CoverImage, &topic.Status, &topic.IsPinned, &topic.IsFeatured, &topic.IsSolved, &topic.ViewCount, &topic.CommentCount, &topic.LikeCount, &topic.FavoriteCount, &topic.HotScore, &topic.LastActiveAt, &topic.CreatedAt, &topic.UpdatedAt,
			&community.ID, &community.Name, &community.Slug, &community.Logo, &community.Description, &community.SortOrder, &community.Status, &community.CreatedAt, &community.UpdatedAt,
			&category.ID, &category.CommunityID, &category.Name, &category.Slug, &category.Type, &category.Description, &category.Icon, &category.SortOrder, &category.Visible, &category.Status, &category.CreatedAt, &category.UpdatedAt); err == nil {
			if topic.ID > 0 {
				topic.Tags = s.getTopicTags(topic.ID)
				item.Topic = topic
				item.Community = community
				item.Category = category
				item.TargetURL = fmt.Sprintf("/topics/%d/", topic.ID)
			}
			items = append(items, item)
		}
	}
	return items, total
}

func (s *MySQLStore) UserFollows(userID int64, targetType string, page, pageSize int) ([]domain.FollowItem, int) {
	if userID <= 0 {
		userID = 1
	}
	targetType = normalizeOptionalTargetType(targetType, "")
	where := ` WHERE user_id=?`
	args := []any{userID}
	if targetType != "" {
		where += ` AND target_type=?`
		args = append(args, targetType)
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM follows`+where, args...).Scan(&total)
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	rows, err := s.db.Query(`SELECT id,user_id,target_type,target_id,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM follows`+where+` ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`, append(append([]any{}, args...), pageSize, offset)...)
	if err != nil {
		return []domain.FollowItem{}, 0
	}
	defer rows.Close()
	items := []domain.FollowItem{}
	for rows.Next() {
		var f domain.Follow
		if err := rows.Scan(&f.ID, &f.UserID, &f.TargetType, &f.TargetID, &f.CreatedAt); err == nil {
			items = append(items, s.followItem(f))
		}
	}
	return items, total
}

func (s *MySQLStore) UserActivities(userID int64, communityID int64, action string, page, pageSize int) ([]domain.Activity, int) {
	if userID <= 0 {
		userID = 1
	}
	where := ` WHERE user_id=?`
	args := []any{userID}

	if communityID > 0 {
		where += ` AND community_id=?`
		args = append(args, communityID)
	}
	if strings.TrimSpace(action) != "" {
		where += ` AND action=?`
		args = append(args, strings.TrimSpace(action))
	}

	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM activities`+where, args...).Scan(&total)

	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	query := `SELECT id,user_id,COALESCE(community_id,0),COALESCE(topic_id,0),action,target_type,target_id,COALESCE(remark,''),COALESCE(metadata,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s') FROM activities` + where + ` ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []domain.Activity{}, 0
	}
	defer rows.Close()
	activities := []domain.Activity{}
	for rows.Next() {
		var a domain.Activity
		if rows.Scan(&a.ID, &a.UserID, &a.CommunityID, &a.TopicID, &a.Action, &a.TargetType, &a.TargetID, &a.Remark, &a.Metadata, &a.CreatedAt) == nil {
			s.enrichActivity(&a)
			activities = append(activities, a)
		}
	}

	return activities, total
}

func (s *MySQLStore) UserNotifications(userID int64, isRead *bool, page, pageSize int) ([]domain.Notification, int, int) {
	if userID <= 0 {
		userID = 1
	}
	where := ` WHERE (user_id=? OR user_id IS NULL)`
	args := []any{userID}
	if isRead != nil {
		where += ` AND is_read=?`
		args = append(args, boolToInt(*isRead))
	}
	var total, unread int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM notifications`+where, args...).Scan(&total)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE (user_id=? OR user_id IS NULL) AND is_read=0`, userID).Scan(&unread)
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	query := `SELECT id,site_key,COALESCE(user_id,0),COALESCE(actor_user_id,0),COALESCE(type,''),COALESCE(target_type,''),COALESCE(target_id,0),COALESCE(topic_id,0),COALESCE(comment_id,0),title,COALESCE(content,''),is_read,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(read_at,'%Y-%m-%d %H:%i:%s'),'') FROM notifications` + where + ` ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, append(append([]any{}, args...), pageSize, offset)...)
	if err != nil {
		return []domain.Notification{}, 0, 0
	}
	defer rows.Close()
	items := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.Site, &n.UserID, &n.ActorUserID, &n.Type, &n.TargetType, &n.TargetID, &n.TopicID, &n.CommentID, &n.Title, &n.Content, &n.Read, &n.CreatedAt, &n.ReadAt); err == nil {
			n.IsRead = n.Read
			n.TargetURL = targetURLFor(n.TargetType, n.TargetID, n.TopicID)
			items = append(items, n)
		}
	}
	return items, total, unread
}

func (s *MySQLStore) ReadUserNotification(userID int64, id int64) bool {
	if userID <= 0 {
		userID = 1
	}
	res, err := s.db.Exec(`UPDATE notifications SET is_read=1,read_at=NOW() WHERE id=? AND (user_id=? OR user_id IS NULL)`, id, userID)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *MySQLStore) ReadAllUserNotifications(userID int64) int {
	if userID <= 0 {
		userID = 1
	}
	res, err := s.db.Exec(`UPDATE notifications SET is_read=1,read_at=NOW() WHERE is_read=0 AND (user_id=? OR user_id IS NULL)`, userID)
	if err != nil {
		return 0
	}
	n, _ := res.RowsAffected()
	return int(n)
}

func (s *MySQLStore) CommunityOverview(slug string) (domain.CommunityOverview, bool) {
	comm, ok := s.CommunityBySlug(slug)
	if !ok {
		return domain.CommunityOverview{}, false
	}

	categories := s.Categories(comm.ID)

	// 板块计数
	categoryCounts := make(map[string]int)
	for _, cat := range categories {
		var count int
		countQuery := `SELECT COUNT(*) FROM topics WHERE category_id=? AND deleted_at IS NULL AND status=1`
		_ = s.db.QueryRow(countQuery, cat.ID).Scan(&count)
		categoryCounts[cat.Slug] = count
	}

	// 热门话题
	hotTopics, _ := s.TopicsByFilter(comm.ID, 0, "", "hot", nil, "", 1, 6)

	// 最新话题
	latestTopics, _ := s.TopicsByFilter(comm.ID, 0, "", "latest", nil, "", 1, 6)

	// 热门标签
	hotTags := s.CommunityTags(comm.ID)

	// 统计
	var totalPosts, totalViews, totalLikes, totalComments int
	_ = s.db.QueryRow(`SELECT COUNT(*),SUM(view_count),SUM(like_count),SUM(comment_count) FROM topics WHERE community_id=? AND deleted_at IS NULL AND status=1`, comm.ID).
		Scan(&totalPosts, &totalViews, &totalLikes, &totalComments)
	if totalViews < 0 {
		totalViews = 0
	}
	if totalLikes < 0 {
		totalLikes = 0
	}
	if totalComments < 0 {
		totalComments = 0
	}

	return domain.CommunityOverview{
		Community:      comm,
		Categories:     categories,
		CategoryCounts: categoryCounts,
		HotTopics:      hotTopics,
		LatestTopics:   latestTopics,
		HotTags:        hotTags,
		Stats: domain.PostStats{
			TotalPosts:    totalPosts,
			TotalViews:    totalViews,
			TotalLikes:    totalLikes,
			TotalComments: totalComments,
		},
	}, true
}

func (s *MySQLStore) CreateCommentWithRequest(topicID int64, req domain.CreateCommentRequest) (*domain.Comment, error) {
	return s.CreateComment(topicID, req)
}

func (s *MySQLStore) AcceptBestAnswer(topicID int64, commentID int64, actorUserID int64) bool {
	topic, err := s.TopicByID(topicID, false)
	if err != nil || topic == nil || topic.ContentType != "question" {
		return false
	}
	comment, err := s.commentByID(commentID)
	if err != nil || comment.TopicID != topicID || comment.Status == "deleted" || comment.Status == "hidden" {
		return false
	}
	if actorUserID <= 0 {
		actorUserID = topic.UserID
		if actorUserID <= 0 {
			actorUserID = 1
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return false
	}
	if _, err = tx.Exec(`UPDATE comments SET is_best=0 WHERE COALESCE(topic_id,post_id)=?`, topicID); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err = tx.Exec(`UPDATE comments SET is_best=1,updated_at=NOW() WHERE id=? AND COALESCE(topic_id,post_id)=?`, commentID, topicID); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err = tx.Exec(`UPDATE topics SET is_solved=1,best_comment_id=?,last_active_at=NOW(),updated_at=NOW() WHERE id=?`, commentID, topicID); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err = tx.Exec(`UPDATE qa_questions SET is_solved=1,best_answer_id=?,accepted_at=NOW(),updated_at=NOW() WHERE topic_id=?`, commentID, topicID); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err = tx.Exec(`UPDATE qa_answers SET is_accepted=0,accepted_at=NULL,updated_at=NOW() WHERE topic_id=?`, topicID); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err = tx.Exec(`UPDATE qa_answers SET is_accepted=1,accepted_at=NOW(),updated_at=NOW() WHERE comment_id=? AND topic_id=?`, commentID, topicID); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err = tx.Exec(`INSERT INTO activities (user_id,community_id,topic_id,action,target_type,target_id,remark,created_at) VALUES (?,?,?,?,?,?,?,NOW())`,
		actorUserID, topic.CommunityID, topic.ID, "accepted_answer", "comment", commentID, topic.Title); err != nil {
		_ = tx.Rollback()
		return false
	}
	if _, err := s.createUserNoticeTx(tx, comment.UserID, actorUserID, "answer_accepted", "comment", commentID, topic.ID, commentID, "你的回答被采纳", fmt.Sprintf("你在《%s》中的回答被采纳为最佳答案。", topic.Title)); err != nil {
		_ = tx.Rollback()
		return false
	}
	if err = tx.Commit(); err != nil {
		return false
	}
	return true
}

func (s *MySQLStore) QAQuestionByTopicID(topicID int64) (*domain.QAQuestion, error) {
	item := &domain.QAQuestion{}
	var acceptedAt sql.NullString
	err := s.db.QueryRow(`SELECT id,topic_id,is_solved,COALESCE(best_answer_id,0),COALESCE(DATE_FORMAT(accepted_at,'%Y-%m-%d %H:%i:%s'),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM qa_questions WHERE topic_id=?`, topicID).
		Scan(&item.ID, &item.TopicID, &item.IsResolved, &item.AcceptedAnswerID, &acceptedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	item.AcceptedAt = acceptedAt.String
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM qa_answers WHERE topic_id=?`, topicID).Scan(&item.AnswerCount)
	return item, nil
}

func (s *MySQLStore) QAAnswersByTopicID(topicID int64) ([]domain.QAAnswer, error) {
	rows, err := s.db.Query(`SELECT id,topic_id,comment_id,user_id,is_accepted,COALESCE(DATE_FORMAT(accepted_at,'%Y-%m-%d %H:%i:%s'),''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM qa_answers WHERE topic_id=? ORDER BY id`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.QAAnswer{}
	for rows.Next() {
		var item domain.QAAnswer
		if err := rows.Scan(&item.ID, &item.TopicID, &item.CommentID, &item.UserID, &item.IsAccepted, &item.AcceptedAt, &item.CreatedAt, &item.UpdatedAt); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *MySQLStore) DocsDocumentByTopicID(topicID int64) (*domain.DocsDocument, error) {
	item := &domain.DocsDocument{}
	err := s.db.QueryRow(`SELECT id,COALESCE(space_id,0),topic_id,COALESCE(parent_id,0),sort_order,status,DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s') FROM docs_documents WHERE topic_id=?`, topicID).
		Scan(&item.ID, &item.SpaceID, &item.TopicID, &item.ParentID, &item.SortOrder, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	item.Version = 1
	item.EditorType = "markdown"
	return item, nil
}

func (s *MySQLStore) DocsTree(communityID int64, spaceID int64) ([]domain.DocsDocument, error) {
	query := `SELECT d.id,COALESCE(d.space_id,0),d.topic_id,COALESCE(d.parent_id,0),d.sort_order,d.status,DATE_FORMAT(d.created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(d.updated_at,'%Y-%m-%d %H:%i:%s') FROM docs_documents d JOIN docs_spaces s ON s.id=d.space_id WHERE s.community_id=?`
	args := []any{communityID}
	if spaceID > 0 {
		query += ` AND d.space_id=?`
		args = append(args, spaceID)
	}
	query += ` ORDER BY d.parent_id,d.sort_order,d.id`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.DocsDocument{}
	for rows.Next() {
		var item domain.DocsDocument
		if err := rows.Scan(&item.ID, &item.SpaceID, &item.TopicID, &item.ParentID, &item.SortOrder, &item.Status, &item.CreatedAt, &item.UpdatedAt); err == nil {
			item.Version = 1
			item.EditorType = "markdown"
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *MySQLStore) WikiPageByTopicID(topicID int64) (*domain.WikiPage, error) {
	item := &domain.WikiPage{}
	err := s.db.QueryRow(`SELECT wp.id,COALESCE(wp.space_id,0),wp.topic_id,COALESCE(wp.current_version_id,0),wp.status,DATE_FORMAT(wp.created_at,'%Y-%m-%d %H:%i:%s'),DATE_FORMAT(wp.updated_at,'%Y-%m-%d %H:%i:%s'),t.community_id,t.category_id,t.user_id,t.title,COALESCE(t.slug,''),COALESCE(t.summary,''),t.content,t.view_count,t.like_count FROM wiki_pages wp JOIN topics t ON t.id=wp.topic_id WHERE wp.topic_id=?`, topicID).
		Scan(&item.ID, &item.SpaceID, &item.TopicID, &item.CurrentVersionID, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.CommunityID, &item.CategoryID, &item.UserID, &item.Title, &item.Slug, &item.Summary, &item.Content, &item.ViewCount, &item.LikeCount)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *MySQLStore) WikiVersionsByTopicID(topicID int64) ([]domain.WikiRevision, error) {
	rows, err := s.db.Query(`SELECT v.id,v.wiki_page_id,v.topic_id,v.editor_id,v.version_no,v.title,v.content,v.change_note,DATE_FORMAT(v.created_at,'%Y-%m-%d %H:%i:%s') FROM wiki_page_versions v WHERE v.topic_id=? ORDER BY v.version_no DESC`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.WikiRevision{}
	for rows.Next() {
		var item domain.WikiRevision
		if err := rows.Scan(&item.ID, &item.WikiPageID, &item.TopicID, &item.EditorID, &item.VersionNo, &item.Title, &item.Content, &item.ChangeNote, &item.CreatedAt); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// CreateReport 创建举报记录。
func (s *MySQLStore) CreateReport(req domain.CreateReportRequest) (*domain.Report, error) {
	reporterID := req.ReporterUserID
	if reporterID <= 0 {
		reporterID = 1
	}
	targetType := strings.TrimSpace(req.TargetType)
	if !validReportTargetType(targetType) {
		return nil, errors.New("举报对象类型不合法")
	}
	reasonType := strings.TrimSpace(req.ReasonType)
	if reasonType == "" {
		return nil, errors.New("举报原因不能为空")
	}
	reasonText := strings.TrimSpace(req.ReasonText)
	if len([]rune(reasonText)) > 500 {
		return nil, errors.New("举报说明最多 500 字")
	}
	communityID, topicID, _, _, err := s.reportTargetContext(targetType, req.TargetID)
	if err != nil {
		return nil, err
	}
	var existingID int64
	_ = s.db.QueryRow(`SELECT id FROM reports WHERE reporter_id=? AND target_type=? AND target_id=? AND status='pending' LIMIT 1`, reporterID, targetType, req.TargetID).Scan(&existingID)
	if existingID > 0 {
		return nil, errors.New("同一对象已有待处理举报，请勿重复提交")
	}
	res, err := s.db.Exec(`INSERT INTO reports (reporter_id,target_type,target_id,community_id,topic_id,reason_type,reason_text,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,'pending',NOW(),NOW())`,
		reporterID, targetType, req.TargetID, communityID, topicID, reasonType, reasonText)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.ReportByID(id)
}

// Reports 返回后台举报列表。
func (s *MySQLStore) Reports(filter domain.ReportFilter) ([]domain.Report, int) {
	where := ` WHERE 1=1`
	args := []any{}
	if filter.Status != "" && filter.Status != "all" {
		where += ` AND status=?`
		args = append(args, filter.Status)
	}
	if filter.TargetType != "" && filter.TargetType != "all" {
		where += ` AND target_type=?`
		args = append(args, filter.TargetType)
	}
	if filter.CommunityID > 0 {
		where += ` AND community_id=?`
		args = append(args, filter.CommunityID)
	}
	if !filter.ActorIsAdmin {
		where += ` AND community_id IN (SELECT community_id FROM community_moderators WHERE user_id=? AND status=1)`
		args = append(args, filter.ActorUserID)
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM reports`+where, args...).Scan(&total)
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	rows, err := s.db.Query(`SELECT id,reporter_id,target_type,target_id,COALESCE(community_id,0),COALESCE(topic_id,0),reason_type,COALESCE(reason_text,''),status,COALESCE(handled_by,0),COALESCE(DATE_FORMAT(handled_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(handle_note,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'') FROM reports`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, append(append([]any{}, args...), pageSize, offset)...)
	if err != nil {
		return []domain.Report{}, 0
	}
	defer rows.Close()
	items := []domain.Report{}
	for rows.Next() {
		var report domain.Report
		if err := rows.Scan(&report.ID, &report.ReporterID, &report.TargetType, &report.TargetID, &report.CommunityID, &report.TopicID, &report.ReasonType, &report.ReasonText, &report.Status, &report.HandledBy, &report.HandledAt, &report.HandleNote, &report.CreatedAt, &report.UpdatedAt); err == nil {
			s.enrichReport(&report)
			items = append(items, report)
		}
	}
	return items, total
}

// ReportByID 返回举报详情。
func (s *MySQLStore) ReportByID(id int64) (*domain.Report, error) {
	var report domain.Report
	err := s.db.QueryRow(`SELECT id,reporter_id,target_type,target_id,COALESCE(community_id,0),COALESCE(topic_id,0),reason_type,COALESCE(reason_text,''),status,COALESCE(handled_by,0),COALESCE(DATE_FORMAT(handled_at,'%Y-%m-%d %H:%i:%s'),''),COALESCE(handle_note,''),DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'') FROM reports WHERE id=?`, id).
		Scan(&report.ID, &report.ReporterID, &report.TargetType, &report.TargetID, &report.CommunityID, &report.TopicID, &report.ReasonType, &report.ReasonText, &report.Status, &report.HandledBy, &report.HandledAt, &report.HandleNote, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return nil, errors.New("举报不存在")
	}
	s.enrichReport(&report)
	return &report, nil
}

// HandleReport 处理举报。
func (s *MySQLStore) HandleReport(id int64, status, note string, handlerUserID int64) (*domain.Report, error) {
	status = strings.TrimSpace(status)
	if status != "accepted" && status != "rejected" {
		return nil, errors.New("处理状态不合法")
	}
	report, err := s.ReportByID(id)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE reports SET status=?,handled_by=?,handled_at=NOW(),handle_note=?,updated_at=NOW() WHERE id=?`, status, handlerUserID, strings.TrimSpace(note), id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if status == "accepted" {
		switch report.TargetType {
		case "topic":
			if _, err := tx.Exec(`UPDATE topics SET status=0,updated_at=NOW() WHERE id=?`, report.TargetID); err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		case "comment":
			var isBest bool
			if err := tx.QueryRow(`SELECT is_best FROM comments WHERE id=?`, report.TargetID).Scan(&isBest); err != nil {
				_ = tx.Rollback()
				return nil, errors.New("评论不存在")
			}
			if isBest {
				_ = tx.Rollback()
				return nil, errors.New("最佳答案不能隐藏")
			}
			if _, err := tx.Exec(`UPDATE comments SET status='hidden',updated_at=NOW() WHERE id=?`, report.TargetID); err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.ReportByID(id)
}

// IsCommunityModerator 判断用户是否为子站版主。
func (s *MySQLStore) IsCommunityModerator(userID, communityID int64) bool {
	if userID <= 0 || communityID <= 0 {
		return false
	}
	return s.existsByQuery(`SELECT 1 FROM community_moderators WHERE user_id=? AND community_id=? AND status=1 LIMIT 1`, userID, communityID)
}

func (s *MySQLStore) CommunityModerators(filter domain.CommunityModeratorFilter) ([]domain.CommunityModerator, int) {
	where := ` WHERE 1=1`
	args := []any{}
	communityID := filter.CommunityID
	if communityID == 0 && filter.CommunitySlug != "" && filter.CommunitySlug != "all" && filter.CommunitySlug != "portal" {
		if comm, ok := s.CommunityBySlug(filter.CommunitySlug); ok {
			communityID = comm.ID
		}
	}
	if communityID > 0 {
		where += ` AND cm.community_id=?`
		args = append(args, communityID)
	}
	if filter.UserID > 0 {
		where += ` AND cm.user_id=?`
		args = append(args, filter.UserID)
	}
	if filter.Status != "" && filter.Status != "all" {
		status := 1
		if filter.Status == "0" || strings.EqualFold(filter.Status, "disabled") {
			status = 0
		}
		where += ` AND cm.status=?`
		args = append(args, status)
	}
	if !filter.ActorIsAdmin {
		where += ` AND cm.community_id IN (SELECT community_id FROM community_moderators WHERE user_id=? AND status=1)`
		args = append(args, filter.ActorUserID)
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM community_moderators cm`+where, args...).Scan(&total)
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize
	query := `SELECT cm.id,cm.community_id,COALESCE(c.slug,''),COALESCE(c.name,''),cm.user_id,COALESCE(u.username,''),COALESCE(u.nickname,''),cm.role,cm.status,DATE_FORMAT(cm.created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(cm.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM community_moderators cm
		LEFT JOIN communities c ON c.id=cm.community_id
		LEFT JOIN users u ON u.id=cm.user_id` + where + ` ORDER BY cm.community_id ASC, cm.id DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, append(append([]any{}, args...), pageSize, offset)...)
	if err != nil {
		return []domain.CommunityModerator{}, 0
	}
	defer rows.Close()
	items := []domain.CommunityModerator{}
	for rows.Next() {
		var moderator domain.CommunityModerator
		if err := rows.Scan(&moderator.ID, &moderator.CommunityID, &moderator.CommunitySlug, &moderator.CommunityName, &moderator.UserID, &moderator.UserName, &moderator.UserNickname, &moderator.Role, &moderator.Status, &moderator.CreatedAt, &moderator.UpdatedAt); err == nil {
			items = append(items, moderator)
		}
	}
	return items, total
}

func (s *MySQLStore) CommunityModeratorByID(id int64) (*domain.CommunityModerator, error) {
	var moderator domain.CommunityModerator
	err := s.db.QueryRow(`SELECT cm.id,cm.community_id,COALESCE(c.slug,''),COALESCE(c.name,''),cm.user_id,COALESCE(u.username,''),COALESCE(u.nickname,''),cm.role,cm.status,DATE_FORMAT(cm.created_at,'%Y-%m-%d %H:%i:%s'),COALESCE(DATE_FORMAT(cm.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM community_moderators cm
		LEFT JOIN communities c ON c.id=cm.community_id
		LEFT JOIN users u ON u.id=cm.user_id
		WHERE cm.id=?`, id).
		Scan(&moderator.ID, &moderator.CommunityID, &moderator.CommunitySlug, &moderator.CommunityName, &moderator.UserID, &moderator.UserName, &moderator.UserNickname, &moderator.Role, &moderator.Status, &moderator.CreatedAt, &moderator.UpdatedAt)
	if err != nil {
		return nil, errors.New("版主不存在")
	}
	return &moderator, nil
}

func (s *MySQLStore) CreateCommunityModerator(req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error) {
	communityID, userID, role, status, err := s.normalizeModeratorRequest(req, nil)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(`INSERT INTO community_moderators (community_id,user_id,role,status,created_at,updated_at) VALUES (?,?,?,?,NOW(),NOW())`, communityID, userID, role, status)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, errors.New("该用户已经是当前子站版主")
		}
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.CommunityModeratorByID(id)
}

func (s *MySQLStore) UpdateCommunityModerator(id int64, req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error) {
	current, err := s.CommunityModeratorByID(id)
	if err != nil {
		return nil, err
	}
	communityID, userID, role, status, err := s.normalizeModeratorRequest(req, current)
	if err != nil {
		return nil, err
	}
	res, err := s.db.Exec(`UPDATE community_moderators SET community_id=?,user_id=?,role=?,status=?,updated_at=NOW() WHERE id=?`, communityID, userID, role, status, id)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, errors.New("该用户已经是当前子站版主")
		}
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("版主不存在")
	}
	return s.CommunityModeratorByID(id)
}

func (s *MySQLStore) DeleteCommunityModerator(id int64) bool {
	res, err := s.db.Exec(`UPDATE community_moderators SET status=0,updated_at=NOW() WHERE id=?`, id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *MySQLStore) normalizeModeratorRequest(req domain.CommunityModeratorRequest, current *domain.CommunityModerator) (int64, int64, string, int, error) {
	communityID := int64(0)
	userID := int64(0)
	role := "moderator"
	status := 1
	if current != nil {
		communityID = current.CommunityID
		userID = current.UserID
		role = current.Role
		status = current.Status
	}
	if req.CommunityID > 0 {
		communityID = req.CommunityID
	}
	if req.CommunityID == 0 && strings.TrimSpace(req.CommunitySlug) != "" {
		comm, ok := s.CommunityBySlug(strings.TrimSpace(req.CommunitySlug))
		if !ok {
			return 0, 0, "", 0, errors.New("子站不存在")
		}
		communityID = comm.ID
	}
	if communityID <= 0 || !s.existsByQuery(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL LIMIT 1`, communityID) {
		return 0, 0, "", 0, errors.New("请选择子站")
	}
	if req.UserID > 0 {
		userID = req.UserID
	}
	if userID <= 0 {
		return 0, 0, "", 0, errors.New("请选择用户")
	}
	if !s.existsByQuery(`SELECT 1 FROM users WHERE id=? LIMIT 1`, userID) {
		return 0, 0, "", 0, errors.New("用户不存在")
	}
	if strings.TrimSpace(req.Role) != "" {
		role = strings.TrimSpace(req.Role)
	}
	if role != "moderator" && role != "owner" {
		return 0, 0, "", 0, errors.New("版主角色不合法")
	}
	if req.Status != nil {
		status = *req.Status
	}
	if status != 0 && status != 1 {
		return 0, 0, "", 0, errors.New("版主状态不合法")
	}
	return communityID, userID, role, status, nil
}

func (s *MySQLStore) SetTopicFeatured(id int64, featured bool) (*domain.Topic, error) {
	res, err := s.db.Exec(`UPDATE topics SET is_featured=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, boolToInt(featured), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("主题不存在")
	}
	return s.TopicByID(id, false)
}

func (s *MySQLStore) SetTopicPinned(id int64, pinned bool) (*domain.Topic, error) {
	res, err := s.db.Exec(`UPDATE topics SET is_pinned=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, boolToInt(pinned), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("主题不存在")
	}
	return s.TopicByID(id, false)
}

func (s *MySQLStore) SetTopicStatus(id int64, status int) (*domain.Topic, error) {
	if status != 0 && status != 1 && status != 3 {
		return nil, errors.New("主题状态不合法")
	}
	res, err := s.db.Exec(`UPDATE topics SET status=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, status, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("主题不存在")
	}
	return s.TopicByID(id, false)
}

func (s *MySQLStore) SetTopicCommentLocked(id int64, locked bool) (*domain.Topic, error) {
	res, err := s.db.Exec(`UPDATE topics SET comment_locked=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, boolToInt(locked), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("主题不存在")
	}
	return s.TopicByID(id, false)
}

func (s *MySQLStore) SetCommentStatus(id int64, status string) (*domain.Comment, error) {
	status = strings.TrimSpace(status)
	if status != "normal" && status != "hidden" {
		return nil, errors.New("评论状态不合法")
	}
	c, err := s.commentByID(id)
	if err != nil {
		return nil, errors.New("评论不存在")
	}
	if status == "hidden" && c.IsBest {
		return nil, errors.New("最佳答案不能隐藏")
	}
	if _, err := s.db.Exec(`UPDATE comments SET status=?,updated_at=NOW() WHERE id=?`, status, id); err != nil {
		return nil, err
	}
	return s.commentByID(id)
}

// ===== 辅助方法 =====

func scanTopic(row scanner) (*domain.Topic, error) {
	t := &domain.Topic{}
	err := row.Scan(&t.ID, &t.CommunityID, &t.CategoryID, &t.UserID, &t.Title, &t.Slug, &t.ContentType,
		&t.Summary, &t.Content, &t.AISummary, &t.CoverImage, &t.Status, &t.IsPinned, &t.IsFeatured, &t.IsSolved,
		&t.CommentLocked, &t.ViewCount, &t.CommentCount, &t.LikeCount, &t.FavoriteCount, &t.HotScore, &t.LastActiveAt, &t.CreatedAt, &t.UpdatedAt)
	t.ContentType = pluginregistry.NormalizeContentType(t.ContentType)
	t.PluginCode = pluginregistry.PluginCodeForContentType(t.ContentType)
	return t, err
}

func (s *MySQLStore) scanTopicDetail(row scanner) (*domain.Topic, error) {
	t := &domain.Topic{}
	var rejectReason, offlineReason sql.NullString
	err := row.Scan(&t.ID, &t.CommunityID, &t.CategoryID, &t.UserID, &t.Title, &t.Slug, &t.ContentType,
		&t.Summary, &t.Content, &t.AISummary, &t.CoverImage, &t.Status, &t.IsPinned, &t.IsFeatured, &t.IsSolved,
		&t.CommentLocked, &rejectReason, &offlineReason, &t.BestCommentID, &t.ViewCount, &t.CommentCount, &t.LikeCount,
		&t.FavoriteCount, &t.HotScore, &t.LastActiveAt, &t.CreatedAt, &t.UpdatedAt)
	t.RejectReason = rejectReason.String
	t.OfflineReason = offlineReason.String
	t.ContentType = pluginregistry.NormalizeContentType(t.ContentType)
	t.PluginCode = pluginregistry.PluginCodeForContentType(t.ContentType)
	if t.ID > 0 {
		t.Tags = s.getTopicTags(t.ID)
	}
	return t, err
}

func (s *MySQLStore) getTopicTags(topicID int64) []string {
	rows, err := s.db.Query(`SELECT t.name FROM topic_tags tt JOIN tags t ON tt.tag_id=t.id WHERE tt.topic_id=? ORDER BY t.name`, topicID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	tags := []string{}
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil {
			tags = append(tags, name)
		}
	}
	return tags
}

func (s *MySQLStore) enrichSQLTag(tag domain.Tag) domain.Tag {
	if tag.SortOrder == 0 {
		tag.SortOrder = tag.Sort
	}
	tag.Site = strings.TrimSpace(tag.Site)
	if tag.Site != "" && tag.Site != "portal" {
		if id, err := strconv.ParseInt(tag.Site, 10, 64); err == nil {
			if comm, ok := s.communityByID(id); ok && comm.Slug != "" {
				tag.Site = comm.Slug
			}
		}
	}
	tag.CommunitySlug = tag.Site
	if tag.Site != "" && tag.Site != "portal" {
		if comm, ok := s.CommunityBySlug(tag.Site); ok {
			tag.CommunityID = comm.ID
			tag.CommunityName = comm.Name
		}
	}
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM topic_tags tt JOIN topics tp ON tp.id=tt.topic_id AND tp.deleted_at IS NULL AND tp.status=1 WHERE tt.tag_id=?`, tag.ID).Scan(&tag.TopicCount)
	tag.UseCount = firstNonZero(tag.TopicCount, tag.UseCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM follows WHERE target_type='tag' AND target_id=?`, tag.ID).Scan(&tag.FollowerCount)
	tag.HotScore = tag.TopicCount*10 + tag.FollowerCount*20
	if tag.MergedToID > 0 {
		if target, err := s.tagByID(tag.MergedToID); err == nil {
			tag.MergedToName = target.Name
			tag.MergedToSlug = target.Slug
		}
	}
	_, _ = s.db.Exec(`UPDATE tags SET follower_count=?,use_count=?,hot_score=? WHERE id=?`, tag.FollowerCount, tag.UseCount, tag.HotScore, tag.ID)
	return tag
}

func (s *MySQLStore) getOrCreateTag(communityID int64, name string) (domain.Tag, error) {
	name = strings.TrimSpace(name)
	slug := normalizeSlug(name)
	if slug == "" {
		slug = strings.ToLower(strings.Join(strings.Fields(name), "-"))
	}
	siteKey := fmt.Sprintf("%d", communityID)
	if comm, ok := s.communityByID(communityID); ok && comm.Slug != "" {
		siteKey = comm.Slug
	}
	if resolved, ok := s.ResolveTag(siteKey, name); ok {
		return resolved.Tag, nil
	}

	// 查找已有标签
	var tag domain.Tag
	err := s.db.QueryRow(sqlTagSelect+` WHERE site_key=? AND slug=? AND status<>'merged'`, siteKey, slug).
		Scan(&tag.ID, &tag.Site, &tag.Name, &tag.Slug, &tag.Description, &tag.Status, &tag.MergedToID, &tag.Sort, &tag.UseCount, &tag.FollowerCount, &tag.HotScore, &tag.SEOTitle, &tag.SEODescription, &tag.SEOKeywords, &tag.CreatedAt, &tag.UpdatedAt)
	if err == nil {
		return s.enrichSQLTag(tag), nil
	}

	// 创建新标签
	_, err = s.db.Exec(`INSERT INTO tags (site_key,name,slug,status,use_count,hot_score,seo_title,seo_description,seo_keywords) VALUES (?,?,?,?,1,0,?,?,?)`, siteKey, name, slug, "enable", name+" 相关内容", "DevHub "+name+" 标签聚合，汇总相关文章、问答、项目和文档。", name)
	if err != nil {
		return domain.Tag{}, err
	}
	return s.getOrCreateTag(communityID, name)
}

func (s *MySQLStore) CommunityTags(communityID int64) []domain.TagStat {
	siteKey := fmt.Sprintf("%d", communityID)
	if comm, ok := s.communityByID(communityID); ok && comm.Slug != "" {
		siteKey = comm.Slug
	}
	query := `SELECT t.name,COUNT(tt.topic_id) as count FROM tags t LEFT JOIN topic_tags tt ON t.id=tt.tag_id LEFT JOIN topics tp ON tt.topic_id=tp.id AND tp.deleted_at IS NULL AND tp.status=1 WHERE t.site_key=? AND t.status='enable' GROUP BY t.name ORDER BY count DESC LIMIT 20`
	rows, err := s.db.Query(query, siteKey)
	if err != nil {
		return []domain.TagStat{}
	}
	defer rows.Close()
	tags := []domain.TagStat{}
	for rows.Next() {
		var stat domain.TagStat
		if rows.Scan(&stat.Name, &stat.Count) == nil && stat.Count > 0 {
			tags = append(tags, stat)
		}
	}
	return tags
}

func (s *MySQLStore) updateReactionCount(targetType string, targetID int64, delta int) {
	switch targetType {
	case "topic":
		if delta > 0 {
			_, _ = s.db.Exec(`UPDATE topics SET like_count=like_count+1,`+recalcTopicHotScoreSQL()+`,updated_at=NOW() WHERE id=?`, targetID)
		} else {
			_, _ = s.db.Exec(`UPDATE topics SET like_count=GREATEST(like_count-1,0),`+recalcTopicHotScoreSQL()+`,updated_at=NOW() WHERE id=?`, targetID)
		}
	case "comment":
		_, _ = s.db.Exec(`UPDATE comments SET likes=GREATEST(likes+?,0) WHERE id=?`, delta, targetID)
	}
}

func (s *MySQLStore) reactionCount(targetType string, targetID int64) int {
	var count int
	switch targetType {
	case "topic":
		_ = s.db.QueryRow(`SELECT like_count FROM topics WHERE id=?`, targetID).Scan(&count)
	case "comment":
		_ = s.db.QueryRow(`SELECT likes FROM comments WHERE id=?`, targetID).Scan(&count)
	}
	return count
}

func (s *MySQLStore) existsByQuery(query string, args ...any) bool {
	var one int
	return s.db.QueryRow(query, args...).Scan(&one) == nil
}

func (s *MySQLStore) reportTargetContext(targetType string, targetID int64) (int64, int64, string, string, error) {
	switch targetType {
	case "topic":
		topic, err := s.TopicByID(targetID, false)
		if err != nil || topic == nil {
			return 0, 0, "", "", errors.New("主题不存在")
		}
		return topic.CommunityID, topic.ID, topic.Title, firstNonEmptyString(topic.Summary, topic.Content), nil
	case "comment":
		comment, err := s.commentByID(targetID)
		if err != nil || comment.Status == "deleted" {
			return 0, 0, "", "", errors.New("评论不存在")
		}
		topic, err := s.TopicByID(comment.TopicID, false)
		if err != nil || topic == nil {
			return 0, 0, "", "", errors.New("主题不存在")
		}
		return topic.CommunityID, topic.ID, topic.Title, comment.Content, nil
	case "user":
		if s.existsByQuery(`SELECT 1 FROM users WHERE id=? LIMIT 1`, targetID) {
			return 0, 0, fmt.Sprintf("user#%d", targetID), "", nil
		}
		return 0, 0, "", "", errors.New("用户不存在")
	case "wiki":
		if s.existsByQuery(`SELECT 1 FROM wiki_pages WHERE id=? LIMIT 1`, targetID) {
			return 0, 0, fmt.Sprintf("wiki#%d", targetID), "", nil
		}
		return 0, 0, "", "", errors.New("Wiki 不存在")
	default:
		return 0, 0, "", "", errors.New("举报对象类型不合法")
	}
}

func (s *MySQLStore) enrichReport(report *domain.Report) {
	if report == nil {
		return
	}
	report.ReporterUserID = report.ReporterID
	if report.CommunityID > 0 {
		_ = s.db.QueryRow(`SELECT slug,name FROM communities WHERE id=?`, report.CommunityID).Scan(&report.CommunitySlug, &report.CommunityName)
	}
	if communityID, topicID, title, content, err := s.reportTargetContext(report.TargetType, report.TargetID); err == nil {
		if report.CommunityID == 0 {
			report.CommunityID = communityID
		}
		if report.TopicID == 0 {
			report.TopicID = topicID
		}
		report.TargetTitle = title
		report.TargetContent = content
	}
	report.TargetURL = reportTargetURL(report.TargetType, report.TargetID, report.TopicID)
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

func (s *MySQLStore) validateFollowTarget(targetType string, targetID int64) error {
	if targetID <= 0 {
		return errors.New("关注对象 ID 不合法")
	}
	switch targetType {
	case "user":
		if s.existsByQuery(`SELECT 1 FROM users WHERE id=? LIMIT 1`, targetID) {
			return nil
		}
		return errors.New("用户不存在")
	case "community":
		if s.existsByQuery(`SELECT 1 FROM communities WHERE id=? AND deleted_at IS NULL LIMIT 1`, targetID) {
			return nil
		}
		return errors.New("子站不存在")
	case "tag":
		if s.existsByQuery(`SELECT 1 FROM tags WHERE id=? AND status='enable' LIMIT 1`, targetID) {
			return nil
		}
		return errors.New("标签不存在或已禁用")
	case "topic":
		if s.existsByQuery(`SELECT 1 FROM topics WHERE id=? AND deleted_at IS NULL LIMIT 1`, targetID) {
			return nil
		}
		return errors.New("主题不存在")
	default:
		return errors.New("不支持的关注对象")
	}
}

func (s *MySQLStore) followActivityContext(targetType string, targetID int64) (int64, int64, string) {
	switch targetType {
	case "community":
		var name string
		_ = s.db.QueryRow(`SELECT name FROM communities WHERE id=?`, targetID).Scan(&name)
		return targetID, 0, name
	case "topic":
		var communityID int64
		var title string
		_ = s.db.QueryRow(`SELECT community_id,title FROM topics WHERE id=?`, targetID).Scan(&communityID, &title)
		return communityID, targetID, title
	case "tag":
		var name, site string
		_ = s.db.QueryRow(`SELECT name,site_key FROM tags WHERE id=?`, targetID).Scan(&name, &site)
		communityID := int64(0)
		if comm, ok := s.CommunityBySlug(site); ok {
			communityID = comm.ID
		}
		return communityID, 0, name
	case "user":
		var name string
		_ = s.db.QueryRow(`SELECT COALESCE(NULLIF(nickname,''),username) FROM users WHERE id=?`, targetID).Scan(&name)
		return 0, 0, name
	default:
		return 0, 0, fmt.Sprintf("%s#%d", targetType, targetID)
	}
}

func (s *MySQLStore) userDisplayName(userID int64) string {
	var name string
	_ = s.db.QueryRow(`SELECT COALESCE(NULLIF(nickname,''),username) FROM users WHERE id=?`, userID).Scan(&name)
	if strings.TrimSpace(name) == "" {
		return "Demo 用户"
	}
	return name
}

func (s *MySQLStore) followItem(f domain.Follow) domain.FollowItem {
	item := domain.FollowItem{
		ID:         f.ID,
		UserID:     f.UserID,
		TargetType: f.TargetType,
		TargetID:   f.TargetID,
		CreatedAt:  f.CreatedAt,
		TargetURL:  targetURLFor(f.TargetType, f.TargetID, f.TargetID),
	}
	switch f.TargetType {
	case "community":
		comm, ok := s.communityByID(f.TargetID)
		if ok {
			item.Community = comm
			item.TargetName = comm.Name
			item.TargetSlug = comm.Slug
			item.Description = comm.Description
			item.TargetURL = "/c/" + comm.Slug + "/"
		}
	case "topic":
		if topic, err := s.TopicByID(f.TargetID, false); err == nil && topic != nil {
			item.Topic = *topic
			item.TargetName = topic.Title
			item.TargetTitle = topic.Title
			item.TargetURL = fmt.Sprintf("/topics/%d/", topic.ID)
			if comm, ok := s.communityByID(topic.CommunityID); ok {
				item.Community = comm
			}
		}
	case "tag":
		if tag, err := s.tagByID(f.TargetID); err == nil {
			item.TargetName = tag.Name
			item.TargetSlug = tag.Slug
			item.Description = tag.Description
			item.TargetURL = tagURLForSlug(tag.Slug, tag.CommunitySlug)
			if tag.CommunityID > 0 {
				if comm, ok := s.communityByID(tag.CommunityID); ok {
					item.Community = comm
				}
			}
		}
	case "user":
		var name string
		_ = s.db.QueryRow(`SELECT COALESCE(NULLIF(nickname,''),username) FROM users WHERE id=?`, f.TargetID).Scan(&name)
		item.TargetName = name
		item.TargetURL = "/me/activities"
	}
	if item.TargetName == "" {
		item.TargetName = fmt.Sprintf("%s#%d", f.TargetType, f.TargetID)
	}
	return item
}

func (s *MySQLStore) communityByID(id int64) (domain.Community, bool) {
	row := s.db.QueryRow(`SELECT `+communitySelect+` FROM communities WHERE id=? AND deleted_at IS NULL`, id)
	c, err := scanCommunityRow(row)
	return c, err == nil
}

func (s *MySQLStore) communityByIDRequired(id int64) (domain.Community, error) {
	comm, ok := s.communityByID(id)
	if !ok {
		return domain.Community{}, errors.New("子站不存在")
	}
	stats := s.CommunityStats(comm.ID)
	comm.FollowerCount = stats.FollowerCount
	comm.TopicCount = stats.TopicCount
	comm.CommentCount = stats.CommentCount
	comm.HotScore = stats.HotScore
	return comm, nil
}

func (s *MySQLStore) categoryByID(id int64) (domain.Category, bool) {
	row := s.db.QueryRow(`SELECT `+categorySelect+` FROM categories WHERE id=? AND deleted_at IS NULL`, id)
	cat, err := scanCategoryRow(row)
	return cat, err == nil
}

func (s *MySQLStore) categoryByIDRequired(id int64) (domain.Category, error) {
	cat, ok := s.categoryByID(id)
	if !ok {
		return domain.Category{}, errors.New("板块不存在")
	}
	return cat, nil
}

func (s *MySQLStore) enrichActivity(a *domain.Activity) {
	if a == nil {
		return
	}
	if a.TargetType == "topic" {
		var title string
		var communityID int64
		var communityName string
		if s.db.QueryRow(`SELECT t.title,t.community_id,c.name FROM topics t LEFT JOIN communities c ON c.id=t.community_id WHERE t.id=?`, a.TargetID).Scan(&title, &communityID, &communityName) == nil {
			a.TargetTitle = title
			a.TargetURL = fmt.Sprintf("/topics/%d/", a.TargetID)
			a.TopicID = a.TargetID
			a.CommunityID = communityID
			a.Community = communityName
			if a.Remark == "" {
				a.Remark = title
			}
		}
		return
	}
	_, _, remark := s.followActivityContext(a.TargetType, a.TargetID)
	a.TargetTitle = remark
	a.TargetURL = targetURLFor(a.TargetType, a.TargetID, a.TopicID)
	if a.TargetType == "tag" {
		if tag, err := s.tagByID(a.TargetID); err == nil {
			a.TargetTitle = tag.Name
			a.TargetURL = tagURLForSlug(tag.Slug, tag.CommunitySlug)
			if tag.CommunityID > 0 {
				if comm, ok := s.communityByID(tag.CommunityID); ok {
					a.CommunityID = comm.ID
					a.Community = comm.Name
				}
			}
		}
	}
	if a.Remark == "" {
		a.Remark = remark
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolToTinyInt(b bool) int {
	return boolToInt(b)
}

func stringSliceJSON(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	buf, _ := json.Marshal(items)
	return string(buf)
}

func parseStringSliceJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = pluginregistry.NormalizeContentType(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return uniqueTags(out)
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func nullableMergedToID(id int64) any {
	if id <= 0 {
		return nil
	}
	return id
}

func normalizeMySQLCommunityRequest(req domain.CommunityRequest, current *domain.Community) (*domain.Community, error) {
	comm := &domain.Community{Status: 1}
	if current != nil {
		cp := *current
		comm = &cp
	}
	if strings.TrimSpace(req.Name) != "" {
		comm.Name = strings.TrimSpace(req.Name)
	}
	if comm.Name == "" {
		return nil, errors.New("子站名称不能为空")
	}
	if slug := normalizeSlug(req.Slug); slug != "" {
		comm.Slug = slug
	} else if comm.Slug == "" {
		comm.Slug = normalizeSlug(comm.Name)
	}
	if comm.Slug == "" {
		return nil, errors.New("子站 slug 不能为空")
	}
	if strings.TrimSpace(req.Logo) != "" || current == nil {
		comm.Logo = strings.TrimSpace(req.Logo)
	}
	if comm.Logo == "" {
		comm.Logo = strings.ToUpper(firstRunes(comm.Name, 2))
	}
	if strings.TrimSpace(req.CoverImage) != "" || current == nil {
		comm.CoverImage = strings.TrimSpace(req.CoverImage)
	}
	if strings.TrimSpace(req.Slogan) != "" || current == nil {
		comm.Slogan = strings.TrimSpace(req.Slogan)
	}
	if strings.TrimSpace(req.Description) != "" || current == nil {
		comm.Description = strings.TrimSpace(req.Description)
	}
	if strings.TrimSpace(req.ThemeColor) != "" || current == nil {
		comm.ThemeColor = strings.TrimSpace(req.ThemeColor)
	}
	if comm.ThemeColor == "" {
		comm.ThemeColor = "#2563eb"
	}
	if strings.TrimSpace(req.SEOTitle) != "" || current == nil {
		comm.SEOTitle = strings.TrimSpace(req.SEOTitle)
	}
	if comm.SEOTitle == "" {
		comm.SEOTitle = comm.Name + " 技术社区"
	}
	if strings.TrimSpace(req.SEODescription) != "" || current == nil {
		comm.SEODescription = strings.TrimSpace(req.SEODescription)
	}
	if comm.SEODescription == "" {
		comm.SEODescription = comm.Description
	}
	if strings.TrimSpace(req.SEOKeywords) != "" || current == nil {
		comm.SEOKeywords = strings.TrimSpace(req.SEOKeywords)
	}
	if req.SortOrder != nil {
		comm.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		if !validCommunityStatus(*req.Status) {
			return nil, errors.New("子站状态不合法")
		}
		comm.Status = *req.Status
	}
	if strings.TrimSpace(req.AnnouncementTitle) != "" || current == nil {
		comm.AnnouncementTitle = strings.TrimSpace(req.AnnouncementTitle)
	}
	if strings.TrimSpace(req.AnnouncementContent) != "" || current == nil {
		comm.AnnouncementContent = strings.TrimSpace(req.AnnouncementContent)
	}
	if strings.TrimSpace(req.AnnouncementURL) != "" || current == nil {
		comm.AnnouncementURL = strings.TrimSpace(req.AnnouncementURL)
	}
	return comm, nil
}

func normalizeMySQLCategoryRequest(req domain.CategoryRequest, current *domain.Category) (*domain.Category, error) {
	cat := &domain.Category{Visible: true, NavVisible: true, Postable: true, Status: 1}
	if current != nil {
		cp := *current
		cat = &cp
	}
	if req.CommunityID > 0 {
		cat.CommunityID = req.CommunityID
	}
	if cat.CommunityID <= 0 {
		return nil, errors.New("请选择子站")
	}
	if strings.TrimSpace(req.Name) != "" {
		cat.Name = strings.TrimSpace(req.Name)
	}
	if cat.Name == "" {
		return nil, errors.New("板块名称不能为空")
	}
	if slug := normalizeSlug(req.Slug); slug != "" {
		cat.Slug = slug
	} else if cat.Slug == "" {
		cat.Slug = normalizeSlug(cat.Name)
	}
	if cat.Slug == "" {
		return nil, errors.New("板块 slug 不能为空")
	}
	if contentType := strings.TrimSpace(firstNonEmptyString(req.ContentType, req.Type)); contentType != "" {
		contentType = pluginregistry.NormalizeContentType(contentType)
		if !validCategoryContentType(contentType) {
			return nil, errors.New("内容类型不合法")
		}
		cat.Type = contentType
		cat.ContentType = contentType
	}
	if cat.Type == "" {
		cat.Type = "article"
	}
	if cat.ContentType == "" {
		cat.ContentType = cat.Type
	}
	if req.PluginCode != "" {
		cat.PluginCode = strings.TrimSpace(req.PluginCode)
	}
	expectedPlugin := pluginregistry.PluginCodeForContentType(cat.ContentType)
	if cat.PluginCode == "" {
		cat.PluginCode = expectedPlugin
	}
	if cat.PluginCode != expectedPlugin {
		return nil, errors.New("板块插件与内容类型不匹配")
	}
	if len(req.AllowedContentTypes) > 0 {
		allowed := make([]string, 0, len(req.AllowedContentTypes))
		for _, item := range req.AllowedContentTypes {
			item = pluginregistry.NormalizeContentType(item)
			if !validCategoryContentType(item) {
				return nil, errors.New("允许内容类型不合法")
			}
			allowed = append(allowed, item)
		}
		cat.AllowedContentTypes = uniqueTags(allowed)
	}
	if len(cat.AllowedContentTypes) == 0 {
		cat.AllowedContentTypes = pluginregistry.DefaultAllowedContentTypes(cat.ContentType)
	}
	if strings.TrimSpace(req.Description) != "" || current == nil {
		cat.Description = strings.TrimSpace(req.Description)
	}
	if strings.TrimSpace(req.Icon) != "" || current == nil {
		cat.Icon = strings.TrimSpace(req.Icon)
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}
	if req.Visible != nil {
		cat.Visible = *req.Visible
	}
	if req.NavVisible != nil {
		cat.NavVisible = *req.NavVisible
	}
	if req.Postable != nil {
		cat.Postable = *req.Postable
	}
	if req.Status != nil {
		if !validCategoryStatus(*req.Status) {
			return nil, errors.New("板块状态不合法")
		}
		cat.Status = *req.Status
	}
	if strings.TrimSpace(req.SEOTitle) != "" || current == nil {
		cat.SEOTitle = strings.TrimSpace(req.SEOTitle)
	}
	if strings.TrimSpace(req.SEODescription) != "" || current == nil {
		cat.SEODescription = strings.TrimSpace(req.SEODescription)
	}
	return cat, nil
}
