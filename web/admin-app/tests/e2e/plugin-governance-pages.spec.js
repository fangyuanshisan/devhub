import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';
import { expectPluginBoundaryNotice, openPluginPage } from './helpers/pluginHelpers.js';

test.describe('plugin governance pages', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('renders grouped plugin navigation and keeps legacy routes working', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await expect(page).toHaveURL(/\/admin-next\/plugins\/overview/);
    await expect(page.getByTestId('plugin-overview-page')).toBeVisible();

    const subNav = page.getByTestId('admin-sub-nav');
    await expect(subNav).toBeVisible();
    for (const group of ['overview', 'manage', 'packages', 'security', 'config', 'runtime', 'logs', 'market']) {
      await expect(subNav.getByTestId(`admin-sub-nav-group-${group}`)).toBeVisible();
    }

    await page.goto('/admin-next/plugins/governance');
    await expect(page).toHaveURL(/\/admin-next\/plugins\/overview/);
    await page.goto('/admin-next/plugins/manifest');
    await expect(page).toHaveURL(/\/admin-next\/plugins\/install/);
    await page.goto('/admin-next/plugins/diagnostics');
    await expect(page).toHaveURL(/\/admin-next\/plugins\/hooks/);
  });

  test('opens implemented plugin function pages from every group', async ({ page }) => {
    await openPluginPage(page, 'list', /\/plugins\/list/, 'admin-plugins-page');
    await expect(page.getByTestId('admin-primary-nav-plugins')).toHaveClass(/active/);
    await expect(page.getByTestId('admin-sub-nav-group-manage')).toHaveClass(/active/);
    await expect(page.getByTestId('admin-sub-nav-list')).toHaveClass(/active/);
    await expect(page.getByTestId('admin-breadcrumb')).toContainText('插件');
    await expect(page.getByTestId('admin-breadcrumb')).toContainText('插件管理');
    await expect(page.getByTestId('admin-breadcrumb')).toContainText('插件列表');

    await openPluginPage(page, 'content', /\/plugins\/content/, 'plugin-content-hub-page');
    await openPluginPage(page, 'permissions', /\/plugins\/permissions/, 'plugin-permissions-page');
    await openPluginPage(page, 'install', /\/plugins\/install/, 'plugin-install-page');
    await openPluginPage(page, 'packageUploads', /\/plugins\/packages\/uploads/, 'plugin-package-upload-lifecycle-page');
    await openPluginPage(page, 'remotePackages', /\/plugins\/packages\/remote/, 'plugin-remote-packages-page');
    await openPluginPage(page, 'versions', /\/plugins\/versions/, 'plugin-versions-page');
    await openPluginPage(page, 'trustedPublishers', /\/plugins\/trusted-publishers/, 'plugin-trusted-publishers-page');
    await openPluginPage(page, 'configKeys', /\/plugins\/config-keys/, 'plugin-config-keys-page');
    await openPluginPage(page, 'approvals', /\/plugins\/approvals/, 'plugin-approvals-page');
    await openPluginPage(page, 'operations', /\/plugins\/operations/, 'plugin-operations-page');
    await openPluginPage(page, 'dependencies', /\/plugins\/dependencies/, 'plugin-dependencies-page');
    await openPluginPage(page, 'hooks', /\/plugins\/hooks/, 'plugin-hooks-page');
    await openPluginPage(page, 'navigation', /\/plugins\/navigation/, 'plugin-navigation-page');
    await openPluginPage(page, 'searchIndex', /\/plugins\/search-index/, 'plugin-search-index-page');
    await openPluginPage(page, 'events', /\/plugins\/events/, 'plugin-events-page');
    await openPluginPage(page, 'remoteIndexes', /\/plugins\/remote-indexes/, 'plugin-remote-indexes-page');
    await openPluginPage(page, 'developer', /\/plugins\/developer/, 'plugin-developer-page');
    await openPluginPage(page, 'audit', /\/plugins\/audit/, 'plugin-audit-page');
  });

  test('shows shared safety boundary and no unavailable marketplace/runtime entries', async ({ page }) => {
    await page.goto('/admin-next/plugins/packages/uploads');
    await expectPluginBoundaryNotice(page);
    await expect(page.getByText('远程市场')).toHaveCount(0);
    await expect(page.getByText('动态加载入口')).toHaveCount(0);
    await expect(page.getByText('第三方代码运行入口')).toHaveCount(0);

    await page.goto('/admin-next/plugins/config-keys');
    await expectPluginBoundaryNotice(page);
  });

  test('legacy plugin content routes still open', async ({ page }) => {
    await page.goto('/admin-next/qa');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.goto('/admin-next/docs');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.goto('/admin-next/wiki');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.goto('/admin-next/projects');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.goto('/admin-next/jobs');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await page.goto('/admin-next/ai-works');
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
  });
});
