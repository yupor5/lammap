<template>
  <div class="space-y-4 min-h-0">
    <div class="flex items-center justify-between mb-2 flex-wrap gap-2">
      <h4 class="font-semibold text-gray-800">回复话术</h4>
      <div class="flex items-center gap-2 flex-wrap">
        <el-select v-model="language" size="small" class="w-24" placeholder="语言">
          <el-option label="English" value="en" />
          <el-option label="中文" value="zh" />
        </el-select>
        <el-select v-model="tone" size="small" class="w-28" placeholder="语气">
          <el-option label="专业" value="professional" />
          <el-option label="热情" value="friendly" />
          <el-option label="简洁" value="concise" />
          <el-option label="催单" value="urgent" />
        </el-select>
        <el-button type="primary" size="small" :loading="generatingReplies" :disabled="!parsedParams" @click="handleAiGenerateAll">
          AI 生成话术
        </el-button>
      </div>
    </div>

    <div v-for="(version, index) in versions" :key="index" class="border border-gray-200 rounded-lg flex flex-col min-h-0">
      <div class="flex items-center justify-between bg-gray-50 px-4 py-2 shrink-0">
        <div class="flex items-center gap-2">
          <span class="font-medium text-sm text-gray-700">{{ version.title }}</span>
          <el-tag size="small" type="info">{{ version.language === 'en' ? 'English' : '中文' }}</el-tag>
        </div>
        <div class="flex gap-1">
          <el-button text size="small" @click="handleCopy(version.content)">
            <el-icon><DocumentCopy /></el-icon>
            复制
          </el-button>
          <el-button text size="small" :loading="regeneratingIndex === index" :disabled="!parsedParams" @click="handleRegenerate(index)">
            <el-icon><RefreshRight /></el-icon>
            重新生成
          </el-button>
        </div>
      </div>
      <div class="p-4 text-sm text-gray-700 whitespace-pre-wrap leading-relaxed break-words min-w-0">
        {{ version.content }}
      </div>
    </div>

    <div v-if="generatingReplies" class="text-center py-4">
      <el-icon class="is-loading" :size="20"><Loading /></el-icon>
      <span class="ml-2 text-sm text-gray-500">AI 生成任务进行中，请稍候…</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, RefreshRight, Loading } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { inferReplyLangFromContent } from '@/stores/quote'

const props = defineProps<{
  versions: { title: string; content: string; language: string }[]
  parsedParams?: Record<string, unknown> | null
}>()

const emit = defineEmits<{
  'update:versions': [versions: { title: string; content: string; language: string }[]]
}>()

const tone = ref('professional')
const language = ref<'zh' | 'en'>('en')
const regeneratingIndex = ref(-1)
const generatingReplies = ref(false)

watch(
  () => props.versions,
  (v) => {
    const lang = v?.[0]?.language
    if (lang === 'zh' || lang === 'en') {
      language.value = lang
    }
  },
  { immediate: true, deep: true }
)

function handleCopy(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

/** 与 NewQuote 中 pollGenerateJob 一致：轮询报价生成异步任务 */
async function pollQuoteGenerateJob(jobId: number): Promise<Record<string, unknown>> {
  const start = Date.now()
  let interval = 1000
  while (true) {
    // eslint-disable-next-line no-await-in-loop
    const res: any = await request.get(`/quotes/generate-jobs/${jobId}`)
    const status = String(res.data?.status || '')
    if (status === 'succeeded') {
      return res.data as Record<string, unknown>
    }
    if (status === 'failed') {
      throw new Error(String(res.data?.errorMsg || '生成失败'))
    }
    const elapsed = Date.now() - start
    if (elapsed > 5 * 60 * 1000) {
      throw new Error('生成超时（超过 5 分钟仍未完成）')
    }
    if (elapsed > 10000 && interval < 2000) interval = 2000
    if (elapsed > 30000 && interval < 3000) interval = 3000
    if (elapsed > 60000 && interval < 5000) interval = 5000
    // eslint-disable-next-line no-await-in-loop
    await new Promise((r) => setTimeout(r, interval))
  }
}

function mapReplyVersionsFromResult(rawResult: Record<string, unknown> | null): { title: string; content: string; language: string }[] {
  if (!rawResult) return []
  const rv = rawResult.replyVersions
  if (!Array.isArray(rv) || rv.length === 0) return []
  return rv
    .map((row: unknown) => {
      const r = row as Record<string, unknown>
      return {
        title: String(r.title ?? ''),
        content: String(r.content ?? ''),
        language: inferReplyLangFromContent(String(r.content ?? '')),
      }
    })
    .filter((x) => x.content.trim() !== '')
}

async function fetchReplyVersionsFromJob(jobId: number): Promise<{ title: string; content: string; language: string }[]> {
  const jobData = await pollQuoteGenerateJob(jobId)
  let rawResult = jobData?.result as Record<string, unknown> | null
  if (!rawResult && typeof jobData?.resultJson === 'string' && (jobData.resultJson as string).trim()) {
    try {
      rawResult = JSON.parse(jobData.resultJson as string) as Record<string, unknown>
    } catch {
      rawResult = null
    }
  }
  const mapped = mapReplyVersionsFromResult(rawResult)
  if (mapped.length === 0) {
    throw new Error('未获取到回复话术')
  }
  return mapped
}

function buildGeneratePayload(extra: Record<string, unknown> = {}) {
  return {
    ...(props.parsedParams || {}),
    _replyTone: tone.value,
    _replyLanguage: language.value,
    ...extra,
  }
}

async function handleAiGenerateAll() {
  if (!props.parsedParams) {
    ElMessage.warning('请先解析客户需求')
    return
  }
  generatingReplies.value = true
  try {
    const created: any = await request.post('/quotes/generate-jobs', buildGeneratePayload())
    const jobId = Number(created.data?.jobId)
    if (!jobId) throw new Error('创建生成任务失败')
    const next = await fetchReplyVersionsFromJob(jobId)
    emit('update:versions', next)
    ElMessage.success(`已按「${toneLabel(tone.value)}」语气、${language.value === 'zh' ? '中文' : 'English'} 生成 3 条话术`)
  } catch (e) {
    console.error(e)
    ElMessage.error(e instanceof Error ? e.message : '生成失败，请稍后重试')
  } finally {
    generatingReplies.value = false
  }
}

function toneLabel(t: string) {
  const m: Record<string, string> = {
    professional: '专业',
    friendly: '热情',
    concise: '简洁',
    urgent: '催单',
  }
  return m[t] || t
}

async function handleRegenerate(index: number) {
  if (!props.parsedParams) {
    ElMessage.warning('缺少解析参数，无法重新生成')
    return
  }

  regeneratingIndex.value = index
  try {
    const created: any = await request.post(
      '/quotes/generate-jobs',
      buildGeneratePayload({ _replyIndex: index })
    )
    const jobId = Number(created.data?.jobId)
    if (!jobId) throw new Error('创建任务失败')
    const full = await fetchReplyVersionsFromJob(jobId)
    emit('update:versions', full)
    ElMessage.success('已重新生成')
  } catch (e) {
    console.error(e)
    ElMessage.error(e instanceof Error ? e.message : '重新生成失败')
  } finally {
    regeneratingIndex.value = -1
  }
}
</script>
