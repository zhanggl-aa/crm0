import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '../api/billing'
import type { BillingInfo } from '../api/billing'

export const useBillingStore = defineStore('billing', () => {
  const info = ref<BillingInfo | null>(null)
  const loading = ref(false)

  async function fetchInfo() {
    loading.value = true
    try {
      info.value = await api.getBillingInfo()
    } finally {
      loading.value = false
    }
  }

  async function checkout(priceId: string) {
    return api.createCheckout(priceId)
  }

  async function portal() {
    return api.createPortalSession()
  }

  return { info, loading, fetchInfo, checkout, portal }
})
