<script setup>

// 返回设备的 host:port，若 ip 已含端口则直接使用，否则追加默认 8000
const getDeviceAddr = (ip) => {
  if (!ip) return ip
  const lastColon = ip.lastIndexOf(':')
  if (lastColon === -1) return ip + ':8000'
  return /^\d+$/.test(ip.slice(lastColon + 1)) ? ip : ip + ':8000'
}


import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick, reactive, getCurrentInstance } from 'vue'
import axios from 'axios'
import { More, Plus, Loading, Warning, Refresh, Download, Delete, Close, Switch, Upload, QuestionFilled, List, Timer, InfoFilled, ArrowRight, ArrowDown, FolderOpened, CircleCheck, CircleClose, Setting, BellFilled, WarningFilled } from '@element-plus/icons-vue'
import { Events } from '@wailsio/runtime'
// 导入模型管理组件
import ModelManagement from './components/modelManagement.vue'
import InstanceManagement from './components/instanceManagement.vue'
import NetworkManagement from './components/networkmanagement.vue'
import BackupManagement from './components/backupManagement.vue'
import AiAssistant from './components/aiAssistant.vue'  // 【使用新的完整版本：设备选择+模型管理+流式优化】
import RpaAgent from './components/rpaAgent.vue'
import CustomerService from './components/customerService.vue'
import OpencecsManagement from './components/opencecsManagement.vue'



// 导入 xterm.js 及其相关依赖
import 'xterm/css/xterm.css'
// 导入Element Plus消息组件
import { ElMessage, ElMessageBox } from 'element-plus'
// 导入云机管理公共函数
import { handleDeleteContainer, handleBatchDeleteContainers, setDependencies } from './services/cloudMachineFunctions.js'
// 导入API服务
import {
  discoverDevices, 
  getContainers, 
  startContainer, 
  stopContainer, 
  deleteContainer, 
  createContainer,
  containersMemoryCache,

  startProjection,

  startProjectionBatchControl,
  stopProjectionBatchControl,
  getDockerNetworks,
  createDockerNetwork,
  deleteDockerNetwork,
  updateDockerNetwork,
  createV0V2Device,
  pullDockerImage,
  getImagePullProgress,
  setDevicePassword,
  closeDevicePassword,
  getDevicePassword,
  saveDevicePassword,
  removeDevicePassword,
  switchPhoneModel,
  resetAndroidContainer,
  restartAndroidContainer,
  CloseProjectionWindow,
  getDeviceVersionInfo,
  upgradeDevice,
  deleteLocalImage,
  getDeviceAllCloudMachines,
  renameAndroidContainer,
  setMacVlanIp,
  setDeviceGPS,
  getAnnouncement,
  getAndroidCacheVersions,
  getAndroidContainersList,
  triggerAndroidRefresh,
  getScreenshotVersions,
  getScreenshots,
  clearScreenshotCache
} from './services/api.js'
// 导入设备心跳检测服务
import { 
  updateMonitoredDevices, 
  startDeviceHeartbeat, 
  getDevicesStatus 
} from './services/deviceHeartbeat.js'
import { countryMap, getCountryEnglishName } from './countryData.js'

import CryptoJS from 'crypto-js';
import QRCode from 'qrcode'


// 导入V3 bindings
import {
  GetLocalImages,
  GetImages,
  DownloadImage,
  CancelImageDownload,
  CancelImageUpload,
  IsImageDownloaded,
  LoadImageToDevice,
  DeleteLocalImage,
  UpgradeDeviceWithNewAPI,
  GetUserDataDir,
  HttpRequest,
  ListSharedDirFiles,
  InstallAPK,
  UploadFileToCloudMachine,
  UploadFileToSharedDir,
  OpenSharedDirectory,
  OpenLocalImageDirectory,
  SelectImageFile,
  SelectVideoFile,
  SyncAuthorization,
  ToggleProjectionWindowTop,
  ArrangeProjectionWindows,
  UpdateMonitoredDevices,
  StartDeviceHeartbeat,
  GetDevicesStatus,
  ResetAllDevicesOffline,
  ForceRefreshDeviceInfo,
  UpdateDevicePasswords,
  GetMirrorList,
  GetPhoneVCode,
  DeviceShake,
  GetUserRabbetList,
  Register,
  UploadGoogleCert,
  GetStoragePath,
  SetStoragePath,
  SelectDirectory,
  GetSharedDirPath,
  SetSharedDirPath,
} from '../bindings/edgeclient/app'

// 导入主机管理组件
import HostManagement from './components/HostManagement.vue'
import StreamManagement from './components/StreamManagement.vue'
import BatchTaskManagement from './components/BatchTaskManagement.vue'
import InterconnectedCloudMachines from './components/InterconnectedCloudMachines.vue'

// 导入更新相关组件
import UpdateMenu from './components/UpdateMenu.vue'
import UpdateDialog from './components/UpdateDialog.vue'

import { useUpdateService } from './services/updateService.js'

const { state: updateState, checkForUpdate: checkForUpdates, getVersionInfo: getUpdateVersionInfo } = useUpdateService()

// 使用自定义国际化 - 不依赖 vue-i18n
const { proxy } = getCurrentInstance()

// 初始化语言设置
try {
  const savedLocale = localStorage.getItem('app-locale')
  if (savedLocale) {
    console.log('Loading saved locale:', savedLocale)
    proxy.$i18n.setLocale(savedLocale)
  } else {
    // 如果没有保存的语言，检测浏览器语言
    const browserLang = navigator.language || navigator.userLanguage
    const defaultLocale = browserLang && browserLang.startsWith('en') ? 'en-US' : 'zh-CN'
    console.log('Setting default locale:', defaultLocale, 'from browser:', browserLang)
    proxy.$i18n.setLocale(defaultLocale)
  }
} catch (e) {
  console.warn('Cannot load saved locale:', e)
}

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

// 响应式数据
const devices = ref([])
const activeDevice = ref(null)
const instances = ref([]) // 当前选中设备的容器列表，每个坑位只显示一个容器，优先显示running状态
const allInstances = ref([]) // 当前选中设备的所有容器，用于备份切换
const cloudMachines = ref([]) // 当前选中设备的云机列表
const deviceCloudMachinesCache = ref(new Map()) // 设备云机缓存，存储每个设备的云机列表，键为设备IP
const cloudManageMode = ref('slot') // slot: 坑位模式, batch: 批量模式
const cloudMachineGroups = ref([]) // 云机分组数据
const selectedCloudDevice = ref(null) // 当前选中的云机设备
const selectedCloudMachines = ref([]) // 选中的云机ID数组

// 更新相关状态
const updateDialogVisible = ref(false) // 更新提示弹窗可见性
const updateInfo = ref(null) // 更新信息

const layoutMode = ref('grid')
const loading = ref(false)
const backupLoading = ref(false) // 切换备份时的加载状态
const activeTab = ref('host-management')
// 批量投屏控制状态管理
// 批量模式：使用单独的 ref
const batchModeProjectionControlling = ref(false)
// 坑位模式：按设备IP记录
const slotModeProjectionControlStatus = ref({})
// 计算当前上下文的批量控制状态
const isBatchProjectionControlling = computed(() => {
  if (cloudManageMode.value === 'batch') {
    // 批量模式：使用全局状态
    return batchModeProjectionControlling.value
  } else {
    // 坑位模式：按设备IP
    const deviceIp = selectedCloudDevice.value?.ip
    return deviceIp ? (slotModeProjectionControlStatus.value[deviceIp] || false) : false
  }
})


// 导入CloudManagement组件
import CloudManagement from './components/CloudManagement.vue'

// 导入批量上传对话框组件
import BatchUploadDialog from './components/BatchUploadDialog.vue'

// 导入语言切换组件
import LanguageSwitcher from './components/LanguageSwitcher.vue'

// 任务队列状态管理
const taskQueue = ref([])
const runningTasksCount = computed(() => {
  return taskQueue.value.filter(task => task.status === 'running').length
})

// ===== 设置弹窗 =====
const settingsDialogVisible = ref(false)
const storagePathInfo = ref({ path: '', isDefault: true, defaultPath: '' })
const settingsLoading = ref(false)

const openSettingsDialog = async () => {
  settingsDialogVisible.value = true
  try {
    const result = await GetStoragePath()
    if (result && result.success !== false) {
      storagePathInfo.value = {
        path: result.path || '',
        isDefault: result.isDefault !== false,
        defaultPath: result.defaultPath || result.path || ''
      }
    }
  } catch (e) {
    console.error('获取保存路径失败', e)
  }
}

const handleSelectDirectory = async () => {
  settingsLoading.value = true
  try {
    const result = await SelectDirectory()
    if (result && result.success && result.path) {
      storagePathInfo.value.path = result.path
      storagePathInfo.value.isDefault = false
    } else if (result && !result.success && result.message !== '用户取消选择') {
      ElMessage.error(result.message || '选择目录失败')
    }
  } catch (e) {
    ElMessage.error('选择目录失败: ' + e.message)
  } finally {
    settingsLoading.value = false
  }
}

const handleSaveSettings = async () => {
  settingsLoading.value = true
  try {
    const result = await SetStoragePath(storagePathInfo.value.path)
    if (result && result.success) {
      ElMessage.success('保存路径设置成功')
      // 刷新路径信息
      const info = await GetStoragePath()
      if (info) {
        storagePathInfo.value = {
          path: info.path || '',
          isDefault: info.isDefault !== false,
          defaultPath: info.defaultPath || info.path || ''
        }
      }
      settingsDialogVisible.value = false

      // 自动刷新镜像、机型、备份页面数据
      fetchLocalCachedImages()
      checkAllImagesDownloadStatus()
      if (modelManagementRef.value) {
        modelManagementRef.value.fetchLocalModels()
        modelManagementRef.value.checkNeedShowButtons()
      }
      if (backupManagementRef.value) {
        backupManagementRef.value.fetchBackups()
      }
    } else {
      ElMessage.error(result?.message || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败: ' + e.message)
  } finally {
    settingsLoading.value = false
  }
}

const handleResetStoragePath = async () => {
  settingsLoading.value = true
  try {
    const result = await SetStoragePath('')
    if (result && result.success) {
      ElMessage.success('已恢复默认路径')
      storagePathInfo.value = {
        path: result.path || storagePathInfo.value.defaultPath,
        isDefault: true,
        defaultPath: storagePathInfo.value.defaultPath
      }

      // 自动刷新镜像、机型、备份页面数据
      fetchLocalCachedImages()
      checkAllImagesDownloadStatus()
      if (modelManagementRef.value) {
        modelManagementRef.value.fetchLocalModels()
        modelManagementRef.value.checkNeedShowButtons()
      }
      if (backupManagementRef.value) {
        backupManagementRef.value.fetchBackups()
      }
    } else {
      ElMessage.error(result?.message || '恢复默认失败')
    }
  } catch (e) {
    ElMessage.error('恢复默认失败: ' + e.message)
  } finally {
    settingsLoading.value = false
  }
}
// ===== 设置弹窗 END =====

// 机型管理组件引用
const modelManagementRef = ref(null)

// 实例管理组件引用
const instanceManagementRef = ref(null)

// 网络管理组件引用
const networkManagementRef = ref(null)

// 备份管理组件引用
const backupManagementRef = ref(null)

// AI助理组件引用
const aiAssistantRef = ref(null)

// RPA Agent 组件引用
const rpaAgentRef = ref(null)

// OpenCecs 管理组件引用
const opencecsManagementRef = ref(null)

// 客服组件引用
const customerServiceRef = ref(null)
// 客服未读消息数
const customerServiceUnreadCount = ref(0)

// 处理未读消息数变化
const handleUnreadCountChange = (count) => {
  customerServiceUnreadCount.value = count
  console.log('客服未读消息数更新:', count)
}



// 临时变量
const tempModelName = ref('') // 临时存储机型选择
const tempModelId = ref('') // 临时存储机型ID
const switchModelType = ref('online') // 机型切换类型: online, local, backup
const switchModelDialogVisible = ref(false) // 机型切换对话框可见性
const currentSwitchContainer = ref(null) // 当前要切换机型的容器
const switchingModel = ref(false) // 机型切换加载状态
// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const contextMenuRef = ref(null)
const contextMenuSlot = ref(0)

// API详情相关
const apiDetailsVisible = ref(false)
const apiDetailsData = ref(null)

// IP测试弹窗
const ipTestVisible = ref(false)
const testIp = ref('')
// S5代理设置弹窗
const s5ProxyDialogVisible = ref(false)
const s5ProxyForm = ref({
  cloudMachineName: '',
  s5ServerAddress: '',
  s5Port: '',
  username: '',
  password: '',
  vpcInfo: '',
  dnsMode: 'server' // local: 本地域名解析, server: 服务端域名解析
})
const s5ProxyLoading = ref(false)

// 重命名相关状态
const renameDialogVisible = ref(false)
const renameForm = ref({
  name: '',
  newName: '',
  prefix: '',
  standby: ''
})
const renameLoading = ref(false)

// 公告弹窗相关状态
const announcementVisible = ref(false)
const announcementData = ref({
  title: '',
  content: '',
  displayDuration: 0
})
let announcementTimer = null
let countdownTimer = null
const countdown = ref(0) // 倒计时秒数

// 处理重命名
const handleRename = (targetContainer = null) => {
  // 获取云机对象：优先使用传入的参数，否则尝试从右键菜单选中的索引获取
  let container = targetContainer
  
  // 如果传入的是鼠标事件对象（可能是直接点击而非传递了row），则视为null
  if (container instanceof Event || (container && container.type === 'click')) {
    container = null
  }
  console.log('处理重命名', instances.value[contextMenuSlot.value], contextMenuSlot.value, instances.value)
  if (!container) {
    // 尝试从右键菜单获取
    // 检查是否是批量模式
    if (cloudManageMode.value === 'batch') {
      // 批量模式下，contextMenuSlot.value 可能不适用或者需要特殊处理
      // 尝试通过 contextMenuContainer 获取
      container = contextMenuContainer.value
    } else {
      container = instances.value.find(item => item.indexNum == contextMenuSlot.value)
    }
  }

  if (!container) {
    ElMessage.warning('未找到选中的云机')
    return
  }
  
  // 获取云机名称
  let containerName = ''
  if (typeof container.name === 'string') {
    containerName = container.name
  } else if (typeof container.ID === 'string') {
    containerName = container.ID
  } else if (container.name && typeof container.name === 'object') {
     // 适配可能的数据结构
     containerName = container.name.name || container.name.id || String(container.name)
  } else {
     containerName = String(container.name || container.ID || '')
  }

  // 处理名称显示，如果包含_则只显示最后一部分
  let simpleName = containerName
  let prefix = ''
  
  // 检查是否符合特定格式（包含下划线）
  if (containerName.includes('_')) {
    const lastUnderscoreIndex = containerName.lastIndexOf('_')
    if (lastUnderscoreIndex !== -1 && lastUnderscoreIndex < containerName.length - 1) {
      simpleName = containerName.substring(lastUnderscoreIndex + 1)
      prefix = containerName.substring(0, lastUnderscoreIndex + 1)
    }
  }

  renameForm.value.name = simpleName // 显示给用户的简单名称（不含前缀）
  //renameForm.value.newName = simpleName // 输入框默认显示简单名称
  renameForm.value.prefix = prefix // 保存前缀
  renameForm.value.standby = containerName // 保存原始完整名称
  
  // 保存设备信息，用于批量模式下也能正确找到设备
  if (container) {
    renameForm.value.deviceIp = container.deviceIp
    renameForm.value.deviceVersion = container.deviceVersion
  }
  
  renameDialogVisible.value = true
  contextMenuVisible.value = false
}

// 提交重命名
const submitRename = async () => {
  if (!renameForm.value.newName || renameForm.value.newName.trim() === '') {
    ElMessage.warning('请输入新名称')
    return
  }

  // 校验新名称：不允许包含中文和特殊字符_
  const chineseRegex = /[\u4e00-\u9fa5]/
  if (chineseRegex.test(renameForm.value.newName) || renameForm.value.newName.includes('_')) {
    ElMessage.error('云机名称不允许包含中文和特殊字符_')
    return
  }
  
  // 构建完整的新名称
  const finalNewName = renameForm.value.prefix + renameForm.value.newName
  
  if (finalNewName === renameForm.value.name) {
    ElMessage.info('名称未发生变化')
    renameDialogVisible.value = false
    return
  }
  
  try {
    renameLoading.value = true
    
    // 确定当前操作的设备
    let targetDevice = activeDevice.value
    
    // 如果是批量模式且没有 activeDevice，尝试从容器信息中获取设备
    if (!targetDevice && cloudManageMode.value === 'batch' && selectedCloudDevice.value) {
        targetDevice = selectedCloudDevice.value
    }
    
    // 优先使用 renameForm 中保存的设备信息（如果有）
    if (renameForm.value.deviceIp) {
      targetDevice = {
        ip: renameForm.value.deviceIp,
        version: renameForm.value.deviceVersion || 'v3'
      }
    }
    
    // 如果还是没有设备信息，尝试从renameForm.value.standby（如果它包含了设备信息）或者通过遍历查找
    if (!targetDevice) {
       // 这是一个最后的尝试，通常 activeDevice 或 selectedCloudDevice 应该被设置
       // 如果容器对象里有 deviceIp，我们可以尝试查找
       // 但在此处我们没有直接访问容器对象，只有名字
    }

    if (!targetDevice) {
        ElMessage.error('未选择设备')
        return
    }

    const result = await renameAndroidContainer(
      targetDevice,
      renameForm.value.standby,
      finalNewName
    )
    
    if (result && result.code === 0) {
      ElMessage.success('修改名称成功')
      renameDialogVisible.value = false
      
      // 清除输入框内容
      renameForm.value.newName = ''
      
      // 刷新容器列表 - 触发后端立即刷新安卓缓存
      await triggerAndroidRefresh([targetDevice.ip])
      
      // 如果是批量模式，更新选中列表中的名称
      if (cloudManageMode.value === 'batch') {
        const oldName = renameForm.value.standby
        const targetMachine = selectedCloudMachines.value.find(
          m => m.deviceIp === targetDevice.ip && m.name === oldName
        )
        
        if (targetMachine) {
          targetMachine.name = finalNewName
          console.log('批量模式下更新云机名称:', oldName, '->', finalNewName)
        }
      }
      
      // 重新初始化分组数据，以更新树形结构
      initCloudMachineGroups()
    } else {
      const errorMsg = result?.message || '修改名称失败'
      ElMessage.error(errorMsg)
    }
    
  } catch (error) {
    console.error('修改名称失败:', error)
    if (error.message === 'Authentication Failed') {
        // 认证失败处理，这里简单提示，实际可能需要弹出认证框
        ElMessage.error('认证失败，请重新登录设备')
    } else {
        const errorMsg = error.response?.data?.message || error.message || '未知错误'
        ElMessage.error(`修改名称失败: ${errorMsg}`)
    }
  } finally {
    renameLoading.value = false
  }
}

// 获取系统公告
const fetchAnnouncement = async () => {
  try {
    const result = await getAnnouncement()
    
    console.log('[fetchAnnouncement] 公告接口返回:', result)
    
    // 检查返回结果是否有效
    if (!result) {
      console.log('[fetchAnnouncement] 公告接口返回为空')
      return
    }
    
    // 检查是否成功获取数据
    if (result.code_id !== 200) {
      console.log('[fetchAnnouncement] 公告接口返回错误:', result.msg)
      return
    }
    
    // 检查data是否为null或空数组（无公告时）
    if (!result.data || (Array.isArray(result.data) && result.data.length === 0)) {
      console.log('[fetchAnnouncement] 当前无公告')
      return
    }
    
    // data为数组，默认取第一条展示
    const firstItem = Array.isArray(result.data) ? result.data[0] : result.data
    if (!firstItem) {
      console.log('[fetchAnnouncement] 公告数据为空')
      return
    }
    
    const { title, content, displayDuration } = firstItem
    console.log('[fetchAnnouncement] 公告数据:', { title, content, displayDuration })
    
    announcementData.value.title = title || '系统通知'
    announcementData.value.content = content || ''
    announcementData.value.displayDuration = displayDuration || 0
    
    // 显示弹窗
    announcementVisible.value = true
    console.log('[fetchAnnouncement] 显示公告弹窗')
    
    // 如果设置了自动关闭时间，启动定时器和倒计时
    if (displayDuration > 0) {
      // 清除之前的定时器（如果存在）
      if (announcementTimer) {
        clearTimeout(announcementTimer)
      }
      if (countdownTimer) {
        clearInterval(countdownTimer)
      }
      
      // 初始化倒计时
      countdown.value = displayDuration
      
      // 启动倒计时（每秒更新）
      countdownTimer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(countdownTimer)
          countdownTimer = null
        }
      }, 1000)
      
      // 设置自动关闭定时器
      announcementTimer = setTimeout(() => {
        announcementVisible.value = false
        console.log('[fetchAnnouncement] 公告自动关闭')
      }, displayDuration * 1000)
      
      console.log(`[fetchAnnouncement] 设置${displayDuration}秒后自动关闭`)
    } else {
      countdown.value = 0
    }
  } catch (error) {
    console.error('[fetchAnnouncement] 获取公告失败:', error)
    // 静默失败，不影响用户使用
  }
}

// 关闭公告弹窗
const closeAnnouncement = () => {
  announcementVisible.value = false
  
  // 清除定时器
  if (announcementTimer) {
    clearTimeout(announcementTimer)
    announcementTimer = null
  }
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
  countdown.value = 0
}


// 设置摇一摇
const handleShake = async () => {
  const container = getCurrentContextMenuContainer()
  if (!container) return

  let targetIP = container.networkName === 'myt' ? container.ip : container.deviceIp
  // OpenCecs 公网设备：deviceIp 含端口（如 1.2.3.4:16039），提取纯 IP
  if (targetIP && targetIP.includes(':')) targetIP = targetIP.split(':')[0]
  let port = extractPort9082(container) || 9082
  
  // 判断是否为V3设备
  // if (container.androidType === 'V3') {
  //   port = extractPort9082(container) || 9082
  //   targetIP = container.deviceIp 
  // } else {
  //   // V2设备尝试使用10008端口
  //   const isDirectNet = container.networkName === 'myt'
    
  //   if (isDirectNet) {
  //     targetIP = container.ip
  //     port = extractPort9082(container) || 9082
  //   }
  // }


  try {
    const password = getDevicePassword(targetIP) || ''
    await DeviceShake(targetIP, Number(port), password)
    ElMessage.success('已发送摇一摇指令')
  } catch (error) {
    console.error('摇一摇失败:', error)
    ElMessage.error(`摇一摇失败: ${error}`)
  }
  closeContextMenu()
}


const gpsDialogVisible = ref(false)
const gpsForm = ref({
  ip: '',
  country: ''
})
const gpsLoading = ref(false)
const gpsContainer = ref(null)

// 设置IP定位
const handleGPS = async () => { 
  const container = getCurrentContextMenuContainer()
  if (!container) return
  
  gpsContainer.value = container
  gpsForm.value.ip = ''
  gpsForm.value.country = ''
  gpsDialogVisible.value = true
  closeContextMenu()
}

const submitGPS = async () => {
  if (!gpsForm.value.country) {
    ElMessage.warning('请选择国家')
    return
  }

  const container = gpsContainer.value
  if (!container) return

  gpsLoading.value = true
  try {
    // 确定目标IP
    // const targetIP = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip 
    //   ? container.ip 
    //   : (container.deviceIp || container.ip)

    // // 确定端口
    // let internalPort = extractPort9082(container) || 9082
    // // 检查是否为V3
    // if (container.androidType === 'V3' || container.version === 'v3' || (activeDevice.value && activeDevice.value.version === 'v3')) {
    //     internalPort = extractPort9082(container) || 9082
    // }
    
    // let finalPort = internalPort
    // // 如果不是直连网络，尝试获取映射端口
    // if (!(container.networkName === 'myt' && container.ip)) {
    //     finalPort = extractPort(container, internalPort) || internalPort
    // }

    let targetIP = container.networkName === 'myt' ? container.ip : container.deviceIp;
    // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
    if (targetIP && targetIP.includes(':')) targetIP = targetIP.split(':')[0]
    let finalPort = extractPort9082(container) || 9082
    
    // 优先使用用户输入的IP，如果为空则使用容器IP
    const currentDeviceIP = gpsForm.value.ip || ''
    
    const countryInfo = countryMap[gpsForm.value.country]
    const lang = countryInfo ? countryInfo.lang : 'en'
    
    // 调用后端API
    await setDeviceGPS(targetIP, finalPort, currentDeviceIP, lang)
    
    ElMessage.success('设置定位指令已发送')
    gpsDialogVisible.value = false

  } catch (error) {
    console.error('设置定位失败:', error)
    ElMessage.error('设置定位失败: ' + (error.message || error))
  } finally {
    gpsLoading.value = false
  }
}

// ===== 上传 Google 证书 =====
const googleCertDialogVisible = ref(false)
const googleCertContainer = ref(null)
const googleCertFile = ref(null)       // File 对象
const googleCertFileName = ref('')     // 显示用文件名
const googleCertLoading = ref(false)
const googleCertInputRef = ref(null)   // 隐藏的 input[file] ref

const handleUploadGoogleCert = () => {
  const container = getCurrentContextMenuContainer()
  if (!container) return
  googleCertContainer.value = container
  googleCertFile.value = null
  googleCertFileName.value = ''
  googleCertDialogVisible.value = true
  closeContextMenu()
}

// 点击"选择文件"按钮触发隐藏 input
const triggerGoogleCertInput = () => {
  googleCertInputRef.value && googleCertInputRef.value.click()
}

// 文件选择后记录
const onGoogleCertFileChange = (e) => {
  const file = e.target.files[0]
  if (!file) return
  googleCertFile.value = file
  googleCertFileName.value = file.name
  // 重置 input 值，允许重复选同一文件也能触发 change
  e.target.value = ''
}

// 确认上传
const submitGoogleCert = async () => {
  if (!googleCertFile.value) {
    ElMessage.warning('请先选择证书文件')
    return
  }
  const container = googleCertContainer.value
  if (!container) return

  let host = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip
    ? container.ip
    : container.deviceIp
  // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
  if (host && host.includes(':')) host = host.split(':')[0]
  const port = extractPort9082(container) || 9082

  googleCertLoading.value = true
  try {
    // 用 FileReader 读取文件内容为 base64，通过 Go IPC 转发（规避跨域）
    const base64Data = await new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => {
        // result 格式: "data:application/octet-stream;base64,xxxx"，只取逗号后面的部分
        const b64 = reader.result.split(',')[1]
        resolve(b64)
      }
      reader.onerror = () => reject(new Error('文件读取失败'))
      reader.readAsDataURL(googleCertFile.value)
    })

    const result = await UploadGoogleCert(host, Number(port), googleCertFileName.value, base64Data)

    if (!result.success) {
      throw new Error(result.message || '上传失败')
    }

    ElMessage.success('Google 证书上传成功')
    googleCertDialogVisible.value = false
  } catch (error) {
    console.error('上传 Google 证书失败:', error)
    ElMessage.error('上传失败: ' + (error.message || error))
  } finally {
    googleCertLoading.value = false
  }
}
// ===== END 上传 Google 证书 =====

// 监听S5代理弹窗关闭，清空vpcInfo
watch(s5ProxyDialogVisible, (newVal) => {
  if (!newVal) {
    s5ProxyForm.value.vpcInfo = ''
  }
})

// 监听备份列表可见且allInstances变化时刷新备份列表
// 使用防抖避免频繁刷新导致选中状态丢失
let backupListRefreshTimer = null
watch(allInstances, () => {
  if (backupListVisible.value) {
    // 清除之前的定时器
    if (backupListRefreshTimer) {
      clearTimeout(backupListRefreshTimer)
    }
    // 延迟刷新，避免频繁触发
    backupListRefreshTimer = setTimeout(() => {
      initBackupList()
    }, 300) // 300ms防抖
  }
}, { deep: true })

// 云机列表加载控制
const cloudMachineLoadingState = ref(new Map()) // 记录每个设备云机加载状态，键为设备IP，值为boolean
const deviceBindStatus = ref(new Map()) // 记录每个设备的绑定状态，键为deviceId，值为0未绑定、1已绑定、2被绑定
const batchLoadTimeout = ref(null) // 批量加载超时定时器
const batchLoadIndex = ref(0) // 批量加载的当前设备索引
// 批量添加设备控制
const isBatchAddingDevices = ref(false) // 是否正在批量添加设备
const batchPendingDevices = ref([]) // 待批量添加的设备队列
const batchAddTimeout = ref(null) // 批量添加防抖定时器
const initGroupsTimeout = ref(null) // 分组初始化防抖定时器
// 型号列表和镜像列表
const phoneModels = ref([]) // 手机型号列表
const localPhoneModels = ref([]) // 本地机型列表
const backupPhoneModels = ref([]) // 备份机型列表
const fetchingBackupModels = ref(false) // 是否正在获取备份机型

// 获取备份机型列表
const fetchBackupModels = async (deviceIp) => {
  if (!deviceIp) {
    console.warn('fetchBackupModels: deviceIp is empty')
    return
  }
  
  fetchingBackupModels.value = true
  // 注意:这里不要立即清空列表,保持之前的数据直到新数据加载完成
  const savedPassword = getDevicePassword(deviceIp)
  let headers = {}
  
  if (savedPassword) {
    // 添加认证头
    const auth = btoa(`admin:${savedPassword}`)
    headers = {
      'Authorization': `Basic ${auth}`
    }
  }
  
  try {
    const response = await axios.get(`http://${getDeviceAddr(deviceIp)}/android/backup/model`, { headers: headers })
    if (response.data && response.data.code === 0) {
      backupPhoneModels.value = response.data.data.list || []
      console.log('备份机型列表加载成功:', backupPhoneModels.value.length)
    } else {
      ElMessage.error(response.data.message || '获取备份机型失败')
      backupPhoneModels.value = []
    }
  } catch (error) {
    console.error('获取备份机型失败:', error)
    ElMessage.error('获取备份机型失败')
    backupPhoneModels.value = []
  } finally {
    fetchingBackupModels.value = false
  }
}



const imageList = ref([]) // 镜像列表
const filteredImageList = ref([]) // 过滤后的镜像列表
const fetchingModels = ref(false) // 是否正在获取型号列表
// 设备选择相关
const selectedHostDevices = ref([]) // 选中的设备列表
const isViewingDeviceDetails = ref(false) // 是否正在查看设备详情
const deviceDetailsDialogVisible = ref(false) // 设备详情弹窗是否可见
const v3DeviceInfoTimer = ref(null) // V3设备信息定时器


// 文件上传相关
const fileInput = ref(null) // 文件输入框引用
const contextMenuContainerId = ref('') // 当前右键菜单操作的容器ID
const contextMenuContainer = ref(null) // 当前右键菜单操作的容器对象
const sharedFilesDialogVisible = ref(false) // 共享文件对话框可见性
const sharedFiles = ref([]) // 共享目录文件列表
const sharedFileTree = ref(null) // 共享目录文件树
const sharedRootPath = ref('') // 共享目录根路径
const selectedFiles = ref([]) // 选中的文件列表
const filesLoading = ref(false) // 文件加载状态
const uploadLoading = ref(false) // 上传加载状态
const fileSortType = ref('time') // 文件排序类型: 'name' | 'time'
const fileSortOrder = ref('desc') // 文件排序顺序: 'asc' | 'desc'

// 单个上传 - 共享目录路径设置
const singleUploadSharedDirInfo = ref({ path: '', isDefault: true, defaultPath: '' })
const singleUploadSharedDirLoading = ref(false)

const loadSingleUploadSharedDirPath = async () => {
  try {
    const result = await GetSharedDirPath()
    if (result) {
      singleUploadSharedDirInfo.value = {
        path: result.path || '',
        isDefault: result.isDefault !== false,
        defaultPath: result.defaultPath || result.path || '',
      }
    }
  } catch (e) {
    console.error('获取共享目录路径失败', e)
  }
}

const handleSelectSingleUploadSharedDir = async () => {
  singleUploadSharedDirLoading.value = true
  try {
    const result = await SelectDirectory()
    if (result && result.success && result.path) {
      const saveResult = await SetSharedDirPath(result.path)
      if (saveResult && saveResult.success) {
        singleUploadSharedDirInfo.value.path = result.path
        singleUploadSharedDirInfo.value.isDefault = false
        ElMessage.success('共享目录路径已保存')
        await loadSharedFiles()
      } else {
        ElMessage.error(saveResult?.message || '保存路径失败')
      }
    } else if (result && !result.success && result.message !== '用户取消选择') {
      ElMessage.error(result.message || '选择目录失败')
    }
  } catch (e) {
    ElMessage.error('选择目录失败')
  } finally {
    singleUploadSharedDirLoading.value = false
  }
}

const handleResetSingleUploadSharedDir = async () => {
  singleUploadSharedDirLoading.value = true
  try {
    const result = await SetSharedDirPath('')
    if (result && result.success) {
      ElMessage.success('已恢复默认目录')
      singleUploadSharedDirInfo.value = {
        path: result.path || singleUploadSharedDirInfo.value.defaultPath,
        isDefault: true,
        defaultPath: singleUploadSharedDirInfo.value.defaultPath,
      }
      await loadSharedFiles()
    } else {
      ElMessage.error(result?.message || '恢复默认失败')
    }
  } catch (e) {
    ElMessage.error('恢复默认失败')
  } finally {
    singleUploadSharedDirLoading.value = false
  }
}
const batchUploadDialogVisible = ref(false) // 批量上传对话框可见性
const batchUploadSelectedMachines = ref([]) // 批量上传选中的云机

// 设置推流弹窗相关
const setStreamDialogVisible = ref(false) // 设置推流弹窗可见性
const streamType = ref('') // 流类型：image-图片, video-视频, app-APP, rtmp-RTMP
const streamFilePath = ref('') // 选择文件或文件夹的路径
const rtmpUrl = ref('') // RTMP推流地址
const setStreamLoading = ref(false) // 设置推流加载状态

const fetchingImages = ref(false) // 是否正在获取镜像列表
// 设备密码管理
const passwordDialogVisible = ref(false)
const passwordForm = ref({
  password: ''
})
const passwordLoading = ref(false)

// 认证对话框
const authDialogVisible = ref(false)
const authForm = ref({
  password: '',
  savePassword: true
})

// 批量认证管理 - 收集所有需要认证的设备，一次性弹出多个输入框
const batchAuthDevices = ref([]) // 待批量认证的设备列表：[{device, callback, password: ''}]
const batchAuthDialogVisible = ref(false) // 批量认证对话框

// 授权同步对话框
const syncAuthDialogVisible = ref(false)
const syncAuthForm = ref({
  username: '',
  password: '',
  saveCredentials: false
})
const syncAuthLoading = ref(false)
const token = ref(localStorage.getItem('token') || null)
const uname = ref(localStorage.getItem('uname') || null)

// 注册对话框
const registerDialogVisible = ref(false)
const registerForm = ref({
  phone: '',
  password: '',
  confirmPassword: '',
  vcode: '',
  vkey: '' // 从获取验证码接口返回的vkey
})
const registerLoading = ref(false)
const sendVcodeLoading = ref(false)
const vcodeCountdown = ref(0)
const vcodeTimer = ref(null)

// 忘记密码对话框
const forgotPasswordDialogVisible = ref(false)
const forgotPasswordForm = ref({
  phone: '',
  newPassword: '',
  confirmPassword: '',
  vcode: '',
  vkey: ''
})
const forgotPasswordErrors = ref({ phone: '', newPassword: '', confirmPassword: '' })
const forgotPasswordLoading = ref(false)
const fpVcodeLoading = ref(false)
const fpVcodeCountdown = ref(0)
const fpVcodeTimer = ref(null)

const authLoading = ref(false)
const authDevice = ref(null)
const authCallback = ref(null)
// 批量认证加载状态
const batchAuthLoading = ref(false)
const batchAuthCollectTimeout = ref(null) // 收集设备的延迟定时器
// 推流设置Loading
// const setStreamLoading = ref(false)
// 同步授权定时器
const syncAuthTimer = ref(null)
// 设备心跳检测定时器
let deviceHeartbeatTimer = null
let heartbeatInitialized = false  // 防止重复初始化
// 安卓容器列表缓存轮询定时器
let androidCacheTimer = null
const androidCacheVersions = ref({})  // 本地版本号快照 {ip: version}
// 设备状态过滤
const deviceFilter = ref('online') // online: 在线设备, offline: 离线设备
const devicesLastUpdateTime = ref(new Map()) // 记录每个设备最后更新时间，键为设备ID，值为时间戳
const devicesStatusCache = ref(new Map()) // 记录每个设备状态，键为设备ID，值为online/offline
// 设备版本信息
const deviceVersionInfo = ref(new Map()) // 记录每个设备的版本信息，键为设备ID，值为{currentVersion, latestVersion}
// 设备固件信息
const deviceFirmwareInfo = ref(new Map()) // 记录每个设备的固件信息，键为设备ID，值为{sdkVersion, deviceModel, originalData}
// 版本信息自动刷新定时器
let versionRefreshInterval = null
// 右侧标签页控制
const currentRightTab = ref('instance') // instance: 实例, image: 镜像, network: 网络, host: 主机

// 监听右侧标签页变化，当切换到镜像管理模块时，默认选中在线镜像
watch(currentRightTab, async (newTab, oldTab) => {
  console.log('右侧标签页变化:', oldTab, '->', newTab)
  
  // 当切换到镜像管理模块时
  if (newTab === 'image' && oldTab !== 'image') {
    console.log('切换到镜像管理模块，默认选中在线镜像')
    
    // 检测镜像缓存是否为空或已被清理
    const cachedImages = getImageListFromLocal()
    const isCacheEmpty = !cachedImages || cachedImages.length === 0
    
    if (isCacheEmpty) {
      console.log('检测到镜像缓存为空，自动重新加载镜像列表...')
      await fetchImageList('')
    }
    
    // 先设置当前选中的镜像分类为在线镜像
    selectedImageCategory.value = 'online'
    
    // 等待DOM更新后再执行后续操作
    await nextTick()
    
    // 加载在线镜像数据，这会自动选中第一个型号的在线镜像
    await switchImageCategory('online')
  }
})

// 从本地存储加载设备列表
const loadDevicesFromLocalStorage = () => {
  try {
    const savedDevices = localStorage.getItem('edgeclient_devices')
    if (savedDevices) {
      const parsedDevices = JSON.parse(savedDevices)
      devices.value = parsedDevices
      console.log('从本地存储加载设备列表成功，共', parsedDevices.length, '个设备')
      
      // ⚠️ 初始化设备状态为 offline，等待心跳检测验证
      // 避免启动时显示假的在线状态
      parsedDevices.forEach(device => {
        devicesStatusCache.value.set(device.id, 'offline')
        devicesLastUpdateTime.value.set(device.id, Date.now())
      })
    }
  } catch (error) {
    console.error('从本地存储加载设备列表失败:', error)
  }
}

// 保存设备列表到本地存储
const saveDevicesToLocalStorage = () => {
  try {
    localStorage.setItem('edgeclient_devices', JSON.stringify(devices.value))
    console.log('设备列表保存到本地存储成功，共', devices.value.length, '个设备')
  } catch (error) {
    console.error('保存设备列表到本地存储失败:', error)
  }
}

// 分组管理方法
const addDeviceGroup = (groupName) => {
  const name = groupName || `新分组${deviceGroups.value.length + 1}`
  if (!deviceGroups.value.includes(name)) {
    deviceGroups.value.push(name)
    saveDeviceGroupsToLocalStorage()
  }
  return name
}

const renameDeviceGroup = (oldName, newName) => {
  const index = deviceGroups.value.indexOf(oldName)
  if (index !== -1 && newName && !deviceGroups.value.includes(newName)) {
    // 更新分组列表
    deviceGroups.value[index] = newName
    // 更新所有属于该分组的设备
    devices.value.forEach(device => {
      if (device.group === oldName) {
        device.group = newName
      }
    })
    // 如果当前筛选的是被重命名的分组，更新筛选
    if (deviceGroupFilter.value === oldName) {
      deviceGroupFilter.value = newName
    }
    saveDevicesToLocalStorage()
    saveDeviceGroupsToLocalStorage()
  }
}

const deleteDeviceGroup = (groupName) => {
  if (groupName === '默认分组') {
    ElMessage.warning('默认分组不能删除')
    return
  }
  const index = deviceGroups.value.indexOf(groupName)
  if (index !== -1) {
    deviceGroups.value.splice(index, 1)
    // 将属于该分组的设备移回默认分组
    devices.value.forEach(device => {
      if (device.group === groupName) {
        device.group = '默认分组'
      }
    })
    // 如果当前筛选的是被删除的分组，重置为全部
    if (deviceGroupFilter.value === groupName) {
      deviceGroupFilter.value = '全部'
    }
    saveDevicesToLocalStorage()
    saveDeviceGroupsToLocalStorage()
    // 重新初始化云机分组
    initCloudMachineGroups()
    ElMessage.success(`分组 "${groupName}" 已删除，设备已移至默认分组`)
  }
}

const moveDeviceToGroup = (deviceId, targetGroup) => {
  const device = devices.value.find(d => d.id === deviceId)
  if (device) {
    const oldGroup = device.group || '默认分组'
    device.group = targetGroup
    saveDevicesToLocalStorage()
    // 重新初始化云机分组
    initCloudMachineGroups()
    ElMessage.success(`设备 ${device.ip} 已从 "${oldGroup}" 移动到 "${targetGroup}"`)
  }
}

const saveDeviceGroupsToLocalStorage = () => {
  try {
    localStorage.setItem('edgeclient_device_groups', JSON.stringify(deviceGroups.value))
  } catch (error) {
    console.error('保存设备分组到本地存储失败:', error)
  }
}

const loadDeviceGroupsFromLocalStorage = () => {
  try {
    const savedGroups = localStorage.getItem('edgeclient_device_groups')
    if (savedGroups) {
      deviceGroups.value = JSON.parse(savedGroups)
    }
  } catch (error) {
    console.error('从本地存储加载设备分组失败:', error)
  }
}

// V3设备信息
const v3DeviceInfo = ref({}) // 存储V3设备详细信息
const v3LatestInfo = ref({}) // 存储最新版本信息
const showUpgradeButton = ref(false) // 是否显示升级按钮
const upgrading = ref(false) // 是否正在升级
const upgradeProgress = ref(0) // SDK升级进度
const v3DeviceInfoLoaded = ref(false) // 标记V3设备信息是否已加载

// 🔧 监听心跳数据变化，自动更新当前激活设备的 v3DeviceInfo
watch([deviceFirmwareInfo, activeDevice], () => {
  if (activeDevice.value && activeDevice.value.version === 'v3') {
    const latestInfo = deviceFirmwareInfo.value.get(activeDevice.value.id)
    if (latestInfo && latestInfo.originalData) {
      // 从心跳数据更新 v3DeviceInfo，保持实时同步
      v3DeviceInfo.value = {
        sdkVersion: latestInfo.sdkVersion || v3DeviceInfo.value.sdkVersion,
        deviceModel: latestInfo.deviceModel || v3DeviceInfo.value.deviceModel,
        originalData: {
          ...v3DeviceInfo.value.originalData,
          ...latestInfo.originalData
        }
      }
      v3DeviceInfoLoaded.value = true
      console.log(`[心跳同步] 更新 v3DeviceInfo: CPU=${latestInfo.originalData.cputemp}°C, 内存=${latestInfo.originalData.memuse}MB`)
    }
  }
}, { deep: true })

const v3DeviceUptimeMinutes = computed(() => {
  if (!v3DeviceInfo.value.originalData?.sysuptime) return '加载中...'
  const seconds = parseInt(v3DeviceInfo.value.originalData.sysuptime)
  if (isNaN(seconds)) return '加载中...'
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}分钟${remainingSeconds}秒`
})

const formatSize = (mbValue) => {
  if (!mbValue || mbValue === '加载中...') return '加载中...'
  const value = parseFloat(mbValue)
  if (isNaN(value)) return mbValue
  if (value >= 1024) {
    return `${(value / 1024).toFixed(2)} GB`
  }
  return `${value} MB`
}
// Docker网络信息
const dockerNetworks = ref([]) // 存储docker网络列表
const dockerNetworksLoading = ref(false) // 标记是否正在加载网络信息
const dockerNetworksError = ref('') // 存储加载错误信息

// 添加macvlan网络弹窗
const addMacvlanDialogVisible = ref(false)
const addMacvlanLoading = ref(false)
const addMacvlanForm = ref({
  networkName: '',
  parentInterface: '',
  subnet: '',
  gateway: '',
  ipRange: '',
  isPrivate: false
})

// 修改网络弹窗
const editNetworkDialogVisible = ref(false)
const editNetworkLoading = ref(false)
const editNetworkForm = ref({
  networkName: '',
  networkID: '',
  subnet: '',
  gateway: '',
  ipRange: '',
  isPrivate: false
})
const currentEditingNetwork = ref(null)

// 镜像列表缓存配置
const IMAGE_CACHE_KEY = 'mytos_image_list' // 本地存储键名
const IMAGE_CACHE_DURATION = 24 * 60 * 60 * 1000 // 缓存有效期24小时
const IMAGE_CACHE_LAST_UPDATE_KEY = 'mytos_image_list_last_update' // 上次更新时间键名

// 镜像卡片相关状态
const localCachedImages = ref([]) // 本地缓存镜像列表
const onlineImagesByModel = ref(new Map()) // 按型号分类的在线镜像列表
const currentOnlineImageModel = ref('') // 当前选中的在线镜像型号
const isLoadingLocalImages = ref(false) // 本地镜像加载状态
const isDownloadingImage = ref(false) // 镜像下载状态
const downloadProgress = ref(0) // 镜像下载进度
const selectedImageCategory = ref('online') // 选中的镜像分类: online 在线镜像, local 本地镜像
const currentDownloadImage = ref(null) // 当前正在下载的镜像
const currentDownloadTaskId = ref(null) // 当前下载任务ID
const downloadStartTime = ref(0) // 下载开始时间戳，用于区分不同的下载会话
const userDataDir = ref('') // 用户数据目录，用于存储下载的镜像
const imageDownloadStatus = ref(new Map()) // 镜像下载状态映射，键为镜像URL，值为boolean
const imageUploadStatus = ref(new Map()) // 镜像上传状态映射，键为镜像URL，值为boolean
const isUploadingImage = ref(false) // 是否正在上传镜像
const uploadProgress = ref(0) // 镜像上传进度
const currentUploadImage = ref(null) // 当前正在上传的镜像
const boxImages = ref([]) // 盒子镜像列表（设备上存在的镜像）
const selectedDeviceForImages = ref(null) // 镜像管理tab选中的设备
const deviceBoxImages = ref([]) // 镜像管理tab选中设备的镜像列表
const isLoadingDeviceImages = ref(false) // 镜像管理tab设备镜像加载状态
const isLoadingBoxImages = ref(false) // 盒子镜像加载状态

// 在线镜像列表筛选状态
const imageCategory = ref('simulator') // 'simulator' | 'container'
const containerAndroidVersion = ref(10) // 10, 12, 14

// 镜像使用说明 - 模拟器 vs 容器对比数据
const imageTypeCompareData = [
  { dimension: '运行原理', simulator: '基于 QEMU 等虚拟机技术，完整模拟 ARM 硬件', container: '基于 Linux 容器（如 Docker）轻量隔离，共享宿主机内核' },
  { dimension: '启动速度', simulator: '较慢，通常需要 30 秒以上', container: '快，通常 3~10 秒即可启动' },
  { dimension: '资源占用', simulator: '较高，每实例需独立分配 CPU/内存', container: '较低，多实例可共享宿主机资源，密度更高' },
  { dimension: '并发数量', simulator: '受限于宿主机性能，同时运行数量较少', container: '可大规模并发，单机支持数十甚至上百个实例' },
  { dimension: '安卓兼容性', simulator: '与真机行为高度一致，兼容性更好', container: '部分依赖底层硬件的功能可能受限' },
  { dimension: '图形界面', simulator: '支持完整 GPU 渲染，画面流畅', container: '使用软件渲染，图形性能稍弱' },
  { dimension: '适用场景', simulator: 'UI 测试、游戏运行、强兼容性需求', container: '自动化脚本、批量任务、高密度部署' },
]

// 监听 tab 切换，处理 P1 不支持 Android 12 的情况
watch(currentOnlineImageModel, (newModel) => {
  if (newModel === 'P1' && containerAndroidVersion.value === 12) {
    containerAndroidVersion.value = 10
  }
})

// 获取筛选后的镜像列表
const getFilteredImages = (images, model) => {
  // console.log('getFilteredImages', images, model)
  if (!images) return []
  
  return images.filter(img => {
    // 1. 模拟器 vs 容器
    if (imageCategory.value === 'simulator') {
      // Q/P 系列模拟器：sys_ver == 5
      return img.sys_ver == 5
    } else {
      // 容器：sys_ver != 5
      if (img.sys_ver == 5) return false
      
      // 2. 安卓版本筛选
      // 假设 sys_ver 对应安卓版本 (10, 12, 14)
      return img.os_ver == `and${containerAndroidVersion.value}`
    }
  })
}

// 设备分组相关
const deviceGroups = ref(['默认分组']) // 设备分组列表
const deviceGroupFilter = ref('全部') // 当前选中的分组过滤
const editingDeviceGroup = ref(null) // 当前编辑分组的设备

// 按IP地址比较（用于正确排序）
const compareIPs = (ip1, ip2) => {
  const parts1 = ip1.split('.').map(Number)
  const parts2 = ip2.split('.').map(Number)
  for (let i = 0; i < 4; i++) {
    if (parts1[i] < parts2[i]) return -1
    if (parts1[i] > parts2[i]) return 1
  }
  return 0
}

// 计算属性：按分组过滤后的设备列表
const filteredDevicesByGroup = computed(() => {
  const filtered = devices.value.filter(device => {
    const status = devicesStatusCache.value.get(device.id) || 'online'
    const matchesStatus = deviceFilter.value === 'online' ? status === 'online' : status === 'offline'
    const matchesGroup = deviceGroupFilter.value === '全部' || device.group === deviceGroupFilter.value
    return matchesStatus && matchesGroup
  }).map(device => {
    const deviceContainers = deviceCloudMachinesCache.value.get(device.ip) || []
    return {
      ...device,
      containers: deviceContainers
    }
  })
  // 按IP地址排序
  return filtered.sort((a, b) => compareIPs(a.ip, b.ip))
})

// 计算属性：按分组整理的设备树形结构
const deviceGroupsTree = computed(() => {
  const groups = {}
  // 确保默认分组存在
  if (!groups['默认分组']) {
    groups['默认分组'] = []
  }
  
  filteredDevicesByGroup.value.forEach(device => {
    const groupName = device.group || '默认分组'
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(device)
  })
  
  return Object.entries(groups).map(([name, devices]) => ({
    name,
    devices,
    count: devices.length
  }))
})

// 计算属性：与当前上传镜像兼容的设备列表
const filteredDevices = computed(() => {
  // devicesLastUpdateTime.value.size
  const filtered = devices.value.filter(device => {
    const status = devicesStatusCache.value.get(device.id) || 'online' // 默认在线
    return deviceFilter.value === 'online' ? status === 'online' : status === 'offline'
  }).map(device => {
    // 添加容器数量信息
    const deviceContainers = deviceCloudMachinesCache.value.get(device.ip) || []
    return {
      ...device,
      containers: deviceContainers
    }
  })
  // 按IP地址排序
  return filtered.sort((a, b) => compareIPs(a.ip, b.ip))
})

// 创建云机对话框
const createDialogVisible = ref(false)
const createDevice = ref(null) // 当前操作的设备
const createMode = ref('batch') // batch: 批量创建, slot: 单个坑位创建
const selectedBatchDevices = ref([]) // 批量设备创建时选中的设备列表

const batchDeviceTypeFilter = ref('p_series') // p_series, other_series

// 监听批量设备类型筛选变化，更新镜像列表
watch(batchDeviceTypeFilter, (newVal) => {
  if (createMode.value === 'multi-device-batch') {
    const type = newVal === 'p_series' ? 'P1' : 'C1'
    filterImageList(type)
  }
})

// 过滤后的设备列表（用于批量设备创建）
const filteredBatchDevices = computed(() => {
  return devices.value.filter(device => {
    const isPSeries = device.name && device.name.toLowerCase().includes('p1')
    if (batchDeviceTypeFilter.value === 'p_series') {
      return isPSeries
    } else {
      return !isPSeries
    }
  })
})

const isPSeriesOrBatchP = computed(() => {
  if (createDevice.value) {
    return createDevice.value.name && createDevice.value.name.toLowerCase().includes('p1')
  }
  if (createMode.value === 'multi-device-batch') {
    return batchDeviceTypeFilter.value === 'p_series'
  }
  return false
})

// 按设备类型过滤本地镜像：P系列设备只显示P系列镜像，非P系列只显示非P系列镜像
// 若镜像的 availableModels 为空（老镜像无元数据），则不过滤，全部显示
const filteredLocalCachedImages = computed(() => {
  const isP = isPSeriesOrBatchP.value
  return localCachedImages.value.filter(image => {
    const models = image.availableModels
    if (!models || models.length === 0) {
      // 无型号信息的镜像不过滤，保持兼容
      return true
    }
    const hasPModel = models.some(m => m && m.toLowerCase().includes('p1'))
    return isP ? hasPModel : !hasPModel
  })
})

const showV3Options = computed(() => {
  if (createDevice.value) {
    return createDevice.value.version === 'v3'
  }
  if (createMode.value === 'multi-device-batch') {
    if (selectedBatchDevices.value.length > 0) {
        const selected = devices.value.filter(d => selectedBatchDevices.value.includes(d.ip))
        return selected.some(d => d.version === 'v3')
    }
    return true
  }
  return false
})

watch(selectedBatchDevices, async (newVal) => {
  if (createMode.value === 'multi-device-batch' && newVal.length > 0) {
     const firstDevice = devices.value.find(d => newVal.includes(d.ip))
     if (firstDevice) {
        // Update image list based on device type
        const type = firstDevice.name || 'C1'
        await fetchImageList(type)
        
        if (firstDevice.version === 'v3') {
           await getV3PhoneModels(firstDevice.ip)
           await getLocalPhoneModels(firstDevice.ip)
           await fetchBackupModels(firstDevice.ip)
           fetchNetworkCards(firstDevice.ip)
        }
     }
  }
})

const currentSlot = ref(0) // 当前操作的坑位
const createLoading = ref(false) // 创建云机时的加载状态
const createCancelled = ref(false) // 创建操作取消标志
// 添加蒙版相关状态
const sdkLoadingVisible = ref(false) // SDK加载蒙版显示状态
const sdkLoadingMessage = ref('加载MYT SDK中') // SDK加载提示信息
const sdkLoadingProgress = ref(0) // SDK加载进度（0-100）
const createForm = ref({
  createType: 'simulator', // simulator, container
  
  // Container mode specific fields
  containerAndroidVersion: '10', // 10, 12, 14
  containerSandboxMode: true,
  containerEnforce: true, // 安全模式，默认开启
  containerDataDiskSize: '16G',
  containerName: 'T000',
  containerCount: 1,
  containerImageSelect: '',
      containerCustomImageUrl: '',
      containerResolution: '720x1280x320',
  containerCustomResolution: {
    width: '',
    height: '',
    dpi: ''
  },
  containerDns: '223.5.5.5',
  containerCustomDns: '',
  containerNetworkCardType: 'private', // private-私有网卡, public-公有网卡
  containerMytBridgeName: '', // 容器模式的 myt_bridge网卡名
  containerMacVlanIp: '', // 容器模式的 MacVlan IP

  // Simulator mode (shared/original) fields
  name: 'T000',
  androidVersion: '14', // 安卓版本：11, 13, 14, 15, 16
  modelName: 'random',
  count: 1,
  startSlot: 1,
  imageSelect: 'registry.magicloud.tech/magicloud/dobox-android13:Q1',
  customImageUrl: '',
  imageCategory: 'online', // online: 在线镜像, local: 本地镜像
  localImageUrl: '', // 本地镜像URL
  imageSource: 'pc', // pc: 在线镜像, local: 本地镜像
  cacheToLocal: false, // 是否缓存到本地创建
  networkMode: 'bridge',
  ipaddr: '',
  resolution: 'default',
  customResolution: {
    width: '',
    height: '',
    dpi: ''
  },
  sandboxSize: 28,
  dns: '223.5.5.5',
  customDns: '',
  countryCode: 'CN', // 机型国家代码，默认为中国
  // S5代理设置（SDK版本>=25时支持）
  s5Type: '0', // 代理类型，0-不开启代理，1-本地域名解析tun2socks，2-服务器域名解析tun2proxy
  s5IP: '', // 代理服务器IP
  s5Port: '', // 代理服务器端口
  s5User: '', // 代理用户名
  s5Password: '', // 代理密码
  enableMagisk: false,
  enableGMS: false,
  enforce: true, // 安全模式，默认开启
  longitud: '',  // 经度
  latitude: '',  // 纬度
  lockScreenPassword: '',  // 锁屏密码
  modelType: 'online', // 机型类型：online-在线机型，local-本地机型
  modelStatic: 'random', // 备份机型名称
  localModel: 'random', // 本地机型名称
  // 网络管理分组
  vpcGroupId: '', // 选择的分组ID
  vpcNodeId: '', // 选择的节点ID
  vpcSelectMode: 'specified', // specified-指定节点, random-随机节点
  randomFile: false, // 随机系统文件，默认关闭
  networkCardType: 'private', // private-私有网卡, public-公有网卡
  mytBridgeName: '', // myt_bridge网卡名
  macVlanIp: '', // macVlan IP
  adbPort: 5555, // ADB端口，默认555，设置0不开启ADB
  selectedSlots: [] // 批量创建选中的坑位
})

// 监听 P 系列设备选择变化，如果选中了 Android 12 则重置为 10；非P设备时清除超出12的坑位
watch(isPSeriesOrBatchP, (isP) => {
  if (isP && createForm.value.containerAndroidVersion === '12') {
    createForm.value.containerAndroidVersion = '10'
  }
  // 切换为非P设备时，过滤掉超出12坑位范围的已选坑位
  if (!isP && createForm.value.selectedSlots && createForm.value.selectedSlots.length > 0) {
    createForm.value.selectedSlots = createForm.value.selectedSlots.filter(s => s <= 12)
  }
})

const handleModelTypeChange = async (val) => {
  if (val === 'online') {
    if (!createForm.value.modelName) createForm.value.modelName = 'random'
    // 确保在线机型列表已加载
    if (phoneModels.value.length === 0 && createDevice.value && createDevice.value.ip) {
      await getV3PhoneModels(createDevice.value.ip)
    }
  } else if (val === 'local') {
    if (!createForm.value.localModel) createForm.value.localModel = 'random'
    // 确保本地机型列表已加载
    if (localPhoneModels.value.length === 0 && createDevice.value && createDevice.value.ip) {
      await getLocalPhoneModels(createDevice.value.ip)
    }
  } else if (val === 'backup') {
    if (!createForm.value.modelStatic) createForm.value.modelStatic = 'random'
    // 每次切换到备份机型都重新加载列表
    if (createDevice.value && createDevice.value.ip) {
      await fetchBackupModels(createDevice.value.ip)
    }
  }
}

const createDeviceFirmwareInfo = computed(() => {
  if (!createDevice.value) return null
  const cached = deviceFirmwareInfo.value.get(createDevice.value.id)
  if (cached) return cached
  if (createDevice.value.version === 'v3') {
    const info = v3DeviceInfo.value
    if (info?.originalData?.ip && info.originalData.ip === createDevice.value.ip) {
      return info
    }
  }
  return null
})

const createDeviceStorageInfo = computed(() => {
  const info = createDeviceFirmwareInfo.value
  const total = Number(info?.originalData?.mmctotal)
  const used = Number(info?.originalData?.mmcuse)
  if (!total || Number.isNaN(total) || Number.isNaN(used)) {
    return { text: '加载中...', isLow: false, remainingGb: null, remainingPercent: 0 }
  }
  const remainingMb = total - used
  const remainingGb = remainingMb / 1024
  const remainingPercent = Math.max(0, Math.min(100, (remainingMb / total) * 100))
  return {
    text: formatSize(remainingMb),
    isLow: remainingGb < 10,
    remainingGb,
    remainingPercent
  }
})

const createDeviceApiVersion = computed(() => {
  if (!createDevice.value) return '加载中...'
  const cached = deviceVersionInfo.value.get(createDevice.value.id)
  if (cached?.currentVersion !== undefined && cached?.currentVersion !== null) {
    return String(cached.currentVersion)
  }
  return '加载中...'
})

const createDeviceApiLatestVersion = computed(() => {
  if (!createDevice.value) return null
  const cached = deviceVersionInfo.value.get(createDevice.value.id)
  if (cached?.latestVersion !== undefined && cached?.latestVersion !== null) {
    return String(cached.latestVersion)
  }
  return null
})

const createDeviceApiVersionNumber = computed(() => {
  const version = createDeviceApiVersion.value
  if (!version || version === '加载中...') return null
  const value = parseFloat(version)
  return Number.isNaN(value) ? null : value
})

const createDeviceApiLatestVersionNumber = computed(() => {
  const version = createDeviceApiLatestVersion.value
  if (!version || version === '加载中...') return null
  const value = parseFloat(version)
  return Number.isNaN(value) ? null : value
})

const createDeviceApiNeedsUpgrade = computed(() => {
  const current = createDeviceApiVersionNumber.value
  const latest = createDeviceApiLatestVersionNumber.value
  if (current === null || latest === null) return false
  return current < latest
})

// 判断当前创建设备是否为公网设备（OpenCecs），公网设备 IP 格式为 publicIp:publicPort
const isPublicNetworkDevice = computed(() => {
  return createDevice.value && createDevice.value.ip && createDevice.value.ip.includes(':')
})

// 判断当前活动设备是否为公网设备（用于更新镜像弹窗）
const isActiveDevicePublic = computed(() => {
  return activeDevice.value && activeDevice.value.ip && activeDevice.value.ip.includes(':')
})

// 机型国家列表相关状态
const countryList = ref([]) // 机型国家列表
const countryListLoading = ref(false) // 机型国家列表加载状态
const vpcGroupList = ref([]) // 网络分组列表
const vpcNodeList = ref([]) // 网络节点列表
const networkCardList = ref([]) // 网卡列表
const fetchingNetworkCards = ref(false) // 网卡列表加载状态
const hasMacVlan = ref(false) // 是否存在macVlan配置
const currentDeviceMacVlanInfo = ref({ subnet: '', gw: '' }) // 当前设备的MacVlan子网和网关信息

// 镜像筛选条件
const imageFilters = reactive({
  name: '',
  url: '',
  includeCompatible: false // 是否包含兼容镜像
})

// 计算属性：根据筛选条件过滤后的镜像列表
const isPSeries = computed(() => {
  if (!createDevice.value || !createDevice.value.name) return false
  const name = createDevice.value.name.toLowerCase()
  // 假设 P 系列以 p 开头 (e.g. p1, p2)
  return name.startsWith('p')
})

const filteredContainerImages = computed(() => {
  const images = imageList.value
  const version = createForm.value.containerAndroidVersion
  
  if (!images || images.length === 0) return []
  
  return images.filter(img => {
    // 检查 Android 版本 (os_ver)
    // 例如: version='10' -> target='and10'
    if (img.sys_ver == 5) {
      return false
    }

    const targetVer = `and${version}`
    if (img.os_ver !== targetVer) {
      return false
    }
    
    // 确定用于过滤的设备名称
    let targetDeviceName = ''
    if (createDevice.value) {
        targetDeviceName = createDevice.value.name
    } else if (createMode.value === 'multi-device-batch') {
        if (selectedBatchDevices.value.length > 0) {
           const firstDevice = devices.value.find(d => selectedBatchDevices.value.includes(d.ip))
           if (firstDevice) {
             targetDeviceName = firstDevice.name
           }
        }
        
        if (!targetDeviceName) {
           // 批量模式下根据筛选器判断
           if (batchDeviceTypeFilter.value === 'p_series') {
               targetDeviceName = 'P1'
           } else {
               targetDeviceName = 'C1' // 其他系列默认为C1，根据实际情况调整
           }
        }
    } else {
        return false
    }
    
    // 参考 filteredImageList 的过滤逻辑，使用 getCompatibleTypes
    const compatibleTypes = getCompatibleTypes(targetDeviceName)
    
    if (img.ttype && compatibleTypes.includes(img.ttype)) {
      return true
    }
    
    if (Array.isArray(img.ttype2)) {
      for (const t of img.ttype2) {
        if (compatibleTypes.includes(t)) {
          return true
        }
      }
    }
    
    return false
  })
})

const filteredContainerImagesForUpdate = computed(() => {
  console.log('filteredContainerImagesForUpdate', createDevice.value)
  const images = imageList.value
  if (!images || images.length === 0) return []
  
  if (!updateImageContainer.value || !updateImageContainer.value.image) return images

  // 找到当前镜像对应的 os_ver
  const currentImageUrl = updateImageContainer.value.image
  const currentImage = images.find(img => img.url === currentImageUrl)
  
  // 如果找不到当前镜像的信息，返回所有镜像 (或者根据需求处理)
  if (!currentImage || !currentImage.os_ver) {
      return images
  }

  const targetVer = currentImage.os_ver

  return images.filter(img => {
    // 检查 Android 版本 (os_ver)
    if (img.os_ver !== targetVer) {
      return false
    }

    if (img.sys_ver == 5) {
      return false
    }
    
    // 如果是 P 系列设备，要求 ttype2 包含 'p1_v2'
    const isP1V2 = Array.isArray(img.ttype2) && img.ttype2.includes(createDevice.value.name)
    return isP1V2
    // if (isPSeries.value) {
    //   return isP1V2
    // } else {
    //   return !isP1V2
    // }
  })
})

// 监听 filteredContainerImages 变化，自动选中第一条
// V3 更新镜像：在 filteredImageList（已按设备类型过滤）基础上，额外按当前容器的 os_ver 过滤
const filteredImageListForUpdate = computed(() => {
  const images = filteredImageList.value
  if (!images || images.length === 0) return []
  
  if (!updateImageContainer.value || !updateImageContainer.value.image) return images

  // 从完整镜像列表中查找当前容器镜像的 os_ver
  const currentImageUrl = updateImageContainer.value.image
  const currentImage = imageList.value.find(img => img.url === currentImageUrl)
  
  // 如果找不到当前镜像或没有 os_ver 信息，返回原列表
  if (!currentImage || !currentImage.os_ver) {
    return images
  }

  const targetVer = currentImage.os_ver
  return images.filter(img => img.os_ver === targetVer)
})

// 根据选择的安卓版本过滤在线镜像列表（用于模拟器创建）
const androidVersionFilteredImageList = computed(() => {
  const images = filteredImageList.value
  if (!images || images.length === 0) return []
  const ver = createForm.value.androidVersion
  if (!ver) return images
  const osVer = `and${ver}`
  return images.filter(img => img.os_ver === osVer)
})

// 根据选择的安卓版本过滤在线机型列表
const androidVersionFilteredPhoneModels = computed(() => {
  const models = phoneModels.value
  if (!models || models.length === 0) return []
  const ver = createForm.value.androidVersion
  if (!ver) return models
  return models.filter(m => {
    if (!m.android_version) return true // 无该字段时不过滤
    return String(m.android_version) === String(ver)
  })
})

// 监听安卓版本变化，自动选中第一条镜像
watch(() => createForm.value.androidVersion, () => {
  if (createForm.value.createType !== 'container' && createForm.value.imageCategory === 'online') {
    const filtered = androidVersionFilteredImageList.value
    if (filtered && filtered.length > 0) {
      createForm.value.imageSelect = filtered[0].url
    } else {
      createForm.value.imageSelect = ''
    }
  }
})

watch(filteredContainerImages, (newVal) => {
  if (createForm.value.createType === 'container') {
    if (newVal && newVal.length > 0) {
      createForm.value.containerImageSelect = newVal[0].url
    } else {
      createForm.value.containerImageSelect = ''
    }
  }
}, { immediate: true })

// 监听 createType 变化，切换到容器模式时自动选中第一条
watch(() => createForm.value.createType, (newVal) => {
  if (newVal === 'container') {
    const images = filteredContainerImages.value
    if (images && images.length > 0) {
      createForm.value.containerImageSelect = images[0].url
    } else {
      createForm.value.containerImageSelect = ''
    }
  }
})

const filteredImageListWithFilters = computed(() => {
  let result = [...filteredImageList.value]
  
  // 根据名称筛选
  if (imageFilters.name) {
    const nameFilter = imageFilters.name.toLowerCase()
    result = result.filter(image => image.name.toLowerCase().includes(nameFilter))
  }
  
  // 根据URL筛选
  if (imageFilters.url) {
    const urlFilter = imageFilters.url.toLowerCase()
    result = result.filter(image => image.url.toLowerCase().includes(urlFilter))
  }
  
  // 根据当前选择的设备型号筛选
  if (activeDevice.value) {
    const currentDeviceType = activeDevice.value.name || 'C1'
    if (imageFilters.includeCompatible) {
      // 包含兼容镜像：筛选ttype等于当前型号或ttype2中包含当前型号的镜像
      result = result.filter(image => {
        return image.ttype === currentDeviceType || 
               (Array.isArray(image.ttype2) && image.ttype2.includes(currentDeviceType))
      })
    } else {
      // 不包含兼容镜像：只筛选ttype等于当前型号的镜像
      result = result.filter(image => image.ttype === currentDeviceType)
    }
  }
  
  return result
})

// 计算属性：从镜像列表中提取可用的设备型号
const availableModels = computed(() => {
  const models = new Set()
  imageList.value.forEach(image => {
    if (image.ttype) {
      models.add(image.ttype)
    }
    if (Array.isArray(image.ttype2)) {
      image.ttype2.forEach(ttype => {
        models.add(ttype)
      })
    }
  })
  return Array.from(models).sort()
})

// 计算属性：将设备中已下载的镜像与线上镜像列表对应起来
const matchedBoxImages = computed(() => {
  return boxImages.value.map(boxImage => {
    // 遍历在线镜像列表，查找匹配的镜像
    for (const image of imageList.value) {
      // 从在线镜像URL中提取镜像名称（去掉registry部分）
      let onlineImageName = image.url
      if (onlineImageName.includes('/')) {
        const parts = onlineImageName.split('/')
        onlineImageName = parts.slice(1).join('/')
      }
      
      // 检查盒子镜像名称是否包含在线镜像名称的关键部分
      if (boxImage.name.includes(onlineImageName) || 
          boxImage.name.includes(onlineImageName.replace(':', '_')) ||
          boxImage.name.includes(onlineImageName.split('/').pop())) {
        // 找到匹配的在线镜像，返回合并后的镜像信息
        return {
          ...boxImage,
          onlineImageName: image.name, // 使用在线镜像列表中的名称
          onlineImageUrl: image.url,   // 保存在线镜像的完整URL
          matched: true
        }
      }
    }
    
    // 没有找到匹配的在线镜像，返回原始镜像信息
    return {
      ...boxImage,
      onlineImageName: boxImage.name, // 使用设备中镜像的名称
      onlineImageUrl: boxImage.url,   // 使用设备中镜像的URL
      matched: false
    }
  })
})

// 获取指定设备的镜像列表（用于镜像管理Tab）
const fetchDeviceBoxImages = async (device) => {
  if (!device) {
    deviceBoxImages.value = []
    return
  }
  
  try {
    isLoadingDeviceImages.value = true
    console.log('获取设备镜像列表，设备:', device.ip)
    
    // 从本地存储获取密码
    const savedPassword = getDevicePassword(device.ip);
    const deviceImages = await GetImages(device.ip, device.version || 'v3', savedPassword || '')
    
    deviceBoxImages.value = processRawDeviceImages(deviceImages)
    
    console.log('获取设备镜像列表成功，共', deviceBoxImages.value.length, '个镜像')
  } catch (error) {
    console.error('获取设备镜像列表失败:', error)
    ElMessage.error('获取设备镜像列表失败: ' + error.message)
    deviceBoxImages.value = []
  } finally {
    isLoadingDeviceImages.value = false
  }
}

// 镜像管理tab选中的设备变化处理
const handleDeviceSelectForImages = (device) => {
  if (device) {
    selectedDeviceForImages.value = device
    fetchDeviceBoxImages(device)
  }
}

// 计算属性：将选定设备中已下载的镜像与线上镜像列表对应起来（用于镜像管理Tab）
const matchedDeviceBoxImages = computed(() => {
  return deviceBoxImages.value.map(boxImage => {
    // 遍历在线镜像列表，查找匹配的镜像
    for (const image of imageList.value) {
      // 从在线镜像URL中提取镜像名称（去掉registry部分）
      let onlineImageName = image.url
      if (onlineImageName.includes('/')) {
        const parts = onlineImageName.split('/')
        onlineImageName = parts.slice(1).join('/')
      }
      
      // 检查盒子镜像名称是否包含在线镜像名称的关键部分
      if (boxImage.name.includes(onlineImageName) || 
          boxImage.name.includes(onlineImageName.replace(':', '_')) ||
          boxImage.name.includes(onlineImageName.split('/').pop())) {
        // 找到匹配的在线镜像，返回合并后的镜像信息
        return {
          ...boxImage,
          onlineImageName: image.name, // 使用在线镜像列表中的名称
          onlineImageUrl: image.url,   // 保存在线镜像的完整URL
          matched: true
        }
      }
    }
    
    // 没有找到匹配的在线镜像，返回原始镜像信息
    return {
      ...boxImage,
      onlineImageName: boxImage.name, // 使用设备中镜像的名称
      onlineImageUrl: boxImage.url,   // 使用设备中镜像的URL
      matched: false
    }
  })
})

// 删除设备镜像（用于镜像管理Tab）
const handleDeleteDeviceImage = async (image) => {
    if (!selectedDeviceForImages.value) return

    // 尝试从image对象中获取ID，如果没有则使用name
    let imageId = image.name
    if (image.original && (image.original.id || image.original.Id)) {
        imageId = image.original.id || image.original.Id
    } else if (image.url) {
        // 兼容旧逻辑，使用url作为id
         imageId = image.url
    }

    ElMessageBox.confirm(
    `确定要删除镜像 "${image.name}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
        const savedPassword = getDevicePassword(selectedDeviceForImages.value.ip)

        // 尝试使用axios直接调用设备API删除镜像
        const apiUrl = `http://${selectedDeviceForImages.value.version === 'v3' ? getDeviceAddr(selectedDeviceForImages.value.ip) : selectedDeviceForImages.value.ip + ':81'}/android/image?image=${encodeURIComponent(imageId)}`;
        const headers = {};
        if (savedPassword) {
            const auth = btoa(`admin:${savedPassword}`);
            headers['Authorization'] = `Basic ${auth}`;
        }
      
        const axiosResponse = await axios.delete(apiUrl, { headers });
        const response = axiosResponse.data;
        
        if (response && response.code === 0) {
            ElMessage.success('镜像删除成功')
            // 刷新列表
            fetchDeviceBoxImages(selectedDeviceForImages.value)
        } else {
            ElMessage.error(response?.message || '镜像删除失败')
        }
    } catch(error) {
        console.error('删除镜像失败:', error)
        
        // 处理认证错误
        if (error.response?.status === 401 || error.response?.data?.code === 61) {
            console.log('删除镜像认证失败，需要显示认证对话框')
            // 显示认证对话框
            showAuthDialog(selectedDeviceForImages.value, async (password) => {
                // 认证成功后重新尝试删除
                console.log('认证回调被调用，开始重试删除镜像，密码长度:', password ? password.length : 0)
                try {
                    const apiUrl = `http://${selectedDeviceForImages.value.version === 'v3' ? getDeviceAddr(selectedDeviceForImages.value.ip) : selectedDeviceForImages.value.ip + ':81'}/android/image?image=${encodeURIComponent(imageId)}`;
                    const auth = btoa(`admin:${password}`);
                    const headers = {
                        'Authorization': `Basic ${auth}`
                    };
                    
                    console.log('重试删除镜像 API URL:', apiUrl)
                    console.log('重试删除镜像请求头:', headers)
                    
                    const axiosResponse = await axios.delete(apiUrl, { headers });
                    const response = axiosResponse.data;
                    
                    console.log('重试删除镜像响应:', response)
                    
                    if (response && response.code === 0) {
                        ElMessage.success('镜像删除成功')
                        // 刷新列表
                        fetchDeviceBoxImages(selectedDeviceForImages.value)
                    } else {
                        ElMessage.error(response?.message || '镜像删除失败')
                    }
                } catch (retryError) {
                    console.error('认证后重试删除镜像失败:', retryError)
                    console.error('错误详情:', retryError.response?.data)
                    ElMessage.error('删除镜像失败: ' + (retryError.response?.data?.message || retryError.message))
                }
            })
            return
        }
        
        ElMessage.error('删除镜像失败: ' + error.message)
    }
  }).catch(() => {
    // 取消删除
  })
}

// 刷新镜像列表
const refreshImageList = async () => {
  if (activeDevice.value) {
    const deviceType = activeDevice.value.name || 'C1'
    await fetchImageList(deviceType)
    
    // 刷新盒子镜像列表
    await fetchBoxImages()
    
    // 检查在线镜像的上传状态
    await checkOnlineImagesUploadStatus()
    
    // 热更新测试：显示一个消息
    ElMessage.info('镜像列表已刷新 - 热更新测试')
  } else {
    await fetchImageList('')
  }
}


//打开本地镜像目录
const handleOpen = async () => {
  try {
    const result = await OpenLocalImageDirectory()
    if (result.success) {
      ElMessage.success('已打开本地镜像目录')
    } else {
      ElMessage.error(result.message)
    }
  } catch (error) {
    console.error('打开本地镜像目录失败:', error)
    ElMessage.error('打开本地镜像目录失败')
  }
}


// 刷新在线镜像


// 刷新本地镜像


// 刷新盒子镜像



// 重置筛选条件



// 处理镜像列表排序
const handleImageSort = (column) => {
  const { prop, order } = column
  if (!prop || !order) return
  
  filteredImageList.value.sort((a, b) => {
    const aVal = a[prop] || ''
    const bVal = b[prop] || ''
    
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      return order === 'ascending' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal)
    }
    
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      return order === 'ascending' ? aVal - bVal : bVal - aVal
    }
    
    return 0
  })
}

// 格式化文件大小为人类可读的格式
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  if (typeof bytes !== 'number') return bytes
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 计算MacVlan IP范围
const calculateIpRange = (startIp, count) => {
  if (!startIp || count <= 0) return ''
  
  try {
    const parts = startIp.split('.').map(Number)
    if (parts.length !== 4 || parts.some(isNaN)) return '无效的IP地址'
    
    // 计算结束IP
    let [a, b, c, d] = parts
    d += count - 1
    
    // 处理进位
    while (d > 255) {
      d -= 256
      c += 1
    }
    while (c > 255) {
      c -= 256
      b += 1
    }
    while (b > 255) {
      b -= 256
      a += 1
    }
    
    if (a > 255) return '超出IP范围'
    
    const endIp = `${a}.${b}.${c}.${d}`
    return `${startIp} - ${endIp}`
  } catch (error) {
    return '计算失败'
  }
}

// 获取MacVlan IP输入框的placeholder
const getMacVlanIpPlaceholder = () => {
  if (currentDeviceMacVlanInfo.value.subnet) {
    // 从subnet中提取建议的起始IP
    // 例如: 10.10.0.0/16 -> 建议从 10.10.0.10 开始
    const subnetParts = currentDeviceMacVlanInfo.value.subnet.split('/')
    if (subnetParts.length > 0) {
      const ipParts = subnetParts[0].split('.')
      if (ipParts.length === 4) {
        // 建议从 x.x.x.10 开始,避免保留地址
        ipParts[3] = '10'
        return `建议起始IP: ${ipParts.join('.')} (请确保未被占用)`
      }
    }
  }
  return '请输入起始IP地址,例如: 10.10.0.10'
}

// 更新镜像对话框
const updateImageDialogVisible = ref(false)
const updateImageContainer = ref(null) // 当前要更新镜像的容器
const updateImageLoading = ref(false) // 更新镜像时的加载状态

// 批量更新镜像对话框
const batchUpdateImageDialogVisible = ref(false)
const batchUpdateImageLoading = ref(false)
const batchUpdateImageStatusText = ref('') // 当前操作进度文字
// 分组数据结构：每项对应一种设备类型（P系列 / 非P系列）
// { groupKey: 'p'|'non-p', groupLabel: string, deviceName: string,
//   containers: [], hasV2: bool, hasV3: bool,
//   androidType: 'V3'|'V2',          // 用户选择的版本
//   v2AndroidVersion: 10|12|14,      // V2时选择的安卓版本
//   selectedUrl: '',                 // 选中的镜像URL
//   customUrl: '' }                  // 自定义地址
const batchUpdateImageGroups = ref([])

// 根据设备名称判断是否P系列
const isBatchImagePSeries = (deviceName) => {
  const n = (deviceName || '').toLowerCase()
  return n.includes('p1') || n.startsWith('p')
}

// 为某个分组获取 V3 镜像列表
const getBatchUpdateV3List = (deviceName) => {
  const images = imageList.value
  if (!images || images.length === 0) return []
  const compatibleTypes = getCompatibleTypes(deviceName || '')
  return images.filter(img => {
    if (img.sys_ver != 5) return false
    if (!img.ttype && !img.ttype2) return true
    if (img.ttype && compatibleTypes.includes(img.ttype)) return true
    if (Array.isArray(img.ttype2)) {
      for (const t of img.ttype2) {
        if (compatibleTypes.includes(t)) return true
      }
    }
    return false
  })
}

// 为某个分组获取 V2 镜像列表
const getBatchUpdateV2List = (deviceName, androidVersion) => {
  const images = imageList.value
  if (!images || images.length === 0) return []
  const ver = `and${androidVersion}`
  return images.filter(img => {
    if (img.sys_ver == 5) return false
    if (img.os_ver !== ver) return false
    if (deviceName) {
      return Array.isArray(img.ttype2) && img.ttype2.includes(deviceName)
    }
    return true
  })
}
const updateImageForm = ref({
  imageSelect: '',
  customImageUrl: '',
  modelName: '',
  enableMagisk: false,
  enableGMS: false,
  dns: '',
  customDns: '',
  resolution: 'default',
  customResolution: {
    width: '720',
    height: '1280',
    dpi: '320'
  },
  vpcGroupId: '',
  vpcNodeId: '',
  vpcSelectMode: 'specified',
  randomFile: false, // 随机系统文件，默认关闭
  enforce: true, // 安全模式，默认开启
  networkCardType: 'private', // private-私有网卡, public-公有网卡
  mytBridgeName: '', // myt_bridge网卡名
  macVlanIp: '' // macVlan IP
})
// 上一次的镜像选择记录
const lastImageSelection = ref({
  imageSelect: '',
  customImageUrl: '',
  imageCategory: 'online',
  localImageUrl: '',
  imageSource: 'pc'
})

// 处理镜像选择变化
const handleImageSelectChange = (value) => {
  if (value !== 'custom') {
    createForm.value.customImageUrl = ''
  }
}

// 处理网络模式变化
const handleNetworkModeChange = (value) => {
  if (value !== 'myt') {
    createForm.value.ipaddr = ''
  }
}

const upgradeCreateDeviceApiVersion = async () => {
  if (!createDevice.value) return
  const device = createDevice.value
  const versionInfo = deviceVersionInfo.value.get(device.id)
  if (!versionInfo || !versionInfo.latestVersion) {
    ElMessage.error(`设备 ${device.ip} 未获取到最新版本信息`)
    return
  }
  const refreshDeviceApiVersion = () => {
    lastCheckTime.value.set(device.id, 0)
    addToVersionCheckQueue(device, true)
    batchProcessVersionCheckQueue()
  }
  try {
    await ElMessageBox.confirm(`确定要升级设备 ${device.ip} 的API版本吗？升级过程可能需要几分钟时间。`, '升级确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    let password = device.password
    if (!password) {
      password = getDevicePassword(device.ip)
    }
    const result = await UpgradeDeviceWithNewAPI(device.ip, versionInfo.latestVersion, password || '')
    if (result.success) {
      ElMessage.success(`设备 ${device.ip} 升级成功: ${result.message}`)
      refreshDeviceApiVersion()
      return
    }
    if (result.errorType === 'auth_required') {
      const savedPassword = getDevicePassword(device.ip)
      if (!savedPassword) {
        showAuthDialog(device, async (newPassword) => {
          const retryResult = await UpgradeDeviceWithNewAPI(device.ip, versionInfo.latestVersion, newPassword || '')
          if (retryResult.success) {
            ElMessage.success(`设备 ${device.ip} 升级成功: ${retryResult.message}`)
            refreshDeviceApiVersion()
          } else {
            ElMessage.error(`设备 ${device.ip} 升级失败: ${retryResult.message}`)
          }
        })
        ElMessage.warning('设备需要认证，请输入设备密码')
      } else {
        ElMessage.error('设备密码错误，请重新输入')
        showAuthDialog(device, async (newPassword) => {
          const retryResult = await UpgradeDeviceWithNewAPI(device.ip, versionInfo.latestVersion, newPassword || '')
          if (retryResult.success) {
            ElMessage.success(`设备 ${device.ip} 升级成功: ${retryResult.message}`)
            refreshDeviceApiVersion()
          } else {
            ElMessage.error(`设备 ${device.ip} 升级失败: ${retryResult.message}`)
          }
        })
      }
      return
    }
    ElMessage.error(`设备 ${device.ip} 升级失败: ${result.message}`)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`设备 ${device.ip} 升级失败: ${error.message}`)
    }
  }
}

const slotStates = ref({})
const runningSlots = ref(new Set())

// ---- 坑位到期时间缓存工具 ----
const SLOT_CACHE_TTL = 24 * 3600 * 1000 // 缓存有效期 1 天（毫秒）
const SLOT_WARN_SECONDS = 3 * 24 * 3600  // 3 天内到期 → 即将到期

const getSlotCacheKey = (deviceId) => `slotStates_${deviceId}`

// 将 API 返回的 child 对象转为 { [slot]: { state, expireTs } }
const convertChild = (child) => {
  const now = Math.floor(Date.now() / 1000)
  const converted = {}
  for (const [slot, expireTs] of Object.entries(child)) {
    const ts = parseInt(expireTs, 10)
    let state
    if (isNaN(ts) || ts === 0 || ts < now) {
      state = 2 // 已到期
    } else if (ts - now < SLOT_WARN_SECONDS) {
      state = 1 // 即将到期（3天内）
    } else {
      state = 0 // 正常有效
    }
    converted[slot] = { state, expireTs: ts }
  }
  return converted
}

// 从缓存读取，返回 converted 对象或 null（缓存不存在/已过期）
const loadSlotCache = (deviceId) => {
  try {
    const raw = localStorage.getItem(getSlotCacheKey(deviceId))
    if (!raw) return null
    const { cachedAt, data } = JSON.parse(raw)
    if (Date.now() - cachedAt > SLOT_CACHE_TTL) return null
    return data
  } catch {
    return null
  }
}

// 将 converted 对象写入缓存
const saveSlotCache = (deviceId, converted) => {
  try {
    localStorage.setItem(getSlotCacheKey(deviceId), JSON.stringify({
      cachedAt: Date.now(),
      data: converted
    }))
  } catch {}
}

// 判断缓存中是否有某坑位已到期（需要重新查询）
const hasCacheExpiredSlot = (cached, slot) => {
  if (!cached) return true
  const info = cached[slot]
  if (!info) return false // 坑位不存在于 child，不触发重查
  return info.state === 2
}

// 拉取并更新 slotStates，同时写缓存
const fetchAndCacheSlotStates = (deviceId) => {
  return GetUserRabbetList(deviceId).then(res => {
    console.log('GetUserRabbetList result:', res)
    if (res.data && res.data.data && res.data.data.length > 0) {
      const converted = convertChild(res.data.data[0].child || {})
      saveSlotCache(deviceId, converted)
      slotStates.value = converted
    } else {
      slotStates.value = {}
    }
  }).catch(err => {
    console.error('GetUserRabbetList error:', err)
    slotStates.value = {}
  })
}

// 加载坑位状态：优先使用缓存，若目标坑位已到期则重新请求
const loadSlotStates = (deviceId, slot) => {
  const cached = loadSlotCache(deviceId)
  if (cached && !hasCacheExpiredSlot(cached, String(slot))) {
    // 命中缓存且目标坑位未到期，直接使用
    slotStates.value = cached
  } else {
    // 缓存过期、不存在，或目标坑位已到期 → 重新请求
    fetchAndCacheSlotStates(deviceId)
  }
}

// Helper to parse running slots
const updateRunningSlots = (device, containers) => {
  runningSlots.value.clear()
  const list = device.version === 'v3' ? (containers.data?.list || []) : (containers || [])
  list.forEach(c => {
    if (c.status === 'running') {
      let snum
      if (device.version === 'v3') {
        snum = parseInt(c.snum)
      } else {
        const name = c.names?.[0] || c.Name || ''
        const match = name.match(/(\d+)/)
        if (match) snum = parseInt(match[1])
      }
      if (snum) runningSlots.value.add(snum)
    }
  })
}

const getSlotClass = (slot) => {
  const info = slotStates.value[slot]
  const state = info ? info.state : undefined
  if (state === 0) return 'slot-blue'
  if (state === 1) return 'slot-yellow'
  if (state === 2) return 'slot-red'
  return 'slot-gray'
}

// Watcher for selectedSlots removed as logic is moved to submit handler

// 显示创建云机对话框
const showCreateDialog = async (device, mode, slot = 0, localImage = null) => {
  console.log('showCreateDialog', device)

  if (device && device.id) {
    loadSlotStates(device.id, slot)
    
    // Fetch containers to check running status
    getContainers(device).then(containers => {
      updateRunningSlots(device, containers)
    }).catch(err => {
      console.error('getContainers error:', err)
      runningSlots.value.clear()
    })
  } else {
    slotStates.value = {}
    runningSlots.value.clear()
  }

  createDevice.value = device
  createMode.value = mode
  currentSlot.value = slot
  
  // 使用上一次的选择初始化表单
  if (createMode.value === 'multi-device-batch' || createMode.value === 'batch') {
    // 从localStorage加载上次选择的坑位
    const savedSlots = localStorage.getItem('createDialog_selectedSlots')
    let previousSelectedSlots = savedSlots ? JSON.parse(savedSlots) : []
    
    // 根据当前设备类型过滤坑位：非P设备只允许 1-12 的坑位
    const currentDeviceIsP = device && device.name && device.name.toLowerCase().startsWith('p')
    const maxSlotForDevice = currentDeviceIsP ? 24 : 12
    previousSelectedSlots = previousSelectedSlots.filter(s => s >= 1 && s <= maxSlotForDevice)
    
    createForm.value = {
      createType: 'simulator',
      selectedSlots: previousSelectedSlots, // 恢复上次选择的坑位（已过滤超出范围的）
      
      // Container mode specific fields
      containerAndroidVersion: '10', // 10, 12, 14
      containerSandboxMode: true,
      containerEnforce: true, // 安全模式，默认开启
      containerDataDiskSize: '16G',
      containerName: 'T000',
      containerCount: 1,
      containerImageSelect: '',
      containerCustomImageUrl: '',
      containerResolution: '720x1280x320',
      containerCustomResolution: {
        width: '720',
        height: '1280',
        dpi: '320'
      },
      containerDns: '223.5.5.5',
      containerCustomDns: '',
      containerNetworkCardType: 'private', // private-私有网卡, public-公有网卡
      containerMytBridgeName: '', // 容器模式的 myt_bridge网卡名
      containerMacVlanIp: '', // 容器模式的 MacVlan IP

      name: 'T000',
      androidVersion: '14', // 安卓版本：11, 13, 14, 15, 16
      modelName: 'random', // 默认随机机型
      modelType: mode === 'multi-device-batch' ? 'online' : 'online', // 默认在线机型
      count: 1,
      startSlot: 1,
      imageSelect: '',
      customImageUrl: '',
      imageCategory: 'online',
      localImageUrl: '',
      imageSource: 'pc',
      cacheToLocal: false,
      networkMode: 'bridge',
      ipaddr: '',
      resolution: '720x1280x320',
      customResolution: {
        width: '720',
        height: '1280',
        dpi: '320'
      },
      sandboxSize: 28,
      dns: '223.5.5.5',
      customDns: '',
      countryCode: 'CN', // 默认为中国
      // S5代理设置
      s5Type: '0',
      s5IP: '',
      s5Port: '',
      s5User: '',
      s5Password: '',
      enableMagisk: false,
      enableGMS: false,
      enforce: true, // 安全模式，默认开启
      adbPort: 5555, // ADB端口，默认555，设置0不开启ADB
      // 网络管理分组
      vpcGroupId: '', // 选择的分组ID
      vpcNodeId: '', // 选择的节点ID
      vpcSelectMode: 'specified', // specified-指定节点, random-随机节点
      networkCardType: 'private', // private-私有网卡, public-公有网卡
      mytBridgeName: '' // myt_bridge网卡名
    }
  } else {
    createForm.value = {
      createType: 'simulator',

      // Container mode specific fields
      containerAndroidVersion: '10', // 10, 12, 14
      containerSandboxMode: true,
      containerEnforce: true, // 安全模式，默认开启
      containerDataDiskSize: '16G',
      containerName: 'T000',
      containerCount: 1,
      containerImageSelect: '',
      containerCustomImageUrl: '',
      containerResolution: '720x1280x320',
      containerCustomResolution: {
        width: '720',
        height: '1280',
        dpi: '320'
      },
      containerDns: '223.5.5.5',
      containerCustomDns: '',
      containerNetworkCardType: 'private', // private-私有网卡, public-公有网卡
      containerMytBridgeName: '', // 容器模式的 myt_bridge网卡名
      containerMacVlanIp: '', // 容器模式的 MacVlan IP

      name: 'T000',
      androidVersion: '14', // 安卓版本：11, 13, 14, 15, 16
      modelName: 'random', // 默认随机机型
      modelType: 'online', // 默认在线机型
      count: 1,
      startSlot: slot,
      imageSelect: '',
      customImageUrl: '',
      imageCategory: 'online',
      localImageUrl: '',
      imageSource: 'pc',
      cacheToLocal: false,
      networkMode: 'bridge',
      ipaddr: '',
      resolution: 'default',
      customResolution: {
        width: '720',
        height: '1280',
        dpi: '320'
      },
      sandboxSize: 28,
      dns: '223.5.5.5',
      customDns: '',
      countryCode: 'CN', // 默认为中国
      // S5代理设置
      s5Type: '0',
      s5IP: '',
      s5Port: '',
      s5User: '',
      s5Password: '',
      enableMagisk: false,
      enableGMS: false,
      enforce: true, // 安全模式，默认开启
      adbPort: 5555, // ADB端口，默认5555，设置0不开启ADB
      // 网络管理分组
      vpcGroupId: '', // 选择的分组ID
      vpcNodeId: '', // 选择的节点ID
      vpcSelectMode: 'specified', // specified-指定节点, random-随机节点
      networkCardType: 'private', // private-私有网卡, public-公有网卡
      mytBridgeName: '' // myt_bridge网卡名
    }
  }
  
  // 获取设备类型
  const deviceType = device ? (device.name || 'C1') : 'C1'
  
  // 先显示创建对话框
  createDialogVisible.value = true

  if (mode === 'multi-device-batch') {
    selectedBatchDevices.value = []
    batchDeviceTypeFilter.value = 'p_series' // 默认选择P系列
    createForm.value.modelType = 'online' // 批量模式下默认为在线机型
    // 加载镜像列表
    const type = batchDeviceTypeFilter.value === 'p_series' ? 'P1' : 'C1'
    await fetchImageList(type)
    // 加载本地镜像列表
    await fetchLocalCachedImages()
    return
  }

  if (device && device.version === 'v3') {
    if (!deviceFirmwareInfo.value.get(device.id)) {
      await fetchV3DeviceInfo(device)
    }
  }

  if (device && !deviceVersionInfo.value.get(device.id)) {
    addToVersionCheckQueue(device, true)
    batchProcessVersionCheckQueue()
  }
  
  // 获取网卡列表 (仅V3设备)
  if (device.version === 'v3') {
    fetchNetworkCards()
  }
  
  // 加载本地镜像列表，确保用户可以在对话框中选择本地镜像
  await fetchLocalCachedImages()
  
  // 如果提供了本地镜像，直接设置表单
  if (localImage) {
    createForm.value.imageCategory = 'local'
    createForm.value.localImageUrl = localImage.url
    createForm.value.imageSource = 'local'
    console.log('从本地镜像创建云机，镜像URL:', localImage.url)
    
    // 如果是V3设备，仍然需要获取型号列表和本地机型列表
    if (device.version === 'v3') {
      await getV3PhoneModels(device.ip)
      await getLocalPhoneModels(device.ip)
      await fetchBackupModels(device.ip)
    }
    
    return
  }
  
  // 如果是V0-V2设备，显示SDK加载蒙版并执行加载流程
  if (device.version !== 'v3') {
    try {
      // 显示SDK加载蒙版
      sdkLoadingVisible.value = true
      sdkLoadingMessage.value = '加载MYT SDK中'
      
      // 调用后端创建V0-V2设备的SDK
      await createV0V2Device(device)
      
      // 更新提示信息
      sdkLoadingMessage.value = '加载镜像列表中'
      
      // 获取镜像列表
      await fetchImageList(deviceType)
      
      // 设置镜像选择：始终默认选第一个镜像（按安卓版本过滤后）
      const filtered = androidVersionFilteredImageList.value
      if (filtered && filtered.length > 0) {
        createForm.value.imageSelect = filtered[0].url
        createForm.value.imageCategory = 'online'
        createForm.value.localImageUrl = ''
        createForm.value.imageSource = 'pc'
      }
      
      // 隐藏SDK加载蒙版
      sdkLoadingVisible.value = false
    } catch (error) {
      console.error('加载MYT SDK失败:', error)
      ElMessage.error(`加载MYT SDK失败：${error.message}`)
      // 隐藏SDK加载蒙版
      sdkLoadingVisible.value = false
      // 关闭创建对话框
      createDialogVisible.value = false
    }
  } else {
    // V3设备正常流程
    // 如果是V3设备，获取型号列表
    await getV3PhoneModels(device.ip)
    
    // 获取本地机型列表
    await getLocalPhoneModels(device.ip)

    // 获取备份机型列表
    await fetchBackupModels(device.ip)
    
    // 获取机型国家列表
    await getCountryList(device.ip)
    
    // 获取镜像列表
    await fetchImageList(deviceType)
    
    // 设置镜像选择：始终默认选第一个镜像（按安卓版本过滤后）
    const filteredByVer = androidVersionFilteredImageList.value
    if (filteredByVer && filteredByVer.length > 0) {
      createForm.value.imageSelect = filteredByVer[0].url
      createForm.value.imageCategory = 'online'
      createForm.value.localImageUrl = ''
      createForm.value.imageSource = 'pc'
    }
  }
  
  // 重置网络管理相关字段
  createForm.value.vpcGroupId = ''
  createForm.value.vpcNodeId = ''
  createForm.value.vpcSelectMode = 'specified'
  vpcGroupList.value = []
  vpcNodeList.value = []
  
  // 如果是V3设备，获取分组列表
  if (device.version === 'v3') {
    await fetchVpcGroupList(device.ip)
  }
}

// 获取V3手机型号列表，参考api/main.go中的getV3PhoneModels实现
const getV3PhoneModels = async (deviceIP) => {
  try {
    fetchingModels.value = true
    // 尝试使用已保存的密码
    const savedPassword = getDevicePassword(deviceIP)
    let headers = {}
    
    if (savedPassword) {
      // 添加认证头
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/android/phoneModel?page=0`, {
      headers: headers
    })
    
    // 解析响应数据
    if (response.data.code === 0 && response.data.data) {
      const result = response.data.data
      phoneModels.value = result.list || []
      console.log('获取V3手机型号列表成功，共', phoneModels.value.length, '个型号')
    } else if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
      // 认证失败，显示认证对话框
      return new Promise((resolve, reject) => {
        showAuthDialog({ ip: deviceIP, version: 'v3' }, async (password) => {
          try {
            const auth = btoa(`admin:${password}`)
            const authResponse = await axios.get(`http://${getDeviceAddr(deviceIP)}/android/phoneModel?page=0`, {
              headers: {
                'Authorization': `Basic ${auth}`
              }
            })
            
            if (authResponse.data.code === 0 && authResponse.data.data) {
              const result = authResponse.data.data
              phoneModels.value = result.list || []
              console.log('获取V3手机型号列表成功，共', phoneModels.value.length, '个型号')
              resolve(phoneModels.value)
            } else {
              ElMessage.error('获取手机型号列表失败: ' + (authResponse.data.message || '未知错误'))
              reject(new Error('获取手机型号列表失败'))
            }
          } catch (error) {
            console.error('获取V3手机型号列表失败:', error)
            ElMessage.error('获取手机型号列表失败: ' + error.message)
            reject(error)
          }
        })
        
        // 30秒后检查是否有响应
        setTimeout(() => {
          console.log('30秒超时检查')
        }, 30000)
      })
    } else {
      console.error('获取V3手机型号列表失败:', response.data.message)
      ElMessage.error('获取手机型号列表失败: ' + (response.data.message || '未知错误'))
    }
  } catch (error) {
    console.error('获取V3手机型号列表失败:', error)
    ElMessage.error('获取手机型号列表失败: ' + error.message)
  } finally {
    fetchingModels.value = false
  }
}

// 获取本地机型列表，使用 /phoneModel 接口
const getLocalPhoneModels = async (deviceIP) => {
  try {
    fetchingModels.value = true
    const savedPassword = getDevicePassword(deviceIP)
    let headers = {}
    
    if (savedPassword) {
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/phoneModel`, {
      headers: headers
    })
    
    // 解析响应数据
    if (response.data.code === 0 && response.data.data) {
      const result = response.data.data
      localPhoneModels.value = result.list || []
      console.log('获取本地机型列表成功，共', localPhoneModels.value.length, '个机型')
    } else if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
      return new Promise((resolve, reject) => {
        showAuthDialog({ ip: deviceIP, version: 'v3' }, async (password) => {
          try {
            const auth = btoa(`admin:${password}`)
            const authResponse = await axios.get(`http://${getDeviceAddr(deviceIP)}/phoneModel`, {
              headers: {
                'Authorization': `Basic ${auth}`
              }
            })
            
            if (authResponse.data.code === 0 && authResponse.data.data) {
              const result = authResponse.data.data
              localPhoneModels.value = result.list || []
              console.log('获取本地机型列表成功，共', localPhoneModels.value.length, '个机型')
              resolve(localPhoneModels.value)
            } else {
              ElMessage.error('获取本地机型列表失败: ' + (authResponse.data.message || '未知错误'))
              reject(new Error('获取本地机型列表失败'))
            }
          } catch (error) {
            console.error('获取本地机型列表失败:', error)
            ElMessage.error('获取本地机型列表失败: ' + error.message)
            reject(error)
          }
        })
      })
    } else {
      console.error('获取本地机型列表失败:', response.data.message)
      ElMessage.error('获取本地机型列表失败: ' + (response.data.message || '未知错误'))
    }
  } catch (error) {
    console.error('获取本地机型列表失败:', error)
    ElMessage.error('获取本地机型列表失败: ' + error.message)
  } finally {
    fetchingModels.value = false
  }
}

// 计算当前显示的机型列表
const displayedModels = computed(() => {
  let models = []
  let useNameAsId = false
  
  switch (switchModelType.value) {
    case 'online':
      models = phoneModels.value
      // 切换机型弹窗：根据容器当前镜像的安卓版本过滤在线机型
      if (currentSwitchContainer.value && currentSwitchContainer.value.image) {
        const currentImageUrl = currentSwitchContainer.value.image
        const currentImage = imageList.value.find(img => img.url === currentImageUrl)
        if (currentImage && currentImage.os_ver) {
          const verMatch = currentImage.os_ver.match(/and(\d+)/i)
          if (verMatch && verMatch[1]) {
            const targetVer = verMatch[1]
            models = models.filter(m => {
              if (!m.android_version) return true // 无该字段时不过滤
              return String(m.android_version) === String(targetVer)
            })
          }
        }
      }
      break
    case 'local':
      models = localPhoneModels.value
      useNameAsId = true
      break
    case 'backup':
      models = backupPhoneModels.value
      useNameAsId = true
      break
    default:
      models = []
  }
  
  if (models && models.length > 0) {
    // 对于本地和备份机型，使用name作为id
    let processedModels = models
    if (useNameAsId) {
      processedModels = models.map(m => ({
        ...m,
        id: m.name // 覆盖或添加id为name
      }))
    }
    
    // 添加随机选项
    return [{ id: 'random', name: '随机机型' }, ...processedModels]
  }
  return []
})

// 处理机型类型切换
const handleSwitchModelTypeChange = async () => {
  tempModelId.value = '' // 重置选择
  if (!currentSwitchContainer.value) return
  
  const deviceIp = currentSwitchContainer.value.deviceIp
  
  if (switchModelType.value === 'local') {
     await getLocalPhoneModels(deviceIp)
  } else if (switchModelType.value === 'backup') {
     await fetchBackupModels(deviceIp)
  }
}

// 切换云机机型函数
const switchCloudMachineModel = async (device, containerName, modelId, modelName) => {
  if (!device || !containerName || !modelId) {
    console.error('切换机型参数不完整:', device, containerName, modelId, modelName)
    throw new Error('参数不能为空')
  }
  
  // 处理随机机型情况
  let finalModelId = modelId
  let finalModelName = modelName
  let modelType = 'online'
  
  // 检查 modelId 是否为对象 (包含 type 和 value)
  if (typeof modelId === 'object' && modelId !== null && modelId.value) {
    finalModelId = modelId.value
    modelType = modelId.type || 'online'
  }
  
  if (finalModelId === 'random') {
    if (phoneModels.value.length === 0 && modelType === 'online') {
       // 尝试获取一下? 或者报错
       // 如果是online且为空，可能还没获取
    }
    
    // 随机选择一个机型
    let list = []
    if (modelType === 'local') list = localPhoneModels.value
    else if (modelType === 'backup') list = backupPhoneModels.value
    else list = phoneModels.value
    
    if (list.length === 0) {
      console.error('随机机型选择失败：没有可用机型列表')
      throw new Error('随机机型选择失败：没有可用机型列表')
    }
    
    const randomIndex = Math.floor(Math.random() * list.length)
    const randomModel = list[randomIndex]
    
    if (modelType === 'online') {
      finalModelId = randomModel.id
      finalModelName = randomModel.name
    } else {
      finalModelId = randomModel.name
      finalModelName = randomModel.name
    }
    
    console.log(`随机选择的机型：${finalModelName} (${finalModelId})`)
  }
  
  // 确保 Online 机型使用的是 ID 而不是 Name
  // 有些情况下 modelId 可能是 name (如果 el-select 绑定的是 id 但数据源里 id=name)
  // 如果是 Online 且 modelId 与某个机型的 name 相同但 id 不同，尝试找到正确的 id
  if (modelType === 'online' && phoneModels.value.length > 0) {
    // 尝试在列表中查找
    const foundModel = phoneModels.value.find(m => m.id === finalModelId || m.name === finalModelId)
    if (foundModel) {
      // 如果找到了模型，确保使用其 ID
      // 注意：有些后端返回的数据可能 id 和 name 是一样的，这取决于后端实现
      // 但如果不一样，我们必须使用 id
      if (foundModel.id !== finalModelId) {
        console.log(`[switchCloudMachineModel] 修正 modelId: ${finalModelId} -> ${foundModel.id} (name: ${foundModel.name})`)
        finalModelId = foundModel.id
      }
      finalModelName = foundModel.name
    }
  }
  
  try {
    // 尝试使用已保存的密码
    const savedPassword = getDevicePassword(device.ip)
    let headers = {}
    
    if (savedPassword) {
      // 添加认证头
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    // 根据设备版本确定API端口和端点
    const port = device.version === 'v3' ? '8000' : '81'
    const apiEndpoint = device.version === 'v3' 
      ? `/android/switchModel` 
      : `/android/container/${containerName}/switchModel`
    
    // 准备请求数据
    const requestData = {
      name: containerName,
      modelId: ''
    }
    
    // V3设备需要添加更多参数
    if (device.version === 'v3') {
      Object.assign(requestData, {
        modelName: finalModelName,
        localModel: '',
        modelStatic: '',
        latitude: 0,
        longitude: 0,
        locateIp: '',
        locateQueryMethod: 'ip-api'
      })
      
      // 根据类型设置参数
      if (modelType === 'local') {
        requestData.localModel = finalModelId
      } else if (modelType === 'backup') {
        requestData.modelStatic = finalModelId
      } else {
        requestData.modelId = finalModelId
      }
    } else {
      // V0-V2 只有 modelId
      requestData.modelId = finalModelId
    }
    
    // 调用切换机型API
    const response = await axios.post(
      `http://${device.version === 'v3' ? getDeviceAddr(device.ip) : device.ip + ':' + port}${apiEndpoint}`,
      requestData,
      { headers: headers }
    )
    
    // 解析响应数据
    if (response.data.code === 0) {
      console.log(`云机 ${containerName} 切换机型成功: ${modelName}`)
      return true
    } else if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
      // 认证失败，显示认证对话框
      return new Promise((resolve, reject) => {
        showAuthDialog(device, async (password) => {
          try {
            const auth = btoa(`admin:${password}`)
            
            // 准备认证重试的请求数据
            const authRequestData = {
              name: containerName,
              modelId: finalModelId
            }
            
            // V3设备需要添加更多参数
            if (device.version === 'v3') {
              Object.assign(authRequestData, {
                modelName: finalModelName,
                localModel: '',
                modelStatic: '',
                latitude: 0,
                longitude: 0,
                locateIp: '',
                locateQueryMethod: 'ip-api'
              })
            }
            
            const authResponse = await axios.post(
              `http://${device.version === 'v3' ? getDeviceAddr(device.ip) : device.ip + ':' + port}${apiEndpoint}`,
              authRequestData,
              { 
                headers: {
                  'Authorization': `Basic ${auth}`
                }
              }
            )
            
            if (authResponse.data.code === 0) {
              console.log(`云机 ${containerName} 切换机型成功: ${modelName}`)
              resolve(true)
            } else {
              console.error(`切换机型失败: ${authResponse.data.message}`)
              reject(new Error(authResponse.data.message || '切换机型失败'))
            }
          } catch (error) {
            console.error(`切换云机机型失败: ${error.message}`)
            reject(error)
          }
        })
      })
    } else {
      console.error(`切换机型失败: ${response.data.message}`)
      throw new Error(response.data.message || '切换机型失败')
    }
  } catch (error) {
    console.error(`切换云机机型失败: ${error.message}`)
    throw error
  }
}

// 获取机型国家列表
const getCountryList = async (deviceIP) => {
  try {
    countryListLoading.value = true
    // 尝试使用已保存的密码
    const savedPassword = getDevicePassword(deviceIP)
    let headers = {}
    
    if (savedPassword) {
      // 添加认证头
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/android/countryCode`, {
      headers: headers
    })
    
    // 解析响应数据
    if (response.data.code === 0 && response.data.data) {
      const result = response.data.data
      countryList.value = result.list || []
      console.log('获取机型国家列表成功，共', countryList.value.length, '个国家')
    } else if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
      // 认证失败，显示认证对话框
      return new Promise((resolve, reject) => {
        showAuthDialog({ ip: deviceIP, version: 'v3' }, async (password) => {
          try {
            const auth = btoa(`admin:${password}`)
            const authResponse = await axios.get(`http://${getDeviceAddr(deviceIP)}/android/countryCode`, {
              headers: {
                'Authorization': `Basic ${auth}`
              }
            })
            
            if (authResponse.data.code === 0 && authResponse.data.data) {
              const result = authResponse.data.data
              countryList.value = result.list || []
              console.log('获取机型国家列表成功，共', countryList.value.length, '个国家')
              resolve(countryList.value)
            } else {
              ElMessage.error('获取机型国家列表失败: ' + (authResponse.data.message || '未知错误'))
              reject(new Error('获取机型国家列表失败'))
            }
          } catch (error) {
            console.error('获取机型国家列表失败:', error)
            ElMessage.error('获取机型国家列表失败: ' + error.message)
            reject(error)
          } finally {
            countryListLoading.value = false
          }
        })
      })
    } else {
      console.error('获取机型国家列表失败:', response.data.message)
      ElMessage.error('获取机型国家列表失败: ' + (response.data.message || '未知错误'))
    }
  } catch (error) {
    console.error('获取机型国家列表失败:', error)
    ElMessage.error('获取机型国家列表失败: ' + error.message)
  } finally {
    countryListLoading.value = false
  }
}

// 获取网络分组列表
const fetchVpcGroupList = async (deviceIP) => {
  try {
    const savedPassword = getDevicePassword(deviceIP)
    let headers = {}
    
    if (savedPassword) {
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/mytVpc/group`, {
      headers: headers
    })
    
    if (response.data.code === 0) {
      vpcGroupList.value = response.data.data?.list || []
    } else {
      console.error('获取分组列表失败:', response.data.message)
    }
  } catch (error) {
    console.error('获取分组列表失败:', error)
  }
}

// 处理分组选择变化
const handleVpcGroupChange = (groupId) => {
  createForm.value.vpcNodeId = ''
  createForm.value.vpcSelectMode = 'specified'
  updateImageForm.value.vpcNodeId = ''
  updateImageForm.value.vpcSelectMode = 'specified'
  
  if (!groupId) {
    vpcNodeList.value = []
    return
  }
  
  const selectedGroup = vpcGroupList.value.find(g => g.id === groupId)
  if (selectedGroup?.vpcs?.list) {
    vpcNodeList.value = selectedGroup.vpcs.list
  } else {
    vpcNodeList.value = []
  }
}

// 获取网卡列表
const fetchNetworkCards = async (ip) => {
  const deviceIP = typeof ip === 'string' ? ip : (createDevice.value ? createDevice.value.ip : null)
  if (!deviceIP) return
  // 根据创建类型决定使用哪个网卡类型字段
  const type = createForm.value.createType === 'container' 
    ? createForm.value.containerNetworkCardType 
    : createForm.value.networkCardType
  
  fetchingNetworkCards.value = true
  networkCardList.value = []
  hasMacVlan.value = false // 重置macVlan状态
  currentDeviceMacVlanInfo.value = { subnet: '', gw: '' } // 重置MacVlan信息
  
  try {
    const savedPassword = getDevicePassword(deviceIP)
    let headers = {}
    
    if (savedPassword) {
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    if (type === 'private') {
      const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/mytBridge`, {
        headers: headers
      })
      
      if (response.data.code === 0) {
        networkCardList.value = (response.data.data?.list || []).map(item => ({
          label: item.name,
          value: item.name
        }))
      } else {
        console.error('获取私有网卡列表失败:', response.data.message)
        networkCardList.value = []
      }
      // 若当前已选网卡不在列表中（包括列表为空），则清空选择
      const validValues = networkCardList.value.map(c => c.value)
      if (updateImageForm.value.mytBridgeName && !validValues.includes(updateImageForm.value.mytBridgeName)) {
        updateImageForm.value.mytBridgeName = ''
      }
    } else {
      const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/macvlan`, {
        headers: headers
      })
      
      if (response.data.code === 0) {
        const info = response.data.data
        
        // 检查macVlan字段
        if (info.macVlan !== null && info.macVlan !== undefined) {
          hasMacVlan.value = true
          
          // 解析MacVlan信息
          try {
            let macVlanData = info.macVlan
            if (typeof macVlanData === 'string') {
              macVlanData = JSON.parse(macVlanData)
            }
            
            // 提取subnet和gateway信息
            if (macVlanData && macVlanData.IPAM && Array.isArray(macVlanData.IPAM.Config) && macVlanData.IPAM.Config.length > 0) {
              currentDeviceMacVlanInfo.value.subnet = macVlanData.IPAM.Config[0].Subnet || ''
              currentDeviceMacVlanInfo.value.gw = macVlanData.IPAM.Config[0].Gateway || ''
            }
          } catch (parseError) {
            console.error('解析MacVlan信息失败:', parseError)
          }
        } else {
          hasMacVlan.value = false
        }
        
        const list = []
        if (info.netWork_eth0) {
          list.push({
            label: 'ETH0',
            value: 'eth0'
          })
        }
        if (info.network4g) {
          list.push({
            label: '4G',
            value: '4g'
          })
        }
        networkCardList.value = list
      } else {
        console.error('获取公有网卡信息失败:', response.data.message)
      }
    }
  } catch (error) {
    console.error('获取网卡列表失败:', error)
  } finally {
    fetchingNetworkCards.value = false
  }
}

// 获取网卡列表（更新镜像专用）
const fetchNetworkCardsForUpdate = async (deviceIP) => {
  if (!deviceIP) return
  
  const type = updateImageForm.value.networkCardType
  
  fetchingNetworkCards.value = true
  networkCardList.value = []
  hasMacVlan.value = false // 重置macVlan状态
  currentDeviceMacVlanInfo.value = { subnet: '', gw: '' } // 重置MacVlan信息
  
  try {
    const savedPassword = getDevicePassword(deviceIP)
    let headers = {}
    
    if (savedPassword) {
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }
    
    if (type === 'private') {
      const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/mytBridge`, {
        headers: headers
      })
      
      if (response.data.code === 0) {
        networkCardList.value = (response.data.data?.list || []).map(item => ({
          label: item.name,
          value: item.name
        }))
      } else {
        console.error('获取私有网卡列表失败:', response.data.message)
        networkCardList.value = []
      }
      // 若当前已选网卡不在列表中（包括列表为空），则清空选择
      const validValues = networkCardList.value.map(c => c.value)
      if (updateImageForm.value.mytBridgeName && !validValues.includes(updateImageForm.value.mytBridgeName)) {
        updateImageForm.value.mytBridgeName = ''
      }
    } else {
      const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/macvlan`, {
        headers: headers
      })
      
      if (response.data.code === 0) {
        const info = response.data.data
        
        // 检查macVlan字段
        if (info.macVlan !== null && info.macVlan !== undefined) {
          hasMacVlan.value = true
          
          // 解析MacVlan信息
          try {
            let macVlanData = info.macVlan
            if (typeof macVlanData === 'string') {
              macVlanData = JSON.parse(macVlanData)
            }
            
            // 提取subnet和gateway信息
            if (macVlanData && macVlanData.IPAM && Array.isArray(macVlanData.IPAM.Config) && macVlanData.IPAM.Config.length > 0) {
              currentDeviceMacVlanInfo.value.subnet = macVlanData.IPAM.Config[0].Subnet || ''
              currentDeviceMacVlanInfo.value.gw = macVlanData.IPAM.Config[0].Gateway || ''
            }
          } catch (parseError) {
            console.error('解析MacVlan信息失败:', parseError)
          }
        } else {
          hasMacVlan.value = false
        }
        
        const list = []
        if (info.netWork_eth0) {
          list.push({
            label: 'ETH0',
            value: 'eth0'
          })
        }
        if (info.network4g) {
          list.push({
            label: '4G',
            value: '4g'
          })
        }
        networkCardList.value = list
      } else {
        console.error('获取公有网卡信息失败:', response.data.message)
      }
    }
  } catch (error) {
    console.error('获取网卡列表失败:', error)
  } finally {
    fetchingNetworkCards.value = false
  }
}

// 处理网卡类型变化
const handleNetworkCardTypeChange = () => {
  if (createForm.value.networkCardType === 'public') {
    const version = createDeviceApiVersionNumber.value
    if (!version || version < 65) {
      ElMessage.warning('公有网卡需要SDK版本>65')
      createForm.value.networkCardType = 'private'
      createForm.value.mytBridgeName = ''
      createForm.value.macVlanIp = ''
      return
    }
    createForm.value.vpcGroupId = ''
    createForm.value.vpcNodeId = ''
  }
  createForm.value.mytBridgeName = ''
  createForm.value.macVlanIp = ''
  fetchNetworkCards()
}

// 处理容器模式网卡类型变化
const handleContainerNetworkCardTypeChange = () => {
  if (createForm.value.containerNetworkCardType === 'public') {
    const version = createDeviceApiVersionNumber.value
    if (!version || version < 65) {
      ElMessage.warning('公有网卡需要SDK版本>65')
      createForm.value.containerNetworkCardType = 'private'
      createForm.value.containerMytBridgeName = ''
      createForm.value.containerMacVlanIp = ''
      return
    }
    createForm.value.vpcGroupId = ''
    createForm.value.vpcNodeId = ''
    createForm.value.containerMytBridgeName = '' // 切换到公有网卡时清空私有网卡选择
  } else {
    createForm.value.containerMacVlanIp = '' // 切换到私有网卡时清空MacVlan IP
  }
  fetchNetworkCards()
}

// 处理更新镜像网卡类型变化
const handleUpdateNetworkCardTypeChange = () => {
  // 获取当前操作的设备IP
  const container = updateImageContainer.value
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp }
  } else if (!targetDevice && createDevice.value) {
    targetDevice = createDevice.value
  }

  if (updateImageForm.value.networkCardType === 'public') {
    let version = null
    if (targetDevice && targetDevice.id) {
      const cached = deviceVersionInfo.value.get(targetDevice.id)
      if (cached?.currentVersion) {
        version = parseFloat(cached.currentVersion)
      }
    }

    if (!version || version < 65) {
      ElMessage.warning('公有网卡需要SDK版本>65')
      updateImageForm.value.networkCardType = 'private'
      updateImageForm.value.mytBridgeName = ''
      updateImageForm.value.macVlanIp = ''
      return
    }
    updateImageForm.value.vpcGroupId = ''
    updateImageForm.value.vpcNodeId = ''
  }

  updateImageForm.value.mytBridgeName = ''
  updateImageForm.value.macVlanIp = ''
  
  if (targetDevice && targetDevice.ip) {
    fetchNetworkCardsForUpdate(targetDevice.ip)
  }
}

// 获取随机节点ID
const getRandomVpcNodeId = () => {
  if (!vpcNodeList.value || vpcNodeList.value.length === 0) {
    return null
  }
  
  const randomIndex = Math.floor(Math.random() * vpcNodeList.value.length)
  return vpcNodeList.value[randomIndex].id
}

// 提取节点显示名称
const extractNodeDisplayName = (remarks) => {
  if (!remarks) return ''
  const parts = remarks.split('_')
  return parts.length > 0 ? parts[parts.length - 1] : remarks
}

// 从本地存储读取镜像列表
const getImageListFromLocal = () => {
  try {
    const saved = localStorage.getItem(IMAGE_CACHE_KEY)
    if (saved) {
      return JSON.parse(saved)
    }
  } catch (error) {
    console.error('读取本地镜像列表缓存失败:', error)
  }
  return []
}

// 将镜像列表保存到本地存储
const saveImageListToLocal = (data) => {
  try {
    localStorage.setItem(IMAGE_CACHE_KEY, JSON.stringify(data))
    localStorage.setItem(IMAGE_CACHE_LAST_UPDATE_KEY, Date.now().toString())
  } catch (error) {
    console.error('保存镜像列表到本地缓存失败:', error)
  }
}

// 检查是否需要更新镜像列表
const shouldUpdateImageList = () => {
  try {
    const lastUpdate = localStorage.getItem(IMAGE_CACHE_LAST_UPDATE_KEY)
    if (!lastUpdate) {
      return true // 没有上次更新时间，需要更新
    }
    const now = Date.now()
    return now - parseInt(lastUpdate) > IMAGE_CACHE_DURATION
  } catch (error) {
    console.error('检查是否需要更新镜像列表失败:', error)
    return true // 出错时需要更新
  }
}

// 获取镜像列表，使用本地缓存，自动更新
const fetchImageList = async (deviceType) => {
  try {
    fetchingImages.value = true
    
    // 从本地存储加载镜像列表
    let localImages = getImageListFromLocal()
    
    // 首先使用本地缓存的数据
    if (localImages.length > 0) {
      // console.log('使用本地缓存的镜像列表，共', localImages.length, '个镜像')
      imageList.value = localImages
      filterImageList(deviceType)
      await categorizeOnlineImages() // 按型号分类在线镜像
    } else {
      // 本地缓存为空，使用默认镜像
      // console.log('本地缓存为空，使用默认镜像列表')
      imageList.value = [
        { name: 'registry.magicloud.tech/magicloud/dobox-android13:Q1', url: 'registry.magicloud.tech/magicloud/dobox-android13:Q1' }
      ];
      filterImageList(deviceType);
      await categorizeOnlineImages() // 按型号分类在线镜像
    }
    
    // 无论是否需要更新，都尝试联网更新本地缓存（后台自动更新）
    try {
      console.log('尝试联网更新镜像列表...')
      
      // 使用Wails IPC调用后端GetMirrorList函数获取镜像列表
      const response = await GetMirrorList()
      
      // console.log('获取镜像列表成功，返回数据:', response)
      
      // 处理响应数据
      if (response.code === '200' && response.data && Array.isArray(response.data)) {
        // 对数据按ID降序排序
        const sortedData = [...response.data].sort((a, b) => {
          const idA = parseInt(a.id) || 0
          const idB = parseInt(b.id) || 0
          return idB - idA
        })
        
        // 保存到本地存储
        saveImageListToLocal(sortedData)
        
        // 更新镜像列表
        imageList.value = sortedData
        // console.log('更新镜像列表成功，共', imageList.value.length, '个镜像')
        
        // 过滤匹配当前设备类型的镜像
        filterImageList(deviceType)
        // 按型号分类在线镜像
        await categorizeOnlineImages()
      } else {
        console.error('获取镜像列表失败: 无效的返回数据');
        ElMessage.warning('无法联网获取最新镜像列表，使用本地缓存！')
      }
    } catch (error) {
      console.error('更新镜像列表失败:', error)
      ElMessage.warning('无法联网获取最新镜像，使用本地缓存老镜像列表！')
    }
  } catch (error) {
    console.error('获取镜像列表失败:', error)
    ElMessage.error('获取镜像列表失败: ' + error.message)
  } finally {
    fetchingImages.value = false
  }
}

// 获取兼容的设备类型列表，参考api/main.go中的getCompatibleTypes实现
const getCompatibleTypes = (deviceType) => {
  switch (deviceType) {
    case 'q1_10':
    case 'c1_10':
      return ['q1_10', 'c1_10']
    case 'q1':
    case 'C1':
    case '': // 空值处理为C1
      return ['q1', 'C1']
    case 'r1_v3':
      return ['r1_v3', 'q1_v3', 'c1_v3']
    case 'c1_v3':
      return ['c1_v3', 'q1_v3', 'r1_v3']
    default:
      return [deviceType]
  }
}

// 过滤镜像列表，根据ttype匹配设备型号_版本
const filterImageList = (deviceType) => {
  if (!deviceType) {
    filteredImageList.value = imageList.value
    console.log('设备类型为空，返回所有镜像，共', filteredImageList.value.length, '个')
    return
  }
  
  // console.log('正在过滤镜像列表，设备类型:', deviceType)
  
  // 获取兼容的设备类型列表
  const compatibleTypes = getCompatibleTypes(deviceType)
  // console.log('兼容的设备类型:', compatibleTypes)
  
  // 过滤镜像列表，检查ttype和ttype2字段是否包含设备型号_版本
  filteredImageList.value = imageList.value.filter(image => {

  if(image.sys_ver != 5) {
    return false
   } 
  //  console.log('正在检查镜像:', image)
    // 检查ttype字段
    if (image.ttype && compatibleTypes.includes(image.ttype)) {
      return true
    }
    
    // 检查ttype2字段（数组）
    if (Array.isArray(image.ttype2)) {
      for (const t of image.ttype2) {
        if (compatibleTypes.includes(t)) {
          return true
        }
      }
    }
    
    // 没有匹配的ttype或ttype2字段，过滤掉
    return false
  })
  
  // console.log('过滤镜像列表成功，共', filteredImageList.value.length, '个镜像')
  
  // 如果过滤后没有镜像，添加默认镜像
  if (filteredImageList.value.length === 0) {
    // console.log('过滤后没有匹配的镜像，添加默认镜像')
    filteredImageList.value = [
      { name: 'registry.magicloud.tech/magicloud/dobox-android13:Q1', url: 'registry.magicloud.tech/magicloud/dobox-android13:Q1' }
    ];
  }

  // 默认选中留给 watcher 处理（会经过安卓版本过滤）
}

// 监听 filteredImageList 变化，自动选中第一条（按安卓版本过滤后）
watch(filteredImageList, () => {
  if (createForm.value.imageCategory === 'online' && createForm.value.createType !== 'container') {
    const filtered = androidVersionFilteredImageList.value
    if (filtered && filtered.length > 0) {
      createForm.value.imageSelect = filtered[0].url
    } else {
      createForm.value.imageSelect = ''
    }
  }
}, { immediate: true })

// 按型号分类在线镜像
const categorizeOnlineImages = async () => {
  const imagesByModel = new Map()
  let firstModel = null // 记录第一个出现的型号
  
  imageList.value.forEach(image => {
    // 检查 ttype2 数组，如果是 Q1 或 P1 则添加到对应分类
    if (Array.isArray(image.ttype2)) {
      // 按优先级处理：P1优先于Q1（按定义顺序）
      const models = []
      if (image.ttype2.includes('p1_v3')) {
        models.push('P1')
      }
      if (image.ttype2.includes('q1_v3')) {
        models.push('Q1')
      }
      if (image.ttype2.includes('r1_v3')) {
        models.push('R1')
      }
      if (image.ttype2.includes('c1_v3')) {
        models.push('C1')
      }
      
      // 处理该镜像属于的所有型号
      models.forEach(displayModel => {
        if (!imagesByModel.has(displayModel)) {
          imagesByModel.set(displayModel, [])
          // 记录第一个出现的型号
          if (!firstModel) {
            firstModel = displayModel
          }
        }
        imagesByModel.get(displayModel).push(image)
      })
    }
  })
  
  onlineImagesByModel.value = imagesByModel
  // console.log('按型号分类在线镜像成功，共', imagesByModel.size, '个型号')
  
  // 自动选中第一个标签（使用记录的第一个型号）
  if (imagesByModel.size > 0 && !currentOnlineImageModel.value) {
    currentOnlineImageModel.value = firstModel || Array.from(imagesByModel.keys())[0]
    console.log('自动选中第一个型号:', currentOnlineImageModel.value)
  }
  
  // 检查每个在线镜像的下载状态
  await checkAllImagesDownloadStatus()
}

// 获取本地缓存镜像列表
const fetchLocalCachedImages = async () => {
  try {
    isLoadingLocalImages.value = true
    // console.log('获取本地缓存镜像列表')
    
    // 使用Wails的原生通信方式，获取本地镜像列表
    const localImages = await GetLocalImages()
    // console.log('获取本地缓存镜像列表成功，返回数据:', localImages)
    
    // 处理返回的本地镜像数据
    let processedImages = await Promise.all(localImages.map(async image => {
      // 移除available_models中的重复型号
      const uniqueModels = Array.from(new Set(image.available_models || []))
      
      // 获取基础镜像名称（去掉.tar.gz后缀）
      const baseName = image.name.replace('.tar.gz', '')
      
      // 优先使用从JSON文件中获取的镜像名称
      let displayName = image.image_name || baseName
      
      // 检查是否是特别版镜像
      if (baseName.includes('特别版')) {
        // 特别版镜像处理：从文件路径中提取设备型号标识
        // 例如：从 "dobox-P14_v3_all_202601091434.tar.gz" 中提取 "P14_v3"
        const fileName = image.name
        const modelMatch = fileName.match(/dobox-(\w+_v\d+)_all_/)
        if (modelMatch && modelMatch[1]) {
          // 添加设备型号到特别版镜像名称中
          displayName = `${baseName} (${modelMatch[1]})`
        } else {
          // 如果无法提取型号，使用原始名称
          displayName = baseName
        }
      } else {
        // 非特别版镜像尝试匹配在线镜像列表中的名称
        // 遍历在线镜像列表，查找匹配的镜像
        for (const onlineImage of imageList.value) {
          // 从在线镜像URL中提取关键信息用于匹配
          const onlineImageUrl = onlineImage.url
          
          // 提取在线镜像的关键标识（支持多种URL格式）
          let onlineImageKey = onlineImageUrl
          if (onlineImageUrl.includes('/')) {
            onlineImageKey = onlineImageUrl.split('/').pop()
          }
          
          // 提取镜像的基础名称（不含标签）
          const onlineImageBase = onlineImageKey.split(':')[0]
          
          // 提取镜像的完整标识（包含标签）
          const onlineImageFull = onlineImageKey
          // console.log('onlineImageFull', onlineImageFull)
          
          // 三种匹配方式，优先级递减：
          // 1. 本地镜像名称包含在线镜像的完整标识
          // 2. 本地镜像名称包含在线镜像的基础名称
          // 3. 在线镜像URL包含本地镜像名称的关键部分
          // if (baseName.includes(onlineImageFull) || 
          //     baseName.includes(onlineImageBase) ||
          //     onlineImageUrl.includes(baseName)) {
          //   // 如果找到匹配项，使用在线镜像的名称
          //   displayName = onlineImage.name
          //   break
          // }
        }
      }

      console.log('baseName:', baseName, 'displayName:', displayName, 'image_name:', image.image_name)
      
      return {
        name: displayName, // 优先使用JSON文件中的名称，其次是在线镜像名称，否则使用文件名
        originalName: baseName, // 保存原始文件名，用于参考
        path: image.path, // 镜像文件路径（保留原始path属性，用于删除操作）
        url: image.path, // 镜像文件路径（兼容原有代码）
        size: image.size, // 镜像大小
        createTime: new Date(image.createTime).toLocaleString(), // 创建时间
        availableModels: uniqueModels, // 去重后的可用设备型号
        onlineUrl: image.online_url || '' // 从JSON metadata读取的在线镜像地址
      }
    }))
    
    localCachedImages.value = processedImages
    
    console.log('获取本地缓存镜像列表成功，共', localCachedImages.value.length, '个镜像')
  } catch (error) {
    console.error('获取本地缓存镜像列表失败:', error)
    ElMessage.error('获取本地缓存镜像列表失败: ' + error.message)
    // 失败时使用空数组，不显示模拟数据
    localCachedImages.value = []
  } finally {
    isLoadingLocalImages.value = false
  }
}

// 处理原始设备镜像数据
const processRawDeviceImages = (deviceImages) => {
    if (Array.isArray(deviceImages)) {
      return deviceImages
        .map(image => {
        // 处理RepoTags数组，获取第一个标签作为镜像名称
        let imageName = '未知镜像'
        let imageUrl = ''
        if (Array.isArray(image.imageUrl) && image.imageUrl.length > 0) {
          imageName = image.imageUrl[0]
          imageUrl = image.imageUrl[0]
        } else if (Array.isArray(image.Image) && image.Image.length > 0) {
          imageName = image.Image[0]
          imageUrl = image.Image[0]
        } else if (image.imageUrl) {
          imageName = image.imageUrl
          imageUrl = image.imageUrl
        } else if (image.Image) {
          imageName = image.Image
          imageUrl = image.Image
        } else if (image.id || image.Id) {
          // 如果所有名称字段都为空，使用镜像ID的前12位作为名称
          const imageId = image.id || image.Id
          imageName = `镜像-${imageId.substring(0, 12)}`
          imageUrl = imageId
        }
        
        // 处理大小字段，转换为人类可读的格式
        let size = '未知大小'
        if (image.size) {
          if (typeof image.size === 'number') {
            size = formatFileSize(image.size)
          } else {
            size = image.size
          }
        }
        
        // 处理创建时间字段
        let createTime = '未知时间'
        if (image.createTime) {
          if (typeof image.createTime === 'number') {
            // 检查是否是秒级时间戳（Docker API返回的是秒）
            if (image.createTime < 10000000000) {
              // 秒级时间戳，转换为毫秒
              createTime = new Date(image.createTime * 1000).toLocaleString()
            } else {
              // 毫秒级时间戳
              createTime = new Date(image.createTime).toLocaleString()
            }
          } else {
            createTime = image.createTime
          }
        }
        
        return {
          name: imageName,
          url: imageUrl,
          size: size,
          createTime: createTime,
          original: image
        }
      })
      // 过滤掉没有有效名称和URL的镜像
      .filter(image => {
        // 确保name和url是字符串类型，然后再调用trim()方法
        const validName = typeof image.name === 'string' && image.name.trim() !== ''
        const validUrl = typeof image.url === 'string' && image.url.trim() !== ''
        return validName && validUrl
      })
    } else {
      return []
    }
}

// 获取盒子镜像列表（设备上存在的镜像）
const fetchBoxImages = async () => {
  if (!activeDevice.value) {
    boxImages.value = []
    return
  }
  
  try {
    isLoadingBoxImages.value = true
    console.log('获取盒子镜像列表，设备:', activeDevice.value.ip)
    
    // 使用Wails的原生通信方式，获取设备上的镜像列表
    // 从本地存储获取密码
    const savedPassword = getDevicePassword(activeDevice.value.ip);
    const deviceImages = await GetImages(activeDevice.value.ip, activeDevice.value.version, savedPassword || '')
    console.log('获取盒子镜像列表成功，返回数据:', deviceImages)
    
    // 处理返回的设备镜像数据
    boxImages.value = processRawDeviceImages(deviceImages)
    
    console.log('获取盒子镜像列表成功，共', boxImages.value.length, '个镜像')
    
    // 检查在线镜像的上传状态：如果盒子中有这个在线镜像，就说明是已上传的
    await checkOnlineImagesUploadStatus()
  } catch (error) {
    console.error('获取盒子镜像列表失败:', error)
    ElMessage.error('获取盒子镜像列表失败: ' + error.message)
    boxImages.value = []
  } finally {
    isLoadingBoxImages.value = false
  }
}

// 检查在线镜像的上传状态：如果盒子中有这个在线镜像，就说明是已上传的
const checkOnlineImagesUploadStatus = async () => {
  try {
    console.log('开始检查在线镜像的上传状态')
    
    // 清空当前上传状态
    imageUploadStatus.value.clear()
    
    // 遍历所有在线镜像
    for (const [model, images] of onlineImagesByModel.value.entries()) {
      for (const image of images) {
        // 只检查已下载的镜像
        if (imageDownloadStatus.value.get(image.url)) {
          // 检查盒子中是否有这个镜像
          const isUploaded = await checkImageInBox(image)
          imageUploadStatus.value.set(image.url, isUploaded)
        }
      }
    }
    
    console.log('检查在线镜像的上传状态完成')
  } catch (error) {
    console.error('检查在线镜像的上传状态失败:', error)
  }
}

// 检查在线镜像是否在盒子中
const checkImageInBox = async (image) => {
  try {
    // 从在线镜像URL中提取镜像名称
    const imageUrl = image.url
    console.log('检查镜像是否在盒子中:', imageUrl)
    
    // 提取镜像名称（去掉registry部分）
    let imageName = imageUrl
    if (imageUrl.includes('/')) {
      // 例如：registry.cn-guangzhou.aliyuncs.com/mytos/dobox:P14_v3_all_202512312124
      // 提取为：mytos/dobox:P14_v3_all_202512312124
      const parts = imageUrl.split('/')
      imageName = parts.slice(1).join('/')
    }
    
    console.log('提取的镜像名称:', imageName)
    
    // 遍历盒子镜像列表
    for (const boxImage of boxImages.value) {
      console.log('盒子镜像:', boxImage.name)
      // 检查盒子镜像名称是否包含在线镜像名称的关键部分
      if (boxImage.name.includes(imageName) || 
          boxImage.name.includes(imageName.replace(':', '_')) ||
          boxImage.name.includes(imageName.split('/').pop())) {
        console.log('镜像在盒子中找到:', imageName)
        return true
      }
    }
    
    console.log('镜像在盒子中未找到:', imageName)
    return false
  } catch (error) {
    console.error('检查镜像是否在盒子中失败:', error)
    return false
  }
}

// 下载在线镜像到本地
const downloadOnlineImage = async (image) => {
  try {
    // 如果已有下载任务正在进行，先完全清理旧状态
    if (isDownloadingImage.value || currentDownloadTaskId.value) {
      console.log('检测到旧的下载任务，彻底清理状态')
      const oldTaskId = currentDownloadTaskId.value
      
      // 立即清空所有状态
      isDownloadingImage.value = false
      currentDownloadImage.value = null
      currentDownloadTaskId.value = null
      downloadProgress.value = 0
      downloadStartTime.value = 0
      
      // 取消旧任务
      if (oldTaskId) {
        const oldTask = taskQueue.value.find(t => t.id === oldTaskId)
        if (oldTask && oldTask.status === 'running') {
          oldTask.status = 'canceled'
          oldTask.endTime = new Date()
          oldTask.progress = 0
        }
      }
      
      // 延长等待时间，确保后端旧任务的事件不会影响新任务
      // 这个延迟很重要，给后端足够的时间停止发送旧事件
      await new Promise(resolve => setTimeout(resolve, 1000))
    }
    
    // 生成新的下载会话时间戳
    const sessionStartTime = Date.now()
    downloadStartTime.value = sessionStartTime
    
    console.log('==================== 开始新的下载任务 ====================')
    console.log('镜像名称:', image.name)
    console.log('镜像URL:', image.url)
    console.log('会话时间戳:', sessionStartTime)
    
    // 添加下载任务到任务队列
    const taskId = addTaskToQueue('downloadImage', [{ imageUrl: image.url, imageName: image.name }], {
      imageName: image.name,
      imageUrl: image.url,
      sessionStartTime: sessionStartTime  // 保存会话时间戳到任务元数据
    })
    
    console.log('创建任务ID:', taskId)
    
    // 关键：先设置所有状态，再执行任务
    currentDownloadTaskId.value = taskId
    currentDownloadImage.value = image
    isDownloadingImage.value = true
    downloadProgress.value = 0
    
    console.log('当前下载任务ID已设置:', currentDownloadTaskId.value)
    console.log('当前下载镜像URL:', currentDownloadImage.value?.url)
    
    // 执行下载任务，设置状态为running
    executeTask(taskId)
    
    // 使用Wails的原生通信方式，调用后端下载镜像，传递完整的镜像元数据
    const result = await DownloadImage(image)
    console.log('下载镜像请求结果:', result)
    
    // 注意：这里不再需要轮询获取进度，进度会通过事件监听自动更新
    // 下载结果也会通过事件监听处理
  } catch (error) {
    console.error('下载镜像请求失败:', error)
    ElMessage.error(`下载镜像请求失败: ${error.message}`)
    
    // 出错时彻底清理状态
    isDownloadingImage.value = false
    currentDownloadImage.value = null
    currentDownloadTaskId.value = null
    downloadProgress.value = 0
    downloadStartTime.value = 0
  }
}

// 检查镜像是否已下载
const checkImageDownloadStatus = async (imageUrl) => {
  try {
    const result = await IsImageDownloaded(imageUrl)
    // console.log('检查镜像下载状态:', imageUrl, result)
    imageDownloadStatus.value.set(imageUrl, result.downloaded)
    return result.downloaded
  } catch (error) {
    console.error('检查镜像下载状态失败:', error)
    return false
  }
}

// 检查所有在线镜像的下载状态
const checkAllImagesDownloadStatus = async () => {
  try {
    console.log('开始检查所有在线镜像的下载状态')
    for (const [model, images] of onlineImagesByModel.value.entries()) {
      for (const image of images) {
        await checkImageDownloadStatus(image.url)
      }
    }
    console.log('检查所有在线镜像的下载状态完成')
  } catch (error) {
    console.error('检查所有在线镜像的下载状态失败:', error)
  }
}

// 检查镜像的上传状态
const checkImageUploadStatus = async (imageUrl) => {
  try {
    if (!activeDevice.value) {
      return false
    }
    
    // 获取本地镜像路径
    const result = await IsImageDownloaded(imageUrl)
    if (!result.downloaded || !result.local_path) {
      return false
    }
    
    const localPath = result.local_path
    const imageName = localPath.split('\\').pop().split('/').pop().replace('.tar.gz', '')
    
    // 获取设备上的镜像列表
    // 从本地存储获取密码
    const savedPassword = getDevicePassword(activeDevice.value.ip);
    const boxImages = await GetImages(activeDevice.value.ip, activeDevice.value.version, savedPassword || '')
    console.log('设备上的镜像列表:', boxImages)
    
    // 检查镜像是否在设备上
    let isUploaded = false
    if (Array.isArray(boxImages)) {
      for (const img of boxImages) {
        let imgName = ''
        if (Array.isArray(img.imageUrl) && img.imageUrl.length > 0) {
          imgName = img.imageUrl[0]
        } else if (Array.isArray(img.Image) && img.Image.length > 0) {
          imgName = img.Image[0]
        } else if (img.imageUrl) {
          imgName = img.imageUrl
        } else if (img.Image) {
          imgName = img.Image
        }
        
        // 检查镜像名称是否匹配
        if (imgName.includes(imageName)) {
          isUploaded = true
          break
        }
      }
    }
    
    // 更新上传状态
    imageUploadStatus.value.set(imageUrl, isUploaded)
    return isUploaded
  } catch (error) {
    console.error('检查镜像上传状态失败:', error)
    return false
  }
}

// 上传镜像到设备
// 设备选择对话框状态
const showDeviceSelectionDialog = ref(false)
const selectedDevicesForUpload = ref([])
const currentUploadingImage = ref(null)
const isUploadingToMultipleDevices = ref(false)

// 计算属性：与当前上传镜像兼容的设备列表
const compatibleDevicesForUpload = computed(() => {
  // 添加调试信息
  console.log('compatibleDevicesForUpload 计算属性执行中:')
  console.log('设备列表:', devices.value)
  console.log('设备列表类型:', typeof devices.value)
  console.log('设备列表长度:', Array.isArray(devices.value) ? devices.value.length : '不是数组')
  console.log('当前上传镜像:', currentUploadingImage.value)
  
  // 确保设备列表是数组
  if (!Array.isArray(devices.value)) {
    console.log('设备列表不是数组，返回空数组')
    return []
  }
  
  // 如果没有设备，直接返回空数组
  if (devices.value.length === 0) {
    console.log('没有设备，返回空数组')
    return []
  }
  
  // 如果没有选择镜像，显示所有设备
  if (!currentUploadingImage.value) {
    console.log('没有选择镜像，返回所有设备')
    return [...devices.value]
  }
  
  const image = currentUploadingImage.value
  let imageModels = []
  
  // 安全处理镜像的兼容性信息
  if (Array.isArray(image.availableModels)) {
    imageModels = image.availableModels
    console.log('镜像 availableModels:', image.availableModels)
  } else {
    // 处理在线镜像的兼容性信息
    if (typeof image.ttype === 'string' && image.ttype.trim()) {
      imageModels.push(image.ttype)
      console.log('添加ttype到imageModels:', image.ttype)
    }
    if (Array.isArray(image.ttype2)) {
      const validTypes = image.ttype2.filter(type => typeof type === 'string' && type.trim())
      imageModels = [...imageModels, ...validTypes]
      console.log('ttype2数组:', image.ttype2)
      console.log('ttype2中有效类型:', validTypes)
    }
    console.log('镜像 ttype:', image.ttype)
    console.log('镜像 ttype2:', image.ttype2)
    console.log('镜像支持的设备型号:', imageModels)
  }
  
  // 如果镜像没有指定可用型号，则显示所有设备
  if (imageModels.length === 0) {
    console.log('镜像没有指定可用型号，返回所有设备')
    return [...devices.value]
  }
  
  // 过滤与镜像兼容的设备
  const compatibleDevices = devices.value.filter(device => {
    // 安全获取设备型号
    const deviceType = typeof device === 'object' && device !== null ? device.name || '' : ''
    console.log('检查设备:', deviceType, '设备ID:', device?.id, '设备IP:', device?.ip)
    
    // 检查每个模型是否匹配
    const isCompatible = imageModels.some(model => {
      // 标准化模型名称：去除空格，替换下划线为破折号，转为小写
      const normalizeName = (name) => {
        return name?.trim().replace(/_/g, '-').toLowerCase() || ''
      }
      
      const normalizedModel = normalizeName(model)
      const normalizedDeviceType = normalizeName(deviceType)
      const matches = normalizedModel === normalizedDeviceType
      
      console.log('  检查模型:', model, '(标准化:', normalizedModel, ') 与设备型号:', deviceType, '(标准化:', normalizedDeviceType, ') 匹配:', matches)
      return matches
    })
    
    console.log('  设备', deviceType, '兼容:', isCompatible)
    return isCompatible
  })
  
  console.log('兼容设备列表:', compatibleDevices)
  console.log('兼容设备列表长度:', compatibleDevices.length)
  
  // 将结果转换为普通数组，并且将数组中的每个设备对象也转换为普通对象，确保UI能够正确识别
  const resultArray = compatibleDevices.map(device => {
    // 将设备对象转换为普通对象
    return {
      ...device,
      id: device.id,
      name: device.name,
      ip: device.ip,
      port: device.port,
      version: device.version
    }
  })
  
  console.log('转换为普通数组和普通对象后的结果:', resultArray)
  console.log('转换后数组长度:', resultArray.length)
  
  return resultArray
})

const uploadImageToDevice = async (image) => {
  try {
    // 检查镜像是否已下载
    const downloadStatus = await checkImageDownloadStatus(image.url)
    if (!downloadStatus) {
      ElMessage.error('镜像未下载到本地，请先下载')
      return
    }
    
    // 获取本地镜像路径
    const result = await IsImageDownloaded(image.url)
    if (!result.downloaded || !result.local_path) {
      ElMessage.error('无法获取本地镜像路径')
      return
    }
    
    // 显示设备选择对话框,设置localPath与本地镜像一致
    currentUploadingImage.value = {
      ...image,
      localPath: result.local_path
    }
    selectedDevicesForUpload.value = []
    showDeviceSelectionDialog.value = true
  } catch (error) {
    console.error('上传镜像到设备失败:', error)
    ElMessage.error('操作失败')
  }
}

// 设备选择表格的引用
const deviceSelectionTableRef = ref(null)

// 刷新上传设备列表的状态
const refreshingDevicesForUpload = ref(false)

// 刷新上传设备列表
const refreshDeviceListForUpload = async () => {
  try {
    refreshingDevicesForUpload.value = true
    // 等待一下，确保获取最新状态
    await new Promise(resolve => setTimeout(resolve, 500))
    
    // 📊 诊断信息
    console.log('========== 设备列表诊断 ==========')
    console.log('总设备数:', devices.value.length)
    console.log('兼容设备数:', compatibleDevicesForUpload.value.length)
    console.log('在线兼容设备数:', sortedCompatibleDevicesList.value.length)
    
    // 统计各状态设备数量
    const statusCount = { online: 0, offline: 0, unknown: 0 }
    devices.value.forEach(device => {
      const status = devicesStatusCache.value.get(device.id)
      if (status === 'online') statusCount.online++
      else if (status === 'offline') statusCount.offline++
      else statusCount.unknown++
    })
    console.log('状态统计:', statusCount)
    
    // 列出离线但 isOnline=true 的设备（数据不一致）
    const inconsistentDevices = devices.value.filter(device => {
      const status = devicesStatusCache.value.get(device.id)
      return device.isOnline && status !== 'online'
    })
    if (inconsistentDevices.length > 0) {
      console.warn('⚠️ 发现状态不一致的设备:', inconsistentDevices.map(d => `${d.ip} (isOnline=true, status=${devicesStatusCache.value.get(d.id)})`))
    }
    console.log('==================================')
    
    ElMessage.success(`已刷新设备列表，当前在线设备: ${sortedCompatibleDevicesList.value.length} 台`)
  } catch (error) {
    console.error('刷新设备列表失败:', error)
    ElMessage.error('刷新设备列表失败: ' + error.message)
  } finally {
    refreshingDevicesForUpload.value = false
  }
}

// 处理上传设备选择变化
const handleUploadDeviceSelectionChange = (selection) => {
  selectedDevicesForUpload.value = selection
}

// 检查设备是否可选择（仅在线设备可选）
const checkDeviceSelectable = (row) => {
  return devicesStatusCache.value.get(row.id) === 'online'
}

// 获取设备行的class名称（离线设备灰色）
const getDeviceRowClassName = ({ row }) => {
  return devicesStatusCache.value.get(row.id) !== 'online' ? 'device-offline-row' : ''
}

// 获取设备存储信息（用于上传设备选择表格）
const getDeviceStorageInfo = (deviceId) => {
  const firmwareInfo = deviceFirmwareInfo.value.get(deviceId)
  if (!firmwareInfo?.originalData) {
    return null
  }
  
  const total = Number(firmwareInfo.originalData.mmctotal) || 0
  const used = Number(firmwareInfo.originalData.mmcuse) || 0
  
  if (total === 0) {
    return null
  }
  
  const free = total - used
  const freeGb = free / 1024
  const freeText = freeGb >= 1 ? `${freeGb.toFixed(1)} GB` : `${free.toFixed(0)} MB`
  
  return {
    total,
    used,
    free,
    freeGb,
    freeText,
    isLow: freeGb < 10  // 可用空间小于10GB视为不足
  }
}

// 处理设备选择对话框关闭
const handleDeviceSelectionDialogClose = () => {
  // 清除选择状态
  selectedDevicesForUpload.value = []
  // 清除表格的选择状态
  if (deviceSelectionTableRef.value) {
    deviceSelectionTableRef.value.clearSelection()
  }
}

// 处理设备选择后上传镜像
const handleUploadAfterDeviceSelection = async () => {
  if (selectedDevicesForUpload.value.length === 0) {
    ElMessage.error('请选择至少一个设备')
    return
  }
  
  if (!currentUploadingImage.value) {
    ElMessage.error('请选择要上传的镜像')
    return
  }
  
  // 上传前再次检查设备在线状态
  const offlineDevices = selectedDevicesForUpload.value.filter(device => {
    const status = devicesStatusCache.value.get(device.id)
    return status !== 'online' && !device.isOnline
  })
  
  if (offlineDevices.length > 0) {
    const offlineIPs = offlineDevices.map(d => d.ip).join(', ')
    ElMessage.error(`以下设备已离线，无法上传: ${offlineIPs}`)
    return
  }
  
  // 创建上传镜像任务
  const taskId = addTaskToQueue('uploadImage', selectedDevicesForUpload.value, {
    imageName: currentUploadingImage.value.name,
    imagePath: currentUploadingImage.value.localPath || currentUploadingImage.value.url,
    imageSize: currentUploadingImage.value.size
  })
  
  try {
    showDeviceSelectionDialog.value = false
    isUploadingToMultipleDevices.value = true
    isUploadingImage.value = true
    currentUploadImage.value = currentUploadingImage.value
    uploadProgress.value = 0
    
    // 更新任务状态为运行中
    const task = taskQueue.value.find(t => t.id === taskId)
    if (task) {
      task.status = 'running'
      task.startTime = new Date()
    }
    
    console.log('开始批量上传镜像:', currentUploadingImage.value.name, '到设备:', selectedDevicesForUpload.value.map(d => d.ip))
    
    // 获取本地镜像路径
    let localPath = ''
    if (currentUploadingImage.value.localPath) {
      // 本地镜像直接使用本地路径
      localPath = currentUploadingImage.value.localPath
    } else {
      // 在线镜像需要获取下载路径
      const result = await IsImageDownloaded(currentUploadingImage.value.url)
      if (!result.downloaded || !result.local_path) {
        ElMessage.error('无法获取本地镜像路径')
        isUploadingImage.value = false
        currentUploadImage.value = null
        uploadProgress.value = 0
        isUploadingToMultipleDevices.value = false
        return
      }
      localPath = result.local_path
      
      // 更新任务的本地镜像路径
      const task = taskQueue.value.find(t => t.id === taskId)
      if (task) {
        task.imagePath = localPath
      }
    }
    
    // 批量上传到所选设备，按设备分组，同一设备的任务串行执行
    // 策略：不同设备可以并行，同一设备的任务必须串行
    
    // 按设备IP分组
    const devicesByIP = {}
    selectedDevicesForUpload.value.forEach(device => {
      if (!devicesByIP[device.ip]) {
        devicesByIP[device.ip] = []
      }
      devicesByIP[device.ip].push(device)
    })
    
    const deviceIPs = Object.keys(devicesByIP)
    const results = []
    let completedCount = 0
    const totalCount = selectedDevicesForUpload.value.length
    
    // 使用信号量控制并发设备数（最多2个设备同时上传）
    const maxConcurrentDevices = 2
    const runningCount = { value: 0 }
    
    // 串行执行同一设备的任务
    const processDeviceSerially = async (deviceIP, devices) => {
      for (const device of devices) {
        // 在每个设备上传前都检查任务是否已取消
        const checkTask = taskQueue.value.find(t => t.id === taskId)
        if (checkTask && checkTask.status === 'canceled') {
          console.log('[上传镜像] 任务已取消,停止上传设备:', device.ip)
          // 将剩余设备标记为已取消,不上传
          results.push({ device: device.ip, success: false, message: '任务已取消' })
          completedCount++
          uploadProgress.value = Math.round((completedCount / totalCount) * 100)
          
          const task = taskQueue.value.find(t => t.id === taskId)
          if (task) {
            task.completed = completedCount
            task.progress = Math.round((completedCount / totalCount) * 100)
          }
          continue
        }
        
        try {
          const imageName = currentUploadingImage.value.name
          const version = device.version || 'v3'
          const password = getDevicePassword(device.ip)
          
          console.log('[上传镜像] 开始上传到设备:', device.ip)
          
          // 上传前再次检查任务状态
          const preCheckTask = taskQueue.value.find(t => t.id === taskId)
          if (preCheckTask && preCheckTask.status === 'canceled') {
            console.log('[上传镜像] 上传前检测到任务已取消,跳过设备:', device.ip)
            results.push({ device: device.ip, success: false, message: '任务已取消' })
            completedCount++
            uploadProgress.value = Math.round((completedCount / totalCount) * 100)
            
            const task = taskQueue.value.find(t => t.id === taskId)
            if (task) {
              task.completed = completedCount
              task.progress = Math.round((completedCount / totalCount) * 100)
            }
            continue
          }
          
          const loadResult = await LoadImageToDevice(device.ip, localPath, version, password || '')
          
          // 检查是否因为取消而失败
          if (loadResult.canceled) {
            console.log('[上传镜像] 检测到后端取消标志,停止后续上传')
            results.push({ device: device.ip, success: false, message: '任务已取消' })
            // 直接退出循环,不再处理后续设备
            break
          }
          
          // 上传完成后检查任务状态,如果已取消则不记录成功
          const postCheckTask = taskQueue.value.find(t => t.id === taskId)
          if (postCheckTask && postCheckTask.status === 'canceled') {
            console.log('[上传镜像] 上传完成后检测到任务已取消,不记录结果:', device.ip)
            results.push({ device: device.ip, success: false, message: '任务已取消' })
          } else if (loadResult.success) {
            results.push({ device: device.ip, success: true })
          } else {
            const errorMsg = loadResult.message || '未知错误'
            results.push({ device: device.ip, success: false, message: errorMsg })
          }
        } catch (error) {
          const errorMsg = error.message || error.toString() || '未知异常'
          results.push({ device: device.ip, success: false, message: errorMsg })
        } finally {
          completedCount++
          uploadProgress.value = Math.round((completedCount / totalCount) * 100)
          
          const task = taskQueue.value.find(t => t.id === taskId)
          if (task) {
            task.completed = completedCount
            task.progress = Math.round((completedCount / totalCount) * 100)
          }
          
          runningCount.value--
        }
      }
    }
    
    // 并发处理不同设备
    const devicePromises = []
    for (const deviceIP of deviceIPs) {
      // 检查任务是否已取消
      const task = taskQueue.value.find(t => t.id === taskId)
      if (task && task.status === 'canceled') {
        console.log('[上传镜像] 任务已取消,停止添加新设备:', taskId)
        break
      }
      
      // 等待有可用的并发槽位
      while (runningCount.value >= maxConcurrentDevices) {
        // 在等待期间也检查任务状态
        const task = taskQueue.value.find(t => t.id === taskId)
        if (task && task.status === 'canceled') {
          console.log('[上传镜像] 等待期间任务已取消:', taskId)
          break
        }
        await new Promise(resolve => setTimeout(resolve, 100))
      }
      
      // 再次检查任务是否已取消
      const taskFinal = taskQueue.value.find(t => t.id === taskId)
      if (taskFinal && taskFinal.status === 'canceled') {
        break
      }
      
      runningCount.value++
      devicePromises.push(processDeviceSerially(deviceIP, devicesByIP[deviceIP]))
    }
    
    await Promise.all(devicePromises)
    
    // 统计成功和失败
    const successCount = results.filter(r => r.success).length
    const failCount = results.filter(r => !r.success).length
    
    // 更新任务状态
    const resultTask = taskQueue.value.find(t => t.id === taskId)
    if (resultTask) {
      // 如果任务已被取消,保持取消状态
      if (resultTask.status === 'canceled') {
        console.log('[上传镜像] 任务已取消,不更新为完成状态')
      } else {
        resultTask.status = failCount > 0 ? 'failed' : 'completed'
        resultTask.endTime = new Date()
        resultTask.completed = successCount
        resultTask.failed = failCount
        resultTask.progress = 100
        
        // 记录失败的设备及错误信息
        resultTask.failedTargets = results.filter(r => !r.success).map(r => ({
          deviceIP: r.device,
          error: r.message || '未知错误'
        }))
      }
    }
    
    // 只在任务未取消时显示成功/失败消息
    const canceledTask = taskQueue.value.find(t => t.id === taskId)
    if (!canceledTask || canceledTask.status !== 'canceled') {
      if (successCount > 0) {
        ElMessage.success(`成功上传到 ${successCount} 个设备`)
      }
      
      if (failCount > 0) {
        const failDetails = results.filter(r => !r.success).map(r => `${r.device}: ${r.message || '未知错误'}`).join('\n')
        ElMessage.error(`上传失败 ${failCount} 个设备:\n${failDetails}`)
      }
    }
    
    // 更新盒子镜像列表
    await fetchBoxImages()
    
  } catch (error) {
    console.error('批量上传镜像失败:', error)
    ElMessage.error('批量上传失败: ' + error.message)
    
    // 更新任务状态为失败
    const task = taskQueue.value.find(t => t.id === taskId)
    if (task) {
      task.status = 'failed'
      task.endTime = new Date()
      task.progress = 0
      task.error = error.message
    }
  } finally {
    // 隐藏加载状态
    isUploadingToMultipleDevices.value = false
    isUploadingImage.value = false
    currentUploadImage.value = null
    uploadProgress.value = 0
  }
}

// 兼容设备列表的ref变量
// 🔧 方案1修复：直接使用 compatibleDevicesForUpload，不再需要 compatibleDevicesList
// 排序后的兼容设备列表（按IP地址排序，只显示在线设备）
const sortedCompatibleDevicesList = computed(() => {
  // 直接使用 compatibleDevicesForUpload 计算属性作为数据源（实时更新）
  const devicesList = compatibleDevicesForUpload.value
  
  if (!devicesList || devicesList.length === 0) {
    return []
  }
  
  // 过滤出在线设备（严格模式：只信任 devicesStatusCache，忽略 device.isOnline）
  const onlineDevices = devicesList.filter(device => {
    const status = devicesStatusCache.value.get(device.id)
    const isOnline = status === 'online'
    // 调试日志
    if (!isOnline && device.isOnline) {
      console.log(`[设备过滤] 设备 ${device.ip} 的 device.isOnline=true 但 statusCache=${status}，已过滤`)
    }
    return isOnline
  })
  
  // 按IP地址排序
  return [...onlineDevices].sort((a, b) => {
    // 将IP地址转换为数字数组进行比较
    const ipA = a.ip.split('.').map(Number)
    const ipB = b.ip.split('.').map(Number)
    for (let i = 0; i < 4; i++) {
      if (ipA[i] !== ipB[i]) {
        return ipA[i] - ipB[i]
      }
    }
    return 0
  })
})

// 设置云机管理函数的依赖
onMounted(() => {
  setDependencies({
    getDevicePassword: getDevicePassword,
    showAuthDialog: showAuthDialog,
    refreshContainerList: fetchAndroidContainers,
    refreshDeviceInfo: (deviceIp) => {
      // 刷新特定设备的信息
      if (activeDevice.value && activeDevice.value.ip === deviceIp) {
        fetchAndroidContainers()
      }
    }
  })
})

// 程序启动时自动检查更新
onMounted(async () => {
  try {
    await getUpdateVersionInfo()
    console.log('[App] 当前版本:', updateState.currentVersion)
    
    if (updateState.autoCheck) {
      console.log('[App] 自动检查更新已开启，开始检查更新...')
      setTimeout(async () => {
        await checkForUpdates()
      }, 2000)
    }
  } catch (error) {
    console.error('[App] 初始化更新服务失败:', error)
  }
})

// 监听更新状态变化，自动弹出更新提示
watch(
  () => updateState.hasUpdate,
  (hasUpdate, prevHasUpdate) => {
    if (hasUpdate && !prevHasUpdate) {
      console.log('[App] 发现新版本，自动显示更新提示')
      updateInfo.value = updateState.updateInfo
      updateDialogVisible.value = true
    }
  }
)

// 🔧 方案1修复：不再需要 watch 监听器，因为 sortedCompatibleDevicesList 直接使用 compatibleDevicesForUpload

// 监听机型类型变化，切换时获取本地机型列表
watch(
  () => createForm.value.modelType,
  async (newType, oldType) => {
    if (newType === 'local' && oldType === 'online') {
      // 从在线机型切换到本地机型时，获取本地机型列表
      if (createDevice.value && createDevice.value.version === 'v3') {
        await getLocalPhoneModels(createDevice.value.ip)
      }
    }
  }
)

// 处理云机管理模式变化
const handleCloudManageModeChange = (mode) => {
  console.log('云机管理模式变化:', mode, '当前模式:', cloudManageMode.value)
  
  // 切换模式时清空所有选择状态，防止批量操作时使用旧数据
  if (cloudManageMode.value !== mode) {
    selectedCloudMachines.value = [] // 清空批量模式下的云机选择
    selectedCloudDevice.value = null  // 清空坑位模式下的设备选择，让用户重新选择
    activeDevice.value = null // 清空当前活跃设备，防止在切换模式后仍然刷新截图
    treeSelectedKeys.value = []       // 清空树形结构的选中状态
    console.log('已清空所有选择状态')
  }
  
  // 切换到批量模式前，先同步计算过滤后的数据，避免闪烁
  if (mode === 'batch') {
    cloudMachineGroups.value = computeCloudMachineGroups('batch')
  }
  
  cloudManageMode.value = mode
  
  // 切换模式时重新初始化分组数据，应用不同的过滤规则
  // 批量模式：只显示运行中的云机
  // 坑位模式：显示所有云机
  // 注意：批量模式已经在上面同步计算过了，这里的调用会被防抖机制处理
  initCloudMachineGroups()
  
  // 如果切换到坑位模式，自动选择第一个设备
  if (mode === 'slot') {
    // 获取第一个在线设备
    const firstOnlineDevice = devices.value.find(device => {
      return devicesStatusCache.value.get(device.id) === 'online'
    })
    
    if (firstOnlineDevice) {
      console.log('坑位模式：自动选择第一个在线设备:', firstOnlineDevice.ip)
      selectedCloudDevice.value = firstOnlineDevice
      activeDevice.value = firstOnlineDevice

      // 立即用已有缓存渲染，避免等待期间显示空白
      const cachedMachines = deviceCloudMachinesCache.value.get(firstOnlineDevice.ip)
      if (cachedMachines && cachedMachines.length > 0) {
        instances.value = cachedMachines
        allInstances.value = cachedMachines
        updateCloudMachines()
      }

      // 后台触发刷新，完成后重新初始化分组
      fetchAndroidContainers(firstOnlineDevice, true).then(() => {
        initCloudMachineGroups()
      })
    } else {
      console.log('坑位模式：未找到在线设备')
    }
  }
  
  // 如果切换到批量模式，检查是否有设备未加载云机数据，如果有则自动加载
  if (mode === 'batch') {
    
    // 获取所有在线但没有云机数据的设备
    const devicesToLoad = devices.value.filter(device => {
      // 检查设备是否在线
      const isOnline = devicesStatusCache.value.get(device.id) === 'online'
      if (!isOnline) return false
      
      // 检查是否已有缓存数据
      const cachedData = deviceCloudMachinesCache.value.get(device.ip)
      const hasData = cachedData && cachedData.length > 0
      
      return !hasData
    })
    
    if (devicesToLoad.length > 0) {
      console.log(`批量模式：发现 ${devicesToLoad.length} 个设备未加载云机数据，开始自动加载`)
      // 并行加载，限制并发数为5，避免网络拥塞
      const batchSize = 5
      const loadBatch = async (index) => {
        if (index >= devicesToLoad.length) return
        
        const batch = devicesToLoad.slice(index, index + batchSize)
        const promises = batch.map(device => fetchAndroidContainers(device, true)) // 使用true强制显示加载状态
        
        await Promise.allSettled(promises)
        // 每批加载完更新一次界面
        initCloudMachineGroups()
        
        // 递归加载下一批
        loadBatch(index + batchSize)
      }
      
      loadBatch(0)
    }
  }
}

// 处理选中云机设备变化
const handleSelectedCloudDeviceChange = (device) => {
  console.log('选中云机设备变化:', device)
  selectedCloudDevice.value = device
  // 切换设备时清空版本快照，强制下次轮询立即拉取新设备截图
  screenshotLocalVersions = {}
  
  // 同步更新activeDevice，确保实例数据正确更新
  if (device && activeDevice.value?.ip !== device.ip) {
    console.log('同步更新activeDevice:', device)
    activeDevice.value = device

    // 立即用新设备的已有缓存渲染界面，避免停留在旧设备数据
    const cachedMachines = deviceCloudMachinesCache.value.get(device.ip)
    if (cachedMachines && cachedMachines.length > 0) {
      // 已有缓存：直接更新 instances / cloudMachines，界面立即切换
      instances.value = cachedMachines
      allInstances.value = cachedMachines
      updateCloudMachines()
    } else {
      // 无缓存：清空界面，等待加载
      instances.value = []
      allInstances.value = []
      cloudMachines.value = []
    }

    // 后台触发刷新（不阻塞界面），拿到最新数据后自动更新
    fetchAndroidContainers(device, true)
  }
}

// 批量操作处理
const handleBatchAction = async (action, selectedData = [], cardOrientation = null) => {
  // 对于"停止批量控制"操作，不需要检查是否选中云机
  const isStoppingControl = action === 'projection-control' && isBatchProjectionControlling.value
  
  if (!isStoppingControl) {
    // 根据云机管理模式检查是否有选中的云机
    let hasSelectedMachines = false
    if (cloudManageMode.value === 'slot') {
      // 坑位模式：selectedData 是坑位号数组
      hasSelectedMachines = Array.isArray(selectedData) && selectedData.length > 0
    } else {
      // 批量模式：selectedData 是云机对象数组
      hasSelectedMachines = Array.isArray(selectedData) && selectedData.length > 0
    }
    
    if (!hasSelectedMachines) {
      console.error('没有选中的云机')
      ElMessage.error('请先选中要操作的云机')
      return
    }
  }
  
  
  loading.value = true
  try {
    switch (action) {
      case 'restart':
        // 实现批量重启功能
        console.log(`执行批量重启操作`)
        
        let containersToRestart = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          containersToRestart = instances.value.filter(inst => 
            Array.isArray(selectedData) && selectedData.includes(inst.indexNum) && 
            inst.status === 'running' // 只选择运行中的容器
          )
        } else {
          // 批量模式：直接使用传递的云机对象数组
          containersToRestart = selectedData.filter(machine => 
            machine.status === 'running' // 只选择运行中的容器
          )
        }
        
        if (containersToRestart.length === 0) {
          ElMessage.warning('没有选中运行中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(
            `确定要重启选中的 ${containersToRestart.length} 个运行中的云机吗？重启后容器将会停止并重新启动。`, 
            '批量重启云机', 
            {
              confirmButtonText: '确定',
              cancelButtonText: '取消',
              type: 'warning'
            }
          )
          
          // 为每个容器添加设备信息
          const restartTargets = containersToRestart.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('restart', restartTargets)
          executeTask(taskId)
          
          ElMessage.success('批量重启任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量重启操作')
            ElMessage.info('已取消批量重启操作')
          } else {
            console.error('批量重启失败:', error)
            ElMessage.error(`批量重启失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'reset':
        // 实现批量重置功能
        console.log(`执行批量重置操作`)
        
        let resetContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          resetContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          resetContainersToOperate = selectedData
        }
        
        if (resetContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        try {
          // 显示确认对话框，让用户选择是否开机
          let startAfterReset = true
          await new Promise((resolve, reject) => {
            ElMessageBox.confirm(
              `确定要重置选中的 ${resetContainersToOperate.length} 个容器吗？重置后容器将会被恢复到初始状态。`,
              '批量重置容器',
              {
                confirmButtonText: '重置并开机',
                cancelButtonText: '仅重置(不开机)',
                distinguishCancelAndClose: true,
                type: 'warning',
                showClose: true
              }
            ).then(() => {
              startAfterReset = true
              console.log('[批量重置] 用户选择: 重置并开机, start=true')
              resolve()
            }).catch((action) => {
              console.log('[批量重置] catch action:', action, typeof action)
              if (action === 'close') {
                // 用户点了 X 或按 ESC，取消操作
                reject('cancel')
              } else {
                // 用户点了"仅重置(不开机)"按钮
                startAfterReset = false
                console.log('[批量重置] 用户选择: 仅重置(不开机), start=false')
                resolve()
              }
            })
          })
          
          // 检查设备版本是否支持重置功能
          let allDevicesSupported = true
          const unsupportedDevices = new Set()
          
          if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
            if (selectedCloudDevice.value.version !== 'v3') {
              allDevicesSupported = false
              unsupportedDevices.add(selectedCloudDevice.value.ip)
            }
          } else {
            // 批量模式下检查所有选中容器的设备版本
            resetContainersToOperate.forEach(container => {
              if ((container.deviceVersion || 'v3') !== 'v3') {
                allDevicesSupported = false
                unsupportedDevices.add(container.deviceIp)
              }
            })
          }
          
          if (!allDevicesSupported) {
            ElMessage.warning(`以下设备不支持重置容器功能: ${Array.from(unsupportedDevices).join(', ')}`)
            break
          }
          
          // 为每个容器添加设备信息
          const resetTargets = resetContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列，通过 metadata 传递 start 参数
          const taskId = addTaskToQueue('reset', resetTargets, { start: startAfterReset })
          executeTask(taskId)
          
          ElMessage.success('批量重置任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量重置操作')
            ElMessage.info('已取消批量重置操作')
          } else {
            console.error('批量重置失败:', error)
            ElMessage.error(`批量重置失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'projection':
        // 实现批量投屏功能
        console.log(`执行批量投屏操作`)
        
        let projectionContainersToOperate = []
        let allContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          allContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          allContainersToOperate = selectedData
        }
        
        // 只保留运行中的云机
        projectionContainersToOperate = allContainersToOperate.filter(container => container.status === 'running')
        
        if (allContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        if (projectionContainersToOperate.length === 0) {
          ElMessage.warning('选中的云机都没有处于运行状态')
          break
        }
        
        // 如果有部分云机未运行，给出提示
        if (projectionContainersToOperate.length < allContainersToOperate.length) {
          const notRunningCount = allContainersToOperate.length - projectionContainersToOperate.length
          ElMessage.info(`有 ${notRunningCount} 个选中的云机未处于运行状态，将只对运行中的 ${projectionContainersToOperate.length} 个云机打开投屏`)
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要对选中的 ${projectionContainersToOperate.length} 个云机打开投屏吗？`, '批量投屏', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'info'
          })
          
          // 计算 orient 参数（如果有 cardOrientation）
          // horizontal: 横屏 = 1, vertical: 竖屏 = 0
          const customOrient = cardOrientation ? (cardOrientation === 'horizontal' ? 1 : 0) : null
          console.log('[App.vue 批量投屏] cardOrientation:', cardOrientation, ', customOrient:', customOrient)
          
          // 执行批量投屏操作，不使用任务队列
          let successCount = 0
          let failCount = 0
          
          for (const container of projectionContainersToOperate) {
            try {
              if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                // 坑位模式
                await startProjection({ ip: selectedCloudDevice.value.ip }, container, customOrient)
              } else {
                // 批量模式
                await startProjection({ ip: container.deviceIp }, container, customOrient)
              }
              successCount++
            } catch (error) {
              console.error(`对云机 ${container.name || container.ID} 打开投屏失败:`, error)
              failCount++
            }
          }
          
          // 显示操作结果
          if (successCount > 0) {
            ElMessage.success(`成功对 ${successCount} 个云机打开投屏`)
          }
          if (failCount > 0) {
            ElMessage.warning(`对 ${failCount} 个云机打开投屏失败`)
          }
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量投屏操作')
            ElMessage.info('已取消批量投屏操作')
          } else {
            console.error('批量投屏失败:', error)
            ElMessage.error(`批量投屏失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'projection-control':
        // 批量投屏控制
        console.log(`执行批量投屏控制`)

        // 如果正在进行批量控制，则停止
        if (isBatchProjectionControlling.value) {
          try {
            await stopProjectionBatchControl()
            // 根据模式清除对应的状态
            if (cloudManageMode.value === 'batch') {
              batchModeProjectionControlling.value = false
            } else {
              const deviceIp = selectedCloudDevice.value?.ip
              if (deviceIp) {
                delete slotModeProjectionControlStatus.value[deviceIp]
              }
            }
          } catch (error) {
            console.error('停止批量投屏控制失败:', error)
            ElMessage.error(`停止批量投屏控制失败: ${error.message || '未知错误'}`)
          }
          break
        }

        let controlContainersToOperate = []
        let allControlContainersToOperate = []

        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          allControlContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          allControlContainersToOperate = selectedData
        }

        // 只保留运行中的云机
        controlContainersToOperate = allControlContainersToOperate.filter(container => container.status === 'running')

        if (allControlContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }

        if (controlContainersToOperate.length === 0) {
          ElMessage.warning('选中的云机都没有处于运行状态')
          break
        }

        if (controlContainersToOperate.length < allControlContainersToOperate.length) {
          const notRunningCount = allControlContainersToOperate.length - controlContainersToOperate.length
          ElMessage.info(`有 ${notRunningCount} 个选中的云机未处于运行状态，将只对运行中的 ${controlContainersToOperate.length} 个云机进行批量投屏控制`)
        }

        try {
          await ElMessageBox.confirm(`确定要对选中的 ${controlContainersToOperate.length} 个云机进行批量投屏控制吗？`, '批量投屏控制', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'info'
          })

          const device = (cloudManageMode.value === 'slot' && selectedCloudDevice.value)
            ? { ip: selectedCloudDevice.value.ip }
            : null

          // 计算 customOrient: horizontal -> 1, vertical -> 0
          const customOrient = cardOrientation ? (cardOrientation === 'horizontal' ? 1 : 0) : null
          console.log('[App.vue projection-control] cardOrientation:', cardOrientation, ', customOrient:', customOrient)
          
          await startProjectionBatchControl(device, controlContainersToOperate, `批量投屏控制`, customOrient)
          
          // 根据模式设置对应的状态
          if (cloudManageMode.value === 'batch') {
            batchModeProjectionControlling.value = true
          } else {
            const deviceIp = selectedCloudDevice.value?.ip
            if (deviceIp) {
              slotModeProjectionControlStatus.value[deviceIp] = true
            }
          }
        } catch (error) {
          if (error === 'cancel') {
            console.log('用户取消了批量投屏控制操作')
            ElMessage.info('已取消批量投屏控制操作')
          } else {
            console.error('批量投屏控制失败:', error)
            ElMessage.error(`批量投屏控制失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'shutdown':
        // 实现批量关机功能
        console.log(`执行批量关机操作`)
        
        let shutdownContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          shutdownContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          shutdownContainersToOperate = selectedData
        }
        
        if (shutdownContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要关闭选中的 ${shutdownContainersToOperate.length} 个容器吗？关闭后容器将会停止运行。`, '批量关机容器', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          })
          
          // 为每个容器添加设备信息
          const shutdownTargets = shutdownContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('shutdown', shutdownTargets)
          executeTask(taskId)
          
          ElMessage.success('批量关机任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量关机操作')
            ElMessage.info('已取消批量关机操作')
          } else {
            console.error('批量关机失败:', error)
            ElMessage.error(`批量关机失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'switch-backup':
        // 实现批量切换备份功能（自动切换到创建时间最新的备份）
        console.log(`执行批量切换备份操作`)
        
        {
          let switchBackupContainersToOperate = []
          
          // 根据云机管理模式获取需要操作的容器
          if (cloudManageMode.value === 'slot') {
            // 坑位模式：根据选中的坑位号获取容器实例
            switchBackupContainersToOperate = instances.value.filter(inst =>
              Array.isArray(selectedData) && selectedData.includes(inst.indexNum)
            )
          } else {
            // 批量模式：直接使用传递的云机对象数组
            switchBackupContainersToOperate = selectedData
          }
          
          if (switchBackupContainersToOperate.length === 0) {
            ElMessage.warning('没有选中的云机')
            break
          }
          
          // 对每个选中的容器，在 allInstances 中查找同坑位创建时间最新的其他备份
          const switchBackupTargets = []
          const noBackupSlots = []
          
          for (const container of switchBackupContainersToOperate) {
            const slotNum = container.indexNum
            const deviceIp = cloudManageMode.value === 'slot' && selectedCloudDevice.value
              ? selectedCloudDevice.value.ip
              : container.deviceIp
            
            // 查找同一设备同一坑位的所有容器（备份）
            // 坑位模式：从 allInstances 查找（包含当前设备所有容器）
            // 批量模式：从 deviceCloudMachinesCache 查找
            const allDeviceContainers = cloudManageMode.value === 'slot'
              ? allInstances.value
              : (deviceCloudMachinesCache.value.get(deviceIp) || [])
            
            const slotAllContainers = allDeviceContainers.filter(inst =>
              inst.indexNum === slotNum &&
              inst.name &&
              inst.name !== container.name
            )
            
            if (slotAllContainers.length === 0) {
              noBackupSlots.push(slotNum)
              continue
            }
            
            // 按创建时间降序排序，取最新的一条
            const latestBackup = slotAllContainers.sort((a, b) => {
              const timeA = a.created ? new Date(a.created).getTime() : 0
              const timeB = b.created ? new Date(b.created).getTime() : 0
              return timeB - timeA
            })[0]
            
            switchBackupTargets.push({
              currentContainer: container,
              backupContainer: latestBackup,
              deviceIp,
              slotNum
            })
          }
          
          if (noBackupSlots.length > 0) {
            ElMessage.warning(`坑位 ${noBackupSlots.join('、')} 没有可切换的备份，将跳过`)
          }
          
          if (switchBackupTargets.length === 0) {
            ElMessage.warning('所有选中的云机都没有可切换的备份')
            break
          }
          
          try {
            await ElMessageBox.confirm(
              `确定要对选中的 ${switchBackupTargets.length} 个云机切换到最新备份吗？\n操作将依次关闭当前运行容器并启动最新备份。`,
              '批量切换备份',
              {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
              }
            )
            
            const device = cloudManageMode.value === 'slot' && selectedCloudDevice.value
              ? selectedCloudDevice.value
              : null
            
            // 初始化进度列表
            batchSwitchBackupTotal.value = switchBackupTargets.length
            batchSwitchBackupDone.value = 0
            batchSwitchBackupProgressList.value = switchBackupTargets.map(t => ({
              slotNum: t.slotNum,
              deviceIp: t.deviceIp,
              currentName: t.currentContainer.name,
              backupName: t.backupContainer.name,
              status: 'pending',
              message: '等待中'
            }))
            batchSwitchBackupProgressVisible.value = true
            
            let successCount = 0
            let failCount = 0
            
            for (let i = 0; i < switchBackupTargets.length; i++) {
              const target = switchBackupTargets[i]
              const progressItem = batchSwitchBackupProgressList.value[i]
              progressItem.status = 'running'
              progressItem.message = '关机中...'
              
              try {
                const targetDevice = device || { ip: target.deviceIp }
                
                // 1. 关机当前运行的容器
                if (target.currentContainer.status === 'running') {
                  await authRetry(targetDevice, async (password) => {
                    await stopContainer(targetDevice, target.currentContainer.name, password)
                  })
                  await new Promise(resolve => setTimeout(resolve, 500))
                }
                
                // 2. 启动备份容器
                progressItem.message = '启动备份中...'
                await authRetry(targetDevice, async (password) => {
                  await startContainer(targetDevice, target.backupContainer.name, password)
                })
                
                progressItem.status = 'success'
                progressItem.message = '切换成功'
                successCount++
              } catch (err) {
                console.error(`坑位 ${target.slotNum} 切换备份失败:`, err)
                progressItem.status = 'failed'
                progressItem.message = err.message || '切换失败'
                failCount++
              }
              
              batchSwitchBackupDone.value = i + 1
            }
            
            // 3. 刷新容器列表
            if (device) {
              await fetchAndroidContainers(device)
            } else {
              // 批量模式：按设备分组刷新
              const deviceIpSet = new Set(switchBackupTargets.map(t => t.deviceIp))
              for (const ip of deviceIpSet) {
                const d = devices.value.find(dev => dev.ip === ip)
                if (d) await fetchAndroidContainers(d)
              }
            }
            
            if (successCount > 0) {
              ElMessage.success(`批量切换备份成功：${successCount} 个`)
            }
            if (failCount > 0) {
              ElMessage.warning(`${failCount} 个切换失败`)
            }
          } catch (error) {
            if (error === 'cancel') {
              ElMessage.info('已取消批量切换备份操作')
            } else {
              console.error('批量切换备份失败:', error)
              ElMessage.error(`批量切换备份失败: ${error.message || '未知错误'}`)
            }
          }
        }
        break
      case 'delete':
        // 实现批量删除功能
        console.log(`执行批量删除操作`)
        
        let deleteContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          deleteContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          deleteContainersToOperate = selectedData
        }
        
        if (deleteContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要删除选中的 ${deleteContainersToOperate.length} 个云机吗？删除后数据将无法恢复。`, '批量删除云机', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'danger'
          })
          
          // 为每个容器添加设备信息
          const deleteTargets = deleteContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('delete', deleteTargets)
          executeTask(taskId)
          
          ElMessage.success('批量删除任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量删除操作')
            ElMessage.info('已取消批量删除操作')
          } else {
            console.error('批量删除失败:', error)
            ElMessage.error(`批量删除失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'switchModel':
        // 实现批量切换机型功能
        console.log(`执行批量切换机型操作`)
        
        let switchModelContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          switchModelContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          switchModelContainersToOperate = selectedData
        }
        
        if (switchModelContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        // 检查是否所有选中的云机都处于运行状态
        const runningContainers = switchModelContainersToOperate.filter(container => container.status === 'running')
        
        if (runningContainers.length === 0) {
          ElMessage.warning('没有选中已运行的云机，批量切换机型操作仅支持已运行的云机')
          break
        }
        
        if (runningContainers.length < switchModelContainersToOperate.length) {
          ElMessage.warning(`部分选中的云机未运行，仅对 ${runningContainers.length} 个已运行的云机执行操作`)
          // 更新要操作的容器列表，只包含已运行的云机
          switchModelContainersToOperate = runningContainers
        }
        
        // 检查设备版本是否支持切换机型功能
        let newIsV3Device = false
        let newTargetDevice = null
        
        if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
          newIsV3Device = selectedCloudDevice.value.version === 'v3'
          newTargetDevice = selectedCloudDevice.value
        } else if (cloudManageMode.value === 'batch') {
          // 批量模式下，确保所有选中的云机都在同一个设备上
          const deviceIps = new Set(switchModelContainersToOperate.map(machine => machine.deviceIp))
          if (deviceIps.size !== 1) {
            ElMessage.error('批量切换机型功能只支持同一设备上的云机')
            break
          }
          
          const deviceIp = Array.from(deviceIps)[0]
          const versions = new Set(switchModelContainersToOperate.map(machine => machine.deviceVersion))
          newIsV3Device = versions.has('v3') && versions.size === 1
          
          if (newIsV3Device) {
            newTargetDevice = { ip: deviceIp, version: 'v3' }
          }
        }
        
        if (!newIsV3Device || !newTargetDevice) {
          ElMessage.error('只有V3版本设备支持批量切换机型功能')
          break
        }
        
        try {
          // 显示加载状态提示
          const loadingMsg = ElMessage({
            message: '加载机型列表中...',
            type: 'info',
            duration: 0
          })
          
          // 获取可用的手机型号列表
          await getV3PhoneModels(newTargetDevice.ip)
          
          // 关闭加载提示
          setTimeout(() => {
            ElMessage.closeAll()
          }, 100)
          
          if (phoneModels.value.length === 0) {
            ElMessage.warning('未获取到可用的机型列表')
            break
          }
          
          // 为每个容器添加设备信息
          batchSwitchModelTargets.value = newContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 设置操作类型为'new'，表示这是通过批量新机按钮触发的操作
          batchSwitchModelOperationType.value = 'new'
          
          // 初始化机型槽
          initModelSlots()
          
          // 打开批量切换机型对话框
          batchSwitchModelDialogVisible.value = true
        } catch (error) {
          console.error('批量切换机型失败:', error)
          ElMessage.error(`批量切换机型失败: ${error.message || '未知错误'}`)
        }
        break
      case 'new':
        // 实现批量切换机型功能（批量新机）
        console.log(`执行批量切换机型操作（批量新机）`)
        
        let newContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          newContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          newContainersToOperate = selectedData
        }
        
        if (newContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        // 检查是否所有选中的云机都处于运行状态
        const runningNewContainers = newContainersToOperate.filter(container => container.status === 'running')
        
        if (runningNewContainers.length === 0) {
          ElMessage.warning('没有选中已运行的云机，批量新机操作仅支持已运行的云机')
          break
        }
        
        if (runningNewContainers.length < newContainersToOperate.length) {
          ElMessage.warning(`部分选中的云机未运行，仅对 ${runningNewContainers.length} 个已运行的云机执行操作`)
          // 更新要操作的容器列表，只包含已运行的云机
          newContainersToOperate = runningNewContainers
        }
        
        // 检查设备版本是否支持切换机型功能
        let isV3Device = false
        let targetDevice = null
        
        if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
          isV3Device = selectedCloudDevice.value.version === 'v3'
          targetDevice = selectedCloudDevice.value
        } else if (cloudManageMode.value === 'batch') {
          // 批量模式下，确保所有选中的云机都在同一个设备上
          const deviceIps = new Set(newContainersToOperate.map(machine => machine.deviceIp))
          if (deviceIps.size !== 1) {
            ElMessage.error('批量切换机型功能只支持同一设备上的云机')
            break
          }
          const deviceIp = Array.from(deviceIps)[0]
          // const versions = new Set(newContainersToOperate.map(machine => machine.deviceVersion))
          // isV3Device = versions.has('v3') && versions.size === 1
          isV3Device = true
          
          if (isV3Device) {
            targetDevice = { ip: deviceIp, version: 'v3' }
          }
        }
        if (!isV3Device || !targetDevice) {
          ElMessage.error('只有V3版本设备支持批量切换机型功能')
          break
        }
        
        try {
          // 显示加载状态提示
          const loadingMsg = ElMessage({
            message: '加载机型列表中...',
            type: 'info',
            duration: 0
          })
          
          // 获取可用的手机型号列表
          await getV3PhoneModels(targetDevice.ip)
          
          // 关闭加载提示
          setTimeout(() => {
            ElMessage.closeAll()
          }, 100)
          
          if (phoneModels.value.length === 0) {
            ElMessage.warning('未获取到可用的机型列表')
            break
          }
          
          // 为每个容器添加设备信息
          batchSwitchModelTargets.value = newContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 检查是否有V2类型容器
          const v2Containers = batchSwitchModelTargets.value.filter(c => c.androidType === 'V2')
          const nonV2Containers = batchSwitchModelTargets.value.filter(c => c.androidType !== 'V2')
          
          if (v2Containers.length > 0 && nonV2Containers.length > 0) {
            ElMessage.warning(`检测到 ${v2Containers.length} 个V2容器，V2容器不支持指定机型，将自动分配到随机机型`)
          } else if (v2Containers.length > 0) {
            ElMessage.info('所选容器均为V2类型，将使用一键新机功能（不支持指定机型）')
          }
          
          // 初始化机型槽
          initModelSlots()
          
          // 打开批量切换机型对话框
          batchSwitchModelDialogVisible.value = true
        } catch (error) {
          console.error('批量切换机型失败:', error)
          ElMessage.error(`批量切换机型失败: ${error.message || '未知错误'}`)
        }
        break
      case 'upload':
        // 实现批量上传文件功能
        console.log(`执行批量上传文件操作`)
        
        let uploadContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          uploadContainersToOperate = instances.value.filter(inst => Array.isArray(selectedData) && selectedData.includes(inst.indexNum))
        } else {
          // 批量模式：直接使用传递的云机对象数组
          uploadContainersToOperate = selectedData
        }
        
        if (uploadContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        // 检查设备信息是否完整
        if (cloudManageMode.value === 'slot' && (!selectedCloudDevice.value || !selectedCloudDevice.value.ip)) {
          ElMessage.warning('设备信息不完整，无法上传文件')
          break
        } else if (cloudManageMode.value === 'batch') {
          // 批量模式下，无需检查是否有相同的设备，因为handleBatchUpload已经支持多设备上传
          // 这里的检查限制被移除了，以支持多设备批量上传
        }
        
        // 设置批量上传选中的云机
        batchUploadSelectedMachines.value = uploadContainersToOperate
        
        // 打开批量上传对话框
        batchUploadDialogVisible.value = true
        break
      case 'update-image': {
        // 批量更新镜像
        console.log('执行批量更新镜像操作')

        let updateImageContainersTmp = []

        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          updateImageContainersTmp = instances.value.filter(inst =>
            Array.isArray(selectedData) && selectedData.includes(inst.indexNum)
          )
        } else {
          updateImageContainersTmp = selectedData
        }

        if (updateImageContainersTmp.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }

        // 确保镜像列表已加载
        if (imageList.value.length === 0) {
          await fetchImageList('')
        }

        // 按设备类型（P系列 / 非P系列）分组
        const groupMap = new Map() // key: 'p' | 'non-p'
        for (const c of updateImageContainersTmp) {
          const deviceIp = c.deviceIp || (cloudManageMode.value === 'slot' && selectedCloudDevice.value ? selectedCloudDevice.value.ip : '')
          const device = devices.value.find(d => d.ip === deviceIp)
          const deviceName = device?.name || (cloudManageMode.value === 'slot' ? selectedCloudDevice.value?.name : '') || ''
          const isPSeries = isBatchImagePSeries(deviceName)
          const groupKey = isPSeries ? 'p' : 'non-p'
          if (!groupMap.has(groupKey)) {
            groupMap.set(groupKey, {
              groupKey,
              groupLabel: isPSeries ? 'P系列设备' : 'CRQ系列设备',
              deviceName,
              containers: [],
              hasV2: false,
              hasV3: false,
              androidType: 'V3',
              v2AndroidVersion: 10,
              selectedUrl: '',
              customUrl: ''
            })
          }
          const group = groupMap.get(groupKey)
          const containerWithIp = cloudManageMode.value === 'slot' && selectedCloudDevice.value
            ? { ...c, deviceIp: selectedCloudDevice.value.ip }
            : c
          group.containers.push(containerWithIp)
          if (c.androidType === 'V2') group.hasV2 = true
          else group.hasV3 = true
        }

        // 为每组设置默认 androidType
        for (const group of groupMap.values()) {
          if (group.hasV3) {
            group.androidType = 'V3'
          } else {
            group.androidType = 'V2'
          }
          // 默认选中第一条镜像
          if (group.androidType === 'V3') {
            const v3list = getBatchUpdateV3List(group.deviceName)
            group.selectedUrl = v3list.length > 0 ? v3list[0].url : ''
          }
        }

        batchUpdateImageGroups.value = Array.from(groupMap.values())
        batchUpdateImageDialogVisible.value = true
        break
      }
      default:
        console.error('未知的批量操作:', action)
        ElMessage.error('未知的批量操作')
    }
  } catch (error) {
    console.error('执行批量操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    loading.value = false
  }
}

// 上传本地镜像到设备
const uploadLocalImageToDevice = async (image) => {
  try {
    // 显示设备选择对话框
    currentUploadingImage.value = {
      ...image,
      localPath: image.url // 本地镜像直接使用URL作为本地路径
    }
    selectedDevicesForUpload.value = []
    showDeviceSelectionDialog.value = true
  } catch (error) {
    console.error('上传本地镜像到设备失败:', error)
    ElMessage.error('操作失败')
  }
}

// 删除本地缓存镜像
const deleteLocalCachedImage = async (image) => {
  return new Promise((resolve) => {
    ElMessageBox.confirm(`确定要删除镜像 ${image.name} 吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      try {
        const result = await DeleteLocalImage(image.path)
        if (result.code === 0) {
          ElMessage.success('删除镜像成功')
          await fetchLocalCachedImages()
          resolve(true)
        } else {
          ElMessage.error(`删除镜像失败: ${result.message || '未知错误'}`)
          resolve(false)
        }
      } catch (error) {
        console.error('删除本地镜像失败:', error)
        ElMessage.error('操作失败')
        resolve(false)
      }
    }).catch(() => {
      resolve(false)
    })
  })
}

// 删除已下载的在线镜像
const deleteDownloadedImage = async (onlineImage) => {
  // 先刷新本地镜像列表，确保数据最新
  await fetchLocalCachedImages()

  const imageUrl = onlineImage.url

  // 通过本地镜像记录的 onlineUrl 字段与在线镜像 URL 精确匹配
  const localImage = localCachedImages.value.find(localImg => localImg.onlineUrl === imageUrl)

  if (localImage) {
    // 使用找到的本地镜像执行与本地镜像删除完全相同的操作
    const success = await deleteLocalCachedImage(localImage)
    if (success) {
      imageDownloadStatus.value.set(onlineImage.url, false)
    }
  } else {
    ElMessage.error('未找到对应的本地镜像文件')
  }
}

// 下载进度更新事件处理函数
const handleDownloadProgress = (event) => {
  // Wails 事件系统传递的是 WailsEvent 对象，需要从 event.data 获取实际数据
  const data = event.data || event
  
  if (data && data.progress !== undefined) {
    // 严格验证：必须有活动的下载任务
    if (!isDownloadingImage.value) {
      return
    }
    
    if (!currentDownloadTaskId.value) {
      return
    }
    
    if (!currentDownloadImage.value) {
      return
    }
    
    // 验证任务是否还在运行
    const currentTask = taskQueue.value.find(task => task.id === currentDownloadTaskId.value)
    if (!currentTask) {
      return
    }
    
    if (currentTask.status !== 'running') {
      return
    }
    
    // 检查会话时间戳是否匹配
    const taskSessionTime = currentTask.sessionStartTime || 0
    if (taskSessionTime !== downloadStartTime.value) {
      console.log(`[进度事件] 忽略：会话时间不匹配 (任务:${taskSessionTime}, 当前:${downloadStartTime.value})`)
      return
    }
    
    const newProgress = Math.round(data.progress)
    const currentProgress = currentTask.progress || 0
    
    // 严格的进度验证逻辑
    // 1. 如果新进度比当前进度小超过3%，视为异常（可能是旧任务）
    if (newProgress < currentProgress - 3) {
      console.log(`[进度事件] 忽略进度回退: ${currentProgress}% -> ${newProgress}%`)
      return
    }
    
    // 2. 如果当前进度已经超过20%，但新进度小于10%，明显是新旧任务混杂
    if (currentProgress > 20 && newProgress < 10) {
      console.log(`[进度事件] 忽略跨度异常: ${currentProgress}% -> ${newProgress}%`)
      return
    }
    
    // 3. 限制单次进度跳跃不能超过30%（正常下载不会有这么大的跳跃）
    if (newProgress - currentProgress > 30) {
      console.log(`[进度事件] 忽略异常跳跃: ${currentProgress}% -> ${newProgress}%`)
      return
    }
    
    // 更新进度值
    downloadProgress.value = data.progress
    currentTask.progress = newProgress
  }
}

// 下载完成事件处理函数
const handleDownloadComplete = async (event) => {
  console.log('下载完成事件:', event)
  
  // Wails 事件系统传递的是 WailsEvent 对象，需要从 event.data 获取实际数据
  const data = event.data || event
  
  // 如果没有活动的下载任务，忽略此事件（可能是取消任务后的残留事件）
  if (!isDownloadingImage.value || !currentDownloadTaskId.value) {
    console.log('收到下载完成事件但无活动任务，忽略')
    return
  }
  
  // 验证任务状态
  const currentTask = taskQueue.value.find(task => task.id === currentDownloadTaskId.value)
  if (!currentTask || (currentTask.status !== 'running' && currentTask.status !== 'pending')) {
    console.log('收到下载完成事件但任务状态异常，忽略')
    return
  }
  
  try {
    if (data.success) {
      // 下载成功，确保进度显示为100%
      downloadProgress.value = 100
      
      // 更新任务状态
      currentTask.status = 'completed'
      currentTask.progress = 100
      currentTask.endTime = new Date()
      currentTask.completed = 1
      
      ElMessage.success(`镜像${currentDownloadImage.value?.name}下载成功`)
      
      // 重新获取本地缓存镜像列表
      await fetchLocalCachedImages()
      
      // 更新在线镜像列表的下载状态
      if (currentDownloadImage.value) {
        imageDownloadStatus.value.set(currentDownloadImage.value.url, true)
      }
    } else {
      // 下载失败
      currentTask.status = 'failed'
      currentTask.endTime = new Date()
      currentTask.failed = 1
      currentTask.error = data.message
      
      ElMessage.error(`下载镜像失败: ${data.message}`)
    }
  } finally {
    // 延迟清理状态，确保UI有时间显示最终状态
    setTimeout(() => {
      isDownloadingImage.value = false
      currentDownloadImage.value = null
      currentDownloadTaskId.value = null
      downloadProgress.value = 0
      downloadStartTime.value = 0
    }, 300)
  }
}

// 处理上传进度事件
const handleUploadProgress = (event) => {
  const data = event.data || event
  
  if (data && data.progress !== undefined) {
    uploadProgress.value = data.progress
    
    const runningUploadTask = taskQueue.value.find(task => 
      task.type === 'uploadImage' && task.status === 'running'
    )
    
    if (runningUploadTask) {
      if (runningUploadTask.deviceIps && runningUploadTask.deviceIps.length > 0) {
        if (data.deviceIP && runningUploadTask.deviceProgress) {
          const deviceIP = data.deviceIP
          
          if (!runningUploadTask.deviceProgress[deviceIP]) {
            runningUploadTask.deviceProgress[deviceIP] = { 
              total: 1,
              completed: 0, 
              failed: 0,
              currentProgress: 0
            }
          }
          
          runningUploadTask.deviceProgress[deviceIP].currentProgress = Math.round(data.progress)
          
          if (data.progress >= 100) {
            runningUploadTask.deviceProgress[deviceIP].completed = 1
            runningUploadTask.deviceProgress[deviceIP].currentProgress = 100
          }
        }
      } else {
        runningUploadTask.progress = Math.round(data.progress)
      }
    }
  }
}

// 处理上传完成事件
const handleUploadComplete = (event) => {
  const data = event.data || event
  
  if (!isUploadingImage.value) {
    return
  }
  
  try {
    if (!data || typeof data !== 'object') {
      return
    }
    
    const isSuccess = data.success === true || data.success === 'true' || data.success === 1 || data.success === '1'
    
    const runningUploadTask = taskQueue.value.find(task => 
      task.type === 'uploadImage' && task.status === 'running'
    )
    
    if (!runningUploadTask) {
      return
    }
    
    if (data.deviceIP && runningUploadTask.deviceProgress) {
      const deviceIP = data.deviceIP
      
      if (!runningUploadTask.deviceProgress[deviceIP]) {
        runningUploadTask.deviceProgress[deviceIP] = { 
          total: 1,
          completed: 0, 
          failed: 0,
          currentProgress: 0
        }
      }
      
      if (isSuccess) {
        runningUploadTask.deviceProgress[deviceIP].completed = 1
        runningUploadTask.deviceProgress[deviceIP].currentProgress = 100
      } else {
        runningUploadTask.deviceProgress[deviceIP].failed = 1
        runningUploadTask.deviceProgress[deviceIP].currentProgress = 0
        runningUploadTask.deviceProgress[deviceIP].error = data.message || '上传失败'
        
        // 立即收集失败原因到 failedTargets
        const existingIndex = runningUploadTask.failedTargets.findIndex(t => t.deviceIP === deviceIP)
        if (existingIndex >= 0) {
          runningUploadTask.failedTargets[existingIndex].error = data.message || '上传失败'
        } else {
          runningUploadTask.failedTargets.push({
            deviceIP: deviceIP,
            error: data.message || '上传失败'
          })
        }
      }
    }
    
    const allDevicesDone = runningUploadTask.deviceIps.every(ip => {
      const dev = runningUploadTask.deviceProgress[ip]
      return dev && (dev.completed > 0 || dev.failed > 0)
    })
    
    if (allDevicesDone) {
      let successCount = 0
      let failCount = 0
      
      runningUploadTask.deviceIps.forEach(ip => {
        const dev = runningUploadTask.deviceProgress[ip]
        if (dev) {
          if (dev.completed > 0) successCount++
          if (dev.failed > 0) failCount++
        }
      })
      
      runningUploadTask.completed = successCount
      runningUploadTask.failed = failCount
      runningUploadTask.progress = 100
      
      if (failCount === 0) {
        runningUploadTask.status = 'completed'
        runningUploadTask.endTime = new Date()
        if (currentUploadImage.value) {
          ElMessage.success(`镜像 ${currentUploadImage.value.name} 上传成功`)
          imageUploadStatus.value.set(currentUploadImage.value.url, true)
        }
      } else {
        runningUploadTask.status = 'failed'
        runningUploadTask.endTime = new Date()
        ElMessage.warning(`镜像 ${currentUploadImage.value?.name || '镜像'} 上传部分失败，成功 ${successCount} 个，失败 ${failCount} 个`)
      }
    } else {
      // 更新总进度（基于已完成的设备数）
      const doneCount = runningUploadTask.deviceIps.filter(ip => {
        const dev = runningUploadTask.deviceProgress[ip]
        return dev && (dev.completed > 0 || dev.failed > 0)
      }).length
      runningUploadTask.progress = Math.round((doneCount / runningUploadTask.deviceIps.length) * 100)
    }
  } finally {
    // 只有当没有运行中的任务时才重置状态
    const stillRunning = taskQueue.value.some(t => t.type === 'uploadImage' && t.status === 'running')
    if (!stillRunning) {
      isUploadingImage.value = false
      uploadProgress.value = 0
      currentUploadImage.value = null
    }
  }
}

// 显示设备密码设置对话框


// 设置设备密码
const handleSetPassword = async () => {
  if (!activeDevice.value) return
  
  if (!passwordForm.value.password) {
    ElMessage.error('请输入密码')
    return
  }
  
  try {
    // 检查设备是否需要认证
    if (activeDevice.value.version === 'v3') {
      try {
        // 先尝试无密码访问，看是否需要认证
        const result = await getContainers(activeDevice.value, null);
        // 如果认证失败，说明设备已经设置了密码，需要先认证
        if (result && result.code === 61 && result.message === 'Authentication Failed') {
          // 使用authRetry处理认证
          await authRetry(activeDevice.value, async (currentPassword) => {
            passwordLoading.value = true
            const result = await setDevicePassword(activeDevice.value, passwordForm.value.password, currentPassword)
            if (result.success) {
              ElMessage.success('密码设置成功')
              passwordDialogVisible.value = false
              // 保存新密码到本地存储并同步到后端
              await saveDevicePassword(activeDevice.value.ip, passwordForm.value.password)
            } else {
              ElMessage.error(`密码设置失败: ${result.message}`)
            }
          })
          return
        }
      } catch (error) {
        // 无密码访问失败，需要认证
        await authRetry(activeDevice.value, async (currentPassword) => {
          passwordLoading.value = true
          const result = await setDevicePassword(activeDevice.value, passwordForm.value.password, currentPassword)
          if (result.success) {
            ElMessage.success('密码设置成功')
            passwordDialogVisible.value = false
            // 保存新密码到本地存储并同步到后端
            await saveDevicePassword(activeDevice.value.ip, passwordForm.value.password)
          } else {
            ElMessage.error(`密码设置失败: ${result.message}`)
          }
        })
        return
      }
    }
    
    // 不需要认证或已经认证成功，直接设置密码
    passwordLoading.value = true
    // 获取当前保存的密码
    const currentPassword = getDevicePassword(activeDevice.value.ip);
    const result = await setDevicePassword(activeDevice.value, passwordForm.value.password, currentPassword)
    if (result.success) {
      ElMessage.success('密码设置成功')
      passwordDialogVisible.value = false
      // 保存新密码到本地存储并同步到后端
      await saveDevicePassword(activeDevice.value.ip, passwordForm.value.password)
    } else {
      ElMessage.error(`密码设置失败: ${result.message}`)
    }
  } catch (error) {
    console.error('设置密码失败:', error)
    ElMessage.error(`设置密码失败: ${error.message}`)
  } finally {
    passwordLoading.value = false
  }
}

// 显示密码设置对话框
const showPasswordDialog = () => {
  if (!activeDevice.value) {
    ElMessage.error('请先选择设备')
    return
  }
  
  passwordForm.value.password = ''
  passwordDialogVisible.value = true
}

// 关闭设备密码
const handleClosePassword = async () => {
  if (!activeDevice.value) return
  
  try {
    const confirmResult = await ElMessageBox.confirm('确定要关闭设备密码吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 用户确认关闭密码
    passwordLoading.value = true
    
    // 获取当前保存的密码
    const currentPassword = getDevicePassword(activeDevice.value.ip);
    const result = await closeDevicePassword(activeDevice.value, currentPassword)
    
    if (result.success) {
      ElMessage.success('密码关闭成功')
      // 清除本地存储的密码
      removeDevicePassword(activeDevice.value.ip)
    } else {
      ElMessage.error(`密码关闭失败: ${result.message}`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('关闭密码失败:', error)
      ElMessage.error(`关闭密码失败: ${error.message}`)
    }
  } finally {
    passwordLoading.value = false
  }
}


// 清理磁盘数据
const handleCleanDisk = async () => { 
   try {
        await ElMessageBox.confirm(
            `确定要清理设备"${activeDevice.value.ip}"磁盘数据吗？`,
            '清理磁盘数据会重启设备，重启时间为5~10分钟，请耐心等待',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        collapseRightSidebar()

        const taskId = `cleanDisk_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        const task = {
            id: taskId,
            type: 'cleanDisk',
            deviceIP: activeDevice.value.ip,
            status: 'running',
            progress: 0,
            startTime: new Date(),
            completed: 0,
            failed: 0,
            totalSteps: 6,
            currentStep: 0,
            steps: []
        }
        taskQueue.value.unshift(task)

        // 🔧 使用 authRetry 处理认证
        await authRetry(activeDevice.value, async (password) => {
            let headers = {}
            if (password) {
                const auth = btoa(`admin:${password}`)
                headers['Authorization'] = `Basic ${auth}`
            }

            const response = await fetch(
                `http://${getDeviceAddr(activeDevice.value.ip)}/server/device/reset`,
                {
                    method: 'POST',
                    headers: headers
                }
            )

            if (!response.ok) {
                // 🔧 如果是401错误，直接抛出让authRetry处理，不标记任务失败
                if (response.status === 401) {
                    throw new Error('Authentication Failed')
                }
                
                // 其他错误才标记任务失败
                const currentTask = taskQueue.value.find(t => t.id === taskId)
                if (currentTask) {
                    currentTask.status = 'failed'
                    currentTask.endTime = new Date()
                    currentTask.error = `接口请求失败: ${response.status}`
                }
                
                ElMessage.error('清理失败')
                return
            }

            const reader = response.body.getReader()
            const decoder = new TextDecoder()
            let buffer = ''
            let taskCompleted = false

            while (!taskCompleted) {
                const { done, value } = await reader.read()
                if (done) break

                buffer += decoder.decode(value, { stream: true })
                const lines = buffer.split('\n')
                buffer = lines.pop()

                for (const line of lines) {
                    if (line.trim()) {
                        console.log('SSE返回:', line)
                        
                        const currentTask = taskQueue.value.find(t => t.id === taskId)
                        if (!currentTask || taskCompleted) continue

                        const stepMatch = line.match(/\[STEP\s+(\d+)\]/i)
                        const infoMatch = line.match(/\[INFO\]/)
                        const resetMatch = line.match(/Reset sequence completed/i)
                        const rebootMatch = line.match(/Rebooting/i)

                        if (stepMatch) {
                            currentTask.currentStep = parseInt(stepMatch[1])
                            currentTask.progress = Math.round((currentTask.currentStep / currentTask.totalSteps) * 100)
                            currentTask.steps.push(line)
                        } else if (infoMatch) {
                            currentTask.steps.push(line)
                        }

                        if (resetMatch || rebootMatch) {
                            currentTask.progress = 100
                            currentTask.status = 'completed'
                            currentTask.completed = 1
                            currentTask.endTime = new Date()
                            taskCompleted = true
                            
                            ElMessage.success('设备清理完成，正在重启')
                            
                            isViewingDeviceDetails.value = false
                        }
                    }
                }
            }
            
            // 🔧 检查流结束时任务状态
            const finalTask = taskQueue.value.find(t => t.id === taskId)
            if (finalTask && !taskCompleted) {
                // 流已结束但没有收到完成标记，可能是最后一行数据在buffer中
                if (buffer.trim()) {
                    console.log('SSE最后一行:', buffer)
                    const resetMatch = buffer.match(/Reset sequence completed/i)
                    const rebootMatch = buffer.match(/Rebooting/i)
                    
                    if (resetMatch || rebootMatch) {
                        finalTask.progress = 100
                        finalTask.status = 'completed'
                        finalTask.completed = 1
                        finalTask.endTime = new Date()
                        taskCompleted = true
                        
                        ElMessage.success('设备清理完成，正在重启')
                        isViewingDeviceDetails.value = false
                    }
                }
                
                // 如果仍未完成，检查是否至少接收到了步骤信息
                if (!taskCompleted && finalTask.currentStep >= finalTask.totalSteps) {
                    console.log('所有步骤已完成，标记任务为成功')
                    finalTask.progress = 100
                    finalTask.status = 'completed'
                    finalTask.completed = 1
                    finalTask.endTime = new Date()
                    
                    ElMessage.success('设备清理完成，正在重启')
                    isViewingDeviceDetails.value = false
                }
            }
        })
    } catch (error) {
        if (error !== 'cancel') {
            console.error('清理失败:', error)
            ElMessage.error('清理失败，请检查网络连接')
        }
    }
}

// 显示认证对话框（批量模式）
const showAuthDialog = (device, callback) => {
  console.log(`[认证] 收到认证请求: ${device.ip}`)
  
  // 检查该设备是否已经在批量认证列表中
  const existingIndex = batchAuthDevices.value.findIndex(item => item.device.ip === device.ip)
  if (existingIndex !== -1) {
    console.log(`[认证] 设备 ${device.ip} 已在批量认证列表中，更新回调函数`)
    batchAuthDevices.value[existingIndex].callback = callback
    return
  }
  
  // 添加到批量认证列表
  batchAuthDevices.value.push({
    device,
    callback,
    password: '',          // 密码输入框
    savePassword: true,    // 是否保存密码
    status: 'pending',     // pending: 等待输入, verifying: 验证中, success: 成功, failed: 失败
    errorMsg: ''           // 错误信息
  })
  
  console.log(`[认证] 当前批量认证列表长度: ${batchAuthDevices.value.length}`)
  
  // 清除之前的定时器
  if (batchAuthCollectTimeout.value) {
    clearTimeout(batchAuthCollectTimeout.value)
  }
  
  // 延迟200ms收集所有设备，避免设备陆续添加时频繁弹窗
  batchAuthCollectTimeout.value = setTimeout(() => {
    if (batchAuthDevices.value.length > 0) {
      console.log(`[认证] 📋 开始批量认证，共 ${batchAuthDevices.value.length} 个设备`)
      batchAuthDialogVisible.value = true
    }
  }, 200)
}


// 显示授权同步对话框
const showSyncAuthDialog = () => {
  const syncAuthDeviceStr = localStorage.getItem('syncAuthCredentials')
  console.log('syncAuthDeviceStr:', syncAuthDeviceStr)
  if(syncAuthDeviceStr) {
    try {
      const syncAuthDevice = JSON.parse(syncAuthDeviceStr)
      syncAuthForm.value = {
        username: syncAuthDevice.username || '',
        password: syncAuthDevice.password || '',
        saveCredentials: syncAuthDevice.saveCredentials || false
      }
    } catch (error) {
      console.error('Failed to parse syncAuthCredentials:', error)
      syncAuthForm.value = {
        username: '',
        password: '',
        saveCredentials: false
      }
    }
  } else {
      syncAuthForm.value = {
      username: '',
      password: '',
      saveCredentials: false
    }
  }
  syncAuthDialogVisible.value = true
}

// 处理授权过期（错误码3030）
const handleAuthExpired = async () => {
  console.log('授权过期，错误码3030')
  
  // 检查本地存储是否有记住的凭证
  const savedCredentials = localStorage.getItem('syncAuthCredentials')
  
  if (savedCredentials) {
    try {
      const credentials = JSON.parse(savedCredentials)
      if (credentials.saveCredentials) {
        console.log('发现已保存的凭证，自动重新登录')
        // 静默处理，不打扰用户（自动重新登录在后台进行）
        // ElMessage.warning(t('auth.loginExpiredAutoRelogin'))
        
        // 使用保存的凭证重新登录
        const params = {
          username: credentials.username,
          password: CryptoJS.MD5(credentials.password).toString(),
          _ts: new Date().getTime(),
        }
        
        const sortedKeys = Object.keys(params).sort()
        const paramStr = sortedKeys.map(k => `${params[k]}`).join('#') + '#' + '454&*&*fsdff'
        const sign = CryptoJS.MD5(paramStr).toString()
        const formData = new URLSearchParams()
        formData.append('type', 'login')
        
        const data = {
          uname: params.username,
          pwd: params.password,
          _ts: params._ts,
          _sign: sign
        }
        formData.append('data', JSON.stringify(data))
        
        const response = await fetch('https://moyunteng.com/api/sp_api.php', {
          method: 'POST',
          body: formData
        })
        const result = await response.json()
        
        if (result.code == 200) {
          console.log('自动重新登录成功')
          // 静默处理成功，不显示提示
          // ElMessage.success('自动重新登录成功')
          token.value = result.data.token
          uname.value = result.data.uname
          localStorage.setItem('token', result.data.token)
          localStorage.setItem('uname', result.data.uname)
          // 重新获取设备绑定状态
          fetchDeviceBindStatus()
        } else {
          console.error('自动重新登录失败:', result.msg)
          // 失败时才显示错误提示
          ElMessage.error('登录已过期，请重新登录')
          clearAuthAndRefresh()
        }
      } else {
        clearAuthAndRefresh()
      }
    } catch (error) {
      console.error('解析保存的凭证失败:', error)
      clearAuthAndRefresh()
    }
  } else {
    clearAuthAndRefresh()
  }
}

// 清空登录信息并刷新页面
const clearAuthAndRefresh = () => {
  ElMessage.warning('登录已过期，请重新登录')
  token.value = null
  uname.value = null
  localStorage.removeItem('token')
  localStorage.removeItem('uname')
  localStorage.removeItem('syncAuthCredentials')
  // 刷新页面
  window.location.reload()
}

// 处理授权同步提交
const handleSyncAuthSubmit = async () => {
  if (!syncAuthForm.value.username || !syncAuthForm.value.password) {
    ElMessage.warning(proxy.$i18n.t('common.enterUsernameAndPassword'))
    return
  }

  syncAuthLoading.value = true
  try {
    // 准备签名参数
    const params = {
      username: syncAuthForm.value.username,
      password: CryptoJS.MD5(syncAuthForm.value.password).toString(),
      _ts: new Date().getTime(),
    }

    // const data = {
    //   uname: params.username,
    //   pwd: CryptoJS.MD5(params.password).toString(),
    //   _ts: new Date().getTime(),
    // }

    const sortedKeys = Object.keys(params).sort();
    const paramStr = sortedKeys.map(k => `${params[k]}`).join('#') + '#' + '454&*&*fsdff';
    console.log('paramStr:', paramStr)
    const sign = CryptoJS.MD5(paramStr).toString();
    const formData = new URLSearchParams();
    formData.append('type', 'login');
    console.log('data:', sign)
    // 将除 type 外的所有参数放入 data 对象中
    const data = {
        uname: params.username,
        pwd: params.password,
        _ts: params._ts,
        _sign: sign
    }
    // 将 data 转为字符串格式传入
    formData.append('data', JSON.stringify(data));
    
    try { 
        // 模拟请求，实际项目中请使用fetch或其他HTTP库进行请求
       const response = await fetch('https://moyunteng.com/api/sp_api.php', {
          method: 'POST',
          body: formData
       })
       const result = await response.json()
       console.log('response:', result)
      if (result.code == 200) {
        ElMessage.success(proxy.$i18n.t('common.loginSuccess'))
        token.value = result.data.token
        uname.value = result.data.uname
        localStorage.setItem('token', result.data.token)
        localStorage.setItem('uname', result.data.uname)
        localStorage.setItem('uid', result.data.uid)
        // 登录成功后启动同步授权定时器
        startSyncAuthTimer()
        fetchDeviceBindStatus()
       } else{
        ElMessage.error(result.msg)
       }
     } catch (error) {
      console.error(error)
      ElMessage.error(error)
    }
    
    
    if (syncAuthForm.value.saveCredentials) {
      // 保存凭证到本地存储
      localStorage.setItem('syncAuthCredentials', JSON.stringify({
        username: syncAuthForm.value.username,
        password: syncAuthForm.value.password,
        saveCredentials: syncAuthForm.value.saveCredentials
      }))
    } else {
      // 未勾选记住凭证，清除旧的保存凭证
      localStorage.removeItem('syncAuthCredentials')
    }
    syncAuthDialogVisible.value = false
  } catch (error) {
    console.error(proxy.$i18n.t('common.syncAuthFailed') + ':', error)
    ElMessage.error(proxy.$i18n.t('common.syncAuthFailed') + ': ' + error.message)
  } finally {
    syncAuthLoading.value = false
  }
}

// 处理授权同步取消
const handleSyncAuthCancel = () => {
  syncAuthDialogVisible.value = false
}

// 处理用户信息更新（从子组件接收）
const handleUpdateUserInfo = (userInfo) => {
  if (userInfo.token) {
    token.value = userInfo.token
  }
  if (userInfo.uname) {
    uname.value = userInfo.uname
  }
  if (userInfo.uid) {
    // uid 存储在 localStorage 中，已经在子组件中处理
  }
  // 登录成功后启动同步授权定时器
  startSyncAuthTimer()
  // 重新获取设备绑定状态
  fetchDeviceBindStatus()
}

// 打开注册对话框
const openRegisterDialog = () => {
  syncAuthDialogVisible.value = false
  registerDialogVisible.value = true
  registerForm.value = {
    phone: '',
    password: '',
    confirmPassword: '',
    vcode: '',
    vkey: ''
  }
}

// 发送验证码
const sendVcode = async () => {
  if (!registerForm.value.phone) {
    ElMessage.warning('请输入手机号')
    return
  }
  
  // 验证手机号格式
  const phoneReg = /^1[3-9]\d{9}$/
  if (!phoneReg.test(registerForm.value.phone)) {
    ElMessage.warning(proxy.$i18n.t('common.enterCorrectPhone'))
    return
  }
  
  sendVcodeLoading.value = true
  try {
    const result = await GetPhoneVCode(registerForm.value.phone, token.value || '')
    console.log('获取验证码结果:', result)
    if (result.code == 200) {
      // 保存vkey供注册时使用
      if (result.data && result.data.vkey) {
        registerForm.value.vkey = result.data.vkey
      }
      ElMessage.success(proxy.$i18n.t('common.vcodeSentSuccess'))
      // 开始倒计时
      vcodeCountdown.value = 60
      vcodeTimer.value = setInterval(() => {
        vcodeCountdown.value--
        if (vcodeCountdown.value <= 0) {
          clearInterval(vcodeTimer.value)
          vcodeTimer.value = null
        }
      }, 1000)
    } else {
      ElMessage.error(result.msg || proxy.$i18n.t('common.sendCode') + '失败')
    }
  } catch (error) {
    console.error('发送验证码失败:', error)
    ElMessage.error(proxy.$i18n.t('common.sendCode') + '失败: ' + error.message)
  } finally {
    sendVcodeLoading.value = false
  }
}

// 处理注册提交
const handleRegisterSubmit = async () => {
  if (!registerForm.value.phone) {
    ElMessage.warning(proxy.$i18n.t('common.enterPhone'))
    return
  }
  if (!registerForm.value.password) {
    ElMessage.warning(proxy.$i18n.t('common.enterPassword'))
    return
  }
  if (!registerForm.value.confirmPassword) {
    ElMessage.warning(proxy.$i18n.t('common.enterConfirmPassword'))
    return
  }
  if (registerForm.value.password !== registerForm.value.confirmPassword) {
    ElMessage.warning(proxy.$i18n.t('common.passwordMismatch'))
    return
  }
  if (!registerForm.value.vcode) {
    ElMessage.warning(proxy.$i18n.t('common.enterVCode'))
    return
  }
  if (!registerForm.value.vkey) {
    ElMessage.warning(proxy.$i18n.t('common.getVCodeFirst'))
    return
  }

  registerLoading.value = true
  try {
    const result = await Register(
      registerForm.value.phone,
      registerForm.value.password,
      registerForm.value.vcode,
      registerForm.value.vkey
    )
    
    if (result.code === 200 || result.code === 0) {
      ElMessage.success(proxy.$i18n.t('common.registerSuccess'))
      registerDialogVisible.value = false
      // 注册成功后自动填充登录表单
      syncAuthForm.value.username = registerForm.value.phone
      syncAuthForm.value.password = registerForm.value.password
      syncAuthDialogVisible.value = true
    } else {
      ElMessage.error(result.message || result.msg || proxy.$i18n.t('common.registerFailed'))
    }
  } catch (error) {
    console.error(proxy.$i18n.t('common.registerFailed') + ':', error)
    ElMessage.error(proxy.$i18n.t('common.registerFailed') + ': ' + error.message)
  } finally {
    registerLoading.value = false
  }
}

// 处理注册取消
const handleRegisterCancel = () => {
  registerDialogVisible.value = false
  // 清除倒计时
  if (vcodeTimer.value) {
    clearInterval(vcodeTimer.value)
    vcodeTimer.value = null
    vcodeCountdown.value = 0
  }
}

// 忘记密码：打开弹窗
const openForgotPasswordDialog = () => {
  forgotPasswordForm.value = { phone: '', newPassword: '', confirmPassword: '', vcode: '', vkey: '' }
  forgotPasswordErrors.value = { phone: '', newPassword: '', confirmPassword: '' }
  fpVcodeCountdown.value = 0
  if (fpVcodeTimer.value) {
    clearInterval(fpVcodeTimer.value)
    fpVcodeTimer.value = null
  }
  forgotPasswordDialogVisible.value = true
}

// 忘记密码：关闭弹窗
const handleForgotPasswordClose = () => {
  forgotPasswordDialogVisible.value = false
  if (fpVcodeTimer.value) {
    clearInterval(fpVcodeTimer.value)
    fpVcodeTimer.value = null
    fpVcodeCountdown.value = 0
  }
}

// 忘记密码：获取验证码
const sendForgotPasswordVcode = async () => {
  forgotPasswordErrors.value.phone = ''
  if (!forgotPasswordForm.value.phone) {
    forgotPasswordErrors.value.phone = '手机号码不能为空'
    return
  }
  const phoneReg = /^1[3-9]\d{9}$/
  if (!phoneReg.test(forgotPasswordForm.value.phone)) {
    forgotPasswordErrors.value.phone = '请输入正确的手机号'
    return
  }
  fpVcodeLoading.value = true
  try {
    const result = await GetPhoneVCode(forgotPasswordForm.value.phone, token.value || '')
    if (result.code == 200 || result.code == 0) {
      forgotPasswordForm.value.vkey = result.data && result.data.vkey ? result.data.vkey : ''
      ElMessage.success('验证码已发送')
      fpVcodeCountdown.value = 60
      fpVcodeTimer.value = setInterval(() => {
        fpVcodeCountdown.value--
        if (fpVcodeCountdown.value <= 0) {
          clearInterval(fpVcodeTimer.value)
          fpVcodeTimer.value = null
        }
      }, 1000)
    } else {
      ElMessage.error(result.msg || '发送验证码失败')
    }
  } catch (error) {
    console.error('发送验证码失败:', error)
    ElMessage.error('发送验证码失败: ' + error.message)
  } finally {
    fpVcodeLoading.value = false
  }
}

// 忘记密码：验证码按钮文字
const fpVcodeButtonText = computed(() => {
  return fpVcodeCountdown.value > 0 ? `${fpVcodeCountdown.value}s后重试` : '获取验证码'
})

// 忘记密码：是否倒计时中
const fpIsCountingDown = computed(() => fpVcodeCountdown.value > 0)

// 忘记密码：提交重置
const handleForgotPasswordSubmit = async () => {
  forgotPasswordErrors.value = { phone: '', newPassword: '', confirmPassword: '' }
  let hasError = false
  if (!forgotPasswordForm.value.phone) {
    forgotPasswordErrors.value.phone = '手机号码不能为空'
    hasError = true
  } else if (!/^1[3-9]\d{9}$/.test(forgotPasswordForm.value.phone)) {
    forgotPasswordErrors.value.phone = '请输入正确的手机号'
    hasError = true
  }
  if (!forgotPasswordForm.value.newPassword) {
    forgotPasswordErrors.value.newPassword = '新密码不能为空'
    hasError = true
  }
  if (!forgotPasswordForm.value.confirmPassword) {
    forgotPasswordErrors.value.confirmPassword = '确认新密码不能为空'
    hasError = true
  } else if (forgotPasswordForm.value.newPassword !== forgotPasswordForm.value.confirmPassword) {
    forgotPasswordErrors.value.confirmPassword = '两次输入的密码不一致'
    hasError = true
  }
  if (!forgotPasswordForm.value.vcode) {
    ElMessage.warning('请输入验证码')
    hasError = true
  }
  if (!forgotPasswordForm.value.vkey) {
    ElMessage.warning('请先获取验证码')
    hasError = true
  }
  if (hasError) return

  forgotPasswordLoading.value = true
  try {
    const resp = await fetch('https://www.moyunteng.com/api/api.php', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        type: 'find_pwd',
        data: JSON.stringify({
          uname: forgotPasswordForm.value.phone,
          pwd: CryptoJS.MD5(forgotPasswordForm.value.newPassword).toString(),
          vcode: forgotPasswordForm.value.vcode,
          vkey: forgotPasswordForm.value.vkey
        })
      })
    })
    const result = await resp.json()
    if (result.code == 0 || result.code == 200) {
      ElMessage.success('密码重置成功，请重新登录')
      handleForgotPasswordClose()
      syncAuthDialogVisible.value = true
    } else {
      ElMessage.error(result.msg || '密码重置失败')
    }
  } catch (error) {
    console.error('密码重置失败:', error)
    ElMessage.error('密码重置失败: ' + error.message)
  } finally {
    forgotPasswordLoading.value = false
  }
}

// 忘记密码弹窗内跳转注册
const openRegisterFromForgot = () => {
  handleForgotPasswordClose()
  openRegisterDialog()
}

// 处理批量认证提交
const handleBatchAuthSubmit = async () => {
  // 检查是否所有设备都输入了密码
  const devicesWithoutPassword = batchAuthDevices.value.filter(item => !item.password)
  if (devicesWithoutPassword.length > 0) {
    ElMessage.warning(`还有 ${devicesWithoutPassword.length} 个设备未输入密码`)
    return
  }
  
  batchAuthLoading.value = true
  
  try {
    console.log(`[认证] 🚀 开始批量认证 ${batchAuthDevices.value.length} 个设备`)
    
    // 并行验证所有设备
    const authPromises = batchAuthDevices.value.map(async (item) => {
      item.status = 'verifying'
      
      try {
        // 验证密码是否正确
        await getContainers(item.device, item.password)
        
        // 认证成功
        item.status = 'success'
        console.log(`[认证] ✅ 设备 ${item.device.ip} 认证成功`)
        
        // 保存密码到本地存储并同步到后端
        if (item.savePassword) {
          await saveDevicePassword(item.device.ip, item.password)
        }
        
        // 执行回调函数
        if (item.callback) {
          try {
            await item.callback(item.password)
          } catch (error) {
            console.error(`[认证] 回调执行失败 (${item.device.ip}):`, error)
          }
        }
        
        return { device: item.device.ip, success: true }
      } catch (error) {
        // 认证失败
        item.status = 'failed'
        item.errorMsg = '密码错误'
        console.error(`[认证] ❌ 设备 ${item.device.ip} 认证失败:`, error)
        return { device: item.device.ip, success: false, error: '密码错误' }
      }
    })
    
    // 等待所有认证完成
    const results = await Promise.all(authPromises)
    
    // 统计结果
    const successCount = results.filter(r => r.success).length
    const failCount = results.filter(r => !r.success).length
    
    console.log(`[认证] 📊 批量认证完成: 成功 ${successCount} 个, 失败 ${failCount} 个`)
    
    if (failCount === 0) {
      // 全部成功
      ElMessage.success(`所有设备认证成功 (${successCount} 个)`)
      batchAuthDialogVisible.value = false
      batchAuthDevices.value = []
    } else if (successCount === 0) {
      // 全部失败
      ElMessage.error('所有设备认证失败，请检查密码')
    } else {
      // 部分成功
      ElMessage.warning(`${successCount} 个设备认证成功，${failCount} 个失败`)
      // 移除成功的设备，保留失败的继续输入
      batchAuthDevices.value = batchAuthDevices.value.filter(item => item.status === 'failed')
    }
  } catch (error) {
    console.error('[认证] 批量认证过程出错:', error)
    ElMessage.error('批量认证失败')
  } finally {
    batchAuthLoading.value = false
  }
}

// 处理批量认证取消
const handleBatchAuthCancel = () => {
  console.log(`[认证] ⚠️ 用户取消批量认证，共 ${batchAuthDevices.value.length} 个设备`)
  
  batchAuthDialogVisible.value = false
  batchAuthDevices.value = []
  
  ElMessage.info('已取消设备认证')
}

// 获取V3设备SDK版本信息
const getV3SDKVersion = async (deviceIP) => {
  try {
    // 尝试使用已保存的密码
    const savedPassword = getDevicePassword(deviceIP);
    let headers = {};
    
    if (savedPassword) {
      const auth = btoa(`admin:${savedPassword}`);
      headers = {
        'Authorization': `Basic ${auth}`
      };
    }
    
    const response = await axios.get(`http://${getDeviceAddr(deviceIP)}/api/v1/device/info`, {
      headers: headers,
      timeout: 5000
    });
    
    if (response.data.code === 0 && response.data.data) {
      const version = response.data.data.sdkVersion || response.data.data.version || '0';
      console.log('获取V3设备SDK版本成功:', version);
      return version;
    }
    return '0';
  } catch (error) {
    console.error('获取V3设备SDK版本失败:', error);
    return '0';
  }
}

// 认证重试函数
const authRetry = async (device, callback) => {
  try {
    console.log('认证重试 - 开始认证流程')
    console.log('认证重试 - 设备信息:', device)
    
    // 检查设备是否需要认证
    // v0-v2 设备不需要认证
    if (device.version === 'v0' || device.version === 'v1' || device.version === 'v2') {
      console.log('认证重试 - v0-v2设备，不需要认证')
      return await callback(null);
    }
    
    // 对于V3设备，智能认证流程（优化：优先使用保存的密码，避免无谓的401）
    if (device.version === 'v3') {
      console.log('认证重试 - V3设备，检查认证策略')
      
      // 【优化】先检查是否有保存的密码
      const savedPassword = getDevicePassword(device.ip)
      
      if (savedPassword) {
        console.log('认证重试 - 发现已保存的密码，直接使用认证')
        try {
          // 直接使用保存的密码，避免先401再重试
          const result = await callback(savedPassword)
          console.log('认证重试 - 使用保存密码成功')
          return result
        } catch (error) {
          console.log('认证重试 - 保存的密码失败:', error)
          
          // 检查是否是认证错误
          if ((error.response && error.response.status === 401) || error.message === 'Authentication Failed') {
            console.log('认证重试 - 保存的密码已失效，需要重新认证')
            
            // 密码失效，显示认证对话框
            return new Promise((resolve, reject) => {
              console.log('认证重试 - 显示认证对话框')
              showAuthDialog(device, async (password) => {
                try {
                  console.log('认证重试 - 收到认证对话框的密码')
                  
                  // 保存新密码
                  await saveDevicePassword(device.ip, password)
                  console.log('认证重试 - 新密码已保存')
                  
                  // 执行原始回调
                  const result = await callback(password)
                  console.log('认证重试 - 回调执行成功')
                  resolve(result)
                } catch (error) {
                  console.error('认证重试 - 回调执行失败:', error)
                  reject(error)
                }
              })
            })
          } else {
            // 不是401错误，直接抛出
            console.log('认证重试 - 不是401错误，直接抛出')
            throw error
          }
        }
      } else {
        console.log('认证重试 - 没有保存的密码，尝试无密码访问')
        
        try {
          // 没有保存密码，尝试无密码访问
          const result = await callback(null)
          console.log('认证重试 - 无密码访问成功')
          return result
        } catch (error) {
          console.log('认证重试 - 无密码访问失败:', error)
          
          // 检查是否是认证错误
          if ((error.response && error.response.status === 401) || error.message === 'Authentication Failed') {
            console.log('认证重试 - 需要认证，显示认证对话框')
            
            // 显示认证对话框
            return new Promise((resolve, reject) => {
              showAuthDialog(device, async (password) => {
                try {
                  console.log('认证重试 - 收到认证对话框的密码')
                  
                  // 保存密码
                  await saveDevicePassword(device.ip, password)
                  console.log('认证重试 - 密码已保存')
                  
                  // 执行原始回调
                  const result = await callback(password)
                  console.log('认证重试 - 回调执行成功')
                  resolve(result)
                } catch (error) {
                  console.error('认证重试 - 回调执行失败:', error)
                  reject(error)
                }
              })
            })
          } else {
            // 不是401错误，直接抛出
            console.log('认证重试 - 不是401错误，直接抛出')
            throw error
          }
        }
      }
    }
    
    // 默认返回空密码
    console.log('认证重试 - 默认返回空密码')
    return await callback(null);
  } catch (error) {
    console.error('认证重试失败:', error)
    throw error
  }
}



// 显示设备详情弹窗
const showDeviceDetails = async (device) => {
  console.log('显示设备详情弹窗:', device);
  await handleDeviceSelect(device);
  deviceDetailsDialogVisible.value = true;
  isViewingDeviceDetails.value = true;
  
  console.log('根据当前选中的镜像分类自动获取镜像列表:', selectedImageCategory.value);
  await switchImageCategory(selectedImageCategory.value);
  
  await fetchDeviceDetailCloudMachines();

   // 🔧 移除定时查询，使用心跳机制自动更新设备信息
  // 设备信息会随心跳自动更新，不需要前端定时刷新

}


// 切换镜像分类


// 树形结构选中节点
const treeSelectedKeys = ref([]) // 存储树形结构选中的节点ID

// 备份列表相关
const backupListVisible = ref(false) // 备份列表显示状态
const backupTableRef = ref(null) // 备份列表表格引用
const backupCurrentSlot = ref(0) // 当前操作的坑位（备份相关）
const switchingBackupSlot = ref(null) // 当前正在切换备份的坑位
const backupList = ref([]) // 备份列表数据
const selectedBackupList = ref([]) // 选中的备份列表（用于批量操作）
const backupGroups = ref(['默认分组', '测试分组', '生产分组']) // 备份分组
const selectedBackupGroup = ref('默认分组') // 当前选中的分组
const sortBy = ref('createTime') // 排序字段

// 批量切换备份进度对话框
const batchSwitchBackupProgressVisible = ref(false) // 进度对话框显示状态
const batchSwitchBackupProgressList = ref([]) // 每条进度项: { slotNum, currentName, backupName, status: 'pending'|'running'|'success'|'failed', message }
const batchSwitchBackupTotal = ref(0) // 总数
const batchSwitchBackupDone = ref(0) // 已完成数（成功+失败）
const sortOrder = ref('descending') // 排序顺序

// 初始化备份列表数据
const initBackupList = () => {
  // 保存当前选中的备份ID列表
  const selectedIds = selectedBackupList.value.map(item => item.id)
  
  // 显示当前坑位的所有容器，除了当前运行的那个
  const slotContainers = allInstances.value.filter(inst => 
    inst.indexNum === backupCurrentSlot.value && 
    inst.name // 排除空容器
  );
  
  // 直接映射容器数据，如果没有则为空数组
  backupList.value = slotContainers.map((container, index) => ({
    id: container.name,
    name: container.name,
    createTime: container.created,
    remark: container.image,
    group: '默认分组',
    status: container.status
  }));
  
  // 恢复选中状态：使用表格的toggleRowSelection方法
  if (selectedIds.length > 0 && backupTableRef.value) {
    nextTick(() => {
      // 清空当前选中
      backupTableRef.value.clearSelection()
      // 重新选中之前选中的行
      backupList.value.forEach(row => {
        if (selectedIds.includes(row.id)) {
          backupTableRef.value.toggleRowSelection(row, true)
        }
      })
    })
  }
}

// 显示备份列表
const showBackupList = async (slotNum) => {
  backupCurrentSlot.value = slotNum
  // 清空之前的选中状态
  selectedBackupList.value = []
  
  // 先显示弹窗，并显示加载状态
  backupListVisible.value = true
  backupLoading.value = true
  
  try {
    // 先用缓存数据初始化，避免空白
    initBackupList()
    
    // 实时刷新数据
    if (activeDevice.value) {
      await triggerAndroidRefresh([activeDevice.value.ip])
      // 数据刷新后重新初始化列表
      initBackupList()
    }
  } catch (error) {
    console.error('刷新备份列表失败:', error)
  } finally {
    backupLoading.value = false
  }
}

// 处理备份列表选中状态变化
const handleBackupSelectionChange = (selection) => {
  selectedBackupList.value = selection
}

// 显示创建云机对话框


// 查找可用的坑位，参考api/main.go中的findAvailableIdx实现
const findAvailableSlot = (device, startSlot = 1, count = 1) => {
  if (!device) return -1
  
  // 根据设备型号确定最大坑位数
  let maxSlots = 12 // 默认12个坑位
  if (device.name && device.name.toLowerCase().includes('p1')) {
    maxSlots = 24 // P1型号24个坑位
  }
  
  // 获取当前设备的所有容器
  const deviceContainers = deviceCloudMachinesCache.value.get(device.ip) || []
  
  // 创建已使用坑位的集合，只考虑运行中的容器
  const usedSlots = new Set()
  deviceContainers.forEach(machine => {
    if (machine.status === 'running' && machine.indexNum) {
      usedSlots.add(machine.indexNum)
    }
  })
  
  // 检查当前运行中的容器数量是否已达到上限
  if (usedSlots.size >= maxSlots) {
    return -1
  }
  
  // 检查请求的坑位数是否超过剩余可用坑位数
  const remainingSlots = maxSlots - usedSlots.size
  if (count > remainingSlots) {
    return -1
  }
  
  // 查找从startSlot开始，连续count个可用的坑位
  for (let i = 1; i <= maxSlots; i++) {
    // 如果是批量创建，检查从i开始的连续count个坑位是否都可用
    if (count > 1) {
      let allAvailable = true
      for (let j = 0; j < count; j++) {
        // 检查当前坑位是否超出范围或已被使用
        if (i + j > maxSlots || usedSlots.has(i + j)) {
          allAvailable = false
          break
        }
      }
      if (allAvailable) {
        return i
      }
    } else {
      // 单个坑位：直接返回第一个可用的坑位
      if (!usedSlots.has(i)) {
        return i
      }
    }
  }
  
  // 没有找到可用的坑位
  return -1
}

// 检查指定坑位是否已被占用
const isSlotOccupied = (device, slot) => {
  if (!device) return true
  
  // 获取当前设备的所有容器
  const deviceContainers = deviceCloudMachinesCache.value.get(device.ip) || []
  
  // 检查是否有运行中的容器占用该坑位
  return deviceContainers.some(machine => 
    machine.status === 'running' && machine.indexNum === slot
  )
}

// 检查并关闭指定坑位的运行中容器
const checkAndStopContainer = async (device, slot) => {
  if (!device) return true
  
  // 获取当前设备的所有容器
  const deviceContainers = deviceCloudMachinesCache.value.get(device.ip) || []
  
  // 查找指定坑位的运行中容器
  const runningContainer = deviceContainers.find(machine => 
    machine.status === 'running' && machine.indexNum === slot
  )
  
  if (runningContainer) {
    // 有运行中的容器，需要先关闭
    try {
      ElMessage.info(`正在关闭坑位 ${slot} 上运行的容器: ${runningContainer.name}`)
      await authRetry(device, async (password) => {
          await stopContainer(device, runningContainer.name, password)
        })
      // 等待1秒，确保容器完全关闭
      await new Promise(resolve => setTimeout(resolve, 1000))
      return true
    } catch (error) {
      ElMessage.error(`关闭容器失败：${error.message}`)
      return false
    }
  }
  
  // 没有运行中的容器，直接返回
  return true
}

// 创建云机
const createCloudMachine = async (device, slot, modelName, cancelCheck = null, options = {}) => {
  if (!device) {
    ElMessage.error('设备不能为空')
    return false
  }

  if (typeof cancelCheck === 'function' && cancelCheck()) {
    throw new Error('任务已取消')
  }
  
  // 优先从 options.formOverride 读取参数，避免后台批量任务干扰界面表单
  const form = options.formOverride ? { ...createForm.value, ...options.formOverride } : createForm.value
  
  try {
    // 获取镜像信息
    let imageUrl = ''
    // 检查是否是本地镜像
    const isLocalImage = form.imageCategory === 'local'
    
    if (isLocalImage) {
      // 本地镜像
      imageUrl = form.localImageUrl
      if (!imageUrl) {
        ElMessage.error('请选择本地镜像')
        return false
      }
    } else {
      // 在线镜像
      imageUrl = form.imageSelect
      if (form.imageSelect === 'custom') {
        imageUrl = form.customImageUrl
      }
    }
    
    // 根据设备版本选择创建方式
    if (device.version === 'v3') {
      // 如果是在线机型，需要选择机型
      if (form.modelType === 'online' && !modelName) {
        ElMessage.error('请选择机型')
        return false
      }
      // V3 API：使用8000端口
      createLoading.value = true
      await createV3CloudMachine(device, slot, modelName, cancelCheck, options)
      createLoading.value = false
    } else {
      // V0-V2 API：根据镜像来源执行不同的逻辑
      // 1. 显示蒙版
      sdkLoadingVisible.value = true
      
      // 2. 处理镜像
      if (isLocalImage) {
        // 本地镜像：先检查设备上是否已存在该镜像
        sdkLoadingMessage.value = '检查设备上的镜像...'
        
        try {
          // 获取设备上的镜像列表

          const savedPassword = getDevicePassword(device.ip)
          const deviceImages = await GetImages(device.ip, device.version, savedPassword || '')

          console.log('设备上的镜像列表:', deviceImages)
          
          // 从本地镜像路径中提取镜像名称
          const imagePathParts = imageUrl.split('\\')
          let localImageName = imagePathParts[imagePathParts.length - 1]
          localImageName = localImageName.replace('.tar.gz', '')
          localImageName = localImageName.toLowerCase().replace(/[^a-z0-9_-]/g, '_')
          const expectedImageName = `local/${localImageName}:latest`
          
          // 检查设备上是否已存在该镜像
          let imageExists = false
          if (Array.isArray(deviceImages)) {
            // V0-V2 格式：数组
            imageExists = deviceImages.some(img => {
              const repoTags = img.RepoTags || img.imageUrl || img.Image
              if (Array.isArray(repoTags)) {
                return repoTags.includes(expectedImageName)
              }
              return repoTags === expectedImageName
            })
          } else if (deviceImages.list) {
            // V3 格式：包含list字段的对象
            imageExists = deviceImages.list.some(img => {
              return img.imageUrl === expectedImageName || img.Image === expectedImageName
            })
          }
          
          console.log('检查镜像是否存在:', expectedImageName, '结果:', imageExists)
          
          if (!imageExists) {
            // 设备上不存在该镜像，需要推送
            sdkLoadingMessage.value = '正在推送本地镜像到设备...'
            
            // 重置进度
            sdkLoadingProgress.value = 0
            
            // 调用后端的LoadImageToDevice函数，将本地镜像加载到设备
            console.log('调用后端LoadImageToDevice函数，镜像URL:', imageUrl)
            const password = getDevicePassword(device.ip)
            const loadResult = await LoadImageToDevice(device.ip, imageUrl, device.version, password || '')
            console.log('推送本地镜像结果:', loadResult)
            
            // 设置进度为100%
            sdkLoadingProgress.value = 100
            
            // 短暂延迟，让用户看到100%进度
            await new Promise(resolve => setTimeout(resolve, 500))
            
            // 检查加载结果
            if (!loadResult.success) {
              ElMessage.error(`推送本地镜像失败：${loadResult.message || '未知错误'}`)
              sdkLoadingVisible.value = false
              return false
            }
            
            // 使用后端返回的真实镜像名称
            if (loadResult.imageName) {
              imageUrl = loadResult.imageName
            } else {
              // 如果后端没有返回镜像名称，使用默认的本地镜像名称格式
              imageUrl = expectedImageName
            }
          } else {
            // 设备上已存在该镜像，直接使用预期的镜像名称
            imageUrl = expectedImageName
            console.log('设备上已存在该镜像，跳过推送步骤')
          }
        } catch (loadError) {
          console.error('处理本地镜像失败:', loadError)
          ElMessage.error(`处理本地镜像失败：${loadError.message}`)
          sdkLoadingVisible.value = false
          return false
        }
      } else {
        // 在线镜像：检查是否需要缓存到本地创建
        const cacheToLocal = createForm.value.cacheToLocal || false
        
        if (cacheToLocal) {
          // 在线镜像缓存到本地创建：先下载到本地，再推送到设备
          sdkLoadingMessage.value = '正在下载镜像到本地...'
          
          // 获取用户数据目录
          const userDataDir = await GetUserDataDir()
          
          // 下载镜像到本地
          try {
            // 重置进度
            sdkLoadingProgress.value = 0
            
            // 启动进度查询定时器
            const progressQueryTimer = setInterval(async () => {
              try {
                const progressResult = await getImagePullProgress()
                if (progressResult && progressResult.progress !== undefined) {
                  sdkLoadingProgress.value = progressResult.progress
                }
              } catch (err) {
                console.error('获取下载进度失败:', err)
              }
            }, 500)
            
            // 调用后端DownloadImage函数
            await DownloadImage(imageUrl)
            
            // 清除进度查询定时器
            clearInterval(progressQueryTimer)
            
            // 设置进度为100%
            sdkLoadingProgress.value = 100
            await new Promise(resolve => setTimeout(resolve, 500))
          } catch (downloadError) {
            console.error('下载镜像到本地失败:', downloadError)
            ElMessage.error(`下载镜像到本地失败：${downloadError.message}`)
            sdkLoadingVisible.value = false
            return false
          }
          
          // 推送本地镜像到设备
          sdkLoadingMessage.value = '正在推送镜像到设备...'
          
          try {
            // 重置进度
            sdkLoadingProgress.value = 0
            
            // 构建正确的本地镜像文件名
            let localImageName = imageUrl.split('/').pop()
            // 替换冒号为下划线，添加.tar.gz后缀
            localImageName = localImageName.replace(':', '_') + '.tar.gz'
            // 调用后端LoadImageToDevice函数
            const password = getDevicePassword(device.ip)
            await LoadImageToDevice(device.ip, `${userDataDir}/${localImageName}`, device.version, password || '')
            
            // 设置进度为100%
            sdkLoadingProgress.value = 100
            await new Promise(resolve => setTimeout(resolve, 500))
          } catch (loadError) {
            console.error('推送镜像到设备失败:', loadError)
            ElMessage.error(`推送镜像到设备失败：${loadError.message}`)
            sdkLoadingVisible.value = false
            return false
          }
        } else {
          // 在线镜像直接创建：使用Docker API拉取镜像
          sdkLoadingMessage.value = '正在下载镜像...'
          
          try {
            // 重置进度
            sdkLoadingProgress.value = 0
            
            // 启动进度查询定时器
            const progressQueryTimer = setInterval(async () => {
              try {
                const progressResult = await getImagePullProgress()
                if (progressResult && progressResult.progress !== undefined) {
                  sdkLoadingProgress.value = progressResult.progress
                }
              } catch (err) {
                console.error('获取镜像拉取进度失败:', err)
              }
            }, 500)
            
            // 调用后端的pullDockerImage函数，开始拉取镜像
            await pullDockerImage(device.ip, imageUrl)
            
            // 清除进度查询定时器
            clearInterval(progressQueryTimer)
            
            // 设置进度为100%
            sdkLoadingProgress.value = 100
            
            // 短暂延迟，让用户看到100%进度
            await new Promise(resolve => setTimeout(resolve, 500))
          } catch (pullError) {
            console.error('拉取镜像失败:', pullError)
            ElMessage.error(`拉取镜像失败：${pullError.message}`)
            sdkLoadingVisible.value = false
            return false
          }
        }
      }
      
      // 3. 镜像处理完成，创建云机
      sdkLoadingMessage.value = '正在创建云机...'
      
      // 4. 调用createV0V2Device函数创建云机
      const createParams = {
        name: createForm.value.name,
        count: createForm.value.count,
        startSlot: createForm.value.startSlot,
        imageSelect: isLocalImage ? imageUrl : createForm.value.imageSelect,
        customImageUrl: isLocalImage ? '' : createForm.value.customImageUrl,
        networkMode: createForm.value.networkMode,
        ipaddr: createForm.value.ipaddr,
        resolution: createForm.value.resolution,
        sandboxSize: createForm.value.sandboxSize,
        dns: createForm.value.dns,
        sandbox: false, // V3才有沙盒选项，V0-V2暂时不使用
        longitud: createForm.value.longitud,
        latitude: createForm.value.latitude,
        PINCode: createForm.value.lockScreenPassword,
        vpcID: createForm.value.vpcNodeId
      }
      
      // 5. 调用后端创建云机
      await createV0V2Device(device, createParams)
      
      // 6. 隐藏蒙版
      sdkLoadingVisible.value = false
    }
    
    // 刷新云机列表
    await fetchAndroidContainers(device, true)
    
    // 保存上一次的镜像选择
    lastImageSelection.value = {
      imageSelect: createForm.value.imageSelect,
      customImageUrl: createForm.value.customImageUrl,
      imageCategory: createForm.value.imageCategory,
      localImageUrl: createForm.value.localImageUrl,
      imageSource: createForm.value.imageSource
    }
    
    ElMessage.success('云机创建成功')
    return true
  } catch (error) {
    console.error('创建云机失败:', error)
    ElMessage.error(`创建云机失败：${error.message}`)
    // 确保蒙版已关闭
    sdkLoadingVisible.value = false
    createLoading.value = false
    throw error // 抛出异常，保留错误信息
  } finally {
    createLoading.value = false
  }
}

// 创建V3云机，参考api/main.go中的handleCreateTask和createV3Container实现
const createV3CloudMachine = async (device, slot, modelName, cancelCheck = null, options = {}) => {
  const isCanceled = () => (typeof cancelCheck === 'function' && cancelCheck()) || createCancelled.value
  // 优先从 options.formOverride 读取参数，避免后台批量任务干扰界面表单
  const form = options.formOverride ? { ...createForm.value, ...options.formOverride } : createForm.value
  // 获取镜像URL，优先使用自定义镜像地址
  let imageUrl = ''
  // 检查是否是本地镜像
  const isLocalImage = form.imageCategory === 'local'
  
  if (isLocalImage) {
    // 本地镜像
    imageUrl = form.localImageUrl
    if (!imageUrl) {
      ElMessage.error('请选择本地镜像')
      return false
    }
    
    // 推送本地镜像到设备
    sdkLoadingVisible.value = true
    sdkLoadingMessage.value = '检查设备上的镜像...'
    
    try {
      // 获取设备上的镜像列表
      const savedPassword = getDevicePassword(device.ip)
      const deviceImages = await GetImages(device.ip, device.version, savedPassword || '')
      console.log('设备上的镜像列表:', deviceImages)
      
      // 获取本地镜像的onlineUrl（从缓存读取）
      let localImageOnlineUrl = ''
      if (localCachedImages.value.length === 0) {
        console.log('[单创建] 本地镜像列表为空，尝试加载...')
        await fetchLocalImages()
      }
      const cachedImage = localCachedImages.value.find(img => img.path === imageUrl)
      if (cachedImage && cachedImage.onlineUrl) {
        localImageOnlineUrl = cachedImage.onlineUrl
      }
      
      // 确定要检查的镜像名称（优先使用onlineUrl，其次使用本地标签）
      let expectedImageName = ''
      let isPushedFromOnline = false
      
      if (localImageOnlineUrl) {
        expectedImageName = localImageOnlineUrl
        isPushedFromOnline = true
        console.log('[单创建] 使用online_url检查镜像:', expectedImageName)
      } else {
        // 从本地镜像路径中提取镜像名称
        const imagePathParts = imageUrl.split('\\')
        let localImageName = imagePathParts[imagePathParts.length - 1]
        localImageName = localImageName.replace('.tar.gz', '')
        localImageName = localImageName.toLowerCase().replace(/[^a-z0-9_-]/g, '_')
        expectedImageName = `local/${localImageName}:latest`
        console.log('[单创建] 未找到online_url，使用本地标签检查:', expectedImageName)
      }
      
      // 检查设备上是否已存在该镜像
      let imageExists = false
      if (Array.isArray(deviceImages)) {
        // V0-V2 格式：数组
        imageExists = deviceImages.some(img => {
          const repoTags = img.RepoTags || img.imageUrl || img.Image
          if (Array.isArray(repoTags)) {
            return repoTags.includes(expectedImageName)
          }
          return repoTags === expectedImageName
        })
      } else if (deviceImages && deviceImages.list) {
        // V3 格式：包含list字段的对象
        imageExists = deviceImages.list.some(img => {
          return img.imageUrl === expectedImageName || img.Image === expectedImageName
        })
      }
      
      console.log('检查镜像是否存在:', expectedImageName, '结果:', imageExists)
      
      if (!imageExists) {
        // 设备上不存在该镜像，需要推送
        sdkLoadingMessage.value = '正在推送本地镜像到设备...'
        
        // 重置进度
        sdkLoadingProgress.value = 0
        
        // 调用后端的LoadImageToDevice函数，将本地镜像加载到设备
        console.log('调用后端LoadImageToDevice函数，镜像URL:', imageUrl)
        const password = getDevicePassword(device.ip)
        const loadResult = await LoadImageToDevice(device.ip, imageUrl, device.version, password || '')
        console.log('推送本地镜像结果:', loadResult)
        
        // 设置进度为100%
        sdkLoadingProgress.value = 100
        
        // 短暂延迟，让用户看到100%进度
        await new Promise(resolve => setTimeout(resolve, 500))
        
        // 隐藏蒙版
        sdkLoadingVisible.value = false
        
        // 检查加载结果
        if (!loadResult.success) {
          ElMessage.error(`推送本地镜像失败：${loadResult.message || '未知错误'}`)
          return false
        }
        
        // 使用后端返回的真实镜像名称
        if (loadResult.imageName) {
          imageUrl = loadResult.imageName
        } else {
          // 如果后端没有返回镜像名称，使用默认的本地镜像名称格式
          imageUrl = expectedImageName
        }
      } else {
        // 设备上已存在该镜像，直接使用预期的镜像名称
        imageUrl = expectedImageName
        console.log('设备上已存在该镜像，跳过推送步骤')
        sdkLoadingVisible.value = false
      }
    } catch (loadError) {
      console.error('处理本地镜像失败:', loadError)
      ElMessage.error(`处理本地镜像失败：${loadError.message}`)
      sdkLoadingVisible.value = false
      return false
    }
  } else {
    // 在线镜像
    imageUrl = form.imageSelect
    if (form.imageSelect === 'custom') {
      imageUrl = form.customImageUrl
    }
    
    // 检查是否需要缓存到本地创建
    const cacheToLocal = form.cacheToLocal || false
    
    if (!cacheToLocal) {
      // 不缓存到本地，直接调用V3 API拉取镜像
      sdkLoadingVisible.value = true
      sdkLoadingMessage.value = '正在拉取镜像...'
      sdkLoadingProgress.value = 0
      
      // 查找当前正在运行的创建任务，用于更新进度
      const currentCreateTask = taskQueue.value.find(t => t.type === 'create' && t.status === 'running')
      
      // 构造拉取镜像的请求参数
      const pullImageParams = {
        imageUrl: imageUrl
      }
      
      // 先获取设备密码
      let password = null
      try {
        // 尝试获取已保存的密码
        password = getDevicePassword(device.ip)
        console.log('使用已保存的密码:', password ? '***' : '无')
      } catch (error) {
        console.error('获取密码失败:', error)
      }
      
      // 构造请求头
      let headers = {}
      if (password) {
        const auth = btoa(`admin:${password}`)
        headers = {
          'Authorization': `Basic ${auth}`
        }
      }
      
      // 调用V3 API拉取镜像
      const pullImageUrl = `http://${getDeviceAddr(device.ip)}/android/pullImage`
      console.log('调用V3 API拉取镜像:', pullImageUrl, pullImageParams)
      
      // 发送POST请求
      console.log('开始发送拉取镜像请求')
      
      // 等待镜像拉取完成
      await new Promise((resolve, reject) => {
        try {
          // 使用fetch API处理流式响应
          
          // 构建请求选项
          const requestOptions = {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              ...headers
            },
            body: JSON.stringify(pullImageParams)
          }
          
          // 发送fetch请求
          fetch(pullImageUrl, requestOptions)
            .then(response => {
              if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`)
              }
              
              // 检查是否支持流式响应
              if (!response.body) {
                throw new Error('Response body is not a readable stream')
              }
              
              // 获取可读流
              const reader = response.body.getReader()
              
              let buffer = ''
              let progress = 0
              
              // 处理流式数据
              const processStream = async () => {
                try {
                  // 检查是否已取消
                  if (isCanceled()) {
                    reader.cancel()
                    throw new Error('创建取消')
                  }
                  
                  // 读取下一个数据块
                  const { done, value } = await reader.read()
                  
                  if (done) {
                    sdkLoadingProgress.value = 100
                    if (currentCreateTask) {
                      currentCreateTask.progress = 100
                    }
                    nextTick()
                    setTimeout(() => {
                      sdkLoadingVisible.value = false
                      resolve()
                    }, 500)
                    return
                  }
                  
                  // 检查是否已取消
                  if (isCanceled()) {
                    reader.cancel()
                    throw new Error('创建取消')
                  }
                  
                  // 处理数据块
                  const chunkStr = new TextDecoder('utf-8').decode(value)
                  buffer += chunkStr
                  
                  // 处理每一行数据
                  const lines = buffer.split('\n')
                  
                  for (let i = 0; i < lines.length; i++) {
                    const line = lines[i].trim()
                    if (line) {
                      if (line.startsWith('data: ')) {
                        const jsonStr = line.substring(6).trim()
                        if (jsonStr) {
                          try {
                            const data = JSON.parse(jsonStr)
                            
                            // 处理进度信息
                            if (data.progressDetail && data.progressDetail.current !== undefined && data.progressDetail.total !== undefined && data.progressDetail.total > 0) {
                              // 计算进度百分比
                              progress = (data.progressDetail.current / data.progressDetail.total) * 100
                              // 直接更新进度
                              sdkLoadingProgress.value = progress
                              if (currentCreateTask) {
                                currentCreateTask.progress = Math.round(progress)
                              }
                              nextTick()
                            } 
                            // 备用进度计算方式
                            else if (data.progress && typeof data.progress === 'string') {
                              // 处理类似 "10%" 的进度字符串
                              const progressMatch = data.progress.match(/(\d+)%/)
                              if (progressMatch) {
                                progress = parseFloat(progressMatch[1])
                                sdkLoadingProgress.value = progress
                                if (currentCreateTask) {
                                  currentCreateTask.progress = Math.round(progress)
                                }
                                nextTick()
                              } else {
                                // 尝试从进度字符串中提取MB/GB信息
                                const sizeMatch = data.progress.match(/(\d+\.\d+)MB\/(\d+\.\d+)GB/)
                                if (sizeMatch) {
                                  const currentMB = parseFloat(sizeMatch[1])
                                  const totalGB = parseFloat(sizeMatch[2])
                                  const totalMB = totalGB * 1024
                                  if (totalMB > 0) {
                                    progress = (currentMB / totalMB) * 100
                                    sdkLoadingProgress.value = progress
                                    if (currentCreateTask) {
                                      currentCreateTask.progress = Math.round(progress)
                                    }
                                    nextTick()
                                  }
                                }
                              }
                            }
                            
                            // 处理状态信息
                            if (data.status) {
                              // 保持基础消息不变，确保进度条持续显示
                              if (data.status.includes('Downloading')) {
                                sdkLoadingMessage.value = '正在拉取镜像... (下载中)'
                              } else if (data.status.includes('Pulling')) {
                                sdkLoadingMessage.value = '正在拉取镜像... (准备中)'
                              } else if (data.status.includes('Extracting')) {
                                sdkLoadingMessage.value = '正在拉取镜像... (解压中)'
                              } else {
                                sdkLoadingMessage.value = `正在拉取镜像... (${data.status})`
                              }
                              nextTick()
                            }
                          } catch (error) {
                            // 尝试直接从原始数据中提取进度信息
                            try {
                              // 尝试匹配类似 "current: 123, total: 456" 的模式
                              const currentMatch = chunkStr.match(/current:\s*(\d+)/)
                              const totalMatch = chunkStr.match(/total:\s*(\d+)/)
                              if (currentMatch && totalMatch) {
                                const current = parseInt(currentMatch[1])
                                const total = parseInt(totalMatch[1])
                                if (total > 0) {
                                  progress = (current / total) * 100
                                  sdkLoadingProgress.value = progress
                                  if (currentCreateTask) {
                                    currentCreateTask.progress = Math.round(progress)
                                  }
                                  nextTick()
                                }
                              }
                            } catch (e) {
                              // 忽略错误
                            }
                          }
                        }
                      }
                    }
                  }
                  
                  // 清空buffer，因为我们已经处理了所有行
                  buffer = ''
                  
                  // 继续处理下一个数据块
                  await processStream()
                } catch (error) {
                  sdkLoadingVisible.value = false
                  reject(error)
                }
              }
              
              // 开始处理流式数据
              processStream()
            })
            .catch(error => {
              sdkLoadingVisible.value = false
              // 检查是否是认证错误
              if (error.message.includes('401')) {
                // 显示认证对话框
                showAuthDialog(device, async (password) => {
                  try {
                    // 保存密码并同步到后端
                    await saveDevicePassword(device.ip, password)
                    
                    // 使用新密码重新拉取镜像
                    const pullImageParams = {
                      imageUrl: imageUrl
                    }
                    
                    // 重置进度条
                    sdkLoadingProgress.value = 0
                    
                    // 查找当前正在运行的创建任务，用于更新进度
                    const currentCreateTask = taskQueue.value.find(t => t.type === 'create' && t.status === 'running')
                    
                    const auth = btoa(`admin:${password}`)
                    const authHeaders = {
                      'Authorization': `Basic ${auth}`,
                      'Content-Type': 'application/json'
                    }
                    
                    // 使用新密码发送请求
                    const response = await fetch(pullImageUrl, {
                      method: 'POST',
                      headers: authHeaders,
                      body: JSON.stringify(pullImageParams)
                    })
                    
                    if (!response.ok) {
                      throw new Error(`HTTP error! status: ${response.status}`)
                    }
                    
                    // 检查是否支持流式响应
                    if (!response.body) {
                      throw new Error('Response body is not a readable stream')
                    }
                    
                    // 获取可读流
                    const reader = response.body.getReader()
                    
                    let buffer = ''
                    
                    // 处理流式数据
                    const processStream = async () => {
                      try {
                        // 检查是否已取消
                        if (isCanceled()) {
                          reader.cancel()
                          throw new Error('创建取消')
                        }
                        
                        // 读取下一个数据块
                        const { done, value } = await reader.read()
                        
                        if (done) {
                          sdkLoadingProgress.value = 100
                          if (currentCreateTask) {
                            currentCreateTask.progress = 100
                          }
                          nextTick()
                          setTimeout(() => {
                            sdkLoadingVisible.value = false
                            resolve()
                          }, 500)
                          return
                        }
                        
                        // 检查是否已取消
                        if (isCanceled()) {
                          reader.cancel()
                          throw new Error('创建取消')
                        }
                        
                        // 处理数据块
                        const chunkStr = new TextDecoder('utf-8').decode(value)
                        buffer += chunkStr
                        
                        // 处理每一行数据
                        const lines = buffer.split('\n')
                        
                        for (let i = 0; i < lines.length; i++) {
                          const line = lines[i].trim()
                          if (line.startsWith('data: ')) {
                            const jsonStr = line.substring(6).trim()
                            if (jsonStr) {
                              try {
                                const data = JSON.parse(jsonStr)
                                
                                if (data.progressDetail && data.progressDetail.current !== undefined && data.progressDetail.total !== undefined && data.progressDetail.total > 0) {
                                  const progress = (data.progressDetail.current / data.progressDetail.total) * 100
                                  sdkLoadingProgress.value = progress
                                  if (currentCreateTask) {
                                    currentCreateTask.progress = Math.round(progress)
                                  }
                                  nextTick()
                                }
                                // 备用进度计算方式
                                else if (data.progress && typeof data.progress === 'string') {
                                  // 处理类似 "10%" 的进度字符串
                                  const progressMatch = data.progress.match(/(\d+)%/)
                                  if (progressMatch) {
                                    const progress = parseFloat(progressMatch[1])
                                    sdkLoadingProgress.value = progress
                                    if (currentCreateTask) {
                                      currentCreateTask.progress = Math.round(progress)
                                    }
                                    nextTick()
                                  } else {
                                    // 尝试从进度字符串中提取MB/GB信息
                                    const sizeMatch = data.progress.match(/(\d+\.\d+)MB\/(\d+\.\d+)GB/)
                                    if (sizeMatch) {
                                      const currentMB = parseFloat(sizeMatch[1])
                                      const totalGB = parseFloat(sizeMatch[2])
                                      const totalMB = totalGB * 1024
                                      if (totalMB > 0) {
                                        const progress = (currentMB / totalMB) * 100
                                        sdkLoadingProgress.value = progress
                                        if (currentCreateTask) {
                                          currentCreateTask.progress = Math.round(progress)
                                        }
                                        nextTick()
                                      }
                                    }
                                  }
                                }
                                
                                if (data.status) {
                                  // 保持基础消息不变，确保进度条持续显示
                                  if (data.status.includes('Downloading')) {
                                    sdkLoadingMessage.value = '正在拉取镜像... (下载中)'
                                  } else if (data.status.includes('Pulling')) {
                                    sdkLoadingMessage.value = '正在拉取镜像... (准备中)'
                                  } else if (data.status.includes('Extracting')) {
                                    sdkLoadingMessage.value = '正在拉取镜像... (解压中)'
                                  } else {
                                    sdkLoadingMessage.value = `正在拉取镜像... (${data.status})`
                                  }
                                  nextTick()
                                }
                              } catch (error) {
                                // 尝试直接从原始数据中提取进度信息
                                try {
                                  // 尝试匹配类似 "current: 123, total: 456" 的模式
                                  const currentMatch = chunkStr.match(/current:\s*(\d+)/)
                                  const totalMatch = chunkStr.match(/total:\s*(\d+)/)
                                  if (currentMatch && totalMatch) {
                                    const current = parseInt(currentMatch[1])
                                    const total = parseInt(totalMatch[1])
                                    if (total > 0) {
                                      const progress = (current / total) * 100
                                      sdkLoadingProgress.value = progress
                                      nextTick()
                                    }
                                  }
                                } catch (e) {
                                  // 忽略错误
                                }
                              }
                            }
                          }
                        }
                        
                        buffer = '' // 清空buffer，因为我们已经处理了所有行
                        
                        // 继续处理下一个数据块
                        await processStream()
                      } catch (error) {
                        sdkLoadingVisible.value = false
                        reject(error)
                      }
                    }
                    
                    // 开始处理流式数据
                    processStream()
                  } catch (error) {
                    sdkLoadingVisible.value = false
                    ElMessage.error(`拉取镜像失败：${error.message}`)
                    reject(error)
                  }
                })
              } else {
                ElMessage.error(`拉取镜像失败：${error.message}`)
                reject(error)
              }
            })
        } catch (error) {
          sdkLoadingVisible.value = false
          reject(error)
        }
      })

    }
  }
  
  // 解析分辨率
  let doboxWidth = ''
  let doboxHeight = ''
  let doboxDpi = ''
  
  if (form.resolution === 'default') {
    // 机型默认分辨率，传递空值
    doboxWidth = ''
    doboxHeight = ''
    doboxDpi = ''
  } else if (form.resolution === 'custom') {
    // 自定义分辨率
    doboxWidth = form.customResolution.width || '720'
    doboxHeight = form.customResolution.height || '1280'
    doboxDpi = form.customResolution.dpi || '320'
  } else {
    // 预设分辨率
    const parts = (form.resolution || '').split('x')
    if (parts.length === 3) {
      doboxWidth = parts[0]
      doboxHeight = parts[1]
      doboxDpi = parts[2]
    } else {
      // 兼容旧格式
      doboxWidth = '720'
      doboxHeight = '1280'
      doboxDpi = '320'
    }
  }
  
  // 从手机型号列表中查找对应的ModelId
  let modelId = ''
  const model = phoneModels.value.find(m => m.name === modelName)
  if (model) {
    modelId = model.id || ''
  } 
  
  // 确保ModelId不为空，使用默认值
  if (!modelId) {
    // 使用默认的ModelId，根据api/main.go中的默认型号设置
    modelId = '17' // 默认型号ID，对应InfinixX6880
  }
  
  // 处理随机机型分配（从按安卓版本过滤后的机型列表中随机选择）
  if (modelName === 'random' || !modelName) {
    const versionFilteredModels = androidVersionFilteredPhoneModels.value
    if (versionFilteredModels && versionFilteredModels.length > 0) {
      // 从按安卓版本过滤后的机型列表中随机选择一个
      const randomModel = versionFilteredModels[Math.floor(Math.random() * versionFilteredModels.length)]
      modelId = randomModel.id
      modelName = randomModel.name
      console.log('随机选择的机型:', modelName, 'ID:', modelId)
    } else if (phoneModels.value && phoneModels.value.length > 0) {
      // 如果过滤后没有机型，从全部机型中随机选择
      const randomModel = phoneModels.value[Math.floor(Math.random() * phoneModels.value.length)]
      modelId = randomModel.id
      modelName = randomModel.name
      console.log('按版本过滤后无机型，从全部机型随机选择:', modelName, 'ID:', modelId)
    } else {
      // 如果没有可用机型，使用默认值
      modelId = '17' // 默认型号ID，对应InfinixX6880
      modelName = 'InfinixX6880' // 默认型号名称
      console.log('没有可用机型，使用默认机型:', modelName, 'ID:', modelId)
    }
  }
  
  // 处理DNS设置
  let dnsValue = form.dns;
  if (form.dns === 'custom' && form.customDns) {
    dnsValue = form.customDns;
  }
  
  // 检查目标坑位的状态，决定Start参数的值
  let shouldStart = true
  
  if (options.start !== undefined) {
    shouldStart = options.start
    console.log(`[createV3CloudMachine] 使用options.start: ${shouldStart}`)
  } else {
    // 从缓存中获取当前设备的容器列表
    const deviceContainers = deviceCloudMachinesCache.value.get(device.ip) || []
    // 查找该坑位是否已有容器
    const existingMachine = deviceContainers.find(m => m.indexNum === slot)
    
    if (existingMachine) {
        if (existingMachine.status === 'running') {
        shouldStart = false
        console.log(`坑位 ${slot} 已有运行中的云机，设置 Start=false`)
        } else {
        shouldStart = true
        console.log(`坑位 ${slot} 云机处于非运行状态，设置 Start=true`)
        }
    }
  }

  // 构造与api/main.go中V3CreateContainerReq完全匹配的参数结构
  // 根据机型类型选择传递的参数
  let params = {}
  if (form.createType === 'container') {
    // 🔧 修复：从 containerDataDiskSize (16G/32G/64G) 提取数字
    const diskSizeNum = parseInt(form.containerDataDiskSize) || 16  // 默认 16GB
    
    params = {
      dns: dnsValue,
      imageUrl: imageUrl,
      name: `${Date.now()}_${slot}_${form.containerName || form.name}${slot}`,
      doboxDpi: doboxDpi,
      doboxFps: '24',
      doboxHeight: doboxHeight,
      doboxWidth: doboxWidth,
      indexNum: slot,
      start: shouldStart,
      sandboxSize: (form.containerSandboxMode === false) ? '' : `${diskSizeNum}GB`,  // ✅ 使用正确的容器数据盘大小
      // offset: 0
    }

    // 容器模式添加网卡配置
    if (form.containerNetworkCardType === 'public' && form.containerMacVlanIp) {
      // 公有网卡(MacVlan)配置
      params.macVlanIp = form.containerMacVlanIp
    } else if (form.containerNetworkCardType === 'private' && form.containerMytBridgeName) {
      // 私有网卡(myt_bridge)配置
      params.mytBridgeName = form.containerMytBridgeName
    }

    // 添加VPC网络管理配置
    if (form.vpcGroupId) {
      params.VpcGroupId = form.vpcGroupId
      // 参考随机机型实现，在发送请求时才真正随机选择节点
      if (form.vpcSelectMode === 'random') {
        params.VpcID = getRandomVpcNodeId()
      } else {
        params.VpcID = form.vpcNodeId || ''
      }
    }
  } else {
    params = {
      Name: `${Date.now()}_${slot}_${form.name}${slot}`, // 格式：timestamp_idx_nameidx
      IndexNum: slot,
      ImageUrl: imageUrl,
      SandboxSize: (form.sandboxMode === false) ? '' : `${form.sandboxSize}GB`,
      Dns: dnsValue,
      CountryCode: form.countryCode, // 机型国家代码
      // S5代理设置（SDK版本>=25时支持）
      S5Type: form.s5Type,
      S5IP: form.s5IP,
      S5Port: form.s5Port,
      S5User: form.s5User,
      S5Password: form.s5Password,
      DoboxFps: '24', // 默认24FPS
      DoboxWidth: doboxWidth,
      DoboxHeight: doboxHeight,
      DoboxDpi: doboxDpi,
      LocateIp: '', // 定位IP，暂时为空
      Longitude: form.longitud, // 经度
      Latitude: form.latitude, // 纬度
      start: shouldStart,
      Mgenable: form.enableMagisk ? '1' : '0', // 0-关，1-开
      Gmsenable: form.enableGMS ? '1' : '0', // 0-关，1-开
      enforce: form.enforce !== false, // 安全模式，默认开启
      adbPort: (form.enforce !== false && form.adbPort !== undefined) ? Number(form.adbPort) : 0, // ADB端口，安全模式下生效，0不开启ADB
      PINCode: form.lockScreenPassword, // 锁屏密码
      randomFile: form.randomFile || false, // 随机系统文件
     // VpcID: form.vpcNodeId || '', // VPC节点ID
      mytBridgeName: form.mytBridgeName,
      macVlanIp: form.macVlanIp,
    }
    
    // 根据机型类型添加对应的参数
    if (form.modelType === 'local') {
      // 本地机型：使用 LocalModel 参数
      params.LocalModel = form.localModel || ''
      console.log('使用本地机型参数:', params.LocalModel)
    } else if(form.modelType === 'online') {
      // 在线机型：使用 modelId 和 modelName 参数
      params.ModelId = modelId
      params.ModelName = modelName
      console.log('使用在线机型参数:', params.ModelId, params.ModelName)
    } else {
        params.modelStatic = form.modelStatic
    }
    
    // 如果是独立IP模式，添加网络配置
    if (form.networkMode === 'myt' && form.ipaddr) {
      // 直接使用表单中的IP地址
      params.Network = {
        Ip: form.ipaddr,
        Gw: '', // 后端会处理网络配置
        Subnet: '' // 后端会处理网络配置
      }
    }
    
    // 添加VPC网络管理配置
    if (form.vpcGroupId) {
      params.VpcGroupId = form.vpcGroupId
      // 参考随机机型实现，在发送请求时才真正随机选择节点
      if (form.vpcNodeId === 'random') {
        params.VpcID = getRandomVpcNodeId()
      } else {
        params.VpcID = form.vpcNodeId || ''
      }
    }
  }
  
  // 直接使用axios调用设备的V3 API，与api/main.go中的createV3Container实现一致
  let apiUrl = `http://${getDeviceAddr(device.ip)}/android`
  if (form.createType === 'container') {
    apiUrl = `http://${getDeviceAddr(device.ip)}/androidV2`
  }
  console.log('调用V3 API创建容器:', apiUrl, params)
  
  // 尝试使用已保存的密码
  const savedPassword = getDevicePassword(device.ip)
  let headers = {}
  
  if (savedPassword) {
    // 添加认证头
    const auth = btoa(`admin:${savedPassword}`)
    headers = {
      'Authorization': `Basic ${auth}`
    }
  }
  
  try {
    // 发送POST请求到设备的8000端口/android端点
    const response = await axios.post(apiUrl, params, {
      headers: headers
    })
    console.log('V3 API创建容器成功，返回数据:', response.data)
    
    // 检查响应状态
    if (response.data.code !== 0) {
      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        // 认证失败，显示认证对话框
        return new Promise((resolve, reject) => {
          showAuthDialog(device, async (password) => {
            try {
              const auth = btoa(`admin:${password}`)
              const authResponse = await axios.post(apiUrl, params, {
                headers: {
                  'Authorization': `Basic ${auth}`
                }
              })
              
              console.log('V3 API创建容器成功，返回数据:', authResponse.data)
              
              // 检查响应状态
              if (authResponse.data.code !== 0) {
                throw new Error(`创建失败: ${authResponse.data.message || '未知错误'}`)
              }
              
              resolve(authResponse.data)
            } catch (error) {
              console.error('创建容器失败:', error)
              reject(error)
            }
          })
        })
      } else {
        throw new Error(`创建失败: ${response.data.message || '未知错误'}`)
      }
    }
    
    return response.data
  } catch (error) {
    if (error.response && error.response.data && error.response.data.code === 61 && error.response.data.message === 'Authentication Failed') {
      // 认证失败，显示认证对话框
      return new Promise((resolve, reject) => {
        showAuthDialog(device, async (password) => {
          try {
            const auth = btoa(`admin:${password}`)
            const authResponse = await axios.post(apiUrl, params, {
              headers: {
                'Authorization': `Basic ${auth}`
              }
            })
            
            console.log('V3 API创建容器成功，返回数据:', authResponse.data)
            
            // 检查响应状态
            if (authResponse.data.code !== 0) {
              throw new Error(`创建失败: ${authResponse.data.message || '未知错误'}`)
            }
            
            resolve(authResponse.data)
          } catch (error) {
            console.error('创建容器失败:', error)
            reject(error)
          }
        })
      })
    } else {
      throw error
    }
  }
}

// 处理创建表单提交
const handleCreateSubmit = async () => {
  // 保存选中的坑位到localStorage（用于下次记忆）
  if (createMode.value === 'batch' || createMode.value === 'multi-device-batch') {
    const slots = createForm.value.selectedSlots || []
    localStorage.setItem('createDialog_selectedSlots', JSON.stringify(slots))
    console.log('已保存坑位选择:', slots)
  }
  
  // IP计算辅助函数
  const calculateSingleIp = (startIp, offset) => {
    if (!startIp || offset === 0) return startIp
    try {
      const parts = startIp.split('.').map(Number)
      if (parts.length !== 4 || parts.some(isNaN)) return startIp
      
      let [a, b, c, d] = parts
      d += offset
      
      // 处理进位逻辑
      while (d > 255) { d -= 256; c += 1; }
      while (c > 255) { c -= 256; b += 1; }
      while (b > 255) { b -= 256; a += 1; }
      
      if (a > 255) return startIp
      return `${a}.${b}.${c}.${d}`
    } catch (e) {
      return startIp
    }
  }

  // 重置取消标志
  createCancelled.value = false

  // 校验容器名称：不允许包含下划线
  const chineseRegex = /[\u4e00-\u9fa5]/
  if (chineseRegex.test(createForm.value.name) || createForm.value.name.includes('_')) {
    ElMessage.error('云机名称不允许包含中文和特殊字符_')
    return
  }
  
  // 准备目标设备列表
  let targetDevices = []
  if (createMode.value === 'multi-device-batch') {
    targetDevices = devices.value.filter(d => selectedBatchDevices.value.includes(d.ip))
    if (targetDevices.length === 0) {
      ElMessage.warning('请选择设备')
      return
    }
  } else if (createDevice.value) {
    targetDevices = [createDevice.value]
  }

  // 校验必须选择坑位
  if ((createMode.value === 'batch' || createMode.value === 'multi-device-batch') && (!createForm.value.selectedSlots || createForm.value.selectedSlots.length === 0)) {
    ElMessage.warning('请选择坑位')
    return
  }

  let globalIpIndex = 0
  
  // 批量创建模式下的容器模式逻辑
  if (createForm.value.createType === 'container') {
    let imageUrl = createForm.value.containerImageSelect
    if (createForm.value.containerAndroidVersion === 'custom' || imageUrl === 'custom') {
        imageUrl = createForm.value.containerCustomImageUrl
        if (!imageUrl) {
            ElMessage.error('请输入自定义镜像地址')
            return
        }
    } else if (!imageUrl) {
      ElMessage.error('请选择镜像地址')
      return
    }
    
    createLoading.value = true
    try {
      const { 
        containerName: name, 
        containerCount: count, 
        startSlot, 
        containerAndroidVersion: androidVersion, 
        containerResolution: resolution, 
        containerDns: dns, 
        containerCustomDns: customDns, 
        containerSandboxMode: sandboxMode, 
        containerDataDiskSize: dataDiskSize, 
        containerCustomResolution: customResolution 
      } = createForm.value
      
      // 校验自定义分辨率
      if (resolution === 'custom') {
        if (!customResolution.width || !customResolution.height || !customResolution.dpi) {
           ElMessage.error('自定义分辨率时，设备宽、设备长和DPI都必须填写')
           createLoading.value = false
           return
        }
      }

      const targets = []
      
      for (const device of targetDevices) {
        // 容器模式：支持新批量模式（指定坑位）
        if (createMode.value === 'batch' && createForm.value.selectedSlots && createForm.value.selectedSlots.length > 0) {
            for (const slot of createForm.value.selectedSlots) {
                // Check if slot has running container
                const hasRunning = isSlotOccupied(device, slot)

                for (let k = 0; k < count; k++) {
                    // Determine start status
                    // 0对应边框为蓝色，1会黄色，2为红色，无实例对应灰色。
                    // 判断逻辑：
                    // 1. 未登录时：判断当前坑位是否有开机状态的云机，如果有则关机，否则开机
                    // 2. 已登录时：
                    //    - 未绑定(0)或被绑定(2)：判断当前坑位是否有开机状态的云机，如果有则关机，否则开机
                    //    - 已绑定(1)：根据坑位状态判断
                    //      * 状态0或1：如果有开机云机则关机，否则开机
                    //      * 状态2或无实例：默认关机
                    let shouldStart = false
                    if (k === 0) {
                       const isLoggedIn = !!token.value
                       const bindStatus = deviceBindStatus.value.get(device.id) || 0
                       const slotInfo = slotStates.value[slot]
                       const state = slotInfo ? slotInfo.state : undefined
                       
                       // 未登录或已登录但未绑定/被绑定时：直接根据是否有运行中的云机判断
                       if (!isLoggedIn || bindStatus === 0 || bindStatus === 2) {
                           shouldStart = !hasRunning
                       } else {
                           // 已登录且已绑定(1)：根据坑位状态判断
                           if (state === 0 || state === 1) {
                               shouldStart = !hasRunning
                           } else {
                               // state === 2 (Expired) or undefined (No Instance)
                               shouldStart = false
                           }
                       }
                    }

                    // 计算当前实例的 IP
                    const currentMacVlanIp = calculateSingleIp(
                        createForm.value.containerMacVlanIp || createForm.value.macVlanIp, 
                        globalIpIndex
                    )
            
                    targets.push({
                        createType: 'container', // 标记为容器创建
                        slot: slot,
                        start: shouldStart,
                        modelName: 'random', 
                        modelType: 'online',
                        deviceIp: device.ip,
                        deviceVersion: device.version,
                        deviceId: device.id,
                        imageUrl: imageUrl,
                        isLocalImage: false,
                        
                        // Container specific fields
                        name: name ? `${name}_${slot}` : undefined,
                        androidVersion,
                        resolution,
                        customResolution: resolution === 'custom' ? { ...customResolution } : undefined,
                        dns,
                        customDns,
                        sandboxMode,
                        dataDiskSize,
                        
                        vpcGroupId: createForm.value.vpcGroupId || '',
                        vpcNodeId: createForm.value.vpcSelectMode === 'random' ? 'random' : createForm.value.vpcNodeId,
                        mytBridgeName: createForm.value.mytBridgeName,
                        macVlanIp: currentMacVlanIp, // 优先使用容器模式的 MacVlan IP
                        networkCardType: createForm.value.containerNetworkCardType // 网卡类型
                    })
                    globalIpIndex++
                }
            }
        } else {
            // 旧模式：连续分配
            let availableSlot = startSlot
            
            for (let i = 0; i < count; i++) {
              // 自动查找下一个可用坑位
              while (isSlotOccupied(device, availableSlot)) {
                availableSlot++
              }
              
              // 计算当前实例的 IP
              const currentMacVlanIp = calculateSingleIp(
                createForm.value.containerMacVlanIp || createForm.value.macVlanIp, 
                globalIpIndex
              )
      
              targets.push({
                createType: 'container', // 标记为容器创建
                slot: availableSlot,
                modelName: 'random', 
                modelType: 'online',
                deviceIp: device.ip,
                deviceVersion: device.version,
                deviceId: device.id,
                imageUrl: imageUrl,
                isLocalImage: false,
                
                // Container specific fields
                name: name ? `${name}_${availableSlot}` : undefined,
                androidVersion,
                resolution,
                customResolution: resolution === 'custom' ? { ...customResolution } : undefined,
                dns,
                customDns,
                sandboxMode,
                dataDiskSize,
                
                vpcGroupId: createForm.value.vpcGroupId || '',
                vpcNodeId: createForm.value.vpcSelectMode === 'random' ? 'random' : createForm.value.vpcNodeId,
                mytBridgeName: createForm.value.mytBridgeName,
                macVlanIp: currentMacVlanIp, // 优先使用容器模式的 MacVlan IP
                networkCardType: createForm.value.containerNetworkCardType // 网卡类型
              })
              availableSlot++
              globalIpIndex++
            }
        }
      }
      
      const taskId = addTaskToQueue('create', targets)
      executeTask(taskId)
      
      createDialogVisible.value = false
      createLoading.value = false
      ElMessage.success(`已添加 ${targets.length} 个云机创建任务到队列`)
    } catch (e) {
      console.error(e)
      createLoading.value = false
      ElMessage.error('创建任务失败')
    }
    return
  }

  // V3设备需要检查机型
  if (createDevice.value.version === 'v3') {
    if (createForm.value.modelType === 'online') {
      // 在线机型：需要选择机型
      if (!createForm.value.modelName || createForm.value.modelName === undefined) {
        ElMessage.error('请选择机型')
        return
      }
    } else if (createForm.value.modelType === 'backup') {
      // 备份机型：需要选择机型
      if (!createForm.value.modelStatic) {
        ElMessage.error('请选择备份机型')
        return
      }
    }
    // 本地机型：不需要选择机型，localModel 可选填
  }

  // 检查锁屏密码规则：4到8位纯数字
  const lockScreenPassword = createForm.value.lockScreenPassword
  if (lockScreenPassword) {
    const passwordRegex = /^\d{4,8}$/
    if (!passwordRegex.test(lockScreenPassword)) {
      ElMessage.error('锁屏密码必须是4到8位纯数字')
      return
    }
  }

  createLoading.value = true
  try {
    if (createMode.value === 'batch' || createMode.value === 'multi-device-batch') {
      // 批量创建 - 使用任务队列系统，支持无限制创建
      const { modelName, count, startSlot, modelType, localModel, modelStatic, selectedSlots } = createForm.value
      
      // 准备批量创建目标，自动分配可用坑位
      const targets = []
      
      for (const device of targetDevices) {
        // 新批量模式：指定坑位 + 单坑位数量
        if (createMode.value === 'batch' && selectedSlots && selectedSlots.length > 0) {
            for (const slot of selectedSlots) {
                // Check if slot has running container
                const hasRunning = isSlotOccupied(device, slot)
                
                for (let k = 0; k < count; k++) {
                     // Determine start status
                     // 0对应边框为蓝色，1会黄色，2为红色，无实例对应灰色。
                     // 判断逻辑：
                     // 1. 未登录时：判断当前坑位是否有开机状态的云机，如果有则关机，否则开机
                     // 2. 已登录时：
                     //    - 未绑定(0)或被绑定(2)：判断当前坑位是否有开机状态的云机，如果有则关机，否则开机
                     //    - 已绑定(1)：根据坑位状态判断
                     //      * 状态0或1：如果有开机云机则关机，否则开机
                     //      * 状态2或无实例：默认关机
                     let shouldStart = false
                     if (k === 0) {
                        const isLoggedIn = !!token.value
                        const bindStatus = deviceBindStatus.value.get(device.id) || 0
                        const slotInfo = slotStates.value[slot]
                        const state = slotInfo ? slotInfo.state : undefined
                        
                        // 未登录或已登录但未绑定/被绑定时：直接根据是否有运行中的云机判断
                        if (!isLoggedIn || bindStatus === 0 || bindStatus === 2) {
                            shouldStart = !hasRunning
                        } else {
                            // 已登录且已绑定(1)：根据坑位状态判断
                            if (state === 0 || state === 1) {
                                shouldStart = !hasRunning
                            } else {
                                // state === 2 (Expired) or undefined (No Instance)
                                shouldStart = false
                            }
                        }
                     }

                     // 获取本地镜像的 onlineUrl（从缓存读取）
                     let localImageOnlineUrl = ''
                     if (createForm.value.imageCategory === 'local' && createForm.value.localImageUrl) {
                         if (localCachedImages.value.length === 0) {
                             console.log(`[批量创建] 本地镜像列表为空，尝试加载...`)
                             await fetchLocalImages()
                         }
                         const cachedImage = localCachedImages.value.find(img => img.path === createForm.value.localImageUrl)
                         if (cachedImage && cachedImage.onlineUrl) {
                           localImageOnlineUrl = cachedImage.onlineUrl
                         }
                     }
                     
                     // 计算当前实例的 IP
                     const currentMacVlanIp = calculateSingleIp(createForm.value.macVlanIp, globalIpIndex)

                     // Randomize model selection if 'random' is selected（从按安卓版本过滤后的机型中随机）
                     let targetModelName = modelName
                     if (modelType === 'online' && modelName === 'random') {
                        const vfModels = androidVersionFilteredPhoneModels.value
                        const pool = vfModels && vfModels.length > 0 ? vfModels : phoneModels.value
                        if (pool.length > 0) {
                          const randomIndex = Math.floor(Math.random() * pool.length)
                          targetModelName = pool[randomIndex].name
                        }
                     }
            
                     let targetLocalModel = localModel
                     if (modelType === 'local' && (localModel === 'random' || localModel === '') && localPhoneModels.value.length > 0) {
                        const randomIndex = Math.floor(Math.random() * localPhoneModels.value.length)
                        targetLocalModel = localPhoneModels.value[randomIndex].name
                     }

                     let targetModelStatic = modelStatic
                     if (modelType === 'backup' && modelStatic === 'random' && backupPhoneModels.value.length > 0) {
                        const randomIndex = Math.floor(Math.random() * backupPhoneModels.value.length)
                        targetModelStatic = backupPhoneModels.value[randomIndex].name
                     }

                     targets.push({
                        slot: slot,
                        start: shouldStart,
                        modelName: targetModelName,
                        modelType: modelType,
                        localModel: targetLocalModel,
                        modelStatic: targetModelStatic,
                        deviceIp: device.ip,
                        deviceVersion: device.version,
                        deviceId: device.id,
                        imageUrl: createForm.value.imageSelect === 'custom' ? createForm.value.customImageUrl : createForm.value.imageSelect,
                        isLocalImage: createForm.value.imageCategory === 'local',
                        localImageUrl: createForm.value.localImageUrl,
                        localImageOnlineUrl: localImageOnlineUrl,
                        vpcGroupId: createForm.value.vpcGroupId || '',
                        vpcNodeId: createForm.value.vpcSelectMode === 'random' ? 'random' : createForm.value.vpcNodeId,
                        mytBridgeName: createForm.value.mytBridgeName,
                        macVlanIp: currentMacVlanIp
                     })
                     
                     globalIpIndex++
                }
            }
        } else {
            let availableSlot = startSlot
            
            for (let i = 0; i < count; i++) {
              // 自动查找下一个可用坑位
              while (isSlotOccupied(device, availableSlot)) {
                availableSlot++
              }
              
              // 获取本地镜像的 onlineUrl（从缓存读取）
              let localImageOnlineUrl = ''
              if (createForm.value.imageCategory === 'local' && createForm.value.localImageUrl) {
                // 确保本地镜像列表已加载
                if (localCachedImages.value.length === 0) {
                  console.log(`[批量创建] 本地镜像列表为空，尝试加载...`)
                  await fetchLocalImages()
                }
                
                const cachedImage = localCachedImages.value.find(img => img.path === createForm.value.localImageUrl)
                if (cachedImage && cachedImage.onlineUrl) {
                  localImageOnlineUrl = cachedImage.onlineUrl
                  console.log(`[批量创建] 从缓存获取online_url: ${localImageOnlineUrl}`)
                } else {
                  console.warn(`[批量创建] 未找到本地镜像的online_url: ${createForm.value.localImageUrl}`)
                }
              }
              
              // 计算当前实例的 IP
              const currentMacVlanIp = calculateSingleIp(createForm.value.macVlanIp, globalIpIndex)
      
              // Randomize model selection if 'random' is selected（从按安卓版本过滤后的机型中随机）
              let targetModelName = modelName
              if (modelType === 'online' && modelName === 'random') {
                const vfModels = androidVersionFilteredPhoneModels.value
                const pool = vfModels && vfModels.length > 0 ? vfModels : phoneModels.value
                if (pool.length > 0) {
                  const randomIndex = Math.floor(Math.random() * pool.length)
                  targetModelName = pool[randomIndex].name
                }
              }
      
              let targetLocalModel = localModel
              if (modelType === 'local' && (localModel === 'random' || localModel === '') && localPhoneModels.value.length > 0) {
                const randomIndex = Math.floor(Math.random() * localPhoneModels.value.length)
                targetLocalModel = localPhoneModels.value[randomIndex].name
              }
      
              let targetModelStatic = modelStatic
              if (modelType === 'backup' && modelStatic === 'random' && backupPhoneModels.value.length > 0) {
                const randomIndex = Math.floor(Math.random() * backupPhoneModels.value.length)
                targetModelStatic = backupPhoneModels.value[randomIndex].name
              }
      
              targets.push({
                slot: availableSlot,
                modelName: targetModelName,
                modelType: modelType,
                localModel: targetLocalModel,
                modelStatic: targetModelStatic,
                deviceIp: device.ip,
                deviceVersion: device.version,
                deviceId: device.id,
                imageUrl: createForm.value.imageSelect === 'custom' ? createForm.value.customImageUrl : createForm.value.imageSelect,
                isLocalImage: createForm.value.imageCategory === 'local',
                localImageUrl: createForm.value.localImageUrl,
                localImageOnlineUrl: localImageOnlineUrl, // 本地镜像对应的在线地址
                // 网络管理参数
                vpcGroupId: createForm.value.vpcGroupId || '',
                vpcNodeId: createForm.value.vpcSelectMode === 'random' ? 'random' : createForm.value.vpcNodeId,
                mytBridgeName: createForm.value.mytBridgeName,
                macVlanIp: currentMacVlanIp // macVlan IP
              })
              availableSlot++
              globalIpIndex++
            }
        }
      }
      
      // 添加到任务队列
      const taskId = addTaskToQueue('create', targets)
      executeTask(taskId)
      
      // 关闭创建对话框，不影响已添加的任务
      createDialogVisible.value = false
      createLoading.value = false
      ElMessage.success(`已添加 ${targets.length} 个云机创建任务到队列`)
    } else {
      // 单个坑位创建：使用任务队列系统
      const { modelName } = createForm.value
      
      // 检查是否已取消
      if (createCancelled.value) {
        ElMessage.info('创建取消')
        return
      }
      
      // 检查指定坑位是否可用
      const deviceContainers = deviceCloudMachinesCache.value.get(createDevice.value.ip) || []
      const usedSlots = new Set()
      deviceContainers.forEach(machine => {
        if (machine.status === 'running' && machine.indexNum) {
          usedSlots.add(machine.indexNum)
        }
      })
      
      // 检查当前运行中的容器数量是否已达到上限
      let maxSlots = 12
      if (createDevice.value.name && createDevice.value.name.toLowerCase().includes('p1')) {
        maxSlots = 24
      }
      
      if (usedSlots.size >= maxSlots) {
        ElMessage.error(`${createDevice.value.name}型号最大只能运行${maxSlots}个云机`)
        createLoading.value = false
        return
      }

      // 检查指定坑位是否已被占用
      // if (usedSlots.has(currentSlot.value)) {
      //   ElMessage.error(`坑位 ${currentSlot.value} 已被占用，请选择其他坑位`)
      //   createLoading.value = false
      //   return
      // }

      // 准备单个创建目标，添加到任务队列
      // Randomize model selection if 'random' is selected
      // let targetModelName = modelName
      // if (createForm.value.modelType === 'online' && modelName === 'random' && phoneModels.value.length > 0) {
      //   const randomIndex = Math.floor(Math.random() * phoneModels.value.length)
      //   targetModelName = phoneModels.value[randomIndex].name
      // }

      let targetLocalModel = createForm.value.localModel
      if (createForm.value.modelType === 'local' && (targetLocalModel === 'random' || targetLocalModel === '') && localPhoneModels.value.length > 0) {
        const randomIndex = Math.floor(Math.random() * localPhoneModels.value.length)
        targetLocalModel = localPhoneModels.value[randomIndex].name
      }

      let targetModelStatic = createForm.value.modelStatic
      if (createForm.value.modelType === 'backup' && targetModelStatic === 'random' && backupPhoneModels.value.length > 0) {
        const randomIndex = Math.floor(Math.random() * backupPhoneModels.value.length)
        targetModelStatic = backupPhoneModels.value[randomIndex].name
      }

      const target = {
        slot: currentSlot.value,
        modelName: modelName,
        modelType: createForm.value.modelType,
        localModel: targetLocalModel,
        modelStatic: targetModelStatic,
        deviceIp: createDevice.value.ip,
        deviceVersion: createDevice.value.version,
        deviceId: createDevice.value.id,
        imageUrl: createForm.value.imageSelect === 'custom' ? createForm.value.customImageUrl : createForm.value.imageSelect,
        isLocalImage: createForm.value.imageCategory === 'local',
        localImageUrl: createForm.value.localImageUrl,
        // 网络管理参数
        vpcGroupId: createForm.value.vpcGroupId || '',
        vpcNodeId: createForm.value.vpcSelectMode === 'random' ? 'random' : createForm.value.vpcNodeId,
        mytBridgeName: createForm.value.mytBridgeName,
        macVlanIp: createForm.value.macVlanIp // macVlan IP
      }
      
      // 添加到任务队列
      const taskId = addTaskToQueue('create', [target])
      executeTask(taskId)
      
      // 关闭创建对话框
      createDialogVisible.value = false
      createLoading.value = false
      ElMessage.success('云机创建任务已添加到队列')
    }
  } catch (error) {
    if (error.message === '创建取消') {
      ElMessage.info('创建取消')
    } else {
      console.error('创建云机失败:', error)
      ElMessage.error(`创建云机失败：${error.message}`)
    }
  } finally {
    createLoading.value = false
  }
}

// 处理创建表单取消
const handleCreateCancel = () => {
  // 取消时也保存坑位选择，以便下次恢复
  if (createMode.value === 'batch' || createMode.value === 'multi-device-batch') {
    const slots = createForm.value.selectedSlots || []
    localStorage.setItem('createDialog_selectedSlots', JSON.stringify(slots))
  }
  createDialogVisible.value = false
  createLoading.value = false
}

// 显示更新镜像对话框
const showUpdateImageDialog = async (container) => {
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
  }

  // 刷新容器列表，确保拿到最新数据
  if (targetDevice) {
    await fetchAndroidContainers(targetDevice, false)
    // 从最新缓存中找到对应容器（按 name 匹配）
    const freshList = deviceCloudMachinesCache.value.get(targetDevice.ip) || []
    const freshContainer = freshList.find(inst => inst.name === container.name)
    if (freshContainer) {
      container = freshContainer
    }
  }

  // 确保 targetDevice 包含 androidType，供 resetAndroidContainer 使用
  if (targetDevice && container) {
    targetDevice.androidType = container.androidType
  }

  // Update createDevice for isPSeries computed property
  createDevice.value = targetDevice 

  if (!targetDevice) {
    ElMessage.error('没有选中设备')
    return
  }
  
  updateImageContainer.value = container
  
  // 解析分辨率回显
  let resolutionValue = '720x1280x320' // 默认值
  let customRes = {
    width: '720',
    height: '1280',
    dpi: '320'
  }

  // 检查是否是容器模式 (V2) 且有详细分辨率信息
  if (container.androidType === 'V2' && container.doboxWidth && container.doboxHeight) {
    const w = String(container.doboxWidth)
    const h = String(container.doboxHeight)
    const dpi = String(container.doboxDpi || '320')
    
    if (w === '720' && h === '1280') {
      resolutionValue = '720x1280x320'
    } else if (w === '1080' && h === '1920') {
      resolutionValue = '1080x1920x420'
    } else {
      resolutionValue = 'custom'
      customRes = {
        width: w,
        height: h,
        dpi: dpi
      }
    }
  }


  // dns：判断是否是预设值，否则归为 custom
  const dnsRaw = container.dns || ''
  const dnsPresets = ['223.5.5.5', '8.8.8.8', '']
  const dnsSelectValue = dnsPresets.includes(dnsRaw) ? dnsRaw : 'custom'
  const customDnsValue = dnsPresets.includes(dnsRaw) ? '' : dnsRaw

  // networkName==='myt' 表示公有网卡
  const isPublicNetworkCard = container.networkName === 'myt'

  // 重置表单
  updateImageForm.value = {
    imageSelect: '',
    customImageUrl: '',
    modelName: container.modelName || '',
    enableMagisk: container.mgenable === '1',
    enableGMS: container.gmsenable === '1',
    dns: dnsSelectValue,
    customDns: customDnsValue,
    resolution: resolutionValue,
    customResolution: customRes,
    longitud: '',
    latitude: '',
    vpcGroupId: '',
    vpcNodeId: '',
    vpcSelectMode: 'specified',
    randomFile: container.randomFile || false,
    networkCardType: isPublicNetworkCard ? 'public' : 'private',
    mytBridgeName: container.mytBridgeName || '',
    macVlanIp: container.macVlanIp || '',
    enforce: container.enforce !== false // 安全模式，从容器数据读取（undefined/true均视为开启）
  }
  
  // 获取设备类型
  const deviceType = targetDevice.name || 'C1'
  
  // 获取镜像列表
   await fetchImageList(deviceType)
   
   // 获取网络分组列表
   await fetchVpcGroupList(targetDevice.ip)

   // 获取网卡列表 (仅V3设备)
   if (targetDevice.version === 'v3') {
     // 由于 fetchNetworkCards 依赖 createDevice.value 和 createForm.value，我们需要临时设置一下
     // 但更好的方式是让 fetchNetworkCards 接受参数，或者在这里直接调用底层 API
     // 考虑到复用性，我们修改 fetchNetworkCards 使其更通用，或者在这里手动调用
     
     // 方案：直接复用 fetchNetworkCards，但在调用前确保 createForm 的值被正确设置
     // 注意：fetchNetworkCards 使用的是 createForm.value，而这里是 updateImageForm
     // 所以我们需要改造 fetchNetworkCards 或者 复制一份逻辑
     // 为了避免副作用，我们在这里复制一份逻辑，专门用于 updateImageForm
     await fetchNetworkCardsForUpdate(targetDevice.ip)
   }
   
   // 获取当前容器使用的镜像
  const currentImageUrl = container.image || ''
  const cleanedUrl = currentImageUrl.toLowerCase()
  
  // 查找当前镜像是否在列表中 - 优化匹配逻辑
  let currentImage = null
  
  // 1. 精确匹配：检查image.url是否与currentImageUrl完全匹配
  currentImage = filteredImageList.value.find(image => {
    return image.url && image.url.toLowerCase() === cleanedUrl
  })
  
  // 2. 如果没有精确匹配，尝试模糊匹配：检查image.url是否是currentImageUrl的一部分
  if (!currentImage) {
    currentImage = filteredImageList.value.find(image => {
      return image.url && cleanedUrl.includes(image.url.toLowerCase())
    })
  }
  
  // 3. 如果仍然没有找到，尝试在完整镜像列表中查找
  if (!currentImage) {
    currentImage = imageList.value.find(image => {
      return image.url && image.url.toLowerCase() === cleanedUrl
    })
  }
  
  // 4. 最后尝试在完整镜像列表中进行模糊匹配
  if (!currentImage) {
    currentImage = imageList.value.find(image => {
      return image.url && cleanedUrl.includes(image.url.toLowerCase())
    })
  }
  
  if (currentImage) {
    // 如果在列表中，选择它
    updateImageForm.value.imageSelect = currentImage.url
  } else if (currentImageUrl) {
    // 如果不在列表中但有镜像URL，选择自定义并填写
    updateImageForm.value.imageSelect = 'custom'
    updateImageForm.value.customImageUrl = currentImageUrl
  } else {
    // 如果没有镜像URL，设置默认镜像为过滤后的第一个镜像
    if (filteredImageList.value.length > 0) {
      updateImageForm.value.imageSelect = filteredImageList.value[0].url
    }
  }
  
  // 如果是V3设备，获取型号列表
  if (targetDevice.version === 'v3') {
    await getV3PhoneModels(targetDevice.ip)
  }
  
  updateImageDialogVisible.value = true
}

// 处理更新镜像提交
const handleUpdateImageSubmit = async () => {
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && updateImageContainer.value && updateImageContainer.value.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === updateImageContainer.value.deviceIp) || { ip: updateImageContainer.value.deviceIp, version: 'v3' }
  }

  if (!updateImageContainer.value || !targetDevice) {
    ElMessage.error('没有选中容器或设备')
    return
  }
  
  updateImageLoading.value = true
  try {
    // 获取镜像URL
    let imageUrl = updateImageForm.value.imageSelect
    if (updateImageForm.value.imageSelect === 'custom') {
      imageUrl = updateImageForm.value.customImageUrl
    }
    
    // 确保镜像URL不为空
    if (!imageUrl) {
      ElMessage.error('请选择或输入镜像URL')
      return
    }
    
    // 调用后端API更新镜像
    console.log('调用后端API更新镜像:', updateImageContainer.value.name, imageUrl)
    
    if (targetDevice.version === 'v3') {
      // V3设备使用switchImage API
      // 从手机型号列表中查找对应的ModelId
      let modelId = ''
      const model = phoneModels.value.find(m => m.name === updateImageForm.value.modelName)
      if (model) {
        modelId = model.id || ''
      } 
      
      // 确保ModelId不为空，使用默认值
      if (!modelId) {
        modelId = '' // 默认空
      }
      
      // 处理DNS设置
      let dnsValue = updateImageForm.value.dns;
      if (updateImageForm.value.dns === 'custom' && updateImageForm.value.customDns) {
        dnsValue = updateImageForm.value.customDns;
      }
      
      // 解析分辨率
      let doboxWidth = ''
      let doboxHeight = ''
      let doboxDpi = ''
      
      if (updateImageForm.value.resolution === 'default') {
        // 机型默认分辨率，传递空值
        doboxWidth = ''
        doboxHeight = ''
        doboxDpi = ''
      } else if (updateImageForm.value.resolution === 'custom') {
        // 自定义分辨率
        doboxWidth = updateImageForm.value.customResolution.width || '720'
        doboxHeight = updateImageForm.value.customResolution.height || '1280'
        doboxDpi = updateImageForm.value.customResolution.dpi || '320'
      } else {
        // 预设分辨率
        const parts = updateImageForm.value.resolution.split('x')
        if (parts.length === 3) {
          doboxWidth = parts[0]
          doboxHeight = parts[1]
          doboxDpi = parts[2]
        } else {
          // 兼容旧格式
          doboxWidth = '720'
          doboxHeight = '1280'
          doboxDpi = '320'
        }
      }
      
      // 构造请求体，参考老客户端的V3SwitchImageReq结构体
      const switchImageReq = {
        name: updateImageContainer.value.name,
        modelId: modelId,
        modelName: updateImageForm.value.modelName,
        imageUrl: imageUrl,
        mgenable: updateImageForm.value.enableMagisk ? '1' : '0',
        gmsenable: updateImageForm.value.enableGMS ? '1' : '0',
        dns: dnsValue,
        doboxWidth: doboxWidth,
        doboxHeight: doboxHeight,
        doboxDpi: doboxDpi,
        randomFile: updateImageForm.value.randomFile, // 随机系统文件
        enforce: updateImageForm.value.enforce !== false, // 安全模式，默认开启
        mytBridgeName: updateImageForm.value.mytBridgeName, // 网卡参数
        macVlanIp: updateImageForm.value.macVlanIp // macVlan IP
      }
      
      // 添加VPC网络管理配置
      if (updateImageForm.value.vpcGroupId) {
        switchImageReq.vpcGroupId = updateImageForm.value.vpcGroupId
        if (updateImageForm.value.vpcSelectMode === 'random') {
          switchImageReq.vpcID = getRandomVpcNodeId()
        } else {
          switchImageReq.vpcID = updateImageForm.value.vpcNodeId || ''
        }
      }
      
      console.log('发送V3 switchImage请求:', switchImageReq)
      
      // 使用authRetry处理认证
      await authRetry(targetDevice, async (password) => {
        let headers = {}
        if (password) {
          const auth = btoa(`admin:${password}`)
          headers = {
            'Authorization': `Basic ${auth}`
          }
        }
        
        let switchUrl = `http://${getDeviceAddr(targetDevice.ip)}/android/switchImage`
        let requestBody = switchImageReq

        // 检查是否是容器模式 (V2)
        if (updateImageContainer.value.androidType === 'V2') {
          switchUrl = `http://${getDeviceAddr(targetDevice.ip)}/androidV2/switchImage`
          requestBody = {
            imageUrl: imageUrl,
            name: updateImageContainer.value.name,
            dns: dnsValue,
            doboxDpi: doboxDpi,
            doboxFps: '24', // 默认24FPS
            doboxHeight: doboxHeight,
            doboxWidth: doboxWidth,
            enforce: updateImageForm.value.enforce !== false // 安全模式
          }
          if (updateImageForm.value.networkCardType === 'public' && updateImageForm.value.macVlanIp) {
            requestBody.macVlanIp = updateImageForm.value.macVlanIp
          } else if (updateImageForm.value.networkCardType === 'private' && updateImageForm.value.mytBridgeName) {
            requestBody.mytBridgeName = updateImageForm.value.mytBridgeName
          }
        }
        
        // 发送前关闭投屏窗口
        try {
          await CloseProjectionWindow(updateImageContainer.value.name);
        } catch(e) { console.warn('关闭投屏窗口失败:', e); }

        // 使用axios调用V3 API
        const response = await axios.post(switchUrl, requestBody, {
          headers: headers
        })
        console.log('V3 switchImage成功，返回数据:', response.data)
        
        // 检查响应状态
        if (response.data.code !== 0) {
          if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
            throw new Error('Authentication Failed')
          } else {
            throw new Error(`切换镜像失败: ${response.data.message || '未知错误'}`)
          }
        }
      })
    } else {
      // 对于V2及以前的设备，根据要求，不调用API，只返回功能正在开发中的提示
      ElMessage.info('功能正在开发中')
      updateImageDialogVisible.value = false
      return
    }
    
    // 关闭对话框
    updateImageDialogVisible.value = false

    // 立即更新本地缓存中该容器的关键字段（设备端重建中，接口可能返回旧值）
    const containerName = updateImageContainer.value.name
    const cachedList = deviceCloudMachinesCache.value.get(targetDevice.ip)
    if (cachedList) {
      const idx = cachedList.findIndex(m => m.name === containerName)
      if (idx !== -1) {
        const finalDns = updateImageForm.value.dns === 'custom'
          ? (updateImageForm.value.customDns || '')
          : (updateImageForm.value.dns || '')
        const isPublic = updateImageForm.value.networkCardType === 'public'
        cachedList[idx] = {
          ...cachedList[idx],
          dns: finalDns,
          randomFile: updateImageForm.value.randomFile,
          mgenable: updateImageForm.value.enableMagisk ? '1' : '0',
          gmsenable: updateImageForm.value.enableGMS ? '1' : '0',
          networkName: isPublic ? 'myt' : (updateImageForm.value.mytBridgeName || cachedList[idx].networkName),
          macVlanIp: isPublic ? (updateImageForm.value.macVlanIp || '') : '',
          mytBridgeName: isPublic ? '' : (updateImageForm.value.mytBridgeName || ''),
          ip: isPublic ? (updateImageForm.value.macVlanIp || cachedList[idx].ip) : cachedList[idx].ip,
          image: imageUrl,
          enforce: updateImageForm.value.enforce !== false, // 安全模式
        }
        deviceCloudMachinesCache.value.set(targetDevice.ip, [...cachedList])
      }
    }
    
    // 刷新容器列表，传递isUserInitiated=true确保强制刷新
    await fetchAndroidContainers(targetDevice, true)
    
    ElMessage.success('镜像更新成功')
  } catch (error) {
    console.error('更新镜像失败:', error)
    ElMessage.error(`更新镜像失败：${error.message}`)
  } finally {
    updateImageLoading.value = false
  }
}

// 处理更新镜像取消
const handleUpdateImageCancel = () => {
  updateImageDialogVisible.value = false
}

// 批量更新镜像 - 等待下载完成（轮询 isDownloadingImage）
const waitForDownloadComplete = () => {
  return new Promise((resolve, reject) => {
    if (!isDownloadingImage.value) {
      resolve(true)
      return
    }
    const timer = setInterval(() => {
      if (!isDownloadingImage.value) {
        clearInterval(timer)
        // 检查最近一次下载是否成功（task 状态）
        const lastTask = [...taskQueue.value].reverse().find(t => t.type === 'downloadImage')
        if (lastTask && lastTask.status === 'failed') {
          reject(new Error(lastTask.error || '下载失败'))
        } else {
          resolve(true)
        }
      }
    }, 500)
    // 超时 30 分钟
    setTimeout(() => {
      clearInterval(timer)
      reject(new Error('下载超时'))
    }, 30 * 60 * 1000)
  })
}

// 批量更新镜像 - 执行提交
// 轮询任务进度，直到任务完成或超时
// onProgress(successCount, failCount, total) 回调用于实时更新进度条
const pollTaskStatus = async (deviceIp, taskId, headers, totalCount, groupLabel, onProgress) => {
  const maxWaitMs = 10 * 60 * 1000 // 最长等待10分钟
  const pollInterval = 3000 // 每3秒查询一次
  const startTime = Date.now()
  let successCount = 0
  let failCount = 0
  const failDetails = []

  while (Date.now() - startTime < maxWaitMs) {
    try {
      const res = await axios.get(
        `http://${getDeviceAddr(deviceIp)}/android/task-status?taskId=${taskId}`,
        { headers }
      )
      const data = res.data?.data
      if (res.data?.code === 0 && data) {
        const status = data.status
        const total = data.total || totalCount
        const successNames = data.successNames || []
        const failedMap = data.failedMap || {}
        const currentSuccess = successNames.length
        const currentFail = Object.keys(failedMap).length
        const processed = currentSuccess + currentFail

        batchUpdateImageStatusText.value = `设备 ${deviceIp}(${groupLabel}) 更新中：${processed}/${total}`

        // 实时回调，每次轮询都更新进度
        if (onProgress) {
          onProgress(currentSuccess, currentFail, total)
        }

        if (status === 'completed' || status === 'done' || status === 'finished') {
          successCount = currentSuccess
          failCount = currentFail
          for (const [name, reason] of Object.entries(failedMap)) {
            failDetails.push({ deviceIP: `${deviceIp}(${groupLabel})`, machineName: name, error: reason })
          }
          break
        } else if (status === 'failed' || status === 'error') {
          failCount = totalCount
          failDetails.push({ deviceIP: `${deviceIp}(${groupLabel})`, error: '任务失败' })
          break
        }
      }
    } catch (e) {
      // 查询失败，继续重试
    }
    await new Promise(resolve => setTimeout(resolve, pollInterval))
  }

  // 超时处理：若仍未完成，视为成功（任务已提交）
  if (successCount === 0 && failCount === 0) {
    successCount = totalCount
  }

  return { successCount, failCount, failDetails }
}

const executeBatchUpdateImage = async () => {
  // 校验每组都已选镜像
  for (const group of batchUpdateImageGroups.value) {
    const addr = group.selectedUrl === 'custom' ? group.customUrl.trim() : group.selectedUrl.trim()
    if (!addr) {
      ElMessage.warning(`请为「${group.groupLabel}」选择或输入镜像地址`)
      return
    }
  }

  // 收集所有容器作为 targets（用于任务队列展示设备信息）
  const allContainers = []
  for (const group of batchUpdateImageGroups.value) {
    for (const c of group.containers) {
      allContainers.push(c)
    }
  }

  // 创建任务队列条目
  const imageLabels = [...new Set(batchUpdateImageGroups.value.map(g =>
    g.selectedUrl === 'custom' ? (g.customUrl.trim() || '自定义') : (g.selectedUrl.trim() || '')
  ))].join(', ')
  const queueTaskId = addTaskToQueue('updateImage', allContainers, {
    imageLabels
  })
  const queueTask = taskQueue.value.find(t => t.id === queueTaskId)
  if (queueTask) {
    queueTask.status = 'running'
    queueTask.startTime = new Date()
  }

  // 关闭弹窗，后台执行
  batchUpdateImageDialogVisible.value = false
  ElMessage.info('批量更新镜像任务已创建，可在任务队列中查看进度')

  // 快照当前分组数据，避免弹窗关闭后数据被清空
  const groups = JSON.parse(JSON.stringify(batchUpdateImageGroups.value))

  // 后台异步执行
  ;(async () => {
    let successCount = 0
    let failCount = 0
    const failDetails = []

    try {
      // 收集所有需要处理的镜像 URL（自定义 URL 不做下载检查）
      const urlsToProcess = []
      for (const group of groups) {
        const imageAddress = group.selectedUrl === 'custom' ? group.customUrl.trim() : group.selectedUrl.trim()
        if (group.selectedUrl === 'custom') continue
        const existing = urlsToProcess.find(u => u.imageUrl === imageAddress)
        if (existing) {
          existing.groups.push(group)
        } else {
          const imgObj = imageList.value.find(img => img.url === imageAddress) || { url: imageAddress, name: imageAddress }
          urlsToProcess.push({ imageUrl: imageAddress, imageObj: imgObj, groups: [group] })
        }
      }

      // 逐一检查并下载镜像
      for (const item of urlsToProcess) {
        const isDownloaded = await checkImageDownloadStatus(item.imageUrl)
        if (!isDownloaded) {
          await downloadOnlineImage(item.imageObj)
          try {
            await waitForDownloadComplete()
          } catch (err) {
            throw new Error(`镜像「${item.imageObj.name || item.imageUrl}」下载失败：${err.message}`)
          }
          const downloadedNow = await checkImageDownloadStatus(item.imageUrl)
          if (!downloadedNow) {
            throw new Error(`镜像「${item.imageObj.name || item.imageUrl}」下载后验证失败`)
          }
        }
      }

      // 收集需要推送的设备，结构：Map<imageUrl, Set<deviceIp>>
      const imagePushTargets = new Map()
      for (const group of groups) {
        const imageAddress = group.selectedUrl === 'custom' ? group.customUrl.trim() : group.selectedUrl.trim()
        if (group.selectedUrl === 'custom') continue
        const targetContainers = group.androidType === 'V2'
          ? group.containers.filter(c => c.androidType === 'V2')
          : group.containers.filter(c => c.androidType !== 'V2')
        for (const container of targetContainers) {
          const deviceIp = container.deviceIp || ''
          if (!deviceIp) continue
          if (!imagePushTargets.has(imageAddress)) imagePushTargets.set(imageAddress, new Set())
          imagePushTargets.get(imageAddress).add(deviceIp)
        }
      }

      // 逐一推送镜像到各设备
      for (const [imageUrl, deviceIpSet] of imagePushTargets.entries()) {
        const downloadResult = await IsImageDownloaded(imageUrl)
        if (!downloadResult.downloaded || !downloadResult.local_path) {
          throw new Error(`无法获取镜像本地路径：${imageUrl}`)
        }
        const localPath = downloadResult.local_path
        for (const deviceIp of deviceIpSet) {
          const device = devices.value.find(d => d.ip === deviceIp)
          const version = device?.version || ''
          const password = getDevicePassword(deviceIp)

          // 检查设备上是否已存在该镜像，避免重复推送
          try {
            const deviceImages = await GetImages(deviceIp, version, password || '')
            let imageAlreadyExists = false
            if (Array.isArray(deviceImages)) {
              imageAlreadyExists = deviceImages.some(img => {
                const repoTags = img.RepoTags || img.imageUrl || img.Image
                if (Array.isArray(repoTags)) {
                  return repoTags.some(tag => tag === imageUrl || tag.includes(imageUrl))
                }
                return repoTags === imageUrl || (typeof repoTags === 'string' && repoTags.includes(imageUrl))
              })
            } else if (deviceImages && deviceImages.list) {
              imageAlreadyExists = deviceImages.list.some(img => {
                const imgUrl = img.imageUrl || img.Image
                return imgUrl === imageUrl || (typeof imgUrl === 'string' && imgUrl.includes(imageUrl))
              })
            }
            if (imageAlreadyExists) {
              console.log(`[批量更新镜像] 设备 ${deviceIp} 已存在镜像 ${imageUrl}，跳过推送`)
              continue
            }
          } catch (e) {
            console.warn(`[批量更新镜像] 检查设备 ${deviceIp} 镜像列表失败，继续推送:`, e)
          }

          const loadResult = await LoadImageToDevice(deviceIp, localPath, version, password || '')
          if (!loadResult.success && !loadResult.canceled) {
            throw new Error(`推送镜像到设备 ${deviceIp} 失败：${loadResult.message || '未知错误'}`)
          }
        }
      }

      // 调用 change-image 接口
      const allSubTasks = []
      const totalContainers = allContainers.length || 1
      for (const group of groups) {
        const imageAddress = group.selectedUrl === 'custom' ? group.customUrl.trim() : group.selectedUrl.trim()
        const apiPath = group.androidType === 'V2' ? '/androidV2/change-image' : '/android/change-image'
        const targetContainers = group.androidType === 'V2'
          ? group.containers.filter(c => c.androidType === 'V2')
          : group.containers.filter(c => c.androidType !== 'V2')

        const deviceContainerMap = new Map()
        for (const container of targetContainers) {
          const deviceIp = container.deviceIp || ''
          if (!deviceIp) continue
          if (!deviceContainerMap.has(deviceIp)) deviceContainerMap.set(deviceIp, [])
          const containerName = container.name || container.Name || container.names?.[0] || ''
          if (containerName) deviceContainerMap.get(deviceIp).push(containerName)
        }

        for (const [deviceIp, containerNames] of deviceContainerMap.entries()) {
          allSubTasks.push(async () => {
            try {
              const savedPassword = getDevicePassword(deviceIp)
              const headers = {}
              if (savedPassword) {
                headers['Authorization'] = `Basic ${btoa(`admin:${savedPassword}`)}`
              }
              
              try {
                for (const cn of containerNames) {
                  await CloseProjectionWindow(cn);
                }
              } catch(e) { console.warn('关闭投屏窗口失败:', e); }

              const response = await axios.post(
                `http://${getDeviceAddr(deviceIp)}${apiPath}`,
                { containerNames, image: imageAddress },
                { headers }
              )
              if (response.data && response.data.code === 0) {
                const remoteTaskId = response.data.data?.taskId
                if (remoteTaskId) {
                  // 记录本子任务开始前的累计数，用于计算增量
                  const baseSuccess = successCount
                  const baseFail = failCount
                  const pollResult = await pollTaskStatus(
                    deviceIp, remoteTaskId, headers, containerNames.length, group.groupLabel,
                    (curSuccess, curFail, _total) => {
                      // 每次轮询实时更新进度条（增量 = 本轮已完成数 - 上一轮）
                      if (queueTask) {
                        const totalDone = baseSuccess + baseFail + curSuccess + curFail
                        queueTask.completed = baseSuccess + curSuccess
                        queueTask.failed = baseFail + curFail
                        queueTask.progress = Math.min(
                          Math.round(totalDone / totalContainers * 100),
                          99 // 未最终确认完成前最多到99%
                        )
                      }
                    }
                  )
                  successCount += pollResult.successCount
                  failCount += pollResult.failCount
                  if (pollResult.failDetails.length > 0) {
                    failDetails.push(...pollResult.failDetails)
                  }
                } else {
                  successCount += containerNames.length
                }
              } else {
                failCount += containerNames.length
                failDetails.push({ deviceIP: `${deviceIp}(${group.groupLabel})`, error: response.data?.message || '未知错误' })
              }
            } catch (err) {
              failCount += containerNames.length
              failDetails.push({ deviceIP: `${deviceIp}(${group.groupLabel})`, error: err.message || '请求失败' })
            }
            // 子任务完成后同步最终进度
            if (queueTask) {
              queueTask.completed = successCount
              queueTask.failed = failCount
              queueTask.progress = Math.round((successCount + failCount) / totalContainers * 100)
            }
          })
        }
      }

      if (allSubTasks.length === 0) {
        if (queueTask) {
          queueTask.status = 'failed'
          queueTask.endTime = new Date()
          queueTask.error = '未能获取到有效的容器信息'
        }
        return
      }

      await Promise.all(allSubTasks.map(fn => fn()))

      // 完成
      if (queueTask) {
        queueTask.completed = successCount
        queueTask.failed = failCount
        queueTask.progress = 100
        queueTask.status = failCount === 0 ? 'completed' : 'failed'
        queueTask.endTime = new Date()
        if (failDetails.length > 0) {
          queueTask.failedTargets = failDetails
        }
      }
    } catch (error) {
      if (queueTask) {
        queueTask.status = 'failed'
        queueTask.endTime = new Date()
        queueTask.error = error.message
      }
      ElMessage.error(`批量更新镜像失败：${error.message}`)
    }
  })()
}

// 切换备份
const switchBackup = async (backupId) => {
  console.log(`切换坑位 ${backupCurrentSlot.value} 到备份 ${backupId}`)
  
  if (!activeDevice.value) {
    ElMessage.error('没有选中设备')
    backupListVisible.value = false
    return
  }
  
  // 记录当前正在切换的坑位
  switchingBackupSlot.value = backupCurrentSlot.value
  backupLoading.value = true
  
  try {
    // 1. 找到当前坑位运行的容器并关机
    const runningContainer = instances.value.find(inst => 
      inst.indexNum === backupCurrentSlot.value && 
      inst.status === 'running'
    );
    
    if (runningContainer) {
      ElMessage.info(`正在关机当前运行的容器: ${runningContainer.name}`)
      await authRetry(activeDevice.value, async (password) => {
        await stopContainer(activeDevice.value, runningContainer.name, password)
      })
      // 等待1秒，确保容器完全关闭
      await new Promise(resolve => setTimeout(resolve, 1000))
    }
    
    // 2. 找到要切换的备份容器并开机
    // 从allInstances中查找，因为instances只包含每个坑位的一个容器
    const backupContainer = allInstances.value.find(inst => 
      inst.name === backupId
    );
    
    if (backupContainer) {
      ElMessage.info(`正在开机备份容器: ${backupContainer.name}`)
      await authRetry(activeDevice.value, async (password) => {
        await startContainer(activeDevice.value, backupContainer.name, password)
      })
      // 等待2秒，确保容器完全启动
      await new Promise(resolve => setTimeout(resolve, 2000))
      ElMessage.success(`成功切换坑位 ${backupCurrentSlot.value} 到备份 ${backupId}`)
    } else {
      ElMessage.error('未找到指定的备份容器')
    }
    
    // 3. 刷新容器列表，重试最多3次，确保获取最新状态
    for (let i = 0; i < 3; i++) {
      await fetchAndroidContainers(activeDevice.value)
      // 检查是否获取到了正确的状态
      const updatedContainer = instances.value.find(inst => 
        inst.indexNum === backupCurrentSlot.value && 
        (inst.status === 'running' || inst.status === 'created')
      );
      if (updatedContainer) {
        break;
      }
      // 等待1秒后重试
      await new Promise(resolve => setTimeout(resolve, 1000))
    }
  } catch (error) {
    console.error('切换备份失败:', error)
    ElMessage.error('切换备份失败: ' + (error.message || '未知错误'))
  } finally {
    backupLoading.value = false
    backupListVisible.value = false
    switchingBackupSlot.value = null
  }
}

// 删除单个备份
const deleteBackup = async (backupId) => {
  console.log(`删除备份: ${backupId}`)
  
  if (!activeDevice.value) {
    ElMessage.error('没有选中设备')
    return
  }
  
  try {
    await ElMessageBox.confirm('确定要删除该备份吗？删除后数据将无法恢复。', '删除备份', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    backupLoading.value = true
    
    await authRetry(activeDevice.value, async (password) => {
      await deleteContainer(activeDevice.value, backupId, password)
    })
    
    ElMessage.success('删除备份成功')
    
    // 重新获取容器列表，更新 allInstances
    await fetchAndroidContainers(activeDevice.value, true)
    
    // 刷新备份列表
    initBackupList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除备份失败:', error)
      ElMessage.error('删除备份失败: ' + (error.message || '未知错误'))
    }
  } finally {
    backupLoading.value = false
  }
}

// 批量删除备份
const batchDeleteBackup = async () => {
  if (selectedBackupList.value.length === 0) {
    ElMessage.warning('请先选择要删除的备份')
    return
  }
  
  console.log(`批量删除备份: ${selectedBackupList.value}`)
  
  if (!activeDevice.value) {
    ElMessage.error('没有选中设备')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedBackupList.value.length} 个备份吗？删除后数据将无法恢复。`, '批量删除备份', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    backupLoading.value = true
    
    for (const backupId of selectedBackupList.value) {
      await authRetry(activeDevice.value, async (password) => {
        await deleteContainer(activeDevice.value, backupId, password)
      })
    }
    
    ElMessage.success(`成功删除 ${selectedBackupList.value.length} 个备份`)
    
    // 清空选择
    selectedBackupList.value = []
    
    // 重新获取容器列表，更新 allInstances
    await fetchAndroidContainers(activeDevice.value, true)
    
    // 刷新备份列表
    initBackupList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除备份失败:', error)
      ElMessage.error('批量删除备份失败: ' + (error.message || '未知错误'))
    }
  } finally {
    backupLoading.value = false
  }
}

// 添加新分组
const addBackupGroup = () => {
  const newGroup = `新分组${backupGroups.value.length + 1}`
  backupGroups.value.push(newGroup)
  selectedBackupGroup.value = newGroup
  // 功能正在开发中提示
  ElMessage.info('功能正在开发中')
}

// 添加新机型
const addNewModel = () => {
  const newModel = `新机型${cloudMachineGroups.value.length}`
  cloudMachineGroups.value.push({
    id: `model-${Date.now()}`,
    name: newModel,
    devices: devices.value.map(device => {
      return {
        id: device.id,
        ip: device.ip,
        cloudMachines: []
      };
    })
  })
  ElMessage.success(`成功创建新机型: ${newModel}`)
}

// 支持从默认机型拖拽坑位到新机型
const handleDragAndDrop = (sourceModelId, targetModelId, slot) => {
  // 找到源机型和目标机型
  const sourceModel = cloudMachineGroups.value.find(model => model.id === sourceModelId)
  const targetModel = cloudMachineGroups.value.find(model => model.id === targetModelId)
  
  if (!sourceModel || !targetModel) return
  
  // 遍历所有设备
  devices.value.forEach(device => {
    // 在源机型中找到该设备的云机列表
    const sourceDevice = sourceModel.devices.find(d => d.ip === device.ip)
    const targetDevice = targetModel.devices.find(d => d.ip === device.ip)
    
    if (!sourceDevice || !targetDevice) return
    
    // 找到要移动的云机
    const cloudMachineIndex = sourceDevice.cloudMachines.findIndex(machine => machine.indexNum === slot)
    if (cloudMachineIndex === -1) return
    
    // 移动云机
    const [movedMachine] = sourceDevice.cloudMachines.splice(cloudMachineIndex, 1)
    targetDevice.cloudMachines.push(movedMachine)
  })
  
  ElMessage.success(`成功将坑位 ${slot} 从 ${sourceModel.name} 移动到 ${targetModel.name}`)
}

// 处理树形结构拖拽的验证函数
const handleDrop = (draggingNode, dropNode, dropType) => {
  // 只允许将云机节点拖拽到机型节点下的设备节点
  if (draggingNode.data.screenshot && dropNode.data.cloudMachines) {
    return true
  }
  return false
}

// 处理树形结构拖拽的完成函数
const handleNodeDrop = (draggingNode, dropNode, dropType) => {
  if (draggingNode.data.screenshot && dropNode.data.cloudMachines) {
    // 云机节点被拖拽到设备节点
    
    // 找到源机型和目标机型
    let sourceModel = null
    let targetModel = null
    
    // 查找源机型
    for (const model of cloudMachineGroups.value) {
      for (const device of model.devices) {
        for (const machine of device.cloudMachines) {
          if (machine.id === draggingNode.data.id) {
            sourceModel = model
            break
          }
        }
        if (sourceModel) break
      }
      if (sourceModel) break
    }
    
    // 查找目标机型
    for (const model of cloudMachineGroups.value) {
      for (const device of model.devices) {
        if (device.id === dropNode.data.id) {
          targetModel = model
          break
        }
      }
      if (targetModel) break
    }
    
    if (sourceModel && targetModel) {
      // 执行拖拽操作
      handleDragAndDrop(sourceModel.id, targetModel.id, draggingNode.data.indexNum)
    }
  }
}

// 计算排序后的备份列表
const sortedBackupList = computed(() => {
  return [...backupList.value].sort((a, b) => {
    let aVal = a[sortBy.value]
    let bVal = b[sortBy.value]  
    
    if (sortBy.value === 'createTime') {
      aVal = new Date(aVal).getTime()
      bVal = new Date(bVal).getTime()
    }
    
    if (sortOrder.value === 'ascending') {
      return aVal > bVal ? 1 : -1
    } else {
      return aVal < bVal ? 1 : -1
    }
  })
})

// 切换排序
const changeSort = (field) => {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'ascending' ? 'descending' : 'ascending'
  } else {
    sortBy.value = field
    sortOrder.value = 'descending'
  }
}

// 同步计算云机分组数据（无防抖）
const computeCloudMachineGroups = (mode = cloudManageMode.value) => {
  console.log('computeCloudMachineGroups called with mode:', mode)
  console.log('devices.value:', devices.value)
  console.log('deviceGroups.value:', deviceGroups.value)
  
  // 根据设备分组生成云机分组数据
  const groups = []
  
  // 遍历所有设备分组
  deviceGroups.value.forEach(groupName => {
    // 批量模式下只显示在线设备
    const groupDevices = devices.value.filter(device => {
      const deviceGroup = device.group || '默认分组'
      const isInGroup = deviceGroup === groupName
      const isOnline = mode === 'batch' 
        ? devicesStatusCache.value.get(device.id) === 'online'
        : true
      return isInGroup && isOnline
    }).sort((a, b) => compareIPs(a.ip, b.ip))
    
    if (groupDevices.length > 0) {
      // 为该分组创建云机分组
      const groupData = {
        id: `group-${groupName}`,
        name: groupName,
        devices: groupDevices.map(device => {
          // 从deviceCloudMachinesCache中获取对应设备的云机数据
          const deviceCloudMachines = deviceCloudMachinesCache.value.get(device.ip) || []
          // 批量模式下只显示运行中的云机
          const filteredCloudMachines = mode === 'batch' 
            ? deviceCloudMachines.filter(machine => machine.status === 'running')
            : deviceCloudMachines
          
          console.log('Device:', device.ip, 'has', filteredCloudMachines.length, 'cloud machines (mode:', mode, ')')
          
          // 对云机按照从ID中提取的坑位号进行排序
          const sortedCloudMachines = [...filteredCloudMachines].sort((a, b) => {
            const aParts = a.id.split('_')
            const bParts = b.id.split('_')
            const aSlot = aParts.length >= 2 ? parseInt(aParts[1]) : 0
            const bSlot = bParts.length >= 2 ? parseInt(bParts[1]) : 0
            return aSlot - bSlot
          })
          
          return {
            id: device.id,
            ip: device.ip,
            name: device.name,
            group: groupName,
            cloudMachines: sortedCloudMachines
          }
        })
      }
      
      groups.push(groupData)
    }
  })
  
  console.log('computed cloudMachineGroups:', groups)
  
  return groups
}

// 初始化云机分组数据（带防抖）
const initCloudMachineGroups = () => {
  if (initGroupsTimeout.value) {
    clearTimeout(initGroupsTimeout.value)
  }
  
  initGroupsTimeout.value = setTimeout(() => {
    cloudMachineGroups.value = computeCloudMachineGroups()
    
    // 不默认选择设备，保持selectedCloudDevice为空，显示12个空坑位
    // 只有在用户点击设备列表时才选择设备并加载对应数据
  }, 100)
}

// 包装删除容器函数，删除后刷新备份列表
const handleDeleteContainerWithBackupRefresh = async (container) => {
  const result = await handleDeleteContainer(container)
  // 如果删除成功且备份列表可见，刷新备份列表
  if (result?.success && backupListVisible.value) {
    initBackupList()
  }
  return result
}

// 关闭设备详情弹窗


// 设备详情弹窗关闭前的回调函数
const handleDeviceDetailsDialogClose = () => {
  console.log('设备详情弹窗关闭前回调');
  collapseRightSidebar();
}

// 显示设备详情弹窗


// 删除设备镜像
const handleDeleteImage = async (image) => {
  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      `确定要删除镜像 "${image.onlineImageName}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    // 直接调用Wails IPC删除镜像
    console.log('调用Wails IPC删除镜像:', image.url);
    const savedPassword = getDevicePassword(activeDevice.value.ip);
    
    let response = null;
    let deleteSuccess = false;
    
    // 尝试使用axios直接调用设备API删除镜像
    try {
      console.log('尝试使用axios直接调用设备API删除镜像');
      const apiUrl = `http://${activeDevice.value.version === 'v3' ? getDeviceAddr(activeDevice.value.ip) : activeDevice.value.ip + ':81'}/android/image?image=${encodeURIComponent(image.url)}`;
      const headers = {};
      if (savedPassword) {
        const auth = btoa(`admin:${savedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }
      
      const axiosResponse = await axios.delete(apiUrl, { headers });
      response = axiosResponse.data;
      console.log('使用axios直接调用设备API删除镜像结果:', response);
      if (response && response.code === 0) {
        deleteSuccess = true;
        refreshImageList()
      }
    } catch (error) {
      console.error('删除镜像API调用失败:', error);
      
      // 处理认证错误
      if (error.response?.status === 401 || error.response?.data?.code === 61) {
        console.log('删除镜像认证失败，需要显示认证对话框')
        // 显示认证对话框
        showAuthDialog(activeDevice.value, async (password) => {
          // 认证成功后重新尝试删除
          console.log('认证回调被调用，开始重试删除镜像，密码长度:', password ? password.length : 0)
          try {
            console.log('认证后重新尝试删除镜像');
            const apiUrl = `http://${activeDevice.value.version === 'v3' ? getDeviceAddr(activeDevice.value.ip) : activeDevice.value.ip + ':81'}/android/image?image=${encodeURIComponent(image.url)}`;
            const auth = btoa(`admin:${password}`);
            const headers = {
              'Authorization': `Basic ${auth}`
            };
            
            console.log('重试删除镜像 API URL:', apiUrl)
            console.log('重试删除镜像请求头:', headers)
            
            const axiosResponse = await axios.delete(apiUrl, { headers });
            const response = axiosResponse.data;
            console.log('认证后删除镜像结果:', response);
            
            if (response && response.code === 0) {
              ElMessage.success('镜像删除成功');
              refreshImageList();
            } else {
              ElMessage.error(response?.message || '镜像删除失败');
            }
          } catch (retryError) {
            console.error('认证后重试删除镜像失败:', retryError);
            console.error('错误详情:', retryError.response?.data);
            ElMessage.error('删除镜像失败: ' + (retryError.response?.data?.message || retryError.message));
          }
        });
        return;
      }
      
      // 如果所有API调用都失败，尝试从本地镜像列表中移除该镜像
      console.log('尝试从本地镜像列表中移除该镜像');
      
      // 找到该镜像在boxImages中的索引（matchedBoxImages是计算属性，不能直接修改）
      const imageIndex = boxImages.value.findIndex(img => img.url === image.url);
      if (imageIndex !== -1) {
        // 从原始数据源中移除该镜像
        boxImages.value.splice(imageIndex, 1);
        // 更新在线镜像状态
        await checkAllImagesDownloadStatus();
        // 显示成功消息
        ElMessage.success('镜像删除成功');
        return;
      }
      
      throw error;
    }

    // 根据API返回结果判断是否删除成功
    if (deleteSuccess) {
      // 显示成功消息
      ElMessage.success('镜像删除成功');
      // 重新获取设备镜像列表，刷新显示
      await switchImageCategory(selectedImageCategory.value);
    } else {
      // 显示失败消息
      ElMessage.error(`删除镜像失败: ${response?.message || '未知错误'}`);
    }
  } catch (error) {
    // 如果是用户取消操作，不显示错误消息
    if (error !== 'cancel') {
      console.error('删除镜像失败:', error);
      ElMessage.error(`删除镜像失败: ${error.message || '未知错误'}`);
    }
  }
}

// 设备选择事件处理
const handleDeviceSelect = async (device) => {
  console.log('handleDeviceSelect called with device:', device)
  if (device) {
    console.log('Setting activeDevice and selectedCloudDevice to:', device.ip)
    activeDevice.value = device
    selectedCloudDevice.value = device // 同步到selectedCloudDevice，确保云机管理也能获取到设备
    
    // 重置上传状态，避免切换设备后上传按钮被禁用
    isUploadingImage.value = false
    currentUploadImage.value = null
    uploadProgress.value = 0
    // 清除上传状态，因为不同设备的上传状态是独立的
    imageUploadStatus.value.clear()
    
    // 切换设备时停止当前截图刷新
    stopScreenshotRefresh()
    
    // 清空选中的云机列表，避免不同设备的云机混淆
    selectedCloudMachines.value = []
    
    // 立即清空右侧内容，提供更好的用户体验
    instances.value = []
    allInstances.value = []
    updateCloudMachines()
    
    // 异步获取选中设备的容器列表，isUserInitiated=true表示用户主动操作
    console.log('Calling fetchAndroidContainers for device:', device.ip)
    fetchAndroidContainers(device, true)
      .then(() => {
        console.log('After fetchAndroidContainers, instances.value:', instances.value)
        
        // 获取盒子镜像列表
        console.log('Calling fetchBoxImages')
        return fetchBoxImages()
      })
      .then(() => {
        // 检查在线镜像的下载和上传状态
        console.log('Calling checkAllImagesDownloadStatus')
        return checkAllImagesDownloadStatus()
      })
      .then(() => {
        // 检查在线镜像的上传状态
        console.log('Calling checkOnlineImagesUploadStatus')
        return checkOnlineImagesUploadStatus()
      })
      .then(() => {
        // 使用nextTick确保DOM更新后再执行后续操作，避免布局挤压
        nextTick(() => {
          // 刷新云机列表
          console.log('Calling updateCloudMachines')
          updateCloudMachines()
          console.log('After updateCloudMachines, cloudMachines.value:', cloudMachines.value)
          
          // 开始截图刷新初始化（不再自动刷新）
          // startScreenshotRefresh()
          
          console.log('handleDeviceSelect completed')
        })
      })
      .catch(error => {
        console.error('Error in fetchAndroidContainers:', error)
        // 即使出错，也确保云机列表已更新
        nextTick(() => {
          updateCloudMachines()
        })
      })
    
    // 如果是V3设备，获取详细信息
    if (device.version === 'v3') {
      console.log('Fetching V3 device info for:', device.ip)
      
      // 🔧 优先从心跳数据中读取
      const cachedInfo = deviceFirmwareInfo.value.get(device.id)
      if (cachedInfo && cachedInfo.originalData) {
        console.log('[快速加载] 从心跳数据加载 V3 设备信息')
        v3DeviceInfo.value = {
          sdkVersion: cachedInfo.sdkVersion,
          deviceModel: cachedInfo.deviceModel,
          originalData: cachedInfo.originalData
        }
        v3DeviceInfoLoaded.value = true
        
        // 仍然调用 fetchV3LatestInfo 获取最新版本信息
        fetchV3LatestInfo(device)
      } else {
        // 如果心跳数据还没准备好，使用传统方式获取
        console.log('[传统加载] 直接请求设备接口')
        fetchV3DeviceInfo(device)
      }
    } else {
      // 非V3设备，清空之前的V3设备信息
      v3DeviceInfo.value = {}
      v3LatestInfo.value = {}
      showUpgradeButton.value = false
    }
    
    // 获取设备的Docker网络列表
    console.log('Fetching Docker networks for:', device.ip)
    fetchDockerNetworks(device)
  } else {
    console.error('handleDeviceSelect called with null/undefined device')
  }
}

const handleHostDeviceSelectionChange = async (selection) => {
  selectedHostDevices.value = selection
  
  // 如果只选择了一个设备，设置为当前活动设备
  // if (selection.length === 1) {
  //   await handleDeviceSelect(selection[0])
  // }
}




const fetchV3DeviceInfo = async (device) => {
  if (!device || device.version !== 'v3') return
  
  try {
    // 重置加载状态
    v3DeviceInfoLoaded.value = false
    
    // 使用authRetry处理认证
    await authRetry(device, async (password) => {
      let headers = {}
      if (password) {
        const auth = btoa(`admin:${password}`)
        headers = {
          'Authorization': `Basic ${auth}`
        }
      }
      
      const url = `http://${getDeviceAddr(device.ip)}/info/device`
      const response = await axios.get(url, { 
        timeout: 1500,
        headers: headers
      })
      console.log('V3设备信息API返回:', response.data)
      
      if (response.data.code === 0) {
        const data = response.data.data
        // 适配实际API返回格式，映射字段名
        const deviceInfo = {
          sdkVersion: data.version, // 实际返回的version字段对应SDK版本
          deviceModel: data.model, // 实际返回的model字段对应设备型号
          hardwareVersion: '未知', // API未返回硬件版本
          firmwareVersion: '未知', // API未返回固件版本
          // 保留原始数据，方便后续扩展
          originalData: data
        }
        
        // 使用nextTick确保UI更新
        await nextTick()
        v3DeviceInfo.value = deviceInfo
        
        // 将固件信息存储到全局Map中，以便在设备列表中显示
        // 使用set方法修改现有Map，避免替换整个对象导致的重新渲染
        deviceFirmwareInfo.value.set(device.id, deviceInfo)
        
        v3DeviceInfoLoaded.value = true
        console.log('V3设备信息更新完成:', v3DeviceInfo.value)
        
        // 获取最新版本信息
        await fetchV3LatestInfo(device)
        
        // ⚠️ 不再手动设置设备在线状态，由心跳检测系统统一管理
        // devicesStatusCache.value.set(device.id, 'online')
      } else if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        throw new Error('Authentication Failed')
      }
    })
  } catch (error) {
    console.error('获取V3设备信息失败:', error)
    // 接口超时或失败时，不清空设备信息，保持现有数据
    v3DeviceInfoLoaded.value = true // 即使失败也标记为已加载，避免一直显示加载中
    
    // ⚠️ 不手动标记设备为离线，由心跳检测系统统一管理
    // devicesStatusCache.value.set(device.id, 'offline')

    // 更新设备最后更新时间
    // devicesLastUpdateTime.value.set(device.id, Date.now())
    
    // 不需要调用API版本信息，因为设备已离线
  }
}

// 获取设备绑定状态
const fetchDeviceBindStatus = async () => {
  try {
    // 1. 获取所有设备的deviceId
    const deviceIds = devices.value.map(device => device.id)
    if(deviceIds.length === 0) return

    console.log('fetchDeviceBindStatus', deviceIds)
    
    // 2. 构造请求数据

    const formData = new URLSearchParams()
    formData.append('type', 'user_host_oper');
    const data = {
       act: 'get',
       data: JSON.stringify({ host: deviceIds }),
       token: token.value
    }
    formData.append('data', JSON.stringify(data));


    console.log('formData:', formData)

    const response = await fetch('https://www.moyunteng.com/api/api.php', {
          method: 'POST',
          body: formData
     })  
 
       const result = await response.json()
       console.log('response:', result)
    
    // 4. 处理API返回结果
    if (result.code == 200) {
      // 清空之前的绑定状态
      deviceBindStatus.value.clear()
      
      // 遍历返回的绑定状态数据，将结果存入deviceBindStatus
      Object.entries(result.data).forEach(([deviceId, status]) => {
        deviceBindStatus.value.set(deviceId, status)
      })
      
      console.log('设备绑定状态获取成功:', deviceBindStatus.value)
    } else if (result.code == 3030) {
      handleAuthExpired()
    }
     else {
      console.error('获取设备绑定状态失败:', result.message)
      ElMessage.error('获取设备绑定状态失败: ' + result.message)
    }
  } catch (error) {
    console.error('获取设备绑定状态异常:', error)
    ElMessage.error('获取设备绑定状态异常: ' + error.message)
  }
}

const fetchDockerNetworks = async (device) => {
  if (!device) return
  
  try {
    dockerNetworksLoading.value = true
    dockerNetworksError.value = ''
    
    // 使用后端代理获取Docker网络列表
    const networks = await getDockerNetworks(device)
    dockerNetworks.value = networks
    console.log('Docker网络列表:', dockerNetworks.value)
  } catch (error) {
    console.error('获取Docker网络列表失败:', error)
    dockerNetworks.value = []
    dockerNetworksError.value = `获取网络列表失败: ${error.message}`
  } finally {
    dockerNetworksLoading.value = false
  }
}

// 云机设备选择事件处理
const handleCloudDeviceSelect = async (device) => {
  console.log('handleCloudDeviceSelect called with device:', device)
  if (device) {
    console.log('Setting selectedCloudDevice and activeDevice to:', device.ip)
    selectedCloudDevice.value = device
    activeDevice.value = device // 同步到activeDevice，确保主机管理和云机管理使用同一设备
    
    // 切换设备时停止当前截图刷新，清空版本快照强制下次立即拉取新设备截图
    stopScreenshotRefresh()
    screenshotLocalVersions = {}
    
    // 清空选中的云机列表，避免不同设备的云机混淆
    selectedCloudMachines.value = []
    
    // 立即清空右侧内容，提供更好的用户体验
    instances.value = []
    allInstances.value = []
    updateCloudMachines()
    
    // 异步获取选中设备的容器列表，isUserInitiated=true表示用户主动操作
    console.log('Calling fetchAndroidContainers for device:', device.ip)
    fetchAndroidContainers(device, true)
      .then(() => {
        // 使用nextTick确保DOM更新后再执行后续操作，避免布局挤压
        nextTick(() => {
          // 刷新云机列表
          console.log('Calling updateCloudMachines')
          updateCloudMachines()
          
          // 开始截图定时刷新（每1秒）
          startScreenshotRefresh()
          console.log('handleCloudDeviceSelect completed')
        })
      })
      .catch(error => {
        console.error('Error in fetchAndroidContainers:', error)
        // 即使出错，也确保云机列表已更新
        nextTick(() => {
          updateCloudMachines()
        })
      })
  } else {
    console.error('handleCloudDeviceSelect called with null/undefined device')
  }
}

// 获取V3设备详细信息


// 获取V3最新版本信息
const fetchV3LatestInfo = async (device) => {
  if (!device || device.version !== 'v3') return
  
  // 检查设备状态，如果已标记为离线，直接返回，不调用接口
  const deviceStatus = devicesStatusCache.value.get(device.id)
  if (deviceStatus === 'offline') {
    console.log('设备已标记为离线，跳过获取V3最新版本信息:', device.ip)
    return
  }
  
  try {
    // 使用authRetry处理认证
    await authRetry(device, async (password) => {
      let headers = {}
      if (password) {
        const auth = btoa(`admin:${password}`)
        headers = {
          'Authorization': `Basic ${auth}`
        }
      }
      
      const url = `http://${getDeviceAddr(device.ip)}/info`
      const response = await axios.get(url, { 
        timeout: 1000,
        headers: headers
      })
      
      if (response.data.code === 0) {
        const data = response.data.data
        // 适配实际API返回格式
        v3LatestInfo.value = {
          version: data.latestVersion.toString(), // 将数字版本转换为字符串
          currentVersion: data.currentVersion.toString(),
          originalData: data
        }
        console.log('V3最新版本信息:', v3LatestInfo.value)
        
        // 检查是否需要显示升级按钮
        // 比较currentVersion和latestVersion数字版本
        // 只有当确实需要升级时才显示升级按钮，否则保持当前状态
        if (data.currentVersion < data.latestVersion) {
          showUpgradeButton.value = true
        } else {
          // 如果不需要升级，不改变当前状态，避免刷新后升级按钮消失
          // showUpgradeButton.value = false
        }
      } else if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        throw new Error('Authentication Failed')
      }
    })
  } catch (error) {
    console.error('获取V3最新版本信息失败:', error)
    // 只重置版本信息，不重置showUpgradeButton状态，避免刷新后升级按钮消失
    v3LatestInfo.value = {}
    // 保持showUpgradeButton当前状态，不设置为false

    // ⚠️ 不手动标记设备为离线，由心跳检测系统统一管理
    // devicesStatusCache.value.set(device.id, 'offline')

    // 更新设备最后更新时间
    // devicesLastUpdateTime.value.set(device.id, Date.now())
  }
}

// 升级SDK


// 获取Docker网络列表


// 显示添加macvlan网络弹窗


// 国家提示
const fetchCountryList =  () => { 
  // ElMessage.info('支持搜索')
  // this.$message({
  //     message: '请选择一个选项',
  //     type: 'info', // 或者 'success', 'warning', 'error' 等类型
  // });
  createForm.value.countryCode = ''
}


// 添加macvlan网络
const handleAddMacvlanSubmit = async () => {
  if (!activeDevice.value) return
  
  try {
    addMacvlanLoading.value = true
    
    // 调用后端API创建macvlan网络
    console.log('添加macvlan网络:', addMacvlanForm.value)
    
    const result = await createDockerNetwork(activeDevice.value, addMacvlanForm.value)
    
    if (result.success) {
      ElMessage.success('macvlan网络添加成功')
      addMacvlanDialogVisible.value = false
      
      // 刷新网络列表
      await fetchDockerNetworks(activeDevice.value)
    } else {
      ElMessage.error(`添加macvlan网络失败: ${result.message}`)
    }
  } catch (error) {
    console.error('添加macvlan网络失败:', error)
    ElMessage.error(`添加macvlan网络失败: ${error.message}`)
  } finally {
    addMacvlanLoading.value = false
  }
}

// 取消添加macvlan网络
const handleAddMacvlanCancel = () => {
  addMacvlanDialogVisible.value = false
}

// 显示添加macvlan网络弹窗
const showAddMacvlanDialog = () => {
  // 重置表单
  addMacvlanForm.value = {
    networkName: '',
    parentInterface: '',
    subnet: '',
    gateway: '',
    ipRange: '',
    isPrivate: false
  }
  addMacvlanDialogVisible.value = true
}

// 处理修改网络


// 修改网络提交
const handleEditNetworkSubmit = async () => {
  if (!activeDevice.value || !currentEditingNetwork.value) return
  
  try {
    editNetworkLoading.value = true
    
    // 调用后端API更新网络
    const result = await updateDockerNetwork(activeDevice.value, editNetworkForm.value.networkID, editNetworkForm.value)
    
    if (result.success) {
      ElMessage.success('网络修改成功')
      editNetworkDialogVisible.value = false
      
      // 刷新网络列表
      await fetchDockerNetworks(activeDevice.value)
    } else {
      ElMessage.error(`修改网络失败: ${result.message}`)
    }
  } catch (error) {
    console.error('修改网络失败:', error)
    ElMessage.error(`修改网络失败: ${error.message}`)
  } finally {
    editNetworkLoading.value = false
  }
}

// 取消修改网络
const handleEditNetworkCancel = () => {
  editNetworkDialogVisible.value = false
  currentEditingNetwork.value = null
}

// 处理删除网络


// 标签页切换事件处理
// 标签页切换事件处理
const handleTabChange = (tabName) => {
  console.log('Tab changed to:', tabName)
  
  // 处理截图刷新的开启与暂停
  if (tabName === 'cloud-management') {
    // 切换回云机管理页面，尝试启动截图刷新
    // startScreenshotRefresh 内部会检查是否有选中的设备或云机
    startScreenshotRefresh()
    
    // 如果是坑位模式且没有选中设备，自动选择第一个在线设备
    if (cloudManageMode.value === 'slot' && !selectedCloudDevice.value) {
      const firstOnlineDevice = devices.value.find(device => {
        return devicesStatusCache.value.get(device.id) === 'online'
      })
      if (firstOnlineDevice) {
        console.log('切换到云机管理页面(坑位模式)，自动选择第一个在线设备:', firstOnlineDevice.name)
        selectedCloudDevice.value = firstOnlineDevice
        activeDevice.value = firstOnlineDevice
        fetchAndroidContainers(firstOnlineDevice, true).then(() => {
          initCloudMachineGroups()
        })
      }
    }
    
    // 如果有选中设备，截图轮询已通过 startScreenshotRefresh 启动，无需手动触发
  } else {
    // 离开云机管理页面，停止截图刷新
    stopScreenshotRefresh()
  }
  
  // 切换到机型管理页面时，触发机型列表数据获取
  if (tabName === 'model-management' && modelManagementRef.value) {
    console.log('切换到机型管理页面，触发机型列表数据获取')
    modelManagementRef.value.fetchPhoneModels()
  }

  // 切换到实例管理页面时，触发实例列表数据获取
  if (tabName === 'instance-management') {
    console.log('切换到实例管理页面，触发实例列表数据获取')
    instanceManagementRef.value.fetchInstances()
  }

  // 切换到网络管理页面时，触发网络列表数据获取
  if (tabName === 'network-management') {
    console.log('切换到网络管理页面，触发网络列表数据获取')
    networkManagementRef.value.fetchNetworks()
  }

  // 切换到备份管理页面时，触发备份列表数据获取
  if (tabName === 'backup-management') {
    console.log('切换到备份管理页面，触发备份列表数据获取')
    backupManagementRef.value.fetchBackups()
  }

  // 切换到AI助理页面时，触发AI助理数据获取
  if (tabName === 'ai-assistant') {
    console.log('切换到AI助理页面，触发AI助理数据获取')
    aiAssistantRef.value.fetchAiAssistant()
  }

  // 切换到RPA Agent页面时，触发RPA Agent数据获取
  if (tabName === 'rpa-agent') {
    console.log('切换到RPA Agent页面')
    rpaAgentRef.value?.fetchRpaAgent()
  }

  // 切换到OpenCecs页面时，触发初始化数据获取
  if (tabName === 'opencecs-management') {
    console.log('切换到OpenCecs页面，触发初始化数据获取')
    opencecsManagementRef.value?.init()
  }

  // 切换到客服页面时，触发资产信息上报
  if (tabName === 'customer-service') {
    console.log('切换到客服页面，触发资产信息上报')
    // 延迟一下确保组件已经渲染
    nextTick(() => {
      if (customerServiceRef.value && token.value) {
        // 重新生成 ticket
        if (customerServiceRef.value.initCustomerServiceUrl) {
          customerServiceRef.value.initCustomerServiceUrl()
        }
        if (customerServiceRef.value.reportUserAssets) {
          customerServiceRef.value.reportUserAssets()
        }
      }
    })
  }
}


// 监听选中的云机变化，取消勾选时停止相关截图请求
watch(selectedCloudMachines, (newSelected, oldSelected) => {
  if (oldSelected.length !== newSelected.length) {
    console.log('选中的云机变化:', oldSelected.length, '->', newSelected.length);
  }
  
  // 只有在云机管理标签页才处理截图刷新
  if (activeTab.value !== 'cloud-management') {
    return;
  }
  
  // 如果从无选中变为有选中，启动截图定时刷新
  if (oldSelected.length === 0 && newSelected.length > 0) {
    startScreenshotRefresh();
  }
  
  // 如果从有选中变为无选中，且仍在批量模式下（非模式切换导致的清空），停止截图刷新
  if (oldSelected.length > 0 && newSelected.length === 0 && cloudManageMode.value === 'batch') {
    stopScreenshotRefresh();
  }
}, { deep: true });

// 生命周期钩子
let refreshInterval = null
let hourlyRefreshInterval = null
let nextDeviceIndex = 0 // 用于半小时逐个设备刷新

onMounted(() => {
  // 从本地存储加载设备列表
  loadDevicesFromLocalStorage()
  // 从本地存储加载设备分组
  loadDeviceGroupsFromLocalStorage()
  // 发现设备
  discoverAndLoadDevices()

  // 初始化云机分组
  initCloudMachineGroups()
  // 自动缓存所有镜像，避免用户进入镜像管理界面还要手动刷新
  fetchImageList('')
  
  // 获取系统公告
  fetchAnnouncement()
  
  // 定时刷新设备数据（10秒）- 已禁用，避免频繁刷新
  // refreshInterval = setInterval(() => {
  //   refreshData()
  // }, 10000)
  
  // 每半小时刷新一次设备和云机列表（逐个设备）
  hourlyRefreshInterval = setInterval(() => {
    refreshDevicesContainersOneByOne()
  }, 30 * 60 * 1000) // 30分钟
  
  // 初始化版本检查队列 - 每2秒检查一个设备
  initVersionCheckQueue()
  
  // 初始化设备绑定状态查询队列 - 每1秒检查一次
  initBindCheckQueue()
  
  // 每10秒自动将所有设备添加到版本检查队列
  // versionRefreshInterval = setInterval(() => {
  //   autoGetAllDeviceVersions()
  // }, 10000) // 10秒
  
  // 初始加载时获取一次版本信息
  setTimeout(() => {
    autoGetAllDeviceVersions()
  }, 1000) // 延迟1秒执行，确保设备列表已加载
  
  // ========== 启动设备心跳检测服务 ==========
  // 延迟3秒启动，无论设备列表是否为空都启动服务
  heartbeatInitialized = false // 确保每次挂载都能重新初始化
  setTimeout(() => {
    console.log('[启动] 准备初始化设备心跳检测，当前设备数量:', devices.value.length)
    initDeviceHeartbeat()
  }, 3000) // 3秒启动，确保前端已准备好
  
  // 检查是否已有token，如果有则启动同步授权定时器
  if (token.value) {
    startSyncAuthTimer()
  }
  
  // 使用Wails的事件API注册下载进度事件监听器
  const initEventListeners = () => {
    if (!Events || !Events.On) {
      setTimeout(initEventListeners, 500)
      return
    }
    
    console.log('[事件监听器] 初始化下载进度事件监听器')
    
    // 注册事件监听器
    Events.On('download-progress', (data) => {
      // 严格验证：必须有活动的下载任务
      if (!isDownloadingImage.value || !currentDownloadTaskId.value || !currentDownloadImage.value) {
        return
      }
      
      let progress = 0
      if (data && data.data && typeof data.data.progress === 'number') {
        progress = data.data.progress
      } else if (data && typeof data.progress === 'number') {
        progress = data.progress
      }
      
      // 查找当前活动的下载任务
      const currentTask = taskQueue.value.find(t => 
        t.id === currentDownloadTaskId.value && 
        t.type === 'downloadImage' && 
        t.status === 'running'
      )
      
      if (!currentTask) {
        return
      }
      
      // 检查会话时间戳
      const taskSessionTime = currentTask.sessionStartTime || 0
      if (taskSessionTime !== downloadStartTime.value) {
        return
      }
      
      // 防止进度回退和异常跳跃
      const newProgress = Math.round(progress)
      const currentProgress = currentTask.progress || 0
      
      // 回退检测
      if (newProgress < currentProgress - 3) {
        return
      }
      
      // 跨度异常检测
      if (currentProgress > 20 && newProgress < 10) {
        return
      }
      
      // 异常跳跃检测
      if (newProgress - currentProgress > 30) {
        return
      }
      
      // 更新进度
      downloadProgress.value = progress
      currentTask.progress = newProgress
    })
    Events.On('download-complete', (data) => {
      handleDownloadComplete(data)
    })
    Events.On('upload-progress', (data) => {
      handleUploadProgress(data)
    })
    Events.On('upload-complete', (data) => {
      handleUploadComplete(data)
    })
  }
  
  // 延迟注册事件监听器，确保Events模块已加载
  setTimeout(initEventListeners, 100)
  // 设置BroadcastChannel监听器，处理来自投屏窗口的IPC调用
  const ipcChannel = new BroadcastChannel('wails-ipc-child');
  ipcChannel.onmessage = async (event) => {
    if (event.data && event.data.type === 'ipc-request') {
      const { funcName, args, requestId } = event.data;
      console.log('[Main] 收到投屏窗口IPC请求:', funcName, args);
      
      try {
        let result = null;
        
        // 根据函数名调用对应的后端方法
        switch (funcName) {
          case 'ToggleProjectionWindowTop':
            if (typeof ToggleProjectionWindowTop === 'function') {
              result = await ToggleProjectionWindowTop(args);
            }
            break;
          case 'ArrangeProjectionWindows':
            if (typeof ArrangeProjectionWindows === 'function') {
              result = await ArrangeProjectionWindows(args);
            }
            break;
          default:
            console.warn('[Main] 未知的函数名:', funcName);
        }
        
        // 发送响应回投屏窗口
        ipcChannel.postMessage({
          type: 'ipc-response',
          requestId: requestId,
          result: result
        });
      } catch (error) {
        console.error('[Main] IPC调用失败:', error);
        ipcChannel.postMessage({
          type: 'ipc-response',
          requestId: requestId,
          error: error.message
        });
      }
    }
  };
  
  // 将频道保存到窗口对象，以便后续清理
  window.$wailsIpcChannel = ipcChannel;
  
  // 设置事件通道监听器，处理来自投屏窗口的Wails事件
  const eventChannel = new BroadcastChannel('wails-events');
  eventChannel.onmessage = async (event) => {
    if (event.data && event.data.type === 'wails-event') {
      const { event: eventName, data } = event.data;
      console.log('[Main] 收到投屏窗口事件:', eventName, data);
      
      // 根据事件名称调用对应的后端方法
      switch (eventName) {
        case 'ToggleProjectionWindowTop':
          if (typeof ToggleProjectionWindowTop === 'function') {
            await ToggleProjectionWindowTop(data);
          }
          break;
        case 'ArrangeProjectionWindows':
          if (typeof ArrangeProjectionWindows === 'function') {
            await ArrangeProjectionWindows(data || {});
          }
          break;
        default:
          console.warn('[Main] 未处理的事件:', eventName);
      }
    }
  };
  
  // 将事件频道保存到窗口对象
  window.$wailsEventChannel = eventChannel;

  // 注册全局设备添加函数，供 opencecsManagement 等外部组件调用
  window.addDiscoveredDevice = (device) => {
    handleAddDevice(device)
  }

  // 注册全局设备移除函数，按 IP 移除设备（供 opencecsManagement 刷新时清理旧设备）
  window.removeDiscoveredDevice = (deviceIp) => {
    const idx = devices.value.findIndex(d => d.ip === deviceIp)
    if (idx !== -1) {
      const removed = devices.value[idx]
      devicesStatusCache.value.delete(removed.id)
      devicesLastUpdateTime.value.delete(removed.id)
      devices.value.splice(idx, 1)
      // 清除该设备的云机缓存，避免旧容器对象（携带旧 deviceIp）被复用
      deviceCloudMachinesCache.value.delete(deviceIp)
      // 如果当前正在查看的设备被移除，清空容器列表，避免旧 deviceIp 的容器残留
      if (activeDevice.value && activeDevice.value.ip === deviceIp) {
        instances.value = []
        allInstances.value = []
      }
      saveDevicesToLocalStorage()
      console.log(`[removeDiscoveredDevice] 已移除设备: ${deviceIp}`)
    }
  }

  // 注册全局按来源批量移除设备函数（供 OpenCecs 等模块清理所有注入的设备）
  window.removeDevicesBySource = (source) => {
    const toRemove = devices.value.filter(d => d.source === source)
    if (toRemove.length === 0) return
    for (const device of toRemove) {
      devicesStatusCache.value.delete(device.id)
      devicesLastUpdateTime.value.delete(device.id)
      deviceCloudMachinesCache.value.delete(device.ip)
      if (activeDevice.value && activeDevice.value.ip === device.ip) {
        instances.value = []
        allInstances.value = []
      }
    }
    devices.value = devices.value.filter(d => d.source !== source)
    saveDevicesToLocalStorage()
    console.log(`[removeDevicesBySource] 已移除 ${toRemove.length} 个 ${source} 设备`)
  }

  // 暴露 Go IPC 函数供 opencecsManagement 等外部组件调用
  window.goHttpRequest = HttpRequest
  window.goForceRefreshDeviceInfo = ForceRefreshDeviceInfo
  window.goGetDevicesStatus = GetDevicesStatus

  // 暴露设备列表给 opencecsManagement 退出登录时兜底查找
  window.getDevicesList = () => devices.value

  // 注册专用的 OpenCecs 设备全部清理函数（退出登录时调用，确保万无一失）
  window.removeAllOpenCecsDevicesFromHost = (extraIps = []) => {
    console.log('========== [App.vue] removeAllOpenCecsDevicesFromHost 被调用 ==========')
    console.log('[App.vue] 传入的 extraIps:', extraIps)
    const extraIpSet = new Set(extraIps)
    const before = devices.value.length
    console.log(`[App.vue] 当前设备总数: ${before}`)
    
    // 逐个设备检查匹配情况
    devices.value.forEach((d, i) => {
      const matchSource = d.source === 'opencecs'
      const matchName = d.name === 'opencecs'
      const matchIp = extraIpSet.has(d.ip)
      const matchPort = d.ip && d.ip.includes(':')
      const matched = matchSource || matchName || matchIp || matchPort
      console.log(`[App.vue] 设备[${i}]: ip=${d.ip}, name=${d.name}, source=${d.source}, id=${d.id} → ${matched ? '✅ 匹配删除' : '❌ 保留'}${matchSource ? ' (source)' : ''}${matchName ? ' (name)' : ''}${matchIp ? ' (IP)' : ''}${matchPort ? ' (IP:port格式)' : ''}`)
    })
    
    // 找出所有 OpenCecs 设备：按 source / name / IP列表 / IP:port格式 多重匹配
    const toRemove = devices.value.filter(d => {
      if (d.source === 'opencecs') return true
      if (d.name === 'opencecs') return true
      if (extraIpSet.has(d.ip)) return true
      // 兜底：OpenCecs 公网设备的 IP 格式为 publicIP:port（含冒号+端口号）
      // 正常局域网设备为纯 IP（如 10.10.11.46），不会含冒号
      if (d.ip && d.ip.includes(':')) return true
      return false
    })
    
    if (toRemove.length === 0) {
      console.log('[App.vue] devices.value 中无匹配设备')
      // 即使内存中没有，也要检查并清理 localStorage 中的残留
      try {
        const saved = localStorage.getItem('edgeclient_devices')
        if (saved) {
          const savedDevices = JSON.parse(saved)
          const cleaned = savedDevices.filter(d => {
            if (d.source === 'opencecs') return false
            if (d.name === 'opencecs') return false
            if (extraIpSet.has(d.ip)) return false
            if (d.ip && d.ip.includes(':')) return false
            return true
          })
          if (cleaned.length < savedDevices.length) {
            localStorage.setItem('edgeclient_devices', JSON.stringify(cleaned))
            // 同步到 devices.value
            devices.value = cleaned
            console.log(`[App.vue] ✅ 已从 localStorage 清理 ${savedDevices.length - cleaned.length} 个 opencecs 残留设备`)
            return savedDevices.length - cleaned.length
          }
        }
      } catch (e) {
        console.error('[App.vue] 清理 localStorage 残留失败:', e)
      }
      console.log('[App.vue] ⚠️ localStorage 中也无 opencecs 残留设备')
      return 0
    }
    
    // 清理缓存
    for (const device of toRemove) {
      devicesStatusCache.value.delete(device.id)
      devicesLastUpdateTime.value.delete(device.id)
      deviceCloudMachinesCache.value.delete(device.ip)
      if (activeDevice.value && activeDevice.value.ip === device.ip) {
        activeDevice.value = null
        instances.value = []
        allInstances.value = []
      }
    }
    
    // 从列表中移除
    const removeIps = new Set(toRemove.map(d => d.ip))
    const removeIds = new Set(toRemove.map(d => d.id))
    devices.value = devices.value.filter(d => !removeIps.has(d.ip) && !removeIds.has(d.id))
    
    saveDevicesToLocalStorage()
    console.log(`[App.vue] ✅ 已移除 ${toRemove.length} 个设备 (${before} → ${devices.value.length})，IP: ${toRemove.map(d => d.ip).join(', ')}`)
    return toRemove.length
  }
})

// 处理容器操作
const handleContainerAction = async (container, action) => {
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
  }

  if (!targetDevice) {
    ElMessage.error('没有选中设备')
    return
  }

  loading.value = true
  try {
    switch (action) {
      case 'restart':
        // 重启容器前，清空该容器的截图缓存，避免显示旧截图
        clearContainerScreenshotCache(targetDevice, container)
        
        // 重启容器：V3设备使用/android/restart API，V0-V2设备使用stop+start
        await authRetry(targetDevice, async (password) => {
          await restartAndroidContainer(targetDevice, container.name || container.ID, password)
        })
        ElMessage.success('重启成功')
        break
      case 'stop':
        await authRetry(targetDevice, async (password) => {
          await stopContainer(targetDevice, container.name || container.ID, password)
        })
        ElMessage.success('关闭成功')
        break
      case 'delete':
        // 添加删除确认提示
        await ElMessageBox.confirm('确定要删除此容器吗？此操作不可恢复。', '删除容器', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        await authRetry(targetDevice, async (password) => {
          await deleteContainer(targetDevice, container.name || container.ID, password)
        })
        ElMessage.success('删除成功')
        break
      default:
        ElMessage.error('未知操作')
        break
    }
    // 操作成功后刷新容器列表
    await fetchAndroidContainers(targetDevice, true)
  } catch (error) {
    console.error(`${action}容器失败:`, error)
    const errorMsg = error.response?.data?.message || error.message || `${action}失败`
    ElMessage.error(errorMsg)
  } finally {
    loading.value = false
  }
}

// 右键菜单相关函数
const handleContextMenu = (event, slot) => {
  event.preventDefault()
  // 添加全局点击事件监听，用于关闭菜单
  // 使用 setTimeout 确保当前点击不会立即触发关闭
  setTimeout(() => {
    window.addEventListener('click', closeContextMenu)
  }, 0)
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  
  // 处理 slot 参数，它可能是索引(坑位模式)或对象(批量模式)
  let container = null
  let slotIndex = slot
  
  if (typeof slot === 'object' && slot !== null) {
    // 批量模式下直接传入了对象
    container = slot
    slotIndex = slot.indexNum // 尝试获取索引用于后续逻辑
  } else if (cloudManageMode.value === 'slot') {
    // 坑位模式：从当前设备的容器列表中查找
    container = instances.value.find(inst => inst.indexNum === slot)
    slotIndex = slot
  } else {
    // 批量模式但传入了索引（防御性代码）
    container = selectedCloudMachines.value.find(machine => machine.indexNum === slot)
    slotIndex = slot
  }
  
  contextMenuSlot.value = slotIndex // 存储索引或标识符
  contextMenuContainer.value = container // 存储解析出的容器对象
  
  // 关键修复：批量模式下自动修正 activeDevice
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    // 如果当前没有 activeDevice，或者 activeDevice 与当前操作的容器不属于同一设备
    if (!activeDevice.value || activeDevice.value.ip !== container.deviceIp) {
      const device = devices.value.find(d => d.ip === container.deviceIp)
      if (device) {
        activeDevice.value = device
        console.log('批量模式下自动切换 activeDevice:', device)
      } else {
        // 如果找不到设备对象，构造临时对象
        activeDevice.value = { 
          ip: container.deviceIp, 
          version: container.deviceVersion || 'v3',
          name: container.deviceName || 'Unknown'
        }
        console.log('批量模式下设置临时 activeDevice:', activeDevice.value)
      }
    }
  }

  console.log('Context menu opened for slot:', slot, 'container:', container, 'mode:', cloudManageMode.value)
  contextMenuVisible.value = true

  // 调整菜单位置，防止超出屏幕底部
  nextTick(() => {
    if (contextMenuRef.value) {
      const menu = contextMenuRef.value
      const { height } = menu.getBoundingClientRect()
      const { innerHeight } = window
      
      // 如果菜单超出底部，向上显示
      if (contextMenuPosition.value.y + height > innerHeight) {
        let newY = event.clientY - height
        // 确保不超出顶部
        if (newY < 0) newY = 0
        contextMenuPosition.value.y = newY
      }
    }
  })
}

const closeContextMenu = () => {
  contextMenuVisible.value = false
  // 移除全局点击事件监听
  window.removeEventListener('click', closeContextMenu)
  // 清空上下文菜单相关的引用
  contextMenuContainer.value = null
}

onBeforeUnmount(() => {
  window.removeEventListener('click', closeContextMenu)
  
  // 停止设备心跳检测
  if (deviceHeartbeatTimer) {
    clearInterval(deviceHeartbeatTimer)
    console.log('[心跳] 已停止设备心跳检测定时器')
  }
  // 停止安卓缓存轮询
  if (androidCacheTimer) {
    clearInterval(androidCacheTimer)
    console.log('[安卓轮询] 已停止安卓缓存轮询定时器')
  }
})

// 获取当前上下文菜单对应的云机
const getCurrentContextMenuContainer = () => {
  // 优先直接返回已存储的容器对象（这是最准确的，特别是在批量模式下）
  if (contextMenuContainer.value) {
    return contextMenuContainer.value
  }

  if (cloudManageMode.value === 'slot') {
    // 坑位模式：从当前设备的容器列表中查找
    return instances.value.find(inst => inst.indexNum === contextMenuSlot.value)
  } else {
    // 批量模式：从所有选中的云机列表中查找
    return selectedCloudMachines.value.find(machine => machine.indexNum === contextMenuSlot.value)
  }
}

const handleUpdateImage = () => {
  const container = getCurrentContextMenuContainer()
  if (container) {
    showUpdateImageDialog(container)
  }
  closeContextMenu()
}

const handleRestart = () => {
  const container = getCurrentContextMenuContainer()
  if (container) {
    handleContainerAction(container, 'restart')
  }
  closeContextMenu()
}

const setS5Agent = async () => {
  const container = getCurrentContextMenuContainer()
  if(container) {
    // 初始化表单数据（先使用缓存数据，避免闪烁）
    s5ProxyForm.value = {
      cloudMachineName: formatInstanceName(container.name) || container.name,
      s5ServerAddress: container.s5IP || '',
      s5Port: container.s5Port || '',
      username: container.s5User || '',
      password: container.s5Password || '',
      dnsMode: container.s5Type == 0 || container.s5Type == 2 ? 'server' : 'local'
    }
    
    // 显示S5代理设置弹窗
    s5ProxyDialogVisible.value = true
    
    // 每次都调用接口获取最新的s5代理详情
    try {
      // 构造请求参数
      let host = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip
        ? container.ip
        : container.deviceIp
      // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
      if (host && host.includes(':')) host = host.split(':')[0]
      const port = extractPort9082(container) || 9082
      const requestParams = {
        url: `http://${host}:${port}/proxy`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        },
        body: ''
      };
      
      console.log('正在获取最新S5代理详情...');
      // 调用Wails后端的HttpRequest函数获取s5代理详情
      const response = await HttpRequest(requestParams);
      console.log('获取S5代理详情响应:', response);
      
      // 处理响应
      let result;
      if (response.body && typeof response.body === 'object') {
        result = response.body;
      } else {
        // 如果body不是JSON对象，尝试解析raw字段
        try {
          result = JSON.parse(response.raw);
        } catch (e) {
          result = { code: response.status };
        }
      }

      console.log('获取S5代理详情结果:', result);
      
      // 更新表单数据为最新数据
      if (result.code === 200 || result.success) {
        parseAndFillSocks5Url(result.data.addr)
        s5ProxyForm.value.dnsMode = result.data.type == 1 ? 'local' : 'server';
        console.log('S5代理信息已更新为最新数据');
      }
    } catch (error) {
      console.error('获取S5代理详情失败:', error);
      // 获取失败不影响弹窗显示，继续使用缓存数据
      console.log('使用缓存的S5代理信息');
    }
  }
  closeContextMenu()
}

const parseAndFillSocks5Url = (url) => {
  // 检查url是否为空或undefined
  if (!url) {
    return false;
  }
  
  const socks5Regex = /^socks5:\/\/(?:([^:]+):([^@]+)@)?([^:]+):(\d+)$/;
  let match = url.match(socks5Regex);
  if (match) {
      // 根据匹配结果提取各部分信息
      // match[1] = username, match[2] = password, match[3] = host, match[4] = port
      s5ProxyForm.value.username = match[1] || '';
      s5ProxyForm.value.password = match[2] || '';
      s5ProxyForm.value.s5ServerAddress = match[3] || '';
      s5ProxyForm.value.s5Port = match[4] || '';
      
      return true;
  }
  return false;
}

// 解析S5信息并自动填写
const parseVpcInfo = () => {
  const vpcInfo = s5ProxyForm.value.vpcInfo?.trim()
  
  if (!vpcInfo) {
    return
  }
  
  // 格式: 地址:端口:用户名:密码 (用户名和密码可选)
  const parts = vpcInfo.split(':')
  
  if (parts.length < 2) {
    ElMessage.warning('S5信息格式不正确，至少需要地址和端口')
    return
  }
  
  // 解析地址和端口
  s5ProxyForm.value.s5ServerAddress = parts[0] || ''
  s5ProxyForm.value.s5Port = parts[1] || ''
  
  // 解析用户名（可选，第3部分）
  if (parts.length >= 3 && parts[2]) {
    s5ProxyForm.value.username = parts[2]
  } else {
    s5ProxyForm.value.username = ''
  }
  
  // 解析密码（可选，第4部分）
  if (parts.length >= 4 && parts[3]) {
    s5ProxyForm.value.password = parts[3]
  } else {
    s5ProxyForm.value.password = ''
  }
  
  ElMessage.success('S5信息解析成功')
}

const handleS5ProxySubmit = async () => {
  const container = getCurrentContextMenuContainer()
  if(!container) return
  
  // 表单验证
  if (!s5ProxyForm.value.s5ServerAddress) {
    ElMessage.error('请输入s5服务器地址')
    return
  }
  
  if (!s5ProxyForm.value.s5Port) {
    ElMessage.error('请输入s5端口')
    return
  }
  
  try {
    s5ProxyLoading.value = true
    
    // 准备请求参数
    const dnsModeValue = s5ProxyForm.value.dnsMode === 'local' ? '1' : '2'

    let host = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip
      ? container.ip
      : container.deviceIp
    // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
    if (host && host.includes(':')) host = host.split(':')[0]
    const port = extractPort9082(container) || 9082
    const requestParams = {
      url: `http://${host}:${port}/proxy?cmd=2&ip=${s5ProxyForm.value.s5ServerAddress}&port=${s5ProxyForm.value.s5Port}&usr=${s5ProxyForm.value.username}&pwd=${s5ProxyForm.value.password}&type=${dnsModeValue}`,
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
      body: ''
    };
    
    // 调用Wails后端的HttpRequest函数
    const response = await HttpRequest(requestParams);
    console.log('设置SOCKS5代理请求响应:', response);
    
    // 处理响应
    let result;
    if (response.body && typeof response.body === 'object') {
      result = response.body;
    } else {
      // 如果body不是JSON对象，尝试解析raw字段
      try {
        result = JSON.parse(response.raw);
      } catch (e) {
        result = { code: response.status };
      }
    }
    
    // 处理响应结果
    if (result.code === 200 || result.success) {
      ElMessage.success('S5代理设置成功')
      s5ProxyDialogVisible.value = false
    } else {
      ElMessage.error('S5代理设置失败: ' + (result.msg || result.message || '未知错误'))
    }
  } catch (error) {
    console.error('设置S5代理失败:', error)
    ElMessage.error('设置S5代理失败: ' + error.message)
  } finally {
    s5ProxyLoading.value = false
  }
}

const closeS5Agent = async  () => { 
  const container = getCurrentContextMenuContainer()
  console.log('关闭S5代理:', container, extractPort9082(container) || 9082)
  if(container) {
    try {
      let host = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip
        ? container.ip
        : container.deviceIp
      // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
      if (host && host.includes(':')) host = host.split(':')[0]
      const port = extractPort9082(container) || 9082
     // 构造请求参数
      const requestParams = {
        url: `http://${host}:${port}/proxy?cmd=3`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        },
        body: ''
      };
      
      // 调用Wails后端的HttpRequest函数
      const response = await HttpRequest(requestParams);
      console.log('关闭SOCKS5代理请求响应:', response);
      
      // 处理响应
      let result;
      if (response.body && typeof response.body === 'object') {
        result = response.body;
      } else {
        // 如果body不是JSON对象，尝试解析raw字段
        try {
          result = JSON.parse(response.raw);
        } catch (e) {
          result = { code: response.status };
        }
      }
      console.log('关闭SOCKS5代理响应:', result)
      
      // 根据API返回结果显示相应消息
      if (result.code == 200 || result.success) {
        ElMessage.success('SOCKS5代理已关闭')
      } else {
        ElMessage.warning(`关闭SOCKS5代理失败: ${result.msg || result.message || '未知错误'}`)
      }
    } catch (error) {
      console.error('关闭SOCKS5代理失败:', error)
      ElMessage.error(`关闭SOCKS5代理失败: ${error.message}`)
    }
  }
  closeContextMenu()
}

const handleDelete = () => {
  const container = getCurrentContextMenuContainer()
  if (container) {
    handleContainerAction(container, 'delete')
  }
  closeContextMenu()
}

// 计算每个坑位的备份数量
const getBackupCount = (slotNum) => {
  return allInstances.value.filter(inst => 
    inst.indexNum === slotNum && 
    inst.status === 'shutdown'
  ).length
}

const handleResetContainer = async () => {
  const container = getCurrentContextMenuContainer()
  console.log('handleResetContainer called with container:', container)
  
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
  }

  // 确保 targetDevice 包含 androidType，供 resetAndroidContainer 使用
  if (targetDevice && container) {
    targetDevice.androidType = container.androidType
  }
  
  console.log('targetDevice:', targetDevice)
  
  if (container && targetDevice && targetDevice.version === 'v3') {
    try {
      // 使用自定义弹窗让用户选择重置后是否开机
      let startAfterReset = true
      await new Promise((resolve, reject) => {
        ElMessageBox.confirm(
          '确定要重置此容器吗？重置后容器将会被恢复到初始状态。',
          '重置容器',
          {
            confirmButtonText: '重置并开机',
            cancelButtonText: '仅重置(不开机)',
            distinguishCancelAndClose: true,
            type: 'warning',
            showClose: true
          }
        ).then(() => {
          startAfterReset = true
          console.log('[重置容器] 用户选择: 重置并开机, start=true')
          resolve()
        }).catch((action) => {
          console.log('[重置容器] catch action:', action, typeof action)
          if (action === 'close') {
            // 用户点了 X 或按 ESC，取消操作
            reject('cancel')
          } else {
            // 用户点了"仅重置(不开机)"按钮
            startAfterReset = false
            console.log('[重置容器] 用户选择: 仅重置(不开机), start=false')
            resolve()
          }
        })
      })
      
      // 重置容器前，清空该容器的截图缓存，避免显示旧截图
      clearContainerScreenshotCache(targetDevice, container)
      
      // 调用重置容器API
      const result = await resetAndroidContainer(targetDevice, container.name, null, startAfterReset)
      
      if (result.code === 0) {
        ElMessage.success('容器重置成功')
        // 刷新容器列表
        await fetchAndroidContainers(targetDevice, true)
      } else {
        ElMessage.error(`容器重置失败: ${result.message || '未知错误'}`)
      }
    } catch (error) {
      if (error !== 'cancel') {
        console.error('重置容器失败:', error)
        ElMessage.error(`重置容器失败: ${error.message || '未知错误'}`)
      }
    }
  } else if (targetDevice && targetDevice.version !== 'v3') {
    ElMessage.warning('该设备版本不支持重置容器功能')
  } else {
    ElMessage.warning('请先选择设备')
  }
  closeContextMenu()
}

// 设置推流
const handleSetStream = async () => {
  try {
    // 获取当前容器
    const container = getCurrentContextMenuContainer()
    if (!container) {
      console.warn('设置推流：无法获取容器信息')
      // 重置表单
      streamType.value = ''
      streamFilePath.value = ''
      rtmpUrl.value = ''
      setStreamDialogVisible.value = true
      return
    }

    // 确定目标设备
    let targetDevice = activeDevice.value
    if (cloudManageMode.value === 'batch' && container.deviceIp) {
      targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
    }

    if (!targetDevice) {
      console.warn('设置推流：无法获取目标设备')
      // 重置表单
      streamType.value = ''
      streamFilePath.value = ''
      rtmpUrl.value = ''
      setStreamDialogVisible.value = true
      return
    }

    // 获取 9082 端口（使用 extractPort9082 自动处理公网设备端口映射）
    const port = extractPort9082(container) || 9082
    let host = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip
      ? container.ip
      : targetDevice.ip
    // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
    if (host && host.includes(':')) host = host.split(':')[0]

    // 调用停止摄像头接口
    const stopUrl = `http://${host}:${port}/camera?cmd=stop`
    console.log('停止摄像头接口:', stopUrl)
    
    try {
      await HttpRequest({
        url: stopUrl,
        method: 'GET'
      })
      console.log('摄像头已停止')
    } catch (error) {
      console.warn('停止摄像头失败（可能未启动）:', error)
      // 忽略错误，继续显示设置推流对话框
    }
  } catch (error) {
    console.error('设置推流准备失败:', error)
  }

  // 重置表单
  streamType.value = ''
  streamFilePath.value = ''
  rtmpUrl.value = ''
  setStreamDialogVisible.value = true
}

// 选择流类型变化处理
// const handleStreamTypeChange = (type) => {
//   streamFilePath.value = ''
//   rtmpUrl.value = ''
// }

// 选择文件夹处理
const selectStreamFolder = async () => {
  try {
    let result
    if (streamType.value === 'image') {
      result = await SelectImageFile()
    } else if (streamType.value === 'video') {
      result = await SelectVideoFile()
    } else {
      return
    }
    
    if (result && result.success && result.path) {
      streamFilePath.value = result.path
    }
  } catch (error) {
    console.error('选择文件失败:', error)
    ElMessage.error('选择文件失败')
  }
}

// 安装APP处理
// const handleInstallApp = async () => {
  // TODO: 实现APP安装逻辑
  // ElMessage.info('APP安装功能开发中')
// }

const qrCodeUrl = ref('')
const appDownloadQrCodeUrl = ref('')
const qrCodeLoading = ref(false)

// 生成APP连接二维码
const generateAppQRCode = async () => {
  qrCodeLoading.value = true
  try {
    const container = getCurrentContextMenuContainer()
    if (!container) {
      // 可能是批量操作或其他情况，暂时忽略或提示
      console.warn('生成二维码失败：无法获取容器信息')
      return
    }

    let n = container.name || ''
    // 处理名称显示，将长ID替换为4位随机数
    if (n && n.includes('_')) {
      const parts = n.split('_')
      // 如果第一部分是长ID（比如长度大于10），则替换为4位随机数
      if (parts.length >= 2 && parts[0].length > 10) {
         const randomPrefix = Math.floor(1000 + Math.random() * 9000)
         const suffix = n.substring(n.indexOf('_'))
         n = `${randomPrefix}${suffix}`
      }
    }
    let ip = container.networkName == 'myt' ? container.ip : activeDevice.value?.ip || ''
    // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
    if (ip && ip.includes(':')) ip = ip.split(':')[0]
    
    // 使用 extractPort 自动处理公网端口映射（OpenCecs 设备会返回映射后的公网端口）
    const t = extractPort(container, 10000) || 10000
    const u = extractPort(container, 10001) || 10001
    const i = extractPort(container, 9082) || 9082
    
    // c: tcp 2375 或 10008
    // let c = extractPort(container, 2375)
    // if (!c) c = extractPort(container, 10008)
    
    const ct = extractPort(container, 10006) || 10006
    const cu = extractPort(container, 10007) || 10007

    const rawStr = `n=${n}&ip=${ip}&t=${t}&u=${u}&i=${i}&ct=${ct}&cu=${cu}`
    console.log('二维码原始字符串:', rawStr)
    
    // Base64 编码 (处理中文)
    const base64Str = CryptoJS.enc.Base64.stringify(CryptoJS.enc.Utf8.parse(rawStr))
    
    qrCodeUrl.value = await QRCode.toDataURL(base64Str)

    // 生成APP下载二维码
    const downloadUrl = 'http://d.moyunteng.com/download/mytcloudphone_v1.3_20260302.apk'
    appDownloadQrCodeUrl.value = await QRCode.toDataURL(downloadUrl)

  } catch (error) {
    console.error('生成二维码失败:', error)
    ElMessage.error('生成二维码失败')
  } finally {
    qrCodeLoading.value = false
  }
}

// 监听流类型变化
watch(streamType, async (newType) => {
  // 切换类型时清空文件路径
  streamFilePath.value = ''
  
  if (newType === 'app') {
    await generateAppQRCode()
  } else {
    qrCodeUrl.value = ''
    appDownloadQrCodeUrl.value = ''
  }
})

// 确认设置推流
const confirmSetStream = async () => {
  if (!streamType.value) {
    ElMessage.warning('请选择推流类型')
    return
  }
  
  if (streamType.value === 'app') {
    // APP模式下，只是为了展示二维码，点击确定后直接关闭即可
    setStreamDialogVisible.value = false
    return
  }

  if ((streamType.value === 'image' || streamType.value === 'video') && !streamFilePath.value) {
    ElMessage.warning('请选择文件')
    return
  }
  
  if (streamType.value === 'rtmp' && !rtmpUrl.value) {
    ElMessage.warning('请输入RTMP推流地址')
    return
  }
  
  try {
    setStreamLoading.value = true
    const container = getCurrentContextMenuContainer()
    if (!container) {
      ElMessage.error('无法获取当前容器信息')
      return
    }

    if (streamType.value === 'image' || streamType.value === 'video') {
      // 上传文件到云机
      await authRetry(activeDevice.value, async (password) => {
        const result = await UploadFileToCloudMachine(
          activeDevice.value.ip,
          activeDevice.value.version || 'v3',
          container.name,
          streamFilePath.value,
          password
        )
        
        if (!result.success) {
          if (result.message && (result.message.includes('Authentication Failed') || result.message.includes('401'))) {
            throw new Error('Authentication Failed')
          }
          throw new Error(result.message)
        }

        // 上传成功后调用 modifydev 接口
        const fileName = streamFilePath.value.split(/[/\\]/).pop()
        const targetPath = `/storage/emulated/0/upload/${fileName}`
        let host = (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') && container.ip
          ? container.ip
          : activeDevice.value.ip
        // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
        if (host && host.includes(':')) host = host.split(':')[0]
        const modifyDevUrl = `http://${host}:${extractPort9082(container) || 9082}/modifydev`
        
        // 使用 x-www-form-urlencoded 格式发送数据
        // 注意：这里手动拼接字符串，避免 URLSearchParams 对路径进行编码，因为后端可能不支持解码
        const bodyStr = `cmd=4&type=${streamType.value}&path=${targetPath}`

        console.log('调用modifydev接口:', modifyDevUrl, bodyStr)

        const modifyResult = await HttpRequest({
          url: modifyDevUrl,
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: bodyStr
        })

        console.log('modifydev接口返回结果:', modifyResult)
 
        if (!modifyResult.success) {
          throw new Error(`推流设置接口调用失败: ${modifyResult.status}`)
        }
      })
    }
    
    // TODO: 调用后端API设置推流
    ElMessage.success('推流设置成功')
    setStreamDialogVisible.value = false
  } catch (error) {
    if (error !== 'cancel') {
      console.error('设置推流失败:', error)
      ElMessage.error(`设置推流失败: ${error.message || '未知错误'}`)
    }
  } finally {
    setStreamLoading.value = false
  }
}

// 取消设置推流
const cancelSetStream = () => {
  setStreamDialogVisible.value = false
  streamType.value = ''
  streamFilePath.value = ''
  rtmpUrl.value = ''
}
const handleShutdown = () => {
  const container = getCurrentContextMenuContainer()
  if (container) {
    handleContainerAction(container, 'stop')
  }
  closeContextMenu()
}

// MacVlanIP相关状态
const macVlanDialogVisible = ref(false)
const macVlanForm = ref({
  name: '', // 容器名称
  ip: ''    // MacVlanIP
})
const macVlanLoading = ref(false)

// 处理设置MacVlanIP
const handleSetMacVlanIP = () => {
  const container = getCurrentContextMenuContainer()
  if (!container) {
    ElMessage.warning('未找到选中的云机')
    return
  }
  
  // 初始化表单
  macVlanForm.value.name = container.name || container.ID || ''
  macVlanForm.value.ip = container.macVlanIP || '' 
  
  macVlanDialogVisible.value = true
  closeContextMenu()
}

// 提交MacVlanIP
const confirmSetMacVlanIP = async () => {
  if (!macVlanForm.value.ip) {
    ElMessage.warning('请输入MacVlanIP')
    return
  }
  
  // 简单的IP格式校验
  const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/
  if (!ipRegex.test(macVlanForm.value.ip)) {
    ElMessage.warning('请输入有效的IP地址')
    return
  }
  
  try {
    macVlanLoading.value = true
    
    // 确定目标设备
    let targetDevice = activeDevice.value
    if (cloudManageMode.value === 'batch' && !targetDevice) {
       // 批量模式下如果没有activeDevice，尝试从container获取deviceIp
       const container = getCurrentContextMenuContainer()
       if (container && container.deviceIp) {
         targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
       }
    }
    
    if (!targetDevice) {
      ElMessage.error('无法确定目标设备')
      return
    }

    await authRetry(targetDevice, async (password) => {
      await setMacVlanIp(targetDevice, macVlanForm.value.name, macVlanForm.value.ip, password)
      ElMessage.success('设置MacVlanIP成功')
      macVlanDialogVisible.value = false
    })
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('设置MacVlanIP失败:', error)
      ElMessage.error(`设置MacVlanIP失败: ${error.message || '未知错误'}`)
    }
  } finally {
    macVlanLoading.value = false
  }
}

// 一键新机 (for androidType === 'V2')
const handleOneKeyNewDevice = async () => {
  const container = getCurrentContextMenuContainer()
  console.log('handleOneKeyNewDevice called with container:', container)
  
  if (!container) {
    ElMessage.warning('未找到容器信息')
    closeContextMenu()
    return
  }
  
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
  }
  
  try {
    const loadingMsg = ElMessage({
      message: '正在执行一键新机...',
      type: 'info',
      duration: 0
    })
    
    // 判断使用哪个IP和端口
    let host, port
    if (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt') {
      // myt网络：使用容器IP + 9082端口
      host = container.ip
      port = 9082
    } else {
      // 非myt网络：使用端口映射
      host = targetDevice.ip
      // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
      if (host && host.includes(':')) host = host.split(':')[0]
      port = extractPort9082(container) || 9082
    }
    
    const modifyDevUrl = `http://${host}:${port}/modifydev?cmd=2`
    console.log('调用一键新机API:', modifyDevUrl)
    
    const result = await HttpRequest({
      url: modifyDevUrl,
      method: 'GET'
    })
    
    ElMessage.closeAll()
    
    if (result.success) {
      ElMessage.success('一键新机成功')
      // 刷新容器列表
      if (targetDevice) {
        await fetchAndroidContainers(targetDevice, true)
      }
    } else {
      ElMessage.error(`一键新机失败: ${result.status}`)
    }
  } catch (error) {
    ElMessage.closeAll()
    console.error('一键新机失败:', error)
    ElMessage.error(`一键新机失败: ${error.message || '未知错误'}`)
  } finally {
    closeContextMenu()
  }
}

const handleSwitchModel = async () => {
  const container = getCurrentContextMenuContainer()
  console.log('handleSwitchModel called with container:', container)
  
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && container && container.deviceIp) {
    // 在批量模式下，根据 deviceIp 查找设备
    targetDevice = devices.value.find(d => d.ip === container.deviceIp) || { ip: container.deviceIp, version: 'v3' }
  }
  console.log('targetDevice:', targetDevice)
  
  if (container && targetDevice && targetDevice.version === 'v3') {
    try {
      // 显示加载状态提示
      const loadingMsg = ElMessage({
        message: '加载机型列表中...',
        type: 'info',
        duration: 0
      })
      
      // 先获取可用的手机型号列表
      console.log('Calling getV3PhoneModels with deviceIP:', targetDevice.ip)
      await getV3PhoneModels(targetDevice.ip)
      
      console.log('Phone models fetched:', phoneModels.value)
      
      // if (phoneModels.value.length === 0) {
      //   ElMessage.warning('未获取到可用的机型列表')
      //   closeContextMenu()
      //   return
      // }
      
      // 重置临时变量
      tempModelName.value = ''
      tempModelId.value = ''
      switchModelType.value = 'online' // 重置为默认线上机型
      currentSwitchContainer.value = container
      
      // 打开机型切换对话框
      switchModelDialogVisible.value = true
    } catch (error) {
      if (error !== 'cancel') {
        console.error('切换机型失败:', error)
        ElMessage.error(`切换机型失败: ${error.message || '未知错误'}`)
      }
    } finally {
      // 关闭加载提示
      setTimeout(() => {
        ElMessage.closeAll()
      }, 100)
    }
  } else if (targetDevice && targetDevice.version !== 'v3') {
    ElMessage.warning('该设备版本不支持切换机型功能')
  } else {
    ElMessage.warning('请先选择设备')
  }
  closeContextMenu()
}


const confirmSwitchModel = async () => {
  if (!tempModelId.value) {
    ElMessage.warning('请选择机型')
    return
  }
  
  // 确定目标设备
  let targetDevice = activeDevice.value
  if (cloudManageMode.value === 'batch' && currentSwitchContainer.value && currentSwitchContainer.value.deviceIp) {
    targetDevice = devices.value.find(d => d.ip === currentSwitchContainer.value.deviceIp) || { ip: currentSwitchContainer.value.deviceIp, version: 'v3' }
  }
  
  if (currentSwitchContainer.value && targetDevice) {
    try {
      // 开始切换，设置加载状态
      switchingModel.value = true
      
      let finalModelId = tempModelId.value
      
      // 处理随机机型
      if (finalModelId === 'random') {
        let list = []
        switch (switchModelType.value) {
          case 'online': 
            list = phoneModels.value; 
            if (currentSwitchContainer.value && currentSwitchContainer.value.image) {
              const currentImageUrl = currentSwitchContainer.value.image
              const currentImage = imageList.value.find(img => img.url === currentImageUrl)
              if (currentImage && currentImage.os_ver) {
                const verMatch = currentImage.os_ver.match(/and(\d+)/i)
                if (verMatch && verMatch[1]) {
                  const targetVer = verMatch[1]
                  const filteredList = list.filter(m => {
                    if (!m.android_version) return true // 无该字段时不过滤
                    return String(m.android_version) === String(targetVer)
                  })
                  if (filteredList.length > 0) {
                    list = filteredList
                  }
                }
              }
            }
            break;
          case 'local': list = localPhoneModels.value; break;
          case 'backup': list = backupPhoneModels.value; break;
        }
        
        if (!list || list.length === 0) {
          ElMessage.warning('当前列表为空，无法随机选择')
          switchingModel.value = false
          return
        }
        
        const randomModel = list[Math.floor(Math.random() * list.length)]
        
        // 根据不同类型获取对应的值
        if (switchModelType.value === 'online') {
          finalModelId = randomModel.id
        } else {
          // 本地和备份机型使用name
          finalModelId = randomModel.name
        }
        
        console.log(`随机选择了机型: ${randomModel.name} (${finalModelId})`)
      }
      
      // 调用切换机型API
      const result = await switchPhoneModel(
        targetDevice, 
        currentSwitchContainer.value.name, 
        {
          value: finalModelId,
          type: switchModelType.value
        }
      )
      
      if (result.code === 0) {
        ElMessage.success('机型切换成功')
        // 刷新容器列表
        await fetchAndroidContainers(targetDevice, true)
      } else {
        ElMessage.error(`机型切换失败: ${result.message || '未知错误'}`)
      }
    } catch (error) {
      console.error('切换机型失败:', error)
      ElMessage.error(`切换机型失败: ${error.message || '未知错误'}`)
    } finally {
      // 关闭对话框
      switchModelDialogVisible.value = false
      tempModelName.value = ''
      tempModelId.value = ''
      currentSwitchContainer.value = null
      switchingModel.value = false
    }
  }
}

const cancelSwitchModel = () => {
  // 关闭对话框并重置变量
  switchModelDialogVisible.value = false
  tempModelName.value = ''
  tempModelId.value = ''
  currentSwitchContainer.value = null
}


const handleSwitchBackup = () => {
  showBackupList(contextMenuSlot.value)
  closeContextMenu()
}



// 处理文件上传
const handleFileUpload = () => {
  console.log('handleFileUpload called')
  console.log('contextMenuContainer.value:', contextMenuContainer.value)
  console.log('activeDevice.value:', activeDevice.value)
  console.log('cloudMachines.value:', cloudMachines.value)
  
  const container = contextMenuContainer.value
  console.log('Using container from contextMenuContainer.value:', container)
  
  if (container) {
    contextMenuContainerId.value = container.containerId || container.id || container.indexNum || contextMenuSlot.value
    console.log('Set contextMenuContainerId.value:', contextMenuContainerId.value)
    
    // 确保我们有设备信息
    if (!activeDevice.value || !activeDevice.value.ip) {
      // 尝试从容器信息中获取设备IP
      let deviceIp = null
      
      // 检查容器对象的各种可能的设备IP属性
      if (container.deviceIp) {
        deviceIp = container.deviceIp
      } else if (container.device_ip) {
        deviceIp = container.device_ip
      } else if (container.ip) {
        deviceIp = container.ip
      }
      
      console.log('Extracted deviceIp:', deviceIp)
      
      if (deviceIp) {
        activeDevice.value = {
          ip: deviceIp,
          version: 'v3' // 默认版本
        }
        console.log('Set activeDevice.value:', activeDevice.value)
      } else {
        // 如果无法从容器中获取设备IP，尝试使用云机列表中的信息
        const cloudMachine = cloudMachines.value.find(cm => cm.indexNum === container.indexNum)
        if (cloudMachine && cloudMachine.deviceIp) {
          deviceIp = cloudMachine.deviceIp
          activeDevice.value = {
            ip: deviceIp,
            version: 'v3' // 默认版本
          }
          console.log('Set activeDevice.value from cloudMachine:', activeDevice.value)
        } else {
          // 作为最后的 fallback，使用一个默认的设备IP（仅用于测试）
          console.warn('No device IP found, using default for testing')
          activeDevice.value = {
            ip: '127.0.0.1',
            version: 'v3'
          }
        }
      }
    }
    
    console.log('Final activeDevice.value:', activeDevice.value)
    
    if (activeDevice.value && activeDevice.value.ip) {
      // 打开共享文件选择对话框
      loadSharedFiles()
    } else {
      console.error('Device info incomplete:', { activeDevice: activeDevice.value, container: container })
      ElMessage.error('设备信息不完整，无法上传文件')
    }
  } else {
    console.error('No container found in contextMenuContainer.value')
    ElMessage.error('未找到容器信息，无法上传文件')
  }
  closeContextMenu()
}

// 排序文件树节点
const sortFileTreeNode = (node) => {
  if (!node || !node.children || node.children.length === 0) {
    return
  }
  
  node.children.sort((a, b) => {
    // 目录始终在文件之前
    if (a.isDir && !b.isDir) return -1
    if (!a.isDir && b.isDir) return 1
    
    // 根据排序类型和顺序排序
    let comparison = 0
    if (fileSortType.value === 'name') {
      comparison = a.name.localeCompare(b.name, 'zh-CN')
    } else if (fileSortType.value === 'time') {
      comparison = (a.modTime || 0) - (b.modTime || 0)
    }
    
    return fileSortOrder.value === 'asc' ? comparison : -comparison
  })
  
  // 递归排序子目录
  node.children.forEach(child => {
    if (child.isDir) {
      sortFileTreeNode(child)
    }
  })
}

// 切换文件排序
const changeFileSort = (type) => {
  if (fileSortType.value === type) {
    fileSortOrder.value = fileSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    fileSortType.value = type
    fileSortOrder.value = 'desc'
  }
  
  // 重新排序文件树
  if (sharedFileTree.value) {
    sortFileTreeNode(sharedFileTree.value)
  }
}

// 加载共享目录文件
const loadSharedFiles = async () => {
  try {
    filesLoading.value = true
    loadSingleUploadSharedDirPath()
    const result = await ListSharedDirFiles()
    if (result.success) {
      sharedFiles.value = []
      sharedFileTree.value = result.tree
      sharedRootPath.value = result.rootPath
      selectedFiles.value = []
      
      // 排序文件树
      if (sharedFileTree.value) {
        sortFileTreeNode(sharedFileTree.value)
      }
      
      // 默认展开第一个shared目录
      if (sharedFileTree.value && sharedFileTree.value.isDir) {
        sharedFileTree.value.expanded = true
      }
      
      sharedFilesDialogVisible.value = true
    } else {
      ElMessage.error(`加载共享文件失败: ${result.message}`)
    }
  } catch (error) {
    ElMessage.error(`加载共享文件失败: ${error.message}`)
  } finally {
    filesLoading.value = false
  }
}


// 刷新共享目录
const handleUploadRefresh = async () => {
  try {
    filesLoading.value = true
    const result = await ListSharedDirFiles()
    if (result.success) {
      sharedFiles.value = []
      sharedFileTree.value = result.tree
      sharedRootPath.value = result.rootPath
      selectedFiles.value = []
      
      // 排序文件树
      if (sharedFileTree.value) {
        sortFileTreeNode(sharedFileTree.value)
      }
      
      // 默认展开第一个shared目录
      if (sharedFileTree.value && sharedFileTree.value.isDir) {
        sharedFileTree.value.expanded = true
      }
      
      sharedFilesDialogVisible.value = true
      ElMessage.success('刷新成功')
    } else {
      ElMessage.error(`加载共享文件失败: ${result.message}`)
    }
  } catch (error) {
    ElMessage.error(`加载共享文件失败: ${error.message}`)
  } finally {
    filesLoading.value = false
  }
}

// 递归处理目录树，提取所有文件
const processFileTree = (node, callback) => {
  if (node.isDir) {
    if (node.children && node.children.length > 0) {
      node.children.forEach(child => {
        processFileTree(child, callback)
      })
    }
  } else {
    callback(node)
  }
}

// 计算目录的选择状态
const getDirectorySelectionState = (node) => {
  if (!node.isDir || !node.children || node.children.length === 0) {
    return selectedFiles.value.includes(node.path) ? 'checked' : 'unchecked'
  }
  
  let checkedCount = 0
  let totalCount = 0
  
  // 递归计算所有子文件的选择状态
  const calculateSelection = (currentNode) => {
    if (currentNode.isDir) {
      if (currentNode.children && currentNode.children.length > 0) {
        currentNode.children.forEach(child => calculateSelection(child))
      }
    } else {
      totalCount++
      if (selectedFiles.value.includes(currentNode.path)) {
        checkedCount++
      }
    }
  }
  
  calculateSelection(node)
  
  if (totalCount === 0) {
    return 'unchecked'
  } else if (checkedCount === totalCount) {
    return 'checked'
  } else if (checkedCount > 0) {
    return 'indeterminate'
  } else {
    return 'unchecked'
  }
}

// 处理文件选择变化
const handleNodeSelectionChange = (node, isDirectoryClick = false) => {
  if (node.isDir && isDirectoryClick) {
    // 点击目录的选择框，全选或取消全选
    const files = []
    
    // 递归收集所有文件
    const collectFiles = (currentNode) => {
      if (currentNode.isDir) {
        if (currentNode.children && currentNode.children.length > 0) {
          currentNode.children.forEach(child => collectFiles(child))
        }
      } else {
        files.push(currentNode.path)
      }
    }
    
    collectFiles(node)
    
    // 检查目录是否已全选
    const allSelected = files.every(filePath => selectedFiles.value.includes(filePath))
    
    if (allSelected) {
      // 取消选择所有文件
      selectedFiles.value = selectedFiles.value.filter(filePath => !files.includes(filePath))
    } else {
      // 选择所有文件
      files.forEach(filePath => {
        if (!selectedFiles.value.includes(filePath)) {
          selectedFiles.value.push(filePath)
        }
      })
    }
  }
  // 文件的选择由v-model自动处理
}

// 处理目录节点展开/折叠
const toggleNodeExpanded = (node) => {
  if (node.isDir) {
    if (node.expanded === undefined) {
      node.expanded = true
    } else {
      node.expanded = !node.expanded
    }
  }
}

// 收集目录下所有文件路径
const collectSharedFilePaths = (node) => {
  const filePaths = []
  const collect = (n) => {
    if (n.children) {
      n.children.forEach(child => {
        if (!child.isDir) {
          filePaths.push(child.path)
        } else {
          collect(child)
        }
      })
    }
  }
  collect(node)
  return filePaths
}

// 判断共享目录是否全选
const isSharedDirectoryFullySelected = (node) => {
  if (!node || !node.isDir) return false
  if (!node.children || node.children.length === 0) return false
  
  const filePaths = collectSharedFilePaths(node)
  if (filePaths.length === 0) return false
  
  return filePaths.every(filePath => selectedFiles.value.includes(filePath))
}

// 判断共享目录是否部分选中
const isSharedDirectoryPartiallySelected = (node) => {
  if (!node || !node.isDir) return false
  if (!node.children || node.children.length === 0) return false
  
  const filePaths = collectSharedFilePaths(node)
  if (filePaths.length === 0) return false
  
  const allSelected = filePaths.every(filePath => selectedFiles.value.includes(filePath))
  if (allSelected) return false
  
  const someSelected = filePaths.some(filePath => selectedFiles.value.includes(filePath))
  return someSelected
}

// 处理共享目录选择变化
const handleSharedNodeSelectionChange = (node) => {
  if (!node || !node.isDir || !node.children) return
  
  const filePaths = collectSharedFilePaths(node)
  const currentState = isSharedDirectoryFullySelected(node)
  const shouldCheck = !currentState
  
  if (shouldCheck) {
    selectedFiles.value = [...new Set([...selectedFiles.value, ...filePaths])]
  } else {
    selectedFiles.value = selectedFiles.value.filter(filePath => !filePaths.includes(filePath))
  }
}

// 处理共享文件勾选变化
const handleSharedFileCheckChange = (path, checked) => {
  if (checked) {
    if (!selectedFiles.value.includes(path)) {
      selectedFiles.value = [...selectedFiles.value, path]
    }
  } else {
    selectedFiles.value = selectedFiles.value.filter(p => p !== path)
  }
}



// 处理上传文件到云机
const handleUploadToCloudMachine = async () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请选择要上传的文件')
    return
  }
  
  if (!activeDevice.value || !activeDevice.value.ip || !contextMenuContainerId.value) {
    ElMessage.error('设备信息不完整，无法上传文件')
    return
  }
  
  try {
    uploadLoading.value = true
    const savedPassword = getDevicePassword(activeDevice.value.ip)
    
    // 逐个上传文件
    let successCount = 0
    for (const filePath of selectedFiles.value) {
      // 检查是否是APK文件
      if (filePath.toLowerCase().endsWith('.apk')) {
        // 安装APK
        const result = await InstallAPK(
          activeDevice.value.ip,
          activeDevice.value.version || 'v3',
          contextMenuContainerId.value,
          filePath,
          savedPassword || ''
        )
        
        if (result.success) {
          successCount++
          ElMessage.success(`APK安装成功: ${filePath.split('\\').pop().split('/').pop()}`)
        } else {
          ElMessage.error(`APK安装失败: ${result.message}`)
        }
      } else {
        // 普通文件上传
        const result = await UploadFileToCloudMachine(
          activeDevice.value.ip,
          activeDevice.value.version || 'v3',
          contextMenuContainerId.value,
          filePath,
          savedPassword || ''
        )
        
        if (result.success) {
          successCount++
        } else {
          ElMessage.error(`上传文件失败: ${result.message}`)
        }
      }
    }
    
    if (successCount > 0) {
      ElMessage.success(`成功处理 ${successCount} 个文件`)
      sharedFilesDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error(`处理文件失败: ${error.message}`)
  } finally {
    uploadLoading.value = false
  }
}

// 打开共享目录
const openSharedDirectory = async () => {
  try {
    const result = await OpenSharedDirectory()
    if (result.success) {
      ElMessage.success('已打开共享目录')
    } else {
      ElMessage.error(`打开共享目录失败: ${result.message}`)
    }
  } catch (error) {
    ElMessage.error(`打开共享目录失败: ${error.message}`)
  }
}

// 处理批量上传
const handleBatchUpload = async (uploadData) => {
  const { files, machines, cloudManageMode, selectedCloudDevice, apkOptions } = uploadData
  
  if (!files || files.length === 0) {
    ElMessage.warning('没有选中的文件')
    return
  }
  
  if (!machines || machines.length === 0) {
    ElMessage.warning('没有选中的云机')
    return
  }
  
  // 按设备分组，每台设备单独创建上传任务
  const devicesMap = new Map()
  
  if (cloudManageMode === 'slot' && selectedCloudDevice) {
    // 坑位模式：所有机器在同一设备上
    const deviceIP = selectedCloudDevice.ip
    const deviceVersion = selectedCloudDevice.version || 'v3'
    devicesMap.set(deviceIP, {
      deviceIP,
      deviceVersion,
      machines: machines.map(m => {
        const containerId = m.containerId ?? m.containerID ?? m.id ?? m.name ?? m.indexNum ?? ''
        return {
          ...m,
          containerID: containerId === '' ? '' : String(containerId)
        }
      })
    })
  } else if (cloudManageMode === 'batch') {
    // 批量模式：按设备IP分组，不同机器可能在不同设备上
    for (const machine of machines) {
      const deviceIP = machine.deviceIp
      const deviceVersion = machine.deviceVersion || 'v3'
      
      if (!devicesMap.has(deviceIP)) {
        devicesMap.set(deviceIP, {
          deviceIP,
          deviceVersion,
          machines: []
        })
      }
      
      const deviceInfo = devicesMap.get(deviceIP)
      const containerId = machine.containerId ?? machine.containerID ?? machine.id ?? machine.name ?? machine.indexNum ?? ''
      deviceInfo.machines.push({
        ...machine,
        containerID: containerId === '' ? '' : String(containerId)
      })
    }
  }
  
  if (devicesMap.size === 0) {
    ElMessage.error('设备信息不完整，无法上传文件')
    return
  }
  
  // 为每个设备创建上传任务
  const uploadTasks = []
  for (const [deviceIP, deviceInfo] of devicesMap) {
    for (const filePath of files) {
      const isAPK = filePath.toLowerCase().endsWith('.apk')
      uploadTasks.push({
        filePath,
        isAPK,
        deviceIP: deviceInfo.deviceIP,
        deviceVersion: deviceInfo.deviceVersion,
        machines: deviceInfo.machines,
        apkOptions: apkOptions || {} // 传递 APK 安装选项
      })
    }
  }
  
  // 添加到任务队列
  const taskId = addTaskToQueue('uploadFile', uploadTasks, {
    fileCount: files.length,
    machineCount: machines.length,
    deviceCount: devicesMap.size,
    hasAPK: files.some(f => f.toLowerCase().endsWith('.apk')),
    apkOptions: apkOptions || {} // 保存到任务元数据
  })
  
  // 执行任务
  executeTask(taskId)
  
  // 关闭对话框
  batchUploadDialogVisible.value = false
  
  ElMessage.success(`批量上传任务已添加到队列，共 ${files.length} 个文件上传到 ${machines.length} 个云机（${devicesMap.size} 台设备）`)
}

// 处理文件选择
const handleFileSelect = async (event) => {
  const files = event.target.files
  if (files.length === 0) return

  try {
    const file = files[0]
    
    if (activeDevice.value && activeDevice.value.ip && contextMenuContainerId.value) {
      try {
        const savedPassword = getDevicePassword(activeDevice.value.ip)
        // 尝试获取文件路径，兼容不同浏览器环境
        let filePath = file.name
        
        // 注意：在Wails环境中，我们需要使用特殊方法获取文件路径
        // 但由于直接传递File对象会导致postMessage错误，我们将使用后端方法来处理
        // 这里我们将文件名传递给后端，后端会处理实际的文件读取
        
        if (!filePath) {
          ElMessage.error('无法获取文件名')
          return
        }

        const result = await UploadFileToSharedDir(
          activeDevice.value.ip,
          activeDevice.value.version || 'v3',
          contextMenuContainerId.value,
          filePath,
          savedPassword || ''
        )
        
        if (result.success) {
          ElMessage.success(`文件上传成功: ${file.name}`)
        } else {
          ElMessage.error(`文件上传失败: ${result.message}`)
        }
      } catch (error) {
        ElMessage.error(`上传失败: ${error.message}`)
      }
    } else {
      ElMessage.error('设备信息不完整，无法上传文件')
    }
  } catch (error) {
    ElMessage.error(`处理文件失败: ${error.message}`)
  }
  
  // 清空文件输入，允许重复选择同一个文件
  event.target.value = ''
  // 清空存储的容器ID
  contextMenuContainerId.value = ''
}


// 组件销毁时清理资源
onBeforeUnmount(() => {
  // 清除定时器
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
  
  // 清除每半小时刷新定时器
  if (hourlyRefreshInterval) {
    clearInterval(hourlyRefreshInterval)
    hourlyRefreshInterval = null
  }
  
  // 清除版本信息刷新定时器
  if (versionRefreshInterval) {
    clearInterval(versionRefreshInterval)
    versionRefreshInterval = null
  }
  
  // 清除同步授权定时器
  if (syncAuthTimer.value) {
    clearInterval(syncAuthTimer.value)
    syncAuthTimer.value = null
  }
  
  // 停止截图刷新并释放资源
  stopScreenshotRefresh()
  
  // 清除批量添加防抖定时器
  if (batchAddTimeout.value) {
    clearTimeout(batchAddTimeout.value)
    batchAddTimeout.value = null
  }
  
  // 清除分组初始化防抖定时器
  if (initGroupsTimeout.value) {
    clearTimeout(initGroupsTimeout.value)
    initGroupsTimeout.value = null
  }
  
  // 清除版本检查防抖定时器
  if (versionCheckTimer) {
    clearTimeout(versionCheckTimer)
    versionCheckTimer = null
  }
  
  // 使用Wails的事件API移除下载进度事件监听器
  Events.Off('download-progress')
  // 使用Wails的事件API移除上传进度事件监听器
  Events.Off('upload-progress')
  // 注意：EventsOnce 监听的事件不需要手动移除，它们只触发一次
})

// 半小时逐个设备刷新云机列表
const refreshDevicesContainersOneByOne = async () => {
  console.log('开始逐个设备刷新云机列表')
  
  // 如果没有设备，直接返回
  if (devices.value.length === 0) {
    console.log('没有设备需要刷新')
    return
  }
  
  // 计算下一个要刷新的设备索引
  const device = devices.value[nextDeviceIndex]
  nextDeviceIndex = (nextDeviceIndex + 1) % devices.value.length
  
  console.log(`正在刷新设备 ${device.ip} 的云机列表`)
  try {
    // 使用fetchAndroidContainers函数刷新设备云机列表，isUserInitiated=false表示后台加载
    await fetchAndroidContainers(device, false)
    // 更新云机分组数据
    initCloudMachineGroups()
    console.log(`设备 ${device.ip} 云机列表刷新成功`)
  } catch (error) {
    console.error(`设备 ${device.ip} 云机列表刷新失败:`, error)
  }
}

// 批量加载云机列表，每10个设备一批，超时10秒
const batchLoadCloudMachines = async () => {
  console.log('开始批量加载云机列表')
  
  if (devices.value.length === 0) {
    console.log('没有设备需要加载')
    return
  }
  
  // 计算批次
  const BATCH_SIZE = 10;
  const totalDevices = devices.value.length;
  const totalBatches = Math.ceil(totalDevices / BATCH_SIZE);
  
  for (let batchIndex = 0; batchIndex < totalBatches; batchIndex++) {
    // 计算当前批次的设备范围
    const startIndex = batchIndex * BATCH_SIZE;
    const endIndex = Math.min(startIndex + BATCH_SIZE, totalDevices);
    const batchDevices = devices.value.slice(startIndex, endIndex);
    
    console.log(`正在加载第 ${batchIndex + 1}/${totalBatches} 批设备云机列表，共 ${batchDevices.length} 个设备`)
    
    // 并行加载当前批次的设备云机列表
    const loadPromises = batchDevices.map(device => 
      // 使用fetchAndroidContainers函数，isUserInitiated=false表示后台加载
      fetchAndroidContainers(device, false).catch(error => {
        console.error(`设备 ${device.ip} 云机列表加载失败:`, error)
        return null
      })
    );
    
    // 等待当前批次加载完成或超时
    try {
      await Promise.all(loadPromises);
    } catch (error) {
      console.error(`第 ${batchIndex + 1} 批设备云机列表加载失败:`, error)
    }
    
    // 更新云机分组数据
    initCloudMachineGroups();
  }
  
  console.log('所有批次设备云机列表加载完成')
}

// 计算属性
const slotGroups = computed(() => {
  const groups = {}
  for (let i = 1; i <= 12; i++) {
    groups[i] = instances.value.filter(inst => inst.indexNum === i)
  }
  return groups
})

const runningCloudMachines = computed(() => {
  return cloudMachines.value.filter(m => m.status === 'running')
})



const groupedInstances = computed(() => {
  let result = []
  const slotCountMap = new Map()
  
  let maxSlots = 12
  if (selectedCloudDevice.value && selectedCloudDevice.value.name && selectedCloudDevice.value.name.toLowerCase().includes('p1')) {
    maxSlots = 24
  }
  
  for (let slotNum = 1; slotNum <= maxSlots; slotNum++) {
    const slotInstances = instances.value.filter(inst => inst.indexNum === slotNum)
    if (slotInstances.length > 0) {
      slotInstances.forEach((inst, index) => {
        result.push({
          slotNum,
          isFirstInSlot: index === 0,
          instanceCount: slotInstances.length,
          ...inst
        })
      })
      slotCountMap.set(slotNum, slotInstances.length)
    } else {
      result.push({
        slotNum,
        isFirstInSlot: true,
        instanceCount: 1,
        name: '',
        ip: '',
        image: '',
        createTime: '',
        status: 'shutdown',
        modelName: ''
      })
      slotCountMap.set(slotNum, 1)
    }
  }

  return result
  
})

const deviceDetailCloudMachines = ref([])
const deviceDetailCloudMachinesLoading = ref(false)

const fetchDeviceDetailCloudMachines = async () => {
  if (!activeDevice.value || !activeDevice.value.ip) return
  
  try {
    deviceDetailCloudMachinesLoading.value = true
    const data = await getDeviceAllCloudMachines(activeDevice.value.ip, '', true)
    console.log('fetchDeviceDetailCloudMachines', data)
    if (data && data.code == 0) {
      deviceDetailCloudMachines.value = data.data.list || []
    } else if (Array.isArray(data)) {
      deviceDetailCloudMachines.value = data
    } else {
      deviceDetailCloudMachines.value = []
    }
    
    console.log('设备详情云机数据获取成功:', deviceDetailCloudMachines.value)
  } catch (error) {
    console.error('获取设备详情云机数据失败:', error)
    deviceDetailCloudMachines.value = []
  } finally {
    deviceDetailCloudMachinesLoading.value = false
  }
}

const handleDeviceDetailDeleteContainer = async (container) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除云机 "${container.name || container.ID}" 吗？删除后数据将无法恢复。`,
      '删除云机',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'danger'
      }
    )
    
    const deviceIp = container.deviceIp || activeDevice.value?.ip
    const deviceVersion = container.deviceVersion || activeDevice.value?.version || 'v3'
    const containerName = container.name || container.ID
    
    if (!deviceIp || !containerName) {
      ElMessage.error('无法获取设备信息或容器名称')
      return
    }
    
    const port = deviceVersion === 'v3' ? '8000' : '81'
    const savedPassword = getDevicePassword(deviceIp)
    let headers = {}
    
    if (savedPassword) {
      const auth = btoa(`admin:${savedPassword}`)
      headers = { 'Authorization': `Basic ${auth}` }
    }
    
    let response
    if (deviceVersion === 'v3') {
      response = await axios.delete(`http://${getDeviceAddr(deviceIp)}/android/?name=${containerName}`, { headers })
    } else {
      response = await axios.delete(`http://${getDeviceAddr(deviceIp)}/android/?name=${containerName}`, { headers })
    }
    
    if (response.data && response.data.code === 0) {
      ElMessage.success('删除云机成功')
      await fetchDeviceDetailCloudMachines()
    } else {
      const errorMsg = response.data?.message || '删除失败'
      ElMessage.error(`删除云机失败: ${errorMsg}`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除云机失败:', error)
      ElMessage.error('删除云机失败')
    }
  }
}

const deviceDetailSlotFoldStatus = ref({})

const toggleSlotFold = (slotNum) => {
  deviceDetailSlotFoldStatus.value[slotNum] = !deviceDetailSlotFoldStatus.value[slotNum]
}

const deviceDetailGroupedInstances = computed(() => {
  if (!activeDevice.value) return []

  const result = []
  
  let maxSlots = 12
  if (activeDevice.value && activeDevice.value.name && activeDevice.value.name.toLowerCase().includes('p1')) {
    maxSlots = 24
  }
  
  for (let slotNum = 1; slotNum <= maxSlots; slotNum++) {
    const slotInstances = deviceDetailCloudMachines.value.filter(inst => inst.indexNum === slotNum)
    const isExpanded = deviceDetailSlotFoldStatus.value[slotNum] === true
    
    const runningInstances = slotInstances.filter(inst => inst.status === 'running')
    const stoppedInstances = slotInstances.filter(inst => inst.status !== 'running')
    
    if (slotInstances.length > 0) {
      if (runningInstances.length > 0) {
        const allChildren = [...runningInstances.slice(1), ...stoppedInstances]
        const firstInstance = {
          ...runningInstances[0],
          id: `slot-${slotNum}-${runningInstances[0].name || 'first'}`,
          slotNum,
          isFirstInSlot: true,
          instanceCount: slotInstances.length,
          isExpanded: isExpanded,
          showFoldButton: stoppedInstances.length > 0 || runningInstances.length > 1,
          hasChildren: allChildren.length > 0,
          _children: allChildren.map((child, idx) => ({
            ...child,
            id: `slot-${slotNum}-child-${idx}-${child.name || 'child'}`
          }))
        }
        result.push(firstInstance)
      } else {
        const allChildren = stoppedInstances.slice(1)
        const firstInstance = {
          ...stoppedInstances[0],
          id: `slot-${slotNum}-${stoppedInstances[0].name || 'first'}`,
          slotNum,
          isFirstInSlot: true,
          instanceCount: slotInstances.length,
          isExpanded: isExpanded,
          showFoldButton: stoppedInstances.length > 1,
          hasChildren: allChildren.length > 0,
          _children: allChildren.map((child, idx) => ({
            ...child,
            id: `slot-${slotNum}-child-${idx}-${child.name || 'child'}`
          }))
        }
        result.push(firstInstance)
      }
    } else {
      result.push({
        id: `slot-${slotNum}-empty`,
        slotNum,
        isFirstInSlot: true,
        instanceCount: 1,
        isExpanded: false,
        showFoldButton: false,
        hasChildren: false,
        _children: [],
        name: '',
        ip: '',
        image: '',
        createTime: '',
        status: 'shutdown',
        modelName: ''
      })
    }
  }

  return result
  
})


const filteredCloudMachines = computed(() => {
  return cloudMachines.value.filter(m => 
    m.status === 'running' && 
    (selectedCloudMachines.value.length === 0 || selectedCloudMachines.value.includes(m.id))
  )
})

// 计算属性：按名称索引云机，方便模板中快速访问
const cloudMachinesByName = computed(() => {
  const map = new Map();
  cloudMachines.value.forEach(machine => {
    map.set(machine.name, machine);
  });
  // console.log('cloudMachinesByName:', map)
  return map;
})

// 生命周期钩子已经在文件顶部定义，此处移除重复定义


// 添加设备（处理来自AddDeviceDialog的事件）
const handleAddDevice = (device) => {
  if (isBatchAddingDevices.value) {
    batchPendingDevices.value.push(device)
    return
  }
  
  try {
    // 按 id 或 ip 去重：同 ip 的设备已存在则更新其 id/name，不重复添加
    const existingByIp = devices.value.find(d => d.ip === device.ip)
    if (existingByIp) {
      // 用新数据更新已有设备（修正 id/name 等字段）
      Object.assign(existingByIp, device)
      console.log(`设备 ${device.ip} 已存在，更新数据`)
      saveDevicesToLocalStorage()
      return
    }
    const existingById = devices.value.find(d => d.id === device.id)
    if (existingById) {
      console.log(`设备 ${device.ip} 已存在（id重复），跳过`)
      return
    }
    
    if (!device.group) {
      device.group = '默认分组'
    }
    
    devices.value.push(device)
    devicesStatusCache.value.set(device.id, 'offline')  // ✅ 改为 offline，由心跳检测验证
    devicesLastUpdateTime.value.set(device.id, Date.now())
    
    saveDevicesToLocalStorage()
    initCloudMachineGroups()
    autoGetAllDeviceVersions()
    
    // ✅ 手动触发心跳监控列表更新
    if (heartbeatInitialized) {
      console.log('[添加设备] 手动触发心跳监控更新')
      updateHeartbeatDevices()
    }
    
    console.log(`设备 ${device.ip} 添加成功，分组: ${device.group}`)
  } catch (error) {
    console.error('添加设备失败:', error)
  }
}

// 批量添加设备（处理来自AddDeviceDialog的批量事件）
const handleBatchAddDevices = async (devicesToAdd) => {
  if (!devicesToAdd || devicesToAdd.length === 0) return
  
  isBatchAddingDevices.value = true
  batchPendingDevices.value = []
  
  try {
    const devicesToProcess = []
    
    for (const device of devicesToAdd) {
      const existingDevice = devices.value.find(d => d.id === device.id)
      if (!existingDevice) {
        if (!device.group) {
          device.group = '默认分组'
        }
        devicesToProcess.push(device)
      } else {
        console.log(`设备 ${device.ip} 已存在，跳过`)
      }
    }
    
    if (devicesToProcess.length === 0) {
      ElMessage.warning('所选设备均已存在')
      return
    }
    
    for (const device of devicesToProcess) {
      devices.value.push(device)
      devicesStatusCache.value.set(device.id, 'offline')  // ✅ 改为 offline，由心跳检测验证
      devicesLastUpdateTime.value.set(device.id, Date.now())
    }
    
    saveDevicesToLocalStorage()
    initCloudMachineGroups()
    autoGetAllDeviceVersions()
    
    // ✅ 手动触发心跳监控列表更新
    if (heartbeatInitialized) {
      console.log('[批量添加设备] 手动触发心跳监控更新')
      updateHeartbeatDevices()
    }
    
    console.log(`批量添加设备成功: ${devicesToProcess.length} 个`)
  } catch (error) {
    console.error('批量添加设备失败:', error)
    ElMessage.error('批量添加设备失败: ' + error.message)
  } finally {
    isBatchAddingDevices.value = false
    batchPendingDevices.value = []
  }
}

// 发现设备并加载数据


// 后台10个一批加载云机列表（只处理在线设备）
const loadContainersInBatches = async () => {
  if (devices.value.length === 0) return

  // 只对在线设备发请求，避免对离线设备产生无效 HTTP 超时
  const onlineDevices = devices.value.filter(device =>
    devicesStatusCache.value.get(device.id) === 'online'
  )

  if (onlineDevices.length === 0) {
    console.log('没有在线设备，跳过云机列表加载')
    return
  }

  console.log(`开始后台10个一批加载云机列表，在线设备 ${onlineDevices.length} 台`)

  // 将在线设备分成每10个一组
  const batchSize = 10
  for (let i = 0; i < onlineDevices.length; i += batchSize) {
    const batch = onlineDevices.slice(i, i + batchSize)
    console.log(`加载第 ${Math.floor(i / batchSize) + 1} 组设备云机列表，共 ${batch.length} 个设备`)

    // 并行加载该组设备的云机列表，isUserInitiated=false表示后台加载
    const promises = batch.map(device => {
      return fetchAndroidContainers(device, false)
        .then(() => {
          console.log(`设备 ${device.ip} 云机列表加载成功`)
        })
    })

    // 使用Promise.allSettled，即使有设备加载失败，其他设备仍能继续加载
    await Promise.allSettled(promises)

    // 更新云机分组数据
    initCloudMachineGroups()

    // 短暂延迟，避免服务器压力过大
    await new Promise(resolve => setTimeout(resolve, 1000))
  }

  console.log('所有在线设备云机列表加载完成')
}

// 刷新数据
const refreshData = async () => {
  if (activeDevice.value) {
    await fetchAndroidContainers(activeDevice.value, true) // isUserInitiated=true表示用户主动刷新
  }
}


// ADB端口禁用列表（这些端口已被其他服务占用）
const FORBIDDEN_ADB_PORTS = new Set([9082, 9083, 10000, 10001, 10006, 10007, 10008])

// 校验 ADB 端口
const validateAdbPort = (value) => {
  if (value === null || value === undefined || value === '') {
    createForm.value.adbPort = 5555
    return
  }
  const port = Number(value)
  if (isNaN(port) || port < 0) {
    ElMessage.warning('ADB端口不能小于0，已重置为默认值5555')
    createForm.value.adbPort = 5555
    return
  }
  if (port > 65535) {
    ElMessage.warning('ADB端口不能超过65535，已重置为65535')
    createForm.value.adbPort = 65535
    return
  }
  if (port !== 0 && FORBIDDEN_ADB_PORTS.has(port)) {
    ElMessage.warning(`端口 ${port} 已被系统服务占用（9082/9083/10000/10001/10006/10007/10008），请使用其他端口`)
    createForm.value.adbPort = 5555
    return
  }
}

// 从容器实例中动态获取 ADB 端口
// 优先读取实例的 adbPort 字段，否则从 portBindings 中排除已知端口后推断
const getInstanceAdbPort = (instance) => {
  if (!instance) return 5555
  // 优先使用实例中保存的 adbPort 字段
  if (instance.adbPort && instance.adbPort !== 0) {
    return Number(instance.adbPort)
  }
  // 从 portBindings 中推断：排除已知端口后，剩余的可能是 ADB 端口
  const knownPorts = new Set([8000, 9082, 9083, 10000, 10001, 10006, 10007, 10008])
  const bindings = instance.portBindings || instance.PortBindings
  if (bindings) {
    for (const [key] of Object.entries(bindings)) {
      const portNum = parseInt(key.split('/')[0])
      if (!isNaN(portNum) && !knownPorts.has(portNum)) {
        // 找到一个非已知端口，可能就是 ADB 端口
        return portNum
      }
    }
  }
  // 默认回退到 5555
  return 5555
}

// 从容器信息中提取端口映射
const extractPort = (container, portNumber) => {
  // 优先从容器的缓存中获取端口映射，避免重复计算
  // 注意：OpenCecs 公网设备不缓存，因为端口映射表可能在首次调用后才填充
  const isPublicDevice = container.deviceIp && container.deviceIp.includes(':')
  const cacheKey = `_cachedPort${portNumber}`
  if (!isPublicDevice && container[cacheKey]) {
    return container[cacheKey]
  }
  
  let mappedPort = null
  const portKeyTcp = `${portNumber}/tcp`
  const portKeyUdp = `${portNumber}/udp`
  
  // 支持V3 API格式: container.portBindings['port/tcp'][0].HostPort
  if (container.portBindings) {
    const portBinding = container.portBindings[portKeyTcp] || container.portBindings[portKeyUdp]
    if (portBinding && portBinding.length > 0) {
      mappedPort = portBinding[0].HostPort
    }
  }
  
  // 支持Docker API原生格式: container.Ports数组
  if (!mappedPort && container.Ports && Array.isArray(container.Ports)) {
    // 查找PrivatePort为指定端口的端口映射（先TCP后UDP）
    const port = container.Ports.find(p => p.PrivatePort === portNumber && p.Type === 'tcp')
      || container.Ports.find(p => p.PrivatePort === portNumber && p.Type === 'udp')
    if (port && port.PublicPort) {
      mappedPort = port.PublicPort
    }
  }
  
  // 兼容旧版Docker API格式（通过端口名查找）
  if (!mappedPort && container.NetworkSettings && container.NetworkSettings.Ports) {
    const portBinding = container.NetworkSettings.Ports[portKeyTcp] || container.NetworkSettings.Ports[portKeyUdp]
    if (portBinding && portBinding.length > 0) {
      mappedPort = portBinding[0].HostPort
    }
  }
  
  // 兼容V3 API的另一种格式
  if (!mappedPort && container.PortBindings) {
    const portBinding = container.PortBindings[portKeyTcp] || container.PortBindings[portKeyUdp]
    if (portBinding && portBinding.length > 0) {
      mappedPort = portBinding[0].HostPort
    }
  }

  // OpenCecs 公网设备：将 HostPort（局域网端口）转换为公网端口映射
  // 当 deviceIp 包含 ":"（格式为 publicIp:publicPort）时，说明是公网设备
  if (mappedPort && container.deviceIp && container.deviceIp.includes(':')) {
    // 先精确匹配 deviceIp，如果找不到则按公网 IP 前缀模糊匹配
    // （因为每次创建 8000 端口映射可能得到不同的公网端口，旧容器的 deviceIp 可能过时）
    let portMap = window.openCecsPortMap?.get(container.deviceIp)
    if (!portMap && window.openCecsPortMap) {
      const ipPrefix = container.deviceIp.split(':')[0] + ':'
      for (const [key, map] of window.openCecsPortMap) {
        if (key.startsWith(ipPrefix)) {
          portMap = map
          break
        }
      }
    }
    if (portMap) {
      const publicPort = portMap.get(Number(mappedPort))
      if (publicPort) {
        mappedPort = publicPort
      }
    }
  }
  
  // 缓存端口映射结果，避免重复计算
  container[cacheKey] = mappedPort
  return mappedPort
}

// 从容器信息中提取9082端口的映射端口
const extractPort9082 = (container) => {
  if (container && (container.networkName === 'myt' || container.networkMode === 'myt' || container.network === 'myt')) {
    return 9082
  }
  return extractPort(container, 9082)
};

// 全局辅助函数：为公网设备解析端口映射
// 返回 { ip, port } 对象，公网设备返回 publicIp + publicPort，局域网设备返回原始值
window.resolveOpenCecsAddress = (deviceIp, hostPort) => {
  if (!deviceIp || !deviceIp.includes(':')) {
    // 局域网设备：直接返回
    return { ip: deviceIp, port: hostPort }
  }
  const publicIp = deviceIp.split(':')[0]
  const portMap = window.openCecsPortMap?.get(deviceIp)
  const publicPort = portMap?.get(Number(hostPort))
  if (publicPort) {
    return { ip: publicIp, port: publicPort }
  }
  // 未找到映射，返回原始值（可能URL会有问题但不影响局域网设备）
  return { ip: deviceIp, port: hostPort }
};

// 获取SDK端口
const getSDKPort = (version, sys_ver) => {
  // v3设备且系统版本为5时使用8000端口，否则使用81端口
  if (version === 'v3' && sys_ver === '5') {
    return '8000'
  }
  return '81'
}

// 获取端口映射信息
const getPortMappings = (instance, device) => {
  console.log('getPortMappings', device)
  // 保留原始 device 用于查 openCecsPortMap
  const originalDevice = device
  const isPublicDevice = device && device.includes(':')
  // OpenCecs 公网设备：device 可能是 publicIp:publicPort，提取纯 IP
  if (isPublicDevice) device = device.split(':')[0]

  // 查找 OpenCecs 端口映射表
  let portMap = null
  if (isPublicDevice && window.openCecsPortMap) {
    portMap = window.openCecsPortMap.get(originalDevice)
    if (!portMap) {
      const ipPrefix = device + ':'
      for (const [key, map] of window.openCecsPortMap) {
        if (key.startsWith(ipPrefix)) { portMap = map; break }
      }
    }
  }

  // Docker 8000 端口：公网设备用映射端口，局域网设备用 8000
  const dockerPort = portMap ? (portMap.get(8000) || 8000) : 8000
  const dockerUrl = isPublicDevice ? `http://${device}:${dockerPort}/docker` : `http://${getDeviceAddr(device)}/docker`

  // 如果没有实例（空坑位），则显示原始端口
  if (!instance) {
    return {
      androidApi: {
        originalPort: 9082,
        mappedPort: portMap ? (portMap.get(9082) || 9082) : 9082,
        description: '安卓设备管理API',
        url: `${device}:${portMap ? (portMap.get(9082) || 9082) : 9082}`,
        isMapped: !!(portMap && portMap.get(9082))
      },
      controlApi: {
        originalPort: 9083,
        mappedPort: portMap ? (portMap.get(9083) || 9083) : 9083,
        description: 'RPA自动化API',
        url: `${device}:${portMap ? (portMap.get(9083) || 9083) : 9083}`,
        isMapped: !!(portMap && portMap.get(9083))
      },
      adb: {
        originalPort: 5555,
        mappedPort: portMap ? (portMap.get(5555) || 5555) : 5555,
        description: 'AndroidADB(默认)',
        url: `${device}:${portMap ? (portMap.get(5555) || 5555) : 5555}`,
        isMapped: !!(portMap && portMap.get(5555))
      },
      dockerApi: {
        originalPort: 8000,
        mappedPort: dockerPort,
        description: 'Docker管理接口',
        url: dockerUrl,
        isMapped: dockerPort !== 8000
      }
    }
  }
  
  // 从容器实例中提取真实端口映射
  // myt 网络模式：容器有独立IP，直接用原始端口，不存在端口映射
  const isMytNetwork = instance && (instance.networkName === 'myt' || instance.networkMode === 'myt')
  const androidApiPort = isMytNetwork ? 9082 : (extractPort(instance, 9082) || 9082)
  const controlApiPort = isMytNetwork ? 9083 : (extractPort(instance, 9083) || 9083)
  // 从容器实例中动态获取 ADB 端口（优先从 adbPort 字段读取，否则从 portBindings 中推断）
  const instanceAdbPort = getInstanceAdbPort(instance)
  const adbPort = isMytNetwork ? instanceAdbPort : (extractPort(instance, instanceAdbPort) || instanceAdbPort)
  
  return {
    androidApi: {
      originalPort: 9082,
      mappedPort: androidApiPort,
      description: '安卓设备管理API',
      url: `${device}:${androidApiPort}`,
      isMapped: androidApiPort !== 9082
    },
    controlApi: {
      originalPort: 9083,
      mappedPort: controlApiPort,
      description: 'RPA自动化API',
      url: `${device}:${controlApiPort}`,
      isMapped: controlApiPort !== 9083
    },
    adb: {
      originalPort: instanceAdbPort,
      mappedPort: adbPort,
      description: 'AndroidADB',
      url: `${device}:${adbPort}`,
      isMapped: adbPort !== instanceAdbPort
    },
    dockerApi: {
      originalPort: 8000,
      mappedPort: dockerPort,
      description: 'Docker管理接口',
      url: dockerUrl,
      isMapped: dockerPort !== 8000
    }
  }
}

// 复制到剪贴板
const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(err => {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  })
}

// 显示API详情
const showApiDetails = () => {
  if (!contextMenuContainer.value) return
  

  const container = getCurrentContextMenuContainer()
  // 获取设备信息
  let device = null
  if (cloudManageMode.value === 'slot') {
    // 坑位模式：从activeDevice获取设备信息
    device = activeDevice.value
    if (!device) {
      ElMessage.error('未选择设备')
      return
    }
  } else {
    // 批量模式：从contextMenuContainer获取设备信息
    if (contextMenuContainer.value && contextMenuContainer.value.deviceIp) {
      // 查找对应的设备
      device = devices.value.find(d => d.ip === contextMenuContainer.value.deviceIp)
      if (!device) {
        // 如果找不到设备，创建一个临时设备对象
        device = {
          ip: contextMenuContainer.value.deviceIp,
          version: contextMenuContainer.value.deviceVersion || 'v3',
          name: 'unknown'
        }
      }
    } else {
      ElMessage.error('未选择设备')
      return
    }
  }

  console.log('device:', container)
  
  // 构建API详情数据
  apiDetailsData.value = {
    slotNum: contextMenuSlot.value,
    instanceName: contextMenuContainer.value ? (contextMenuContainer.value.name || '未命名实例') : '空坑位',
    deviceIp: container.networkName == 'myt' ? container.ip : device.ip,
    deviceVersion: device.version || '未知',
    portMappings: getPortMappings(contextMenuContainer.value, container.networkName == 'myt' ? container.ip : device.ip),
    hasInstance: !!contextMenuContainer.value
  }
  
  // 显示API详情对话框
  apiDetailsVisible.value = true
  
  // 隐藏右键菜单
  contextMenuVisible.value = false
}

// 获取云机截图URL
const getCloudMachineScreenshotUrl = (device, container) => {
  // 优先从容器的缓存中获取截图URL，避免重复计算
  // 注意：OpenCecs 公网设备不缓存，端口映射表可能后续才填充
  const isPublicDev = container.deviceIp && container.deviceIp.includes(':')
  if (!isPublicDev && container._cachedScreenshotUrl) {
    return container._cachedScreenshotUrl;
  }
  
  let mappedPort = extractPort9082(container);
  
  // 如果没有找到映射端口，尝试使用默认的10000+坑位号作为端口
  if (!mappedPort && container.indexNum) {
    mappedPort = 10000 + container.indexNum;
  }
  
  if (!mappedPort) {
    return null;
  }
  
  // 使用容器所属设备IP，而不是当前选中设备
  const deviceIp = container.deviceIp || (device ? device.ip : null);
  if (!deviceIp) {
    return null;
  }

  let screenshotUrl;
  if(container.networkName == 'myt') {
    screenshotUrl = `http://${container.ip}:9082/task=snap&level=1?v=202603152350`;
  } else if (deviceIp.includes(':')) {
    // OpenCecs 公网设备：extractPort 已自动返回公网端口，只需提取纯 IP
    const publicIp = deviceIp.split(':')[0]
    screenshotUrl = `http://${publicIp}:${mappedPort}/task=snap&level=1?v=202603152350`;
  } else {
    // 局域网设备：直接使用 deviceIp:HostPort
    screenshotUrl = `http://${deviceIp}:${mappedPort}/task=snap&level=1?v=202603152350`;
  }
  
  // 缓存截图URL，避免重复计算
  container._cachedScreenshotUrl = screenshotUrl;
  return screenshotUrl;
};

// 辅助函数：合并云机状态，保留截图数据
const mergeCloudMachineState = (oldMachines, newMachine) => {
  if (!oldMachines || oldMachines.length === 0) return newMachine;
  
  const oldMachine = oldMachines.find(m => m.id === newMachine.id);
  if (oldMachine) {
    newMachine.screenshotData = oldMachine.screenshotData;
    newMachine.screenshotError = oldMachine.screenshotError;
    newMachine.hasLoadedOnce = oldMachine.hasLoadedOnce;
  }
  return newMachine;
};

// 清空指定容器的截图缓存，用于重启/重置操作
const clearContainerScreenshotCache = (device, container) => {
  if (!device || !container) return;
  
  const cacheKey = `${device.ip}_${container.name}`
  screenshotDataCache.value.delete(cacheKey)
  // 替换新 Map 对象触发响应式更新
  screenshotDataCache.value = new Map(screenshotDataCache.value)
  // 通知后端清空，下次后端轮询会重新抓图
  clearScreenshotCache(device.ip, container.name || '').catch(() => {})

  // console.log(`已清空容器 ${container.name} 的截图缓存并触发重新加载`);
};



// 更新云机列表
const updateCloudMachines = () => {
  if (cloudManageMode.value === 'batch') {
    // 批量模式：从deviceCloudMachinesCache获取所有选中设备的云机数据
    const selectedDeviceIps = new Set(selectedCloudMachines.value.map(machine => machine.deviceIp));
    let allCloudMachines = [];
    
    // 遍历所有选中的设备IP
    selectedDeviceIps.forEach(deviceIp => {
      // 从缓存中获取该设备的所有云机数据
      const deviceCloudMachines = deviceCloudMachinesCache.value.get(deviceIp) || [];
      // 筛选出被选中的云机
      const selectedCloudMachinesForDevice = deviceCloudMachines.filter(machine => {
        return selectedCloudMachines.value.some(selectedMachine => selectedMachine.id === machine.id);
      });
      // 将选中的云机添加到总列表中
    allCloudMachines = [...allCloudMachines, ...selectedCloudMachinesForDevice];
  });
  
  // 对云机按照从ID中提取的坑位号进行排序
  allCloudMachines.sort((a, b) => {
    // 从ID中提取坑位号，例如从"8569541a74175bfe052739c4321ea31b_2_T0002"中提取"2"
    const aParts = a.id.split('_');
    const bParts = b.id.split('_');
    
    // 获取坑位号（第二个元素）并转换为数字
    const aSlot = aParts.length >= 2 ? parseInt(aParts[1]) : 0;
    const bSlot = bParts.length >= 2 ? parseInt(bParts[1]) : 0;
    
    return aSlot - bSlot;
  });
  
  cloudMachines.value = allCloudMachines;
  } else {
    // 坑位模式：从instances.value获取云机数据
    // 先保存当前的云机列表，用于保留截图状态
    const currentCloudMachines = [...cloudMachines.value];
    
    cloudMachines.value = instances.value.map(inst => {
      const screenshotUrl = getCloudMachineScreenshotUrl(activeDevice.value, inst);
      const newMachine = {
        id: inst.name,
        name: inst.name,
        status: inst.status,
        screenshot: screenshotUrl,
        screenshotData: null, // 存储截图URL
        screenshotError: false, // 存储截图加载状态
        hasLoadedOnce: false, // 标记是否已成功加载过至少一次
        ip: inst.ip,
        modelName: inst.modelName,
        deviceIp: inst.deviceIp, // 添加设备IP属性，确保截图URL生成时能获取到设备IP
        indexNum: inst.indexNum, // 添加坑位编号属性，确保截图URL生成时能获取到坑位编号
      };
      
      // 合并状态
      return mergeCloudMachineState(currentCloudMachines, newMachine);
    });
  }
  
  // 云机列表更新后，重新初始化云机分组数据
  initCloudMachineGroups();
}

// ========== 安卓容器截图缓存（后端轮询驱动）==========
// 后端每 800ms 抓一次截图存缓存，前端 300ms 拉版本号，有更新才拉 base64
// 彻底消除前端对每个坑位独立发 IPC 请求的性能开销

// 截图数据缓存 Map<"ip_containerName", base64DataURL>
// 每次更新都替换整个 Map 对象，确保 Vue 响应式能检测到变化
const screenshotDataCache = ref(new Map())
// 每台设备的本地版本号快照，用于对比后端是否有新截图
let screenshotLocalVersions = {}
// 截图版本轮询定时器
let screenshotCacheTimer = null
// 防并发标志：避免同一时刻多个轮询 tick 重叠执行
let screenshotFetching = false

// 300ms 轮询：比较版本号，有变化才拉取该设备的截图数据
const fetchScreenshotCacheIfUpdated = async () => {
  if (screenshotFetching) return
  screenshotFetching = true
  try {
    const versions = await getScreenshotVersions()
    // console.log('[截图轮询] versions:', versions, '| mode:', cloudManageMode.value, '| selectedDevice:', selectedCloudDevice.value?.ip)

    // Wails 会把 Go 空 map 序列化为 null，视为无数据但不阻断逻辑
    if (!versions || Object.keys(versions).length === 0) {
      // console.log('[截图轮询] versions 为空（后端尚无截图数据），跳过')
      return
    }

    // 找出有版本变化的设备
    const targetIps = []
    if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
      const ip = selectedCloudDevice.value.ip
      // console.log(`[截图轮询] 坑位模式 ip=${ip} 后端版本=${versions[ip]} 本地版本=${screenshotLocalVersions[ip]}`)
      if (versions[ip] !== undefined && versions[ip] !== screenshotLocalVersions[ip]) {
        targetIps.push(ip)
      }
    } else if (cloudManageMode.value === 'batch') {
      const ipSet = new Set(selectedCloudMachines.value.map(m => m.deviceIp).filter(Boolean))
      const ips = ipSet.size > 0 ? ipSet : new Set(Object.keys(versions))
      for (const ip of ips) {
        if (versions[ip] !== undefined && versions[ip] !== screenshotLocalVersions[ip]) {
          targetIps.push(ip)
        }
      }
    } else {
      // console.log('[截图轮询] 无匹配模式或无选中设备，跳过')
    }

    // console.log('[截图轮询] targetIps:', targetIps)
    if (targetIps.length === 0) return

    // 并行拉取有变化的设备截图数据
    let hasUpdate = false
    await Promise.all(targetIps.map(async (ip) => {
      const snapshots = await getScreenshots(ip)
      // console.log(`[截图轮询] getScreenshots(${ip}) 返回:`, snapshots ? Object.keys(snapshots).length + ' 条' : 'null/undefined')
      if (!snapshots) return
      let count = 0
      for (const [key, dataURL] of Object.entries(snapshots)) {
        if (dataURL) {
          screenshotDataCache.value.set(key, dataURL)
          hasUpdate = true
          count++
        }
      }
      // console.log(`[截图轮询] 设备 ${ip} 写入缓存 ${count} 张`)
      screenshotLocalVersions[ip] = versions[ip]
    }))

    if (hasUpdate) {
      screenshotDataCache.value = new Map(screenshotDataCache.value)
      // console.log('[截图轮询] screenshotDataCache 已更新，共', screenshotDataCache.value.size, '条')
    }
  } catch (e) {
    console.error('[截图轮询] 异常:', e)
  } finally {
    screenshotFetching = false
  }
}

// 启动截图缓存轮询（切换到云机管理页面时调用）
const startScreenshotRefresh = () => {
  if (screenshotCacheTimer) {
    // console.log('[截图轮询] 定时器已存在，不重复启动')
    return
  }
  console.log('[截图轮询] 启动定时器 150ms')
  screenshotCacheTimer = setInterval(fetchScreenshotCacheIfUpdated, 150)
  fetchScreenshotCacheIfUpdated()
}

// 停止截图缓存轮询（离开云机管理页面时调用）
const stopScreenshotRefresh = () => {
  if (screenshotCacheTimer) {
    // console.log('[截图轮询] 停止定时器')
    clearInterval(screenshotCacheTimer)
    screenshotCacheTimer = null
  }
  screenshotFetching = false
}

// 获取指定容器的截图数据（供 CloudManagement 透传给 ScreenshotImage）
const getContainerScreenshotData = (deviceIp, containerName) => {
  return screenshotDataCache.value.get(`${deviceIp}_${containerName}`) || ''
}

// 监听云机管理模式变化：清空版本快照，强制立即拉取新模式下的截图
watch(cloudManageMode, () => {
  screenshotLocalVersions = {}
  screenshotFetching = false   // 重置并发锁，防止上一轮请求残留导致下次 tick 被跳过
  // 切模式时 selectedCloudMachines 清空会触发 watch(selectedCloudMachines) → stopScreenshotRefresh()
  // 必须在此重新启动定时器，否则后续 300ms 轮询永远不再触发
  screenshotCacheTimer && clearInterval(screenshotCacheTimer)
  screenshotCacheTimer = null
  startScreenshotRefresh()
  // 不清空 screenshotDataCache，保留已有图片避免闪烁
})



// 将ArrayBuffer转换为base64
const arrayBufferToBase64 = (buffer) => {
  let binary = '';
  const bytes = new Uint8Array(buffer);
  const len = bytes.byteLength;
  for (let i = 0; i < len; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return window.btoa(binary);
};

// 获取容器列表


// 树形结构勾选事件处理
// 加载选中云机的截图
const handleTreeCheck = (data, checkedInfo) => {
  // 处理来自CloudManagement.vue的事件数据：{selectedMachines, selectedDevices, treeSelectedKeys}
  let selectedMachines = []
  let selectedDevices = []
  let treeSelectedKeysValue = []
  
  if (data && data.selectedMachines && data.selectedDevices && data.treeSelectedKeys) {
    // 来自CloudManagement.vue的参数格式
    selectedMachines = data.selectedMachines
    selectedDevices = data.selectedDevices
    treeSelectedKeysValue = data.treeSelectedKeys
  } else {
    // 来自el-tree组件的原始参数格式（兼容旧代码）
    return
  }
  
  // 更新选中的节点ID数组
  treeSelectedKeys.value = treeSelectedKeysValue
  
  // 更新选中的云机列表，确保每次都重新赋值
  selectedCloudMachines.value = [...selectedMachines]
}

// 任务队列核心逻辑

// 生成唯一任务ID
const generateTaskId = () => {
  return Date.now().toString(36) + Math.random().toString(36).substring(2, 9)
}

// 获取指定设备IP的进度百分比
const getDeviceProgress = (task, deviceIP) => {
  if (!task.deviceProgress) return 0
  
  const deviceData = task.deviceProgress[deviceIP]
  if (!deviceData) return 0
  
  // 如果有实时进度数据且任务正在进行中，使用实时进度
  if (deviceData.currentProgress !== undefined && task.status === 'running') {
    return deviceData.currentProgress
  }
  
  // 否则使用完成/总数的百分比
  return Math.round((deviceData.completed / deviceData.total) * 100)
}

// 获取指定设备IP的进度状态
const getDeviceProgressStatus = (task, deviceIP) => {
  if (!task.deviceProgress) return ''
  
  const deviceData = task.deviceProgress[deviceIP]
  if (!deviceData) return ''
  
  if (deviceData.completed === deviceData.total) return 'success'
  if (deviceData.failed > 0 && deviceData.completed + deviceData.failed === deviceData.total) return 'exception'
  return ''
}

// 获取指定设备IP的进度状态文字
const getDeviceProgressText = (task, deviceIP) => {
  if (!task.deviceProgress) return { text: '等待中', icon: 'Timer' }
  
  const deviceData = task.deviceProgress[deviceIP]
  if (!deviceData) return { text: '等待中', icon: 'Timer' }
  
  // 针对批量任务，优先根据设备自身的进度判断状态，避免被全局任务状态覆盖
  if ((task.type === 'uploadFile' || task.type === 'uploadImage')) {
    if (deviceData.total > 0) {
      if (deviceData.completed === deviceData.total) {
        return { text: '完成', icon: 'CircleCheck' }
      }
      if (deviceData.failed > 0 && deviceData.completed + deviceData.failed >= deviceData.total) {
        return { text: '失败', icon: 'CircleClose' }
      }
    }
  }
  
  // 任务状态映射
  const statusMap = {
    'pending': { text: '等待中', icon: 'Timer' },
    'running': { text: '上传中', icon: 'Loading' },
    'completed': { text: '完成', icon: 'CircleCheck' },
    'failed': { text: '失败', icon: 'CircleClose' },
    'canceled': { text: '已取消', icon: 'Close' }
  }
  
  // 如果任务已完成或失败，显示最终状态
  if (task.status === 'completed' || task.status === 'failed' || task.status === 'canceled') {
    return statusMap[task.status] || { text: '未知', icon: 'QuestionFilled' }
  }
  
  // 任务进行中，根据设备状态显示
  if (deviceData.completed > 0) {
    return { text: '完成', icon: 'CircleCheck' }
  }
  if (deviceData.failed > 0) {
    return { text: '失败', icon: 'CircleClose' }
  }
  if (deviceData.currentProgress > 0) {
    return { text: '上传中', icon: 'Loading' }
  }
  
  // 检查该设备是否在当前批次中（对于分批上传任务）
  if (task.deviceIps && task.type === 'uploadImage') {
    // 如果进度刚开始，显示等待中
    return { text: '等待中', icon: 'Timer' }
  }
  
  return { text: '等待中', icon: 'Timer' }
}

// 获取任务的目标设备/云机名称显示
const getTaskTargetDisplay = (task) => {
  if (!task) return { short: '', full: [] }
  
  // 对于上传文件任务，从targets中获取云机名称
  if (task.type === 'uploadFile' && task.targets && task.targets.length > 0) {
    const names = []
    task.targets.forEach(target => {
      if (target.machines && target.machines.length > 0) {
        target.machines.forEach(machine => {
          if (machine.name) names.push(formatInstanceName(machine.name))
        })
      }
    })
    // 去重
    const uniqueNames = [...new Set(names)]
    if (uniqueNames.length > 0) {
      return {
        short: uniqueNames.length > 1 ? `${uniqueNames[0]}...` : uniqueNames[0],
        full: uniqueNames
      }
    }
  }
  
  // 对于其他任务，使用deviceIps
  if (task.deviceIps && task.deviceIps.length > 0) {
    return {
      short: task.deviceIps.length > 1 ? `${task.deviceIps[0]}...` : task.deviceIps[0],
      full: task.deviceIps
    }
  }
  
  // 对于创建任务，从targets中获取槽位信息
  if (task.type === 'create' && task.targets && task.targets.length > 0) {
    const slots = task.targets.map(t => `坑位${t.slot}`)
    return {
      short: slots.length > 1 ? `${slots[0]}...` : slots[0],
      full: slots
    }
  }
  
  return { short: '', full: [] }
}

// 添加任务到队列
const addTaskToQueue = (taskType, targets, metadata = {}) => {
  console.log(`[addTaskToQueue] taskType: ${taskType}, targets length: ${targets?.length}, targets:`, targets)
  
  // 从targets中提取设备IP信息
  const deviceIps = new Set()
  if (targets && targets.length > 0) {
    targets.forEach(target => {
      // 支持多种设备IP字段格式（deviceIp小写、deviceIP大写、ip字段）
      if (target.deviceIp) {
        deviceIps.add(target.deviceIp)
      } else if (target.deviceIP) {
        deviceIps.add(target.deviceIP)
      } else if (target.ip) {
        // 设备对象直接有ip字段（如selectedDevicesForUpload中的设备）
        deviceIps.add(target.ip)
      } else if (target.device && target.device.ip) {
        // 嵌套在device对象中
        deviceIps.add(target.device.ip)
      }
    })
  }
  
  console.log(`[addTaskToQueue] deviceIps:`, Array.from(deviceIps))

  // 根据操作类型设置不同的超时时间
  const getTimeout = (type) => {
    switch (type) {
      case 'restart': // 重启需要stop+start，通常需要更长时间
      case 'reset':   // 重置操作通常需要更长时间
        return 30000  // 30秒
      case 'shutdown': // 关机操作
        return 15000  // 15秒
      case 'delete':   // 删除操作
        return 200000  // 20秒
      case 'create':   // 创建操作不使用超时限制
        return 0      // 不限制超时
      case 'switchModel': // 切换机型
        return 20000  // 20秒
      case 'uploadFile':
        return 0      // 不限制超时
      case 'uploadImage': // 批量上传镜像不限制超时
        return 0
      case 'downloadImage': // 下载镜像不限制超时
        return 0
      case 'updateImage': // 批量更新镜像不限制超时
        return 0
      case 'upload': // 上传文件
        return 0  // 不限制超时
      default:
        return 500000   // 默认5秒
    }
  }
  
  // 初始化每个设备的进度统计
  const deviceProgress = {}
  for (const deviceIP of deviceIps) {
    // 根据任务类型设置每个设备的总任务数
    let total = 0
    if (taskType === 'uploadImage') {
      total = 1 // 每个设备上传一个镜像
    } else if (taskType === 'uploadFile') {
      // 统计该设备关联的云机数量
      const deviceTargets = targets.filter(t => 
        (t.deviceIP && t.deviceIP === deviceIP) ||
        (t.ip && t.ip === deviceIP) ||
        (t.device && t.device.ip && t.device.ip === deviceIP)
      )
      total = deviceTargets.reduce((count, target) => {
        return count + (target.machines?.length || 0)
      }, 0)
    }
    
    deviceProgress[deviceIP] = {
      total: total,
      completed: 0,
      failed: 0
    }
  }
  
  // 计算uploadFile任务的总云机数量
  let totalMachines = targets.length
  if (taskType === 'uploadFile') {
    totalMachines = targets.reduce((count, target) => {
      return count + (target.machines?.length || 0)
    }, 0)
  }
  
  const task = {
    id: generateTaskId(),
    type: taskType,
    status: 'pending',
    total: taskType === 'uploadFile' ? totalMachines : targets.length,
    completed: 0,
    failed: 0,
    progress: 0,
    targets: [...targets],
    deviceIps: Array.from(deviceIps), // 操作的设备IP列表
    deviceProgress, // 每个设备的进度统计
    timeout: getTimeout(taskType), // 根据操作类型设置超时
    startTime: null,
    endTime: null,
    failedTargets: [],
    currentStep: taskType === 'create' && targets.some(t => t.isLocalImage) ? 'image' : null, // 创建任务：先推送镜像
    imageProgress: taskType === 'create' && targets.some(t => t.isLocalImage) ? 0 : null,
    ...metadata // 合并额外的任务元数据
  }
  taskQueue.value.unshift(task)
  return task.id
}

// 处理复制云机任务（SSE 流式进度）
const handleStartCopyTask = ({ device, name, indexNum, count, version }) => {
  const ip = device.ip
  const endpoint = version === 'v3'
    ? `http://${getDeviceAddr(ip)}/android/copy?name=${encodeURIComponent(name)}&indexNum=${indexNum}&count=${count}`
    : `http://${getDeviceAddr(ip)}/androidV2/copy?name=${encodeURIComponent(name)}&indexNum=${indexNum}&count=${count}`

  // 创建任务对象并加入队列
  const taskId = generateTaskId()
  const task = {
    id: taskId,
    type: 'copy',
    status: 'running',
    total: count,
    completed: 0,
    failed: 0,
    progress: 0,
    targets: [{ deviceIp: ip }],
    deviceIps: [ip],
    deviceProgress: {},
    timeout: 0,
    startTime: Date.now(),
    endTime: null,
    failedTargets: [],
    currentStep: null,
    imageProgress: null,
    copyLogs: [],           // 流式日志
    sourceName: name,
    copyVersion: version
  }
  taskQueue.value.unshift(task)

  // 通过 fetch 读取 SSE 流
  fetch(endpoint)
    .then(async res => {
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        // 按行解析 SSE data
        const lines = buffer.split('\n')
        buffer = lines.pop() // 保留未完成的一行
        for (const line of lines) {
          const trimmed = line.trim()
          if (!trimmed.startsWith('data:')) continue
          const jsonStr = trimmed.slice(5).trim()
          if (!jsonStr) continue
          try {
            const data = JSON.parse(jsonStr)
            const t = taskQueue.value.find(t => t.id === taskId)
            if (!t) break
            if (data.status === 'done') {
              // 最终汇总
              const successCount = data.success?.length || 0
              const failedMap = data.failed || {}
              const failedCount = Object.keys(failedMap).length
              t.completed = successCount
              t.failed = failedCount
              t.total = data.total || count
              t.progress = 100
              t.status = failedCount > 0 ? (successCount > 0 ? 'completed' : 'failed') : 'completed'
              t.endTime = Date.now()
              // 补充失败详情到 failedTargets
              Object.entries(failedMap).forEach(([machineName, error]) => {
                t.failedTargets.push({ deviceIP: ip, machineName, error })
              })
            } else {
              // 逐条进度
              t.total = data.total || count
              // 更新最新进度（当前处理到第几个）
              if (data.current > t.completed + t.failed) {
                if (data.status === 'success') {
                  t.completed = data.current
                } else if (data.status === 'failed') {
                  t.failed += 1
                }
              }
              const processed = t.completed + t.failed
              t.progress = t.total > 0 ? Math.round((processed / t.total) * 100) : 0
              // 追加日志（去掉相同 name+status 的旧条目，只保留最新状态）
              const existIdx = t.copyLogs.findIndex(l => l.name === data.name && l.status !== 'success' && l.status !== 'failed')
              if (existIdx >= 0) {
                t.copyLogs.splice(existIdx, 1, { current: data.current, total: data.total, name: data.name, status: data.status, message: data.message })
              } else {
                t.copyLogs.push({ current: data.current, total: data.total, name: data.name, status: data.status, message: data.message })
              }
            }
          } catch (e) {
            // JSON解析失败忽略
          }
        }
      }
      // 流结束后确保状态为完成
      const t = taskQueue.value.find(t => t.id === taskId)
      if (t && t.status === 'running') {
        t.status = t.failed > 0 ? (t.completed > 0 ? 'completed' : 'failed') : 'completed'
        t.progress = 100
        t.endTime = Date.now()
      }
    })
    .catch(err => {
      const t = taskQueue.value.find(t => t.id === taskId)
      if (t) {
        t.status = 'failed'
        t.endTime = Date.now()
        t.failedTargets.push({ deviceIP: ip, error: err.message || '网络错误' })
      }
    })
}

// 执行单个任务
const executeTask = async (taskId) => {
  const task = taskQueue.value.find(t => t.id === taskId)
  console.log(`[executeTask] taskId: ${taskId}, task:`, task)
  
  if (!task || task.status !== 'pending') {
    console.log(`[executeTask] 任务不存在或状态不是pending，返回`)
    return
  }

  if (task.status === 'canceled') {
    return
  }
  
  console.log(`[executeTask] 开始执行任务，targets数量: ${task.targets.length}`)
  
  task.status = 'running'
  task.startTime = new Date()
  task.completed = 0
  task.failed = 0
  task.progress = 0
  task.failedTargets = []
  
  // 初始化每个设备的任务总数（用于上传文件/镜像任务）
  if (task.deviceProgress) {
    for (const deviceIP of task.deviceIps) {
      if (task.type === 'uploadFile') {
        // 统计该设备关联的云机数量（支持多种字段格式）
        const deviceTargets = task.targets.filter(t => 
          (t.deviceIP && t.deviceIP === deviceIP) ||
          (t.ip && t.ip === deviceIP) ||
          (t.device && t.device.ip && t.device.ip === deviceIP)
        )
        const machineCount = deviceTargets.reduce((count, target) => {
          return count + (target.machines?.length || 0)
        }, 0)
        task.deviceProgress[deviceIP] = {
          total: machineCount,
          completed: 0,
          failed: 0
        }
      } else if (task.type === 'uploadImage') {
        // 统计该设备的镜像上传任务数（支持多种字段格式）
        const deviceTargets = task.targets.filter(t => 
          (t.deviceIP && t.deviceIP === deviceIP) ||
          (t.ip && t.ip === deviceIP) ||
          (t.device && t.device.ip && t.device.ip === deviceIP)
        )
        task.deviceProgress[deviceIP] = {
          total: deviceTargets.length,
          completed: 0,
          failed: 0
        }
      }
    }
  }
  
  try {
    // 处理下载镜像任务
    if (task.type === 'downloadImage') {
      // 下载镜像任务不需要并发执行，直接返回，由后端事件处理进度和结果
      return
    }
    
    // 按设备IP分组，同一设备的任务串行执行，不同设备的任务可以并行
    // 这样可以避免同一设备被并发请求压垮
    const targetsByDevice = {}
    task.targets.forEach(target => {
      // 支持多种设备IP字段格式（deviceIp小写、deviceIP大写、ip字段）
      const deviceIP = target.deviceIp || target.deviceIP || target.ip || (target.device && target.device.ip)
      if (!deviceIP) {
        console.log(`[executeTask] 跳过没有deviceIP的target:`, target)
        return
      }
      if (!targetsByDevice[deviceIP]) {
        targetsByDevice[deviceIP] = []
      }
      targetsByDevice[deviceIP].push(target)
    })
    
    const deviceIPs = Object.keys(targetsByDevice)
    console.log(`[executeTask] deviceIPs: ${deviceIPs}, targetsByDevice:`, targetsByDevice)
    
    if (deviceIPs.length === 0) {
      console.log(`[executeTask] 没有设备IP，任务完成`)
      task.status = 'completed'
      task.endTime = new Date()
      return
    }
    
    const maxConcurrentDevices = 4 // 最多同时处理4个设备
    const runningCount = { value: 0 }
    
    // 串行执行同一设备的所有任务 (上传任务除外,上传任务完全并发)
    const processDeviceSerially = async (deviceIP, targets) => {
      try {
        if (task.status === 'canceled') {
          return
        }
        
        // ========== 特殊处理: 上传任务完全并发 ==========
        if (task.type === 'uploadFile') {
          // 创建所有上传任务(所有文件×所有容器,完全并发)
          const allUploadPromises = targets.map(async (target) => {
            if (task.status === 'canceled') {
              return
            }
            
            try {
              const { filePath, isAPK, deviceIP: uploadDeviceIP, deviceVersion, machines, apkOptions } = target
              const targetDevice = { ip: uploadDeviceIP, version: deviceVersion || 'v3' }
              
              const runUploadWithAuth = async (password, containerID) => {
                let result
                if (isAPK) {
                  // 传递 APK 安装选项
                  result = await InstallAPK(
                    uploadDeviceIP,
                    deviceVersion,
                    containerID,
                    filePath,
                    password || '',
                    apkOptions || {}
                  )
                } else {
                  result = await UploadFileToCloudMachine(
                    uploadDeviceIP,
                    deviceVersion,
                    containerID,
                    filePath,
                    password || ''
                  )
                }

                if (result && (result.code === 61 || result.message === 'Authentication Failed' || result.errorType === 'auth_required')) {
                  throw new Error('Authentication Failed')
                }
                return result
              }
              
              // 为每个文件创建所有容器的上传任务
              const machineUploadPromises = machines.map(async (machine) => {
                const containerID = machine.containerID
                const displayName = machine.name || machine.id || machine.indexNum || '云机'
                
                try {
                  const fileResult = await authRetry(targetDevice, async (password) => {
                    return await runUploadWithAuth(password, containerID)
                  })
                  
                  if (fileResult && fileResult.success) {
                    // 更新全局任务进度
                    task.completed++
                    console.log(`文件 ${filePath.split('\\').pop().split('/').pop()} 上传到云机 ${displayName} 成功`)
                    
                    // 立即更新设备进度
                    if (task.deviceProgress && uploadDeviceIP) {
                      if (!task.deviceProgress[uploadDeviceIP]) {
                        task.deviceProgress[uploadDeviceIP] = { total: 0, completed: 0, failed: 0 }
                      }
                      task.deviceProgress[uploadDeviceIP].completed++
                    }
                    
                    // APK安装成功
                    if (isAPK && fileResult.installed && fileResult.uploadPath) {
                      console.log(`APK安装成功,文件保留在: ${fileResult.uploadPath}`)
                    }
                    
                    return { success: true, machine: displayName }
                  } else {
                    // 更新全局任务进度
                    task.failed++
                    console.error(`文件 ${filePath.split('\\').pop().split('/').pop()} 上传到云机 ${displayName} 失败`)
                    
                    // 立即更新设备进度（失败）
                    if (task.deviceProgress && uploadDeviceIP) {
                      if (!task.deviceProgress[uploadDeviceIP]) {
                        task.deviceProgress[uploadDeviceIP] = { total: 0, completed: 0, failed: 0 }
                      }
                      task.deviceProgress[uploadDeviceIP].failed++
                    }
                    
                    return { success: false, machine: displayName, error: fileResult?.message }
                  }
                } catch (error) {
                  // 更新全局任务进度
                  task.failed++
                  console.error(`上传文件 ${filePath.split('\\').pop().split('/').pop()} 到云机 ${displayName} 失败:`, error.message)
                  
                  // 立即更新设备进度（异常）
                  if (task.deviceProgress && uploadDeviceIP) {
                    if (!task.deviceProgress[uploadDeviceIP]) {
                      task.deviceProgress[uploadDeviceIP] = { total: 0, completed: 0, failed: 0 }
                    }
                    task.deviceProgress[uploadDeviceIP].failed++
                  }
                  
                  return { success: false, machine: displayName, error: error.message }
                }
              })
              
              // 等待该文件的所有容器上传完成
              await Promise.allSettled(machineUploadPromises)
              
            } catch (error) {
              console.error(`处理上传任务失败:`, error)
            }
          })
          
          // 等待所有文件的所有容器上传完成(完全并发)
          await Promise.allSettled(allUploadPromises)
          return
        }
        // ========== 上传任务并发处理结束 ==========
        
        // 其他任务类型保持串行
        if (task.type === 'create' && targets.length > 0 && targets[0].isLocalImage) {
          const target = targets[0]
          const deviceVersion = target.deviceVersion || 'v3'
          
          const onlineUrl = target.localImageOnlineUrl
          console.log(`[批量创建] 设备 ${deviceIP} 本地镜像在线地址: ${onlineUrl}`)
          
          const savedPassword = getDevicePassword(deviceIP)
          const deviceImages = await GetImages(deviceIP, deviceVersion, savedPassword || '')
          console.log(`[批量创建] 设备 ${deviceIP} 镜像列表:`, deviceImages)
          
          let imageExists = false
          if (onlineUrl) {
            if (Array.isArray(deviceImages)) {
              imageExists = deviceImages.some(img => {
                const repoTags = img.RepoTags || img.imageUrl || img.Image
                if (Array.isArray(repoTags)) {
                  return repoTags.includes(onlineUrl) || repoTags.some(tag => tag.includes(onlineUrl))
                }
                return repoTags === onlineUrl || (typeof repoTags === 'string' && repoTags.includes(onlineUrl))
              })
            } else if (deviceImages && deviceImages.list) {
              imageExists = deviceImages.list.some(img => {
                const imgUrl = img.imageUrl || img.Image
                return imgUrl === onlineUrl || (typeof imgUrl === 'string' && imgUrl.includes(onlineUrl))
              })
            }
          }
          
          console.log(`[批量创建] 检查镜像 ${onlineUrl} 是否存在:`, imageExists)
          
          if (!imageExists) {
            console.log(`[批量创建] 设备 ${deviceIP} 镜像不存在，开始推送本地镜像...`)
            
            task.imageProgress = 0
            const progressInterval = setInterval(() => {
              if (task.status === 'canceled') {
                clearInterval(progressInterval)
                return
              }
              if (task.imageProgress < 90) {
                task.imageProgress += 10
              }
            }, 200)
            
            const password = getDevicePassword(deviceIP)
            const loadResult = await LoadImageToDevice(deviceIP, target.localImageUrl, deviceVersion, password || '')
            
            clearInterval(progressInterval)
            if (task.status === 'canceled') {
              return
            }
            task.imageProgress = 100
            
            console.log(`[批量创建] 设备 ${deviceIP} 镜像推送结果:`, loadResult)
            
            if (!loadResult.success) {
              console.error(`[批量创建] 设备 ${deviceIP} 镜像推送失败:`, loadResult.message)
              for (const t of targets) {
                task.failed++
                task.failedTargets.push({
                  ...t,
                  error: `镜像推送失败: ${loadResult.message || '未知错误'}`
                })
              }
              return
            }
          } else {
            console.log(`[批量创建] 设备 ${deviceIP} 镜像已存在（在线地址: ${onlineUrl}），跳过推送`)
            if (task.status !== 'canceled') {
              task.imageProgress = 100
            }
          }
          
          if (task.status === 'canceled') {
            return
          }
          task.currentStep = 'create'
        }
        
        for (const target of targets) {
          if (task.status === 'canceled') {
            break
          }
          try {
          // 根据任务类型创建对应的操作Promise
          const createOperationPromise = async () => {
            // 确保 containerName 是字符串
            let containerName = ''
            const targetName = target.name || target.id || target.ID
            if (typeof targetName === 'string') {
              containerName = targetName
            } else if (targetName && typeof targetName === 'object') {
              containerName = targetName.name || targetName.id || String(targetName)
            } else {
              containerName = String(targetName || '')
            }
            
            // uploadFile任务已在processDeviceSerially开头并发处理,这里不再处理
            switch (task.type) {
                case 'restart': {
                  const device = { ip: target.deviceIp, version: target.deviceVersion || 'v3' }
                  const actualDevice = cloudManageMode.value === 'slot' && selectedCloudDevice.value ? selectedCloudDevice.value : device
                  
                  // 重启容器前，清空该容器的截图缓存，避免显示旧截图
                  clearContainerScreenshotCache(actualDevice, target)
                  
                  if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                    await restartAndroidContainer(selectedCloudDevice.value, containerName)
                  } else {
                    await restartAndroidContainer(device, containerName)
                  }
                  return true
                }
                
                case 'reset': {
                  const device = { ip: target.deviceIp, version: target.deviceVersion || 'v3' }
                  const actualDevice = cloudManageMode.value === 'slot' && selectedCloudDevice.value ? selectedCloudDevice.value : device
                  
                  // 重置容器前，清空该容器的截图缓存，避免显示旧截图
                  clearContainerScreenshotCache(actualDevice, target)
                  
                  // 从任务中获取 start 参数（metadata 通过 spread 合并到 task 上），默认为 true
                  const resetStart = task.start !== undefined ? task.start : true
                  console.log('[重置任务执行] start参数:', resetStart, 'task.start:', task.start)
                  if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                    await resetAndroidContainer(selectedCloudDevice.value, containerName, null, resetStart)
                  } else {
                    await resetAndroidContainer(device, containerName, null, resetStart)
                  }
                  return true
                }
                
                case 'shutdown': {
                  if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                    await stopContainer(selectedCloudDevice.value, containerName)
                  } else {
                    await stopContainer({ ip: target.deviceIp, version: target.deviceVersion || 'v3' }, containerName)
                  }
                  return true
                }
                
                case 'delete': {
                  if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                    await deleteContainer(selectedCloudDevice.value, containerName)
                  } else {
                    await deleteContainer({ ip: target.deviceIp, version: target.deviceVersion || 'v3' }, containerName)
                  }
                  return true
                }
                
                case 'uploadImage': {
                  // 从 target 中提取设备信息和镜像路径
                  const uploadDeviceIP = target.deviceIP || target.ip || (target.device && target.device.ip)
                  const deviceVersion = target.deviceVersion || (target.device && target.device.version) || 'v3'
                  const imagePath = target.imagePath || task.imagePath
                  const imageName = task.imageName
                  
                  if (!uploadDeviceIP || !imagePath) {
                    throw new Error('缺少设备IP或镜像路径')
                  }
                  
                  const password = getDevicePassword(uploadDeviceIP)
                  const result = await LoadImageToDevice(uploadDeviceIP, imagePath, deviceVersion, password || '')
                  
                  if (result.success) {
                    return true
                  } else {
                    throw new Error(result.message || '上传失败')
                  }
                }
                
                case 'switchModel': {
                  // 获取传递的 modelInfo
                  const modelInfo = task.modelInfo || task.modelId
                  const modelName = task.modelName
                  
                  // V2容器使用一键新机接口，不支持指定机型
                  if (target.androidType === 'V2') {
                    // V2容器只能调用一键新机接口
                    const device = cloudManageMode.value === 'slot' && selectedCloudDevice.value 
                      ? selectedCloudDevice.value 
                      : { ip: target.deviceIp, version: target.deviceVersion || 'v3' }
                    
                    // 判断使用哪个IP和端口
                    let host, port
                    if (target.networkName === 'myt' || target.networkMode === 'myt' || target.network === 'myt') {
                      // myt网络：使用容器IP + 9082端口
                      host = target.ip
                      port = 9082
                    } else {
                      // 非myt网络：使用端口映射
                      host = device.ip
                      // OpenCecs 公网设备：deviceIp 含端口，提取纯 IP
                      if (host && host.includes(':')) host = host.split(':')[0]
                      port = extractPort9082(target) || 9082
                    }
                    
                    const modifyDevUrl = `http://${host}:${port}/modifydev?cmd=2`
                    console.log(`[批量新机-V2容器] 调用一键新机API: ${modifyDevUrl}`)
                    
                    const result = await HttpRequest({
                      url: modifyDevUrl,
                      method: 'GET'
                    })
                    
                    if (result.success) {
                      console.log(`[批量新机-V2容器] ${containerName} 一键新机成功`)
                      return true
                    } else {
                      throw new Error(`一键新机失败: ${result.status}`)
                    }
                  } else {
                    // 非V2容器使用切换机型接口
                    if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                      await switchCloudMachineModel(selectedCloudDevice.value, containerName, modelInfo, modelName)
                    } else {
                      await switchCloudMachineModel({ ip: target.deviceIp, version: target.deviceVersion || 'v3' }, containerName, modelInfo, modelName)
                    }
                    return true
                  }
                }
                
                case 'create': {
                  if (task.status === 'canceled') {
                    throw new Error('任务已取消')
                  }
                  const device = { ip: target.deviceIp, version: target.deviceVersion || 'v3', id: target.deviceId }
                  
                  // 通过 formOverride 传递任务参数，避免修改共享的 createForm 导致界面抖动
                  const formOverride = {
                    createType: target.createType || 'simulator',
                    imageCategory: target.isLocalImage ? 'local' : 'online',
                    imageSelect: target.isLocalImage ? '' : target.imageUrl,
                    customImageUrl: '',
                    localImageUrl: target.isLocalImage ? target.localImageUrl : '',
                    modelType: target.modelType || 'online',
                    localModel: target.localModel || '',
                    modelStatic: target.modelStatic || '',
                    vpcGroupId: target.vpcGroupId || '',
                    vpcNodeId: target.vpcNodeId || '',
                    vpcSelectMode: target.vpcNodeId === 'random' ? 'random' : 'specified',
                    macVlanIp: target.macVlanIp || '',
                    containerMacVlanIp: target.macVlanIp || '',
                    mytBridgeName: target.mytBridgeName || '',
                    containerNetworkCardType: target.networkCardType || 'private',
                    resolution: target.resolution || createForm.value.resolution,
                    customResolution: target.customResolution || createForm.value.customResolution,
                    dns: target.dns || createForm.value.dns,
                    customDns: target.customDns || createForm.value.customDns,
                    sandboxMode: target.sandboxMode !== undefined ? target.sandboxMode : createForm.value.sandboxMode,
                    sandboxSize: target.dataDiskSize ? parseInt(target.dataDiskSize) : createForm.value.sandboxSize,
                  }
                  
                  const success = await createCloudMachine(
                    device,
                    target.slot,
                    target.modelName,
                    () => task.status === 'canceled',
                    { start: target.start, formOverride }
                  )

                  // 容器创建成功后，为 OpenCecs 公网设备自动创建端口映射
                  // 仅对公网设备（IP 包含 ":"，即 publicIp:publicPort 格式）生效，局域网设备跳过
                  if (success && target.deviceIp && target.deviceIp.includes(':')) {
                    try {
                      if (opencecsManagementRef.value) {
                        // 从 OpenCecs 实例列表中找到匹配的实例 ID
                        let matchedInstance = opencecsManagementRef.value.getInstanceByDeviceIp?.(target.deviceIp)
                        
                        if (!matchedInstance) {
                          // fallback：按公网 IP 前缀从 openCecsPortMap 反查已知的 deviceIp
                          console.warn(`[端口映射] getInstanceByDeviceIp 未匹配到实例 (${target.deviceIp})，尝试 fallback...`)
                          // 刷新一次实例列表再重试
                          try {
                            await opencecsManagementRef.value.fetchInstances?.()
                            // 等待端口映射设置完成
                            await new Promise(r => setTimeout(r, 5000))
                            matchedInstance = opencecsManagementRef.value.getInstanceByDeviceIp?.(target.deviceIp)
                            if (matchedInstance) {
                              console.log(`[端口映射] fallback 成功找到实例: ${matchedInstance.instance_id}`)
                            }
                          } catch (retryErr) {
                            console.warn('[端口映射] fallback 刷新实例失败:', retryErr)
                          }
                        }
                        
                        if (matchedInstance) {
                          console.log(`[端口映射] 开始为公网设备创建容器端口映射 (${target.deviceIp}), instanceId=${matchedInstance.instance_id}`)
                          await opencecsManagementRef.value.ensureContainerPortMappings(matchedInstance.instance_id, target.deviceIp)
                        } else {
                          console.warn(`[端口映射] ⚠️ 公网设备 ${target.deviceIp} 未找到匹配的 OpenCecs 实例，跳过端口映射。请检查 OpenCecs 是否已登录并刷新实例列表。`)
                        }
                      } else {
                        console.warn(`[端口映射] ⚠️ opencecsManagementRef 不可用，无法为公网设备 ${target.deviceIp} 创建端口映射`)
                      }
                    } catch (e) {
                      console.warn('[端口映射] 容器创建后自动映射失败:', e)
                    }
                  }

                  return success
                }
                
                default:
                  throw new Error(`未知的任务类型: ${task.type}`)
              }
            }
          
          // 创建带超时的Promise（超时后返回null表示需要验证）
          const operationWithTimeout = async () => {
            if (task.timeout === 0) {
              return await createOperationPromise()
            }
            
            const timeoutMs = task.timeout
            console.log(`开始执行${task.type}操作，超时时间: ${timeoutMs}ms`)
            
            let timeoutId
            const timeoutPromise = new Promise((_, reject) => {
              timeoutId = setTimeout(() => {
                reject(new Error('任务超时'))
              }, timeoutMs)
            })
            
            try {
              const result = await Promise.race([
                createOperationPromise(),
                timeoutPromise
              ])
              clearTimeout(timeoutId)
              return result
            } catch (error) {
              clearTimeout(timeoutId)
              if (error.message === '任务超时') {
                console.warn(`操作超时但可能仍在执行中: ${target.name || target.id}`)
                return null
              }
              throw error
            }
          }
          
          let result = await operationWithTimeout()
          
          // 对于restart和reset操作，如果超时则验证实际结果
          if (result === null && (task.type === 'restart' || task.type === 'reset')) {
            console.log('超时后验证操作结果...')
            await new Promise(resolve => setTimeout(resolve, 5000))
            
            try {
              const verifyContainerName = target.name || target.id || target.ID
              const verifyDevice = cloudManageMode.value === 'slot' && selectedCloudDevice.value
                ? selectedCloudDevice.value
                : { ip: target.deviceIp, version: target.deviceVersion || 'v3' }
              
              const containers = await getContainers(verifyDevice)
              const targetContainer = containers.find(c => 
                (c.name || c.ID || c.id) === verifyContainerName ||
                (c.Names && c.Names.includes(verifyContainerName))
              )
              
              if (targetContainer) {
                if (targetContainer.status === 'running') {
                  console.log('验证成功：容器已执行操作并运行中')
                  result = true
                } else {
                  console.warn('验证结果：容器状态异常', targetContainer.status)
                  result = false
                }
              } else {
                console.warn('验证结果：找不到容器（可能被删除）')
                result = task.type === 'delete' ? true : false
              }
            } catch (verifyError) {
              console.error('验证操作结果失败:', verifyError)
              result = false
            }
          }
          
          // 如果结果包含manualCount标记，说明内部已经处理了计数，这里不再重复计数
          if (result && typeof result === 'object' && result.manualCount) {
            // 既然已经手动处理了，这里不需要做任何事
            // 但如果是非uploadFile任务（理论上不应该进入这里，因为只有uploadFile返回这个对象），
            // 我们可能还是需要更新deviceProgress。
            // 不过目前只有uploadFile返回这个，且uploadFile内部已经更新了deviceProgress。
          } else if (result === true) {
            task.completed++
            // 对于非uploadFile任务，更新对应设备的进度
            if (task.deviceProgress && target.deviceIP && task.type !== 'uploadFile') {
              if (!task.deviceProgress[target.deviceIP]) {
                task.deviceProgress[target.deviceIP] = { total: 0, completed: 0, failed: 0 }
              }
              task.deviceProgress[target.deviceIP].completed++
            }
          } else if (result === false) {
            task.failed++
            task.failedTargets.push({
              ...target,
              error: '执行失败'
            })
            // 对于非uploadFile任务，更新对应设备的进度
            if (task.deviceProgress && target.deviceIP && task.type !== 'uploadFile') {
              if (!task.deviceProgress[target.deviceIP]) {
                task.deviceProgress[target.deviceIP] = { total: 0, completed: 0, failed: 0 }
              }
              task.deviceProgress[target.deviceIP].failed++
            }
          } else {
            task.failed++
            task.failedTargets.push({
              ...target,
              error: '任务超时'
            })
            // 更新对应设备的进度
            if (task.deviceProgress && target.deviceIP) {
              if (!task.deviceProgress[target.deviceIP]) {
                task.deviceProgress[target.deviceIP] = { total: 0, completed: 0, failed: 0 }
              }
              task.deviceProgress[target.deviceIP].failed++
            }
          }
        } catch (error) {
          if (task.status === 'canceled' || error.message === '任务已取消') {
            break
          }
          task.failed++
          task.failedTargets.push({
            ...target,
            error: error.message || '未知错误'
          })
          console.error(`执行任务失败: ${error.message}`)
          if (task.deviceProgress && target.deviceIP && task.type !== 'uploadFile') {
            if (!task.deviceProgress[target.deviceIP]) {
              task.deviceProgress[target.deviceIP] = { total: 0, completed: 0, failed: 0 }
            }
            task.deviceProgress[target.deviceIP].failed++
          }
        } finally {
          if (task.status !== 'canceled') {
            if (task.type === 'create') {
              if (task.currentStep === 'image' && task.imageProgress !== null) {
                task.progress = Math.round(task.imageProgress / 2)
              } else if (task.currentStep === 'create') {
                const currentSlot = task.completed + task.failed
                task.progress = 50 + Math.round(currentSlot / task.total * 50)
              } else {
                const currentSlot = task.completed + task.failed
                task.progress = Math.round(currentSlot / task.total * 100)
              }
            } else {
              task.progress = Math.round((task.completed + task.failed) / task.total * 100)
            }
          }
        }
        }
      } finally {
        runningCount.value--
      }
    }
    
    // 并发处理不同设备
    const devicePromises = []
    for (const deviceIP of deviceIPs) {
      // 等待有可用的并发槽位
      while (runningCount.value >= maxConcurrentDevices) {
        await new Promise(resolve => setTimeout(resolve, 100))
      }
      
      runningCount.value++
      devicePromises.push(processDeviceSerially(deviceIP, targetsByDevice[deviceIP]))
    }
    
    await Promise.all(devicePromises)
    
    if (task.status === 'canceled') {
      return
    }
    
    // 任务完成
    task.status = task.failed === 0 ? 'completed' : 'failed'
    task.endTime = new Date()
    
    // 计算实际上传成功的云机数量（uploadFile类型）
    let actualSuccessMachines = task.completed
    let actualFailMachines = task.failed
    if (task.type === 'uploadFile') {
      actualSuccessMachines = task.targets.filter(t => t.machines).reduce((count, target) => {
        return count + (target.machines?.length || 0)
      }, 0)
      actualFailMachines = 0
    }
    
    // 显示任务完成通知
    const taskTypeText = task.type === 'restart' ? '批量重启' : task.type === 'reset' ? '批量重置' : task.type === 'shutdown' ? '批量关机' : task.type === 'create' ? '批量创建' : task.type === 'delete' ? '批量删除' : task.type === 'switchModel' ? (task.operation === 'new' ? '批量新机' : '批量切换机型') : task.type === 'uploadFile' ? '批量上传' : task.type === 'uploadImage' ? '批量上传镜像' : task.type === 'downloadImage' ? '下载镜像' : task.type === 'updateImage' ? '批量更新镜像' : '批量操作'
    if (task.status === 'completed') {
      if (task.type === 'uploadFile') {
        ElMessage.success(`${taskTypeText}任务已完成，成功上传到 ${actualSuccessMachines} 个云机`)
      } else {
        ElMessage.success(`${taskTypeText}任务已完成，成功 ${task.completed} 个云机`)
      }
    } else {
      if (task.type === 'uploadFile') {
        ElMessage.warning(`${taskTypeText}任务部分失败，成功 ${actualSuccessMachines} 个云机，失败 ${actualFailMachines} 个云机，请查看任务列表获取详细信息`)
      } else {
        ElMessage.warning(`${taskTypeText}任务部分失败，成功 ${task.completed} 个云机，失败 ${task.failed} 个云机，请查看任务列表获取详细信息`)
      }
    }
    
    // 刷新容器列表
    if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
      await fetchAndroidContainers(selectedCloudDevice.value, true)
    }
    
    // 如果是删除操作且备份列表可见，刷新备份列表
    if (task.type === 'delete' && backupListVisible.value) {
      initBackupList()
    }
  } catch (error) {
    task.status = 'failed'
    task.endTime = new Date()
    console.error(`任务执行失败: ${error.message}`)
  }
}

// 取消任务
const cancelTask = async (taskId) => {
  const task = taskQueue.value.find(t => t.id === taskId)
  if (task && task.status === 'running') {
    // 取消下载镜像任务的特殊处理
    if (task.type === 'downloadImage') {
      console.log('取消下载任务:', taskId)
      
      try {
        // 调用后端API取消下载并删除未完成的文件
        await CancelImageDownload()
        console.log('取消下载API调用成功')
      } catch (error) {
        console.error('取消下载API调用失败:', error)
      }
      
      // 先标记任务为已取消
      task.status = 'canceled'
      task.endTime = new Date()
      task.progress = 0
      
      // 延迟重置下载状态，避免残留事件干扰
      setTimeout(() => {
        // 再次确认是当前任务才重置
        if (currentDownloadTaskId.value === taskId) {
          isDownloadingImage.value = false
          currentDownloadImage.value = null
          currentDownloadTaskId.value = null
          downloadProgress.value = 0
          downloadStartTime.value = 0
          console.log('下载状态已完全重置')
        }
      }, 200)
    } else if (task.type === 'uploadImage') {
      // 取消上传镜像任务的特殊处理
      console.log('[上传镜像] 取消上传任务:', taskId)
      
      try {
        // 调用后端API取消上传
        await CancelImageUpload()
        console.log('[上传镜像] 取消上传API调用成功')
      } catch (error) {
        console.error('[上传镜像] 取消上传API调用失败:', error)
      }
      
      // 先标记任务为已取消
      task.status = 'canceled'
      task.endTime = new Date()
      
      // 延迟重置上传状态，避免残留事件干扰
      setTimeout(() => {
        if (isUploadingImage.value || isUploadingToMultipleDevices.value) {
          isUploadingImage.value = false
          isUploadingToMultipleDevices.value = false
          currentUploadImage.value = null
          uploadProgress.value = 0
          console.log('[上传镜像] 上传状态已完全重置')
        }
      }, 200)
    } else {
      task.status = 'canceled'
      task.endTime = new Date()
      task.progress = 0
      if (task.type === 'create') {
        task.imageProgress = 0
        task.currentStep = null
      }
    }
    
    ElMessage.info('任务已取消')
  }
}

// 重试任务（只重试失败的）
const retryFailedTask = async (taskId) => {
  const task = taskQueue.value.find(t => t.id === taskId)
  if (!task || task.status !== 'failed' || task.failedTargets.length === 0) return
  
  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      `确定要重试失败的 ${task.failedTargets.length} 个任务吗？`, 
      '重试失败任务', 
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 从失败目标中提取设备IP信息
    const failedDeviceIps = new Set()
    task.failedTargets.forEach(target => {
      if (target.deviceIp) {
        failedDeviceIps.add(target.deviceIp)
      } else if (target.device && target.device.ip) {
        failedDeviceIps.add(target.device.ip)
      }
    })
    
    // 创建新任务，只包含失败的目标
    const newTask = {
      ...task,
      id: generateTaskId(),
      status: 'pending',
      total: task.failedTargets.length,
      completed: 0,
      failed: 0,
      progress: 0,
      targets: task.failedTargets.map(target => {
        // 对于 uploadImage 类型，确保 target 包含重试所需的信息
        if (task.type === 'uploadImage') {
          return {
            ...target,
            deviceIP: target.deviceIP || target.ip || (target.device && target.device.ip),
            deviceVersion: target.deviceVersion || (target.device && target.device.version) || 'v3',
            imagePath: task.imagePath || target.imagePath || ''
          }
        }
        return target
      }),
      deviceIps: Array.from(failedDeviceIps), // 只包含失败目标相关的设备IP
      failedTargets: [],
      startTime: null,
      endTime: null
    }
    
    // 添加到队列并执行
    taskQueue.value.unshift(newTask)
    executeTask(newTask.id)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重试任务失败:', error)
      ElMessage.error(`重试任务失败: ${error.message || '未知错误'}`)
    }
  }
}

// 清理任务
const handleClearTasks = async () => {
  try {
    // 可清理的状态：已完成、已取消、失败
    const cleanableStatuses = ['completed', 'canceled', 'failed']
    const hasCleanableTasks = taskQueue.value.some(task => cleanableStatuses.includes(task.status))
    
    if (hasCleanableTasks) {
      await ElMessageBox.confirm(
        '确定要清理所有已完成、已取消和失败的任务吗？', 
        '清理任务队列', 
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'info'
        }
      )
      
      // 过滤掉已完成、已取消、失败的任务
      taskQueue.value = taskQueue.value.filter(task => !cleanableStatuses.includes(task.status))
      ElMessage.success('任务清理成功')
    } else {
      ElMessage.info('没有可清理的任务')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清理任务失败:', error)
      ElMessage.error(`清理任务失败: ${error.message || '未知错误'}`)
    }
  }
}

// 显示更新提示弹窗
const handleShowUpdateDialog = (info) => {
  updateInfo.value = info
  updateDialogVisible.value = true
}

// 批量切换机型确认函数
const confirmBatchSwitchModel = async () => {
  if (!isModelSlotsValid.value) {
    ElMessage.warning('请确保所有机型都已选择且每个机型至少分配一个坑位')
    return
  }
  
  if (batchSwitchModelTargets.value.length === 0) {
    ElMessage.warning('没有要切换的云机')
    return
  }
  
  try {
    // 准备确认信息
    const modelInfo = []
    modelSlots.value.forEach((modelSlot, index) => {
      if (modelSlot.assignedSlots.length > 0) {
        if (modelSlot.modelId === 'random') {
          modelInfo.push(`随机 (${modelSlot.assignedSlots.length}个坑位)`)
        } else {
          let modelName = modelSlot.modelId
          if (!modelSlot.type || modelSlot.type === 'online') {
            const model = phoneModels.value.find(m => m.id === modelSlot.modelId)
            if (model) modelName = model.name
          }
          modelInfo.push(`${modelName} (${modelSlot.assignedSlots.length}个坑位)`)
        }
      }
    })
    
    // 显示确认对话框
    await ElMessageBox.confirm(
      `确定要执行批量新机操作吗？\n分配情况：\n${modelInfo.join('\n')}`, 
      '批量新机确认', 
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 标记为正在切换机型
    batchSwitchingModel.value = true
    
    // 确定操作类型（如果是通过批量新机按钮触发的，则显示为批量新机）
    const operationType = batchSwitchModelOperationType.value || 'switchModel'
    
    // 为每个机型创建一个任务
    modelSlots.value.forEach((modelSlot) => {
      if (modelSlot.assignedSlots.length === 0) {
        return
      }
      
      let model = null
      let modelName = ''
      
      // 处理随机机型情况
      if (modelSlot.modelId === 'random') {
        modelName = '随机'
      } else {
        if (!modelSlot.type || modelSlot.type === 'online') {
          const m = phoneModels.value.find(m => m.id === modelSlot.modelId)
          if (m) {
            model = { id: m.id, name: m.name }
            modelName = m.name
          }
        } else {
          // 本地和备份机型直接使用modelId(即name)
          model = { id: modelSlot.modelId, name: modelSlot.modelId }
          modelName = modelSlot.modelId
        }
        
        if (!model) {
          return
        }
      }
      
      // 将批量切换机型任务添加到任务队列
      const taskId = addTaskToQueue('switchModel', modelSlot.assignedSlots, {
        modelId: modelSlot.modelId === 'random' ? 'random' : model.id, 
        
        // 这里我们传递 modelInfo 对象，类似于单机切换
        modelInfo: {
          value: modelSlot.modelId === 'random' ? 'random' : modelSlot.modelId,
          type: modelSlot.type || 'online'
        },
        modelName: modelName,
        operation: operationType,
        timeout: 30000 // 30秒超时
      })
      
      // 执行任务
      executeTask(taskId)
    })
    
    // 关闭对话框
    batchSwitchModelDialogVisible.value = false
    
    ElMessage.success(`批量新机任务已添加到队列并开始执行`)

    // 重置状态
    batchSwitchingModel.value = false
    selectedBatchModelId.value = ''
    selectedBatchModelName.value = ''
    batchSwitchModelTargets.value = []
    batchSwitchModelOperationType.value = 'switchModel' // 重置为默认操作类型
    modelSlots.value = [] // 清空机型分配槽
    draggingSlot.value = null // 重置拖拽状态
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量切换机型失败:', error)
      ElMessage.error(`批量切换机型失败: ${error.message || '未知错误'}`)
      batchSwitchingModel.value = false
    }
  }
}

// 处理批量切换机型取消操作
const handleBatchSwitchModelCancel = () => {
  // 关闭对话框
  batchSwitchModelDialogVisible.value = false
  
  // 重置相关状态
  setTimeout(() => {
    selectedBatchModelId.value = ''
    selectedBatchModelName.value = ''
    batchSwitchModelTargets.value = []
    batchSwitchModelOperationType.value = 'switchModel'
    modelSlots.value = [] // 清空机型分配槽
    draggingSlot.value = null // 重置拖拽状态
  }, 100)
}

// 批量新机 - 机型槽管理
const addNewModelSlot = async (type = 'online') => {
  // 检查是否需要获取机型列表
  const targets = batchSwitchModelTargets.value
  if (targets.length > 0) {
    const deviceIp = targets[0].deviceIp || (targets[0].device && targets[0].device.ip)
    
    if (deviceIp) {
      if (type === 'local') {
        // 如果本地机型列表为空，尝试获取
        if (localPhoneModels.value.length === 0) {
           await getLocalPhoneModels(deviceIp)
        }
      } else if (type === 'backup') {
        // 如果备份机型列表为空，尝试获取
        if (backupPhoneModels.value.length === 0) {
           await fetchBackupModels(deviceIp)
        }
      } else if (type === 'online') {
        // 如果线上机型列表为空，尝试获取
        if (phoneModels.value.length === 0) {
           await getV3PhoneModels(deviceIp)
        }
      }
    }
  }

  modelSlots.value.push({
    modelId: '',
    type: type, // 'online', 'local', 'backup'
    assignedSlots: []
  })
}

const removeModelSlot = (index) => {
  // 将该机型槽中的坑位返回到可用列表
  const modelSlot = modelSlots.value[index]
  if (modelSlot) {
    modelSlot.assignedSlots = []
  }
  // 删除该机型槽
  modelSlots.value.splice(index, 1)
}

// 批量新机 - 拖拽处理
const handleSlotDragStart = (event, slot) => {
  // V2容器不允许拖拽
  if (slot.androidType === 'V2') {
    event.preventDefault()
    ElMessage.warning('V2容器不支持指定机型，只能使用随机机型')
    return
  }
  draggingSlot.value = slot.id || slot.indexNum
  event.dataTransfer.setData('text/plain', JSON.stringify(slot))
}

const handleSlotDropInModelSlot = (event, modelSlotIndex) => {
  event.preventDefault()
  const slot = JSON.parse(event.dataTransfer.getData('text/plain'))
  const slotId = slot.id || slot.indexNum
  
  // V2容器不允许拖拽到非随机机型槽
  if (slot.androidType === 'V2' && modelSlots.value[modelSlotIndex].modelId !== 'random') {
    ElMessage.warning('V2容器不支持指定机型，只能使用随机机型')
    draggingSlot.value = null
    return
  }
  
  // 从所有机型槽中移除该坑位
  modelSlots.value.forEach(modelSlot => {
    modelSlot.assignedSlots = modelSlot.assignedSlots.filter(s => (s.id || s.indexNum) !== slotId)
  })
  
  // 添加到当前机型槽
  modelSlots.value[modelSlotIndex].assignedSlots.push(slot)
  draggingSlot.value = null
}

const handleSlotDropInAvailableArea = (event, slot) => {
  event.preventDefault()
  const draggedSlot = JSON.parse(event.dataTransfer.getData('text/plain'))
  const slotId = draggedSlot.id || draggedSlot.indexNum
  
  // 从所有机型槽中移除该坑位
  modelSlots.value.forEach(modelSlot => {
    modelSlot.assignedSlots = modelSlot.assignedSlots.filter(s => (s.id || s.indexNum) !== slotId)
  })
  
  draggingSlot.value = null
}

// 初始化机型槽
const initModelSlots = () => {
  // 确保至少有一个机型槽
  if (modelSlots.value.length === 0) {
    // 创建一个初始机型槽，并将所有选中的坑位默认放入其中
    modelSlots.value = [{
      modelId: 'random',
      assignedSlots: [...batchSwitchModelTargets.value] // 默认将所有选中的坑位放入第一个机型槽
    }]
  }
}

// 执行任务队列
const executeTaskQueue = () => {
  for (const task of taskQueue.value) {
    if (task.status === 'pending') {
      executeTask(task.id)
    }
  }
}

// 批量切换机型相关状态
const batchSwitchModelDialogVisible = ref(false)
const batchSwitchingModel = ref(false)
const selectedBatchModelName = ref('')
const selectedBatchModelId = ref('')
const batchSwitchModelTargets = ref([])
const batchSwitchModelOperationType = ref('switchModel') // 'switchModel' 或 'new'

// 批量新机 - 机型分配相关状态
const modelSlots = ref([]) // 机型分配槽列表
const draggingSlot = ref(null) // 当前拖拽的槽位
const isModelSlotsValid = computed(() => {
  // 验证所有机型都已选择且每个机型至少分配一个坑位
  return modelSlots.value.length > 0 && 
         modelSlots.value.every(slot => slot.modelId && slot.assignedSlots.length > 0) &&
         // 确保每个坑位只分配给一个机型
         new Set(modelSlots.value.flatMap(slot => {
           return slot.assignedSlots.map(s => s.id || s.indexNum);
         })).size === batchSwitchModelTargets.value.length;
})

// 批量操作处理


// 从设备名称中提取设备型号
const getDeviceTypeName = (deviceName) => {
  if (!deviceName) return 'unknown'
  // 从设备名称中提取型号，例如q1_v2 -> q1, p1_v3 -> p1
  const parts = deviceName.split('_')
  return parts[0] || 'unknown'
}

// 格式化实例名称，隐藏特定格式的前缀
const formatInstanceName = (name) => {
  if (!name) return name
  
  // 匹配格式：[任意字符]_数字_名称
  // 例如：p1e847b84af914895b56a14557d1813d_2_T00022222222 -> T00022222222
  // 例如：p1e847b84af914895b56a14557d1813d_4_sjz_cs -> sjz_cs
  const match = name.match(/^.+_\d+_(.+)$/)
  if (match && match.length > 1) {
    return match[1] // 只返回最后一个下划线后的名称部分
  }
  
  return name // 返回原始名称
}


// 格式化实例机型名称
const formatInstanceModel = (path) => {
  if (!path) return ''
  
  // 匹配格式：[任意字符]_数字_名称
  // 例如：p1e847b84af914895b56a14557d1813d_2_T00022222222 -> T00022222222
  // 例如：p1e847b84af914895b56a14557d1813d_4_sjz_cs -> sjz_cs
  // 例如：/mmc/data/.../22111317PG -> 22111317PG
  
  // 尝试从路径中提取最后一部分作为机型
  const pathParts = path.split('/')
  const lastPart = pathParts[pathParts.length - 1]
  
  // 检查是否是有效的机型名称（非空且不包含路径分隔符）
  if (lastPart && !lastPart.includes('/')) {
    return lastPart
  }
  
  return path // 无法提取时返回原始名称
}

// 获取镜像显示名称 - 优化版
const getImageDisplayName = (imageUrl) => {
  if (!imageUrl) return '未知镜像'
  
  // 清理镜像URL，移除可能的协议和多余路径
  const cleanedUrl = imageUrl.toLowerCase()
  
  // 1. 从镜像列表中查找匹配的镜像
  let matchedImage = null
  
  // 精确匹配：检查image.url是否与imageUrl完全匹配
  matchedImage = imageList.value.find(image => {
    return image.url && image.url.toLowerCase() === cleanedUrl
  })

  
  // 如果没有精确匹配，尝试模糊匹配：检查image.url是否是imageUrl的一部分
  if (!matchedImage) {
    matchedImage = imageList.value.find(image => {
      return image.url && cleanedUrl.includes(image.url.toLowerCase())
    })
  }
  
  // 如果没有找到，尝试反向匹配：检查imageUrl是否是image.url的一部分
  if (!matchedImage) {
    matchedImage = imageList.value.find(image => {
      return image.url && image.url.toLowerCase().includes(cleanedUrl)
    })
  }
  
  // 如果仍然没有找到，尝试匹配镜像名称
  if (!matchedImage) {
    // 从URL中提取镜像名称部分用于匹配
    const urlParts = cleanedUrl.split('/')
    const urlNamePart = urlParts[urlParts.length - 1]
    
    matchedImage = imageList.value.find(image => {
      return image.name && image.name.toLowerCase().includes(urlNamePart)
    })
  }
  
  if (matchedImage) {
    // 2. 如果找到匹配的镜像，返回其名称
    return matchedImage.name || matchedImage.url
  }
  
  // 3. 如果没有找到匹配的镜像，从URL中提取用户友好的名称
  
  // 处理本地文件路径
  if (cleanedUrl.includes('\\') || cleanedUrl.endsWith('.tar.gz')) {
    // 本地镜像路径，提取文件名
    const pathParts = cleanedUrl.split(/[\\/]/)
    let fileName = pathParts[pathParts.length - 1]
    // 移除文件扩展名
    fileName = fileName.replace('.tar.gz', '')
    fileName = fileName.replace('.tar', '')
    return fileName
  }
  
  // 处理Docker镜像URL
  // 例如：registry.magicloud.tech/magicloud/dobox-android13:Q1 -> dobox-android13:Q1
  // 例如：docker.io/library/nginx:latest -> nginx:latest
  const parts = cleanedUrl.split('/')
  if (parts.length > 0) {
    let imageName = parts[parts.length - 1]
    
    // 进一步优化：如果镜像名包含registry或magicloud等关键词，尝试提取更友好的名称
    const friendlyNameMatch = imageName.match(/([a-zA-Z0-9_-]+):([a-zA-Z0-9._-]+)$/)
    if (friendlyNameMatch) {
      return friendlyNameMatch[0] // 返回 镜像名:标签 格式
    }
    
    return imageName
  }
  
  // 4. 否则返回原始URL
  return imageUrl
}

// 获取设备类型颜色
const getDeviceTypeColor = (deviceName) => {
  const deviceType = getDeviceTypeName(deviceName)
  const colorMap = {
    'q1': '#409EFF', // 蓝色
    'p1': '#67C23A', // 绿色
    'm48': '#E6A23C', // 黄色
    'c1': '#F56C6C', // 红色
    'a1': '#909399' // 灰色
  }
  return colorMap[deviceType] || '#909399' // 默认灰色
}

// 解析容器坑位编号
const parseContainerSlot = (container, device) => {
  console.log('parseContainerSlot called with container:', container?.name, 'device:', device?.ip, 'version:', device?.version);
  
  // 1. 识别系统插件容器，直接返回null
  const image = container.Image || container.image;
  const name = container.Name || container.Names?.[0];
  const isSystemContainer = image?.includes('myt_sdk') || 
                           image?.includes('myt_vpc_plugin') ||
                           name?.includes('myt_sdk') ||
                           name?.includes('myt_vpc_plugin');
  
  if (isSystemContainer) {
    console.log('System container detected, returning null');
    return null;
  }
  
  // 2. 优先从容器的indexNum字段获取
  if (container.indexNum) {
    console.log('Found slot from indexNum:', container.indexNum);
    return container.indexNum;
  }
  
  // 对于Docker API返回的容器，尝试从Labels获取idx
  if (container.Config && container.Config.Labels && container.Config.Labels.idx) {
    const slot = parseInt(container.Config.Labels.idx);
    console.log('Found slot from Config.Labels.idx:', slot);
    return slot;
  }
  
  // 尝试从docker inspect的Labels直接获取idx（不同Docker API版本可能有不同的字段）
  if (container.Labels && container.Labels.idx) {
    const slot = parseInt(container.Labels.idx);
    console.log('Found slot from Labels.idx:', slot);
    return slot;
  }
  
  // 从设备路径推断idx（参考api/main.go的逻辑）
  if (container.HostConfig && container.HostConfig.Devices) {
    for (const dev of container.HostConfig.Devices) {
      if (dev.PathInContainer && dev.PathInContainer.includes('/dev/vndbinder')) {
        const parts = dev.PathOnHost.split('binder')
        if (parts.length > 1) {
          const num = parseInt(parts[1])
          if (!isNaN(num)) {
            const slot = Math.floor(num / 3);
            console.log('Found slot from device path:', slot);
            return slot;
          }
        }
      }
    }
  }
  
  // 从容器名称中提取坑位编号，例如 "android-1" -> 1
  if (container.Name || container.Names) {
    const name = container.Name || container.Names[0];
    if (name) {
      const match = name.match(/-(\d+)/);
      if (match && match[1]) {
        const slot = parseInt(match[1]);
        console.log('Found slot from container name:', slot);
        return slot;
      }
    }
  }
  
  // 默认返回null
  console.log('No slot found, returning null');
  return null;
}

// 处理主机管理设备选择变化


// 单个设备删除处理


// 批量删除设备处理


// 获取设备版本信息 - 优化版
const handleGetDeviceVersion = async (device, isFromUpgrade = false) => {
  try {
    const now = Date.now()
    const lastTime = lastCheckTime.value.get(device.id) || 0
    
    // 检查是否在最小检查间隔内
    if (now - lastTime < MIN_CHECK_INTERVAL) {
      console.log(`设备 ${device.ip} 最近已查询过，跳过本次检查`)
      if (!isFromUpgrade) {
        ElMessage.warning(`设备 ${device.ip} 最近已查询过，请稍后再试`)
      }
      return
    }
    
    // 只在手动调用时显示全局loading
    if (!isFromUpgrade) {
      loading.value = true
    }
    
    console.log('获取设备版本信息:', device.ip)
    
    // 手动查询添加到优先级队列
    addToVersionCheckQueue(device, true)
    
    // 立即触发一次队列处理
    batchProcessVersionCheckQueue()
    
    // 同时获取版本信息
    fetchV3LatestInfo(device)
    
    // 如果是手动调用，优化loading显示
    if (!isFromUpgrade) {
      // 设置1.5秒超时，避免loading显示太久
      const loadingTimeout = setTimeout(() => {
        if (loading.value) {
          loading.value = false
          ElMessage.info(`设备 ${device.ip} 版本检查正在进行中，请稍候查看结果`)
        }
      }, 1500)
      
      // 监听版本信息变化，及时关闭loading
      const checkVersionUpdate = setInterval(() => {
        const version = deviceVersionInfo.value.get(device.id)
        const lastCheck = lastCheckTime.value.get(device.id) || 0
        
        if (lastCheck >= now && version) {
          clearInterval(checkVersionUpdate)
          clearTimeout(loadingTimeout)
          if (loading.value) {
            loading.value = false
            ElMessage.success(`设备 ${device.ip} 版本信息已更新`)
          }
        }
      }, 200)
      
      // 最多检查10次
      setTimeout(() => {
        clearInterval(checkVersionUpdate)
      }, 2000)
    }
  } catch (error) {
    console.error('获取设备版本信息失败:', error)
    if (!isFromUpgrade) {
      loading.value = false
      ElMessage.error(`获取设备 ${device.ip} 版本信息失败: ${error.message}`)
    } else {
      console.error('升级后版本检查失败:', error)
    }
  }
}

// 升级设备


// 清理客户端数据功能


// API版本检查队列
const versionCheckQueue = ref([])
const priorityQueue = ref([]) // 优先级队列，用于手动查询
const isProcessingQueue = ref(false)
const versionCheckInterval = ref(null)
const lastCheckTime = ref(new Map()) // 记录每个设备的最后检查时间，避免频繁查询
const MAX_CONCURRENT_CHECKS = 3 // 最大并发检查数
const CHECK_INTERVAL = 500 // 检查间隔缩短到500ms
const MIN_CHECK_INTERVAL = 3000 // 同一设备最小检查间隔3秒

// 设备绑定状态查询队列
const deviceBindCheckQueue = ref([])
const isProcessingBindQueue = ref(false)
const deviceBindCheckInterval = ref(null)
const lastBindCheckTime = ref(new Map()) // 记录每个设备的最后绑定状态检查时间
const MAX_CONCURRENT_BIND_CHECKS = 1 // 最大并发绑定状态查询数
const BIND_CHECK_INTERVAL = 1000 // 绑定状态查询间隔1秒
const MIN_BIND_CHECK_INTERVAL = 5000 // 同一设备最小绑定状态检查间隔5秒

// 添加设备到版本检查队列
const addToVersionCheckQueue = (device, isPriority = false) => {
  const now = Date.now()
  const lastTime = lastCheckTime.value.get(device.id) || 0
  
  // 避免短时间内重复查询同一设备
  if (now - lastTime < MIN_CHECK_INTERVAL) {
    console.log(`设备 ${device.ip} 最近已查询过，跳过本次检查`)
    return
  }
  
  // 检查是否已在队列中
  const isInMainQueue = versionCheckQueue.value.some(item => item.id === device.id)
  const isInPriorityQueue = priorityQueue.value.some(item => item.id === device.id)
  
  if (isInMainQueue || isInPriorityQueue) {
    // console.log(`设备 ${device.ip} 已在检查队列中，跳过重复添加`)
    return
  }
  
  if (isPriority) {
    priorityQueue.value.push(device)
    // console.log(`设备 ${device.ip} 已添加到优先级检查队列`)
  } else {
    versionCheckQueue.value.push(device)
    // console.log(`设备 ${device.ip} 已添加到版本检查队列`)
  }
}

// 处理版本检查队列 - 支持并发检查
const processVersionCheckQueue = async () => {
  if (isProcessingQueue.value) {
    return
  }
  
  isProcessingQueue.value = true
  
  let currentDevice = null
  
  try {
    // 优先处理优先级队列
    currentDevice = priorityQueue.value.shift() || versionCheckQueue.value.shift()
    
    if (!currentDevice) {
      isProcessingQueue.value = false
      return
    }
    
    console.log(`开始检查设备 ${currentDevice.ip} 版本信息`)
    
    // 记录检查时间
    lastCheckTime.value.set(currentDevice.id, Date.now())
    
    // 设置API调用超时
    const versionInfo = await Promise.race([
      getDeviceVersionInfo(currentDevice),
      new Promise((_, reject) => setTimeout(() => reject(new Error('API调用超时')), 3000))
    ])
    
    // 更新设备版本信息缓存
    if (versionInfo.code === 0 && versionInfo.data) {
      // 优化Map更新：只在数据变化时更新
      const currentVersion = deviceVersionInfo.value.get(currentDevice.id)
      const needUpdate = !currentVersion || 
                       currentVersion.currentVersion !== versionInfo.data.currentVersion ||
                       currentVersion.latestVersion !== versionInfo.data.latestVersion
      
      if (needUpdate) {
        // 直接更新现有Map，避免替换整个对象导致的重新渲染
        deviceVersionInfo.value.set(currentDevice.id, {
          currentVersion: versionInfo.data.currentVersion,
          latestVersion: versionInfo.data.latestVersion
        })
        // console.log(`设备 ${currentDevice.ip} 版本信息已更新:`, versionInfo.data)
      } else {
        // console.log(`设备 ${currentDevice.ip} 版本信息未变化，跳过更新`)
      }
    }
  } catch (error) {
    // console.error(`处理设备 ${currentDevice?.ip} 版本检查失败:`, error)

    // 标记当前处理的设备为离线
    if (currentDevice) {
      // 更新设备状态为离线
      // 更新设备最后更新时间，确保设备列表立即重新过滤
      devicesLastUpdateTime.value.set(currentDevice.id, Date.now())
      // console.log(`设备 ${currentDevice.ip} 版本检查失败，已标记为离线`)
    }
  } finally {
    isProcessingQueue.value = false
  }
}

// 批量处理版本检查队列 - 支持并发
const batchProcessVersionCheckQueue = async () => {
  // 最多同时处理MAX_CONCURRENT_CHECKS个设备
  const tasks = []
  for (let i = 0; i < MAX_CONCURRENT_CHECKS; i++) {
    tasks.push(processVersionCheckQueue())
  }
  await Promise.all(tasks)
}

// 初始化版本检查队列定时器
const initVersionCheckQueue = () => {
  // 清理之前的定时器
  if (versionCheckInterval.value) {
    clearInterval(versionCheckInterval.value)
  }
  
  // 每500ms处理一次队列，支持并发
  versionCheckInterval.value = setInterval(() => {
    batchProcessVersionCheckQueue()
  }, CHECK_INTERVAL)
  
  // console.log('版本检查队列定时器已启动，每500ms检查一次，支持最大并发数:', MAX_CONCURRENT_CHECKS)
}

// 添加设备到绑定状态查询队列
const addToBindCheckQueue = () => {
  const now = Date.now()
  const lastTime = lastBindCheckTime.value.get('global') || 0
  
  // 避免短时间内重复查询
  if (now - lastTime < MIN_BIND_CHECK_INTERVAL) {
    console.log('最近已查询过设备绑定状态，跳过本次检查')
    return
  }
  
  // 检查是否已在队列中
  if (deviceBindCheckQueue.value.length > 0) {
    console.log('设备绑定状态查询已在队列中，跳过重复添加')
    return
  }
  
  // 添加到队列
  deviceBindCheckQueue.value.push('bind-check')
  // console.log('设备绑定状态查询已添加到队列')
}

// 处理绑定状态查询队列
const processBindCheckQueue = async () => {
  if (isProcessingBindQueue.value) {
    return
  }
  
  isProcessingBindQueue.value = true
  
  try {
    // 从队列中取出任务
    const task = deviceBindCheckQueue.value.shift()
    
    if (!task) {
      isProcessingBindQueue.value = false
      return
    }
    
    console.log('开始查询设备绑定状态')
    
    // 记录检查时间
    lastBindCheckTime.value.set('global', Date.now())
    
    // 调用设备绑定状态查询API
    await fetchDeviceBindStatus()
    
    // console.log('设备绑定状态查询完成')
  } catch (error) {
    console.error('设备绑定状态查询失败:', error)
  } finally {
    isProcessingBindQueue.value = false
  }
}

// 批量处理绑定状态查询队列 - 支持并发
const batchProcessBindCheckQueue = async () => {
  // 最多同时处理MAX_CONCURRENT_BIND_CHECKS个任务
  const tasks = []
  for (let i = 0; i < MAX_CONCURRENT_BIND_CHECKS; i++) {
    tasks.push(processBindCheckQueue())
  }
  await Promise.all(tasks)
}

// 初始化绑定状态查询队列定时器
const initBindCheckQueue = () => {
  // 清理之前的定时器
  if (deviceBindCheckInterval.value) {
    clearInterval(deviceBindCheckInterval.value)
  }
  
  // 每1秒处理一次队列，支持并发
  deviceBindCheckInterval.value = setInterval(() => {
    batchProcessBindCheckQueue()
  }, BIND_CHECK_INTERVAL)
  
  // console.log('设备绑定状态查询队列定时器已启动，每1秒检查一次，支持最大并发数:', MAX_CONCURRENT_BIND_CHECKS)
}

// 自动获取所有设备版本信息 - 优化版
let versionCheckTimer = null

const autoGetAllDeviceVersions = async () => {
  if (versionCheckTimer) {
    clearTimeout(versionCheckTimer)
  }
  
  versionCheckTimer = setTimeout(async () => {
    try {
      console.log('自动获取所有设备版本信息 - 优化版')
      
      const devicesToCheck = []
      const now = Date.now()
      
      for (const device of devices.value) {
        const lastTime = lastCheckTime.value.get(device.id) || 0
        if (now - lastTime >= MIN_CHECK_INTERVAL) {
          devicesToCheck.push(device)
          
          if (device.version === 'v3') {
            fetchV3DeviceInfo(device)
          }
          
          if (token.value) {
            addToBindCheckQueue()
          }
        }
      }
      
      for (const device of devicesToCheck) {
        addToVersionCheckQueue(device)
      }
      
      // console.log(`本次自动检查共添加 ${devicesToCheck.length} 个设备到队列`)
    } catch (error) {
      console.error('自动获取所有设备版本信息失败:', error)
    }
  }, 500)
}

// 启动批量升级

// 发现并加载设备

// 后台10个一批加载云机列表

// 发现并加载设备
const discoverAndLoadDevices = async () => {
  loading.value = true
  try {
    // 记录开始发现设备的时间
    const discoveryTime = Date.now()
    
    // 发现设备
    const discoveredDevices = await discoverDevices()
    
    // 以设备ID为比对，更新现有设备的IP信息
    const discoveredDevicesMap = new Map(discoveredDevices.map(device => [device.id, device]))
    const discoveredDeviceIds = new Set(discoveredDevicesMap.keys())
    
    // ⚠️ 完全禁用状态更新，避免与心跳检测系统冲突
    // 心跳检测系统会通过 TCP Ping + HTTP 验证来管理设备状态
    // mDNS 发现功能仅用于手动添加设备时的设备列表展示
    // console.log('[发现设备] mDNS 扫描完成，发现', discoveredDevices.length, '个设备')
    // console.log('[发现设备] 设备状态由心跳检测系统统一管理，不在此处更新')
    
    // 不修改任何设备信息，不修改任何状态
    
    // 设备列表保持不变，只更新现有设备的状态
    
    // 不自动设置默认设备和获取容器列表，保持selectedCloudDevice为空，显示12个空坑位
    activeDevice.value = null
    selectedCloudDevice.value = null
    instances.value = []
    updateCloudMachines()
    
    // 设备列表更新后，重新初始化云机分组
    initCloudMachineGroups()
    
    // 3. 加载完成后，检查哪些设备没有被更新
    // 注意：这里不再简单地将未更新的设备标记为离线，
    // 而是保留它们的状态，因为有些设备可能能访问Docker API但在本次加载中失败
    const updatedDeviceIds = new Set()
    devicesLastUpdateTime.value.forEach((time, deviceId) => {
      if (time >= discoveryTime) {
        updatedDeviceIds.add(deviceId)
      }
    })
    
    // 4. 对于本次发现且成功加载的设备，确保标记为在线
    // 对于本次发现但未成功加载的设备，保持原有状态，不自动标记为离线
    // 这样可以避免能访问Docker API的设备被误判为离线
  } catch (error) {
    console.error('发现设备失败:', error)
    // 失败时保持设备列表不变
  } finally {
    loading.value = false
  }
  
  // 云机列表加载由 initDeviceHeartbeat 首次完成状态更新后触发，确保只请求在线设备
  
  // 获取设备绑定状态
  if(token.value) {
    fetchDeviceBindStatus()
  }
}


// 刷新当前列表 - 手动触发API版本检查和存储查询
const refchDevices = async () => {
  loading.value = true
  
  try {
    // console.log('[手动刷新] 开始手动刷新设备API版本和存储信息')
    
    // 获取所有在线设备
    const onlineDevices = devices.value.filter(device => {
      return devicesStatusCache.value.get(device.id) === 'online'
    })
    
    if (onlineDevices.length === 0) {
      ElMessage.warning('没有在线设备可以刷新')
      loading.value = false
      return
    }
    
    console.log(`[手动刷新] 找到 ${onlineDevices.length} 个在线设备`)
    
    // 🔧 调用后端强制刷新API版本和存储信息(不管在线离线)
    const deviceIPs = onlineDevices.map(device => device.ip)
    
    // 调用后端接口强制刷新
    await ForceRefreshDeviceInfo(deviceIPs)
    
    ElMessage.success(`已触发 ${onlineDevices.length} 个在线设备的API版本和存储信息刷新`)
    
    // console.log('[手动刷新] ✅ 刷新请求已发送到后端,预计1-2秒内完成')
    
  } catch (error) {
    console.error('[手动刷新] 刷新失败:', error)
    ElMessage.error(`刷新失败: ${error.message}`)
  } finally {
    loading.value = false
  }
}

// 切换镜像分类
const switchImageCategory = async (category) => {
  selectedImageCategory.value = category
  console.log('切换镜像分类:', category)
  
  if (category === 'local') {
    await fetchLocalCachedImages()
    
    // 如果是V3设备，确保已经获取了型号列表
    if (activeDevice.value && activeDevice.value.version === 'v3') {
      await getV3PhoneModels(activeDevice.value.ip)
    }
  } else if (category === 'online') {
    // 在线镜像分类，确保已经获取了在线镜像列表
    if (activeDevice.value) {
      const deviceType = activeDevice.value.name || 'C1'
      await fetchImageList(deviceType)
    } else {
      await fetchImageList('')
    }
    
    // 确保categorizeOnlineImages已经执行完毕，onlineImagesByModel有数据
    await nextTick()
    
    // 如果categorizeOnlineImages没有自动选中，这里作为兜底逻辑
    if (!currentOnlineImageModel.value && onlineImagesByModel.value.size > 0) {
      if (onlineImagesByModel.value.has('Q1')) {
        currentOnlineImageModel.value = 'Q1'
        console.log('[兜底] 默认选中在线镜像型号: Q1')
      } else {
        const firstModel = Array.from(onlineImagesByModel.value.keys())[0]
        currentOnlineImageModel.value = firstModel
        console.log('[兜底] 默认选中第一个在线镜像型号:', firstModel)
      }
    }
  } else if (category === 'device') {
     console.log('切换到设备镜像分类')
     // 如果没有选中的设备，且设备列表不为空，默认选中第一个
     if (!selectedDeviceForImages.value && devices.value.length > 0) {
         handleDeviceSelectForImages(devices.value[0])
     }
  }
}

// 关闭设备详情弹窗
const collapseRightSidebar = () => {
  console.log('关闭设备详情弹窗');
  deviceDetailsDialogVisible.value = false;
  // 退出查看详情模式，显示勾选框
  isViewingDeviceDetails.value = false;

  // 🔧 已移除定时器，无需清理

}



// ========== 设备心跳检测服务 ==========

// 初始化设备心跳检测
const initDeviceHeartbeat = async () => {
  console.log('[心跳] initDeviceHeartbeat 被调用, heartbeatInitialized=', heartbeatInitialized)
  // 防止重复初始化
  if (heartbeatInitialized) {
    console.log('[心跳] ⚠️ 心跳检测已初始化，忽略重复调用')
    return
  }
  
  heartbeatInitialized = true
  
  try {
    console.log('[心跳] ========== 开始初始化设备心跳检测服务 ==========');
    console.log('[心跳] 当前设备列表:', devices.value);
    
    // 🔥 应用启动时强制重置所有设备为离线状态
    console.log('[心跳] 正在重置所有设备为离线状态...');
    await ResetAllDevicesOffline();
    console.log('[心跳] ✓ 已重置所有设备为离线状态');
    
    // 收集所有设备IP
    const deviceIPs = devices.value.map(device => device.ip);
    
    if (deviceIPs.length === 0) {
      console.warn('[心跳] ⚠️ 当前设备列表为空，但仍启动心跳服务（设备添加后会自动监控）');
    }
    
    console.log('[心跳] ✓ 准备监控的设备IP列表:', deviceIPs);
    
    // 检查函数是否存在
    if (typeof updateMonitoredDevices !== 'function') {
      console.error('[心跳] ❌ updateMonitoredDevices 函数不存在！');
      return;
    }
    if (typeof startDeviceHeartbeat !== 'function') {
      console.error('[心跳] ❌ startDeviceHeartbeat 函数不存在！');
      return;
    }
    if (typeof getDevicesStatus !== 'function') {
      console.error('[心跳] ❌ getDevicesStatus 函数不存在！');
      return;
    }
    
    console.log('[心跳] ✓ 所有必需函数都存在');
    
    // 获取设备密码映射
    let devicePasswords = {};
    try {
      const storedPasswords = localStorage.getItem('devicePasswords');
      if (storedPasswords) {
        devicePasswords = JSON.parse(storedPasswords);
        console.log('[心跳] ✓ 已加载设备密码映射:', Object.keys(devicePasswords).length, '个设备');
      }
    } catch (error) {
      console.warn('[心跳] ⚠️ 加载设备密码失败:', error);
    }
    
    // 更新后端设备密码映射
    if (typeof UpdateDevicePasswords === 'function') {
      console.log('[心跳] 正在调用 UpdateDevicePasswords...');
      await UpdateDevicePasswords(devicePasswords);
      console.log('[心跳] ✓ 已更新后端设备密码映射');
    } else {
      console.warn('[心跳] ⚠️ UpdateDevicePasswords 函数不存在，跳过密码更新');
    }
    
    // 更新后端监控设备列表（携带名称映射）
    const deviceNamesMap = {}
    devices.value.forEach(d => { deviceNamesMap[d.ip] = d.name || d.ip })
    console.log('[心跳] 正在调用 updateMonitoredDevices...');
    await updateMonitoredDevices(deviceIPs, deviceNamesMap);
    console.log('[心跳] ✓ 已更新后端监控设备列表');
    
    // 启动后端心跳检测服务
    console.log('[心跳] 正在调用 startDeviceHeartbeat...');
    await startDeviceHeartbeat();
    console.log('[心跳] ✓ 已启动后端心跳检测服务');
    
    // 🔧 启动前端状态轮询（每1秒查询一次，实时更新设备信息）
    deviceHeartbeatTimer = setInterval(async () => {
      // console.log('[心跳] ⏰ 定时器触发，开始轮询设备状态...');
      await fetchDevicesStatusFromBackend();
    }, 1000); // 1秒，实时响应设备状态变化
    
    console.log('[心跳] ✓ 前端轮询定时器已启动');

    // 🚀 启动安卓容器列表缓存轮询（每2秒查询版本号，有变化才拉完整数据）
    androidCacheTimer = setInterval(fetchAndroidCacheIfUpdated, 2000)
    // console.log('[安卓轮询] ✓ 前端安卓缓存轮询定时器已启动（间隔2秒）');
    
    // 立即执行一次状态更新，拿到真实在线/离线状态后再加载云机列表
    // console.log('[心跳] 立即执行第一次状态查询...');
    await fetchDevicesStatusFromBackend();

    // 首次心跳完成后，devicesStatusCache 已是真实状态，安全地加载在线设备云机列表
    loadContainersInBatches()
    
    // console.log('[心跳] ========== ✅ 设备心跳检测服务启动成功！ ==========');
    // console.log('[心跳] 监控设备数:', deviceIPs.length);
    // console.log('[心跳] 后端检测间隔: 4秒');
    // console.log('[心跳] 前端轮询间隔: 1秒');
    // console.log('[心跳] 超时时间: 2秒');
  } catch (error) {
    console.error('[心跳] ❌ 初始化设备心跳检测失败:');
    console.error('[心跳] 错误详情:', error);
    console.error('[心跳] 错误堆栈:', error.stack);
  }
}

// ========== 安卓容器列表缓存轮询 ==========
// 每2秒调用，先比较版本号，有变化才拉完整数据，避免无效传输
const fetchAndroidCacheIfUpdated = async () => {
  try {
    const versions = await getAndroidCacheVersions()
    if (!versions || Object.keys(versions).length === 0) return

    const changedIps = []
    for (const [ip, version] of Object.entries(versions)) {
      if (version !== androidCacheVersions.value[ip]) {
        changedIps.push(ip)
      }
    }

    if (changedIps.length === 0) return

    // 拉取有变化的设备完整数据
    const cacheData = await getAndroidContainersList(changedIps)
    if (!cacheData) return

    for (const [ip, cacheEntry] of Object.entries(cacheData)) {
      // 更新本地版本号快照
      androidCacheVersions.value[ip] = versions[ip]

      // 找到对应设备
      const device = devices.value.find(d => d.ip === ip)
      if (!device) continue

      // 离线或错误状态：清空该设备缓存
      if (cacheEntry.status === 'offline') {
        deviceCloudMachinesCache.value.set(ip, [])
        if (activeDevice.value && activeDevice.value.ip === ip) {
          instances.value = []
          allInstances.value = []
          updateCloudMachines()
        }
        continue
      }

      // 无数据或认证失败：保留当前缓存不清空
      if (cacheEntry.status === 'error' && !cacheEntry.list) continue
      if (cacheEntry.status === 'auth_fail') continue

      // 有数据：解析并更新 deviceCloudMachinesCache（复用 _parseV3RawContainers）
      const raw = cacheEntry.list
      if (!raw) continue

      // 兼容后端返回的完整响应结构 {code, data:{list:[...]}}
      let rawContainers = []
      if (raw.data && raw.data.list) {
        rawContainers = raw.data.list
      } else if (raw.list) {
        rawContainers = raw.list
      }

      if (!Array.isArray(rawContainers)) continue

      _parseV3RawContainers(device, rawContainers)
    }
  } catch (e) {
    // 静默失败，不影响其他功能
    console.warn('[安卓轮询] 缓存更新异常:', e)
  }
}

// 从后端获取设备状态并更新前端缓存
const fetchDevicesStatusFromBackend = async () => {
  try {
    const statusMap = await getDevicesStatus();
    
    if (!statusMap || Object.keys(statusMap).length === 0) {
      console.warn('[心跳] ⚠️ 后端返回的状态为空');
      return;
    }
    console.log('[心跳] 后端返回状态，设备数:', Object.keys(statusMap).length, statusMap);
    
    // console.log('[心跳] ✓ 收到设备状态更新:');
    // console.log('[心跳] 状态数据样例:', Object.entries(statusMap)[0]); // 打印第一个设备的完整数据
    
    // 更新前端状态缓存
    let updatedCount = 0;
    let changedCount = 0;
    
    for (const [ip, statusInfo] of Object.entries(statusMap)) {
      const device = devices.value.find(d => d.ip === ip);
      
      if (!device) {
        console.warn(`[心跳] ⚠️ 找不到IP为 ${ip} 的设备`);
        continue;
      }
      
      const newStatus = statusInfo.status;
      const oldStatus = devicesStatusCache.value.get(device.id);
      
      // console.log(`[心跳] 处理设备 ${ip} (ID: ${device.id}):`, {
      //   status: statusInfo.status,
      //   apiVersion: statusInfo.apiVersion,
      //   storageTotal: statusInfo.storageTotal,
      //   oldStatus: oldStatus
      // });
      
      // 更新状态缓存
      devicesStatusCache.value.set(device.id, newStatus);
      devicesLastUpdateTime.value.set(device.id, Date.now());
      updatedCount++;
      
      // 如果设备离线，清除版本和存储信息，显示"未知"
      if (newStatus === 'offline') {
        // 清除 API 版本信息
        deviceVersionInfo.value.delete(device.id);
        // 清除存储信息
        deviceFirmwareInfo.value.delete(device.id);
        // console.log(`[心跳] 🔒 设备 ${ip} 离线，已清除缓存数据`);
      } else {
        // 设备在线，更新数据
        
        // 如果有 API 版本信息，更新到 deviceVersionInfo
        if (statusInfo.apiVersion && statusInfo.apiVersion !== '') {
          const currentVersionInfo = deviceVersionInfo.value.get(device.id) || {};
          
          // 🔧 智能更新策略: 只有当新版本号更大时才更新,避免覆盖升级后的最新版本
          const currentApiVersion = parseInt(currentVersionInfo.currentVersion || '0');
          const newApiVersion = parseInt(statusInfo.apiVersion || '0');
          
          // 如果新版本号更大或相等,或者没有现有版本信息,则更新
          if (newApiVersion >= currentApiVersion || !currentVersionInfo.currentVersion) {
            deviceVersionInfo.value.set(device.id, {
              ...currentVersionInfo,
              currentVersion: statusInfo.apiVersion,
              latestVersion: statusInfo.latestVersion || currentVersionInfo.latestVersion,
              lastUpdateTime: Date.now()
            });
            
            if (newApiVersion > currentApiVersion) {
              // console.log(`[心跳] 📈 设备 ${ip} API版本更新: ${currentVersionInfo.currentVersion} -> ${statusInfo.apiVersion}`);
            }
          } else {
            // console.log(`[心跳] ⏸️ 设备 ${ip} API版本未更新(新=${statusInfo.apiVersion}, 当前=${currentVersionInfo.currentVersion})`);
          }
        }
        
        // 🔧 更新 deviceFirmwareInfo (包括 responseTime)
        const currentFirmwareInfo = deviceFirmwareInfo.value.get(device.id) || {};
        
        // 🔧 缓存上次的延迟值: 只有新延迟>0时才更新,否则保留旧值
        const newResponseTime = statusInfo.responseTime && statusInfo.responseTime > 0 
          ? statusInfo.responseTime 
          : (currentFirmwareInfo.responseTime || 0);
        
        const updatedFirmwareInfo = {
          ...currentFirmwareInfo,
          responseTime: newResponseTime  // TCP Ping延迟(毫秒) - 缓存上次有效值
        };
        
        // 如果有存储信息，更新完整设备信息
        if (statusInfo.storageTotal && statusInfo.storageTotal > 0) {
          updatedFirmwareInfo.sdkVersion = statusInfo.sdkVersion || currentFirmwareInfo.sdkVersion;  // SDK版本
          updatedFirmwareInfo.deviceModel = statusInfo.deviceModel || currentFirmwareInfo.deviceModel;  // 设备型号
          updatedFirmwareInfo.originalData = {
              // 保留旧数据
              ...(currentFirmwareInfo.originalData || {}),
              // 更新存储信息（MB）
              mmctotal: statusInfo.storageTotal || 0,
              mmcuse: statusInfo.storageUsed || 0,
              mmcfree: statusInfo.storageFree || 0,
              // 更新 CPU 信息
              cputemp: statusInfo.cpuTemp || 0,
              cpuload: statusInfo.cpuLoad || '0%',
              // 更新内存信息（MB）
              memtotal: statusInfo.memoryTotal || 0,
              memuse: statusInfo.memoryUsed || 0,
              // 更新网络信息
              speed: statusInfo.speed || '0',
              network4g: statusInfo.network4g || 'n',
              netWork_eth0: statusInfo.networkEth0 || 'n',
              // 更新硬盘信息
              mmcread: statusInfo.mmcRead || '0',
              mmcwrite: statusInfo.mmcWrite || '0',
              mmcmodel: statusInfo.mmcModel || '',
              mmctemp: statusInfo.mmcTemp || '0',
              // 更新系统运行时间
              sysuptime: statusInfo.sysUptime || '0',
              // 更新设备基本信息
              model: statusInfo.deviceModel || (currentFirmwareInfo.originalData?.model || ''),
              version: statusInfo.sdkVersion || (currentFirmwareInfo.originalData?.version || ''),
              ip: statusInfo.ip || ip,
              ip_1: statusInfo.ip_1 || '',
              hwaddr: statusInfo.hwaddr || '',
              hwaddr_1: statusInfo.hwaddr_1 || '',
              deviceId: statusInfo.deviceId || ''
            };
          updatedFirmwareInfo.lastUpdateTime = Date.now();
          
          // console.log(`[心跳] 💾 设备 ${ip} 完整信息更新: SDK=${statusInfo.sdkVersion}, CPU=${statusInfo.cpuTemp}°C/${statusInfo.cpuLoad}, 内存=${statusInfo.memoryUsed}/${statusInfo.memoryTotal}MB, 磁盘=${statusInfo.mmcModel}/${statusInfo.mmcTemp}°C, 网速=${statusInfo.speed}, 延迟=${statusInfo.responseTime}ms`);
        }
        
        // 🔧 更新 deviceFirmwareInfo (始终更新，包括只有responseTime的情况)
        deviceFirmwareInfo.value.set(device.id, updatedFirmwareInfo);
      }
      
      // 如果状态发生变化，输出日志
      if (oldStatus !== newStatus) {
        const ts = new Date().toLocaleTimeString()
        changedCount++;
        
        // ✅ 如果设备从离线变为在线（含首次判定在线：oldStatus === undefined），触发后端立即刷新安卓缓存
        if ((oldStatus === 'offline' || oldStatus === undefined) && newStatus === 'online') {
          console.log(`[心跳][${ts}] ✅ 设备上线  IP: ${ip}  名称: ${device.name || '-'}  ID: ${device.id}  延迟: ${statusInfo.responseTime ?? '-'}ms  API: ${statusInfo.apiVersion || '-'}`)
          triggerAndroidRefresh([ip]).catch(err => {
            console.error(`[心跳][${ts}] ❌ 触发设备 ${ip} 安卓缓存刷新失败:`, err);
          });
        }
        
        // ✅ 如果设备从在线变为离线，且是坑位模式当前选中设备，清空安卓列表
        if (oldStatus === 'online' && newStatus === 'offline') {
          console.log(`[心跳][${ts}] ❌ 设备离线  IP: ${ip}  名称: ${device.name || '-'}  ID: ${device.id}  上次延迟: ${statusInfo.responseTime ?? '-'}ms`)
          if (cloudManageMode.value === 'slot' && selectedCloudDevice.value?.ip === ip) {
            instances.value = []
            allInstances.value = []
            console.log(`[心跳][${ts}] 🗑️ 坑位模式：设备 ${ip} 离线，已清空安卓列表`)
          }
        }
        
        // 其他状态变化（如首次判定离线：undefined -> offline）
        if (!((oldStatus === 'offline' || oldStatus === undefined) && newStatus === 'online') &&
            !(oldStatus === 'online' && newStatus === 'offline')) {
          console.log(`[心跳][${ts}] 🔄 设备状态变化  IP: ${ip}  名称: ${device.name || '-'}  ${oldStatus ?? '首次'} -> ${newStatus}`)
        }
      }
    }
    
    // console.log(`[心跳] 更新完成: ${updatedCount}个设备已更新, ${changedCount}个状态发生变化`);
    // console.log('[心跳] 当前状态缓存:', Array.from(devicesStatusCache.value.entries()));
  } catch (error) {
    console.error('[心跳] ❌ 获取设备状态失败:');
    console.error('[心跳] 错误详情:', error);
    console.error('[心跳] 错误堆栈:', error.stack);
  }
}

// 更新监控设备列表（当设备列表变化时调用）
const updateHeartbeatDevices = async () => {
  try {
    const deviceIPs = devices.value.map(device => device.ip);
    const deviceNamesMap = {}
    devices.value.forEach(d => { deviceNamesMap[d.ip] = d.name || d.ip })
    await updateMonitoredDevices(deviceIPs, deviceNamesMap);
    // console.log('[心跳] 已更新监控设备列表');
  } catch (error) {
    // console.error('[心跳] 更新监控设备列表失败:', error);
  }
}

// ⚠️ 暂时禁用自动 watch，避免意外触发
// 改为在添加/删除设备时手动调用 updateHeartbeatDevices()
// 
// 监听设备列表变化,自动更新监控列表
// 只监听设备IP列表,避免在设备属性更新时重复触发
// const deviceIPsList = computed(() => {
//   const ipList = devices.value.map(d => d.ip).sort().join(',')
//   console.log('[心跳Watch] computed deviceIPsList 被计算:', ipList)
//   return ipList
// })

// watch(deviceIPsList, async (newIPs, oldIPs) => {
//   console.log('[心跳Watch] watch 触发!', {
//     heartbeatInitialized,
//     newIPs,
//     oldIPs,
//     same: newIPs === oldIPs
//   })
//   
//   if (heartbeatInitialized && newIPs && newIPs !== oldIPs) {
//     console.log('[心跳Watch] ⚠️⚠️⚠️ 设备IP列表发生变化,更新监控列表')
//     console.log('[心跳Watch] 旧IP列表:', oldIPs || '空')
//     console.log('[心跳Watch] 新IP列表:', newIPs)
//     await updateHeartbeatDevices()
//   } else {
//     console.log('[心跳Watch] 条件不满足,不触发更新')
//   }
// })

// 将后端缓存的原始容器列表解析为前端 deviceCloudMachinesCache 格式
// rawContainers: /android 接口返回的 list 数组（已含 deviceIp 字段）
const _parseV3RawContainers = (device, rawContainers) => {
  // 保存所有原始容器数据，用于备份列表
  const allRawContainers = rawContainers.map(container => ({
    ...container,
    status: container.status === 'running' ? 'running' : container.status,
    deviceIp: device.ip
  }))

  // 按坑位分组，每个坑位只保留一个容器，优先保留 running 状态
  const containersBySlot = new Map()
  rawContainers.forEach(container => {
    const slotNum = container.indexNum
    if (slotNum) {
      const existing = containersBySlot.get(slotNum)
      const processed = {
        ...container,
        containerId: container.id || container.ID || container.containerId || container.containerID || '',
        status: container.status === 'running' ? 'running' : 'shutdown',
        deviceIp: device.ip
      }
      if (!existing || (processed.status === 'running' && existing.status !== 'running')) {
        containersBySlot.set(slotNum, processed)
      }
    }
  })

  const processedContainers = Array.from(containersBySlot.values())

  // 获取之前的缓存，用于保留截图状态
  const previousCache = deviceCloudMachinesCache.value.get(device.ip) || []

  const deviceCloudMachines = processedContainers.map(inst => {
    const screenshotUrl = getCloudMachineScreenshotUrl(device, inst)
    const newMachine = {
      id: `${device.ip}_${inst.name}`,
      containerId: inst.containerId || inst.containerID || inst.id || inst.ID || '',
      name: inst.name,
      status: inst.status,
      screenshot: screenshotUrl,
      screenshotData: null,
      screenshotError: false,
      hasLoadedOnce: false,
      ip: inst.ip,
      modelName: inst.modelName,
      modelPath: inst.modelPath,
      deviceIp: device.ip,
      indexNum: inst.indexNum,
      portBindings: inst.portBindings,
      image: inst.image,
      created: inst.created,
      dns: inst.dns,
      androidType: inst.androidType,
      networkName: inst.networkName,
      randomFile: inst.randomFile || false,
      enforce: inst.enforce !== undefined ? inst.enforce : true,
      mgenable: inst.mgenable || '0',
      gmsenable: inst.gmsenable || '0',
      doboxWidth: inst.doboxWidth || '',
      doboxHeight: inst.doboxHeight || '',
      doboxDpi: inst.doboxDpi || '',
      doboxFps: inst.doboxFps || '',
      macVlanIp: inst.networkName === 'myt' ? (inst.ip || '') : '',
      mytBridgeName: inst.networkName !== 'myt' ? (inst.networkName || '') : '',
      adbPort: inst.adbPort || 5555, // ADB端口，默认5555
    }
    return mergeCloudMachineState(previousCache, newMachine)
  })

  devicesLastUpdateTime.value.set(device.id, Date.now())
  deviceCloudMachinesCache.value.set(device.ip, deviceCloudMachines)

  if (activeDevice.value && activeDevice.value.ip === device.ip) {
    instances.value = processedContainers
    allInstances.value = allRawContainers
    updateCloudMachines()
  } else if (cloudManageMode.value === 'batch') {
    // 批量模式下 activeDevice 为 null，需主动重新计算分组数据以刷新界面
    initCloudMachineGroups()
  }
}

// 获取安卓云机列表
// isUserInitiated=true：触发后端立即刷新后读最新缓存（操作后强制更新）
// isUserInitiated=false：直接读后端缓存（日常展示，极快）
const fetchAndroidContainers = async (device, isUserInitiated = false) => {
  if (!device) return

  // 只有用户主动刷新或设备未加载过时才显示加载状态
  if (isUserInitiated || !deviceCloudMachinesCache.value.has(device.ip)) {
    cloudMachineLoadingState.value.set(device.ip, true)
  }

  try {
    // isUserInitiated=true：触发后端立即刷新，等 800ms 后读最新数据
    if (isUserInitiated) {
      triggerAndroidRefresh([device.ip]).catch(() => {})
      await new Promise(resolve => setTimeout(resolve, 800))
    }

    // 从后端缓存读取（IPC 调用，极快，无外部网络请求）
    const cacheMap = await getAndroidContainersList([device.ip])
    const deviceCache = cacheMap[device.ip]

    // 认证失败：触发 authRetry 弹窗重新认证，认证后触发后端重新轮询
    if (deviceCache && deviceCache.status === 'auth_fail') {
      await authRetry(device, async () => {
        triggerAndroidRefresh([device.ip]).catch(() => {})
      })
      return
    }

    // 从缓存中提取原始列表
    let rawList = null
    if (deviceCache && deviceCache.list) {
      const resp = deviceCache.list
      // 后端存的是完整 /android 响应：{code:0, data:{list:[...]}}
      if (resp.code === 0 && resp.data && resp.data.list) {
        rawList = resp.data.list
      } else if (resp.code === 0 && resp.list) {
        rawList = resp.list
      } else if (resp.code === 0) {
        rawList = [] // 空列表合法
      }
    }

    // 缓存无数据（设备刚上线、尚未完成首次轮询）：回退到直接请求
    if (rawList === null) {
      await authRetry(device, async (password = null) => {
        const containers = await getContainers(device, password)
        if (!containers) return
        if (containers.code === 61 && containers.message === 'Authentication Failed') {
          throw new Error('Authentication Failed')
        }
        if (containers.code === 0) {
          const list = (containers.data && containers.data.list) ? containers.data.list
            : containers.list ? containers.list : []
          _parseV3RawContainers(device, list)
        }
      })
      return
    }

    // V3：直接解析
    if (device.version === 'v3') {
      _parseV3RawContainers(device, rawList)
      return
    }

    // V0-V2：回退到 Docker API（后端轮询仅支持 V3）
    await authRetry(device, async (password = null) => {
      const dockerContainers = await getContainers(device, password)
      if (!dockerContainers) return

      const dockerArr = Array.isArray(dockerContainers)
        ? dockerContainers
        : (dockerContainers?.code === 0 && dockerContainers?.data?.list) ? dockerContainers.data.list : []

      const processedContainers = dockerArr.filter(c => {
        const image = c.Image || c.image
        const name = c.Name || c.Names?.[0]
        return !(image?.includes('myt_sdk') || image?.includes('myt_vpc_plugin') ||
                 name?.includes('myt_sdk') || name?.includes('myt_vpc_plugin'))
      }).filter(c => parseContainerSlot(c, device) !== null)
        .map((c, index) => {
          const slotNum = parseContainerSlot(c, device) || (index + 1)
          return {
            name: c.Names ? c.Names[0]?.replace('/', '') || `container-${index}` : `container-${index}`,
            status: c.Status?.startsWith('Up') ? 'running' : 'shutdown',
            indexNum: slotNum,
            ip: `${device.ip.split('.')[0]}.${device.ip.split('.')[1]}.3.${slotNum + 100}`,
            deviceIp: device.ip,
            containerId: c.Id || c.ID || c.id || '',
            image: c.Image,
            createTime: new Date(c.Created * 1000).toLocaleString('zh-CN', {
              year: 'numeric', month: '2-digit', day: '2-digit',
              hour: '2-digit', minute: '2-digit', second: '2-digit'
            }),
            modelName: 'Q1',
            Ports: c.Ports,
            NetworkSettings: c.NetworkSettings,
            PortBindings: c.PortBindings,
            portBindings: c.portBindings,
            rawContainer: c
          }
        })

      const previousCache = deviceCloudMachinesCache.value.get(device.ip) || []
      const deviceCloudMachines = processedContainers.map(inst => {
        const screenshotUrl = getCloudMachineScreenshotUrl(device, inst)
        return mergeCloudMachineState(previousCache, {
          id: `${device.ip}_${inst.name}`,
          name: inst.name,
          status: inst.status,
          screenshot: screenshotUrl,
          screenshotData: null,
          screenshotError: false,
          hasLoadedOnce: false,
          ip: inst.ip,
          modelName: inst.modelName,
          modelPath: inst.modelPath,
          deviceIp: device.ip,
          indexNum: inst.indexNum,
          PortBindings: inst.PortBindings,
          portBindings: inst.portBindings,
        })
      })

      deviceCloudMachinesCache.value.set(device.ip, deviceCloudMachines)
      if (activeDevice.value && activeDevice.value.ip === device.ip) {
        instances.value = processedContainers
        allInstances.value = processedContainers
        updateCloudMachines()
      }
    })
  } catch (error) {
    if (error.message !== 'Authentication Failed') {
      console.error('获取安卓云机列表失败:', device.ip, error)
    }
    deviceCloudMachinesCache.value.set(device.ip, [])
  } finally {
    cloudMachineLoadingState.value.set(device.ip, false)
  }
}


// 启动同步授权定时器（每10分钟执行一次）
const startSyncAuthTimer = () => {
  // 清理之前的定时器
  if (syncAuthTimer.value) {
    clearInterval(syncAuthTimer.value)
  }
  
  // 设置新的定时器，每10分钟执行一次
  syncAuthTimer.value = setInterval(() => {
    console.log('自动执行同步授权操作')
    handleSyncAuthorization()
  }, 30 * 60 * 1000) // 10分钟 = 10 * 60 * 1000毫秒
  
  console.log('同步授权定时器已启动，每10分钟执行一次')
}

// 从HostManagement.vue移回的函数
const handleSyncAuthorization = async () => {
  const currentToken = token.value 
  if (!currentToken) {
    ElMessage.warning('请先获取授权token');
    return;
  }
  
  const devices = filteredDevices.value;
  if (devices.length === 0) {
    ElMessage.warning('请先添加设备');
    return;
  }
  
  const deviceIPs = devices.map(device => device.ip);
  console.log('deviceIPs:', deviceIPs)
  
  try {
    let successCount = 0;
    for (const deviceIP of deviceIPs) {
      try {
        const result = await SyncAuthorization(currentToken, deviceIP);
        if (result.success) {
          successCount++;
        } else {
          console.error(`设备 ${deviceIP} 同步授权失败:`, result.message);
        }
      } catch (deviceError) {
        console.error(`设备 ${deviceIP} 同步授权出错:`, deviceError);
      }
    }
    
    if (successCount === deviceIPs.length) {
      ElMessage.success(t('common.syncAuthCompletedForAllDevices', { count: successCount }));
    } else if (successCount > 0) {
      ElMessage.warning(t('common.syncAuthPartialSuccess', { count: successCount }));
    } else {
      ElMessage.error(t('common.syncAuthAllFailed'));
    }
  } catch (error) {
    console.error('同步授权失败:', error);
    ElMessage.error(t('common.syncAuthFailed'));
  }
}

const clearCache = async () => {
  try {
    await ElMessageBox.confirm('确定要清理所有客户端数据并重新加载吗？这将清除所有设备信息、密码和镜像缓存。', '清理客户端数据确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 显示加载状态
    ElMessage({ message: '正在清理客户端数据...', type: 'info' })
    
    console.log('[清理缓存] 开始清理所有客户端数据...')
    
    // 1. 清理localStorage缓存
    console.log('[清理缓存] 清理 localStorage...')
    localStorage.removeItem('edgeclient_devices')  // ✅ 修复：正确的设备列表key
    localStorage.removeItem('devicePasswords')
    localStorage.removeItem('mytos_image_list')
    localStorage.removeItem('mytos_image_list_last_update')
    localStorage.removeItem('deviceGroups')  // 清理设备分组
    localStorage.removeItem('deviceGroupFilter')  // 清理分组过滤器
    
    console.log('[清理缓存] ✓ localStorage 清理完成')
    
    // 2. 清理内存缓存（Map类型）
    console.log('[清理缓存] 清理 Map 缓存...')
    deviceCloudMachinesCache.value.clear()
    cloudMachineLoadingState.value.clear()
    devicesLastUpdateTime.value.clear()
    devicesStatusCache.value.clear()
    deviceVersionInfo.value.clear()
    deviceFirmwareInfo.value.clear()  // 清理固件信息
    onlineImagesByModel.value.clear()
    imageDownloadStatus.value.clear()
    imageUploadStatus.value.clear()
    deviceBindStatus.value.clear()  // 清理绑定状态
    
    console.log('[清理缓存] ✓ Map 缓存清理完成')
    
    // 3. 清理内存缓存（ref类型）
    console.log('[清理缓存] 清理 ref 缓存...')
    devices.value = []
    activeDevice.value = null
    instances.value = []
    allInstances.value = []
    cloudMachines.value = []
    selectedCloudMachines.value = []
    phoneModels.value = []
    imageList.value = []
    filteredImageList.value = []
    sharedFiles.value = []
    selectedFiles.value = []
    dockerNetworks.value = []
    localCachedImages.value = []
    deviceBoxImages.value = []
    isViewingDeviceDetails.value = false
    deviceDetailsDialogVisible.value = false
    
    console.log('[清理缓存] ✓ ref 缓存清理完成')


    // 🔧 通知后端清空监控设备列表
    console.log('[清理缓存] 通知后端清空监控列表...')
    await updateMonitoredDevices([], {})
    console.log('[清理缓存] ✓ 后端监控列表已清空')

    // 🔧 通知后端清空设备密码（否则心跳检测仍会用旧密码认证，导致设备被判定为在线）
    console.log('[清理缓存] 通知后端清空设备密码...')
    await UpdateDevicePasswords({})
    console.log('[清理缓存] ✓ 后端设备密码已清空')

    // 🔧 已移除定时器，无需清理
    
    // 4. 清理api.js中的内存缓存
    console.log('[清理缓存] 清理 API 缓存...')
    containersMemoryCache.clear()
    
    console.log('[清理缓存] ✓ API 缓存清理完成')
    
    // 5. 重新加载所有数据
    console.log('[清理缓存] 重新加载数据...')
    await refreshData()
    
    // 6. 自动重新加载镜像列表
    console.log('[清理缓存] 重新加载镜像列表...')
    await fetchImageList('')
    
    console.log('[清理缓存] ========== ✅ 客户端数据清理完成 ==========')
    ElMessage({ message: '客户端数据清理完成，数据已重新加载', type: 'success' })
  } catch (error) {
    if (error !== 'cancel') {
      console.error('[清理缓存] ❌ 清理失败:', error)
      ElMessage.error('清理客户端数据失败: ' + error.message)
    }
  }
}

// 批量清理磁盘数据
const handleBatchDeleteHosts = async () => {
  if (selectedHostDevices.value.length === 0) {
    ElMessage.warning('请先选择要清理的主机')
    return
  }

  try {
    await ElMessageBox.confirm(
        `确定要清理选中的 ${selectedHostDevices.value.length} 个设备的磁盘数据吗？`,
        '清理磁盘数据会重启设备，重启时间为5~10分钟，请耐心等待',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        }
    )

    collapseRightSidebar()

    // 批量创建任务并执行
    for (const device of selectedHostDevices.value) {
        const taskId = `cleanDisk_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        const task = {
            id: taskId,
            type: 'cleanDisk',
            deviceIP: device.ip,
            status: 'running',
            progress: 0,
            startTime: new Date(),
            completed: 0,
            failed: 0,
            totalSteps: 6,
            currentStep: 0,
            steps: []
        }
        taskQueue.value.unshift(task)

        // 异步执行清理，不阻塞循环
        // 🔧 不使用 .catch()，让错误在 cleanDeviceDisk 内部处理
        // 避免认证等待时被误判为失败
        cleanDeviceDisk(device, taskId)
    }

    ElMessage.success(`已开始清理 ${selectedHostDevices.value.length} 个设备，请查看任务列表`)
    selectedHostDevices.value = []

  } catch (error) {
    if (error !== 'cancel') {
        console.error('批量清理失败:', error)
        ElMessage.error('批量清理操作失败')
    }
  }
}

// 单个设备清理逻辑
const cleanDeviceDisk = async (device, taskId) => {
    try {
        // 🔧 使用 authRetry 处理认证
        await authRetry(device, async (password) => {
            let headers = {}
            if (password) {
                const auth = btoa(`admin:${password}`)
                headers['Authorization'] = `Basic ${auth}`
            }

            const response = await fetch(
                `http://${getDeviceAddr(device.ip)}/server/device/reset`,
                {
                    method: 'POST',
                    headers: headers
                }
            )

            if (!response.ok) {
                // 🔧 如果是401错误，直接抛出让authRetry处理，不标记任务失败
                if (response.status === 401) {
                    throw new Error('Authentication Failed')
                }
                
                // 其他错误才标记任务失败
                const currentTask = taskQueue.value.find(t => t.id === taskId)
                if (currentTask) {
                    currentTask.status = 'failed'
                    currentTask.endTime = new Date()
                    currentTask.error = `接口请求失败: ${response.status}`
                }
                
                return
            }

            const reader = response.body.getReader()
            const decoder = new TextDecoder()
            let buffer = ''
            let taskCompleted = false

            while (!taskCompleted) {
                const { done, value } = await reader.read()
                if (done) break

                buffer += decoder.decode(value, { stream: true })
                const lines = buffer.split('\n')
                buffer = lines.pop()

                for (const line of lines) {
                    if (line.trim()) {
                        // console.log(`[${device.ip}] SSE:`, line)
                        
                        const currentTask = taskQueue.value.find(t => t.id === taskId)
                        if (!currentTask || taskCompleted) continue

                        const stepMatch = line.match(/\[STEP\s+(\d+)\]/i)
                        const infoMatch = line.match(/\[INFO\]/)
                        const resetMatch = line.match(/Reset sequence completed/i)
                        const rebootMatch = line.match(/Rebooting/i)

                        if (stepMatch) {
                            currentTask.currentStep = parseInt(stepMatch[1])
                            currentTask.progress = Math.round((currentTask.currentStep / currentTask.totalSteps) * 100)
                            currentTask.steps.push(line)
                        } else if (infoMatch) {
                            currentTask.steps.push(line)
                        }

                        if (resetMatch || rebootMatch) {
                            currentTask.progress = 100
                            currentTask.status = 'completed'
                            currentTask.completed = 1
                            currentTask.endTime = new Date()
                            taskCompleted = true
                            
                            if (activeDevice.value && activeDevice.value.ip === device.ip) {
                                 isViewingDeviceDetails.value = false
                            }
                        }
                    }
                }
            }
            
            // 🔧 检查流结束时任务状态
            const finalTask = taskQueue.value.find(t => t.id === taskId)
            if (finalTask && !taskCompleted) {
                // 流已结束但没有收到完成标记，可能是最后一行数据在buffer中
                if (buffer.trim()) {
                    console.log(`[${device.ip}] SSE最后一行:`, buffer)
                    const resetMatch = buffer.match(/Reset sequence completed/i)
                    const rebootMatch = buffer.match(/Rebooting/i)
                    
                    if (resetMatch || rebootMatch) {
                        finalTask.progress = 100
                        finalTask.status = 'completed'
                        finalTask.completed = 1
                        finalTask.endTime = new Date()
                        taskCompleted = true
                        
                        if (activeDevice.value && activeDevice.value.ip === device.ip) {
                            isViewingDeviceDetails.value = false
                        }
                    }
                }
                
                // 如果仍未完成，检查是否至少接收到了步骤信息
                if (!taskCompleted && finalTask.currentStep >= finalTask.totalSteps) {
                    console.log(`[${device.ip}] 所有步骤已完成，标记任务为成功`)
                    finalTask.progress = 100
                    finalTask.status = 'completed'
                    finalTask.completed = 1
                    finalTask.endTime = new Date()
                    
                    if (activeDevice.value && activeDevice.value.ip === device.ip) {
                        isViewingDeviceDetails.value = false
                    }
                }
            }
        })
    } catch (error) {
        // 🔧 在这里处理错误，更新任务状态
        console.error(`[${device.ip}] 清理异常:`, error)
        const currentTask = taskQueue.value.find(t => t.id === taskId)
        if (currentTask && currentTask.status !== 'completed') {
            currentTask.status = 'failed'
            currentTask.endTime = new Date()
            currentTask.error = `执行异常: ${error.message || '未知错误'}`
        }
    }
}


// 退出登录
const logOut = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('uname')
  // localStorage.removeItem('syncAuthCredentials')
  token.value = null
  uname.value = null
  window.location.reload()
}

const handleBatchDeleteDevices = async () => {
  if (selectedHostDevices.value.length === 0) {
    ElMessage.warning('请先选择要删除的设备')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedHostDevices.value.length} 个设备吗？`, '批量删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger'
    })
    
    // 批量移除设备
    const deviceIpsToDelete = new Set(selectedHostDevices.value.map(d => d.ip))
    devices.value = devices.value.filter(d => !deviceIpsToDelete.has(d.ip))
    
    // 保存到本地存储
    saveDevicesToLocalStorage()
    
    // 清除设备相关的缓存
    deviceIpsToDelete.forEach(ip => {
      deviceCloudMachinesCache.value.delete(ip)
      // 如果删除的是当前激活的设备，清空激活状态
      if (activeDevice.value && activeDevice.value.ip === ip) {
        activeDevice.value = null
        instances.value = []
        allInstances.value = []
        cloudMachines.value = []
      }
    })

    // 清理已删除设备的认证密码
    const passwords = JSON.parse(localStorage.getItem('devicePasswords') || '{}')
    deviceIpsToDelete.forEach(ip => {
      delete passwords[ip]
    })
    localStorage.setItem('devicePasswords', JSON.stringify(passwords))
    await UpdateDevicePasswords(passwords)
    
    // 清空选择
    selectedHostDevices.value = []

    // 通知后端更新监控设备列表，停止对已删除设备的心跳检测
    await updateHeartbeatDevices()
    
    ElMessage.success(`成功删除 ${deviceIpsToDelete.size} 个设备`)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除设备失败:', error)
      ElMessage.error('批量删除设备失败')
    }
  }
}


// 绑定主机
const handleBindsTest = async () => { 
  try {
    // 1. 检查是否已登录
    // const currentToken = token.value 
    // if (!currentToken) {
    //   ElMessage.warning('请先登录获取授权token');
    //   return;
    // }
    
    // 2. 获取选中的设备
    const selectedDevices = [...selectedHostDevices.value];
    // if (selectedDevices.length === 0) {
    //   ElMessage.warning('请先选择要绑定的设备');
    //   return;
    // }
    
    // 3. 过滤出未绑定的设备
    const unboundDevices = selectedDevices.filter(device => {
      const bindStatus = deviceBindStatus.value.get(device.id) || 0;
      return bindStatus !== 1; // 1表示已绑定
    });
    
    if (unboundDevices.length === 0) {
      ElMessage.warning('所有选中的设备都已绑定');
      return;
    }
    
    // 4. 提取设备ID
    const deviceIds = unboundDevices.map(device => device.id);
    console.log('要绑定的设备ID:', deviceIds);
    
    
    // 6. 调用绑定API
    const formData = new URLSearchParams()
    formData.append('type', 'user_host_oper');
    const data = {
       act: 'batchBind',
       data: JSON.stringify({ host: deviceIds }),
       token: token.value
    }
    formData.append('data', JSON.stringify(data));


    console.log('formData:', formData)
    
    const response = await fetch('https://www.moyunteng.com/api/api.php', {
          method: 'POST',
          body: formData
     })  
 
    const result = await response.json()
    console.log('response:', result)
    
    // 7. 处理响应结果
    if (result.code == 200) {
      // 更新绑定状态
      result.data.forEach(deviceId => {
        deviceBindStatus.value.set(deviceId, 1); // 标记为已绑定
      });
      
      // 清空选中设备
      selectedHostDevices.value = [];
      
      ElMessage.success(`成功绑定 ${result.data.length} 个设备`);
    } else {
      ElMessage.error(`绑定失败: ${result.msg}`);
    }
    
  } catch (error) {
    console.error('绑定设备失败:', error);
    ElMessage.error('绑定设备失败: ' + error.message);
  }
}
</script>

<template>
  <el-container class="app-container" style="height: 100vh;">
    <!-- 终端遮罩层 -->
    <div id="terminal-overlay" class="terminal-overlay"></div>
    <!-- 主要内容区域 -->
    <el-main class="app-main" style="height: 100%;">
      <!-- 悬浮任务列表按钮和语言切换 -->
      <div style="position: fixed; top: 18px; right: 34px; z-index: 1000; display: flex; gap: 8px; align-items: center;">
        <!-- 语言切换器 -->
        <LanguageSwitcher />
        
        <!-- 任务队列 -->
        <el-dropdown trigger="click" popper-class="task-list-dropdown">
          <el-button type="primary" size="default" class="el-tooltip__trigger" style="padding: 8px 10px;">
            <el-icon><Timer /></el-icon>
            {{ t('common.taskQueue') }}
            <el-badge v-if="runningTasksCount > 0" :value="runningTasksCount" type="danger" :max="99"></el-badge>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu style="width: 400px; max-height: 500px; overflow-y: auto;">
              <div style="padding: 12px; font-weight: bold; border-bottom: 1px solid #e4e7ed; color: #303133; display: flex; justify-content: space-between; align-items: center;">
                <span>{{ t('common.taskQueue') }}</span>
                <el-button type="link" size="small" @click="handleClearTasks" :disabled="taskQueue.length === 0">
                  {{ t('common.delete') }}{{ t('common.taskQueue') }}
                </el-button>
              </div>
              <div v-if="taskQueue.length === 0" style="padding: 20px; text-align: center; color: #909399;">
                {{ t('common.taskQueue') }}为空
              </div>
              <el-dropdown-item v-else v-for="task in taskQueue" :key="task.id" style="padding: 0; border-bottom: 1px solid #f0f0f0; text-align: left !important; display: block;">
                <div style="padding: 10px 15px; text-align: left !important; display: block; width: 100%;">
                  <div style="display: flex; justify-content: space-between; align-items: center; text-align: left;">
                    <div style="text-align: left !important;">
                      <span style="font-weight: bold;">
                {{ task.type === 'restart' ? '批量重启' : task.type === 'reset' ? '批量重置' : task.type === 'shutdown' ? '批量关机' : task.type === 'create' ? '批量创建' : task.type === 'delete' ? '批量删除' : task.type === 'switchModel' ? (task.operation === 'new' ? '批量新机' : '批量切换机型') : task.type === 'uploadFile' ? '批量上传' : task.type === 'uploadImage' ? '批量上传镜像' : task.type === 'downloadImage' ? '下载镜像' : task.type === 'cleanDisk' ? '清理磁盘' : task.type === 'updateImage' ? '批量更新镜像' : task.type === 'copy' ? '复制云机' : '未知任务' }}
              </span>
                      <el-tooltip :content="getTaskTargetDisplay(task).full.join('\n')" placement="top" :disabled="getTaskTargetDisplay(task).full.length <= 1">
                <span style="color: #409eff; font-size: 12px; margin-left: 4px; cursor: default;">
                  {{ getTaskTargetDisplay(task).short }}
                </span>
              </el-tooltip>
                      <el-tag :type="task.status === 'running' ? 'warning' : task.status === 'completed' ? 'success' : task.status === 'failed' ? 'danger' : 'info'" size="small" style="margin-left: 8px;">
                        {{ task.status === 'running' ? '运行中' : task.status === 'completed' ? '已完成' : task.status === 'failed' ? '失败' : task.status === 'canceled' ? '已取消' : '等待中' }}
                      </el-tag>
                    </div>
                    <div style="font-size: 12px; color: #909399; text-align: right !important;">{{ new Date(task.startTime || 0).toLocaleTimeString() }}</div>
                  </div>
                  
                  <!-- 批量上传镜像/文件按设备IP分进度条显示 -->
                  <div v-if="(task.type === 'uploadImage' || task.type === 'uploadFile') && task.deviceIps && task.deviceIps.length > 0" class="device-progress-list" style="margin-top: 8px;">
                    <div v-for="deviceIP in task.deviceIps" :key="deviceIP" class="device-progress-item" style="display: flex; align-items: center; margin-bottom: 4px;">
                      <span style="width: 100px; font-size: 12px; color: #606266; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ deviceIP }}</span>
                      <el-progress 
                        :percentage="getDeviceProgress(task, deviceIP)" 
                        :status="getDeviceProgressStatus(task, deviceIP)"
                        :stroke-width="8"
                        style="flex: 1;"
                      ></el-progress>
                      <component :is="getDeviceProgressText(task, deviceIP).icon" v-if="getDeviceProgressText(task, deviceIP).icon === 'CircleCheck'" style="width: 16px; height: 16px; color: #67c23a; margin-left: 8px;" />
                      <component :is="getDeviceProgressText(task, deviceIP).icon" v-else-if="getDeviceProgressText(task, deviceIP).icon === 'CircleClose'" style="width: 16px; height: 16px; color: #f56c6c; margin-left: 8px;" />
                      <component :is="getDeviceProgressText(task, deviceIP).icon" v-else-if="getDeviceProgressText(task, deviceIP).icon === 'Loading'" style="width: 16px; height: 16px; color: #409eff; margin-left: 8px;" class="rotating" />
                      <el-icon v-else style="width: 16px; height: 16px; color: #909399; margin-left: 8px;"><Timer /></el-icon>
                    </div>
                  </div>
                  <!-- 其他任务显示总进度 -->
                  <div v-else style="margin-top: 5px;">
                    <el-progress :percentage="task.progress" :status="task.status === 'completed' ? 'success' : task.status === 'failed' ? 'exception' : ''" :stroke-width="10"></el-progress>
                  </div>
                  
                  <!-- 清理磁盘任务显示步骤 -->
                  <div v-if="task.type === 'cleanDisk' && task.steps && task.steps.length > 0" style="margin-top: 8px;">
                    <div style="font-size: 12px; color: #606266; margin-bottom: 4px;">执行步骤:</div>
                    <div style="max-height: 150px; overflow-y: auto; background: #f5f7fa; padding: 8px; border-radius: 4px; font-size: 11px; color: #606266; font-family: monospace;">
                      <div v-for="(step, index) in task.steps" :key="index" style="margin-bottom: 2px; word-break: break-all;">{{ step }}</div>
                    </div>
                  </div>
                  
                  <!-- 任务摘要信息 -->
                  <div v-if="task.type === 'cleanDisk'" style="margin-top: 5px; font-size: 12px; text-align: left !important; display: block;">
                    <div v-if="task.status === 'completed'" style="color: #67c23a;">
                      设备清理完毕，正在重启
                    </div>
                    <div v-else style="color: #606266;">
                      设备: {{ task.deviceIP }} | 步骤: {{ task.currentStep }}/{{ task.totalSteps }}
                    </div>
                  </div>
                  <!-- 创建任务分步显示进度 -->
                  <div v-else-if="task.type === 'create'" style="margin-top: 5px; font-size: 12px; text-align: left !important; display: block;">
                    <template v-if="task.currentStep === 'image' && task.imageProgress !== null">
                      <span style="color: #409eff;">推送镜像: {{ task.imageProgress }}%</span>
                    </template>
                    <template v-else-if="task.currentStep === 'create'">
                      <span style="color: #67c23a;">创建第{{ task.completed + task.failed }}个 ({{ task.completed + task.failed }}/{{ task.total }})</span>
                    </template>
                    <template v-else>
                      <span style="color: #606266;">创建第{{ task.completed + task.failed }}个 ({{ task.completed + task.failed }}/{{ task.total }})</span>
                    </template>
                    <span style="color: #909399; margin-left: 8px;">成功: {{ task.completed }}, 失败: {{ task.failed }}</span>
                  </div>
                  <div v-else style="margin-top: 5px; font-size: 12px; color: #606266; text-align: left !important; display: block;">
                      总进度: {{ task.progress }}% (成功: {{ task.completed }}, 失败: {{ task.failed }})
                    </div>
                    <div v-if="(task.type === 'uploadImage' || task.type === 'downloadImage') && task.imageName" style="margin-top: 3px; font-size: 12px; color: #67c23a; text-align: left !important; display: block;">
                      镜像名称: {{ task.imageName }}
                    </div>
                    <div v-if="task.type === 'uploadFile'" style="margin-top: 3px; font-size: 12px; color: #67c23a; text-align: left !important; display: block;">
                      文件数: {{ task.fileCount || task.targets?.length || 0 }}, 云机数: {{ task.machineCount || 0 }}
                    </div>
                    <!-- 复制云机：流式日志列表 -->
                    <div v-if="task.type === 'copy' && task.copyLogs && task.copyLogs.length > 0" style="margin-top: 6px; max-height: 160px; overflow-y: auto; background: #f5f7fa; border-radius: 4px; padding: 6px 8px;">
                      <div v-for="(log, li) in task.copyLogs" :key="li" style="display:flex; align-items:center; gap:4px; font-size:11px; margin-bottom:3px; font-family:monospace;">
                        <span style="color:#909399; flex-shrink:0;">[{{ log.current }}/{{ log.total }}]</span>
                        <span style="color:#303133; flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;" :title="log.name">{{ log.name }}</span>
                        <el-tag size="small" :type="log.status === 'success' ? 'success' : log.status === 'failed' ? 'danger' : 'info'" style="flex-shrink:0; padding:0 4px; height:16px; line-height:16px;">
                          {{ log.status === 'success' ? '✓' : log.status === 'failed' ? '✗' : '…' }}
                        </el-tag>
                        <span style="color:#606266; max-width:120px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; flex-shrink:0;" :title="log.message">{{ log.message }}</span>
                      </div>
                    </div>
                    <div v-if="task.failed > 0" style="margin-top: 3px; font-size: 12px; color: #f56c6c; text-align: left !important; display: block;">
                      <template v-if="task.type === 'create'">
                        <div style="text-align: left !important; display: block;">失败原因:</div>
                        <div v-for="(failedTarget, index) in task.failedTargets.slice(0, 3)" :key="index" style="margin-top: 2px; text-align: left !important; display: block;">
                          坑位{{ failedTarget.slot }}: {{ failedTarget.error || '未知错误' }}
                        </div>
                        <div v-if="task.failedTargets.length > 3" style="margin-top: 2px; color: #909399; text-align: left !important; display: block;">
                          ...等{{ task.failedTargets.length }}个失败记录
                        </div>
                      </template>
                      <template v-else-if="task.type === 'uploadFile'">
                        <div style="text-align: left !important; display: block;">失败原因:</div>
                        <div v-for="(failedTarget, index) in task.failedTargets.slice(0, 3)" :key="index" style="margin-top: 2px; text-align: left !important; display: block;">
                          {{ failedTarget.filePath?.split('\\').pop() || '未知文件' }} -> {{ formatInstanceName(failedTarget.machineName) || '未知云机' }}: {{ failedTarget.error || '未知错误' }}
                        </div>
                        <div v-if="task.failedTargets.length > 3" style="margin-top: 2px; color: #909399; text-align: left !important; display: block;">
                          ...等{{ task.failedTargets.length }}个失败记录
                        </div>
                      </template>
                      <template v-else>
                        <div style="text-align: left !important; display: block;">
                          <span v-for="(failedTarget, index) in task.failedTargets.slice(0, 3)" :key="index" style="display: block; margin-bottom: 2px;">
                            {{ failedTarget.machineName ? `${failedTarget.deviceIP || '未知设备'} ${failedTarget.machineName}` : (failedTarget.deviceIP || '未知设备') }}: {{ failedTarget.error || '未知错误' }}
                          </span>
                          <div v-if="task.failedTargets.length > 3" style="margin-top: 2px; color: #909399;">
                            ...等{{ task.failedTargets.length }}个失败记录
                          </div>
                        </div>
                      </template>
                    </div>
                  <div style="margin-top: 8px; display: flex; gap: 5px; justify-content: flex-start; text-align: left !important;">
                    <el-button v-if="task.status === 'running'" type="danger" size="small" @click="cancelTask(task.id)">取消</el-button>
                    <el-button v-if="task.status === 'failed'" type="primary" size="small" @click="retryFailedTask(task.id)">重试失败</el-button>
                  </div>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        
        <!-- 设置按钮 -->
        <el-tooltip content="设置" placement="bottom">
          <el-button type="default" size="default" style="padding: 8px 10px;" @click="openSettingsDialog">
            <el-icon><Setting /></el-icon>
          </el-button>
        </el-tooltip>
        
        <!-- 更新菜单组件 -->
        <UpdateMenu @show-update-dialog="handleShowUpdateDialog" />
      </div>
      
      <!-- 标签页 -->
      <el-tabs v-model="activeTab" size="large" class="modern-tabs fixed-tabs" @tab-change="handleTabChange" style="height: 100%;" :scrollable="true">
        <template #append>
          <!-- 其他菜单按钮 -->
          <el-dropdown trigger="click">
            <el-button type="link" size="large" class="menu-button">
              <el-icon><More /></el-icon>
            </el-button>
            <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item>{{ t('common.settings') }}</el-dropdown-item>
            <el-dropdown-item divided>{{ t('menu.logout') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
          </el-dropdown>
        </template>
          <!-- 主机管理 -->
          <el-tab-pane :label="t('menu.hostManagement')" name="host-management" style="height: 100%;">
            <HostManagement 
              :devices="devices"
              :active-device="activeDevice"
              :instances="instances"
              :loading="loading"
              :token="token"
              :uname="uname"
              :filtered-devices="filteredDevices"
              :device-firmware-info="deviceFirmwareInfo"
              :device-version-info="deviceVersionInfo"
              :selected-host-devices="selectedHostDevices"
              :is-viewing-device-details="isViewingDeviceDetails"
              :active-tab="activeTab"
              :grouped-instances="groupedInstances"
              :filtered-image-List="filteredImageList"
              :docker-networks="dockerNetworks"
              :box-images="boxImages"
              :device-filter="deviceFilter"
              :devices-status-cache="devicesStatusCache"
              :device-bind-status="deviceBindStatus"
              :device-groups="deviceGroups"
              :device-group-filter="deviceGroupFilter"
              :filtered-devices-by-group="filteredDevicesByGroup"
              :device-groups-tree="deviceGroupsTree"
              :is-batch-projection-controlling="isBatchProjectionControlling"
              @update:device-filter="deviceFilter = $event"
              @update:device-group-filter="deviceGroupFilter = $event"
              @refch-devices="refchDevices"
              @handle-batch-delete-devices="handleBatchDeleteDevices"
              @show-sync-auth-dialog="showSyncAuthDialog"
              @handle-sync-authorization="handleSyncAuthorization"
              @clear-cache="clearCache"
              @log-Out="logOut"
              @handle-host-device-selection-change="handleHostDeviceSelectionChange"
              @show-device-details="showDeviceDetails"
              @show-create-dialog="showCreateDialog"
              @collapse-right-sidebar="collapseRightSidebar"
              @fetch-android-containers="fetchAndroidContainers"
              @handle-batch-action="handleBatchAction"
              @refresh-image-list="refreshImageList"
              @fetch-docker-networks="fetchDockerNetworks"
              @fetch-v3-device-info="fetchV3DeviceInfo"
              @show-password-dialog="showPasswordDialog"
              @handle-close-password="handleClosePassword"
              @show-add-macvlan-dialog="showAddMacvlanDialog"
              @switch-image-category="switchImageCategory" 
              @start-container="startContainer"
              @stop-container="stopContainer"
              @start-projection="startProjection"
              @show-update-image-dialog="showUpdateImageDialog"
              @handle-delete-container="handleDeleteContainerWithBackupRefresh"
              @handle-add-device="handleAddDevice"
              @handle-batch-add-devices="handleBatchAddDevices"
              @handle-binds-test="handleBindsTest"
              @handle-get-device-version="handleGetDeviceVersion"
              @add-device-group="addDeviceGroup"
              @rename-device-group="renameDeviceGroup"
              @delete-device-group="deleteDeviceGroup"
              @move-device-to-group="moveDeviceToGroup"
              @handle-batch-delete-hosts="handleBatchDeleteHosts"
            />
          </el-tab-pane>
        <!-- 云机管理 -->
        <el-tab-pane :label="t('menu.cloudManagement')" name="cloud-management">
          <CloudManagement 
            :active-device="activeDevice"
            :cloud-manage-mode="cloudManageMode"
            :cloud-machine-groups="cloudMachineGroups"
            :selected-cloud-device="selectedCloudDevice"
            :imageList="imageList"
            :instances="instances"
            :all-instances="allInstances"
            :cloud-machines="cloudMachines"
            :cloud-machines-by-name="cloudMachinesByName"
            :selected-cloud-machines="selectedCloudMachines"
            :devices="devices"
            :devices-status-cache="devicesStatusCache"
            :loading="loading"
            :tree-selected-keys="treeSelectedKeys"
            :is-batch-projection-controlling="isBatchProjectionControlling"
            @show-create-dialog="showCreateDialog"
            @handle-context-menu="handleContextMenu"
            @show-backup-list="showBackupList"
            @handle-batch-action="handleBatchAction"
            @handle-delete-container="handleDeleteContainerWithBackupRefresh"
            @start-projection="startProjection"
            @handle-node-drop="handleNodeDrop"
            @cloud-manage-mode-change="handleCloudManageModeChange"
            @selected-cloud-device-change="handleSelectedCloudDeviceChange"
            @fetch-android-containers="fetchAndroidContainers"
            @device-added="handleAddDevice"
            @devices-added="handleBatchAddDevices"
            @handle-batch-add-devices="handleBatchAddDevices"
            @handle-tree-check="handleTreeCheck"
            @move-device-to-group="moveDeviceToGroup"
            @start-copy-task="handleStartCopyTask"
            :screenshot-cache="screenshotDataCache"
            
          />
          
          <!-- 批量上传对话框 -->
          <BatchUploadDialog
            v-model:visible="batchUploadDialogVisible"
            :selected-machines="batchUploadSelectedMachines"
            :cloud-manage-mode="cloudManageMode"
            :selected-cloud-device="selectedCloudDevice"
            @upload="handleBatchUpload"
            @open-shared-directory="openSharedDirectory"
          />
        </el-tab-pane>

        
        
        <!-- 镜像管理 -->
        <el-tab-pane :label="t('menu.imageManagement')" name="image-management">
          <el-row :gutter="12" style="height: 100%; min-height: 600px;">
            <el-col :span="24" class="device-left-col" style="height: 100%;">
              <el-card shadow="hover" class="device-card image-management-card" style="height: 100%;">
                <template #header>
                  <div class="card-header" style="display: flex; justify-content: space-between; align-items: center;">
                    <div class="device-list-tabs">
                      <el-button 
                        type="link" 
                        size="small" 
                        class="device-tab" 
                        :class="{ active: selectedImageCategory === 'online' }"
                        @click="selectedImageCategory = 'online'; switchImageCategory('online')"
                      >{{ $t('common.onlineImage') }}</el-button>
                      <el-button 
                        type="link" 
                        size="small" 
                        class="device-tab"
                        :class="{ active: selectedImageCategory === 'local' }"
                        @click="selectedImageCategory = 'local'; switchImageCategory('local')"
                      >{{ $t('common.localImage') }}</el-button>
                      <el-button 
                        type="link" 
                        size="small" 
                        class="device-tab"
                        :class="{ active: selectedImageCategory === 'device' }"
                        @click="selectedImageCategory = 'device'; switchImageCategory('device')"
                      >{{ $t('common.deviceImage') }}</el-button>
                      <el-button 
                        type="link" 
                        size="small" 
                        class="device-tab"
                        :class="{ active: selectedImageCategory === 'guide' }"
                        @click="selectedImageCategory = 'guide'"
                      >{{ $t('image.usageGuide') }}</el-button>
                    </div>
                    <div style="display: flex; gap: 8px;">
                      <el-button 
                        type="primary" 
                        size="small" 
                        @click="refreshImageList" 
                        :disabled="fetchingImages"
                        class="refresh-button"
                      >
                        <el-icon :class="{ 'is-rotating': fetchingImages }"><Refresh /></el-icon> {{ $t('common.refreshOnlineImages') }}
                      </el-button>
                      <el-button 
                        type="primary" 
                        size="small" 
                        @click="handleOpen" 
                        class="refresh-button"
                      >
                        <el-icon ><FolderOpened /></el-icon> {{ $t('common.openLocalImageDirectory') }}
                      </el-button>
                    </div>
                  </div>
                </template>
                
                <!-- 在线镜像列表 -->
                <div v-if="selectedImageCategory === 'online'" class="image-list-container">
                  <el-tabs v-model="currentOnlineImageModel" type="border-card" class="online-image-tabs">
                    <el-tab-pane 
                      v-for="[model, images] in onlineImagesByModel" 
                      :key="model" 
                      :label="model || $t('common.other')"
                      :name="model || $t('common.other')"
                    >
                      <!-- 筛选控制区域 -->
                      <div class="filter-controls" style="margin-bottom: 15px;">
                        <el-radio-group v-model="imageCategory" size="small" style="margin-bottom: 10px; display: block;">
                          <el-radio-button label="simulator">{{ $t('common.simulator') }}</el-radio-button>
                          <el-radio-button label="container">{{ $t('common.container') }}</el-radio-button>
                        </el-radio-group>

                        <div v-if="imageCategory === 'container'" class="version-filter">
                           <span style="margin-right: 10px; font-size: 12px;">{{ $t('common.androidVersion') }}:</span>
                           <el-radio-group v-model="containerAndroidVersion" size="small">
                             <el-radio :label="10">Android 10</el-radio>
                             <!-- Q1 支持 Android 12，P1 不支持 -->
                             <el-radio v-if="model === 'Q1'" :label="12">Android 12</el-radio>
                             <el-radio :label="14">Android 14</el-radio>
                           </el-radio-group>
                        </div>
                      </div>

                      <el-table 
                        :data="getFilteredImages(images, model)" 
                        stripe 
                        size="small" 
                        class="image-table"
                      >
                        <el-table-column :label="$t('common.imageName')" align="center" width="260">
                          <template #default="scope">
                            <span class="image-name">{{ scope.row.name }}</span>
                          </template>
                        </el-table-column>
                        <el-table-column :label="$t('common.updateContent')" align="center">
                          <template #default="scope">
                            <span v-html="scope.row.udesc ? scope.row.udesc.replace(/\n/g, '|') : ''"></span>
                          </template>
                        </el-table-column>
                        <el-table-column :label="$t('common.operation')" fixed="right" align="center" width="220">
                          <template #default="scope">
                            <div class="table-actions">
                              <!-- 未下载状态 -->
                              <el-button 
                                v-if="!imageDownloadStatus.get(scope.row.url)"
                                type="primary" 
                                size="small"
                                @click="downloadOnlineImage(scope.row)"
                                :disabled="isDownloadingImage"
                              >
                                <el-icon><Download /></el-icon> {{ $t('common.download') }}
                              </el-button>
                              <!-- 已下载状态 -->
                              <template v-else>
                                <el-button 
                                  type="success" 
                                  size="small"
                                  @click="uploadImageToDevice(scope.row)"
                                  
                                >
                                  <el-icon><Upload /></el-icon> {{ $t('common.uploadToDevice') }}
                                </el-button>
                                <el-button 
                                  type="danger" 
                                  size="small"
                                  @click="deleteDownloadedImage(scope.row)"
                                >
                                  <el-icon><Delete /></el-icon> {{ $t('common.delete') }}
                                </el-button>
                              </template>
                            </div>
                          </template>
                        </el-table-column>
                      </el-table>
                    </el-tab-pane>
                  </el-tabs>
                </div>
                
                <!-- 设备镜像列表 -->
                <div v-else-if="selectedImageCategory === 'device'" class="image-list-container" style="height: calc(100% - 50px);">
                  <el-row style="height: 100%;">
                    <!-- 左侧：设备列表 -->
                    <el-col :span="6" style="height: 100%; border-right: 1px solid #EBEEF5;">
                      <div class="device-list-sidebar" style="height: 100%; overflow-y: auto;">
                        <el-menu
                          :default-active="selectedDeviceForImages ? selectedDeviceForImages.ip : ''"
                          class="el-menu-vertical-demo"
                          style="border-right: none;"
                        >
                          <el-menu-item 
                            v-for="device in devices" 
                            :key="device.ip" 
                            :index="device.ip"
                            @click="handleDeviceSelectForImages(device)"
                          >
                            <el-icon><Monitor /></el-icon>
                            <span>{{ device.ip }}</span>
                            <el-tag size="small" type="info" style="margin-left: 5px;">{{ device.name || $t('common.device') }}</el-tag>
                          </el-menu-item>
                          
                          <div v-if="devices.length === 0" style="padding: 20px; text-align: center; color: #909399;">
                            {{ $t('common.noDevice') }}
                          </div>
                        </el-menu>
                      </div>
                    </el-col>
                    
                    <!-- 右侧：镜像列表 -->
                    <el-col :span="18" style="height: 100%; padding-left: 10px;">
                        <div v-if="selectedDeviceForImages" class="image-list-content" style="height: 100%; overflow-y: auto;">
                          <div v-if="isLoadingDeviceImages" class="image-loading" style="padding: 20px;">
                            <el-skeleton :rows="5" animated></el-skeleton>
                          </div>
                          <div v-else-if="matchedDeviceBoxImages.length === 0" class="no-images" style="padding: 20px; text-align: center;">
                            <el-empty :description="$t('common.noDownloadedImages')" :image-size="100"></el-empty>
                          </div>
                          <div v-else class="image-items" style="display: flex; flex-wrap: wrap; align-content: flex-start;">
                            <div v-for="(image, index) in matchedDeviceBoxImages" :key="index" class="image-item" style="margin: 10px;">
                              <el-card shadow="hover" class="image-card">
                                <template #header>
                                  <div class="card-header" style="display: flex; justify-content: space-between; align-items: center;">
                                    <span :title="image.onlineImageName">{{ image.onlineImageName }}</span>
                                    <el-button
                                      type="danger"
                                      size="small"
                                      @click="handleDeleteDeviceImage(image)"
                                      class="delete-button"
                                    >
                                    <el-icon><Delete /></el-icon> {{ $t('common.delete') }}
                                    </el-button>
                                  </div>
                                </template>
                                <div class="image-info">
                                  <div class="image-details" style="font-size: 13px; line-height: 1.5;">
                                    <p>{{ $t('common.onlineImageName') }}: {{ image.onlineImageName }}</p>
                                    <p>{{ $t('common.imageSize') }}: {{ image.size }}</p>
                                    <p>{{ $t('common.createTime') }}: {{ image.createTime }}</p>
                                    <p v-if="image.matched" style="color: #67c23a;">{{ $t('common.matchedWithOnlineImage') }}</p>
                                    <p v-else style="color: #909399;">{{ $t('common.notMatchedWithOnlineImage') }}</p>
                                    <p v-if="image.matched" style="font-size: 12px; color: #909399; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">{{ $t('common.onlineURL') }}: {{ image.onlineImageUrl }}</p>
                                    <p v-else style="font-size: 12px; color: #909399; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">{{ $t('common.deviceURL') }}: {{ image.url }}</p>
                                  </div>
                                </div>
                              </el-card>
                            </div>
                          </div>
                        </div>
                        
                        <div v-else class="no-device-selected" style="height: 100%; display: flex; justify-content: center; align-items: center;">
                          <el-empty :description="$t('common.pleaseSelectDeviceFirst')" :image-size="100"></el-empty>
                        </div>
                    </el-col>
                  </el-row>
                </div>
                <div v-else-if="selectedImageCategory === 'local'" class="image-list-container">
                  <el-table 
                    :data="localCachedImages" 
                    stripe 
                    size="small" 
                    class="image-table"
                  >
                    <el-table-column :label="$t('common.imageName')" align="center">
                      <template #default="scope">
                        <span class="image-name">{{ scope.row.name }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column :label="$t('common.filePath')" show-overflow-tooltip align="center">
                      <template #default="scope">
                        <span class="image-url">{{ scope.row.url }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column :label="$t('common.size')" align="center">
                      <template #default="scope">
                        <span>{{ formatFileSize(scope.row.size || 0) }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column :label="$t('common.createTime')" align="center">
                      <template #default="scope">
                        <span>{{ scope.row.createTime }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column :label="$t('common.availableDeviceModels')" show-overflow-tooltip align="center">
                      <template #default="scope">
                        <el-tag 
                          v-for="model in scope.row.availableModels" 
                          :key="model" 
                          size="small"
                          style="margin-right: 4px; margin-bottom: 4px;"
                        >
                          {{ model }}
                        </el-tag>
                        <span v-if="!scope.row.availableModels || scope.row.availableModels.length === 0">{{ $t('common.universal') }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column :label="$t('common.operation')" fixed="right" align="center" width="280">
                      <template #default="scope">
                        <div class="table-actions">
                          <el-button 
                            type="success" 
                            size="small" 
                            @click="uploadLocalImageToDevice(scope.row)" 
                            :disabled="false" 
                          >
                            <el-icon><Upload /></el-icon> {{ $t('common.uploadToDevice') }}
                          </el-button>
                          <el-button 
                            type="danger" 
                            size="small"
                            @click="deleteLocalCachedImage(scope.row)"
                          >
                            <el-icon><Delete /></el-icon> {{ $t('common.delete') }}
                          </el-button>
                        </div>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>

                <!-- 使用说明 -->
                <div v-else-if="selectedImageCategory === 'guide'" class="image-list-container image-guide-container">
                  <div class="image-guide-content" v-html="$t('image.guideContent')"></div>
                </div>

                <!-- 下载进度弹窗 -->
              
                
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>
        

        <!-- 机型管理 -->
        <el-tab-pane :label="t('menu.modelManagement')" name="model-management">
          <ModelManagement 
            ref="modelManagementRef"
            :devices="devices" 
            :activeDevice="activeDevice" 
            :selectedHostDevices="selectedHostDevices"
            :devicesStatusCache="devicesStatusCache"
          />
        </el-tab-pane>

        <!-- 网络管理 -->
        <el-tab-pane :label="t('menu.networkManagement')" name="network-management">
          <NetworkManagement 
            ref="networkManagementRef"
            :devices="devices" 
            :activeDevice="activeDevice" 
            :selectedHostDevices="selectedHostDevices"
            :token="token"
            :device-firmware-info="deviceFirmwareInfo"
            :device-version-info="deviceVersionInfo"
            :devices-status-cache="devicesStatusCache"
          />
        </el-tab-pane>


        <!-- 备份管理 -->
        <el-tab-pane :label="t('menu.backupManagement')" name="backup-management">
          <BackupManagement 
            ref="backupManagementRef"
            :devices="devices" 
            :device-firmware-info="deviceFirmwareInfo"
            :devices-status-cache="devicesStatusCache"
          />
        </el-tab-pane>


        <!-- 实例管理 -->
        <el-tab-pane :label="t('menu.instanceManagement')" name="instance-management">
          <instanceManagement 
            ref="instanceManagementRef"
            :devices="devices" 
            @handle-sync-authorization="handleSyncAuthorization"
            v-model:token="token"
            @update-user-info="handleUpdateUserInfo"
          />
        </el-tab-pane>

        <!-- 客服 -->
        <!-- <el-tab-pane name="customer-service">
          <template #label>
            <el-badge :value="customerServiceUnreadCount" :hidden="customerServiceUnreadCount === 0" :max="99">
              <span>客服</span>
            </el-badge>
          </template>
          <CustomerService 
             ref="customerServiceRef"
            :devices="devices" 
            :token="token"
            :device-firmware-info="deviceFirmwareInfo"
            :devices-status-cache="devicesStatusCache"
            :device-version-info="deviceVersionInfo"
            @unread-count-change="handleUnreadCountChange"
            @show-sync-auth-dialog="showSyncAuthDialog"
            @show-register-dialog="openRegisterDialog"
          />
        </el-tab-pane> -->

         <!-- 批量任务 -->
        <el-tab-pane :label="t('menu.batchTask')" name="batch-task">
          <BatchTaskManagement 
            :devices="devices"
            :instances="instances"
            :all-instances="allInstances"
            :cloud-machines="cloudMachines"
            :device-cloud-machines-cache="deviceCloudMachinesCache"
            :devices-status-cache="devicesStatusCache"
            :loading="loading"
          />
        </el-tab-pane>

        <!-- 流媒体 -->
        <el-tab-pane :label="t('menu.streamManagement')" name="stream-management">
          <StreamManagement 
            :devices="devices"
            :devices-status-cache="devicesStatusCache"
            :cloud-machine-groups="cloudMachineGroups"
            :fetch-android-containers="fetchAndroidContainers"
            :device-cloud-machines-cache="deviceCloudMachinesCache"
          />
        </el-tab-pane>

        <!-- AI助理 -->
        <el-tab-pane :label="t('menu.aiAssistant')" name="ai-assistant">
          <AiAssistant 
            ref="aiAssistantRef"
            :devices="devices" 
            :token="token"
            :devices-status-cache="devicesStatusCache"
          />
        </el-tab-pane>

        <!-- RPA Agent -->
        <el-tab-pane :label="t('menu.rpaAgent')" name="rpa-agent">
          <RpaAgent
            ref="rpaAgentRef"
            :devices="devices"
            :token="token"
            :devices-status-cache="devicesStatusCache"
          />
        </el-tab-pane>

        <!-- 互联云机 -->
        <el-tab-pane :label="t('menu.interconnectedCloudMachines')" name="interconnected-cloud-machines">
          <InterconnectedCloudMachines :token="token" />
        </el-tab-pane>

        <!-- opencecs -->
        <el-tab-pane label="OpenCecs" name="opencecs-management">
          <OpencecsManagement ref="opencecsManagementRef" />
        </el-tab-pane>

      </el-tabs>
    </el-main>
  </el-container>
  
  <!-- 设备选择对话框 -->
  <el-dialog
    v-model="showDeviceSelectionDialog"
    :title="t('dialog.addDeviceTitle')"
    width="600px"
    :close-on-click-modal="false"
    @close="handleDeviceSelectionDialogClose"
  >
    <!-- 上传进度已整合到任务列表 -->
    
    <div class="device-selection-container">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h4 style="margin: 0;">
          {{ $t('common.pleaseSelectDeviceForUpload') }}
          <span style="color: #409eff; font-weight: 500; margin-left: 8px;">
            ({{ $t('model.onlineDeviceCount', { count: sortedCompatibleDevicesList.length }) }})
          </span>
        </h4>
        <el-button 
          type="primary" 
          size="small" 
          :icon="Refresh" 
          @click="refreshDeviceListForUpload"
          :loading="refreshingDevicesForUpload"
        >
          刷新列表
        </el-button>
      </div>
      
      <!-- 空状态提示 -->
      <div v-if="sortedCompatibleDevicesList.length === 0" class="empty-devices">
        <el-empty :description="$t('common.noCompatibleDevices')" :image-size="100"></el-empty>
      </div>
      
      <!-- 设备列表 -->
      <el-table 
        v-else
        ref="deviceSelectionTableRef"
        :data="sortedCompatibleDevicesList" 
        stripe 
        size="small" 
        max-height="400"
        class="device-selection-table"
        @selection-change="handleUploadDeviceSelectionChange"
        :row-class-name="getDeviceRowClassName"
      >
        <el-table-column 
          type="selection" 
          width="55"
          :selectable="checkDeviceSelectable"
        ></el-table-column>
        <el-table-column prop="name" :label="$t('image.deviceModel')" width="120"></el-table-column>
        <el-table-column prop="ip" :label="$t('model.deviceIP')" width="150" sortable></el-table-column>
        <el-table-column :label="$t('common.status')" width="80" align="center">
          <template #default="scope">
            <el-tag 
              :type="devicesStatusCache.get(scope.row.id) === 'online' ? 'success' : 'danger'" 
              size="small"
            >
              {{ devicesStatusCache.get(scope.row.id) === 'online' ? $t('common.online') : $t('common.offline') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('image.availableSpace')" width="140" align="center">
          <template #default="scope">
            <span v-if="devicesStatusCache.get(scope.row.id) !== 'online'" style="color: #909399;">
              未知
            </span>
            <span v-else-if="getDeviceStorageInfo(scope.row.id)">
              {{ getDeviceStorageInfo(scope.row.id).freeText }}
              <span 
                v-if="getDeviceStorageInfo(scope.row.id).isLow" 
                style="color: #F56C6C; font-size: 12px;"
              >
                {{ $t('image.insufficient') }}
              </span>
            </span>
            <span v-else style="color: #909399;">{{ $t('common.loading') }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="showDeviceSelectionDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button 
          type="primary" 
          @click="handleUploadAfterDeviceSelection"
          :loading="isUploadingToMultipleDevices"
        >
          {{ t('common.upload') }}
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 云机重命名弹窗 -->
  <el-dialog
    v-model="renameDialogVisible"
    :title="t('cloudMachine.renameCloudMachine')"
    width="400px"
    :close-on-click-modal="false"
  >
    <el-form :model="renameForm" label-width="80px">
      <el-form-item :label="$t('cloudMachine.currentName')">
        <el-input v-model="renameForm.name" disabled></el-input>
      </el-form-item>
      <el-form-item :label="$t('cloudMachine.newName')">
        <el-input v-model="renameForm.newName" :placeholder="$t('cloudMachine.enterNewName')"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="renameDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitRename" :loading="renameLoading">
          {{ t('common.confirm') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
  
  <!-- 系统公告弹窗 -->
  <el-dialog
    v-model="announcementVisible"
    :title="announcementData.title"
    width="520px"
    :close-on-click-modal="false"
    :show-close="false"
    @close="closeAnnouncement"
    class="announcement-dialog"
  >
    <div class="announcement-content">
      <div class="announcement-icon">
        <el-icon :size="48" color="#409EFF">
          <BellFilled />
        </el-icon>
      </div>
      <div class="announcement-text">
        {{ announcementData.content }}
      </div>
    </div>
    <template #footer>
      <div class="announcement-footer">
        <el-button type="primary" @click="closeAnnouncement" size="large">
          {{ $t('common.understood') }}
          <span v-if="countdown > 0" class="countdown-badge">
            {{ countdown }}s
          </span>
        </el-button>
      </div>
    </template>
  </el-dialog>
  
  <!-- 切换备份悬浮窗口 -->
  <el-dialog
    v-model="backupListVisible"
    :title="t('cloudMachine.switchBackup')"
    width="70%"
  >
    <!-- 切换备份时的覆盖层 -->
    <div 
      v-if="backupLoading" 
      class="switching-backup-overlay-dialog"
    >
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>切换中...</span>
    </div>
    
    <!-- 坑位选择 -->
    <!-- <div class="backup-slot-section" style="margin-bottom: 16px;">
      <span class="backup-section-label">坑位：</span>
      <el-select 
        v-model="currentSlot" 
        placeholder="选择坑位" 
        style="width: 150px; margin-right: 12px;"
        @change="initBackupList"
        :disabled="backupLoading"
      >
        <el-option 
          v-for="i in 12" 
          :key="i" 
          :label="i" 
          :value="i"
        ></el-option>
      </el-select>
    </div> -->
    

    
    <!-- 备份列表 -->
    <div class="backup-list-container">
      <!-- 排序栏和批量操作 -->
      <div class="backup-sort-bar" style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <div style="display: flex; gap: 16px;">
          <span class="backup-section-label">{{ $t('common.sortBy') }}</span>
          <el-button 
            type="link" 
            size="small" 
            @click="changeSort('name')"
            :class="{ active: sortBy === 'name' }"
            :disabled="backupLoading"
          >
            {{ $t('common.name') }} {{ sortBy === 'name' ? (sortOrder === 'ascending' ? '↑' : '↓') : '' }}
          </el-button>
          <el-button 
            type="link" 
            size="small" 
            @click="changeSort('createTime')"
            :class="{ active: sortBy === 'createTime' }"
            :disabled="backupLoading"
          >
            {{ $t('common.createTimeSort') }} {{ sortBy === 'createTime' ? (sortOrder === 'ascending' ? '↑' : '↓') : '' }}
          </el-button>
        </div>
        <div>
          <span v-if="selectedBackupList.length > 0" style="margin-right: 12px; color: #909399;">{{ $t('common.selectedItems', { count: selectedBackupList.length }) }}</span>
          <el-button 
            type="danger" 
            size="small" 
            @click="batchDeleteBackup"
            :loading="backupLoading"
            :disabled="backupLoading || selectedBackupList.length === 0"
          >
            {{ $t('common.batchDelete') }}
          </el-button>
        </div>
      </div>
      
      <!-- 备份列表表格 -->
      <el-table 
        :data="sortedBackupList" 
        stripe 
        size="small" 
        style="width: 100%;" 
        :disabled="backupLoading"
        @selection-change="handleBackupSelectionChange"
        :row-key="row => row.id"
        ref="backupTableRef"
      >
        <el-table-column type="selection" width="50"></el-table-column>
        <el-table-column prop="name" :label="$t('cloudMachine.backupName')" width="150" show-overflow-tooltip>
           <template #default="scope">
            {{ formatInstanceName(scope.row.name) }}
          </template>
        </el-table-column>
        <el-table-column prop="createTime" :label="$t('common.createTimeSort')" width="180"></el-table-column>
        <el-table-column :label="$t('cloudMachine.remark')">
          <template #default="scope">
            {{ getImageDisplayName(scope.row.remark) }}
          </template>
        </el-table-column>

        <el-table-column prop="status" :label="$t('common.status')" width="120">
          <template #default="scope">
            <el-tag size="small" type="info" style="padding: 2px 8px;">{{ scope.row.status === 'running' ? $t('cloudMachine.running') : (scope.row.status === 'shutdown' || scope.row.status === 'exited') ? $t('cloudMachine.shutdown') : scope.row.status === 'created' ? $t('cloudMachine.created') : $t('cloudMachine.restarting') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="scope">
            <el-button 
              size="small" 
              type="primary" 
              @click="switchBackup(scope.row.id)"
              :loading="backupLoading"
              :disabled="backupLoading"
              style="margin-right: 8px;"
            >
              切换
            </el-button>
             <el-button 
              size="small" 
              type="primary" 
              @click="handleRename(scope.row)"
              style="margin-right: 8px;"
            >
              修改名称
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteBackup(scope.row.id)"
              :loading="backupLoading"
              :disabled="backupLoading"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
        <template #empty>
          <div style="padding: 20px; text-align: center; color: #909399;">
            当前坑位没有可用的备份
          </div>
        </template>
      </el-table>
    </div>
  </el-dialog>
  
  <!-- 批量切换备份进度对话框 -->
  <el-dialog
    v-model="batchSwitchBackupProgressVisible"
    title="批量切换备份进度"
    width="520px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="batchSwitchBackupDone >= batchSwitchBackupTotal"
  >
    <div style="padding: 0 4px;">
      <!-- 总进度条 -->
      <div style="margin-bottom: 16px;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
          <span style="font-size: 13px; color: #606266;">总进度</span>
          <span style="font-size: 13px; color: #303133; font-weight: 500;">
            {{ batchSwitchBackupDone }} / {{ batchSwitchBackupTotal }}
          </span>
        </div>
        <el-progress
          :percentage="batchSwitchBackupTotal > 0 ? Math.round(batchSwitchBackupDone / batchSwitchBackupTotal * 100) : 0"
          :status="batchSwitchBackupDone >= batchSwitchBackupTotal
            ? (batchSwitchBackupProgressList.some(i => i.status === 'failed') ? 'exception' : 'success')
            : ''"
          :stroke-width="12"
        />
      </div>
      <!-- 每台云机进度列表 -->
      <div style="max-height: 320px; overflow-y: auto;">
        <div
          v-for="item in batchSwitchBackupProgressList"
          :key="item.slotNum + '-' + item.deviceIp"
          style="display: flex; align-items: center; padding: 7px 0; border-bottom: 1px solid #f0f0f0; gap: 8px;"
        >
          <!-- 坑位 + 设备IP -->
          <div style="min-width: 100px; font-size: 12px; color: #606266; flex-shrink: 0;">
            <span>坑位 {{ item.slotNum }}</span>
            <span v-if="cloudManageMode === 'batch'" style="display: block; color: #909399; font-size: 11px;">{{ item.deviceIp }}</span>
          </div>
          <!-- 备份名称 -->
          <div style="flex: 1; min-width: 0; font-size: 12px; color: #303133; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" :title="item.backupName">
            → {{ item.backupName }}
          </div>
          <!-- 状态 -->
          <div style="min-width: 82px; text-align: right; flex-shrink: 0;">
            <el-tag
              v-if="item.status === 'pending'"
              size="small"
              type="info"
            >等待中</el-tag>
            <el-tag
              v-else-if="item.status === 'running'"
              size="small"
              type="warning"
            >
              <el-icon class="is-loading" style="margin-right: 3px;"><Loading /></el-icon>
              {{ item.message }}
            </el-tag>
            <el-tag
              v-else-if="item.status === 'success'"
              size="small"
              type="success"
            >成功</el-tag>
            <el-tag
              v-else-if="item.status === 'failed'"
              size="small"
              type="danger"
              :title="item.message"
            >失败</el-tag>
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button
        type="primary"
        :disabled="batchSwitchBackupDone < batchSwitchBackupTotal"
        @click="batchSwitchBackupProgressVisible = false"
      >
        {{ batchSwitchBackupDone < batchSwitchBackupTotal ? '处理中...' : '关闭' }}
      </el-button>
    </template>
  </el-dialog>

  <!-- 创建云机弹窗 -->
  <div class="create-dialog-container">
    <el-dialog
      v-model="createDialogVisible"
      :title="createMode === 'batch' ? $t('common.batchCreateCloudMachine') : $t('common.createCloudMachine')"
      width="900px"
      :before-close="handleCreateCancel"
    >
      <!-- 弹窗内容 -->
      <div class="create-dialog-content">
        <!-- SDK加载蒙版 -->
        <!-- <div v-if="sdkLoadingVisible" class="sdk-loading-overlay">
          <div class="sdk-loading-content">
            <el-loading-spinner class="is-medium"></el-loading-spinner>
            <div class="sdk-loading-text">{{ sdkLoadingMessage }}</div>
            <div v-if="sdkLoadingMessage.includes('下载镜像') || sdkLoadingMessage.includes('拉取镜像')" class="sdk-loading-progress">
              <el-progress 
                :percentage="sdkLoadingProgress" 
                :stroke-width="16" 
                :show-text="true"
                status="success"
              ></el-progress>
            </div>
          </div>
        </div> -->
      <!-- 批量创建模式切换 -->
      <div style="margin-bottom: 20px; text-align: center;">
        <div class="create-type-switch">
          <div class="create-type-side create-type-left">
            <span class="create-type-tag create-type-tag-success">{{ $t('common.googleGreen') }}</span>
            <span class="create-type-tag create-type-tag-success">{{ $t('common.realParameters') }}</span>
          </div>
          <el-radio-group v-model="createForm.createType">
            <el-radio-button label="simulator">{{ $t('common.simulator') }}</el-radio-button>
            <el-radio-button label="container">{{ $t('common.container') }}</el-radio-button>
          </el-radio-group>
          <div class="create-type-side create-type-right">
            <span class="create-type-tag create-type-tag-muted">{{ $t('common.oldVersionImage') }}</span>
          </div>
        </div>
        <div style="margin-top: 8px; font-size: 12px; color: #606266; display: flex; justify-content: center; gap: 16px; flex-wrap: wrap;">
          <template v-if="createMode !== 'multi-device-batch'">
            <div style="display: flex; align-items: center; gap: 8px;">
              <span>{{ $t('common.deviceRemainingSpace') }}:</span>
              <span :style="{ color: createDeviceStorageInfo.isLow ? '#F56C6C' : '#606266' }">{{ createDeviceStorageInfo.text }}</span>
              <el-progress
                :percentage="createDeviceStorageInfo.remainingPercent"
                :stroke-width="6"
                :show-text="false"
                :status="createDeviceStorageInfo.isLow ? 'exception' : 'success'"
                style="width: 140px;"
              ></el-progress>
            </div>
            <span v-if="createDeviceStorageInfo.isLow" style="color: #F56C6C;">{{ $t('common.insufficientSpace') }}</span>
            <div style="display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
              <span>{{ $t('common.apiVersion') }}: {{ createDeviceApiVersion }}</span>
              <span v-if="createDeviceApiLatestVersion">/ {{ $t('common.latestVersion') }} {{ createDeviceApiLatestVersion }}</span>
              <el-button v-if="createDeviceApiNeedsUpgrade" size="small" type="warning" @click="upgradeCreateDeviceApiVersion">{{ $t('common.upgrade') }}</el-button>
              <span v-if="createDeviceApiNeedsUpgrade" style="color: #F56C6C;">{{ $t('common.notLatestVersion') }}</span>
            </div>
          </template>
        </div>
      </div>

      <!-- 容器模式表单 -->
      <div v-if="createForm.createType === 'container'" class="create-dialog-container-mode" style="padding: 0 20px;">
        <el-form :model="createForm" label-width="100px">
          <el-form-item :label="$t('common.deviceIP')">
            <template v-if="createMode === 'multi-device-batch'">
              <div style="margin-bottom: 10px;">
                <el-radio-group v-model="batchDeviceTypeFilter" size="small" @change="selectedBatchDevices = []">
                  <el-radio-button label="p_series">{{ $t('common.pSeries') }}</el-radio-button>
                  <el-radio-button label="other_series">{{ $t('common.otherSeries') }}</el-radio-button>
                </el-radio-group>
              </div>
              <el-select 
                v-model="selectedBatchDevices" 
                multiple 
                :placeholder="$t('common.pleaseSelectDevice')"
                collapse-tags
                collapse-tags-tooltip
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="device in filteredBatchDevices"
                  :key="device.id"
                  :label="device.ip + (device.name ? ' (' + device.name + ')' : '')"
                  :value="device.ip"
                >
                </el-option>
              </el-select>
            </template>
            <template v-else>
              <el-input :value="createDevice ? createDevice.ip : ''" disabled></el-input>
            </template>
          </el-form-item>
          <el-form-item :label="$t('common.androidVersion')">
            <el-radio-group v-model="createForm.containerAndroidVersion">
              <el-radio label="10">Android 10</el-radio>
              <el-radio v-if="!isPSeriesOrBatchP" label="12">Android 12</el-radio>
              <el-radio label="14">Android 14</el-radio>
              <el-radio label="custom">{{ $t('common.customImage') }}</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="createMode !== 'multi-device-batch' || selectedBatchDevices.length > 0" :label="$t('common.imageAddress')">
             <el-select v-if="createForm.containerAndroidVersion !== 'custom'" v-model="createForm.containerImageSelect" :placeholder="$t('common.pleaseSelect')" style="width: 100%;">
                <!-- <el-option label="自定义镜像" value="custom"></el-option> -->
                <el-option 
                  v-for="image in filteredContainerImages" 
                  :key="image.url" 
                  :label="image.name" 
                  :value="image.url"
                ></el-option>
             </el-select>
             <div v-if="createForm.containerAndroidVersion === 'custom' || createForm.containerImageSelect === 'custom'" style="width: 100%;">
                <el-input v-model="createForm.containerCustomImageUrl"></el-input>
             </div>
          </el-form-item>

          <div style="display: flex; gap: 20px;">
            <el-form-item :label="$t('common.name')" style="flex: 1;">
              <el-input v-model="createForm.containerName" placeholder="T100"></el-input>
            </el-form-item>
            <el-form-item v-if="createMode !== 'batch'" :label="$t('common.cloudMachineCount')" style="flex: 1;">
              <el-input-number v-model="createForm.containerCount" :min="1" :max="24" style="width: 100%;"></el-input-number>
            </el-form-item>
          </div>

          <el-form-item v-if="createMode === 'batch'" :label="$t('common.slot')">
              <!-- <div style="margin-bottom: 5px;">选择坑位 ({{ isPSeries ? '24' : '12' }}坑位):</div> -->
              <div>
                  <div style="margin-bottom: 5px; font-size: 12px; display: flex; gap: 10px; flex-wrap: wrap;">
                      <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #409EFF; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.normal') }}</span>
                      <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #E6A23C; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.expiringSoon') }}</span>
                      <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #F56C6C; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.expired') }}</span>
                      <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #909399; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.noInstance') }}</span>
                  </div>
                  <el-checkbox-group v-model="createForm.selectedSlots" size="small" class="batch-slot-checkbox-group" style="display: grid; grid-template-columns: repeat(6, 1fr); gap: 5px;">
                      <el-checkbox-button v-for="i in (isPSeries ? 24 : 12)" :key="i" :label="i" style="width: 100%; margin: 0;" :class="getSlotClass(i)">
                          {{ i }}
                      </el-checkbox-button>
                  </el-checkbox-group>
              </div>
          </el-form-item>
          <el-form-item v-if="createMode === 'batch'" :label="$t('common.singleSlotCount')">
              <el-input-number 
              v-model="createForm.containerCount" 
              :min="1" 
              style="width: 100%;"
              :placeholder="$t('common.quantity')"
              ></el-input-number>
          </el-form-item>

          <el-form-item :label="$t('common.resolution')">
             <el-select v-model="createForm.containerResolution" :placeholder="$t('common.pleaseSelect')" style="width: 100%;">
                <el-option label="720 X 1280" value="720x1280x320"></el-option>
                <el-option label="1080 X 1920" value="1080x1920x420"></el-option>
                <el-option label="1200 X 1920（平板）" value="1200x1920x240"></el-option>
                <el-option label="1600 X 2560（平板）" value="1600x2560x320"></el-option>
                <el-option :label="$t('common.customResolution')" value="custom"></el-option>
             </el-select>
             
             <!-- 自定义分辨率输入框 -->
             <div v-if="createForm.containerResolution === 'custom'" class="custom-resolution-container" style="margin-top: 15px;">
                <div style="display: flex; gap: 20px; margin-bottom: 15px;">
                  <div style="flex: 1; display: flex; align-items: center;">
                    <label style="width: 60px; color: #606266;">{{ $t('common.deviceWidth') }}</label>
                    <el-input v-model="createForm.containerCustomResolution.width" style="flex: 1;"></el-input>
                  </div>
                  <div style="flex: 1; display: flex; align-items: center;">
                    <label style="width: 60px; color: #606266;">{{ $t('common.deviceHeight') }}</label>
                    <el-input v-model="createForm.containerCustomResolution.height" style="flex: 1;"></el-input>
                  </div>
                </div>
                <div style="display: flex; gap: 20px; align-items: center;">
                  <div style="flex: 1; display: flex; align-items: center;">
                    <label style="width: 60px; color: #606266;">DPI</label>
                    <el-input v-model="createForm.containerCustomResolution.dpi" style="flex: 1;"></el-input>
                  </div>
                  <div style="flex: 1; color: #f56c6c; font-size: 12px;">
                    {{ $t('common.resolutionWarning') }}
                  </div>
                </div>
             </div>
          </el-form-item>

          <div style="display: flex; gap: 20px;">
            <el-form-item :label="$t('common.dnsType')" style="flex: 1;">
              <el-select v-model="createForm.containerDns" :placeholder="$t('common.pleaseSelect')" style="width: 100%;">
                <el-option :label="$t('common.aliDNS')" value="223.5.5.5"></el-option>
                <el-option :label="$t('common.googleDNS')" value="8.8.8.8"></el-option>
                <el-option :label="$t('common.custom')" value="custom"></el-option>
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('common.dnsAddress')" style="flex: 1;">
              <el-input v-if="createForm.containerDns === 'custom'" v-model="createForm.containerCustomDns" placeholder="223.5.5.5"></el-input>
              <el-input v-else :value="createForm.containerDns" disabled></el-input>
            </el-form-item>
          </div>

          <div style="display: flex; gap: 20px; align-items: center;">
            <el-form-item :label="$t('common.sandboxMode')" style="flex: 1;">
               <el-switch v-model="createForm.containerSandboxMode" :active-text="$t('common.enable')" inline-prompt disabled></el-switch>
            </el-form-item>
            <el-form-item :label="$t('common.dataDiskSize')" style="flex: 1;">
               <el-radio-group v-model="createForm.containerDataDiskSize">
                 <el-radio label="16G">16G</el-radio>
                 <el-radio label="32G">32G</el-radio>
                 <el-radio label="64G">64G</el-radio>
               </el-radio-group>
            </el-form-item>
          </div>

            <el-form-item :label="$t('common.secureMode')" style="flex: 1;">
               <el-switch v-model="createForm.containerEnforce" :active-text="$t('common.enable')" :inactive-text="$t('common.disable')" inline-prompt></el-switch>
            </el-form-item>

          <div v-if="createMode !== 'multi-device-batch'" style="display: flex; gap: 20px; align-items: center;">
            <el-form-item :label="$t('common.networkManagement')" style="flex: 1;">
              <el-select v-model="createForm.vpcGroupId" :placeholder="$t('common.selectGroup')" clearable @change="handleVpcGroupChange" style="width: 130px;" :disabled="createForm.containerNetworkCardType === 'public' && createForm.containerMacVlanIp">
                <el-option v-for="group in vpcGroupList" :key="group.id" :label="group.alias" :value="group.id"></el-option>
              </el-select>
              <el-select v-if="createForm.vpcGroupId && createForm.vpcSelectMode === 'specified'" v-model="createForm.vpcNodeId" :placeholder="$t('common.selectNode')" style="width: 130px; margin-left: 10px;">
                <el-option v-for="node in vpcNodeList" :key="node.id" :label="node.remarks" :value="node.id"></el-option>
              </el-select>
              <el-radio-group v-if="createForm.vpcGroupId" v-model="createForm.vpcSelectMode" style="margin-left: 10px;">
                <el-radio label="specified">{{ $t('common.specifiedNode') }}</el-radio>
                <el-radio label="random">{{ $t('common.randomNode') }}</el-radio>
              </el-radio-group>
            </el-form-item>
          </div>
          
          <!-- 容器模式网卡类型选择 -->
          <el-form-item v-if="createMode !== 'multi-device-batch'" :label="$t('common.networkCardType')">
            <el-radio-group v-model="createForm.containerNetworkCardType" @change="handleContainerNetworkCardTypeChange">
              <el-radio label="private">{{ $t('common.privateNetworkCard') }}({{ $t('common.sharedIP') }})</el-radio>
              <el-radio label="public" :disabled="isPublicNetworkDevice">{{ $t('common.publicNetworkCard') }}({{ $t('common.independentIP') }})</el-radio>
            </el-radio-group>
            
            <!-- 网卡类型功能说明 -->
            <!-- <div style="margin-top: 8px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; font-size: 12px; line-height: 1.6; color: #606266;">
              <div style="margin-bottom: 6px;">
                <span style="font-weight: bold; color: #409EFF;">私有网卡：</span>
                在设备内创建独立的网关和掩码，为每个容器分配该网关下的IP地址。可实现容器间网络隔离，仍可使用网络管理的IP代理功能。
              </div>
              <div>
                <span style="font-weight: bold; color: #67C23A;">公有网卡：</span>
                容器直接使用设备所在局域网的网关和掩码，与设备处于同一网段。容器间无法实现网络隔离，且设置后将无法使用网络管理的IP代理功能。
              </div>
            </div> -->
            
            <!-- 公有网卡 macVlan 提示 -->
            <div v-if="createForm.containerNetworkCardType === 'public' && !hasMacVlan && !fetchingNetworkCards" style="margin-top: 5px; font-size: 12px; line-height: 1.2;">
              <span style="color: #F56C6C;">{{ $t('common.noMacVlanDetected') }}
              </span>
            </div>
          </el-form-item>
          
          <!-- 容器模式私有网卡选择 -->
          <el-form-item :label="$t('common.networkCardSelection')" v-if="createMode !== 'multi-device-batch' && createForm.containerNetworkCardType === 'private'" key="container-nic-select">
            <el-select 
              v-model="createForm.containerMytBridgeName" 
              :placeholder="$t('common.pleaseSelectNetworkCard')" 
              :loading="fetchingNetworkCards"
              clearable
              filterable
            >
              <el-option
                v-for="item in networkCardList"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          
          <!-- 容器模式 MacVlan IP 输入框 -->
          <el-form-item 
            v-if="createForm.containerNetworkCardType === 'public' && hasMacVlan" 
            label="MacVlan IP"
          >
            <el-input v-model="createForm.containerMacVlanIp" :placeholder="getMacVlanIpPlaceholder()"></el-input>
            
            <!-- 批量创建时显示IP范围 -->
            <div v-if="createForm.containerMacVlanIp && createMode === 'batch' && ((createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.containerCount) > 1" style="margin-top: 5px; font-size: 12px; color: #409EFF;">
              <el-icon><InfoFilled /></el-icon>
              {{ $t('common.batchCreateTip').replace('{count}', (createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.containerCount).replace('{range}', calculateIpRange(createForm.containerMacVlanIp, (createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.containerCount)) }}
            </div>
            
            <!-- MacVlan网络信息和注意事项 -->
            <div style="margin-top: 8px;">
              <div v-if="currentDeviceMacVlanInfo.subnet || currentDeviceMacVlanInfo.gw" style="font-size: 12px; color: #606266; margin-bottom: 5px;">
                <span v-if="currentDeviceMacVlanInfo.subnet">子网: {{ currentDeviceMacVlanInfo.subnet }}</span>
                <span v-if="currentDeviceMacVlanInfo.gw" style="margin-left: 10px;">网关: {{ currentDeviceMacVlanInfo.gw }}</span>
              </div>
              <el-alert 
                type="warning" 
                :closable="false"
                style="padding: 8px 12px;"
              >
                <template #title>
                  <div style="font-size: 12px; line-height: 1.6;">
                    <div style="font-weight: bold; margin-bottom: 4px;">⚠️ 重要提示</div>
                    <div>1. 请确保起始IP在子网范围内</div>
                    <div>2. <span style="color: #F56C6C; font-weight: bold;">请务必确认IP地址未被占用</span>,否则会造成IP冲突导致无法访问</div>
                    <div>3. 批量创建时将按顺序使用连续的IP地址(需手动确保可用)</div>
                    <div v-if="createMode === 'batch' && createForm.containerMacVlanIp">4. 当前将使用IP范围: {{ calculateIpRange(createForm.containerMacVlanIp, (createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.containerCount) }}</div>
                  </div>
                </template>
              </el-alert>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- 左右分栏布局 -->
      <div v-else class="create-dialog-left-right">
        <!-- 左侧内容 -->
        <div class="create-dialog-left">
          <el-form :model="createForm" label-width="100px">
            <el-form-item :label="$t('common.deviceIP')">
              <template v-if="createMode === 'multi-device-batch'">
                <div style="margin-bottom: 10px;">
                  <el-radio-group v-model="batchDeviceTypeFilter" size="small" @change="selectedBatchDevices = []">
                    <el-radio-button label="p_series">{{ $t('common.pSeries') }}</el-radio-button>
                    <el-radio-button label="other_series">{{ $t('common.otherSeries') }}</el-radio-button>
                  </el-radio-group>
                </div>
                <el-select 
                  v-model="selectedBatchDevices" 
                  multiple 
                  :placeholder="$t('common.pleaseSelectDevice')"
                  collapse-tags
                  collapse-tags-tooltip
                  filterable
                  style="width: 100%"
                >
                  <el-option
                    v-for="device in filteredBatchDevices"
                    :key="device.id"
                    :label="device.ip + (device.name ? ' (' + device.name + ')' : '')"
                    :value="device.ip"
                  >
                  </el-option>
                </el-select>
              </template>
              <template v-else>
                <el-input :value="createDevice ? createDevice.ip : ''" disabled></el-input>
              </template>
            </el-form-item>
            <el-form-item :label="$t('common.containerName')">
              <el-input v-model="createForm.name" :placeholder="$t('common.enterContainerNamePrefix')"></el-input>
            </el-form-item>
            <el-form-item :label="$t('common.slot')">
              <template v-if="createMode === 'batch'">
                <!-- <div style="margin-bottom: 5px;">选择坑位 ({{ isPSeries ? '24' : '12' }}坑位):</div> -->
                <div>
                    <div style="margin-bottom: 5px; font-size: 12px; display: flex; gap: 10px; flex-wrap: wrap;">
                        <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #409EFF; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.normal') }}</span>
                        <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #E6A23C; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.expiringSoon') }}</span>
                        <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #F56C6C; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.expired') }}</span>
                        <span style="display: flex; align-items: center;"><span style="display: inline-block; width: 10px; height: 10px; background-color: #909399; margin-right: 4px; border-radius: 2px;"></span>{{ $t('common.noInstance') }}</span>
                    </div>
                    <el-checkbox-group v-model="createForm.selectedSlots" size="small" class="batch-slot-checkbox-group" style="display: grid; grid-template-columns: repeat(6, 1fr); gap: 5px;">
                        <el-checkbox-button v-for="i in (isPSeries ? 24 : 12)" :key="i" :label="i" style="width: 100%; margin: 0;" :class="getSlotClass(i)">
                            {{ i }}
                        </el-checkbox-button>
                    </el-checkbox-group>
                </div>
              </template>
              <template v-else>
                <el-input v-model="createForm.startSlot" disabled></el-input>
              </template>
            </el-form-item>
            <el-form-item v-if="createMode === 'batch'" :label="$t('common.singleSlotCountLabel')">
                <el-input-number 
                v-model="createForm.count" 
                :min="1" 
                style="width: 100%;"
                :placeholder="$t('common.quantity')"
                ></el-input-number>
            </el-form-item>
            <!-- 安卓版本 -->
            <el-form-item :label="$t('common.androidVersion')">
              <el-select v-model="createForm.androidVersion" style="width: 100%;">
                <el-option label="Android 11" value="11"></el-option>
                <el-option label="Android 13" value="13"></el-option>
                <el-option label="Android 14" value="14"></el-option>
                <el-option label="Android 15" value="15"></el-option>
                <el-option label="Android 16" value="16"></el-option>
              </el-select>
            </el-form-item>
            <!-- 镜像分类 -->
            <el-form-item :label="$t('common.imageCategory')">
              <el-radio-group v-model="createForm.imageCategory">
                <el-radio label="online">{{ $t('common.onlineImage') }}</el-radio>
                <el-radio label="local">{{ $t('common.localImage') }}</el-radio>
              </el-radio-group>
            </el-form-item>
            
            <!-- 在线镜像选择 -->
            <el-form-item v-if="createForm.imageCategory === 'online'" :label="$t('common.imageSelection')">
              <el-select v-model="createForm.imageSelect" @change="handleImageSelectChange" :loading="fetchingImages" style="width: 100%;" filterable>
                <el-option :label="$t('common.customImage')" value="custom"></el-option>
                <!-- 使用按安卓版本过滤的镜像列表 -->
                <el-option 
                  v-for="image in androidVersionFilteredImageList" 
                  :key="image.url" 
                  :label="image.name" 
                  :value="image.url"
                ></el-option>
              </el-select>
            </el-form-item>
            
            <!-- 本地镜像选择（已按设备类型过滤：P系列设备只显示P系列镜像，非P系列只显示非P系列镜像） -->
            <el-form-item v-if="createForm.imageCategory === 'local'" :label="$t('common.imageSelection')">
              <el-select v-model="createForm.localImageUrl" :placeholder="$t('common.pleaseSelectLocalImage')" style="width: 100%;" filterable>
                <el-option 
                  v-for="image in filteredLocalCachedImages" 
                  :key="image.url" 
                  :label="image.name" 
                  :value="image.url"
                ></el-option>
              </el-select>
            </el-form-item>
            
            <!-- 自定义镜像地址 -->
            <el-form-item v-if="createForm.imageCategory === 'online' && createForm.imageSelect === 'custom'" :label="$t('common.customImageAddress')">
              <el-input v-model="createForm.customImageUrl" :placeholder="$t('common.enterImageAddress')"></el-input>
            </el-form-item>
            
            <!-- 在线镜像缓存到本地创建选项 -->
            <el-form-item v-if="createForm.imageCategory === 'online'" :label="$t('common.creationMethod')">
              <div style="display: flex; align-items: center;">
                <el-checkbox v-model="createForm.cacheToLocal">{{ $t('common.cacheToLocal') }}</el-checkbox>
                <el-tooltip 
                  :content="$t('common.cacheToLocalTip')" 
                  placement="top" 
                  effect="dark"
                  :popper-options="{
                    modifiers: [
                      {
                        name: 'offset',
                        options: {
                          offset: [0, 10]
                        }
                      }
                    ]
                  }"
                >
                  <el-icon style="margin-left: 8px; cursor: help; color: #909399; font-size: 14px;">
                    <QuestionFilled />
                  </el-icon>
                </el-tooltip>
              </div>
            </el-form-item>
            
            <!-- S5代理设置（SDK版本>=25时支持） -->
            <el-form-item :label="$t('common.s5Proxy')" style="margin-bottom: 0;">
              <el-select v-model="createForm.s5Type" :placeholder="$t('common.pleaseSelectProxyType')" style="width: 100%;">
                <el-option :label="$t('common.noProxy')" value="0"></el-option>
                <el-option :label="$t('common.localDNSParsing')" value="1"></el-option>
                <el-option :label="$t('common.serverDNSParsing')" value="2"></el-option>
              </el-select>
            </el-form-item>

            <!-- 是否设置锁屏密码 -->
            <el-form-item :label="$t('common.lockScreenPassword')" style="margin-bottom: 0;">
              <el-input v-model="createForm.lockScreenPassword" :placeholder="$t('common.noLockScreen')" style="width: 100%;"></el-input>
            </el-form-item>

             <el-form-item :label="$t('common.secureMode')">
              <el-switch v-model="createForm.enforce" :active-text="$t('common.enable')" :inactive-text="$t('common.disable')" inline-prompt></el-switch>
            </el-form-item>

            <!-- ADB端口（安全模式下显示） -->
            <el-form-item v-if="createForm.enforce" label="ADB端口">
              <el-input-number v-model="createForm.adbPort" :min="0" :max="65535" :step="1" controls-position="right" style="width: 200px;" @change="validateAdbPort"></el-input-number>
              <span style="margin-left: 8px; color: #909399; font-size: 12px;">设置0不开启ADB</span>
            </el-form-item>
            
            <div v-if="createForm.s5Type !== '0'" style="margin-top: 0; margin-bottom: 20px; padding: 15px; background-color: #f5f7fa; border-radius: 4px;">
              <div style="display: flex; flex-direction: column; gap: 10px;">
                <div style="display: flex; gap: 10px; align-items: center;">
                  <label style="width: 40px; text-align: right;">IP:</label>
                  <el-input v-model="createForm.s5IP" :placeholder="$t('common.enterIP')" style="flex: 1;"></el-input>
                  <label style="margin-left: 20px;">{{ $t('common.port') }}:</label>
                  <el-input v-model="createForm.s5Port" :placeholder="$t('common.enterPort')" style="flex: 1;"></el-input>
                </div>
                <div style="display: flex; gap: 10px; align-items: center;">
                  <label style="width: 40px; text-align: right;">{{ $t('common.user') }}:</label>
                  <el-input v-model="createForm.s5User" :placeholder="$t('common.enterUser')" style="flex: 1;"></el-input>
                  <label style="margin-left: 20px;">{{ $t('common.password') }}:</label>
                  <el-input v-model="createForm.s5Password" type="password" :placeholder="$t('common.enterPassword2')" style="flex: 1;"></el-input>
                </div>
              </div>
            </div>
          </el-form>
        </div>
        
        <!-- 右侧内容 -->
        <div class="create-dialog-right">
          <el-form :model="createForm" label-width="100px">
           
            <el-form-item v-if="createForm.networkMode === 'myt'" :label="$t('common.ipAddress')">
              <el-input v-model="createForm.ipaddr" :placeholder="$t('common.enterStartIP')"></el-input>
            </el-form-item>
            <el-form-item :label="$t('common.dnsAddress')">
              <el-select v-model="createForm.dns" :placeholder="$t('common.pleaseSelectDNSAddress')">
                <el-option :label="'223.5.5.5 (' + $t('common.aliCloud') + ')'" value="223.5.5.5"></el-option>
                <el-option label="8.8.8.8 (Google)" value="8.8.8.8"></el-option>
                <el-option :label="$t('common.custom')" value="custom"></el-option>
              </el-select>
              <el-input v-if="createForm.dns === 'custom'" v-model="createForm.customDns" :placeholder="$t('common.enterCustomDNSAddress')" style="margin-top: 10px;"></el-input>
            </el-form-item>
            <el-form-item :label="$t('common.sandboxSize')">
              <el-input-number v-model="createForm.sandboxSize" :min="1" :max="2000" :step="1" suffix="GB"></el-input-number>
            </el-form-item>
            <el-form-item :label="$t('common.resolution')">
              <el-select v-model="createForm.resolution">
                <el-option :label="$t('common.defaultResolution')" value="default"></el-option>
                <el-option label="720x1280x320" value="720x1280x320"></el-option>
                <el-option label="1080x1920x420" value="1080x1920x420"></el-option>
                <el-option label="1200x1920x240（平板）" value="1200x1920x240"></el-option>
                <el-option label="1600x2560x320（平板）" value="1600x2560x320"></el-option>
              </el-select>
              <div v-if="createForm.resolution === 'custom'" class="custom-resolution-form" style="margin-top: 10px; display: flex; gap: 10px;">
                <el-input v-model="createForm.customResolution.width" :placeholder="$t('common.width')" style="width: 100px;"></el-input>
                <span>x</span>
                <el-input v-model="createForm.customResolution.height" :placeholder="$t('common.height')" style="width: 100px;"></el-input>
                <span>x</span>
                <el-input v-model="createForm.customResolution.dpi" placeholder="DPI" style="width: 100px;"></el-input>
              </div>
            </el-form-item>
            <!-- 地区选择（SDK版本>=25时显示） -->
            <el-form-item :label="$t('common.modelCountry')">
              <el-select 
                v-model="createForm.countryCode" 
                :placeholder="$t('common.pleaseSelectModelCountry')" 
                :loading="countryListLoading" 
                filterable
                @focus="fetchCountryList"
              >
                <el-option 
                  v-for="country in countryList" 
                  :key="country.countryCode" 
                  :label="`${country.countryName} (${getCountryEnglishName(country.countryCode)})`" 
                  :value="country.countryCode"
                ></el-option>
              </el-select>
            </el-form-item>
            <!-- V3设备显示型号选择 -->
            <el-form-item v-if="showV3Options && createMode !== 'multi-device-batch'" :label="$t('common.modelType')">
              <el-radio-group v-model="createForm.modelType" @change="handleModelTypeChange">
                <el-radio label="online">{{ $t('common.onlineModel') }}</el-radio>
                <el-radio label="local">{{ $t('common.localModel') }}</el-radio>
                <el-radio label="backup">{{ $t('common.backupModel') }}</el-radio>
              </el-radio-group>
            </el-form-item>
            <!-- V3设备显示在线型号选择 -->
            <el-form-item v-if="showV3Options && createForm.modelType === 'online'" :label="$t('common.phoneModel')">
              <el-select v-model="createForm.modelName" :placeholder="$t('common.pleaseSelectPhoneModel')" :loading="fetchingModels" filterable>
                <el-option :label="$t('common.random')" value="random"></el-option>
                <el-option 
                  v-for="model in androidVersionFilteredPhoneModels" 
                  :key="model.id" 
                  :label="model.name" 
                  :value="model.name"
                ></el-option>
              </el-select>
            </el-form-item>
            <!-- V3设备显示本地型号选择 -->
            <el-form-item v-if="showV3Options && createForm.modelType === 'local'" :label="$t('common.localModel')">
              <el-select v-model="createForm.localModel" :placeholder="$t('common.pleaseSelectLocalModel')" :loading="fetchingModels" clearable filterable>
                <el-option :label="$t('common.random')" value="random"></el-option>
                <el-option 
                  v-for="model in localPhoneModels" 
                  :key="model.id || model.name" 
                  :label="model.name" 
                  :value="model.name"
                ></el-option>
              </el-select>
            </el-form-item>
            <!-- V3设备显示备份机型选择 -->
            <el-form-item v-if="showV3Options && createForm.modelType === 'backup'" :label="$t('common.backupModel')">
              <el-select v-model="createForm.modelStatic" :placeholder="$t('common.pleaseSelectBackupModel')" :loading="fetchingBackupModels" filterable>
                <el-option :label="$t('common.random')" value="random"></el-option>
                <el-option 
                  v-for="model in backupPhoneModels" 
                  :key="model.name" 
                  :label="model.name" 
                  :value="model.name"
                ></el-option>
              </el-select>
            </el-form-item>
            
            <el-form-item :label="$t('common.advancedOptions')">
              <el-checkbox v-model="createForm.enableMagisk">{{ $t('common.enableMagisk') }}</el-checkbox>
              <el-checkbox v-model="createForm.enableGMS" style="margin-left: 20px;">{{ $t('common.enableGMS') }}</el-checkbox>
            </el-form-item>

            <el-form-item :label="$t('common.latitude') + ' / ' + $t('common.longitude')">
              <el-input v-model="createForm.latitude" :placeholder="$t('common.latitude')" style="width: 110px;"></el-input>
              <span style="margin: 0 8px;">/</span>
              <el-input v-model="createForm.longitud" :placeholder="$t('common.longitude')" style="width: 110px;"></el-input>
            </el-form-item>

            <el-form-item v-if="createMode !== 'multi-device-batch'" :label="$t('common.networkManagement')">
              <el-select v-model="createForm.vpcGroupId" :placeholder="$t('common.selectGroup')" clearable @change="handleVpcGroupChange" style="width: 130px;" :disabled="createForm.networkCardType === 'public' && createForm.macVlanIp">
                <el-option v-for="group in vpcGroupList" :key="group.id" :label="group.alias" :value="group.id" />
              </el-select>
              <el-select v-if="createForm.vpcGroupId && createForm.vpcSelectMode === 'specified'" v-model="createForm.vpcNodeId" :placeholder="$t('common.selectNode')" style="width: 130px; margin-left: 10px;">
                <el-option v-for="node in vpcNodeList" :key="node.id" :label="node.remarks" :value="node.id" />
              </el-select>
              <el-radio-group v-if="createForm.vpcGroupId" v-model="createForm.vpcSelectMode">
                <el-radio label="specified">{{ $t('common.specifiedNode') }}</el-radio>
                <el-radio label="random">{{ $t('common.randomNode') }}</el-radio>
              </el-radio-group>
              <!-- <div v-if="createForm.networkCardType === 'public' && createForm.macVlanIp" style="margin-top: 5px; font-size: 12px; color: #909399;">
                提示：设置了MacVlan（公有网卡）后，网络管理选项被禁用
              </div> -->
            </el-form-item>

            <el-form-item :label="$t('common.randomSystemFiles')">
              <el-switch v-model="createForm.randomFile"></el-switch>
            </el-form-item>

            <!-- V3设备网卡选择 -->
            <template v-if="showV3Options && createMode !== 'multi-device-batch'">
              <el-form-item :label="$t('common.networkCardType')">
                <el-radio-group v-model="createForm.networkCardType" @change="handleNetworkCardTypeChange">
                  <el-radio label="private">{{ $t('common.privateNetworkCard') }}({{ $t('common.sharedIP') }})</el-radio>
                  <el-radio label="public" :disabled="isPublicNetworkDevice">{{ $t('common.publicNetworkCard') }}({{ $t('common.independentIP') }})</el-radio>
                </el-radio-group>
                
                <!-- 网卡类型功能说明 -->
                <!-- <div style="margin-top: 8px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; font-size: 12px; line-height: 1.6; color: #606266;">
                  <div style="margin-bottom: 6px;">
                    <span style="font-weight: bold; color: #409EFF;">私有网卡：</span>
                    在设备内创建独立的网关和掩码，为每个虚拟机分配该网关下的IP地址。可实现虚拟机间网络隔离，仍可使用网络管理的IP代理功能。
                  </div>
                  <div>
                    <span style="font-weight: bold; color: #67C23A;">公有网卡：</span>
                    虚拟机直接使用设备所在局域网的网关和掩码，与设备处于同一网段。虚拟机间无法实现网络隔离，且设置后将无法使用网络管理的IP代理功能。
                  </div>
                </div> -->
                
                <!-- 公有网卡 macVlan 提示 -->
                <div v-if="createForm.networkCardType === 'public' && !hasMacVlan && !fetchingNetworkCards" style="margin-top: 5px; font-size: 12px; line-height: 1.2;">
                  <span style="color: #F56C6C;">{{ $t('common.noMacVlanDetected') }}
                  </span>
                </div>
              </el-form-item>
              
              <el-form-item :label="$t('common.networkCardSelection')" v-if="createForm.networkCardType === 'private'" key="create-nic-select">
                <el-select 
                  v-model="createForm.mytBridgeName" 
                  :placeholder="$t('common.pleaseSelectNetworkCard')" 
                  :loading="fetchingNetworkCards"
                  clearable
                  filterable
                >
                  <el-option
                    v-for="item in networkCardList"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              
              <!-- MacVlan IP 输入框 -->
              <el-form-item 
                v-if="createForm.networkCardType === 'public' && hasMacVlan" 
                label="MacVlan IP"
              >
                 <el-input v-model="createForm.macVlanIp" :placeholder="getMacVlanIpPlaceholder()"></el-input>
                 
                 <!-- 批量创建时显示IP范围 -->
                 <div v-if="createForm.macVlanIp && createMode === 'batch' && ((createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.count) > 1" style="margin-top: 5px; font-size: 12px; color: #409EFF;">
                   <el-icon><InfoFilled /></el-icon>
                   {{ $t('common.batchCreateTip').replace('{count}', (createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.count).replace('{range}', calculateIpRange(createForm.macVlanIp, (createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.count)) }}
                 </div>
                 
                 <!-- MacVlan网络信息和注意事项 -->
                 <div style="margin-top: 8px;">
                   <div v-if="currentDeviceMacVlanInfo.subnet || currentDeviceMacVlanInfo.gw" style="font-size: 12px; color: #606266; margin-bottom: 5px;">
                     <span v-if="currentDeviceMacVlanInfo.subnet">子网: {{ currentDeviceMacVlanInfo.subnet }}</span>
                     <span v-if="currentDeviceMacVlanInfo.gw" style="margin-left: 10px;">网关: {{ currentDeviceMacVlanInfo.gw }}</span>
                   </div>
                   <el-alert 
                     type="warning" 
                     :closable="false"
                     style="padding: 8px 12px;"
                   >
                     <template #title>
                       <div style="font-size: 12px; line-height: 1.6;">
                         <div style="font-weight: bold; margin-bottom: 4px;">⚠️ 重要提示</div>
                         <div>1. 请确保起始IP在子网范围内</div>
                         <div>2. <span style="color: #F56C6C; font-weight: bold;">请务必确认IP地址未被占用</span>,否则会造成IP冲突导致无法访问</div>
                         <div>3. 批量创建时将按顺序使用连续的IP地址(需手动确保可用)</div>
                         <div v-if="createMode === 'batch' && createForm.macVlanIp">4. 当前将使用IP范围: {{ calculateIpRange(createForm.macVlanIp, (createForm.selectedSlots ? createForm.selectedSlots.length : 1) * createForm.count) }}</div>
                       </div>
                     </template>
                   </el-alert>
                 </div>
              </el-form-item>
            </template>

          </el-form>
        </div>
      </div>
    </div>
    
    <!-- 弹窗底部 -->
    <template #footer>
      <div class="create-dialog-footer">
        <el-button @click="handleCreateCancel">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleCreateSubmit">{{ $t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>
  </div>
  
  <!-- 更新镜像弹窗 -->
  <el-dialog
    v-model="updateImageDialogVisible"
    :title="$t('common.updateImage')"
    width="900px"
    :before-close="handleUpdateImageCancel"
  >
    <!-- 弹窗内容 -->
    <div class="create-dialog-content">
      <!-- 容器模式 (V2) -->
      <div v-if="updateImageContainer && updateImageContainer.androidType === 'V2'" class="create-dialog-container-mode" style="padding: 0 20px;">
        <div class="create-dialog-container-mode-title">{{ $t('common.updateImageWarning') }}</div>
        <el-form :model="updateImageForm" label-width="100px">
          <el-form-item :label="$t('common.imageAddress')">
             <el-select v-model="updateImageForm.imageSelect" :placeholder="$t('common.pleaseSelect')" style="width: 100%;">
              <el-option :label="$t('common.customImage')" value="custom"></el-option>
                <el-option 
                  v-for="image in filteredContainerImagesForUpdate" 
                  :key="image.url" 
                  :label="image.name" 
                  :value="image.url"
                ></el-option>
             </el-select>
             
             <!-- 自定义镜像URL输入框 -->
             <div v-if="updateImageForm.imageSelect === 'custom'" style="margin-top: 10px;width: 100%;">
               <el-input 
                 v-model="updateImageForm.customImageUrl" 
                 :placeholder="$t('common.enterCustomImageAddress')"
                 clearable
               ></el-input>
             </div>
          </el-form-item>

          <div style="display: flex; gap: 20px;">
            <el-form-item label="名称" style="flex: 1;">
              <el-input :value="updateImageContainer ? (() => {
                const nameParts = updateImageContainer.name.split('_');
                return nameParts[nameParts.length - 1] || updateImageContainer.name;
              })() : ''" disabled></el-input>
            </el-form-item>
            <!-- <el-form-item label="云机数量" style="flex: 1;">
              <el-input-number :model-value="1" disabled style="width: 100%;"></el-input-number>
            </el-form-item> -->
          </div>

          <el-form-item label="分辨率">
             <el-select v-model="updateImageForm.resolution" placeholder="请选择" style="width: 100%;">
                <el-option label="720 X 1280" value="720x1280x320"></el-option>
                <el-option label="1080 X 1920" value="1080x1920x420"></el-option>
                <el-option label="自定义分辨率" value="custom"></el-option>
             </el-select>
             
             <!-- 自定义分辨率输入框 -->
             <div v-if="updateImageForm.resolution === 'custom'" class="custom-resolution-container" style="margin-top: 15px;">
                <div style="display: flex; gap: 20px; margin-bottom: 15px;">
                  <div style="flex: 1; display: flex; align-items: center;">
                    <label style="width: 60px; color: #606266;">设备宽</label>
                    <el-input v-model="updateImageForm.customResolution.width" style="flex: 1;"></el-input>
                  </div>
                  <div style="flex: 1; display: flex; align-items: center;">
                    <label style="width: 60px; color: #606266;">设备长</label>
                    <el-input v-model="updateImageForm.customResolution.height" style="flex: 1;"></el-input>
                  </div>
                </div>
                <div style="display: flex; gap: 20px; align-items: center;">
                  <div style="flex: 1; display: flex; align-items: center;">
                    <label style="width: 60px; color: #606266;">DPI</label>
                    <el-input v-model="updateImageForm.customResolution.dpi" style="flex: 1;"></el-input>
                  </div>
                  <div style="flex: 1; color: #f56c6c; font-size: 12px;">
                    请注意，自定义分辨率可能引发样式适配异常
                  </div>
                </div>
             </div>
          </el-form-item>

          <div style="display: flex; gap: 20px;">
            <el-form-item label="DNS 类型" style="flex: 1;">
              <el-select v-model="updateImageForm.dns" placeholder="请选择" style="width: 100%;">
                <el-option label="阿里DNS(223.5.5.5)" value="223.5.5.5"></el-option>
                <el-option label="Google(8.8.8.8)" value="8.8.8.8"></el-option>
                <el-option label="自定义" value="custom"></el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="DNS 地址" style="flex: 1;">
              <el-input v-if="updateImageForm.dns === 'custom'" v-model="updateImageForm.customDns" placeholder="223.5.5.5"></el-input>
              <el-input v-else :value="updateImageForm.dns" disabled></el-input>
            </el-form-item>
          </div>
          
          <div style="display: flex; gap: 20px; align-items: center;">
            <el-form-item label="网络管理" style="flex: 1;">
              <el-select v-model="updateImageForm.vpcGroupId" placeholder="选择分组" clearable @change="handleVpcGroupChange" style="width: 130px;" :disabled="updateImageForm.networkCardType === 'public' && updateImageForm.macVlanIp">
                <el-option v-for="group in vpcGroupList" :key="group.id" :label="group.alias" :value="group.id" />
              </el-select>
              <el-select v-if="updateImageForm.vpcGroupId && updateImageForm.vpcSelectMode === 'specified'" v-model="updateImageForm.vpcNodeId" placeholder="选择节点" style="width: 130px; margin-left: 10px;">
                <el-option v-for="node in vpcNodeList" :key="node.id" :label="extractNodeDisplayName(node.remarks)" :value="node.id" />
              </el-select>
              <el-radio-group v-if="updateImageForm.vpcGroupId" v-model="updateImageForm.vpcSelectMode" style="margin-left: 10px;">
                <el-radio label="specified">指定节点</el-radio>
                <el-radio label="random">随机节点</el-radio>
              </el-radio-group>
            </el-form-item>
          </div>

          <el-form-item label="网卡类型">
            <el-radio-group v-model="updateImageForm.networkCardType" @change="handleUpdateNetworkCardTypeChange">
              <el-radio label="private">{{ $t('common.privateNetworkCard') }}({{ $t('common.sharedIP') }})</el-radio>
              <el-radio label="public" :disabled="isActiveDevicePublic">{{ $t('common.publicNetworkCard') }}({{ $t('common.independentIP') }})</el-radio>
            </el-radio-group>
            
            <!-- 网卡类型功能说明 -->
            <!-- <div style="margin-top: 8px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; font-size: 12px; line-height: 1.6; color: #606266;">
              <div style="margin-bottom: 6px;">
                <span style="font-weight: bold; color: #409EFF;">私有网卡：</span>
                在设备内创建独立的网关和掩码，为每个容器分配该网关下的IP地址。可实现容器间网络隔离，仍可使用网络管理的IP代理功能。
              </div>
              <div>
                <span style="font-weight: bold; color: #67C23A;">公有网卡：</span>
                容器直接使用设备所在局域网的网关和掩码，与设备处于同一网段。容器间无法实现网络隔离，且设置后将无法使用网络管理的IP代理功能。
              </div>
            </div> -->
            
            <div v-if="updateImageForm.networkCardType === 'public' && !hasMacVlan && !fetchingNetworkCards" style="margin-top: 5px; font-size: 12px; line-height: 1.2;">
              <span style="color: #F56C6C;">未检测到MacVlan配置，请前往网络管理-公有网卡创建</span>
            </div>
          </el-form-item>

          <el-form-item v-if="updateImageForm.networkCardType === 'private'" label="网卡选择">
            <el-select
              v-model="updateImageForm.mytBridgeName"
              placeholder="请选择网卡"
              :loading="fetchingNetworkCards"
              clearable
              filterable
              style="width: 100%;"
            >
              <el-option
                v-for="item in networkCardList"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item v-if="updateImageForm.networkCardType === 'public' && hasMacVlan" label="MacVlan IP">
            <el-input v-model="updateImageForm.macVlanIp" :placeholder="getMacVlanIpPlaceholder()"></el-input>
            
            <!-- MacVlan网络信息和注意事项 -->
            <div style="margin-top: 8px;">
              <div v-if="currentDeviceMacVlanInfo.subnet || currentDeviceMacVlanInfo.gw" style="font-size: 12px; color: #606266; margin-bottom: 5px;">
                <span v-if="currentDeviceMacVlanInfo.subnet">子网: {{ currentDeviceMacVlanInfo.subnet }}</span>
                <span v-if="currentDeviceMacVlanInfo.gw" style="margin-left: 10px;">网关: {{ currentDeviceMacVlanInfo.gw }}</span>
              </div>
              <el-alert 
                type="warning" 
                :closable="false"
                style="padding: 8px 12px;"
              >
                <template #title>
                  <div style="font-size: 12px; line-height: 1.6;">
                    <div style="font-weight: bold; margin-bottom: 4px;">⚠️ 重要提示</div>
                    <div>1. 请确保IP在子网范围内</div>
                    <div>2. <span style="color: #F56C6C; font-weight: bold;">请务必确认IP地址未被占用</span>,否则会造成IP冲突导致无法访问</div>
                  </div>
                </template>
              </el-alert>
            </div>
          </el-form-item>

          <el-form-item :label="$t('common.secureMode')">
            <el-switch v-model="updateImageForm.enforce" :active-text="$t('common.enable')" :inactive-text="$t('common.disable')" inline-prompt></el-switch>
          </el-form-item>
        </el-form>
      </div>

      <!-- 模拟器模式 (V0/V1/V3) -->
      <div v-else class="create-dialog-left-right">
        <!-- 左侧内容 -->
        <div class="create-dialog-left">
          <el-form :model="updateImageForm" label-width="100px">
            <el-form-item label="设备IP">
              <el-input :value="activeDevice ? activeDevice.ip : ''" disabled></el-input>
            </el-form-item>
            <el-form-item label="容器名称">
              <el-input :value="updateImageContainer ? (() => {
                const nameParts = updateImageContainer.name.split('_');
                return nameParts[nameParts.length - 1] || updateImageContainer.name;
              })() : ''" disabled></el-input>
            </el-form-item>
            <el-form-item label="当前镜像">
              <el-input :value="updateImageContainer ? getImageDisplayName(updateImageContainer.image) : ''" disabled></el-input>
            </el-form-item>
            <!-- V3设备显示型号选择 -->
            <!-- <el-form-item v-if="activeDevice && activeDevice.version === 'v3'" label="手机型号">
              <el-select v-model="updateImageForm.modelName" placeholder="请选择手机型号" :loading="fetchingModels" filterable>
                <el-option 
                  v-for="model in phoneModels" 
                  :key="model.id" 
                  :label="model.name" 
                  :value="model.name"
                ></el-option>
              </el-select>
            </el-form-item> -->
            <el-form-item :label="$t('common.imageSelection')">
              <el-select v-model="updateImageForm.imageSelect" @change="handleImageSelectChange" :loading="fetchingImages" style="width: 100%;" filterable>
                <el-option :label="$t('common.customImage')" value="custom"></el-option>
                <!-- 使用从API获取的镜像列表（按 os_ver 过滤） -->
                <el-option 
                  v-for="image in filteredImageListForUpdate" 
                  :key="image.url" 
                  :label="image.name" 
                  :value="image.url"
                ></el-option>
              </el-select>
            </el-form-item>
            <el-form-item v-if="updateImageForm.imageSelect === 'custom'" :label="$t('common.customImageAddress')">
              <el-input v-model="updateImageForm.customImageUrl" :placeholder="$t('common.enterImageAddress')"></el-input>
            </el-form-item>
            <el-form-item label="DNS地址">
              <el-select v-model="updateImageForm.dns" placeholder="请选择DNS地址">
                <el-option label="223.5.5.5 (阿里云)" value="223.5.5.5"></el-option>
                <el-option label="8.8.8.8 (Google)" value="8.8.8.8"></el-option>
                <el-option label="自定义" value="custom"></el-option>
              </el-select>
              <el-input v-if="updateImageForm.dns === 'custom'" v-model="updateImageForm.customDns" placeholder="输入自定义DNS地址" style="margin-top: 10px;"></el-input>
            </el-form-item>
            
            <el-form-item label="网络管理">
              <el-select v-model="updateImageForm.vpcGroupId" placeholder="选择分组" clearable @change="handleVpcGroupChange" :disabled="updateImageForm.networkCardType === 'public' && updateImageForm.macVlanIp" style="width: 130px;">
                <el-option v-for="group in vpcGroupList" :key="group.id" :label="group.alias" :value="group.id" />
              </el-select>
              <el-select v-if="updateImageForm.vpcGroupId && updateImageForm.vpcSelectMode === 'specified'" v-model="updateImageForm.vpcNodeId" placeholder="选择节点" style="width: 130px; margin-left: 10px;">
                <el-option v-for="node in vpcNodeList" :key="node.id" :label="extractNodeDisplayName(node.remarks)" :value="node.id" />
              </el-select>
              <el-radio-group v-if="updateImageForm.vpcGroupId" v-model="updateImageForm.vpcSelectMode">
                <el-radio label="specified">指定节点</el-radio>
                <el-radio label="random">随机节点</el-radio>
              </el-radio-group>
            </el-form-item>
            
            <el-form-item label="随机系统文件">
              <el-switch v-model="updateImageForm.randomFile"></el-switch>
            </el-form-item>

            <el-form-item :label="$t('common.secureMode')">
              <el-switch v-model="updateImageForm.enforce" :active-text="$t('common.enable')" :inactive-text="$t('common.disable')" inline-prompt></el-switch>
            </el-form-item>
          </el-form>
        </div>
        
        <!-- 右侧内容 -->
        <div class="create-dialog-right">
          <el-form :model="updateImageForm" label-width="100px">
            <!-- V3设备显示高级选项 -->
            <el-form-item v-if="activeDevice && activeDevice.version === 'v3'" label="高级选项">
              <el-checkbox v-model="updateImageForm.enableMagisk">启用Magisk</el-checkbox>
              <el-checkbox v-model="updateImageForm.enableGMS" style="margin-left: 20px;">启用GMS</el-checkbox>
            </el-form-item>
            <!-- V3设备网卡选择 -->
            <template v-if="activeDevice && activeDevice.version === 'v3'">
              <el-form-item label="网卡类型">
                <el-radio-group v-model="updateImageForm.networkCardType" @change="handleUpdateNetworkCardTypeChange">
                  <el-radio label="private">{{ $t('common.privateNetworkCard') }}({{ $t('common.sharedIP') }})</el-radio>
                  <el-radio label="public" :disabled="isActiveDevicePublic">{{ $t('common.publicNetworkCard') }}({{ $t('common.independentIP') }})</el-radio>
                </el-radio-group>
                
                <!-- 网卡类型功能说明 -->
                <!-- <div style="margin-top: 8px; padding: 8px 12px; background: #f5f7fa; border-radius: 4px; font-size: 12px; line-height: 1.6; color: #606266;">
                  <div style="margin-bottom: 6px;">
                    <span style="font-weight: bold; color: #409EFF;">私有网卡：</span>
                    在设备内创建独立的网关和掩码，为每个虚拟机分配该网关下的IP地址。可实现虚拟机间网络隔离，仍可使用网络管理的IP代理功能。
                  </div>
                  <div>
                    <span style="font-weight: bold; color: #67C23A;">公有网卡：</span>
                    虚拟机直接使用设备所在局域网的网关和掩码，与设备处于同一网段。虚拟机间无法实现网络隔离，且设置后将无法使用网络管理的IP代理功能。
                  </div>
                </div> -->
                
                <!-- 公有网卡 macVlan 提示 -->
                <div v-if="updateImageForm.networkCardType === 'public' && !hasMacVlan && !fetchingNetworkCards" style="margin-top: 5px; font-size: 12px; line-height: 1.2;">
                  <span style="color: #F56C6C;">
                    未检测到MacVlan配置，请前往<span style="color: #409EFF; cursor: pointer; text-decoration: underline;" @click="activeTab = 'network'; activeNetworkTab = 'public-nic'">网络管理-公有网卡</span>创建
                  </span>
                </div>
              </el-form-item>
              
              
              <el-form-item label="网卡选择" v-if="updateImageForm.networkCardType === 'private'" key="update-nic-select">
                <el-select 
                  v-model="updateImageForm.mytBridgeName" 
                  placeholder="请选择网卡" 
                  :loading="fetchingNetworkCards"
                  clearable
                  filterable
                >
                  <el-option
                    v-for="item in networkCardList"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              
              
              <!-- MacVlan IP 输入框 -->
              <el-form-item 
                v-if="updateImageForm.networkCardType === 'public' && hasMacVlan" 
                label="MacVlan IP"
              >
                 <el-input v-model="updateImageForm.macVlanIp" :placeholder="getMacVlanIpPlaceholder()"></el-input>
                 
                 <!-- MacVlan网络信息和注意事项 -->
                 <div style="margin-top: 8px;">
                   <div v-if="currentDeviceMacVlanInfo.subnet || currentDeviceMacVlanInfo.gw" style="font-size: 12px; color: #606266; margin-bottom: 5px;">
                     <span v-if="currentDeviceMacVlanInfo.subnet">子网: {{ currentDeviceMacVlanInfo.subnet }}</span>
                     <span v-if="currentDeviceMacVlanInfo.gw" style="margin-left: 10px;">网关: {{ currentDeviceMacVlanInfo.gw }}</span>
                   </div>
                   <el-alert 
                     type="warning" 
                     :closable="false"
                     style="padding: 8px 12px;"
                   >
                     <template #title>
                         <div style="font-size: 12px; line-height: 1.6;">
                         <div style="font-weight: bold; margin-bottom: 4px;">{{ $t('common.importantTip') }}</div>
                         <div>1. {{ $t('common.ensureIPInSubnet') }}</div>
                         <div>2. <span style="color: #F56C6C; font-weight: bold;">{{ $t('common.ensureIPNotUsed') }}</span>,{{ $t('common.ipConflictWarning') }}</div>
                       </div>
                     </template>
                   </el-alert>
                 </div>
              </el-form-item>
            </template>
          </el-form>
        </div>
      </div>
    </div>
    
    <!-- 弹窗底部 -->
    <template #footer>
      <div class="create-dialog-footer">
        <el-button @click="handleUpdateImageCancel">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleUpdateImageSubmit" :loading="updateImageLoading">{{ $t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 批量更新镜像对话框 -->
  <el-dialog
    v-model="batchUpdateImageDialogVisible"
    :title="$t('common.batchUpdateImage')"
    width="640px"
    :close-on-click-modal="false"
  >
    <div v-for="(group, gIdx) in batchUpdateImageGroups" :key="group.groupKey" :style="{ marginBottom: gIdx < batchUpdateImageGroups.length - 1 ? '20px' : '0' }">
      <!-- 分组标题 -->
      <div style="font-weight:600;font-size:14px;color:#303133;padding:6px 0 10px 0;border-bottom:1px solid #ebeef5;margin-bottom:12px;">
        {{ group.groupLabel }}
        <span style="font-weight:400;color:#909399;font-size:12px;margin-left:8px;">{{ $t('common.totalCloudMachines', { count: group.containers.length }) }}</span>
      </div>

      <el-form label-width="90px">
        <!-- 若同时含 V2 和 V3，让用户选择版本 -->
        <el-form-item v-if="group.hasV2 && group.hasV3" :label="$t('common.updateVersion')">
          <el-radio-group v-model="group.androidType" @change="group.selectedUrl = ''; group.customUrl = ''">
            <el-radio label="V3">{{ $t('common.v3Simulator') }}</el-radio>
            <el-radio label="V2">{{ $t('common.v2Container') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-else :label="$t('common.versionType')">
          <span style="color:#606266;">{{ group.androidType === 'V2' ? $t('common.v2Container') : $t('common.v3Simulator') }}</span>
        </el-form-item>

        <!-- V2 模式：安卓版本选择 -->
        <el-form-item v-if="group.androidType === 'V2'" :label="$t('common.androidVersion')">
          <el-radio-group v-model="group.v2AndroidVersion" @change="group.selectedUrl = ''">
            <el-radio :label="10">Android 10</el-radio>
            <el-radio v-if="!isBatchImagePSeries(group.deviceName)" :label="12">Android 12</el-radio>
            <el-radio :label="14">Android 14</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 镜像选择 -->
        <el-form-item :label="$t('common.imageSelection')">
          <el-select
            v-model="group.selectedUrl"
            filterable
            style="width: 100%;"
            :placeholder="$t('common.pleaseSelectImage')"
          >
            <el-option :label="$t('common.customImage')" value="custom" />
            <template v-if="group.androidType === 'V3'">
              <el-option
                v-for="img in getBatchUpdateV3List(group.deviceName)"
                :key="img.url"
                :label="img.name"
                :value="img.url"
              />
            </template>
            <template v-else>
              <el-option
                v-for="img in getBatchUpdateV2List(group.deviceName, group.v2AndroidVersion)"
                :key="img.url"
                :label="img.name"
                :value="img.url"
              />
            </template>
          </el-select>
          <div
            v-if="group.androidType === 'V2' && getBatchUpdateV2List(group.deviceName, group.v2AndroidVersion).length === 0 && group.selectedUrl !== 'custom'"
            style="color:#909399;font-size:12px;margin-top:4px;"
          >
            {{ $t('common.noImageForVersion') }}
          </div>
        </el-form-item>

        <!-- 自定义地址 -->
        <el-form-item v-if="group.selectedUrl === 'custom'" :label="$t('common.customAddress')">
          <el-input
            v-model="group.customUrl"
            :placeholder="$t('common.enterImageURL')"
            clearable
          />
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="create-dialog-footer">
        <el-button @click="batchUpdateImageDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="executeBatchUpdateImage"
        >{{ $t('common.confirmUpdate') }}</el-button>
      </div>
    </template>
  </el-dialog>
  
  <!-- IP连接测试弹窗 -->
  <el-dialog
    v-model="ipTestVisible"
    :title="$t('common.ipTestTitle')"
    width="500px"
    center
  >
    <div class="ip-test-container">
      <el-form label-width="80px">
        <el-form-item :label="$t('common.ipAddress')">
          <el-input v-model="testIp" :placeholder="$t('common.enterIP')" size="small"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="small" style="margin-right: 10px;" @click="ElMessage.info('功能正在开发中')">{{ $t('common.startTest') }}</el-button>
          <el-button size="small" @click="ipTestVisible = false">{{ $t('common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
  </el-dialog>
  
  <!-- 添加macvlan网络弹窗 -->
  <el-dialog
    v-model="addMacvlanDialogVisible"
    :title="$t('common.addMacvlanNetwork')"
    width="600px"
    :before-close="handleAddMacvlanCancel"
  >
    <div class="add-macvlan-content">
      <el-form :model="addMacvlanForm" label-width="120px">
        <el-form-item :label="$t('common.networkName')" required>
          <el-input v-model="addMacvlanForm.networkName" :placeholder="$t('common.enterNetworkName')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.physicalInterface')" required>
          <el-input v-model="addMacvlanForm.parentInterface" :placeholder="$t('common.enterPhysicalInterface')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.subnet')" required>
          <el-input v-model="addMacvlanForm.subnet" :placeholder="$t('common.enterSubnet')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.gateway')" required>
          <el-input v-model="addMacvlanForm.gateway" :placeholder="$t('common.enterGateway')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.ipRange')">
          <el-input v-model="addMacvlanForm.ipRange" :placeholder="$t('common.enterIPRange')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.isolationMode')">
          <el-checkbox v-model="addMacvlanForm.isPrivate">{{ $t('common.enablePrivateIsolation') }}</el-checkbox>
          <div style="font-size: 12px; color: #909399; margin-top: 5px;">
            {{ $t('common.isolationTip') }}
          </div>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleAddMacvlanCancel">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleAddMacvlanSubmit" :loading="addMacvlanLoading">{{ $t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 修改网络弹窗 -->
  <el-dialog
    v-model="editNetworkDialogVisible"
    :title="$t('common.editNetwork')"
    width="600px"
    :before-close="handleEditNetworkCancel"
  >
    <div class="edit-network-content">
      <el-form :model="editNetworkForm" label-width="120px">
        <el-form-item :label="$t('common.networkName')" required>
          <el-input v-model="editNetworkForm.networkName" :placeholder="$t('common.enterNetworkName')" disabled></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.subnet')" required>
          <el-input v-model="editNetworkForm.subnet" :placeholder="$t('common.enterSubnet')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.gateway')" required>
          <el-input v-model="editNetworkForm.gateway" :placeholder="$t('common.enterGateway')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.ipRange')">
          <el-input v-model="editNetworkForm.ipRange" :placeholder="$t('common.enterIPRange')"></el-input>
        </el-form-item>
        <el-form-item :label="$t('common.isolationMode')">
          <el-checkbox v-model="editNetworkForm.isPrivate">{{ $t('common.enablePrivateIsolation') }}</el-checkbox>
          <div style="font-size: 12px; color: #909399; margin-top: 5px;">
            {{ $t('common.isolationTip') }}
          </div>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleEditNetworkCancel">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleEditNetworkSubmit" :loading="editNetworkLoading">{{ $t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>

  <!-- API详情对话框 -->
  <el-dialog 
    v-model="apiDetailsVisible" 
    :title="$t('common.apiDetails')" 
    width="600px"
    :close-on-click-modal="false"
    style="max-height: 600px;"
  >
    <div v-if="apiDetailsData" class="api-details-content">
      <div class="api-details-header">
        <p><strong>{{ $t('common.slot') }}:</strong> {{ apiDetailsData.slotNum }}</p>
        <p><strong>{{ $t('common.instanceName') }}:</strong> {{ apiDetailsData.instanceName }}</p>
        <p><strong>{{ $t('common.deviceIP') }}:</strong> {{ apiDetailsData.deviceIp }}</p>
        <p><strong>{{ $t('common.deviceVersion') }}:</strong> {{ apiDetailsData.deviceVersion }}</p>
      </div>
      
      <div class="api-details-table">
        <h4>{{ $t('common.portMappingInfo') }}</h4>
        <el-table :data="Object.values(apiDetailsData.portMappings)" size="small" class="port-mapping-table">
          <el-table-column prop="description" :label="$t('common.service')" width="150"></el-table-column>
          <el-table-column :label="$t('common.portMapping')" width="120">
            <template #default="{ row }">
              {{ row.originalPort }} → {{ row.mappedPort }}
            </template>
          </el-table-column>
          <el-table-column prop="url" :label="$t('common.accessAddress')" min-width="200">
            <template #default="{ row }">
              <span style="cursor: pointer; color: #409EFF;" @click="copyToClipboard(row.url)">
                {{ row.url }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <div class="api-details-footer">
        <p style="color: #909399; font-size: 12px; margin-top: 10px;">
          {{ $t('common.clickToCopy') }}
        </p>
      </div>
    </div>
    
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="apiDetailsVisible = false">{{ $t('common.close') }}</el-button>
      </span>
    </template>
  </el-dialog>

  <!-- S5代理设置弹窗 -->
  <el-dialog
    v-model="s5ProxyDialogVisible"
    :title="$t('common.setS5Proxy')"
    width="550px"
  >
    <el-form :model="s5ProxyForm" label-width="120px">
      <el-form-item :label="$t('common.cloudMachineName')">
        <el-input
          v-model="s5ProxyForm.cloudMachineName"
          :placeholder="$t('common.cloudMachineName')"
          readonly
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.s5Info')">
        <el-input
          v-model="s5ProxyForm.vpcInfo"
          :placeholder="$t('common.s5InfoFormat')"
          @blur="parseVpcInfo"
          clearable
        >
          <template #append>
            <el-button @click="parseVpcInfo">{{ $t('common.parseAndFill') }}</el-button>
          </template>
        </el-input>
        <div style="font-size: 12px; color: #909399; margin-top: 4px;">
          {{ $t('common.s5InfoExample') }}
        </div>
      </el-form-item>
      <el-form-item :label="$t('common.s5ServerAddress')" required>
        <el-input
          v-model="s5ProxyForm.s5ServerAddress"
          :placeholder="$t('common.enterS5ServerAddress')"
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.s5Port')" required>
        <el-input
          v-model="s5ProxyForm.s5Port"
          :placeholder="$t('common.enterS5Port')"
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.username')">
        <el-input
          v-model="s5ProxyForm.username"
          :placeholder="$t('common.enterUsername')"
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.password')">
        <el-input
          v-model="s5ProxyForm.password"
          :placeholder="$t('common.enterPassword')"
          type="password"
          show-password
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.dnsMode')">
        <el-radio-group v-model="s5ProxyForm.dnsMode">
          <el-radio label="local">{{ $t('common.localDNS') }}</el-radio>
          <el-radio label="server">{{ $t('common.serverDNS') }}</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="s5ProxyDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleS5ProxySubmit" :loading="s5ProxyLoading">
          {{ s5ProxyLoading ? $t('common.submitting') : $t('common.submitNow') }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 右键菜单 -->
  <div 
    v-if="contextMenuVisible"
    ref="contextMenuRef"
    class="context-menu"
    :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
  >
  <div class="context-menu-item" @click="handleSetStream">
     <el-icon><Setting /></el-icon>
     <span>{{ $t('cloudMachine.setStream') }}</span>
   </div>
   <div class="context-menu-item" @click="handleUpdateImage">
     <el-icon><Refresh /></el-icon>
     <span>{{ $t('cloudMachine.updateImage') }}</span>
   </div>
   <div class="context-menu-item" @click="showApiDetails">
     <el-icon><InfoFilled /></el-icon>
     <span>{{ $t('cloudMachine.apiDetails') }}</span>
   </div>
   <div class="context-menu-item" @click="handleRename">
     <el-icon><Edit /></el-icon>
     <span>{{ $t('cloudMachine.renameDevice') }}</span>
   </div>
   <div class="context-menu-item" @click="handleShake">
     <el-icon><Edit /></el-icon>
     <span>{{ $t('cloudMachine.shake') }}</span>
   </div>
   <div class="context-menu-item" @click="handleGPS">
     <el-icon><Edit /></el-icon>
     <span>{{ $t('cloudMachine.setGPS') }}</span>
   </div>
   <div class="context-menu-item" @click="handleUploadGoogleCert">
     <el-icon><Edit /></el-icon>
     <span>{{ $t('cloudMachine.uploadGoogleCert') }}</span>
   </div>
   <div class="context-menu-item" @click="handleRestart">
     <el-icon><Refresh /></el-icon>
     <span>{{ $t('cloudMachine.restart') }}</span>
   </div>
   <div class="context-menu-item" @click="handleDelete">
     <el-icon><Delete /></el-icon>
     <span>{{ $t('common.delete') }}</span>
   </div>
   <div class="context-menu-item" @click="handleShutdown">
     <el-icon><Close /></el-icon>
     <span>{{ $t('cloudMachine.shutdown') }}</span>
   </div>
   <div class="context-menu-item" @click="setS5Agent">
     <el-icon><Close /></el-icon>
     <span>{{ $t('cloudMachine.setS5Agent') }}</span>
   </div>
   <div class="context-menu-item" @click="closeS5Agent">
     <el-icon><Close /></el-icon>
     <span>{{ $t('cloudMachine.closeS5Agent') }}</span>
   </div>
   <div class="context-menu-item" @click="handleSwitchBackup">
     <el-icon><Switch /></el-icon>
     <span>{{ $t('cloudMachine.switchBackup') }}</span>
   </div>
   <div class="context-menu-item" @click="handleFileUpload">
     <el-icon><Upload /></el-icon>
     <span>{{ $t('cloudMachine.fileUpload') }}</span>
   </div>
   <div v-if="getCurrentContextMenuContainer()?.androidType === 'V2'" class="context-menu-item" @click="handleOneKeyNewDevice">
     <el-icon><Switch /></el-icon>
     <span>{{ $t('cloudMachine.oneKeyNewDevice') }}</span>
   </div>
   <div v-else class="context-menu-item" @click="handleSwitchModel">
     <el-icon><Switch /></el-icon>
     <span>{{ $t('cloudMachine.switchModel') }}</span>
   </div>
   <div class="context-menu-item" @click="handleResetContainer">
     <el-icon><Refresh /></el-icon>
     <span>{{ $t('cloudMachine.resetContainer') }}</span>
   </div>
  </div>
  
  <!-- 文件上传输入 -->
  <input
    ref="fileInput"
    type="file"
    style="display: none"
    @change="handleFileSelect"
  />

  
  <!-- 密码设置对话框 -->
  <el-dialog
    v-model="passwordDialogVisible"
    :title="$t('common.setDevicePassword')"
    width="400px"
  >
    <el-form :model="passwordForm" label-width="80px">
      <el-form-item label="密码">
        <el-input
          v-model="passwordForm.password"
          type="password"
          placeholder="请输入密码"
          show-password
        ></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSetPassword" :loading="passwordLoading">
          {{ passwordLoading ? '设置中...' : '确定' }}
        </el-button>
      </span>
    </template>
  </el-dialog>


  <!-- 授权同步对话框 -->
  <el-dialog
    v-model="syncAuthDialogVisible"
    :title="$t('common.syncAuthLogin')"
    width="400px"
  >
    <el-form :model="syncAuthForm" label-width="80px">
      <el-form-item :label="$t('common.username')" required>
        <el-input
          v-model="syncAuthForm.username"
          :placeholder="$t('common.enterUsername')"
          autocomplete="off"
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.password')" required>
        <el-input
          v-model="syncAuthForm.password"
          type="password"
          :placeholder="$t('common.enterPassword')"
          show-password
        ></el-input>
      </el-form-item>
      <el-form-item>
        <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">
          <el-checkbox v-model="syncAuthForm.saveCredentials">{{ $t('common.rememberCredentials') }}</el-checkbox>
          <el-link type="primary" :underline="false" @click="openForgotPasswordDialog" style="font-size: 13px;">忘记密码</el-link>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleSyncAuthCancel">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSyncAuthSubmit" :loading="syncAuthLoading">
          {{ syncAuthLoading ? $t('common.loggingIn') : $t('common.login') }}
        </el-button>
        <el-button type="success" @click="openRegisterDialog">{{ $t('common.register') }}</el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 注册对话框 -->
  <el-dialog
    v-model="registerDialogVisible"
    :title="$t('common.userRegistration')"
    width="400px"
  >
    <el-form :model="registerForm" label-width="100px">
      <el-form-item :label="$t('common.phoneNumber')" required>
        <el-input
          v-model="registerForm.phone"
          :placeholder="$t('common.enterPhoneNumber')"
          autocomplete="off"
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.loginPassword')" required>
        <el-input
          v-model="registerForm.password"
          type="password"
          :placeholder="$t('common.enterLoginPassword')"
          show-password
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.confirmPassword')" required>
        <el-input
          v-model="registerForm.confirmPassword"
          type="password"
          :placeholder="$t('common.enterPasswordAgain')"
          show-password
        ></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.phoneVerificationCode')" required>
        <div style="display: flex; gap: 10px;">
          <el-input
            v-model="registerForm.vcode"
            :placeholder="$t('common.enterVerificationCode')"
            style="flex: 1;"
          ></el-input>
          <el-button 
            @click="sendVcode" 
            :loading="sendVcodeLoading"
            :disabled="vcodeCountdown > 0"
            style="width: 120px;"
          >
            {{ vcodeCountdown > 0 ? `${vcodeCountdown}${$t('common.retryAfterSeconds')}` : $t('common.getVerificationCode') }}
          </el-button>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleRegisterCancel">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleRegisterSubmit" :loading="registerLoading">
          {{ registerLoading ? $t('common.registering') : $t('common.register') }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 忘记密码对话框 -->
  <el-dialog
    v-model="forgotPasswordDialogVisible"
    title="忘记密码"
    width="400px"
    @close="handleForgotPasswordClose"
  >
    <el-form :model="forgotPasswordForm" label-width="0">
      <el-form-item>
        <el-input
          v-model="forgotPasswordForm.phone"
          placeholder="手机号码"
          autocomplete="off"
          clearable
        ></el-input>
        <div v-if="forgotPasswordErrors.phone" class="fp-error-app">
          <el-icon style="margin-right:3px;"><WarningFilled /></el-icon>{{ forgotPasswordErrors.phone }}
        </div>
      </el-form-item>
      <el-form-item>
        <el-input
          v-model="forgotPasswordForm.newPassword"
          type="password"
          placeholder="新密码"
          show-password
          clearable
        ></el-input>
        <div v-if="forgotPasswordErrors.newPassword" class="fp-error-app">
          <el-icon style="margin-right:3px;"><WarningFilled /></el-icon>{{ forgotPasswordErrors.newPassword }}
        </div>
      </el-form-item>
      <el-form-item>
        <el-input
          v-model="forgotPasswordForm.confirmPassword"
          type="password"
          placeholder="确认新密码"
          show-password
          clearable
        ></el-input>
        <div v-if="forgotPasswordErrors.confirmPassword" class="fp-error-app">
          <el-icon style="margin-right:3px;"><WarningFilled /></el-icon>{{ forgotPasswordErrors.confirmPassword }}
        </div>
      </el-form-item>
      <el-form-item>
        <div style="display: flex; gap: 10px; width: 100%;">
          <el-input
            v-model="forgotPasswordForm.vcode"
            placeholder="手机验证码"
            style="flex: 1;"
            clearable
          ></el-input>
          <el-button
            type="primary"
            @click="sendForgotPasswordVcode"
            :loading="fpVcodeLoading"
            :disabled="fpIsCountingDown"
            style="white-space: nowrap;"
          >
            {{ fpVcodeButtonText }}
          </el-button>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <div style="width: 100%; display: flex; flex-direction: column;">
        <el-button
          type="primary"
          style="width: 100%; margin-bottom: 10px;"
          :loading="forgotPasswordLoading"
          @click="handleForgotPasswordSubmit"
        >
          {{ forgotPasswordLoading ? '重置中...' : '重置密码' }}
        </el-button>
        <div style="text-align: center; font-size: 13px; color: #666;">
          还没有账号？<el-link type="primary" :underline="false" @click="openRegisterFromForgot">立即注册</el-link>
        </div>
      </div>
    </template>
  </el-dialog>

  <!-- 共享文件选择对话框 -->
  <el-dialog
    v-model="sharedFilesDialogVisible"
    title="选择要上传的文件"
    width="600px"
  >
    <div v-loading="filesLoading" element-loading-text="加载文件中...">
      <div class="dialog-header" style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <h3 style="margin: 0;">文件列表</h3>
          <el-button 
            type="text" 
            size="small" 
            @click="changeFileSort('name')"
            :class="{ 'sort-active': fileSortType === 'name' }"
          >
            名称 {{ fileSortType === 'name' ? (fileSortOrder === 'asc' ? '↑' : '↓') : '' }}
          </el-button>
          <el-button 
            type="text" 
            size="small" 
            @click="changeFileSort('time')"
            :class="{ 'sort-active': fileSortType === 'time' }"
          >
            时间 {{ fileSortType === 'time' ? (fileSortOrder === 'asc' ? '↑' : '↓') : '' }}
          </el-button>
        </div>
        <el-button type="primary" size="small" @click="openSharedDirectory">
          打开共享目录
        </el-button>
      </div>

      <!-- 共享目录路径设置 -->
      <div style="margin-bottom: 12px; padding: 10px 12px; background: #f5f7fa; border-radius: 4px; border: 1px solid #e4e7ed;">
        <div style="font-size: 12px; color: #606266; margin-bottom: 8px; font-weight: 500;">📂 文件来源目录</div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <el-input
            v-model="singleUploadSharedDirInfo.path"
            placeholder="共享目录路径"
            size="small"
            :readonly="true"
            style="flex: 1;"
          />
          <el-button size="small" type="primary" :loading="singleUploadSharedDirLoading" @click="handleSelectSingleUploadSharedDir">
            浏览
          </el-button>
        </div>
        <div style="display: flex; align-items: center; gap: 8px; margin-top: 6px;">
          <el-tag v-if="singleUploadSharedDirInfo.isDefault" type="info" size="small">默认目录</el-tag>
          <el-tag v-else type="success" size="small">自定义目录</el-tag>
          <el-button
            v-if="!singleUploadSharedDirInfo.isDefault"
            type="text"
            size="small"
            style="color: #909399; padding: 0;"
            :loading="singleUploadSharedDirLoading"
            @click="handleResetSingleUploadSharedDir"
          >
            恢复默认
          </el-button>
        </div>
      </div>

      <div v-if="sharedFileTree" class="file-tree">
        <!-- 递归渲染目录树 -->
        <div class="tree-node">
          <div class="node-checkbox">
            <el-checkbox
              :model-value="isSharedDirectoryFullySelected(sharedFileTree)"
              :indeterminate="isSharedDirectoryPartiallySelected(sharedFileTree)"
              @change="() => handleSharedNodeSelectionChange(sharedFileTree)"
            ></el-checkbox>
          </div>
          <span class="node-icon" v-if="sharedFileTree.isDir && sharedFileTree.children && sharedFileTree.children.length > 0" @click.stop="toggleNodeExpanded(sharedFileTree)">
            {{ sharedFileTree.expanded ? '▼' : '▶' }}
          </span>
          <span class="node-icon" v-else-if="sharedFileTree.isDir">
            📁
          </span>
          <span class="node-name" @click="toggleNodeExpanded(sharedFileTree)">{{ sharedFileTree.name }}</span>
        </div>
        <!-- 递归渲染子节点 -->
        <template v-if="sharedFileTree.expanded && sharedFileTree.isDir && sharedFileTree.children && sharedFileTree.children.length > 0">
          <div class="tree-children">
            <div v-for="child in sharedFileTree.children" :key="child.path">
              <div class="tree-node">
                <div class="node-checkbox" v-if="child.isDir">
                  <el-checkbox
                    :model-value="isSharedDirectoryFullySelected(child)"
                    :indeterminate="isSharedDirectoryPartiallySelected(child)"
                    @change="() => handleSharedNodeSelectionChange(child)"
                  ></el-checkbox>
                </div>
                <div class="node-checkbox" v-else>
                  <el-checkbox
                    :model-value="selectedFiles.includes(child.path)"
                    @change="(val) => handleSharedFileCheckChange(child.path, val)"
                  >{{ child.name }}</el-checkbox>
                </div>
                <span class="node-icon" v-if="child.isDir && child.children && child.children.length > 0" @click.stop="toggleNodeExpanded(child)">
                  {{ child.expanded ? '▼' : '▶' }}
                </span>
                <span class="node-icon" v-else-if="child.isDir">
                  📁
                </span>
                <span class="node-name" v-if="child.isDir" @click="toggleNodeExpanded(child)">{{ child.name }}</span>
                <span class="node-size" v-if="!child.isDir">{{ (child.size / 1024).toFixed(2) }} KB</span>
                <span class="node-date" v-if="!child.isDir">{{ new Date(child.modTime * 1000).toLocaleString() }}</span>

              </div>
              <!-- 递归渲染子目录 -->
              <template v-if="child.expanded && child.isDir && child.children && child.children.length > 0">
                <div class="tree-children">
                  <div v-for="grandchild in child.children" :key="grandchild.path">
                    <div class="tree-node">
                      <div class="node-checkbox" v-if="grandchild.isDir">
                        <el-checkbox
                          :model-value="isSharedDirectoryFullySelected(grandchild)"
                          :indeterminate="isSharedDirectoryPartiallySelected(grandchild)"
                          @change="() => handleSharedNodeSelectionChange(grandchild)"
                        ></el-checkbox>
                      </div>
                      <div class="node-checkbox" v-else>
                        <el-checkbox
                          :model-value="selectedFiles.includes(grandchild.path)"
                          @change="(val) => handleSharedFileCheckChange(grandchild.path, val)"
                        >{{ grandchild.name }}</el-checkbox>
                      </div>
                      <span class="node-icon" v-if="grandchild.isDir && grandchild.children && grandchild.children.length > 0" @click.stop="toggleNodeExpanded(grandchild)">
                        {{ grandchild.expanded ? '▼' : '▶' }}
                      </span>
                      <span class="node-icon" v-else-if="grandchild.isDir">
                        📁
                      </span>
                      <span class="node-name" v-if="grandchild.isDir" @click="toggleNodeExpanded(grandchild)">{{ grandchild.name }}</span>
                      <span class="node-size" v-if="!grandchild.isDir">{{ (grandchild.size / 1024).toFixed(2) }} KB</span>
                      <span class="node-date" v-if="!grandchild.isDir">{{ new Date(grandchild.modTime * 1000).toLocaleString() }}</span>

                    </div>
                    <!-- 递归渲染更深层次的子目录 -->
                    <template v-if="grandchild.expanded && grandchild.isDir && grandchild.children && grandchild.children.length > 0">
                      <div class="tree-children">
                        <div v-for="greatgrandchild in grandchild.children" :key="greatgrandchild.path">
                          <div class="tree-node">
                            <div class="node-checkbox" v-if="greatgrandchild.isDir">
                              <el-checkbox
                                :model-value="isSharedDirectoryFullySelected(greatgrandchild)"
                                :indeterminate="isSharedDirectoryPartiallySelected(greatgrandchild)"
                                @change="() => handleSharedNodeSelectionChange(greatgrandchild)"
                              ></el-checkbox>
                            </div>
                            <div class="node-checkbox" v-else>
                              <el-checkbox
                                :model-value="selectedFiles.includes(greatgrandchild.path)"
                                @change="(val) => handleSharedFileCheckChange(greatgrandchild.path, val)"
                              >{{ greatgrandchild.name }}</el-checkbox>
                            </div>
                            <span class="node-icon" v-if="greatgrandchild.isDir && greatgrandchild.children && greatgrandchild.children.length > 0" @click.stop="toggleNodeExpanded(greatgrandchild)">
                              {{ greatgrandchild.expanded ? '▼' : '▶' }}
                            </span>
                            <span class="node-icon" v-else-if="greatgrandchild.isDir">
                              📁
                            </span>
                            <span class="node-name" v-if="greatgrandchild.isDir" @click="toggleNodeExpanded(greatgrandchild)">{{ greatgrandchild.name }}</span>
                            <span class="node-size" v-if="!greatgrandchild.isDir">{{ (greatgrandchild.size / 1024).toFixed(2) }} KB</span>
                            <span class="node-date" v-if="!greatgrandchild.isDir">{{ new Date(greatgrandchild.modTime * 1000).toLocaleString() }}</span>

                          </div>
                          <!-- 递归渲染更深层次的子目录 -->
                          <template v-if="greatgrandchild.expanded && greatgrandchild.isDir && greatgrandchild.children && greatgrandchild.children.length > 0">
                            <div class="tree-children">
                              <div v-for="deepchild in greatgrandchild.children" :key="deepchild.path">
                                <div class="tree-node">
                                  <div class="node-checkbox" v-if="deepchild.isDir">
                                    <el-checkbox
                                      :model-value="isSharedDirectoryFullySelected(deepchild)"
                                      :indeterminate="isSharedDirectoryPartiallySelected(deepchild)"
                                      @change="() => handleSharedNodeSelectionChange(deepchild)"
                                    ></el-checkbox>
                                  </div>
                                  <div class="node-checkbox" v-else>
                                    <el-checkbox
                                      :model-value="selectedFiles.includes(deepchild.path)"
                                      @change="(val) => handleSharedFileCheckChange(deepchild.path, val)"
                                    >{{ deepchild.name }}</el-checkbox>
                                  </div>
                                  <span class="node-icon" v-if="deepchild.isDir">
                                    📁
                                  </span>
                                  <span class="node-name" v-if="deepchild.isDir">{{ deepchild.name }}</span>
                                  <span class="node-size" v-if="!deepchild.isDir">{{ (deepchild.size / 1024).toFixed(2) }} KB</span>
                                  <span class="node-date" v-if="!deepchild.isDir">{{ new Date(deepchild.modTime * 1000).toLocaleString() }}</span>
                                </div>
                              </div>
                            </div>
                          </template>
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </template>
      </div>
      <div v-else class="no-files">
        <el-empty description="共享目录中没有文件"></el-empty>
        <el-button type="primary" style="margin-top: 16px;" @click="openSharedDirectory">
          打开共享目录
        </el-button>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="sharedFilesDialogVisible = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleUploadToCloudMachine" 
          :loading="uploadLoading"
          :disabled="selectedFiles.length === 0"
        >
          {{ uploadLoading ? '上传中...' : '上传' }}
        </el-button>
         <el-button 
          type="primary" 
          @click="handleUploadRefresh" 
        >
          刷新共享文件
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 批量认证对话框 -->
  <el-dialog
    v-model="batchAuthDialogVisible"
    title="设备批量认证"
    width="600px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
  >
    <el-alert
      :title="`需要对 ${batchAuthDevices.length} 个设备进行认证`"
      type="warning"
      :closable="false"
      style="margin-bottom: 16px;"
    >
      <template #default>
        <div style="font-size: 13px;">
          请为以下设备输入认证密码
        </div>
      </template>
    </el-alert>
    
    <!-- 设备列表 -->
    <div style="max-height: 400px; overflow-y: auto;">
      <el-form label-width="100px">
        <div 
          v-for="(item, index) in batchAuthDevices" 
          :key="item.device.ip"
          style="padding: 12px; margin-bottom: 12px; border: 1px solid #dcdfe6; border-radius: 4px;"
          :style="{
            borderColor: item.status === 'success' ? '#67c23a' : item.status === 'failed' ? '#f56c6c' : '#dcdfe6',
            backgroundColor: item.status === 'success' ? '#f0f9ff' : item.status === 'failed' ? '#fef0f0' : '#fff'
          }"
        >
          <!-- 设备标题 -->
          <div style="display: flex; align-items: center; margin-bottom: 8px;">
            <span style="font-weight: bold; font-size: 14px;">设备 {{ index + 1 }}: {{ item.device.ip }}</span>
            <el-tag 
              v-if="item.status === 'verifying'" 
              type="info" 
              size="small" 
              style="margin-left: 8px;"
            >
              验证中...
            </el-tag>
            <el-tag 
              v-else-if="item.status === 'success'" 
              type="success" 
              size="small" 
              style="margin-left: 8px;"
            >
              ✓ 认证成功
            </el-tag>
            <el-tag 
              v-else-if="item.status === 'failed'" 
              type="danger" 
              size="small" 
              style="margin-left: 8px;"
            >
              ✗ 认证失败
            </el-tag>
          </div>
          
          <!-- 密码输入 -->
          <el-form-item label="密码" :required="true" style="margin-bottom: 8px;">
            <el-input
              v-model="item.password"
              type="password"
              placeholder="请输入设备密码"
              show-password
              :disabled="item.status === 'verifying' || item.status === 'success'"
              @keyup.enter="handleBatchAuthSubmit"
            >
              <template #append v-if="item.status === 'success'">
                <el-icon color="#67c23a"><CircleCheck /></el-icon>
              </template>
              <template #append v-else-if="item.status === 'failed'">
                <el-icon color="#f56c6c"><CircleClose /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          
          <!-- 错误提示 -->
          <div v-if="item.status === 'failed' && item.errorMsg" style="color: #f56c6c; font-size: 12px; margin-top: 4px;">
            {{ item.errorMsg }}
          </div>
          
          <!-- 保存密码选项 -->
          <el-form-item style="margin-bottom: 0;">
            <el-checkbox 
              v-model="item.savePassword"
              :disabled="item.status === 'verifying' || item.status === 'success'"
            >
              自动保存密码
            </el-checkbox>
          </el-form-item>
        </div>
      </el-form>
    </div>
    
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleBatchAuthCancel" :disabled="batchAuthLoading">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleBatchAuthSubmit" 
          :loading="batchAuthLoading"
          :disabled="batchAuthDevices.every(item => item.status === 'success')"
        >
          {{ batchAuthLoading ? '认证中...' : '确定' }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 机型切换对话框 -->
  <el-dialog
    v-model="switchModelDialogVisible"
    :title="`切换机型 - ${currentSwitchContainer?.name || ''}`"
    width="500px"
  >
    <div style="padding: 10px 0;">
      <el-form label-width="80px">
        <el-form-item label="机型来源">
          <el-radio-group v-model="switchModelType" @change="handleSwitchModelTypeChange">
            <el-radio-button label="online">线上机型</el-radio-button>
            <el-radio-button label="local">本地机型</el-radio-button>
            <el-radio-button label="backup">备份机型</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择机型">
          <el-select v-model="tempModelId" style="width: 100%;" placeholder="请选择要切换的机型" filterable :loading="fetchingModels || fetchingBackupModels">
            <el-option 
              v-for="model in displayedModels" 
              :key="model.id" 
              :label="model.name" 
              :value="model.id"
            ></el-option>
          </el-select>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="cancelSwitchModel" :disabled="switchingModel">取消</el-button>
        <el-button type="primary" @click="confirmSwitchModel" :loading="switchingModel" :disabled="switchingModel">
          {{ switchingModel ? '正在切换' : '确定' }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 设置推流弹窗 -->
  <el-dialog
    v-model="setStreamDialogVisible"
    title="设置推流"
    width="50%"
    :close-on-click-modal="false"
  >
    <el-form >
      <el-form-item label="推流类型">
        <el-select v-model="streamType" style="width: 100%;" placeholder="请选择推流类型">
          <el-option label="图片" value="image"></el-option>
          <el-option label="视频" value="video"></el-option>
          <el-option label="APP" value="app"></el-option>
          <!-- <el-option label="RTMP" value="rtmp"></el-option> -->
        </el-select>
      </el-form-item>
      
      <el-form-item v-if="streamType === 'image' || streamType === 'video'" label="文件路径">
        <el-input v-model="streamFilePath" placeholder="请选择文件" readonly>
          <template #append>
            <el-button @click="selectStreamFolder">选择文件</el-button>
          </template>
        </el-input>
        <p style="margin-top: 20px;color: red;">选择图片或视频会自动推送到设备内</p>
      </el-form-item>
      
      <el-form-item v-if="streamType === 'app'">
        <div class="qrcode-container" v-loading="qrCodeLoading" style="display: flex; flex-direction: column; align-items: center; width: 100%;">
          <h4>扫码连接</h4>
          <img v-if="qrCodeUrl" :src="qrCodeUrl" alt="连接二维码" style="width: 200px; height: 200px;" />
          <div v-else-if="!qrCodeLoading">二维码生成失败</div>
          
          <div style="margin-top: 10px;">
            <el-popover
              placement="bottom"
              :width="200"
              trigger="hover"
            >
              <template #reference>
                <el-button link type="primary">APP下载地址</el-button>
              </template>
              <div style="text-align: center;">
                <img v-if="appDownloadQrCodeUrl" :src="appDownloadQrCodeUrl" style="width: 150px; height: 150px;" />
                <div v-else>生成中...</div>
                <div style="font-size: 12px; margin-top: 5px;">扫码下载APP</div>
              </div>
            </el-popover>
          </div>
          <div style="margin-top: 10px;color: red;">
            <p>注意：推流手机必须与设备同在一个局域网内>否则无法连接</p>
            <p>使用方法：安装APP后扫码增加云机，如出现相机黑屏，请如下操作</p>
            <p>扩展服务>设置摄像头视频源>手机摄像头映射>保存</p>
          </div>
        </div>
      </el-form-item>
      
      <!-- <el-form-item v-if="streamType === 'rtmp'" label="RTMP地址">
        <el-input v-model="rtmpUrl" placeholder="请输入RTMP推流地址，如 rtmp://example.com/live/stream"></el-input>
      </el-form-item> -->
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="cancelSetStream">取消</el-button>
        <el-button type="primary" @click="confirmSetStream" :loading="setStreamLoading">确定</el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 设备详情弹窗 -->
  <el-dialog
    v-model="deviceDetailsDialogVisible"
    :title="$t('common.deviceDetails')"
    :before-close="handleDeviceDetailsDialogClose"
    center
    :close-on-click-modal="true"
    :close-on-press-escape="true"
    class="device-details-dialog"
    style="--el-dialog-width: 1300px;"
    top="6vh"
  >
    <div v-if="activeDevice" class="device-details-content">
      <!-- 悬浮功能栏 -->
      <div class="floating-toolbar">
        <div class="device-info">
          <span class="device-ip-text">{{ activeDevice?.ip || $t('common.noDeviceSelected') }}</span>
          <el-button 
            type="primary" 
            size="small" 
            @click="activeDevice && (fetchAndroidContainers(activeDevice, true), fetchDeviceDetailCloudMachines())"
            :disabled="!activeDevice || loading"
            class="refresh-button"
          >
            <el-icon :class="{ 'is-rotating': loading }"><Refresh /></el-icon> {{ $t('common.refreshCloudMachines') }}
          </el-button>
          <span class="divider">|</span>
          <el-button 
            :type="currentRightTab === 'instance' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'instance' }"
            @click="currentRightTab = 'instance'"
          >{{ $t('common.instance') }}</el-button>
          <!-- <el-button 
            :type="currentRightTab === 'image' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'image' }"
            @click="currentRightTab = 'image'"
          >{{ $t('common.image') }}</el-button> -->
          <!-- <el-button 
            :type="currentRightTab === 'network' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'network' }"
            @click="currentRightTab = 'network'"
          >{{ $t('common.network') }}</el-button> -->
          <el-button 
            :type="currentRightTab === 'host' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'host' }"
            @click="currentRightTab = 'host'"
          >{{ $t('common.host') }}</el-button>
        </div>
        
        <!-- 批量操作按钮 -->
        <el-space wrap class="batch-actions">
          <template v-if="currentRightTab === 'instance'">
            <!-- <el-button @click="handleBatchAction('restart')" size="small">批量重启</el-button>
            <el-button @click="handleBatchAction('reset')" size="small">批量重置</el-button>
            <el-button @click="handleBatchAction('start')" size="small">批量启动</el-button>
            <el-button @click="handleBatchAction('shutdown')" size="small">批量关机</el-button> -->
            <!-- <el-button type="danger" @click="handleBatchAction('delete')" size="small">批量删除</el-button> -->
            <!-- <el-button type="primary" @click="handleBatchAction('new')" size="small">批量新机</el-button> -->
          </template>
          <el-button 
            v-if="currentRightTab === 'image'"
            type="info" 
            size="small" 
            @click="refreshImageList" 
            :disabled="!activeDevice || fetchingImages"
            class="refresh-button"
          >
            <el-icon :class="{ 'is-rotating': fetchingImages }"><Refresh /></el-icon> {{ $t('common.refreshImages') }}
          </el-button>
          <el-button 
            v-if="currentRightTab === 'network'"
            type="info" 
            size="small" 
            @click="activeDevice && fetchDockerNetworks(activeDevice)" 
            :disabled="!activeDevice || dockerNetworksLoading"
            class="refresh-button"
          >
            <el-icon :class="{ 'is-rotating': dockerNetworksLoading }"><Refresh /></el-icon> {{ $t('common.refreshNetworkList') }}
          </el-button>
          <!-- 🔧 移除刷新主机信息按钮，使用心跳机制自动更新 -->
        </el-space>
      </div>
      
      <!-- 右侧内容区域，根据标签页切换显示不同内容 -->
      <!-- 实例标签页 -->
      <div v-if="currentRightTab === 'instance'" class="table-container" style="padding-bottom: 20px;">
        <el-table 
          :data="deviceDetailGroupedInstances" 
          stripe 
          size="small" 
          class="slot-table"
          :tree-props="{ children: '_children', hasChildren: 'hasChildren' }"
          row-key="id"
        >
          <el-table-column :label="$t('common.slot')" width="140" align="center">
            <template #default="scope">
              <div class="slot-cell-content">
                <span class="slot-number">
                  <!-- <el-button 
                    v-if="scope.row.showFoldButton"
                    size="mini" 
                    type="text"
                    class="fold-button"
                  >
                  </el-button> -->
                  {{ scope.row.slotNum }}
                </span>
                <el-button 
                  v-if="scope.row.isFirstInSlot"
                  size="mini" 
                  type="primary" 
                  @click="showCreateDialog(activeDevice, 'slot', scope.row.slotNum)"
                  class="create-button"
                  :disabled="!activeDevice"
                >
                  {{ $t('common.create') }}
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.instanceName')" width="110">
            <template #default="scope">
              {{ formatInstanceName(scope.row.name) }}
            </template>
          </el-table-column>
          <el-table-column prop="ip" :label="$t('common.ipAddress')" width="130"></el-table-column>
          <el-table-column :label="$t('common.systemImage')" width="200" align="center">
            <template #default="scope">
              <div 
                style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                :title="getImageDisplayName(scope.row.image)"
              >
                {{ getImageDisplayName(scope.row.image) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.createTime')" width="160">
            <template #default="scope">
              {{ scope.row.created ? new Date(scope.row.created).toLocaleString('zh-CN') : scope.row.createTime }}
            </template>
          </el-table-column>
          <el-table-column prop="status" :label="$t('common.status')" width="100" align="center">
            <template #default="scope">
              <el-tag 
                :type="scope.row.status === 'running' ? 'success' : scope.row.status === 'restarting' ? 'warning' : 'info'"
                size="small"
                class="status-tag-normal"
              >
                {{ scope.row.status === 'running' ? $t('common.running') : (scope.row.status === 'shutdown' || scope.row.status === 'exited') ? $t('common.shutdown') : scope.row.status === 'created' ? $t('common.created') : $t('common.restarting') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="modelName" :label="$t('common.model')" width="140">
              <template #default="scope">
              {{ formatInstanceModel(scope.row.modelPath) || $t('common.none') }}
            </template>
          </el-table-column>
          <el-table-column :label="$t('common.operation')" width="200" fixed="right" align="center">
            <template #default="scope">
              <el-space size="mini" wrap>
                <el-button 
                  size="mini" 
                  :type="scope.row.status === 'running' ? 'warning' : 'success'" 
                  @click="() => {
                    // 检查是否为空坑位（没有云机）
                    if (!scope.row.name || scope.row.name === '') {
                      ElMessage.warning($t('common.slotNoMachine').replace('{slot}', scope.row.slotNum));
                      return;
                    }
                    
                    const action = scope.row.status === 'running' ? $t('common.shutdown') : $t('common.startUp');
                    ElMessageBox.confirm($t('common.confirmAction').replace('{action}', action).replace('{name}', scope.row.name), $t('common.operationConfirm'), {
                      confirmButtonText: $t('common.confirm'),
                      cancelButtonText: $t('common.cancel'),
                      type: 'warning'
                    }).then(async () => {
                      try {
                        if (scope.row.status === 'running') {
                          await stopContainer(activeDevice, scope.row.name);
                        } else {
                          await startContainer(activeDevice, scope.row.name);
                        }
                        ElMessage.success($t('common.actionSuccess').replace('{action}', action));
                        await fetchDeviceDetailCloudMachines();
                      } catch (error) {
                        ElMessage.error($t('common.actionFailed').replace('{action}', action).replace('{error}', error.message));
                      }
                    }).catch(() => {});
                  }"
                >
                  {{ scope.row.status === 'running' ? $t('common.shutdown') : $t('common.startUp') }}
                </el-button>
                <el-button 
                  size="mini" 
                  type="danger"
                  @click="() => {
                    // 检查是否为空坑位（没有云机）
                    if (!scope.row.name || scope.row.name === '') {
                      ElMessage.warning($t('common.slotCannotDelete').replace('{slot}', scope.row.slotNum));
                      return;
                    }
                    handleDeviceDetailDeleteContainer(scope.row);
                  }"
                >
                  {{ $t('common.delete') }}
                </el-button>
                <template v-if="scope.row.status === 'running'">
                  <el-button 
                    size="mini" 
                    type="primary"
                    @click="() => { try { startProjection({ ip: scope.row.networkName == 'myt' ? scope.row.ip : activeDevice?.ip }, scope.row) } catch (error) { console.error('打开投屏失败:', error) } }"
                  >
                    {{ $t('common.openProjection') }}
                  </el-button>
                  <el-button 
                    size="mini" 
                    type="info"
                    @click="showUpdateImageDialog(scope.row)"
                  >
                    {{ $t('common.updateImage') }}
                  </el-button>
                </template>
              </el-space>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <!-- 主机标签页 -->
      <div v-if="currentRightTab === 'host'" class="host-info-container">
        <el-card shadow="hover" class="host-info-card">
          
          
          <div v-if="activeDevice" class="host-info-content">
            <!-- 基本信息 -->
            <el-descriptions :title="$t('common.basicInfo')" :column="2" border>
              <el-descriptions-item :label="$t('common.hostName')">{{ activeDevice.name }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.hostIP')">{{ activeDevice.ip }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.deviceID')">{{ (activeDevice.version === 'v3' && v3DeviceInfo.originalData?.deviceId) ? v3DeviceInfo.originalData.deviceId : activeDevice.id }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.deviceVersion')">{{ activeDevice.version }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.firmwareVersion')">{{ v3DeviceInfo.originalData?.version || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.modelVersion')">{{ v3DeviceInfo.originalData?.model || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.deviceUptime')">{{ v3DeviceUptimeMinutes }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.cpuTemp')">{{ v3DeviceInfo.originalData?.cputemp || $t('common.loading') }}°C</el-descriptions-item>
              <el-descriptions-item :label="$t('common.cpuLoad')">{{ v3DeviceInfo.originalData?.cpuload || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.memoryTotal')">{{ formatSize(v3DeviceInfo.originalData?.memtotal) }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.memoryUsed')">{{ formatSize(v3DeviceInfo.originalData?.memuse) }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.diskTotal')">{{ formatSize(v3DeviceInfo.originalData?.mmctotal) }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.diskUsed')">{{ formatSize(v3DeviceInfo.originalData?.mmcuse) }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.diskModel')">{{ v3DeviceInfo.originalData?.mmcmodel || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.diskTemp')">{{ v3DeviceInfo.originalData?.mmctemp || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.diskRead')">{{ v3DeviceInfo.originalData?.mmcread || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.diskWrite')">{{ v3DeviceInfo.originalData?.mmcwrite || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.networkIP')">{{ v3DeviceInfo.originalData?.ip || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.macAddress')">{{ v3DeviceInfo.originalData?.hwaddr || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.networkSpeed')">{{ v3DeviceInfo.originalData?.speed || $t('common.loading') }}</el-descriptions-item>
              <el-descriptions-item :label="$t('common.eth0Network')">{{ v3DeviceInfo.originalData?.netWork_eth0 || $t('common.loading') }}</el-descriptions-item>
            </el-descriptions>
            
            <!-- V3设备额外信息 -->
            <div v-if="activeDevice.version === 'v3'" class="v3-info-section">
          
              
              <!-- 密码管理功能 -->
              <div class="password-section" style="margin-top: 20px;">
                <el-divider>{{ $t('common.passwordManagement') }}</el-divider>
                <el-space>
                  <el-button 
                    type="primary" 
                    size="small" 
                    @click="showPasswordDialog"
                    :disabled="!activeDevice"
                  >
                    {{ $t('common.setPassword') }}
                  </el-button>
                  <el-button 
                    type="warning" 
                    size="small" 
                    @click="handleClosePassword"
                    :disabled="!activeDevice"
                  >
                    {{ $t('common.closePassword') }}
                  </el-button>
                  <span class="password-hint" style="color: #606266; font-size: 12px;">
                    {{ $t('common.passwordHint') }}
                  </span>
                </el-space>
              </div>

              <div style="text-align: end;">
                <el-button type="warning" size="small" @click="handleCleanDisk">
                  {{ $t('common.cleanDiskData') }}
                </el-button>
              </div>
              
              <!-- SDK升级功能 - 不在设备详情页面显示 -->
              <div v-if="showUpgradeButton && !isViewingDeviceDetails" class="upgrade-section">
                <el-divider>{{ $t('common.sdkUpgrade') }}</el-divider>
                <el-space>
                  <span>{{ $t('common.currentSDKVersion') }}: {{ v3LatestInfo.originalData?.currentVersion }}</span>
                  <span>{{ $t('common.latestSDKVersion') }}: {{ v3LatestInfo.originalData?.latestVersion }}</span>
                  <el-button 
                    type="primary" 
                    @click="upgradeSDK"
                    :loading="upgrading"
                    :disabled="upgrading"
                  >
                    <template v-if="upgrading && upgradeProgress > 0">
                      {{ $t('common.upgrading') }} ({{ upgradeProgress.toFixed(1) }}%)
                    </template>
                    <template v-else>
                      {{ $t('common.upgradeSDK') }}
                    </template>
                  </el-button>
                </el-space>
              </div>
            </div>
          </div>
          
          <div v-else class="no-device-selected">
            <el-empty :description="$t('common.selectDeviceFirst')" :image-size="100"></el-empty>
          </div>
        </el-card>
      </div>
      
      <!-- 网络标签页 -->
      <div v-if="currentRightTab === 'network'" class="network-info-container">
        <el-card shadow="hover" class="network-info-card">
          <template #header>
            <div class="card-header">
              <span>Docker网络列表</span>
              <el-space>
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="showAddMacvlanDialog()" 
                  :disabled="!activeDevice"
                >
                  <el-icon><Plus /></el-icon> 添加macvlan网络
                </el-button>
              </el-space>
            </div>
          </template>
          
          <div v-if="activeDevice" class="network-info-content">
            <!-- 加载状态 -->
            <div v-if="dockerNetworksLoading" class="network-loading">
              <el-skeleton :rows="5" animated></el-skeleton>
            </div>
            <!-- 错误状态 -->
            <div v-else-if="dockerNetworksError" class="network-error">
              <el-alert
                title="加载失败"
                :description="dockerNetworksError"
                type="error"
                show-icon
              ></el-alert>
            </div>
            
            <!-- 网络列表 -->
            <div v-else-if="dockerNetworks.length > 0" class="table-container">
              <el-table :data="dockerNetworks" stripe size="small" class="network-table">
                  <el-table-column prop="Name" label="网络名称" width="180"></el-table-column>
                  <el-table-column prop="ID" label="网络ID" width="120"></el-table-column>
                  <el-table-column label="网段" width="150">
                    <template #default="scope">
                      <span v-if="scope.row.IPAM?.Config?.[0]?.Subnet">{{ scope.row.IPAM.Config[0].Subnet }}</span>
                      <span v-else>无</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="Driver" label="驱动" width="100"></el-table-column>
                  <el-table-column prop="Scope" label="范围" width="100"></el-table-column>
                  <el-table-column label="网关" width="150">
                    <template #default="scope">
                      <span v-if="scope.row.IPAM?.Config?.[0]?.Gateway">{{ scope.row.IPAM.Config[0].Gateway }}</span>
                      <span v-else>无</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="IP范围" width="150">
                    <template #default="scope">
                      <span v-if="scope.row.IPAM?.Config?.[0]?.IPRange">{{ scope.row.IPAM.Config[0].IPRange }}</span>
                      <span v-else>无</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="Containers" label="容器数量" width="120">
                    <template #default="scope">
                      <span>{{ Object.keys(scope.row.Containers || {}).length }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="Internal" label="私有" width="80">
                    <template #default="scope">
                      <el-tag v-if="scope.row.Internal" type="danger" size="small">是</el-tag>
                      <el-tag v-else type="success" size="small">否</el-tag>
                    </template>
                  </el-table-column>
              </el-table>
            </div>
            <!-- 空状态 -->
            <div v-else class="network-empty">
              <el-empty description="暂无网络信息" :image-size="100"></el-empty>
            </div>
          </div>
          
          <div v-else class="no-device-selected">
            <el-empty description="请先选择设备" :image-size="100"></el-empty>
          </div>
        </el-card>
      </div>
      
      <!-- 镜像标签页 -->
      <div v-if="currentRightTab === 'image'" class="image-list-container">
        <div v-if="activeDevice" class="image-list-content">
          <div v-if="isLoadingBoxImages" class="image-loading">
            <el-skeleton :rows="5" animated></el-skeleton>
          </div>
          <div v-else-if="matchedBoxImages.length === 0" class="no-images">
            <el-empty description="暂无已下载镜像信息" :image-size="100"></el-empty>
          </div>
          <div v-else class="image-items">
            <div v-for="(image, index) in matchedBoxImages" :key="index" class="image-item">
              <el-card shadow="hover" class="image-card">
                <template #header>
                  <div class="card-header">
                    <span>{{ image.onlineImageName }}</span>
                    <el-button
                      type="danger"
                      size="small"
                      @click="handleDeleteImage(image)"
                      class="delete-button"
                    >
                    <el-icon><Delete /></el-icon> 删除
                    </el-button>
                  </div>
                </template>
                <div class="image-info">
                  <div class="image-details">
                    <p>线上镜像名称: {{ image.onlineImageName }}</p>
                    <p>镜像大小: {{ image.size }}</p>
                    <p>创建时间: {{ image.createTime }}</p>
                    <p v-if="image.matched" style="color: #67c23a;">✓ 已与线上镜像匹配</p>
                    <p v-else style="color: #909399;">✗ 未与线上镜像匹配</p>
                    <p v-if="image.matched" style="font-size: 12px; color: #909399;">线上URL: {{ image.onlineImageUrl }}</p>
                    <p v-else style="font-size: 12px; color: #909399;">设备中URL: {{ image.url }}</p>
                  </div>
                </div>
              </el-card>
            </div>
          </div>
        </div>
        
        <div v-else class="no-device-selected">
          <el-empty description="请先选择设备" :image-size="100"></el-empty>
        </div>
      </div>
    </div>
    
    <div v-else class="no-device-selected">
      <el-empty description="请先选择设备" :image-size="100"></el-empty>
    </div>
  </el-dialog>

  <!-- 终端相关HTML结构 -->
  <div id="login-overlay" style="display: none;">
    <div class="login-box">
      <h2 style="text-align:center">Docker Console</h2>
      <div class="form-group">
        <label>容器 ID 或 名称</label>
        <input type="text" id="containerId" placeholder="例如: T1001">
      </div>
      <div class="form-group">
        <label>Shell 程序</label>
        <select id="shell">
          <option value="/bin/sd">/bin/sd</option>
          <option value="/bin/bash">/bin/bash</option>
          <option value="/bin/sh">/bin/sh</option>
        </select>
      </div>
      <button class="btn" onclick="connectDocker()">连 接</button>
    </div>
  </div>

  <div id="terminal"></div>

  <!-- 批量切换机型对话框 -->
  <el-dialog
    v-model="batchSwitchModelDialogVisible"
    title="批量新机"
    width="900px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :before-close="handleBatchSwitchModelCancel"
  >
    <!-- 切换机型时的加载覆盖层 -->
    <div 
      v-if="batchSwitchingModel" 
      class="switching-model-overlay"
    >
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>切换机型中...</span>
    </div>
    
    <div class="batch-switch-model-content">
      <!-- 左侧：坑位列表 -->
      <div class="slots-section" style="width: 30%; float: left; margin-right: 20px;">
        <h3 style="margin-bottom: 10px;">可选坑位</h3>
        <div class="available-slots">
          <div 
            v-for="target in batchSwitchModelTargets" 
            :key="target.id || target.indexNum"
            class="slot-item"
            :class="{ 
              'dragging': draggingSlot === target.id || draggingSlot === target.indexNum,
              'v2-container': target.androidType === 'V2'
            }"
            :draggable="target.androidType !== 'V2'"
            @dragstart="handleSlotDragStart($event, target)"
            @dragover.prevent
            @dragenter.prevent
            @drop="handleSlotDropInAvailableArea($event, target)"
            :title="target.androidType === 'V2' ? 'V2容器只能使用随机机型' : ''"
          >
            <!-- {{ cloudManageMode === 'slot' ? `坑位${target.indexNum}` : `云机${target.ip}` }} -->
              {{ formatInstanceName(target.name) }}
              <span v-if="target.androidType === 'V2'" style="color: #F56C6C; font-size: 10px; margin-left: 4px;">(V2)</span>
          </div>
        </div>
      </div>
      
      <!-- 右侧：机型分配区域 -->
      <div class="models-section" style="width: 65%; float: right;">
        <div class="models-header" style="display: flex; justify-content: space-between; margin-bottom: 16px;">
          <h3>机型分配</h3>
          <div class="models-actions">
            <el-button 
              type="primary" 
              @click="addNewModelSlot('online')"
              :disabled="batchSwitchingModel"
            >
              添加线上
            </el-button>
            <el-button 
              type="primary" 
              @click="addNewModelSlot('local')"
              :disabled="batchSwitchingModel"
            >
              添加本地
            </el-button>
            <el-button 
              type="primary" 
              @click="addNewModelSlot('backup')"
              :disabled="batchSwitchingModel"
            >
              添加备份
            </el-button>
          </div>
        </div>
        
        <!-- 机型列表 -->
        <div class="model-slots-container">
          <!-- 确保至少显示一个机型分配槽 -->
          <div 
            v-for="(modelSlot, index) in modelSlots" 
            :key="index"
            class="model-slot"
            @dragover.prevent
            @dragenter.prevent
            @drop="handleSlotDropInModelSlot($event, index)"
          >
            <div class="model-slot-header">
              <div class="model-selector">
                <el-select 
                  v-model="modelSlot.modelId" 
                  :placeholder="modelSlot.type === 'local' ? '选择本地机型' : (modelSlot.type === 'backup' ? '选择备份机型' : '选择线上机型')" 
                  style="width: 200px;"
                  :disabled="batchSwitchingModel"
                >
                  <!-- 添加随机选项 -->
                  <el-option 
                    label="随机" 
                    value="random"
                  >
                    <div style="display: flex; justify-content: space-between;">
                      <span>随机</span>
                      <span style="color: #909399; font-size: 12px;">random</span>
                    </div>
                  </el-option>
                  
                  <!-- 线上机型 -->
                  <template v-if="!modelSlot.type || modelSlot.type === 'online'">
                    <el-option 
                      v-for="model in phoneModels" 
                      :key="model.id" 
                      :label="model.name" 
                      :value="model.id"
                    >
                      <div style="display: flex; justify-content: space-between;">
                        <span>{{ model.name }}</span>
                        <span style="color: #909399; font-size: 12px;">{{ model.id }}</span>
                      </div>
                    </el-option>
                  </template>

                  <!-- 本地机型 -->
                  <template v-else-if="modelSlot.type === 'local'">
                    <el-option 
                      v-for="model in localPhoneModels" 
                      :key="model.name" 
                      :label="model.name" 
                      :value="model.name"
                    >
                      <div style="display: flex; justify-content: space-between;">
                        <span>{{ model.name }}</span>
                      </div>
                    </el-option>
                  </template>

                  <!-- 备份机型 -->
                  <template v-else-if="modelSlot.type === 'backup'">
                    <el-option 
                      v-for="model in backupPhoneModels" 
                      :key="model.name" 
                      :label="model.name" 
                      :value="model.name"
                    >
                      <div style="display: flex; justify-content: space-between;">
                        <span>{{ model.name }}</span>
                      </div>
                    </el-option>
                  </template>
                </el-select>
                
                <!-- 显示机型类型标签 -->
                <el-tag 
                  size="small" 
                  :type="!modelSlot.type || modelSlot.type === 'online' ? 'primary' : (modelSlot.type === 'local' ? 'success' : 'warning')"
                  style="margin-left: 10px;"
                >
                  {{ !modelSlot.type || modelSlot.type === 'online' ? '线上机型' : (modelSlot.type === 'local' ? '本地机型' : '备份机型') }}
                </el-tag>
              </div>
              <div class="model-actions">
                <el-button 
                  type="danger" 
                  size="mini" 
                  @click="removeModelSlot(index)"
                  :disabled="batchSwitchingModel || modelSlots.length <= 1"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
            
            <div class="assigned-slots">
              <div 
                v-for="assignedSlot in modelSlot.assignedSlots" 
                :key="assignedSlot.id || assignedSlot.indexNum"
                class="assigned-slot-item"
                :class="{ 'v2-container': assignedSlot.androidType === 'V2' }"
                :draggable="assignedSlot.androidType !== 'V2'"
                @dragstart="handleSlotDragStart($event, assignedSlot)"
                :title="assignedSlot.androidType === 'V2' ? 'V2容器只能使用随机机型' : ''"
              >
                <!-- {{ cloudManageMode === 'slot' ? `坑位${assignedSlot.indexNum}` : `云机${assignedSlot.ip}` }} -->
                  {{ formatInstanceName(assignedSlot.name) }}
                  <span v-if="assignedSlot.androidType === 'V2'" style="color: #F56C6C; font-size: 10px; margin-left: 4px;">(V2)</span>
              </div>
              <div v-if="modelSlot.assignedSlots.length === 0" class="no-slots-assigned">
                拖放坑位到此处
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleBatchSwitchModelCancel" :disabled="batchSwitchingModel">
          取消
        </el-button>
        <el-button type="primary" @click="confirmBatchSwitchModel" :disabled="batchSwitchingModel || !isModelSlotsValid">
          确定
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 更新提示弹窗 -->
  <UpdateDialog
    v-model:visible="updateDialogVisible"
    :update-info="updateInfo"
  />

  <!-- 设置MacVlanIP对话框 -->
  <el-dialog
    v-model="macVlanDialogVisible"
    title="设置MacVlanIP"
    width="400px"
  >
    <el-form :model="macVlanForm" label-width="100px">
      <el-form-item label="容器名称">
        <el-input v-model="macVlanForm.name" disabled></el-input>
      </el-form-item>
      <el-form-item label="MacVlanIP" required>
        <el-input v-model="macVlanForm.ip" placeholder="请输入MacVlanIP"></el-input>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="macVlanDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmSetMacVlanIP" :loading="macVlanLoading">
          {{ $t('common.confirm') }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- GPS定位设置弹窗 -->
  <el-dialog
    v-model="gpsDialogVisible"
    :title="$t('common.setIPLocation')"
    width="400px"
    :close-on-click-modal="false"
  >
    <el-form :model="gpsForm" label-width="80px">
      <el-form-item :label="$t('common.locationIP')">
        <el-input v-model="gpsForm.ip" :placeholder="$t('common.leaveEmptyForCurrentIP')"></el-input>
      </el-form-item>
      <el-form-item :label="$t('common.countryRegion')">
        <el-select v-model="gpsForm.country" :placeholder="$t('common.pleaseSelectCountryRegion')" filterable>
          <el-option
            v-for="(info, code) in countryMap"
            :key="code"
            :label="`${info.name} (${info.en})`"
            :value="code"
          >
            <span style="float: left">{{ info.name }} ({{ info.en }})</span>
            <span style="float: right; color: #8492a6; font-size: 13px">{{ code }}</span>
          </el-option>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="gpsDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitGPS" :loading="gpsLoading">
          {{ $t('common.confirm') }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 上传 Google 证书弹窗 -->
  <el-dialog
    v-model="googleCertDialogVisible"
    :title="$t('common.uploadGoogleCert')"
    width="460px"
    :close-on-click-modal="false"
  >
    <!-- 隐藏的文件选择 input，仅接受 pem / xml -->
    <input
      ref="googleCertInputRef"
      type="file"
      accept=".pem,.xml"
      style="display: none"
      @change="onGoogleCertFileChange"
    />

    <div style="padding: 8px 0;">
      <div style="margin-bottom: 12px; color: #606266; font-size: 13px;">
        {{ $t('common.supportedFormats') }}：<strong>.pem</strong>、<strong>.xml</strong>
      </div>

      <!-- 文件选择区 -->
      <div
        style="
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 14px 16px;
          border: 1px dashed #d9d9d9;
          border-radius: 6px;
          background: #fafafa;
          cursor: pointer;
        "
        @click="triggerGoogleCertInput"
      >
        <el-icon :size="22" style="color: #409eff; flex-shrink: 0;"><Upload /></el-icon>
        <span
          v-if="googleCertFileName"
          style="flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #303133; font-size: 14px;"
        >{{ googleCertFileName }}</span>
        <span v-else style="flex: 1; color: #aaa; font-size: 14px;">{{ $t('common.clickToSelectCert') }}</span>
        <el-button
          v-if="googleCertFileName"
          type="primary"
          link
          size="small"
          style="flex-shrink: 0;"
          @click.stop="triggerGoogleCertInput"
        >{{ $t('common.reselect') }}</el-button>
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="googleCertDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="submitGoogleCert"
          :loading="googleCertLoading"
          :disabled="!googleCertFileName"
        >
          {{ googleCertLoading ? $t('common.uploading') : $t('common.confirmUpload') }}
        </el-button>
      </span>
    </template>
  </el-dialog>

  <!-- ===== 设置弹窗 ===== -->
  <el-dialog
    v-model="settingsDialogVisible"
    :title="$t('common.settings')"
    width="500px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div style="padding: 8px 0;">
      <div style="margin-bottom: 20px;">
        <div style="font-weight: bold; margin-bottom: 12px; color: #303133; font-size: 14px;">
          文件保存路径
        </div>
        <div style="font-size: 12px; color: #909399; margin-bottom: 12px; line-height: 1.6;">
          设置下载镜像、本地机型、备份机型、备份云机等文件的保存位置。<br>
          默认保存在 C 盘系统目录，建议修改到其他磁盘以避免 C 盘空间不足。
        </div>

        <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
          <el-input
            v-model="storagePathInfo.path"
            placeholder="请选择或输入保存路径"
            style="flex: 1;"
            :readonly="true"
          >
            <template #prefix>
              <el-icon><FolderOpened /></el-icon>
            </template>
          </el-input>
          <el-button
            type="primary"
            :loading="settingsLoading"
            @click="handleSelectDirectory"
          >
            浏览
          </el-button>
        </div>

        <div style="display: flex; align-items: center; gap: 8px; margin-top: 8px;">
          <el-tag v-if="storagePathInfo.isDefault" type="info" size="small">当前为默认路径</el-tag>
          <el-tag v-else type="success" size="small">已自定义路径</el-tag>
          <el-button
            v-if="!storagePathInfo.isDefault"
            type="text"
            size="small"
            style="color: #909399;"
            @click="handleResetStoragePath"
            :loading="settingsLoading"
          >
            恢复默认
          </el-button>
        </div>

        <div v-if="storagePathInfo.defaultPath" style="margin-top: 10px; font-size: 12px; color: #c0c4cc;">
          默认路径：{{ storagePathInfo.defaultPath }}
        </div>

        <el-alert
          title="修改保存路径后，已有文件不会自动迁移，请手动将旧目录中的文件复制到新路径。"
          type="warning"
          show-icon
          :closable="false"
          style="margin-top: 12px;"
        />
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="settingsDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="settingsLoading"
          @click="handleSaveSettings"
        >
          保存
        </el-button>
      </span>
    </template>
  </el-dialog>
  <!-- ===== 设置弹窗 END ===== -->

</template>

<style >
/* 加载图标旋转动画 */
.rotating {
  animation: rotate 1.5s linear infinite;
}

.fp-error-app {
  display: flex;
  align-items: center;
  color: #f56c6c;
  font-size: 12px;
  margin-top: 4px;
  line-height: 1.4;
}


@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 批量新机界面样式 */
.batch-slot-checkbox-group .el-checkbox-button__inner {
  border-radius: 0 !important;
  width: 100%;
  border: 2px solid #dcdfe6;
  padding: 5px 16px !important;
  box-shadow: none !important; /* Remove box-shadow used for active border */
}

.batch-slot-checkbox-group .el-checkbox-button.is-checked .el-checkbox-button__inner {
    background-color: #409eff;
    border-color: #409eff !important;
    color: #fff;
}

.batch-slot-checkbox-group .el-checkbox-button:first-child .el-checkbox-button__inner,
.batch-slot-checkbox-group .el-checkbox-button:last-child .el-checkbox-button__inner {
    border-radius: 0 !important;
}

.batch-switch-model-content {
  overflow: hidden;
}

.slot-item {
  padding: 8px 12px;
  margin: 4px 0;
  background-color: #f0f2f5;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: move;
  user-select: none;
}

.slot-item:hover {
  background-color: #e4e7ed;
}

.slot-item.dragging {
  opacity: 0.5;
}

/* V2容器样式 - 灰色显示，不可拖动 */
.slot-item.v2-container {
  background-color: #f5f5f5;
  border: 1px dashed #dcdfe6;
  cursor: not-allowed;
  opacity: 0.7;
}

.slot-item.v2-container:hover {
  background-color: #f5f5f5;
}

.available-slots {
  max-height: 400px;
  overflow-y: auto;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.model-slot {
  padding: 15px;
  margin-bottom: 15px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #fafafa;
}

.model-slot-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.assigned-slots {
  min-height: 50px;
  padding: 10px;
  border: 1px dashed #c0c4cc;
  border-radius: 4px;
  background-color: #fff;
}

.assigned-slot-item {
  display: inline-block;
  padding: 6px 12px;
  margin: 4px;
  background-color: #ecf5ff;
  border: 1px solid #d9ecff;
  border-radius: 4px;
  cursor: move;
}

/* V2容器在已分配区域的样式 */
.assigned-slot-item.v2-container {
  background-color: #f5f5f5;
  border: 1px dashed #dcdfe6;
  cursor: not-allowed;
  opacity: 0.7;
}

.no-slots-assigned {
  color: #909399;
  text-align: center;
  padding: 10px;
}

/* 清除浮动 */
.batch-switch-model-content::after {
  content: "";
  display: table;
  clear: both;
}
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* 终端遮罩层样式 */
.terminal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 9998;
  display: none;
}

/* 终端相关样式 */
#terminal {
  position: fixed;
  top: 30%;
  left: 35%;
  width: 90%;
  height: 90%;
  /* margin-top: -200px;
  margin-left: -450px; */
  transform: translate(-35%, -30%);
  display: none;
  z-index: 9999;
  background-color: #000000; /* 黑色背景，类似Windows命令提示符 */
  overflow: hidden;
  border: 1px solid #333333;
  border-radius: 4px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

/* 终端标题栏样式 */
.terminal-header {
  width: 100%;
  height: 32px;
  background: linear-gradient(180deg, #2d2d2d 0%, #1a1a1a 100%);
  border-bottom: 1px solid #333333;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px;
  box-sizing: border-box;
  cursor: move;
}

.terminal-title {
  color: #cccccc;
  font-size: 12px;
  font-weight: 500;
  user-select: none;
}

/* 终端内容区域样式 */
.terminal-content {
  width: 100%;
  height: calc(100% - 32px);
  overflow: hidden;
  position: relative;
  text-align: left;
  background-color: #000000; /* 黑色背景 */
}

/* 确保xterm终端文本左对齐 */
.terminal-content .xterm {
  text-align: left !important;
  background-color: #000000 !important; /* 黑色背景 */
}

/* 确保xterm终端行左对齐 */
.terminal-content .xterm-rows {
  justify-content: flex-start !important;
}

/* 设置xterm终端的背景色 */
.terminal-content .xterm-background {
  background-color: #000000 !important; /* 黑色背景 */
}

/* 登录覆盖层样式 - 暂时隐藏，因为我们直接连接终端 */
#login-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
  display: none;
}

.login-box {
  background: #fff;
  padding: 30px;
  border-radius: 8px;
  width: 350px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #666;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px;
  box-sizing: border-box;
}

.btn {
  width: 100%;
  padding: 10px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

/* SDK加载蒙版样式 */
.sdk-loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
  overflow: hidden;
}

.sdk-loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
}

/* 创建云机对话框样式 */
.create-dialog-container .el-dialog {
  margin-top: 20px !important;
  margin-bottom: 20px !important;
}

.sdk-loading-text {
  margin-top: 15px;
  color: #606266;
  font-size: 16px;
}

.sdk-loading-progress {
  margin-top: 25px;
  width: 100%;
  max-width: 700px;
}

/* 扫描设备样式 */
.scanning-devices {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  font-size: 16px;
  color: #606266;
}

.scanning-devices .el-icon {
  margin-right: 10px;
  font-size: 20px;
}

.no-devices {
  padding: 40px 0;
  text-align: center;
}

.no-devices .tip {
  margin-top: 10px;
  color: #909399;
  font-size: 14px;
}

.scanned-devices-list {
  height: 400px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.scanned-devices-list .search-box {
  flex-shrink: 0;
}

.scanned-devices-list .el-table {
  flex: 1;
  overflow-y: auto;
  margin-bottom: 10px;
}

.scanned-devices-list .selected-info {
  flex-shrink: 0;
  margin-top: 0;
}

.selected-info {
  color: #606266;
  margin-top: 10px;
}

/* 右侧栏样式 */
.device-right-col {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.device-right-col .table-container {
  flex: 1;
  overflow-y: auto;
}

/* 镜像管理样式 */
.image-info-container {
  padding: 0;
  height: calc(100% - 60px);
  min-height: 340px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  position: relative;
  z-index: 1;
  overflow: hidden;
}

/* 右侧内容区域样式 */
.el-col:nth-child(2) {
  position: relative;
  z-index: 10;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 500px;
}

/* 悬浮工具栏样式 */
.floating-toolbar {
  position: sticky;
  top: 0;
  z-index: 20;
  background: white;
  padding: 10px 0;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
}

/* 镜像分类标签样式 */
.image-category-tabs {
  padding: 0;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.image-category-tabs > div {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0;
}

.online-images-container,
.local-images-container,
.box-images-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 0;
}

.online-images-list,
.local-image-items,
.box-image-items {
  flex: 1;
  overflow: hidden;
  border: none;
  border-radius: 0;
  margin: 0;
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

/* 筛选条件样式 */
.image-filters {
  padding: 15px;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.image-table .cell {
  display: block !important;
}

/* 表格容器样式 */
.image-table-wrapper {
  flex: 1;
  overflow: auto;
  max-height: calc(100vh - 400px);
}

/* 表格样式 */
.image-table-wrapper .el-table {
  width: 100%;
  margin-bottom: 0;
  table-layout: auto !important;
}

/* 确保表格列可以自适应 */
.image-table-wrapper .el-table th {
  white-space: nowrap;
}

/* 确保表格内容可以换行 */
.image-table-wrapper .el-table td {
  word-break: break-all;
}

/* 确保表格不溢出容器 */
.el-table {
  width: 100%;
  overflow: hidden;
}

/* 确保卡片容器适应内容窗口 */
.el-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.el-card__body {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 确保标签页内容适应容器 */
.el-tabs__content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 确保标签页面板适应容器 */
.el-tab-pane {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.image-loading {
  padding: 20px 0;
}

.no-images {
  padding: 50px 0;
  text-align: center;
}

.online-images-by-model {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.model-images-section {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 15px;
  background-color: #fafafa;
}

.model-title {
  font-size: 18px;
  margin-bottom: 15px;
  color: #303133;
  font-weight: bold;
}

.image-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.image-item {
  width: 100%;
  background-color: white;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 15px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.image-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.image-item-header {
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.image-name {
  font-weight: bold;
  color: #303133;
  font-size: 16px;
}

.image-item-body {
  margin-bottom: 15px;
}

.image-detail {
  margin-bottom: 8px;
  font-size: 14px;
}

.image-detail .label {
  color: #909399;
  margin-right: 8px;
}

.image-detail .value {
  color: #303133;
  word-break: break-all;
}

.image-item-actions {
  display: flex;
  justify-content: flex-end;
}

/* 镜像卡片头部样式 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 删除按钮样式 */
.delete-button {
  margin-left: auto;
}

.download-progress-container {
  margin-top: 20px;
  padding: 20px;
  background-color: #f0f9eb;
  border: 1px solid #e1f3d8;
  border-radius: 8px;
}

.download-progress-text {
  margin-top: 10px;
  text-align: center;
  color: #67c23a;
  font-size: 14px;
}

.no-device {
  padding: 50px 0;
  text-align: center;
}

.no-device-selected {
  padding: 50px 0;
  text-align: center;
}

/* 创建云机弹窗左右分栏布局 */
.create-dialog-left-right {
  display: flex;
  gap: 20px;
}

.create-dialog-left {
  flex: 1;
}

.create-dialog-right {
  flex: 1;
}

/* 增加标题和内容之间的间距 */
.create-dialog-content {
  /* margin-top: 40px !important; */
  position: relative;
}

/* 增加表单项之间的间距 */
.create-dialog-content .el-form-item {
  margin-bottom: 30px !important;
}

/* 增加左右栏内部的间距 */
.create-dialog-left, .create-dialog-right {
  padding: 10px;
}

/* 增加单选按钮组之间的间距 */


/* 增加高级选项的间距 */
.create-dialog-content .el-checkbox {
  margin-right: 20px;
 
}

.create-type-option {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.create-type-switch {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: center;
  position: relative;
}

.create-type-side {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.create-type-left {
  margin-right: 6px;
}

.create-type-right {
  margin-left: 6px;
}

.create-type-tag {
  border: 1px solid;
  border-radius: 10px;
  padding: 0 6px;
  font-size: 12px;
  line-height: 18px;
}

.create-type-tag-success {
  color: #67c23a;
  border-color: #67c23a;
  background: transparent;
}

.create-type-tag-muted {
  color: #909399;
  border-color: #dcdfe6;
  background: transparent;
}

/* 确保坑位和数量在一行显示 */
.batch-slot-config {
  display: flex !important;
  align-items: center !important;
  gap: 10px !important;
}

.batch-slot-config span {
  margin: 0 !important;
  padding: 0 !important;
}

/* 设备创建按钮样式 */
.device-create-btn {
  color: #000000 !important;
  background-color: transparent !important;
  border-color: #000000 !important;
}

.device-create-btn:hover {
  color: #333333 !important;
  background-color: rgba(0, 0, 0, 0.1) !important;
  border-color: #333333 !important;
}

/* 小加号按钮样式 */
.small-plus-btn {
  width: 28px !important;
  height: 28px !important;
  padding: 0 !important;
  margin: 0 !important;
  border: none !important;
  background: transparent !important;
  color: #000000 !important;
  font-size: 16px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.small-plus-btn:hover {
  background: rgba(0, 0, 0, 0.1) !important;
  border-radius: 4px !important;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f7fa;
  color: #303133;
  font-size: 14px;
}

#app {
  width: 100%;
  height: 100vh;
}

/* 主容器布局 */
.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f8f9fa;
  width: 100%;
  position: relative;
  overflow: hidden;
}

/* 主要内容区域 */
.app-main {
  padding: 0;
  background: #f8f9fa;
  height: 100vh;
  width: 100%;
  overflow: hidden;
}

/* 固定标签栏样式 */
.fixed-tabs {
  position: sticky;
  top: 0;
  z-index: 100;
  background-color: #ffffff;
  border-radius: 0;
  width: 100%;
}

/* 标签页内容区域 */
.el-tabs__content {
  height: calc(100vh - 100px);
  overflow: auto;
  padding: 20px;
  background: #ffffff;
  border-radius: 0;
  box-shadow: none;
  border: 1px solid #e9ecef;
  border-top: none;
  width: 100%;
}

/* 确保整个页面填充窗口 */
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
  width: 100%;
}

/* 修复卡片边框问题 */
.el-card {
  border-radius: 8px;
  margin-bottom: 16px;
  border: 1px solid #e9ecef;
  background: #ffffff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

/* 修复表格滚动问题 */
.el-table__body-wrapper {
  overflow: auto;
  height: auto;
  max-height: 600px;
}

/* 镜像管理卡片内的表格不限制高度，由外层容器控制滚动 */
.image-management-card .el-table__body-wrapper {
  overflow: visible;
  height: auto;
  max-height: none;
}

/* 修复左侧设备列表滚动 */
.device-card {
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
  width: 100%;
  box-sizing: border-box;
}

/* 确保所有设备菜单都能正常滚动 */
.device-menu {
  flex: 1;
  overflow: auto;
  max-height: calc(100vh - 300px);
  border: none;
  background-color: transparent;
  width: 100%;
  box-sizing: border-box;
}

/* 设备列表容器样式 */
.device-list-container {
  flex: 1;
  overflow: auto;
  width: 100%;
  box-sizing: border-box;
  min-height: 0;
}

/* 批量模式容器样式 */
.batch-mode-container {
  flex: 1;
  overflow: hidden;
  width: 100%;
  box-sizing: border-box;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* 设备列表滚动样式 */
.device-list-scroll {
  flex: 1;
  overflow-y: auto !important;
  height: 100% !important;
  position: relative;
}

/* 强制修复el-scrollbar样式 */
.batch-mode-container :deep(.el-scrollbar) {
  height: 100% !important;
  width: 100% !important;
  overflow: auto !important;
  position: relative;
}

.batch-mode-container :deep(.el-scrollbar__wrap) {
  overflow-y: scroll !important;
  height: 100% !important;
  width: 100% !important;
  margin-right: -17px; /* 为滚动条预留空间 */
}

.batch-mode-container :deep(.el-scrollbar__view) {
  height: auto !important;
  overflow: visible !important;
  width: 100% !important;
}

/* 树形结构样式修复 */
.batch-mode-container :deep(.el-tree) {
  height: auto !important;
  width: 100% !important;
  overflow: visible !important;
  min-height: 100%;
}

/* 确保树形节点能正确显示 */
.batch-mode-container :deep(.el-tree-node) {
  height: auto !important;
  overflow: visible !important;
}

.batch-mode-container :deep(.el-tree-node__children) {
  height: auto !important;
  overflow: visible !important;
}

/* 强制显示滚动条 */
.batch-mode-container :deep(.el-scrollbar__bar) {
  opacity: 1 !important;
}

/* 原生滚动容器样式 */
.native-scroll-container {
  height: 100%;
  overflow-y: auto;
  padding-right: 8px;
  box-sizing: border-box;
}

/* 确保滚动条始终可见 */
.native-scroll-container::-webkit-scrollbar {
  width: 17px;
}

.native-scroll-container::-webkit-scrollbar-track {
  background: #f1f1f1;
}

.native-scroll-container::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.native-scroll-container::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 修复表格容器滚动 */
.el-table-container {
  overflow: auto;
  max-height: calc(100vh - 220px);
}

/* 确保所有卡片都能正确适应窗口大小 */
.el-card {
  box-sizing: border-box;
  width: 100%;
}

/* 修复云机管理页面的表格滚动 */
.cloud-machine-table .el-table__body-wrapper {
  max-height: calc(100vh - 220px);
}

/* 确保所有内容区域都能正确滚动 */
.el-tabs__content {
  overflow: auto;
  height: calc(100vh - 100px);
  box-sizing: border-box;
}

/* 设备列表项样式 */
.device-menu-item {
  border-radius: 4px;
  margin: 4px 0;
  transition: all 0.2s ease;
}

/* 设备列表项悬停效果 */
.device-menu-item:hover {
  background-color: rgba(64, 158, 255, 0.1) !important;
}

/* 设备列表项选中样式 */
.device-menu-item.is-active {
  background-color: rgba(64, 158, 255, 0.2) !important;
  border-left: 3px solid #409EFF !important;
}

/* 设备类型图标样式 */
.device-type-icon {
  transition: all 0.2s ease;
}

/* 设备IP文本样式 */
.device-ip {
  font-size: 14px;
  font-weight: 500;
}

/* 云机截图样式 */
.cloud-machine-screenshot-img {
  width: 100%;
  height: auto;
  max-height: 200px;
  object-fit: contain;
  border-radius: 4px;
}

/* 截图加载状态 */
.screenshot-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  background-color: #f5f7fa;
  border-radius: 4px;
  color: #909399;
}

/* 截图离线状态 */
.screenshot-offline {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  background-color: #fef0f0;
  border-radius: 4px;
  color: #f56c6c;
}

/* 截图空状态 */
.screenshot-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  background-color: #f5f7fa;
  border-radius: 4px;
  color: #909399;
}

/* 加载图标旋转动画 */
/* 只让图标旋转，而不是整个按钮 */
el-icon.is-loading {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 截图错误状态 */
.screenshot-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #fef6f6;
  color: #e6a23c;
  gap: 5px;
}

/* 截图图片样式优化，保持原始比例 */
.cloud-machine-screenshot-img {
  width: 100%;
  height: auto;
  max-height: 100%;
  object-fit: contain;
  background-color: #f5f7fa;
  border-radius: 0;
  margin: 0 auto;
}

/* 截图加载中状态 */
.screenshot-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f5f7fa;
  color: #606266;
  gap: 5px;
}

/* 截图离线状态 */
.screenshot-offline {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f5f7fa;
  color: #606266;
}

/* 文件树样式 */
.file-tree {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 8px;
  background-color: #f9f9f9;
}

/* 排序按钮样式 */
.sort-active {
  color: #409eff;
  font-weight: 600;
}

/* 树形节点样式 */
.tree-node {
  display: flex;
  align-items: center;
  padding: 6px 0;
  margin: 2px 0;
  border-radius: 4px;
  transition: background-color 0.2s ease;
}

.tree-node:hover {
  background-color: rgba(64, 158, 255, 0.1);
}

/* 节点复选框 */
.node-checkbox {
  margin-right: 8px;
  flex-shrink: 0;
}

/* 节点内容 */
.node-content {
  display: flex;
  align-items: center;
  flex: 1;
  cursor: pointer;
}

/* 节点名称 */
.node-name {
  flex: 1;
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 节点大小 */
.node-size {
  font-size: 12px;
  color: #909399;
  margin-left: 16px;
  width: 80px;
  text-align: right;
}

/* 节点日期 */
.node-date {
  font-size: 12px;
  color: #909399;
  margin-left: 16px;
  width: 120px;
  text-align: right;
}

/* 节点复选框 */
.node-checkbox {
  margin-left: 16px;
}

/* 树形子节点 */
.tree-children {
  margin-left: 28px;
  border-left: 1px dashed #dcdfe6;
  padding-left: 8px;
}

/* 无文件状态 */
.no-files {
  text-align: center;
  padding: 40px 0;
}

/* 对话框头部 */
.dialog-header {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 滚动条样式 */
.file-tree::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.file-tree::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.file-tree::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.file-tree::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 截图空状态 */
.screenshot-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f5f7fa;
  color: #606266;
}

/* 修复窗口大小变化时的布局问题 */
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

/* 确保所有容器都使用box-sizing */
* {
  box-sizing: border-box;
}

/* 表格容器，修复滚动问题 */
.table-container {
  overflow: hidden;
  max-height: calc(100vh - 200px);
  width: 100%;
  box-sizing: border-box;
  margin-top: 0;
  padding: 0;
  margin-right: 0;
}

/* 修复表格操作按钮被遮挡的问题 */
.table-container .el-table {
  min-width: auto;
  width: auto;
}

/* 简化修复方案：移除固定列，让表格自动处理列宽 */

/* 表格容器样式 */
.table-container {
  overflow: auto;
  max-height: calc(100vh - 200px);
  width: 100%;
  box-sizing: border-box;
  margin-top: 0;
  padding: 0;
  position: relative;
}

/* 表格基本样式 */
.table-container .el-table {
  width: 100%;
  min-width: 100%;
  position: relative;
  overflow: auto !important;
}

/* 优化操作列样式：减少按钮数量后调整宽度 */

/* 修复表格容器 */
.table-container {
  overflow: auto !important;
  max-height: calc(100vh - 200px) !important;
  width: 100% !important;
  position: relative;
}

/* 修复表格布局 */
.el-table {
  width: 100% !important;
  min-width: 100% !important;
  overflow: auto !important;
  table-layout: auto !important; /* 自适应列宽 */
}

/* 固定表头 */
.el-table__header-wrapper {
  background: white !important;
  z-index: 10 !important;
}

/* 修复表格内容滚动 */

/* 设备详情弹窗样式 - 固定高度，允许内容区域滚动 */
/* 添加自定义class到el-dialog组件 */
.device-details-dialog {
  /* 限制弹窗最大高度 */
  max-height: 90vh;
}

/* 固定弹窗内容区域高度 */
.device-details-dialog .el-dialog__body {
  height: 84vh;
  overflow: hidden;
  padding: 0;
}

/* 固定内容容器高度 */
.device-details-content {
  height: 100%;
  padding: 20px;
  background-color: #f5f7fa;
  overflow: hidden;
}

/* 镜像标签页内容区域，允许垂直滚动 */
.image-list-container {
  height: 100%;
  /* overflow: hidden; */
}

/* 使用说明内容区域 */
.image-guide-container {
  overflow-y: auto !important;
}

.image-guide-content {
  padding: 20px 24px;
  /* max-width: 900px; */
}

.guide-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 4px 0;
}

.guide-section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 12px 0;
}

.guide-text {
  color: #606266;
  line-height: 1.8;
  margin-bottom: 10px;
}

.guide-list {
  color: #606266;
  line-height: 2;
  padding-left: 20px;
  margin: 0 0 4px 0;
}

.guide-list li {
  margin-bottom: 4px;
}

.guide-steps {
  padding: 0 8px;
}

.guide-compare-table {
  font-size: 13px;
}

/* 镜像管理卡片：body 区域限高并只内部滚动 */
.image-management-card .el-card__body {
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.image-management-card .image-list-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  height: 0; /* 配合 flex:1 让其撑满剩余空间并触发滚动 */
}

.image-management-card .online-image-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.image-management-card .online-image-tabs .el-tabs__content {
  flex: 1;
  overflow-y: auto;
  height: 0;
}

.image-management-card .online-image-tabs .el-tab-pane {
  height: 100%;
}

/* 镜像列表内容，允许垂直滚动 */
.image-list-content {
  height: calc(100% - 60px);
  overflow-y: auto;
  padding: 0 10px;
}

/* 优化滚动条样式 */
.image-list-content::-webkit-scrollbar {
  width: 8px;
}

.image-list-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.image-list-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.image-list-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
.el-table__body-wrapper {
  overflow: auto !important;
  max-height: calc(100vh - 250px) !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

/* 调整固定列宽度 */
.el-table-column--fixed-right {
  position: sticky !important;
  width: 200px !important; /* 调整宽度，适应两个按钮 */
  right: 0 !important;
  z-index: 10 !important;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1) !important;
  border-left: 1px solid #ebeef5 !important;
}

/* 确保固定列容器显示 */
.el-table__fixed-right {
  display: block !important;
  position: sticky !important;
  right: 0 !important;
  z-index: 10 !important;
  background: white !important;
  width: 200px !important; /* 调整宽度 */
}

/* 调整操作列宽度 */
.el-table-column:last-child {
  width: 200px !important;
  min-width: 200px !important;
}

/* 修复所有操作列宽度 */
.el-table-column[label="操作"] {
  width: 200px !important;
  min-width: 200px !important;
}

/* 修复操作按钮样式 */
.el-table .el-button {
  white-space: nowrap;
  min-width: 60px;
  margin: 0 2px;
  padding: 4px 10px;
  font-size: 12px;
}

/* 确保操作按钮组能完整显示 */
.el-table__cell .el-space {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-wrap: nowrap;
  width: 100%;
  justify-content: flex-start;
}

/* 修复单元格样式 */
.el-table__cell {
  overflow: visible !important;
  white-space: nowrap !important;
  padding: 8px 12px !important;
}

/* 调整操作列单元格宽度 */
.el-table__cell[data-column-key="操作"] {
  width: 200px !important;
  min-width: 200px !important;
}

/* 确保固定列内容正确显示 */
.el-table__fixed-right .el-table__body-wrapper {
  overflow: visible !important;
  max-height: none !important;
}

/* 修复固定列与主表对齐 */
.el-table__fixed-right {
  height: 100% !important;
}

/* 确保操作按钮不被遮挡 */
.el-table-column:last-child .el-table__cell {
  padding-right: 10px !important;
  width: 200px !important;
  min-width: 200px !important;
}

/* 修复Element Plus内置样式冲突 */
:deep(.el-table__header-wrapper .el-table__header) {
  width: 100% !important;
}

:deep(.el-table__body) {
  width: 100% !important;
}

/* 确保固定列不影响表格缩放 */
.el-table--scrollable-x {
  overflow-x: auto !important;
}

/* 确保表格可以正常缩放 */
.el-table__inner-wrapper {
  width: 100% !important;
}

/* 优化固定列表头样式 */
.el-table__fixed-right .el-table__header-wrapper {
  background: white !important;
  border-bottom: 1px solid #ebeef5 !important;
}

/* 修复固定列单元格样式 */
.el-table__fixed-right .el-table__body .el-table__cell {
  padding: 8px 12px !important;
  width: 200px !important;
  min-width: 200px !important;
}

/* 确保固定列按钮完整显示 */
.el-table__fixed-right .el-button {
  flex-shrink: 0 !important;
  margin: 0 2px !important;
}

/* 云机管理标签页布局 */
.el-tab-pane[name="cloud-management"] {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 云机管理网格布局 */
.el-tab-pane[name="cloud-management"] .el-row {
  flex: 1;
  display: flex;
  min-height: 0;
}

/* 云机管理列布局 */
.el-tab-pane[name="cloud-management"] .el-col {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* 云机管理右侧栏布局 */
.cloud-management-right-column {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* 云机列表容器 */
.cloud-machine-container {
  height: calc(100vh - 205px);
  overflow: auto;
  width: 100%;
  box-sizing: border-box;
  padding: 0;
  min-width: 0;
}

/* 批量模式容器 */
.batch-mode-container {
  min-height: 0;
}

/* 原生滚动容器 */
.native-scroll-container {
  min-height: 0;
}

/* 云机坑位容器 */
.cloud-machine-slots {
  width: 100%;
  box-sizing: border-box;
  min-height: 0;
}

/* 主标签页容器 */
.modern-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* 主标签页内容 */
.modern-tabs .el-tabs__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}



/* 修复表格容器的溢出问题 */
.table-container {
  overflow: auto !important;
}

/* 修复响应式布局 */
@media (max-width: 1200px) {
  .el-table {
    overflow-x: auto !important;
  }
}



/* 云机列表容器 */
.cloud-machine-container {
  height: calc(100vh - 205px);
  overflow: auto;
  width: 100%;
  box-sizing: border-box;
  padding: 0;
  min-width: 0;
}

/* 确保云机管理列表的操作列能完整显示 */
.cloud-machine-container .el-table {
  min-width: auto;
  width: auto;
}

/* 修复主机管理列表窗口拖大后的标题栏错位问题 */
.floating-toolbar {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 0;
  position: static;
}

/* 确保浮动工具栏在窗口放大时正确对齐 */
.device-info {
  display: flex;
  align-items: center;
  gap: 0px;
  flex-shrink: 0;
}

/* 确保批量操作按钮组能正确换行 */
.batch-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  flex: 1;
  justify-content: flex-end;
}

/* 修复标签页内容在窗口放大时的对齐问题 */
.modern-tabs .el-tabs__content {
  padding: 20px;
  box-sizing: border-box;
  width: 100%;
}

/* 云机列表容器 */
.cloud-machine-container {
  height: calc(100vh - 205px);
  overflow: auto;
  padding-right: 8px;
  min-width: 0;
}

/* 云机列表滚动 */
.cloud-machine-grid {
  overflow: visible;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

/* 修复设备显示清单滚动 */
.device-list-scroll {
  overflow: auto;
  max-height: calc(100vh - 250px);
}

/* 浅色云机卡片头部，最小化空白 */
.cloud-machine-card-header-light {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  font-weight: 600;
  padding: 4px 8px;
  background: #f8f9fa;
  color: #495057;
  margin: 0;
  border-bottom: 1px solid #e9ecef;
  line-height: 1.2;
}

/* 仅在云机卡片范围内移除默认padding */
.cloud-machine-card .el-card__header {
  padding: 0 !important;
  border-bottom: none !important;
}

.cloud-machine-card .el-card__body {
  padding: 0 !important;
}

/* 增大截图区域，适配180x320截图，恢复层次感背景 */
.cloud-machine-screenshot-large {
  width: 100%;
  /* 180x320的宽高比为9:16，所以高度应该是宽度的16/9倍 */
  aspect-ratio: 9/16;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #606266;
  margin: 0;
  font-weight: 500;
  overflow: hidden;
  border: 1px solid #dee2e6;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.05);
}

/* 确保截图加载状态和错误状态也能完全填充容器 */
.screenshot-loading,
.screenshot-error,
.screenshot-offline,
.screenshot-empty {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 5px;
  background-color: #f5f7fa;
  color: #606266;
}

/* 确保坑位模式下截图加载状态的文本正确显示，防止堆叠 */
.cloud-machine-screenshot-large .screenshot-loading > div {
  line-height: 1.5 !important;
  height: auto !important;
  white-space: normal;
  margin: 4px 0;
}

.cloud-machine-screenshot-large .screenshot-loading > div > div {
  line-height: 1.5 !important;
  height: auto !important;
  white-space: normal;
}

/* 修复device-details-content内部表格的滚动问题 */
.device-details-content .table-container {
  overflow: hidden;
  position: relative;
  height: calc(100% - 40px);
}

.device-details-content .table-container .el-table {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.device-details-content .table-container .el-table__header-wrapper {
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 10;
  background-color: white;
}

.device-details-content .table-container .el-table__body-wrapper {
  flex: 1;
  overflow-y: auto;
  overflow-x: auto;
  max-height: none !important;
  height: auto;
}

.device-details-content .table-container .el-table__inner-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.device-details-content .table-container .el-table__body {
  flex: 1;
}

/* 最小化信息区域和按钮的空白 */
.cloud-machine-info-small {
  padding: 4px 8px;
  background-color: #ffffff;
}

.cloud-machine-info-small .el-button {
  font-size: 11px;
  padding: 3px 8px;
  height: 24px;
  line-height: 24px;
}

/* 确保设备列表滚动正常 */
.device-menu {
  overflow: auto;
  max-height: calc(100vh - 250px);
}

/* 修复窗口拉大后的对齐问题 */
.el-row {
  width: 100%;
}

.el-col {
  padding: 0;
}

/* 修复标签栏和表格的偏移问题 */
.modern-tabs .el-tabs__content {
  width: 100%;
  box-sizing: border-box;
}

/* 修复云机卡片高度问题 */
.cloud-machine-card {
  height: auto;
}

/* 标签页样式 */
.modern-tabs {
  background-color: transparent;
  --el-tabs-header-height: 48px;
  margin-bottom: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.modern-tabs .el-tabs__header {
  margin-bottom: 0;
  background: #ffffff;
  border-radius: 8px 8px 0 0;
  padding: 0 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  border-bottom: 1px solid #e9ecef;
  overflow: visible;
}

.modern-tabs .el-tabs__nav {
  height: 48px;
}

.modern-tabs .el-tabs__item {
  height: 48px;
  line-height: 48px;
  /* font-size: 13px; */
  font-weight: 500;
  color: #6c757d;
  /* margin: 0 4px; */
  padding: 0 20px;
  position: relative;
  transition: all 0.3s ease;
}

.modern-tabs .el-tabs__item:hover {
  color: #409EFF;
}

.modern-tabs .el-tabs__item.is-active {
  color: #409EFF;
  font-weight: 600;
}

.modern-tabs .el-tabs__active-bar {
  background: #409EFF;
  height: 2px;
  border-radius: 2px;
}

.modern-tabs .el-tabs__content {
  background: #ffffff;
  border-radius: 0 0 8px 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  border: 1px solid #e9ecef;
  border-top: none;
}

/* 卡片样式 */
.el-card {
  border-radius: 8px;
  margin-bottom: 16px;
  border: 1px solid #e9ecef;
  background: #ffffff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.el-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  font-size: 14px;
  padding: 0;
}

.card-header span {
  color: #495057;
  font-weight: 600;
}

/* 设备卡片 */
.device-card {
  height: calc(100vh - 120px);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 设备列表头部 */
.device-list-header {
  padding: 8px 12px;
  border-bottom: 1px solid #e9ecef;
  margin-bottom: 8px;
  width: 100%;
  box-sizing: border-box;
}

/* 设备列表标签页 */
.device-list-tabs {
  display: flex;
  gap: 16px;
  width: 100%;
  box-sizing: border-box;
}

.device-tab {
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 13px;
  color: #6c757d;
  transition: all 0.3s ease;
}

.device-tab:hover {
  color: #409EFF;
}

.device-tab.active {
  color: #409EFF;
  background-color: #e7f3ff;
  font-weight: 500;
}

/* 设备菜单 */
.device-menu {
  border: none;
  background-color: transparent;
  padding: 4px 0;
  overflow: auto;
  flex: 1;
  max-height: calc(100vh - 300px);
  width: 100%;
  box-sizing: border-box;
}

.device-menu-item {
  border-radius: 6px;
  margin: 2px 8px;
  height: auto;
  padding: 8px 12px;
  transition: all 0.3s ease;
}

.device-menu-item:hover {
  background: #e7f3ff;
  color: #409EFF;
}

.device-menu-item.is-active {
  background: #e7f3ff;
  color: #409EFF;
  border-right: none;
  border-left: 3px solid #409EFF;
}

/* 设备菜单项内容 */
.device-menu-item-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

/* 状态标签 - 圆形图标样式 */
.status-tag {
  border-radius: 50%;
  width: 10px;
  height: 10px;
  padding: 0;
  min-width: 10px;
  margin-right: 8px;
  transition: all 0.3s ease;
}

.device-menu-item.is-active .status-tag,
.device-menu-item:hover .status-tag {
  background-color: #409EFF;
  color: #ffffff;
  border-color: transparent;
}

/* IP地址样式 */
.device-ip {
  font-size: 14px;
  color: #495057;
  transition: color 0.3s ease;
  font-family: 'Courier New', monospace;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.device-menu-item:hover .device-ip {
  color: #409EFF;
}

/* 创建按钮样式 */
.create-button {
  width: 28px;
  height: 28px;
  padding: 0;
  min-width: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background-color: #409EFF;
  border: none;
  transition: all 0.3s ease;
}

.create-button:hover {
  background-color: #66b1ff;
  transform: scale(1.1);
}

.create-button .el-icon {
  font-size: 14px;
  margin-right: 0;
}

/* 悬浮功能栏样式 */
.floating-toolbar {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  padding: 12px 16px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  width: 100%;
  box-sizing: border-box;
  position: relative;
}

/* 设备信息部分 */
.device-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 设备IP文本 */
.device-ip-text {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #409EFF;
  font-size: 14px;
}

/* 分隔符 */
.divider {
  color: #909399;
  font-weight: 300;
}

/* 工具栏按钮 */
.toolbar-button {
  padding: 6px 16px;
  border-radius: 6px;
  font-weight: 500;
  transition: all 0.3s ease;
  margin: 0;
}

.toolbar-button.active {
  background-color: #409EFF;
  color: #ffffff;
  border: none;
}

.toolbar-button:not(.active):hover {
  color: #409EFF;
  background-color: #e7f3ff;
}

/* 批量操作按钮组 */
.batch-actions {
  margin-bottom: 0;
  padding: 0;
  background: transparent;
  border: none;
  gap: 8px;
}

/* 状态标签 - 正常宽度样式 */
.status-tag-normal {
  font-size: 11px;
  padding: 4px 12px;
  border-radius: 20px;
  font-weight: 500;
  color: #ffffff;
}

/* 确保表格内容对齐 */
.el-table {
  margin-top: 0;
}

/* 修复标题栏偏移 */
.modern-tabs .el-tabs__header {
  padding: 0 8px;
}

/* 修复表格行高 */
.el-table__row {
  height: 40px;
}

/* 修复表格单元格内边距 */
.el-table__cell {
  padding: 8px 12px;
}

/* 坑位表格样式 */
.slot-table {
  margin-top: 16px;
  border-radius: 8px;
  border: 1px solid #e9ecef;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  width: 100% !important;
  box-sizing: border-box;
}

.slot-table .el-table__header-wrapper {
  background: #f8f9fa;
  width: 100% !important;
}

.slot-table .el-table__header-wrapper th {
  background: transparent;
  color: #495057;
  font-weight: 600;
  border-bottom: 1px solid #dee2e6;
  padding: 10px 0;
  min-width: 0;
}

.slot-table .el-table__body-wrapper {
  background-color: #ffffff;
  width: 100% !important;
  overflow-y: auto;
  max-height: 600px;
  margin-bottom: 0;
}

.slot-table .el-table__body {
  width: 100% !important;
}

.slot-table .el-table__row {
  transition: all 0.3s ease;
  width: 100% !important;
}

.slot-table .el-table__row:hover {
  background-color: #f8f9fa;
}

.slot-table .el-table__cell {
  padding: 8px 0;
  border-bottom: 1px solid #f1f3f5;
  min-width: 0;
  box-sizing: border-box;
}

.slot-table .el-table--striped .el-table__body tr.el-table__row--striped {
  background-color: #f8f9fa;
}

.slot-table .el-table__row:last-child .el-table__cell {
  border-bottom: none;
}

/* 确保表格容器和表格能正确对齐 */
.el-table {
  width: 100% !important;
  box-sizing: border-box;
  margin: 0;
}

.el-table__header {
  width: 100% !important;
}

.el-table__body {
  width: 100% !important;
}

/* 修复表格容器和工具栏的对齐问题 */
.cloud-machine-slots {
  width: 100%;
  box-sizing: border-box;
}

/* 修复云机管理坑位模式的布局问题 */
.cloud-machine-slots .cloud-machine-grid {
  width: 100%;
  box-sizing: border-box;
}

/* 云机管理样式 */
.device-list-card {
  margin-bottom: 20px;
  border-radius: 8px;
}

.device-list-scroll {
  max-height: 140px;
  border-radius: 8px;
  overflow: hidden;
}

.device-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 12px;
}

.device-checkbox {
  margin: 0;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.device-checkbox:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.device-checkbox .el-checkbox__label {
  font-size: 13px;
  font-weight: 500;
}

/* 云机网格样式 */
.cloud-machine-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 10px;
  margin-top: 20px;
}

.cloud-machine-card {
  margin-bottom: 0;
  height: auto;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
  position: relative;
  padding: 0;
  /* 移除卡片的默认内边距 */
  --el-card-padding: 0;
}

/* 云机卡片内容容器 - 移除 min-height 限制 */
.cloud-machine-card .el-card__body {
  padding: 0;
  height: auto;
  min-height: 0; /* 移除最小高度限制，让 aspect-ratio 正常工作 */
}

/* 云机卡片内容 */
.cloud-machine-card .el-card__body > * {
  width: 100%;
  height: auto;
}

/* 云机信息和按钮部分 */
.cloud-machine-info-small {
  padding: 10px;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
}

/* 云机操作按钮 */
.cloud-machine-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
  width: 100%;
}

/* 云机截图样式 */
.cloud-machine-screenshot-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  display: block;
}

/* 云机截图容器 - 移除 min-height，让 aspect-ratio 控制高度 */
.cloud-machine-screenshot-large {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  padding: 0;
  position: relative;
  overflow: hidden;
}

/* 云机截图容器中的图片 */
.cloud-machine-screenshot-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  position: absolute;
  top: 0;
  left: 0;
}

.cloud-machine-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.cloud-machine-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  font-weight: 600;
  padding: 12px 16px;
  background: #409EFF;
  color: #ffffff;
  margin: 0;
}

.cloud-machine-card-header span {
  color: #ffffff;
}

.cloud-machine-screenshot {
  width: 100%;
  height: 120px;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #606266;
  border-radius: 0;
  margin: 0;
  font-weight: 500;
}

.cloud-machine-screenshot-small {
  width: 100%;
  height: 60px;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #606266;
  border-radius: 6px;
  font-weight: 500;
}

.screenshot-loading-small,
.screenshot-error-small,
.screenshot-empty-small {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.screenshot-loading-small {
  color: #909399;
}

.screenshot-error-small {
  color: #f56c6c;
}

.screenshot-empty-small {
  color: #c0c4cc;
}

.cloud-machine-info {
  padding: 16px;
  background-color: #ffffff;
}

.cloud-machine-actions {
  width: 100%;
  justify-content: center;
  gap: 8px;
}

.cloud-machine-actions .el-button {
  border-radius: 6px;
  transition: all 0.3s ease;
}

.cloud-machine-actions .el-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.connect-button {
  margin-top: 8px;
  border-radius: 6px;
  background: #409EFF;
  border: none;
  transition: all 0.3s ease;
}

.connect-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(64, 158, 255, 0.4);
}

/* VPC管理样式 */
.vpc-proxy-card, .vpc-host-card {
  height: 100%;
  border-radius: 8px;
}

.ip-test-card {
  margin-top: 20px;
  border-radius: 8px;
  background: #f8f9fa;
}

.ip-test-card .el-space {
  width: 100%;
  padding: 12px;
}



/* 功能占位符 */
.feature-card {
  height: 350px;
  border-radius: 8px;
  background: #ffffff;
}

.feature-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: calc(100% - 60px);
  flex-direction: column;
  gap: 16px;
}

.feature-placeholder .el-empty {
  --el-empty-description-font-size: 16px;
  --el-empty-description-color: #606266;
}

.feature-placeholder .el-empty__image {
  margin-bottom: 16px;
}

/* 按钮样式优化 */
.el-button {
  border-radius: 6px;
  transition: all 0.3s ease;
  font-weight: 500;
  text-transform: none;
  letter-spacing: 0.3px;
}

.el-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.el-button--primary {
  background: #409EFF;
  border: none;
}

.el-button--primary:hover {
  background: #66b1ff;
  box-shadow: 0 6px 20px rgba(64, 158, 255, 0.4);
}

.el-button--danger {
  background: #F56C6C;
  border: none;
}

.el-button--danger:hover {
  background: #f78989;
  box-shadow: 0 6px 20px rgba(245, 108, 108, 0.4);
}

/* 输入框样式优化 */
.el-input {
  border-radius: 6px;
  overflow: hidden;
}

.el-input__wrapper {
  border-radius: 6px;
  transition: all 0.3s ease;
}

.el-input__wrapper:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 选择器样式优化 */
.el-select {
  border-radius: 6px;
  overflow: hidden;
}

.el-select .el-input__wrapper {
  border-radius: 6px;
}

/* 标签样式优化 */
.el-tag {
  border-radius: 20px;
  padding: 4px 12px;
  font-weight: 500;
  font-size: 11px;
  letter-spacing: 0.5px;
}

/* 状态标签样式优化，确保文字清晰可见 */
.status-tag {
  color: #ffffff !important;
  font-weight: 600;
  background-color: #67C23A !important;
}

.status-tag.warning {
  background-color: #E6A23C !important;
}

.status-tag.info {
  background-color: #909399 !important;
}

/* 修复云机卡片头状态标签颜色 */
.cloud-machine-card-header .el-tag {
  color: #ffffff !important;
}

/* 菜单按钮样式 */
.menu-button {
  margin: 0;
  padding: 0 16px;
  color: #6c757d;
  transition: all 0.3s ease;
  background: transparent;
  border: none;
  font-size: 16px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.menu-button:hover {
  color: #409EFF;
  background-color: #f8f9fa;
}

.menu-button .el-icon {
  font-size: 20px;
  margin-right: 0;
}

/* 修复标签栏右侧按钮区域 */
.el-tabs__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

/* 确保标签栏和按钮区域对齐，允许溢出时滚动 */
.modern-tabs .el-tabs__nav-wrap {
  flex: 1;
  overflow: hidden;
  min-width: 0;
  position: relative;
}

/* 标签过多时滚动箭头样式 */
.modern-tabs .el-tabs__nav-prev,
.modern-tabs .el-tabs__nav-next {
  line-height: 48px;
  height: 48px;
  width: 32px;
  text-align: center;
  cursor: pointer;
  color: #6c757d;
  font-size: 16px;
  background: #ffffff;
  position: absolute;
  top: 0;
  z-index: 10;
}

.modern-tabs .el-tabs__nav-prev {
  left: 0;
}

.modern-tabs .el-tabs__nav-next {
  right: 0;
}

.modern-tabs .el-tabs__nav-prev:hover,
.modern-tabs .el-tabs__nav-next:hover {
  color: #409EFF;
}

/* 标签容器可滚动时的内边距 */
.modern-tabs .el-tabs__nav-wrap.is-scrollable {
  padding: 0 32px;
}

/* 标签内容区域滚动时不显示滚动条 */
.modern-tabs .el-tabs__nav-scroll {
  overflow: hidden;
  position: relative;
}

/* 修复下拉菜单显示 */
.el-dropdown {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 48px;
}

/* 分隔线样式 */
.el-divider {
  background: #e9ecef;
  margin: 16px 0;
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background-color: #f1f1f1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background-color: #c1c1c1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background-color: #a8a8a8;
}

/* 切换备份时的覆盖层样式 */
.switching-backup-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 100;
  color: white;
  font-size: 14px;
  border-radius: 8px;
}

/* 切换备份弹窗的覆盖层样式 */
.switching-backup-overlay-dialog {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.7);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  color: #303133;
  font-size: 16px;
  border-radius: 8px;
}

.switching-backup-overlay .is-loading,
.switching-backup-overlay-dialog .is-loading {
  font-size: 24px;
  margin-bottom: 8px;
  color: #409EFF;
}

.switching-backup-overlay .is-loading svg,
.switching-backup-overlay-dialog .is-loading svg,
.is-rotating svg {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 坑位单元格样式 */
.slot-cell-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.slot-cell-content .create-button {
  margin-left: 8px;
  padding: 4px 8px;
  width: auto;
  height: auto;
  display: inline-block;
}

.slot-cell-content .fold-button {
  font-size: 14px;
  font-weight: bold;
  padding: 4px 8px;
  margin-right: 4px;
}

.slot-cell-content .slot-number {
  font-size: 14px;
  font-weight: bold;
  padding: 4px 8px;
  margin-right: 4px;
  color: #606266;
}

/* 刷新按钮样式 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header .refresh-button {
  margin-left: 10px;
}

/* 批量创建按钮样式 */
.batch-create-btn {
  margin-left: 5px;
  padding: 0;
  color: #606266;
}

.device-menu-item-content {
  display: flex;
  align-items: center;
}

.switching-backup-overlay span {
  color: white;
  font-weight: 600;
}

.switching-backup-overlay-dialog span {
  color: #303133;
  font-weight: 600;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .cloud-machine-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  }
}

@media (max-width: 992px) {
  .app-main {
    padding: 12px;
  }
  
  .el-col {
    margin-bottom: 16px;
  }
  
  .cloud-machine-grid {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  }
}

@media (max-width: 768px) {
  .app-header {
    padding: 0 12px;
  }
  
  .header-title {
    font-size: 16px;
  }
  
  .el-tabs--border-card > .el-tabs__content {
    padding: 12px;
  }
  
  .cloud-machine-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 8px;
  }
}

/* 主机信息容器样式 */
.host-info-container {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  padding: 10px 0;
}

/* 主机信息卡片样式 */
.host-info-card {
  height: auto;
  min-height: 300px;
}

/* 确保主机信息卡片内容可以滚动 */
.host-info-content {
  max-height: calc(100vh - 280px);
  overflow-y: auto;
  padding: 0 10px;
}

/* 美化滚动条 */
.host-info-container::-webkit-scrollbar,
.host-info-content::-webkit-scrollbar,
.network-info-container::-webkit-scrollbar,
.network-info-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.host-info-container::-webkit-scrollbar-track,
.host-info-content::-webkit-scrollbar-track,
.network-info-container::-webkit-scrollbar-track,
.network-info-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.host-info-container::-webkit-scrollbar-thumb,
.host-info-content::-webkit-scrollbar-thumb,
.network-info-container::-webkit-scrollbar-thumb,
.network-info-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.host-info-container::-webkit-scrollbar-thumb:hover,
.host-info-content::-webkit-scrollbar-thumb:hover,
.network-info-container::-webkit-scrollbar-thumb:hover,
.network-info-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 网络信息容器样式 */
.network-info-container {
  padding: 0;
}

/* 网络信息卡片样式 */
.network-info-card {
  height: auto;
  min-height: 300px;
}

/* 确保网络信息卡片内容可以滚动 */
.network-info-content {
  padding: 0;
}

/* 网络列表表格样式 */
.network-table {
  width: 100%;
}

/* 网络加载状态样式 */
.network-loading {
  padding: 20px 0;
}

/* 网络错误信息样式 */
.network-error {
  margin-bottom: 20px;
}

/* 网络空状态样式 */
.network-empty {
  padding: 50px 0;
  text-align: center;
}

/* 升级进度对话框样式 */
.upgrade-info {
  margin-bottom: 20px;
  text-align: center;
}

.progress-container {
  margin-top: 20px;
}

.progress-details {
  margin-top: 10px;
  text-align: center;
  font-size: 14px;
  color: #606266;
}

/* 右键菜单样式 */
.context-menu {
  position: fixed;
  z-index: 1000;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  min-width: 150px;
  max-height: 60vh;
  overflow-y: auto;
}

.context-menu-item {
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: background-color 0.3s;
}

.context-menu-item:hover {
  background-color: #f5f7fa;
}

.context-menu-item .el-icon {
  margin-right: 8px;
  font-size: 14px;
}

.context-menu-item span {
  font-size: 14px;
  color: #303133;
}

/* 文件项样式 */
.file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 4px 0;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 16px;
}

.file-size {
  width: 100px;
  text-align: right;
  margin-right: 16px;
  color: #909399;
}

.file-date {
  width: 180px;
  text-align: right;
  color: #909399;
  font-size: 12px;
}

.no-files {
  padding: 40px 0;
  text-align: center;
}

/* 设备选择对话框样式 */
.device-selection-container {
  /* 移除容器的滚动条，使用 table 自身的滚动控制 */
}

.device-selection-table {
  width: 100%;
}

/* 离线设备行样式 */
.device-selection-table :deep(.device-offline-row) {
  background-color: #f5f7fa !important;
  color: #909399 !important;
  opacity: 0.6;
}

.device-selection-table :deep(.device-offline-row:hover) {
  background-color: #f0f2f5 !important;
}

.device-selection-table :deep(.device-offline-row .el-checkbox__input) {
  cursor: not-allowed;
}

/* 上传进度覆盖层 */
.upload-progress-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.9);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.upload-progress-dialog {
  background-color: #fff;
  padding: 30px;
  border-radius: 10px;
  text-align: center;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.upload-progress-text {
  margin-top: 15px;
  font-size: 14px;
  color: #606266;
}

.xterm .xterm-viewport {
  overflow: hidden;
}

.el-table .cell {
  display: flex;
  align-items: center;
}

.create-dialog-container-mode-title {
  text-align: center;
  color: red;
  padding-bottom: 15px;
  font-size: 18px;
}

.el-radio-group {
  display: block !important;
}

.batch-slot-checkbox-group .el-checkbox-button__inner {
  width: 100%;
  padding: 8px 0;
  border-radius: 4px !important;
  border: 2px solid #dcdfe6;
}

.batch-slot-checkbox-group .slot-blue .el-checkbox-button__inner {
  border-color: #409EFF;
  color: #409EFF;
}

.batch-slot-checkbox-group .slot-yellow .el-checkbox-button__inner {
  border-color: #E6A23C;
  color: #E6A23C;
}

.batch-slot-checkbox-group .slot-red .el-checkbox-button__inner {
  border-color: #F56C6C;
  color: #F56C6C;
}

.batch-slot-checkbox-group .slot-gray .el-checkbox-button__inner {
  border-color: #909399;
  color: #909399;
}

.batch-slot-checkbox-group .el-checkbox-button.is-checked .el-checkbox-button__inner {
  background-color: #409EFF;
  border-color: #409EFF;
  color: #fff;
  box-shadow: none;
}

/* 公告弹窗样式 */
.announcement-dialog .el-dialog__header {
  /* background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); */
  /* padding: 20px; */
  /* margin: 0; */
  /* border-radius: 8px 8px 0 0; */
  text-align: center;
}

.announcement-dialog :deep(.el-dialog__title) {
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  width: 100%;
  display: block;
  text-align: center;
}

.announcement-dialog :deep(.el-dialog__headerbtn) {
  display: none;
}

.announcement-dialog :deep(.el-dialog__body) {
  padding: 30px 20px;
}

.announcement-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.announcement-icon {
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 20%, 50%, 80%, 100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-10px);
  }
  60% {
    transform: translateY(-5px);
  }
}

.announcement-text {
  font-size: 15px;
  line-height: 1.8;
  color: #606266;
  white-space: pre-wrap;
  text-align: center;
  padding: 0 10px;
  max-height: 400px;
  overflow-y: auto;
}

.announcement-footer {
  display: flex;
  justify-content: center;
  padding: 10px 0;
}

.announcement-footer .el-button {
  min-width: 140px;
  position: relative;
  font-size: 16px;
  padding: 12px 30px;
  border-radius: 6px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s ease;
}

.announcement-footer .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.countdown-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 24px;
  margin-left: 8px;
  padding: 0 8px;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  backdrop-filter: blur(10px);
  animation: pulse 1s infinite;
}

.el-badge__content.is-fixed {
  top: 10px !important;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}
</style>
