#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if git -C "${REPO_ROOT}" rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT="$(git -C "${REPO_ROOT}" rev-parse --show-toplevel)"
fi

cd "${REPO_ROOT}"

SCREENSHOT_DIR="${DEVHUB_PLUGIN_IA_SCREENSHOT_DIR:-/workspace/.devhub/screenshots/plugin-ia}"
ORIGIN="${DEVHUB_E2E_ORIGIN:-http://devhub:8090}"
MATRIX_PACKAGE_ROOT="storage/plugins/packages/.matrix"

rm -rf "${MATRIX_PACKAGE_ROOT}"
mkdir -p "${MATRIX_PACKAGE_ROOT}"
cp -R "examples/plugins/official_links" "${MATRIX_PACKAGE_ROOT}/official_links"
cp -R "examples/plugins/official_webhook_notify" "${MATRIX_PACKAGE_ROOT}/official_webhook_notify"

docker compose run --rm \
  -e "DEVHUB_E2E_ORIGIN=${ORIGIN}" \
  -e "DEVHUB_PLUGIN_IA_SCREENSHOT_DIR=${SCREENSHOT_DIR}" \
  admin-e2e node <<'NODE'
const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const origin = process.env.DEVHUB_E2E_ORIGIN || 'http://devhub:8090';
const outDir = process.env.DEVHUB_PLUGIN_IA_SCREENSHOT_DIR || '/workspace/.devhub/screenshots/plugin-ia';
fs.mkdirSync(outDir, { recursive: true });

const adminUser = {
  id: 1,
  username: 'admin',
  nickname: '超级管理员',
  role: '超级管理员',
  role_code: 'super_admin',
  permissions: ['*'],
};

const officialAnnouncementConfig = {
  enabled: true,
  message: '欢迎使用 DevHub 官方公告插件',
  link_text: '查看详情',
  link_url: '/',
  dismissible: false,
};

const officialWebhookExternalServiceConfig = {
  enabled: true,
  endpoint_url: 'http://localhost:19090/hooks/content.after_create',
  health_check_path: '/health',
  timeout_ms: 1000,
  failure_policy: 'warn',
  auth_type: 'bearer',
  token: 's8-browser-matrix-token-value',
  warning_threshold: 1,
  error_threshold: 2,
};

const adminLoginPayload = { account: 'admin', password: 'admin123' };

const domains = [
  ['plugin-overview', '/admin-next/plugins/overview', 'plugin-overview-domain', '插件列表'],
  ['plugin-packages', '/admin-next/plugins/packages', 'plugin-packages-domain', '安装与升级'],
  ['webhook-governance', '/admin-next/plugins/webhooks', 'plugin-webhooks-page', 'Webhook 与回调'],
  ['trusted-publishers', '/admin-next/plugins/publishers', 'plugin-publishers-domain', '安全与发布者'],
  ['admin-logs', '/admin-next/plugins/runtime', 'plugin-runtime-domain', '运行与审计'],
];

const systemDomains = [
  ['system-settings', '/admin-next/system', 'system-security-page-header', '敏感配置与运行安全状态'],
  ['effective-config', '/admin-next/system?tab=effective', 'system-effective-runtime', '当前生效配置'],
  ['secret-center', '/admin-next/system?tab=secret-center', 'system-security-page-header', 'SecretCenter'],
];

