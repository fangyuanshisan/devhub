package store

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

// TimeLayout 是接口中时间字符串的统一格式。
const TimeLayout = "2006-01-02 15:04:05"

// MemoryStore 是线程安全的内存数据仓储，用于演示环境和本地开发。
type MemoryStore struct {
	mu                  sync.RWMutex
	nextPostID          int64
	nextCommentID       int64
	nextNoticeID        int64
	nextLogID           int64
	nextReactionID      int64
	nextFavoriteID      int64
	nextFollowID        int64
	nextActivityID      int64
	nextReportID        int64
	nextModeratorID     int64
	nextCommunityID     int64
	nextCategoryID      int64
	nextTagID           int64
	nextTagAliasID      int64
	nextUserID          int64
	nextHookExecutionID int64
	sites               map[string]domain.Site
	boards              map[string]domain.Board
	communities         map[int64]*domain.Community
	categories          map[int64]*domain.Category
	plugins             map[string]*domain.Plugin
	communityPlugins    map[int64]map[string]*domain.CommunityPlugin
	pluginMigrations    map[string][]domain.PluginMigration // plugin_code -> records
	hookExecutions      []domain.HookExecution
	qaQuestions         map[int64]*domain.QAQuestion
	qaAnswers           map[int64]*domain.QAAnswer
	docsSpaces          map[int64]*domain.DocsSpace
	docsDocuments       map[int64]*domain.DocsDocument
	wikiPages           map[int64]*domain.WikiPage
	wikiVersions        map[int64]*domain.WikiRevision
	tags                map[int64]*domain.Tag
	tagAliases          map[int64]*domain.TagAlias
	boardOrder          []string
	siteOrder           []string
	posts               map[int64]*domain.Post
	comments            map[int64]*domain.Comment
	notices             map[int64]*domain.Notification
	reactions           map[string]*domain.Reaction
	favorites           map[string]*domain.Favorite
	follows             map[string]*domain.Follow
	activities          map[int64]*domain.Activity
	reports             map[int64]*domain.Report
	moderators          map[int64]*domain.CommunityModerator
	commentLocks        map[int64]bool
	users               map[int64]*domain.AdminUser
	roles               map[int64]domain.AdminRole
	settings            domain.AdminSettings
	logs                []domain.AdminLog
}

func countPluginMenus(def domain.Plugin, area string) int {
	n := 0
	for _, m := range def.Menus {
		a := strings.TrimSpace(m.Area)
		if a == "" {
			a = strings.TrimSpace(m.Location)
		}
		if area == "" || a == area {
			n++
		}
	}
	return n
}

func (s *MemoryStore) PluginImpact(code string) (domain.PluginImpact, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.PluginImpact{}, errors.New("plugin_code 不能为空")
	}
	p, ok := s.PluginByCode(code)
	if !ok || p.Code == "" {
		return domain.PluginImpact{}, errors.New("插件不存在")
	}

	enabledCommunities := 0
	disabledCommunities := 0
	globalEnabled := p.Status == pluginregistry.StatusEnabled
	for id := range s.communities {
		if id == 0 {
			continue
		}
		cp := s.communityPlugins[id][code]
		if globalEnabled && cp != nil && cp.Status == pluginregistry.StatusEnabled {
			enabledCommunities++
		} else {
			disabledCommunities++
		}
	}

	categories := 0
	for _, cat := range s.categories {
		if cat == nil {
			continue
		}
		if strings.TrimSpace(cat.PluginCode) == code {
			categories++
		}
	}

	topics := 0
	pending := 0
	recent := 0
	cutoff := time.Now().AddDate(0, 0, -7)
	for _, post := range s.posts {
		if post == nil {
			continue
		}
		ct := contentTypeForBoard(post.Board)
		if pluginregistry.PluginCodeForContentType(ct) != code {
			continue
		}
		topics++
		if memoryTopicStatus(post) == 2 {
			pending++
		}
		if ts, err := time.Parse(TimeLayout, post.CreatedAt); err == nil && ts.After(cutoff) {
			recent++
		}
	}
	configs := 0
	if strings.TrimSpace(p.ConfigJSON) != "" {
		configs++
	}
	for _, byPlugin := range s.communityPlugins {
		if cp := byPlugin[code]; cp != nil && strings.TrimSpace(cp.ConfigJSON) != "" {
			configs++
		}
	}
	pendingMigrations := 0
	for _, item := range s.pluginMigrations[code] {
		if item.Status == "pending" {
			pendingMigrations++
		}
	}
	recentHookErrors := s.recentHookErrorsCountLocked(code, 0)

	return domain.PluginImpact{
		PluginCode:               code,
		ExistingContentsCount:    topics,
		EnabledCommunitiesCount:  enabledCommunities,
		DisabledCommunitiesCount: disabledCommunities,
		CategoriesCount:          categories,
		TopicsCount:              topics,
		RecentContentsCount:      recent,
		PendingTopicsCount:       pending,
		PendingContentsCount:     pending,
		MenusCount:               len(p.Menus),
		FrontendMenusCount:       countPluginMenus(p, "frontend"),
		ModeratorMenusCount:      countPluginMenus(p, "moderator"),
		AdminMenusCount:          countPluginMenus(p, "admin"),
		ConfigsCount:             configs,
		PendingMigrationsCount:   pendingMigrations,
		RecentHookErrorsCount:    recentHookErrors,
	}, nil
}

func (s *MemoryStore) CommunityPluginImpact(communityID int64, code string) (domain.PluginImpact, error) {
	code = strings.TrimSpace(code)
	if communityID <= 0 {
		return domain.PluginImpact{}, errors.New("community_id 不合法")
	}
	if _, ok := s.communities[communityID]; !ok {
		return domain.PluginImpact{}, errors.New("子站不存在")
	}
	impact, err := s.PluginImpact(code)
	if err != nil {
		return domain.PluginImpact{}, err
	}

	// Narrow counts to the specified community where possible.
	enabledCommunities := 0
	disabledCommunities := 1
	global, ok := s.plugins[code]
	if ok && global != nil && global.Status == pluginregistry.StatusEnabled {
		cp := s.communityPlugins[communityID][code]
		if cp != nil && cp.Status == pluginregistry.StatusEnabled {
			enabledCommunities = 1
			disabledCommunities = 0
		}
	}
	categories := 0
	for _, cat := range s.categories {
		if cat == nil || cat.CommunityID != communityID {
			continue
		}
		if strings.TrimSpace(cat.PluginCode) == code {
			categories++
		}
	}
	topics := 0
	pending := 0
	recent := 0
	cutoff := time.Now().AddDate(0, 0, -7)
	for _, post := range s.posts {
		if post == nil {
			continue
		}
		if s.communityIDBySlugLocked(post.Site) != communityID {
			continue
		}
		ct := contentTypeForBoard(post.Board)
		if pluginregistry.PluginCodeForContentType(ct) != code {
			continue
		}
		topics++
		if memoryTopicStatus(post) == 2 {
			pending++
		}
		if ts, err := time.Parse(TimeLayout, post.CreatedAt); err == nil && ts.After(cutoff) {
			recent++
		}
	}
	configs := 0
	if global, ok := s.plugins[code]; ok && global != nil && strings.TrimSpace(global.ConfigJSON) != "" {
		configs++
	}
	if cp := s.communityPlugins[communityID][code]; cp != nil && strings.TrimSpace(cp.ConfigJSON) != "" {
		configs++
	}

	impact.EnabledCommunitiesCount = enabledCommunities
	impact.DisabledCommunitiesCount = disabledCommunities
	impact.CategoriesCount = categories
	impact.TopicsCount = topics
	impact.ExistingContentsCount = topics
	impact.RecentContentsCount = recent
	impact.PendingTopicsCount = pending
	impact.PendingContentsCount = pending
	impact.ConfigsCount = configs
	impact.RecentHookErrorsCount = s.recentHookErrorsCountLocked(code, communityID)
	return impact, nil
}

func (s *MemoryStore) recentHookErrorsCountLocked(pluginCode string, communityID int64) int {
	cutoff := time.Now().AddDate(0, 0, -7)
	count := 0
	for _, record := range s.hookExecutions {
		if record.PluginCode != pluginCode || record.Success {
			continue
		}
		if communityID > 0 && record.CommunityID != communityID {
			continue
		}
		if ts, err := time.Parse(TimeLayout, record.FinishedAt); err == nil && ts.Before(cutoff) {
			continue
		}
		count++
	}
	return count
}

// NewMemoryStore 创建内存仓储并写入演示数据。
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		nextPostID:          1,
		nextCommentID:       1,
		nextNoticeID:        1,
		nextLogID:           1,
		nextReactionID:      1,
		nextFavoriteID:      1,
		nextFollowID:        1,
		nextActivityID:      1,
		nextReportID:        1,
		nextModeratorID:     1,
		nextCommunityID:     1,
		nextCategoryID:      1,
		nextTagID:           1,
		nextTagAliasID:      1,
		nextUserID:          1,
		nextHookExecutionID: 1,
		sites:               map[string]domain.Site{},
		boards:              map[string]domain.Board{},
		communities:         map[int64]*domain.Community{},
		categories:          map[int64]*domain.Category{},
		plugins:             map[string]*domain.Plugin{},
		communityPlugins:    map[int64]map[string]*domain.CommunityPlugin{},
		pluginMigrations:    map[string][]domain.PluginMigration{},
		hookExecutions:      []domain.HookExecution{},
		qaQuestions:         map[int64]*domain.QAQuestion{},
		qaAnswers:           map[int64]*domain.QAAnswer{},
		docsSpaces:          map[int64]*domain.DocsSpace{},
		docsDocuments:       map[int64]*domain.DocsDocument{},
		wikiPages:           map[int64]*domain.WikiPage{},
		wikiVersions:        map[int64]*domain.WikiRevision{},
		tags:                map[int64]*domain.Tag{},
		tagAliases:          map[int64]*domain.TagAlias{},
		boardOrder:          []string{"all", "community", "qa", "opensource", "ai", "jobs", "wiki", "docs"},
		siteOrder:           []string{"php", "go", "java", "ai", "frontend"},
		posts:               map[int64]*domain.Post{},
		comments:            map[int64]*domain.Comment{},
		notices:             map[int64]*domain.Notification{},
		reactions:           map[string]*domain.Reaction{},
		favorites:           map[string]*domain.Favorite{},
		follows:             map[string]*domain.Follow{},
		activities:          map[int64]*domain.Activity{},
		reports:             map[int64]*domain.Report{},
		moderators:          map[int64]*domain.CommunityModerator{},
		commentLocks:        map[int64]bool{},
		users:               map[int64]*domain.AdminUser{},
		roles:               map[int64]domain.AdminRole{},
		settings: domain.AdminSettings{
			SiteName:          "DevHub",
			Copyright:         "© 2026 DevHub",
			DefaultPageSize:   20,
			ReviewTimeoutHour: 24,
			PasswordRule:      "至少 8 位，包含字母和数字",
			CaptchaEnabled:    true,
			SearchDefault:     "portal",
			SearchSort:        "time",
			HotViewWeight:     1,
			HotLikeWeight:     8,
			HotCommentWeight:  15,
		},
	}
	s.seed()
	return s
}

