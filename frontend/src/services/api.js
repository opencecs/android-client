// API服务文件 - 处理设备发现和API调用
import axios from 'axios';
import { ElMessage } from 'element-plus';
import {
  DiscoverDevices,
  GetContainers,
  StartContainer,
  StopContainer,
  DeleteContainer,
  CreateContainer,
  GetImages,
  GetVPCProxies,
  GetVPCHosts,
  GetDockerNetworks,
  CreateMacvlanNetwork,
  DeleteDockerNetwork,
  UpdateDockerNetwork,
  CheckMytSdkContainer,
  LoadImageToDevice,
  CreateMytSdkContainer,
  CreateV0V2Device,
  PullDockerImage,
  GetImagePullProgress,
  SetDevicePassword,
  CloseDevicePassword,
  UpdateDevicePasswords,
  StartProjectionWindow,
  CloseProjectionWindow,
  SetDeviceGPS,
  CleanupProjectionWindows,
  CleanupProjectionProcesses,
  GetAnnouncement,
  UpdateMonitoredDevices,
  StartDeviceHeartbeat,
  GetDevicesStatus,
  GetDeviceStatus,
  GetAndroidCacheVersions,
  GetAndroidContainersList,
  TriggerAndroidRefresh,
  GetScreenshotVersions,
  GetScreenshots,
  ClearScreenshotCache
} from '../../bindings/edgeclient/app';

// 设备发现相关
const DISCOVERY_PORT = 7678;
const DISCOVERY_MESSAGE = 'lgcloud';
const DISCOVERY_TIMEOUT = 2000;

// API端口配置
const API_PORTS = {
  'v0-v2': 81,
  'v3': 8000
};

/**
 * 返回设备的 host:port 字符串。
 * 若 ip 已含端口（如 "1.2.3.4:8187"），直接返回；否则追加默认端口 8000。
 */
export function getDeviceAddr(ip) {
  if (!ip) return ip
  // 判断是否已含端口：IPv4 形如 "a.b.c.d:port"，IPv6 形如 "[...]:port"
  const lastColon = ip.lastIndexOf(':')
  if (lastColon === -1) return ip + ':8000'
  const afterColon = ip.slice(lastColon + 1)
  // afterColon 全为数字则说明已含端口
  if (/^\d+$/.test(afterColon)) return ip
  return ip + ':8000'
}

// 设备信息结构
class DeviceInfo {
  constructor(ip, type, id, name, version = 'v0') {
    this.ip = ip;
    this.type = type;
    this.id = id;
    this.name = name;
    this.version = version;
    this.isOnline = true;
    this.lastSeen = new Date();
  }
}

// 从localStorage获取设备缓存
function getCachedDevices() {
  try {
    const cached = localStorage.getItem('deviceCache');
    if (cached) {
      const devices = JSON.parse(cached);
      // 将字符串日期转换为Date对象
      return devices.map(device => ({
        ...device,
        lastSeen: new Date(device.lastSeen)
      }));
    }
  } catch (error) {
    console.error('读取设备缓存失败:', error);
  }
  return [];
}

// 保存设备缓存到localStorage
function saveDeviceCache(devices) {
  try {
    localStorage.setItem('deviceCache', JSON.stringify(devices));
  } catch (error) {
    console.error('保存设备缓存失败:', error);
  }
}

// 云机列表内存缓存
export const containersMemoryCache = new Map();

// 保存云机列表到内存缓存
function saveContainersToMemoryCache(deviceIp, containers) {
  containersMemoryCache.set(deviceIp, {
    data: containers,
    timestamp: Date.now()
  });
}

// 从内存缓存获取云机列表
function getContainersFromMemoryCache(deviceIp) {
  return containersMemoryCache.get(deviceIp)?.data || null;
}

// 解析设备版本
function parseDeviceVersion(deviceName) {
  console.log('parseDeviceVersion called with deviceName:', deviceName);
  if (deviceName.includes('_v3')) return 'v3';
  if (deviceName.includes('_v2')) return 'v2';
  if (deviceName.includes('_10')) return 'v1';
  if (deviceName.includes('_')) return 'v0';
  return 'v0';
}

// 设备发现函数 - 实现增量更新和缓存
async function discoverDevices() {
  try {
    // 从缓存加载设备列表
    const cachedDevices = getCachedDevices();

    // 使用Wails IPC调用后端函数，增加超时时间到10秒
    console.log('使用Wails IPC调用后端DiscoverDevices函数，超时时间10秒');

    // 调用后端函数
    const discoveredDevices = await DiscoverDevices();
    console.log('设备发现请求成功，返回数据:', discoveredDevices);

    // 处理返回的设备数据
    const onlineDevices = discoveredDevices.map(d => new DeviceInfo(d.ip, d.type, d.id, d.name, d.version));

    // 创建IP到设备的映射，便于快速查找
    const onlineDeviceMap = new Map(onlineDevices.map(device => [device.ip, device]));

    // 增量更新设备列表
    const updatedDevices = [];
    const updatedDeviceIps = new Set();

    // 1. 更新缓存中存在且当前在线的设备
    cachedDevices.forEach(cachedDevice => {
      if (onlineDeviceMap.has(cachedDevice.ip)) {
        // 设备在线，更新设备信息
        const onlineDevice = onlineDeviceMap.get(cachedDevice.ip);
        updatedDevices.push(onlineDevice);
        updatedDeviceIps.add(cachedDevice.ip);
        onlineDeviceMap.delete(cachedDevice.ip); // 从在线设备映射中移除，剩下的就是新设备
      } else {
        // 设备离线，更新状态
        cachedDevice.isOnline = false;
        cachedDevice.lastSeen = new Date();
        updatedDevices.push(cachedDevice);
        updatedDeviceIps.add(cachedDevice.ip);
      }
    });

    // 2. 添加新发现的设备
    onlineDeviceMap.forEach(newDevice => {
      updatedDevices.push(newDevice);
      updatedDeviceIps.add(newDevice.ip);
    });

    // 3. 保存更新后的设备列表到缓存
    saveDeviceCache(updatedDevices);

    return updatedDevices;
  } catch (error) {
    console.error('设备发现请求失败:', error);

    // 失败时从缓存加载设备列表，并将所有设备标记为离线
    const cachedDevices = getCachedDevices();
    if (cachedDevices.length > 0) {
      console.log('使用缓存设备数据，并标记为离线');
      const offlineDevices = cachedDevices.map(device => ({
        ...device,
        isOnline: false,
        lastSeen: new Date()
      }));
      saveDeviceCache(offlineDevices);
      return offlineDevices;
    }

    // 缓存为空时返回模拟数据
    console.log('使用模拟设备数据作为 fallback');

    const mockDevices = [
    ];

    saveDeviceCache(mockDevices);
    return mockDevices;
  }
}

