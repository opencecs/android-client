<template>
  <div class="opencecs-container">

    <!-- 顶部栏 -->
    <div class="top-bar">
      <div class="top-bar-title">{{ $t('common.cloudInstanceManagement') }}</div>
      <div class="top-right-actions">
        <!-- 未登录 -->
        <template v-if="!userInfo">
          <button class="btn-login" @click="loginDialogVisible = true">{{ $t('common.login') }}</button>
        </template>
        <!-- 已登录 -->
        <template v-else>
          <div class="user-info">
            <img v-if="userInfo.avatar" :src="userInfo.avatar" class="user-avatar" />
            <span class="user-name">{{ userInfo.username }}</span>
            <button class="btn-logout" @click="handleLogout">{{ $t('common.logout') }}</button>
          </div>
        </template>
      </div>
    </div>

    <!-- 实例列表（登录后显示） -->
    <div v-if="userInfo" class="instance-section">
      <!-- 统计卡片 -->
      <div class="stats-bar">
        <div class="stat-item">
          <span class="stat-label">{{ $t('common.allInstances') }}</span>
          <span class="stat-value">{{ instanceTotal }}</span>
        </div>
        <div class="stat-item running">
          <span class="stat-dot"></span>
          <span class="stat-label">{{ $t('common.running') }}</span>
          <span class="stat-value">{{ instanceStats.running }}</span>
        </div>
        <div class="stat-item stopped">
          <span class="stat-dot"></span>
          <span class="stat-label">{{ $t('common.stopped') }}</span>
          <span class="stat-value">{{ instanceStats.stopped }}</span>
        </div>
        <div class="stat-item expired">
          <span class="stat-dot"></span>
          <span class="stat-label">{{ $t('common.expired') }}</span>
          <span class="stat-value">{{ instanceStats.expired }}</span>
        </div>
        <button class="btn-refresh" :disabled="listLoading" @click="fetchInstances">
          {{ listLoading ? $t('common.refreshing') : $t('common.refresh') }}
        </button>
      </div>

      <!-- 加载中 -->
      <div v-if="listLoading" class="list-loading">{{ $t('common.loadingData') }}</div>
      <!-- 空状态 -->
      <div v-else-if="instanceList.length === 0" class="list-empty">{{ $t('common.noInstanceData') }}</div>

      <!-- 表格 -->
      <div v-else class="table-wrap">
        <table class="instance-table">
          <thead>
            <tr>
              <th>{{ $t('common.publicAddress') }}</th>
              <th>{{ $t('common.instanceNameLabel') }}</th>
              <th>{{ $t('common.coreBoard') }}</th>
              <th>{{ $t('common.memory') }}</th>
              <th>{{ $t('common.system') }}</th>
              <th>{{ $t('common.ipAddress') }}</th>
              <th>{{ $t('common.statusLabel') }}</th>
              <th>{{ $t('common.billingMode') }}</th>
              <th>{{ $t('common.expiryTime') }}</th>
              <th>{{ $t('common.operationLabel') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredInstanceList" :key="item.instance_id">
              <td class="monospace" :title="item.instance_id">{{ getPublicAddress(item.instance_id) }}</td>
              <td>{{ item.instance_name }}</td>
              <td>{{ item.board_name }}</td>
              <td>{{ item.memory }}GB</td>
              <td>{{ item.image_name }}</td>
              <td class="monospace">{{ item.ip_address }}</td>
              <td>
                <span>
                  {{ statusLabel(item.status) }}
                </span>
              </td>
              <td>{{ billingLabel(item.billing_mode) }}</td>
              <td :class="isExpireSoon(item.expire_at) ? 'expire-soon' : ''">
                {{ item.expire_at || '-' }}
              </td>
              <td class="action-cell">
                <button 
                  class="btn-action start" 
                  :disabled="item.status === 'running' || item.operating" 
                  @click="operateInstance(item, 'start')"
                >{{ $t('common.start') }}</button>
                <button 
                  class="btn-action stop" 
                  :disabled="item.status !== 'running' || item.operating" 
                  @click="operateInstance(item, 'stop')"
                >{{ $t('common.stop') }}</button>
                <button 
                  class="btn-action restart" 
                  :disabled="item.status !== 'running' || item.operating" 
                  @click="operateInstance(item, 'reboot')"
                >{{ $t('common.restart') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 未登录提示 -->
    <div v-else class="not-login-tip">
      <p>{{ $t('common.loginFirst') }}</p>
      <button class="btn-login" @click="loginDialogVisible = true">{{ $t('common.loginNow') }}</button>
    </div>

    <!-- ==================== 登录弹窗 ==================== -->
    <div v-if="loginDialogVisible" class="dialog-mask" @click.self="loginDialogVisible = false">
      <div class="dialog-box">
        <button class="dialog-close" @click="loginDialogVisible = false">✕</button>
        <h2 class="dialog-title">{{ $t('common.login') }}</h2>
        <p class="dialog-subtitle">{{ $t('common.welcomeOpenCecs') }}</p>

        <div class="login-tabs">
          <span :class="['login-tab', loginTab === 'phone' ? 'active' : '']" @click="switchTab('phone')">{{ $t('common.phoneTab') }}</span>
          <span :class="['login-tab', loginTab === 'account' ? 'active' : '']" @click="switchTab('account')">{{ $t('common.accountTab') }}</span>
        </div>

        <!-- 手机号 + 验证码 -->
        <div v-if="loginTab === 'phone'" class="login-form">
          <div class="form-item">
            <label>手机号 <span class="required">*</span></label>
            <input v-model="loginForm.phone" type="text" class="form-input" placeholder="请输入手机号" maxlength="20" />
          </div>
          <div class="form-item">
            <label>验证码 <span class="required">*</span></label>
            <div class="input-code-wrap">
              <input v-model="loginForm.code" type="text" class="form-input" placeholder="请输入验证码" maxlength="10" />
              <button class="btn-get-code" :disabled="codeCounting || sendingCode" @click="handleSendCode">
                <span v-if="sendingCode">发送中...</span>
                <span v-else-if="codeCounting">{{ codeCountdown }}s 后重发</span>
                <span v-else>获取验证码</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 账号 + 密码 -->
        <div v-if="loginTab === 'account'" class="login-form">
          <div class="form-item">
            <label>账号 <span class="required">*</span></label>
            <input v-model="loginForm.account" type="text" class="form-input" placeholder="请输入账号" />
          </div>
          <div class="form-item">
            <label>密码 <span class="required">*</span></label>
            <div class="input-password-wrap">
              <input v-model="loginForm.password" :type="showPassword ? 'text' : 'password'" class="form-input" placeholder="请输入密码" />
             <!-- <span class="eye-icon" @click="showPassword = !showPassword">{{ showPassword ? '👁' : '🙈' }}</span> -->
            </div>
          </div>
        </div>

        <div class="login-options">
          <label class="remember-me">
            <input type="checkbox" v-model="loginForm.remember" />
            <span>记住我</span>
          </label>
        </div>

        <div v-if="loginError" class="login-error">{{ loginError }}</div>

        <button class="btn-submit" :disabled="loginLoading" @click="handleLogin">
          <span v-if="loginLoading">登录中...</span>
          <span v-else>立即登录</span>
        </button>
      </div>
    </div>



  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

// Go IPC — 直接从 Wails bindings 导入（GetContainers 已确认在 bindings 中）
import { GetContainers } from '../../bindings/edgeclient/app'
import { ElMessage, ElMessageBox } from 'element-plus'

const BASE_URL = 'https://www.opencecs.com/api/v1'

// ==================== 用户 / 登录 ====================
const loginDialogVisible = ref(false)
const loginTab = ref('phone')
const showPassword = ref(false)
const loginLoading = ref(false)
const loginError = ref('')
const userInfo = ref(null)

const savedUser = localStorage.getItem('opencecs_user')
if (savedUser) {
  try { userInfo.value = JSON.parse(savedUser) } catch (e) { /* ignore */ }
}

const codeCounting = ref(false)
const sendingCode = ref(false)
const codeCountdown = ref(60)
let countdownTimer = null

const loginForm = ref({ phone: '', code: '', account: '', password: '', remember: false })

const switchTab = (tab) => { loginTab.value = tab; loginError.value = '' }

const handleSendCode = async () => {
  const phone = loginForm.value.phone.trim()
  if (!phone) { loginError.value = '请先输入手机号'; return }
  loginError.value = ''
  sendingCode.value = true
  try {
    const res = await fetch(`${BASE_URL}/auth/sms/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone, scene: 'login' })
    })
    const data = await res.json()
    if (data.code === 200) {
      codeCountdown.value = 60
      codeCounting.value = true
      countdownTimer = setInterval(() => {
        codeCountdown.value--
        if (codeCountdown.value <= 0) { clearInterval(countdownTimer); codeCounting.value = false }
      }, 1000)
    } else {
      loginError.value = data.message || '验证码发送失败，请重试'
    }
  } catch (e) {
    loginError.value = '网络错误，请检查连接后重试'
  } finally {
    sendingCode.value = false
  }
}

const handleLogin = async () => {
  loginError.value = ''
  if (loginTab.value === 'phone') {
    if (!loginForm.value.phone.trim()) { loginError.value = '请输入手机号'; return }
    if (!loginForm.value.code.trim()) { loginError.value = '请输入验证码'; return }
    loginLoading.value = true
    try {
      const res = await fetch(`${BASE_URL}/auth/login/phone`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone: loginForm.value.phone.trim(), code: loginForm.value.code.trim(), remember_me: loginForm.value.remember })
      })
      const data = await res.json()
      if (data.code === 200) { onLoginSuccess(data.data) }
      else { loginError.value = data.message || '登录失败，请检查手机号或验证码' }
    } catch (e) { loginError.value = '网络错误，请检查连接后重试' }
    finally { loginLoading.value = false }
  } else {
    if (!loginForm.value.account.trim()) { loginError.value = '请输入账号'; return }
    if (!loginForm.value.password) { loginError.value = '请输入密码'; return }
    loginLoading.value = true
    try {
      const res = await fetch(`${BASE_URL}/auth/login/account`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ account: loginForm.value.account.trim(), password: loginForm.value.password, remember_me: loginForm.value.remember })
      })
      const data = await res.json()
      if (data.code === 200) { onLoginSuccess(data.data) }
      else { loginError.value = data.message || '登录失败，请检查账号或密码' }
    } catch (e) { loginError.value = '网络错误，请检查连接后重试' }
    finally { loginLoading.value = false }
  }
}

const resetForm = () => {
  loginForm.value = { phone: '', code: '', account: '', password: '', remember: false }
  loginError.value = ''
  showPassword.value = false
  codeCounting.value = false
  clearInterval(countdownTimer)
}

const onLoginSuccess = (userData) => {
  if (userData?.token) localStorage.setItem('opencecs_token', userData.token)
  userInfo.value = {
    user_id: userData.user_id,
    username: userData.username,
    phone: userData.phone,
    avatar: userData.avatar,
    user_type: userData.user_type,
    expire_time: userData.expire_time
  }
  localStorage.setItem('opencecs_user', JSON.stringify(userInfo.value))
  loginDialogVisible.value = false
  resetForm()
  fetchInstances()
}

// 清除所有由 OpenCecs 注入到主机页面的设备
// clearCache=true 时同时清除端口映射缓存（退出登录时使用）
// clearCache=false 时保留缓存（刷新时使用，下次可从缓存恢复）
const removeAllOpenCecsDevices = (clearCache = true) => {
  // 收集所有已知的 OpenCecs 设备 IP（从内存 + localStorage 中恢复）
  const allKnownIps = new Set(addedDeviceIps)
  try {
    const raw = localStorage.getItem(PORT_MAP_STORAGE_KEY)
    if (raw) {
      for (const ip of Object.keys(JSON.parse(raw))) {
        allKnownIps.add(ip)
      }
    }
  } catch (e) { /* ignore */ }
  try {
    const raw2 = localStorage.getItem(ADDED_DEVICES_KEY)
    if (raw2) {
      for (const ip of JSON.parse(raw2)) {
        allKnownIps.add(ip)
      }
    }
  } catch (e) { /* ignore */ }

  console.log(`[OpenCecs 退出] 准备清理设备，已知IP: ${allKnownIps.size} 个`, [...allKnownIps])

  // 调用 App.vue 注册的专用清理函数（直接操作 devices 数组，确保万无一失）
  if (typeof window.removeAllOpenCecsDevicesFromHost === 'function') {
    const removed = window.removeAllOpenCecsDevicesFromHost([...allKnownIps])
    console.log(`[OpenCecs 退出] App.vue 清理完成，移除了 ${removed} 个设备`)
  } else {
    // 降级方案：逐个按 IP 移除
    console.warn('[OpenCecs 退出] removeAllOpenCecsDevicesFromHost 不可用，使用降级方案')
    if (typeof window.removeDevicesBySource === 'function') {
      window.removeDevicesBySource('opencecs')
    }
    if (allKnownIps.size > 0 && typeof window.removeDiscoveredDevice === 'function') {
      for (const ip of allKnownIps) {
        window.removeDiscoveredDevice(ip)
      }
    }
  }

  // 清理本组件内部状态
  addedDeviceIps.clear()
  deviceIpToInstanceMap.clear()
  window.openCecsPortMap?.clear()
  localStorage.removeItem(ADDED_DEVICES_KEY)
  
  if (clearCache) {
    // 退出登录时清除端口映射缓存
    localStorage.removeItem(PORT_MAP_STORAGE_KEY)
    localStorage.removeItem(INSTANCE_DEVICE_MAP_KEY)
    console.log('[OpenCecs] 已清除端口映射缓存')
  } else {
    console.log('[OpenCecs] 保留端口映射缓存（刷新模式）')
  }
}

const handleLogout = () => {
  console.log('========== [OpenCecs] handleLogout 被调用 ==========')
  console.log('[OpenCecs] addedDeviceIps:', [...addedDeviceIps])
  console.log('[OpenCecs] window.removeAllOpenCecsDevicesFromHost 是否可用:', typeof window.removeAllOpenCecsDevicesFromHost)
  if (typeof window.getDevicesList === 'function') {
    const allDevices = window.getDevicesList()
    console.log('[OpenCecs] 当前设备列表:', allDevices.map(d => ({ip: d.ip, name: d.name, source: d.source, id: d.id})))
  }
  
  // ⚠️ 先递增 setupVersion，终止所有后台异步操作（setupAllPortMappings 等）
  setupVersion++
  console.log(`[OpenCecs] 已递增 setupVersion 到 ${setupVersion}，终止后台异步操作`)
  
  userInfo.value = null
  instanceList.value = []
  instanceTotal.value = 0
  instanceStats.value = { running: 0, stopped: 0, expired: 0 }
  localStorage.removeItem('opencecs_token')
  localStorage.removeItem('opencecs_user')
  
  // 登出时清除已注入的设备
  removeAllOpenCecsDevices()
  
  console.log('========== [OpenCecs] handleLogout 完成 ==========')
  if (typeof window.getDevicesList === 'function') {
    const remaining = window.getDevicesList()
    console.log('[OpenCecs] 清理后设备列表:', remaining.map(d => ({ip: d.ip, name: d.name, source: d.source})))
  }
  
  // 延迟二次清理：防止已在执行中的异步操作在清理完成后又加回设备
  setTimeout(() => {
    console.log('[OpenCecs] 延迟二次清理...')
    removeAllOpenCecsDevices()
  }, 2000)
}

// ==================== 实例列表 ====================
const instanceList = ref([])
const instanceTotal = ref(0)
const instanceStats = ref({ running: 0, stopped: 0, expired: 0 })
const listLoading = ref(false)

// 记录由 OpenCecs 添加到主机列表的设备 IP，用于刷新时先清除旧设备
const addedDeviceIps = new Set()
// 记录 deviceIp → { instanceId, instance } 的精确映射，供容器创建后查找实例
const deviceIpToInstanceMap = new Map()
// 响应式版本号：每次端口映射更新时递增，触发 getPublicAddress 重新计算
const portMapVersion = ref(0)
// 全局端口映射表：deviceIp → Map<privatePort, publicPort>
// 供 App.vue 构建 URL 时将容器 HostPort 替换为公网端口
// 持久化到 localStorage，页面刷新后仍可用
const PORT_MAP_STORAGE_KEY = 'opencecs_port_map'
const ADDED_DEVICES_KEY = 'opencecs_added_devices'
// 实例到设备的映射缓存：instanceId → { publicIp, deviceIp }
const INSTANCE_DEVICE_MAP_KEY = 'opencecs_instance_device_map'
// 版本计数器：每次 fetchInstances 递增，setupAllPortMappings 检查是否过期
let setupVersion = 0

// 立即保存已注入的设备 IP 列表到 localStorage
const saveAddedDeviceIps = () => {
  try {
    localStorage.setItem(ADDED_DEVICES_KEY, JSON.stringify([...addedDeviceIps]))
  } catch (e) { /* ignore */ }
}

// 从 localStorage 恢复端口映射
function loadPortMapFromStorage() {
  try {
    const raw = localStorage.getItem(PORT_MAP_STORAGE_KEY)
    if (raw) {
      const obj = JSON.parse(raw)
      const map = new Map()
      for (const [deviceIp, ports] of Object.entries(obj)) {
        const portMap = new Map()
        for (const [priv, pub] of Object.entries(ports)) {
          portMap.set(Number(priv), Number(pub))
        }
        map.set(deviceIp, portMap)
      }
      console.log('[端口映射] 从 localStorage 恢复端口映射:', obj)
      return map
    }
  } catch (e) {
    console.warn('[端口映射] 恢复 localStorage 失败', e)
  }
  return new Map()
}

// 保存端口映射到 localStorage
function savePortMapToStorage() {
  try {
    const obj = {}
    for (const [deviceIp, portMap] of window.openCecsPortMap) {
      obj[deviceIp] = Object.fromEntries(portMap)
    }
    localStorage.setItem(PORT_MAP_STORAGE_KEY, JSON.stringify(obj))
    // 同步推送到 Go 后端（截图轮询需要使用端口映射）
    import('@wailsio/runtime').then(({ Call }) => {
      if (Call && Call.ByName) {
        Call.ByName('main.App.SetOpenCecsPortMap', obj)
        console.log('[端口映射] 已推送到 Go 后端')
      }
    }).catch(() => {})
  } catch (e) {
    console.warn('[端口映射] 保存 localStorage 失败', e)
  }
}

// 保存实例到设备的映射缓存到 localStorage
function saveInstanceDeviceMap(instanceId, publicIp, deviceIp) {
  try {
    const raw = localStorage.getItem(INSTANCE_DEVICE_MAP_KEY)
    const map = raw ? JSON.parse(raw) : {}
    map[instanceId] = { publicIp, deviceIp }
    localStorage.setItem(INSTANCE_DEVICE_MAP_KEY, JSON.stringify(map))
    console.log(`[端口映射缓存] 已保存实例设备映射: ${instanceId} → ${deviceIp}`)
  } catch (e) {
    console.warn('[端口映射缓存] 保存实例设备映射失败', e)
  }
}

// 从 localStorage 加载实例到设备的映射缓存
function loadInstanceDeviceMap() {
  try {
    const raw = localStorage.getItem(INSTANCE_DEVICE_MAP_KEY)
    if (raw) {
      const map = JSON.parse(raw)
      console.log('[端口映射缓存] 从 localStorage 恢复实例设备映射:', map)
      return map
    }
  } catch (e) {
    console.warn('[端口映射缓存] 恢复实例设备映射失败', e)
  }
  return {}
}

// 检查某个实例是否有缓存的端口映射
function hasCachedPortMapping(instanceId) {
  const instanceDeviceMap = loadInstanceDeviceMap()
  const cachedInfo = instanceDeviceMap[instanceId]
  if (!cachedInfo || !cachedInfo.deviceIp) return false
  // 同时检查端口映射表中是否有对应的 deviceIp 数据
  const portMapRaw = localStorage.getItem(PORT_MAP_STORAGE_KEY)
  if (!portMapRaw) return false
  try {
    const portMapObj = JSON.parse(portMapRaw)
    return !!portMapObj[cachedInfo.deviceIp] && Object.keys(portMapObj[cachedInfo.deviceIp]).length > 0
  } catch (e) {
    return false
  }
}

// 获取缓存的端口映射信息
function getCachedPortMapping(instanceId) {
  const instanceDeviceMap = loadInstanceDeviceMap()
  const cachedInfo = instanceDeviceMap[instanceId]
  if (!cachedInfo || !cachedInfo.deviceIp) return null
  const portMapRaw = localStorage.getItem(PORT_MAP_STORAGE_KEY)
  if (!portMapRaw) return null
  try {
    const portMapObj = JSON.parse(portMapRaw)
    const ports = portMapObj[cachedInfo.deviceIp]
    if (!ports || Object.keys(ports).length === 0) return null
    return { ...cachedInfo, ports }
  } catch (e) {
    return null
  }
}

if (!window.openCecsPortMap || window.openCecsPortMap.size === 0) {
  window.openCecsPortMap = loadPortMapFromStorage()
  // 启动时也推送到 Go 后端
  if (window.openCecsPortMap.size > 0) {
    savePortMapToStorage()
  }
}

// 只显示 image_name 包含 "myt"（不区分大小写）的实例
const filteredInstanceList = computed(() =>
  instanceList.value.filter(i => i.image_name?.toLowerCase().includes('myt'))
)

const fetchInstances = async () => {
  console.log('[OpenCecs] fetchInstances 被调用')
  const token = localStorage.getItem('opencecs_token')
  if (!token) {
    console.log('[OpenCecs] 无 token，跳过')
    return
  }
  listLoading.value = true

  // 递增版本号，使上一次仍在运行的 setupAllPortMappings 自动终止
  const currentVersion = ++setupVersion

  // 刷新前先移除上一次由 OpenCecs 添加的所有设备（保留端口映射缓存）
  removeAllOpenCecsDevices(false)

  try {
    const res = await fetch(`${BASE_URL}/cecs/instances?page=1&page_size=99999`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    const data = await res.json()
    if (data.code === 200) {
      const list = data.data.list || []
      instanceList.value = list
      // 只统计 image_name 包含 myt 的实例
      const mytList = list.filter(i => i.image_name?.toLowerCase().includes('myt'))
      instanceTotal.value = mytList.length
      instanceStats.value = {
        running: mytList.filter(i => i.status === 'running').length,
        stopped: mytList.filter(i => i.status === 'stopped').length,
        expired: mytList.filter(i => i.status === 'expired').length,
      }
      // 数据已获取，立即关闭加载状态，让用户看到列表
      listLoading.value = false

      const runningInstances = list.filter(i => i.status === 'running' && i.image_name?.toLowerCase().includes('myt'))
      console.log(`[端口映射] 获取到 ${list.length} 个实例, ${runningInstances.length} 个符合条件且运行中`)
      if (runningInstances.length > 0) {
        // 端口映射在后台异步进行，不阻塞列表显示
        setupAllPortMappings(runningInstances, currentVersion).catch(e => {
          console.error('[端口映射] setupAllPortMappings 异常:', e)
        })
      }
    } else {
      // API 返回非 200，清空数据并移除已添加的设备
      console.warn('[OpenCecs] API 返回非200:', data.code, data.message)
      instanceList.value = []
      instanceTotal.value = 0
      instanceStats.value = { running: 0, stopped: 0, expired: 0 }
      removeAllOpenCecsDevices(false)
    }
  } catch (e) {
    console.error('获取实例列表失败', e)
    // 网络错误，清空数据并移除已添加的设备
    instanceList.value = []
    instanceTotal.value = 0
    instanceStats.value = { running: 0, stopped: 0, expired: 0 }
    removeAllOpenCecsDevices(false)
  } finally {
    listLoading.value = false
  }
}

const statusLabel = (status) => {
  const map = { running: '运行中', stopped: '已停止', expired: '已过期', creating: '创建中' }
  return map[status] || status
}
const billingLabel = (mode) => {
  const map = { monthly: '包月', hourly: '按小时', yearly: '包年' }
  return map[mode] || mode
}
const isExpireSoon = (expireAt) => {
  if (!expireAt) return false
  const diff = new Date(expireAt).getTime() - Date.now()
  return diff > 0 && diff < 3 * 24 * 3600 * 1000 // 3天内到期
}

// 根据实例ID获取公网IP+端口地址，用于替代实例ID显示
const getPublicAddress = (instanceId) => {
  // 读取响应式版本号，使 Vue 感知端口映射变化后重新渲染
  void portMapVersion.value
  // 先从内存中的 deviceIpToInstanceMap 反查（精确匹配）
  for (const [deviceIp, instance] of deviceIpToInstanceMap) {
    if (instance.instance_id === instanceId) {
      return deviceIp
    }
  }
  // 再从 localStorage 缓存查找
  const instanceDeviceMap = loadInstanceDeviceMap()
  const cachedInfo = instanceDeviceMap[instanceId]
  if (cachedInfo && cachedInfo.deviceIp) {
    return cachedInfo.deviceIp
  }
  // 无映射信息时回退显示实例ID
  return instanceId
}

const operateInstance = async (item, action) => {
  const actionName = action === 'start' ? '启动' : action === 'stop' ? '停止' : '重启'
  try {
    await ElMessageBox.confirm(`确定要${actionName}实例 ${item.instance_name} 吗？`, '操作确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    item.operating = true
    const token = localStorage.getItem('opencecs_token')
    const res = await fetch(`${BASE_URL}/cecs/instances/${item.instance_id}/${action}`, {
      method: 'POST',
      headers: { 
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success(`${actionName}指令已发送`)
      setTimeout(() => {
        fetchInstances()
      }, 2000)
    } else {
      ElMessage.error(`${actionName}失败: ${data.message || '未知错误'}`)
    }
  } catch (e) {
    if (e !== 'cancel') {
      console.error(`实例${actionName}错误`, e)
      ElMessage.error(`请求失败，请检查网络后再试`)
    }
  } finally {
    if (item) item.operating = false
  }
}

// 暴露 init 方法，供父组件在切换到该标签页时调用
// 切换标签页时自动刷新页面数据
const init = () => {
  console.log('[OpenCecs] init 被调用，切换标签页触发刷新')
  if (userInfo.value) {
    fetchInstances()
  }
}

// 根据设备公网 IP（如 1.2.3.4:5001）查找对应的 OpenCecs 实例
// 先精确匹配，再按 IP 前缀模糊匹配（因为 8000 端口映射每次可能分配不同的公网端口）
const getInstanceByDeviceIp = (deviceIp) => {
  let result = deviceIpToInstanceMap.get(deviceIp)
  if (!result && deviceIp && deviceIp.includes(':')) {
    const ipPrefix = deviceIp.split(':')[0] + ':'
    for (const [key, instance] of deviceIpToInstanceMap) {
      if (key.startsWith(ipPrefix)) {
        result = instance
        break
      }
    }
  }
  return result || null
}



/**
 * 同步本地登录 token 到设备
 * 检查 localStorage 中是否有 'token'（非 opencecs_token），
 * 如果有则调用设备的 /user/sync 接口将 token 同步过去。
 * @param {string} deviceIp - 设备的公网 IP:Port（通过 8000 端口映射获得）
 */
const syncTokenToDevice = async (deviceIp) => {
  const mytToken = localStorage.getItem('token')
  if (!mytToken) {
    console.log(`[Token同步] 本地无登录 token，跳过同步 (${deviceIp})`)
    return
  }
  const syncUrl = `http://${deviceIp}/user/sync?mytToken=${encodeURIComponent(mytToken)}`
  try {
    console.log(`[Token同步] 开始同步 token 到设备 (${deviceIp})...`)
    if (typeof window.goHttpRequest === 'function') {
      const result = await window.goHttpRequest({
        Method: 'GET',
        URL: syncUrl,
        Body: '',
        Headers: {}
      })
      if (result && result.success) {
        console.log(`[Token同步] ✅ token 同步成功 (${deviceIp}):`, JSON.stringify(result.body).substring(0, 200))
      } else {
        console.warn(`[Token同步] ⚠️ token 同步失败 (${deviceIp}):`, result)
      }
    } else {
      // 降级方案：直接通过 fetch 调用
      const res = await fetch(syncUrl)
      const data = await res.json()
      console.log(`[Token同步] ✅ token 同步成功 (${deviceIp}, fetch):`, data)
    }
  } catch (e) {
    console.warn(`[Token同步] ❌ token 同步异常 (${deviceIp}):`, e.message)
  }
}

/**
 * 统一处理所有实例的端口映射（优化版 — 支持缓存）
 * 
 * 流程：
 *  1. 先检查 localStorage 缓存，有缓存的实例直接恢复（跳过网络请求）
 *  2. 无缓存的实例走完整流程：删除旧映射 → 创建 8000 映射 → 查询容器 → 批量创建容器映射
 * 
 * 优化：
 *  1. 缓存命中时跳过所有网络请求，大幅加速刷新
 *  2. 合并获取+删除为单步并行操作
 *  3. 并行获取容器列表和设备信息
 *  4. 尽早注入设备到主机列表
 *
 * @param {Array} instances - 运行中的实例列表
 * @param {number} version - 版本号，用于检测是否过期
 */
const setupAllPortMappings = async (instances, version) => {
  console.log(`[端口映射] setupAllPortMappings v${version} 开始, 处理 ${instances.length} 个实例`)
  const token = localStorage.getItem('opencecs_token')
  if (!token) return

  const isStale = () => setupVersion !== version

  // ===== 阶段0 & 1：验证缓存与获取概览 =====
  const cachedInstances = []
  const uncachedInstances = []
  const instanceInfos = []

  await Promise.all(instances.map(async (instance) => {
    const instanceId = instance.instance_id
    const cached = getCachedPortMapping(instanceId)

    let overview = null
    let listData = null
    let ovData = null

    // 无论有无缓存，都拉取线上真实的端口映射列表和概览来验证/获取
    try {
      const [ovRes, listRes] = await Promise.all([
        fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings/overview`, { headers: { 'Authorization': `Bearer ${token}` } }),
        fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings?page=1&page_size=99999`, { headers: { 'Authorization': `Bearer ${token}` } })
      ])
      ovData = await ovRes.json()
      listData = await listRes.json()
    } catch (e) {
      console.error(`[端口映射] 请求线上概览或规则失败 (${instanceId})`, e)
    }

    let isCacheValid = false
    // 校验缓存数据的有效性（和线上返回的映射列表是否完全一致）
    if (cached && cached.ports && ovData?.code === 200 && listData?.code === 200) {
      overview = ovData.data
      const list = listData.data?.list || []
      const onlinePorts = {}
      for (const p of list) {
        if (p.private_port && p.public_port) {
          onlinePorts[p.private_port] = p.public_port
        }
      }
      
      const cachedKeys = Object.keys(cached.ports)
      const onlineKeys = Object.keys(onlinePorts)
      if (cachedKeys.length > 0 && cachedKeys.length === onlineKeys.length) {
        // 判断私有端口对公有端口是否一一对应
        const allMatch = cachedKeys.every(k => Number(cached.ports[k]) === Number(onlinePorts[k]))
        const publicIp = overview?.nat_public_ip
        const cachedPublicIp = cached.deviceIp.split(':')[0]
        // 还要保证公有IP一致
        if (allMatch && publicIp === cachedPublicIp) {
          isCacheValid = true
          console.log(`[端口映射] 缓存校验通过 (${instanceId}): IP=${publicIp}, 端口数=${cachedKeys.length}`)
        } else {
          console.warn(`[端口映射] 缓存校验未通过 (${instanceId}): allMatch=${allMatch}, publicIp=${publicIp}, cached=${cachedPublicIp}`)
        }
      } else {
        console.warn(`[端口映射] 缓存校验未通过 (${instanceId}): 长度不一致 cached=${cachedKeys.length}, online=${onlineKeys.length}`)
      }
    }

    if (isCacheValid) {
      cachedInstances.push({ instance, cached })
    } else {
      uncachedInstances.push(instance)
      if (ovData?.code === 200) {
        overview = ovData.data
      } else if (ovData) {
        console.warn(`[端口映射] 概览API非200 (${instanceId}): code=${ovData.code}, msg=${ovData.message}`)
      }

      // 如果属于缓存无效或无缓存的，走原本彻底删除重建的逻辑：先删除旧有所有列表
      const list = (listData?.code === 200 ? listData.data?.list : null) || []
      const mappingIds = list.map(p => p.mapping_id).filter(Boolean)

      if (mappingIds.length > 0) {
        try {
          await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings/batch`, {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
            body: JSON.stringify({ mapping_ids: mappingIds })
          })
          console.log(`[端口映射] 删除旧映射完成 (${instanceId}), 共 ${mappingIds.length} 条`)
        } catch (e) {
          console.warn(`[端口映射] 删除旧映射失败 (${instanceId})`, e)
        }
      }
      instanceInfos.push({ instance, instanceId, overview })
    }
  }))

  console.log(`[端口映射] 缓存命中且校验通过: ${cachedInstances.length} 个, 需要彻底重建: ${uncachedInstances.length} 个`)

  // ===== 处理缓存命中的实例：直接恢复，不发网络请求 =====
  for (const { instance, cached } of cachedInstances) {
    if (isStale()) return
    const { deviceIp, ports } = cached
    const instanceId = instance.instance_id
    console.log(`[端口映射缓存] 恢复实例 ${instanceId}: deviceIp=${deviceIp}`)

    // 恢复端口映射到全局 Map
    const portMap = new Map()
    for (const [priv, pub] of Object.entries(ports)) {
      portMap.set(Number(priv), Number(pub))
    }
    window.openCecsPortMap.set(deviceIp, portMap)

    // 注入设备到主机列表
    addedDeviceIps.add(deviceIp)
    deviceIpToInstanceMap.set(deviceIp, instance)
    portMapVersion.value++
    const deviceObj = {
      ip: deviceIp,
      type: 'android',
      id: instance.instance_id || deviceIp,
      name: 'opencecs',
      version: 'v3',
      isOnline: true,
      lastSeen: new Date(),
      group: '默认分组',
      source: 'opencecs'
    }
    if (typeof window.addDiscoveredDevice === 'function') {
      window.addDiscoveredDevice(deviceObj)
      saveAddedDeviceIps()
      console.log(`[端口映射缓存] ✅ 设备已恢复注入: ${deviceIp}`)
      // 异步同步本地 token 到设备（不阻塞后续流程）
      syncTokenToDevice(deviceIp)
    }

    // 异步获取设备真实 ID 和名称（不阻塞后续流程）
    ;(async () => {
      try {
        if (typeof window.goHttpRequest === 'function') {
          const result = await window.goHttpRequest({ Method: 'GET', URL: `http://${deviceIp}/info/device`, Body: '', Headers: {} })
          if (result && result.success && result.body) {
            const infoData = result.body
            const fetchedId = infoData?.deviceId || infoData?.data?.deviceId
            if (fetchedId) {
              const firstChar = fetchedId[0]?.toLowerCase()
              const nameMap = { r: 'r1_v3', c: 'c1_v3', q: 'q1_v3', p: 'p1_v3' }
              deviceObj.id = fetchedId
              deviceObj.name = nameMap[firstChar] || 'opencecs'
              if (typeof window.addDiscoveredDevice === 'function') {
                window.addDiscoveredDevice(deviceObj)
                console.log(`[端口映射缓存] 📝 设备信息已更新: ${deviceIp} → id=${fetchedId}, name=${deviceObj.name}`)
              }
            }
          }
        }
      } catch (e) {
        console.warn(`[端口映射缓存] 获取设备信息失败 (${deviceIp}): ${e.message}`)
      }
    })()
  }

  // 保存恢复后的端口映射
  if (cachedInstances.length > 0) {
    savePortMapToStorage()
  }

  // 如果所有实例都命中缓存，直接结束
  if (uncachedInstances.length === 0) {
    console.log(`[端口映射] 所有实例均命中缓存，setupAllPortMappings v${version} 完成`)
    return
  }

  if (isStale()) { console.log(`[端口映射] v${version} 已过期，终止`); return }

  // ===== 以下仅处理未命中缓存/缓存失效的实例（已执行完删除旧映射逻辑） =====

  if (isStale()) { console.log(`[端口映射] v${version} 已过期，终止`); return }

  // ===== 阶段2：并行创建 8000 端口映射 =====
  const deviceInfos = await Promise.all(instanceInfos.map(async ({ instance, instanceId, overview }) => {
    const publicIp = overview?.nat_public_ip
    if (!publicIp) {
      console.warn(`[端口映射] ⏭️ 跳过实例 ${instanceId}: 无公网IP (overview=${JSON.stringify(overview)})`)
      return null
    }
    try {
      const res = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ protocol: 'TCP', private_port: 8000, remark: 'API' })
      })
      const data = await res.json()
      if (data.code === 200 && data.data?.public_port) {
        const deviceIp = `${publicIp}:${data.data.public_port}`
        console.log(`[端口映射] 8000 映射成功 (${instanceId}): ${deviceIp}`)
        return { instance, instanceId, publicIp, deviceIp }
      } else {
        console.warn(`[端口映射] 8000端口映射失败 (${instanceId}):`, data.message)
      }
    } catch (e) {
      console.error(`[端口映射] 8000端口映射网络错误 (${instanceId})`, e)
    }
    return null
  }))

  if (isStale()) { console.log(`[端口映射] v${version} 已过期，终止`); return }

  const validDeviceInfos = deviceInfos.filter(Boolean)
  console.log(`[端口映射] 有效设备: ${validDeviceInfos.length} / ${deviceInfos.length}`)

  // ===== 阶段3：查询容器 → 创建映射 → 构建 portMap → 注入设备 =====
  // 等待 1 秒让新 8000 端口映射生效（从 2s 缩短到 1s）
  await new Promise(r => setTimeout(r, 1000))

  await Promise.all(validDeviceInfos.map(async ({ instance, instanceId, publicIp, deviceIp }) => {
    // 每个实例独立 try/catch，避免一个失败导致其他实例也无法处理
    try {
      if (isStale()) return

      // 辅助函数：带超时的 fetch（设备直连请求可能长时间无响应）
      const fetchWithTimeout = (url, options = {}, timeoutMs = 10000) => {
        const controller = new AbortController()
        const timer = setTimeout(() => controller.abort(), timeoutMs)
        return fetch(url, { ...options, signal: controller.signal })
          .finally(() => clearTimeout(timer))
      }

      // 3a. 先注入设备到主机列表（确保即使后续请求超时，设备也能显示）
      addedDeviceIps.add(deviceIp)
      deviceIpToInstanceMap.set(deviceIp, instance)
      portMapVersion.value++

      // 先用实例名作为默认设备名，后续查到真实信息再更新
      const deviceObj = {
        ip: deviceIp,
        type: 'android',
        id: instance.instance_id || deviceIp,
        name: 'opencecs',
        version: 'v3',
        isOnline: true,
        lastSeen: new Date(),
        group: '默认分组',
        source: 'opencecs'
      }
      if (typeof window.addDiscoveredDevice === 'function') {
        window.addDiscoveredDevice(deviceObj)
        // 立即保存已注入的设备 IP 到 localStorage，确保下次刷新能找到并清除
        saveAddedDeviceIps()
        console.log(`[端口映射] ✅ 设备已注入主机列表: ${deviceIp} (${deviceObj.name})`)
        // 异步同步本地 token 到设备（不阻塞后续流程）
        syncTokenToDevice(deviceIp)
      }

      // 3b. 先获取设备信息（通过 Go 后端 IPC，稳定可靠）
      // 等 3 秒让 NAT 端口映射完全生效
      await new Promise(r => setTimeout(r, 3000))

      // 3b-1. 获取设备真实 ID 和名称（优先级最高，不依赖容器查询）
      let deviceId = deviceIp
      let deviceName = 'opencecs'
      try {
        if (typeof window.goHttpRequest === 'function') {
          console.log(`[端口映射] 查询设备信息 (${deviceIp}) via goHttpRequest...`)
          const result = await window.goHttpRequest({ Method: 'GET', URL: `http://${deviceIp}/info/device`, Body: '', Headers: {} })
          console.log(`[端口映射] goHttpRequest 返回 (${deviceIp}):`, JSON.stringify(result).substring(0, 300))
          if (result && result.success && result.body) {
            const infoData = result.body
            const fetchedId = infoData?.deviceId || infoData?.data?.deviceId
            if (fetchedId) {
              deviceId = fetchedId
              const firstChar = fetchedId[0]?.toLowerCase()
              const nameMap = { r: 'r1_v3', c: 'c1_v3', q: 'q1_v3', p: 'p1_v3' }
              if (nameMap[firstChar]) deviceName = nameMap[firstChar]
              console.log(`[端口映射] 设备信息获取成功 (${deviceIp}): id=${fetchedId}, name=${deviceName}`)
            }
          }
        }
      } catch (e) {
        console.warn(`[端口映射] 获取设备信息失败 (${deviceIp}): ${e.message}`)
      }

      // 3c. 立即更新设备 ID/名称（不等容器查询）
      if (deviceId !== deviceIp || deviceName !== deviceObj.name) {
        deviceObj.id = deviceId
        deviceObj.name = deviceName
        if (typeof window.addDiscoveredDevice === 'function') {
          window.addDiscoveredDevice(deviceObj)
          console.log(`[端口映射] 📝 设备信息已更新: ${deviceIp} → id=${deviceId}, name=${deviceName}`)
        }
      }

      if (isStale()) return

      // 3b-2. 查询容器列表（带 15 秒超时，防止 Go IPC 无限挂起）
      const withTimeout = (promise, ms) => Promise.race([
        promise,
        new Promise((_, reject) => setTimeout(() => reject(new Error(`超时 ${ms}ms`)), ms))
      ])

      let containers = []
      for (let attempt = 1; attempt <= 3; attempt++) {
        try {
          console.log(`[端口映射] 查询容器 (${deviceIp}), 第${attempt}次 (Go IPC)...`)
          const result = await withTimeout(GetContainers(deviceIp, 'v3', ''), 15000)
          console.log(`[端口映射] GetContainers 返回 (${deviceIp}):`, JSON.stringify(result).substring(0, 200))
          if (result && result.code === 0) {
            containers = result.data?.list || result.list || []
            console.log(`[端口映射] 查询容器成功 (${deviceIp}), 第${attempt}次, 共 ${containers.length} 个`)
            break
          }
        } catch (e) {
          console.warn(`[端口映射] 查询容器失败 (${deviceIp}), 第${attempt}次: ${e.message}`)
          if (attempt < 3) await new Promise(r => setTimeout(r, 2000))
        }
      }

      // 3d. 收集所有 HostPort 并批量创建映射
      if (containers.length > 0) {
        const seenPorts = new Set()
        const rules = []
        for (const container of containers) {
          if (!container.portBindings) continue
          for (const [key, bindings] of Object.entries(container.portBindings)) {
            if (!bindings || bindings.length === 0) continue
            const hostPort = parseInt(bindings[0].HostPort)
            if (!hostPort || isNaN(hostPort)) continue
            if (seenPorts.has(hostPort)) continue
            seenPorts.add(hostPort)
            const protocol = key.split('/')[1]?.toUpperCase() || 'TCP'
            rules.push({ protocol, private_port: hostPort, remark: `${key}_${hostPort}` })
          }
        }
        console.log(`[端口映射] 容器端口规则 (${instanceId}): ${rules.length} 条`)

        if (rules.length > 0) {
          try {
            const batchRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings/batch`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
              body: JSON.stringify({ rules })
            })
            const batchData = await batchRes.json()
            console.log(`[端口映射] 批量创建映射结果 (${instanceId}):`, batchData)

            // 验证：重新查询已有映射，逐个补建缺失的端口
            await verifyAndRetryMissingPorts(instanceId, rules, token)
          } catch (e) {
            console.warn(`[端口映射] 批量创建映射失败 (${instanceId})`, e)
          }
        }
      } else {
        console.warn(`[端口映射] 未查询到容器 (${deviceIp})`)
      }

      if (isStale()) return

      // 3e. 查询最终完整映射，构建 portMap，并缓存到 localStorage
      try {
        const listRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings?page=1&page_size=200`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
        const listData = await listRes.json()
        const allMappings = listData.data?.list || listData.data || []
        if (allMappings.length > 0) {
          const portMap = new Map()
          for (const m of allMappings) {
            if (m.private_port && m.public_port) {
              portMap.set(Number(m.private_port), Number(m.public_port))
            }
          }
          window.openCecsPortMap.set(deviceIp, portMap)
          savePortMapToStorage()
          // 保存实例到设备映射缓存
          saveInstanceDeviceMap(instanceId, publicIp, deviceIp)
          portMapVersion.value++
          console.log(`[端口映射] 端口映射表已构建并缓存 (${deviceIp}):`, Object.fromEntries(portMap))
        }
      } catch (e) {
        console.warn(`[端口映射] 查询端口映射列表失败 (${instanceId})`, e)
      }
    } catch (e) {
      console.error(`[端口映射] ❌ 实例处理异常 (${instanceId}, ${deviceIp}):`, e)
    }
  }))

  console.log(`[端口映射] setupAllPortMappings v${version} 完成`)
}

/**
 * 验证批量创建后是否所有端口都已映射，若有遗漏则逐个补建。
 * 批量 API 可能因服务端限制静默丢弃部分规则，此函数确保 100% 覆盖。
 * @param {string} instanceId - 实例 ID
 * @param {Array} expectedRules - 期望已创建的规则列表 [{protocol, private_port, remark}]
 * @param {string} token - 鉴权 token
 */
const verifyAndRetryMissingPorts = async (instanceId, expectedRules, token) => {
  try {
    // 重新查询已有映射
    const checkRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings?page=1&page_size=99999`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    const checkData = await checkRes.json()
    const currentMappings = (checkData.code === 200 ? checkData.data?.list : null) || []
    const mappedPorts = new Set(currentMappings.map(m => Number(m.private_port)).filter(Boolean))

    // 找出仍未映射的端口
    const missingRules = expectedRules.filter(r => !mappedPorts.has(r.private_port))
    if (missingRules.length === 0) {
      console.log(`[端口映射] ✅ 验证通过，所有 ${expectedRules.length} 个端口已映射 (${instanceId})`)
      return
    }

    console.warn(`[端口映射] ⚠️ 批量创建后仍有 ${missingRules.length} 个端口未映射 (${instanceId})，逐个补建:`, 
      missingRules.map(r => r.private_port))

    // 逐个创建缺失的端口映射
    let successCount = 0
    for (const rule of missingRules) {
      try {
        const singleRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
          body: JSON.stringify({ protocol: rule.protocol, private_port: rule.private_port, remark: rule.remark })
        })
        const singleData = await singleRes.json()
        if (singleData.code === 200) {
          successCount++
          console.log(`[端口映射] 补建成功: ${rule.private_port} (${rule.remark})`)
        } else {
          console.warn(`[端口映射] 补建失败: ${rule.private_port} (${rule.remark}):`, singleData.message)
        }
      } catch (e) {
        console.warn(`[端口映射] 补建异常: ${rule.private_port} (${rule.remark}):`, e.message)
      }
    }

    console.log(`[端口映射] 补建完成 (${instanceId}): ${successCount}/${missingRules.length} 成功`)
  } catch (e) {
    console.error(`[端口映射] 验证补建过程异常 (${instanceId}):`, e)
  }
}

/**
 * 容器创建成功后调用：检查该实例是否已有容器端口映射规则，
 * 如果没有则查询容器 portBindings 并批量创建映射。
 * @param {string} instanceId - OpenCecs 实例 ID
 * @param {string} deviceIp - 设备的公网 IP:Port（通过 8000 端口映射获得）
 */
const ensureContainerPortMappings = async (instanceId, deviceIp) => {
  const token = localStorage.getItem('opencecs_token')
  if (!token || !instanceId || !deviceIp) return

  try {
    // 等待 3 秒让新创建的容器完成注册，端口绑定信息可查
    console.log(`[端口映射] 等待 3s 让新容器注册 (${instanceId})...`)
    await new Promise(r => setTimeout(r, 3000))

    // 1. 通过 Go IPC 查询容器 portBindings（比直接 HTTP 更可靠，不受 NAT 影响）
    const withTimeout = (promise, ms) => Promise.race([
      promise,
      new Promise((_, reject) => setTimeout(() => reject(new Error(`超时 ${ms}ms`)), ms))
    ])

    let containers = []
    for (let attempt = 1; attempt <= 3; attempt++) {
      try {
        console.log(`[端口映射] 查询容器 (${deviceIp}), 第${attempt}次 (Go IPC)...`)
        const result = await withTimeout(GetContainers(deviceIp, 'v3', ''), 15000)
        console.log(`[端口映射] GetContainers 返回 (${deviceIp}):`, JSON.stringify(result).substring(0, 200))
        if (result && result.code === 0) {
          containers = result.data?.list || result.list || []
          console.log(`[端口映射] 查询容器成功 (${deviceIp}), 第${attempt}次, 共 ${containers.length} 个`)
          break
        }
      } catch (e) {
        console.warn(`[端口映射] 查询容器失败 (${deviceIp}), 第${attempt}次: ${e.message}`)
        if (attempt < 3) await new Promise(r => setTimeout(r, 2000))
      }
    }

    if (containers.length === 0) {
      console.warn(`[端口映射] 未查询到容器 (${deviceIp}), 跳过端口映射`)
      return
    }

    // 2. 收集所有容器的 HostPort
    const allHostPorts = new Set()
    const rules = []
    for (const container of containers) {
      if (!container.portBindings) continue
      for (const [key, bindings] of Object.entries(container.portBindings)) {
        if (!bindings || bindings.length === 0) continue
        const hostPort = parseInt(bindings[0].HostPort)
        if (!hostPort || isNaN(hostPort)) continue
        if (allHostPorts.has(hostPort)) continue
        allHostPorts.add(hostPort)
        const protocol = key.split('/')[1]?.toUpperCase() || 'TCP'
        rules.push({ protocol, private_port: hostPort, remark: `${key}_${hostPort}` })
      }
    }

    if (allHostPorts.size === 0) {
      console.warn(`[端口映射] 容器无端口绑定 (${deviceIp}), 跳过`)
      return
    }

    console.log(`[端口映射] 收集到 ${rules.length} 个容器端口规则 (${instanceId})`)

    // 3. 查询已有映射，找出缺少的端口
    const listRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings?page=1&page_size=99999`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    const listData = await listRes.json()
    const existingMappings = (listData.code === 200 ? listData.data?.list : null) || []
    const existingPrivPorts = new Set(existingMappings.map(m => Number(m.private_port)).filter(Boolean))

    // 4. 只为未映射的端口创建规则
    console.log(`[端口映射] 已有映射端口 (${instanceId}):`, [...existingPrivPorts])
    const newRules = rules.filter(r => !existingPrivPorts.has(r.private_port))
    console.log(`[端口映射] 需要新建的规则 (${instanceId}):`, JSON.stringify(newRules))
    if (newRules.length > 0) {
      const batchRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings/batch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ rules: newRules })
      })
      const batchData = await batchRes.json()
      console.log(`[端口映射] 容器创建后批量映射完成 (${instanceId}), 新建 ${newRules.length} 条, 结果:`, batchData)

      // 验证：重新查询已有映射，逐个补建缺失的端口
      await verifyAndRetryMissingPorts(instanceId, newRules, token)
    } else {
      console.log(`[端口映射] 容器创建后所有端口已映射 (${instanceId}), 跳过创建`)
    }

    // 5. 始终重新查询完整映射并更新 openCecsPortMap
    const finalRes = await fetch(`${BASE_URL}/cecs/instances/${instanceId}/port-mappings?page=1&page_size=200`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    const finalData = await finalRes.json()
    const allMappings = finalData.data?.list || finalData.data || []
    if (allMappings.length > 0 && deviceIp) {
      const portMap = new Map()
      for (const m of allMappings) {
        if (m.private_port && m.public_port) {
          portMap.set(Number(m.private_port), Number(m.public_port))
        }
      }
      // 同时以 deviceIp 和 IP 前缀的形式存入，确保 extractPort 模糊匹配能命中
      window.openCecsPortMap?.set(deviceIp, portMap)
      savePortMapToStorage()
      // 同步保存实例到设备映射缓存，下次刷新时可直接恢复
      const publicIp = deviceIp.includes(':') ? deviceIp.split(':')[0] : deviceIp
      saveInstanceDeviceMap(instanceId, publicIp, deviceIp)
      portMapVersion.value++
      console.log(`[端口映射] 容器创建后端口映射表已更新并缓存 (${deviceIp}):`, Object.fromEntries(portMap))
    }
  } catch (e) {
    console.error(`[端口映射] ensureContainerPortMappings 失败 (${instanceId})`, e)
  }
}

defineExpose({ init, fetchInstances, ensureContainerPortMappings, getInstanceByDeviceIp })
</script>

<style scoped>
/* ===== 整体容器 ===== */
.opencecs-container {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 400px;
  background: #0d0d1a;
  color: #e0e0e0;
  font-size: 14px;
  box-sizing: border-box;
}

/* ===== 顶部栏 ===== */
.top-bar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: space-between;
  padding: 14px 24px;
  background: #13132a;
  border-bottom: 1px solid #252545;
}
.top-bar-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
}
.top-right-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}
.btn-login {
  padding: 7px 20px;
  background: #2dd4b8;
  color: #000;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}
.btn-login:hover { opacity: 0.85; }
.user-info { display: flex; align-items: center; gap: 8px; }
.user-avatar { width: 28px; height: 28px; border-radius: 50%; border: 1px solid #2dd4b8; object-fit: cover; }
.user-name { color: #fff; font-size: 14px; font-weight: 500; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { padding: 6px 14px; background: transparent; color: #f56c6c; border: 1px solid #f56c6c; border-radius: 6px; font-size: 13px; cursor: pointer; transition: background 0.2s; }
.btn-logout:hover { background: rgba(245,108,108,0.1); }

/* ===== 统计栏 ===== */
.instance-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px 24px;
  overflow: hidden;
}
.stats-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 16px;
  background: #13132a;
  border: 1px solid #252545;
  border-radius: 10px;
  padding: 12px 20px;
}
.stat-item { display: flex; align-items: center; gap: 6px; }
.stat-label { color: #8899bb; font-size: 13px; }
.stat-value { font-weight: 700; color: #fff; font-size: 15px; }
.stat-dot { width: 8px; height: 8px; border-radius: 50%; }
.stat-item.running .stat-dot { background: #2dd4b8; }
.stat-item.running .stat-value { color: #2dd4b8; }
.stat-item.stopped .stat-dot { background: #909399; }
.stat-item.stopped .stat-value { color: #909399; }
.stat-item.expired .stat-dot { background: #f56c6c; }
.stat-item.expired .stat-value { color: #f56c6c; }
.btn-refresh { margin-left: auto; padding: 6px 16px; background: transparent; border: 1px solid #2dd4b8; color: #2dd4b8; border-radius: 6px; font-size: 13px; cursor: pointer; transition: background 0.2s; }
.btn-refresh:hover:not(:disabled) { background: #2dd4b818; }
.btn-refresh:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== 表格 ===== */
.table-wrap {
  flex: 1;
  overflow: auto;
  border-radius: 10px;
  border: 1px solid #252545;
}
.table-wrap::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
.table-wrap::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}
.table-wrap::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}
.table-wrap::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

.instance-table { width: 100%; border-collapse: collapse; }
.instance-table th {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #13132a;
  color: #8899bb;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  text-align: center;
  border-bottom: 1px solid #252545;
  white-space: nowrap;
}
.instance-table td {
  padding: 11px 14px;
  border-bottom: 1px solid #1e1e38;
  color: #ccd;
  font-size: 13px;
  white-space: nowrap;
  text-align: center;
}
.instance-table tr:last-child td { border-bottom: none; }
.instance-table tr:hover td { background: #13132a; }
.monospace { font-family: monospace; font-size: 12px; color: #aab; }
.expire-soon { color: #e6a23c !important; font-weight: 600; }

/* 状态标签 */
.status-tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}
/* .status-tag.running { background: rgba(45,212,184,0.15); color: #2dd4b8; } */
/* .status-tag.stopped { background: rgba(144,147,153,0.15); color: #909399; } */
/* .status-tag.expired { background: rgba(245,108,108,0.15); color: #f56c6c; } */
/* .status-tag.creating { background: rgba(64,158,255,0.15); color: #409eff; } */

/* 操作按钮 */
.btn-action {
  padding: 4px 12px;
  background: transparent;
  border: 1px solid #2dd4b8;
  color: #2dd4b8;
  border-radius: 5px;
  font-size: 12px;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-action:hover { background: #2dd4b818; }

/* 空/加载 */
.list-loading, .list-empty {
  text-align: center;
  color: #8899bb;
  padding: 40px 0;
  font-size: 14px;
}
.not-login-tip {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  gap: 16px;
  color: #8899bb;
}

/* ===== 遮罩 ===== */
.dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

/* ===== 登录弹窗 ===== */
.dialog-box {
  position: relative;
  width: 460px;
  background: #1a1a2e;
  border-radius: 16px;
  padding: 40px 40px 32px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.6);
  max-height: 90vh;
  overflow-y: auto;
}
.port-dialog-box { width: 540px; }
.dialog-close { position: absolute; top: 16px; right: 20px; background: transparent; border: none; color: #aaa; font-size: 18px; cursor: pointer; }
.dialog-close:hover { color: #fff; }
.dialog-title { text-align: center; color: #fff; font-size: 26px; font-weight: 700; margin: 0 0 8px; }
.dialog-subtitle { text-align: center; color: #8899bb; font-size: 13px; margin: 0 0 28px; }

/* Tab */
.login-tabs { display: flex; background: #252540; border-radius: 8px; padding: 4px; margin-bottom: 24px; }
.login-tab { flex: 1; text-align: center; padding: 8px 0; border-radius: 6px; color: #8899bb; font-size: 14px; cursor: pointer; transition: background 0.2s, color 0.2s; user-select: none; }
.login-tab.active { background: #353560; color: #fff; font-weight: 600; }

/* 表单 */
.login-form { display: flex; flex-direction: column; gap: 16px; margin-bottom: 16px; }
.form-item { display: flex; flex-direction: column; gap: 6px; }
.form-item label { color: #ccd; font-size: 13px; font-weight: 500; }
.required { color: #f56c6c; margin-left: 2px; }
.form-input {
  width: 100%; padding: 10px 14px; background: #252540; border: 1px solid #353560;
  border-radius: 8px; color: #fff; font-size: 14px; outline: none; box-sizing: border-box; transition: border-color 0.2s;
}
.form-input:focus { border-color: #2dd4b8; }
.form-input::placeholder { color: #556; }
.form-select { appearance: none; cursor: pointer; }

/* 验证码行 */
.input-code-wrap { display: flex; gap: 10px; align-items: center; }
.input-code-wrap .form-input { flex: 1; }
.btn-get-code { flex-shrink: 0; padding: 10px 14px; background: #252540; border: 1px solid #2dd4b8; border-radius: 8px; color: #2dd4b8; font-size: 13px; font-weight: 600; cursor: pointer; white-space: nowrap; transition: background 0.2s; }
.btn-get-code:hover:not(:disabled) { background: #2dd4b818; }
.btn-get-code:disabled { border-color: #444; color: #666; cursor: not-allowed; }

/* 密码行 */
.input-password-wrap { position: relative; }
.input-password-wrap .form-input { padding-right: 40px; }
.eye-icon { position: absolute; right: 12px; top: 50%; transform: translateY(-50%); cursor: pointer; font-size: 16px; color: #8899bb; user-select: none; }

/* 记住我 */
.login-options { display: flex; align-items: center; margin-bottom: 14px; }
.remember-me { display: flex; align-items: center; gap: 6px; color: #aab; font-size: 13px; cursor: pointer; }
.remember-me input[type="checkbox"] { accent-color: #2dd4b8; width: 15px; height: 15px; cursor: pointer; }

/* 错误提示 */
.login-error { color: #f56c6c; font-size: 13px; margin-bottom: 10px; padding: 8px 12px; background: rgba(245,108,108,0.1); border-radius: 6px; border: 1px solid rgba(245,108,108,0.3); }

/* 提交按钮 */
.btn-submit { width: 100%; padding: 12px 0; background: #2dd4b8; color: #000; border: none; border-radius: 8px; font-size: 16px; font-weight: 700; cursor: pointer; transition: opacity 0.2s; margin-bottom: 4px; }
.btn-submit:hover:not(:disabled) { opacity: 0.88; }
.btn-submit:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== 端口映射新弹窗 ===== */
.port-add-dialog {
  width: 420px;
  background: #1c1c30;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.7);
}
.pad-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 20px;
  background: #1e3a5f;
  border-bottom: 1px solid #25406a;
}
.pad-header-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid #4a9eff;
  color: #4a9eff;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}
.pad-header-title {
  flex: 1;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
}
.pad-close {
  background: transparent;
  border: none;
  color: #8899bb;
  font-size: 16px;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}
.pad-close:hover { color: #fff; }
.pad-body { padding: 20px 20px 10px; }
.pad-field { margin-bottom: 16px; }
.pad-label { display: block; color: #b0bdd4; font-size: 13px; margin-bottom: 8px; }
/* 协议单选 */
.pad-radio-group { display: flex; gap: 20px; }
.pad-radio {
  display: flex;
  align-items: center;
  gap: 7px;
  color: #ccd;
  font-size: 13px;
  cursor: pointer;
  user-select: none;
}
.pad-radio input[type="radio"] { display: none; }
.pad-radio-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid #556;
  position: relative;
  flex-shrink: 0;
  transition: border-color 0.2s;
}
.pad-radio.active .pad-radio-dot {
  border-color: #2dd4b8;
}
.pad-radio.active .pad-radio-dot::after {
  content: '';
  position: absolute;
  inset: 3px;
  border-radius: 50%;
  background: #2dd4b8;
}
/* 内网端口输入 */
.pad-port-input-wrap {
  display: flex;
  align-items: center;
  background: #252540;
  border: 1px solid #353560;
  border-radius: 8px;
  overflow: hidden;
}
.pad-port-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #fff;
  font-size: 14px;
  padding: 10px 14px;
  min-width: 0;
}
.pad-port-input::placeholder { color: #556; }
.pad-port-input::-webkit-outer-spin-button,
.pad-port-input::-webkit-inner-spin-button { -webkit-appearance: none; }
.pad-stepper {
  width: 36px;
  height: 40px;
  background: #2a2a50;
  border: none;
  border-left: 1px solid #353560;
  color: #8899bb;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  flex-shrink: 0;
}
.pad-stepper:hover { background: #353570; color: #2dd4b8; }
/* 备注输入 */
.pad-remark-wrap { position: relative; }
.pad-remark-input {
  width: 100%;
  padding: 10px 50px 10px 14px;
  background: #252540;
  border: 1px solid #353560;
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s;
}
.pad-remark-input:focus { border-color: #2dd4b8; }
.pad-remark-input::placeholder { color: #556; }
.pad-remark-count {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 11px;
  color: #556;
  pointer-events: none;
  white-space: nowrap;
}
/* 提示框 */
.pad-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  margin-bottom: 10px;
}
.pad-tip.info { background: #1a3050; color: #7bbfff; border: 1px solid #254870; }
.pad-tip.warn { background: #2e2008; color: #e6a23c; border: 1px solid #5a400a; }
.pad-tip-icon { flex-shrink: 0; font-size: 14px; }
/* 底部按钮 */
.pad-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 20px;
  border-top: 1px solid #252545;
}
.pad-btn-cancel {
  padding: 8px 24px;
  background: transparent;
  border: 1px solid #353560;
  border-radius: 8px;
  color: #aab;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}
.pad-btn-cancel:hover { background: #252540; }
.pad-btn-add {
  padding: 8px 28px;
  background: #2dd4b8;
  border: none;
  border-radius: 8px;
  color: #000;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: opacity 0.2s;
}
.pad-btn-add:hover:not(:disabled) { opacity: 0.88; }
.pad-btn-add:disabled { opacity: 0.5; cursor: not-allowed; }

/* ===== 端口映射概览信息 ===== */
.pad-overview {
  margin: 0 20px 0;
  padding: 14px 16px;
  background: #14142a;
  border: 1px solid #252545;
  border-radius: 8px;
  margin-top: 14px;
}
.pad-overview-loading, .pad-overview-err {
  font-size: 12px;
  color: #8899bb;
  text-align: center;
  padding: 4px 0;
}
.pad-overview-err { color: #f56c6c; }
.pad-overview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 20px;
}
.pad-ov-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.pad-ov-label {
  font-size: 11px;
  color: #6677aa;
  white-space: nowrap;
}
.pad-ov-value {
  font-size: 13px;
  color: #ccd;
  word-break: break-all;
}
.pad-ov-value.monospace { font-family: monospace; color: #2dd4b8; }
.pad-ov-used { color: #2dd4b8; font-weight: 600; }
.pad-ov-sep { color: #556; margin: 0 3px; }
.pad-ov-quota { color: #8899bb; }

/* ===== 初始化进度条 ===== */
.pad-init-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 10px 20px 0;
  padding: 9px 14px;
  background: #1a2a40;
  border: 1px solid #254870;
  border-radius: 8px;
  color: #7bbfff;
  font-size: 12px;
}
.pad-init-spinner {
  display: inline-block;
  width: 13px;
  height: 13px;
  border: 2px solid #4a9eff44;
  border-top-color: #4a9eff;
  border-radius: 50%;
  animation: pad-spin 0.7s linear infinite;
  flex-shrink: 0;
}
@keyframes pad-spin { to { transform: rotate(360deg); } }

/* ===== 添加成功结果 ===== */
.action-cell {
  display: flex;
  gap: 6px;
  align-items: center;
}
.btn-action {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  border: none;
  color: #fff;
  transition: opacity 0.2s;
}
.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  filter: grayscale(1);
}
.btn-action:hover:not(:disabled) {
  opacity: 0.8;
}
.btn-action.start { background: #67c23a; }
.btn-action.stop { background: #f56c6c; }
.btn-action.restart { background: #e6a23c; }

.pad-add-result {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 10px;
  padding: 10px 14px;
  background: rgba(45,212,184,0.08);
  border: 1px solid rgba(45,212,184,0.3);
  border-radius: 8px;
  color: #2dd4b8;
  font-size: 13px;
  line-height: 1.5;
}
.pad-add-result-icon {
  flex-shrink: 0;
  font-weight: 700;
  font-size: 15px;
}
.pad-add-result strong { color: #fff; }
</style>
