const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainJsPath, 'utf8');

const zhHelpContent = `<div class="model-guide-container" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    📄 页面概览
  </div>
  <p style="margin: 0; padding-left: 8px;">本页面用于管理云机模拟器所使用的手机机型模板，支持查看线上机型、下载模板、编辑配置，以及将本地机型推送到设备。</p>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🌐 线上机型
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>展示服务器上所有可用的机型模板列表，包含机型 ID 和名称。</li>
    <li><b>下载：</b> 若本地尚未存有该机型模板，操作列会显示“下载”按钮，点击即可将模板下载到本地。</li>
    <li><b>编辑：</b> 若本地已有对应模板，操作列会显示“编辑”按钮，点击可修改机型的 overlay、prop 等配置项，保存后会自动生成新的机型文件并存入本地。</li>
    <li>右上角“打开本地机型下载目录”可快速打开模板的本地存储目录。</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    💾 本地机型
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>展示已存储在本地的全部机型模板，这些机型可直接推送到在线设备使用。</li>
    <li><b>推送（单个）：</b> 点击某一行的“推送”按钮，在弹窗中选择目标设备后点击“开始推送”即可。</li>
    <li><b>批量推送：</b> 勾选多个机型（或点击“全选”），再点击顶部“推送已选机型”按钮，统一推送到选中设备。</li>
    <li><b>刷新：</b> 重新扫描本地目录，更新机型列表。</li>
    <li><b>采集工具：</b> 扫描二维码下载手机信息采集 APK，用于从真机采集机型数据后导入本地。注意：仅支持采集对应 Android 版本的手机（如 Android 14 只能采集 Android 14 设备）。</li>
    <li>右上角“打开本地机型目录”可快速查看本地机型文件。</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📝 机型配置编辑
  </div>
  <p style="margin: 0; padding-left: 8px;">在线上机型页面点击“编辑”后进入配置编辑弹窗。可修改该机型的设备参数。配置分为以下几个部分：</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>model (模板名称)：</b> 只读，显示当前编辑的机型名称，不可修改。</li>
  </ul>
  
  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Overlay 参数（设备基础属性，影响云表现）</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>EMMCID:</b> eMMC 存储芯片唯一标识，空 <code> </code> 表示随机生成。</li>
    <li><b>EMMCSD:</b> eMMC 存储序列号，空 <code> </code> 表示随机生成。</li>
    <li><b>GRID:</b> GPU图形相关标识符，空 <code> </code> 表示随机生成。</li>
    <li><b>MEDIAID:</b> 媒体存储设备唯一标识，空 <code> </code> 表示随机生成。</li>
    <li><b>RNDID:</b> 随机设备标识符，空 <code> </code> 表示随机生成。</li>
    <li><b>SERIALNUMBER:</b> 设备序列号 (SN 码)，空 <code> </code> 表示随机生成。</li>
    <li><b>Aaid:</b> 匿名广告标识符 (Anonymous Advertising ID)，空 <code> </code> 表示随机生成。</li>
    <li><b>Android Id:</b> Android 系统唯一设备标识符，空 <code> </code> 表示随机生成。</li>
    <li><b>Bluetooth Name:</b> 设备蓝牙展示名称，空 <code> </code> 表示随机生成。</li>
    <li><b>GnssModel:</b> GPS/定位模组型号，例如 <code>MTK_M90_Default_MN_PDB_default</code>，影响定位识别，一般保持默认。</li>
    <li><b>Oaid:</b> 开放匿名设备标识符 (Open Anonymous Device Identifier)，空 <code> </code> 表示随机生成。</li>
    <li><b>OviLbs:</b> Overlay 底层 libs 版本标识，通常为数字 (如 <code>1</code>)，一般保持默认。</li>
    <li><b>Serial Number:</b> 设备序列号短字段，空 <code> </code> 表示随机生成。</li>
    <li><b>Vaid:</b> 厂商广告标识符 (Vendor Advertising ID)，空 <code> </code> 表示随机生成。</li>
    <li><b>Wifi List:</b> WiFi 扫描列表数量，数字类型（如 <code>10</code>），影响 WiFi 列表模拟条目数。</li>
    <li><b>Wifi Name:</b> 设备连接的 WiFi 名称，空 <code> </code> 表示随机生成。</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Prop 参数 (Android 系统属性，只读展示)</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>Ro.build.version.security_patch:</b> Android 安全补丁日期，格式 <code>YYYY-MM-DD</code>（如 <code>2023-12-05</code>），该字段为只读，不可修改。</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">关于 <code>空格</code> 占位符</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>字段值全为 <code>空格</code>，表示该字段在云机启动时会自动随机生成，无需手动填写。</li>
    <li>若需要固定某个字段的值，可将对应位置的 <code>空格</code> 替换为实际字符（需符合字段规定的字符集，如十六进制字段只能填 <code>0-9,a-f</code>）。</li>
    <li>输入了不符合规则的字符时，系统会自动将其还原为 <code>空格</code> 并给出提示。</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">保存规则</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>点击“保存”后，系统会以原机型名称 + 3 位随机数（如 <code>PJA110_182</code>）命名，生成一个新的本地机型文件，不会覆盖原装模板。</li>
    <li>保存成功后可在“本地机型” tab 中看到新生成的机型，并可直接推送到设备使用。</li>
    <li><span style="color: #F56C6C;">⚠️ 修改配置可能导致云机无法正常开机。不确定的字段请保持默认值，修改前建议记录原始值。</span></li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📤 推送设备
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>推送弹窗会列出当前所有在线设备，勾选目标设备后点击“开始推送”。</li>
    <li>推送后会自动给各设备在线状态，离线设备无法推送。</li>
    <li>支持同时推送多个机型到多台设备，推送完成后会展示成功/失败汇总。</li>
    <li>点击“刷新列表”可重新获取最新的在线设备状态。</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🔍 搜索与分页
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>顶部搜索框支持按机型名称关键词过滤，实时生效。</li>
    <li>列表默认每页显示 12 条，可通过底部分页组件翻页或调整每页数量。</li>
  </ul>
</div>`;

