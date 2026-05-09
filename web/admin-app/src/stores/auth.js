import { defineStore } from 'pinia';
import { login as loginAPI, me as meAPI, logout as logoutAPI } from '@/api/admin';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: sessionStorage.getItem('devhub_admin_token') || '',
    refreshToken: sessionStorage.getItem('devhub_admin_refresh_token') || '',
    user: JSON.parse(sessionStorage.getItem('devhub_admin_user') || 'null'),
    adminContext: null,
  }),
  getters: {
    authed: (state) => Boolean(state.token),
    permissions: (state) => state.user?.permissions || ['*'],
    role: (state) => state.user?.role_code || state.user?.role || '',
  },
  actions: {
    async login(payload) {
      const session = await loginAPI(payload);
      this.token = session.access_token || session.token;
      this.refreshToken = session.refresh_token || '';
      this.user = session.user;
      sessionStorage.setItem('devhub_admin_token', this.token);
      sessionStorage.setItem('devhub_admin_refresh_token', this.refreshToken);
      sessionStorage.setItem('devhub_admin_user', JSON.stringify(this.user));
      return session;
    },
    async fetchMe() {
      const data = await meAPI();
      this.user = data.user || this.user;
      this.adminContext = data.admin_context || null;
      sessionStorage.setItem('devhub_admin_user', JSON.stringify(this.user));
    },
    async logout() {
      if (this.refreshToken) await logoutAPI(this.refreshToken).catch(() => null);
      this.token = '';
      this.refreshToken = '';
      this.user = null;
      this.adminContext = null;
      sessionStorage.removeItem('devhub_admin_token');
      sessionStorage.removeItem('devhub_admin_refresh_token');
      sessionStorage.removeItem('devhub_admin_user');
    },
    can(permission) {
      const permissions = this.permissions;
      return permissions.includes('*') || permissions.includes(permission);
    },
  },
});
