<template>
  <div class="churn-page">
    <div class="page-header">
      <h2>{{ t('churn.title') }}</h2>
      <el-button type="primary" @click="handleTrigger" :loading="analyticsStore.loading">{{ t('churn.triggerPrediction') }}</el-button>
    </div>

    <el-row :gutter="20">
      <el-col :span="10">
        <el-card :header="t('churn.riskDistribution')">
          <v-chart :option="chartOption" style="height: 300px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card :header="t('churn.customerChurnRisk')">
          <el-table :data="analyticsStore.churnPredictions" v-loading="analyticsStore.loading" stripe max-height="400">
            <el-table-column :label="t('subscriptions.customer')" min-width="140">
              <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
            </el-table-column>
            <el-table-column :label="t('churn.riskScore')" width="160">
              <template #default="{ row }">
                <el-progress :percentage="Math.round(row.risk_score * 100)" :color="riskColor(row.risk_score)" :stroke-width="16" :text-inside="true" />
              </template>
            </el-table-column>
            <el-table-column :label="t('churn.riskLevel')" width="100">
              <template #default="{ row }">
                <el-tag :type="row.risk_level === 'high' ? 'danger' : row.risk_level === 'medium' ? 'warning' : 'success'" size="small">{{ row.risk_level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('churn.keyFactors')" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="(val, key) in (row.factors || {})" :key="key" size="small" style="margin: 2px">{{ key }}: {{ (Number(val) * 100).toFixed(0) }}%</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { useAnalyticsStore } from '../../stores/analytics'

use([BarChart, TitleComponent, TooltipComponent, GridComponent, CanvasRenderer])

const { t } = useI18n()
const analyticsStore = useAnalyticsStore()

const riskColor = (score: number) => score > 0.7 ? '#F56C6C' : score > 0.4 ? '#E6A23C' : '#67C23A'

const chartOption = computed(() => {
  const high = analyticsStore.churnPredictions.filter(p => p.risk_level === 'high').length
  const medium = analyticsStore.churnPredictions.filter(p => p.risk_level === 'medium').length
  const low = analyticsStore.churnPredictions.filter(p => p.risk_level === 'low').length
  return {
    tooltip: {},
    xAxis: { type: 'category', data: [t('churn.highRisk'), t('churn.mediumRisk'), t('churn.lowRisk')] },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: [
      { value: high, itemStyle: { color: '#F56C6C' } },
      { value: medium, itemStyle: { color: '#E6A23C' } },
      { value: low, itemStyle: { color: '#67C23A' } }
    ]}]
  }
})

const handleTrigger = async () => {
  try {
    await analyticsStore.triggerChurn()
    ElMessage.success(t('churn.predictionComplete'))
    analyticsStore.fetchChurn()
  } catch { ElMessage.error(t('churn.predictionFailed')) }
}

onMounted(() => analyticsStore.fetchChurn())
</script>

<style scoped>
.churn-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
