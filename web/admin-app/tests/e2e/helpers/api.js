export const adminToken = 'devhub-admin-1';
export const userToken = (id = 1) => `devhub-user-${id}`;

export function adminHeaders() {
  return { Authorization: `Bearer ${adminToken}` };
}

export function userHeaders(id = 1) {
  return { Authorization: `Bearer ${userToken(id)}` };
}

export function uniqueTitle(prefix) {
  return `${prefix} ${Date.now()} ${Math.floor(Math.random() * 1000)}`;
}

export async function adminGet(request, url) {
  return request.get(url, { headers: adminHeaders() });
}

export async function adminPost(request, url, data = {}) {
  return request.post(url, {
    headers: { ...adminHeaders(), 'Content-Type': 'application/json' },
    data,
  });
}

export async function userPost(request, url, data = {}, userID = 1) {
  return request.post(url, {
    headers: { ...userHeaders(userID), 'Content-Type': 'application/json' },
    data,
  });
}

export async function ensurePluginEnabled(request, code) {
  await adminPost(request, `/api/v1/admin/plugins/${code}/enable`);
}

export async function createTestTopic(request, overrides = {}) {
  const payload = {
    community_id: 1,
    community_slug: 'php',
    category_id: 101,
    content_type: 'article',
    title: uniqueTitle('E2E Admin Topic'),
    summary: 'E2E 后台测试内容摘要',
    content: '这是一段用于后台 E2E 自动化测试的正文内容，长度满足发布校验。',
    tags: [],
    ...overrides,
  };
  const response = await userPost(request, '/api/v1/topics', payload);
  if (!response.ok()) throw new Error(await response.text());
  const body = await response.json();
  return body.topic || body;
}

export async function createReportForTopic(request, topicID, reasonText = 'E2E admin report') {
  const response = await userPost(request, '/api/v1/reports', {
    target_type: 'topic',
    target_id: Number(topicID),
    reason_type: 'spam',
    reason_text: reasonText,
  });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}
