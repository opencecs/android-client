<template>
  <div class="interconnected-cloud-machines">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-checkbox
          v-model="selectAll"
          :indeterminate="isIndeterminate"
          @change="handleSelectAll"
          style="margin-right: 8px;"
        />
        <span class="selected-info" v-if="selectedIds.size > 0">{{ $t('common.selected', { count: selectedIds.size }) }}</span>
        <span class="selected-info" v-else>{{ $t('common.totalMachines', { count: machines.length }) }}</span>
      </div>
      <div class="toolbar-right">
        <template v-if="selectedIds.size > 0">
          <el-button type="danger" size="small" :icon="Delete" @click="handleBatchDelete">{{ $t('common.batchDeleteBtn') }}</el-button>
          <el-divider direction="vertical" />
        </template>
        <el-button
          v-if="isBatchControlling || selectedIds.size >= 2"
          :type="isBatchControlling ? 'warning' : 'success'"
          size="small"
          :icon="isBatchControlling ? VideoPause : VideoPlay"
          @click="handleBatchControl"
        >{{ isBatchControlling ? $t('common.stopControlBtn') : $t('common.batchControlBtn') }}</el-button>
        <el-button type="primary" size="small" :icon="Plus" @click="openAddDialog">{{ $t('common.addCloudMachine') }}</el-button>
      </div>
    </div>

    <!-- 云机列表 -->
    <div class="machine-grid" v-if="machines.length > 0">
      <div
        v-for="(machine, index) in machines"
        :key="machine.name"
        class="machine-card"
        :class="{ selected: selectedIds.has(machine.name) }"
        @click="toggleSelect(machine.name)"
      >
        <!-- 顶部名称栏：白底，选择框 + 名称 + 删除 -->
        <div class="card-header">
          <el-checkbox
            :model-value="selectedIds.has(machine.name)"
            @change="toggleSelect(machine.name)"
            @click.stop
          />
          <span class="card-name" :title="machine.name">{{ machine.name }}</span>
          <div class="card-ops" @click.stop>
            <el-tooltip :content="$t('common.delete')" placement="top">
              <el-button type="danger" size="small" :icon="Delete" circle text @click="handleDeleteOne(machine)" />
            </el-tooltip>
          </div>
        </div>

        <!-- 截图区域（点击投屏） -->
        <div class="screenshot-area" @click.stop="handleProjection(machine)">
          <img
            v-if="getScreenshot(machine)"
            :src="getScreenshot(machine)"
            class="screenshot-img"
          />
          <div v-else class="screenshot-placeholder">
            <el-icon class="placeholder-icon"><Monitor /></el-icon>
            <span>{{ $t('common.clickToProject') }}</span>
          </div>
          <!-- 投屏悬停蒙层 -->
          <!-- <div class="projection-hint">
            <el-icon><VideoPlay /></el-icon>
            <span>投屏</span>
          </div> -->
          <!-- 右下角序号角标 -->
          <span class="card-index">{{ index + 1 }}</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <el-empty
      v-else
      :description="$t('common.noMachineHint')"
      style="margin-top: 60px;"
    >
      <el-button type="primary" :icon="Plus" @click="openAddDialog">{{ $t('common.addCloudMachine') }}</el-button>
    </el-empty>

    <!-- 添加云机弹窗 -->
    <el-dialog
      v-model="addDialogVisible"
      :title="$t('common.addMachineDialog')"
      width="560px"
      destroy-on-close
      @close="resetAddForm"
    >
      <div class="add-dialog-body">
        <el-input
          v-model="batchKeyInput"
          type="textarea"
          :rows="6"
          :placeholder="$t('common.enterKeyHint')"
        />
        <el-alert
          v-if="parseError"
          :title="parseError"
          type="error"
          :closable="true"
          @close="parseError = ''"
          style="margin-top: 8px;"
        />
      </div>
      <template #footer>
        <el-button @click="addDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :disabled="!batchKeyInput.trim()" @click="confirmAdd">{{ $t('common.confirmAddBtn') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Monitor, VideoPlay, VideoPause } from '@element-plus/icons-vue'
