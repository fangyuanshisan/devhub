package service

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLStorePluginPlatformConsistency(t *testing.T) {
	if os.Getenv("DEVHUB_MYSQL_TESTS") != "1" {
		t.Skip("set DEVHUB_MYSQL_TESTS=1 and DB_NAME to a disposable test database to run MySQLStore plugin consistency checks")
	}

	cfg := mysqlTestConfig()
	if !strings.Contains(strings.ToLower(cfg.Database), "test") {
		t.Fatalf("refusing to run MySQL integration test against non-test database %q", cfg.Database)
	}

	repo, err := store.NewMySQLStore(cfg)
	if err != nil {
		t.Fatalf("NewMySQLStore failed: %v", err)
	}
	defer repo.Close()

	db, err := sql.Open("mysql", mysqlDSN(cfg))
	if err != nil {
		t.Fatalf("open mysql failed: %v", err)
	}
	defer db.Close()

	assertMySQLPluginSchema(t, db)

	svc := New(repo)
	actor := domain.ActorContext{
		UserID:      1,
		Permissions: []string{"qa.question.create", "docs.document.create", "wiki.page.create", "core.topic.create"},
	}

	t.Run("global plugin disable blocks matching content type", func(t *testing.T) {
		if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusDisabled); err != nil {
			t.Fatalf("disable qa failed: %v", err)
		}
		defer func() {
			if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
				t.Fatalf("restore qa failed: %v", err)
			}
		}()
		_, err := svc.CreateTopic(mysqlTestTopicRequest(1, 102, "question", actor, "global disabled"))
		if err == nil || !strings.Contains(err.Error(), "插件未启用") {
			t.Fatalf("expected global disabled plugin to block question create, got %v", err)
		}
	})

	t.Run("community plugin disable only blocks that community", func(t *testing.T) {
		if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
			t.Fatalf("enable qa failed: %v", err)
		}
		if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusDisabled); err != nil {
			t.Fatalf("disable qa for community 1 failed: %v", err)
		}
		defer func() {
			if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusEnabled); err != nil {
				t.Fatalf("restore community qa failed: %v", err)
			}
		}()

		_, err := svc.CreateTopic(mysqlTestTopicRequest(1, 102, "question", actor, "community disabled"))
		if err == nil || !strings.Contains(err.Error(), "当前子站未启用插件") {
			t.Fatalf("expected community disabled plugin to block community 1 create, got %v", err)
		}

		topic, err := svc.CreateTopic(mysqlTestTopicRequest(2, 202, "question", actor, "community enabled"))
		if err != nil {
			t.Fatalf("expected community 2 create to pass while community 1 is disabled, got %v", err)
		}
		if topic.PluginCode != "qa" || topic.ContentType != "question" {
			t.Fatalf("expected normalized qa topic, got plugin=%q type=%q", topic.PluginCode, topic.ContentType)
		}
	})

	t.Run("failed migration blocks enable and retry restores readiness", func(t *testing.T) {
		if _, err := svc.InjectFailedPluginMigrationForTest("qa", "qa_questions", "mysql integration failed migration", "mysql-integration"); err != nil {
			t.Fatalf("inject failed migration: %v", err)
		}
		if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err == nil || !strings.Contains(err.Error(), "失败迁移") {
			t.Fatalf("expected failed migration to block global enable, got %v", err)
		}
		if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusEnabled); err == nil || !strings.Contains(err.Error(), "失败迁移") {
			t.Fatalf("expected failed migration to block community enable, got %v", err)
		}
		if _, err := svc.RunPluginMigration("qa", "qa_questions", "mysql-integration"); err != nil {
			t.Fatalf("retry migration failed: %v", err)
		}
		if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
			t.Fatalf("enable after retry should pass: %v", err)
		}
	})

	t.Run("config schema validation is enforced", func(t *testing.T) {
		if _, err := svc.SetPluginConfig("qa", `{"allow_anonymous_answer":false,"default_question_status":"bad"}`); err == nil || !strings.Contains(err.Error(), "enum") {
			t.Fatalf("expected invalid enum to fail, got %v", err)
		}
		if _, err := svc.SetCommunityPluginConfig(1, "docs", `{"allow_public_spaces":true,"max_tree_depth":"deep"}`); err == nil || (!strings.Contains(err.Error(), "类型") && !strings.Contains(err.Error(), "integer")) {
			t.Fatalf("expected invalid community config type to fail, got %v", err)
		}
		if _, err := svc.SetCommunityPluginConfig(1, "docs", `{"allow_public_spaces":true,"max_tree_depth":5}`); err != nil {
			t.Fatalf("valid community config should pass: %v", err)
		}
	})

	t.Run("hook execution and plugin audit are queryable", func(t *testing.T) {
		if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
			t.Fatalf("enable qa failed: %v", err)
		}
		if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusEnabled); err != nil {
			t.Fatalf("enable community qa failed: %v", err)
		}
		if err := svc.DispatchHook(pluginregistry.HookEvent{
			Name: pluginregistry.HookBeforeCreateContent,
			Mode: pluginregistry.HookBlocking,
			Ctx: pluginregistry.HookContext{
				PluginCode:  "qa",
				ContentType: "question",
				CommunityID: 1,
				CategoryID:  102,
				ActorType:   pluginregistry.HookActorSystem,
			},
		}); err != nil {
			t.Fatalf("dispatch hook failed: %v", err)
		}
		records, err := svc.HookExecutions("qa", 10)
		if err != nil {
			t.Fatalf("HookExecutions failed: %v", err)
		}
		if len(records) == 0 {
			t.Fatal("expected hook execution records")
		}

		svc.AppendAdminLog(domain.AdminLog{
			Site:        "portal",
			Type:        "system",
			Actor:       "mysql-integration",
			ActorType:   "system",
			Action:      "plugin.mysql.integration",
			Target:      "plugins#qa",
			TargetType:  "plugin",
			CommunityID: 1,
			Metadata:    `{"plugin_code":"qa","store":"mysql"}`,
		})
		logs, total := svc.AdminLogsByFilter(domain.AdminLogFilter{Action: "plugin.mysql.integration", Target: "plugins#qa", Page: 1, PageSize: 10})
		if total == 0 || len(logs) == 0 {
			t.Fatalf("expected plugin audit log to be queryable, total=%d logs=%#v", total, logs)
		}
	})
}