func (s *MemoryStore) PluginMigrations(pluginCode string) ([]domain.PluginMigration, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.pluginMigrations[pluginCode]
	out := make([]domain.PluginMigration, 0, len(items))
	out = append(out, items...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func (s *MemoryStore) AppendPluginMigration(record domain.PluginMigration) (domain.PluginMigration, error) {
	if record.MigrationVersion == "" {
		record.MigrationVersion = record.Version
	}
	if record.Version == "" {
		record.Version = record.MigrationVersion
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.Version = strings.TrimSpace(record.Version)
	record.MigrationVersion = strings.TrimSpace(record.MigrationVersion)
	record.MigrationName = strings.TrimSpace(record.MigrationName)
	if record.PluginCode == "" || record.MigrationName == "" {
		return domain.PluginMigration{}, errors.New("plugin_code 和 migration_name 不能为空")
	}
	// Prevent duplicates by (plugin_code, version, migration_name).
	for _, it := range s.pluginMigrations[record.PluginCode] {
		if it.PluginCode == record.PluginCode && it.Version == record.Version && it.MigrationName == record.MigrationName {
			return it, nil
		}
	}
	record.ID = int64(len(s.pluginMigrations[record.PluginCode]) + 1)
	if record.CreatedAt == "" {
		record.CreatedAt = Now()
	}
	s.pluginMigrations[record.PluginCode] = append(s.pluginMigrations[record.PluginCode], record)
	return record, nil
}

func (s *MemoryStore) SavePluginMigration(record domain.PluginMigration) (domain.PluginMigration, error) {
	if record.MigrationVersion == "" {
		record.MigrationVersion = record.Version
	}
	if record.Version == "" {
		record.Version = record.MigrationVersion
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	record.Version = strings.TrimSpace(record.Version)
	record.MigrationVersion = strings.TrimSpace(record.MigrationVersion)
	record.MigrationName = strings.TrimSpace(record.MigrationName)
	if record.PluginCode == "" || record.MigrationName == "" {
		return domain.PluginMigration{}, errors.New("plugin_code 和 migration_name 不能为空")
	}
	if record.Status == "" {
		record.Status = "pending"
	}
	now := Now()
	if record.CreatedAt == "" {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	items := s.pluginMigrations[record.PluginCode]
	for i, it := range items {
		if it.PluginCode == record.PluginCode && it.Version == record.Version && it.MigrationName == record.MigrationName {
			record.ID = it.ID
			if record.CreatedAt == "" {
				record.CreatedAt = it.CreatedAt
			}
			if record.CreatedAt == "" {
				record.CreatedAt = now
			}
			items[i] = record
			s.pluginMigrations[record.PluginCode] = items
			return record, nil
		}
	}
	record.ID = int64(len(items) + 1)
	s.pluginMigrations[record.PluginCode] = append(items, record)
	return record, nil
}

func (s *MemoryStore) AppendHookExecution(record domain.HookExecution) (domain.HookExecution, error) {
	record.HookName = strings.TrimSpace(record.HookName)
	record.PluginCode = strings.TrimSpace(record.PluginCode)
	if record.HookName == "" || record.PluginCode == "" {
		return domain.HookExecution{}, errors.New("hook_name 和 plugin_code 不能为空")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if record.ID == 0 {
		record.ID = s.nextHookExecutionID
		s.nextHookExecutionID++
	}
	if record.CreatedAt == "" {
		record.CreatedAt = Now()
	}
	if record.StartedAt == "" {
		record.StartedAt = record.CreatedAt
	}
	if record.FinishedAt == "" {
		record.FinishedAt = record.StartedAt
	}
	s.hookExecutions = append(s.hookExecutions, record)
	return record, nil
}

func (s *MemoryStore) HookExecutions(pluginCode string, limit int) ([]domain.HookExecution, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.HookExecution, 0)
	for i := len(s.hookExecutions) - 1; i >= 0 && len(out) < limit; i-- {
		record := s.hookExecutions[i]
		if pluginCode != "" && record.PluginCode != pluginCode {
			continue
		}
		out = append(out, record)
	}
	return out, nil
}

func (s *MemoryStore) HookStats(pluginCode string) ([]domain.HookStats, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	s.mu.RLock()
	defer s.mu.RUnlock()
	type agg struct {
		stat    domain.HookStats
		totalMS int
	}
	items := map[string]*agg{}
	for _, record := range s.hookExecutions {
		if pluginCode != "" && record.PluginCode != pluginCode {
			continue
		}
		key := record.PluginCode + "\x00" + record.HookName
		item := items[key]
		if item == nil {
			item = &agg{stat: domain.HookStats{
				HookName:   record.HookName,
				PluginCode: record.PluginCode,
				Mode:       record.Mode,
				Blocking:   record.Blocking,
			}}
			items[key] = item
		}
		item.stat.ExecutionCount++
		item.totalMS += record.DurationMS
		if record.FinishedAt > item.stat.LastExecutedAt {
			item.stat.LastExecutedAt = record.FinishedAt
		}
		if !record.Success {
			item.stat.FailureCount++
			if record.FinishedAt > item.stat.LastFailedAt {
				item.stat.LastFailedAt = record.FinishedAt
				item.stat.LastError = record.ErrorMessage
			}
		}
	}
	out := make([]domain.HookStats, 0, len(items))
	for _, item := range items {
		if item.stat.ExecutionCount > 0 {
			item.stat.AvgDurationMS = float64(item.totalMS) / float64(item.stat.ExecutionCount)
		}
		out = append(out, item.stat)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PluginCode == out[j].PluginCode {
			return out[i].HookName < out[j].HookName
		}
		return out[i].PluginCode < out[j].PluginCode
	})
	return out, nil
}

// Now 返回符合接口格式的当前时间字符串。
func Now() string { return time.Now().Format(TimeLayout) }

// Health 返回内存仓储状态和关键资源数量。
func (s *MemoryStore) Health() domain.HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return domain.HealthStatus{
		OK:    true,
		Time:  Now(),
		Store: "memory",
		Counts: map[string]int{
			"sites":         len(s.sites),
			"boards":        len(s.boards),
			"posts":         len(s.posts),
			"comments":      len(s.comments),
			"reports":       len(s.reports),
			"moderators":    len(s.moderators),
			"notifications": len(s.notices),
			"reactions":     len(s.reactions),
			"favorites":     len(s.favorites),
			"follows":       len(s.follows),
			"activities":    len(s.activities),
			"users":         len(s.users),
			"communities":   len(s.communities),
			"categories":    len(s.categories),
			"plugins":       len(s.plugins),
			"tags":          len(s.tags),
			"tag_aliases":   len(s.tagAliases),
		},
	}
}

func communitySeedData() []domain.Community {
	return []domain.Community{
		{ID: 1, Name: "PHP", Slug: "php", Logo: "PHP", CoverImage: "/frontend-assets/community-php.jpg", Slogan: "框架生态、工程实践与线上问题复盘", Description: "PHP 技术社区，讨论 Laravel、Hyperf、Swoole、Composer、性能优化和工程实践。", ThemeColor: "#6478c8", SEOTitle: "PHP 技术社区", SEODescription: "DevHub PHP 技术社区，聚合 Laravel、Hyperf、Swoole、Composer、性能优化、招聘和问答内容。", SEOKeywords: "PHP,Laravel,Hyperf,Swoole,Composer", SortOrder: 1, Status: 1, AnnouncementTitle: "PHP 子站公告", AnnouncementContent: "欢迎分享 PHP 工程实践、框架生态和线上排障经验。", AnnouncementURL: "/c/php/topics/new/"},
		{ID: 2, Name: "Go", Slug: "go", Logo: "GO", CoverImage: "/frontend-assets/community-go.jpg", Slogan: "并发、服务端与云原生实践", Description: "Go 技术社区，关注 Gin、Gorm、gRPC、微服务、并发模型和云原生工程。", ThemeColor: "#0891b2", SEOTitle: "Go 技术社区", SEODescription: "DevHub Go 技术社区，聚合 Gin、Gorm、gRPC、微服务、并发、Docker 和性能优化内容。", SEOKeywords: "Go,Gin,Gorm,gRPC,微服务,Docker", SortOrder: 2, Status: 1, AnnouncementTitle: "Go 子站公告", AnnouncementContent: "欢迎发布 Go 项目结构、并发治理和云原生实践。", AnnouncementURL: "/c/go/topics/new/"},
		{ID: 3, Name: "Java", Slug: "java", Logo: "JAVA", CoverImage: "/frontend-assets/community-java.jpg", Slogan: "后端架构、中间件与企业工程实践", Description: "Java 技术社区，覆盖 Spring Boot、MyBatis、JVM、消息队列、微服务和性能调优。", ThemeColor: "#ea580c", SEOTitle: "Java 技术社区", SEODescription: "DevHub Java 技术社区，聚合 Spring Boot、MyBatis、JVM、微服务、消息队列和企业开发内容。", SEOKeywords: "Java,Spring Boot,MyBatis,JVM,微服务", SortOrder: 3, Status: 1, AnnouncementTitle: "Java 子站公告", AnnouncementContent: "欢迎沉淀 Java 后端架构、中间件和 JVM 调优经验。", AnnouncementURL: "/c/java/topics/new/"},
		{ID: 4, Name: "AI", Slug: "ai", Logo: "AI", CoverImage: "/frontend-assets/community-ai.jpg", Slogan: "Agent、RAG、Prompt 与工作流实践", Description: "AI 技术社区，讨论 AI Agent、RAG、Prompt、OpenAI、Claude、Codex 和工作流工程。", ThemeColor: "#7c3aed", SEOTitle: "AI 技术社区", SEODescription: "DevHub AI 技术社区，聚合 AI Agent、RAG、Prompt、OpenAI、Claude、Codex 和工作流内容。", SEOKeywords: "AI Agent,RAG,Prompt,OpenAI,Claude,Codex", SortOrder: 4, Status: 1, AnnouncementTitle: "AI 子站公告", AnnouncementContent: "欢迎分享 AI 应用开发、Agent 工作流和 RAG 落地经验。", AnnouncementURL: "/c/ai/topics/new/"},
		{ID: 5, Name: "Frontend", Slug: "frontend", Logo: "FE", CoverImage: "/frontend-assets/community-frontend.jpg", Slogan: "Vue、React、TypeScript 与前端工程化", Description: "Frontend 技术社区，聚合 Vue、React、TypeScript、Vite、性能优化和前端工程化内容。", ThemeColor: "#16a34a", SEOTitle: "Frontend 技术社区", SEODescription: "DevHub Frontend 技术社区，聚合 Vue、React、TypeScript、Vite、性能优化和前端工程化内容。", SEOKeywords: "Frontend,Vue,React,TypeScript,Vite", SortOrder: 5, Status: 1, AnnouncementTitle: "前端子站公告", AnnouncementContent: "欢迎分享前端工程化、组件设计和性能优化实践。", AnnouncementURL: "/c/frontend/topics/new/"},
	}
}

func defaultCategorySeeds(communityID int64) []domain.Category {
	defs := []struct {
		name string
		slug string
		ct   string
		icon string
		desc string
	}{
		{"社区", "community", "article", "Message", "技术讨论、经验复盘和社区动态。"},
		{"问答中心", "qa", "question", "QuestionFilled", "可被采纳最佳答案的问题与解答。"},
		{"开源项目", "opensource", "project", "FolderOpened", "项目发布、库推荐和开源协作。"},
		{"AI作品", "ai", "ai_work", "MagicStick", "AI 应用、Prompt、Agent 和工作流作品。"},
		{"招聘内推", "jobs", "job", "Briefcase", "招聘、内推和团队介绍。"},
		{"Wiki", "wiki", "wiki", "Document", "长期沉淀的知识条目。"},
		{"文档", "docs", "doc", "Notebook", "规范、教程和操作文档。"},
	}
	out := make([]domain.Category, 0, len(defs))
	for i, def := range defs {
		id := communityID*100 + int64(i) + 1
		contentType := pluginregistry.NormalizeContentType(def.ct)
		out = append(out, domain.Category{
			ID:                  id,
			CommunityID:         communityID,
			Name:                def.name,
			Slug:                def.slug,
			Type:                contentType,
			ContentType:         contentType,
			PluginCode:          pluginregistry.PluginCodeForContentType(contentType),
			AllowedContentTypes: pluginregistry.DefaultAllowedContentTypes(contentType),
			Description:         def.desc,
			Icon:                def.icon,
			SortOrder:           i,
			Visible:             true,
			NavVisible:          true,
			Postable:            true,
			Status:              1,
			CreatedAt:           Now(),
			UpdatedAt:           Now(),
		})
	}
	return out
}

func (s *MemoryStore) seedPluginsLocked() {
	now := Now()
	for _, def := range pluginregistry.Definitions() {
		cp := def
		cp.CreatedAt = now
		cp.UpdatedAt = now
		s.plugins[cp.Code] = &cp
	}
}

func (s *MemoryStore) seedCommunityPluginsLocked() {
	now := Now()
	for id := range s.communities {
		if _, ok := s.communityPlugins[id]; !ok {
			s.communityPlugins[id] = map[string]*domain.CommunityPlugin{}
		}
		for _, def := range pluginregistry.Definitions() {
			s.communityPlugins[id][def.Code] = &domain.CommunityPlugin{
				ID:          0,
				CommunityID: id,
				PluginCode:  def.Code,
				Status:      pluginregistry.StatusEnabled,
				SortOrder:   0,
				ConfigJSON:  "",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
		}
	}
}

func (s *MemoryStore) ensureDocsSpaceLocked(communityID int64) int64 {
	for _, space := range s.docsSpaces {
		if space != nil && space.CommunityID == communityID && space.Slug == "default" {
			return space.ID
		}
	}
	id := int64(len(s.docsSpaces) + 1)
	s.docsSpaces[id] = &domain.DocsSpace{
		ID:          id,
		CommunityID: communityID,
		Name:        "默认文档空间",
		Slug:        "default",
		Description: "由 Docs 插件自动创建的默认空间。",
		Status:      1,
		CreatedAt:   Now(),
		UpdatedAt:   Now(),
	}
	return id
}

func (s *MemoryStore) ensureWikiPageLocked(topic domain.Topic) {
	if _, ok := s.wikiPages[topic.ID]; ok {
		return
	}
	spaceID := int64(topic.CommunityID)
	id := topic.ID
	versionID := int64(len(s.wikiVersions) + 1)
	s.wikiPages[id] = &domain.WikiPage{
		ID:               id,
		SpaceID:          spaceID,
		TopicID:          topic.ID,
		CommunityID:      topic.CommunityID,
		CategoryID:       topic.CategoryID,
		UserID:           topic.UserID,
		Title:            topic.Title,
		Summary:          topic.Summary,
		Content:          topic.Content,
		Status:           1,
		CurrentVersionID: versionID,
		CreatedAt:        topic.CreatedAt,
		UpdatedAt:        topic.UpdatedAt,
	}
	s.wikiVersions[versionID] = &domain.WikiRevision{
		ID:         versionID,
		WikiPageID: id,
		TopicID:    topic.ID,
		EditorID:   topic.UserID,
		VersionNo:  1,
		Title:      topic.Title,
		Content:    topic.Content,
		ChangeNote: "初始版本",
		CreatedAt:  topic.CreatedAt,
	}
}

func (s *MemoryStore) seedPluginContentRowsLocked(topic domain.Topic) {
	switch topic.PluginCode {
	case "qa":
		s.qaQuestions[topic.ID] = &domain.QAQuestion{
			ID:          topic.ID,
			TopicID:     topic.ID,
			AnswerCount: 0,
			IsResolved:  topic.IsSolved,
			CreatedAt:   topic.CreatedAt,
			UpdatedAt:   topic.UpdatedAt,
		}
	case "docs":
		spaceID := s.ensureDocsSpaceLocked(topic.CommunityID)
		s.docsDocuments[topic.ID] = &domain.DocsDocument{
			ID:         topic.ID,
			SpaceID:    spaceID,
			TopicID:    topic.ID,
			ParentID:   0,
			SortOrder:  0,
			Status:     1,
			Version:    1,
			EditorType: "markdown",
			CreatedAt:  topic.CreatedAt,
			UpdatedAt:  topic.UpdatedAt,
		}
	case "wiki":
		s.ensureWikiPageLocked(topic)
	}
}

func (s *MemoryStore) countWikiVersionsLocked(wikiPageID int64) int {
	count := 0
	for _, item := range s.wikiVersions {
		if item != nil && item.WikiPageID == wikiPageID {
			count++
		}
	}
	return count
}

func (s *MemoryStore) seedCommunitiesAndCategories() {
	for _, comm := range communitySeedData() {
		c := comm
		now := Now()
		c.CreatedAt = now
		c.UpdatedAt = now
		s.communities[c.ID] = &c
		if c.ID >= s.nextCommunityID {
			s.nextCommunityID = c.ID + 1
		}
		for _, cat := range defaultCategorySeeds(c.ID) {
			cp := cat
			s.categories[cp.ID] = &cp
			if cp.ID >= s.nextCategoryID {
				s.nextCategoryID = cp.ID + 1
			}
		}
	}
}

// seed 初始化站点、板块、帖子、评论、通知和后台管理演示数据。
func (s *MemoryStore) seed() {
	s.seedPluginsLocked()

	s.sites["portal"] = domain.Site{Key: "portal", Name: "DevHub", Logo: "DH", Title: "DevHub", Sub: "总网站 · 多技术子站内容集合", Pub: "发布内容", Description: "聚合 PHP、Go、Java、AI、Frontend 内容", Color: "#2563eb", Status: "enable", Sort: 0}
	s.sites["php"] = domain.Site{Key: "php", Name: "PHP", Logo: "PHP", Title: "PHP 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 PHP 内容", Description: "PHP 技术社区", Color: "#7c3aed", Status: "enable", Sort: 1}
	s.sites["go"] = domain.Site{Key: "go", Name: "Go", Logo: "GO", Title: "Go 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 Go 内容", Description: "Go 技术社区", Color: "#06b6d4", Status: "enable", Sort: 2}
	s.sites["java"] = domain.Site{Key: "java", Name: "Java", Logo: "JAVA", Title: "Java 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 Java 内容", Description: "Java 技术社区", Color: "#f97316", Status: "enable", Sort: 3}
	s.sites["ai"] = domain.Site{Key: "ai", Name: "AI", Logo: "AI", Title: "AI 子网站", Sub: "子网站 · 7 个板块", Pub: "发布 AI 内容", Description: "AI Agent、RAG、Prompt 与工作流社区", Color: "#7c3aed", Status: "enable", Sort: 4}
	s.sites["frontend"] = domain.Site{Key: "frontend", Name: "Frontend", Logo: "FE", Title: "Frontend 子网站", Sub: "子网站 · 7 个板块", Pub: "发布前端内容", Description: "Vue、React、TypeScript 与前端工程化社区", Color: "#16a34a", Status: "enable", Sort: 5}
	s.seedCommunitiesAndCategories()
	s.seedCommunityPluginsLocked()

	boardNames := map[string]string{"all": "全部", "community": "社区", "qa": "问答中心", "opensource": "开源项目", "ai": "AI作品", "jobs": "招聘内推", "wiki": "Wiki", "docs": "文档"}
	for i, key := range s.boardOrder {
		s.boards[key] = domain.Board{Key: key, Name: boardNames[key], Site: "all", Sort: i, Visible: true}
	}

	s.roles[1] = domain.AdminRole{ID: 1, Name: "超级管理员", Builtin: true, Description: "拥有所有模块操作权限", Permissions: []string{"*"}, UserCount: 1}
	s.roles[2] = domain.AdminRole{ID: 2, Name: "站点管理员", Builtin: true, Description: "负责授权子站的内容和举报治理", Permissions: []string{"dashboard.read", "site.read", "site.write", "board.read", "board.write", "post.read", "post.create", "post.update", "post.delete", "topic.moderate", "comment.read", "comment.moderate", "report.read", "report.handle", "moderator.read", "notification.write", "log.read", "plugin.read", "qa.question.create", "qa.question.audit", "docs.document.create", "docs.document.audit", "docs.space.manage", "wiki.page.create", "wiki.page.audit", "wiki.page.version.rollback", "projects.project.create", "projects.project.audit", "jobs.job.create", "jobs.job.audit", "ai_works.work.create", "ai_works.work.audit"}, UserCount: 1}
	s.roles[3] = domain.AdminRole{ID: 3, Name: "内容审核员", Builtin: true, Description: "负责授权子站的内容审核和评论治理", Permissions: []string{"dashboard.read", "post.read", "post.update", "topic.moderate", "comment.read", "comment.moderate", "report.read", "report.handle", "plugin.read", "qa.question.audit", "docs.document.audit", "wiki.page.audit"}, UserCount: 1}
	defaultPassword, _ := hashPassword("admin123")
	s.users[1] = &domain.AdminUser{ID: 1, Username: "admin", Nickname: "超级管理员", Avatar: "", Phone: "13800000001", Email: "admin@devhub.local", PasswordHash: defaultPassword, Status: "normal", RoleID: 1, RoleName: "超级管理员", CreatedAt: "2026-04-01 09:00:00", LastLoginAt: "2026-05-06 09:30:00"}
	s.users[2] = &domain.AdminUser{ID: 2, Username: "operator", Nickname: "运营管理员", Avatar: "", Phone: "13800000002", Email: "operator@devhub.local", PasswordHash: defaultPassword, Status: "normal", RoleID: 2, RoleName: "运营管理员", CreatedAt: "2026-04-08 09:00:00", LastLoginAt: "2026-05-05 18:20:00"}
	s.users[3] = &domain.AdminUser{ID: 3, Username: "auditor", Nickname: "内容审核员", Avatar: "", Phone: "13800000003", Email: "auditor@devhub.local", PasswordHash: defaultPassword, Status: "normal", RoleID: 3, RoleName: "内容审核员", CreatedAt: "2026-04-12 09:00:00", LastLoginAt: "2026-05-06 10:12:00"}
	if frontPassword, err := hashPassword("a123456"); err == nil {
		s.users[4] = &domain.AdminUser{ID: 4, Username: "liuwei", Nickname: "方圆十三", Avatar: "", Phone: "13800000004", Email: "liuwei@devhub.local", PasswordHash: frontPassword, Status: "normal", RoleName: "普通用户", CreatedAt: "2026-05-12 09:00:00", LastLoginAt: "2026-05-12 09:00:00"}
	}
	for _, u := range s.users {
		if u.ID >= s.nextUserID {
			s.nextUserID = u.ID + 1
		}
	}
	for _, moderator := range []domain.CommunityModerator{
		{CommunityID: 1, UserID: 2, Role: "moderator", Status: 1},
		{CommunityID: 2, UserID: 3, Role: "moderator", Status: 1},
	} {
		moderator.ID = s.nextModeratorID
		s.nextModeratorID++
		moderator.CreatedAt = Now()
		moderator.UpdatedAt = moderator.CreatedAt
		cp := moderator
		s.moderators[cp.ID] = &cp
	}

	seedPosts := []domain.Post{
		{Site: "php", Board: "community", Title: "Laravel 社区系统如何设计积分和通知？", Summary: "从用户行为、积分流水、通知触发器、异步队列几个角度拆解社区积分系统。", Author: "LaravelChen", Views: 2380, Likes: 128, Comments: 0, CreatedAt: "2026-05-01 09:00:00", Tags: []string{"Laravel", "积分", "通知"}, Content: "一个社区系统的积分与通知不应该散落在业务代码中，而应该通过事件、积分规则表、通知模板和队列消费组合起来。"},
		{Site: "php", Board: "qa", Title: "PHP-FPM 502 问题应该如何排查？", Summary: "从 Nginx、PHP-FPM、慢日志、进程池、超时配置和内存占用逐层定位。", Author: "SwooleDev", Views: 1890, Likes: 84, Comments: 0, CreatedAt: "2026-05-02 10:00:00", Tags: []string{"PHP-FPM", "Nginx", "502"}, Content: "502 通常不是一个单点问题。建议先确认 Nginx upstream 是否能连通 PHP-FPM，再检查 PHP-FPM 进程池是否耗尽。"},
		{Site: "php", Board: "opensource", Title: "PHP Package Starter：Composer 包模板", Summary: "用于快速创建 Composer 包、测试、CI 和发布流程的模板。", Author: "OpenSourceHub", Views: 860, Likes: 38, Comments: 0, CreatedAt: "2026-05-02 11:00:00", Tags: []string{"Composer", "代码规范", "开源项目"}, Content: "PHP Package Starter 提供 PSR-4、PHPUnit、静态分析和 GitHub Actions 发布流程。"},
		{Site: "php", Board: "jobs", Title: "招聘：PHP 后端工程师", Summary: "负责社区系统、支付系统和内部管理平台开发。", Author: "Hiring Desk", Views: 760, Likes: 26, Comments: 0, CreatedAt: "2026-05-02 12:00:00", Tags: []string{"招聘内推", "Laravel", "MySQL"}, Content: "岗位要求熟悉 Laravel 或 Hyperf，有 Redis、MySQL 和队列系统经验。"},
		{Site: "php", Board: "wiki", Title: "PHP Wiki：Laravel 请求生命周期", Summary: "从 public/index.php 到 Kernel、Middleware、Router、Controller 的完整链路。", Author: "DevHub Wiki", Views: 980, Likes: 45, Comments: 0, CreatedAt: "2026-04-29 12:00:00", Tags: []string{"Wiki", "Laravel", "生命周期"}, Content: "Laravel 请求从 public/index.php 进入，随后创建应用容器，加载 HTTP Kernel，经过中间件管道，进入路由匹配。"},
		{Site: "php", Board: "docs", Title: "PHP 文档：Composer 自动加载", Summary: "解释 PSR-4、autoload、classmap、files 自动加载的适用场景。", Author: "ComposerBot", Views: 1320, Likes: 62, Comments: 0, CreatedAt: "2026-04-28 12:00:00", Tags: []string{"Composer", "PSR-4"}, Content: "Composer 自动加载最常见的是 PSR-4。它通过命名空间前缀映射到目录路径。"},
		{Site: "go", Board: "community", Title: "Go 项目应该如何组织目录结构？", Summary: "讨论 cmd、internal、pkg、api、configs、scripts 等目录的边界。", Author: "GopherLin", Views: 2100, Likes: 101, Comments: 0, CreatedAt: "2026-05-01 12:00:00", Tags: []string{"Go", "项目结构"}, Content: "Go 项目结构的重点不是目录越多越好，而是边界清晰。cmd 放入口，internal 放内部业务。"},
		{Site: "go", Board: "qa", Title: "Goroutine 泄漏如何定位？", Summary: "从 pprof、context、channel 阻塞、select default 等角度排查。", Author: "GoTrace", Views: 1760, Likes: 92, Comments: 0, CreatedAt: "2026-05-03 08:30:00", Tags: []string{"Goroutine", "pprof", "泄漏"}, Content: "Goroutine 泄漏常见原因是 channel 永久阻塞、context 未取消、后台循环没有退出条件。"},
		{Site: "go", Board: "opensource", Title: "Go CLI Starter：命令行工具模板", Summary: "一个用于快速启动 Cobra/Viper 命令行项目的模板。", Author: "OpenSourceHub", Views: 860, Likes: 39, Comments: 0, CreatedAt: "2026-04-27 12:00:00", Tags: []string{"CLI", "Cobra", "开源项目"}, Content: "Go CLI Starter 集成 Cobra 命令组织、Viper 配置读取、日志初始化和版本输出。"},
		{Site: "go", Board: "jobs", Title: "招聘：Go 云原生工程师", Summary: "负责微服务、网关、任务调度和容器平台建设。", Author: "Hiring Desk", Views: 940, Likes: 34, Comments: 0, CreatedAt: "2026-04-27 13:00:00", Tags: []string{"招聘内推", "微服务", "Docker"}, Content: "岗位要求熟悉 Go、Docker、gRPC 和高并发服务治理。"},
		{Site: "go", Board: "wiki", Title: "Go Wiki：并发模式速览", Summary: "整理 worker pool、fan-in/fan-out、context 取消和限流模式。", Author: "DevHub Wiki", Views: 790, Likes: 41, Comments: 0, CreatedAt: "2026-04-27 14:00:00", Tags: []string{"并发", "Wiki", "gRPC"}, Content: "Go 并发设计常用模式包括 worker pool、fan-in/fan-out、context 取消、超时和限流。"},
		{Site: "go", Board: "docs", Title: "Go 文档：context 超时控制", Summary: "说明 WithCancel、WithTimeout、WithDeadline 的区别和常见误区。", Author: "DevHub Docs", Views: 1120, Likes: 58, Comments: 0, CreatedAt: "2026-04-26 12:00:00", Tags: []string{"context", "超时控制"}, Content: "context 用于跨 API 边界传递取消信号、超时和请求级元数据。"},
		{Site: "java", Board: "community", Title: "Spring Boot 项目如何分层更清晰？", Summary: "Controller、Application Service、Domain、Infrastructure 的职责拆分。", Author: "SpringLee", Views: 2320, Likes: 116, Comments: 0, CreatedAt: "2026-05-02 13:00:00", Tags: []string{"Spring Boot", "分层"}, Content: "Spring Boot 项目常见问题是 Controller 太厚、Service 太杂。建议拆分接口层、应用服务、领域模型、基础设施。"},
		{Site: "java", Board: "qa", Title: "JVM Full GC 频繁如何排查？", Summary: "结合 GC 日志、堆转储、对象增长趋势和内存区域定位问题。", Author: "JvmDoctor", Views: 2600, Likes: 149, Comments: 0, CreatedAt: "2026-05-03 09:00:00", Tags: []string{"JVM", "Full GC", "性能"}, Content: "Full GC 频繁需要先看 GC 日志确认触发原因，再通过 heap dump 分析大对象和引用链。"},
		{Site: "java", Board: "opensource", Title: "Spring Boot Admin Starter：后台脚手架", Summary: "包含 RBAC、审计日志、配置中心和内容管理基础模块。", Author: "OpenSourceHub", Views: 1180, Likes: 52, Comments: 0, CreatedAt: "2026-05-03 10:00:00", Tags: []string{"Spring Boot", "MyBatis", "开源项目"}, Content: "该脚手架适合中后台管理系统，内置权限、菜单、审计和代码生成基础能力。"},
		{Site: "java", Board: "ai", Title: "Spring AI 接入大模型的项目结构建议", Summary: "把模型配置、提示词模板、工具调用和业务流程拆开。", Author: "AIJava", Views: 1480, Likes: 73, Comments: 0, CreatedAt: "2026-04-30 12:00:00", Tags: []string{"Spring AI", "LLM"}, Content: "Spring AI 接入大模型时，不建议把 Prompt 和业务逻辑全部写在 Controller。"},
		{Site: "java", Board: "jobs", Title: "招聘：Java 平台工程师", Summary: "负责中台服务、消息队列、缓存治理和 JVM 性能优化。", Author: "Hiring Desk", Views: 980, Likes: 36, Comments: 0, CreatedAt: "2026-04-30 13:00:00", Tags: []string{"招聘内推", "JVM", "消息队列"}, Content: "岗位要求熟悉 Spring Boot、MyBatis、Redis、消息队列和 JVM 调优。"},
		{Site: "java", Board: "wiki", Title: "Java Wiki：JVM 内存区域速览", Summary: "堆、方法区、虚拟机栈、本地方法栈、程序计数器的基础说明。", Author: "DevHub Wiki", Views: 990, Likes: 54, Comments: 0, CreatedAt: "2026-04-25 12:00:00", Tags: []string{"JVM", "Wiki"}, Content: "JVM 运行时内存区域包括线程共享的堆和方法区，以及线程私有的虚拟机栈、本地方法栈、程序计数器。"},
		{Site: "java", Board: "docs", Title: "Java 文档：MyBatis 映射规范", Summary: "记录 Mapper、XML、分页、批量写入和事务边界约定。", Author: "DevHub Docs", Views: 870, Likes: 33, Comments: 0, CreatedAt: "2026-04-25 13:00:00", Tags: []string{"MyBatis", "代码规范"}, Content: "MyBatis 使用中需要明确 Mapper 职责、SQL 命名、分页约定、批量写入和事务边界。"},
		{Site: "ai", Board: "community", Title: "AI Agent 工作流如何设计得可维护？", Summary: "从工具调用、上下文压缩、任务状态和人工确认四个角度拆解 Agent 工程化。", Author: "AgentNotes", Views: 1560, Likes: 88, Comments: 0, CreatedAt: "2026-05-04 09:00:00", Tags: []string{"AI Agent", "工作流", "Codex"}, Content: "AI Agent 的重点不是把模型接进来，而是把任务边界、工具权限、状态恢复和人工确认做清楚。"},
		{Site: "ai", Board: "qa", Title: "如何优化 Prompt 获得更稳定的 AI 回复？", Summary: "从 Prompt 结构、上下文、示例、约束和输出格式讨论优化方法。", Author: "PromptLab", Views: 1340, Likes: 65, Comments: 0, CreatedAt: "2026-05-04 10:00:00", Tags: []string{"Prompt", "OpenAI", "Claude"}, Content: "好的 Prompt 应该结构清晰、上下文完整、提供示例，并明确要求模型在信息不足时说明不确定性。"},
		{Site: "ai", Board: "opensource", Title: "RAG Starter：知识库问答项目模板", Summary: "一个用于验证文档切分、向量检索和答案引用的 RAG 项目模板。", Author: "RAGHub", Views: 1180, Likes: 56, Comments: 0, CreatedAt: "2026-05-04 11:00:00", Tags: []string{"RAG", "OpenAI", "工作流"}, Content: "RAG Starter 包含文档导入、切分、Embedding、检索、重排和答案引用展示。"},
		{Site: "ai", Board: "jobs", Title: "招聘：AI 应用开发工程师", Summary: "负责 AI 应用开发、Agent 工作流和内部效率工具建设。", Author: "Hiring Desk", Views: 920, Likes: 45, Comments: 0, CreatedAt: "2026-05-04 12:00:00", Tags: []string{"招聘内推", "AI Agent"}, Content: "岗位要求熟悉主流大模型 API，有 RAG 或 Agent 应用开发经验优先。"},
		{Site: "ai", Board: "wiki", Title: "AI Wiki：RAG 基础概念", Summary: "整理向量、Embedding、召回、重排和引用生成的基础知识。", Author: "DevHub Wiki", Views: 760, Likes: 35, Comments: 0, CreatedAt: "2026-05-04 13:00:00", Tags: []string{"RAG", "Wiki"}, Content: "RAG 通过检索外部知识补充模型上下文，核心流程包括切分、向量化、召回、重排和生成。"},
		{Site: "ai", Board: "docs", Title: "AI 文档：OpenAI API 调用约定", Summary: "记录模型调用、错误处理、重试和日志脱敏的基础规范。", Author: "DevHub Docs", Views: 830, Likes: 42, Comments: 0, CreatedAt: "2026-05-04 14:00:00", Tags: []string{"OpenAI", "代码规范"}, Content: "调用大模型 API 时应统一封装超时、重试、错误分类、审计日志和敏感数据脱敏。"},
		{Site: "frontend", Board: "community", Title: "Vue 3 Composition API 最佳实践", Summary: "分享 Vue 3 组合式 API 的使用技巧、状态拆分和类型推导经验。", Author: "FrontendLab", Views: 1890, Likes: 95, Comments: 0, CreatedAt: "2026-05-05 09:00:00", Tags: []string{"Vue", "TypeScript", "代码规范"}, Content: "Composition API 适合把复杂逻辑拆成可复用的 composable，同时保持组件模板足够清晰。"},
		{Site: "frontend", Board: "qa", Title: "React Hooks 常见误区有哪些？", Summary: "讨论依赖项、闭包、useMemo 滥用和异步副作用的常见问题。", Author: "ReactNotes", Views: 1680, Likes: 78, Comments: 0, CreatedAt: "2026-05-05 10:00:00", Tags: []string{"React", "TypeScript", "性能优化"}, Content: "React Hooks 需要注意依赖项的正确设置和闭包陷阱，不要把 useMemo 当作默认性能优化方案。"},
		{Site: "frontend", Board: "opensource", Title: "Vite + React TypeScript 模板项目", Summary: "一个包含路由、状态管理、请求封装和代码规范的前端模板。", Author: "OpenSourceHub", Views: 1240, Likes: 62, Comments: 0, CreatedAt: "2026-05-05 11:00:00", Tags: []string{"Vite", "React", "TypeScript"}, Content: "模板包含 ESLint、Prettier、路由、状态管理、请求封装和基础布局。"},
		{Site: "frontend", Board: "jobs", Title: "招聘：高级前端工程师", Summary: "负责中后台产品、组件库和性能优化。", Author: "Hiring Desk", Views: 860, Likes: 31, Comments: 0, CreatedAt: "2026-05-05 12:00:00", Tags: []string{"招聘内推", "Vue", "React"}, Content: "岗位要求熟悉 Vue 或 React，具备 TypeScript、工程化和性能优化经验。"},
		{Site: "frontend", Board: "wiki", Title: "Frontend Wiki：TypeScript 类型收窄", Summary: "整理 typeof、in、谓词函数和可辨识联合类型。", Author: "DevHub Wiki", Views: 720, Likes: 29, Comments: 0, CreatedAt: "2026-05-05 13:00:00", Tags: []string{"TypeScript", "Wiki"}, Content: "TypeScript 类型收窄可以通过 typeof、in、instanceof、自定义谓词和可辨识联合类型实现。"},
		{Site: "frontend", Board: "docs", Title: "前端文档：Vite 构建性能优化", Summary: "说明依赖预构建、代码分割、缓存和构建分析。", Author: "DevHub Docs", Views: 900, Likes: 44, Comments: 0, CreatedAt: "2026-05-05 14:00:00", Tags: []string{"Vite", "性能优化"}, Content: "Vite 构建优化可以从依赖预构建、动态导入、manualChunks 和缓存策略入手。"},
	}
	for _, p := range seedPosts {
		p.UpdatedAt = p.CreatedAt
		s.createPostLocked(&p)
	}
	s.rebuildTagsFromPostsLocked()

	s.createCommentLocked(1, 0, "Corwien", "", "不错，图文并茂！", "2026-05-03 10:12:00", 12)
	s.createCommentLocked(1, 1, "LaravelChen", "Corwien", "thank you", "2026-05-03 10:30:00", 5)
	s.createCommentLocked(1, 2, "DevHubUser", "LaravelChen", "这个回复层级展示很清楚，后续可以加折叠。", "2026-05-03 11:00:00", 2)
	s.createCommentLocked(1, 0, "Ysll", "", "不错不错，学习学习。", "2026-05-03 12:15:00", 8)
	s.createCommentLocked(7, 0, "GoTrace", "", "Go 目录结构重点还是边界清晰。", "2026-05-04 09:20:00", 7)
	s.createCommentLocked(13, 0, "SpringLee", "", "分层之后测试和维护都会轻松很多。", "2026-05-04 10:20:00", 6)
	s.createCommentLocked(20, 0, "AgentNotes", "", "工具权限和状态恢复确实是 Agent 落地的关键。", "2026-05-05 09:20:00", 9)
	s.createCommentLocked(26, 0, "FrontendLab", "", "组合式 API 拆分逻辑后组件会清爽很多。", "2026-05-06 09:20:00", 5)

	s.createNoticeLocked("LaravelChen 回复了你的评论", "thank you")
	s.createNoticeLocked("你的帖子获得了新的点赞", "JVM Full GC 频繁如何排查？")
	s.createNoticeLocked("Go 子网站有新的问答", "Goroutine 泄漏如何定位？")
	s.appendLogLocked("login", "admin", "管理员登录", "后台系统", "127.0.0.1")
	s.appendLogLocked("operation", "operator", "配置首页推荐", "posts#10", "127.0.0.1")
}

// createPostLocked 在调用方已持有写锁时创建帖子。
func (s *MemoryStore) createPostLocked(p *domain.Post) *domain.Post {
	p.ID = s.nextPostID
	s.nextPostID++
	if p.UserID <= 0 {
		p.UserID = 1
	}
	if p.Author == "" {
		p.Author = "SUI.CHEN"
	}
	if p.CreatedAt == "" {
		p.CreatedAt = Now()
	}
	if p.UpdatedAt == "" {
		p.UpdatedAt = p.CreatedAt
	}
	if p.Status == "" {
		p.Status = "publish"
	}
	cp := *p
	cp.Tags = uniqueTags(cp.Tags)
	s.posts[cp.ID] = &cp
	return &cp
}

func (s *MemoryStore) rebuildTagsFromPostsLocked() {
	existing := map[string]domain.Tag{}
	for _, tag := range s.tags {
		if tag == nil {
			continue
		}
		key := tagKey(tag.Site, tag.Slug)
		if key != ":" {
			cp := *tag
			cp.UseCount = 0
			cp.TopicCount = 0
			existing[key] = cp
		}
	}
	s.tags = map[int64]*domain.Tag{}
	s.nextTagID = 1
	for _, tag := range existing {
		if tag.ID <= 0 {
			tag.ID = s.nextTagID
		}
		if tag.ID >= s.nextTagID {
			s.nextTagID = tag.ID + 1
		}
		cp := tag
		s.tags[cp.ID] = &cp
	}
	for _, post := range s.posts {
		if post == nil {
			continue
		}
		for _, name := range uniqueTags(post.Tags) {
			tag := normalizeTag(domain.Tag{Site: post.Site, Name: name, Status: "enable"})
			if tag.Slug == "" {
				continue
			}
			key := tagKey(tag.Site, tag.Slug)
			if current, ok := s.findTagBySiteSlugLocked(tag.Site, tag.Slug); ok {
				if current.Name == "" {
					current.Name = tag.Name
				}
				current.UseCount++
				current.TopicCount = current.UseCount
				current.UpdatedAt = Now()
				continue
			}
			if old, ok := existing[key]; ok {
				tag = old
				tag.UseCount = 0
				tag.TopicCount = 0
			}
			if tag.CreatedAt == "" {
				tag.CreatedAt = Now()
			}
			tag.UpdatedAt = Now()
			tag.UseCount = 1
			tag.TopicCount = 1
			tag.ID = s.nextTagID
			s.nextTagID++
			cp := tag
			s.tags[cp.ID] = &cp
		}
	}
}

func memoryHotScore(p *domain.Post) int {
	if p == nil {
		return 0
	}
	return p.Views + p.Comments*5 + p.Likes*3
}

func memoryHotScoreWithFavorites(p *domain.Post, favoriteCount int) int {
	if p == nil {
		return 0
	}
	return p.Views + p.Comments*5 + p.Likes*3 + favoriteCount*4
}

func memoryPostVisible(p *domain.Post) bool {
	if p == nil {
		return false
	}
	switch strings.TrimSpace(p.Status) {
	case "", "publish", "published", "normal":
		return true
	default:
		return false
	}
}

func memoryTopicStatus(p *domain.Post) int {
	if memoryPostVisible(p) {
		return 1
	}
	if p != nil && strings.TrimSpace(p.Status) == "deleted" {
		return 3
	}
	return 0
}

func memorySetPostStatus(p *domain.Post, status int) {
	if p == nil {
		return
	}
	if status == 1 {
		p.Status = "publish"
	} else if status == 3 {
		p.Status = "deleted"
	} else {
		p.Status = "hidden"
	}
}

func (s *MemoryStore) hotScoreLocked(p *domain.Post) int {
	if p == nil {
		return 0
	}
	return memoryHotScoreWithFavorites(p, s.favoriteCountLocked("topic", p.ID))
}

// createCommentLocked 在调用方已持有写锁时创建评论并更新帖子评论数。
func (s *MemoryStore) createCommentLocked(postID, parentID int64, author, to, text, createdAt string, likes int) *domain.Comment {
	c := &domain.Comment{ID: s.nextCommentID, PostID: postID, ParentID: parentID, Author: author, To: to, Text: text, Status: "normal", Likes: likes, CreatedAt: createdAt}
	s.nextCommentID++
	if c.Author == "" {
		c.Author = "SUI.CHEN"
	}
	if c.CreatedAt == "" {
		c.CreatedAt = Now()
	}
	c.TopicID = postID
	c.UserID = 1
	c.UserName = c.Author
	c.Content = c.Text
	c.LikeCount = c.Likes
	c.UpdatedAt = c.CreatedAt
	s.comments[c.ID] = c
	if p, ok := s.posts[postID]; ok {
		p.Comments++
		p.UpdatedAt = c.CreatedAt
	}
	return c
}

func (s *MemoryStore) appendActivityLocked(userID, communityID int64, action, targetType string, targetID, topicID int64, remark string) *domain.Activity {
	a := &domain.Activity{
		ID:          s.nextActivityID,
		UserID:      userID,
		CommunityID: communityID,
		TopicID:     topicID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Remark:      remark,
		CreatedAt:   Now(),
	}
	s.nextActivityID++
	s.enrichActivityLocked(a)
	s.activities[a.ID] = a
	return a
}

// createNoticeLocked 在调用方已持有写锁时创建站内通知。
func (s *MemoryStore) createNoticeLocked(title, content string) *domain.Notification {
	return s.createNoticeForSiteLocked("portal", title, content)
}

func (s *MemoryStore) createNoticeForSiteLocked(site, title, content string) *domain.Notification {
	n := &domain.Notification{ID: s.nextNoticeID, Site: normalizeSiteScope(site), UserID: 1, Title: title, Content: content, CreatedAt: Now()}
	n.IsRead = n.Read
	s.nextNoticeID++
	s.notices[n.ID] = n
	return n
}

func (s *MemoryStore) createUserNoticeLocked(userID, actorUserID int64, noticeType, targetType string, targetID, topicID, commentID int64, title, content string) *domain.Notification {
	if userID <= 0 || actorUserID == userID {
		return nil
	}
	n := &domain.Notification{
		ID:          s.nextNoticeID,
		Site:        "portal",
		UserID:      userID,
		ActorUserID: actorUserID,
		Type:        noticeType,
		TargetType:  targetType,
		TargetID:    targetID,
		TopicID:     topicID,
		CommentID:   commentID,
		Title:       strings.TrimSpace(title),
		Content:     strings.TrimSpace(content),
		TargetURL:   targetURLFor(targetType, targetID, topicID),
		CreatedAt:   Now(),
	}
	s.nextNoticeID++
	s.notices[n.ID] = n
	return n
}

// appendLogLocked 在调用方已持有写锁时追加后台操作日志。
func (s *MemoryStore) appendLogLocked(logType, actor, action, target, ip string) {
	s.appendLogForSiteLocked("portal", logType, actor, "", action, target, ip)
}

func (s *MemoryStore) appendLogForSiteLocked(site, logType, actor, role, action, target, ip string) {
	log := domain.AdminLog{ID: s.nextLogID, Site: normalizeSiteScope(site), Type: logType, Actor: actor, Role: role, Action: action, Target: target, IP: ip, CreatedAt: Now()}
	if actor == "system" || actor == "" {
		log.ActorType = "system"
	} else {
		log.ActorType = "admin_user"
	}
	s.logs = append(s.logs, log)
	s.nextLogID++
}

// ValidateSite 校验站点筛选值，portal 和空值表示全站。
func (s *MemoryStore) ValidateSite(site string) bool {
	if site == "portal" || site == "" {
		return true
	}
	_, ok := s.sites[site]
	return ok
}

// ValidateBoard 校验板块筛选值，all 和空值表示全部板块。
func (s *MemoryStore) ValidateBoard(board string) bool {
	if board == "all" || board == "" {
		return true
	}
	_, ok := s.boards[board]
	return ok
}

// AdminLogin 校验后台演示账号。当前阶段仍使用演示密码，但会校验后台人员存在和状态。
func (s *MemoryStore) AdminLogin(account, password string) (*domain.AdminSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	for _, u := range s.users {
		if u.Username != account && u.Email != account && u.Phone != account {
			continue
		}
		if u.RoleID <= 0 {
			return nil, errors.New("账号或密码错误")
		}
		if !memoryPasswordMatches(u.PasswordHash, password) {
			return nil, errors.New("账号或密码错误")
		}
		if u.Status == "forbidden" {
			return nil, errors.New("账号已被禁用")
		}
		u.LastLoginAt = Now()
		s.appendLogLocked("login", u.Username, "管理员登录", "后台系统", "127.0.0.1")
		token := fmt.Sprintf("devhub-admin-%d", u.ID)
		auth := s.memoryAuthUserLocked(u.ID, "admin")
		return &domain.AdminSession{
			Token:        token,
			AccessToken:  token,
			RefreshToken: token + "-refresh",
			ExpiresIn:    int64(accessTokenTTL.Seconds()),
			TokenType:    "admin",
			Audience:     "devhub_admin",
			User: domain.AdminLoginUser{
				ID:          auth.ID,
				Username:    auth.Username,
				Nickname:    auth.Nickname,
				Email:       auth.Email,
				Phone:       auth.Phone,
				Role:        auth.RoleName,
				RoleCode:    auth.RoleCode,
				Sites:       auth.Sites,
				Permissions: auth.Permissions,
			},
		}, nil
	}
	return nil, errors.New("账号或密码错误")
}

func (s *MemoryStore) UserLogin(account, password string) (*domain.AdminSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	for _, u := range s.users {
		if u.Username != account && u.Email != account && u.Phone != account {
			continue
		}
		if !memoryPasswordMatches(u.PasswordHash, password) {
			return nil, errors.New("账号或密码错误")
		}
		if u.Status == "forbidden" {
			return nil, errors.New("账号已被禁用")
		}
		u.LastLoginAt = Now()
		auth := s.memoryAuthUserLocked(u.ID, "user")
		return s.memoryUserSessionLocked(auth), nil
	}
	return nil, errors.New("账号或密码错误")
}

func (s *MemoryStore) RefreshSession(refreshToken string) (*domain.AdminSession, error) {
	id, err := parseMemoryTokenUserIDWithPrefix(refreshToken, "devhub-user-")
	if err != nil {
		return nil, errors.New("refresh token 无效")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	auth := s.memoryAuthUserLocked(id, "user")
	if auth.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	return s.memoryUserSessionLocked(auth), nil
}

func (s *MemoryStore) RefreshAdminSession(refreshToken string) (*domain.AdminSession, error) {
	id, err := parseMemoryTokenUserIDWithPrefix(refreshToken, "devhub-admin-")
	if err != nil {
		return nil, errors.New("refresh token 无效")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	auth := s.memoryAuthUserLocked(id, "admin")
	if auth.ID == 0 {
		return nil, errors.New("后台账号不存在")
	}
	return s.memoryAdminSessionLocked(auth), nil
}

func (s *MemoryStore) Register(req domain.RegisterRequest) (*domain.AdminSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码至少 6 位")
	}
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	for _, u := range s.users {
		if u.Username == username {
			return nil, errors.New("用户名已存在")
		}
		if email != "" && u.Email == email {
			return nil, errors.New("邮箱已存在")
		}
		if phone != "" && u.Phone == phone {
			return nil, errors.New("手机号已存在")
		}
	}
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = username
	}
	id := s.nextUserID
	s.nextUserID++
	s.users[id] = &domain.AdminUser{
		ID:           id,
		Username:     username,
		Nickname:     nickname,
		Email:        email,
		Phone:        phone,
		PasswordHash: hash,
		Status:       "normal",
		RoleName:     "普通用户",
		CreatedAt:    Now(),
	}
	auth := s.memoryAuthUserLocked(id, "user")
	return s.memoryUserSessionLocked(auth), nil
}

func memoryPasswordMatches(hash, password string) bool {
	if hash != "" {
		return checkPassword(hash, password)
	}
	// Backward compatibility for in-memory data created before password hashes existed.
	return password == "admin123" || password == "123456"
}

func (s *MemoryStore) Logout(refreshToken string) error { return nil }

func (s *MemoryStore) AuthUser(accessToken string) (*domain.AuthUser, error) {
	if !strings.HasPrefix(strings.TrimSpace(accessToken), "devhub-user-") {
		return nil, errors.New("token 无效")
	}
	id, err := parseMemoryTokenUserIDWithPrefix(accessToken, "devhub-user-")
	if err != nil {
		return nil, errors.New("token 无效")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user := s.memoryAuthUserLocked(id, "user")
	if user.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *MemoryStore) AuthAdmin(accessToken string) (*domain.AuthUser, error) {
	if !strings.HasPrefix(strings.TrimSpace(accessToken), "devhub-admin-") {
		return nil, errors.New("后台 token 无效")
	}
	id, err := parseMemoryTokenUserIDWithPrefix(accessToken, "devhub-admin-")
	if err != nil {
		return nil, errors.New("后台 token 无效")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user := s.memoryAuthUserLocked(id, "admin")
	if user.ID == 0 {
		return nil, errors.New("后台账号不存在")
	}
	return &user, nil
}

func parseMemoryTokenUserIDWithPrefix(token, prefix string) (int64, error) {
	token = strings.TrimPrefix(strings.TrimSpace(token), prefix)
	token = strings.TrimSuffix(token, "-refresh")
	return strconv.ParseInt(token, 10, 64)
}

func (s *MemoryStore) memoryAdminSessionLocked(auth domain.AuthUser) *domain.AdminSession {
	token := fmt.Sprintf("devhub-admin-%d", auth.ID)
	return &domain.AdminSession{
		Token:        token,
		AccessToken:  token,
		RefreshToken: token + "-refresh",
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		TokenType:    "admin",
		Audience:     "devhub_admin",
		User: domain.AdminLoginUser{
			ID:          auth.ID,
			Username:    auth.Username,
			Nickname:    auth.Nickname,
			Email:       auth.Email,
			Phone:       auth.Phone,
			Role:        auth.RoleName,
			RoleCode:    auth.RoleCode,
			Sites:       auth.Sites,
			Permissions: auth.Permissions,
		},
	}
}

func (s *MemoryStore) memoryUserSessionLocked(auth domain.AuthUser) *domain.AdminSession {
	token := fmt.Sprintf("devhub-user-%d", auth.ID)
	return &domain.AdminSession{
		Token:        token,
		AccessToken:  token,
		RefreshToken: token + "-refresh",
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		TokenType:    "user",
		Audience:     "devhub_frontend",
		User: domain.AdminLoginUser{
			ID:       auth.ID,
			Username: auth.Username,
			Nickname: auth.Nickname,
			Email:    auth.Email,
			Phone:    auth.Phone,
			Role:     auth.RoleName,
			RoleCode: auth.RoleCode,
			Sites:    auth.Sites,
		},
	}
}

func (s *MemoryStore) memoryAuthUserLocked(userID int64, identity string) domain.AuthUser {
	u, ok := s.users[userID]
	if !ok {
		return domain.AuthUser{}
	}
	roleCode := "user"
	roleName := "普通用户"
	sites := []string{}
	perms := []string{}
	if identity == "admin" {
		switch u.RoleID {
		case 1:
			roleCode = "super_admin"
			sites = []string{"*"}
		case 2:
			roleCode = "site_admin"
			sites = []string{"php"}
		case 3:
			roleCode = "moderator"
			sites = []string{"go"}
		}
		roleName = u.RoleName
		if role, ok := s.roles[u.RoleID]; ok {
			perms = append(perms, role.Permissions...)
		}
	}
	if identity == "user" {
		// Frontend users have a baseline set of permissions for publishing and interaction.
		perms = []string{"post.create", "core.topic.create", "comment.read", "qa.question.create", "qa.answer.create", "docs.document.create", "wiki.page.create", "projects.project.create", "jobs.job.create", "ai_works.work.create"}
	}
	return domain.AuthUser{
		ID:          u.ID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Email:       u.Email,
		Phone:       u.Phone,
		Status:      u.Status,
		RoleCode:    roleCode,
		RoleName:    roleName,
		Sites:       sites,
		Permissions: perms,
		TokenType:   identity,
		Identity:    identity,
		Audience:    map[string]string{"admin": "devhub_admin", "user": "devhub_frontend"}[identity],
	}
}

// GetSite 按 key 获取站点配置。
func (s *MemoryStore) GetSite(key string) (domain.Site, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site, ok := s.sites[key]
	return site, ok
}

// CreateSite 创建子站配置。
func (s *MemoryStore) CreateSite(req domain.Site) (domain.Site, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.TrimSpace(req.Key)
	if key == "" || key == "portal" {
		return domain.Site{}, errors.New("子站标识不合法")
	}
	if _, ok := s.sites[key]; ok {
		return domain.Site{}, errors.New("子站已存在")
	}
	site := domain.Site{
		Key:         key,
		Name:        strings.TrimSpace(req.Name),
		Logo:        strings.TrimSpace(req.Logo),
		Title:       strings.TrimSpace(req.Title),
		Sub:         strings.TrimSpace(req.Sub),
		Pub:         strings.TrimSpace(req.Pub),
		Description: strings.TrimSpace(req.Description),
		Color:       strings.TrimSpace(req.Color),
		Status:      strings.TrimSpace(req.Status),
		Sort:        req.Sort,
	}
	if site.Name == "" {
		site.Name = key
	}
	if site.Logo == "" {
		site.Logo = strings.ToUpper(firstRunes(key, 4))
	}
	if site.Title == "" {
		site.Title = site.Name + " 子网站"
	}
	if site.Pub == "" {
		site.Pub = "发布 " + site.Name + " 内容"
	}
	if site.Status == "" {
		site.Status = "enable"
	}
	s.sites[key] = site
	s.siteOrder = append(s.siteOrder, key)
	s.appendLogLocked("system", "admin", "新增子站", fmt.Sprintf("sites#%s", key), "127.0.0.1")
	return site, nil
}

// UpdateSite 更新站点配置，portal 不允许被禁用。
func (s *MemoryStore) UpdateSite(key string, req domain.Site) (domain.Site, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	site, ok := s.sites[key]
	if !ok {
		return domain.Site{}, false
	}
	oldName := site.Name
	oldTitle := site.Title
	oldPub := site.Pub
	site.Name = strings.TrimSpace(req.Name)
	site.Logo = strings.TrimSpace(req.Logo)
	site.Title = strings.TrimSpace(req.Title)
	site.Sub = strings.TrimSpace(req.Sub)
	site.Pub = strings.TrimSpace(req.Pub)
	site.Description = strings.TrimSpace(req.Description)
	site.Color = strings.TrimSpace(req.Color)
	site.Sort = req.Sort
	if req.Status != "" {
		if key == "portal" && req.Status == "disable" {
			req.Status = "enable"
		}
		site.Status = strings.TrimSpace(req.Status)
	}
	if key != "portal" && site.Name != "" && site.Name != oldName {
		if site.Title == "" || site.Title == oldTitle || site.Title == oldName+" 子网站" {
			site.Title = site.Name + " 子网站"
		}
		if site.Pub == "" || site.Pub == oldPub || site.Pub == "发布 "+oldName+" 内容" {
			site.Pub = "发布 " + site.Name + " 内容"
		}
	}
	s.sites[key] = site
	s.appendLogLocked("system", "admin", "更新子站配置", fmt.Sprintf("sites#%s", key), "127.0.0.1")
	return site, true
}

// ListSites 按展示顺序返回站点列表。
func (s *MemoryStore) ListSites() []domain.Site {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Site{s.sites["portal"]}
	for _, k := range s.siteOrder {
		out = append(out, s.sites[k])
	}
	return out
}

// ListBoards 按展示顺序返回板块列表。
func (s *MemoryStore) ListBoards() []domain.Board {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Board, 0, len(s.boardOrder))
	for _, k := range s.boardOrder {
		out = append(out, s.boards[k])
	}
	return out
}

// CreateBoard 创建板块配置。
func (s *MemoryStore) CreateBoard(req domain.Board) (domain.Board, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.TrimSpace(req.Key)
	if key == "" || key == "all" {
		return domain.Board{}, errors.New("板块标识不合法")
	}
	if _, ok := s.boards[key]; ok {
		return domain.Board{}, errors.New("板块已存在")
	}
	board := domain.Board{Key: key, Name: strings.TrimSpace(req.Name), Site: strings.TrimSpace(req.Site), Sort: req.Sort, Visible: req.Visible}
	if board.Name == "" {
		board.Name = key
	}
	if board.Site == "" {
		board.Site = "all"
	}
	s.boards[key] = board
	s.boardOrder = append(s.boardOrder, key)
	s.appendLogLocked("system", "admin", "新增板块", fmt.Sprintf("boards#%s", key), "127.0.0.1")
	return board, nil
}

// UpdateBoard 更新板块配置。
func (s *MemoryStore) UpdateBoard(key string, req domain.Board) (domain.Board, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	board, ok := s.boards[key]
	if !ok {
		return domain.Board{}, false
	}
	board.Name = strings.TrimSpace(req.Name)
	board.Site = strings.TrimSpace(req.Site)
	board.Sort = req.Sort
	board.Visible = req.Visible
	s.boards[key] = board
	s.appendLogLocked("system", "admin", "更新板块配置", fmt.Sprintf("boards#%s", key), "127.0.0.1")
	return board, true
}

// Plugins 返回系统插件注册信息。
func (s *MemoryStore) Plugins() []domain.Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Plugin, 0, len(pluginregistry.Definitions()))
	seen := map[string]bool{}
	for _, def := range pluginregistry.Definitions() {
		seen[def.Code] = true
		if runtime, ok := s.plugins[def.Code]; ok {
			plugin := pluginregistry.MergeRuntimeState(def, *runtime)
			out = append(out, withResolvedPluginConfig(plugin, runtime.ConfigJSON, ""))
			continue
		}
		out = append(out, withResolvedPluginConfig(def, "", ""))
	}
	for code, runtime := range s.plugins {
		if seen[code] {
			continue
		}
		out = append(out, withResolvedPluginConfig(*runtime, runtime.ConfigJSON, ""))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// PluginByCode 获取系统插件。
func (s *MemoryStore) PluginByCode(code string) (domain.Plugin, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	def, ok := pluginregistry.DefinitionByCode(code)
	if ok {
		if runtime, ok := s.plugins[def.Code]; ok {
			plugin := pluginregistry.MergeRuntimeState(def, *runtime)
			return withResolvedPluginConfig(plugin, runtime.ConfigJSON, ""), true
		}
		return withResolvedPluginConfig(def, "", ""), true
	}
	if runtime, ok := s.plugins[strings.TrimSpace(code)]; ok {
		return withResolvedPluginConfig(*runtime, runtime.ConfigJSON, ""), true
	}
	return domain.Plugin{}, false
}

func withResolvedPluginConfig(plugin domain.Plugin, globalConfigJSON, communityConfigJSON string) domain.Plugin {
	plugin.ConfigJSON = strings.TrimSpace(firstNonEmptyString(communityConfigJSON, globalConfigJSON))
	plugin.ResolvedConfig = pluginregistry.ResolvePluginConfig(plugin, globalConfigJSON, communityConfigJSON)
	return pluginregistry.ApplyLifecycle(plugin)
}

func (s *MemoryStore) SavePlugin(plugin domain.Plugin) (domain.Plugin, error) {
	plugin.Code = strings.TrimSpace(firstNonEmptyString(plugin.Code, plugin.PluginCode))
	if plugin.Code == "" {
		return domain.Plugin{}, errors.New("插件 code 不能为空")
	}
	if plugin.Status == "" {
		plugin.Status = pluginregistry.StatusDisabled
	}
	if !pluginregistry.ValidGlobalStatus(plugin.Status) {
		return domain.Plugin{}, errors.New("插件状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := Now()
	if existing, ok := s.plugins[plugin.Code]; ok {
		if plugin.CreatedAt == "" {
			plugin.CreatedAt = existing.CreatedAt
		}
		if plugin.ConfigJSON == "" {
			plugin.ConfigJSON = existing.ConfigJSON
		}
	}
	if plugin.CreatedAt == "" {
		plugin.CreatedAt = now
	}
	plugin.UpdatedAt = now
	cp := plugin
	s.plugins[plugin.Code] = &cp
	return withResolvedPluginConfig(plugin, plugin.ConfigJSON, ""), nil
}

// SetPluginStatus 设置插件运行状态。
func (s *MemoryStore) SetPluginStatus(code, status string) (domain.Plugin, error) {
	status = strings.TrimSpace(status)
	if !pluginregistry.ValidGlobalStatus(status) {
		return domain.Plugin{}, errors.New("插件状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		if runtime, exists := s.plugins[code]; exists {
			runtime.Status = status
			runtime.UpdatedAt = Now()
			return withResolvedPluginConfig(*runtime, runtime.ConfigJSON, ""), nil
		}
		return domain.Plugin{}, errors.New("插件不存在")
	}
	runtime, ok := s.plugins[def.Code]
	if !ok {
		cp := def
		runtime = &cp
		s.plugins[def.Code] = runtime
	}
	runtime.Status = status
	runtime.UpdatedAt = Now()
	plugin := pluginregistry.MergeRuntimeState(def, *runtime)
	return withResolvedPluginConfig(plugin, runtime.ConfigJSON, ""), nil
}

func (s *MemoryStore) SetPluginConfig(code, configJSON string) (domain.Plugin, error) {
	configJSON = strings.TrimSpace(configJSON)
	s.mu.Lock()
	defer s.mu.Unlock()
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		if runtime, exists := s.plugins[code]; exists {
			if err := pluginregistry.ValidateConfigJSON(*runtime, configJSON); err != nil {
				return domain.Plugin{}, err
			}
			runtime.ConfigJSON = configJSON
			runtime.UpdatedAt = Now()
			return withResolvedPluginConfig(*runtime, runtime.ConfigJSON, ""), nil
		}
		return domain.Plugin{}, errors.New("插件不存在")
	}
	if err := pluginregistry.ValidateConfigJSON(def, configJSON); err != nil {
		return domain.Plugin{}, err
	}
	runtime, ok := s.plugins[def.Code]
	if !ok {
		cp := def
		runtime = &cp
		s.plugins[def.Code] = runtime
	}
	runtime.ConfigJSON = configJSON
	runtime.UpdatedAt = Now()
	plugin := pluginregistry.MergeRuntimeState(def, *runtime)
	return withResolvedPluginConfig(plugin, runtime.ConfigJSON, ""), nil
}

func (s *MemoryStore) CommunityPlugins(communityID int64) ([]domain.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.communities[communityID]; !ok {
		return nil, errors.New("子站不存在")
	}
	runtime := s.communityPlugins[communityID]
	out := make([]domain.Plugin, 0, len(pluginregistry.Definitions()))
	for _, def := range pluginregistry.Definitions() {
		merged := def
		if global, ok := s.plugins[def.Code]; ok {
			merged = pluginregistry.MergeRuntimeState(def, *global)
		}
		merged.GlobalStatus = merged.Status
		merged.CommunityStatus = pluginregistry.StatusDisabled
		if cp, ok := runtime[def.Code]; ok && cp != nil {
			merged.CommunityStatus = cp.Status
			merged.SortOrder = cp.SortOrder
			merged = withResolvedPluginConfig(merged, merged.ConfigJSON, cp.ConfigJSON)
		} else {
			merged = withResolvedPluginConfig(merged, merged.ConfigJSON, "")
		}
		if merged.GlobalStatus == pluginregistry.StatusEnabled && merged.CommunityStatus == pluginregistry.StatusEnabled {
			merged.Status = pluginregistry.StatusEnabled
		} else {
			merged.Status = pluginregistry.StatusDisabled
		}
		out = append(out, pluginregistry.ApplyLifecycle(merged))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func (s *MemoryStore) SetCommunityPluginStatus(communityID int64, code, status string) (domain.Plugin, error) {
	status = strings.TrimSpace(status)
	if status != pluginregistry.StatusEnabled && status != pluginregistry.StatusDisabled {
		return domain.Plugin{}, errors.New("插件状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.communities[communityID]; !ok {
		return domain.Plugin{}, errors.New("子站不存在")
	}
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	global, ok := s.plugins[def.Code]
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	if global.Status != pluginregistry.StatusEnabled && status == pluginregistry.StatusEnabled {
		return domain.Plugin{}, errors.New("插件全局未启用，不能在子站启用")
	}
	if _, ok := s.communityPlugins[communityID]; !ok {
		s.communityPlugins[communityID] = map[string]*domain.CommunityPlugin{}
	}
	cp := s.communityPlugins[communityID][def.Code]
	if cp == nil {
		cp = &domain.CommunityPlugin{CommunityID: communityID, PluginCode: def.Code, CreatedAt: Now()}
		s.communityPlugins[communityID][def.Code] = cp
	}
	cp.Status = status
	cp.UpdatedAt = Now()
	plugin := pluginregistry.MergeRuntimeState(def, *global)
	plugin.GlobalStatus = global.Status
	plugin.CommunityStatus = cp.Status
	plugin = withResolvedPluginConfig(plugin, global.ConfigJSON, cp.ConfigJSON)
	if plugin.GlobalStatus == pluginregistry.StatusEnabled && plugin.CommunityStatus == pluginregistry.StatusEnabled {
		plugin.Status = pluginregistry.StatusEnabled
	} else {
		plugin.Status = pluginregistry.StatusDisabled
	}
	return plugin, nil
}

func (s *MemoryStore) SetCommunityPluginConfig(communityID int64, code, configJSON string) (domain.Plugin, error) {
	configJSON = strings.TrimSpace(configJSON)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.communities[communityID]; !ok {
		return domain.Plugin{}, errors.New("子站不存在")
	}
	def, ok := pluginregistry.DefinitionByCode(code)
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	if err := pluginregistry.ValidateConfigJSON(def, configJSON); err != nil {
		return domain.Plugin{}, err
	}
	global, ok := s.plugins[def.Code]
	if !ok {
		return domain.Plugin{}, errors.New("插件不存在")
	}
	if _, ok := s.communityPlugins[communityID]; !ok {
		s.communityPlugins[communityID] = map[string]*domain.CommunityPlugin{}
	}
	cp := s.communityPlugins[communityID][def.Code]
	if cp == nil {
		cp = &domain.CommunityPlugin{CommunityID: communityID, PluginCode: def.Code, Status: pluginregistry.StatusDisabled, CreatedAt: Now()}
		s.communityPlugins[communityID][def.Code] = cp
	}
	cp.ConfigJSON = configJSON
	cp.UpdatedAt = Now()
	plugin := pluginregistry.MergeRuntimeState(def, *global)
	plugin.GlobalStatus = global.Status
	plugin.CommunityStatus = cp.Status
	plugin = withResolvedPluginConfig(plugin, global.ConfigJSON, cp.ConfigJSON)
	if plugin.GlobalStatus == pluginregistry.StatusEnabled && plugin.CommunityStatus == pluginregistry.StatusEnabled {
		plugin.Status = pluginregistry.StatusEnabled
	} else {
		plugin.Status = pluginregistry.StatusDisabled
	}
	return plugin, nil
}

func (s *MemoryStore) ReorderCommunityPlugins(communityID int64, codes []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.communities[communityID]; !ok {
		return 0, errors.New("子站不存在")
	}
	if _, ok := s.communityPlugins[communityID]; !ok {
		s.communityPlugins[communityID] = map[string]*domain.CommunityPlugin{}
	}
	updated := 0
	for i, code := range codes {
		def, ok := pluginregistry.DefinitionByCode(code)
		if !ok {
			continue
		}
		cp := s.communityPlugins[communityID][def.Code]
		if cp == nil {
			cp = &domain.CommunityPlugin{CommunityID: communityID, PluginCode: def.Code, Status: pluginregistry.StatusDisabled, CreatedAt: Now()}
			s.communityPlugins[communityID][def.Code] = cp
		}
		cp.SortOrder = i
		cp.UpdatedAt = Now()
		updated++
	}
	return updated, nil
}

// ListPosts 按站点、板块、关键词和标签筛选帖子，并按 ID 倒序返回。
func (s *MemoryStore) ListPosts(site, board, q, tag string) []domain.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site = strings.TrimSpace(site)
	board = strings.TrimSpace(board)
	q = strings.ToLower(strings.TrimSpace(q))
	tag = strings.ToLower(strings.TrimSpace(tag))
	out := make([]domain.Post, 0)
	for _, p := range s.posts {
		if !memoryPostVisible(p) {
			continue
		}
		if site != "" && site != "portal" && p.Site != site {
			continue
		}
		if board != "" && board != "all" && p.Board != board {
			continue
		}
		if tag != "" && !hasTag(p.Tags, tag) {
			continue
		}
		if q != "" && !postContains(p, q) {
			continue
		}
		cp := *p
		cp.Tags = append([]string(nil), p.Tags...)
		cp.CommentLocked = s.commentLocks[p.ID]
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

// GetPost 获取帖子详情，increaseView 为 true 时增加浏览量。
func (s *MemoryStore) GetPost(id int64, increaseView bool) (*domain.Post, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok || !memoryPostVisible(p) {
		return nil, false
	}
	if increaseView {
		p.Views++
	}
	cp := *p
	cp.Tags = append([]string(nil), p.Tags...)
	cp.CommentLocked = s.commentLocks[p.ID]
	return &cp, true
}

// CreatePost 创建帖子并生成发布通知。
func (s *MemoryStore) CreatePost(req domain.CreatePostRequest) (*domain.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.Site == "portal" || req.Site == "" {
		return nil, errors.New("帖子必须发布到具体子网站：php / go / java")
	}
	if _, ok := s.sites[req.Site]; !ok {
		return nil, errors.New("无效子网站")
	}
	if req.Board == "all" || req.Board == "" {
		return nil, errors.New("帖子必须发布到具体板块")
	}
	if _, ok := s.boards[req.Board]; !ok {
		return nil, errors.New("无效板块")
	}
	p := &domain.Post{Site: req.Site, Board: req.Board, Title: strings.TrimSpace(req.Title), Summary: strings.TrimSpace(req.Summary), Content: strings.TrimSpace(req.Content), Author: strings.TrimSpace(req.Author), Tags: req.Tags}
	if p.Summary == "" {
		p.Summary = firstRunes(p.Content, 80)
	}
	created := s.createPostLocked(p)
	s.rebuildTagsFromPostsLocked()
	s.createNoticeLocked("你发布了新的内容", created.Title)
	cp := *created
	cp.Tags = append([]string(nil), created.Tags...)
	return &cp, nil
}

// UpdatePost 部分更新帖子字段。
func (s *MemoryStore) UpdatePost(id int64, req domain.UpdatePostRequest) (*domain.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("帖子不存在")
	}
	if req.Site != nil {
		if *req.Site == "portal" || *req.Site == "" {
			return nil, errors.New("帖子必须属于具体子网站")
		}
		if _, ok := s.sites[*req.Site]; !ok {
			return nil, errors.New("无效子网站")
		}
		p.Site = *req.Site
	}
	if req.Board != nil {
		if *req.Board == "all" || *req.Board == "" {
			return nil, errors.New("帖子必须属于具体板块")
		}
		if _, ok := s.boards[*req.Board]; !ok {
			return nil, errors.New("无效板块")
		}
		p.Board = *req.Board
	}
	if req.Title != nil {
		p.Title = strings.TrimSpace(*req.Title)
	}
	if req.Summary != nil {
		p.Summary = strings.TrimSpace(*req.Summary)
	}
	if req.Content != nil {
		p.Content = strings.TrimSpace(*req.Content)
	}
	if req.Status != nil {
		p.Status = strings.TrimSpace(*req.Status)
	}
	if req.Pinned != nil {
		p.Pinned = *req.Pinned
	}
	if req.Recommended != nil {
		p.Recommended = *req.Recommended
	}
	if req.RejectReason != nil {
		p.RejectReason = strings.TrimSpace(*req.RejectReason)
	}
	if req.OfflineReason != nil {
		p.OfflineReason = strings.TrimSpace(*req.OfflineReason)
	}
	if req.Tags != nil {
		p.Tags = uniqueTags(*req.Tags)
	}
	p.UpdatedAt = Now()
	s.rebuildTagsFromPostsLocked()
	s.appendLogLocked("operation", "admin", "更新帖子", fmt.Sprintf("posts#%d", id), "127.0.0.1")
	cp := *p
	cp.Tags = append([]string(nil), p.Tags...)
	return &cp, nil
}

// DeletePost 删除帖子及其所有评论。
func (s *MemoryStore) DeletePost(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.posts[id]; !ok {
		return false
	}
	delete(s.posts, id)
	for cid, c := range s.comments {
		if c.PostID == id {
			delete(s.comments, cid)
		}
	}
	s.rebuildTagsFromPostsLocked()
	s.appendLogLocked("operation", "admin", "删除帖子", fmt.Sprintf("posts#%d", id), "127.0.0.1")
	return true
}

// LikePost 增加帖子点赞数并生成通知。
func (s *MemoryStore) LikePost(id int64) (*domain.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("帖子不存在")
	}
	p.Likes++
	p.UpdatedAt = Now()
	s.createNoticeLocked("你的帖子获得了新的点赞", p.Title)
	cp := *p
	cp.Tags = append([]string(nil), p.Tags...)
	return &cp, nil
}

// HotPosts 按热度分数返回帖子列表。
func (s *MemoryStore) HotPosts(site string, limit int) []domain.Post {
	out := s.ListPosts(site, "all", "", "")
	sort.Slice(out, func(i, j int) bool {
		si := out[i].Views + out[i].Likes*8 + out[i].Comments*15
		sj := out[j].Views + out[j].Likes*8 + out[j].Comments*15
		return si > sj
	})
	return limitPosts(out, limit)
}

// Feed 按发布时间返回帖子列表。
func (s *MemoryStore) Feed(site string, limit int) []domain.Post {
	out := s.ListPosts(site, "all", "", "")
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return limitPosts(out, limit)
}

// TagStats 统计站点内标签使用次数。
func (s *MemoryStore) TagStats(site string) []domain.TagStat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tagStatsLocked(site, "", "enable", 0)
}

func (s *MemoryStore) AdminTags(site, q, status string) []domain.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q = strings.ToLower(strings.TrimSpace(q))
	status = strings.TrimSpace(status)
	out := make([]domain.Tag, 0, len(s.tags))
	for _, tag := range s.tags {
		if tag == nil {
			continue
		}
		cp := s.enrichTagLocked(*tag)
		if site != "" && site != "portal" && cp.Site != site && cp.CommunitySlug != site {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(tag.Name+" "+tag.Slug+" "+tag.Description), q) {
			continue
		}
		if status == "" || status == "all" {
			if cp.Status == "merged" {
				continue
			}
		} else if cp.Status != status {
			continue
		}
		out = append(out, cp)
	}
	sortTags(out)
	return out
}

func (s *MemoryStore) AdminTagByID(id int64) (domain.Tag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tag, ok := s.tags[id]
	if !ok || tag == nil {
		return domain.Tag{}, false
	}
	return s.enrichTagLocked(*tag), true
}

func (s *MemoryStore) CreateTag(req domain.Tag) (domain.Tag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tag := normalizeTag(req)
	if tag.SortOrder != 0 && tag.Sort == 0 {
		tag.Sort = tag.SortOrder
	}
	if tag.Name == "" {
		return domain.Tag{}, errors.New("标签名称不能为空")
	}
	if _, ok := s.findTagBySiteSlugLocked(tag.Site, tag.Slug); ok {
		return domain.Tag{}, errors.New("标签已存在")
	}
	now := Now()
	tag.ID = s.nextTagID
	s.nextTagID++
	tag.CreatedAt = now
	tag.UpdatedAt = now
	cp := tag
	s.tags[cp.ID] = &cp
	s.appendLogLocked("operation", "admin", "新增标签", tag.Name, "127.0.0.1")
	return s.enrichTagLocked(cp), nil
}

func (s *MemoryStore) UpdateTag(id int64, req domain.Tag) (domain.Tag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.tags[id]
	if !ok || current == nil {
		return domain.Tag{}, false
	}
	tag := normalizeTag(req)
	if tag.Name == "" {
		return domain.Tag{}, false
	}
	if tag.SortOrder != 0 && tag.Sort == 0 {
		tag.Sort = tag.SortOrder
	}
	for _, item := range s.tags {
		if item == nil || item.ID == id {
			continue
		}
		if item.Site == tag.Site && item.Slug == tag.Slug {
			return domain.Tag{}, false
		}
	}
	tag.ID = id
	tag.CreatedAt = current.CreatedAt
	tag.UpdatedAt = Now()
	tag.UseCount = current.UseCount
	tag.TopicCount = current.TopicCount
	tag.FollowerCount = current.FollowerCount
	tag.HotScore = current.HotScore
	tag.MergedToID = current.MergedToID
	*current = tag
	s.appendLogLocked("operation", "admin", "更新标签", fmt.Sprintf("tags#%d", id), "127.0.0.1")
	return s.enrichTagLocked(*current), true
}

func (s *MemoryStore) SetTagStatus(id int64, status string) (domain.Tag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tag, ok := s.tags[id]
	if !ok || tag == nil {
		return domain.Tag{}, false
	}
	status = normalizeTagStatus(status)
	if status == "" {
		return domain.Tag{}, false
	}
	tag.Status = status
	tag.UpdatedAt = Now()
	s.appendLogLocked("operation", "admin", "更新标签状态", fmt.Sprintf("tags#%d:%s", id, status), "127.0.0.1")
	return s.enrichTagLocked(*tag), true
}

func (s *MemoryStore) TagBySlug(site, slugOrName string) (domain.Tag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tagBySlugLocked(site, slugOrName, true)
}

func (s *MemoryStore) ResolveTag(site, slugOrName string) (domain.TagResolveResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resolveTagLocked(site, slugOrName, true)
}

func (s *MemoryStore) TagSuggestions(site, q string, limit int) []domain.TagStat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q = strings.TrimSpace(q)
	items := s.tagStatsLocked(site, "", "enable", 0)
	if q == "" {
		if limit > 0 && len(items) > limit {
			return items[:limit]
		}
		return items
	}
	matched := make([]domain.TagStat, 0, len(items))
	seen := map[int64]bool{}
	needle := strings.ToLower(q)
	slugNeedle := normalizeSlug(q)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name+" "+item.Slug+" "+item.Description), needle) {
			matched = append(matched, item)
			seen[item.ID] = true
		}
	}
	for _, alias := range s.tagAliases {
		if alias == nil {
			continue
		}
		if site != "" && site != "portal" && alias.Site != site {
			continue
		}
		if !strings.Contains(strings.ToLower(alias.Alias+" "+alias.AliasSlug), needle) && (slugNeedle == "" || !strings.Contains(strings.ToLower(alias.AliasSlug), slugNeedle)) {
			continue
		}
		tag, ok := s.tags[alias.TagID]
		if !ok || tag == nil || tag.Status != "enable" {
			continue
		}
		stat := memoryTagToStat(s.enrichTagLocked(*tag))
		stat.MatchedAlias = alias.Alias
		if !seen[stat.ID] {
			matched = append(matched, stat)
			seen[stat.ID] = true
		}
	}
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].MatchedAlias == "" && matched[j].MatchedAlias != "" {
			return false
		}
		if matched[i].MatchedAlias != "" && matched[j].MatchedAlias == "" {
			return true
		}
		if matched[i].Count == matched[j].Count {
			return matched[i].Name < matched[j].Name
		}
		return matched[i].Count > matched[j].Count
	})
	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}
	return matched
}

