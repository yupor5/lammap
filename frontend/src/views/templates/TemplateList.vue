<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-800">模板管理</h1>
      <el-button type="primary" :icon="Plus" @click="handleAdd">新建模板</el-button>
    </div>

    <div class="grid grid-cols-12 gap-4" style="height: calc(100vh - 180px);">
      <!-- 左侧分类 -->
      <div class="col-span-3 bg-white rounded-lg border border-gray-100 shadow-sm overflow-hidden">
        <div class="p-4 border-b border-gray-100">
          <h3 class="font-semibold text-gray-800">模板分类</h3>
        </div>
        <el-menu :default-active="activeCategory" @select="handleCategorySelect">
          <el-menu-item v-for="cat in categories" :key="cat.value" :index="cat.value">
            <el-icon><component :is="cat.icon" /></el-icon>
            <span>{{ cat.label }}</span>
          </el-menu-item>
        </el-menu>
      </div>

      <!-- 右侧编辑区 -->
      <div class="col-span-9 bg-white rounded-lg border border-gray-100 shadow-sm flex flex-col overflow-hidden">
        <div class="p-4 border-b border-gray-100 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <el-select v-model="selectedTemplate" placeholder="选择模板" class="w-60">
              <el-option v-for="t in currentTemplates" :key="t.id" :label="t.name" :value="t.id" />
            </el-select>
          </div>
          <div class="flex gap-2">
            <el-button size="small" @click="handlePreview">预览</el-button>
            <el-button size="small" type="primary" @click="handleSave">保存</el-button>
          </div>
        </div>

        <!-- 变量插入区 -->
        <div class="px-4 py-2 border-b border-gray-100 flex items-center gap-2 flex-wrap">
          <span class="text-sm text-gray-500">插入变量：</span>
          <el-button v-for="v in variables" :key="v.key" size="small" @click="insertVariable(v.key)">
            {{ v.label }}
          </el-button>
        </div>

        <!-- 编辑器 -->
        <div class="flex-1 p-4">
          <el-input
            ref="editorRef"
            v-model="templateContent"
            type="textarea"
            :autosize="{ minRows: 15, maxRows: 30 }"
            placeholder="在此编辑模板内容，使用上方按钮插入变量…"
            class="h-full"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Document, Message, ChatDotRound, List } from '@element-plus/icons-vue'

const activeCategory = ref('quotation')
const selectedTemplate = ref(1)
const templateContent = ref(`Dear {{customer_name}},

Thank you for your inquiry regarding {{product_name}}.

We are pleased to offer the following quotation:

Product: {{product_name}}
Quantity: {{quantity}}
Unit Price: {{price}}
Lead Time: {{lead_time}}
Payment Terms: {{payment_terms}}

This quotation is valid for 30 days.

Best regards`)

const editorRef = ref()

const categories = [
  { label: '报价单模板', value: 'quotation', icon: Document },
  { label: '邮件回复模板', value: 'email', icon: Message },
  { label: 'WhatsApp/微信模板', value: 'chat', icon: ChatDotRound },
  { label: '参数确认模板', value: 'confirmation', icon: List },
]

const templates = ref([
  { id: 1, name: '标准报价邮件', category: 'quotation' },
  { id: 2, name: '简洁报价邮件', category: 'quotation' },
  { id: 3, name: '标准邮件回复', category: 'email' },
  { id: 4, name: 'WhatsApp 快速回复', category: 'chat' },
  { id: 5, name: '参数确认模板', category: 'confirmation' },
])

const variables = [
  { key: 'customer_name', label: '客户名' },
  { key: 'product_name', label: '产品名' },
  { key: 'quantity', label: '数量' },
  { key: 'price', label: '价格' },
  { key: 'lead_time', label: '交期' },
  { key: 'payment_terms', label: '付款条款' },
]

const currentTemplates = computed(() =>
  templates.value.filter((t) => t.category === activeCategory.value)
)

function handleCategorySelect(key: string) {
  activeCategory.value = key
  const first = currentTemplates.value[0]
  if (first) selectedTemplate.value = first.id
}

function insertVariable(key: string) {
  templateContent.value += `{{${key}}}`
}

function handlePreview() {
  ElMessage.info('预览功能开发中')
}

function handleSave() {
  ElMessage.success('模板已保存')
}

function handleAdd() {
  ElMessage.info('新建模板功能开发中')
}
</script>
