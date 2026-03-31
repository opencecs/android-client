<template>
  <div class="ai-assistant">
    <!-- 左侧设备列表 -->
    <div class="device-list-section">
      <el-card class="device-card">
        <template #header>
          <div class="card-header">
            <span>{{ $t('common.onlineDevicesLabel') }}</span>
          </div>
        </template>
        
        <div class="device-list">
          <div 
            v-for="device in onlineDevices" 
            :key="device.id"
            :class="['device-item', { 'active': selectedDevice?.id === device.id }]"
            @click="selectDevice(device)"
          >
            <div class="device-info">
              <div class="device-ip">{{ device.ip }}</div>
              <!-- <div class="device-status">
                <el-tag size="small" type="success">AI</el-tag>
              </div> -->
            </div>
          </div>
          
          <el-empty 
            v-if="onlineDevices.length === 0" 
            :description="$t('common.noAIDevice')"
            :image-size="60"
          />
        </div>
      </el-card>
    </div>
    
    <!-- 中间模型选择区域 -->
    <div class="model-section">
      <el-card class="model-card">
        <template #header>
          <div class="card-header">
            <span>{{ $t('common.modelManagement') }}</span>
            <el-button 
              type="primary" 
              size="small"
              @click="handleImportModel"
              :loading="importing"
              :disabled="!selectedDevice"
            >
              {{ $t('common.importModel') }}
            </el-button>
          </div>
        </template>
        
        <div class="model-content">
          <div v-loading="loadingModels" class="model-selector-container">
            <div v-if="modelList.length === 0 && !loadingModels" class="empty-hint">
              <el-empty :description="$t('common.noModel')" :image-size="80" />
            </div>
            
            <div v-else class="model-selector">
              <el-form label-position="top">
                <el-form-item :label="$t('common.chatModel')">
                  <el-select 
                    v-model="selectedModelName" 
                    :placeholder="$t('common.selectChatModel')" 
                    style="width: 100%"
                    @change="handleModelChange"
                    size="large"
                  >
                    <el-option
                      v-for="model in modelList"
                      :key="model.name"
                      :label="model.name"
                      :value="model.name"
                    >
                      <div class="model-option">
                        <span class="model-option-name">{{ model.name }}</span>
                        <div class="model-option-actions">
                          <el-tag 
                            v-if="loadedModel?.name === model.name" 
                            size="small" 
                            type="success"
                            style="margin-right: 8px"
                          >
                            {{ $t('common.running') }}
                          </el-tag>
                          <el-button
                            type="danger"
                            size="small"
                            :icon="Delete"
                            circle
                            @click.stop="handleDeleteModel(model.name)"
                            :title="$t('common.delete')"
                          />
                        </div>
                      </div>
                    </el-option>
                  </el-select>
                </el-form-item>

                <!-- 启动/停止按钮 -->
                <el-form-item>
                  <div style="display: flex; gap: 8px; width: 100%;">
                    <el-button
                      type="success"
                      size="large"
                      style="flex: 1;"
                      @click="handleLoadModel"
                      :loading="loadingModel"
                      :disabled="!selectedModelName || !selectedDevice"
                    >
                      {{ loadingModel ? $t('common.startingModel') : $t('common.startModel') }}
                    </el-button>
                    <el-button
                      type="warning"
                      size="large"
                      @click="handleStopLLMService"
                      :loading="stoppingService"
                      :disabled="!selectedDevice"
                      :title="$t('common.stopLLMService')"
                    >
                      {{ $t('common.stopModel') }}
                    </el-button>
                  </div>
                </el-form-item>

                <!-- 运行状态 -->
                <div v-if="loadedModel" style="display: flex; flex-wrap: wrap; gap: 6px; margin-top: 4px;">
                  <el-tag v-if="loadedModel" type="success" size="small">
                    💬 {{ loadedModel.name }}
                  </el-tag>
                </div>
              </el-form>
            </div>
          </div>
        </div>
      </el-card>
    </div>
    
    <!-- 右侧对话区域 -->
    <div class="chat-section">
      <el-card class="chat-card">
        <template #header>
          <div class="card-header">
            <span>
              {{ $t('common.aiChat') }}
              <el-tag v-if="loadedModel" size="small" type="success" style="margin-left: 8px">
                {{ loadedModel.name }}
              </el-tag>
            </span>
            <div class="header-actions">
              <el-button 
                v-if="sending"
                type="warning" 
                size="small"
                @click="stopGeneration"
              >
                {{ $t('common.stopGeneration') }}
              </el-button>
              <el-button 
                v-if="canContinueGenerate"
                type="success" 
                size="small"
                @click="continueGeneration"
                :icon="VideoPlay"
              >
                {{ $t('common.continueGenerate') }}
              </el-button>
              <el-button 
                type="primary" 
                size="small"
                @click="startNewChat"
                :disabled="chatMessages.length === 0"
                :icon="ChatLineRound"
              >
                {{ $t('common.newChat') }}
              </el-button>
              <el-button 
                type="primary" 
                size="small"
                @click="showSettings = true"
                :icon="Setting"
              >
                {{ $t('common.settings') }}
              </el-button>
              <el-button 
                type="danger" 
                size="small"
                @click="clearChat"
                :disabled="chatMessages.length === 0"
              >
                {{ $t('common.clearChat') }}
              </el-button>
            </div>
          </div>
        </template>
        
        <div class="chat-content">
          <!-- 聊天消息区域 -->
          <div class="chat-messages" ref="chatMessagesRef" @click="handleCopyCode">
            <div v-if="chatMessages.length === 0 && !loadedModel" class="empty-chat">
              <el-empty :description="$t('common.selectDeviceAndLoadModel')" :image-size="100" />
            </div>
            <div v-else-if="chatMessages.length === 0" class="empty-chat">
              <div class="welcome-container">
                <el-icon :size="80" color="#409eff"><ChatDotRound /></el-icon>
                <h2>{{ $t('common.helloAssistant') }}</h2>
                <p>{{ $t('common.howCanIHelp') }}</p>
              </div>
            </div>
            <div v-else class="messages-list">
              <div v-for="(msg, index) in chatMessages" :key="index" class="message-item-new">
                <!-- 用户消息 -->
                <div v-if="msg.role === 'user'" class="user-msg-new">
                  <div class="user-avatar-new">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2">
                      <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
                      <circle cx="12" cy="7" r="4" />
                    </svg>
                  </div>
                  <div class="user-content-new">
                    <!-- 图片 -->
                    <div v-if="msg.files && msg.files.length > 0" class="msg-images-new">
                      <img 
                        v-for="(file, imgIndex) in msg.files" 
                        :key="imgIndex" 
                        :src="file.data || file.preview" 
                        :alt="file.name"
                        class="msg-image-new"
                        @click="previewImage(file.data || file.preview)"
                      />
                    </div>
                    <!-- 文字 -->
                    <div v-if="msg.content">{{ msg.content }}</div>
                  </div>
                </div>
                
                <!-- AI 消息 -->
                <div v-else class="ai-msg-new">
                  <div class="ai-avatar-new">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                      stroke="currentColor" stroke-width="2">
                      <path d="M12 8V4H8" />
                      <rect width="16" height="12" x="4" y="8" rx="2" />
                      <path d="M2 14h2" />
                      <path d="M20 14h2" />
                      <path d="M15 13v2" />
                      <path d="M9 13v2" />
                    </svg>
                  </div>
                  
                  <div class="ai-content-wrap-new">
                    <!-- 思考过程 -->
                    <div v-if="msg.reasoning || msg.isThinking" class="think-block-new">
                      <div @click="msg.thinkingExpanded = !msg.thinkingExpanded" class="think-header-new">
                        <span class="think-icon-new" :class="{ 'expanded': msg.thinkingExpanded }">▶</span>
                        <span v-if="msg.isThinking">{{ $t('common.thinkingInProgress') }}</span>
                        <span v-else>{{ $t('common.thinkingProcess') }}</span>
                        <span v-if="msg.thinkingDuration" class="think-time-new">({{ msg.thinkingDuration }}s)</span>
                      </div>
                      <div v-show="msg.thinkingExpanded" class="think-content-new" v-html="formatMessageContent(msg.reasoning, false)"></div>
                    </div>
                    
                    <!-- 正文内容 -->
                    <div 
                      class="ai-main-content-new" 
                      :class="{ 'typing': msg.streaming && !msg.isThinking }"
                      v-html="formatMessageContent(msg.content, msg.streaming)"
                    ></div>
                    
                    <!-- 性能指标 -->
                    <div v-if="!msg.streaming && msg.stats && msg.role === 'assistant'" class="metrics-bar-new">
                      <span>⏱️ {{ msg.stats.duration }}s</span>
                      <span>⚡ {{ msg.stats.tokensPerSecond }} tokens/s</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 经验库反馈条 -->
          <transition name="feedback-slide">
            <div v-if="feedbackBar.visible" class="feedback-bar">
              <div class="feedback-bar-inner">
                <div class="feedback-left">
                  <span class="feedback-icon">💡</span>
                  <span class="feedback-text">{{ $t('common.feedbackQuestion') }}</span>
                  <span class="feedback-countdown">{{ $t('common.closeInSeconds', { seconds: feedbackBar.countdown }) }}</span>
                </div>
                <div class="feedback-actions">
                  <button class="feedback-btn useful" @click="submitFeedback(true)" :disabled="feedbackBar.saving">
                    {{ $t('common.feedbackUseful') }}
                  </button>
                  <button class="feedback-btn useless" @click="submitFeedback(false)" :disabled="feedbackBar.saving">
                    {{ $t('common.feedbackUseless') }}
                  </button>
                  <button class="feedback-btn dismiss" @click="dismissFeedbackBar()">
                    ✕
                  </button>
                </div>
              </div>
              <!-- 倒计时进度条 -->
              <div class="feedback-progress">
                <div class="feedback-progress-bar" :style="{ width: (feedbackBar.countdown / 10 * 100) + '%' }"></div>
              </div>
            </div>
          </transition>

          <!-- 输入区域 -->
          <div class="chat-input-area">
            <!-- 上传的文件预览 -->
            <div v-if="uploadedFiles.length > 0" class="uploaded-files-preview">
              <div 
                v-for="(file, index) in uploadedFiles" 
                :key="index"
                class="file-preview-item"
              >
                <!-- 图片预览 -->
                <div v-if="file.type.startsWith('image/')" class="image-preview">
                  <img :src="file.preview" :alt="file.name" />
                  <div class="file-overlay">
                    <span class="file-name">{{ file.name }}</span>
                    <el-button 
                      type="danger" 
                      size="small" 
                      circle
                      @click="removeFile(index)"
                      class="remove-btn"
                    >
                      ×
                    </el-button>
                  </div>
                </div>
                <!-- 视频预览 -->
                <div v-else-if="file.type.startsWith('video/')" class="video-preview">
                  <video :src="file.preview" controls></video>
                  <div class="file-overlay">
                    <span class="file-name">{{ file.name }}</span>
                    <el-button 
                      type="danger" 
                      size="small" 
                      circle
                      @click="removeFile(index)"
                      class="remove-btn"
                    >
                      ×
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
            
            <div :class="['input-wrapper', { 'disabled': !canChat || sending }]">
              <!-- 工具栏 -->
              <div class="editor-toolbar">
                <!-- 上传按钮 -->
                <div class="toolbar-group">
                  <input
                    ref="fileInputRef"
                    type="file"
                    accept="image/*,video/*"
                    multiple
                    style="display: none"
                    @change="handleFileSelect"
                  />
                  <el-tooltip :content="$t('common.uploadMediaTooltip')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="triggerFileUpload"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <el-icon :size="18">
                        <Upload />
                      </el-icon>
                    </button>
                  </el-tooltip>
                </div>
                
                <!-- 格式化按钮 -->
                <!-- <div class="toolbar-group">
                  <el-tooltip :content="$t('aiAssistant.bold')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="formatText('bold')"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <span class="toolbar-icon">B</span>
                    </button>
                  </el-tooltip>
                  
                  <el-tooltip :content="$t('aiAssistant.italic')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="formatText('italic')"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <span class="toolbar-icon italic">I</span>
                    </button>
                  </el-tooltip>
                  
                  <el-tooltip :content="$t('aiAssistant.underline')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="formatText('underline')"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <span class="toolbar-icon underline">U</span>
                    </button>
                  </el-tooltip>
                  
                  <el-tooltip :content="$t('aiAssistant.strikethrough')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="formatText('strikeThrough')"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <span class="toolbar-icon strikethrough">S</span>
                    </button>
                  </el-tooltip>
                </div> -->
                
                <!-- <div class="toolbar-divider"></div> -->
                
                <!-- 列表按钮 -->
                <!-- <div class="toolbar-group">
                  <el-tooltip :content="$t('aiAssistant.unorderedList')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="formatText('insertUnorderedList')"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <el-icon :size="18">
                        <List />
                      </el-icon>
                    </button>
                  </el-tooltip>
                  
                  <el-tooltip :content="$t('aiAssistant.orderedList')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="formatText('insertOrderedList')"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <el-icon :size="18">
                        <Tickets />
                      </el-icon>
                    </button>
                  </el-tooltip>
                </div> -->
                
                <!-- <div class="toolbar-divider"></div> -->
                
                <!-- 清除格式按钮 -->
                <!-- <div class="toolbar-group">
                  <el-tooltip :content="$t('aiAssistant.clearFormat')" placement="top">
                    <button
                      class="toolbar-btn"
                      @click="clearFormat"
                      :disabled="!canChat || sending"
                      type="button"
                    >
                      <el-icon :size="18">
                        <RefreshLeft />
                      </el-icon>
                    </button>
                  </el-tooltip>
                </div> -->
              </div>
              
              <!-- 富文本编辑器 -->
              <div
                ref="editorRef"
                class="rich-editor"
                contenteditable="true"
                :data-placeholder="editorPlaceholder"
                @input="handleEditorInput"
                @keydown="handleEditorKeydown"
                @paste="handleEditorPaste"
                @focus="handleEditorFocus"
                @blur="handleEditorBlur"
                @dragover="handleDragOver"
                @dragleave="handleDragLeave"
                @drop="handleDrop"
              ></div>
              
              <!-- 发送按钮 -->
              <div class="input-actions">
                <el-button 
                  type="primary" 
                  @click="sendMessage"
                  :loading="sending"
                  :disabled="!canChat || (!inputMessage.trim() && uploadedFiles.length === 0)"
                  circle
                  :icon="Promotion"
                  size="large"
                />
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>
    
    <!-- 设置对话框 -->
    <el-dialog
      v-model="showSettings"
      :title="$t('common.settings')"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-tabs v-model="activeSettingTab" type="border-card">
        <!-- 服务管理 -->
        <el-tab-pane :label="$t('common.serviceManagement')" name="service">
          <div class="settings-section">
            <el-alert
              :title="$t('common.stopServiceWarning')"
              type="warning"
              :closable="false"
              style="margin-bottom: 20px;"
            />
            
            <el-form label-width="120px">
              <el-form-item :label="$t('common.currentDevice')">
                <el-tag v-if="selectedDevice" type="success">{{ selectedDevice.ip }}</el-tag>
                <span v-else style="color: #999;">{{ $t('common.noDeviceSelected') }}</span>
              </el-form-item>
              
              <el-form-item :label="$t('common.currentModel')">
                <el-tag v-if="loadedModel" type="success">{{ loadedModel.name }}</el-tag>
                <span v-else style="color: #999;">{{ $t('common.noModelLoaded') }}</span>
              </el-form-item>
              
              <el-form-item>
                <el-button
                  type="danger"
                  @click="handleStopLLMService"
                  :loading="stoppingService"
                  :disabled="!selectedDevice || !loadedModel"
                >
                  {{ stoppingService ? $t('common.stoppingService') : $t('common.stopLLMService') }}
                </el-button>
              </el-form-item>
            </el-form>
            
            <el-divider content-position="left">{{ $t('common.deviceManagement') }}</el-divider>
            
            <el-alert
              :title="$t('common.resetDeviceWarning')"
              type="warning"
              :closable="false"
              style="margin-bottom: 20px;"
            />
            
            <el-form label-width="120px">
              <el-form-item :label="$t('common.resetOperation')">
                <el-button
                  type="danger"
                  @click="handleResetDevice"
                  :loading="resettingDevice"
                  :disabled="!selectedDevice"
                  :icon="RefreshRight"
                  plain
                >
                  {{ resettingDevice ? $t('common.resettingDevice') : $t('common.resetDevice') }}
                </el-button>
                <div style="margin-top: 8px; font-size: 12px; color: #909399;">
                  ⚠️ {{ $t('common.resetConfirmMessage') }}
                </div>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
        
        <!-- 系统提示词 -->
        <el-tab-pane :label="$t('common.systemPrompt')" name="prompt">
          <div class="settings-section">
            <el-alert
              :title="$t('common.systemPromptDesc')"
              type="info"
              :closable="false"
              style="margin-bottom: 20px;"
            />
            
            <el-form label-width="120px">
              <el-form-item :label="$t('common.enablePrompt')">
                <el-switch v-model="systemPromptEnabled" />
              </el-form-item>
              
              <el-form-item :label="$t('common.promptContent')">
                <el-input
                  v-model="systemPrompt"
                  type="textarea"
                  :rows="6"
                  :disabled="!systemPromptEnabled"
                />
              </el-form-item>
              
              <el-form-item>
                <el-button
                  type="primary"
                  @click="saveSystemPrompt"
                >
                  {{ $t('common.saveSettings') }}
                </el-button>
                <el-button @click="resetSystemPrompt">
                  {{ $t('common.restoreDefault') }}
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup>
import { getCurrentInstance } from 'vue'
const { proxy } = getCurrentInstance()
const t = proxy.$t