func (s *MemoryStore) TagTopics(tagID int64, communityID int64, contentType string, sortBy string, page, pageSize int) ([]domain.Topic, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tag, ok := s.tags[tagID]
	if !ok || tag == nil || tag.Status != "enable" {
		return []domain.Topic{}, 0
	}
	contentType = strings.TrimSpace(contentType)
	topics := make([]domain.Topic, 0)
	for _, post := range s.posts {
		if !memoryPostVisible(post) {
			continue
		}
		topic, ok := s.topicFromPostLocked(post.ID, false)
		if !ok {
			continue
		}
		if communityID > 0 && topic.CommunityID != communityID {
			continue
		}
		if contentType != "" && contentType != "all" && topic.ContentType != contentType {
			continue
		}
		if sortBy == "unsolved" && (topic.ContentType != "question" || topic.IsSolved) {
			continue
		}
		if tag.Site != "" && tag.Site != "portal" && post.Site != tag.Site {
			continue
		}
		if !memoryTopicHasTag(topic, tag.Name) && !memoryTopicHasTag(topic, tag.Slug) {
			continue
		}
		topics = append(topics, topic)
	}
	sortTopicsForSearch(topics, sortBy)
	return paginateTopics(topics, page, pageSize)
}

func (s *MemoryStore) AdminTagTopics(id int64, page, pageSize int) ([]domain.Topic, int) {
	s.mu.RLock()
	tag, ok := s.tags[id]
	if !ok || tag == nil {
		s.mu.RUnlock()
		return []domain.Topic{}, 0
	}
	communityID := int64(0)
	if tag.Site != "" && tag.Site != "portal" {
		communityID = s.communityIDBySlugLocked(tag.Site)
	}
	tagID := tag.ID
	s.mu.RUnlock()
	return s.TagTopics(tagID, communityID, "", "latest", page, pageSize)
}

