<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-800">产品资料库</h1>
      <div class="flex gap-2">
        <el-button :icon="MagicStick" @click="openAiExamples">AI 生成示例</el-button>
        <el-button :icon="Upload" @click="handleImport">导入 Excel</el-button>
        <el-button type="primary" :icon="Plus" @click="showAddDialog = true">新建产品</el-button>
      </div>
    </div>

    <!-- 搜索 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-4">
      <el-input
        v-model="searchText"
        placeholder="搜索产品名称、SKU、分类…"
        :prefix-icon="Search"
        clearable
        size="large"
        class="max-w-lg"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      />
    </div>

    <!-- 产品列表 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm">
      <el-table :data="productStore.products" v-loading="productStore.loading" class="w-full">
        <el-table-column prop="name" label="产品名称" min-width="150">
          <template #default="{ row }">
            <el-button text type="primary" @click="$router.push(`/products/${row.id}`)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="sku" label="SKU / 型号" width="120" />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="material" label="材质" width="120" />
        <el-table-column prop="price" label="参考价格" width="120" align="right">
          <template #default="{ row }">USD {{ (row.price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="moq" label="MOQ" width="80" align="right" />
        <el-table-column prop="leadTime" label="默认交期" width="100" />
        <el-table-column prop="attachments" label="附件数" width="80" align="center" />
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="$router.push(`/products/${row.id}`)">编辑</el-button>
            <el-button text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="productStore.products.length === 0 && !productStore.loading" class="p-8 text-center text-gray-400">
        暂无产品，点击右上角"新建产品"或"导入 Excel"添加
      </div>
      <div class="flex justify-end p-4" v-if="productStore.total > 20">
        <el-pagination v-model:current-page="page" :page-size="20" :total="productStore.total" layout="total, prev, pager, next" @current-change="loadProducts" />
      </div>
    </div>

    <!-- Excel 导入弹窗 -->
    <el-dialog v-model="showImportDialog" title="导入产品 (Excel)" width="700px">
      <el-upload
        drag
        :auto-upload="false"
        accept=".xls,.xlsx"
        :limit="1"
        :on-change="handleImportFileChange"
      >
        <el-icon class="el-icon--upload"><Upload /></el-icon>
        <div class="el-upload__text">拖拽 Excel 文件到此或 <em>点击选择</em></div>
        <template #tip><div class="el-upload__tip">支持 .xls / .xlsx，表头需包含"产品名称/name"等列名</div></template>
      </el-upload>
      <div v-if="importPreview.length > 0" class="mt-4">
        <p class="text-sm text-gray-500 mb-2">预览前 5 条（共 {{ importPreview.length }} 条）：</p>
        <el-table :data="importPreview.slice(0, 5)" border size="small" max-height="250">
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="sku" label="SKU" width="100" />
          <el-table-column prop="category" label="分类" width="80" />
          <el-table-column prop="material" label="材质" width="100" />
          <el-table-column prop="price" label="价格" width="80" align="right" />
          <el-table-column prop="moq" label="MOQ" width="60" align="right" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" :loading="importing" :disabled="importPreview.length === 0" @click="confirmImport">
          确认导入 ({{ importPreview.length }} 条)
        </el-button>
      </template>
    </el-dialog>

    <!-- AI 生成示例 -->
    <el-dialog v-model="aiExamplesVisible" title="AI 生成产品示例" width="720px">
      <div class="space-y-3">
        <div class="text-sm text-gray-500">
          生成的示例会先展示预览，你可以确认后再导入到产品库（最大 5 条）。
        </div>
        <el-form label-width="90px">
          <el-form-item label="条数">
            <el-input-number v-model="aiExampleCount" :min="1" :max="5" />
          </el-form-item>
          <el-form-item label="补充说明">
            <div class="w-full space-y-2">
              <el-autocomplete
                v-model="hintSearch"
                :fetch-suggestions="queryHintSuggestions"
                placeholder="搜索历史说明（支持中文/拼音）…"
                size="small"
                style="width: 100%"
                @select="onSelectHint"
              />
              <div class="flex gap-2">
                <el-button size="small" :loading="aiHintLoading" @click="runAiHint">AI 生成补充说明</el-button>
                <el-button size="small" @click="saveHintToHistory" :disabled="!aiExampleHint.trim()">保存到历史</el-button>
              </div>
              <el-input
                v-model="aiExampleHint"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 8 }"
                placeholder="（可选）例如：只生成面料类；强调字段包含克重/gsm、门幅；SKU 为空也可；默认币种 USD 等"
                @blur="saveHintToHistory"
              />
            </div>
          </el-form-item>
        </el-form>

        <div v-if="aiExampleItems.length > 0" class="pt-2">
          <div class="text-sm text-gray-500 mb-2">预览（{{ aiExampleItems.length }} 条）：</div>
          <el-table :data="aiExampleItems" border size="small" max-height="260">
            <el-table-column prop="name" label="产品名称" min-width="160" />
            <el-table-column prop="sku" label="SKU" width="110" />
            <el-table-column prop="category" label="分类" width="100" />
            <el-table-column prop="material" label="材质" width="140" />
            <el-table-column prop="price" label="参考价格" width="110" align="right">
              <template #default="{ row }">USD {{ (row.price || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="moq" label="MOQ" width="80" align="right" />
            <el-table-column prop="leadTime" label="交期" width="120" />
          </el-table>
        </div>
      </div>
      <template #footer>
        <el-button @click="aiExamplesVisible = false">取消</el-button>
        <el-button :loading="aiExampleLoading" @click="runAiExamples">开始生成</el-button>
        <el-button
          type="primary"
          :disabled="aiExampleItems.length === 0"
          :loading="aiExampleImporting"
          @click="importAiExamples"
        >
          导入到产品库 ({{ aiExampleItems.length }} 条)
        </el-button>
      </template>
    </el-dialog>

    <!-- 新建产品弹窗 -->
    <el-dialog v-model="showAddDialog" title="新建产品" width="600px">
      <el-form :model="newProduct" label-width="100px">
        <el-form-item label="产品名称"><el-input v-model="newProduct.name" /></el-form-item>
        <el-form-item label="SKU / 型号"><el-input v-model="newProduct.sku" /></el-form-item>
        <el-form-item label="分类"><el-input v-model="newProduct.category" /></el-form-item>
        <el-form-item label="材质"><el-input v-model="newProduct.material" /></el-form-item>
        <el-form-item label="参考价格"><el-input-number v-model="newProduct.price" :precision="2" :min="0" /></el-form-item>
        <el-form-item label="MOQ"><el-input-number v-model="newProduct.moq" :min="1" /></el-form-item>
        <el-form-item label="默认交期"><el-input v-model="newProduct.leadTime" placeholder="例：25-30天" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAdd">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useProductStore } from '@/stores/product'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Plus, Search, MagicStick } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { getSuggestions, rememberValue } from '@/utils/paramMemory'

