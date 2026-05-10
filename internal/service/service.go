package service

import (
	"errors"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

// Repository 定义业务服务依赖的数据访问能力。
type Repository interface {
	Health() domain.HealthStatus
	ValidateSite(site string) bool
	ValidateBoard(board string) bool
	AdminLogin(account, password string) (*domain.AdminSession, error)
	UserLogin(account, password string) (*domain.AdminSession, error)
	Register(req domain.RegisterRequest) (*domain.AdminSession, error)
	RefreshSession(refreshToken string) (*domain.AdminSession, error)
	RefreshAdminSession(refreshToken string) (*domain.AdminSession, error)
	Logout(refreshToken string) error
	AuthUser(accessToken string) (*domain.AuthUser, error)
	AuthAdmin(accessToken string) (*domain.AuthUser, error)
	GetSite(key string) (domain.Site, bool)
	CreateSite(req domain.Site) (domain.Site, error)
	UpdateSite(key string, req domain.Site) (domain.Site, bool)
	ListSites() []domain.Site
	ListBoards() []domain.Board
	CreateBoard(req domain.Board) (domain.Board, error)
	UpdateBoard(key string, req domain.Board) (domain.Board, bool)
	Plugins() []domain.Plugin
	PluginByCode(code string) (domain.Plugin, bool)
	SetPluginStatus(code, status string) (domain.Plugin, error)
	CommunityPlugins(communityID int64) ([]domain.Plugin, error)
	SetCommunityPluginStatus(communityID int64, code, status string) (domain.Plugin, error)
	SetCommunityPluginConfig(communityID int64, code, configJSON string) (domain.Plugin, error)
	ReorderCommunityPlugins(communityID int64, codes []string) (int, error)
	ListPosts(site, board, q, tag string) []domain.Post
	GetPost(id int64, increaseView bool) (*domain.Post, bool)
	CreatePost(req domain.CreatePostRequest) (*domain.Post, error)
	UpdatePost(id int64, req domain.UpdatePostRequest) (*domain.Post, error)
	DeletePost(id int64) bool
	LikePost(id int64) (*domain.Post, error)
	HotPosts(site string, limit int) []domain.Post
	Feed(site string, limit int) []domain.Post
	TagStats(site string) []domain.TagStat
	TagBySlug(site, slugOrName string) (domain.Tag, bool)
	ResolveTag(site, slugOrName string) (domain.TagResolveResult, bool)
	TagTopics(tagID int64, communityID int64, contentType string, sort string, page, pageSize int) ([]domain.Topic, int)
	TagSuggestions(site, q string, limit int) []domain.TagStat
	AdminTags(site, q, status string) []domain.Tag
	AdminTagByID(id int64) (domain.Tag, bool)
	AdminTagTopics(id int64, page, pageSize int) ([]domain.Topic, int)
	CreateTag(req domain.Tag) (domain.Tag, error)
	UpdateTag(id int64, req domain.Tag) (domain.Tag, bool)
	SetTagStatus(id int64, status string) (domain.Tag, bool)
	TagAliases(tagID int64) ([]domain.TagAlias, error)
	AddTagAlias(tagID int64, alias string) (domain.TagAlias, error)
	DeleteTagAlias(tagID, aliasID int64) error
	MergeTag(sourceTagID, targetTagID int64) (domain.Tag, error)
	RecalculateTag(tagID int64) (domain.Tag, error)
	RecalculateAllTags() (int, error)
	BoardCounts(site, q string) map[string]int
	PostStats(site string) domain.PostStats
	CommentsTree(postID int64) []*domain.Comment
	LikeComment(id int64) (*domain.Comment, error)
	DeleteOwnComment(id int64, author string) error
	DeleteComment(id int64) bool
	AdminOverview(site string) domain.AdminOverview
	AdminUsers() []domain.AdminUser
	UpdateUserStatus(id int64, status, note string) bool
	AdminRoles() []domain.AdminRole
	AdminPermissions() []domain.AdminPermission
	AdminComments(site string) []domain.AdminComment
	AdminTopics(site, board, q string) []domain.Post
	UpdateCommentStatus(id int64, status string) bool
	AdminSettings() domain.AdminSettings
	UpdateAdminSettings(req domain.AdminSettings) domain.AdminSettings
	AdminLogs(site string) []domain.AdminLog
	AdminLogsByFilter(filter domain.AdminLogFilter) ([]domain.AdminLog, int)
	AppendAdminLog(log domain.AdminLog)
	PushNotification(req domain.PushNotificationRequest) *domain.Notification
	Notices(site string) []domain.Notification
	ReadNotice(id int64) bool
	ReadAllNotices(site string) int
	UnreadNoticeCount(site string) int
	UserProfile() domain.UserProfile

	// ===== 新增：DevHub 通用社区系统 =====
	Communities() []domain.Community
	CommunityBySlug(slug string) (domain.Community, bool)
	Categories(communityID int64) []domain.Category
	CommunityStats(communityID int64) domain.CommunityStats
	CreateCommunity(req domain.CommunityRequest) (domain.Community, error)
	UpdateCommunity(id int64, req domain.CommunityRequest) (domain.Community, error)
	SetCommunityStatus(id int64, status int) (domain.Community, error)
	ReorderCommunities(ids []int64) int
	CreateCategory(communityID int64, req domain.CategoryRequest) (domain.Category, error)
	UpdateCategory(id int64, req domain.CategoryRequest) (domain.Category, error)
	SetCategoryStatus(id int64, status int) (domain.Category, error)
	ReorderCategories(ids []int64) int
	TopicsByFilter(communityID, categoryID int64, contentType, sort string, isSolved *bool, tag string, page, pageSize int) ([]domain.Topic, int)
	TopicByID(id int64, increaseView bool) (*domain.Topic, error)
	CreateTopic(req domain.CreateTopicRequest) (*domain.Topic, error)
	UpdateTopic(id int64, req domain.UpdateTopicRequest) (*domain.Topic, error)
	DeleteTopic(id int64) bool
	SearchTopics(req domain.SearchRequest) ([]domain.Topic, int)
	ToggleReaction(userID int64, targetID int64, targetType, reactionType string) (bool, int, error)
	ToggleFavorite(userID int64, targetID int64, targetType string) (bool, error)
	ToggleFollow(userID int64, targetID int64, targetType string) (bool, error)
	TopicInteraction(userID int64, topicID int64) (domain.TopicInteraction, error)
	UserFavorites(userID int64, targetType string, page, pageSize int) ([]domain.FavoriteItem, int)
	UserFollows(userID int64, targetType string, page, pageSize int) ([]domain.FollowItem, int)
	UserActivities(userID int64, communityID int64, action string, page, pageSize int) ([]domain.Activity, int)
	UserNotifications(userID int64, isRead *bool, page, pageSize int) ([]domain.Notification, int, int)
	ReadUserNotification(userID int64, id int64) bool
	ReadAllUserNotifications(userID int64) int
	CommunityOverview(slug string) (domain.CommunityOverview, bool)
	TopicComments(topicID int64, sort string, page, pageSize int) ([]*domain.Comment, int)
	CommentByID(id int64) (*domain.Comment, error)
	CreateCommentWithRequest(topicID int64, req domain.CreateCommentRequest) (*domain.Comment, error)
	AcceptBestAnswer(topicID int64, commentID int64, actorUserID int64) bool
	CreateReport(req domain.CreateReportRequest) (*domain.Report, error)
	Reports(filter domain.ReportFilter) ([]domain.Report, int)
	ReportByID(id int64) (*domain.Report, error)
	HandleReport(id int64, status, note string, handlerUserID int64) (*domain.Report, error)
	IsCommunityModerator(userID, communityID int64) bool
	CommunityModerators(filter domain.CommunityModeratorFilter) ([]domain.CommunityModerator, int)
	CommunityModeratorByID(id int64) (*domain.CommunityModerator, error)
	CreateCommunityModerator(req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error)
	UpdateCommunityModerator(id int64, req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error)
	DeleteCommunityModerator(id int64) bool
	SetTopicFeatured(id int64, featured bool) (*domain.Topic, error)
	SetTopicPinned(id int64, pinned bool) (*domain.Topic, error)
	SetTopicStatus(id int64, status int) (*domain.Topic, error)
	SetTopicCommentLocked(id int64, locked bool) (*domain.Topic, error)
	SetCommentStatus(id int64, status string) (*domain.Comment, error)
}

// Service 封装业务入口，向 HTTP 层提供稳定的调用接口。
type Service struct {
	repo Repository
}

// New 创建业务服务实例。
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Health 返回服务及当前数据源状态。
func (s *Service) Health() domain.HealthStatus { return s.repo.Health() }

// ValidateSite 校验站点筛选值是否可用。
func (s *Service) ValidateSite(site string) bool { return s.repo.ValidateSite(site) }

// ValidateBoard 校验板块筛选值是否可用。
func (s *Service) ValidateBoard(board string) bool { return s.repo.ValidateBoard(board) }

// AdminLogin 校验后台账号并返回轻量会话。
func (s *Service) AdminLogin(account, password string) (*domain.AdminSession, error) {
	return s.repo.AdminLogin(account, password)
}

// UserLogin 校验前台用户账号并返回前台登录态。
func (s *Service) UserLogin(account, password string) (*domain.AdminSession, error) {
	return s.repo.UserLogin(account, password)
}

// Register 创建前台用户并返回登录态。
func (s *Service) Register(req domain.RegisterRequest) (*domain.AdminSession, error) {
	return s.repo.Register(req)
}

// RefreshSession 使用 refresh token 轮转登录态。
func (s *Service) RefreshSession(refreshToken string) (*domain.AdminSession, error) {
	return s.repo.RefreshSession(refreshToken)
}

// RefreshAdminSession 使用后台 refresh token 轮转后台登录态。
func (s *Service) RefreshAdminSession(refreshToken string) (*domain.AdminSession, error) {
	return s.repo.RefreshAdminSession(refreshToken)
}

// Logout 撤销 refresh token。
func (s *Service) Logout(refreshToken string) error { return s.repo.Logout(refreshToken) }

// AuthUser 解析 access token 并返回当前用户。
func (s *Service) AuthUser(accessToken string) (*domain.AuthUser, error) {
	return s.repo.AuthUser(accessToken)
}

// AuthAdmin 解析后台 access token 并返回当前后台人员上下文。
func (s *Service) AuthAdmin(accessToken string) (*domain.AuthUser, error) {
	return s.repo.AuthAdmin(accessToken)
}

// GetSite 按 key 获取站点配置。
func (s *Service) GetSite(key string) (domain.Site, bool) { return s.repo.GetSite(key) }

// CreateSite 创建子站配置。
func (s *Service) CreateSite(req domain.Site) (domain.Site, error) { return s.repo.CreateSite(req) }

// UpdateSite 更新站点配置。
func (s *Service) UpdateSite(key string, req domain.Site) (domain.Site, bool) {
	return s.repo.UpdateSite(key, req)
}

// ListSites 返回所有站点配置。
func (s *Service) ListSites() []domain.Site { return s.repo.ListSites() }

// ListBoards 返回所有板块配置。
func (s *Service) ListBoards() []domain.Board { return s.repo.ListBoards() }

// CreateBoard 创建板块配置。
func (s *Service) CreateBoard(req domain.Board) (domain.Board, error) {
	return s.repo.CreateBoard(req)
}

// UpdateBoard 更新板块配置。
func (s *Service) UpdateBoard(key string, req domain.Board) (domain.Board, bool) {
	return s.repo.UpdateBoard(key, req)
}

// Plugins 返回系统插件注册信息与运行状态。
func (s *Service) Plugins() []domain.Plugin { return s.repo.Plugins() }

// PluginByCode 按插件唯一标识获取插件。
func (s *Service) PluginByCode(code string) (domain.Plugin, bool) {
	return s.repo.PluginByCode(code)
}

// CommunityPlugins returns plugin list with community runtime state overlay.
func (s *Service) CommunityPlugins(communityID int64) ([]domain.Plugin, error) {
	return s.repo.CommunityPlugins(communityID)
}

// SetCommunityPluginStatus updates per-community plugin enablement.
func (s *Service) SetCommunityPluginStatus(communityID int64, code, status string) (domain.Plugin, error) {
	return s.repo.SetCommunityPluginStatus(communityID, code, status)
}

// SetCommunityPluginConfig updates per-community plugin config blob.
func (s *Service) SetCommunityPluginConfig(communityID int64, code, configJSON string) (domain.Plugin, error) {
	return s.repo.SetCommunityPluginConfig(communityID, code, configJSON)
}

// ReorderCommunityPlugins updates per-community plugin sort order.
func (s *Service) ReorderCommunityPlugins(communityID int64, codes []string) (int, error) {
	return s.repo.ReorderCommunityPlugins(communityID, codes)
}

// IsPluginEnabled checks whether a plugin is globally enabled.
// Core is always enabled and not persisted in plugins table.
func (s *Service) IsPluginEnabled(pluginCode string) bool {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return true
	}
	plugin, ok := s.repo.PluginByCode(pluginCode)
	return ok && plugin.Status == pluginregistry.StatusEnabled
}

