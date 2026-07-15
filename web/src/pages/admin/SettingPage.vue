<template>
  <section class="admin-page" :aria-busy="loading || saving">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Settings</p>
        <h1>站点设置</h1>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadSetting">刷新</BaseButton>
    </header>

    <form class="settings-form" :aria-busy="saving" @submit.prevent="saveSetting">
      <section class="settings-panel">
        <h2>基础信息</h2>
        <BaseInput v-model="form.site_name" label="站点名称" />
        <BaseInput v-model="form.author" label="作者" />
        <BaseTextarea v-model="form.essay" label="站点签名" :rows="4" />
      </section>

      <section class="settings-panel settings-panel--appearance">
        <h2>主题外观</h2>
        <p id="site-theme-hint" class="settings-panel__description">
          主题色由后台统一发布；访客只能在前台切换明暗模式。
        </p>

        <fieldset class="settings-theme-picker" aria-describedby="site-theme-hint">
          <legend class="mist-field__label">站点主题色</legend>
          <div class="settings-theme-picker__options">
            <label
              v-for="option in themeOptions"
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

        <section
          class="settings-theme-preview"
          :data-theme="form.theme"
          :data-mode="form.mode"
          aria-labelledby="settings-theme-preview-title"
        >
          <p class="settings-theme-preview__eyebrow">Theme preview</p>
          <h3 id="settings-theme-preview-title">{{ selectedTheme.label }}</h3>
          <p>{{ selectedTheme.preview }}</p>
          <span class="settings-theme-preview__action" aria-hidden="true">主操作</span>
        </section>
      </section>

      <section class="settings-panel">
        <h2>社交链接</h2>
        <BaseInput
          v-for="item in socialItems"
          :key="item.key"
          v-model="form.social_links[item.key]"
          :label="item.label"
        />
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
import { useToast } from '@/composables/useToast'
import { getSettingDetail, updateSetting } from '@/api/modules/setting'
import { useSettingStore } from '@/stores/setting'
import type { LobbySetting, SettingPayload, ThemeName } from '@/types/setting'

const socialItems = [
  { key: 'gitee', label: 'Gitee' },
  { key: 'github', label: 'GitHub' },
  { key: 'bilibili', label: 'Bilibili' },
  { key: 'douyin', label: 'Douyin' },
] as const

const themeOptions = [
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
] as const satisfies ReadonlyArray<{
  value: ThemeName
  label: string
  description: string
  preview: string
}>

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const message = ref('')
const toast = useToast()
const settingStore = useSettingStore()

const form = reactive<SettingPayload>({
  site_name: '',
  author: '',
  essay: '',
  theme: 'mist-sea-salt',
  mode: 'light',
  social_links: {
    gitee: '',
    github: '',
    bilibili: '',
    douyin: '',
  },
})

const selectedTheme = computed(
  () => themeOptions.find((option) => option.value === form.theme) ?? themeOptions[0],
)

onMounted(() => {
  void loadSetting()
})

function assignSetting(setting: LobbySetting) {
  form.site_name = setting.site_name
  form.author = setting.author
  form.essay = setting.essay
  form.theme = setting.theme
  form.mode = setting.mode
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