const productStore = useProductStore()

const searchText = ref('')
const showAddDialog = ref(false)
const submitting = ref(false)
const page = ref(1)

const newProduct = reactive({
  name: '', sku: '', category: '', material: '', price: 0, moq: 100, leadTime: '',
})

function formatTime(t: string) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 10)
}

async function loadProducts() {
  await productStore.fetchProducts({ search: searchText.value, page: page.value, pageSize: 20 })
}

function handleSearch() {
  page.value = 1
  loadProducts()
}

const showImportDialog = ref(false)
const importFile = ref<File | null>(null)
const importPreview = ref<any[]>([])
const importing = ref(false)

const aiExamplesVisible = ref(false)
const aiExampleCount = ref(3)
const aiExampleHint = ref('')
const hintSearch = ref('')
const aiHintLoading = ref(false)
const aiExampleLoading = ref(false)
const aiExampleImporting = ref(false)
const aiExampleItems = ref<any[]>([])

function openAiExamples() {
  aiExamplesVisible.value = true
  aiExampleItems.value = []
  aiExampleLoading.value = false
  aiExampleImporting.value = false
  hintSearch.value = ''
  if (aiExampleCount.value < 1) aiExampleCount.value = 3
  if (aiExampleCount.value > 5) aiExampleCount.value = 5
}

function queryHintSuggestions(queryString: string, cb: (arg: Array<{ value: string }>) => void) {
  const list = getSuggestions('productExampleHint', queryString).map((v) => ({ value: v }))
  cb(list)
}

function onSelectHint(item: { value: string }) {
  aiExampleHint.value = item.value
  hintSearch.value = ''
  saveHintToHistory()
}

function saveHintToHistory() {
  rememberValue('productExampleHint', aiExampleHint.value)
}

async function runAiHint() {
  aiHintLoading.value = true
  try {
    const res: any = await request.post('/ai/compose-product-example-hint', {
      count: aiExampleCount.value,
      currentHint: aiExampleHint.value,
    })
    const content = String(res.data?.content || '').trim()
    if (!content) {
      ElMessage.warning('AI 未生成到补充说明')
      return
    }
    aiExampleHint.value = content
    saveHintToHistory()
  } catch (e) {
    console.error(e)
    ElMessage.error('生成补充说明失败')
  } finally {
    aiHintLoading.value = false
  }
}

