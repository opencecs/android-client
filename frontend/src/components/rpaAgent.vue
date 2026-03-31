<template>
  <div class="rpa-agent">
    <!-- 左栏：容器选择 + 进度面板 -->
    <div class="left-panel">
      <el-card class="device-card">
        <template #header>
          <div class="card-header">
            <span>{{ $t('common.selectContainer') }}</span>
            <div style="display:flex;gap:6px">
              <el-button size="small" :icon="Refresh" circle @click="refreshContainers" :title="$t('common.refreshModelList')" :loading="containersLoading" />
              <el-button size="small" :icon="Setting" circle @click="showSettings = true" :title="$t('common.rpaSettings')" />
            </div>
          </div>
        </template>

        <!-- 加载中 -->
        <div v-if="containersLoading" style="text-align:center;padding:16px">
          <el-icon class="rotating"><Loading /></el-icon>
          <span style="font-size:12px;color:#909399;margin-left:6px">{{ $t('common.loadingContainers') }}</span>
        </div>

        <!-- 容器列表（按主机 IP 分组） -->
        <div v-else class="container-list">
          <template v-if="containerGroups.length > 0">
            <div v-for="group in containerGroups" :key="group.deviceIP" class="device-group">
              <!-- 主机标题行 -->
              <div class="group-header" @click="toggleGroup(group.deviceIP)">
                <el-icon style="font-size:12px;color:#909399">
                  <component :is="collapsedGroups.has(group.deviceIP) ? ArrowRight : ArrowDown" />
                </el-icon>
                <span class="group-ip">{{ group.deviceIP }}</span>
                <el-tag size="small" type="info" style="margin-left:auto">{{ group.containers.length }} {{ $t('common.times') }}</el-tag>
              </div>

              <!-- 容器条目 -->
              <template v-if="!collapsedGroups.has(group.deviceIP)">
                <div
                  v-for="ct in group.containers"
                  :key="ct.key"
                  :class="['container-item', { active: selectedContainerKeys.includes(ct.key) }]"
                  @click="toggleContainer(ct)"
                >
                  <el-checkbox :model-value="selectedContainerKeys.includes(ct.key)" @change="toggleContainer(ct)" @click.stop />
                  <div class="ct-info">
                    <span class="ct-name">{{ ct.name }}</span>
                    <div class="ct-meta">
                      <el-tag size="small" :type="ct.networkMode === 'myt' ? 'success' : 'warning'" style="font-size:10px">
                        {{ ct.networkMode === 'myt' ? $t('common.bridged') : $t('common.nonBridged') }}
                      </el-tag>
                      <span class="ct-port">#{{ ct.indexNum }} · :{{ ct.rpaPort }}</span>
                    </div>
                  </div>
                </div>

                <!-- 该主机无 running 容器 -->
                <div v-if="group.containers.length === 0" class="group-empty">
                  <el-icon color="#e6a23c"><Warning /></el-icon>
                  <span>{{ $t('common.noRunningContainer') }}</span>
                </div>
              </template>
            </div>
          </template>

          <!-- 完全没有数据 -->
          <template v-else>
            <el-empty
              v-if="onlineDeviceIPs.length === 0"
              :description="$t('common.noSupportedDevice')"
              :image-size="50"
            />
            <el-alert v-else type="warning" :closable="false" style="font-size:12px">
              {{ $t('common.noAndroidContainer') }}
            </el-alert>
          </template>
        </div>

        <div class="device-actions" v-if="allContainers.length > 0">
          <el-button size="small" @click="selectAllContainers">{{ $t('common.selectAllBtn') }}</el-button>
          <el-button size="small" @click="selectedContainerKeys = []">{{ $t('common.clearBtn') }}</el-button>
        </div>
      </el-card>

      <!-- 任务进度面板 -->
      <el-card class="progress-card" v-if="taskResults.length > 0">
        <template #header>
          <div class="card-header">
            <span>{{ $t('common.executionProgress') }}</span>
            <el-button size="small" text @click="taskResults = []">{{ $t('common.clearBtn') }}</el-button>
          </div>
        </template>
        <el-collapse accordion>
          <el-collapse-item
            v-for="result in taskResults"
            :key="result.key"
            :name="result.key"
          >
            <template #title>
              <div class="progress-title">
                <el-icon v-if="result.status === 'running'" class="rotating"><Loading /></el-icon>
                <el-icon v-else-if="result.status === 'success'" color="#67c23a"><CircleCheck /></el-icon>
                <el-icon v-else-if="result.status === 'failed'" color="#f56c6c"><CircleClose /></el-icon>
                <el-icon v-else color="#909399"><Clock /></el-icon>
                <span class="progress-ip">{{ result.containerName }}</span>
                <el-tag size="small" :type="result.status === 'success' ? 'success' : result.status === 'failed' ? 'danger' : result.status === 'running' ? 'warning' : 'info'">
                  {{ statusLabel(result.status) }}
                </el-tag>
              </div>
            </template>
            <div class="progress-log">
              <div v-for="(logItem, i) in result.logs" :key="i" :class="['log-line', logItem.type]">
                <span class="log-time">{{ logItem.time }}</span>
                <span class="log-text">{{ logItem.text }}</span>
              </div>
              <div v-if="result.summary" class="log-summary">
                <el-tag :type="result.status === 'success' ? 'success' : 'danger'" size="small">{{ result.summary }}</el-tag>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </el-card>
    </div>

    <!-- 右栏：任务区 + 对话区 -->
    <div class="right-panel">
      <!-- 任务卡片（始终显示，执行中禁用切换） -->
      <div class="task-cards">
        <div class="task-cards-title">{{ $t('common.selectTaskType') }}</div>
        <div class="cards-grid">
          <div
            v-for="card in taskCards"
            :key="card.type"
            :class="['task-card', { active: selectedTaskType === card.type, disabled: isRunning }]"
            @click="!isRunning && selectTaskType(card.type)"
          >
            <div class="task-card-icon">{{ card.icon }}</div>
            <div class="task-card-name">{{ $t('common.' + card.nameKey) }}</div>
          </div>
        </div>
      </div>

      <!-- 对话区 -->
      <div class="chat-area" ref="chatAreaRef">
        <div v-for="(msg, i) in chatMessages" :key="i" :class="['chat-msg', msg.role]">
          <div class="msg-bubble">
            <div class="msg-role">{{ msg.role === 'user' ? 'You' : 'Agent' }}</div>
            <div class="msg-content" v-html="renderMarkdown(msg.content)"></div>
            <!-- tool_call 展示 -->
            <div v-if="msg.toolCalls && msg.toolCalls.length" class="tool-calls">
              <div v-for="(tc, ti) in msg.toolCalls" :key="ti" class="tool-call-item">
                <el-tag size="small" type="info">🔧 {{ tc.name }}</el-tag>
                <span class="tool-args" v-if="tc.args">{{ JSON.stringify(tc.args) }}</span>
              </div>
            </div>
            <!-- tool 结果 -->
            <div v-if="msg.role === 'tool'" class="tool-result">
              <el-tag size="small" :type="msg.success === false ? 'danger' : 'success'">
                {{ msg.success === false ? 'Failed' : 'Success' }}
              </el-tag>
              <span class="tool-result-text">{{ msg.resultText }}</span>
            </div>
          </div>
        </div>
        <div v-if="streamingText" class="chat-msg assistant">
          <div class="msg-bubble">
            <div class="msg-role">Agent</div>
            <div class="msg-content streaming">{{ streamingText }}<span class="cursor">|</span></div>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="input-area">
        <!-- ask_user 等待回复时显示特殊提示 -->
        <div v-if="waitingForUser" class="waiting-hint">
          <el-icon color="#e6a23c"><QuestionFilled /></el-icon>
          {{ $t('common.agentWaiting') }}
        </div>
        <div class="input-row">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :rows="2"
            :placeholder="inputPlaceholder"
            :disabled="isRunning && !waitingForUser"
            @keydown.enter.prevent="handleEnter"
            resize="none"
          />
          <div class="input-btns">
            <el-button
              v-if="!isRunning"
              type="primary"
              :disabled="!canStart"
              @click="startTask"
              :icon="VideoPlay"
            >
              {{ $t('common.executeTask') }}
            </el-button>
            <el-button
              v-if="isRunning && !waitingForUser"
              type="danger"
              @click="stopAllTasks"
            >
              {{ $t('common.stopModel') }}
            </el-button>
            <el-button
              v-if="waitingForUser"
              type="primary"
              @click="replyToAgent"
            >
              {{ $t('common.replyBtn') }}
            </el-button>
          </div>
        </div>
        <div class="input-tips" v-if="!isRunning">
          {{ $t('common.selectedContainers', { count: selectedContainerKeys.length }) }}
          <template v-if="selectedTaskType"> · {{ $t('common.taskLabel') }}{{ $t('common.' + currentTaskCard?.nameKey) }}</template>
        </div>
      </div>
    </div>

    <!-- RPA 设置抽屉 -->
    <el-drawer v-model="showSettings" :title="$t('common.rpaSettings')" direction="rtl" size="460px" @open="onSettingsOpen">
      <div style="padding:0 16px">

        <!-- 每台设备独立配置（仅显示支持的设备） -->
        <el-divider content-position="left">{{ $t('common.deviceModelConfig') }}</el-divider>
        <el-empty v-if="r1qDevices.length === 0" :description="$t('common.noSupportedOnlineDevice')" :image-size="40" style="padding:16px 0" />
        <el-collapse v-model="settingsOpenPanels">
          <el-collapse-item
            v-for="d in r1qDevices"
            :key="d.ip"
            :name="d.ip"
          >
            <template #title>
              <div style="display:flex;align-items:center;gap:8px;width:100%;padding-right:8px">
                <el-tag size="small" :type="props.devicesStatusCache.get(d.id) === 'online' ? 'success' : 'info'">
                  {{ props.devicesStatusCache.get(d.id) === 'online' ? $t('common.online') : $t('common.offline') }}
                </el-tag>
                <span style="font-weight:500">{{ d.ip }}</span>
                <!-- 模型运行状态 -->
                <template v-if="modelStatusMap[d.ip]">
                  <el-tag size="small" :type="modelStatusMap[d.ip].llm === 'running' ? 'success' : modelStatusMap[d.ip].llm === 'starting' ? 'warning' : 'danger'" style="margin-left:auto">
                    <el-icon v-if="modelStatusMap[d.ip].llm === 'starting'" class="rotating"><Loading /></el-icon>
                    LLM {{ modelStatusLabel(modelStatusMap[d.ip].llm) }}
                  </el-tag>
                </template>
              </div>
            </template>

            <el-form label-width="120px" style="padding:8px 0 0 0">
              <el-form-item :label="$t('common.llmModel')">
                <el-select
                  :model-value="settingsDraft[d.ip]?.llmModel || ''"
                  @update:model-value="v => setDraft(d.ip, 'llmModel', v)"
                  :placeholder="$t('common.selectChatModelRPA')"
                  style="width:100%"
                  filterable
                  :loading="settingsLoadingMap[d.ip]"
                  :no-data-text="settingsModelMap[d.ip] === undefined ? '' + $t('common.clickRefreshToLoad') + '' : '' + $t('common.noModelOnDevice') + ''"
                >
                  <el-option v-for="m in (settingsModelMap[d.ip] || [])" :key="m.name" :label="m.name" :value="m.name" />
                </el-select>
              </el-form-item>
              <el-form-item label=" ">
                <el-button size="small" :loading="settingsLoadingMap[d.ip]" @click="loadModelListForIp(d.ip)">
                  {{ $t('common.refreshModelList') }}
                </el-button>
                <el-button
                  size="small"
                  type="primary"
                  :loading="modelStartingMap[d.ip]"
                  :disabled="!settingsDraft[d.ip]?.llmModel || props.devicesStatusCache.get(d.id) !== 'online'"
                  @click="checkAndStartModels(d.ip)"
                  style="margin-left:8px"
                >
                  {{ $t('common.checkAndStartModel') }}
                </el-button>
              </el-form-item>
            </el-form>
          </el-collapse-item>
        </el-collapse>

        <!-- 全局参数 -->
        <el-divider content-position="left">{{ $t('common.agentGlobalParams') }}</el-divider>
        <el-form label-width="120px">
          <el-form-item :label="$t('common.breakerRoundLimit')">
            <el-input-number v-model="globalConfig.maxRounds" :min="0" :max="200" style="width:100%" />
            <div style="font-size:11px;color:#909399;margin-top:4px">{{ $t('common.breakerHint') }}</div>
          </el-form-item>
          <el-form-item :label="$t('common.stepDelay')">
            <el-input-number v-model="globalConfig.stepDelayMs" :min="0" :max="5000" :step="100" style="width:100%" />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <el-button @click="saveSettings">{{ $t('common.saveConfig') }}</el-button>
        <el-button @click="showSettings = false">{{ $t('common.closeBtn') }}</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
