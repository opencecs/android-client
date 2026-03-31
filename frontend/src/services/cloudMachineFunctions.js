// 云机管理公共函数库
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'
import { DeleteContainer, DiscoverDevicesManually } from '../../bindings/edgeclient/app'
import { getDeviceAddr } from './api.js'

// 删除云机容器
const handleDeleteContainer = async (container) => {
  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      `确定要删除云机 "${container.name || container.ID}" 吗？删除后数据将无法恢复。`,
      '删除云机',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'danger'
      }
    )

    // 获取设备信息
    const deviceIp = container.deviceIp
    const deviceVersion = container.deviceVersion || 'v3'
    
    // 确保 containerName 是字符串
    let containerName = ''
    if (typeof container.name === 'string') {
      containerName = container.name
    } else if (typeof container.ID === 'string') {
      containerName = container.ID
    } else if (container.name && typeof container.name === 'object') {
      // 如果 name 是对象，尝试获取其 name 属性或转为字符串
      containerName = container.name.name || container.name.id || String(container.name)
    } else {
      containerName = String(container.name || container.ID || '')
    }

    if (!deviceIp || !containerName) {
      ElMessage.error('无法获取设备信息或容器名称')
      return { success: false, error: '无法获取设备信息或容器名称' }
    }

    // 根据设备版本确定API端口
    const port = deviceVersion === 'v3' ? '8000' : '81'
    
    // 尝试使用已保存的密码
    const savedPassword = getDevicePassword(deviceIp)
    let headers = {}
    
    if (savedPassword) {
      // 添加认证头
      const auth = btoa(`admin:${savedPassword}`)
      headers = {
        'Authorization': `Basic ${auth}`
      }
    }

    // 根据设备版本调用不同的API
    let response
    if (deviceVersion === 'v3') {
      // V3设备删除API：使用 /android?name=containerName 格式
      const apiUrl = `http://${getDeviceAddr(deviceIp)}/android?name=${encodeURIComponent(containerName)}`
      console.log('使用V3设备删除API:', apiUrl)
      response = await axios.delete(apiUrl, {
        headers: headers
      })
    } else {
      // V0-V2设备：使用Wails IPC调用后端函数
      console.log('使用Wails IPC调用后端DeleteContainer函数')
      const result = await DeleteContainer(deviceIp, deviceVersion, containerName)
      console.log('删除容器成功，返回数据:', result)
      
      // V0-V2设备返回模拟数据
      if (result && result.success !== false) {
        ElMessage.success(`云机 "${containerName}" 删除成功`)
        
        // 触发容器列表刷新
        if (typeof refreshContainerList === 'function') {
          refreshContainerList()
        }
        
        // 触发设备信息刷新
        if (typeof refreshDeviceInfo === 'function') {
          refreshDeviceInfo(deviceIp)
        }
        
        return { success: true }
      } else {
        const errorMsg = result?.message || '删除失败'
        ElMessage.error(`删除云机失败: ${errorMsg}`)
        return { success: false, error: errorMsg }
      }
    }

    // 检查响应结果（仅适用于V3设备）
    if (response && response.data && response.data.code === 0) {
      ElMessage.success(`云机 "${containerName}" 删除成功`)
      
      // 触发容器列表刷新
      if (typeof refreshContainerList === 'function') {
        refreshContainerList()
      }
      
      // 触发设备信息刷新
      if (typeof refreshDeviceInfo === 'function') {
        refreshDeviceInfo(deviceIp)
      }
      
      return { success: true }
    } else {
      const errorMsg = response?.data?.message || '删除失败'
      ElMessage.error(`删除云机失败: ${errorMsg}`)
      return { success: false, error: errorMsg }
    }

  } catch (error) {
    if (error === 'cancel') {
      // 用户取消了操作
      console.log('用户取消了删除操作')
      ElMessage.info('已取消删除操作')
      return { success: false, canceled: true }
    }
    
    console.error('删除云机失败:', error)
    
    // 处理认证错误
    if (error.response?.status === 401 || error.response?.data?.code === 61) {
      // 显示认证对话框
      const authResult = await showAuthDialog({ ip: container.deviceIp, version: container.deviceVersion })
      if (authResult) {
        // 重新尝试删除
        return await handleDeleteContainer(container)
      }
      return { success: false, error: '认证失败' }
    }
    
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    ElMessage.error(`删除云机失败: ${errorMsg}`)
    return { success: false, error: errorMsg }
  }
}

