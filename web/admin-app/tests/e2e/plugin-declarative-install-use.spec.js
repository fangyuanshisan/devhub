import { expect, test } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { adminHeaders } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

const projectRoot = path.resolve(process.cwd(), '../..');
const fixtureScript = path.join(projectRoot, 'scripts/build-plugin-package-fixtures.sh');
const fixtureDist = path.join(projectRoot, 'scripts/fixtures/plugin-packages/dist');

test.describe('declarative plugin real install to use closure', () => {
  test('official_links fixture installs and governs real declared capabilities', async ({ page, request }) => {
    const suffix = `s15_${Date.now()}`;
    execFileSync(fixtureScript, ['--suffix', suffix], { cwd: projectRoot, stdio: 'inherit' });
    const manifest = JSON.parse(fs.readFileSync(path.join(fixtureDist, `manifest-${suffix}.json`), 'utf8'));
    const fixture = manifest.links;
    const code = fixture.plugin_code;
    const contentType = fixture.content_type;

    const upload = await uploadFixture(request, fixture.zip);
    expect(upload.upload_id).toBeTruthy();
    expect(upload.can_promote).toBeTruthy();
    expect(upload.manifest_validation?.valid).toBeTruthy();
    expect(JSON.stringify(upload)).toContain(contentType);
    expect(JSON.stringify(upload.migration_plan || upload.install_dry_run || upload)).not.toContain('will_execute":true');
    expect(JSON.stringify(upload)).not.toContain('scripts/install.sh');

    const promoted = await promoteUpload(request, upload.upload_id);
    expect(promoted.package_path).toBe(`storage/plugins/packages/${code}`);

    const packageBeforeDryRun = await firstPackage(request, code);
    expect(packageBeforeDryRun.source_upload_id).toBe(upload.upload_id);

    const installWithoutDryRun = await adminPost(request, '/api/v1/admin/plugins/packages/install', {
      path: promoted.package_path,
    });
    expect(installWithoutDryRun.ok()).toBeFalsy();
    await expectErrorCode(installWithoutDryRun, /plugin_package_install_dry_run_required/);

    const dryRun = await dryRunPackage(request, promoted.package_path);
    expect(dryRun.dry_run_id).toBeTruthy();
    expect(dryRun.package.code).toBe(code);
    expect(dryRun.package.version).toBe('1.0.0');
    expect(dryRun.migration_plan || []).toEqual(expect.arrayContaining([
      expect.objectContaining({ path: 'migrations/001_init.sql', source: 'migrations', will_execute: false }),
    ]));
    expect(JSON.stringify(dryRun)).not.toContain('001_schema.sql');

    const installed = await installPackage(request, promoted.package_path, dryRun.dry_run_id, dryRun.risk_report?.level);
    expect(installed.plugin.code).toBe(code);
    expect(installed.plugin.status).toBe('disabled');
    expect(installed.install_result.installed).toBeTruthy();
    expect(installed.install_result.created_migrations).toBe(1);
    expect(installed.install_result.created_permissions).toBeGreaterThanOrEqual(4);
    expect(installed.install_result.created_menus).toBeGreaterThanOrEqual(2);

    let plugin = await getPlugin(request, code);
    expect(plugin.status).toBe('disabled');
    expect(JSON.stringify(plugin.content_type_definitions || plugin.content_types)).toContain(contentType);
    expect(JSON.stringify(plugin.permissions)).toContain(`${code}.link.create`);
    expect(JSON.stringify(plugin.menus)).toContain('友情链接');

    const enabled = await adminPostJSON(request, `/api/v1/admin/plugins/${code}/enable`);
    expect(enabled.status).toBe('enabled');
    plugin = await getPlugin(request, code);
    expect(plugin.status).toBe('enabled');

    const communityEnabled = await adminPostJSON(request, `/api/v1/admin/communities/1/plugins/${code}/enable`);
    expect(communityEnabled.status).toBe('enabled');

    const globalConfig = await adminPutJSON(request, `/api/v1/admin/plugins/${code}/config`, {
      config_json: {
        enabled: true,
        title: 'E2E 友情链接',
        max_links: 12,
        display_position: 'footer',
      },
    });
    expect(globalConfig.resolved_config?.title || globalConfig.config?.title || JSON.stringify(globalConfig)).toContain('E2E 友情链接');

    const communityConfig = await adminPutJSON(request, `/api/v1/admin/communities/1/plugins/${code}/config`, {
      config_json: {
        enabled: true,
        title: 'PHP 子站友情链接',
        max_links: 8,
        display_position: 'sidebar',
      },
    });
    expect(JSON.stringify(communityConfig.resolved_config || communityConfig)).toContain('PHP 子站友情链接');

    const category = await createCategory(request, {
      name: `友情链接 ${suffix}`,
      slug: `links-${suffix.replaceAll('_', '-')}`,
      type: contentType,
      content_type: contentType,
      plugin_code: code,
      allowed_content_types: [contentType],
      description: 'S15 声明型插件验收板块',
      visible: true,
      nav_visible: true,
    });
    expect(category.id).toBeTruthy();
    expect(category.plugin_code).toBe(code);
    expect(category.allowed_content_types).toContain(contentType);

    const adminMenus = await adminGetJSON(request, '/api/v1/admin/plugin-menus');
    expect(JSON.stringify(adminMenus.items || [])).toContain(`${code}.admin.links`);

    const menuPreview = await adminGetJSON(request, `/api/v1/admin/plugins/${code}/menus/preview?community_slug=php&category_id=${category.id}`);
    const frontendMenu = (menuPreview.items || []).find((item) => item.code === `${code}.frontend.links`);
    expect(frontendMenu).toBeTruthy();
    expect(frontendMenu.visible).toBeTruthy();

    const permissions = await adminGetJSON(request, '/api/v1/admin/permissions');
    const permissionGroup = (permissions.items || []).find((item) => item.code === `plugin.${code}`);
    expect(permissionGroup).toBeTruthy();
    expect(permissionGroup.ops).toEqual(expect.arrayContaining([
      `${code}.config.manage`,
      `${code}.link.create`,
      `${code}.link.manage`,
      `${code}.menu.view`,
    ]));

    const createdTopic = await createTopic(request, {
      community_id: 1,
      community_slug: 'php',
      category_id: category.id,
      content_type: contentType,
      plugin_code: code,
      title: `E2E 友情链接 ${suffix}`,
      summary: 'S15 真实声明型插件内容创建摘要',
      content: 'https://example.com 是用于 S15 声明型插件业务闭环验收的友情链接内容。',
    });
    expect(createdTopic.topic.plugin_code).toBe(code);
    expect(createdTopic.topic.content_type).toBe(contentType);

    const disallowed = await createTopicRaw(request, {
      community_id: 1,
      community_slug: 'php',
      category_id: 101,
      content_type: contentType,
      plugin_code: code,
      title: `E2E 不允许板块 ${suffix}`,
      summary: 'S15 不允许板块验收摘要',
      content: '这条内容应被板块绑定校验拦截，不能发布到默认文章板块。',
    });
    expect(disallowed.ok()).toBeFalsy();
    await expectErrorCode(disallowed, /plugin_category_not_supported|当前板块|内容类型/);

    const disabledCommunity = await adminPostJSON(request, `/api/v1/admin/communities/1/plugins/${code}/disable`);
    expect(disabledCommunity.status).toBe('disabled');
    const communityMenuPreview = await adminGetJSON(request, `/api/v1/admin/plugins/${code}/menus/preview?community_slug=php&category_id=${category.id}`);
    const hiddenCommunityMenu = (communityMenuPreview.items || []).find((item) => item.code === `${code}.frontend.links`);
    expect(hiddenCommunityMenu?.visible).toBeFalsy();
    expect(hiddenCommunityMenu?.reason_code).toBe('plugin_community_disabled');
    const communityDisabledCreate = await createTopicRaw(request, {
      community_id: 1,
      community_slug: 'php',
      category_id: category.id,
      content_type: contentType,
      plugin_code: code,
      title: `E2E 子站禁用 ${suffix}`,
      summary: 'S15 子站禁用阻断摘要',
      content: '这条内容应被子站插件禁用状态拦截。',
    });
    expect(communityDisabledCreate.ok()).toBeFalsy();
    await expectErrorCode(communityDisabledCreate, /plugin_community_disabled|当前子站未启用/);

    await adminPostJSON(request, `/api/v1/admin/communities/1/plugins/${code}/enable`);
    const disabledGlobal = await adminPostJSON(request, `/api/v1/admin/plugins/${code}/disable`);
    expect(disabledGlobal.status).toBe('disabled');
    const globalDisabledCreate = await createTopicRaw(request, {
      community_id: 1,
      community_slug: 'php',
      category_id: category.id,
      content_type: contentType,
      plugin_code: code,
      title: `E2E 全局禁用 ${suffix}`,
      summary: 'S15 全局禁用阻断摘要',
      content: '这条内容应被插件全局禁用状态拦截。',
    });
    expect(globalDisabledCreate.ok()).toBeFalsy();
    await expectErrorCode(globalDisabledCreate, /plugin_disabled|插件未启用/);

    await adminPostJSON(request, `/api/v1/admin/plugins/${code}/enable`);
    await adminPostJSON(request, `/api/v1/admin/communities/1/plugins/${code}/enable`);
    const archived = await adminPostJSON(request, `/api/v1/admin/plugins/${code}/archive`);
    expect(archived.status).toBe('archived');
    const archivedCreate = await createTopicRaw(request, {
      community_id: 1,
      community_slug: 'php',
      category_id: category.id,
      content_type: contentType,
      plugin_code: code,
      title: `E2E 归档阻断 ${suffix}`,
      summary: 'S15 归档阻断摘要',
      content: '这条内容应被插件归档状态拦截。',
    });
    expect(archivedCreate.ok()).toBeFalsy();
    await expectErrorCode(archivedCreate, /plugin_disabled|archived|插件未启用|已归档/);

    const historical = await adminGetJSON(request, `/api/v1/topics/${createdTopic.topic.id}`);
    expect(historical.id).toBe(createdTopic.topic.id);
    expect(historical.plugin_code).toBe(code);
    expect(historical.content_type).toBe(contentType);

    const audits = await adminGetJSON(request, `/api/v1/admin/audit-logs?plugin_code=${encodeURIComponent(code)}&page_size=120`);
    const auditText = JSON.stringify(audits);
    expect(auditText).toContain('plugin.package.installed');
    expect(auditText).toContain('plugin_status');
    expect(auditText).toContain('plugin_config');
    expect(auditText).toContain('plugin_archive');

    await seedAdminSession(page);
    await page.goto('/admin-next/plugins/overview?tab=list');
    await expect(page.getByText(code)).toBeVisible();
    await page.goto('/admin-next/plugins/packages?tab=install&workspace_tab=repository');
    await expect(page.getByTestId('plugin-package-repository-workspace')).toBeVisible();
    await page.getByTestId('plugin-package-repo-scan').click();
    await expect(page.getByTestId('plugin-package-repo-table')).toContainText(code);
  });
});

