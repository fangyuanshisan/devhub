package domain

// Site 描述一个站点或子站的展示配置。
type Site struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Title       string `json:"title"`
	Sub         string `json:"sub"`
	Pub         string `json:"pub"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Status      string `json:"status"`
	Sort        int    `json:"sort"`
}

// Board 描述内容板块的基础配置。
type Board struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Site    string `json:"site"`
	Sort    int    `json:"sort"`
	Visible bool   `json:"visible"`
}

// ContentTypeDefinition 描述一个内容类型的统一声明。
type ContentTypeDefinition struct {
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	PluginCode       string   `json:"plugin_code"`
	Aliases          []string `json:"aliases,omitempty"`
	CreatePermission string   `json:"create_permission,omitempty"`
	EditPermission   string   `json:"edit_permission,omitempty"`
	DeletePermission string   `json:"delete_permission,omitempty"`
	AuditPermission  string   `json:"audit_permission,omitempty"`
	DefaultStatus    string   `json:"default_status,omitempty"`
	AllowComment     bool     `json:"allow_comment"`
	AllowLike        bool     `json:"allow_like"`
	AllowFavorite    bool     `json:"allow_favorite"`
	SEOType          string   `json:"seo_type,omitempty"`
}

// PermissionDefinition 描述插件或 Core 暴露的权限点。
type PermissionDefinition struct {
	PluginCode  string `json:"plugin_code"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// MenuDefinition 描述插件暴露的菜单入口。
type MenuDefinition struct {
	PluginCode string `json:"plugin_code"`
	Code       string `json:"code,omitempty"`
	Key        string `json:"key,omitempty"`
	Title      string `json:"title"`
	Short      string `json:"short,omitempty"`
	Path       string `json:"path"`
	Location   string `json:"location,omitempty"`
	Area       string `json:"area,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Permission string `json:"permission,omitempty"`
	SortOrder  int    `json:"sort_order,omitempty"`
}

// RouteDefinition 描述插件声明的路由元数据。
type RouteDefinition struct {
	PluginCode string `json:"plugin_code"`
	Area       string `json:"area"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Handler    string `json:"handler,omitempty"`
}