// 返回设备的 host:port，若 ip 已含端口则直接使用，否则追加默认 8000
const getDeviceAddr = (ip) => {
  if (!ip) return ip
  const lastColon = ip.lastIndexOf(':')
  if (lastColon === -1) return ip + ':8000'
  return /^\d+$/.test(ip.slice(lastColon + 1)) ? ip : ip + ':8000'
}


import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User, ChatDotRound, Promotion, Delete, Upload, Setting, List, Tickets, RefreshLeft, RefreshRight, VideoPlay, ChatLineRound, QuestionFilled } from '@element-plus/icons-vue'
import { UploadLLMModel, SelectZipFile, GetLLMModelList, AISendMessage, AICancelMessage, AIStartModel, AIStopModel, AIGetModels, AIResetDevice, AIGetDeviceInfo } from '../../bindings/edgeclient/app'
import { Events } from '@wailsio/runtime'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

// 配置 Markdown-it 渲染器（参考 aiAssistantNew.vue）
const md = new MarkdownIt({
  html: true,
  linkify: true,
  highlight: function (str, lang) {
    const code = md.utils.escapeHtml(str)
    const language = lang && hljs.getLanguage(lang) ? lang : 'text'
    let highlighted = code

    if (lang && hljs.getLanguage(lang)) {
      try {
        highlighted = hljs.highlight(str, { language: lang, ignoreIllegals: true }).value
      } catch (__) { }
    }

    // 构建代码块 HTML（使用新样式类名）
    return `<div class="code-block-wrapper">
      <div class="code-block-header">
        <span class="code-lang">${language}</span>
        <button class="copy-btn" data-code="${encodeURIComponent(str)}">${t('aiAssistant.copy')}</button>
      </div>
      <pre class="code-block"><code class="hljs language-${language}">${highlighted}</code></pre>
    </div>`
  }
})

// Props
const props = defineProps({
  devices: {
    type: Array,
    default: () => []
  },
  token: {
    type: String,
    default: ''
  },
  devicesStatusCache: {
    type: Map,
    default: () => new Map()
  }
})

// 认证头辅助函数
const getAuthHeaders = (deviceIP) => {
  const savedPassword = localStorage.getItem('devicePasswords')
  const passwords = JSON.parse(savedPassword || '{}')
  const password = passwords[deviceIP] || null
  if (password) {
    const auth = btoa(`admin:${password}`)
    return {
      'Authorization': `Basic ${auth}`
    }
  }
  return {}
}

// 获取设备认证 token
const getDeviceToken = (deviceIP) => {
  const savedPassword = localStorage.getItem('devicePasswords')
  const passwords = JSON.parse(savedPassword || '{}')
  const password = passwords[deviceIP] || null
  if (password) {
    const auth = btoa(`admin:${password}`)
    return `Basic ${auth}`
  }
  return ''
}

// 本地OpenAI配置
const localOpenAIConfig = ref({
  model: 'qwen2.5-3b',
  temperature: 0.7,
  maxTokens: 2000
})

// 响应式数据
const selectedDevice = ref(null)
const loading = ref(false)
const importing = ref(false)
const modelList = ref([])
const selectedModel = ref(null)
const selectedModelName = ref('') // 对话模型名称
const loadedModel = ref(null)       // 已加载的对话模型
const loadingModels = ref(false)
const loadingModel = ref(false)
const deviceModelsCache = ref(new Map()) // 缓存设备的 model 信息
const startingService = ref(false) // 启动服务加载状态
const stoppingService = ref(false) // 停止服务加载状态
const resettingDevice = ref(false) // 重置设备加载状态

// 对话相关
const chatMessages = ref([])
const inputMessage = ref('')
const sending = ref(false)
const chatMessagesRef = ref(null)
const abortController = ref(null) // 保留兼容（不再使用，由 currentSessionId 取代）
const currentSessionId = ref(null) // 当前 AI 对话 session（Go 后端管理）
const fileInputRef = ref(null) // 文件输入框引用
const uploadedFiles = ref([]) // 已上传的文件列表
const wasStoppedManually = ref(false) // 是否手动停止了生成

// 富文本编辑器相关
const editorRef = ref(null)
const editorPlaceholder = computed(() => t('aiAssistant.editorPlaceholder'))
const editorFocused = ref(false)



// 打字机效果相关（已弃用，改为即时流式输出）
// const chunkBuffer = ref([]) // 内容缓冲队列
// const isTyping = ref(false) // 是否正在打字（内容）
// const typingSpeed = 3 // 基础打字速度（毫秒/批次）
// const charsPerBatch = 3 // 每批次处理的字符数（提高批量处理数量）


// ===== 经验库反馈条 =====
const feedbackBar = ref({
  visible: false,       // 是否显示
  countdown: 10,        // 倒计时秒数
  question: '',         // 用户问题（用于存储）
  answer: '',           // AI 回答
  timer: null,          // 倒计时定时器
  saving: false,        // 正在保存
})

// 设置相关
const showSettings = ref(false)
const activeSettingTab = ref('service')
const systemPromptEnabled = ref(false)
const systemPrompt = ref('')
const defaultSystemPrompt = computed(() => t('aiAssistant.defaultPrompt'))

// 打字机效果已移除，改为即时流式输出以提升用户体验
// 打字机效果已移除，改为即时流式输出以提升用户体验

// 从本地存储加载系统提示词设置
const loadSystemPromptSettings = () => {
  try {
    const savedEnabled = localStorage.getItem('ai_system_prompt_enabled')
    const savedPrompt = localStorage.getItem('ai_system_prompt')
    
    if (savedEnabled !== null) {
      systemPromptEnabled.value = savedEnabled === 'true'
    }
    
    if (savedPrompt) {
      systemPrompt.value = savedPrompt
    } else {
      systemPrompt.value = defaultSystemPrompt.value
    }
  } catch (error) {
    console.error('[loadSystemPromptSettings] 加载设置失败:', error)
  }
}

// 保存系统提示词到本地
const saveSystemPrompt = () => {
  try {
    localStorage.setItem('ai_system_prompt_enabled', systemPromptEnabled.value.toString())
    localStorage.setItem('ai_system_prompt', systemPrompt.value)
    ElMessage.success(t('aiAssistant.settingsSaved'))
  } catch (error) {
    console.error('[saveSystemPrompt] 保存失败:', error)
    ElMessage.error(t('aiAssistant.saveFailed') + error.message)
  }
}

// 恢复默认提示词
const resetSystemPrompt = () => {
  systemPrompt.value = defaultSystemPrompt.value
  ElMessage.info(t('aiAssistant.restoredDefault'))
}

// 初始化时加载设置
loadSystemPromptSettings()


// 获取设备详细信息（包括 model、rk182x 字段）
const fetchDeviceInfo = async (deviceIP) => {
  // 检查缓存
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
        // 缓存完整 data 对象（含 model、rk182x 等字段）
        deviceModelsCache.value.set(deviceIP, data.data)
        return data.data
      }
    }
  } catch (error) {
    console.warn(`[fetchDeviceInfo] 获取设备 ${deviceIP} 信息失败:`, error)
  }
  
  return null
}

// 计算支持 AI 的在线设备（显示 r1q 和 eces-rk3588-rk1828 机型）
const onlineDevices = computed(() => {
  console.log('******原始设备列表:', props.devices)
  
  // 先过滤在线设备（使用 devicesStatusCache 判断）
  const online = props.devices.filter(device => {
    const status = props.devicesStatusCache.get(device.id)
    return status === 'online'
  })
  
  console.log('******在线设备:', online.length)
  
  // 进一步过滤支持 AI 的设备：
  // 条件1：rk182x === 'y'（插了算力棒）且 version 包含 'r1q'（r1q 系列设备）
  // 条件2：model 或 version 包含 'eces-rk3588-rk1828'
  const aiSupportedDevices = online.filter(device => {
    const info = deviceModelsCache.value.get(device.ip)
    if (!info) return false
    const version = (info.version || '').toLowerCase()
    const model = (info.model || '').toLowerCase()
    const isR1q = info.rk182x === 'y' && version.includes('r1q')
    const isEces = model.includes('eces-rk3588-rk1828') || version.includes('eces-rk3588-rk1828')
    return isR1q || isEces
  })
  
  console.log('******支持AI的设备列表:', aiSupportedDevices)
  return aiSupportedDevices
})

// 计算是否可以对话
const canChat = computed(() => {
  return !!loadedModel.value && !!selectedDevice.value
})

// 计算是否可以继续生成
const canContinueGenerate = computed(() => {
  // 满足条件：1) 不在发送中 2) 手动停止过 3) 有对话消息 4) 最后一条是助手消息
  if (sending.value || !wasStoppedManually.value || chatMessages.value.length === 0) {
    return false
  }
  
  const lastMsg = chatMessages.value[chatMessages.value.length - 1]
  return lastMsg.role === 'assistant'
})

// 打字机效果已移除，改为即时流式输出

// ---- 流式渲染性能优化 ----
// 批量 chunk 缓冲区：积累内容后定时 flush，减少 Vue 响应式更新频率
let _chunkBuffer = ''         // 待 flush 的正常内容
let _reasoningBuffer = ''     // 待 flush 的思考内容
let _flushTimer = null        // flush 定时器句柄

// scrollToBottom 节流：流式过程中每 80ms 最多滚动一次
let _scrollThrottleTimer = null
const _scheduleScroll = () => {
  if (_scrollThrottleTimer) return
  _scrollThrottleTimer = setTimeout(() => {
    _scrollThrottleTimer = null
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  }, 80)
}

// 将缓冲区内容一次性写入消息对象
const _flushChunkBuffer = () => {
  _flushTimer = null
  const msgs = chatMessages.value
  if (!msgs.length) return
  const lastMsg = msgs[msgs.length - 1]
  if (lastMsg.role !== 'assistant' || !lastMsg.streaming) return

  if (_chunkBuffer) {
    lastMsg.content += _chunkBuffer
    lastMsg.contentTokens = (lastMsg.contentTokens || 0) + _chunkBuffer.length
    _chunkBuffer = ''
  }
  if (_reasoningBuffer) {
    lastMsg.reasoning = (lastMsg.reasoning || '') + _reasoningBuffer
    lastMsg.reasoningTokens = (lastMsg.reasoningTokens || 0) + _reasoningBuffer.length
    _reasoningBuffer = ''
  }
  _scheduleScroll()
}

// 触发定时 flush（50ms 内合并）
const _scheduleFlusher = () => {
  if (_flushTimer) return
  _flushTimer = setTimeout(_flushChunkBuffer, 50)
}

// 流式结束时立即 flush 剩余内容
const _forceFlush = () => {
  if (_flushTimer) { clearTimeout(_flushTimer); _flushTimer = null }
  if (_scrollThrottleTimer) { clearTimeout(_scrollThrottleTimer); _scrollThrottleTimer = null }
  _flushChunkBuffer()
  // 结束后确保滚到底部
  nextTick(() => scrollToBottom())
}
// ---- 流式渲染性能优化 end ----

// 处理流式内容片段（批量合并，50ms flush 一次）
const handleStreamChunk = (chunk) => {
  if (chatMessages.value.length > 0) {
    const lastMsg = chatMessages.value[chatMessages.value.length - 1]
    if (lastMsg.role === 'assistant' && lastMsg.streaming) {
      // 记录内容开始时间（如果还没有）
      if (!lastMsg.contentStartTime) {
        lastMsg.contentStartTime = Date.now()
      }
      // 写入缓冲区，等待批量 flush
      _chunkBuffer += chunk
      _scheduleFlusher()
    }
  }
}

// 处理流式思考过程（批量合并，50ms flush 一次）
const handleStreamReasoning = (reasoning) => {
  if (chatMessages.value.length > 0) {
    const lastMsg = chatMessages.value[chatMessages.value.length - 1]
    if (lastMsg.role === 'assistant' && lastMsg.streaming) {
      // 初始化思考状态
      if (!lastMsg.reasoning) {
        lastMsg.reasoning = ''
        lastMsg.reasoningStartTime = Date.now()
        lastMsg.isThinking = true
        lastMsg.thinkingExpanded = true // 思考时展开显示
      }
      // 写入缓冲区，等待批量 flush
      _reasoningBuffer += reasoning
      _scheduleFlusher()
    }
  }
}

