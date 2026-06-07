<template>
  <section class="auth-panel" id="login">
    <div class="auth-tabs" v-if="!user">
      <button type="button" :class="{ active: mode === 'login' }" @click="setMode('login')">登录</button>
      <button type="button" :class="{ active: mode === 'register' }" @click="setMode('register')">注册</button>
    </div>

    <div v-if="user" class="auth-user">
      <span>{{ displayName }}</span>
      <button type="button" :disabled="submitting" @click="handleLogout">退出</button>
    </div>

    <form v-else-if="mode === 'login'" class="auth-form" @submit.prevent="handleLogin">
      <input v-model.trim="loginForm.account" autocomplete="username" placeholder="账号 / 邮箱 / 手机" />
      <input v-model="loginForm.password" autocomplete="current-password" placeholder="密码" type="password" />
      <button type="submit" :disabled="submitting">{{ submitting ? '登录中...' : '登录' }}</button>
      <p v-if="message" class="form-message" :class="{ success: messageType === 'success' }">{{ message }}</p>
    </form>

    <form v-else class="auth-form" id="register" @submit.prevent="handleRegister">
      <input v-model.trim="registerForm.username" autocomplete="username" placeholder="用户名" />
      <input v-model.trim="registerForm.nickname" autocomplete="nickname" placeholder="昵称，可选" />
      <input v-model.trim="registerForm.email" autocomplete="email" placeholder="邮箱，可选" type="email" />
      <input v-model="registerForm.password" autocomplete="new-password" placeholder="密码，至少 6 位" type="password" />
      <button type="submit" :disabled="submitting">{{ submitting ? '注册中...' : '注册并登录' }}</button>
      <p v-if="message" class="form-message" :class="{ success: messageType === 'success' }">{{ message }}</p>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { login, logout, register, restoreSession, type DevHubUser } from '../lib/session';

const mode = ref<'login' | 'register'>('login');
const message = ref('');
const messageType = ref<'error' | 'success'>('error');
const submitting = ref(false);
const user = ref<DevHubUser | null>(null);

const loginForm = reactive({
  account: '',
  password: '',
});

const registerForm = reactive({
  username: '',
  nickname: '',
  email: '',
  password: '',
});

const displayName = computed(() => user.value?.nickname || user.value?.username || '已登录用户');

onMounted(async () => {
  user.value = await restoreSession();
  const hash = window.location.hash.replace('#', '');
  if (hash === 'register') mode.value = 'register';
  if (hash === 'login') mode.value = 'login';
  window.addEventListener('hashchange', syncModeFromHash);
  window.addEventListener('devhub:session-change', syncUserFromEvent as EventListener);
});

function syncModeFromHash() {
  const hash = window.location.hash.replace('#', '');
  if (hash === 'register') setMode('register', false);
  if (hash === 'login') setMode('login', false);
}

function syncUserFromEvent(event: CustomEvent<{ user: DevHubUser | null }>) {
  user.value = event.detail?.user || null;
}

function setMode(value: 'login' | 'register', updateHash = true) {
  mode.value = value;
  message.value = '';
  if (updateHash) window.history.replaceState(null, '', `#${value}`);
}

function setMessage(value: string, type: 'error' | 'success' = 'error') {
  message.value = value;
  messageType.value = type;
}

async function handleLogin() {
  if (!loginForm.account || !loginForm.password) {
    setMessage('请输入账号和密码');
    return;
  }
  submitting.value = true;
  message.value = '';
  try {
    const session = await login(loginForm.account, loginForm.password);
    user.value = session.user || null;
    setMessage('登录成功', 'success');
  } catch (error: any) {
    setMessage(error?.data?.error || '登录失败');
  } finally {
    submitting.value = false;
  }
}

async function handleRegister() {
  if (!registerForm.username) {
    setMessage('请输入用户名');
    return;
  }
  if (registerForm.password.length < 6) {
    setMessage('密码至少 6 位');
    return;
  }
  submitting.value = true;
  message.value = '';
  try {
    const session = await register({
      username: registerForm.username,
      nickname: registerForm.nickname,
      email: registerForm.email,
      password: registerForm.password,
    });
    user.value = session.user || null;
    setMessage('注册成功，已自动登录', 'success');
  } catch (error: any) {
    setMessage(error?.data?.error || '注册失败');
  } finally {
    submitting.value = false;
  }
}

async function handleLogout() {
  submitting.value = true;
  try {
    await logout();
    user.value = null;
    setMessage('已退出登录', 'success');
  } finally {
    submitting.value = false;
  }
}
</script>
