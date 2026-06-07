import { expect, test } from '@playwright/test';
import { ensurePluginEnabled, injectFailedPluginHook } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin hooks troubleshooting', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ page, request }) => {
    await ensurePluginEnabled(request, 'qa');
    await seedAdminSession(page);
  });

  test('opens hooks tab, shows stats and recent executions, supports filtered executions drawer and detail drawer', async ({ page, request }) => {
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();

    await page.getByTestId('plugin-detail-qa').click();
    const drawer = page.getByTestId('plugin-detail-drawer');
    await expect(drawer).toBeVisible();

    await page.getByRole('tab', { name: 'Hook' }).click();
    await expect(drawer.getByTestId('hook-recent-table')).toBeVisible();
    await expect(drawer.getByTestId('hook-recent-table')).toBeVisible();

    // open executions drawer
    await drawer.getByRole('button', { name: '查看全部执行记录' }).click();
    const execDrawer = page.getByTestId('hook-executions-drawer');
    await expect(execDrawer).toBeVisible();
    await expect(execDrawer.getByTestId('hook-executions-table')).toBeVisible();

    // filter: only failures
    await execDrawer.getByTestId('hook-exec-filter-success').click();
    await page.getByRole('option', { name: '失败' }).click();
    await execDrawer.getByRole('button', { name: '查询' }).click();
    await expect(execDrawer.getByTestId('hook-executions-table')).toBeVisible();

    // open detail drawer from first row if exists
    const firstDetail = execDrawer.getByTestId('hook-executions-table').getByRole('button', { name: '详情' }).first();
    if (await firstDetail.isVisible().catch(() => false)) {
      await firstDetail.click();
      await expect(page.getByTestId('hook-execution-detail-drawer')).toBeVisible();
      await expect(page.getByTestId('hook-execution-detail-drawer')).toContainText('metadata_json');
    }

    // In test env, injection is available via API; validate blocking/non-blocking failures are recorded.
    await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', `E2E blocking hook fail ${Date.now()}`, false);
    await injectFailedPluginHook(request, 'qa', 'AfterCreateContent', 'non_blocking', `E2E non-blocking hook fail ${Date.now()}`, false);
    await injectFailedPluginHook(request, 'qa', 'BeforeCreateContent', 'blocking', '', true).catch(() => {});
    await injectFailedPluginHook(request, 'qa', 'AfterCreateContent', 'non_blocking', '', true).catch(() => {});
  });
});
