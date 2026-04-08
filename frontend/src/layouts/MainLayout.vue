<template>
  <el-container class="h-screen">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="border-r border-gray-200 transition-all duration-300">
      <div class="flex items-center justify-center h-14 border-b border-gray-200">
        <span v-if="!isCollapse" class="text-lg font-bold text-blue-600">QuotePro</span>
        <span v-else class="text-lg font-bold text-blue-600">Q</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        router
        class="!border-r-0"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataBoard /></el-icon>
          <template #title>工作台</template>
        </el-menu-item>
        <el-menu-item index="/quotes/new">
          <el-icon><DocumentAdd /></el-icon>
          <template #title>新建报价</template>
        </el-menu-item>
        <el-menu-item index="/quotes/history">
          <el-icon><Document /></el-icon>
          <template #title>历史报价</template>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon><Goods /></el-icon>
          <template #title>产品资料库</template>
        </el-menu-item>
        <el-menu-item index="/templates">
          <el-icon><Files /></el-icon>
          <template #title>模板管理</template>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <template #title>设置</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="flex items-center justify-between border-b border-gray-200 bg-white h-14 px-4">
        <div class="flex items-center gap-3">
          <el-button :icon="isCollapse ? Expand : Fold" text @click="isCollapse = !isCollapse" />
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute">{{ currentRoute }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="flex items-center gap-3">
          <el-button type="primary" :icon="DocumentAdd" @click="$router.push('/quotes/new')">
            生成新报价
          </el-button>
          <el-dropdown>
            <div class="flex items-center gap-2 cursor-pointer">
              <el-avatar :size="32">{{ userInitial }}</el-avatar>
              <span class="text-sm text-gray-600">{{ userName }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$router.push('/settings')">设置</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="bg-gray-50 p-6">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { Expand, Fold, DocumentAdd, DataBoard, Document, Goods, Files, Setting } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isCollapse = ref(false)

onMounted(async () => {
  if (userStore.token && !userStore.userInfo) {
    try {
      await userStore.fetchProfile()
    } catch {
      // fetchProfile handles logout on failure
    }
  }
})

const activeMenu = computed(() => route.path)

const routeNameMap: Record<string, string> = {
  '/dashboard': '工作台',
  '/quotes/new': '新建报价',
  '/quotes/history': '历史报价',
  '/products': '产品资料库',
  '/templates': '模板管理',
  '/settings': '设置',
}

const currentRoute = computed(() => {
  const p = route.path
  if (routeNameMap[p]) return routeNameMap[p]
  const m = p.match(/^\/quotes\/(\d+)$/)
  if (m) return `报价详情 #${m[1]}`
  return ''
})

const userName = computed(() => userStore.userInfo?.name || '用户')
const userInitial = computed(() => (userStore.userInfo?.name || 'U').charAt(0).toUpperCase())

function handleLogout() {
  userStore.logout()
  router.push('/login')
}
</script>
