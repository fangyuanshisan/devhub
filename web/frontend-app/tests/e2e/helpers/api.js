import { expect } from '@playwright/test';

export const adminToken = 'devhub-admin-1';
export const userToken = (id = 1) => `devhub-user-${id}`;

export function uniqueTitle(prefix) {
  return `${prefix} ${Date.now()} ${Math.floor(Math.random() * 1000)}`;
}

export function authHeaders(token = userToken()) {
  return { Authorization: `Bearer ${token}` };
}

export async function apiGet(request, url, token = userToken()) {
  return request.get(url, { headers: authHeaders(token) });
}

export async function apiPost(request, url, data, token = userToken()) {
  return request.post(url, {
    headers: { ...authHeaders(token), 'Content-Type': 'application/json' },
    data,
  });
}

export async function createTestTopic(request, overrides = {}, token = userToken()) {
  const payload = {
    community_id: 1,
    community_slug: 'php',
    category_id: 101,
    content_type: 'article',
    title: uniqueTitle('E2E Topic'),
    summary: 'E2E 自动化测试内容摘要',
    content: '这是一段用于 E2E 自动化测试的正文内容，长度满足发布校验。',
    tags: [],
    ...overrides,
  };
  const response = await apiPost(request, '/api/v1/topics', payload, token);
  expect(response.ok(), await response.text()).toBeTruthy();
  const body = await response.json();
  return body.topic || body;
}

export async function createReportForTopic(request, topicID, reasonText = 'E2E report') {
  const response = await apiPost(request, '/api/v1/reports', {
    target_type: 'topic',
    target_id: Number(topicID),
    reason_type: 'spam',
    reason_text: reasonText,
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

export async function getPlugin(request, code) {
  const response = await apiGet(request, '/api/v1/admin/plugins', adminToken);
  expect(response.ok(), await response.text()).toBeTruthy();
  const body = await response.json();
  return (body.items || []).find((item) => item.code === code);
}

export async function setPluginEnabled(request, code, enabled) {
  const action = enabled ? 'enable' : 'disable';
  const response = await apiPost(request, `/api/v1/admin/plugins/${code}/${action}`, {}, adminToken);
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

export async function archivePlugin(request, code) {
  const response = await apiPost(request, `/api/v1/admin/plugins/${code}/archive`, {}, adminToken);
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

export async function restorePlugin(request, code) {
  const response = await apiPost(request, `/api/v1/admin/plugins/${code}/restore`, {}, adminToken);
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

export async function setCategoryEnabled(request, categoryID, enabled) {
  const action = enabled ? 'enable' : 'disable';
  const response = await apiPost(request, `/api/v1/admin/categories/${categoryID}/${action}`, {}, adminToken);
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

export async function ensurePluginStatus(request, code, status) {
  const plugin = await getPlugin(request, code);
  if (plugin?.status === 'archived' && status !== 'archived') {
    await restorePlugin(request, code).catch(() => {});
  }
  if (!plugin || plugin.status !== status) {
    await setPluginEnabled(request, code, status === 'enabled');
  }
}

export async function enableCommunityPlugin(request, communityID, code, expectOK = true) {
  const response = await apiPost(request, `/api/v1/admin/communities/${communityID}/plugins/${code}/enable`, {}, adminToken);
  if (expectOK) {
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  }
  return response;
}

export async function expectSeoBasics(page) {
  await expect(page.locator('h1').first()).toBeVisible();
  await expect(page.locator('article').first()).toBeVisible();
  await expect(page.locator('title')).toHaveCount(1);
}
