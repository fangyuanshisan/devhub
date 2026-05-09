import axios from 'axios';
import { ElMessage } from 'element-plus';

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 12000,
});

http.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('devhub_admin_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

http.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.error || error.response?.data?.message || error.message || '请求失败';
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