// IsPluginEnabledForCommunity checks whether a plugin is enabled for a community,
// requiring both global and community status enabled.
func (s *Service) IsPluginEnabledForCommunity(communityID int64, pluginCode string) bool {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return true
	}
	items, err := s.repo.CommunityPlugins(communityID)
	if err != nil {
		return false
	}
	for _, item := range items {
		if item.Code == pluginCode {
			return item.Status == pluginregistry.StatusEnabled
		}
	}
	return false
}

// ListEnabledPluginsForCommunity returns plugins enabled for a community.
func (s *Service) ListEnabledPluginsForCommunity(communityID int64) ([]domain.Plugin, error) {
	items, err := s.repo.CommunityPlugins(communityID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Plugin, 0, len(items))
	for _, item := range items {
		if item.Status == pluginregistry.StatusEnabled {
			out = append(out, item)
		}
	}
	return out, nil
}

// ValidateTopicPluginAccess validates plugin enablement and category bindings for publishing.
// It returns the normalized contentType and resolved pluginCode.
func (s *Service) ValidateTopicPluginAccess(communityID, categoryID int64, contentType string) (string, string, error) {
	contentType = pluginregistry.NormalizeContentType(strings.TrimSpace(contentType))
	if contentType == "" {
		return "", "", errors.New("内容类型不能为空")
	}
	if !pluginregistry.ValidContentType(contentType) {
		return "", "", errors.New("内容类型不合法")
	}
	pluginCode := pluginregistry.PluginCodeForContentType(contentType)
	if pluginCode != pluginregistry.CoreCode {
		if !s.IsPluginEnabled(pluginCode) {
			return "", "", errors.New("插件全局未启用")
		}
		if !s.IsPluginEnabledForCommunity(communityID, pluginCode) {
			return "", "", errors.New("当前子站未启用该插件")
		}
	}
	if categoryID > 0 {
		categories := s.repo.Categories(communityID)
		var category *domain.Category
		for i := range categories {
			if categories[i].ID == categoryID {
				category = &categories[i]
				break
			}
		}
		if category == nil {
			return "", "", errors.New("板块不存在")
		}
		categoryType := pluginregistry.NormalizeContentType(firstNonEmpty(category.ContentType, category.Type))
		expectedPlugin := firstNonEmpty(category.PluginCode, pluginregistry.PluginCodeForContentType(categoryType))
		if expectedPlugin != pluginCode {
			return "", "", errors.New("当前板块未绑定对应插件")
		}
		if !pluginregistry.ContentTypeAllowed(category.AllowedContentTypes, contentType) {
			return "", "", errors.New("内容类型与板块不匹配")
		}
	}
	return contentType, pluginCode, nil
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func requiredCreatePermission(contentType, pluginCode string) string {
	contentType = strings.TrimSpace(contentType)
	pluginCode = strings.TrimSpace(pluginCode)
	switch contentType {
	case "question":
		return "qa.question.create"
	case "document":
		return "docs.document.create"
	case "wiki_page":
		return "wiki.page.create"
	}
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return "post.create"
	}
	return ""
}

