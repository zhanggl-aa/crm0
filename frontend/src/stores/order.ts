import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '../api/order'
import type { Order, OrderMetrics } from '../api/order'

export const useOrderStore = defineStore('order', () => {
  const orders = ref<Order[]>([])
  const total = ref(0)
  const metrics = ref<OrderMetrics | null>(null)
  const loading = ref(false)

  async function fetchOrders(params: { page?: number; page_size?: number; platform?: string; status?: string } = {}) {
    loading.value = true
    try {
      const res = await api.listOrders(params)
      orders.value = res.items || []
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  async function fetchMetrics() {
    try {
      metrics.value = await api.getOrderMetrics()
    } catch { /* ignore */ }
  }

  return { orders, total, metrics, loading, fetchOrders, fetchMetrics }
})
