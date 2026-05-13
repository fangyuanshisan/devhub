import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package repository', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('scans repository and shows discovered packages list + detail', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();

    await expect(page.getByTestId('plugin-package-repo-panel')).toBeVisible();
    await page.getByTestId('plugin-package-repo-root').fill('plugins-local/repository-fixtures');
    await page.getByTestId('plugin-package-repo-scan').click();

    const table = page.getByTestId('plugin-package-repo-table');
    await expect(table).toBeVisible();
    await expect(table).toContainText('demo_notice');
    await expect(table).toContainText('checksum');
    await expect(page.getByTestId('plugin-package-repo-summary')).toBeVisible();

    // Keyword search.
    await page.getByPlaceholder('keyword：code/name/path').fill('demo_notice');
    await page.getByTestId('plugin-package-repo-refresh').click();
    await expect(table).toContainText('demo_notice');

    // Open detail drawer from the demo_notice row.
    const row = table.getByRole('row', { name: /demo_notice/ }).first();
    await row.getByTestId('plugin-package-repo-detail-btn').click();
    const drawer = page.getByTestId('plugin-package-repo-detail-content');
    await expect(drawer).toBeVisible();
    await expect(page.getByRole('heading', { name: '插件包详情' })).toBeVisible();
  });
});
