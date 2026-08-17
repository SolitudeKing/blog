import type { AboutContent, LobbySetting } from '@/types/setting'

/**
 * 关于页文案的字段目录。
 *
 * 与 home_content 同构：聚合在 `site_settings.about_content` 一列 JSON 中，
 * 不分主题，整组替换。后台 Settings 页的「关于文案」段直接遍历
 * `aboutContentFields` 渲染；新增字段只动本文件 + 后端 `sectioncontent`
 * 包的常量。
 *
 * 关于页 hero 使用本组的独立文案（about_heading / about_lead /
 * about_signature / about_portrait_line1/2），默认值刻意不复述首页的
 * 问候语、essay、作者名与站点名，避免两个板块内容冗余。
 */
export type AboutContentKey = keyof AboutContent

export interface AboutContentField {
  key: AboutContentKey
  label: string
  hint: string
  max: number
  multiline?: boolean
}

// 与后端 sectioncontent/about_content.go 的 *MaxRunes 常量保持一一对应。
export const aboutContentFields: ReadonlyArray<AboutContentField> = [
  {
    key: 'about_kicker',
    label: '关于小标',
    hint: 'about-hero 标题上方的小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'about_heading',
    label: '关于标题',
    hint: 'about-hero 的大标题；最多 80 字。',
    max: 80,
  },
  {
    key: 'about_lead',
    label: '关于导语',
    hint: 'about-hero 标题下的多行描述；最多 240 字。',
    max: 240,
    multiline: true,
  },
  {
    key: 'about_signature',
    label: '关于签名行',
    hint: '导语下的签名短语；最多 80 字。',
    max: 80,
  },
  {
    key: 'about_contact_label',
    label: '"和我联系"按钮',
    hint: 'about-hero 左侧主操作按钮；最多 16 字。',
    max: 16,
  },
  {
    key: 'about_reading_label',
    label: '"阅读文章"按钮',
    hint: 'about-hero 右侧副操作按钮；最多 16 字。',
    max: 16,
  },
  {
    key: 'about_portrait_line1',
    label: '肖像题注第一行',
    hint: 'about-portrait 下方题注的第一行；最多 40 字。',
    max: 40,
  },
  {
    key: 'about_portrait_line2',
    label: '肖像题注第二行',
    hint: 'about-portrait 下方题注的第二行；最多 80 字。',
    max: 80,
  },
  {
    key: 'about_principles_kicker',
    label: '原则小标',
    hint: '"Publishing principles" 段上方小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'about_principles_heading',
    label: '原则标题',
    hint: '"Publishing principles" 段标题；最多 80 字。',
    max: 80,
  },
  {
    key: 'about_principles_intro',
    label: '原则导语',
    hint: '原则标题下的多行描述；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'principle_1_title',
    label: '原则一标题',
    hint: '第一条发布原则的标题；最多 40 字。',
    max: 40,
  },
  {
    key: 'principle_1_description',
    label: '原则一描述',
    hint: '第一条发布原则的描述；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'principle_2_title',
    label: '原则二标题',
    hint: '第二条发布原则的标题；最多 40 字。',
    max: 40,
  },
  {
    key: 'principle_2_description',
    label: '原则二描述',
    hint: '第二条发布原则的描述；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'principle_3_title',
    label: '原则三标题',
    hint: '第三条发布原则的标题；最多 40 字。',
    max: 40,
  },
  {
    key: 'principle_3_description',
    label: '原则三描述',
    hint: '第三条发布原则的描述；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'about_contact_kicker',
    label: '联系小标',
    hint: '"Say hello" 段上方小标；最多 32 字。',
    max: 32,
  },
  {
    key: 'about_contact_heading_with',
    label: '联系标题（有链接）',
    hint: '已配置社交链接时显示的标题；最多 80 字。',
    max: 80,
  },
  {
    key: 'about_contact_heading_empty',
    label: '联系标题（无链接）',
    hint: '未配置社交链接时显示的标题；最多 80 字。',
    max: 80,
  },
  {
    key: 'about_contact_intro_with',
    label: '联系导语（有链接）',
    hint: '已配置社交链接时显示的描述；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'about_contact_intro_empty',
    label: '联系导语（无链接）',
    hint: '未配置社交链接时显示的描述；最多 160 字。',
    max: 160,
    multiline: true,
  },
  {
    key: 'about_contact_empty_cta',
    label: '"先读一篇文章"按钮',
    hint: '未配置社交链接时的兜底跳转按钮；最多 16 字。',
    max: 16,
  },
]

