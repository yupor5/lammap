<template>
  <div class="flex items-center justify-between p-2 rounded bg-gray-50 hover:bg-gray-100 transition-colors">
    <span class="text-sm text-gray-600 shrink-0 w-20">{{ label }}</span>
    <div class="flex-1 ml-2">
      <el-autocomplete
        v-if="isEditing && !!memoryKey && !numeric"
        v-model="localValue"
        size="small"
        :fetch-suggestions="querySearch"
        @select="handleSelect"
        @blur="handleConfirm"
        @keyup.enter="handleConfirm"
        ref="inputRef"
        placeholder="支持搜索历史输入…"
        style="width: 100%"
      />
      <el-input
        v-else-if="isEditing"
        v-model="localValue"
        size="small"
        @blur="handleConfirm"
        @keyup.enter="handleConfirm"
        ref="inputRef"
      />
      <div
        v-else
        class="flex items-center gap-1 cursor-pointer"
        @click="handleClickDisplay"
      >
        <span v-if="modelValue" class="text-sm text-gray-800">{{ modelValue }}</span>
        <el-tag v-else type="warning" size="small">未识别（点击填写）</el-tag>
        <el-button text size="small" :icon="Edit" @click.stop="startEdit" class="ml-auto" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { Edit } from '@element-plus/icons-vue'
import { getSuggestions, type MemoryKey } from '@/utils/paramMemory'

const props = defineProps<{
  label: string
  modelValue: string | number
  /** 确认时以数字写回（数量、MOQ 等） */
  numeric?: boolean
  /** 开启输入记忆（下拉可搜索，按最近更新时间排序） */
  memoryKey?: MemoryKey
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const isEditing = ref(false)
const localValue = ref('')
const inputRef = ref()

function startEdit() {
  localValue.value = String(props.modelValue || '')
  isEditing.value = true
  nextTick(() => inputRef.value?.focus())
}

defineExpose({ startEdit })

function handleClickDisplay() {
  // 空值/未识别时允许直接点击进入编辑；有值时避免误触
  if (props.modelValue === '' || props.modelValue === 0 || props.modelValue == null) {
    startEdit()
  }
}

function querySearch(queryString: string, cb: (arg: Array<{ value: string }>) => void) {
  if (!props.memoryKey) {
    cb([])
    return
  }
  const list = getSuggestions(props.memoryKey, queryString).map((v) => ({ value: v }))
  cb(list)
}

function handleSelect(item: { value: string }) {
  localValue.value = item.value
  handleConfirm()
}

function handleConfirm() {
  if (props.numeric) {
    const n = parseFloat(localValue.value)
    emit('update:modelValue', Number.isFinite(n) ? n : 0)
  } else {
    emit('update:modelValue', localValue.value)
  }
  isEditing.value = false
}
</script>
