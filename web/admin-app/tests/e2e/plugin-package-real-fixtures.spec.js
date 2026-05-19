import { expect, test } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { adminHeaders } from './helpers/api.js';
import { seedAdminSession } from './helpers/auth.js';

const projectRoot = path.resolve(process.cwd(), '../..');
const fixtureScript = path.join(projectRoot, 'scripts/build-plugin-package-fixtures.sh');
const fixtureDist = path.join(projectRoot, 'scripts/fixtures/plugin-packages/dist');

test.describe('plugin package real fixture upload promote install chain', () => {
  test('valid, blocked and deprecated zip fixtures verify S12 chain', async ({ page, request }) => {
    const suffix = `e2e_${Date.now()}`;
    execFileSync(fixtureScript, ['--suffix', suffix], { cwd: projectRoot, stdio: 'inherit' });
    const manifest = JSON.parse(fs.readFileSync(path.join(fixtureDist, `manifest-${suffix}.json`), 'utf8'));

    const blockedUpload = await uploadFixture(request, manifest.blocked.zip);
    expect(blockedUpload.upload_id).toBeTruthy();
    expect(['blocked', 'failed']).toContain(String(blockedUpload.status).toLowerCase());
    expect(JSON.stringify(blockedUpload)).toContain('plugin_package_dangerous_file');

    const blockedPromote = await adminPost(request, `/api/v1/admin/plugins/packages/uploads/${blockedUpload.upload_id}/promote`, { force: false });
    expect(blockedPromote.ok()).toBeFalsy();
    await expectErrorCode(blockedPromote, /plugin_package_promote_blocked|plugin_package_upload_invalid_status/);

    const blockedRepo = await listPackages(request, manifest.blocked.plugin_code);
    expect((blockedRepo.items || []).some((item) => item.code === manifest.blocked.plugin_code)).toBeFalsy();

    const validUpload = await uploadFixture(request, manifest.valid.zip);
    expect(validUpload.upload_id).toBeTruthy();
    expect(['ok', 'warning', 'staged']).toContain(String(validUpload.status).toLowerCase());
    expect(validUpload.can_promote).toBeTruthy();
    expect(JSON.stringify(validUpload.migration_plan || validUpload.install_dry_run || validUpload)).not.toContain('will_execute":true');

    const promoted = await promoteUpload(request, validUpload.upload_id);
    expect(promoted.package_path).toBe(`storage/plugins/packages/${manifest.valid.plugin_code}`);
    expect(promoted.message).toContain('仍未安装插件');

    const validDetail = await getUploadDetail(request, validUpload.upload_id);
    expect(validDetail.record.status).toBe('promoted');
    expect(validDetail.record.promoted_path).toBe(promoted.package_path);

    const repo = await listPackages(request, manifest.valid.plugin_code);
    const repoItem = (repo.items || []).find((item) => item.code === manifest.valid.plugin_code);
    expect(repoItem).toBeTruthy();
    expect(repoItem.source_upload_id).toBe(validUpload.upload_id);
    expect(repoItem.promoted_at).toBeTruthy();
    expect(repoItem.checksum_status).toMatch(/ok|warning|missing/);

    const directUploadInstall = await adminPost(request, '/api/v1/admin/plugins/packages/install', {
      path: validUpload.package_path,
    });
    expect(directUploadInstall.ok()).toBeFalsy();
    await expectErrorCode(directUploadInstall, /plugin_package_install_source_invalid|plugin_package_install_dry_run_required/);

    const installWithoutDryRun = await adminPost(request, '/api/v1/admin/plugins/packages/install', {
      path: promoted.package_path,
    });
    expect(installWithoutDryRun.ok()).toBeFalsy();
    await expectErrorCode(installWithoutDryRun, /plugin_package_install_dry_run_required/);

    const dryRun = await dryRunPackage(request, promoted.package_path);
    expect(dryRun.dry_run_id).toBeTruthy();
    expect(dryRun.package.code).toBe(manifest.valid.plugin_code);
    expect(dryRun.package.version).toBe('1.0.0');
    expect(dryRun.migration_plan || []).toEqual(expect.arrayContaining([
      expect.objectContaining({ path: 'migrations/001_init.sql', source: 'migrations', will_execute: false }),
    ]));
    expect(JSON.stringify(dryRun)).not.toContain('001_schema.sql');

    const mismatchInstall = await adminPost(request, '/api/v1/admin/plugins/packages/install', {
      path: promoted.package_path,
      dry_run_id: dryRun.dry_run_id.slice(0, -1),
    });
    expect(mismatchInstall.ok()).toBeFalsy();
    await expectErrorCode(mismatchInstall, /plugin_package_install_dry_run_invalid/);

    const installed = await installPackage(request, promoted.package_path, dryRun.dry_run_id, dryRun.risk_report?.level);
    expect(installed.plugin.code).toBe(manifest.valid.plugin_code);
    expect(installed.plugin.status).toBe('disabled');
    expect(installed.install_result.installed).toBeTruthy();
    expect(installed.install_result.created_migrations).toBeGreaterThanOrEqual(1);

    const installedPlugin = await getPlugin(request, manifest.valid.plugin_code);
    expect(installedPlugin.status).toBe('disabled');

    const installAgain = await adminPost(request, '/api/v1/admin/plugins/packages/install', {
      path: promoted.package_path,
      dry_run_id: dryRun.dry_run_id,
    });
    expect(installAgain.ok()).toBeFalsy();
    await expectErrorCode(installAgain, /plugin_package_already_installed|plugin_package_install_dry_run/);

    const deprecatedUpload = await uploadFixture(request, manifest.deprecated.zip);
    expect(deprecatedUpload.upload_id).toBeTruthy();
    const deprecatedText = JSON.stringify(deprecatedUpload);
    expect(deprecatedText).toContain('001_schema.sql 已废弃');
    expect(deprecatedText).toContain('migrations/001_init.sql');
    expect(deprecatedText).not.toContain('"path":"001_schema.sql","name"');

    const audits = await adminGetJSON(request, `/api/v1/admin/audit-logs?metadata=${encodeURIComponent(validUpload.upload_id)}&page_size=80`);
    const auditText = JSON.stringify(audits);
    expect(auditText).toContain('plugin.package.uploaded');
    expect(auditText).toContain('plugin.package.promote');
    expect(auditText).toContain('plugin.package.install');

    await seedAdminSession(page);
    await page.goto('/admin-next/plugins/packages?tab=uploads');
    await expect(page.getByTestId('plugin-package-upload-lifecycle-page')).toBeVisible();
    await page.getByTestId('upload-filter-keyword').fill(manifest.blocked.plugin_code);
    await page.getByTestId('upload-filter-submit').click();
    await expect(page.getByTestId('upload-list')).toContainText(manifest.blocked.plugin_code);
    await expect(page.getByTestId('upload-list')).toContainText(/已阻断|失败/);

    await page.goto('/admin-next/plugins/packages?tab=install&workspace_tab=repository');
    await expect(page.getByTestId('plugin-package-repository-workspace')).toBeVisible();
    await page.getByTestId('plugin-package-repo-scan').click();
    await expect(page.getByTestId('plugin-package-repo-table')).toContainText(manifest.valid.plugin_code);
  });
});

