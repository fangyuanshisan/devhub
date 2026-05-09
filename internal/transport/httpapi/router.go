package httpapi

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
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
		api.GET("/stats", srv.stats)
		api.GET("/tags", srv.tags)
		api.GET("/tags/hot", srv.hotTags)
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
		api.GET("/communities/:slug/categories", srv.listCategories)
		api.GET("/communities/:slug/tags", srv.listCommunityTags)
		api.GET("/topics", srv.listTopics)
		api.GET("/topics/:id", srv.getTopic)
		api.GET("/topics/:id/comments", srv.topicComments)
		api.POST("/topics/:id/comments", srv.createTopicComment)
		api.POST("/topics/:id/comments/:commentId/replies", srv.replyTopicComment)
		api.POST("/topics/:id/comments/:commentId/accept", srv.acceptTopicComment)
		api.POST("/topics", srv.createTopic)
		api.PUT("/topics/:id", srv.authRequired(), srv.updateTopic)
		api.DELETE("/topics/:id", srv.authRequired(), srv.deleteTopic)
		api.POST("/topics/:id/like", srv.likeTopic)
		api.POST("/topics/:id/favorite", srv.favoriteTopic)
		api.GET("/topics/:id/interaction", srv.topicInteraction)
		api.POST("/topics/:id/solve", srv.authRequired(), srv.solveTopic)
		api.GET("/search/topics", srv.searchTopics)
		api.POST("/actions/toggle", srv.toggleReaction)
		api.POST("/reactions/toggle", srv.toggleReaction)
		api.POST("/favorites/toggle", srv.toggleFavorite)
		api.POST("/follows/toggle", srv.toggleFollow)
		api.GET("/activities", srv.userActivities)
		api.GET("/me/favorites", srv.myFavorites)
		api.GET("/me/follows", srv.myFollows)
		api.GET("/me/activities", srv.myActivities)
		api.GET("/me/notifications", srv.myNotifications)
		api.POST("/me/notifications/read-all", srv.readAllMyNotifications)
		api.POST("/me/notifications/:id/read", srv.readMyNotification)
		api.POST("/reports", srv.optionalAuth(), srv.createReport)

		admin := api.Group("/admin")
		{
			admin.POST("/login", srv.adminLogin)
			admin.POST("/refresh", srv.refreshSession)
			admin.POST("/logout", srv.logout)
			protected := admin.Group("", srv.authRequired(), srv.adminContext())
			protected.GET("/me", srv.adminMe)
			protected.GET("/overview", srv.requirePermission("dashboard.read"), srv.adminOverview)
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
			protected.POST("/tags", srv.requirePermission("post.update"), srv.createAdminTag)
			protected.PUT("/tags/:id", srv.requirePermission("post.update"), srv.updateAdminTag)
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
	r.StaticFile("/admin-next", "./web/admin-vue/index.html")
	r.StaticFile("/admin-next/", "./web/admin-vue/index.html")
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next")
	})
	r.GET("/admin/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/admin-next")
	})
	r.GET("/site/:site", func(c *gin.Context) {
		c.File(fmt.Sprintf("./web/frontend/site/%s/index.html", c.Param("site")))
	})
	r.GET("/site/:site/", func(c *gin.Context) {
		c.File(fmt.Sprintf("./web/frontend/site/%s/index.html", c.Param("site")))
	})
	r.GET("/c/:site", func(c *gin.Context) {
		c.File(fmt.Sprintf("./web/frontend/site/%s/index.html", c.Param("site")))
	})
	r.GET("/c/:site/", func(c *gin.Context) {
		c.File(fmt.Sprintf("./web/frontend/site/%s/index.html", c.Param("site")))
	})
	r.GET("/c/:site/topics/new", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/c/%s/topics/new/index.html", c.Param("site")))
	})
	r.GET("/c/:site/topics/new/", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/c/%s/topics/new/index.html", c.Param("site")))
	})
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
	r.GET("/tags/:tag", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/tags/%s/index.html", c.Param("tag")))
	})
	r.GET("/tags/:tag/", func(c *gin.Context) {
		serveFrontendFile(c, fmt.Sprintf("./web/frontend/tags/%s/index.html", c.Param("tag")))
	})
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