func mysqlTestConfig() store.MySQLConfig {
	return store.MySQLConfig{
		Host:     envOrDefault("DB_HOST", "127.0.0.1"),
		Port:     envOrDefault("DB_PORT", "3307"),
		User:     envOrDefault("DB_USER", "devhub"),
		Password: envOrDefault("DB_PASSWORD", "Devhub_123456"),
		Database: envOrDefault("DB_NAME", "devhub_test"),
	}
}

func mysqlDSN(cfg store.MySQLConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func mysqlTestTopicRequest(communityID, categoryID int64, contentType string, actor domain.ActorContext, suffix string) domain.CreateTopicRequest {
	now := time.Now().UnixNano()
	return domain.CreateTopicRequest{
		UserID:           1,
		CommunityID:      communityID,
		CategoryID:       categoryID,
		ContentType:      contentType,
		Title:            fmt.Sprintf("MySQL Plugin E2E %s %d", suffix, now),
		Summary:          "MySQLStore plugin consistency integration test",
		Content:          "MySQLStore plugin consistency integration test content.",
		Tags:             []string{"mysql-e2e"},
		ActorPermissions: append([]string{}, actor.Permissions...),
		ActorContext:     actor,
	}
}

func assertMySQLPluginSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, table := range []string{"plugins", "community_plugins", "plugin_migrations", "hook_executions", "admin_logs"} {
		if !mysqlTableExists(t, db, table) {
			t.Fatalf("expected table %s to exist", table)
		}
	}
	for table, columns := range map[string][]string{
		"topics":            {"plugin_code"},
		"categories":        {"plugin_code", "allowed_content_types"},
		"plugins":           {"plugin_code", "status", "config_json"},
		"community_plugins": {"community_id", "plugin_code", "status", "sort_order", "config_json"},
		"plugin_migrations": {"plugin_code", "version", "migration_name", "status", "error_message"},
		"hook_executions":   {"plugin_code", "hook_name", "success", "metadata_json"},
		"admin_logs":        {"old_value", "new_value", "metadata_json"},
	} {
		for _, column := range columns {
			if !mysqlColumnExists(t, db, table, column) {
				t.Fatalf("expected column %s.%s to exist", table, column)
			}
		}
	}
}

func mysqlTableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=?`, table).Scan(&count); err != nil {
		t.Fatalf("checking table %s failed: %v", table, err)
	}
	return count > 0
}

func mysqlColumnExists(t *testing.T, db *sql.DB, table, column string) bool {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND COLUMN_NAME=?`, table, column).Scan(&count); err != nil {
		t.Fatalf("checking column %s.%s failed: %v", table, column, err)
	}
	return count > 0
}
