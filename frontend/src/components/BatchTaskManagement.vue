<template>
  <div class="batch-task-management">
    <el-row :gutter="16" style="height: 100%;">
      <!-- 左侧：设备列表 -->
      <el-col :span="9" style="height: 100%;">
        <el-card shadow="hover" style="height: 100%;">
          <template #header>
            <div class="card-header" style="display: flex; justify-content: space-between; align-items: center;">
              <span style="font-weight: bold; font-size: 15px;">📱 {{ $t('batchTask.selectDeviceTitle') }}</span>
              <el-tag v-if="selectedDevices.length > 0" type="success" size="small">
                {{ $t('batchTask.selected') }} {{ selectedDevices.length }} {{ $t('batchTask.unit') }}
              </el-tag>
            </div>
          </template>

          <!-- 搜索栏 -->
          <div style="margin-bottom: 10px;">
            <el-input
              v-model="searchKeyword"
              :placeholder="$t('batchTask.searchPlaceholder')"
              clearable
              prefix-icon="Search"
              size="small"
            />
          </div>

          <!-- 快速操作按钮 -->
          <div style="margin-bottom: 10px; display: flex; align-items: center; justify-content: space-between;">
            <div style="display: flex; gap: 5px; flex-wrap: wrap;">
              <el-button size="small" @click="selectAll">{{ $t('batchTask.selectAll') }}</el-button>
              <el-button size="small" @click="clearSelection">{{ $t('batchTask.clear') }}</el-button>
              <el-button size="small" type="primary" @click="batchOpenProjection">{{ $t('batchTask.openAll') }}</el-button>
              <el-button size="small" type="warning" @click="batchCloseProjection">{{ $t('batchTask.closeAll') }}</el-button>
            </div>
            <el-tag type="info" size="small">
              {{ $t('batchTask.total') }} {{ filteredDevices.length }} {{ $t('batchTask.unitMachines') }}
            </el-tag>
          </div>

          <!-- 设备列表 -->
          <div style="max-height: calc(100vh - 300px); overflow-y: auto;">
            <el-checkbox-group v-model="selectedDevices">
              <div
                v-for="device in filteredDevices"
                :key="device.id"
                class="device-item"
                :class="{ 'selected': selectedDevices.includes(device.id) }"
              >
                <el-checkbox :label="device.id">
                  <div style="display: flex; align-items: center; gap: 8px; width: 100%;">
                    <el-tag type="primary" size="small">{{ device.deviceIP }}</el-tag>
                    <span style="flex: 1; color: #409EFF; font-size: 12px;">{{ formatContainerName(device.containerName) }}</span>
                    <el-tag size="small">{{ device.containerShortID }}</el-tag>
                    <el-button 
                      size="small" 
                      type="success" 
                      @click.stop="openSingleProjection(device)"
                      style="padding: 4px 8px; font-size: 11px;"
                    >
                      {{ $t('batchTask.projection') }}
                    </el-button>
                  </div>
                </el-checkbox>
              </div>
            </el-checkbox-group>

            <el-empty v-if="filteredDevices.length === 0" :description="$t('batchTask.noDevices')" />
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：操作面板 -->
      <el-col :span="15" style="height: 100%;">
        <el-card shadow="hover" style="height: 100%;">
          <template #header>
            <div class="card-header">
              <span style="font-weight: bold;">⚡ {{ $t('batchTask.batchOperationTitle') }}</span>
            </div>
          </template>

          <!-- 操作类型切换 -->
          <el-tabs v-model="activeOperationType" style="height: 100%;">
            <!-- 批量执行命令 -->
            <el-tab-pane :label="'📝 ' + $t('batchTask.batchExecuteCmd')" name="command">
              <div style="height: calc(100vh - 280px); overflow-y: auto; padding: 8px 16px;">
                <!-- 命令输入区 -->
                <div style="margin-bottom: 10px;">
                  <div style="margin-bottom: 6px; display: flex; justify-content: space-between; align-items: center;">
                    <span style="font-weight: bold; font-size: 13px;">{{ $t('batchTask.inputAdbCmd') }}</span>
                    <el-button
                      type="info"
                      size="small"
                      text
                      @click="showCommandHelp"
                      style="font-size: 11px;"
                    >
                      {{ $t('batchTask.viewCmdExample') }}
                    </el-button>
                  </div>
                  <el-input
                    v-model="command"
                    type="textarea"
                    :rows="4"
                    :placeholder="$t('batchTask.cmdExamplePlaceholder')"
                    style="font-family: 'Courier New', monospace; font-size: 12px;"
                  />
                </div>

                <!-- 常用命令模板 -->
