package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

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
	SavePlugin(plugin domain.Plugin) (domain.Plugin, error)
	SetPluginStatus(code, status string) (domain.Plugin, error)
	SetPluginConfig(code, configJSON string) (domain.Plugin, error)
	PluginImpact(code string) (domain.PluginImpact, error)
	CommunityPluginImpact(communityID int64, code string) (domain.PluginImpact, error)
	PluginMigrations(pluginCode string) ([]domain.PluginMigration, error)
	AppendPluginMigration(record domain.PluginMigration) (domain.PluginMigration, error)
	SavePluginMigration(record domain.PluginMigration) (domain.PluginMigration, error)
	// Plugin approvals (v1.5.0-P1-07).
	AppendPluginApprovalRequest(record domain.PluginApprovalRequest) (domain.PluginApprovalRequest, error)
	SavePluginApprovalRequest(record domain.PluginApprovalRequest) (domain.PluginApprovalRequest, error)
	PluginApprovalRequests(filter domain.PluginApprovalFilter) ([]domain.PluginApprovalRequest, int, error)
	PluginApprovalRequestByID(id int64) (domain.PluginApprovalRequest, bool)
	// Plugin package uploads (v1.6.0-P0-02).
	AppendPluginPackageUpload(record domain.PluginPackageUploadRecord) (domain.PluginPackageUploadRecord, error)
	SavePluginPackageUpload(record domain.PluginPackageUploadRecord) (domain.PluginPackageUploadRecord, error)
	PluginPackageUploadByUploadID(uploadID string) (domain.PluginPackageUploadRecord, bool)
	PluginPackageUploads(filter domain.PluginPackageUploadFilter) ([]domain.PluginPackageUploadRecord, int, error)
	// Plugin operations (v1.6.0-P0-06).
	AppendPluginOperationSnapshot(record domain.PluginOperationSnapshot) (domain.PluginOperationSnapshot, error)
	SavePluginOperationSnapshot(record domain.PluginOperationSnapshot) (domain.PluginOperationSnapshot, error)
	PluginOperationSnapshotByOperationID(operationID string) (domain.PluginOperationSnapshot, bool)
	PluginOperationSnapshots(filter domain.PluginOperationFilter) ([]domain.PluginOperationSnapshot, int, error)
	DeletePluginByCode(code string) error
	DeleteCommunityPluginsByCode(code string) (int, error)
	DeletePluginMigrationsByPlugin(code string) (int, error)
	DeletePluginMigrationsByVersion(code, version string) (int, error)
	DeletePluginConfigVersionsByPlugin(code string) (int, error)
	// Trusted publishers (v1.6.0-P0-03).
	AppendPluginTrustedPublisher(record domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error)
	SavePluginTrustedPublisher(record domain.PluginTrustedPublisher) (domain.PluginTrustedPublisher, error)
	DeletePluginTrustedPublisher(id int64) error
	PluginTrustedPublisherByID(id int64) (domain.PluginTrustedPublisher, bool)
	PluginTrustedPublisherByKey(publisherID, publicKeyID string) (domain.PluginTrustedPublisher, bool)
	PluginTrustedPublishers(filter domain.PluginTrustedPublisherFilter) ([]domain.PluginTrustedPublisher, int, error)
	// Remote plugin indexes (v1.6.0-P0-04).
	AppendPluginRemoteIndex(record domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error)
	SavePluginRemoteIndex(record domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error)
	DeletePluginRemoteIndex(id int64) error
	PluginRemoteIndexByID(id int64) (domain.PluginRemoteIndexSource, bool)
	PluginRemoteIndexBySourceID(sourceID string) (domain.PluginRemoteIndexSource, bool)
	PluginRemoteIndexes(filter domain.PluginRemoteIndexFilter) ([]domain.PluginRemoteIndexSource, int, error)
	AppendHookExecution(record domain.HookExecution) (domain.HookExecution, error)
	HookExecutions(pluginCode string, limit int) ([]domain.HookExecution, error)
	HookStats(pluginCode string) ([]domain.HookStats, error)
	HookExecutionsByFilter(filter domain.HookExecutionFilter) ([]domain.HookExecution, int, error)
	// PluginReadiness is computed in service; no repo method.
	CommunityPlugins(communityID int64) ([]domain.Plugin, error)
	SetCommunityPluginStatus(communityID int64, code, status string) (domain.Plugin, error)
	SetCommunityPluginConfig(communityID int64, code, configJSON string) (domain.Plugin, error)
	ReorderCommunityPlugins(communityID int64, codes []string) (int, error)
	QAQuestionByTopicID(topicID int64) (*domain.QAQuestion, error)
	QAAnswersByTopicID(topicID int64) ([]domain.QAAnswer, error)
	DocsDocumentByTopicID(topicID int64) (*domain.DocsDocument, error)
	DocsTree(communityID int64, spaceID int64) ([]domain.DocsDocument, error)
	WikiPageByTopicID(topicID int64) (*domain.WikiPage, error)
	WikiVersionsByTopicID(topicID int64) ([]domain.WikiRevision, error)
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

	// ===== Plugin config versions (v1.5.0-P1-05) =====
	AppendPluginConfigVersion(record domain.PluginConfigVersion) (domain.PluginConfigVersion, error)
	PluginConfigVersions(pluginCode, scope string, communityID int64, page, pageSize int) ([]domain.PluginConfigVersion, int, error)
	PluginConfigVersionByID(id int64) (domain.PluginConfigVersion, bool)

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
	repo  Repository
	hooks *pluginregistry.HookBus
}

// New 创建业务服务实例。
func New(repo Repository) *Service {
	bus := pluginregistry.NewHookBus()
	pluginregistry.RegisterBuiltinHookHandlers(bus)
	return &Service{repo: repo, hooks: bus}
}

// DispatchHook exposes HookBus dispatch for platform-governance actions (enable/disable/config/etc).
// Keep this minimal: callers should only send non-business platform events.
func (s *Service) DispatchHook(event pluginregistry.HookEvent) error {
	return s.dispatchHook(event)
}

// RegisterHookHandler registers a built-in hook handler. This is intended for
// tests and compile-time system plugins, not third-party dynamic loading.
func (s *Service) RegisterHookHandler(name, pluginCode string, handler pluginregistry.HookHandler) {
	if s == nil || s.hooks == nil {
		return
	}
	s.hooks.Register(name, pluginCode, handler)
}

// SetHookFailureInjectionForTest configures a test/dev-only HookBus failure rule.
func (s *Service) SetHookFailureInjectionForTest(pluginCode, hookName string, mode pluginregistry.HookMode, errorMessage string) error {
	if s == nil || s.hooks == nil {
		return nil
	}
	pluginCode = strings.TrimSpace(pluginCode)
	hookName = strings.TrimSpace(hookName)
	if pluginCode == "" || hookName == "" {
		return errors.New("plugin_code 和 hook_name 不能为空")
	}
	if _, ok := s.repo.PluginByCode(pluginCode); !ok {
		return pluginNotFound(pluginCode)
	}
	if mode != "" && mode != pluginregistry.HookBlocking && mode != pluginregistry.HookNonBlocking {
		return errors.New("hook mode 不合法")
	}
	s.hooks.SetFailureInjection(pluginregistry.HookFailureRule{
		PluginCode: pluginCode,
		HookName:   hookName,
		Mode:       mode,
		Error:      strings.TrimSpace(errorMessage),
	})
	return nil
}

func (s *Service) dispatchHook(event pluginregistry.HookEvent) error {
	if s == nil || s.hooks == nil {
		return nil
	}
	event.Ctx.HookName = event.Name
	if event.Ctx.CategoryID == 0 && event.Ctx.ChannelID > 0 {
		event.Ctx.CategoryID = event.Ctx.ChannelID
	}
	if event.Ctx.ChannelID == 0 && event.Ctx.CategoryID > 0 {
		event.Ctx.ChannelID = event.Ctx.CategoryID
	}
	if event.Ctx.UserID == 0 && event.Ctx.Actor.UserID > 0 {
		event.Ctx.UserID = event.Ctx.Actor.UserID
	}
	if event.Ctx.AdminUserID == 0 && event.Ctx.Actor.AdminID > 0 {
		event.Ctx.AdminUserID = event.Ctx.Actor.AdminID
	}
	if event.Ctx.ActorID == 0 {
		event.Ctx.ActorID = actorIDFromHookContext(event.Ctx)
	}
	if event.Ctx.ActorType == "" {
		event.Ctx.ActorType = actorTypeFromActor(event.Ctx.Actor)
	}

	if !s.shouldRunHook(event) {
		return nil
	}

	results, err := s.hooks.DispatchWithResults(event)
	for _, result := range results {
		record := hookExecutionRecord(event, result)
		if saved, saveErr := s.repo.AppendHookExecution(record); saveErr == nil {
			record = saved
		}
		if !record.Success {
			s.auditHookFailure(record)
		}
	}
	if event.Mode == pluginregistry.HookBlocking {
		return err
	}
	return nil
}

