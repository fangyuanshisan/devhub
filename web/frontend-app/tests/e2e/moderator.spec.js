import { expect, test } from '@playwright/test';
import { frontendUser, phpModeratorUser, seedUserSession } from './helpers/auth.js';
import { apiGet, userToken } from './helpers/api.js';

test.describe('frontend moderator workspace boundaries', () => {
  test('regular frontend user cannot access moderator workspace data', async ({ page }) => {
    await seedUserSession(page, frontendUser);
    for (const path of ['/moderator', '/moderator/reports', '/moderator/topics', '/moderator/comments', '/moderator/audit-logs']) {
      await page.goto(path);
      await expect(page.locator('[data-moderator-app]')).toBeVisible();
      await expect(page.locator('[data-moderator-message]')).toContainText(/不是启用状态子站版主|无权限|无法加载版主子站/);
    }
  });

  test('php moderator can open authorized pages and is blocked from go community API', async ({ page, request }) => {
    await seedUserSession(page, phpModeratorUser);
    for (const path of ['/moderator', '/moderator/reports', '/moderator/topics', '/moderator/comments', '/moderator/audit-logs']) {
      await page.goto(path);
      await expect(page.locator('[data-moderator-app]')).toBeVisible();
      await expect(page.locator('[data-community-select]')).toContainText('PHP');
      await expect(page.locator('[data-plugin-links]')).toBeVisible();
    }

    const allowed = await apiGet(request, '/api/v1/moderator/topics?community_id=1&page_size=5', userToken(2));
    expect(allowed.ok(), await allowed.text()).toBeTruthy();

    const forbidden = await apiGet(request, '/api/v1/moderator/topics?community_id=2&page_size=5', userToken(2));
    expect(forbidden.status()).toBe(403);
  });
});

