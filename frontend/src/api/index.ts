import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import i18n from '../i18n'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('crm0_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response.data
  },
  (error) => {
    const t = i18n.global.t
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        localStorage.removeItem('crm0_token')
        localStorage.removeItem('crm0_user')
        router.push('/login')
        ElMessage.error(t('auth.loginExpired'))
      } else if (status === 403) {
        ElMessage.error(t('auth.noPermission'))
      } else if (status === 404) {
        ElMessage.error(t('auth.resourceNotFound'))
      } else if (status >= 500) {
        ElMessage.error(t('auth.serverError'))
      } else {
        const message = data?.message || data?.error || t('auth.requestFailed')
        ElMessage.error(message)
      }
    } else if (error.request) {
      ElMessage.error(t('auth.networkError'))
    } else {
      ElMessage.error(t('auth.sendFailed'))
    }
    return Promise.reject(error)
  }
)

export default api
