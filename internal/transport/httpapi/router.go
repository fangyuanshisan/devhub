package httpapi

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// Server 保存 HTTP 处理器依赖的业务服务。
type Server struct {
	svc *service.Service
}

// NewRouter 创建 Gin 路由，并挂载前台、后台 API 以及静态资源。
func NewRouter(svc *service.Service) *gin.Engine {
	srv := &Server{svc: svc}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(), srv.optionalAuth())

	api := r.Group("/api/v1")
	{
		api.GET("/health", srv.health)
		auth := api.Group("/auth")
		{
			auth.POST("/register", srv.register)
			auth.POST("/login", srv.frontLogin)
			auth.POST("/refresh", srv.refreshSession)
			auth.POST("/logout", srv.logout)
			auth.GET("/me", srv.authRequired(), srv.authMe)
		}
		api.GET("/sites", srv.listSites)
		api.GET("/sites/:site", srv.getSite)
		api.GET("/sites/:site/overview", srv.siteOverview)
		api.GET("/boards", srv.listBoards)
		api.GET("/plugins", srv.plugins)
		api.GET("/stats", srv.stats)
		api.GET("/tags", srv.tags)
		api.GET("/tags/hot", srv.hotTags)
		api.GET("/tags/suggestions", srv.tagSuggestions)
		api.GET("/tags/suggest", srv.tagSuggestions)
		api.GET("/tags/by-slug/:tag", srv.getTag)
		api.GET("/tags/:tag", srv.getTag)
		api.GET("/tags/:tag/topics", srv.tagTopics)
		api.GET("/posts", srv.listPosts)
		api.POST("/posts", srv.authRequired(), srv.createPost)
		api.GET("/posts/:id", srv.getPost)
		api.PUT("/posts/:id", srv.authRequired(), srv.updatePost)
		api.DELETE("/posts/:id", srv.authRequired(), srv.deletePost)
		api.POST("/posts/:id/like", srv.authRequired(), srv.likePost)
		api.GET("/posts/:id/comments", srv.comments)
		api.POST("/posts/:id/comments", srv.authRequired(), srv.createComment)
		api.POST("/comments/:id/like", srv.authRequired(), srv.likeComment)
		api.DELETE("/comments/:id", srv.authRequired(), srv.deleteComment)
		api.GET("/search", srv.search)
		api.GET("/hot", srv.hot)
		api.GET("/feed", srv.feed)
		api.GET("/notifications", srv.notifications)
		api.GET("/notifications/unread-count", srv.unreadNotifications)
		api.POST("/notifications/read-all", srv.authRequired(), srv.readAllNotifications)
		api.POST("/notifications/:id/read", srv.authRequired(), srv.readNotification)
		api.GET("/users/me", srv.authRequired(), srv.me)

		// ===== 新增：DevHub 通用社区系统 API =====
		api.GET("/communities", srv.listCommunities)
		api.GET("/communities/:slug", srv.getCommunity)
		api.GET("/communities/:slug/home", srv.communityOverview)
		api.GET("/communities/:slug/overview", srv.communityOverview)
		api.GET("/communities/:slug/stats", srv.communityStats)
		api.GET("/communities/:slug/categories", srv.listCategories)
		api.GET("/communities/:slug/plugins", srv.communityPlugins)
		api.GET("/communities/:slug/tags", srv.listCommunityTags)
		api.GET("/communities/:slug/tags/:tag", srv.getCommunityTag)
		api.GET("/communities/:slug/tags/:tag/topics", srv.communityTagTopics)
		api.GET("/communities/:slug/moderators", srv.listCommunityModerators)
		api.GET("/topics", srv.listTopics)
		api.GET("/topics/:id", srv.getTopic)
		api.GET("/topics/:id/qa", srv.topicQA)
		api.GET("/topics/:id/docs", srv.topicDocs)
		api.GET("/topics/:id/wiki/versions", srv.topicWikiVersions)
		api.GET("/topics/:id/comments", srv.topicComments)
		api.POST("/topics/:id/comments", srv.userAuthRequired(), srv.createTopicComment)
		api.POST("/topics/:id/comments/:commentId/replies", srv.userAuthRequired(), srv.replyTopicComment)
		api.POST("/topics/:id/comments/:commentId/accept", srv.userAuthRequired(), srv.acceptTopicComment)
		api.POST("/topics", srv.userAuthRequired(), srv.createTopic)
		api.PUT("/topics/:id", srv.userAuthRequired(), srv.updateTopic)
		api.DELETE("/topics/:id", srv.userAuthRequired(), srv.deleteTopic)
		api.POST("/topics/:id/like", srv.userAuthRequired(), srv.likeTopic)
		api.POST("/topics/:id/favorite", srv.userAuthRequired(), srv.favoriteTopic)
		api.GET("/topics/:id/interaction", srv.topicInteraction)
		api.POST("/topics/:id/solve", srv.userAuthRequired(), srv.solveTopic)
		api.GET("/search/topics", srv.searchTopics)
		api.POST("/actions/toggle", srv.userAuthRequired(), srv.toggleReaction)
		api.POST("/reactions/toggle", srv.userAuthRequired(), srv.toggleReaction)
		api.POST("/favorites/toggle", srv.userAuthRequired(), srv.toggleFavorite)
		api.POST("/follows/toggle", srv.userAuthRequired(), srv.toggleFollow)
		api.GET("/activities", srv.userAuthRequired(), srv.userActivities)
		api.GET("/me/favorites", srv.userAuthRequired(), srv.myFavorites)
		api.GET("/me/follows", srv.userAuthRequired(), srv.myFollows)
		api.GET("/me/activities", srv.userAuthRequired(), srv.myActivities)
		api.GET("/me/notifications", srv.userAuthRequired(), srv.myNotifications)
		api.POST("/me/notifications/read-all", srv.userAuthRequired(), srv.readAllMyNotifications)
		api.POST("/me/notifications/:id/read", srv.userAuthRequired(), srv.readMyNotification)
		api.POST("/reports", srv.userAuthRequired(), srv.createReport)

		moderator := api.Group("/moderator", srv.moderatorAuthRequired())
		{
			moderator.GET("/communities", srv.moderatorCommunities)
			moderator.GET("/plugin-menus", srv.moderatorPluginMenus)
			moderator.GET("/dashboard", srv.moderatorDashboard)
			moderator.GET("/reports", srv.moderatorReports)
			moderator.POST("/reports/:id/handle", srv.handleModeratorReport)
			moderator.GET("/topics", srv.moderatorTopics)
			moderator.POST("/topics/:id/feature", srv.featureModeratorTopic)
			moderator.POST("/topics/:id/unfeature", srv.unfeatureModeratorTopic)
			moderator.POST("/topics/:id/pin", srv.pinModeratorTopic)
			moderator.POST("/topics/:id/unpin", srv.unpinModeratorTopic)
			moderator.POST("/topics/:id/hide", srv.hideModeratorTopic)
			moderator.POST("/topics/:id/restore", srv.restoreModeratorTopic)
			moderator.POST("/topics/:id/lock-comments", srv.lockModeratorTopicComments)
			moderator.POST("/topics/:id/unlock-comments", srv.unlockModeratorTopicComments)
			moderator.GET("/comments", srv.moderatorComments)
			moderator.POST("/comments/:id/hide", srv.hideModeratorComment)
			moderator.POST("/comments/:id/restore", srv.restoreModeratorComment)
			moderator.GET("/audit-logs", srv.moderatorAuditLogs)
		}

		admin := api.Group("/admin")
		{
			admin.POST("/login", srv.adminLogin)
			admin.POST("/refresh", srv.refreshAdminSession)
			admin.POST("/logout", srv.logout)
			protected := admin.Group("", srv.adminAuthRequired(), srv.adminContext())
			protected.GET("/me", srv.adminMe)
			protected.GET("/plugins", srv.requirePermission("plugin.read"), srv.adminPlugins)
			protected.GET("/plugin-menus", srv.adminPluginMenus)
			protected.GET("/plugins/:code/impact", srv.requirePermission("plugin.read"), srv.adminPluginImpact)
			protected.GET("/plugins/:code/hooks", srv.requirePermission("plugin.read"), srv.adminPluginHooks)
			protected.POST("/plugins/:code/hooks/:name/e2e-fail", srv.requirePermission("plugin.write"), srv.injectFailedAdminPluginHookForTest)
			protected.GET("/plugins/:code/audit-logs", srv.requirePermission("plugin.read"), srv.adminPluginAuditLogs)
			protected.GET("/plugins/:code/migrations", srv.requirePermission("plugin.read"), srv.adminPluginMigrations)
			protected.POST("/plugins/:code/migrations/run", srv.requirePermission("plugin.write"), srv.runAdminPluginMigrations)
			protected.POST("/plugins/:code/migrations/:name/retry", srv.requirePermission("plugin.write"), srv.retryAdminPluginMigration)
			protected.POST("/plugins/:code/migrations/:name/e2e-fail", srv.requirePermission("plugin.write"), srv.injectFailedAdminPluginMigrationForTest)
			protected.POST("/plugins/:code/enable", srv.requirePermission("plugin.write"), srv.enableAdminPlugin)
			protected.POST("/plugins/:code/disable", srv.requirePermission("plugin.write"), srv.disableAdminPlugin)
			protected.POST("/plugins/:code/archive", srv.requirePermission("plugin.write"), srv.archiveAdminPlugin)
			protected.POST("/plugins/:code/restore", srv.requirePermission("plugin.write"), srv.restoreAdminPlugin)
			protected.PUT("/plugins/:code/config", srv.requirePermission("plugin.write"), srv.updateAdminPluginConfig)
			protected.GET("/overview", srv.requirePermission("dashboard.read"), srv.adminOverview)
			protected.GET("/communities", srv.requirePermission("site.read"), srv.adminCommunities)
			protected.POST("/communities", srv.requirePermission("site.write"), srv.createAdminCommunity)
			protected.GET("/communities/:id", srv.requirePermission("site.read"), srv.adminCommunity)
			protected.PUT("/communities/:id", srv.requirePermission("site.write"), srv.updateAdminCommunity)
			protected.POST("/communities/:id/enable", srv.requirePermission("site.write"), srv.enableAdminCommunity)
			protected.POST("/communities/:id/disable", srv.requirePermission("site.write"), srv.disableAdminCommunity)
			protected.POST("/communities/reorder", srv.requirePermission("site.write"), srv.reorderAdminCommunities)
			protected.GET("/communities/:id/plugins", srv.requirePermission("site.read"), srv.adminCommunityPlugins)
			protected.GET("/communities/:id/plugins/:code/impact", srv.requirePermission("site.read"), srv.adminCommunityPluginImpact)
			protected.POST("/communities/:id/plugins/:code/enable", srv.requirePermission("site.write"), srv.enableAdminCommunityPlugin)
			protected.POST("/communities/:id/plugins/:code/disable", srv.requirePermission("site.write"), srv.disableAdminCommunityPlugin)
			protected.PUT("/communities/:id/plugins/:code/config", srv.requirePermission("site.write"), srv.updateAdminCommunityPluginConfig)
			protected.PUT("/communities/:id/plugins/sort", srv.requirePermission("site.write"), srv.reorderAdminCommunityPlugins)
			protected.GET("/communities/:id/categories", srv.requirePermission("board.read"), srv.adminCommunityCategories)
			protected.POST("/communities/:id/categories", srv.requirePermission("board.write"), srv.createAdminCommunityCategory)
			protected.PUT("/categories/:id", srv.requirePermission("board.write"), srv.updateAdminCategory)
			protected.POST("/categories/:id/enable", srv.requirePermission("board.write"), srv.enableAdminCategory)
			protected.POST("/categories/:id/disable", srv.requirePermission("board.write"), srv.disableAdminCategory)
			protected.POST("/categories/reorder", srv.requirePermission("board.write"), srv.reorderAdminCategories)
			protected.GET("/posts", srv.requirePermission("post.read"), srv.adminPosts)
			protected.POST("/posts", srv.requirePermission("post.create"), srv.createAdminPost)
			protected.PUT("/posts/:id", srv.requirePermission("post.update"), srv.updateAdminPost)
			protected.DELETE("/posts/:id", srv.requirePermission("post.delete"), srv.deleteAdminPost)
			protected.POST("/topics/batch", srv.requirePermission("topic.moderate"), srv.batchAdminTopics)
			protected.POST("/sites", srv.requirePermission("site.write"), srv.createAdminSite)
			protected.PUT("/sites/:site", srv.requirePermission("site.write"), srv.updateAdminSite)
			protected.POST("/boards", srv.requirePermission("board.write"), srv.createAdminBoard)
			protected.PUT("/boards/:board", srv.requirePermission("board.write"), srv.updateAdminBoard)
			protected.GET("/users", srv.requirePermission("user.read"), srv.adminUsers)
			protected.PUT("/users/:id/status", srv.requirePermission("user.write"), srv.updateAdminUserStatus)
			protected.GET("/roles", srv.requirePermission("role.read"), srv.adminRoles)
			protected.GET("/permissions", srv.requirePermission("role.read"), srv.adminPermissions)
			protected.GET("/tags", srv.requirePermission("post.read"), srv.adminTags)
			protected.GET("/tags/:id", srv.requirePermission("post.read"), srv.adminTag)
			protected.GET("/tags/:id/topics", srv.requirePermission("post.read"), srv.adminTagTopics)
			protected.GET("/tags/:id/aliases", srv.requirePermission("post.read"), srv.adminTagAliases)
			protected.POST("/tags", srv.requirePermission("post.update"), srv.createAdminTag)
			protected.PUT("/tags/:id", srv.requirePermission("post.update"), srv.updateAdminTag)
			protected.POST("/tags/:id/aliases", srv.requirePermission("post.update"), srv.createAdminTagAlias)
			protected.DELETE("/tags/:id/aliases/:aliasId", srv.requirePermission("post.update"), srv.deleteAdminTagAlias)
			protected.POST("/tags/:id/enable", srv.requirePermission("post.update"), srv.enableAdminTag)
			protected.POST("/tags/:id/disable", srv.requirePermission("post.update"), srv.disableAdminTag)
			protected.POST("/tags/:id/merge", srv.requirePermission("post.update"), srv.mergeAdminTag)
			protected.POST("/tags/:id/recalculate", srv.requirePermission("post.update"), srv.recalculateAdminTag)
			protected.POST("/tags/recalculate-all", srv.requirePermission("post.update"), srv.recalculateAllAdminTags)
			protected.GET("/comments", srv.requirePermission("comment.read"), srv.adminComments)
			protected.PUT("/comments/:id/status", srv.requirePermission("comment.moderate"), srv.updateAdminCommentStatus)
			protected.POST("/comments/:id/hide", srv.requirePermission("comment.moderate"), srv.hideAdminComment)
			protected.POST("/comments/:id/restore", srv.requirePermission("comment.moderate"), srv.restoreAdminComment)
			protected.DELETE("/comments/:id", srv.requirePermission("comment.moderate"), srv.deleteAdminComment)
			protected.POST("/comments/batch", srv.requirePermission("comment.moderate"), srv.batchAdminComments)
			protected.GET("/reports", srv.requirePermission("report.read"), srv.adminReports)
			protected.GET("/reports/:id", srv.requirePermission("report.read"), srv.adminReport)
			protected.POST("/reports/:id/handle", srv.requirePermission("report.handle"), srv.handleAdminReport)
			protected.POST("/reports/batch-handle", srv.requirePermission("report.handle"), srv.batchAdminReports)
			protected.GET("/moderators", srv.requirePermission("moderator.read"), srv.adminModerators)
			protected.POST("/moderators", srv.requirePermission("moderator.write"), srv.createAdminModerator)
			protected.PUT("/moderators/:id", srv.requirePermission("moderator.write"), srv.updateAdminModerator)
			protected.DELETE("/moderators/:id", srv.requirePermission("moderator.write"), srv.deleteAdminModerator)
			protected.POST("/topics/:id/feature", srv.requirePermission("topic.moderate"), srv.featureAdminTopic)
			protected.POST("/topics/:id/pin", srv.requirePermission("topic.moderate"), srv.pinAdminTopic)
			protected.POST("/topics/:id/hide", srv.requirePermission("topic.moderate"), srv.hideAdminTopic)
			protected.POST("/topics/:id/restore", srv.requirePermission("topic.moderate"), srv.restoreAdminTopic)
			protected.POST("/topics/:id/lock-comments", srv.requirePermission("topic.moderate"), srv.lockAdminTopicComments)
			protected.POST("/topics/:id/unlock-comments", srv.requirePermission("topic.moderate"), srv.unlockAdminTopicComments)
			protected.GET("/settings", srv.requirePermission("setting.read"), srv.adminSettings)
			protected.PUT("/settings", srv.requirePermission("setting.write"), srv.updateAdminSettings)
			protected.GET("/logs", srv.requirePermission("log.read"), srv.adminLogs)
			protected.GET("/audit-logs", srv.requirePermission("log.read"), srv.adminAuditLogs)
			protected.POST("/notifications", srv.requirePermission("notification.write"), srv.pushAdminNotification)
		}
	}

	r.Static("/_astro", "./web/frontend/_astro")
	r.Static("/frontend-assets", "./web/frontend/frontend-assets")
	r.Static("/admin-next/assets", "./web/admin-vue/assets")
	r.GET("/admin-next/sites", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next/communities")
	})
	r.GET("/admin-next/sites/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next/communities")
	})
	r.HEAD("/admin-next/sites", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next/communities")
	})
	r.HEAD("/admin-next/sites/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next/communities")
	})
	r.StaticFile("/", "./web/frontend/index.html")
	r.StaticFile("/index.html", "./web/frontend/index.html")
	r.StaticFile("/search", "./web/frontend/search/index.html")
	r.StaticFile("/search/", "./web/frontend/search/index.html")
	r.GET("/topics/new", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/topics/new/index.html")
	})
	r.GET("/topics/new/", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/topics/new/index.html")
	})
	r.GET("/me/favorites", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/me/favorites/index.html")
	})
	r.GET("/me/favorites/", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/me/favorites/index.html")
	})
	r.GET("/me/follows", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/me/follows/index.html")
	})
	r.GET("/me/follows/", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/me/follows/index.html")
	})
	r.GET("/me/activities", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/me/activities/index.html")
	})
	r.GET("/me/activities/", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/me/activities/index.html")
	})
	r.GET("/notifications", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/notifications/index.html")
	})
	r.GET("/notifications/", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/notifications/index.html")
	})
	r.GET("/me/notifications", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/notifications/index.html")
	})
	r.GET("/me/notifications/", func(c *gin.Context) {
		serveFrontendFile(c, "./web/frontend/notifications/index.html")
	})
	r.GET("/moderator", srv.moderatorWorkbenchPage)
	r.GET("/moderator/", srv.moderatorWorkbenchPage)
	r.GET("/moderator/reports", srv.moderatorWorkbenchPage)
	r.GET("/moderator/reports/", srv.moderatorWorkbenchPage)
	r.GET("/moderator/topics", srv.moderatorWorkbenchPage)
	r.GET("/moderator/topics/", srv.moderatorWorkbenchPage)
	r.GET("/moderator/comments", srv.moderatorWorkbenchPage)
	r.GET("/moderator/comments/", srv.moderatorWorkbenchPage)
	r.GET("/moderator/audit-logs", srv.moderatorWorkbenchPage)
	r.GET("/moderator/audit-logs/", srv.moderatorWorkbenchPage)
	r.StaticFile("/admin-next", "./web/admin-vue/index.html")
	r.StaticFile("/admin-next/", "./web/admin-vue/index.html")
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next")
	})
	r.GET("/admin/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next")
	})
	r.GET("/site/:site", srv.redirectSiteToCommunity)
	r.GET("/site/:site/", srv.redirectSiteToCommunity)
	r.HEAD("/site/:site", srv.redirectSiteToCommunity)
	r.HEAD("/site/:site/", srv.redirectSiteToCommunity)
	r.GET("/c/:site", srv.communitySEOPage)
	r.GET("/c/:site/", srv.communitySEOPage)
	r.GET("/c/:site/topics/new", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/c/%s/topics/new/index.html", c.Param("site")))
	})
	r.GET("/c/:site/topics/new/", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/c/%s/topics/new/index.html", c.Param("site")))
	})
	r.GET("/c/:site/tags/:tag", srv.communityTagSEOPage)
	r.GET("/c/:site/tags/:tag/", srv.communityTagSEOPage)
	r.GET("/site/:site/topics/new", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/site/%s/topics/new/index.html", c.Param("site")))
	})
	r.GET("/site/:site/topics/new/", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/site/%s/topics/new/index.html", c.Param("site")))
	})
	r.GET("/posts/:id", srv.redirectPostToTopic)
	r.GET("/posts/:id/", srv.redirectPostToTopic)
	r.GET("/topics/:id", srv.topicSEOPage)
	r.GET("/topics/:id/", srv.topicSEOPage)
	r.GET("/tags/:tag", srv.tagSEOPage)
	r.GET("/tags/:tag/", srv.tagSEOPage)
	r.GET("/robots.txt", srv.robots)
	r.GET("/sitemap.xml", srv.sitemap)
	r.GET("/sitemap-index.xml", srv.sitemap)
	r.GET("/admin/:site", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next?site="+c.Param("site"))
	})
	r.GET("/admin/:site/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next?site="+c.Param("site"))
	})
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
			fail(c, http.StatusNotFound, "接口不存在")
			return
		}
		if strings.HasPrefix(path, "/admin-next/") {
			c.File("./web/admin-vue/index.html")
			return
		}
		c.File("./web/frontend/index.html")
	})

	return r
}

func serveFrontendFile(c *gin.Context, file string) {
	if _, err := os.Stat(file); err == nil {
		c.File(file)
		return
	}
	c.File("./web/frontend/index.html")
}

func (s *Server) moderatorWorkbenchPage(c *gin.Context) {
	serveFrontendFile(c, "./web/frontend/moderator/index.html")
}

func (s *Server) redirectPostToTopic(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/topics/"+id+"/")
}

func (s *Server) redirectSiteToCommunity(c *gin.Context) {
	slug := strings.Trim(strings.TrimSpace(c.Param("site")), "/")
	if slug == "" {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/c/"+pathEsc(slug)+"/")
}

func (s *Server) communitySEOPage(c *gin.Context) {
	slug := strings.Trim(strings.TrimSpace(c.Param("site")), "/")
	comm, ok := s.svc.CommunityBySlug(slug)
	if !ok || comm.Slug == "" {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.communityNotFoundHTML()))
		return
	}
	if comm.Status != 1 {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.communityDisabledHTML(comm)))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.renderCommunityHTML(c, comm)))
}

func (s *Server) communityNotFoundHTML() string {
	return fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>子站不存在 - DevHub</title><meta name="description" content="该子站不存在或已被关闭。"><meta name="robots" content="noindex,follow"><link rel="canonical" href="/"><link rel="stylesheet" href="%s"></head><body><main class="article-shell"><article class="article-main"><h1>子站不存在</h1><p>该子站不存在或已被关闭。</p><p><a href="/">返回 DevHub 首页</a></p></article></main></body></html>`, esc(frontendStylesheetHref()))
}

func (s *Server) communityDisabledHTML(comm domain.Community) string {
	title := "子站暂不可用 - DevHub"
	return fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s</title><meta name="description" content="该子站暂不可用。"><meta name="robots" content="noindex,follow"><link rel="canonical" href="/c/%s/"><link rel="stylesheet" href="%s"></head><body><main class="article-shell"><article class="article-main"><h1>%s 暂不可用</h1><p>该子站当前已禁用或归档。</p><p><a href="/">返回 DevHub 首页</a></p></article></main></body></html>`, esc(title), pathEsc(comm.Slug), esc(frontendStylesheetHref()), esc(comm.Name))
}

