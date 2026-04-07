<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <el-button text :icon="ArrowLeft" @click="$router.back()">返回</el-button>
        <h1 class="text-2xl font-bold text-gray-800">报价详情 #{{ $route.params.id }}</h1>
        <el-tag type="info">草稿</el-tag>
      </div>
      <div class="flex gap-2">
        <el-button type="success" @click="handleDuplicate">基于此报价新建</el-button>
        <el-button type="primary">保存修改</el-button>
      </div>
    </div>
    <div class="bg-white rounded-lg border border-gray-100 shadow-sm p-8 text-center text-gray-400">
      <el-icon :size="64"><Document /></el-icon>
      <p class="mt-4 text-lg">报价详情页复用新建报价页布局（加载已有数据）</p>
      <p class="mt-2">后端对接后自动加载报价数据</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowLeft, Document } from '@element-plus/icons-vue'

const router = useRouter()

function handleDuplicate() {
  router.push('/quotes/new')
}
</script>
