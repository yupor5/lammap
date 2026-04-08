<template>
  <div class="space-y-4" v-loading="loading">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <el-button text :icon="ArrowLeft" @click="$router.back()">返回</el-button>
        <h1 class="text-2xl font-bold text-gray-800">报价详情 #{{ quote?.quoteNumber || $route.params.id }}</h1>
        <el-tag v-if="quote" :type="statusType(quote.status)">{{ quote.status }}</el-tag>
      </div>
      <div class="flex gap-2">
        <el-button :icon="Download" @click="handleExportPdf">导出 PDF</el-button>
        <el-button @click="handleExportExcel">导出 Excel</el-button>
        <el-button type="success" @click="handleDuplicate">基于此报价新建</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存修改</el-button>
      </div>
    </div>

    <template v-if="quote">
      <div class="grid grid-cols-12 gap-4" style="height: calc(100vh - 180px);">
        <!-- 左侧: 基本信息 -->
        <div class="col-span-4 bg-white rounded-lg border border-gray-100 shadow-sm p-4 overflow-y-auto">
          <h3 class="font-semibold text-gray-800 mb-4">基本信息</h3>
          <el-form label-width="80px">
            <el-form-item label="客户名称"><el-input v-model="quote.customerName" /></el-form-item>
            <el-form-item label="国家"><el-input v-model="quote.country" /></el-form-item>
            <el-form-item label="币种"><el-input v-model="quote.currency" /></el-form-item>
            <el-form-item label="交付地址"><el-input v-model="quote.deliveryAddress" /></el-form-item>
            <el-form-item label="状态">
              <el-select v-model="quote.status" class="w-full">
                <el-option label="草稿" value="草稿" />
                <el-option label="已发送" value="已发送" />
                <el-option label="待确认" value="待确认" />
                <el-option label="已成交" value="已成交" />
              </el-select>
            </el-form-item>
            <el-form-item label="交期"><el-input v-model="quote.leadTime" placeholder="例：25-30天" /></el-form-item>
            <el-form-item label="条款"><el-input v-model="quote.terms" type="textarea" :rows="3" placeholder="付款条款、贸易条款等" /></el-form-item>
            <el-form-item label="备注"><el-input v-model="quote.remarks" type="textarea" :rows="3" placeholder="内部备注或给客户的附加说明" /></el-form-item>
          </el-form>

          <h3 class="font-semibold text-gray-800 mb-4 mt-6">原始需求</h3>
          <el-input v-model="quote.rawRequirement" type="textarea" :rows="6" placeholder="客户原始需求内容" />
        </div>

        <!-- 右侧: 报价明细 -->
        <div class="col-span-8 bg-white rounded-lg border border-gray-100 shadow-sm p-4 overflow-y-auto">
          <h3 class="font-semibold text-gray-800 mb-4">报价明细</h3>
          <el-table :data="quote.items" border size="small" show-summary :summary-method="getSummary">
            <el-table-column prop="productName" label="产品名称" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.productName" size="small" />
              </template>
            </el-table-column>
            <el-table-column prop="model" label="型号" width="100">
              <template #default="{ row }">
                <el-input v-model="row.model" size="small" />
              </template>
            </el-table-column>
            <el-table-column prop="specs" label="规格" min-width="150">
              <template #default="{ row }">
                <el-input v-model="row.specs" size="small" />
              </template>
            </el-table-column>
            <el-table-column prop="quantity" label="数量" width="100" align="right">
              <template #default="{ row }">
                <el-input-number v-model="row.quantity" :min="1" size="small" controls-position="right" @change="recalcRow(row)" />
              </template>
            </el-table-column>
            <el-table-column prop="unitPrice" label="单价" width="110" align="right">
              <template #default="{ row }">
                <el-input-number v-model="row.unitPrice" :min="0" :precision="2" size="small" controls-position="right" @change="recalcRow(row)" />
              </template>
            </el-table-column>
            <el-table-column prop="totalPrice" label="小计" width="110" align="right">
              <template #default="{ row }">
                {{ quote.currency }} {{ (row.totalPrice || 0).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
              </template>
            </el-table-column>
          </el-table>

          <div class="mt-4 text-right">
            <span class="text-lg font-bold">
              总计: {{ quote.currency }} {{ totalAmount.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
            </span>
          </div>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="bg-white rounded-lg border border-gray-100 shadow-sm p-16 text-center text-gray-400">
      <el-icon :size="64"><Document /></el-icon>
      <p class="mt-4 text-lg">报价不存在</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Document, Download } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { exportQuotePdf } from '@/utils/exportPdf'
import { exportQuoteExcel } from '@/utils/exportExcel'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const quote = ref<any>(null)

const totalAmount = computed(() => {
  if (!quote.value?.items) return 0
  return quote.value.items.reduce((sum: number, item: any) => sum + (item.totalPrice || 0), 0)
})

function statusType(status: string) {
  const map: Record<string, string> = { '草稿': 'info', '已发送': 'success', '待确认': 'warning', '已成交': '' }
  return map[status] || 'info'
}

function recalcRow(row: any) {
  row.totalPrice = Number((row.quantity * row.unitPrice).toFixed(2))
  quote.value.totalAmount = totalAmount.value
}

function getSummary({ columns, data }: { columns: any[]; data: any[] }) {
  const sums: string[] = []
  columns.forEach((_: any, index: number) => {
    if (index === 0) sums[index] = '合计'
    else if (index === columns.length - 1) {
      const t = data.reduce((s, item) => s + (item.totalPrice || 0), 0)
      sums[index] = `${quote.value?.currency || ''} ${t.toLocaleString(undefined, { minimumFractionDigits: 2 })}`
    } else sums[index] = ''
  })
  return sums
}

async function loadQuote() {
  loading.value = true
  try {
    const res: any = await request.get(`/quotes/${route.params.id}`)
    quote.value = res.data
  } catch {
    ElMessage.error('加载报价失败')
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  if (!quote.value) return
  saving.value = true
  try {
    await request.put(`/quotes/${route.params.id}`, quote.value)
    ElMessage.success('保存成功')
  } catch {
    // error handled in interceptor
  } finally {
    saving.value = false
  }
}

function handleExportPdf() {
  if (!quote.value) return
  exportQuotePdf({
    quoteNumber: quote.value.quoteNumber,
    customerName: quote.value.customerName,
    country: quote.value.country,
    currency: quote.value.currency,
    deliveryAddress: quote.value.deliveryAddress,
    leadTime: quote.value.leadTime,
    remarks: quote.value.remarks,
    terms: quote.value.terms,
    items: quote.value.items || [],
    totalAmount: totalAmount.value,
  })
  ElMessage.success('PDF 已导出')
}

function handleExportExcel() {
  if (!quote.value) return
  exportQuoteExcel({
    quoteNumber: quote.value.quoteNumber,
    customerName: quote.value.customerName,
    country: quote.value.country,
    currency: quote.value.currency,
    deliveryAddress: quote.value.deliveryAddress,
    leadTime: quote.value.leadTime,
    remarks: quote.value.remarks,
    terms: quote.value.terms,
    items: quote.value.items || [],
    totalAmount: totalAmount.value,
  })
  ElMessage.success('Excel 已导出')
}

function handleDuplicate() {
  router.push({ path: '/quotes/new', query: { from: String(route.params.id) } })
}

onMounted(() => {
  loadQuote()
})
</script>
