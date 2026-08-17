/**
 * Hero / section 标题文案排版工具，用于 archive / search / about 板块
 * 的标题 `<div role="heading">` 元素。
 *
 * 排版约定：
 * - 仅含句末标点（. 。 ; ；）或纯字符的文案，例如 "打捞一段想法"、
 *   "让内容按自己的节奏生长"，保持单行展示，文本不换行。
 * - 含分隔符的文案从首个分隔符处切为两行；分隔符本身不出现在
 *   渲染文本中，仅作为切分标记使用。
 *
 * 分隔符集合：
 * - 单字符可见分隔符 `，` `,` `、`（1 个 code unit）—— 保留在 first
 *   末尾，与剩余内容形成"逗号收尾 + 第二行"的两行错落。
 * - 字面 `\n`（反斜杠 + n，2 个 code unit）—— admin 在单行 `<input>`
 *   里可输入的"两行标记"。浏览器对单行 input 会丢弃真正的换行符，
 *   因此这里用字面两字符序列作为可在表单里直接键入的标记；该标记仅
 *   用于切分，slice 时不放入任一段，且 first / second 中残留的 `\n`
 *   一律剥离，不作为可见字符、也不产生额外换行。
 *
 * 切分失败（无可拆分隔符或分隔符之后无内容）时返回 null，模板走单行
 * 样式即可，避免空第二行或重复排版。
 *
 * 注意：空格 *不* 作为分隔符，避免半角空格意外触发换行。
 */
export interface HeroTitleLines {
  first: string
  second: string
}

const HERO_TITLE_BREAK_PATTERN = /[，,、]|\\n/

export function splitHeroTitle(text: string | null | undefined): HeroTitleLines | null {
  if (!text) {
    return null
  }
  const trimmed = text.trim()
  if (!trimmed) {
    return null
  }
  const match = HERO_TITLE_BREAK_PATTERN.exec(trimmed)
  if (!match || match.index < 1) {
    return null
  }
  const breakIndex = match.index
  const separator = match[0]
  const separatorLen = separator.length
  // 字面 "\n"（反斜杠 + n）是 admin 在单行 input 里可输入的"两行标记"，
  // 仅作为切分点：slice 时不放入任一段。其余 visible 分隔符（， , 、）
  // 仍是 1 个 code unit，保留在 first 末尾。first / second 中残留的
  // 字面 "\n" 一律剥离，文本里看不到任何标记。
  const rawFirst =
    separator === '\\n' ? trimmed.slice(0, breakIndex) : trimmed.slice(0, breakIndex + separatorLen)
  const first = rawFirst.replace(/\\n/g, '')
  const second = trimmed.slice(breakIndex + separatorLen).replace(/\\n/g, '').trim()
  if (!second) {
    return null
  }
  return { first, second }
}