func (s *Server) renderCommunityHTML(c *gin.Context, comm domain.Community) string {
	stats := s.svc.CommunityStats(comm.ID)
	categories := s.visibleCommunityCategories(comm.ID)
	pinned, _ := s.svc.TopicsByFilter(comm.ID, 0, "", "latest", nil, "", 1, 12)
	pinned = filterTopicsByState(pinned, func(topic domain.Topic) bool { return topic.IsPinned })
	featured, _ := s.svc.TopicsByFilter(comm.ID, 0, "", "featured", nil, "", 1, 8)
	latest, _ := s.svc.TopicsByFilter(comm.ID, 0, "", "latest", nil, "", 1, 8)
	hot, _ := s.svc.TopicsByFilter(comm.ID, 0, "", "hot", nil, "", 1, 8)
	unsolved := false
	questions, _ := s.svc.TopicsByFilter(comm.ID, 0, "question", "latest", &unsolved, "", 1, 8)
	tags := s.communityHotTags(comm.ID, comm.Slug)
	moderators, _ := s.svc.CommunityModerators(domain.CommunityModeratorFilter{CommunityID: comm.ID, Status: "1", Page: 1, PageSize: 20, ActorIsAdmin: true})

	title := firstNonEmpty(comm.SEOTitle, comm.Name+" 技术社区") + " - DevHub"
	description := firstNonEmpty(comm.SEODescription, comm.Description, comm.Slogan, comm.Name+" 技术社区")
	canonicalPath := fmt.Sprintf("/c/%s/", comm.Slug)
	canonicalURL := absoluteURL(c, canonicalPath)
	themeColor := firstNonEmpty(comm.ThemeColor, "#2563eb")
	jsonLD := fmt.Sprintf(`{"@context":"https://schema.org","@type":"WebSite","name":%q,"description":%q,"url":%q,"publisher":{"@type":"Organization","name":"DevHub"}}`, comm.Name+" - DevHub", description, canonicalURL)

	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <meta name="description" content="%s">
  <meta name="keywords" content="%s">
  <link rel="canonical" href="%s">
  <meta property="og:site_name" content="DevHub">
  <meta property="og:type" content="website">
  <meta property="og:title" content="%s">
  <meta property="og:description" content="%s">
  <meta property="og:url" content="%s">
  <meta name="theme-color" content="%s">
  <link rel="stylesheet" href="%s">
  <script type="application/ld+json">%s</script>
</head>
<body>
  <header class="site-header">
    <div class="top-accent"></div>
    <div class="header-inner">
      <a class="brand" href="/"><span>DH</span><strong>DevHub</strong></a>
      <nav class="site-nav" aria-label="主导航">
        <a href="/">首页</a>
        <a href="/c/%s/">%s</a>
        <a href="/search/?scope=community&amp;community_slug=%s">搜索</a>
        <a href="/me/follows">我的关注</a>
        <a href="/notifications">通知</a>
      </nav>
      <div class="header-actions">
        <a class="publish-link" href="/c/%s/topics/new/">发布内容</a>
        <a href="/me/follows">我的关注</a>
        <a class="is-hidden" href="/moderator" data-moderator-entry>版主工作台</a>
        <span data-user-status><a href="/#login">登录</a></span>
      </div>
    </div>
  </header>
  <main>
    <section class="site-hero community-hero" style="--accent:%s" data-community-hero data-community-id="%d" data-community="%s">
      <span>%s 子站</span>
      <h1>%s</h1>
      <p>%s</p>
      <div class="hero-actions">
        <a class="primary-link" href="/c/%s/topics/new/">在当前子站发布</a>
        <button class="secondary-action" type="button" data-community-follow>关注子站</button>
        <span class="action-message" data-community-message></span>
      </div>
      <div class="community-stats">%s</div>
    </section>
    <nav class="board-tabs js-board-tabs" aria-label="子站板块">%s</nav>
    <section class="content-layout">
      <div>
        %s
        %s
        %s
        %s
        %s
      </div>
      <aside class="sidebar">
        %s
        <section><h2>热门标签</h2><div class="tag-cloud">%s</div></section>
        <section><h2>版主</h2>%s</section>
        <section class="community-rules"><h2>子站简介</h2><p>%s</p></section>
      </aside>
    </section>
  </main>
  <script>
    (() => {
      const root = document.querySelector('[data-community-hero]');
      const button = document.querySelector('[data-community-follow]');
      const message = document.querySelector('[data-community-message]');
      const id = Number(root?.dataset.communityId || 0);
      const token = () => localStorage.getItem('devhub_user_token') || localStorage.getItem('devhub_access_token') || '';
      const userStatus = document.querySelector('[data-user-status]');
      const moderatorEntry = document.querySelector('[data-moderator-entry]');
      const headers = (extra = {}) => token() ? {...extra, Authorization: 'Bearer ' + token()} : extra;
      const setMessage = (text, danger = false) => {
        if (!message) return;
        message.textContent = text || '';
        message.style.color = danger ? '#dc2626' : '';
      };
      const syncUser = () => {
        if (!token()) return;
        fetch('/api/v1/auth/me', {headers: headers()}).then((response) => response.ok ? response.json() : null).then((user) => {
          if (!user) return;
          const name = user.nickname || user.username || '已登录';
          if (userStatus) userStatus.innerHTML = '<span>' + name + '</span>';
          if (moderatorEntry) moderatorEntry.classList.toggle('is-hidden', !Boolean(user.is_moderator));
          localStorage.setItem('devhub_user', JSON.stringify(user));
        }).catch(() => {});
      };
      syncUser();
      button?.addEventListener('click', () => {
        if (!id) return;
        if (!token()) {
          setMessage('请先登录后再关注子站', true);
          return;
        }
        button.disabled = true;
        fetch('/api/v1/follows/toggle', {
          method: 'POST',
          headers: headers({'Content-Type': 'application/json'}),
          body: JSON.stringify({target_type: 'community', target_id: id}),
        }).then(async (response) => {
          const data = await response.json().catch(() => ({}));
          if (!response.ok) throw new Error(data.error || '操作失败');
          button.textContent = data.followed ? '已关注子站' : '关注子站';
          setMessage(data.followed ? '已关注子站' : '已取消关注');
        }).catch((err) => setMessage(err?.message || '操作失败', true)).finally(() => {
          button.disabled = false;
        });
      });
    })();
  </script>
</body>
</html>`,
		esc(title), esc(description), esc(comm.SEOKeywords), esc(canonicalPath), esc(title), esc(description), esc(canonicalURL), esc(themeColor), esc(frontendStylesheetHref()), jsonLD,
		pathEsc(comm.Slug), esc(comm.Name), queryEsc(comm.Slug), pathEsc(comm.Slug), esc(themeColor), comm.ID, esc(comm.Slug),
		esc(comm.Name), esc(firstNonEmpty(comm.SEOTitle, comm.Name+" 技术社区")), esc(firstNonEmpty(comm.Slogan, comm.Description)),
		pathEsc(comm.Slug), communityStatsHTML(stats), communityCategoryNavHTML(comm.Slug, categories),
		communityTopicSectionHTML("置顶内容", pinned, comm.Slug, "暂无置顶内容。"),
		communityTopicSectionHTML("精华内容", featured, comm.Slug, "暂无精华内容。"),
		communityTopicSectionHTML("最新内容", latest, comm.Slug, "还没有内容，欢迎发布第一篇。"),
		communityTopicSectionHTML("热门内容", hot, comm.Slug, "暂无热门内容。"),
		communityTopicSectionHTML("未解决问答", questions, comm.Slug, "暂无未解决问答。"),
		communityAnnouncementHTML(comm), communityTagsHTML(tags, comm.Slug), communityModeratorsHTML(moderators), esc(comm.Description))
}

func (s *Server) topicSEOPage(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.topicNotFoundHTML()))
		return
	}
	topic, err := s.svc.TopicByID(id, true)
	if err != nil || topic == nil || topic.ID == 0 {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.topicNotFoundHTML()))
		return
	}
	if topic.Status != 1 {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.topicHiddenHTML(topic)))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.renderTopicHTML(c, topic)))
}

func (s *Server) topicNotFoundHTML() string {
	return fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>内容不存在 - DevHub</title><meta name="description" content="该内容不存在或已被删除。"><link rel="canonical" href="/"><link rel="stylesheet" href="%s"></head><body><main class="article-shell"><article class="article-main"><h1>内容不存在</h1><p>该内容不存在或已被删除。</p><p><a href="/">返回 DevHub 首页</a></p></article></main></body></html>`, esc(frontendStylesheetHref()))
}

func (s *Server) tagSEOPage(c *gin.Context) {
	raw := strings.Trim(strings.TrimSpace(c.Param("tag")), "/")
	if raw == "" {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.tagNotFoundHTML()))
		return
	}
	site := strings.TrimSpace(c.Query("community_slug"))
	resolved, ok := s.resolveTagForPage(site, raw)
	if !ok || resolved.Tag.ID == 0 || resolved.Tag.Status != "enable" {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.tagNotFoundHTML()))
		return
	}
	if resolved.RedirectURL != "" {
		c.Redirect(http.StatusMovedPermanently, resolved.RedirectURL)
		return
	}
	tag := resolved.Tag
	if site == "" || site == "portal" {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.renderGlobalTagHTML(c, tag)))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.renderTagHTML(c, tag)))
}

func (s *Server) communityTagSEOPage(c *gin.Context) {
	site := strings.Trim(strings.TrimSpace(c.Param("site")), "/")
	raw := strings.Trim(strings.TrimSpace(c.Param("tag")), "/")
	if site == "" || raw == "" {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.tagNotFoundHTML()))
		return
	}
	comm, ok := s.svc.CommunityBySlug(site)
	if !ok || comm.Slug == "" || comm.Status != 1 {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.tagNotFoundHTML()))
		return
	}
	resolved, ok := s.resolveTagForPage(comm.Slug, raw)
	if !ok || resolved.Tag.ID == 0 || resolved.Tag.Status != "enable" {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(s.tagNotFoundHTML()))
		return
	}
	if resolved.RedirectURL != "" {
		c.Redirect(http.StatusMovedPermanently, resolved.RedirectURL)
		return
	}
	tag := resolved.Tag
	tag.CommunityID = comm.ID
	tag.CommunitySlug = comm.Slug
	tag.CommunityName = comm.Name
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(s.renderTagHTML(c, tag)))
}

func (s *Server) tagBySlugForPage(site, raw string) (domain.Tag, bool) {
	resolved, ok := s.resolveTagForPage(site, raw)
	if !ok {
		return domain.Tag{}, false
	}
	return resolved.Tag, true
}

func (s *Server) resolveTagForPage(site, raw string) (domain.TagResolveResult, bool) {
	resolved, ok := s.svc.ResolveTag(site, raw)
	if ok {
		segment := tagPathSegment(firstNonEmpty(resolved.Tag.Slug, resolved.Tag.Name))
		canonical := tagHrefFromSegment(segment, "")
		if site != "" && site != "portal" {
			canonical = tagHrefFromSegment(segment, site)
		}
		resolved.RedirectURL = canonical
		if resolved.ResolvedBy == "direct" && strings.EqualFold(strings.Trim(raw, "/"), segment) {
			resolved.RedirectURL = ""
		}
		return resolved, true
	}
	if decoded, err := url.PathUnescape(raw); err == nil && decoded != raw {
		return s.resolveTagForPage(site, decoded)
	}
	return domain.TagResolveResult{}, false
}

func (s *Server) tagNotFoundHTML() string {
	return fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>标签不存在 - DevHub</title><meta name="description" content="该标签不存在或已被禁用。"><meta name="robots" content="noindex,follow"><link rel="canonical" href="/"><link rel="stylesheet" href="%s"></head><body><main class="article-shell"><article class="article-main"><h1>标签不存在</h1><p>该标签不存在或已被禁用。</p><p><a href="/">返回 DevHub 首页</a></p></article></main></body></html>`, esc(frontendStylesheetHref()))
}

func (s *Server) renderGlobalTagHTML(c *gin.Context, tag domain.Tag) string {
	tag.Site = "portal"
	tag.CommunityID = 0
	tag.CommunitySlug = ""
	tag.CommunityName = ""
	slug := firstNonEmpty(tag.Slug, tagPathSegment(tag.Name))
	if slug == "" {
		slug = tagPathSegment(c.Param("tag"))
	}
	if slug != "" {
		tag.Slug = slug
	}
	topics, total := s.svc.TopicsByFilter(0, 0, "", "latest", nil, firstNonEmpty(tag.Slug, tag.Name), 1, 12)
	if total == 0 && tag.Name != "" && tag.Slug != tag.Name {
		topics, total = s.svc.TopicsByFilter(0, 0, "", "latest", nil, tag.Name, 1, 12)
	}
	tag.TopicCount = total
	hotTopics, _ := s.svc.TopicsByFilter(0, 0, "", "hot", nil, firstNonEmpty(tag.Slug, tag.Name), 1, 8)
	featuredTopics, _ := s.svc.TopicsByFilter(0, 0, "", "featured", nil, firstNonEmpty(tag.Slug, tag.Name), 1, 8)
	unsolved := false
	unsolvedTopics, _ := s.svc.TopicsByFilter(0, 0, "question", "latest", &unsolved, firstNonEmpty(tag.Slug, tag.Name), 1, 8)
	return s.renderTagHTMLWithTopics(c, tag, topics, total, hotTopics, featuredTopics, unsolvedTopics)
}

func (s *Server) renderTagHTML(c *gin.Context, tag domain.Tag) string {
	topics, total := s.svc.TagTopics(tag.ID, tag.CommunityID, "", "latest", 1, 12)
	hotTopics, _ := s.svc.TagTopics(tag.ID, tag.CommunityID, "", "hot", 1, 8)
	featuredTopics, _ := s.svc.TagTopics(tag.ID, tag.CommunityID, "", "featured", 1, 8)
	unsolvedTopics, _ := s.svc.TagTopics(tag.ID, tag.CommunityID, "", "unsolved", 1, 8)
	return s.renderTagHTMLWithTopics(c, tag, topics, total, hotTopics, featuredTopics, unsolvedTopics)
}

func (s *Server) renderTagHTMLWithTopics(c *gin.Context, tag domain.Tag, topics []domain.Topic, total int, hotTopics, featuredTopics, unsolvedTopics []domain.Topic) string {
	communityName := tag.CommunityName
	if communityName == "" && tag.CommunityID > 0 {
		if comm, ok := s.communityByID(tag.CommunityID); ok {
			communityName = comm.Name
			tag.CommunitySlug = comm.Slug
		}
	}
	isCommunityTag := tag.CommunitySlug != "" && tag.CommunitySlug != "portal"
	titleBase := firstNonEmpty(tag.SEOTitle, tag.Name+" 相关内容")
	description := firstNonEmpty(tag.SEODescription, tag.Description, "DevHub "+tag.Name+" 标签聚合，汇总相关文章、问答、项目和文档。")
	if isCommunityTag && tag.SEOTitle == "" {
		titleBase = firstNonEmpty(communityName, tag.CommunitySlug) + " " + tag.Name + " 标签"
	}
	if isCommunityTag && tag.SEODescription == "" {
		description = firstNonEmpty(tag.Description, "DevHub "+firstNonEmpty(communityName, tag.CommunitySlug)+" 子站 "+tag.Name+" 标签聚合，汇总相关文章、问答、项目和文档。")
	}
	title := titleBase + " - DevHub"
	tagSegment := tagPathSegment(firstNonEmpty(tag.Slug, tag.Name))
	canonicalPath := fmt.Sprintf("/tags/%s/", tagSegment)
	if isCommunityTag {
		canonicalPath = fmt.Sprintf("/c/%s/tags/%s/", tag.CommunitySlug, tagSegment)
	}
	canonicalURL := absoluteURL(c, canonicalPath)
	related := s.svc.TagSuggestions(tag.Site, "", 18)
	jsonLD := fmt.Sprintf(`{"@context":"https://schema.org","@type":"CollectionPage","name":%q,"description":%q,"url":%q,"publisher":{"@type":"Organization","name":"DevHub"}}`, tag.Name+" - DevHub", description, canonicalURL)
	communityLink := ""
	if tag.CommunitySlug != "" {
		communityLink = fmt.Sprintf(`<a href="/c/%s/">%s</a>`, pathEsc(tag.CommunitySlug), esc(firstNonEmpty(communityName, tag.CommunitySlug)))
	}
	heroLabel := "标签"
	if isCommunityTag {
		heroLabel = "子站标签"
	}
	publishHref := "/topics/new/"
	if isCommunityTag {
		publishHref = "/c/" + pathEsc(tag.CommunitySlug) + "/topics/new/"
	}
	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <meta name="description" content="%s">
  <meta name="keywords" content="%s">
  <link rel="canonical" href="%s">
  <meta property="og:site_name" content="DevHub">
  <meta property="og:type" content="website">
  <meta property="og:title" content="%s">
  <meta property="og:description" content="%s">
  <meta property="og:url" content="%s">
  <meta name="theme-color" content="#2563eb">
  <link rel="stylesheet" href="%s">
  <script type="application/ld+json">%s</script>
</head>
<body>
  <header class="site-header">
    <div class="top-accent"></div>
    <div class="header-inner">
      <a class="brand" href="/"><span>DH</span><strong>DevHub</strong></a>
      <nav class="site-nav" aria-label="主导航">
        <a href="/">首页</a>
        %s
        <a href="/search/?tag=%s">搜索标签</a>
        <a href="%s">发布内容</a>
      </nav>
    </div>
  </header>
  <main>
    <section class="page-title tag-hero" data-tag-id="%d" data-tag-slug="%s">
      <span>%s</span>
      <h1>%s</h1>
      <p>%s</p>
      <div class="hero-actions">
        <button class="secondary-action" type="button" data-tag-follow>关注标签</button>
        <a class="primary-link" href="%s">发布相关内容</a>
        <span class="action-message" data-tag-message></span>
      </div>
      <div class="community-stats"><span><strong>%d</strong>内容</span><span><strong>%d</strong>关注</span>%s</div>
    </section>
    <section class="content-layout">
      <div>
        %s
        %s
        %s
        %s
      </div>
      <aside class="sidebar">
        <section><h2>相关标签</h2><div class="tag-cloud">%s</div></section>
        <section class="community-rules"><h2>标签说明</h2><p>%s</p></section>
      </aside>
    </section>
  </main>
  <script>
    (() => {
      const hero = document.querySelector('[data-tag-id]');
      const button = document.querySelector('[data-tag-follow]');
      const message = document.querySelector('[data-tag-message]');
      const id = Number(hero?.dataset.tagId || 0);
      const token = () => localStorage.getItem('devhub_user_token') || localStorage.getItem('devhub_access_token') || '';
      const setMessage = (text, danger = false) => {
        if (!message) return;
        message.textContent = text || '';
        message.style.color = danger ? '#dc2626' : '';
      };
      button?.addEventListener('click', () => {
        if (!id) return;
        button.disabled = true;
        const headers = {'Content-Type': 'application/json'};
        const accessToken = token();
        if (accessToken) headers.Authorization = 'Bearer ' + accessToken;
        fetch('/api/v1/follows/toggle', {method: 'POST', headers, body: JSON.stringify({target_type: 'tag', target_id: id})})
          .then(async (response) => {
            const data = await response.json().catch(() => ({}));
            if (!response.ok) throw new Error(data.error || '操作失败');
            button.textContent = data.followed ? '已关注标签' : '关注标签';
            setMessage(data.followed ? '已关注标签' : '已取消关注');
          })
          .catch((err) => setMessage(err?.message || '操作失败', true))
          .finally(() => { button.disabled = false; });
      });
    })();
  </script>
</body>
</html>`,
		esc(title), esc(description), esc(firstNonEmpty(tag.SEOKeywords, tag.Name)), esc(canonicalURL), esc(title), esc(description), esc(canonicalURL), esc(frontendStylesheetHref()), jsonLD,
		communityLink, queryEsc(tag.Name), esc(publishHref), tag.ID, esc(tag.Slug), esc(heroLabel), esc(tag.Name), esc(description),
		esc(publishHref),
		total, tag.FollowerCount, tagCommunityStatHTML(tag),
		tagTopicSectionHTML("最新内容", topics, "暂无相关内容。"),
		tagTopicSectionHTML("热门内容", hotTopics, "暂无热门内容。"),
		tagTopicSectionHTML("精华内容", featuredTopics, "暂无精华内容。"),
		tagTopicSectionHTML("未解决问答", unsolvedTopics, "暂无未解决问答。"),
		relatedTagsHTML(related, tag.Slug), esc(firstNonEmpty(tag.Description, description)))
}

