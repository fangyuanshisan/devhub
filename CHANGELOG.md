# Changelog

## v1.6.0 (2026-05-15)

Current `VERSION` is `v1.6.0`. v1.6.0 closes the plugin package upload and distribution-prep track: zip upload sandbox, upload lifecycle, real Ed25519 signature verification, trusted publishers, read-only remote indexes, version repository, upgrade diff, operation recovery previews, config key rotation, and admin plugin UI grouping are in place. See `docs/releases/v1.6.0.md`.


- Added admin zip plugin package upload into a controlled sandbox: `.zip` only, 20MB upload limit, 50MB extracted limit, 5MB per-file limit, 300-file limit, depth limit, nested archive blocking, zip slip checks, symlink/special-file blocking, and package-root detection.
- Added upload detail and promote APIs. Promote copies validated staging packages into `storage/plugins/packages/{code}` but does not install, enable, execute code, run SQL, or dynamically load frontend assets.
- Added the admin upload panel under `/admin-next/plugins/install`, showing zip scan, package scan, checksum, signature, risk report, manifest validation, dry-run, blocked reasons, suggestions, and promote action.
- Added backend and Playwright coverage for valid zip upload, invalid type, zip slip, dangerous files, checksum/risk reuse, promote, auth rejection, and audit logging.
- Added `plugin_package_uploads` lifecycle records for uploaded zip packages, with list/detail/rescan/import-approval/approve/reject/promote/cancel/delete/cleanup APIs.
- Added `/admin-next/plugins/packages/uploads` upload package management page with lifecycle filters, scan snapshots, action reasons, import approval, promote, cancellation, deletion, and cleanup controls.
- Kept the v1.6 boundary strict: promote is not install; uploaded packages still do not execute plugin code, run SQL, dynamically load frontend assets, or connect to a remote marketplace.
- Added Ed25519 real signature verification for plugin packages and an admin trusted publishers management page/API; signed package dry-run, repository scan, upload detail, promote, install, and approval execution now use verified/trusted/unknown/blocked/revoked signature states in `risk_report`.
- Added plugin config encryption keyring + rotation tooling: `enc:v2:<key_id>:...` ciphertext format, multi-key decryption for legacy `enc:v1`, admin key status endpoint and rotation dry-run/re-encrypt APIs, plus `/admin-next/plugins/config-keys` UI (no key material exposure, no KMS/Vault, no scheduled rotation).
- Added read-only remote plugin indexes, plugin package version repository, structured upgrade diffs, and operation recovery previews; these remain metadata / governance flows only and do not download, install, execute code, run SQL, or dynamically load assets.
- Closed v1.6.0 acceptance with Go tests/build, admin build + full admin E2E (`62 passed`), frontend build + frontend E2E (`17 passed`), SEO curl checks for `/topics/1/` and `/c/php/`, and documented that zip export download/signing packaging is still a v1.7 follow-up.
- Refined the admin plugin governance UI information architecture into six function groups, added compatibility redirects for older plugin routes, extracted lightweight shared plugin display components, added grouped plugin API exports, and replaced the old navigation E2E with `plugin-governance-pages.spec.js`.

## v1.5.0 (2026-05-14)

Current `VERSION` is `v1.5.0`. v1.5.0 closes the plugin package governance track: local package dry-run, checksums and risk reports, local repository scanning, local package install, config version history, sensitive config encryption, approval flow, signing/trusted-source draft, installed-plugin export, and the matching docs/E2E surface are now in place. See `docs/releases/v1.5.0.md`.

- Added the local plugin package governance track: package dry-run, checksum/risk reporting, repository scanning, install closure, config history, sensitive config encryption, approval flow, signing/trusted-source draft, and local plugin export, all with admin-facing UI and tests.
- Kept the package governance boundary strict: no zip uploads, no remote market, no remote install/download, no dynamic loading, no third-party code execution, no external SQL execution, and no user-data export.

## Unreleased

