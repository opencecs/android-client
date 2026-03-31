<template>
  <div class="model-management-container">
    <div class="model-content">
      <!-- 设备选择区域 -->
      <!-- <div class="device-select-content">
          <el-select 
            v-model="selectedDeviceId" 
            placeholder="请选择设备" 
            style="width: 300px; margin-right: 10px;"
            @change="handleDeviceChange"
          >
            <el-option 
              v-for="device in devices" 
              :key="device.id" 
              :label="device.ip" 
              :value="device.id"
            ></el-option>
          </el-select>
          <el-button 
            type="primary" 
            @click="fetchPhoneModels" 
            :loading="fetchingModels"
          >
            获取机型数据
          </el-button>
        </div> -->

      <!-- 机型列表 -->
      <el-card class="model-list-card">
        <!-- 提示信息 -->
        <el-alert
          :title="$t('model.simulatorOnlyHint')"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 15px;"
        />
        <!-- 筛选标签 -->
        <div class="model-filter-container">
          <div class="filter-item" :class="{ active: activeTab === 'online' }" @click="switchTab('online')">
            {{ $t('model.onlineModels') }}
          </div>
          <div class="filter-item" :class="{ active: activeTab === 'local' }" @click="switchTab('local')">
            {{ $t('model.localModels') }}
          </div>
          <div class="filter-item" :class="{ active: activeTab === 'help' }" @click="switchTab('help')">
            {{ $t('model.usageGuide') }}
          </div>
          <div v-if="activeTab !== 'help'" style="margin-left: auto; display: flex; align-items: center; gap: 10px; padding-right: 20px;">
            <el-button v-if="activeTab === 'local'" type="success" size="small" @click="handleCollectionTool">
              采集工具
            </el-button>
            <el-button v-if="activeTab === 'local'" type="default" size="small" @click="handleRefreshLocalModels">
              刷新
            </el-button>
            <el-button type="primary" size="small" @click="handleOpenDirectory">
              {{ $t(activeTab === 'local' ? 'model.openLocalDir' : 'model.openDownloadDir') }}
            </el-button>
          </div>
        </div>

        <!-- 批量操作栏（本地机型时显示） -->
        <div v-if="activeTab === 'local'" class="batch-action-bar">
          <el-checkbox
            :model-value="isAllSelected"
            :indeterminate="isIndeterminate"
            @change="handleSelectAll"
            style="margin-right: 12px;"
          >{{ $t('common.selectAllBtn') }}</el-checkbox>
          <span class="selected-count" v-if="selectedModels.length > 0">
            {{ $t('model.selectedCount', { count: selectedModels.length }) }}
          </span>
          <span class="selected-count" v-else style="color: #909399;">{{ $t('model.noModelSelected') }}</span>
          <el-button
            type="primary"
            size="small"
            :disabled="selectedModels.length === 0"
            @click="handleBatchPush"
            style="margin-left: 12px;"
          >
            {{ $t('model.pushSelected', { count: selectedModels.length }) }}
          </el-button>
        </div>

        <!-- 搜索框 -->
        <div v-if="activeTab !== 'help'" class="model-search-container" style="display: flex; gap: 15px; margin-bottom: 15px; align-items: flex-start;">
          <el-radio-group v-if="activeTab === 'online'" v-model="selectedAndroidVersion" @change="fetchPhoneModels" size="medium">
            <el-radio-button label="all">全部安卓版本</el-radio-button>
            <el-radio-button label="11">Android 11</el-radio-button>
            <el-radio-button label="13">Android 13</el-radio-button>
            <el-radio-button label="14">Android 14</el-radio-button>
            <el-radio-button label="15">Android 15</el-radio-button>
            <el-radio-button label="16">Android 16</el-radio-button>
          </el-radio-group>
          <el-input v-model="searchKeyword" :placeholder="$t('model.searchPlaceholder')" clearable size="medium"
            style="width: 300px"></el-input>
        </div>

        <!-- 机型列表 -->
        <div v-if="activeTab !== 'help'" class="model-list">
          <el-table
            :data="paginatedPhoneModels"
            stripe
            size="small"
            class="model-table"
            :row-class-name="getRowClassName"
            @row-click="handleRowClick"
          >
            <!-- 本地机型显示勾选列 -->
            <el-table-column v-if="activeTab === 'local'" width="50" align="center">
              <template #default="scope">
                <el-checkbox
                  :model-value="isModelSelected(scope.row)"
                  @change="(val) => toggleModelSelection(scope.row, val)"
                  @click.stop
                />
              </template>
            </el-table-column>
            <!-- 只在显示线上机型时显示机型ID -->
            <el-table-column v-if="activeTab !== 'local'" :label="$t('model.modelId')" prop="id" align="center"></el-table-column>
            <el-table-column :label="$t('model.modelName')" prop="name" align="center"></el-table-column>
            <el-table-column :label="$t('common.operation')" fixed="right" align="center">
              <template #default="scope">
                <div class="table-actions">
                  <el-button type="success" size="small" @click.stop="downloadTemplate(scope.row.id, scope.row.name)"
                    :loading="downloadingModelId === scope.row.id" v-if="showDownloadButton[scope.row.id]">
                    下载
                  </el-button>
                  <!-- 根据当前标签页显示不同按钮 -->
                  <template v-if="activeTab === 'local'">
                    <!-- 本地机型显示单独推送按钮 -->
                    <el-button type="primary" size="small" @click.stop="handleModelPush(scope.row)">
                      {{ $t('model.push') }}
                    </el-button>
                    <el-tag v-if="pushedModelNames.includes(scope.row.name)" type="success" size="small" style="margin-left: 4px;">{{ $t('model.pushed') }}</el-tag>
                  </template>
                  <template v-else>
                    <!-- 线上机型显示编辑按钮 -->
                    <el-button type="primary" size="small" @click.stop="handleModelEdit(scope.row)"
                      v-if="showEditButton[scope.row.id]">
                      编辑
                    </el-button>
                  </template>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="allPhoneModels.length === 0 && !fetchingModels" class="empty-state">
            <el-empty :description="$t('model.noModelData')" :image-size="100"></el-empty>
          </div>

          <!-- 分页组件 -->
          <div class="pagination-container" v-if="total > 0">
            <div class="pagination-wrapper">
              <el-pagination background layout="prev, pager, next, jumper" :current-page="currentPage"
                :page-size="pageSize" :total="total" @current-change="handleCurrentChange"
                @size-change="handleSizeChange"></el-pagination>
              <div class="total-text">{{ $t('model.totalModels') }}{{ total }}</div>
            </div>
          </div>
        </div>

        <!-- 使用说明 -->
        <div v-if="activeTab === 'help'" class="help-content">
          <div v-html="$t('model.helpContent')"></div>
        
        </div>
      </el-card>
    </div>
  </div>

  <!-- 机型配置编辑弹窗 -->
  <el-dialog v-model="configDialogVisible" width="70%" :before-close="closeConfigDialog" class="config-dialog"
    append-to-body>
    <template #header>
      <div class="dialog-header">
        <div class="dialog-title">
          {{ $t('model.configEdit', { name: editingModel?.name || '' }) }}
        </div>
        <div class="dialog-warning">
          {{ $t('model.configWarning') }}
        </div>
      </div>
    </template>
    <div class="dialog-content">
      <!-- 当前版本显示 -->
      <!-- <div class="version-info">
        当前版本: v4
      </div>
       -->
      <el-form :model="modelConfig" label-position="right" label-width="210px" class="config-form" inline>
        <!-- 顶层简单字段 -->
        <el-form-item :label="$t('model.templateName')" v-if="modelConfig.name">
          <el-input v-model="modelConfig.name" style="width: 300px" disabled></el-input>
        </el-form-item>

        <el-form-item label="supported_version" v-if="modelConfig.supported_version">
          <el-input v-model="modelConfig.supported_version" style="width: 300px" disabled></el-input>
        </el-form-item>

        <template v-for="(value, key) in otherFields" :key="`other-${key}`">
          <el-form-item :label="key">
            <el-input v-if="typeof value === 'string'" v-model="modelConfig[key]" style="width: 300px" @input="(val) => {
              const fullPath = key;
              if (regexPatterns[fullPath]) {
                modelConfig[key] = validateAndAdjustInput(val, regexPatterns[fullPath]);
              }
            }" disabled></el-input>
            <el-input-number v-else-if="typeof value === 'number'" v-model="modelConfig[key]" :min="-999999"
              :max="999999" style="width: 150px" disabled></el-input-number>

          </el-form-item>
        </template>

        <!-- 固定字段 (Overlay) 部分 -->
        <!-- <div class="section-header">固定字段 (Overlay)</div> -->
        <template v-for="(value, key) in overlayFields" :key="`overlay-${key}`">
          <el-form-item :label="formatFieldLabel(key)">
            <el-input v-if="typeof value === 'string'" v-model="modelConfig.overlay[key]" style="width: 470px" @input="(val) => {
              const fullPath = `overlay.${key}`;
              if (regexPatterns[fullPath]) {
                modelConfig.overlay[key] = validateAndAdjustInput(val, regexPatterns[fullPath]);
              }
            }"></el-input>
            <el-input-number v-else-if="typeof value === 'number'" v-model="modelConfig.overlay[key]" :min="-999999"
              :max="999999" style="width: 150px"></el-input-number>
            <el-switch v-else-if="typeof value === 'boolean'" v-model="modelConfig.overlay[key]"></el-switch>
          </el-form-item>
        </template>

        <!-- Prop 字段部分 -->
        <!-- <div class="section-header">Prop 字段</div> -->
        <template v-if="modelConfig.prop && typeof modelConfig.prop === 'object'">
          <template v-for="(value, key) in modelConfig.prop" :key="`prop-${key}`">
            <el-form-item :label="formatFieldLabel(key)">
              <el-input v-if="typeof value === 'string'" v-model="modelConfig.prop[key]" style="width: 300px" @input="(val) => {
                const fullPath = `prop.${key}`;
                if (regexPatterns[fullPath]) {
                  modelConfig.prop[key] = validateAndAdjustInput(val, regexPatterns[fullPath]);
                }
              }"></el-input>
              <el-input-number v-else-if="typeof value === 'number'" v-model="modelConfig.prop[key]" :min="-999999"
                :max="999999" style="width: 150px"></el-input-number>
              <el-switch v-else-if="typeof value === 'boolean'" v-model="modelConfig.prop[key]"></el-switch>
              <el-input v-else :value="JSON.stringify(value, null, 2)" type="textarea" :rows="4" style="width: 470px"
                @change="handlePropComplexChange(key, $event)"></el-input>
            </el-form-item>
          </template>
        </template>

        <!-- 其他顶层字段 -->
        <!-- <div class="section-header" v-if="otherFieldsCount > 0">其他字段</div> -->
        <!-- <template v-for="(value, key) in otherFields" :key="`other-${key}`">
          <el-form-item :label="key">
            <el-input
              v-if="typeof value === 'string'"
              v-model="modelConfig[key]"
              style="width: 300px"
            ></el-input>
            <el-input-number
              v-else-if="typeof value === 'number'"
              v-model="modelConfig[key]"
              :min="-999999"
              :max="999999"
              style="width: 150px"
            ></el-input-number>
            <el-switch
              v-else-if="typeof value === 'boolean'"
              v-model="modelConfig[key]"
            ></el-switch>
          </el-form-item>
        </template> -->
      </el-form>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="closeConfigDialog">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveConfigChanges">{{ $t('common.save') }}</el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 采集工具二维码弹窗 -->
  <el-dialog v-model="qrcodeDialogVisible" :title="$t('model.collectionToolDownload')" width="40%" :before-close="closeQrcodeDialog" append-to-body>
    <div class="qrcode-dialog-content">
      <div class="qrcode-description">
        {{ $t('model.scanQrToDownload') }}
      </div>
      <div class="qrcode-container">
        <canvas ref="qrcodeCanvas"></canvas>
      </div>
      <div class="qrcode-url">
        <p>{{ $t('model.collectionToolNote') }}</p>
        <p style="color: red;">{{ $t('model.collectionToolWarning1') }}</p>
        <p style="color: red;">{{ $t('model.collectionToolWarning2') }}</p>
        <!-- {{ collectionToolUrl }} -->

      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="closeQrcodeDialog">{{ $t('common.closeBtn') }}</el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 推送设备选择弹窗 -->
  <el-dialog v-model="pushDialogVisible" :title="$t('model.selectPushDevice')" width="60%" :before-close="closePushDialog" append-to-body>
    <div class="push-dialog-content">
      <!-- 已选机型展示 -->
      <div class="push-models-info">
        <span class="push-models-label">{{ $t('model.modelsToPush') }}</span>
        <div class="push-models-tags">
          <el-tag
            v-for="m in pushingModels"
            :key="m.name"
            type="primary"
            size="small"
            style="margin: 2px 4px 2px 0;"
          >{{ m.name }}</el-tag>
        </div>
      </div>
      <div class="dialog-header-bar">
        <div class="dialog-description">
          {{ $t('model.selectTargetDevices') }}
          <span class="device-count">({{ $t('model.onlineDeviceCount', { count: onlineDevices.length }) }})</span>
        </div>
        <el-button 
          type="primary" 
          size="small" 
          :icon="RefreshIcon" 
          @click="refreshOnlineDevices"
          :loading="refreshingDevices"
          :disabled="isPushing"
        >
          {{ $t('model.refreshList') }}
        </el-button>
      </div>

      <el-table :data="onlineDevices" stripe size="small" class="device-table" ref="deviceTable"
        @selection-change="handleDeviceSelectionChange">
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column prop="ip" :label="$t('model.deviceIP')" align="center"></el-table-column>
        <el-table-column :label="$t('common.status')" align="center" width="100">
          <template #default="scope">
            <el-tag type="success" size="small">{{ $t('common.online') }}</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <!-- 推送进度条 -->
      <div v-if="isPushing || pushProgress.total > 0" class="push-progress-area">
        <div class="push-progress-title">
          <span v-if="isPushing">{{ $t('model.pushing', { current: pushProgress.current, total: pushProgress.total }) }}</span>
          <span v-else-if="pushProgress.current === pushProgress.total && pushProgress.total > 0" style="color: #67c23a;">{{ $t('model.pushComplete', { total: pushProgress.total }) }}</span>
        </div>
        <el-progress
          :percentage="pushProgress.total > 0 ? Math.round((pushProgress.current / pushProgress.total) * 100) : 0"
          :status="isPushing ? '' : (pushProgress.hasError ? 'exception' : 'success')"
          :stroke-width="12"
        />
        <div v-if="pushProgress.currentModelName" class="push-progress-current">
          {{ $t('model.current') }}{{ pushProgress.currentModelName }}
        </div>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="closePushDialog" :disabled="isPushing">取消</el-button>
        <el-button type="primary" @click="confirmPush" :loading="isPushing">{{ $t('model.startPush') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { Refresh as RefreshIcon } from '@element-plus/icons-vue'
import QRCode from 'qrcode'
import {
  GetLocalModels,
  GetModelConfig,
  PushModelToDevices,
  SaveModelConfig,
  NeedShowDownloadButton,
  HasLocalTemplate,
  CheckModelButtonStatusBatch,
  DownloadTemplate,
  GetPhoneTemplates,
  OpenLocalModelDirectory,
  OpenDownloadModelDirectory
} from '../../bindings/edgeclient/app'

// 接收父组件传递的属性
const props = defineProps({
  devices: {
    type: Array,
    default: () => []
  },
  activeDevice: {
    type: Object,
    default: null
  },
  selectedHostDevices: {
    type: Array,
    default: () => []
  },
  devicesStatusCache: {
    type: Map,
    default: () => new Map()
  }
})

// 组件内部状态
// const selectedDeviceId = ref('')
const selectedDevice = ref(null)
const allPhoneModels = ref([]) // 存储所有数据，用于前端分页
const phoneModels = ref([]) // 存储当前页数据
const fetchingModels = ref(false)
const activeTab = ref('online') // 当前选中的标签页：online-线上机型，local-本级
const searchKeyword = ref('') // 搜索关键词
const selectedAndroidVersion = ref('all') // 安卓版本筛选

// 正则表达式相关状态和函数
const regexPatterns = ref({}) // 存储原始正则表达式，使用完整路径作为键

// 检测字符串是否为正则表达式
const isRegexPattern = (str) => {
  // 只识别当前代码可以处理的简单正则表达式格式：[字符集]{长度}
  return /^\[[^\]]+\]\{\d+\}$/.test(str);
};

// 检测字符串是否为复杂正则表达式（无法处理的格式）
const isComplexRegexPattern = (str) => {
  // 检测包含多个正则表达式片段或特殊字符的复杂格式
  return typeof str === 'string' && (
    // 包含多个[]{}组合（如 [0-9a-f]{8}-[0-9a-f]{4}）
    (str.match(/\[.*?\]\{\d+\}/g) || []).length > 1 ||
    // 包含多个{}组合
    (str.match(/\{\d+\}/g) || []).length > 1 ||
    // 包含[]但后面没有{}（如 [0-9a-f]{8}-[0-9a-f]）
    /\[.*?\]\D(?!\{\d+\})/.test(str)
  );
};

// 获取正则表达式匹配的字符集
const getRegexCharSet = (str) => {
  const match = str.match(/^\[([^\]]+)\]/);
  return match ? match[1] : '';
};