// HookDefinition 描述插件可声明的扩展 Hook。
type HookDefinition struct {
	PluginCode    string `json:"plugin_code"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Critical      bool   `json:"critical"`
	FailurePolicy string `json:"failure_policy,omitempty"`
}

// PluginManifest 描述插件声明层，不直接承载运行时流程。
type PluginManifest struct {
	Code            string                      `json:"code"`
	PluginCode      string                      `json:"plugin_code,omitempty"`
	Name            string                      `json:"name"`
	Version         string                      `json:"version"`
	Description     string                      `json:"description,omitempty"`
	IsSystem        bool                        `json:"is_system"`
	ContentTypes    []string                    `json:"content_types,omitempty"`
	ContentTypeDefs []ContentTypeDefinition     `json:"content_type_definitions,omitempty"`
	Permissions     []PermissionDefinition      `json:"permissions,omitempty"`
	Menus           []MenuDefinition            `json:"menus,omitempty"`
	Routes          []RouteDefinition           `json:"routes,omitempty"`
	ConfigSchema    any                         `json:"config_schema,omitempty"`
	Dependencies    []string                    `json:"dependencies,omitempty"`
	MinCoreVersion  string                      `json:"min_core_version,omitempty"`
	Hooks           []HookDefinition            `json:"hooks,omitempty"`
	Migrations      []PluginMigrationDefinition `json:"migrations,omitempty"`
}

// PluginMigrationDefinition describes a built-in plugin migration declaration.
type PluginMigrationDefinition struct {
	PluginCode        string   `json:"plugin_code"`
	MigrationVersion  string   `json:"migration_version"`
	MigrationName     string   `json:"migration_name"`
	Direction         string   `json:"direction"`
	Checksum          string   `json:"checksum,omitempty"`
	Tables            []string `json:"tables,omitempty"`
	RollbackSupported bool     `json:"rollback_supported"`
	Description       string   `json:"description,omitempty"`
}

// PluginMigration represents a plugin migration execution record.
type PluginMigration struct {
	ID                int64  `json:"id"`
	PluginCode        string `json:"plugin_code"`
	MigrationVersion  string `json:"migration_version"`
	Version           string `json:"version,omitempty"`
	MigrationName     string `json:"migration_name"`
	Direction         string `json:"direction,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
	Status            string `json:"status"`
	StartedAt         string `json:"started_at,omitempty"`
	FinishedAt        string `json:"finished_at,omitempty"`
	DurationMS        int    `json:"duration_ms,omitempty"`
	ExecutedAt        string `json:"executed_at,omitempty"`
	ExecutionTimeMS   int    `json:"execution_time_ms,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`
	Executor          string `json:"executor,omitempty"`
	RollbackSupported bool   `json:"rollback_supported"`
	Description       string `json:"description,omitempty"`
	Declared          bool   `json:"declared"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

// HookExecution represents one built-in plugin HookBus handler execution.
type HookExecution struct {
	ID           int64  `json:"id"`
	HookName     string `json:"hook_name"`
	PluginCode   string `json:"plugin_code"`
	Mode         string `json:"mode"`
	ContentType  string `json:"content_type,omitempty"`
	ContentID    int64  `json:"content_id,omitempty"`
	CommunityID  int64  `json:"community_id,omitempty"`
	CategoryID   int64  `json:"category_id,omitempty"`
	ActorType    string `json:"actor_type,omitempty"`
	ActorID      int64  `json:"actor_id,omitempty"`
	UserID       int64  `json:"user_id,omitempty"`
	AdminUserID  int64  `json:"admin_user_id,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	StartedAt    string `json:"started_at,omitempty"`
	FinishedAt   string `json:"finished_at,omitempty"`
	DurationMS   int    `json:"duration_ms"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	Blocking     bool   `json:"blocking"`
	Metadata     string `json:"metadata_json,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// HookStats summarizes HookBus execution observability for a plugin hook.
type HookStats struct {
	HookName       string  `json:"hook_name"`
	PluginCode     string  `json:"plugin_code"`
	Mode           string  `json:"mode"`
	Blocking       bool    `json:"blocking"`
	ExecutionCount int     `json:"execution_count"`
	FailureCount   int     `json:"failure_count"`
	AvgDurationMS  float64 `json:"avg_duration_ms"`
	LastExecutedAt string  `json:"last_executed_at,omitempty"`
	LastFailedAt   string  `json:"last_failed_at,omitempty"`
	LastError      string  `json:"last_error,omitempty"`
}

// PluginHealth summarizes the runtime governance health of a plugin.
type PluginHealth struct {
	Status                 string `json:"status"`
	ConfigStatus           string `json:"config_status"`
	MigrationStatus        string `json:"migration_status"`
	HookStatus             string `json:"hook_status"`
	DependencyStatus       string `json:"dependency_status"`
	RecentError            string `json:"recent_error,omitempty"`
	SuggestedAction        string `json:"suggested_action,omitempty"`
	StatusReason           string `json:"status_reason,omitempty"`
	PendingMigrationsCount int    `json:"pending_migrations_count"`
	FailedMigrationsCount  int    `json:"failed_migrations_count"`
	HookFailureCount       int    `json:"hook_failure_count"`
	LastHookError          string `json:"last_hook_error,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
}

// Plugin 描述系统插件的注册与运行状态。
type Plugin struct {
	PluginManifest
	Status          string        `json:"status"`
	GlobalStatus    string        `json:"global_status,omitempty"`
	CommunityStatus string        `json:"community_status,omitempty"`
	SortOrder       int           `json:"sort_order,omitempty"`
	ConfigJSON      string        `json:"config_json,omitempty"`
	ResolvedConfig  any           `json:"resolved_config,omitempty"`
	Health          *PluginHealth `json:"health,omitempty"`
	CreatedAt       string        `json:"created_at,omitempty"`
	UpdatedAt       string        `json:"updated_at,omitempty"`
}

// PluginImpact summarizes the governance impact scope for disabling/enabling a plugin.
// It intentionally stays lightweight: only counts that are cheap and stable to compute.
type PluginImpact struct {
	PluginCode               string `json:"plugin_code"`
	ExistingContentsCount    int    `json:"existing_contents_count"`
	EnabledCommunitiesCount  int    `json:"enabled_communities_count"`
	DisabledCommunitiesCount int    `json:"disabled_communities_count"`
	CategoriesCount          int    `json:"categories_count"`
	TopicsCount              int    `json:"topics_count"`
	RecentContentsCount      int    `json:"recent_contents_count"`
	PendingTopicsCount       int    `json:"pending_topics_count"`
	PendingContentsCount     int    `json:"pending_contents_count"`
	MenusCount               int    `json:"menus_count"`
	FrontendMenusCount       int    `json:"frontend_menus_count"`
	ModeratorMenusCount      int    `json:"moderator_menus_count"`
	AdminMenusCount          int    `json:"admin_menus_count"`
	ConfigsCount             int    `json:"configs_count"`
	PendingMigrationsCount   int    `json:"pending_migrations_count"`
	RecentHookErrorsCount    int    `json:"recent_hook_errors_count"`
}

// CommunityPlugin 表示子站对某个插件的启用状态与配置。
type CommunityPlugin struct {
	ID          int64  `json:"id"`
	CommunityID int64  `json:"community_id"`
	PluginCode  string `json:"plugin_code"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
	ConfigJSON  string `json:"config_json,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// PluginMenu 是 MenuDefinition 的兼容别名。
type PluginMenu = MenuDefinition

// PluginPermission 是 PermissionDefinition 的兼容别名。
type PluginPermission = PermissionDefinition

// PluginRoute 是 RouteDefinition 的兼容别名。
type PluginRoute = RouteDefinition

// TagStat 表示标签及其关联内容数量。
type TagStat struct {
	ID             int64  `json:"id,omitempty"`
	Name           string `json:"name"`
	Slug           string `json:"slug,omitempty"`
	Site           string `json:"site,omitempty"`
	CommunityID    int64  `json:"community_id,omitempty"`
	CommunitySlug  string `json:"community_slug,omitempty"`
	Description    string `json:"description,omitempty"`
	TopicCount     int    `json:"topic_count,omitempty"`
	Count          int    `json:"count"`
	FollowerCount  int    `json:"follower_count,omitempty"`
	HotScore       int    `json:"hot_score,omitempty"`
	Status         string `json:"status,omitempty"`
	MatchedAlias   string `json:"matched_alias,omitempty"`
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	SEOKeywords    string `json:"seo_keywords,omitempty"`
}

// Tag 表示可被后台管理的内容标签。
type Tag struct {
	ID             int64  `json:"id"`
	Site           string `json:"site"`
	CommunityID    int64  `json:"community_id,omitempty"`
	CommunitySlug  string `json:"community_slug,omitempty"`
	CommunityName  string `json:"community_name,omitempty"`
	Name           string `json:"name" binding:"required"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	Sort           int    `json:"sort"`
	SortOrder      int    `json:"sort_order,omitempty"`
	UseCount       int    `json:"use_count"`
	TopicCount     int    `json:"topic_count,omitempty"`
	FollowerCount  int    `json:"follower_count,omitempty"`
	HotScore       int    `json:"hot_score,omitempty"`
	MergedToID     int64  `json:"merged_to_id,omitempty"`
	MergedToName   string `json:"merged_to_name,omitempty"`
	MergedToSlug   string `json:"merged_to_slug,omitempty"`
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	SEOKeywords    string `json:"seo_keywords,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// TagAlias 表示标签别名。
type TagAlias struct {
	ID            int64  `json:"id"`
	TagID         int64  `json:"tag_id"`
	Site          string `json:"site"`
	CommunityID   int64  `json:"community_id,omitempty"`
	CommunitySlug string `json:"community_slug,omitempty"`
	Alias         string `json:"alias"`
	AliasSlug     string `json:"alias_slug"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// TagResolveResult 表示标签 slug/别名/合并态解析结果。
type TagResolveResult struct {
	Tag          Tag    `json:"tag"`
	MatchedAlias string `json:"matched_alias,omitempty"`
	Requested    string `json:"requested,omitempty"`
	CanonicalURL string `json:"canonical_url,omitempty"`
	RedirectURL  string `json:"redirect_url,omitempty"`
	ResolvedBy   string `json:"resolved_by,omitempty"`
}

// Post 表示社区帖子、文档、Wiki 等内容实体。
type Post struct {
	ID            int64    `json:"id"`
	UserID        int64    `json:"user_id,omitempty"`
	Site          string   `json:"site"`
	Board         string   `json:"board"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	Author        string   `json:"author"`
	Status        string   `json:"status"`
	Pinned        bool     `json:"pinned"`
	Recommended   bool     `json:"recommended"`
	CommentLocked bool     `json:"comment_locked"`
	RejectReason  string   `json:"reject_reason,omitempty"`
	OfflineReason string   `json:"offline_reason,omitempty"`
	Views         int      `json:"views"`
	Likes         int      `json:"likes"`
	Comments      int      `json:"comments"`
	Tags          []string `json:"tags"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// Comment 表示帖子下的评论节点，Replies 用于返回树形评论结构。
type Comment struct {
	ID            int64      `json:"id"`
	PostID        int64      `json:"post_id"`
	TopicID       int64      `json:"topic_id,omitempty"`
	ParentID      int64      `json:"parent_id"`
	ReplyToUserID int64      `json:"reply_to_user_id,omitempty"`
	UserID        int64      `json:"user_id,omitempty"`
	UserName      string     `json:"user_name,omitempty"`
	Author        string     `json:"author"`
	To            string     `json:"to,omitempty"`
	Text          string     `json:"text"`
	Content       string     `json:"content"`
	ContentHTML   string     `json:"content_html,omitempty"`
	Status        string     `json:"status"`
	Likes         int        `json:"likes"`
	LikeCount     int        `json:"like_count"`
	IsBest        bool       `json:"is_best"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at,omitempty"`
	Replies       []*Comment `json:"replies,omitempty"`
}

// Notification 表示站内通知。
type Notification struct {
	ID          int64  `json:"id"`
	Site        string `json:"site,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	ActorUserID int64  `json:"actor_user_id,omitempty"`
	Type        string `json:"type,omitempty"`
	TargetType  string `json:"target_type,omitempty"`
	TargetID    int64  `json:"target_id,omitempty"`
	TopicID     int64  `json:"topic_id,omitempty"`
	CommentID   int64  `json:"comment_id,omitempty"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Read        bool   `json:"read"`
	IsRead      bool   `json:"is_read"`
	TargetURL   string `json:"target_url,omitempty"`
	CreatedAt   string `json:"created_at"`
	ReadAt      string `json:"read_at,omitempty"`
}

// UserProfile 表示当前前台用户的个人资料和内容统计。
type UserProfile struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	Posts    int    `json:"posts"`
	Comments int    `json:"comments"`
	Likes    int    `json:"likes"`
}

// HealthStatus 表示服务和当前数据仓储的健康信息。
type HealthStatus struct {
	OK       bool           `json:"ok"`
	Time     string         `json:"time"`
	Store    string         `json:"store"`
	Database string         `json:"database,omitempty"`
	Counts   map[string]int `json:"counts,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// PostStats 汇总帖子相关核心指标。
type PostStats struct {
	TotalPosts    int `json:"total_posts"`
	TotalViews    int `json:"total_views"`
	TotalLikes    int `json:"total_likes"`
	TotalComments int `json:"total_comments"`
}

// SiteOverview 聚合子站首页所需的站点、板块、统计和内容列表。
type SiteOverview struct {
	Site        Site           `json:"site"`
	Boards      []Board        `json:"boards"`
	BoardCounts map[string]int `json:"board_counts"`
	Stats       PostStats      `json:"stats"`
	Tags        []TagStat      `json:"tags"`
	HotPosts    []Post         `json:"hot_posts"`
	LatestPosts []Post         `json:"latest_posts"`
}

// PageResponse 是列表接口通用分页响应。
type PageResponse struct {
	Items    any  `json:"items"`
	Total    int  `json:"total"`
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	HasMore  bool `json:"has_more,omitempty"`
	Filters  any  `json:"filters,omitempty"`
}

// CreatePostRequest 是创建帖子的请求体。
type CreatePostRequest struct {
	Site        string   `json:"site" binding:"required"`
	Board       string   `json:"board" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content" binding:"required"`
	Author      string   `json:"author"`
	Status      string   `json:"status"`
	Pinned      bool     `json:"pinned"`
	Recommended bool     `json:"recommended"`
	Tags        []string `json:"tags"`
}

// UpdatePostRequest 是更新帖子的请求体，指针字段用于区分未传值和传入零值。
type UpdatePostRequest struct {
	Site          *string   `json:"site"`
	Board         *string   `json:"board"`
	Title         *string   `json:"title"`
	Summary       *string   `json:"summary"`
	Content       *string   `json:"content"`
	Status        *string   `json:"status"`
	Pinned        *bool     `json:"pinned"`
	Recommended   *bool     `json:"recommended"`
	RejectReason  *string   `json:"reject_reason"`
	OfflineReason *string   `json:"offline_reason"`
	Tags          *[]string `json:"tags"`
}

// CreateCommentRequest 是创建评论或回复的请求体。
type CreateCommentRequest struct {
	UserID        int64  `json:"user_id"`
	Author        string `json:"author"`
	Text          string `json:"text"`
	Content       string `json:"content"`
	ParentID      int64  `json:"parent_id"`
	ActorUserID   int64  `json:"-"`
	ActorUserName string `json:"-"`
}

// AdminUser 表示后台用户管理列表中的用户信息。
type AdminUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	PasswordHash  string `json:"-"`
	Status        string `json:"status"`
	RoleID        int64  `json:"role_id"`
	RoleName      string `json:"role_name"`
	Posts         int    `json:"posts"`
	Comments      int    `json:"comments"`
	CreatedAt     string `json:"created_at"`
	LastLoginAt   string `json:"last_login_at"`
	ViolationNote string `json:"violation_note,omitempty"`
}

// AdminLoginRequest 是后台登录请求体。
type AdminLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
}