// 处理流式完成
const handleStreamDone = (data) => {
  // 强制 flush 所有缓冲内容，然后再做收尾处理
  _forceFlush()

  if (chatMessages.value.length > 0) {
    const lastMsg = chatMessages.value[chatMessages.value.length - 1]
    if (lastMsg.role === 'assistant' && lastMsg.streaming) {
      // 立即记录接口实际完成时间
      const actualEndTime = Date.now()
      
      // 如果有思考过程，标记思考完成并延迟折叠
      if (lastMsg.reasoning && lastMsg.reasoningStartTime) {
        lastMsg.isThinking = false
        const thinkingDuration = ((actualEndTime - lastMsg.reasoningStartTime) / 1000).toFixed(1)
        lastMsg.thinkingDuration = thinkingDuration
        
        // 延迟1秒后自动折叠思考过程
        setTimeout(() => {
          if (lastMsg && !lastMsg.streaming) {
            lastMsg.thinkingExpanded = false
          }
        }, 1000)
      }
      
      // 直接标记流式完成（content/reasoning 已由 _forceFlush 写入，data 作为兜底）
      lastMsg.streaming = false
      lastMsg.content = lastMsg.content || data.content || ''
      lastMsg.reasoning = lastMsg.reasoning || data.reasoning || ''
      lastMsg.time = formatTime(new Date())
      
      // 计算统计信息 - 使用接口实际完成时间
      
      // 计算各个时间节点
      const requestToEnd = (actualEndTime - lastMsg.requestStartTime) / 1000
      const contentToEnd = lastMsg.contentStartTime ? (actualEndTime - lastMsg.contentStartTime) / 1000 : null
      const reasoningToEnd = lastMsg.reasoningStartTime ? (actualEndTime - lastMsg.reasoningStartTime) / 1000 : null
      
      console.log('[handleStreamDone] 时间统计:', {
        requestToEnd: requestToEnd.toFixed(3) + 's',
        contentToEnd: contentToEnd ? contentToEnd.toFixed(3) + 's' : 'N/A',
        reasoningToEnd: reasoningToEnd ? reasoningToEnd.toFixed(3) + 's' : 'N/A'
      })
      
      // 优先使用后端返回的 usage 和 timings 信息
      if (data.usage) {
        const usage = data.usage
        const timings = data.timings // llama.cpp 提供的详细性能数据
        
        // console.log('[handleStreamDone] ========== Usage 详细信息 ==========')
        // console.log('[handleStreamDone] 后端返回usage完整对象:', JSON.stringify(usage, null, 2))
        // console.log('[handleStreamDone] - prompt_tokens:', usage.prompt_tokens)
        // console.log('[handleStreamDone] - completion_tokens:', usage.completion_tokens)
        // console.log('[handleStreamDone] - total_tokens:', usage.total_tokens)
        
        // if (timings) {
        //   console.log('[handleStreamDone] ========== Timings 详细信息 ==========')
        //   console.log('[handleStreamDone] 后端返回timings对象:', JSON.stringify(timings, null, 2))
        //   console.log('[handleStreamDone] - prompt_ms:', timings.prompt_ms, 'ms')
        //   console.log('[handleStreamDone] - predicted_ms:', timings.predicted_ms, 'ms')
        //   console.log('[handleStreamDone] - predicted_per_second:', timings.predicted_per_second, 'tokens/s')
        // }
        
        // console.log('[handleStreamDone] 前端统计 - reasoning字符:', lastMsg.reasoningTokens || 0, ', content字符:', lastMsg.contentTokens || 0)
        // console.log('[handleStreamDone] 实际内容长度:', (lastMsg.content || '').length, '字符, reasoning:', (lastMsg.reasoning || '').length, '字符')
        
        // 使用 completion_tokens 作为 token 数
        let totalTokens = 0
        if (usage.completion_tokens !== undefined && usage.completion_tokens > 0) {
          totalTokens = usage.completion_tokens
          // console.log('[handleStreamDone] 使用 completion_tokens:', totalTokens)
        } else if (usage.output_tokens !== undefined && usage.output_tokens > 0) {
          totalTokens = usage.output_tokens
          // console.log('[handleStreamDone] 使用 output_tokens:', totalTokens)
        } else {
          // 如果后端没有提供准确的 tokens，使用前端估算
          console.warn('[handleStreamDone] ⚠️ 后端未提供 completion_tokens，使用前端估算')
          const contentLength = (lastMsg.content || '').length
          const reasoningLength = (lastMsg.reasoning || '').length
          totalTokens = Math.ceil((contentLength + reasoningLength) * 1.7)
          // console.log('[handleStreamDone] 前端估算tokens:', totalTokens)
        }
        
        // console.log('[handleStreamDone] ========== 时间和速度计算 ==========')
        // console.log('[handleStreamDone] timings对象存在:', !!timings)
        // console.log('[handleStreamDone] predicted_ms存在:', timings ? timings.predicted_ms : 'N/A')
        // console.log('[handleStreamDone] predicted_per_second存在:', timings ? timings.predicted_per_second : 'N/A')
        
        // 优先使用后端 timings 中的精确数据
        let displayDuration, tokensPerSecond
        
        if (timings && timings.predicted_ms !== undefined) {
          // 使用后端提供的精确生成时间和速度
          // prompt_ms: 提示词处理时间，predicted_ms: 实际生成时间
          const promptTime = (timings.prompt_ms || 0) / 1000
          const predictedTime = timings.predicted_ms / 1000
          displayDuration = promptTime + predictedTime // 总耗时
          
          // console.log('[handleStreamDone] ✅ 进入 timings 分支')
          // console.log('[handleStreamDone] - prompt处理时间:', promptTime.toFixed(2), 's')
          // console.log('[handleStreamDone] - 生成时间:', predictedTime.toFixed(2), 's')
          // console.log('[handleStreamDone] - 总耗时:', displayDuration.toFixed(2), 's')
          
          // 优先使用后端计算的速度
          if (timings.predicted_per_second !== undefined) {
            tokensPerSecond = timings.predicted_per_second.toFixed(1)
            // console.log('[handleStreamDone] ✅✅ 使用后端 predicted_per_second:', tokensPerSecond, 'tokens/s')
          } else {
            // 如果没有后端速度，用生成时间计算
            tokensPerSecond = totalTokens > 0 && predictedTime > 0
              ? (totalTokens / predictedTime).toFixed(1)
              : '0.0'
            // console.log('[handleStreamDone] 📊 使用 timings.predicted_ms 计算速度')
            // console.log('[handleStreamDone] - 生成时间:', predictedTime.toFixed(2), 's')
            // console.log('[handleStreamDone] - 计算速度:', tokensPerSecond, 'tokens/s')
          }
        } else {
          // console.log('[handleStreamDone] ⚠️ 未进入 timings 分支')
          // 如果没有 timings，使用前端记录的时间
          displayDuration = requestToEnd
          tokensPerSecond = totalTokens > 0 && displayDuration > 0
            ? (totalTokens / displayDuration).toFixed(1)
            : '0.0'
          // console.log('[handleStreamDone] ⚠️ 后端未提供 timings，使用前端时间')
          // console.log('[handleStreamDone] - 请求总时间:', displayDuration.toFixed(2), 's')
          // console.log('[handleStreamDone] - 计算速度:', tokensPerSecond, 'tokens/s')
        }
        
        // console.log('[handleStreamDone] ========== 最终统计 ==========')
        // console.log('[handleStreamDone] 显示耗时:', displayDuration.toFixed(2), 's')
        // console.log('[handleStreamDone] tokens:', totalTokens)
        // console.log('[handleStreamDone] tokens/s:', tokensPerSecond)
        // console.log('[handleStreamDone] ==========================================')
        
        lastMsg.stats = {
          duration: displayDuration.toFixed(2),
          tokensPerSecond: tokensPerSecond,
          totalTokens: totalTokens
        }
      } else {
        // console.log('[handleStreamDone] 后端未返回usage信息，使用前端估算')
        
        // 前端估算的 token 数
        const contentLength = (lastMsg.content || '').length
        const reasoningLength = (lastMsg.reasoning || '').length
        const totalLength = contentLength + reasoningLength
        const totalTokens = Math.ceil(totalLength * 1.7)
        
        // console.log('[handleStreamDone] 内容总长度:', totalLength, '字符')
        // console.log('[handleStreamDone] 估算tokens:', totalTokens)
        
        // 使用完整的请求时间
        const displayDuration = requestToEnd
        const generationDuration = requestToEnd
        
        // console.log('[handleStreamDone] 使用完整请求时间:', displayDuration.toFixed(2), 's')
        
        const totalTokensPerSecond = totalTokens > 0 && generationDuration > 0
          ? (totalTokens / generationDuration).toFixed(1)
          : '0.0'
        
        // console.log(`[handleStreamDone] 前端估算 - 显示耗时: ${displayDuration.toFixed(2)}s, 计算耗时: ${generationDuration.toFixed(2)}s, tokens: ${totalTokens}, tokens/s: ${totalTokensPerSecond}`)
        
        lastMsg.stats = {
          duration: displayDuration.toFixed(2),
          tokensPerSecond: totalTokensPerSecond,
          totalTokens: totalTokens
        }
      }
      
      scrollToBottom()

      // 回答完成后，弹出反馈条（10秒倒计时）
      triggerFeedbackBar(lastMsg.content || '')
    }
  }
}

// ===== 经验库：反馈条触发 =====
const triggerFeedbackBar = (answerContent) => {
  if (!answerContent || answerContent.length < 10) return

  // 找最后一条用户消息
  const userMsgs = chatMessages.value.filter(m => m.role === 'user')
  if (userMsgs.length === 0) return
  const lastQuestion = userMsgs[userMsgs.length - 1].content || ''

  // 清理上一次未完成的倒计时
  if (feedbackBar.value.timer) {
    clearInterval(feedbackBar.value.timer)
  }

  feedbackBar.value = {
    visible: true,
    countdown: 10,
    question: lastQuestion,
    answer: answerContent,
    timer: null,
    saving: false,
  }

  // 启动倒计时
  feedbackBar.value.timer = setInterval(() => {
    feedbackBar.value.countdown -= 1
    if (feedbackBar.value.countdown <= 0) {
      dismissFeedbackBar()
    }
  }, 1000)
}

// 关闭反馈条（不记录）
const dismissFeedbackBar = () => {
  if (feedbackBar.value.timer) {
    clearInterval(feedbackBar.value.timer)
  }
  feedbackBar.value.visible = false
  feedbackBar.value.timer = null
}

// 用户点击反馈（关闭反馈条）
const submitFeedback = (isUseful) => {
  if (feedbackBar.value.timer) {
    clearInterval(feedbackBar.value.timer)
  }
  feedbackBar.value.visible = false
  feedbackBar.value.timer = null
}


// 监听设备列表变化，自动获取设备模型信息
watch(() => props.devices, async (newDevices) => {
  console.log('[watch devices] 设备列表变化，开始获取模型信息')
  
  const onlineDeviceList = newDevices.filter(device => device.status === 'online' || !device.status)
  
  // 只获取缓存中没有的设备信息
  const needFetchDevices = onlineDeviceList.filter(device => !deviceModelsCache.value.has(device.ip))
  
  if (needFetchDevices.length > 0) {
    console.log('[watch devices] 需要获取模型信息的设备数:', needFetchDevices.length)
    const fetchPromises = needFetchDevices.map(device => fetchDeviceInfo(device.ip))
    await Promise.allSettled(fetchPromises)
  }
}, { deep: true, immediate: true })

// AI Assistant Component
const fetchAiAssistant = async () => {
  await refreshDevices()
}

// 刷新设备列表并获取设备模型信息
const refreshDevices = async () => {
  loading.value = true
  
  try {
    // 获取所有在线设备的模型信息
    const onlineDeviceList = props.devices.filter(device => device.status === 'online' || !device.status)
    
    console.log('[refreshDevices] 开始获取设备模型信息，在线设备数:', onlineDeviceList.length)
    
    // 并发获取所有设备的模型信息
    const fetchPromises = onlineDeviceList.map(device => fetchDeviceInfo(device.ip))
    await Promise.allSettled(fetchPromises)
    
    console.log('[refreshDevices] 设备模型信息缓存:', Object.fromEntries(deviceModelsCache.value))
  } catch (error) {
    console.error('[refreshDevices] 获取设备信息失败:', error)
  } finally {
    loading.value = false
  }
}

// 查询设备当前运行中的模型，同步到 loadedModel
const syncRunningModels = async () => {
  if (!selectedDevice.value) return
  try {
    const result = await AIGetModels(selectedDevice.value.ip)
    if (!result?.success) return
    const raw = result.data
    // 502 = 启动中，视为无运行模型
    if (raw?.code === 502 || raw?.code === '502') {
      loadedModel.value = null
      return
    }
    // 兼容两种格式
    const modelsList = raw?.object === 'list' ? raw
      : (raw?.data?.object === 'list' ? raw.data : null)
    if (!modelsList?.data?.length) {
      loadedModel.value = null
      return
    }
    const runningIds = modelsList.data.map(m => m.id)
    console.log('[syncRunningModels] 设备运行中的模型:', runningIds)

    // 对话模型：在 modelList 中找匹配的
    const runningChat = modelList.value.find(m => runningIds.includes(m.name))
    loadedModel.value = runningChat || null
    if (runningChat) {
      selectedModelName.value = runningChat.name
      selectedModel.value = runningChat
      localOpenAIConfig.value.model = runningChat.name  // 同步模型名，避免发送时用旧值
    }

    if (runningChat) console.log('[syncRunningModels] 对话模型已同步:', runningChat.name)
  } catch (e) {
    console.warn('[syncRunningModels] 查询运行状态失败:', e)
  }
}

// 选择设备
const selectDevice = async (device) => {
  selectedDevice.value = device

  // 重置状态
  loadedModel.value = null
  chatMessages.value = []
  selectedModelName.value = ''

  // 获取模型列表，然后同步运行状态
  await fetchModelList()
}

// 获取模型列表
const fetchModelList = async () => {
  if (!selectedDevice.value) {
    return
  }
  
  loadingModels.value = true
  modelList.value = []
  selectedModel.value = null
  selectedModelName.value = ''
  
  try {
    console.log('[fetchModelList] 获取模型列表:', selectedDevice.value.ip)
    const result = await GetLLMModelList(selectedDevice.value.ip, getDeviceToken(selectedDevice.value.ip))
    console.log('[fetchModelList] 返回结果:', result)
    
    if (result.success && result.data && result.data.list) {
      modelList.value = result.data.list

      // 查询设备当前运行中的模型，同步状态（不再盲目触发 handleLoadModel）
      await syncRunningModels()

      // 如果没有运行中模型且用户未手动选择，不自动选中（避免 @change 不触发的问题）
      // 用户自行从下拉框选择即可

    } else {
      ElMessage.warning(result.message || t('aiAssistant.fetchModelFailed'))
    }
  } catch (error) {
    console.error('[fetchModelList] 获取模型列表失败:', error)
    ElMessage.error(t('aiAssistant.fetchModelFailed') + ': ' + error.message)
  } finally {
    loadingModels.value = false
  }
}

