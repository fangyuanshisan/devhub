import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package install', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('installs a valid local package (disabled by default)', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();

    await page.getByTestId('plugin-package-repo-root').fill('plugins-local/repository-fixtures');
    await page.getByTestId('plugin-package-repo-scan').click();

    const table = page.getByTestId('plugin-package-repo-table');
    await expect(table).toBeVisible();
    await expect(table).toContainText('demo_notice_install');

    const row = table.getByRole('row').filter({ hasText: 'plugins-local/repository-fixtures/demo_notice_install' }).first();
    await expect(row.getByTestId('plugin-package-repo-install-btn')).toBeVisible();
    await row.getByTestId('plugin-package-repo-install-btn').click();
    const dialog = page.getByTestId('plugin-package-repo-install-content');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('确认安装');
    await expect(dialog).toContainText('disabled');
    const statusTag = page.getByTestId('plugin-package-repo-install-status');
    await expect(statusTag).toBeVisible();

    const statusText = String(await statusTag.innerText()).toLowerCase();
    const confirmBtn = page.getByTestId('plugin-package-repo-install-confirm');
    if (statusText.includes('blocked')) {
      // If another flow already installed this plugin in the shared E2E database, the install preview becomes blocked.
      // In this case the confirm button stays disabled; assert a stable conflict hint instead.
      await expect(dialog).toContainText(/manifest 校验失败|插件 code 已存在/);
      await page.keyboard.press('Escape');
    } else {
      await expect(confirmBtn).toBeEnabled();
      await confirmBtn.click();
      // If the plugin was already installed by another flow (e.g. approvals), accept the stable error code.
      await expect(page.getByText(/安装成功|安装完成|plugin_package_already_installed/).first()).toBeVisible();
    }

    // Verify plugin appears in plugin list and is disabled.
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
    const pluginTable = page.getByTestId('plugin-table');
    await expect(pluginTable).toContainText('demo_notice_install');
    await expect(pluginTable).toContainText('已禁用');
  });
});