// 获取正则表达式的长度
const getRegexLength = (str) => {
  const match = str.match(/\{(\d+)\}$/);
  return match ? parseInt(match[1], 10) : 0;
};

// 转换正则表达式为星号显示
const convertRegexToAsterisks = (str) => {
  if (isRegexPattern(str)) {
    const length = getRegexLength(str);
    return '*'.repeat(length);
  }
  return str;
};

// 验证字符是否符合正则表达式字符集
const isValidCharForRegex = (char, regexPattern) => {
  const charSet = getRegexCharSet(regexPattern);
  if (!charSet) return false;

  // 简单处理常见字符集：0-9, a-f, A-Z等
  if (charSet === '0-9') {
    return /^[0-9]$/.test(char);
  } else if (charSet === '0-9a-f') {
    return /^[0-9a-f]$/.test(char);
  } else if (charSet === '0-9a-zA-Z') {
    return /^[0-9a-zA-Z]$/.test(char);
  } else if (charSet === 'a-zA-Z') {
    return /^[a-zA-Z]$/.test(char);
  }
  // 其他字符集暂时返回true，后续可以扩展
  return true;
};

// 验证并调整用户输入，确保符合正则要求和长度
const validateAndAdjustInput = (inputStr, originalRegex) => {
  if (!isRegexPattern(originalRegex)) {
    return inputStr;
  }

  const regexLength = getRegexLength(originalRegex);
  const charSet = getRegexCharSet(originalRegex);
  let adjustedInput = '';
  let hasInvalidChar = false;

  // 验证每个字符并调整长度
  for (let i = 0; i < regexLength; i++) {
    const char = inputStr[i] || '*';

    if (char === '*' || isValidCharForRegex(char, originalRegex)) {
      adjustedInput += char;
    } else {
      adjustedInput += '*';
      hasInvalidChar = true;
    }
  }

  // 如果有无效字符，显示错误提示
  if (hasInvalidChar) {
    ElMessage.warning(`请输入符合规则的字符 (${charSet})`);
  }

  return adjustedInput;
};

