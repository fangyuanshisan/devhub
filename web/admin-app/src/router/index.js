import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

export const menuRoutes = [
  { path: '/dashboard', name: 'dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '控制台', short: '概览', icon: 'DataBoard', permission: 'dashboard.read', keepAlive: true } },
  { path: '/content', name: 'content', component: () => import('@/views/Content.vue'), meta: { title: '内容管理', short: '内容', icon: 'Document', permission: 'post.read', keepAlive: true } },
  { path: '/comments', name: 'comments', component: () => import('@/views/Comments.vue'), meta: { title: '评论审核', short: '评论', icon: 'ChatDotRound', permission: 'comment.read', keepAlive: true } },
  { path: '/reports', name: 'reports', component: () => import('@/views/Reports.vue'), meta: { title: '举报管理', short: '举报', icon: 'Warning', permission: 'report.read', keepAlive: true } },
  { path: '/moderators', name: 'moderators', component: () => import('@/views/Moderators.vue'), meta: { title: '版主管理', short: '版主', icon: 'UserFilled', permission: 'moderator.read', keepAlive: true } },
  { path: '/audit-logs', name: 'auditLogs', component: () => import('@/views/AuditLogs.vue'), meta: { title: '治理审计', short: '审计', icon: 'Tickets', permission: 'log.read', keepAlive: true } },
  { path: '/communities', name: 'communities', component: () => import('@/views/Communities.vue'), meta: { title: '子站管理', short: '子站', icon: 'SetUp', permission: 'site.read', keepAlive: true } },
  { path: '/sites', name: 'sites', component: () => import('@/views/Communities.vue'), meta: { title: '子站管理', short: '子站', icon: 'SetUp', permission: 'site.read', keepAlive: true } },
  { path: '/users', name: 'users', component: () => import('@/views/Users.vue'), meta: { title: '用户权限', short: '用户', icon: 'User', permission: 'user.read', keepAlive: true } },
  { path: '/operation', name: 'operation', component: () => import('@/views/Operation.vue'), meta: { title: '运营工具', short: '运营', icon: 'Promotion', permission: 'notification.write', keepAlive: true } },
  { path: '/statistics', name: 'statistics', component: () => import('@/views/Statistics.vue'), meta: { title: '数据统计', short: '数据', icon: 'TrendCharts', permission: 'dashboard.read', keepAlive: true } },
  { path: '/system', name: 'system', component: () => import('@/views/System.vue'), meta: { title: '系统设置', short: '系统', icon: 'Setting', permission: 'setting.read', keepAlive: true } },
];

const router = createRouter({
  history: createWebHistory('/admin-next/'),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
    ...menuRoutes,
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
  ],
});

router.beforeEach((to) => {
  const auth = useAuthStore();
  if (to.path !== '/login' && !auth.authed) return '/login';
  if (to.path !== '/login' && to.meta.permission && !auth.can(to.meta.permission)) return '/dashboard';
  if (to.path === '/login' && auth.authed) return '/dashboard';
  return true;
});

export default router;