// RefreshTokenRequest 是刷新 Access Token 的请求体。
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterRequest 是前台用户注册请求体。
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}

// AdminSession 表示登录成功后返回的轻量会话信息。
type AdminSession struct {
	Token        string         `json:"token"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	ExpiresIn    int64          `json:"expires_in"`
	User         AdminLoginUser `json:"user"`
	TokenType    string         `json:"token_type,omitempty"`
	Audience     string         `json:"aud,omitempty"`
}

// AdminLoginUser 表示后台当前登录用户信息。
type AdminLoginUser struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Email       string   `json:"email,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Role        string   `json:"role"`
	RoleCode    string   `json:"role_code"`
	Sites       []string `json:"sites"`
	Permissions []string `json:"permissions,omitempty"`
}

// AuthUser 是认证中间件解析出的当前用户上下文。
type AuthUser struct {
	ID                   int64                `json:"id"`
	Username             string               `json:"username"`
	Nickname             string               `json:"nickname"`
	Email                string               `json:"email,omitempty"`
	Phone                string               `json:"phone,omitempty"`
	Status               string               `json:"status"`
	RoleCode             string               `json:"role_code"`
	RoleName             string               `json:"role_name"`
	Sites                []string             `json:"sites"`
	Permissions          []string             `json:"permissions"`
	IsModerator          bool                 `json:"is_moderator"`
	ModeratedCommunities []CommunityModerator `json:"moderated_communities,omitempty"`
	TokenType            string               `json:"token_type,omitempty"`
	Audience             string               `json:"aud,omitempty"`
	Identity             string               `json:"identity,omitempty"`
}

