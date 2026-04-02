import { createApp, reactive } from 'vue'
import App from './App.vue'
import './style.css';
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import axios from 'axios'

// 翻译消息数据
const messages = {
  'zh-CN': {
    common: {
      sortBy: '排序方式',
      createTimeSort: '创建时间',
      selectedItems: '已选择 {count} 项',
      backupName: '备份名称',
      cloudMachine: 'Machine',
      statusInfo: '状态信息',
      statusLabel: '状态',
      projection: '投屏',
      confirm: '确认',
      cancel: '取消',
      save: '保存',
      delete: '删除',
      close: '关闭',
      settings: '设置',
      upload: '上传',
      refresh: '刷新',
      cloudMachine: '云机',
      image: '镜像',
      device: '设备',
      taskQueue: '任务队列',
      deleteSelected: '删除选中',
      addDevice: '添加设备',
      bindHost: '绑定主机',
      unbindHost: '主机解绑',
      syncAuth: '授权同步',
      batchUpgradeAPI: '批量升级API',
      clearClientData: '清理客户端数据',
      batchCleanDisk: '批量清理磁盘数据',
      logout: '退出登录',
      onlineDevices: '在线设备',
      allDevices: '全部设备',
      offlineDevices: '离线设备',
      hostFirmwareVersion: '主机固件版本',
      nvmeStorage: 'NVME存储空间',
      bindStatus: '绑定状态',
      apiVersion: 'API版本',
      operation: '操作',
      view: '查看',
      upgrade: '升级',
      loading: '加载中',
      current: '当前',
      latest: '最新',
      group: '分组',
      defaultGroup: '默认分组',
      unknown: '未知',
      phoneNumber: '手机号码',
      verificationCode: '验证码',
      submit: '立即提交',
      unbindWarning: '注意：解绑后该主机下购买的所有云机实例都将会被清空,再次进行绑定需要重新购买云机实例才可以使用。(数据不会清理,若有需求请自行清理),该操作不可逆,请知晓风险,谨慎操作!',
      sendCode: '发送验证码',
      resendCode: '重新发送',
      addGroup: '添加新分组',
      editGroup: '编辑分组名称',
      deleteGroup: '删除分组',
      create: '创建',
      enterNewGroupName: '请输入新分组名称',
      groupNameCannotBeEmpty: '分组名称不能为空',
      collapseRightSidebar: '收起右侧栏',
      hostInfo: '主机信息',
      refreshHostInfo: '刷新主机信息',
      basicInfo: '基本信息',
      v3DeviceInfo: 'V3设备信息',
      hostName: '主机名称',
      hostIP: '主机IP',
      deviceID: '设备ID',
      deviceVersion: '设备版本',
      deviceModel: '设备型号',
      cpuTemp: 'CPU温度',
      cpuLoad: 'CPU负载',
      memoryUsage: '内存使用',
      storageUsage: '存储使用',
      ipAddress: 'IP地址',
      macAddress: 'MAC地址',
      firmwareVersion: '固件版本',
      modelVersion: '型号版本',
      deviceUptime: '设备运行时间',
      memoryTotal: '内存总大小',
      memoryUsed: '内存已使用',
      diskTotal: '磁盘总大小',
      diskUsed: '磁盘已使用',
      diskModel: '磁盘型号',
      diskTemp: '磁盘温度',
      diskRead: '磁盘读取量',
      diskWrite: '磁盘写入量',
      networkIP: '网口IP',
      networkSpeed: '网口速率',
      eth0Network: 'ETH0网卡',
      loading: '加载中...',
      passwordManagement: '密码管理',
      setPassword: '设置密码',
      closePassword: '关闭密码',
      passwordHint: '注：设置密码后，访问设备8000端口的v3 API时需要输入密码',
      sdkUpgrade: 'SDK升级',
      currentSDKVersion: '当前SDK版本',
      latestSDKVersion: '最新SDK版本',
      upgradeSDK: '升级SDK',
      upgrading: '升级中',
      cleanDiskData: '清理磁盘数据',
      latitude: '纬度',
      longitude: '经度',
      selectDeviceFirst: '请先选择设备',
      loadFailed: '加载失败',
      time: '构建时间',
      type: '系统类型',
      dockerNetworkList: 'Docker网络列表',
      addMacvlanNetwork: '添加macvlan网络',
      editNetwork: '修改网络',
      networkName: '网络名称',
      physicalInterface: '物理网口',
      subnet: '网段',
      gateway: '网关',
      ipRange: 'IP范围',
      isolationMode: '隔离模式',
      enablePrivateIsolation: '开启私有隔离',
      isolationTip: '开启后，该网络内的容器只能与同网络内的容器通信',
      enterNetworkName: '输入网络名称，例如：mynetwork',
      enterPhysicalInterface: '输入物理网口，例如：eth0',
      enterSubnet: '输入网段，例如：192.168.1.0/24',
      enterGateway: '输入网关，例如：192.168.1.1',
      enterIPRange: '输入IP范围，例如：192.168.1.100/28',
      apiDetails: 'API详情',
      slot: '坑位',
      instanceName: '实例名称',
      deviceIP: '设备IP',
      deviceVersion: '设备版本',
      portMappingInfo: '端口映射信息',
      service: '服务',
      portMapping: '端口映射',
      accessAddress: '访问地址',
      clickToCopy: '点击访问地址可复制到剪贴板',
      close: '关闭',
      setS5Proxy: '设置S5代理',
      cloudMachineName: '云机名称',
      s5Info: 'S5信息',
      s5InfoFormat: '格式: 地址:端口:用户名:密码 (用户名密码可选)',
      s5InfoExample: '例如: 192.168.1.1:22:root:123456 或 192.168.1.1:22',
      parseAndFill: '解析并填写',
      s5ServerAddress: 's5服务器地址',
      s5Port: 's5端口',
      enterS5ServerAddress: '请输入s5服务器地址',
      enterS5Port: '请输入s5端口',
      dnsMode: '域名解析模式',
      localDNS: '本地域名解析',
      serverDNS: '服务端域名解析（默认）',
      submitting: '提交中...',
      submitNow: '立即提交',
      batchCreateCloudMachine: '批量创建云机',
      createCloudMachine: '创建云机',
      googleGreen: '谷歌三绿',
      realParameters: '参数真实',
      simulator: '模拟器',
      container: '容器',
      oldVersionImage: '老版本镜像',
      deviceRemainingSpace: '设备剩余空间',
      insufficientSpace: '剩余空间不足，可能会无法开机/创建',
      apiVersion: 'API版本',
      latestVersion: '最新',
      notLatestVersion: '不是最新版本，升级可以支持最新功能',
      androidVersion: '安卓版本',
      pSeries: 'P系列',
      otherSeries: '其他系列',
      pleaseSelectDevice: '请选择设备',
      customImage: '自定义镜像',
      imageAddress: '镜像地址',
      pleaseSelect: '请选择',
      name: '名称',
      cloudMachineCount: '云机数量',
      slot: '坑位',
      normal: '正常',
      expiringSoon: '即将过期',
      expired: '已过期',
      noInstance: '无实例',
      singleSlotCount: '单坑位数量',
      quantity: '数量',
      resolution: '分辨率',
      customResolution: '自定义分辨率',
      deviceWidth: '设备宽',
      deviceHeight: '设备长',
      resolutionWarning: '请注意，自定义分辨率可能引发样式适配异常',
      dnsType: 'DNS 类型',
      dnsAddress: 'DNS 地址',
      aliDNS: '阿里DNS(223.5.5.5)',
      googleDNS: 'Google(8.8.8.8)',
      custom: '自定义',
      sandboxMode: '沙盒模式?',
      enable: '开启',
      disable: '关闭',
      secureMode: '安全模式',
      dataDiskSize: '数据盘大小',
      networkManagement: '网络管理',
      selectGroup: '选择分组',
      selectNode: '选择节点',
      specifiedNode: '指定节点',
      randomNode: '随机节点',
      networkCardType: '网卡类型',
      privateNetworkCard: '私有网卡',
      publicNetworkCard: '公有网卡',
      noMacVlanDetected: '未检测到MacVlan配置，请前往网络管理-公有网卡创建',
      enterStartIPAddress: '请输入起始IP地址，例如：10.10.0.253',
      batchCreateTip: '提示：批量创建{count}台，分配IP范围：{range}',
      containerName: '容器名称',
      enterContainerNamePrefix: '输入容器名称前缀，例如：T000',
      singleSlotCountLabel: '单个坑位数量',
      imageCategory: '镜像分类',
      onlineImage: '在线镜像',
      localImage: '本地镜像',
      imageSelection: '镜像选择',
      pleaseSelectLocalImage: '请选择本地镜像',
      customImageAddress: '自定义镜像地址',
      enterImageAddress: '输入镜像地址，例如：registry.example.com/myimage:tag',
      creationMethod: '创建方式',
      cacheToLocal: '缓存到本地创建',
      cacheToLocalTip: '选中后，在线镜像将先下载到本地，然后推送到设备创建云机',
      s5Proxy: 'S5代理',
      pleaseSelectProxyType: '请选择代理类型',
      noProxy: '不开启代理',
      localDNSParsing: '本地域名解析(tun2socks)',
      serverDNSParsing: '服务器域名解析(tun2proxy)',
      lockScreenPassword: '锁屏密码',
      noLockScreen: '不填没有锁屏',
      port: '端口',
      enterIP: '输入IP',
      enterPort: '输入端口',
      user: '用户',
      enterUser: '输入用户',
      enterPassword2: '输入密码',
      ipAddress: 'IP地址',
      enterStartIP: '输入IP地址起始，例如：192.168.1.100',
      dnsAddress: 'DNS地址',
      pleaseSelectDNSAddress: '请选择DNS地址',
      aliCloud: '阿里云',
      enterCustomDNSAddress: '输入自定义DNS地址',
      sandboxSize: '沙盒大小',
      defaultResolution: '机型默认分辨率',
      width: '宽度',
      height: '高度',
      modelCountry: '机型国家',
      pleaseSelectModelCountry: '请选择机型国家,支持搜索',
      modelType: '机型类型',
      onlineModel: '在线机型',
      localModel: '本地机型',
      backupModel: '备份机型',
      phoneModel: '手机型号',
      pleaseSelectPhoneModel: '请选择手机型号',
      random: '随机',
      pleaseSelectLocalModel: '请选择本地机型',
      pleaseSelectBackupModel: '请选择备份机型',
      advancedOptions: '高级选项',
      enableMagisk: '启用Magisk',
      enableGMS: '启用GMS',
      randomSystemFiles: '随机系统文件',
      networkCardSelection: '网卡选择',
      pleaseSelectNetworkCard: '请选择网卡',
      syncAuthLogin: '授权同步登录',
      username: '用户名',
      password: '密码',
      enterUsername: '请输入用户名',
      enterPassword: '请输入密码',
      rememberCredentials: '记住凭证',
      login: '登录',
      loggingIn: '登录中...',
      register: '注册',
      enterUsernameAndPassword: '请输入用户名和密码',
      loginSuccess: '登录成功',
      syncAuthFailed: '授权同步失败',
      userRegistration: '用户注册',
      phoneNumber: '手机号',
      loginPassword: '登录密码',
      confirmPassword: '确认密码',
      phoneVerificationCode: '手机验证码',
      enterPhoneNumber: '请输入手机号',
      enterLoginPassword: '请输入登录密码',
      enterPasswordAgain: '请再次输入密码',
      enterVerificationCode: '请输入验证码',
      getVerificationCode: '获取验证码',
      retryAfterSeconds: '秒后重试',
      registering: '注册中...',
      enterPhone: '请输入手机号',
      enterPassword: '请输入密码',
      enterConfirmPassword: '请输入确认密码',
      passwordMismatch: '两次输入的密码不一致',
      enterVCode: '请输入验证码',
      getVCodeFirst: '请先获取验证码',
      registerSuccess: '注册成功',
      registerFailed: '注册失败',
      vcodeSentSuccess: '验证码发送成功',
      enterCorrectPhone: '请输入正确的手机号',
      syncAuthCompletedForAllDevices: '已为所有 {count} 个设备完成同步授权',
      syncAuthPartialSuccess: '部分设备同步授权成功，共 {count} 个设备',
      syncAuthAllFailed: '所有设备同步授权失败',
      notLoggedIn: '未登录',
      unbound: '未绑定',
      bound: '已绑定',
      boundByOthers: '被绑定',
      deviceUpgradeSuccess: '设备 {ip} 升级成功',
      deviceUpgradeSuccessWithMessage: '设备 {ip} 升级成功: {message}',
      addNewGroup: '添加新分组',
      deleteCurrentGroup: '删除当前分组',
      moreActions: '更多操作',
      openProjection: '打开投屏',
      updateImage: '更新镜像',
      batchActions: '批量操作',
      batchRestart: '批量重启',
      batchReset: '批量重置',
      batchProjection: '批量投屏',
      batchShutdown: '批量关机',
      batchUpdateImage: '批量更新镜像',
      batchDelete: '批量删除',
      batchSwitchBackup: '批量切换备份',
      deviceDetails: '设备详情',
      setDevicePassword: '设置设备密码',
      deviceAuthentication: '设备认证',
      unknownDevice: '未知设备',
      addDevicesToGroup: '添加设备到分组',
      selectPushDevice: '选择推送设备',
      deviceUpgrade: '设备升级',
      noDeviceSelected: '未选择设备',
      refreshCloudMachines: '刷新云机',
      instance: '实例',
      image: '镜像',
      network: '网络',
      host: '主机',
      refreshImages: '刷新镜像',
      refreshNetworkList: '刷新网络列表',
      slot: '坑位',
      create: '创建',
      instanceName: '实例名称',
      ipAddress: 'IP地址',
      systemImage: '系统镜像',
      createTime: '创建时间',
      status: '状态',
      running: '运行中',
      shutdown: '关机',
      created: '已创建',
      restarting: '重启中',
      model: '机型',
      none: '无',
      startUp: '开机',
      operationConfirm: '操作确认',
      confirmAction: '确定要{action}实例 {name}吗？',
      actionSuccess: '{action}成功',
      actionFailed: '{action}失败：{error}',
      slotNoMachine: '坑位 {slot} 暂无云机，请先创建云机',
      slotCannotDelete: '坑位 {slot} 暂无云机，无法删除',
      noDownloadedImages: '暂无已下载镜像信息',
      pleaseSelectDeviceFirst: '请先在左侧选择设备',
      onlineImageName: '线上镜像名称',
      imageSize: '镜像大小',
      matchedWithOnlineImage: '✓ 已与线上镜像匹配',
      notMatchedWithOnlineImage: '✗ 未与线上镜像匹配',
      onlineURL: '线上URL',
      deviceURL: '设备中URL',
      refreshOnlineImages: '刷新在线镜像',
      openLocalImageDirectory: '打开本地镜像目录',
      deviceImage: '设备镜像',
      other: '其他',
      simulator: '模拟器',
      container: '容器',
      androidVersion: '安卓版本',
      imageName: '镜像名称',
      updateContent: '更新内容',
      download: '下载',
      uploadToDevice: '上传到设备',
      filePath: '文件路径',
      size: '大小',
      availableDeviceModels: '可用设备型号',
      universal: '通用',
      noDevice: '暂无设备',
      noCompatibleDevices: '暂无兼容设备',
      pleaseSelectDeviceForUpload: '请选择要上传镜像的设备',
      refreshLocalImages: '刷新本地镜像',
      noLocalCachedImages: '暂无本地缓存镜像',
      noOnlineImages: '暂无符合条件的在线镜像',
      filterImageName: '筛选镜像名称',
      filterImageURL: '筛选镜像URL',
      compatibleImages: '兼容镜像',
      imagePath: '镜像路径',
      updateImageWarning: '更新镜像方式创建，不保证能正常开机',
      enterCustomImageAddress: '请输入自定义镜像地址',
      batchUpdateImage: '批量更新镜像',
      updateVersion: '更新版本',
      versionType: '版本类型',
      v3Simulator: 'V3（模拟器）',
      v2Container: 'V2（容器）',
      noImageForVersion: '暂无该版本镜像，可选择自定义镜像',
      customAddress: '自定义地址',
      enterImageURL: '请输入镜像地址',
      confirmUpdate: '确定更新',
      totalCloudMachines: '共 {count} 个云机',
      ipTestTitle: 'IP连接测试',
      testResult: '测试结果',
      startTest: '开始测试',
      testing: '测试中...',
      importantTip: '⚠️ 重要提示',
      ensureIPInSubnet: '请确保IP在子网范围内',
      ensureIPNotUsed: '请务必确认IP地址未被占用',
      ipConflictWarning: '否则会造成IP冲突导致无法访问',
      pleaseSelectImage: '请选择镜像',
      batchUpdateImageBtn: '批量更新镜像',
      setIPLocation: '设置IP定位',
      locationIP: '定位IP',
      leaveEmptyForCurrentIP: '留空则使用当前云机IP',
      countryRegion: '国家/地区',
      pleaseSelectCountryRegion: '请选择国家/地区',
      setMacVlanIP: '设置MacVlan IP',
      uploadGoogleCert: '上传 Google 证书',
      supportedFormats: '支持格式',
      clickToSelectCert: '点击选择证书文件',
      reselect: '重新选择',
      uploading: '上传中...',
      confirmUpload: '确认上传',
      searchGroupOrIP: '搜索分组名或设备IP',
      dropToGroup: '拖放到此分组',
      cloudMachineManagement: '云机管理',
      cloudMachineColon: '云机: {ip}',
      batchManagement: '批量管理',
      refreshCloudMachine: '刷新云机',
      stopControl: '停止控制',
      batchControl: '批量控制',
      batchCloseProjection: '批量关闭投屏',
      batchUpload: '批量上传',
      batchNewDevice: '批量新机',
      zoom: '缩放',
      layout: '布局',
      grid: '网格',
      list: '列表',
      switchToHorizontal: '切换到横屏',
      switchToVertical: '切换到竖屏',
      vertical: '竖屏',
      horizontal: '横屏',
      switching: '切换中...',
      running: '运行中',
      backupCount: '{count}备份',
      empty: '空',
      booting: '开机中...',
      shutdownClickBackup: '已关机 | 点击切换备份',
      emptySlot: '空坑位',
      switchBackupBtn: '切换',
      copy: '复制',
      slotNoMachineCannotDelete: '坑位 {slot} 暂无云机，无法删除',
      selectedCloudMachineList: '已选云机列表 ({count}台)',
      noSelectedCloudMachine: '暂无选中的云机',
      pleaseSelectCloudMachine: '请在左侧树形结构中勾选需要操作的云机',
      openProjectionBtn: '打开投屏',
      slot: '坑位',
      instanceNameLabel: '实例名称',
      systemImage: '系统镜像',
      createTime: '创建时间',
      statusLabel: '状态',
      shutdownStatus: '关机',
      restartingStatus: '重启中',
      modelLabel: '机型',
      noneLabel: '无',
      operationLabel: '操作',
      create: '创建',
      selectDevicesToAdd: '选择要添加到该分组的设备：',
      selectAllLabel: '全选',
      defaultGroup: '默认分组',
      allDevicesInGroup: '所有设备都已在该分组中',
      addCount: '添加 ({count})',
      copyCloudMachine: '复制云机',
      cloudMachineName: '云机名称',
      targetSlot: '目标坑位',
      targetSlotHint: '复制到的目标坑位编号',
      copyCount: '复制数量',
      confirmCopy: '确认复制',
      batchCloseProjectionSuccess: '批量关闭投屏成功',
      batchCloseProjectionFailed: '批量关闭投屏失败',
      batchUploadFileTo: '批量上传文件到云机',
      closeAll: '全部关闭',
      noRunningCloudMachine: '没有选中已运行的云机，批量新机操作仅支持已运行的云机',
      imageSelection: '镜像选择',
      searchIP: '搜索IP',
      all: '全部',
      group: '分组',
      noDeviceSelected: '未选择设备',
      deviceNetworkLatency: '设备网络延迟',
      networkName: '网络名称',
      networkID: '网络ID',
      subnet: '网段',
      driver: '驱动',
      scope: '范围',
      gateway: '网关',
      ipRange: 'IP范围',
      containerCount: '容器数量',
      private: '私有',
      yes: '是',
      no: '否',
      edit: '编辑',
      delete: '删除',
      noNetworkInfo: '暂无网络信息',
      confirm: '确定',
      cancel: '取消',
      // OpenCecs
      cloudInstanceManagement: '云实例管理',
      login: '登录',
      logout: '退出',
      allInstances: '全部',
      running: '运行中',
      stopped: '已停止',
      expired: '已过期',
      refreshing: '刷新中...',
      refresh: '刷新',
      loadingData: '加载中...',
      noInstanceData: '暂无实例数据',
      instanceID: '实例ID',
      instanceNameLabel: '实例名称',
      coreBoard: '核心板',
      memory: '内存',
      system: '系统',
      ipAddress: 'IP地址',
      statusLabel: '状态',
      billingMode: '计费方式',
      expiryTime: '到期时间',
      operationLabel: '操作',
      start: '启动',
      stop: '停止',
      restart: '重启',
      loginFirst: '请先登录以查看您的云实例',
      loginNow: '立即登录',
      welcomeOpenCecs: '欢迎使用 OpenCecs - 全球领先的云服务提供商',
      phoneTab: '手机号',
      accountTab: '账号',
      // StreamManagement
      streamManagement: '流管理 (Streams)',
      forwardStreamOBS: '转发流(OBS)',
      p2pStreamOBS: '点对点(OBS)',
      pcCamera: 'PC摄像头',
      userGuide: '使用说明',
      enterRoomNumber: '输入房间号',
      newStream: '新建流',
      refreshList: '刷新列表',
      streamCode: '推流码/房间',
      sourceIP: '源IP',
      bitrate: '码率 (Video/Audio)',
      noStreamInfo: '暂无流信息，请点击上方新建',
      active: '活跃 (Active)',
      idle: '待推流 (Idle)',
      distribute: '分发',
      disconnect: '断开',
      addP2PStream: '添加P2P流',
      pushAddress: '推流地址',
      deviceIP: '设备IP',
      cloudMachine: '云机',
      listenAddress: '监听地址',
      startP2P: '启动P2P',
      stopP2P: '停止P2P',
      noP2PStreamInfo: '暂无P2P流信息，请点击上方添加',
      addCameraStream: '添加摄像头推流',
      camera: '摄像头',
      resolutionBitrate: '分辨率/码率',
      defaultCamera: '默认摄像头',
      streaming: '推流中',
      streamStopped: '已停止',
      noCameraStream: '暂无摄像头推流，请点击上方添加',
      cameraPreview: '摄像头实时预览',
      deviceCloudRelation: '设备-云机关系 (Device & Cloud Machines)',
      distributeToCloud: '分发到云机',
      autoRefreshCapture: '自动刷新中（每 {interval}s 抓帧）',
      autoPreviewAfterStart: '启动推流后自动预览',
      capturingFrame: '抓取画面中…',
      refreshingFrame: '刷新中…',
      startStreamToPreview: '在左侧「摄像头模式」启动推流后自动预览',
      configuredDistribution: '已配置分发 (Active)',
      refreshStatus: '刷新状态',
      protocol: '协议',
      queryingStatus: '正在查询状态...',
      noActiveDistribution: '暂无活跃分发',
      modifyStream: '修改流',
      projection: '投屏',
      inactiveDistribution: '失效分发 (Inactive/Unverified)',
      statusInfo: '状态信息',
      noInactiveDistribution: '暂无失效分发',
      unverified: '未验证',
      distributeStreamDialog: '分发流到云机',
      stream: '流',
      selectOrEnterRoom: '选择或输入房间号',
      device: '设备',
      selectDevice: '选择设备',
      selectCloudMachine: '选择云机',
      resolution: '分辨率',
      auto: '自动',
      confirmDistribute: '确认分发',
      addP2PStreamDialog: '添加P2P流',
      p2pPort: 'P2P端口',
      portOccupied: '端口已被占用，请选择其他端口',
      remark: '备注',
      optionalDefault: '可选，默认 p2p-端口',
      fillPort: '请填写端口',
      addCameraStreamDialog: '添加摄像头推流',
      devicePort: '设备端口',
      selectCloudMachineFirst: '（请先选择云机）',
      selectCamera: '选择摄像头',
      none: '无',
      width: '宽',
      height: '高',
      fpsAndBitrate: '帧率/码率',
      confirmAdd: '确认添加',
      // InterconnectedCloudMachines
      selected: '已选 {count} 台',
      totalMachines: '共 {count} 台云机',
      batchDeleteBtn: '批量删除',
      stopControlBtn: '停止控制',
      batchControlBtn: '批量控制',
      addCloudMachine: '添加云机',
      clickToProject: '点击投屏',
      noMachineHint: '暂无云机，点击「添加云机」通过密钥导入',
      addMachineDialog: '添加云机',
      enterKeyHint: '请输入密钥，多个密钥用逗号或换行分隔，支持批量粘贴',
      confirmAddBtn: '确认添加',
      // BatchTaskManagement
      selectDevices: '📱 选择设备',
      selectedCount: '已选 {count} 个',
      searchDeviceOrContainer: '搜索设备IP或容器',
      selectAllBtn: '全选',
      clearBtn: '清空',
      openAll: '全部打开',
      closeAllBtn: '全部关闭',
      totalDevices: '共 {count} 台',
      projectionBtn: '投屏',
      noAvailableDevice: '暂无可用设备',
      batchOperation: '⚡ 批量操作',
      batchExecCommand: '📝 批量执行命令',
      inputADBCommand: '输入ADB命令：',
      viewCommandExample: '查看命令示例',
      quickCommands: '🔖 快捷命令',
      clickToFill: '点击按钮快速填充',
      basicOps: '📱 基础操作',
      appManagement: '📦 应用管理',
      systemInfo: '🔧 系统信息',
      loopCount: '循环次数：',
      times: '次',
      executing: '执行中 (第{current}/{total}轮)...',
      executeNow: '🚀 立即执行',
      stopBtn: '⏹ 停止',
      execResult: '📊 执行结果',
      roundSeparator: '第 {loop} / {total} 轮',
      errorPrefix: '错误',
      closeBtn: '关闭',
      adbCommandRef: '📖 ADB命令参考',
      // 命令名称
      cmdInputText: '输入文本',
      cmdTapScreen: '点击屏幕',
      cmdSwipeScreen: '滑动屏幕',
      cmdHomeKey: 'Home键',
      cmdBackKey: '返回键',
      cmdPowerKey: '电源键',
      cmdVolUp: '音量+',
      cmdVolDown: '音量-',
      cmdScreenshot: '截屏',
      cmdRecordStart: '录屏开始',
      cmdOpenSettings: '打开设置',
      cmdOpenDial: '打开拨号',
      cmdListApps: '列出应用',
      cmdListSysApps: '列出系统应用',
      cmdList3rdApps: '列出第三方应用',
      cmdFindApp: '查找应用',
      cmdClearAppData: '清除应用数据',
      cmdUninstallApp: '卸载应用',
      cmdForceStopApp: '强制停止应用',
      cmdStartApp: '启动应用',
      cmdSysVersion: '系统版本',
      cmdDeviceModel: '设备型号',
      cmdDeviceBrand: '设备品牌',
      cmdAndroidId: 'Android ID',
      cmdIpAddress: 'IP地址',
      cmdCpuInfo: 'CPU信息',
      cmdMemInfo: '内存信息',
      cmdStorage: '存储空间',
      cmdBattery: '电池状态',
      cmdScreenRes: '屏幕分辨率',
      cmdScreenDensity: '屏幕密度',
      cmdCurrentActivity: '当前Activity',
      // 命令参考
      commonAdbExamples: '常用ADB命令示例：',
      inputText: '输入文本',
      tapCoordinate: '点击坐标 (x=500, y=800)',
      swipeScreen: '滑动屏幕',
      pressHomeKey: '按Home键 (3=HOME, 4=BACK, 26=POWER)',
      listAllPackages: '列出所有应用包名',
      installApp: '安装应用',
      uninstallApp: '卸载应用',
      startSettings: '启动设置',
      checkBatteryInfo: '查看电池信息',
      screenshot: '截屏',
      getSystemVersion: '获取系统版本',
      supportedVars: '支持的变量替换：',
      deviceIpAddress: '设备IP地址',
      containerFullId: '容器完整ID',
      containerShortId: '容器短ID (前12位)',
      containerName: '容器名称',
      currentTimestamp: '当前时间戳',
      precautions: '注意事项：',
      cmdExecSequentially: '命令会在每个选中的设备上依次执行',
      quotesNeededForSpaces: '包含空格的文本参数需要用引号包裹',
      rootRequired: '某些命令需要root权限才能执行',
      // AI Assistant
      onlineDevicesLabel: '在线设备',
      noAIDevice: '暂无支持AI的设备 (需要 r1q 或 eces-rk3588-rk1828 机型)',
      modelManagement: '模型管理',
      importModel: '导入模型',
      noModel: '暂无模型，请先导入',
      chatModel: '对话模型',
      selectChatModel: '请选择对话模型',
      startingModel: '启动中...',
      startModel: '启动模型',
      stopModel: '停止',
      aiChat: 'AI 对话',
      stopGeneration: '停止生成',
      continueGenerate: '继续生成',
      newChat: '新对话',
      settings: '设置',
      clearChat: '清空对话',
      selectDeviceAndLoadModel: '请先选择设备并加载模型',
      helloAssistant: '你好！我是魔云腾助手',
      howCanIHelp: '有什么我可以帮助你的吗？',
      thinkingProcess: '思考过程',
      thinkingInProgress: '思考中...',
      feedbackQuestion: '这次回答对你有帮助吗？',
      feedbackUseful: '👍 有效，记住它',
      feedbackUseless: '👎 没用',
      serviceManagement: '服务管理',
      stopServiceWarning: '停止服务后将无法进行对话，需要重新选择模型加载',
      currentDevice: '当前设备',
      currentModel: '当前模型',
      noModelLoaded: '未加载模型',
      stoppingService: '停止中...',
      stopLLMService: '停止LLM服务',
      deviceManagement: '设备管理',
      resetDeviceWarning: '重置设备将清除设备的硬件配置和状态，需要谨慎操作',
      resetOperation: '重置操作',
      resettingDevice: '重置中...',
      resetDevice: '重置设备',
      resetDeviceNote: '⚠️ 重置设备后需要重新配置和加载模型',
      noDeviceSelected: '未选择设备',
      closeInSeconds: '{seconds}s 后关闭',
      uploadMediaTooltip: '上传图片或视频',
      running: '运行中',
      deleteModel: '删除模型',
      resetConfirmMessage: '重置设备后需要重新配置和加载模型',
      copyBtn: '复制',
      systemPrompt: '系统提示词',
      systemPromptDesc: '系统提示词会在每次对话时自动添加到消息开头，用于设定AI的角色和行为',
      enablePrompt: '启用提示词',
      promptContent: '提示词内容',
      saveSettings: '保存设置',
      restoreDefault: '恢复默认',
      // RPA Agent
      selectContainer: '选择容器',
      loadingContainers: '加载容器列表...',
      bridged: '桥接',
      nonBridged: '非桥接',
      noRunningContainer: '暂无运行中容器',
      noSupportedDevice: '暂无支持的在线设备（需 r1q 或 eces-rk3588-rk1828 机型）',
      noAndroidContainer: '所有支持的设备均无运行中的安卓容器。请先在设备管理页面创建并启动云机，然后点击刷新。',
      executionProgress: '执行进度',
      selectTaskType: '选择任务类型',
      executeTask: '执行任务',
      replyBtn: '回复',
      agentWaiting: 'Agent 正在等待你的回复',
      selectedContainers: '已选 {count} 个容器',
      taskLabel: '任务：',
      rpaSettings: 'RPA Agent 设置',
      deviceModelConfig: '各设备模型配置（r1q / eces-rk3588-rk1828 机型）',
      noSupportedOnlineDevice: '暂无支持的在线设备',
      online: '在线',
      offline: '离线',
      llmModel: 'LLM 模型',
      selectChatModelRPA: '选择对话模型',
      refreshModelList: '刷新模型列表',
      checkAndStartModel: '检查并启动模型',
      agentGlobalParams: 'Agent 全局参数',
      breakerRoundLimit: '熔断轮次上限',
      breakerHint: '0 = 无上限（仅 task_done 或停止退出）',
      stepDelay: '操作间隔(ms)',
      saveConfig: '保存配置',
      // RPA Task Cards
      taskBrowseVideo: '刷视频',
      taskSendMessage: '发消息',
      taskInstallApp: '装应用',
      taskLikeContent: '点赞',
      taskSearch: '搜索',
      taskCustom: '自定义',
      // RPA Status Labels
      statusPending: '等待',
      statusRunning: '执行中',
      statusSuccess: '成功',
      statusFailed: '失败',
      statusStopped: '已停止',
      // RPA Messages
      agentWaitingReply: 'Agent 有问题等待你回复...',
      taskExecuting: '任务执行中...',
      taskInputHint: '输入任务描述，例如：打开抖音向上列10条视频',
      modelStatusRunning: '运行',
      modelStatusStarting: '启动',
      modelStatusStopped: '未启动',
      modelStatusUnknown: '未知',
      refreshContainerList: '刷新容器列表',
      clickRefreshToLoad: '点击刷新加载',
      noModelOnDevice: '该设备暂无模型',
      // HostManagement image tab
      onlineImages: '在线镜像',
      localImages: '本地镜像',
      noMatchingImages: '暂无符合条件的在线镜像',
      filterImageName: '筛选镜像名称',
      filterImageURL: '筛选镜像URL',
      compatibleImages: '兼容镜像',
      resetFilter: '重置筛选',
      refreshOnlineImages: '刷新在线镜像',
      noDescription: '暂无描述',
      downloadImage: '下载',
      selectDeviceFirst: '请先选择设备',
      selectGroupPlaceholder: '选择分组',
      sharedIP: '共享IP',
      independentIP: '独立IP',
      deleteGroupConfirm: '删除分组确认',
      deleteGroupMessage: '确定要删除分组 "{group}" 吗？该分组下的所有设备将移至"默认分组"',
      editGroup: '编辑分组',
      enterNewGroupName: '请输入新的分组名称',
      groupNameCannotBeEmpty: '分组名称不能为空',
      importantNotice: '重要提示（点击展开/收起）',
      supportScope: '支持范围',
      supportScopeDesc: '仅支持模拟器镜像类型的云机备份，',
      containerNotSupported: '容器模式镜像不支持',
      backupImportFunction: '备份导入功能',
      importPreparation: '导入前准备',
      importPreparationDesc: '请确保目标设备已下载对应的安卓镜像，镜像未下载将导致导入失败或者启动失败',
      deviceCompatibility: '设备兼容性',
      deviceCompatibilityDesc: '备份文件必须与目标设备类型完全匹配（',
      cqrReusable: 'CQR系列镜像可复用',
      pSeriesNotReusable: 'P系列不能复用',
      backupFileList: '备份文件列表',
      openImportFolder: '打开导入文件夹',
      noBackupFiles: '暂无备份文件',
      copyToImportFolder: '请将云机备份文件（.tar.gz 格式）复制到导入文件夹中',
      thenRefresh: '然后点击右上角"刷新"按钮',
      fileName: '文件名',
      fileSize: '文件大小',
      import: '导入',
      batchImport: '批量导入',
      selectDevice: '选择设备',
      configureSlots: '配置坑位',
      executeImport: '执行导入',
      searchDeviceIP: '搜索设备IP',
      deviceIP: '设备IP',
      hostFirmwareVersion: '主机固件版本',
      nvmeStorage: 'NVME存储空间',
      noAvailableDevices: '暂无可用设备',
      nextStep: '下一步',
      previousStep: '上一步',
      startImport: '开始导入',
      selectAll: '全选',
      clear: '清空',
      batchSetCopyCount: '批量设置复制份数',
      slotLabel: '坑位',
      copies: '份',
      currentBackupFile: '当前备份文件:',
      overallProgress: '总体进度',
      success: '成功',
      failed: '失败',
      importingNow: '正在导入...',
      device: '设备',
      successfulImports: '成功导入的云机',
      importFailed: '导入失败',
      complete: '完成'
    },



    backup: {
      add: '新增',
      addBackupMachine: '新增云机备份',
      addBackupModel: '新增机型备份',
      all: '全部',
      availableSlots: '可用坑位',
      backupMachine: '备份云机',
      backupMachineList: '备份云机列表',
      backupModel: '备份机型',
      backupModelList: '备份机型列表',
      batchImport: '批量导入',
      batchImportBackupModel: '批量导入备份机型',
      cancel: '取消',
      confirm: '确定',
      delete: '删除',
      deviceIP: '设备IP',
      deviceList: '设备列表',
      deviceStatus: '设备状态',
      download: '下载',
      downloading: '下载中',
      ensureImageDownloaded: '确保镜像已下载',
      enterMachineName: '输入云机名称',
      enterModelName: '输入机型名称',
      export: '导出',
      group: '分组',
      import: '导入',
      importBackupMachine: '导入备份云机',
      importBtn: '导入',
      index: '序号',
      machineName: '云机名称',
      modelName: '机型名称',
      noAvailableSlots: '无可用坑位',
      openLocalBackupMachine: '打开本地云机备份目录',
      openLocalBackupModel: '打开本地机型备份目录',
      operation: '操作',
      pleaseSelectDevice: '请选择设备',
      pleaseSelectMachine: '请选择云机',
      searchMachineName: '搜索云机名称',
      searchModelName: '搜索机型名称',
      selectDevice: '选择设备',
      selectMachine: '选择云机',
      size: '大小',
      slotNumber: '坑位编号',
      tip: '提示',
      usageGuide: '使用说明'
    },

    batchTask: {
      selectDeviceTitle: '选择设备',
      selected: '已选',
      unit: '个',
      searchPlaceholder: '搜索设备IP或容器',
      selectAll: '全选',
      clear: '清空',
      openAll: '全部打开',
      closeAll: '全部关闭',
      total: '共',
      unitMachines: '台',
      projection: '投屏',
      noDevices: '暂无可用设备',
      batchOperationTitle: '批量操作',
      batchExecuteCmd: '批量执行命令',
      inputAdbCmd: '输入ADB命令：',
      viewCmdExample: '查看命令示例',
      cmdExamplePlaceholder: '示例命令：\ninput text \'Hello World\'\ninput tap 500 800\npm list packages',
      quickCmd: '快捷命令',
      clickToFill: '点击按钮快速填充',
      basicOps: '基础操作',
      appOps: '应用管理',
      sysOps: '系统信息',
      executing: '执行中 (第',
      loopSuffix: '轮)',
      executeNow: '立即执行',
      stop: '停止',
      executionResultTitle: '执行结果',
      loopPrefix: '第',
      errorStr: '错误',
      cmdReference: 'ADB命令参考',
      commonCmdExamples: '常用ADB命令示例：',
      cmdInputText: '输入文本',
      cmdTap: '点击屏幕',
      cmdSwipe: '滑动屏幕',
      cmdHome: 'Home键',
      cmdBack: '返回键',
      cmdPower: '电源键',
      cmdVolUp: '音量+',
      cmdVolDown: '音量-',
      cmdScreencap: '截屏',
      cmdRecord: '录屏开始',
      cmdStartSettings: '启动设置',
      cmdDial: '打开拨号',
      cmdListPkgs: '列出所有应用包名',
      cmdListSys: '列出系统应用',
      cmdListThird: '列出第三方应用',
      cmdFindApp: '查找应用',
      cmdClearApp: '清除应用数据',
      cmdUninstall: '卸载应用',
      cmdForceStop: '强制停止应用',
      cmdStartApp: '启动应用',
      cmdSysVer: '系统版本',
      cmdModel: '设备型号',
      cmdBrand: '设备品牌',
      cmdAndroidId: 'Android ID',
      cmdIp: 'IP地址',
      cmdCpu: 'CPU信息',
      cmdMem: '内存信息',
      cmdStorage: '存储空间',
      cmdBattery: '电池状态',
      cmdResolution: '屏幕分辨率',
      cmdDensity: '屏幕密度',
      cmdCurrentAct: '当前Activity',
      supportedVars: '支持的变量替换：',
      varDeviceIp: '设备IP地址',
      varContainerId: '容器完整ID',
      varContainerShortId: '容器短ID (前12位)',
      varContainerName: '容器名称',
      varTimestamp: '当前时间戳',
      precautions: '注意事项：',
      note1: '命令会在每个选中的设备上依次执行',
      note2: '包含空格的文本参数需要用引号包裹',
      note3: '某些命令需要root权限才能执行',
      cmdInstall: '安装应用',
      cmdGetVersion: '获取系统版本'
    },

    stream: {
      streamManagementTitle: '流管理 (Streams)',
      forwardStream: '转发流(OBS)',
      p2pStream: '点对点(OBS)',
      inputRoomId: '输入房间号',
      refreshList: '刷新列表',
      pushCodeRoom: '推流码/房间',
      sourceIp: '源IP',
      bitrate: '码率 (Video/Audio)',
      noStreamInfo: '暂无流信息，请点击上方新建',
      distribute: '分发',
      disconnect: '断开',
      deviceCloudRelation: '设备-云机关系 (Device & Cloud Machines)',
      distributeToCloud: '分发到云机',
      activeDistribution: '已配置分发 (Active)',
      refreshStatus: '刷新状态',
      pushAddress: '推流地址',
      noActiveDistribution: '暂无分发流信息',
      inactiveDistribution: '失效分发 (Inactive/Unverified)',
      deviceOfflineOrNotExist: '设备离线或不存在',
      startP2P: '启动P2P',
      webrtcAddress: 'WEBRTC地址',
      rtmpAddress: 'RTMP地址',

      queryingStatus: '正在查询状态...',
      activeStatus: '推流中',
      idleStatus: '闲置',
      streamGuideHtml: `<div class="stream-guide" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-connection" style="margin-right: 8px;"></i> 转发流（OBS）
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>打开 OBS，进入 <b>设置 -> 直播</b>，服务选择 <b>自定义</b>。</li>
    <li>将首页展示的 WebRTC 推流地址 或 RTMP 推流地址 填入"服务器"栏（点击地址可一键复制）。</li>
    <li>在"推流码"栏填入自定义的推流码。点击确定后在 OBS 中点击 <b>开始直播</b>。</li>
    <li>回到本页面，点击 <b>新建流</b>，输入与推流码 <b>相同的名称</b> 后确认。
      <div style="color: #e6a23c; font-size: 12px; margin-top: 4px;"><i class="el-icon-info"></i> 命名一致后该条目将永远保留在列表中，推流状态均可见。</div>
    </li>
    <li>列表该项对应条目且状态变为 <b>活跃</b> 后，点击 <b>分发</b> 按钮，选择目标设备和云机。</li>
    <li>在云机上打开 <b>相机</b> 即可查看对应画面。</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-link" style="margin-right: 8px;"></i> 点对点（P2P / OBS）
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>切换到 <b>点对点(OBS)</b> 标签，点击 <b>添加 P2P 流</b>。</li>
    <li>在弹窗中选择 <b>设备</b> 和 <b>云机</b>，保存后列表会显示对应的 <b>监听地址</b>（点击可一键复制）。</li>
    <li>打开 OBS，进入 <b>设置 -> 直播</b>，服务选择 <b>自定义</b>，将监听地址填入"服务器"栏，无需填写推流码，点击确定。</li>
    <li>在 OBS 中点击 <b>开始直播</b>，然后在在列表中的 <b>启动 P2P</b> 按钮。</li>
    <li>在对应云机上打开 <b>相机</b> 即可查看实时画面。</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-video-camera" style="margin-right: 8px;"></i> PC 摄像头
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>切换到 <b>PC摄像头</b> 标签，点击 <b>添加摄像头推流</b>。</li>
    <li>在弹窗中选择 <b>设备</b>、<b>云机</b> 以及可用的 <b>摄像头</b>，配置分辨率与码率后保存。</li>
    <li>在列表中点击 <b>启动</b> 按钮，右侧预览区会自动启动展示实时画面。</li>
    <li>在对应云机上打开 <b>相机</b> 即可查看推流画面。</li>
  </ol>
</div>`
    },

    instance: {
      alipayTip: '请使用支付宝扫描二维码进行支付',
      authSync: '授权同步',
      confirmBuy: '确认购买',
      details: '详情',
      detailTitle: '实例详情',
      host: '归属主机',
      instanceIP: '实例 / IP',
      instances: '实例列表',
      loadingPackages: '正在加载套餐...',
      loadingQR: '正在生成支付二维码...',
      login: '登录',
      loginRequired: '需要登录后才能进行授权或操作',
      noInstance: '暂无实例',
      noPackages: '无套餐',
      purchaseRenew: '购买/续费授权',
      query: '查询',
      register: '注册',
      scanToPay: '扫码支付',
      searchPlaceholder: '搜索主机 IP 或实例',
      selectPackage: '选择套餐',
      statusExpired: '已过期',
      statusExpiring: '即将过期',
      statusNormal: '正常',
      subtotal: '小计',
      totalCount: '共有 ',
      instancePrefix: '实例',
      username: '账号',
      validUntil: '有效期至',
      password: '密码',
      rememberCredentials: '记住凭证',
      forgotPassword: '忘记密码',
      cancel: '取消',
      loggingIn: '登录中...',
      userRegistration: '用户注册',
      phone: '手机号',
      loginPassword: '登录密码',
      confirmPassword: '确认密码',
      verificationCode: '验证码',
      enterUsernameOrPhone: '请输入用户名/手机号',
      enterPassword: '请输入密码',
      enterPhone: '请输入手机号',
      enterPasswordAgain: '请再次输入密码',
      enterVerificationCode: '请输入验证码',
      sendVerificationCode: '发送验证码',
      registering: '注册中...',
      resetPassword: '重置密码',
      resettingPassword: '重置中...',
      phoneNumber: '手机号码',
      newPassword: '新密码',
      confirmNewPassword: '确认新密码',
      phoneVerificationCode: '手机验证码',
      noAccount: '还没有账号？',
      registerNow: '立即注册',
      instanceSlot: '实例位',
      selectedInstances: '已选实例',
      unit: '个'
    },
    network: {
      nodeManagement: '节点管理',
      nodeAllocation: '节点分配',
      domainFilter: '域名过滤',
      domainDirect: '域名直连',
      privateNic: '私有网卡',
      publicNic: '公有网卡',
      usageGuide: '使用说明',
      macvlanHint: 'MacVlan由于网卡混杂模式工作原理，不支持无线网卡，必须使用有线网络连接才可正常工作。',
      nodeManagementList: '节点管理列表',
      searchGroup: '搜索分组',
      selectedNodes: '已选节点',
      batchDeleteNodes: '批量删除节点',
      batchSpeedTest: '批量测速',
      type: '类型',
      alias: '别名',
      address: '地址',
      port: '端口',
      protocol: '协议',
      latency: '延迟',
      deviceList: '设备列表',
      assignVpc: '分配节点',
      addGroup: '添加分组',
      addDomain: '添加域名',
      addRule: '添加规则',
      batchClearVpc: '批量清空',
      clearDirect: '清空直连',
      clearFilter: '清空过滤',
      clearVpc: '清空节点',
      closeDns: '关闭DNS',
      closeDnsWhitelist: '白名单模式',
      createNic: '创建网卡',
      deleteGroup: '删除分组',
      deleteGroupNode: '删除',
      directDomain: '直连域名',
      domainRuleDesc: '域名匹配规则',
      domainRules: '域名规则',
      editGroupAlias: '编辑别名',
      filterDomain: '过滤域名',
      fullRuleDesc: '完整匹配规则',
      gateway: '网关',
      globalRules: '全局规则',
      index: '序号',
      keywordRuleDesc: '关键字匹配',
      machineIP: '云机IP',
      machineName: '云机名称',
      nicName: '网卡名称',
      nodeName: '节点名称',
      noDomainDirectRules: '暂无直连规则',
      noDomainRules: '暂无域名规则',
      noPublicNicData: '暂无公有网卡数据',
      noVpcContainers: '暂无容器',
      openDns: '开启DNS',
      privateNicDesc: '私有网卡（VPC网络）说明',
      privateNicDescDetail: '私有网络会在宿主机创建一个独立的虚拟网关，并为云机分配在该虚拟局域网内的内部 IP，用来隔离或桥接外部网络代理。',
      privateNicList: '私有网卡列表',
      publicNicDesc: '公有网卡（MacVlan）说明',
      publicNicDescDetail: '公有网卡会将云机的通讯接口直接绑定并透传到真实局域网，使云机能拿到你所处真实局域网内的 IP 地址。',
      queryContainerDomainDirect: '查询容器域名直连',
      queryContainerDomainFilter: '查询容器域名过滤',
      queryGlobalDomainFilter: '查询全局域名过滤',
      selectContainer: '选择容器',
      selectContainerOrQuery: '选择容器并配置',
      selectContainerToQuery: '选择容器进行查询',
      selectDeviceFirst: '请先选择左侧设备',
      selectVpcContainer: '选择 VPC 容器',
      setContainerDomainFilter: '设置容器域名过滤',
      setDomainDirect: '设置域名直连',
      setGlobalDomainFilter: '设置全局域名过滤',
      sourceGroup: '源分组',
      speedTest: '测速',
      updateGroup: '更新分组',
      updateMacVlan: '更新 MacVlan',
      updateMacVlanTitle: '更新 MacVLAN 配置',
      vpcNodes: 'VPC节点'
    },
    model: {
      simulatorOnlyHint: '本机型配置仅在使用 V3（模拟器）镜像时生效，容器镜像由于机制差异，部分可能不生效。',
      onlineModels: '在线机型',
      localModels: '本地机型',
      usageGuide: '使用说明',
      openDownloadDir: '打开下载目录',
      openLocalDir: '打开本地目录',
      selectedCount: '已选 {count} 个机型',
      noModelSelected: '未选机型',
      pushSelected: '推送到设备 ({count})',
      searchPlaceholder: '搜索机型',
      modelId: '机型ID',
      modelName: '机型名称',
      push: '编辑',
      pushed: '已推送',
      noModelData: '暂无机型数据',
      totalModels: '共 ',
      helpContent: `<div class="model-guide-container" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    📄 页面概览
  </div>
  <p style="margin: 0; padding-left: 8px;">本页面用于管理云机模拟器所使用的手机机型模板，支持查看线上机型、下载模板、编辑配置，以及将本地机型推送到设备。</p>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🌐 线上机型
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>展示服务器上所有可用的机型模板列表，包含机型 ID 和名称。</li>
    <li><b>下载：</b> 若本地尚未存有该机型模板，操作列会显示“下载”按钮，点击即可将模板下载到本地。</li>
    <li><b>编辑：</b> 若本地已有对应模板，操作列会显示“编辑”按钮，点击可修改机型的 overlay、prop 等配置项，保存后会自动生成新的机型文件并存入本地。</li>
    <li>右上角“打开本地机型下载目录”可快速打开模板的本地存储目录。</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    💾 本地机型
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>展示已存储在本地的全部机型模板，这些机型可直接推送到在线设备使用。</li>
    <li><b>推送（单个）：</b> 点击某一行的“推送”按钮，在弹窗中选择目标设备后点击“开始推送”即可。</li>
    <li><b>批量推送：</b> 勾选多个机型（或点击“全选”），再点击顶部“推送已选机型”按钮，统一推送到选中设备。</li>
    <li><b>刷新：</b> 重新扫描本地目录，更新机型列表。</li>
    <li><b>采集工具：</b> 扫描二维码下载手机信息采集 APK，用于从真机采集机型数据后导入本地。注意：仅支持采集对应 Android 版本的手机（如 Android 14 只能采集 Android 14 设备）。</li>
    <li>右上角“打开本地机型目录”可快速查看本地机型文件。</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📝 机型配置编辑
  </div>
  <p style="margin: 0; padding-left: 8px;">在线上机型页面点击“编辑”后进入配置编辑弹窗。可修改该机型的设备参数。配置分为以下几个部分：</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>model (模板名称)：</b> 只读，显示当前编辑的机型名称，不可修改。</li>
  </ul>
  
  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Overlay 参数（设备基础属性，影响云表现）</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>EMMCID:</b> eMMC 存储芯片唯一标识，空 <code> </code> 表示随机生成。</li>
    <li><b>EMMCSD:</b> eMMC 存储序列号，空 <code> </code> 表示随机生成。</li>
    <li><b>GRID:</b> GPU图形相关标识符，空 <code> </code> 表示随机生成。</li>
    <li><b>MEDIAID:</b> 媒体存储设备唯一标识，空 <code> </code> 表示随机生成。</li>
    <li><b>RNDID:</b> 随机设备标识符，空 <code> </code> 表示随机生成。</li>
    <li><b>SERIALNUMBER:</b> 设备序列号 (SN 码)，空 <code> </code> 表示随机生成。</li>
    <li><b>Aaid:</b> 匿名广告标识符 (Anonymous Advertising ID)，空 <code> </code> 表示随机生成。</li>
    <li><b>Android Id:</b> Android 系统唯一设备标识符，空 <code> </code> 表示随机生成。</li>
    <li><b>Bluetooth Name:</b> 设备蓝牙展示名称，空 <code> </code> 表示随机生成。</li>
    <li><b>GnssModel:</b> GPS/定位模组型号，例如 <code>MTK_M90_Default_MN_PDB_default</code>，影响定位识别，一般保持默认。</li>
    <li><b>Oaid:</b> 开放匿名设备标识符 (Open Anonymous Device Identifier)，空 <code> </code> 表示随机生成。</li>
    <li><b>OviLbs:</b> Overlay 底层 libs 版本标识，通常为数字 (如 <code>1</code>)，一般保持默认。</li>
    <li><b>Serial Number:</b> 设备序列号短字段，空 <code> </code> 表示随机生成。</li>
    <li><b>Vaid:</b> 厂商广告标识符 (Vendor Advertising ID)，空 <code> </code> 表示随机生成。</li>
    <li><b>Wifi List:</b> WiFi 扫描列表数量，数字类型（如 <code>10</code>），影响 WiFi 列表模拟条目数。</li>
    <li><b>Wifi Name:</b> 设备连接的 WiFi 名称，空 <code> </code> 表示随机生成。</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Prop 参数 (Android 系统属性，只读展示)</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>Ro.build.version.security_patch:</b> Android 安全补丁日期，格式 <code>YYYY-MM-DD</code>（如 <code>2023-12-05</code>），该字段为只读，不可修改。</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">关于 <code>空格</code> 占位符</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>字段值全为 <code>空格</code>，表示该字段在云机启动时会自动随机生成，无需手动填写。</li>
    <li>若需要固定某个字段的值，可将对应位置的 <code>空格</code> 替换为实际字符（需符合字段规定的字符集，如十六进制字段只能填 <code>0-9,a-f</code>）。</li>
    <li>输入了不符合规则的字符时，系统会自动将其还原为 <code>空格</code> 并给出提示。</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">保存规则</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>点击“保存”后，系统会以原机型名称 + 3 位随机数（如 <code>PJA110_182</code>）命名，生成一个新的本地机型文件，不会覆盖原装模板。</li>
    <li>保存成功后可在“本地机型” tab 中看到新生成的机型，并可直接推送到设备使用。</li>
    <li><span style="color: #F56C6C;">⚠️ 修改配置可能导致云机无法正常开机。不确定的字段请保持默认值，修改前建议记录原始值。</span></li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📤 推送设备
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>推送弹窗会列出当前所有在线设备，勾选目标设备后点击“开始推送”。</li>
    <li>推送后会自动给各设备在线状态，离线设备无法推送。</li>
    <li>支持同时推送多个机型到多台设备，推送完成后会展示成功/失败汇总。</li>
    <li>点击“刷新列表”可重新获取最新的在线设备状态。</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🔍 搜索与分页
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>顶部搜索框支持按机型名称关键词过滤，实时生效。</li>
    <li>列表默认每页显示 12 条，可通过底部分页组件翻页或调整每页数量。</li>
  </ul>
</div>`,
      configEdit: '编辑机型配置: {name}',
      configWarning: '修改机型配置可能影响云机的正常运行。',
      templateName: '模板名称',
      collectionToolDownload: '采集工具下载',
      scanQrToDownload: '扫码下载',
      collectionToolNote: '安装到手机上即可进行采集！',
      collectionToolWarning1: '重要！不可擅自提供给他人使用',
      collectionToolWarning2: '因此产生的风控概不负责。',
      selectPushDevice: '选择目标设备',
      modelsToPush: '待推送机型：',
      selectTargetDevices: '选择推送设备',
      onlineDeviceCount: '在线 {count} ',
      refreshList: '刷新设备',
      deviceIP: '设备IP',
      pushing: '正在推送 {current}/{total}',
      pushComplete: '推送完成 ({total})',
      current: '当前: ',
      startPush: '开始推送'
    },
    image: {
      usageGuide: '使用说明',
      guideContent: `<div class="image-guide-container" style="font-size: 13px; color: #606266; line-height: 2;">
  <h3 style="color: #303133; font-size: 16px; margin-bottom: 20px; display: flex; align-items: center;">
    <i class="el-icon-picture-outline" style="margin-right: 8px;"></i> 镜像管理使用说明
  </h3>

  <!-- 1. 功能介绍 -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    📄 功能介绍
  </div>
  <p style="margin-top: 0; margin-bottom: 12px;">镜像管理模块用于管理运行在设备上的 Android 系统镜像，支持从云端下载镜像到本地、将本地镜像上传到设备，以及查看和删除设备上已有的镜像。主要包含三个功能页签：</p>
  <ul style="padding-left: 20px; margin-bottom: 24px;">
    <li><b>在线镜像：</b> 列表云端可用镜像列表，可按设备型号、类型（模拟器/容器）、Android 版本等条件筛选，支持一键下载到本地或直接上传到设备。</li>
    <li><b>本地镜像：</b> 查看已下载到本地的镜像文件，可将本地镜像上传至指定设备，也可删除本地缓存文件以释放磁盘空间。</li>
    <li><b>设备镜像：</b> 选择某台设备，查看该设备上已安装的镜像列表，并可对其进行删除操作。</li>
  </ul>

  <!-- 2. 典型使用流程 -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    🔄 典型使用流程
  </div>
  <div style="margin-bottom: 30px; position: relative;">
    <ol style="padding-left: 10px; margin: 0; list-style: none;">
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">1</span> <div><b style="color: #409eff;">选择在线镜像</b><br><span style="color: #409eff;">在「在线镜像」列表中，找对你的设备和系统要求的镜像，核对版本和类型。</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">2</span> <div><b style="color: #409eff;">下载到本地</b><br><span style="color: #409eff;">点击「下载」按钮，将镜像文件缓存到本机存储，以便下次使用。</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">3</span> <div><b style="color: #409eff;">上传到设备</b><br><span style="color: #409eff;">在「本地镜像」或「在线镜像」中点击「上传到设备」，选中相应的设备完成推送。</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">4</span> <div><b style="color: #409eff;">验证安装</b><br><span style="color: #409eff;">切换到「设备镜像」页签，选择对应设备，确认镜像已成功安装在设备磁盘空间中。</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">5</span> <div><b style="color: #409eff;">启动云机</b><br><span style="color: #409eff;">回到「云机」页面，使用已上传的镜像创建云机并启动运行。</span></div></li>
    </ol>
  </div>

  <!-- 3. 模拟器 vs 容器：有什么区别？ -->
  <div style="font-weight: bold; margin-bottom: 12px; font-size: 14px;">模拟器 vs 容器：有什么区别？</div>
  <table style="width: 100%; border-collapse: collapse; margin-bottom: 24px; text-align: left; font-size: 12px;">
    <thead>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <th style="padding: 10px 0; width: 15%; color:#606266;">对比维度</th>
        <th style="padding: 10px 0; width: 42%; color:#409eff;">🖥️ 模拟器 (Simulator)</th>
        <th style="padding: 10px 0; width: 43%; color:#e6a23c;">📦 容器 (Container)</th>
      </tr>
    </thead>
    <tbody>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">运行原理</td>
        <td style="padding: 8px 0;">基于 QEMU 等虚拟机技术，完整模拟 ARM 硬件</td>
        <td style="padding: 8px 0;">基于 Linux 容器（如 Docker）轻量隔离，共享宿主机内核</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">启动速度</td>
        <td style="padding: 8px 0;">较慢，通常需要 50 秒以上</td>
        <td style="padding: 8px 0;">快，通常 3~10 秒即可启动</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">资源占用</td>
        <td style="padding: 8px 0;">较高，每实例需独立分配 CPU/内存</td>
        <td style="padding: 8px 0;">较低，多实例可共享宿主机资源，密度越高</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">并发数量</td>
        <td style="padding: 8px 0;">受限于宿主机性能，同时运行数量较少</td>
        <td style="padding: 8px 0;">没有硬屏障，单机支持数十甚至上百个实例</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">安卓兼容性</td>
        <td style="padding: 8px 0;">与真机行为高度一致，兼容性更好</td>
        <td style="padding: 8px 0;">部分依赖底层硬件的功能可能受限</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">图形界面</td>
        <td style="padding: 8px 0;">支持完整 GPU 渲染，画面流畅</td>
        <td style="padding: 8px 0;">视网显卡渲染，图形性能稍弱</td>
      </tr>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">适用场景</td>
        <td style="padding: 8px 0;">UI 测试、游戏运行、重度图形需求</td>
        <td style="padding: 8px 0;">自动化脚本、批量任务、高密度部署</td>
      </tr>
    </tbody>
  </table>

  <!-- 4. 选型建议 -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #e6a23c;">
    💡 选型建议
  </div>
  <div style="background-color: #f0f9eb; padding: 12px; border-radius: 4px; border-left: 4px solid #67c23a; margin-bottom: 8px;">
    <div style="display: flex; align-items: flex-start;">
      <i class="el-icon-success" style="color: #67c23a; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
      <div>
        <div style="font-weight: bold; color: #67c23a; margin-bottom: 4px;">推荐使用容器镜像的场景：</div>
        <div style="color: #67c23a;">需要大规模并发运行多个 Android 实例（如自动化脚本、批量任务）；对启动速度和资源利用率有较高要求；无需图形界面的无头运行场景。</div>
      </div>
    </div>
  </div>
  <div style="background-color: #fdf6ec; padding: 12px; border-radius: 4px; border-left: 4px solid #e6a23c; margin-bottom: 24px;">
    <div style="display: flex; align-items: flex-start;">
       <i class="el-icon-warning" style="color: #e6a23c; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
       <div>
         <div style="font-weight: bold; color: #e6a23c; margin-bottom: 4px;">推荐使用模拟器镜像的场景：</div>
         <div style="color: #e6a23c;">需要完整的 Android 图形界面体验；需要与真机行为高度一致的测试环境；部分对硬件加载有强依赖的应用场景。</div>
       </div>
    </div>
  </div>

  <!-- 5. 注意事项 -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #F56C6C;">
    ⚠️ 注意事项
  </div>
  <ul style="padding-left: 20px; color: #606266; margin-bottom: 0;">
    <li>镜像文件体积较大（通常 1GB 以上），下载前请确保本机磁盘空间充足。</li>
    <li>上传镜像到设备需要设备在线且网络稳定，上传过程中请勿断开连接。</li>
    <li>删除设备上的镜像后，使用该镜像的实例将无法再启动，请谨慎操作。</li>
    <li><b>P1 型号</b>设备不支持 Android 12 容器镜像，请选择 Android 10 或 Android 14。</li>
  </ul>
</div>`
    },
    addDevice: {
      title: '添加设备',
      scanDiscovery: '扫描发现',
      manualAdd: '手动添加',
      scanDevices: '扫描设备',
      scanning: '正在扫描...',
      autoDiscoverHint: '自动发现局域网内的设备',
      searchPlaceholder: '搜索设备IP、名称或ID',
      selectAllToggle: '全选 / 反选',
      devicesCount: '共 {total} 个设备，过滤后 {filtered} 个，已选 {selected} 个',
      deviceIP: '设备IP',
      deviceName: '设备名称',
      deviceID: '设备ID',
      version: '版本',
      status: '状态',
      added: '已添加',
      canAdd: '可添加',
      addSelectedDevices: '添加已选设备',
      noDevicesFound: '暂未发现设备，请点击扫描按钮',
      enterDeviceIP: '输入设备IP地址',
      multipleIPHint: '支持逗号或换行分隔多个IP，例如：192.168.1.100, 192.168.1.101',
      enterIPPlaceholder: '请输入设备IP地址，多个地址用逗号或换行分隔',
      addDevice: '添加设备',
      querying: '正在查询...',
      clear: '清空',
      successAdded: '成功添加 {count} 个设备',
      noResponseDevices: '以下设备无响应 ({count}个)：',
      noDeviceDiscovered: '未发现任何设备',
      noResponseLabel: '无响应的设备：',
      close: '关闭',
      // 批量任务相关翻译
      batchTask: '批量任务',
      executeNow: '立即执行',
      scheduledExecution: '定时执行',
      commandTemplate: '命令模板',
      executionHistory: '执行历史',
      addTemplate: '添加模板',
      variableReplacement: '变量替换',
      executing: '执行中',
      executionSuccess: '执行成功',
      executionFailed: '执行失败',
      onceSchedule: '一次性定时',
      periodicSchedule: '周期性定时',
      cronExpression: 'Cron表达式',
      targetDevices: '目标设备',
      executionResult: '执行结果',
      exportHistory: '导出历史',
      selectTargets: '选择目标',
      commandEditor: '命令编辑器',
      scheduleConfig: '定时配置',
      templateManager: '模板管理',
      historyViewer: '历史查看',
      enterCommand: '请输入命令',
      taskName: '任务名称',
      createTask: '创建任务',
      updateTask: '更新任务',
      deleteTask: '删除任务',
      enableTask: '启用任务',
      disableTask: '禁用任务',
      totalDevices: '共{total}个设备',
      successCount: '成功{count}个',
      failedCount: '失败{count}个',
      duration: '耗时',
      output: '输出',
      exportJSON: '导出JSON',
      exportCSV: '导出CSV'
    },
    menu: {
      hostManagement: '主机',
      cloudManagement: '云机',
      imageManagement: '镜像',
      modelManagement: '机型',
      networkManagement: '网络',
      backupManagement: '备份',
      instanceManagement: '实例',
      streamManagement: '流媒体',
      batchTask: '批量任务',
      aiAssistant: 'AI助理',
      rpaAgent: 'RPA Agent',
      logout: '退出',
      about: '关于',
      interconnectedCloudMachines: '互联云机'
    },
    update: {
      checkUpdate: '检查更新',
      newVersionAvailable: '发现新版本',
      currentVersion: '当前版本',
      noUpdate: '已是最新版本',
      downloading: '检查中...',
      about: '关于',
      autoCheckUpdate: '自动检查更新',
      autoDownloadUpdate: '自动下载更新',
      hasNewVersion: '有新版本',
      updateSettings: '更新设置'
    },
    dialog: {
      addDeviceTitle: '选择上传设备'
    },
    app: {
      title: '魔云腾-V3-客户端',
      description: 'ARM边缘计算设备管理客户端'
    },
    cloudMachine: {
      currentName: '当前名称',
      newName: '新名称',
      enterNewName: '请输入新名称',
      running: '运行中',
      backupName: '备份名称',
      remark: '备注',
      created: '已创建',
      restarting: '重启中',
      renameCloudMachine: '修改云机名称',
      switchBackup: '切换备份',
      batchAddToGroup: '批量添加分组',
      slotMode: '坑位',
      batchMode: '批量',
      setStream: '设置推流',
      updateImage: '更新镜像',
      apiDetails: 'API详情',
      renameDevice: '修改云机名称',
      shake: '摇一摇',
      setGPS: '设置GPS',
      uploadGoogleCert: '上传google证书',
      restart: '重启',
      shutdown: '关机',
      setS5Agent: '设置S5代理',
      closeS5Agent: '关闭安卓内SOCKS5代理',
      fileUpload: '上传文件',
      oneKeyNewDevice: '一键新机',
      switchModel: '切换机型',
      resetContainer: '重置容器'
    },
    aiAssistant: {
      bold: '加粗',
      clearFormat: '清除格式',
      clearHistoryConfirm: '确定要清空当前历史对话吗？此操作不可恢复。',
      continueFailed: '继续对话失败。',
      copiedClipboard: '已复制到剪贴板。',
      copiedText: '文本已复制。',
      copy: '复制',
      copyFailed: '复制失败。',
      defaultPrompt: '你是人工智能助手。',
      deleteModelFailed: '删除模型失败。',
      deleteModelSuccess: '模型删除成功。',
      deviceResetFailed: '设备重置失败。',
      deviceResetSuccess: '设备重置成功。',
      editorPlaceholder: '请输入对话内容，或将文件拖至此处 / 粘贴输入（Shift+Enter 换行）',
      fetchModelFailed: '获取模型列表失败。',
      fileProcessFailed: '文件处理失败。',
      fileSelectFailed: '文件选择失败。',
      fileTooLarge: '文件过大。',
      generationStopped: '生成已停止。',
      historyCleared: '对话记录已清空。',
      italic: '斜体',
      llmStartFailed: 'LLM 服务启动失败。',
      llmStartSuccess: 'LLM 服务启动成功。',
      llmStopFailed: 'LLM 服务停止失败。',
      llmStopSuccess: 'LLM 服务已停止。',
      loadModelFailed: '模型加载失败。',
      modelAlreadyRunning: '该模型已在运行。',
      modelStartSuccess: '模型启动成功。',
      newChatStarted: '已开启新对话。',
      notImgVideo: '非图片或视频文件。',
      orderedList: '有序列表',
      pastedFiles: '粘贴的文件',
      regenerateFailed: '重新生成失败。',
      restoredDefault: '已恢复默认设置。',
      saveFailed: '保存失败。',
      selectDeviceAndModel: '选择设备和模型。',
      selectDeviceFirst: '请先选择设备。',
      selectModelFirst: '请先选择模型。',
      selectedFiles: '选中的文件',
      sendFailed: '发送失败。',
      settingsSaved: '设置已保存。',
      startingModel: '正在启动模型...',
      strikethrough: '删除线',
      underline: '下划线',
      unorderedList: '无序列表',
    },

  },
  'en-US': {
    common: {
      sortBy: 'Sort By',
      createTimeSort: 'Create Time',
      selectedItems: '{count} selected',
      backupName: 'Backup Name',
      cloudMachine: 'Machine',
      statusInfo: 'Status Info',
      statusLabel: 'Status',
      projection: 'Projection',
      confirm: 'Confirm',
      cancel: 'Cancel',
      save: 'Save',
      delete: 'Delete',
      close: 'Close',
      settings: 'Settings',
      upload: 'Upload',
      refresh: 'Refresh',
      cloudMachine: 'Cloud Machine',
      image: 'Image',
      device: 'Device',
      taskQueue: 'Task Queue',
      deleteSelected: 'Delete',
      addDevice: 'Add Device',
      bindHost: 'Bind Host',
      unbindHost: 'Unbind',
      syncAuth: 'Sync License',
      batchUpgradeAPI: 'Upgrade API',
      clearClientData: 'Clear Client Data',
      batchCleanDisk: 'Clean Disk',
      logout: 'Logout',
      onlineDevices: 'Online Devices',
      allDevices: 'All Devices',
      offlineDevices: 'Offline Devices',
      hostFirmwareVersion: 'Firmware Version',
      nvmeStorage: 'NVME Storage',
      bindStatus: 'Bind',
      apiVersion: 'API Version',
      operation: 'Operation',
      view: 'View',
      upgrade: 'Upgrade',
      loading: 'Loading',
      current: 'Current',
      latest: 'Latest',
      group: 'Group',
      defaultGroup: 'Default Group',
      unknown: 'Unknown',
      phoneNumber: 'Phone Number',
      verificationCode: 'Verification Code',
      submit: 'Submit',
      unbindWarning: 'Warning: All cloud instances purchased under this host will be cleared after unbinding. You need to repurchase cloud instances to use them again. (Data will not be cleared, please clean it manually if needed). This operation is irreversible, please be aware of the risks and operate carefully!',
      sendCode: 'Send Code',
      resendCode: 'Resend',
      addGroup: 'Add Group',
      editGroup: 'Edit Group',
      deleteGroup: 'Delete Group',
      create: 'Create',
      enterNewGroupName: 'Enter new group name',
      groupNameCannotBeEmpty: 'Group name cannot be empty',
      collapseRightSidebar: 'Collapse',
      hostInfo: 'Host Info',
      refreshHostInfo: 'Refresh Host Info',
      basicInfo: 'Basic Info',
      v3DeviceInfo: 'V3 Device Info',
      hostName: 'Host Name',
      hostIP: 'Host IP',
      deviceID: 'Device ID',
      deviceVersion: 'Device Version',
      deviceModel: 'Device Model',
      cpuTemp: 'CPU Temp',
      cpuLoad: 'CPU Load',
      memoryUsage: 'Memory Usage',
      storageUsage: 'Storage Usage',
      ipAddress: 'IP Address',
      macAddress: 'MAC Address',
      passwordManagement: 'Password Management',
      setPassword: 'Set Password',
      closePassword: 'Close Password',
      passwordHint: 'Note: After setting password, you need to enter password when accessing v3 API on device port 8000',
      sdkUpgrade: 'SDK Upgrade',
      currentSDKVersion: 'Current SDK Version',
      latestSDKVersion: 'Latest SDK Version',
      upgradeSDK: 'Upgrade SDK',
      upgrading: 'Upgrading',
      cleanDiskData: 'Clean Disk Data',
      latitude: 'Latitude',
      longitude: 'Longitude',
      selectDeviceFirst: 'Please select a device first',
      loadFailed: 'Load Failed',
      time: 'Build Time',
      type: 'System Type',
      dockerNetworkList: 'Docker Network List',
      addMacvlanNetwork: 'Add Macvlan Network',
      editNetwork: 'Edit Network',
      networkName: 'Network Name',
      physicalInterface: 'Physical Interface',
      subnet: 'Subnet',
      gateway: 'Gateway',
      ipRange: 'IP Range',
      isolationMode: 'Isolation Mode',
      enablePrivateIsolation: 'Enable Private Isolation',
      isolationTip: 'When enabled, containers in this network can only communicate with other containers in the same network',
      enterNetworkName: 'Enter network name, e.g.: mynetwork',
      enterPhysicalInterface: 'Enter physical interface, e.g.: eth0',
      enterSubnet: 'Enter subnet, e.g.: 192.168.1.0/24',
      enterGateway: 'Enter gateway, e.g.: 192.168.1.1',
      enterIPRange: 'Enter IP range, e.g.: 192.168.1.100/28',
      apiDetails: 'API Details',
      slot: 'Slot',
      instanceName: 'Instance Name',
      deviceIP: 'Device IP',
      deviceVersion: 'Device Version',
      portMappingInfo: 'Port Mapping Info',
      service: 'Service',
      portMapping: 'Port Mapping',
      accessAddress: 'Access Address',
      clickToCopy: 'Click on the access address to copy to clipboard',
      close: 'Close',
      setS5Proxy: 'Set S5 Proxy',
      cloudMachineName: 'Cloud Machine Name',
      s5Info: 'S5 Info',
      s5InfoFormat: 'Format: address:port:username:password (username and password are optional)',
      s5InfoExample: 'E.g.: 192.168.1.1:22:root:123456 or 192.168.1.1:22',
      parseAndFill: 'Parse and Fill',
      s5ServerAddress: 'S5 Server Address',
      s5Port: 'S5 Port',
      enterS5ServerAddress: 'Enter S5 server address',
      enterS5Port: 'Enter S5 port',
      dnsMode: 'DNS Resolution Mode',
      localDNS: 'Local DNS Resolution',
      serverDNS: 'Server DNS Resolution (Default)',
      submitting: 'Submitting...',
      submitNow: 'Submit Now',
      batchCreateCloudMachine: 'Batch Create Cloud Machine',
      createCloudMachine: 'Create Cloud Machine',
      googleGreen: 'Google Green',
      realParameters: 'Real Parameters',
      simulator: 'Simulator',
      container: 'Container',
      oldVersionImage: 'Old Version Image',
      deviceRemainingSpace: 'Device Remaining Space',
      insufficientSpace: 'Insufficient space, may fail to start/create',
      apiVersion: 'API Version',
      latestVersion: 'Latest',
      notLatestVersion: 'Not the latest version, upgrade to support new features',
      androidVersion: 'Android Version',
      pSeries: 'P Series',
      otherSeries: 'Other Series',
      pleaseSelectDevice: 'Please select device',
      customImage: 'Custom Image',
      imageAddress: 'Image Address',
      pleaseSelect: 'Please select',
      name: 'Name',
      cloudMachineCount: 'Cloud Machine Count',
      slot: 'Slot',
      normal: 'Normal',
      expiringSoon: 'Expiring Soon',
      expired: 'Expired',
      noInstance: 'No Instance',
      singleSlotCount: 'Single Slot Count',
      quantity: 'Quantity',
      resolution: 'Resolution',
      customResolution: 'Custom Resolution',
      deviceWidth: 'Device Width',
      deviceHeight: 'Device Height',
      resolutionWarning: 'Note: Custom resolution may cause style adaptation issues',
      dnsType: 'DNS Type',
      dnsAddress: 'DNS Address',
      aliDNS: 'Ali DNS(223.5.5.5)',
      googleDNS: 'Google(8.8.8.8)',
      custom: 'Custom',
      sandboxMode: 'Sandbox Mode?',
      enable: 'Enable',
      disable: 'Disable',
      secureMode: 'Secure Mode',
      dataDiskSize: 'Data Disk Size',
      networkManagement: 'Network Management',
      selectGroup: 'Select Group',
      selectNode: 'Select Node',
      specifiedNode: 'Specified Node',
      randomNode: 'Random Node',
      networkCardType: 'Network Card Type',
      privateNetworkCard: 'Private Network Card',
      publicNetworkCard: 'Public Network Card',
      noMacVlanDetected: 'No MacVlan configuration detected, please go to Network Management - Public Network Card to create',
      enterStartIPAddress: 'Enter start IP address, e.g.: 10.10.0.253',
      batchCreateTip: 'Tip: Batch create {count} machines, IP range: {range}',
      containerName: 'Container Name',
      enterContainerNamePrefix: 'Enter container name prefix, e.g.: T000',
      singleSlotCountLabel: 'Single Slot Count',
      imageCategory: 'Image Category',
      onlineImage: 'Online Image',
      localImage: 'Local Image',
      imageSelection: 'Image Selection',
      pleaseSelectLocalImage: 'Please select local image',
      customImageAddress: 'Custom Image Address',
      enterImageAddress: 'Enter image address, e.g.: registry.example.com/myimage:tag',
      creationMethod: 'Creation Method',
      cacheToLocal: 'Cache to Local',
      cacheToLocalTip: 'When checked, online image will be downloaded to local first, then pushed to device to create cloud machine',
      s5Proxy: 'S5 Proxy',
      pleaseSelectProxyType: 'Please select proxy type',
      noProxy: 'No Proxy',
      localDNSParsing: 'Local DNS Parsing(tun2socks)',
      serverDNSParsing: 'Server DNS Parsing(tun2proxy)',
      lockScreenPassword: 'Lock Screen Password',
      noLockScreen: 'No lock screen if left empty',
      port: 'Port',
      enterIP: 'Enter IP',
      enterPort: 'Enter Port',
      user: 'User',
      enterUser: 'Enter User',
      enterPassword2: 'Enter Password',
      ipAddress: 'IP Address',
      enterStartIP: 'Enter start IP address, e.g.: 192.168.1.100',
      dnsAddress: 'DNS Address',
      pleaseSelectDNSAddress: 'Please select DNS address',
      aliCloud: 'Alibaba Cloud',
      enterCustomDNSAddress: 'Enter custom DNS address',
      sandboxSize: 'Sandbox Size',
      defaultResolution: 'Model Default Resolution',
      width: 'Width',
      height: 'Height',
      modelCountry: 'Model Country',
      pleaseSelectModelCountry: 'Please select model country, supports search',
      modelType: 'Model Type',
      onlineModel: 'Online Model',
      localModel: 'Local Model',
      backupModel: 'Backup Model',
      phoneModel: 'Phone Model',
      pleaseSelectPhoneModel: 'Please select phone model',
      random: 'Random',
      pleaseSelectLocalModel: 'Please select local model',
      pleaseSelectBackupModel: 'Please select backup model',
      advancedOptions: 'Advanced Options',
      enableMagisk: 'Enable Magisk',
      enableGMS: 'Enable GMS',
      randomSystemFiles: 'Random System Files',
      networkCardSelection: 'Network Card Selection',
      pleaseSelectNetworkCard: 'Please select network card',
      firmwareVersion: 'Firmware Version',
      modelVersion: 'Model Version',
      deviceUptime: 'Device Uptime',
      memoryTotal: 'Total Memory',
      memoryUsed: 'Memory Used',
      diskTotal: 'Total Disk',
      diskUsed: 'Disk Used',
      diskModel: 'Disk Model',
      diskTemp: 'Disk Temp',
      diskRead: 'Disk Read',
      diskWrite: 'Disk Write',
      networkIP: 'Network IP',
      networkSpeed: 'Network Speed',
      eth0Network: 'ETH0 Network',
      syncAuthLogin: 'Authorization Login',
      username: 'Username',
      password: 'Password',
      enterUsername: 'Enter username',
      enterPassword: 'Enter password',
      rememberCredentials: 'Remember credentials',
      login: 'Login',
      loggingIn: 'Logging in...',
      register: 'Register',
      enterUsernameAndPassword: 'Please enter username and password',
      loginSuccess: 'Login successful',
      syncAuthFailed: 'Authorization sync failed',
      userRegistration: 'User Registration',
      phoneNumber: 'Phone',
      loginPassword: 'Password',
      confirmPassword: 'Password',
      phoneVerificationCode: 'Code',
      enterPhoneNumber: 'Enter phone number',
      enterLoginPassword: 'Enter login password',
      enterPasswordAgain: 'Enter password again',
      enterVerificationCode: 'Enter verification code',
      getVerificationCode: 'Get Code',
      retryAfterSeconds: 's retry',
      registering: 'Registering...',
      enterPhone: 'Please enter phone number',
      enterPassword: 'Please enter password',
      enterConfirmPassword: 'Please enter confirm password',
      passwordMismatch: 'Passwords do not match',
      enterVCode: 'Please enter verification code',
      getVCodeFirst: 'Please get verification code first',
      registerSuccess: 'Registration successful',
      registerFailed: 'Registration failed',
      vcodeSentSuccess: 'Verification code sent successfully',
      enterCorrectPhone: 'Please enter a valid phone number',
      syncAuthCompletedForAllDevices: 'Sync authorization completed for all {count} devices',
      syncAuthPartialSuccess: 'Partial success: {count} devices authorized',
      syncAuthAllFailed: 'All devices sync authorization failed',
      notLoggedIn: 'Not Logged In',
      unbound: 'Unbound',
      bound: 'Bound',
      boundByOthers: 'Others',
      deviceUpgradeSuccess: 'Device {ip} upgraded successfully',
      deviceUpgradeSuccessWithMessage: 'Device {ip} upgraded successfully: {message}',
      addNewGroup: 'Add New Group',
      deleteCurrentGroup: 'Delete Current Group',
      moreActions: 'More',
      openProjection: 'Open Projection',
      updateImage: 'Update Image',
      batchActions: 'Batch Actions',
      batchRestart: 'Batch Restart',
      batchReset: 'Batch Reset',
      batchProjection: 'Batch Projection',
      batchShutdown: 'Batch Shutdown',
      batchDelete: 'Batch Delete',
      batchSwitchBackup: 'Batch Switch Backup',
      deviceDetails: 'Device Details',
      setDevicePassword: 'Set Device Password',
      deviceAuthentication: 'Device Authentication',
      unknownDevice: 'Unknown Device',
      addDevicesToGroup: 'Add Devices to Group',
      selectPushDevice: 'Select Push Device',
      deviceUpgrade: 'Device Upgrade',
      noDeviceSelected: 'No Device Selected',
      refreshCloudMachines: 'Refresh Machines',
      instance: 'Instance',
      image: 'Image',
      network: 'Network',
      host: 'Host',
      refreshImages: 'Refresh Images',
      refreshNetworkList: 'Refresh Network List',
      slot: 'Slot',
      create: 'Create',
      instanceName: 'Instance Name',
      ipAddress: 'IP Address',
      systemImage: 'System Image',
      createTime: 'Create Time',
      status: 'Status',
      running: 'Running',
      shutdown: 'Shutdown',
      created: 'Created',
      restarting: 'Restarting',
      model: 'Model',
      none: 'None',
      startUp: 'Start',
      operationConfirm: 'Operation Confirm',
      confirmAction: 'Are you sure to {action} instance {name}?',
      actionSuccess: '{action} successful',
      actionFailed: '{action} failed: {error}',
      slotNoMachine: 'Slot {slot} has no machine, please create one first',
      slotCannotDelete: 'Slot {slot} has no machine, cannot delete',
      noDownloadedImages: 'No downloaded images',
      pleaseSelectDeviceFirst: 'Please select a device on the left first',
      onlineImageName: 'Online Image Name',
      imageSize: 'Image Size',
      matchedWithOnlineImage: '✓ Matched with online image',
      notMatchedWithOnlineImage: '✗ Not matched with online image',
      onlineURL: 'Online URL',
      deviceURL: 'Device URL',
      refreshOnlineImages: 'Refresh Online Images',
      openLocalImageDirectory: 'Open Local Image Directory',
      deviceImage: 'Device Image',
      other: 'Other',
      simulator: 'Simulator',
      container: 'Container',
      androidVersion: 'Android Version',
      imageName: 'Image Name',
      updateContent: 'Update Content',
      download: 'Download',
      uploadToDevice: 'Upload to Device',
      filePath: 'File Path',
      size: 'Size',
      availableDeviceModels: 'Available Device Models',
      universal: 'Universal',
      noDevice: 'No Device',
      noCompatibleDevices: 'No Compatible Devices',
      pleaseSelectDeviceForUpload: 'Please select device for uploading image',
      refreshLocalImages: 'Refresh Local Images',
      noLocalCachedImages: 'No local cached images',
      noOnlineImages: 'No matching online images',
      filterImageName: 'Filter image name',
      filterImageURL: 'Filter image URL',
      compatibleImages: 'Compatible Images',
      imagePath: 'Image Path',
      updateImageWarning: 'Update by image method, normal boot is not guaranteed',
      enterCustomImageAddress: 'Enter custom image address',
      batchUpdateImage: 'Batch Update Image',
      updateVersion: 'Update Version',
      versionType: 'Version Type',
      v3Simulator: 'V3 (Simulator)',
      v2Container: 'V2 (Container)',
      noImageForVersion: 'No image for this version, you can choose custom image',
      customAddress: 'Custom Address',
      enterImageURL: 'Enter image address',
      confirmUpdate: 'Confirm Update',
      totalCloudMachines: '{count} cloud machines',
      ipTestTitle: 'IP Connection Test',
      testResult: 'Test Result',
      startTest: 'Start Test',
      testing: 'Testing...',
      importantTip: '⚠️ Important Notice',
      ensureIPInSubnet: 'Ensure IP is within subnet range',
      ensureIPNotUsed: 'Make sure the IP address is not occupied',
      ipConflictWarning: 'otherwise it may cause IP conflict and access failure',
      pleaseSelectImage: 'Please select image',
      batchUpdateImageBtn: 'Batch Update Image',
      setIPLocation: 'Set IP Location',
      locationIP: 'Location IP',
      leaveEmptyForCurrentIP: 'Leave empty to use current cloud machine IP',
      countryRegion: 'Country/Region',
      pleaseSelectCountryRegion: 'Please select country/region',
      setMacVlanIP: 'Set MacVlan IP',
      uploadGoogleCert: 'Upload Google Certificate',
      supportedFormats: 'Supported formats',
      clickToSelectCert: 'Click to select certificate file',
      reselect: 'Reselect',
      uploading: 'Uploading...',
      confirmUpload: 'Confirm Upload',
      searchGroupOrIP: 'Search group name or device IP',
      dropToGroup: 'Drop to this group',
      cloudMachineManagement: 'Cloud Machine Management',
      cloudMachineColon: 'Machine: {ip}',
      batchManagement: 'Batch Management',
      refreshCloudMachine: 'Refresh Machines',
      stopControl: 'Stop Control',
      batchControl: 'Batch Control',
      batchCloseProjection: 'Batch Close Projection',
      batchUpload: 'Batch Upload',
      batchNewDevice: 'Batch New Device',
      zoom: 'Zoom',
      layout: 'Layout',
      grid: 'Grid',
      list: 'List',
      switchToHorizontal: 'Switch to Horizontal',
      switchToVertical: 'Switch to Vertical',
      vertical: 'Portrait',
      horizontal: 'Landscape',
      switching: 'Switching...',
      running: 'Running',
      backupCount: '{count} Backups',
      empty: 'Empty',
      booting: 'Booting...',
      shutdownClickBackup: 'Shutdown | Click to switch backup',
      emptySlot: 'Empty Slot',
      switchBackupBtn: 'Switch',
      copy: 'Copy',
      slotNoMachineCannotDelete: 'Slot {slot} has no machine, cannot delete',
      selectedCloudMachineList: 'Selected Machines ({count})',
      noSelectedCloudMachine: 'No selected machines',
      pleaseSelectCloudMachine: 'Please select machines from the tree on the left',
      openProjectionBtn: 'Open Projection',
      slot: 'Slot',
      instanceNameLabel: 'Instance Name',
      systemImage: 'System Image',
      createTime: 'Create Time',
      statusLabel: 'Status',
      shutdownStatus: 'Shutdown',
      restartingStatus: 'Restarting',
      modelLabel: 'Model',
      noneLabel: 'None',
      operationLabel: 'Operation',
      create: 'Create',
      selectDevicesToAdd: 'Select devices to add to this group:',
      selectAllLabel: 'Select All',
      defaultGroup: 'Default Group',
      allDevicesInGroup: 'All devices are already in this group',
      addCount: 'Add ({count})',
      copyCloudMachine: 'Copy Cloud Machine',
      cloudMachineName: 'Machine Name',
      targetSlot: 'Target Slot',
      targetSlotHint: 'Target slot number to copy to',
      copyCount: 'Copy Count',
      confirmCopy: 'Confirm Copy',
      batchCloseProjectionSuccess: 'Batch close projection successful',
      batchCloseProjectionFailed: 'Batch close projection failed',
      batchUploadFileTo: 'Batch Upload Files to Cloud Machines',
      closeAll: 'Close All',
      noRunningCloudMachine: 'No running machines selected, batch new device only supports running machines',
      imageSelection: 'Image Selection',
      searchIP: 'Search IP',
      all: 'All',
      group: 'Group',
      noDeviceSelected: 'No Device Selected',
      deviceNetworkLatency: 'Device Network Latency',
      networkName: 'Network Name',
      networkID: 'Network ID',
      subnet: 'Subnet',
      driver: 'Driver',
      scope: 'Scope',
      gateway: 'Gateway',
      ipRange: 'IP Range',
      containerCount: 'Containers',
      private: 'Private',
      yes: 'Yes',
      no: 'No',
      edit: 'Edit',
      delete: 'Delete',
      noNetworkInfo: 'No network information',
      confirm: 'Confirm',
      cancel: 'Cancel',
      // OpenCecs
      cloudInstanceManagement: 'Cloud Instance Management',
      login: 'Login',
      logout: 'Logout',
      allInstances: 'All',
      running: 'Running',
      stopped: 'Stopped',
      expired: 'Expired',
      refreshing: 'Refreshing...',
      refresh: 'Refresh',
      loadingData: 'Loading...',
      noInstanceData: 'No instance data',
      instanceID: 'Instance ID',
      instanceNameLabel: 'Instance Name',
      coreBoard: 'Core Board',
      memory: 'Memory',
      system: 'System',
      ipAddress: 'IP Address',
      statusLabel: 'Status',
      billingMode: 'Billing',
      expiryTime: 'Expiry',
      operationLabel: 'Actions',
      start: 'Start',
      stop: 'Stop',
      restart: 'Restart',
      loginFirst: 'Please login to view your cloud instances',
      loginNow: 'Login Now',
      welcomeOpenCecs: 'Welcome to OpenCecs - Leading Cloud Service Provider',
      phoneTab: 'Phone',
      accountTab: 'Account',
      // StreamManagement
      streamManagement: 'Stream Management',
      forwardStreamOBS: 'Forward (OBS)',
      p2pStreamOBS: 'P2P (OBS)',
      pcCamera: 'PC Camera',
      userGuide: 'User Guide',
      enterRoomNumber: 'Enter room number',
      newStream: 'New Stream',
      refreshList: 'Refresh List',
      streamCode: 'Stream Code/Room',
      sourceIP: 'Source IP',
      bitrate: 'Bitrate (Video/Audio)',
      noStreamInfo: 'No stream info, click above to create',
      active: 'Active',
      idle: 'Idle',
      distribute: 'Distribute',
      disconnect: 'Disconnect',
      addP2PStream: 'Add P2P Stream',
      pushAddress: 'Push Address',
      deviceIP: 'Device IP',
      cloudMachine: 'Cloud Machine',
      listenAddress: 'Listen Address',
      startP2P: 'Start P2P',
      stopP2P: 'Stop P2P',
      noP2PStreamInfo: 'No P2P stream info, click above to add',
      addCameraStream: 'Add Camera Stream',
      camera: 'Camera',
      resolutionBitrate: 'Resolution/Bitrate',
      defaultCamera: 'Default Camera',
      streaming: 'Streaming',
      streamStopped: 'Stopped',
      noCameraStream: 'No camera stream, click above to add',
      cameraPreview: 'Camera Live Preview',
      deviceCloudRelation: 'Device & Cloud Machines',
      distributeToCloud: 'Distribute to Cloud',
      autoRefreshCapture: 'Auto refreshing (capture every {interval}s)',
      autoPreviewAfterStart: 'Auto preview after starting stream',
      capturingFrame: 'Capturing frame...',
      refreshingFrame: 'Refreshing...',
      startStreamToPreview: 'Start camera stream to auto preview',
      configuredDistribution: 'Configured Distribution (Active)',
      refreshStatus: 'Refresh Status',
      protocol: 'Protocol',
      queryingStatus: 'Querying status...',
      noActiveDistribution: 'No active distribution',
      modifyStream: 'Modify Stream',
      projection: 'Projection',
      inactiveDistribution: 'Inactive Distribution',
      statusInfo: 'Status Info',
      noInactiveDistribution: 'No inactive distribution',
      unverified: 'Unverified',
      distributeStreamDialog: 'Distribute Stream to Cloud',
      stream: 'Stream',
      selectOrEnterRoom: 'Select or enter room number',
      device: 'Device',
      selectDevice: 'Select Device',
      selectCloudMachine: 'Select Cloud Machine',
      resolution: 'Resolution',
      auto: 'Auto',
      confirmDistribute: 'Confirm Distribute',
      addP2PStreamDialog: 'Add P2P Stream',
      p2pPort: 'P2P Port',
      portOccupied: 'Port is occupied, please choose another',
      remark: 'Remark',
      optionalDefault: 'Optional, default p2p-port',
      fillPort: 'Please fill in the port',
      addCameraStreamDialog: 'Add Camera Stream',
      devicePort: 'Device Port',
      selectCloudMachineFirst: '(Please select cloud machine first)',
      selectCamera: 'Select Camera',
      none: 'None',
      width: 'Width',
      height: 'Height',
      fpsAndBitrate: 'FPS/Bitrate',
      confirmAdd: 'Confirm Add',
      // InterconnectedCloudMachines
      selected: '{count} Selected',
      totalMachines: '{count} Cloud Machines',
      batchDeleteBtn: 'Batch Delete',
      stopControlBtn: 'Stop Control',
      batchControlBtn: 'Batch Control',
      addCloudMachine: 'Add Cloud Machine',
      clickToProject: 'Click to Project',
      noMachineHint: 'No machines, click "Add Cloud Machine" to import via key',
      addMachineDialog: 'Add Cloud Machine',
      enterKeyHint: 'Enter keys, separate multiple keys with comma or newline',
      confirmAddBtn: 'Confirm Add',
      // BatchTaskManagement
      selectDevices: '📱 Select Devices',
      selectedCount: '{count} Selected',
      searchDeviceOrContainer: 'Search device IP or container',
      selectAllBtn: 'Select All',
      clearBtn: 'Clear',
      openAll: 'Open All',
      closeAllBtn: 'Close All',
      totalDevices: '{count} Total',
      projectionBtn: 'Project',
      noAvailableDevice: 'No available devices',
      batchOperation: '⚡ Batch Operations',
      batchExecCommand: '📝 Batch Execute Command',
      inputADBCommand: 'Enter ADB Command:',
      viewCommandExample: 'View Command Examples',
      quickCommands: '🔖 Quick Commands',
      clickToFill: 'Click button to fill',
      basicOps: '📱 Basic Operations',
      appManagement: '📦 App Management',
      systemInfo: '🔧 System Info',
      loopCount: 'Loop Count:',
      times: 'times',
      executing: 'Executing (Round {current}/{total})...',
      executeNow: '🚀 Execute Now',
      stopBtn: '⏹ Stop',
      execResult: '📊 Execution Results',
      roundSeparator: 'Round {loop} / {total}',
      errorPrefix: 'Error',
      closeBtn: 'Close',
      adbCommandRef: '📖 ADB Command Reference',
      // Command Names
      cmdInputText: 'Input Text',
      cmdTapScreen: 'Tap Screen',
      cmdSwipeScreen: 'Swipe Screen',
      cmdHomeKey: 'Home Key',
      cmdBackKey: 'Back Key',
      cmdPowerKey: 'Power Key',
      cmdVolUp: 'Volume Up',
      cmdVolDown: 'Volume Down',
      cmdScreenshot: 'Screenshot',
      cmdRecordStart: 'Start Recording',
      cmdOpenSettings: 'Open Settings',
      cmdOpenDial: 'Open Dialer',
      cmdListApps: 'List Apps',
      cmdListSysApps: 'List System Apps',
      cmdList3rdApps: 'List 3rd-party Apps',
      cmdFindApp: 'Find App',
      cmdClearAppData: 'Clear App Data',
      cmdUninstallApp: 'Uninstall App',
      cmdForceStopApp: 'Force Stop App',
      cmdStartApp: 'Start App',
      cmdSysVersion: 'System Version',
      cmdDeviceModel: 'Device Model',
      cmdDeviceBrand: 'Device Brand',
      cmdAndroidId: 'Android ID',
      cmdIpAddress: 'IP Address',
      cmdCpuInfo: 'CPU Info',
      cmdMemInfo: 'Memory Info',
      cmdStorage: 'Storage Space',
      cmdBattery: 'Battery Status',
      cmdScreenRes: 'Screen Resolution',
      cmdScreenDensity: 'Screen Density',
      cmdCurrentActivity: 'Current Activity',
      // Command Reference
      commonAdbExamples: 'Common ADB Command Examples:',
      inputText: 'Input Text',
      tapCoordinate: 'Tap Coordinate (x=500, y=800)',
      swipeScreen: 'Swipe Screen',
      pressHomeKey: 'Press Home Key (3=HOME, 4=BACK, 26=POWER)',
      listAllPackages: 'List All Packages',
      installApp: 'Install App',
      uninstallApp: 'Uninstall App',
      startSettings: 'Start Settings',
      checkBatteryInfo: 'Check Battery Info',
      screenshot: 'Screenshot',
      getSystemVersion: 'Get System Version',
      supportedVars: 'Supported Variables:',
      deviceIpAddress: 'Device IP Address',
      containerFullId: 'Full Container ID',
      containerShortId: 'Short Container ID (first 12 chars)',
      containerName: 'Container Name',
      currentTimestamp: 'Current Timestamp',
      precautions: 'Precautions:',
      cmdExecSequentially: 'Commands will be executed sequentially on selected devices.',
      quotesNeededForSpaces: 'Parameters with spaces need to be enclosed in quotes.',
      rootRequired: 'Some commands require root access to execute.',
      // AI Assistant
      onlineDevicesLabel: 'Online Devices',
      noAIDevice: 'No AI-compatible devices (requires r1q or eces-rk3588-rk1828)',
      modelManagement: 'Model Management',
      importModel: 'Import Model',
      noModel: 'No models, please import first',
      chatModel: 'Chat Model',
      selectChatModel: 'Select chat model',
      startingModel: 'Starting...',
      startModel: 'Start Model',
      stopModel: 'Stop',
      aiChat: 'AI Chat',
      stopGeneration: 'Stop Generation',
      continueGenerate: 'Continue',
      newChat: 'New Chat',
      settings: 'Settings',
      clearChat: 'Clear Chat',
      selectDeviceAndLoadModel: 'Please select device and load model',
      helloAssistant: 'Hello! I am MYT Assistant',
      howCanIHelp: 'How can I help you?',
      thinkingProcess: 'Thinking Process',
      thinkingInProgress: 'Thinking...',
      feedbackQuestion: 'Was this answer helpful?',
      feedbackUseful: '👍 Useful, remember it',
      feedbackUseless: '👎 Not useful',
      serviceManagement: 'Service Management',
      stopServiceWarning: 'Stopping service will disable chat. You need to reload model.',
      currentDevice: 'Current Device',
      currentModel: 'Current Model',
      noModelLoaded: 'No model loaded',
      stoppingService: 'Stopping...',
      stopLLMService: 'Stop LLM Service',
      deviceManagement: 'Device Management',
      resetDeviceWarning: 'Resetting device will clear hardware config and state. Use with caution.',
      resetOperation: 'Reset Operation',
      resettingDevice: 'Resetting...',
      resetDevice: 'Reset Device',
      resetDeviceNote: '⚠️ Reconfiguration and model reload needed after reset',
      noDeviceSelected: 'No device selected',
      closeInSeconds: '{seconds}s to close',
      uploadMediaTooltip: 'Upload image or video',
      running: 'Running',
      deleteModel: 'Delete Model',
      resetConfirmMessage: 'Reconfiguration and model reload needed after reset',
      copyBtn: 'Copy',
      systemPrompt: 'System Prompt',
      systemPromptDesc: 'System prompt is automatically added at the beginning of each conversation to set AI role and behavior',
      enablePrompt: 'Enable Prompt',
      promptContent: 'Prompt Content',
      saveSettings: 'Save Settings',
      restoreDefault: 'Restore Default',
      // RPA Agent
      selectContainer: 'Select Container',
      loadingContainers: 'Loading containers...',
      bridged: 'Bridged',
      nonBridged: 'Non-bridged',
      noRunningContainer: 'No running containers',
      noSupportedDevice: 'No supported online devices (requires r1q or eces-rk3588-rk1828)',
      noAndroidContainer: 'No running Android containers on supported devices. Create and start cloud machines first, then refresh.',
      executionProgress: 'Execution Progress',
      selectTaskType: 'Select Task Type',
      executeTask: 'Execute Task',
      replyBtn: 'Reply',
      agentWaiting: 'Agent is waiting for your reply',
      selectedContainers: '{count} containers selected',
      taskLabel: 'Task:',
      rpaSettings: 'RPA Agent Settings',
      deviceModelConfig: 'Device Model Config (r1q / eces-rk3588-rk1828)',
      noSupportedOnlineDevice: 'No supported online devices',
      online: 'Online',
      offline: 'Offline',
      llmModel: 'LLM Model',
      selectChatModelRPA: 'Select chat model',
      refreshModelList: 'Refresh Model List',
      checkAndStartModel: 'Check & Start Model',
      agentGlobalParams: 'Agent Global Parameters',
      breakerRoundLimit: 'Breaker Round Limit',
      breakerHint: '0 = unlimited (exits on task_done or stop)',
      stepDelay: 'Step Delay (ms)',
      saveConfig: 'Save Config',
      // RPA Task Cards
      taskBrowseVideo: 'Browse Videos',
      taskSendMessage: 'Send Message',
      taskInstallApp: 'Install App',
      taskLikeContent: 'Like Content',
      taskSearch: 'Search',
      taskCustom: 'Custom',
      // RPA Status Labels
      statusPending: 'Pending',
      statusRunning: 'Running',
      statusSuccess: 'Success',
      statusFailed: 'Failed',
      statusStopped: 'Stopped',
      // RPA Messages
      agentWaitingReply: 'Agent is waiting for your reply...',
      taskExecuting: 'Task executing...',
      taskInputHint: 'Enter task description, e.g.: Open Douyin and scroll 10 videos',
      modelStatusRunning: 'Running',
      modelStatusStarting: 'Starting',
      modelStatusStopped: 'Not Started',
      modelStatusUnknown: 'Unknown',
      refreshContainerList: 'Refresh Container List',
      clickRefreshToLoad: 'Click Refresh to Load',
      noModelOnDevice: 'No models on this device',
      // HostManagement image tab
      onlineImages: 'Online Images',
      localImages: 'Local Images',
      noMatchingImages: 'No matching online images',
      filterImageName: 'Filter image name',
      filterImageURL: 'Filter image URL',
      compatibleImages: 'Compatible Images',
      resetFilter: 'Reset Filter',
      refreshOnlineImages: 'Refresh Online Images',
      noDescription: 'No description',
      downloadImage: 'Download',
      selectDeviceFirst: 'Please select a device first',
      selectGroupPlaceholder: 'Select group',
      sharedIP: 'Shared IP',
      independentIP: 'Independent IP',
      deleteGroupConfirm: 'Delete Group Confirm',
      deleteGroupMessage: 'Are you sure to delete group "{group}"? All devices will be moved to "Default Group"',
      editGroup: 'Edit Group',
      enterNewGroupName: 'Enter new group name',
      groupNameCannotBeEmpty: 'Group name cannot be empty',
      importantNotice: 'Important Notice (click to expand/collapse)',
      supportScope: 'Scope',
      supportScopeDesc: 'Only supports simulator image cloud machine backups, ',
      containerNotSupported: 'Container images not supported',
      backupImportFunction: ' for backup import',
      importPreparation: 'Before Import',
      importPreparationDesc: 'Please ensure the target device has downloaded the corresponding Android image',
      deviceCompatibility: 'Compatibility',
      deviceCompatibilityDesc: 'Backup files must fully match target device type (',
      cqrReusable: 'CQR series images are reusable',
      pSeriesNotReusable: 'P series cannot be reused',
      backupFileList: 'Backup File List',
      openImportFolder: 'Open Import Folder',
      noBackupFiles: 'No backup files',
      copyToImportFolder: 'Copy cloud machine backup files (.tar.gz format) to the import folder',
      thenRefresh: 'Then click the "Refresh" button at top right',
      fileName: 'File Name',
      fileSize: 'File Size',
      import: 'Import',
      batchImport: 'Batch Import',
      selectDevice: 'Select Device',
      configureSlots: 'Configure Slots',
      executeImport: 'Execute Import',
      searchDeviceIP: 'Search device IP',
      hostFirmwareVersion: 'Firmware Version',
      nvmeStorage: 'NVME Storage',
      noAvailableDevices: 'No available devices',
      nextStep: 'Next',
      previousStep: 'Previous',
      startImport: 'Start Import',
      selectAll: 'Select All',
      clear: 'Clear',
      batchSetCopyCount: 'Batch Set Copy Count',
      slotLabel: 'Slot',
      copies: 'copies',
      currentBackupFile: 'Current backup file:',
      overallProgress: 'Overall Progress',
      success: 'Success',
      failed: 'Failed',
      importingNow: 'Importing...',
      device: 'Device',
      successfulImports: 'Successfully Imported',
      importFailed: 'Import Failed',
      complete: 'Complete'
    },



    backup: {
      add: 'Add',
      addBackupMachine: 'Add Machine Backup',
      addBackupModel: 'Add Model Backup',
      all: 'All',
      availableSlots: 'Available Slots',
      backupMachine: 'Backup Machine',
      backupMachineList: 'Machine Backup List',
      backupModel: 'Backup Model',
      backupModelList: 'Model Backup List',
      batchImport: 'Batch Import',
      batchImportBackupModel: 'Batch Import Backup Model',
      cancel: 'Cancel',
      confirm: 'Confirm',
      delete: 'Delete',
      deviceIP: 'Device IP',
      deviceList: 'Device List',
      deviceStatus: 'Status',
      download: 'Download',
      downloading: 'Downloading...',
      ensureImageDownloaded: 'Ensure image is downloaded',
      enterMachineName: 'Enter Machine Name',
      enterModelName: 'Enter Model Name',
      export: 'Export',
      group: 'Group',
      import: 'Import',
      importBackupMachine: 'Import Machine Backup',
      importBtn: 'Import',
      index: 'Index',
      machineName: 'Machine Name',
      modelName: 'Model Name',
      noAvailableSlots: 'No Available Slots',
      openLocalBackupMachine: 'Open Local Machine Backup Dir',
      openLocalBackupModel: 'Open Local Model Backup Dir',
      operation: 'Actions',
      pleaseSelectDevice: 'Please select a device',
      pleaseSelectMachine: 'Please select a machine',
      searchMachineName: 'Search machine...',
      searchModelName: 'Search model...',
      selectDevice: 'Select Device',
      selectMachine: 'Select Machine',
      size: 'Size',
      slotNumber: 'Slot No.',
      tip: 'Tip',
      usageGuide: 'Usage Guide'
    },

    batchTask: {
      selectDeviceTitle: 'Select Device',
      selected: 'Selected',
      unit: '',
      searchPlaceholder: 'Search Device IP or Container',
      selectAll: 'Select All',
      clear: 'Clear',
      openAll: 'Open All',
      closeAll: 'Close All',
      total: 'Total',
      unitMachines: 'Machines',
      projection: 'Projection',
      noDevices: 'No devices available',
      batchOperationTitle: 'Batch Operation',
      batchExecuteCmd: 'Batch Execute Command',
      inputAdbCmd: 'Enter ADB Command:',
      viewCmdExample: 'View Examples',
      cmdExamplePlaceholder: 'Example:\ninput text \'Hello World\'\ninput tap 500 800\npm list packages',
      quickCmd: 'Quick Commands',
      clickToFill: 'Click to fill',
      basicOps: 'Basic Operations',
      appOps: 'App Management',
      sysOps: 'System Info',
      executing: 'Executing (Round ',
      loopSuffix: ')',
      executeNow: 'Execute',
      stop: 'Stop',
      executionResultTitle: 'Execution Results',
      loopPrefix: 'Round',
      errorStr: 'Error',
      cmdReference: 'ADB Command Reference',
      commonCmdExamples: 'Common ADB Command Examples:',
      cmdInputText: 'Input Text',
      cmdTap: 'Tap Screen',
      cmdSwipe: 'Swipe Screen',
      cmdHome: 'Home Key',
      cmdBack: 'Back Key',
      cmdPower: 'Power Key',
      cmdVolUp: 'Volume +',
      cmdVolDown: 'Volume -',
      cmdScreencap: 'Screenshot',
      cmdRecord: 'Start Recording',
      cmdStartSettings: 'Open Settings',
      cmdDial: 'Open Dialer',
      cmdListPkgs: 'List Packages',
      cmdListSys: 'List System Apps',
      cmdListThird: 'List Third-party Apps',
      cmdFindApp: 'Find App',
      cmdClearApp: 'Clear App Data',
      cmdUninstall: 'Uninstall App',
      cmdForceStop: 'Force Stop App',
      cmdStartApp: 'Start App',
      cmdSysVer: 'System Version',
      cmdModel: 'Device Model',
      cmdBrand: 'Device Brand',
      cmdAndroidId: 'Android ID',
      cmdIp: 'IP Address',
      cmdCpu: 'CPU Info',
      cmdMem: 'Memory Info',
      cmdStorage: 'Storage Space',
      cmdBattery: 'Battery Status',
      cmdResolution: 'Screen Resolution',
      cmdDensity: 'Screen Density',
      cmdCurrentAct: 'Current Activity',
      supportedVars: 'Supported Variables:',
      varDeviceIp: 'Device IP',
      varContainerId: 'Full Container ID',
      varContainerShortId: 'Short Container ID (12 chars)',
      varContainerName: 'Container Name',
      varTimestamp: 'Current Timestamp',
      precautions: 'precautions:',
      note1: 'Commands execute sequentially on selected devices',
      note2: 'Wrap text containing spaces in quotes',
      note3: 'Some commands require root access',
      cmdInstall: 'Install App',
      cmdGetVersion: 'Get Sys Version'
    },

    stream: {
      streamManagementTitle: 'Stream Management',
      forwardStream: 'Forward Stream (OBS)',
      p2pStream: 'Peer-to-Peer (OBS)',
      inputRoomId: 'Enter Room ID',
      refreshList: 'Refresh List',
      pushCodeRoom: 'Stream Key/Room',
      sourceIp: 'Source IP',
      bitrate: 'Bitrate (Video/Audio)',
      noStreamInfo: 'No stream info, click New Stream to create',
      distribute: 'Distribute',
      disconnect: 'Disconnect',
      deviceCloudRelation: 'Device & Cloud Machines',
      distributeToCloud: 'Distribute to Machine',
      activeDistribution: 'Active Distribution',
      refreshStatus: 'Refresh Status',
      pushAddress: 'Push URL',
      noActiveDistribution: 'No Active Distribution',
      inactiveDistribution: 'Inactive/Unverified Distribution',
      deviceOfflineOrNotExist: 'Device Offline or Not Found',
      startP2P: 'Start P2P',
      webrtcAddress: 'WEBRTC URL',
      rtmpAddress: 'RTMP URL',

      queryingStatus: 'Querying status...',
      activeStatus: 'Active',
      idleStatus: 'Idle',
      streamGuideHtml: `<div class="stream-guide" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-connection" style="margin-right: 8px;"></i> Forward Stream (OBS)
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>Open OBS, go to <b>Settings -> Stream</b>, and set Service to <b>Custom</b>.</li>
    <li>Copy the WebRTC or RTMP URL and paste it into "Server" (click URL to copy).</li>
    <li>Enter a custom stream key. Click OK, then click <b>Start Streaming</b> in OBS.</li>
    <li>Return here, click <b>New Stream</b>, and enter the <b>exact same name</b> as your stream key.
      <div style="color: #e6a23c; font-size: 12px; margin-top: 4px;"><i class="el-icon-info"></i> With the same name, this entry remains in the list displaying real-time stream status.</div>
    </li>
    <li>Once the status is <b>Active</b>, click <b>Distribute</b> and select the target device and machine.</li>
    <li>Open the <b>Camera</b> app inside the cloud machine to view the stream.</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-link" style="margin-right: 8px;"></i> Peer-to-Peer (P2P / OBS)
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>Switch to <b>P2P (OBS)</b> tab and click <b>Add P2P Stream</b>.</li>
    <li>Select <b>Device</b> and <b>Machine</b>. The list will display a <b>Listen Address</b> (click to copy).</li>
    <li>Open OBS, go to <b>Settings -> Stream</b>, set Service to <b>Custom</b>, and paste the listen address into "Server". Leave Stream Key empty.</li>
    <li>Click <b>Start Streaming</b> in OBS, then click the <b>Start P2P</b> button in the list here.</li>
    <li>Open the <b>Camera</b> app inside the cloud machine to view the feed.</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-video-camera" style="margin-right: 8px;"></i> PC Camera
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>Switch to <b>PC Camera</b> tab and click <b>Add Camera Stream</b>.</li>
    <li>Select <b>Device, Machine</b>, and <b>Camera</b>. Config resolution & bitrate, save.</li>
    <li>Click <b>Start</b>. The right preview area will automatically show live feed.</li>
    <li>Open the <b>Camera</b> app inside the cloud machine to view the live feed.</li>
  </ol>
</div>`
    },

    instance: {
      alipayTip: 'Please scan the QR code via Alipay to pay',
      authSync: 'Auth Sync',
      confirmBuy: 'Confirm Purchase',
      details: 'Details',
      detailTitle: 'Instance Details',
      host: 'Host',
      instanceIP: 'Instance / IP',
      instances: 'Instances',
      loadingPackages: 'Loading packages...',
      loadingQR: 'Generating Payment QR...',
      login: 'Login',
      loginRequired: 'Login required for authorization',
      noInstance: 'No Instances found',
      noPackages: 'No packages',
      purchaseRenew: 'Purchase/Renew Auth',
      query: 'Query',
      register: 'Register',
      scanToPay: 'Scan to Pay',
      searchPlaceholder: 'Search Host IP or Machine Name',
      selectPackage: 'Select Package',
      statusExpired: 'Expired',
      statusExpiring: 'Expiring Soon',
      statusNormal: 'Normal',
      subtotal: 'Subtotal',
      totalCount: 'Total ',
      instancePrefix: 'Instance',
      username: 'Username',
      validUntil: 'Valid Until',
      password: 'Password',
      rememberCredentials: 'Remember Credentials',
      forgotPassword: 'Forgot Password',
      cancel: 'Cancel',
      loggingIn: 'Logging in...',
      userRegistration: 'User Registration',
      phone: 'Phone',
      loginPassword: 'Login Password',
      confirmPassword: 'Confirm Password',
      verificationCode: 'Verification Code',
      enterUsernameOrPhone: 'Enter username / phone',
      enterPassword: 'Enter password',
      enterPhone: 'Enter phone number',
      enterPasswordAgain: 'Enter password again',
      enterVerificationCode: 'Enter verification code',
      sendVerificationCode: 'Send Code',
      registering: 'Registering...',
      resetPassword: 'Reset Password',
      resettingPassword: 'Resetting...',
      phoneNumber: 'Phone Number',
      newPassword: 'New Password',
      confirmNewPassword: 'Confirm New Password',
      phoneVerificationCode: 'Phone Verification Code',
      noAccount: 'No account yet?',
      registerNow: 'Register Now',
      instanceSlot: 'Instance Slot',
      selectedInstances: 'Selected Instances',
      unit: ''
    },
    network: {
      nodeManagement: 'Node Management',
      nodeAllocation: 'Node Allocation',
      domainFilter: 'Domain Filter',
      domainDirect: 'Domain Direct',
      privateNic: 'Private NIC',
      publicNic: 'Public NIC',
      usageGuide: 'Usage Guide',
      macvlanHint: 'Due to the promiscuous mode mechanism, MacVlan does not support wireless network cards. A wired network connection is strictly required.',
      nodeManagementList: 'Node Management List',
      searchGroup: 'Search Group',
      selectedNodes: 'Selected Nodes',
      batchDeleteNodes: 'Batch Delete Nodes',
      batchSpeedTest: 'Batch Speed Test',
      type: 'Type',
      alias: 'Alias',
      address: 'Address',
      port: 'Port',
      protocol: 'Protocol',
      latency: 'Latency',
      deviceList: 'Device List',
      assignVpc: 'Assign Node',
      addGroup: 'Add Group',
      addDomain: 'Add Domain',
      addRule: 'Add Rule',
      batchClearVpc: 'Batch Clear',
      clearDirect: 'Clear Direct',
      clearFilter: 'Clear Filter',
      clearVpc: 'Clear Node',
      closeDns: 'Close DNS',
      closeDnsWhitelist: 'Whitelist Mode',
      createNic: 'Create NIC',
      deleteGroup: 'Delete Group',
      deleteGroupNode: 'Delete',
      directDomain: 'Direct Domain',
      domainRuleDesc: 'Domain Match Profile',
      domainRules: 'Domain Rules',
      editGroupAlias: 'Edit Alias',
      filterDomain: 'Filter Domain',
      fullRuleDesc: 'Full Match Profile',
      gateway: 'Gateway',
      globalRules: 'Global Rules',
      index: 'Index',
      keywordRuleDesc: 'Keyword Match',
      machineIP: 'Machine IP',
      machineName: 'Machine Name',
      nicName: 'NIC Name',
      nodeName: 'Node Name',
      noDomainDirectRules: 'No Direct Rules',
      noDomainRules: 'No Domain Rules',
      noPublicNicData: 'No Public NIC Data',
      noVpcContainers: 'No Containers',
      openDns: 'Open DNS',
      privateNicDesc: 'Private NIC (VPC Network) Desc',
      privateNicDescDetail: 'Creates an isolated virtual gateway on the host allocating internal IPs for machines to isolate traffic or bridge to proxy nodes.',
      privateNicList: 'Private NIC List',
      publicNicDesc: 'Public NIC (MacVlan) Desc',
      publicNicDescDetail: 'Binds the cloud machine directly to the physical LAN, allowing it to obtain an IP address on your real local network router.',
      queryContainerDomainDirect: 'Query Container Direct',
      queryContainerDomainFilter: 'Query Container Filter',
      queryGlobalDomainFilter: 'Query Global Filter',
      selectContainer: 'Select Container',
      selectContainerOrQuery: 'Select Container to Set',
      selectContainerToQuery: 'Select Container to Query',
      selectDeviceFirst: 'Please select a device first',
      selectVpcContainer: 'Select VPC Container',
      setContainerDomainFilter: 'Set Container Domain Filter',
      setDomainDirect: 'Set Domain Direct',
      setGlobalDomainFilter: 'Set Global Domain Filter',
      sourceGroup: 'Source Group',
      speedTest: 'Speed Test',
      updateGroup: 'Update Group',
      updateMacVlan: 'Update MacVlan',
      updateMacVlanTitle: 'Update MacVLAN Config',
      vpcNodes: 'VPC Nodes'
    },
    model: {
      simulatorOnlyHint: 'These configurations only take effect when using V3 Simulator images. Container images may ignore them due to architectural differences.',
      onlineModels: 'Online Models',
      localModels: 'Local Models',
      usageGuide: 'Usage Guide',
      openDownloadDir: 'Open Download Dir',
      openLocalDir: 'Open Local Dir',
      selectedCount: 'Selected {count} models',
      noModelSelected: 'No models selected',
      pushSelected: 'Push to Device ({count})',
      searchPlaceholder: 'Search Models...',
      modelId: 'Model ID',
      modelName: 'Model Name',
      push: 'Edit',
      pushed: 'Pushed',
      noModelData: 'No Model Data Found',
      totalModels: 'Total: ',
      helpContent: `<div class="model-guide-container" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    📄 Overview
  </div>
  <p style="margin: 0; padding-left: 8px;">This page manages phone templates for the cloud machines. It supports viewing online templates, downloading, configuration editing, and pushing local templates to devices.</p>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🌐 Online Models
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Displays available model templates on the server, showing model ID and name.</li>
    <li><b>Download:</b> A "Download" button appears for templates not available locally. Click to fetch the template.</li>
    <li><b>Edit:</b> If already downloaded, an "Edit" button allows modifying overlay and prop variables. Saving creates a new variant.</li>
    <li>The top right button quickly opens your local templates directory.</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    💾 Local Models
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Shows stored local templates ready for pushing to online devices.</li>
    <li><b>Push (Individual):</b> Click "Push" on a row, select target devices in the popup, and hit "Start Pushing".</li>
    <li><b>Batch Push:</b> Select multiple templates, then use the top "Push Selected" button to deploy uniformly.</li>
    <li><b>Refresh:</b> Rescan your local directory and update the list.</li>
    <li><b>Capture Tool:</b> Scan the QR code to install an APK on a real phone and capture physical hardware identifiers for cloning. (Only supported on matching Android versions).</li>
    <li>The top right button quickly opens your local templates directory.</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📝 Editor Parameters
  </div>
  <p style="margin: 0; padding-left: 8px;">Clicking "Edit" on a template allows tweaking device overlays:</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>model (Template Name):</b> Read-only base model tag.</li>
  </ul>
  
  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Overlay Variables (Hardware ID Attributes)</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>EMMCID / EMMCSD:</b> eMMC storage unique chips/SN. <code>Spaces</code> = random on boot.</li>
    <li><b>GRID:</b> GPU graphic identifier. <code>Spaces</code> = random on boot.</li>
    <li><b>MEDIAID / RNDID:</b> Random media and device IDs.</li>
    <li><b>SERIALNUMBER / Serial Number:</b> Full and short Device SN. <code>Spaces</code> = random.</li>
    <li><b>Aaid / Vaid:</b> Anonymous & Vendor Ads IDs.</li>
    <li><b>Android Id:</b> Primary unresettable Android identifier. <code>Spaces</code> = random.</li>
    <li><b>Bluetooth / Wifi Name:</b> Wireless display names.</li>
    <li><b>GnssModel / OviLbs:</b> GPS module routing flags. Usually defaults are sufficient.</li>
    <li><b>Wifi List:</b> Numeric cap of simulated Wi-Fi scan results.</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Prop Variables (Read-Only OS Properties)</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>Ro.build.version.security_patch:</b> OS patch date (e.g., <code>2023-12-05</code>).</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Regarding <code>Blank Spaces</code></p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Fields consisting entirely of spaces will instruct the container to inject randomized hex values upon startup automatically.</li>
    <li>If rewriting them, enforce format boundaries (e.g., hex identifiers must be <code>0-9, a-f</code>).</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Saving Rules</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Clicking "Save" generates a new localized variant appending random digits (e.g., <code>PJA110_182</code>). Baseline templates are never overwritten.</li>
    <li><span style="color: #F56C6C;">⚠️ Inputting invalid config variables can break OS boots. Leave unknown fields completely blank to use resilient randomizers safely.</span></li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📤 Pushing Devices
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>The popup relays online statuses. Offline endpoints will be filtered.</li>
    <li>Support pushing batches of templates across massive arrays of cloud instances.</li>
    <li>Always click "Refresh List" before multi-targeting to weed out disconnected hosts.</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🔍 Search & Pagination
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Top input bounds immediate text-filtered lookups on model names.</li>
  </ul>
</div>`,
      configEdit: 'Edit Config: {name}',
      configWarning: 'Editing configurations may alter the cloud machine operational status.',
      templateName: 'Template Name',
      collectionToolDownload: 'Collection Tool Download',
      scanQrToDownload: 'Scan QR to Download',
      collectionToolNote: 'Install on a phone to start collecting!',
      collectionToolWarning1: 'Do not distribute unauthorized copies!',
      collectionToolWarning2: 'We are not responsible for account bans related to misuse.',
      selectPushDevice: 'Select Target Device',
      modelsToPush: 'Models to Push:',
      selectTargetDevices: 'Select Target Devices',
      onlineDeviceCount: '{count} Online',
      refreshList: 'Refresh List',
      deviceIP: 'Device IP',
      pushing: 'Pushing {current}/{total}',
      pushComplete: 'Push Complete ({total})',
      current: 'Current: ',
      startPush: 'Start Push'
    },
    image: {
      usageGuide: 'Usage Guide',
      guideContent: `<div class="image-guide-container" style="font-size: 13px; color: #606266; line-height: 2;">
  <h3 style="color: #303133; font-size: 16px; margin-bottom: 20px; display: flex; align-items: center;">
    <i class="el-icon-picture-outline" style="margin-right: 8px;"></i> Image Management Guide
  </h3>

  <!-- 1. Introduction -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    📄 Feature Introduction
  </div>
  <p style="margin-top: 0; margin-bottom: 12px;">The Image Management module manages Android system images running on your devices. It supports downloading images from the cloud, pushing them to devices, and managing existing images. It consists of three main tabs:</p>
  <ul style="padding-left: 20px; margin-bottom: 24px;">
    <li><b>Online Images:</b> Lists available cloud images. You can filter by device model, type, or Android version, and download or push them.</li>
    <li><b>Local Images:</b> Views templates downloaded locally. Push them directly to connected devices or delete them to free up disk space.</li>
    <li><b>Device Images:</b> Verify images natively installed on specific devices and safely delete them if unneeded.</li>
  </ul>

  <!-- 2. Workflow -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    🔄 Typical Workflow
  </div>
  <div style="margin-bottom: 30px; position: relative;">
    <ol style="padding-left: 10px; margin: 0; list-style: none;">
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">1</span> <div><b style="color: #409eff;">Select Online Image</b><br><span style="color: #409eff;">Find an image compatible with your device in the Online list.</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">2</span> <div><b style="color: #409eff;">Download Locally</b><br><span style="color: #409eff;">Click Download to cache the image locally for faster reuse.</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">3</span> <div><b style="color: #409eff;">Push to Device</b><br><span style="color: #409eff;">Trigger a device push to deploy the firmware physically.</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">4</span> <div><b style="color: #409eff;">Verify Installation</b><br><span style="color: #409eff;">Check the Device Images tab to confirm success block availability.</span></div></li>
      <li style="margin-bottom: 12px; display: flex; align-items: flex-start;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; height: 20px; flex-shrink: 0; border-radius: 50%; line-height: 20px;">5</span> <div><b style="color: #409eff;">Boot Instance</b><br><span style="color: #409eff;">Deploy a new active cloud machine leveraging this image.</span></div></li>
    </ol>
  </div>

  <!-- 3. Simulator vs Container -->
  <div style="font-weight: bold; margin-bottom: 12px; font-size: 14px;">Simulator vs Container: Differences</div>
  <table style="width: 100%; border-collapse: collapse; margin-bottom: 24px; text-align: left; font-size: 12px;">
    <thead>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <th style="padding: 10px 0; width: 15%; color:#606266;">Comparison</th>
        <th style="padding: 10px 0; width: 42%; color:#409eff;">🖥️ Simulator</th>
        <th style="padding: 10px 0; width: 43%; color:#e6a23c;">📦 Container</th>
      </tr>
    </thead>
    <tbody>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Architecture</td>
        <td style="padding: 8px 0;">VM imitating ARM hardware (QEMU)</td>
        <td style="padding: 8px 0;">Lightweight Linux orchestration</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Boot Speed</td>
        <td style="padding: 8px 0;">Slower, typically 50+ secs</td>
        <td style="padding: 8px 0;">Extremely fast, usually 3~10 secs</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Resource Use</td>
        <td style="padding: 8px 0;">Higher, requires isolated CPU blocks</td>
        <td style="padding: 8px 0;">Extremely Low, multiple instances share host pools</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Density</td>
        <td style="padding: 8px 0;">Strict limits; few deployments</td>
        <td style="padding: 8px 0;">Unlimited soft barriers; tens to hundreds</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Compatibility</td>
        <td style="padding: 8px 0;">Behaves close to true hardware devices</td>
        <td style="padding: 8px 0;">Dependence on host OS bounds hooks</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Graphics</td>
        <td style="padding: 8px 0;">Fluid GPU rendering layers</td>
        <td style="padding: 8px 0;">Softer graphics bounds</td>
      </tr>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Use Cases</td>
        <td style="padding: 8px 0;">UI tests, Gaming, Heavy graphics</td>
        <td style="padding: 8px 0;">Automation scripts, High density deployments</td>
      </tr>
    </tbody>
  </table>

  <!-- 4. Suggestions -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #e6a23c;">
    💡 Selection Advice
  </div>
  <div style="background-color: #f0f9eb; padding: 12px; border-radius: 4px; border-left: 4px solid #67c23a; margin-bottom: 8px;">
    <div style="display: flex; align-items: flex-start;">
      <i class="el-icon-success" style="color: #67c23a; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
      <div>
        <div style="font-weight: bold; color: #67c23a; margin-bottom: 4px;">Recommended Container usage:</div>
        <div style="color: #67c23a;">When requiring enormous scalability across parallel scripts and lightweight headless setups. Rapid spin ups.</div>
      </div>
    </div>
  </div>
  <div style="background-color: #fdf6ec; padding: 12px; border-radius: 4px; border-left: 4px solid #e6a23c; margin-bottom: 24px;">
    <div style="display: flex; align-items: flex-start;">
       <i class="el-icon-warning" style="color: #e6a23c; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
       <div>
         <div style="font-weight: bold; color: #e6a23c; margin-bottom: 4px;">Recommended Simulator usage:</div>
         <div style="color: #e6a23c;">Deep graphical dependencies, high accuracy hardware modeling testcases, Android core behavior bounds.</div>
       </div>
    </div>
  </div>

  <!-- 5. Warnings -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #F56C6C;">
    ⚠️ Important Notes
  </div>
  <ul style="padding-left: 20px; color: #606266; margin-bottom: 0;">
    <li>Firmware is generally 1GB+; ensure host disk capacity permits holding payloads.</li>
    <li>Networks must remain connected uninterrupted while pumping huge files to device endpoints.</li>
    <li>Purging an image active machines rely upon causes unrecoverable crash states. Audit environments cautiously.</li>
    <li><b>P1 series devices</b> do not support Android 12 container variants. Choose A10 or A14.</li>
  </ul>
</div>`
    },
    addDevice: {
      title: 'Add Device',
      scanDiscovery: 'Scan',
      manualAdd: 'Manual',
      scanDevices: 'Scan Devices',
      scanning: 'Scanning...',
      autoDiscoverHint: 'Auto discover devices in LAN',
      searchPlaceholder: 'Search IP, name or ID',
      selectAllToggle: 'Select All / Deselect',
      devicesCount: 'Total: {total}, Filtered: {filtered}, Selected: {selected}',
      deviceIP: 'Device IP',
      deviceName: 'Device Name',
      deviceID: 'Device ID',
      version: 'Version',
      status: 'Status',
      added: 'Added',
      canAdd: 'Available',
      addSelectedDevices: 'Add Selected',
      noDevicesFound: 'No devices found, please scan',
      enterDeviceIP: 'Enter Device IP',
      multipleIPHint: 'Support comma or line break separation, e.g.: 192.168.1.100, 192.168.1.101',
      enterIPPlaceholder: 'Enter device IP address, separate multiple addresses with comma or line break',
      addDevice: 'Add Device',
      querying: 'Querying...',
      clear: 'Clear',
      successAdded: 'Successfully added {count} devices',
      noResponseDevices: 'No response from these devices ({count}):',
      noDeviceDiscovered: 'No device discovered',
      noResponseLabel: 'No response devices:',
      close: 'Close',
      // Batch Task related translations
      batchTask: 'Batch Task',
      executeNow: 'Execute Now',
      scheduledExecution: 'Scheduled Execution',
      commandTemplate: 'Command Template',
      executionHistory: 'Execution History',
      addTemplate: 'Add Template',
      variableReplacement: 'Variable Replacement',
      executing: 'Executing',
      executionSuccess: 'Success',
      executionFailed: 'Failed',
      onceSchedule: 'Once',
      periodicSchedule: 'Periodic',
      cronExpression: 'Cron Expression',
      targetDevices: 'Target Devices',
      executionResult: 'Result',
      exportHistory: 'Export History',
      selectTargets: 'Select Targets',
      commandEditor: 'Command Editor',
      scheduleConfig: 'Schedule Config',
      templateManager: 'Template Manager',
      historyViewer: 'History Viewer',
      enterCommand: 'Enter command',
      taskName: 'Task Name',
      createTask: 'Create Task',
      updateTask: 'Update Task',
      deleteTask: 'Delete Task',
      enableTask: 'Enable Task',
      disableTask: 'Disable Task',
      totalDevices: 'Total: {total} devices',
      successCount: 'Success: {count}',
      failedCount: 'Failed: {count}',
      duration: 'Duration',
      output: 'Output',
      exportJSON: 'Export JSON',
      exportCSV: 'Export CSV'
    },
    menu: {
      hostManagement: 'Host',
      cloudManagement: 'Phone',
      imageManagement: 'Image',
      modelManagement: 'Model',
      networkManagement: 'Network',
      backupManagement: 'Backup',
      instanceManagement: 'Instance',
      streamManagement: 'Stream',
      batchTask: 'Batch Task',
      aiAssistant: 'AI Agent',
      rpaAgent: 'RPA Agent',
      logout: 'Logout',
      about: 'About',
      interconnectedCloudMachines: 'Interconnected'
    },
    update: {
      checkUpdate: 'Check for Updates',
      newVersionAvailable: 'New Version Available',
      currentVersion: 'Current Version',
      noUpdate: 'Already up to date',
      downloading: 'Checking...',
      about: 'About',
      autoCheckUpdate: 'Auto Check Updates',
      autoDownloadUpdate: 'Auto Download Updates',
      hasNewVersion: 'New Version',
      updateSettings: 'Update Settings'
    },
    dialog: {
      addDeviceTitle: 'Add Device'
    },
    app: {
      title: 'MoYunTeng V3 Client',
      description: 'ARM Edge Computing Device Management Client'
    },
    cloudMachine: {
      currentName: 'Current Name',
      newName: 'New Name',
      enterNewName: 'Please enter new name',
      running: 'Running',
      backupName: 'Backup Name',
      remark: 'Remark',
      created: 'Created',
      restarting: 'Restarting',
      renameCloudMachine: 'Rename Cloud Machine',
      switchBackup: 'Switch Backup',
      batchAddToGroup: 'Batch Add to Group',
      slotMode: 'Slot',
      batchMode: 'Batch',
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
      resetContainer: 'Reset Container'
    }
  }
}

