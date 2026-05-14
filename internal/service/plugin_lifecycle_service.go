package service

import (
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

// Plugins 返回系统插件注册信息与运行状态。
func (s *Service) Plugins() []domain.Plugin { return s.repo.Plugins() }

// PluginByCode 按插件唯一标识获取插件。
func (s *Service) PluginByCode(code string) (domain.Plugin, bool) {
	return s.repo.PluginByCode(code)
}

// CommunityPlugins returns plugin list with community runtime state overlay.
func (s *Service) CommunityPlugins(communityID int64) ([]domain.Plugin, error) {
	return s.repo.CommunityPlugins(communityID)
}

// SetCommunityPluginStatus updates per-community plugin enablement.
func (s *Service) SetCommunityPluginStatus(communityID int64, code, status string) (domain.Plugin, error) {
	if strings.TrimSpace(status) == pluginregistry.StatusEnabled {
		if err := s.validatePluginEnableReadiness(code); err != nil {
			return domain.Plugin{}, err
		}
	}
	return s.repo.SetCommunityPluginStatus(communityID, code, status)
}

// SetCommunityPluginConfig updates per-community plugin config blob.
func (s *Service) SetCommunityPluginConfig(communityID int64, code, configJSON string) (domain.Plugin, error) {
	code = strings.TrimSpace(code)
	items, _ := s.repo.CommunityPlugins(communityID)
	var current domain.Plugin
	for _, it := range items {
		if it.Code == code {
			current = it
			break
		}
	}
	if current.Code == "" {
		// fallback to global definition (schema)
		current, _ = s.repo.PluginByCode(code)
	}
	if current.Code == "" {
		return domain.Plugin{}, pluginNotFound(code)
	}

	res, err := s.encryptPluginConfigJSON(current, current.ConfigJSON, configJSON)
	if err != nil {
		return domain.Plugin{}, err
	}
	return s.repo.SetCommunityPluginConfig(communityID, code, res.EncryptedJSON)
}

// ReorderCommunityPlugins updates per-community plugin sort order.
func (s *Service) ReorderCommunityPlugins(communityID int64, codes []string) (int, error) {
	return s.repo.ReorderCommunityPlugins(communityID, codes)
}

// IsPluginEnabled checks whether a plugin is globally enabled.
// Core is always enabled and not persisted in plugins table.
func (s *Service) IsPluginEnabled(pluginCode string) bool {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return true
	}
	plugin, ok := s.repo.PluginByCode(pluginCode)
	return ok && plugin.Status == pluginregistry.StatusEnabled
}

// IsPluginEnabledForCommunity checks whether a plugin is enabled for a community,
// requiring both global and community status enabled.
func (s *Service) IsPluginEnabledForCommunity(communityID int64, pluginCode string) bool {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return true
	}
	items, err := s.repo.CommunityPlugins(communityID)
	if err != nil {
		return false
	}
	for _, item := range items {
		if item.Code == pluginCode {
			return item.Status == pluginregistry.StatusEnabled
		}
	}
	return false
}

// ListEnabledPluginsForCommunity returns plugins enabled for a community.
func (s *Service) ListEnabledPluginsForCommunity(communityID int64) ([]domain.Plugin, error) {
	items, err := s.repo.CommunityPlugins(communityID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Plugin, 0, len(items))
	for _, item := range items {
		if item.Status == pluginregistry.StatusEnabled {
			out = append(out, item)
		}
	}
	return out, nil
}
