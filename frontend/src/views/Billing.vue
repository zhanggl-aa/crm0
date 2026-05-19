<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useBillingStore } from '../stores/billing'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const billingStore = useBillingStore()

const currentPlan = computed(() => billingStore.info?.plan || 'free')

const plans = computed(() => [
  { key: 'free', name: t('billing.free'), price: t('billing.freePrice'), features: t('billing.freeFeatures') },
  { key: 'pro', name: t('billing.pro'), price: t('billing.proPrice'), features: t('billing.proFeatures') },
  { key: 'enterprise', name: t('billing.enterprise'), price: t('billing.enterprisePrice'), features: t('billing.enterpriseFeatures') }
])

async function handleUpgrade(planKey: string) {
  if (planKey === currentPlan.value) return
  try {
    await billingStore.checkout(planKey)
    ElMessage.success(t('billing.checkoutCreated'))
  } catch { /* handled */ }
}

async function handlePortal() {
  try {
    await billingStore.portal()
    ElMessage.success(t('billing.portalCreated'))
  } catch { /* handled */ }
}

onMounted(() => billingStore.fetchInfo())
</script>

<template>
  <div class="billing-page">
    <h2>{{ t('billing.title') }}</h2>

    <el-card v-if="billingStore.info" style="margin-bottom: 24px">
      <div style="display: flex; justify-content: space-between; align-items: center">
        <div>
          <h3>{{ t('billing.currentPlan') }}: {{ billingStore.info.plan }}</h3>
          <p style="color: #909399">Status: {{ billingStore.info.status }}</p>
        </div>
        <el-button v-if="billingStore.info.plan !== 'free'" @click="handlePortal">{{ t('billing.manageBilling') }}</el-button>
      </div>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="8" v-for="plan in plans" :key="plan.key">
        <el-card shadow="hover" class="plan-card" :class="{ 'plan-card--current': plan.key === currentPlan }">
          <div class="plan-name">{{ plan.name }}</div>
          <div class="plan-price">{{ plan.price }}</div>
          <p class="plan-features">{{ plan.features }}</p>
          <el-button
            :type="plan.key === currentPlan ? 'info' : 'primary'"
            :disabled="plan.key === currentPlan"
            style="width: 100%"
            @click="handleUpgrade(plan.key)"
          >
            {{ plan.key === currentPlan ? t('billing.current') : t('billing.upgrade') }}
          </el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.billing-page { padding: 20px; }
.plan-card { text-align: center; padding: 20px 0; }
.plan-card--current { border: 2px solid #409eff; }
.plan-name { font-size: 20px; font-weight: 600; margin-bottom: 8px; }
.plan-price { font-size: 28px; font-weight: 700; color: #409eff; margin-bottom: 12px; }
.plan-features { color: #909399; font-size: 13px; margin-bottom: 20px; line-height: 1.8; }
</style>
