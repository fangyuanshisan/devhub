<template>
  <router-view v-if="$route.path === '/login'" />
  <AdminLayout v-else />
</template>

<script setup>
import AdminLayout from '@/layout/AdminLayout.vue';
</script>
