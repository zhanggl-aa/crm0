import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '../api/integration'
import type { PlatformIntegration } from '../api/integration'

export const useIntegrationStore = defineStore('integration', () => {
  const integrations = ref<PlatformIntegration[]>([])
  const loading = ref(false)

  async function fetchIntegrations() {
    loading.value = true
    try {
      const res = await api.listIntegrations()
      integrations.value = res.items || []
    } finally {
      loading.value = false
    }
  }

  async function connect(platform: string, code: string) {
    await api.connectPlatform(platform, code)
    await fetchIntegrations()
  }

  async function disconnect(platform: string) {
    await api.disconnectPlatform(platform)
    await fetchIntegrations()
  }

  async function sync(platform: string) {
    await api.triggerSync(platform)
    await fetchIntegrations()
  }

  return { integrations, loading, fetchIntegrations, connect, disconnect, sync }
})