func (s *Server) topicHiddenHTML(topic *domain.Topic) string {
	title := "内容已隐藏 - DevHub"
	description := "该内容因社区治理规则已被隐藏。"
	return fmt.Sprintf(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s</title><meta name="description" content="%s"><meta name="robots" content="noindex,follow"><link rel="canonical" href="/"><link rel="stylesheet" href="%s"></head><body><main class="article-shell"><article class="article-main"><h1>内容已隐藏</h1><p>该主题已被管理员或版主隐藏。</p><p><a href="/">返回 DevHub 首页</a></p></article></main></body></html>`, esc(title), esc(description), esc(frontendStylesheetHref()))
}

func (s *Server) renderTopicHTML(c *gin.Context, topic *domain.Topic) string {
	community, category := s.topicSEOContext(topic)
	communitySlug := community.Slug
	if communitySlug == "" {
		communitySlug = "portal"
	}
	communityName := community.Name
	if communityName == "" {
		communityName = communitySlug
	}
	categoryName := category.Name
	if categoryName == "" {
		categoryName = contentTypeLabel(topic.ContentType)
	}
	description := seoDescription(topic.Summary, topic.Content)
	canonicalPath := fmt.Sprintf("/topics/%d/", topic.ID)
	canonicalURL := absoluteURL(c, canonicalPath)
	title := topic.Title + " - DevHub"
	contentHTML := paragraphsHTML(topic.Content)
	tagLinks := s.topicTagLinks(topic, communitySlug)
	stylesheetHref := frontendStylesheetHref()
	jsonLD := fmt.Sprintf(`{"@context":"https://schema.org","@type":"Article","headline":%q,"description":%q,"datePublished":%q,"dateModified":%q,"author":{"@type":"Person","name":"DevHub 用户"},"publisher":{"@type":"Organization","name":"DevHub"},"mainEntityOfPage":%q}`,
		topic.Title, description, topic.CreatedAt, firstNonEmpty(topic.UpdatedAt, topic.CreatedAt), canonicalURL)

	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <meta name="description" content="%s">
  <link rel="canonical" href="%s">
  <meta property="og:site_name" content="DevHub">
  <meta property="og:type" content="article">
  <meta property="og:title" content="%s">
  <meta property="og:description" content="%s">
  <meta property="og:url" content="%s">
  <meta name="theme-color" content="#2563eb">
  <link rel="stylesheet" href="%s">
  <script type="application/ld+json">%s</script>
</head>
<body>
  <header class="site-header">
    <div class="top-accent"></div>
    <div class="header-inner">
      <a class="brand" href="/"><span>DH</span><strong>DevHub</strong></a>
      <nav class="site-nav" aria-label="主导航">
        <a href="/">首页</a>
        <a href="/c/%s/">%s</a>
        <a href="/search/?scope=community&amp;community_slug=%s&amp;content_type=%s">%s</a>
        <a href="/me/activities">我的动态</a>
        <a href="/notifications">通知</a>
      </nav>
      <div class="header-actions">
        <a class="publish-link" href="/c/%s/topics/new/">发布内容</a>
        <a href="/me/favorites">收藏</a>
        <a href="/me/follows">关注</a>
        <a class="is-hidden" href="/moderator" data-moderator-entry>版主工作台</a>
        <span data-user-status><a href="/#login">登录</a></span>
      </div>
    </div>
  </header>
  <main class="article-shell">
    <article class="article-main" data-topic-detail data-topic-id="%d">
      <nav class="breadcrumb" aria-label="面包屑">
        <a href="/">首页</a><span>/</span><a href="/c/%s/">%s</a><span>/</span><a href="/search/?scope=community&amp;community_slug=%s&amp;content_type=%s">%s</a>
      </nav>
      <header class="article-header">
        <h1>%s</h1>
        <p>%s</p>
        <div class="article-meta">
          <span>DevHub 用户</span>
          <a href="/c/%s/">%s</a>
          <span>%s</span>
          <time datetime="%s">%s 发布</time>
          <span>%d 浏览</span>
          <span>%d 评论</span>
          <span>%d 赞</span>
          <span>%d 收藏</span>
        </div>
      </header>
      %s
      <div class="article-content topic-content">%s</div>
      <div class="article-tags">%s</div>
	      <div class="article-actions">
        <button type="button" data-topic-action="like" data-liked="false">点赞 <strong data-like-count>%d</strong></button>
        <button type="button" data-topic-action="favorite" data-favorited="false">收藏 <strong data-favorite-count>%d</strong></button>
        <button type="button" data-topic-action="follow" data-followed="false">关注主题</button>
        <a href="#comments">评论</a>
        <button type="button" data-topic-action="report">举报</button>
        <span class="action-message" data-action-message></span>
      </div>
      <dialog class="report-dialog" data-report-dialog>
        <form method="dialog" data-report-form>
          <h2>举报内容</h2>
          <input type="hidden" name="target_type" value="topic">
          <input type="hidden" name="target_id" value="%d">
          <label>举报原因
            <select name="reason_type">
              <option value="spam">广告 / 垃圾内容</option>
              <option value="abuse">攻击谩骂</option>
              <option value="illegal">违法违规</option>
              <option value="other">其他</option>
            </select>
          </label>
          <label>补充说明
            <textarea name="reason_text" maxlength="500" placeholder="可选，最多 500 字"></textarea>
          </label>
          <div class="report-actions"><button type="submit" value="submit">提交举报</button><button type="button" data-report-close>取消</button></div>
          <p class="action-message" data-report-message></p>
        </form>
      </dialog>
      <section id="comments" class="related-topics" data-comment-root>
        <div class="section-head"><h2>评论 <strong data-comment-total>%d</strong></h2><span data-comment-status>%s</span></div>
        <form class="comment-form" data-comment-form>
          <textarea name="content" minlength="2" maxlength="5000" placeholder="%s" %s></textarea>
          <button type="submit" data-comment-submit %s>发表评论</button>
          <div class="action-message" data-comment-message></div>
        </form>
        <div class="comment-list" data-topic-comments>评论将由前端运行时加载。</div>
      </section>
    </article>
  </main>
  <script>
	  (() => {
	    const id = %d;
	    const isQuestion = %t;
	    const commentLocked = %t;
	    const topicAuthorID = %d;
	    const comments = document.querySelector('[data-topic-comments]');
    const commentForm = document.querySelector('[data-comment-form]');
    const commentTextarea = commentForm?.querySelector('textarea[name="content"]');
    const commentSubmit = document.querySelector('[data-comment-submit]');
    const commentStatus = document.querySelector('[data-comment-status]');
    const commentTotal = document.querySelector('[data-comment-total]');
	    const commentMessage = document.querySelector('[data-comment-message]');
	    const message = document.querySelector('[data-action-message]');
	    const likeButton = document.querySelector('[data-topic-action="like"]');
	    const favoriteButton = document.querySelector('[data-topic-action="favorite"]');
	    const followButton = document.querySelector('[data-topic-action="follow"]');
	    const reportDialog = document.querySelector('[data-report-dialog]');
	    const reportForm = document.querySelector('[data-report-form]');
	    const reportMessage = document.querySelector('[data-report-message]');
	    let commentPage = 1;
	    let commentHasMore = false;
	    let commentLoading = false;
	    const accessToken = () => localStorage.getItem('devhub_user_token') || localStorage.getItem('devhub_access_token') || '';
	    const userStatus = document.querySelector('[data-user-status]');
	    const moderatorEntry = document.querySelector('[data-moderator-entry]');
	    const currentUserID = () => {
	      try {
	        const raw = localStorage.getItem('devhub_user');
	        if (!raw) return 1;
	        const parsed = JSON.parse(raw);
	        return Number(parsed.id || parsed.user_id || 1) || 1;
	      } catch (_) {
	        return 1;
	      }
	    };
    const headers = (extra = {}) => {
      const token = accessToken();
      return token ? {...extra, Authorization: 'Bearer ' + token} : extra;
    };
    const syncUser = () => {
      if (!accessToken()) return;
      fetch('/api/v1/auth/me', {headers: headers()}).then(response => response.ok ? response.json() : null).then(user => {
        if (!user) return;
        const name = user.nickname || user.username || '已登录';
        if (userStatus) userStatus.innerHTML = '<span>' + name + '</span>';
        if (moderatorEntry) moderatorEntry.classList.toggle('is-hidden', !Boolean(user.is_moderator));
        localStorage.setItem('devhub_user', JSON.stringify(user));
      }).catch(() => {});
    };
    const setNodeMessage = (node, text, danger = false) => {
      if (!node) return;
      node.textContent = text || '';
      node.style.color = danger ? '#dc2626' : '';
    };
    const setMessage = (text, danger = false) => setNodeMessage(message, text, danger);
    const setCommentMessage = (text, danger = false) => setNodeMessage(commentMessage, text, danger);
    const renderState = (state = {}) => {
      if (likeButton) {
        const liked = Boolean(state.liked);
        likeButton.dataset.liked = String(liked);
        likeButton.classList.toggle('active', liked);
        likeButton.innerHTML = (liked ? '已赞 ' : '点赞 ') + '<strong data-like-count>' + Number(state.like_count || 0) + '</strong>';
      }
      if (favoriteButton) {
        const favorited = Boolean(state.favorited);
        favoriteButton.dataset.favorited = String(favorited);
        favoriteButton.classList.toggle('active', favorited);
        favoriteButton.innerHTML = (favorited ? '已收藏 ' : '收藏 ') + '<strong data-favorite-count>' + Number(state.favorite_count || 0) + '</strong>';
      }
      if (followButton) {
        const followed = Boolean(state.followed);
        followButton.dataset.followed = String(followed);
        followButton.classList.toggle('active', followed);
        followButton.textContent = followed ? '已关注主题' : '关注主题';
      }
    };
    syncUser();
    fetch('/api/v1/topics/' + id + '/interaction', { headers: headers() }).then(r => r.ok ? r.json() : null).then(data => {
	      if (data) renderState(data);
	    }).catch(() => {});
	    if (commentLocked) {
	      setCommentMessage('评论已锁定');
	      if (commentStatus) commentStatus.textContent = '评论已锁定';
	    }
	    loadComments();
    commentForm?.addEventListener('submit', event => {
      event.preventDefault();
      if (commentLocked) {
        setCommentMessage('评论已锁定', true);
        return;
      }
      const content = String(commentTextarea?.value || '').trim();
      if (content.length < 2) {
        setCommentMessage('评论内容至少 2 个字符', true);
        return;
      }
      setCommentBusy(true);
      postJSON('/api/v1/topics/' + id + '/comments', {content})
        .then(data => {
          commentTextarea.value = '';
          setCommentMessage('评论已发布');
          if (data.topic) updateTopicCounts(data.topic);
	          return reloadComments();
        })
        .catch(err => setCommentMessage(err?.message || '评论发布失败，请稍后再试', true))
        .finally(() => setCommentBusy(false));
    });
    comments?.addEventListener('click', event => {
	      const replyButton = event.target.closest('[data-reply]');
	      const acceptButton = event.target.closest('[data-accept]');
	      const reportButton = event.target.closest('[data-report-comment]');
	      if (replyButton) {
        if (commentLocked) {
          setCommentMessage('评论已锁定', true);
          return;
        }
        const idValue = Number(replyButton.dataset.reply || 0);
        const form = comments.querySelector('[data-reply-form="' + idValue + '"]');
        form?.classList.toggle('is-open');
      }
      if (acceptButton) {
        const idValue = Number(acceptButton.dataset.accept || 0);
        acceptButton.disabled = true;
        postJSON('/api/v1/topics/' + id + '/comments/' + idValue + '/accept')
          .then(data => {
            setCommentMessage('已采纳最佳答案');
            if (data.topic) updateTopicCounts(data.topic);
	            return reloadComments();
          })
          .catch(err => setCommentMessage(err?.message || '采纳失败', true))
	          .finally(() => { acceptButton.disabled = false; });
	      }
	      if (reportButton) {
	        openReport('comment', Number(reportButton.dataset.reportComment || 0));
	      }
	    });
    comments?.addEventListener('submit', event => {
      const form = event.target.closest('[data-reply-form]');
      if (!form) return;
      event.preventDefault();
      const parentID = Number(form.dataset.replyForm || 0);
      const textarea = form.querySelector('textarea');
      const button = form.querySelector('button');
      const content = String(textarea?.value || '').trim();
      if (content.length < 2) {
        setCommentMessage('回复内容至少 2 个字符', true);
        return;
      }
      button.disabled = true;
      postJSON('/api/v1/topics/' + id + '/comments/' + parentID + '/replies', {content})
        .then(data => {
          textarea.value = '';
          form.classList.remove('is-open');
          setCommentMessage('回复已发布');
          if (data.topic) updateTopicCounts(data.topic);
	          return reloadComments();
        })
        .catch(err => setCommentMessage(err?.message || '回复失败，请稍后再试', true))
        .finally(() => { button.disabled = false; });
    });
    likeButton?.addEventListener('click', () => postJSON('/api/v1/topics/' + id + '/like').then(renderState).then(() => setMessage('点赞状态已更新')).catch(errorMessage));
    favoriteButton?.addEventListener('click', () => postJSON('/api/v1/topics/' + id + '/favorite').then(renderState).then(() => setMessage('收藏状态已更新')).catch(errorMessage));
    followButton?.addEventListener('click', () => postJSON('/api/v1/follows/toggle', {target_type: 'topic', target_id: id}).then(data => {
      renderState({liked: likeButton?.dataset.liked === 'true', favorited: favoriteButton?.dataset.favorited === 'true', followed: data.followed, like_count: Number(document.querySelector('[data-like-count]')?.textContent || 0), favorite_count: Number(document.querySelector('[data-favorite-count]')?.textContent || 0)});
      setMessage(data.followed ? '已关注主题' : '已取消关注主题');
    }).catch(errorMessage));
	    document.querySelector('[data-topic-action="report"]')?.addEventListener('click', () => openReport('topic', id));
	    reportForm?.addEventListener('submit', event => {
	      event.preventDefault();
	      const payload = {
	        target_type: reportForm.target_type.value,
	        target_id: Number(reportForm.target_id.value || 0),
	        reason_type: reportForm.reason_type.value,
	        reason_text: reportForm.reason_text.value,
	      };
	      setNodeMessage(reportMessage, '提交中');
	      postJSON('/api/v1/reports', payload)
	        .then(() => {
	          setNodeMessage(reportMessage, '举报已提交');
	          setMessage('举报已提交');
	          setTimeout(() => reportDialog?.close(), 500);
	        })
	        .catch(err => setNodeMessage(reportMessage, err?.message || '举报失败', true));
	    });
	    document.querySelector('[data-report-close]')?.addEventListener('click', () => reportDialog?.close());
	    comments?.addEventListener('click', event => {
	      const moreButton = event.target.closest('[data-load-more-comments]');
	      if (!moreButton || commentLoading) return;
	      moreButton.disabled = true;
	      loadComments(commentPage + 1, true).finally(() => { moreButton.disabled = false; });
	    });
    function postJSON(url, body) {
      return fetch(url, {
        method: 'POST',
        headers: headers({'Content-Type': 'application/json'}),
        body: body ? JSON.stringify(body) : undefined,
      }).then(async r => {
        const data = await r.json().catch(() => ({}));
        if (!r.ok) throw new Error(data.error || '操作失败，请稍后再试');
        return data;
      });
    }
	    function reloadComments() {
	      commentPage = 1;
	      return loadComments(1, false);
	    }
	    function loadComments(page = 1, append = false) {
	      if (!comments) return Promise.resolve();
	      commentLoading = true;
	      if (commentStatus) commentStatus.textContent = '加载中';
	      return fetch('/api/v1/topics/' + id + '/comments?sort=best&page=' + page + '&page_size=20', { headers: headers() })
	        .then(r => r.ok ? r.json() : Promise.reject(new Error('评论加载失败')))
	        .then(data => {
	          const items = data.items || [];
	          if (commentTotal) commentTotal.textContent = Number(data.total || items.length);
	          commentPage = Number(data.page || page);
	          commentHasMore = Boolean(data.has_more);
	          if (commentStatus) commentStatus.textContent = commentLocked ? '评论已锁定' : (commentHasMore ? '可继续加载' : '已加载');
	          const html = items.map(renderComment).join('');
	          if (append) {
	            const oldMore = comments.querySelector('[data-load-more-comments]');
	            oldMore?.remove();
	            comments.insertAdjacentHTML('beforeend', html + renderMoreButton());
	          } else {
	            comments.innerHTML = items.length ? html + renderMoreButton() : '<div class="empty-state">还没有评论，来写下第一条回答。</div>';
	          }
	        })
	        .catch(() => {
	          if (commentStatus) commentStatus.textContent = '加载失败';
	          comments.innerHTML = '<div class="empty-state">评论暂时加载失败，请稍后刷新。</div>';
	        })
	        .finally(() => { commentLoading = false; });
	    }
	    function renderMoreButton() {
	      return commentHasMore ? '<div class="comment-more"><button type="button" data-load-more-comments>加载更多评论</button></div>' : '';
	    }
    function renderComment(item) {
      const replies = Array.isArray(item.replies) && item.replies.length ? '<div class="comment-replies">' + item.replies.map(renderComment).join('') + '</div>' : '';
      const best = item.is_best ? '<span class="state-pill solved">最佳答案</span>' : '';
	      const canAccept = isQuestion && !item.is_best && currentUserID() === Number(topicAuthorID || 1);
	      const accept = canAccept ? '<button type="button" data-accept="' + Number(item.id || 0) + '">采纳</button>' : '';
	      const report = '<button type="button" data-report-comment="' + Number(item.id || 0) + '">举报</button>';
	      const reply = commentLocked ? '' : '<button type="button" data-reply="' + Number(item.id || 0) + '">回复</button>';
      return '<article class="comment-item" id="comment-' + Number(item.id || 0) + '">' +
        '<div class="comment-meta"><strong>' + escapeHTML(item.user_name || item.author || 'DevHub 用户') + '</strong>' + best + '<span>' + escapeHTML(item.created_at || '') + '</span></div>' +
        '<p>' + escapeHTML(item.content || item.text || '') + '</p>' +
        '<div class="comment-actions">' + reply + accept + report + '</div>' +
        '<form class="reply-form" data-reply-form="' + Number(item.id || 0) + '"><textarea maxlength="5000" placeholder="回复这条评论"></textarea><button type="submit">发布回复</button></form>' +
        replies +
      '</article>';
    }
    function setCommentBusy(busy) {
      if (commentSubmit) commentSubmit.disabled = busy || commentLocked;
      if (commentStatus) commentStatus.textContent = commentLocked ? '评论已锁定' : (busy ? '提交中' : '运行时加载');
    }
    function updateTopicCounts(topic) {
      if (commentTotal && typeof topic.comment_count === 'number') commentTotal.textContent = topic.comment_count;
      const meta = document.querySelector('.article-meta');
      if (meta && typeof topic.comment_count === 'number') {
        const spans = meta.querySelectorAll('span');
        if (spans[3]) spans[3].textContent = topic.comment_count + ' 评论';
      }
    }
    function errorMessage(err) { setMessage(err?.message || '操作失败，请稍后再试', true); }
    function openReport(targetType, targetID) {
      if (!reportDialog || !reportForm) return;
      reportForm.target_type.value = targetType;
      reportForm.target_id.value = String(targetID);
      reportForm.reason_type.value = 'spam';
      reportForm.reason_text.value = '';
      setNodeMessage(reportMessage, '');
      if (typeof reportDialog.showModal === 'function') reportDialog.showModal();
      else {
        const reason = window.prompt('请输入举报原因：spam / abuse / illegal / other', 'spam') || 'spam';
        postJSON('/api/v1/reports', {target_type: targetType, target_id: targetID, reason_type: reason}).then(() => setMessage('举报已提交')).catch(errorMessage);
      }
    }
    function escapeHTML(value) { return String(value || '').replace(/[&<>"']/g, char => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char])); }
  })();
  </script>
</body>
</html>`,
		esc(title), esc(description), esc(canonicalPath), esc(title), esc(description), esc(canonicalURL), esc(stylesheetHref), jsonLD,
		pathEsc(communitySlug), esc(communityName), queryEsc(communitySlug), queryEsc(topic.ContentType), esc(categoryName),
		pathEsc(communitySlug), topic.ID,
		pathEsc(communitySlug), esc(communityName), queryEsc(communitySlug), queryEsc(topic.ContentType), esc(categoryName),
		esc(topic.Title), esc(description), pathEsc(communitySlug), esc(communityName), esc(categoryName), esc(topic.CreatedAt), esc(topic.CreatedAt),
		topic.ViewCount, topic.CommentCount, topic.LikeCount, topic.FavoriteCount,
		aiSummaryHTML(topic.AISummary), contentHTML, tagLinks, topic.LikeCount, topic.FavoriteCount,
		topic.ID, topic.CommentCount, ternary(topic.CommentLocked, "评论已锁定", "运行时加载"),
		ternary(topic.CommentLocked, "评论已锁定", "写下你的评论或回答"), ternary(topic.CommentLocked, "disabled", ""), ternary(topic.CommentLocked, "disabled", ""),
		topic.ID, topic.ContentType == "question", topic.CommentLocked, topic.UserID)
}

func (s *Server) topicSEOContext(topic *domain.Topic) (domain.Community, domain.Category) {
	var community domain.Community
	for _, item := range s.svc.Communities() {
		if item.ID == topic.CommunityID {
			community = item
			break
		}
	}
	var category domain.Category
	for _, item := range s.svc.Categories(topic.CommunityID) {
		if item.ID == topic.CategoryID {
			category = item
			break
		}
	}
	return community, category
}

func (s *Server) topicTagLinks(topic *domain.Topic, communitySlug string) string {
	if len(topic.Tags) == 0 {
		return ""
	}
	links := make([]string, 0, len(topic.Tags))
	for _, tag := range topic.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		links = append(links, fmt.Sprintf(`<a href="%s">%s</a>`, esc(tagHref(tag, communitySlug)), esc(tag)))
	}
	return strings.Join(links, "")
}

func (s *Server) robots(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("User-agent: *\nAllow: /\n\nSitemap: "+absoluteURL(c, "/sitemap.xml")+"\n"))
}

func (s *Server) sitemap(c *gin.Context) {
	urls := []string{"/"}
	for _, comm := range s.svc.Communities() {
		if comm.Status == 1 && comm.Slug != "" {
			urls = append(urls, "/c/"+comm.Slug+"/")
		}
	}
	topics, _ := s.svc.TopicsByFilter(0, 0, "", "latest", nil, "", 1, 5000)
	for _, topic := range topics {
		if topic.Status == 1 {
			urls = append(urls, fmt.Sprintf("/topics/%d/", topic.ID))
		}
	}
	seenTags := map[string]bool{}
	seenTagPages := map[string]bool{}
	for _, tag := range s.svc.AdminTags("", "", "enable") {
		if tag.Status != "enable" {
			continue
		}
		segment := tagPathSegment(firstNonEmpty(tag.Slug, tag.Name))
		if segment == "" {
			continue
		}
		if !seenTags[segment] {
			seenTags[segment] = true
			urls = append(urls, "/tags/"+segment+"/")
		}
		if tag.CommunitySlug != "" && tag.CommunitySlug != "portal" {
			page := "/c/" + tag.CommunitySlug + "/tags/" + segment + "/"
			if !seenTagPages[page] {
				seenTagPages[page] = true
				urls = append(urls, page)
			}
		}
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, path := range urls {
		b.WriteString("  <url><loc>")
		b.WriteString(esc(absoluteURL(c, path)))
		b.WriteString("</loc></url>\n")
	}
	b.WriteString("</urlset>\n")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(b.String()))
}

// health 返回服务健康状态。
func (s *Server) health(c *gin.Context) {
	status := s.svc.Health()
	if !status.OK {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (s *Server) listSites(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"items": s.svc.ListSites()}) }

func (s *Server) getSite(c *gin.Context) {
	site, ok := s.svc.GetSite(c.Param("site"))
	if !ok {
		fail(c, http.StatusNotFound, "子网站不存在")
		return
	}
	c.JSON(http.StatusOK, site)
}