func (s *MemoryStore) TagAliases(tagID int64) ([]domain.TagAlias, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.tags[tagID]; !ok {
		return nil, errors.New("标签不存在")
	}
	return s.tagAliasListLocked(tagID), nil
}

func (s *MemoryStore) AddTagAlias(tagID int64, alias string) (domain.TagAlias, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tag, ok := s.tags[tagID]
	if !ok || tag == nil {
		return domain.TagAlias{}, errors.New("标签不存在")
	}
	if tag.Status != "enable" {
		return domain.TagAlias{}, errors.New("仅启用标签可添加别名")
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return domain.TagAlias{}, errors.New("别名不能为空")
	}
	if len([]rune(alias)) > 50 {
		return domain.TagAlias{}, errors.New("别名长度不能超过 50 个字符")
	}
	aliasSlug := normalizeSlug(alias)
	if aliasSlug == "" {
		aliasSlug = strings.ToLower(strings.Join(strings.Fields(alias), "-"))
	}
	if aliasSlug == "" {
		return domain.TagAlias{}, errors.New("别名 slug 不能为空")
	}
	if _, ok := s.findTagBySiteSlugLocked(tag.Site, aliasSlug); ok {
		return domain.TagAlias{}, errors.New("别名 slug 与现有标签冲突")
	}
	if _, ok := s.findTagAliasBySiteSlugLocked(tag.Site, aliasSlug); ok {
		return domain.TagAlias{}, errors.New("别名 slug 已存在")
	}
	for _, item := range s.tagAliases {
		if item == nil || item.TagID != tagID {
			continue
		}
		if strings.EqualFold(item.AliasSlug, aliasSlug) {
			return domain.TagAlias{}, errors.New("别名已存在")
		}
	}
	now := Now()
	item := domain.TagAlias{
		ID:            s.nextTagAliasID,
		TagID:         tagID,
		Site:          tag.Site,
		CommunityID:   s.communityIDBySlugLocked(tag.Site),
		CommunitySlug: tag.Site,
		Alias:         alias,
		AliasSlug:     aliasSlug,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.nextTagAliasID++
	cp := item
	s.tagAliases[item.ID] = &cp
	s.appendLogLocked("audit", "admin", "新增标签别名", fmt.Sprintf("tags#%d/aliases#%d", tagID, item.ID), "127.0.0.1")
	return item, nil
}

func (s *MemoryStore) DeleteTagAlias(tagID, aliasID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tags[tagID]; !ok {
		return errors.New("标签不存在")
	}
	item, ok := s.tagAliases[aliasID]
	if !ok || item == nil || item.TagID != tagID {
		return errors.New("标签别名不存在")
	}
	delete(s.tagAliases, aliasID)
	s.appendLogLocked("audit", "admin", "删除标签别名", fmt.Sprintf("tags#%d/aliases#%d", tagID, aliasID), "127.0.0.1")
	return nil
}

func (s *MemoryStore) RecalculateTag(tagID int64) (domain.Tag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tag, ok := s.tags[tagID]
	if !ok || tag == nil {
		return domain.Tag{}, errors.New("标签不存在")
	}
	updated := s.enrichTagLocked(*tag)
	tag.UseCount = updated.UseCount
	tag.FollowerCount = updated.FollowerCount
	tag.HotScore = updated.HotScore
	tag.UpdatedAt = Now()
	if tag.Status == "merged" {
		tag.UseCount = 0
		tag.TopicCount = 0
		tag.FollowerCount = 0
		tag.HotScore = 0
	}
	s.appendLogLocked("audit", "admin", "重算标签统计", fmt.Sprintf("tags#%d", tagID), "127.0.0.1")
	return s.enrichTagLocked(*tag), nil
}

func (s *MemoryStore) RecalculateAllTags() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := 0
	now := Now()
	for _, tag := range s.tags {
		if tag == nil {
			continue
		}
		enriched := s.enrichTagLocked(*tag)
		tag.UseCount = enriched.UseCount
		tag.FollowerCount = enriched.FollowerCount
		tag.HotScore = enriched.HotScore
		if tag.Status == "merged" {
			tag.UseCount = 0
			tag.FollowerCount = 0
			tag.HotScore = 0
		}
		tag.UpdatedAt = now
		updated++
	}
	s.appendLogLocked("audit", "admin", "批量重算标签统计", fmt.Sprintf("tags:%d", updated), "127.0.0.1")
	return updated, nil
}

func (s *MemoryStore) MergeTag(sourceTagID, targetTagID int64) (domain.Tag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sourceTagID <= 0 || targetTagID <= 0 || sourceTagID == targetTagID {
		return domain.Tag{}, errors.New("源标签和目标标签不能相同")
	}
	source, ok := s.tags[sourceTagID]
	if !ok || source == nil {
		return domain.Tag{}, errors.New("源标签不存在")
	}
	target, ok := s.tags[targetTagID]
	if !ok || target == nil {
		return domain.Tag{}, errors.New("目标标签不存在")
	}
	if source.Status == "merged" {
		return domain.Tag{}, errors.New("源标签已合并")
	}
	if target.Status != "enable" {
		return domain.Tag{}, errors.New("目标标签必须为启用状态")
	}
	if source.Site != target.Site {
		return domain.Tag{}, errors.New("仅支持同一子站范围合并")
	}
	names := []string{source.Name, source.Slug}
	for _, alias := range s.tagAliasListLocked(sourceTagID) {
		names = append(names, alias.Alias, alias.AliasSlug)
	}
	for _, post := range s.posts {
		if post == nil || post.Site != source.Site {
			continue
		}
		replaced := false
		for _, item := range names {
			if item == "" || !hasTagOrSlug(post.Tags, item, normalizeSlug(item)) {
				continue
			}
			replaced = true
			break
		}
		if replaced {
			next := make([]string, 0, len(post.Tags)+1)
			for _, tagName := range post.Tags {
				keep := true
				for _, item := range names {
					if strings.EqualFold(strings.TrimSpace(tagName), strings.TrimSpace(item)) || normalizeSlug(tagName) == normalizeSlug(item) {
						keep = false
						break
					}
				}
				if keep {
					next = append(next, tagName)
				}
			}
			next = append(next, target.Name)
			post.Tags = uniqueTags(next)
			post.UpdatedAt = Now()
		}
	}
	for _, follow := range s.follows {
		if follow == nil || follow.TargetType != "tag" || follow.TargetID != sourceTagID {
			continue
		}
		key := followKey(follow.UserID, "tag", targetTagID)
		if _, exists := s.follows[key]; !exists {
			follow.TargetID = targetTagID
			follow.UpdatedAt = Now()
			s.follows[key] = follow
		}
		delete(s.follows, followKey(follow.UserID, "tag", sourceTagID))
	}
	for aliasID, alias := range s.tagAliases {
		if alias == nil || alias.TagID != sourceTagID {
			continue
		}
		if alias.AliasSlug == target.Slug {
			delete(s.tagAliases, aliasID)
			continue
		}
		if existing, ok := s.findTagAliasBySiteSlugLocked(target.Site, alias.AliasSlug); ok && existing.TagID != sourceTagID {
			delete(s.tagAliases, aliasID)
			continue
		}
		alias.TagID = targetTagID
		alias.Site = target.Site
		alias.UpdatedAt = Now()
	}
	source.Status = "merged"
	source.MergedToID = targetTagID
	source.UseCount = 0
	source.FollowerCount = 0
	source.HotScore = 0
	source.UpdatedAt = Now()
	target.UpdatedAt = Now()
	enriched := s.enrichTagLocked(*target)
	target.UseCount = enriched.UseCount
	target.FollowerCount = enriched.FollowerCount
	target.HotScore = enriched.HotScore
	s.appendLogLocked("audit", "admin", "合并标签", fmt.Sprintf("tags#%d->%d", sourceTagID, targetTagID), "127.0.0.1")
	return s.enrichTagLocked(*target), nil
}

// BoardCounts 统计站点内各板块帖子数量，可叠加关键词过滤。
func (s *MemoryStore) BoardCounts(site, q string) map[string]int {
	result := map[string]int{}
	posts := s.ListPosts(site, "all", q, "")
	result["all"] = len(posts)
	for _, b := range s.boardOrder {
		if b != "all" {
			result[b] = 0
		}
	}
	for _, p := range posts {
		result[p.Board]++
	}
	return result
}

// PostStats 汇总站点内帖子、浏览、点赞和评论数。
func (s *MemoryStore) PostStats(site string) domain.PostStats {
	posts := s.ListPosts(site, "all", "", "")
	stats := domain.PostStats{TotalPosts: len(posts)}
	for _, p := range posts {
		stats.TotalViews += p.Views
		stats.TotalLikes += p.Likes
		stats.TotalComments += p.Comments
	}
	return stats
}

// CommentsTree 将指定帖子的评论组装成树形结构。
func (s *MemoryStore) CommentsTree(postID int64) []*domain.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	roots, _ := s.commentsTreeLocked(postID, "", 1, 1000)
	return roots
}

func (s *MemoryStore) commentsTreeLocked(postID int64, sortBy string, page, pageSize int) ([]*domain.Comment, int) {
	nodes := map[int64]*domain.Comment{}
	ids := make([]int, 0)
	for _, c := range s.comments {
		if c.PostID == postID && c.Status != "deleted" && c.Status != "hidden" {
			ids = append(ids, int(c.ID))
		}
	}
	sort.Ints(ids)
	for _, id := range ids {
		c := s.comments[int64(id)]
		cp := *c
		s.normalizeCommentLocked(&cp)
		cp.Replies = nil
		nodes[cp.ID] = &cp
	}
	roots := []*domain.Comment{}
	for _, id := range ids {
		c := nodes[int64(id)]
		if c.ParentID > 0 {
			if parent, ok := nodes[c.ParentID]; ok {
				parent.Replies = append(parent.Replies, c)
				continue
			}
		}
		roots = append(roots, c)
	}
	switch sortBy {
	case "latest":
		sort.Slice(roots, func(i, j int) bool { return roots[i].ID > roots[j].ID })
	case "best":
		sort.Slice(roots, func(i, j int) bool {
			if roots[i].IsBest != roots[j].IsBest {
				return roots[i].IsBest
			}
			return roots[i].ID < roots[j].ID
		})
	default:
		sort.Slice(roots, func(i, j int) bool { return roots[i].ID < roots[j].ID })
	}
	total := len(roots)
	return paginateSlice(roots, page, pageSize), total
}

// LikeComment 增加评论点赞数。
func (s *MemoryStore) LikeComment(id int64) (*domain.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.comments[id]
	if !ok {
		return nil, errors.New("评论不存在")
	}
	c.Likes++
	cp := *c
	return &cp, nil
}