<div style="margin-bottom: 10px;">
                  <div style="margin-bottom: 6px; display: flex; justify-content: space-between; align-items: center;">
                    <span style="font-weight: bold; font-size: 13px;">🔖 {{ $t('batchTask.quickCmd') }}</span>
                    <span style="color: #999; font-size: 10px;">{{ $t('batchTask.clickToFill') }}</span>
                  </div>
                  
                  <!-- 分类显示命令 -->
                  <div style="background: #f5f7fa; padding: 6px; border-radius: 4px;">
                    <!-- 基础操作 -->
                    <div style="margin-bottom: 5px;">
                      <div style="font-size: 10px; color: #666; margin-bottom: 3px; font-weight: 600;">📱 {{ $t('batchTask.basicOps') }}</div>
                      <div style="display: flex; flex-wrap: wrap; gap: 3px;">
                        <el-button
                          v-for="template in basicTemplates"
                          :key="template.name"
                          size="small"
                          @click="applyTemplate(template.command)"
                          style="padding: 4px 8px; font-size: 11px; height: 26px;"
                        >
                          {{ template.name }}
                        </el-button>
                      </div>
                    </div>
                    
                    <!-- 应用管理 -->
                    <div style="margin-bottom: 5px;">
                      <div style="font-size: 10px; color: #666; margin-bottom: 3px; font-weight: 600;">📦 {{ $t('batchTask.appOps') }}</div>
                      <div style="display: flex; flex-wrap: wrap; gap: 3px;">
                        <el-button
                          v-for="template in appTemplates"
                          :key="template.name"
                          size="small"
                          @click="applyTemplate(template.command)"
                          style="padding: 4px 8px; font-size: 11px; height: 26px;"
                        >
                          {{ template.name }}
                        </el-button>
                      </div>
                    </div>
                    
                    <!-- 系统信息 -->
                    <div>
                      <div style="font-size: 10px; color: #666; margin-bottom: 3px; font-weight: 600;">🔧 {{ $t('batchTask.sysOps') }}</div>
                      <div style="display: flex; flex-wrap: wrap; gap: 3px;">
                        <el-button
                          v-for="template in systemTemplates"
                          :key="template.name"
                          size="small"
                          @click="applyTemplate(template.command)"
                          style="padding: 4px 8px; font-size: 11px; height: 26px;"
                        >
                          {{ template.name }}
                        </el-button>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 循环次数 + 执行按钮 -->
                <div style="margin-bottom: 10px; display: flex; align-items: center; gap: 8px;">
                  <span style="white-space: nowrap; font-size: 13px; color: #606266;">{{ $t('common.loopCount') }}</span>
                  <el-input-number
                    v-model="loopCount"
                    :min="1"
                    :max="9999"
                    :step="1"
                    size="default"
                    controls-position="right"
                    style="width: 120px;"
                  />
                  <span style="font-size: 12px; color: #999;">次</span>
                  <el-button
                    type="primary"
                    size="default"
                    :disabled="!canExecute"
                    :loading="executing"
                    @click="executeCommand"
                    style="flex: 1; font-weight: 600;"
                  >
                    {{ executing ? $t('batchTask.executing') + currentLoop + '/' + loopCount + $t('batchTask.loopSuffix') + '...' : '🚀 ' + $t('batchTask.executeNow') }}
                  </el-button>
                  <el-button
                    v-if="executing"
                    type="danger"
                    size="default"
                    @click="stopExecution"
                    style="font-weight: 600;"
                  >
                    ⏹ {{ $t('batchTask.stop') }}
                  </el-button>
                </div>


                <!-- 执行结果 -->
                <div v-if="executionResults.length > 0">
                  <div style="margin-bottom: 6px; display: flex; justify-content: space-between; align-items: center;">
                    <span style="font-weight: bold; font-size: 13px;">📊 {{ $t('batchTask.executionResultTitle') }}</span>
                    <div>
                      <el-tag type="success" size="small">✓ {{ successCount }}</el-tag>
                      <el-tag type="danger" size="small" style="margin-left: 5px;">✗ {{ failedCount }}</el-tag>
                    </div>
                  </div>
                  
                  <div style="max-height: 380px; overflow-y: auto; border: 1px solid #e4e7ed; border-radius: 4px; padding: 6px; background: #f9f9f9;">
                    <template v-for="(result, index) in executionResults" :key="index">
                      <!-- 轮次分隔符 -->
                      <div
                        v-if="result.isSeparator"
                        style="display: flex; align-items: center; margin: 6px 0; gap: 6px;"
                      >
                        <div style="flex: 1; height: 1px; background: #dcdfe6;"></div>
                        <span style="font-size: 11px; color: #909399; white-space: nowrap;">{{ $t('batchTask.loopPrefix') }} {{ result.loop }} / {{ result.total }} {{ $t('batchTask.loopSuffix') }}</span>
                        <div style="flex: 1; height: 1px; background: #dcdfe6;"></div>
                      </div>
                      <!-- 正常结果行 -->
                      <div
                        v-else
                        style="margin-bottom: 6px; padding: 6px; background: white; border-radius: 3px; border-left: 3px solid;"
                        :style="{ borderLeftColor: result.success ? '#67C23A' : '#F56C6C' }"
                      >
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 3px;">
                          <div style="display: flex; align-items: center; gap: 6px; flex-wrap: wrap;">
                            <el-tag :type="result.success ? 'success' : 'danger'" size="small" style="height: 20px; padding: 0 6px; font-size: 11px;">
                              {{ result.success ? '✓' : '✗' }}
                            </el-tag>
                            <span style="font-weight: 600; font-size: 12px;">{{ result.deviceIP }}</span>
                            <span v-if="result.containerName" style="color: #409EFF; font-size: 10px;">
                              {{ formatContainerName(result.containerName) }}
                            </span>
                            <span style="color: #999; font-size: 10px;">{{ result.containerShortID }}</span>
                          </div>
                          <span style="color: #999; font-size: 10px;">{{ result.duration }}</span>
                        </div>
                        <div v-if="result.output" style="font-family: 'Courier New', monospace; font-size: 10px; color: #666; background: #f5f5f5; padding: 5px; border-radius: 2px; max-height: 150px; overflow-y: auto; line-height: 1.3; white-space: pre-wrap; word-break: break-all;">
                          {{ result.output }}
                        </div>
                        <div v-if="result.error" style="font-family: 'Courier New', monospace; font-size: 10px; color: #F56C6C; background: #FEF0F0; padding: 5px; border-radius: 2px; margin-top: 3px; line-height: 1.3; white-space: pre-wrap; word-break: break-all;">
                          {{ $t('batchTask.errorStr') }}: {{ result.error }}
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>

    <!-- 命令帮助对话框 -->
    <el-dialog
      v-model="helpDialogVisible"
      :title="'📖 ' + $t('batchTask.cmdReference')"
      width="600px"
    >
      <div style="line-height: 2;">
        <h4>📱 {{ $t('batchTask.commonCmdExamples') }}</h4>
        <ul style="list-style: none; padding: 0;">
          <li><code>input text "Hello World"</code> - 输入文本</li>
          <li><code>input tap 500 800</code> - 点击坐标 (x=500, y=800)</li>
          <li><code>input swipe 300 1000 300 300 500</code> - 滑动屏幕</li>
          <li><code>input keyevent 3</code> - 按Home键 (3=HOME, 4=BACK, 26=POWER)</li>
          <li><code>pm list packages</code> - 列出所有应用包名</li>
          <li><code>pm install -r /sdcard/app.apk</code> - 安装应用</li>
          <li><code>pm uninstall com.example.app</code> - 卸载应用</li>
          <li><code>am start -n com.android.settings/.Settings</code> - 启动设置</li>
          <li><code>dumpsys battery</code> - 查看电池信息</li>
          <li><code>screencap /sdcard/screen.png</code> - 截屏</li>
          <li><code>getprop ro.build.version.release</code> - 获取系统版本</li>
        </ul>
        
        <h4 style="margin-top: 20px;">💡 {{ $t('batchTask.supportedVars') }}</h4>
        <ul style="list-style: none; padding: 0;">
          <li><code>{device_ip}</code> - 设备IP地址</li>
          <li><code>{container_id}</code> - 容器完整ID</li>
          <li><code>{container_short_id}</code> - 容器短ID (前12位)</li>
          <li><code>{container_name}</code> - 容器名称</li>
          <li><code>{timestamp}</code> - 当前时间戳</li>
        </ul>
        
        <h4 style="margin-top: 20px;">⚠️ {{ $t('batchTask.precautions') }}</h4>
        <ul>
          <li>{{ $t('batchTask.note1') }}</li>
          <li>{{ $t('batchTask.note2') }}</li>
          <li>{{ $t('batchTask.note3') }}</li>
        </ul>
      </div>
      <template #footer>
        <el-button @click="helpDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, getCurrentInstance } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ExecuteBatchCommand } from '../../bindings/edgeclient/app'
