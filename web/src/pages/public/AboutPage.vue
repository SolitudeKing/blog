<template>
  <section class="about-page" aria-labelledby="about-title">
    <header class="about-hero">
      <div class="about-hero__copy">
        <p class="about-kicker">{{ aboutContent.about_kicker }}</p>
        <div
          id="about-title"
          role="heading"
          aria-level="1"
          :title="aboutContent.about_heading"
        >
          <template v-if="aboutTitleLines">
            <span>{{ aboutTitleLines.first }}</span>
            <span class="about-title__line">{{ aboutTitleLines.second }}</span>
          </template>
          <template v-else>{{ aboutContent.about_heading }}</template>
        </div>
        <p class="about-hero__lead">{{ aboutContent.about_lead }}</p>
        <p class="about-hero__signature">{{ aboutContent.about_signature }}</p>
        <div class="about-hero__actions">
          <a class="mist-button mist-button--lg" href="#contact" @click="focusContact">
            {{ aboutContent.about_contact_label }}
          </a>
          <RouterLink class="mist-button mist-button--secondary mist-button--lg" to="/archives">
            {{ aboutContent.about_reading_label }}
          </RouterLink>
        </div>
      </div>

      <div class="about-hero__visual">
        <figure class="about-portrait mist-luminous" :aria-labelledby="portraitCaptionId">
          <SvgIcon class="about-portrait__field" name="about-field" />
          <span class="about-portrait__monogram" aria-hidden="true">{{ authorInitial }}</span>
          <figcaption :id="portraitCaptionId" class="about-portrait__caption">
            <span>{{ aboutContent.about_portrait_line1 }}</span>
            <span>{{ aboutContent.about_portrait_line2 }}</span>
          </figcaption>
        </figure>
      </div>
    </header>

    <section class="about-section" aria-labelledby="principles-title">
      <div class="about-principles">
        <header class="about-principles__intro">
          <p class="about-kicker">{{ aboutContent.about_principles_kicker }}</p>
          <div
            id="principles-title"
            role="heading"
            aria-level="2"
            :title="aboutContent.about_principles_heading"
          >
            <template v-if="aboutPrinciplesTitleLines">
              <span>{{ aboutPrinciplesTitleLines.first }}</span>
              <span class="about-principles-title__line">{{ aboutPrinciplesTitleLines.second }}</span>
            </template>
            <template v-else>{{ aboutContent.about_principles_heading }}</template>
          </div>
          <p>{{ aboutContent.about_principles_intro }}</p>
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
          <p class="about-kicker">{{ aboutContent.about_contact_kicker }}</p>
          <div
            id="contact-title"
            role="heading"
            aria-level="2"
            :title="aboutContactHeading"
          >
            <template v-if="aboutContactTitleLines">
              <span>{{ aboutContactTitleLines.first }}</span>
              <span class="about-contact-title__line">{{ aboutContactTitleLines.second }}</span>
            </template>
            <template v-else>{{ aboutContactHeading }}</template>
          </div>
          <p>
            {{
              socialEntries.length
                ? aboutContent.about_contact_intro_with
                : aboutContent.about_contact_intro_empty
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
          {{ aboutContent.about_contact_empty_cta }}
        </RouterLink>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { resolveAboutPrinciples, resolveLobbyAboutContent } from '@/config/aboutContent'
import { useSettingStore } from '@/stores/setting'
import { splitHeroTitle } from '@/utils/heroTitle'
import { createSocialLinkEntries } from '@/utils/socialLinks'

const portraitCaptionId = 'about-portrait-caption'

const setting = useSettingStore()
const aboutContent = computed(() => resolveLobbyAboutContent(setting.lobby))
const principles = computed(() => resolveAboutPrinciples(aboutContent.value))
const aboutTitleLines = computed(() => splitHeroTitle(aboutContent.value.about_heading))
const aboutPrinciplesTitleLines = computed(() =>
  splitHeroTitle(aboutContent.value.about_principles_heading),
)
const aboutContactHeading = computed(() =>
  socialEntries.value.length
    ? aboutContent.value.about_contact_heading_with
    : aboutContent.value.about_contact_heading_empty,
)
const aboutContactTitleLines = computed(() => splitHeroTitle(aboutContactHeading.value))
const authorInitial = computed(() => {
  const author = setting.lobby?.author?.trim() ?? ''
  return author.charAt(0).toUpperCase() || 'S'
})
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
