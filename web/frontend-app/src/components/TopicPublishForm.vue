<template>
  <section class="publish-layout">
    <form class="publish-form" @submit.prevent="submit">
      <div v-if="message" class="form-message" :class="{ error: messageType === 'error', success: messageType === 'success' }">{{ message }}</div>

      <label>
        <span>所属子站</span>
        <select v-model="form.community_slug" @change="loadCommunityData">
          <option value="">请选择子站</option>
          <option v-for="community in communities" :key="community.slug" :value="community.slug">{{ community.name }}</option>
        </select>
      </label>

      <label>
        <span>所属板块</span>
        <select v-model.number="form.category_id" @change="syncCategoryType">
          <option :value="0">请选择板块</option>
          <option v-for="category in categories" :key="category.id" :value="category.id">{{ category.name }}</option>
        </select>
      </label>

      <label>
        <span>内容类型</span>
        <select v-model="form.content_type" @change="selectCategoryForType(form.content_type)">
          <option v-for="item in contentTypes" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
      </label>

      <label>
        <span>标题</span>
        <input v-model.trim="form.title" maxlength="120" placeholder="4 到 120 个字符" />
      </label>

      <label>
        <span>摘要</span>
        <textarea v-model.trim="form.summary" maxlength="300" rows="3" placeholder="可选，最多 300 个字符" />
      </label>

      <label>
        <span>正文</span>
        <textarea v-model.trim="form.content" rows="14" placeholder="支持 Markdown 原文，至少 10 个字符" />
      </label>

      <fieldset class="tag-picker">
        <legend>标签</legend>
        <p>最多选择 5 个标签，标签来自当前子站已启用标签。</p>
        <input v-model.trim="tagKeyword" type="search" placeholder="搜索标签建议" @input="loadTagSuggestions" />
        <div>
          <label v-for="tag in tags" :key="tag.name">
            <input type="checkbox" :value="tag.name" :checked="form.tags.includes(tag.name)" @change="toggleTag(tag.name)" />
            <span>{{ tag.name }}</span>
            <small>{{ tag.topic_count || tag.count || 0 }}</small>
          </label>
        </div>
        <p v-if="!tags.length">当前子站暂无可选标签，请先联系管理员在后台启用标签。</p>
      </fieldset>

      <div class="publish-actions">
        <button type="submit" :disabled="submitting">{{ submitting ? '发布中...' : '发布' }}</button>
        <a :href="cancelHref">取消</a>
      </div>
    </form>

    <aside class="publish-help">
      <h2>发布说明</h2>
      <p>请选择合适的子站和板块。问答、开源项目、AI 作品、招聘、Wiki 和文档都会统一进入 DevHub Topic 内容流。</p>
      <h3>当前子站</h3>
      <p>{{ currentCommunity?.description || '从总站发布时，需要先选择一个技术子站。' }}</p>
      <h3>内容类型</h3>
      <p>{{ currentTypeDescription }}</p>
    </aside>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ofetch } from 'ofetch';
import { authRequest, hasSession } from '../lib/session';

const props = defineProps<{ defaultCommunity?: string }>();

type Community = { id: number; name: string; slug: string; description: string };
type Category = { id: number; name: string; slug: string; type?: string; content_type?: string; postable?: boolean; status?: number };
type Tag = { id?: number; name: string; slug?: string; count: number; topic_count?: number; description?: string };

const contentTypes = [
  { value: 'article', label: '社区', desc: '适合分享经验、讨论方案和复盘实践。' },
  { value: 'question', label: '问答中心', desc: '适合提出具体问题，后续可采纳答案。' },
  { value: 'project', label: '开源项目', desc: '适合介绍项目、组件、工具和维护经验。' },
  { value: 'ai_work', label: 'AI作品', desc: '适合发布 Agent、Prompt、RAG 和自动化工作流。' },
  { value: 'job', label: '招聘内推', desc: '适合发布招聘、内推和团队岗位信息。' },
  { value: 'wiki', label: 'Wiki', desc: '适合沉淀知识索引和可维护条目。' },
  { value: 'doc', label: '文档', desc: '适合发布教程、手册和实践指南。' },
];

const communities = ref<Community[]>([]);
const categories = ref<Category[]>([]);
const tags = ref<Tag[]>([]);
const allowedTagNames = ref<Set<string>>(new Set());
const tagKeyword = ref('');
let tagRequestID = 0;
const submitting = ref(false);
const message = ref('');
const messageType = ref<'error' | 'success'>('error');

const form = reactive({
  community_slug: props.defaultCommunity || '',
  category_id: 0,
  content_type: initialContentType(),
  title: '',
  summary: '',
  content: '',
  tags: [] as string[],
});

const currentCommunity = computed(() => communities.value.find((item) => item.slug === form.community_slug));
const currentTypeDescription = computed(() => contentTypes.find((item) => item.value === form.content_type)?.desc || '');
const cancelHref = computed(() => form.community_slug ? `/c/${form.community_slug}/` : '/');

onMounted(async () => {
  await loadCommunities();
  if (!form.community_slug && communities.value[0]) form.community_slug = communities.value[0].slug;
  await loadCommunityData();
});