// 下拉框选择模型（不自动加载，等用户点「启动模型」）
const handleModelChange = (modelName) => {
  const model = modelList.value.find(m => m.name === modelName)
  if (model) {
    selectedModel.value = model
    console.log('[handleModelChange] 选择对话模型:', model.name)
  }
}


// 轮询检查模型运行状态
const pollModelStatus = async (modelName, maxAttempts = 15, interval = 2000) => {
  let progressMessage = null
  
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      console.log(`[pollModelStatus] 第 ${attempt}/${maxAttempts} 次检查...`)
      
      const progress = Math.round((attempt / maxAttempts) * 100)
      const progressText = `${t('aiAssistant.startingModel')} ${progress}% (${attempt}/${maxAttempts})`
      
      if (progressMessage) progressMessage.close()
      progressMessage = ElMessage.info({ message: progressText, duration: 0, showClose: false })
      
      const result = await AIGetModels(selectedDevice.value.ip)
      
      if (result?.success) {
        const raw = result.data
        const rawCode = raw?.code
        console.log(`[pollModelStatus] 第 ${attempt} 次 code=${rawCode} raw:`, JSON.stringify(raw))

        // 502 = 启动中，继续等待
        if (rawCode === 502 || rawCode === '502') {
          console.log(`[pollModelStatus] 设备启动中(502)，继续等待...`)
        } else {
          // 两种就绪格式:
          // A) 直接返回 { object:'list', data:[...] }（无 code 字段）
          // B) 包在外层 { code:0, data:{ object:'list', data:[...] } }
          const modelsList = raw?.object === 'list' ? raw
            : (raw?.data?.object === 'list' ? raw.data : null)

          if (modelsList?.data?.length > 0) {
            const runningModels = modelsList.data
            console.log(`[pollModelStatus] 当前运行模型:`, runningModels.map(m => m.id))
            const targetModel = runningModels.find(m => m.id === modelName)
            if (targetModel) {
              console.log(`[pollModelStatus] ✅ 模型已启动:`, targetModel)
              if (progressMessage) progressMessage.close()
              const elapsedTime = Math.ceil(attempt * interval / 1000)
              ElMessage.success(t('aiAssistant.modelStartSuccess', { time: elapsedTime }))
              return true
            } else {
              console.log(`[pollModelStatus] 模型尚未出现在列表中，继续等待...`)
            }
          } else {
            console.log(`[pollModelStatus] 模型列表为空，继续等待...`)
          }
        }
      } else {
        console.warn(`[pollModelStatus] 第 ${attempt} 次请求失败`)
      }
      
      if (attempt < maxAttempts) {
        console.log(`[pollModelStatus] 等待 ${interval}ms 后继续检查...`)
        await new Promise(resolve => setTimeout(resolve, interval))
      }
    } catch (error) {
      console.error(`[pollModelStatus] 第 ${attempt} 次检查出错:`, error)
      if (attempt < maxAttempts) {
        await new Promise(resolve => setTimeout(resolve, interval))
      }
    }
  }
  
  if (progressMessage) progressMessage.close()
  const totalTime = maxAttempts * interval / 1000
  console.error(`[pollModelStatus] 轮询超时，模型未能在 ${totalTime} 秒内启动`)
  return false
}

// 加载模型
const handleLoadModel = async () => {
  if (!selectedDevice.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return
  }
  
  if (!selectedModel.value) {
    ElMessage.warning(t('aiAssistant.selectModelFirst'))
    return
  }
  
  console.log('[handleLoadModel] 对话模型:', selectedModel.value.name)
  
  loadingModel.value = true
  
  try {
    // 检查模型是否已运行
    let chatRunning = false

    try {
      const modelsResult = await AIGetModels(selectedDevice.value.ip)
      if (modelsResult?.success) {
        const raw = modelsResult.data
        console.log('[handleLoadModel] 当前运行模型 raw:', JSON.stringify(raw))
        // 兼容两种格式（直接 list 或 502/带code外层）
        if (raw?.code !== 502 && raw?.code !== '502') {
          const modelsList = raw?.object === 'list' ? raw
            : (raw?.data?.object === 'list' ? raw.data : null)
          if (modelsList?.data?.length > 0) {
            const runningModels = modelsList.data
            chatRunning = !!runningModels.find(m => m.id === selectedModel.value.name)
            console.log('[handleLoadModel] chatRunning:', chatRunning)
          }
        }
      }
    } catch (error) {
      console.warn('[handleLoadModel] 检查模型运行状态失败:', error)
    }

    console.log('[handleLoadModel] chatRunning:', chatRunning)

    if (!chatRunning) {
      // startLLMService 内部已包含串行启动 + 轮询等待
      const startResult = await startLLMService(true)
      if (!startResult) {
        return
      }
    } else {
      console.log('[handleLoadModel] 模型已运行，直接同步状态')
      ElMessage.success(t('aiAssistant.modelAlreadyRunning'))
    }
    
    loadedModel.value = selectedModel.value
    localOpenAIConfig.value.model = selectedModel.value?.name || localOpenAIConfig.value.model
    chatMessages.value = []
  } catch (error) {
    console.error('[handleLoadModel] 加载模型失败:', error)
    ElMessage.closeAll()
    ElMessage.error(t('aiAssistant.loadModelFailed') + error.message)
  } finally {
    loadingModel.value = false
  }
}

// 富文本编辑器方法
// 格式化文本
const formatText = (command, value = null) => {
  document.execCommand(command, false, value)
  editorRef.value?.focus()
}

// 清除格式
const clearFormat = () => {
  const selection = window.getSelection()
  if (selection.rangeCount > 0) {
    const range = selection.getRangeAt(0)
    const text = range.toString()
    if (text) {
      document.execCommand('removeFormat', false, null)
      document.execCommand('insertText', false, text)
    }
  }
  editorRef.value?.focus()
}

// 处理编辑器输入
const handleEditorInput = () => {
  if (editorRef.value) {
    // 获取纯文本用于验证
    inputMessage.value = editorRef.value.innerText || ''
  }
}

// 处理编辑器按键
const handleEditorKeydown = (event) => {
  // Enter 发送，Shift+Enter 换行
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
  
  // 快捷键
  if (event.ctrlKey || event.metaKey) {
    switch (event.key.toLowerCase()) {
      case 'b':
        event.preventDefault()
        formatText('bold')
        break
      case 'i':
        event.preventDefault()
        formatText('italic')
        break
      case 'u':
        event.preventDefault()
        formatText('underline')
        break
    }
  }
}

// 处理粘贴
const handleEditorPaste = async (event) => {
  event.preventDefault()
  
  const clipboardData = event.clipboardData || event.originalEvent?.clipboardData
  if (!clipboardData) return
  
  // 检查是否包含文件（图片或视频）
  const items = Array.from(clipboardData.items || [])
  const files = []
  
  for (const item of items) {
    if (item.kind === 'file') {
      const file = item.getAsFile()
      if (file) {
        files.push(file)
      }
    }
  }
  
  // 如果有文件，处理文件粘贴
  if (files.length > 0) {
    await handlePastedFiles(files)
    return
  }
  
  // 没有文件，处理文本粘贴
  const text = clipboardData.getData('text/plain') || ''
  if (text) {
    document.execCommand('insertText', false, text)
  }
}

// 处理粘贴的文件
const handlePastedFiles = async (files) => {
  if (files.length === 0) return
  
  const maxSize = 1000 * 1024 * 1024 // 100MB
  let successCount = 0
  
  for (const file of files) {
    // 检查文件类型
    if (!file.type.startsWith('image/') && !file.type.startsWith('video/')) {
      ElMessage.warning(`不支持的文件类型: ${file.type || '未知'}`)
      continue
    }
    
    // 检查文件大小
    if (file.size > maxSize) {
      ElMessage.warning(`文件 ${file.name || '未命名'} ${t('aiAssistant.fileTooLarge')}`)
      continue
    }
    
    // 创建预览URL
    const preview = URL.createObjectURL(file)
    
    // 生成文件名（粘贴的文件可能没有名称）
    const fileName = file.name || `pasted-${Date.now()}.${file.type.split('/')[1]}`
    
    // 添加到上传列表
    uploadedFiles.value.push({
      file: file,
      name: fileName,
      type: file.type,
      size: file.size,
      preview: preview
    })
    
    successCount++
  }
  
  if (successCount > 0) {
    ElMessage.success(t('aiAssistant.pastedFiles', { count: successCount }))
  }
}

// 拖拽相关
const isDragging = ref(false)

const handleDragOver = (event) => {
  event.preventDefault()
  event.stopPropagation()
  
  // 检查是否拖拽的是文件
  if (event.dataTransfer?.types.includes('Files')) {
    isDragging.value = true
    event.dataTransfer.dropEffect = 'copy'
    
    // 添加拖拽样式
    if (editorRef.value) {
      editorRef.value.classList.add('dragging')
      // 暂时禁用编辑
      editorRef.value.setAttribute('contenteditable', 'false')
    }
  }
}

const handleDragLeave = (event) => {
  event.preventDefault()
  event.stopPropagation()
  
  // 只有当鼠标真正离开编辑器区域时才移除样式
  const relatedTarget = event.relatedTarget
  if (!editorRef.value?.contains(relatedTarget)) {
    isDragging.value = false
    
    // 移除拖拽样式并恢复编辑
    if (editorRef.value) {
      editorRef.value.classList.remove('dragging')
      editorRef.value.setAttribute('contenteditable', 'true')
    }
  }
}

const handleDrop = async (event) => {
  event.preventDefault()
  event.stopPropagation()
  
  isDragging.value = false
  
  // 移除拖拽样式并恢复编辑
  if (editorRef.value) {
    editorRef.value.classList.remove('dragging')
    editorRef.value.setAttribute('contenteditable', 'true')
  }
  
  // 获取拖拽的文件
  const files = Array.from(event.dataTransfer?.files || [])
  
  if (files.length > 0) {
    await handlePastedFiles(files)
  }
}

// 编辑器焦点
const handleEditorFocus = () => {
  editorFocused.value = true
}

const handleEditorBlur = () => {
  editorFocused.value = false
}

// 获取编辑器HTML内容
const getEditorHtml = () => {
  return editorRef.value?.innerHTML || ''
}

// 格式化消息内容的缓存（避免重复格式化）
const formatCache = new Map()
const MAX_CACHE_SIZE = 1000

// 【最终优化方案】结合节流+渲染优化
// 流式时仍然使用markdown渲染，但通过RAF节流降低渲染频率
let renderPending = false
let lastRenderContent = ''

const formatMessageContent = (content, isStreaming = false) => {
  if (!content) return ''
  
  // 流式输出时：使用RAF节流，避免过于频繁的渲染
  if (isStreaming) {
    // 如果内容没变化，返回缓存
    if (content === lastRenderContent && formatCache.has(content)) {
      return formatCache.get(content)
    }
    
    // 使用RAF节流（但不影响最终渲染）
    if (!renderPending) {
      renderPending = true
      requestAnimationFrame(() => {
        renderPending = false
      })
    }
    
    // 直接渲染（markdown-it很快，主要是浏览器的DOM更新慢）
    const result = md.render(content)
    lastRenderContent = content
    formatCache.set(content, result)
    return result
  }
  
  // 完成后：使用完整Markdown渲染（带缓存）
  if (formatCache.has(content)) {
    return formatCache.get(content)
  }
  
  const startTime = performance.now()
  const result = md.render(content)
  const endTime = performance.now()
  
  console.log(`[formatMessageContent] 完整渲染耗时: ${(endTime - startTime).toFixed(2)}ms, 内容长度: ${content.length}`)
  
  // 缓存结果
  if (formatCache.size >= MAX_CACHE_SIZE) {
    const entries = Array.from(formatCache.entries())
    formatCache.clear()
    entries.slice(-Math.floor(MAX_CACHE_SIZE / 2)).forEach(([k, v]) => formatCache.set(k, v))
  }
  formatCache.set(content, result)
  
  return result
}

// 处理代码复制（事件委托）
nextTick(() => {
  document.addEventListener('click', (e) => {
    const btn = e.target.closest('.copy-btn')
    if (btn && btn.dataset.code) {
      const code = decodeURIComponent(btn.dataset.code)
      navigator.clipboard.writeText(code).then(() => {
        const originalHtml = btn.innerHTML
        btn.innerHTML = '<span style="color: #10b981;">已复制!</span>'
        setTimeout(() => {
          btn.innerHTML = originalHtml
        }, 2000)
        ElMessage.success(t('aiAssistant.copiedText'))
      }).catch(() => {
        ElMessage.error(t('aiAssistant.copyFailed'))
      })
    }
  })
})

// 清空编辑器
const clearEditor = () => {
  if (editorRef.value) {
    editorRef.value.innerHTML = ''
    inputMessage.value = ''
  }
}

// 发送消息
const sendMessage = async () => {
  const message = inputMessage.value.trim()
  const hasFiles = uploadedFiles.value.length > 0
  
  // 至少需要消息或文件之一
  if (!message && !hasFiles) {
    return
  }

  // 检查是否可以对话
  if (!canChat.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceAndModel'))
    return
  }
  
  // 清除手动停止标记（开始新的对话）
  wasStoppedManually.value = false
  
  // 准备文件数据
  const userMessageFiles = []
  
  // 如果有上传的文件,转换为base64
  if (hasFiles) {
    for (const uploadedFile of uploadedFiles.value) {
      try {
        const base64 = await fileToBase64(uploadedFile.file)
        userMessageFiles.push({
          name: uploadedFile.name,
          type: uploadedFile.type,
          data: base64 // API 需要的 base64 数据
        })
      } catch (error) {
        console.error('[sendMessage] 文件转base64失败:', error)
        ElMessage.error(`文件 ${uploadedFile.name} ${t('aiAssistant.fileProcessFailed')}`)
      }
    }
  }
  
  // 添加用户消息
  const userMessage = {
    role: 'user',
    content: message || '[发送了图片/视频]',
    time: formatTime(new Date()),
    files: userMessageFiles.map(f => ({
      name: f.name,
      type: f.type,
      data: f.data, // 保存 base64 数据到消息中用于显示
      preview: uploadedFiles.value.find(uf => uf.name === f.name)?.preview // 也保留 preview URL
    }))
  }
  chatMessages.value.push(userMessage)
  
  // 清空输入和文件
  clearEditor()
  clearUploadedFiles()
  
  // 滚动到底部
  scrollToBottom()
  
  sending.value = true
  
  try {
    // 使用本地OpenAI API(包含文件信息)
    await sendToLocalOpenAI(message, userMessageFiles)
  } catch (error) {
    console.error('[sendMessage] 发送消息失败:', error)
    ElMessage.error(t('aiAssistant.sendFailed') + error.message)
    chatMessages.value.push({
      role: 'assistant',
      content: '抱歉，发送消息时出错了。',
      time: formatTime(new Date())
    })
  } finally {
    sending.value = false
    scrollToBottom()
  }
}