import { getDevicePassword, startProjection, cleanupProjectionWindows, cleanupProjectionProcesses } from '../services/api.js'

// Props
const props = defineProps({
  devices: {
    type: Array,
    default: () => []
  },
  instances: {
    type: Array,
    default: () => []
  },
  allInstances: {
    type: Array,
    default: () => []
  },
  cloudMachines: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  deviceCloudMachinesCache: {
    type: Map,
    default: () => new Map()
  },
  devicesStatusCache: {
    type: Map,
    default: () => new Map()
  }
})

// State
const { proxy } = getCurrentInstance()
const $t = proxy.$t
const activeOperationType = ref('command') // 'command' 或 'apk'
const selectedDevices = ref([])
const command = ref('')
const searchKeyword = ref('')
const executing = ref(false)
const executionResults = ref([])
const helpDialogVisible = ref(false)
const loopCount = ref(1)       // 循环次数
const currentLoop = ref(0)     // 当前第几轮
const stopFlag = ref(false)    // 停止标志

// 快速命令模板
// 基础操作命令
const basicTemplates = computed(() => [
  { name: $t('batchTask.cmdInputText'), command: 'input text "Hello"' },
  { name: $t('batchTask.cmdTap'), command: 'input tap 500 800' },
  { name: $t('batchTask.cmdSwipe'), command: 'input swipe 300 1000 300 300 500' },
  { name: $t('batchTask.cmdHome'), command: 'input keyevent 3' },
  { name: $t('batchTask.cmdBack'), command: 'input keyevent 4' },
  { name: $t('batchTask.cmdPower'), command: 'input keyevent 26' },
  { name: $t('batchTask.cmdVolUp'), command: 'input keyevent 24' },
  { name: $t('batchTask.cmdVolDown'), command: 'input keyevent 25' },
  { name: $t('batchTask.cmdScreencap'), command: 'screencap /sdcard/upload/screen.png' },
  { name: $t('batchTask.cmdRecord'), command: 'screenrecord /sdcard/upload/video.mp4' },
  { name: $t('batchTask.cmdStartSettings'), command: 'am start -n com.android.settings/.Settings' },
  { name: $t('batchTask.cmdDial'), command: 'am start -a android.intent.action.DIAL' }
])

