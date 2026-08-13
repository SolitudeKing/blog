<template>
  <section class="about-page" aria-labelledby="about-title">
    <header class="about-hero">
      <div class="about-hero__copy">
        <p class="about-kicker">About the keeper</p>
        <div id="about-title" role="heading" aria-level="1">你好，我是 {{ author }}</div>
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
          <SvgIcon class="about-portrait__field" name="about-field" />
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
          <div id="principles-title" role="heading" aria-level="2">让内容按自己的节奏生长</div>
          <p>这些原则约束这个博客的设计与维护方式，也帮助阅读始终停留在内容本身。</p>
        </header>

        <ol class="about-principle-list">
          <li v-for="(principle, index) in principles" :key="principle.title" class="about-principle">
            <span class="about-principle__index" aria-hidden="true">{{ formatIndex(index) }}</span>
            <div>
              <div role="heading" aria-level="3">{{ principle.title }}</div>
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
          <div id="contact-title" role="heading" aria-level="2">
            {{ socialEntries.length ? '在这些地方找到我' : '联系方式暂时停泊' }}
          </div>
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
            <SvgIcon name="arrow-up-right" />
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
import SvgIcon from '@/components/base/SvgIcon.vue'
import { useSettingStore } from '@/stores/setting'
import { createSocialLinkEntries } from '@/utils/socialLinks'

interface Principle {
  title: string
  description: string
}

const portraitCaptionId = 'about-portrait-caption'
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
const socialEntries = computed(() => createSocialLinkEntries(setting.lobby?.social_links))

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
