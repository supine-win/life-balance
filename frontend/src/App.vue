<template>
  <a-config-provider :locale="zhCN">
    <component :is="layoutComponent">
      <router-view />
    </component>
  </a-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import MainLayout from '@/layouts/MainLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import DisplayLayout from '@/layouts/DisplayLayout.vue'

const route = useRoute()

const layoutComponent = computed(() => {
  const layout = route.meta.layout as string
  switch (layout) {
    case 'auth':
    case 'setup':
      return AuthLayout
    case 'display':
      return DisplayLayout
    case 'main':
    case 'admin':
    default:
      return MainLayout
  }
})
</script>

<style>
#app {
  width: 100%;
  min-height: 100vh;
}
</style>
