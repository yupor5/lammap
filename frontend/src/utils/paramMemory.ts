export type MemoryKey =
  | 'customerName'
  | 'country'
  | 'currency'
  | 'deliveryAddress'
  | 'productName'
  | 'model'
  | 'material'
  | 'size'
  | 'color'
  | 'packaging'
  | 'paymentTerms'
  | 'leadTime'
  | 'validityPeriod'
  | 'productExampleHint'

export interface MemoryItem {
  value: string
  updatedAt: number
}

const STORAGE_KEY = 'lammap-param-memory-v1'
const MAX_ITEMS_PER_KEY = 30

/** 用于展示与存储：去掉首尾空白，并把连续空白压成单个空格 */
function normalizeDisplayValue(raw: unknown): string {
  return String(raw ?? '')
    .trim()
    .replace(/\s+/g, ' ')
}

/** 用于去重：在 normalizeDisplayValue 基础上忽略大小写 */
function normalizeForDedupe(s: string): string {
  return normalizeDisplayValue(s).toLowerCase()
}

function sortNewestFirst(items: MemoryItem[]): MemoryItem[] {
  return [...items].sort((a, b) => b.updatedAt - a.updatedAt)
}

/** 同一条说明只保留一条：取 updatedAt 最新的一条（已按新在前排序后去重） */
function dedupeNewestFirst(items: MemoryItem[]): MemoryItem[] {
  const sorted = sortNewestFirst(items)
  const out: MemoryItem[] = []
  const seen = new Set<string>()
  for (const it of sorted) {
    const k = normalizeForDedupe(it.value)
    if (!k) continue
    if (seen.has(k)) continue
    seen.add(k)
    out.push(it)
  }
  return out
}

function normalizeQuery(q: string) {
  return (q || '').trim().toLowerCase().replace(/\s+/g, '')
}

function safePinyin(s: string) {
  try {
    // 运行时依赖：构建侧无需 Node 的 require 类型
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const g: any = globalThis as any
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const mod: any = g.__lammapPinyinPro || null
    const pinyinFn = mod?.pinyin as undefined | ((text: string, opt: any) => string[] | string)
    if (!pinyinFn) return ''
    const out = pinyinFn(String(s || ''), { toneType: 'none', type: 'array' })
    const arr = Array.isArray(out) ? out : String(out).split(/\s+/)
    return normalizeQuery(arr.join(''))
  } catch {
    return ''
  }
}

function loadAll(): Record<string, MemoryItem[]> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as Record<string, MemoryItem[]>
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

function saveAll(all: Record<string, MemoryItem[]>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(all))
  } catch {
    // ignore
  }
}

export function rememberValue(key: MemoryKey, raw: unknown) {
  const v = normalizeDisplayValue(raw)
  if (!v) return
  const now = Date.now()
  const all = loadAll()
  const list = Array.isArray(all[key]) ? [...all[key]] : []
  const norm = normalizeForDedupe(v)
  const existingIdx = list.findIndex((x) => normalizeForDedupe(x.value) === norm)
  if (existingIdx >= 0) {
    list[existingIdx] = { value: v, updatedAt: now }
  } else {
    list.unshift({ value: v, updatedAt: now })
  }
  all[key] = dedupeNewestFirst(list).slice(0, MAX_ITEMS_PER_KEY)
  saveAll(all)
}

export function getSuggestions(key: MemoryKey, query: string): string[] {
  const q = normalizeQuery(query)
  const all = loadAll()
  const list = Array.isArray(all[key]) ? all[key] : []
  const sorted = dedupeNewestFirst(list)
  const filtered = q
    ? sorted.filter((x) => {
        const v = String(x.value || '')
        const vNorm = normalizeQuery(v)
        if (vNorm.includes(q)) return true
        const py = safePinyin(v)
        return py && py.includes(q)
      })
    : sorted
  return filtered.map((x) => x.value)
}

