import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('admin core pages', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('opens content management and runs a title search', async ({ page }) => {
    await page.goto('/admin-next/content');
    await expect(page.getByTestId('admin-content-page')).toBeVisible();
    await page.getByPlaceholder('请输入').fill('Go');
    await page.getByRole('button', { name: '查询' }).first().click();
    await expect(page.getByRole('tab', { name: '内容管理' })).toBeVisible();
  });

  test('opens comments management and filters comments', async ({ page }) => {
    await page.goto('/admin-next/comments');
    await expect(page.getByTestId('admin-comments-page')).toBeVisible();
    await page.getByPlaceholder('内容 / 作者 / 标题').fill('Go');
    await expect(page.locator('.el-table')).toBeVisible();
  });

  test('opens communities management', async ({ page }) => {
    await page.goto('/admin-next/communities');
    await expect(page.getByTestId('admin-communities-page')).toBeVisible();
    await page.getByPlaceholder('名称 / slug').fill('php');
    await expect(page.getByRole('tab', { name: '子站管理' })).toBeVisible();
  });

  test('opens tags management and searches tags', async ({ page }) => {
    await page.goto('/admin-next/tags');
    await expect(page.getByTestId('admin-tags-page')).toBeVisible();
    await page.getByPlaceholder('名称 / slug / 描述').fill('go');
    await page.getByRole('button', { name: '查询' }).first().click();
    await expect(page.getByRole('tab', { name: '标签管理' })).toBeVisible();
  });

  test('opens audit logs and filters actions', async ({ page }) => {
    await page.goto('/admin-next/audit-logs');
    await expect(page.getByTestId('admin-audit-logs-page')).toBeVisible();
    await page.getByPlaceholder('动作关键词').fill('plugin');
    await page.getByRole('button', { name: '查询' }).first().click();
    await expect(page.getByText('治理审计日志')).toBeVisible();
  });
});