func (s *Service) shouldRunHook(event pluginregistry.HookEvent) bool {
	code := strings.TrimSpace(event.Ctx.PluginCode)
	if code == "" || code == pluginregistry.CoreCode {
		return false
	}
	// Lifecycle hooks are emitted exactly when status changes, so they must run
	// even for the disabled transition.
	if event.Name == pluginregistry.HookAfterPluginEnabled || event.Name == pluginregistry.HookAfterPluginDisabled {
		return true
	}
	if !s.IsPluginEnabled(code) {
		return false
	}
	if event.Ctx.CommunityID > 0 && !s.IsPluginEnabledForCommunity(event.Ctx.CommunityID, code) {
		return false
	}
	return true
}

func hookExecutionRecord(event pluginregistry.HookEvent, result pluginregistry.HookExecutionResult) domain.HookExecution {
	ctx := event.Ctx
	meta := map[string]any{
		"handler_index": result.HandlerIndex,
	}
	for k, v := range ctx.Metadata {
		meta[k] = v
	}
	metadata := ""
	if raw, err := json.Marshal(meta); err == nil && string(raw) != "null" {
		metadata = string(raw)
	}
	return domain.HookExecution{
		HookName:     result.HookName,
		PluginCode:   result.PluginCode,
		Mode:         string(result.Mode),
		ContentType:  ctx.ContentType,
		ContentID:    ctx.ContentID,
		CommunityID:  ctx.CommunityID,
		CategoryID:   ctx.CategoryID,
		ActorType:    string(ctx.ActorType),
		ActorID:      ctx.ActorID,
		UserID:       ctx.UserID,
		AdminUserID:  ctx.AdminUserID,
		RequestID:    ctx.RequestID,
		StartedAt:    formatHookTime(result.StartedAt),
		FinishedAt:   formatHookTime(result.FinishedAt),
		DurationMS:   result.DurationMS,
		Success:      result.Success,
		ErrorMessage: result.ErrorMessage,
		Blocking:     result.Blocking,
		Metadata:     metadata,
	}
}

func (s *Service) auditHookFailure(record domain.HookExecution) {
	action := "plugin.hook.failed"
	if record.Blocking {
		action = "plugin.hook.blocked"
	}
	metadata, _ := json.Marshal(map[string]any{
		"plugin_code":   record.PluginCode,
		"hook_name":     record.HookName,
		"mode":          record.Mode,
		"blocking":      record.Blocking,
		"content_type":  record.ContentType,
		"content_id":    record.ContentID,
		"community_id":  record.CommunityID,
		"category_id":   record.CategoryID,
		"request_id":    record.RequestID,
		"error":         record.ErrorMessage,
		"duration_ms":   record.DurationMS,
		"hook_metadata": record.Metadata,
	})
	site := "portal"
	if record.CommunityID > 0 {
		site = fmt.Sprintf("community:%d", record.CommunityID)
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      site,
		Type:      "system",
		Actor:     firstNonEmpty(record.ActorType, "system"),
		ActorType: firstNonEmpty(record.ActorType, "system"),
		ActorID:   record.ActorID,
		Action:    action,
		Target:    fmt.Sprintf("hooks#%s:%s", record.PluginCode, record.HookName),
		OldValue:  "",
		NewValue:  "",
		Metadata:  string(metadata),
	})
}

func formatHookTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func actorIDFromHookContext(ctx pluginregistry.HookContext) int64 {
	if ctx.ActorID > 0 {
		return ctx.ActorID
	}
	if ctx.AdminUserID > 0 {
		return ctx.AdminUserID
	}
	if ctx.UserID > 0 {
		return ctx.UserID
	}
	return actorIDFromActor(ctx.Actor)
}

// ActorContextFromAuthUser converts server-authenticated identity into a trusted actor context.
func ActorContextFromAuthUser(user domain.AuthUser, communityScopes []int64) domain.ActorContext {
	ctx := domain.ActorContext{
		UserID:          user.ID,
		IsAdmin:         user.TokenType == "admin",
		IsModerator:     user.IsModerator || user.RoleCode == "moderator",
		CommunityScopes: uniquePositiveInt64s(communityScopes),
		Sites:           append([]string{}, user.Sites...),
		Permissions:     append([]string{}, user.Permissions...),
		TokenType:       user.TokenType,
		RoleCode:        user.RoleCode,
	}
	if ctx.IsAdmin {
		ctx.AdminID = user.ID
		ctx.UserID = 0
	}
	return ctx
}