// 返回设备的 host:port，若 ip 已含端口则直接使用，否则追加默认 8000
const getDeviceAddr = (ip) => {
  if (!ip) return ip
  const lastColon = ip.lastIndexOf(':')
  if (lastColon === -1) return ip + ':8000'
  return /^\d+$/.test(ip.slice(lastColon + 1)) ? ip : ip + ':8000'
}


import { ref, computed, watch, nextTick, onMounted, getCurrentInstance } from 'vue'
// i18n helper for script-level translations
const _instance = getCurrentInstance()
const i18n = { t: (key, params) => _instance?.proxy?.$t(key, params) || key }

import { ElMessage } from 'element-plus'
import {
  Setting, VideoPlay, Loading, CircleCheck, CircleClose, Clock, QuestionFilled,
  Refresh, ArrowRight, ArrowDown, Warning
} from '@element-plus/icons-vue'
import {
  AISendMessage,
  AIGetModels, AIStartModel, AIStopModel,
  GetLLMModelList,
  RpaGetScreenContext, RpaClick, RpaSwipe,
  RpaSendText, RpaKeyPress, RpaOpenApp, RpaStopApp,
  RpaCloseConnection, RpaGetToolsJSON, RpaBuildSystemPrompt, RpaCallFixedTool,
  RpaGetAndroidContainers, TriggerAndroidRefresh
} from '../../bindings/edgeclient/app'
import { Events } from '@wailsio/runtime'
import MarkdownIt from 'markdown-it'

// ============================================================
// Markdown 渲染
// ============================================================
const md = new MarkdownIt({ html: false, linkify: true })
const renderMarkdown = (text) => {
  if (!text) return ''
  return md.render(String(text))
}

// ============================================================
// Props & Expose
// ============================================================
const props = defineProps({
  devices: { type: Array, default: () => [] },
  token: { type: String, default: '' },
  devicesStatusCache: { type: Map, default: () => new Map() }
})

// ============================================================
// 设置（按设备 IP 独立保存）
// ============================================================
const SETTINGS_KEY = 'rpa_agent_config_v2'   // { [ip]: { llmModel } }
const GLOBAL_KEY   = 'rpa_agent_global_v3'   // { maxRounds, stepDelayMs }

const globalConfig  = ref({ maxRounds: 0, stepDelayMs: 1500 })
const deviceConfigs = ref({})   // { [ip]: { llmModel } }

const getDeviceConfig = (ip) => deviceConfigs.value[ip] || { llmModel: '' }

const loadSettings = () => {
  try {
    const g = localStorage.getItem(GLOBAL_KEY)
    if (g) Object.assign(globalConfig.value, JSON.parse(g))
    const d = localStorage.getItem(SETTINGS_KEY)
    if (d) deviceConfigs.value = JSON.parse(d)
  } catch (e) { /* ignore */ }
}

const showSettings = ref(false)

// ---- 抽屉内部状态 ----
const settingsOpenPanels = ref([])
const settingsDraft      = ref({})
const settingsModelMap   = ref({})
const settingsLoadingMap = ref({})

const setDraft = (ip, field, value) => {
  if (!settingsDraft.value[ip]) settingsDraft.value[ip] = { llmModel: '' }
  settingsDraft.value[ip][field] = value
}

const onSettingsOpen = async () => {
  const draft = {}
  for (const d of r1qDevices.value) {
    draft[d.ip] = { ...getDeviceConfig(d.ip) }
  }
  settingsDraft.value = draft

  const openIps = selectedContainers.value.map(ct => ct.deviceIP)
  settingsOpenPanels.value = [...new Set(openIps)]

  await Promise.allSettled([
    ...settingsOpenPanels.value.filter(ip => settingsModelMap.value[ip] === undefined).map(ip => loadModelListForIp(ip)),
    ...r1qDevices.value.map(d => {
      const cfg = getDeviceConfig(d.ip)
      return refreshModelStatus(d.ip, cfg.llmModel)
    })
  ])
}

