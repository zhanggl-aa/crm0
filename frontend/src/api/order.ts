import api from './index'

export interface Order {
  id: string
  customer_id?: string
  integration_id: string
  tenant_id: string
  platform: string
  platform_order_id: string
  status: string
  currency: string
  total: number
  items: any[]
  ordered_at?: string
  created_at: string
  updated_at: string
}

export interface OrderMetrics {
  total_orders: number
  total_revenue: number
  avg_order_value: number
  by_platform: { platform: string; count: number; revenue: number }[]
}

export function listOrders(params: { page?: number; page_size?: number; platform?: string; status?: string }) {
  const query = new URLSearchParams()
  if (params.page) query.set('page', String(params.page))
  if (params.page_size) query.set('page_size', String(params.page_size))
  if (params.platform) query.set('platform', params.platform)
  if (params.status) query.set('status', params.status)
  return api.get(`/orders?${query}`) as Promise<{ items: Order[]; total: number; page: number; page_size: number }>
}

export function getOrder(id: string) {
  return api.get(`/orders/${id}`) as Promise<Order>
}

export function getOrderMetrics() {
  return api.get('/orders/metrics') as Promise<OrderMetrics>
}
