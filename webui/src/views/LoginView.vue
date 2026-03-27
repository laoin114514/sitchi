<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import { login } from '@/api/auth'
import LoginCard from '@/components/LoginCard.vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const theme = ref<'dark' | 'light'>('dark')

const applyTheme = (nextTheme: 'dark' | 'light') => {
  theme.value = nextTheme
  document.documentElement.setAttribute('data-theme', nextTheme)
  localStorage.setItem('theme', nextTheme)
}

const toggleTheme = () => {
  applyTheme(theme.value === 'dark' ? 'light' : 'dark')
}

const handleSubmit = async (payload: { userCode: string; password: string }) => {
  if (!payload.userCode || !payload.password) {
    ElMessage.warning('请输入完整账号密码')
    return
  }

  loading.value = true
  try {
    const response = await login({
      user_code: payload.userCode,
      password: payload.password,
    })

    if (!response.data) {
      throw new Error('登录响应数据为空')
    }

    authStore.setAuth({
      accessToken: response.data.access_token,
      refreshToken: response.data.refresh_token,
      userId: response.data.user_id,
      moduleCode: response.data.module_code,
      roles: response.data.roles,
    })

    ElMessage.success('登录成功')
    await router.push('/')
  } catch (error) {
    const message = error instanceof Error ? error.message : '登录失败'
    ElMessage.error(message)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const localTheme = localStorage.getItem('theme')
  if (localTheme === 'light' || localTheme === 'dark') {
    applyTheme(localTheme)
    return
  }
  applyTheme('dark')
})
</script>

<template>
  <div class="login-view">
    <div class="bg-layer"></div>
    <LoginCard :loading="loading" :theme="theme" @submit="handleSubmit" @toggle-theme="toggleTheme" />
  </div>
</template>

<style scoped>
.login-view {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 24px;
  background: radial-gradient(circle at top left, rgba(58, 109, 240, 0.16), transparent 42%), var(--color-bg);
}

.bg-layer {
  position: absolute;
  inset: 0;
  background: var(--color-gradient);
  pointer-events: none;
}

:deep(.el-form-item__label) {
  color: var(--color-text-secondary);
}

:deep(.el-input__wrapper) {
  background: var(--color-bg-secondary);
  box-shadow: 0 0 0 1px var(--color-border) inset;
}

:deep(.el-input__inner) {
  color: var(--color-text);
}

:deep(.el-button--primary) {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

:deep(.el-button--primary:hover) {
  background: var(--color-primary-hover);
  border-color: var(--color-primary-hover);
}
</style>
