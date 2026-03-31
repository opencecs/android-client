const fs = require('fs');
const path = require('path');

const vuePath = path.join(__dirname, 'src', 'components', 'StreamManagement.vue');
let content = fs.readFileSync(vuePath, 'utf8');

// Fix the doubled $t interpolation
content = content.replace(/\{\{\s*\$t\('\{\{\s*\$t\('stream\.webrtcAddress'\)\s*\}\}'\)\s*\}\}/g, `{{ $t('stream.webrtcAddress') }}`);
content = content.replace(/\{\{\s*\$t\('\{\{\s*\$t\('stream\.rtmpAddress'\)\s*\}\}'\)\s*\}\}/g, `{{ $t('stream.rtmpAddress') }}`);

fs.writeFileSync(vuePath, content);
console.log('SUCCESS: Fixed extra $t in StreamManagement.vue');
