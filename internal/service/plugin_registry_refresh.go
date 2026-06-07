package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
)

type pluginRegistryRefreshEvent struct {
	Trigger     string
	PluginCode  string
	CommunityID int64
	ActorType   string
	ActorID     int64
	ActorName   string
	OldVersion  string
	NewVersion  string
	Status      string
}

// RefreshPluginRegistry validates the current runtime plugin view after a
// lifecycle/configuration write. DevHub currently derives runtime state from
// the store on read, so refresh is a focused consistency barrier plus audit
// trail rather than a dynamic code loader.
func (s *Service) RefreshPluginRegistry(trigger, pluginCode string) error {
	return s.refreshPluginRegistry(pluginRegistryRefreshEvent{
		Trigger:    trigger,
		PluginCode: pluginCode,
		ActorType:  "system",
		ActorName:  "system",
	})
}

func (s *Service) refreshPluginRegistry(event pluginRegistryRefreshEvent) error {
	if s == nil || s.repo == nil {
		return nil
	}
	event.Trigger = strings.TrimSpace(event.Trigger)
	if event.Trigger == "" {
		event.Trigger = "manual"
	}
	event.PluginCode = strings.TrimSpace(event.PluginCode)
	event.ActorType = firstNonEmpty(strings.TrimSpace(event.ActorType), "system")
	event.ActorName = firstNonEmpty(strings.TrimSpace(event.ActorName), event.ActorType)

	started := time.Now()
	s.pluginRuntimeMu.Lock()
	defer s.pluginRuntimeMu.Unlock()

	if s.pluginRuntimeErr != nil {
		if err := s.pluginRuntimeErr(event); err != nil {
			s.auditPluginRegistryRefresh(event, "failed", started, err)
			return err
		}
	}

	plugins := s.repo.Plugins()
	if len(plugins) == 0 {
		err := errors.New("插件运行态为空")
		s.auditPluginRegistryRefresh(event, "failed", started, err)
		return err
	}
	for _, item := range plugins {
		if strings.TrimSpace(item.Code) == "" {
			err := errors.New("插件运行态包含空 plugin_code")
			s.auditPluginRegistryRefresh(event, "failed", started, err)
			return err
		}
	}
	if event.PluginCode != "" {
		plugin, ok := s.repo.PluginByCode(event.PluginCode)
		if !ok || strings.TrimSpace(plugin.Code) == "" {
			err := fmt.Errorf("插件运行态未找到：%s", event.PluginCode)
			s.auditPluginRegistryRefresh(event, "failed", started, err)
			return err
		}
		event.Status = firstNonEmpty(event.Status, plugin.Status)
		event.NewVersion = firstNonEmpty(event.NewVersion, plugin.Version)
	}
	if event.CommunityID > 0 {
		if _, err := s.repo.CommunityPlugins(event.CommunityID); err != nil {
			err = fmt.Errorf("子站插件运行态不可用：%w", err)
			s.auditPluginRegistryRefresh(event, "failed", started, err)
			return err
		}
	}

	s.storePluginRuntimeSnapshot(plugins, event.CommunityID)
	if event.Trigger != "startup" {
		s.auditPluginRegistryRefresh(event, "success", started, nil)
	}
	return nil
}

func (s *Service) storePluginRuntimeSnapshot(plugins []domain.Plugin, communityID int64) {
	pluginsCopy := append([]domain.Plugin(nil), plugins...)
	byCode := make(map[string]domain.Plugin, len(pluginsCopy))
	for _, plugin := range pluginsCopy {
		byCode[strings.TrimSpace(plugin.Code)] = plugin
	}

	var communityItems []domain.Plugin
	if communityID > 0 {
		if items, err := s.repo.CommunityPlugins(communityID); err == nil {
			communityItems = append([]domain.Plugin(nil), items...)
		}
	}

	s.pluginRuntimeSnapshotMu.Lock()
	defer s.pluginRuntimeSnapshotMu.Unlock()
	s.pluginRuntimePlugins = pluginsCopy
	s.pluginRuntimePluginsByCode = byCode
	if communityID <= 0 {
		s.pluginRuntimeCommunities = map[int64][]domain.Plugin{}
	}
	if communityID > 0 && communityItems != nil {
		if s.pluginRuntimeCommunities == nil {
			s.pluginRuntimeCommunities = map[int64][]domain.Plugin{}
		}
		s.pluginRuntimeCommunities[communityID] = communityItems
	}
}

func (s *Service) runtimePluginsSnapshot() ([]domain.Plugin, bool) {
	if s == nil {
		return nil, false
	}
	s.pluginRuntimeSnapshotMu.RLock()
	defer s.pluginRuntimeSnapshotMu.RUnlock()
	if len(s.pluginRuntimePlugins) == 0 {
		return nil, false
	}
	return append([]domain.Plugin(nil), s.pluginRuntimePlugins...), true
}

func (s *Service) runtimePluginByCodeSnapshot(code string) (domain.Plugin, bool) {
	if s == nil {
		return domain.Plugin{}, false
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.Plugin{}, false
	}
	s.pluginRuntimeSnapshotMu.RLock()
	defer s.pluginRuntimeSnapshotMu.RUnlock()
	if len(s.pluginRuntimePluginsByCode) == 0 {
		return domain.Plugin{}, false
	}
	plugin, ok := s.pluginRuntimePluginsByCode[code]
	return plugin, ok
}

func (s *Service) runtimeCommunityPluginsSnapshot(communityID int64) ([]domain.Plugin, bool) {
	if s == nil || communityID <= 0 {
		return nil, false
	}
	s.pluginRuntimeSnapshotMu.RLock()
	defer s.pluginRuntimeSnapshotMu.RUnlock()
	items, ok := s.pluginRuntimeCommunities[communityID]
	if !ok {
		return nil, false
	}
	return append([]domain.Plugin(nil), items...), true
}

func (s *Service) auditPluginRegistryRefresh(event pluginRegistryRefreshEvent, status string, started time.Time, err error) {
	if s == nil || s.repo == nil {
		return
	}
	action := "plugin.registry.reload." + strings.TrimSpace(status)
	if status == "success" && strings.TrimSpace(event.Trigger) != "" {
		action = "plugin.registry.reload." + strings.TrimSpace(event.Trigger)
	}
	errorMessage := ""
	if err != nil {
		errorMessage = strings.TrimSpace(err.Error())
		if len(errorMessage) > 300 {
			errorMessage = errorMessage[:300]
		}
	}
	target := "plugins#registry"
	if event.PluginCode != "" {
		target = fmt.Sprintf("plugins#%s", event.PluginCode)
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(event.ActorName, event.ActorType, "system"),
		ActorType: firstNonEmpty(event.ActorType, "system"),
		ActorID:   event.ActorID,
		Action:    action,
		Target:    target,
		Metadata: mustJSON(map[string]any{
			"plugin_code":  event.PluginCode,
			"trigger":      event.Trigger,
			"community_id": event.CommunityID,
			"duration_ms":  int64(time.Since(started).Milliseconds()),
			"old_version":  event.OldVersion,
			"new_version":  event.NewVersion,
			"status":       firstNonEmpty(status, event.Status),
			"error":        errorMessage,
		}),
		CreatedAt: Now(),
	})
}

func (s *Service) setPluginRegistryRefreshFailureForTest(fn func(pluginRegistryRefreshEvent) error) {
	s.pluginRuntimeErr = fn
}
