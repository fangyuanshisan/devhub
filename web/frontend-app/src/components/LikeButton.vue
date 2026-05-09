<template>
  <button class="like-button" :class="{ liked }" type="button" @click="like">
    <span>{{ liked ? '已赞' : '点赞' }}</span>
    <strong>{{ count }}</strong>
  </button>
  <p v-if="message" class="form-message">{{ message }}</p>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { authRequest, hasSession } from '../lib/session';

const props = defineProps<{ postId: number; initialLikes: number }>();
const count = ref(props.initialLikes);
const liked = ref(false);
const message = ref('');

async function like() {
  if (liked.value) return;
  if (!hasSession()) {
    message.value = '请先登录后再点赞';
    return;
  }
  try {
    const result = await authRequest<any>(`/api/v1/topics/${props.postId}/like`, {
      method: 'POST',
    });
    count.value = result.count ?? count.value + 1;
    liked.value = true;
    message.value = '';
  } catch (error: any) {
    message.value = error?.data?.error || '点赞失败，请稍后再试';
  }
}
</script>
