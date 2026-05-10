import { ofetch } from 'ofetch';
import { z } from 'zod';
import { fallbackBoards, fallbackPosts, fallbackSites, fallbackTags } from './fallback';
import { boardToContentType, contentTypeToBoard } from './site-config';
import type { Board, CommentNode, CommunityStats, Post, Site, TagStat } from './types';
import { authRequest } from './session';

const API_BASE = process.env.FRONTEND_API_BASE || '';
const API_PREFIX = '/api/v1';

const siteSchema = z.object({
  key: z.string(),
  name: z.string(),
  logo: z.string().default('DH'),
  title: z.string(),
  sub: z.string().default(''),
  pub: z.string().default(''),
  description: z.string().default(''),
  color: z.string().default('#2563eb'),
  status: z.string().default('active'),
  sort: z.number().default(0),
});

const boardSchema = z.object({
  key: z.string(),
  name: z.string(),
  site: z.string().default('portal'),
  sort: z.number().default(0),
  visible: z.boolean().default(true),
});

const communitySchema = z.object({
  id: z.number(),
  name: z.string(),
  slug: z.string(),
  logo: z.string().default(''),
  cover_image: z.string().default(''),
  slogan: z.string().default(''),
  description: z.string().default(''),
  theme_color: z.string().default('#2563eb'),
  seo_title: z.string().default(''),
  seo_description: z.string().default(''),
  seo_keywords: z.string().default(''),
  sort_order: z.number().default(0),
  status: z.number().default(1),
  follower_count: z.number().default(0),
  topic_count: z.number().default(0),
  comment_count: z.number().default(0),
  hot_score: z.number().default(0),
  announcement_title: z.string().default(''),
  announcement_content: z.string().default(''),
  announcement_url: z.string().default(''),
});

const categorySchema = z.object({
  id: z.number(),
  community_id: z.number().default(0),
  name: z.string(),
  slug: z.string(),
  type: z.string().default('article'),
  content_type: z.string().default('article'),
  description: z.string().default(''),
  icon: z.string().default(''),
  sort_order: z.number().default(0),
  visible: z.boolean().default(true),
  nav_visible: z.boolean().default(true),
  postable: z.boolean().default(true),
  status: z.number().default(1),
});

const communityStatsSchema = z.object({
  topic_count: z.number().default(0),
  comment_count: z.number().default(0),
  question_count: z.number().default(0),
  unsolved_count: z.number().default(0),
  follower_count: z.number().default(0),
  today_topic_count: z.number().default(0),
  today_comment_count: z.number().default(0),
  moderator_count: z.number().default(0),
  hot_score: z.number().default(0),
});

const tagStatSchema = z.object({
  id: z.number().optional(),
  name: z.string(),
  slug: z.string().default(''),
  site: z.string().default(''),
  community_id: z.number().default(0),
  community_slug: z.string().default(''),
  description: z.string().default(''),
  topic_count: z.number().default(0),
  count: z.number().default(0),
  follower_count: z.number().default(0),
  status: z.string().default('enable'),
  seo_title: z.string().default(''),
  seo_description: z.string().default(''),
  seo_keywords: z.string().default(''),
});

const tagSchema = tagStatSchema.extend({
  id: z.number(),
  sort: z.number().default(0),
  sort_order: z.number().default(0),
  use_count: z.number().default(0),
  community_name: z.string().default(''),
  created_at: z.string().default(''),
  updated_at: z.string().default(''),
});

const postSchema = z.object({
  id: z.number(),
  site: z.string(),
  board: z.string(),
  title: z.string(),
  summary: z.string().default(''),
  content: z.string().default(''),
  author: z.string().default('DevHub'),
  status: z.string().default('publish'),
  pinned: z.boolean().default(false),
  recommended: z.boolean().default(false),
  solved: z.boolean().default(false),
  content_type: z.string().default('article'),
  favorite_count: z.number().default(0),
  hot_score: z.number().default(0),
  last_active_at: z.string().default(''),
  ai_summary: z.string().default(''),
  views: z.number().default(0),
  likes: z.number().default(0),
  comments: z.number().default(0),
  tags: z.array(z.string()).default([]),
  created_at: z.string().default(''),
  updated_at: z.string().default(''),
});

