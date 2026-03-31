<template> 
  <div class="customer-service">
    <!-- 未登录提示 -->
    <div v-if="!isLoggedIn" class="login-required">
      <el-empty description="">
        <template #description>
          <div style="margin-bottom: 20px; font-size: 14px; color: #909399;">
            请登录后使用客服功能
          </div>
          <div style="display: flex; gap: 10px; justify-content: center;">
            <el-button type="primary" @click="emit('showSyncAuthDialog')">登录</el-button>
            <el-button @click="emit('showRegisterDialog')">注册</el-button>
          </div>
        </template>
      </el-empty>
    </div>
    <!-- 已登录：iframe 客服窗口 -->
    <div v-else class="customer-service-content">
      <!-- iframeKey 变化时强制销毁并重建 iframe -->
      <iframe
        v-if="customerServiceUrl"
        ref="customerServiceFrame"
        @load="handleIframeLoad"
        :key="iframeKey"
        :src="customerServiceUrl"
        frameborder="0"
        class="customer-service-iframe"
      ></iframe>
    </div>
  </div>
</template>

<script setup> 
  import { ref, computed, onMounted, onUnmounted, onActivated } from 'vue'
  import { Call } from '@wailsio/runtime'

  const CHAT_ORIGIN = 'https://chat.moyunteng.com'

  // iframe key：每次自增，强制 Vue 销毁并重建 iframe
  const iframeKey = ref(0)
  const customerServiceFrame = ref(null)
  
  // 定义 props
  const props = defineProps({
    devices: {
      type: Array,
      default: () => []
    },
    deviceFirmwareInfo: {
      type: Map,
      default: () => new Map()
    },
    devicesStatusCache: {
      type: Map,
      default: () => new Map()
    },
    deviceVersionInfo: {
      type: Map,
      default: () => new Map()
    },
    token: {
      type: String,
      default: ''
    }
  })
  
  // 定义 emits
  const emit = defineEmits(['unreadCountChange', 'showSyncAuthDialog', 'showRegisterDialog'])

  // 登录状态（与 instanceManagement.vue 保持一致）
  const isLoggedIn = computed(() => !!props.token || !!localStorage.getItem('token'))
  
  // 从 localStorage 获取用户信息
  const uid = localStorage.getItem('uid') || ''
  const uname = localStorage.getItem('uname') || ''

  // 客服 URL（ticket 生成后填入）
  const customerServiceUrl = ref('')
  // 缓存最近一次的资产参数，供 ticket 刷新后复用
  let cachedAssetsParam = ''
  // 当前客服页是否可见
  const isChatPageVisible = ref(false)
  // 首次挂载标志，防止 onMounted + onActivated 同时触发导致调用两次
  let isMounted = false
  // 单飞控制：同一时刻只发起一次 ticket 请求
  let inflightTicketPromise = null

  // 调用 Go 后端生成 JWT ticket，拼接客服 URL（自动携带已缓存的 assets）
  const initCustomerServiceUrl = async () => {
    if (inflightTicketPromise) {
      return inflightTicketPromise
    }

    inflightTicketPromise = (async () => {
      try {
        const result = await Call.ByName('main.App.GenerateChatTicket', uid, uname)
        if (result && result.success && result.ticket) {
          const base = `${CHAT_ORIGIN}/?ticket=${result.ticket}`
          customerServiceUrl.value = cachedAssetsParam ? `${base}&assets=${cachedAssetsParam}` : base
          console.log('客服 ticket 生成成功（有效期 300s）', result.ticket)
        } else {
          throw new Error(result?.error || '生成失败')
        }
      } catch (error) {
        console.error('生成 ticket 失败，降级使用 user_id/user_name:', error)
        const base = `${CHAT_ORIGIN}/?user_id=${uid}&user_name=${uname}`
        customerServiceUrl.value = cachedAssetsParam ? `${base}&assets=${cachedAssetsParam}` : base
      } finally {
        inflightTicketPromise = null
      }
    })()

    return inflightTicketPromise
  }

  // 防重入标志：避免 token 过期时并发触发多次刷新
  let isRefreshingTicket = false

  const postTicketToIframe = (ticket) => {
    const targetWindow = customerServiceFrame.value?.contentWindow
    if (!targetWindow || !ticket) return false

    targetWindow.postMessage({
      type: 'chat:new-ticket',
      ticket
    }, '*')
    console.log('[customerService] 已向 iframe 发送新 ticket')
    return true
  }

  const postPageVisibleToIframe = (visible) => {
    const targetWindow = customerServiceFrame.value?.contentWindow
    if (!targetWindow) return false

    targetWindow.postMessage({
      type: 'chat:page-visible',
      visible
    }, '*')
    console.log('[customerService] 已通知 iframe 页面可见状态:', visible)
    return true
  }

  const syncPageVisibleToIframe = () => {
    postPageVisibleToIframe(isChatPageVisible.value)
  }

  const setPageVisible = (visible) => {
    isChatPageVisible.value = !!visible
    syncPageVisibleToIframe()

    if (isChatPageVisible.value) {
      unreadCount.value = 0
      emit('unreadCountChange', 0)
      console.log('[customerService] 客服页可见，清空宿主未读角标')
    }
  }

  const handleIframeLoad = () => {
    syncPageVisibleToIframe()
  }

  // 重新生成 ticket，并优先通过 postMessage 交给 iframe 自行续期
  const refreshTicketForIframe = async () => {
    if (isRefreshingTicket) return
    isRefreshingTicket = true
    console.log('[customerService] 收到 refresh-ticket，重新生成 ticket...')
    try {
      const result = await Call.ByName('main.App.GenerateChatTicket', uid, uname)
      if (result && result.success && result.ticket) {
        const delivered = postTicketToIframe(result.ticket)
        if (!delivered) {
          await initCustomerServiceUrl()
          iframeKey.value++
          console.log('[customerService] iframe 不可用，回退为重建 iframe')
        }
      } else {
        throw new Error(result?.error || '生成失败')
      }
    } finally {
      isRefreshingTicket = false
    }
  }

  // 监听 iframe postMessage（客服页面发出 token_expired 通知）
  const handleIframeMessage = (event) => {
    // 只处理来自客服域的消息
    if (event.origin !== CHAT_ORIGIN) {
      console.log('[customerService] 忽略非客服域消息:', event.origin, event.data)
      return
    }
    const data = event.data
    if (!data) return
    // 支持多种续票消息格式：
    //   { type: 'token_expired' }
    //   { event: 'token_expired' }
    //   { type: 'chat:refresh-ticket' }
    const msgType = (typeof data === 'object') ? (data.type || data.event) : data
    if (msgType === 'token_expired' || msgType === 'chat:refresh-ticket') {
      refreshTicketForIframe()
      return
    }

    if (msgType === 'chat:unread-summary') {
      if (isChatPageVisible.value) {
        unreadCount.value = 0
        emit('unreadCountChange', 0)
        console.log('[customerService] 客服页可见，忽略未读累计')
        return
      }
      const incomingVisitorID = data.visitor_id == null ? '' : String(data.visitor_id)
      const currentVisitorID = uid == null ? '' : String(uid)
      if (incomingVisitorID && currentVisitorID && incomingVisitorID !== currentVisitorID) {
        console.log('[customerService] 忽略其他访客未读消息:', incomingVisitorID, currentVisitorID)
        return
      }
      const nextUnreadCount = Number(data.total_unread)
      if (!Number.isNaN(nextUnreadCount)) {
        unreadCount.value = nextUnreadCount
        emit('unreadCountChange', unreadCount.value)
        console.log('[customerService] 收到未读消息推送:', data.total_unread)
      } else {
        console.log('[customerService] 未读消息格式无效:', data)
      }
    }
  }

  const notifyIframeShutdown = () => {
    try {
      customerServiceFrame.value?.contentWindow?.postMessage({ type: 'chat:shutdown' }, '*')
      console.log('[customerService] 已通知 iframe 执行 shutdown')
    } catch (error) {
      console.error('[customerService] 通知 iframe shutdown 失败:', error)
    }
  }
  
  // 未读消息数量
  const unreadCount = ref(0)
  
  // 构建资产信息并拼接到客服 URL
  const reportUserAssets = async () => {
    try {
      console.log('开始构建用户资产信息')
      console.log('设备总数:', props.devices.length)
      
      // 获取在线设备列表
      const onlineDevices = props.devices.filter(device => {
        const status = props.devicesStatusCache.get(device.id) || 'online'
        return status === 'online'
      })
      
      console.log('在线设备数:', onlineDevices.length)
      
      // 检查有多少设备已经有固件信息和版本信息
      const devicesWithInfo = onlineDevices.filter(device => {
        const hasFirmwareInfo = props.deviceFirmwareInfo.has(device.id)
        const hasVersionInfo = props.deviceVersionInfo.has(device.id)
        return hasFirmwareInfo && hasVersionInfo
      })
      
      console.log('已有完整信息的设备数:', devicesWithInfo.length)
      
      // 如果没有任何设备有完整信息，延迟一下再试
      if (devicesWithInfo.length === 0 && onlineDevices.length > 0) {
        console.log('设备信息未加载完成，3秒后重试...')
        setTimeout(() => {
          reportUserAssets()
        }, 3000)
        return
      }
      
      // 只处理有完整信息的设备
      const deviceList = devicesWithInfo.map(device => {
        const firmwareInfo = props.deviceFirmwareInfo.get(device.id) || {}
        const versionInfo = props.deviceVersionInfo.get(device.id) || {}
        
        console.log('设备固件信息:', device.id, firmwareInfo)
        console.log('设备版本信息:', device.id, versionInfo)
        
        return {
          device_firmware_version: firmwareInfo.originalData?.model || 'unknown',
          device_id: device.id || '',
          device_sdk_version: versionInfo.currentVersion ? String(versionInfo.currentVersion) : '0',
          device_disk_size: formatDiskSize(firmwareInfo.originalData?.mmctotal || 0)
        }
      })
      
      // 资产只在宿主侧缓存，等待首次加载或 token 续期时一并传给 iframe
      if (deviceList.length > 0) {
        const nextAssetsParam = encodeURIComponent(JSON.stringify(deviceList))
        const assetsChanged = nextAssetsParam !== cachedAssetsParam
        cachedAssetsParam = nextAssetsParam
        console.log('资产信息已缓存，设备数:', deviceList.length)

        // 首次 iframe 可能在资产准备完成前就已创建，此时补一次首帧重建
        const iframeMissingAssets = customerServiceUrl.value && !customerServiceUrl.value.includes('&assets=')
        if (isLoggedIn.value && assetsChanged && iframeMissingAssets) {
          await initCustomerServiceUrl()
          iframeKey.value++
          console.log('[customerService] 首次补齐 assets 后重建 iframe')
        }
      } else {
        console.log('没有可拼接的设备信息或客服 URL 未就绪')
      }
    } catch (error) {
      console.error('构建用户资产信息异常:', error)
    }
  }
  
  // 格式化磁盘大小
  const formatDiskSize = (sizeInKB) => {
    if (!sizeInKB || sizeInKB === 0) return '0 GB'
    const sizeInGB = (sizeInKB / 1024).toFixed(2)
    return `${sizeInGB} GB`
  }
  
  // 组件挂载时生成 ticket 并启动定时器
  onMounted(async () => {
    if (isLoggedIn.value) {
      await initCustomerServiceUrl()
    }
    // 注册 postMessage 监听，捕获 iframe 发出的 token_expired
    window.addEventListener('message', handleIframeMessage)
    window.addEventListener('beforeunload', notifyIframeShutdown)
    // onMounted 执行完后标记，后续 onActivated 才真正刷新
    isMounted = true
  })

  // 每次切换到客服页面时重新生成 ticket（keep-alive 场景）
  // keep-alive 首次挂载时 onMounted + onActivated 均会触发，用 isMounted 跳过首次
  onActivated(async () => {
    if (!isMounted) return
    if (isLoggedIn.value) {
      await initCustomerServiceUrl()
      iframeKey.value++
    }
  })
  
  // 组件卸载时清除定时器并移除监听
  onUnmounted(() => {
    notifyIframeShutdown()
    window.removeEventListener('message', handleIframeMessage)
    window.removeEventListener('beforeunload', notifyIframeShutdown)
  })
  
  // 暴露方法给父组件调用
  defineExpose({
    reportUserAssets,
    initCustomerServiceUrl,
    setPageVisible
  })

</script>

<style scoped>
.customer-service {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.customer-service-header {
  padding: 12px 20px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.customer-service-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.customer-service-content {
  flex: 1;
  overflow: hidden;
}

.customer-service-iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