func (s *Server) siteOverview(c *gin.Context) {
	site := c.Param("site")
	if !s.svc.ValidateSite(site) || site == "" {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	overview, ok := s.svc.SiteOverview(site, intQuery(c, "limit", 6))
	if !ok {
		fail(c, http.StatusNotFound, "子网站不存在")
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (s *Server) listBoards(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": s.svc.ListBoards()})
}

func (s *Server) plugins(c *gin.Context) {
	items := []domain.Plugin{}
	for _, plugin := range s.svc.Plugins() {
		if plugin.Status == pluginregistry.StatusEnabled {
			plugin.ConfigJSON = ""
			plugin.ResolvedConfig = nil
			items = append(items, plugin)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) stats(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	if !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	c.JSON(http.StatusOK, s.svc.PostStats(site))
}

func (s *Server) tags(c *gin.Context) {
	site := firstQuery(c, "community_slug", "site")
	if site == "" {
		site = "portal"
	}
	if site != "portal" && !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	status := strings.TrimSpace(c.Query("status"))
	if status != "" && status != "enable" && status != "all" {
		fail(c, http.StatusBadRequest, "标签状态不合法")
		return
	}
	items := s.svc.TagStats(site)
	if q := strings.ToLower(strings.TrimSpace(c.Query("q"))); q != "" || status == "all" {
		adminStatus := "enable"
		if status == "all" {
			adminStatus = "all"
		}
		tags := s.svc.AdminTags(site, q, adminStatus)
		items = make([]domain.TagStat, 0, len(tags))
		for _, tag := range tags {
			items = append(items, tagToStat(tag))
		}
	}
	limit := intQuery(c, "limit", 0)
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) tagSuggestions(c *gin.Context) {
	site := firstQuery(c, "community_slug", "site")
	if site != "" && site != "portal" && !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	limit := intQuery(c, "limit", 20)
	if limit > 50 {
		limit = 50
	}
	c.JSON(http.StatusOK, gin.H{"items": s.svc.TagSuggestions(site, c.Query("q"), limit)})
}

func (s *Server) getTag(c *gin.Context) {
	site := firstQuery(c, "community_slug", "site")
	if site != "" && site != "portal" && !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	resolved, ok := s.svc.ResolveTag(site, c.Param("tag"))
	if !ok || resolved.Tag.Status != "enable" {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	c.JSON(http.StatusOK, resolved.Tag)
}

func (s *Server) getCommunityTag(c *gin.Context) {
	site := strings.TrimSpace(c.Param("slug"))
	if site == "" || !s.svc.ValidateSite(site) {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	resolved, ok := s.svc.ResolveTag(site, c.Param("tag"))
	if !ok || resolved.Tag.Status != "enable" {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	c.JSON(http.StatusOK, resolved.Tag)
}

func (s *Server) tagTopics(c *gin.Context) {
	site := firstQuery(c, "community_slug", "site")
	if site != "" && site != "portal" && !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	resolved, ok := s.svc.ResolveTag(site, c.Param("tag"))
	if !ok || resolved.Tag.Status != "enable" {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	tag := resolved.Tag
	page, pageSize := pagination(c)
	sortBy := c.DefaultQuery("sort", "latest")
	contentType := pluginregistry.NormalizeContentType(c.Query("content_type"))
	var topics []domain.Topic
	var total int
	if site == "" || site == "portal" {
		var isSolved *bool
		if sortBy == "unsolved" {
			unsolved := false
			isSolved = &unsolved
			contentType = "question"
		}
		topics, total = s.svc.TopicsByFilter(0, 0, contentType, sortBy, isSolved, firstNonEmpty(tag.Slug, tag.Name), page, pageSize)
	} else {
		topics, total = s.svc.TagTopics(tag.ID, tag.CommunityID, contentType, sortBy, page, pageSize)
	}
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    topics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  page*pageSize < total,
		Filters:  gin.H{"tag": tag.Slug, "community_slug": tag.CommunitySlug, "content_type": contentType, "sort": sortBy},
	})
}

func (s *Server) communityTagTopics(c *gin.Context) {
	site := strings.TrimSpace(c.Param("slug"))
	if site == "" || !s.svc.ValidateSite(site) {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	resolved, ok := s.svc.ResolveTag(site, c.Param("tag"))
	if !ok || resolved.Tag.Status != "enable" {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	tag := resolved.Tag
	page, pageSize := pagination(c)
	sortBy := c.DefaultQuery("sort", "latest")
	contentType := pluginregistry.NormalizeContentType(c.Query("content_type"))
	topics, total := s.svc.TagTopics(tag.ID, tag.CommunityID, contentType, sortBy, page, pageSize)
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    topics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  page*pageSize < total,
		Filters:  gin.H{"tag": tag.Slug, "community_slug": tag.CommunitySlug, "content_type": contentType, "sort": sortBy},
	})
}

func (s *Server) hotTags(c *gin.Context) {
	limit := intQuery(c, "limit", 20)
	if limit > 50 {
		limit = 50
	}
	type hotTag struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
		CommunityID   int64  `json:"community_id"`
		CommunitySlug string `json:"community_slug"`
		TopicCount    int    `json:"topic_count"`
		Count         int    `json:"count"`
	}
	items := []hotTag{}
	for _, comm := range s.svc.Communities() {
		tags := s.svc.TagStats(comm.Slug)
		if len(tags) == 0 {
			topics, _ := s.svc.TopicsByFilter(comm.ID, 0, "", "hot", nil, "", 1, 500)
			counts := map[string]int{}
			for _, topic := range topics {
				for _, tag := range topic.Tags {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						counts[tag]++
					}
				}
			}
			for name, count := range counts {
				tags = append(tags, domain.TagStat{Name: name, Count: count})
			}
		}
		for index, tag := range tags {
			items = append(items, hotTag{
				ID:            int64(index + 1),
				Name:          tag.Name,
				Slug:          strings.ToLower(strings.Join(strings.Fields(tag.Name), "-")),
				CommunityID:   comm.ID,
				CommunitySlug: comm.Slug,
				TopicCount:    tag.Count,
				Count:         tag.Count,
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].TopicCount == items[j].TopicCount {
			return items[i].Name < items[j].Name
		}
		return items[i].TopicCount > items[j].TopicCount
	})
	if len(items) > limit {
		items = items[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// listPosts 处理帖子列表查询，并统一返回分页结构。
func (s *Server) listPosts(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	board := c.DefaultQuery("board", "all")
	q := c.Query("q")
	tag := c.Query("tag")
	if !s.svc.ValidateSite(site) || !s.svc.ValidateBoard(board) {
		fail(c, http.StatusBadRequest, "筛选参数不合法")
		return
	}
	posts := s.svc.ListPosts(site, board, q, tag)
	page, pageSize := pagination(c)
	c.JSON(http.StatusOK, domain.PageResponse{Items: paginate(posts, page, pageSize), Total: len(posts), Page: page, PageSize: pageSize})
}

func (s *Server) getPost(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	p, exists := s.svc.GetPost(id, true)
	if !exists {
		fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	c.JSON(http.StatusOK, p)
}

func (s *Server) createPost(c *gin.Context) {
	// Legacy posts write APIs are deprecated. Use /api/v1/topics instead.
	fail(c, http.StatusGone, "posts 写接口已废弃，请使用 /api/v1/topics")
}

func (s *Server) updatePost(c *gin.Context) {
	fail(c, http.StatusGone, "posts 写接口已废弃，请使用 /api/v1/topics")
}

func (s *Server) deletePost(c *gin.Context) {
	fail(c, http.StatusGone, "posts 写接口已废弃，请使用 /api/v1/topics")
}

func (s *Server) likePost(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	p, err := s.svc.LikePost(id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, p)
}

func (s *Server) comments(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if _, exists := s.svc.GetPost(id, false); !exists {
		fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": s.svc.CommentsTree(id)})
}

func (s *Server) createComment(c *gin.Context) {
	postID, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	cmt, err := s.svc.CreateCommentWithRequest(postID, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, cmt)
}

func (s *Server) likeComment(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	cmt, err := s.svc.LikeComment(id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, cmt)
}

func (s *Server) deleteComment(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	author := c.Query("author")
	if author == "" {
		var req struct {
			Author string `json:"author"`
		}
		_ = c.ShouldBindJSON(&req)
		author = req.Author
	}
	if err := s.svc.DeleteOwnComment(id, author); err != nil {
		status := http.StatusForbidden
		if err.Error() == "评论不存在" {
			status = http.StatusNotFound
		}
		if err.Error() == "缺少评论作者" {
			status = http.StatusBadRequest
		}
		fail(c, status, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) deleteAdminComment(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	comment, err := s.svc.CommentByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	topic, err := s.svc.TopicByID(comment.TopicID, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	if !s.svc.DeleteComment(id) {
		fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	s.audit(c, "audit", "删除评论", fmt.Sprintf("comments#%d", id))
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// search 处理全文搜索，附带当前筛选条件下的板块计数。
func (s *Server) search(c *gin.Context) {
	if c.Query("community_slug") != "" || c.Query("content_type") != "" || c.Query("category_slug") != "" || c.Query("q") != "" || c.Query("keyword") != "" {
		s.searchTopics(c)
		return
	}
	scope := c.DefaultQuery("scope", "portal")
	keyword := c.Query("keyword")
	board := c.DefaultQuery("board", "all")
	if !s.svc.ValidateSite(scope) || !s.svc.ValidateBoard(board) {
		fail(c, http.StatusBadRequest, "搜索参数不合法")
		return
	}
	posts := s.svc.ListPosts(scope, board, keyword, "")
	page, pageSize := pagination(c)
	c.JSON(http.StatusOK, gin.H{
		"items":        paginate(posts, page, pageSize),
		"total":        len(posts),
		"page":         page,
		"page_size":    pageSize,
		"board_counts": s.svc.BoardCounts(scope, keyword),
	})
}

func (s *Server) hot(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	if !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": s.svc.HotPosts(site, intQuery(c, "limit", 8))})
}

func (s *Server) feed(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	if !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": s.svc.Feed(site, intQuery(c, "limit", 8))})
}

func (s *Server) notifications(c *gin.Context) {
	page, pageSize := pagination(c)
	items, total, unread := s.svc.UserNotifications(currentUserID(c), readFilter(c.Query("is_read")), page, pageSize)
	c.JSON(http.StatusOK, gin.H{"items": items, "unread": unread, "unread_count": unread, "total": total, "page": page, "page_size": pageSize, "has_more": page*pageSize < total})
}

func (s *Server) unreadNotifications(c *gin.Context) {
	_, _, unread := s.svc.UserNotifications(currentUserID(c), nil, 1, 1)
	c.JSON(http.StatusOK, gin.H{"unread": unread, "unread_count": unread})
}

func (s *Server) readNotification(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.svc.ReadUserNotification(currentUserID(c), id) {
		fail(c, http.StatusNotFound, "通知不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"read": true})
}

func (s *Server) readAllNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"read": true, "updated": s.svc.ReadAllUserNotifications(currentUserID(c))})
}

func (s *Server) me(c *gin.Context) {
	if user, ok := currentUser(c); ok {
		c.JSON(http.StatusOK, user)
		return
	}
	c.JSON(http.StatusOK, s.svc.UserProfile())
}

// adminLogin 返回后台登录态。当前仍使用演示密码，但会校验用户和状态。
func (s *Server) adminLogin(c *gin.Context) {
	var req domain.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	session, err := s.svc.AdminLogin(req.Account, req.Password)
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	c.JSON(http.StatusOK, session)
}

func (s *Server) frontLogin(c *gin.Context) {
	var req domain.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	session, err := s.svc.UserLogin(req.Account, req.Password)
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	c.JSON(http.StatusOK, session)
}

func (s *Server) register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	session, err := s.svc.Register(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, session)
}

func (s *Server) refreshSession(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	session, err := s.svc.RefreshSession(req.RefreshToken)
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	c.JSON(http.StatusOK, session)
}

func (s *Server) refreshAdminSession(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	session, err := s.svc.RefreshAdminSession(req.RefreshToken)
	if err != nil {
		fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	c.JSON(http.StatusOK, session)
}

func (s *Server) logout(c *gin.Context) {
	var req domain.RefreshTokenRequest
	_ = c.ShouldBindJSON(&req)
	if err := s.svc.Logout(req.RefreshToken); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"logout": true})
}

func (s *Server) adminMe(c *gin.Context) {
	user, _ := currentUser(c)
	adminCtx, _ := currentAdminContext(c)
	c.JSON(http.StatusOK, gin.H{"user": user, "admin_context": adminCtx})
}

func (s *Server) adminPlugins(c *gin.Context) {
	items := s.svc.Plugins()
	for i := range items {
		if health, err := s.svc.PluginHealth(items[i].Code); err == nil {
			items[i].Health = &health
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) adminPluginImpact(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	impact, err := s.svc.PluginImpact(code)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, impact)
}

func (s *Server) adminPluginHooks(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if _, ok := s.svc.PluginByCode(code); !ok {
		fail(c, http.StatusNotFound, "插件不存在")
		return
	}
	stats, err := s.svc.HookStats(code)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	recent, err := s.svc.HookExecutions(code, 20)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": stats, "recent_executions": recent})
}

func (s *Server) injectFailedAdminPluginHookForTest(c *gin.Context) {
	if !pluginTestInjectionEnabled() {
		fail(c, http.StatusNotFound, "测试 Hook 注入接口未启用")
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	name, err := url.PathUnescape(strings.TrimSpace(c.Param("name")))
	if err != nil {
		name = strings.TrimSpace(c.Param("name"))
	}
	var req struct {
		Mode         string `json:"mode"`
		ErrorMessage string `json:"error_message"`
		Clear        bool   `json:"clear"`
	}
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	message := strings.TrimSpace(req.ErrorMessage)
	if req.Clear {
		message = ""
	}
	mode := pluginregistry.HookMode(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = pluginregistry.HookBlocking
	}
	if err := s.svc.SetHookFailureInjectionForTest(code, name, mode, message); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	operation := "hook_failure_test_injection"
	if req.Clear {
		operation = "hook_failure_test_injection_clear"
	}
	s.auditStructured(c, "system", "plugin.hook.test_injection", fmt.Sprintf("hooks#%s:%s", code, name),
		nil,
		gin.H{"mode": string(mode), "hook_name": name, "enabled": !req.Clear},
		gin.H{"plugin_code": code, "hook_name": name, "mode": string(mode), "operation": operation, "test_injection": true, "error": message})
	c.JSON(http.StatusOK, gin.H{"plugin_code": code, "hook_name": name, "mode": string(mode), "enabled": !req.Clear})
}

func (s *Server) adminPluginAuditLogs(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if _, ok := s.svc.PluginByCode(code); !ok {
		fail(c, http.StatusNotFound, "插件不存在")
		return
	}
	page, pageSize := pagination(c)
	target := strings.TrimSpace(c.Query("target"))
	if target == "" {
		target = code
	}
	filter := domain.AdminLogFilter{
		Site:        "portal",
		Type:        strings.TrimSpace(c.DefaultQuery("type", "all")),
		Action:      strings.TrimSpace(c.Query("action")),
		Target:      target,
		TargetType:  strings.TrimSpace(c.Query("target_type")),
		TargetID:    int64Query(c, "target_id", 0),
		PluginCode:  strings.TrimSpace(c.DefaultQuery("plugin_code", code)),
		ActorType:   strings.TrimSpace(c.Query("actor_type")),
		Actor:       strings.TrimSpace(c.Query("actor")),
		ActorID:     int64Query(c, "actor_user_id", 0),
		CommunityID: int64Query(c, "community_id", 0),
		Metadata:    strings.TrimSpace(c.Query("metadata")),
		RequestID:   strings.TrimSpace(c.Query("request_id")),
		StartTime:   strings.TrimSpace(c.Query("start_time")),
		EndTime:     strings.TrimSpace(c.Query("end_time")),
		Page:        page,
		PageSize:    pageSize,
	}
	items, total := s.svc.AdminLogsByFilter(filter)
	c.JSON(http.StatusOK, domain.PageResponse{Items: items, Total: total, Page: page, PageSize: pageSize, HasMore: page*pageSize < total})
}

func (s *Server) adminPluginMigrations(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if _, ok := s.svc.PluginByCode(code); !ok {
		fail(c, http.StatusNotFound, "插件不存在")
		return
	}
	items, err := s.svc.PluginMigrations(code)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	pending, failed, success := 0, 0, 0
	for _, item := range items {
		switch item.Status {
		case "pending", "running":
			pending++
		case "failed":
			failed++
		case "success":
			success++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"summary": gin.H{
			"plugin_code": code,
			"total":       len(items),
			"pending":     pending,
			"failed":      failed,
			"success":     success,
		},
	})
}

func (s *Server) runAdminPluginMigrations(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	executor := auditActor(c)
	items, err := s.svc.RunAllPluginMigrations(code, executor)
	if err != nil {
		s.auditStructured(c, "system", "plugin.migration.failed", fmt.Sprintf("plugins#%s/migrations", code),
			nil,
			gin.H{"status": "failed"},
			gin.H{"plugin_code": code, "operation": "plugin_migration_run", "error": err.Error()})
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "plugin.migration.run", fmt.Sprintf("plugins#%s/migrations", code),
		nil,
		gin.H{"status": "success", "count": len(items)},
		gin.H{"plugin_code": code, "operation": "plugin_migration_run", "count": len(items)})
	s.auditStructured(c, "system", "plugin.migration.success", fmt.Sprintf("plugins#%s/migrations", code),
		nil,
		gin.H{"status": "success", "count": len(items)},
		gin.H{"plugin_code": code, "operation": "plugin_migration_success", "count": len(items)})
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) retryAdminPluginMigration(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	name, err := url.PathUnescape(strings.TrimSpace(c.Param("name")))
	if err != nil {
		name = strings.TrimSpace(c.Param("name"))
	}
	executor := auditActor(c)
	item, err := s.svc.RunPluginMigration(code, name, executor)
	if err != nil {
		s.auditStructured(c, "system", "plugin.migration.failed", fmt.Sprintf("plugins#%s/migrations#%s", code, name),
			nil,
			gin.H{"status": "failed"},
			gin.H{"plugin_code": code, "migration_name": name, "operation": "plugin_migration_retry", "error": err.Error()})
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "plugin.migration.retry", fmt.Sprintf("plugins#%s/migrations#%s", code, item.MigrationName),
		nil,
		gin.H{"status": item.Status, "migration_name": item.MigrationName},
		gin.H{"plugin_code": code, "migration_name": item.MigrationName, "operation": "plugin_migration_retry"})
	s.auditStructured(c, "system", "plugin.migration.success", fmt.Sprintf("plugins#%s/migrations#%s", code, item.MigrationName),
		nil,
		gin.H{"status": item.Status, "migration_name": item.MigrationName},
		gin.H{"plugin_code": code, "migration_name": item.MigrationName, "operation": "plugin_migration_success"})
	c.JSON(http.StatusOK, item)
}

func (s *Server) injectFailedAdminPluginMigrationForTest(c *gin.Context) {
	if !pluginMigrationFailureInjectionEnabled() {
		fail(c, http.StatusNotFound, "测试迁移注入接口未启用")
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	name, err := url.PathUnescape(strings.TrimSpace(c.Param("name")))
	if err != nil {
		name = strings.TrimSpace(c.Param("name"))
	}
	var req struct {
		ErrorMessage string `json:"error_message"`
	}
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&req)
	}
	executor := auditActor(c)
	item, err := s.svc.InjectFailedPluginMigrationForTest(code, name, req.ErrorMessage, executor)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "plugin.migration.failed", fmt.Sprintf("plugins#%s/migrations#%s", code, item.MigrationName),
		nil,
		gin.H{"status": "failed", "migration_name": item.MigrationName, "error_message": item.ErrorMessage},
		gin.H{"plugin_code": code, "migration_name": item.MigrationName, "operation": "plugin_migration_test_injection", "test_injection": true, "error": item.ErrorMessage})
	c.JSON(http.StatusOK, item)
}

func pluginMigrationFailureInjectionEnabled() bool {
	return pluginTestInjectionEnabled()
}

func pluginTestInjectionEnabled() bool {
	return os.Getenv("DEVHUB_E2E_TESTING") == "1" || os.Getenv("CMS_STORE") == "memory"
}

func (s *Server) adminCommunityPluginImpact(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	impact, err := s.svc.CommunityPluginImpact(id, code)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, impact)
}

func (s *Server) adminPluginMenus(c *gin.Context) {
	ctx := s.actorContext(c)
	if user, ok := currentUser(c); ok {
		ctx.Permissions = user.Permissions
		ctx.TokenType = user.TokenType
		ctx.RoleCode = user.RoleCode
		ctx.IsAdmin = user.TokenType == "admin"
		if ctx.IsAdmin {
			ctx.AdminID = user.ID
			ctx.UserID = 0
		}
	}
	menus := s.filteredPluginMenus(ctx, 0, "admin")
	c.JSON(http.StatusOK, gin.H{"items": menus})
}

func (s *Server) enableAdminPlugin(c *gin.Context) {
	s.setAdminPluginStatus(c, pluginregistry.StatusEnabled)
}

func (s *Server) disableAdminPlugin(c *gin.Context) {
	s.setAdminPluginStatus(c, pluginregistry.StatusDisabled)
}

func (s *Server) archiveAdminPlugin(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	before, _ := s.svc.PluginByCode(code)
	impact, _ := s.svc.PluginImpact(code)
	plugin, err := s.svc.ArchivePlugin(code)
	if err != nil {
		s.auditStructured(c, "system", "plugin.archive.failed", fmt.Sprintf("plugins#%s", code),
			gin.H{"status": before.Status},
			gin.H{"status": before.Status},
			gin.H{"scope": "global", "plugin_code": code, "operation": "plugin_archive_failed", "error": err.Error(), "impact_summary": impact})
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "plugin.archived", fmt.Sprintf("plugins#%s", plugin.Code),
		gin.H{"status": before.Status},
		gin.H{"status": plugin.Status},
		gin.H{"scope": "global", "plugin_code": plugin.Code, "operation": "plugin_archive", "impact_summary": impact})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) restoreAdminPlugin(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	before, _ := s.svc.PluginByCode(code)
	plugin, err := s.svc.RestorePlugin(code)
	if err != nil {
		s.auditStructured(c, "system", "plugin.restore.failed", fmt.Sprintf("plugins#%s", code),
			gin.H{"status": before.Status},
			gin.H{"status": before.Status},
			gin.H{"scope": "global", "plugin_code": code, "operation": "plugin_restore_failed", "error": err.Error()})
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "plugin.restored", fmt.Sprintf("plugins#%s", plugin.Code),
		gin.H{"status": before.Status},
		gin.H{"status": plugin.Status},
		gin.H{"scope": "global", "plugin_code": plugin.Code, "operation": "plugin_restore"})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) setAdminPluginStatus(c *gin.Context, status string) {
	code := strings.TrimSpace(c.Param("code"))
	before, _ := s.svc.PluginByCode(code)
	plugin, err := s.svc.SetPluginStatus(code, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "更新插件状态", fmt.Sprintf("plugins#%s", plugin.Code),
		gin.H{"status": before.Status},
		gin.H{"status": plugin.Status},
		gin.H{"scope": "global", "plugin_code": plugin.Code, "operation": "plugin_status"})
	// Platform governance: emit plugin lifecycle hook.
	hookName := pluginregistry.HookAfterPluginEnabled
	if status == pluginregistry.StatusDisabled {
		hookName = pluginregistry.HookAfterPluginDisabled
	}
	ctx := s.actorContext(c)
	_ = s.svc.DispatchHook(pluginregistry.HookEvent{
		Name: hookName,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode: plugin.Code,
			ActorType:  pluginregistry.HookActorAdmin,
			ActorID:    ctx.AdminID,
			Actor:      ctx,
			Metadata:   map[string]any{"scope": "global", "status": status},
		},
	})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) updateAdminPluginConfig(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	before, _ := s.svc.PluginByCode(code)
	var req struct {
		ConfigJSON any `json:"config_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	payload := ""
	if req.ConfigJSON != nil {
		raw, err := json.Marshal(req.ConfigJSON)
		if err != nil || !json.Valid(raw) {
			fail(c, http.StatusBadRequest, "config_json 必须是合法 JSON")
			return
		}
		payload = string(raw)
	}
	plugin, err := s.svc.SetPluginConfig(code, payload)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "更新插件全局配置", fmt.Sprintf("plugins#%s", plugin.Code),
		gin.H{"config_json": jsonAuditValue(before.ConfigJSON)},
		gin.H{"config_json": jsonAuditValue(plugin.ConfigJSON)},
		gin.H{
			"scope":        "global",
			"plugin_code":  plugin.Code,
			"operation":    "plugin_config",
			"changed_keys": configChangedKeys(before.ConfigJSON, plugin.ConfigJSON),
		})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) authMe(c *gin.Context) {
	user, _ := currentUser(c)
	c.JSON(http.StatusOK, s.frontendUserPayload(user))
}

func (s *Server) frontendUserPayload(user domain.AuthUser) domain.AuthUser {
	items, total := s.svc.CommunityModerators(domain.CommunityModeratorFilter{
		UserID:       user.ID,
		Status:       "1",
		ActorIsAdmin: true,
		Page:         1,
		PageSize:     1000,
	})
	if total <= 0 || len(items) == 0 {
		user.IsModerator = false
		user.ModeratedCommunities = nil
		return user
	}
	user.IsModerator = true
	user.ModeratedCommunities = items
	return user
}

func (s *Server) adminOverview(c *gin.Context) {
	site := adminSiteScope(c)
	c.JSON(http.StatusOK, s.svc.AdminOverview(site))
}

func (s *Server) adminPosts(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	if forced, ok := scopedSite(c, site); ok {
		site = forced
	} else {
		return
	}
	board := c.DefaultQuery("board", "all")
	q := c.Query("q")
	status := c.DefaultQuery("status", "all")
	contentType := pluginregistry.NormalizeContentType(c.Query("content_type"))
	if !s.svc.ValidateSite(site) || !s.svc.ValidateBoard(board) {
		fail(c, http.StatusBadRequest, "筛选参数不合法")
		return
	}
	posts := s.svc.AdminTopics(site, board, q)
	if contentType != "" && contentType != "all" {
		filtered := make([]domain.Post, 0, len(posts))
		for _, post := range posts {
			if adminContentTypeByBoard(post.Board) == contentType {
				filtered = append(filtered, post)
			}
		}
		posts = filtered
	}
	if status != "all" && status != "" {
		filtered := make([]domain.Post, 0, len(posts))
		for _, post := range posts {
			if status == "pinned" && post.Pinned {
				filtered = append(filtered, post)
				continue
			}
			if status == "recommended" && post.Recommended {
				filtered = append(filtered, post)
				continue
			}
			if post.Status == status {
				filtered = append(filtered, post)
			}
		}
		posts = filtered
	}
	page, pageSize := pagination(c)
	c.JSON(http.StatusOK, domain.PageResponse{Items: paginate(posts, page, pageSize), Total: len(posts), Page: page, PageSize: pageSize})
}

func (s *Server) createAdminPost(c *gin.Context) {
	var req domain.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	topicReq, err := s.createPostToTopicRequest(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if user, ok := currentUser(c); ok {
		if user.TokenType == "user" {
			topicReq.UserID = user.ID
		} else {
			topicReq.UserID = currentDemoUserID()
		}
		topicReq.ActorPermissions = user.Permissions
		topicReq.ActorContext = s.actorContext(c)
	} else {
		topicReq.UserID = currentDemoUserID()
	}
	if !ensureSiteAllowed(c, req.Site) {
		return
	}
	if err := s.normalizeCreateTopicRequest(&topicReq); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	permission := service.CreatePermissionForContentType(topicReq.ContentType, topicReq.PluginCode)
	if permission != "" && !service.HasPermission(topicReq.ActorPermissions, permission) {
		fail(c, http.StatusForbidden, fmt.Sprintf("缺少权限 %s，不能创建该类型内容", permission))
		return
	}
	topic, err := s.svc.CreateTopic(topicReq)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Pinned || req.Recommended {
		updateReq := domain.UpdateTopicRequest{}
		if req.Pinned {
			pinned := true
			updateReq.IsPinned = &pinned
		}
		if req.Recommended {
			featured := true
			updateReq.IsFeatured = &featured
		}
		updateReq.ActorContext = s.actorContext(c)
		if updated, err := s.svc.UpdateTopic(topic.ID, updateReq); err == nil {
			topic = updated
		}
	}
	if req.Status != "" && req.Status != "publish" {
		status := postStatusToTopicStatus(req.Status)
		topic, _ = s.svc.SetTopicStatus(topic.ID, status)
	}
	s.audit(c, "operation", "后台创建主题", fmt.Sprintf("topics#%d", topic.ID))
	c.JSON(http.StatusCreated, topicToAdminPost(*topic))
}

func (s *Server) updateAdminPost(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	var req domain.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Site != nil && !ensureSiteAllowed(c, *req.Site) {
		return
	}
	updateReq, err := s.updatePostToTopicRequest(req, topic)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	updateReq.ActorContext = s.actorContext(c)
	updated, err := s.svc.UpdateTopic(id, updateReq)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "operation", "后台更新主题", fmt.Sprintf("topics#%d", id))
	c.JSON(http.StatusOK, topicToAdminPost(*updated))
}

func (s *Server) deleteAdminPost(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	if !s.svc.DeleteTopic(id) {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	s.audit(c, "operation", "后台删除主题", fmt.Sprintf("topics#%d", id))
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) createAdminSite(c *gin.Context) {
	if adminCtx, ok := currentAdminContext(c); ok && !adminCtx.IsGlobal {
		fail(c, http.StatusForbidden, "只有全局管理员可以新增子站")
		return
	}
	var req domain.Site
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	site, err := s.svc.CreateSite(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "新增子站", fmt.Sprintf("sites#%s", site.Key))
	c.JSON(http.StatusCreated, site)
}

func (s *Server) updateAdminSite(c *gin.Context) {
	key := c.Param("site")
	if !ensureSiteAllowed(c, key) {
		return
	}
	var req domain.Site
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	site, ok := s.svc.UpdateSite(key, req)
	if !ok {
		fail(c, http.StatusNotFound, "子网站不存在")
		return
	}
	s.audit(c, "system", "更新子站配置", fmt.Sprintf("sites#%s", key))
	c.JSON(http.StatusOK, site)
}

func (s *Server) createAdminBoard(c *gin.Context) {
	var req domain.Board
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Site != "" && req.Site != "all" && !ensureSiteAllowed(c, req.Site) {
		return
	}
	board, err := s.svc.CreateBoard(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "新增板块", fmt.Sprintf("boards#%s", board.Key))
	c.JSON(http.StatusCreated, board)
}

func (s *Server) updateAdminBoard(c *gin.Context) {
	key := c.Param("board")
	var req domain.Board
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	board, ok := s.svc.UpdateBoard(key, req)
	if !ok {
		fail(c, http.StatusNotFound, "板块不存在")
		return
	}
	s.audit(c, "system", "更新板块配置", fmt.Sprintf("boards#%s", key))
	c.JSON(http.StatusOK, board)
}

func (s *Server) adminCommunities(c *gin.Context) {
	items := s.svc.Communities()
	if !isAdminUser(c) {
		scope := adminSiteScope(c)
		filtered := make([]domain.Community, 0, len(items))
		for _, item := range items {
			if item.Slug == scope || s.svc.IsCommunityModerator(currentUserID(c), item.ID) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (s *Server) adminCommunity(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	comm, ok := s.communityByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	if !s.canManageCommunityConfig(c, comm.ID) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"community": comm, "item": comm, "stats": s.svc.CommunityStats(comm.ID)})
}

func (s *Server) createAdminCommunity(c *gin.Context) {
	if !isAdminUser(c) {
		fail(c, http.StatusForbidden, "只有管理员可以新增子站")
		return
	}
	var req domain.CommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	comm, err := s.svc.CreateCommunity(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "新增子站", fmt.Sprintf("communities#%d", comm.ID))
	c.JSON(http.StatusCreated, comm)
}

func (s *Server) updateAdminCommunity(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	var req domain.CommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	comm, err := s.svc.UpdateCommunity(id, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "更新子站配置", fmt.Sprintf("communities#%d", id))
	c.JSON(http.StatusOK, comm)
}

func (s *Server) enableAdminCommunity(c *gin.Context) {
	s.setAdminCommunityStatus(c, 1)
}

func (s *Server) disableAdminCommunity(c *gin.Context) {
	s.setAdminCommunityStatus(c, 0)
}

func (s *Server) setAdminCommunityStatus(c *gin.Context, status int) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	comm, err := s.svc.SetCommunityStatus(id, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "启用子站"
	if status == 0 {
		action = "禁用子站"
	}
	s.audit(c, "system", action, fmt.Sprintf("communities#%d", id))
	c.JSON(http.StatusOK, comm)
}

func (s *Server) reorderAdminCommunities(c *gin.Context) {
	if !isAdminUser(c) {
		fail(c, http.StatusForbidden, "只有管理员可以排序子站")
		return
	}
	var req domain.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	updated := s.svc.ReorderCommunities(uniqueInt64s(req.IDs))
	s.audit(c, "system", "子站排序", fmt.Sprintf("communities:%d", updated))
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (s *Server) adminCommunityPlugins(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	items, err := s.svc.CommunityPlugins(id)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (s *Server) enableAdminCommunityPlugin(c *gin.Context) {
	s.setAdminCommunityPluginStatus(c, pluginregistry.StatusEnabled)
}

func (s *Server) disableAdminCommunityPlugin(c *gin.Context) {
	s.setAdminCommunityPluginStatus(c, pluginregistry.StatusDisabled)
}

func (s *Server) setAdminCommunityPluginStatus(c *gin.Context, status string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	beforeItems, _ := s.svc.CommunityPlugins(id)
	beforeStatus := ""
	for _, item := range beforeItems {
		if item.Code == code {
			beforeStatus = item.CommunityStatus
			break
		}
	}
	plugin, err := s.svc.SetCommunityPluginStatus(id, code, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "更新子站插件状态", fmt.Sprintf("community_plugins#%d:%s", id, plugin.Code),
		gin.H{"community_status": beforeStatus},
		gin.H{"community_status": plugin.CommunityStatus},
		gin.H{"scope": "community", "community_id": id, "plugin_code": plugin.Code, "operation": "community_plugin_status"})
	// Platform governance: emit plugin lifecycle hook for community scope.
	hookName := pluginregistry.HookAfterPluginEnabled
	if status == pluginregistry.StatusDisabled {
		hookName = pluginregistry.HookAfterPluginDisabled
	}
	ctx := s.actorContext(c)
	_ = s.svc.DispatchHook(pluginregistry.HookEvent{
		Name: hookName,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  plugin.Code,
			CommunityID: id,
			ActorType:   pluginregistry.HookActorAdmin,
			ActorID:     ctx.AdminID,
			Actor:       ctx,
			Metadata:    map[string]any{"scope": "community", "community_id": id, "status": status},
		},
	})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) updateAdminCommunityPluginConfig(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	beforeItems, _ := s.svc.CommunityPlugins(id)
	beforeConfig := ""
	for _, item := range beforeItems {
		if item.Code == code {
			beforeConfig = item.ConfigJSON
			break
		}
	}
	var req struct {
		ConfigJSON any `json:"config_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	payload := ""
	if req.ConfigJSON != nil {
		raw, err := json.Marshal(req.ConfigJSON)
		if err != nil || !json.Valid(raw) {
			fail(c, http.StatusBadRequest, "config_json 必须是合法 JSON")
			return
		}
		payload = string(raw)
	}
	plugin, err := s.svc.SetCommunityPluginConfig(id, code, payload)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.auditStructured(c, "system", "更新子站插件配置", fmt.Sprintf("community_plugins#%d:%s", id, plugin.Code),
		gin.H{"config_json": jsonAuditValue(beforeConfig)},
		gin.H{"config_json": jsonAuditValue(plugin.ConfigJSON)},
		gin.H{
			"scope":        "community",
			"community_id": id,
			"plugin_code":  plugin.Code,
			"operation":    "community_plugin_config",
			"changed_keys": configChangedKeys(beforeConfig, plugin.ConfigJSON),
		})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) reorderAdminCommunityPlugins(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	var req struct {
		Codes []string `json:"codes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	beforeItems, _ := s.svc.CommunityPlugins(id)
	beforeCodes := make([]string, 0, len(beforeItems))
	for _, item := range beforeItems {
		beforeCodes = append(beforeCodes, item.Code)
	}
	updated, err := s.svc.ReorderCommunityPlugins(id, uniqueStrings(req.Codes))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	afterCodes := uniqueStrings(req.Codes)
	s.auditStructured(c, "system", "子站插件排序", fmt.Sprintf("community_plugins#%d", id),
		gin.H{"codes": beforeCodes},
		gin.H{"codes": afterCodes},
		gin.H{"scope": "community", "community_id": id, "operation": "community_plugin_sort", "updated": updated})
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (s *Server) adminCommunityCategories(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	items := s.svc.Categories(id)
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (s *Server) createAdminCommunityCategory(c *gin.Context) {
	communityID, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, communityID) {
		return
	}
	var req domain.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	cat, err := s.svc.CreateCategory(communityID, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "新增子站板块", fmt.Sprintf("categories#%d", cat.ID))
	c.JSON(http.StatusCreated, cat)
}

func (s *Server) updateAdminCategory(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	current, ok := s.categoryByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "板块不存在")
		return
	}
	if !s.canManageCommunityConfig(c, current.CommunityID) {
		return
	}
	var req domain.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	cat, err := s.svc.UpdateCategory(id, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "更新子站板块", fmt.Sprintf("categories#%d", id))
	c.JSON(http.StatusOK, cat)
}

func (s *Server) enableAdminCategory(c *gin.Context) {
	s.setAdminCategoryStatus(c, 1)
}

func (s *Server) disableAdminCategory(c *gin.Context) {
	s.setAdminCategoryStatus(c, 0)
}

func (s *Server) setAdminCategoryStatus(c *gin.Context, status int) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	current, ok := s.categoryByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "板块不存在")
		return
	}
	if !s.canManageCommunityConfig(c, current.CommunityID) {
		return
	}
	cat, err := s.svc.SetCategoryStatus(id, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "启用子站板块"
	if status == 0 {
		action = "禁用子站板块"
	}
	s.audit(c, "system", action, fmt.Sprintf("categories#%d", id))
	c.JSON(http.StatusOK, cat)
}

func (s *Server) reorderAdminCategories(c *gin.Context) {
	var req domain.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.IDs) > 0 {
		cat, ok := s.categoryByID(req.IDs[0])
		if !ok {
			fail(c, http.StatusNotFound, "板块不存在")
			return
		}
		if !s.canManageCommunityConfig(c, cat.CommunityID) {
			return
		}
	}
	updated := s.svc.ReorderCategories(uniqueInt64s(req.IDs))
	s.audit(c, "system", "子站板块排序", fmt.Sprintf("categories:%d", updated))
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (s *Server) adminUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": s.svc.AdminUsers()})
}

func (s *Server) updateAdminUserStatus(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req domain.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if !s.svc.UpdateUserStatus(id, req.Status, req.Note) {
		fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	s.audit(c, "operation", "更新用户状态", fmt.Sprintf("users#%d", id))
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

func (s *Server) adminRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": s.svc.AdminRoles()})
}

func (s *Server) adminPermissions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": s.svc.AdminPermissions()})
}

func (s *Server) adminTags(c *gin.Context) {
	site := c.Query("site")
	if forced, ok := scopedSite(c, site); ok {
		site = forced
	} else {
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": s.svc.AdminTags(site, c.Query("q"), c.DefaultQuery("status", "all"))})
}

func (s *Server) adminTag(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	tag, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if tag.Site != "" && tag.Site != "portal" && !ensureSiteAllowed(c, tag.Site) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": tag, "tag": tag})
}

func (s *Server) adminTagTopics(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	found, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if found.Site != "" && found.Site != "portal" && !ensureSiteAllowed(c, found.Site) {
		return
	}
	page, pageSize := pagination(c)
	topics, total := s.svc.AdminTagTopics(id, page, pageSize)
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    topics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  page*pageSize < total,
		Filters:  gin.H{"tag_id": id, "site": found.Site},
	})
}

func (s *Server) adminTagAliases(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	tag, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if tag.Site != "" && tag.Site != "portal" && !ensureSiteAllowed(c, tag.Site) {
		return
	}
	items, err := s.svc.TagAliases(id)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) createAdminTag(c *gin.Context) {
	var req domain.Tag
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Site != "" && req.Site != "portal" && !ensureSiteAllowed(c, req.Site) {
		return
	}
	tag, err := s.svc.CreateTag(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "operation", "新增标签", fmt.Sprintf("tags#%d", tag.ID))
	c.JSON(http.StatusCreated, tag)
}

func (s *Server) updateAdminTag(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req domain.Tag
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	tag, ok := s.svc.UpdateTag(id, req)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	s.audit(c, "operation", "更新标签", fmt.Sprintf("tags#%d", id))
	c.JSON(http.StatusOK, tag)
}

func (s *Server) createAdminTagAlias(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	tag, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if tag.Site != "" && tag.Site != "portal" && !ensureSiteAllowed(c, tag.Site) {
		return
	}
	var req domain.TagAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := s.svc.AddTagAlias(id, req.Alias)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "audit", "新增标签别名", fmt.Sprintf("tags#%d/aliases#%d", id, item.ID))
	c.JSON(http.StatusCreated, item)
}

func (s *Server) deleteAdminTagAlias(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	aliasID, ok := idParam(c, "aliasId")
	if !ok {
		return
	}
	tag, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if tag.Site != "" && tag.Site != "portal" && !ensureSiteAllowed(c, tag.Site) {
		return
	}
	if err := s.svc.DeleteTagAlias(id, aliasID); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "audit", "删除标签别名", fmt.Sprintf("tags#%d/aliases#%d", id, aliasID))
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) enableAdminTag(c *gin.Context) {
	s.setAdminTagStatus(c, "enable")
}

func (s *Server) disableAdminTag(c *gin.Context) {
	s.setAdminTagStatus(c, "disable")
}

func (s *Server) setAdminTagStatus(c *gin.Context, status string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	found, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if found.Site != "" && found.Site != "portal" && !ensureSiteAllowed(c, found.Site) {
		return
	}
	tag, ok := s.svc.SetTagStatus(id, status)
	if !ok {
		fail(c, http.StatusBadRequest, "标签状态更新失败")
		return
	}
	action := "启用标签"
	if status != "enable" {
		action = "禁用标签"
	}
	s.audit(c, "operation", action, fmt.Sprintf("tags#%d", id))
	c.JSON(http.StatusOK, tag)
}

func (s *Server) mergeAdminTag(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	source, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if source.Site != "" && source.Site != "portal" && !ensureSiteAllowed(c, source.Site) {
		return
	}
	var req domain.TagMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	target, ok := s.svc.AdminTagByID(req.TargetTagID)
	if !ok {
		fail(c, http.StatusBadRequest, "目标标签不存在")
		return
	}
	if target.Site != "" && target.Site != "portal" && !ensureSiteAllowed(c, target.Site) {
		return
	}
	tag, err := s.svc.MergeTag(id, req.TargetTagID)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	targetText := fmt.Sprintf("tags#%d->%d", id, req.TargetTagID)
	if strings.TrimSpace(req.Note) != "" {
		targetText += ":" + strings.TrimSpace(req.Note)
	}
	s.audit(c, "audit", "合并标签", targetText)
	c.JSON(http.StatusOK, tag)
}

func (s *Server) recalculateAdminTag(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	tag, ok := s.svc.AdminTagByID(id)
	if !ok {
		fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	if tag.Site != "" && tag.Site != "portal" && !ensureSiteAllowed(c, tag.Site) {
		return
	}
	recalculated, err := s.svc.RecalculateTag(id)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "audit", "重算标签统计", fmt.Sprintf("tags#%d", id))
	c.JSON(http.StatusOK, recalculated)
}

func (s *Server) recalculateAllAdminTags(c *gin.Context) {
	updated, err := s.svc.RecalculateAllTags()
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "audit", "批量重算标签统计", fmt.Sprintf("tags:%d", updated))
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (s *Server) adminComments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": s.svc.AdminComments(adminSiteScope(c))})
}

func (s *Server) updateAdminCommentStatus(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	comment, err := s.svc.CommentByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	topic, err := s.svc.TopicByID(comment.TopicID, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	var req domain.UpdateCommentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if !s.svc.UpdateCommentStatus(id, req.Status) {
		fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	s.audit(c, "audit", "更新评论状态", fmt.Sprintf("comments#%d", id))
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

func (s *Server) adminSettings(c *gin.Context) {
	c.JSON(http.StatusOK, s.svc.AdminSettings())
}

func (s *Server) updateAdminSettings(c *gin.Context) {
	var req domain.AdminSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	settings := s.svc.UpdateAdminSettings(req)
	s.audit(c, "system", "更新基础参数", "settings")
	c.JSON(http.StatusOK, settings)
}

func (s *Server) adminLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": s.svc.AdminLogs(adminSiteScope(c))})
}

func (s *Server) adminAuditLogs(c *gin.Context) {
	page, pageSize := pagination(c)
	filter := domain.AdminLogFilter{
		Site:        adminSiteScope(c),
		Type:        strings.TrimSpace(c.DefaultQuery("type", "all")),
		ActorType:   strings.TrimSpace(c.Query("actor_type")),
		Action:      strings.TrimSpace(c.Query("action")),
		Target:      strings.TrimSpace(c.Query("target")),
		TargetType:  strings.TrimSpace(c.Query("target_type")),
		TargetID:    int64Query(c, "target_id", 0),
		PluginCode:  strings.TrimSpace(c.Query("plugin_code")),
		Actor:       strings.TrimSpace(c.Query("actor")),
		ActorID:     int64Query(c, "actor_user_id", 0),
		CommunityID: int64Query(c, "community_id", 0),
		Metadata:    strings.TrimSpace(c.Query("metadata")),
		RequestID:   strings.TrimSpace(c.Query("request_id")),
		StartTime:   strings.TrimSpace(c.Query("start_time")),
		EndTime:     strings.TrimSpace(c.Query("end_time")),
		Page:        page,
		PageSize:    pageSize,
	}
	items, total := s.svc.AdminLogsByFilter(filter)
	c.JSON(http.StatusOK, domain.PageResponse{Items: items, Total: total, Page: page, PageSize: pageSize, HasMore: page*pageSize < total})
}

func (s *Server) pushAdminNotification(c *gin.Context) {
	var req domain.PushNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Scope == "" {
		req.Scope = adminSiteScope(c)
	}
	notice := s.svc.PushNotification(req)
	s.audit(c, "operation", "发送通知", fmt.Sprintf("notifications#%d", notice.ID))
	c.JSON(http.StatusCreated, notice)
}

func (s *Server) createReport(c *gin.Context) {
	var req domain.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ReporterUserID = currentUserID(c)
	report, err := s.svc.CreateReport(req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "不存在") {
			status = http.StatusNotFound
		}
		fail(c, status, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"report": report, "item": report})
}

func (s *Server) moderatorCommunities(c *gin.Context) {
	communities := s.moderatorCommunitiesForCurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"items": communities, "total": len(communities)})
}

func (s *Server) moderatorPluginMenus(c *gin.Context) {
	communityID := int64(0)
	if slug := strings.TrimSpace(firstQuery(c, "community_slug", "community", "site")); slug != "" {
		if comm, ok := s.svc.CommunityBySlug(slug); ok {
			communityID = comm.ID
		}
	} else if id := strings.TrimSpace(c.Query("community_id")); id != "" {
		if parsed, err := strconv.ParseInt(id, 10, 64); err == nil {
			communityID = parsed
		}
	}
	if communityID > 0 && !s.canModerateCommunityStrict(c, communityID) {
		return
	}
	moderated := s.moderatorCommunitiesForCurrentUser(c)
	scopes := make([]int64, 0, len(moderated))
	for _, comm := range moderated {
		scopes = append(scopes, comm.ID)
	}
	ctx := s.actorContext(c)
	ctx.IsModerator = true
	ctx.CommunityScopes = uniqueInt64s(scopes)
	menus := []domain.PluginMenu{}
	if communityID > 0 {
		menus = s.filteredPluginMenus(ctx, communityID, "moderator")
	} else {
		seen := map[string]bool{}
		for _, comm := range moderated {
			for _, menu := range s.filteredPluginMenus(ctx, comm.ID, "moderator") {
				key := firstNonEmpty(menu.Code, menu.Key, menu.Path)
				if seen[key] {
					continue
				}
				seen[key] = true
				menus = append(menus, menu)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": menus, "total": len(menus)})
}

func (s *Server) moderatorDashboard(c *gin.Context) {
	communities := s.moderatorCommunitiesForCurrentUser(c)
	communityIDs := make([]int64, 0, len(communities))
	for _, comm := range communities {
		communityIDs = append(communityIDs, comm.ID)
	}
	pendingReports := s.moderatorReportsForCommunities(communityIDs, "pending", "all", 1, 50)
	topics := s.moderatorTopicsForCommunities(communityIDs, "all", "all", "", 1, 50)
	comments := s.moderatorCommentsForCommunities(communityIDs, "all", "", 1, 50)
	allLogs := s.moderatorLogsForCommunities(c, communityIDs, "", "all", "", "", 1, 1000)
	recentLogs := allLogs
	if len(recentLogs) > 5 {
		recentLogs = recentLogs[:5]
	}
	dashboard := domain.ModeratorDashboard{
		ManagedCommunities: communities,
		PendingReportCount: len(pendingReports),
		TopicCount:         len(topics),
		CommentCount:       len(comments),
		TodayActionCount:   countTodayModeratorActions(allLogs),
		RecentReports:      limitReports(pendingReports, 5),
		RecentAuditLogs:    recentLogs,
	}
	c.JSON(http.StatusOK, dashboard)
}

func (s *Server) moderatorReports(c *gin.Context) {
	communityID, ok := s.moderatorRequestedCommunityID(c)
	if !ok {
		return
	}
	page, pageSize := pagination(c)
	communityIDs := s.moderatorCommunityIDs(c, communityID)
	items := s.moderatorReportsForCommunities(communityIDs, c.DefaultQuery("status", "all"), c.DefaultQuery("target_type", "all"), 1, 1000)
	c.JSON(http.StatusOK, domain.PageResponse{Items: paginate(items, page, pageSize), Total: len(items), Page: page, PageSize: pageSize, HasMore: page*pageSize < len(items)})
}

func (s *Server) handleModeratorReport(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	report, err := s.svc.ReportByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	if !s.canModerateCommunityStrict(c, report.CommunityID) {
		return
	}
	var req domain.HandleReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := s.svc.HandleReport(id, req.Status, req.HandleNote, currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.moderatorAudit(c, report.CommunityID, "handle_report", fmt.Sprintf("reports#%d", id))
	c.JSON(http.StatusOK, gin.H{"handled": true, "report": updated})
}

func (s *Server) moderatorTopics(c *gin.Context) {
	communityID, ok := s.moderatorRequestedCommunityID(c)
	if !ok {
		return
	}
	page, pageSize := pagination(c)
	communityIDs := s.moderatorCommunityIDs(c, communityID)
	items := s.moderatorTopicsForCommunities(communityIDs, c.DefaultQuery("status", "all"), c.DefaultQuery("content_type", "all"), strings.TrimSpace(c.Query("keyword")), 1, 1000)
	c.JSON(http.StatusOK, domain.PageResponse{Items: paginate(items, page, pageSize), Total: len(items), Page: page, PageSize: pageSize, HasMore: page*pageSize < len(items)})
}

func (s *Server) featureModeratorTopic(c *gin.Context) {
	s.setModeratorTopicFeatured(c, true)
}

func (s *Server) unfeatureModeratorTopic(c *gin.Context) {
	s.setModeratorTopicFeatured(c, false)
}

func (s *Server) pinModeratorTopic(c *gin.Context) {
	s.setModeratorTopicPinned(c, true)
}

func (s *Server) unpinModeratorTopic(c *gin.Context) {
	s.setModeratorTopicPinned(c, false)
}

func (s *Server) hideModeratorTopic(c *gin.Context) {
	s.setModeratorTopicStatus(c, 0, "hide_topic")
}

func (s *Server) restoreModeratorTopic(c *gin.Context) {
	s.setModeratorTopicStatus(c, 1, "restore_topic")
}

func (s *Server) lockModeratorTopicComments(c *gin.Context) {
	s.setModeratorTopicLock(c, true)
}

func (s *Server) unlockModeratorTopicComments(c *gin.Context) {
	s.setModeratorTopicLock(c, false)
}

func (s *Server) moderatorComments(c *gin.Context) {
	communityID, ok := s.moderatorRequestedCommunityID(c)
	if !ok {
		return
	}
	page, pageSize := pagination(c)
	communityIDs := s.moderatorCommunityIDs(c, communityID)
	items := s.moderatorCommentsForCommunities(communityIDs, c.DefaultQuery("status", "all"), strings.TrimSpace(c.Query("keyword")), 1, 1000)
	c.JSON(http.StatusOK, domain.PageResponse{Items: paginate(items, page, pageSize), Total: len(items), Page: page, PageSize: pageSize, HasMore: page*pageSize < len(items)})
}

func (s *Server) hideModeratorComment(c *gin.Context) {
	s.setModeratorCommentStatus(c, "hidden")
}

func (s *Server) restoreModeratorComment(c *gin.Context) {
	s.setModeratorCommentStatus(c, "normal")
}

func (s *Server) moderatorAuditLogs(c *gin.Context) {
	communityID, ok := s.moderatorRequestedCommunityID(c)
	if !ok {
		return
	}
	page, pageSize := pagination(c)
	communityIDs := s.moderatorCommunityIDs(c, communityID)
	items := s.moderatorLogsForCommunities(c, communityIDs, c.Query("actor_type"), c.DefaultQuery("type", "all"), c.Query("action"), c.Query("target_type"), 1, 1000)
	c.JSON(http.StatusOK, domain.PageResponse{Items: paginate(items, page, pageSize), Total: len(items), Page: page, PageSize: pageSize, HasMore: page*pageSize < len(items)})
}

func (s *Server) adminReports(c *gin.Context) {
	filter := s.reportFilter(c)
	items, total := s.svc.Reports(filter)
	c.JSON(http.StatusOK, domain.PageResponse{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize, HasMore: filter.Page*filter.PageSize < total})
}

func (s *Server) adminReport(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	report, err := s.svc.ReportByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	if !s.canModerateCommunity(c, report.CommunityID) {
		return
	}
	c.JSON(http.StatusOK, report)
}

func (s *Server) handleAdminReport(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	report, err := s.svc.ReportByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	if !s.canModerateCommunity(c, report.CommunityID) {
		return
	}
	var req domain.HandleReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := s.svc.HandleReport(id, req.Status, req.HandleNote, currentUserID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "audit", "处理举报", fmt.Sprintf("reports#%d", id))
	c.JSON(http.StatusOK, gin.H{"handled": true, "report": updated})
}

func (s *Server) batchAdminReports(c *gin.Context) {
	var req domain.BatchModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = strings.TrimSpace(req.Action)
	}
	if status != "accepted" && status != "rejected" {
		fail(c, http.StatusBadRequest, "批量处理状态不合法")
		return
	}
	result := domain.BatchModerationResponse{Action: status, Items: []domain.BatchModerationItem{}}
	for _, id := range uniqueInt64s(req.IDs) {
		item := domain.BatchModerationItem{ID: id}
		report, err := s.svc.ReportByID(id)
		if err != nil {
			item.Error = err.Error()
		} else if !s.canModerateCommunityForBatch(c, report.CommunityID) {
			item.Error = "无权管理该举报"
		} else if _, err := s.svc.HandleReport(id, status, req.HandleNote, currentUserID(c)); err != nil {
			item.Error = err.Error()
		} else {
			item.OK = true
			result.Updated++
		}
		if !item.OK {
			result.Failed++
		}
		result.Items = append(result.Items, item)
	}
	s.audit(c, "audit", "批量处理举报", fmt.Sprintf("reports:%d/%d", result.Updated, len(result.Items)))
	c.JSON(http.StatusOK, result)
}

func (s *Server) adminModerators(c *gin.Context) {
	filter := s.moderatorFilter(c)
	items, total := s.svc.CommunityModerators(filter)
	c.JSON(http.StatusOK, domain.PageResponse{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize, HasMore: filter.Page*filter.PageSize < total})
}

func (s *Server) createAdminModerator(c *gin.Context) {
	if !isAdminUser(c) {
		fail(c, http.StatusForbidden, "只有管理员可以分配版主")
		return
	}
	var req domain.CommunityModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if !s.ensureModeratorRequestScope(c, req) {
		return
	}
	moderator, err := s.svc.CreateCommunityModerator(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "新增版主", fmt.Sprintf("community_moderators#%d", moderator.ID))
	c.JSON(http.StatusCreated, moderator)
}

func (s *Server) updateAdminModerator(c *gin.Context) {
	if !isAdminUser(c) {
		fail(c, http.StatusForbidden, "只有管理员可以管理版主")
		return
	}
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req domain.CommunityModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if !s.ensureModeratorRequestScope(c, req) {
		return
	}
	moderator, err := s.svc.UpdateCommunityModerator(id, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "system", "更新版主", fmt.Sprintf("community_moderators#%d", id))
	c.JSON(http.StatusOK, moderator)
}

func (s *Server) deleteAdminModerator(c *gin.Context) {
	if !isAdminUser(c) {
		fail(c, http.StatusForbidden, "只有管理员可以停用版主")
		return
	}
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	moderator, err := s.svc.CommunityModeratorByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	if !s.canModerateCommunity(c, moderator.CommunityID) {
		return
	}
	if !s.svc.DeleteCommunityModerator(id) {
		fail(c, http.StatusNotFound, "版主不存在")
		return
	}
	s.audit(c, "system", "停用版主", fmt.Sprintf("community_moderators#%d", id))
	c.JSON(http.StatusOK, gin.H{"deleted": true, "disabled": true})
}

func (s *Server) featureAdminTopic(c *gin.Context) {
	s.toggleAdminTopicBool(c, "feature")
}

func (s *Server) pinAdminTopic(c *gin.Context) {
	s.toggleAdminTopicBool(c, "pin")
}

func (s *Server) hideAdminTopic(c *gin.Context) {
	s.setAdminTopicStatus(c, 0, "隐藏主题")
}

func (s *Server) restoreAdminTopic(c *gin.Context) {
	s.setAdminTopicStatus(c, 1, "恢复主题")
}

func (s *Server) lockAdminTopicComments(c *gin.Context) {
	s.setAdminTopicLock(c, true)
}

func (s *Server) unlockAdminTopicComments(c *gin.Context) {
	s.setAdminTopicLock(c, false)
}

func (s *Server) hideAdminComment(c *gin.Context) {
	s.setAdminCommentStatus(c, "hidden")
}

func (s *Server) restoreAdminComment(c *gin.Context) {
	s.setAdminCommentStatus(c, "normal")
}

func (s *Server) batchAdminTopics(c *gin.Context) {
	var req domain.BatchModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := strings.TrimSpace(req.Action)
	result := domain.BatchModerationResponse{Action: action, Items: []domain.BatchModerationItem{}}
	for _, id := range uniqueInt64s(req.IDs) {
		item := domain.BatchModerationItem{ID: id}
		topic, err := s.svc.TopicByID(id, false)
		if err != nil {
			item.Error = "主题不存在"
		} else if !s.canModerateTopicForBatch(c, topic) {
			item.Error = "无权管理该主题"
		} else if err := s.applyBatchTopicAction(id, topic, action); err != nil {
			item.Error = err.Error()
		} else {
			item.OK = true
			result.Updated++
			s.auditPluginContentAction(c, "批量治理主题", topic, fmt.Sprintf("batch:%s", action),
				gin.H{"action": action},
				gin.H{"ok": true},
				gin.H{"batch": true, "note": req.Note})
		}
		if !item.OK {
			result.Failed++
		}
		result.Items = append(result.Items, item)
	}
	s.audit(c, "audit", "批量治理主题", batchAuditTarget("topics", action, result.Updated, len(result.Items), req.Note))
	c.JSON(http.StatusOK, result)
}

func (s *Server) batchAdminComments(c *gin.Context) {
	var req domain.BatchModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := strings.TrimSpace(req.Action)
	result := domain.BatchModerationResponse{Action: action, Items: []domain.BatchModerationItem{}}
	for _, id := range uniqueInt64s(req.IDs) {
		item := domain.BatchModerationItem{ID: id}
		comment, err := s.svc.CommentByID(id)
		if err != nil {
			item.Error = "评论不存在"
		} else {
			topic, topicErr := s.svc.TopicByID(comment.TopicID, false)
			if topicErr != nil {
				item.Error = "主题不存在"
			} else if !s.canModerateTopicForBatch(c, topic) {
				item.Error = "无权管理该评论"
			} else if err := s.applyBatchCommentAction(id, action); err != nil {
				item.Error = err.Error()
			} else {
				item.OK = true
				result.Updated++
			}
		}
		if !item.OK {
			result.Failed++
		}
		result.Items = append(result.Items, item)
	}
	s.audit(c, "audit", "批量治理评论", batchAuditTarget("comments", action, result.Updated, len(result.Items), req.Note))
	c.JSON(http.StatusOK, result)
}

// corsMiddleware 允许本地前端跨域访问后端接口。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (s *Server) authRequired() gin.HandlerFunc {
	return s.userAuthRequired()
}

func (s *Server) userAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		user, err := s.svc.AuthUser(token)
		if err != nil {
			fail(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		c.Set("auth_user", *user)
		c.Next()
	}
}

func (s *Server) moderatorAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		user, err := s.svc.AuthUser(token)
		if err != nil {
			fail(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		moderator, ok := s.moderatorUserForAdmin(*user)
		if !ok {
			fail(c, http.StatusForbidden, "当前用户不是启用状态子站版主")
			c.Abort()
			return
		}
		c.Set("auth_user", moderator)
		c.Set("moderator_context", moderator)
		c.Next()
	}
}

func (s *Server) adminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			fail(c, http.StatusUnauthorized, "后台未登录")
			c.Abort()
			return
		}
		user, err := s.svc.AuthAdmin(token)
		if err == nil && user != nil {
			c.Set("auth_user", *user)
			c.Next()
			return
		}
		frontUser, userErr := s.svc.AuthUser(token)
		if userErr == nil && frontUser != nil {
			if moderator, ok := s.moderatorUserForAdmin(*frontUser); ok {
				c.Set("auth_user", moderator)
				c.Next()
				return
			}
		}
		fail(c, http.StatusUnauthorized, err.Error())
		c.Abort()
	}
}

func (s *Server) optionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := currentUser(c); ok {
			c.Next()
			return
		}
		token := bearerToken(c.GetHeader("Authorization"))
		if token != "" {
			if user, err := s.svc.AuthUser(token); err == nil && user != nil {
				c.Set("auth_user", *user)
			}
		}
		c.Next()
	}
}

func (s *Server) adminContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		site := c.Param("site")
		if site == "" {
			site = c.Query("site")
		}
		if site == "" {
			site = "portal"
		}
		isGlobal := user.RoleCode == "super_admin" || hasSite(user.Sites, "*")
		adminCtx := domain.AdminContext{CurrentUser: user, CurrentSite: site, IsGlobal: isGlobal, Permissions: user.Permissions}
		c.Set("admin_context", adminCtx)
		c.Next()
	}
}

func (s *Server) requirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		if !hasPermission(user.Permissions, permission) {
			fail(c, http.StatusForbidden, "无权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return header
}

func currentUser(c *gin.Context) (domain.AuthUser, bool) {
	value, ok := c.Get("auth_user")
	if !ok {
		return domain.AuthUser{}, false
	}
	user, ok := value.(domain.AuthUser)
	return user, ok
}

func currentAdminContext(c *gin.Context) (domain.AdminContext, bool) {
	value, ok := c.Get("admin_context")
	if !ok {
		return domain.AdminContext{}, false
	}
	adminCtx, ok := value.(domain.AdminContext)
	return adminCtx, ok
}

func (s *Server) audit(c *gin.Context, logType, action, target string) {
	s.auditStructured(c, logType, action, target, nil, nil, nil)
}

func (s *Server) auditStructured(c *gin.Context, logType, action, target string, oldValue, newValue, metadata any) {
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		return
	}
	actor := adminCtx.CurrentUser.Username
	if actor == "" {
		actor = adminCtx.CurrentUser.Nickname
	}
	actorType := "admin_user"
	if adminCtx.CurrentUser.TokenType == "user" {
		actorType = "moderator"
	}
	site := adminCtx.CurrentSite
	if site == "" {
		site = adminSiteScope(c)
	}
	s.svc.AppendAdminLog(domain.AdminLog{
		Site:        site,
		Type:        logType,
		Actor:       actor,
		ActorType:   actorType,
		ActorUserID: adminCtx.CurrentUser.ID,
		ActorID:     adminCtx.CurrentUser.ID,
		Role:        adminCtx.CurrentUser.RoleCode,
		Action:      action,
		Target:      target,
		OldValue:    auditJSON(oldValue),
		NewValue:    auditJSON(newValue),
		Metadata:    auditJSON(metadata),
		IP:          c.ClientIP(),
	})
}

func auditActor(c *gin.Context) string {
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		return "system"
	}
	actor := adminCtx.CurrentUser.Username
	if actor == "" {
		actor = adminCtx.CurrentUser.Nickname
	}
	if actor == "" {
		actor = fmt.Sprintf("admin#%d", adminCtx.CurrentUser.ID)
	}
	return actor
}

func auditJSON(value any) string {
	if value == nil {
		return ""
	}
	if raw, ok := value.(string); ok {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return ""
		}
		if json.Valid([]byte(raw)) {
			return raw
		}
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(buf)
}

func jsonAuditValue(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if json.Valid([]byte(raw)) {
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err == nil {
			return value
		}
	}
	return raw
}

func configChangedKeys(oldRaw, newRaw string) []string {
	oldObj := auditObjectMap(oldRaw)
	newObj := auditObjectMap(newRaw)
	seen := map[string]bool{}
	keys := []string{}
	for key := range oldObj {
		seen[key] = true
		keys = append(keys, key)
	}
	for key := range newObj {
		if !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	out := []string{}
	for _, key := range keys {
		if !reflect.DeepEqual(oldObj[key], newObj[key]) {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func auditObjectMap(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return map[string]any{}
	}
	return value
}

func (s *Server) moderatorUserForAdmin(user domain.AuthUser) (domain.AuthUser, bool) {
	items, total := s.svc.CommunityModerators(domain.CommunityModeratorFilter{
		UserID:       user.ID,
		Status:       "1",
		ActorIsAdmin: true,
		Page:         1,
		PageSize:     1000,
	})
	if total <= 0 || len(items) == 0 {
		return domain.AuthUser{}, false
	}
	sites := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		if item.Status != 1 || strings.TrimSpace(item.CommunitySlug) == "" {
			continue
		}
		if !seen[item.CommunitySlug] {
			seen[item.CommunitySlug] = true
			sites = append(sites, item.CommunitySlug)
		}
	}
	if len(sites) == 0 {
		return domain.AuthUser{}, false
	}
	sort.Strings(sites)
	user.RoleCode = "moderator"
	user.RoleName = "子站版主"
	user.Sites = sites
	user.Permissions = []string{
		"dashboard.read",
		"site.read",
		"board.read",
		"post.read",
		"topic.moderate",
		"comment.read",
		"comment.moderate",
		"report.read",
		"report.handle",
		"moderator.read",
		"log.read",
		"plugin.read",
		"qa.question.audit",
		"docs.document.audit",
		"wiki.page.audit",
	}
	user.TokenType = "user"
	user.Identity = "moderator"
	user.Audience = "devhub_frontend"
	return user, true
}

func hasPermission(perms []string, permission string) bool {
	for _, p := range perms {
		if p == "*" || p == permission {
			return true
		}
		if strings.HasSuffix(p, ".*") && strings.HasPrefix(permission, strings.TrimSuffix(p, "*")) {
			return true
		}
	}
	return false
}

func ctxHasPermission(ctx domain.ActorContext, permission string) bool {
	return permission == "" || hasPermission(ctx.Permissions, permission)
}

func (s *Server) actorContext(c *gin.Context) domain.ActorContext {
	user, _ := currentUser(c)
	scopes := []int64{}
	if user.TokenType == "user" {
		for _, item := range s.moderatorCommunitiesForCurrentUser(c) {
			scopes = append(scopes, item.ID)
		}
	}
	ctx := service.ActorContextFromAuthUser(user, scopes)
	if user.TokenType == "user" && len(scopes) > 0 {
		ctx.IsModerator = true
	}
	return ctx
}

func (s *Server) pluginsForMenuScope(communityID int64) []domain.Plugin {
	if communityID > 0 {
		items, err := s.svc.CommunityPlugins(communityID)
		if err == nil {
			return items
		}
	}
	return s.svc.Plugins()
}

func (s *Server) filteredPluginMenus(ctx domain.ActorContext, communityID int64, area string) []domain.PluginMenu {
	menus := []domain.PluginMenu{}
	for _, plugin := range s.pluginsForMenuScope(communityID) {
		if plugin.Status != pluginregistry.StatusEnabled {
			continue
		}
		if communityID > 0 && !s.svc.IsPluginEnabledForCommunity(communityID, plugin.Code) {
			continue
		}
		if area == "moderator" && communityID > 0 && !int64In(ctx.CommunityScopes, communityID) {
			continue
		}
		for _, menu := range plugin.Menus {
			menuArea := strings.TrimSpace(firstNonEmpty(menu.Area, menu.Location))
			if menuArea != area {
				continue
			}
			if ctxHasPermission(ctx, menu.Permission) {
				menus = append(menus, menu)
			}
		}
	}
	return menus
}

func int64In(items []int64, want int64) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func hasSite(sites []string, site string) bool {
	for _, s := range sites {
		if s == "*" || s == site {
			return true
		}
	}
	return false
}

func ensureSiteAllowed(c *gin.Context, site string) bool {
	adminCtx, ok := currentAdminContext(c)
	if !ok || adminCtx.IsGlobal || site == "" || site == "portal" {
		return true
	}
	if hasSite(adminCtx.CurrentUser.Sites, site) {
		return true
	}
	fail(c, http.StatusForbidden, "无权管理该子站")
	return false
}

func adminSiteScope(c *gin.Context) string {
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		return c.DefaultQuery("site", "portal")
	}
	if adminCtx.IsGlobal {
		return c.DefaultQuery("site", adminCtx.CurrentSite)
	}
	if adminCtx.CurrentSite != "" && adminCtx.CurrentSite != "portal" && hasSite(adminCtx.CurrentUser.Sites, adminCtx.CurrentSite) {
		return adminCtx.CurrentSite
	}
	if len(adminCtx.CurrentUser.Sites) > 0 {
		return adminCtx.CurrentUser.Sites[0]
	}
	return "portal"
}

func scopedSite(c *gin.Context, requested string) (string, bool) {
	adminCtx, ok := currentAdminContext(c)
	if !ok || adminCtx.IsGlobal {
		return requested, true
	}
	if requested == "" || requested == "portal" {
		if len(adminCtx.CurrentUser.Sites) == 0 {
			fail(c, http.StatusForbidden, "未授权任何子站")
			return "", false
		}
		return adminCtx.CurrentUser.Sites[0], true
	}
	if !hasSite(adminCtx.CurrentUser.Sites, requested) {
		fail(c, http.StatusForbidden, "无权管理该子站")
		return "", false
	}
	return requested, true
}

func isAdminUser(c *gin.Context) bool {
	user, ok := currentUser(c)
	if !ok {
		return false
	}
	return user.TokenType == "admin" && (user.RoleCode == "super_admin" || hasPermission(user.Permissions, "*"))
}

func isAdminIdentity(c *gin.Context) bool {
	user, ok := currentUser(c)
	return ok && user.TokenType == "admin"
}

func (s *Server) canModerateCommunity(c *gin.Context, communityID int64) bool {
	if communityID <= 0 {
		if isAdminUser(c) {
			return true
		}
		fail(c, http.StatusForbidden, "只有管理员可以管理全局举报")
		return false
	}
	if user, ok := currentUser(c); ok && user.TokenType == "admin" {
		if user.RoleCode == "super_admin" || hasSite(user.Sites, "*") {
			return true
		}
		if comm, ok := s.communityByID(communityID); ok && hasSite(user.Sites, comm.Slug) {
			return true
		}
		fail(c, http.StatusForbidden, "无权管理该子站内容")
		return false
	}
	if s.svc.IsCommunityModerator(currentUserID(c), communityID) {
		return true
	}
	fail(c, http.StatusForbidden, "无权管理该子站内容")
	return false
}

func (s *Server) canModerateCommunityForBatch(c *gin.Context, communityID int64) bool {
	if communityID <= 0 {
		return false
	}
	if user, ok := currentUser(c); ok && user.TokenType == "admin" {
		if user.RoleCode == "super_admin" || hasSite(user.Sites, "*") {
			return true
		}
		if comm, ok := s.communityByID(communityID); ok && hasSite(user.Sites, comm.Slug) {
			return true
		}
		return false
	}
	return communityID > 0 && s.svc.IsCommunityModerator(currentUserID(c), communityID)
}

func (s *Server) canModerateTopic(c *gin.Context, topic *domain.Topic) bool {
	if topic == nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return false
	}
	return s.canModerateCommunity(c, topic.CommunityID)
}

func (s *Server) canModerateTopicForBatch(c *gin.Context, topic *domain.Topic) bool {
	return topic != nil && s.canModerateCommunityForBatch(c, topic.CommunityID)
}

func (s *Server) canManageCommunityConfig(c *gin.Context, communityID int64) bool {
	if isAdminUser(c) {
		return true
	}
	return s.canModerateCommunity(c, communityID)
}

func (s *Server) communityByID(id int64) (domain.Community, bool) {
	for _, item := range s.svc.Communities() {
		if item.ID == id {
			return item, true
		}
	}
	return domain.Community{}, false
}

func (s *Server) categoryByID(id int64) (domain.Category, bool) {
	for _, comm := range s.svc.Communities() {
		for _, category := range s.svc.Categories(comm.ID) {
			if category.ID == id {
				return category, true
			}
		}
	}
	return domain.Category{}, false
}

func (s *Server) reportFilter(c *gin.Context) domain.ReportFilter {
	page, pageSize := pagination(c)
	communityID := int64(0)
	communitySlug := strings.TrimSpace(firstQuery(c, "community_slug", "site"))
	if (communitySlug == "" || communitySlug == "all" || communitySlug == "portal") && !isAdminUser(c) {
		if scope := adminSiteScope(c); scope != "" && scope != "portal" && scope != "all" {
			communitySlug = scope
		}
	}
	if communitySlug != "" && communitySlug != "all" && communitySlug != "portal" {
		if comm, ok := s.svc.CommunityBySlug(communitySlug); ok {
			communityID = comm.ID
		}
	}
	return domain.ReportFilter{
		Status:        strings.TrimSpace(c.DefaultQuery("status", "all")),
		TargetType:    strings.TrimSpace(c.DefaultQuery("target_type", "all")),
		CommunitySlug: communitySlug,
		CommunityID:   communityID,
		Page:          page,
		PageSize:      pageSize,
		ActorUserID:   currentUserID(c),
		ActorIsAdmin:  isAdminIdentity(c),
	}
}

func (s *Server) moderatorFilter(c *gin.Context) domain.CommunityModeratorFilter {
	page, pageSize := pagination(c)
	communityID := int64(0)
	communitySlug := strings.TrimSpace(firstQuery(c, "community_slug", "site"))
	if (communitySlug == "" || communitySlug == "all" || communitySlug == "portal") && !isAdminUser(c) {
		if scope := adminSiteScope(c); scope != "" && scope != "portal" && scope != "all" {
			communitySlug = scope
		}
	}
	if communitySlug != "" && communitySlug != "all" && communitySlug != "portal" {
		if comm, ok := s.svc.CommunityBySlug(communitySlug); ok {
			communityID = comm.ID
		}
	}
	return domain.CommunityModeratorFilter{
		CommunitySlug: communitySlug,
		CommunityID:   communityID,
		UserID:        int64Query(c, "user_id", 0),
		Status:        strings.TrimSpace(c.DefaultQuery("status", "all")),
		Page:          page,
		PageSize:      pageSize,
		ActorUserID:   currentUserID(c),
		ActorIsAdmin:  isAdminIdentity(c),
	}
}

func (s *Server) ensureModeratorRequestScope(c *gin.Context, req domain.CommunityModeratorRequest) bool {
	if isAdminUser(c) {
		return true
	}
	communityID := req.CommunityID
	if communityID == 0 && strings.TrimSpace(req.CommunitySlug) != "" {
		if comm, ok := s.svc.CommunityBySlug(strings.TrimSpace(req.CommunitySlug)); ok {
			communityID = comm.ID
		}
	}
	return s.canModerateCommunity(c, communityID)
}

func (s *Server) toggleAdminTopicBool(c *gin.Context, action string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	var updated *domain.Topic
	switch action {
	case "feature":
		updated, err = s.svc.SetTopicFeatured(id, !topic.IsFeatured)
	case "pin":
		updated, err = s.svc.SetTopicPinned(id, !topic.IsPinned)
	default:
		err = fmt.Errorf("不支持的操作")
	}
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	auditAction := map[string]string{"feature": "切换精华", "pin": "切换置顶"}[action]
	s.audit(c, "audit", auditAction, fmt.Sprintf("topics#%d", id))
	s.auditPluginContentAction(c, auditAction, topic, action,
		gin.H{"featured": topic.IsFeatured, "pinned": topic.IsPinned},
		gin.H{"featured": updated.IsFeatured, "pinned": updated.IsPinned},
		nil)
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) applyBatchTopicAction(id int64, topic *domain.Topic, action string) error {
	switch action {
	case "feature":
		_, err := s.svc.SetTopicFeatured(id, true)
		return err
	case "unfeature":
		_, err := s.svc.SetTopicFeatured(id, false)
		return err
	case "pin":
		_, err := s.svc.SetTopicPinned(id, true)
		return err
	case "unpin":
		_, err := s.svc.SetTopicPinned(id, false)
		return err
	case "hide":
		_, err := s.svc.SetTopicStatus(id, 0)
		return err
	case "restore":
		_, err := s.svc.SetTopicStatus(id, 1)
		return err
	case "lock-comments":
		_, err := s.svc.SetTopicCommentLocked(id, true)
		return err
	case "unlock-comments":
		_, err := s.svc.SetTopicCommentLocked(id, false)
		return err
	case "delete":
		if !s.svc.DeleteTopic(id) {
			return fmt.Errorf("主题不存在")
		}
		return nil
	default:
		return fmt.Errorf("不支持的批量主题操作")
	}
}

func (s *Server) applyBatchCommentAction(id int64, action string) error {
	switch action {
	case "hide":
		_, err := s.svc.SetCommentStatus(id, "hidden")
		return err
	case "restore":
		_, err := s.svc.SetCommentStatus(id, "normal")
		return err
	case "delete":
		if !s.svc.DeleteComment(id) {
			return fmt.Errorf("评论不存在")
		}
		return nil
	default:
		return fmt.Errorf("不支持的批量评论操作")
	}
}

func (s *Server) setAdminTopicStatus(c *gin.Context, status int, action string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	updated, err := s.svc.SetTopicStatus(id, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "audit", action, fmt.Sprintf("topics#%d", id))
	s.auditPluginContentAction(c, action, topic, "status",
		gin.H{"status": topic.Status},
		gin.H{"status": updated.Status},
		nil)
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) setAdminTopicLock(c *gin.Context, locked bool) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	updated, err := s.svc.SetTopicCommentLocked(id, locked)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "锁定评论"
	if !locked {
		action = "解锁评论"
	}
	s.audit(c, "audit", action, fmt.Sprintf("topics#%d", id))
	s.auditPluginContentAction(c, action, topic, "comments_locked",
		gin.H{"comment_locked": topic.CommentLocked},
		gin.H{"comment_locked": updated.CommentLocked},
		nil)
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) setAdminCommentStatus(c *gin.Context, status string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	comment, err := s.svc.CommentByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	topic, err := s.svc.TopicByID(comment.TopicID, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateTopic(c, topic) {
		return
	}
	updated, err := s.svc.SetCommentStatus(id, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "隐藏评论"
	if status == "normal" {
		action = "恢复评论"
	}
	s.audit(c, "audit", action, fmt.Sprintf("comments#%d", id))
	s.auditPluginContentAction(c, action, topic, "comment_status",
		gin.H{"comment_id": id, "status": comment.Status},
		gin.H{"comment_id": id, "status": updated.Status},
		nil)
	c.JSON(http.StatusOK, gin.H{"comment": updated, "changed": true})
}

func (s *Server) auditPluginContentAction(c *gin.Context, action string, topic *domain.Topic, operation string, oldValue, newValue, metadata any) {
	if topic == nil {
		return
	}
	pluginCode := strings.TrimSpace(topic.PluginCode)
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return
	}
	meta := gin.H{
		"operation":    operation,
		"plugin_code":  pluginCode,
		"content_type": topic.ContentType,
		"content_id":   topic.ID,
		"community_id": topic.CommunityID,
		"category_id":  topic.CategoryID,
	}
	if extra, ok := metadata.(gin.H); ok {
		for k, v := range extra {
			meta[k] = v
		}
	}
	s.auditStructured(c, "audit", action, fmt.Sprintf("plugins#%s/topics#%d", pluginCode, topic.ID), oldValue, newValue, meta)
}

func (s *Server) moderatorCommunitiesForCurrentUser(c *gin.Context) []domain.Community {
	items, _ := s.svc.CommunityModerators(domain.CommunityModeratorFilter{
		UserID:       currentUserID(c),
		Status:       "1",
		ActorIsAdmin: true,
		Page:         1,
		PageSize:     1000,
	})
	seen := map[int64]bool{}
	communities := make([]domain.Community, 0, len(items))
	for _, item := range items {
		if item.Status != 1 || item.CommunityID <= 0 || seen[item.CommunityID] {
			continue
		}
		if comm, ok := s.communityByID(item.CommunityID); ok && comm.Status == 1 {
			communities = append(communities, comm)
			seen[item.CommunityID] = true
		}
	}
	sort.Slice(communities, func(i, j int) bool {
		if communities[i].SortOrder == communities[j].SortOrder {
			return communities[i].ID < communities[j].ID
		}
		return communities[i].SortOrder < communities[j].SortOrder
	})
	return communities
}

func (s *Server) moderatorRequestedCommunityID(c *gin.Context) (int64, bool) {
	communityID := int64Query(c, "community_id", 0)
	if communityID == 0 {
		slug := strings.TrimSpace(firstQuery(c, "community_slug", "site"))
		if slug != "" && slug != "all" && slug != "portal" {
			if comm, ok := s.svc.CommunityBySlug(slug); ok {
				communityID = comm.ID
			} else {
				fail(c, http.StatusNotFound, "子站不存在")
				return 0, false
			}
		}
	}
	if communityID > 0 && !s.canModerateCommunityStrict(c, communityID) {
		return 0, false
	}
	return communityID, true
}

func (s *Server) moderatorCommunityIDs(c *gin.Context, requested int64) []int64 {
	if requested > 0 {
		return []int64{requested}
	}
	communities := s.moderatorCommunitiesForCurrentUser(c)
	ids := make([]int64, 0, len(communities))
	for _, comm := range communities {
		ids = append(ids, comm.ID)
	}
	return ids
}

func (s *Server) moderatorReportsForCommunities(communityIDs []int64, status, targetType string, page, pageSize int) []domain.Report {
	out := []domain.Report{}
	for _, communityID := range uniqueInt64s(communityIDs) {
		if communityID <= 0 {
			continue
		}
		items, _ := s.svc.Reports(domain.ReportFilter{
			Status:       status,
			TargetType:   targetType,
			CommunityID:  communityID,
			Page:         page,
			PageSize:     pageSize,
			ActorIsAdmin: true,
		})
		out = append(out, items...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (s *Server) moderatorTopicsForCommunities(communityIDs []int64, status, contentType, keyword string, page, pageSize int) []domain.Post {
	out := []domain.Post{}
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	for _, communityID := range uniqueInt64s(communityIDs) {
		if communityID <= 0 {
			continue
		}
		site := slugByCommunityID(communityID)
		if comm, ok := s.communityByID(communityID); ok && comm.Slug != "" {
			site = comm.Slug
		}
		posts := s.svc.AdminTopics(site, "all", keyword)
		for _, post := range posts {
			if status != "" && status != "all" {
				switch status {
				case "featured":
					if !post.Recommended {
						continue
					}
				case "pinned":
					if !post.Pinned {
						continue
					}
				case "locked":
					if !post.CommentLocked {
						continue
					}
				default:
					if post.Status != status {
						continue
					}
				}
			}
			if contentType != "" && contentType != "all" && adminContentTypeByBoard(post.Board) != pluginregistry.NormalizeContentType(contentType) {
				continue
			}
			out = append(out, post)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return paginate(out, page, pageSize)
}

func (s *Server) moderatorCommentsForCommunities(communityIDs []int64, status, keyword string, page, pageSize int) []domain.AdminComment {
	out := []domain.AdminComment{}
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	for _, communityID := range uniqueInt64s(communityIDs) {
		if communityID <= 0 {
			continue
		}
		site := slugByCommunityID(communityID)
		if comm, ok := s.communityByID(communityID); ok && comm.Slug != "" {
			site = comm.Slug
		}
		comments := s.svc.AdminComments(site)
		for _, comment := range comments {
			if status != "" && status != "all" && comment.Status != status {
				continue
			}
			if keyword != "" {
				haystack := strings.ToLower(strings.Join([]string{comment.PostTitle, comment.Author, comment.To, comment.Text}, " "))
				if !strings.Contains(haystack, keyword) {
					continue
				}
			}
			out = append(out, comment)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return paginate(out, page, pageSize)
}

func (s *Server) moderatorLogsForCommunities(c *gin.Context, communityIDs []int64, actorType, logType, action, targetType string, page, pageSize int) []domain.AdminLog {
	out := []domain.AdminLog{}
	if strings.TrimSpace(logType) == "" {
		logType = "all"
	}
	for _, communityID := range uniqueInt64s(communityIDs) {
		if communityID <= 0 {
			continue
		}
		filter := domain.AdminLogFilter{
			Type:        logType,
			ActorType:   strings.TrimSpace(actorType),
			Action:      strings.TrimSpace(action),
			TargetType:  strings.TrimSpace(targetType),
			CommunityID: communityID,
			Page:        page,
			PageSize:    pageSize,
		}
		items, _ := s.svc.AdminLogsByFilter(filter)
		out = append(out, items...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return paginate(out, page, pageSize)
}

func (s *Server) canModerateCommunityStrict(c *gin.Context, communityID int64) bool {
	if communityID <= 0 {
		fail(c, http.StatusForbidden, "版主只能治理具体子站")
		return false
	}
	user, ok := currentUser(c)
	if !ok || user.TokenType != "user" {
		fail(c, http.StatusForbidden, "版主工作台仅支持前台用户身份")
		return false
	}
	if s.svc.IsCommunityModerator(user.ID, communityID) {
		return true
	}
	fail(c, http.StatusForbidden, "无权管理该子站内容")
	return false
}

func (s *Server) moderatorAudit(c *gin.Context, communityID int64, action, target string) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	site := slugByCommunityID(communityID)
	if comm, ok := s.communityByID(communityID); ok && comm.Slug != "" {
		site = comm.Slug
	}
	actor := firstNonEmpty(user.Username, user.Nickname, "moderator")
	s.svc.AppendAdminLog(domain.AdminLog{
		Site:        site,
		Type:        "audit",
		Actor:       actor,
		ActorType:   "moderator",
		ActorUserID: user.ID,
		ActorID:     user.ID,
		Role:        "moderator",
		Action:      action,
		Target:      target,
		CommunityID: communityID,
		IP:          c.ClientIP(),
	})
}

func (s *Server) setModeratorTopicFeatured(c *gin.Context, featured bool) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateCommunityStrict(c, topic.CommunityID) {
		return
	}
	updated, err := s.svc.SetTopicFeatured(id, featured)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "feature_topic"
	if !featured {
		action = "unfeature_topic"
	}
	s.moderatorAudit(c, topic.CommunityID, action, fmt.Sprintf("topics#%d", id))
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) setModeratorTopicPinned(c *gin.Context, pinned bool) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateCommunityStrict(c, topic.CommunityID) {
		return
	}
	updated, err := s.svc.SetTopicPinned(id, pinned)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "pin_topic"
	if !pinned {
		action = "unpin_topic"
	}
	s.moderatorAudit(c, topic.CommunityID, action, fmt.Sprintf("topics#%d", id))
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) setModeratorTopicStatus(c *gin.Context, status int, action string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateCommunityStrict(c, topic.CommunityID) {
		return
	}
	updated, err := s.svc.SetTopicStatus(id, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.moderatorAudit(c, topic.CommunityID, action, fmt.Sprintf("topics#%d", id))
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) setModeratorTopicLock(c *gin.Context, locked bool) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateCommunityStrict(c, topic.CommunityID) {
		return
	}
	updated, err := s.svc.SetTopicCommentLocked(id, locked)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "lock_comments"
	if !locked {
		action = "unlock_comments"
	}
	s.moderatorAudit(c, topic.CommunityID, action, fmt.Sprintf("topics#%d", id))
	c.JSON(http.StatusOK, gin.H{"topic": updated, "changed": true})
}

func (s *Server) setModeratorCommentStatus(c *gin.Context, status string) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	comment, err := s.svc.CommentByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	topic, err := s.svc.TopicByID(comment.TopicID, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if !s.canModerateCommunityStrict(c, topic.CommunityID) {
		return
	}
	updated, err := s.svc.SetCommentStatus(id, status)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	action := "hide_comment"
	if status == "normal" {
		action = "restore_comment"
	}
	s.moderatorAudit(c, topic.CommunityID, action, fmt.Sprintf("comments#%d", id))
	c.JSON(http.StatusOK, gin.H{"comment": updated, "changed": true})
}

func countTodayModeratorActions(logs []domain.AdminLog) int {
	today := time.Now().Format("2006-01-02")
	count := 0
	for _, log := range logs {
		if log.ActorType == "moderator" && strings.HasPrefix(log.CreatedAt, today) {
			count++
		}
	}
	return count
}

func limitReports(items []domain.Report, limit int) []domain.Report {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// fail 输出统一错误响应。
func fail(c *gin.Context, code int, msg string) { c.JSON(code, gin.H{"error": msg}) }

// idParam 解析正整数路径参数，解析失败时直接写入 400 响应。
func idParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "ID 不合法")
		return 0, false
	}
	return id, true
}

// intQuery 解析正整数查询参数，缺失或非法时使用默认值。
func intQuery(c *gin.Context, name string, def int) int {
	v, err := strconv.Atoi(c.Query(name))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func int64Query(c *gin.Context, name string, def int64) int64 {
	v, err := strconv.ParseInt(c.Query(name), 10, 64)
	if err != nil || v < 0 {
		return def
	}
	return v
}

// pagination 解析分页参数，并限制 page_size 最大值。
func pagination(c *gin.Context) (int, int) {
	page := intQuery(c, "page", 1)
	pageSize := intQuery(c, "page_size", 10)
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

// paginate 按页码和页大小截取列表。
func paginate[T any](items []T, page, pageSize int) []T {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []T{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

// ===== 新增：DevHub 通用社区系统 API 处理器 =====

func (s *Server) listCommunities(c *gin.Context) {
	communities := s.svc.Communities()
	c.JSON(http.StatusOK, gin.H{"items": communities})
}

func (s *Server) getCommunity(c *gin.Context) {
	comm, ok := s.svc.CommunityBySlug(c.Param("slug"))
	if !ok {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	c.JSON(http.StatusOK, comm)
}

func (s *Server) communityOverview(c *gin.Context) {
	slug := c.Param("slug")
	overview, ok := s.svc.CommunityOverview(slug)
	if !ok {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (s *Server) listCategories(c *gin.Context) {
	slug := c.Param("slug")
	comm, ok := s.svc.CommunityBySlug(slug)
	if !ok {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	categories := s.svc.Categories(comm.ID)
	filtered := make([]domain.Category, 0, len(categories))
	for _, cat := range categories {
		if cat.Status != 1 || !cat.Visible {
			continue
		}
		if s.svc.IsPluginEnabledForCommunity(comm.ID, firstNonEmpty(cat.PluginCode, pluginregistry.CoreCode)) {
			filtered = append(filtered, cat)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": filtered})
}

func (s *Server) communityPlugins(c *gin.Context) {
	comm, ok := s.svc.CommunityBySlug(c.Param("slug"))
	if !ok || comm.Status != 1 {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	items, err := s.svc.CommunityPlugins(comm.ID)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	enabled := make([]domain.Plugin, 0, len(items))
	for _, plugin := range items {
		if plugin.Status == pluginregistry.StatusEnabled {
			plugin.GlobalStatus = ""
			plugin.CommunityStatus = ""
			plugin.ConfigJSON = ""
			plugin.ResolvedConfig = nil
			enabled = append(enabled, plugin)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": enabled})
}

func (s *Server) listCommunityTags(c *gin.Context) {
	slug := c.Param("slug")
	comm, ok := s.svc.CommunityBySlug(slug)
	if !ok {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	topics, _ := s.svc.TopicsByFilter(comm.ID, 0, "", "hot", nil, "", 1, 200)
	counts := map[string]int{}
	for _, topic := range topics {
		for _, tag := range topic.Tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				counts[tag]++
			}
		}
	}
	hotTags := make([]domain.TagStat, 0, len(counts))
	for name, count := range counts {
		hotTags = append(hotTags, domain.TagStat{Name: name, Count: count})
	}
	sort.Slice(hotTags, func(i, j int) bool {
		if hotTags[i].Count == hotTags[j].Count {
			return hotTags[i].Name < hotTags[j].Name
		}
		return hotTags[i].Count > hotTags[j].Count
	})
	if len(hotTags) == 0 {
		hotTags = s.svc.TagStats(comm.Slug)
	}
	c.JSON(http.StatusOK, gin.H{"items": hotTags})
}

func (s *Server) communityStats(c *gin.Context) {
	comm, ok := s.svc.CommunityBySlug(c.Param("slug"))
	if !ok || comm.Status != 1 {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	c.JSON(http.StatusOK, s.svc.CommunityStats(comm.ID))
}

func (s *Server) listCommunityModerators(c *gin.Context) {
	comm, ok := s.svc.CommunityBySlug(c.Param("slug"))
	if !ok || comm.Status != 1 {
		fail(c, http.StatusNotFound, "子站不存在")
		return
	}
	items, total := s.svc.CommunityModerators(domain.CommunityModeratorFilter{CommunityID: comm.ID, Status: "1", Page: 1, PageSize: 50, ActorIsAdmin: true})
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (s *Server) listTopics(c *gin.Context) {
	communityID := int64(0)
	communitySlug := firstQuery(c, "community_slug", "community", "site")
	if communitySlug != "" && communitySlug != "all" && communitySlug != "portal" {
		if comm, ok := s.svc.CommunityBySlug(communitySlug); ok {
			communityID = comm.ID
		}
	}

	categoryID := int64(0)
	categorySlug := firstQuery(c, "category_slug", "category", "board")
	if categorySlug != "" {
		categoryID = s.categoryIDBySlug(communityID, categorySlug)
	}

	contentType := pluginregistry.NormalizeContentType(firstQuery(c, "content_type", "type"))
	if contentType == "" {
		contentType = contentTypeByBoard(categorySlug)
	}
	sort := c.DefaultQuery("sort", "latest")
	var isSolved *bool
	if sort == "unsolved" {
		zero := false
		isSolved = &zero
		sort = "latest"
	}

	if c.Query("solved") == "1" {
		one := true
		isSolved = &one
	} else if c.Query("solved") == "0" {
		zero := false
		isSolved = &zero
	}
	if c.Query("is_solved") == "1" {
		one := true
		isSolved = &one
	} else if c.Query("is_solved") == "0" {
		zero := false
		isSolved = &zero
	}

	tag := c.Query("tag")
	page, pageSize := pagination(c)

	topics, total := s.svc.TopicsByFilter(communityID, categoryID, contentType, sort, isSolved, tag, page, pageSize)
	if c.Query("is_featured") == "1" {
		filtered := make([]domain.Topic, 0, len(topics))
		for _, topic := range topics {
			if topic.IsFeatured {
				filtered = append(filtered, topic)
			}
		}
		topics = filtered
		total = len(filtered)
	}
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    topics,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *Server) getTopic(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, true)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	s.applyTopicInteraction(c, topic)
	c.JSON(http.StatusOK, topic)
}

func (s *Server) topicQA(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if topic.ContentType != "question" {
		fail(c, http.StatusBadRequest, "当前主题不是 question")
		return
	}
	question, err := s.svc.QAQuestionByTopicID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "问答扩展不存在")
		return
	}
	answers, err := s.svc.QAAnswersByTopicID(id)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"topic_id": id, "question": question, "answers": answers})
}

func (s *Server) topicDocs(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if topic.ContentType != "document" {
		fail(c, http.StatusBadRequest, "当前主题不是 document")
		return
	}
	document, err := s.svc.DocsDocumentByTopicID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "文档扩展不存在")
		return
	}
	tree, err := s.svc.DocsTree(topic.CommunityID, document.SpaceID)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"topic_id": id, "document": document, "tree": tree})
}

func (s *Server) topicWikiVersions(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	if topic.ContentType != "wiki_page" {
		fail(c, http.StatusBadRequest, "当前主题不是 wiki_page")
		return
	}
	page, err := s.svc.WikiPageByTopicID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "Wiki 页面扩展不存在")
		return
	}
	versions, err := s.svc.WikiVersionsByTopicID(id)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"topic_id": id, "page": page, "versions": versions})
}

func (s *Server) topicComments(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if _, err := s.svc.TopicByID(id, false); err != nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	sortBy := strings.TrimSpace(c.DefaultQuery("sort", "best"))
	switch sortBy {
	case "latest", "oldest", "best":
	default:
		sortBy = "best"
	}
	page, pageSize := pagination(c)
	comments, total := s.svc.TopicComments(id, sortBy, page, pageSize)
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  page*pageSize < total,
		Filters:  gin.H{"sort": sortBy},
	})
}

func (s *Server) createTopicComment(c *gin.Context) {
	topicID, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ParentID = 0
	s.fillCommentUser(c, &req)
	comment, err := s.svc.CreateCommentWithRequest(topicID, req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "不存在") {
			status = http.StatusNotFound
		}
		fail(c, status, err.Error())
		return
	}
	topic, _ := s.svc.TopicByID(topicID, false)
	c.JSON(http.StatusCreated, gin.H{"comment": comment, "item": comment, "topic": topic})
}

func (s *Server) replyTopicComment(c *gin.Context) {
	topicID, ok := idParam(c, "id")
	if !ok {
		return
	}
	commentID, ok := idParam(c, "commentId")
	if !ok {
		return
	}
	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ParentID = commentID
	s.fillCommentUser(c, &req)
	comment, err := s.svc.CreateCommentWithRequest(topicID, req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "不存在") {
			status = http.StatusNotFound
		}
		fail(c, status, err.Error())
		return
	}
	topic, _ := s.svc.TopicByID(topicID, false)
	c.JSON(http.StatusCreated, gin.H{"comment": comment, "item": comment, "topic": topic})
}

func (s *Server) fillCommentUser(c *gin.Context, req *domain.CreateCommentRequest) {
	user := currentUserOrAuth(c)
	if user.ID <= 0 {
		user.ID = 1
	}
	req.UserID = user.ID
	req.ActorUserID = user.ID
	req.ActorUserName = firstNonEmpty(user.Nickname, user.Username, "Demo 用户")
	if strings.TrimSpace(req.Author) == "" {
		req.Author = req.ActorUserName
	}
}

func (s *Server) acceptTopicComment(c *gin.Context) {
	topicID, ok := idParam(c, "id")
	if !ok {
		return
	}
	commentID, ok := idParam(c, "commentId")
	if !ok {
		return
	}
	if !s.canAcceptAnswer(c, topicID) {
		return
	}
	if s.svc.AcceptBestAnswer(topicID, commentID, currentUserID(c)) {
		topic, _ := s.svc.TopicByID(topicID, false)
		c.JSON(http.StatusOK, gin.H{"accepted": true, "solved": true, "best_comment_id": commentID, "topic": topic})
		return
	}
	fail(c, http.StatusBadRequest, "采纳失败，请确认主题类型和评论状态")
}

func (s *Server) canAcceptAnswer(c *gin.Context, topicID int64) bool {
	topic, err := s.svc.TopicByID(topicID, false)
	if err != nil || topic == nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return false
	}
	if topic.ContentType != "question" {
		fail(c, http.StatusBadRequest, "只有问答主题可以采纳答案")
		return false
	}
	user := currentUserOrAuth(c)
	if user.ID == topic.UserID || user.RoleCode == "super_admin" || hasPermission(user.Permissions, "comment.moderate") {
		return true
	}
	fail(c, http.StatusForbidden, "只有主题作者或管理员可以采纳答案")
	return false
}

func (s *Server) createTopic(c *gin.Context) {
	user := currentUserOrAuth(c)
	if user.ID == 0 {
		user.ID = 1
	}

	var req domain.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.normalizeCreateTopicRequest(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.UserID = user.ID
	req.ActorContext = s.actorContext(c)
	req.ActorPermissions = req.ActorContext.Permissions
	topic, err := s.svc.CreateTopic(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":         topic.ID,
		"detail_url": fmt.Sprintf("/topics/%d/", topic.ID),
		"topic":      topic,
	})
}

func (s *Server) updateTopic(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	var req domain.UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	// Frontend topic edits are user-scoped. Admin/moderator governance must use /api/v1/admin/* or /api/v1/moderator/*.
	topic, err := s.svc.TopicByID(id, false)
	if err != nil || topic == nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	user, _ := currentUser(c)
	if user.ID == 0 || topic.UserID != user.ID {
		fail(c, http.StatusForbidden, "无权编辑该主题")
		return
	}
	if req.ContentType != nil || req.PluginCode != nil || req.CommunityID != nil || req.CategoryID != nil || req.CommunitySlug != nil {
		fail(c, http.StatusBadRequest, "不允许修改主题归属或内容类型")
		return
	}

	req.ActorContext = s.actorContext(c)
	topic, err = s.svc.UpdateTopic(id, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, topic)
}

func (s *Server) deleteTopic(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	topic, err := s.svc.TopicByID(id, false)
	if err != nil || topic == nil {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	user, _ := currentUser(c)
	if user.ID == 0 || topic.UserID != user.ID {
		fail(c, http.StatusForbidden, "无权删除该主题")
		return
	}
	if !s.svc.DeleteTopic(id) {
		fail(c, http.StatusNotFound, "主题不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) likeTopic(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	userID := currentUserID(c)
	liked, count, err := s.svc.ToggleReaction(userID, id, "topic", "like")
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	interaction, _ := s.svc.TopicInteraction(userID, id)
	if interaction.LikeCount == 0 {
		interaction.LikeCount = count
	}
	c.JSON(http.StatusOK, gin.H{"liked": liked, "like_count": interaction.LikeCount, "count": interaction.LikeCount, "hot_score": interaction.HotScore})
}

func (s *Server) favoriteTopic(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	userID := currentUserID(c)
	favorited, err := s.svc.ToggleFavorite(userID, id, "topic")
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	interaction, _ := s.svc.TopicInteraction(userID, id)
	c.JSON(http.StatusOK, gin.H{"favorited": favorited, "favorite_count": interaction.FavoriteCount, "hot_score": interaction.HotScore})
}

func (s *Server) topicInteraction(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	interaction, err := s.svc.TopicInteraction(currentUserID(c), id)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, interaction)
}

func (s *Server) solveTopic(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		CommentID int64 `json:"comment_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	if !s.canAcceptAnswer(c, id) {
		return
	}
	if s.svc.AcceptBestAnswer(id, req.CommentID, currentUserID(c)) {
		c.JSON(http.StatusOK, gin.H{"accepted": true, "solved": true, "best_comment_id": req.CommentID})
		return
	}
	fail(c, http.StatusBadRequest, "采纳失败，请确认主题类型和评论状态")
}

func (s *Server) searchTopics(c *gin.Context) {
	req := s.searchRequestFromQuery(c)
	topics, total := s.svc.SearchTopics(req)
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    topics,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		HasMore:  req.Page*req.PageSize < total,
		Filters:  req,
	})
}

func (s *Server) searchRequestFromQuery(c *gin.Context) domain.SearchRequest {
	page, pageSize := pagination(c)
	scope := strings.TrimSpace(c.DefaultQuery("scope", "all"))
	switch scope {
	case "all", "current", "community":
	default:
		scope = "all"
	}

	communitySlug := strings.TrimSpace(firstQuery(c, "community_slug", "site"))
	communityID := int64(0)
	if (scope == "community" || scope == "current") && communitySlug != "" && communitySlug != "all" && communitySlug != "portal" {
		if comm, ok := s.svc.CommunityBySlug(communitySlug); ok {
			communityID = comm.ID
		}
	} else if scope == "community" && communitySlug == "" {
		scope = "all"
	}

	categoryID := int64(intQuery(c, "category_id", 0))
	if categoryID == 0 {
		categorySlug := firstQuery(c, "category_slug", "category", "board")
		if categorySlug != "" {
			categoryID = s.categoryIDBySlug(communityID, categorySlug)
		}
	}

	contentType := pluginregistry.NormalizeContentType(firstQuery(c, "content_type", "type"))
	if !validContentType(contentType) {
		contentType = ""
	}
	sortBy := strings.TrimSpace(c.DefaultQuery("sort", "latest"))
	switch sortBy {
	case "latest", "active", "hot", "featured", "unsolved":
	default:
		sortBy = "latest"
	}

	req := domain.SearchRequest{
		Keyword:       strings.TrimSpace(firstQuery(c, "keyword", "q")),
		Scope:         scope,
		CommunitySlug: communitySlug,
		CommunityID:   communityID,
		CategoryID:    categoryID,
		ContentType:   contentType,
		Tag:           strings.TrimSpace(c.Query("tag")),
		TagID:         int64(intQuery(c, "tag_id", 0)),
		Sort:          sortBy,
		Page:          page,
		PageSize:      pageSize,
	}
	if req.TagID == 0 && req.Tag != "" {
		resolveSite := ""
		if req.CommunitySlug != "" && req.CommunitySlug != "portal" && req.CommunitySlug != "all" {
			resolveSite = req.CommunitySlug
		}
		if resolved, ok := s.svc.ResolveTag(resolveSite, req.Tag); ok {
			req.TagID = resolved.Tag.ID
			req.Tag = firstNonEmpty(resolved.Tag.Slug, resolved.Tag.Name)
		}
	}
	return req
}

func (s *Server) applyTopicInteraction(c *gin.Context, topic *domain.Topic) {
	if topic == nil || topic.ID <= 0 {
		return
	}
	interaction, err := s.svc.TopicInteraction(currentUserID(c), topic.ID)
	if err != nil {
		return
	}
	topic.Liked = interaction.Liked
	topic.Favorited = interaction.Favorited
	topic.Followed = interaction.Followed
	topic.LikeCount = interaction.LikeCount
	topic.FavoriteCount = interaction.FavoriteCount
	topic.HotScore = interaction.HotScore
	if user, ok := currentUser(c); ok && user.ID == topic.UserID {
		topic.CanEdit = true
		topic.CanDelete = true
	}
}

func (s *Server) toggleReaction(c *gin.Context) {
	var req domain.ToggleReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	reacted, count, err := s.svc.ToggleReaction(currentUserID(c), req.TargetID, req.TargetType, "like")
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"reacted": reacted, "count": count})
}

func (s *Server) toggleFavorite(c *gin.Context) {
	var req domain.ToggleReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	favorited, err := s.svc.ToggleFavorite(currentUserID(c), req.TargetID, req.TargetType)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp := gin.H{"favorited": favorited}
	if strings.TrimSpace(req.TargetType) == "topic" || strings.TrimSpace(req.TargetType) == "" {
		interaction, _ := s.svc.TopicInteraction(currentUserID(c), req.TargetID)
		resp["favorite_count"] = interaction.FavoriteCount
		resp["hot_score"] = interaction.HotScore
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) toggleFollow(c *gin.Context) {
	var req domain.ToggleReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	followed, err := s.svc.ToggleFollow(currentUserID(c), req.TargetID, req.TargetType)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"followed": followed, "target_type": req.TargetType, "target_id": req.TargetID})
}

func (s *Server) userActivities(c *gin.Context) {
	communitySlug := c.Query("community")
	communityID := int64(0)

	if communitySlug != "" && communitySlug != "all" && communitySlug != "portal" {
		if comm, ok := s.svc.CommunityBySlug(communitySlug); ok {
			communityID = comm.ID
		}
	}

	page, pageSize := pagination(c)
	activities, total := s.svc.UserActivities(currentUserID(c), communityID, strings.TrimSpace(c.Query("action")), page, pageSize)
	c.JSON(http.StatusOK, domain.PageResponse{
		Items:    activities,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  page*pageSize < total,
	})
}

func (s *Server) myFavorites(c *gin.Context) {
	page, pageSize := pagination(c)
	items, total := s.svc.UserFavorites(currentUserID(c), strings.TrimSpace(c.Query("target_type")), page, pageSize)
	c.JSON(http.StatusOK, domain.PageResponse{Items: items, Total: total, Page: page, PageSize: pageSize, HasMore: page*pageSize < total})
}

func (s *Server) myFollows(c *gin.Context) {
	page, pageSize := pagination(c)
	items, total := s.svc.UserFollows(currentUserID(c), strings.TrimSpace(c.Query("target_type")), page, pageSize)
	c.JSON(http.StatusOK, domain.PageResponse{Items: items, Total: total, Page: page, PageSize: pageSize, HasMore: page*pageSize < total})
}

func (s *Server) myActivities(c *gin.Context) {
	s.userActivities(c)
}

func (s *Server) myNotifications(c *gin.Context) {
	page, pageSize := pagination(c)
	isRead := readFilter(c.Query("is_read"))
	items, total, unread := s.svc.UserNotifications(currentUserID(c), isRead, page, pageSize)
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize, "has_more": page*pageSize < total, "unread_count": unread})
}

func (s *Server) readMyNotification(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.svc.ReadUserNotification(currentUserID(c), id) {
		fail(c, http.StatusNotFound, "通知不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{"read": true, "id": id})
}

func (s *Server) readAllMyNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"read": true, "updated": s.svc.ReadAllUserNotifications(currentUserID(c))})
}

func readFilter(value string) *bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "read":
		v := true
		return &v
	case "0", "false", "no", "unread":
		v := false
		return &v
	default:
		return nil
	}
}

func currentUserID(c *gin.Context) int64 {
	if user, ok := currentUser(c); ok && user.ID > 0 {
		return user.ID
	}
	return 1
}

// currentUserOrAuth 获取当前用户，未登录返回 demo 用户 ID 1。
func currentUserOrAuth(c *gin.Context) domain.AuthUser {
	if user, ok := currentUser(c); ok {
		return user
	}
	return domain.AuthUser{ID: 1, Username: "demo", Nickname: "Demo 用户"}
}

func firstQuery(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
	}
	return ""
}

func (s *Server) createPostToTopicRequest(req domain.CreatePostRequest) (domain.CreateTopicRequest, error) {
	req.Site = strings.TrimSpace(req.Site)
	req.Board = strings.TrimSpace(req.Board)
	if req.Site == "" || req.Site == "portal" {
		return domain.CreateTopicRequest{}, fmt.Errorf("请选择具体子站")
	}
	communityID := communityIDBySlug(req.Site)
	if communityID <= 0 {
		return domain.CreateTopicRequest{}, fmt.Errorf("子站不存在")
	}
	if req.Board == "" || req.Board == "all" {
		req.Board = "community"
	}
	categoryID := categoryIDByBoard(communityID, req.Board)
	contentType := adminContentTypeByBoard(req.Board)
	return domain.CreateTopicRequest{
		UserID:        currentDemoUserID(),
		CommunityID:   communityID,
		CommunitySlug: req.Site,
		CategoryID:    categoryID,
		Title:         req.Title,
		ContentType:   contentType,
		Summary:       req.Summary,
		Content:       req.Content,
		Tags:          req.Tags,
	}, nil
}

func (s *Server) updatePostToTopicRequest(req domain.UpdatePostRequest, topic *domain.Topic) (domain.UpdateTopicRequest, error) {
	out := domain.UpdateTopicRequest{}
	if req.Site != nil {
		if strings.TrimSpace(*req.Site) != slugByCommunityID(topic.CommunityID) {
			return out, fmt.Errorf("后台编辑不允许修改内容归属子站，请通过迁移专项处理")
		}
	}
	if req.Board != nil {
		board := strings.TrimSpace(*req.Board)
		if board == "" || board == "all" {
			return out, fmt.Errorf("请选择具体板块")
		}
		contentType := adminContentTypeByBoard(board)
		categoryID := categoryIDByBoard(topic.CommunityID, board)
		if categoryID != topic.CategoryID || pluginregistry.NormalizeContentType(contentType) != pluginregistry.NormalizeContentType(topic.ContentType) {
			return out, fmt.Errorf("后台编辑不允许修改内容板块或内容类型，请通过迁移专项处理")
		}
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		out.Title = &title
	}
	if req.Summary != nil {
		summary := strings.TrimSpace(*req.Summary)
		out.Summary = &summary
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		out.Content = &content
	}
	if req.Status != nil {
		status := postStatusToTopicStatus(*req.Status)
		out.Status = &status
	}
	if req.Pinned != nil {
		out.IsPinned = req.Pinned
	}
	if req.Recommended != nil {
		out.IsFeatured = req.Recommended
	}
	if req.Tags != nil {
		tags := uniqueStrings(*req.Tags)
		out.Tags = &tags
	}
	return out, nil
}

func topicToAdminPost(topic domain.Topic) domain.Post {
	site := slugByCommunityID(topic.CommunityID)
	board := boardByContentTypeHTTP(topic.ContentType)
	status := "publish"
	if topic.Status == 0 {
		status = "offline"
	} else if topic.Status == 3 {
		status = "deleted"
	}
	return domain.Post{
		ID:            topic.ID,
		UserID:        topic.UserID,
		Site:          site,
		Board:         board,
		Title:         topic.Title,
		Summary:       topic.Summary,
		Content:       topic.Content,
		Author:        "DevHub 用户",
		Status:        status,
		Pinned:        topic.IsPinned,
		Recommended:   topic.IsFeatured,
		CommentLocked: topic.CommentLocked,
		Views:         topic.ViewCount,
		Likes:         topic.LikeCount,
		Comments:      topic.CommentCount,
		Tags:          append([]string{}, topic.Tags...),
		CreatedAt:     topic.CreatedAt,
		UpdatedAt:     topic.UpdatedAt,
	}
}

func currentDemoUserID() int64 { return 1 }

func communityIDBySlug(slug string) int64 {
	switch strings.TrimSpace(slug) {
	case "php":
		return 1
	case "go":
		return 2
	case "java":
		return 3
	case "ai":
		return 4
	case "frontend":
		return 5
	default:
		return 0
	}
}

func slugByCommunityID(id int64) string {
	switch id {
	case 1:
		return "php"
	case 2:
		return "go"
	case 3:
		return "java"
	case 4:
		return "ai"
	case 5:
		return "frontend"
	default:
		return "portal"
	}
}

func categoryIDByBoard(communityID int64, board string) int64 {
	order := map[string]int64{"community": 1, "qa": 2, "opensource": 3, "ai": 4, "jobs": 5, "wiki": 6, "docs": 7}
	if order[board] == 0 {
		board = "community"
	}
	return communityID*100 + order[board]
}

func adminContentTypeByBoard(board string) string {
	switch board {
	case "qa":
		return "question"
	case "opensource":
		return "project"
	case "ai":
		return "ai_work"
	case "jobs":
		return "job"
	case "wiki":
		return "wiki_page"
	case "docs":
		return "document"
	default:
		return "article"
	}
}

func boardByContentTypeHTTP(contentType string) string {
	switch pluginregistry.NormalizeContentType(contentType) {
	case "question":
		return "qa"
	case "project":
		return "opensource"
	case "ai_work":
		return "ai"
	case "job":
		return "jobs"
	case "wiki_page":
		return "wiki"
	case "document":
		return "docs"
	default:
		return "community"
	}
}

func postStatusToTopicStatus(status string) int {
	switch strings.TrimSpace(status) {
	case "offline", "hidden":
		return 0
	case "deleted":
		return 3
	default:
		return 1
	}
}

func uniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func uniqueInt64s(items []int64) []int64 {
	seen := map[int64]bool{}
	out := []int64{}
	for _, item := range items {
		if item <= 0 || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func (s *Server) normalizeCreateTopicRequest(req *domain.CreateTopicRequest) error {
	req.Title = strings.TrimSpace(req.Title)
	req.Summary = strings.TrimSpace(req.Summary)
	req.Content = strings.TrimSpace(req.Content)
	req.ContentType = pluginregistry.NormalizeContentType(req.ContentType)
	req.CommunitySlug = strings.TrimSpace(req.CommunitySlug)

	if req.CommunityID == 0 && req.CommunitySlug != "" {
		if comm, ok := s.svc.CommunityBySlug(req.CommunitySlug); ok {
			req.CommunityID = comm.ID
		}
	}
	if req.CommunityID == 0 {
		return fmt.Errorf("请选择子站")
	}
	if len([]rune(req.Title)) < 4 || len([]rune(req.Title)) > 120 {
		return fmt.Errorf("标题长度需为 4 到 120 个字符")
	}
	if len([]rune(req.Summary)) > 300 {
		return fmt.Errorf("摘要最多 300 个字符")
	}
	if len([]rune(req.Content)) < 10 {
		return fmt.Errorf("正文至少 10 个字符")
	}
	if len(req.TagIDs) > 5 || len(req.Tags) > 5 {
		return fmt.Errorf("最多选择 5 个标签")
	}

	categories := s.svc.Categories(req.CommunityID)
	var category *domain.Category
	for i := range categories {
		if categories[i].ID == req.CategoryID {
			category = &categories[i]
			break
		}
	}
	if category == nil {
		return fmt.Errorf("请选择当前子站下的板块")
	}
	if category.Status != 1 || !category.Postable {
		return fmt.Errorf("当前板块不可发布")
	}
	if req.ContentType == "" {
		req.ContentType = pluginregistry.NormalizeContentType(firstNonEmpty(category.ContentType, category.Type))
	}
	normalizedType, pluginCode, err := s.svc.ValidateTopicPluginAccess(req.CommunityID, req.CategoryID, req.ContentType)
	if err != nil {
		// Keep user-facing errors stable and explicit.
		if strings.Contains(err.Error(), "全局") {
			return fmt.Errorf("插件未启用，不能发布该类型内容")
		}
		return err
	}
	req.ContentType = normalizedType
	req.PluginCode = pluginCode

	if len(req.TagIDs) > 0 {
		req.Tags = req.Tags[:0]
		allowed := map[int64]string{}
		for _, tag := range s.communityTagsForCreate(req.CommunityID) {
			allowed[tag.ID] = tag.Name
		}
		for _, id := range req.TagIDs {
			name, ok := allowed[id]
			if !ok {
				return fmt.Errorf("标签不属于当前子站")
			}
			req.Tags = append(req.Tags, name)
		}
	}
	if len(req.Tags) > 0 {
		allowed := map[string]bool{}
		site := slugByCommunityID(req.CommunityID)
		for _, tag := range s.communityTagsForCreate(req.CommunityID) {
			allowed[tag.Name] = true
		}
		normalized := make([]string, 0, len(req.Tags))
		for _, name := range req.Tags {
			name = strings.TrimSpace(name)
			if resolved, ok := s.svc.ResolveTag(site, name); ok {
				name = resolved.Tag.Name
			}
			if !allowed[name] {
				return fmt.Errorf("标签不属于当前子站")
			}
			normalized = append(normalized, name)
		}
		req.Tags = uniqueStrings(normalized)
	}
	return nil
}

func (s *Server) communityTagsForCreate(communityID int64) []domain.Tag {
	slug := ""
	for _, comm := range s.svc.Communities() {
		if comm.ID == communityID {
			slug = comm.Slug
			break
		}
	}
	seen := map[string]bool{}
	tags := []domain.Tag{}
	topics, _ := s.svc.TopicsByFilter(communityID, 0, "", "hot", nil, "", 1, 500)
	for _, topic := range topics {
		for _, name := range topic.Tags {
			name = strings.TrimSpace(name)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			tags = append(tags, domain.Tag{ID: int64(len(tags) + 1), Name: name, Site: slug})
		}
	}
	stats := []domain.TagStat{}
	if slug != "" {
		stats = s.svc.TagStats(slug)
	}
	if len(stats) == 0 {
		stats = s.svc.TagStats(fmt.Sprintf("%d", communityID))
	}
	for _, stat := range stats {
		if stat.Name == "" || seen[stat.Name] {
			continue
		}
		seen[stat.Name] = true
		tags = append(tags, domain.Tag{ID: int64(len(tags) + 1), Name: stat.Name, Site: slug})
	}
	return tags
}

func validContentType(contentType string) bool {
	return pluginregistry.ValidContentType(contentType)
}

func (s *Server) categoryIDBySlug(communityID int64, slug string) int64 {
	slug = strings.TrimSpace(slug)
	if slug == "" || slug == "all" {
		return 0
	}
	if communityID == 0 {
		return 0
	}
	for _, category := range s.svc.Categories(communityID) {
		if category.Slug == slug || category.Type == contentTypeByBoard(slug) {
			return category.ID
		}
	}
	return 0
}

func contentTypeByBoard(board string) string {
	switch board {
	case "community", "article":
		return "article"
	case "qa", "question":
		return "question"
	case "opensource", "project":
		return "project"
	case "ai", "ai_work":
		return "ai_work"
	case "jobs", "job":
		return "job"
	case "wiki", "wiki_page":
		return "wiki_page"
	case "docs", "doc", "document":
		return "document"
	case "news":
		return "news"
	default:
		return ""
	}
}

func contentTypeLabel(contentType string) string {
	switch contentType {
	case "article":
		return "社区"
	case "question":
		return "问答中心"
	case "project":
		return "开源项目"
	case "ai_work":
		return "AI作品"
	case "job":
		return "招聘内推"
	case "wiki", "wiki_page":
		return "Wiki"
	case "doc", "document":
		return "文档"
	case "news":
		return "公告"
	default:
		return "内容"
	}
}

func (s *Server) visibleCommunityCategories(communityID int64) []domain.Category {
	categories := s.svc.Categories(communityID)
	out := make([]domain.Category, 0, len(categories))
	for _, category := range categories {
		if category.Status == 1 && category.Visible && category.NavVisible {
			out = append(out, category)
		}
	}
	return out
}

func filterTopicsByState(topics []domain.Topic, keep func(domain.Topic) bool) []domain.Topic {
	out := make([]domain.Topic, 0, len(topics))
	for _, topic := range topics {
		if keep(topic) {
			out = append(out, topic)
		}
	}
	return out
}

func (s *Server) communityHotTags(communityID int64, slug string) []domain.TagStat {
	topics, _ := s.svc.TopicsByFilter(communityID, 0, "", "hot", nil, "", 1, 200)
	counts := map[string]int{}
	for _, topic := range topics {
		for _, tag := range topic.Tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				counts[tag]++
			}
		}
	}
	tags := make([]domain.TagStat, 0, len(counts))
	for name, count := range counts {
		tags = append(tags, domain.TagStat{Name: name, Count: count})
	}
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Count == tags[j].Count {
			return tags[i].Name < tags[j].Name
		}
		return tags[i].Count > tags[j].Count
	})
	if len(tags) > 20 {
		tags = tags[:20]
	}
	if len(tags) == 0 {
		tags = s.svc.TagStats(slug)
	}
	return tags
}

