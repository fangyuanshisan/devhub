# Repository Agent Rules

- Treat tests as manual-only when completing tasks here: do not run `go test`, Playwright, or other verification suites unless the user explicitly asks.
- When a task needs validation, leave the runnable command or script for the user to execute manually.
- Use `scripts/test-all.sh` as the one-click manual verification entrypoint.
- Do not revert user changes or unrelated edits.
