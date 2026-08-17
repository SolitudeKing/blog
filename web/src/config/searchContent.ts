import type { LobbySetting, SearchContent } from '@/types/setting'

/**
 * 搜索页文案的字段目录。
 *
 * 与 home_content 同构：聚合在 `site_settings.search_content` 一列 JSON 中，
 * 不分主题，整组替换。后台 Settings 页的「搜索文案」段直接遍历
 * `searchContentFields` 渲染；新增字段只动本文件 + 后端 `sectioncontent`
 * 包的常量。
 */
export type SearchContentKey = keyof SearchContent

export interface SearchContentField {
  key: SearchContentKey
  label: string
  hint: string
  max: number
  multiline?: boolean
}

// 与后端 sectioncontent/search_content.go 的 *MaxRunes 常量保持一一对应。
export const searchContentFields: ReadonlyArray<SearchContentField> = [
  {
    key: 'search_kicker',
    label: '搜索小标',
    hint: '搜索标题上方的小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'search_heading',
    label: '搜索标题',
    hint: 'search-hero 的大标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'search_intro',
    label: '搜索导语',
    hint: '搜索标题下的多行描述；最多 240 字。',
    max: 240,
    multiline: true,
  },
  {
    key: 'search_placeholder',
    label: '搜索框占位符',
    hint: '关键词输入框的占位提示；最多 64 字。',
    max: 64,
  },
  {
    key: 'search_suggestion_label',
    label: '航标标签',
    hint: '"试试这些航标"航标行的标签文案；最多 32 字。',
    max: 32,
  },
  {
    key: 'search_suggestion_fallbacks',
    label: '航标兜底词',
    hint: '当每日航标接口不可用或没有专题/标签时使用的兜底词，每行一个；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'search_empty_title',
    label: '搜索空状态标题',
    hint: '搜索没有命中时显示的标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'search_empty_description',
    label: '搜索空状态描述',
    hint: '空状态标题下的说明；最多 160 字。',
    max: 160,
    multiline: true,
  },
]

export const defaultSearchContent: Readonly<SearchContent> = {
  search_kicker: 'Search the current',
  search_heading: '打捞一段想法',
  search_intro:
    '输入一个词，沿着标题、摘要、正文、专题与标签寻找。也可以从常用航标开始，看看它会把你带去哪里。',
  search_placeholder: '例如：设计系统、写作、Vue……',
  search_suggestion_label: '试试这些航标',
  search_suggestion_fallbacks: '设计\n代码\n写作',
  search_empty_title: '这片水域还没有记录',
  search_empty_description: '可以尝试缩短关键词，或换一个角度重新出发。',
}

export function createDefaultSearchContent(): SearchContent {
  return { ...defaultSearchContent }
}

type SearchContentInput = Partial<SearchContent> | null | undefined

function resolveText(value: unknown, fallback: string) {
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

export function normalizeSearchContent(value?: SearchContentInput): SearchContent {
  return {
    search_kicker: resolveText(value?.search_kicker, defaultSearchContent.search_kicker),
    search_heading: resolveText(value?.search_heading, defaultSearchContent.search_heading),
    search_intro: resolveText(value?.search_intro, defaultSearchContent.search_intro),
    search_placeholder: resolveText(
      value?.search_placeholder,
      defaultSearchContent.search_placeholder,
    ),
    search_suggestion_label: resolveText(
      value?.search_suggestion_label,
      defaultSearchContent.search_suggestion_label,
    ),
    search_suggestion_fallbacks: resolveText(
      value?.search_suggestion_fallbacks,
      defaultSearchContent.search_suggestion_fallbacks,
    ),
    search_empty_title: resolveText(
      value?.search_empty_title,
      defaultSearchContent.search_empty_title,
    ),
    search_empty_description: resolveText(
      value?.search_empty_description,
      defaultSearchContent.search_empty_description,
    ),
  }
}

/**
 * 渲染层统一入口：始终返回一个完整的 SearchContent，未提供或字段为空白时
 * 回退到 defaultSearchContent。这样模板中可以直接 `{{ searchContent.xxx }}`
 * 而不必处理 undefined / 空字符串。
 */
export function resolveLobbySearchContent(
  setting: LobbySetting | null | undefined,
): SearchContent {
  if (!setting) {
    return { ...defaultSearchContent }
  }
  return normalizeSearchContent(setting.search_content)
}

/**
 * 把 search_suggestion_fallbacks 拆成兜底航标词数组；空行与空白词会被过滤。
 */
export function parseSearchSuggestionFallbacks(content: SearchContent): string[] {
  return content.search_suggestion_fallbacks
    .split(/\r?\n/)
    .map((word) => word.trim())
    .filter(Boolean)
}
