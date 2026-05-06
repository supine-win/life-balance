<template>
  <div class="page-container">
    <!-- 顶部操作栏 -->
    <div class="page-header">
      <div class="header-actions">
        <a-input-search
          v-model:value="searchKeyword"
          placeholder="搜索事件..."
          style="width: 300px"
          @search="handleSearch"
        />
        <a-select v-model:value="filters.category" placeholder="分类" allowClear style="width: 120px" @change="handleFilter">
          <a-select-option value="旅行">旅行</a-select-option>
          <a-select-option value="美食">美食</a-select-option>
          <a-select-option value="运动">运动</a-select-option>
          <a-select-option value="工作">工作</a-select-option>
          <a-select-option value="学习">学习</a-select-option>
          <a-select-option value="家庭">家庭</a-select-option>
          <a-select-option value="摄影">摄影</a-select-option>
        </a-select>
        <a-select v-model:value="filters.mood" placeholder="心情" allowClear style="width: 100px" @change="handleFilter">
          <a-select-option value="happy">😊 开心</a-select-option>
          <a-select-option value="calm">😌 平静</a-select-option>
          <a-select-option value="excited">🤩 兴奋</a-select-option>
          <a-select-option value="tired">😴 疲惫</a-select-option>
          <a-select-option value="sad">😢 难过</a-select-option>
        </a-select>
      </div>
      <div>
        <a-radio-group v-model:value="viewMode" button-style="solid" size="small">
          <a-radio-button value="timeline"><UnorderedListOutlined /> 时间线</a-radio-button>
          <a-radio-button value="grid"><AppstoreOutlined /> 网格</a-radio-button>
        </a-radio-group>
        <a-button type="primary" @click="$router.push('/events/create')" style="margin-left: 12px">
          <PlusOutlined /> 新建事件
        </a-button>
      </div>
    </div>

    <!-- 事件列表 -->
    <a-spin :spinning="eventStore.loading">
      <div v-if="viewMode === 'timeline'" class="timeline-view">
        <a-timeline>
          <a-timeline-item
            v-for="event in eventStore.events"
            :key="event.id"
            :color="event.color || 'blue'"
          >
            <div class="lr-card event-card" @click="goToDetail(event.id)">
              <div class="event-card-header">
                <span class="event-color-dot" :style="{ background: event.color || '#1890ff' }"></span>
                <h4 class="event-title">{{ event.title }}</h4>
                <a-tag v-if="event.category" :color="getCategoryColor(event.category)">{{ event.category }}</a-tag>
                <span class="event-time">{{ formatDate(event.event_time) }}</span>
              </div>
              <p v-if="event.description" class="event-desc">{{ event.description }}</p>
              <div class="event-meta">
                <span v-if="event.location"><EnvironmentOutlined /> {{ event.location }}</span>
                <span v-if="event.mood">{{ getMoodEmoji(event.mood) }}</span>
                <a-tag v-for="tag in (event.tags || []).slice(0, 3)" :key="tag" size="small">{{ tag }}</a-tag>
              </div>
              <!-- 媒体缩略图 -->
              <div v-if="event.media?.length" class="event-media-row">
                <div v-for="m in event.media.slice(0, 4)" :key="m.id" class="media-thumb">
                  <img v-if="m.type === 'image'" :src="`/api/v1/media/${m.id}/file`" class="media-thumbnail" />
                  <div v-else class="media-type-icon">
                    <PlayCircleOutlined />
                  </div>
                </div>
                <span v-if="event.media.length > 4" class="media-more">+{{ event.media.length - 4 }}</span>
              </div>
            </div>
          </a-timeline-item>
        </a-timeline>
      </div>

      <div v-else class="grid-view">
        <a-row :gutter="[16, 16]">
          <a-col v-for="event in eventStore.events" :key="event.id" :xs="24" :sm="12" :md="8" :lg="6">
            <div class="lr-card event-grid-card" @click="goToDetail(event.id)">
              <div v-if="event.media?.length" class="grid-card-cover">
                <img :src="`/api/v1/media/${event.media[0].id}/file`" />
              </div>
              <div class="grid-card-body">
                <h4>{{ event.title }}</h4>
                <p class="text-muted">{{ formatDate(event.event_time) }}</p>
                <div class="grid-card-tags">
                  <a-tag v-for="tag in (event.tags || []).slice(0, 2)" :key="tag" size="small">{{ tag }}</a-tag>
                </div>
              </div>
            </div>
          </a-col>
        </a-row>
      </div>

      <!-- 空状态 -->
      <div v-if="!eventStore.loading && eventStore.events.length === 0" class="empty-state">
        <CalendarOutlined style="font-size: 48px; color: #ccc" />
        <p style="margin-top: 16px; font-size: 16px">还没有任何事件记录</p>
        <a-button type="primary" @click="$router.push('/events/create')">
          <PlusOutlined /> 创建第一个事件
        </a-button>
      </div>

      <!-- 分页 -->
      <div v-if="eventStore.total > 20" class="pagination">
        <a-pagination
          v-model:current="currentPage"
          :total="eventStore.total"
          :pageSize="20"
          @change="handlePageChange"
          show-less-items
        />
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useEventStore } from '@/stores/event'
import dayjs from 'dayjs'
import {
  PlusOutlined, UnorderedListOutlined, AppstoreOutlined,
  CalendarOutlined, EnvironmentOutlined, PlayCircleOutlined,
} from '@ant-design/icons-vue'