// ActorContext 是服务端从 token / admin / moderator scope 计算出的可信权限上下文。
// 客户端请求体不能覆盖该结构。
type ActorContext struct {
	UserID          int64    `json:"user_id,omitempty"`
	AdminID         int64    `json:"admin_id,omitempty"`
	IsAdmin         bool     `json:"is_admin"`
	IsModerator     bool     `json:"is_moderator"`
	CommunityScopes []int64  `json:"community_scopes,omitempty"`
	Sites           []string `json:"sites,omitempty"`
	Permissions     []string `json:"permissions,omitempty"`
	TokenType       string   `json:"token_type,omitempty"`
	RoleCode        string   `json:"role_code,omitempty"`
}

// AdminContext 表示后台请求解析后的权限上下文。
type AdminContext struct {
	CurrentUser AuthUser `json:"current_user"`
	CurrentSite string   `json:"current_site"`
	IsGlobal    bool     `json:"is_global"`
	Permissions []string `json:"permissions"`
}

// AdminRole 表示后台角色及其权限集合。
type AdminRole struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Builtin     bool     `json:"builtin"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	UserCount   int      `json:"user_count"`
}

// AdminPermission 表示后台权限点及其支持的操作。
type AdminPermission struct {
	Code   string   `json:"code"`
	Module string   `json:"module"`
	Name   string   `json:"name"`
	Ops    []string `json:"ops"`
}

// AdminComment 表示后台评论审核列表中的评论信息。
type AdminComment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	PostTitle string `json:"post_title"`
	ParentID  int64  `json:"parent_id"`
	Author    string `json:"author"`
	To        string `json:"to,omitempty"`
	Text      string `json:"text"`
	Status    string `json:"status"`
	Likes     int    `json:"likes"`
	CreatedAt string `json:"created_at"`
}

// AdminLog 表示后台操作日志。
type AdminLog struct {
	ID          int64  `json:"id"`
	Site        string `json:"site"`
	Type        string `json:"type"`
	Actor       string `json:"actor"`
	ActorType   string `json:"actor_type"`
	ActorUserID int64  `json:"actor_user_id"`
	ActorID     int64  `json:"actor_id"`
	Role        string `json:"role,omitempty"`
	Action      string `json:"action"`
	Target      string `json:"target"`
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	CommunityID int64  `json:"community_id"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
	Metadata    string `json:"metadata_json,omitempty"`
	IP          string `json:"ip"`
	CreatedAt   string `json:"created_at"`
}

