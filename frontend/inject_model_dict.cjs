const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainJsContent = fs.readFileSync(mainJsPath, 'utf8');

const zhModelDict = `
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
      helpContent: \`
        <h3 class="guide-title">
          <span style="vertical-align: middle; margin-right: 6px;">ℹ️</span>机型管理使用说明
        </h3>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">📋 功能介绍</h4>
        <p class="guide-text">机型管理模块用于管理运行在设备上的云机环境配置（如品牌、型号等），支持浏览在线可用机型、查阅本地已下载机型，以及将配置推送到真实设备。</p>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">💡 注意事项</h4>
        <ul class="guide-list">
          <li><strong>仅支持V3模拟器镜像</strong>：容器镜像底层机制不同，部分或全机型配置对其可能不生效。</li>
          <li><strong>设备推送</strong>：推送前请确保目标设备在线且网络通畅。</li>
        </ul>
      \`,
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
    },`;

const enModelDict = `
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
      helpContent: \`
        <h3 class="guide-title">
          <span style="vertical-align: middle; margin-right: 6px;">ℹ️</span>Model Management Guide
        </h3>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">📋 Features Introduction</h4>
        <p class="guide-text">The Model module allows you to manage cloud machine hardware configurations, such as device branding and identifiers. You can browse online configs, view your local templates, and push them to devices.</p>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">💡 Important Notes</h4>
        <ul class="guide-list">
          <li><strong>Simulator Only</strong>: Due to differing architectures, these settings consistently work on V3 Simulators but may be limited on Container images.</li>
          <li><strong>Push Requirements</strong>: Please ensure target devices are online and connected prior to pushing.</li>
        </ul>
      \`,
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
    },`;

let insertCount = 0;
mainJsContent = mainJsContent.replace(
  /    image: \{/g,
  function(match) {
    insertCount++;
    if (insertCount === 1) {
      return zhModelDict + '\n' + match;
    } else if (insertCount === 2) {
      return enModelDict + '\n' + match;
    }
    return match;
  }
);

fs.writeFileSync(mainJsPath, mainJsContent);
console.log('SUCCESS: Injected model dicts');