// 应用管理命令
const appTemplates = computed(() => [
  { name: $t('batchTask.cmdListPkgs'), command: 'pm list packages' },
  { name: $t('batchTask.cmdListSys'), command: 'pm list packages -s' },
  { name: $t('batchTask.cmdListThird'), command: 'pm list packages -3' },
  { name: $t('batchTask.cmdFindApp'), command: 'pm list packages | grep "关键词"' },
  { name: $t('batchTask.cmdClearApp'), command: 'pm clear com.example.app' },
  { name: $t('batchTask.cmdUninstall'), command: 'pm uninstall com.example.app' },
  { name: $t('batchTask.cmdForceStop'), command: 'am force-stop com.example.app' },
  { name: $t('batchTask.cmdStartApp'), command: 'monkey -p com.example.app 1' }
])

// 系统信息命令
const systemTemplates = computed(() => [
  { name: $t('batchTask.cmdSysVer'), command: 'getprop ro.build.version.release' },
  { name: $t('batchTask.cmdModel'), command: 'getprop ro.product.model' },
  { name: $t('batchTask.cmdBrand'), command: 'getprop ro.product.brand' },
  { name: $t('batchTask.cmdAndroidId'), command: 'settings get secure android_id' },
  { name: $t('batchTask.cmdIp'), command: 'ifconfig | grep inet' },
  { name: $t('batchTask.cmdCpu'), command: 'cat /proc/cpuinfo | grep "Hardware"' },
  { name: $t('batchTask.cmdMem'), command: 'cat /proc/meminfo | grep "MemTotal"' },
  { name: $t('batchTask.cmdStorage'), command: 'df -h /sdcard' },
  { name: $t('batchTask.cmdBattery'), command: 'dumpsys battery' },
  { name: $t('batchTask.cmdResolution'), command: 'wm size' },
  { name: $t('batchTask.cmdDensity'), command: 'wm density' },
  { name: $t('batchTask.cmdCurrentAct'), command: 'dumpsys window | grep mCurrentFocus' }
])

