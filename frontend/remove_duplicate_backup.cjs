const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainJsContent = fs.readFileSync(mainJsPath, 'utf8');

// The Chinese duplicate block signature:
//    backup: {
//      importBackupMachine: '导入备份云机',
// ...
//      noAvailableSlots: '无可用坑位'
//    }

const zhOldBackupRegex = /,\s*backup: \{\s*importBackupMachine: '导入备份云机',[\s\S]*?noAvailableSlots: '[^']+'\s*\}/;
mainJsContent = mainJsContent.replace(zhOldBackupRegex, '');

// The English duplicate block signature:
//    backup: {
//      importBackupMachine: 'Import Backup Machine',
// ...
//      noAvailableSlots: 'No available slots'
//    }

const enOldBackupRegex = /,\s*backup: \{\s*importBackupMachine: 'Import Backup Machine',[\s\S]*?noAvailableSlots: '[^']+'\s*\}/;
mainJsContent = mainJsContent.replace(enOldBackupRegex, '');

fs.writeFileSync(mainJsPath, mainJsContent);
console.log('SUCCESS: Removed duplicate backup blocks');
