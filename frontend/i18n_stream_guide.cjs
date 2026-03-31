const fs = require('fs');
const path = require('path');

const vuePath = path.join(__dirname, 'src', 'components', 'StreamManagement.vue');
let content = fs.readFileSync(vuePath, 'utf8');

// Fix JS expression
content = content.replace(/'正在查询状态\.\.\.' \: '暂无' \+ \$t\('common\.activeStatus'\) \+ '分发'/g, `refreshingDistributions ? $t('stream.queryingStatus') : $t('stream.noActiveDistribution')`);
// If the ternary operator gets duplicated or messed up, simply replace the whole line:
content = content.replace(/\{\{\s*refreshingDistributions\s*\?\s*'正在查询状态\.\.\.'\s*\:\s*'暂无'\s*\+\s*\$t\('common\.activeStatus'\)\s*\+\s*'分发'\s*\}\}/g, `{{ refreshingDistributions ? $t('stream.queryingStatus') : $t('stream.noActiveDistribution') }}`);

// Fix common prefix mapping to stream prefix
content = content.replace(/\$t\('common\.streamGuideHtml'\)/g, `$t('stream.streamGuideHtml')`);
content = content.replace(/\$t\('common\.activeStatus'\)/g, `$t('stream.activeStatus')`);
content = content.replace(/\$t\('common\.idleStatus'\)/g, `$t('stream.idleStatus')`);

fs.writeFileSync(vuePath, content);
console.log('SUCCESS: Updated StreamManagement.vue expressions');

// ---------------- Update main.js ----------------
const mainPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainPath, 'utf8');

const zhStreamAdd = `
      queryingStatus: '正在查询状态...',
      activeStatus: '推流中',
      idleStatus: '闲置',
      streamGuideHtml: \`<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
<h3 style="color: #409EFF; border-bottom: 2px solid #409EFF; padding-bottom: 5px; margin-top: 0;">流媒体使用说明</h3>
<ul>
<li><b>转发流模式（OBS推流）：</b> 您可以使用OBS等串流软件，将画面推流至此处分配的地址。平台会自动将该流媒体分发和桥接。</li>
<li><b>点对点模式：</b> 使用WebRTC进行的无延迟点对点低延迟投屏推流模式。</li>
</ul>
</div>\`
    },`;

const enStreamAdd = `
      queryingStatus: 'Querying status...',
      activeStatus: 'Active',
      idleStatus: 'Idle',
      streamGuideHtml: \`<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
<h3 style="color: #409EFF; border-bottom: 2px solid #409EFF; padding-bottom: 5px; margin-top: 0;">Stream Usage Guide</h3>
<ul>
<li><b>Forward Stream (OBS):</b> Use OBS or similar streaming software to push video streams to the corresponding allocated addresses. The platform will automatically distribute and bridge the media.</li>
<li><b>P2P Mode:</b> Point-to-point WebRTC low-latency streaming and casting.</li>
</ul>
</div>\`
    },`;

// We inject the new keys before the closing `},` of the `stream: {` block
// For zh-CN
mainText = mainText.replace(
  /      rtmpAddress: 'RTMP地址'\n    \},/g,
  `      rtmpAddress: 'RTMP地址',\n${zhStreamAdd}`
);

// For en-US
mainText = mainText.replace(
  /      rtmpAddress: 'RTMP URL'\n    \},/g,
  `      rtmpAddress: 'RTMP URL',\n${enStreamAdd}`
);

fs.writeFileSync(mainPath, mainText);
console.log('SUCCESS: Updated main.js for stream expansion');
