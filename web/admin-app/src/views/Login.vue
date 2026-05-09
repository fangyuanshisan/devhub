<template>
  <div class="login-page">
    <el-card class="login-card">
      <div class="login-brand"><span>DH</span><div><h1>DevHub Admin</h1><p>多站点 CMS 管理后台</p></div></div>
      <el-form ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="submit">
        <el-form-item prop="account"><el-input v-model="form.account" placeholder="账号" /></el-form-item>
        <el-form-item prop="password"><el-input v-model="form.password" placeholder="密码" type="password" show-password /></el-form-item>
        <el-form-item><el-button type="primary" native-type="submit" :loading="loading" class="full">登录</el-button></el-form-item>
      </el-form>
      <el-alert title="演示账号：admin / admin123" type="info" :closable="false" />
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const auth = useAuthStore();
const formRef = ref();
const loading = ref(false);
const form = reactive({ account: 'admin', password: 'admin123' });
const rules = {
  account: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
};

async function submit() {
  await formRef.value.validate();
  loading.value = true;
  try {
    await auth.login(form);
    router.push('/dashboard');
  } finally {
    loading.value = false;
  }
}
</script>