import { startProjection, startProjectionBatchControl, stopProjectionBatchControl } from '../services/api.js'
import { GetScreenshotProxy } from '../../bindings/edgeclient/app.js'

const props = defineProps({
  token: { type: String, default: '' }
})

// ==================== 状态 ====================
const STORAGE_KEY = 'interconnected_machines'

const machines = ref([])
const selectedIds = ref(new Set())
const addDialogVisible = ref(false)
const batchKeyInput = ref('')
const parseError = ref('')

// 批量控制状态
const isBatchControlling = ref(false)

// 本组件私有截图缓存：Map<"ip_name", base64DataURL>
const localScreenshotCache = ref(new Map())
let screenshotTimer = null
let fetching = false

// ==================== 截图 ====================
/**
 * 直接通过 machine.i（API端口）HTTP 请求截图
 * 接口：http://<ip>:<i>/android/screenshot?name=<name>
 * 返回 base64 图片数据
 */
const getScreenshot = (machine) => {
  const key = `${machine.ip}_${machine.name}`
  return localScreenshotCache.value.get(key) || ''
}

const fetchScreenshots = async () => {
  if (fetching || machines.value.length === 0) return
  fetching = true
  try {
    await Promise.all(machines.value.map(async (machine) => {
      if (!machine.ip || !machine.i) return
      try {
        // 截图接口格式：http://<ip>:<i>/task=snap&level=1
        // 通过 Go 后端代理请求，避免 CORS 问题
        const url = `http://${machine.ip}:${machine.i}/task=snap&level=1`
        const result = await GetScreenshotProxy(url)
        if (result && result.success && result.data) {
          const key = `${machine.ip}_${machine.name}`
          localScreenshotCache.value.set(key, result.data)
          localScreenshotCache.value = new Map(localScreenshotCache.value)
        }
      } catch {
        // 忽略单台失败，不影响其他
      }
    }))
  } finally {
    fetching = false
  }
}

const startScreenshotRefresh = () => {
  stopScreenshotRefresh()
  fetchScreenshots()
  screenshotTimer = setInterval(fetchScreenshots, 500)
}

const stopScreenshotRefresh = () => {
  if (screenshotTimer) {
    clearInterval(screenshotTimer)
    screenshotTimer = null
  }
  fetching = false
}

// ==================== 选择逻辑 ====================
const selectAll = computed(() => machines.value.length > 0 && selectedIds.value.size === machines.value.length)
const isIndeterminate = computed(() => selectedIds.value.size > 0 && selectedIds.value.size < machines.value.length)

const handleSelectAll = (val) => {
  selectedIds.value = val ? new Set(machines.value.map(m => m.name)) : new Set()
}

const toggleSelect = (name) => {
  const next = new Set(selectedIds.value)
  next.has(name) ? next.delete(name) : next.add(name)
  selectedIds.value = next
}

// ==================== 本地存储 ====================
const loadFromStorage = () => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    machines.value = raw ? JSON.parse(raw) : []
  } catch {
    machines.value = []
  }
  startScreenshotRefresh()
}

const saveToStorage = () => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(machines.value))
}

// ==================== Base64 密钥解析 ====================
const parseOneKey = (raw) => {
  const key = raw.trim()
  if (!key) return null
  try {
    let decoded = ''
    try { decoded = atob(key) } catch { decoded = key }
    const params = {}
    decoded.split('&').forEach(seg => {
      const eqIdx = seg.indexOf('=')
      if (eqIdx === -1) return
      const k = seg.slice(0, eqIdx).trim().toLowerCase()
      params[k] = seg.slice(eqIdx + 1).trim()
    })
    if (!params.n || !params.ip) return { _error: '缺少 n 或 ip', _raw: key }
    return {
      name: params.n, ip: params.ip,
      t: params.t || '', u: params.u || '', i: params.i || '',
      c: params.c || '', ct: params.ct || '', cu: params.cu || ''
    }
  } catch (e) {
    return { _error: '解析失败：' + e.message, _raw: key }
  }
}

