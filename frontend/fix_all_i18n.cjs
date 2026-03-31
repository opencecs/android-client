const fs = require('fs');

// Fix HostManagement.vue - group dialogs
let host = fs.readFileSync('src/components/HostManagement.vue', 'utf8');
let hostCount = 0;

const hostReplacements = [
  // Delete group in device group command handler
  [`确定要删除分组 "\${currentGroup}" 吗？该分组下的所有设备将移至"默认分组"`,
   `\${t('common.deleteGroupMessage', {group: currentGroup})}`],
  ['\'删除分组确认\',\n      {\n        confirmButtonText: \'确定\',\n        cancelButtonText: \'取消\',',
   't(\'common.deleteGroupConfirm\'),\n      {\n        confirmButtonText: t(\'common.confirm\'),\n        cancelButtonText: t(\'common.cancel\'),'],
  // Delete group in handleDeleteGroup  
  [`确定要删除分组 "\${localGroupFilter.value}" 吗？该分组下的所有设备将移至"默认分组"`,
   `\${t('common.deleteGroupMessage', {group: localGroupFilter.value})}`],
  ['\'删除分组确认\',\n    {\n      confirmButtonText: \'确定\',\n      cancelButtonText: \'取消\',',
   't(\'common.deleteGroupConfirm\'),\n    {\n      confirmButtonText: t(\'common.confirm\'),\n      cancelButtonText: t(\'common.cancel\'),'],
  // Edit group dialog
  [`ElMessageBox.prompt('请输入新的分组名称', '编辑分组',`,
   `ElMessageBox.prompt(t('common.enterNewGroupName'), t('common.editGroup'),`],
  [`confirmButtonText: '确定',\n    cancelButtonText: '取消',\n    inputValue: localGroupFilter.value,\n    inputPattern: /\\S+/,\n    inputErrorMessage: '分组名称不能为空'`,
   `confirmButtonText: t('common.confirm'),\n    cancelButtonText: t('common.cancel'),\n    inputValue: localGroupFilter.value,\n    inputPattern: /\\S+/,\n    inputErrorMessage: t('common.groupNameCannotBeEmpty')`],
];

for (const [old, newStr] of hostReplacements) {
  if (host.includes(old)) {
    host = host.replace(old, newStr);
    hostCount++;
  } else {
    console.log('❌ Host not found:', old.substring(0, 80));
  }
}
fs.writeFileSync('src/components/HostManagement.vue', host, 'utf8');
console.log(`✅ HostManagement: Applied ${hostCount} replacements`);

// Fix App.vue - network card labels
let app = fs.readFileSync('src/App.vue', 'utf8');
let appCount = 0;

const appReplacements = [
  // Create dialog - container mode network card (2 occurrences)
  [`{{ $t('common.privateNetworkCard') }}(共享IP)`,
   `{{ $t('common.privateNetworkCard') }}({{ $t('common.sharedIP') }})`],
  [`{{ $t('common.publicNetworkCard') }}(独立IP)`,
   `{{ $t('common.publicNetworkCard') }}({{ $t('common.independentIP') }})`],
  // Create dialog - simulator mode network card (hardcoded Chinese)
  [`私有网卡(共享IP)`, `{{ $t('common.privateNetworkCard') }}({{ $t('common.sharedIP') }})`],
  [`公有网卡(独立IP)`, `{{ $t('common.publicNetworkCard') }}({{ $t('common.independentIP') }})`],
];

for (const [old, newStr] of appReplacements) {
  let found = 0;
  while (app.includes(old)) {
    app = app.replace(old, newStr);
    appCount++;
    found++;
  }
  if (found === 0) {
    console.log('❌ App not found:', old);
  }
}
fs.writeFileSync('src/App.vue', app, 'utf8');
console.log(`✅ App.vue: Applied ${appCount} replacements`);

// Fix BatchImport.vue
let batch = fs.readFileSync('src/components/BatchImport.vue', 'utf8');
let batchCount = 0;

