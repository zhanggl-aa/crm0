import api from './index'

export interface ChurnPrediction {
  id: string
  customer_id: string
  risk_score: number
  risk_level: string
  factors: Record<string, unknown>
  predicted_at: string
  customer?: {
    id: string
    name: string
    email: string
    company: string
  }
}

export interface SegmentData {
  id: string
  customer_id: string
  segment_type: string
  segment_name: string
  score: number
  updated_at: string
  customer?: {
    id: string
    name: string
    email: string
    company: string
  }
}

export interface LTVPrediction {
  id: string
  customer_id: string
  predicted_ltv: number
  confidence: number
  expected_lifetime_months: number
  model_version: string
  predicted_at: string
  customer?: {
    id: string
    name: string
    email: string
    company: string
  }
}

export interface ChannelROI {
  channel: string
  ltv: number
  cac: number
  ltv_cac_ratio: number
  customer_count: number
}

export interface NBARecommendation {
  id: string
  customer_id: string
  action_type: string
  action_detail: Record<string, unknown>
  expected_impact: number
  priority: number
  status: string
  created_at: string
  customer?: {
    id: string
    name: string
    email: string
  }
}

export interface AlgorithmTask {
  task_id: string
  status: string
  result?: unknown
}

export interface DashboardOverview {
  total_customers: number
  active_customers: number
  mrr: number
  churn_rate: number
  avg_ltv: number
  high_risk_count: number
  pending_actions: number
  recent_alerts: Array<{
    id: string
    type: string
    message: string
    severity: string
    time: string
  }>
}

export function getChurnPredictions(): Promise<ChurnPrediction[]> {
  return api.get('/analytics/churn/predictions')
}

export function triggerChurnPrediction(): Promise<AlgorithmTask> {
  return api.post('/analytics/churn/trigger-prediction')
}

export function getSegments(params?: { type?: string }): Promise<SegmentData[]> {
  return api.get('/analytics/segments', { params })
}

export function triggerSegmentation(type: string): Promise<AlgorithmTask> {
  return api.post('/analytics/segments/trigger-segmentation', { type })
}

export function getLTVPredictions(): Promise<LTVPrediction[]> {
  return api.get('/analytics/ltv/predictions')
}

export function getChannelROI(): Promise<ChannelROI[]> {
  return api.get('/analytics/ltv/channel-roi')
}

export function getNBARecommendations(): Promise<NBARecommendation[]> {
  return api.get('/analytics/nba/recommendations')
}

export function triggerNBA(customerId?: string): Promise<AlgorithmTask> {
  return api.post('/analytics/nba/trigger-nba', customerId ? { customer_id: customerId } : {})
}

export function getDashboard(): Promise<DashboardOverview> {
  return api.get('/dashboard')
}