func uniquePositiveInt64s(items []int64) []int64 {
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

func (s *Service) QAQuestionByTopicID(topicID int64) (*domain.QAQuestion, error) {
	return s.repo.QAQuestionByTopicID(topicID)
}

func (s *Service) QAAnswersByTopicID(topicID int64) ([]domain.QAAnswer, error) {
	return s.repo.QAAnswersByTopicID(topicID)
}

func (s *Service) DocsDocumentByTopicID(topicID int64) (*domain.DocsDocument, error) {
	return s.repo.DocsDocumentByTopicID(topicID)
}

func (s *Service) DocsTree(communityID int64, spaceID int64) ([]domain.DocsDocument, error) {
	return s.repo.DocsTree(communityID, spaceID)
}

func (s *Service) WikiPageByTopicID(topicID int64) (*domain.WikiPage, error) {
	return s.repo.WikiPageByTopicID(topicID)
}

func (s *Service) WikiVersionsByTopicID(topicID int64) ([]domain.WikiRevision, error) {
	return s.repo.WikiVersionsByTopicID(topicID)
}

// ValidateTopicPluginAccess validates plugin enablement and category bindings for publishing.
// It returns the normalized contentType and resolved pluginCode.
func (s *Service) ValidateTopicPluginAccess(communityID, categoryID int64, contentType string) (string, string, error) {
	contentType = pluginregistry.NormalizeContentType(strings.TrimSpace(contentType))
	if contentType == "" {
		return "", "", errors.New("内容类型不能为空")
	}
	def, ok := s.contentTypeDefinitionByType(contentType)
	if !ok {
		return "", "", errors.New("内容类型不合法")
	}
	pluginCode := firstNonEmpty(def.PluginCode, pluginregistry.PluginCodeForContentType(contentType))
	if pluginCode != pluginregistry.CoreCode {
		plugin, ok := s.PluginByCode(pluginCode)
		if !ok || plugin.Code == "" {
			return "", "", errors.New("内容类型对应插件不存在")
		}
		if !s.IsPluginEnabled(pluginCode) {
			return "", "", fmt.Errorf("插件未启用：插件 %s 当前状态为 %s，不能创建 %s 内容", pluginCode, firstNonEmpty(plugin.Status, "unknown"), contentType)
		}
		if !s.IsPluginEnabledForCommunity(communityID, pluginCode) {
			return "", "", fmt.Errorf("当前子站未启用插件 %s，不能创建 %s 内容", pluginCode, contentType)
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
		expectedPlugin := firstNonEmpty(category.PluginCode, s.pluginCodeForContentType(categoryType))
		if expectedPlugin != pluginCode {
			return "", "", errors.New("当前板块未绑定对应插件")
		}
		if !pluginregistry.ContentTypeAllowed(category.AllowedContentTypes, contentType) {
			return "", "", errors.New("内容类型与板块不匹配")
		}
	}
	return contentType, pluginCode, nil
}

func (s *Service) contentTypeDefinitionByType(contentType string) (domain.ContentTypeDefinition, bool) {
	contentType = pluginregistry.NormalizeContentType(strings.TrimSpace(contentType))
	if contentType == "" {
		return domain.ContentTypeDefinition{}, false
	}
	if def, ok := pluginregistry.ContentTypeDefinitionByType(contentType); ok {
		return def, true
	}
	for _, plugin := range s.repo.Plugins() {
		for _, def := range plugin.ContentTypeDefs {
			if pluginregistry.NormalizeContentType(def.Type) == contentType {
				if def.PluginCode == "" {
					def.PluginCode = plugin.Code
				}
				return def, true
			}
			for _, alias := range def.Aliases {
				if pluginregistry.NormalizeContentType(alias) == contentType {
					if def.PluginCode == "" {
						def.PluginCode = plugin.Code
					}
					def.Type = pluginregistry.NormalizeContentType(def.Type)
					return def, true
				}
			}
		}
		for _, typ := range plugin.ContentTypes {
			if pluginregistry.NormalizeContentType(typ) == contentType {
				return domain.ContentTypeDefinition{
					Type:             contentType,
					Name:             contentType,
					PluginCode:       plugin.Code,
					CreatePermission: firstNonEmpty(permissionByPrefix(plugin.Permissions, plugin.Code, "create"), plugin.Code+"."+contentType+".create"),
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
				}, true
			}
		}
	}
	return domain.ContentTypeDefinition{}, false
}

func permissionByPrefix(perms []domain.PermissionDefinition, pluginCode, suffix string) string {
	for _, perm := range perms {
		code := strings.TrimSpace(perm.Code)
		if code == "" {
			continue
		}
		if strings.HasPrefix(code, pluginCode+".") && strings.HasSuffix(code, "."+suffix) {
			return code
		}
	}
	return ""
}

func (s *Service) pluginCodeForContentType(contentType string) string {
	if def, ok := s.contentTypeDefinitionByType(contentType); ok {
		return firstNonEmpty(def.PluginCode, pluginregistry.CoreCode)
	}
	return pluginregistry.PluginCodeForContentType(contentType)
}

func (s *Service) createPermissionForContentType(contentType, pluginCode string) string {
	if def, ok := s.contentTypeDefinitionByType(contentType); ok && strings.TrimSpace(def.CreatePermission) != "" {
		return strings.TrimSpace(def.CreatePermission)
	}
	return requiredCreatePermission(contentType, pluginCode)
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
	contentType = pluginregistry.NormalizeContentType(strings.TrimSpace(contentType))
	pluginCode = strings.TrimSpace(pluginCode)
	if def, ok := pluginregistry.ContentTypeDefinitionByType(contentType); ok {
		return strings.TrimSpace(def.CreatePermission)
	}
	if pluginCode == "" || pluginCode == pluginregistry.CoreCode {
		return "core.topic.create"
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
	// Compatibility: historically core publishing uses post.create.
	// v1.3.x introduces core.topic.create, but we keep old roles working.
	if permission == "core.topic.create" {
		for _, p := range perms {
			if p == "post.create" {
				return true
			}
		}
	}
	return false
}

// HasPermission applies service-level permission compatibility rules.
func HasPermission(perms []string, permission string) bool {
	return hasPermission(perms, permission)
}

// CreatePermissionForContentType resolves the create permission for a content type.
func CreatePermissionForContentType(contentType, pluginCode string) string {
	return requiredCreatePermission(contentType, pluginCode)
}

// RequireCreatePermission checks whether an actor can create a given content type.
func RequireCreatePermission(perms []string, contentType, pluginCode string) error {
	permission := requiredCreatePermission(contentType, pluginCode)
	if permission == "" || hasPermission(perms, permission) {
		return nil
	}
	return fmt.Errorf("缺少权限 %s，不能创建该类型内容", permission)
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// ValidatePluginManifestJSON validates an external manifest/config-only plugin
// package without installing it or executing any third-party code.
func (s *Service) ValidatePluginManifestJSON(raw []byte) (domain.PluginManifestValidationResult, error) {
	result := pluginregistry.ValidatePluginManifestJSON(raw, s.repo.Plugins(), currentCoreVersion())
	return result, nil
}

// InstallPluginManifest installs the safe manifest + configuration plugin
// metadata. It never executes plugin code or raw SQL.
func (s *Service) InstallPluginManifest(raw []byte) (domain.Plugin, domain.PluginManifestValidationResult, error) {
	return s.installPluginManifestInternal(raw, "", "")
}

func (s *Service) installPluginManifestInternal(raw []byte, sourceType string, packageManifestChecksum string) (domain.Plugin, domain.PluginManifestValidationResult, error) {
	result, err := s.ValidatePluginManifestJSON(raw)
	if err != nil {
		return domain.Plugin{}, result, err
	}
	if !result.Valid {
		return domain.Plugin{}, result, domain.NewPluginError(PluginErrManifestInvalid, "manifest 校验失败").
			WithStatus(400).
			WithDetail("errors", result.Errors).
			WithSuggestion("请根据 errors 修复 manifest 后重试。")
	}
	manifest := result.NormalizedManifest
	manifest.IsSystem = false
	manifest.Status = pluginregistry.StatusDisabled
	manifest.SourceType = firstNonBlank(strings.TrimSpace(sourceType), manifest.SourceType, "manifest")
	manifestJSON, _ := json.Marshal(manifest)
	plugin := domain.Plugin{
		PluginManifest:        manifest,
		Status:                pluginregistry.StatusDisabled,
		SourceType:            manifest.SourceType,
		ManifestJSON:          string(manifestJSON),
		ManifestChecksum:      result.Checksum,
		PackageChecksum:       strings.TrimSpace(packageManifestChecksum),
		CompatibleCoreVersion: firstNonBlank(manifest.CompatibleCoreVersion, manifest.MinCoreVersion),
	}
	saved, err := s.repo.SavePlugin(plugin)
	if err != nil {
		return domain.Plugin{}, result, err
	}
	for _, migration := range manifest.Migrations {
		record := migrationRecordFromDefinition(migration, "pending")
		_, _ = s.repo.SavePluginMigration(record)
	}
	return saved, result, nil
}

func (s *Service) PluginUpgradeDryRun(code string, raw []byte) (domain.PluginUpgradeDryRunResult, error) {
	current, ok := s.repo.PluginByCode(code)
	if !ok {
		return domain.PluginUpgradeDryRunResult{}, pluginNotFound(code)
	}
	manifest, checksum, err := pluginregistry.DecodePluginManifestJSON(raw)
	if err != nil {
		return domain.PluginUpgradeDryRunResult{}, domain.NewPluginError(PluginErrManifestInvalid, "manifest 校验失败").
			WithStatus(400).
			WithDetail("plugin_code", strings.TrimSpace(code)).
			WithDetail("reason", strings.TrimSpace(err.Error())).
			WithSuggestion("请修复 manifest JSON 后重试。")
	}
	existing := make([]domain.Plugin, 0, len(s.repo.Plugins()))
	for _, item := range s.repo.Plugins() {
		if item.Code == code {
			continue
		}
		existing = append(existing, item)
	}
	validation := pluginregistry.ValidatePluginManifest(manifest, existing, currentCoreVersion())
	validation.Checksum = checksum
	normalizedManifest := validation.NormalizedManifest
	if normalizedManifest.Code != code {
		validation.Valid = false
		validation.Errors = append(validation.Errors, fmt.Sprintf("升级预览的 manifest code 必须为 %s", code))
	}
	compatibility := validation.Compatibility.Status
	if compatibility == pluginregistry.CompatibilityWarning {
		compatibility = pluginregistry.CompatibilityCompatible
	}
	if validation.Valid {
		if pluginregistry.CompareVersionStrings(normalizedManifest.Version, current.Version) > 0 {
			if compatibility == "" {
				compatibility = pluginregistry.CompatibilityCompatible
			}
		} else {
			compatibility = pluginregistry.CompatibilityIncompatible
			validation.Valid = false
			validation.Errors = append(validation.Errors, "升级版本必须高于当前版本")
		}
	}
	currentManifest := current.PluginManifest
	changedKeys := topLevelDiffKeys(normalizedManifest, currentManifest)
	diffSections, diffSummary := buildPluginManifestDiff(currentManifest, normalizedManifest)
	riskLevel := "low"
	riskScore := 5
	riskSummary := "升级差异未发现高风险变更"
	if diffSummary.HighRisk > 0 {
		riskLevel = "high"
		riskScore = 70
		riskSummary = fmt.Sprintf("升级差异包含 %d 个高风险变更", diffSummary.HighRisk)
	}
	if !validation.Valid || compatibility == pluginregistry.CompatibilityIncompatible {
		riskLevel = "blocked"
		riskScore = 100
		riskSummary = "升级 dry-run 存在阻断项"
	}
	diff := map[string]any{
		"current": map[string]any{
			"version":                 current.Version,
			"compatible_core_version": current.CompatibleCoreVersion,
			"content_types":           current.ContentTypes,
			"permissions":             current.Permissions,
			"menus":                   current.Menus,
			"routes":                  current.Routes,
			"hooks":                   current.Hooks,
			"migrations":              current.Migrations,
			"dependencies":            current.Dependencies,
		},
		"new": map[string]any{
			"version":                 normalizedManifest.Version,
			"compatible_core_version": normalizedManifest.CompatibleCoreVersion,
			"content_types":           normalizedManifest.ContentTypes,
			"permissions":             normalizedManifest.Permissions,
			"menus":                   normalizedManifest.Menus,
			"routes":                  normalizedManifest.Routes,
			"hooks":                   normalizedManifest.Hooks,
			"migrations":              normalizedManifest.Migrations,
			"dependencies":            normalizedManifest.Dependencies,
		},
	}
	return domain.PluginUpgradeDryRunResult{
		PluginCode:            code,
		CurrentVersion:        current.Version,
		NewVersion:            normalizedManifest.Version,
		CurrentCoreVersion:    currentCoreVersion(),
		CompatibleCoreVersion: normalizedManifest.CompatibleCoreVersion,
		CompatibilityStatus:   compatibility,
		ChangedKeys:           changedKeys,
		Diff:                  diff,
		DiffSections:          diffSections,
		DiffSummary:           diffSummary,
		RiskReport:            domain.PluginPackageRiskReport{Level: riskLevel, Score: riskScore, Summary: riskSummary},
		DependencyDiff:        pluginregistry.DependencyDiff(current.Dependencies, normalizedManifest.Dependencies),
		Validation:            validation,
	}, nil
}

func (s *Service) UpgradePluginManifest(code string, raw []byte) (domain.PluginUpgradeResult, error) {
	return s.upgradePluginManifestWithOperation(PluginOperationOperator{}, 0, "", code, raw)
}

func (s *Service) UpgradePluginManifestWithOperation(operator PluginOperationOperator, approvalID int64, operationID string, code string, raw []byte) (domain.PluginUpgradeResult, error) {
	return s.upgradePluginManifestWithOperation(operator, approvalID, operationID, code, raw)
}

func (s *Service) upgradePluginManifestWithOperation(operator PluginOperationOperator, approvalID int64, operationID string, code string, raw []byte) (domain.PluginUpgradeResult, error) {
	current, ok := s.repo.PluginByCode(code)
	if !ok || current.Code == "" {
		return domain.PluginUpgradeResult{}, pluginNotFound(code)
	}
	if strings.TrimSpace(current.Status) == pluginregistry.StatusArchived {
		return domain.PluginUpgradeResult{}, pluginArchived(code)
	}
	preview, err := s.PluginUpgradeDryRun(code, raw)
	if err != nil {
		return domain.PluginUpgradeResult{}, err
	}
	if !preview.Validation.Valid {
		return domain.PluginUpgradeResult{}, domain.NewPluginError(PluginErrManifestInvalid, "manifest 校验失败").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("errors", preview.Validation.Errors).
			WithSuggestion("请根据 errors 修复 manifest 后重试。")
	}
	if preview.CompatibilityStatus != "compatible" {
		return domain.PluginUpgradeResult{}, domain.NewPluginError(PluginErrCoreVersionIncompat, "升级版本不兼容或版本号未提升").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("compatibility", preview.Validation.Compatibility).
			WithSuggestion("请检查 Core 版本兼容范围和插件版本号。")
	}
	for _, item := range current.Migrations {
		_ = item
	}
	records, err := s.repo.PluginMigrations(code)
	if err != nil {
		return domain.PluginUpgradeResult{}, fmt.Errorf("读取插件迁移失败：%w", err)
	}
	for _, item := range records {
		if strings.TrimSpace(item.Status) == "failed" {
			return domain.PluginUpgradeResult{}, pluginMigrationFailed(code, item.MigrationName)
		}
	}

	opID := strings.TrimSpace(operationID)
	if opID == "" {
		opID = newPluginOperationID()
	}
	beforeManifestRaw, _ := json.Marshal(current.PluginManifest)
	beforeMigrationsRaw, _ := json.Marshal(records)
	snapshot := domain.PluginOperationSnapshot{
		OperationID:            opID,
		OperationType:          domain.PluginOperationTypeUpgrade,
		PluginCode:             code,
		FromVersion:            current.Version,
		ToVersion:              preview.NewVersion,
		PackagePath:            "",
		PackageSource:          "manifest",
		ApprovalID:             approvalID,
		BeforePluginJSON:       buildOperationSnapshotPluginJSON(current),
		BeforeManifestJSON:     string(beforeManifestRaw),
		BeforeConfigJSON:       buildOperationSnapshotConfigJSON(current),
		BeforeMigrationsJSON:   string(beforeMigrationsRaw),
		BeforePermissionsJSON:  mustJSON(scrubAnyForSnapshot(current.Permissions)),
		BeforeMenusJSON:        mustJSON(scrubAnyForSnapshot(current.Menus)),
		BeforeRoutesJSON:       mustJSON(scrubAnyForSnapshot(current.Routes)),
		BeforeDependenciesJSON: mustJSON(scrubAnyForSnapshot(current.Dependencies)),
		BeforeStatus:           current.Status,
		AfterManifestJSON:      scrubManifestJSONForSnapshot(string(raw)),
		DryRunJSON:             mustJSON(scrubAnyForSnapshot(preview)),
		RiskReportJSON:         mustJSON(scrubAnyForSnapshot(preview.RiskReport)),
		DiffJSON:               mustJSON(scrubAnyForSnapshot(map[string]any{"diff_sections": preview.DiffSections, "diff_summary": preview.DiffSummary})),
		ChecksumSummaryJSON:    "",
		SignatureSummaryJSON:   "",
		Status:                 domain.PluginOperationStatusCreated,
		CreatedBy:              operator.ID,
		MetadataJSON: mustJSON(map[string]any{
			"operator_name": strings.TrimSpace(operator.Name),
			"approval_id":   approvalID,
		}),
	}
	if saved, serr := s.repo.AppendPluginOperationSnapshot(snapshot); serr == nil {
		snapshot = saved
	}
	manifest := preview.Validation.NormalizedManifest
	manifest.IsSystem = current.IsSystem
	manifest.Status = current.Status
	manifest.SourceType = firstNonBlank(manifest.SourceType, current.SourceType, "manifest")
	manifestJSON, _ := json.Marshal(manifest)
	updated := current
	updated.PluginManifest = manifest
	updated.Name = manifest.Name
	updated.Version = manifest.Version
	updated.Description = manifest.Description
	updated.SourceType = manifest.SourceType
	updated.ManifestJSON = string(manifestJSON)
	updated.ManifestChecksum = preview.Validation.Checksum
	updated.CompatibleCoreVersion = firstNonBlank(manifest.CompatibleCoreVersion, manifest.MinCoreVersion, current.CompatibleCoreVersion)
	updated.ContentTypes = manifest.ContentTypes
	updated.ContentTypeDefs = manifest.ContentTypeDefs
	updated.Permissions = manifest.Permissions
	updated.Menus = manifest.Menus
	updated.Routes = manifest.Routes
	updated.ConfigSchema = manifest.ConfigSchema
	updated.Dependencies = manifest.Dependencies
	updated.MinCoreVersion = manifest.MinCoreVersion
	updated.Hooks = manifest.Hooks
	updated.Migrations = manifest.Migrations
	updated.Assets = manifest.Assets
	updated.ExternalService = manifest.ExternalService
	updated.Status = s.nextPluginStatusAfterUpgrade(current.Status, updated)
	saved, err := s.repo.SavePlugin(updated)
	if err != nil {
		snapshot.Status = domain.PluginOperationStatusFailed
		if apiErr, ok := err.(*domain.APIError); ok && apiErr != nil {
			snapshot.ErrorCode = apiErr.Code
			snapshot.ErrorMessage = apiErr.Message
		} else {
			snapshot.ErrorCode = "plugin_operation_apply_failed"
			snapshot.ErrorMessage = err.Error()
		}
		_, _ = s.repo.SavePluginOperationSnapshot(snapshot)
		return domain.PluginUpgradeResult{}, err
	}
	for _, migration := range manifest.Migrations {
		record := migrationRecordFromDefinition(migration, "pending")
		if _, merr := s.repo.SavePluginMigration(record); merr != nil {
			// Best-effort rollback: restore plugin manifest/config/status and clean to_version migration residues.
			_, _ = s.repo.SavePlugin(current)
			_, _ = s.repo.DeletePluginMigrationsByVersion(code, updated.Version)
			snapshot.Status = domain.PluginOperationStatusFailed
			snapshot.ErrorCode = "plugin_operation_apply_failed"
			snapshot.ErrorMessage = merr.Error()
			_, _ = s.repo.SavePluginOperationSnapshot(snapshot)
			return domain.PluginUpgradeResult{}, domain.NewPluginError("plugin_operation_apply_failed", "升级写入迁移记录失败，已尝试保护原版本").
				WithStatus(500).
				WithDetail("plugin_code", code).
				WithDetail("operation_id", snapshot.OperationID).
				WithSuggestion("请到“系统插件 -> 操作历史”查看失败详情，并执行 cleanup。")
		}
	}
	snapshot.Status = domain.PluginOperationStatusApplied
	snapshot.ErrorCode = ""
	snapshot.ErrorMessage = ""
	_, _ = s.repo.SavePluginOperationSnapshot(snapshot)
	return domain.PluginUpgradeResult{
		Plugin:                    saved,
		PluginUpgradeDryRunResult: preview,
	}, nil
}

func (s *Service) nextPluginStatusAfterUpgrade(currentStatus string, plugin domain.Plugin) string {
	currentStatus = strings.TrimSpace(currentStatus)
	if currentStatus == pluginregistry.StatusArchived {
		return currentStatus
	}
	if err := pluginregistry.ValidateConfigJSON(plugin, plugin.ConfigJSON); err != nil {
		return pluginregistry.StatusConfigInvalid
	}
	_, summary := pluginregistry.ResolvePluginDependencies(plugin.PluginManifest, s.repo.Plugins())
	if summary.Blocking > 0 {
		return pluginregistry.StatusDependencyMissing
	}
	switch currentStatus {
	case pluginregistry.StatusEnabled, pluginregistry.StatusDisabled:
		return currentStatus
	case pluginregistry.StatusConfigInvalid, pluginregistry.StatusDependencyMissing, pluginregistry.StatusInstalled, pluginregistry.StatusDiscovered, pluginregistry.StatusMigrated, pluginregistry.StatusConfigured, pluginregistry.StatusRunning, pluginregistry.StatusMigrationPending, pluginregistry.StatusMigrationFailed:
		return pluginregistry.StatusDisabled
	default:
		if currentStatus == "" {
			return pluginregistry.StatusDisabled
		}
		return currentStatus
	}
}

func (s *Service) BulkArchivePlugins(codes []string) domain.PluginBulkOperationResult {
	result := domain.PluginBulkOperationResult{}
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		plugin, err := s.ArchivePlugin(code)
		if err != nil {
			result.Failed = append(result.Failed, domain.PluginBulkItemResult{PluginCode: code, Error: err.Error()})
			continue
		}
		result.Succeeded = append(result.Succeeded, domain.PluginBulkItemResult{PluginCode: plugin.Code, Status: plugin.Status})
	}
	return result
}

func (s *Service) BulkRestorePlugins(codes []string) domain.PluginBulkOperationResult {
	result := domain.PluginBulkOperationResult{}
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		plugin, err := s.RestorePlugin(code)
		if err != nil {
			result.Failed = append(result.Failed, domain.PluginBulkItemResult{PluginCode: code, Error: err.Error()})
			continue
		}
		result.Succeeded = append(result.Succeeded, domain.PluginBulkItemResult{PluginCode: plugin.Code, Status: plugin.Status})
	}
	return result
}

func topLevelDiffKeys(newManifest, currentManifest domain.PluginManifest) []string {
	keys := []string{}
	appendIfDiff := func(key string, a, b any) {
		if fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b) {
			keys = append(keys, key)
		}
	}
	appendIfDiff("version", newManifest.Version, currentManifest.Version)
	appendIfDiff("compatible_core_version", newManifest.CompatibleCoreVersion, currentManifest.CompatibleCoreVersion)
	appendIfDiff("content_types", len(newManifest.ContentTypes), len(currentManifest.ContentTypes))
	appendIfDiff("permissions", len(newManifest.Permissions), len(currentManifest.Permissions))
	appendIfDiff("menus", len(newManifest.Menus), len(currentManifest.Menus))
	appendIfDiff("routes", len(newManifest.Routes), len(currentManifest.Routes))
	appendIfDiff("hooks", len(newManifest.Hooks), len(currentManifest.Hooks))
	appendIfDiff("migrations", len(newManifest.Migrations), len(currentManifest.Migrations))
	sort.Strings(keys)
	return keys
}

func currentCoreVersion() string {
	if raw, err := os.ReadFile("VERSION"); err == nil {
		if version := strings.TrimSpace(string(raw)); version != "" {
			return version
		}
	}
	return "v1.4.0"
}

// SetPluginStatus 更新插件状态。
func (s *Service) SetPluginStatus(code, status string) (domain.Plugin, error) {
	code = strings.TrimSpace(code)
	status = strings.TrimSpace(status)
	if status == pluginregistry.StatusEnabled {
		if err := s.validatePluginEnableReadiness(code); err != nil {
			return domain.Plugin{}, err
		}
	}
	return s.repo.SetPluginStatus(code, status)
}

// ArchivePlugin soft-uninstalls a plugin from runtime creation paths while
// preserving content, config, migrations and audit history.
func (s *Service) ArchivePlugin(code string) (domain.Plugin, error) {
	code = strings.TrimSpace(code)
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return domain.Plugin{}, pluginNotFound(code)
	}
	if plugin.Status == pluginregistry.StatusArchived {
		return plugin, nil
	}
	return s.repo.SetPluginStatus(code, pluginregistry.StatusArchived)
}