const topicSchema = z.object({
  id: z.number(),
  community_id: z.number().default(0),
  category_id: z.number().default(0),
  title: z.string(),
  user_id: z.number().default(0),
  slug: z.string().default(''),
  content_type: z.string().default('article'),
  summary: z.string().default(''),
  content: z.string().default(''),
  status: z.number().default(1),
  is_pinned: z.boolean().default(false),
  is_featured: z.boolean().default(false),
  is_solved: z.boolean().default(false),
  view_count: z.number().default(0),
  comment_count: z.number().default(0),
  like_count: z.number().default(0),
  favorite_count: z.number().default(0),
  hot_score: z.number().default(0),
  last_active_at: z.string().default(''),
  ai_summary: z.string().default(''),
  tags: z.array(z.string()).default([]),
  created_at: z.string().default(''),
  updated_at: z.string().default(''),
});

export function endpoint(path: string) {
  if (path.startsWith('http')) return path;
  const apiPath = path.startsWith('/api/') ? path : `${API_PREFIX}${path}`;
  if (!API_BASE) return apiPath;
  return `${API_BASE.replace(/\/$/, '')}${apiPath}`;
}

async function request<T>(path: string, fallback: T, schema?: z.ZodTypeAny): Promise<T> {
  try {
    const data = await ofetch(endpoint(path), { retry: 0, timeout: 3000 });
    if (!schema) return data as T;
    return schema.parse(data) as T;
  } catch (error) {
    console.warn(`[frontend] API fallback for ${path}:`, error);
    return fallback;
  }
}

export async function getSites(): Promise<Site[]> {
  const communityData = await request('/communities', { items: [] }, z.object({ items: z.array(communitySchema) }));
  if (communityData.items.length > 0) {
    const communities = communityData.items.map(communityToSite);
    return [fallbackSites[0], ...communities];
  }
  const schema = z.object({ items: z.array(siteSchema) });
  const data = await request('/sites', { items: fallbackSites }, schema);
  return data.items;
}

export async function getCommunity(slug: string): Promise<Site | undefined> {
  const data = await request(`/communities/${encodeURIComponent(slug)}`, undefined, communitySchema.optional());
  return data ? communityToSite(data) : undefined;
}

export async function getCommunityStats(slug: string): Promise<CommunityStats> {
  return request(`/communities/${encodeURIComponent(slug)}/stats`, {
    topic_count: 0,
    comment_count: 0,
    question_count: 0,
    unsolved_count: 0,
    follower_count: 0,
    today_topic_count: 0,
    today_comment_count: 0,
    moderator_count: 0,
    hot_score: 0,
  }, communityStatsSchema);
}

export async function getCommunityCategories(slug: string): Promise<Board[]> {
  const data = await request(`/communities/${encodeURIComponent(slug)}/categories`, { items: [] }, z.object({ items: z.array(categorySchema) }));
  return data.items.map((category) => categoryToBoard(category, slug));
}

export async function getCommunityTags(slug: string): Promise<TagStat[]> {
  return getTags(slug);
}

export async function getCommunityModerators(slug: string) {
  const data = await request(`/communities/${encodeURIComponent(slug)}/moderators`, { items: [] });
  return data.items || [];
}

export async function getCommunityTopics(slug: string, params: { board?: string; sort?: string; page_size?: number; featured?: boolean } = {}): Promise<Post[]> {
  return getPosts({ ...params, site: slug });
}

