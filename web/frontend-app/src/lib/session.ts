import { ofetch } from 'ofetch';

export type DevHubUser = {
  id?: number;
  username?: string;
  nickname?: string;
  email?: string;
  phone?: string;
  role?: string;
  role_code?: string;
  role_name?: string;
  sites?: string[];
  permissions?: string[];
};

type SessionPayload = {
  token?: string;
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
  user?: DevHubUser;
};

const accessTokenKey = 'devhub_access_token';
const refreshTokenKey = 'devhub_refresh_token';
const userKey = 'devhub_user';

function inBrowser() {
  return typeof window !== 'undefined' && typeof localStorage !== 'undefined';
}

export function getAccessToken() {
  if (!inBrowser()) return '';
  return localStorage.getItem(accessTokenKey) || '';
}

export function getRefreshToken() {
  if (!inBrowser()) return '';
  return localStorage.getItem(refreshTokenKey) || '';
}

export function getStoredUser(): DevHubUser | null {
  if (!inBrowser()) return null;
  try {
    return JSON.parse(localStorage.getItem(userKey) || 'null');
  } catch {
    return null;
  }
}

export function hasSession() {
  return Boolean(getAccessToken() || getRefreshToken());
}

export function authHeaders(extra: Record<string, string> = {}) {
  const token = getAccessToken();
  return token ? { ...extra, Authorization: `Bearer ${token}` } : extra;
}

export function saveSession(session: SessionPayload) {
  if (!inBrowser()) return null;
  const accessToken = session.access_token || session.token || '';
  const refreshToken = session.refresh_token || '';
  if (accessToken) localStorage.setItem(accessTokenKey, accessToken);
  if (refreshToken) localStorage.setItem(refreshTokenKey, refreshToken);
  if (session.user) localStorage.setItem(userKey, JSON.stringify(session.user));
  notifySessionChange(session.user || getStoredUser());
  return session.user || null;
}

export function saveUser(user: DevHubUser | null) {
  if (!inBrowser()) return;
  if (user) {
    localStorage.setItem(userKey, JSON.stringify(user));
  } else {
    localStorage.removeItem(userKey);
  }
  notifySessionChange(user);
}

export function clearSession() {
  if (!inBrowser()) return;
  localStorage.removeItem(accessTokenKey);
  localStorage.removeItem(refreshTokenKey);
  localStorage.removeItem(userKey);
  notifySessionChange(null);
}

export function notifySessionChange(user: DevHubUser | null = getStoredUser()) {
  if (!inBrowser()) return;
  window.dispatchEvent(new CustomEvent('devhub:session-change', {
    detail: { user, token: getAccessToken() },
  }));
}

export async function login(account: string, password: string) {
  const session = await ofetch<SessionPayload>('/api/v1/auth/login', {
    method: 'POST',
    body: { account, password },
  });
  saveSession(session);
  return session;
}

export async function register(payload: {
  username: string;
  nickname?: string;
  email?: string;
  phone?: string;
  password: string;
}) {
  const session = await ofetch<SessionPayload>('/api/v1/auth/register', {
    method: 'POST',
    body: payload,
  });
  saveSession(session);
  return session;
}

export async function refreshSession() {
  const refreshToken = getRefreshToken();
  if (!refreshToken) throw new Error('缺少 refresh token');
  const session = await ofetch<SessionPayload>('/api/v1/auth/refresh', {
    method: 'POST',
    body: { refresh_token: refreshToken },
  });
  saveSession(session);
  return session;
}

export async function fetchCurrentUser() {
  const token = getAccessToken();
  if (!token) return null;
  const user = await ofetch<DevHubUser>('/api/v1/auth/me', {
    headers: authHeaders(),
  });
  saveUser(user);
  return user;
}

export async function restoreSession() {
  if (!hasSession()) {
    clearSession();
    return null;
  }
  try {
    return await fetchCurrentUser();
  } catch {
    try {
      await refreshSession();
      return await fetchCurrentUser();
    } catch {
      clearSession();
      return null;
    }
  }
}

export async function logout() {
  const refreshToken = getRefreshToken();
  try {
    await ofetch('/api/v1/auth/logout', {
      method: 'POST',
      body: { refresh_token: refreshToken },
    });
  } finally {
    clearSession();
  }
}

export async function authRequest<T = unknown>(path: string, options: any = {}): Promise<T> {
  if (!getAccessToken() && getRefreshToken()) {
    await refreshSession().catch(() => null);
  }
  const requestOptions = {
    ...options,
    headers: authHeaders(options.headers || {}),
  };
  try {
    return await ofetch<T>(path, requestOptions);
  } catch (error: any) {
    if ((error?.statusCode === 401 || error?.response?.status === 401) && getRefreshToken()) {
      await refreshSession();
      return ofetch<T>(path, {
        ...options,
        headers: authHeaders(options.headers || {}),
      });
    }
    throw error;
  }
}
