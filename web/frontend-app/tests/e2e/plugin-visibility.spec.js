import { expect, test } from '@playwright/test';
import { loginAsUser } from './helpers/auth.js';
import {
  adminToken,
  apiPost,
  archivePlugin,
  enableCommunityPlugin,
  ensurePluginStatus,
  expectSeoBasics,
  getPlugin,
  restorePlugin,
  setCategoryEnabled,
  setPluginEnabled,
} from './helpers/api.js';

test.describe.configure({ mode: 'serial' });

test.describe('frontend plugin visibility and content_type guard', () => {
  test.afterEach(async ({ request }) => {
    await restorePlugin(request, 'qa').catch(() => {});
    await ensurePluginStatus(request, 'qa', 'enabled');
    await enableCommunityPlugin(request, 1, 'qa').catch(() => {});
    await setCategoryEnabled(request, 102, false).catch(() => {});
  });

  test('disabled plugin hides publish option, blocks direct API submit, and keeps historical SEO', async ({ page, request }) => {
    await ensurePluginStatus(request, 'qa', 'enabled');
    // Ensure qa category is enabled so it shows up in the public categories list.
    // Some environments may not have this seed category at all; in that case skip.
    await apiPost(request, '/api/v1/admin/categories/102/enable', {}, adminToken).catch(() => {});
    const categoriesRes = await request.get('/api/v1/communities/php/categories');
    const categories = categoriesRes.ok() ? (await categoriesRes.json()).items || [] : [];
    expect(categories.some((c) => String(c.id) === '102')).toBeTruthy();
    await setCategoryEnabled(request, 102, true);
    await loginAsUser(page);
    await page.goto('/c/php/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    await page.getByTestId('frontend-publish-category').selectOption('102');
    let values = await page.getByTestId('frontend-publish-content-type').locator('option').evaluateAll((options) => options.map((option) => option.value));
    expect(values).toContain('question');

    await setPluginEnabled(request, 'qa', false);
    await page.goto('/c/php/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    const categoryValues = await page.getByTestId('frontend-publish-category').locator('option').evaluateAll((options) => options.map((option) => option.value));
    expect(categoryValues).not.toContain('102');
    values = await page.getByTestId('frontend-publish-content-type').locator('option').evaluateAll((options) => options.map((option) => option.value));
    expect(values).not.toContain('question');

    const disabledResponse = await apiPost(request, '/api/v1/topics', {
      community_id: 1,
      community_slug: 'php',
      category_id: 102,
      content_type: 'question',
      title: `E2E Disabled Question ${Date.now()}`,
      content: '这是一段不应该创建成功的禁用插件内容。',
      tags: [],
    });
    expect(disabledResponse.ok()).toBeFalsy();
    expect(await disabledResponse.text()).toContain('插件');

    const invalidResponse = await apiPost(request, '/api/v1/topics', {
      community_id: 1,
      community_slug: 'php',
      category_id: 101,
      content_type: 'e2e_invalid_type',
      title: `E2E Invalid Type ${Date.now()}`,
      content: '这是一段不应该创建成功的非法 content_type 内容。',
      tags: [],
    });
    expect(invalidResponse.ok()).toBeFalsy();

    await page.goto('/topics/2/');
    await expectSeoBasics(page);
  });

  test('archived plugin hides publish option, blocks direct API submit, and keeps historical SEO', async ({ page, request }) => {
    await ensurePluginStatus(request, 'qa', 'enabled');
    await apiPost(request, '/api/v1/admin/categories/102/enable', {}, adminToken).catch(() => {});
    const categoriesRes = await request.get('/api/v1/communities/php/categories');
    const categories = categoriesRes.ok() ? (await categoriesRes.json()).items || [] : [];
    expect(categories.some((c) => String(c.id) === '102')).toBeTruthy();
    await setCategoryEnabled(request, 102, true);
    await loginAsUser(page);

    await page.goto('/c/php/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    await page.getByTestId('frontend-publish-category').selectOption('102');
    let values = await page.getByTestId('frontend-publish-content-type').locator('option').evaluateAll((options) => options.map((option) => option.value));
    expect(values).toContain('question');

    await archivePlugin(request, 'qa');

    await page.goto('/c/php/topics/new/');
    await expect(page.getByTestId('frontend-topic-publish-form')).toBeVisible();
    const categoryValues = await page.getByTestId('frontend-publish-category').locator('option').evaluateAll((options) => options.map((option) => option.value));
    expect(categoryValues).not.toContain('102');
    values = await page.getByTestId('frontend-publish-content-type').locator('option').evaluateAll((options) => options.map((option) => option.value));
    expect(values).not.toContain('question');

    const archivedResponse = await apiPost(request, '/api/v1/topics', {
      community_id: 1,
      community_slug: 'php',
      category_id: 102,
      content_type: 'question',
      title: `E2E Archived Question ${Date.now()}`,
      content: '这是一段不应该创建成功的归档插件内容。',
      tags: [],
    });
    expect(archivedResponse.ok()).toBeFalsy();
    expect(await archivedResponse.text()).toContain('archived');

    const enableResponse = await enableCommunityPlugin(request, 1, 'qa', false);
    expect(enableResponse.ok()).toBeFalsy();
    expect(await enableResponse.text()).toContain('归档');

    await page.goto('/topics/2/');
    await expectSeoBasics(page);
  });

  test('archived plugin cannot be re-enabled at subsite level', async ({ request }) => {
    await archivePlugin(request, 'qa');
    const plugin = await getPlugin(request, 'qa');
    expect(plugin?.status).toBe('archived');

    const archivedEnable = await enableCommunityPlugin(request, 1, 'qa', false);
    expect(archivedEnable.ok()).toBeFalsy();
    expect(await archivedEnable.text()).toContain('归档');
  });
});