export async function followCommunity(id: number): Promise<{ followed: boolean }> {
  const response = await authRequest(endpoint('/follows/toggle'), {
    method: 'POST',
    body: { target_type: 'community', target_id: id },
  });
  return response as { followed: boolean };
}

export async function getBoards(): Promise<Board[]> {
  const sites = await getSites();
  const communityBoards = await Promise.all(
    sites.filter((site) => site.key !== 'portal').map(async (site) => {
      const data = await request(`/communities/${encodeURIComponent(site.key)}/categories`, { items: [] }, z.object({ items: z.array(categorySchema) }));
      return data.items.map((category) => categoryToBoard(category, site.key));
    }),
  );
  const merged = dedupeBoards([fallbackBoards, ...communityBoards].flat());
  if (merged.length > 0) return merged;
  const schema = z.object({ items: z.array(boardSchema) });
  const data = await request('/boards', { items: fallbackBoards }, schema);
  return data.items;
}

export async function getPosts(params: { site?: string; board?: string; tag?: string; q?: string; sort?: string; page_size?: number; featured?: boolean } = {}): Promise<Post[]> {
  const topicData = await getTopics(params);
  if (topicData.length > 0) return topicData;

  const search = new URLSearchParams();
  search.set('site', params.site || 'portal');
  search.set('board', params.board || 'all');
  search.set('page_size', String(params.page_size || 12));
  if (params.tag) search.set('tag', params.tag);
  if (params.q) search.set('q', params.q);
  if (params.sort) search.set('sort', params.sort);
  const fallback = filterPosts(params);
  const data = await request(`/posts?${search.toString()}`, { items: fallback }, z.object({ items: z.array(postSchema) }));
  return data.items;
}

export async function getHotPosts(site = 'portal', limit = 8): Promise<Post[]> {
  const topics = await getPosts({ site, sort: 'hot', page_size: limit });
  if (topics.length > 0) return topics.slice(0, limit);
  const fallback = filterPosts({ site }).slice(0, limit);
  const schema = z.object({ items: z.array(postSchema) });
  const data = await request(`/hot?site=${encodeURIComponent(site)}&limit=${limit}`, { items: fallback }, schema);
  return data.items;
}

export async function getTags(site = 'portal'): Promise<TagStat[]> {
  if (site !== 'portal') {
    const data = await request(`/communities/${encodeURIComponent(site)}/tags`, { items: [] }, z.object({ items: z.array(tagStatSchema) }));
    if (data.items.length > 0) return data.items;
  } else {
    const communities = (await getSites()).filter((item) => item.key !== 'portal');
    const groups = await Promise.all(communities.map((item) => getTags(item.key).catch(() => [] as TagStat[])));
    const counts = new Map<string, number>();
    for (const tag of groups.flat()) counts.set(tag.name, (counts.get(tag.name) || 0) + tag.count);
    const aggregated = [...counts.entries()].map(([name, count]) => ({ name, count })).sort((a, b) => b.count - a.count || a.name.localeCompare(b.name));
    if (aggregated.length > 0) return aggregated.slice(0, 30);
  }
  const schema = z.object({ items: z.array(tagStatSchema) });
  const data = await request(`/tags?site=${encodeURIComponent(site)}`, { items: fallbackTags }, schema);
  return data.items;
}

export async function getTag(slug: string, communitySlug = '') {
  if (communitySlug) {
    return request(`/communities/${encodeURIComponent(communitySlug)}/tags/${encodeURIComponent(slug)}`, undefined, tagSchema.optional());
  }
  const query = new URLSearchParams();
  return request(`/tags/${encodeURIComponent(slug)}${query.toString() ? `?${query}` : ''}`, undefined, tagSchema.optional());
}