// 仅发现当前在线的设备（不使用缓存，专门用于设备扫描）
async function discoverOnlineDevicesOnly() {
  try {
    console.log('开始扫描当前在线设备（不使用缓存）');

    // 调用后端函数获取当前在线的设备
    const discoveredDevices = await DiscoverDevices();
    console.log('设备扫描成功，发现设备数:', discoveredDevices.length);

    // 处理返回的设备数据
    const onlineDevices = discoveredDevices.map(d => new DeviceInfo(d.ip, d.type, d.id, d.name, d.version));

    return onlineDevices;
  } catch (error) {
    console.error('设备扫描失败:', error);
    throw error; // 抛出错误，让调用者处理
  }
}

// 获取设备API端口
function getDevicePort(device) {
  if (device.version === 'v3') {
    return API_PORTS['v3'];
  }
  return API_PORTS['v0-v2'];
}

// 获取设备API基础URL
function getDeviceApiUrl(device, endpoint = '') {
  // 使用 getDeviceAddr 智能处理：已含端口的 IP 不会再追加默认端口
  const addr = getDeviceAddr(device.ip)
  return `http://${addr}${endpoint}`;
}

// 从本地存储获取设备密码
function getDevicePassword(deviceIP) {
  try {
    const passwords = JSON.parse(localStorage.getItem('devicePasswords') || '{}');
    return passwords[deviceIP] || null;
  } catch (error) {
    console.error('获取设备密码失败:', error);
    return null;
  }
}

// 保存设备密码到本地存储并同步到后端
async function saveDevicePassword(deviceIP, password) {
  try {
    const passwords = JSON.parse(localStorage.getItem('devicePasswords') || '{}');
    passwords[deviceIP] = password;
    localStorage.setItem('devicePasswords', JSON.stringify(passwords));

    // 同步到后端
    if (typeof UpdateDevicePasswords === 'function') {
      await UpdateDevicePasswords(passwords);
    } else {
      console.warn('[密码同步] ⚠️ UpdateDevicePasswords 函数不存在');
    }
  } catch (error) {
    console.error('保存设备密码失败:', error);
  }
}

// 移除设备密码并同步到后端
async function removeDevicePassword(deviceIP) {
  try {
    const passwords = JSON.parse(localStorage.getItem('devicePasswords') || '{}');
    delete passwords[deviceIP];
    localStorage.setItem('devicePasswords', JSON.stringify(passwords));

    // 同步到后端
    if (typeof UpdateDevicePasswords === 'function') {
      await UpdateDevicePasswords(passwords);
    } else {
      console.warn('[密码同步] ⚠️ UpdateDevicePasswords 函数不存在');
    }
  } catch (error) {
    console.error('移除设备密码失败:', error);
  }
}

// 通用API调用函数 - 使用Wails IPC接口
async function callDeviceApi(device, method, endpoint, data = null) {
  try {
    // 检查是否需要认证
    const password = getDevicePassword(device.ip);

    // 使用Wails IPC调用后端函数
    console.log(`使用Wails IPC调用后端API: deviceIP=${device.ip}, version=${device.version}, endpoint=${endpoint}, method=${method}`);

    // 检查后端是否暴露了CallDeviceApi函数
    if (CallDeviceApi) {
      // 调用Wails IPC接口 - 调用通用的设备API调用函数
      const result = await CallDeviceApi(device.ip, device.version, method, endpoint, data, password);
      console.log('API调用成功，返回数据:', result);
      return result;
    } else {
      // 如果后端没有暴露CallDeviceApi函数，返回模拟数据
      console.log('后端未暴露CallDeviceApi函数，使用模拟数据');

      // 针对删除镜像的endpoint返回模拟成功数据
      if (endpoint.startsWith('/android/image?imageUrl=') && method === 'DELETE') {
        // V3 API: 删除本地镜像
        return { code: 0, message: 'OK', data: null };
      }

      // 默认返回空数据
      return {};
    }
  } catch (error) {
    console.error(`API调用失败: ${device.ip}${endpoint}`, error);
    // 返回错误信息
    return { error: 'API调用失败', message: error.message };
  }
}

// 获取容器列表
async function getContainers(device, password = null) {
  try {
    // 使用提供的密码或从本地存储获取
    let usedPassword = password || getDevicePassword(device.ip);

    // 确保 usedPassword 是一个字符串
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    // 使用Wails IPC调用后端函数，设置5秒超时
    console.log('使用Wails IPC调用后端GetContainers函数，5秒超时');

    // 创建AbortController来实现超时控制
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5秒超时

    const result = await GetContainers(device.ip, device.version, usedPassword);
    clearTimeout(timeoutId); // 清除超时定时器

    console.log('获取容器列表成功，返回数据:', result);

    // 检查是否认证失败
    if (result.code === 61 && result.message === 'Authentication Failed') {
      console.log('认证失败，需要重新认证');
      // 抛出认证错误
      throw new Error('Authentication Failed');
    }

    // 不使用缓存，直接返回最新结果
    return result;
  } catch (error) {
    console.error('获取容器列表失败:', error);

    // 如果是认证失败，重新抛出错误
    if (error.message === 'Authentication Failed') {
      throw error;
    }

    // 其他错误，返回空数据
    console.log('获取容器列表失败，返回空数据');

    if (device.version === 'v3') {
      // V3 API格式空数据
      return { data: { list: [] } };
    } else {
      // V0-V2 Docker API格式空数据
      return [];
    }
  }
}

// ========== 安卓容器列表后台轮询接口 ==========

// 获取所有设备的缓存版本号（极轻量，2秒轮询用）
// 返回 { [ip]: versionTimestamp(UnixMilli) }
async function getAndroidCacheVersions() {
  try {
    return await GetAndroidCacheVersions()
  } catch (e) {
    console.error('[安卓轮询] GetAndroidCacheVersions 失败:', e)
    return {}
  }
}

