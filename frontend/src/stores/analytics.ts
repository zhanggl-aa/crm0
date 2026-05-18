import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  getChurnPredictions as getChurnApi,
  triggerChurnPrediction as triggerChurnApi,
  getSegments as getSegmentsApi,
  triggerSegmentation as triggerSegmentationApi,
  getLTVPredictions as getLTVApi,
  getChannelROI as getChannelROIApi,
  getNBARecommendations as getNBAApi,
  triggerNBA as triggerNBAApi,
  getDashboard as getDashboardApi,
  type ChurnPrediction,
  type SegmentData,
  type LTVPrediction,
  type ChannelROI,
  type NBARecommendation,
  type DashboardOverview
} from '../api/analytics'

export const useAnalyticsStore = defineStore('analytics', () => {
  const churnPredictions = ref<ChurnPrediction[]>([])
  const segments = ref<SegmentData[]>([])
  const ltvPredictions = ref<LTVPrediction[]>([])
  const channelROI = ref<ChannelROI[]>([])
  const nbaRecommendations = ref<NBARecommendation[]>([])
  const dashboard = ref<DashboardOverview | null>(null)
  const loading = ref(false)

  async function fetchChurn() {
    loading.value = true
    try {
      churnPredictions.value = await getChurnApi()
    } finally {
      loading.value = false
    }
  }

  async function triggerChurn() {
    loading.value = true
    try {
      const task = await triggerChurnApi()
      return task
    } finally {
      loading.value = false
    }
  }

  async function fetchSegments(type?: string) {
    loading.value = true
    try {
      segments.value = await getSegmentsApi(type ? { type } : undefined)
    } finally {
      loading.value = false
    }
  }

  async function triggerSegmentation(type: string) {
    loading.value = true
    try {
      const task = await triggerSegmentationApi(type)
      return task
    } finally {
      loading.value = false
    }
  }

  async function fetchLTV() {
    loading.value = true
    try {
      ltvPredictions.value = await getLTVApi()
    } finally {
      loading.value = false
    }
  }

  async function fetchChannelROI() {
    loading.value = true
    try {
      channelROI.value = await getChannelROIApi()
    } finally {
      loading.value = false
    }
  }

  async function fetchNBA() {
    loading.value = true
    try {
      nbaRecommendations.value = await getNBAApi()
    } finally {
      loading.value = false
    }
  }

  async function triggerNBAAction(customerId?: string) {
    loading.value = true
    try {
      const task = await triggerNBAApi(customerId)
      return task
    } finally {
      loading.value = false
    }
  }

  async function fetchDashboard() {
    loading.value = true
    try {
      dashboard.value = await getDashboardApi()
    } finally {
      loading.value = false
    }
  }

  return {
    churnPredictions,
    segments,
    ltvPredictions,
    channelROI,
    nbaRecommendations,
    dashboard,
    loading,
    fetchChurn,
    triggerChurn,
    fetchSegments,
    triggerSegmentation,
    fetchLTV,
    fetchChannelROI,
    fetchNBA,
    triggerNBAAction,
    fetchDashboard
  }
})
