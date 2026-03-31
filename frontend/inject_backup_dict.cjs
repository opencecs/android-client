const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainJsContent = fs.readFileSync(mainJsPath, 'utf8');

const zhBackupDict = `
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
    },`;

const enBackupDict = `
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
    },`;

let insertCount = 0;
// Insert it right before `network: {`
mainJsContent = mainJsContent.replace(
  /    network: \{/g,
  function(match) {
    insertCount++;
    if (insertCount === 1) {
      return zhBackupDict + '\n' + match;
    } else if (insertCount === 2) {
      return enBackupDict + '\n' + match;
    }
    return match;
  }
);

fs.writeFileSync(mainJsPath, mainJsContent);
console.log('SUCCESS: Injected backup dicts');
