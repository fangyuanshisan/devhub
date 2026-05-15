import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package zip upload', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('uploads a valid zip into staging and shows scan, risk, manifest and dry-run', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    const card = page.getByTestId('plugin-package-upload-card');
    await expect(card).toBeVisible();
    await expect(card.getByTestId('plugin-package-upload-limits')).toContainText('20MB');
    await expect(card.getByTestId('plugin-package-upload-safety')).toContainText('不会执行插件代码');
    await expect(card.getByTestId('plugin-package-upload-safety')).toContainText('不会执行 SQL');
    await expect(card.getByTestId('plugin-package-upload-safety')).toContainText('不会动态加载前端资产');

    await card.locator('input[type="file"]').setInputFiles({
      name: 'e2e_upload_ok.zip',
      mimeType: 'application/zip',
      buffer: makeZip({
        'manifest.json': manifestJSON('e2e_upload_ok'),
        'README.md': '# E2E Upload\n',
        'config.example.json': '{}',
      }),
    });
    await card.getByTestId('plugin-package-upload-submit').click();

    const result = page.getByTestId('plugin-package-upload-result');
    await expect(result).toBeVisible();
    await expect(result.getByTestId('plugin-package-upload-status')).toContainText(/ok|warning/);
    await expect(result.getByTestId('plugin-package-upload-zip-scan')).toContainText('total_entries');
    await expect(result.getByTestId('plugin-package-upload-file-scan')).toContainText('total_files');
    await expect(result.getByTestId('plugin-package-upload-checksum')).toContainText('manifest_valid');
    await expect(result.getByTestId('plugin-package-upload-risk-report')).toContainText('risk_report');
    await expect(result.getByTestId('plugin-package-upload-manifest-validate')).toBeVisible();
    await expect(result.getByTestId('plugin-package-upload-dry-run')).toBeVisible();
    await expect(result).not.toContainText('确认安装');
  });

  test('shows invalid type code before upload', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    const card = page.getByTestId('plugin-package-upload-card');
    await card.locator('input[type="file"]').setInputFiles({
      name: 'not-a-plugin.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('not zip'),
    });
    await card.getByTestId('plugin-package-upload-submit').click();
    await expect(card.getByTestId('plugin-package-upload-error')).toContainText('plugin_package_upload_invalid_type');
  });

  test('blocks zip slip and dangerous files with code and suggestion', async ({ page }) => {
    await page.goto('/admin-next/plugins/install');
    const card = page.getByTestId('plugin-package-upload-card');

    await card.locator('input[type="file"]').setInputFiles({
      name: 'zip_slip.zip',
      mimeType: 'application/zip',
      buffer: makeZip({ '../manifest.json': '{}' }),
    });
    await card.getByTestId('plugin-package-upload-submit').click();
    await expect(card.getByTestId('plugin-package-upload-error')).toContainText(/plugin_package_zip_slip_detected|plugin_package_zip_entry_path_invalid/);
    await expect(card.getByTestId('plugin-package-upload-error')).toContainText('建议');

    await page.goto('/admin-next/plugins/install');
    const freshCard = page.getByTestId('plugin-package-upload-card');
    await freshCard.locator('input[type="file"]').setInputFiles({
      name: 'dangerous.zip',
      mimeType: 'application/zip',
      buffer: makeZip({
        'manifest.json': manifestJSON('e2e_upload_dangerous'),
        'README.md': '# Dangerous\n',
        'config.example.json': '{}',
        'hack.sh': '#!/bin/sh\necho bad\n',
      }),
    });
    await freshCard.getByTestId('plugin-package-upload-submit').click();
    const result = page.getByTestId('plugin-package-upload-result');
    await expect(result).toBeVisible();
    await expect(result.getByTestId('plugin-package-upload-status')).toContainText('blocked');
    await expect(result.getByTestId('plugin-package-upload-blocked-reasons')).toContainText('plugin_package_dangerous_file');
    await expect(result.getByTestId('plugin-package-upload-blocked-reasons')).toContainText('建议');
  });

  test('promotes an uploaded package into local repository without direct install', async ({ page }) => {
    const code = `e2e_upload_promote_${Date.now()}`;
    await page.goto('/admin-next/plugins/install');
    const card = page.getByTestId('plugin-package-upload-card');
    await card.locator('input[type="file"]').setInputFiles({
      name: `${code}.zip`,
      mimeType: 'application/zip',
      buffer: makeZip({
        'manifest.json': manifestJSON(code),
        'README.md': '# E2E Upload Promote\n',
        'config.example.json': '{}',
      }),
    });
    await card.getByTestId('plugin-package-upload-submit').click();
    await expect(page.getByTestId('plugin-package-upload-result')).toBeVisible();
    await page.getByTestId('plugin-package-upload-promote').click();
    await expect(page.getByTestId('plugin-package-repo-table')).toContainText(code);
  });
});

