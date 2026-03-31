<template>
  <div class="update-menu">
    <el-dropdown
      trigger="click"
      placement="bottom-end"
      @command="handleCommand"
    >
      <el-button
        class="update-menu-button"
        circle
        :disabled="isUpdating"
      >
        <el-icon size="18"><MoreFilled /></el-icon>
        <span v-if="hasUpdate" class="update-badge">new</span>
      </el-button>

      <template #dropdown>
        <el-dropdown-menu class="update-dropdown-menu">
          <el-dropdown-item command="check" :disabled="isChecking || isUpdating">
            <el-icon><Search /></el-icon>
            <span>{{ checkUpdateText }}</span>
            <span v-if="isChecking" class="menu-badge checking">{{ downloadingText }}</span>
            <span v-else-if="hasUpdate" class="menu-badge available">{{ hasNewVersionText }}</span>
          </el-dropdown-item>

          <el-dropdown-item command="about">
            <el-icon><InfoFilled /></el-icon>
            <span>{{ aboutText }}</span>
          </el-dropdown-item>

          <el-dropdown-divider />

          <el-dropdown-item command="settings">
            <div class="menu-switch-item">
              <div class="menu-switch-label">
                <el-icon><Clock /></el-icon>
                <span>{{ autoCheckUpdateText }}</span>
              </div>
              <el-switch
                v-model="autoCheckLocal"
                size="small"
                @click.stop
                @change="handleAutoCheckChange"
              />
            </div>
          </el-dropdown-item>

          <el-dropdown-item command="auto-update-settings">
            <div class="menu-switch-item">
              <div class="menu-switch-label">
                <el-icon><Download /></el-icon>
                <span>{{ autoDownloadUpdateText }}</span>
              </div>
              <el-switch
                v-model="autoUpdateLocal"
                size="small"
                @click.stop
                @change="handleAutoUpdateChange"
              />
            </div>
          </el-dropdown-item>

          <el-dropdown-divider />

          <!-- <el-dropdown-item command="settings-page">
            <el-icon><Setting /></el-icon>
            <span>更新设置</span>
          </el-dropdown-item> -->
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <AboutDialog
      v-model:visible="aboutDialogVisible"
      :version="currentVersion"
      @check-update="handleCheckUpdate"
    />

    <UpdateSettingsDialog
      v-model:visible="settingsDialogVisible"
      :config="settingsConfig"
      @save="handleSaveSettings"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, getCurrentInstance } from 'vue';
import { MoreFilled, Search, InfoFilled, Clock, Download, Setting } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import AboutDialog from './AboutDialog.vue';
import UpdateSettingsDialog from './UpdateSettingsDialog.vue';
import { useUpdateService } from '../services/updateService.js';

const { proxy } = getCurrentInstance()

// 响应式翻译函数
const t = (key, params) => {
  try {
    const _ = proxy.$i18n.locale
    let text = proxy.$i18n.t(key)
    if (params) {
      Object.keys(params).forEach(param => {
        text = text.replace(`{${param}}`, params[param])
      })
    }
    return text
  } catch (e) {
    console.warn('Translation error:', key, e)
    return key
  }
}

const emit = defineEmits(['check-update', 'show-update-dialog']);

const {
  state,
  checkForUpdate,
  startUpdate,
  configureUpdate,
  getVersionInfo,
  formatLastCheckTime
} = useUpdateService();

const aboutDialogVisible = ref(false);
const settingsDialogVisible = ref(false);
const autoCheckLocal = ref(state.autoCheck);
const autoUpdateLocal = ref(state.autoUpdate);

const settingsConfig = ref({
  autoCheck: state.autoCheck,
  autoUpdate: state.autoUpdate,
  checkInterval: state.checkInterval,
  channel: state.channel
});

const isChecking = computed(() => state.isChecking);
const isUpdating = computed(() => state.isUpdating);
const hasUpdate = computed(() => state.hasUpdate);
const currentVersion = computed(() => state.currentVersion);
const latestVersion = computed(() => state.latestVersion);

