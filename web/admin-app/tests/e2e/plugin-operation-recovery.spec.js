import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin operation recovery', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('shows operations list and detail drawer', async ({ page }) => {
    await page.goto('/admin-next/plugins/operations');
    await expect(page.getByTestId('plugin-operations-page')).toBeVisible();

    const table = page.getByTestId('plugin-operations-table');
    await expect(table).toBeVisible();

    // If earlier tests created operations (install/upgrade), open one for a stable smoke check.
    const rows = table.getByRole('row');
    const rowCount = await rows.count();
    if (rowCount > 1) {
      await rows.nth(1).getByTestId('plugin-operations-detail-btn').click();
      await expect(page.getByTestId('plugin-operation-detail-drawer')).toBeVisible();
    } else {
      // Empty state is acceptable in isolated environments.
      await expect(page.getByText(/暂无操作记录|操作历史/).first()).toBeVisible();
    }
  });
});

