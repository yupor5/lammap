<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-800">历史报价</h1>
      <el-button type="primary" :icon="DocumentAdd" @click="$router.push('/quotes/new')">新建报价</el-button>
    </div>

    <!-- 筛选条件 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-4">
      <div class="grid grid-cols-5 gap-4">
        <el-input v-model="filters.customer" placeholder="客户名" :prefix-icon="Search" clearable @keyup.enter="handleSearch" />
        <el-select v-model="filters.status" placeholder="状态" clearable class="w-full" @change="handleSearch">
          <el-option label="草稿" value="草稿" />
          <el-option label="已发送" value="已发送" />
          <el-option label="待确认" value="待确认" />
          <el-option label="已成交" value="已成交" />
        </el-select>
        <div class="col-span-2"></div>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
    </div>

    <!-- 报价列表 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm">
      <el-table :data="quotes" v-loading="loading" class="w-full">
        <el-table-column prop="quoteNumber" label="报价编号" width="140" />
        <el-table-column prop="customerName" label="客户" min-width="140" />
        <el-table-column label="产品" min-width="140">
          <template #default="{ row }">
            {{ row.items?.[0]?.productName || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="金额" width="130" align="right">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.currency }} {{ (row.totalAmount || 0).toLocaleString() }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="$router.push(`/quotes/${row.id}`)">查看</el-button>
            <el-button text size="small" type="success" @click="handleDuplicate(row)">基于此报价新建</el-button>
            <el-button text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="quotes.length === 0 && !loading" class="p-8 text-center text-gray-400">
        暂无报价记录
      </div>
      <div class="flex justify-end p-4">
        <el-pagination v-model:current-page="page" :page-size="20" :total="total" layout="total, prev, pager, next" @current-change="loadQuotes" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useQuoteStore } from '@/stores/quote'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DocumentAdd, Search } from '@element-plus/icons-vue'
import request from '@/utils/request'

const router = useRouter()
const quoteStore = useQuoteStore()

const loading = ref(false)
const page = ref(1)
const total = ref(0)
const quotes = ref<any[]>([])

const filters = reactive({ customer: '', status: '' })

function statusType(status: string) {
  const map: Record<string, string> = { '草稿': 'info', '已发送': 'success', '待确认': 'warning', '已成交': '' }
  return map[status] || 'info'
}

function formatTime(t: string) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 16)
}

async function loadQuotes() {
  loading.value = true
  try {
    const res: any = await request.get('/quotes', {
      params: {
        page: page.value,
        pageSize: 20,
        customer: filters.customer || undefined,
        status: filters.status || undefined,
      },
    })
    quotes.value = res.data.items || []
    total.value = res.data.total || 0
  } catch {
    // error handled in interceptor
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  loadQuotes()
}

function handleDuplicate(row: any) {
  router.push({ path: '/quotes/new', query: { from: row.id } })
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm('确定删除此报价？', '提示', { type: 'warning' })
  try {
    await request.delete(`/quotes/${row.id}`)
    ElMessage.success('已删除')
    loadQuotes()
  } catch {
    // error handled in interceptor
  }
}

onMounted(() => {
  loadQuotes()
})
</script>
