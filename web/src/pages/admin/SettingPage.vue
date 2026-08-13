<template>
  <section class="admin-page" :aria-busy="loading || saving">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Settings</p>
        <div role="heading" aria-level="1">站点设置</div>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadSetting">刷新</BaseButton>
    </header>

      <form class="settings-form" :aria-busy="saving" @submit.prevent="saveSetting">
        <div class="settings-form__main">
          <section class="settings-panel">
            <div role="heading" aria-level="2">基础信息</div>
            <BaseInput v-model="form.site_name" label="站点名称" />
            <BaseInput v-model="form.author" label="作者" />
            <BaseInput
              v-model="form.author_avatar_url"
              type="url"
              label="作者头像 URL"
              hint="可填写 /uploads/... 相对路径或 https://... 绝对地址；留空时使用默认头像。"
              :maxlength="500"
            />
            <BaseTextarea v-model="form.essay" label="站点签名" :rows="4" />
            <BaseInput
              v-model="form.icp_number"
              label="网站备案号"
              hint="可选；保存后将在前台页脚展示，并链接至工信部备案管理系统。"
              :maxlength="64"
            />
          </section>

          <section class="settings-panel">
            <div role="heading" aria-level="2">社交链接</div>
            <BaseInput
              v-for="item in socialItems"
              :key="item.key"
              v-model="form.social_links[item.key]"
              :label="item.label"
            />
          </section>
        </div>

        <section class="settings-panel settings-panel--appearance">
        <div role="heading" aria-level="2">主题外观</div>
        <p id="site-theme-hint" class="settings-panel__description">
          主题色由后台统一发布；访客只能在前台切换明暗模式。
        </p>

        <fieldset class="settings-theme-picker" aria-describedby="site-theme-hint">
          <legend class="mist-field__label">站点主题色</legend>
          <div class="settings-theme-picker__options">
            <label
              v-for="option in themeAppearanceOptions"
              :key="option.value"
              class="settings-theme-option"
              :class="{ 'is-selected': form.theme === option.value }"
              :data-theme="option.value"
              :data-mode="form.mode"
            >
              <input
                v-model="form.theme"
                class="settings-theme-option__input"
                type="radio"
                name="site-theme"
                :value="option.value"
                :aria-labelledby="`${option.value}-label`"
                :aria-describedby="`${option.value}-description site-theme-hint`"
              />
              <span class="settings-theme-option__swatches" aria-hidden="true">
                <span />
                <span />
                <span />
              </span>
              <span class="settings-theme-option__copy">
                <strong :id="`${option.value}-label`">{{ option.label }}</strong>
                <span :id="`${option.value}-description`">{{ option.description }}</span>
              </span>
            </label>
          </div>
        </fieldset>

        <label class="mist-field" for="site-default-mode">
          <span class="mist-field__label">访客默认模式</span>
          <select id="site-default-mode" v-model="form.mode" class="mist-input">
            <option value="light">浅色</option>
            <option value="dark">深色</option>
          </select>
          <span class="mist-field__hint">
            仅用于没有本地明暗偏好的访客，不会覆盖访客已经保存的选择。
          </span>
        </label>

        <fieldset class="settings-theme-elements" aria-describedby="theme-elements-hint">
          <legend class="mist-field__label">主题元素 · {{ selectedTheme.label }}</legend>
          <p id="theme-elements-hint" class="settings-theme-elements__description">
            以下文案跟随当前主题发布；切换主题可分别维护，未保存前不会影响前台。
          </p>
          <BaseInput
            v-model="selectedThemeElements.home_latest_empty_description"
            name="home-latest-empty-description"
            label="首页文章空状态"
            hint="用于“最近发布”暂时没有文章时的说明，最多 160 字。"
            :maxlength="160"
            required
          />
          <BaseInput
            v-model="selectedThemeElements.home_latest_end_text"
            name="home-latest-end-text"
            label="首页文章结束提示"
            hint="用于全部文章加载完毕后的结束提示，最多 80 字。"
            :maxlength="80"
            required
          />
        </fieldset>

        <section
          class="settings-theme-preview"
          :data-theme="form.theme"
          :data-mode="form.mode"
          aria-labelledby="settings-theme-preview-title"
        >
          <p class="settings-theme-preview__eyebrow">Theme preview</p>
          <div id="settings-theme-preview-title" role="heading" aria-level="3">{{ selectedTheme.label }}</div>
          <p>{{ selectedTheme.preview }}</p>
          <dl class="settings-theme-preview__elements">
            <div>
              <dt>首页空状态</dt>
              <dd>{{ selectedThemeElements.home_latest_empty_description }}</dd>
            </div>
            <div>
              <dt>首页列表结尾</dt>
              <dd>{{ selectedThemeElements.home_latest_end_text }}</dd>
            </div>
          </dl>
          <span class="settings-theme-preview__action" aria-hidden="true">主操作</span>
        </section>
      </section>

        <div class="settings-actions">
        <BaseButton type="submit" :loading="saving">保存设置</BaseButton>
        <p v-if="message" class="settings-message" role="status" aria-live="polite">{{ message }}</p>
        <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
      </div>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import {
  cloneThemeElementMap,
  createDefaultThemeElementMap,
  normalizeThemeElementMap,
  themeAppearanceOptions,
} from '@/config/themeAppearance'
import { useToast } from '@/composables/useToast'
import { getSettingDetail, updateSetting } from '@/api/modules/setting'
import { useSettingStore } from '@/stores/setting'
import type { LobbySetting, SettingPayload } from '@/types/setting'

