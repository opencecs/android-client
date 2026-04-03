<template>
  <div class="batch-import">
    <!-- 重要提示 - 顶部可折叠 -->
    <el-collapse v-model="activeNotice" style="margin-bottom: 20px;">
      <el-collapse-item name="notice">
        <template #title>
          <div style="display: flex; align-items: center; gap: 8px;">
            <el-icon style="color: #E6A23C; font-size: 16px;"><Warning /></el-icon>
            <span style="font-weight: bold;">{{ $t('common.importantNotice') }}</span>
          </div>
        </template>
        
        <div style="padding: 10px;">
          <div style="margin-bottom: 10px;">
            <el-tag type="warning" style="margin-right: 10px;">{{ $t('common.supportScope') }}</el-tag>
            <span>{{ $t('common.supportScopeDesc') }}<span style="color: #F56C6C; font-weight: bold;">{{ $t('common.containerNotSupported') }}</span>{{ $t('common.backupImportFunction') }}</span>
          </div>
          <div style="margin-bottom: 10px;">
            <el-tag type="info" style="margin-right: 10px;">{{ $t('common.importPreparation') }}</el-tag>
            <span>{{ $t('common.importPreparationDesc') }}</span>
          </div>
          <div>
            <el-tag type="danger" style="margin-right: 10px;">{{ $t('common.deviceCompatibility') }}</el-tag>
            <span>{{ $t('common.deviceCompatibilityDesc') }}<span style="color: #67C23A; font-weight: bold;">{{ $t('common.cqrReusable') }}</span>, <span style="color: #F56C6C; font-weight: bold;">{{ $t('common.pSeriesNotReusable') }}</span>)</span>
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- 备份文件列表 -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span style="font-weight: bold;">{{ $t('common.backupFileList') }}</span>
          <div style="display: flex; gap: 10px;">
            <el-button type="success" size="small" @click="openBackupFolder">
              <el-icon><FolderOpened /></el-icon>
              {{ $t('common.openImportFolder') }}
            </el-button>
            <el-button type="primary" size="small" @click="refreshBackupFiles" :loading="loading">
              <el-icon><Refresh /></el-icon>
              {{ $t('common.refresh') }}
            </el-button>
          </div>
        </div>
      </template>

      <!-- 有备份文件时显示表格 -->
      <el-table 
        v-if="backupFiles.length > 0" 
        :data="backupFiles" 
        style="width: 100%" 
        v-loading="loading"
      >
        <el-table-column prop="name" :label="$t('common.fileName')" min-width="200" />
        <el-table-column :label="$t('common.fileSize')" width="120" align="center">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.operation')" width="200" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleImport(row)">
              {{ $t('common.import') }}
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 无备份文件时显示空状态 -->
      <div v-if="backupFiles.length === 0 && !loading" style="min-height: 300px;">
        <el-empty description=" ">
          <template #image>
            <el-icon style="font-size: 80px; color: #C0C4CC;">
              <FolderOpened />
            </el-icon>
          </template>
          <template #description>
            <div style="display: flex; flex-direction: column; align-items: center; gap: 20px; padding: 20px;">
              <div style="color: #909399; font-size: 15px; font-weight: 500;">
                {{ $t('common.noBackupFiles') }}
              </div>
              <el-button type="primary" size="large" @click="openBackupFolder">
                <el-icon><FolderOpened /></el-icon>
                {{ $t('common.openImportFolder') }}
              </el-button>
              <div style="color: #C0C4CC; font-size: 13px; max-width: 420px; text-align: center; line-height: 1.8;">
                {{ $t('common.copyToImportFolder') }}<br>{{ $t('common.thenRefresh') }}
              </div>
            </div>
          </template>
        </el-empty>
      </div>

      <!-- 加载中状态 -->
      <div v-if="loading && backupFiles.length === 0" style="min-height: 200px; display: flex; align-items: center; justify-content: center;">
        <el-icon class="is-loading" style="font-size: 40px; color: #409EFF;">
          <Refresh />
        </el-icon>
      </div>
    </el-card>

    <!-- 导入对话框 -->
    <el-dialog
      v-model="importFlowActive"
      :title="$t('common.batchImport')"
      width="70%"
      :close-on-click-modal="false"
      @close="cancelImport"
    >
      <!-- 当前备份文件信息 -->
      <el-alert type="info" :closable="false" style="margin-bottom: 20px;">
        <template #default>
          <div style="display: flex; align-items: center; gap: 10px;">
            <el-icon style="font-size: 20px;"><Document /></el-icon>
            <div>
              <div style="font-weight: bold;">当前备份文件:</div>
              <div style="font-size: 12px; color: #606266; margin-top: 4px;">{{ currentBackupFile?.name }}</div>
            </div>
          </div>
        </template>
      </el-alert>

      <!-- 步骤指示器 -->
      <el-steps :active="currentStep" align-center finish-status="success" style="margin-bottom: 30px;">
        <el-step :title="$t('common.selectDevice')" icon="Monitor" />
        <el-step :title="$t('common.configureSlots')" icon="Grid" />
        <el-step :title="$t('common.executeImport')" icon="Upload" />
      </el-steps>

      <!-- 步骤1: 选择设备 -->
      <div v-show="currentStep === 1">
        <div style="margin-bottom: 15px;">
          <el-input
            v-model="deviceSearchKeyword"
            :placeholder="$t('common.searchDeviceIP')"
            clearable
            prefix-icon="Search"
            style="width: 300px;"
          />
        </div>

        <el-table
          ref="deviceTableRef"
          :data="filteredDevices"
          style="width: 100%"
          @selection-change="handleDeviceSelectionChange"
          max-height="350px"
          border
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="ip" :label="$t('common.deviceIP')" width="150" />
          <el-table-column :label="$t('common.hostFirmwareVersion')" align="center">
            <template #default="{ row }">
              {{ formatSdkVersion(deviceFirmwareInfo.get(row.id)?.sdkVersion) }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.nvmeStorage')" align="center">
            <template #default="{ row }">
              {{ formatStorage(deviceFirmwareInfo.get(row.id)?.originalData?.mmcuse || 0, deviceFirmwareInfo.get(row.id)?.originalData?.mmctotal || 0) }}
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="filteredDevices.length === 0" :description="$t('common.noAvailableDevices')" :image-size="80" />

        <div style="margin-top: 20px; text-align: right;">
          <el-button type="primary" @click="nextStep" :disabled="selectedDevices.length === 0">
            {{ $t('common.nextStep') }}
          </el-button>
        </div>
      </div>

      <!-- 步骤2: 配置坑位 -->
      <div v-show="currentStep === 2">
        <el-tabs v-model="activeDeviceTab" type="card">
          <el-tab-pane
            v-for="device in selectedDevices"
            :key="device.ip"
            :label="`${device.ip}`"
            :name="device.ip"
          >
            <div style="margin-bottom: 15px;">
              <el-button size="small" @click="selectAllSlots(device)">{{ $t('common.selectAll') }}</el-button>
              <el-button size="small" @click="clearAllSlots(device)">{{ $t('common.clear') }}</el-button>
              <el-button size="small" @click="batchSetCopyCount(device)">{{ $t('common.batchSetCopyCount') }}</el-button>
            </div>

            <!-- 坑位选择网格 -->
            <div class="slot-grid" :class="{ 'grid-24': getDeviceSlotCount(device) === 24 }">
              <div
                v-for="slot in getDeviceSlotCount(device)"
                :key="slot"
                class="slot-item"
                :class="{ 'selected': isSlotSelected(device, slot) }"
                @click="toggleSlot(device, slot)"
              >
                <div class="slot-number">{{ $t('common.slotLabel') }} {{ slot }}</div>
                <div v-if="isSlotSelected(device, slot)" class="slot-copy-input">
                  <el-input-number
                    v-model="getSlotConfig(device, slot).copyCount"
                    :min="1"
                    :max="10"
                    size="small"
                    controls-position="right"
                    @click.stop
                  />
                  <span style="font-size: 12px; color: #666;">{{ $t('common.copies') }}</span>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>

        <div style="margin-top: 20px; text-align: right; display: flex; justify-content: space-between;">
          <el-button @click="prevStep">{{ $t('common.previousStep') }}</el-button>
          <el-button type="primary" @click="startImport" :disabled="!hasSelectedSlots()">
            {{ $t('common.startImport') }}
          </el-button>
        </div>
      </div>

      <!-- 步骤3: 执行导入 -->
      <div v-show="currentStep === 3">
        <el-alert
          :title="getImportStatusTitle()"
          :type="importStatus === 'running' ? 'info' : importStatus === 'completed' ? 'success' : 'error'"
          :closable="false"
          style="margin-bottom: 15px;"
        >
          <template v-if="importStatus === 'completed'" #default>
            <div style="font-size: 14px; line-height: 1.8;">
              <div v-if="importProgress.failedTasks === 0">
                <span style="color: #67C23A; font-weight: bold;">🎉 所有云机导入成功！</span>
              </div>
              <div v-else>
                <span style="color: #E6A23C; font-weight: bold;">⚠️ 导入完成，部分任务失败</span>
              </div>
            </div>
          </template>
        </el-alert>

        <!-- 总体进度 -->
        <div style="margin-bottom: 25px;">
          <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
            <span style="font-weight: bold; font-size: 15px;">总体进度</span>
            <div style="display: flex; align-items: center; gap: 8px;">
              <span style="font-size: 18px; font-weight: bold; color: #409EFF;">
                {{ importProgress.completedTasks + importProgress.failedTasks }}
              </span>
              <span style="color: #909399;">/</span>
              <span style="font-size: 16px; color: #606266;">{{ importProgress.totalTasks }}</span>
            </div>
          </div>
          <el-progress
            :percentage="calculatePercentage()"
            :status="getProgressStatus()"
            :stroke-width="20"
            :text-inside="true"
          >
            <template #default="{ percentage }">
              <span style="color: #fff; font-weight: bold;">{{ percentage }}%</span>
            </template>
          </el-progress>
          <div style="margin-top: 12px; display: flex; gap: 20px; font-size: 15px; justify-content: center;">
            <div style="display: flex; align-items: center; gap: 6px;">
              <el-icon style="color: #67C23A; font-size: 18px;"><Check /></el-icon>
              <span style="font-weight: 600;">成功: </span>
              <span style="color: #67C23A; font-weight: bold; font-size: 16px;">{{ importProgress.completedTasks }}</span>
            </div>
            <div style="display: flex; align-items: center; gap: 6px;">
              <el-icon style="color: #F56C6C; font-size: 18px;"><Close /></el-icon>
              <span style="font-weight: 600;">失败: </span>
              <span style="color: #F56C6C; font-weight: bold; font-size: 16px;">{{ importProgress.failedTasks }}</span>
            </div>
          </div>
        </div>

        <!-- 当前任务 -->
        <div v-if="importStatus === 'running' && importProgress.currentDevice" style="margin-bottom: 20px;">
          <div style="display: flex; align-items: center; gap: 10px; padding: 12px; background: #F0F9FF; border: 1px solid #91D5FF; border-radius: 4px;">
            <el-icon class="is-loading" style="color: #409EFF; font-size: 20px;"><Refresh /></el-icon>
            <div>
              <div style="font-weight: bold; color: #303133;">正在导入...</div>
              <div style="font-size: 13px; color: #606266; margin-top: 4px;">
                设备: <span style="font-weight: 600;">{{ importProgress.currentDevice }}</span>
                <span style="margin: 0 8px;">|</span>
                坑位: <span style="font-weight: 600;">{{ importProgress.currentSlot }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 成功列表（仅显示成功的） -->
        <div v-if="importProgress.completedTasks > 0" style="margin-bottom: 20px;">
          <div style="font-weight: bold; font-size: 15px; margin-bottom: 10px; color: #67C23A; display: flex; align-items: center; gap: 6px;">
            <el-icon><Check /></el-icon>
            <span>成功导入的云机 ({{ importProgress.completedTasks }})</span>
          </div>
          <div style="max-height: 200px; overflow-y: auto; background: #F0F9FF; border-radius: 4px; padding: 10px;">
            <div 
              v-for="(detail, index) in successDetails"
              :key="index"
              style="padding: 8px 10px; background: white; border-radius: 4px; margin-bottom: 8px; display: flex; justify-content: space-between; align-items: center; border-left: 3px solid #67C23A;"
            >
              <div style="flex: 1;">
                <div style="display: flex; align-items: center; gap: 10px;">
                  <el-tag type="success" size="small">✓</el-tag>
                  <span style="font-weight: 600; color: #303133;">{{ detail.machine_name }}</span>
                </div>
                <div style="font-size: 12px; color: #909399; margin-top: 4px; padding-left: 32px;">
                  设备: {{ detail.device_ip }} | 坑位: {{ detail.slot_number }}
                </div>
              </div>
              <span style="color: #67C23A; font-size: 12px; white-space: nowrap; margin-left: 10px;">{{ formatDuration(detail.duration) }}</span>
            </div>
          </div>
        </div>

        <!-- 失败列表（仅显示失败的） -->
        <div v-if="importProgress.failedTasks > 0">
          <div style="font-weight: bold; font-size: 15px; margin-bottom: 10px; color: #F56C6C; display: flex; align-items: center; gap: 6px;">
            <el-icon><Close /></el-icon>
            <span>导入失败 ({{ importProgress.failedTasks }})</span>
          </div>
          <div style="max-height: 200px; overflow-y: auto; background: #FEF0F0; border-radius: 4px; padding: 10px;">
            <div 
              v-for="(detail, index) in failedDetails"
              :key="index"
              style="padding: 8px 10px; background: white; border-radius: 4px; margin-bottom: 8px; border-left: 3px solid #F56C6C;"
            >
              <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 6px;">
                <el-tag type="danger" size="small">✗</el-tag>
                <span style="font-weight: 600; color: #303133;">{{ detail.machine_name || '未知' }}</span>
              </div>
              <div style="font-size: 12px; color: #909399; padding-left: 32px;">
                设备: {{ detail.device_ip }} | 坑位: {{ detail.slot_number }}
              </div>
              <div style="color: #F56C6C; font-size: 12px; margin-top: 6px; padding-left: 32px; background: #FEF0F0; padding: 6px; border-radius: 3px;">
                {{ detail.message }}
              </div>
            </div>
          </div>
        </div>

        <div v-if="importStatus !== 'running'" style="margin-top: 20px; text-align: right;">
          <el-button type="primary" size="large" @click="finishImport">
            <el-icon><Check /></el-icon>
            {{ $t('common.complete') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Check, Close, Warning, InfoFilled, FolderOpened, Refresh, Upload, Delete } from '@element-plus/icons-vue'
import { ListBackupFiles, DeleteBackupFile, StartBatchImport, OpenBackupMachineDir, GetBatchImportProgress } from '../../bindings/edgeclient/app'
import { Events } from '@wailsio/runtime'

// Props
const props = defineProps({
  devices: {
    type: Array,
    default: () => []
  },
  deviceFirmwareInfo: {
    type: Map,
    default: () => new Map()
  },
  devicesStatusCache: {
    type: Map,
    default: () => new Map()
  }
})

// State
const loading = ref(false)
const backupFiles = ref([])
const importFlowActive = ref(false)
const currentStep = ref(1)
const currentBackupFile = ref(null)
const activeNotice = ref([]) // 折叠面板默认收起

// 设备选择表格 ref
const deviceTableRef = ref(null)

// 设备选择
const deviceSearchKeyword = ref('')
const selectedDevices = ref([])
const activeDeviceTab = ref('')

// 坑位配置
const slotConfigs = ref(new Map())

// 导入状态
const importStatus = ref('')
const importProgress = ref({
  totalTasks: 0,
  completedTasks: 0,
  failedTasks: 0,
  currentDevice: '',
  currentSlot: 0,
  details: []
})

// 轮询定时器
let pollTimer = null

// 获取设备的坑位数量
const getDeviceSlotCount = (device) => {
  if (device.name && device.name.toLowerCase().includes('p1')) {
    return 24
  }
  return 12
}

// 格式化固件版本
const formatSdkVersion = (version) => {
  if (!version) return '未知'
  return version
}

// 格式化存储空间
const formatStorage = (used, total) => {
  if (!total || total === 0) return '0/0 MB'
  if (total >= 1000) {
    const usedGB = (used / 1024).toFixed(1)
    const totalGB = (total / 1024).toFixed(1)
    return `${usedGB} GB/${totalGB} GB`
  }
  return `${used || 0} MB/${total || 0} MB`
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

// 格式化耗时（毫秒 -> 秒/分钟）
const formatDuration = (ms) => {
  if (!ms || ms === 0) return '0秒'
  
  const seconds = ms / 1000
  
  // 小于60秒，显示秒
  if (seconds < 60) {
    return seconds.toFixed(1) + '秒'
  }
  
  // 大于等于60秒，显示分钟
  const minutes = seconds / 60
  return minutes.toFixed(1) + '分钟'
}

// 过滤设备列表（只显示在线设备）
const filteredDevices = computed(() => {
  // 先过滤在线设备
  const onlineDevices = props.devices.filter(device => {
    const status = props.devicesStatusCache.get(device.id)
    return status === 'online'
  })
  
  // 如果有搜索关键词，进一步过滤
  if (!deviceSearchKeyword.value) {
    return onlineDevices
  }
  
  const keyword = deviceSearchKeyword.value.toLowerCase()
  return onlineDevices.filter(d => d.ip.toLowerCase().includes(keyword))
})

// 加载备份文件列表
const refreshBackupFiles = async () => {
  loading.value = true
  try {
    const result = await ListBackupFiles()
    console.log('📁 ListBackupFiles 返回结果:', result)
    
    if (result.success) {
      backupFiles.value = result.files || []
      console.log('✅ 备份文件数量:', backupFiles.value.length)
      console.log('📋 文件列表:', backupFiles.value)
    } else {
      console.error('❌ 加载失败:', result.message)
      ElMessage.error(result.message || '加载备份文件失败')
    }
  } catch (error) {
    console.error('❌ 加载备份文件异常:', error)
    ElMessage.error('加载备份文件失败')
  } finally {
    loading.value = false
  }
}

// 打开导入文件夹
const openBackupFolder = async () => {
  try {
    const result = await OpenBackupMachineDir()
    console.log('📂 OpenBackupMachineDir 返回结果:', result)
    
    if (result.success) {
      ElMessage.success('已打开导入文件夹')
    } else {
      ElMessage.error(result.message || '打开文件夹失败')
    }
  } catch (error) {
    console.error('❌ 打开文件夹异常:', error)
    ElMessage.error('打开文件夹失败')
  }
}

// 删除备份文件
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除备份文件 "${row.name}" 吗？`, '提示', {
      type: 'warning'
    })

    const result = await DeleteBackupFile(row.name)
    if (result.success) {
      ElMessage.success('删除成功')
      refreshBackupFiles()
    } else {
      ElMessage.error(result.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除备份文件失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 开始导入流程
const handleImport = (row) => {
  currentBackupFile.value = row
  currentStep.value = 1
  selectedDevices.value = []
  slotConfigs.value = new Map()
  importFlowActive.value = true
  // 弹窗打开后清除表格的视觉选中状态
  nextTick(() => {
    deviceTableRef.value?.clearSelection()
  })
}

// 取消导入
const cancelImport = () => {
  stopPolling() // 停止轮询
  importFlowActive.value = false
  currentStep.value = 1
  selectedDevices.value = []
  slotConfigs.value = new Map()
  deviceTableRef.value?.clearSelection()
}

// 完成导入
const finishImport = () => {
  stopPolling() // 停止轮询
  importFlowActive.value = false
  currentStep.value = 1
  selectedDevices.value = []
  slotConfigs.value = new Map()
  importProgress.value = {
    totalTasks: 0,
    completedTasks: 0,
    failedTasks: 0,
    currentDevice: '',
    currentSlot: 0,
    details: []
  }
  importStatus.value = ''
}

// 设备选择变化
const handleDeviceSelectionChange = (selection) => {
  selectedDevices.value = selection
  if (selection.length > 0 && !activeDeviceTab.value) {
    activeDeviceTab.value = selection[0].ip
  }
}

// 下一步
const nextStep = () => {
  if (currentStep.value === 1 && selectedDevices.value.length === 0) {
    ElMessage.warning('请至少选择一个设备')
    return
  }
  
  if (currentStep.value === 1) {
    if (selectedDevices.value.length > 0) {
      activeDeviceTab.value = selectedDevices.value[0].ip
    }
  }
  
  currentStep.value++
}

// 上一步
const prevStep = () => {
  currentStep.value--
}

// 切换坑位选中状态
const toggleSlot = (device, slotNumber) => {
  const key = `${device.ip}_${slotNumber}`
  if (!slotConfigs.value.has(key)) {
    slotConfigs.value.set(key, {
      deviceIP: device.ip,
      slotNumber: slotNumber,
      copyCount: 1
    })
  } else {
    slotConfigs.value.delete(key)
  }
}

// 判断坑位是否被选中
const isSlotSelected = (device, slotNumber) => {
  return slotConfigs.value.has(`${device.ip}_${slotNumber}`)
}

// 获取坑位配置
const getSlotConfig = (device, slotNumber) => {
  const key = `${device.ip}_${slotNumber}`
  if (!slotConfigs.value.has(key)) {
    const config = {
      deviceIP: device.ip,
      slotNumber: slotNumber,
      copyCount: 1
    }
    slotConfigs.value.set(key, config)
  }
  return slotConfigs.value.get(key)
}

// 全选坑位
const selectAllSlots = (device) => {
  const slotCount = getDeviceSlotCount(device)
  for (let i = 1; i <= slotCount; i++) {
    const key = `${device.ip}_${i}`
    if (!slotConfigs.value.has(key)) {
      slotConfigs.value.set(key, {
        deviceIP: device.ip,
        slotNumber: i,
        copyCount: 1
      })
    }
  }
}

// 清空坑位选择
const clearAllSlots = (device) => {
  const slotCount = getDeviceSlotCount(device)
  for (let i = 1; i <= slotCount; i++) {
    slotConfigs.value.delete(`${device.ip}_${i}`)
  }
}

// 批量设置复制份数
const batchSetCopyCount = async (device) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入复制份数 (1-10)', '批量设置', {
      inputPattern: /^([1-9]|10)$/,
      inputErrorMessage: '请输入1-10之间的数字'
    })

    const count = parseInt(value)
    slotConfigs.value.forEach((config, key) => {
      if (config.deviceIP === device.ip) {
        config.copyCount = count
      }
    })

    ElMessage.success('设置成功')
  } catch (error) {
    // 用户取消
  }
}

// 判断是否有选中的坑位
const hasSelectedSlots = () => {
  return slotConfigs.value.size > 0
}

// 计算进度百分比
const calculatePercentage = () => {
  if (importProgress.value.totalTasks === 0) return 0
  const completed = importProgress.value.completedTasks + importProgress.value.failedTasks
  return Math.floor((completed / importProgress.value.totalTasks) * 100)
}

// 获取进度条状态
const getProgressStatus = () => {
  if (importStatus.value !== 'completed') return undefined
  if (importProgress.value.failedTasks === 0) return 'success'
  if (importProgress.value.completedTasks === 0) return 'exception'
  return 'warning'
}

// 获取导入状态标题
const getImportStatusTitle = () => {
  if (importStatus.value === 'running') {
    return '正在导入，请稍候...'
  } else if (importStatus.value === 'completed') {
    if (importProgress.value.failedTasks === 0) {
      return `导入完成！成功导入 ${importProgress.value.completedTasks} 个云机`
    } else if (importProgress.value.completedTasks === 0) {
      return '导入失败'
    } else {
      return `导入完成！成功 ${importProgress.value.completedTasks} 个，失败 ${importProgress.value.failedTasks} 个`
    }
  }
  return '导入失败'
}

// 成功的详情列表
const successDetails = computed(() => {
  return importProgress.value.details.filter(d => d.success)
})

// 失败的详情列表
const failedDetails = computed(() => {
  return importProgress.value.details.filter(d => !d.success)
})

// 轮询进度
const pollProgress = async () => {
  try {
    console.log('🔄 [Poll] 轮询进度...')
    const result = await GetBatchImportProgress()
    
    console.log('📦 [Poll] 轮询返回:', result)
    
    if (result.success && result.progress) {
      const progress = result.progress
      
      console.log('✅ [Poll] 进度数据:')
      console.log('  - total_tasks:', progress.total_tasks)
      console.log('  - completed_tasks:', progress.completed_tasks)
      console.log('  - failed_tasks:', progress.failed_tasks)
      console.log('  - details长度:', progress.details ? progress.details.length : 0)
      
      // 更新进度
      importProgress.value = {
        totalTasks: progress.total_tasks || 0,
        completedTasks: progress.completed_tasks || 0,
        failedTasks: progress.failed_tasks || 0,
        currentDevice: progress.current_device || '',
        currentSlot: progress.current_slot || 0,
        details: progress.details || []
      }
      
      const percentage = importProgress.value.totalTasks > 0 
        ? Math.floor(((importProgress.value.completedTasks + importProgress.value.failedTasks) / importProgress.value.totalTasks) * 100)
        : 0
      
      console.log('✅ [Poll] 进度已更新: ' + percentage + '%')
      
      // 如果任务完成，停止轮询
      if (result.status === 'completed' || result.status === 'failed') {
        console.log('🎉 [Poll] 任务已完成，停止轮询，状态:', result.status)
        stopPolling()
        importStatus.value = result.status
        
        // 显示完成消息
        const failed = importProgress.value.failedTasks
        const completed = importProgress.value.completedTasks
        
        if (result.status === 'completed') {
          if (failed === 0) {
            ElMessage.success(`所有导入任务已完成！成功: ${completed}`)
          } else {
            ElMessage.warning(`导入完成，成功: ${completed}, 失败: ${failed}`)
          }
        } else {
          ElMessage.error('导入任务失败')
        }
      }
    } else {
      console.warn('⚠️ [Poll] 轮询返回数据无效:', result)
    }
  } catch (error) {
    console.error('❌ [Poll] 轮询异常:', error)
  }
}

// 启动轮询
const startPolling = () => {
  console.log('🎬 [Poll] 启动轮询，间隔: 500ms')
  stopPolling() // 先清除旧的定时器
  pollTimer = setInterval(pollProgress, 500)
}

// 停止轮询
const stopPolling = () => {
  if (pollTimer) {
    console.log('🛑 [Poll] 停止轮询')
    clearInterval(pollTimer)
    pollTimer = null
  }
}

// 开始导入
const startImport = async () => {
  const devicesConfig = []
  let totalTasksCount = 0 // 手动计算总任务数
  
  selectedDevices.value.forEach(device => {
    const deviceSlots = []
    
    slotConfigs.value.forEach((config, key) => {
      if (config.deviceIP === device.ip) {
        deviceSlots.push({
          slot_number: config.slotNumber,
          copy_count: config.copyCount,
          machine_name: ''
        })
        totalTasksCount += config.copyCount // 累加每个坑位的复制份数
      }
    })
    
    if (deviceSlots.length > 0) {
      const slotCount = getDeviceSlotCount(device)
      devicesConfig.push({
        device_ip: device.ip,
        device_type: slotCount === 12 ? '12slots' : '24slots',
        slot_configs: deviceSlots
      })
    }
  })
  
  if (devicesConfig.length === 0) {
    ElMessage.warning('请至少选择一个坑位')
    return
  }
  
  console.log('📊 [StartImport] 手动计算的总任务数:', totalTasksCount)
  console.log('📋 [StartImport] 配置详情:', devicesConfig)
  
  // 立即初始化进度（使用手动计算的值）
  importProgress.value = {
    totalTasks: totalTasksCount,
    completedTasks: 0,
    failedTasks: 0,
    currentDevice: '',
    currentSlot: 0,
    details: []
  }
  
  console.log('✅ [StartImport] 进度初始化完成:', importProgress.value)
  
  currentStep.value = 3
  importStatus.value = 'running'
  
  console.log('🎬 [StartImport] 当前步骤:', currentStep.value)
  console.log('🎬 [StartImport] 导入状态:', importStatus.value)
  
  try {
    console.log('🚀 [StartImport] 调用 StartBatchImport API...')
    const result = await StartBatchImport(currentBackupFile.value.name, devicesConfig)
    
    console.log('📦 [StartImport] StartBatchImport 返回结果:', result)
    
    if (!result.success) {
      console.error('❌ [StartImport] 启动失败:', result.message)
      ElMessage.error(result.message || '启动导入任务失败')
      importStatus.value = 'failed'
      return
    }
    
    console.log('✅ [StartImport] 任务启动成功，等待后端事件...')
    
    // 后端返回的任务信息（用于验证）
    if (result.task && result.task.progress) {
      const taskProgress = result.task.progress
      console.log('✅ [StartImport] 后端返回的总任务数:', taskProgress.total_tasks || 0)
      
      // 如果后端返回的值与前端计算的不一致，使用后端的值
      if (taskProgress.total_tasks && taskProgress.total_tasks !== totalTasksCount) {
        console.warn('⚠️ [StartImport] 任务数不匹配! 前端计算:', totalTasksCount, '后端返回:', taskProgress.total_tasks)
        importProgress.value.totalTasks = taskProgress.total_tasks
      }
    }
    
    console.log('✅ [StartImport] 初始化完成，总任务数:', importProgress.value.totalTasks)
    console.log('👂 [StartImport] 现在等待后端发送事件: batch-import:progress 和 batch-import:complete')
    
    // 启动轮询（作为事件系统的备选方案）
    console.log('🔄 [StartImport] 启动轮询机制作为备选方案')
    startPolling()
    
  } catch (error) {
    console.error('❌ [StartImport] 启动导入任务异常:', error)
    ElMessage.error('启动导入任务失败')
    importStatus.value = 'failed'
    stopPolling() // 失败时停止轮询
  }
}

// 暴露方法给父组件
defineExpose({
  refreshBackupFiles
})

// 初始化
onMounted(() => {
  refreshBackupFiles()
  
  // 延迟注册事件监听器，确保Events模块已加载
  const initEventListeners = () => {
    if (!Events || !Events.On) {
      console.warn('⚠️ [BatchImport] Events 未就绪，500ms后重试...')
      setTimeout(initEventListeners, 500)
      return
    }
    
    console.log('🎯 [BatchImport] Events 已就绪，开始注册事件监听器')
    console.log('🔍 [BatchImport] Events 对象:', Events)
    console.log('🔍 [BatchImport] Events.On 类型:', typeof Events.On)
    
    // 监听进度更新事件
    const progressHandler = (data) => {
      console.log('📊 [EVENT] ========== 收到进度事件 ==========')
      console.log('📊 [EVENT] 事件名: batch-import:progress')
      console.log('📦 [EVENT] 数据类型:', typeof data)
      console.log('📦 [EVENT] 原始数据:', data)
      console.log('📦 [EVENT] JSON格式:', JSON.stringify(data, null, 2))
      
      if (data && data.progress) {
        const progress = data.progress
        
        console.log('✅ [EVENT] 进度数据解析成功')
        console.log('  - total_tasks:', progress.total_tasks)
        console.log('  - completed_tasks:', progress.completed_tasks)
        console.log('  - failed_tasks:', progress.failed_tasks)
        console.log('  - details长度:', progress.details ? progress.details.length : 0)
        
        // 使用 snake_case 字段名
        importProgress.value = {
          totalTasks: progress.total_tasks || 0,
          completedTasks: progress.completed_tasks || 0,
          failedTasks: progress.failed_tasks || 0,
          currentDevice: progress.current_device || '',
          currentSlot: progress.current_slot || 0,
          details: progress.details || []
        }
        
        const percentage = importProgress.value.totalTasks > 0 
          ? Math.floor(((importProgress.value.completedTasks + importProgress.value.failedTasks) / importProgress.value.totalTasks) * 100)
          : 0
        
        console.log('✅ [EVENT] 进度已更新到 Vue state:')
        console.log('  - totalTasks:', importProgress.value.totalTasks)
        console.log('  - completedTasks:', importProgress.value.completedTasks)
        console.log('  - failedTasks:', importProgress.value.failedTasks)
        console.log('  - percentage:', percentage + '%')
        console.log('  - detailsCount:', importProgress.value.details.length)
        console.log('📊 [EVENT] ========== 进度事件处理完成 ==========')
      } else {
        console.error('⚠️ [EVENT] 进度事件数据结构异常!')
        console.error('  - data存在:', !!data)
        console.error('  - data.progress存在:', data && !!data.progress)
        console.error('  - 完整data:', data)
      }
    }
    
    Events.On('batch-import:progress', progressHandler)
    console.log('✅ [BatchImport] batch-import:progress 监听器已注册')
    
    // 监听完成事件
    const completeHandler = (data) => {
      console.log('🎉 [EVENT] ========== 收到完成事件 ==========')
      console.log('🎉 [EVENT] 事件名: batch-import:complete')
      console.log('📦 [EVENT] 数据类型:', typeof data)
      console.log('📦 [EVENT] 原始数据:', data)
      console.log('📦 [EVENT] JSON格式:', JSON.stringify(data, null, 2))
      
      importStatus.value = data.status || 'completed'
      console.log('✅ [EVENT] importStatus 已更新为:', importStatus.value)
      
      if (data && data.progress) {
        const progress = data.progress
        
        console.log('✅ [EVENT] 完成事件进度数据解析成功')
        console.log('  - total_tasks:', progress.total_tasks)
        console.log('  - completed_tasks:', progress.completed_tasks)
        console.log('  - failed_tasks:', progress.failed_tasks)
        
        // 使用 snake_case 字段名
        importProgress.value = {
          totalTasks: progress.total_tasks || 0,
          completedTasks: progress.completed_tasks || 0,
          failedTasks: progress.failed_tasks || 0,
          currentDevice: progress.current_device || '',
          currentSlot: progress.current_slot || 0,
          details: progress.details || []
        }
        
        console.log('✅ [EVENT] 最终进度已更新到 Vue state:')
        console.log('  - totalTasks:', importProgress.value.totalTasks)
        console.log('  - completedTasks:', importProgress.value.completedTasks)
        console.log('  - failedTasks:', importProgress.value.failedTasks)
        console.log('  - details数量:', importProgress.value.details.length)
      } else {
        console.error('⚠️ [EVENT] 完成事件数据结构异常!')
        console.error('  - data存在:', !!data)
        console.error('  - data.progress存在:', data && !!data.progress)
      }
      
      if (data.status === 'completed') {
        const failed = importProgress.value.failedTasks
        const completed = importProgress.value.completedTasks
        
        console.log('🎊 [EVENT] 显示完成消息: 成功=' + completed + ', 失败=' + failed)
        
        if (failed === 0) {
          ElMessage.success(`所有导入任务已完成！成功: ${completed}`)
        } else {
          ElMessage.warning(`导入完成，成功: ${completed}, 失败: ${failed}`)
        }
      } else if (data.status === 'failed') {
        console.error('❌ [EVENT] 任务状态为失败')
        ElMessage.error('导入任务失败')
      }
      
      console.log('🎉 [EVENT] ========== 完成事件处理完成 ==========')
    }
    
    Events.On('batch-import:complete', completeHandler)
    console.log('✅ [BatchImport] batch-import:complete 监听器已注册')
    
    console.log('✅ [BatchImport] 所有事件监听器注册完成')
    console.log('📝 [BatchImport] 监听的事件列表:')
    console.log('   1. batch-import:progress')
    console.log('   2. batch-import:complete')
  }
  
  // 延迟注册，确保Events模块已加载
  setTimeout(initEventListeners, 100)
})
</script>

<style scoped>
.batch-import {
  padding: 20px;
  box-sizing: border-box;
}

.slot-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 10px;
  margin-top: 10px;
}

.slot-grid.grid-24 {
  grid-template-columns: repeat(8, 1fr);
}

.slot-item {
  border: 2px solid #DCDFE6;
  border-radius: 4px;
  padding: 15px 10px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  background: #fff;
}

.slot-item:hover {
  border-color: #409EFF;
  background: #ECF5FF;
}

.slot-item.selected {
  border-color: #409EFF;
  background: #ECF5FF;
}

.slot-number {
  font-weight: bold;
  margin-bottom: 10px;
  color: #303133;
}

.slot-copy-input {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
}
</style>