// 响应式翻译文本
const checkUpdateText = computed(() => t('update.checkUpdate'));
const aboutText = computed(() => t('update.about'));
const autoCheckUpdateText = computed(() => t('update.autoCheckUpdate'));
const autoDownloadUpdateText = computed(() => t('update.autoDownloadUpdate'));
const downloadingText = computed(() => t('update.downloading'));
const hasNewVersionText = computed(() => t('update.hasNewVersion'));

watch(() => state.autoCheck, (val) => {
  autoCheckLocal.value = val;
});

watch(() => state.autoUpdate, (val) => {
  autoUpdateLocal.value = val;
});

watch(() => [state.autoCheck, state.autoUpdate, state.checkInterval, state.channel], () => {
  settingsConfig.value = {
    autoCheck: state.autoCheck,
    autoUpdate: state.autoUpdate,
    checkInterval: state.checkInterval,
    channel: state.channel
  };
});

const handleCommand = (command) => {
  switch (command) {
    case 'check':
      handleCheckUpdate();
      break;
    case 'about':
      aboutDialogVisible.value = true;
      break;
    case 'settings':
      settingsDialogVisible.value = true;
      break;
    case 'settings-page':
      settingsDialogVisible.value = true;
      break;
    default:
      break;
  }
};

const handleCheckUpdate = async () => {
  await checkForUpdate();

  if (state.hasUpdate) {
    emit('show-update-dialog', state.updateInfo);
    ElMessage.success(`发现新版本: ${state.latestVersion}`);
  } else if (!state.errorMessage) {
    ElMessage.info('当前已是最新版本');
  } else if (state.errorMessage) {
    ElMessage.warning(state.errorMessage);
  }
};

const handleAutoCheckChange = async (value) => {
  const success = await configureUpdate(
    value,
    autoUpdateLocal.value,
    settingsConfig.value.checkInterval,
    settingsConfig.value.channel
  );

  if (success) {
    ElMessage.success(value ? '已开启自动检查更新' : '已关闭自动检查更新');
  } else {
    autoCheckLocal.value = !value;
    ElMessage.error('配置保存失败');
  }
};

const handleAutoUpdateChange = async (value) => {
  const success = await configureUpdate(
    autoCheckLocal.value,
    value,
    settingsConfig.value.checkInterval,
    settingsConfig.value.channel
  );

  if (success) {
    ElMessage.success(value ? '已开启自动下载更新' : '已关闭自动下载更新');
  } else {
    autoUpdateLocal.value = !value;
    ElMessage.error('配置保存失败');
  }
};

const handleSaveSettings = async (config) => {
  const success = await configureUpdate(
    config.autoCheck,
    config.autoUpdate,
    config.checkInterval,
    config.channel
  );

  if (success) {
    ElMessage.success('设置已保存');
    settingsDialogVisible.value = false;
  } else {
    ElMessage.error('设置保存失败');
  }
};

getVersionInfo();
</script>

<style scoped>
.update-menu {
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
}

.update-menu-button {
  width: 36px;
  height: 36px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: #909399;
  transition: all 0.2s;
  position: relative;
}

.update-menu-button:hover {
  background: #f5f7fa;
  color: #409eff;
}

.update-menu-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.update-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 32px;
  height: 16px;
  padding: 0 6px;
  background: #f56c6c;
  color: #fff;
  font-size: 10px;
  font-weight: bold;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

:deep(.update-dropdown-menu) {
  min-width: 200px;
  padding: 4px 0;
}

:deep(.update-dropdown-menu .el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  line-height: 1.5;
}

:deep(.update-dropdown-menu .el-dropdown-menu__item.is-disabled) {
  opacity: 0.6;
}

.menu-badge {
  margin-left: auto;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.menu-badge.checking {
  background: #e6f7ff;
  color: #1890ff;
}

.menu-badge.available {
  background: #f6ffed;
  color: #52c41a;
}

.menu-switch-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 12px;
}

.menu-switch-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

:deep(.el-dropdown-menu__item .el-switch) {
  flex-shrink: 0;
}
</style>
