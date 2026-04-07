<template>
  <div class="min-h-screen flex">
    <!-- 左侧品牌展示 -->
    <div class="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-blue-600 to-blue-800 p-12 flex-col justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">QuotePro</h1>
        <p class="text-blue-200 mt-1">报价回复助手</p>
      </div>
      <div class="space-y-6">
        <h2 class="text-4xl font-bold text-white leading-tight">
          3 分钟生成<br />专业报价回复
        </h2>
        <p class="text-blue-100 text-lg leading-relaxed">
          客户发一句需求，自动生成报价单、交期说明、<br />参数确认清单和可发送附件。
        </p>
        <div class="grid grid-cols-2 gap-4 mt-8">
          <div class="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
            <div class="text-2xl font-bold text-white">5 分钟</div>
            <div class="text-blue-200 text-sm">平均出价时间</div>
          </div>
          <div class="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
            <div class="text-2xl font-bold text-white">95%</div>
            <div class="text-blue-200 text-sm">参数识别准确率</div>
          </div>
          <div class="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
            <div class="text-2xl font-bold text-white">3 种</div>
            <div class="text-blue-200 text-sm">回复话术版本</div>
          </div>
          <div class="bg-white/10 rounded-lg p-4 backdrop-blur-sm">
            <div class="text-2xl font-bold text-white">1 键</div>
            <div class="text-blue-200 text-sm">导出 PDF/Excel</div>
          </div>
        </div>
      </div>
      <p class="text-blue-300 text-sm">&copy; 2026 QuotePro. All rights reserved.</p>
    </div>

    <!-- 右侧登录表单 -->
    <div class="flex-1 flex items-center justify-center p-8 bg-white">
      <div class="w-full max-w-md">
        <div class="lg:hidden mb-8">
          <h1 class="text-2xl font-bold text-blue-600">QuotePro</h1>
          <p class="text-gray-500">报价回复助手</p>
        </div>

        <h2 class="text-2xl font-bold text-gray-800 mb-2">
          {{ isRegister ? '创建账号' : '欢迎回来' }}
        </h2>
        <p class="text-gray-500 mb-8">
          {{ isRegister ? '注册后即可开始使用' : '登录后开始高效报价' }}
        </p>

        <el-form ref="formRef" :model="form" :rules="rules" size="large">
          <el-form-item v-if="isRegister" prop="name">
            <el-input v-model="form.name" placeholder="姓名" prefix-icon="User" />
          </el-form-item>
          <el-form-item v-if="isRegister" prop="company">
            <el-input v-model="form.company" placeholder="公司名称" prefix-icon="OfficeBuilding" />
          </el-form-item>
          <el-form-item prop="email">
            <el-input v-model="form.email" placeholder="邮箱" prefix-icon="Message" />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              prefix-icon="Lock"
              show-password
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" class="w-full" :loading="loading" @click="handleSubmit">
              {{ isRegister ? '注册' : '登录' }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="text-center mt-4">
          <span class="text-gray-500">
            {{ isRegister ? '已有账号？' : '还没有账号？' }}
          </span>
          <el-button type="primary" link @click="isRegister = !isRegister">
            {{ isRegister ? '去登录' : '立即注册' }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const isRegister = ref(false)
const loading = ref(false)

const form = reactive({
  name: '',
  company: '',
  email: '',
  password: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  company: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      if (isRegister.value) {
        await userStore.register(form)
        ElMessage.success('注册成功')
      } else {
        await userStore.login(form.email, form.password)
        ElMessage.success('登录成功')
      }
      router.push('/dashboard')
    } catch {
      // error handled in interceptor
    } finally {
      loading.value = false
    }
  })
}
</script>
