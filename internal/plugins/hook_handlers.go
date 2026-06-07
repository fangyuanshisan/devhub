package plugins

import (
	"errors"
	"strings"

	pluginaiworks "devhub-gin-backend/internal/plugins/aiworks"
	plugindocs "devhub-gin-backend/internal/plugins/docs"
	pluginjobs "devhub-gin-backend/internal/plugins/jobs"
	pluginprojects "devhub-gin-backend/internal/plugins/projects"
	pluginqa "devhub-gin-backend/internal/plugins/qa"
	pluginwiki "devhub-gin-backend/internal/plugins/wiki"
)

// RegisterBuiltinHookHandlers wires minimal hook handlers for built-in system plugins.
// This is NOT a dynamic loader; it's a compile-time registry for governance hooks.
func RegisterBuiltinHookHandlers(bus *HookBus) {
	if bus == nil {
		return
	}
	registerQA(bus)
	registerDocs(bus)
	registerWiki(bus)
	// projects/jobs/ai_works currently only have platform declarations; keep hooks empty.
	_ = pluginprojects.Code
	_ = pluginjobs.Code
	_ = pluginaiworks.Code
}

func registerQA(bus *HookBus) {
	bus.Register(HookBeforeCreateContent, pluginqa.Code, func(e HookEvent) error {
		if NormalizeContentType(e.Ctx.ContentType) != "question" {
			return errors.New("qa 插件仅允许创建 question")
		}
		return nil
	})
	// Example non-blocking hook: best-effort side effect placeholder.
	bus.Register(HookAfterCreateContent, pluginqa.Code, func(e HookEvent) error { return nil })
}

func registerDocs(bus *HookBus) {
	bus.Register(HookBeforeCreateContent, plugindocs.Code, func(e HookEvent) error {
		if NormalizeContentType(e.Ctx.ContentType) != "document" {
			return errors.New("docs 插件仅允许创建 document")
		}
		return nil
	})
	bus.Register(HookAfterCreateContent, plugindocs.Code, func(e HookEvent) error { return nil })
}

func registerWiki(bus *HookBus) {
	bus.Register(HookBeforeCreateContent, pluginwiki.Code, func(e HookEvent) error {
		if NormalizeContentType(e.Ctx.ContentType) != "wiki_page" {
			return errors.New("wiki 插件仅允许创建 wiki_page")
		}
		return nil
	})
	bus.Register(HookBeforeUpdateContent, pluginwiki.Code, func(e HookEvent) error {
		// Minimal guard: do not allow empty content type for wiki update hooks.
		if strings.TrimSpace(e.Ctx.ContentType) == "" {
			return errors.New("wiki 更新缺少 content_type")
		}
		return nil
	})
	bus.Register(HookAfterUpdateContent, pluginwiki.Code, func(e HookEvent) error { return nil })
}
