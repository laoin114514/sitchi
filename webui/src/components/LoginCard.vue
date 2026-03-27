<script setup lang="ts">
import { computed, reactive } from 'vue'
import { Moon, Sunny } from '@element-plus/icons-vue'

interface LoginForm {
  userCode: string
  password: string
}

const props = defineProps<{
  loading?: boolean
  theme: 'dark' | 'light'
}>()

const emit = defineEmits<{
  submit: [payload: LoginForm]
  toggleTheme: []
}>()

const form = reactive<LoginForm>({
  userCode: '',
  password: '',
})

const rules = {
  userCode: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const themeIcon = computed(() => (props.theme === 'dark' ? Sunny : Moon))

const onSubmit = () => {
  emit('submit', {
    userCode: form.userCode.trim(),
    password: form.password,
  })
}

const onToggleTheme = () => {
  emit('toggleTheme')
}
</script>

<template>
  <div class="login-card">
    <div class="login-header">
      <div>
        <h1 class="login-title">Sitchi</h1>
        <p class="login-subtitle">系统管理登录</p>
      </div>
      <el-button class="theme-btn" circle text @click="onToggleTheme">
        <el-icon :size="18">
          <component :is="themeIcon" />
        </el-icon>
      </el-button>
    </div>

    <el-form :model="form" :rules="rules" label-position="top">
      <el-form-item label="账号" prop="userCode">
        <el-input v-model="form.userCode" placeholder="请输入管理员账号" autocomplete="username" />
      </el-form-item>

      <el-form-item label="密码" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="请输入密码"
          autocomplete="current-password"
          @keyup.enter="onSubmit"
        />
      </el-form-item>

      <el-button class="submit-btn" type="primary" :loading="props.loading" @click="onSubmit">
        登录
      </el-button>
    </el-form>
  </div>
</template>

<style scoped>
.login-card {
  width: 420px;
  padding: 32px;
  border-radius: 16px;
  border: 1px solid var(--color-border);
  background: var(--color-card);
  box-shadow: 0 16px 40px var(--color-shadow);
}

.login-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.login-title {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text);
}

.login-subtitle {
  margin: 6px 0 0;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.theme-btn {
  color: var(--color-text-secondary);
}

.submit-btn {
  width: 100%;
  margin-top: 6px;
}
</style>