// DeleteOwnComment 仅允许评论作者删除自己的评论及其子回复。
func (s *MemoryStore) DeleteOwnComment(id int64, author string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	author = strings.TrimSpace(author)
	if author == "" {
		return errors.New("缺少评论作者")
	}
	c, ok := s.comments[id]
	if !ok {
		return errors.New("评论不存在")
	}
	if c.Author != author {
		return errors.New("只能删除自己的评论")
	}
	removed := s.deleteCommentCascadeLocked(id)
	if p, ok := s.posts[c.PostID]; ok {
		p.Comments -= removed
		if p.Comments < 0 {
			p.Comments = 0
		}
		p.UpdatedAt = Now()
	}
	s.appendLogLocked("audit", author, "删除自己的评论", fmt.Sprintf("comments#%d", id), "127.0.0.1")
	return nil
}

// DeleteComment 级联删除评论及其子回复。
func (s *MemoryStore) DeleteComment(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.comments[id]
	if !ok {
		return false
	}
	removed := s.deleteCommentCascadeLocked(id)
	if p, ok := s.posts[c.PostID]; ok {
		p.Comments -= removed
		if p.Comments < 0 {
			p.Comments = 0
		}
		p.UpdatedAt = Now()
	}
	s.appendLogLocked("audit", "auditor", "删除评论", fmt.Sprintf("comments#%d", id), "127.0.0.1")
	return true
}

// deleteCommentCascadeLocked 在调用方已持有写锁时递归删除评论树。
func (s *MemoryStore) deleteCommentCascadeLocked(id int64) int {
	removed := 0
	for cid, c := range s.comments {
		if c.ParentID == id {
			removed += s.deleteCommentCascadeLocked(cid)
		}
	}
	if _, ok := s.comments[id]; ok {
		delete(s.comments, id)
		removed++
	}
	return removed
}

// Notices 返回通知列表，按 ID 倒序排列。
func (s *MemoryStore) Notices(site string) []domain.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Notification, 0, len(s.notices))
	for _, n := range s.notices {
		if !notificationInSite(*n, site) {
			continue
		}
		out = append(out, *n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

// ReadNotice 将指定通知标记为已读。
func (s *MemoryStore) ReadNotice(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.notices[id]
	if !ok {
		return false
	}
	n.Read = true
	n.IsRead = true
	n.ReadAt = Now()
	return true
}

// ReadAllNotices 将所有未读通知标记为已读，并返回更新数量。
func (s *MemoryStore) ReadAllNotices(site string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, n := range s.notices {
		if !notificationInSite(*n, site) {
			continue
		}
		if !n.Read {
			n.Read = true
			n.IsRead = true
			if n.ReadAt == "" {
				n.ReadAt = Now()
			}
			count++
		}
	}
	return count
}

// UnreadNoticeCount 返回未读通知数量。
func (s *MemoryStore) UnreadNoticeCount(site string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, n := range s.notices {
		if !n.Read && notificationInSite(*n, site) {
			count++
		}
	}
	return count
}

// UserProfile 返回演示用户资料和统计。
func (s *MemoryStore) UserProfile() domain.UserProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	totalPosts := 0
	totalComments := 0
	totalLikes := 0
	for _, p := range s.posts {
		if p.Author == "SUI.CHEN" {
			totalPosts++
		}
		totalLikes += p.Likes
	}
	for _, c := range s.comments {
		if c.Author == "SUI.CHEN" {
			totalComments++
		}
	}
	return domain.UserProfile{ID: 1, Name: "SUI.CHEN", Bio: "DevHub 用户主页 · 关注 PHP / Go / Java 技术内容", Posts: totalPosts, Comments: totalComments, Likes: totalLikes}
}

// AdminOverview 返回后台首页聚合统计。
func (s *MemoryStore) AdminOverview(site string) domain.AdminOverview {
	stats := s.PostStats(site)
	posts := s.ListPosts(site, "all", "", "")
	status := map[string]int{"draft": 0, "review": 0, "publish": 0, "offline": 0}
	siteStats := map[string]*domain.SiteStat{}
	boardStats := map[string]*domain.BoardStat{}
	for _, siteKey := range append([]string{"portal"}, s.siteOrder...) {
		if site != "" && site != "portal" && siteKey != site {
			continue
		}
		siteStats[siteKey] = &domain.SiteStat{Site: siteKey}
	}
	for _, board := range s.boardOrder {
		boardStats[board] = &domain.BoardStat{Board: board}
	}
	for _, p := range posts {
		status[p.Status]++
		if st, ok := siteStats[p.Site]; ok {
			st.Posts++
			st.Views += p.Views
			st.Likes += p.Likes
			st.Comments += p.Comments
		}
		if st, ok := boardStats[p.Board]; ok {
			st.Posts++
			st.Views += p.Views
			st.Likes += p.Likes
			st.Comments += p.Comments
		}
	}
	siteOut := make([]domain.SiteStat, 0, len(s.siteOrder))
	for _, siteKey := range s.siteOrder {
		if st, ok := siteStats[siteKey]; ok {
			siteOut = append(siteOut, *st)
		}
	}
	boardOut := make([]domain.BoardStat, 0, len(s.boardOrder))
	for _, board := range s.boardOrder {
		if board != "all" {
			boardOut = append(boardOut, *boardStats[board])
		}
	}
	return domain.AdminOverview{
		Stats:              stats,
		StatusDistribution: status,
		SiteStats:          siteOut,
		BoardStats:         boardOut,
		TopPosts:           s.HotPosts(site, 10),
		SearchKeywords: []domain.KeywordStat{
			{Keyword: "JVM", Count: 36, Scope: "java"},
			{Keyword: "Laravel", Count: 28, Scope: "php"},
			{Keyword: "Goroutine", Count: 21, Scope: "go"},
			{Keyword: "context", Count: 18, Scope: "go"},
		},
		UserStats: domain.UserAdminStats{TotalUsers: len(s.users), ActiveUsers: 3, Forbidden: 0, NewThisWeek: 1},
	}
}

// AdminUsers 返回后台用户列表，并补充每个用户的帖子和评论数。
func (s *MemoryStore) AdminUsers() []domain.AdminUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.adminUsersLocked()
}

func (s *MemoryStore) adminUsersLocked() []domain.AdminUser {
	out := make([]domain.AdminUser, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		for _, p := range s.posts {
			if p.Author == cp.Username || p.Author == cp.Nickname {
				cp.Posts++
			}
		}
		for _, c := range s.comments {
			if c.Author == cp.Username || c.Author == cp.Nickname {
				cp.Comments++
			}
		}
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// UpdateUserStatus 更新用户状态和违规备注。
func (s *MemoryStore) UpdateUserStatus(id int64, status, note string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return false
	}
	u.Status = strings.TrimSpace(status)
	u.ViolationNote = strings.TrimSpace(note)
	s.appendLogLocked("operation", "admin", "更新用户状态", fmt.Sprintf("users#%d", id), "127.0.0.1")
	return true
}

// AdminRoles 返回后台角色列表。
func (s *MemoryStore) AdminRoles() []domain.AdminRole {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AdminRole, 0, len(s.roles))
	for _, role := range s.roles {
		out = append(out, role)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// AdminPermissions 返回后台权限点清单。
func (s *MemoryStore) AdminPermissions() []domain.AdminPermission {
	return []domain.AdminPermission{
		{Code: "content", Module: "内容管理", Name: "帖子 / 评论 / 标签 / 文档", Ops: []string{"查", "增", "删", "改", "审核"}},
		{Code: "plugin", Module: "系统插件", Name: "qa / docs / wiki 插件", Ops: []string{"查", "启用", "禁用"}},
		{Code: "qa", Module: "问答插件", Name: "问题 / 回答 / 采纳", Ops: []string{"查", "增", "审核", "采纳"}},
		{Code: "docs", Module: "文档插件", Name: "空间 / 文档树 / 文档", Ops: []string{"查", "增", "改", "管理"}},
		{Code: "wiki", Module: "Wiki 插件", Name: "页面 / 版本 / 回滚", Ops: []string{"查", "增", "改", "回滚"}},
		{Code: "site", Module: "站点配置", Name: "子站 / 板块 / 搜索范围", Ops: []string{"查", "增", "删", "改"}},
		{Code: "user", Module: "用户管理", Name: "用户信息 / 行为 / 违规处理", Ops: []string{"查", "改", "审核"}},
		{Code: "operation", Module: "运营管理", Name: "推荐 / 通知 / 热门 / 草稿箱", Ops: []string{"查", "增", "删", "改"}},
		{Code: "moderator", Module: "版主管理", Name: "子站版主分配 / 停用", Ops: []string{"查", "增", "删", "改"}},
		{Code: "statistics", Module: "数据统计", Name: "内容 / 用户 / 搜索统计", Ops: []string{"查", "导出"}},
		{Code: "system", Module: "系统设置", Name: "参数 / 日志 / 备份恢复", Ops: []string{"查", "改", "删"}},
	}
}

// AdminComments 返回后台评论审核列表。
func (s *MemoryStore) AdminComments(site string) []domain.AdminComment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AdminComment, 0, len(s.comments))
	for _, c := range s.comments {
		title := ""
		if p, ok := s.posts[c.PostID]; ok {
			if site != "" && site != "portal" && p.Site != site {
				continue
			}
			title = p.Title
		}
		out = append(out, domain.AdminComment{ID: c.ID, PostID: c.PostID, PostTitle: title, ParentID: c.ParentID, Author: c.Author, To: c.To, Text: c.Text, Status: c.Status, Likes: c.Likes, CreatedAt: c.CreatedAt})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

// AdminTopics 返回后台内容列表，包含隐藏内容。
func (s *MemoryStore) AdminTopics(site, board, q string) []domain.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site = strings.TrimSpace(site)
	board = strings.TrimSpace(board)
	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]domain.Post, 0, len(s.posts))
	for _, p := range s.posts {
		if site != "" && site != "portal" && p.Site != site {
			continue
		}
		if board != "" && board != "all" && p.Board != board {
			continue
		}
		if q != "" && !postContains(p, q) {
			continue
		}
		cp := *p
		cp.Tags = append([]string(nil), p.Tags...)
		cp.CommentLocked = s.commentLocks[p.ID]
		if !memoryPostVisible(p) {
			cp.Status = "offline"
		}
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

// UpdateCommentStatus 更新评论审核状态。
func (s *MemoryStore) UpdateCommentStatus(id int64, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.comments[id]
	if !ok {
		return false
	}
	c.Status = strings.TrimSpace(status)
	s.appendLogLocked("audit", "auditor", "更新评论状态", fmt.Sprintf("comments#%d", id), "127.0.0.1")
	return true
}

// AdminSettings 返回后台基础参数。
func (s *MemoryStore) AdminSettings() domain.AdminSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// UpdateAdminSettings 覆盖后台基础参数。
func (s *MemoryStore) UpdateAdminSettings(req domain.AdminSettings) domain.AdminSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = req
	s.appendLogLocked("system", "admin", "更新基础参数", "settings", "127.0.0.1")
	return s.settings
}

// AdminLogs 返回后台操作日志，按 ID 倒序排列。
func (s *MemoryStore) AdminLogs(site string) []domain.AdminLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.AdminLog{}
	for _, log := range s.logs {
		if !logInSite(log, site) {
			continue
		}
		out = append(out, enrichAdminLog(log, s.adminUsersLocked()))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (s *MemoryStore) AdminLogsByFilter(filter domain.AdminLogFilter) ([]domain.AdminLog, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []domain.AdminLog{}
	users := s.adminUsersLocked()
	for _, log := range s.logs {
		log = enrichAdminLog(log, users)
		if !adminLogMatches(log, filter) {
			continue
		}
		items = append(items, log)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	total := len(items)
	page, pageSize := normalizeMemoryPage(filter.Page, filter.PageSize)
	return paginateSlice(items, page, pageSize), total
}

// PushNotification 创建后台推送通知并记录操作日志。
func (s *MemoryStore) PushNotification(req domain.PushNotificationRequest) *domain.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "portal"
	}
	n := s.createNoticeForSiteLocked(scope, strings.TrimSpace(req.Title), strings.TrimSpace(req.Content))
	s.appendLogForSiteLocked(scope, "operation", "operator", "", "发送通知", n.Title, "127.0.0.1")
	return n
}

func (s *MemoryStore) AppendAdminLog(log domain.AdminLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if log.ID == 0 {
		log.ID = s.nextLogID
		s.nextLogID++
	}
	log.Site = normalizeSiteScope(log.Site)
	if log.ActorType == "" {
		if log.Actor == "system" || log.Actor == "" {
			log.ActorType = "system"
		} else {
			log.ActorType = "admin_user"
		}
	}
	if log.ActorID == 0 {
		log.ActorID = log.ActorUserID
	}
	if log.CreatedAt == "" {
		log.CreatedAt = Now()
	}
	s.logs = append(s.logs, log)
}

// limitPosts 按 limit 截断帖子列表。
func limitPosts(posts []domain.Post, limit int) []domain.Post {
	if limit <= 0 || limit > len(posts) {
		limit = len(posts)
	}
	return posts[:limit]
}

// postContains 判断帖子文本字段中是否包含关键词。
func postContains(p *domain.Post, q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	haystack := strings.ToLower(strings.Join([]string{p.Title, p.Summary, p.Author, p.Content, strings.Join(p.Tags, " ")}, " "))
	return strings.Contains(haystack, q)
}

func adminLogMatches(log domain.AdminLog, filter domain.AdminLogFilter) bool {
	if !logInSite(log, filter.Site) {
		return false
	}
	if filter.Type != "" && filter.Type != "all" && log.Type != filter.Type {
		return false
	}
	if filter.ActorType != "" && filter.ActorType != "all" && log.ActorType != filter.ActorType {
		return false
	}
	if filter.Action != "" && !strings.Contains(strings.ToLower(log.Action), strings.ToLower(filter.Action)) {
		return false
	}
	if filter.TargetType != "" && filter.TargetType != "all" && !strings.EqualFold(log.TargetType, filter.TargetType) {
		return false
	}
	if filter.TargetID > 0 && log.TargetID != filter.TargetID {
		return false
	}
	if filter.PluginCode != "" && !adminLogContainsPluginCode(log, filter.PluginCode) {
		return false
	}
	if filter.ActorID > 0 && log.ActorUserID != filter.ActorID {
		return false
	}
	if filter.CommunityID > 0 {
		site := fallbackSiteByCommunityID(filter.CommunityID)
		if site == "" || (!logInSite(log, site) && !logInSite(log, fmt.Sprintf("community:%d", filter.CommunityID))) {
			return false
		}
	}
	if filter.Actor != "" && !strings.Contains(strings.ToLower(log.Actor), strings.ToLower(filter.Actor)) {
		return false
	}
	if filter.Target != "" && !strings.Contains(strings.ToLower(log.Target), strings.ToLower(filter.Target)) {
		return false
	}
	if filter.Metadata != "" && !strings.Contains(strings.ToLower(log.Metadata), strings.ToLower(filter.Metadata)) {
		return false
	}
	if filter.RequestID != "" && !strings.Contains(strings.ToLower(log.Metadata), strings.ToLower(filter.RequestID)) {
		return false
	}
	if filter.StartTime != "" && log.CreatedAt < filter.StartTime {
		return false
	}
	if filter.EndTime != "" && log.CreatedAt > filter.EndTime {
		return false
	}
	return true
}

func adminLogContainsPluginCode(log domain.AdminLog, pluginCode string) bool {
	pluginCode = strings.ToLower(strings.TrimSpace(pluginCode))
	if pluginCode == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		log.Target,
		log.Metadata,
		log.OldValue,
		log.NewValue,
	}, " "))
	return strings.Contains(haystack, pluginCode)
}

// hasTag 判断标签列表是否包含指定标签，大小写不敏感。
func hasTag(tags []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, t := range tags {
		name := strings.TrimSpace(t)
		if strings.ToLower(name) == want || normalizeSlug(name) == want {
			return true
		}
	}
	return false
}

func hasTagOrSlug(tags []string, wantName, wantSlug string) bool {
	wantName = strings.ToLower(strings.TrimSpace(wantName))
	wantSlug = strings.ToLower(strings.TrimSpace(wantSlug))
	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		nameLower := strings.ToLower(name)
		slug := normalizeSlug(name)
		if slug == "" {
			slug = strings.ToLower(strings.Join(strings.Fields(name), "-"))
		}
		if (wantName != "" && nameLower == wantName) || (wantSlug != "" && slug == wantSlug) {
			return true
		}
	}
	return false
}

func memoryTopicHasTag(topic domain.Topic, tag string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if tag == "" {
		return true
	}
	for _, name := range topic.Tags {
		name = strings.TrimSpace(name)
		slug := strings.ToLower(strings.Join(strings.Fields(name), "-"))
		if strings.ToLower(name) == tag || slug == tag {
			return true
		}
	}
	return false
}

func memoryTopicHasTagID(topic domain.Topic, tagID int64) bool {
	if tagID <= 0 {
		return true
	}
	for index := range topic.Tags {
		if int64(index+1) == tagID {
			return true
		}
	}
	return false
}

func memoryTopicMatchesKeyword(topic domain.Topic, keyword string, s *MemoryStore) bool {
	needle := strings.ToLower(strings.TrimSpace(keyword))
	if needle == "" {
		return true
	}
	communityName := fallbackSiteByCommunityID(topic.CommunityID)
	if site, ok := s.sites[communityName]; ok {
		communityName += " " + site.Name
	}
	categoryName := ""
	for _, board := range s.boards {
		if categoryIDForBoard(topic.CommunityID, board.Key) == topic.CategoryID {
			categoryName = board.Name + " " + board.Key
			break
		}
	}
	haystack := strings.ToLower(strings.Join([]string{
		topic.Title,
		topic.Summary,
		topic.Content,
		strings.Join(topic.Tags, " "),
		communityName,
		categoryName,
	}, " "))
	return strings.Contains(haystack, needle)
}

func sortTopicsForSearch(topics []domain.Topic, sortBy string) {
	switch sortBy {
	case "active":
		sort.Slice(topics, func(i, j int) bool {
			left := topics[i].LastActiveAt
			if left == "" {
				left = topics[i].UpdatedAt
			}
			right := topics[j].LastActiveAt
			if right == "" {
				right = topics[j].UpdatedAt
			}
			if left == right {
				return topics[i].ID > topics[j].ID
			}
			return left > right
		})
	case "hot":
		sort.Slice(topics, func(i, j int) bool {
			if topics[i].HotScore == topics[j].HotScore {
				return topics[i].CreatedAt > topics[j].CreatedAt
			}
			return topics[i].HotScore > topics[j].HotScore
		})
	case "featured":
		sort.Slice(topics, func(i, j int) bool {
			if topics[i].UpdatedAt == topics[j].UpdatedAt {
				return topics[i].ID > topics[j].ID
			}
			return topics[i].UpdatedAt > topics[j].UpdatedAt
		})
	case "unsolved":
		sort.Slice(topics, func(i, j int) bool {
			left := firstNonEmptyString(topics[i].LastActiveAt, topics[i].UpdatedAt, topics[i].CreatedAt)
			right := firstNonEmptyString(topics[j].LastActiveAt, topics[j].UpdatedAt, topics[j].CreatedAt)
			if left == right {
				return topics[i].ID > topics[j].ID
			}
			return left > right
		})
	default:
		sort.Slice(topics, func(i, j int) bool {
			if topics[i].IsPinned != topics[j].IsPinned {
				return topics[i].IsPinned
			}
			if topics[i].CreatedAt == topics[j].CreatedAt {
				return topics[i].ID > topics[j].ID
			}
			return topics[i].CreatedAt > topics[j].CreatedAt
		})
	}
}

