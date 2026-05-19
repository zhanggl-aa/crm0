<template>
  <div class="nba-page">
    <div class="page-header">
      <h2>{{ t('nba.title') }}</h2>
      <el-button type="primary" @click="handleTrigger" :loading="analyticsStore.loading">{{ t('nba.generateRecommendations') }}</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="6" v-for="rec in analyticsStore.nbaRecommendations" :key="rec.id">
        <el-card shadow="hover" style="margin-bottom: 16px">
          <div class="nba-card">
            <div class="nba-header">
              <el-tag :type="actionType(rec.action_type)" size="large">{{ actionLabel(rec.action_type) }}</el-tag>
              <el-tag type="info" size="small">{{ t('nba.priority') }} {{ rec.priority }}</el-tag>
            </div>
            <h4 style="margin: 12px 0 8px">{{ rec.customer?.name || t('nba.customer') }}</h4>
            <p style="color: #909399; font-size: 13px; margin-bottom: 8px">{{ actionDetail(rec.action_detail) }}</p>
            <div class="nba-impact">
              <span>{{ t('nba.expectedImpact') }}</span>
              <el-progress :percentage="Math.round(rec.expected_impact * 100)" :stroke-width="12" style="flex: 1; margin-left: 8px" />
            </div>
            <div style="margin-top: 8px; display: flex; justify-content: space-between; align-items: center">
              <el-tag :type="rec.status === 'pending' ? 'warning' : rec.status === 'completed' ? 'success' : 'info'" size="small">{{ rec.status }}</el-tag>
              <span style="color: #C0C4CC; font-size: 12px">{{ rec.created_at?.slice(0, 10) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!analyticsStore.nbaRecommendations.length && !analyticsStore.loading" :description="t('nba.noRecommendations')" />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useAnalyticsStore } from '../../stores/analytics'

const { t } = useI18n()
const analyticsStore = useAnalyticsStore()

const actionType = (type: string) => type === 'call' ? 'danger' : type === 'discount' ? 'warning' : type === 'email' ? '' : 'success'
const actionLabel = (type: string) => type === 'call' ? t('nba.call') : type === 'discount' ? t('nba.discount') : type === 'email' ? t('nba.email') : type === 'feature_guide' ? t('nba.featureGuide') : type
const actionDetail = (d: Record<string, unknown>) => {
  if (!d || !Object.keys(d).length) return ''
  return Object.entries(d).map(([k, v]) => `${k}: ${v}`).join(', ')
}

const handleTrigger = async () => {
  try {
    await analyticsStore.triggerNBAAction()
    ElMessage.success(t('nba.recommendationsGenerated'))
    analyticsStore.fetchNBA()
  } catch { ElMessage.error(t('nba.generationFailed')) }
}

onMounted(() => analyticsStore.fetchNBA())
</script>

<style scoped>
.nba-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.nba-card { text-align: left; }
.nba-header { display: flex; justify-content: space-between; align-items: center; }
.nba-impact { display: flex; align-items: center; font-size: 13px; color: #606266; }
</style>