// 转换用户输入（带*）回正则表达式
const convertInputToRegex = (inputStr, originalRegex) => {
  if (!isRegexPattern(originalRegex)) {
    return inputStr;
  }

  const regexLength = getRegexLength(originalRegex);
  const charSet = getRegexCharSet(originalRegex);
  let result = '';
  let currentAsteriskCount = 0;

  // 遍历用户输入的每个字符，构建组合正则表达式
  for (let i = 0; i < regexLength; i++) {
    const char = inputStr[i] || '*';

    if (char === '*') {
      // 遇到星号，累加星号计数
      currentAsteriskCount++;
    } else if (isValidCharForRegex(char, originalRegex)) {
      // 遇到有效字符，先处理之前累积的星号
      if (currentAsteriskCount > 0) {
        result += `[${charSet}]{${currentAsteriskCount}}`;
        currentAsteriskCount = 0;
      }
      // 添加有效字符
      result += char;
    } else {
      // 遇到无效字符，转换为星号
      if (currentAsteriskCount > 0) {
        result += `[${charSet}]{${currentAsteriskCount}}`;
        currentAsteriskCount = 0;
      }
      result += `[${charSet}]{1}`;
    }
  }

  // 处理末尾可能剩余的星号
  if (currentAsteriskCount > 0) {
    result += `[${charSet}]{${currentAsteriskCount}}`;
  }

  return result;
};