func hasPermission(perms []string, permission string) bool {
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return true
	}
	for _, p := range perms {
		if p == "*" || p == permission {
			return true
		}
	}
	return false
}

// SetPluginStatus 更新插件状态。
func (s *Service) SetPluginStatus(code, status string) (domain.Plugin, error) {
	return s.repo.SetPluginStatus(code, status)
}

// ListPosts 按站点、板块、关键词和标签筛选帖子列表。
func (s *Service) ListPosts(site, board, q, tag string) []domain.Post {
	return s.repo.ListPosts(site, board, q, tag)
}

// GetPost 获取帖子详情，increaseView 为 true 时同步增加浏览量。
func (s *Service) GetPost(id int64, increaseView bool) (*domain.Post, bool) {
	return s.repo.GetPost(id, increaseView)
}

// CreatePost 创建帖子。
func (s *Service) CreatePost(req domain.CreatePostRequest) (*domain.Post, error) {
	return s.repo.CreatePost(req)
}

// UpdatePost 更新帖子。
func (s *Service) UpdatePost(id int64, req domain.UpdatePostRequest) (*domain.Post, error) {
	return s.repo.UpdatePost(id, req)
}

// DeletePost 删除帖子及其评论。
func (s *Service) DeletePost(id int64) bool { return s.repo.DeletePost(id) }

