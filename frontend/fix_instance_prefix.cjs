const fs = require('fs');
const path = require('path');

const vuePath = path.join(__dirname, 'src', 'components', 'instanceManagement.vue');
let vueText = fs.readFileSync(vuePath, 'utf8');

vueText = vueText.replace(/Instance-\{\{ key \}\}/g, `{{ $t('instance.instancePrefix') }}-{{ key }}`);
vueText = vueText.replace(/Instance-\{\{ item\.instanceKey \}\}/g, `{{ $t('instance.instancePrefix') }}-{{ item.instanceKey }}`);
vueText = vueText.replace(/Instance-\{\{ scope\.row\.slot \}\}/g, `{{ $t('instance.instancePrefix') }}-{{ scope.row.slot }}`);

fs.writeFileSync(vuePath, vueText);
console.log('SUCCESS: updated instanceManagement.vue');

const mainPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainPath, 'utf8');

// Replace instancePrefix in zh-CN
mainText = mainText.replace(
  /    instance: \{([\s\S]*?)totalCount: '共有 \{count\} 台云机实例'/g,
  "    instance: {$1totalCount: '共有 ',\n      instancePrefix: '云机'"
);

// Replace instancePrefix in en-US
mainText = mainText.replace(
  /    instance: \{([\s\S]*?)totalCount: 'Total \{count\} instances'/g,
  "    instance: {$1totalCount: 'Total ',\n      instancePrefix: 'Instance'"
);

fs.writeFileSync(mainPath, mainText);
console.log('SUCCESS: updated main.js');