const loadModelListForIp = async (ip) => {
  settingsLoadingMap.value = { ...settingsLoadingMap.value, [ip]: true }
  try {
    const result = await GetLLMModelList(ip, props.token)
    if (result?.success) {
      settingsModelMap.value = { ...settingsModelMap.value, [ip]: result.data?.list || [] }
    } else {
      settingsModelMap.value = { ...settingsModelMap.value, [ip]: [] }
      ElMessage.error(`${ip} 获取模型列表失败: ${result?.message || t('aiAssistant.fetchModelFailed')}`)
    }
  } catch (e) {
    settingsModelMap.value = { ...settingsModelMap.value, [ip]: [] }
    ElMessage.error(`${ip} ${t('rpaAgent.fetchModelFailed')}${e.message}`)
  } finally {
    settingsLoadingMap.value = { ...settingsLoadingMap.value, [ip]: false }
  }
}

watch(settingsOpenPanels, (newVal, oldVal) => {
  const newlyOpened = newVal.filter(ip => !oldVal.includes(ip))
  for (const ip of newlyOpened) {
    if (settingsModelMap.value[ip] === undefined) {
      loadModelListForIp(ip)
    }
  }
})

const saveSettings = () => {
  for (const ip of Object.keys(settingsDraft.value)) {
    deviceConfigs.value[ip] = { ...settingsDraft.value[ip] }
  }
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(deviceConfigs.value))
  localStorage.setItem(GLOBAL_KEY, JSON.stringify(globalConfig.value))
  ElMessage.success(t('rpaAgent.settingsSaved'))
}

// ============================================================
// 模型状态检查 & 启动
// ============================================================
const modelStatusMap   = ref({})
const modelStartingMap = ref({})

const modelStatusLabel = (s) => {
  if (s === 'running')  return t('rpaAgent.running')
  if (s === 'starting') return t('rpaAgent.starting')
  if (s === 'stopped')  return t('rpaAgent.stopped')
  return i18n.t('common.modelStatusUnknown')
}

const refreshModelStatus = async (ip, llmModel) => {
  if (!llmModel) {
    modelStatusMap.value = { ...modelStatusMap.value, [ip]: null }
    return
  }
  try {
    const r = await AIGetModels(ip)
    if (!r?.success) {
      modelStatusMap.value = { ...modelStatusMap.value, [ip]: { llm: 'unknown' } }
      return
    }
    const raw = r.data
    if (raw?.code === 502 || raw?.code === '502') {
      modelStatusMap.value = { ...modelStatusMap.value, [ip]: { llm: 'stopped' } }
      return
    }
    const ml = raw?.object === 'list' ? raw : (raw?.data?.object === 'list' ? raw.data : null)
    const runningModels = ml?.data || []
    const modelMatched = runningModels.some(m => {
      const mid = (m.id || '')
      return mid === llmModel || mid.endsWith('/' + llmModel) || llmModel.endsWith('/' + mid) || mid.split('/').pop() === llmModel.split('/').pop()
    })
    modelStatusMap.value = {
      ...modelStatusMap.value,
      [ip]: { llm: modelMatched ? 'running' : 'stopped' }
    }
  } catch (e) {
    modelStatusMap.value = { ...modelStatusMap.value, [ip]: { llm: 'unknown' } }
  }
}

const checkAndStartModels = async (ip) => {
  const cfg = settingsDraft.value[ip] || deviceConfigs.value[ip] || {}
  const llmModel = cfg.llmModel
  if (!llmModel) { ElMessage.warning(`${ip} ${t('rpaAgent.noLlmModel')}`); return }

  modelStartingMap.value = { ...modelStartingMap.value, [ip]: true }

  try {
    modelStatusMap.value = { ...modelStatusMap.value, [ip]: { llm: 'starting' } }
    ElMessage.info(`${ip} ${t('rpaAgent.stoppingOld')}`)
    await AIStopModel(ip)
    await sleep(2000)

    const listResult = await GetLLMModelList(ip, props.token)
    const allModels = listResult?.data?.list || []

    const parseFiles = (model) => {
      const files = model.files || []
      let modelPath = '', weightPath = '', vocabPath = '', embedPath = '', chatTemplateFile = ''
      files.forEach(f => {
        const n = (f.filePath || '').split('/').pop()
        if (n.endsWith('.rknn') && !n.includes('vision')) modelPath = f.filePath
        else if (n.endsWith('.weight') && !n.includes('vision')) weightPath = f.filePath
        else if (n.endsWith('.gguf')) vocabPath = f.filePath
        else if (n.includes('.embed.bin')) embedPath = f.filePath
        else if (n.endsWith('.jinja')) chatTemplateFile = f.filePath  // 直接从文件列表取 jinja 路径
      })
      return { modelPath, weightPath, vocabPath, embedPath, chatTemplateFile }
    }

    const modelConfig = { host: '0.0.0.0', port: 8081, timeout: 30, models: {} }
    const chatModel = allModels.find(m => m.name === llmModel)
    if (!chatModel) { ElMessage.error(`${ip} ${t('rpaAgent.modelNotFound')} "${llmModel}"`); return }
    const f = parseFiles(chatModel)
    modelConfig.models[llmModel] = {
      alias: llmModel, model: f.modelPath, weight: f.weightPath,
      model2: '', weight2: '', model3: '', weight3: '',
      vocab: f.vocabPath, embed: f.embedPath,
      'mel-filter': '', 'ctx-size': 2048, 'predict': 512,
      'temp': 0.1, 'top-k': 1, 'top-p': 0.8,
      'repeat-penalty': 1.1, 'presence-penalty': 1.0, 'frequency-penalty': 1.0,
      'img-start': '', 'img-end': '', 'img-content': '',
      'audio-start': '', 'audio-end': '', 'audio-content': '',
      'img-width': 0, 'img-height': 0, 'chat-template-file': f.chatTemplateFile, embedding: false
    }

    const startResult = await AIStartModel({ deviceIp: ip, modelConfig })
    if (!startResult?.success) {
      ElMessage.error(`${ip} ${t('rpaAgent.startFailed')}${startResult?.message || t('common.unknownError')}`)
      await refreshModelStatus(ip, llmModel)
      return
    }

    ElMessage.info(`${ip} ${t('rpaAgent.startingWait')}`)
    let ready = false
    for (let i = 0; i < 36; i++) {
      await sleep(5000)
      await refreshModelStatus(ip, llmModel)
      if (modelStatusMap.value[ip]?.llm === 'running') { ready = true; break }
    }
    if (ready) {
      ElMessage.success(`${ip} ${t('rpaAgent.modelReady')}`)
    } else {
      ElMessage.error(`${ip} ${t('rpaAgent.startTimeout')}`)
    }
  } catch (e) {
    ElMessage.error(`${ip} 启动失败: ${e.message}`)
    await refreshModelStatus(ip, llmModel)
  } finally {
    modelStartingMap.value = { ...modelStartingMap.value, [ip]: false }
  }
}

// ============================================================
// 安卓容器列表
// ============================================================
// containerGroups: [{deviceIP, containers:[{key, name, networkMode, indexNum, rpaIp, rpaPort, deviceIP}]}]
const containerGroups   = ref([])
const containersLoading = ref(false)
const collapsedGroups   = ref(new Set())

// ============================================================
// 支持设备过滤（RPA Agent 支持 r1q 和 eces-rk3588-rk1828 机型设备）
// ============================================================
const deviceModelsCache = ref(new Map())  // Map<ip, deviceInfo>

// 认证头辅助（复用 aiAssistant.vue 相同逻辑）
const getAuthHeaders = (deviceIP) => {
  try {
    const savedPassword = localStorage.getItem('devicePasswords')
    const passwords = JSON.parse(savedPassword || '{}')
    const password = passwords[deviceIP] || null
    if (password) {
      return { 'Authorization': `Basic ${btoa(`admin:${password}`)}` }
    }
  } catch (_e) { /* ignore */ }
  return {}
}