// 文件转base64
const fileToBase64 = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

// 清空已上传的文件
const clearUploadedFiles = () => {
  // 释放所有预览URL
  uploadedFiles.value.forEach(file => {
    if (file.preview) {
      URL.revokeObjectURL(file.preview)
    }
  })
  uploadedFiles.value = []
}

// 发送到本地OpenAI（流式）— Go 后端处理，通过 Wails 事件推送 chunk
const sendToLocalOpenAI = async (message, files = []) => {
  console.log('[sendToLocalOpenAI] 发送消息:', message, '文件数量:', files.length)
  
  // 记录请求开始时间
  const requestStartTime = Date.now()
  
  // 添加一个空的assistant消息用于接收流式内容
  const assistantMessage = {
    role: 'assistant',
    content: '',
    reasoning: '', // 思考过程内容
    isThinking: false, // 是否正在思考
    thinkingExpanded: false, // 思考区域是否展开
    thinkingDuration: null, // 思考用时（秒）
    streaming: true,
    time: '',
    requestStartTime: requestStartTime, // 记录请求开始时间
    reasoningStartTime: null,
    contentStartTime: null,
    reasoningTokens: 0,
    contentTokens: 0
  }
  chatMessages.value.push(assistantMessage)
  scrollToBottom()
  
  // 构建消息历史（保留最近10条，约5轮对话，配合 ctx-size=2048）
  // 确保第一条是 user（OpenAI API 要求 system 之后首条必须是 user）
  let historySlice = chatMessages.value.filter(msg => !msg.streaming).slice(-10)
  // 若截断后第一条是 assistant，往后找第一条 user 开始
  while (historySlice.length > 0 && historySlice[0].role === 'assistant') {
    historySlice = historySlice.slice(1)
  }
  const recentMessages = historySlice
    .map(msg => {
      const msgData = {
        role: msg.role,
        content: msg.content
      }
      
      // 如果是用户消息且有文件,添加文件信息(仅在当前消息)
      if (msg.role === 'user' && files.length > 0 && msg === chatMessages.value[chatMessages.value.length - 2]) {
        // 构建多模态内容
        const content = []
        
        // 添加文本内容
        if (message) {
          content.push({
            type: 'text',
            text: message
          })
        }
        
        // 添加文件内容
        files.forEach(file => {
          if (file.type.startsWith('image/')) {
            content.push({
              type: 'image_url',
              image_url: {
                url: file.data
              }
            })
          } else if (file.type.startsWith('video/')) {
            content.push({
              type: 'video_url',
              video_url: {
                url: file.data
              }
            })
          }
        })
        
        msgData.content = content
      }
      
      return msgData
    })
  
  // 如果启用了系统提示词，在消息历史开头添加
  if (systemPromptEnabled.value && systemPrompt.value.trim()) {
    recentMessages.unshift({
      role: 'system',
      content: systemPrompt.value.trim()
    })
  }

  // 生成唯一 sessionId，Go 后端用于区分会话
  const sessionId = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  currentSessionId.value = sessionId

  // 注册事件监听
  const chunkEvent = `ai:chunk:${sessionId}`
  const doneEvent = `ai:done:${sessionId}`
  const errorEvent = `ai:error:${sessionId}`

  const cleanup = () => {
    Events.Off(chunkEvent)
    Events.Off(doneEvent)
    Events.Off(errorEvent)
    currentSessionId.value = null
  }

  Events.On(chunkEvent, (event) => {
    const data = event?.data
    if (data?.content) handleStreamChunk(data.content)
    if (data?.reasoning) handleStreamReasoning(data.reasoning)
  })

  Events.On(doneEvent, (event) => {
    const data = event?.data
    cleanup()
    handleStreamDone({
      content: data?.content || '',
      reasoning: data?.reasoning || '',
      usage: data?.usage || null
    })
  })

  Events.On(errorEvent, (event) => {
    const data = event?.data
    cleanup()
    const errMsg = data?.message || '未知错误'
    console.error('[sendToLocalOpenAI] Go 后端报错:', errMsg)
    // 移除 streaming 消息
    chatMessages.value.pop()
    sending.value = false
    ElMessage.error(t('aiAssistant.sendFailed') + errMsg)
    chatMessages.value.push({
      role: 'assistant',
      content: '抱歉，发送消息时出错了。',
      time: formatTime(new Date()),
      streaming: false
    })
  })

  try {
    // 调用 Go 后端（异步，立即返回；chunk 通过事件推送）
    // model 优先用设备上实际加载的模型名，避免 localOpenAIConfig 过期
    const actualModel = loadedModel.value?.name || localOpenAIConfig.value.model
    await AISendMessage({
      deviceIp: selectedDevice.value.ip,
      model: actualModel,
      messages: recentMessages,
      temperature: localOpenAIConfig.value.temperature,
      maxTokens: localOpenAIConfig.value.maxTokens,
      sessionId: sessionId
    })
  } catch (error) {
    cleanup()
    console.error('[sendToLocalOpenAI] 调用失败:', error)
    chatMessages.value.pop()
    sending.value = false
    ElMessage.error(t('aiAssistant.sendFailed') + error.message)
    chatMessages.value.push({
      role: 'assistant',
      content: '抱歉，发送消息时出错了。',
      time: formatTime(new Date()),
      streaming: false
    })
  }
}

// 清空对话
const clearChat = () => {
  ElMessageBox.confirm(t('aiAssistant.clearHistoryConfirm'), t('common.prompt'), {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    chatMessages.value = []
    ElMessage.success(t('aiAssistant.historyCleared'))
  }).catch(() => {})
}

// 触发文件上传
const triggerFileUpload = () => {
  fileInputRef.value?.click()
}

// 处理文件选择
const handleFileSelect = async (event) => {
  const files = Array.from(event.target.files || [])
  
  if (files.length === 0) {
    return
  }
  
  // 验证文件类型和大小
  const maxSize = 1000 * 1024 * 1024 // 100MB
  
  for (const file of files) {
    // 检查文件类型
    if (!file.type.startsWith('image/') && !file.type.startsWith('video/')) {
      ElMessage.warning(`文件 ${file.name} ${t('aiAssistant.notImgVideo')}`)
      continue
    }
    
    console.log(maxSize, file.size)
    // 检查文件大小
    if (file.size > maxSize) {
      ElMessage.warning(`文件 ${file.name} 超过1000MB限制`)
      continue
    }
    
    // 创建预览URL
    const preview = URL.createObjectURL(file)
    
    // 添加到上传列表
    uploadedFiles.value.push({
      file: file,
      name: file.name,
      type: file.type,
      size: file.size,
      preview: preview
    })
  }
  
  // 清空input,允许重复选择同一文件
  event.target.value = ''
  
  if (uploadedFiles.value.length > 0) {
    ElMessage.success(t('aiAssistant.selectedFiles', { count: files.length }))
  }
}

// 移除文件
const removeFile = (index) => {
  const file = uploadedFiles.value[index]
  // 释放预览URL
  if (file.preview) {
    URL.revokeObjectURL(file.preview)
  }
  uploadedFiles.value.splice(index, 1)
}

// 复制消息
const copyMessage = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success(t('aiAssistant.copiedClipboard'))
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error(t('aiAssistant.copyFailed'))
  }
}

// 重新生成最后一条回复
const regenerateMessage = async () => {
  if (chatMessages.value.length < 2) {
    return
  }
  
  // 移除最后一条assistant消息
  chatMessages.value.pop()
  
  // 获取最后一条用户消息
  const lastUserMsg = [...chatMessages.value].reverse().find(msg => msg.role === 'user')
  if (!lastUserMsg) {
    return
  }
  
  // 重新发送
  sending.value = true
  try {
    await sendToLocalOpenAI(lastUserMsg.content)
  } catch (error) {
    console.error('[regenerateMessage] 重新生成失败:', error)
    ElMessage.error(t('aiAssistant.regenerateFailed') + error.message)
  } finally {
    sending.value = false
  }
}

// 停止生成
const stopGeneration = () => {
  // 通知 Go 后端取消当前会话
  if (currentSessionId.value) {
    AICancelMessage(currentSessionId.value).catch(e => console.warn('[stopGeneration] 取消失败:', e))
    currentSessionId.value = null
  }
  // 兼容旧的 abortController（保留，以防其他地方还在用）
  if (abortController.value) {
    abortController.value.abort()
    abortController.value = null
  }
  
  // 标记最后一条消息为非流式
  if (chatMessages.value.length > 0) {
    const lastMsg = chatMessages.value[chatMessages.value.length - 1]
    if (lastMsg.streaming) {
      lastMsg.streaming = false
      lastMsg.time = formatTime(new Date())
      ElMessage.info(t('aiAssistant.generationStopped'))
    }
  }
  
  sending.value = false
  wasStoppedManually.value = true // 标记为手动停止
}

// 继续生成
const continueGeneration = async () => {
  if (!canChat.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceAndModel'))
    return
  }
  
  if (chatMessages.value.length === 0) {
    return
  }
  
  // 获取最后一条助手消息
  const lastMsg = chatMessages.value[chatMessages.value.length - 1]
  if (lastMsg.role !== 'assistant') {
    return
  }
  
  // 清除手动停止标记
  wasStoppedManually.value = false
  
  // 构建继续生成的提示
  const continuePrompt = '继续'
  
  // 添加用户消息
  chatMessages.value.push({
    role: 'user',
    content: continuePrompt,
    time: formatTime(new Date()),
    streaming: false
  })
  
  scrollToBottom()
  
  // 发送请求
  sending.value = true
  try {
    await sendToLocalOpenAI(continuePrompt)
  } catch (error) {
    console.error('[continueGeneration] 继续生成失败:', error)
    ElMessage.error(t('aiAssistant.continueFailed') + error.message)
  } finally {
    sending.value = false
  }
}

// 开启新对话
const startNewChat = () => {
  if (chatMessages.value.length === 0) {
    return
  }
  
  ElMessageBox.confirm(
    '开启新对话将清空当前对话记录，是否继续？',
    '确认开启新对话',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    chatMessages.value = []
    wasStoppedManually.value = false
    // 清空输入框
    inputMessage.value = ''
    if (editorRef.value) {
      editorRef.value.innerHTML = ''
    }
    // 清空上传的文件
    uploadedFiles.value.forEach(file => {
      if (file.preview) {
        URL.revokeObjectURL(file.preview)
      }
    })
    uploadedFiles.value = []
    
    ElMessage.success(t('aiAssistant.newChatStarted'))
  }).catch(() => {
    // 用户取消
  })
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (chatMessagesRef.value) {
      chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
    }
  })
}

// 代码复制功能
const handleCopyCode = (e) => {
  const btn = e.target.closest('.copy-btn')
  if (btn && btn.dataset.code) {
    const code = decodeURIComponent(btn.dataset.code)
    navigator.clipboard.writeText(code).then(() => {
      btn.textContent = '已复制!'
      setTimeout(() => {
        btn.textContent = '复制'
      }, 2000)
      ElMessage.success(t('aiAssistant.copiedText'))
    }).catch((err) => {
      console.error('复制失败:', err)
      ElMessage.error(t('aiAssistant.copyFailed'))
    })
  }
}

