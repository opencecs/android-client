const fs = require('fs');
const path = 'src/components/StreamManagement.vue';
let content = fs.readFileSync(path, 'utf8');

// Replace line 520
content = content.replace(
  'title="{{ $t(\'common.addCamera\') }}"',
  ':title="$t(\'common.addCamera\')"'
);

// Add the i18n setup to the script tag block (around line 595)
if (!content.includes('const _instance = getCurrentInstance()')) {
  content = content.replace(
    "import { ref, onMounted, computed, onBeforeUnmount, watch } from 'vue'",
    "import { ref, onMounted, computed, onBeforeUnmount, watch, getCurrentInstance } from 'vue'\n\nconst _instance = getCurrentInstance()\nconst $t = (key, params) => _instance?.proxy?.$t(key, params) || key\n"
  );
}

// Replace the string literal in JS
content = content.replace(
  "item.statusReason = '{{ $t(\\'common.normal\\') }}'",
  "item.statusReason = $t('common.normal')"
);

fs.writeFileSync(path, content);
console.log('SUCCESS: Fixed StreamManagement.vue line 520 and 1042 syntax errors');
