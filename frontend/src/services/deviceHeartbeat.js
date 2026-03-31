
// ========== 设备心跳检测相关 ==========

import {
  UpdateMonitoredDevices,
  StartDeviceHeartbeat,
  GetDevicesStatus,
  GetDeviceStatus,
  ResetAllDevicesOffline
} from '../../bindings/edgeclient/app.js';

// 更新需要监控的设备列表，names 为 {ip: name} 映射
async function updateMonitoredDevices(deviceIPs, names = {}) {
  try {
    console.log('[心跳] 更新监控设备列表:', deviceIPs);
    await UpdateMonitoredDevices(deviceIPs, names);
    return { success: true };
  } catch (error) {
    console.error('[心跳] 更新监控设备列表失败:', error);
    throw error;
  }
}

// 启动设备心跳检测服务
async function startDeviceHeartbeat() {
  try {
    console.log('[心跳] 启动设备心跳检测服务');
    await StartDeviceHeartbeat();
    return { success: true };
  } catch (error) {
    console.error('[心跳] 启动心跳检测失败:', error);
    throw error;
  }
}

// 获取所有设备状态
async function getDevicesStatus() {
  try {
    const statusMap = await GetDevicesStatus();
    return statusMap;
  } catch (error) {
    console.error('[心跳] 获取设备状态失败:', error);
    throw error;
  }
}

// 获取单个设备状态
async function getDeviceStatus(deviceIP) {
  try {
    const status = await GetDeviceStatus(deviceIP);
    return status;
  } catch (error) {
    console.error('[心跳] 获取设备状态失败:', error);
    throw error;
  }
}

// 重置所有设备为离线状态(应用启动时调用)
async function resetAllDevicesOffline() {
  try {
    console.log('[心跳] 重置所有设备为离线状态');
    await ResetAllDevicesOffline();
    return { success: true };
  } catch (error) {
    console.error('[心跳] 重置设备状态失败:', error);
    throw error;
  }
}

export { updateMonitoredDevices, startDeviceHeartbeat, getDevicesStatus, getDeviceStatus, resetAllDevicesOffline };
