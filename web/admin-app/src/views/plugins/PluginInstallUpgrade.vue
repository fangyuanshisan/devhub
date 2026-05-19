<template>
  <section class="plugin-page" data-testid="plugin-install-page">
    <div class="plugin-page-header">
      <div>
        <div class="eyebrow">安装与治理</div>
        <h2>安装升级</h2>
        <p class="muted">Manifest 校验、预检、安装与升级入口。不会执行第三方代码，不支持远程下载与动态加载。</p>
      </div>
      <div class="primary-actions">
        <el-button type="primary" plain data-testid="plugin-manifest-validate" @click="openManifestDialog('validate')">{{ t('plugin.ops.validateManifest') }}</el-button>
        <el-button type="primary" plain data-testid="plugin-manifest-dry-run" @click="openManifestDialog('dry-run')">{{ t('plugin.ops.dryRun') }}</el-button>
        <el-button type="success" plain data-testid="plugin-manifest-install" @click="openManifestDialog('install')">{{ t('plugin.ops.install') }}</el-button>
      </div>
    </div>

    <el-alert type="info" show-icon :closable="false" class="mb" title="限制：不自动安装依赖、不远程下载、不动态加载、不执行第三方代码。" />

    <el-tabs v-model="packageTab" class="package-workspace-tabs" data-testid="plugin-package-workspace-tabs">
      <el-tab-pane label="本地仓库" name="repository" />
      <el-tab-pane label="初始化插件包" name="template" />
      <el-tab-pane label="上传 zip" name="upload" />
    </el-tabs>

    <el-alert
      type="info"
      show-icon
      :closable="false"
      class="mb"
      title="插件包治理分层：远程索引不等于自动安装，暂存区不等于已安装，预检通过不等于可启用，启用前检查通过后才允许进入启用流程，升级不会执行第三方代码。"
    />

    <el-card v-if="packageTab === 'template'" shadow="never" class="mb" data-testid="plugin-package-template-card">
      <template #header>
        <div class="card-head">
          <strong>初始化插件包</strong>
          <span class="muted">在 storage/plugins/packages/{code} 下生成声明型模板，并自动执行插件包预检。</span>
        </div>
      </template>

      <el-form :model="templateForm" label-width="120px" class="template-form" data-testid="plugin-package-template-form">
        <div class="template-grid">
          <el-form-item label="插件编码">
            <el-input v-model="templateForm.code" data-testid="plugin-template-code" placeholder="demo_notice" clearable />
          </el-form-item>
          <el-form-item label="插件名称">
            <el-input v-model="templateForm.name" data-testid="plugin-template-name" placeholder="示例公告插件" clearable />
          </el-form-item>
          <el-form-item label="内容类型">
            <el-input v-model="templateForm.content_type" data-testid="plugin-template-content-type" placeholder="notice" clearable />
          </el-form-item>
          <el-form-item label="内容类型名">
            <el-input v-model="templateForm.content_name" data-testid="plugin-template-content-name" placeholder="公告" clearable />
          </el-form-item>
          <el-form-item label="描述" class="wide">
            <el-input v-model="templateForm.description" data-testid="plugin-template-description" placeholder="声明型插件描述" clearable />
          </el-form-item>
          <el-form-item label="作者">
            <el-input v-model="templateForm.author" data-testid="plugin-template-author" placeholder="DevHub Team" clearable />
          </el-form-item>
        </div>
        <div class="filter-actions" style="justify-content: space-between">
          <div class="template-switches">
            <el-checkbox v-model="templateForm.with_config" data-testid="plugin-template-with-config">配置模型</el-checkbox>
            <el-checkbox v-model="templateForm.with_hooks" data-testid="plugin-template-with-hooks">Hook 声明</el-checkbox>
            <el-checkbox v-model="templateForm.with_migration" data-testid="plugin-template-with-migration">migration 声明</el-checkbox>
          </div>
          <div class="filter-actions">
            <el-button plain :loading="templateLoading" data-testid="plugin-template-preview" @click="previewPackageTemplate">预览</el-button>
            <el-button type="primary" :loading="templateLoading" data-testid="plugin-template-create" @click="createPackageTemplate">初始化插件包</el-button>
          </div>
        </div>
      </el-form>

      <el-alert
        type="info"
        show-icon
        :closable="false"
        class="mb"
        title="后台初始化不会生成 registry.example.go；registry 接入说明会写入 docs/registry-example.md，避免 .go 文件触发危险文件阻断。"
      />
      <el-alert v-if="templateError" type="error" show-icon :closable="false" class="mb" :title="templateError" />

      <el-card v-if="templateResult" shadow="never" class="mb" data-testid="plugin-template-result">
        <template #header>
          <div class="card-head">
            <strong>{{ templateResult.template?.code || '-' }}</strong>
            <el-tag :type="templateResult.status === 'blocked' ? 'danger' : templateResult.status === 'warning' ? 'warning' : 'success'" effect="plain">
              {{ genericStatusLabel(templateResult.status || 'preview') }}
            </el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border class="mb">
          <el-descriptions-item label="插件包路径">{{ templateResult.template?.package_path || templateResult.template?.output_dir || '-' }}</el-descriptions-item>
          <el-descriptions-item label="预检状态">{{ genericStatusLabel(templateResult.dry_run?.status) }}</el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <el-tag :type="riskLevelType(templateResult.dry_run?.risk_report?.level)" effect="plain">
              {{ packageRiskLabel(templateResult.dry_run?.risk_report?.level) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Manifest 是否有效">{{ templateResult.dry_run?.manifest_validation?.valid ?? '-' }}</el-descriptions-item>
        </el-descriptions>
        <div class="result-grid mb">
          <div class="result-box">
            <h4>将生成文件</h4>
            <div class="tag-wrap">
              <el-tag v-for="file in templateResult.template?.files || []" :key="file" effect="plain">{{ file }}</el-tag>
            </div>
          </div>
          <div class="result-box">
            <h4>警告 / 错误</h4>
            <ul class="result-list">
              <li v-for="(item, idx) in templateWarnings" :key="`tpl-w-${idx}`">{{ item }}</li>
              <li v-for="(item, idx) in templateErrors" :key="`tpl-e-${idx}`">{{ item }}</li>
              <li v-if="!templateWarnings.length && !templateErrors.length" class="muted">-</li>
            </ul>
          </div>
        </div>
        <div class="filter-actions" style="justify-content: flex-end">
          <el-button v-if="templateResult.template?.package_path" @click="useTemplatePathForDryRun">填入预检路径</el-button>
          <el-button
            v-if="templateResult.dry_run && String(templateResult.dry_run.status || '').toLowerCase() !== 'blocked'"
            type="primary"
            plain
            data-testid="plugin-template-submit-approval"
            @click="submitTemplateInstallApproval"
          >
            提交安装审批
          </el-button>
        </div>
      </el-card>
    </el-card>

    <el-card v-if="packageTab === 'upload'" shadow="never" class="mb" data-testid="plugin-package-upload-card">
      <template #header>
        <div class="card-head">
          <strong>上传插件包 zip</strong>
          <span class="muted">上传后只进入安全沙箱，不会安装插件。</span>
        </div>
      </template>
      <div class="upload-grid">
        <div>
          <el-upload
            :auto-upload="false"
            :limit="1"
            accept=".zip"
            :on-change="onZipUploadChange"
            :on-remove="onZipUploadRemove"
            data-testid="plugin-package-upload-input"
          >
            <el-button type="primary" plain>选择 zip 文件</el-button>
          </el-upload>
          <div class="muted" style="margin-top: 8px" data-testid="plugin-package-upload-limits">
            上传上限 20MB；解压后总大小 50MB；单文件 5MB；文件数 300；目录深度 8；内嵌压缩包、symlink、hardlink、特殊设备文件、路径穿越会被阻断。
          </div>
        </div>
        <div class="result-box">
          <h4>安全边界</h4>
          <ul class="result-list" data-testid="plugin-package-upload-safety">
            <li>zip 只会进入 storage/plugins/uploads 与 staging/quarantine 沙箱。</li>
            <li>不会执行插件代码，不会执行 SQL，不会动态加载前端资产。</li>
            <li>安装仍需预检 / 审批 / 安装流程；本区域不显示直接安装按钮。</li>
            <li>不支持远程市场、远程下载、在线更新或脚本沙箱。</li>
          </ul>
        </div>
      </div>
      <div class="filter-actions" style="justify-content: flex-end; margin-top: 12px">
        <el-button type="primary" :loading="zipUploadLoading" data-testid="plugin-package-upload-submit" @click="submitZipUpload">上传并扫描</el-button>
      </div>
      <el-alert v-if="zipUploadError" type="error" show-icon :closable="false" class="mb" :title="zipUploadError" data-testid="plugin-package-upload-error" />
      <el-card v-if="zipUploadResult" shadow="never" class="mb" data-testid="plugin-package-upload-result">
        <template #header>
          <div class="card-head">
            <strong>{{ zipUploadResult.upload_id || '-' }}</strong>
            <el-tag :type="zipUploadResult.status === 'blocked' ? 'danger' : zipUploadResult.status === 'warning' ? 'warning' : 'success'" effect="plain" data-testid="plugin-package-upload-status">
            {{ genericStatusLabel(zipUploadResult.status) }}
            </el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border class="mb" data-testid="plugin-package-upload-info">
          <el-descriptions-item label="文件名">{{ zipUploadResult.filename || '-' }}</el-descriptions-item>
          <el-descriptions-item label="暂存路径">{{ zipUploadResult.staging_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="插件包路径">{{ zipUploadResult.package_path || '-' }}</el-descriptions-item>
          <el-descriptions-item label="可转入仓库">{{ zipUploadResult.can_promote ? '是' : '否' }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions :column="3" border class="mb" data-testid="plugin-package-upload-zip-scan">
          <el-descriptions-item label="条目数">{{ zipUploadResult.zip_scan?.total_entries ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="压缩大小">{{ zipUploadResult.zip_scan?.compressed_size ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="解压大小">{{ zipUploadResult.zip_scan?.uncompressed_size ?? 0 }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions :column="3" border class="mb" data-testid="plugin-package-upload-file-scan">
          <el-descriptions-item label="文件数">{{ zipUploadResult.file_scan?.total_files ?? 0 }}</el-descriptions-item>
          <el-descriptions-item label="允许文件">{{ (zipUploadResult.file_scan?.allowed_files || []).length }}</el-descriptions-item>
          <el-descriptions-item label="危险文件">{{ (zipUploadResult.file_scan?.dangerous_files || []).length }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions :column="3" border class="mb" data-testid="plugin-package-upload-checksum">
          <el-descriptions-item label="校验状态">
            <el-tag :type="checksumStatusType(zipUploadResult.checksum?.status)" effect="plain">{{ genericStatusLabel(zipUploadResult.checksum?.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="签名">{{ genericStatusLabel(zipUploadResult.signature?.verification_status || 'missing') }}</el-descriptions-item>
          <el-descriptions-item label="Manifest 有效">{{ zipUploadResult.manifest_validation?.valid ?? false }}</el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-if="zipUploadResult.risk_report"
          :type="riskLevelType(zipUploadResult.risk_report.level)"
          show-icon
          :closable="false"
          class="mb"
          data-testid="plugin-package-upload-risk-report"
          :title="`风险报告：${packageRiskLabel(zipUploadResult.risk_report.level)} / score=${zipUploadResult.risk_report.score ?? 0}`"
        >
          <div class="muted">{{ zipUploadResult.risk_report.summary || '-' }}</div>
        </el-alert>
        <div class="result-grid mb">
          <div class="result-box" data-testid="plugin-package-upload-manifest-validate">
            <h4>Manifest 校验</h4>
            <pre class="json-box compact">{{ formatJSON(zipUploadResult.manifest_validation || {}) }}</pre>
          </div>
          <div class="result-box" data-testid="plugin-package-upload-dry-run">
            <h4>安装预检</h4>
            <pre class="json-box compact">{{ formatJSON(zipUploadResult.install_dry_run?.install_preview || zipUploadResult.install_dry_run || {}) }}</pre>
          </div>
        </div>
        <div v-if="(zipUploadResult.errors || []).length || (zipUploadResult.risk_report?.items || []).length" class="mb" data-testid="plugin-package-upload-blocked-reasons">
          <h4>阻断原因 / 修复建议</h4>
          <ul class="result-list">
            <li v-for="(item, idx) in zipUploadResult.errors || []" :key="`upload-err-${idx}`">{{ item }}</li>
            <li v-for="(item, idx) in zipUploadResult.risk_report?.items || []" :key="`upload-risk-${idx}`">
              {{ pluginReasonText(item.code) }}：{{ item.message }}；建议：{{ item.suggestion || '-' }}
            </li>
          </ul>
        </div>
        <div class="filter-actions" style="justify-content: flex-end">
          <el-button v-if="zipUploadResult.package_path" data-testid="plugin-package-upload-fill-dryrun" @click="useUploadedPackageForDryRun">填入预检路径</el-button>
          <el-button v-if="zipUploadResult.can_promote" type="success" plain :loading="zipPromoteLoading" data-testid="plugin-package-upload-promote" @click="promoteUploadedPackage">
            转入本地仓库
          </el-button>
        </div>
      </el-card>
    </el-card>

    <div v-if="packageTab === 'repository'" data-testid="plugin-package-repository-workspace">
    <div class="filter-panel mb">
      <div>
        <strong>升级目标插件</strong>
        <span class="muted">升级预览/升级需要选择一个已安装插件。</span>
      </div>
      <div class="filter-actions">
        <div data-testid="plugin-upgrade-target-select">
          <el-select v-model="targetCode" filterable clearable placeholder="选择插件" style="min-width: 260px">
            <el-option v-for="p in items" :key="p.code" :label="`${p.name} (${p.code})`" :value="p.code" />
          </el-select>
        </div>
        <el-button type="warning" plain :disabled="!targetCode" data-testid="plugin-upgrade-preview-selected" @click="openManifestDialog('upgrade-dry-run', findPlugin(targetCode))">{{ t('plugin.ops.upgradePreview') }}</el-button>
        <el-button type="danger" plain :disabled="!targetCode" data-testid="plugin-upgrade-selected" @click="openManifestDialog('upgrade', findPlugin(targetCode))">{{ t('plugin.ops.upgrade') }}</el-button>
      </div>
    </div>

    <div class="filter-panel mb" data-testid="plugin-package-dryrun-panel">
      <div>
          <strong>本地插件包预检</strong>
        <div class="muted" style="margin-top: 6px">
          只做安全读取、文件扫描、manifest 校验与安装预览；不会安装插件，不执行插件代码，不执行 SQL，不动态加载前端资产。
        </div>
      </div>
      <div class="filter-actions">
        <el-input v-model="packagePath" data-testid="plugin-package-path-input" placeholder="例如：examples/plugins/demo_notice" clearable />
        <el-button type="primary" :loading="packageLoading" data-testid="plugin-package-dry-run" @click="runPackageDryRun">扫描 / 预检</el-button>
        <el-button v-if="packageResult" data-testid="plugin-package-clear" @click="clearPackageResult">清空</el-button>
      </div>
    </div>

    <div class="filter-panel mb" data-testid="plugin-package-repo-panel">
      <div>
        <strong>本地插件仓库</strong>
        <div class="muted" style="margin-top: 6px">
          扫描仓库目录下的一级子目录作为插件包（只读扫描/校验/预览）；不会安装插件，不执行代码/SQL，不动态加载前端资产。
        </div>
      </div>
      <div class="filter-actions">
        <el-input v-model="repoRoot" data-testid="plugin-package-repo-root" placeholder="默认：storage/plugins/packages" clearable />
        <el-button type="primary" plain :loading="repoLoading" data-testid="plugin-package-repo-scan" @click="scanRepository(true)">扫描</el-button>
        <el-button :loading="repoLoading" data-testid="plugin-package-repo-refresh" @click="scanRepository()">刷新</el-button>
      </div>
    </div>

    <div v-if="repoSummary" class="mb" data-testid="plugin-package-repo-summary">
      <el-tag effect="plain">总数 {{ repoSummary.total ?? 0 }}</el-tag>
      <el-tag type="success" effect="plain">正常 {{ repoSummary.ok ?? 0 }}</el-tag>
      <el-tag type="warning" effect="plain">有警告 {{ repoSummary.warning ?? 0 }}</el-tag>
      <el-tag type="danger" effect="plain">已阻断 {{ repoSummary.blocked ?? 0 }}</el-tag>
      <el-tag type="info" effect="plain">无效 {{ repoSummary.invalid ?? 0 }}</el-tag>
    </div>

    <el-alert v-if="repoError" type="error" show-icon :closable="false" class="mb" :title="repoError" />

    <el-card shadow="never" class="mb" data-testid="plugin-package-repo-card">
      <template #header>
        <div class="card-head">
          <strong>已发现插件包</strong>
          <span class="muted">（第 {{ repoPage }} 页 / 共 {{ repoTotal }} 条）</span>
        </div>
      </template>

      <div class="filter-actions mb" style="gap: 10px" data-testid="plugin-package-repo-filters">
        <el-input v-model="repoKeyword" placeholder="关键词：编码 / 名称 / 路径" clearable style="max-width: 320px" @change="scanRepository(true)" />
        <el-select v-model="repoStatus" placeholder="状态" clearable style="width: 160px" @change="scanRepository(true)">
          <el-option label="全部" value="all" />
          <el-option label="正常" value="ok" />
          <el-option label="有警告" value="warning" />
          <el-option label="已阻断" value="blocked" />
          <el-option label="无效" value="invalid" />
        </el-select>
        <el-select v-model="repoRiskLevel" placeholder="风险等级" clearable style="width: 160px" @change="scanRepository(true)">
          <el-option label="低风险" value="low" />
          <el-option label="中风险" value="medium" />
          <el-option label="高风险" value="high" />
          <el-option label="已阻断" value="blocked" />
        </el-select>
        <el-select v-model="repoChecksumStatus" placeholder="checksum" clearable style="width: 170px" @change="scanRepository(true)">
          <el-option label="ok" value="ok" />
          <el-option label="warning" value="warning" />
          <el-option label="failed" value="failed" />
          <el-option label="missing" value="missing" />
        </el-select>
        <el-select v-model="repoManifestValid" placeholder="Manifest 是否有效" clearable style="width: 170px" @change="scanRepository(true)">
          <el-option label="是" value="true" />
          <el-option label="否" value="false" />
        </el-select>
      </div>

      <el-table :data="repoItems" border size="small" v-loading="repoLoading" data-testid="plugin-package-repo-table">
        <el-table-column prop="code" label="插件编码" min-width="150" />
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column prop="version" label="版本" width="110" />
        <el-table-column prop="path" label="路径" min-width="260" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'blocked' ? 'danger' : row.status === 'warning' ? 'warning' : row.status === 'invalid' ? 'info' : 'success'" effect="plain">
              {{ genericStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="risk_level" label="风险" width="110">
          <template #default="{ row }">
            <el-tag :type="riskLevelType(row.risk_level)" effect="plain">{{ packageRiskLabel(row.risk_level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="checksum_status" label="校验状态" width="120">
          <template #default="{ row }">
            <el-tag :type="checksumStatusType(row.checksum_status)" effect="plain">{{ genericStatusLabel(row.checksum_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源上传包" min-width="180">
          <template #default="{ row }">
            <div v-if="row.source_upload_id">
              <div class="mono small">{{ row.source_upload_id }}</div>
              <div class="muted small">{{ row.promoted_at || '已转入本地仓库' }}</div>
            </div>
            <span v-else class="muted">本地包</span>
          </template>
        </el-table-column>
        <el-table-column label="签名" width="190">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 6px; flex-wrap: wrap">
              <el-tag
                :type="signatureTrustType(row?.signature?.trust_status || (row?.signature_found ? 'unknown' : 'unsigned'))"
                effect="plain"
                data-testid="plugin-package-repo-signature-trust"
              >
                {{ trustLevelLabel(row?.signature?.trust_status || (row?.signature_found ? 'unknown' : 'unsigned')) }}
              </el-tag>
              <el-tag
                :type="signatureVerifyType(row?.signature?.verification_status)"
                effect="plain"
                data-testid="plugin-package-repo-signature-verify"
              >
                {{ genericStatusLabel(row?.signature?.verification_status || (row?.signature_found ? '-' : 'missing')) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="total_files" label="文件数" width="90" />
        <el-table-column prop="total_size" label="大小" width="110" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <span data-testid="plugin-package-repo-detail-btn" style="display: inline-flex" @click="openRepoDetail(row)">
              <el-button link type="primary">详情</el-button>
            </span>
            <span data-testid="plugin-package-repo-dryrun-btn" style="display: inline-flex" @click="dryRunRepoPackage(row)">
              <el-button link type="warning">预检</el-button>
            </span>
            <span data-testid="plugin-package-repo-install-btn" style="display: inline-flex" @click="openRepoInstall(row)">
              <el-button
                link
                type="success"
                @click.stop="openRepoInstall(row)"
              >
                安装
              </el-button>
            </span>
          </template>
        </el-table-column>
      </el-table>

      <div class="filter-actions" style="justify-content: flex-end; margin-top: 10px">
        <el-pagination
          v-model:current-page="repoPage"
          v-model:page-size="repoPageSize"
          layout="prev, pager, next, sizes, total"
          :total="repoTotal"
          :page-sizes="[10, 20, 50, 100]"
          @current-change="scanRepository()"
          @size-change="scanRepository(true)"
        />
      </div>
    </el-card>

    <el-alert v-if="packageError" type="error" show-icon :closable="false" class="mb" :title="packageError" />

    <el-card v-if="packageResult" class="mb" shadow="never" data-testid="plugin-package-result">
      <template #header>
        <div class="card-head">
          <strong>预检结果</strong>
          <el-tag :type="packageResult.status === 'blocked' ? 'danger' : packageResult.status === 'warning' ? 'warning' : 'success'" effect="plain">
            {{ genericStatusLabel(packageResult.status) }}
          </el-tag>
        </div>
      </template>

      <el-alert
        v-if="packageResult.risk_report && packageResult.risk_report.level"
        :type="riskLevelType(packageResult.risk_report.level)"
        show-icon
        :closable="false"
        class="mb"
        data-testid="plugin-package-risk-alert"
      >
        <template #title>
          <span>风险等级：</span>
          <el-tag :type="riskLevelType(packageResult.risk_report.level)" effect="plain" data-testid="plugin-package-risk-level">
            {{ packageRiskLabel(packageResult.risk_report.level) }}
          </el-tag>
          <span class="muted" style="margin-left: 8px" data-testid="plugin-package-risk-score">score={{ packageResult.risk_report.score ?? 0 }}</span>
        </template>
        <div class="muted" data-testid="plugin-package-risk-summary">{{ packageResult.risk_report.summary || '-' }}</div>
      </el-alert>

      <el-alert
        v-if="(packageResult.errors || []).length"
        type="error"
        show-icon
        :closable="false"
        class="mb"
        title="阻断原因"
      >
        <ul class="result-list">
          <li v-for="(item, idx) in packageResult.errors" :key="`pkg-err-${idx}`">{{ item }}</li>
        </ul>
      </el-alert>

      <el-alert
        v-if="(packageResult.warnings || []).length"
        type="warning"
        show-icon
        :closable="false"
        class="mb"
        title="警告"
      >
        <ul class="result-list">
          <li v-for="(item, idx) in packageResult.warnings" :key="`pkg-warn-${idx}`">{{ item }}</li>
        </ul>
      </el-alert>

      <el-descriptions :column="2" border class="mb" data-testid="plugin-package-info">
        <el-descriptions-item label="路径">{{ packageResult.package?.path || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目录">{{ packageResult.package?.dir_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="插件编码">{{ packageResult.package?.code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="版本">{{ packageResult.package?.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="阻断原因">{{ pluginReasonText(packageResult.blocked_code) }}</el-descriptions-item>
        <el-descriptions-item label="Manifest">{{ yesNo(packageResult.package?.manifest_found) }}</el-descriptions-item>
        <el-descriptions-item label="README">{{ yesNo(packageResult.package?.readme_found) }}</el-descriptions-item>
        <el-descriptions-item label="配置示例">{{ yesNo(packageResult.package?.config_example_found) }}</el-descriptions-item>
        <el-descriptions-item label="校验文件">{{ yesNo(packageResult.package?.checksum_found) }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions :column="3" border class="mb" data-testid="plugin-package-file-scan">
        <el-descriptions-item label="文件数">{{ packageResult.file_scan?.total_files ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="总大小">{{ packageResult.file_scan?.total_size ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="允许文件">{{ (packageResult.file_scan?.allowed_files || []).length }}</el-descriptions-item>
        <el-descriptions-item label="未知文件">{{ (packageResult.file_scan?.unknown_files || []).length }}</el-descriptions-item>
        <el-descriptions-item label="危险文件">{{ (packageResult.file_scan?.dangerous_files || []).length }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions v-if="packageResult.checksum" :column="3" border class="mb" data-testid="plugin-package-checksum">
        <el-descriptions-item label="算法">{{ packageResult.checksum.algorithm || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="checksumStatusType(packageResult.checksum.status)" effect="plain" data-testid="plugin-package-checksum-status">
            {{ genericStatusLabel(packageResult.checksum.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="匹配">{{ (packageResult.checksum.matched || []).length }}</el-descriptions-item>
        <el-descriptions-item label="不匹配">{{ (packageResult.checksum.mismatched || []).length }}</el-descriptions-item>
        <el-descriptions-item label="缺失">{{ (packageResult.checksum.missing || []).length }}</el-descriptions-item>
        <el-descriptions-item label="额外文件">{{ (packageResult.checksum.extra || []).length }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions v-if="packageResult.signature" :column="3" border class="mb" data-testid="plugin-package-signature">
        <el-descriptions-item label="信任状态">
          <el-tag :type="signatureTrustType(packageResult.signature.trust_status)" effect="plain" data-testid="plugin-package-signature-trust">
            {{ trustLevelLabel(packageResult.signature.trust_status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="验签状态">
          <el-tag :type="signatureVerifyType(packageResult.signature.verification_status)" effect="plain" data-testid="plugin-package-signature-verify">
            {{ genericStatusLabel(packageResult.signature.verification_status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="算法">{{ packageResult.signature.algorithm || '-' }}</el-descriptions-item>
        <el-descriptions-item label="签名载荷">{{ packageResult.signature.payload || '-' }}</el-descriptions-item>
        <el-descriptions-item label="指纹">{{ packageResult.signature.fingerprint || '-' }}</el-descriptions-item>
        <el-descriptions-item label="发布者 ID">{{ packageResult.signature.publisher_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="public_key_id">{{ packageResult.signature.public_key_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="签名文件数">{{ packageResult.signature.signed_files_count ?? 0 }}</el-descriptions-item>
      </el-descriptions>

      <div v-if="packageResult.signature && (packageResult.signature.unsigned_files || []).length" class="mb" data-testid="plugin-package-signature-unsigned-files">
        <h4 style="margin: 0 0 8px">未签名文件</h4>
        <div class="tag-wrap">
          <el-tag v-for="item in packageResult.signature.unsigned_files || []" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
        </div>
      </div>

      <div v-if="packageResult.signature && (packageResult.signature.messages || []).length" class="mb" data-testid="plugin-package-signature-messages">
        <h4 style="margin: 0 0 8px">签名提示</h4>
        <ul class="result-list">
          <li v-for="(item, idx) in packageResult.signature.messages || []" :key="`pkg-sig-msg-${idx}`">{{ item }}</li>
        </ul>
      </div>

      <div v-if="packageResult.checksum && (packageResult.checksum.mismatched || []).length" class="mb" data-testid="plugin-package-checksum-mismatched">
        <h4 style="margin: 0 0 8px">校验不匹配</h4>
        <el-table :data="packageResult.checksum.mismatched" size="small" border>
          <el-table-column prop="path" label="路径" min-width="240" />
          <el-table-column prop="expected" label="预期值" min-width="240" />
          <el-table-column prop="actual" label="实际值" min-width="240" />
        </el-table>
      </div>

      <div v-if="packageResult.checksum && (packageResult.checksum.extra || []).length" class="mb" data-testid="plugin-package-checksum-extra">
        <h4 style="margin: 0 0 8px">额外文件（未被校验覆盖）</h4>
        <div class="tag-wrap">
          <el-tag v-for="item in packageResult.checksum.extra || []" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
        </div>
      </div>

      <div class="result-grid mb">
        <div class="result-box">
          <h4>危险文件</h4>
          <div class="tag-wrap">
            <el-tag v-for="item in packageResult.file_scan?.dangerous_files || []" :key="item.path" type="danger" effect="plain">
              {{ item.path }}
            </el-tag>
            <span v-if="!(packageResult.file_scan?.dangerous_files || []).length" class="muted">-</span>
          </div>
        </div>
        <div class="result-box">
          <h4>未知文件</h4>
          <div class="tag-wrap">
            <el-tag v-for="item in packageResult.file_scan?.unknown_files || []" :key="item.path" type="warning" effect="plain">
              {{ item.path }}
            </el-tag>
            <span v-if="!(packageResult.file_scan?.unknown_files || []).length" class="muted">-</span>
          </div>
        </div>
      </div>

      <div v-if="packageResult.risk_report && (packageResult.risk_report.items || []).length" class="mb" data-testid="plugin-package-risk-items">
        <h4 style="margin: 0 0 8px">风险报告明细</h4>
        <el-table :data="packageResult.risk_report.items" size="small" border>
          <el-table-column prop="level" label="等级" width="110">
            <template #default="{ row }">
              <el-tag :type="riskLevelType(row.level)" effect="plain">{{ packageRiskLabel(row.level) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="code" label="编码" min-width="240" />
          <el-table-column prop="path" label="路径" min-width="220" />
          <el-table-column prop="message" label="说明" min-width="280" />
          <el-table-column prop="suggestion" label="建议" min-width="260" />
        </el-table>
      </div>

      <el-alert
        v-if="packageResult.manifest_validation && packageResult.manifest_validation.valid === false"
        type="error"
        show-icon
        :closable="false"
        class="mb"
        title="Manifest 校验失败"
      >
        <ul class="result-list">
          <li v-for="(item, idx) in packageResult.manifest_validation.errors || []" :key="`pkg-manifest-err-${idx}`">{{ item }}</li>
        </ul>
      </el-alert>

      <el-descriptions :column="2" border class="mb" data-testid="plugin-package-dry-run-summary">
        <el-descriptions-item label="内容类型">{{ packageResult.install_dry_run?.impact_summary?.content_types_count ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="权限">{{ packageResult.install_dry_run?.impact_summary?.permissions_count ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="前端挂载">{{ packageResult.install_dry_run?.impact_summary?.menus_count ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="路由">{{ packageResult.install_dry_run?.impact_summary?.routes_count ?? 0 }}</el-descriptions-item>
      </el-descriptions>

      <pre class="json-box compact" data-testid="plugin-package-install-preview">{{ formatJSON(packageResult.install_dry_run?.install_preview || {}) }}</pre>
    </el-card>
    </div>

    <el-dialog v-model="repoDetailVisible" width="920px" destroy-on-close data-testid="plugin-package-repo-detail-dialog">
      <section class="action-panel in-drawer" data-testid="plugin-package-repo-detail-content">
        <div class="action-panel-header">
          <div>
            <h3>插件包详情</h3>
            <p class="muted">仅展示扫描、校验、风险与预检预览；不会安装，不执行代码/SQL，不加载前端资产。</p>
          </div>
          <div class="action-panel-tools">
            <el-button @click="repoDetailVisible = false">{{ t('common.close') }}</el-button>
          </div>
        </div>

        <el-alert v-if="repoDetailError" type="error" show-icon :closable="false" class="mb" :title="repoDetailError" />
        <el-skeleton v-if="repoDetailLoading" :rows="8" animated />

        <el-card v-if="repoDetail" shadow="never">
          <template #header>
            <div class="card-head">
              <strong>{{ repoDetail.package?.code || '-' }}</strong>
              <el-tag :type="repoDetail.status === 'blocked' ? 'danger' : repoDetail.status === 'warning' ? 'warning' : 'success'" effect="plain">
                {{ genericStatusLabel(repoDetail.status) }}
              </el-tag>
            </div>
          </template>

          <el-alert
            v-if="repoDetail.risk_report && repoDetail.risk_report.level"
            :type="riskLevelType(repoDetail.risk_report.level)"
            show-icon
            :closable="false"
            class="mb"
          >
            <template #title>
              <span>风险等级：</span>
              <el-tag :type="riskLevelType(repoDetail.risk_report.level)" effect="plain">
                {{ packageRiskLabel(repoDetail.risk_report.level) }}
              </el-tag>
              <span class="muted" style="margin-left: 8px">score={{ repoDetail.risk_report.score ?? 0 }}</span>
            </template>
            <div class="muted">{{ repoDetail.risk_report.summary || '-' }}</div>
          </el-alert>

          <el-descriptions :column="2" border class="mb">
            <el-descriptions-item label="路径">{{ repoDetail.package?.path || '-' }}</el-descriptions-item>
            <el-descriptions-item label="目录">{{ repoDetail.package?.dir_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ repoDetail.package?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="版本">{{ repoDetail.package?.version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="校验文件">{{ repoDetail.package?.checksum_found ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="阻断原因">{{ pluginReasonText(repoDetail.blocked_code) }}</el-descriptions-item>
          </el-descriptions>

          <el-descriptions v-if="repoDetail.checksum" :column="3" border class="mb">
            <el-descriptions-item label="校验算法">{{ repoDetail.checksum.algorithm || '-' }}</el-descriptions-item>
            <el-descriptions-item label="校验状态">
              <el-tag :type="checksumStatusType(repoDetail.checksum.status)" effect="plain">
                {{ genericStatusLabel(repoDetail.checksum.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="匹配">{{ (repoDetail.checksum.matched || []).length }}</el-descriptions-item>
            <el-descriptions-item label="不匹配">{{ (repoDetail.checksum.mismatched || []).length }}</el-descriptions-item>
            <el-descriptions-item label="缺失">{{ (repoDetail.checksum.missing || []).length }}</el-descriptions-item>
            <el-descriptions-item label="额外文件">{{ (repoDetail.checksum.extra || []).length }}</el-descriptions-item>
          </el-descriptions>

          <el-descriptions v-if="repoDetail.signature" :column="3" border class="mb" data-testid="plugin-package-repo-detail-signature">
            <el-descriptions-item label="信任状态">
              <el-tag :type="signatureTrustType(repoDetail.signature.trust_status)" effect="plain" data-testid="plugin-package-repo-detail-signature-trust">
                {{ trustLevelLabel(repoDetail.signature.trust_status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="验签状态">
              <el-tag :type="signatureVerifyType(repoDetail.signature.verification_status)" effect="plain" data-testid="plugin-package-repo-detail-signature-verify">
                {{ genericStatusLabel(repoDetail.signature.verification_status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="publisher_id">{{ repoDetail.signature.publisher_id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="public_key_id">{{ repoDetail.signature.public_key_id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="fingerprint">{{ repoDetail.signature.fingerprint || '-' }}</el-descriptions-item>
            <el-descriptions-item label="已签名文件数">{{ repoDetail.signature.signed_files_count ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="未签名文件数">{{ (repoDetail.signature.unsigned_files || []).length }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="repoDetail.signature && (repoDetail.signature.signed_files || []).length" class="mb" data-testid="plugin-package-repo-detail-signed-files">
            <h4 style="margin: 0 0 8px">已签名文件</h4>
            <div class="tag-wrap">
              <el-tag v-for="item in repoDetail.signature.signed_files || []" :key="item" type="success" effect="plain">{{ item }}</el-tag>
            </div>
          </div>

          <div v-if="repoDetail.signature && (repoDetail.signature.unsigned_files || []).length" class="mb" data-testid="plugin-package-repo-detail-unsigned-files">
            <h4 style="margin: 0 0 8px">未签名文件</h4>
            <div class="tag-wrap">
              <el-tag v-for="item in repoDetail.signature.unsigned_files || []" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
            </div>
          </div>

          <pre class="json-box compact">{{ formatJSON(repoDetail.install_dry_run?.install_preview || {}) }}</pre>
        </el-card>
      </section>
    </el-dialog>

    <el-dialog v-model="repoInstallVisible" width="820px" destroy-on-close data-testid="plugin-package-repo-install-dialog">
      <template #header>
        <div class="card-head">
          <strong>确认安装（本地插件包）</strong>
          <span class="muted">安装前必须基于本地仓库包重新执行 dry-run；安装后默认 disabled。</span>
        </div>
      </template>
      <section v-if="repoInstallVisible" class="action-panel in-drawer" data-testid="plugin-package-repo-install-content">
        <el-alert type="info" show-icon :closable="false" class="mb" title="边界：不会执行第三方代码 / 不会执行根目录 SQL / 不会动态加载前端资产；只能从本地仓库包安装，且必须使用当前安装 dry-run 计划。" />

        <el-alert v-if="repoInstallError" type="error" show-icon :closable="false" class="mb" :title="repoInstallError" />

        <div v-if="repoInstallDetail" class="mb">
          <el-descriptions :column="2" border class="mb">
            <el-descriptions-item label="插件编码">{{ repoInstallDetail.package?.code || '-' }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ repoInstallDetail.package?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="版本">{{ repoInstallDetail.package?.version || '-' }}</el-descriptions-item>
            <el-descriptions-item label="路径">{{ repoInstallDetail.package?.path || '-' }}</el-descriptions-item>
            <el-descriptions-item label="安装 dry-run">
              <el-tag :type="repoInstallDetail.dry_run_id ? 'success' : 'warning'" effect="plain">
                {{ repoInstallDetail.dry_run_id ? '当前有效' : '需要重新 dry-run' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="计划过期时间">{{ repoInstallDetail.expires_at || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="repoInstallDetail.status === 'blocked' ? 'danger' : repoInstallDetail.status === 'warning' ? 'warning' : 'success'" effect="plain" data-testid="plugin-package-repo-install-status">
                {{ genericStatusLabel(repoInstallDetail.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="风险等级">
              <el-tag :type="riskLevelType(repoInstallDetail.risk_report?.level)" effect="plain">
                {{ packageRiskLabel(repoInstallDetail.risk_report?.level) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="校验状态">
              <el-tag :type="checksumStatusType(repoInstallDetail.checksum?.status)" effect="plain">
                {{ genericStatusLabel(repoInstallDetail.checksum?.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="签名">
              <div style="display: inline-flex; gap: 6px; flex-wrap: wrap" data-testid="plugin-package-repo-install-signature">
                <el-tag :type="signatureTrustType(repoInstallDetail.signature?.trust_status)" effect="plain">
                  {{ trustLevelLabel(repoInstallDetail.signature?.trust_status) }}
                </el-tag>
                <el-tag :type="signatureVerifyType(repoInstallDetail.signature?.verification_status)" effect="plain">
                  {{ genericStatusLabel(repoInstallDetail.signature?.verification_status) }}
                </el-tag>
                <el-tag v-if="repoInstallDetail.signature?.fingerprint" type="info" effect="plain">
                  {{ repoInstallDetail.signature.fingerprint }}
                </el-tag>
              </div>
            </el-descriptions-item>
          </el-descriptions>

          <el-alert v-if="repoInstallDetail.risk_report?.summary" :type="riskLevelType(repoInstallDetail.risk_report?.level)" show-icon :closable="false" class="mb" :title="repoInstallDetail.risk_report.summary" />

          <el-alert v-if="(repoInstallDetail.errors || []).length" type="error" show-icon :closable="false" class="mb" title="阻断/错误">
            <ul class="result-list">
              <li v-for="(item, idx) in repoInstallDetail.errors" :key="`repo-install-err-${idx}`">{{ item }}</li>
            </ul>
          </el-alert>

          <el-alert v-if="(repoInstallDetail.warnings || []).length" type="warning" show-icon :closable="false" class="mb" title="警告">
            <ul class="result-list">
              <li v-for="(item, idx) in repoInstallDetail.warnings" :key="`repo-install-warn-${idx}`">{{ item }}</li>
            </ul>
          </el-alert>

          <div class="filter-actions mb" style="justify-content: space-between; gap: 10px">
            <div v-if="String(repoInstallDetail.risk_report?.level || '').toLowerCase() !== 'low'" style="display: flex; align-items: center; gap: 10px">
              <span class="muted">确认风险等级：</span>
              <el-select v-model="repoInstallConfirmRiskLevel" placeholder="确认风险等级" style="width: 180px">
                <el-option label="中风险" value="medium" />
                <el-option label="高风险" value="high" />
              </el-select>
            </div>
            <div class="filter-actions" style="justify-content: flex-end">
              <el-button :loading="repoInstallLoading" @click="repoInstallVisible = false">取消</el-button>
              <el-button
                type="primary"
                plain
                :loading="repoInstallLoading"
                :disabled="!repoInstallDetail || String(repoInstallDetail.status || '').toLowerCase() === 'blocked'"
                data-testid="plugin-package-repo-install-approval"
                @click="submitRepoInstallApproval"
              >
                提交安装审批
              </el-button>
              <el-button type="success" :loading="repoInstallLoading" :disabled="!repoInstallDetail || !repoInstallDetail.dry_run_id || String(repoInstallDetail.status || '').toLowerCase() === 'blocked'" data-testid="plugin-package-repo-install-confirm" @click="confirmRepoInstall">确认安装</el-button>
            </div>
          </div>
        </div>

        <div v-else class="muted" style="padding: 12px 0">
          <el-skeleton :rows="6" animated />
        </div>
      </section>
    </el-dialog>

    <el-empty v-if="!items.length && !loading" description="暂无插件数据" />
    <el-skeleton v-if="loading" :rows="6" animated />

    <!-- ===== copied/kept from legacy Plugins.vue to avoid E2E regression ===== -->
    <el-drawer v-model="manifestDialogVisible" append-to-body destroy-on-close size="820px" :with-header="false" class="plugin-action-drawer">
      <section class="action-panel in-drawer" data-testid="plugin-manifest-panel">
        <div class="action-panel-header">
          <div>
            <h3>{{ manifestDialogTitle }}</h3>
            <p>{{ manifestDialogTip }}</p>
          </div>
          <div class="action-panel-tools">
            <el-button :data-testid="wizardStep > 0 ? 'plugin-result-close' : 'plugin-manifest-cancel'" @click="manifestDialogVisible = false">{{ wizardStep > 0 ? t('common.close') : t('common.cancel') }}</el-button>
            <el-button v-if="wizardStep > 0 && !isWizardResultStep" :disabled="manifestLoading" @click="wizardBack">{{ t('common.back') }}</el-button>
            <el-button v-if="canShortcutConfirm" :loading="manifestLoading" type="warning" plain @click="confirmWizardAction">
              {{ manifestDialogAction === 'install' ? t('plugin.wizard.confirmInstall') : t('plugin.wizard.confirmUpgrade') }}
            </el-button>
            <el-button :loading="manifestLoading" :disabled="!wizardCanProceed" type="primary" data-testid="plugin-manifest-submit" @click="submitManifestAction">{{ manifestDialogActionLabel }}</el-button>
          </div>
        </div>
        <el-steps :active="wizardStep" finish-status="success" align-center class="wizard-steps">
          <el-step v-for="step in wizardSteps" :key="step" :title="step" />
        </el-steps>
        <el-input
          v-if="wizardStep === 0"
          v-model="manifestText"
          type="textarea"
          :rows="24"
          data-testid="plugin-manifest-input"
          :placeholder="t('plugin.ops.manifestPlaceholder')"
        />
        <div v-else data-testid="plugin-result-panel">
          <el-alert v-if="resultErrors.length" :title="t('plugin.ops.errors')" type="error" show-icon :closable="false" class="mb">
            <ul class="result-list">
              <li v-for="(item, idx) in resultErrors" :key="`err-${idx}`">{{ item }}</li>
            </ul>
          </el-alert>
          <el-alert v-if="resultWarnings.length" :title="t('plugin.ops.warnings')" type="warning" show-icon :closable="false" class="mb">
            <ul class="result-list">
              <li v-for="(item, idx) in resultWarnings" :key="`warn-${idx}`">{{ item }}</li>
            </ul>
          </el-alert>

          <template v-if="manifestDialogAction === 'upgrade' || manifestDialogAction === 'upgrade-dry-run'">
            <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
              <el-descriptions-item :label="t('plugin.ops.currentVersion')">{{ resultUpgrade.current_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.newVersion')">{{ resultUpgrade.new_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.currentCoreVersion')">{{ resultUpgrade.current_core_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.compatibilityStatus')">{{ compatibilityLabel(resultUpgrade.compatibility_status) }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.minCoreVersion')">{{ resultCompatibility(resultUpgrade).min_core_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.compatibleCoreVersion')">{{ resultCompatibility(resultUpgrade).compatible_core_version || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.changedKeys')" :span="2">{{ (resultUpgrade.changed_keys || []).join(', ') || '-' }}</el-descriptions-item>
            </el-descriptions>
            <div class="result-grid mb">
              <div class="result-box full-width" data-testid="plugin-upgrade-dependency-matrix">
                <h4>{{ t('plugin.dependencies.matrix') }}</h4>
                <div class="tag-wrap mb">
                  <el-tag :type="dependencySummaryType(resultUpgrade.validation?.dependency_summary)" effect="plain">
                    {{ dependencySummaryText(resultUpgrade.validation?.dependency_summary) }}
                  </el-tag>
                  <el-tag :type="compatibilityStatusType(resultCompatibility(resultUpgrade).status)" effect="plain">
                    {{ compatibilityLabel(resultCompatibility(resultUpgrade).status) }}
                  </el-tag>
                </div>
                <el-table :data="dependencyRows(resultUpgrade)" border stripe :empty-text="t('common.none')">
                  <el-table-column prop="code" :label="t('plugin.code')" min-width="140" />
                  <el-table-column prop="version" :label="t('plugin.dependencies.requiredVersion')" min-width="140" />
                  <el-table-column :label="t('plugin.dependencies.required')" width="100">
                    <template #default="{ row }">{{ row.required ? t('plugin.dependencies.requiredDep') : t('plugin.dependencies.optionalDep') }}</template>
                  </el-table-column>
                  <el-table-column :label="t('plugin.status')" width="150">
                    <template #default="{ row }">
                      <el-tag :type="dependencyStatusType(row.status, row.satisfied)" effect="plain">{{ dependencyStatusLabel(row.status) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="current_version" :label="t('plugin.dependencies.currentVersion')" width="130" />
                  <el-table-column prop="message" :label="t('plugin.dependencies.message')" min-width="220" />
                </el-table>
              </div>
              <div class="result-box full-width" data-testid="plugin-upgrade-dependency-diff">
                <h4>{{ t('plugin.dependencies.diff') }}</h4>
                <pre class="json-box compact">{{ formatJSON(resultUpgrade.dependency_diff || {}) }}</pre>
              </div>
            </div>
            <el-alert v-if="isWizardConfirmStep" :title="t('plugin.wizard.confirmUpgradeTip')" type="warning" show-icon :closable="false" class="mb" />
            <div class="result-grid">
              <div class="result-box">
                <h4>{{ t('plugin.ops.diffCurrent') }}</h4>
                <pre class="json-box compact">{{ formatJSON(resultUpgrade.diff?.current || {}) }}</pre>
              </div>
              <div class="result-box">
                <h4>{{ t('plugin.ops.diffNew') }}</h4>
                <pre class="json-box compact">{{ formatJSON(resultUpgrade.diff?.new || wizardResult || {}) }}</pre>
              </div>
            </div>
          </template>

          <template v-else>
            <el-descriptions :column="2" border class="mb" data-testid="plugin-result-summary">
              <el-descriptions-item :label="t('plugin.ops.resultValid')">{{ manifestResult.valid ? t('common.yes') : t('common.no') }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.sourceType')">{{ manifestResult.install_preview?.source_type || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.checksum')"><span class="mono">{{ manifestResult.checksum || '-' }}</span></el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.initialStatus')">{{ pluginStatusLabel(manifestResult.install_preview?.initial_status) }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.dependencies')">{{ dependencySummaryText(manifestResult.dependency_summary) }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.compatibilityStatus')">
                <el-tag :type="compatibilityStatusType(manifestResult.compatibility?.status)" effect="plain">{{ compatibilityLabel(manifestResult.compatibility?.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="t('plugin.ops.impactSummary')">
                {{ t('plugin.ops.contentTypesCount') }} {{ manifestResult.impact_summary?.content_types_count ?? 0 }}，
                {{ t('plugin.ops.permissionsCount') }} {{ manifestResult.impact_summary?.permissions_count ?? 0 }}，
                {{ t('plugin.ops.menusCount') }} {{ manifestResult.impact_summary?.menus_count ?? 0 }}，
                {{ t('plugin.ops.routesCount') }} {{ manifestResult.impact_summary?.routes_count ?? 0 }}
              </el-descriptions-item>
            </el-descriptions>
            <div class="result-grid mb">
              <div class="result-box full-width" data-testid="plugin-dependency-summary">
                <h4>{{ t('plugin.dependencies.matrix') }}</h4>
                <div class="tag-wrap mb">
                  <el-tag :type="dependencySummaryType(manifestResult.dependency_summary)" effect="plain">
                    {{ dependencySummaryText(manifestResult.dependency_summary) }}
                  </el-tag>
                  <span class="muted">{{ (manifestResult.compatibility?.messages || []).join('；') || '-' }}</span>
                </div>
                <el-table :data="dependencyRows(manifestResult)" border stripe :empty-text="t('common.none')">
                  <el-table-column prop="code" :label="t('plugin.code')" min-width="140" />
                  <el-table-column prop="version" :label="t('plugin.dependencies.requiredVersion')" min-width="140" />
                  <el-table-column :label="t('plugin.dependencies.required')" width="100">
                    <template #default="{ row }">{{ row.required ? t('plugin.dependencies.requiredDep') : t('plugin.dependencies.optionalDep') }}</template>
                  </el-table-column>
                  <el-table-column :label="t('plugin.status')" width="150">
                    <template #default="{ row }">
                      <el-tag :type="dependencyStatusType(row.status, row.satisfied)" effect="plain">{{ dependencyStatusLabel(row.status) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="current_version" :label="t('plugin.dependencies.currentVersion')" width="130" />
                  <el-table-column prop="message" :label="t('plugin.dependencies.message')" min-width="220" />
                </el-table>
              </div>
            </div>
            <el-alert v-if="isWizardConfirmStep" :title="t('plugin.wizard.confirmInstallTip')" type="warning" show-icon :closable="false" class="mb" />
            <div class="result-grid">
              <div class="result-box">
                <h4>{{ t('plugin.ops.contentTypeConflicts') }}</h4>
                <div class="tag-wrap">
                  <el-tag v-for="item in manifestResult.content_type_conflicts || []" :key="item" type="danger" effect="plain">{{ item }}</el-tag>
                  <span v-if="!(manifestResult.content_type_conflicts || []).length" class="muted">-</span>
                </div>
              </div>
              <div class="result-box">
                <h4>{{ t('plugin.ops.permissionConflicts') }}</h4>
                <div class="tag-wrap">
                  <el-tag v-for="item in manifestResult.permission_conflicts || []" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
                  <span v-if="!(manifestResult.permission_conflicts || []).length" class="muted">-</span>
                </div>
              </div>
              <div class="result-box">
                <h4>{{ t('plugin.ops.migrationPlan') }}</h4>
                <div class="tag-wrap">
                  <el-tag v-for="item in manifestResult.migration_plan || []" :key="item.migration_name || item.name" effect="plain">{{ item.migration_name || item.name }} / {{ item.migration_version || item.version }}</el-tag>
                  <span v-if="!(manifestResult.migration_plan || []).length" class="muted">-</span>
                </div>
              </div>
              <div class="result-box">
                <h4>{{ t('plugin.ops.installPreview') }}</h4>
                <pre class="json-box compact">{{ formatJSON(isWizardResultStep ? wizardResult : manifestResult.install_preview || {}) }}</pre>
              </div>
            </div>
          </template>
        </div>
      </section>
    </el-drawer>
    <!-- ===== end legacy wizard ===== -->
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';
import {
  createPluginApproval,
  createPluginPackageTemplate,
  dryRunPluginManifest,
  dryRunPluginPackage,
  dryRunPluginUpgrade,
  getPluginPackageDetail,
  getPluginPackageUpload,
  installPluginManifest,
  installPluginPackage,
  listPluginPackages,
  promotePluginPackageUpload,
  previewPluginPackageTemplate,
  plugins,
  upgradePlugin,
  uploadPluginPackageZip,
  validatePluginManifest,
} from '@/api/admin';
import { t } from '@/i18n';
import { genericStatusLabel, packageRiskLabel, pluginStatusLabel, trustLevelLabel } from '@/i18n/formatters';
import { pluginMessageText, pluginReasonText } from '@/modules/plugins/statusText';

const router = useRouter();
const route = useRoute();
const packageTabs = new Set(['repository', 'template', 'upload']);

const items = ref([]);
const loading = ref(false);
const targetCode = ref('');
const packageTab = ref(packageTabs.has(String(route.query.workspace_tab || '')) ? String(route.query.workspace_tab) : 'repository');

const templateForm = ref({
  code: '',
  name: '',
  content_type: '',
  content_name: '',
  description: '',
  author: '',
  with_config: true,
  with_hooks: true,
  with_migration: true,
});
const templateLoading = ref(false);
const templateError = ref('');
const templateResult = ref(null);
const templateWarnings = computed(() => [
  ...(templateResult.value?.warnings || []),
  ...(templateResult.value?.dry_run?.warnings || []),
]);
const templateErrors = computed(() => [
  ...(templateResult.value?.errors || []),
  ...(templateResult.value?.dry_run?.errors || []),
]);

const packagePath = ref('examples/plugins/demo_notice');
const packageLoading = ref(false);
const packageResult = ref(null);
const packageError = ref('');
const zipUploadFile = ref(null);
const zipUploadLoading = ref(false);
const zipUploadError = ref('');
const zipUploadResult = ref(null);
const zipPromoteLoading = ref(false);

const repoRoot = ref('storage/plugins/packages');
const repoLoading = ref(false);
const repoError = ref('');
const repoItems = ref([]);
const repoSummary = ref(null);
const repoPage = ref(1);
const repoPageSize = ref(20);
const repoTotal = ref(0);
const repoKeyword = ref('');
const repoStatus = ref('all');
const repoRiskLevel = ref('');
const repoChecksumStatus = ref('');
const repoManifestValid = ref('');

const repoDetailVisible = ref(false);
const repoDetailLoading = ref(false);
const repoDetailError = ref('');
const repoDetail = ref(null);

const repoInstallVisible = ref(false);
const repoInstallLoading = ref(false);
const repoInstallError = ref('');
const repoInstallDetail = ref(null);
const repoInstallConfirmRiskLevel = ref('');

watch(packageTab, async (next) => {
  const query = { ...route.query };
  if (next === 'repository') delete query.workspace_tab;
  else query.workspace_tab = next;
  await router.replace({ query });
});

watch(() => route.query.workspace_tab, (value) => {
  const next = packageTabs.has(String(value || '')) ? String(value) : 'repository';
  if (packageTab.value !== next) packageTab.value = next;
});

onMounted(load);

function templatePayload() {
  return {
    code: String(templateForm.value.code || '').trim(),
    name: String(templateForm.value.name || '').trim(),
    content_type: String(templateForm.value.content_type || '').trim(),
    content_name: String(templateForm.value.content_name || '').trim(),
    description: String(templateForm.value.description || '').trim(),
    author: String(templateForm.value.author || '').trim(),
    with_config: !!templateForm.value.with_config,
    with_hooks: !!templateForm.value.with_hooks,
    with_migration: !!templateForm.value.with_migration,
  };
}

function formatAPIError(e, fallback) {
  const data = e?.response?.data;
  const code = String(data?.code || '').trim();
  const message = String(data?.message || data?.error || '').trim();
  const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
  const parts = [];
  if (code) parts.push(`[${code}]`);
  if (message) parts.push(message);
  if (suggestion) parts.push(`建议：${suggestion}`);
  return parts.join(' ') || String(e?.message || fallback);
}

function onZipUploadChange(file) {
  zipUploadFile.value = file?.raw || null;
  zipUploadError.value = '';
}

function onZipUploadRemove() {
  zipUploadFile.value = null;
}

async function submitZipUpload() {
  zipUploadError.value = '';
  if (!zipUploadFile.value) {
    zipUploadError.value = '请先选择 .zip 插件包';
    return;
  }
  const name = String(zipUploadFile.value.name || '');
  if (!name.toLowerCase().endsWith('.zip')) {
    zipUploadError.value = '[plugin_package_upload_invalid_type] 只允许上传 .zip 插件包 建议：不支持 tar/gz/rar/7z。';
    return;
  }
  zipUploadLoading.value = true;
  try {
    const form = new FormData();
    form.append('file', zipUploadFile.value);
    zipUploadResult.value = await uploadPluginPackageZip(form);
    ElMessage.success('上传扫描完成');
  } catch (e) {
    zipUploadResult.value = null;
    zipUploadError.value = formatAPIError(e, '上传失败');
  } finally {
    zipUploadLoading.value = false;
  }
}

async function useUploadedPackageForDryRun() {
  const path = String(zipUploadResult.value?.package_path || '').trim();
  if (!path) return;
  packagePath.value = path;
  packageTab.value = 'repository';
  await runPackageDryRun();
}

async function promoteUploadedPackage() {
  const uploadId = String(zipUploadResult.value?.upload_id || '').trim();
  if (!uploadId) return;
  zipPromoteLoading.value = true;
  zipUploadError.value = '';
  try {
    const res = await promotePluginPackageUpload(uploadId);
    ElMessage.success(res?.message || '已转入本地插件仓库');
    if (res?.package_path) {
      packagePath.value = res.package_path;
      repoRoot.value = 'storage/plugins/packages';
    }
    packageTab.value = 'repository';
    zipUploadResult.value = await getPluginPackageUpload(uploadId).catch(() => zipUploadResult.value);
    await scanRepository(true);
  } catch (e) {
    zipUploadError.value = formatAPIError(e, '转入本地仓库失败');
  } finally {
    zipPromoteLoading.value = false;
  }
}

async function previewPackageTemplate() {
  templateError.value = '';
  templateLoading.value = true;
  try {
    templateResult.value = await previewPluginPackageTemplate(templatePayload());
  } catch (e) {
    templateResult.value = null;
    templateError.value = formatAPIError(e, '预览失败');
  } finally {
    templateLoading.value = false;
  }
}

async function createPackageTemplate() {
  templateError.value = '';
  templateLoading.value = true;
  try {
    templateResult.value = await createPluginPackageTemplate(templatePayload());
    ElMessage.success(templateResult.value?.message || '插件包模板已初始化');
    const path = String(templateResult.value?.template?.package_path || '').trim();
    if (path) packagePath.value = path;
    packageTab.value = 'repository';
    await scanRepository(true);
  } catch (e) {
    templateResult.value = null;
    templateError.value = formatAPIError(e, '初始化失败');
  } finally {
    templateLoading.value = false;
  }
}

function useTemplatePathForDryRun() {
  const path = String(templateResult.value?.template?.package_path || '').trim();
  if (!path) return;
  packagePath.value = path;
  packageTab.value = 'repository';
  runPackageDryRun();
}

async function submitTemplateInstallApproval() {
  const path = String(templateResult.value?.template?.package_path || '').trim();
  const code = String(templateResult.value?.template?.code || '').trim();
  if (!path || !code) return;
  templateLoading.value = true;
  templateError.value = '';
  try {
    const res = await createPluginApproval({
      action: 'install',
      package_path: path,
      plugin_code: code,
      reason: `从后台初始化插件包提交安装审批：${path}`,
    });
    ElMessage.success(`已提交审批 #${res?.id || ''}`.trim());
    await router.push('/plugins/approvals');
  } catch (e) {
    templateError.value = formatAPIError(e, '提交审批失败');
  } finally {
    templateLoading.value = false;
  }
}

async function load() {
  loading.value = true;
  try {
    const list = await plugins();
    items.value = list.items || [];
  } finally {
    loading.value = false;
  }
}

async function scanRepository(resetPage = false) {
  if (resetPage) repoPage.value = 1;
  repoError.value = '';
  repoLoading.value = true;
  try {
    const resp = await listPluginPackages({
      root: repoRoot.value || '',
      status: repoStatus.value || 'all',
      keyword: repoKeyword.value || '',
      risk_level: repoRiskLevel.value || '',
      checksum_status: repoChecksumStatus.value || '',
      manifest_valid: repoManifestValid.value || '',
      page: repoPage.value,
      page_size: repoPageSize.value,
    });
    repoItems.value = resp.items || [];
    repoSummary.value = resp.summary || null;
    repoTotal.value = resp.pagination?.total ?? 0;
  } catch (e) {
    const data = e?.response?.data;
    const code = String(data?.code || '').trim();
    const message = String(data?.message || data?.error || '').trim();
    const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
    const parts = [];
    if (code) parts.push(`[${code}]`);
    if (message) parts.push(message);
    if (suggestion) parts.push(`建议：${suggestion}`);
    repoError.value = parts.join(' ') || String(e?.message || '扫描失败');
    repoItems.value = [];
    repoSummary.value = null;
    repoTotal.value = 0;
  } finally {
    repoLoading.value = false;
  }
}

async function openRepoDetail(row) {
  const path = String(row?.path || '').trim();
  if (!path) return;
  repoDetailVisible.value = true;
  repoDetailLoading.value = true;
  repoDetailError.value = '';
  repoDetail.value = null;
  try {
    repoDetail.value = await getPluginPackageDetail({ path });
  } catch (e) {
    const data = e?.response?.data;
    const code = String(data?.code || '').trim();
    const message = String(data?.message || data?.error || '').trim();
    const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
    const parts = [];
    if (code) parts.push(`[${code}]`);
    if (message) parts.push(message);
    if (suggestion) parts.push(`建议：${suggestion}`);
    repoDetailError.value = parts.join(' ') || String(e?.message || '加载详情失败');
  } finally {
    repoDetailLoading.value = false;
  }
}

function canInstallRepoPackage(row) {
  const status = String(row?.status || 'ok').toLowerCase();
  const risk = String(row?.risk_level || '').toLowerCase();
  const checksum = String(row?.checksum_status || 'ok').toLowerCase();
  if (!row?.manifest_found) return false;
  if (!row?.code) return false;
  if (status === 'invalid') return false;
  if (status !== 'ok' && status !== 'warning') return false;
  if (risk === 'blocked') return false;
  if (checksum === 'failed') return false;
  return true;
}

async function openRepoInstall(row) {
  const path = String(row?.path || '').trim();
  if (!path) return;
  repoInstallVisible.value = true;
  repoInstallLoading.value = true;
  repoInstallError.value = '';
  repoInstallDetail.value = null;
  repoInstallConfirmRiskLevel.value = '';
  try {
    const detail = await getPluginPackageDetail({ path });
    repoInstallDetail.value = detail;
    repoInstallConfirmRiskLevel.value = String(detail?.risk_report?.level || '').toLowerCase();
  } catch (e) {
    const data = e?.response?.data;
    const code = String(data?.code || '').trim();
    const message = String(data?.message || data?.error || '').trim();
    const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
    const parts = [];
    if (code) parts.push(`[${code}]`);
    if (message) parts.push(message);
    if (suggestion) parts.push(`建议：${suggestion}`);
    repoInstallError.value = parts.join(' ') || String(e?.message || '加载安装信息失败');
  } finally {
    repoInstallLoading.value = false;
  }
}

async function confirmRepoInstall() {
  const path = String(repoInstallDetail.value?.package?.path || '').trim();
  if (!path) return;
  repoInstallLoading.value = true;
  repoInstallError.value = '';
  try {
    const risk = String(repoInstallDetail.value?.risk_report?.level || '').toLowerCase();
    const payload = {
      path,
      dry_run_id: String(repoInstallDetail.value?.dry_run_id || '').trim(),
      confirm_risk_level: risk && risk !== 'low' ? String(repoInstallConfirmRiskLevel.value || '').toLowerCase() : '',
    };
    const res = await installPluginPackage(payload);
    ElMessage.success(res?.message || '安装成功（默认 disabled）');
    repoInstallVisible.value = false;
    await load();
    await scanRepository();
  } catch (e) {
    const data = e?.response?.data;
    const code = String(data?.code || '').trim();
    const message = String(data?.message || data?.error || '').trim();
    const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
    const parts = [];
    if (code) parts.push(`[${code}]`);
    if (message) parts.push(message);
    if (suggestion) parts.push(`建议：${suggestion}`);
    repoInstallError.value = parts.join(' ') || String(e?.message || '安装失败');
  } finally {
    repoInstallLoading.value = false;
  }
}

async function submitRepoInstallApproval() {
  const path = String(repoInstallDetail.value?.package?.path || '').trim();
  if (!path) return;
  repoInstallLoading.value = true;
  repoInstallError.value = '';
  try {
    const payload = {
      action: 'install',
      package_path: path,
      plugin_code: String(repoInstallDetail.value?.package?.code || '').trim(),
      reason: `从本地插件仓库提交安装审批：${path}`,
    };
    const res = await createPluginApproval(payload);
    ElMessage.success(`已提交审批 #${res?.id || ''}`.trim());
    repoInstallVisible.value = false;
    await router.push('/plugins/approvals');
  } catch (e) {
    const data = e?.response?.data;
    const code = String(data?.code || '').trim();
    const message = String(data?.message || data?.error || '').trim();
    const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
    const parts = [];
    if (code) parts.push(`[${code}]`);
    if (message) parts.push(message);
    if (suggestion) parts.push(`建议：${suggestion}`);
    repoInstallError.value = parts.join(' ') || String(e?.message || '提交审批失败');
  } finally {
    repoInstallLoading.value = false;
  }
}

async function dryRunRepoPackage(row) {
  const path = String(row?.path || '').trim();
  if (!path) return;
  packagePath.value = path;
  await runPackageDryRun();
}

function findPlugin(code) {
  return (items.value || []).find((p) => (p.code || p.plugin_code) === code) || null;
}

async function runPackageDryRun() {
  const path = String(packagePath.value || '').trim();
  if (!path) {
    packageError.value = '请先输入插件包路径';
    return;
  }
  packageError.value = '';
  packageLoading.value = true;
  try {
    packageResult.value = await dryRunPluginPackage({ path });
  } catch (e) {
    packageResult.value = null;
    const data = e?.response?.data;
    const code = String(data?.code || '').trim();
    const message = String(data?.message || data?.error || '').trim();
    const suggestion = String(data?.suggestion || data?.details?.suggestion || '').trim();
    const parts = [];
    if (code) parts.push(`[${code}]`);
    if (message) parts.push(message);
    if (suggestion) parts.push(`建议：${suggestion}`);
    packageError.value = parts.join(' ') || String(e?.message || '预检失败');
  } finally {
    packageLoading.value = false;
  }
}

function clearPackageResult() {
  packageResult.value = null;
  packageError.value = '';
}

function yesNo(value) {
  return value ? '是' : '否';
}

function riskLevelType(level) {
  const v = String(level || '').toLowerCase();
  if (v === 'blocked') return 'danger';
  if (v === 'high') return 'warning';
  if (v === 'medium') return 'warning';
  return 'success';
}

function checksumStatusType(status) {
  const v = String(status || '').toLowerCase();
  if (v === 'failed') return 'danger';
  if (v === 'warning') return 'warning';
  if (v === 'missing') return 'warning';
  return 'success';
}

function signatureTrustType(status) {
  const v = String(status || '').toLowerCase();
  if (v === 'trusted') return 'success';
  if (v === 'unknown') return 'warning';
  if (v === 'unsigned') return 'warning';
  if (v === 'revoked') return 'danger';
  if (v === 'blocked') return 'danger';
  return 'info';
}

function signatureVerifyType(status) {
  const v = String(status || '').toLowerCase();
  if (v === 'verified') return 'success';
  if (v === 'publisher_unknown') return 'warning';
  if (v === 'missing') return 'warning';
  if (v === 'unsupported') return 'danger';
  if (v === 'failed') return 'danger';
  return 'info';
}

// === legacy wizard state/methods (copied with minimal adjustments) ===
const manifestDialogVisible = ref(false);
const manifestDialogAction = ref('validate');
const manifestDialogTarget = ref(null);
const manifestText = ref('');
const manifestLoading = ref(false);
const wizardStep = ref(0);
const wizardValidation = ref({});
const wizardPreview = ref({});
const wizardResult = ref({});
const wizardDisplayPayload = computed(() => {
  if (isWizardResultStep.value) return wizardResult.value || {};
  if (isWizardPreviewStep.value) return wizardPreview.value || {};
  return wizardValidation.value || {};
});
const manifestResult = computed(() => wizardDisplayPayload.value?.validation || wizardDisplayPayload.value || {});
const resultErrors = computed(() => Array.isArray(manifestResult.value?.errors) ? manifestResult.value.errors : []);
const resultWarnings = computed(() => Array.isArray(manifestResult.value?.warnings) ? manifestResult.value.warnings : []);

const manifestDialogTitle = computed(() => {
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradePreview');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgrade');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.install');
  return t('plugin.ops.validateManifest');
});

const manifestDialogTip = computed(() => {
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRunTip');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradeTip');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgradeConfirmTip');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.installTip');
  return t('plugin.ops.validateTip');
});

const wizardSteps = computed(() => {
  if (manifestDialogAction.value === 'validate') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult')];
  if (manifestDialogAction.value === 'dry-run') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult'), t('plugin.wizard.dryRunPreview')];
  if (manifestDialogAction.value === 'install') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult'), t('plugin.wizard.dryRunPreview'), t('plugin.wizard.confirmInstall'), t('plugin.wizard.installResult')];
  if (manifestDialogAction.value === 'upgrade-dry-run') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.compatibilityMatrix')];
  if (manifestDialogAction.value === 'upgrade') return [t('plugin.wizard.manifestInput'), t('plugin.wizard.compatibilityMatrix'), t('plugin.wizard.confirmUpgrade'), t('plugin.wizard.upgradeResult')];
  return [t('plugin.wizard.manifestInput'), t('plugin.wizard.validationResult')];
});

const isWizardConfirmStep = computed(() =>
  (manifestDialogAction.value === 'install' && wizardStep.value === 3)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 2),
);
const isWizardResultStep = computed(() =>
  (manifestDialogAction.value === 'install' && wizardStep.value === 4)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 3),
);
const isWizardPreviewStep = computed(() =>
  (manifestDialogAction.value === 'dry-run' && wizardStep.value === 2)
  || (manifestDialogAction.value === 'install' && wizardStep.value === 2)
  || (manifestDialogAction.value === 'upgrade-dry-run' && wizardStep.value === 1)
  || (manifestDialogAction.value === 'upgrade' && wizardStep.value === 1),
);

const canShortcutConfirm = computed(() => false);

const manifestDialogActionLabel = computed(() => {
  if (manifestDialogAction.value === 'validate' && wizardStep.value === 1) return t('common.close');
  if (manifestDialogAction.value === 'dry-run' && wizardStep.value === 2) return t('common.close');
  if (manifestDialogAction.value === 'upgrade-dry-run' && wizardStep.value === 1) return t('common.close');
  if (isWizardConfirmStep.value) return manifestDialogAction.value === 'install' ? t('plugin.wizard.confirmInstall') : t('plugin.wizard.confirmUpgrade');
  if (manifestDialogAction.value === 'install' && wizardStep.value === 2) return t('common.next');
  if (manifestDialogAction.value === 'upgrade' && wizardStep.value === 1) return t('common.next');
  if ((manifestDialogAction.value === 'dry-run' || manifestDialogAction.value === 'install') && wizardStep.value === 1) return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'dry-run') return t('plugin.ops.dryRun');
  if (manifestDialogAction.value === 'upgrade-dry-run') return t('plugin.ops.upgradePreview');
  if (manifestDialogAction.value === 'upgrade') return t('plugin.ops.upgrade');
  if (manifestDialogAction.value === 'install') return t('plugin.ops.install');
  return t('plugin.ops.validateManifest');
});

const wizardCanProceed = computed(() => {
  if (manifestLoading.value) return false;
  if ((manifestDialogAction.value === 'dry-run' || manifestDialogAction.value === 'install') && wizardStep.value === 1 && resultErrors.value.length) return false;
  if (manifestDialogAction.value === 'install' && wizardStep.value === 2 && resultErrors.value.length) return false;
  return true;
});

function openManifestDialog(action, row = null) {
  manifestDialogAction.value = action;
  manifestDialogTarget.value = row;
  wizardStep.value = 0;
  wizardValidation.value = {};
  wizardPreview.value = {};
  wizardResult.value = {};
  manifestDialogVisible.value = true;
}

function wizardBack() {
  wizardStep.value = Math.max(0, wizardStep.value - 1);
}

async function submitManifestAction() {
  manifestLoading.value = true;
  try {
    let manifest;
    try {
      manifest = JSON.parse(manifestText.value || '{}');
    } catch {
      ElMessage.error(t('plugin.ops.manifestInvalidJson'));
      return;
    }

    if (manifestDialogAction.value === 'validate') {
      wizardValidation.value = await validatePluginManifest({ manifest });
      wizardStep.value = 1;
      return;
    }
    if (manifestDialogAction.value === 'dry-run') {
      if (wizardStep.value === 0) {
        wizardValidation.value = await validatePluginManifest({ manifest });
        wizardStep.value = 1;
        return;
      }
      if (wizardStep.value === 1) {
        wizardPreview.value = await dryRunPluginManifest({ manifest });
        wizardStep.value = 2;
        return;
      }
      manifestDialogVisible.value = false;
      return;
    }
    if (manifestDialogAction.value === 'install') {
      if (wizardStep.value === 0) {
        wizardValidation.value = await validatePluginManifest({ manifest });
        wizardStep.value = 1;
        return;
      }
      if (wizardStep.value === 1) {
        wizardPreview.value = await dryRunPluginManifest({ manifest });
        wizardStep.value = 2;
        return;
      }
      if (wizardStep.value === 2) {
        wizardStep.value = 3;
        return;
      }
      if (wizardStep.value === 3) {
        wizardResult.value = await installPluginManifest({ manifest });
        wizardStep.value = 4;
        await load();
        return;
      }
      manifestDialogVisible.value = false;
      return;
    }
    if (manifestDialogAction.value === 'upgrade-dry-run') {
      const code = manifestDialogTarget.value?.code || targetCode.value;
      wizardPreview.value = await dryRunPluginUpgrade(code, { manifest });
      wizardStep.value = 1;
      return;
    }
    if (manifestDialogAction.value === 'upgrade') {
      const code = manifestDialogTarget.value?.code || targetCode.value;
      if (wizardStep.value === 0) {
        wizardPreview.value = await dryRunPluginUpgrade(code, { manifest });
        wizardStep.value = 1;
        return;
      }
      if (wizardStep.value === 1) {
        wizardStep.value = 2;
        return;
      }
      if (wizardStep.value === 2) {
        await ElMessageBox.confirm(t('plugin.ops.upgradeConfirmTip'), t('plugin.ops.upgrade'), { type: 'warning' });
        wizardResult.value = await upgradePlugin(code, { manifest });
        wizardStep.value = 3;
        await load();
        return;
      }
      manifestDialogVisible.value = false;
    }
  } finally {
    manifestLoading.value = false;
  }
}

function confirmWizardAction() {}

// helpers copied from Plugins.vue
function formatJSON(obj) {
  try {
    return JSON.stringify(obj || {}, null, 2);
  } catch {
    return String(obj || '');
  }
}

function compatibilityLabel(status) {
  if (status === 'compatible') return '兼容';
  if (status === 'warning') return '警告';
  if (status === 'incompatible') return '不兼容';
  return status || '-';
}

function resultCompatibility(result) {
  return result?.compatibility || {};
}

function compatibilityStatusType(status) {
  if (status === 'compatible') return 'success';
  if (status === 'warning') return 'warning';
  if (status === 'incompatible') return 'danger';
  return 'info';
}

function dependencyRows(result) {
  return result?.validation?.dependencies || result?.dependencies || [];
}

function dependencySummaryType(summary) {
  if (summary === 'blocked') return 'danger';
  if (summary === 'warning') return 'warning';
  if (summary === 'pass') return 'success';
  return 'info';
}

function dependencySummaryText(summary) {
  if (summary === 'blocked') return '已阻断';
  if (summary === 'warning') return '警告';
  if (summary === 'pass') return '通过';
  return summary || '-';
}

function dependencyStatusType(status, satisfied) {
  if (satisfied) return 'success';
  if (status === 'optional_missing') return 'warning';
  if (status === 'version_mismatch') return 'danger';
  if (status === 'missing') return 'danger';
  return 'info';
}

function dependencyStatusLabel(status) {
  return genericStatusLabel(status);
}
</script>

<style scoped>
.template-form {
  margin-bottom: 12px;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(260px, 1fr));
  column-gap: 18px;
}

.template-grid .wide {
  grid-column: span 2;
}

.template-switches {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.upload-grid {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) minmax(280px, 1fr);
  gap: 16px;
}

@media (max-width: 1100px) {
  .template-grid,
  .upload-grid {
    grid-template-columns: 1fr;
  }

  .template-grid .wide {
    grid-column: span 1;
  }
}
</style>
