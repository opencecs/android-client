const fs = require('fs');

const filePath = 'src/components/backupManagement.vue';
let text = fs.readFileSync(filePath, 'utf8');

// Check if already translated
if (text.includes('Backup Management Usage Guide')) {
    console.log('Already translated!');
    process.exit(0);
}

// The Chinese guide block starts at the outer wrapper div inside the help tab-pane
// We find the exact block: from '<div style="margin: 20px; padding: 0;">' 
// to its closing '</div>' right before '</div>\n                </el-tab-pane>'

const marker = '<!-- 使用说明 -->';
const markerIdx = text.indexOf(marker);
if (markerIdx === -1) throw new Error('marker not found');

// Find the content div after the tab-pane opening
const outerDivMarker = '<div style="height: 100%; overflow-y: auto; box-sizing: border-box;">';
const outerStart = text.indexOf(outerDivMarker, markerIdx);
if (outerStart === -1) throw new Error('outer div not found');

// Find the closing </el-tab-pane> for this help tab
const tabPaneClose = '</el-tab-pane>';
const tabCloseIdx = text.indexOf(tabPaneClose, outerStart);
if (tabCloseIdx === -1) throw new Error('tab close not found');

// Extract the Chinese block (from outerStart to just before tabCloseIdx)
const zhBlock = text.substring(outerStart, tabCloseIdx).trimEnd();

// Build the English block by translating all Chinese text
let enBlock = zhBlock;

const translations = [
    ['备份管理使用说明', 'Backup Management Usage Guide'],
    ['📦 备份机型', '📦 Backup Model'],
    ['用于将设备上的云机配置（机型模板）导出保存到本地，或将本地备份的机型导入到其他设备。', 'Export VM configurations (model templates) from devices to local storage, or import locally backed-up models to other devices.'],
    ['<strong>导出</strong>：在左侧选择设备后，右侧列表会显示该设备上的备份机型。点击"导出"按钮，将机型文件下载到本地保存。', '<strong>Export</strong>: After selecting a device on the left, the right panel shows its backup models. Click "Export" to download the model file to local storage.'],
    ['：若本地已存在同名机型文件，操作列将显示"导入"按钮。点击后选择目标设备，即可将该机型批量导入到所选设备。', ': If a local model file with the same name exists, an "Import" button will appear. Click it and select target devices to batch-import the model.'],
    ['<strong>新增</strong>：点击右上角"新增"按钮，选择设备上的云机并命名，系统将对该云机的机型进行备份。', '<strong>Add</strong>: Click the "Add" button in the upper right, select a VM on the device and name it. The system will back up that VM\'s model.'],
    ['：点击"删除"按钮，可删除设备上保存的对应备份机型记录。', ': Click the "Delete" button to remove the corresponding backup model record from the device.'],
    ['点击"打开本地备份机型目录"可快速查看本地已保存的机型文件。', 'Click "Open Local Backup Model Directory" to quickly browse locally saved model files.'],
    ['☁️ 备份云机', '☁️ Backup VM'],
    ['用于将设备上的完整云机镜像文件导出到本地，或将本地已下载的云机镜像恢复到指定设备和坑位。', 'Export complete VM image files from devices to local storage, or restore locally downloaded VM images to specified devices and slots.'],
    ['<strong>下载</strong>：点击"下载"按钮，将设备上的云机备份文件下载到本机保存。下载过程中按钮会显示"下载中"状态，请勿重复点击。', '<strong>Download</strong>: Click "Download" to save the VM backup file locally. The button will show "Downloading" status during the process — do not click repeatedly.'],
    ['<strong>导入</strong>：本地已下载的云机备份会显示"导入"按钮。点击后选择目标设备、填写坑位号及新云机名称，确认后将进行恢复操作。', '<strong>Import</strong>: Locally downloaded VM backups will show an "Import" button. Click it, select the target device, fill in the slot number and new VM name, then confirm to perform the restore.'],
    ['<strong>新增</strong>：点击右上角"新增"按钮，选择设备上的云机，系统将对该云机进行整机备份（文件较大，请耐心等待）。', '<strong>Add</strong>: Click the "Add" button in the upper right, select a VM on the device. The system will perform a full VM backup (files are large — please be patient).'],
    ['<strong>删除</strong>：删除操作将同时删除设备端备份记录及本地已下载的文件，请谨慎操作。', '<strong>Delete</strong>: The delete operation removes both the device-side backup record and locally downloaded files — please proceed with caution.'],
    ['点击"打开本地备份云机目录"可查看本地已下载的云机镜像文件。', 'Click "Open Local Backup VM Directory" to view locally downloaded VM image files.'],
    ['⚠️ 注意：导入备份云机前请确保镜像文件已完整下载到本地，且目标设备处于在线状态。', '⚠️ Note: Before importing a backup VM, ensure the image file has been fully downloaded locally and the target device is online.'],
    ['🚀 批量导入', '🚀 Batch Import'],
    ['支持将本地已有的云机备份文件批量分发并导入到多台设备，适合大规模初始化场景。', 'Supports batch distribution and import of locally available VM backup files to multiple devices — ideal for large-scale initialization scenarios.'],
    ['在"批量导入"标签页中，选择本地备份文件，再勾选目标设备，点击"开始导入"即可批量执行。', 'In the "Batch Import" tab, select local backup files, check target devices, then click "Start Import" to execute in batch.'],
    ['导入进度会实时显示，成功/失败结果会分别统计汇报。', 'Import progress is displayed in real-time, with success/failure results reported separately.'],
    ['批量导入期间请保持设备在线，避免网络中断导致导入失败。', 'Keep devices online during batch import to avoid import failures caused by network interruptions.'],
];

for (const [zh, en] of translations) {
    enBlock = enBlock.split(zh).join(en);
}

// Also translate $t() references that appear in the content
// {{ $t('backup.importBtn') }} and {{ $t('common.delete') }} are dynamic, keep as-is

// Wrap zhBlock with v-if and enBlock with v-else
const zhWrapped = zhBlock.replace(
    '<div style="height: 100%; overflow-y: auto; box-sizing: border-box;">',
    '<div v-if="$i18n.locale === \'zh-CN\'" style="height: 100%; overflow-y: auto; box-sizing: border-box;">'
);

const enWrapped = enBlock.replace(
    '<div style="height: 100%; overflow-y: auto; box-sizing: border-box;">',
    '<div v-else style="height: 100%; overflow-y: auto; box-sizing: border-box;">'
);

// Replace original block
const newContent = text.substring(0, outerStart) + zhWrapped + '\n' + enWrapped + '\n                ' + text.substring(tabCloseIdx);

fs.writeFileSync(filePath, newContent);
console.log('SUCCESS: Injected English translation for Backup guide.');
