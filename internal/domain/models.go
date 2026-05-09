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

// TagStat 表示标签及其关联内容数量。
type TagStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Tag 表示可被后台管理的内容标签。
type Tag struct {
	ID          int64  `json:"id"`
	Site        string `json:"site"`
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Sort        int    `json:"sort"`
	UseCount    int    `json:"use_count"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
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

// AdminSession 表示后台登录成功后返回的轻量会话信息。
type AdminSession struct {
	Token        string         `json:"token"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	ExpiresIn    int64          `json:"expires_in"`
	User         AdminLoginUser `json:"user"`
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
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Email       string   `json:"email,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Status      string   `json:"status"`
	RoleCode    string   `json:"role_code"`
	RoleName    string   `json:"role_name"`
	Sites       []string `json:"sites"`
	Permissions []string `json:"permissions"`
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
	ActorUserID int64  `json:"actor_user_id"`
	Role        string `json:"role,omitempty"`
	Action      string `json:"action"`
	Target      string `json:"target"`
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	CommunityID int64  `json:"community_id"`
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
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Category 描述内容板块/分类，支持多种内容类型。
type Category struct {
	ID          int64  `json:"id"`
	CommunityID int64  `json:"community_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
	Visible     bool   `json:"visible"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Topic 表示社区主题内容，支持多种内容类型。
type Topic struct {
	ID            int64    `json:"id"`
	CommunityID   int64    `json:"community_id"`
	CategoryID    int64    `json:"category_id"`
	UserID        int64    `json:"user_id"`
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	ContentType   string   `json:"content_type"`
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	AISummary     string   `json:"ai_summary,omitempty"`
	CoverImage    string   `json:"cover_image,omitempty"`
	Status        int      `json:"status"`
	IsPinned      bool     `json:"is_pinned"`
	IsFeatured    bool     `json:"is_featured"`
	IsSolved      bool     `json:"is_solved"`
	CommentLocked bool     `json:"comment_locked"`
	RejectReason  string   `json:"reject_reason,omitempty"`
	OfflineReason string   `json:"offline_reason,omitempty"`
	BestCommentID int64    `json:"best_comment_id,omitempty"`
	ViewCount     int      `json:"view_count"`
	CommentCount  int      `json:"comment_count"`
	LikeCount     int      `json:"like_count"`
	FavoriteCount int      `json:"favorite_count"`
	HotScore      int      `json:"hot_score"`
	LastActiveAt  string   `json:"last_active_at,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Liked         bool     `json:"liked,omitempty"`
	Favorited     bool     `json:"favorited,omitempty"`
	Followed      bool     `json:"followed,omitempty"`
	CanEdit       bool     `json:"can_edit,omitempty"`
	CanDelete     bool     `json:"can_delete,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
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
	Action      string `json:"action"`
	Target      string `json:"target"`
	TargetType  string `json:"target_type"`
	Actor       string `json:"actor"`
	ActorID     int64  `json:"actor_user_id"`
	CommunityID int64  `json:"community_id"`
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
	ID          int64  `json:"id"`
	CommunityID int64  `json:"community_id"`
	CategoryID  int64  `json:"category_id,omitempty"`
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Summary     string `json:"summary"`
	Content     string `json:"content"`
	Status      int    `json:"status"`
	ViewCount   int    `json:"view_count"`
	LikeCount   int    `json:"like_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WikiRevision 表示 Wiki 版本记录。
type WikiRevision struct {
	ID         int64  `json:"id"`
	WikiPageID int64  `json:"wiki_page_id"`
	EditorID   int64  `json:"editor_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ChangeNote string `json:"change_note,omitempty"`
	CreatedAt  string `json:"created_at"`
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
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	Status        string   `json:"status"`
	TagIDs        []int64  `json:"tag_ids"`
	Tags          []string `json:"tags"`
}

// UpdateTopicRequest 是更新主题的请求体。
type UpdateTopicRequest struct {
	CommunityID   *int64    `json:"community_id"`
	CommunitySlug *string   `json:"community_slug"`
	CategoryID    *int64    `json:"category_id"`
	Title         *string   `json:"title"`
	ContentType   *string   `json:"content_type"`
	Summary       *string   `json:"summary"`
	Content       *string   `json:"content"`
	Status        *int      `json:"status"`
	IsPinned      *bool     `json:"is_pinned"`
	IsFeatured    *bool     `json:"is_featured"`
	IsSolved      *bool     `json:"is_solved"`
	CommentLocked *bool     `json:"comment_locked"`
	Tags          *[]string `json:"tags"`
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
	ContentType   string `json:"content_type"`
	Tag           string `json:"tag"`
	TagID         int64  `json:"tag_id"`
	Sort          string `json:"sort"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}
