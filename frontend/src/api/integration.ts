import api from './index'

export interface PlatformIntegration {
  id: string
  tenant_id: string
  platform: string
  sync_status: string
  last_synced_at?: string
  created_at: string
  updated_at: string
}

export function listIntegrations() {
  return api.get('/integrations') as Promise<{ items: PlatformIntegration[] }>
}

export function connectPlatform(platform: string, code: string) {
  return api.post('/integrations/connect', { platform, code })
}

export function disconnectPlatform(platform: string) {
  return api.post(`/integrations/${platform}/disconnect`)
}

export function triggerSync(platform: string) {
  return api.post(`/integrations/${platform}/sync`)
}
