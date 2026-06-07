export type Site = {
  key: string;
  name: string;
  logo: string;
  cover_image?: string;
  title: string;
  sub: string;
  pub: string;
  description: string;
  color: string;
  status: string;
  sort: number;
  slogan?: string;
  seo_title?: string;
  seo_description?: string;
  seo_keywords?: string;
  follower_count?: number;
  topic_count?: number;
  comment_count?: number;
  hot_score?: number;
  announcement_title?: string;
  announcement_content?: string;
  announcement_url?: string;
};

export type Board = {
  key: string;
  name: string;
  site: string;
  type?: string;
  sort: number;
  visible: boolean;
};

export type Post = {
  id: number;
  site: string;
  board: string;
  title: string;
  summary: string;
  content: string;
  author: string;
  status: string;
  pinned: boolean;
  recommended: boolean;
  solved?: boolean;
  content_type?: string;
  favorite_count?: number;
  hot_score?: number;
  last_active_at?: string;
  ai_summary?: string;
  views: number;
  likes: number;
  comments: number;
  tags: string[];
  created_at: string;
  updated_at: string;
};

export type TagStat = {
  id?: number;
  name: string;
  slug?: string;
  site?: string;
  community_id?: number;
  community_slug?: string;
  description?: string;
  topic_count?: number;
  count: number;
  follower_count?: number;
  status?: string;
  seo_title?: string;
  seo_description?: string;
  seo_keywords?: string;
};

export type CommunityStats = {
  topic_count: number;
  comment_count: number;
  question_count: number;
  unsolved_count: number;
  follower_count: number;
  today_topic_count: number;
  today_comment_count: number;
  moderator_count: number;
  hot_score: number;
};

export type CommentNode = {
  id: number;
  post_id: number;
  parent_id: number;
  author: string;
  to?: string;
  text: string;
  status: string;
  likes: number;
  created_at: string;
  replies?: CommentNode[];
};
