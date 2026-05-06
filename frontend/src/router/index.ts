import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 安装初始化
    {
      path: '/setup',
      name: 'setup',
      component: () => import('@/views/setup/SetupWizard.vue'),
      meta: { layout: 'setup', public: true },
    },
    // 认证
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { layout: 'auth', public: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/Register.vue'),
      meta: { layout: 'auth', public: true },
    },
    // 主页 - 重定向
    {
      path: '/',
      redirect: '/events',
    },
    // 事件管理
    {
      path: '/events',
      name: 'events',
      component: () => import('@/views/events/EventList.vue'),
      meta: { layout: 'main', title: '事件列表' },
    },
    {
      path: '/events/calendar',
      name: 'events-calendar',
      component: () => import('@/views/events/EventCalendar.vue'),
      meta: { layout: 'main', title: '日历视图' },
    },
    {
      path: '/events/map',
      name: 'events-map',
      component: () => import('@/views/events/EventMap.vue'),
      meta: { layout: 'main', title: '地图视图' },
    },
    {
      path: '/events/create',
      name: 'events-create',
      component: () => import('@/views/events/EventForm.vue'),
      meta: { layout: 'main', title: '创建事件' },
    },
    {
      path: '/events/create/voice',
      name: 'events-create-voice',
      component: () => import('@/views/events/VoiceInput.vue'),
      meta: { layout: 'main', title: '语音录入' },
    },
    {
      path: '/events/create/chat',
      name: 'events-create-chat',
      component: () => import('@/views/events/ChatInput.vue'),
      meta: { layout: 'main', title: '对话式录入' },
    },
    {
      path: '/events/:id',
      name: 'events-detail',
      component: () => import('@/views/events/EventDetail.vue'),
      meta: { layout: 'main', title: '事件详情' },
    },
    {
      path: '/events/:id/edit',
      name: 'events-edit',
      component: () => import('@/views/events/EventForm.vue'),
      meta: { layout: 'main', title: '编辑事件' },
    },
    // 展示模式
    {
      path: '/display',
      name: 'display',
      component: () => import('@/views/display/DisplayHome.vue'),
      meta: { layout: 'main', title: '展示模式' },
    },
    {
      path: '/display/slideshow/:id',
      name: 'slideshow',
      component: () => import('@/views/display/SlideshowPlayer.vue'),
      meta: { layout: 'display', title: '幻灯片播放' },
    },
    {
      path: '/display/templates',
      name: 'display-templates',
      component: () => import('@/views/display/TemplateList.vue'),
      meta: { layout: 'main', title: '模板管理' },
    },
    // 媒体库
    {
      path: '/media',
      name: 'media',
      component: () => import('@/views/media/MediaLibrary.vue'),
      meta: { layout: 'main', title: '媒体库' },
    },
    // Webhook 管理
    {
      path: '/webhooks',
      name: 'webhooks',
      component: () => import('@/views/webhooks/WebhookList.vue'),
      meta: { layout: 'main', title: 'Webhook' },
    },
    // 个人设置
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/settings/ProfileSettings.vue'),
      meta: { layout: 'main', title: '设置' },
    },
    // 管理后台
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/admin/Dashboard.vue'),
      meta: { layout: 'admin', title: '管理后台', permission: 'admin:users' },
    },
    {
      path: '/admin/users',
      name: 'admin-users',
      component: () => import('@/views/admin/UserList.vue'),
      meta: { layout: 'admin', title: '用户管理', permission: 'admin:users' },
    },
    {
      path: '/admin/roles',
      name: 'admin-roles',
      component: () => import('@/views/admin/RoleList.vue'),
      meta: { layout: 'admin', title: '角色管理', permission: 'admin:users' },
    },
    {
      path: '/admin/config',
      name: 'admin-config',
      component: () => import('@/views/admin/ConfigEditor.vue'),
      meta: { layout: 'admin', title: '系统配置', permission: 'admin:config' },
    },
    {
      path: '/admin/audit-logs',
      name: 'admin-audit',
      component: () => import('@/views/admin/AuditLogList.vue'),
      meta: { layout: 'admin', title: '审计日志', permission: 'admin:users' },
    },
    {
      path: '/admin/backup',
      name: 'admin-backup',
      component: () => import('@/views/admin/BackupManager.vue'),
      meta: { layout: 'admin', title: '数据备份', permission: 'admin:backup' },
    },
    // 404
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/auth/Login.vue'), // fallback
    },
  ],
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  // 公开页面直接放行
  if (to.meta.public) {
    next()
    return
  }

  const authStore = useAuthStore()

  // 检查是否已认证
  if (!authStore.isAuthenticated) {
    // 尝试从本地存储恢复
    authStore.restoreSession()
    if (!authStore.isAuthenticated) {
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }
  }

  // 检查权限
  if (to.meta.permission && !authStore.hasPermission(to.meta.permission as string)) {
    next({ name: 'events' })
    return
  }

  next()
})

export default router
