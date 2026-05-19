<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { setLocale, getLocale } from '../i18n'
import {
  Menu as IconMenu,
  Odometer,
  User,
  Tickets,
  DataAnalysis,
  Setting,
  SwitchButton,
  PieChart,
  Histogram,
  Opportunity,
  ShoppingCart,
  Link,
  CreditCard
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const isCollapse = ref(false)

const currentLocale = ref(getLocale())

const activeMenu = computed(() => {
  return route.path
})

const userName = computed(() => {
  return authStore.user?.name || authStore.user?.email || 'User'
})

const userRole = computed(() => {
  const role = authStore.user?.role || ''
  if (currentLocale.value === 'zh') {
    const roleMap: Record<string, string> = { admin: '管理员', manager: '经理', member: '成员' }
    return roleMap[role] || role
  }
  return role
})

function handleLogout() {
  authStore.logout()
}

function handleSelect(key: string) {
  router.push(key)
}

function toggleSidebar() {
  isCollapse.value = !isCollapse.value
}

function switchLocale(locale: 'en' | 'zh') {
  setLocale(locale)
  currentLocale.value = locale
}
</script>

<template>
  <el-container class="app-layout">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="app-aside">
      <div class="logo-area">
        <h1 v-if="!isCollapse" class="logo-text">CRM0</h1>
        <h1 v-else class="logo-text-mini">C0</h1>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="true"
        background-color="#1d1e1f"
        text-color="#bfcbd9"
        active-text-color="#409eff"
        router
        class="app-menu"
        @select="handleSelect"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>{{ t('nav.dashboard') }}</template>
        </el-menu-item>

        <el-menu-item index="/customers">
          <el-icon><User /></el-icon>
          <template #title>{{ t('nav.customers') }}</template>
        </el-menu-item>

        <el-menu-item index="/subscriptions">
          <el-icon><Tickets /></el-icon>
          <template #title>{{ t('nav.subscriptions') }}</template>
        </el-menu-item>

        <el-menu-item index="/orders">
          <el-icon><ShoppingCart /></el-icon>
          <template #title>{{ t('nav.orders') }}</template>
        </el-menu-item>

        <el-menu-item index="/integrations">
          <el-icon><Link /></el-icon>
          <template #title>{{ t('nav.integrations') }}</template>
        </el-menu-item>

        <el-sub-menu index="analytics">
          <template #title>
            <el-icon><DataAnalysis /></el-icon>
            <span>{{ t('nav.analytics') }}</span>
          </template>
          <el-menu-item index="/analytics/churn">
            <el-icon><PieChart /></el-icon>
            <template #title>{{ t('nav.churnPrediction') }}</template>
          </el-menu-item>
          <el-menu-item index="/analytics/segments">
            <el-icon><DataAnalysis /></el-icon>
            <template #title>{{ t('nav.customerSegments') }}</template>
          </el-menu-item>
          <el-menu-item index="/analytics/ltv">
            <el-icon><Histogram /></el-icon>
            <template #title>{{ t('nav.ltvAnalysis') }}</template>
          </el-menu-item>
          <el-menu-item index="/analytics/nba">
            <el-icon><Opportunity /></el-icon>
            <template #title>{{ t('nav.nbaRecommendations') }}</template>
          </el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/billing">
          <el-icon><CreditCard /></el-icon>
          <template #title>{{ t('nav.billing') }}</template>
        </el-menu-item>

        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <template #title>{{ t('nav.settings') }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container class="main-container">
      <el-header class="app-header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="toggleSidebar">
            <IconMenu />
          </el-icon>
          <span class="header-title">{{ t('nav.systemTitle') }}</span>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click" @command="switchLocale" style="margin-right: 16px">
            <span class="locale-switch">
              {{ currentLocale === 'zh' ? '中文' : 'EN' }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="en" :disabled="currentLocale === 'en'">English</el-dropdown-item>
                <el-dropdown-item command="zh" :disabled="currentLocale === 'zh'">中文</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown trigger="click">
            <span class="user-info">
              <el-icon><User /></el-icon>
              <span class="user-name">{{ userName }}</span>
              <el-tag size="small" type="info" class="user-role-tag">{{ userRole }}</el-tag>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/settings')">
                  <el-icon><Setting /></el-icon>{{ t('nav.settings') }}
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>{{ t('auth.logout') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.app-layout {
  height: 100vh;
  overflow: hidden;
}

.app-aside {
  background-color: #1d1e1f;
  transition: width 0.3s;
  overflow: hidden;
}

.logo-area {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #333;
}

.logo-text {
  color: #409eff;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: 2px;
  margin: 0;
}

.logo-text-mini {
  color: #409eff;
  font-size: 20px;
  font-weight: 700;
  margin: 0;
}

.app-menu {
  border-right: none;
  height: calc(100vh - 60px);
  overflow-y: auto;
}

.app-menu:not(.el-menu--collapse) {
  width: 220px;
}

.main-container {
  overflow: hidden;
}

.app-header {
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  transition: color 0.2s;
}

.collapse-btn:hover {
  color: #409eff;
}

.header-title {
  font-size: 16px;
  color: #303133;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
}

.locale-switch {
  cursor: pointer;
  color: #606266;
  font-size: 14px;
  font-weight: 500;
  padding: 4px 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.locale-switch:hover {
  color: #409eff;
  border-color: #409eff;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #606266;
  font-size: 14px;
}

.user-info:hover {
  color: #409eff;
}

.user-role-tag {
  margin-left: 4px;
}

.app-main {
  background: #f5f7fa;
  overflow-y: auto;
  padding: 20px;
}
</style>
