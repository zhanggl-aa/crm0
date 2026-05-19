<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIntegrationStore } from '../stores/integration'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const store = useIntegrationStore()

const platforms = computed(() => [
  { key: 'tiktok_shop', name: t('integrations.tiktokShop'), icon: '🎵', color: '#000000' },
  { key: 'shopify', name: t('integrations.shopify'), icon: '🛒', color: '#96bf48' },
  { key: 'amazon', name: t('integrations.amazon'), icon: '📦', color: '#ff9900' },
  { key: 'meta', name: t('integrations.meta'), icon: '📱', color: '#1877f2' }
])

function getStatus(key: string) {
  const ig = store.integrations.find(i => i.platform === key)
  return ig?.sync_status || 'disconnected'
}

function getLastSynced(key: string) {
  const ig = store.integrations.find(i => i.platform === key)
  return ig?.last_synced_at
}

function statusLabel(status: string) {
  if (status === 'connected') return t('integrations.connected')
  if (status === 'syncing') return t('integrations.syncing')
  return t('integrations.disconnected')
}

function statusType(status: string) {
  if (status === 'connected') return 'success'
  if (status === 'syncing') return 'warning'
  return 'info'
}

async function handleConnect(key: string) {
  try {
    await store.connect(key, 'demo_oauth_code')
    ElMessage.success(t('integrations.connectSuccess'))
  } catch { /* handled */ }
}

async function handleDisconnect(key: string) {
  try {
    await store.disconnect(key)
    ElMessage.success(t('integrations.disconnectSuccess'))
  } catch { /* handled */ }
}

async function handleSync(key: string) {
  try {
    await store.sync(key)
    ElMessage.success(t('integrations.syncStarted'))
  } catch { /* handled */ }
}

onMounted(() => store.fetchIntegrations())
</script>

<template>
  <div class="integrations-page">
    <h2>{{ t('integrations.title') }}</h2>

    <el-row :gutter="20" style="margin-top: 20px" v-loading="store.loading">
      <el-col :span="12" v-for="p in platforms" :key="p.key">
        <el-card shadow="hover" class="platform-card">
          <div class="platform-header">
            <span class="platform-icon">{{ p.icon }}</span>
            <div>
              <h3>{{ p.name }}</h3>
              <el-tag :type="statusType(getStatus(p.key))" size="small">{{ statusLabel(getStatus(p.key)) }}</el-tag>
            </div>
          </div>
          <p v-if="getLastSynced(p.key)" class="last-synced">{{ t('integrations.lastSynced') }}: {{ new Date(getLastSynced(p.key)!).toLocaleString() }}</p>
          <div class="platform-actions">
            <el-button v-if="getStatus(p.key) === 'disconnected'" type="primary" @click="handleConnect(p.key)">{{ t('integrations.connect') }}</el-button>
            <template v-else>
              <el-button type="success" @click="handleSync(p.key)" :disabled="getStatus(p.key) === 'syncing'">{{ t('integrations.sync') }}</el-button>
              <el-button type="danger" @click="handleDisconnect(p.key)">{{ t('integrations.disconnect') }}</el-button>
            </template>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.integrations-page { padding: 20px; }
.platform-card { margin-bottom: 16px; }
.platform-header { display: flex; align-items: center; gap: 16px; margin-bottom: 12px; }
.platform-icon { font-size: 40px; }
.platform-header h3 { margin: 0 0 4px; }
.last-synced { color: #909399; font-size: 12px; margin: 8px 0; }
.platform-actions { display: flex; gap: 8px; margin-top: 12px; }
</style>
