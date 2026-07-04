<template>
  <main class="login-page creamy-page">
    <form class="login-card" @submit.prevent="submit">
      <h1>后台登录</h1>
      <BaseInput v-model="form.username" label="账号" autocomplete="username" />
      <BaseInput v-model="form.password" label="密码" type="password" autocomplete="current-password" />
      <BaseButton type="submit" :loading="loading">登录</BaseButton>
      <p v-if="error" class="login-card__error">{{ error }}</p>
    </form>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const form = reactive({
  username: 'admin',
  password: 'admin',
})

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(form)
    await router.push(String(route.query.redirect ?? '/admin'))
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'login failed'
  } finally {
    loading.value = false
  }
}
</script>