const batchReplacements = [
  [`<span style="font-weight: bold;">重要提示（点击展开/收起）</span>`,
   `<span style="font-weight: bold;">{{ $t('common.importantNotice') }}</span>`],
  [`<el-tag type="warning" style="margin-right: 10px;">支持范围</el-tag>`,
   `<el-tag type="warning" style="margin-right: 10px;">{{ $t('common.supportScope') }}</el-tag>`],
  [`仅支持模拟器镜像类型的云机备份，<span style="color: #F56C6C; font-weight: bold;">容器模式镜像不支持</span>备份导入功能`,
   `{{ $t('common.supportScopeDesc') }}<span style="color: #F56C6C; font-weight: bold;">{{ $t('common.containerNotSupported') }}</span>{{ $t('common.backupImportFunction') }}`],
  [`<el-tag type="info" style="margin-right: 10px;">导入前准备</el-tag>`,
   `<el-tag type="info" style="margin-right: 10px;">{{ $t('common.importPreparation') }}</el-tag>`],
  [`请确保目标设备已下载对应的安卓镜像，镜像未下载将导致导入失败或者启动失败`,
   `{{ $t('common.importPreparationDesc') }}`],
  [`<el-tag type="danger" style="margin-right: 10px;">设备兼容性</el-tag>`,
   `<el-tag type="danger" style="margin-right: 10px;">{{ $t('common.deviceCompatibility') }}</el-tag>`],
  [`备份文件必须与目标设备类型完全匹配（<span style="color: #67C23A; font-weight: bold;">CQR系列镜像可复用</span>，<span style="color: #F56C6C; font-weight: bold;">P系列不能复用</span>）`,
   `{{ $t('common.deviceCompatibilityDesc') }}<span style="color: #67C23A; font-weight: bold;">{{ $t('common.cqrReusable') }}</span>, <span style="color: #F56C6C; font-weight: bold;">{{ $t('common.pSeriesNotReusable') }}</span>)`],
  [`<span style="font-weight: bold;">备份文件列表</span>`,
   `<span style="font-weight: bold;">{{ $t('common.backupFileList') }}</span>`],
  // Open import folder button text (in header)
  [`<el-icon><FolderOpened /></el-icon>\r\n              打开导入文件夹`,
   `<el-icon><FolderOpened /></el-icon>\r\n              {{ $t('common.openImportFolder') }}`],
  [`<el-icon><Refresh /></el-icon>\r\n              刷新`,
   `<el-icon><Refresh /></el-icon>\r\n              {{ $t('common.refresh') }}`],
  [`<el-table-column prop="name" label="文件名" min-width="200" />`,
   `<el-table-column prop="name" :label="$t('common.fileName')" min-width="200" />`],
  [`label="文件大小"`, `:label="$t('common.fileSize')"`],
  [`label="操作"`, `:label="$t('common.operation')"`],
  // Import/Delete buttons
  [`\r\n              导入\r\n            </el-button>\r\n            <el-button type="danger" size="small" @click="handleDelete(row)">\r\n              删除`,
   `\r\n              {{ $t('common.import') }}\r\n            </el-button>\r\n            <el-button type="danger" size="small" @click="handleDelete(row)">\r\n              {{ $t('common.delete') }}`],
  // Empty state
  [`暂无备份文件`, `{{ $t('common.noBackupFiles') }}`],
  // Open import folder button (empty state)
  [`<el-icon><FolderOpened /></el-icon>\r\n                打开导入文件夹`,
   `<el-icon><FolderOpened /></el-icon>\r\n                {{ $t('common.openImportFolder') }}`],
  [`请将云机备份文件（.tar.gz 格式）复制到导入文件夹中<br>然后点击右上角"刷新"按钮`,
   `{{ $t('common.copyToImportFolder') }}<br>{{ $t('common.thenRefresh') }}`],
  // Import dialog
  [`title="批量导入"`, `:title="$t('common.batchImport')"`],
  [`<el-step title="选择设备"`, `<el-step :title="$t('common.selectDevice')"`],
  [`<el-step title="配置坑位"`, `<el-step :title="$t('common.configureSlots')"`],
  [`<el-step title="执行导入"`, `<el-step :title="$t('common.executeImport')"`],
  [`placeholder="搜索设备IP"`, `:placeholder="$t('common.searchDeviceIP')"`],
  [`prop="ip" label="设备IP"`, `prop="ip" :label="$t('common.deviceIP')"`],
  [`label="主机固件版本"`, `:label="$t('common.hostFirmwareVersion')"`],
  [`label="NVME存储空间"`, `:label="$t('common.nvmeStorage')"`],
  [`description="暂无可用设备"`, `:description="$t('common.noAvailableDevices')"`],
  [`\r\n            下一步\r\n`, `\r\n            {{ $t('common.nextStep') }}\r\n`],
  [`>全选</el-button>`, `>{{ $t('common.selectAll') }}</el-button>`],
  [`>清空</el-button>`, `>{{ $t('common.clear') }}</el-button>`],
  [`>批量设置复制份数</el-button>`, `>{{ $t('common.batchSetCopyCount') }}</el-button>`],
  [`<div class="slot-number">坑位`, `<div class="slot-number">{{ $t('common.slotLabel') }}`],
  [`<span style="font-size: 12px; color: #666;">份</span>`,
   `<span style="font-size: 12px; color: #666;">{{ $t('common.copies') }}</span>`],
  [`<el-button @click="prevStep">上一步</el-button>`,
   `<el-button @click="prevStep">{{ $t('common.previousStep') }}</el-button>`],
  [`\r\n            开始导入\r\n`, `\r\n            {{ $t('common.startImport') }}\r\n`],
  [`<el-icon><Check /></el-icon>\r\n            完成`,
   `<el-icon><Check /></el-icon>\r\n            {{ $t('common.complete') }}`],
];

for (const [old, newStr] of batchReplacements) {
  if (batch.includes(old)) {
    batch = batch.replace(old, newStr);
    batchCount++;
  } else {
    console.log('❌ Batch not found:', old.substring(0, 80));
  }
}
fs.writeFileSync('src/components/BatchImport.vue', batch, 'utf8');
console.log(`✅ BatchImport.vue: Applied ${batchCount} replacements`);