const legacyRoutes = [
  ['/admin-next/plugins/list', /\/admin-next\/plugins\/overview.*tab=list/],
  ['/admin-next/plugins/content', /\/admin-next\/plugins\/overview.*tab=content/],
  ['/admin-next/plugins/config', /\/admin-next\/plugins\/overview.*tab=config/],
  ['/admin-next/plugins/navigation', /\/admin-next\/plugins\/overview.*tab=navigation/],
  ['/admin-next/plugins/permissions', /\/admin-next\/plugins\/overview.*tab=permissions/],
  ['/admin-next/plugins/developer', /\/admin-next\/plugins\/overview.*tab=developer/],
  ['/admin-next/plugins/install', /\/admin-next\/plugins\/packages.*tab=install/],
  ['/admin-next/plugins/packages/local', /\/admin-next\/plugins\/packages.*tab=install/],
  ['/admin-next/plugins/packages/install', /\/admin-next\/plugins\/packages.*tab=install/],
  ['/admin-next/plugins/packages/export', /\/admin-next\/plugins\/packages.*tab=install/],
  ['/admin-next/plugins/remote-indexes', /\/admin-next\/plugins\/packages.*tab=remote-indexes/],
  ['/admin-next/plugins/packages/uploads', /\/admin-next\/plugins\/packages.*tab=uploads/],
  ['/admin-next/plugins/package-uploads', /\/admin-next\/plugins\/packages.*tab=uploads/],
  ['/admin-next/plugins/packages/remote', /\/admin-next\/plugins\/packages.*tab=remote-packages/],
  ['/admin-next/plugins/remote-packages', /\/admin-next\/plugins\/packages.*tab=remote-packages/],
  ['/admin-next/plugins/versions', /\/admin-next\/plugins\/packages.*tab=versions/],
  ['/admin-next/plugins/upgrade-diff', /\/admin-next\/plugins\/packages.*tab=versions/],
  ['/admin-next/plugins/dependencies', /\/admin-next\/plugins\/packages.*tab=dependencies/],
  ['/admin-next/plugins/approvals', /\/admin-next\/plugins\/packages.*tab=approvals/],
  ['/admin-next/plugins/events', /\/admin-next\/plugins\/webhooks.*tab=events/],
  ['/admin-next/plugins/webhook-events', /\/admin-next\/plugins\/webhooks.*tab=events/],
  ['/admin-next/plugins/webhook-deliveries', /\/admin-next\/plugins\/webhooks.*tab=deliveries/],
  ['/admin-next/plugins/webhook-retry', /\/admin-next\/plugins\/webhooks.*tab=exceptions/],
  ['/admin-next/plugins/webhook-circuits', /\/admin-next\/plugins\/webhooks.*tab=exceptions/],
  ['/admin-next/plugins/webhook-secrets', /\/admin-next\/plugins\/webhooks.*tab=secrets/],
  ['/admin-next/plugins/callback-tokens', /\/admin-next\/plugins\/webhooks.*tab=callback_tokens/],
  ['/admin-next/plugins/callback-requests', /\/admin-next\/plugins\/webhooks.*tab=callback_requests/],
  ['/admin-next/plugins/webhooks?tab=secrets', /\/admin-next\/plugins\/webhooks.*tab=secrets/],
  ['/admin-next/plugins/webhooks?tab=callback_tokens', /\/admin-next\/plugins\/webhooks.*tab=callback_tokens/],
  ['/admin-next/plugins/trusted-publishers', /\/admin-next\/plugins\/publishers.*tab=list/],
  ['/admin-next/plugins/config-keys', /\/admin-next\/plugins\/publishers.*tab=config-keys/],
  ['/admin-next/plugins/security', /\/admin-next\/plugins\/publishers.*tab=config-keys/],
  ['/admin-next/plugins/operations', /\/admin-next\/plugins\/runtime.*tab=operations/],
  ['/admin-next/plugins/audit', /\/admin-next\/plugins\/runtime.*tab=audit/],
  ['/admin-next/plugins/hooks', /\/admin-next\/plugins\/runtime.*tab=hooks/],
  ['/admin-next/plugins/diagnostics', /\/admin-next\/plugins\/runtime.*tab=hooks/],
  ['/admin-next/plugins/search-index', /\/admin-next\/plugins\/runtime.*tab=search-index/],
  ['/admin-next/system/settings', /\/admin-next\/system(?:\?|$)/],
  ['/admin-next/system/effective-config', /\/admin-next\/system.*tab=effective/],
  ['/admin-next/system/secret-center', /\/admin-next\/system.*tab=secret-center/],
  ['/admin-next/system/secrets', /\/admin-next\/system.*tab=secret-center/],
  ['/admin-next/system/external-service/http-allowlist', /\/admin-next\/system.*tab=external/],
  ['/admin-next/system/http-allowlist', /\/admin-next\/system.*tab=external/],
];