- Added a local plugin package spec draft (`docs/PLUGIN_PACKAGE.md`) and an admin dry-run API for scanning and previewing local plugin packages (`POST /api/v1/admin/plugins/packages/dry-run`); this is a safe read/validate/preview flow only (no install, no code/SQL execution, no dynamic frontend asset loading).
- Added an admin UI section under `/admin-next/plugins/install` for running local plugin package dry-run previews, plus a safe example package at `examples/plugins/demo_notice/`.
- Extended local plugin package dry-run with sha256 `checksums.json` verification, strengthened dangerous-file rules (symlink/executable/hidden dirs), and a backend-generated `risk_report` (low/medium/high/blocked) rendered in the admin UI.
- Added a local plugin package repository scanner (`GET /api/v1/admin/plugins/packages`) with filters/pagination plus a detail endpoint (`GET /api/v1/admin/plugins/packages/detail`), and surfaced the repository view under the admin install/upgrade page.
- Added a minimal local plugin package install flow (`POST /api/v1/admin/plugins/packages/install`) that re-runs dry-run server-side and installs manifest-only plugins as `disabled` with `source_type=local_package` (no code/SQL execution, no dynamic frontend asset loading).
- Added a draft plugin package signature + trusted publisher model: optional `publisher.json`/`signature.json`, local-only `storage/plugins/trusted_publishers.json`, Ed25519 verification over `sha256(raw checksums.json bytes)`, risk-report integration, and admin UI display in the package repository/dry-run/detail/install confirmation flows.
- Added plugin config version history for global/community plugin configs, including redacted diff views, version details, audit linkage via `config_version_id`, and a rollback **dry-run** preview API/UI (no actual rollback write).
- Added backend encryption for sensitive plugin config fields (AES-256-GCM with `enc:v1:` prefix) while keeping API/audit/history outputs redacted; supports placeholder retention (`[ENCRYPTED]`/`******`) and empty-string clearing.
- Added a minimal plugin approval flow for high-risk operations (install/upgrade): approval requests + admin approval center (`/admin-next/plugins/approvals`) + execution with mandatory re-dry-run checks; direct local package install and plugin upgrade execution now require `plugin.approve`.
- Refactored plugin governance backend structure by splitting plugin-related HTTP handlers and service methods into focused modules (no API/behavior changes); updated E2E compose defaults so Playwright runs against an internal `devhub` service (`DEVHUB_E2E_ORIGIN=http://devhub:8090` by default).
- Added installed declarative plugin export to local plugin package directories (`POST /api/v1/admin/plugins/:code/export/dry-run` and `/export`), generating `manifest.json`, README, redacted `config.example.json`, `checksums.json`, optional docs/migration/signature stubs, and an admin detail-drawer export panel; exports never include sensitive config, user data, runtime code, SQL, zip downloads, or remote publishing.

## v1.4.0 (2026-05-12)

Current `VERSION` is `v1.4.0`. The plugin-content governance enhancement work is now validated with Go tests/build plus Docker-based admin build and Playwright; see `docs/releases/v1.4.0.md`.

