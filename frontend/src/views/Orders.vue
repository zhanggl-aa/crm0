<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOrderStore } from '../stores/order'

const { t } = useI18n()
const store = useOrderStore()
const page = ref(1)
const platformFilter = ref('')
const statusFilter = ref('')

function loadData() {
  store.fetchOrders({
    page: page.value,
    page_size: 20,
    platform: platformFilter.value || undefined,
    status: statusFilter.value || undefined
  })
  store.fetchMetrics()
}

onMounted(loadData)
</script>

<template>
  <div class="orders-page">
    <h2>{{ t('orders.title') }}</h2>

    <el-row :gutter="16" style="margin: 20px 0" v-if="store.metrics">
      <el-col :span="8">
        <el-card shadow="never">
          <el-statistic :title="t('orders.totalOrders')" :value="store.metrics.total_orders" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <el-statistic :title="t('orders.totalRevenue')" :value="store.metrics.total_revenue" prefix="$" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <el-statistic :title="t('orders.avgOrderValue')" :value="store.metrics.avg_order_value" prefix="$" />
        </el-card>
      </el-col>
    </el-row>

    <div style="display: flex; gap: 12px; margin-bottom: 16px">
      <el-select v-model="platformFilter" :placeholder="t('orders.platform')" clearable style="width: 150px" @change="loadData">
        <el-option label="TikTok Shop" value="tiktok_shop" />
        <el-option label="Shopify" value="shopify" />
        <el-option label="Amazon" value="amazon" />
        <el-option label="Meta" value="meta" />
      </el-select>
      <el-select v-model="statusFilter" :placeholder="t('orders.status')" clearable style="width: 150px" @change="loadData">
        <el-option label="Pending" value="pending" />
        <el-option label="Completed" value="completed" />
        <el-option label="Refunded" value="refunded" />
      </el-select>
    </div>

    <el-table :data="store.orders" v-loading="store.loading" stripe>
      <el-table-column prop="platform_order_id" :label="t('orders.orderId')" min-width="160" show-overflow-tooltip />
      <el-table-column prop="platform" :label="t('orders.platform')" width="130" />
      <el-table-column prop="status" :label="t('orders.status')" width="110">
        <template #default="{ row }">
          <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'pending' ? 'warning' : 'danger'" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('orders.total')" width="130">
        <template #default="{ row }">{{ row.currency }} {{ row.total }}</template>
      </el-table-column>
      <el-table-column prop="ordered_at" :label="t('orders.orderedAt')" width="180">
        <template #default="{ row }">{{ row.ordered_at ? new Date(row.ordered_at).toLocaleString() : '-' }}</template>
      </el-table-column>
    </el-table>

    <el-pagination v-model:current-page="page" :total="store.total" layout="total, prev, pager, next" style="margin-top: 16px; justify-content: flex-end" @current-change="loadData" />

    <el-empty v-if="!store.orders.length && !store.loading" :description="t('orders.noOrders')" />
  </div>
</template>

<style scoped>
.orders-page { padding: 20px; }
</style>