// LikePost 给帖子点赞。
func (s *Service) LikePost(id int64) (*domain.Post, error) { return s.repo.LikePost(id) }

// HotPosts 返回按热度排序的帖子。
func (s *Service) HotPosts(site string, limit int) []domain.Post {
	return s.repo.HotPosts(site, limit)
}

// Feed 返回按发布时间排序的信息流。
func (s *Service) Feed(site string, limit int) []domain.Post { return s.repo.Feed(site, limit) }

// TagStats 返回站点下的标签统计。
func (s *Service) TagStats(site string) []domain.TagStat { return s.repo.TagStats(site) }

// TagBySlug 按 slug 或名称获取启用标签。
func (s *Service) TagBySlug(site, slugOrName string) (domain.Tag, bool) {
	return s.repo.TagBySlug(site, slugOrName)
}

// ResolveTag 按 slug、别名或合并关系解析标签。
func (s *Service) ResolveTag(site, slugOrName string) (domain.TagResolveResult, bool) {
	return s.repo.ResolveTag(site, slugOrName)
}

// TagTopics 返回标签关联内容。
func (s *Service) TagTopics(tagID int64, communityID int64, contentType string, sort string, page, pageSize int) ([]domain.Topic, int) {
	return s.repo.TagTopics(tagID, communityID, contentType, sort, page, pageSize)
}

