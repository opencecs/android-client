import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

// 从localStorage获取保存的语言设置，默认为中文
const getDefaultLocale = () => {
  const savedLocale = localStorage.getItem('app-locale')
  if (savedLocale) {
    return savedLocale
  }
  
  // 如果没有保存的设置，根据浏览器语言判断
  const browserLang = navigator.language || navigator.userLanguage
  if (browserLang.startsWith('en')) {
    return 'en-US'
  }
  return 'zh-CN'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getDefaultLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  },
  globalInjection: true // 全局注入 $t 方法
})

// 切换语言的辅助函数
export const switchLocale = (locale) => {
  i18n.global.locale.value = locale
  localStorage.setItem('app-locale', locale)
  
  // 更新 Element Plus 的语言设置
  if (typeof window !== 'undefined') {
    document.documentElement.setAttribute('lang', locale)
  }
}

// 获取当前语言
export const getCurrentLocale = () => {
  return i18n.global.locale.value
}

export default i18n
