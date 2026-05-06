<template>
  <a-layout class="main-layout" style="min-height: 100vh">
    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      :width="240"
      class="layout-sider"
    >
      <div class="logo">
        <CameraOutlined style="font-size: 24px" />
        <span v-if="!collapsed" class="logo-text">LifeRecorder</span>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
      >
        <a-menu-item key="events" @click="$router.push('/events')">
          <CalendarOutlined />
          <span>事件列表</span>
        </a-menu-item>
        <a-menu-item key="calendar" @click="$router.push('/events/calendar')">
          <AppstoreOutlined />
          <span>日历视图</span>
        </a-menu-item>
        <a-menu-item key="create" @click="$router.push('/events/create')">
          <PlusOutlined />
          <span>创建事件</span>
        </a-menu-item>
        <a-sub-menu key="input-group">
          <template #title><AudioOutlined /><span>快速录入</span></template>
          <a-menu-item key="voice" @click="$router.push('/events/create/voice')">
            <AudioOutlined /> 语音录入
          </a-menu-item>
          <a-menu-item key="chat" @click="$router.push('/events/create/chat')">
            <MessageOutlined /> 对话录入
          </a-menu-item>
        </a-sub-menu>
        <a-menu-divider />
        <a-menu-item key="media" @click="$router.push('/media')">
          <PictureOutlined />
          <span>媒体库</span>
        </a-menu-item>
        <a-menu-item key="display" @click="$router.push('/display')">
          <PlayCircleOutlined />
          <span>展示模式</span>
        </a-menu-item>
        <a-menu-item key="webhooks" @click="$router.push('/webhooks')">
          <ApiOutlined />
          <span>Webhook</span>
        </a-menu-item>
        <a-menu-divider />
        <a-menu-item v-if="authStore.isAdmin" key="admin" @click="$router.push('/admin')">
          <SettingOutlined />
          <span>管理后台</span>
        </a-menu-item>
        <a-menu-item key="settings" @click="$router.push('/settings')">
          <UserOutlined />
          <span>个人设置</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <!-- 右侧内容 -->
    <a-layout>
      <!-- 顶栏 -->
      <a-layout-header class="layout-header">
        <div class="header-left">
          <a-button type="text" @click="collapsed = !collapsed">
            <MenuUnfoldOutlined v-if="collapsed" />
            <MenuFoldOutlined v-else />
          </a-button>
          <h3 class="page-title">{{ currentTitle }}</h3>
        </div>
        <div class="header-right">
          <a-dropdown>
            <a-button type="text">
              <UserOutlined /> {{ authStore.username }}
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="$router.push('/settings')">个人设置</a-menu-item>
                <a-menu-divider />
                <a-menu-item @click="handleLogout">退出登录</a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- 内容区 -->
      <a-layout-content class="layout-content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  CameraOutlined, CalendarOutlined, AppstoreOutlined,
  PlusOutlined, AudioOutlined, MessageOutlined,
  PictureOutlined, PlayCircleOutlined, ApiOutlined,
  SettingOutlined, UserOutlined, MenuUnfoldOutlined,
  MenuFoldOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const collapsed = ref(false)

const selectedKeys = computed(() => {
  const path = route.path
  if (path.startsWith('/events/calendar')) return ['calendar']
  if (path.startsWith('/events/create/voice')) return ['voice']
  if (path.startsWith('/events/create/chat')) return ['chat']
  if (path.startsWith('/events/create')) return ['create']
  if (path.startsWith('/events')) return ['events']
  if (path.startsWith('/media')) return ['media']
  if (path.startsWith('/display')) return ['display']
  if (path.startsWith('/webhooks')) return ['webhooks']
  if (path.startsWith('/admin')) return ['admin']
  if (path.startsWith('/settings')) return ['settings']
  return ['events']
})

const currentTitle = computed(() => (route.meta.title as string) || 'LifeRecorder')

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}
.layout-sider {
  overflow: auto;
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 10;
}
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
}
.logo-text {
  white-space: nowrap;
}
.layout-header {
  background: var(--lr-bg-white);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  position: sticky;
  top: 0;
  z-index: 9;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.page-title {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}
.header-right {
  display: flex;
  align-items: center;
}
.layout-content {
  margin-left: 240px;
  padding: 24px;
  min-height: calc(100vh - 56px);
  transition: margin-left var(--lr-transition-normal);
}
.layout-sider.ant-layout-sider-collapsed + .a-layout .layout-content {
  margin-left: 64px;
}
</style>
