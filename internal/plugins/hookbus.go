package plugins

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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
	failures map[string]HookFailureRule
	mu       sync.RWMutex
}

func NewHookBus() *HookBus {
	return &HookBus{
		handlers: map[string]map[string][]HookHandler{},
		failures: map[string]HookFailureRule{},
	}
}

// HookFailureRule is a test/dev-only failure injection rule.
// It is used to exercise HookBus governance without changing built-in plugin code.
type HookFailureRule struct {
	PluginCode string
	HookName   string
	Mode       HookMode
	Error      string
}

func hookFailureKey(pluginCode, hookName string) string {
	return strings.TrimSpace(pluginCode) + "\x00" + strings.TrimSpace(hookName)
}

// SetFailureInjection sets or clears a test/dev-only HookBus failure rule.
func (b *HookBus) SetFailureInjection(rule HookFailureRule) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	key := hookFailureKey(rule.PluginCode, rule.HookName)
	if key == "\x00" {
		return
	}
	if strings.TrimSpace(rule.Error) == "" {
		delete(b.failures, key)
		return
	}
	if rule.Mode == "" {
		rule.Mode = HookBlocking
	}
	b.failures[key] = rule
}

func (b *HookBus) failureInjection(pluginCode, hookName string) (HookFailureRule, bool) {
	if b == nil {
		return HookFailureRule{}, false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	rule, ok := b.failures[hookFailureKey(pluginCode, hookName)]
	return rule, ok
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
		if rule, ok := b.failureInjection(pluginCode, name); ok {
			return injectedFailureResult(event, rule)
		}
		return nil, nil
	}
	handlers := group[pluginCode]
	if len(handlers) == 0 {
		if rule, ok := b.failureInjection(pluginCode, name); ok {
			return injectedFailureResult(event, rule)
		}
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
	if rule, ok := b.failureInjection(pluginCode, name); ok {
		injected, injectedErr := injectedFailureResult(event, rule)
		results = append(results, injected...)
		if injectedErr != nil {
			if len(injected) > 0 && injected[0].Blocking {
				return results, injectedErr
			}
			nonBlockingErrs = append(nonBlockingErrs, injectedErr)
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

func injectedFailureResult(event HookEvent, rule HookFailureRule) ([]HookExecutionResult, error) {
	started := time.Now()
	finished := time.Now()
	message := strings.TrimSpace(rule.Error)
	if message == "" {
		message = "injected hook failure"
	}
	err := errors.New(message)
	mode := event.Mode
	if rule.Mode != "" {
		mode = rule.Mode
	}
	result := HookExecutionResult{
		HookName:     strings.TrimSpace(event.Name),
		PluginCode:   strings.TrimSpace(event.Ctx.PluginCode),
		Mode:         mode,
		Blocking:     mode == HookBlocking,
		HandlerIndex: -1,
		StartedAt:    started,
		FinishedAt:   finished,
		DurationMS:   int(finished.Sub(started).Milliseconds()),
		Success:      false,
		ErrorMessage: message,
	}
	return []HookExecutionResult{result}, err
}