// RestorePlugin brings an archived plugin back to an installed/disabled state.
// It deliberately does not auto-enable the plugin; admins must re-enable it
// after readiness checks pass.
func (s *Service) RestorePlugin(code string) (domain.Plugin, error) {
	code = strings.TrimSpace(code)
	if err := s.validatePluginRestoreReadiness(code); err != nil {
		return domain.Plugin{}, err
	}
	return s.repo.SetPluginStatus(code, pluginregistry.StatusDisabled)
}

func (s *Service) validatePluginEnableReadiness(code string) error {
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return pluginNotFound(code)
	}
	if strings.TrimSpace(plugin.InstallStatus) == pluginregistry.StatusDiscovered {
		return pluginNotInstalled(code)
	}
	switch strings.TrimSpace(plugin.Status) {
	case pluginregistry.StatusDiscovered:
		return pluginNotInstalled(code)
	case pluginregistry.StatusArchived:
		return pluginArchived(code)
	case pluginregistry.StatusMigrationFailed:
		return pluginMigrationFailed(code, "")
	}
	return s.validatePluginRuntimeReadiness(plugin)
}

func (s *Service) validatePluginRestoreReadiness(code string) error {
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return pluginNotFound(code)
	}
	if plugin.Status != pluginregistry.StatusArchived {
		return domain.NewPluginError(PluginErrArchived, "插件未归档，无需恢复").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithSuggestion("请刷新插件状态后重试。")
	}
	return s.validatePluginRuntimeReadiness(plugin)
}