// 下载相关状态
const downloadingModelId = ref('') // 存储当前正在下载的机型ID
const showDownloadButton = ref({}) // 存储每个机型是否显示下载按钮
const showEditButton = ref({}) // 存储每个机型是否显示编辑按钮

// 编辑相关状态
const editingModel = ref(null)
const editingModelName = ref('') // 存储原始机型名称
const modelConfig = ref({})
const configDialogVisible = ref(false)

// 推送相关状态
const pushDialogVisible = ref(false)
const selectedDevices = ref([]) // 选中的设备列表
const pushingModels = ref([]) // 当前正在推送的机型列表（支持批量）
const refreshingDevices = ref(false) // 刷新设备列表加载状态
const deviceTable = ref(null) // 设备选择表格 ref
const pushedModelNames = ref([]) // 本次会话已成功推送过的机型名称

// 推送进度状态
const isPushing = ref(false)
const pushProgress = ref({ current: 0, total: 0, currentModelName: '', hasError: false })

// 本地机型多选状态
const selectedModels = ref([]) // 已勾选的机型列表

// 全选状态
const isAllSelected = computed(() => {
  const localList = filteredPhoneModels.value
  return localList.length > 0 && selectedModels.value.length === localList.length
})

const isIndeterminate = computed(() => {
  return selectedModels.value.length > 0 && selectedModels.value.length < filteredPhoneModels.value.length
})

// 判断某个机型是否已选
const isModelSelected = (model) => {
  return selectedModels.value.some(m => m.name === model.name)
}

// 切换单个机型选中状态
const toggleModelSelection = (model, selected) => {
  if (selected) {
    if (!isModelSelected(model)) {
      selectedModels.value.push(model)
    }
  } else {
    selectedModels.value = selectedModels.value.filter(m => m.name !== model.name)
  }
}

// 全选/取消全选
const handleSelectAll = (val) => {
  if (val) {
    // 全选当前筛选结果
    selectedModels.value = [...filteredPhoneModels.value]
  } else {
    selectedModels.value = []
  }
}

// 点击行切换选中（仅本地机型）
const handleRowClick = (row) => {
  if (activeTab.value !== 'local') return
  toggleModelSelection(row, !isModelSelected(row))
}

// 获取行样式（高亮已选中机型 / 已推送机型）
const getRowClassName = ({ row }) => {
  if (activeTab.value === 'local') {
    if (isModelSelected(row)) return 'selected-model-row'
    if (pushedModelNames.value.includes(row.name)) return 'pushed-model-row'
  }
  return ''
}

// 切换标签时清空已选机型
const switchTab = (tab) => {
  activeTab.value = tab
  selectedModels.value = []
}

// 二维码相关状态
const qrcodeDialogVisible = ref(false)
const qrcodeCanvas = ref(null)
const collectionToolUrl = 'https://d.moyunteng.com/download/devinfo/devinfo_v2beta_20260206.apk'

// 分页相关状态
const currentPage = ref(1)
const pageSize = ref(12)

// 计算总条数，根据筛选结果动态变化
const total = computed(() => {
  return filteredPhoneModels.value.length
})

// 计算属性：筛选后的机型列表
const filteredPhoneModels = computed(() => {
  let models = []

  // 根据当前选中的标签页筛选机型
  if (activeTab.value === 'local') {
    // 本地机型：直接返回本地机型列表
    models = localModels.value
  } else if (activeTab.value === 'online') {
    // 线上机型：显示所有从API获取的机型
    models = allPhoneModels.value
  } else {
    // 使用说明等其他 tab：返回空列表
    return []
  }

  // 如果有搜索关键词，则根据名称过滤
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.trim().toLowerCase()
    models = models.filter(model =>
      model.name && model.name.toLowerCase().includes(keyword)
    )
  }

  return models
})

// 本地机型列表
const localModels = ref([]) // 存储本地机型列表

// 获取本地机型列表
const fetchLocalModels = async () => {
  try {
    const models = await GetLocalModels()
    localModels.value = Array.isArray(models) ? models : []
  } catch (error) {
    console.error('获取本地机型列表失败:', error)
    localModels.value = []
  } finally {
    // 过滤掉已不存在于本地列表中的已选机型，避免推送时报错
    const validNames = new Set(localModels.value.map(m => m.name))
    selectedModels.value = selectedModels.value.filter(m => validNames.has(m.name))
  }
}

// 计算属性：分页后的机型列表
const paginatedPhoneModels = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredPhoneModels.value.slice(start, end)
})

// 监听标签页切换，重置页码并更新本地机型列表
watch(activeTab, (newTab) => {
  currentPage.value = 1

  // 如果切换到本地机型标签页，重新获取本地机型列表
  if (newTab === 'local') {
    fetchLocalModels()
  }

  // 更新total值
  total.value = filteredPhoneModels.value.length
})

// 打开目录
const handleOpenDirectory = async () => {
  try {
    let result;
    if (activeTab.value === 'local') {
      result = await OpenLocalModelDirectory()
    } else {
      result = await OpenDownloadModelDirectory()
    }

    if (result.success) {
      ElMessage.success(activeTab.value === 'local' ? '已打开本地机型目录' : '已打开本地机型下载目录')
    } else {
      ElMessage.error('打开目录失败: ' + result.message)
    }
  } catch (error) {
    console.error('打开目录失败:', error)
    ElMessage.error('打开目录失败: ' + error.message)
  }
}