export const defaultAboutContent: Readonly<AboutContent> = {
  about_kicker: 'About the keeper',
  about_heading: '关于我，也关于这座博客',
  about_lead:
    '我是这座博客的维护者，也是一名长期主义的记录者。比起追逐热点，更在意那些经得起时间检验的工程实践与真实判断。',
  about_signature: '记录与维护，皆在字里行间',
  about_contact_label: '和我联系',
  about_reading_label: '阅读文章',
  about_principles_kicker: 'Publishing principles',
  about_principles_heading: '让内容按自己的节奏生长',
  about_principles_intro:
    '这些原则约束这个博客的设计与维护方式，也帮助阅读始终停留在内容本身。',
  principle_1_title: '内容先于装饰',
  principle_1_description:
    '排版、留白与动效都服务于理解；移除背景效果之后，文章仍然应该清楚而完整。',
  principle_2_title: '让系统承担复杂',
  principle_2_description:
    '主题、组件和状态保持稳定契约，把维护成本留在系统内部，而不是交给每一篇内容。',
  principle_3_title: '为长期阅读留白',
  principle_3_description:
    '不追逐每一次短暂变化，让归档、链接与文字在更长时间里仍然可以被重新找到。',
  about_contact_kicker: 'Say hello',
  about_contact_heading_with: '在这些地方找到我',
  about_contact_heading_empty: '联系方式暂时停泊',
  about_contact_intro_with: '选择你习惯的平台，继续聊写作、技术或长期维护。',
  about_contact_intro_empty: '站点尚未公开社交链接，你仍可以从归档继续阅读。',
  about_contact_empty_cta: '先读一篇文章',
  about_portrait_line1: 'Blog keeper',
  about_portrait_line2: '记录，是为了更好地想起',
}

export function createDefaultAboutContent(): AboutContent {
  return { ...defaultAboutContent }
}

type AboutContentInput = Partial<AboutContent> | null | undefined

function resolveText(value: unknown, fallback: string) {
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

export function normalizeAboutContent(value?: AboutContentInput): AboutContent {
  return {
    about_kicker: resolveText(value?.about_kicker, defaultAboutContent.about_kicker),
    about_heading: resolveText(value?.about_heading, defaultAboutContent.about_heading),
    about_lead: resolveText(value?.about_lead, defaultAboutContent.about_lead),
    about_signature: resolveText(value?.about_signature, defaultAboutContent.about_signature),
    about_contact_label: resolveText(
      value?.about_contact_label,
      defaultAboutContent.about_contact_label,
    ),
    about_reading_label: resolveText(
      value?.about_reading_label,
      defaultAboutContent.about_reading_label,
    ),
    about_principles_kicker: resolveText(
      value?.about_principles_kicker,
      defaultAboutContent.about_principles_kicker,
    ),
    about_principles_heading: resolveText(
      value?.about_principles_heading,
      defaultAboutContent.about_principles_heading,
    ),
    about_principles_intro: resolveText(
      value?.about_principles_intro,
      defaultAboutContent.about_principles_intro,
    ),
    principle_1_title: resolveText(value?.principle_1_title, defaultAboutContent.principle_1_title),
    principle_1_description: resolveText(
      value?.principle_1_description,
      defaultAboutContent.principle_1_description,
    ),
    principle_2_title: resolveText(value?.principle_2_title, defaultAboutContent.principle_2_title),
    principle_2_description: resolveText(
      value?.principle_2_description,
      defaultAboutContent.principle_2_description,
    ),
    principle_3_title: resolveText(value?.principle_3_title, defaultAboutContent.principle_3_title),
    principle_3_description: resolveText(
      value?.principle_3_description,
      defaultAboutContent.principle_3_description,
    ),
    about_contact_kicker: resolveText(
      value?.about_contact_kicker,
      defaultAboutContent.about_contact_kicker,
    ),
    about_contact_heading_with: resolveText(
      value?.about_contact_heading_with,
      defaultAboutContent.about_contact_heading_with,
    ),
    about_contact_heading_empty: resolveText(
      value?.about_contact_heading_empty,
      defaultAboutContent.about_contact_heading_empty,
    ),
    about_contact_intro_with: resolveText(
      value?.about_contact_intro_with,
      defaultAboutContent.about_contact_intro_with,
    ),
    about_contact_intro_empty: resolveText(
      value?.about_contact_intro_empty,
      defaultAboutContent.about_contact_intro_empty,
    ),
    about_contact_empty_cta: resolveText(
      value?.about_contact_empty_cta,
      defaultAboutContent.about_contact_empty_cta,
    ),
    about_portrait_line1: resolveText(
      value?.about_portrait_line1,
      defaultAboutContent.about_portrait_line1,
    ),
    about_portrait_line2: resolveText(
      value?.about_portrait_line2,
      defaultAboutContent.about_portrait_line2,
    ),
  }
}

/**
 * 渲染层统一入口：始终返回一个完整的 AboutContent，未提供或字段为空白时
 * 回退到 defaultAboutContent。这样模板中可以直接 `{{ aboutContent.xxx }}`
 * 而不必处理 undefined / 空字符串。
 */
export function resolveLobbyAboutContent(setting: LobbySetting | null | undefined): AboutContent {
  if (!setting) {
    return { ...defaultAboutContent }
  }
  return normalizeAboutContent(setting.about_content)
}

/**
 * 把 3 组原则字段折叠成渲染用数组；AboutPage 的 `v-for` 直接遍历它。
 */
export function resolveAboutPrinciples(content: AboutContent): Array<{
  title: string
  description: string
}> {
  return [
    { title: content.principle_1_title, description: content.principle_1_description },
    { title: content.principle_2_title, description: content.principle_2_description },
    { title: content.principle_3_title, description: content.principle_3_description },
  ]
}
