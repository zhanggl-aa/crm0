import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '../api/onboarding'

export const useOnboardingStore = defineStore('onboarding', () => {
  const step = ref(0)
  const completedSteps = ref<number[]>([])
  const demoDataSeeded = ref(false)
  const completed = ref(false)
  const loading = ref(false)

  async function fetchStatus() {
    loading.value = true
    try {
      const res = await api.getOnboardingStatus() as any
      step.value = res.step
      completedSteps.value = res.completed_steps || []
      demoDataSeeded.value = res.demo_data_seeded
      completed.value = res.completed
    } finally {
      loading.value = false
    }
  }

  async function completeStep(s: number) {
    loading.value = true
    try {
      const res = await api.completeOnboardingStep(s) as any
      step.value = res.step
      completedSteps.value = res.completed_steps || []
      completed.value = res.completed
    } finally {
      loading.value = false
    }
  }

  async function seedDemo() {
    loading.value = true
    try {
      const res = await api.seedDemoData() as any
      demoDataSeeded.value = res.demo_data_seeded
      step.value = res.step
      completedSteps.value = res.completed_steps || []
      completed.value = res.completed
    } finally {
      loading.value = false
    }
  }

  async function skip() {
    loading.value = true
    try {
      await api.skipOnboarding()
      completed.value = true
    } finally {
      loading.value = false
    }
  }

  return { step, completedSteps, demoDataSeeded, completed, loading, fetchStatus, completeStep, seedDemo, skip }
})
