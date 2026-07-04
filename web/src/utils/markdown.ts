import MarkdownIt from 'markdown-it'

export interface TocItem {
  id: string
  level: number
  text: string
}

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
})

export function renderMarkdown(source: string) {
  const toc: TocItem[] = []
  const tokens = markdown.parse(source, {})

  for (let index = 0; index < tokens.length; index += 1) {
    const token = tokens[index]
    if (token.type !== 'heading_open') {
      continue
    }
    const inline = tokens[index + 1]
    if (!inline || inline.type !== 'inline') {
      continue
    }
    const level = Number(token.tag.replace('h', ''))
    const text = inline.content
    const id = uniqueHeadingId(slugify(text), toc)
    token.attrSet('id', id)
    toc.push({ id, level, text })
  }

  return {
    html: markdown.renderer.render(tokens, markdown.options, {}),
    toc,
  }
}

function slugify(value: string) {
  const slug = value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-|-$/g, '')

  return slug || 'section'
}

function uniqueHeadingId(base: string, toc: TocItem[]) {
  const used = new Set(toc.map((item) => item.id))
  if (!used.has(base)) {
    return base
  }
  let index = 2
  while (used.has(`${base}-${index}`)) {
    index += 1
  }
  return `${base}-${index}`
}
