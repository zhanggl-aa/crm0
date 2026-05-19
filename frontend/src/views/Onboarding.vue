<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useOnboardingStore } from '../stores/onboarding'
import { ElMessage } from 'element-plus'

const router = useRouter()
const { t } = useI18n()
const store = useOnboardingStore()

onMounted(() => {
  store.fetchStatus()
})

async function handleNext() {
  await store.completeStep(store.step + 1)
}

async function handleSeedDemo() {
  try {
    await store.seedDemo()
    ElMessage.success(t('onboarding.demoSeeded'))
  } catch { /* handled */ }
}

async function handleSkip() {
  try {
    await store.skip()
    ElMessage.success(t('onboarding.skipped'))
    router.push('/dashboard')
  } catch { /* handled */ }
}

function handleFinish() {
  router.push('/dashboard')
}
</script>

<template>
  <div class="onboarding-container">
    <div class="onboarding-card">
      <h1 class="onboarding-title">CRM0</h1>
      <p class="onboarding-subtitle">{{ t('onboarding.subtitle') }}</p>

      <el-steps :active="store.step" align-center style="margin: 32px 0">
        <el-step :title="t('onboarding.step0Title')" :description="t('onboarding.step0Desc')" />
        <el-step :title="t('onboarding.step1Title')" :description="t('onboarding.step1Desc')" />
        <el-step :title="t('onboarding.step2Title')" :description="t('onboarding.step2Desc')" />
        <el-step :title="t('onboarding.step3Title')" :description="t('onboarding.step3Desc')" />
      </el-steps>

      <div class="step-content" v-loading="store.loading">
        <!-- Step 0: Welcome -->
        <div v-if="store.step === 0" class="step-panel">
          <h2>{{ t('onboarding.title') }}</h2>
          <p style="color: #909399; margin: 16px 0">{{ t('onboarding.subtitle') }}</p>
          <div class="step-actions">
            <el-button type="primary" size="large" @click="handleNext">{{ t('onboarding.next') }}</el-button>
            <el-button size="large" @click="handleSkip">{{ t('onboarding.skip') }}</el-button>
          </div>
        </div>

        <!-- Step 1: Connect Store -->
        <div v-if="store.step === 1" class="step-panel">
          <h2>{{ t('onboarding.step1Title') }}</h2>
          <p style="color: #909399; margin: 16px 0">{{ t('onboarding.step1Desc') }}</p>
          <el-row :gutter="16" style="margin: 20px 0">
            <el-col :span="12" v-for="p in ['tiktok_shop', 'shopify', 'amazon', 'meta']" :key="p">
              <el-card shadow="hover" style="margin-bottom: 12px; text-align: center; cursor: pointer">
                <h4>{{ p === 'tiktok_shop' ? t('integrations.tiktokShop') : p === 'shopify' ? t('integrations.shopify') : p === 'amazon' ? t('integrations.amazon') : t('integrations.meta') }}</h4>
              </el-card>
            </el-col>
          </el-row>
          <div class="step-actions">
            <el-button type="primary" size="large" @click="handleNext">{{ t('onboarding.next') }}</el-button>
            <el-button size="large" @click="handleNext">{{ t('onboarding.connectLater') }}</el-button>
          </div>
        </div>

        <!-- Step 2: Import Data -->
        <div v-if="store.step === 2" class="step-panel">
          <h2>{{ t('onboarding.step2Title') }}</h2>
          <p style="color: #909399; margin: 16px 0">{{ t('onboarding.step2Desc') }}</p>
          <el-button
            type="success"
            size="large"
            @click="handleSeedDemo"
            :loading="store.loading"
            :disabled="store.demoDataSeeded"
            style="margin: 20px 0"
          >
            {{ store.demoDataSeeded ? '✓' : '' }} {{ t('onboarding.seedDemoData') }}
          </el-button>
          <div class="step-actions">
            <el-button type="primary" size="large" @click="handleNext">{{ t('onboarding.next') }}</el-button>
          </div>
        </div>

        <!-- Step 3: Done -->
        <div v-if="store.step >= 3" class="step-panel">
          <el-result icon="success" :title="t('onboarding.step3Title')" :sub-title="t('onboarding.step3Desc')">
            <template #extra>
              <el-button type="primary" size="large" @click="handleFinish">{{ t('onboarding.goToDashboard') }}</el-button>
            </template>
          </el-result>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.onboarding-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.onboarding-card {
  width: 700px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}
.onboarding-title {
  text-align: center;
  font-size: 32px;
  font-weight: 700;
  color: #409eff;
  letter-spacing: 4px;
  margin: 0;
}
.onboarding-subtitle {
  text-align: center;
  color: #909399;
  margin: 8px 0 0;
}
.step-panel {
  text-align: center;
  min-height: 200px;
}
.step-actions {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  gap: 12px;
}
</style>
