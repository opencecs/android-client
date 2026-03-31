<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="480px"
    :close-on-click-modal="!isMandatory && !isComplete"
    :close-on-press-escape="!isMandatory && !isComplete"
    :show-close="!isUpdating || isComplete"
    destroy-on-close
  >
    <div class="update-content">
      <div v-if="!isComplete">
        <div class="version-comparison">
          <div class="version-item current">
            <span class="version-label">当前版本</span>
            <span class="version-value">{{ currentVersion }}</span>
          </div>
          <el-icon class="arrow-icon"><Right /></el-icon>
          <div class="version-item latest">
            <span class="version-label">最新版本</span>
            <span class="version-value highlight">{{ updateInfo?.version }}</span>
          </div>
        </div>

        <div class="update-info">
          <div class="info-row">
            <span class="info-label">文件大小</span>
            <span class="info-value">{{ formatFileSize(updateInfo?.fileSize || 0) }}</span>
          </div>
          <div v-if="updateInfo?.mandatory" class="mandatory-badge">
            <el-icon><WarningFilled /></el-icon>
            <span>此版本为强制更新</span>
          </div>
        </div>

        <div class="release-notes" v-if="updateInfo?.releaseNotes">
          <div class="notes-header" @click="showNotes = !showNotes">
            <span>更新日志</span>
            <el-icon :class="{ 'is-expand': showNotes }">
              <ArrowDown />
            </el-icon>
          </div>
          <el-collapse-transition>
            <div v-show="showNotes" class="notes-content">
              <pre>{{ updateInfo.releaseNotes }}</pre>
            </div>
          </el-collapse-transition>
        </div>

        <div class="download-progress" v-if="isUpdating">
          <div class="progress-header">
            <span>下载进度</span>
            <span class="progress-percent">{{ downloadProgress.toFixed(1) }}%</span>
          </div>
          <el-progress
            :percentage="downloadProgress"
            :stroke-width="8"
            :status="downloadProgress >= 100 ? 'success' : ''"
          />
          <div class="progress-detail">
            <span>已下载: {{ formatFileSize(downloadedBytes) }}</span>
            <span v-if="totalBytes > 0">/ {{ formatFileSize(totalBytes) }}</span>
          </div>
        </div>

        <div class="error-message" v-if="errorMessage">
          <el-alert
            type="error"
            :closable="false"
            show-icon
          >
            {{ errorMessage }}
          </el-alert>
        </div>
      </div>

      <div v-else class="update-complete">
        <el-icon class="success-icon"><SuccessFilled /></el-icon>
        <h3>更新完成</h3>
        <p class="complete-message">新版本已安装完成</p>
        <p class="countdown-message">{{ countdownMessage }}</p>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer" v-if="!isUpdating && !isComplete">
        <el-button
          v-if="!isMandatory"
          @click="handleLater"
        >
          稍后提醒
        </el-button>
        <el-button
          type="primary"
          @click="handleUpdateNow"
        >
          {{ isDownloading ? '下载中...' : '立即更新' }}
        </el-button>
      </div>
      <div class="dialog-footer updating" v-else-if="isUpdating && !isComplete">
        <el-button
          v-if="!isMandatory && downloadProgress < 100"
          @click="handleCancel"
        >
          取消
        </el-button>
        <span class="updating-status" v-if="downloadProgress < 100">
          正在下载更新包...
        </span>
        <span class="updating-status" v-else>
          下载完成，正在准备安装...
        </span>
      </div>
      <div class="dialog-footer" v-else-if="isComplete">
        <span class="restarting-message">正在重启应用...</span>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { Right, WarningFilled, ArrowDown, SuccessFilled } from '@element-plus/icons-vue';
import { useUpdateService } from '../services/updateService.js';

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  updateInfo: {
    type: Object,
    default: null
  }
});

const emit = defineEmits(['update:visible', 'later', 'cancel']);

const {
  state,
  startUpdate,
  cancelUpdate,
  restartApp,
  formatFileSize
} = useUpdateService();

const showNotes = ref(false);
const countdown = ref(10);
let countdownTimer = null;

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const isMandatory = computed(() => props.updateInfo?.mandatory || false);
const currentVersion = computed(() => state.currentVersion);
const isUpdating = computed(() => state.isUpdating);
const isDownloading = computed(() => state.isUpdating && state.downloadProgress < 100);
const downloadProgress = computed(() => state.downloadProgress);
const errorMessage = computed(() => state.errorMessage);