async function uploadFixture(request, zipName) {
  const filePath = path.join(fixtureDist, zipName);
  const response = await request.post('/api/v1/admin/plugins/packages/upload', {
    headers: adminHeaders(),
    multipart: {
      file: {
        name: zipName,
        mimeType: 'application/zip',
        buffer: fs.readFileSync(filePath),
      },
    },
  });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function promoteUpload(request, uploadID) {
  const response = await adminPost(request, `/api/v1/admin/plugins/packages/uploads/${uploadID}/promote`, { force: false });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function firstPackage(request, keyword) {
  const body = await adminGetJSON(request, `/api/v1/admin/plugins/packages?keyword=${encodeURIComponent(keyword)}&page_size=50`);
  const item = (body.items || []).find((pkg) => pkg.code === keyword);
  if (!item) throw new Error(`package not found: ${keyword}`);
  return item;
}

async function dryRunPackage(request, packagePath) {
  const response = await adminPost(request, '/api/v1/admin/plugins/packages/dry-run', { path: packagePath });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function installPackage(request, packagePath, dryRunID, riskLevel) {
  const payload = { path: packagePath, dry_run_id: dryRunID };
  const normalizedRisk = String(riskLevel || '').toLowerCase();
  if (normalizedRisk && normalizedRisk !== 'low') payload.confirm_risk_level = normalizedRisk;
  const response = await adminPost(request, '/api/v1/admin/plugins/packages/install', payload);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function getPlugin(request, code) {
  const body = await adminGetJSON(request, '/api/v1/admin/plugins');
  const item = (body.items || []).find((plugin) => plugin.code === code);
  if (!item) throw new Error(`plugin not found: ${code}`);
  return item;
}

async function createCategory(request, payload) {
  const response = await adminPost(request, '/api/v1/admin/communities/1/categories', payload);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function createTopic(request, payload) {
  const response = await createTopicRaw(request, payload);
  if (!response.ok()) throw new Error(await response.text());
  const body = await response.json();
  return { topic: adminPostToTopic(body) };
}

async function createTopicRaw(request, payload) {
  return adminPost(request, '/api/v1/admin/posts', {
    site: payload.community_slug || 'php',
    board: 'community',
    category_id: payload.category_id,
    content_type: payload.content_type,
    plugin_code: payload.plugin_code,
    title: payload.title,
    summary: payload.summary,
    content: payload.content,
  });
}

async function adminPostJSON(request, url, data = {}) {
  const response = await adminPost(request, url, data);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function adminPutJSON(request, url, data = {}) {
  const response = await request.put(url, {
    headers: { ...adminHeaders(), 'Content-Type': 'application/json' },
    data,
  });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function adminPost(request, url, data = {}) {
  return request.post(url, {
    headers: { ...adminHeaders(), 'Content-Type': 'application/json' },
    data,
  });
}

async function adminGetJSON(request, url) {
  const response = await request.get(url, { headers: adminHeaders() });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function expectErrorCode(response, pattern) {
  const text = await response.text();
  expect(text).toMatch(pattern);
}

function adminPostToTopic(post) {
  return {
    id: post.id,
    plugin_code: post.plugin_code,
    content_type: post.content_type,
    title: post.title,
    summary: post.summary,
    content: post.content,
    category_id: post.category_id,
    community_id: post.community_id,
  };
}
