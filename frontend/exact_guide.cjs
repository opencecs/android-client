const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainJsPath, 'utf8');

const newZhGuide = `<div class="stream-guide" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-connection" style="margin-right: 8px;"></i> 转发流（OBS）
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>打开 OBS，进入 <b>设置 -> 直播</b>，服务选择 <b>自定义</b>。</li>
    <li>将首页展示的 WebRTC 推流地址 或 RTMP 推流地址 填入"服务器"栏（点击地址可一键复制）。</li>
    <li>在"推流码"栏填入自定义的推流码。点击确定后在 OBS 中点击 <b>开始直播</b>。</li>
    <li>回到本页面，点击 <b>新建流</b>，输入与推流码 <b>相同的名称</b> 后确认。
      <div style="color: #e6a23c; font-size: 12px; margin-top: 4px;"><i class="el-icon-info"></i> 命名一致后该条目将永远保留在列表中，推流状态均可见。</div>
    </li>
    <li>列表该项对应条目且状态变为 <b>活跃</b> 后，点击 <b>分发</b> 按钮，选择目标设备和云机。</li>
    <li>在云机上打开 <b>相机</b> 即可查看对应画面。</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-link" style="margin-right: 8px;"></i> 点对点（P2P / OBS）
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>切换到 <b>点对点(OBS)</b> 标签，点击 <b>添加 P2P 流</b>。</li>
    <li>在弹窗中选择 <b>设备</b> 和 <b>云机</b>，保存后列表会显示对应的 <b>监听地址</b>（点击可一键复制）。</li>
    <li>打开 OBS，进入 <b>设置 -> 直播</b>，服务选择 <b>自定义</b>，将监听地址填入"服务器"栏，无需填写推流码，点击确定。</li>
    <li>在 OBS 中点击 <b>开始直播</b>，然后在在列表中的 <b>启动 P2P</b> 按钮。</li>
    <li>在对应云机上打开 <b>相机</b> 即可查看实时画面。</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-video-camera" style="margin-right: 8px;"></i> PC 摄像头
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>切换到 <b>PC摄像头</b> 标签，点击 <b>添加摄像头推流</b>。</li>
    <li>在弹窗中选择 <b>设备</b>、<b>云机</b> 以及可用的 <b>摄像头</b>，配置分辨率与码率后保存。</li>
    <li>在列表中点击 <b>启动</b> 按钮，右侧预览区会自动启动展示实时画面。</li>
    <li>在对应云机上打开 <b>相机</b> 即可查看推流画面。</li>
  </ol>
</div>`;

const newEnGuide = `<div class="stream-guide" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-connection" style="margin-right: 8px;"></i> Forward Stream (OBS)
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>Open OBS, go to <b>Settings -> Stream</b>, and set Service to <b>Custom</b>.</li>
    <li>Copy the WebRTC or RTMP URL and paste it into "Server" (click URL to copy).</li>
    <li>Enter a custom stream key. Click OK, then click <b>Start Streaming</b> in OBS.</li>
    <li>Return here, click <b>New Stream</b>, and enter the <b>exact same name</b> as your stream key.
      <div style="color: #e6a23c; font-size: 12px; margin-top: 4px;"><i class="el-icon-info"></i> With the same name, this entry remains in the list displaying real-time stream status.</div>
    </li>
    <li>Once the status is <b>Active</b>, click <b>Distribute</b> and select the target device and machine.</li>
    <li>Open the <b>Camera</b> app inside the cloud machine to view the stream.</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-link" style="margin-right: 8px;"></i> Peer-to-Peer (P2P / OBS)
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>Switch to <b>P2P (OBS)</b> tab and click <b>Add P2P Stream</b>.</li>
    <li>Select <b>Device</b> and <b>Machine</b>. The list will display a <b>Listen Address</b> (click to copy).</li>
    <li>Open OBS, go to <b>Settings -> Stream</b>, set Service to <b>Custom</b>, and paste the listen address into "Server". Leave Stream Key empty.</li>
    <li>Click <b>Start Streaming</b> in OBS, then click the <b>Start P2P</b> button in the list here.</li>
    <li>Open the <b>Camera</b> app inside the cloud machine to view the feed.</li>
  </ol>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-radius: 4px; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    <i class="el-icon-video-camera" style="margin-right: 8px;"></i> PC Camera
  </div>
  <ol style="margin: 0; padding-left: 20px;">
    <li>Switch to <b>PC Camera</b> tab and click <b>Add Camera Stream</b>.</li>
    <li>Select <b>Device, Machine</b>, and <b>Camera</b>. Config resolution & bitrate, save.</li>
    <li>Click <b>Start</b>. The right preview area will automatically show live feed.</li>
    <li>Open the <b>Camera</b> app inside the cloud machine to view the live feed.</li>
  </ol>
</div>`;


// Need to be careful to match my previous replacement pattern
mainText = mainText.replace(/streamGuideHtml: `[\s\S]*?<\/div>`/g, function(match, offset, string) {
  const before = string.slice(Math.max(0, offset - 100), offset);
  if (before.includes("idleStatus: '闲置'") || before.includes("idleStatus: 'Idle'")) {
    // Determine if it is the chinese block
    if (before.includes("idleStatus: '闲置'")) return 'streamGuideHtml: `' + newZhGuide + '`';
    return 'streamGuideHtml: `' + newEnGuide + '`';
  }
  return match;
});


fs.writeFileSync(mainJsPath, mainText);
console.log('SUCCESS: Put back identical HTML structure');
