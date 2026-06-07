import { expect, test } from '@playwright/test';
import { loginAsUser } from './helpers/auth.js';
import { apiGet, apiPost, userToken } from './helpers/api.js';

async function ensureTopicInteractionOff(request, topicID) {
  const response = await apiGet(request, `/api/v1/topics/${topicID}/interaction`);
  if (!response.ok()) return;
  const state = await response.json();
  if (state.liked) await apiPost(request, `/api/v1/topics/${topicID}/like`);
  if (state.favorited) await apiPost(request, `/api/v1/topics/${topicID}/favorite`);
  if (state.followed) await apiPost(request, '/api/v1/follows/toggle', { target_type: 'topic', target_id: topicID });
}

test.describe('frontend logged-in interactions', () => {
  test.beforeEach(async ({ page, request }) => {
    await ensureTopicInteractionOff(request, 1);
    await loginAsUser(page);
  });

  test('logged-in user can like, favorite, follow and comment on topic detail', async ({ page }) => {
    await page.goto('/topics/1/');
    await expect(page.locator('h1').first()).toBeVisible();
    await expect(page.locator('article').first()).toBeVisible();
    await expect(page.locator('[data-comment-root]')).toBeVisible();
    await expect(page.locator('[data-topic-action="like"]')).toBeVisible();
    await expect(page.locator('[data-topic-action="favorite"]')).toBeVisible();
    await expect(page.locator('[data-topic-action="follow"]')).toBeVisible();

    await page.locator('[data-topic-action="like"]').click();
    await expect(page.locator('[data-action-message]')).toContainText('点赞状态已更新');
    await expect(page.locator('[data-topic-action="like"]')).toHaveAttribute('data-liked', 'true');

    await page.locator('[data-topic-action="favorite"]').click();
    await expect(page.locator('[data-action-message]')).toContainText('收藏状态已更新');
    await expect(page.locator('[data-topic-action="favorite"]')).toHaveAttribute('data-favorited', 'true');

    await page.locator('[data-topic-action="follow"]').click();
    await expect(page.locator('[data-action-message]')).toContainText('已关注主题');
    await expect(page.locator('[data-topic-action="follow"]')).toHaveAttribute('data-followed', 'true');

    const comment = `E2E Comment ${Date.now()}`;
    await page.locator('[data-comment-form] textarea[name="content"]').fill(comment);
    await page.locator('[data-comment-submit]').click();
    await expect(page.locator('[data-comment-message]')).toContainText('评论已发布');
    await expect(page.locator('[data-topic-comments]')).toContainText(comment);

    await page.goto('/me/favorites');
    await expect(page.locator('[data-favorite-list]')).toBeVisible();
    await expect(page.locator('[data-favorite-list]')).toContainText(/Laravel|主题|E2E|还没有收藏内容/);

    await page.goto('/me/follows');
    await expect(page.locator('[data-follow-list]')).toBeVisible();
    await expect(page.locator('[data-follow-list]')).toContainText(/内容|主题|还没有关注任何内容/);
  });

  test('user center pages keep frontend user session', async ({ page }) => {
    for (const [url, selector] of [
      ['/notifications', '[data-notification-list]'],
      ['/me/activities', '[data-activity-list]'],
      ['/me/favorites', '[data-favorite-list]'],
      ['/me/follows', '[data-follow-list]'],
    ]) {
      await page.goto(url);
      await expect(page.locator(selector)).toBeVisible();
      await expect(page.getByTestId('frontend-user-menu')).toBeVisible();
    }
    await expect(page.evaluate(() => localStorage.getItem('devhub_user_token'))).resolves.toBe(userToken(1));
  });
});

