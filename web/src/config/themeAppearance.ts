import type {
  LobbySetting,
  ThemeElementMap,
  ThemeElements,
  ThemeName,
} from '@/types/setting'

export interface ThemeAppearanceOption {
  value: ThemeName
  label: string
  description: string
  preview: string
}

export const themeAppearanceOptions = [
  {
    value: 'mist-sea-salt',
    label: '雾境海盐',
    description: '清爽安静的海岸晨雾与透明蓝玻璃。',
    preview: '深海蓝交互色穿过轻盈海雾，适合专注阅读。',
  },
  {
    value: 'mist-forest',
    label: '雾境青森',
    description: '自然平静的森林晨雾与露水冷光。',
    preview: '深森绿交互色配合青绿雾层，适合长时间浏览。',
  },
] as const satisfies ReadonlyArray<ThemeAppearanceOption>

export const defaultThemeElements: Readonly<ThemeElementMap> = {
  'mist-sea-salt': {
    home_latest_empty_description: '第一篇文章正在潮汐之外酝酿。',
    home_latest_end_text: '已经读到潮汐尽头',
  },
  'mist-forest': {
    home_latest_empty_description: '第一篇文章正在林雾之间酝酿。',
    home_latest_end_text: '已经走到林径尽头',
  },
}

export const neutralThemeElements: Readonly<ThemeElements> = {
  home_latest_empty_description: '第一篇文章正在酝酿中。',
  home_latest_end_text: '已经读完全部文章',
}

type ThemeElementMapInput = Partial<
  Record<ThemeName, Partial<ThemeElements> | null | undefined>
>

function resolveText(value: unknown, fallback: string) {
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

export function createDefaultThemeElementMap(): ThemeElementMap {
  return cloneThemeElementMap(defaultThemeElements)
}

export function normalizeThemeElementMap(value?: ThemeElementMapInput | null): ThemeElementMap {
  return {
    'mist-sea-salt': normalizeThemeElements(
      value?.['mist-sea-salt'],
      defaultThemeElements['mist-sea-salt'],
    ),
    'mist-forest': normalizeThemeElements(
      value?.['mist-forest'],
      defaultThemeElements['mist-forest'],
    ),
  }
}

export function cloneThemeElementMap(value: Readonly<ThemeElementMap>): ThemeElementMap {
  return {
    'mist-sea-salt': { ...value['mist-sea-salt'] },
    'mist-forest': { ...value['mist-forest'] },
  }
}

export function resolveLobbyThemeElements(setting: LobbySetting | null): ThemeElements {
  if (!setting) {
    return { ...neutralThemeElements }
  }

  return normalizeThemeElements(
    setting.theme_elements?.[setting.theme],
    defaultThemeElements[setting.theme],
  )
}

function normalizeThemeElements(
  value: Partial<ThemeElements> | null | undefined,
  fallback: Readonly<ThemeElements>,
): ThemeElements {
  return {
    home_latest_empty_description: resolveText(
      value?.home_latest_empty_description,
      fallback.home_latest_empty_description,
    ),
    home_latest_end_text: resolveText(
      value?.home_latest_end_text,
      fallback.home_latest_end_text,
    ),
  }
}
