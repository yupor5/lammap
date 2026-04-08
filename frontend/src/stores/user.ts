import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'

export interface UserInfo {
  id: number
  email: string
  name: string
  company: string
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)
  const token = ref(localStorage.getItem('token') || '')

  async function login(email: string, password: string) {
    const res: any = await request.post('/auth/login', { email, password })
    token.value = res.data.token
    localStorage.setItem('token', res.data.token)
    userInfo.value = res.data.user
    return res
  }

  async function register(data: { email: string; password: string; name: string; company: string }) {
    const res: any = await request.post('/auth/register', data)
    token.value = res.data.token
    localStorage.setItem('token', res.data.token)
    userInfo.value = res.data.user
    return res
  }

  async function fetchProfile() {
    try {
      const res: any = await request.get('/auth/profile')
      userInfo.value = res.data
      return res
    } catch {
      logout()
      throw new Error('获取用户信息失败')
    }
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  return { userInfo, token, login, register, fetchProfile, logout }
})
