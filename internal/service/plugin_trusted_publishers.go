package service

import (
	"encoding/base64"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type TrustedPublisherOperator struct {
	ID   int64
	Name string
}

func (s *Service) trustedPublishersConfig() pluginregistry.TrustedPublishersConfig {
	items, _, _ := s.repo.PluginTrustedPublishers(domain.PluginTrustedPublisherFilter{Page: 1, PageSize: 1000})
	if len(items) > 0 {
		return pluginregistry.TrustedPublishersConfigFromDomain(items)
	}
	cfg, found, err := pluginregistry.LoadTrustedPublishers()
	if err == nil && found {
		return cfg
	}
	return pluginregistry.TrustedPublishersConfig{}
}

func (s *Service) ListPluginTrustedPublishers(filter domain.PluginTrustedPublisherFilter) (domain.PluginTrustedPublisherListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	items, total, err := s.repo.PluginTrustedPublishers(filter)
	if err != nil {
		return domain.PluginTrustedPublisherListResponse{}, err
	}
	summary := map[string]int{"total": total, "trusted": 0, "blocked": 0, "revoked": 0, "unknown": 0}
	all, _, _ := s.repo.PluginTrustedPublishers(domain.PluginTrustedPublisherFilter{Page: 1, PageSize: 1000})
	for _, it := range all {
		switch strings.TrimSpace(it.Status) {
		case "trusted", "blocked", "revoked":
			summary[strings.TrimSpace(it.Status)]++
		default:
			summary["unknown"]++
		}
	}
	return domain.PluginTrustedPublisherListResponse{Items: items, Pagination: domain.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total}, Summary: summary}, nil
}

func (s *Service) GetPluginTrustedPublisher(id int64) (domain.PluginTrustedPublisher, error) {
	it, ok := s.repo.PluginTrustedPublisherByID(id)
	if !ok {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_not_found", "可信发布者不存在").WithStatus(404).WithDetail("id", id)
	}
	return it, nil
}

func (s *Service) CreatePluginTrustedPublisher(operator TrustedPublisherOperator, req domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error) {
	record, err := normalizeTrustedPublisher(req)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	if _, ok := s.repo.PluginTrustedPublisherByKey(record.PublisherID, record.PublicKeyID); ok {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_duplicate", "可信发布者 public_key_id 已存在").
			WithStatus(409).WithDetail("publisher_id", record.PublisherID).WithDetail("public_key_id", record.PublicKeyID).
			WithSuggestion("请修改 publisher_id/public_key_id，或编辑已有记录。")
	}
	record.CreatedBy = operator.ID
	record.UpdatedBy = operator.ID
	record.CreatedAt = Now()
	record.UpdatedAt = Now()
	out, err := s.repo.AppendPluginTrustedPublisher(record)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	s.logTrustedPublisher(operator, "plugin.trusted_publisher.created", out, "")
	return out, nil
}

func (s *Service) UpdatePluginTrustedPublisher(operator TrustedPublisherOperator, id int64, req domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error) {
	existing, err := s.GetPluginTrustedPublisher(id)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	record, err := normalizeTrustedPublisher(req)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	if other, ok := s.repo.PluginTrustedPublisherByKey(record.PublisherID, record.PublicKeyID); ok && other.ID != id {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_duplicate", "可信发布者 public_key_id 已存在").
			WithStatus(409).WithDetail("publisher_id", record.PublisherID).WithDetail("public_key_id", record.PublicKeyID)
	}
	record.ID = id
	record.CreatedBy = existing.CreatedBy
	record.CreatedAt = existing.CreatedAt
	record.UpdatedBy = operator.ID
	record.UpdatedAt = Now()
	if record.Status == "" {
		record.Status = existing.Status
	}
	out, err := s.repo.SavePluginTrustedPublisher(record)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	s.logTrustedPublisher(operator, "plugin.trusted_publisher.updated", out, "")
	return out, nil
}

func (s *Service) SetPluginTrustedPublisherStatus(operator TrustedPublisherOperator, id int64, status, comment string) (domain.PluginTrustedPublisher, error) {
	record, err := s.GetPluginTrustedPublisher(id)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "trusted", "blocked", "revoked":
	default:
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_invalid_status", "可信发布者状态不合法").
			WithStatus(400).WithDetail("status", status).WithSuggestion("状态仅支持 trusted / blocked / revoked。")
	}
	record.Status = status
	record.UpdatedBy = operator.ID
	record.UpdatedAt = Now()
	if status == "blocked" {
		record.BlockedAt = Now()
	}
	if status == "revoked" {
		record.RevokedAt = Now()
	}
	if status == "trusted" {
		record.BlockedAt = ""
		record.RevokedAt = ""
	}
	out, err := s.repo.SavePluginTrustedPublisher(record)
	if err != nil {
		return domain.PluginTrustedPublisher{}, err
	}
	s.logTrustedPublisher(operator, "plugin.trusted_publisher."+status, out, comment)
	return out, nil
}

func (s *Service) DeletePluginTrustedPublisher(operator TrustedPublisherOperator, id int64) error {
	record, err := s.GetPluginTrustedPublisher(id)
	if err != nil {
		return err
	}
	if err := s.repo.DeletePluginTrustedPublisher(id); err != nil {
		return err
	}
	s.logTrustedPublisher(operator, "plugin.trusted_publisher.deleted", record, "")
	return nil
}

func normalizeTrustedPublisher(req domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error) {
	req.PublisherID = strings.TrimSpace(req.PublisherID)
	req.Name = strings.TrimSpace(req.Name)
	req.PublicKeyID = strings.TrimSpace(req.PublicKeyID)
	req.PublicKeyAlgorithm = strings.ToLower(strings.TrimSpace(req.PublicKeyAlgorithm))
	req.PublicKey = strings.TrimSpace(req.PublicKey)
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status == "" {
		req.Status = "trusted"
	}
	if req.PublisherID == "" || req.Name == "" || req.PublicKeyID == "" {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_invalid_key", "publisher_id/name/public_key_id 必填").WithStatus(400)
	}
	if req.PublicKeyAlgorithm != "ed25519" {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_invalid_key", "public_key_algorithm 仅支持 ed25519").WithStatus(400).WithDetail("public_key_algorithm", req.PublicKeyAlgorithm)
	}
	raw, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil || len(raw) != 32 {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_invalid_key", "public_key 必须是 base64 Ed25519 公钥").WithStatus(400).WithSuggestion("请填写 32 字节 Ed25519 public key 的 base64 编码。")
	}
	if req.Status != "trusted" && req.Status != "blocked" && req.Status != "revoked" {
		return domain.PluginTrustedPublisher{}, domain.NewPluginError("plugin_trusted_publisher_invalid_status", "可信发布者状态不合法").WithStatus(400).WithDetail("status", req.Status)
	}
	req.Fingerprint = pluginregistry.FingerprintTrustedPublisherPublicKey(req.PublicKey)
	return req, nil
}

func (s *Service) logTrustedPublisher(operator TrustedPublisherOperator, action string, record domain.PluginTrustedPublisher, comment string) {
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    action,
		Target:    record.PublisherID,
		Metadata:  mustJSON(map[string]any{"publisher_id": record.PublisherID, "public_key_id": record.PublicKeyID, "fingerprint": record.Fingerprint, "status": record.Status, "comment": comment}),
		CreatedAt: Now(),
	})
}
