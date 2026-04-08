<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between mb-2">
      <h4 class="font-semibold text-gray-800">推荐附带发送的资料</h4>
      <el-button size="small" type="primary" @click="handleDownloadAll">
        <el-icon><Download /></el-icon>
        打包下载
      </el-button>
    </div>

    <div v-for="(att, index) in attachments" :key="index"
      class="flex items-center gap-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
    >
      <el-checkbox v-model="att.selected" />
      <el-icon :size="20" class="text-gray-400"><Document /></el-icon>
      <span class="flex-1 text-sm text-gray-700">{{ att.name }}</span>
      <el-button text size="small" @click="handlePreview(att)">预览</el-button>
    </div>

    <el-button class="w-full" @click="handleGenerateEmailPack">生成邮件附件包</el-button>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { Download, Document } from '@element-plus/icons-vue'

const props = defineProps<{
  attachments: { name: string; url?: string; selected: boolean }[]
}>()

function handleDownloadAll() {
  ElMessage.info('打包下载功能开发中')
}

function handleGenerateEmailPack() {
  ElMessage.info('邮件附件包功能开发中')
}

function handlePreview(att: { name: string; url?: string }) {
  const url = String(att.url || '').trim()
  if (!url) {
    ElMessage.warning('暂无可预览链接（请先上传附件或在生成结果中提供 url）')
    return
  }
  // 直接新窗口打开：pdf/图片会预览，其它类型由浏览器处理下载/打开
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>
