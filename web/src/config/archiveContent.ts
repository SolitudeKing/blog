import type { ArchiveContent, LobbySetting } from '@/types/setting'

/**
 * 归档页文案的字段目录。
 *
 * 与 home_content 同构：聚合在 `site_settings.archive_content` 一列 JSON 中，
 * 不分主题，整组替换。后台 Settings 页的「归档文案」段直接遍历
 * `archiveContentFields` 渲染；新增字段只动本文件 + 后端 `sectioncontent`
 * 包的常量。
 */
export type ArchiveContentKey = keyof ArchiveContent

export interface ArchiveContentField {
  key: ArchiveContentKey
  label: string
  hint: string
  max: number
  multiline?: boolean
}

// 与后端 sectioncontent/archive_content.go 的 *MaxRunes 常量保持一一对应。
export const archiveContentFields: ReadonlyArray<ArchiveContentField> = [
  {
    key: 'archive_kicker',
    label: '归档小标',
    hint: '归档标题上方的小标，页面会在其后追加年份范围，例如 "Archive · 2024—2026"；最多 32 字。',
    max: 32,
  },
  {
    key: 'archive_heading',
    label: '归档标题',
    hint: 'archive-hero 卡片的大标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'archive_intro',
    label: '归档导语',
    hint: '归档标题下的多行描述；最多 240 字。',
    max: 240,
    multiline: true,
  },
  {
    key: 'archive_empty_title',
    label: '归档空状态标题',
    hint: '没有任何已发布文章时显示的标题；最多 64 字。',
    max: 64,
  },
  {
    key: 'archive_empty_description',
    label: '归档空状态描述',
    hint: '空状态标题下的说明；最多 160 字。',
    max: 160,
    multiline: true,
  },
]

export const defaultArchiveContent: Readonly<ArchiveContent> = {
  archive_kicker: 'Archive',
  archive_heading: '所有足迹，都有刻度',
  archive_intro:
    '从最近一次发布向过去回望。这里按年份与月份整理已经公开的文章，让每一段记录都能被重新抵达。',
  archive_empty_title: '还没有归档内容',
  archive_empty_description: '发布文章后会按年/月自动汇总到这里。',
}

export function createDefaultArchiveContent(): ArchiveContent {
  return { ...defaultArchiveContent }
}

type ArchiveContentInput = Partial<ArchiveContent> | null | undefined

function resolveText(value: unknown, fallback: string) {
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

export function normalizeArchiveContent(value?: ArchiveContentInput): ArchiveContent {
  return {
    archive_kicker: resolveText(value?.archive_kicker, defaultArchiveContent.archive_kicker),
    archive_heading: resolveText(value?.archive_heading, defaultArchiveContent.archive_heading),
    archive_intro: resolveText(value?.archive_intro, defaultArchiveContent.archive_intro),
    archive_empty_title: resolveText(
      value?.archive_empty_title,
      defaultArchiveContent.archive_empty_title,
    ),
    archive_empty_description: resolveText(
      value?.archive_empty_description,
      defaultArchiveContent.archive_empty_description,
    ),
  }
}

/**
 * 渲染层统一入口：始终返回一个完整的 ArchiveContent，未提供或字段为空白时
 * 回退到 defaultArchiveContent。这样模板中可以直接 `{{ archiveContent.xxx }}`
 * 而不必处理 undefined / 空字符串。
 */
export function resolveLobbyArchiveContent(
  setting: LobbySetting | null | undefined,
): ArchiveContent {
  if (!setting) {
    return { ...defaultArchiveContent }
  }
  return normalizeArchiveContent(setting.archive_content)
}
