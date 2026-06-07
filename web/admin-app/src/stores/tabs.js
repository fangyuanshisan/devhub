import { defineStore } from 'pinia';

const HOME_PATH = '/dashboard';
const HOME_TAB = { path: HOME_PATH, title: '控制台', name: 'dashboard' };

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    tabs: [{ ...HOME_TAB }],
  }),
  actions: {
    add(route) {
      if (!route.meta?.title || route.path === '/login') return;
      if (!this.tabs.some((tab) => tab.path === route.path)) {
        this.tabs.push({ path: route.path, title: route.meta.title, name: route.name });
      }
    },
    remove(path) {
      if (path === HOME_PATH) return;
      this.tabs = this.tabs.filter((tab) => tab.path !== path);
    },
    closeOthers(path) {
      this.tabs = this.tabs.filter((tab) => tab.path === HOME_PATH || tab.path === path);
      if (!this.tabs.some((tab) => tab.path === HOME_PATH)) this.tabs.unshift({ ...HOME_TAB });
    },
    closeLeft(path) {
      const currentIndex = this.tabs.findIndex((tab) => tab.path === path);
      if (currentIndex <= 0) return;
      this.tabs = this.tabs.filter((tab, index) => tab.path === HOME_PATH || index >= currentIndex);
    },
    closeRight(path) {
      const currentIndex = this.tabs.findIndex((tab) => tab.path === path);
      if (currentIndex < 0) return;
      this.tabs = this.tabs.filter((tab, index) => tab.path === HOME_PATH || index <= currentIndex);
    },
    closeAll() {
      this.tabs = [{ ...HOME_TAB }];
    },
  },
});
