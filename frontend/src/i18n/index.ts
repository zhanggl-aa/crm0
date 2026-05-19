import { createI18n } from 'vue-i18n'
import en from './en'
import zh from './zh'

export type MessageSchema = typeof en

const savedLocale = localStorage.getItem('crm0_locale')
const browserLang = navigator.language.startsWith('zh') ? 'zh' : 'en'
const defaultLocale = savedLocale || browserLang

const i18n = createI18n<[MessageSchema], 'en' | 'zh'>({
  legacy: false,
  locale: defaultLocale,
  fallbackLocale: 'en',
  messages: { en, zh }
})

export function setLocale(locale: 'en' | 'zh') {
  i18n.global.locale.value = locale
  localStorage.setItem('crm0_locale', locale)
}

export function getLocale(): 'en' | 'zh' {
  return i18n.global.locale.value as 'en' | 'zh'
}

export default i18n
