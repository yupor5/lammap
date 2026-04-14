<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between mb-2">
      <h4 class="font-semibold text-gray-800">推荐附带发送的资料</h4>
      <el-button size="small" type="primary" :disabled="loading" @click="handleDownloadAll">
        <el-icon><Download /></el-icon>
        打包下载
      </el-button>
    </div>

    <div
      v-for="(att, index) in attachments"
      :key="`${index}-${att.name}`"
      class="flex items-center gap-3 p-3 border border-gray-200 rounded-lg transition-colors"
      :class="String(att.url || '').trim() ? 'hover:bg-gray-50' : 'bg-gray-50 opacity-60'"
    >
      <el-checkbox v-model="att.selected" :disabled="loading" />
      <el-icon :size="20" class="text-gray-400"><Document /></el-icon>
      <span class="flex-1 text-sm text-gray-700">{{ att.name }}</span>
      <el-tag v-if="isGenerating(index)" size="small" type="warning">生成中</el-tag>
      <el-tag v-else-if="att.source === 'upload'" size="small" type="success">上传</el-tag>
      <el-tag v-else-if="att.source === 'ai_generated'" size="small" type="primary">AI已生成</el-tag>
      <el-tag v-else size="small" type="info">AI建议</el-tag>
      <el-button
        v-if="canAiGenerate(att)"
        text
        size="small"
        :disabled="loading || !att.selected || isGenerating(index)"
        @click="emit('ai-generate', att, index)"
      >
        AI生成
      </el-button>
      <el-button
        text
        size="small"
        :disabled="!String(att.url || '').trim() || loading"
        @click="handlePreview(att)"
      >
        预览
      </el-button>
    </div>

    <el-button class="w-full" :loading="loading" :disabled="loading" @click="handleGenerateEmailPack">
      生成邮件附件包
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { Download, Document } from '@element-plus/icons-vue'

const props = defineProps<{
  attachments: { name: string; url?: string; selected: boolean; source?: 'upload' | 'ai' | 'ai_generated' }[]
  loading?: boolean
  /** 正在 AI 生成中的行下标（与父组件 Set 同步，按行隔离并行任务） */
  generatingIndices?: number[]
}>()

function isGenerating(index: number) {
  return (props.generatingIndices || []).includes(index)
}

const emit = defineEmits<{
  (e: 'download-all'): void
  (e: 'generate-pack'): void
  (
    e: 'ai-generate',
    att: { name: string; url?: string; selected: boolean; source?: string },
    index: number
  ): void
}>()

function canAiGenerate(att: { url?: string; source?: string }) {
  return !String(att.url || '').trim() && String(att.source || '') === 'ai'
}

function handleDownloadAll() {
  if (!props.attachments?.length) {
    ElMessage.warning('暂无附件可打包')
    return
  }
  emit('download-all')
}

function handleGenerateEmailPack() {
  if (props.loading) return
  if (!props.attachments?.length) {
    ElMessage.warning('暂无附件可生成')
    return
  }
  emit('generate-pack')
}

function handlePreview(att: { name: string; url?: string }) {
  const url = String(att.url || '').trim()
  if (!url) {
    return
  }
  // 直接新窗口打开：pdf/图片会预览，其它类型由浏览器处理下载/打开
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>
