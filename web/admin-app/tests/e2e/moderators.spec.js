import { expect, test } from '@playwright/test';
import { adminGet, adminPost, uniqueTitle, userHeaders } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';
import { expectAdminPageReady } from './helpers/selectors.js';

async function registerUniqueUser(request) {
  const username = uniqueTitle('e2e_mod').toLowerCase().replace(/[^a-z0-9_]+/g, '_');
  const response = await request.post('/api/v1/auth/register', {
    data: {
      username,
      nickname: username,
      email: `${username}@example.test`,
      password: 'admin123',
    },
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  const body = await response.json();
  return body.user;
}

test.describe('admin moderators detailed flow', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('opens moderator page and verifies create, update, disable APIs are reflected in list', async ({ page, request }) => {
    const user = await registerUniqueUser(request);
    const created = await adminPost(request, '/api/v1/admin/moderators', {
      community_slug: 'java',
      user_id: user.id,
      role: 'moderator',
      status: 1,
    });
    expect(created.ok(), await created.text()).toBeTruthy();
    const moderator = await created.json();

    await page.goto('/admin-next/moderators');
    await expectAdminPageReady(page, 'admin-moderators-page');
    await expect(page.getByTestId('admin-moderators-table')).toContainText(user.username);

    const updated = await request.put(`/api/v1/admin/moderators/${moderator.id}`, {
      headers: { Authorization: 'Bearer devhub-admin-1' },
      data: { community_slug: 'java', user_id: user.id, role: 'owner', status: 1 },
    });
    expect(updated.ok(), await updated.text()).toBeTruthy();
    await page.reload();
    await expect(page.getByTestId('admin-moderators-table')).toContainText('站长');

    const disabled = await request.delete(`/api/v1/admin/moderators/${moderator.id}`, {
      headers: { Authorization: 'Bearer devhub-admin-1' },
    });
    expect(disabled.ok(), await disabled.text()).toBeTruthy();
    await page.reload();
    await expect(page.getByTestId('admin-moderators-table')).toContainText('停用');

    const denied = await request.get('/api/v1/admin/moderators', { headers: userHeaders(1) });
    expect([401, 403]).toContain(denied.status());

    const listed = await adminGet(request, '/api/v1/admin/moderators?community_slug=java&status=all');
    expect(listed.ok(), await listed.text()).toBeTruthy();
  });
});