func communityStatsHTML(stats domain.CommunityStats) string {
	items := []struct {
		Label string
		Value int
	}{
		{"主题", stats.TopicCount},
		{"评论", stats.CommentCount},
		{"问答", stats.QuestionCount},
		{"未解决", stats.UnsolvedCount},
		{"关注", stats.FollowerCount},
		{"今日主题", stats.TodayTopicCount},
		{"今日评论", stats.TodayCommentCount},
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf(`<span><strong>%d</strong>%s</span>`, item.Value, esc(item.Label)))
	}
	return strings.Join(parts, "")
}

func communityCategoryNavHTML(slug string, categories []domain.Category) string {
	links := []string{fmt.Sprintf(`<a class="active" href="/c/%s/">全部</a>`, pathEsc(slug))}
	for _, category := range categories {
		label := category.Name
		if label == "" {
			label = contentTypeLabel(firstNonEmpty(category.ContentType, category.Type))
		}
		links = append(links, fmt.Sprintf(`<a href="/c/%s/?category_slug=%s&amp;content_type=%s">%s</a>`, pathEsc(slug), queryEsc(category.Slug), queryEsc(firstNonEmpty(category.ContentType, category.Type)), esc(label)))
	}
	return strings.Join(links, "")
}

func communityTopicSectionHTML(title string, topics []domain.Topic, slug, emptyText string) string {
	var b strings.Builder
	b.WriteString(`<section><div class="section-head"><h2>`)
	b.WriteString(esc(title))
	b.WriteString(`</h2><a href="/search/?scope=community&amp;community_slug=`)
	b.WriteString(queryEsc(slug))
	b.WriteString(`">查看全部</a></div>`)
	if len(topics) == 0 {
		b.WriteString(`<div class="empty-state">`)
		b.WriteString(esc(emptyText))
		b.WriteString(`</div></section>`)
		return b.String()
	}
	b.WriteString(`<div class="post-list">`)
	for _, topic := range topics {
		b.WriteString(communityTopicCardHTML(topic, slug))
	}
	b.WriteString(`</div></section>`)
	return b.String()
}