const router = useRouter()
const eventStore = useEventStore()
const viewMode = ref('timeline')
const searchKeyword = ref('')
const currentPage = ref(1)
const filters = ref<Record<string, string>>({
  category: '',
  mood: '',
})

onMounted(() => {
  eventStore.fetchEvents()
})

function handleSearch() {
  eventStore.fetchEvents({ keyword: searchKeyword.value, page: 1 })
  currentPage.value = 1
}

function handleFilter() {
  eventStore.fetchEvents({
    category: filters.value.category || undefined,
    mood: filters.value.mood || undefined,
    page: 1,
  })
  currentPage.value = 1
}

function handlePageChange(page: number) {
  eventStore.fetchEvents({
    keyword: searchKeyword.value || undefined,
    category: filters.value.category || undefined,
    page,
  })
}

function goToDetail(id: string) {
  router.push(`/events/${id}`)
}

function formatDate(dateStr: string) {
  return dayjs(dateStr).format('YYYY-MM-DD HH:mm')
}

function getCategoryColor(category: string): string {
  const colors: Record<string, string> = {
    '旅行': 'blue', '美食': 'orange', '运动': 'green',
    '工作': 'purple', '学习': 'cyan', '家庭': 'pink', '摄影': 'gold',
  }
  return colors[category] || 'default'
}

function getMoodEmoji(mood: string): string {
  const emojis: Record<string, string> = {
    happy: '😊', calm: '😌', excited: '🤩', tired: '😴', sad: '😢',
  }
  return emojis[mood] || ''
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.header-actions {
  display: flex;
  gap: 12px;
}
.event-card {
  cursor: pointer;
  margin-bottom: 0;
  width: 100%;
}
.event-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
.event-title {
  flex: 1;
  margin: 0;
  font-size: 15px;
}
.event-time {
  color: var(--lr-text-muted);
  font-size: 13px;
}
.event-desc {
  margin: 8px 0 0 28px;
  color: var(--lr-text-secondary);
  font-size: 13px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.event-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  margin-left: 28px;
  font-size: 13px;
  color: var(--lr-text-muted);
}
.event-media-row {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  margin-left: 28px;
}
.media-thumb {
  width: 60px;
  height: 60px;
  border-radius: 4px;
  overflow: hidden;
  background: #f0f0f0;
}
.media-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 24px;
  color: #999;
}
.media-more {
  display: flex;
  align-items: center;
  color: var(--lr-text-muted);
  font-size: 13px;
}
.grid-card-cover {
  height: 150px;
  overflow: hidden;
  border-radius: var(--lr-border-radius) var(--lr-border-radius) 0 0;
}
.grid-card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.grid-card-body {
  padding: 12px;
}
.grid-card-body h4 {
  margin: 0 0 4px;
  font-size: 14px;
}
.text-muted {
  color: var(--lr-text-muted);
  font-size: 12px;
  margin: 0;
}
.grid-card-tags {
  margin-top: 8px;
}
.pagination {
  text-align: center;
  margin-top: 24px;
}
</style>
