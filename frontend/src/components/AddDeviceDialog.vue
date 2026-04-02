<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('addDevice.title')"
    width="1050px"
    height="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-tabs v-model="activeTab">
      <!-- 扫描发现Tab -->
      <el-tab-pane :label="t('addDevice.scanDiscovery')" name="scan">
        <div class="scan-tab-content">
          <div class="scan-header" style="margin-bottom: 16px;">
            <el-button 
              type="success" 
              @click="handleScanDevices"
              :loading="isScanning"
              :disabled="isScanning"
            >
              <el-icon v-if="!isScanning"><Search /></el-icon>
              <span>{{ isScanning ? t('addDevice.scanning') : t('addDevice.scanDevices') }}</span>
            </el-button>
            <span class="scan-hint" style="margin-left: 12px; color: #909399; font-size: 13px;">
              {{ t('addDevice.autoDiscoverHint') }}
            </span>
          </div>
          
          <!-- 扫描结果表格 -->
          <div v-if="scannedDevices.length > 0" class="scan-result-section">
            <div class="search-box" style="margin-bottom: 12px; display: flex; align-items: center; gap: 12px;">
              <el-input
                v-model="searchKeyword"
                :placeholder="t('addDevice.searchPlaceholder')"
                :prefix-icon="Search"
                clearable
                style="width: 280px;"
              />
              <el-checkbox 
                v-model="isAllSelected" 
                @change="toggleSelectAll"
                style="margin-left: 8px;"
              >
                {{ t('addDevice.selectAllToggle') }}
              </el-checkbox>
              <span style="margin-left: auto; color: #909399; font-size: 13px;">
                {{ t('addDevice.devicesCount', { total: scannedDevices.length, filtered: filteredScannedDevices.length, selected: selectedDeviceIds.size }) }}
              </span>
            </div>
            
            <el-table 
              ref="scanTableRef"
              :data="filteredScannedDevices" 
              stripe 
              style="width: 100%;"
              max-height="400"
              @selection-change="handleSelectionChange"
            >
              <el-table-column type="selection" width="50" :selectable="(row) => !isDeviceAdded(row.id)" />
              <el-table-column prop="ip" :label="t('addDevice.deviceIP')" width="150"></el-table-column>
              <el-table-column prop="name" :label="t('addDevice.deviceName')" min-width="180"></el-table-column>
              <el-table-column prop="id" :label="t('addDevice.deviceID')" min-width="200"></el-table-column>
              <el-table-column prop="version" :label="t('addDevice.version')" width="100">
                <template #default="scope">
                  <el-tag size="small" :type="scope.row.version === 'v3' ? 'success' : 'warning'">
                    {{ scope.row.version }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('addDevice.status')" width="100">
                <template #default="scope">
                  <el-tag v-if="isDeviceAdded(scope.row.id)" type="success" size="small">{{ t('addDevice.added') }}</el-tag>
                  <el-tag v-else type="info" size="small">{{ t('addDevice.canAdd') }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            
            <div class="batch-actions" style="margin-top: 16px; text-align: right;">
              <el-button 
                type="primary" 
                size="large"
                :disabled="selectedDevices.length === 0"
                @click="handleBatchAdd"
              >
                {{ t('addDevice.addSelectedDevices') }} ({{ selectedDevices.length }})
              </el-button>
            </div>
          </div>
          
          <!-- 无扫描结果 -->
          <div v-else-if="!isScanning" class="no-devices">
            <el-empty :description="t('addDevice.noDevicesFound')" :image-size="80"></el-empty>
          </div>
        </div>
      </el-tab-pane>
      
      <!-- 手动添加Tab -->
      <el-tab-pane :label="t('addDevice.manualAdd')" name="manual">
        <div class="manual-tab-content">
          <div class="manual-input-section">
            <div class="input-label">{{ t('addDevice.enterDeviceIP') }}</div>
            <div class="input-hint" style="margin-bottom: 12px; color: #909399; font-size: 13px;">
              {{ t('addDevice.multipleIPHint') }}
            </div>
            <el-input
              v-model="manualInput"
              type="textarea"
              :rows="6"
              :placeholder="t('addDevice.enterIPPlaceholder')"
              resize="vertical"
            />
          </div>
          
          <div class="manual-actions" style="margin-top: 16px;">
            <el-button 
              type="primary" 
              @click="handleManualAdd"
              :loading="addingDevices"
              :disabled="!manualInput.trim() || addingDevices"
            >
              <el-icon v-if="!addingDevices"><Plus /></el-icon>
              <span>{{ addingDevices ? t('addDevice.querying') : t('addDevice.addDevice') }}</span>
            </el-button>
            <el-button @click="manualInput = ''" :disabled="addingDevices">{{ t('addDevice.clear') }}</el-button>
          </div>
          
          <!-- 添加结果 -->
          <div v-if="manualAddResult" class="manual-result" style="margin-top: 20px;">
            <div v-if="manualAddResult.success && manualAddResult.devices.length > 0" class="success-result">
              <el-alert
                :title="t('addDevice.successAdded', { count: manualAddResult.addedCount })"
                type="success"
                :closable="false"
                show-icon
              />
              <div v-if="manualAddResult.failedIPs && manualAddResult.failedIPs.length > 0" class="failed-list" style="margin-top: 12px;">
                <div style="color: #E6A23C; font-weight: 500; margin-bottom: 8px;">
                  {{ t('addDevice.noResponseDevices', { count: manualAddResult.failedIPs.length }) }}
                </div>
                <div style="color: #909399; font-size: 13px; word-break: break-all;">
                  {{ manualAddResult.failedIPs.join(', ') }}
                </div>
              </div>
            </div>
            <div v-else class="error-result">
              <el-alert
                :title="manualAddResult.error || t('addDevice.noDeviceDiscovered')"
                type="error"
                :closable="false"
                show-icon
              />
              <div v-if="manualAddResult.failedIPs && manualAddResult.failedIPs.length > 0" class="failed-list" style="margin-top: 12px;">
                <div style="color: #909399; font-size: 13px; word-break: break-all;">
                  {{ t('addDevice.noResponseLabel') }}{{ manualAddResult.failedIPs.join(', ') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ t('addDevice.close') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, getCurrentInstance } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import { discoverOnlineDevicesOnly } from '../services/api.js'
import { handleManualDeviceDiscovery } from '../services/cloudMachineFunctions.js'

// 国际化支持
const { proxy } = getCurrentInstance()
const t = (key, params) => {
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
  existingDeviceIds: {
    type: Set,
    default: () => new Set()
  }
})

const emit = defineEmits(['update:visible', 'device-added', 'devices-added', 'scan-start', 'scan-complete'])

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})
const activeTab = ref('scan')
const isScanning = ref(false)  // AddDeviceDialog 内部使用的扫描状态
const addingDevices = ref(false)
const scannedDevices = ref([])
const manualInput = ref('')
const manualAddResult = ref(null)
const searchKeyword = ref('')
const selectedDevices = ref([])
const scanTableRef = ref(null)  // 表格引用

