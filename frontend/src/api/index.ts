import api from './client'

// ==================== 认证 API ====================

export interface LoginInput {
  username: string
  password: string
}

export interface RegisterInput {
  username: string
  email: string
  password: string
  nickname?: string
}

export interface AuthResult {
  user: UserInfo
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface UserInfo {
  id: string
  username: string
  email: string
  nickname: string
  avatar: string
  roles: string[]
  permissions: string[]
}

export const authApi = {
  login: (data: LoginInput) => api.post<AuthResult>('/auth/login', data),
  register: (data: RegisterInput) => api.post<AuthResult>('/auth/register', data),
  refresh: (refreshToken: string) => api.post<AuthResult>('/auth/refresh', { refresh_token: refreshToken }),
  me: () => api.get<UserInfo>('/auth/me'),
  logout: () => api.post('/auth/logout'),
}

// ==================== 事件 API ====================

export interface Event {
  id: string
  user_id: string
  title: string
  description: string
  event_time: string
  end_time?: string
  location: string
  latitude?: number
  longitude?: number
  mood: string
  tags: string[]
  participants: string[]
  category: string
  color: string
  priority: number
  is_confirmed: boolean
  source: string
  source_data: Record<string, unknown>
  media: Media[]
  created_at: string
  updated_at: string
}

export interface Media {
  id: string
  type: string
  storage_path: string
  thumbnail_path: string
  original_name: string
  mime_type: string
  size: number
  width?: number
  height?: number
  media_metadata?: MediaMetadataResult
}

export interface CreateEventInput {
  title: string
  description?: string
  event_time: string
  end_time?: string
  location?: string
  latitude?: number
  longitude?: number
  mood?: string
  tags?: string[]
  participants?: string[]
  category?: string
  color?: string
  priority?: number
  media_ids?: string[]
  is_confirmed?: boolean
  source?: string
  source_data?: Record<string, unknown>
}

export interface EventListParams {
  page?: number
  page_size?: number
  start_time?: string
  end_time?: string
  category?: string
  mood?: string
  keyword?: string
  confirmed?: boolean
  sort?: string
  order?: string
}

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export const eventApi = {
  list: (params: EventListParams) => api.get<PageResult<Event>>('/events', { params }),
  get: (id: string) => api.get<Event>(`/events/${id}`),
  create: (data: CreateEventInput) => api.post<Event>('/events', data),
  update: (id: string, data: Partial<CreateEventInput>) => api.put<Event>(`/events/${id}`, data),
  delete: (id: string) => api.delete(`/events/${id}`),
  confirm: (id: string, modifications?: Record<string, unknown>) =>
    api.post(`/events/confirm/${id}`, { modifications }),
  calendar: (year: number, month: number) =>
    api.get(`/events/calendar/${year}/${month}`),
  applySuggestions: (eventId: string, mediaId: string, fields: string[]) =>
    api.post(`/events/${eventId}/apply-suggestions`, { media_id: mediaId, fields }),
}

// ==================== 媒体 API ====================

export interface UploadResult {
  media: Media
  metadata?: MediaMetadataResult
  suggestions?: MetadataSuggestions
}

export interface MediaMetadataResult {
  id: string
  media_id: string
  shot_time?: string
  shot_time_src: string
  location_name: string
  location_src: string
  camera_make: string
  camera_model: string
  lens_model: string
  focal_length?: number
  aperture?: number
  shutter_speed: string
  iso?: number
  video_codec: string
  audio_codec: string
  frame_rate?: number
  bitrate?: number
  iptc_keywords: string[]
  iptc_caption: string
  filename_pattern: string
  raw_exif: Record<string, unknown>
  suggestions: MetadataSuggestions
  extract_status: string
  extract_error: string
}

export interface MetadataSuggestions {
  event_time?: string
  location?: string
  tags?: string[]
  category?: string
  hint_text?: string
  device_text?: string
  scene_text?: string
  confidence?: Record<string, number>
}

export const mediaApi = {
  upload: (file: File, eventId?: string) => {
    const formData = new FormData()
    formData.append('file', file)
    if (eventId) formData.append('event_id', eventId)
    return api.post<Media>('/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  uploadWithMetadata: (file: File, eventId?: string) => {
    const formData = new FormData()
    formData.append('file', file)
    if (eventId) formData.append('event_id', eventId)
    return api.post<UploadResult>('/media/upload-with-metadata', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  list: (params: { page?: number; page_size?: number; type?: string }) =>
    api.get<PageResult<Media>>('/media', { params }),
  get: (id: string) => api.get<Media>(`/media/${id}`),
  getMetadata: (id: string) => api.get<MediaMetadataResult>(`/media/${id}/metadata`),
  getFileUrl: (id: string) => `/api/v1/media/${id}/file`,
  delete: (id: string) => api.delete(`/media/${id}`),
}

// ==================== AI API ====================

export const aiApi = {
  transcribe: (file: File, language?: string) => {
    const formData = new FormData()
    formData.append('file', file)
    if (language) formData.append('language', language)
    return api.post('/ai/transcribe', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  extractEvent: (text: string, conversationHistory?: Array<{ role: string; content: string }>) =>
    api.post('/ai/extract-event', { text, conversation_history: conversationHistory }),
  chat: (message: string, conversationId?: string) =>
    api.post('/ai/chat', { message, conversation_id: conversationId }),
}

// ==================== 配置 API ====================

export const configApi = {
  get: (category?: string) => api.get('/config', { params: { category } }),
  update: (configs: Array<{ category: string; key: string; value: string }>) =>
    api.put('/config', { configs }),
}

// ==================== Webhook API ====================

export interface Webhook {
  id: string
  name: string
  url: string
  events: string[]
  is_active: boolean
  retry_count: number
  timeout_ms: number
  created_at: string
}

export const webhookApi = {
  list: () => api.get<Webhook[]>('/webhooks'),
  create: (data: Partial<Webhook> & { url: string; events: string[] }) =>
    api.post<Webhook>('/webhooks', data),
  update: (id: string, data: Partial<Webhook>) => api.put<Webhook>(`/webhooks/${id}`, data),
  delete: (id: string) => api.delete(`/webhooks/${id}`),
  test: (id: string) => api.post(`/webhooks/${id}/test`),
  deliveries: (id: string, page?: number) =>
    api.get(`/webhooks/${id}/deliveries`, { params: { page } }),
}

// ==================== 系统 API ====================

export const systemApi = {
  health: () => api.get('/system/health'),
  backup: (options?: { type?: string; scope?: Record<string, unknown> }) =>
    api.post('/system/backup', options),
  backups: () => api.get('/system/backups'),
  restore: (backupId: string, options?: Record<string, unknown>) =>
    api.post('/system/restore', { backup_id: backupId, options }),
}

// ==================== 安装 API ====================

export const setupApi = {
  status: () => api.get('/setup/status'),
  init: (data: { email: string; password: string }) => api.post('/setup/init', data),
}

// ==================== 管理 API ====================

export const adminApi = {
  users: (params?: { page?: number; keyword?: string }) =>
    api.get('/admin/users', { params }),
  updateUser: (id: string, data: Record<string, unknown>) =>
    api.put(`/admin/users/${id}`, data),
  roles: () => api.get('/admin/roles'),
  permissions: () => api.get('/admin/permissions'),
  auditLogs: (params?: { page?: number; user_id?: string; action?: string }) =>
    api.get('/admin/audit-logs', { params }),
}
