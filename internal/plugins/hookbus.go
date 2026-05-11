package plugins

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
)

// HookName is the stable name for a hook event.
// Keep these as constants to avoid string scattering.
const (
	HookBeforeCreateContent   = "BeforeCreateContent"
	HookAfterCreateContent    = "AfterCreateContent"
	HookBeforeUpdateContent   = "BeforeUpdateContent"
	HookAfterUpdateContent    = "AfterUpdateContent"
	HookBeforeModerateContent = "BeforeModerateContent"
	HookAfterModerateContent  = "AfterModerateContent"
	HookBeforeBuildSEO        = "BeforeBuildSEO"
	HookAfterBuildSEO         = "AfterBuildSEO"
	HookAfterPluginEnabled    = "AfterPluginEnabled"
	HookAfterPluginDisabled   = "AfterPluginDisabled"
	HookAfterCreateComment    = "AfterCreateComment"
	HookOnSearchIndex         = "OnSearchIndex"
	HookOnNotificationBuild   = "OnNotificationBuild"
	HookOnSEOBuild            = "OnSEOBuild"
)

// HookMode controls whether a hook can block the main flow.
type HookMode string

const (
	HookBlocking    HookMode = "blocking"
	HookNonBlocking HookMode = "non_blocking"
)

// HookActorType describes who triggers the hook.
type HookActorType string

const (
	HookActorUser      HookActorType = "user"
	HookActorAdmin     HookActorType = "admin_user"
	HookActorModerator HookActorType = "moderator"
	HookActorSystem    HookActorType = "system"
)

// HookContext is the governance-friendly context carried by hook execution.
// It is intentionally decoupled from HTTP layer request types.
type HookContext struct {
	HookName    string              `json:"hook_name,omitempty"`
	RequestID   string              `json:"request_id,omitempty"`
	PluginCode  string              `json:"plugin_code,omitempty"`
	ContentType string              `json:"content_type,omitempty"`
	ChannelID   int64               `json:"channel_id,omitempty"`
	CommunityID int64               `json:"community_id,omitempty"`
	CategoryID  int64               `json:"category_id,omitempty"`
	ContentID   int64               `json:"content_id,omitempty"`
	ActorType   HookActorType       `json:"actor_type,omitempty"`
	ActorID     int64               `json:"actor_id,omitempty"`
	UserID      int64               `json:"user_id,omitempty"`
	AdminUserID int64               `json:"admin_user_id,omitempty"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
	Actor       domain.ActorContext `json:"-"`
}

// HookEvent carries typed payloads for hook handlers.
// Some events only have Context+PluginCode; others carry Topic/Requests.
type HookEvent struct {
	Name string
	Mode HookMode
	Ctx  HookContext

	Topic         *domain.Topic
	PreviousTopic *domain.Topic
	Request       *domain.CreateTopicRequest
	UpdateRequest *domain.UpdateTopicRequest
	Comment       *domain.Comment
	SearchRequest *domain.SearchRequest
	SearchResults []domain.Topic
	Notification  *domain.Notification
	SEOHTML       string
}

// HookHandler is a built-in plugin hook handler.
type HookHandler func(HookEvent) error

// HookExecutionResult is the in-memory result emitted by HookBus dispatch.
// Service/store layers decide how to persist it.
type HookExecutionResult struct {
	HookName     string
	PluginCode   string
	Mode         HookMode
	Blocking     bool
	HandlerIndex int
	StartedAt    time.Time
	FinishedAt   time.Time
	DurationMS   int
	Success      bool
	ErrorMessage string
}

// HookBus dispatches hooks for built-in system plugins.
// It is NOT a third-party dynamic plugin execution environment.
type HookBus struct {
	handlers map[string]map[string][]HookHandler // name -> plugin_code -> handlers
}

func NewHookBus() *HookBus {
	return &HookBus{handlers: map[string]map[string][]HookHandler{}}
}

// Register registers a handler for a hook name and plugin code.
func (b *HookBus) Register(name, pluginCode string, handler HookHandler) {
	name = strings.TrimSpace(name)
	pluginCode = strings.TrimSpace(pluginCode)
	if b == nil || name == "" || pluginCode == "" || handler == nil {
		return
	}
	if _, ok := b.handlers[name]; !ok {
		b.handlers[name] = map[string][]HookHandler{}
	}
	b.handlers[name][pluginCode] = append(b.handlers[name][pluginCode], handler)
}

// HandlerCount returns registered handler count for a hook/plugin pair.
func (b *HookBus) HandlerCount(name, pluginCode string) int {
	if b == nil {
		return 0
	}
	return len(b.handlers[strings.TrimSpace(name)][strings.TrimSpace(pluginCode)])
}

// Dispatch executes hook handlers.
//
// For blocking hooks, the first error stops execution and is returned.
// For non-blocking hooks, errors are aggregated and returned as a single error,
// so callers can decide whether to log/audit without blocking the main flow.
func (b *HookBus) Dispatch(event HookEvent) error {
	_, err := b.DispatchWithResults(event)
	return err
}

// DispatchWithResults executes hook handlers and returns per-handler execution
// results so the platform can persist observability records.
func (b *HookBus) DispatchWithResults(event HookEvent) ([]HookExecutionResult, error) {
	if b == nil {
		return nil, nil
	}
	name := strings.TrimSpace(event.Name)
	pluginCode := strings.TrimSpace(event.Ctx.PluginCode)
	if name == "" || pluginCode == "" {
		return nil, nil
	}
	group := b.handlers[name]
	if len(group) == 0 {
		return nil, nil
	}
	handlers := group[pluginCode]
	if len(handlers) == 0 {
		return nil, nil
	}

	var nonBlockingErrs []error
	results := make([]HookExecutionResult, 0, len(handlers))
	for i, handler := range handlers {
		if handler == nil {
			continue
		}
		started := time.Now()
		err := handler(event)
		finished := time.Now()
		result := HookExecutionResult{
			HookName:     name,
			PluginCode:   pluginCode,
			Mode:         event.Mode,
			Blocking:     event.Mode == HookBlocking,
			HandlerIndex: i,
			StartedAt:    started,
			FinishedAt:   finished,
			DurationMS:   int(finished.Sub(started).Milliseconds()),
			Success:      err == nil,
		}
		if err != nil {
			result.ErrorMessage = err.Error()
		}
		results = append(results, result)
		if err != nil {
			if event.Mode == HookBlocking {
				return results, err
			}
			nonBlockingErrs = append(nonBlockingErrs, err)
		}
	}
	if len(nonBlockingErrs) == 0 {
		return results, nil
	}
	if len(nonBlockingErrs) == 1 {
		return results, nonBlockingErrs[0]
	}
	return results, errors.New(fmt.Sprintf("hook %s non-blocking errors: %v", name, nonBlockingErrs))
}