// 批量删除云机
const handleBatchDeleteContainers = async (containers) => {
  if (!containers || containers.length === 0) {
    ElMessage.warning('没有选中的云机')
    return { success: false, error: '没有选中的云机' }
  }

  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      `确定要删除选中的 ${containers.length} 个云机吗？删除后数据将无法恢复。`,
      '批量删除云机',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'danger'
      }
    )

    // 按设备分组，避免频繁切换设备
    const containersByDevice = {}
    containers.forEach(container => {
      const deviceIp = container.deviceIp
      if (!containersByDevice[deviceIp]) {
        containersByDevice[deviceIp] = []
      }
      containersByDevice[deviceIp].push(container)
    })

    let successCount = 0
    let failCount = 0
    const failedContainers = []

    // 按设备逐个处理
    for (const [deviceIp, deviceContainers] of Object.entries(containersByDevice)) {
      for (const container of deviceContainers) {
        try {
          const result = await handleDeleteContainer(container)
          if (result.success) {
            successCount++
          } else {
            failCount++
            failedContainers.push({
              container: container.name || container.ID,
              error: result.error
            })
          }
        } catch (error) {
          failCount++
          failedContainers.push({
            container: container.name || container.ID,
            error: error.message || '未知错误'
          })
        }
      }
    }

    // 显示操作结果
    if (successCount > 0) {
      ElMessage.success(`成功删除 ${successCount} 个云机`)
    }
    if (failCount > 0) {
      const failDetails = failedContainers.map(f => `${f.container}: ${f.error}`).join('\n')
      ElMessage.error(`删除失败 ${failCount} 个云机:\n${failDetails}`)
    }

    return {
      success: failCount === 0,
      successCount,
      failCount,
      failedContainers
    }

  } catch (error) {
    if (error === 'cancel') {
      console.log('用户取消了批量删除操作')
      ElMessage.info('已取消批量删除操作')
      return { success: false, canceled: true }
    }
    
    console.error('批量删除云机失败:', error)
    ElMessage.error(`批量删除失败: ${error.message || '未知错误'}`)
    return { success: false, error: error.message }
  }
}

// 获取设备密码（需要从主组件传入）
let getDevicePassword = () => ''

// 显示认证对话框（需要从主组件传入）
let showAuthDialog = () => Promise.resolve(null)

// 刷新容器列表（需要从主组件传入）
let refreshContainerList = () => {}

// 刷新设备信息（需要从主组件传入）
let refreshDeviceInfo = () => {}

// 设置依赖函数
const setDependencies = (dependencies) => {
  if (dependencies.getDevicePassword) {
    getDevicePassword = dependencies.getDevicePassword
  }
  if (dependencies.showAuthDialog) {
    showAuthDialog = dependencies.showAuthDialog
  }
  if (dependencies.refreshContainerList) {
    refreshContainerList = dependencies.refreshContainerList
  }
  if (dependencies.refreshDeviceInfo) {
    refreshDeviceInfo = dependencies.refreshDeviceInfo
  }
}

const handleManualDeviceDiscovery = async (ips) => {
  try {
    if (!ips || ips.trim() === '') {
      ElMessage.warning('请输入设备IP地址')
      return { success: false, error: '未提供IP地址' }
    }

    const result = await DiscoverDevicesManually(ips)
    console.log('DiscoverDevicesManually result:', result)
    
    if (result == null) {
      ElMessage.error('发现设备失败')
      return { success: false, error: '发现设备失败' }
    }

    // Wails 返回的是数组 [devices, failedIPs]
    const [devices, failedIPs] = result
    console.log('devices:', devices, 'failedIPs:', failedIPs)

    if (devices && devices.length > 0) {
      // 添加到设备列表
      for (const device of devices) {
        if (typeof window.addDiscoveredDevice === 'function') {
          window.addDiscoveredDevice(device)
        }
      }
      
      if (failedIPs && failedIPs.length > 0) {
        ElMessage.warning(`成功添加 ${devices.length} 个设备，${failedIPs.length} 个设备无响应: ${failedIPs.join(', ')}`)
        return {
          success: true,
          addedCount: devices.length,
          failedIPs: failedIPs,
          devices: devices
        }
      } else {
        ElMessage.success(`成功添加 ${devices.length} 个设备`)
        return {
          success: true,
          addedCount: devices.length,
          failedIPs: [],
          devices: devices
        }
      }
    } else {
      if (failedIPs && failedIPs.length > 0) {
        ElMessage.error(`所有设备均无响应: ${failedIPs.join(', ')}`)
        return {
          success: false,
          addedCount: 0,
          failedIPs: failedIPs,
          devices: [],
          error: '设备无响应'
        }
      } else {
        ElMessage.error('未发现任何设备')
        return {
          success: false,
          addedCount: 0,
          failedIPs: [],
          devices: [],
          error: '未发现设备'
        }
      }
    }
  } catch (error) {
    console.error('手动发现设备失败:', error)
    ElMessage.error(`发现设备失败: ${error.message || '未知错误'}`)
    return { success: false, error: error.message || '未知错误' }
  }
}

export {
  handleDeleteContainer,
  handleBatchDeleteContainers,
  handleManualDeviceDiscovery,
  setDependencies
}