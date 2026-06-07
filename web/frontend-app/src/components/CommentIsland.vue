<template>
  <section class="comment-island">
    <div class="section-head">
      <h2>评论</h2>
      <span>{{ comments.length }} 条</span>
    </div>
    <form class="comment-form" @submit.prevent="submit">
      <textarea v-model="text" placeholder="写下你的看法" />
      <button type="submit">发布评论</button>
    </form>
    <p v-if="message" class="form-message">{{ message }}</p>
    <div class="comment-list">
      <article v-for="item in comments" :key="item.id" class="comment-item">
        <strong>{{ item.author }}</strong>
        <p>{{ item.text }}</p>
        <span>{{ item.created_at }}</span>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ofetch } from 'ofetch';
import { authRequest, getStoredUser, hasSession } from '../lib/session';

const props = defineProps<{ postId: number }>();
const comments = ref<any[]>([]);
const text = ref('');
const message = ref('');

onMounted(load);

async function load() {
  const data = await ofetch(`/api/v1/topics/${props.postId}/comments`).catch(() => ({ items: [] }));
  comments.value = data.items || [];
}

async function submit() {
  if (!hasSession()) {
    message.value = '请先登录后再评论';
    return;
  }
  try {
    const user = getStoredUser();
    await authRequest(`/api/v1/topics/${props.postId}/comments`, {
      method: 'POST',
      body: {
        author: user?.nickname || user?.username || 'DevHub 用户',
        text: text.value,
      },
    });
    text.value = '';
    message.value = '';
    await load();
  } catch (error: any) {
    message.value = error?.data?.error || '评论发布失败';
  }
}
</script>