// 格式化时间
const formatTime = (date) => {
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

// 导入模型
const handleImportModel = async () => {
  if (!selectedDevice.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return
  }
  
  let result
  
  try {
    // 使用文件选择对话框
    console.log('打开文件选择对话框...')
    result = await SelectZipFile()
    console.log('文件选择结果:', result)
  } catch (error) {
    console.error('打开文件对话框失败:', error)
    // 用户取消操作，不显示错误提示
    return
  }
  
  // 检查文件选择结果
  if (!result.success || !result.path) {
    if (result.message && result.message !== '用户取消选择') {
      ElMessage.error(result.message || t('aiAssistant.fileSelectFailed'))
    }
    // 用户取消选择，静默返回
    return
  }
  
  const filePath = result.path
  const fileName = filePath.split('\\').pop().split('/').pop()
  const deviceIp = selectedDevice.value.ip
  
  console.log('准备上传:', {
    fileName: fileName,
    filePath: filePath,
    deviceIp: deviceIp,
    token: getDeviceToken(deviceIp) ? '已设置' : '未设置'
  })
  
  // 确认上传
  try {
    await ElMessageBox.confirm(
      `确定要将模型 "${fileName}" 上传到设备 ${deviceIp} 吗？`,
      '确认上传',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    // 用户取消确认，静默返回
    return
  }
  
  // 执行上传
  importing.value = true
  
  try {
    ElMessage.info({
      message: `正在上传模型到 ${deviceIp}，请耐心等待...`,
      duration: 0,
      showClose: true
    })
    
    console.log('[UploadLLMModel] 调用Go函数上传...')
    
    // 使用Go函数上传
    const uploadResult = await UploadLLMModel(deviceIp, filePath, getDeviceToken(deviceIp))
    
    console.log('[UploadLLMModel] 上传结果:', uploadResult)
    
    ElMessage.closeAll()
    
    if (uploadResult.success || uploadResult.code === 0) {
      ElMessage.success({
        message: '模型导入成功',
        duration: 3000
      })
      // 上传成功后自动刷新模型列表
      await fetchModelList()
    } else {
      ElMessage.error({
        message: uploadResult.message || uploadResult.msg || '模型导入失败',
        duration: 5000,
        showClose: true
      })
    }
  } catch (error) {
    console.error('导入模型失败 - 详细错误:', error)
    
    ElMessage.closeAll()
    
    let errorMessage = '导入失败: '
    
    if (error.message) {
      errorMessage += error.message
    } else {
      errorMessage += '未知错误'
    }
    
    ElMessage.error({
      message: errorMessage,
      duration: 5000,
      showClose: true
    })
  } finally {
    importing.value = false
  }
}

// 删除模型
const handleDeleteModel = async (modelName) => {
  if (!selectedDevice.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除模型 "${modelName}" 吗？此操作无法撤销。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    // 用户取消删除
    return
  }
  
  try {
    const url = `http://${getDeviceAddr(selectedDevice.value.ip)}/lm/local?name=${encodeURIComponent(modelName)}`
    const response = await fetch(url, {
      method: 'DELETE',
      headers: getAuthHeaders(selectedDevice.value.ip)
    })
    
    if (response.ok) {
      const data = await response.json()
      if (data.code === 0) {
        ElMessage.success(t('aiAssistant.deleteModelSuccess'))
        
        // 如果删除的是当前加载的模型，清空加载状态
        if (loadedModel.value?.name === modelName) {
          loadedModel.value = null
          chatMessages.value = []
        }
        
        // 如果删除的是当前选中的模型，清空选择
        if (selectedModelName.value === modelName) {
          selectedModelName.value = ''
          selectedModel.value = null
        }
        
        // 刷新模型列表
        await fetchModelList()
      } else {
        ElMessage.error(data.message || t('aiAssistant.deleteModelFailed'))
      }
    } else {
      ElMessage.error(t('aiAssistant.deleteModelFailed'))
    }
  } catch (error) {
    console.error('[handleDeleteModel] 删除模型失败:', error)
    ElMessage.error(t('aiAssistant.deleteModelFailed') + ': ' + error.message)
  }
}

// 启动LLM服务的核心逻辑(内部方法,可复用)
// 串行模式：先启动对话模型并等待就绪，再启动 Embedding 模型并等待就绪
const startLLMService = async (showMessage = true) => {
  if (!selectedDevice.value) {
    if (showMessage) ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return false
  }
  
  if (!selectedModelName.value || !selectedModel.value) {
    if (showMessage) ElMessage.warning(t('aiAssistant.selectModelFirst'))
    return false
  }
  
  try {
    // ========== 解析模型文件 ==========
    const parseModelFiles = (model) => {
      const files = model.files || []
      let modelPath = '', weightPath = '', model2Path = '', weight2Path = ''
      let vocabPath = '', embedPath = '', chatTemplateFile = ''
      files.forEach(file => {
        const filePath = file.filePath
        const fileName = filePath.split('/').pop()
        if (fileName.includes('.rknn') && !fileName.includes('vision')) modelPath = filePath
        else if (fileName.includes('.weight') && !fileName.includes('vision')) weightPath = filePath
        else if (fileName.includes('vision') && fileName.includes('.rknn')) model2Path = filePath
        else if (fileName.includes('vision') && fileName.includes('.weight')) weight2Path = filePath
        else if (fileName.endsWith('.gguf')) vocabPath = filePath
        else if (fileName.includes('.embed.bin')) embedPath = filePath
        else if (fileName.endsWith('.jinja')) chatTemplateFile = filePath
      })
      return { modelPath, weightPath, model2Path, weight2Path, vocabPath, embedPath, chatTemplateFile }
    }

    // ========== 构建对话模型配置 ==========
    const chatFiles = parseModelFiles(selectedModel.value)
    const isVLM = !!(chatFiles.model2Path || chatFiles.weight2Path)
    console.log('[startLLMService] 对话模型:', selectedModelName.value, '  isVLM:', isVLM)
    console.log('[startLLMService] 对话模型文件:', JSON.stringify(chatFiles))

    const missingFiles = []
    if (!chatFiles.modelPath) missingFiles.push('模型文件(.rknn)')
    if (!chatFiles.weightPath) missingFiles.push('权重文件(.weight)')
    if (!chatFiles.vocabPath) missingFiles.push('词表文件(.gguf)')
    if (!chatFiles.embedPath) missingFiles.push('嵌入文件(.embed.bin)')
    if (isVLM && !chatFiles.model2Path) missingFiles.push('视觉模型文件(vision_*.rknn)')
    if (isVLM && !chatFiles.weight2Path) missingFiles.push('视觉权重文件(vision_*.weight)')
    if (missingFiles.length > 0) {
      const errMsg = `对话模型文件不完整，缺少: ${missingFiles.join(', ')}`
      console.error('[startLLMService]', errMsg)
      if (showMessage) ElMessage.error(errMsg)
      return false
    }

    // ========== 启动模型 ==========
    const modelConfig = {
      host: "0.0.0.0",
      port: 8081,
      timeout: 30,
      models: {}
    }

    // 对话模型 embedding:false
    modelConfig.models[selectedModelName.value] = isVLM ? {
      alias: selectedModelName.value,
      model: chatFiles.modelPath,
      weight: chatFiles.weightPath,
      model2: chatFiles.model2Path,
      weight2: chatFiles.weight2Path,
      model3: "", weight3: "",
      vocab: chatFiles.vocabPath,
      embed: chatFiles.embedPath,
      "mel-filter": "", "ctx-size": 2048, "predict": -1,
      "temp": 0.8, "top-k": 1, "top-p": 0.8,
      "repeat-penalty": 1.1, "presence-penalty": 1.0, "frequency-penalty": 1.0,
      "img-start": "<|vision_start|>", "img-end": "<|vision_end|>", "img-content": "<|image_pad|>",
      "audio-start": "", "audio-end": "", "audio-content": "",
      "img-width": 392, "img-height": 392,
      "chat-template-file": chatFiles.chatTemplateFile,
      "embedding": false
    } : {
      alias: selectedModelName.value,
      model: chatFiles.modelPath,
      weight: chatFiles.weightPath,
      model2: "", weight2: "", model3: "", weight3: "",
      vocab: chatFiles.vocabPath,
      embed: chatFiles.embedPath,
      "mel-filter": "", "ctx-size": 2048, "predict": -1,
      "temp": 0.8, "top-k": 1, "top-p": 0.8,
      "repeat-penalty": 1.1, "presence-penalty": 1.0, "frequency-penalty": 1.0,
      "img-start": "", "img-end": "", "img-content": "",
      "audio-start": "", "audio-end": "", "audio-content": "",
      "img-width": 0, "img-height": 0,
      "chat-template-file": chatFiles.chatTemplateFile,
      "embedding": false
    }

    console.log('[startLLMService] 启动配置（models 数量:', Object.keys(modelConfig.models).length, '):', JSON.stringify(modelConfig, null, 2))
    const startResult = await AIStartModel({ deviceIp: selectedDevice.value.ip, modelConfig: modelConfig })
    console.log('[startLLMService] 启动响应:', JSON.stringify(startResult))

    if (!startResult?.success) {
      const msg = startResult?.message || '模型启动失败'
      console.error('[startLLMService] 启动失败:', msg, '  data:', JSON.stringify(startResult?.data))
      if (showMessage) ElMessage.error(msg)
      return false
    }

    // 等待对话模型就绪（最多 3 分钟，每 5 秒一次）
    console.log('[startLLMService] 等待对话模型就绪...')
    const chatReady = await pollModelStatus(selectedModelName.value, 36, 5000)
    if (!chatReady) {
      const errMsg = '对话模型启动超时（超过3分钟），请检查设备状态'
      console.error('[startLLMService]', errMsg)
      if (showMessage) ElMessage.error(errMsg)
      return false
    }
    console.log('[startLLMService] ✅ 对话模型已就绪:', selectedModelName.value)



    if (showMessage) ElMessage.success(t('aiAssistant.llmStartSuccess'))
    return true

  } catch (error) {
    console.error('[startLLMService] 异常:', error)
    if (showMessage) ElMessage.error(t('aiAssistant.llmStartFailed') + error.message)
    return false
  }
}

// 启动LLM服务(UI按钮调用)
const handleStartLLMService = async () => {
  if (!selectedDevice.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return
  }
  
  if (!selectedModelName.value || !selectedModel.value) {
    ElMessage.warning(t('aiAssistant.selectModelFirst'))
    return
  }
  
  try {
    startingService.value = true
    
    const result = await startLLMService(true)
    
    if (result) {
      // 标记模型为已加载
      loadedModel.value = selectedModel.value
      localOpenAIConfig.value.model = selectedModel.value?.name || localOpenAIConfig.value.model
    }
  } catch (error) {
    console.error('[handleStartLLMService] 启动LLM服务失败:', error)
    ElMessage.error(t('aiAssistant.llmStartFailed') + error.message)
  } finally {
    startingService.value = false
  }
}

// 停止LLM服务
const handleStopLLMService = async () => {
  if (!selectedDevice.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return
  }
  
  try {
    await ElMessageBox.confirm(
      '确定要停止LLM服务吗？',
      '确认停止',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    // 用户取消
    return
  }
  
  try {
    stoppingService.value = true
    
    const result = await AIStopModel(selectedDevice.value.ip)
    
    if (result?.success) {
      ElMessage.success(t('aiAssistant.llmStopSuccess'))
      // 清空加载状态
      loadedModel.value = null
      chatMessages.value = []
    } else {
      ElMessage.error(result?.message || t('aiAssistant.llmStopFailed'))
    }
  } catch (error) {
    console.error('[handleStopLLMService] 停止LLM服务失败:', error)
    ElMessage.error(t('aiAssistant.llmStopFailed') + ': ' + error.message)
  } finally {
    stoppingService.value = false
  }
}

// 重置设备
const handleResetDevice = async () => {
  if (!selectedDevice.value) {
    ElMessage.warning(t('aiAssistant.selectDeviceFirst'))
    return
  }
  
  try {
    await ElMessageBox.confirm(
      '重置设备将清除设备的硬件配置和状态，此操作不可恢复，确定要继续吗？',
      '警告',
      {
        confirmButtonText: '确定重置',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    )
  } catch {
    // 用户取消
    return
  }
  
  try {
    resettingDevice.value = true
    
    const result = await AIResetDevice(selectedDevice.value.ip)
    
    if (result?.success) {
      ElMessage.success(t('aiAssistant.deviceResetSuccess'))
      // 清空相关状态
      loadedModel.value = null
      chatMessages.value = []
      modelList.value = []
      selectedModel.value = null
      selectedModelName.value = ''
      // 关闭设置对话框
      showSettings.value = false
      // 刷新模型列表
      await fetchModelList()
      
      // 提示用户重新选择设备
      ElMessageBox.alert(
        '设备已重置，请重新选择模型并加载',
        '提示',
        {
          confirmButtonText: '知道了',
          type: 'success'
        }
      )
    } else {
      ElMessage.error(result?.message || t('aiAssistant.deviceResetFailed'))
    }
  } catch (error) {
    console.error('[handleResetDevice] 重置设备失败:', error)
    ElMessage.error(t('aiAssistant.deviceResetFailed') + ': ' + error.message)
  } finally {
    resettingDevice.value = false
  }
}

// 图片预览
const previewImage = (url) => {
  // 创建预览遮罩层
  const overlay = document.createElement('div')
  overlay.style.cssText = `
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    cursor: pointer;
  `
  
  const img = document.createElement('img')
  img.src = url
  img.style.cssText = `
    max-width: 90%;
    max-height: 90%;
    object-fit: contain;
  `
  
  overlay.appendChild(img)
  document.body.appendChild(overlay)
  
  overlay.onclick = () => {
    document.body.removeChild(overlay)
  }
}

defineExpose({
  fetchAiAssistant
})
</script>

<style scoped>
.ai-assistant {
  display: flex;
  height: 100%;
  gap: 16px;
  padding: 16px;
  overflow: hidden;
  background: #f8f9fa;
}

/* 左侧设备列表/配置 */
.device-list-section {
  width: 280px;
  flex-shrink: 0;
}

.device-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border-radius: 12px;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid #e8eaed;
}

.device-card :deep(.el-card__header) {
  background: #ffffff;
  border-bottom: 1px solid #e8eaed;
  padding: 16px 20px;
}

.device-card :deep(.el-card__header .card-header) {
  color: #303133;
}

.device-card :deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  font-size: 15px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

/* 设备列表 */
.device-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.device-item {
  padding: 14px 16px;
  margin-bottom: 10px;
  border: 1.5px solid #e8eaed;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: #fafbfc;
  position: relative;
  overflow: hidden;
}

.device-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 3px;
  height: 100%;
  background: #3b82f6;
  transform: scaleY(0);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.device-item:hover {
  border-color: #c3cad4;
  background: #ffffff;
  transform: translateX(4px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.device-item:hover::before {
  transform: scaleY(1);
}

.device-item.active {
  border-color: #3b82f6;
  background: #f0f5ff;
  box-shadow: 0 2px 12px rgba(59, 130, 246, 0.12);
}

.device-item.active::before {
  transform: scaleY(1);
}

.device-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-left: 4px;
}

.device-ip {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

/* 中间模型选择区域 */
.model-section {
  width: 320px;
  flex-shrink: 0;
}

.model-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border-radius: 12px;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid #e8eaed;
}

.model-card :deep(.el-card__header) {
  background: #ffffff;
  border-bottom: 1px solid #e8eaed;
  padding: 16px 20px;
}

.model-card :deep(.el-card__header .card-header) {
  color: #303133;
}

.model-card :deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 20px;
}

.model-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.model-selector-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.empty-hint {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.model-selector {
  flex: 1;
  overflow-y: auto;
}

.model-selector :deep(.el-form) {
  padding: 0;
}

.model-selector :deep(.el-form-item) {
  margin-bottom: 20px;
}

.model-selector :deep(.el-form-item__label) {
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
  font-size: 14px;
}

.model-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.model-option-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.model-option-actions {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 4px;
}

.model-option-actions :deep(.el-button) {
  width: 24px;
  height: 24px;
  padding: 0;
  font-size: 12px;
}

.model-option-actions :deep(.el-button):hover {
  transform: scale(1.1);
}

/* 右侧对话区域 */
.chat-section {
  flex: 1;
  min-width: 0;
}

.chat-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border-radius: 12px;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid #e8eaed;
}

.chat-card :deep(.el-card__header) {
  background: #ffffff;
  border-bottom: 1px solid #e8eaed;
  padding: 18px 28px;
}

.chat-card :deep(.el-card__header .card-header) {
  color: #303133;
  font-size: 16px;
}

.chat-card :deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fafbfc;
  position: relative;
}

.chat-content::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 100px;
  background: linear-gradient(180deg, rgba(59, 130, 246, 0.02) 0%, transparent 100%);
  pointer-events: none;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 32px 28px;
  scroll-behavior: smooth;
}

.empty-chat {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-container {
  text-align: center;
  color: #606266;
  animation: fadeIn 0.6s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.welcome-container h2 {
  margin: 24px 0 12px;
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.welcome-container p {
  font-size: 15px;
  color: #909399;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 36px;
}

.message-item {
  display: flex;
  gap: 16px;
  animation: slideInUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  margin-top: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s;
}

.message-item .message-avatar:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.message-item.user .message-avatar {
  background: #3b82f6;
  color: #fff;
}

.message-item.assistant .message-avatar {
  background: #f5f8ff;
  border: 2px solid #e0e7ff;
  color: #5b8def;
}

.message-wrapper {
  flex: 0 1 auto;
  max-width: 80%;
  min-width: 0;
}

.message-item.user .message-wrapper {
  max-width: fit-content;
  max-width: -moz-fit-content;
  flex: 0 0 auto;
}

.message-content {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid #e8eaed;
  width: fit-content;
  width: -moz-fit-content;
  max-width: 100%;
  min-width: 80px;
}

.message-item.assistant .message-content {
  border-top-left-radius: 6px;
  width: 100%;
}

.message-item.user .message-content {
  background: #3b82f6;
  color: #fff;
  border-top-right-radius: 6px;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.2);
  border: 1px solid #3b82f6;
  width: fit-content;
  width: -moz-fit-content;
  min-width: 80px;
}

.message-wrapper:hover .message-content {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.message-item.user:hover .message-content {
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.3);
}

/* 思考过程区域 */
.thinking-section {
  margin: 12px 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  background: #f9fafb;
}

.thinking-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
  background: #f3f4f6;
  font-size: 13px;
}

.thinking-header:hover {
  background: #e5e7eb;
}

.thinking-icon {
  flex-shrink: 0;
  color: #6b7280;
}

.thinking-status {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.thinking-active {
  color: #3b82f6;
}

.thinking-done {
  color: #6b7280;
}

.thinking-time {
  color: #9ca3af;
  font-size: 12px;
  font-weight: 400;
}

.expand-icon {
  flex-shrink: 0;
  color: #9ca3af;
  transition: transform 0.3s;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.thinking-content {
  padding: 12px 16px;
  font-size: 13px;
  line-height: 1.8;
  color: #4b5563;
  background: #ffffff;
  border-top: 1px solid #e5e7eb;
}

.thinking-tag {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  color: #dc2626;
  font-size: 12px;
  margin: 4px 0;
  opacity: 0.8;
}

.thinking-text {
  white-space: pre-wrap;
  word-wrap: break-word;
  padding: 8px 0;
  color: #374151;
}

.thinking-cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  background: #3b82f6;
  animation: blink-cursor 1s infinite;
  margin-left: 2px;
  vertical-align: text-bottom;
}

/* 消息文本 */
.message-text {
  padding: 16px;
  word-wrap: break-word;
  /* white-space: pre-wrap; */
  line-height: 1.9;
  font-size: 15px;
  color: #303133;
  font-weight: 400;
  letter-spacing: 0.2px;
}

.message-text :deep(p) {
  margin: 0.5em 0;
}

.message-text :deep(p:first-child) {
  margin-top: 0;
}

.message-text :deep(p:last-child) {
  margin-bottom: 0;
}


.message-item.user .message-text {
  color: rgba(255, 255, 255, 0.98);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

/* 光标闪烁动画 */
.cursor-blink {
  display: inline-block;
  width: 2px;
  height: 1.1em;
  background: #3b82f6;
  animation: blink-cursor 1.2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  margin-left: 2px;
  vertical-align: text-bottom;
  border-radius: 1px;
  box-shadow: 0 0 4px rgba(59, 130, 246, 0.3);
}

@keyframes blink-cursor {
  0%, 100% {
    opacity: 1;
    transform: scaleY(1);
  }
  50% {
    opacity: 0.3;
    transform: scaleY(0.9);
  }
}

/* 性能指标 */
.message-metrics {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 16px 16px 16px;
  font-size: 11px;
  color: #999;
  font-family: 'SF Mono', 'Consolas', 'Monaco', 'JetBrains Mono', monospace;
  border-top: 1px solid #f0f0f0;
  padding-top: 12px;
  margin: 8px 16px 0 16px;
}

.metric-item {
  display: flex;
  align-items: center;
  gap: 6px;
  background: #f8f9fa;
  padding: 4px 10px;
  border-radius: 12px;
  font-weight: 600;
  transition: all 0.2s;
}

.metric-item:hover {
  background: #e9ecef;
  transform: translateY(-1px);
}

/* 操作按钮 */
.message-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  opacity: 0;
  transition: all 0.3s;
  transform: translateY(-4px);
}

.message-wrapper:hover .message-actions {
  opacity: 1;
  transform: translateY(0);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  font-size: 12px;
  color: #6b7280;
  background: #ffffff;
  border: 1.5px solid #e5e7eb;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.action-btn:hover {
  color: #3b82f6;
  background: #f5f8ff;
  border-color: #bfdbfe;
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.12);
}

.action-btn:active {
  transform: scale(0.95) translateY(-1px);
}

.action-btn svg {
  flex-shrink: 0;
}

/* 输入区域 */
.chat-input-area {
  flex-shrink: 0;
  padding: 20px 28px 24px;
  background: #ffffff;
  border-top: 1px solid #e8eaed;
}

/* 上传文件预览区域 */
.uploaded-files-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 12px;
}

.file-preview-item {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid #e8eaed;
  background: #fff;
}

.image-preview,
.video-preview {
  position: relative;
  width: 150px;
  height: 150px;
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-preview video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.file-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.8), transparent);
  padding: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.file-name {
  font-size: 12px;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.remove-btn {
  background: rgba(255, 77, 79, 0.9) !important;
  border: none !important;
  color: #fff !important;
  font-size: 20px !important;
  font-weight: bold !important;
  min-width: 24px !important;
  width: 24px !important;
  height: 24px !important;
  padding: 0 !important;
}

.remove-btn:hover {
  background: rgba(255, 77, 79, 1) !important;
}

/* 消息中的文件显示 */
.message-files {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.message-file-item {
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e8eaed;
  background: #f8f9fa;
}

.message-image {
  max-width: 300px;
  max-height: 300px;
  display: block;
  cursor: pointer;
  transition: transform 0.2s;
}

.message-image:hover {
  transform: scale(1.02);
}

.message-video {
  max-width: 400px;
  max-height: 300px;
  display: block;
}

.input-wrapper {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0;
  background: #ffffff;
  border-radius: 16px;
  border: 1.5px solid #e0e3e8;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: none;
  overflow: hidden;
}

.input-wrapper:hover {
  border-color: #b4b9c1;
}

.input-wrapper:focus-within {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08);
}

.input-wrapper.disabled {
  background: #f8f9fa;
  border-color: #e0e3e8;
  cursor: not-allowed;
}

/* 工具栏样式 */
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-bottom: 1px solid #e8eaed;
  flex-wrap: wrap;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  color: #606266;
  transition: all 0.2s;
  padding: 0;
}

