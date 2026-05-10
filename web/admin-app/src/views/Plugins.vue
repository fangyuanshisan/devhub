<template>
  <section class="page-card">
    <div class="toolbar">
      <div>
        <h2>系统插件</h2>
        <p>Core 只保留通用能力，问答、文档、Wiki 通过系统插件启用。</p>
      </div>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table :data="items" border stripe>
      <el-table-column prop="code" label="插件标识" width="120" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column label="内容类型" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="type in row.content_types || []" :key="type" class="mr-6">{{ type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="权限" min-width="220">
        <template #default="{ row }">
          <el-tag v-for="permission in (row.permissions || []).slice(0, 3)" :key="permission.code" class="mr-6" type="info">
            {{ permission.code }}
          </el-tag>
          <span v-if="(row.permissions || []).length > 3" class="muted">+{{ row.permissions.length - 3 }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="260" />
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="row.status === 'enabled' ? 'success' : 'info'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260">
        <template #default="{ row }">
          <el-button v-if="canOpen(row)" link type="primary" @click="openPlugin(row)">进入管理</el-button>
          <el-button link type="info" @click="openManifest(row)">查看声明</el-button>
          <el-button v-if="row.status !== 'enabled'" link type="success" @click="setStatus(row, 'enabled')">启用</el-button>
          <el-button v-if="row.status === 'enabled'" link type="warning" @click="setStatus(row, 'disabled')">禁用</el-button>
        </template>
      </el-table-column>
    </el-table>
  </section>

  <el-dialog v-model="manifestDialog" :title="`${manifestTarget?.name || ''} 插件声明`" width="760px">
    <el-descriptions :column="1" border>
      <el-descriptions-item label="code">{{ manifestTarget?.code }}</el-descriptions-item>
      <el-descriptions-item label="version">{{ manifestTarget?.version }}</el-descriptions-item>
      <el-descriptions-item label="status">{{ manifestTarget?.status }}</el-descriptions-item>
      <el-descriptions-item label="content_types">{{ (manifestTarget?.content_types || []).join(', ') || '-' }}</el-descriptions-item>
      <el-descriptions-item label="permissions">{{ (manifestTarget?.permissions || []).map((item) => item.code).join(', ') || '-' }}</el-descriptions-item>
    </el-descriptions>
    <el-form label-width="110px" class="manifest-form">
      <el-form-item label="config_schema">
        <el-input :model-value="formatJSON(manifestTarget?.config_schema)" type="textarea" :rows="8" readonly />
      </el-form-item>
      <el-form-item label="routes">
        <el-input :model-value="formatJSON(manifestTarget?.routes || [])" type="textarea" :rows="8" readonly />
      </el-form-item>
      <el-form-item label="hooks">
        <el-input :model-value="formatJSON(manifestTarget?.hooks || [])" type="textarea" :rows="6" readonly />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="manifestDialog = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import { disablePlugin, enablePlugin, plugins } from '@/api/admin';
import { useAuthStore } from '@/stores/auth';

const auth = useAuthStore();
const router = useRouter();
const items = ref([]);
const manifestDialog = ref(false);
const manifestTarget = ref(null);
const routeMap = {
  qa: '/qa',
  docs: '/docs',
  wiki: '/wiki',
  projects: '/projects',
  jobs: '/jobs',
  ai_works: '/ai-works',
};

async function load() {
  const data = await plugins();
  items.value = data.items || [];
}

async function setStatus(row, status) {
  if (status === 'disabled') {
    await ElMessageBox.confirm(
      '禁用全局插件后，所有子站都不能继续新发布该插件内容，但历史内容详情仍可访问。是否继续？',
      '禁用确认',
      { type: 'warning' },
    );
  }
  if (status === 'enabled') await enablePlugin(row.code);
  else await disablePlugin(row.code);
  ElMessage.success('插件状态已更新');
  await load();
}

function canOpen(row) {
  const target = routeMap[row.code];
  return Boolean(target && row.status === 'enabled' && hasPermission(row));
}

function hasPermission(row) {
  const required = row.menus?.find((item) => item.area === 'admin')?.permission || row.permissions?.[0]?.code || '';
  return !required || auth.can(required);
}

function openPlugin(row) {
  const target = routeMap[row.code];
  if (!target) return;
  router.push(target);
}

function openManifest(row) {
  manifestTarget.value = row;
  manifestDialog.value = true;
}

function formatJSON(value) {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return '{}';
  }
}

onMounted(load);
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; }
.toolbar h2 { margin: 0 0 6px; }
.toolbar p { margin: 0; color: #64748b; }
.mr-6 { margin-right: 6px; }
.muted { color: #64748b; }
.manifest-form { margin-top: 16px; }
</style>
