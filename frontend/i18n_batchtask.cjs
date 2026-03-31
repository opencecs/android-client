const fs = require('fs');
const path = require('path');

const vuePath = path.join(__dirname, 'src', 'components', 'BatchTaskManagement.vue');
let content = fs.readFileSync(vuePath, 'utf8');

// Insert useI18n if needed
if (!content.includes('const { t } = useI18n()') && !content.includes('getCurrentInstance()')) {
  content = content.replace(/import \{ ref, computed, watch \} from 'vue'/, "import { ref, computed, watch, getCurrentInstance } from 'vue'");
  content = content.replace(/const activeOperationType = ref\('command'\)/, "const { proxy } = getCurrentInstance()\nconst $t = proxy.$t\nconst activeOperationType = ref('command')");
} else if (!content.includes('const $t = proxy.$t') && content.includes('getCurrentInstance')) {
  content = content.replace(/const activeOperationType = ref\('command'\)/, "const { proxy } = getCurrentInstance()\nconst $t = proxy.$t\nconst activeOperationType = ref('command')");
}

// ---------------- Batch Task Vue Replace ----------------
content = content.replace(/<span style="font-weight: bold; font-size: 15px;">📱 选择设备<\/span>/g, `<span style="font-weight: bold; font-size: 15px;">📱 {{ $t('batchTask.selectDeviceTitle') }}</span>`);
content = content.replace(/已选 \{\{ selectedDevices.length \}\} 个/g, `{{ $t('batchTask.selected') }} {{ selectedDevices.length }} {{ $t('batchTask.unit') }}`);
content = content.replace(/placeholder="搜索设备IP或容器"/g, `:placeholder="$t('batchTask.searchPlaceholder')"`);
content = content.replace(/>全选<\/el-button>/g, `>{{ $t('batchTask.selectAll') }}</el-button>`);
content = content.replace(/>清空<\/el-button>/g, `>{{ $t('batchTask.clear') }}</el-button>`);
content = content.replace(/>全部打开<\/el-button>/g, `>{{ $t('batchTask.openAll') }}</el-button>`);
content = content.replace(/>全部关闭<\/el-button>/g, `>{{ $t('batchTask.closeAll') }}</el-button>`);
content = content.replace(/共 \{\{ filteredDevices.length \}\} 台/g, `{{ $t('batchTask.total') }} {{ filteredDevices.length }} {{ $t('batchTask.unitMachines') }}`);
content = content.replace(/>\s*投屏\s*<\/el-button>/g, `>\n                      {{ $t('batchTask.projection') }}\n                    </el-button>`);
content = content.replace(/description="暂无可用设备"/g, `:description="$t('batchTask.noDevices')"`);

content = content.replace(/<span style="font-weight: bold;">⚡ 批量操作<\/span>/g, `<span style="font-weight: bold;">⚡ {{ $t('batchTask.batchOperationTitle') }}</span>`);
content = content.replace(/label="📝 批量执行命令"/g, `:label="'📝 ' + $t('batchTask.batchExecuteCmd')"`);
content = content.replace(/>输入ADB命令：<\/span>/g, `>{{ $t('batchTask.inputAdbCmd') }}</span>`);
content = content.replace(/>\s*查看命令示例\s*<\/el-button>/g, `>\n                      {{ $t('batchTask.viewCmdExample') }}\n                    </el-button>`);
content = content.replace(/placeholder="示例命令：&#10;input text 'Hello World'&#10;input tap 500 800&#10;pm list packages"/g, `:placeholder="$t('batchTask.cmdExamplePlaceholder')"`);
content = content.replace(/>🔖 快捷命令<\/span>/g, `>🔖 {{ $t('batchTask.quickCmd') }}</span>`);
content = content.replace(/>点击按钮快速填充<\/span>/g, `>{{ $t('batchTask.clickToFill') }}</span>`);
content = content.replace(/>📱 基础操作<\/div>/g, `>📱 {{ $t('batchTask.basicOps') }}</div>`);
content = content.replace(/>📦 应用管理<\/div>/g, `>📦 {{ $t('batchTask.appOps') }}</div>`);
content = content.replace(/>🔧 系统信息<\/div>/g, `>🔧 {{ $t('batchTask.sysOps') }}</div>`);

