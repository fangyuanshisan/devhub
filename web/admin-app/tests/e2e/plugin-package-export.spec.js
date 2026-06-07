import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package export', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('exports an installed declarative plugin package', async ({ page }) => {
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();

    await page.getByTestId('plugin-detail-qa').click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();
    await page.getByTestId('plugin-export-open').click();

    const dialog = page.getByTestId('plugin-export-dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('不导出敏感配置');
    await expect(dialog).toContainText('用户数据');
    await expect(dialog).toContainText('不生成 zip');
    await expect(dialog).not.toContainText('zip 下载');
    await expect(dialog.getByRole('button', { name: '远程发布' })).toHaveCount(0);
    await expect(dialog).not.toContainText('插件市场');

    await page.getByTestId('plugin-export-dry-run').click();
    await expect(page.getByTestId('plugin-export-preview')).toBeVisible();
    await expect(page.getByTestId('plugin-export-preview')).toContainText('不包含敏感配置');
    await expect(page.getByTestId('plugin-export-preview')).toContainText('不包含用户数据');
    await expect(page.getByTestId('plugin-export-files')).toContainText('manifest.json');
    await expect(page.getByTestId('plugin-export-files')).toContainText('README.md');
    await expect(page.getByTestId('plugin-export-files')).toContainText('config.example.json');
    await expect(page.getByTestId('plugin-export-files')).toContainText('checksums.json');
    await expect(page.getByTestId('plugin-export-submit')).toBeEnabled();

    await page.getByTestId('plugin-export-submit').click();
    await page.getByRole('button', { name: '确认导出' }).last().click();
    await expect(page.getByTestId('plugin-export-result')).toBeVisible();
    await expect(page.getByTestId('plugin-export-result')).toContainText('storage/plugins/exports/');
    await expect(page.getByTestId('plugin-export-result')).toContainText('generated');
    await expect(page.getByTestId('plugin-export-result')).toContainText('dry-run 验证');
  });
});