- Closed v1.4.0 acceptance: full Go + Docker build + admin/frontend Playwright + SEO curl checks pass (admin E2E `35 passed`, frontend E2E `17 passed`), and no long-term `test.skip/test.only` remains in Playwright suites.
- Began `v1.4.0` plugin-content governance enhancement work: `GET /api/v1/admin/posts` now supports precise `plugin_code + content_type` filtering and returns plugin ownership fields for admin post rows.
- Reworked the admin plugin area into function pages under `/admin-next/plugins/overview|list|content|install|config|dependencies|hooks|events|search-index|navigation|permissions|audit|developer`, while keeping legacy plugin routes compatible.
- Upgraded `PluginContent` into a fuller governance page with plugin name/code/status/health/type-count header, disabled/archived history notices, aligned filter/batch layout, result details, audit jump query metadata, and recent-governance entry from the detail drawer.
- Expanded PluginContent batch governance from hide/restore to approve/reject, pin/unpin, and feature/unfeature while keeping structured plugin audit metadata.
- Added minimal PluginContent E2E coverage for archived-history governance, precise plugin filtering, hide/restore regression, and a batch pin/unpin chain.
- Reorganized the documentation system so `docs/README.md` is the canonical entry point, `docs/PROJECT_PROGRESS.md` holds current state, `docs/PLUGIN_SYSTEM_ROADMAP.md` holds long-term and next-version plugin goals, and historical task material is archived under `docs/archive/`.
- Added `docs/releases/v1.3.5.md` as the next-stage draft for plugin-governance UI and install / upgrade wizard closure.
- Archived the old root `更新.md` product task document into `docs/archive/2026-05-09-product-requirements.md`; it is no longer a current acceptance source.
- Began Stage B plugin-governance experience work in the admin UI with `vue-i18n` and a default zh-CN dictionary for plugin-center wording, status labels, config panels, audit labels, and PluginContent actions.
- Completed another plugin-governance i18n cleanup pass for the plugin detail drawer, community plugin config drawer, PluginConfigEditor hints, PluginContent content statuses, and audit action labels; technical values such as `plugin_code`, `content_type`, Hook names, and JSON keys remain visible as raw values where useful.
- Upgraded plugin config editing from JSON-only to a basic schema-driven form mode plus JSON advanced mode, with effective-config preview and config-diff display.
- Added a plugin Hooks troubleshooting view in the admin plugin detail drawer, including hook stats, recent `hook_executions`, a filterable executions drawer with pagination, an execution-detail drawer, and audit-log jump links.
- Added plugin SDK/template documentation plus `go run ./cmd/devhub plugin:new` for generating manifest-only plugin skeletons that validate through the existing ManifestValidator and config schema checks.
- Enhanced the generic PluginContent governance page with content-type filtering, detail drawer, multi-select, batch hide, batch restore, and audit-log entry points while reusing the existing audited backend batch topic API.
- Connected PluginContent audit-log entry points to the generic audit log page with prefilled action, target type, and plugin metadata filters.
- Aligned Stage B documentation wording so basic schema-driven forms, effective config, config diff, PluginContent batch hide/restore, and audit-log jumps are treated as landed baseline capabilities, while deep schema support and advanced batch governance remain future work.
- Added plugin SDK/template documentation under `docs/plugins/`, including manifest, config schema, Hook, migration, permission, menu/route, and external ecosystem guides.
- Added built-in plugin lifecycle response fields and soft-uninstall archive/restore APIs; archived plugins block new content and community enablement while preserving history, config, migrations, audit logs, and SEO.
- Extended plugin status storage for `archived` and `migration_failed`, with matching MySQL schema/migration updates.
- Added frontend and admin E2E coverage for archived-plugin entry linkage: archived plugin content types disappear from publish pages, direct archived `content_type` submissions fail, community enablement is blocked, historical topic SEO remains accessible, and PluginContent can still govern archived historical content.
- Added a manifest validator plus admin validate/dry-run/install APIs for safe manifest + configuration-style plugins; installation records metadata and pending migrations but does not execute third-party code.
- Added plugin health summary APIs and bulk archive/restore APIs for the governance center.
- Extended dynamic content-type permission lookup so manifest-installed plugins can participate in the unified create-permission chain.
- Added a P2 upgrade dry-run preview API and UI branch so admins can inspect current/new version compatibility, changed keys, and diff without performing a real upgrade.
- Added the real plugin upgrade execution API and UI entry, preserving config/migration/audit history while updating version and manifest metadata.
- Documented that external-service webhooks, upgrade flows, plugin package upload/signing, remote marketplace, dynamic loading, sandboxing, hard uninstall, and migration down remain future work.
- Extended the admin plugin governance center UI with health summary cards, manifest validate / dry-run / install entry panels, bulk archive / restore actions, and a clearer runtime/archive-status banner in the plugin detail drawer.
- Moved plugin manifest validate / dry-run / install and upgrade/bulk result panels from inline page sections into drawers, keeping the same actions and test anchors while avoiding the plugin list being pushed into a nested long-scroll layout.
- Reworked `/admin-next/plugins` into a clearer governance layout with a page action header, list/status-governance views, compact summary cards, a filter panel, a bulk-action panel, and a reduced table action column.
- Turned manifest validate / dry-run / install and upgrade preview / execution into drawer-based step flows with structured validation, impact, compatibility, confirmation, and result panels.
- Added bulk archive / restore impact previews, succeeded / failed result tables, and audit-log jump actions in the plugin governance UI.
- Updated the admin plugin-governance E2E suite for the new action grouping and step-flow UI; latest admin check passes with `21 passed / 2 skipped`.
- Re-aligned plugin documentation with current code facts: `v1.3.5` is now treated as an implemented-but-unreleased governance closure, with remaining requirements reorganized into release cleanup, `v1.4` platform enhancement, and later plugin distribution work.
- Began v1.4.0-P1-10 unified plugin governance error codes: plugin endpoints now return `code/message/details/suggestion` while keeping legacy `error`, and permission denials on plugin routes expose `details.permission_code` for actionable fixes.
- Added admin `GET /api/v1/admin/plugins/:code/readiness` and a plugin-detail “操作诊断” tab to explain why enable/upgrade/config actions are blocked, including dependency/core/config_schema/migration checks.
- Enhanced the plugin-detail permissions tab with missing/high-risk highlights, filters, and reference lookup across menus/routes/content types.
- Began v1.4.0-P1-11 frontend entry/menu visibility governance: added `/api/v1/navigation`, `/api/v1/communities/:slug/navigation`, `/api/v1/communities/:slug/create-options`, and admin `/api/v1/admin/plugins/:code/menus/preview` for unified navigation/create gating and actionable “why hidden” diagnostics.

