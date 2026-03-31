const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainJsPath, 'utf8');

const zhGuideContent = `<div class="image-guide-container" style="font-size: 13px; color: #606266; line-height: 2;">
  <h3 style="color: #303133; font-size: 16px; margin-bottom: 20px; display: flex; align-items: center;">
    <i class="el-icon-picture-outline" style="margin-right: 8px;"></i> 镜像管理使用说明
  </h3>

  <!-- 1. 功能介绍 -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    📄 功能介绍
  </div>
  <p style="margin-top: 0; margin-bottom: 12px;">镜像管理模块用于管理运行在设备上的 Android 系统镜像，支持从云端下载镜像到本地、将本地镜像上传到设备，以及查看和删除设备上已有的镜像。主要包含三个功能页签：</p>
  <ul style="padding-left: 20px; margin-bottom: 24px;">
    <li><b>在线镜像：</b> 列表云端可用镜像列表，可按设备型号、类型（模拟器/容器）、Android 版本等条件筛选，支持一键下载到本地或直接上传到设备。</li>
    <li><b>本地镜像：</b> 查看已下载到本地的镜像文件，可将本地镜像上传至指定设备，也可删除本地缓存文件以释放磁盘空间。</li>
    <li><b>设备镜像：</b> 选择某台设备，查看该设备上已安装的镜像列表，并可对其进行删除操作。</li>
  </ul>

  <!-- 2. 典型使用流程 -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    🔄 典型使用流程
  </div>
  <div style="margin-bottom: 30px; position: relative;">
    <ol style="padding-left: 10px; margin: 0; list-style: none;">
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">1</span> <div><b style="color: #409eff;">选择在线镜像</b><br><span style="color: #409eff;">在「在线镜像」列表中，找对你的设备和系统要求的镜像，核对版本和类型。</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">2</span> <div><b style="color: #409eff;">下载到本地</b><br><span style="color: #409eff;">点击「下载」按钮，将镜像文件缓存到本机存储，以便下次使用。</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">3</span> <div><b style="color: #409eff;">上传到设备</b><br><span style="color: #409eff;">在「本地镜像」或「在线镜像」中点击「上传到设备」，选中相应的设备完成推送。</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">4</span> <div><b style="color: #409eff;">验证安装</b><br><span style="color: #409eff;">切换到「设备镜像」页签，选择对应设备，确认镜像已成功安装在设备磁盘空间中。</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">5</span> <div><b style="color: #409eff;">启动云机</b><br><span style="color: #409eff;">回到「云机」页面，使用已上传的镜像创建云机并启动运行。</span></div></li>
    </ol>
  </div>

  <!-- 3. 模拟器 vs 容器：有什么区别？ -->
  <div style="font-weight: bold; margin-bottom: 12px; font-size: 14px;">模拟器 vs 容器：有什么区别？</div>
  <table style="width: 100%; border-collapse: collapse; margin-bottom: 24px; text-align: left; font-size: 12px;">
    <thead>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <th style="padding: 10px 0; width: 15%; color:#606266;">对比维度</th>
        <th style="padding: 10px 0; width: 42%; color:#409eff;">🖥️ 模拟器 (Simulator)</th>
        <th style="padding: 10px 0; width: 43%; color:#e6a23c;">📦 容器 (Container)</th>
      </tr>
    </thead>
    <tbody>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">运行原理</td>
        <td style="padding: 8px 0;">基于 QEMU 等虚拟机技术，完整模拟 ARM 硬件</td>
        <td style="padding: 8px 0;">基于 Linux 容器（如 Docker）轻量隔离，共享宿主机内核</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">启动速度</td>
        <td style="padding: 8px 0;">较慢，通常需要 50 秒以上</td>
        <td style="padding: 8px 0;">快，通常 3~10 秒即可启动</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">资源占用</td>
        <td style="padding: 8px 0;">较高，每实例需独立分配 CPU/内存</td>
        <td style="padding: 8px 0;">较低，多实例可共享宿主机资源，密度越高</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">并发数量</td>
        <td style="padding: 8px 0;">受限于宿主机性能，同时运行数量较少</td>
        <td style="padding: 8px 0;">没有硬屏障，单机支持数十甚至上百个实例</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">安卓兼容性</td>
        <td style="padding: 8px 0;">与真机行为高度一致，兼容性更好</td>
        <td style="padding: 8px 0;">部分依赖底层硬件的功能可能受限</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">图形界面</td>
        <td style="padding: 8px 0;">支持完整 GPU 渲染，画面流畅</td>
        <td style="padding: 8px 0;">视网显卡渲染，图形性能稍弱</td>
      </tr>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">适用场景</td>
        <td style="padding: 8px 0;">UI 测试、游戏运行、重度图形需求</td>
        <td style="padding: 8px 0;">自动化脚本、批量任务、高密度部署</td>
      </tr>
    </tbody>
  </table>

  <!-- 4. 选型建议 -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #e6a23c;">
    💡 选型建议
  </div>
  <div style="background-color: #f0f9eb; padding: 12px; border-radius: 4px; border-left: 4px solid #67c23a; margin-bottom: 8px;">
    <div style="display: flex; align-items: flex-start;">
      <i class="el-icon-success" style="color: #67c23a; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
      <div>
        <div style="font-weight: bold; color: #67c23a; margin-bottom: 4px;">推荐使用容器镜像的场景：</div>
        <div style="color: #67c23a;">需要大规模并发运行多个 Android 实例（如自动化脚本、批量任务）；对启动速度和资源利用率有较高要求；无需图形界面的无头运行场景。</div>
      </div>
    </div>
  </div>
  <div style="background-color: #fdf6ec; padding: 12px; border-radius: 4px; border-left: 4px solid #e6a23c; margin-bottom: 24px;">
    <div style="display: flex; align-items: flex-start;">
       <i class="el-icon-warning" style="color: #e6a23c; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
       <div>
         <div style="font-weight: bold; color: #e6a23c; margin-bottom: 4px;">推荐使用模拟器镜像的场景：</div>
         <div style="color: #e6a23c;">需要完整的 Android 图形界面体验；需要与真机行为高度一致的测试环境；部分对硬件加载有强依赖的应用场景。</div>
       </div>
    </div>
  </div>

  <!-- 5. 注意事项 -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #F56C6C;">
    ⚠️ 注意事项
  </div>
  <ul style="padding-left: 20px; color: #606266; margin-bottom: 0;">
    <li>镜像文件体积较大（通常 1GB 以上），下载前请确保本机磁盘空间充足。</li>
    <li>上传镜像到设备需要设备在线且网络稳定，上传过程中请勿断开连接。</li>
    <li>删除设备上的镜像后，使用该镜像的实例将无法再启动，请谨慎操作。</li>
    <li><b>P1 型号</b>设备不支持 Android 12 容器镜像，请选择 Android 10 或 Android 14。</li>
  </ul>
</div>`;

