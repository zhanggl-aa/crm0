import api from './index'

export interface BillingInfo {
  tenant_id: string
  plan: string
  status: string
  stripe_customer_id: string
  current_period_end?: string
}

export function getBillingInfo() {
  return api.get('/billing/info') as Promise<BillingInfo>
}

export function createCheckout(priceId: string) {
  return api.post('/billing/checkout', { price_id: priceId })
}

export function createPortalSession() {
  return api.post('/billing/portal')
}
