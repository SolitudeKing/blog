<template>
  <section class="about-page" aria-labelledby="about-title">
    <header class="about-hero">
      <div class="about-hero__copy">
        <p class="about-kicker">About the keeper</p>
        <h1 id="about-title">你好，我是 {{ author }}</h1>
        <p class="about-hero__lead">{{ essay }}</p>
        <p class="about-hero__signature">{{ siteName }} 的维护者</p>
        <div class="about-hero__actions">
          <a class="mist-button mist-button--lg" href="#contact" @click="focusContact">和我联系</a>
          <RouterLink class="mist-button mist-button--secondary mist-button--lg" to="/archives">
            阅读文章
          </RouterLink>
        </div>
      </div>

      <div class="about-hero__visual">
        <figure class="about-portrait mist-luminous" :aria-labelledby="portraitCaptionId">
          <svg class="about-portrait__field" viewBox="0 0 620 720" fill="none" stroke="currentColor" aria-hidden="true">
            <path
              d="M-40 292c88-75 132 70 220-5s132-55 220 4 132 74 220-8 132-32 220-5"
              stroke-width="1.5"
              opacity=".76"
            />
            <path
              d="M-40 332c88-52 132 50 220-2s132-35 220 3 132 46 220-5 132-22 220-3"
              stroke-width="1"
              opacity=".42"
            />
            <path
              d="M-40 374c88-32 132 31 220-1s132-19 220 2 132 28 220-4 132-12 220-2"
              stroke-width=".8"
              opacity=".26"
            />
            <path d="M42 92c74 30 82 94 40 150s-37 113 14 162 54 119 2 188" stroke-width="1" opacity=".22" />
            <path d="M516 72c-70 51-78 117-28 166s43 116-12 164-56 112-14 176" stroke-width="1" opacity=".22" />
            <circle cx="310" cy="326" r="196" stroke-width="1" opacity=".15" />
            <circle cx="310" cy="326" r="148" stroke-width="1" opacity=".22" />
          </svg>
          <span class="about-portrait__monogram" aria-hidden="true">{{ authorInitial }}</span>
          <figcaption :id="portraitCaptionId" class="about-portrait__caption">
            <span>{{ author }}</span>
            <span>{{ siteName }}</span>
          </figcaption>
        </figure>
      </div>
    </header>

    <section class="about-section" aria-labelledby="principles-title">
      <div class="about-principles">
        <header class="about-principles__intro">
          <p class="about-kicker">Publishing principles</p>
          <h2 id="principles-title">让内容按自己的节奏生长</h2>
          <p>这些原则约束这个博客的设计与维护方式，也帮助阅读始终停留在内容本身。</p>
        </header>

        <ol class="about-principle-list">
          <li v-for="(principle, index) in principles" :key="principle.title" class="about-principle">
            <span class="about-principle__index" aria-hidden="true">{{ formatIndex(index) }}</span>
            <div>
              <h3>{{ principle.title }}</h3>
              <p>{{ principle.description }}</p>
            </div>
          </li>
        </ol>
      </div>
    </section>

    <section
      id="contact"
      class="about-section about-contact-section"
      aria-labelledby="contact-title"
      tabindex="-1"
    >
      <div class="about-contact mist-glass--strong">
        <div>
          <p class="about-kicker">Say hello</p>
          <h2 id="contact-title">
            {{ socialEntries.length ? '在这些地方找到我' : '联系方式暂时停泊' }}
          </h2>
          <p>
            {{
              socialEntries.length
                ? '选择你习惯的平台，继续聊写作、技术或长期维护。'
                : '站点尚未公开社交链接，你仍可以从归档继续阅读。'
            }}
          </p>
        </div>

        <nav v-if="socialEntries.length" class="about-contact__links" aria-label="作者社交链接">
          <a
            v-for="entry in socialEntries"
            :key="entry.key"
            :href="entry.href"
            :target="entry.external ? '_blank' : undefined"
            :rel="entry.external ? 'noopener noreferrer' : undefined"
          >
            <span>{{ entry.label }}</span>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
              <path d="M7 17 17 7M8 7h9v9" />
            </svg>
          </a>
        </nav>

        <RouterLink v-else class="mist-button mist-button--secondary mist-button--lg" to="/archives">
          先读一篇文章
        </RouterLink>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useSettingStore } from '@/stores/setting'

interface Principle {
  title: string
  description: string
}

interface SocialEntry {
  key: string
  label: string
  href: string
  external: boolean
}

const portraitCaptionId = 'about-portrait-caption'
const socialLabels: Record<string, string> = {
  github: 'GitHub',
  gitee: 'Gitee',
  bilibili: 'Bilibili',
  douyin: '抖音',
  email: '电子邮件',
  mail: '电子邮件',
  rss: 'RSS',
}
const principles: Principle[] = [
  {
    title: '内容先于装饰',
    description: '排版、留白与动效都服务于理解；移除背景效果之后，文章仍然应该清楚而完整。',
  },
  {
    title: '让系统承担复杂',
    description: '主题、组件和状态保持稳定契约，把维护成本留在系统内部，而不是交给每一篇内容。',
  },
  {
    title: '为长期阅读留白',
    description: '不追逐每一次短暂变化，让归档、链接与文字在更长时间里仍然可以被重新找到。',
  },
]

const setting = useSettingStore()
const siteName = computed(() => setting.lobby?.site_name?.trim() || 'Solitude Blog')
const author = computed(() => setting.lobby?.author?.trim() || 'Solitude King')
const essay = computed(
  () => setting.lobby?.essay?.trim() || '关于写作、技术与长期维护的个人记录。',
)
const authorInitial = computed(() => Array.from(author.value)[0]?.toUpperCase() || 'S')
const socialEntries = computed<SocialEntry[]>(() => {
  const links = setting.lobby?.social_links ?? {}
  return Object.entries(links).flatMap(([key, value]) => {
    const href = normalizeSocialUrl(value)
    if (!href) {
      return []
    }
    const normalizedKey = key.trim().toLowerCase()
    return [
      {
        key,
        label: socialLabels[normalizedKey] ?? formatSocialLabel(key),
        href,
        external: href.startsWith('http://') || href.startsWith('https://'),
      },
    ]
  })
})

function normalizeSocialUrl(value: string) {
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

function formatSocialLabel(key: string) {
  return key
    .trim()
    .replace(/[-_]+/g, ' ')
    .replace(/\b\w/g, (character) => character.toUpperCase())
}

function formatIndex(index: number) {
  return String(index + 1).padStart(2, '0')
}

function focusContact(event: MouseEvent) {
  event.preventDefault()
  const target = document.getElementById('contact')
  if (!target) {
    return
  }
  const url = new URL(window.location.href)
  url.hash = 'contact'
  window.history.replaceState(window.history.state, '', url)
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  target.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
  window.setTimeout(() => target.focus({ preventScroll: true }), reduceMotion ? 0 : 360)
}
</script>