const fetchDeviceInfo = async (deviceIP) => {
  if (deviceModelsCache.value.has(deviceIP)) {
    return deviceModelsCache.value.get(deviceIP)
  }
  try {
    const response = await fetch(`http://${getDeviceAddr(deviceIP)}/info/device`, {
      method: 'GET',
      headers: getAuthHeaders(deviceIP)
    })
    if (response.ok) {
      const data = await response.json()
      if (data.code === 0 && data.data) {
        deviceModelsCache.value.set(deviceIP, data.data)
        return data.data
      }
    }
  } catch (e) {
    console.warn(`[RpaAgent] 获取设备 ${deviceIP} 信息失败:`, e)
  }
  return null
}

const isR1qDevice = (deviceIP) => {
  const info = deviceModelsCache.value.get(deviceIP)
  if (!info) return false
  const version = (info.version || '').toLowerCase()
  const model = (info.model || '').toLowerCase()
  // 支持 r1q 和 eces-rk3588-rk1828 机型
  const isR1q = info.rk182x === 'y' && version.includes('r1q')
  const isEces = model.includes('eces-rk3588-rk1828') || version.includes('eces-rk3588-rk1828')
  return isR1q || isEces
}

// 在线且为支持机型的设备列表（用于设置抽屉）
const r1qDevices = computed(() =>
  props.devices.filter(d =>
    props.devicesStatusCache.get(d.id) === 'online' && isR1qDevice(d.ip)
  )
)

// 在线且为支持机型的设备 IP 列表
const onlineDeviceIPs = computed(() =>
  r1qDevices.value.map(d => d.ip)
)

// 所有容器平铺列表
const allContainers = computed(() => {
  const list = []
  for (const group of containerGroups.value) {
    for (const ct of group.containers) list.push(ct)
  }
  return list
})

// 加载容器列表（读 Go 后端缓存）
const loadContainers = async (triggerRefresh = false) => {
  if (onlineDeviceIPs.value.length === 0) {
    containerGroups.value = []
    return
  }
  containersLoading.value = true
  try {
    if (triggerRefresh) {
      // 先触发一次强制刷新
      console.log('[RpaAgent] 触发安卓容器缓存刷新 ips=', onlineDeviceIPs.value)
      await TriggerAndroidRefresh(onlineDeviceIPs.value)
      await sleep(1500)  // 等待轮询完成
    }

    console.log('[RpaAgent] 读取安卓容器列表 ips=', onlineDeviceIPs.value)
    const result = await RpaGetAndroidContainers(onlineDeviceIPs.value)
    console.log('[RpaAgent] 容器列表结果:', result)

    const groups = []
    for (const ip of onlineDeviceIPs.value) {
      const deviceResult = result?.[ip]
      const containers = []
      if (deviceResult && deviceResult.status === 'ok' && Array.isArray(deviceResult.containers)) {
        for (const ct of deviceResult.containers) {
          containers.push({
            key:         `${ct.deviceIP}_${ct.name}`,
            name:        ct.name,
            networkMode: ct.networkMode || 'bridge',
            indexNum:    ct.indexNum || 0,
            containerIP: ct.containerIP || '',
            rpaIp:       ct.rpaIp || ip,
            rpaPort:     ct.rpaPort || 9083,
            deviceIP:    ct.deviceIP || ip,
          })
        }
      }
      groups.push({ deviceIP: ip, containers, error: deviceResult?.error || '' })
    }
    containerGroups.value = groups
  } catch (e) {
    console.error('[RpaAgent] 加载容器列表失败:', e)
    ElMessage.error(t('rpaAgent.loadContainerFailed') + e.message)
  } finally {
    containersLoading.value = false
  }
}

const refreshContainers = () => loadContainers(true)

// 设备列表变化时，拉取未缓存设备的 r1q 信息
watch(() => props.devices, async (newDevices) => {
  const onlineDevices = newDevices.filter(d => props.devicesStatusCache.get(d.id) === 'online')
  const needFetch = onlineDevices.filter(d => !deviceModelsCache.value.has(d.ip))
  if (needFetch.length > 0) {
    await Promise.allSettled(needFetch.map(d => fetchDeviceInfo(d.ip)))
  }
}, { deep: true })

// 在线设备变化时自动刷新容器列表
watch(onlineDeviceIPs, (newIps, oldIps) => {
  const changed = newIps.length !== oldIps?.length ||
    newIps.some((ip, i) => ip !== oldIps?.[i])
  if (changed) loadContainers(false)
}, { deep: true })

// 折叠/展开分组
const toggleGroup = (deviceIP) => {
  const s = new Set(collapsedGroups.value)
  if (s.has(deviceIP)) s.delete(deviceIP)
  else s.add(deviceIP)
  collapsedGroups.value = s
}

// ============================================================
// 容器选择
// ============================================================
const selectedContainerKeys = ref([])

const toggleContainer = (ct) => {
  const idx = selectedContainerKeys.value.indexOf(ct.key)
  if (idx >= 0) selectedContainerKeys.value.splice(idx, 1)
  else selectedContainerKeys.value.push(ct.key)
}

const selectAllContainers = () => {
  selectedContainerKeys.value = allContainers.value.map(ct => ct.key)
}

const selectedContainers = computed(() =>
  allContainers.value.filter(ct => selectedContainerKeys.value.includes(ct.key))
)

// ============================================================
// 任务卡片
// ============================================================
const taskCards = [
  { type: 'browse_video', nameKey: 'taskBrowseVideo', icon: '📱' },
  { type: 'send_message', nameKey: 'taskSendMessage', icon: '💬' },
  { type: 'install_app', nameKey: 'taskInstallApp', icon: '📦' },
  { type: 'like_content', nameKey: 'taskLikeContent', icon: '❤️' },
  { type: 'search', nameKey: 'taskSearch', icon: '🔍' },
  { type: 'custom', nameKey: 'taskCustom', icon: '⚙️' }
]
const selectedTaskType = ref('')
const currentTaskCard = computed(() => taskCards.find(c => c.type === selectedTaskType.value))

const selectTaskType = (type) => {
  selectedTaskType.value = type
  const hints = {
    browse_video: '打开抖音，向上滑动刷5条视频',
    send_message: '打开微信，给文件传输助手发送一条消息：测试',
    install_app: '安装应用，APK路径：',
    like_content: '打开抖音，给当前视频点赞',
    search: '打开抖音，搜索关键词：',
    custom: ''
  }
  if (!inputMessage.value) inputMessage.value = hints[type] || ''
}

// ============================================================
// 对话消息 & 流式文本
// ============================================================
const chatMessages  = ref([])
const streamingText = ref('')
const chatAreaRef   = ref(null)

const scrollToBottom = () => {
  nextTick(() => {
    if (chatAreaRef.value) chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight
  })
}

const addMsg = (role, content, extra = {}) => {
  chatMessages.value.push({ role, content, ...extra })
  scrollToBottom()
}

// ============================================================
// 任务状态
// ============================================================
const isRunning       = ref(false)
const waitingForUser  = ref(false)
const taskResults     = ref([])   // [{key, containerName, deviceIP, status, logs[], summary}]
const abortFlags      = ref({})   // containerKey → true 表示要中止

const inputMessage    = ref('')
const inputPlaceholder = computed(() => {
  if (waitingForUser.value) return i18n.t('common.agentWaitingReply')
  if (isRunning.value) return i18n.t('common.taskExecuting')
  return i18n.t('common.taskInputHint')
})

const canStart = computed(() =>
  selectedContainerKeys.value.length > 0 && inputMessage.value.trim().length > 0
)

// ============================================================
// 状态标签
// ============================================================
const statusLabel = (status) => {
  const map = { pending: t('common.pending'), running: t('common.running'), success: t('common.success'), failed: t('common.failed'), stopped: t('common.stopped') }
  return map[status] || status
}

// ============================================================
// 全并发模式：所有选中容器同时执行，不再按设备串行
// 802 / ctx-size=16384 支持多容器并发推理，502 由 agentLoop 内部重试处理
// ============================================================