// 批量获取指定设备的完整容器缓存
// ips: string[] - 空数组返回所有设备
// 返回 { [ip]: { list, version, status, error, failCount } }
async function getAndroidContainersList(ips = []) {
  try {
    return await GetAndroidContainersList(ips)
  } catch (e) {
    console.error('[安卓轮询] GetAndroidContainersList 失败:', e)
    return {}
  }
}

// 手动触发指定设备立即刷新（操作后调用，不阻塞）
async function triggerAndroidRefresh(ips = []) {
  if (!ips || ips.length === 0) return
  try {
    await TriggerAndroidRefresh(ips)
  } catch (e) {
    console.error('[安卓轮询] TriggerAndroidRefresh 失败:', e)
  }
}

// ========== 安卓容器截图后台缓存接口 ==========

// 获取所有设备的截图版本号（前端 500ms 轮询用，极轻量）
// 返回 { [ip]: versionTimestamp(UnixMilli) }
async function getScreenshotVersions() {
  try {
    return await GetScreenshotVersions()
  } catch (e) {
    return {}
  }
}

// 获取指定设备下所有容器的最新截图 base64
// 返回 { ["ip_containerName"]: "data:image/jpeg;base64,..." }
async function getScreenshots(ip) {
  try {
    // Go 端 GetScreenshots(ips []string)，传入单个 IP 的数组
    return await GetScreenshots(ip ? [ip] : [])
  } catch (e) {
    console.error('[截图轮询] GetScreenshots 失败:', e)
    return {}
  }
}

// 清空指定容器的截图缓存（重启/重置时调用）
// containerName 为空时清空该设备所有容器截图
async function clearScreenshotCache(ip, containerName = '') {
  try {
    // Go 端 ClearScreenshotCache(ips []string)，ips 为设备 IP 数组
    await ClearScreenshotCache(ip ? [ip] : [])
  } catch (e) {
    console.error('[截图轮询] ClearScreenshotCache 失败:', e)
  }
}

// 启动容器
async function startContainer(device, containerId, password = null) {
  try {
    let usedPassword = password || getDevicePassword(device.ip);

    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    console.log('使用Wails IPC调用后端StartContainer函数');
    const result = await StartContainer(device.ip, device.version, containerId, usedPassword);
    console.log('启动容器返回数据:', result);

    if (result.code === 61 && result.message === 'Authentication Failed') {
      console.log('认证失败，需要重新认证');
      throw new Error('Authentication Failed');
    }

    if (result.success === false) {
      console.error('启动容器失败:', result.message);
      throw new Error(result.message || '启动容器失败');
    }

    return result;
  } catch (error) {
    console.error('启动容器失败:', error);
    throw error;
  }
}

// 停止容器
async function stopContainer(device, containerId, password = null) {
  try {
    let usedPassword = password || getDevicePassword(device.ip);

    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    console.log('使用Wails IPC调用后端StopContainer函数');
    const result = await StopContainer(device.ip, device.version, containerId, usedPassword);
    console.log('停止容器返回数据:', result);

    if (result.code === 61 && result.message === 'Authentication Failed') {
      console.log('认证失败，需要重新认证');
      throw new Error('Authentication Failed');
    }

    if (result.success === false) {
      console.error('停止容器失败:', result.message);
      throw new Error(result.message || '停止容器失败');
    }

    return result;
  } catch (error) {
    console.error('停止容器失败:', error);
    throw error;
  }
}

// 删除容器
async function deleteContainer(device, containerId, password = null) {
  try {
    // 确保 containerId 是字符串
    let safeContainerId = ''
    if (typeof containerId === 'string') {
      safeContainerId = containerId
    } else if (containerId && typeof containerId === 'object') {
      safeContainerId = containerId.name || containerId.id || String(containerId)
    } else {
      safeContainerId = String(containerId || '')
    }

    if (!safeContainerId) {
      throw new Error('容器ID无效')
    }

    if (device.version === 'v3') {
      const apiUrl = getDeviceApiUrl(device, `/android?name=${encodeURIComponent(safeContainerId)}`);
      console.log('使用HTTP API调用V3删除容器:', apiUrl);

      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {};

      if (usedPassword) {
        const auth = btoa(`admin:${usedPassword}`);
        headers = {
          'Authorization': `Basic ${auth}`
        };
      }

      const response = await axios.delete(apiUrl, { headers: headers });
      console.log('删除容器返回数据:', response.data);

      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        console.log('认证失败，需要重新认证');
        throw new Error('Authentication Failed');
      }

      if (response.data.success === false) {
        console.error('删除容器失败:', response.data.message);
        throw new Error(response.data.message || '删除容器失败');
      }

      return response.data;
    } else {
      console.log('使用Wails IPC调用后端DeleteContainer函数');
      const result = await DeleteContainer(device.ip, device.version, containerId);
      console.log('删除容器返回数据:', result);

      if (result.success === false) {
        console.error('删除容器失败:', result.message);
        throw new Error(result.message || '删除容器失败');
      }

      return result;
    }
  } catch (error) {
    console.error('删除容器失败:', error);
    throw error;
  }
}

// 创建容器
async function createContainer(device, params) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端CreateContainer函数', device.ip, device.version, params);
    const result = await CreateContainer(device.ip, device.version, params);
    console.log('创建容器成功，返回数据:', result);
    return result;
  } catch (error) {
    console.error('创建容器失败:', error);
    // 失败时抛出错误，不返回模拟的成功结果
    throw error;
  }
}

// 获取镜像列表
async function getImages(device, password = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端GetImages函数');
    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }
    const result = await GetImages(device.ip, device.version, usedPassword);
    console.log('获取镜像列表成功，返回数据:', result);
    return result;
  } catch (error) {
    console.error('获取镜像列表失败:', error);
    // 失败时返回模拟数据
    return [];
  }
}

// 获取VPC代理列表
async function getVpcProxies() {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端GetVPCProxies函数');
    const result = await GetVPCProxies();
    console.log('获取VPC代理列表成功，返回数据:', result);
    return result;
  } catch (error) {
    console.error('获取VPC代理列表失败:', error);
    // 失败时返回模拟数据
    return [
      { id: 'proxy-1', name: '代理服务器1', ip: '10.0.0.1' },
      { id: 'proxy-2', name: '代理服务器2', ip: '10.0.0.2' }
    ];
  }
}

