<template>
  <div class="login-page">
    <div class="login-card">
      <h2>注册 LifeRecorder</h2>
      <a-form :model="form" @finish="handleRegister" layout="vertical">
        <a-form-item name="username" :rules="[{required:true,message:'请输入用户名'},{min:3,message:'至少3个字符'}]">
          <a-input v-model:value="form.username" placeholder="用户名" size="large" />
        </a-form-item>
        <a-form-item name="email" :rules="[{required:true,message:'请输入邮箱'},{type:'email',message:'邮箱格式不正确'}]">
          <a-input v-model:value="form.email" placeholder="邮箱" size="large" />
        </a-form-item>
        <a-form-item name="password" :rules="[{required:true,message:'请输入密码'},{min:8,message:'至少8个字符'}]">
          <a-input-password v-model:value="form.password" placeholder="密码" size="large" />
        </a-form-item>
        <a-form-item name="nickname">
          <a-input v-model:value="form.nickname" placeholder="昵称（可选）" size="large" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" block size="large">注册</a-button>
        </a-form-item>
        <div style="text-align:center;color:#666">已有账号？ <router-link to="/login">登录</router-link></div>
      </a-form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useAuthStore } from '@/stores/auth'
const router = useRouter()
const authStore = useAuthStore()
const form = reactive({ username: '', email: '', password: '', nickname: '' })
async function handleRegister() {
  try {
    await authStore.register(form)
    message.success('注册成功')
    router.push('/events')
  } catch (err: any) { message.error(err.message || '注册失败') }
}
</script>
<style scoped>
.login-page{min-height:100vh;display:flex;align-items:center;justify-content:center;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%)}
.login-card{background:#fff;border-radius:12px;padding:40px;width:400px;box-shadow:0 8px 32px rgba(0,0,0,.15)}
.login-card h2{text-align:center;margin-bottom:24px}
</style>
