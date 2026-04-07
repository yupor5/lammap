import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export interface ParsedParams {
  customerName: string
  country: string
  currency: string
  deliveryAddress: string
  productName: string
  model: string
  material: string
  size: string
  color: string
  quantity: number
  packaging: string
  moq: number
  paymentTerms: string
  leadTime: string
  validityPeriod: string
  includeShipping: boolean | null
  unconfirmed: string[]
}

export interface QuoteItem {
  productName: string
  model: string
  specs: string
  quantity: number
  unitPrice: number
  totalPrice: number
}

export interface Quote {
  id?: number
  customerName: string
  status: string
  params: ParsedParams
  items: QuoteItem[]
  replyVersions: { title: string; content: string; language: string }[]
  confirmationList: { question: string; questionEn: string; checked: boolean }[]
  attachments: { name: string; selected: boolean }[]
  totalAmount: number
  currency: string
  createdAt?: string
  updatedAt?: string
}

export const useQuoteStore = defineStore('quote', () => {
  const currentQuote = ref<Quote | null>(null)
  const parsedParams = ref<ParsedParams | null>(null)
  const isLoading = ref(false)

  async function parseRequirement(content: string, files?: File[]) {
    isLoading.value = true
    try {
      const formData = new FormData()
      formData.append('content', content)
      if (files) {
        files.forEach((f) => formData.append('files', f))
      }
      const res: any = await request.post('/quotes/parse', formData)
      parsedParams.value = res.data
      return res.data
    } finally {
      isLoading.value = false
    }
  }

  async function generateQuote(params: ParsedParams) {
    isLoading.value = true
    try {
      const res: any = await request.post('/quotes/generate', params)
      currentQuote.value = res.data
      return res.data
    } finally {
      isLoading.value = false
    }
  }

  async function saveQuote(quote: Quote) {
    const res: any = quote.id
      ? await request.put(`/quotes/${quote.id}`, quote)
      : await request.post('/quotes', quote)
    return res.data
  }

  async function fetchQuotes(params?: Record<string, any>) {
    const res: any = await request.get('/quotes', { params })
    return res.data
  }

  async function fetchQuote(id: number) {
    const res: any = await request.get(`/quotes/${id}`)
    currentQuote.value = res.data
    return res.data
  }

  function reset() {
    currentQuote.value = null
    parsedParams.value = null
  }

  return { currentQuote, parsedParams, isLoading, parseRequirement, generateQuote, saveQuote, fetchQuotes, fetchQuote, reset }
})
