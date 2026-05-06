<template>
  <div class="login-page">
    <div class="login-card">
      <h2><CameraOutlined /> LifeRecorder</h2>
      <p class="subtitle">记录生活中的每一个精彩瞬间</p>
      <a-form :model="form" @finish="handleLogin" layout="vertical">
        <a-form-item name="username" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model:value="form.username" placeholder="用户名或邮箱" size="large">
            <template #prefix><UserOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item name="password" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model:value="form.password" placeholder="密码" size="large">
            <template #prefix><LockOutlined /></template>
          </a-input-password>
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" :loading="authStore.loading" block size="large">
            登录
          </a-button>
        </a-form-item>
        <div class="login-footer">
          还没有账号？ <router-link to="/register">立即注册</router-link>
        </div>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { useAuthStore } from '@/stores/auth'
import { CameraOutlined, UserOutlined, LockOutlined } from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const form = reactive({ username: '', password: '' })

async function handleLogin() {
  try {
    await authStore.login(form.username, form.password)
    message.success('登录成功')
    const redirect = (route.query.redirect as string) || '/events'
    router.push(redirect)
  } catch (err: any) {
    message.error(err.message || '登录失败')
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  width: 400px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}
.login-card h2 {
  text-align: center;
  margin-bottom: 8px;
  font-size: 24px;
}
.subtitle {
  text-align: center;
  color: #999;
  margin-bottom: 32px;
}
.login-footer {
  text-align: center;
  color: #666;
}
</style>
