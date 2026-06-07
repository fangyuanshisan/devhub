package plugins

import (
	"fmt"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

const (
	FrontendMountHomeSection            = "frontend.home.section"
	FrontendMountCommunitySection       = "frontend.community.section"
	FrontendMountAdminPluginPreview     = "admin.plugin.detail.preview"
	FrontendComponentAnnouncementCard   = "official.announcement.card"
	FrontendRenderModeOfficialComponent = "official_component"
)

type FrontendMountPointSpec struct {
	MountPoint  string
	Area        string
	Description string
}

type FrontendComponentSpec struct {
	ComponentKey       string
	Description        string
	AllowedMountPoints []string
	Deprecated         bool
}

var frontendMountAllowlist = map[string]FrontendMountPointSpec{
	FrontendMountHomeSection: {
		MountPoint:  FrontendMountHomeSection,
		Area:        "frontend",
		Description: "前台首页官方受控挂载位",
	},
	FrontendMountCommunitySection: {
		MountPoint:  FrontendMountCommunitySection,
		Area:        "frontend",
		Description: "子站首页官方受控挂载位",
	},
	FrontendMountAdminPluginPreview: {
		MountPoint:  FrontendMountAdminPluginPreview,
		Area:        "admin",
		Description: "后台插件详情官方预览挂载位",
	},
}

var frontendComponentAllowlist = map[string]FrontendComponentSpec{
	FrontendComponentAnnouncementCard: {
		ComponentKey:       FrontendComponentAnnouncementCard,
		Description:        "官方公告卡片",
		AllowedMountPoints: []string{FrontendMountHomeSection, FrontendMountCommunitySection, FrontendMountAdminPluginPreview},
	},
}

func FrontendMountPointSpecs() []FrontendMountPointSpec {
	out := make([]FrontendMountPointSpec, 0, len(frontendMountAllowlist))
	for _, spec := range frontendMountAllowlist {
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MountPoint < out[j].MountPoint })
	return out
}

func FrontendComponentSpecs() []FrontendComponentSpec {
	out := make([]FrontendComponentSpec, 0, len(frontendComponentAllowlist))
	for _, spec := range frontendComponentAllowlist {
		spec.AllowedMountPoints = append([]string(nil), spec.AllowedMountPoints...)
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ComponentKey < out[j].ComponentKey })
	return out
}

func IsFrontendMountPointAllowed(mountPoint string) bool {
	_, ok := frontendMountAllowlist[strings.TrimSpace(mountPoint)]
	return ok
}

func IsFrontendComponentAllowed(componentKey string) bool {
	_, ok := frontendComponentAllowlist[strings.TrimSpace(componentKey)]
	return ok
}

func NormalizeFrontendMount(mount domain.FrontendMountDefinition, pluginCode string) domain.FrontendMountDefinition {
	mount.PluginCode = firstNonBlank(mount.PluginCode, pluginCode)
	mount.MountPoint = strings.TrimSpace(firstNonBlank(mount.MountPoint, mount.Slot))
	mount.Slot = strings.TrimSpace(mount.Slot)
	mount.ComponentKey = strings.TrimSpace(mount.ComponentKey)
	mount.RenderMode = strings.TrimSpace(mount.RenderMode)
	mount.ConfigRef = strings.TrimSpace(mount.ConfigRef)
	return mount
}

func ValidateFrontendMount(mount domain.FrontendMountDefinition) []string {
	mount = NormalizeFrontendMount(mount, mount.PluginCode)
	errs := []string{}
	if mount.MountPoint == "" {
		errs = append(errs, "FRONTEND_MOUNT_NOT_ALLOWED：前端挂载点不能为空")
	} else if !IsFrontendMountPointAllowed(mount.MountPoint) {
		errs = append(errs, fmt.Sprintf("FRONTEND_MOUNT_NOT_ALLOWED：前端挂载点不在官方允许列表中：%s", mount.MountPoint))
	}
	if mount.ComponentKey == "" {
		errs = append(errs, "FRONTEND_COMPONENT_NOT_ALLOWED：前端组件不能为空")
	} else if !IsFrontendComponentAllowed(mount.ComponentKey) {
		errs = append(errs, fmt.Sprintf("FRONTEND_COMPONENT_NOT_ALLOWED：前端组件不在官方允许列表中：%s", mount.ComponentKey))
	}
	if mount.RenderMode != "" && mount.RenderMode != FrontendRenderModeOfficialComponent {
		errs = append(errs, fmt.Sprintf("FRONTEND_RENDER_MODE_NOT_ALLOWED：当前版本只支持官方组件挂载，不支持 render_mode=%s", mount.RenderMode))
	}
	if mount.MountPoint != "" && mount.ComponentKey != "" {
		if spec, ok := frontendComponentAllowlist[mount.ComponentKey]; ok && !stringInSlice(mount.MountPoint, spec.AllowedMountPoints) {
			errs = append(errs, fmt.Sprintf("FRONTEND_COMPONENT_NOT_ALLOWED：组件 %s 不允许挂载到 %s", mount.ComponentKey, mount.MountPoint))
		}
	}
	return errs
}

func SanitizeFrontendMountProps(props map[string]any) map[string]any {
	if len(props) == 0 {
		return nil
	}
	out := map[string]any{}
	for key, value := range props {
		if isSensitiveFrontendPropKey(key) {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isSensitiveFrontendPropKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
	return strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "authorization") ||
		strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "credential")
}

func stringInSlice(value string, items []string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