// ============================================================
// 启动任务（以容器为粒度，完全并发）
// ============================================================
const startTask = async () => {
  if (!canStart.value) return

  // 检查所有选中容器的 deviceIP 是否都配置了 LLM 模型
  const needIPs = [...new Set(selectedContainers.value.map(ct => ct.deviceIP))]
  const unconfigured = needIPs.filter(ip => !getDeviceConfig(ip).llmModel)
  if (unconfigured.length > 0) {
    ElMessage.warning(`${t('rpaAgent.unconfiguredNodes')} ${unconfigured.join(', ')}`)
    showSettings.value = true
    return
  }

  const userTask = inputMessage.value.trim()
  isRunning.value = true
  chatMessages.value = []
  streamingText.value = ''
  waitingForUser.value = false
  abortFlags.value = {}

  addMsg('user', userTask)

  // 为每个容器初始化进度记录
  taskResults.value = selectedContainers.value.map(ct => ({
    key:           ct.key,
    containerName: `${ct.deviceIP} / ${ct.name}`,
    deviceIP:      ct.deviceIP,
    status:        'pending',
    logs:          [],
    summary:       ''
  }))

  console.log(`[RpaAgent] 启动任务，容器数=${selectedContainers.value.length}，任务=${userTask}，模式=全并发`)

  // 全并发：所有容器同时开始，502 由 agentLoop 内部重试处理
  try {
    await Promise.allSettled(selectedContainers.value.map(ct => runAgentLoop(ct, userTask)))
  } finally {
    isRunning.value = false
    waitingForUser.value = false
  }
}

