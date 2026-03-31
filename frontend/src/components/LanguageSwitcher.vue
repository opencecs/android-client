<template>
  <el-dropdown trigger="click" @command="handleCommand" class="language-switcher">
    <el-button type="primary" size="default" style="padding: 8px 10px;">
      <!-- <el-icon><Connection /></el-icon> -->
      {{ currentLanguageLabel }}
    </el-button>
    <template #dropdown>
      <el-dropdown-menu class="language-dropdown">
        <el-dropdown-item 
          command="zh-CN" 
          :class="{ 'is-active': currentLocale === 'zh-CN' }"
        >
          <span class="language-option">
            <!-- <span class="flag-icon">🇨🇳</span> -->
            <span class="language-name">CN/简体中文</span>
            <svg v-if="currentLocale === 'zh-CN'" class="check-icon" viewBox="0 0 16 16" fill="none">
              <path d="M13.5 4.5L6 12L2.5 8.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
        </el-dropdown-item>
        <el-dropdown-item 
          command="en-US" 
          :class="{ 'is-active': currentLocale === 'en-US' }"
        >
          <span class="language-option">
            <!-- <span class="flag-icon">🇺🇸</span> -->
            <span class="language-name">US/English</span>
            <svg v-if="currentLocale === 'en-US'" class="check-icon" viewBox="0 0 16 16" fill="none">
              <path d="M13.5 4.5L6 12L2.5 8.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { computed, getCurrentInstance } from 'vue'
import { ElMessage } from 'element-plus'
import { Connection } from '@element-plus/icons-vue'

const { proxy } = getCurrentInstance()

const currentLocale = computed(() => proxy.$i18n.getLocale())

const currentLanguageLabel = computed(() => {
  return currentLocale.value === 'zh-CN' ? 'CN/简体中文' : 'US/English'
})

const handleCommand = (command) => {
  if (command === currentLocale.value) {
    return
  }
  
  proxy.$i18n.setLocale(command)
  
  const message = command === 'zh-CN' 
    ? '语言已切换为简体中文' 
    : 'Language switched to English'
  
  ElMessage.success(message)
}
</script>

<style scoped>
.language-switcher {
  margin-left: 0;
}

.language-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
  width: 100%;
}

.flag-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.language-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
}

.check-icon {
  width: 16px;
  height: 16px;
  color: #409eff;
  flex-shrink: 0;
  animation: checkIn 0.3s ease;
}

@keyframes checkIn {
  0% {
    opacity: 0;
    transform: scale(0.5);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

::deep(.language-dropdown) {
  padding: 8px 4px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(102, 126, 234, 0.1);
}

::deep(.el-dropdown-menu__item) {
  padding: 10px 16px;
  border-radius: 8px;
  margin: 2px 0;
  transition: all 0.2s ease;
}

::deep(.el-dropdown-menu__item:hover) {
  background-color: #f5f7fa;
}

::deep(.el-dropdown-menu__item.is-active) {
  background: linear-gradient(135deg, #ecf5ff 0%, #e8f4ff 100%);
  color: #409eff;
  font-weight: 500;
}
</style>
