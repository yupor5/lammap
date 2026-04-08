<template>
  <div class="space-y-6">
    <!-- 顶部操作区 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">工作台</h1>
        <p class="text-gray-500 mt-1">欢迎回来，快速开始你的报价工作</p>
      </div>
      <el-button type="primary" size="large" :icon="DocumentAdd" @click="$router.push('/quotes/new')">
        生成新报价
      </el-button>
    </div>

    <!-- 今日工作统计 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="item in statCards" :key="item.label"
        class="bg-white rounded-lg p-5 border border-gray-100 shadow-sm hover:shadow-md transition-shadow"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ item.label }}</p>
            <p class="text-3xl font-bold mt-1" :class="item.color">{{ item.value }}</p>
          </div>
          <div class="w-12 h-12 rounded-full flex items-center justify-center" :class="item.bgColor">
            <el-icon :size="24" :class="item.iconColor"><component :is="item.icon" /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 最近报价 -->
      <div class="lg:col-span-2 bg-white rounded-lg border border-gray-100 shadow-sm">
        <div class="flex items-center justify-between p-4 border-b border-gray-100">
          <h3 class="text-lg font-semibold text-gray-800">最近报价</h3>
          <el-button text type="primary" @click="$router.push('/quotes/history')">查看全部</el-button>
        </div>
        <el-table :data="dashboardStore.recentQuotes" v-loading="dashboardStore.loading" class="w-full">
          <el-table-column prop="customerName" label="客户名" min-width="120" />
          <el-table-column label="产品" min-width="120">
            <template #default="{ row }">
              {{ row.items?.[0]?.productName || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="金额" min-width="100">
            <template #default="{ row }">
              <span class="font-semibold">{{ row.currency }} {{ (row.totalAmount || 0).toLocaleString() }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="updatedAt" label="更新时间" width="160">
            <template #default="{ row }">
              {{ formatTime(row.updatedAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button text size="small" type="primary" @click="$router.push(`/quotes/${row.id}`)">查看</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="dashboardStore.recentQuotes.length === 0 && !dashboardStore.loading" class="p-8 text-center text-gray-400">
          暂无报价记录，点击右上角"生成新报价"开始
        </div>
      </div>

      <!-- 快速入口 -->
      <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-4">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">快速入口</h3>
        <div class="space-y-3">
          <el-button class="w-full !justify-start" size="large" :icon="DocumentAdd" @click="$router.push('/quotes/new')">
            新建报价
          </el-button>
          <el-button class="w-full !justify-start" size="large" :icon="Upload" @click="$router.push('/products')">
            导入产品资料
          </el-button>
          <el-button class="w-full !justify-start" size="large" :icon="Files" @click="$router.push('/templates')">
            管理模板
          </el-button>
          <el-button class="w-full !justify-start" size="large" :icon="Clock" @click="$router.push('/quotes/history')">
            查看历史
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useDashboardStore } from '@/stores/dashboard'
import { DocumentAdd, Upload, Files, Clock, Document, Checked, Promotion, TrendCharts } from '@element-plus/icons-vue'

const dashboardStore = useDashboardStore()

const statCards = computed(() => [
  { label: '今日新报价', value: dashboardStore.stats.todayQuotes, icon: Document, color: 'text-blue-600', bgColor: 'bg-blue-50', iconColor: 'text-blue-500' },
  { label: '待确认参数', value: dashboardStore.stats.pendingParams, icon: Checked, color: 'text-orange-600', bgColor: 'bg-orange-50', iconColor: 'text-orange-500' },
  { label: '待发送', value: dashboardStore.stats.toSend, icon: Promotion, color: 'text-green-600', bgColor: 'bg-green-50', iconColor: 'text-green-500' },
  { label: '本周已发送', value: dashboardStore.stats.weekSent, icon: TrendCharts, color: 'text-purple-600', bgColor: 'bg-purple-50', iconColor: 'text-purple-500' },
])

function statusTagType(status: string) {
  const map: Record<string, string> = { '草稿': 'info', '已发送': 'success', '待确认': 'warning', '已成交': '' }
  return map[status] || 'info'
}

function formatTime(t: string) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 16)
}

onMounted(() => {
  dashboardStore.fetchAll()
})
</script>
