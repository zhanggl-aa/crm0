<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAnalyticsStore } from '../stores/analytics'
import { ElMessage } from 'element-plus'
import {
  User,
  Money,
  Histogram,
  Warning,
  DataAnalysis,
  PieChart,
  Opportunity
} from '@element-plus/icons-vue'

const router = useRouter()
const { t } = useI18n()
const analyticsStore = useAnalyticsStore()

const dashboard = computed(() => analyticsStore.dashboard)
const loading = computed(() => analyticsStore.loading)

const alerts = computed(() => {
  return dashboard.value?.recent_alerts || []
})

const alertSeverityType = (severity: string) => {
  const map: Record<string, string> = {
    critical: 'danger',
    high: 'danger',
    medium: 'warning',
    low: 'info',
    info: 'info'
  }
  return map[severity] || 'info'
}

const alertSeverityLabel = (severity: string) => {
  const map: Record<string, string> = {
    critical: t('dashboard.critical'),
    high: t('dashboard.high'),
    medium: t('dashboard.medium'),
    low: t('dashboard.low'),
    info: t('dashboard.info')
  }
  return map[severity] || severity
}

async function triggerChurn() {
  try {
    await analyticsStore.triggerChurn()
    ElMessage.success(t('dashboard.churnTriggered'))
    analyticsStore.fetchDashboard()
  } catch {
    // handled by interceptor
  }
}

async function triggerSegmentation() {
  try {
    await analyticsStore.triggerSegmentation('rfm')
    ElMessage.success(t('dashboard.segmentationTriggered'))
    analyticsStore.fetchDashboard()
  } catch {
    // handled by interceptor
  }
}

function goToLTV() {
  router.push('/analytics/ltv')
}

function formatTime(time: string) {
  if (!time) return ''
  return new Date(time).toLocaleString()
}

onMounted(() => {
  analyticsStore.fetchDashboard()
})
</script>