func communityIDBySite(site string) int64 {
	switch site {
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

func fallbackCommunityIDBySite(site string) int64 {
	return communityIDBySite(site)
}

func categoryIDForBoard(communityID int64, board string) int64 {
	order := map[string]int64{
		"community":  1,
		"qa":         2,
		"opensource": 3,
		"ai":         4,
		"jobs":       5,
		"wiki":       6,
		"docs":       7,
	}
	if communityID <= 0 {
		return order[board]
	}
	return communityID*100 + order[board]
}

func boardByCategoryID(categoryID int64) string {
	order := map[int64]string{
		1: "community",
		2: "qa",
		3: "opensource",
		4: "ai",
		5: "jobs",
		6: "wiki",
		7: "docs",
	}
	if categoryID > 100 {
		categoryID = categoryID % 100
	}
	if board := order[categoryID]; board != "" {
		return board
	}
	return "community"
}

func siteByCommunityID(communityID int64) string {
	switch communityID {
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
		return ""
	}
}

func fallbackSiteByCommunityID(communityID int64) string {
	return siteByCommunityID(communityID)
}

func contentTypeForBoard(board string) string {
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

func boardByContentType(contentType string) string {
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

func paginateTopics(topics []domain.Topic, page, pageSize int) ([]domain.Topic, int) {
	total := len(topics)
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []domain.Topic{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return topics[start:end], total
}

func paginateSlice[T any](items []T, page, pageSize int) []T {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
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

// uniqueTags 去除空标签和重复标签，保留首次出现的原始写法。
func uniqueTags(tags []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, t)
	}
	return out
}

func reactionKey(userID int64, targetType string, targetID int64, reactionType string) string {
	return fmt.Sprintf("%d:%s:%d:%s", userID, strings.TrimSpace(targetType), targetID, strings.TrimSpace(reactionType))
}

func favoriteKey(userID int64, targetType string, targetID int64) string {
	return fmt.Sprintf("%d:%s:%d", userID, strings.TrimSpace(targetType), targetID)
}

func followKey(userID int64, targetType string, targetID int64) string {
	return fmt.Sprintf("%d:%s:%d", userID, strings.TrimSpace(targetType), targetID)
}

func normalizeOptionalTargetType(targetType string, fallback string) string {
	targetType = strings.TrimSpace(targetType)
	if targetType == "" || targetType == "all" {
		return fallback
	}
	return targetType
}

func targetURLFor(targetType string, targetID int64, topicID int64) string {
	switch targetType {
	case "topic":
		if targetID > 0 {
			return fmt.Sprintf("/topics/%d/", targetID)
		}
	case "comment":
		if topicID > 0 {
			return fmt.Sprintf("/topics/%d/#comments", topicID)
		}
	case "community":
		if slug := fallbackSiteByCommunityID(targetID); slug != "" {
			return "/c/" + slug + "/"
		}
	case "tag":
		return fmt.Sprintf("/tags/%d/", targetID)
	case "user":
		return "/me/activities"
	}
	if topicID > 0 {
		return fmt.Sprintf("/topics/%d/", topicID)
	}
	return "/"
}

func (s *MemoryStore) favoriteCountLocked(targetType string, targetID int64) int {
	count := 0
	for _, fav := range s.favorites {
		if fav.TargetType == targetType && fav.TargetID == targetID {
			count++
		}
	}
	return count
}

func (s *MemoryStore) followCountLocked(targetType string, targetID int64) int {
	count := 0
	for _, follow := range s.follows {
		if follow.TargetType == targetType && follow.TargetID == targetID {
			count++
		}
	}
	return count
}

func (s *MemoryStore) reactionExistsLocked(userID int64, targetType string, targetID int64, reactionType string) bool {
	_, ok := s.reactions[reactionKey(userID, targetType, targetID, reactionType)]
	return ok
}

func (s *MemoryStore) favoriteExistsLocked(userID int64, targetType string, targetID int64) bool {
	_, ok := s.favorites[favoriteKey(userID, targetType, targetID)]
	return ok
}

func tagKey(site, slug string) string {
	return strings.TrimSpace(site) + ":" + strings.TrimSpace(slug)
}

func tagURLForSlug(slug, communitySlug string) string {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "/tags/"
	}
	if communitySlug != "" && communitySlug != "portal" {
		return "/c/" + communitySlug + "/tags/" + slug + "/"
	}
	return "/tags/" + slug + "/"
}

func (s *MemoryStore) findTagBySiteSlugLocked(site, slug string) (*domain.Tag, bool) {
	site = strings.TrimSpace(site)
	slug = strings.TrimSpace(slug)
	for _, tag := range s.tags {
		if tag != nil && tag.Site == site && tag.Slug == slug {
			return tag, true
		}
	}
	return nil, false
}

func (s *MemoryStore) findTagAliasBySiteSlugLocked(site, aliasSlug string) (*domain.TagAlias, bool) {
	site = strings.TrimSpace(site)
	aliasSlug = strings.TrimSpace(aliasSlug)
	for _, item := range s.tagAliases {
		if item != nil && item.Site == site && item.AliasSlug == aliasSlug {
			return item, true
		}
	}
	return nil, false
}

func (s *MemoryStore) tagAliasListLocked(tagID int64) []domain.TagAlias {
	items := make([]domain.TagAlias, 0)
	for _, item := range s.tagAliases {
		if item == nil || item.TagID != tagID {
			continue
		}
		cp := *item
		cp.CommunityID = s.communityIDBySlugLocked(cp.Site)
		cp.CommunitySlug = cp.Site
		items = append(items, cp)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].AliasSlug == items[j].AliasSlug {
			return items[i].ID < items[j].ID
		}
		return items[i].AliasSlug < items[j].AliasSlug
	})
	return items
}

func (s *MemoryStore) findMergeTargetLocked(tag *domain.Tag) (*domain.Tag, bool) {
	if tag == nil || tag.MergedToID <= 0 {
		return nil, false
	}
	target, ok := s.tags[tag.MergedToID]
	if !ok || target == nil {
		return nil, false
	}
	return target, true
}

func (s *MemoryStore) resolveTagLocked(site, slugOrName string, enabledOnly bool) (domain.TagResolveResult, bool) {
	site = strings.TrimSpace(site)
	slugOrName = strings.TrimSpace(slugOrName)
	if slugOrName == "" {
		return domain.TagResolveResult{}, false
	}
	normalizedSlug := normalizeSlug(slugOrName)
	if normalizedSlug == "" {
		normalizedSlug = strings.ToLower(strings.Join(strings.Fields(slugOrName), "-"))
	}
	match := func(tag *domain.Tag) bool {
		if tag == nil {
			return false
		}
		if site != "" && site != "portal" && tag.Site != site {
			return false
		}
		return strings.EqualFold(tag.Slug, slugOrName) || strings.EqualFold(tag.Name, slugOrName) || (normalizedSlug != "" && strings.EqualFold(tag.Slug, normalizedSlug))
	}
	var directPortal *domain.Tag
	var directFallback *domain.Tag
	for _, tag := range s.tags {
		if !match(tag) {
			continue
		}
		if site == "" || site == "portal" {
			if tag.Site == "portal" {
				directPortal = tag
				break
			}
			if directFallback == nil {
				directFallback = tag
			}
			continue
		}
		directFallback = tag
		break
	}
	if directPortal != nil || directFallback != nil {
		current := directPortal
		if current == nil {
			current = directFallback
		}
		if current != nil {
			if current.Status == "merged" {
				target, ok := s.findMergeTargetLocked(current)
				if !ok {
					return domain.TagResolveResult{}, false
				}
				if enabledOnly && target.Status != "enable" {
					return domain.TagResolveResult{}, false
				}
				resolved := s.enrichTagLocked(*target)
				return domain.TagResolveResult{Tag: resolved, Requested: slugOrName, ResolvedBy: "merged"}, true
			}
			if enabledOnly && current.Status != "enable" {
				return domain.TagResolveResult{}, false
			}
			return domain.TagResolveResult{Tag: s.enrichTagLocked(*current), Requested: slugOrName, ResolvedBy: "direct"}, true
		}
	}
	for _, item := range s.tagAliases {
		if item == nil {
			continue
		}
		if site != "" && site != "portal" && item.Site != site {
			continue
		}
		if !strings.EqualFold(item.Alias, slugOrName) && !strings.EqualFold(item.AliasSlug, slugOrName) && (normalizedSlug == "" || !strings.EqualFold(item.AliasSlug, normalizedSlug)) {
			continue
		}
		tag, ok := s.tags[item.TagID]
		if !ok || tag == nil {
			continue
		}
		if tag.Status == "merged" {
			target, ok := s.findMergeTargetLocked(tag)
			if !ok {
				return domain.TagResolveResult{}, false
			}
			if enabledOnly && target.Status != "enable" {
				return domain.TagResolveResult{}, false
			}
			return domain.TagResolveResult{
				Tag:          s.enrichTagLocked(*target),
				MatchedAlias: item.Alias,
				Requested:    slugOrName,
				ResolvedBy:   "alias",
			}, true
		}
		if enabledOnly && tag.Status != "enable" {
			return domain.TagResolveResult{}, false
		}
		return domain.TagResolveResult{
			Tag:          s.enrichTagLocked(*tag),
			MatchedAlias: item.Alias,
			Requested:    slugOrName,
			ResolvedBy:   "alias",
		}, true
	}
	return domain.TagResolveResult{}, false
}

func (s *MemoryStore) tagBySlugLocked(site, slugOrName string, enabledOnly bool) (domain.Tag, bool) {
	result, ok := s.resolveTagLocked(site, slugOrName, enabledOnly)
	if !ok {
		return domain.Tag{}, false
	}
	return result.Tag, true
}

func (s *MemoryStore) enrichTagLocked(tag domain.Tag) domain.Tag {
	if tag.SortOrder == 0 {
		tag.SortOrder = tag.Sort
	}
	tag.CommunitySlug = tag.Site
	tag.CommunityID = s.communityIDBySlugLocked(tag.Site)
	if comm := s.communityByIDLocked(tag.CommunityID); comm.ID > 0 {
		tag.CommunityName = comm.Name
	}
	tag.TopicCount = s.tagTopicCountLocked(tag)
	tag.UseCount = tag.TopicCount
	tag.FollowerCount = s.followCountLocked("tag", tag.ID)
	tag.HotScore = tag.TopicCount*10 + tag.FollowerCount*20
	if target, ok := s.findMergeTargetLocked(&tag); ok {
		tag.MergedToName = target.Name
		tag.MergedToSlug = target.Slug
	}
	return tag
}

func (s *MemoryStore) tagTopicCountLocked(tag domain.Tag) int {
	count := 0
	for _, post := range s.posts {
		if !memoryPostVisible(post) {
			continue
		}
		if tag.Site != "" && tag.Site != "portal" && post.Site != tag.Site {
			continue
		}
		if hasTagOrSlug(post.Tags, tag.Name, tag.Slug) {
			count++
		}
	}
	return count
}

func (s *MemoryStore) tagStatsLocked(site, q, status string, limit int) []domain.TagStat {
	q = strings.ToLower(strings.TrimSpace(q))
	status = strings.TrimSpace(status)
	items := make([]domain.TagStat, 0, len(s.tags))
	for _, tag := range s.tags {
		if tag == nil {
			continue
		}
		cp := s.enrichTagLocked(*tag)
		if site != "" && site != "portal" && cp.Site != site && cp.CommunitySlug != site {
			continue
		}
		if status != "" && status != "all" && cp.Status != status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(cp.Name+" "+cp.Slug+" "+cp.Description), q) {
			continue
		}
		if limit == 0 && cp.TopicCount <= 0 && cp.UseCount <= 0 {
			continue
		}
		items = append(items, domain.TagStat{
			ID:             cp.ID,
			Name:           cp.Name,
			Slug:           cp.Slug,
			Site:           cp.Site,
			CommunityID:    cp.CommunityID,
			CommunitySlug:  cp.CommunitySlug,
			Description:    cp.Description,
			TopicCount:     cp.TopicCount,
			Count:          cp.TopicCount,
			FollowerCount:  cp.FollowerCount,
			Status:         cp.Status,
			SEOTitle:       cp.SEOTitle,
			SEODescription: cp.SEODescription,
			SEOKeywords:    cp.SEOKeywords,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].TopicCount == items[j].TopicCount {
			if items[i].CommunitySlug == items[j].CommunitySlug {
				return items[i].Name < items[j].Name
			}
			return items[i].CommunitySlug < items[j].CommunitySlug
		}
		return items[i].TopicCount > items[j].TopicCount
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *MemoryStore) followExistsLocked(userID int64, targetType string, targetID int64) bool {
	_, ok := s.follows[followKey(userID, targetType, targetID)]
	return ok
}

func (s *MemoryStore) topicFromPostLocked(id int64, increaseView bool) (domain.Topic, bool) {
	p, ok := s.posts[id]
	if !ok {
		return domain.Topic{}, false
	}
	status := memoryTopicStatus(p)
	if increaseView {
		p.Views++
	}
	communityID := s.communityIDBySlugLocked(p.Site)
	userID := p.UserID
	if userID <= 0 {
		userID = 1
	}
	favoriteCount := s.favoriteCountLocked("topic", p.ID)
	return domain.Topic{
		ID:            p.ID,
		CommunityID:   communityID,
		CategoryID:    s.categoryIDForBoardLocked(communityID, p.Board),
		UserID:        userID,
		Title:         p.Title,
		ContentType:   contentTypeForBoard(p.Board),
		PluginCode:    pluginregistry.PluginCodeForContentType(contentTypeForBoard(p.Board)),
		Summary:       p.Summary,
		Content:       p.Content,
		Status:        status,
		IsPinned:      p.Pinned,
		IsFeatured:    p.Recommended,
		IsSolved:      s.topicIsSolvedLocked(p),
		CommentLocked: s.commentLocks[p.ID],
		BestCommentID: s.bestCommentIDLocked(p.ID),
		ViewCount:     p.Views,
		CommentCount:  p.Comments,
		LikeCount:     p.Likes,
		FavoriteCount: favoriteCount,
		HotScore:      memoryHotScoreWithFavorites(p, favoriteCount),
		LastActiveAt:  p.UpdatedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Tags:          append([]string{}, p.Tags...),
	}, true
}

func (s *MemoryStore) topicIsSolvedLocked(p *domain.Post) bool {
	if p == nil || contentTypeForBoard(p.Board) != "question" {
		return false
	}
	return s.bestCommentIDLocked(p.ID) > 0
}

func (s *MemoryStore) bestCommentIDLocked(topicID int64) int64 {
	for _, c := range s.comments {
		if c.PostID == topicID && c.IsBest && c.Status != "deleted" && c.Status != "hidden" {
			return c.ID
		}
	}
	return 0
}

func (s *MemoryStore) normalizeCommentLocked(c *domain.Comment) {
	if c == nil {
		return
	}
	c.TopicID = c.PostID
	if c.UserID == 0 {
		c.UserID = 1
	}
	if c.UserName == "" {
		c.UserName = firstNonEmptyString(c.Author, "DevHub 用户")
	}
	if c.Author == "" {
		c.Author = c.UserName
	}
	if c.Content == "" {
		c.Content = c.Text
	}
	if c.Text == "" {
		c.Text = c.Content
	}
	c.LikeCount = c.Likes
	if c.UpdatedAt == "" {
		c.UpdatedAt = c.CreatedAt
	}
}

func (s *MemoryStore) communityByIDLocked(id int64) domain.Community {
	if comm, ok := s.communities[id]; ok {
		cp := *comm
		s.applyCommunityStatsLocked(&cp)
		return cp
	}
	return domain.Community{}
}

func (s *MemoryStore) categoryByIDLocked(communityID, categoryID int64) domain.Category {
	if cat, ok := s.categories[categoryID]; ok && (communityID == 0 || cat.CommunityID == communityID) {
		cp := *cat
		if cp.ContentType == "" {
			cp.ContentType = cp.Type
		}
		return cp
	}
	return domain.Category{}
}

func (s *MemoryStore) categoryIDForBoardLocked(communityID int64, board string) int64 {
	board = strings.TrimSpace(board)
	for _, cat := range s.categories {
		if cat.CommunityID == communityID && (cat.Slug == board || cat.Type == contentTypeForBoard(board) || cat.ContentType == contentTypeForBoard(board)) {
			return cat.ID
		}
	}
	return categoryIDForBoard(communityID, board)
}

func (s *MemoryStore) boardByCategoryIDLocked(categoryID int64) string {
	if cat, ok := s.categories[categoryID]; ok {
		if cat.Slug != "" {
			return cat.Slug
		}
		return boardByContentType(firstNonEmptyString(cat.ContentType, cat.Type))
	}
	return boardByCategoryID(categoryID)
}

func (s *MemoryStore) validateFollowTargetLocked(targetType string, targetID int64) error {
	if targetID <= 0 {
		return errors.New("关注对象 ID 不合法")
	}
	switch targetType {
	case "user":
		if _, ok := s.users[targetID]; ok {
			return nil
		}
		return errors.New("用户不存在")
	case "community":
		if _, ok := s.communities[targetID]; ok {
			return nil
		}
		return errors.New("子站不存在")
	case "tag":
		if tag, ok := s.tags[targetID]; ok && tag != nil && tag.Status == "enable" {
			return nil
		}
		return errors.New("标签不存在或已禁用")
	case "topic":
		if _, ok := s.posts[targetID]; ok {
			return nil
		}
		return errors.New("主题不存在")
	default:
		return errors.New("不支持的关注对象")
	}
}

func (s *MemoryStore) followTargetNameLocked(targetType string, targetID int64) string {
	switch targetType {
	case "user":
		if user, ok := s.users[targetID]; ok {
			if user.Nickname != "" {
				return user.Nickname
			}
			return user.Username
		}
	case "community":
		if comm := s.communityByIDLocked(targetID); comm.Name != "" {
			return comm.Name
		}
	case "tag":
		if tag, ok := s.tags[targetID]; ok && tag != nil {
			return tag.Name
		}
		return fmt.Sprintf("标签 #%d", targetID)
	case "topic":
		if p, ok := s.posts[targetID]; ok {
			return p.Title
		}
	}
	return fmt.Sprintf("%s#%d", targetType, targetID)
}

func (s *MemoryStore) followItemLocked(follow *domain.Follow) domain.FollowItem {
	item := domain.FollowItem{
		ID:         follow.ID,
		UserID:     follow.UserID,
		TargetType: follow.TargetType,
		TargetID:   follow.TargetID,
		TargetName: s.followTargetNameLocked(follow.TargetType, follow.TargetID),
		CreatedAt:  follow.CreatedAt,
		TargetURL:  targetURLFor(follow.TargetType, follow.TargetID, follow.TargetID),
	}
	switch follow.TargetType {
	case "community":
		item.Community = s.communityByIDLocked(follow.TargetID)
		item.TargetSlug = item.Community.Slug
		item.Description = item.Community.Description
		item.TargetURL = "/c/" + item.Community.Slug + "/"
	case "topic":
		if topic, ok := s.topicFromPostLocked(follow.TargetID, false); ok {
			item.Topic = topic
			item.TargetTitle = topic.Title
			item.TargetURL = fmt.Sprintf("/topics/%d/", topic.ID)
			item.Community = s.communityByIDLocked(topic.CommunityID)
		}
	case "tag":
		if tag, ok := s.tags[follow.TargetID]; ok && tag != nil {
			cp := s.enrichTagLocked(*tag)
			item.TargetName = cp.Name
			item.TargetSlug = cp.Slug
			item.Description = cp.Description
			item.TargetURL = tagURLForSlug(cp.Slug, cp.CommunitySlug)
			item.Community = s.communityByIDLocked(cp.CommunityID)
		} else {
			item.TargetSlug = fmt.Sprintf("%d", follow.TargetID)
			item.TargetURL = fmt.Sprintf("/tags/%d/", follow.TargetID)
		}
	case "user":
		if user, ok := s.users[follow.TargetID]; ok {
			item.TargetName = firstNonEmptyString(user.Nickname, user.Username)
			item.Description = user.RoleName
		}
	}
	return item
}

func (s *MemoryStore) enrichActivityLocked(a *domain.Activity) {
	if a == nil {
		return
	}
	if a.TargetType == "topic" {
		if topic, ok := s.topicFromPostLocked(a.TargetID, false); ok {
			a.TopicID = topic.ID
			a.TargetTitle = topic.Title
			a.TargetURL = fmt.Sprintf("/topics/%d/", topic.ID)
			if comm := s.communityByIDLocked(topic.CommunityID); comm.Name != "" {
				a.Community = comm.Name
				a.CommunityID = topic.CommunityID
			}
			if a.Remark == "" {
				a.Remark = topic.Title
			}
		}
		return
	}
	a.TargetTitle = s.followTargetNameLocked(a.TargetType, a.TargetID)
	a.TargetURL = targetURLFor(a.TargetType, a.TargetID, a.TopicID)
	if a.TargetType == "community" {
		if comm := s.communityByIDLocked(a.TargetID); comm.Name != "" {
			a.Community = comm.Name
		}
	}
	if a.TargetType == "tag" {
		if tag, ok := s.tags[a.TargetID]; ok && tag != nil {
			cp := s.enrichTagLocked(*tag)
			a.TargetTitle = cp.Name
			a.TargetURL = tagURLForSlug(cp.Slug, cp.CommunitySlug)
			if comm := s.communityByIDLocked(cp.CommunityID); comm.Name != "" {
				a.Community = comm.Name
				a.CommunityID = comm.ID
			}
		}
	}
	if a.Remark == "" {
		a.Remark = a.TargetTitle
	}
}

func normalizeTag(tag domain.Tag) domain.Tag {
	tag.Site = strings.TrimSpace(tag.Site)
	tag.Name = strings.TrimSpace(tag.Name)
	tag.Slug = strings.TrimSpace(tag.Slug)
	tag.Description = strings.TrimSpace(tag.Description)
	tag.Status = strings.TrimSpace(tag.Status)
	tag.SEOTitle = strings.TrimSpace(tag.SEOTitle)
	tag.SEODescription = strings.TrimSpace(tag.SEODescription)
	tag.SEOKeywords = strings.TrimSpace(tag.SEOKeywords)
	if tag.Status == "" {
		tag.Status = "enable"
	}
	if tag.Status == "1" || tag.Status == "enabled" {
		tag.Status = "enable"
	}
	if tag.Status == "0" || tag.Status == "disabled" {
		tag.Status = "disable"
	}
	if tag.Status == "2" || tag.Status == "merged" {
		tag.Status = "merged"
	}
	if tag.Slug == "" {
		tag.Slug = normalizeSlug(tag.Name)
		if tag.Slug == "" {
			tag.Slug = strings.ToLower(strings.Join(strings.Fields(tag.Name), "-"))
		}
	}
	if tag.Site == "" {
		tag.Site = "portal"
	}
	if tag.SortOrder != 0 && tag.Sort == 0 {
		tag.Sort = tag.SortOrder
	}
	if tag.SortOrder == 0 {
		tag.SortOrder = tag.Sort
	}
	return tag
}

func normalizeTagStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "enable", "enabled", "1", "true":
		return "enable"
	case "disable", "disabled", "0", "false":
		return "disable"
	case "merged", "2":
		return "merged"
	default:
		return ""
	}
}

func memoryTagToStat(tag domain.Tag) domain.TagStat {
	return domain.TagStat{
		ID:             tag.ID,
		Name:           tag.Name,
		Slug:           tag.Slug,
		Site:           tag.Site,
		CommunityID:    tag.CommunityID,
		CommunitySlug:  tag.CommunitySlug,
		Description:    tag.Description,
		TopicCount:     tag.TopicCount,
		Count:          firstNonZero(tag.TopicCount, tag.UseCount),
		FollowerCount:  tag.FollowerCount,
		HotScore:       tag.HotScore,
		Status:         tag.Status,
		SEOTitle:       tag.SEOTitle,
		SEODescription: tag.SEODescription,
		SEOKeywords:    tag.SEOKeywords,
	}
}

func sortTags(tags []domain.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		leftSort := tags[i].Sort
		if leftSort == 0 {
			leftSort = tags[i].SortOrder
		}
		rightSort := tags[j].Sort
		if rightSort == 0 {
			rightSort = tags[j].SortOrder
		}
		if leftSort == rightSort {
			if tags[i].UseCount == tags[j].UseCount {
				if tags[i].Site == tags[j].Site {
					return tags[i].Name < tags[j].Name
				}
				return tags[i].Site < tags[j].Site
			}
			return tags[i].UseCount > tags[j].UseCount
		}
		return leftSort < rightSort
	})
}

// firstRunes 按 rune 截取摘要，避免中文被截断成非法字符。
func firstRunes(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "..."
}

func normalizeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", ",", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
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
		} else {
			lastDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}

func validCommunityStatus(status int) bool {
	return status == 0 || status == 1 || status == 2
}

func validCategoryStatus(status int) bool {
	return status == 0 || status == 1
}

func validCategoryContentType(contentType string) bool {
	return pluginregistry.ValidContentType(contentType)
}

// ===== 新增：DevHub 通用社区系统方法 =====