export async function getTagTopics(slug: string, params: { community_slug?: string; content_type?: string; sort?: string; page?: number; page_size?: number } = {}) {
  const query = new URLSearchParams();
  const base = params.community_slug
    ? `/communities/${encodeURIComponent(params.community_slug)}/tags/${encodeURIComponent(slug)}/topics`
    : `/tags/${encodeURIComponent(slug)}/topics`;
  if (params.community_slug && !base.startsWith('/communities/')) query.set('community_slug', params.community_slug);
  if (params.content_type) query.set('content_type', params.content_type);
  if (params.sort) query.set('sort', params.sort);
  query.set('page', String(params.page || 1));
  query.set('page_size', String(params.page_size || 12));
  const data = await request(`${base}?${query}`, { items: [], total: 0, page: 1, page_size: 12 }, z.object({ items: z.array(topicSchema), total: z.number().default(0), page: z.number().default(1), page_size: z.number().default(12) }));
  return { ...data, items: data.items.map(topicToPost) };
}

export async function getTagSuggestions(params: { community_slug?: string; q?: string; limit?: number } = {}): Promise<TagStat[]> {
  const query = new URLSearchParams();
  if (params.community_slug) query.set('community_slug', params.community_slug);
  if (params.q) query.set('q', params.q);
  query.set('limit', String(params.limit || 20));
  const data = await request(`/tags/suggestions?${query}`, { items: [] }, z.object({ items: z.array(tagStatSchema) }));
  return data.items;
}

export async function followTag(id: number): Promise<{ followed: boolean }> {
  return authRequest('/api/v1/follows/toggle', {
    method: 'POST',
    body: { target_type: 'tag', target_id: id },
  });
}

export async function getPost(id: number): Promise<Post | undefined> {
  const fallback = fallbackPosts.find((post) => post.id === id);
  try {
    const topic = await request<z.infer<typeof topicSchema> | undefined>(`/topics/${id}`, undefined, topicSchema);
    if (topic) return topicToPost(topic);
  } catch {
    // request() already falls back; keep the old posts API as compatibility.
  }
  if (!fallback) return undefined;
  return request(`/posts/${id}`, fallback, postSchema);
}

export async function getComments(postID: number): Promise<CommentNode[]> {
  const topicData = await request(`/topics/${postID}/comments`, { items: [] as CommentNode[] });
  if ((topicData.items || []).length > 0) return topicData.items || [];
  const data = await request(`/posts/${postID}/comments`, { items: [] as CommentNode[] });
  return data.items || [];
}

export async function allStaticPosts(): Promise<Post[]> {
  return getPosts({ site: 'portal', page_size: 100 });
}

export async function searchPosts(params: { site?: string; board?: string; tag?: string; q?: string; sort?: string; page_size?: number } = {}): Promise<Post[]> {
  const topics = await getTopics(params, true);
  if (topics.length > 0 || params.q || params.tag) return topics;
  return getPosts(params);
}

async function getTopics(params: { site?: string; board?: string; tag?: string; q?: string; sort?: string; page_size?: number; featured?: boolean } = {}, useSearch = false): Promise<Post[]> {
  const query = new URLSearchParams();
  const site = params.site || 'portal';
  if (site && site !== 'portal') query.set('community_slug', site);
  query.set('page_size', String(params.page_size || 12));
  query.set('sort', params.sort || 'latest');
  const contentType = boardToContentType(params.board || 'all');
  if (contentType) query.set('content_type', contentType);
  if (params.tag) query.set('tag', params.tag);
  if (params.q) query.set('keyword', params.q);
  if (params.featured || params.sort === 'featured') query.set('is_featured', '1');
  const path = useSearch || params.q ? `/search/topics?${query.toString()}` : `/topics?${query.toString()}`;
  const data = await request(path, { items: [] }, z.object({ items: z.array(topicSchema) }));
  return data.items.map(topicToPost).filter((post) => !params.featured || post.recommended);
}