func communityTopicCardHTML(topic domain.Topic, slug string) string {
	summary := firstNonEmpty(topic.Summary, firstRunes(topic.Content, 120))
	pills := []string{`<span class="type-pill">` + esc(contentTypeLabel(topic.ContentType)) + `</span>`}
	if topic.IsPinned {
		pills = append(pills, `<span class="state-pill">置顶</span>`)
	}
	if topic.IsFeatured {
		pills = append(pills, `<span class="state-pill">精华</span>`)
	}
	if topic.ContentType == "question" {
		if topic.IsSolved {
			pills = append(pills, `<span class="state-pill solved">已解决</span>`)
		} else {
			pills = append(pills, `<span class="state-pill unsolved">未解决</span>`)
		}
	}
	tagLinks := make([]string, 0, len(topic.Tags))
	for _, tag := range topic.Tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}
		tagLinks = append(tagLinks, fmt.Sprintf(`<a href="%s">%s</a>`, esc(tagHref(tag, slug)), esc(tag)))
	}
	return fmt.Sprintf(`<article class="post-card"><div class="post-card-top">%s<a class="site-pill" href="/c/%s/">%s</a></div><h2><a href="/topics/%d/">%s</a></h2><p>%s</p><div class="post-tags">%s</div><footer><span>%s 发布</span><span>%d 浏览</span><span>%d 评论</span><span>%d 赞</span></footer></article>`,
		strings.Join(pills, ""), pathEsc(slug), esc(slug), topic.ID, esc(topic.Title), esc(summary), strings.Join(tagLinks, ""), esc(topic.CreatedAt), topic.ViewCount, topic.CommentCount, topic.LikeCount)
}