func (s *MemoryStore) Communities() []domain.Community {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Community, 0, len(s.communities))
	for _, comm := range s.communities {
		cp := *comm
		s.applyCommunityStatsLocked(&cp)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SortOrder == out[j].SortOrder {
			return out[i].ID < out[j].ID
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out
}

func (s *MemoryStore) CommunityBySlug(slug string) (domain.Community, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slug = strings.TrimSpace(slug)
	for _, comm := range s.communities {
		if comm.Slug == slug {
			cp := *comm
			s.applyCommunityStatsLocked(&cp)
			return cp, true
		}
	}
	return domain.Community{}, false
}

func (s *MemoryStore) communityIDBySlugLocked(slug string) int64 {
	slug = strings.TrimSpace(slug)
	for _, comm := range s.communities {
		if comm.Slug == slug {
			return comm.ID
		}
	}
	return fallbackCommunityIDBySite(slug)
}

func (s *MemoryStore) communitySlugByIDLocked(id int64) string {
	if comm, ok := s.communities[id]; ok {
		return comm.Slug
	}
	return fallbackSiteByCommunityID(id)
}

func (s *MemoryStore) Categories(communityID int64) []domain.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Category{}
	for _, cat := range s.categories {
		if communityID > 0 && cat.CommunityID != communityID {
			continue
		}
		cp := *cat
		if cp.ContentType == "" {
			cp.ContentType = cp.Type
		}
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SortOrder == out[j].SortOrder {
			return out[i].ID < out[j].ID
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out
}

func (s *MemoryStore) CommunityStats(communityID int64) domain.CommunityStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.communityStatsLocked(communityID)
}

func (s *MemoryStore) CreateCommunity(req domain.CommunityRequest) (domain.Community, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	comm, err := s.normalizeCommunityRequestLocked(req, nil)
	if err != nil {
		return domain.Community{}, err
	}
	for _, existing := range s.communities {
		if existing.Slug == comm.Slug {
			return domain.Community{}, errors.New("子站 slug 已存在")
		}
	}
	comm.ID = s.nextCommunityID
	s.nextCommunityID++
	now := Now()
	comm.CreatedAt = now
	comm.UpdatedAt = now
	s.communities[comm.ID] = comm
	for _, cat := range defaultCategorySeeds(comm.ID) {
		cp := cat
		if cp.ID >= s.nextCategoryID {
			cp.ID = s.nextCategoryID
		}
		s.nextCategoryID = cp.ID + 1
		cp.CommunityID = comm.ID
		s.categories[cp.ID] = &cp
	}
	cp := *comm
	s.applyCommunityStatsLocked(&cp)
	return cp, nil
}

func (s *MemoryStore) UpdateCommunity(id int64, req domain.CommunityRequest) (domain.Community, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.communities[id]
	if !ok {
		return domain.Community{}, errors.New("子站不存在")
	}
	updated, err := s.normalizeCommunityRequestLocked(req, current)
	if err != nil {
		return domain.Community{}, err
	}
	for _, existing := range s.communities {
		if existing.ID != id && existing.Slug == updated.Slug {
			return domain.Community{}, errors.New("子站 slug 已存在")
		}
	}
	updated.ID = id
	updated.CreatedAt = current.CreatedAt
	updated.UpdatedAt = Now()
	s.communities[id] = updated
	cp := *updated
	s.applyCommunityStatsLocked(&cp)
	return cp, nil
}

func (s *MemoryStore) SetCommunityStatus(id int64, status int) (domain.Community, error) {
	if status != 0 && status != 1 && status != 2 {
		return domain.Community{}, errors.New("子站状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	comm, ok := s.communities[id]
	if !ok {
		return domain.Community{}, errors.New("子站不存在")
	}
	comm.Status = status
	comm.UpdatedAt = Now()
	cp := *comm
	s.applyCommunityStatsLocked(&cp)
	return cp, nil
}

func (s *MemoryStore) ReorderCommunities(ids []int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := 0
	for i, id := range ids {
		if comm, ok := s.communities[id]; ok {
			comm.SortOrder = i + 1
			comm.UpdatedAt = Now()
			updated++
		}
	}
	return updated
}

func (s *MemoryStore) CreateCategory(communityID int64, req domain.CategoryRequest) (domain.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.communities[communityID]; !ok {
		return domain.Category{}, errors.New("子站不存在")
	}
	cat, err := s.normalizeCategoryRequestLocked(req, nil)
	if err != nil {
		return domain.Category{}, err
	}
	cat.CommunityID = communityID
	for _, existing := range s.categories {
		if existing.CommunityID == communityID && existing.Slug == cat.Slug {
			return domain.Category{}, errors.New("板块 slug 已存在")
		}
	}
	cat.ID = s.nextCategoryID
	s.nextCategoryID++
	now := Now()
	cat.CreatedAt = now
	cat.UpdatedAt = now
	s.categories[cat.ID] = cat
	return *cat, nil
}

func (s *MemoryStore) UpdateCategory(id int64, req domain.CategoryRequest) (domain.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.categories[id]
	if !ok {
		return domain.Category{}, errors.New("板块不存在")
	}
	updated, err := s.normalizeCategoryRequestLocked(req, current)
	if err != nil {
		return domain.Category{}, err
	}
	for _, existing := range s.categories {
		if existing.ID != id && existing.CommunityID == updated.CommunityID && existing.Slug == updated.Slug {
			return domain.Category{}, errors.New("板块 slug 已存在")
		}
	}
	updated.ID = id
	updated.CreatedAt = current.CreatedAt
	updated.UpdatedAt = Now()
	s.categories[id] = updated
	return *updated, nil
}

func (s *MemoryStore) SetCategoryStatus(id int64, status int) (domain.Category, error) {
	if status != 0 && status != 1 {
		return domain.Category{}, errors.New("板块状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cat, ok := s.categories[id]
	if !ok {
		return domain.Category{}, errors.New("板块不存在")
	}
	cat.Status = status
	cat.Visible = status == 1
	cat.NavVisible = status == 1
	cat.UpdatedAt = Now()
	return *cat, nil
}

func (s *MemoryStore) ReorderCategories(ids []int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := 0
	for i, id := range ids {
		if cat, ok := s.categories[id]; ok {
			cat.SortOrder = i + 1
			cat.UpdatedAt = Now()
			updated++
		}
	}
	return updated
}

func (s *MemoryStore) normalizeCommunityRequestLocked(req domain.CommunityRequest, current *domain.Community) (*domain.Community, error) {
	comm := &domain.Community{Status: 1}
	if current != nil {
		cp := *current
		comm = &cp
	}
	if strings.TrimSpace(req.Name) != "" {
		comm.Name = strings.TrimSpace(req.Name)
	}
	if comm.Name == "" {
		return nil, errors.New("子站名称不能为空")
	}
	slug := normalizeSlug(req.Slug)
	if slug != "" {
		comm.Slug = slug
	} else if comm.Slug == "" {
		comm.Slug = normalizeSlug(comm.Name)
	}
	if comm.Slug == "" {
		return nil, errors.New("子站 slug 不能为空")
	}
	comm.Logo = strings.TrimSpace(req.Logo)
	if comm.Logo == "" && current == nil {
		comm.Logo = strings.ToUpper(firstRunes(comm.Name, 2))
	}
	if strings.TrimSpace(req.CoverImage) != "" || current == nil {
		comm.CoverImage = strings.TrimSpace(req.CoverImage)
	}
	if strings.TrimSpace(req.Slogan) != "" || current == nil {
		comm.Slogan = strings.TrimSpace(req.Slogan)
	}
	if strings.TrimSpace(req.Description) != "" || current == nil {
		comm.Description = strings.TrimSpace(req.Description)
	}
	if strings.TrimSpace(req.ThemeColor) != "" || current == nil {
		comm.ThemeColor = strings.TrimSpace(req.ThemeColor)
	}
	if comm.ThemeColor == "" {
		comm.ThemeColor = "#2563eb"
	}
	comm.SEOTitle = strings.TrimSpace(req.SEOTitle)
	if comm.SEOTitle == "" && current != nil {
		comm.SEOTitle = current.SEOTitle
	}
	if comm.SEOTitle == "" {
		comm.SEOTitle = comm.Name + " 技术社区"
	}
	comm.SEODescription = strings.TrimSpace(req.SEODescription)
	if comm.SEODescription == "" && current != nil {
		comm.SEODescription = current.SEODescription
	}
	if comm.SEODescription == "" {
		comm.SEODescription = comm.Description
	}
	comm.SEOKeywords = strings.TrimSpace(req.SEOKeywords)
	if comm.SEOKeywords == "" && current != nil {
		comm.SEOKeywords = current.SEOKeywords
	}
	if req.SortOrder != nil {
		comm.SortOrder = *req.SortOrder
	} else if current == nil && comm.SortOrder == 0 {
		comm.SortOrder = len(s.communities) + 1
	}
	if req.Status != nil {
		if !validCommunityStatus(*req.Status) {
			return nil, errors.New("子站状态不合法")
		}
		comm.Status = *req.Status
	}
	comm.AnnouncementTitle = strings.TrimSpace(req.AnnouncementTitle)
	if comm.AnnouncementTitle == "" && current != nil {
		comm.AnnouncementTitle = current.AnnouncementTitle
	}
	comm.AnnouncementContent = strings.TrimSpace(req.AnnouncementContent)
	if comm.AnnouncementContent == "" && current != nil {
		comm.AnnouncementContent = current.AnnouncementContent
	}
	comm.AnnouncementURL = strings.TrimSpace(req.AnnouncementURL)
	if comm.AnnouncementURL == "" && current != nil {
		comm.AnnouncementURL = current.AnnouncementURL
	}
	return comm, nil
}

func (s *MemoryStore) normalizeCategoryRequestLocked(req domain.CategoryRequest, current *domain.Category) (*domain.Category, error) {
	cat := &domain.Category{Visible: true, NavVisible: true, Postable: true, Status: 1}
	if current != nil {
		cp := *current
		cat = &cp
	}
	if req.CommunityID > 0 {
		if _, ok := s.communities[req.CommunityID]; !ok {
			return nil, errors.New("子站不存在")
		}
		cat.CommunityID = req.CommunityID
	}
	if strings.TrimSpace(req.Name) != "" {
		cat.Name = strings.TrimSpace(req.Name)
	}
	if cat.Name == "" {
		return nil, errors.New("板块名称不能为空")
	}
	slug := normalizeSlug(req.Slug)
	if slug != "" {
		cat.Slug = slug
	} else if cat.Slug == "" {
		cat.Slug = normalizeSlug(cat.Name)
	}
	if cat.Slug == "" {
		return nil, errors.New("板块 slug 不能为空")
	}
	contentType := strings.TrimSpace(firstNonEmptyString(req.ContentType, req.Type))
	if contentType != "" {
		contentType = pluginregistry.NormalizeContentType(contentType)
		if !validCategoryContentType(contentType) {
			return nil, errors.New("内容类型不合法")
		}
		cat.Type = contentType
		cat.ContentType = contentType
	}
	if cat.Type == "" {
		cat.Type = "article"
		cat.ContentType = "article"
	}
	if cat.ContentType == "" {
		cat.ContentType = cat.Type
	}
	if req.PluginCode != "" {
		cat.PluginCode = strings.TrimSpace(req.PluginCode)
	}
	expectedPlugin := pluginregistry.PluginCodeForContentType(cat.ContentType)
	if cat.PluginCode == "" {
		cat.PluginCode = expectedPlugin
	}
	if cat.PluginCode != expectedPlugin {
		return nil, errors.New("板块插件与内容类型不匹配")
	}
	if cat.CommunityID == 0 {
		return nil, errors.New("板块必须绑定子站")
	}
	if expectedPlugin != pluginregistry.CoreCode {
		global := s.plugins[expectedPlugin]
		if global == nil || global.Status != pluginregistry.StatusEnabled {
			return nil, errors.New("插件全局未启用，不能绑定该插件板块")
		}
		cp := s.communityPlugins[cat.CommunityID][expectedPlugin]
		if cp == nil || cp.Status != pluginregistry.StatusEnabled {
			return nil, errors.New("当前子站未启用该插件，不能绑定该插件板块")
		}
	}
	if len(req.AllowedContentTypes) > 0 {
		allowed := make([]string, 0, len(req.AllowedContentTypes))
		for _, item := range req.AllowedContentTypes {
			item = pluginregistry.NormalizeContentType(item)
			if !validCategoryContentType(item) {
				return nil, errors.New("允许内容类型不合法")
			}
			allowed = append(allowed, item)
		}
		cat.AllowedContentTypes = uniqueTags(allowed)
	}
	if len(cat.AllowedContentTypes) == 0 {
		cat.AllowedContentTypes = pluginregistry.DefaultAllowedContentTypes(cat.ContentType)
	}
	if strings.TrimSpace(req.Description) != "" || current == nil {
		cat.Description = strings.TrimSpace(req.Description)
	}
	if strings.TrimSpace(req.Icon) != "" || current == nil {
		cat.Icon = strings.TrimSpace(req.Icon)
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	} else if current == nil && cat.SortOrder == 0 {
		cat.SortOrder = len(s.categories) + 1
	}
	if req.Visible != nil {
		cat.Visible = *req.Visible
	}
	if req.NavVisible != nil {
		cat.NavVisible = *req.NavVisible
	}
	if req.Postable != nil {
		cat.Postable = *req.Postable
	}
	if req.Status != nil {
		if !validCategoryStatus(*req.Status) {
			return nil, errors.New("板块状态不合法")
		}
		cat.Status = *req.Status
	}
	cat.SEOTitle = strings.TrimSpace(req.SEOTitle)
	if cat.SEOTitle == "" && current != nil {
		cat.SEOTitle = current.SEOTitle
	}
	cat.SEODescription = strings.TrimSpace(req.SEODescription)
	if cat.SEODescription == "" && current != nil {
		cat.SEODescription = current.SEODescription
	}
	return cat, nil
}

func (s *MemoryStore) applyCommunityStatsLocked(comm *domain.Community) {
	if comm == nil {
		return
	}
	stats := s.communityStatsLocked(comm.ID)
	comm.FollowerCount = stats.FollowerCount
	comm.TopicCount = stats.TopicCount
	comm.CommentCount = stats.CommentCount
	comm.HotScore = stats.HotScore
}

func (s *MemoryStore) communityStatsLocked(communityID int64) domain.CommunityStats {
	stats := domain.CommunityStats{}
	slug := s.communitySlugByIDLocked(communityID)
	for _, p := range s.posts {
		if p.Site != slug || !memoryPostVisible(p) {
			continue
		}
		stats.TopicCount++
		stats.CommentCount += p.Comments
		stats.HotScore += p.Views + p.Comments*5 + p.Likes*3 + s.favoriteCountLocked("topic", p.ID)*4
		if contentTypeForBoard(p.Board) == "question" {
			stats.QuestionCount++
			if !s.topicIsSolvedLocked(p) {
				stats.UnsolvedCount++
			}
		}
		if strings.HasPrefix(p.CreatedAt, time.Now().Format("2006-01-02")) {
			stats.TodayTopicCount++
		}
	}
	for _, c := range s.comments {
		if c.Status != "normal" {
			continue
		}
		if p, ok := s.posts[c.PostID]; ok && p.Site == slug && strings.HasPrefix(c.CreatedAt, time.Now().Format("2006-01-02")) {
			stats.TodayCommentCount++
		}
	}
	for key := range s.follows {
		parts := strings.Split(key, ":")
		if len(parts) == 3 && parts[1] == "community" {
			if id, _ := strconv.ParseInt(parts[2], 10, 64); id == communityID {
				stats.FollowerCount++
			}
		}
	}
	for _, moderator := range s.moderators {
		if moderator.CommunityID == communityID && moderator.Status == 1 {
			stats.ModeratorCount++
		}
	}
	return stats
}

func (s *MemoryStore) TopicsByFilter(communityID, categoryID int64, contentType, sortBy string, isSolved *bool, tag string, page, pageSize int) ([]domain.Topic, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	contentType = pluginregistry.NormalizeContentType(contentType)

	// 从 posts 转换为 topics
	topics := []domain.Topic{}
	for _, p := range s.posts {
		if !memoryPostVisible(p) {
			continue
		}
		favoriteCount := s.favoriteCountLocked("topic", p.ID)
		topic := domain.Topic{
			ID:            p.ID,
			CommunityID:   s.communityIDBySlugLocked(p.Site),
			CategoryID:    s.categoryIDForBoardLocked(s.communityIDBySlugLocked(p.Site), p.Board),
			UserID:        p.UserID,
			Title:         p.Title,
			ContentType:   contentTypeForBoard(p.Board),
			PluginCode:    pluginregistry.PluginCodeForContentType(contentTypeForBoard(p.Board)),
			Summary:       p.Summary,
			Content:       p.Content,
			Status:        memoryTopicStatus(p),
			IsPinned:      p.Pinned,
			IsFeatured:    p.Recommended,
			IsSolved:      s.topicIsSolvedLocked(p),
			CommentLocked: s.commentLocks[p.ID],
			BestCommentID: s.bestCommentIDLocked(p.ID),
			ViewCount:     p.Views,
			CommentCount:  p.Comments,
			LikeCount:     p.Likes,
			FavoriteCount: favoriteCount,
			HotScore:      memoryHotScoreWithFavorites(p, favoriteCount),
			LastActiveAt:  p.UpdatedAt,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		}
		if topic.UserID <= 0 {
			topic.UserID = 1
		}
		topic.Tags = p.Tags
		topics = append(topics, topic)
	}

	// 筛选
	filtered := []domain.Topic{}
	for _, t := range topics {
		if communityID > 0 && t.CommunityID != communityID {
			continue
		}
		if categoryID > 0 && t.CategoryID != categoryID {
			continue
		}
		if contentType != "" && contentType != "all" && t.ContentType != contentType {
			continue
		}
		if isSolved != nil && (t.ContentType != "question" || t.IsSolved != *isSolved) {
			continue
		}
		if sortBy == "featured" && !t.IsFeatured {
			continue
		}
		if sortBy == "unsolved" && (t.ContentType != "question" || t.IsSolved) {
			continue
		}
		if tag != "" && !hasTag(t.Tags, tag) {
			continue
		}
		filtered = append(filtered, t)
	}

	// 排序
	sortType := sortBy
	switch sortType {
	case "hot":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].HotScore > filtered[j].HotScore })
	case "featured":
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].IsFeatured != filtered[j].IsFeatured {
				return filtered[i].IsFeatured
			}
			return filtered[i].ID > filtered[j].ID
		})
	case "solved":
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].IsSolved != filtered[j].IsSolved {
				return filtered[i].IsSolved
			}
			return filtered[i].ID > filtered[j].ID
		})
	case "active", "unsolved":
		sort.Slice(filtered, func(i, j int) bool {
			left := firstNonEmptyString(filtered[i].LastActiveAt, filtered[i].UpdatedAt, filtered[i].CreatedAt)
			right := firstNonEmptyString(filtered[j].LastActiveAt, filtered[j].UpdatedAt, filtered[j].CreatedAt)
			if left == right {
				return filtered[i].ID > filtered[j].ID
			}
			return left > right
		})
	default: // latest, active
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID > filtered[j].ID })
	}

	total := len(filtered)
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []domain.Topic{}, total
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total
}

func (s *MemoryStore) TopicByID(id int64, increaseView bool) (*domain.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 从 posts 查找
	for _, p := range s.posts {
		if p.ID == id {
			if increaseView {
				p.Views++
			}

			topic, _ := s.topicFromPostLocked(id, false)
			return &topic, nil
		}
	}
	return nil, errors.New("主题不存在")
}

func (s *MemoryStore) CreateTopic(req domain.CreateTopicRequest) (*domain.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req.ContentType = pluginregistry.NormalizeContentType(req.ContentType)
	req.PluginCode = pluginregistry.PluginCodeForContentType(req.ContentType)
	site := s.communitySlugByIDLocked(req.CommunityID)
	board := boardByContentType(req.ContentType)
	if req.CategoryID > 0 {
		board = s.boardByCategoryIDLocked(req.CategoryID)
	}
	if site == "" {
		return nil, errors.New("子站不存在")
	}
	if board == "" {
		board = "community"
	}
	userID := req.UserID
	if userID <= 0 {
		userID = 1
	}
	s.nextPostID++
	now := Now()
	post := &domain.Post{
		ID:          s.nextPostID,
		UserID:      userID,
		Site:        site,
		Board:       board,
		Title:       strings.TrimSpace(req.Title),
		Summary:     strings.TrimSpace(req.Summary),
		Content:     strings.TrimSpace(req.Content),
		Author:      "DevHub 用户",
		Status:      "publish",
		Pinned:      false,
		Recommended: false,
		Views:       0,
		Likes:       0,
		Comments:    0,
		Tags:        append([]string{}, req.Tags...),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.posts[post.ID] = post
	s.rebuildTagsFromPostsLocked()
	s.appendActivityLocked(userID, req.CommunityID, "created_topic", "topic", post.ID, post.ID, post.Title)
	s.appendLogLocked("operation", "DevHub 用户", "创建主题", fmt.Sprintf("topics#%d", post.ID), "127.0.0.1")

	topic := domain.Topic{
		ID:            post.ID,
		CommunityID:   req.CommunityID,
		CategoryID:    req.CategoryID,
		UserID:        userID,
		Title:         post.Title,
		ContentType:   req.ContentType,
		PluginCode:    req.PluginCode,
		Summary:       post.Summary,
		Content:       post.Content,
		Status:        1,
		IsPinned:      false,
		IsFeatured:    false,
		IsSolved:      false,
		ViewCount:     0,
		CommentCount:  0,
		LikeCount:     0,
		FavoriteCount: 0,
		HotScore:      0,
		LastActiveAt:  now,
		Tags:          post.Tags,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.seedPluginContentRowsLocked(topic)
	return &topic, nil
}

func (s *MemoryStore) UpdateTopic(id int64, req domain.UpdateTopicRequest) (*domain.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("主题不存在")
	}
	if req.CommunitySlug != nil || req.CommunityID != nil {
		communityID := int64(0)
		if req.CommunityID != nil {
			communityID = *req.CommunityID
		}
		if communityID == 0 && req.CommunitySlug != nil {
			communityID = s.communityIDBySlugLocked(strings.TrimSpace(*req.CommunitySlug))
		}
		site := s.communitySlugByIDLocked(communityID)
		if site == "" {
			return nil, errors.New("子站不存在")
		}
		p.Site = site
	}
	if req.CategoryID != nil && *req.CategoryID > 0 {
		p.Board = s.boardByCategoryIDLocked(*req.CategoryID)
	}
	if req.ContentType != nil && strings.TrimSpace(*req.ContentType) != "" {
		contentType := pluginregistry.NormalizeContentType(*req.ContentType)
		p.Board = boardByContentType(contentType)
		*req.ContentType = contentType
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, errors.New("标题不能为空")
		}
		p.Title = title
	}
	if req.Summary != nil {
		p.Summary = strings.TrimSpace(*req.Summary)
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if content == "" {
			return nil, errors.New("正文不能为空")
		}
		p.Content = content
	}
	if req.Status != nil {
		memorySetPostStatus(p, *req.Status)
	}
	if req.IsPinned != nil {
		p.Pinned = *req.IsPinned
	}
	if req.IsFeatured != nil {
		p.Recommended = *req.IsFeatured
	}
	if req.CommentLocked != nil {
		s.commentLocks[id] = *req.CommentLocked
	}
	if req.Tags != nil {
		p.Tags = uniqueTags(*req.Tags)
	}
	p.UpdatedAt = Now()
	if contentTypeForBoard(p.Board) == "question" {
		if q, ok := s.qaQuestions[id]; ok && q != nil {
			q.IsResolved = s.topicIsSolvedLocked(p)
			q.AcceptedAnswerID = s.bestCommentIDLocked(id)
			q.AnswerCount = p.Comments
			q.UpdatedAt = p.UpdatedAt
		}
	}
	if contentTypeForBoard(p.Board) == "document" {
		if doc, ok := s.docsDocuments[id]; ok && doc != nil {
			doc.Version++
			doc.UpdatedAt = p.UpdatedAt
		}
	}
	if contentTypeForBoard(p.Board) == "wiki_page" {
		if page, ok := s.wikiPages[id]; ok && page != nil {
			versionID := int64(len(s.wikiVersions) + 1)
			versionNo := s.countWikiVersionsLocked(page.ID) + 1
			s.wikiVersions[versionID] = &domain.WikiRevision{
				ID:         versionID,
				WikiPageID: page.ID,
				TopicID:    id,
				EditorID:   p.UserID,
				VersionNo:  versionNo,
				Title:      p.Title,
				Content:    p.Content,
				ChangeNote: "更新页面",
				CreatedAt:  p.UpdatedAt,
			}
			page.Title = p.Title
			page.Summary = p.Summary
			page.Content = p.Content
			page.UpdatedAt = p.UpdatedAt
			page.CurrentVersionID = versionID
		}
	}
	s.rebuildTagsFromPostsLocked()
	topic, _ := s.topicFromPostLocked(id, false)
	return &topic, nil
}

func (s *MemoryStore) DeleteTopic(id int64) bool {
	return s.DeletePost(id)
}

func (s *MemoryStore) SearchTopics(req domain.SearchRequest) ([]domain.Topic, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req.ContentType = pluginregistry.NormalizeContentType(req.ContentType)

	all := make([]domain.Topic, 0, len(s.posts))
	for _, p := range s.posts {
		if !memoryPostVisible(p) {
			continue
		}
		status := memoryTopicStatus(p)
		communityID := s.communityIDBySlugLocked(p.Site)
		favoriteCount := s.favoriteCountLocked("topic", p.ID)
		topic := domain.Topic{
			ID:            p.ID,
			CommunityID:   communityID,
			CategoryID:    s.categoryIDForBoardLocked(communityID, p.Board),
			UserID:        p.UserID,
			Title:         p.Title,
			ContentType:   contentTypeForBoard(p.Board),
			PluginCode:    pluginregistry.PluginCodeForContentType(contentTypeForBoard(p.Board)),
			Summary:       p.Summary,
			Content:       p.Content,
			Status:        status,
			IsPinned:      p.Pinned,
			IsFeatured:    p.Recommended,
			IsSolved:      s.topicIsSolvedLocked(p),
			CommentLocked: s.commentLocks[p.ID],
			BestCommentID: s.bestCommentIDLocked(p.ID),
			ViewCount:     p.Views,
			CommentCount:  p.Comments,
			LikeCount:     p.Likes,
			FavoriteCount: favoriteCount,
			HotScore:      memoryHotScoreWithFavorites(p, favoriteCount),
			LastActiveAt:  p.UpdatedAt,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
			Tags:          append([]string{}, p.Tags...),
		}
		all = append(all, topic)
	}

	filtered := make([]domain.Topic, 0, len(all))
	for _, topic := range all {
		if req.CommunityID > 0 && topic.CommunityID != req.CommunityID {
			continue
		}
		if req.CategoryID > 0 && topic.CategoryID != req.CategoryID {
			continue
		}
		if req.ContentType != "" && topic.ContentType != req.ContentType {
			continue
		}
		if req.PluginCode != "" && req.PluginCode != topic.PluginCode {
			continue
		}
		if req.Sort == "featured" && !topic.IsFeatured {
			continue
		}
		if req.Sort == "unsolved" && (topic.ContentType != "question" || topic.IsSolved) {
			continue
		}
		if req.TagID > 0 && !memoryTopicHasTagID(topic, req.TagID) {
			continue
		}
		if req.TagID == 0 && req.Tag != "" && !memoryTopicHasTag(topic, req.Tag) {
			continue
		}
		if req.Keyword != "" && !memoryTopicMatchesKeyword(topic, req.Keyword, s) {
			continue
		}
		filtered = append(filtered, topic)
	}

	sortTopicsForSearch(filtered, req.Sort)
	return paginateTopics(filtered, req.Page, req.PageSize)
}

