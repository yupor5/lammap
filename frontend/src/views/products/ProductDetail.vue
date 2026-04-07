<template>
  <div class="space-y-6">
    <div class="flex items-center gap-3">
      <el-button text :icon="ArrowLeft" @click="$router.back()">返回</el-button>
      <h1 class="text-2xl font-bold text-gray-800">产品详情</h1>
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
        <el-upload drag multiple :auto-upload="false">
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">拖拽文件或 <em>点击上传</em></div>
          <template #tip><div class="el-upload__tip">支持图片、PDF、认证文件</div></template>
        </el-upload>
      </div>
    </div>

    <div class="flex justify-end">
      <el-button type="primary" size="large">保存修改</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ArrowLeft, UploadFilled } from '@element-plus/icons-vue'

const product = reactive({
  name: '不锈钢桌腿', sku: 'STL-70B', category: '五金', description: '高品质304不锈钢桌腿，适用于各类办公桌和餐桌',
  size: '70cm', material: 'SS304', color: '黑色', process: '粉末喷涂', packaging: '单独纸箱包装',
  price: 12.50, moq: 100, leadTime: '25-30天', paymentTerms: 'T/T 30% deposit',
})
</script>
