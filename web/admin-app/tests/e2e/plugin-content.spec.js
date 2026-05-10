import { expect, test } from '@playwright/test';
import { adminPost, ensurePluginEnabled } from './helpers/api.js';
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
      await expect(page.getByTestId('plugin-content-page')).toContainText(`content_type：${type}`);
      await expect(page.getByTestId('plugin-content-community-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-status-filter')).toBeVisible();
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
});
