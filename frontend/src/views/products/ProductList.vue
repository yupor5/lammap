<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-800">产品资料库</h1>
      <div class="flex gap-2">
        <el-button :icon="Upload" @click="handleImport">导入 Excel</el-button>
        <el-button type="primary" :icon="Plus" @click="showAddDialog = true">新建产品</el-button>
      </div>
    </div>

    <!-- 搜索 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-4">
      <el-input v-model="searchText" placeholder="搜索产品名称、SKU、分类…" prefix-icon="Search" clearable size="large" class="max-w-lg" />
    </div>

    <!-- 产品列表 -->
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm">
      <el-table :data="filteredProducts" class="w-full">
        <el-table-column prop="name" label="产品名称" min-width="150">
          <template #default="{ row }">
            <el-button text type="primary" @click="$router.push(`/products/${row.id}`)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="sku" label="SKU / 型号" width="120" />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="material" label="材质" width="120" />
        <el-table-column prop="price" label="参考价格" width="120" align="right">
          <template #default="{ row }">USD {{ row.price.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="moq" label="MOQ" width="80" align="right" />
        <el-table-column prop="leadTime" label="默认交期" width="100" />
        <el-table-column prop="attachments" label="附件数" width="80" align="center" />
        <el-table-column prop="updatedAt" label="更新时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="$router.push(`/products/${row.id}`)">编辑</el-button>
            <el-button text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

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
        <el-button type="primary" @click="handleAdd">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Plus } from '@element-plus/icons-vue'

const searchText = ref('')
const showAddDialog = ref(false)

const newProduct = reactive({
  name: '', sku: '', category: '', material: '', price: 0, moq: 100, leadTime: '',
})

const products = ref([
  { id: 1, name: '不锈钢桌腿', sku: 'STL-70B', category: '五金', material: 'SS304', price: 12.50, moq: 100, leadTime: '25-30天', attachments: 3, updatedAt: '2026-04-07' },
  { id: 2, name: '铝合金灯座', sku: 'AL-A3', category: '灯具', material: '铝合金 6063', price: 8.00, moq: 200, leadTime: '20-25天', attachments: 2, updatedAt: '2026-04-06' },
  { id: 3, name: '定制螺丝 M8x30', sku: 'SC-M8-30', category: '紧固件', material: '碳钢 8.8级', price: 0.15, moq: 5000, leadTime: '15-20天', attachments: 1, updatedAt: '2026-04-05' },
  { id: 4, name: '铜管接头', sku: 'CP-H12', category: '管件', material: '黄铜 H62', price: 3.20, moq: 500, leadTime: '15-20天', attachments: 2, updatedAt: '2026-04-04' },
])

const filteredProducts = computed(() => {
  if (!searchText.value) return products.value
  const s = searchText.value.toLowerCase()
  return products.value.filter((p) =>
    p.name.toLowerCase().includes(s) || p.sku.toLowerCase().includes(s) || p.category.toLowerCase().includes(s)
  )
})

function handleImport() {
  ElMessage.info('Excel 导入功能开发中')
}

function handleAdd() {
  ElMessage.success('产品已添加')
  showAddDialog.value = false
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确定删除产品 "${row.name}"？`, '提示', { type: 'warning' })
  products.value = products.value.filter((p) => p.id !== row.id)
  ElMessage.success('已删除')
}
</script>
