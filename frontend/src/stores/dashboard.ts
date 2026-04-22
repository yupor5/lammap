import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export interface DashboardStats {
  todayQuotes: number
  pendingParams: number
  toSend: number
  weekSent: number
}

export interface RecentQuote {
  id: number
  quoteNumber: string
  customerName: string
  currency: string
  totalAmount: number
  status: string
  updatedAt: string
  items: { productName: string }[]
}

export const useDashboardStore = defineStore('dashboard', () => {
  const stats = ref<DashboardStats>({ todayQuotes: 0, pendingParams: 0, toSend: 0, weekSent: 0 })
  const recentQuotes = ref<RecentQuote[]>([])
  const loading = ref(false)

  async function fetchStats() {
    try {
      const res: any = await request.get('/dashboard/stats')
      stats.value = res.data
    } catch {
      // keep defaults
    }
  }

  async function fetchRecent() {
    try {
      const res: any = await request.get('/dashboard/recent')
      recentQuotes.value = res.data || []
    } catch {
      // keep empty
    }
  }

  async function fetchAll() {
    loading.value = true
    try {
      await Promise.all([fetchStats(), fetchRecent()])
    } finally {
      loading.value = false
    }
  }

  return { stats, recentQuotes, loading, fetchStats, fetchRecent, fetchAll }
})
