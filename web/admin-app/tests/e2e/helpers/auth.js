import { expect } from '@playwright/test';

export const adminUser = {
  id: 1,
  username: 'admin',
  nickname: '超级管理员',
  role: '超级管理员',
  role_code: 'super_admin',
  permissions: ['*'],
};

export async function seedAdminSession(page, user = adminUser) {
  await page.goto('/admin-next/login');
  await page.evaluate((currentUser) => {
    sessionStorage.setItem('devhub_admin_token', `devhub-admin-${currentUser.id || 1}`);
    sessionStorage.setItem('devhub_admin_refresh_token', `devhub-admin-${currentUser.id || 1}-refresh`);
    sessionStorage.setItem('devhub_admin_user', JSON.stringify(currentUser));
  }, user);
}

export async function loginAdminByUI(page) {
  await page.goto('/admin-next/login');
  await expect(page.getByTestId('admin-login-form')).toBeVisible();
  await page.getByRole('button', { name: '登录' }).click();
  await expect(page).toHaveURL(/\/admin-next\/dashboard/);
}

export async function clearAdminSession(page) {
  await page.evaluate(() => sessionStorage.clear());
}
