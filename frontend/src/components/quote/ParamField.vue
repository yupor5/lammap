<template>
  <div class="flex items-center justify-between p-2 rounded bg-gray-50 hover:bg-gray-100 transition-colors">
    <span class="text-sm text-gray-600 shrink-0 w-20">{{ label }}</span>
    <div class="flex-1 ml-2">
      <el-input
        v-if="isEditing"
        v-model="localValue"
        size="small"
        @blur="handleConfirm"
        @keyup.enter="handleConfirm"
        ref="inputRef"
      />
      <div v-else class="flex items-center gap-1">
        <span v-if="modelValue" class="text-sm text-gray-800">{{ modelValue }}</span>
        <el-tag v-else type="warning" size="small">未识别</el-tag>
        <el-button text size="small" :icon="Edit" @click="startEdit" class="ml-auto" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { Edit } from '@element-plus/icons-vue'

const props = defineProps<{
  label: string
  modelValue: string | number
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

function handleConfirm() {
  emit('update:modelValue', localValue.value)
  isEditing.value = false
}
</script>
