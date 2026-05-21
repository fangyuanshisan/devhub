package service

import (
	"fmt"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

// FrontendMountsForRuntime returns only official allowlisted frontend mounts
// from enabled plugins. It does not load plugin-provided JS, HTML, or iframe
// URLs; unknown historical declarations are skipped with warnings.
func (s *Service) FrontendMountsForRuntime(mountPoint string, communityID int64) domain.PluginFrontendMountRuntimeResult {
	mountPoint = strings.TrimSpace(mountPoint)
	result := domain.PluginFrontendMountRuntimeResult{
		MountPoint: mountPoint,
		Items:      []domain.PluginFrontendMountRuntimeItem{},
		Warnings:   []string{},
	}
	if !pluginregistry.IsFrontendMountPointAllowed(mountPoint) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("FRONTEND_MOUNT_NOT_ALLOWED：前端挂载点不在官方允许列表中：%s", mountPoint))
		return result
	}
	for _, plugin := range s.repo.Plugins() {
		if plugin.Status != pluginregistry.StatusEnabled && plugin.Status != pluginregistry.StatusRunning {
			continue
		}
		if communityID > 0 && !s.IsPluginEnabledForCommunity(communityID, plugin.Code) {
			continue
		}
		for _, rawMount := range plugin.FrontendMounts {
			mount := pluginregistry.NormalizeFrontendMount(rawMount, plugin.Code)
			if mount.MountPoint != mountPoint {
				continue
			}
			if errs := pluginregistry.ValidateFrontendMount(mount); len(errs) > 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s：当前版本不支持该前端挂载声明，已跳过", plugin.Code))
				continue
			}
			result.Items = append(result.Items, domain.PluginFrontendMountRuntimeItem{
				PluginCode:   plugin.Code,
				MountPoint:   mount.MountPoint,
				ComponentKey: mount.ComponentKey,
				Status:       "valid",
				Message:      "官方 allowlist 挂载",
				Props:        pluginregistry.SanitizeFrontendMountProps(mount.Props),
				ConfigRef:    mount.ConfigRef,
			})
		}
	}
	return result
}