async function uploadFixture(request, zipName) {
  const filePath = path.join(fixtureDist, zipName);
  const response = await request.post('/api/v1/admin/plugins/packages/upload', {
    headers: adminHeaders(),
    multipart: {
      file: {
        name: zipName,
        mimeType: 'application/zip',
        buffer: fs.readFileSync(filePath),
      },
    },
  });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function promoteUpload(request, uploadID) {
  const response = await adminPost(request, `/api/v1/admin/plugins/packages/uploads/${uploadID}/promote`, { force: false });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function getUploadDetail(request, uploadID) {
  return adminGetJSON(request, `/api/v1/admin/plugins/packages/uploads/${uploadID}`);
}

async function listPackages(request, keyword) {
  return adminGetJSON(request, `/api/v1/admin/plugins/packages?keyword=${encodeURIComponent(keyword)}&page_size=50`);
}

async function dryRunPackage(request, packagePath) {
  const response = await adminPost(request, '/api/v1/admin/plugins/packages/dry-run', { path: packagePath });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function installPackage(request, packagePath, dryRunID, riskLevel) {
  const payload = { path: packagePath, dry_run_id: dryRunID };
  const normalizedRisk = String(riskLevel || '').toLowerCase();
  if (normalizedRisk && normalizedRisk !== 'low') payload.confirm_risk_level = normalizedRisk;
  const response = await adminPost(request, '/api/v1/admin/plugins/packages/install', payload);
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function getPlugin(request, code) {
  const body = await adminGetJSON(request, '/api/v1/admin/plugins');
  const item = (body.items || []).find((plugin) => plugin.code === code);
  if (!item) throw new Error(`plugin not found: ${code}`);
  return item;
}

async function adminPost(request, url, data = {}) {
  return request.post(url, {
    headers: { ...adminHeaders(), 'Content-Type': 'application/json' },
    data,
  });
}

async function adminGetJSON(request, url) {
  const response = await request.get(url, { headers: adminHeaders() });
  if (!response.ok()) throw new Error(await response.text());
  return response.json();
}

async function expectErrorCode(response, pattern) {
  const text = await response.text();
  expect(text).toMatch(pattern);
}