.toolbar-btn:hover:not(:disabled) {
  background: #e8eaed;
  color: #303133;
}

.toolbar-btn:active:not(:disabled) {
  background: #d0d7de;
}

.toolbar-btn:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.toolbar-icon {
  font-weight: 600;
  font-size: 14px;
  font-family: Arial, sans-serif;
}

.toolbar-icon.italic {
  font-style: italic;
}

.toolbar-icon.underline {
  text-decoration: underline;
}

.toolbar-icon.strikethrough {
  text-decoration: line-through;
}

.toolbar-divider {
  width: 1px;
  height: 20px;
  background: #d0d7de;
  margin: 0 4px;
}

/* 富文本编辑器样式 */
.rich-editor {
  flex: 1;
  min-height: 60px;
  max-height: 200px;
  overflow-y: auto;
  padding: 14px 56px 14px 16px; /* 右侧留出发送按钮空间 */
  font-size: 15px;
  line-height: 1.6;
  color: #303133;
  outline: none;
  word-wrap: break-word;
  word-break: break-word;
  transition: all 0.3s ease;
  position: relative;
}

/* 拖拽时的样式 */
.rich-editor.dragging {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 2px dashed #409eff !important;
  border-radius: 6px;
}

.rich-editor.dragging::before {
  display: none !important;
}

.rich-editor.dragging::after {
  content: '📁 松开以上传文件';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: #409eff;
  font-size: 14px;
  font-weight: 600;
  pointer-events: none;
  white-space: nowrap;
  z-index: 10;
}

.rich-editor:empty:before {
  content: attr(data-placeholder);
  color: #adb5bd;
  pointer-events: none;
}

.rich-editor:focus {
  outline: none;
}

/* 富文本内容样式 */
.rich-editor b,
.rich-editor strong {
  font-weight: 700;
}

.rich-editor i,
.rich-editor em {
  font-style: italic;
}

.rich-editor u {
  text-decoration: underline;
}

.rich-editor s,
.rich-editor strike {
  text-decoration: line-through;
}

.rich-editor ul,
.rich-editor ol {
  margin: 8px 0;
  padding-left: 24px;
}

.rich-editor li {
  margin: 4px 0;
}

.rich-editor p {
  margin: 4px 0;
}

/* 编辑器滚动条 */
.rich-editor::-webkit-scrollbar {
  width: 6px;
}

.rich-editor::-webkit-scrollbar-thumb {
  background: #d0d7de;
  border-radius: 10px;
}

.rich-editor::-webkit-scrollbar-thumb:hover {
  background: #a8b0b8;
}

.rich-editor::-webkit-scrollbar-track {
  background: transparent;
}

/* 发送按钮容器 */
.input-wrapper > .input-actions {
  position: absolute;
  right: 12px;
  bottom: 12px;
}

