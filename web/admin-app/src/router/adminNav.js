// Admin navigation config (module -> functional groups -> pages).
//
// Principles:
// - Level 1: modules (global domains)
// - Level 2: functional groups under a module
// - Level 3: concrete pages (clickable routes)
// - In-page tabs are NOT part of this config (should stay within pages and sync to URL query)

export const adminModules = [
  { key: 'home', title: '首页', short: '首页', icon: 'DataBoard', entry: '/dashboard', permission: 'dashboard.read' },
  { key: 'users', title: '用户', short: '用户', icon: 'User', entry: '/users', permission: 'user.read' },
  { key: 'content', title: '内容', short: '内容', icon: 'Document', entry: '/content', permission: 'post.read' },
  { key: 'communities', title: '社区', short: '社区', icon: 'SetUp', entry: '/communities', permission: 'site.read' },
  { key: 'plugins', title: '插件', short: '插件', icon: 'Connection', entry: '/plugins/overview', permission: 'plugin.read' },
  { key: 'system', title: '系统', short: '系统', icon: 'Setting', entry: '/system', permission: 'setting.read' },
  { key: 'maintenance', title: '维护', short: '维护', icon: 'Tickets', entry: '/audit-logs', permission: 'log.read' },
];

export const adminModuleGroups = {
  home: [
    { key: 'overview', title: '概览', items: [{ key: 'dashboard', label: '控制台', path: '/dashboard', permission: 'dashboard.read' }] },
    { key: 'stats', title: '统计', items: [{ key: 'statistics', label: '数据统计', path: '/statistics', permission: 'dashboard.read' }] },
  ],
  users: [
    { key: 'accounts', title: '账号与权限', items: [{ key: 'users', label: '用户权限', path: '/users', permission: 'user.read' }] },
  ],
  content: [
    {
      key: 'governance',
      title: '内容治理',
      items: [
        { key: 'content', label: '内容管理', path: '/content', permission: 'post.read' },
        { key: 'comments', label: '评论审核', path: '/comments', permission: 'comment.read' },
        { key: 'reports', label: '举报管理', path: '/reports', permission: 'report.read' },
        { key: 'tags', label: '标签管理', path: '/tags', permission: 'post.read' },
      ],
    },
  ],
  communities: [
    { key: 'base', title: '子站治理', items: [{ key: 'communities', label: '子站管理', path: '/communities', permission: 'site.read' }] },
    { key: 'moderation', title: '治理授权', items: [{ key: 'moderators', label: '版主管理', path: '/moderators', permission: 'moderator.read' }] },
  ],
  plugins: [
    {
      key: 'governance',
      title: '插件管理',
      items: [
        { key: 'overview', label: '插件总览', path: '/plugins/overview', permission: 'plugin.read' },
        { key: 'packages', label: '插件包治理', path: '/plugins/packages', permission: 'plugin.read' },
        { key: 'webhooks', label: 'Webhook 治理', path: '/plugins/webhooks', permission: 'plugin.read' },
        { key: 'publishers', label: '发布者与信任', path: '/plugins/publishers', permission: 'plugin.read' },
        { key: 'runtime', label: '运行记录 / 审计', path: '/plugins/runtime', permission: 'plugin.read' },
      ],
    },
  ],
  system: [
    { key: 'settings', title: '基础设置', items: [{ key: 'system', label: '系统设置', path: '/system', permission: 'setting.read' }] },
    { key: 'operation', title: '运营与通知', items: [{ key: 'operation', label: '运营工具', path: '/operation', permission: 'notification.write' }] },
  ],
  maintenance: [
    { key: 'audit', title: '日志审计', items: [{ key: 'auditLogs', label: '治理审计', path: '/audit-logs', permission: 'log.read' }] },
  ],
};