// AdminSettings 表示后台基础参数配置。
type AdminSettings struct {
	SiteName          string `json:"site_name"`
	Copyright         string `json:"copyright"`
	DefaultPageSize   int    `json:"default_page_size"`
	ReviewTimeoutHour int    `json:"review_timeout_hour"`
	PasswordRule      string `json:"password_rule"`
	CaptchaEnabled    bool   `json:"captcha_enabled"`
	SearchDefault     string `json:"search_default"`
	SearchSort        string `json:"search_sort"`
	HotViewWeight     int    `json:"hot_view_weight"`
	HotLikeWeight     int    `json:"hot_like_weight"`
	HotCommentWeight  int    `json:"hot_comment_weight"`
}

// AdminOverview 聚合后台首页需要的内容、站点、板块、搜索和用户统计。
type AdminOverview struct {
	Stats              PostStats      `json:"stats"`
	StatusDistribution map[string]int `json:"status_distribution"`
	SiteStats          []SiteStat     `json:"site_stats"`
	BoardStats         []BoardStat    `json:"board_stats"`
	TopPosts           []Post         `json:"top_posts"`
	SearchKeywords     []KeywordStat  `json:"search_keywords"`
	UserStats          UserAdminStats `json:"user_stats"`
}

// SiteStat 表示单个站点维度的内容统计。
type SiteStat struct {
	Site     string `json:"site"`
	Posts    int    `json:"posts"`
	Views    int    `json:"views"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
}

// BoardStat 表示单个板块维度的内容统计。
type BoardStat struct {
	Board    string `json:"board"`
	Posts    int    `json:"posts"`
	Views    int    `json:"views"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
}

// KeywordStat 表示搜索关键词统计。
type KeywordStat struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
	Scope   string `json:"scope"`
}

// UserAdminStats 表示后台用户维度的汇总统计。
type UserAdminStats struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
	Forbidden   int `json:"forbidden"`
	NewThisWeek int `json:"new_this_week"`
}

// UpdateCommentStatusRequest 是后台更新评论审核状态的请求体。
type UpdateCommentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateUserStatusRequest 是后台更新用户状态的请求体。
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Note   string `json:"note"`
}

// PushNotificationRequest 是后台推送站内通知的请求体。
type PushNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Scope   string `json:"scope"`
	UserID  int64  `json:"user_id"`
}

// ===== 新增：DevHub 通用社区系统数据模型 =====

// Community 描述一个子站（技术社区）。
type Community struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Slug                string `json:"slug"`
	Logo                string `json:"logo"`
	CoverImage          string `json:"cover_image"`
	Slogan              string `json:"slogan"`
	Description         string `json:"description"`
	ThemeColor          string `json:"theme_color"`
	SEOTitle            string `json:"seo_title"`
	SEODescription      string `json:"seo_description"`
	SEOKeywords         string `json:"seo_keywords"`
	SortOrder           int    `json:"sort_order"`
	Status              int    `json:"status"`
	FollowerCount       int    `json:"follower_count"`
	TopicCount          int    `json:"topic_count"`
	CommentCount        int    `json:"comment_count"`
	HotScore            int    `json:"hot_score"`
	AnnouncementTitle   string `json:"announcement_title"`
	AnnouncementContent string `json:"announcement_content"`
	AnnouncementURL     string `json:"announcement_url"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

// Category 描述内容板块/分类，支持多种内容类型。
type Category struct {
	ID                  int64    `json:"id"`
	CommunityID         int64    `json:"community_id"`
	Name                string   `json:"name"`
	Slug                string   `json:"slug"`
	Type                string   `json:"type"`
	ContentType         string   `json:"content_type"`
	PluginCode          string   `json:"plugin_code"`
	AllowedContentTypes []string `json:"allowed_content_types,omitempty"`
	Description         string   `json:"description"`
	Icon                string   `json:"icon"`
	SortOrder           int      `json:"sort_order"`
	Visible             bool     `json:"visible"`
	NavVisible          bool     `json:"nav_visible"`
	Postable            bool     `json:"postable"`
	SEOTitle            string   `json:"seo_title"`
	SEODescription      string   `json:"seo_description"`
	Status              int      `json:"status"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

