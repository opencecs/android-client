<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('menu.about')"
    width="420px"
    destroy-on-close
    append-to-body
  >
    <div class="about-content">
      <div class="app-header">
        <img src="/src/assets/images/appicon.png" alt="Logo" class="app-logo" />
        <h2 class="app-name">{{ t('app.title') }}</h2>
        <p class="app-slogan">{{ t('app.description') }}</p>
      </div>

      <el-divider />

      <div class="version-info">
        <div class="info-row">
          <span class="info-label">{{ t('update.currentVersion') }}</span>
          <span class="info-value">v{{ version }}</span>
        </div>
        <div class="info-row" v-if="buildTime">
          <span class="info-label">{{ t('common.time') }}</span>
          <span class="info-value">{{ formatBuildTime }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">{{ t('common.type') }}</span>
          <span class="info-value">{{ platformInfo }}</span>
        </div>
      </div>

      <el-divider />

      <div class="links-section">
        <div class="link-item" @click="openLink(officialSite)">
          <el-icon><Link /></el-icon>
          <span>官方网站</span>
        </div>
        <div class="link-item" @click="openLink(documentation)">
          <el-icon><Document /></el-icon>
          <span>开发文档</span>
        </div>
      </div>

      <el-divider />

      <!-- <div class="copyright-section">
        <p class="copyright">© {{ currentYear }} 武汉魔云腾网络科技有限公司 版权所有</p>
        <p class="description">
          本软件基于开源项目开发，感谢所有贡献者的辛勤付出。
        </p>
      </div> -->

      <!-- <div class="open-source-credits">
        <el-collapse v-model="activeNames">
          <el-collapse-item title="开源组件致谢" name="1">
            <div class="credits-list">
              <div class="credit-item">
                <span class="credit-name">Wails</span>
                <span class="credit-desc">跨平台桌面应用框架</span>
              </div>
              <div class="credit-item">
                <span class="credit-name">Vue 3</span>
                <span class="credit-desc">前端框架</span>
              </div>
              <div class="credit-item">
                <span class="credit-name">Element Plus</span>
                <span class="credit-desc">UI组件库</span>
              </div>
              <div class="credit-item">
                <span class="credit-name">Go</span>
                <span class="credit-desc">后端编程语言</span>
              </div>
              <div class="credit-item">
                <span class="credit-name">Docker</span>
                <span class="credit-desc">容器运行时</span>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div> -->
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button
          type="primary"
          :loading="isChecking"
          @click="handleCheckUpdate"
        >
          <el-icon><Search /></el-icon>
          {{ isChecking ? t('update.downloading') : t('update.checkUpdate') }}
        </el-button>
        <el-button @click="dialogVisible = false">
          {{ t('common.close') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, getCurrentInstance } from 'vue';
import { Link, Document, Warning, Search } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useUpdateService } from '../services/updateService.js';

const { proxy } = getCurrentInstance()

// 响应式翻译函数 - 确保语言切换时重新计算
const t = (key, params) => {
  // 通过访问 proxy.$i18n.locale 建立响应式依赖
  const _ = proxy.$i18n.locale
  let text = proxy.$i18n.t(key)
  if (params) {
    Object.keys(params).forEach(param => {
      text = text.replace(`{${param}}`, params[param])
    })
  }
  return text
}

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  version: {
    type: String,
    default: '1.0.0'
  }
});

const emit = defineEmits(['update:visible', 'check-update']);

const {
  state,
  checkForUpdate,
  getVersionInfo
} = useUpdateService();

const activeNames = ref([]);
const currentYear = new Date().getFullYear();

const officialSite = ref('https://www.moyunteng.com');
const documentation = ref('https://dev.moyunteng.com');

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const buildTime = computed(() => {
  const buildTimeStr = state.currentVersion?.buildTime || '';
  return buildTimeStr;
});

const formatBuildTime = computed(() => {
  if (!buildTime.value) return '-';
  try {
    const date = new Date(buildTime.value);
    return date.toLocaleString('zh-CN');
  } catch {
    return buildTime.value;
  }
});

const platformInfo = computed(() => {
  const platform = navigator.platform.toLowerCase();
  const arch = navigator.userAgent.includes('x64') ? '64位' : '64位';
  return `${platform} ${arch}`;
});

const isChecking = computed(() => state.isChecking);

watch(() => props.visible, async (val) => {
  if (val && !state.currentVersion) {
    await getVersionInfo();
  }
});

const openLink = (url) => {
  if (url && typeof window !== 'undefined') {
    window.open(url, '_blank');
  }
};

const handleCheckUpdate = async () => {
  await checkForUpdate();

  if (state.hasUpdate) {
    ElMessage.success(t('update.newVersionAvailable') + `: ${state.latestVersion}`);
    emit('check-update', state.updateInfo);
  } else if (!state.errorMessage) {
    ElMessage.info(t('update.noUpdate'));
  } else if (state.errorMessage) {
    ElMessage.warning(state.errorMessage);
  }
};
</script>

<style scoped>
.about-content {
  padding: 8px 0;
}

.app-header {
  text-align: center;
  padding: 16px 0;
}

.app-logo {
  width: 80px;
  height: 80px;
  margin-bottom: 12px;
  border-radius: 16px;
  object-fit: contain;
}

.app-name {
  margin: 0 0 8px;
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.app-slogan {
  margin: 0;
  font-size: 13px;
  color: #909399;
}

.version-info {
  padding: 0 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.info-label {
  color: #909399;
  font-size: 14px;
}

.info-value {
  color: #303133;
  font-size: 14px;
  font-weight: 500;
}

.links-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.link-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: #f5f7fa;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.link-item:hover {
  background: #ecf5ff;
  color: #409eff;
}

.link-item .el-icon {
  font-size: 16px;
}

.link-item span {
  font-size: 14px;
}

.copyright-section {
  text-align: center;
  padding: 8px 0;
}

.copyright {
  margin: 0 0 8px;
  font-size: 13px;
  color: #606266;
}

.description {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.open-source-credits {
  margin-top: 8px;
}

.credits-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.credit-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}

.credit-name {
  font-weight: 500;
  color: #303133;
  font-size: 13px;
}

.credit-desc {
  color: #909399;
  font-size: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: center;
  gap: 12px;
}

.dialog-footer .el-button {
  min-width: 120px;
}
</style>
