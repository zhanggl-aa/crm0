<template>
  <div class="subscriptions-page">
    <div class="page-header">
      <h2>{{ t('subscriptions.title') }}</h2>
      <el-button type="primary" @click="showCreateDialog = true"><el-icon><Plus /></el-icon> {{ t('subscriptions.addSubscription') }}</el-button>
    </div>

    <el-row :gutter="16" style="margin-bottom: 20px">
      <el-col :span="6"><el-card shadow="never"><el-statistic :title="t('subscriptions.totalMRR')" :value="metrics.total_mrr" prefix="¥" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><el-statistic :title="t('subscriptions.activeSubscriptions')" :value="metrics.active_count" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><el-statistic :title="t('subscriptions.trialing')" :value="metrics.trial_count" /></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><el-statistic :title="t('subscriptions.canceled')" :value="metrics.canceled_count" /></el-card></el-col>
    </el-row>

    <el-table :data="subscriptions" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="100" show-overflow-tooltip />
      <el-table-column :label="t('subscriptions.customer')" min-width="150">
        <template #default="{ row }">{{ row.customer?.name || row.customer_id }}</template>
      </el-table-column>
      <el-table-column :label="t('subscriptions.plan')" min-width="150">
        <template #default="{ row }">{{ row.plan?.name || row.plan_id }}</template>
      </el-table-column>
      <el-table-column prop="status" :label="t('common.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : row.status === 'trial' ? 'warning' : 'danger'" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="mrr" label="MRR" width="120">
        <template #default="{ row }">¥{{ row.mrr }}</template>
      </el-table-column>
      <el-table-column prop="started_at" :label="t('subscriptions.startDate')" width="180" />
      <el-table-column :label="t('common.actions')" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="warning" @click="editSub(row)">{{ t('common.edit') }}</el-button>
          <el-popconfirm :title="t('subscriptions.cancelConfirm')" @confirm="handleCancel(row.id)">
            <template #reference><el-button link type="danger">{{ t('subscriptions.cancelSubscription') }}</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination v-model:current-page="page" :total="total" layout="total, prev, pager, next" style="margin-top: 16px; justify-content: flex-end" @current-change="loadData" />

    <el-dialog v-model="showCreateDialog" :title="editingSub ? t('subscriptions.editSubscription') : t('subscriptions.addSubscription')" width="450">
      <el-form :model="form" label-width="80px">
        <el-form-item :label="t('subscriptions.customerId')" v-if="!editingSub"><el-input v-model="form.customer_id" /></el-form-item>
        <el-form-item :label="t('subscriptions.planId')" v-if="!editingSub"><el-input v-model="form.plan_id" /></el-form-item>
        <el-form-item :label="t('common.status')" v-if="editingSub">
          <el-select v-model="form.status">
            <el-option :label="t('subscriptions.trial')" value="trial" />
            <el-option :label="t('common.active')" value="active" />
            <el-option :label="t('subscriptions.canceled')" value="canceled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { list, create, update, remove, getMetrics } from '../api/subscription'
import type { Subscription, SubscriptionMetrics } from '../api/subscription'

const { t } = useI18n()
const subscriptions = ref<Subscription[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const metrics = ref<SubscriptionMetrics>({ total_mrr: 0, active_count: 0, canceled_count: 0, trial_count: 0, avg_mrr_per_customer: 0 })
const showCreateDialog = ref(false)
const editingSub = ref<Subscription | null>(null)
const saving = ref(false)
const form = ref({ customer_id: '', plan_id: '', status: 'trial' })

const loadData = async () => {
  loading.value = true
  try {
    const [listRes, metricsRes] = await Promise.all([list({ page: page.value, page_size: 20 }), getMetrics()])
    subscriptions.value = listRes.items
    total.value = listRes.total
    metrics.value = metricsRes
  } finally { loading.value = false }
}

const editSub = (s: Subscription) => {
  editingSub.value = s
  form.value.status = s.status
  showCreateDialog.value = true
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingSub.value) {
      await update(editingSub.value.id, { status: form.value.status })
      ElMessage.success(t('subscriptions.subscriptionUpdated'))
    } else {
      await create({ customer_id: form.value.customer_id, plan_id: form.value.plan_id })
      ElMessage.success(t('subscriptions.subscriptionCreated'))
    }
    showCreateDialog.value = false
    editingSub.value = null
    loadData()
  } catch { ElMessage.error(t('common.saveFailed')) } finally { saving.value = false }
}

const handleCancel = async (id: string) => {
  await remove(id)
  ElMessage.success(t('subscriptions.subscriptionCanceled'))
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.subscriptions-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
