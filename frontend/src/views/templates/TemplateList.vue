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
            <el-select v-model="selectedTemplateId" placeholder="选择模板" class="w-60" @change="handleTemplateSelect">
              <el-option v-for="t in templateStore.templates" :key="t.id" :label="t.name" :value="t.id" />
            </el-select>
            <el-tag v-if="currentTemplateSourceLabel" size="small" type="info">{{ currentTemplateSourceLabel }}</el-tag>
            <el-input v-model="templateName" placeholder="模板名称" class="w-48" size="default" />
          </div>
          <div class="flex gap-2">
            <el-select v-model="aiLang" size="small" class="w-28">
              <el-option label="中文" value="zh" />
              <el-option label="English" value="en" />
            </el-select>
            <el-button size="small" :loading="aiGenerating" @click="openAIGenerateDialog">AI 生成</el-button>
            <el-button size="small" type="danger" :disabled="!selectedTemplateId" @click="handleDeleteTemplate">删除</el-button>
            <el-button size="small" type="primary" :loading="saving" @click="handleSave">保存</el-button>
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
          />
        </div>
      </div>
    </div>
  </div>

  <el-dialog v-model="aiDialogVisible" title="AI 生成模板" width="720px">
    <div class="space-y-3">
      <div class="text-sm text-gray-500">
        将按当前分类「{{ activeCategoryLabel() }}」生成一份模板，并自动保存到模板列表中。
      </div>
      <el-form label-width="90px">
        <el-form-item label="语言">
          <el-select v-model="aiLang" class="w-40">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </el-form-item>
        <el-form-item label="补充说明">
          <el-input
            v-model="aiExtraHint"
            type="textarea"
            :autosize="{ minRows: 6, maxRows: 10 }"
            placeholder="你可以补充：\n- 语气/风格：更正式/更简短/更像中国外贸业务员/更偏欧美商务邮件\n- 条款偏好：FOB/CIF/有效期/样品政策/售后与质保\n- 信息缺口：强调需要客户确认哪些参数（颜色/尺寸/包装/交期等）\n- 目标渠道：邮件更长更正式；WhatsApp/微信更短更口语\n- 禁止项：不要写死价格/交期，只写“待确认/供参考”\n\n示例：请用中国外贸业务员常用表达，语气专业但不生硬；价格用“subject to final confirmation”。"
          />
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <div class="flex items-center justify-between w-full">
        <el-button @click="aiDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="aiGenerating" @click="handleAIGenerate">开始生成</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useTemplateStore } from '@/stores/template'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Document, Message, ChatDotRound, List } from '@element-plus/icons-vue'
import request from '@/utils/request'

const templateStore = useTemplateStore()

const activeCategory = ref('quotation')
const selectedTemplateId = ref<number | null>(null)

const currentTemplateSourceLabel = computed(() => {
  const id = selectedTemplateId.value
  if (!id) return ''
  const t = templateStore.templates.find((x) => x.id === id)
  const s = (t?.source || 'user').toLowerCase()
  if (s === 'ai') return 'AI'
  if (s === 'system') return '系统'
  return '手工'
})
const templateName = ref('')
const templateContent = ref('')
const saving = ref(false)
const editorRef = ref()
const aiGenerating = ref(false)
const aiLang = ref<'zh' | 'en'>('zh')
const aiDialogVisible = ref(false)
const aiExtraHint = ref('')

const categories = [
  { label: '报价单模板', value: 'quotation', icon: Document },
  { label: '邮件回复模板', value: 'email', icon: Message },
  { label: 'WhatsApp/微信模板', value: 'chat', icon: ChatDotRound },
  { label: '参数确认模板', value: 'confirmation', icon: List },
]

const variables = [
  { key: 'customer_name', label: '客户名' },
  { key: 'product_name', label: '产品名' },
  { key: 'quantity', label: '数量' },
  { key: 'price', label: '价格' },
  { key: 'lead_time', label: '交期' },
  { key: 'payment_terms', label: '付款条款' },
]

async function loadTemplates() {
  await templateStore.fetchTemplates(activeCategory.value, aiLang.value)
  if (templateStore.templates.length > 0) {
    selectedTemplateId.value = templateStore.templates[0].id
    handleTemplateSelect(templateStore.templates[0].id)
  } else {
    selectedTemplateId.value = null
    templateName.value = ''
    templateContent.value = ''
  }
}

function handleCategorySelect(key: string) {
  activeCategory.value = key
  loadTemplates()
}

function handleTemplateSelect(id: number) {
  const tmpl = templateStore.templates.find((t) => t.id === id)
  if (tmpl) {
    templateName.value = tmpl.name
    templateContent.value = tmpl.content
  }
}

function insertVariable(key: string) {
  templateContent.value += `{{${key}}}`
}

function activeCategoryLabel() {
  return categories.find((c) => c.value === activeCategory.value)?.label || activeCategory.value
}

function openAIGenerateDialog() {
  aiExtraHint.value = ''
  aiDialogVisible.value = true
}

async function handleAIGenerate() {
  aiGenerating.value = true
  try {
    const res: any = await request.post('/ai/generate-template', {
      category: activeCategory.value,
      categoryLabel: activeCategoryLabel(),
      language: aiLang.value,
      nameHint: templateName.value.trim() || '',
      extraHint: aiExtraHint.value.trim() || '',
    })
    const tmpl = res.data as { id: number; name: string; content: string }
    // 生成后自动选中并填充编辑器
    await templateStore.fetchTemplates(activeCategory.value, aiLang.value)
    selectedTemplateId.value = tmpl.id
    templateName.value = tmpl.name
    templateContent.value = tmpl.content
    aiDialogVisible.value = false
    ElMessage.success('已生成模板')
  } catch {
    // error handled in interceptor
  } finally {
    aiGenerating.value = false
  }
}

async function handleSave() {
  if (!templateName.value.trim()) {
    ElMessage.warning('请输入模板名称')
    return
  }
  saving.value = true
  try {
    if (selectedTemplateId.value) {
      await templateStore.updateTemplate(selectedTemplateId.value, {
        name: templateName.value,
        category: activeCategory.value,
        language: aiLang.value,
        content: templateContent.value,
      })
      ElMessage.success('模板已更新')
    } else {
      const created = await templateStore.createTemplate({
        name: templateName.value,
        category: activeCategory.value,
        language: aiLang.value,
        content: templateContent.value,
      })
      selectedTemplateId.value = created.id
      ElMessage.success('模板已创建')
    }
    loadTemplates()
  } catch {
    // error handled in interceptor
  } finally {
    saving.value = false
  }
}

function handleAdd() {
  selectedTemplateId.value = null
  templateName.value = ''
  templateContent.value = ''
}

async function handleDeleteTemplate() {
  if (!selectedTemplateId.value) return
  await ElMessageBox.confirm('确定删除此模板？', '提示', { type: 'warning' })
  try {
    await templateStore.deleteTemplate(selectedTemplateId.value)
    ElMessage.success('已删除')
    loadTemplates()
  } catch {
    // error handled in interceptor
  }
}

onMounted(() => {
  loadTemplates()
})
</script>
