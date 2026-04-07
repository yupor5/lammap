<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between mb-2">
      <h4 class="font-semibold text-gray-800">建议向客户确认以下信息</h4>
      <el-button size="small" @click="handleCopyAll">批量复制</el-button>
    </div>

    <div v-for="(item, index) in items" :key="index" class="border border-gray-200 rounded-lg p-3">
      <div class="flex items-start gap-3">
        <el-checkbox v-model="item.checked" class="mt-0.5" />
        <div class="flex-1">
          <p class="text-sm text-gray-800">{{ item.question }}</p>
          <p class="text-sm text-blue-600 mt-1">{{ item.questionEn }}</p>
        </div>
        <div class="flex gap-1 shrink-0">
          <el-button text size="small" @click="handleCopy(item.questionEn)">
            <el-icon><DocumentCopy /></el-icon>
          </el-button>
        </div>
      </div>
      <div class="flex gap-2 mt-2 ml-8">
        <el-tag
          :type="item.checked ? 'success' : 'info'"
          size="small"
          class="cursor-pointer"
          @click="item.checked = true"
        >
          已确认
        </el-tag>
        <el-tag size="small" class="cursor-pointer" type="info" @click="item.checked = true">
          不需确认
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { DocumentCopy } from '@element-plus/icons-vue'

const props = defineProps<{
  items: { question: string; questionEn: string; checked: boolean }[]
}>()

function handleCopy(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

function handleCopyAll() {
  const text = props.items.map((item) => item.questionEn).join('\n\n')
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制所有确认项')
}
</script>