func (s *Server) redirectPostToTopic(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/topics/"+id+"/")
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
      <div class="header-actions"><a class="publish-link" href="/c/%s/topics/new/">发布内容</a><a href="/me/favorites">收藏</a><a href="/me/follows">关注</a><a href="/admin-next">后台</a></div>
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
	    const accessToken = () => localStorage.getItem('devhub_access_token') || '';
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
		links = append(links, fmt.Sprintf(`<a href="/search/?tag=%s&amp;scope=community&amp;community_slug=%s">%s</a>`, queryEsc(tag), queryEsc(communitySlug), esc(tag)))
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

func (s *Server) stats(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	if !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	c.JSON(http.StatusOK, s.svc.PostStats(site))
}

func (s *Server) tags(c *gin.Context) {
	site := c.DefaultQuery("site", "portal")
	if !s.svc.ValidateSite(site) {
		fail(c, http.StatusBadRequest, "无效子网站")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": s.svc.TagStats(site)})
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
	var req domain.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if strings.HasPrefix(c.FullPath(), "/api/v1/admin/") && !ensureSiteAllowed(c, req.Site) {
		return
	}
	p, err := s.svc.CreatePost(req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "operation", "创建帖子", fmt.Sprintf("posts#%d", p.ID))
	c.JSON(http.StatusCreated, p)
}

func (s *Server) updatePost(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if strings.HasPrefix(c.FullPath(), "/api/v1/admin/") {
		post, exists := s.svc.GetPost(id, false)
		if !exists {
			fail(c, http.StatusNotFound, "帖子不存在")
			return
		}
		if !ensureSiteAllowed(c, post.Site) {
			return
		}
	}
	var req domain.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if strings.HasPrefix(c.FullPath(), "/api/v1/admin/") && req.Site != nil && !ensureSiteAllowed(c, *req.Site) {
		return
	}
	p, err := s.svc.UpdatePost(id, req)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.audit(c, "operation", "更新帖子", fmt.Sprintf("posts#%d", p.ID))
	c.JSON(http.StatusOK, p)
}

func (s *Server) deletePost(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if strings.HasPrefix(c.FullPath(), "/api/v1/admin/") {
		post, exists := s.svc.GetPost(id, false)
		if !exists {
			fail(c, http.StatusNotFound, "帖子不存在")
			return
		}
		if !ensureSiteAllowed(c, post.Site) {
			return
		}
	}
	if !s.svc.DeletePost(id) {
		fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	s.audit(c, "operation", "删除帖子", fmt.Sprintf("posts#%d", id))
	c.JSON(http.StatusOK, gin.H{"deleted": true})
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
	s.adminLogin(c)
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

func (s *Server) authMe(c *gin.Context) {
	user, _ := currentUser(c)
	c.JSON(http.StatusOK, user)
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
	if !s.svc.ValidateSite(site) || !s.svc.ValidateBoard(board) {
		fail(c, http.StatusBadRequest, "筛选参数不合法")
		return
	}
	posts := s.svc.AdminTopics(site, board, q)
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
	topicReq.UserID = currentUserID(c)
	if !ensureSiteAllowed(c, req.Site) {
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
		Action:      strings.TrimSpace(c.Query("action")),
		Target:      strings.TrimSpace(c.Query("target")),
		TargetType:  strings.TrimSpace(c.Query("target_type")),
		Actor:       strings.TrimSpace(c.Query("actor")),
		ActorID:     int64Query(c, "actor_user_id", 0),
		CommunityID: int64Query(c, "community_id", 0),
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
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		return
	}
	actor := adminCtx.CurrentUser.Username
	if actor == "" {
		actor = adminCtx.CurrentUser.Nickname
	}
	site := adminCtx.CurrentSite
	if site == "" {
		site = adminSiteScope(c)
	}
	s.svc.AppendAdminLog(domain.AdminLog{
		Site:   site,
		Type:   logType,
		Actor:  actor,
		Role:   adminCtx.CurrentUser.RoleCode,
		Action: action,
		Target: target,
		IP:     c.ClientIP(),
	})
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
	if adminCtx.CurrentSite != "" && adminCtx.CurrentSite != "portal" {
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
		return currentUserID(c) == 1
	}
	return user.ID == 1 || user.RoleCode == "super_admin" || hasPermission(user.Permissions, "*")
}

func (s *Server) canModerateCommunity(c *gin.Context, communityID int64) bool {
	if isAdminUser(c) {
		return true
	}
	if communityID <= 0 {
		fail(c, http.StatusForbidden, "只有管理员可以管理全局举报")
		return false
	}
	if s.svc.IsCommunityModerator(currentUserID(c), communityID) {
		return true
	}
	fail(c, http.StatusForbidden, "无权管理该子站内容")
	return false
}

func (s *Server) canModerateCommunityForBatch(c *gin.Context, communityID int64) bool {
	if isAdminUser(c) {
		return true
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

func (s *Server) reportFilter(c *gin.Context) domain.ReportFilter {
	page, pageSize := pagination(c)
	communityID := int64(0)
	communitySlug := strings.TrimSpace(firstQuery(c, "community_slug", "site"))
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
		ActorIsAdmin:  isAdminUser(c),
	}
}

func (s *Server) moderatorFilter(c *gin.Context) domain.CommunityModeratorFilter {
	page, pageSize := pagination(c)
	communityID := int64(0)
	communitySlug := strings.TrimSpace(firstQuery(c, "community_slug", "site"))
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
		ActorIsAdmin:  isAdminUser(c),
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
	s.audit(c, "audit", map[string]string{"feature": "切换精华", "pin": "切换置顶"}[action], fmt.Sprintf("topics#%d", id))
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
	c.JSON(http.StatusOK, gin.H{"comment": updated, "changed": true})
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
	c.JSON(http.StatusOK, gin.H{"items": categories})
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

	contentType := firstQuery(c, "content_type", "type")
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

	topic, err := s.svc.UpdateTopic(id, req)
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

	contentType := strings.TrimSpace(firstQuery(c, "content_type", "type"))
	if !validContentType(contentType) {
		contentType = ""
	}
	sortBy := strings.TrimSpace(c.DefaultQuery("sort", "latest"))
	switch sortBy {
	case "latest", "active", "hot", "featured", "unsolved":
	default:
		sortBy = "latest"
	}

	return domain.SearchRequest{
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
		communityID := communityIDBySlug(strings.TrimSpace(*req.Site))
		if communityID <= 0 {
			return out, fmt.Errorf("子站不存在")
		}
		out.CommunityID = &communityID
	}
	if req.Board != nil {
		board := strings.TrimSpace(*req.Board)
		if board == "" || board == "all" {
			return out, fmt.Errorf("请选择具体板块")
		}
		communityID := topic.CommunityID
		if out.CommunityID != nil {
			communityID = *out.CommunityID
		}
		categoryID := categoryIDByBoard(communityID, board)
		contentType := adminContentTypeByBoard(board)
		out.CategoryID = &categoryID
		out.ContentType = &contentType
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
		return "wiki"
	case "docs":
		return "doc"
	default:
		return "article"
	}
}

func boardByContentTypeHTTP(contentType string) string {
	switch contentType {
	case "question":
		return "qa"
	case "project":
		return "opensource"
	case "ai_work":
		return "ai"
	case "job":
		return "jobs"
	case "wiki":
		return "wiki"
	case "doc":
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
	req.ContentType = strings.TrimSpace(req.ContentType)
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
	if req.ContentType == "" {
		req.ContentType = category.Type
	}
	if category.Type != "" && req.ContentType != category.Type {
		return fmt.Errorf("内容类型与板块不匹配")
	}
	if !validContentType(req.ContentType) {
		return fmt.Errorf("内容类型不合法")
	}

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
		for _, tag := range s.communityTagsForCreate(req.CommunityID) {
			allowed[tag.Name] = true
		}
		for _, name := range req.Tags {
			if !allowed[name] {
				return fmt.Errorf("标签不属于当前子站")
			}
		}
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
	switch contentType {
	case "article", "question", "project", "ai_work", "job", "wiki", "doc", "news":
		return true
	default:
		return false
	}
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
	case "wiki":
		return "wiki"
	case "docs", "doc":
		return "doc"
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
	case "wiki":
		return "Wiki"
	case "doc":
		return "文档"
	case "news":
		return "公告"
	default:
		return "内容"
	}
}

func esc(value string) string {
	return html.EscapeString(value)
}

func pathEsc(value string) string {
	return strings.ReplaceAll(urlQueryEsc(value), "+", "%20")
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
			return value
		}
	}
	return ""
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
