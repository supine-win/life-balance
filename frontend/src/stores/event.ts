import { defineStore } from 'pinia'
import { ref } from 'vue'
import { eventApi, type Event, type EventListParams, type CreateEventInput } from '@/api'

export const useEventStore = defineStore('event', () => {
  const events = ref<Event[]>([])
  const currentEvent = ref<Event | null>(null)
  const total = ref(0)
  const loading = ref(false)
  const filters = ref<EventListParams>({
    page: 1,
    page_size: 20,
    sort: 'event_time',
    order: 'desc',
  })

  async function fetchEvents(params?: EventListParams) {
    loading.value = true
    try {
      const query = { ...filters.value, ...params }
      const { data } = await eventApi.list(query)
      const result = (data as any).data || data
      events.value = result.items || []
      total.value = result.total || 0
      return result
    } finally {
      loading.value = false
    }
  }

  async function fetchEvent(id: string) {
    loading.value = true
    try {
      const { data } = await eventApi.get(id)
      currentEvent.value = (data as any).data || data
      return currentEvent.value
    } finally {
      loading.value = false
    }
  }

  async function createEvent(input: CreateEventInput) {
    const { data } = await eventApi.create(input)
    const event = (data as any).data || data
    events.value.unshift(event)
    return event
  }

  async function updateEvent(id: string, input: Partial<CreateEventInput>) {
    const { data } = await eventApi.update(id, input)
    const updated = (data as any).data || data
    const idx = events.value.findIndex((e) => e.id === id)
    if (idx >= 0) events.value[idx] = updated
    if (currentEvent.value?.id === id) currentEvent.value = updated
    return updated
  }

  async function deleteEvent(id: string) {
    await eventApi.delete(id)
    events.value = events.value.filter((e) => e.id !== id)
    if (currentEvent.value?.id === id) currentEvent.value = null
  }

  async function confirmEvent(id: string, modifications?: Record<string, unknown>) {
    const { data } = await eventApi.confirm(id, modifications)
    const updated = (data as any).data || data
    const idx = events.value.findIndex((e) => e.id === id)
    if (idx >= 0) events.value[idx] = updated
    return updated
  }

  return {
    events,
    currentEvent,
    total,
    loading,
    filters,
    fetchEvents,
    fetchEvent,
    createEvent,
    updateEvent,
    deleteEvent,
    confirmEvent,
  }
})
