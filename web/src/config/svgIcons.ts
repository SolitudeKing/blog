export const svgIconNames = [
  'about-field',
  'archive-wave',
  'arrow-left',
  'arrow-right',
  'arrow-up-right',
  'article-cover',
  'article',
  'book-open',
  'brand-waves',
  'chevron-down',
  'chevron-up',
  'close',
  'dashboard',
  'document',
  'document-lines',
  'empty-article',
  'empty-image',
  'empty-inbox',
  'empty-notice',
  'external-link',
  'info',
  'link',
  'list',
  'logout',
  'media',
  'moon',
  'not-found-field',
  'plus',
  'rss',
  'search',
  'search-minus',
  'settings',
  'sidebar',
  'sidebar-collapse',
  'sidebar-expand',
  'sun',
  'topic-grid',
] as const

export type SvgIconName = (typeof svgIconNames)[number]

export interface SvgIconDefinition {
  viewBox: string
  content: string
  fill: string
  stroke: string
  strokeLinecap?: 'butt' | 'round' | 'square' | 'inherit'
  strokeLinejoin?: 'round' | 'inherit' | 'miter' | 'bevel'
  strokeWidth?: string
}

const rawIcons = import.meta.glob<string>('../assets/icons/*.svg', {
  eager: true,
  import: 'default',
  query: '?raw',
})

function readAttribute(attributes: string, name: string) {
  return attributes.match(new RegExp(`\\b${name}="([^"]+)"`, 'i'))?.[1]
}

function parseIcon(name: SvgIconName): SvgIconDefinition {
  const source = rawIcons[`../assets/icons/${name}.svg`]
  const match = source?.match(/^<svg\b([^>]*)>([\s\S]*?)<\/svg>\s*$/i)
  if (!match) {
    throw new Error(`无法加载 SVG 图标：${name}`)
  }
  const attributes = match[1]
  return {
    viewBox: readAttribute(attributes, 'viewBox') ?? '0 0 24 24',
    content: match[2].trim(),
    fill: readAttribute(attributes, 'fill') ?? 'none',
    stroke: readAttribute(attributes, 'stroke') ?? 'currentColor',
    strokeLinecap: readAttribute(attributes, 'stroke-linecap') as SvgIconDefinition['strokeLinecap'],
    strokeLinejoin: readAttribute(attributes, 'stroke-linejoin') as SvgIconDefinition['strokeLinejoin'],
    strokeWidth: readAttribute(attributes, 'stroke-width'),
  }
}

export const svgIcons = Object.fromEntries(
  svgIconNames.map((name) => [name, parseIcon(name)]),
) as Record<SvgIconName, SvgIconDefinition>