content = content.replace(/\{\{ executing \? `执行中 \(第\$\{currentLoop\}\/\$\{loopCount\}轮\)\.\.\.` : '🚀 立即执行' \}\}/g, `{{ executing ? $t('batchTask.executing') + currentLoop + '/' + loopCount + $t('batchTask.loopSuffix') + '...' : '🚀 ' + $t('batchTask.executeNow') }}`);
content = content.replace(/>\s*⏹ 停止\s*<\/el-button>/g, `>\n                    ⏹ {{ $t('batchTask.stop') }}\n                  </el-button>`);
content = content.replace(/>📊 执行结果<\/span>/g, `>📊 {{ $t('batchTask.executionResultTitle') }}</span>`);
content = content.replace(/>第 \{\{ result.loop \}\} \/ \{\{ result.total \}\} 轮<\/span>/g, `>{{ $t('batchTask.loopPrefix') }} {{ result.loop }} / {{ result.total }} {{ $t('batchTask.loopSuffix') }}</span>`);
content = content.replace(/错误: \{\{ result.error \}\}/g, `{{ $t('batchTask.errorStr') }}: {{ result.error }}`);

content = content.replace(/title="📖 ADB命令参考"/g, `:title="'📖 ' + $t('batchTask.cmdReference')"`);
content = content.replace(/>📱 常用ADB命令示例：<\/h4>/g, `>📱 {{ $t('batchTask.commonCmdExamples') }}</h4>`);
content = content.replace(/>输入文本<\/li>/g, `>{{ $t('batchTask.cmdInputText') }}</li>`);
content = content.replace(/>点击坐标 \(x=500, y=800\)<\/li>/g, `>{{ $t('batchTask.cmdTap') }}</li>`);
content = content.replace(/>滑动屏幕<\/li>/g, `>{{ $t('batchTask.cmdSwipe') }}</li>`);
content = content.replace(/>按Home键 \(3=HOME, 4=BACK, 26=POWER\)<\/li>/g, `>{{ $t('batchTask.cmdHome') }}</li>`);
content = content.replace(/>列出所有应用包名<\/li>/g, `>{{ $t('batchTask.cmdListPkgs') }}</li>`);
content = content.replace(/>安装应用<\/li>/g, `>{{ $t('batchTask.cmdInstall') }}</li>`);
content = content.replace(/>卸载应用<\/li>/g, `>{{ $t('batchTask.cmdUninstall') }}</li>`);
content = content.replace(/>启动设置<\/li>/g, `>{{ $t('batchTask.cmdStartSettings') }}</li>`);
content = content.replace(/>查看电池信息<\/li>/g, `>{{ $t('batchTask.cmdBattery') }}</li>`);
content = content.replace(/>截屏<\/li>/g, `>{{ $t('batchTask.cmdScreencap') }}</li>`);
content = content.replace(/>获取系统版本<\/li>/g, `>{{ $t('batchTask.cmdGetVersion') }}</li>`);

content = content.replace(/>💡 支持的变量替换：<\/h4>/g, `>💡 {{ $t('batchTask.supportedVars') }}</h4>`);
content = content.replace(/>设备IP地址<\/li>/g, `>{{ $t('batchTask.varDeviceIp') }}</li>`);
content = content.replace(/>容器完整ID<\/li>/g, `>{{ $t('batchTask.varContainerId') }}</li>`);
content = content.replace(/>容器短ID \(前12位\)<\/li>/g, `>{{ $t('batchTask.varContainerShortId') }}</li>`);
content = content.replace(/>容器名称<\/li>/g, `>{{ $t('batchTask.varContainerName') }}</li>`);
content = content.replace(/>当前时间戳<\/li>/g, `>{{ $t('batchTask.varTimestamp') }}</li>`);

content = content.replace(/>⚠️ 注意事项：<\/h4>/g, `>⚠️ {{ $t('batchTask.precautions') }}</h4>`);
content = content.replace(/>命令会在每个选中的设备上依次执行<\/li>/g, `>{{ $t('batchTask.note1') }}</li>`);
content = content.replace(/>包含空格的文本参数需要用引号包裹<\/li>/g, `>{{ $t('batchTask.note2') }}</li>`);
content = content.replace(/>某些命令需要root权限才能执行<\/li>/g, `>{{ $t('batchTask.note3') }}</li>`);

// Let's replace the script arrays keys dynamically using getter properties instead of static arrays
if (!content.includes('get basicTemplates()')) {
  content = content.replace(/const basicTemplates = \[/, "const basicTemplates = computed(() => [");
  content = content.replace(/const appTemplates = \[/, "const appTemplates = computed(() => [");
  content = content.replace(/const systemTemplates = \[/, "const systemTemplates = computed(() => [");
  
  // They are closed with ]
  // We need to properly close the computed if they are computed
  content = content.replace(/\]\s*\n\s*\/\/ 应用管理命令/g, "])\n\n// 应用管理命令");
  content = content.replace(/\]\s*\n\s*\/\/ 系统信息命令/g, "])\n\n// 系统信息命令");
  content = content.replace(/\]\s*\n\s*\/\/ 保留旧的/g, "])\n\n// 保留旧的");

  // In template, `basicTemplates` is iterated. Vue unwraps refs in templates automatically, so no template changes needed!
}