// 采集工具
const handleCollectionTool = async () => {
  qrcodeDialogVisible.value = true
  // 等待DOM更新后生成二维码
  await nextTick()
  generateQRCode()
}

// 生成二维码
const generateQRCode = async () => {
  try {
    if (qrcodeCanvas.value) {
      await QRCode.toCanvas(qrcodeCanvas.value, collectionToolUrl, {
        width: 300,
        margin: 2,
        color: {
          dark: '#000000',
          light: '#ffffff'
        }
      })
    }
  } catch (error) {
    console.error('生成二维码失败:', error)
    ElMessage.error('生成二维码失败: ' + error.message)
  }
}

// 关闭二维码弹窗
const closeQrcodeDialog = () => {
  qrcodeDialogVisible.value = false
}

// 刷新本地机型列表
const handleRefreshLocalModels = async () => {
  try {
    // ElMessage.info('正在刷新本地机型列表...')
    await fetchLocalModels()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('刷新本地机型列表失败:', error)
    ElMessage.error('刷新失败: ' + error.message)
  }
}

// 获取手机型号列表
const fetchPhoneModels = async () => {
  try {
    fetchingModels.value = true
    currentPage.value = 1 // 切换设备或刷新时重置到第一页

    // 使用Wails IPC调用后端GetPhoneTemplates函数获取机型列表
    const response = await GetPhoneTemplates(0, selectedAndroidVersion.value)

    // 解析响应数据
    let apiModels = []
    if (response.code_id == 200 && response.data) {
      const result = response.data
      // 检查result是否直接是数组（后端可能直接返回机型列表）
      if (Array.isArray(result)) {
        apiModels = result // 直接赋值数组
      } else {
        apiModels = result.list || [] // 否则取list字段
      }

      allPhoneModels.value = apiModels // 更新机型列表
      total.value = apiModels.length // 总条数为API返回的数据长度

      // 检查每个机型是否需要显示按钮
      checkNeedShowButtons()

      // 获取本地机型列表
      fetchLocalModels()

      // ElMessage.success(`获取机型数据成功，共 ${total.value} 个机型`)
    } else {
      ElMessage.error('获取机型数据失败: ' + (response.message || '未知错误'))
      // 请求失败时，清空线上机型列表
      allPhoneModels.value = []
      total.value = 0
      showDownloadButton.value = {}

      // 仍然尝试获取本地机型列表
      fetchLocalModels()
    }
  } catch (error) {
    console.error('获取机型数据失败:', error)
    ElMessage.error('获取机型数据失败: ' + error.message)

    // 发生错误时，清空线上机型列表
    allPhoneModels.value = []
    total.value = 0

    // 仍然尝试获取本地机型列表
    fetchLocalModels()
  } finally {
    fetchingModels.value = false
  }
}

// 处理机型操作
const handleModelAction = (model) => {
  console.log('机型操作:', model)
  ElMessage.info(`查看机型 ${model.name} 的详情`)
}

// 存储原始配置，用于保存时恢复被删除的字段
const originalConfig = ref({})

// 处理机型编辑
const handleModelEdit = async (model) => {
  try {
    // 设置编辑中的状态
    editingModel.value = model
    editingModelName.value = model.name // 保存原始机型名称

    // 调用后端方法获取机型配置
    const config = await GetModelConfig(model.name)

    // 保存原始配置，用于后续合并
    originalConfig.value = JSON.parse(JSON.stringify(config))

    // 深拷贝配置，避免直接修改原始数据
    const configCopy = JSON.parse(JSON.stringify(config))

    // 重置正则表达式存储
    regexPatterns.value = {}

    // 存储原始正则表达式并转换为星号显示
    const storeRegexAndConvert = (obj, parentPath = '') => {
      // 使用数组存储要删除的键，避免在遍历过程中修改对象
      const keysToDelete = []

      for (const [key, value] of Object.entries(obj)) {
        // 构建完整路径
        const fullPath = parentPath ? `${parentPath}.${key}` : key

        if (typeof value === 'object' && value !== null) {
          // 递归处理对象
          storeRegexAndConvert(value, fullPath)
        } else if (typeof value === 'string') {
          if (isRegexPattern(value)) {
            // 存储原始正则表达式，使用完整路径作为键
            regexPatterns.value[fullPath] = value
            // 转换为星号显示
            obj[key] = convertRegexToAsterisks(value)
          } else if (isComplexRegexPattern(value)) {
            // 标记复杂正则表达式字段为删除
            keysToDelete.push(key)
          }
        }
      }

      // 删除所有无法处理的复杂正则表达式字段
      for (const key of keysToDelete) {
        delete obj[key]
      }
    }

    // 处理所有字段
    storeRegexAndConvert(configCopy)

    modelConfig.value = configCopy

    console.log('获取机型配置成功:', configCopy)
    console.log('存储的正则表达式:', regexPatterns.value)
    console.log('原始机型名称:', editingModelName.value)
    // 打开编辑弹窗
    configDialogVisible.value = true
    // ElMessage.success(`获取机型 ${model.name} 配置成功`)
  } catch (error) {
    console.error(`获取机型 ${model.name} 配置失败:`, error)
    ElMessage.error(`获取机型 ${model.name} 配置失败: ${error.message}`)
    // 出错时清除编辑中的状态
    editingModel.value = null
  }
}

// 在线设备列表（使用 computed 自动响应状态变化）
const onlineDevices = computed(() => {
  console.log('[机型推送] 计算在线设备列表')
  console.log('[机型推送] 所有设备:', props.devices.length)
  console.log('[机型推送] 设备状态缓存:', props.devicesStatusCache)
  
  const online = props.devices.filter(device => {
    const status = props.devicesStatusCache.get(device.id)
    const isOnline = status === 'online'
    console.log(`[机型推送] 设备 ${device.ip} (${device.id}): 状态=${status}, 在线=${isOnline}`)
    return isOnline
  })
  
  console.log('[机型推送] 在线设备数量:', online.length)
  return online
})

// 处理单个机型推送（点击行内"推送"按钮）
const handleModelPush = async (model) => {
  try {
    pushingModels.value = [model]
    selectedDevices.value = []
    pushDialogVisible.value = true
    await nextTick()
    deviceTable.value?.clearSelection()
  } catch (error) {
    console.error('打开推送弹窗失败:', error)
    ElMessage.error('打开推送弹窗失败: ' + error.message)
  }
}

