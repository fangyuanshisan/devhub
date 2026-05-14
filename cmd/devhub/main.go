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
}

func runPluginNew(args []string) {
	fs := flag.NewFlagSet("plugin:new", flag.ExitOnError)
	code := fs.String("code", "", "plugin code, lowercase letters/digits/underscore")
	name := fs.String("name", "", "plugin display name")
	contentType := fs.String("content_type", "", "content type, defaults to {code}_item")
	contentName := fs.String("content_name", "", "content type display name")
	description := fs.String("description", "", "plugin description")
	author := fs.String("author", "", "plugin author")
	output := fs.String("output", "examples/plugins", "output parent directory")
	withConfig := fs.Bool("with_config", true, "include config_schema example")
	withHooks := fs.Bool("with_hooks", true, "include hook declarations")
	withMigration := fs.Bool("with_migration", true, "include migration declaration")
	force := fs.Bool("force", false, "overwrite existing output directory")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	result, err := scaffold.Generate(scaffold.Options{
		Code:                   *code,
		Name:                   *name,
		ContentType:            *contentType,
		ContentName:            *contentName,
		Description:            *description,
		Author:                 *author,
		Output:                 *output,
		WithConfig:             *withConfig,
		WithHooks:              *withHooks,
		WithMigration:          *withMigration,
		IncludeRegistryExample: true,
		Force:                  *force,
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