// Topic 表示社区主题内容，支持多种内容类型。
type Topic struct {
	ID            int64         `json:"id"`
	CommunityID   int64         `json:"community_id"`
	CategoryID    int64         `json:"category_id"`
	UserID        int64         `json:"user_id"`
	Title         string        `json:"title"`
	Slug          string        `json:"slug"`
	ContentType   string        `json:"content_type"`
	PluginCode    string        `json:"plugin_code"`
	Summary       string        `json:"summary"`
	Content       string        `json:"content"`
	AISummary     string        `json:"ai_summary,omitempty"`
	CoverImage    string        `json:"cover_image,omitempty"`
	Status        int           `json:"status"`
	IsPinned      bool          `json:"is_pinned"`
	IsFeatured    bool          `json:"is_featured"`
	IsSolved      bool          `json:"is_solved"`
	CommentLocked bool          `json:"comment_locked"`
	RejectReason  string        `json:"reject_reason,omitempty"`
	OfflineReason string        `json:"offline_reason,omitempty"`
	BestCommentID int64         `json:"best_comment_id,omitempty"`
	ViewCount     int           `json:"view_count"`
	CommentCount  int           `json:"comment_count"`
	LikeCount     int           `json:"like_count"`
	FavoriteCount int           `json:"favorite_count"`
	HotScore      int           `json:"hot_score"`
	LastActiveAt  string        `json:"last_active_at,omitempty"`
	Tags          []string      `json:"tags,omitempty"`
	Liked         bool          `json:"liked,omitempty"`
	Favorited     bool          `json:"favorited,omitempty"`
	Followed      bool          `json:"followed,omitempty"`
	CanEdit       bool          `json:"can_edit,omitempty"`
	CanDelete     bool          `json:"can_delete,omitempty"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
	QAQuestion    *QAQuestion   `json:"qa_question,omitempty"`
	DocsDocument  *DocsDocument `json:"docs_document,omitempty"`
	WikiPage      *WikiPage     `json:"wiki_page,omitempty"`
}

// TopicTag 表示主题与标签的关联。
type TopicTag struct {
	TopicID int64 `json:"topic_id"`
	TagID   int64 `json:"tag_id"`
}

// Reaction 表示用户的点赞/表情反应。
type Reaction struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	TargetType   string `json:"target_type"`
	TargetID     int64  `json:"target_id"`
	ReactionType string `json:"reaction_type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// Favorite 表示用户的收藏。
type Favorite struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// Follow 表示用户的关注。
type Follow struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// Activity 表示用户动态。
type Activity struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	CommunityID int64  `json:"community_id,omitempty"`
	TopicID     int64  `json:"topic_id,omitempty"`
	Action      string `json:"action"`
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	Remark      string `json:"remark,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	TargetTitle string `json:"target_title,omitempty"`
	TargetURL   string `json:"target_url,omitempty"`
	Community   string `json:"community,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// TopicInteraction 表示当前用户对主题的互动状态和主题统计。
type TopicInteraction struct {
	Liked         bool `json:"liked"`
	Favorited     bool `json:"favorited"`
	Followed      bool `json:"followed"`
	LikeCount     int  `json:"like_count"`
	FavoriteCount int  `json:"favorite_count"`
	HotScore      int  `json:"hot_score"`
}

// FavoriteItem 表示“我的收藏”列表中的一条收藏及其目标内容。
type FavoriteItem struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	CreatedAt  string    `json:"created_at"`
	Topic      Topic     `json:"topic,omitempty"`
	Community  Community `json:"community,omitempty"`
	Category   Category  `json:"category,omitempty"`
	TargetURL  string    `json:"target_url,omitempty"`
}

// FollowItem 表示“我的关注”列表中的一条关注及其目标摘要。
type FollowItem struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	TargetType  string    `json:"target_type"`
	TargetID    int64     `json:"target_id"`
	TargetName  string    `json:"target_name"`
	TargetSlug  string    `json:"target_slug,omitempty"`
	TargetTitle string    `json:"target_title,omitempty"`
	TargetURL   string    `json:"target_url,omitempty"`
	Description string    `json:"description,omitempty"`
	Community   Community `json:"community,omitempty"`
	Topic       Topic     `json:"topic,omitempty"`
	CreatedAt   string    `json:"created_at"`
}