const parseKeys = () => {
  parseError.value = ''
  const lines = batchKeyInput.value.split(/[\n,]/).map(l => l.trim()).filter(Boolean)
  if (!lines.length) { parseError.value = '请输入密钥'; return [] }
  return lines.map(l => parseOneKey(l)).filter(Boolean)
}

// ==================== 添加云机 ====================
const openAddDialog = () => {
  batchKeyInput.value = ''
  parseError.value = ''
  addDialogVisible.value = true
}

const resetAddForm = () => {
  batchKeyInput.value = ''
  parseError.value = ''
}

const confirmAdd = () => {
  parseError.value = ''
  const parsed = parseKeys()
  if (!parsed.length) { if (!parseError.value) parseError.value = '未能解析出任何云机信息'; return }

  const valid = parsed.filter(m => !m._error)
  const invalid = parsed.filter(m => m._error)
  if (!valid.length) { parseError.value = `所有密钥解析失败（${invalid.length} 条），请检查格式`; return }

  let addCount = 0, skipCount = 0
  const existingNames = new Set(machines.value.map(m => m.name))
  for (const m of valid) {
    if (existingNames.has(m.name)) { skipCount++; continue }
    machines.value.push({ ...m })
    existingNames.add(m.name)
    addCount++
  }

  saveToStorage()
  startScreenshotRefresh()
  addDialogVisible.value = false

  const parts = []
  if (addCount > 0) parts.push(`成功添加 ${addCount} 台`)
  if (skipCount > 0) parts.push(`${skipCount} 台名称重复已跳过`)
  if (invalid.length > 0) parts.push(`${invalid.length} 条密钥解析失败`)
  addCount > 0 ? ElMessage.success(parts.join('，')) : ElMessage.warning(parts.join('，'))
}

// ==================== 投屏 ====================
const handleProjection = async (machine) => {
  try {
    const device = { ip: machine.ip }
    // 用密钥字段 t(TCP视频/10000) u(UDP控制/10001) 构造 portBindings
    const portBindings = {}
    if (machine.t) {
      portBindings[`10000/tcp`] = [{ HostPort: String(machine.t) }]
    }
    if (machine.u) {
      portBindings[`10001/tcp`] = [{ HostPort: String(machine.u) }]
      portBindings[`10001/udp`] = [{ HostPort: String(machine.u) }]
    }
    const containerInfo = {
      name: machine.name,
      ID: machine.name,
      portBindings,
      width: 360,
      height: 640
    }
    await startProjection(device, containerInfo, 0)
  } catch (e) {
    ElMessage.error('投屏失败：' + (e.message || e))
  }
}

// ==================== 批量控制 ====================
// 当前操作目标：有选中则操作选中，否则操作全部
const targetMachines = computed(() => {
  if (selectedIds.value.size > 0) {
    return machines.value.filter(m => selectedIds.value.has(m.name))
  }
  return machines.value
})

