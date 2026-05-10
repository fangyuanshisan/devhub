# Changelog

## v1.1.5

DevHub v1.1.5 is the frontend UI polish release.

### Changed

- Unified frontend visual tokens for colors, typography, spacing, borders, shadows, radius, and responsive page width.
- Polished the frontend header, site switcher, search box, logged-in user menu, publish button, and moderator workspace entry.
- Improved homepage, community pages, topic cards, topic detail typography, search page, publish page, user-center pages, notification cards, empty states, and lightweight moderator workspace visuals.
- Added responsive refinements for desktop, tablet, and mobile layouts.

### Unchanged

- No API, Store, database schema, route, auth, follow, publish, comment, moderator-permission, or admin-next business logic was changed.
- `/topics/:id`, `/c/:slug`, and `/tags/:tag` remain Go-rendered SEO pages and were not converted to CSR shells.

## v1.1.4

DevHub v1.1.4 is the frontend login-state and permission-entry fix release.

### Fixed

- Frontend user login state is restored consistently across header, Go-rendered community pages, and Go-rendered topic pages.
- Community follow, favorites, follows, activities, and notifications pages now send the frontend user token and no longer misreport logged-in users as unauthenticated.
- Normal frontend users no longer see the full admin-next backend entry in the frontend user menu.
- Community moderators see the `/moderator` workspace entry based on `is_moderator`.
- Publishing `question` content now matches the question category instead of defaulting to an article category.
- The admin-next menu now exposes only one community management entry.

### Changed

- `GET /api/v1/auth/me` now includes `is_moderator` and `moderated_communities`.
- `/admin-next/sites` remains as a hidden compatibility route and redirects to `/admin-next/communities`.
- Topic creation validation now prefers `categories.content_type` and falls back to legacy `categories.type`.

## v1.2.0

DevHub v1.2.0 is the tag system enhancement release.

### Added

- Go-rendered Baidu-friendly tag aggregation SEO pages at `/tags/:tag/`.
- Go-rendered community tag aggregation SEO pages at `/c/:communitySlug/tags/:tag/`.
- Public tag detail, tag-topic aggregation, and tag suggestion APIs.
- Public community tag detail and community tag-topic aggregation APIs.
- Tag follow UX on the tag SEO page using existing `POST /api/v1/follows/toggle`.
- Publish-page tag suggestions scoped to the selected community.
- admin-next tag management at `/admin-next/tags`, including CRUD, enable/disable, SEO fields, and related-topic viewing.
- MySQL schema and startup migration support for tag `follower_count`, SEO fields, and `enable/disable` status.
- Dynamic sitemap entries for enabled global tags and enabled community tag pages.

### Changed

- Topic and community tag links now point to canonical tag pages instead of only search filters; community context links use `/c/:communitySlug/tags/:tag/`.
- Tags are first-class manageable records in MemoryStore and MySQLStore, while still preserving existing topic tag behavior.
- `/topics/:id` and `/c/:slug` SEO output remains Go-rendered and unchanged in responsibility.

### Known Limitations

- Tag merge, tag aliasing, and tag trend statistics are still not part of v1.2.0.
- Tag custom redirect/canonical migration after future merges remains planned.
- Sitemap is still a single dynamic file and is not yet sharded.

## v1.1.3

DevHub v1.1.3 is the independent moderator workspace MVP release.

### Added

- Independent frontend moderator workspace at `/moderator`, `/moderator/reports`, `/moderator/topics`, `/moderator/comments`, and `/moderator/audit-logs`.
- Dedicated `/api/v1/moderator/*` APIs that use frontend `users` tokens and `community_moderators` scope checks.
- Moderator dashboard, managed community list, scoped reports, scoped topics, scoped comments, and scoped audit-log views.
- Moderator actions for handling reports, feature/unfeature, pin/unpin, hide/restore topics, lock/unlock comments, and hide/restore comments.
- Moderator audit-log writes with `actor_type=moderator`, `actor_id=users.id`, and community scope.

### Changed

