const fs = require('fs');
const path = require('path');

const vuePath = path.join(__dirname, 'src', 'components', 'StreamManagement.vue');
let content = fs.readFileSync(vuePath, 'utf8');

if (!content.includes('const { t } = useI18n()') && !content.includes('getCurrentInstance()')) {
  // Simple check... we will use $t in template so proxy doesn't matter if it's only in template
}

content = content.replace(/<span>流管理 \(Streams\)<\/span>/g, `<span>{{ $t('stream.streamManagementTitle') }}</span>`);
content = content.replace(/label="转发流\(OBS\)"/g, `:label="$t('stream.forwardStream')"`);
content = content.replace(/label="点对点\(OBS\)"/g, `:label="$t('stream.p2pStream')"`);

content = content.replace(/placeholder="输入房间号"/g, `:placeholder="$t('stream.inputRoomId')"`);
content = content.replace(/>\s*刷新列表\s*<\/el-button>/g, `>{{ $t('stream.refreshList') }}</el-button>`);

content = content.replace(/<th>推流码\/房间<\/th>/g, `<th>{{ $t('stream.pushCodeRoom') }}</th>`);
content = content.replace(/<th>状态<\/th>/g, `<th>{{ $t('common.statusLabel') }}</th>`);
content = content.replace(/<th>源IP<\/th>/g, `<th>{{ $t('stream.sourceIp') }}</th>`);
content = content.replace(/<th>码率 \(Video\/Audio\)<\/th>/g, `<th>{{ $t('stream.bitrate') }}</th>`);
content = content.replace(/<th>操作<\/th>/g, `<th>{{ $t('common.operation') }}</th>`);
content = content.replace(/暂无流信息，请点击上方新建/g, `{{ $t('stream.noStreamInfo') }}`);
content = content.replace(/<\/i>\s*分发\s*<\/el-button>/g, `></i> {{ $t('stream.distribute') }}</el-button>`);
content = content.replace(/<\/i>\s*断开\s*<\/el-button>/g, `></i> {{ $t('stream.disconnect') }}</el-button>`);
content = content.replace(/<\/i>\s*删除\s*<\/el-button>/g, `></i> {{ $t('common.delete') }}</el-button>`);

content = content.replace(/<span>设备-云机关系 \(Device & Cloud Machines\)<\/span>/g, `<span>{{ $t('stream.deviceCloudRelation') }}</span>`);
content = content.replace(/>\s*分发到云机\s*<\/el-button>/g, `>{{ $t('stream.distributeToCloud') }}</el-button>`);

content = content.replace(/已配置分发 \(Active\)/g, `{{ $t('stream.activeDistribution') }}`);
content = content.replace(/>\s*刷新状态\s*<\/el-button>/g, `>{{ $t('stream.refreshStatus') }}</el-button>`);
content = content.replace(/<th>推流地址<\/th>/g, `<th>{{ $t('stream.pushAddress') }}</th>`);
content = content.replace(/<th>设备IP<\/th>/g, `<th>{{ $t('backup.deviceIP') }}</th>`);
content = content.replace(/<th>云机<\/th>/g, `<th>{{ $t('common.cloudMachine') }}</th>`);
content = content.replace(/<th>协议<\/th>/g, `<th>{{ $t('network.protocol') }}</th>`);
content = content.replace(/暂无\{\{ \$t\('common\.activeStatus'\) \}\}分发/g, `{{ $t('stream.noActiveDistribution') }}`);
content = content.replace(/暂无common\.activeStatus分发/g, `{{ $t('stream.noActiveDistribution') }}`);