// 保留旧的 quickTemplates 以兼容其他地方的引用
const quickTemplates = [
  ...basicTemplates.value,
  ...appTemplates.value,
  ...systemTemplates.value
]

// 格式化设备列表
const allDevices = computed(() => {
  const devices = []
  
  console.log('BatchTask - Props:', {
    allInstances: props.allInstances?.length,
    deviceCloudMachinesCache: props.deviceCloudMachinesCache?.size,
    devicesCount: props.devices?.length
  })
  
  // 创建设备IP到设备对象的映射，用于获取设备版本和在线状态
  const deviceMap = new Map()
  if (props.devices && props.devices.length > 0) {
    props.devices.forEach(device => {
      deviceMap.set(device.ip, device)
    })
  }
  
  // 优先策略：
  // 1. 如果 deviceCloudMachinesCache 存在，从缓存获取所有设备的容器
  // 2. 否则使用 allInstances (当前选中设备的容器)
  
  if (props.deviceCloudMachinesCache && props.deviceCloudMachinesCache.size > 0) {
    // 从缓存获取所有设备的容器
    props.deviceCloudMachinesCache.forEach((containers, deviceIP) => {
      const device = deviceMap.get(deviceIP)
      
      // ✅ 只处理在线设备的容器
      const deviceStatus = device ? props.devicesStatusCache.get(device.id) : null
      if (deviceStatus !== 'online') {
        console.log(`[批量任务] 跳过离线设备: ${deviceIP}`)
        return // 跳过离线设备
      }
      
      if (Array.isArray(containers)) {
        containers.forEach(container => {
          // 只显示运行中的容器
          if (container.status === 'running') {
            // 容器ID优先级: containerId > ID > id
            const containerId = container.containerId || container.ID || container.id
            if (containerId) {
              devices.push({
                id: `${deviceIP}_${containerId}`,
                deviceIP: deviceIP,
                deviceVersion: device?.version || 'v3',
                containerID: containerId,  // 真实的Docker容器ID
                containerShortID: containerId.substring(0, 12),
                containerName: container.name || `容器${containerId.substring(0, 8)}`
              })
            }
          }
        })
      }
    })
  } else if (props.allInstances && props.allInstances.length > 0) {
    // 从当前选中设备的容器列表获取
    props.allInstances.forEach(inst => {
      if (inst.status === 'running') {
        // 容器ID优先级: containerId > ID > id
        const containerId = inst.containerId || inst.ID || inst.id
        if (containerId) {
          const deviceIP = inst.deviceIp || inst.ip
          const device = deviceMap.get(deviceIP)
          
          // ✅ 只处理在线设备的容器
          const deviceStatus = device ? props.devicesStatusCache.get(device.id) : null
          if (deviceStatus !== 'online') {
            return // 跳过离线设备
          }
          
          devices.push({
            id: `${deviceIP}_${containerId}`,
            deviceIP: deviceIP,
            deviceVersion: device?.version || inst.deviceVersion || 'v3',
            containerID: containerId,  // 真实的Docker容器ID
            containerShortID: containerId.substring(0, 12),
            containerName: inst.name || `容器${containerId.substring(0, 8)}`
          })
        }
      }
    })
  }
  
  console.log('BatchTask - allDevices:', devices.length, devices.slice(0, 3))
  
  return devices
})

