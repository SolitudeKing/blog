import { computed, reactive } from 'vue'
import type { ModeName, ThemeName } from '@/types/setting'

export type { ModeName, ThemeName } from '@/types/setting'

export interface ThemeState {
  theme: ThemeName
  mode: ModeName
}

export interface ThemeInitOptions {
  syncThemeColor?: boolean
}

const MODE_STORAGE_KEY = 'blog:mode'
const SITE_APPEARANCE_STORAGE_KEY = 'blog:site-appearance'
const LEGACY_STORAGE_KEY = 'blog:theme'
const DEFAULT_THEME: ThemeName = 'mist-sea-salt'
const DEFAULT_MODE: ModeName = 'light'

const state = reactive<ThemeState>({
  theme: DEFAULT_THEME,
  mode: DEFAULT_MODE,
})
const preference = reactive({ hasLocalMode: false })

let initialized = false
let storageListenerAttached = false
let syncThemeColorEnabled = true
let initialDomMode: ModeName | null = null
let serverDefaultMode: ModeName | null = null

function isThemeName(value: unknown): value is ThemeName {
  return value === 'mist-sea-salt' || value === 'mist-forest'
}

function isModeName(value: unknown): value is ModeName {
  return value === 'light' || value === 'dark'
}

function normalizeAppearance(value: unknown): ThemeState | null {
  if (!value || typeof value !== 'object') {
    return null
  }

  const candidate = value as { theme?: unknown; mode?: unknown }
  if (!isThemeName(candidate.theme) || !isModeName(candidate.mode)) {
    return null
  }

  return {
    theme: candidate.theme,
    mode: candidate.mode,
  }
}

function parseJson(raw: string | null): unknown {
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as unknown
  } catch {
    return null
  }
}

function parseStoredMode(raw: string | null): ModeName | null {
  if (!raw) {
    return null
  }
  if (isModeName(raw)) {
    return raw
  }

  const parsed = parseJson(raw)
  if (isModeName(parsed)) {
    return parsed
  }
  if (parsed && typeof parsed === 'object') {
    const candidate = parsed as { mode?: unknown }
    return isModeName(candidate.mode) ? candidate.mode : null
  }

  return null
}

function parseSiteAppearance(raw: string | null): ThemeState | null {
  return normalizeAppearance(parseJson(raw))
}

function writeMode(mode: ModeName) {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.setItem(MODE_STORAGE_KEY, mode)
  } catch {
    // Storage can be blocked without blocking an in-page mode change.
  }
}

function removeMode() {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.removeItem(MODE_STORAGE_KEY)
  } catch {
    // Storage can be blocked without blocking an in-page mode change.
  }
}

function readStoredMode(): ModeName | null {
  if (typeof window === 'undefined') {
    return null
  }

  try {
    const storedMode = parseStoredMode(window.localStorage.getItem(MODE_STORAGE_KEY))
    if (storedMode) {
      window.localStorage.removeItem(LEGACY_STORAGE_KEY)
      return storedMode
    }

    window.localStorage.removeItem(MODE_STORAGE_KEY)
    const legacyMode = parseStoredMode(window.localStorage.getItem(LEGACY_STORAGE_KEY))
    window.localStorage.removeItem(LEGACY_STORAGE_KEY)
    if (legacyMode) {
      window.localStorage.setItem(MODE_STORAGE_KEY, legacyMode)
    }
    return legacyMode
  } catch {
    return null
  }
}

function readCachedSiteAppearance(): ThemeState | null {
  if (typeof window === 'undefined') {
    return null
  }

  try {
    const appearance = parseSiteAppearance(
      window.localStorage.getItem(SITE_APPEARANCE_STORAGE_KEY),
    )
    if (!appearance) {
      window.localStorage.removeItem(SITE_APPEARANCE_STORAGE_KEY)
    }
    return appearance
  } catch {
    return null
  }
}

function writeCachedSiteAppearance(value: ThemeState) {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.setItem(SITE_APPEARANCE_STORAGE_KEY, JSON.stringify(value))
  } catch {
    // The server value still applies to the current document when storage is blocked.
  }
}

function readDomTheme(): ThemeName | null {
  if (typeof document === 'undefined') {
    return null
  }

  const { theme } = document.documentElement.dataset
  return isThemeName(theme) ? theme : null
}

function readDomMode(): ModeName | null {
  if (typeof document === 'undefined') {
    return null
  }

  const { mode } = document.documentElement.dataset
  return isModeName(mode) ? mode : null
}

function readSystemMode(): ModeName {
  return typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : DEFAULT_MODE
}

function syncThemeColor() {
  if (
    !syncThemeColorEnabled ||
    typeof document === 'undefined' ||
    typeof window === 'undefined'
  ) {
    return
  }

  const root = document.documentElement
  let meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'theme-color'
    document.head.append(meta)
  }

  const update = () => {
    const background = window.getComputedStyle(root).getPropertyValue('--bg-primary').trim()
    if (background) {
      meta.content = background
    }
  }

  update()
  window.requestAnimationFrame?.(update)
}

