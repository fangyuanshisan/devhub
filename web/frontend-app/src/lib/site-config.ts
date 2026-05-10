export const defaultSiteKey = 'portal';

export const communityNavItems = [
  { key: 'article', label: '社区', contentType: 'article' },
  { key: 'question', label: '问答', contentType: 'question' },
  { key: 'doc', label: '文档', contentType: 'doc' },
  { key: 'wiki', label: 'Wiki', contentType: 'wiki' },
  { key: 'project', label: '开源', contentType: 'project' },
  { key: 'ai_work', label: '作品', contentType: 'ai_work' },
] as const;

export const contentTypeOptions = [
  { value: 'article', label: '社区', desc: '适合分享经验、讨论方案和复盘实践。' },
  { value: 'question', label: '问答中心', desc: '适合提出具体问题，后续可采纳答案。' },
  { value: 'project', label: '开源项目', desc: '适合介绍项目、组件、维护经验和协作进展。' },
  { value: 'ai_work', label: '智能作品', desc: '适合发布工具、自动化流程和实验型作品。' },
  { value: 'job', label: '招聘内推', desc: '适合发布招聘、内推和协作岗位信息。' },
  { value: 'wiki', label: 'Wiki', desc: '适合沉淀知识索引和可维护条目。' },
  { value: 'doc', label: '文档', desc: '适合发布教程、手册和实践指南。' },
] as const;

export const sortTabs = [
  { key: 'latest', name: '最新' },
  { key: 'active', name: '活跃' },
  { key: 'hot', name: '热门' },
  { key: 'featured', name: '精华' },
  { key: 'unsolved', name: '未解决' },
] as const;

export const searchCategories = [
  { key: 'all', type: '', name: '全部' },
  ...contentTypeOptions.map((item) => ({ key: item.value, type: item.value, name: item.label })),
] as const;

export const contentTypeLabelMap: Record<string, string> = Object.fromEntries(
  contentTypeOptions.map((item) => [item.value, item.label]),
);

contentTypeLabelMap.news = '动态';

export function contentTypeLabel(contentType: string) {
  return contentTypeLabelMap[contentType] || '内容';
}

export function contentTypeCommentLabel(contentType: string) {
  return contentType === 'question' ? '回答' : '评论';
}

export function boardToContentType(board: string) {
  return {
    community: 'article',
    qa: 'question',
    opensource: 'project',
    ai: 'ai_work',
    jobs: 'job',
    wiki: 'wiki',
    docs: 'doc',
  }[board] || '';
}

export function contentTypeToBoard(contentType: string) {
  return {
    article: 'community',
    question: 'qa',
    project: 'opensource',
    ai_work: 'ai',
    job: 'jobs',
    wiki: 'wiki',
    doc: 'docs',
    news: 'community',
  }[contentType] || 'community';
}

export function tagSegment(tag: string) {
  return encodeURIComponent(String(tag).trim().toLowerCase().replace(/\s+/g, '-'));
}
