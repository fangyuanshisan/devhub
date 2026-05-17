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
      key: 'overview',
      title: '插件概览',
      items: [
        { key: 'overview', label: '概览', path: '/plugins/overview', permission: 'plugin.read' },
      ],
    },
    {
      key: 'manage',
      title: '插件管理',
      items: [
        { key: 'list', label: '插件列表', path: '/plugins/list', permission: 'plugin.read' },
        { key: 'content', label: '内容治理', path: '/plugins/content', permission: 'plugin.read' },
        { key: 'permissions', label: '权限矩阵', path: '/plugins/permissions', permission: 'plugin.read' },
      ],
    },
    {
      key: 'packages',
      title: '插件包治理',
      items: [
        { key: 'install', label: '本地插件仓库', path: '/plugins/install', permission: 'plugin.read' },
        { key: 'packageUploads', label: 'zip 上传包', path: '/plugins/packages/uploads', permission: 'plugin.read' },
        { key: 'remotePackages', label: '远程插件包', path: '/plugins/packages/remote', permission: 'plugin.read' },
        { key: 'versions', label: '版本仓库', path: '/plugins/versions', permission: 'plugin.read' },
      ],
    },
    {
      key: 'security',
      title: '插件安全',
      items: [
        { key: 'dependencies', label: '依赖 / 兼容性', path: '/plugins/dependencies', permission: 'plugin.read' },
        { key: 'trustedPublishers', label: '可信发布者', path: '/plugins/trusted-publishers', permission: 'plugin.read' },
        { key: 'configKeys', label: '密钥轮换', path: '/plugins/config-keys', permission: 'plugin.manage' },
      ],
    },
    {
      key: 'config',
      title: '插件配置',
      items: [
        { key: 'config', label: '配置中心', path: '/plugins/config', permission: 'plugin.read' },
      ],
    },
    {
      key: 'runtime',
      title: '运行时治理',
      items: [
        { key: 'hooks', label: 'Hook 排障', path: '/plugins/hooks', permission: 'plugin.read' },
        { key: 'navigation', label: '前台入口', path: '/plugins/navigation', permission: 'plugin.read' },
        { key: 'searchIndex', label: '搜索索引', path: '/plugins/search-index', permission: 'plugin.read' },
        { key: 'webhooks', label: 'Webhook 治理', path: '/plugins/webhooks', permission: 'plugin.read' },
        { key: 'events', label: '事件通知', path: '/plugins/events', permission: 'plugin.read' },
      ],
    },
    {
      key: 'logs',
      title: '插件日志',
      items: [
        { key: 'approvals', label: '安装 / 升级审批', path: '/plugins/approvals', permission: 'plugin.read' },
        { key: 'operations', label: '操作历史', path: '/plugins/operations', permission: 'plugin.read' },
        { key: 'audit', label: '审计', path: '/plugins/audit', permission: 'plugin.read' },
      ],
    },
    {
      key: 'market',
      title: '插件市场 / 开发者',
      items: [
        { key: 'remoteIndexes', label: '远程索引', path: '/plugins/remote-indexes', permission: 'plugin.read' },
        { key: 'developer', label: '开发者工具', path: '/plugins/developer', permission: 'plugin.read' },
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
