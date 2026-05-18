import api from './index'

export interface CustomerListParams {
  page: number
  page_size: number
  search?: string
  status?: string
}

export interface Customer {
  id: string
  tenant_id: string
  name: string
  email: string
  company: string
  phone: string
  status: string
  tags: string[] | null
  custom_fields: Record<string, unknown> | null
  acquired_channel: string
  created_at: string
  updated_at: string
}

export interface CustomerInsight {
  customer: Customer
  churn_prediction: {
    id: string
    customer_id: string
    risk_score: number
    risk_level: string
    factors: Record<string, unknown>
    predicted_at: string
  } | null
  ltv: {
    id: string
    customer_id: string
    predicted_ltv: number
    confidence: number
    expected_lifetime_months: number
    model_version: string
    predicted_at: string
  } | null
  segments: Array<{
    id: string
    customer_id: string
    segment_type: string
    segment_name: string
    score: number
    updated_at: string
  }>
  recommendations: Array<{
    id: string
    customer_id: string
    action_type: string
    action_detail: Record<string, unknown>
    expected_impact: number
    priority: number
    status: string
    created_at: string
  }>
}

export interface CreateCustomerData {
  name: string
  email: string
  company?: string
  phone?: string
  status?: string
  tags?: string[]
  custom_fields?: Record<string, unknown>
  acquired_channel?: string
}

export interface UpdateCustomerData {
  name?: string
  email?: string
  company?: string
  phone?: string
  status?: string
  tags?: string[]
  custom_fields?: Record<string, unknown>
  acquired_channel?: string
}

export interface CustomerListResponse {
  items: Customer[]
  total: number
  page: number
  page_size: number
}

export function list(params: CustomerListParams): Promise<CustomerListResponse> {
  return api.get('/customers', { params })
}

export function get(id: string): Promise<Customer> {
  return api.get(`/customers/${id}`)
}

export function create(data: CreateCustomerData): Promise<Customer> {
  return api.post('/customers', data)
}

export function update(id: string, data: UpdateCustomerData): Promise<Customer> {
  return api.put(`/customers/${id}`, data)
}

export function remove(id: string): Promise<void> {
  return api.delete(`/customers/${id}`)
}

export function getInsights(id: string): Promise<CustomerInsight> {
  return api.get(`/customers/${id}/insights`)
}

export function getEvents(id: string, params?: { page?: number; page_size?: number }): Promise<{
  items: Array<{
    id: string
    customer_id: string
    event_type: string
    properties: Record<string, unknown>
    occurred_at: string
  }>
  total: number
}> {
  return api.get(`/customers/${id}/events`, { params })
}
