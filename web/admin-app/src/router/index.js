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

export const pluginRoutes = [
  { path: '/plugins/overview', name: 'pluginsOverview', component: () => import('@/views/plugins/PluginOverview.vue'), meta: { title: '插件概览', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'overview', moduleKey: 'plugins', navGroupKey: 'overview', navPageKey: 'overview' } },
  { path: '/plugins/list', name: 'pluginsList', component: () => import('@/views/plugins/PluginList.vue'), meta: { title: '插件列表', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'list', moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'list' } },
  { path: '/plugins/content', name: 'pluginsContent', component: () => import('@/views/plugins/PluginContentHub.vue'), meta: { title: '内容治理', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'content', moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'content' } },
  { path: '/plugins/install', name: 'pluginsInstall', component: () => import('@/views/plugins/PluginInstallUpgrade.vue'), meta: { title: '本地插件仓库', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'install', moduleKey: 'plugins', navGroupKey: 'packages', navPageKey: 'install' } },
  { path: '/plugins/packages', redirect: '/plugins/install' },
  { path: '/plugins/packages/local', redirect: '/plugins/install' },
  { path: '/plugins/packages/install', redirect: '/plugins/install' },
  { path: '/plugins/packages/export', redirect: '/plugins/install' },
  { path: '/plugins/packages/uploads', name: 'pluginPackageUploads', component: () => import('@/views/plugins/PluginPackageUploads.vue'), meta: { title: 'zip 上传包', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'packageUploads', moduleKey: 'plugins', navGroupKey: 'packages', navPageKey: 'packageUploads' } },
  { path: '/plugins/packages/remote', name: 'pluginRemotePackages', component: () => import('@/views/plugins/PluginRemotePackages.vue'), meta: { title: '远程插件包', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'remotePackages', moduleKey: 'plugins', navGroupKey: 'packages', navPageKey: 'remotePackages' } },
  { path: '/plugins/versions', name: 'pluginVersions', component: () => import('@/views/plugins/PluginVersions.vue'), meta: { title: '版本仓库', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'versions', moduleKey: 'plugins', navGroupKey: 'packages', navPageKey: 'versions' } },
  { path: '/plugins/upgrade-diff', redirect: '/plugins/versions' },
  { path: '/plugins/operations', name: 'pluginOperations', component: () => import('@/views/plugins/PluginOperations.vue'), meta: { title: '操作历史', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'operations', moduleKey: 'plugins', navGroupKey: 'logs', navPageKey: 'operations' } },
  { path: '/plugins/trusted-publishers', name: 'pluginTrustedPublishers', component: () => import('@/views/plugins/PluginTrustedPublishers.vue'), meta: { title: '可信发布者', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'trustedPublishers', moduleKey: 'plugins', navGroupKey: 'security', navPageKey: 'trustedPublishers' } },
  { path: '/plugins/config-keys', name: 'pluginConfigKeys', component: () => import('@/views/plugins/PluginConfigKeys.vue'), meta: { title: '密钥轮换', permission: 'plugin.manage', subNavGroup: 'plugins', subNavKey: 'configKeys', moduleKey: 'plugins', navGroupKey: 'security', navPageKey: 'configKeys' } },
  { path: '/plugins/security', redirect: '/plugins/config-keys' },
  { path: '/plugins/approvals', name: 'pluginsApprovals', component: () => import('@/views/plugins/PluginApprovals.vue'), meta: { title: '安装 / 升级审批', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'approvals', moduleKey: 'plugins', navGroupKey: 'logs', navPageKey: 'approvals' } },
  { path: '/plugins/config', name: 'pluginsConfig', component: () => import('@/views/plugins/PluginConfigHub.vue'), meta: { title: '配置中心', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'config', moduleKey: 'plugins', navGroupKey: 'config', navPageKey: 'config' } },
  { path: '/plugins/dependencies', name: 'pluginsDependencies', component: () => import('@/views/plugins/PluginDependencies.vue'), meta: { title: '依赖 / 兼容性', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'dependencies', moduleKey: 'plugins', navGroupKey: 'security', navPageKey: 'dependencies' } },
  { path: '/plugins/hooks', name: 'pluginsHooks', component: () => import('@/views/plugins/PluginHooks.vue'), meta: { title: 'Hook 排障', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'hooks', moduleKey: 'plugins', navGroupKey: 'runtime', navPageKey: 'hooks' } },
  { path: '/plugins/webhooks', name: 'pluginsWebhooks', component: () => import('@/views/plugins/PluginWebhooks.vue'), meta: { title: 'Webhook 治理', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'webhooks', moduleKey: 'plugins', navGroupKey: 'runtime', navPageKey: 'webhooks' } },
  { path: '/plugins/events', name: 'pluginsEvents', component: () => import('@/views/plugins/PluginEvents.vue'), meta: { title: '事件通知', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'events', moduleKey: 'plugins', navGroupKey: 'runtime', navPageKey: 'events' } },
  { path: '/plugins/search-index', name: 'pluginsSearchIndex', component: () => import('@/views/plugins/PluginSearchIndex.vue'), meta: { title: '搜索索引', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'searchIndex', moduleKey: 'plugins', navGroupKey: 'runtime', navPageKey: 'searchIndex' } },
  { path: '/plugins/navigation', name: 'pluginsNavigation', component: () => import('@/views/plugins/PluginNavigation.vue'), meta: { title: '前台入口', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'navigation', moduleKey: 'plugins', navGroupKey: 'runtime', navPageKey: 'navigation' } },
  { path: '/plugins/permissions', name: 'pluginsPermissions', component: () => import('@/views/plugins/PluginPermissions.vue'), meta: { title: '权限矩阵', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'permissions', moduleKey: 'plugins', navGroupKey: 'manage', navPageKey: 'permissions' } },
  { path: '/plugins/audit', name: 'pluginsAudit', component: () => import('@/views/plugins/PluginAudit.vue'), meta: { title: '审计', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'audit', moduleKey: 'plugins', navGroupKey: 'logs', navPageKey: 'audit' } },
  { path: '/plugins/remote-indexes', name: 'pluginRemoteIndexes', component: () => import('@/views/plugins/PluginRemoteIndexes.vue'), meta: { title: '远程索引', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'remoteIndexes', moduleKey: 'plugins', navGroupKey: 'market', navPageKey: 'remoteIndexes' } },
  { path: '/plugins/developer', name: 'pluginsDeveloper', component: () => import('@/views/plugins/PluginDeveloper.vue'), meta: { title: '开发者工具', permission: 'plugin.read', subNavGroup: 'plugins', subNavKey: 'developer', moduleKey: 'plugins', navGroupKey: 'market', navPageKey: 'developer' } },

  // legacy compat routes
  { path: '/plugins/governance', redirect: '/plugins/overview' },
  { path: '/plugins/manifest', redirect: '/plugins/install' },
  { path: '/plugins/diagnostics', redirect: '/plugins/hooks' },
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
