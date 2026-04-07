<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between mb-2">
      <h4 class="font-semibold text-gray-800">回复话术</h4>
      <el-select v-model="tone" size="small" class="w-28">
        <el-option label="专业" value="professional" />
        <el-option label="热情" value="friendly" />
        <el-option label="简洁" value="concise" />
        <el-option label="催单" value="urgent" />
      </el-select>
    </div>

    <div v-for="(version, index) in versions" :key="index" class="border border-gray-200 rounded-lg overflow-hidden">
      <div class="flex items-center justify-between bg-gray-50 px-4 py-2">
        <div class="flex items-center gap-2">
          <span class="font-medium text-sm text-gray-700">{{ version.title }}</span>
          <el-tag size="small" type="info">{{ version.language === 'en' ? 'English' : '中文' }}</el-tag>
        </div>
        <div class="flex gap-1">
          <el-button text size="small" @click="handleCopy(version.content)">
            <el-icon><DocumentCopy /></el-icon>
            复制
          </el-button>
          <el-button text size="small">
            <el-icon><RefreshRight /></el-icon>
            重新生成
          </el-button>
        </div>
      </div>
      <div class="p-4 text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">
        {{ version.content }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, RefreshRight } from '@element-plus/icons-vue'

defineProps<{
  versions: { title: string; content: string; language: string }[]
}>()

const tone = ref('professional')

function handleCopy(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}
</script>