// 批量推送已勾选机型
const handleBatchPush = async () => {
  if (selectedModels.value.length === 0) {
    ElMessage.warning('请先勾选要推送的机型')
    return
  }
  pushingModels.value = [...selectedModels.value]
  selectedDevices.value = []
  pushDialogVisible.value = true
  await nextTick()
  deviceTable.value?.clearSelection()
}

// 刷新在线设备列表
const refreshOnlineDevices = async () => {
  try {
    refreshingDevices.value = true
    // 触发一次心跳检测更新（可选，如果需要立即刷新）
    // 这里我们只需要等待一下，因为 onlineDevices 是 computed 会自动更新
    await new Promise(resolve => setTimeout(resolve, 500))
    
    ElMessage.success(`已刷新设备列表，当前在线设备: ${onlineDevices.value.length} 台`)
  } catch (error) {
    console.error('刷新设备列表失败:', error)
    ElMessage.error('刷新设备列表失败: ' + error.message)
  } finally {
    refreshingDevices.value = false
  }
}

// 处理设备选择变化
const handleDeviceSelectionChange = (devices) => {
  selectedDevices.value = devices
}

// 确认推送
const confirmPush = async () => {
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请选择要推送的设备')
    return
  }

  try {
    // 推送前再次检查设备在线状态
    const offlineDevices = selectedDevices.value.filter(device => {
      const status = props.devicesStatusCache.get(device.id)
      return status !== 'online' && !device.isOnline
    })

    if (offlineDevices.length > 0) {
      const offlineIPs = offlineDevices.map(d => d.ip).join(', ')
      ElMessage.error(`以下设备已离线，无法推送: ${offlineIPs}`)
      return
    }

    // 初始化进度
    const total = pushingModels.value.length
    isPushing.value = true
    pushProgress.value = { current: 0, total, currentModelName: '', hasError: false }

    // 依次推送每个机型到所有选中设备
    let successCount = 0
    let failMessages = []

    for (const model of pushingModels.value) {
      pushProgress.value.currentModelName = model.name
      try {
        console.log('推送机型:', model.name, '到设备:', selectedDevices.value)
        const result = await PushModelToDevices({
          modelName: model.name,
          devices: selectedDevices.value
        })
        if (result.success) {
          successCount++
        } else {
          failMessages.push(`${model.name}: ${result.message}`)
          pushProgress.value.hasError = true
        }
      } catch (err) {
        failMessages.push(`${model.name}: ${err.message}`)
        pushProgress.value.hasError = true
      }
      pushProgress.value.current++
    }

    pushProgress.value.currentModelName = ''
    isPushing.value = false

    if (failMessages.length === 0) {
      ElMessage.success(
        `成功推送 ${successCount} 个机型到 ${selectedDevices.value.length} 台设备`
      )
      // 记录已推送成功的机型
      for (const model of pushingModels.value) {
        if (!pushedModelNames.value.includes(model.name)) {
          pushedModelNames.value.push(model.name)
        }
      }
      pushDialogVisible.value = false
      // 推送完成后清空已选机型
      selectedModels.value = []
    } else {
      if (successCount > 0) {
        // 部分成功，记录成功推送的机型（排除失败的）
        const failedNames = new Set(failMessages.map(msg => msg.split(':')[0].trim()))
        for (const model of pushingModels.value) {
          if (!failedNames.has(model.name) && !pushedModelNames.value.includes(model.name)) {
            pushedModelNames.value.push(model.name)
          }
        }
        ElMessage.warning(
          `${successCount} 个成功，${failMessages.length} 个失败：${failMessages.join('；')}`
        )
      } else {
        ElMessage.error(`推送失败：${failMessages.join('；')}`)
      }
    }
  } catch (error) {
    isPushing.value = false
    pushProgress.value.currentModelName = ''
    console.error('推送失败:', error)
    ElMessage.error('推送失败: ' + error.message)
  }
}

// 关闭推送弹窗
const closePushDialog = () => {
  if (isPushing.value) return
  pushDialogVisible.value = false
  selectedDevices.value = []
  pushingModels.value = []
  isPushing.value = false
  pushProgress.value = { current: 0, total: 0, currentModelName: '', hasError: false }
  deviceTable.value?.clearSelection()
}

// 关闭配置编辑弹窗
const closeConfigDialog = () => {
  configDialogVisible.value = false
  // 清除编辑中的状态
  editingModel.value = null
}

// 获取数字精度
const getNumberPrecision = (value) => {
  if (Number.isInteger(value)) return 0
  return value.toString().split('.')[1]?.length || 0
}

// 处理复杂类型字段变化
const handleComplexTypeChange = (key, value) => {
  try {
    const parsedValue = JSON.parse(value)
    modelConfig.value[key] = parsedValue
    ElMessage.success(`字段 ${key} 更新成功`)
  } catch (error) {
    console.warn(`解析 ${key} 字段失败:`, error)
    ElMessage.warning(`字段 ${key} JSON格式错误，请检查输入`)
  }
}

// 处理深层对象变化
const handleDeepObjectChange = (key, subKey, value) => {
  try {
    const parsedValue = JSON.parse(value)
    modelConfig.value[key][subKey] = parsedValue
    ElMessage.success(`字段 ${key}.${subKey} 更新成功`)
  } catch (error) {
    console.warn(`解析 ${key}.${subKey} 字段失败:`, error)
    ElMessage.warning(`字段 ${key}.${subKey} JSON格式错误，请检查输入`)
  }
}

// 处理嵌套数组项变化
const handleArrayItemChange = (key, subKey, index, value) => {
  try {
    const parsedValue = JSON.parse(value)
    modelConfig.value[key][subKey][index] = parsedValue
    ElMessage.success(`字段 ${key}.${subKey}[${index}] 更新成功`)
  } catch (error) {
    console.warn(`解析 ${key}.${subKey}[${index}] 字段失败:`, error)
    ElMessage.warning(`字段 ${key}.${subKey}[${index}] JSON格式错误，请检查输入`)
  }
}

// 处理顶层数组项变化
const handleTopArrayItemChange = (key, index, value) => {
  try {
    const parsedValue = JSON.parse(value)
    modelConfig.value[key][index] = parsedValue
    ElMessage.success(`字段 ${key}[${index}] 更新成功`)
  } catch (error) {
    console.warn(`解析 ${key}[${index}] 字段失败:`, error)
    ElMessage.warning(`字段 ${key}[${index}] JSON格式错误，请检查输入`)
  }
}

