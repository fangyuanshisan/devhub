import { expect, test } from '@playwright/test';

const adminUser = {
  id: 1,
  username: 'admin',
  nickname: '超级管理员',
  role: '超级管理员',
  role_code: 'super_admin',
  permissions: ['*'],
};

async function login(page) {
  await page.goto('/admin-next/login');
  await page.getByRole('button', { name: '登录' }).click();
  await expect(page).toHaveURL(/\/admin-next\/dashboard/);
}

async function seedAdminSession(page) {
  await page.goto('/admin-next/login');
  await page.evaluate((user) => {
    sessionStorage.setItem('devhub_admin_token', 'devhub-admin-1');
    sessionStorage.setItem('devhub_admin_refresh_token', 'devhub-admin-1-refresh');
    sessionStorage.setItem('devhub_admin_user', JSON.stringify(user));
  }, adminUser);
}

async function ensurePluginEnabled(request, code) {
  await request.post(`/api/v1/admin/plugins/${code}/enable`, {
    headers: { Authorization: 'Bearer devhub-admin-1' },
  });
}

test.describe('plugin governance center', () => {
  test.beforeEach(async ({ page, request }) => {
    await ensurePluginEnabled(request, 'qa');
    await seedAdminSession(page);
  });

  test('opens plugin center and filters plugin list', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await expect(page.getByTestId('admin-plugins-page')).toBeVisible();
    await expect(page.getByTestId('plugin-stats')).toContainText('全部插件');
    await expect(page.getByText('问答插件')).toBeVisible();

    await page.getByPlaceholder('搜索 code / name').fill('qa');
    await expect(page.getByText('问答插件')).toBeVisible();
    await expect(page.getByText('文档插件')).toBeHidden();
  });

  test('opens plugin detail tabs and shows schema validation errors', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await page.getByTestId('plugin-detail-qa').click();
    await expect(page.getByTestId('plugin-detail-drawer')).toBeVisible();

    for (const tabName of ['概览', '内容类型', '权限', '菜单', '配置', 'Hooks', '路由', '审计']) {
      await page.getByRole('tab', { name: tabName }).click();
      await expect(page.getByRole('tab', { name: tabName })).toHaveAttribute('aria-selected', 'true');
    }

    await page.getByRole('tab', { name: '配置' }).click();
    await expect(page.getByRole('button', { name: 'config_schema' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'resolved_config' })).toBeVisible();

    await page.getByTestId('json-clear-object').click();
    await expect(page.getByTestId('schema-error-box')).toContainText('required');
    await expect(page.getByTestId('plugin-global-config-save')).toBeDisabled();
  });

  test('shows impact before global disable and blocks community enable when globally disabled', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await page.getByTestId('plugin-disable-qa').click();
    await expect(page.getByRole('dialog')).toContainText('历史内容详情页和 SEO 不受影响');
    await expect(page.getByRole('dialog')).toContainText('将影响子站');
    await page.getByRole('button', { name: '确认禁用' }).click();
    await expect(page.getByText('插件状态已更新')).toBeVisible();

    await page.goto('/admin-next/communities');
    await page.getByTestId('community-plugins-1').click();
    await expect(page.getByTestId('community-plugin-drawer')).toBeVisible();
    await expect(page.getByText('该插件已被全局禁用')).toBeVisible();
  });

  test('opens community plugin config and blocks invalid schema values', async ({ page }) => {
    await page.goto('/admin-next/communities');
    await page.getByTestId('community-plugins-1').click();
    await expect(page.getByTestId('community-plugin-drawer')).toBeVisible();
    await expect(page.getByTestId('community-plugin-drawer').getByText('子站已启用').first()).toBeVisible();

    await page.getByTestId('community-plugin-config-qa').click();
    await expect(page.getByTestId('community-plugin-config-dialog')).toBeVisible();
    await expect(page.getByTestId('plugin-json-editor').getByText('子站 config_json')).toBeVisible();
    await page.getByTestId('json-clear-object').click();
    await expect(page.getByTestId('schema-error-box')).toContainText('required');
    await expect(page.getByTestId('community-plugin-config-save')).toBeDisabled();
  });

  test('opens generic plugin content page with filters', async ({ page }) => {
    await page.goto('/admin-next/plugins');
    await page.getByTestId('plugin-manage-qa').click();
    await expect(page).toHaveURL(/\/admin-next\/qa/);
    await expect(page.getByTestId('plugin-content-page')).toBeVisible();
    await expect(page.getByText('content_type：')).toBeVisible();
    await expect(page.getByTestId('plugin-content-community-filter')).toBeVisible();
    await expect(page.getByTestId('plugin-content-status-filter')).toBeVisible();
    await page.getByTestId('plugin-content-back').click();
    await expect(page).toHaveURL(/\/admin-next\/plugins/);
  });
});