// ============================================================
// 单容器 agentLoop（无轮次上限，直到 task_done / abort）
// ============================================================
const runAgentLoop = async (container, userTask) => {
  const { key, name: containerName, deviceIP, rpaIp, rpaPort } = container
  const devCfg = getDeviceConfig(deviceIP)
  // RPA 请求对象（传给 RpaXxx 系列 IPC 方法）
  const rpaReq = { deviceIp: rpaIp, rpaPort, containerId: containerName, password: '' }

  const result = taskResults.value.find(r => r.key === key)
  result.status = 'running'

  const log = (text, type = 'info') => {
    const time = new Date().toLocaleTimeString()
    result.logs.push({ time, text, type })
    console.log(`[AgentLoop][${deviceIP}/${containerName}][${type}] ${text}`)
    // 多容器时加标识
    if (selectedContainers.value.length > 1) {
      addMsg('assistant', `[${deviceIP}/${containerName}] ${text}`)
    }
  }

  log(`开始执行任务：${userTask}`)
  log(`容器=${containerName} 网络=${container.networkMode} 坑位=${container.indexNum} RPA=${rpaIp}:${rpaPort}`)

  // ── 启动模型并等待就绪的公共辅助（agentLoop 内复用）──
  // 返回 true=就绪, false=失败
  const startModelAndWait = async () => {
    const listResult = await GetLLMModelList(deviceIP, props.token)
    const allModels = listResult?.data?.list || []
    const chatModel = allModels.find(m => m.name === devCfg.llmModel)
    if (!chatModel) {
      log(`设备上找不到模型 "${devCfg.llmModel}"，请检查设置`, 'error')
      return false
    }
    const parseFiles = (model) => {
      const files = model.files || []
      let modelPath = '', weightPath = '', vocabPath = '', embedPath = '', chatTemplateFile = ''
      files.forEach(f => {
        const n = (f.filePath || '').split('/').pop()
        if (n.endsWith('.rknn') && !n.includes('vision')) modelPath = f.filePath
        else if (n.endsWith('.weight') && !n.includes('vision')) weightPath = f.filePath
        else if (n.endsWith('.gguf')) vocabPath = f.filePath
        else if (n.includes('.embed.bin')) embedPath = f.filePath
        else if (n.endsWith('.jinja')) chatTemplateFile = f.filePath
      })
      return { modelPath, weightPath, vocabPath, embedPath, chatTemplateFile }
    }
    const chatFiles = parseFiles(chatModel)
    const modelCfg = {
      host: '0.0.0.0', port: 8081, timeout: 30,
      models: {
        [devCfg.llmModel]: {
          alias: devCfg.llmModel,
          model: chatFiles.modelPath, weight: chatFiles.weightPath,
          model2: '', weight2: '', model3: '', weight3: '',
          vocab: chatFiles.vocabPath, embed: chatFiles.embedPath,
          'mel-filter': '', 'ctx-size': 2048, 'predict': 512,
          'temp': 0.1, 'top-k': 1, 'top-p': 0.8,
          'repeat-penalty': 1.1, 'presence-penalty': 1.0, 'frequency-penalty': 1.0,
          'img-start': '', 'img-end': '', 'img-content': '',
          'audio-start': '', 'audio-end': '', 'audio-content': '',
          'img-width': 0, 'img-height': 0,
          'chat-template-file': chatFiles.chatTemplateFile, embedding: false
        }
      }
    }
    // AIStartModel 本身可能返回 502（服务正在重启中），重试几次
    let startOk = false
    for (let si = 0; si < 6; si++) {
      const startResult = await AIStartModel({ deviceIp: deviceIP, modelConfig: modelCfg })
      if (startResult?.success) { startOk = true; break }
      log(`启动请求 502，5 秒后重试 (${si+1}/6)...`, 'warn')
      await sleep(5000)
    }
    if (!startOk) {
      log('模型启动请求持续失败', 'error')
      return false
    }
    log('等待模型就绪...')
    for (let i = 0; i < 36; i++) {
      await sleep(5000)
      try {
        const pr = await AIGetModels(deviceIP)
        const raw = pr?.data
        if (raw?.code !== 502 && raw?.code !== '502') {
          const ml = raw?.object === 'list' ? raw : (raw?.data?.object === 'list' ? raw.data : null)
          if (ml?.data?.find(m => {
            const mid = m.id || ''
            return mid === devCfg.llmModel || mid.split('/').pop() === devCfg.llmModel.split('/').pop()
          })) { log('模型已就绪'); return true }
        }
      } catch (_e) { /* ignore */ }
    }
    log('模型启动超时（3分钟）', 'error')
    return false
  }

  // 0. 检查/启动模型
  if (devCfg.llmModel) {
    log('检查模型运行状态...')
    try {
      const modelsResult = await AIGetModels(deviceIP)
      let modelRunning = false
      if (modelsResult?.success) {
        const raw = modelsResult.data
        if (raw?.code !== 502 && raw?.code !== '502') {
          const modelsList = raw?.object === 'list' ? raw : (raw?.data?.object === 'list' ? raw.data : null)
          if (modelsList?.data?.length > 0) {
            modelRunning = !!modelsList.data.find(m => {
              const mid = (m.id || '')
              return mid === devCfg.llmModel || mid.endsWith('/' + devCfg.llmModel) || devCfg.llmModel.endsWith('/' + mid) || mid.split('/').pop() === devCfg.llmModel.split('/').pop()
            })
          }
        }
      }
      if (modelRunning) {
        log('模型已在运行')
      } else {
        log('模型未运行，正在启动...', 'warn')
        const ok = await startModelAndWait()
        if (!ok) {
          result.status = 'failed'
          result.summary = '模型启动失败'
          return
        }
      }
    } catch (e) {
      log(`模型状态检查失败: ${e.message}，继续尝试...`, 'warn')
    }
  }


  // 1. 构建工具定义和 system prompt
  let tools = []
  let systemPrompt = ''
  try {
    tools = await RpaGetToolsJSON(selectedTaskType.value || 'custom')
    systemPrompt = await RpaBuildSystemPrompt(
      selectedTaskType.value || 'custom',
      '1080x1920',
      ''
    )
  } catch (e) {
    log(`获取工具定义失败: ${e.message}`, 'error')
    result.status = 'failed'
    result.summary = '初始化失败'
    return
  }

  // 2. 初始消息
  const messages = [
    { role: 'system', content: systemPrompt },
    { role: 'user', content: userTask }
  ]

  // custom 任务：代码强制先执行 get_screen_context，把当前界面节点注入历史
  // 让模型看到真实屏幕再决策，而不是凭空猜操作
  if ((selectedTaskType.value || 'custom') === 'custom') {
    log(`[custom] 自动获取当前界面节点...`)
    let ctxResult
    try {
      ctxResult = await executeToolCall('get_screen_context', {}, rpaReq, container)
    } catch (e) {
      ctxResult = { success: false, message: e.message }
    }
    const nodesSummary = ctxResult?.content || (ctxResult?.success === false ? `获取失败: ${ctxResult?.message}` : JSON.stringify(ctxResult))
    log(`[custom] 当前界面: ${nodesSummary.slice(0, 100)}`)
    // 注入到历史：模型第一轮已知当前屏幕状态，直接决策下一步
    messages.push({ role: 'assistant', content: `我执行了 get_screen_context，参数：{}` })
    messages.push({ role: 'user',      content: `get_screen_context 执行结果：${nodesSummary}` })
  }





  // 3. Agent 循环（无轮次上限，0 = 无限，> 0 = 熔断）
  const maxRounds = globalConfig.value.maxRounds || 0
  let round = 0
  let noToolRetryCount = 0   // 模型连续无工具调用的次数（超过3次强制结束）
  let swipeCount = 0         // swipe 执行次数，注入到工具结果让模型计数
  let pendingToolCalls = []  // 模型一次输出多个 tool_calls 时，暂存待执行的队列

  // 从任务描述中提取目标滑动次数（如"刷5条" → 5）
  const swipeTargetMatch = userTask.match(/(?:刷|滑动|滑|看)\s*(\d+)\s*(?:条|次|个|下)/)
  const swipeTarget = swipeTargetMatch ? parseInt(swipeTargetMatch[1]) : 0

  while (true) {
    // abort 检查
    if (abortFlags.value[key]) {
      log('任务已被用户停止', 'warn')
      result.status = 'stopped'
      result.summary = '用户停止'
      try { await RpaCloseConnection(rpaReq) } catch (_e) { /* ignore */ }
      return
    }

    // 熔断检查
    if (maxRounds > 0 && round >= maxRounds) {
      log(`超过熔断轮次 ${maxRounds}，任务结束`, 'warn')
      result.status = 'failed'
      result.summary = '超过最大轮次'
      try { await RpaCloseConnection(rpaReq) } catch (_e) { /* ignore */ }
      return
    }

    round++
    log(`第 ${round} 轮决策...`)

    if (round > 1 && globalConfig.value.stepDelayMs > 0) {
      await sleep(globalConfig.value.stepDelayMs)
    }


    // ── 优先消费暂存队列，队列空时才调 LLM ──
    let tc
    if (pendingToolCalls.length > 0) {
      tc = pendingToolCalls.shift()
      log(`[队列] 执行暂存工具: ${tc.name} (剩余 ${pendingToolCalls.length} 个)`)
    } else {
      // 调用 LLM（SSE 流式）
      // noToolRetryCount > 0 时切换到 toolChoice='required' 强制模型调工具
      const forceToolChoice = noToolRetryCount > 0 ? 'required' : 'auto'
      let llmResponse
      // LLM 调用失败时（502 模型重载中），最多等待重试 3 次，轮询直到模型就绪
      let llmRetry = 0
      try {
        while (true) {
          try {
            llmResponse = await callLLM(messages, tools, deviceIP, devCfg.llmModel, forceToolChoice)
            break  // 成功
          } catch (e) {
            const is502 = e.message && (e.message.includes('502') || e.message.includes('Target Service'))
            if (is502 && llmRetry < 3) {
              llmRetry++
              log(`LLM 502（模型重载中），等待模型就绪... (${llmRetry}/3)`, 'warn')
              // 轮询等待模型真正就绪（最多等 90 秒）
              let recovered = false
              for (let pi = 0; pi < 18; pi++) {
                await sleep(5000)
                try {
                  const pr = await AIGetModels(deviceIP)
                  const raw = pr?.data
                  if (raw?.code !== 502 && raw?.code !== '502') {
                    const ml = raw?.object === 'list' ? raw : (raw?.data?.object === 'list' ? raw.data : null)
                    if (ml?.data?.find(m => {
                      const mid = m.id || ''
                      return mid === devCfg.llmModel || mid.split('/').pop() === devCfg.llmModel.split('/').pop()
                    })) { recovered = true; break }
                  }
                } catch (_e2) { /* ignore */ }
              }
              if (recovered) {
                log('模型已恢复，继续推理', 'warn')
                continue
              }
              // 90 秒未自动恢复，主动重启模型
              log('模型 90 秒未恢复，主动重启...', 'warn')
              const ok = await startModelAndWait()
              if (ok) { continue }  // 重启成功，重试推理
              // 重启也失败，放弃
            }
            log(`LLM 调用失败: ${e.message}`, 'error')
            result.status = 'failed'
            result.summary = 'LLM 调用失败'
            try { await RpaCloseConnection(rpaReq) } catch (_e) { /* ignore */ }
            return
          }
        }
      } finally {
        // lock removed: 全并发模式无需互斥锁
      }

      const { text, toolCalls } = llmResponse

      if (text) {
        if (selectedContainers.value.length === 1) {
          addMsg('assistant', text, { toolCalls })
        }
        log(`LLM: ${text.slice(0, 100)}${text.length > 100 ? '...' : ''}`)
      }

      if (!text && (!toolCalls || toolCalls.length === 0)) {
        noToolRetryCount++
        if (noToolRetryCount >= 5) {
          log('模型连续 5 次无有效响应，任务中止', 'error')
          result.status = 'failed'
          result.summary = 'LLM 持续空响应'
          try { await RpaCloseConnection(rpaReq) } catch (_e) { /* ignore */ }
          return
        }
        log(`LLM 空响应（第${noToolRetryCount}次），切换 required 重试...`, 'warn')
        continue
      }

      if (!toolCalls || toolCalls.length === 0) {
        noToolRetryCount++
        if (noToolRetryCount >= 5) {
          log('模型连续 5 次未调工具，强制结束任务', 'warn')
          result.status = 'failed'
          result.summary = text?.slice(0, 80) || '模型未调工具'
          try { await RpaCloseConnection(rpaReq) } catch (_e) { /* ignore */ }
          return
        }
        log(`模型未调工具（第${noToolRetryCount}次），toolChoice=required 重试...`, 'warn')
        if (noToolRetryCount >= 2) {
          if (text) messages.push({ role: 'assistant', content: text })
          messages.push({
            role: 'user',
            content: `你必须调用工具。当前任务：${userTask}。请直接调用工具，不要输出文字。`
          })
        }
        continue
      }

      // 模型一次输出多个 tool_calls：取第一个执行，其余暂存队列
      tc = toolCalls[0]
      if (toolCalls.length > 1) {
        // 过滤掉 argsStr 解析失败的（如 get_screen_context 空参数是正常的，保留；其他空参的跳过）
        const rest = toolCalls.slice(1).filter(t => t.name === 'get_screen_context' || (t.args && Object.keys(t.args).length > 0))
        if (rest.length > 0) {
          pendingToolCalls.push(...rest)
          log(`模型输出了 ${toolCalls.length} 个工具调用，暂存后 ${rest.length} 个: ${rest.map(t=>t.name).join(', ')}`)
        }
      }
    }

    noToolRetryCount = 0
    const toolName = tc.name
    const toolArgs = tc.args || {}

    // 防御：args 是字符串说明 JSON 截断/损坏，跳过该工具调用
    if (typeof toolArgs === 'string') {
      log(`工具 ${toolName} 参数解析失败（截断），跳过`, 'warn')
      continue
    }

    log(`调用工具: ${toolName} args=${JSON.stringify(toolArgs)}`)



    // task_done → 结束任务
    if (toolName === 'task_done') {
      const success = toolArgs.success !== false
      const summary = toolArgs.summary || ''
      log(summary, success ? 'success' : 'error')
      result.status  = success ? 'success' : 'failed'
      result.summary = summary
      try { await RpaCloseConnection(rpaReq) } catch (_e) { /* ignore */ }
      return
    }

    // ask_user → 暂停等待用户输入
    if (toolName === 'ask_user') {
      const question = toolArgs.question || '请确认操作'
      log(`等待用户回复: ${question}`, 'warn')
      if (selectedContainers.value.length === 1) {
        addMsg('assistant', `❓ ${question}`)
        waitingForUser.value = true
      }
      const userReply = await waitForUserReply(300000)
      waitingForUser.value = false
      const reply = userReply === null ? '用户未回复，任务超时' : userReply
      if (userReply !== null) log(`用户回复: ${userReply}`)
      messages.push({ role: 'assistant', content: `我执行了 ${toolName}，参数：${JSON.stringify(toolArgs)}` })
      messages.push({ role: 'user',      content: `${toolName} 执行结果：${reply}` })
      noToolRetryCount = 0
      continue
    }

    // 执行工具
    let toolResult
    try {
      toolResult = await executeToolCall(toolName, toolArgs, rpaReq, container)
    } catch (e) {
      toolResult = { success: false, message: e.message }
    }

    // 精简工具结果（防止 XML 树撑爆 token）
    let resultSummary
    if (toolResult.success !== false) {
      // get_screen_context: 直接用 content 字段（已是精简节点摘要），完整传给模型
      if (toolName === 'get_screen_context') {
        resultSummary = toolResult.content || JSON.stringify(toolResult)
      // open_app: 直接用 nodes 字段（启动后自动抓取的界面节点）
      } else if (toolName === 'open_app') {
        resultSummary = toolResult.nodes || '界面节点暂未就绪，请调用get_screen_context重试'
      } else {
        const data = toolResult.output ?? toolResult.data ?? toolResult
        const raw = typeof data === 'string' ? data : JSON.stringify(data)
        resultSummary = raw.slice(0, 200)
      }
    } else {
      const errMsg = (toolResult.message || '未知错误').slice(0, 100)
      resultSummary = `失败: ${errMsg}。请继续尝试其他方式。`
    }

    // swipe 成功时注入计数信息，让模型知道进度
    if (toolName === 'swipe' && toolResult.success !== false) {
      swipeCount++
      if (swipeTarget > 0) {
        resultSummary += `【已滑动第 ${swipeCount} 条，目标 ${swipeTarget} 条${swipeCount >= swipeTarget ? '，已完成，请调用task_done' : '，继续滑动'}】`
      } else {
        resultSummary += `【已滑动第 ${swipeCount} 条】`
      }
    }

    log(`工具结果[${toolName}]: ${resultSummary.slice(0, 120)}`, toolResult.success !== false ? 'success' : 'warn')

    // 追加本轮工具调用+结果到历史，进入下一轮 LLM 决策
    messages.push({ role: 'assistant', content: `我执行了 ${toolName}，参数：${JSON.stringify(toolArgs)}` })
    messages.push({ role: 'user',      content: `${toolName} 执行结果：${resultSummary}` })
    noToolRetryCount = 0
  }
}

