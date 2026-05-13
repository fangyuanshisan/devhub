import axios from 'axios';
import { ElMessage } from 'element-plus';

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 12000,
});

function formatAPIError(data) {
  if (!data || typeof data !== 'object') return '';
  const code = String(data.code || '').trim();
  const message = String(data.message || data.error || '').trim();
  const suggestion = String(data.suggestion || data.details?.suggestion || '').trim();
  const parts = [];
  if (code) parts.push(`[${code}]`);
  if (message) parts.push(message);
  if (suggestion) parts.push(`建议：${suggestion}`);
  return parts.join(' ');
}

http.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('devhub_admin_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

http.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const data = error.response?.data;
    const formatted = formatAPIError(data);
    const message = formatted || data?.error || data?.message || error.message || '请求失败';
    ElMessage.error(message);
    if (error.response?.status === 401) {
      sessionStorage.removeItem('devhub_admin_token');
      sessionStorage.removeItem('devhub_admin_refresh_token');
      sessionStorage.removeItem('devhub_admin_user');
      if (!location.pathname.includes('/login')) location.href = '/admin-next/login';
    }
    return Promise.reject(error);
  },
);
