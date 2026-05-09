# Changelog

## v1.0.0

DevHub v1.0.0 is the first runnable archive release of the project.

### Added

- Multi-community DevHub structure for Portal, PHP, Go, Java, AI, and Frontend communities.
- Topic publishing, listing, detail, editing, and compatibility `sites/posts` APIs.
- Search and filtering by keyword, community, content type, tag, status, featured, and unsolved questions.
- Basic tag capability, hot tags, and community tag aggregation.
- Topic likes, favorites, follows, user activities, and notifications.
- Comments, replies, question solved state, best-answer acceptance, and unsolved filtering.
- Topic and comment reports, moderator-scoped governance, featured, pinned, hidden, restored, and comment-lock moderation.
- admin-next backend for content CRUD, comments, reports, moderator CRUD, batch governance, and audit logs.
- Baidu-friendly dynamic SEO HTML for `/topics/:id`, dynamic sitemap, and robots.txt.
- MemoryStore and MySQLStore modes with MySQL 8 schema coverage.
- Testing, deployment, SEO, backup, rollback, and release archive documentation.
- Basic GitHub Actions CI for Go tests/builds, frontend build, admin build, SQL checks, and docs checks.

### Changed

- Default port is unified to `8090` across `main.go`, `dev.sh`, README, docs, and admin dev proxy.
- `/topics/:id` keeps Go-rendered SEO HTML and uses the current Astro CSS asset from the build output.
- admin-next frontend API wrappers are aligned with backend moderator, batch governance, report, and audit-log routes.
- Documentation has been consolidated around the v1.0.0 runnable release scope.

### Known Limitations

- Advanced tag features are planned for v1.1.0, including tag detail SEO pages, tag admin, aliases, merging, and trends.
- Runtime comment likes are not part of v1.0.0.
- Accepted questions support changing the best answer, but do not yet support canceling solved status.
- Tag-follow and user-follow backend support exists, while richer frontend entry points remain future work.
- Sitemap output is dynamic but not yet sharded for very large content volumes.
- Production deployment still needs environment-specific process supervision, reverse proxy, HTTPS, logging, and backup scheduling.
