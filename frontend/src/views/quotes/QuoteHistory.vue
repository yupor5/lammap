<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-800">历史报价</h1>
      <el-button type="primary" :icon="DocumentAdd" @click="$router.push('/quotes/new')">新建报价</el-button>
    </div>

    <!-- 筛选条件 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-4">
      <div class="grid grid-cols-5 gap-4">
        <el-input v-model="filters.customer" placeholder="客户名" prefix-icon="Search" clearable />
        <el-input v-model="filters.product" placeholder="产品" clearable />
        <el-date-picker v-model="filters.dateRange" type="daterange" start-placeholder="开始日期" end-placeholder="结束日期" class="w-full" />
        <el-select v-model="filters.status" placeholder="状态" clearable class="w-full">
          <el-option label="草稿" value="草稿" />
          <el-option label="已发送" value="已发送" />
          <el-option label="待确认" value="待确认" />
          <el-option label="已成交" value="已成交" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
    </div>

    <!-- 报价列表 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm">
      <el-table :data="quotes" class="w-full">
        <el-table-column prop="id" label="报价编号" width="100" />
        <el-table-column prop="customerName" label="客户" min-width="140" />
        <el-table-column prop="product" label="产品" min-width="140" />
        <el-table-column prop="amount" label="金额" width="130" align="right">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.currency }} {{ row.amount.toLocaleString() }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdBy" label="创建人" width="100" />
        <el-table-column prop="updatedAt" label="更新时间" width="160" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="$router.push(`/quotes/${row.id}`)">查看</el-button>
            <el-button text size="small" type="success" @click="handleDuplicate(row)">基于此报价新建</el-button>
            <el-button text size="small" @click="handleExport(row)">导出</el-button>
            <el-button text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="flex justify-end p-4">
        <el-pagination v-model:current-page="page" :page-size="20" :total="total" layout="total, prev, pager, next" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DocumentAdd } from '@element-plus/icons-vue'

const router = useRouter()
const page = ref(1)
const total = ref(50)

const filters = reactive({
  customer: '',
  product: '',
  dateRange: null as [Date, Date] | null,
  status: '',
})

const quotes = ref([
  { id: 'QT-2026-001', customerName: 'ABC Trading Co.', product: '不锈钢桌腿 70cm', amount: 15000, currency: 'USD', status: '已发送', createdBy: '张三', updatedAt: '2026-04-07 10:30' },
  { id: 'QT-2026-002', customerName: 'Euro Home Ltd.', product: '铝合金灯座 A3', amount: 8500, currency: 'EUR', status: '已成交', createdBy: '张三', updatedAt: '2026-04-06 09:15' },
  { id: 'QT-2026-003', customerName: '东方贸易有限公司', product: '五金配件套装', amount: 52000, currency: 'CNY', status: '待确认', createdBy: '李四', updatedAt: '2026-04-05 16:45' },
  { id: 'QT-2026-004', customerName: 'Global Parts Inc.', product: '定制螺丝 M8x30', amount: 3200, currency: 'USD', status: '草稿', createdBy: '张三', updatedAt: '2026-04-05 14:20' },
  { id: 'QT-2026-005', customerName: 'Smith & Co.', product: '铜管接头 1/2"', amount: 6800, currency: 'USD', status: '已发送', createdBy: '李四', updatedAt: '2026-04-04 11:00' },
])

function statusType(status: string) {
  const map: Record<string, string> = { '草稿': 'info', '已发送': 'success', '待确认': 'warning', '已成交': '' }
  return map[status] || 'info'
}

function handleSearch() {
  ElMessage.info('搜索功能开发中')
}

function handleDuplicate(row: any) {
  router.push({ path: '/quotes/new', query: { from: row.id } })
}

function handleExport(row: any) {
  ElMessage.info(`导出 ${row.id}`)
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm('确定删除此报价？', '提示', { type: 'warning' })
  ElMessage.success('已删除')
}
</script>
