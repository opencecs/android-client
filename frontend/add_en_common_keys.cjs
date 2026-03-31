const fs = require('fs');
let content = fs.readFileSync('src/main.js', 'utf8');

// Add English common keys
const oldEnCommon = "selectGroupPlaceholder: 'Select group'";
const newEnCommon = `selectGroupPlaceholder: 'Select group',
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
      complete: 'Complete'`;

if (content.includes(oldEnCommon)) {
  content = content.replace(oldEnCommon, newEnCommon);
  console.log('✅ English common keys added');
} else {
  console.log('❌ Target not found');
}

fs.writeFileSync('src/main.js', content, 'utf8');
console.log('Done');