// 获取VPC主机列表
async function getVpcHosts() {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端GetVPCHosts函数');
    const result = await GetVPCHosts();
    console.log('获取VPC主机列表成功，返回数据:', result);
    return result;
  } catch (error) {
    console.error('获取VPC主机列表失败:', error);
    // 失败时返回模拟数据
    return [
    ];
  }
}

// 启动投屏 (使用 Wails V3 多窗口)
function isWindowsPlatform() {
  if (typeof navigator === 'undefined') {
    return false;
  }
  return /windows/i.test(navigator.userAgent || '');
}


function formatInstanceName(name) {
  if (!name) return '';
  const parts = name.split('_');
  if (parts.length > 2) {
    return parts[parts.length - 1];
  }
  return name;
}

function getMappedPort(portBindings, key) {
  if (!portBindings || !Array.isArray(portBindings[key]) || portBindings[key].length === 0) {
    return 0;
  }
  const hostPort = portBindings[key][0] && portBindings[key][0].HostPort;
  const parsed = parseInt(hostPort, 10);
  return Number.isFinite(parsed) ? parsed : 0;
}

// 辅助函数：按 deviceIp 查找端口映射表，支持 IP 前缀模糊匹配
function findPortMap(deviceIp) {
  if (!deviceIp || !deviceIp.includes(':') || !window.openCecsPortMap) return null
  let portMap = window.openCecsPortMap.get(deviceIp)
  if (!portMap) {
    const ipPrefix = deviceIp.split(':')[0] + ':'
    for (const [key, map] of window.openCecsPortMap) {
      if (key.startsWith(ipPrefix)) {
        portMap = map
        break
      }
    }
  }
  return portMap
}


async function startProjection(device, containerInfo, customOrient = null) {
  try {
    console.log('[api.js] startProjection 被调用');
    console.log('[api.js] customOrient 参数:', customOrient);
    console.log('[api.js] device.ip:', device.ip);
    console.log('[api.js] containerInfo:', containerInfo);

    if (!containerInfo) {
      throw new Error('容器信息不完整');
    }

    const isWindows = isWindowsPlatform();
    let tcpPort = 0;
    let udpPort = 0;
    let controlPort = 0;

    // 判断是否是myt或macvlan网络（使用容器IP直连）
    const isMytOrMacvlan = containerInfo.NetworkName === 'myt' ||
      containerInfo.NetworkMode === 'myt' ||
      containerInfo.network === 'myt' ||
      containerInfo.networkName === 'myt' ||
      containerInfo.networkMode === 'myt';

    if (isWindows) {
      if (isMytOrMacvlan && containerInfo.ip) {
        // myt/macvlan网络：使用容器IP直连，端口固定为10000和10001
        tcpPort = 10000;
        controlPort = 10001;
      } else {
        // 普通网络：需要端口映射
        if (!containerInfo.portBindings) {
          throw new Error('容器端口绑定信息不完整');
        }
        tcpPort = getMappedPort(containerInfo.portBindings, '10000/tcp') || getMappedPort(containerInfo.portBindings, '10000/udp') || 10000;
        controlPort = getMappedPort(containerInfo.portBindings, '10001/tcp') || getMappedPort(containerInfo.portBindings, '10001/udp') || 10001;
        // OpenCecs 公网设备：将 HostPort 转换为公网端口
        const portMap = findPortMap(device.ip)
        if (portMap) {
          tcpPort = portMap.get(tcpPort) || tcpPort
          controlPort = portMap.get(controlPort) || controlPort
        }
      }
      if (!tcpPort || !controlPort) {
        throw new Error('无法获取视频或控制端口');
      }
    } else {
      if (!containerInfo.portBindings) {
        throw new Error('容器端口绑定信息不完整');
      }
      tcpPort = getMappedPort(containerInfo.portBindings, '10008/tcp');
      udpPort = getMappedPort(containerInfo.portBindings, '10008/udp');
      if (!tcpPort || !udpPort) {
        throw new Error('无法获取视频或控制端口');
      }
    }


    const containerID = containerInfo.ID || containerInfo.name || containerInfo.id;
    const containerName = containerInfo.name || containerInfo.Names?.[0] || containerID;

    const width = parseInt(containerInfo.width, 10) || 360;
    const height = parseInt(containerInfo.height, 10) || 640;
    // 如果传入了 customOrient，使用自定义值；否则根据宽高自动判断
    const orient = customOrient !== null ? customOrient : (width >= height ? 1 : 0);
    console.log('[api.js] width:', width, ', height:', height);
    console.log('[api.js] customOrient:', customOrient, ', 最终 orient:', orient);

    const shortName = formatInstanceName(containerName);
    const term = shortName || device.ip;

    // 判断使用哪个IP：myt/macvlan网络使用容器IP，其他使用设备IP
    let deviceIP = isMytOrMacvlan && containerInfo.ip ? containerInfo.ip : device.ip;
    // OpenCecs 公网设备：提取纯 IP
    if (deviceIP && deviceIP.includes(':')) deviceIP = deviceIP.split(':')[0]

    const result = await StartProjectionWindow({
      DeviceIP: deviceIP,
      TCPPort: tcpPort,
      UDPPort: isWindows ? controlPort : udpPort,
      ControlPort: isWindows ? controlPort : 0,
      Orient: orient,
      Term: term,
      ContainerID: containerID,
      ContainerName: containerName,
      Width: width,
      Height: height
    });

    if (result.success) {
      console.log('投屏窗口操作成功:', result);
      if (typeof ElMessage !== 'undefined') {
        if (result.focused) {
          ElMessage.info('投屏窗口已聚焦');
        } else {
          ElMessage.success(result.message);
        }
      }
    } else {
      console.error('投屏窗口创建失败:', result.message);
      if (typeof ElMessage !== 'undefined') {
        ElMessage.error('启动投屏失败: ' + result.message);
      }
    }

    return result;
  } catch (error) {
    console.error('启动投屏失败:', error);
    if (error.message && error.message.includes('端口映射信息不完整')) {
      if (typeof ElMessage !== 'undefined') {
        ElMessage.error('SDK版本过旧，请升级支持WebRTC版本');
      }
    }
    throw error;
  }
}

