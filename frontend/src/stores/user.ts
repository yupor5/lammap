import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface UserInfo {
  name: string
  email?: string
  company?: string
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)

  async function login(email: string, _password: string) {
    userInfo.value = { name: email.split('@')[0] || '用户', email }
    localStorage.setItem('token', 'dev-token')
  }

  async function register(payload: { name: string; company: string; email: string; password: string }) {
    userInfo.value = {
      name: payload.name,
      email: payload.email,
      company: payload.company,
    }
    localStorage.setItem('token', 'dev-token')
  }

  function logout() {
    userInfo.value = null
    localStorage.removeItem('token')
  }

  return { userInfo, login, register, logout }
})
