<template>
  <div class="page-container">
    <a-card :title="isEdit ? '编辑事件' : '创建事件'" :bordered="false">
      <a-form
        :model="form"
        :label-col="{ span: 4 }"
        :wrapper-col="{ span: 16 }"
        @finish="handleSubmit"
      >
        <a-form-item label="标题" name="title" :rules="[{ required: true, message: '请输入标题' }]">
          <a-input v-model:value="form.title" placeholder="事件标题" size="large" />
        </a-form-item>

        <a-form-item label="描述">
          <a-textarea v-model:value="form.description" placeholder="详细描述..." :rows="4" />
        </a-form-item>

        <a-form-item label="时间" name="event_time" :rules="[{ required: true, message: '请选择时间' }]">
          <a-date-picker
            v-model:value="form.event_time"
            show-time
            format="YYYY-MM-DD HH:mm"
            placeholder="选择时间"
            style="width: 100%"
          />
        </a-form-item>

        <a-form-item label="地点">
          <a-input v-model:value="form.location" placeholder="地点" />
        </a-form-item>

        <a-form-item label="分类">
          <a-select v-model:value="form.category" placeholder="选择分类" allowClear>
            <a-select-option value="旅行">🌍 旅行</a-select-option>
            <a-select-option value="美食">🍜 美食</a-select-option>
            <a-select-option value="运动">🏃 运动</a-select-option>
            <a-select-option value="工作">💼 工作</a-select-option>
            <a-select-option value="学习">📚 学习</a-select-option>
            <a-select-option value="家庭">🏠 家庭</a-select-option>
            <a-select-option value="摄影">📷 摄影</a-select-option>
            <a-select-option value="朋友">👥 朋友</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="心情">
          <a-radio-group v-model:value="form.mood">
            <a-radio-button value="happy">😊 开心</a-radio-button>
            <a-radio-button value="calm">😌 平静</a-radio-button>
            <a-radio-button value="excited">🤩 兴奋</a-radio-button>
            <a-radio-button value="tired">😴 疲惫</a-radio-button>
            <a-radio-button value="sad">😢 难过</a-radio-button>
          </a-radio-group>
        </a-form-item>

        <a-form-item label="标签">
          <a-select
            v-model:value="form.tags"
            mode="tags"
            placeholder="添加标签，回车确认"
            :token-separators="[',']"
          />
        </a-form-item>

        <a-form-item label="参与人">
          <a-select
            v-model:value="form.participants"
            mode="tags"
            placeholder="添加参与人，回车确认"
          />
        </a-form-item>

        <a-form-item label="颜色">
          <input type="color" v-model="form.color" style="width: 40px; height: 32px; border: none; cursor: pointer" />
        </a-form-item>

        <a-form-item label="附件">
          <a-upload
            :file-list="fileList"
            :custom-request="handleUpload"
            :multiple="true"
            list-type="picture-card"
            @change="handleFileChange"
            accept="image/*,video/*,audio/*"
          >
            <div v-if="fileList.length < 9">
              <PlusOutlined />
              <div style="margin-top: 8px">上传</div>
            </div>
          </a-upload>
        </a-form-item>

        <!-- 元数据建议卡片（上传图片后自动出现） -->
        <a-form-item v-if="metadataSuggestions" label="智能建议" :wrapper-col="{ span: 20 }">
          <div class="metadata-suggestion-card">
            <p v-if="metadataSuggestions.hint_text" class="hint-text">💡 {{ metadataSuggestions.hint_text }}</p>
            <p v-if="metadataSuggestions.device_text" class="hint-text">📷 {{ metadataSuggestions.device_text }}</p>
            <div class="suggestion-actions" style="margin-top: 12px">
              <a-space>
                <a-button size="small" type="primary" @click="applyAllSuggestions">
                  <CheckOutlined /> 一键采纳所有建议
                </a-button>
                <a-button size="small" @click="applySuggestion('time')">
                  采纳时间
                </a-button>
                <a-button size="small" @click="applySuggestion('location')">
                  采纳地点
                </a-button>
                <a-button size="small" @click="applySuggestion('tags')">
                  采纳标签
                </a-button>
              </a-space>
            </div>
            <div class="suggestion-details" style="margin-top: 12px">
              <div v-if="metadataSuggestions.event_time" class="suggestion-item">
                <span class="suggestion-label">时间：</span>
                <span class="suggestion-value">{{ metadataSuggestions.event_time }}</span>
                <a-tag v-if="metadataSuggestions.confidence?.time" color="blue">
                  {{ Math.round(metadataSuggestions.confidence.time * 100) }}%
                </a-tag>
              </div>
              <div v-if="metadataSuggestions.location" class="suggestion-item">
                <span class="suggestion-label">地点：</span>
                <span class="suggestion-value">{{ metadataSuggestions.location }}</span>
                <a-tag v-if="metadataSuggestions.confidence?.location" color="blue">
                  {{ Math.round(metadataSuggestions.confidence.location * 100) }}%
                </a-tag>
              </div>
              <div v-if="metadataSuggestions.tags?.length" class="suggestion-item">
                <span class="suggestion-label">标签：</span>
                <span class="suggestion-value">
                  <a-tag v-for="tag in metadataSuggestions.tags" :key="tag">{{ tag }}</a-tag>
                </span>
              </div>
            </div>
          </div>
        </a-form-item>

        <a-form-item :wrapper-col="{ offset: 4, span: 16 }">
          <a-space>
            <a-button type="primary" html-type="submit" :loading="submitting">
              {{ isEdit ? '更新' : '创建' }}
            </a-button>
            <a-button @click="$router.back()">取消</a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useEventStore } from '@/stores/event'