func (s *MemoryStore) ToggleReaction(userID int64, targetID int64, targetType, reactionType string) (bool, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userID <= 0 {
		userID = 1
	}
	targetType = strings.TrimSpace(targetType)
	reactionType = strings.TrimSpace(reactionType)
	if reactionType == "" {
		reactionType = "like"
	}
	if targetType != "topic" && targetType != "comment" {
		return false, 0, errors.New("不支持的点赞对象")
	}
	if reactionType != "like" {
		return false, 0, errors.New("不支持的互动类型")
	}
	if targetType == "topic" {
		if _, ok := s.posts[targetID]; !ok {
			return false, 0, errors.New("主题不存在")
		}
	}
	if targetType == "comment" {
		if _, ok := s.comments[targetID]; !ok {
			return false, 0, errors.New("评论不存在")
		}
	}

	key := reactionKey(userID, targetType, targetID, reactionType)
	if _, exists := s.reactions[key]; exists {
		delete(s.reactions, key)
		if targetType == "topic" {
			p := s.posts[targetID]
			if p.Likes > 0 {
				p.Likes--
			}
			p.UpdatedAt = Now()
			return false, p.Likes, nil
		}
		if c, ok := s.comments[targetID]; ok {
			if c.Likes > 0 {
				c.Likes--
			}
			return false, c.Likes, nil
		}
		return false, 0, nil
	}

	now := Now()
	s.reactions[key] = &domain.Reaction{
		ID:           s.nextReactionID,
		UserID:       userID,
		TargetType:   targetType,
		TargetID:     targetID,
		ReactionType: reactionType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.nextReactionID++
	if targetType == "topic" {
		p := s.posts[targetID]
		p.Likes++
		p.UpdatedAt = now
		communityID := s.communityIDBySlugLocked(p.Site)
		s.appendActivityLocked(userID, communityID, "liked", "topic", targetID, targetID, p.Title)
		s.createUserNoticeLocked(int64(1), userID, "topic_liked", "topic", targetID, targetID, 0, "你的内容被点赞", fmt.Sprintf("主题《%s》获得了新的点赞。", p.Title))
		return true, p.Likes, nil
	}
	c := s.comments[targetID]
	c.Likes++
	return true, c.Likes, nil
}

func (s *MemoryStore) ToggleFavorite(userID int64, targetID int64, targetType string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userID <= 0 {
		userID = 1
	}
	targetType = strings.TrimSpace(targetType)
	if targetType == "" {
		targetType = "topic"
	}
	if targetType != "topic" {
		return false, errors.New("不支持的收藏对象")
	}
	p, ok := s.posts[targetID]
	if !ok {
		return false, errors.New("主题不存在")
	}

	key := favoriteKey(userID, targetType, targetID)
	if _, exists := s.favorites[key]; exists {
		delete(s.favorites, key)
		p.UpdatedAt = Now()
		return false, nil
	}

	now := Now()
	s.favorites[key] = &domain.Favorite{
		ID:         s.nextFavoriteID,
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.nextFavoriteID++
	communityID := s.communityIDBySlugLocked(p.Site)
	s.appendActivityLocked(userID, communityID, "favorited", "topic", targetID, targetID, p.Title)
	s.createUserNoticeLocked(int64(1), userID, "topic_favorited", "topic", targetID, targetID, 0, "你的内容被收藏", fmt.Sprintf("主题《%s》被收藏。", p.Title))
	p.UpdatedAt = now
	return true, nil
}

func (s *MemoryStore) ToggleFollow(userID int64, targetID int64, targetType string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userID <= 0 {
		userID = 1
	}
	targetType = strings.TrimSpace(targetType)
	if targetType == "" {
		return false, errors.New("关注对象类型不能为空")
	}
	if err := s.validateFollowTargetLocked(targetType, targetID); err != nil {
		return false, err
	}

	key := followKey(userID, targetType, targetID)
	if _, exists := s.follows[key]; exists {
		delete(s.follows, key)
		return false, nil
	}

	now := Now()
	s.follows[key] = &domain.Follow{
		ID:         s.nextFollowID,
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.nextFollowID++
	communityID := int64(0)
	remark := s.followTargetNameLocked(targetType, targetID)
	topicID := int64(0)
	if targetType == "community" {
		communityID = targetID
	}
	if targetType == "tag" {
		if tag, ok := s.tags[targetID]; ok && tag != nil {
			communityID = s.communityIDBySlugLocked(tag.Site)
		}
	}
	if targetType == "topic" {
		if p, ok := s.posts[targetID]; ok {
			communityID = s.communityIDBySlugLocked(p.Site)
			topicID = targetID
		}
	}
	s.appendActivityLocked(userID, communityID, "followed", targetType, targetID, topicID, remark)
	if targetType == "user" {
		s.createUserNoticeLocked(targetID, userID, "user_followed", "user", targetID, 0, 0, "你有新的关注者", "有用户关注了你。")
	}
	return true, nil
}

func (s *MemoryStore) TopicInteraction(userID int64, topicID int64) (domain.TopicInteraction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.posts[topicID]
	if !ok {
		return domain.TopicInteraction{}, errors.New("主题不存在")
	}
	if userID <= 0 {
		userID = 1
	}
	favoriteCount := s.favoriteCountLocked("topic", topicID)
	return domain.TopicInteraction{
		Liked:         s.reactionExistsLocked(userID, "topic", topicID, "like"),
		Favorited:     s.favoriteExistsLocked(userID, "topic", topicID),
		Followed:      s.followExistsLocked(userID, "topic", topicID),
		LikeCount:     p.Likes,
		FavoriteCount: favoriteCount,
		HotScore:      memoryHotScoreWithFavorites(p, favoriteCount),
	}, nil
}

func (s *MemoryStore) UserFavorites(userID int64, targetType string, page, pageSize int) ([]domain.FavoriteItem, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if userID <= 0 {
		userID = 1
	}
	targetType = normalizeOptionalTargetType(targetType, "topic")
	items := make([]domain.FavoriteItem, 0)
	for _, fav := range s.favorites {
		if fav.UserID != userID {
			continue
		}
		if targetType != "" && fav.TargetType != targetType {
			continue
		}
		item := domain.FavoriteItem{
			ID:         fav.ID,
			UserID:     fav.UserID,
			TargetType: fav.TargetType,
			TargetID:   fav.TargetID,
			CreatedAt:  fav.CreatedAt,
			TargetURL:  targetURLFor(fav.TargetType, fav.TargetID, fav.TargetID),
		}
		if fav.TargetType == "topic" {
			if topic, ok := s.topicFromPostLocked(fav.TargetID, false); ok {
				item.Topic = topic
				item.Community = s.communityByIDLocked(topic.CommunityID)
				item.Category = s.categoryByIDLocked(topic.CommunityID, topic.CategoryID)
			}
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt == items[j].CreatedAt {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})
	total := len(items)
	return paginateSlice(items, page, pageSize), total
}

func (s *MemoryStore) UserFollows(userID int64, targetType string, page, pageSize int) ([]domain.FollowItem, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if userID <= 0 {
		userID = 1
	}
	targetType = normalizeOptionalTargetType(targetType, "")
	items := make([]domain.FollowItem, 0)
	for _, follow := range s.follows {
		if follow.UserID != userID {
			continue
		}
		if targetType != "" && follow.TargetType != targetType {
			continue
		}
		item := s.followItemLocked(follow)
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt == items[j].CreatedAt {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})
	total := len(items)
	return paginateSlice(items, page, pageSize), total
}

func (s *MemoryStore) UserActivities(userID int64, communityID int64, action string, page, pageSize int) ([]domain.Activity, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if userID <= 0 {
		userID = 1
	}
	action = strings.TrimSpace(action)
	out := make([]domain.Activity, 0)
	for _, activity := range s.activities {
		if activity.UserID != userID {
			continue
		}
		if communityID > 0 && activity.CommunityID != communityID {
			continue
		}
		if action != "" && activity.Action != action {
			continue
		}
		cp := *activity
		s.enrichActivityLocked(&cp)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt == out[j].CreatedAt {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt > out[j].CreatedAt
	})
	total := len(out)
	return paginateSlice(out, page, pageSize), total
}

func (s *MemoryStore) UserNotifications(userID int64, isRead *bool, page, pageSize int) ([]domain.Notification, int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if userID <= 0 {
		userID = 1
	}
	out := make([]domain.Notification, 0)
	unread := 0
	for _, notice := range s.notices {
		if notice.UserID != 0 && notice.UserID != userID {
			continue
		}
		read := notice.Read || notice.IsRead
		if !read {
			unread++
		}
		if isRead != nil && read != *isRead {
			continue
		}
		cp := *notice
		cp.Read = read
		cp.IsRead = read
		if cp.TargetURL == "" {
			cp.TargetURL = targetURLFor(cp.TargetType, cp.TargetID, cp.TopicID)
		}
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt == out[j].CreatedAt {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt > out[j].CreatedAt
	})
	total := len(out)
	return paginateSlice(out, page, pageSize), total, unread
}

func (s *MemoryStore) ReadUserNotification(userID int64, id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userID <= 0 {
		userID = 1
	}
	n, ok := s.notices[id]
	if !ok || (n.UserID != 0 && n.UserID != userID) {
		return false
	}
	now := Now()
	n.Read = true
	n.IsRead = true
	n.ReadAt = now
	return true
}

func (s *MemoryStore) ReadAllUserNotifications(userID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userID <= 0 {
		userID = 1
	}
	count := 0
	now := Now()
	for _, n := range s.notices {
		if n.UserID != 0 && n.UserID != userID {
			continue
		}
		if !n.Read && !n.IsRead {
			count++
		}
		n.Read = true
		n.IsRead = true
		if n.ReadAt == "" {
			n.ReadAt = now
		}
	}
	return count
}

func (s *MemoryStore) CommunityOverview(slug string) (domain.CommunityOverview, bool) {
	comm, ok := s.CommunityBySlug(slug)
	if !ok {
		return domain.CommunityOverview{}, false
	}

	categories := s.Categories(comm.ID)

	// 板块计数
	categoryCounts := make(map[string]int)
	for _, cat := range categories {
		categoryCounts[cat.Slug] = len(s.ListPosts(slug, cat.Slug, "", ""))
	}

	// 热门话题
	hotTopics := []domain.Topic{}
	hotResult, _ := s.TopicsByFilter(comm.ID, 0, "", "hot", nil, "", 1, 6)
	for _, t := range hotResult {
		hotTopics = append(hotTopics, t)
	}

	// 最新话题
	latestTopics := []domain.Topic{}
	latestResult, _ := s.TopicsByFilter(comm.ID, 0, "", "latest", nil, "", 1, 6)
	for _, t := range latestResult {
		latestTopics = append(latestTopics, t)
	}

	// 热门标签
	hotTags := s.TagStats(slug)

	// 统计
	posts := s.ListPosts(slug, "", "", "")
	stats := domain.PostStats{TotalPosts: len(posts)}
	for _, p := range posts {
		stats.TotalViews += p.Views
		stats.TotalLikes += p.Likes
		stats.TotalComments += p.Comments
	}

	return domain.CommunityOverview{
		Community:      comm,
		Categories:     categories,
		CategoryCounts: categoryCounts,
		HotTopics:      hotTopics,
		LatestTopics:   latestTopics,
		HotTags:        hotTags,
		Stats:          stats,
	}, true
}

func (s *MemoryStore) TopicComments(topicID int64, sortBy string, page, pageSize int) ([]*domain.Comment, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.posts[topicID]; !ok {
		return []*domain.Comment{}, 0
	}
	return s.commentsTreeLocked(topicID, sortBy, page, pageSize)
}

// CommentByID 返回评论详情。
func (s *MemoryStore) CommentByID(id int64) (*domain.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.comments[id]
	if !ok || c.Status == "deleted" {
		return nil, errors.New("评论不存在")
	}
	cp := *c
	s.normalizeCommentLocked(&cp)
	return &cp, nil
}

func (s *MemoryStore) CreateComment(topicID int64, author string, text string, parentID int64) (*domain.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[topicID]
	if !ok {
		return nil, errors.New("帖子不存在")
	}
	to := ""
	replyToUserID := int64(0)
	if parentID > 0 {
		parent, ok := s.comments[parentID]
		if !ok || parent.PostID != topicID {
			return nil, errors.New("父评论不存在")
		}
		to = parent.Author
		replyToUserID = parent.UserID
		if replyToUserID == 0 {
			replyToUserID = 1
		}
	}
	c := s.createCommentLocked(topicID, parentID, author, to, text, "", 0)
	c.UserID = 1
	c.TopicID = topicID
	c.ReplyToUserID = replyToUserID
	c.Content = c.Text
	c.UserName = firstNonEmptyString(c.Author, "Demo 用户")
	c.UpdatedAt = c.CreatedAt
	p.UpdatedAt = Now()
	communityID := s.communityIDBySlugLocked(p.Site)
	targetType := "topic"
	targetID := topicID
	if parentID > 0 {
		targetType = "comment"
		targetID = c.ID
	}
	s.appendActivityLocked(1, communityID, "commented", targetType, targetID, topicID, p.Title)
	if parentID > 0 {
		s.createUserNoticeLocked(replyToUserID, 1, "comment_replied", "comment", c.ID, topicID, c.ID, "你的评论有新的回复", fmt.Sprintf("%s 回复了你的评论。", c.Author))
	} else {
		s.createUserNoticeLocked(1, 1, "topic_commented", "topic", topicID, topicID, c.ID, "你的主题有新的评论", fmt.Sprintf("%s 评论了《%s》。", c.Author, p.Title))
	}
	cp := *c
	s.normalizeCommentLocked(&cp)
	return &cp, nil
}

func (s *MemoryStore) CreateCommentWithRequest(topicID int64, req domain.CreateCommentRequest) (*domain.Comment, error) {
	userID := req.UserID
	if userID <= 0 {
		userID = 1
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		content = strings.TrimSpace(req.Text)
	}
	if len([]rune(content)) < 2 {
		return nil, errors.New("评论内容至少 2 个字符")
	}
	if len([]rune(content)) > 5000 {
		return nil, errors.New("评论内容最多 5000 个字符")
	}
	author := strings.TrimSpace(req.Author)
	if author == "" {
		author = "Demo 用户"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[topicID]
	if !ok {
		return nil, errors.New("主题不存在")
	}
	if !memoryPostVisible(p) {
		return nil, errors.New("主题已隐藏")
	}
	if s.commentLocks[topicID] {
		return nil, errors.New("评论已锁定")
	}
	to := ""
	replyToUserID := int64(0)
	if req.ParentID > 0 {
		parent, ok := s.comments[req.ParentID]
		if !ok || parent.PostID != topicID || parent.Status == "deleted" || parent.Status == "hidden" {
			return nil, errors.New("父评论不存在")
		}
		to = parent.Author
		replyToUserID = parent.UserID
	}
	c := s.createCommentLocked(topicID, req.ParentID, author, to, content, "", 0)
	c.UserID = userID
	c.UserName = author
	c.Author = author
	c.Content = content
	c.Text = content
	c.TopicID = topicID
	c.ReplyToUserID = replyToUserID
	c.UpdatedAt = c.CreatedAt
	p.UpdatedAt = Now()
	communityID := s.communityIDBySlugLocked(p.Site)
	targetType := "topic"
	targetID := topicID
	if req.ParentID > 0 {
		targetType = "comment"
		targetID = req.ParentID
	}
	s.appendActivityLocked(userID, communityID, "commented", targetType, targetID, topicID, p.Title)
	if req.ParentID > 0 {
		s.createUserNoticeLocked(replyToUserID, userID, "comment_replied", "comment", c.ID, topicID, c.ID, "你的评论有新的回复", fmt.Sprintf("%s 回复了你的评论。", author))
	} else {
		s.createUserNoticeLocked(p.UserID, userID, "topic_commented", "topic", topicID, topicID, c.ID, "你的主题有新的评论", fmt.Sprintf("%s 评论了《%s》。", author, p.Title))
	}
	if contentTypeForBoard(p.Board) == "question" {
		answerID := int64(len(s.qaAnswers) + 1)
		s.qaAnswers[answerID] = &domain.QAAnswer{
			ID:         answerID,
			TopicID:    topicID,
			CommentID:  c.ID,
			UserID:     userID,
			IsAccepted: false,
			CreatedAt:  c.CreatedAt,
			UpdatedAt:  c.CreatedAt,
		}
		if q, ok := s.qaQuestions[topicID]; ok && q != nil {
			q.AnswerCount++
			q.UpdatedAt = Now()
		}
		s.createUserNoticeLocked(p.UserID, userID, "qa.question.answered", "topic", topicID, topicID, c.ID, "你的问题有了新回答", fmt.Sprintf("%s 回答了《%s》。", author, p.Title))
	}
	cp := *c
	s.normalizeCommentLocked(&cp)
	return &cp, nil
}

func (s *MemoryStore) AcceptBestAnswer(topicID int64, commentID int64, actorUserID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[topicID]
	if !ok || contentTypeForBoard(p.Board) != "question" {
		return false
	}
	c, ok := s.comments[commentID]
	if !ok || c.PostID != topicID || c.Status == "deleted" || c.Status == "hidden" {
		return false
	}
	if actorUserID <= 0 {
		actorUserID = p.UserID
		if actorUserID <= 0 {
			actorUserID = 1
		}
	}
	for _, item := range s.comments {
		if item.PostID == topicID {
			item.IsBest = false
		}
	}
	c.IsBest = true
	now := Now()
	p.UpdatedAt = now
	c.UpdatedAt = now
	if q, ok := s.qaQuestions[topicID]; ok && q != nil {
		q.IsResolved = true
		q.AcceptedAnswerID = commentID
		q.AcceptedAt = now
		q.UpdatedAt = now
	}
	for _, answer := range s.qaAnswers {
		if answer == nil || answer.TopicID != topicID {
			continue
		}
		answer.IsAccepted = answer.CommentID == commentID
		if answer.IsAccepted {
			answer.AcceptedAt = now
		} else {
			answer.AcceptedAt = ""
		}
		answer.UpdatedAt = now
	}
	communityID := s.communityIDBySlugLocked(p.Site)
	s.appendActivityLocked(actorUserID, communityID, "accepted_answer", "comment", commentID, topicID, p.Title)
	receiverID := c.UserID
	if receiverID == 0 {
		receiverID = 1
	}
	s.createUserNoticeLocked(receiverID, actorUserID, "answer_accepted", "comment", commentID, topicID, commentID, "你的回答被采纳", fmt.Sprintf("你在《%s》中的回答被采纳为最佳答案。", p.Title))
	s.createUserNoticeLocked(receiverID, actorUserID, "qa.answer.accepted", "comment", commentID, topicID, commentID, "问答已采纳最佳答案", fmt.Sprintf("《%s》已采纳你的回答。", p.Title))
	return true
}

func (s *MemoryStore) QAQuestionByTopicID(topicID int64) (*domain.QAQuestion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.qaQuestions[topicID]
	if !ok || item == nil {
		return nil, errors.New("问答扩展不存在")
	}
	cp := *item
	return &cp, nil
}

func (s *MemoryStore) QAAnswersByTopicID(topicID int64) ([]domain.QAAnswer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []domain.QAAnswer{}
	for _, item := range s.qaAnswers {
		if item != nil && item.TopicID == topicID {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (s *MemoryStore) DocsDocumentByTopicID(topicID int64) (*domain.DocsDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.docsDocuments[topicID]
	if !ok || item == nil {
		return nil, errors.New("文档扩展不存在")
	}
	cp := *item
	return &cp, nil
}

func (s *MemoryStore) DocsTree(communityID int64, spaceID int64) ([]domain.DocsDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []domain.DocsDocument{}
	for _, item := range s.docsDocuments {
		if item == nil {
			continue
		}
		if spaceID > 0 && item.SpaceID != spaceID {
			continue
		}
		items = append(items, *item)
	}
	return items, nil
}

func (s *MemoryStore) WikiPageByTopicID(topicID int64) (*domain.WikiPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.wikiPages[topicID]
	if !ok || item == nil {
		return nil, errors.New("Wiki 页面扩展不存在")
	}
	cp := *item
	return &cp, nil
}

func (s *MemoryStore) WikiVersionsByTopicID(topicID int64) ([]domain.WikiRevision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	page, ok := s.wikiPages[topicID]
	if !ok || page == nil {
		return nil, errors.New("Wiki 页面扩展不存在")
	}
	items := []domain.WikiRevision{}
	for _, item := range s.wikiVersions {
		if item != nil && item.WikiPageID == page.ID {
			items = append(items, *item)
		}
	}
	return items, nil
}

// CreateReport 创建举报记录。
func (s *MemoryStore) CreateReport(req domain.CreateReportRequest) (*domain.Report, error) {
	reporterID := req.ReporterUserID
	if reporterID <= 0 {
		reporterID = 1
	}
	targetType := strings.TrimSpace(req.TargetType)
	if !validReportTargetType(targetType) {
		return nil, errors.New("举报对象类型不合法")
	}
	reasonType := strings.TrimSpace(req.ReasonType)
	if reasonType == "" {
		return nil, errors.New("举报原因不能为空")
	}
	reasonText := strings.TrimSpace(req.ReasonText)
	if len([]rune(reasonText)) > 500 {
		return nil, errors.New("举报说明最多 500 字")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	communityID, topicID, title, content, err := s.reportTargetContextLocked(targetType, req.TargetID)
	if err != nil {
		return nil, err
	}
	for _, existing := range s.reports {
		if existing.ReporterUserID == reporterID && existing.TargetType == targetType && existing.TargetID == req.TargetID && existing.Status == "pending" {
			return nil, errors.New("同一对象已有待处理举报，请勿重复提交")
		}
	}
	now := Now()
	report := &domain.Report{
		ID:             s.nextReportID,
		ReporterID:     reporterID,
		ReporterUserID: reporterID,
		TargetType:     targetType,
		TargetID:       req.TargetID,
		CommunityID:    communityID,
		TopicID:        topicID,
		ReasonType:     reasonType,
		ReasonText:     reasonText,
		Status:         "pending",
		TargetTitle:    title,
		TargetContent:  content,
		TargetURL:      reportTargetURL(targetType, req.TargetID, topicID),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.nextReportID++
	s.enrichReportLocked(report)
	s.reports[report.ID] = report
	cp := *report
	return &cp, nil
}

// Reports 返回后台举报列表。
func (s *MemoryStore) Reports(filter domain.ReportFilter) ([]domain.Report, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []domain.Report{}
	for _, report := range s.reports {
		if filter.Status != "" && filter.Status != "all" && report.Status != filter.Status {
			continue
		}
		if filter.TargetType != "" && filter.TargetType != "all" && report.TargetType != filter.TargetType {
			continue
		}
		if filter.CommunityID > 0 && report.CommunityID != filter.CommunityID {
			continue
		}
		if !filter.ActorIsAdmin && !s.isCommunityModeratorLocked(filter.ActorUserID, report.CommunityID) {
			continue
		}
		cp := *report
		s.enrichReportLocked(&cp)
		items = append(items, cp)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	total := len(items)
	page, pageSize := normalizeMemoryPage(filter.Page, filter.PageSize)
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []domain.Report{}, total
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total
}

// ReportByID 返回举报详情。
func (s *MemoryStore) ReportByID(id int64) (*domain.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	report, ok := s.reports[id]
	if !ok {
		return nil, errors.New("举报不存在")
	}
	cp := *report
	s.enrichReportLocked(&cp)
	return &cp, nil
}

// HandleReport 处理举报。
func (s *MemoryStore) HandleReport(id int64, status, note string, handlerUserID int64) (*domain.Report, error) {
	status = strings.TrimSpace(status)
	if status != "accepted" && status != "rejected" {
		return nil, errors.New("处理状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	report, ok := s.reports[id]
	if !ok {
		return nil, errors.New("举报不存在")
	}
	if status == "accepted" {
		if err := s.hideReportTargetLocked(report); err != nil {
			return nil, err
		}
	}
	report.Status = status
	report.HandledBy = handlerUserID
	report.HandledAt = Now()
	report.HandleNote = strings.TrimSpace(note)
	report.UpdatedAt = report.HandledAt
	s.enrichReportLocked(report)
	cp := *report
	return &cp, nil
}

// IsCommunityModerator 判断用户是否为子站版主。
func (s *MemoryStore) IsCommunityModerator(userID, communityID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isCommunityModeratorLocked(userID, communityID)
}

func (s *MemoryStore) CommunityModerators(filter domain.CommunityModeratorFilter) ([]domain.CommunityModerator, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	communityID := filter.CommunityID
	if communityID == 0 && filter.CommunitySlug != "" && filter.CommunitySlug != "all" && filter.CommunitySlug != "portal" {
		communityID = s.communityIDBySlugLocked(filter.CommunitySlug)
	}
	items := []domain.CommunityModerator{}
	for _, moderator := range s.moderators {
		if communityID > 0 && moderator.CommunityID != communityID {
			continue
		}
		if filter.UserID > 0 && moderator.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && filter.Status != "all" {
			want := 1
			if filter.Status == "0" || strings.EqualFold(filter.Status, "disabled") {
				want = 0
			}
			if moderator.Status != want {
				continue
			}
		}
		if !filter.ActorIsAdmin && !s.isCommunityModeratorLocked(filter.ActorUserID, moderator.CommunityID) {
			continue
		}
		items = append(items, s.enrichModeratorLocked(*moderator))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CommunityID == items[j].CommunityID {
			return items[i].ID > items[j].ID
		}
		return items[i].CommunityID < items[j].CommunityID
	})
	total := len(items)
	page, pageSize := normalizeMemoryPage(filter.Page, filter.PageSize)
	return paginateSlice(items, page, pageSize), total
}

func (s *MemoryStore) CommunityModeratorByID(id int64) (*domain.CommunityModerator, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	moderator, ok := s.moderators[id]
	if !ok {
		return nil, errors.New("版主不存在")
	}
	cp := s.enrichModeratorLocked(*moderator)
	return &cp, nil
}

func (s *MemoryStore) CreateCommunityModerator(req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	moderator, err := s.normalizeModeratorRequestLocked(req, nil)
	if err != nil {
		return nil, err
	}
	for _, existing := range s.moderators {
		if existing.CommunityID == moderator.CommunityID && existing.UserID == moderator.UserID {
			return nil, errors.New("该用户已经是当前子站版主")
		}
	}
	moderator.ID = s.nextModeratorID
	s.nextModeratorID++
	now := Now()
	moderator.CreatedAt = now
	moderator.UpdatedAt = now
	s.moderators[moderator.ID] = moderator
	cp := s.enrichModeratorLocked(*moderator)
	return &cp, nil
}

func (s *MemoryStore) UpdateCommunityModerator(id int64, req domain.CommunityModeratorRequest) (*domain.CommunityModerator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.moderators[id]
	if !ok {
		return nil, errors.New("版主不存在")
	}
	updated, err := s.normalizeModeratorRequestLocked(req, current)
	if err != nil {
		return nil, err
	}
	for _, existing := range s.moderators {
		if existing.ID != id && existing.CommunityID == updated.CommunityID && existing.UserID == updated.UserID {
			return nil, errors.New("该用户已经是当前子站版主")
		}
	}
	updated.ID = current.ID
	updated.CreatedAt = current.CreatedAt
	updated.UpdatedAt = Now()
	s.moderators[id] = updated
	cp := s.enrichModeratorLocked(*updated)
	return &cp, nil
}

func (s *MemoryStore) DeleteCommunityModerator(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	moderator, ok := s.moderators[id]
	if !ok {
		return false
	}
	moderator.Status = 0
	moderator.UpdatedAt = Now()
	return true
}

func (s *MemoryStore) SetTopicFeatured(id int64, featured bool) (*domain.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("主题不存在")
	}
	p.Recommended = featured
	p.UpdatedAt = Now()
	topic, _ := s.topicFromPostLocked(id, false)
	return &topic, nil
}

func (s *MemoryStore) SetTopicPinned(id int64, pinned bool) (*domain.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("主题不存在")
	}
	p.Pinned = pinned
	p.UpdatedAt = Now()
	topic, _ := s.topicFromPostLocked(id, false)
	return &topic, nil
}

func (s *MemoryStore) SetTopicStatus(id int64, status int) (*domain.Topic, error) {
	if status != 0 && status != 1 && status != 3 {
		return nil, errors.New("主题状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("主题不存在")
	}
	memorySetPostStatus(p, status)
	p.UpdatedAt = Now()
	topic, _ := s.topicFromPostLocked(id, false)
	return &topic, nil
}

func (s *MemoryStore) SetTopicCommentLocked(id int64, locked bool) (*domain.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return nil, errors.New("主题不存在")
	}
	s.commentLocks[id] = locked
	p.UpdatedAt = Now()
	topic, _ := s.topicFromPostLocked(id, false)
	return &topic, nil
}

func (s *MemoryStore) SetCommentStatus(id int64, status string) (*domain.Comment, error) {
	status = strings.TrimSpace(status)
	if status != "normal" && status != "hidden" {
		return nil, errors.New("评论状态不合法")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.comments[id]
	if !ok || c.Status == "deleted" {
		return nil, errors.New("评论不存在")
	}
	if status == "hidden" && c.IsBest {
		return nil, errors.New("最佳答案不能隐藏")
	}
	c.Status = status
	c.UpdatedAt = Now()
	cp := *c
	s.normalizeCommentLocked(&cp)
	return &cp, nil
}

func (s *MemoryStore) reportTargetContextLocked(targetType string, targetID int64) (int64, int64, string, string, error) {
	switch targetType {
	case "topic":
		p, ok := s.posts[targetID]
		if !ok {
			return 0, 0, "", "", errors.New("主题不存在")
		}
		return s.communityIDBySlugLocked(p.Site), p.ID, p.Title, firstNonEmptyString(p.Summary, p.Content), nil
	case "comment":
		c, ok := s.comments[targetID]
		if !ok || c.Status == "deleted" {
			return 0, 0, "", "", errors.New("评论不存在")
		}
		p, ok := s.posts[c.PostID]
		if !ok {
			return 0, 0, "", "", errors.New("主题不存在")
		}
		return s.communityIDBySlugLocked(p.Site), p.ID, p.Title, c.Content, nil
	case "user", "wiki":
		return 0, 0, fmt.Sprintf("%s#%d", targetType, targetID), "", nil
	default:
		return 0, 0, "", "", errors.New("举报对象类型不合法")
	}
}

func (s *MemoryStore) enrichReportLocked(report *domain.Report) {
	if report == nil {
		return
	}
	report.ReporterUserID = report.ReporterID
	if report.CommunityID > 0 {
		comm := s.communityByIDLocked(report.CommunityID)
		report.CommunitySlug = comm.Slug
		report.CommunityName = comm.Name
	}
	if report.TargetTitle == "" || report.TargetContent == "" {
		if communityID, topicID, title, content, err := s.reportTargetContextLocked(report.TargetType, report.TargetID); err == nil {
			if report.CommunityID == 0 {
				report.CommunityID = communityID
			}
			if report.TopicID == 0 {
				report.TopicID = topicID
			}
			report.TargetTitle = title
			report.TargetContent = content
		}
	}
	report.TargetURL = reportTargetURL(report.TargetType, report.TargetID, report.TopicID)
}

func (s *MemoryStore) hideReportTargetLocked(report *domain.Report) error {
	switch report.TargetType {
	case "topic":
		if p, ok := s.posts[report.TargetID]; ok {
			memorySetPostStatus(p, 0)
			p.UpdatedAt = Now()
			return nil
		}
	case "comment":
		if c, ok := s.comments[report.TargetID]; ok {
			if c.IsBest {
				return errors.New("最佳答案不能隐藏")
			}
			c.Status = "hidden"
			c.UpdatedAt = Now()
			return nil
		}
	}
	return nil
}

func (s *MemoryStore) isCommunityModeratorLocked(userID, communityID int64) bool {
	if userID <= 0 || communityID <= 0 {
		return false
	}
	for _, moderator := range s.moderators {
		if moderator.UserID == userID && moderator.CommunityID == communityID && moderator.Status == 1 {
			return true
		}
	}
	return false
}

func (s *MemoryStore) normalizeModeratorRequestLocked(req domain.CommunityModeratorRequest, current *domain.CommunityModerator) (*domain.CommunityModerator, error) {
	moderator := &domain.CommunityModerator{Role: "moderator", Status: 1}
	if current != nil {
		cp := *current
		moderator = &cp
	}
	communityID := req.CommunityID
	if communityID == 0 && strings.TrimSpace(req.CommunitySlug) != "" {
		communityID = s.communityIDBySlugLocked(strings.TrimSpace(req.CommunitySlug))
	}
	if communityID > 0 {
		if s.communitySlugByIDLocked(communityID) == "" {
			return nil, errors.New("子站不存在")
		}
		moderator.CommunityID = communityID
	}
	if moderator.CommunityID <= 0 {
		return nil, errors.New("请选择子站")
	}
	if req.UserID > 0 {
		moderator.UserID = req.UserID
	}
	if moderator.UserID <= 0 {
		return nil, errors.New("请选择用户")
	}
	if _, ok := s.users[moderator.UserID]; !ok {
		return nil, errors.New("用户不存在")
	}
	if strings.TrimSpace(req.Role) != "" {
		moderator.Role = strings.TrimSpace(req.Role)
	}
	if moderator.Role == "" {
		moderator.Role = "moderator"
	}
	if moderator.Role != "moderator" && moderator.Role != "owner" {
		return nil, errors.New("版主角色不合法")
	}
	if req.Status != nil {
		moderator.Status = *req.Status
	}
	if moderator.Status != 0 && moderator.Status != 1 {
		return nil, errors.New("版主状态不合法")
	}
	return moderator, nil
}

func (s *MemoryStore) enrichModeratorLocked(m domain.CommunityModerator) domain.CommunityModerator {
	if comm := s.communityByIDLocked(m.CommunityID); comm.ID > 0 {
		m.CommunitySlug = comm.Slug
		m.CommunityName = comm.Name
	}
	if user, ok := s.users[m.UserID]; ok {
		m.UserName = user.Username
		m.UserNickname = user.Nickname
	}
	return m
}

func validReportTargetType(targetType string) bool {
	switch targetType {
	case "topic", "comment", "user", "wiki":
		return true
	default:
		return false
	}
}

func reportTargetURL(targetType string, targetID, topicID int64) string {
	switch targetType {
	case "topic":
		return fmt.Sprintf("/topics/%d/", targetID)
	case "comment":
		if topicID > 0 {
			return fmt.Sprintf("/topics/%d/#comment-%d", topicID, targetID)
		}
	}
	return "/"
}

func normalizeMemoryPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}
