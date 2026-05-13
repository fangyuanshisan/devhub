import { expect, test } from '@playwright/test';
import { adminGet, adminPost, archivePlugin, ensurePluginEnabled, restorePlugin } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

test.describe('admin plugin content detailed flow', () => {
  test.beforeEach(async ({ page, request }) => {
    for (const code of ['qa', 'docs', 'wiki']) {
      await ensurePluginEnabled(request, code);
    }
    await seedAdminSession(page);
  });

  test('opens qa/docs/wiki plugin content pages and filters rows', async ({ page }) => {
    for (const path of ['qa', 'docs', 'wiki']) {
      await page.goto(`/admin-next/${path}`);
      await expect(page.getByTestId('plugin-content-page')).toBeVisible();
      await expect(page.getByTestId('plugin-content-header')).toContainText(path);
      await expect(page.getByTestId('plugin-content-type-count')).toBeVisible();
      await expect(page.getByTestId('plugin-content-health')).toBeVisible();
      await expect(page.getByTestId('plugin-content-community-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-status-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-type-filter')).toBeVisible();
      await expect(page.getByTestId('plugin-content-search')).toBeVisible();
      await page.getByTestId('plugin-content-search').fill('E2E');
      await page.getByTestId('plugin-content-query').click();
      await expect(page.getByTestId('plugin-content-table')).toBeVisible();
    }
  });

  test('filters admin posts by plugin_code and content_type precisely', async ({ request }) => {
    const qa = await adminGet(request, '/api/v1/admin/posts?site=portal&board=all&plugin_code=qa&content_type=question&page_size=50');
    expect(qa.ok(), await qa.text()).toBeTruthy();
    const qaBody = await qa.json();
    expect((qaBody.items || []).length).toBeGreaterThan(0);
    for (const item of qaBody.items || []) {
      expect(item.plugin_code).toBe('qa');
      expect(item.content_type).toBe('question');
    }

    const docs = await adminGet(request, '/api/v1/admin/posts?site=portal&board=all&plugin_code=docs&content_type=document&page_size=50');
    expect(docs.ok(), await docs.text()).toBeTruthy();
    const docsBody = await docs.json();
    for (const item of docsBody.items || []) {
      expect(item.plugin_code).toBe('docs');
      expect(item.content_type).toBe('document');
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
      await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
      await page.getByTestId('plugin-content-batch-result').getByRole('button', { name: '关闭' }).click();
      await expect(page.getByTestId('plugin-content-table')).toBeVisible();

      const hiddenRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
      await hiddenRow.locator('.el-checkbox__input').click();
      await page.getByTestId('plugin-content-batch-restore').click();
      await page.getByRole('button', { name: '确认' }).click();
      await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
      await page.getByTestId('plugin-content-batch-result').getByRole('button', { name: '关闭' }).click();
      await page.getByTestId('plugin-content-back').click();
      await expect(page).toHaveURL(/\/admin-next\/plugins\/content/);
    } finally {
      await restorePlugin(request, 'qa').catch(() => {});
      await ensurePluginEnabled(request, 'qa').catch(() => {});
    }
  });

  test('runs minimal batch pin and unpin governance chain', async ({ page }) => {
    await page.goto('/admin-next/qa');
    await page.getByTestId('plugin-content-search').fill('PHP-FPM');
    await page.getByTestId('plugin-content-query').click();
    await expect(page.getByTestId('plugin-content-table')).toContainText('PHP-FPM');

    let firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
    await firstRow.locator('.el-checkbox__input').click();
    await page.getByTestId('plugin-content-batch-pin').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
    await page.getByTestId('plugin-content-batch-result').getByRole('button', { name: '关闭' }).click();

    firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
    await firstRow.locator('.el-checkbox__input').click();
    await page.getByTestId('plugin-content-batch-unpin').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
  });

  test('runs minimal batch feature and unfeature governance chain', async ({ page }) => {
    await page.goto('/admin-next/qa');
    await page.getByTestId('plugin-content-search').fill('PHP-FPM');
    await page.getByTestId('plugin-content-query').click();
    await expect(page.getByTestId('plugin-content-table')).toContainText('PHP-FPM');

    let firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
    await firstRow.locator('.el-checkbox__input').click();
    await page.getByTestId('plugin-content-batch-feature').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
    await page.getByTestId('plugin-content-batch-result').getByRole('button', { name: '关闭' }).click();

    firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
    await firstRow.locator('.el-checkbox__input').click();
    await page.getByTestId('plugin-content-batch-unfeature').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
  });

  test('runs minimal batch reject and approve governance chain', async ({ page }) => {
    await page.goto('/admin-next/qa');
    await page.getByTestId('plugin-content-search').fill('PHP-FPM');
    await page.getByTestId('plugin-content-query').click();
    await expect(page.getByTestId('plugin-content-table')).toContainText('PHP-FPM');

    let firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
    await firstRow.locator('.el-checkbox__input').click();
    await page.getByTestId('plugin-content-batch-reject').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
    await page.getByTestId('plugin-content-batch-result').getByRole('button', { name: '关闭' }).click();

    firstRow = page.getByTestId('plugin-content-table').locator('.el-table__body-wrapper .el-table__row').first();
    await firstRow.locator('.el-checkbox__input').click();
    await page.getByTestId('plugin-content-batch-approve').click();
    await page.getByRole('button', { name: '确认' }).click();
    await expect(page.getByTestId('plugin-content-batch-result')).toContainText('成功数');
  });
});
