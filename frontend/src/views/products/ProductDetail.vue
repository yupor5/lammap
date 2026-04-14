<template>
  <div class="space-y-6" v-loading="loading">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <el-button text :icon="ArrowLeft" @click="$router.back()">返回</el-button>
        <h1 class="text-2xl font-bold text-gray-800">{{ isNew ? '新建产品' : '产品详情' }}</h1>
      </div>
      <el-button type="primary" size="large" :loading="saving" @click="handleSave">保存修改</el-button>
    </div>

    <div class="grid grid-cols-2 gap-6">
      <!-- 基本信息 -->
      <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-6">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">基本信息</h3>
        <el-form label-width="80px">
          <el-form-item label="产品名"><el-input v-model="product.name" /></el-form-item>
          <el-form-item label="型号"><el-input v-model="product.sku" /></el-form-item>
          <el-form-item label="分类"><el-input v-model="product.category" /></el-form-item>
          <el-form-item label="简介"><el-input v-model="product.description" type="textarea" :rows="3" /></el-form-item>
        </el-form>
      </div>

      <!-- 规格参数 -->
      <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-6">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">规格参数</h3>
        <el-form label-width="80px">
          <el-form-item label="尺寸"><el-input v-model="product.size" /></el-form-item>
          <el-form-item label="材质"><el-input v-model="product.material" /></el-form-item>
          <el-form-item label="颜色"><el-input v-model="product.color" /></el-form-item>
          <el-form-item label="工艺"><el-input v-model="product.process" /></el-form-item>
          <el-form-item label="包装"><el-input v-model="product.packaging" /></el-form-item>
        </el-form>
      </div>

      <!-- 报价规则 -->
      <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-6">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">报价规则</h3>
        <el-form label-width="100px">
          <el-form-item label="基础价格"><el-input-number v-model="product.price" :precision="2" /></el-form-item>
          <el-form-item label="MOQ"><el-input-number v-model="product.moq" /></el-form-item>
          <el-form-item label="默认交期"><el-input v-model="product.leadTime" /></el-form-item>
          <el-form-item label="付款条款"><el-input v-model="product.paymentTerms" /></el-form-item>
        </el-form>
      </div>

      <!-- 附件资料 -->
      <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-6">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">附件资料</h3>
        <el-upload
          v-model:file-list="uploadFiles"
          drag
          multiple
          :http-request="handleUpload"
          :limit="20"
          accept=".pdf,.png,.jpg,.jpeg,.webp,.doc,.docx,.xls,.xlsx"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">拖拽文件或 <em>点击上传</em></div>
          <template #tip><div class="el-upload__tip">支持图片、PDF、认证文件</div></template>
        </el-upload>

        <div v-if="attachments.length > 0" class="mt-4">
          <div class="text-sm text-gray-500 mb-2">已上传附件（{{ attachments.length }}）</div>
          <el-table :data="attachments" border size="small" max-height="260">
            <el-table-column prop="fileName" label="文件名" min-width="220">
              <template #default="{ row }">
                <a class="text-blue-600 hover:underline" :href="row.filePath" target="_blank" rel="noopener noreferrer">
                  {{ row.fileName }}
                </a>
              </template>
            </el-table-column>
            <el-table-column prop="fileSize" label="大小" width="100" align="right">
              <template #default="{ row }">{{ formatSize(row.fileSize) }}</template>
            </el-table-column>
            <el-table-column prop="createdAt" label="上传时间" width="140">
              <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button text size="small" type="danger" @click="handleDeleteAttachment(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProductStore } from '@/stores/product'
import { ElMessage } from 'element-plus'
import { ArrowLeft, UploadFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()

const loading = ref(false)
const saving = ref(false)
const isNew = computed(() => route.params.id === 'new')

const product = reactive({
  name: '', sku: '', category: '', description: '',
  size: '', material: '', color: '', process: '', packaging: '',
  price: 0, moq: 100, leadTime: '', paymentTerms: '',
})

const uploadFiles = ref<any[]>([])
const attachments = ref<any[]>([])

function formatTime(t: string) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 10)
}

function formatSize(sz: number) {
  const n = Number(sz) || 0
  if (n <= 0) return '-'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}

async function loadAttachments() {
  if (isNew.value) {
    attachments.value = []
    return
  }
  const res: any = await request.get('/attachments', { params: { productId: Number(route.params.id) } })
  attachments.value = res.data || []
}

async function loadProduct() {
  if (isNew.value) return
  loading.value = true
  try {
    const data = await productStore.fetchProduct(Number(route.params.id))
    Object.assign(product, data)
    await loadAttachments()
  } catch {
    ElMessage.error('加载产品失败')
    router.back()
  } finally {
    loading.value = false
  }
}

async function handleUpload(opt: any) {
  if (isNew.value) {
    ElMessage.warning('请先保存创建产品，再上传附件')
    return
  }
  const file = opt?.file as File
  if (!file) return
  const form = new FormData()
  form.append('file', file)
  form.append('productId', String(route.params.id))
  try {
    const res: any = await request.post('/attachments', form, { headers: { 'Content-Type': 'multipart/form-data' } })
    const att = res.data
    if (att) attachments.value = [att, ...attachments.value]
    ElMessage.success('上传成功')
  } catch (e) {
    console.error(e)
    ElMessage.error('上传失败')
  } finally {
    try {
      opt?.onSuccess?.()
    } catch {
      // ignore
    }
    // 刷新产品附件数（后端会回写 attachments 字段）
    try {
      const data = await productStore.fetchProduct(Number(route.params.id))
      Object.assign(product, data)
    } catch {
      // ignore
    }
  }
}

async function handleDeleteAttachment(row: any) {
  if (!row?.id) return
  try {
    await request.delete(`/attachments/${row.id}`)
    attachments.value = attachments.value.filter((x) => x.id !== row.id)
    ElMessage.success('已删除')
    try {
      const data = await productStore.fetchProduct(Number(route.params.id))
      Object.assign(product, data)
    } catch {
      // ignore
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('删除失败')
  }
}

async function handleSave() {
  if (!product.name.trim()) {
    ElMessage.warning('请输入产品名称')
    return
  }
  saving.value = true
  try {
    if (isNew.value) {
      const created = await productStore.createProduct({ ...product })
      ElMessage.success('产品已创建')
      router.replace(`/products/${created.id}`)
      await loadProduct()
    } else {
      await productStore.updateProduct(Number(route.params.id), { ...product })
      ElMessage.success('产品已更新')
    }
  } catch {
    // error handled in interceptor
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadProduct()
})
</script>