// 获取Docker网络列表

async function startProjectionBatchControl(device, containers, term = '批量投屏控制', customOrient = null) {
  try {
    if (!isWindowsPlatform()) {
      throw new Error('此功能仅支持 Windows 平台');
    }
    if (!Array.isArray(containers) || containers.length == 0) {
      throw new Error('没有要操作的容器');
    }

    const runningContainers = containers.filter(container => container.status === 'running');
    if (runningContainers.length == 0) {
      throw new Error('没有处于运行状态的容器');
    }

    const firstContainer = runningContainers[0];
    if (!firstContainer) {
      throw new Error('容器信息不完整');
    }

    const baseDeviceIp = (device && device.ip) || firstContainer.deviceIp || firstContainer.deviceIP || firstContainer.ip;
    if (!baseDeviceIp) {
      throw new Error('无法获取设备IP');
    }

    // 判断是否是myt或macvlan网络（使用容器IP直连）
    const isMytOrMacvlan = firstContainer.NetworkName === 'myt' ||
      firstContainer.NetworkMode === 'myt' ||
      firstContainer.network === 'myt' ||
      firstContainer.networkName === 'myt' ||
      firstContainer.networkMode === 'myt';

    let videoPort, controlPort;
    if (isMytOrMacvlan && firstContainer.ip) {
      // myt/macvlan网络：使用容器IP直连，端口固定为10000和10001
      videoPort = 10000;
      controlPort = 10001;
    } else {
      // 普通网络：需要端口映射
      if (!firstContainer.portBindings) {
        throw new Error('容器端口绑定信息不完整');
      }
      videoPort = getMappedPort(firstContainer.portBindings, '10000/tcp') || getMappedPort(firstContainer.portBindings, '10000/udp');
      controlPort = getMappedPort(firstContainer.portBindings, '10001/tcp') || getMappedPort(firstContainer.portBindings, '10001/udp');
      // OpenCecs 公网设备：将 HostPort 转换为公网端口
      const portMap = findPortMap(baseDeviceIp)
      if (portMap) {
        videoPort = portMap.get(videoPort) || videoPort
        controlPort = portMap.get(controlPort) || controlPort
      }
    }

    if (!videoPort || !controlPort) {
      throw new Error('无法获取视频或控制端口');
    }

    const listItems = [];
    for (const container of runningContainers) {
      // 判断当前容器是否是myt/macvlan网络
      const containerIsMyt = container.NetworkName === 'myt' ||
        container.NetworkMode === 'myt' ||
        container.network === 'myt' ||
        container.networkName === 'myt' ||
        container.networkMode === 'myt';

      let containerDeviceIp, cport;
      if (containerIsMyt && container.ip) {
        // myt/macvlan网络：使用容器IP + 固定端口10001
        containerDeviceIp = container.ip;
        cport = 10001;
      } else {
        // 普通网络：使用设备IP + 端口映射
        containerDeviceIp = container.deviceIp || container.deviceIP || baseDeviceIp;
        cport = container.portBindings
          ? (getMappedPort(container.portBindings, '10001/tcp') || getMappedPort(container.portBindings, '10001/udp'))
          : 0;
        // OpenCecs 公网设备：解析端口映射并提取纯 IP
        const portMap = findPortMap(containerDeviceIp)
        if (portMap && cport) {
          cport = portMap.get(cport) || cport
        }
        if (containerDeviceIp.includes(':')) {
          containerDeviceIp = containerDeviceIp.split(':')[0]
        }
      }

      if (!containerDeviceIp || !cport) {
        console.warn('跳过无效的容器:', container);
        continue;
      }
      listItems.push(`${containerDeviceIp}:${cport}`);
    }

    if (listItems.length == 0) {
      throw new Error('没有有效的容器列表');
    }

    const width = parseInt(firstContainer.width, 10) || 360;
    const height = parseInt(firstContainer.height, 10) || 640;
    // 如果传入了 customOrient，使用自定义值；否则根据宽高自动判断
    const orient = customOrient !== null ? customOrient : (width >= height ? 1 : 0);
    const resolvedTerm = term || '批量投屏控制';

    // 判断使用哪个IP：myt/macvlan网络使用容器IP，其他使用设备IP
    let deviceIP = isMytOrMacvlan && firstContainer.ip ? firstContainer.ip : baseDeviceIp;
    // OpenCecs 公网设备：提取纯 IP
    if (deviceIP && deviceIP.includes(':')) deviceIP = deviceIP.split(':')[0]

    const result = await StartProjectionWindow({
      DeviceIP: deviceIP,
      TCPPort: videoPort,
      UDPPort: controlPort,
      ControlPort: controlPort,
      Orient: orient,
      List: listItems.join('#'),
      Term: resolvedTerm,
      ContainerID: 'batch_control',
      ContainerName: resolvedTerm,
      Width: width,
      Height: height
    });

    if (result.success) {
      console.log('批量投屏控制成功:', result);
      if (typeof ElMessage !== 'undefined') {
        if (result.focused) {
          ElMessage.info('批量投屏控制窗口已聚焦');
        } else {
          ElMessage.success(result.message || '批量投屏控制成功');
        }
      }
    } else {
      console.error('批量投屏控制失败:', result.message);
      if (typeof ElMessage !== 'undefined') {
        ElMessage.error('批量投屏控制失败: ' + result.message);
      }
    }

    return result;
  } catch (error) {
    console.error('批量投屏控制失败:', error);
    if (typeof ElMessage !== 'undefined') {
      ElMessage.error('批量投屏控制失败: ' + (error.message || '未知错误'));
    }
    throw error;
  }
}

async function stopProjectionBatchControl() {
  try {
    const result = await CloseProjectionWindow('batch_control');
    if (result.success) {
      console.log('停止控制:', result);
      if (typeof ElMessage !== 'undefined') {
        ElMessage.success(result.message || '停止控制');
      }
    } else {
      console.error('停止控制:', result.message);
      if (typeof ElMessage !== 'undefined') {
        ElMessage.error('停止控制: ' + result.message);
      }
    }
    return result;
  } catch (error) {
    console.error('停止控制:', error);
    if (typeof ElMessage !== 'undefined') {
      ElMessage.error('停止控制: ' + (error.message || '停止控制'));
    }
    throw error;
  }
}