// 过滤设备列表
const filteredDevices = computed(() => {
  if (!searchKeyword.value) {
    return allDevices.value
  }
  
  const keyword = searchKeyword.value.toLowerCase()
  return allDevices.value.filter(device => 
    device.deviceIP.toLowerCase().includes(keyword) ||
    device.containerName.toLowerCase().includes(keyword) ||
    device.containerID.toLowerCase().includes(keyword)
  )
})

// 判断是否可以执行
const canExecute = computed(() => {
  return selectedDevices.value.length > 0 && command.value.trim().length > 0 && !executing.value
})

// 统计执行结果（过滤分隔符行）
const successCount = computed(() => {
  return executionResults.value.filter(r => !r.isSeparator && r.success).length
})

const failedCount = computed(() => {
  return executionResults.value.filter(r => !r.isSeparator && !r.success).length
})

// 全选
const selectAll = () => {
  selectedDevices.value = filteredDevices.value.map(d => d.id)
}

// 清空选择
const clearSelection = () => {
  selectedDevices.value = []
}

// 应用模板
const applyTemplate = (templateCommand) => {
  command.value = templateCommand
}

// 格式化容器名称，只显示最后一段（例如：192.168.1.1_1770995999275_1_T0001 -> T0001）
const formatContainerName = (name) => {
  if (!name) return ''
  const parts = name.split('_')
  return parts.length > 0 ? parts[parts.length - 1] : name
}

// 显示命令帮助
const showCommandHelp = () => {
  helpDialogVisible.value = true
}

// 停止执行
const stopExecution = () => {
  stopFlag.value = true
  ElMessage.warning('已发送停止信号，等待当前轮次完成后停止...')
}

