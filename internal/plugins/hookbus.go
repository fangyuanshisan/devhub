package plugins

import (
	"errors"
	"fmt"
	"strings"

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
	RequestID   string              `json:"request_id,omitempty"`
	PluginCode  string              `json:"plugin_code,omitempty"`
	ContentType string              `json:"content_type,omitempty"`
	CommunityID int64               `json:"community_id,omitempty"`
	CategoryID  int64               `json:"category_id,omitempty"`
	ContentID   int64               `json:"content_id,omitempty"`
	ActorType   HookActorType       `json:"actor_type,omitempty"`
	ActorID     int64               `json:"actor_id,omitempty"`
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

// Dispatch executes hook handlers.
//
// For blocking hooks, the first error stops execution and is returned.
// For non-blocking hooks, errors are aggregated and returned as a single error,
// so callers can decide whether to log/audit without blocking the main flow.
func (b *HookBus) Dispatch(event HookEvent) error {
	if b == nil {
		return nil
	}
	name := strings.TrimSpace(event.Name)
	pluginCode := strings.TrimSpace(event.Ctx.PluginCode)
	if name == "" || pluginCode == "" {
		return nil
	}
	group := b.handlers[name]
	if len(group) == 0 {
		return nil
	}
	handlers := group[pluginCode]
	if len(handlers) == 0 {
		return nil
	}

	var nonBlockingErrs []error
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if err := handler(event); err != nil {
			if event.Mode == HookBlocking {
				return err
			}
			nonBlockingErrs = append(nonBlockingErrs, err)
		}
	}
	if len(nonBlockingErrs) == 0 {
		return nil
	}
	if len(nonBlockingErrs) == 1 {
		return nonBlockingErrs[0]
	}
	return errors.New(fmt.Sprintf("hook %s non-blocking errors: %v", name, nonBlockingErrs))
}
