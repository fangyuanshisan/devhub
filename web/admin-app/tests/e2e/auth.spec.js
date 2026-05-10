import { expect, test } from '@playwright/test';
import { clearAdminSession, loginAdminByUI } from './helpers/auth.js';

test.describe('admin auth', () => {
  test('admin can login and protected pages require an admin session', async ({ page }) => {
    await loginAdminByUI(page);
    await expect(page.getByRole('tab', { name: '控制台' })).toBeVisible();

    await clearAdminSession(page);
    await page.goto('/admin-next/content');
    await expect(page).toHaveURL(/\/admin-next\/login/);
  });
});
