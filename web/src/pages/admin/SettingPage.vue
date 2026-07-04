<template>
  <section class="admin-page">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Settings</p>
        <h1>站点设置</h1>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadSetting">刷新</BaseButton>
    </header>

    <form class="settings-form" @submit.prevent="saveSetting">
      <section class="settings-panel">
        <h2>基础信息</h2>
        <BaseInput v-model="form.site_name" label="站点名称" />
        <BaseInput v-model="form.author" label="作者" />
        <BaseTextarea v-model="form.essay" label="站点签名" :rows="4" />

        <div class="settings-form__row">
          <label class="cui-field">
            <span class="cui-field__label">主题</span>
            <select v-model="form.theme" class="cui-input">
              <option value="forest">Forest</option>
              <option value="strawberry">Strawberry</option>
            </select>
          </label>

          <label class="cui-field">
            <span class="cui-field__label">模式</span>
            <select v-model="form.mode" class="cui-input">
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </select>
          </label>
        </div>
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
        <p v-if="message" class="settings-message">{{ message }}</p>
        <p v-if="error" class="admin-page__error">{{ error }}</p>
      </div>
    </form>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import { getSettingDetail, updateSetting } from '@/api/modules/setting'
import type { SettingPayload } from '@/types/setting'

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

const form = reactive<SettingPayload>({
  site_name: '',
  author: '',
  essay: '',
  theme: 'forest',
  mode: 'light',
  social_links: {
    gitee: '',
    github: '',
    bilibili: '',
    douyin: '',
  },
})

onMounted(() => {
  loadSetting()
})

async function loadSetting() {
  loading.value = true
  error.value = ''
  message.value = ''
  try {
    const setting = await getSettingDetail()
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
    applyTheme()
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
    const saved = await updateSetting({ ...form, social_links: { ...form.social_links } })
    form.site_name = saved.site_name
    form.author = saved.author
    form.essay = saved.essay
    form.theme = saved.theme
    form.mode = saved.mode
    form.social_links = {
      gitee: saved.social_links.gitee ?? '',
      github: saved.social_links.github ?? '',
      bilibili: saved.social_links.bilibili ?? '',
      douyin: saved.social_links.douyin ?? '',
    }
    applyTheme()
    message.value = '设置已保存'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存站点设置失败'
  } finally {
    saving.value = false
  }
}

function applyTheme() {
  document.documentElement.dataset.theme = form.theme
  document.documentElement.dataset.mode = form.mode
}
</script>
