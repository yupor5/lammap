import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export interface ParsedParams {
  /** 后端解析结果结构版本（可选，未知字段前端忽略） */
  schemaVersion?: number
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
  remark?: string
}

export interface Quote {
  id?: number
  customerName: string
  status: string
  params: ParsedParams
  items: QuoteItem[]
  replyVersions: { title: string; content: string; language: string }[]
  confirmationList: { question: string; questionEn: string; checked: boolean }[]
  attachments: { name: string; url?: string; selected: boolean; source?: 'upload' | 'ai' | 'ai_generated' }[]
  templateMeta?: any
  renderedContents?: any
  totalAmount: number
  currency: string
  createdAt?: string
  updatedAt?: string
}

/** 与后端 quote_normalize.inferReplyLangFromContent 一致：标签语言跟正文走，避免 AI 写 language:"en" 却输出中文 */
export function inferReplyLangFromContent(s: string): 'zh' | 'en' {
  const t = s.trim()
  if (!t) return 'en'
  let cjk = 0
  for (const ch of t) {
    const cp = ch.codePointAt(0)!
    if (cp >= 0x4e00 && cp <= 0x9fff) cjk++
  }
  return cjk >= 2 ? 'zh' : 'en'
}

export function mapGeneratePayload(data: Record<string, unknown>, parsed: ParsedParams): Quote {
  const rows = (data.items as Record<string, unknown>[] | undefined) || []
  const items: QuoteItem[] = rows.map((row) => {
    const qty = Number(row.quantity) || 0
    const up = Number(row.unitPrice ?? row.unit_price) || 0
    let tp = Number(row.totalPrice ?? row.amount ?? row.total_price) || 0
    if (!tp && qty && up) tp = Math.round(qty * up * 100) / 100
    return {
      productName: String(row.productName ?? row.product_name ?? ''),
      model: String(row.model ?? ''),
      specs: String(row.specs ?? row.spec ?? ''),
      quantity: qty,
      unitPrice: up,
      totalPrice: tp,
      remark: String(row.remark ?? ''),
    }
  })
  let replyVersions = data.replyVersions as Quote['replyVersions']
  const rv = data.replyVersions as Record<string, string> | Quote['replyVersions']
  if (rv && !Array.isArray(rv) && typeof rv === 'object') {
    const short = String(rv.short ?? '')
    const prof = String(rv.professional ?? '')
    const follow = String(rv.followup ?? '')
    replyVersions = [
      { title: '简短成交版 (WhatsApp/微信)', content: short, language: inferReplyLangFromContent(short) },
      { title: '专业邮件版', content: prof, language: inferReplyLangFromContent(prof) },
      { title: '追单版', content: follow, language: inferReplyLangFromContent(follow) },
    ]
  }
  if (!replyVersions || !Array.isArray(replyVersions)) {
    replyVersions = []
  }
  replyVersions = replyVersions.map((r) => ({
    ...r,
    language: inferReplyLangFromContent(String(r.content ?? '')),
  }))
  const confirmationList = (data.confirmationList as Quote['confirmationList']) || []
  const attachments = (data.attachments as Quote['attachments']) || []
  const totalAmount = Number(data.totalAmount) || items.reduce((s, i) => s + (i.totalPrice || 0), 0)
  return {
    customerName: String(data.customerName ?? parsed.customerName ?? ''),
    status: String(data.status ?? '草稿'),
    params: parsed,
    items,
    replyVersions,
    confirmationList,
    attachments,
    totalAmount,
    currency: String(data.currency ?? parsed.currency ?? 'USD'),
  }
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
      const mapped = mapGeneratePayload(res.data as Record<string, unknown>, params)
      currentQuote.value = mapped
      return mapped
    } finally {
      isLoading.value = false
    }
  }

  async function saveQuote(quote: Record<string, unknown>) {
    const id = quote.id as number | undefined
    const res: any = id
      ? await request.put(`/quotes/${id}`, quote)
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