func tagCommunityStatHTML(tag domain.Tag) string {
	if tag.CommunitySlug == "" {
		return ""
	}
	label := firstNonEmpty(tag.CommunityName, tag.CommunitySlug)
	return fmt.Sprintf(`<span><a href="/c/%s/">%s</a></span>`, pathEsc(tag.CommunitySlug), esc(label))
}

func tagTopicSectionHTML(title string, topics []domain.Topic, emptyText string) string {
	var b strings.Builder
	b.WriteString(`<section><div class="section-head"><h2>`)
	b.WriteString(esc(title))
	b.WriteString(`</h2></div>`)
	if len(topics) == 0 {
		b.WriteString(`<div class="empty-state">`)
		b.WriteString(esc(emptyText))
		b.WriteString(`</div></section>`)
		return b.String()
	}
	b.WriteString(`<div class="post-list">`)
	for _, topic := range topics {
		b.WriteString(communityTopicCardHTML(topic, slugByCommunityID(topic.CommunityID)))
	}
	b.WriteString(`</div></section>`)
	return b.String()
}

func relatedTagsHTML(tags []domain.TagStat, currentSlug string) string {
	links := make([]string, 0, len(tags))
	for _, tag := range tags {
		segment := firstNonEmpty(tag.Slug, tagPathSegment(tag.Name))
		if segment == "" || segment == currentSlug || strings.TrimSpace(tag.Name) == "" {
			continue
		}
		links = append(links, fmt.Sprintf(`<a href="%s">%s<span>%d</span></a>`, esc(tagHrefFromSegment(segment, tag.CommunitySlug)), esc(tag.Name), firstNonZeroHTTP(tag.TopicCount, tag.Count)))
	}
	if len(links) == 0 {
		return `<span>暂无相关标签</span>`
	}
	return strings.Join(links, "")
}

