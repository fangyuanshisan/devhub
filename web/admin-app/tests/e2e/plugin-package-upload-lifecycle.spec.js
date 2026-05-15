import { expect, test } from '@playwright/test';
import { seedAdminSession } from './helpers/auth.js';

test.describe('plugin package upload lifecycle', () => {
  test.beforeEach(async ({ page }) => {
    await seedAdminSession(page);
  });

  test('uploads, rescans, approves, promotes, cancels/deletes and blocks dangerous uploads', async ({ page }) => {
    const code = `e2e_upload_lifecycle_${Date.now()}`;
    await page.goto('/admin-next/plugins/packages/uploads');
    await expect(page.getByTestId('plugin-package-upload-lifecycle-page')).toContainText('上传包只是 staging');
    await expect(page.getByTestId('plugin-package-upload-lifecycle-page')).toContainText('不会执行插件代码');
    await expect(page.getByTestId('plugin-package-upload-lifecycle-page')).not.toContainText('远程市场');
    await expect(page.getByTestId('plugin-package-upload-lifecycle-page')).not.toContainText('动态加载入口');

    await page.getByTestId('upload-zip-picker').locator('input[type="file"]').setInputFiles({
      name: `${code}.zip`,
      mimeType: 'application/zip',
      buffer: makeZip({
        'manifest.json': manifestJSON(code),
        'README.md': '# Lifecycle\n',
        'config.example.json': '{}',
      }),
    });
    await page.getByTestId('upload-zip-submit').click();
    await expect(page.getByTestId('upload-detail-drawer')).toBeVisible();
    await expect(page.getByTestId('upload-list')).toContainText(code);
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('staged');
    await expect(page.getByTestId('upload-risk-report')).toContainText('summary');
    await expect(page.getByTestId('upload-checksum')).toContainText('signature');
    await expect(page.getByTestId('upload-manifest-validation')).toContainText('valid');
    await expect(page.getByTestId('upload-dry-run')).toContainText('dependency_summary');

    await page.getByTestId('upload-rescan').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('staged');
    await page.getByTestId('upload-submit-approval').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('approval_pending');
    await page.getByTestId('upload-approve').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('approved');
    await page.getByTestId('upload-promote').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('promoted');
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('不会安装插件');
    await expect(page.getByTestId('plugin-package-upload-lifecycle-page')).not.toContainText('上传后直接安装');

    const cancelCode = `${code}_cancel`;
    await page.getByRole('button', { name: 'Close' }).click().catch(() => {});
    await page.getByTestId('upload-zip-picker').locator('input[type="file"]').setInputFiles({
      name: `${cancelCode}.zip`,
      mimeType: 'application/zip',
      buffer: makeZip({
        'manifest.json': manifestJSON(cancelCode),
        'README.md': '# Cancel\n',
        'config.example.json': '{}',
      }),
    });
    await page.getByTestId('upload-zip-submit').click();
    await page.getByTestId('upload-cancel').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('canceled');
    await page.getByTestId('upload-delete').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('deleted');

    await page.getByRole('button', { name: 'Close' }).click().catch(() => {});
    await page.getByTestId('upload-zip-picker').locator('input[type="file"]').setInputFiles({
      name: `${code}_dangerous.zip`,
      mimeType: 'application/zip',
      buffer: makeZip({
        'manifest.json': manifestJSON(`${code}_dangerous`),
        'README.md': '# Dangerous\n',
        'hack.sh': '#!/bin/sh\necho bad\n',
      }),
    });
    await page.getByTestId('upload-zip-submit').click();
    await expect(page.getByTestId('upload-detail-drawer')).toContainText('blocked');
    await expect(page.getByTestId('upload-promote')).toBeDisabled();
  });
});

function manifestJSON(code) {
  return JSON.stringify({
    code,
    name: `Lifecycle ${code}`,
    version: '1.0.0',
    compatible_core_version: '>=1.4.0',
    is_system: false,
    content_types: [`${code}_item`],
    content_type_definitions: [{
      type: `${code}_item`,
      name: 'Lifecycle Item',
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
    menus: [{ code: `${code}.admin`, title: 'Lifecycle', path: `/admin-next/${code}`, location: 'admin', area: 'admin', permission: `${code}.item.audit` }],
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
  for (const byte of buf) crc = (crc >>> 8) ^ CRC_TABLE[(crc ^ byte) & 0xff];
  return (crc ^ -1) >>> 0;
}

const CRC_TABLE = Array.from({ length: 256 }, (_, n) => {
  let c = n;
  for (let k = 0; k < 8; k += 1) c = (c & 1) ? (0xedb88320 ^ (c >>> 1)) : (c >>> 1);
  return c >>> 0;
});
