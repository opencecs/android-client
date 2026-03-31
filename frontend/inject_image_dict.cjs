const fs = require('fs');
const path = require('path');

const mainJsPath = path.join(__dirname, 'src', 'main.js');
let mainJsContent = fs.readFileSync(mainJsPath, 'utf8');

const zhGuideContent = `
        <h3 class="guide-title">
          <span style="vertical-align: middle; margin-right: 6px;">ℹ️</span>镜像管理使用说明
        </h3>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">📋 功能介绍</h4>
        <p class="guide-text">镜像管理模块用于管理运行在设备上的系统镜像，支持从云端下载镜像到本地、将本地镜像上传到设备，以及查看和删除设备上已有的镜像。主要包含三个功能页签：</p>
        <ul class="guide-list">
          <li><strong>在线镜像</strong>：浏览可用镜像列表，可按设备型号等筛选，支持一键下载到本地或上传到设备。</li>
          <li><strong>本地镜像</strong>：查看已下载到本地的镜像文件，可将镜像上传至设备，也可删除本地缓存以释放磁盘。</li>
          <li><strong>设备镜像</strong>：选择某台设备，查看该设备上已安装的镜像列表，并进行删除操作。</li>
        </ul>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">🔄 典型使用流程</h4>
        <ol class="guide-list" style="list-style-type: decimal;">
          <li><strong>选择在线镜像</strong>：在「在线镜像」中，找到目标镜像。</li>
          <li><strong>下载到本地</strong>：点击「下载」按钮，将文件缓存到本地目录。</li>
          <li><strong>上传到设备</strong>：点击「上传到设备」，选择目标设备完成推送。</li>
          <li><strong>验证及使用</strong>：在「设备镜像」中检查，并在「云机」页面使用已上传的镜像。</li>
        </ol>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">⚠️ 注意事项</h4>
        <ul class="guide-list">
          <li>镜像文件通常体积较大，下载前请确保磁盘空间充足。</li>
          <li>上传镜像到设备需要设备在线且网络稳定。</li>
          <li>部分型号不兼容特定版本的容器镜像，请注意设备型号限制。</li>
        </ul>
`;

const enGuideContent = `
        <h3 class="guide-title">
          <span style="vertical-align: middle; margin-right: 6px;">ℹ️</span>Image Management Guide
        </h3>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">📋 Module Features</h4>
        <p class="guide-text">The Image Management module manages system images across your devices, supporting downloads from the cloud and direct pushes to devices. It consists of three main tabs:</p>
        <ul class="guide-list">
          <li><strong>Online Image</strong>: Browse available cloud images, filter by device parameters, and download or push them.</li>
          <li><strong>Local Image</strong>: View your locally cached images on disk. You can push them to devices or delete them to free up space.</li>
          <li><strong>Device Image</strong>: Inspect installed images on an individual device and delete them if no longer needed.</li>
        </ul>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">🔄 Typical Workflow</h4>
        <ol class="guide-list" style="list-style-type: decimal;">
          <li><strong>Find Image</strong>: Locate the desired image under 'Online Image'.</li>
          <li><strong>Download</strong>: Cache the image to your local directory by clicking 'Download'.</li>
          <li><strong>Upload</strong>: Push the image directly to your connected devices.</li>
          <li><strong>Verify/Use</strong>: Verify under 'Device Image' and deploy new cloud machines using this image.</li>
        </ol>
        <hr style="border: none; border-top: 1px solid #EBEEF5; margin: 16px 0;"/>
        <h4 class="guide-section-title">⚠️ Important Notes</h4>
        <ul class="guide-list">
          <li>Images are generally large; make sure your disk has enough free space.</li>
          <li>A stable network connection is required when pushing images to devices.</li>
          <li>Certain device models may not be compatible with all container OS versions.</li>
        </ul>
`;

let insertCount = 0;
mainJsContent = mainJsContent.replace(
  /    addDevice: \{/g,
  function(match) {
    insertCount++;
    if (insertCount === 1) {
      // First one is zh-CN
      return "    image: {\n      usageGuide: '使用说明',\n      guideContent: `" + zhGuideContent + "`\n    },\n    addDevice: {";
    } else if (insertCount === 2) {
      // Second one is en-US
      return "    image: {\n      usageGuide: 'Usage Guide',\n      guideContent: `" + enGuideContent + "`\n    },\n    addDevice: {";
    }
    return match;
  }
);

fs.writeFileSync(mainJsPath, mainJsContent);
console.log('SUCCESS: Injected image usage dictionary');
