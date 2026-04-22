import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export interface Template {
  id: number
  name: string
  category: string
  language: 'zh' | 'en'
  content: string
  /** user | ai | system */
  source?: string
  createdAt: string
  updatedAt: string
}

export const useTemplateStore = defineStore('template', () => {
  const templates = ref<Template[]>([])
  const loading = ref(false)

  async function fetchTemplates(category?: string, language?: 'zh' | 'en') {
    loading.value = true
    try {
      const params: Record<string, any> = {}
      if (category) params.category = category
      if (language) params.language = language
      const res: any = await request.get('/templates', { params })
      templates.value = res.data || []
      return templates.value
    } finally {
      loading.value = false
    }
  }

  async function fetchTemplate(id: number) {
    const res: any = await request.get(`/templates/${id}`)
    return res.data as Template
  }

  async function createTemplate(data: Partial<Template>) {
    const res: any = await request.post('/templates', data)
    return res.data as Template
  }

  async function updateTemplate(id: number, data: Partial<Template>) {
    const res: any = await request.put(`/templates/${id}`, data)
    return res.data as Template
  }

  async function deleteTemplate(id: number) {
    await request.delete(`/templates/${id}`)
    templates.value = templates.value.filter((t) => t.id !== id)
  }

  return { templates, loading, fetchTemplates, fetchTemplate, createTemplate, updateTemplate, deleteTemplate }
})