## v1.3.4

DevHub v1.3.4 is the plugin failure-governance and platform-foundation closure release.

- Added E2E/API-only failed plugin migration injection guarded by `DEVHUB_E2E_TESTING=1` or `CMS_STORE=memory`.
- Verified failed plugin migrations block both global plugin enablement and per-community plugin enablement until retry succeeds.
- Added audit coverage for migration failure injection, retry, and success recovery.
- Added admin Playwright coverage for the migration tab failure reason, retry action, restore flow, and plugin audit lookup.
- Added E2E/API-only HookBus failure injection guarded by `DEVHUB_E2E_TESTING=1` or `CMS_STORE=memory`.
- Verified blocking Hook failures block content creation without dirty writes and record `hook_executions` plus `plugin.hook.blocked` audit.
- Verified non-blocking Hook failures keep content creation successful while recording `hook_executions` plus `plugin.hook.failed` audit.
- Added admin Playwright coverage for Hooks tab failure summaries and plugin audit lookup.
- Tightened the plugin permission matrix around `ContentTypeDefinition.create_permission`; `post.create` now remains documented and tested only as a `core.topic.create` compatibility bridge, not as a plugin-content create permission.
- Added API tests proving `post.create` cannot create plugin-owned content, plugin create permissions can create their own content types, and frontend user tokens cannot call plugin governance APIs.
- Ran a dedicated MySQLStore / legacy-database upgrade pass for plugin platform schema, plugin migrations, hook executions, audit logs, global/community plugin state, failed migration readiness, and config schema validation.
- Hardened MySQL plugin upgrade migrations: `004_community_plugins.sql` now tolerates numeric-order execution before `005`, and `005_core_plugins.sql` now adds plugin fields idempotently.
- Added lightweight plugin health status reasons and Hook-derived `hook_warning` / `hook_error` summaries for the admin plugin governance center.
- Expanded plugin audit filtering by plugin code, community, action, actor, target, metadata, request id, and time range.
- Archived the v1.3.4 testing matrix into automated, partially automated, manual, uncovered, and skipped categories, and scoped P1 to plugin experience work rather than new plugin-market capabilities.
- Keep plugin marketplace, package upload/install, remote install/update, Go dynamic loading, and third-party sandboxing out of the current implementation scope.

