import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package dry-run', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('dry-runs a safe local plugin package and shows scan + preview', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();
    await expect(page.getByTestId('plugin-package-dryrun-panel')).toBeVisible();

    await page.getByTestId('plugin-package-path-input').fill('examples/plugins/demo_notice');
    await page.getByTestId('plugin-package-dry-run').click();

    const panel = page.getByTestId('plugin-package-result');
    await expect(panel).toBeVisible();
    await expect(panel.getByTestId('plugin-package-info')).toContainText('demo_notice');
    await expect(panel.getByTestId('plugin-package-file-scan')).toContainText('total_files');
    await expect(panel.getByTestId('plugin-package-install-preview')).toBeVisible();
    await expect(page.getByTestId('plugin-package-dryrun-panel')).toContainText('不会安装插件');
  });

  test('shows structured error code when path is invalid', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await page.getByTestId('plugin-package-path-input').fill('../etc/passwd');
    await page.getByTestId('plugin-package-dry-run').click();
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();
    await expect(page.getByTestId('plugin-install-page').getByRole('alert').filter({ hasText: 'plugin_package_path_invalid' })).toBeVisible();
  });

  test('shows structured error code when package does not exist', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await page.getByTestId('plugin-package-path-input').fill('examples/plugins/does_not_exist');
    await page.getByTestId('plugin-package-dry-run').click();
    await expect(page.getByTestId('plugin-install-page').getByRole('alert').filter({ hasText: 'plugin_package_not_found' })).toBeVisible();
  });
});
