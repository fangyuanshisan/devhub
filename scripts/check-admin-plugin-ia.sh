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

const domains = [
  ['overview', '/admin-next/plugins/overview', 'plugin-overview-domain', '插件总览'],
  ['packages', '/admin-next/plugins/packages', 'plugin-packages-domain', '插件包治理'],
  ['webhooks', '/admin-next/plugins/webhooks', 'plugin-webhooks-page', 'Webhook 治理'],
  ['publishers', '/admin-next/plugins/publishers', 'plugin-publishers-domain', '发布者与信任'],
  ['runtime', '/admin-next/plugins/runtime', 'plugin-runtime-domain', '运行记录 / 审计'],
];

const legacyRoutes = [
  ['/admin-next/plugins/remote-indexes', /\/admin-next\/plugins\/packages.*tab=remote-indexes/],
  ['/admin-next/plugins/packages/uploads', /\/admin-next\/plugins\/packages.*tab=uploads/],
  ['/admin-next/plugins/events', /\/admin-next\/plugins\/webhooks.*tab=events/],
  ['/admin-next/plugins/webhooks?tab=secrets', /\/admin-next\/plugins\/webhooks.*tab=secrets/],
  ['/admin-next/plugins/webhooks?tab=callback_tokens', /\/admin-next\/plugins\/webhooks.*tab=callback_tokens/],
  ['/admin-next/plugins/audit', /\/admin-next\/plugins\/runtime.*tab=audit/],
  ['/admin-next/plugins/hooks', /\/admin-next\/plugins\/runtime.*tab=hooks/],
];

async function seedSession(page) {
  await page.goto(`${origin}/admin-next/login`, { waitUntil: 'domcontentloaded' });
  await page.evaluate((currentUser) => {
    sessionStorage.setItem('devhub_admin_token', `devhub-admin-${currentUser.id || 1}`);
    sessionStorage.setItem('devhub_admin_refresh_token', `devhub-admin-${currentUser.id || 1}-refresh`);
    sessionStorage.setItem('devhub_admin_user', JSON.stringify(currentUser));
  }, adminUser);
}

async function main() {
  const browser = await chromium.launch();
  const context = await browser.newContext({ viewport: { width: 1366, height: 768 } });
  const page = await context.newPage();
  const errors = [];
  page.on('pageerror', (error) => errors.push(error.message));
  page.on('console', (msg) => {
    const text = msg.text();
    if (msg.type() === 'error' && !text.includes('Failed to load resource')) errors.push(text);
  });

  await seedSession(page);

  const report = { origin, domains: [], legacyRoutes: [] };
  for (const [key, route, testid, title] of domains) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    await page.getByTestId(testid).waitFor({ timeout: 10000 });
    const pageTitle = await page.locator('h2').first().innerText().catch(() => '');
    const breadcrumb = await page.getByTestId('admin-breadcrumb').innerText().catch(() => '');
    const hasBlank = await page.getByText('No Data').count();
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth);
    await page.screenshot({ path: path.join(outDir, `${key}-1366.png`), fullPage: true });
    if (!pageTitle || !pageTitle.includes(title)) throw new Error(`${route} 页面标题异常：${pageTitle}`);
    if (!breadcrumb || !breadcrumb.includes('插件')) throw new Error(`${route} 面包屑异常：${breadcrumb}`);
    if (hasBlank) throw new Error(`${route} 仍存在 No Data 英文空状态`);
    if (overflow) throw new Error(`${route} 1366 宽度存在横向溢出`);
    report.domains.push({ key, route, title: pageTitle, breadcrumb });
  }

  await page.setViewportSize({ width: 1024, height: 768 });
  await page.goto(`${origin}/admin-next/plugins/overview?tab=list`, { waitUntil: 'networkidle' });
  await page.getByTestId('admin-plugins-page').waitFor({ timeout: 10000 });
  await page.screenshot({ path: path.join(outDir, 'overview-list-1024.png'), fullPage: true });
  const overflow1024 = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth);
  if (overflow1024) throw new Error('插件列表 1024 宽度存在横向溢出');

  for (const [route, expected] of legacyRoutes) {
    await page.goto(`${origin}${route}`, { waitUntil: 'networkidle' });
    if (!expected.test(page.url())) throw new Error(`${route} 未跳转到预期 Tab，实际：${page.url()}`);
    report.legacyRoutes.push({ route, url: page.url() });
  }

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