## v1.3.3

DevHub v1.3.3 is the plugin platform governance closure release.

### Changed

- Added Service-level plugin enable readiness checks for both global and per-community enable actions.
- Plugin enable now checks plugin existence, global config schema validity, enabled dependencies, and failed plugin migrations before allowing `enabled`.
- Kept built-in pending up/no-op migrations non-blocking for enable, while surfacing them through plugin health and the migration tab; failed migrations block enable until retried or resolved.
- Clarified v1.3.3 documentation boundaries for lifecycle states, config schema validation, HookBus observability, migration no-op runner, plugin permissions, `post.create` compatibility, and admin plugin governance center coverage.
- Added `docs/releases/v1.3.3.md` and updated README, docs index, API, architecture, testing, project progress, changelog, and VERSION to the v1.3.3 release line.

### Known Limitations

- Plugin lifecycle states are accepted by schema/Store but still do not form a full automatic state machine.
- Plugin migrations remain built-in up/no-op records; migration down, true rollback, pre-migration backup, and external plugin migration packages remain follow-up work.
- HookBus remains for built-in plugins only; third-party dynamic hooks, webhooks, remote execution, sandboxing, and plugin marketplace capabilities are not implemented.

## v1.3.2

DevHub v1.3.2 is the plugin platform governance enhancement release.

### Changed

- Calibrated plugin-platform documentation to distinguish completed capabilities, partial capabilities, reserved concepts, and future roadmap items before continuing new plugin work.
- Moved HookBus into the plugin platform layer (`internal/plugins`) and registered minimal built-in hook handlers for system plugins.
- Added `hook_executions` runtime records for built-in HookBus execution, plus `/api/v1/admin/plugins/:code/hooks` for Hook statistics and recent executions.
- Recorded blocking Hook failures as `plugin.hook.blocked` and non-blocking Hook failures as `plugin.hook.failed` audit entries.
- Added lightweight plugin health summaries to admin plugin responses and the admin plugin governance UI.
- Added `/api/v1/admin/plugins/:code/audit-logs` for plugin-scoped audit queries.
- Added structured `plugin_code` audit metadata for plugin content governance actions such as hide/restore, pin/feature, comment governance, and batch topic moderation.
- Enforced `config_schema` validation when saving plugin `config_json` (both global `plugins.config_json` and per-community `community_plugins.config_json`).
- Added schema default values to `resolved_config.effective` and recorded plugin config audit diffs via `metadata_json.changed_keys`.
- Added `plugin_migrations` table (schema + migration) for tracking plugin migration execution state.
- Added built-in plugin migration declarations for qa/docs/wiki, plus admin APIs and UI for listing, running, and retrying first-stage up/no-op migrations.
- Recorded plugin migration run/retry/success/failure actions in structured audit logs.
- Enhanced `/admin-next/plugins` towards a plugin governance center baseline UI (stats cards, filter toolbar, clearer status/capability badges).
- Upgraded the admin plugin detail drawer into a tabbed governance view and replaced the global-config textarea with a JSON editor powered by `json-editor-vue` + `Ajv` client-side schema validation.
- Upgraded the admin community plugin drawer with filtering, clearer status/override indicators, and a JSON editor for `community_plugins.config_json` powered by `json-editor-vue` + `Ajv` schema validation.
- Added lightweight plugin impact analysis endpoints and surfaced impact hints in disable confirmations; added an audit tab to the admin plugin detail drawer (backed by `admin/audit-logs`) and improved the generic PluginContent page with community/status filters.
- Extended the plugin status model beyond `enabled` / `disabled` to support governance states such as `config_invalid` and `migration_pending`, while keeping content creation strictly gated on global `enabled` plus community `enabled`.
- Expanded plugin impact analysis counts to include existing contents, enabled/disabled communities, recent contents, pending contents, config overrides, and pending migrations; disable confirmations now surface the richer impact context without implying historical content or SEO deletion.
- Archived a plugin-governance acceptance pass covering Go tests/build, Docker Node admin build, impact APIs, audit logs, config schema failures, global/community plugin state limits, moderator menus, and `/topics/:id` SEO regression.
- Added a fixed Docker-based admin Playwright E2E runner (`admin-e2e`) using `mcr.microsoft.com/playwright:v1.59.1-noble`, with containerized admin build and a minimal plugin-governance browser test suite.
- Added a fixed Docker-based frontend Playwright E2E runner (`frontend-e2e`) with containerized frontend build and a first-stage public navigation / SEO smoke suite.
- Expanded admin Playwright E2E coverage from the plugin-governance center to login, content, comments, communities, tags, and audit-log smoke paths.
- Archived a plugin-system acceptance pass with Go tests/build, admin Docker build, frontend 14-test E2E pass, admin 15-test E2E pass, and `/topics/:id` / `/c/:slug` SEO curl regressions.
- Fixed admin plugin-governance E2E state isolation by restoring globally disabled plugins in `finally` and aligning impact-dialog assertions with the current real impact fields.