watch(() => props.visible, (val) => {
  if (val) {
    // 打开对话框时重置状态
    activeTab.value = 'scan'
    manualInput.value = ''
    manualAddResult.value = null
    // 自动开始扫描
    handleScanDevices()
  }
})

const handleClose = () => {
  // 关闭对话框
  dialogVisible.value = false
  // 关闭时通知父组件恢复按钮状态
  emit('scan-complete')
}

const isDeviceAdded = (deviceId) => {
  return props.existingDeviceIds.has(deviceId)
}

const filteredScannedDevices = computed(() => {
  let devices = scannedDevices.value
  
  // 搜索过滤
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    devices = devices.filter(device => 
      device.ip.toLowerCase().includes(keyword) ||
      device.name.toLowerCase().includes(keyword) ||
      device.id.toLowerCase().includes(keyword)
    )
  }
  
  // IP排序
  return devices.sort((a, b) => {
    const ipA = a.ip.split('.').map(num => parseInt(num))
    const ipB = b.ip.split('.').map(num => parseInt(num))
    for (let i = 0; i < 4; i++) {
      if (ipA[i] !== ipB[i]) {
        return ipA[i] - ipB[i]
      }
    }
    return 0
  })
})

const selectedDeviceIds = computed(() => {
  return new Set(selectedDevices.value.map(d => d.id))
})