func communityAnnouncementHTML(comm domain.Community) string {
	if strings.TrimSpace(comm.AnnouncementTitle) == "" && strings.TrimSpace(comm.AnnouncementContent) == "" {
		return ""
	}
	link := ""
	if strings.TrimSpace(comm.AnnouncementURL) != "" {
		link = fmt.Sprintf(`<p><a href="%s">查看公告链接</a></p>`, esc(comm.AnnouncementURL))
	}
	return fmt.Sprintf(`<section class="community-rules"><h2>%s</h2><p>%s</p>%s</section>`, esc(firstNonEmpty(comm.AnnouncementTitle, "子站公告")), esc(comm.AnnouncementContent), link)
}

func communityTagsHTML(tags []domain.TagStat, slug string) string {
	if len(tags) == 0 {
		return `<span>暂无标签</span>`
	}
	links := make([]string, 0, len(tags))
	for _, tag := range tags {
		if strings.TrimSpace(tag.Name) == "" {
			continue
		}
		segment := firstNonEmpty(tag.Slug, tagPathSegment(tag.Name))
		links = append(links, fmt.Sprintf(`<a href="%s">%s<span>%d</span></a>`, esc(tagHrefFromSegment(segment, slug)), esc(tag.Name), firstNonZeroHTTP(tag.TopicCount, tag.Count)))
	}
	return strings.Join(links, "")
}

func communityModeratorsHTML(items []domain.CommunityModerator) string {
	if len(items) == 0 {
		return `<p>暂无公开版主。</p>`
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		name := firstNonEmpty(item.UserNickname, item.UserName, fmt.Sprintf("UID %d", item.UserID))
		role := "版主"
		if item.Role == "owner" {
			role = "站长"
		}
		parts = append(parts, fmt.Sprintf(`<p><strong>%s</strong><span> %s</span></p>`, esc(name), esc(role)))
	}
	return strings.Join(parts, "")
}

func esc(value string) string {
	return html.EscapeString(value)
}

func pathEsc(value string) string {
	return url.PathEscape(value)
}

func queryEsc(value string) string {
	return urlQueryEsc(value)
}

func batchAuditTarget(targetType, action string, updated, total int, note string) string {
	target := fmt.Sprintf("%s:%s:%d/%d", targetType, action, updated, total)
	note = strings.TrimSpace(note)
	if note == "" {
		return target
	}
	runes := []rune(note)
	if len(runes) > 80 {
		note = string(runes[:80])
	}
	return target + " " + note
}

func urlQueryEsc(value string) string {
	replacer := strings.NewReplacer("%", "%25", " ", "+", "&", "%26", "?", "%3F", "#", "%23", "=", "%3D", "/", "%2F")
	return replacer.Replace(value)
}

func tagPathSegment(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	slug := normalizeSlug(name)
	if slug != "" {
		return slug
	}
	return name
}

func tagHref(tagName, communitySlug string) string {
	return tagHrefFromSegment(tagPathSegment(tagName), communitySlug)
}

func tagCanonicalPath(tag domain.Tag) string {
	return tagHrefFromSegment(tagPathSegment(firstNonEmpty(tag.Slug, tag.Name)), tag.CommunitySlug)
}

func tagHrefFromSegment(segment, communitySlug string) string {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return "/tags/"
	}
	if communitySlug != "" && communitySlug != "portal" {
		return "/c/" + pathEsc(communitySlug) + "/tags/" + pathEsc(segment) + "/"
	}
	return "/tags/" + pathEsc(segment) + "/"
}

func firstNonZeroHTTP(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func normalizeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", ",", "-")
	value = strings.Trim(replacer.Replace(value), "-")
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			continue
		}
		if r == '-' {
			if lastDash {
				continue
			}
			lastDash = true
			b.WriteRune(r)
			continue
		}
		lastDash = false
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}

func tagToStat(tag domain.Tag) domain.TagStat {
	return domain.TagStat{
		ID:             tag.ID,
		Name:           tag.Name,
		Slug:           tag.Slug,
		Site:           tag.Site,
		CommunityID:    tag.CommunityID,
		CommunitySlug:  tag.CommunitySlug,
		Description:    tag.Description,
		TopicCount:     tag.TopicCount,
		Count:          firstNonZeroHTTP(tag.TopicCount, tag.UseCount),
		FollowerCount:  tag.FollowerCount,
		Status:         tag.Status,
		SEOTitle:       tag.SEOTitle,
		SEODescription: tag.SEODescription,
		SEOKeywords:    tag.SEOKeywords,
	}
}

func frontendStylesheetHref() string {
	const fallback = "/_astro/index.css"

	if htmlBytes, err := os.ReadFile("./web/frontend/index.html"); err == nil {
		html := string(htmlBytes)
		marker := `<link rel="stylesheet" href="`
		start := strings.Index(html, marker)
		if start >= 0 {
			start += len(marker)
			if end := strings.Index(html[start:], `"`); end > 0 {
				href := strings.TrimSpace(html[start : start+end])
				if strings.HasPrefix(href, "/_astro/") && strings.HasSuffix(href, ".css") {
					return href
				}
			}
		}
	}

	matches, err := filepath.Glob("./web/frontend/_astro/*.css")
	if err == nil && len(matches) > 0 {
		sort.Strings(matches)
		return "/_astro/" + filepath.Base(matches[0])
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstRunes(value string, n int) string {
	runes := []rune(strings.TrimSpace(value))
	if n <= 0 || len(runes) <= n {
		return string(runes)
	}
	return string(runes[:n]) + "..."
}

func ternary(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

func seoDescription(summary, content string) string {
	text := strings.TrimSpace(summary)
	if text == "" {
		text = strings.TrimSpace(content)
	}
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) > 155 {
		return string(runes[:155])
	}
	if text == "" {
		return "DevHub 技术社区内容详情。"
	}
	return text
}

func paragraphsHTML(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "<p>暂无正文。</p>"
	}
	parts := strings.Split(content, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, "<p>"+esc(part)+"</p>")
	}
	if len(out) == 0 {
		return "<p>" + esc(content) + "</p>"
	}
	return strings.Join(out, "")
}

func aiSummaryHTML(summary string) string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return ""
	}
	return `<section class="ai-summary"><h2>AI 摘要</h2><p>` + esc(summary) + `</p></section>`
}

func absoluteURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	}
	host := c.Request.Host
	if host == "" {
		host = "127.0.0.1:8090"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return scheme + "://" + host + path
}
