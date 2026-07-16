export interface SocialLinkEntry {
  key: string
  label: string
  href: string
  external: boolean
}

const socialLabelMap: Readonly<Record<string, string>> = {
  github: 'GitHub',
  gitee: 'Gitee',
  bilibili: 'Bilibili',
  douyin: '抖音',
  email: '电子邮件',
  mail: '电子邮件',
  rss: 'RSS',
}

/** 公开链接只接受明确的 Web 与邮件协议，拒绝 javascript: 等可执行地址。 */
export function normalizeSocialUrl(value: unknown) {
  if (typeof value !== 'string') {
    return ''
  }
  const candidate = value.trim()
  if (!candidate) {
    return ''
  }

  try {
    const url = new URL(candidate)
    return ['http:', 'https:', 'mailto:'].includes(url.protocol) ? url.toString() : ''
  } catch {
    return ''
  }
}

export function createSocialLinkEntries(
  links: Readonly<Record<string, string>> | null | undefined,
): SocialLinkEntry[] {
  return Object.entries(links ?? {}).flatMap(([key, value]) => {
    const href = normalizeSocialUrl(value)
    if (!href) {
      return []
    }
    const normalizedKey = key.trim().toLowerCase()
    return [
      {
        key,
        label: socialLabelMap[normalizedKey] ?? formatSocialLabel(key),
        href,
        external: href.startsWith('http://') || href.startsWith('https://'),
      },
    ]
  })
}

function formatSocialLabel(key: string) {
  return key
    .trim()
    .replace(/[-_]+/g, ' ')
    .replace(/\b\w/g, (character) => character.toUpperCase())
}