### Baseline Notes

- Current plugin runtime state is still based on `plugins.status`, `community_plugins.status`, and `plugin_migrations.status`; the expanded statuses are accepted by schema/Store, but only global `enabled` is publish-enabled until a full lifecycle state machine is implemented.
- `plugin_migrations` now supports built-in up/no-op migration listing, execution records, failed-state retry, audit, and an admin migration tab; migration down, real rollback, pre-migration backup, and external plugin migration packages remain follow-up work.
- HookBus dispatch exists for built-in plugins and now persists execution records/statistics; retry policy, alerting, and external monitoring remain follow-up work.
- Plugin health is a lightweight governance summary, not a Prometheus/Grafana-style monitoring system.
- Plugin config diff is currently top-level `changed_keys`; deep-path diff, version history, rollback, and gray release remain follow-up work.

## Next

- (empty)

## v1.3.1

DevHub v1.3.1 is the plugin-entry hardening and permission-boundary release.

### Changed

- Reframed the complete plugin system as the highest-priority long-term roadmap, split into P0 platform closure, P1 platform enhancements, P2 plugin distribution, and P3 advanced runtime capabilities.
- Sealed `Service.CreatePost` as a legacy/deprecated business entry so normal writes must go through `Service.CreateTopic` and plugin publishing validation.
- Kept `/api/v1/posts` write endpoints deprecated with `410 Gone`; read compatibility remains.
- Hardened `POST /api/v1/admin/posts` with dynamic plugin create permission checks on top of the legacy base `post.create` gate.
- Forbid normal admin content editing from changing site/community, board/category, `content_type`, or `plugin_code`; ownership/type migration must be handled by a future migration-specific workflow.
- Disabled site and board selectors in the admin content edit form to match the backend ownership-change policy.
- Added plugin create permissions to demo site-admin role seeds for MemoryStore and MySQLStore.
- Standardized plugin platform contracts with manifest/content-type/permission/menu/route/hook structure tests.
- Added trusted server-side `ActorContext` injection for topic creation instead of trusting request-body permissions.
- Added writable global `plugins.config_json`, admin config API, config merge view, and public API config scrubbing.
- Expanded the minimal internal HookBus call points to content create/update/delete, comment creation, search, notification, and SEO events.
- Added structured plugin audit fields (`old_value`, `new_value`, `metadata_json`) and writes for plugin status/config/sort governance actions.
- Made the admin plugin page consume manifest-declared admin menu paths instead of hardcoded plugin route maps.
- Improved the admin global plugin page with an explanatory card, status badges, capability summaries, a tabbed plugin-detail drawer, JSON config/schema display, and clearer enable/disable confirmations.
- Improved the admin community plugin drawer with global/community status badges, enablement summaries, disabled-reason hints, schema reference display, JSON formatting/validation, and reliable sort-order updates.
- Added tests for plugin mappings, config JSON validation, public config hiding, plugin audit logs, and moderator plugin-menu scope filtering.

