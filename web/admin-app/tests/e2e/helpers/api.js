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

export async function pluginHealthSummary(request) {
  const response = await adminGet(request, '/api/v1/admin/plugins/health');
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function pluginHealth(request, code) {
  const response = await adminGet(request, `/api/v1/admin/plugins/${code}/health`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
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
  const plugin = await adminGet(request, '/api/v1/admin/plugins').then((response) => response.json()).then((body) => (body.items || []).find((item) => item.code === code));
  if (plugin?.status === 'archived') {
    await restorePlugin(request, code).catch(() => {});
  }
  await adminPost(request, `/api/v1/admin/plugins/${code}/enable`);
}

export async function archivePlugin(request, code) {
  const response = await adminPost(request, `/api/v1/admin/plugins/${code}/archive`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function restorePlugin(request, code) {
  const response = await adminPost(request, `/api/v1/admin/plugins/${code}/restore`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function disablePlugin(request, code) {
  const response = await adminPost(request, `/api/v1/admin/plugins/${code}/disable`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function validatePluginManifest(request, payload) {
  const response = await adminPost(request, '/api/v1/admin/plugins/manifest/validate', payload);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function dryRunPluginManifest(request, payload) {
  const response = await adminPost(request, '/api/v1/admin/plugins/dry-run', payload);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function installPluginManifest(request, payload) {
  const response = await adminPost(request, '/api/v1/admin/plugins/install', payload);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function bulkArchivePlugins(request, codes) {
  const response = await adminPost(request, '/api/v1/admin/plugins/bulk-archive', { codes });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function bulkRestorePlugins(request, codes) {
  const response = await adminPost(request, '/api/v1/admin/plugins/bulk-restore', { codes });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function enableCommunityPlugin(request, communityID, code) {
  return adminPost(request, `/api/v1/admin/communities/${communityID}/plugins/${code}/enable`);
}

export async function disableCommunityPlugin(request, communityID, code) {
  const response = await adminPost(request, `/api/v1/admin/communities/${communityID}/plugins/${code}/disable`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function injectFailedPluginMigration(request, code, migrationName, errorMessage = 'E2E forced migration failure') {
  const response = await adminPost(request, `/api/v1/admin/plugins/${code}/migrations/${encodeURIComponent(migrationName)}/e2e-fail`, {
    error_message: errorMessage,
  });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function retryPluginMigration(request, code, migrationName) {
  const response = await adminPost(request, `/api/v1/admin/plugins/${code}/migrations/${encodeURIComponent(migrationName)}/retry`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function pluginMigrations(request, code) {
  const response = await adminGet(request, `/api/v1/admin/plugins/${code}/migrations`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function pluginAuditLogs(request, code, params = {}) {
  const search = new URLSearchParams(params);
  const suffix = search.toString() ? `?${search.toString()}` : '';
  const response = await adminGet(request, `/api/v1/admin/plugins/${code}/audit-logs${suffix}`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function pluginHooks(request, code) {
  const response = await adminGet(request, `/api/v1/admin/plugins/${code}/hooks`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function injectFailedPluginHook(request, code, hookName, mode, errorMessage, clear = false) {
  const response = await adminPost(request, `/api/v1/admin/plugins/${code}/hooks/${encodeURIComponent(hookName)}/e2e-fail`, {
    mode,
    error_message: errorMessage,
    clear,
  });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

export async function setCategoryEnabled(request, categoryID, enabled) {
  const action = enabled ? 'enable' : 'disable';
  const response = await adminPost(request, `/api/v1/admin/categories/${categoryID}/${action}`);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
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