// TagSuggestions 返回发布页标签建议。
func (s *Service) TagSuggestions(site, q string, limit int) []domain.TagStat {
	return s.repo.TagSuggestions(site, q, limit)
}

// AdminTags 返回后台标签管理列表。
func (s *Service) AdminTags(site, q, status string) []domain.Tag {
	return s.repo.AdminTags(site, q, status)
}

// AdminTagByID 返回后台标签详情。
func (s *Service) AdminTagByID(id int64) (domain.Tag, bool) { return s.repo.AdminTagByID(id) }

// AdminTagTopics 返回后台标签关联内容。
func (s *Service) AdminTagTopics(id int64, page, pageSize int) ([]domain.Topic, int) {
	return s.repo.AdminTagTopics(id, page, pageSize)
}

// CreateTag 创建标签配置。
func (s *Service) CreateTag(req domain.Tag) (domain.Tag, error) { return s.repo.CreateTag(req) }

// UpdateTag 更新标签配置。
func (s *Service) UpdateTag(id int64, req domain.Tag) (domain.Tag, bool) {
	return s.repo.UpdateTag(id, req)
}

// TagAliases 返回标签别名列表。
func (s *Service) TagAliases(tagID int64) ([]domain.TagAlias, error) { return s.repo.TagAliases(tagID) }

// AddTagAlias 新增标签别名。
func (s *Service) AddTagAlias(tagID int64, alias string) (domain.TagAlias, error) {
	return s.repo.AddTagAlias(tagID, alias)
}

// DeleteTagAlias 删除标签别名。
func (s *Service) DeleteTagAlias(tagID, aliasID int64) error {
	return s.repo.DeleteTagAlias(tagID, aliasID)
}

// MergeTag 合并标签。
func (s *Service) MergeTag(sourceTagID, targetTagID int64) (domain.Tag, error) {
	return s.repo.MergeTag(sourceTagID, targetTagID)
}

// RecalculateTag 重算单个标签统计。
func (s *Service) RecalculateTag(tagID int64) (domain.Tag, error) {
	return s.repo.RecalculateTag(tagID)
}

// RecalculateAllTags 重算全部标签统计。
func (s *Service) RecalculateAllTags() (int, error) { return s.repo.RecalculateAllTags() }

// SetTagStatus 启用或禁用标签。
func (s *Service) SetTagStatus(id int64, status string) (domain.Tag, bool) {
	return s.repo.SetTagStatus(id, status)
}

// BoardCounts 返回站点内各板块的帖子数量。
func (s *Service) BoardCounts(site, q string) map[string]int {
	return s.repo.BoardCounts(site, q)
}

// PostStats 返回站点内容汇总统计。
func (s *Service) PostStats(site string) domain.PostStats { return s.repo.PostStats(site) }

// SiteOverview 聚合子站首页所需的数据。
func (s *Service) SiteOverview(site string, limit int) (domain.SiteOverview, bool) {
	meta, ok := s.repo.GetSite(site)
	if !ok {
		return domain.SiteOverview{}, false
	}
	return domain.SiteOverview{
		Site:        meta,
		Boards:      s.repo.ListBoards(),
		BoardCounts: s.repo.BoardCounts(site, ""),
		Stats:       s.repo.PostStats(site),
		Tags:        s.repo.TagStats(site),
		HotPosts:    s.repo.HotPosts(site, limit),
		LatestPosts: s.repo.Feed(site, limit),
	}, true
}

// CommentsTree 返回帖子的树形评论列表。
func (s *Service) CommentsTree(postID int64) []*domain.Comment {
	return s.repo.CommentsTree(postID)
}

// LikeComment 给评论点赞。
func (s *Service) LikeComment(id int64) (*domain.Comment, error) {
	return s.repo.LikeComment(id)
}

// DeleteOwnComment 仅允许评论作者删除自己的评论及其子回复。
func (s *Service) DeleteOwnComment(id int64, author string) error {
	return s.repo.DeleteOwnComment(id, author)
}

