const fs = require('fs');
const path = require('path');

const vuePath = path.join(__dirname, 'src', 'components', 'StreamManagement.vue');
let content = fs.readFileSync(vuePath, 'utf8');

content = content.replace(/ElMessage\.error\('\{\{\s*\$t\('stream\.deviceOfflineOrNotExist'\)\s*\}\}'\)/g, `ElMessage.error(_instance.proxy.$t('stream.deviceOfflineOrNotExist'))`);

fs.writeFileSync(vuePath, content);
console.log('SUCCESS: Fixed JS compilation block again');