// Report 表示举报记录。
type Report struct {
	ID             int64  `json:"id"`
	ReporterID     int64  `json:"reporter_id"`
	ReporterUserID int64  `json:"reporter_user_id"`
	TargetType     string `json:"target_type"`
	TargetID       int64  `json:"target_id"`
	CommunityID    int64  `json:"community_id,omitempty"`
	CommunitySlug  string `json:"community_slug,omitempty"`
	CommunityName  string `json:"community_name,omitempty"`
	TopicID        int64  `json:"topic_id,omitempty"`
	ReasonType     string `json:"reason_type"`
	ReasonText     string `json:"reason_text,omitempty"`
	Status         string `json:"status"`
	HandledBy      int64  `json:"handled_by,omitempty"`
	HandledAt      string `json:"handled_at,omitempty"`
	HandleNote     string `json:"handle_note,omitempty"`
	TargetTitle    string `json:"target_title,omitempty"`
	TargetContent  string `json:"target_content,omitempty"`
	TargetURL      string `json:"target_url,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// CommunityModerator 表示子站版主。
type CommunityModerator struct {
	ID            int64  `json:"id"`
	CommunityID   int64  `json:"community_id"`
	CommunitySlug string `json:"community_slug,omitempty"`
	CommunityName string `json:"community_name,omitempty"`
	UserID        int64  `json:"user_id"`
	UserName      string `json:"user_name,omitempty"`
	UserNickname  string `json:"user_nickname,omitempty"`
	Role          string `json:"role"`
	Status        int    `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// ModeratorDashboard 表示独立版主工作台概览。
type ModeratorDashboard struct {
	ManagedCommunities []Community `json:"managed_communities"`
	PendingReportCount int         `json:"pending_report_count"`
	TopicCount         int         `json:"topic_count"`
	CommentCount       int         `json:"comment_count"`
	TodayActionCount   int         `json:"today_action_count"`
	RecentReports      []Report    `json:"recent_reports"`
	RecentAuditLogs    []AdminLog  `json:"recent_audit_logs"`
}

// CreateReportRequest 是前台创建举报的请求体。
type CreateReportRequest struct {
	ReporterUserID int64  `json:"-"`
	TargetType     string `json:"target_type" binding:"required"`
	TargetID       int64  `json:"target_id" binding:"required"`
	ReasonType     string `json:"reason_type" binding:"required"`
	ReasonText     string `json:"reason_text"`
}

// ReportFilter 是后台举报列表筛选条件。
type ReportFilter struct {
	Status        string `json:"status"`
	TargetType    string `json:"target_type"`
	CommunitySlug string `json:"community_slug"`
	CommunityID   int64  `json:"community_id"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	ActorUserID   int64  `json:"-"`
	ActorIsAdmin  bool   `json:"-"`
}

// HandleReportRequest 是处理举报的请求体。
type HandleReportRequest struct {
	Status     string `json:"status" binding:"required"`
	HandleNote string `json:"handle_note"`
}

// CommunityModeratorFilter 是后台版主列表筛选条件。
type CommunityModeratorFilter struct {
	CommunitySlug string `json:"community_slug"`
	CommunityID   int64  `json:"community_id"`
	UserID        int64  `json:"user_id"`
	Status        string `json:"status"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	ActorUserID   int64  `json:"-"`
	ActorIsAdmin  bool   `json:"-"`
}

// CommunityModeratorRequest 是新增或更新版主的请求体。
type CommunityModeratorRequest struct {
	CommunityID   int64  `json:"community_id"`
	CommunitySlug string `json:"community_slug"`
	UserID        int64  `json:"user_id"`
	Role          string `json:"role"`
	Status        *int   `json:"status"`
}

// CommunityRequest 是后台新增或更新子站的请求体。
type CommunityRequest struct {
	Name                string `json:"name"`
	Slug                string `json:"slug"`
	Logo                string `json:"logo"`
	CoverImage          string `json:"cover_image"`
	Slogan              string `json:"slogan"`
	Description         string `json:"description"`
	ThemeColor          string `json:"theme_color"`
	SEOTitle            string `json:"seo_title"`
	SEODescription      string `json:"seo_description"`
	SEOKeywords         string `json:"seo_keywords"`
	SortOrder           *int   `json:"sort_order"`
	Status              *int   `json:"status"`
	AnnouncementTitle   string `json:"announcement_title"`
	AnnouncementContent string `json:"announcement_content"`
	AnnouncementURL     string `json:"announcement_url"`
}

// CategoryRequest 是后台新增或更新子站板块的请求体。
type CategoryRequest struct {
	CommunityID         int64    `json:"community_id"`
	Name                string   `json:"name"`
	Slug                string   `json:"slug"`
	Type                string   `json:"type"`
	ContentType         string   `json:"content_type"`
	PluginCode          string   `json:"plugin_code"`
	AllowedContentTypes []string `json:"allowed_content_types"`
	Description         string   `json:"description"`
	Icon                string   `json:"icon"`
	SortOrder           *int     `json:"sort_order"`
	Visible             *bool    `json:"visible"`
	NavVisible          *bool    `json:"nav_visible"`
	Postable            *bool    `json:"postable"`
	SEOTitle            string   `json:"seo_title"`
	SEODescription      string   `json:"seo_description"`
	Status              *int     `json:"status"`
}

// ReorderRequest 表示后台排序请求。
type ReorderRequest struct {
	IDs []int64 `json:"ids"`
}

// TagAliasRequest 是后台新增标签别名的请求体。
type TagAliasRequest struct {
	Alias string `json:"alias" binding:"required"`
}

// TagMergeRequest 是后台标签合并请求体。
type TagMergeRequest struct {
	TargetTagID int64  `json:"target_tag_id" binding:"required"`
	Note        string `json:"note"`
}

// CommunityStats 表示子站统计信息。
type CommunityStats struct {
	TopicCount        int `json:"topic_count"`
	CommentCount      int `json:"comment_count"`
	QuestionCount     int `json:"question_count"`
	UnsolvedCount     int `json:"unsolved_count"`
	FollowerCount     int `json:"follower_count"`
	TodayTopicCount   int `json:"today_topic_count"`
	TodayCommentCount int `json:"today_comment_count"`
	ModeratorCount    int `json:"moderator_count"`
	HotScore          int `json:"hot_score"`
}

// BatchModerationRequest 是后台批量治理请求体。
type BatchModerationRequest struct {
	IDs        []int64 `json:"ids" binding:"required"`
	Action     string  `json:"action"`
	Status     string  `json:"status"`
	Note       string  `json:"note"`
	HandleNote string  `json:"handle_note"`
}

// BatchModerationItem 表示批量治理中单个对象的处理结果。
type BatchModerationItem struct {
	ID    int64  `json:"id"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// BatchModerationResponse 表示批量治理汇总结果。
type BatchModerationResponse struct {
	Action  string                `json:"action"`
	Updated int                   `json:"updated"`
	Failed  int                   `json:"failed"`
	Items   []BatchModerationItem `json:"items"`
}

// AdminLogFilter 是后台治理审计日志筛选条件。
type AdminLogFilter struct {
	Site        string `json:"site"`
	Type        string `json:"type"`
	ActorType   string `json:"actor_type"`
	Action      string `json:"action"`
	Target      string `json:"target"`
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	PluginCode  string `json:"plugin_code"`
	Actor       string `json:"actor"`
	ActorID     int64  `json:"actor_user_id"`
	CommunityID int64  `json:"community_id"`
	Metadata    string `json:"metadata"`
	RequestID   string `json:"request_id"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}

// ModerationResult 表示治理动作后的目标状态。
type ModerationResult struct {
	Topic   *Topic   `json:"topic,omitempty"`
	Comment *Comment `json:"comment,omitempty"`
	Report  *Report  `json:"report,omitempty"`
	Action  string   `json:"action,omitempty"`
	Changed bool     `json:"changed"`
}

// WikiPage 表示 Wiki 页面。
type WikiPage struct {
	ID               int64  `json:"id"`
	SpaceID          int64  `json:"space_id,omitempty"`
	TopicID          int64  `json:"topic_id,omitempty"`
	CommunityID      int64  `json:"community_id"`
	CategoryID       int64  `json:"category_id,omitempty"`
	UserID           int64  `json:"user_id"`
	Title            string `json:"title"`
	Slug             string `json:"slug"`
	Summary          string `json:"summary"`
	Content          string `json:"content"`
	CurrentVersionID int64  `json:"current_version_id,omitempty"`
	Status           int    `json:"status"`
	ViewCount        int    `json:"view_count"`
	LikeCount        int    `json:"like_count"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// WikiRevision 表示 Wiki 版本记录。
type WikiRevision struct {
	ID         int64  `json:"id"`
	WikiPageID int64  `json:"wiki_page_id"`
	TopicID    int64  `json:"topic_id,omitempty"`
	EditorID   int64  `json:"editor_id"`
	VersionNo  int    `json:"version_no,omitempty"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ChangeNote string `json:"change_note,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// QAQuestion 表示 question 内容的扩展行。
type QAQuestion struct {
	ID               int64  `json:"id"`
	TopicID          int64  `json:"topic_id"`
	AnswerCount      int    `json:"answer_count"`
	IsResolved       bool   `json:"is_resolved"`
	AcceptedAnswerID int64  `json:"accepted_answer_id,omitempty"`
	AcceptedAt       string `json:"accepted_at,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

// QAAnswer 表示基于评论映射出的回答扩展行。
type QAAnswer struct {
	ID         int64  `json:"id"`
	TopicID    int64  `json:"topic_id"`
	CommentID  int64  `json:"comment_id"`
	UserID     int64  `json:"user_id"`
	IsAccepted bool   `json:"is_accepted"`
	AcceptedAt string `json:"accepted_at,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// DocsSpace 表示文档空间。
type DocsSpace struct {
	ID          int64  `json:"id"`
	CommunityID int64  `json:"community_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// DocsDocument 表示文档扩展行。
type DocsDocument struct {
	ID         int64  `json:"id"`
	SpaceID    int64  `json:"space_id,omitempty"`
	TopicID    int64  `json:"topic_id"`
	ParentID   int64  `json:"parent_id,omitempty"`
	SortOrder  int    `json:"sort_order"`
	Status     int    `json:"status"`
	Version    int    `json:"version"`
	EditorType string `json:"editor_type,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// ===== 新增：请求/响应类型 =====

// CreateTopicRequest 是创建主题的请求体。
type CreateTopicRequest struct {
	UserID        int64    `json:"user_id"`
	CommunityID   int64    `json:"community_id"`
	CommunitySlug string   `json:"community_slug"`
	CategoryID    int64    `json:"category_id"`
	Title         string   `json:"title"`
	ContentType   string   `json:"content_type"`
	PluginCode    string   `json:"plugin_code"`
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	Status        string   `json:"status"`
	TagIDs        []int64  `json:"tag_ids"`
	Tags          []string `json:"tags"`

	// ActorPermissions is derived from the authenticated user context.
	// It is not part of the public API payload.
	ActorPermissions []string     `json:"-"`
	ActorContext     ActorContext `json:"-"`
}

// UpdateTopicRequest 是更新主题的请求体。
type UpdateTopicRequest struct {
	CommunityID   *int64    `json:"community_id"`
	CommunitySlug *string   `json:"community_slug"`
	CategoryID    *int64    `json:"category_id"`
	Title         *string   `json:"title"`
	ContentType   *string   `json:"content_type"`
	PluginCode    *string   `json:"plugin_code"`
	Summary       *string   `json:"summary"`
	Content       *string   `json:"content"`
	Status        *int      `json:"status"`
	IsPinned      *bool     `json:"is_pinned"`
	IsFeatured    *bool     `json:"is_featured"`
	IsSolved      *bool     `json:"is_solved"`
	CommentLocked *bool     `json:"comment_locked"`
	Tags          *[]string `json:"tags"`

	ActorContext ActorContext `json:"-"`
}

// ToggleReactionRequest 是切换点赞/收藏/关注的请求体。
type ToggleReactionRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   int64  `json:"target_id" binding:"required"`
}

// CommunityOverview 表示子站首页概览。
type CommunityOverview struct {
	Community      Community      `json:"community"`
	Categories     []Category     `json:"categories"`
	CategoryCounts map[string]int `json:"category_counts"`
	HotTopics      []Topic        `json:"hot_topics"`
	LatestTopics   []Topic        `json:"latest_topics"`
	HotTags        []TagStat      `json:"hot_tags"`
	Stats          PostStats      `json:"stats"`
}

// SearchRequest 是搜索请求参数。
type SearchRequest struct {
	Keyword       string `json:"keyword"`
	Scope         string `json:"scope"`
	CommunitySlug string `json:"community_slug"`
	CommunityID   int64  `json:"community_id"`
	CategoryID    int64  `json:"category_id"`
	PluginCode    string `json:"plugin_code"`
	ContentType   string `json:"content_type"`
	Tag           string `json:"tag"`
	TagID         int64  `json:"tag_id"`
	Sort          string `json:"sort"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}