content = content.replace(/失效分发 \(Inactive\/Unverified\)/g, `{{ $t('stream.inactiveDistribution') }}`);
content = content.replace(/<th>状态信息<\/th>/g, `<th>{{ $t('common.statusInfo') }}</th>`);
content = content.replace(/设备离线或不存在/g, `{{ $t('stream.deviceOfflineOrNotExist') }}`);
content = content.replace(/>\s*启动P2P\s*<\/el-button>/g, `>{{ $t('stream.startP2P') }}</el-button>`);
content = content.replace(/>\s*投屏\s*<\/el-button>/g, `>{{ $t('common.projection') }}</el-button>`);

content = content.replace(/common\.webrtcAddress/g, `{{ $t('stream.webrtcAddress') }}`);
content = content.replace(/common\.rtmpAddress/g, `{{ $t('stream.rtmpAddress') }}`);

fs.writeFileSync(vuePath, content);
console.log('SUCCESS: Updated StreamManagement.vue');

// ---------------- Update main.js ----------------
const mainPath = path.join(__dirname, 'src', 'main.js');
let mainText = fs.readFileSync(mainPath, 'utf8');

const zhStreamDict = `
    stream: {
      streamManagementTitle: '流管理 (Streams)',
      forwardStream: '转发流(OBS)',
      p2pStream: '点对点(OBS)',
      inputRoomId: '输入房间号',
      refreshList: '刷新列表',
      pushCodeRoom: '推流码/房间',
      sourceIp: '源IP',
      bitrate: '码率 (Video/Audio)',
      noStreamInfo: '暂无流信息，请点击上方新建',
      distribute: '分发',
      disconnect: '断开',
      deviceCloudRelation: '设备-云机关系 (Device & Cloud Machines)',
      distributeToCloud: '分发到云机',
      activeDistribution: '已配置分发 (Active)',
      refreshStatus: '刷新状态',
      pushAddress: '推流地址',
      noActiveDistribution: '暂无分发流信息',
      inactiveDistribution: '失效分发 (Inactive/Unverified)',
      deviceOfflineOrNotExist: '设备离线或不存在',
      startP2P: '启动P2P',
      webrtcAddress: 'WEBRTC地址',
      rtmpAddress: 'RTMP地址'
    },`;

const enStreamDict = `
    stream: {
      streamManagementTitle: 'Stream Management',
      forwardStream: 'Forward Stream (OBS)',
      p2pStream: 'Peer-to-Peer (OBS)',
      inputRoomId: 'Enter Room ID',
      refreshList: 'Refresh List',
      pushCodeRoom: 'Stream Key/Room',
      sourceIp: 'Source IP',
      bitrate: 'Bitrate (Video/Audio)',
      noStreamInfo: 'No stream info, click New Stream to create',
      distribute: 'Distribute',
      disconnect: 'Disconnect',
      deviceCloudRelation: 'Device & Cloud Machines',
      distributeToCloud: 'Distribute to Machine',
      activeDistribution: 'Active Distribution',
      refreshStatus: 'Refresh Status',
      pushAddress: 'Push URL',
      noActiveDistribution: 'No Active Distribution',
      inactiveDistribution: 'Inactive/Unverified Distribution',
      deviceOfflineOrNotExist: 'Device Offline or Not Found',
      startP2P: 'Start P2P',
      webrtcAddress: 'WEBRTC URL',
      rtmpAddress: 'RTMP URL'
    },`;


let insertCount = 0;
mainText = mainText.replace(
  /\n\s*instance: \{/g,
  function(match) {
    insertCount++;
    if (insertCount === 1) {
      return '\n' + zhStreamDict + match;
    } else if (insertCount === 2) {
      return '\n' + enStreamDict + match;
    }
    return match;
  }
);

mainText = mainText.replace(/common: \{/g, "common: {\n      cloudMachine: '云机',\n      statusInfo: '状态信息',\n      statusLabel: '状态',\n      projection: '投屏',");
mainText = mainText.replace(/common: \{\s*cloudMachine: '云机'/g, "common: {\n      cloudMachine: 'Machine'");

fs.writeFileSync(mainPath, mainText);
console.log('SUCCESS: Updated main.js for stream');
