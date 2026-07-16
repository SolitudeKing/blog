export type TopicCatalogLabel = 'NODES' | 'CODE' | 'JOTTING'

export interface TopicCatalogEntry {
  label: TopicCatalogLabel
  name: string
  slug: string
  description: string
}

/**
 * 首页专题目录的唯一来源。
 *
 * 后台仍可维护其他专题，但公开首页只按这里的顺序展示约定的三个入口，
 * 避免测试数据或后台排序变化破坏站点信息架构。
 */
export const topicCatalog = [
  {
    label: 'NODES',
    name: '雾里拾笺',
    slug: 'nodes',
    description: '收拢阅读、学习与技术实践中散落的知识微光。',
  },
  {
    label: 'CODE',
    name: '微光造物',
    slug: 'code',
    description: '记录灵感如何经由设计、代码与实验长成作品。',
  },
  {
    label: 'JOTTING',
    name: '风过留痕',
    slug: 'jotting',
    description: '安放日常见闻、片刻心绪与未成体系的思考。',
  },
] as const satisfies ReadonlyArray<TopicCatalogEntry>

export function findTopicBySlug(value: unknown): TopicCatalogEntry | null {
  if (typeof value !== 'string') {
    return null
  }
  const slug = value.trim().toLowerCase()
  return topicCatalog.find((topic) => topic.slug === slug) ?? null
}

export function normalizeTopicLabel(value: unknown) {
  return typeof value === 'string' ? value.trim().toUpperCase() : ''
}