function filterPosts(params: { site?: string; board?: string; tag?: string; q?: string; sort?: string }): Post[] {
  const filtered = fallbackPosts.filter((post) => {
    const siteMatched = !params.site || params.site === 'portal' || post.site === params.site;
    const boardMatched = !params.board || params.board === 'all' || post.board === params.board;
    const tagMatched = !params.tag || post.tags.includes(params.tag);
    const qMatched = !params.q || `${post.title} ${post.summary} ${post.content}`.toLowerCase().includes(params.q.toLowerCase());
    const solvedMatched = params.sort !== 'unsolved' || (post.content_type === 'question' && !post.solved);
    const featuredMatched = params.sort !== 'featured' || post.recommended;
    return siteMatched && boardMatched && tagMatched && qMatched && solvedMatched && featuredMatched;
  });
  return filtered.sort((a, b) => {
    if (params.sort === 'hot') return (b.hot_score || b.views + b.comments * 5 + b.likes * 3) - (a.hot_score || a.views + a.comments * 5 + a.likes * 3);
    if (params.sort === 'active') return Date.parse(b.last_active_at || b.updated_at || b.created_at) - Date.parse(a.last_active_at || a.updated_at || a.created_at);
    return Date.parse(b.created_at) - Date.parse(a.created_at);
  });
}

function topicToPost(topic: z.infer<typeof topicSchema>): Post {
  const site = { 1: 'php', 2: 'go', 3: 'java', 4: 'ai', 5: 'frontend' }[topic.community_id] || 'portal';
  const board = contentTypeToBoard(topic.content_type);
  return {
    id: topic.id,
    site,
    board,
    title: topic.title,
    summary: topic.summary,
    content: topic.content,
    author: '社区用户',
    status: topic.status === 1 ? 'publish' : 'draft',
    pinned: topic.is_pinned,
    recommended: topic.is_featured,
    solved: topic.is_solved,
    content_type: topic.content_type,
    favorite_count: topic.favorite_count,
    hot_score: topic.hot_score,
    last_active_at: topic.last_active_at,
    ai_summary: topic.ai_summary,
    views: topic.view_count,
    likes: topic.like_count,
    comments: topic.comment_count,
    tags: topic.tags,
    created_at: topic.created_at,
    updated_at: topic.updated_at,
  };
}

function communityToSite(community: z.infer<typeof communitySchema>): Site {
  const fallback = fallbackSites.find((site) => site.key === community.slug);
  return {
    key: community.slug,
    name: community.name,
    logo: community.logo || fallback?.logo || community.name.slice(0, 2),
    cover_image: community.cover_image,
    title: community.seo_title || fallback?.title || `${community.name} 开发者站`,
    sub: community.slogan || fallback?.sub || community.description,
    pub: fallback?.pub || `发布 ${community.name} 内容`,
    description: community.description || fallback?.description || `${community.name} 技术社区`,
    color: community.theme_color || fallback?.color || '#2563eb',
    status: community.status === 1 ? 'active' : 'disabled',
    sort: community.sort_order,
    slogan: community.slogan,
    seo_title: community.seo_title,
    seo_description: community.seo_description,
    seo_keywords: community.seo_keywords,
    follower_count: community.follower_count,
    topic_count: community.topic_count,
    comment_count: community.comment_count,
    hot_score: community.hot_score,
    announcement_title: community.announcement_title,
    announcement_content: community.announcement_content,
    announcement_url: community.announcement_url,
  };
}

function categoryToBoard(category: z.infer<typeof categorySchema>, site: string): Board {
  return {
    key: category.slug,
    name: category.name,
    site,
    type: category.content_type || category.type,
    sort: category.sort_order,
    visible: category.visible && category.nav_visible && category.status === 1,
  };
}

function dedupeBoards(boards: Board[]): Board[] {
  const seen = new Set<string>();
  return boards
    .filter((board) => board.visible)
    .sort((a, b) => a.sort - b.sort)
    .filter((board) => {
      if (seen.has(board.key)) return false;
      seen.add(board.key);
      return true;
    });
}