// 执行命令
const executeCommand = async () => {
  // 验证是否选择了设备
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请选择安卓')
    return
  }
  
  if (!command.value.trim()) {
    ElMessage.warning('请输入命令')
    return
  }

  // 构建目标列表 - 注意字段名要与后端的 Target 结构体匹配
  const targets = selectedDevices.value.map(id => {
    const device = allDevices.value.find(d => d.id === id)
    if (!device) {
      console.error(`找不到设备: ${id}`)
      return null
    }
    const password = getDevicePassword(device.deviceIP)
    return {
      device_ip: device.deviceIP,
      container_id: device.containerID,
      container_name: device.containerName,
      device_version: device.deviceVersion,
      password: password || ''
    }
  }).filter(t => t !== null)
  
  if (targets.length === 0) {
    ElMessage.error('没有找到有效的设备，请重新选择')
    return
  }

  const totalLoops = loopCount.value || 1

  try {
    executing.value = true
    stopFlag.value = false
    executionResults.value = []
    currentLoop.value = 0

    ElMessage.info(`开始执行命令，目标设备: ${targets.length} 个，循环: ${totalLoops} 次`)

    for (let loop = 1; loop <= totalLoops; loop++) {
      if (stopFlag.value) {
        ElMessage.warning(`已在第 ${loop - 1} 轮后停止`)
        break
      }

      currentLoop.value = loop

      console.log(`执行第 ${loop}/${totalLoops} 轮 - command:`, command.value)

      const result = await ExecuteBatchCommand(
        targets,
        command.value,
        `批量执行_第${loop}轮_${new Date().toLocaleString()}`
      )

      console.log(`第 ${loop} 轮结果:`, result)

      // 解析并追加结果
      if (result && result.success && result.history && result.history.results) {
        const loopResults = result.history.results.map(r => {
          const device = allDevices.value.find(d =>
            d.deviceIP === r.device_ip && d.containerID === r.container_id
          )
          return {
            loop,
            deviceIP: r.device_ip,
            containerShortID: r.container_id ? r.container_id.substring(0, 12) : 'unknown',
            containerName: device?.containerName || '',
            success: r.success,
            output: r.output || '',
            error: r.error || '',
            duration: `${r.duration || 0}ms`
          }
        })
        // 多轮时在结果前插入轮次分隔符
        if (totalLoops > 1) {
          executionResults.value.push({ isSeparator: true, loop, total: totalLoops })
        }
        executionResults.value.push(...loopResults)
      } else if (result && !result.success) {
        ElMessage.error(`第 ${loop} 轮执行失败：${result.message || '未知错误'}`)
        break
      } else {
        console.error(`第 ${loop} 轮返回格式异常:`, result)
        ElMessage.error(`第 ${loop} 轮返回格式异常`)
        break
      }
    }

    if (!stopFlag.value) {
      const success = executionResults.value.filter(r => !r.isSeparator && r.success).length
      const failed = executionResults.value.filter(r => !r.isSeparator && !r.success).length
      if (failed === 0) {
        ElMessage.success(`✓ 全部执行成功！(${success} 条，共 ${totalLoops} 轮)`)
      } else {
        ElMessage.warning(`执行完成：成功 ${success} 条，失败 ${failed} 条，共 ${totalLoops} 轮`)
      }
    }

  } catch (error) {
    console.error('执行命令失败:', error)
    if (error.message && error.message.includes('deviceIP')) {
      ElMessage.error('请选择安卓')
    } else {
      ElMessage.error(`执行失败: ${error.message || error}`)
    }
  } finally {
    executing.value = false
    currentLoop.value = 0
    stopFlag.value = false
  }
}

// 单个设备打开投屏
const openSingleProjection = async (device) => {
  try {
    console.log('[批量任务] 打开单个设备投屏:', device.deviceIP, device.containerID)
    
    // 从 allInstances 或 deviceCloudMachinesCache 中找到完整的容器信息
    let fullContainerInfo = null
    
    // 优先从 deviceCloudMachinesCache 获取
    if (props.deviceCloudMachinesCache && props.deviceCloudMachinesCache.size > 0) {
      const containers = props.deviceCloudMachinesCache.get(device.deviceIP)
      if (containers && Array.isArray(containers)) {
        fullContainerInfo = containers.find(c => {
          const cid = c.containerId || c.ID || c.id
          return cid === device.containerID
        })
      }
    }
    
    // 如果没找到，从 allInstances 获取
    if (!fullContainerInfo && props.allInstances && props.allInstances.length > 0) {
      fullContainerInfo = props.allInstances.find(inst => {
        const cid = inst.containerId || inst.ID || inst.id
        const ip = inst.deviceIp || inst.ip
        return cid === device.containerID && ip === device.deviceIP
      })
    }
    
    if (!fullContainerInfo) {
      ElMessage.error('无法获取容器完整信息，请刷新设备列表')
      console.error('[批量任务] 未找到容器信息:', device.containerID)
      return
    }
    
    console.log('[批量任务] 找到完整容器信息:', fullContainerInfo)
    
    // 直接使用完整的容器对象调用投屏API
    await startProjection({ ip: device.deviceIP }, fullContainerInfo)
    ElMessage.success(`已打开 ${device.deviceIP} 的投屏`)
  } catch (error) {
    console.error('[批量任务] 打开投屏失败:', error)
    ElMessage.error(`打开投屏失败: ${error.message || error}`)
  }
}