async function loadCommunities() {
  try {
    const data = await ofetch('/api/v1/communities');
    communities.value = data.items || [];
  } catch {
    setMessage('子站加载失败，请稍后刷新重试');
  }
}

async function loadCommunityData() {
  form.category_id = 0;
  form.tags = [];
  tagKeyword.value = '';
  categories.value = [];
  tags.value = [];
  allowedTagNames.value = new Set();
  if (!form.community_slug) return;
  try {
    const [categoryData, tagData] = await Promise.all([
      ofetch(`/api/v1/communities/${encodeURIComponent(form.community_slug)}/categories`),
      ofetch(`/api/v1/tags/suggestions?community_slug=${encodeURIComponent(form.community_slug)}&limit=30`),
    ]);
    categories.value = categoryData.items || [];
    setTags(tagData.items || []);
    selectCategoryForType(form.content_type);
  } catch {
    setMessage('板块或标签加载失败，请切换子站后重试');
  }
}

async function loadTagSuggestions() {
  if (!form.community_slug) return;
  const requestID = ++tagRequestID;
  try {
    const query = new URLSearchParams({
      community_slug: form.community_slug,
      q: tagKeyword.value,
      limit: '30',
    });
    const data = await ofetch(`/api/v1/tags/suggestions?${query}`);
    if (requestID === tagRequestID) setTags(data.items || []);
  } catch {
    if (requestID === tagRequestID) tags.value = [];
  }
}

function setTags(items: Tag[]) {
  tags.value = items;
  const next = new Set(allowedTagNames.value);
  for (const tag of items) {
    if (tag.name) next.add(tag.name);
  }
  allowedTagNames.value = next;
}

function syncCategoryType() {
  const category = categories.value.find((item) => item.id === form.category_id);
  const categoryType = categoryContentType(category);
  if (categoryType) form.content_type = categoryType;
}

function initialContentType() {
  if (typeof window === 'undefined') return 'article';
  const value = new URLSearchParams(window.location.search).get('type') || new URLSearchParams(window.location.search).get('content_type') || 'article';
  return contentTypes.some((item) => item.value === value) ? value : 'article';
}

function categoryContentType(category?: Category) {
  return category?.content_type || category?.type || '';
}

function selectableCategories() {
  return categories.value.filter((category) => category.status !== 0 && category.postable !== false);
}

function selectCategoryForType(contentType: string) {
  const matched = selectableCategories().find((category) => categoryContentType(category) === contentType);
  const fallback = selectableCategories()[0] || categories.value[0];
  if (matched) {
    form.category_id = matched.id;
    form.content_type = categoryContentType(matched);
    return;
  }
  if (fallback) {
    form.category_id = fallback.id;
    form.content_type = categoryContentType(fallback) || form.content_type;
  }
}

function toggleTag(name: string) {
  if (form.tags.includes(name)) {
    form.tags = form.tags.filter((item) => item !== name);
    return;
  }
  if (form.tags.length >= 5) {
    setMessage('最多选择 5 个标签');
    return;
  }
  form.tags.push(name);
}

function validate() {
  if (!form.community_slug) return '请选择子站';
  if (!form.category_id) return '请选择板块';
  const selectedCategory = categories.value.find((item) => item.id === form.category_id);
  if (!selectedCategory) return '请选择当前子站下的板块';
  const selectedType = categoryContentType(selectedCategory);
  if (selectedType && selectedType !== form.content_type) {
    const target = selectableCategories().find((category) => categoryContentType(category) === form.content_type);
    return target ? `当前板块不支持该内容类型，请切换到「${target.name}」板块` : `当前子站未配置${contentTypes.find((item) => item.value === form.content_type)?.label || form.content_type}板块`;
  }
  if (form.title.length < 4 || form.title.length > 120) return '标题长度需为 4 到 120 个字符';
  if (form.summary.length > 300) return '摘要最多 300 个字符';
  if (form.content.length < 10) return '正文至少 10 个字符';
  if (form.tags.length > 5) return '最多选择 5 个标签';
  const allowed = allowedTagNames.value;
  if (form.tags.some((tag) => !allowed.has(tag))) return '请选择当前子站已启用标签';
  return '';
}

async function submit() {
  if (!hasSession()) {
    setMessage('请先登录后再发布内容');
    return;
  }
  const error = validate();
  if (error) {
    setMessage(error);
    return;
  }
  submitting.value = true;
  message.value = '';
  try {
    const result = await authRequest<any>('/api/v1/topics', {
      method: 'POST',
      body: {
        community_slug: form.community_slug,
        category_id: form.category_id,
        content_type: form.content_type,
        title: form.title,
        summary: form.summary,
        content: form.content,
        tags: form.tags,
        status: 'published',
      },
    });
    messageType.value = 'success';
    message.value = '发布成功，正在跳转...';
    const detailURL = result.detail_url || (result.id ? `/topics/${result.id}/` : `/c/${form.community_slug}/`);
    window.location.assign(detailURL);
  } catch (error: any) {
    setMessage(error?.data?.error || '发布失败，请稍后再试');
  } finally {
    submitting.value = false;
  }
}

function setMessage(value: string) {
  messageType.value = 'error';
  message.value = value;
}
</script>
