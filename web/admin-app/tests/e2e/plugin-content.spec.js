import { expect, test } from '@playwright/test';
import { adminPost, archivePlugin, ensurePluginEnabled, restorePlugin } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

test.describe('admin plugin content detailed flow', () => {
  test.beforeEach(async ({ page, request }) => {
    for (const code of ['qa', 'docs', 'wiki']) {
      await ensurePluginEnabled(request, code);
    }
    await seedAdminSession(page);
  });

  test('opens qa/docs/wiki plugin content pages and filters rows', async ({ page }) => {
    for (const [path, type] of [['qa', 'question'], ['docs', 'document'], ['wiki', 'wiki_page']]) {
      await page.goto(`/admin-next/${path}`);
      await expect(page.getByTestId('plugin-content-page')).toBeVisible();
      await expect(page.getByTestId('plugin-content-page')).toContainText(`内容类型：${type}`);
      await expect(page.getByTestId('plugin-content-community-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-status-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-type-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-search')).toBeVisible();
      await page.getByTestId('plugin-content-search').fill('E2E');
      await page.getByTestId('plugin-content-query').click();
      await expect(page.getByTestId('plugin-content-table')).toBeVisible();
    }
  });

  test('performs one plugin content moderation operation through admin API and checks audit log', async ({ page, request }) => {
    const topicID = 2;
    const title = 'PHP-FPM';

    await page.goto('/admin-next/qa');
    await page.getByTestId('plugin-content-search').fill(title);
    await page.getByTestId('plugin-content-query').click();
    await expect(page.getByTestId('plugin-content-table')).toContainText(title);

    const hide = await adminPost(request, `/api/v1/admin/topics/${topicID}/hide`);
    expect(hide.ok(), await hide.text()).toBeTruthy();
    const restore = await adminPost(request, `/api/v1/admin/topics/${topicID}/restore`);
    expect(restore.ok(), await restore.text()).toBeTruthy();

    await page.goto('/admin-next/audit-logs');
    await page.getByTestId('admin-audit-action-search').fill('恢复主题');
    await page.getByRole('button', { name: '查询' }).first().click();
    await expect(page.locator('.el-table')).toContainText('恢复主题');
  });

  test('keeps archived plugin history manageable with batch hide and restore', async ({ page, request }) => {
    try {
      await archivePlugin(request, 'qa');

      await page.goto('/admin-next/qa');
      await expect(page.getByTestId('plugin-content-page')).toBeVisible();
      await expect(page.getByTestId('plugin-content-archived-tip')).toContainText('只能查看和治理历史内容');
      await page.getByTestId('plugin-content-search').fill('PHP-FPM');
      await page.getByTestId('plugin-content-query').click();
      await expect(page.getByTestId('plugin-content-table')).toContainText('PHP-FPM');

      const firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
      await firstRow.locator('.el-checkbox__input').click();
      await expect(page.getByText('已选择 1 条内容')).toBeVisible();

      await page.getByTestId('plugin-content-batch-hide').click();
      await page.getByRole('button', { name: '确认' }).click();
      await expect(page.locator('.el-message__content').filter({ hasText: '批量操作已完成' }).first()).toBeVisible();
      await expect(page.getByTestId('plugin-content-table')).toBeVisible();

      const hiddenRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
      await hiddenRow.locator('.el-checkbox__input').click();
      await page.getByTestId('plugin-content-batch-restore').click();
      await page.getByRole('button', { name: '确认' }).click();
      await expect(page.locator('.el-message__content').filter({ hasText: '批量操作已完成' }).first()).toBeVisible();
      await page.getByTestId('plugin-content-back').click();
      await expect(page).toHaveURL(/\/admin-next\/plugins/);
    } finally {
      await restorePlugin(request, 'qa').catch(() => {});
      await ensurePluginEnabled(request, 'qa').catch(() => {});
    }
  });
});