// ============================================================
// 消息历史裁剪
// 历史格式：system + [user, assistant, user, assistant, ...]
// assistant 后面跟 user（工具结果），成对裁剪保持上下文一致性
// ============================================================
const trimMessages = (messages, maxEstimatedTokens = 1800) => {
  if (!messages || messages.length === 0) return messages
  const systemMsgs = messages.filter(m => m.role === 'system')
  const nonSystem  = messages.filter(m => m.role !== 'system')
  const estimateTokens = (msgs) => Math.ceil(JSON.stringify(msgs).length / 2)
  const systemTokens = estimateTokens(systemMsgs)
  let budget = Math.max(100, maxEstimatedTokens - systemTokens)

  if (nonSystem.length === 0) return systemMsgs

  // 从末尾往前，assistant+user 成对保留（工具调用轮次）
  const kept = []
  let i = nonSystem.length - 1
  while (i >= 0) {
    // 尝试取 user（工具结果）+ 前面的 assistant（工具调用） 作为一对
    if (i >= 1 && nonSystem[i].role === 'user' && nonSystem[i-1].role === 'assistant') {
      const group = [nonSystem[i-1], nonSystem[i]]
      const cost = Math.ceil(JSON.stringify(group).length / 2)
      if (budget - cost < 50) break
      budget -= cost
      kept.unshift(...group)
      i -= 2
    } else {
      const cost = Math.ceil(JSON.stringify(nonSystem[i]).length / 2)
      if (budget - cost < 50) break
      budget -= cost
      kept.unshift(nonSystem[i])
      i--
    }
  }

  // 确保首条消息是 user（任务描述）
  if (kept.length === 0 || kept[0].role !== 'user') {
    const firstUser = nonSystem.find(m => m.role === 'user')
    if (firstUser && !kept.includes(firstUser)) kept.unshift(firstUser)
  }

  const finalResult = [...systemMsgs, ...kept]
  if (finalResult.length < messages.length) {
    console.log(`[trimMessages] 裁剪: ${messages.length} → ${finalResult.length} 条消息`)
  }
  return finalResult
}

// ============================================================
// 调用 LLM（SSE 流式）
// ============================================================
const callLLM = (messages, tools, deviceIp, llmModel, toolChoice = 'auto') => {
  return new Promise((resolve, reject) => {
    const sessionId = `rpa-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`
    let fullText = ''
    const toolCallMap = {}

    const chunkEvent = `ai:chunk:${sessionId}`
    const doneEvent  = `ai:done:${sessionId}`
    const errorEvent = `ai:error:${sessionId}`

    const cleanup = () => {
      Events.Off(chunkEvent)
      Events.Off(doneEvent)
      Events.Off(errorEvent)
    }

    Events.On(chunkEvent, (event) => {
      const data = event?.data
      if (data?.content) {
        fullText += data.content
        if (selectedContainers.value.length === 1) streamingText.value += data.content
      }
      if (data?.tool_calls) {
        for (const delta of data.tool_calls) {
          const idx = delta.index ?? 0
          if (!toolCallMap[idx]) toolCallMap[idx] = { id: delta.id || `call_${idx}`, name: '', argsStr: '' }
          if (delta.function?.name) toolCallMap[idx].name += delta.function.name
          if (delta.function?.arguments) toolCallMap[idx].argsStr += delta.function.arguments
          if (delta.id) toolCallMap[idx].id = delta.id
        }
      }
    })

    Events.On(doneEvent, (event) => {
      cleanup()
      streamingText.value = ''
      const data = event?.data
      let toolCalls = []
      if (data?.tool_calls && data.tool_calls.length > 0) {
        toolCalls = data.tool_calls.map(tc => ({
          id:   tc.id   || `call_${Math.random().toString(36).slice(2,6)}`,
          name: tc.name || '',
          args: tc.args || {}
        })).filter(tc => tc.name)
      } else {
        toolCalls = Object.values(toolCallMap).map(tc => {
          let args = {}
          try { args = JSON.parse(tc.argsStr || '{}') } catch (_e) { args = {} }
          return { id: tc.id, name: tc.name, args }
        }).filter(tc => tc.name)
      }
      resolve({ text: fullText, toolCalls })
    })

    Events.On(errorEvent, (event) => {
      cleanup()
      streamingText.value = ''
      reject(new Error(event?.data?.message || 'LLM 请求失败'))
    })

    const CTX_SIZE      = 2048  // 与 AIStartModel ctx-size 保持一致
    const OUTPUT_RESERVE = 256
    const SAFETY_MARGIN  = 50
    const toolsTokens = Math.ceil(JSON.stringify(tools || []).length / 2)
    let effectiveTools = tools || []
    const toolsBudget = Math.floor(CTX_SIZE * 0.5)
    if (toolsTokens > toolsBudget) {
      const sorted = [...effectiveTools].sort((a, b) => JSON.stringify(a).length - JSON.stringify(b).length)
      let budget = toolsBudget
      effectiveTools = []
      for (const t of sorted) {
        const cost = Math.ceil(JSON.stringify(t).length / 2)
        if (budget - cost < 0) break
        budget -= cost
        effectiveTools.push(t)
      }
      console.warn(`[callLLM] tools 超限，裁剪: ${tools.length} → ${effectiveTools.length}`)
    }
    const effectiveToolsTokens = Math.ceil(JSON.stringify(effectiveTools).length / 2)
    const msgBudget = Math.max(300, CTX_SIZE - effectiveToolsTokens - OUTPUT_RESERVE - SAFETY_MARGIN)
    const trimmedMessages = trimMessages(messages, msgBudget)
    console.log(`[callLLM] device=${deviceIp} ctx=${CTX_SIZE} toolsTokens=${effectiveToolsTokens} msgBudget=${msgBudget} tools=${effectiveTools.length}/${(tools||[]).length}`)

    AISendMessage({
      deviceIp:    deviceIp,
      model:       llmModel,
      messages:    trimmedMessages,
      temperature: 0.1,
      maxTokens:   OUTPUT_RESERVE,
      sessionId:   sessionId,
      tools:       effectiveTools,
      toolChoice:  toolChoice
    }).catch(e => {
      cleanup()
      streamingText.value = ''
      reject(e)
    })
  })
}

