import api from './index'

export function getOnboardingStatus() {
  return api.get('/onboarding/status')
}

export function completeOnboardingStep(step: number) {
  return api.post('/onboarding/complete-step', { step })
}

export function seedDemoData() {
  return api.post('/onboarding/seed-demo')
}

export function skipOnboarding() {
  return api.post('/onboarding/skip')
}