// DeleteComment 删除评论及其子回复。
func (s *Service) DeleteComment(id int64) bool { return s.repo.DeleteComment(id) }

// AdminOverview 返回后台首页统计。
func (s *Service) AdminOverview(site string) domain.AdminOverview { return s.repo.AdminOverview(site) }

// AdminUsers 返回后台用户列表。
func (s *Service) AdminUsers() []domain.AdminUser { return s.repo.AdminUsers() }

// UpdateUserStatus 更新后台用户状态和备注。
func (s *Service) UpdateUserStatus(id int64, status, note string) bool {
	return s.repo.UpdateUserStatus(id, status, note)
}

// AdminRoles 返回后台角色列表。
func (s *Service) AdminRoles() []domain.AdminRole { return s.repo.AdminRoles() }

// AdminPermissions 返回后台权限点列表。
func (s *Service) AdminPermissions() []domain.AdminPermission {
	return s.repo.AdminPermissions()
}

// AdminComments 返回后台评论审核列表。
func (s *Service) AdminComments(site string) []domain.AdminComment { return s.repo.AdminComments(site) }

// AdminTopics 返回后台内容列表，包含隐藏内容。
func (s *Service) AdminTopics(site, board, q string) []domain.Post {
	return s.repo.AdminTopics(site, board, q)
}

// UpdateCommentStatus 更新评论审核状态。
func (s *Service) UpdateCommentStatus(id int64, status string) bool {
	return s.repo.UpdateCommentStatus(id, status)
}

// AdminSettings 返回后台基础参数。
func (s *Service) AdminSettings() domain.AdminSettings { return s.repo.AdminSettings() }

// UpdateAdminSettings 更新后台基础参数。
func (s *Service) UpdateAdminSettings(req domain.AdminSettings) domain.AdminSettings {
	return s.repo.UpdateAdminSettings(req)
}

// AdminLogs 返回后台操作日志。
func (s *Service) AdminLogs(site string) []domain.AdminLog { return s.repo.AdminLogs(site) }

// AdminLogsByFilter 返回可筛选和分页的后台治理审计日志。
func (s *Service) AdminLogsByFilter(filter domain.AdminLogFilter) ([]domain.AdminLog, int) {
	return s.repo.AdminLogsByFilter(filter)
}

// AppendAdminLog 写入带站点和操作者上下文的后台操作日志。
func (s *Service) AppendAdminLog(log domain.AdminLog) { s.repo.AppendAdminLog(log) }

// PushNotification 创建一条站内通知。
func (s *Service) PushNotification(req domain.PushNotificationRequest) *domain.Notification {
	return s.repo.PushNotification(req)
}

// Notices 返回当前用户的通知列表。
func (s *Service) Notices(site string) []domain.Notification { return s.repo.Notices(site) }

// ReadNotice 将单条通知标记为已读。
func (s *Service) ReadNotice(id int64) bool { return s.repo.ReadNotice(id) }

// ReadAllNotices 将所有未读通知标记为已读，并返回更新数量。
func (s *Service) ReadAllNotices(site string) int { return s.repo.ReadAllNotices(site) }

// UnreadNoticeCount 返回未读通知数量。
func (s *Service) UnreadNoticeCount(site string) int { return s.repo.UnreadNoticeCount(site) }

// UserProfile 返回当前用户资料。
func (s *Service) UserProfile() domain.UserProfile { return s.repo.UserProfile() }

// ===== 新增：DevHub 通用社区系统服务 =====

// Communities 返回所有子站列表。
func (s *Service) Communities() []domain.Community { return s.repo.Communities() }

// CommunityBySlug 按 slug 获取子站。
func (s *Service) CommunityBySlug(slug string) (domain.Community, bool) {
	return s.repo.CommunityBySlug(slug)
}

// Categories 返回指定子站的板块列表。
func (s *Service) Categories(communityID int64) []domain.Category {
	return s.repo.Categories(communityID)
}

// CommunityStats 返回子站统计。
func (s *Service) CommunityStats(communityID int64) domain.CommunityStats {
	return s.repo.CommunityStats(communityID)
}

// CreateCommunity 新增子站。
func (s *Service) CreateCommunity(req domain.CommunityRequest) (domain.Community, error) {
	return s.repo.CreateCommunity(req)
}

// UpdateCommunity 更新子站。
func (s *Service) UpdateCommunity(id int64, req domain.CommunityRequest) (domain.Community, error) {
	return s.repo.UpdateCommunity(id, req)
}