// 格式化字段标签
const formatFieldLabel = (key) => {
  // 将下划线转为空格，首字母大写
  return key.replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

// 计算属性：overlay字段
const overlayFields = computed(() => {
  if (modelConfig.value.overlay && typeof modelConfig.value.overlay === 'object') {
    return modelConfig.value.overlay
  }
  return {}
})



// 计算属性：其他顶层字段
const otherFields = computed(() => {
  const excludeKeys = ['overlay', 'prop', 'name', 'supported_version', 'template_name']
  const result = {}

  for (const [key, value] of Object.entries(modelConfig.value)) {
    if (!excludeKeys.includes(key) && typeof value !== 'object') {
      result[key] = value
    }
  }

  return result
})

// 计算属性：其他字段数量
const otherFieldsCount = computed(() => {
  return Object.keys(otherFields.value).length
})



// 处理复杂Prop字段变化
const handlePropComplexChange = (key, value) => {
  try {
    const parsedValue = JSON.parse(value)
    modelConfig.value.prop[key] = parsedValue
    ElMessage.success(`Prop字段 ${key} 更新成功`)
  } catch (error) {
    console.warn(`解析Prop字段 ${key} 失败:`, error)
    ElMessage.warning(`Prop字段 ${key} JSON格式错误，请检查输入`)
  }
}

// 保存配置修改
// 深度合并两个对象，target的优先级高于source
const deepMerge = (target, source) => {
  const result = { ...source }

  for (const key in target) {
    if (target.hasOwnProperty(key)) {
      if (typeof target[key] === 'object' && target[key] !== null &&
        typeof source[key] === 'object' && source[key] !== null &&
        !Array.isArray(target[key]) && !Array.isArray(source[key])) {
        // 递归合并对象
        result[key] = deepMerge(target[key], source[key])
      } else {
        // 其他类型直接使用target的值
        result[key] = target[key]
      }
    }
  }

  return result
}

const saveConfigChanges = async () => {
  try {
    // 深拷贝配置，避免直接修改原始数据
    const configToSave = JSON.parse(JSON.stringify(modelConfig.value))

    // 将用户输入转换回正则表达式
    const convertToRegex = (obj, parentPath = '') => {
      for (const [key, value] of Object.entries(obj)) {
        // 构建完整路径
        const fullPath = parentPath ? `${parentPath}.${key}` : key

        if (typeof value === 'object' && value !== null) {
          // 递归处理对象
          convertToRegex(value, fullPath)
        } else if (typeof value === 'string' && regexPatterns.value[fullPath]) {
          // 转换用户输入回正则表达式
          const originalRegex = regexPatterns.value[fullPath]
          obj[key] = convertInputToRegex(value, originalRegex)
        }
      }
    }

    // 处理所有字段
    convertToRegex(configToSave)

    // 深度合并原始配置和修改后的配置，确保所有字段都被提交
    // 使用修改后的配置覆盖原始配置中的对应字段，同时保留原始配置中的其他字段
    const finalConfigToSave = deepMerge(configToSave, originalConfig.value)

    console.log('保存配置修改:', finalConfigToSave)

    // 生成新的机型名称：原始名称 + _ + 三个随机数
    const randomNum = Math.floor(100 + Math.random() * 900); // 生成100-999的随机数
    const newModelName = `${editingModelName.value}_${randomNum}`;

    // 调用后端Go方法保存配置并生成新的zip包，使用新的机型名称
    const result = await SaveModelConfig(newModelName, finalConfigToSave)
    console.log('保存结果:', result)

    if (result.success) {
      ElMessage.success(result.message)
      closeConfigDialog()

      // 重新获取本地机型列表，确保新机型能正确显示
      fetchLocalModels()

      // 重新检查按钮显示状态
      checkNeedShowButtons()
    } else {
      ElMessage.error(result.message)
    }
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error(`保存配置失败: ${error.message}`)
  }
}

// 处理页码变化
const handleCurrentChange = (page) => {
  currentPage.value = page
  // 前端分页，不需要重新请求数据
}

// 处理每页条数变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1 // 重置到第一页
  // 前端分页，不需要重新请求数据
}

// 检查每个机型是否需要显示下载按钮和编辑按钮（批量接口，一次 IPC 完成）
const checkNeedShowButtons = async () => {
  if (allPhoneModels.value.length === 0) return
  try {
    const models = allPhoneModels.value.map(m => ({ id: m.id, name: m.name }))
    const res = await CheckModelButtonStatusBatch(models)
    if (res?.success) {
      const newShowDownloadButton = {}
      const newShowEditButton = {}
      for (const [id, status] of Object.entries(res.result || {})) {
        newShowDownloadButton[id] = status.needDownload
        newShowEditButton[id] = status.hasLocal
      }
      showDownloadButton.value = newShowDownloadButton
      showEditButton.value = newShowEditButton
    }
  } catch (error) {
    console.error('批量检查机型按钮状态失败:', error)
  }
}

// 下载模板文件
const downloadTemplate = async (modelId, modelName) => {
  try {
    // 设置下载中的状态
    downloadingModelId.value = modelId

    // 调用后端方法下载模板
    const result = await DownloadTemplate(modelId.toString())

    console.log('****', result)

    if (result.success) {
      ElMessage.success(`机型 ${modelName} 模板下载成功`)
      // 更新按钮显示状态
      checkNeedShowButtons()
    } else {
      ElMessage.error(`机型 ${modelName} 模板下载失败: ${result.message}`)
    }
  } catch (error) {
    console.error(`下载机型 ${modelName} 模板失败:`, error)
    ElMessage.error(`机型 ${modelName} 模板下载失败: ${error.message}`)
  } finally {
    // 清除下载中的状态
    downloadingModelId.value = ''
  }
}

// 组件初始化时不再自动调用获取机型列表，而是通过父组件调用
// onMounted(() => {
//   fetchPhoneModels()
// })

// 暴露方法给父组件
defineExpose({
  fetchPhoneModels,
  fetchLocalModels,
  checkNeedShowButtons
})
</script>

<style scoped>
.model-management-container {
  /* padding: 20px; */
  height: 100%;
  box-sizing: border-box;
}

.model-content {
  height: 100%;
}

/* .model-content {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
} */

.device-select-card,
.model-list-card {
  margin-bottom: 20px;
}

.model-list-card {
  position: relative;
}

.device-select-content {
  display: flex;
  align-items: center;
  padding: 20px 0;
}

.model-filter-container {
  display: flex;
  margin-bottom: 20px;
  border-bottom: 1px solid #ebeef5;
}

.filter-item {
  padding: 10px 20px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  margin-right: 10px;
  border-bottom: 2px solid transparent;
  transition: all 0.3s ease;
}

.filter-item:hover {
  color: #409eff;
}