import { mediaApi, type MetadataSuggestions } from '@/api'
import dayjs, { type Dayjs } from 'dayjs'
import { PlusOutlined, CheckOutlined } from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const eventStore = useEventStore()

const isEdit = computed(() => !!route.params.id)
const submitting = ref(false)
const fileList = ref<any[]>([])
const metadataSuggestions = ref<MetadataSuggestions | null>(null)

const form = ref({
  title: '',
  description: '',
  event_time: null as Dayjs | null,
  location: '',
  category: '',
  mood: '',
  tags: [] as string[],
  participants: [] as string[],
  color: '#1890ff',
  media_ids: [] as string[],
})

onMounted(async () => {
  if (isEdit.value) {
    const event = await eventStore.fetchEvent(route.params.id as string)
    if (event) {
      form.value = {
        title: event.title,
        description: event.description || '',
        event_time: dayjs(event.event_time),
        location: event.location || '',
        category: event.category || '',
        mood: event.mood || '',
        tags: event.tags || [],
        participants: event.participants || [],
        color: event.color || '#1890ff',
        media_ids: event.media?.map((m: any) => m.id) || [],
      }
    }
  }
})

async function handleUpload(options: any) {
  const { file, onSuccess, onError } = options
  try {
    const result = await mediaApi.uploadWithMetadata(file)
    const data = result.data as any
    const mediaData = data.data?.media || data.media
    const suggestions = data.data?.suggestions || data.suggestions

    if (mediaData) {
      form.value.media_ids.push(mediaData.id)
      onSuccess({ id: mediaData.id }, file)
    }

    // 显示元数据建议
    if (suggestions && Object.keys(suggestions).length > 0) {
      metadataSuggestions.value = suggestions
      message.success('已提取照片元数据，请查看智能建议')
    }
  } catch (err: any) {
    onError(err)
    message.error('上传失败')
  }
}

function handleFileChange({ fileList: newFileList }: any) {
  fileList.value = newFileList
}

function applyAllSuggestions() {
  if (!metadataSuggestions.value) return
  applySuggestion('time')
  applySuggestion('location')
  applySuggestion('tags')
  message.success('已采纳所有建议')
}

function applySuggestion(field: string) {
  if (!metadataSuggestions.value) return
  const sug = metadataSuggestions.value

  switch (field) {
    case 'time':
      if (sug.event_time) {
        form.value.event_time = dayjs(sug.event_time)
      }
      break
    case 'location':
      if (sug.location) {
        form.value.location = sug.location
      }
      break
    case 'tags':
      if (sug.tags) {
        const existing = new Set(form.value.tags)
        sug.tags.forEach((tag: string) => {
          if (!existing.has(tag)) {
            form.value.tags.push(tag)
          }
        })
      }
      break
  }
  message.success(`已采纳${field === 'time' ? '时间' : field === 'location' ? '地点' : '标签'}建议`)
}

async function handleSubmit() {
  submitting.value = true
  try {
    const input = {
      ...form.value,
      event_time: form.value.event_time?.toISOString() || '',
    }

    if (isEdit.value) {
      await eventStore.updateEvent(route.params.id as string, input)
      message.success('事件已更新')
    } else {
      await eventStore.createEvent(input)
      message.success('事件已创建')
    }
    router.push('/events')
  } catch (err: any) {
    message.error(err.message || '操作失败')
  } finally {
    submitting.value = false
  }
}
</script>
