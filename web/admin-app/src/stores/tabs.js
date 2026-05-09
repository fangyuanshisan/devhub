import { defineStore } from 'pinia';

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    tabs: [{ path: '/dashboard', title: '控制台', name: 'dashboard' }],
  }),
  actions: {
    add(route) {
      if (!route.meta?.title || route.path === '/login') return;
      if (!this.tabs.some((tab) => tab.path === route.path)) {
        this.tabs.push({ path: route.path, title: route.meta.title, name: route.name });
      }
    },
    remove(path) {
      if (path === '/dashboard') return;
      this.tabs = this.tabs.filter((tab) => tab.path !== path);
    },
  },
});
