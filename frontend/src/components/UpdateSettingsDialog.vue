<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('update.updateSettings')"
    width="480px"
    destroy-on-close
    append-to-body
  >
    <div class="settings-content">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-position="top"
      >
        <el-form-item label="更新渠道" prop="channel">
          <el-select
            v-model="formData.channel"
            placeholder="选择更新渠道"
            style="width: 100%"
          >
            <el-option
              v-for="channel in channels"
              :key="channel.key"
              :label="channel.name"
              :value="channel.key"
            >
              <span>{{ channel.name }}</span>
              <span class="channel-desc">{{ channel.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>

        <el-divider content-position="left">自动检查</el-divider>

        <el-form-item label="自动检查更新">
          <div class="form-switch-item">
            <div class="switch-label">
              <span class="label-text">启用自动检查</span>
              <span class="label-desc">在后台定期检查是否有新版本</span>
            </div>
            <el-switch
              v-model="formData.autoCheck"
              @change="handleAutoCheckChange"
            />
          </div>
        </el-form-item>

        <el-form-item v-if="formData.autoCheck" label="检查间隔" prop="checkInterval">
          <el-input-number
            v-model="formData.checkInterval"
            :min="300"
            :max="86400"
            :step="300"
            controls-position="right"
            style="width: 100%"
          />
          <div class="input-hint">单位：秒，建议设置不小于 300 秒（5 分钟）</div>
        </el-form-item>

        <el-divider content-position="left">自动下载</el-divider>

        <el-form-item label="自动下载更新">
          <div class="form-switch-item">
            <div class="switch-label">
              <span class="label-text">启用自动下载</span>
              <span class="label-desc">发现新版本后自动下载更新包</span>
            </div>
            <el-switch
              v-model="formData.autoUpdate"
              @change="handleAutoUpdateChange"
            />
          </div>
        </el-form-item>

        <el-form-item v-if="formData.autoUpdate && !formData.autoCheck" label="提示">
          <el-alert
            type="info"
            :closable="false"
            show-icon
          >
            建议同时开启自动检查更新，以便及时发现新版本
          </el-alert>
        </el-form-item>

        <el-divider content-position="left">检查记录</el-divider>

        <div class="check-record">
          <div class="record-item">
            <span class="record-label">上次检查时间</span>
            <span class="record-value">{{ lastCheckTimeFormatted }}</span>
          </div>
          <div class="record-item">
            <span class="record-label">当前版本</span>
            <span class="record-value">v{{ currentVersion }}</span>
          </div>
          <div class="record-item">
            <span class="record-label">最新版本</span>
            <span class="record-value">{{ latestVersion || '未知' }}</span>
          </div>
        </div>

        <el-form-item v-if="formData.autoCheck">
          <el-button
            type="primary"
            plain
            :loading="isChecking"
            @click="handleManualCheck"
          >
            <el-icon><Search /></el-icon>
            立即检查更新
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleReset">
          恢复默认
        </el-button>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="isSaving"
          @click="handleSave"
        >
          保存设置
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, getCurrentInstance } from 'vue';
import { Search } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useUpdateService } from '../services/updateService.js';

// 国际化支持
const { proxy } = getCurrentInstance()
const t = (key) => proxy.$i18n.t(key)

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  config: {
    type: Object,
    default: () => ({
      autoCheck: true,
      autoUpdate: false,
      checkInterval: 43200,
      channel: 'published'
    })
  }
});

const emit = defineEmits(['update:visible', 'save']);

const {
  state,
  checkForUpdate,
  configureUpdate,
  formatLastCheckTime
} = useUpdateService();

const formRef = ref(null);
const isSaving = ref(false);
const isChecking = ref(false);

const channels = ref([
  {
    key: 'published',
    name: '稳定版',
    description: '经过充分测试的稳定版本'
  },
  {
    key: 'test',
    name: '测试版',
    description: '包含新功能的测试版本'
  }
]);

const defaultConfig = {
  autoCheck: true,
  autoUpdate: false,
  checkInterval: 43200,
  channel: 'published'
};

const formData = ref({ ...defaultConfig });

const rules = {
  checkInterval: [
    { required: true, message: '请输入检查间隔', trigger: 'blur' },
    { type: 'number', min: 300, max: 86400, message: '检查间隔必须在 300-86400 秒之间', trigger: 'blur' }
  ]
};

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const currentVersion = computed(() => state.currentVersion);
const latestVersion = computed(() => state.latestVersion);
const lastCheckTimeFormatted = computed(() => formatLastCheckTime());

watch(() => props.visible, (val) => {
  if (val) {
    formData.value = {
      autoCheck: props.config.autoCheck ?? defaultConfig.autoCheck,
      autoUpdate: props.config.autoUpdate ?? defaultConfig.autoUpdate,
      checkInterval: props.config.checkInterval ?? defaultConfig.checkInterval,
      channel: props.config.channel ?? defaultConfig.channel
    };
  }
});

watch(() => formData.value.autoCheck, (val) => {
  if (!val) {
    formData.value.autoUpdate = false;
  }
});

const handleAutoCheckChange = (val) => {
  if (!val) {
    formData.value.autoUpdate = false;
  }
};

const handleAutoUpdateChange = (val) => {
  if (val && !formData.value.autoCheck) {
    ElMessage.warning('建议同时开启自动检查更新');
  }
};

const handleManualCheck = async () => {
  isChecking.value = true;
  await checkForUpdate();
  isChecking.value = false;

  if (state.hasUpdate) {
    ElMessage.success(`发现新版本: ${state.latestVersion}`);
  } else if (!state.errorMessage) {
    ElMessage.info('当前已是最新版本');
  } else if (state.errorMessage) {
    ElMessage.warning(state.errorMessage);
  }
};

const handleReset = () => {
  formData.value = { ...defaultConfig };
  ElMessage.info('已恢复默认设置');
};

const handleSave = async () => {
  isSaving.value = true;

  try {
    const success = await configureUpdate(
      formData.value.autoCheck,
      formData.value.autoUpdate,
      formData.value.checkInterval,
      formData.value.channel
    );

    if (success) {
      ElMessage.success('设置已保存');
      dialogVisible.value = false;
      emit('save', formData.value);
    } else {
      ElMessage.error('设置保存失败');
    }
  } catch (error) {
    console.error('保存设置异常:', error);
    ElMessage.error('设置保存失败');
  } finally {
    isSaving.value = false;
  }
};
</script>

<style scoped>
.settings-content {
  padding: 8px 0;
}

.form-switch-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.switch-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.label-text {
  font-size: 14px;
  color: #303133;
}

.label-desc {
  font-size: 12px;
  color: #909399;
}

.input-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.check-record {
  background: #f5f7fa;
  border-radius: 6px;
  padding: 12px 16px;
}

.record-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}

.record-item:not(:last-child) {
  border-bottom: 1px solid #ebeef5;
}

.record-label {
  color: #909399;
  font-size: 13px;
}

.record-value {
  color: #303133;
  font-size: 13px;
  font-weight: 500;
}

.channel-desc {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}

:deep(.el-select .el-option) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.el-divider__text) {
  font-size: 13px;
  color: #606266;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
