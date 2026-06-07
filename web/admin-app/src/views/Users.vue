<template>
  <el-card shadow="never">
    <template #header>用户管理</template>
    <el-input v-model="keyword" placeholder="搜索用户名 / 手机 / 邮箱" clearable class="table-search" />
    <el-table :data="filtered" stripe border><el-table-column prop="id" label="ID" width="80" sortable /><el-table-column prop="nickname" label="昵称" /><el-table-column prop="username" label="用户名" /><el-table-column prop="email" label="邮箱" /><el-table-column prop="role_name" label="角色" /><el-table-column prop="status" label="状态"><template #default="{ row }"><el-tag :type="row.status === 'normal' ? 'success' : 'danger'">{{ row.status }}</el-tag></template></el-table-column><el-table-column label="操作" width="150"><template #default="{ row }"><el-button link @click="set(row, 'normal')">解禁</el-button><el-button link type="danger" @click="set(row, 'forbidden')">禁用</el-button></template></el-table-column></el-table>
  </el-card>
</template>

<script setup>
import { computed, ref } from 'vue';
import { updateUserStatus, users } from '@/api/admin';
const rows = ref([]);
const keyword = ref('');
const filtered = computed(() => rows.value.filter((u) => `${u.username} ${u.nickname} ${u.email} ${u.phone}`.toLowerCase().includes(keyword.value.toLowerCase())));
async function load() { rows.value = (await users()).items || []; }
async function set(row, status) { await updateUserStatus(row.id, { status, note: status === 'forbidden' ? '后台禁用' : '' }); await load(); }
load();
</script>
