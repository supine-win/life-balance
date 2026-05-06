<template>
  <div class="setup-page">
    <div class="setup-card">
      <h1>🎬 LifeRecorder 安装向导</h1>
      <p>欢迎使用 LifeRecorder，让我们开始初始化设置</p>
      <a-steps :current="currentStep" style="margin-bottom:32px">
        <a-step title="数据库" />
        <a-step title="管理员" />
        <a-step title="完成" />
      </a-steps>
      <div v-if="currentStep===0">
        <a-form layout="vertical">
          <a-form-item label="数据库类型">
            <a-select v-model:value="dbType"><a-select-option value="sqlite">SQLite (推荐)</a-select-option><a-select-option value="postgres">PostgreSQL</a-select-option></a-select>
          </a-form-item>
          <a-button type="primary" @click="currentStep=1">下一步</a-button>
        </a-form>
      </div>
      <div v-if="currentStep===1">
        <a-form :model="adminForm" @finish="handleInit" layout="vertical">
          <a-form-item label="邮箱" name="email" :rules="[{required:true,type:'email'}]">
            <a-input v-model:value="adminForm.email" placeholder="管理员邮箱" />
          </a-form-item>
          <a-form-item label="密码" name="password" :rules="[{required:true,min:8}]">
            <a-input-password v-model:value="adminForm.password" placeholder="管理员密码" />
          </a-form-item>
          <a-space><a-button @click="currentStep=0">上一步</a-button><a-button type="primary" html-type="submit">完成安装</a-button></a-space>
        </a-form>
      </div>
      <div v-if="currentStep===2">
        <a-result status="success" title="安装完成" sub-title="LifeRecorder 已准备就绪">
          <template #extra><a-button type="primary" @click="$router.push('/login')">开始使用</a-button></template>
        </a-result>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive } from 'vue'
import { message } from 'ant-design-vue'
import { setupApi } from '@/api'
const currentStep = ref(0)
const dbType = ref('sqlite')
const adminForm = reactive({ email: '', password: '' })
async function handleInit() {
  try {
    await setupApi.init(adminForm)
    currentStep.value = 2
    message.success('安装成功')
  } catch (err: any) { message.error(err.message || '安装失败') }
}
</script>
<style scoped>
.setup-page{min-height:100vh;display:flex;align-items:center;justify-content:center;background:#f0f2f5}
.setup-card{background:#fff;border-radius:12px;padding:40px;width:600px;box-shadow:0 2px 8px rgba(0,0,0,.1)}
.setup-card h1{text-align:center;margin-bottom:8px}
</style>