func (s *Service) validatePluginRuntimeReadiness(plugin domain.Plugin) error {
	if err := pluginregistry.ValidateConfigJSON(plugin, plugin.ConfigJSON); err != nil {
		msg := strings.TrimSpace(err.Error())
		path := ""
		if strings.HasPrefix(msg, "$") {
			parts := strings.Fields(msg)
			if len(parts) > 0 {
				path = parts[0]
			}
		}
		// config_json invalid against config_schema should be surfaced as schema-invalid
		// for a more actionable UI (field path).
		return pluginConfigSchemaInvalid(plugin.Code, path, msg)
	}
	checks, _, dependencyErrors, _ := pluginregistry.ValidatePluginDependencies(plugin.PluginManifest, s.repo.Plugins())
	if len(dependencyErrors) > 0 {
		for _, check := range checks {
			if check.Satisfied {
				continue
			}
			// Only blocking checks should block readiness; optional missing remains warning-only.
			if !check.Required && check.Status != pluginregistry.DependencySelfDependency && check.Status != pluginregistry.DependencyCircularDependency {
				continue
			}
			return dependencyAPIError(plugin.Code, check)
		}
		return domain.NewPluginError(PluginErrDependencyMissing, "插件依赖未满足，无法执行该操作").
			WithStatus(400).
			WithDetail("plugin_code", plugin.Code).
			WithDetail("errors", dependencyErrors).
			WithSuggestion("请先修复依赖插件状态或版本后重试。")
	}
	migrations, err := s.pluginMigrationsWithDefinitions(plugin.Code)
	if err != nil {
		return fmt.Errorf("插件迁移状态不可用：%w", err)
	}
	for _, item := range migrations {
		if strings.TrimSpace(item.Status) == "failed" {
			return pluginMigrationFailed(plugin.Code, item.MigrationName)
		}
	}
	return nil
}

// SetPluginConfig updates global plugin config_json.
func (s *Service) SetPluginConfig(code, configJSON string) (domain.Plugin, error) {
	code = strings.TrimSpace(code)
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return domain.Plugin{}, pluginNotFound(code)
	}

	res, err := s.encryptPluginConfigJSON(plugin, plugin.ConfigJSON, configJSON)
	if err != nil {
		return domain.Plugin{}, err
	}
	return s.repo.SetPluginConfig(code, res.EncryptedJSON)
}

func (s *Service) PluginImpact(code string) (domain.PluginImpact, error) {
	return s.repo.PluginImpact(code)
}

func (s *Service) CommunityPluginImpact(communityID int64, code string) (domain.PluginImpact, error) {
	return s.repo.CommunityPluginImpact(communityID, code)
}

func (s *Service) PluginMigrations(pluginCode string) ([]domain.PluginMigration, error) {
	return s.pluginMigrationsWithDefinitions(pluginCode)
}

func (s *Service) AppendPluginMigration(record domain.PluginMigration) (domain.PluginMigration, error) {
	return s.repo.AppendPluginMigration(record)
}

