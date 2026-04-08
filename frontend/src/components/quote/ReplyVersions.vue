<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between mb-2">
      <h4 class="font-semibold text-gray-800">回复话术</h4>
      <div class="flex items-center gap-2">
        <el-select v-model="language" size="small" class="w-24" @change="handleRegenerateAll">
          <el-option label="English" value="en" />
          <el-option label="中文" value="zh" />
        </el-select>
        <el-select v-model="tone" size="small" class="w-28" @change="handleRegenerateAll">
          <el-option label="专业" value="professional" />
          <el-option label="热情" value="friendly" />
          <el-option label="简洁" value="concise" />
          <el-option label="催单" value="urgent" />
        </el-select>
      </div>
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
          <el-button text size="small" :loading="regeneratingIndex === index" @click="handleRegenerate(index)">
            <el-icon><RefreshRight /></el-icon>
            重新生成
          </el-button>
        </div>
      </div>
      <div class="p-4 text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">
        {{ version.content }}
      </div>
    </div>

    <div v-if="regeneratingAll" class="text-center py-4">
      <el-icon class="is-loading" :size="20"><Loading /></el-icon>
      <span class="ml-2 text-sm text-gray-500">正在根据新语气/语言重新生成...</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, RefreshRight, Loading } from '@element-plus/icons-vue'
import request from '@/utils/request'

const props = defineProps<{
  versions: { title: string; content: string; language: string }[]
  parsedParams?: any
}>()

const emit = defineEmits<{
  'update:versions': [versions: { title: string; content: string; language: string }[]]
}>()

const tone = ref('professional')
const language = ref('en')
const regeneratingIndex = ref(-1)
const regeneratingAll = ref(false)

function handleCopy(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

async function handleRegenerate(index: number) {
  if (!props.parsedParams) {
    ElMessage.warning('缺少解析参数，无法重新生成')
    return
  }

  regeneratingIndex.value = index
  try {
    const res: any = await request.post('/quotes/generate', {
      ...props.parsedParams,
      _replyTone: tone.value,
      _replyLanguage: language.value,
      _replyIndex: index,
    })

    if (res.data?.replyVersions?.[index]) {
      const updated = [...props.versions]
      updated[index] = res.data.replyVersions[index]
      emit('update:versions', updated)
      ElMessage.success('已重新生成')
    }
  } catch {
    ElMessage.info('重新生成失败（后端未连接时为正常现象）')
  } finally {
    regeneratingIndex.value = -1
  }
}

async function handleRegenerateAll() {
  if (!props.parsedParams) return

  regeneratingAll.value = true
  try {
    const res: any = await request.post('/quotes/generate', {
      ...props.parsedParams,
      _replyTone: tone.value,
      _replyLanguage: language.value,
    })

    if (res.data?.replyVersions) {
      emit('update:versions', res.data.replyVersions)
      ElMessage.success(`已按 ${tone.value} 语气、${language.value === 'zh' ? '中文' : 'English'} 重新生成`)
    }
  } catch {
    ElMessage.info('重新生成失败（后端未连接时为正常现象）')
  } finally {
    regeneratingAll.value = false
  }
}
</script>