const handleBatchControl = async () => {
  // 停止
  if (isBatchControlling.value) {
    try {
      await stopProjectionBatchControl()
      isBatchControlling.value = false
    } catch (e) {
      ElMessage.error('停止批量控制失败：' + (e.message || e))
    }
    return
  }

  // 启动
  const targets = targetMachines.value.filter(m => m.t && m.u)
  if (!targets.length) {
    ElMessage.warning('没有可用的云机（需要有效的 t/u 端口）')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定对 ${targets.length} 台云机进行批量控制吗？`,
      '批量控制',
      { type: 'info', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch { return }

  try {
    // 构造兼容 startProjectionBatchControl 的容器数组
    // 每台机器的 t=TCP视频(10000映射), u=UDP控制(10001映射)
    const containers = targets.map(m => ({
      status: 'running',
      deviceIp: m.ip,
      name: m.name,
      ID: m.name,
      width: 360,
      height: 640,
      portBindings: {
        '10000/tcp': [{ HostPort: String(m.t) }],
        '10001/tcp': [{ HostPort: String(m.u) }],
        '10001/udp': [{ HostPort: String(m.u) }],
      }
    }))
    const firstTarget = targets[0]
    const device = { ip: firstTarget.ip }
    await startProjectionBatchControl(device, containers, '互联云机批量控制', 0)
    isBatchControlling.value = true
  } catch (e) {
    ElMessage.error('批量控制失败：' + (e.message || e))
  }
}

// ==================== 删除 ====================
const handleDeleteOne = async (machine) => {
  try {
    await ElMessageBox.confirm(`确定删除云机「${machine.name}」吗？`, '提示', {
      type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消'
    })
    machines.value = machines.value.filter(m => m.name !== machine.name)
    selectedIds.value.delete(machine.name)
    saveToStorage()
    ElMessage.success('已删除')
  } catch { /* 取消 */ }
}

const handleBatchDelete = async () => {
  if (!selectedIds.value.size) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedIds.value.size} 台云机吗？此操作不可恢复。`,
      '批量删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
    machines.value = machines.value.filter(m => !selectedIds.value.has(m.name))
    selectedIds.value = new Set()
    saveToStorage()
    ElMessage.success('批量删除成功')
  } catch { /* 取消 */ }
}

// ==================== 生命周期 ====================
onMounted(() => { loadFromStorage() })
onUnmounted(() => { stopScreenshotRefresh() })
</script>

<style scoped>
.interconnected-cloud-machines {
  height: 100%;
  padding: 12px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow: hidden;
}

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  flex-shrink: 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.selected-info {
  font-size: 13px;
  color: #606266;
  margin-top: 0px;
}

/* 云机网格 */
.machine-grid {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 10px;
  align-content: start;
}

/* 云机卡片 */
.machine-card {
  position: relative;
  background: #fff;
  border: 2px solid #ebeef5;
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  user-select: none;
  display: flex;
  flex-direction: column;
}

.machine-card:hover {
  border-color: #c6e0ff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.12);
}

.machine-card.selected {
  border-color: #409eff;
  box-shadow: 0 0 0 1px #409eff;
}

/* 顶部名称栏：白底，与截图明确分隔 */
.card-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 6px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
  min-width: 0;
}

.card-name {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-ops {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.card-ops .el-button {
  color: #c0c4cc !important;
}

.card-ops .el-button:hover {
  color: #ff4d4f !important;
}

/* 截图区域：保持 9:16 竖屏比例 */
.screenshot-area {
  width: 100%;
  aspect-ratio: 9 / 16;
  background: #1a1a1a;
  overflow: hidden;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
}

.screenshot-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.screenshot-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #666;
  font-size: 11px;
}

.placeholder-icon {
  font-size: 24px;
  color: #555;
}

/* 投屏悬停蒙层 */
.projection-hint {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 12px;
  opacity: 0;
  transition: opacity 0.2s;
  pointer-events: none;
}

.screenshot-area:hover .projection-hint {
  opacity: 1;
}

.projection-hint .el-icon {
  font-size: 22px;
}

/* 右下角序号角标 */
.card-index {
  position: absolute;
  bottom: 4px;
  right: 6px;
  font-size: 11px;
  font-weight: 700;
  color: rgba(255,255,255,0.9);
  text-shadow: 0 1px 3px rgba(0,0,0,0.8);
  z-index: 10;
  line-height: 1;
  pointer-events: none;
}

/* 端口信息浮层（悬停显示） */

.port-row {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
}

.port-label {
  color: #b0b8c4;
}

.port-value {
  color: #e0e6ef;
  font-family: 'Consolas', monospace;
}

/* 添加弹窗 */
.add-dialog-body {
  padding: 0 4px;
}
</style>