function applyToDom(value: ThemeState) {
  if (typeof document === 'undefined') {
    return
  }

  const root = document.documentElement
  root.dataset.theme = value.theme
  root.dataset.mode = value.mode
  root.style.colorScheme = value.mode
  syncThemeColor()
}

function emitAppearanceChange(value: ThemeState) {
  if (typeof document === 'undefined') {
    return
  }

  document.dispatchEvent(
    new CustomEvent<ThemeState>('mistappearancechange', {
      detail: { ...value },
    }),
  )
}

function commit(value: ThemeState, options: { emit?: boolean } = {}) {
  const changed = state.theme !== value.theme || state.mode !== value.mode
  state.theme = value.theme
  state.mode = value.mode
  applyToDom(value)

  if (changed && options.emit !== false) {
    emitAppearanceChange(value)
  }
}

function fallbackMode() {
  return serverDefaultMode ?? initialDomMode ?? readSystemMode()
}

function applyCachedAppearance(value: ThemeState) {
  serverDefaultMode = value.mode
  commit({
    theme: value.theme,
    mode: preference.hasLocalMode ? state.mode : value.mode,
  })
}

function onStorage(event: StorageEvent) {
  if (event.key === MODE_STORAGE_KEY) {
    const storedMode = parseStoredMode(event.newValue)
    if (storedMode) {
      preference.hasLocalMode = true
      commit({ theme: state.theme, mode: storedMode })
      return
    }
    if (event.newValue === null) {
      preference.hasLocalMode = false
      commit({ theme: state.theme, mode: fallbackMode() })
    }
    return
  }

  if (event.key === SITE_APPEARANCE_STORAGE_KEY) {
    const appearance = parseSiteAppearance(event.newValue)
    if (appearance) {
      applyCachedAppearance(appearance)
    }
    return
  }

  if (event.key === LEGACY_STORAGE_KEY && !preference.hasLocalMode) {
    const legacyMode = parseStoredMode(event.newValue)
    if (legacyMode) {
      preference.hasLocalMode = true
      writeMode(legacyMode)
      commit({ theme: state.theme, mode: legacyMode })
    }
    return
  }

  if (event.key === null) {
    preference.hasLocalMode = false
    commit({ theme: state.theme, mode: fallbackMode() })
  }
}

export function initTheme(options: ThemeInitOptions = {}) {
  if (options.syncThemeColor !== undefined) {
    syncThemeColorEnabled = options.syncThemeColor
  }
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return { ...state }
  }
  if (initialized) {
    applyToDom(state)
    return { ...state }
  }

  initialized = true
  const storedMode = readStoredMode()
  const cachedAppearance = readCachedSiteAppearance()
  const domMode = readDomMode()

  preference.hasLocalMode = storedMode !== null
  initialDomMode = storedMode ? null : domMode
  serverDefaultMode = cachedAppearance?.mode ?? null

  commit(
    {
      theme: cachedAppearance?.theme ?? readDomTheme() ?? DEFAULT_THEME,
      mode: storedMode ?? cachedAppearance?.mode ?? domMode ?? readSystemMode(),
    },
    { emit: false },
  )

  if (!storageListenerAttached) {
    window.addEventListener('storage', onStorage)
    storageListenerAttached = true
  }

  return { ...state }
}

function ensureInitialized() {
  if (!initialized) {
    initTheme()
  }
}

function setMode(mode: ModeName) {
  if (!isModeName(mode)) {
    return
  }

  ensureInitialized()
  preference.hasLocalMode = true
  writeMode(mode)
  commit({ theme: state.theme, mode })
}

function clearModePreference() {
  ensureInitialized()
  preference.hasLocalMode = false
  removeMode()
  commit({ theme: state.theme, mode: fallbackMode() })
}

function syncFromServer(payload: { theme: unknown; mode: unknown }) {
  const normalized = normalizeAppearance(payload)
  if (!normalized) {
    return false
  }

  ensureInitialized()
  serverDefaultMode = normalized.mode
  writeCachedSiteAppearance(normalized)
  commit({
    theme: normalized.theme,
    mode: preference.hasLocalMode ? state.mode : normalized.mode,
  })
  return true
}

export function useTheme() {
  const hasLocalModePreference = computed(() => preference.hasLocalMode)

  return {
    theme: computed(() => state.theme),
    mode: computed(() => state.mode),
    hasLocalModePreference,
    hasLocalPreference: hasLocalModePreference,
    setMode,
    cycleMode() {
      setMode(state.mode === 'light' ? 'dark' : 'light')
    },
    clearModePreference,
    clearPreference: clearModePreference,
    syncFromServer,
  }
}
