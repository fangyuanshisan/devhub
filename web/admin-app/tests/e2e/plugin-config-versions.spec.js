import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin config versions (history + rollback dry-run)', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('records config versions and shows diff + rollback preview', async ({ page }) => {
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();

    const qaRow = page.getByRole('row', { name: /问答插件/ });
    await expect(qaRow).toBeVisible();
    await qaRow.getByRole('button', { name: '详情' }).first().click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();

    await page.getByRole('tab', { name: '配置' }).click();
    const globalPanel = page.getByTestId('plugin-global-config-panel');
    await expect(globalPanel).toBeVisible();

    // Toggle a boolean field twice to generate versions.
    const field = page.getByTestId('config-field-allow_anonymous_answer');
    if (await field.isVisible()) {
      await field.click();
    } else {
      // fallback: switch to json mode and edit minimal object
      await page.getByTestId('config-mode-toggle').getByText('JSON').click();
      await page.getByTestId('plugin-config-json-mode').click();
    }
    await globalPanel.getByTestId('plugin-global-config-save').click();
    await expect(page.locator('.el-message').first()).toBeVisible();

    if (await field.isVisible()) {
      await field.click();
    }
    await globalPanel.getByTestId('plugin-global-config-save').click();
    await expect(page.locator('.el-message').first()).toBeVisible();

    await globalPanel.getByTestId('plugin-config-versions-open').click();
    await expect(page.getByTestId('plugin-config-versions-dialog')).toBeVisible();
    await expect(page.getByTestId('plugin-config-versions-table')).toBeVisible();

    // Open detail.
    const firstRow = page.getByTestId('plugin-config-versions-table').getByRole('row').nth(1);
    await firstRow.getByTestId('plugin-config-version-detail-btn').click();
    await expect(page.getByTestId('plugin-config-version-detail-drawer')).toBeVisible();

    // Rollback dry-run preview.
    await page.getByTestId('plugin-config-version-detail-drawer').getByLabel('Close').click();
    await firstRow.getByTestId('plugin-config-version-rollback-preview-btn').click();
    await expect(page.getByTestId('plugin-config-rollback-dryrun-drawer')).toBeVisible();
    await expect(page.getByTestId('plugin-config-rollback-preview')).toBeVisible();
  });
});