<template>
  <div class="dashboard-page">
    <h2 class="page-title">{{ t('dashboard.title') }}</h2>

    <el-skeleton :loading="loading" animated :rows="5">
      <template #default>
        <!-- Stat cards -->
        <el-row :gutter="20" class="stat-row">
          <el-col :span="6">
            <el-card shadow="hover" class="stat-card stat-card--blue">
              <div class="stat-content">
                <div class="stat-info">
                  <div class="stat-label">{{ t('dashboard.totalCustomers') }}</div>
                  <div class="stat-value">{{ dashboard?.total_customers ?? 0 }}</div>
                </div>
                <el-icon class="stat-icon"><User /></el-icon>
              </div>
              <div class="stat-sub">{{ t('common.active') }}: {{ dashboard?.active_customers ?? 0 }}</div>
            </el-card>
          </el-col>

          <el-col :span="6">
            <el-card shadow="hover" class="stat-card stat-card--green">
              <div class="stat-content">
                <div class="stat-info">
                  <div class="stat-label">{{ t('dashboard.mrr') }}</div>
                  <div class="stat-value">¥{{ (dashboard?.mrr ?? 0).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</div>
                </div>
                <el-icon class="stat-icon"><Money /></el-icon>
              </div>
              <div class="stat-sub">{{ t('dashboard.monthlyRevenue') }}</div>
            </el-card>
          </el-col>

          <el-col :span="6">
            <el-card shadow="hover" class="stat-card stat-card--red">
              <div class="stat-content">
                <div class="stat-info">
                  <div class="stat-label">{{ t('dashboard.churnRate') }}</div>
                  <div class="stat-value">{{ ((dashboard?.churn_rate ?? 0) * 100).toFixed(1) }}%</div>
                </div>
                <el-icon class="stat-icon"><Histogram /></el-icon>
              </div>
              <div class="stat-sub">{{ t('dashboard.highRiskCustomers') }}: {{ dashboard?.high_risk_count ?? 0 }}</div>
            </el-card>
          </el-col>

          <el-col :span="6">
            <el-card shadow="hover" class="stat-card stat-card--orange">
              <div class="stat-content">
                <div class="stat-info">
                  <div class="stat-label">{{ t('dashboard.avgLTV') }}</div>
                  <div class="stat-value">¥{{ (dashboard?.avg_ltv ?? 0).toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</div>
                </div>
                <el-icon class="stat-icon"><Histogram /></el-icon>
              </div>
              <div class="stat-sub">{{ t('dashboard.pendingActions') }}: {{ dashboard?.pending_actions ?? 0 }}</div>
            </el-card>
          </el-col>
        </el-row>

        <!-- Bottom row -->
        <el-row :gutter="20" class="bottom-row">
          <el-col :span="14">
            <el-card shadow="hover" class="alerts-card">
              <template #header>
                <div class="card-header">
                  <span><el-icon><Warning /></el-icon> {{ t('dashboard.recentAlerts') }}</span>
                </div>
              </template>
              <div v-if="alerts.length === 0" class="empty-state">
                <el-empty :description="t('dashboard.noAlerts')" :image-size="80" />
              </div>
              <div v-else class="alerts-list">
                <div
                  v-for="alert in alerts"
                  :key="alert.id"
                  class="alert-item"
                >
                  <el-tag
                    :type="alertSeverityType(alert.severity)"
                    size="small"
                    class="alert-tag"
                  >
                    {{ alertSeverityLabel(alert.severity) }}
                  </el-tag>
                  <span class="alert-message">{{ alert.message }}</span>
                  <span class="alert-time">{{ formatTime(alert.time) }}</span>
                </div>
              </div>
            </el-card>
          </el-col>

          <el-col :span="10">
            <el-card shadow="hover" class="actions-card">
              <template #header>
                <div class="card-header">
                  <span><el-icon><DataAnalysis /></el-icon> {{ t('dashboard.quickActions') }}</span>
                </div>
              </template>
              <div class="quick-actions">
                <el-button
                  type="danger"
                  size="large"
                  class="action-btn"
                  @click="triggerChurn"
                >
                  <el-icon><PieChart /></el-icon>
                  {{ t('dashboard.triggerChurnPrediction') }}
                </el-button>
                <el-button
                  type="warning"
                  size="large"
                  class="action-btn"
                  @click="triggerSegmentation"
                >
                  <el-icon><DataAnalysis /></el-icon>
                  {{ t('dashboard.runSegmentation') }}
                </el-button>
                <el-button
                  type="success"
                  size="large"
                  class="action-btn"
                  @click="goToLTV"
                >
                  <el-icon><Histogram /></el-icon>
                  {{ t('dashboard.viewLTVReport') }}
                </el-button>
                <el-button
                  type="primary"
                  size="large"
                  class="action-btn"
                  @click="router.push('/analytics/nba')"
                >
                  <el-icon><Opportunity /></el-icon>
                  {{ t('dashboard.viewNBA') }}
                </el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </template>
    </el-skeleton>
  </div>
</template>

<style scoped>
.dashboard-page {
  max-width: 1400px;
}

.page-title {
  font-size: 22px;
  color: #303133;
  margin: 0 0 20px 0;
  font-weight: 600;
}

.stat-row {
  margin-bottom: 20px;
}

.stat-card {
  border-radius: 8px;
  border: none;
}

.stat-card--blue {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
  color: #fff;
}

.stat-card--green {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
  color: #fff;
}

.stat-card--red {
  background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
  color: #fff;
}

.stat-card--orange {
  background: linear-gradient(135deg, #e6a23c 0%, #ebb563 100%);
  color: #fff;
}

.stat-card :deep(.el-card__body) {
  padding: 20px;
}

.stat-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 13px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
}

.stat-icon {
  font-size: 48px;
  opacity: 0.3;
}

.stat-sub {
  font-size: 12px;
  opacity: 0.8;
  margin-top: 8px;
}

.bottom-row {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.alerts-list {
  max-height: 300px;
  overflow-y: auto;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.alert-item:last-child {
  border-bottom: none;
}

.alert-tag {
  flex-shrink: 0;
}

.alert-message {
  flex: 1;
  font-size: 14px;
  color: #606266;
}

.alert-time {
  flex-shrink: 0;
  font-size: 12px;
  color: #909399;
}

.empty-state {
  padding: 20px 0;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-btn {
  width: 100%;
  justify-content: flex-start;
  height: 48px;
  font-size: 15px;
}
</style>