### Known Limitations

- `post.create` remains a compatibility bridge for `core.topic.create`; it is not the long-term primary permission.
- HookBus is still minimal: search/notification/SEO currently dispatch events but do not yet have full plugin business handlers, retry, or unified error logging.
- `plugins.config_json` and `community_plugins.config_json` validate JSON syntax only; `config_schema` enforcement remains follow-up work.
- The improved admin plugin UI still needs a real browser acceptance matrix; there is no automated browser test runner in the repo yet.
- Plugin impact analysis currently provides lightweight count-only endpoints; affected-object detail lists (e.g. impacted category IDs) are still follow-up work.
- Non-plugin historical audit logs may still only have `admin_logs.target` text summaries.
- `project`, `job`, and `ai_work` are plugin-owned but still lack dedicated extension tables and full business workflows.
- Plugin packages, marketplace, remote install/update, and dynamic loading are not implemented in v1.3.1; they are staged as P2/P3 plugin-platform roadmap items rather than permanent exclusions.

## v1.3.0

DevHub v1.3.0 is the Core + Plugins architecture split release.

Current status: v1.3.0 is code-level integrated for built-in `qa`, `docs`, `wiki`, `projects`, `jobs`, and `ai_works` system plugins, global plugin state, per-community plugin state, and publishing validation. Some browser acceptance, plugin-specific product workflows, and fine-grained permission matrices remain follow-up work and are listed below.

### Added

- Built-in plugin registry with `qa`, `docs`, `wiki`, `projects`, `jobs`, and `ai_works` system plugins.
- MySQL `plugins` table with `installed`, `enabled`, and `disabled` states.
- Per-community plugin enablement via the `community_plugins` table.
- `topics.plugin_code` plus `categories.plugin_code` and `categories.allowed_content_types`.
- Plugin-owned tables: `qa_questions`, `qa_answers`, `docs_spaces`, `docs_documents`, `wiki_spaces`, `wiki_pages`, and `wiki_page_versions`.
- Admin plugin APIs and lightweight admin-next plugin management / plugin content entries.
- Public community plugin API, admin community plugin APIs, and moderator plugin menu API.

### Changed

- `question`, `document`, `wiki_page`, `project`, `job`, and `ai_work` are now owned by `qa`, `docs`, `wiki`, `projects`, `jobs`, and `ai_works` plugins rather than hardcoded as Core-only types.
- Topic publishing validates category plugin binding, global plugin status, per-community plugin status, and allowed content types.
- Legacy `doc` / `wiki` request values are normalized to `document` / `wiki_page` for compatibility.
- `project`, `job`, and `ai_work` have plugin ownership, publish validation, permissions, and menus; plugin-specific extension tables and full product workflows remain follow-up work.

### Known Limitations

- Plugin marketplace, package upload, remote update, and dynamic loading are not v1.3.0 implementation scope; they are staged in the longer plugin-platform roadmap.
- Plugin route loading is currently registry metadata plus Core dispatch; dynamic route/runtime loading is a later-stage platform capability.
- Dedicated Docs tree editing UI and Wiki collaboration / rollback UI remain follow-up work.
- Community plugin `config_json` and sort have a minimal admin UI, but still need full browser matrix acceptance and stronger product polish.
- Publishing currently enforces minimal permission-code checks for plugin-owned types. Core-compatible `article` and `news` still use a coarse permission (`core.topic.create`, compatible with legacy `post.create`) and do not yet support fine-grained per-type permission matrices.