const isAllSelected = computed(() => {
  if (filteredScannedDevices.value.length === 0) return false
  return filteredScannedDevices.value.every(device => 
    selectedDeviceIds.value.has(device.id) || isDeviceAdded(device.id)
  )
})

const handleSelectionChange = (selection) => {
  selectedDevices.value = selection
}

const toggleSelectAll = () => {
  const table = scanTableRef.value
  if (!table) return
  
  const visibleDevices = filteredScannedDevices.value
  const selectableDevices = visibleDevices.filter(device => !isDeviceAdded(device.id))
  
  // 判断当前是否全选
  const allSelected = selectableDevices.length > 0 && 
    selectableDevices.every(device => selectedDeviceIds.value.has(device.id))
  
  if (allSelected) {
    // 反选：清空选择
    table.clearSelection()
  } else {
    // 全选：选择所有可选的设备
    table.clearSelection()
    visibleDevices.forEach(device => {
      if (!isDeviceAdded(device.id)) {
        table.toggleRowSelection(device, true)
      }
    })
  }
}

const handleBatchAdd = () => {
  if (selectedDevices.value.length === 0) return
  
  emit('devices-added', selectedDevices.value)
  
  ElMessage.success(`已添加 ${selectedDevices.value.length} 个设备`)
  handleClose()
}

const handleScanDevices = async () => {
  isScanning.value = true
  manualAddResult.value = null
  
  try {
    // 使用新的仅在线设备扫描函数，不使用缓存
    const devices = await discoverOnlineDevicesOnly()
    // 只过滤v3版本的设备
    scannedDevices.value = devices.filter(device => device.version === 'v3')
    
    if (scannedDevices.value.length === 0) {
      ElMessage.info('未发现v3版本设备')
    } else {
      ElMessage.success(`发现 ${scannedDevices.value.length} 个设备`)
    }
  } catch (error) {
    console.error('扫描设备失败:', error)
    ElMessage.error('扫描设备失败: ' + error.message)
  } finally {
    isScanning.value = false
    emit('scan-complete')
  }
}

const handleManualAdd = async () => {
  if (!manualInput.value.trim()) {
    ElMessage.warning('请输入设备IP地址')
    return
  }
  
  addingDevices.value = true
  manualAddResult.value = null
  
  try {
    const result = await handleManualDeviceDiscovery(manualInput.value)
    manualAddResult.value = result
    
    if (result.success && result.devices && result.devices.length > 0) {
      // 批量发送添加成功的所有设备，避免循环单次发送造成心跳服务拥堵
      emit('devices-added', result.devices)
      
      ElMessage.success(`已添加 ${result.devices.length} 个设备`)
      handleClose()
    }
  } catch (error) {
    console.error('手动添加设备失败:', error)
    ElMessage.error('添加设备失败: ' + error.message)
  } finally {
    addingDevices.value = false
  }
}
</script>

<style scoped>
.scan-tab-content,
.manual-tab-content {
  min-height: 200px;
}

.scan-header {
  display: flex;
  align-items: center;
}

.no-devices {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.manual-input-section {
  margin-bottom: 16px;
}

.input-label {
  font-weight: 500;
  margin-bottom: 8px;
}

.manual-actions {
  display: flex;
  gap: 12px;
}
</style>