// SetCommunityStatus 设置子站状态。
func (s *Service) SetCommunityStatus(id int64, status int) (domain.Community, error) {
	return s.repo.SetCommunityStatus(id, status)
}

// ReorderCommunities 更新子站排序。
func (s *Service) ReorderCommunities(ids []int64) int { return s.repo.ReorderCommunities(ids) }

// CreateCategory 新增板块。
func (s *Service) CreateCategory(communityID int64, req domain.CategoryRequest) (domain.Category, error) {
	return s.repo.CreateCategory(communityID, req)
}

// UpdateCategory 更新板块。
func (s *Service) UpdateCategory(id int64, req domain.CategoryRequest) (domain.Category, error) {
	return s.repo.UpdateCategory(id, req)
}

// SetCategoryStatus 设置板块状态。
func (s *Service) SetCategoryStatus(id int64, status int) (domain.Category, error) {
	return s.repo.SetCategoryStatus(id, status)
}

// ReorderCategories 更新板块排序。
func (s *Service) ReorderCategories(ids []int64) int { return s.repo.ReorderCategories(ids) }

// TopicsByFilter 按筛选条件返回主题列表。
func (s *Service) TopicsByFilter(communityID, categoryID int64, contentType, sort string, isSolved *bool, tag string, page, pageSize int) ([]domain.Topic, int) {
	return s.repo.TopicsByFilter(communityID, categoryID, contentType, sort, isSolved, tag, page, pageSize)
}

// TopicByID 获取主题详情。
func (s *Service) TopicByID(id int64, increaseView bool) (*domain.Topic, error) {
	return s.repo.TopicByID(id, increaseView)
}

// CreateTopic 创建主题。
func (s *Service) CreateTopic(req domain.CreateTopicRequest) (*domain.Topic, error) {
	normalizedType, pluginCode, err := s.ValidateTopicPluginAccess(req.CommunityID, req.CategoryID, req.ContentType)
	if err != nil {
		return nil, err
	}
	req.ContentType = normalizedType
	req.PluginCode = pluginCode

	if perm := requiredCreatePermission(normalizedType, pluginCode); perm != "" {
		if !hasPermission(req.ActorPermissions, perm) {
			return nil, errors.New("无权发布该类型内容")
		}
	}
	return s.repo.CreateTopic(req)
}

// UpdateTopic 更新主题。
func (s *Service) UpdateTopic(id int64, req domain.UpdateTopicRequest) (*domain.Topic, error) {
	return s.repo.UpdateTopic(id, req)
}

// DeleteTopic 删除主题。
func (s *Service) DeleteTopic(id int64) bool {
	return s.repo.DeleteTopic(id)
}

// SearchTopics 搜索主题。
func (s *Service) SearchTopics(req domain.SearchRequest) ([]domain.Topic, int) {
	return s.repo.SearchTopics(req)
}

// ToggleReaction 切换点赞。
func (s *Service) ToggleReaction(userID int64, targetID int64, targetType, reactionType string) (bool, int, error) {
	return s.repo.ToggleReaction(userID, targetID, targetType, reactionType)
}

// ToggleFavorite 切换收藏。
func (s *Service) ToggleFavorite(userID int64, targetID int64, targetType string) (bool, error) {
	return s.repo.ToggleFavorite(userID, targetID, targetType)
}

// ToggleFollow 切换关注。
func (s *Service) ToggleFollow(userID int64, targetID int64, targetType string) (bool, error) {
	return s.repo.ToggleFollow(userID, targetID, targetType)
}

// TopicInteraction 获取当前用户对主题的互动状态。
func (s *Service) TopicInteraction(userID int64, topicID int64) (domain.TopicInteraction, error) {
	return s.repo.TopicInteraction(userID, topicID)
}

// UserFavorites 获取用户收藏列表。
func (s *Service) UserFavorites(userID int64, targetType string, page, pageSize int) ([]domain.FavoriteItem, int) {
	return s.repo.UserFavorites(userID, targetType, page, pageSize)
}

// UserFollows 获取用户关注列表。
func (s *Service) UserFollows(userID int64, targetType string, page, pageSize int) ([]domain.FollowItem, int) {
	return s.repo.UserFollows(userID, targetType, page, pageSize)
}

// UserActivities 获取用户动态。
func (s *Service) UserActivities(userID int64, communityID int64, action string, page, pageSize int) ([]domain.Activity, int) {
	return s.repo.UserActivities(userID, communityID, action, page, pageSize)
}

// UserNotifications 获取用户通知列表、总数和未读数。
func (s *Service) UserNotifications(userID int64, isRead *bool, page, pageSize int) ([]domain.Notification, int, int) {
	return s.repo.UserNotifications(userID, isRead, page, pageSize)
}