// 批量打开投屏
const batchOpenProjection = async () => {
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请先选择要打开投屏的设备')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要对选中的 ${selectedDevices.value.length} 个设备打开投屏吗？`,
      '批量打开投屏',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    
    let successCount = 0
    let failCount = 0
    
    for (const deviceId of selectedDevices.value) {
      const device = allDevices.value.find(d => d.id === deviceId)
      if (!device) continue
      
      try {
        // 从 deviceCloudMachinesCache 或 allInstances 获取完整容器信息
        let fullContainerInfo = null
        
        if (props.deviceCloudMachinesCache && props.deviceCloudMachinesCache.size > 0) {
          const containers = props.deviceCloudMachinesCache.get(device.deviceIP)
          if (containers && Array.isArray(containers)) {
            fullContainerInfo = containers.find(c => {
              const cid = c.containerId || c.ID || c.id
              return cid === device.containerID
            })
          }
        }
        
        if (!fullContainerInfo && props.allInstances && props.allInstances.length > 0) {
          fullContainerInfo = props.allInstances.find(inst => {
            const cid = inst.containerId || inst.ID || inst.id
            const ip = inst.deviceIp || inst.ip
            return cid === device.containerID && ip === device.deviceIP
          })
        }
        
        if (!fullContainerInfo) {
          console.error(`[批量任务] 未找到容器信息: ${device.containerID}`)
          failCount++
          continue
        }
        
        await startProjection({ ip: device.deviceIP }, fullContainerInfo)
        successCount++
        console.log(`[批量任务] 成功打开投屏: ${device.deviceIP}`)
      } catch (error) {
        failCount++
        console.error(`[批量任务] 打开投屏失败 ${device.deviceIP}:`, error)
      }
    }
    
    if (successCount > 0) {
      ElMessage.success(`成功打开 ${successCount} 个设备的投屏`)
    }
    if (failCount > 0) {
      ElMessage.warning(`${failCount} 个设备打开投屏失败`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('[批量任务] 批量打开投屏失败:', error)
      ElMessage.error('批量打开投屏失败')
    }
  }
}

// 批量关闭投屏
const batchCloseProjection = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要关闭所有投屏窗口吗？',
      '批量关闭投屏',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 调用清理窗口
    await cleanupProjectionWindows()
    // 调用清理进程
    await cleanupProjectionProcesses()
    
    ElMessage.success('批量关闭投屏成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('[批量任务] 批量关闭投屏失败:', error)
      ElMessage.error('批量关闭投屏失败')
    }
  }
}

// 监听设备变化
watch([() => props.allInstances, () => props.deviceCloudMachinesCache], () => {
  console.log('Device data changed')
  // 设备列表更新时，清理已选择但不存在的设备
  const validIds = allDevices.value.map(d => d.id)
  selectedDevices.value = selectedDevices.value.filter(id => validIds.includes(id))
}, { immediate: true, deep: true })
</script>

<style scoped>
.batch-task-management {
  height: 100%;
  padding: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.device-item {
  padding: 10px;
  margin-bottom: 6px;
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  transition: all 0.3s;
  cursor: pointer;
}

.device-item:hover {
  background-color: #F5F7FA;
  border-color: #409EFF;
}

.device-item.selected {
  background-color: #ECF5FF;
  border-color: #409EFF;
}

code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  color: #E6A23C;
}

:deep(.el-checkbox__label) {
  width: 100%;
}
</style>
