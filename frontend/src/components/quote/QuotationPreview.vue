<template>
  <div class="space-y-4">
    <!-- 操作栏 -->
    <div class="flex items-center justify-between">
      <h4 class="font-semibold text-gray-800">报价单预览</h4>
      <div class="flex gap-2">
        <el-button size="small" :icon="DocumentCopy" @click="handleCopyText">复制文本</el-button>
        <el-button size="small" :icon="Download" @click="handleExportPDF">导出 PDF</el-button>
        <el-button size="small" @click="handleExportExcel">导出 Excel</el-button>
      </div>
    </div>

    <!-- 客户信息 -->
    <div class="bg-gray-50 rounded-lg p-4">
      <div class="grid grid-cols-2 gap-3 text-sm">
        <div><span class="text-gray-500">客户：</span>{{ quote.customerName }}</div>
        <div><span class="text-gray-500">币种：</span>{{ quote.currency }}</div>
      </div>
    </div>

    <!-- 产品明细表 -->
    <el-table :data="localItems" border size="small" show-summary :summary-method="getSummary">
      <el-table-column prop="productName" label="产品名称" min-width="140" />
      <el-table-column prop="model" label="型号" width="100" />
      <el-table-column prop="specs" label="规格" min-width="150" />
      <el-table-column prop="quantity" label="数量" width="80" align="right">
        <template #default="{ row, $index }">
          <el-input-number v-model="row.quantity" :min="1" size="small" controls-position="right" @change="recalcRow($index)" />
        </template>
      </el-table-column>
      <el-table-column prop="unitPrice" label="单价" width="100" align="right">
        <template #default="{ row, $index }">
          <el-input-number v-model="row.unitPrice" :min="0" :precision="2" size="small" controls-position="right" @change="recalcRow($index)" />
        </template>
      </el-table-column>
      <el-table-column prop="totalPrice" label="小计" width="100" align="right">
        <template #default="{ row }">
          {{ quote.currency }} {{ row.totalPrice.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Quote, QuoteItem } from '@/stores/quote'
import { ElMessage } from 'element-plus'
import { DocumentCopy, Download } from '@element-plus/icons-vue'

const props = defineProps<{ quote: Quote }>()
const emit = defineEmits<{ 'update:items': [items: QuoteItem[]] }>()

const localItems = ref<QuoteItem[]>([...props.quote.items])

watch(() => props.quote.items, (newItems) => {
  localItems.value = [...newItems]
})

function recalcRow(index: number) {
  const item = localItems.value[index]
  item.totalPrice = Number((item.quantity * item.unitPrice).toFixed(2))
  emit('update:items', localItems.value)
}

function getSummary({ columns, data }: { columns: any[]; data: QuoteItem[] }) {
  const sums: string[] = []
  columns.forEach((_col: any, index: number) => {
    if (index === 0) {
      sums[index] = '合计'
    } else if (index === columns.length - 1) {
      const total = data.reduce((sum, item) => sum + (item.totalPrice || 0), 0)
      sums[index] = `${props.quote.currency} ${total.toLocaleString(undefined, { minimumFractionDigits: 2 })}`
    } else {
      sums[index] = ''
    }
  })
  return sums
}

function handleCopyText() {
  const text = localItems.value.map((item) =>
    `${item.productName} | ${item.model} | ${item.specs} | x${item.quantity} | ${props.quote.currency} ${item.unitPrice} | ${props.quote.currency} ${item.totalPrice}`
  ).join('\n')
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

function handleExportPDF() {
  ElMessage.info('PDF 导出功能开发中')
}

function handleExportExcel() {
  ElMessage.info('Excel 导出功能开发中')
}
</script>