.filter-item.active {
  color: #409eff;
  border-bottom-color: #409eff;
  font-weight: 500;
}

.model-list {
  /* padding: 20px 0; */
}

.model-table {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* .table-actions {
  display: flex;
  gap: 5px;
} */

.empty-state {
  padding: 40px 0;
  text-align: center;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 0 4px 4px;
}

/* 分页容器样式 */
.pagination-wrapper {
  display: flex;
  align-items: center;
  gap: 20px;
}

/* 自定义总条数文本样式 */
.total-text {
  font-size: 14px;
  color: #606266;
  line-height: 32px;
  margin-left: 10px;
}

/* 配置表单样式 */
.config-form {
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 10px;
}

/* 对象字段样式 */
.object-field {
  margin-bottom: 20px;
}

.object-field .el-card {
  border: 1px dashed #dcdfe6;
  background-color: #f5f7fa;
}

/* 数组字段样式 */
.array-field {
  margin-bottom: 20px;
}

.array-field .el-card {
  border: 1px dashed #dcdfe6;
  background-color: #f5f7fa;
}

.array-item {
  margin-bottom: 15px;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #ffffff;
}

.array-item-header {
  margin-bottom: 10px;
  font-weight: bold;
  color: #303133;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.array-item-index {
  color: #409eff;
}

/* 对话框内容样式 */
.dialog-content {
  padding: 10px;
}

/* 版本信息样式 */
.version-info {
  margin-bottom: 20px;
  font-weight: bold;
  font-size: 16px;
  color: #303133;
}

/* 配置表单样式 */
.config-form {
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 10px;
}

/* 表单行样式 */
.config-form .el-form-item {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
}

/* 标签样式 */
.config-form .el-form-item__label {
  font-weight: 500;
  color: #303133;
  width: 180px;
  text-align: right;
  padding-right: 20px;
}

/* 输入框样式 */
.config-form .el-input {
  margin-right: 10px;
}

/* 行内表单样式 */
.config-form .el-form--inline .el-form-item {
  width: auto;
}

/* 部分标题样式 */
.section-header {
  margin: 25px 0 15px 0;
  padding: 8px 15px;
  background-color: #ecf5ff;
  color: #409eff;
  font-weight: bold;
  border-radius: 4px;
  font-size: 14px;
  border-left: 4px solid #409eff;
}

/* Prop字段项样式 */
.prop-item {
  margin-bottom: 10px;
  padding: 10px;
  background-color: #fafafa;
  border-radius: 4px;
}

/* JSON字段样式 */
.json-help {
  margin-top: 5px;
  font-size: 12px;
  color: #909399;
}

/* 表单滚动条样式 */
.config-form::-webkit-scrollbar {
  width: 6px;
}

.config-form::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.config-form::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.config-form::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 按钮样式调整 */
.dialog-footer {
  text-align: right;
}

/* 弹窗标题样式 */
.dialog-header {
  display: flex;
  /* flex-direction: column; */
  gap: 5px;
  justify-content: normal;
}

.dialog-title {
  font-size: 16px;
  font-weight: 500;
  color: #000;
}

.dialog-warning {
  font-size: 16px;
  color: #f56c6c;
  font-weight: normal;
}

/* 固定字段区域样式 */
.object-field {
  margin-bottom: 15px;
}

/* 深层对象字段样式 */
.deep-object-field .el-textarea {
  width: 100%;
}

/* 推送弹窗样式 */
.push-dialog-content {
  padding: 10px 0;
}

/* 推送进度条区域 */
.push-progress-area {
  margin-top: 16px;
  padding: 12px 14px;
  background: #f5f7fa;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.push-progress-title {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
}

.push-progress-current {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}

/* 待推送机型展示 */
.push-models-info {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 14px;
  padding: 10px 12px;
  background: #f0f7ff;
  border-radius: 6px;
  border: 1px solid #c6e0ff;
}

.push-models-label {
  font-size: 13px;
  color: #409eff;
  font-weight: 500;
  white-space: nowrap;
  line-height: 22px;
}

.push-models-tags {
  display: flex;
  flex-wrap: wrap;
  flex: 1;
}

.dialog-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.dialog-description {
  margin: 0;
  font-size: 14px;
  color: #606266;
  display: flex;
  align-items: center;
  gap: 8px;
}

.device-count {
  color: #409eff;
  font-weight: 500;
}

.device-table {
  max-height: 400px;
  overflow-y: auto;
}

.device-table .el-table__header-wrapper {
  position: sticky;
  top: 0;
  z-index: 10;
}

/* 二维码弹窗样式 */
.qrcode-dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0;
}

.qrcode-description {
  margin-bottom: 20px;
  font-size: 16px;
  color: #303133;
  font-weight: 500;
}

.qrcode-container {
  margin-bottom: 20px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.qrcode-container canvas {
  display: block;
}

.qrcode-url {
  font-size: 12px;
  color: #909399;
  word-break: break-all;
  text-align: center;
  /* max-width: 350px; */
}

/* 批量操作栏 */
.batch-action-bar {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  min-height: 40px;
}

.selected-count {
  font-size: 13px;
  color: #409eff;
}

/* 已选中机型行高亮 */
.model-table :deep(.selected-model-row) {
  background-color: #ecf5ff !important;
}

.model-table :deep(.selected-model-row:hover > td) {
  background-color: #d9ecff !important;
}

.model-table :deep(.selected-model-row > td) {
  background-color: #ecf5ff !important;
}

/* 已推送机型行高亮 */
.model-table :deep(.pushed-model-row) {
  background-color: #f0f9eb !important;
}

.model-table :deep(.pushed-model-row:hover > td) {
  background-color: #e1f3d8 !important;
}

.model-table :deep(.pushed-model-row > td) {
  background-color: #f0f9eb !important;
}

/* 使用说明 */
.help-content {
  padding: 10px 4px 20px;
  max-height: calc(100vh - 280px);
  overflow-y: auto;
}

.help-section {
  margin-bottom: 24px;
}

.help-section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
  padding: 6px 12px;
  background: #f0f7ff;
  border-left: 4px solid #409eff;
  border-radius: 0 4px 4px 0;
}

.help-section p,
.help-section ul {
  font-size: 13px;
  color: #606266;
  line-height: 1.8;
  margin: 0 0 6px 0;
  padding-left: 16px;
}

.help-section ul {
  padding-left: 28px;
}

.help-section li {
  margin-bottom: 4px;
}

.help-section code {
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 3px;
  padding: 1px 5px;
  font-size: 12px;
  color: #e6a23c;
}
</style>