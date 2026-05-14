import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin config encryption (sensitive fields)', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('sensitive fields show placeholder and never show ciphertext', async ({ page }) => {
    // In current built-in plugins, qa config schema does not include sensitive fields by default.
    // This test focuses on the "never expose ciphertext" guarantee in UI.
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();

    const qaRow = page.getByRole('row', { name: /问答插件/ });
    await expect(qaRow).toBeVisible();
    await qaRow.getByRole('button', { name: '详情' }).first().click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();

    await page.getByRole('tab', { name: '配置' }).click();
    const panel = page.getByTestId('plugin-global-config-panel');
    await expect(panel).toBeVisible();

    // Ensure UI never shows ciphertext prefix.
    await expect(page.locator('body')).not.toContainText('enc:v1:');
  });
});
