import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin approvals', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('submit, approve and execute an install approval', async ({ page }) => {
    // Submit install approval from local repository install dialog.
    await page.goto('/admin-next/plugins/install');
    await expect(page.getByTestId('plugin-install-page')).toBeVisible();

    await page.getByTestId('plugin-package-repo-root').fill('plugins-local/repository-fixtures');
    await page.getByTestId('plugin-package-repo-scan').click();

    const table = page.getByTestId('plugin-package-repo-table');
    await expect(table).toBeVisible();
    await expect(table).toContainText('demo_notice_install');

    // Click install on the row matching the full path to avoid fixed-column row mismatches.
    const row = table.getByRole('row').filter({ hasText: 'plugins-local/repository-fixtures/demo_notice_install' }).first();
    await expect(row).toBeVisible();
    await row.getByTestId('plugin-package-repo-install-btn').click();

    const dialog = page.getByTestId('plugin-package-repo-install-content');
    await expect(dialog).toBeVisible();
    await expect(dialog).toContainText('确认安装');
    await expect(page.getByTestId('plugin-package-repo-install-content')).toContainText('demo_notice_install');
    await expect(page.getByTestId('plugin-package-repo-install-status')).toContainText(/ok|warning/);
    await expect(page.getByTestId('plugin-package-repo-install-approval')).toBeEnabled();
    await page.getByTestId('plugin-package-repo-install-approval').click();

    // Should redirect to approvals page.
    await expect(page.getByTestId('plugin-approvals-page')).toBeVisible();

    const approvalsTable = page.getByTestId('plugin-approvals-table');
    await expect(approvalsTable).toContainText('demo_notice_install');

    // Open detail and approve.
    const approvalRow = approvalsTable.getByRole('row').filter({ hasText: 'demo_notice_install' }).first();
    await approvalRow.getByTestId('plugin-approvals-detail-btn').click();
    await expect(page.getByTestId('plugin-approval-detail-drawer')).toBeVisible();

    await page.getByTestId('plugin-approval-approve').click();
    await expect(page.getByTestId('plugin-approval-review-dialog')).toBeVisible();
    await page.getByTestId('plugin-approval-review-comment').fill('风险可接受，允许安装');
    await page.getByTestId('plugin-approval-review-confirm').click();

    // Execute approved request.
    await page.getByTestId('plugin-approval-execute').click();
    const confirmBtn = page.locator('.el-message-box__btns .el-button--primary').first();
    await expect(confirmBtn).toBeVisible();
    await confirmBtn.click();

    // Verify plugin appears in plugin list and is disabled.
    await page.goto('/admin-next/plugins/list');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
    const pluginTable = page.getByTestId('plugin-table');
    await expect(pluginTable).toContainText('demo_notice_install');
    await expect(pluginTable).toContainText('已禁用');
  });
});