function manifestJSON(code) {
  return JSON.stringify({
    code,
    name: `Upload ${code}`,
    version: '1.0.0',
    compatible_core_version: '>=1.4.0',
    is_system: false,
    content_types: [`${code}_item`],
    content_type_definitions: [{
      type: `${code}_item`,
      name: 'Upload Item',
      plugin_code: code,
      create_permission: `${code}.item.create`,
      edit_permission: `${code}.item.edit`,
      delete_permission: `${code}.item.delete`,
      audit_permission: `${code}.item.audit`,
      default_status: 'draft',
      allow_comment: true,
      allow_like: true,
      allow_favorite: true,
      seo_type: 'Article',
    }],
    permissions: [
      { code: `${code}.item.create`, name: 'create', scope: 'community' },
      { code: `${code}.item.edit`, name: 'edit', scope: 'own' },
      { code: `${code}.item.delete`, name: 'delete', scope: 'own' },
      { code: `${code}.item.audit`, name: 'audit', scope: 'community' },
    ],
    menus: [{ code: `${code}.admin`, title: 'Upload', path: `/admin-next/${code}`, location: 'admin', area: 'admin', permission: `${code}.item.audit` }],
    routes: [{ area: 'admin', method: 'GET', path: `/api/v1/admin/${code}`, handler: 'reserved', auth: 'admin', permission: `${code}.item.audit` }],
  });
}

function makeZip(entries) {
  const chunks = [];
  const central = [];
  let offset = 0;
  for (const [name, content] of Object.entries(entries)) {
    const nameBuf = Buffer.from(name);
    const data = Buffer.isBuffer(content) ? content : Buffer.from(String(content));
    const crc = crc32(data);
    const local = Buffer.alloc(30);
    local.writeUInt32LE(0x04034b50, 0);
    local.writeUInt16LE(20, 4);
    local.writeUInt16LE(0, 6);
    local.writeUInt16LE(0, 8);
    local.writeUInt32LE(crc, 14);
    local.writeUInt32LE(data.length, 18);
    local.writeUInt32LE(data.length, 22);
    local.writeUInt16LE(nameBuf.length, 26);
    chunks.push(local, nameBuf, data);

    const head = Buffer.alloc(46);
    head.writeUInt32LE(0x02014b50, 0);
    head.writeUInt16LE(20, 4);
    head.writeUInt16LE(20, 6);
    head.writeUInt32LE(crc, 16);
    head.writeUInt32LE(data.length, 20);
    head.writeUInt32LE(data.length, 24);
    head.writeUInt16LE(nameBuf.length, 28);
    head.writeUInt32LE(offset, 42);
    central.push(head, nameBuf);
    offset += local.length + nameBuf.length + data.length;
  }
  const centralStart = offset;
  const centralSize = central.reduce((sum, b) => sum + b.length, 0);
  const end = Buffer.alloc(22);
  end.writeUInt32LE(0x06054b50, 0);
  end.writeUInt16LE(Object.keys(entries).length, 8);
  end.writeUInt16LE(Object.keys(entries).length, 10);
  end.writeUInt32LE(centralSize, 12);
  end.writeUInt32LE(centralStart, 16);
  return Buffer.concat([...chunks, ...central, end]);
}

function crc32(buf) {
  let crc = -1;
  for (const byte of buf) {
    crc = (crc >>> 8) ^ CRC_TABLE[(crc ^ byte) & 0xff];
  }
  return (crc ^ -1) >>> 0;
}

const CRC_TABLE = Array.from({ length: 256 }, (_, n) => {
  let c = n;
  for (let k = 0; k < 8; k += 1) c = (c & 1) ? (0xedb88320 ^ (c >>> 1)) : (c >>> 1);
  return c >>> 0;
});