// ============================================================
// 执行工具调用
// exec_cmd 特殊处理：注入 device_ip 和 container_name，走 Go 后端 HTTP exec
// ============================================================
const executeToolCall = async (toolName, args, rpaReq, container) => {
  // exec_cmd → 通过 Go 后端 execAndroidCmdForRpa（HTTP :8000/android/exec）
  if (toolName === 'exec_cmd') {
    const enrichedArgs = {
      ...args,
      device_ip:      container.deviceIP,
      container_name: container.name,
    }
    console.log(`[executeToolCall] exec_cmd device=${container.deviceIP} container=${container.name} cmd=${args.cmd}`)
    return await RpaCallFixedTool('exec_cmd', enrichedArgs)
  }

  // 层2：RPA 界面操作工具
  switch (toolName) {
    case 'get_screen_context':
      return await RpaGetScreenContext(rpaReq)

    case 'click':
      return await RpaClick(rpaReq, args.x ?? 0, args.y ?? 0)

    case 'swipe':
      return await RpaSwipe(rpaReq, args.x0 ?? 540, args.y0 ?? 1200, args.x1 ?? 540, args.y1 ?? 400, args.duration_ms ?? 300)

    case 'send_text':
      return await RpaSendText(rpaReq, args.text ?? '')

    case 'key_press':
      return await RpaKeyPress(rpaReq, args.key_code ?? 4)

    case 'open_app':
      return await RpaOpenApp(rpaReq, args.pkg ?? '')

    case 'stop_app':
      return await RpaStopApp(rpaReq, args.pkg ?? '')

    default:
      return { success: false, message: `未知工具: ${toolName}` }
  }
}

// ============================================================
// ask_user 等待用户回复
// ============================================================
let userReplyResolver = null

const waitForUserReply = (timeoutMs) => {
  return new Promise((resolve) => {
    const timer = setTimeout(() => {
      userReplyResolver = null
      resolve(null)
    }, timeoutMs)
    userReplyResolver = (reply) => {
      clearTimeout(timer)
      resolve(reply)
    }
  })
}

const replyToAgent = () => {
  if (!waitingForUser.value || !inputMessage.value.trim()) return
  const reply = inputMessage.value.trim()
  inputMessage.value = ''
  waitingForUser.value = false
  addMsg('user', reply)
  if (userReplyResolver) { userReplyResolver(reply); userReplyResolver = null }
}

const handleEnter = (e) => {
  if (e.shiftKey) return
  if (waitingForUser.value) replyToAgent()
  else if (canStart.value && !isRunning.value) startTask()
}

// ============================================================
// 停止所有任务
// ============================================================
const stopAllTasks = () => {
  for (const ct of selectedContainers.value) {
    abortFlags.value[ct.key] = true
  }
  isRunning.value = false
  waitingForUser.value = false
  if (userReplyResolver) { userReplyResolver(null); userReplyResolver = null }
}

// ============================================================
// 工具函数
// ============================================================
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms))

// ============================================================
// 生命周期
// ============================================================
onMounted(async () => {
  loadSettings()
  // 先拉取在线设备的 r1q 信息，再加载容器列表
  const onlineDevices = props.devices.filter(d => props.devicesStatusCache.get(d.id) === 'online')
  if (onlineDevices.length > 0) {
    await Promise.allSettled(onlineDevices.map(d => fetchDeviceInfo(d.ip)))
  }
  loadContainers(false)
})

const fetchRpaAgent = async () => {
  loadSettings()
  const onlineDevices = props.devices.filter(d => props.devicesStatusCache.get(d.id) === 'online')
  if (onlineDevices.length > 0) {
    await Promise.allSettled(onlineDevices.map(d => fetchDeviceInfo(d.ip)))
  }
  await loadContainers(false)
}
defineExpose({ fetchRpaAgent })
</script>

<style scoped>
.rpa-agent {
  display: flex;
  height: 100%;
  gap: 12px;
  padding: 12px;
  background: #f5f7fa;
  overflow: hidden;
}

/* ---- 左栏 ---- */
.left-panel {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow-y: auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  font-weight: 600;
}

/* 容器列表 */
.container-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 340px;
  overflow-y: auto;
}

.device-group {
  margin-bottom: 4px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 6px;
  border-radius: 5px;
  cursor: pointer;
  background: #f0f2f5;
  user-select: none;
  font-size: 12px;
  font-weight: 600;
  color: #303133;
}

.group-header:hover {
  background: #e6e8eb;
}

.group-ip {
  flex: 1;
  font-family: monospace;
}

.group-empty {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  font-size: 11px;
  color: #e6a23c;
}

.container-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 10px;
  border-radius: 5px;
  cursor: pointer;
  transition: background 0.15s;
  margin-left: 8px;
}

.container-item:hover, .container-item.active {
  background: #ecf5ff;
}

.ct-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.ct-name {
  font-size: 12px;
  font-weight: 500;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ct-meta {
  display: flex;
  align-items: center;
  gap: 4px;
}

.ct-port {
  font-size: 10px;
  color: #909399;
  font-family: monospace;
}

.device-actions {
  display: flex;
  gap: 6px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #ebeef5;
}

.device-card, .progress-card {
  border-radius: 8px;
}

.progress-card {
  flex: 1;
  overflow-y: auto;
}

.progress-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.progress-ip {
  flex: 1;
  font-size: 12px;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.progress-log {
  padding: 6px 0;
  font-size: 11px;
  max-height: 200px;
  overflow-y: auto;
}

.log-line {
  display: flex;
  gap: 6px;
  padding: 2px 0;
  line-height: 1.5;
}

.log-time { color: #909399; flex-shrink: 0; }
.log-text { color: #606266; word-break: break-all; }
.log-line.success .log-text { color: #67c23a; }
.log-line.error .log-text { color: #f56c6c; }
.log-line.warn .log-text { color: #e6a23c; }

.log-summary { margin-top: 6px; }

.rotating { animation: rotating 1s linear infinite; }
@keyframes rotating {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}

/* ---- 右栏 ---- */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow: hidden;
}

.task-cards { flex-shrink: 0; }

.task-cards-title {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
}

.task-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 6px;
  border-radius: 8px;
  border: 1.5px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.15s;
  background: #fff;
  font-size: 12px;
}

.task-card:hover:not(.disabled) { border-color: #409eff; background: #ecf5ff; }
.task-card.active { border-color: #409eff; background: #ecf5ff; box-shadow: 0 0 0 2px rgba(64,158,255,0.2); }
.task-card.disabled { opacity: 0.6; cursor: not-allowed; pointer-events: none; }
.task-card-icon { font-size: 20px; }
.task-card-name { font-size: 11px; color: #303133; }

.chat-area {
  flex: 1;
  overflow-y: auto;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.chat-msg { display: flex; }
.chat-msg.user { justify-content: flex-end; }
.chat-msg.assistant, .chat-msg.tool { justify-content: flex-start; }

.msg-bubble {
  max-width: 80%;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 13px;
  line-height: 1.6;
}

.chat-msg.user .msg-bubble { background: #409eff; color: #fff; }
.chat-msg.assistant .msg-bubble { background: #f5f7fa; border: 1px solid #e4e7ed; color: #303133; }
.chat-msg.tool .msg-bubble { background: #fdf6ec; border: 1px solid #faecd8; color: #606266; font-size: 12px; }

.msg-role { font-size: 11px; color: #909399; margin-bottom: 4px; }
.chat-msg.user .msg-role { color: rgba(255,255,255,0.7); }

.msg-content { word-break: break-word; }
.msg-content :deep(p) { margin: 0; }
.msg-content :deep(pre) { background: #2d3748; color: #e2e8f0; padding: 8px; border-radius: 4px; overflow-x: auto; font-size: 12px; }
.msg-content :deep(code) { background: rgba(0,0,0,0.1); padding: 1px 4px; border-radius: 3px; font-size: 12px; }

.msg-content.streaming { white-space: pre-wrap; }

.cursor { animation: blink 1s infinite; }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0; } }

.tool-calls { margin-top: 6px; display: flex; flex-direction: column; gap: 3px; }
.tool-call-item { display: flex; align-items: center; gap: 6px; }
.tool-args { font-size: 11px; color: #909399; word-break: break-all; }

.tool-result { margin-top: 4px; display: flex; align-items: center; gap: 6px; }
.tool-result-text { font-size: 11px; color: #606266; word-break: break-all; }

.input-area {
  flex-shrink: 0;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  padding: 10px;
}

.waiting-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #e6a23c;
  margin-bottom: 6px;
}

.input-row { display: flex; gap: 8px; align-items: flex-end; }
.input-row .el-textarea { flex: 1; }
.input-btns { display: flex; flex-direction: column; gap: 6px; flex-shrink: 0; }
.input-tips { font-size: 11px; color: #909399; margin-top: 6px; }
</style>
