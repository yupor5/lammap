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
      <el-button text size="small">预览</el-button>
    </div>

    <el-button class="w-full" @click="handleGenerateEmailPack">生成邮件附件包</el-button>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { Download, Document } from '@element-plus/icons-vue'

defineProps<{
  attachments: { name: string; selected: boolean }[]
}>()

function handleDownloadAll() {
  ElMessage.info('打包下载功能开发中')
}

function handleGenerateEmailPack() {
  ElMessage.info('邮件附件包功能开发中')
}
</script>
