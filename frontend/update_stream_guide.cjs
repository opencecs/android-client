const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainJsPath, 'utf8');

const newZhGuide = `<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
<h3 style="color: #409EFF; border-bottom: 2px solid #409EFF; padding-bottom: 5px; margin-top: 0;">流媒体使用说明</h3>
<ul>
<li><b>转发流模式（OBS推流）：</b> 您可以使用OBS等串流软件，将画面推流至此处分配的地址。平台会自动将该流媒体分发和桥接。</li>
<li><b>点对点模式：</b> 使用WebRTC进行的无延迟点对点低延迟投屏推流模式。</li>
<li><b>PC摄像头模式：</b> 直接将本地电脑的USB摄像头或虚拟摄像头画面挂载桥接到云机内部的相机接口。</li>
</ul>
</div>`;

const newEnGuide = `<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
<h3 style="color: #409EFF; border-bottom: 2px solid #409EFF; padding-bottom: 5px; margin-top: 0;">Stream Usage Guide</h3>
<ul>
<li><b>Forward Stream (OBS):</b> Use OBS or similar software to push video streams to the allocated addresses. The platform will automatically distribute the stream to cloud machines.</li>
<li><b>P2P Mode:</b> Point-to-point WebRTC low-latency streaming and screen mirroring mode.</li>
<li><b>PC Camera Mode:</b> Directly bridge and mount your local PC's USB or virtual cameras to the cloud machine's internal camera interface.</li>
</ul>
</div>`;

// Replace the old HTML guides with the new ones
mainText = mainText.replace(/streamGuideHtml: `<div style="font-family: Arial, sans-serif;[\s\S]*?<\/div>`/g, function(match, offset, string) {
  // If it's the first occurrence, it's Chinese. Otherwise, it's English.
  const before = string.slice(Math.max(0, offset - 100), offset);
  if (before.includes("idleStatus: '闲置'")) {
    return 'streamGuideHtml: `' + newZhGuide + '`';
  } else {
    return 'streamGuideHtml: `' + newEnGuide + '`';
  }
});

fs.writeFileSync(mainJsPath, mainText);
console.log('SUCCESS: Updated HTML Guide');