async function runAiExamples() {
  aiExampleLoading.value = true
  try {
    saveHintToHistory()
    const created: any = await request.post('/products/ai-example-jobs', {
      count: aiExampleCount.value,
      extraHint: aiExampleHint.value,
    })
    const jobId = Number(created.data?.jobId)
    if (!jobId) throw new Error('创建生成任务失败')

    const jobData = await pollProductExampleJob(jobId)
    let rawResult = jobData?.result as Record<string, unknown> | null
    if (!rawResult && typeof jobData?.resultJson === 'string' && jobData.resultJson.trim()) {
      try {
        rawResult = JSON.parse(jobData.resultJson) as Record<string, unknown>
      } catch {
        rawResult = null
      }
    }
    aiExampleItems.value = ((rawResult as any)?.items as any[]) || []
    if (aiExampleItems.value.length === 0) {
      ElMessage.warning('未生成到示例')
    } else {
      ElMessage.success(`已生成 ${aiExampleItems.value.length} 条示例`)
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('生成失败，请稍后重试')
  } finally {
    aiExampleLoading.value = false
  }
}

async function pollProductExampleJob(jobId: number): Promise<Record<string, any>> {
  const start = Date.now()
  let interval = 1000
  while (true) {
    // eslint-disable-next-line no-await-in-loop
    const res: any = await request.get(`/products/ai-example-jobs/${jobId}`)
    const status = String(res.data?.status || '')
    if (status === 'succeeded') {
      return res.data as Record<string, any>
    }
    if (status === 'failed') {
      throw new Error(String(res.data?.errorMsg || '生成失败'))
    }
    const elapsed = Date.now() - start
    if (elapsed > 5 * 60 * 1000) {
      throw new Error('生成超时（超过 5 分钟仍未完成）')
    }
    if (elapsed > 10000 && interval < 2000) interval = 2000
    if (elapsed > 30000 && interval < 3000) interval = 3000
    if (elapsed > 60000 && interval < 5000) interval = 5000
    // eslint-disable-next-line no-await-in-loop
    await new Promise((r) => setTimeout(r, interval))
  }
}

async function importAiExamples() {
  if (aiExampleItems.value.length === 0) return
  aiExampleImporting.value = true
  try {
    saveHintToHistory()
    const res: any = await productStore.importProducts(aiExampleItems.value)
    ElMessage.success(res.message || `成功导入 ${res.created} 个产品`)
    aiExamplesVisible.value = false
    loadProducts()
  } catch (e) {
    console.error(e)
    ElMessage.error('导入失败')
  } finally {
    aiExampleImporting.value = false
  }
}

function handleImport() {
  showImportDialog.value = true
  importFile.value = null
  importPreview.value = []
}

async function handleImportFileChange(uploadFile: any) {
  if (!uploadFile?.raw) return
  importFile.value = uploadFile.raw
  try {
    const { parseProductExcel } = await import('@/utils/exportExcel')
    const rows = await parseProductExcel(uploadFile.raw)
    importPreview.value = rows
    ElMessage.success(`识别到 ${rows.length} 个产品`)
  } catch (err: any) {
    ElMessage.error(err.message || '解析失败')
    importPreview.value = []
  }
}

async function confirmImport() {
  if (importPreview.value.length === 0) {
    ElMessage.warning('没有可导入的产品')
    return
  }
  importing.value = true
  try {
    const res: any = await productStore.importProducts(importPreview.value)
    ElMessage.success(res.message || `成功导入 ${res.created} 个产品`)
    showImportDialog.value = false
    loadProducts()
  } catch {
    ElMessage.error('导入失败')
  } finally {
    importing.value = false
  }
}

async function handleAdd() {
  if (!newProduct.name.trim()) {
    ElMessage.warning('请输入产品名称')
    return
  }
  submitting.value = true
  try {
    await productStore.createProduct({ ...newProduct })
    ElMessage.success('产品已添加')
    showAddDialog.value = false
    Object.assign(newProduct, { name: '', sku: '', category: '', material: '', price: 0, moq: 100, leadTime: '' })
    loadProducts()
  } catch {
    // error handled in interceptor
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除产品 "${row.name}"？`, '提示', { type: 'warning' })
  try {
    await productStore.deleteProduct(row.id)
    ElMessage.success('已删除')
    loadProducts()
  } catch {
    // error handled in interceptor
  }
}

onMounted(() => {
  loadProducts()
})
</script>
