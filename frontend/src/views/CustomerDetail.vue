<template>
  <div class="customer-detail" v-loading="customerStore.loading">
    <el-page-header @back="router.push('/customers')" :title="customer?.name || ''" :content="t('customers.customerDetail')" />

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card :header="t('customerDetail.basicInfo')">
          <el-descriptions :column="2" border v-if="customer">
            <el-descriptions-item :label="t('common.name')">{{ customer.name }}</el-descriptions-item>
            <el-descriptions-item :label="t('common.email')">{{ customer.email }}</el-descriptions-item>
            <el-descriptions-item :label="t('common.company')">{{ customer.company || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('common.phone')">{{ customer.phone || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('common.status')">
              <el-tag :type="customer.status === 'active' ? 'success' : customer.status === 'inactive' ? 'warning' : 'danger'" size="small">{{ customer.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('customers.acquiredChannel')">{{ customer.acquired_channel || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card :header="t('customerDetail.eventHistory')" style="margin-top: 16px">
          <el-timeline v-if="events.length">
            <el-timeline-item v-for="e in events" :key="e.id" :timestamp="e.occurred_at" placement="top">
              <el-card shadow="never">
                <h4>{{ e.event_type }}</h4>
                <p v-if="e.properties && Object.keys(e.properties).length" style="color: #909399; font-size: 13px">{{ JSON.stringify(e.properties) }}</p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
          <el-empty v-else :description="t('customerDetail.noEvents')" />
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card :header="t('customerDetail.smartInsights')">
          <div v-if="insight">
            <div class="insight-item">
              <span class="insight-label">{{ t('customerDetail.churnRisk') }}</span>
              <el-progress v-if="insight.churn_prediction" :percentage="Math.round(insight.churn_prediction.risk_score * 100)" :color="churnColor(insight.churn_prediction.risk_score)" :stroke-width="18" :text-inside="true" />
              <span v-else style="color: #909399">{{ t('customerDetail.notPredicted') }}</span>
            </div>
            <div class="insight-item" v-if="insight.ltv">
              <span class="insight-label">{{ t('customerDetail.predictedLTV') }}</span>
              <span class="insight-value">¥{{ insight.ltv.predicted_ltv.toLocaleString() }}</span>
              <span style="color: #909399; font-size: 12px; margin-left: 8px">{{ t('customerDetail.confidence') }} {{ (insight.ltv.confidence * 100).toFixed(0) }}%</span>
            </div>
            <div class="insight-item" v-if="insight.segments?.length">
              <span class="insight-label">{{ t('customerDetail.customerSegments') }}</span>
              <el-tag v-for="s in insight.segments" :key="s.id" size="small" style="margin-right: 4px">{{ s.segment_name }}</el-tag>
            </div>
            <div class="insight-item" v-if="insight.recommendations?.length">
              <span class="insight-label">{{ t('customerDetail.recommendedActions') }}</span>
              <div v-for="r in insight.recommendations" :key="r.id" class="nba-item">
                <el-tag :type="r.action_type === 'call' ? 'danger' : r.action_type === 'discount' ? 'warning' : r.action_type === 'email' ? '' : 'success'" size="small">{{ r.action_type }}</el-tag>
                <span style="font-size: 13px; margin-left: 8px">{{ t('customerDetail.impact') }}: {{ (r.expected_impact * 100).toFixed(0) }}%</span>
              </div>
            </div>
          </div>
          <el-empty v-else :description="t('customerDetail.noInsights')" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useCustomerStore } from '../stores/customer'
import { getEvents } from '../api/customer'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const customerStore = useCustomerStore()
const customerId = route.params.id as string

const events = ref<any[]>([])
const insight = computed(() => customerStore.currentInsight)
const customer = computed(() => customerStore.currentCustomer)

const churnColor = (score: number) => score > 0.7 ? '#F56C6C' : score > 0.4 ? '#E6A23C' : '#67C23A'

onMounted(async () => {
  await customerStore.fetchCustomer(customerId)
  await customerStore.fetchInsights(customerId)
  try {
    const res = await getEvents(customerId, { page: 1, page_size: 20 })
    events.value = res.items
  } catch { /* ignore */ }
})
</script>

<style scoped>
.customer-detail { padding: 20px; }
.insight-item { margin-bottom: 16px; }
.insight-label { display: block; font-size: 13px; color: #909399; margin-bottom: 6px; }
.insight-value { font-size: 20px; font-weight: 600; color: #409EFF; }
.nba-item { display: flex; align-items: center; margin-bottom: 8px; }
</style>
