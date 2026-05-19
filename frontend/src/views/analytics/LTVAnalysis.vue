<template>
  <div class="ltv-page">
    <div class="page-header"><h2>{{ t('ltv.title') }}</h2></div>

    <el-row :gutter="20">
      <el-col :span="14">
        <el-card :header="t('ltv.customerLTVRanking')">
          <v-chart :option="ltvChartOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card :header="t('ltv.channelROI')">
          <v-chart :option="roiChartOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 16px">
      <el-col :span="14">
        <el-card :header="t('ltv.ltvDetails')">
          <el-table :data="analyticsStore.ltvPredictions" v-loading="analyticsStore.loading" stripe max-height="350">
            <el-table-column :label="t('subscriptions.customer')" min-width="140">
              <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
            </el-table-column>
            <el-table-column :label="t('ltv.predictedLTV')" width="130">
              <template #default="{ row }">¥{{ row.predicted_ltv.toLocaleString() }}</template>
            </el-table-column>
            <el-table-column :label="t('ltv.confidence')" width="100">
              <template #default="{ row }">{{ (row.confidence * 100).toFixed(0) }}%</template>
            </el-table-column>
            <el-table-column :label="t('ltv.expectedLifetime')" width="130">
              <template #default="{ row }">{{ row.expected_lifetime_months }}{{ t('ltv.months') }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card :header="t('ltv.channelROIDetails')">
          <el-table :data="analyticsStore.channelROI" stripe max-height="350">
            <el-table-column prop="channel" :label="t('ltv.channel')" min-width="100" />
            <el-table-column label="CAC" width="90">
              <template #default="{ row }">¥{{ row.cac }}</template>
            </el-table-column>
            <el-table-column label="LTV" width="90">
              <template #default="{ row }">¥{{ row.ltv }}</template>
            </el-table-column>
            <el-table-column label="LTV/CAC" width="90">
              <template #default="{ row }">{{ row.ltv_cac_ratio.toFixed(1) }}</template>
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
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { useAnalyticsStore } from '../../stores/analytics'

use([BarChart, TitleComponent, TooltipComponent, GridComponent, LegendComponent, CanvasRenderer])

const { t } = useI18n()
const analyticsStore = useAnalyticsStore()

const ltvChartOption = computed(() => {
  const top10 = [...analyticsStore.ltvPredictions].sort((a, b) => b.predicted_ltv - a.predicted_ltv).slice(0, 10)
  return {
    tooltip: {},
    xAxis: { type: 'category', data: top10.map(p => p.customer?.name || p.customer_id.slice(0, 8)), axisLabel: { rotate: 30 } },
    yAxis: { type: 'value', name: 'LTV (¥)' },
    series: [{ type: 'bar', data: top10.map(p => p.predicted_ltv), itemStyle: { color: '#409EFF' } }]
  }
})

const roiChartOption = computed(() => {
  const channels = analyticsStore.channelROI
  return {
    tooltip: {}, legend: {},
    xAxis: { type: 'category', data: channels.map(c => c.channel) },
    yAxis: { type: 'value', name: '¥' },
    series: [
      { name: 'CAC', type: 'bar', data: channels.map(c => c.cac), itemStyle: { color: '#F56C6C' } },
      { name: 'LTV', type: 'bar', data: channels.map(c => c.ltv), itemStyle: { color: '#67C23A' } }
    ]
  }
})

onMounted(() => { analyticsStore.fetchLTV(); analyticsStore.fetchChannelROI() })
</script>

<style scoped>
.ltv-page { padding: 20px; }
.page-header { margin-bottom: 16px; }
</style>
