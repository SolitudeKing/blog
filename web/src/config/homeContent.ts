import type { HomeContent, LobbySetting } from '@/types/setting'

/**
 * 主页文案（home-intro + latest posts）的字段目录。
 *
 * 这一组字段聚合在 `site_settings.home_content` 一列 JSON 中，与
 * `theme_elements` 的差异在于：不分主题，整组替换。后台 Settings 页的
 * 「主页文案」段直接遍历 `homeContentFields` 渲染；新增字段只动本文件
 * + 后端 `homecontent` 包的常量。
 */
export type HomeContentKey = keyof HomeContent

export interface HomeContentField {
  key: HomeContentKey
  label: string
  hint: string
  max: number
  multiline?: boolean
}

// 与后端 homecontent.go 的 *MaxRunes 常量保持一一对应。
export const homeContentFields: ReadonlyArray<HomeContentField> = [
  {
    key: 'home_profile_kicker',
    label: '主页 Profile 小标',
    hint: '头像右侧作者身份上方的英文小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'home_heading_prefix',
    label: '主页问候语前缀',
    hint: '作者名前的称呼语，例如 "你好，我是"；最多 16 字。',
    max: 16,
  },
  {
    key: 'home_status_fallback',
    label: '主页状态回落文案',
    hint: '当没有启用中的 Site Notice 时显示；最多 80 字。',
    max: 80,
  },
  {
    key: 'home_intro_heading',
    label: '主页介绍标题',
    hint: 'home-intro 卡片顶部的大标题；最多 80 字。',
    max: 80,
  },
  {
    key: 'home_intro_paragraph',
    label: '主页介绍段落',
    hint: 'home-intro 卡片标题下的多行描述；最多 240 字。',
    max: 240,
    multiline: true,
  },
  {
    key: 'home_action_view_recent_label',
    label: '"查看最近发布"按钮',
    hint: 'home-intro 卡片左下主操作按钮；最多 16 字。',
    max: 16,
  },
  {
    key: 'home_action_view_archive_label',
    label: '"浏览全部归档"按钮',
    hint: 'home-intro 卡片右下副操作按钮；最多 16 字。',
    max: 16,
  },
  {
    key: 'home_latest_kicker',
    label: '最近发布小标',
    hint: '"Latest posts" 段上方小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'home_latest_heading',
    label: '最近发布标题',
    hint: '"Latest posts" 段标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'home_latest_view_all_label',
    label: '"查看全部归档"链接',
    hint: '"Latest posts" 段右上角"查看全部"链接文案；最多 16 字。',
    max: 16,
  },
  {
    key: 'home_latest_empty_title',
    label: '最近发布空状态标题',
    hint: '没有任何已发布文章时显示的标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'home_topics_kicker',
    label: '专题目录小标',
    hint: '"Topics" 段上方小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'home_topics_heading',
    label: '专题目录标题',
    hint: '"Topics" 段标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'home_notice_kicker',
    label: '站点公告小标',
    hint: '启用中的 Site Notice 卡片上方小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'home_notice_action_label',
    label: '"继续阅读"按钮',
    hint: 'Site Notice 卡片上的跳转按钮；最多 16 字。',
    max: 16,
  },
]

export const defaultHomeContent: Readonly<HomeContent> = {
  home_profile_kicker: 'Blog keeper · Solitude',
  home_heading_prefix: '你好，我是',
  home_status_fallback: '持续记录技术、设计与生活',
  home_intro_heading: '一份持续更新的博客，也是公开的思考现场',
  home_intro_paragraph:
    '在这里记录工程实践、设计系统与构建过程。文章保留可复用的方法，也保留问题发生时的真实判断。',
  home_action_view_recent_label: '查看最近发布',
  home_action_view_archive_label: '浏览全部归档',
  home_latest_kicker: 'Latest posts',
  home_latest_heading: '最近发布的博客',
  home_latest_view_all_label: '查看全部归档',
  home_latest_empty_title: '暂时还没有发布文章',
  home_topics_kicker: 'Topics',
  home_topics_heading: '从这些专题进入',
  home_notice_kicker: 'Site notice',
  home_notice_action_label: '继续阅读',
}

export function createDefaultHomeContent(): HomeContent {
  return { ...defaultHomeContent }
}

type HomeContentInput = Partial<HomeContent> | null | undefined

function resolveText(value: unknown, fallback: string) {
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

export function normalizeHomeContent(value?: HomeContentInput): HomeContent {
  return {
    home_profile_kicker: resolveText(value?.home_profile_kicker, defaultHomeContent.home_profile_kicker),
    home_heading_prefix: resolveText(value?.home_heading_prefix, defaultHomeContent.home_heading_prefix),
    home_status_fallback: resolveText(value?.home_status_fallback, defaultHomeContent.home_status_fallback),
    home_intro_heading: resolveText(value?.home_intro_heading, defaultHomeContent.home_intro_heading),
    home_intro_paragraph: resolveText(
      value?.home_intro_paragraph,
      defaultHomeContent.home_intro_paragraph,
    ),
    home_action_view_recent_label: resolveText(
      value?.home_action_view_recent_label,
      defaultHomeContent.home_action_view_recent_label,
    ),
    home_action_view_archive_label: resolveText(
      value?.home_action_view_archive_label,
      defaultHomeContent.home_action_view_archive_label,
    ),
    home_latest_kicker: resolveText(value?.home_latest_kicker, defaultHomeContent.home_latest_kicker),
    home_latest_heading: resolveText(value?.home_latest_heading, defaultHomeContent.home_latest_heading),
    home_latest_view_all_label: resolveText(
      value?.home_latest_view_all_label,
      defaultHomeContent.home_latest_view_all_label,
    ),
    home_latest_empty_title: resolveText(
      value?.home_latest_empty_title,
      defaultHomeContent.home_latest_empty_title,
    ),
    home_topics_kicker: resolveText(value?.home_topics_kicker, defaultHomeContent.home_topics_kicker),
    home_topics_heading: resolveText(value?.home_topics_heading, defaultHomeContent.home_topics_heading),
    home_notice_kicker: resolveText(value?.home_notice_kicker, defaultHomeContent.home_notice_kicker),
    home_notice_action_label: resolveText(
      value?.home_notice_action_label,
      defaultHomeContent.home_notice_action_label,
    ),
  }
}

/**
 * 渲染层统一入口：始终返回一个完整的 HomeContent，未提供或字段为空白时
 * 回退到 defaultHomeContent。这样模板中可以直接 `{{ homeContent.xxx }}`
 * 而不必处理 undefined / 空字符串。
 */
export function resolveLobbyHomeContent(setting: LobbySetting | null | undefined): HomeContent {
  if (!setting) {
    return { ...defaultHomeContent }
  }
  return normalizeHomeContent(setting.home_content)
}