// InjectFailedPluginMigrationForTest writes a failed built-in migration record
// for E2E/API tests. It deliberately reuses the normal migration store path so
// enable-readiness checks, retry and audit can exercise production behavior.
func (s *Service) InjectFailedPluginMigrationForTest(pluginCode, migrationName, errorMessage, executor string) (domain.PluginMigration, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	migrationName = strings.TrimSpace(migrationName)
	if _, ok := s.repo.PluginByCode(pluginCode); !ok {
		return domain.PluginMigration{}, errors.New("插件不存在")
	}
	defs := pluginregistry.MigrationDefinitions(pluginCode)
	var def domain.PluginMigrationDefinition
	for _, item := range defs {
		if item.MigrationName == migrationName {
			def = item
			break
		}
	}
	if def.MigrationName == "" {
		return domain.PluginMigration{}, errors.New("迁移不存在")
	}
	now := time.Now()
	record := migrationRecordFromDefinition(def, "failed")
	record.Executor = firstNonBlank(executor, "e2e")
	record.StartedAt = now.Format("2006-01-02 15:04:05")
	record.FinishedAt = record.StartedAt
	record.ExecutedAt = record.FinishedAt
	record.ErrorMessage = firstNonBlank(errorMessage, "E2E injected failed migration")
	return s.repo.SavePluginMigration(record)
}

func (s *Service) RunPluginMigration(pluginCode, migrationName, executor string) (domain.PluginMigration, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	migrationName = strings.TrimSpace(migrationName)
	plugin, ok := s.repo.PluginByCode(pluginCode)
	if !ok || plugin.Code == "" {
		return domain.PluginMigration{}, errors.New("插件不存在")
	}
	defs := pluginregistry.MigrationDefinitions(pluginCode)
	var def domain.PluginMigrationDefinition
	for _, item := range defs {
		if item.MigrationName == migrationName {
			def = item
			break
		}
	}
	if def.MigrationName == "" {
		return domain.PluginMigration{}, errors.New("迁移不存在")
	}
	if def.Direction != "" && def.Direction != "up" {
		return domain.PluginMigration{}, errors.New("当前仅支持 up migration")
	}
	records, _ := s.repo.PluginMigrations(pluginCode)
	for _, item := range records {
		if sameMigration(item, def) && item.Status == "success" {
			return enrichMigrationRecord(item, def), nil
		}
	}
	now := time.Now()
	running := migrationRecordFromDefinition(def, "running")
	running.Executor = executor
	running.StartedAt = now.Format("2006-01-02 15:04:05")
	running, _ = s.repo.SavePluginMigration(running)

	result := migrationRecordFromDefinition(def, "success")
	result.Executor = executor
	result.StartedAt = running.StartedAt
	result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
	result.DurationMS = int(time.Since(now).Milliseconds())
	result.ExecutionTimeMS = result.DurationMS
	result.ExecutedAt = result.FinishedAt
	// v1.3.2 built-in migrations are idempotent no-op confirmations because the
	// authoritative schema already creates these plugin tables on startup.
	result.Description = firstNonBlank(def.Description, "内置插件迁移表结构已由主 schema 保证，本次记录为 no-op success")
	return s.repo.SavePluginMigration(result)
}