- The frontend user menu now links to the independent moderator workspace.
- Moderator governance no longer needs to enter the full admin-next UI for the MVP workflow.
- MemoryStore frontend registration now creates a real user with a bcrypt password hash, so newly registered accounts can log in with their own password.

### Known Limitations

- Complex RBAC, permission matrix editing, moderator tenure, and performance statistics remain out of scope.
- The moderator workspace is a lightweight runtime API page and not a full replacement for admin-next.
- Super admins should continue to use admin APIs and admin-next for full-system governance.

## v1.1.1

DevHub v1.1.1 is the frontend/admin identity boundary cleanup release.

### Added

- Scoped JWT claims for frontend user tokens and backend admin tokens.
- Separate frontend user login and backend admin login flows.
- Moderator-scoped admin API access for enabled community moderators.
- `actor_type` and `actor_id` audit-log fields for admin, moderator, and system actions.
- Identity-boundary test cases and v1.1.1 release documentation.

### Changed

- `/api/v1/admin/*` now uses backend admin identity by default, with explicit scoped moderator allowance.
- Frontend token storage now prefers `devhub_user_token` and `devhub_user_refresh_token`, while keeping compatibility with old keys.
- Audit log UI now displays and filters actor identity type.

### Known Limitations

- MemoryStore still uses demo seed users for local development.
- MySQL refresh tokens still use `user_id` with `token_type` to distinguish `users.id` from `admin_users.id`.
- Admin-user to frontend-user binding is left for a later productionization pass.

## v1.1.0

DevHub v1.1.0 is the sub-site module enhancement release. It upgrades communities from a simple content filter into independent community spaces with their own profile, SEO, boards, moderators, stats, follow state, and announcements.

### Added

- Enhanced community profile fields: logo, cover image, slogan, theme color, SEO title/description/keywords, counters, hot score, and announcement fields.
- Go-rendered Baidu-friendly community SEO pages for `/c/:slug`, including title, description, canonical, h1, board links, topic links, tag links, stats, moderators, and follow action.
- `/site/:slug` compatibility redirect to canonical `/c/:slug/`.
- Community stats API, public community moderator API, enhanced community tags/categories responses, and community follow counter updates.
- admin-next community management page at `/admin-next/communities`, with create/edit, enable/disable, sort order, SEO fields, announcement fields, frontend links, moderator links, and board management.
- Admin community and category CRUD APIs, reorder APIs, status APIs, and audit-log writes for community/board changes.
- MemoryStore and MySQLStore support for enhanced communities, enhanced categories, community stats, community sitemap filtering, public moderators, and community follow counts.
- Sitemap entries for enabled communities such as `/c/php/`, `/c/go/`, `/c/java/`, `/c/ai/`, and `/c/frontend/`.
- v1.1.0 release documentation and test matrix.

### Changed

- Community pages are now treated as first-class spaces instead of only content-list aliases.
- `/c/:slug` is the canonical community URL; `/site/:slug` remains compatible by redirecting.
- MySQL schema and startup migration helpers now include v1.1.0 community/category fields.
- README, API, SEO, deployment, testing, and project progress docs are updated for the v1.1.0 scope.

### Known Limitations

- v1.1.0 uses enabled categories as the default community navigation; deeper custom navigation is left for a later release.
- Advanced tag features such as aliases, merging, trend statistics, and tag admin remain planned for v1.2.0.
- A complete followed-community feed remains planned for v1.3.0; this release completes follow state, follower count, activities, and "my follows" visibility.
- Comment likes, canceling solved status, recommendation algorithms, reputation, and complex analytics are outside this release.
- Sitemap output is still single-file dynamic output and is not yet sharded for very large installations.

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

- Advanced tag features are planned for v1.2.0, including tag detail SEO pages, tag admin, aliases, merging, and trends.
- Runtime comment likes are not part of v1.0.0.
- Accepted questions support changing the best answer, but do not yet support canceling solved status.
- Tag-follow and user-follow backend support exists, while richer frontend entry points remain future work.
- Sitemap output is dynamic but not yet sharded for very large content volumes.
- Production deployment still needs environment-specific process supervision, reverse proxy, HTTPS, logging, and backup scheduling.