.input-actions {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.input-actions :deep(.el-button) {
  width: 40px;
  height: 40px;
  background: #3b82f6;
  border: none;
  box-shadow: none;
  transition: all 0.2s ease;
  padding: 0;
}

.input-actions :deep(.el-button:hover:not(:disabled)) {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.input-actions :deep(.el-button:active:not(:disabled)) {
  transform: translateY(0);
  box-shadow: none;
}

.input-actions :deep(.el-button:disabled) {
  background: #e9ecef;
  color: #adb5bd;
  cursor: not-allowed;
}

.input-actions :deep(.el-button.is-loading) {
  background: #e9ecef;
  color: #adb5bd;
}

/* 滚动条样式 */
.device-list::-webkit-scrollbar,
.model-list::-webkit-scrollbar,
.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.device-list::-webkit-scrollbar-thumb,
.model-list::-webkit-scrollbar-thumb,
.chat-messages::-webkit-scrollbar-thumb {
  background: #d0d7de;
  border-radius: 10px;
}

.device-list::-webkit-scrollbar-thumb:hover,
.model-list::-webkit-scrollbar-thumb:hover,
.chat-messages::-webkit-scrollbar-thumb:hover {
  background: #a8b0b8;
}

.device-list::-webkit-scrollbar-track,
.model-list::-webkit-scrollbar-track,
.chat-messages::-webkit-scrollbar-track {
  background: #f5f6f7;
  border-radius: 10px;
}

/* 模型文件列表样式 */
.model-files-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e8eaed;
  transition: all 0.2s;
}

.file-item:hover {
  background: #f1f3f5;
  border-color: #d0d7de;
}

.file-name {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
  word-break: break-all;
  flex: 1;
}

.file-size {
  font-size: 12px;
  color: #909399;
  margin-left: 12px;
  white-space: nowrap;
}

/* 设置对话框样式 */
.settings-section {
  padding: 20px;
}

.settings-section :deep(.el-form-item__label) {
  font-weight: 500;
  color: #303133;
}

.settings-section :deep(.el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.settings-section :deep(.el-alert) {
  border-radius: 8px;
}

.settings-section :deep(.el-alert__title) {
  font-size: 14px;
}

/* 代码块样式 */
.message-text :deep(.code-block-wrapper) {
  margin: 12px 0;
  border-radius: 6px;
  overflow: hidden;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.message-text :deep(.code-block-wrapper:first-child) {
  margin-top: 0;
}

.message-text :deep(.code-block-wrapper:last-child) {
  margin-bottom: 0;
}

.message-text :deep(.code-block-header) {
  display: flex;
  /* justify-content: space-between; */
  /* align-items: center; */
  padding: 10px 16px;
  background: #f1f3f5;
  border-bottom: 1px solid #dee2e6;
}

.message-text :deep(.code-language) {
  font-size: 11px;
  color: #495057;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
}

.message-text :deep(.code-actions) {
  display: flex;
  gap: 4px;
  align-items: center;
}

.message-text :deep(.code-action-btn) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 26px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  color: #6c757d;
  transition: all 0.2s ease;
  padding: 0;
  position: relative;
}

.message-text :deep(.code-action-btn:hover) {
  background: #e9ecef;
  color: #212529;
}

.message-text :deep(.code-action-btn:active) {
  background: #dee2e6;
  transform: scale(0.95);
}

.message-text :deep(.code-action-btn svg) {
  width: 16px;
  height: 16px;
}

/* 复制成功提示 */
.message-text :deep(.code-action-btn.copied)::after {
  content: '已复制';
  position: absolute;
  top: -30px;
  right: 0;
  background: #212529;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  white-space: nowrap;
  animation: fadeOut 2s forwards;
}

@keyframes fadeOut {
  0%, 50% { opacity: 1; }
  100% { opacity: 0; }
}

.message-text :deep(.code-block) {
  margin: 0;
  padding: 16px 0;
  background: #ffffff;
  overflow-x: auto;
  font-family: 'Consolas', 'Monaco', 'Courier New', 'Menlo', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #24292e;
}

.message-text :deep(.code-block code) {
  display: block;
  padding: 0;
  background: transparent;
  color: inherit;
  font-family: inherit;
}

.message-text :deep(.code-line) {
  display: flex;
  min-height: 20px;
  padding: 0 16px;
  transition: background-color 0.1s ease;
}

.message-text :deep(.code-line:hover) {
  background-color: #f8f9fa;
}

.message-text :deep(.line-number) {
  display: inline-block;
  width: 42px;
  min-width: 42px;
  padding-right: 16px;
  text-align: right;
  color: #adb5bd;
  user-select: none;
  flex-shrink: 0;
  font-size: 12px;
}

.message-text :deep(.line-content) {
  flex: 1;
  white-space: pre;
  overflow: visible;
}

/* 代码高亮颜色 */
.message-text :deep(.hljs-keyword) {
  color: #cf222e;
  font-weight: 600;
}

.message-text :deep(.hljs-string) {
  color: #0a3069;
}

.message-text :deep(.hljs-number) {
  color: #0550ae;
}

.message-text :deep(.hljs-comment) {
  color: #57606a;
  font-style: italic;
}

.message-text :deep(.hljs-literal) {
  color: #0550ae;
}

.message-text :deep(.hljs-function) {
  color: #8250df;
  font-weight: 500;
}

/* 行内代码 */
.message-text :deep(.inline-code) {
  padding: 2px 6px;
  background: #eff1f3;
  border: 1px solid #d0d7de;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', 'Courier New', 'Menlo', monospace;
  font-size: 85%;
  color: #cf222e;
}

/* 代码块滚动条 */
.message-text :deep(.code-block::-webkit-scrollbar) {
  height: 8px;
}

.message-text :deep(.code-block::-webkit-scrollbar-thumb) {
  background: #d1d5da;
  border-radius: 4px;
}

.message-text :deep(.code-block::-webkit-scrollbar-thumb:hover) {
  background: #959da5;
}

.message-text :deep(.code-block::-webkit-scrollbar-track) {
  background: #f6f8fa;
  border-radius: 4px;
}

/* 消息文本内的 HTML 样式 */
.message-text :deep(br) {
  display: block;
  content: "";
  margin-top: 0.5em;
}

/* === Markdown-it 生成的代码块样式（参考222.html）=== */
.message-text :deep(.code-wrapper) {
  margin: 16px 0;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #374151;
  background: #282c34;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.message-text :deep(.code-header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: #21252b;
  border-bottom: 1px solid #374151;
}

.message-text :deep(.code-header .code-language) {
  color: #9ca3af;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  font-weight: bold;
}

.message-text :deep(.copy-btn) {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #9ca3af;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.message-text :deep(.copy-btn:hover) {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

.message-text :deep(.code-pre) {
  margin: 0 !important;
  padding: 16px !important;
  background: transparent !important;
  overflow-x: auto;
}

.message-text :deep(.code-pre code) {
  font-family: 'Consolas', 'Monaco', 'Courier New', 'JetBrains Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
}

/* Highlight.js 主题调整 */
.message-text :deep(.hljs) {
  background: transparent !important;
  color: #abb2bf;
}

/* Markdown-it 生成的行内代码 */
.message-text :deep(p code:not(.hljs)) {
  background-color: #f1f5f9;
  color: #ec4899;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 0.9em;
}






/* ========== 新样式 - 来自 aiAssistantNew.vue ========== */
.message-item-new {
  margin-bottom: 24px;
}

/* 用户消息 */
.user-msg-new {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  justify-content: flex-end;
  animation: slideInUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.user-avatar-new {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea, #764ba2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
  order: 2;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
  transition: all 0.3s;
}

.user-avatar-new:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

.user-content-new {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  padding: 16px 20px;
  border-radius: 18px 18px 4px 18px;
  max-width: 65%;
  word-wrap: break-word;
  box-shadow: 0 4px 14px rgba(102, 126, 234, 0.35);
  font-size: 15px;
  line-height: 1.7;
  transition: all 0.3s;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.user-content-new:hover {
  box-shadow: 0 6px 18px rgba(102, 126, 234, 0.45);
  transform: translateY(-2px);
}

/* 消息中的图片 */
.msg-images-new {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.msg-image-new {
  max-width: 200px;
  max-height: 200px;
  border-radius: 8px;
  object-fit: cover;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid rgba(255, 255, 255, 0.3);
}

.msg-image-new:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

/* AI 消息 */
.ai-msg-new {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  animation: slideInUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(15px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.ai-avatar-new {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  transition: all 0.3s;
}

.ai-avatar-new:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.ai-content-wrap-new {
  flex: 1;
  background: white;
  border-radius: 18px 18px 18px 4px;
  padding: 22px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  max-width: 82%;
  transition: all 0.3s;
  border: 1px solid rgba(0, 0, 0, 0.05);
}

.ai-content-wrap-new:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
  border-color: rgba(59, 130, 246, 0.2);
}

/* 思考过程 */
.think-block-new {
  margin-bottom: 20px;
  border: 1px solid #dbeafe;
  border-radius: 12px;
  overflow: hidden;
  background: linear-gradient(135deg, #eff6ff 0%, #f0f9ff 100%);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.12);
  transition: all 0.3s;
}

.think-block-new:hover {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.18);
  border-color: #bfdbfe;
}

.think-header-new {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: linear-gradient(135deg, #dbeafe 0%, #e0f2fe 100%);
  cursor: pointer;
  user-select: none;
  font-size: 14px;
  color: #1e40af;
  font-weight: 700;
  transition: all 0.25s;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
}

.think-header-new:hover {
  background: linear-gradient(135deg, #bfdbfe 0%, #dbeafe 100%);
  padding-left: 20px;
}

.think-header-new:active {
  transform: scale(0.98);
}

.think-icon-new {
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 11px;
  color: #3b82f6;
  font-weight: bold;
}

.think-icon-new.expanded {
  transform: rotate(90deg);
}

.think-time-new {
  margin-left: auto;
  color: #3b82f6;
  font-size: 12px;
  font-weight: 600;
  background: white;
  padding: 4px 10px;
  border-radius: 14px;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.think-content-new {
  padding: 18px;
  background: white;
  font-size: 14.5px;
  color: #475569;
  line-height: 1.85;
  border-top: 1px solid #e0f2fe;
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.think-content-new :deep(p) {
  margin: 0.7em 0;
}

.think-content-new :deep(p:first-child) {
  margin-top: 0;
}

.think-content-new :deep(p:last-child) {
  margin-bottom: 0;
}

/* 正文内容 */
.ai-main-content-new {
  font-size: 15.5px;
  line-height: 1.85;
  color: #1e293b;
  word-wrap: break-word;
  word-break: break-word;
}

.ai-main-content-new.typing::after {
  content: '▋';
  animation: blink-new 1s infinite;
  margin-left: 3px;
  color: #3b82f6;
  font-weight: bold;
}

@keyframes blink-new {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* Markdown 元素样式优化 */
.ai-main-content-new :deep(p) {
  margin: 0.9em 0;
  line-height: 1.85;
}

.ai-main-content-new :deep(p:first-child) {
  margin-top: 0;
}

.ai-main-content-new :deep(p:last-child) {
  margin-bottom: 0;
}

.ai-main-content-new :deep(ul),
.ai-main-content-new :deep(ol) {
  margin: 1.1em 0;
  padding-left: 2em;
}

.ai-main-content-new :deep(li) {
  margin: 0.7em 0;
  line-height: 1.85;
}

.ai-main-content-new :deep(li > p) {
  margin: 0.5em 0;
}

.ai-main-content-new :deep(ul > li) {
  list-style-type: disc;
}

.ai-main-content-new :deep(ul > li::marker) {
  color: #3b82f6;
  font-size: 1em;
}

.ai-main-content-new :deep(ol > li) {
  list-style-type: decimal;
}

.ai-main-content-new :deep(ol > li::marker) {
  color: #3b82f6;
  font-weight: 700;
}

.ai-main-content-new :deep(strong) {
  color: #0f172a;
  font-weight: 700;
}

.ai-main-content-new :deep(em) {
  color: #475569;
  font-style: italic;
}

.ai-main-content-new :deep(h1),
.ai-main-content-new :deep(h2),
.ai-main-content-new :deep(h3),
.ai-main-content-new :deep(h4) {
  margin: 1.4em 0 0.7em 0;
  font-weight: 700;
  line-height: 1.4;
  color: #0f172a;
}

.ai-main-content-new :deep(h1) { 
  font-size: 1.75em;
  border-bottom: 2px solid #e5e7eb;
  padding-bottom: 0.3em;
}

.ai-main-content-new :deep(h2) { 
  font-size: 1.5em;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 0.25em;
}

.ai-main-content-new :deep(h3) { 
  font-size: 1.3em; 
}

.ai-main-content-new :deep(h4) { 
  font-size: 1.15em; 
}

.ai-main-content-new :deep(blockquote) {
  margin: 1.2em 0;
  padding: 1em 1.2em;
  border-left: 4px solid #3b82f6;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #475569;
  font-style: italic;
  border-radius: 0 8px 8px 0;
}

.ai-main-content-new :deep(hr) {
  margin: 2em 0;
  border: none;
  border-top: 2px solid #e2e8f0;
}

/* 性能指标 */
.metrics-bar-new {
  display: flex;
  gap: 18px;
  margin-top: 18px;
  padding-top: 14px;
  border-top: 1px solid #e5e7eb;
  font-size: 13px;
  color: #64748b;
}

.metrics-bar-new span {
  display: flex;
  align-items: center;
  gap: 5px;
  font-weight: 600;
  background: #f8fafc;
  padding: 5px 10px;
  border-radius: 8px;
  transition: all 0.2s;
}

.metrics-bar-new span:hover {
  background: #f1f5f9;
  color: #3b82f6;
  transform: translateY(-1px);
}

/* 代码块样式 */
.ai-main-content-new :deep(.code-block-wrapper) {
  margin: 20px 0;
  border-radius: 12px;
  overflow: hidden;
  background: #282c34;
  border: 1px solid #3d4451;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  transition: all 0.3s;
}

.ai-main-content-new :deep(.code-block-wrapper:hover) {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.25);
  border-color: #4b5563;
}

.ai-main-content-new :deep(.code-block-header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #1e2127 0%, #252931 100%);
  border-bottom: 1px solid #3d4451;
}

.ai-main-content-new :deep(.code-lang) {
  color: #61afef;
  font-size: 13px;
  font-family: 'Consolas', 'Monaco', 'JetBrains Mono', monospace;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
  background: rgba(97, 175, 239, 0.15);
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid rgba(97, 175, 239, 0.3);
}

.ai-main-content-new :deep(.copy-btn) {
  color: #abb2bf;
  background: rgba(171, 178, 191, 0.08);
  border: 1px solid rgba(171, 178, 191, 0.2);
  cursor: pointer;
  font-size: 12px;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.2s;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 4px;
}

.ai-main-content-new :deep(.copy-btn:hover) {
  color: #61afef;
  background: rgba(97, 175, 239, 0.15);
  border-color: #61afef;
  transform: translateY(-1px);
}

.ai-main-content-new :deep(.copy-btn:active) {
  transform: translateY(0);
}

.ai-main-content-new :deep(.code-block) {
  margin: 0 !important;
  padding: 20px !important;
  background: #282c34 !important;
  overflow-x: auto;
  font-size: 14px;
  line-height: 1.8;
  font-family: 'Consolas', 'Monaco', 'JetBrains Mono', 'Courier New', monospace;
}

.ai-main-content-new :deep(.code-block::-webkit-scrollbar) {
  height: 8px;
}

.ai-main-content-new :deep(.code-block::-webkit-scrollbar-track) {
  background: #1e2127;
  border-radius: 4px;
}

.ai-main-content-new :deep(.code-block::-webkit-scrollbar-thumb) {
  background: #4b5563;
  border-radius: 4px;
}

.ai-main-content-new :deep(.code-block::-webkit-scrollbar-thumb:hover) {
  background: #6b7280;
}

.ai-main-content-new :deep(.hljs) {
  background: #282c34 !important;
  color: #abb2bf !important;
}

.ai-main-content-new :deep(pre) {
  margin: 0;
  background: #282c34 !important;
}

.ai-main-content-new :deep(pre code) {
  display: block;
  background: transparent !important;
  padding: 0 !important;
}

/* 行内代码 */
.ai-main-content-new :deep(p code:not(.hljs)),
.ai-main-content-new :deep(li code:not(.hljs)),
.ai-main-content-new :deep(h1 code:not(.hljs)),
.ai-main-content-new :deep(h2 code:not(.hljs)),
.ai-main-content-new :deep(h3 code:not(.hljs)),
.ai-main-content-new :deep(h4 code:not(.hljs)) {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #92400e;
  padding: 0.2em 0.6em;
  border-radius: 6px;
  font-size: 0.90em;
  font-family: 'Consolas', 'Monaco', 'JetBrains Mono', monospace;
  font-weight: 600;
  border: 1px solid #fbbf24;
  box-shadow: 0 1px 3px rgba(251, 191, 36, 0.25);
  transition: all 0.2s;
}

.ai-main-content-new :deep(p code:not(.hljs):hover),
.ai-main-content-new :deep(li code:not(.hljs):hover) {
  background: linear-gradient(135deg, #fde68a 0%, #fcd34d 100%);
  box-shadow: 0 2px 5px rgba(251, 191, 36, 0.35);
  transform: translateY(-1px);
}

/* 表格样式 */
.ai-main-content-new :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5em 0;
  font-size: 0.95em;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border-radius: 8px;
  overflow: hidden;
}

.ai-main-content-new :deep(thead) {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}

.ai-main-content-new :deep(th) {
  padding: 12px 16px;
  text-align: left;
  font-weight: 700;
  border-bottom: 2px solid #1e40af;
}

.ai-main-content-new :deep(td) {
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
}

.ai-main-content-new :deep(tbody tr:hover) {
  background: #f8fafc;
}

.ai-main-content-new :deep(tbody tr:last-child td) {
  border-bottom: none;
}



/* ===== 经验库反馈条 ===== */
.feedback-bar {
  margin: 0 0 8px 0;
  border-radius: 10px;
  background: linear-gradient(135deg, #1e3a5f 0%, #1a2f4e 100%);
  border: 1px solid #3b82f6;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.25);
}

.feedback-bar-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  gap: 12px;
}

.feedback-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.feedback-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.feedback-text {
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 500;
}

.feedback-countdown {
  color: #64748b;
  font-size: 12px;
  flex-shrink: 0;
}

.feedback-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.feedback-btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s ease;
}

.feedback-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.feedback-btn.useful {
  background: #16a34a;
  color: white;
}
.feedback-btn.useful:hover:not(:disabled) {
  background: #15803d;
  transform: translateY(-1px);
}

.feedback-btn.useless {
  background: #374151;
  color: #d1d5db;
  border: 1px solid #4b5563;
}
.feedback-btn.useless:hover:not(:disabled) {
  background: #4b5563;
}

.feedback-btn.dismiss {
  background: transparent;
  color: #6b7280;
  padding: 6px 8px;
  font-size: 14px;
}
.feedback-btn.dismiss:hover {
  color: #9ca3af;
}

.feedback-progress {
  height: 3px;
  background: rgba(255,255,255,0.08);
}

.feedback-progress-bar {
  height: 100%;
  background: #3b82f6;
  transition: width 1s linear;
}

/* 滑入动画 */
.feedback-slide-enter-active {
  transition: all 0.3s ease;
}
.feedback-slide-leave-active {
  transition: all 0.2s ease;
}
.feedback-slide-enter-from,
.feedback-slide-leave-to {
  opacity: 0;
  transform: translateY(8px);
}

/* ===== 经验库管理面板 ===== */
.knowledge-panel {
  padding: 4px 0;
}

.knowledge-stats {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.kstat-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 8px;
  border-radius: 8px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}
.kstat-item.useful { background: #f0fdf4; border-color: #bbf7d0; }
.kstat-item.useless { background: #fff7ed; border-color: #fed7aa; }
.kstat-item.vector { background: #eff6ff; border-color: #bfdbfe; }

.kstat-num {
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
  line-height: 1;
}
.kstat-label {
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
}

.knowledge-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.knowledge-actions {
  display: flex;
  gap: 6px;
  align-items: center;
}

.knowledge-list {
  max-height: 360px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.knowledge-item {
  border-radius: 8px;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  background: #fff;
}
.knowledge-item.useful { border-left: 3px solid #22c55e; }
.knowledge-item.useless { border-left: 3px solid #f97316; }

.ki-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.ki-tag {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
}
.ki-tag.useful { background: #dcfce7; color: #16a34a; }
.ki-tag.useless { background: #ffedd5; color: #ea580c; }

.ki-vector {
  font-size: 11px;
  color: #3b82f6;
  background: #eff6ff;
  padding: 2px 6px;
  border-radius: 4px;
}

.ki-time {
  font-size: 11px;
  color: #94a3b8;
}

.ki-question {
  font-size: 13px;
  color: #1e293b;
  font-weight: 500;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ki-answer {
  font-size: 12px;
  color: #64748b;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

</style>


