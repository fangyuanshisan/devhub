package main

import (
	"flag"
	"fmt"
	"os"

	"devhub-gin-backend/internal/plugins/scaffold"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "plugin:new" {
		runPluginNew(os.Args[2:])
		return
	}
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, `  go run ./cmd/devhub plugin:new --code demo_links --name "Demo Links" --content_type demo_link --content_name "链接"`)
	fmt.Fprintln(os.Stderr, `  go run ./cmd/devhub plugin:new --plugin_type external_service --code demo_webhook --name "Demo Webhook"`)
	fmt.Fprintln(os.Stderr, `  go run ./cmd/devhub plugin:new --plugin_type frontend_mount --code demo_mount --name "Demo Mount" --mount_point admin.plugin.detail.preview --component_key official.announcement.card`)
}

func runPluginNew(args []string) {
	fs := flag.NewFlagSet("plugin:new", flag.ExitOnError)
	code := fs.String("code", "", "plugin code, lowercase letters/digits/underscore/hyphen")
	name := fs.String("name", "", "plugin display name")
	pluginType := fs.String("plugin_type", "content", "plugin type: content, external_service, admin_tool, frontend_mount")
	contentType := fs.String("content_type", "", "content type, defaults to {code}_item")
	contentName := fs.String("content_name", "", "content type display name")
	description := fs.String("description", "", "plugin description")
	author := fs.String("author", "", "plugin author")
	mountPoint := fs.String("mount_point", "", "frontend mount point, for frontend_mount templates")
	componentKey := fs.String("component_key", "", "frontend component key, for frontend_mount templates")
	healthCheckPath := fs.String("health_check_path", "", "external service hook path")
	timeoutMS := fs.Int("timeout_ms", 3000, "external service timeout in milliseconds")
	failurePolicy := fs.String("failure_policy", "warn", "external service failure policy: warn, log, ignore")
	output := fs.String("output", "examples/plugins", "output parent directory")
	withConfig := fs.Bool("with_config", true, "include config_schema example")
	withHooks := fs.Bool("with_hooks", true, "include hook declarations")
	withMigration := fs.Bool("with_migration", true, "include migration declaration")
	force := fs.Bool("force", false, "overwrite existing output directory")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	seen := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})
	normalizedPluginType := scaffold.NormalizePluginType(*pluginType)
	if !seen["with_hooks"] {
		*withHooks = normalizedPluginType == scaffold.PluginTypeContent || normalizedPluginType == scaffold.PluginTypeExternalService
	}
	if !seen["with_migration"] {
		*withMigration = normalizedPluginType == scaffold.PluginTypeContent
	}

	result, err := scaffold.Generate(scaffold.Options{
		Code:               *code,
		Name:               *name,
		PluginType:         *pluginType,
		ContentType:        *contentType,
		ContentName:        *contentName,
		Description:        *description,
		Author:             *author,
		MountPoint:         *mountPoint,
		ComponentKey:       *componentKey,
		HealthCheckPath:    *healthCheckPath,
		TimeoutMS:          *timeoutMS,
		FailurePolicy:      *failurePolicy,
		Output:             *output,
		WithConfig:         *withConfig,
		WithHooks:          *withHooks,
		WithMigration:      *withMigration,
		IncludeRegistryDoc: true,
		Force:              *force,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Generated plugin template: %s\n", result.Dir)
	for _, file := range result.Files {
		fmt.Printf("- %s\n", file)
	}
}
