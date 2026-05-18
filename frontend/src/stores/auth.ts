import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, register as registerApi, type LoginData, type RegisterData } from '../api/auth'
import router from '../router'

export interface UserInfo {
  id: string
  tenant_id: string
  email: string
  name: string
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>('')
  const user = ref<UserInfo | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function loadFromStorage() {
    const storedToken = localStorage.getItem('crm0_token')
    const storedUser = localStorage.getItem('crm0_user')
    if (storedToken && storedUser) {
      token.value = storedToken
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        localStorage.removeItem('crm0_token')
        localStorage.removeItem('crm0_user')
      }
    }
  }

  async function login(data: LoginData) {
    const res = await loginApi(data)
    token.value = res.token
    user.value = {
      id: res.user.id,
      tenant_id: res.user.tenant_id,
      email: res.user.email,
      name: res.user.name,
      role: res.user.role
    }
    localStorage.setItem('crm0_token', res.token)
    localStorage.setItem('crm0_user', JSON.stringify(user.value))
    router.push('/dashboard')
  }

  async function register(data: RegisterData) {
    const res = await registerApi(data)
    token.value = res.token
    user.value = {
      id: res.user.id,
      tenant_id: res.user.tenant_id,
      email: res.user.email,
      name: res.user.name,
      role: res.user.role
    }
    localStorage.setItem('crm0_token', res.token)
    localStorage.setItem('crm0_user', JSON.stringify(user.value))
    router.push('/dashboard')
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('crm0_token')
    localStorage.removeItem('crm0_user')
    router.push('/login')
  }

  return {
    token,
    user,
    isAuthenticated,
    loadFromStorage,
    login,
    register,
    logout
  }
})