async function apiRequest(method, urlPath, token, body) {
  const response = await fetch(`${origin}${urlPath}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await response.text();
  let data = {};
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { raw: text };
    }
  }
  if (!response.ok) {
    const message = data?.message || data?.error || data?.raw || response.statusText;
    throw new Error(`${method} ${urlPath} 失败：${response.status} ${message}`);
  }
  return data;
}

async function loginAdminForFixture() {
  const session = await apiRequest('POST', '/api/v1/admin/login', '', adminLoginPayload);
  const token = session.access_token || session.accessToken || session.token;
  const refreshToken = session.refresh_token || session.refreshToken || `${token}-refresh`;
  const user = session.user || adminUser;
  if (!token) throw new Error('无法获取后台登录 token，official_announcement fixture 未准备');
  return { token, refreshToken, user };
}

function normalizeItems(payload) {
  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload?.items)) return payload.items;
  if (Array.isArray(payload?.data?.items)) return payload.data.items;
  return [];
}

function parseConfig(value) {
  if (!value) return {};
  if (typeof value === 'string') {
    try {
      return JSON.parse(value);
    } catch {
      return {};
    }
  }
  if (typeof value === 'object') return value;
  return {};
}

function configMatchesFixture(item) {
  const config = parseConfig(item?.config_json || item?.configJSON || item?.resolved_config || item?.resolvedConfig);
  return Object.entries(officialAnnouncementConfig).every(([key, value]) => config?.[key] === value);
}

function assertTextIncludesAny(text, label, candidates) {
  if (!candidates.some((candidate) => text.includes(candidate))) {
    throw new Error(`${label} 缺少文案：${candidates.join(' / ')}`);
  }
}

async function assertNoSensitiveLeak(locatorOrPage, label) {
  const text = await locatorOrPage.locator?.('body').innerText().catch(() => '')
    || await locatorOrPage.innerText?.().catch(() => '')
    || '';
  const forbidden = [
    's8-browser-matrix-token-value',
    'Authorization: Bearer',
    'Bearer s8-browser',
    'encrypted_value',
    'enc:v',
    'token_hash',
    'callback_token_plaintext',
    'webhook_secret_plaintext',
    'secret plaintext',
    'password plaintext',
    'token plaintext',
    'devhub-admin-1',
  ];
  for (const term of forbidden) {
    if (text.includes(term)) throw new Error(`${label} 泄露敏感字段：${term}`);
  }
}

async function takeMatrixScreenshot(page, name, width) {
  await page.screenshot({ path: path.join(outDir, `${width}-${name}.png`), fullPage: true });
}

async function assertNoBrokenText(page, label) {
  const text = await page.locator('body').innerText().catch(() => '');
  for (const term of ['Cannot read properties', 'route not found', 'undefined', 'null', 'No Data']) {
    if (text.includes(term)) throw new Error(`${label} 出现异常占位：${term}`);
  }
}

async function clickTabByName(page, name, waitForTestId) {
  await page.getByRole('tab', { name, exact: true }).click();
  if (waitForTestId) await page.getByTestId(waitForTestId).waitFor({ timeout: 10000 });
  await page.waitForTimeout(250);
}

async function clickButtonByName(page, name) {
  await page.getByRole('button', { name, exact: true }).first().click();
  await page.waitForTimeout(250);
}

async function getDomainHeading(page, testid) {
  return page.getByTestId(testid).locator('h1, h2').first().innerText().catch(() => '');
}

async function ensureOfficialAnnouncementFixture(token) {
  const pluginsPayload = await apiRequest('GET', '/api/v1/admin/plugins', token);
  const plugins = normalizeItems(pluginsPayload);
  const official = plugins.find((item) => item?.code === 'official_announcement' || item?.plugin_code === 'official_announcement');
  if (!official) {
    throw new Error('fixture 前置失败：后台插件列表未返回 official_announcement，请确认内置插件 registry 已加载');
  }
  const shouldEnableGlobal = official.status !== 'enabled';
  if (shouldEnableGlobal) await apiRequest('POST', '/api/v1/admin/plugins/official_announcement/enable', token);
  const shouldUpdateGlobalConfig = !configMatchesFixture(official);
  if (shouldUpdateGlobalConfig) {
    await apiRequest('PUT', '/api/v1/admin/plugins/official_announcement/config', token, { config_json: officialAnnouncementConfig });
  }

  const communitiesPayload = await apiRequest('GET', '/api/v1/admin/communities', token);
  const communities = normalizeItems(communitiesPayload);
  const community = communities.find((item) => item?.slug === 'php') || communities.find((item) => item?.status === 1) || communities[0];
  if (!community?.id) {
    throw new Error('fixture 前置失败：未找到可用子站，无法启用 official_announcement 子站状态');
  }
  const communityPluginsPayload = await apiRequest('GET', `/api/v1/admin/communities/${community.id}/plugins`, token);
  const communityPlugins = normalizeItems(communityPluginsPayload);
  const communityOfficial = communityPlugins.find((item) => item?.code === 'official_announcement' || item?.plugin_code === 'official_announcement');
  if (!communityOfficial) {
    throw new Error(`fixture 前置失败：子站 ${community.slug || community.id} 插件列表未返回 official_announcement`);
  }
  const communityStatus = communityOfficial.community_status || communityOfficial.communityStatus || communityOfficial.status;
  const shouldEnableCommunity = communityStatus !== 'enabled';
  if (shouldEnableCommunity) await apiRequest('POST', `/api/v1/admin/communities/${community.id}/plugins/official_announcement/enable`, token);
  const shouldUpdateCommunityConfig = !configMatchesFixture(communityOfficial);
  if (shouldUpdateCommunityConfig) {
    await apiRequest('PUT', `/api/v1/admin/communities/${community.id}/plugins/official_announcement/config`, token, { config_json: officialAnnouncementConfig });
  }

  return {
    plugin_code: 'official_announcement',
    global_status: 'enabled',
    community_id: community.id,
    community_slug: community.slug || '',
    config: officialAnnouncementConfig,
    idempotent: true,
    changed_global_status: shouldEnableGlobal,
    changed_global_config: shouldUpdateGlobalConfig,
    changed_community_status: shouldEnableCommunity,
    changed_community_config: shouldUpdateCommunityConfig,
  };
}

async function getAdminPlugins(token) {
  const pluginsPayload = await apiRequest('GET', '/api/v1/admin/plugins', token);
  return normalizeItems(pluginsPayload);
}

async function getPluginByCode(token, code) {
  const plugins = await getAdminPlugins(token);
  return plugins.find((item) => item?.code === code || item?.plugin_code === code);
}

async function ensurePackagePluginFixture(token, code, packagePath) {
  let plugin = await getPluginByCode(token, code);
  if (!plugin) {
    const dry = await apiRequest('POST', '/api/v1/admin/plugins/packages/dry-run', token, { path: packagePath });
    if (!dry?.dry_run_id) throw new Error(`${code} package dry-run 未返回 dry_run_id`);
    const risk = String(dry?.risk_report?.level || dry?.status || '').toLowerCase();
    const installPayload = { path: packagePath, dry_run_id: dry.dry_run_id };
    if (risk && risk !== 'low' && risk !== 'passed') installPayload.confirm_risk_level = risk;
    await apiRequest('POST', '/api/v1/admin/plugins/packages/install', token, installPayload);
    plugin = await getPluginByCode(token, code);
  }
  if (!plugin) throw new Error(`${code} 安装后仍未出现在插件列表`);
  if (plugin.status === 'archived') await apiRequest('POST', `/api/v1/admin/plugins/${code}/restore`, token);
  if (plugin.status !== 'enabled') await apiRequest('POST', `/api/v1/admin/plugins/${code}/enable`, token);
  await apiRequest('POST', `/api/v1/admin/communities/1/plugins/${code}/enable`, token).catch(() => null);
  return getPluginByCode(token, code);
}

async function ensureOfficialLinksFixture(token) {
  const plugin = await ensurePackagePluginFixture(token, 'official_links', 'storage/plugins/packages/.matrix/official_links');
  await apiRequest('PUT', '/api/v1/admin/plugins/official_links/config', token, {
    config_json: {
      enabled: true,
      title: 'S6 友情链接',
      display_position: 'footer',
      max_links: 12,
      show_logo: false,
      open_in_new_tab: true,
    },
  }).catch(() => null);
  return plugin;
}

async function ensureOfficialWebhookNotifyFixture(token) {
  const plugin = await ensurePackagePluginFixture(token, 'official_webhook_notify', 'storage/plugins/packages/.matrix/official_webhook_notify');
  const configResult = { token_configured: true, fallback: false, error: '' };
  try {
    await apiRequest('PUT', '/api/v1/admin/plugins/official_webhook_notify/external-service', token, officialWebhookExternalServiceConfig);
  } catch (error) {
    configResult.token_configured = false;
    configResult.fallback = true;
    configResult.error = error.message || String(error);
    await apiRequest('PUT', '/api/v1/admin/plugins/official_webhook_notify/external-service', token, {
      ...officialWebhookExternalServiceConfig,
      auth_type: 'none',
      token: '',
    });
  }
  return { plugin, external_service_config: configResult };
}

async function createMatrixTopic(token) {
  return apiRequest('POST', '/api/v1/admin/posts', token, {
    site: 'php',
    board: 'community',
    category_id: 101,
    content_type: 'article',
    plugin_code: '',
    title: `S6 Webhook 点击矩阵 ${Date.now()}`,
    summary: 'S6 浏览器点击矩阵触发 Webhook',
    content: '用于触发 official_webhook_notify 的 AfterCreateContent 外部服务投递。',
  });
}

async function seedSession(page, token, refreshToken, user = adminUser) {
  await page.addInitScript(({ accessToken, currentRefreshToken, currentUser }) => {
    sessionStorage.setItem('devhub_admin_token', accessToken);
    sessionStorage.setItem('devhub_admin_refresh_token', currentRefreshToken);
    sessionStorage.setItem('devhub_admin_user', JSON.stringify(currentUser));
  }, { accessToken: token, currentRefreshToken: refreshToken, currentUser: user });
  await page.goto(`${origin}/admin-next/`, { waitUntil: 'domcontentloaded' });
}

async function main() {
  const browser = await chromium.launch();
  const context = await browser.newContext({ viewport: { width: 1366, height: 768 } });
  const page = await context.newPage();
  const errors = [];
  page.on('pageerror', (error) => {
    const message = error.message || '';
    if (message.includes('sessionStorage') && message.includes('sandboxed') && message.includes('allow-same-origin')) return;
    errors.push(message);
  });
  page.on('console', (msg) => {
    const text = msg.text();
    if (msg.type() === 'error' && !text.includes('Failed to load resource')) errors.push(text);
  });

  const { token, refreshToken, user } = await loginAdminForFixture();
  const officialFixture = await ensureOfficialAnnouncementFixture(token);
  const officialLinksFixture = await ensureOfficialLinksFixture(token);
  const officialWebhookFixture = await ensureOfficialWebhookNotifyFixture(token);
  const triggerTopic = await createMatrixTopic(token);
  await seedSession(page, token, refreshToken, user);

  const report = {
    origin,
    fixture: officialFixture,
    officialPlugins: {
      official_links: { status: officialLinksFixture?.status || 'enabled' },
      official_announcement: { status: officialFixture?.global_status || 'enabled' },
      official_webhook_notify: {
        status: officialWebhookFixture?.plugin?.status || 'enabled',
        external_service_config: officialWebhookFixture?.external_service_config || {},
      },
    },
    triggerTopic,
    domains: [],
    legacyRoutes: [],
    detailDrawer: [],
    sensitiveScan: [],
  };
  for (const [key, route, testid, title] of domains) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    await page.getByTestId(testid).waitFor({ timeout: 10000 });
    const pageTitle = await getDomainHeading(page, testid);
    const breadcrumb = await page.getByTestId('admin-breadcrumb').innerText().catch(() => '');
    const hasBlank = await page.getByText('No Data').count();
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth);
    await takeMatrixScreenshot(page, key, 1366);
    if (!pageTitle || !pageTitle.includes(title)) throw new Error(`${route} 页面标题异常：${pageTitle}`);
    if (!breadcrumb || !breadcrumb.includes('插件')) throw new Error(`${route} 面包屑异常：${breadcrumb}`);
    if (hasBlank) throw new Error(`${route} 仍存在 No Data 英文空状态`);
    if (overflow) throw new Error(`${route} 1366 宽度存在横向溢出`);
    await assertNoBrokenText(page, route);
    await assertNoSensitiveLeak(page, route);
    report.domains.push({ key, route, title: pageTitle, breadcrumb });
  }

  await page.setViewportSize({ width: 1024, height: 768 });
  for (const [key, route, testid, title] of domains) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    await page.getByTestId(testid).waitFor({ timeout: 10000 });
    await takeMatrixScreenshot(page, key, 1024);
    const pageTitle = await getDomainHeading(page, testid);
    if (!pageTitle || !pageTitle.includes(title)) throw new Error(`${route} 1024 页面标题异常：${pageTitle}`);
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth);
    if (overflow) throw new Error(`${route} 1024 宽度存在横向溢出`);
    await assertNoBrokenText(page, `${route} 1024`);
    await assertNoSensitiveLeak(page, `${route} 1024`);
  }
  await page.setViewportSize({ width: 1366, height: 768 });

  for (const [key, route, testid, title] of systemDomains) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    await page.getByTestId(testid).waitFor({ timeout: 10000 });
    await page.getByText(title, { exact: false }).first().waitFor({ timeout: 10000 });
    await takeMatrixScreenshot(page, key, 1366);
    await assertNoBrokenText(page, route);
    await assertNoSensitiveLeak(page, route);
    report.domains.push({ key, route, title });
  }
  await page.setViewportSize({ width: 1024, height: 768 });
  for (const [key, route, testid, title] of systemDomains) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    await page.getByTestId(testid).waitFor({ timeout: 10000 });
    await page.getByText(title, { exact: false }).first().waitFor({ timeout: 10000 });
    await takeMatrixScreenshot(page, key, 1024);
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth);
    if (overflow) throw new Error(`${route} 1024 宽度存在横向溢出`);
    await assertNoBrokenText(page, `${route} 1024`);
    await assertNoSensitiveLeak(page, `${route} 1024`);
  }
  await page.setViewportSize({ width: 1366, height: 768 });

  await page.goto(`${origin}/admin-next/system?tab=external`, { waitUntil: 'networkidle' });
  await page.getByText('external_service HTTP Allowlist 管理').waitFor({ timeout: 10000 });
  await page.getByRole('button', { name: '新增 Origin' }).click();
  await page.getByText('新增 HTTP Allowlist Origin').waitFor({ timeout: 10000 });
  await page.getByPlaceholder('http://172.17.0.1:18081').fill('http://0.0.0.0:18081');
  await page.getByPlaceholder('例如：本地 Docker receiver 联调').fill('S8 点击矩阵非法 origin 验证');
  await page.getByText('我确认该 HTTP origin').click();
  await page.getByRole('button', { name: '确认新增' }).click();
  await page.locator('.el-message, .el-message__content')
    .filter({ hasText: /0\.0\.0\.0|wildcard|CIDR|path|query|fragment|userinfo|不允许|非法/ })
    .first()
    .waitFor({ timeout: 10000 });
  await page.keyboard.press('Escape');
  await assertNoSensitiveLeak(page, 'HTTP Allowlist 新增校验');

  await page.goto(`${origin}/admin-next/plugins/overview?tab=list`, { waitUntil: 'networkidle' });
  await page.getByTestId('admin-plugins-page').waitFor({ timeout: 10000 });
  await page.getByTestId('plugin-search').fill('official_announcement');
  await page.getByTestId('plugin-search').fill('official_links');
  await page.getByTestId('plugin-search').fill('official_webhook_notify');
  await page.getByTestId('plugin-filter-reset').click();
  await page.getByTestId('plugin-filter-refresh').click();
  await page.getByTestId('plugin-filter-advanced-toggle').click();
  await page.getByTestId('plugin-filter-advanced-toggle').click();
  await assertNoBrokenText(page, '插件列表按钮');

  await page.goto(`${origin}/admin-next/plugins/packages`, { waitUntil: 'networkidle' });
  await page.getByTestId('plugin-packages-domain').waitFor({ timeout: 10000 });
  await clickButtonByName(page, '查看上传包');
  await page.getByTestId('plugin-package-upload-lifecycle-page').waitFor({ timeout: 10000 });
  await clickTabByName(page, '本地包与预检', 'plugin-install-page');
  await page.getByTestId('plugin-manifest-validate').click();
  await page.getByTestId('plugin-manifest-panel').waitFor({ timeout: 10000 });
  await page.keyboard.press('Escape');
  await page.getByTestId('plugin-package-dry-run').click();
  await page.waitForTimeout(250);
  await page.getByTestId('plugin-package-repo-refresh').click();
  await assertNoBrokenText(page, '安装与升级按钮');
  await assertNoSensitiveLeak(page, '安装与升级按钮');
  if (await page.getByTestId('plugin-package-repo-dryrun-btn').count()) {
    await page.getByTestId('plugin-package-repo-dryrun-btn').first().click();
    await page.waitForTimeout(250);
  }
  if (await page.getByTestId('plugin-package-repo-detail-btn').count()) {
    await page.getByTestId('plugin-package-repo-detail-btn').first().click();
    await page.getByTestId('plugin-package-repo-detail-dialog').waitFor({ timeout: 10000 });
    await page.keyboard.press('Escape');
  }

  await page.goto(`${origin}/admin-next/plugins/webhooks`, { waitUntil: 'networkidle' });
  await page.getByTestId('plugin-webhooks-page').waitFor({ timeout: 10000 });
  await clickTabByName(page, '总览');
  await clickButtonByName(page, '查看投递记录');
  await page.getByTestId('webhook-deliveries-table').waitFor({ timeout: 10000 });
  await page.getByTestId('webhook-deliveries-search').click();
  await clickTabByName(page, '高级治理');
  await clickButtonByName(page, 'Webhook 密钥');
  await page.getByTestId('webhook-secrets-table').waitFor({ timeout: 10000 });
  await clickTabByName(page, '回调 Token', 'callback-tokens-table');
  await clickTabByName(page, '回调请求', 'callback-requests-table');
  await clickTabByName(page, '外部服务执行', 'external-service-executions-table');
  await page.getByTestId('external-service-plugin-select').click();
  await page.keyboard.type('official_webhook_notify');
  await page.keyboard.press('Enter');
  await page.getByTestId('external-service-executions-search').click();
  await page.getByTestId('external-service-executions-table').waitFor({ timeout: 10000 });
  await page.waitForTimeout(1000);
  const retryButton = page.getByTestId('external-service-retry-button').first();
  if (await retryButton.count()) {
    await retryButton.click();
    await page.waitForTimeout(500);
  }
  await assertNoBrokenText(page, 'Webhook 与回调按钮');
  await assertNoSensitiveLeak(page, 'Webhook 与回调按钮');

  await page.goto(`${origin}/admin-next/plugins/publishers`, { waitUntil: 'networkidle' });
  await page.getByTestId('plugin-publishers-domain').waitFor({ timeout: 10000 });
  await clickTabByName(page, '高级治理');
  await clickButtonByName(page, '影响分析');
  if (!page.url().includes('tab=impact')) throw new Error(`发布者影响分析 Tab 跳转异常：${page.url()}`);
  await page.getByText('影响分析').first().waitFor({ timeout: 10000 });
  await assertNoBrokenText(page, '安全与发布者按钮');
  await assertNoSensitiveLeak(page, '安全与发布者按钮');

  await page.goto(`${origin}/admin-next/plugins/runtime`, { waitUntil: 'networkidle' });
  await page.getByTestId('plugin-runtime-domain').waitFor({ timeout: 10000 });
  await clickTabByName(page, 'Hook 排障', 'plugin-hooks-page');
  await clickTabByName(page, '审计日志', 'plugin-audit-page');
  await clickTabByName(page, '操作历史', 'plugin-operations-page');
  await assertNoBrokenText(page, '运行与审计按钮');
  await assertNoSensitiveLeak(page, '运行与审计按钮');

  await page.setViewportSize({ width: 1024, height: 768 });
  await page.goto(`${origin}/admin-next/plugins/overview?tab=list`, { waitUntil: 'networkidle' });
  await page.getByTestId('admin-plugins-page').waitFor({ timeout: 10000 });
  await page.screenshot({ path: path.join(outDir, 'overview-list-1024.png'), fullPage: true });
  const overflow1024 = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth);
  if (overflow1024) throw new Error('插件列表 1024 宽度存在横向溢出');
  await assertNoSensitiveLeak(page, '插件列表 1024');

  async function openDetailByCode(code, label) {
    await page.goto(`${origin}/admin-next/plugins/overview?tab=list`, { waitUntil: 'networkidle' });
    await page.getByTestId('admin-plugins-page').waitFor({ timeout: 10000 });
    if (code) {
      await page.getByTestId('plugin-search').fill(code);
      await page.waitForTimeout(500);
    }
    const button = page.getByTestId(`plugin-detail-${code}`);
    if (await button.count()) {
      await button.first().click();
    } else if (label === 'official-announcement') {
      throw new Error('fixture 已准备，但插件列表仍未显示 official_announcement');
    } else {
      await page.getByRole('button', { name: '查看详情' }).first().click();
    }
    await page.getByTestId('plugin-detail-drawer').waitFor({ timeout: 10000 });
    await page.getByTestId('plugin-detail-tabs').waitFor({ timeout: 10000 });
    await page.screenshot({ path: path.join(outDir, `${label}-detail-1024-overview.png`), fullPage: true });
    const drawerText = await page.getByTestId('plugin-detail-drawer').innerText();
    if (!drawerText.includes('概览')) throw new Error(`${label} 详情抽屉未显示概览 Tab`);
    if (drawerText.includes('No Data')) throw new Error(`${label} 详情抽屉存在 No Data 英文空状态`);
    report.detailDrawer.push({ label, code, viewport: 1024, tab: 'overview' });
    return true;
  }

  async function clickDetailTab(name, screenshotName) {
    await page.getByTestId('plugin-detail-drawer').getByRole('tab', { name, exact: true }).click();
    await page.waitForTimeout(300);
    await page.screenshot({ path: path.join(outDir, screenshotName), fullPage: true });
  }

  async function assertOfficialDetail(code, label, expectedTextList = []) {
    const opened = await openDetailByCode(code, label);
    if (!opened) throw new Error(`${code} 详情未能打开`);
    const drawer = page.getByTestId('plugin-detail-drawer');
    const initialText = await drawer.innerText();
    for (const expected of expectedTextList) {
      if (!initialText.includes(expected)) throw new Error(`${label} 详情缺少文案：${expected}`);
    }
    await clickDetailTab('配置', `${label}-detail-1024-config.png`);
    await clickDetailTab('能力', `${label}-detail-1024-capabilities.png`);
    await clickDetailTab('运行记录', `${label}-detail-1024-runtime.png`);
    await clickDetailTab('技术详情', `${label}-detail-1024-technical.png`);
    await assertNoSensitiveLeak(drawer, `${label} 详情`);
    await page.keyboard.press('Escape');
  }

  if (!await openDetailByCode('docs', 'normal-plugin')) throw new Error('普通插件详情未能打开');
  await clickDetailTab('能力', 'normal-plugin-detail-1024-capabilities.png');
  await clickDetailTab('运行记录', 'normal-plugin-detail-1024-runtime.png');
  await clickDetailTab('技术详情', 'normal-plugin-detail-1024-technical.png');
  const technicalOverflow = await page.evaluate(() => document.querySelector('[data-testid="plugin-detail-drawer"]')?.scrollWidth > document.querySelector('[data-testid="plugin-detail-drawer"]')?.clientWidth);
  if (technicalOverflow) throw new Error('技术详情在 1024 宽度下撑破详情抽屉');
  await clickDetailTab('配置', 'normal-plugin-detail-1024-config.png');
  const versionsButton = page.getByTestId('plugin-config-versions-open');
  if (await versionsButton.count()) {
    await versionsButton.first().click();
    await page.getByTestId('plugin-config-versions-dialog').waitFor({ timeout: 10000 });
    await page.screenshot({ path: path.join(outDir, 'normal-plugin-config-versions-1024.png'), fullPage: true });
    const dialogOverflow = await page.evaluate(() => document.querySelector('.plugin-config-versions-dialog')?.scrollWidth > document.querySelector('.plugin-config-versions-dialog')?.clientWidth);
    if (dialogOverflow) throw new Error('配置版本弹窗在 1024 宽度下存在横向溢出');
    await page.keyboard.press('Escape');
  }
  await page.keyboard.press('Escape');

  await assertOfficialDetail('official_links', 'official-links', ['official_links']);

  await page.goto(`${origin}/admin-next/official-links`, { waitUntil: 'networkidle' });
  await page.getByTestId('plugin-content-page').waitFor({ timeout: 10000 });
  await page.getByTestId('plugin-content-query').click();
  await page.getByTestId('plugin-content-community-filter').click();
  await page.keyboard.press('Escape');
  await page.getByTestId('plugin-content-status-filter').click();
  await page.keyboard.press('Escape');
  await page.screenshot({ path: path.join(outDir, 'official-links-management-1024.png'), fullPage: true });
  await assertNoBrokenText(page, 'official_links 管理页');
  await assertNoSensitiveLeak(page, 'official_links 管理页');

  const officialOpened1024 = await openDetailByCode('official_announcement', 'official-announcement');
  if (!officialOpened1024) throw new Error('official_announcement 详情未能打开');
  const officialDrawerText = await page.getByTestId('plugin-detail-drawer').innerText();
  assertTextIncludesAny(officialDrawerText, 'official_announcement 详情', ['官方公告插件']);
  assertTextIncludesAny(officialDrawerText, 'official_announcement 详情', ['官方内置插件', '官方内置']);
  assertTextIncludesAny(officialDrawerText, 'official_announcement 详情', ['配置 / 前端挂载 / 公告预览', '公告配置']);
  await clickDetailTab('公告配置', 'official-announcement-detail-1024-config.png');
  await clickDetailTab('能力', 'official-announcement-detail-1024-mount.png');
  const mountText = await page.getByTestId('plugin-detail-drawer').innerText();
  assertTextIncludesAny(mountText, 'official_announcement 前端挂载', ['不允许远程 iframe URL', '远程 URL：否', '远程 URL']);
  await clickDetailTab('公告预览', 'official-announcement-detail-1024-preview.png');
  const previewFrame = page.locator('.official-announcement-preview iframe, iframe[src*="official-announcement"]').first();
  await previewFrame.waitFor({ state: 'attached', timeout: 15000 });
  const frameInfo = await previewFrame.evaluate((iframe) => ({
    src: iframe.getAttribute('src') || '',
    sandbox: iframe.getAttribute('sandbox') || '',
    outerHTML: iframe.outerHTML,
  }));
  if (!frameInfo.src.includes('/plugins/official-announcement/iframe')) throw new Error(`公告预览 iframe 路由异常：${frameInfo.src}`);
  if (/^https?:\/\//i.test(frameInfo.src)) throw new Error(`公告预览不应使用远程 iframe URL：${frameInfo.src}`);
  if (frameInfo.sandbox !== 'allow-scripts') throw new Error(`公告预览 sandbox 异常：${frameInfo.sandbox}`);
  for (const term of [/callback_token/i, /webhook_secret/i, /token_hash/i, /authorization/i, /secret=/i, /token=/i]) {
    if (term.test(frameInfo.outerHTML)) throw new Error(`official_announcement iframe DOM 暴露敏感字段：${term}`);
  }
  report.detailDrawer.push({ label: 'official-announcement', code: 'official_announcement', forced: true, tabs: ['公告配置', '前端挂载', '公告预览'] });
  await page.keyboard.press('Escape');

  const webhookOpened = await openDetailByCode('official_webhook_notify', 'official-webhook-notify');
  if (!webhookOpened) throw new Error('official_webhook_notify 详情未能打开');
  await clickDetailTab('运行记录', 'official-webhook-notify-detail-1024-runtime.png');
  await page.getByTestId('plugin-external-service-form').waitFor({ timeout: 10000 });
  const tokenInput = page.getByTestId('plugin-external-service-token');
  if (await tokenInput.count()) {
    const tokenValue = await tokenInput.locator('input,textarea').first().inputValue().catch(() => '');
    if (tokenValue.includes('s8-browser-matrix-token-value')) throw new Error('external_service token 在配置表单中被回显');
  }
  await assertNoSensitiveLeak(page.getByTestId('plugin-detail-drawer'), 'official_webhook_notify 运行记录');
  await clickDetailTab('技术详情', 'official-webhook-notify-detail-1024-technical.png');
  await page.keyboard.press('Escape');

  await page.setViewportSize({ width: 1366, height: 768 });
  const officialOpened1366 = await openDetailByCode('official_announcement', 'official-announcement');
  if (!officialOpened1366) throw new Error('official_announcement 1366 详情未能打开');
  await page.screenshot({ path: path.join(outDir, 'official-announcement-detail-1366-overview.png'), fullPage: true });
  const drawerOverflow1366 = await page.evaluate(() => document.querySelector('[data-testid="plugin-detail-drawer"]')?.scrollWidth > document.querySelector('[data-testid="plugin-detail-drawer"]')?.clientWidth);
  if (drawerOverflow1366) throw new Error('详情抽屉在 1366 宽度下存在横向溢出');
  await page.keyboard.press('Escape');

  for (const [route, expected] of legacyRoutes) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    if (!expected.test(page.url())) throw new Error(`${route} 未跳转到预期 Tab，实际：${page.url()}`);
    await page.reload({ waitUntil: 'networkidle' });
    if (!expected.test(page.url())) throw new Error(`${route} 刷新后 Tab 未保持，实际：${page.url()}`);
    await assertNoBrokenText(page, `旧路由 ${route}`);
    await assertNoSensitiveLeak(page, `旧路由 ${route}`);
    report.legacyRoutes.push({ route, url: page.url() });
  }

  await page.goto(`${origin}/admin-next/plugins/package-uploads?upload_id=s8-matrix#upload-detail`, { waitUntil: 'networkidle' });
  if (!/\/admin-next\/plugins\/packages.*tab=uploads/.test(page.url())) throw new Error(`带 query/hash 的上传旧路由未跳转到上传 Tab：${page.url()}`);
  if (!page.url().includes('upload_id=s8-matrix') || !page.url().includes('#upload-detail')) throw new Error(`上传旧路由未保留 query/hash：${page.url()}`);
  report.legacyRoutes.push({ route: '/admin-next/plugins/package-uploads?upload_id=s8-matrix#upload-detail', url: page.url(), preserved_query_hash: true });

  await page.goto(`${origin}/admin-next/plugins/packages?tab=install`, { waitUntil: 'networkidle' });
  await page.getByTestId('plugin-packages-domain').waitFor({ timeout: 10000 });
  await page.getByTestId('admin-sub-nav-webhooks').click();
  await page.getByTestId('plugin-webhooks-page').waitFor({ timeout: 10000 });
  const historyForwardTarget = page.url();
  await page.goBack({ waitUntil: 'networkidle' });
  if (!/\/admin-next\/plugins\/packages/.test(page.url())) throw new Error(`浏览器后退 Tab 状态异常：${page.url()}`);
  await page.goForward({ waitUntil: 'networkidle' });
  if (page.url() !== historyForwardTarget) throw new Error(`浏览器前进 Tab 状态异常：${page.url()} != ${historyForwardTarget}`);
  await page.screenshot({ path: path.join(outDir, 'legacy-route-webhook-callback-tokens-1366.png'), fullPage: true });
  await assertNoSensitiveLeak(page, '旧路由跳转');

  if (errors.length) throw new Error(`页面错误：${errors.join(' | ')}`);
  fs.writeFileSync(path.join(outDir, 'report.json'), JSON.stringify(report, null, 2));
  await browser.close();
  console.log(`插件 IA 回归通过，截图目录：${outDir}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
NODE