const socialItems = [
  { key: 'gitee', label: 'Gitee' },
  { key: 'github', label: 'GitHub' },
  { key: 'bilibili', label: 'Bilibili' },
  { key: 'douyin', label: 'Douyin' },
] as const

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const message = ref('')
const toast = useToast()
const settingStore = useSettingStore()

const form = reactive<SettingPayload>({
  site_name: '',
  author: '',
  author_avatar_url: '',
  essay: '',
  icp_number: '',
  theme: 'mist-sea-salt',
  mode: 'light',
  theme_elements: createDefaultThemeElementMap(),
  social_links: {
    gitee: '',
    github: '',
    bilibili: '',
    douyin: '',
  },
})

const selectedTheme = computed(
  () =>
    themeAppearanceOptions.find((option) => option.value === form.theme) ??
    themeAppearanceOptions[0],
)
const selectedThemeElements = computed(() => form.theme_elements[form.theme])

onMounted(() => {
  void loadSetting()
})

function assignSetting(setting: LobbySetting) {
  form.site_name = setting.site_name
  form.author = setting.author
  form.author_avatar_url = setting.author_avatar_url
  form.essay = setting.essay
  form.icp_number = setting.icp_number
  form.theme = setting.theme
  form.mode = setting.mode
  form.theme_elements = normalizeThemeElementMap(setting.theme_elements)
  form.social_links = {
    gitee: setting.social_links.gitee ?? '',
    github: setting.social_links.github ?? '',
    bilibili: setting.social_links.bilibili ?? '',
    douyin: setting.social_links.douyin ?? '',
  }
}

async function loadSetting() {
  loading.value = true
  error.value = ''
  message.value = ''
  try {
    const setting = await getSettingDetail()
    settingStore.applyLobby(setting)
    assignSetting(setting)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载站点设置失败'
  } finally {
    loading.value = false
  }
}

async function saveSetting() {
  saving.value = true
  error.value = ''
  message.value = ''
  try {
    const saved = await updateSetting({
      ...form,
      theme_elements: cloneThemeElementMap(form.theme_elements),
      social_links: { ...form.social_links },
    })
    settingStore.applyLobby(saved)
    assignSetting(saved)
    message.value = '设置已保存'
    toast.success('站点设置已保存')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存站点设置失败'
    toast.error(error.value)
  } finally {
    saving.value = false
  }
}
</script>