// ReadUserNotification 将用户单条通知标记为已读。
func (s *Service) ReadUserNotification(userID int64, id int64) bool {
	return s.repo.ReadUserNotification(userID, id)
}

// ReadAllUserNotifications 将用户所有通知标记为已读。
func (s *Service) ReadAllUserNotifications(userID int64) int {
	return s.repo.ReadAllUserNotifications(userID)
}

// CommunityOverview 获取子站概览。
func (s *Service) CommunityOverview(slug string) (domain.CommunityOverview, bool) {
	return s.repo.CommunityOverview(slug)
}

// TopicComments 获取主题评论。
func (s *Service) TopicComments(topicID int64, sort string, page, pageSize int) ([]*domain.Comment, int) {
	return s.repo.TopicComments(topicID, sort, page, pageSize)
}

// CommentByID 获取评论详情。
func (s *Service) CommentByID(id int64) (*domain.Comment, error) { return s.repo.CommentByID(id) }

// AcceptBestAnswer 采纳最佳答案（问答类型）。
func (s *Service) AcceptBestAnswer(topicID int64, commentID int64, actorUserID int64) bool {
	return s.repo.AcceptBestAnswer(topicID, commentID, actorUserID)
}

// CreateReport 创建举报记录。
func (s *Service) CreateReport(req domain.CreateReportRequest) (*domain.Report, error) {
	return s.repo.CreateReport(req)
}

// Reports 返回后台举报列表。
func (s *Service) Reports(filter domain.ReportFilter) ([]domain.Report, int) {
	return s.repo.Reports(filter)
}

// ReportByID 返回举报详情。
func (s *Service) ReportByID(id int64) (*domain.Report, error) { return s.repo.ReportByID(id) }

// HandleReport 处理举报。
func (s *Service) HandleReport(id int64, status, note string, handlerUserID int64) (*domain.Report, error) {
	return s.repo.HandleReport(id, status, note, handlerUserID)
}

// IsCommunityModerator 判断用户是否为指定子站版主。
func (s *Service) IsCommunityModerator(userID, communityID int64) bool {
	return s.repo.IsCommunityModerator(userID, communityID)
}

// CommunityModerators 返回后台版主列表。
func (s *Service) CommunityModerators(filter domain.CommunityModeratorFilter) ([]domain.CommunityModerator, int) {
	return s.repo.CommunityModerators(filter)
}

// CommunityModeratorByID 返回版主详情。
func (s *Service) CommunityModeratorByID(id int64) (*domain.CommunityModerator, error) {
	return s.repo.CommunityModeratorByID(id)
}

// CreateCommunityModerator 新增子站版主。
func (s *Service) CreateCommunityModerator(req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error) {
	return s.repo.CreateCommunityModerator(req)
}

// UpdateCommunityModerator 更新子站版主。
func (s *Service) UpdateCommunityModerator(id int64, req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error) {
	return s.repo.UpdateCommunityModerator(id, req)
}

// DeleteCommunityModerator 停用子站版主。
func (s *Service) DeleteCommunityModerator(id int64) bool {
	return s.repo.DeleteCommunityModerator(id)
}

func (s *Service) SetTopicFeatured(id int64, featured bool) (*domain.Topic, error) {
	return s.repo.SetTopicFeatured(id, featured)
}

func (s *Service) SetTopicPinned(id int64, pinned bool) (*domain.Topic, error) {
	return s.repo.SetTopicPinned(id, pinned)
}

func (s *Service) SetTopicStatus(id int64, status int) (*domain.Topic, error) {
	return s.repo.SetTopicStatus(id, status)
}

func (s *Service) SetTopicCommentLocked(id int64, locked bool) (*domain.Topic, error) {
	return s.repo.SetTopicCommentLocked(id, locked)
}

func (s *Service) SetCommentStatus(id int64, status string) (*domain.Comment, error) {
	return s.repo.SetCommentStatus(id, status)
}

// CreateComment 创建评论（新的Topics API）。
func (s *Service) CreateComment(topicID int64, author string, text string, parentID int64) (*domain.Comment, error) {
	return s.repo.CreateCommentWithRequest(topicID, domain.CreateCommentRequest{
		Author:   author,
		Text:     text,
		ParentID: parentID,
	})
}

// CreateCommentWithRequest 创建评论。
func (s *Service) CreateCommentWithRequest(topicID int64, req domain.CreateCommentRequest) (*domain.Comment, error) {
	return s.repo.CreateCommentWithRequest(topicID, req)
}
