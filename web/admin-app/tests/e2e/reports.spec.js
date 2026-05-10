import { expect, test } from '@playwright/test';
import { createReportForTopic, createTestTopic } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';
import { expectAdminPageReady, findRowByText } from './helpers/selectors.js';

test.describe('admin reports detailed flow', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('opens, filters, handles a report and records audit trail', async ({ page, request }) => {
    const title = `E2E Report Target ${Date.now()}`;
    const topic = await createTestTopic(request, { title });
    await createReportForTopic(request, topic.id, 'E2E report detail');

    await page.goto('/admin-next/reports');
    await expectAdminPageReady(page, 'admin-reports-page');
    await expect(page.getByTestId('admin-reports-table')).toContainText(title);
    await page.getByTestId('admin-report-status-filter').click();
    await page.getByRole('option', { name: '待处理' }).click();
    await page.getByTestId('admin-report-type-filter').click();
    await page.getByRole('option', { name: '主题' }).click();
    await page.getByRole('button', { name: '查询' }).first().click();

    const row = findRowByText(page, title);
    await expect(row).toBeVisible();
    await row.getByRole('button', { name: '驳回' }).click();
    await expect(page.getByTestId('admin-report-handle-dialog')).toBeVisible();
    await page.getByTestId('admin-report-handle-dialog').getByRole('button', { name: '确认' }).click();
    await expect(row).toBeHidden();

    await page.goto('/admin-next/audit-logs');
    await expect(page.getByTestId('admin-audit-logs-page')).toBeVisible();
    await page.getByTestId('admin-audit-action-search').fill('处理举报');
    await page.getByRole('button', { name: '查询' }).first().click();
    await expect(page.locator('.el-table')).toContainText('处理举报');
  });
});

