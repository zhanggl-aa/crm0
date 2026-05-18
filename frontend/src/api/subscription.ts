import api from './index'

export interface SubscriptionListParams {
  page: number
  page_size: number
  status?: string
}

export interface Subscription {
  id: string
  customer_id: string
  plan_id: string
  status: string
  mrr: number
  started_at: string | null
  canceled_at: string | null
  created_at: string
  updated_at: string
  customer?: {
    id: string
    name: string
    email: string
  }
  plan?: {
    id: string
    name: string
    price: number
    billing_cycle: string
  }
}

export interface SubscriptionMetrics {
  total_mrr: number
  active_count: number
  canceled_count: number
  trial_count: number
  avg_mrr_per_customer: number
}

export interface CreateSubscriptionData {
  customer_id: string
  plan_id: string
}

export interface UpdateSubscriptionData {
  status?: string
  plan_id?: string
  canceled_at?: string
}

export interface SubscriptionListResponse {
  items: Subscription[]
  total: number
  page: number
  page_size: number
}

export function list(params: SubscriptionListParams): Promise<SubscriptionListResponse> {
  return api.get('/subscriptions', { params })
}

export function get(id: string): Promise<Subscription> {
  return api.get(`/subscriptions/${id}`)
}

export function create(data: CreateSubscriptionData): Promise<Subscription> {
  return api.post('/subscriptions', data)
}

export function update(id: string, data: UpdateSubscriptionData): Promise<Subscription> {
  return api.put(`/subscriptions/${id}`, data)
}

export function remove(id: string): Promise<void> {
  return api.delete(`/subscriptions/${id}`)
}

export function getMetrics(): Promise<SubscriptionMetrics> {
  return api.get('/subscriptions/metrics')
}