async function getDockerNetworks(device, password = null) {
  try {
    // 使用Wails IPC调用后端函数，设置5秒超时
    console.log('使用Wails IPC调用后端GetDockerNetworks函数，5秒超时');

    // 创建AbortController来实现超时控制
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5秒超时

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await GetDockerNetworks(device.ip, device.version, usedPassword);
    clearTimeout(timeoutId); // 清除超时定时器

    console.log('获取Docker网络列表成功，返回数据:', result);

    return result;
  } catch (error) {
    console.error('获取Docker网络列表失败:', error);

    // 失败时返回空数据
    console.log('获取Docker网络列表失败，返回空数据');
    return [];
  }
}

// 创建Docker网络
async function createDockerNetwork(device, networkConfig, password = null) {
  try {
    // 使用Wails IPC调用后端函数，设置5秒超时
    console.log('使用Wails IPC调用后端CreateMacvlanNetwork函数，5秒超时');

    // 创建AbortController来实现超时控制
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5秒超时

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await CreateMacvlanNetwork(device.ip, device.version, networkConfig, usedPassword);
    clearTimeout(timeoutId); // 清除超时定时器

    console.log('创建Docker网络成功，返回数据:', result);

    return result;
  } catch (error) {
    console.error('创建Docker网络失败:', error);

    // 失败时抛出错误
    throw new Error(`创建网络失败: ${error.message}`);
  }
}

// 删除Docker网络
async function deleteDockerNetwork(device, networkID, password = null) {
  try {
    // 使用Wails IPC调用后端函数，设置5秒超时
    console.log('使用Wails IPC调用后端DeleteDockerNetwork函数，5秒超时');

    // 创建AbortController来实现超时控制
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5秒超时

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await DeleteDockerNetwork(device.ip, device.version, networkID, usedPassword);
    clearTimeout(timeoutId); // 清除超时定时器

    console.log('删除Docker网络成功，返回数据:', result);

    return result;
  } catch (error) {
    console.error('删除Docker网络失败:', error);

    // 失败时抛出错误
    throw new Error(`删除网络失败: ${error.message}`);
  }
}

// 更新Docker网络
async function updateDockerNetwork(device, networkID, networkConfig, password = null) {
  try {
    // 使用Wails IPC调用后端函数，设置5秒超时
    console.log('使用Wails IPC调用后端UpdateDockerNetwork函数，5秒超时');

    // 创建AbortController来实现超时控制
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5秒超时

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await UpdateDockerNetwork(device.ip, device.version, networkID, networkConfig, usedPassword);
    clearTimeout(timeoutId); // 清除超时定时器

    console.log('更新Docker网络成功，返回数据:', result);

    return result;
  } catch (error) {
    console.error('更新Docker网络失败:', error);

    // 失败时抛出错误
    throw new Error(`更新网络失败: ${error.message}`);
  }
}

// 检查myt_sdk容器是否存在
async function checkMytSdkContainer(device, password = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端CheckMytSdkContainer函数');

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await CheckMytSdkContainer(device.ip, device.version, usedPassword);
    console.log('检查myt_sdk容器结果:', result);
    return result;
  } catch (error) {
    console.error('检查myt_sdk容器失败:', error);
    throw error;
  }
}

// 加载镜像到设备
async function loadImageToDevice(device, imagePath, password = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端LoadImageToDevice函数');

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await LoadImageToDevice(device.ip, imagePath, device.version, usedPassword);
    console.log('加载镜像结果:', result);
    return result;
  } catch (error) {
    console.error('加载镜像失败:', error);
    throw error;
  }
}

// 创建myt_sdk容器
async function createMytSdkContainer(device, password = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端CreateMytSdkContainer函数');

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await CreateMytSdkContainer(device.ip, device.version, usedPassword);
    console.log('创建myt_sdk容器结果:', result);
    return result;
  } catch (error) {
    console.error('创建myt_sdk容器失败:', error);
    throw error;
  }
}

// 创建v0-v2设备
async function createV0V2Device(device, createParams = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端CreateV0V2Device函数');
    const result = await CreateV0V2Device(device.ip, createParams);
    console.log('创建v0-v2设备结果:', result);
    return result;
  } catch (error) {
    console.error('创建v0-v2设备失败:', error);
    throw error;
  }
}

// 拉取Docker镜像
async function pullDockerImage(device, imageUrl, password = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端PullDockerImage函数');

    let usedPassword = password || getDevicePassword(device.ip);
    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const result = await PullDockerImage(device.ip, imageUrl, device.version, usedPassword);
    console.log('拉取Docker镜像结果:', result);
    return result;
  } catch (error) {
    console.error('拉取Docker镜像失败:', error);
    throw error;
  }
}

// 获取镜像拉取进度
async function getImagePullProgress() {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端GetImagePullProgress函数');
    const result = await GetImagePullProgress();
    console.log('获取镜像拉取进度结果:', result);
    return result;
  } catch (error) {
    console.error('获取镜像拉取进度失败:', error);
    throw error;
  }
}

// 设置设备密码
async function setDevicePassword(device, password, currentPassword = null) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端SetDevicePassword函数');
    const result = await SetDevicePassword(device.ip, password, currentPassword);
    console.log('设置设备密码成功，返回数据:', result);
    return result;
  } catch (error) {
    console.error('设置设备密码失败:', error);
    throw error;
  }
}

// 关闭设备密码
async function closeDevicePassword(device, currentPassword) {
  try {
    // 使用Wails IPC调用后端函数
    console.log('使用Wails IPC调用后端CloseDevicePassword函数');
    const result = await CloseDevicePassword(device.ip, currentPassword);
    console.log('关闭设备密码成功，返回数据:', result);
    return result;
  } catch (error) {
    console.error('关闭设备密码失败:', error);
    throw error;
  }
}

