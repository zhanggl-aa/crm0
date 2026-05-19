<template>
  <div class="segment-page">
    <div class="page-header">
      <h2>{{ t('segments.title') }}</h2>
      <div>
        <el-select v-model="segmentType" style="width: 140px; margin-right: 12px">
          <el-option :label="t('segments.rfmSegments')" value="rfm" />
          <el-option :label="t('segments.behavioralSegments')" value="behavioral" />
          <el-option :label="t('segments.valueSegments')" value="value" />
        </el-select>
        <el-button type="primary" @click="handleTrigger" :loading="analyticsStore.loading">{{ t('segments.runSegmentation') }}</el-button>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="10">
        <el-card :header="t('segments.segmentDistribution')">
          <v-chart :option="chartOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card :header="t('segments.segmentResults')">
          <el-table :data="analyticsStore.segments" v-loading="analyticsStore.loading" stripe max-height="400">
            <el-table-column :label="t('subscriptions.customer')" min-width="140">
              <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
            </el-table-column>
            <el-table-column prop="segment_type" :label="t('segments.segmentType')" width="100" />
            <el-table-column prop="segment_name" :label="t('segments.segmentName')" width="120" />
            <el-table-column prop="score" :label="t('segments.score')" width="120">
              <template #default="{ row }">{{ row.score.toFixed(2) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { useAnalyticsStore } from '../../stores/analytics'

use([PieChart, TitleComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const { t } = useI18n()
const analyticsStore = useAnalyticsStore()
const segmentType = ref('rfm')

const chartOption = computed(() => {
  const counts: Record<string, number> = {}
  for (const s of analyticsStore.segments) { counts[s.segment_name] = (counts[s.segment_name] || 0) + 1 }
  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{ type: 'pie', radius: ['40%', '70%'], data: Object.entries(counts).map(([name, value]) => ({ name, value })) }]
  }
})

const handleTrigger = async () => {
  try {
    await analyticsStore.triggerSegmentation(segmentType.value)
    ElMessage.success(t('segments.segmentationComplete'))
    analyticsStore.fetchSegments(segmentType.value)
  } catch { ElMessage.error(t('segments.segmentationFailed')) }
}

onMounted(() => analyticsStore.fetchSegments(segmentType.value))
</script>

<style scoped>
.segment-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
