import { createI18n } from 'vue-i18n'

// 完整的内联语言包
const zhCN = {
  common: {
    confirm: '确认',
    cancel: '取消',
    save: '保存',
    delete: '删除',
    edit: '编辑',
    add: '添加',
    close: '关闭',
    search: '搜索',
    refresh: '刷新',
    loading: '加载中...',
    success: '成功',
    failed: '失败',
    warning: '警告',
    error: '错误',
    yes: '是',
    no: '否',
    operation: '操作',
    status: '状态',
    name: '名称',
    type: '类型',
    time: '时间',
    description: '描述',
    settings: '设置',
    upload: '上传',
    download: '下载',
    sortBy: '排序方式',
    createTimeSort: '创建时间',
    selectedItems: '已选择 {count} 项',
    batchDelete: '批量删除'
  },
  app: {
    title: '魔云腾-V3-客户端',
    description: 'ARM边缘计算设备管理客户端'
  },
  menu: {
    hostManagement: '主机',
    cloudManagement: '云机管理',
    modelManagement: '机型',
    instanceManagement: '实例',
    networkManagement: '网络',
    backupManagement: '备份',
    streamManagement: '流媒体',
    aiAssistant: 'AI助理',
    rpaAgent: 'RPA Agent',
    customerService: '客服',
    about: '关于',
    settings: '设置',
    update: '更新',
    logout: '退出'
  },
  device: {
    device: '设备',
    deviceList: '设备列表',
    deviceName: '设备名称',
    deviceIP: '设备IP',
    deviceStatus: '设备状态',
    addDevice: '添加设备',
    deleteDevice: '删除设备',
    online: '在线',
    offline: '离线'
  },
  cloudMachine: {
    cloudMachine: '云机',
    createCloudMachine: '创建云机',
    deleteCloudMachine: '删除云机',
    startCloudMachine: '启动云机',
    stopCloudMachine: '停止云机',
    renameCloudMachine: '重命名云机',
    switchBackup: '切换备份',
    running: '运行中',
    stopped: '已停止',
    selectedCount: '已选择 {count} 个',
    setStream: '设置推流',
    updateImage: '更新镜像',
    apiDetails: 'API详情',
    renameDevice: '修改云机名称',
    shake: '摇一摇',
    setGPS: '设置GPS',
    uploadGoogleCert: '上传google证书',
    restart: '重启',
    shutdown: '关闭',
    setS5Agent: '设置S5代理',
    closeS5Agent: '关闭安卓内SOCKS5代理',
    fileUpload: '上传文件',
    oneKeyNewDevice: '一键新机',
    switchModel: '切换机型',
    resetContainer: '重置容器',
    backupName: '备份名称',
    remark: '备注',
    created: '已创建',
    restarting: '重启中'
  },
  update: {
    checkUpdate: '检查更新',
    newVersionAvailable: '发现新版本',
    currentVersion: '当前版本',
    noUpdate: '已是最新版本',
    downloading: '下载中'
  },
  dialog: {
    addDeviceTitle: '添加设备',
    deleteConfirm: '确认删除',
    operationSuccess: '操作成功',
    operationFailed: '操作失败'
  },
  task: {
    taskQueue: '任务列表'
  }
}

const enUS = {
  common: {
    confirm: 'Confirm',
    cancel: 'Cancel',
    save: 'Save',
    delete: 'Delete',
    edit: 'Edit',
    add: 'Add',
    close: 'Close',
    search: 'Search',
    refresh: 'Refresh',
    loading: 'Loading...',
    success: 'Success',
    failed: 'Failed',
    warning: 'Warning',
    error: 'Error',
    yes: 'Yes',
    no: 'No',
    operation: 'Operation',
    status: 'Status',
    name: 'Name',
    type: 'Type',
    time: 'Time',
    description: 'Description',
    settings: 'Settings',
    upload: 'Upload',
    download: 'Download',
    sortBy: 'Sort By',
    createTimeSort: 'Create Time',
    selectedItems: '{count} selected',
    batchDelete: 'Batch Delete'
  },
  app: {
    title: 'MoYunTeng V3 Client',
    description: 'ARM Edge Computing Device Management Client'
  },
  menu: {
    hostManagement: 'Host Management',
    cloudManagement: 'Cloud Machine Management',
    modelManagement: 'Model Management',
    instanceManagement: 'Instance Management',
    networkManagement: 'Network Management',
    backupManagement: 'Backup Management',
    streamManagement: 'Stream Management',
    aiAssistant: 'AI Assistant',
    rpaAgent: 'RPA Agent',
    customerService: 'Customer Service',
    about: 'About',
    settings: 'Settings',
    update: 'Update',
    logout: 'Logout'
  },
  device: {
    device: 'Device',
    deviceList: 'Device List',
    deviceName: 'Device Name',
    deviceIP: 'Device IP',
    deviceStatus: 'Device Status',
    addDevice: 'Add Device',
    deleteDevice: 'Delete Device',
    online: 'Online',
    offline: 'Offline'
  },
  cloudMachine: {
    cloudMachine: 'Cloud Machine',
    createCloudMachine: 'Create Cloud Machine',
    deleteCloudMachine: 'Delete Cloud Machine',
    startCloudMachine: 'Start Cloud Machine',
    stopCloudMachine: 'Stop Cloud Machine',
    renameCloudMachine: 'Rename Cloud Machine',
    switchBackup: 'Switch Backup',
    running: 'Running',
    stopped: 'Stopped',
    selectedCount: '{count} selected',
    setStream: 'Set Stream',
    updateImage: 'Update Image',
    apiDetails: 'API Details',
    renameDevice: 'Rename Cloud Machine',
    shake: 'Shake',
    setGPS: 'Set GPS',
    uploadGoogleCert: 'Upload Google Certificate',
    restart: 'Restart',
    shutdown: 'Shutdown',
    setS5Agent: 'Set S5 Proxy',
    closeS5Agent: 'Close Android SOCKS5 Proxy',
    fileUpload: 'Upload File',
    oneKeyNewDevice: 'One-Click New Device',
    switchModel: 'Switch Model',
    resetContainer: 'Reset Container',
    backupName: 'Backup Name',
    remark: 'Remark',
    created: 'Created',
    restarting: 'Restarting'
  },
  update: {
    checkUpdate: 'Check for Updates',
    newVersionAvailable: 'New Version Available',
    currentVersion: 'Current Version',
    noUpdate: 'Already up to date',
    downloading: 'Downloading'
  },
  dialog: {
    addDeviceTitle: 'Add Device',
    deleteConfirm: 'Confirm Delete',
    operationSuccess: 'Operation Successful',
    operationFailed: 'Operation Failed'
  },
  task: {
    taskQueue: 'Task Queue'
  }
}

// 从localStorage获取保存的语言设置，默认为中文
const getDefaultLocale = () => {
  try {
    const savedLocale = localStorage.getItem('app-locale')
    if (savedLocale) {
      return savedLocale
    }
  } catch (e) {
    console.warn('Cannot access localStorage:', e)
  }
  
  return 'zh-CN'
}

const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  },
  globalInjection: true,
  missingWarn: false,
  fallbackWarn: false
})

// 切换语言的辅助函数
export const switchLocale = (locale) => {
  i18n.global.locale.value = locale
  try {
    localStorage.setItem('app-locale', locale)
  } catch (e) {
    console.warn('Cannot save to localStorage:', e)
  }
}

// 获取当前语言
export const getCurrentLocale = () => {
  return i18n.global.locale.value
}

export default i18n