const enHelpContent = `<div class="model-guide-container" style="font-size: 13px; color: #606266; line-height: 2; margin-top: -10px;">
  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 16px 0 8px 0; display: flex; align-items: center;">
    📄 Overview
  </div>
  <p style="margin: 0; padding-left: 8px;">This page manages phone templates for the cloud machines. It supports viewing online templates, downloading, configuration editing, and pushing local templates to devices.</p>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🌐 Online Models
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Displays available model templates on the server, showing model ID and name.</li>
    <li><b>Download:</b> A "Download" button appears for templates not available locally. Click to fetch the template.</li>
    <li><b>Edit:</b> If already downloaded, an "Edit" button allows modifying overlay and prop variables. Saving creates a new variant.</li>
    <li>The top right button quickly opens your local templates directory.</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    💾 Local Models
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Shows stored local templates ready for pushing to online devices.</li>
    <li><b>Push (Individual):</b> Click "Push" on a row, select target devices in the popup, and hit "Start Pushing".</li>
    <li><b>Batch Push:</b> Select multiple templates, then use the top "Push Selected" button to deploy uniformly.</li>
    <li><b>Refresh:</b> Rescan your local directory and update the list.</li>
    <li><b>Capture Tool:</b> Scan the QR code to install an APK on a real phone and capture physical hardware identifiers for cloning. (Only supported on matching Android versions).</li>
    <li>The top right button quickly opens your local templates directory.</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📝 Editor Parameters
  </div>
  <p style="margin: 0; padding-left: 8px;">Clicking "Edit" on a template allows tweaking device overlays:</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>model (Template Name):</b> Read-only base model tag.</li>
  </ul>
  
  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Overlay Variables (Hardware ID Attributes)</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>EMMCID / EMMCSD:</b> eMMC storage unique chips/SN. <code>Spaces</code> = random on boot.</li>
    <li><b>GRID:</b> GPU graphic identifier. <code>Spaces</code> = random on boot.</li>
    <li><b>MEDIAID / RNDID:</b> Random media and device IDs.</li>
    <li><b>SERIALNUMBER / Serial Number:</b> Full and short Device SN. <code>Spaces</code> = random.</li>
    <li><b>Aaid / Vaid:</b> Anonymous & Vendor Ads IDs.</li>
    <li><b>Android Id:</b> Primary unresettable Android identifier. <code>Spaces</code> = random.</li>
    <li><b>Bluetooth / Wifi Name:</b> Wireless display names.</li>
    <li><b>GnssModel / OviLbs:</b> GPS module routing flags. Usually defaults are sufficient.</li>
    <li><b>Wifi List:</b> Numeric cap of simulated Wi-Fi scan results.</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Prop Variables (Read-Only OS Properties)</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li><b>Ro.build.version.security_patch:</b> OS patch date (e.g., <code>2023-12-05</code>).</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Regarding <code>Blank Spaces</code></p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Fields consisting entirely of spaces will instruct the container to inject randomized hex values upon startup automatically.</li>
    <li>If rewriting them, enforce format boundaries (e.g., hex identifiers must be <code>0-9, a-f</code>).</li>
  </ul>

  <p style="margin: 12px 0 4px 8px; font-weight: bold;">Saving Rules</p>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Clicking "Save" generates a new localized variant appending random digits (e.g., <code>PJA110_182</code>). Baseline templates are never overwritten.</li>
    <li><span style="color: #F56C6C;">⚠️ Inputting invalid config variables can break OS boots. Leave unknown fields completely blank to use resilient randomizers safely.</span></li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    📤 Pushing Devices
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>The popup relays online statuses. Offline endpoints will be filtered.</li>
    <li>Support pushing batches of templates across massive arrays of cloud instances.</li>
    <li>Always click "Refresh List" before multi-targeting to weed out disconnected hosts.</li>
  </ul>

  <div style="background-color: #ecf5ff; padding: 6px 12px; border-left: 4px solid #409eff; color: #409eff; font-weight: bold; margin: 24px 0 8px 0; display: flex; align-items: center;">
    🔍 Search & Pagination
  </div>
  <ul style="margin: 0; padding-left: 20px;">
    <li>Top input bounds immediate text-filtered lookups on model names.</li>
  </ul>
</div>`;

// Replace helpContent in main.js
mainText = mainText.replace(/helpContent:\s*`[\s\S]*?`\s*,/g, function(match, offset, string) {
  // Check vicinity to identify zh vs en
  const before = string.slice(Math.max(0, offset - 100), offset);
  if (before.includes("totalModels: '共 '") || before.includes("totalModels:")) {
    // If it's near Chinese keys
    if (before.includes('共')) {
      return 'helpContent: `' + zhHelpContent + '`,';
    } else {
      return 'helpContent: `' + enHelpContent + '`,';
    }
  }
  return match;
});

fs.writeFileSync(mainJsPath, mainText);
console.log('SUCCESS: Restored model help config blocks.');