// 响应式 i18n 状态
const i18nState = reactive({
  locale: 'zh-CN'
})

// 自定义的国际化系统（响应式）
const i18n = {
  get locale() {
    return i18nState.locale
  },
  
  set locale(val) {
    i18nState.locale = val
  },
  
  messages,
  
  t(key, params) {
    // 访问响应式的 locale，确保依赖追踪
    const currentLocale = i18nState.locale
    const keys = key.split('.')
    let result = messages[currentLocale]
    
    for (const k of keys) {
      if (result && result[k]) {
        result = result[k]
      } else {
        return key
      }
    }
    
    let text = result || key
    
    // 支持参数插值，如 {count} -> params.count
    if (params && typeof text === 'string') {
      Object.keys(params).forEach(k => {
        text = text.replace(new RegExp(`\\{${k}\\}`, 'g'), params[k])
      })
    }
    
    return text
  },
  
  setLocale(locale) {
    i18nState.locale = locale
    try {
      localStorage.setItem('app-locale', locale)
      if (typeof window !== 'undefined') {
        document.documentElement.setAttribute('lang', locale)
      }
    } catch (e) {
      console.warn('Cannot save locale:', e)
    }
  },
  
  getLocale() {
    return i18nState.locale
  }
}

// 初始化语言
try {
  const savedLocale = localStorage.getItem('app-locale')
  if (savedLocale) {
    i18nState.locale = savedLocale
  }
} catch (e) {
  console.warn('Cannot load saved locale:', e)
}

// 创建应用
const app = createApp(App)

// 全局提供翻译功能（使用绑定的方法确保响应式）
app.config.globalProperties.$t = (key, params) => i18n.t(key, params)
app.config.globalProperties.$i18n = i18n

app.use(ElementPlus)
app.config.globalProperties.$axios = axios
app.mount('#app')