// 切换机型
async function switchPhoneModel(device, containerId, modelInfo, password = null) {
  try {
    if (device.version === 'v3') {
      // V3设备：使用/android/switchModel API
      const apiUrl = getDeviceApiUrl(device, `/android/switchModel`);
      console.log('使用HTTP API调用V3切换机型:', apiUrl);

      // 尝试使用已保存的密码或提供的密码
      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {
        'Content-Type': 'application/json'
      };

      if (usedPassword) {
        // 添加认证头
        const auth = btoa(`admin:${usedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      // 准备请求数据
      const data = {
        name: containerId,
        modelId: '',
        localModel: '',
        modelStatic: ''
      };

      if (typeof modelInfo === 'object' && modelInfo !== null && modelInfo.value) {
        if (modelInfo.type === 'local') {
          data.localModel = modelInfo.value;
        } else if (modelInfo.type === 'backup') {
          data.modelStatic = modelInfo.value;
        } else {
          data.modelId = modelInfo.value;
        }
      } else {
        data.modelId = modelInfo;
      }

      // 使用axios发送POST请求
      const response = await axios.post(apiUrl, data, { headers: headers });
      console.log('V3切换机型成功，返回数据:', response.data);

      // 检查是否认证失败
      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        console.log('认证失败，需要重新认证');
        throw new Error('Authentication Failed');
      }

      return response.data;
    } else {
      // V0-V2设备：不支持切换机型
      console.log('V0-V2设备不支持切换机型');
      return { error: '该设备版本不支持切换机型功能' };
    }
  } catch (error) {
    console.error('切换机型失败:', error);
    throw error;
  }
}

// 重置安卓容器
async function resetAndroidContainer(device, containerId, password = null, start = true) {
  try {
    try {
      await CloseProjectionWindow(containerId);
    } catch (e) {
      console.warn(`关闭投屏窗口失败:`, e);
    }
    
    if (device.version === 'v3') {
      // V3设备：使用/android PUT API重置
      // const apiUrl = getDeviceApiUrl(device, `/android`);
      let apiUrl
      if (device.androidType === 'V2') {
        apiUrl = getDeviceApiUrl(device, `/androidV2`);
      } else {
        apiUrl = getDeviceApiUrl(device, `/android`);
      }
      console.log('使用HTTP API调用V3重置安卓:', apiUrl);

      // 尝试使用已保存的密码或提供的密码
      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {
        'Content-Type': 'application/json'
      };

      if (usedPassword) {
        // 添加认证头
        const auth = btoa(`admin:${usedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      // 准备请求数据
      const data = {
        name: containerId,
        start: start
      };

      // 使用axios发送PUT请求
      const response = await axios.put(apiUrl, data, { headers: headers });
      console.log('V3重置安卓成功，返回数据:', response.data);

      // 检查是否认证失败
      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        console.log('认证失败，需要重新认证');
        throw new Error('Authentication Failed');
      }

      return response.data;
    } else {
      // V0-V2设备：不支持重置功能
      console.log('V0-V2设备不支持重置功能');
      return { error: '该设备版本不支持重置功能' };
    }
  } catch (error) {
    console.error('重置安卓失败:', error);
    throw error;
  }
}

// 重启安卓容器（V3设备使用 /android/restart POST API）
async function restartAndroidContainer(device, containerId, password = null) {
  try {
    try {
      await CloseProjectionWindow(containerId);
    } catch (e) {
      console.warn(`关闭投屏窗口失败:`, e);
    }
    
    if (device.version === 'v3') {
      // V3设备：使用/android/restart POST API重启
      const apiUrl = getDeviceApiUrl(device, `/android/restart`);
      console.log('使用HTTP API调用V3重启安卓:', apiUrl);

      // 尝试使用已保存的密码或提供的密码
      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {
        'Content-Type': 'application/json'
      };

      if (usedPassword) {
        // 添加认证头
        const auth = btoa(`admin:${usedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      // 准备请求数据
      const data = {
        name: containerId
      };

      // 使用axios发送POST请求
      const response = await axios.post(apiUrl, data, { headers: headers });
      console.log('V3重启安卓成功，返回数据:', response.data);

      // 检查是否认证失败
      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        console.log('认证失败，需要重新认证');
        throw new Error('Authentication Failed');
      }

      return response.data;
    } else {
      // V0-V2设备：使用stop+start方式重启
      console.log('V0-V2设备使用stop+start方式重启');
      await stopContainer(device, containerId, password);
      await new Promise(resolve => setTimeout(resolve, 1000));
      await startContainer(device, containerId, password);
      return { success: true };
    }
  } catch (error) {
    console.error('重启安卓失败:', error);
    throw error;
  }
}

// 获取设备版本信息
async function getDeviceVersionInfo(device, password = null) {
  try {
    if (device.version === 'v3') {
      // V3设备：调用/info接口获取版本信息
      const apiUrl = getDeviceApiUrl(device, `/info`);
      console.log('使用HTTP API调用V3获取版本信息:', apiUrl);

      // 尝试使用已保存的密码或提供的密码
      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {};

      if (usedPassword) {
        // 添加认证头
        const auth = btoa(`admin:${usedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      // 使用axios发送GET请求，设置1秒超时
      const response = await axios.get(apiUrl, { headers: headers, timeout: 1000 });
      console.log('V3获取版本信息成功，返回数据:', response.data);

      // 检查是否认证失败
      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        console.log('认证失败，需要重新认证');
        throw new Error('Authentication Failed');
      }

      // 🔧 统一数据类型：将版本号从number转换为string，避免后续类型不匹配
      if (response.data.data) {
        response.data.data.currentVersion = String(response.data.data.currentVersion);
        response.data.data.latestVersion = String(response.data.data.latestVersion);
      }

      return response.data;
    } else {
      // V0-V2设备：不支持版本信息查询
      console.log('V0-V2设备不支持版本信息查询');
      return { error: '该设备版本不支持版本信息查询功能' };
    }
  } catch (error) {
    console.error('获取设备版本信息失败:', error);
    throw error;
  }
}

// 获取设备升级API URL
function getDeviceUpgradeUrl(device) {
  if (device.version === 'v3') {
    return `http://${getDeviceAddr(device.ip)}/server/upgrade`;
  }
  return null;
}

// 升级设备
async function upgradeDevice(device, password = null) {
  try {
    if (device.version === 'v3') {
      // 尝试使用已保存的密码或提供的密码
      const usedPassword = password || getDevicePassword(device.ip);

      // 返回升级相关信息，供前端使用EventSource直接连接
      return {
        url: getDeviceUpgradeUrl(device),
        password: usedPassword
      };
    } else {
      // V0-V2设备：不支持升级功能
      console.log('V0-V2设备不支持升级功能');
      return { error: '该设备版本不支持升级功能' };
    }
  } catch (error) {
    console.error('获取升级信息失败:', error);
    throw error;
  }
}

// 删除本地镜像
async function deleteLocalImage(device, imageUrl, password = null) {
  try {
    // 使用callDeviceApi调用设备API
    const response = await callDeviceApi(device, 'DELETE', `/android/image?imageUrl=${encodeURIComponent(imageUrl)}`, null, password);
    console.log('deleteLocalImage', response);
    return response;
  } catch (error) {
    console.error('删除本地镜像失败:', error);
    throw error;
  }
}

async function getDeviceAllCloudMachines(deviceIP, name = '', running = true) {
  try {
    let usedPassword = getDevicePassword(deviceIP);

    if (typeof usedPassword !== 'string') {
      usedPassword = null;
    }

    const headers = {
      'Content-Type': 'application/json'
    };

    if (usedPassword) {
      const auth = btoa(`admin:${usedPassword}`);
      headers['Authorization'] = `Basic ${auth}`;
    }

    const url = `http://${getDeviceAddr(deviceIP)}/android?name=${encodeURIComponent(name)}&running=false`;
    console.log('调用设备云机列表接口:', url);

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    const response = await axios.get(url, {
      headers,
      signal: controller.signal,
      timeout: 10000
    });

    clearTimeout(timeoutId);

    console.log('获取设备所有云机成功:', response.data);
    return response.data;
  } catch (error) {
    console.error('获取设备所有云机失败:', error);
    throw error;
  }
}

// 重命名云机容器
async function renameAndroidContainer(device, name, newName, password = null) {
  try {
    if (device.version === 'v3') {
      const apiUrl = getDeviceApiUrl(device, `/android/rename`);
      console.log('使用HTTP API调用V3重命名云机:', apiUrl);

      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {
        'Content-Type': 'application/json'
      };

      if (usedPassword) {
        const auth = btoa(`admin:${usedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      const data = {
        name: name,
        newName: newName
      };

      const response = await axios.post(apiUrl, data, { headers: headers });
      console.log('V3重命名云机成功，返回数据:', response.data);

      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        throw new Error('Authentication Failed');
      }

      return response.data;
    } else {
      console.log('V0-V2设备不支持重命名功能');
      return { error: '该设备版本不支持重命名功能' };
    }
  } catch (error) {
    console.error('重命名云机失败:', error);
    throw error;
  }
}

// 设置MacVlanIP
async function setMacVlanIp(device, name, ip, password = null) {
  try {
    if (device.version === 'v3') {
      const apiUrl = getDeviceApiUrl(device, `/android/macvlan`);
      console.log('使用HTTP API调用V3设置MacVlanIP:', apiUrl);

      const usedPassword = password || getDevicePassword(device.ip);
      let headers = {
        'Content-Type': 'application/json'
      };

      if (usedPassword) {
        const auth = btoa(`admin:${usedPassword}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      const data = {
        name: name,
        ip: ip
      };

      const response = await axios.post(apiUrl, data, { headers: headers });
      console.log('V3设置MacVlanIP成功，返回数据:', response.data);

      if (response.data.code === 61 && response.data.message === 'Authentication Failed') {
        throw new Error('Authentication Failed');
      }

      return response.data;
    } else {
      console.log('V0-V2设备不支持设置MacVlanIP功能');
      return { error: '该设备版本不支持设置MacVlanIP功能' };
    }
  } catch (error) {
    console.error('设置MacVlanIP失败:', error);
    throw error;
  }
}

// 设置设备GPS
async function setDeviceGPS(host, port, deviceIP, language) {
  try {
    console.log('使用Wails IPC调用后端SetDeviceGPS函数', host, port, deviceIP, language);
    const result = await SetDeviceGPS(host, Number(port), deviceIP, language);
    console.log('设置设备GPS成功');
    return result;
  } catch (error) {
    console.error('设置设备GPS失败:', error);
    throw error;
  }
}

// 清理所有投屏窗口
async function cleanupProjectionWindows() {
  try {
    console.log('使用Wails IPC调用后端CleanupProjectionWindows函数');
    const result = await CleanupProjectionWindows();
    console.log('清理所有投屏窗口成功');
    return result;
  } catch (error) {
    console.error('清理所有投屏窗口失败:', error);
    throw error;
  }
}

// 清理所有投屏进程
async function cleanupProjectionProcesses() {
  try {
    console.log('使用Wails IPC调用后端CleanupProjectionProcesses函数');
    const result = await CleanupProjectionProcesses();
    console.log('清理所有投屏进程成功');
    return result;
  } catch (error) {
    console.error('清理所有投屏进程失败:', error);
    throw error;
  }
}

// 获取系统公告
async function getAnnouncement() {
  try {
    // 优先使用Go后端方法（如果bindings已生成）
    if (typeof GetAnnouncement === 'function') {
      console.log('使用Wails IPC调用后端GetAnnouncement函数');
      const result = await GetAnnouncement();
      console.log('获取系统公告成功:', result);
      return result;
    } else {
      // 降级到直接调用HTTP API
      console.log('GetAnnouncement绑定未生成，使用axios直接调用');
      const response = await axios.get('https://newapi.moyunteng.com/api/announcement');
      console.log('获取系统公告成功:', response.data);
      return response.data;
    }
  } catch (error) {
    console.error('获取系统公告失败:', error);
    return null;
  }
}

// 导出API服务
export {
  discoverDevices,
  discoverOnlineDevicesOnly,
  getContainers,
  startContainer,
  stopContainer,
  deleteContainer,
  createContainer,
  getImages,
  getVpcProxies,
  getVpcHosts,
  startProjection,
  startProjectionBatchControl,
  stopProjectionBatchControl,
  getDockerNetworks,
  createDockerNetwork,
  deleteDockerNetwork,
  updateDockerNetwork,
  getDeviceApiUrl,
  parseDeviceVersion,
  checkMytSdkContainer,
  loadImageToDevice,
  createMytSdkContainer,
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
  cleanupProjectionWindows,
  cleanupProjectionProcesses,
  getAnnouncement,
  getAndroidCacheVersions,
  getAndroidContainersList,
  triggerAndroidRefresh,
  getScreenshotVersions,
  getScreenshots,
  clearScreenshotCache
};