// In basicTemplates array, replace names:
content = content.replace(/name: '输入文本'/g, `name: $t('batchTask.cmdInputText')`);
content = content.replace(/name: '点击屏幕'/g, `name: $t('batchTask.cmdTap')`);
content = content.replace(/name: '滑动屏幕'/g, `name: $t('batchTask.cmdSwipe')`);
content = content.replace(/name: 'Home键'/g, `name: $t('batchTask.cmdHome')`);
content = content.replace(/name: '返回键'/g, `name: $t('batchTask.cmdBack')`);
content = content.replace(/name: '电源键'/g, `name: $t('batchTask.cmdPower')`);
content = content.replace(/name: '音量\+'/g, `name: $t('batchTask.cmdVolUp')`);
content = content.replace(/name: '音量\-'/g, `name: $t('batchTask.cmdVolDown')`);
content = content.replace(/name: '截屏'/g, `name: $t('batchTask.cmdScreencap')`);
content = content.replace(/name: '录屏开始'/g, `name: $t('batchTask.cmdRecord')`);
content = content.replace(/name: '打开设置'/g, `name: $t('batchTask.cmdStartSettings')`);
content = content.replace(/name: '打开拨号'/g, `name: $t('batchTask.cmdDial')`);

// appTemplates
content = content.replace(/name: '列出应用'/g, `name: $t('batchTask.cmdListPkgs')`);
content = content.replace(/name: '列出系统应用'/g, `name: $t('batchTask.cmdListSys')`);
content = content.replace(/name: '列出第三方应用'/g, `name: $t('batchTask.cmdListThird')`);
content = content.replace(/name: '查找应用'/g, `name: $t('batchTask.cmdFindApp')`);
content = content.replace(/name: '清除应用数据'/g, `name: $t('batchTask.cmdClearApp')`);
content = content.replace(/name: '卸载应用'/g, `name: $t('batchTask.cmdUninstall')`);
content = content.replace(/name: '强制停止应用'/g, `name: $t('batchTask.cmdForceStop')`);
content = content.replace(/name: '启动应用'/g, `name: $t('batchTask.cmdStartApp')`);

// systemTemplates
content = content.replace(/name: '系统版本'/g, `name: $t('batchTask.cmdSysVer')`);
content = content.replace(/name: '设备型号'/g, `name: $t('batchTask.cmdModel')`);
content = content.replace(/name: '设备品牌'/g, `name: $t('batchTask.cmdBrand')`);
content = content.replace(/name: 'Android ID'/g, `name: $t('batchTask.cmdAndroidId')`);
content = content.replace(/name: 'IP地址'/g, `name: $t('batchTask.cmdIp')`);
content = content.replace(/name: 'CPU信息'/g, `name: $t('batchTask.cmdCpu')`);
content = content.replace(/name: '内存信息'/g, `name: $t('batchTask.cmdMem')`);
content = content.replace(/name: '存储空间'/g, `name: $t('batchTask.cmdStorage')`);
content = content.replace(/name: '电池状态'/g, `name: $t('batchTask.cmdBattery')`);
content = content.replace(/name: '屏幕分辨率'/g, `name: $t('batchTask.cmdResolution')`);
content = content.replace(/name: '屏幕密度'/g, `name: $t('batchTask.cmdDensity')`);
content = content.replace(/name: '当前Activity'/g, `name: $t('batchTask.cmdCurrentAct')`);

// Update the quickTemplates merge since they are now Computed refs:
content = content.replace(/\.\.\.basicTemplates/g, "...basicTemplates.value");
content = content.replace(/\.\.\.appTemplates/g, "...appTemplates.value");
content = content.replace(/\.\.\.systemTemplates/g, "...systemTemplates.value");


fs.writeFileSync(vuePath, content);
console.log('SUCCESS: Updated BatchTaskManagement.vue');

// ---------------- Update main.js ----------------
const mainPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainPath, 'utf8');

const zhBatchDict = `
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
      cmdExamplePlaceholder: '示例命令：\\ninput text \\'Hello World\\'\\ninput tap 500 800\\npm list packages',
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
    },`;

const enBatchDict = `
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
      cmdExamplePlaceholder: 'Example:\\ninput text \\'Hello World\\'\\ninput tap 500 800\\npm list packages',
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
    },`;


let insertCount = 0;
mainText = mainText.replace(
  /\n\s*instance: \{/g,
  function(match) {
    insertCount++;
    if (insertCount === 1) {
      return '\n' + zhBatchDict + match;
    } else if (insertCount === 2) {
      return '\n' + enBatchDict + match;
    }
    return match;
  }
);

fs.writeFileSync(mainPath, mainText);
console.log('SUCCESS: Updated main.js for batch task');
