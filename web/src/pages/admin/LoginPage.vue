<template>
  <main class="login-page mist-page">
    <form class="login-card" :aria-busy="loading" @submit.prevent="submit">
      <header class="login-card__header">
        <span class="login-card__mark" aria-hidden="true">S</span>
        <div>
          <p class="login-card__eyebrow">Solitude Admin</p>
          <div role="heading" aria-level="1">欢迎回来</div>
        </div>
      </header>
      <p class="login-card__description">登录后继续管理文章、媒体与站点设置。</p>
      <p
        v-if="error"
        ref="errorSummary"
        class="login-card__error"
        role="alert"
        aria-live="assertive"
        tabindex="-1"
      >
        {{ error }}
      </p>
      <BaseInput
        v-model="form.username"
        name="username"
        label="账号"
        autocomplete="username"
        :disabled="loading"
        required
      />
      <BaseInput
        v-model="form.password"
        name="password"
        label="密码"
        type="password"
        autocomplete="current-password"
        :disabled="loading"
        required
      />
      <BaseButton type="submit" :loading="loading">登录</BaseButton>
    </form>
  </main>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import { ApiError } from '@/api/types'
import { resolveAdminRedirect } from '@/router/guards'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const errorSummary = ref<HTMLElement | null>(null)
const form = reactive({
  username: '',
  password: '',
})

function loginErrorMessage(err: unknown) {
  if (err instanceof ApiError) {
    if (err.code === 10005) {
      return '账号或密码不正确'
    }
    if (err.code === 10006) {
      return '当前账号已被停用'
    }
    if (err.code === 40000 || err.code === 40003) {
      return '尝试次数过多，请稍后再试'
    }
  }
  return '暂时无法登录，请检查网络连接后重试'
}

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(form)
    await router.replace(resolveAdminRedirect(route.query.redirect))
  } catch (err) {
    error.value = loginErrorMessage(err)
    await nextTick()
    errorSummary.value?.focus()
  } finally {
    loading.value = false
  }
}
</script>