func (s *Service) RunAllPluginMigrations(pluginCode, executor string) ([]domain.PluginMigration, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if _, ok := s.repo.PluginByCode(pluginCode); !ok {
		return nil, errors.New("插件不存在")
	}
	defs := pluginregistry.MigrationDefinitions(pluginCode)
	out := make([]domain.PluginMigration, 0, len(defs))
	for _, def := range defs {
		item, err := s.RunPluginMigration(pluginCode, def.MigrationName, executor)
		if err != nil {
			failed := migrationRecordFromDefinition(def, "failed")
			failed.Executor = executor
			failed.StartedAt = time.Now().Format("2006-01-02 15:04:05")
			failed.FinishedAt = failed.StartedAt
			failed.ErrorMessage = err.Error()
			if saved, saveErr := s.repo.SavePluginMigration(failed); saveErr == nil {
				out = append(out, saved)
			}
			return out, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *Service) HookExecutions(pluginCode string, limit int) ([]domain.HookExecution, error) {
	return s.repo.HookExecutions(pluginCode, limit)
}

func (s *Service) HookStats(pluginCode string) ([]domain.HookStats, error) {
	return s.repo.HookStats(pluginCode)
}

func (s *Service) HookExecutionsByFilter(filter domain.HookExecutionFilter) ([]domain.HookExecution, int, error) {
	return s.repo.HookExecutionsByFilter(filter.Normalize())
}

func (s *Service) pluginMigrationsWithDefinitions(pluginCode string) ([]domain.PluginMigration, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	records, err := s.repo.PluginMigrations(pluginCode)
	if err != nil {
		return nil, err
	}
	defs := pluginregistry.MigrationDefinitions(pluginCode)
	if len(defs) == 0 {
		return records, nil
	}
	seen := map[string]bool{}
	out := make([]domain.PluginMigration, 0, len(records)+len(defs))
	for _, record := range records {
		key := migrationKey(record.PluginCode, record.MigrationVersion, record.MigrationName)
		matched := false
		for _, def := range defs {
			if sameMigration(record, def) {
				record = enrichMigrationRecord(record, def)
				matched = true
				break
			}
		}
		if !matched && record.MigrationVersion == "" {
			record.MigrationVersion = record.Version
		}
		record.Declared = matched
		seen[key] = true
		out = append(out, record)
	}
	for _, def := range defs {
		key := migrationKey(def.PluginCode, def.MigrationVersion, def.MigrationName)
		if seen[key] {
			continue
		}
		out = append(out, migrationRecordFromDefinition(def, "pending"))
	}
	return out, nil
}

func migrationRecordFromDefinition(def domain.PluginMigrationDefinition, status string) domain.PluginMigration {
	return domain.PluginMigration{
		PluginCode:        def.PluginCode,
		MigrationVersion:  firstNonBlank(def.MigrationVersion, "1.0.0"),
		Version:           firstNonBlank(def.MigrationVersion, "1.0.0"),
		MigrationName:     def.MigrationName,
		Direction:         firstNonBlank(def.Direction, "up"),
		Checksum:          def.Checksum,
		Status:            firstNonBlank(status, "pending"),
		RollbackSupported: def.RollbackSupported,
		Description:       def.Description,
		Declared:          true,
	}
}

func enrichMigrationRecord(record domain.PluginMigration, def domain.PluginMigrationDefinition) domain.PluginMigration {
	if record.MigrationVersion == "" {
		record.MigrationVersion = firstNonBlank(record.Version, def.MigrationVersion)
	}
	if record.Version == "" {
		record.Version = record.MigrationVersion
	}
	if record.Direction == "" {
		record.Direction = firstNonBlank(def.Direction, "up")
	}
	if record.Checksum == "" {
		record.Checksum = def.Checksum
	}
	record.RollbackSupported = def.RollbackSupported
	if record.Description == "" {
		record.Description = def.Description
	}
	record.Declared = true
	return record
}

func sameMigration(record domain.PluginMigration, def domain.PluginMigrationDefinition) bool {
	return strings.TrimSpace(record.PluginCode) == strings.TrimSpace(def.PluginCode) &&
		firstNonBlank(record.MigrationVersion, record.Version) == firstNonBlank(def.MigrationVersion, "1.0.0") &&
		strings.TrimSpace(record.MigrationName) == strings.TrimSpace(def.MigrationName)
}

func migrationKey(pluginCode, version, name string) string {
	return strings.TrimSpace(pluginCode) + ":" + strings.TrimSpace(version) + ":" + strings.TrimSpace(name)
}

// PluginHealth computes a lightweight runtime health summary for plugin governance UI.
func (s *Service) PluginHealth(code string) (domain.PluginHealth, error) {
	code = strings.TrimSpace(code)
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return domain.PluginHealth{}, errors.New("插件不存在")
	}
	health := domain.PluginHealth{
		Status:           "healthy",
		ConfigStatus:     "valid",
		MigrationStatus:  "ok",
		HookStatus:       "ok",
		DependencyStatus: "ok",
		SuggestedAction:  "无需处理",
		StatusReason:     "插件运行状态正常",
		UpdatedAt:        time.Now().Format("2006-01-02 15:04:05"),
	}

	disabled := plugin.Status == pluginregistry.StatusDisabled || plugin.Status == pluginregistry.StatusArchived
	if plugin.Status == pluginregistry.StatusArchived {
		health.Status = "archived"
		health.SuggestedAction = "如需恢复插件治理能力，请先恢复插件，再手动启用"
		health.StatusReason = "插件已归档，新发布和入口已关闭，历史内容与 SEO 保留"
	} else if plugin.Status == pluginregistry.StatusDisabled {
		health.Status = "disabled"
		health.SuggestedAction = "如需恢复新发布和入口展示，请启用插件"
		health.StatusReason = "插件已全局禁用，仅影响新发布和入口展示"
	} else if plugin.Status == pluginregistry.StatusConfigInvalid {
		health.Status = "config_invalid"
		health.ConfigStatus = "invalid"
		health.SuggestedAction = "检查插件 config_json"
		health.StatusReason = "插件状态标记为配置无效"
	} else if plugin.Status == pluginregistry.StatusMigrationPending {
		health.Status = "migration_pending"
		health.MigrationStatus = "pending"
		health.SuggestedAction = "执行或确认插件迁移"
		health.StatusReason = "插件状态标记为存在待处理迁移"
	} else if plugin.Status == pluginregistry.StatusMigrationFailed {
		health.Status = "error"
		health.MigrationStatus = "failed"
		health.SuggestedAction = "查看并重试失败迁移"
		health.StatusReason = "插件状态标记为存在失败迁移"
	} else if plugin.Status == pluginregistry.StatusDependencyMissing {
		health.Status = "dependency_missing"
		health.DependencyStatus = "missing"
		health.SuggestedAction = "检查依赖插件状态"
		health.StatusReason = "插件状态标记为依赖缺失"
	}

	if err := pluginregistry.ValidateConfigJSON(plugin, plugin.ConfigJSON); err != nil {
		if !disabled {
			health.Status = "config_invalid"
		}
		health.ConfigStatus = "invalid"
		health.RecentError = err.Error()
		if !disabled {
			health.SuggestedAction = "修正插件全局配置"
			health.StatusReason = "插件全局配置未通过 config_schema 校验"
		}
	}

	dependencyChecks, dependencySummary := pluginregistry.ResolvePluginDependencies(plugin.PluginManifest, s.repo.Plugins())
	if dependencySummary.Blocking > 0 {
		for _, dep := range dependencyChecks {
			if dep.Satisfied {
				continue
			}
			if !disabled {
				health.Status = "dependency_missing"
			}
			health.DependencyStatus = dep.Status
			if health.RecentError == "" {
				health.RecentError = dep.Message
			}
			if !disabled {
				health.SuggestedAction = "启用或修复依赖插件"
				health.StatusReason = "依赖插件未启用、版本不匹配或不可用"
			}
			break
		}
	}

	migrations, err := s.pluginMigrationsWithDefinitions(code)
	if err == nil {
		for _, item := range migrations {
			switch strings.TrimSpace(item.Status) {
			case "pending":
				health.PendingMigrationsCount++
			case "failed":
				health.FailedMigrationsCount++
				if item.ErrorMessage != "" {
					health.RecentError = item.ErrorMessage
				}
			}
		}
		if health.FailedMigrationsCount > 0 {
			if !disabled {
				health.Status = "error"
			}
			health.MigrationStatus = "failed"
			if !disabled {
				health.SuggestedAction = "查看并重试失败迁移"
				health.StatusReason = "存在 failed migration"
			}
		} else if health.PendingMigrationsCount > 0 && health.Status == "healthy" && !disabled {
			health.Status = "migration_pending"
			health.MigrationStatus = "pending"
			health.SuggestedAction = "确认并执行待处理迁移"
			health.StatusReason = "存在待处理迁移记录"
		}
	}

	stats, err := s.repo.HookStats(code)
	if err == nil {
		for _, stat := range stats {
			health.HookFailureCount += stat.FailureCount
			if stat.Blocking && stat.FailureCount > 0 {
				health.HookStatus = "hook_error"
			} else if stat.FailureCount > 0 && health.HookStatus != "hook_error" {
				health.HookStatus = "hook_warning"
			}
			if stat.LastError != "" {
				health.LastHookError = stat.LastError
				health.RecentError = stat.LastError
			}
		}
		if health.HookFailureCount > 0 {
			hookStatus := health.HookStatus
			if hookStatus == "" {
				hookStatus = "hook_warning"
			}
			if health.HookFailureCount >= 3 {
				hookStatus = "hook_error"
				health.HookStatus = "hook_error"
			}
			if health.Status == "healthy" && !disabled {
				health.Status = hookStatus
				health.SuggestedAction = "查看 Hooks Tab 的最近失败记录"
				health.StatusReason = "存在 Hook 失败记录"
			}
		}
	}

	return health, nil
}

// ContentTypeCreatePermission resolves a content type create permission using
// both built-in registry definitions and installed manifest plugins.
func (s *Service) ContentTypeCreatePermission(contentType, pluginCode string) string {
	return s.createPermissionForContentType(contentType, pluginCode)
}

// ListPosts 按站点、板块、关键词和标签筛选帖子列表。
func (s *Service) ListPosts(site, board, q, tag string) []domain.Post {
	return s.repo.ListPosts(site, board, q, tag)
}

// GetPost 获取帖子详情，increaseView 为 true 时同步增加浏览量。
func (s *Service) GetPost(id int64, increaseView bool) (*domain.Post, bool) {
	return s.repo.GetPost(id, increaseView)
}

// CreatePost is no longer a business write entry after v1.3.1.
// Keep it only to satisfy the legacy repository contract; callers must use CreateTopic.
func (s *Service) CreatePost(req domain.CreatePostRequest) (*domain.Post, error) {
	return nil, errors.New("posts 写接口已废弃，请使用 CreateTopic")
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
	notice := s.repo.PushNotification(req)
	if notice != nil {
		_ = s.dispatchHook(pluginregistry.HookEvent{
			Name:         pluginregistry.HookOnNotificationBuild,
			Mode:         pluginregistry.HookNonBlocking,
			Ctx:          pluginregistry.HookContext{ActorType: pluginregistry.HookActorSystem},
			Notification: notice,
		})
	}
	return notice
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
	topic, err := s.repo.TopicByID(id, increaseView)
	if err == nil && topic != nil {
		_ = s.dispatchHook(pluginregistry.HookEvent{
			Name:  pluginregistry.HookOnSEOBuild,
			Mode:  pluginregistry.HookNonBlocking,
			Ctx:   pluginregistry.HookContext{PluginCode: topic.PluginCode, ContentType: topic.ContentType, CommunityID: topic.CommunityID, CategoryID: topic.CategoryID, ContentID: topic.ID, ActorType: pluginregistry.HookActorSystem},
			Topic: topic,
		})
	}
	return topic, err
}

// CreateTopic 创建主题。
func (s *Service) CreateTopic(req domain.CreateTopicRequest) (*domain.Topic, error) {
	normalizedType, pluginCode, err := s.ValidateTopicPluginAccess(req.CommunityID, req.CategoryID, req.ContentType)
	if err != nil {
		return nil, err
	}
	req.ContentType = normalizedType
	req.PluginCode = pluginCode
	if len(req.ActorContext.Permissions) > 0 {
		req.ActorPermissions = append([]string{}, req.ActorContext.Permissions...)
	}

	if perm := s.createPermissionForContentType(normalizedType, pluginCode); perm != "" {
		if !hasPermission(req.ActorPermissions, perm) {
			return nil, fmt.Errorf("缺少权限 %s，不能创建该类型内容", perm)
		}
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeCreateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  pluginCode,
			ContentType: normalizedType,
			CommunityID: req.CommunityID,
			CategoryID:  req.CategoryID,
			ActorType:   actorTypeFromActor(req.ActorContext),
			ActorID:     actorIDFromActor(req.ActorContext),
			Actor:       req.ActorContext,
		},
		Request: &req,
	}); err != nil {
		return nil, err
	}
	topic, err := s.repo.CreateTopic(req)
	if err != nil {
		return nil, err
	}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  pluginCode,
			ContentType: normalizedType,
			CommunityID: req.CommunityID,
			CategoryID:  req.CategoryID,
			ContentID:   topic.ID,
			ActorType:   actorTypeFromActor(req.ActorContext),
			ActorID:     actorIDFromActor(req.ActorContext),
			Actor:       req.ActorContext,
		},
		Topic: topic,
	})
	return topic, nil
}

// UpdateTopic 更新主题。
func (s *Service) UpdateTopic(id int64, req domain.UpdateTopicRequest) (*domain.Topic, error) {
	before, _ := s.repo.TopicByID(id, false)
	pluginCode := ""
	if before != nil {
		pluginCode = before.PluginCode
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeUpdateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode: pluginCode,
			ContentType: firstNonEmpty(func() string {
				if req.ContentType != nil {
					return *req.ContentType
				}
				return ""
			}(), func() string {
				if before != nil {
					return before.ContentType
				}
				return ""
			}()),
			CommunityID: func() int64 {
				if before != nil {
					return before.CommunityID
				}
				return 0
			}(),
			CategoryID: func() int64 {
				if before != nil {
					return before.CategoryID
				}
				return 0
			}(),
			ContentID: id,
			ActorType: actorTypeFromActor(req.ActorContext),
			ActorID:   actorIDFromActor(req.ActorContext),
			Actor:     req.ActorContext,
		},
		PreviousTopic: before,
		UpdateRequest: &req,
	}); err != nil {
		return nil, err
	}
	topic, err := s.repo.UpdateTopic(id, req)
	if err != nil {
		return nil, err
	}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterUpdateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  topic.PluginCode,
			ContentType: topic.ContentType,
			CommunityID: topic.CommunityID,
			CategoryID:  topic.CategoryID,
			ContentID:   topic.ID,
			ActorType:   actorTypeFromActor(req.ActorContext),
			ActorID:     actorIDFromActor(req.ActorContext),
			Actor:       req.ActorContext,
		},
		Topic:         topic,
		PreviousTopic: before,
	})
	return topic, nil
}

