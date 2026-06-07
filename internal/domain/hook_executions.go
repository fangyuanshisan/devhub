package domain

import "strings"

// HookExecutionFilter is used by admin troubleshooting APIs to query hook_executions.
// It intentionally stays lightweight and does NOT imply full-text search or archival features.
type HookExecutionFilter struct {
	PluginCode  string
	HookName    string
	ServiceType string
	Mode        string
	Blocking    *bool
	Success     *bool
	ContentType string
	ContentID   int64
	CommunityID int64
	ActorType   string
	ActorID     int64
	RequestID   string
	StartTime   string
	EndTime     string
	Page        int
	PageSize    int
}

func (f HookExecutionFilter) Normalize() HookExecutionFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.HookName = strings.TrimSpace(f.HookName)
	f.ServiceType = strings.TrimSpace(f.ServiceType)
	f.Mode = strings.TrimSpace(f.Mode)
	f.ContentType = strings.TrimSpace(f.ContentType)
	f.ActorType = strings.TrimSpace(f.ActorType)
	f.RequestID = strings.TrimSpace(f.RequestID)
	f.StartTime = strings.TrimSpace(f.StartTime)
	f.EndTime = strings.TrimSpace(f.EndTime)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}