const enGuideContent = `<div class="image-guide-container" style="font-size: 13px; color: #606266; line-height: 2;">
  <h3 style="color: #303133; font-size: 16px; margin-bottom: 20px; display: flex; align-items: center;">
    <i class="el-icon-picture-outline" style="margin-right: 8px;"></i> Image Management Guide
  </h3>

  <!-- 1. Introduction -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    📄 Feature Introduction
  </div>
  <p style="margin-top: 0; margin-bottom: 12px;">The Image Management module manages Android system images running on your devices. It supports downloading images from the cloud, pushing them to devices, and managing existing images. It consists of three main tabs:</p>
  <ul style="padding-left: 20px; margin-bottom: 24px;">
    <li><b>Online Images:</b> Lists available cloud images. You can filter by device model, type, or Android version, and download or push them.</li>
    <li><b>Local Images:</b> Views templates downloaded locally. Push them directly to connected devices or delete them to free up disk space.</li>
    <li><b>Device Images:</b> Verify images natively installed on specific devices and safely delete them if unneeded.</li>
  </ul>

  <!-- 2. Workflow -->
  <div style="font-weight: bold; color: #409eff; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center;">
    🔄 Typical Workflow
  </div>
  <div style="margin-bottom: 30px; position: relative;">
    <ol style="padding-left: 10px; margin: 0; list-style: none;">
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">1</span> <div><b style="color: #409eff;">Select Online Image</b><br><span style="color: #409eff;">Find an image compatible with your device in the Online list.</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">2</span> <div><b style="color: #409eff;">Download Locally</b><br><span style="color: #409eff;">Click Download to cache the image locally for faster reuse.</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">3</span> <div><b style="color: #409eff;">Push to Device</b><br><span style="color: #409eff;">Trigger a device push to deploy the firmware physically.</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">4</span> <div><b style="color: #409eff;">Verify Installation</b><br><span style="color: #409eff;">Check the Device Images tab to confirm success block availability.</span></div></li>
      <li style="margin-bottom: 12px; display: flex;"><span style="color: #409eff; font-size: 16px; margin-right: 10px; font-weight: bold; display: inline-block; width: 20px; text-align: center; border: 1px solid #409eff; border-radius: 50%; line-height: 20px;">5</span> <div><b style="color: #409eff;">Boot Instance</b><br><span style="color: #409eff;">Deploy a new active cloud machine leveraging this image.</span></div></li>
    </ol>
  </div>

  <!-- 3. Simulator vs Container -->
  <div style="font-weight: bold; margin-bottom: 12px; font-size: 14px;">Simulator vs Container: Differences</div>
  <table style="width: 100%; border-collapse: collapse; margin-bottom: 24px; text-align: left; font-size: 12px;">
    <thead>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <th style="padding: 10px 0; width: 15%; color:#606266;">Comparison</th>
        <th style="padding: 10px 0; width: 42%; color:#409eff;">🖥️ Simulator</th>
        <th style="padding: 10px 0; width: 43%; color:#e6a23c;">📦 Container</th>
      </tr>
    </thead>
    <tbody>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Architecture</td>
        <td style="padding: 8px 0;">VM imitating ARM hardware (QEMU)</td>
        <td style="padding: 8px 0;">Lightweight Linux orchestration</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Boot Speed</td>
        <td style="padding: 8px 0;">Slower, typically 50+ secs</td>
        <td style="padding: 8px 0;">Extremely fast, usually 3~10 secs</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Resource Use</td>
        <td style="padding: 8px 0;">Higher, requires isolated CPU blocks</td>
        <td style="padding: 8px 0;">Extremely Low, multiple instances share host pools</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Density</td>
        <td style="padding: 8px 0;">Strict limits; few deployments</td>
        <td style="padding: 8px 0;">Unlimited soft barriers; tens to hundreds</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Compatibility</td>
        <td style="padding: 8px 0;">Behaves close to true hardware devices</td>
        <td style="padding: 8px 0;">Dependence on host OS bounds hooks</td>
      </tr>
      <tr style="border-bottom: 1px dashed #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Graphics</td>
        <td style="padding: 8px 0;">Fluid GPU rendering layers</td>
        <td style="padding: 8px 0;">Softer graphics bounds</td>
      </tr>
      <tr style="border-bottom: 1px solid #ebeef5;">
        <td style="padding: 8px 0; color:#909399;">Use Cases</td>
        <td style="padding: 8px 0;">UI tests, Gaming, Heavy graphics</td>
        <td style="padding: 8px 0;">Automation scripts, High density deployments</td>
      </tr>
    </tbody>
  </table>

  <!-- 4. Suggestions -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #e6a23c;">
    💡 Selection Advice
  </div>
  <div style="background-color: #f0f9eb; padding: 12px; border-radius: 4px; border-left: 4px solid #67c23a; margin-bottom: 8px;">
    <div style="display: flex; align-items: flex-start;">
      <i class="el-icon-success" style="color: #67c23a; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
      <div>
        <div style="font-weight: bold; color: #67c23a; margin-bottom: 4px;">Recommended Container usage:</div>
        <div style="color: #67c23a;">When requiring enormous scalability across parallel scripts and lightweight headless setups. Rapid spin ups.</div>
      </div>
    </div>
  </div>
  <div style="background-color: #fdf6ec; padding: 12px; border-radius: 4px; border-left: 4px solid #e6a23c; margin-bottom: 24px;">
    <div style="display: flex; align-items: flex-start;">
       <i class="el-icon-warning" style="color: #e6a23c; font-size: 18px; margin-right: 10px; margin-top: 2px;"></i>
       <div>
         <div style="font-weight: bold; color: #e6a23c; margin-bottom: 4px;">Recommended Simulator usage:</div>
         <div style="color: #e6a23c;">Deep graphical dependencies, high accuracy hardware modeling testcases, Android core behavior bounds.</div>
       </div>
    </div>
  </div>

  <!-- 5. Warnings -->
  <div style="font-weight: bold; margin-bottom: 8px; font-size: 14px; display: flex; align-items: center; color: #F56C6C;">
    ⚠️ Important Notes
  </div>
  <ul style="padding-left: 20px; color: #606266; margin-bottom: 0;">
    <li>Firmware is generally 1GB+; ensure host disk capacity permits holding payloads.</li>
    <li>Networks must remain connected uninterrupted while pumping huge files to device endpoints.</li>
    <li>Purging an image active machines rely upon causes unrecoverable crash states. Audit environments cautiously.</li>
    <li><b>P1 series devices</b> do not support Android 12 container variants. Choose A10 or A14.</li>
  </ul>
</div>`;

// Replace guideContent in main.js
// There are two blocks of image: { ... guideContent: `...` }
mainText = mainText.replace(/guideContent:\s*`[\s\S]*?`\s*(\n\s*\})/g, function(match, tail, offset, string) {
  const before = string.slice(Math.max(0, offset - 100), offset);
  if (before.includes("usageGuide: 'ʹ˵'") || before.includes("usageGuide: '使用说明'")) {
    return 'guideContent: `' + zhGuideContent + '`' + tail;
  } else {
    return 'guideContent: `' + enGuideContent + '`' + tail;
  }
});

fs.writeFileSync(mainJsPath, mainText);
console.log('SUCCESS: Restored image guide content.');