const isComplete = computed(() => state.state === 'complete' || state.state === 'UpdateStateComplete');

const dialogTitle = computed(() => {
  if (isComplete.value) return '更新完成';
  return '发现新版本';
});

const countdownMessage = computed(() => {
  if (countdown.value > 0) {
    return `${countdown.value}秒后程序即将退出...`;
  }
  return '正在重启应用...';
});

const downloadedBytes = computed(() => {
  if (props.updateInfo?.fileSize) {
    return Math.floor(props.updateInfo.fileSize * (downloadProgress.value / 100));
  }
  return 0;
});

const totalBytes = computed(() => props.updateInfo?.fileSize || 0);

const startCountdown = () => {
  countdown.value = 10;
  countdownTimer = setInterval(async () => {
    countdown.value--;
    if (countdown.value <= 0) {
      clearInterval(countdownTimer);
      countdownTimer = null;
      try {
        await restartApp();
      } catch (error) {
        console.error('[UpdateDialog] 重启应用失败:', error);
      }
    }
  }, 1000);
};

watch(isComplete, (val) => {
  if (val) {
    startCountdown();
  } else {
    if (countdownTimer) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
    countdown.value = 10;
  }
});

watch(() => props.visible, (val) => {
  if (val) {
    showNotes.value = false;
  }
});

const handleLater = () => {
  dialogVisible.value = false;
  emit('later');
};

const handleUpdateNow = async () => {
  await startUpdate();
};

const handleCancel = async () => {
  await cancelUpdate();
  dialogVisible.value = false;
  emit('cancel');
};

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer);
  }
});

defineExpose({
  dialogVisible,
  showNotes,
  isMandatory,
  currentVersion,
  isUpdating,
  isDownloading,
  downloadProgress,
  errorMessage,
  downloadedBytes,
  totalBytes,
  isComplete,
  countdownMessage,
  updateInfo: () => props.updateInfo,
  formatFileSize,
  handleLater,
  handleUpdateNow,
  handleCancel
});
</script>

<style scoped>
.update-content {
  padding: 8px 0;
}

.version-comparison {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.version-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.version-label {
  font-size: 12px;
  color: #909399;
}

.version-value {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.version-value.highlight {
  color: #409eff;
}

.arrow-icon {
  color: #409eff;
  font-size: 20px;
}

.update-info {
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #ebeef5;
}

.info-label {
  color: #909399;
}

.info-value {
  color: #303133;
  font-weight: 500;
}

.mandatory-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 8px 12px;
  background: #fef0f0;
  color: #f56c6c;
  border-radius: 4px;
  font-size: 13px;
}

.release-notes {
  margin-top: 16px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: hidden;
}

.notes-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: #fafafa;
  cursor: pointer;
  user-select: none;
}

.notes-header span {
  font-weight: 500;
  color: #303133;
}

.notes-header .el-icon {
  transition: transform 0.2s;
}

.notes-header .el-icon.is-expand {
  transform: rotate(180deg);
}

.notes-content {
  padding: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.notes-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.6;
  color: #606266;
}

.download-progress {
  margin-top: 16px;
  padding: 12px;
  background: #f0f9eb;
  border-radius: 4px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
  color: #606266;
}

.progress-percent {
  font-weight: 600;
  color: #67c23a;
}

.progress-detail {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  text-align: right;
}

.error-message {
  margin-top: 16px;
}

.dialog-footer {
  display: flex;
  justify-content: center;
  gap: 12px;
}

.dialog-footer.updating {
  justify-content: flex-end;
}

.updating-status {
  color: #909399;
  font-size: 14px;
}

.update-complete {
  text-align: center;
  padding: 40px 20px;
}

.success-icon {
  font-size: 64px;
  color: #67c23a;
  margin-bottom: 16px;
}

.update-complete h3 {
  margin: 0 0 8px;
  font-size: 20px;
  color: #303133;
}

.complete-message {
  margin: 0 0 16px;
  font-size: 14px;
  color: #909399;
}

.countdown-message {
  margin: 0;
  font-size: 14px;
  color: #409eff;
  font-weight: 500;
}

.restarting-message {
  color: #909399;
  font-size: 14px;
}
</style>