// DeleteTopic 删除主题。
func (s *Service) DeleteTopic(id int64) bool {
	before, _ := s.repo.TopicByID(id, false)
	pluginCode := ""
	if before != nil {
		pluginCode = before.PluginCode
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeModerateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode: pluginCode,
			ContentType: func() string {
				if before != nil {
					return before.ContentType
				}
				return ""
			}(),
			CommunityID: func() int64 {
				if before != nil {
					return before.CommunityID
				}
				return 0
			}(),
			CategoryID: func() int64 {
				if before != nil {
					return before.CategoryID
				}
				return 0
			}(),
			ContentID: id,
			ActorType: pluginregistry.HookActorSystem,
			ActorID:   0,
		},
		PreviousTopic: before,
	}); err != nil {
		return false
	}
	deleted := s.repo.DeleteTopic(id)
	if deleted {
		_ = s.dispatchHook(pluginregistry.HookEvent{
			Name: pluginregistry.HookAfterModerateContent,
			Mode: pluginregistry.HookNonBlocking,
			Ctx: pluginregistry.HookContext{
				PluginCode: pluginCode,
				ContentType: func() string {
					if before != nil {
						return before.ContentType
					}
					return ""
				}(),
				CommunityID: func() int64 {
					if before != nil {
						return before.CommunityID
					}
					return 0
				}(),
				CategoryID: func() int64 {
					if before != nil {
						return before.CategoryID
					}
					return 0
				}(),
				ContentID: id,
				ActorType: pluginregistry.HookActorSystem,
				ActorID:   0,
			},
			PreviousTopic: before,
		})
	}
	return deleted
}

// SearchTopics 搜索主题。
func (s *Service) SearchTopics(req domain.SearchRequest) ([]domain.Topic, int) {
	topics, total := s.repo.SearchTopics(req)
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name:          pluginregistry.HookOnSearchIndex,
		Mode:          pluginregistry.HookNonBlocking,
		Ctx:           pluginregistry.HookContext{ActorType: pluginregistry.HookActorSystem},
		SearchRequest: &req,
		SearchResults: topics,
	})
	return topics, total
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
	before, _ := s.repo.TopicByID(id, false)
	pluginCode := ""
	if before != nil {
		pluginCode = before.PluginCode
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeUpdateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{PluginCode: pluginCode, ContentType: func() string {
			if before != nil {
				return before.ContentType
			}
			return ""
		}(), CommunityID: func() int64 {
			if before != nil {
				return before.CommunityID
			}
			return 0
		}(), CategoryID: func() int64 {
			if before != nil {
				return before.CategoryID
			}
			return 0
		}(), ContentID: id, ActorType: pluginregistry.HookActorSystem},
		PreviousTopic: before,
	}); err != nil {
		return nil, err
	}
	topic, err := s.repo.SetTopicFeatured(id, featured)
	if err != nil {
		return nil, err
	}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name:          pluginregistry.HookAfterUpdateContent,
		Mode:          pluginregistry.HookNonBlocking,
		Ctx:           pluginregistry.HookContext{PluginCode: topic.PluginCode, ContentType: topic.ContentType, CommunityID: topic.CommunityID, CategoryID: topic.CategoryID, ContentID: topic.ID, ActorType: pluginregistry.HookActorSystem},
		Topic:         topic,
		PreviousTopic: before,
	})
	return topic, nil
}

func (s *Service) SetTopicPinned(id int64, pinned bool) (*domain.Topic, error) {
	before, _ := s.repo.TopicByID(id, false)
	pluginCode := ""
	if before != nil {
		pluginCode = before.PluginCode
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeUpdateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{PluginCode: pluginCode, ContentType: func() string {
			if before != nil {
				return before.ContentType
			}
			return ""
		}(), CommunityID: func() int64 {
			if before != nil {
				return before.CommunityID
			}
			return 0
		}(), CategoryID: func() int64 {
			if before != nil {
				return before.CategoryID
			}
			return 0
		}(), ContentID: id, ActorType: pluginregistry.HookActorSystem},
		PreviousTopic: before,
	}); err != nil {
		return nil, err
	}
	topic, err := s.repo.SetTopicPinned(id, pinned)
	if err != nil {
		return nil, err
	}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name:          pluginregistry.HookAfterUpdateContent,
		Mode:          pluginregistry.HookNonBlocking,
		Ctx:           pluginregistry.HookContext{PluginCode: topic.PluginCode, ContentType: topic.ContentType, CommunityID: topic.CommunityID, CategoryID: topic.CategoryID, ContentID: topic.ID, ActorType: pluginregistry.HookActorSystem},
		Topic:         topic,
		PreviousTopic: before,
	})
	return topic, nil
}

func (s *Service) SetTopicStatus(id int64, status int) (*domain.Topic, error) {
	before, _ := s.repo.TopicByID(id, false)
	pluginCode := ""
	if before != nil {
		pluginCode = before.PluginCode
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeUpdateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{PluginCode: pluginCode, ContentType: func() string {
			if before != nil {
				return before.ContentType
			}
			return ""
		}(), CommunityID: func() int64 {
			if before != nil {
				return before.CommunityID
			}
			return 0
		}(), CategoryID: func() int64 {
			if before != nil {
				return before.CategoryID
			}
			return 0
		}(), ContentID: id, ActorType: pluginregistry.HookActorSystem},
		PreviousTopic: before,
	}); err != nil {
		return nil, err
	}
	topic, err := s.repo.SetTopicStatus(id, status)
	if err != nil {
		return nil, err
	}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name:          pluginregistry.HookAfterUpdateContent,
		Mode:          pluginregistry.HookNonBlocking,
		Ctx:           pluginregistry.HookContext{PluginCode: topic.PluginCode, ContentType: topic.ContentType, CommunityID: topic.CommunityID, CategoryID: topic.CategoryID, ContentID: topic.ID, ActorType: pluginregistry.HookActorSystem},
		Topic:         topic,
		PreviousTopic: before,
	})
	return topic, nil
}

func (s *Service) SetTopicCommentLocked(id int64, locked bool) (*domain.Topic, error) {
	before, _ := s.repo.TopicByID(id, false)
	pluginCode := ""
	if before != nil {
		pluginCode = before.PluginCode
	}
	if err := s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeUpdateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{PluginCode: pluginCode, ContentType: func() string {
			if before != nil {
				return before.ContentType
			}
			return ""
		}(), CommunityID: func() int64 {
			if before != nil {
				return before.CommunityID
			}
			return 0
		}(), CategoryID: func() int64 {
			if before != nil {
				return before.CategoryID
			}
			return 0
		}(), ContentID: id, ActorType: pluginregistry.HookActorSystem},
		PreviousTopic: before,
	}); err != nil {
		return nil, err
	}
	topic, err := s.repo.SetTopicCommentLocked(id, locked)
	if err != nil {
		return nil, err
	}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name:          pluginregistry.HookAfterUpdateContent,
		Mode:          pluginregistry.HookNonBlocking,
		Ctx:           pluginregistry.HookContext{PluginCode: topic.PluginCode, ContentType: topic.ContentType, CommunityID: topic.CommunityID, CategoryID: topic.CategoryID, ContentID: topic.ID, ActorType: pluginregistry.HookActorSystem},
		Topic:         topic,
		PreviousTopic: before,
	})
	return topic, nil
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
	comment, err := s.repo.CreateCommentWithRequest(topicID, req)
	if err != nil {
		return nil, err
	}
	topic, _ := s.repo.TopicByID(topicID, false)
	pluginCode := ""
	contentType := ""
	communityID := int64(0)
	categoryID := int64(0)
	if topic != nil {
		pluginCode = topic.PluginCode
		contentType = topic.ContentType
		communityID = topic.CommunityID
		categoryID = topic.CategoryID
	}
	actor := domain.ActorContext{UserID: req.ActorUserID, Permissions: nil}
	_ = s.dispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateComment,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  pluginCode,
			ContentType: contentType,
			CommunityID: communityID,
			CategoryID:  categoryID,
			ContentID:   topicID,
			ActorType:   actorTypeFromActor(actor),
			ActorID:     actorIDFromActor(actor),
			Actor:       actor,
			Metadata:    map[string]any{"topic_id": topicID},
		},
		Topic:   topic,
		Comment: comment,
	})
	return comment, nil
}

func actorTypeFromActor(actor domain.ActorContext) pluginregistry.HookActorType {
	if actor.IsAdmin {
		return pluginregistry.HookActorAdmin
	}
	if actor.IsModerator {
		return pluginregistry.HookActorModerator
	}
	if actor.UserID > 0 {
		return pluginregistry.HookActorUser
	}
	return pluginregistry.HookActorSystem
}

func actorIDFromActor(actor domain.ActorContext) int64 {
	if actor.IsAdmin && actor.AdminID > 0 {
		return actor.AdminID
	}
	if actor.UserID > 0 {
		return actor.UserID
	}
	return 0
}
