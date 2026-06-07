import { expect, test } from '@playwright/test';
import { adminToken, apiPost } from './helpers/api.js';

const enabledAnnouncementConfig = {
  enabled: true,
  message: '前台公告 E2E',
  link_text: '查看详情',
  link_url: '/topics/1/',
  dismissible: false,
};

const resetAnnouncementConfig = {
  enabled: true,
  message: '',
  link_text: '',
  link_url: '/',
  dismissible: false,
};

async function putAnnouncementConfig(request, configJSON) {
  const response = await request.put('/api/v1/admin/plugins/official_announcement/config', {
    headers: {
      Authorization: `Bearer ${adminToken}`,
      'Content-Type': 'application/json',
    },
    data: { config_json: configJSON },
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

test.describe.configure({ mode: 'serial' });

test.describe('official announcement frontend mount', () => {
  test.beforeEach(async ({ request }) => {
    const enableResponse = await apiPost(request, '/api/v1/admin/plugins/official_announcement/enable', {}, adminToken);
    expect(enableResponse.ok(), await enableResponse.text()).toBeTruthy();
    await putAnnouncementConfig(request, enabledAnnouncementConfig);
  });

  test.afterEach(async ({ request }) => {
    await putAnnouncementConfig(request, resetAnnouncementConfig).catch(() => {});
  });

  test('homepage renders official announcement iframe content', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('frontend-site-header')).toBeVisible();

    const mount = page.locator('[data-devhub-plugin-mount][data-plugin-code="official_announcement"]');
    await expect(mount).toBeVisible();

    const iframe = mount.locator('iframe');
    await expect(iframe).toHaveCount(1);

    const frame = page.frameLocator('[data-devhub-plugin-mount][data-plugin-code="official_announcement"] iframe');
    await expect(frame.getByText('前台公告 E2E')).toBeVisible();
    await expect(frame.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/topics/1/');
  });
});
