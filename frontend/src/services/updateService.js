import { reactive, onMounted, onUnmounted } from 'vue';
import * as UpdaterService from '../../bindings/edgeclient/updaterservice.js';

const state = reactive({
  state: 'idle',
  isChecking: false,
  isUpdating: false,
  hasUpdate: false,
  updateInfo: null,
  downloadProgress: 0,
  currentVersion: '',
  latestVersion: '',
  errorMessage: '',
  autoCheck: true,
  autoUpdate: false,
  checkInterval: 43200,
  lastCheckTime: '',
  channel: 'published'
});

let stateCheckTimer = null;

const startStatePolling = () => {
  if (stateCheckTimer) return;
  
  stateCheckTimer = setInterval(async () => {
    if (state.isUpdating) {
      try {
        const result = await UpdaterService.GetUpdateState();
        if (result) {
          if (result.state) state.state = result.state;
          if (result.progress !== undefined) state.downloadProgress = result.progress;
          
          console.log('[UpdateService] 更新状态:', result.state, '进度:', result.progress);
          
          if (result.state === 'complete' || result.state === 'UpdateStateComplete') {
            console.log('[UpdateService] 更新完成');
            state.isUpdating = false;
            state.state = 'complete';
            if (stateCheckTimer) {
              clearInterval(stateCheckTimer);
              stateCheckTimer = null;
            }
          }
        }
      } catch (error) {
        console.error('[UpdateService] 获取更新状态异常:', error);
      }
    }
  }, 500);
};

onMounted(() => {
  getUpdateConfig();
  getVersionInfo();
});

onUnmounted(() => {
  if (stateCheckTimer) {
    clearInterval(stateCheckTimer);
    stateCheckTimer = null;
  }
});

export function useUpdateService() {
  const checkForUpdate = async () => {
    state.isChecking = true;
    state.errorMessage = '';

    try {
      const result = await UpdaterService.CheckForUpdate();
      console.log('[UpdateService] 检查更新结果:', result);

      if (result.success) {
        if (result.hasUpdate) {
          state.hasUpdate = true;
          state.updateInfo = {
            version: result.version,
            downloadUrl: result.downloadUrl,
            checksum: result.checksum,
            fileSize: result.fileSize,
            releaseNotes: result.releaseNotes,
            mandatory: result.mandatory
          };
          state.latestVersion = result.version;
        } else {
          state.hasUpdate = false;
          state.updateInfo = null;
        }
      } else {
        state.errorMessage = result.message || '检查更新失败';
      }
    } catch (error) {
      console.error('[UpdateService] 检查更新异常:', error);
      state.errorMessage = error.message || '检查更新失败';
    } finally {
      state.isChecking = false;
    }

    return state;
  };

  const startUpdate = async () => {
    if (!state.hasUpdate || !state.updateInfo) {
      return { success: false, message: '没有可用的更新' };
    }

    state.isUpdating = true;
    state.state = 'downloading';
    state.errorMessage = '';

    try {
      const result = await UpdaterService.StartUpdate();
      console.log('[UpdateService] 开始更新结果:', result);

      if (result.success) {
        startStatePolling();
      } else {
        state.errorMessage = result.message || '更新失败';
        state.isUpdating = false;
        state.state = 'idle';
      }

      return result;
    } catch (error) {
      console.error('[UpdateService] 更新异常:', error);
      state.errorMessage = error.message || '更新失败';
      state.isUpdating = false;
      state.state = 'idle';
      throw error;
    }
  };

  const configureUpdate = async (config) => {
    try {
      const result = await UpdaterService.ConfigureUpdate(
        config.autoCheck,
        config.autoUpdate,
        config.checkInterval,
        config.channel,
        config.proxyURL || ''
      );
      console.log('[UpdateService] 配置更新结果:', result);

      if (result.success) {
        state.autoCheck = config.autoCheck;
        state.autoUpdate = config.autoUpdate;
        state.checkInterval = config.checkInterval;
        state.channel = config.channel;
      }

      return result;
    } catch (error) {
      console.error('[UpdateService] 配置更新异常:', error);
      throw error;
    }
  };

  const getUpdateConfig = async () => {
    try {
      const result = await UpdaterService.GetUpdateConfig();
      console.log('[UpdateService] 获取配置结果:', result);

      if (result) {
        state.autoCheck = result.autoCheck;
        state.autoUpdate = result.autoUpdate;
        state.checkInterval = result.checkInterval;
        state.channel = result.channel;
        state.lastCheckTime = result.lastCheckTime;
      }

      return result;
    } catch (error) {
      console.error('[UpdateService] 获取配置异常:', error);
      throw error;
    }
  };

  const getVersionInfo = async () => {
    try {
      const result = await UpdaterService.GetVersionInfo();
      console.log('[UpdateService] 获取版本信息结果:', result);

      if (result) {
        state.currentVersion = result.version || '未知';
      }

      return result;
    } catch (error) {
      console.error('[UpdateService] 获取版本信息异常:', error);
      state.currentVersion = '未知';
      throw error;
    }
  };

  const cancelUpdate = async () => {
    try {
      const result = await UpdaterService.CancelUpdate();
      console.log('[UpdateService] 取消更新结果:', result);
      state.isUpdating = false;
      return result;
    } catch (error) {
      console.error('[UpdateService] 取消更新异常:', error);
      throw error;
    }
  };

  const getUpdateState = async () => {
    try {
      const result = await UpdaterService.GetUpdateState();
      console.log('[UpdateService] 获取更新状态结果:', result);
      return result;
    } catch (error) {
      console.error('[UpdateService] 获取更新状态异常:', error);
      throw error;
    }
  };

  const getUpdateLog = async () => {
    try {
      const result = await UpdaterService.GetUpdateLog();
      console.log('[UpdateService] 获取更新日志结果:', result);
      return result;
    } catch (error) {
      console.error('[UpdateService] 获取更新日志异常:', error);
      throw error;
    }
  };

  const restartApp = async () => {
    try {
      const result = await UpdaterService.RestartApp();
      console.log('[UpdateService] 重启应用结果:', result);
      return result;
    } catch (error) {
      console.error('[UpdateService] 重启应用异常:', error);
      throw error;
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return {
    state,
    checkForUpdate,
    startUpdate,
    configureUpdate,
    getUpdateConfig,
    getVersionInfo,
    cancelUpdate,
    getUpdateState,
    getUpdateLog,
    restartApp,
    formatFileSize
  };
}
