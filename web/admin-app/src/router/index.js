import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

export const menuRoutes = [
  { path: '/dashboard', name: 'dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '控制台', short: '概览', icon: 'DataBoard', permission: 'dashboard.read', keepAlive: true, moduleKey: 'home', navGroupKey: 'overview', navPageKey: 'dashboard' } },
  { path: '/content', name: 'content', component: () => import('@/views/Content.vue'), meta: { title: '内容管理', short: '内容', icon: 'Document', permission: 'post.read', keepAlive: true, moduleKey: 'content', navGroupKey: 'governance', navPageKey: 'content' } },
  { path: '/comments', name: 'comments', component: () => import('@/views/Comments.vue'), meta: { title: '评论审核', short: '评论', icon: 'ChatDotRound', permission: 'comment.read', keepAlive: true, moduleKey: 'content', navGroupKey: 'governance', navPageKey: 'comments' } },
  { path: '/reports', name: 'reports', component: () => import('@/views/Reports.vue'), meta: { title: '举报管理', short: '举报', icon: 'Warning', permission: 'report.read', keepAlive: true, moduleKey: 'content', navGroupKey: 'governance', navPageKey: 'reports' } },
  { path: '/moderators', name: 'moderators', component: () => import('@/views/Moderators.vue'), meta: { title: '版主管理', short: '版主', icon: 'UserFilled', permission: 'moderator.read', keepAlive: true, moduleKey: 'communities', navGroupKey: 'moderation', navPageKey: 'moderators' } },
  { path: '/audit-logs', name: 'auditLogs', component: () => import('@/views/AuditLogs.vue'), meta: { title: '治理审计', short: '审计', icon: 'Tickets', permission: 'log.read', keepAlive: true, moduleKey: 'maintenance', navGroupKey: 'audit', navPageKey: 'auditLogs' } },
  { path: '/tags', name: 'tags', component: () => import('@/views/Tags.vue'), meta: { title: '标签管理', short: '标签', icon: 'PriceTag', permission: 'post.read', keepAlive: true, moduleKey: 'content', navGroupKey: 'governance', navPageKey: 'tags' } },
  { path: '/plugins', name: 'plugins', redirect: '/plugins/overview', meta: { title: '系统插件', short: '插件', icon: 'Connection', permission: 'plugin.read', keepAlive: true, moduleKey: 'plugins' } },
  { path: '/qa', name: 'qaPlugin', component: () => import('@/views/PluginContent.vue'), meta: { title: '问答管理', short: '问答', icon: 'QuestionFilled', permission: 'qa.question.audit', pluginCode: 'qa', contentType: 'question', hiddenInMenu: true, keepAlive: true, moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/docs', name: 'docsPlugin', component: () => import('@/views/PluginContent.vue'), meta: { title: '文档管理', short: '文档', icon: 'Notebook', permission: 'docs.document.audit', pluginCode: 'docs', contentType: 'document', hiddenInMenu: true, keepAlive: true, moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/wiki', name: 'wikiPlugin', component: () => import('@/views/PluginContent.vue'), meta: { title: 'Wiki 管理', short: 'Wiki', icon: 'Document', permission: 'wiki.page.audit', pluginCode: 'wiki', contentType: 'wiki_page', hiddenInMenu: true, keepAlive: true, moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/projects', name: 'projectsPlugin', component: () => import('@/views/PluginContent.vue'), meta: { title: '项目管理', short: '项目', icon: 'FolderOpened', permission: 'projects.project.audit', pluginCode: 'projects', contentType: 'project', hiddenInMenu: true, keepAlive: true, moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/jobs', name: 'jobsPlugin', component: () => import('@/views/PluginContent.vue'), meta: { title: '招聘管理', short: '招聘', icon: 'Briefcase', permission: 'jobs.job.audit', pluginCode: 'jobs', contentType: 'job', hiddenInMenu: true, keepAlive: true, moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/ai-works', name: 'aiWorksPlugin', component: () => import('@/views/PluginContent.vue'), meta: { title: 'AI 作品管理', short: 'AI作品', icon: 'MagicStick', permission: 'ai_works.work.audit', pluginCode: 'ai_works', contentType: 'ai_work', hiddenInMenu: true, keepAlive: true, moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/communities', name: 'communities', component: () => import('@/views/Communities.vue'), meta: { title: '子站管理', short: '子站', icon: 'SetUp', permission: 'site.read', keepAlive: true, moduleKey: 'communities', navGroupKey: 'base', navPageKey: 'communities' } },
  { path: '/users', name: 'users', component: () => import('@/views/Users.vue'), meta: { title: '用户权限', short: '用户', icon: 'User', permission: 'user.read', keepAlive: true, moduleKey: 'users', navGroupKey: 'accounts', navPageKey: 'users' } },
  { path: '/operation', name: 'operation', component: () => import('@/views/Operation.vue'), meta: { title: '运营工具', short: '运营', icon: 'Promotion', permission: 'notification.write', keepAlive: true, moduleKey: 'system', navGroupKey: 'operation', navPageKey: 'operation' } },
  { path: '/statistics', name: 'statistics', component: () => import('@/views/Statistics.vue'), meta: { title: '数据统计', short: '数据', icon: 'TrendCharts', permission: 'dashboard.read', keepAlive: true, moduleKey: 'home', navGroupKey: 'stats', navPageKey: 'statistics' } },
  { path: '/system', name: 'system', component: () => import('@/views/System.vue'), meta: { title: '系统设置', short: '系统', icon: 'Setting', permission: 'setting.read', keepAlive: true, moduleKey: 'system', navGroupKey: 'settings', navPageKey: 'system' } },
];

const withTab = (path, tab) => (to) => ({ path, query: { ...to.query, tab } });
const withPackageInstallTab = (to) => {
  const query = { ...to.query, tab: 'install' };
  return { path: '/plugins/packages', query };
};

export const pluginRoutes = [
  { path: '/plugins/overview', name: 'pluginsOverview', component: () => import('@/views/plugins/PluginOverviewDomain.vue'), meta: { title: '插件总览', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'overview', moduleKey: 'plugins', navGroupKey: 'governance', navPageKey: 'overview' } },
  { path: '/plugins/packages', name: 'pluginsPackages', component: () => import('@/views/plugins/PluginPackagesDomain.vue'), meta: { title: '插件包治理', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'packages', moduleKey: 'plugins', navGroupKey: 'governance', navPageKey: 'packages' } },
  { path: '/plugins/webhooks', name: 'pluginsWebhooks', component: () => import('@/views/plugins/PluginWebhooks.vue'), meta: { title: 'Webhook 治理', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'webhooks', moduleKey: 'plugins', navGroupKey: 'governance', navPageKey: 'webhooks' } },
  { path: '/plugins/publishers', name: 'pluginsPublishers', component: () => import('@/views/plugins/PluginPublishersDomain.vue'), meta: { title: '发布者与信任', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'publishers', moduleKey: 'plugins', navGroupKey: 'governance', navPageKey: 'publishers' } },
  { path: '/plugins/runtime', name: 'pluginsRuntime', component: () => import('@/views/plugins/PluginRuntimeDomain.vue'), meta: { title: '运行记录 / 审计', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'runtime', moduleKey: 'plugins', navGroupKey: 'governance', navPageKey: 'runtime' } },

  { path: '/plugins/list', redirect: withTab('/plugins/overview', 'list') },
  { path: '/plugins/content', redirect: withTab('/plugins/overview', 'content') },
  { path: '/plugins/config', redirect: withTab('/plugins/overview', 'config') },
  { path: '/plugins/navigation', redirect: withTab('/plugins/overview', 'navigation') },
  { path: '/plugins/permissions', redirect: withTab('/plugins/overview', 'permissions') },
  { path: '/plugins/developer', redirect: withTab('/plugins/overview', 'developer') },

  { path: '/plugins/install', redirect: withPackageInstallTab },
  { path: '/plugins/packages/local', redirect: withPackageInstallTab },
  { path: '/plugins/packages/install', redirect: withPackageInstallTab },
  { path: '/plugins/packages/export', redirect: withPackageInstallTab },
  { path: '/plugins/packages/uploads', redirect: withTab('/plugins/packages', 'uploads') },
  { path: '/plugins/package-uploads', redirect: withTab('/plugins/packages', 'uploads') },
  { path: '/plugins/packages/remote', redirect: withTab('/plugins/packages', 'remote-packages') },
  { path: '/plugins/remote-packages', redirect: withTab('/plugins/packages', 'remote-packages') },
  { path: '/plugins/versions', redirect: withTab('/plugins/packages', 'versions') },
  { path: '/plugins/upgrade-diff', redirect: withTab('/plugins/packages', 'versions') },
  { path: '/plugins/remote-indexes', redirect: withTab('/plugins/packages', 'remote-indexes') },
  { path: '/plugins/dependencies', redirect: withTab('/plugins/packages', 'dependencies') },
  { path: '/plugins/approvals', redirect: withTab('/plugins/packages', 'approvals') },

  { path: '/plugins/events', redirect: withTab('/plugins/webhooks', 'events') },
  { path: '/plugins/webhook-events', redirect: withTab('/plugins/webhooks', 'events') },
  { path: '/plugins/webhook-deliveries', redirect: withTab('/plugins/webhooks', 'deliveries') },
  { path: '/plugins/webhook-retry', redirect: withTab('/plugins/webhooks', 'exceptions') },
  { path: '/plugins/webhook-circuits', redirect: withTab('/plugins/webhooks', 'exceptions') },
  { path: '/plugins/webhook-secrets', redirect: withTab('/plugins/webhooks', 'secrets') },
  { path: '/plugins/callback-tokens', redirect: withTab('/plugins/webhooks', 'callback_tokens') },
  { path: '/plugins/callback-requests', redirect: withTab('/plugins/webhooks', 'callback_requests') },

  { path: '/plugins/trusted-publishers', redirect: withTab('/plugins/publishers', 'list') },
  { path: '/plugins/config-keys', redirect: withTab('/plugins/publishers', 'config-keys') },
  { path: '/plugins/security', redirect: withTab('/plugins/publishers', 'config-keys') },

  { path: '/plugins/operations', redirect: withTab('/plugins/runtime', 'operations') },
  { path: '/plugins/audit', redirect: withTab('/plugins/runtime', 'audit') },
  { path: '/plugins/hooks', redirect: withTab('/plugins/runtime', 'hooks') },
  { path: '/plugins/search-index', redirect: withTab('/plugins/runtime', 'search-index') },

  // legacy compat routes
  { path: '/plugins/governance', redirect: '/plugins/overview' },
  { path: '/plugins/manifest', redirect: withPackageInstallTab },
  { path: '/plugins/diagnostics', redirect: withTab('/plugins/runtime', 'hooks') },
];

const router = createRouter({
  history: createWebHistory('/admin-next/'),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
    ...menuRoutes,
    ...pluginRoutes,
    { path: '/sites', redirect: '/communities' },
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