## v1.2.1

DevHub v1.2.1 is the tag governance enhancement release.

### Added

- Tag aliases with admin CRUD APIs, alias-based suggestion matching, alias URL resolution, and audit-log writes.
- Tag merge APIs and admin-next merge UI, including topic-tag migration, follow deduplication, merged status, and merged-target tracking.
- Tag statistic recalculation for single-tag and all-tag operations in MemoryStore and MySQLStore.
- MySQL schema support for `tags.merged_to_id`, `tags.hot_score`, and the `tag_aliases` table.

### Changed

- Public tag resolution now normalizes direct slug, alias slug, and merged source tags to the canonical target tag.
- Merged and disabled tags no longer enter sitemap, and alias URLs are not emitted as sitemap entries.
- Tag SEO pages prefer 301 redirects to canonical target URLs for alias and merged-source access.
- admin-next tag management now exposes alias management, merged status, merge target selection, and statistic recalculation.

### Known Limitations

- Tag trend analytics, operator dashboards, and large-scale async recalculation jobs are still out of scope.
- AI-assisted tag recommendations remain planned for a later release.

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
- Advanced tag features such as aliases, merging, and tag admin were planned for v1.2.0 and are now delivered in v1.2.0–v1.2.1; tag trend statistics remains out of scope.
- A complete followed-community feed remains planned for a later release; this release completes follow state, follower count, activities, and "my follows" visibility.
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

- Advanced tag features were planned for v1.2.0; tag detail SEO pages and tag admin landed in v1.2.0, while aliases/merging/recalculation landed in v1.2.1. Tag trend statistics remains out of scope.
- Runtime comment likes are not part of v1.0.0.
- Accepted questions support changing the best answer, but do not yet support canceling solved status.
- Tag-follow and user-follow backend support exists, while richer frontend entry points remain future work.
- Sitemap output is dynamic but not yet sharded for very large content volumes.
- Production deployment still needs environment-specific process supervision, reverse proxy, HTTPS, logging, and backup scheduling.

## v1.4.0-P1-07

### Added

- Plugin dependency checks now support structured `dependencies` with `code`, `version`, `required`, and `reason`, while keeping legacy string dependencies compatible.
- Manifest validate, dry-run, install, upgrade dry-run, upgrade, and enable now share dependency and Core compatibility blocking rules.
- Admin plugin install / upgrade flows and detail drawer now show dependency matrix, Core compatibility, blocking reasons, and dependency diff.
- Added backend dependency/version tests and `web/admin-app/tests/e2e/plugin-dependencies.spec.js`.

### Notes

- Version constraints are intentionally lightweight: exact `x.y.z`, comparison operators, and whitespace-combined ranges only.
- DevHub still does not support automatic dependency installation, plugin marketplace, remote install, dynamic loading, script sandbox, plugin signing, migration down, or hard uninstall.

## v1.6.0-P0-04

### Added

- Added readonly remote plugin index sources, fetch, plugin list/detail APIs, SSRF-safe URL validation, trust/compatibility/risk metadata, and admin UI at `/admin-next/plugins/remote-indexes`.
- Added `docs/PLUGIN_REMOTE_INDEX.md` and `docs/examples/plugin-remote-index.example.json`.

### Notes

- Remote indexes are metadata-only: DevHub does not download remote packages, install remote plugins, execute code, dynamically load assets, or auto-trust remote publishers in this phase.

## v1.6.0-P0-05

### Added

- Added plugin package version repository APIs and admin page to aggregate installed, local package, uploaded package, and remote index versions.
- Added structured upgrade diff previews with grouped `diff_sections`, high-risk/blocking rules, sensitive value redaction, and upgrade approval handoff.

### Notes

- Remote index versions remain readonly metadata and cannot be upgraded directly; DevHub still does not auto-upgrade, download remote packages, execute plugin code, execute SQL, or dynamically load assets.
