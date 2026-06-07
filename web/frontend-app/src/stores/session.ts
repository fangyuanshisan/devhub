import { defineStore } from 'pinia';
import {
  clearSession,
  fetchCurrentUser,
  getAccessToken,
  getStoredUser,
  login,
  logout,
  refreshSession,
  register,
  restoreSession,
  saveSession,
  type DevHubUser,
} from '../lib/session';

export const useSessionStore = defineStore('session', {
  state: () => ({
    token: '',
    user: null as DevHubUser | null,
    ready: false,
  }),
  actions: {
    hydrate() {
      this.token = getAccessToken();
      this.user = getStoredUser();
      this.ready = true;
    },
    async restore() {
      const user = await restoreSession();
      this.token = getAccessToken();
      this.user = user;
      this.ready = true;
      return user;
    },
    async login(account: string, password: string) {
      const session = await login(account, password);
      this.token = getAccessToken();
      this.user = session.user || null;
      this.ready = true;
      return session;
    },
    async register(payload: { username: string; nickname?: string; email?: string; phone?: string; password: string }) {
      const session = await register(payload);
      saveSession(session);
      this.token = getAccessToken();
      this.user = session.user || null;
      this.ready = true;
      return session;
    },
    async refresh() {
      const session = await refreshSession();
      this.token = getAccessToken();
      this.user = session.user || this.user;
      return session;
    },
    async fetchMe() {
      const user = await fetchCurrentUser();
      this.token = getAccessToken();
      this.user = user;
      return user;
    },
    async logout() {
      await logout();
      this.token = '';
      this.user = null;
      this.ready = true;
    },
    clear() {
      clearSession();
      this.token = '';
      this.user = null;
      this.ready = true;
    },
  },
});
