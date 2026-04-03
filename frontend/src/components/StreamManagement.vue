<template>
  <div class="stream-management">
    <!-- 顶部状态栏
    <el-card class="status-card">
      <div class="server-info">
        <h3>{{ $t('common.streamServer') }}</h3>
        <p>{{ $t('common.localStreamStarted') }}</p>
        <ul style="font-size: 12px; color: #666; margin: 5px 0; padding-left: 20px;">
          <li><strong>HTTP-FLV (TCP):</strong> 8083 - {{ $t('common.normal') }}</li>
          <li><strong>RTMP (TCP):</strong> 1935 - {{ $t('common.normal') }}</li>
        </ul>
        <p style="margin-top: 10px;">{{ $t('common.createRoomDesc') }}</p>
      </div>
    </el-card> -->

    <div class="main-content">
      <!-- 左侧：推流列表 -->
      <el-card class="stream-list-card" :body-style="{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', height: '100%' }">
        <template #header>
          <div class="card-header">
            <span>{{ $t('stream.streamManagementTitle') }}</span>
          </div>
        </template>
        
        <!-- 标签页 -->
        <el-tabs v-model="activeStreamTab" style="flex: 1; min-height: 0; display: flex; flex-direction: column;">
          <!-- 转发流标签页 -->
          <el-tab-pane :label="$t('stream.forwardStream')" name="forward" style="height: 100%; display: flex; flex-direction: column;">
            <div class="header-actions" style="margin-bottom: 10px;">
              <el-input 
                v-model="newStreamName" 
                :placeholder="$t('stream.inputRoomId')" 
                size="small" 
                style="width: 150px; margin-right: 10px;"
                @keyup.enter="addNewStream"
              />
              <el-button type="success" size="small" @click="addNewStream">{{ $t('common.newStream') }}</el-button>
              <el-button type="primary" size="small" @click="refreshStreams" style="margin-left: 10px;">{{ $t('stream.refreshList') }}</el-button>
            </div>
            <div style="padding: 12px; background: #f5f7fa; border-bottom: 1px solid #ebeef5; font-size: 13px; color: #606266;">
              <div>
                <strong>{{ $t('stream.webrtcAddress') }}</strong> 
                <span 
                  style="cursor: pointer; color: #409eff; text-decoration: underline;" 
                  @click="copyToClipboard(`webrtc://${localIp}/live`)"
                  :title="$t('common.clickToCopy')"
                >
                  webrtc://{{ localIp }}/live
                </span>
              </div>
              <div style="margin-top: 4px;">
                <strong>{{ $t('stream.rtmpAddress') }}</strong> 
                <span 
                  style="cursor: pointer; color: #409eff; text-decoration: underline;" 
                  @click="copyToClipboard(`rtmp://${localIp}:1935/live`)"
                  :title="$t('common.clickToCopy')"
                >
                  rtmp://{{ localIp }}:1935/live
                </span>
              </div>
            </div>
            <div class="native-table-container">
              <table class="native-table">
                <thead>
                  <tr>
                    <th>{{ $t('stream.pushCodeRoom') }}</th>
                    <th>{{ $t('common.statusLabel') }}</th>
                    <th>{{ $t('stream.sourceIp') }}</th>
                    <th>{{ $t('stream.bitrate') }}</th>
                    <th>{{ $t('common.operation') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="activeStreams.length === 0">
                    <td colspan="5" style="text-align: center; color: #909399; padding: 20px;">{{ $t('stream.noStreamInfo') }}</td>
                  </tr>
                  <tr v-for="(stream, index) in activeStreams" :key="index">
                    <td>
                      <div style="font-size: 13px; color: #303133; font-weight: 500;">{{ stream.streamName }}</div>
                    </td>
                    <td>
                      <el-tag size="small" :type="stream.publisherIP ? 'success' : 'info'">
                        {{ stream.publisherIP ? $t('stream.activeStatus') + ' (Active)' : $t('stream.idleStatus') + ' (Idle)' }}
                      </el-tag>
                    </td>
                    <td>{{ stream.publisherIP || '-' }}</td>
                    <td>
                      <span v-if="stream.publisherIP">
                        {{ (stream.videoBitrate / 1024).toFixed(1) }} / {{ (stream.audioBitrate / 1024).toFixed(1) }} kbps
                      </span>
                      <span v-else>-</span>
                    </td>
                    <td>
                      <div class="action-buttons">
                        <el-button size="small" type="primary" :disabled="!stream.publisherIP" @click="openDistributeDialog(stream)">
                          <i class="el-icon-share"></i> {{ $t('stream.distribute') }}</el-button>
                        <el-button size="small" type="warning" :disabled="!stream.publisherIP" @click="handleStopPush(stream)">
                          <i class="el-icon-video-pause"></i> {{ $t('stream.disconnect') }}</el-button>
                        <el-button size="small" type="danger" plain @click="handleDeleteStream(stream)">
                          <i class="el-icon-delete"></i> {{ $t('common.delete') }}</el-button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </el-tab-pane>

          <!-- 点对点标签页 -->
          <el-tab-pane :label="$t('stream.p2pStream')" name="p2p" style="height: 100%; display: flex; flex-direction: column;">

            <div class="header-actions" style="margin-bottom: 10px;">
              <el-button type="warning" size="small" @click="openP2PDialog(null)">
                添加P2P流
              </el-button>
            </div>
            
            <div class="native-table-container">
              <table class="native-table">
                <thead>
                  <tr>
                    <th>{{ $t('stream.pushAddress') }}</th>
                    <th>{{ $t('backup.deviceIP') }}</th>
                    <th class="cloud-machine-col">云机</th>
                    <th>监听地址</th>
                    <th>{{ $t('common.statusLabel') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="p2pStreams.length === 0">
                    <td colspan="5" style="text-align: center; color: #909399; padding: 20px;">暂无P2P流信息，请点击上方添加</td>
                  </tr>
                  <tr v-for="item in p2pStreams" :key="item.id">
                    <td>{{ item.streamName }}</td>
                    <td>{{ item.deviceIp }}</td>
                    <td class="cloud-machine-col">
                      <div class="cloud-machine-name" :title="item.cloudMachineName || item.cloudMachineId">
                        {{ formatInstanceName(item.cloudMachineName || item.cloudMachineId) }}
                      </div>
                    </td>
                    <td>
                      <span 
                        style="cursor: pointer; color: #409eff; text-decoration: underline;" 
                        @click="copyToClipboard(getP2PListenUrl(item))"
                        :title="$t('common.clickToCopy')"
                      >
                        {{ getP2PListenUrl(item) }}
                      </span>
                    </td>
                    <td>
                      <el-button v-if="item.status !== 'active'" type="primary" size="small" @click="startP2PDistribution(item)">{{ $t('stream.startP2P') }}</el-button>
                      <el-button v-else type="warning" size="small" @click="stopP2PDistribution(item)">停止P2P</el-button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </el-tab-pane>

          <!-- 摄像头模式标签页 -->
          <el-tab-pane :label="$t('common.pcCamera')" name="camera" style="height: 100%; display: flex; flex-direction: column;">
            <div class="header-actions" style="margin-bottom: 10px;">
              <el-button type="primary" size="small" @click="openCameraDialog">
                {{ $t('common.addCamera') }}
              </el-button>
            </div>

            <!-- 推流列表 -->
            <div class="native-table-container" style="flex:1; overflow:auto;">
              <table class="native-table">
                <thead>
                  <tr>
                    <th>{{ $t('backup.deviceIP') }}</th>
                    <th class="cloud-machine-col">云机</th>
                    <th class="cloud-machine-col">摄像头</th>
                    <th>分辨率/码率</th>
                    <th>{{ $t('common.statusLabel') }}</th>
                    <th>{{ $t('common.operation') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="cameraStreams.length === 0">
                    <td colspan="6" style="text-align: center; color: #909399; padding: 20px;">暂无摄像头推流，请点击上方添加</td>
                  </tr>
                  <tr v-for="item in cameraStreams" :key="item.id"
                      :style="camPreviewActiveId === item.id ? 'background:#ecf5ff;' : ''"
                      style="cursor:default;">
                    <td>{{ item.deviceIp }}</td>
                    <td class="cloud-machine-col">
                      <div class="cloud-machine-name" :title="item.cloudMachineName || item.cloudMachineId">
                        {{ formatInstanceName(item.cloudMachineName || item.cloudMachineId) }}
                      </div>
                    </td>
                    <td style="font-size:12px; color:#606266;">{{ item.camName || '默认摄像头' }}</td>
                    <td style="font-size:12px;">{{ item.width }}x{{ item.height }}@{{ item.fps }}fps / {{ item.bitrate }}kbps</td>
                    <td>
                      <el-tag size="small" :type="item.status === 'running' ? 'success' : 'info'">
                        {{ item.status === 'running' ? '推流中' : '已停止' }}
                      </el-tag>
                    </td>
                    <td>
                      <div class="action-buttons">
                        <el-button v-if="item.status !== 'running'" type="primary" size="small" @click="startCameraStream(item)" :loading="item.loading">启动</el-button>
                        <el-button v-else type="warning" size="small" @click="stopCameraStream(item)" :loading="item.loading">停止</el-button>
                        <el-button type="danger" size="small" plain @click="removeCameraStream(item)">删除</el-button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </el-tab-pane>

          <!-- 使用说明标签页 -->
          <el-tab-pane :label="$t('common.userGuide')" name="guide" style="height: 100%; display: flex; flex-direction: column; overflow: hidden; min-height: 0;">
            <div class="guide-container" style="flex:1; min-height:0; overflow-y:auto; padding:16px 20px; font-size:13px; color:#303133; line-height:1.8;">
              <div v-html="$t('stream.streamGuideHtml')"></div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-card>

      <el-card class="client-list-card" :body-style="{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', height: '100%' }">
        <template #header>
          <div class="card-header">
            <span style="font-weight: bold;">
              {{ activeStreamTab === 'camera' ? '摄像头实时预览' : '设备-云机关系 (Device & Cloud Machines)' }}
            </span>
            <div class="header-actions" v-if="activeStreamTab !== 'camera'">
              <el-button type="primary" size="small" @click="openDistributeDialog(null)">{{ $t('stream.distributeToCloud') }}</el-button>
            </div>
          </div>
        </template>

        <!-- 摄像头模式：只显示预览 -->
        <div v-if="activeStreamTab === 'camera'"
             style="flex:1; display:flex; flex-direction:column; padding:12px; box-sizing:border-box; overflow:hidden;">
          <!-- 顶部提示行 -->
          <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:10px; flex-shrink:0;">
            <span style="font-size:12px; color:#909399;">
              <el-icon style="vertical-align:middle; margin-right:2px;"><VideoCamera /></el-icon>
              {{ camPreviewActiveId ? `自动刷新中（每 ${camPreviewInterval}s 抓帧）` : '启动推流后自动预览' }}
            </span>
          </div>
          <!-- 加载中（无旧图时） -->
          <div v-if="camPreviewLoading && !camPreviewDataURL"
               style="flex:1; display:flex; align-items:center; justify-content:center; background:#f0f0f0; border-radius:6px; color:#909399; font-size:13px;">
            <el-icon class="is-loading" style="margin-right:6px;"><Loading /></el-icon>
            抓取画面中…
          </div>
          <!-- 有图：叠加半透明遮罩刷新，不闪烁 -->
          <div v-else-if="camPreviewDataURL"
               style="flex:1; position:relative; display:flex; align-items:center; justify-content:center; overflow:hidden; border-radius:6px; background:#000;">
            <img :src="camPreviewDataURL"
                 style="max-width:100%; max-height:100%; object-fit:contain; display:block; border-radius:6px;"
                 :alt="$t('common.preview')" />
            <div v-if="camPreviewLoading"
                 style="position:absolute; inset:0; display:flex; align-items:center; justify-content:center; background:rgba(0,0,0,0.3); border-radius:6px; color:#fff; font-size:13px;">
              <el-icon class="is-loading" style="margin-right:6px;"><Loading /></el-icon>
              刷新中…
            </div>
          </div>
          <!-- 空状态 -->
          <div v-else
               style="flex:1; display:flex; flex-direction:column; align-items:center; justify-content:center; background:#f5f7fa; border-radius:6px; color:#c0c4cc; font-size:13px; gap:10px;">
            <el-icon style="font-size:48px;"><VideoCamera /></el-icon>
            <span>在左侧「摄像头模式」启动推流后自动预览</span>
          </div>
        </div>

        <!-- 非摄像头模式：原有分发内容 -->
        <div v-else class="device-cloud-panel">
          <!-- 已配置分发 -->
          <div style="flex: 1; display: flex; flex-direction: column; min-height: 0; border-bottom: 1px solid #eee; padding-bottom: 10px;">
            <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;">
              <div style="font-weight: bold; color: #67c23a;">
                <i class="el-icon-success"></i> {{ $t('stream.activeDistribution') }}
              </div>
              <el-button 
                type="text" 
                size="small" 
                :loading="refreshingDistributions" 
                :disabled="refreshingDistributions"
                @click="checkDistributionsStatus"
              >
                <i class="el-icon-refresh"></i>{{ $t('stream.refreshStatus') }}</el-button>
            </div>
            <div class="native-table-container">
              <table class="native-table">
                <thead>
                  <tr>
                    <th>{{ $t('stream.pushAddress') }}</th>
                    <th>{{ $t('backup.deviceIP') }}</th>
                    <th class="cloud-machine-col">云机</th>
                    <th>{{ $t('network.protocol') }}</th>
                    <th>{{ $t('common.operation') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="activeDistributions.length === 0">
                    <td colspan="5" style="text-align: center; color: #909399; padding: 20px;">
                      {{ refreshingDistributions ? $t('stream.queryingStatus') : $t('stream.noActiveDistribution') }}
                    </td>
                  </tr>
                  <tr v-for="item in activeDistributions" :key="item.id">
                    <td>{{ item.streamName }}</td>
                    <td>{{ item.deviceIp }}</td>
                    <td class="cloud-machine-col">
                      <div class="cloud-machine-name" :title="item.cloudMachineName || item.cloudMachineId">
                        {{ formatInstanceName(item.cloudMachineName || item.cloudMachineId) }}
                      </div>
                    </td>
                    <td>{{ item.protocol || 'httpflv' }}</td>
                    <td>
                      <div class="action-buttons">
                        <el-button v-if="item.protocol !== 'p2p'" type="primary" size="small" plain @click="openEditDistribution(item)">
                          <i class="el-icon-edit"></i> 修改流
                        </el-button>
                        <el-button v-if="item.protocol === 'p2p' && item.status !== 'active'" type="success" size="small" @click="startP2PDistribution(item)">
                          <i class="el-icon-video-play"></i>{{ $t('stream.startP2P') }}</el-button>
                        <el-button v-if="item.protocol === 'p2p' && item.status === 'active'" type="warning" size="small" @click="stopP2PDistribution(item)">
                          <i class="el-icon-video-pause"></i> 停止P2P
                        </el-button>
                        <el-button type="success" size="small" @click="openProjectionForDistribution(item)">
                          <i class="el-icon-monitor"></i>{{ $t('common.projection') }}</el-button>
                        <el-button type="danger" size="small" plain @click="removeDistribution(item)">
                          <i class="el-icon-delete"></i> {{ $t('common.delete') }}</el-button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 失效分发 -->
          <div style="flex: 1; display: flex; flex-direction: column; min-height: 0; margin-top: 10px;">
            <div style="font-weight: bold; color: #909399; margin-bottom: 8px;">
              <i class="el-icon-warning"></i> {{ $t('stream.inactiveDistribution') }}
            </div>
            <div class="native-table-container">
              <table class="native-table">
                <thead>
                  <tr>
                    <th>{{ $t('stream.pushAddress') }}</th>
                    <th>{{ $t('backup.deviceIP') }}</th>
                    <th class="cloud-machine-col">云机</th>
                    <th>{{ $t('network.protocol') }}</th>
                    <th>{{ $t('common.statusInfo') }}</th>
                    <th>{{ $t('common.operation') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="inactiveDistributions.length === 0">
                    <td colspan="6" style="text-align: center; color: #909399; padding: 20px;">暂无失效分发</td>
                  </tr>
                  <tr v-for="item in inactiveDistributions" :key="item.id">
                    <td style="color: #999;">{{ item.streamName }}</td>
                    <td style="color: #999;">{{ item.deviceIp }}</td>
                    <td class="cloud-machine-col" style="color: #999;">
                      <div class="cloud-machine-name" :title="item.cloudMachineName || item.cloudMachineId">
                        {{ formatInstanceName(item.cloudMachineName || item.cloudMachineId) }}
                      </div>
                    </td>
                    <td style="color: #999;">{{ item.protocol || 'httpflv' }}</td>
                    <td>
                      <div class="inactive-status-scroll">
                        <el-tag size="small" type="info">{{ item.statusReason || '未验证' }}</el-tag>
                      </div>
                    </td>
                    <td>
                      <div class="action-buttons">
                        <el-button v-if="item.protocol !== 'p2p'" type="primary" size="small" plain @click="openEditDistribution(item)">
                          <i class="el-icon-edit"></i> 修改流
                        </el-button>
                        <el-button v-if="item.protocol === 'p2p' && item.status !== 'active'" type="success" size="small" @click="startP2PDistribution(item)">
                          <i class="el-icon-video-play"></i>{{ $t('stream.startP2P') }}</el-button>
                        <el-button v-if="item.protocol === 'p2p' && item.status === 'active'" type="warning" size="small" @click="stopP2PDistribution(item)">
                          <i class="el-icon-video-pause"></i> 停止P2P
                        </el-button>
                        <el-button type="success" size="small" @click="openProjectionForDistribution(item)">
                          <i class="el-icon-monitor"></i>{{ $t('common.projection') }}</el-button>
                        <el-button type="danger" size="small" plain @click="removeDistribution(item)">
                          <i class="el-icon-delete"></i> {{ $t('common.delete') }}</el-button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 分发弹窗 -->
    <el-dialog v-model="distributeDialogVisible" title="分发流到云机" width="640px">
      <div class="distribute-form">
        <el-form label-width="80px">
          <el-form-item label="流">
            <el-select v-model="distributeStreamName" filterable allow-create placeholder="选择或输入房间号" style="width: 100%;">
              <el-option v-for="item in streamOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('common.deviceStr')">
            <el-select v-model="distributeDeviceIp" placeholder="选择设备" style="width: 100%;">
              <el-option
                v-for="device in deviceCloudOptions"
                :key="device.ip"
                :label="`${device.ip}${device.name ? ` (${device.name})` : ''}`"
                :value="device.ip"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('common.cloudMachineStr')">
            <el-select v-model="distributeCloudMachineId" placeholder="选择云机" style="width: 100%;" :disabled="!distributeDevice">
              <el-option
                v-for="machine in distributeAvailableCloudMachines"
                :key="machine.id || machine.name"
                :label="machine.name || machine.id"
                :value="machine.id || machine.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="协议">
            <el-select v-model="distributeProtocol" placeholder="选择协议" style="width: 100%;">
              <el-option label="HTTP-FLV" value="httpflv" />
              <el-option label="RTMP" value="rtmp" />
            </el-select>
          </el-form-item>
          <el-form-item label="分辨率">
            <el-select v-model="distributeResolution" placeholder="分辨率" style="width: 100%;">
              <el-option label="自动" value="1" />
              <el-option label="1920x1080@30" value="2" />
              <el-option label="1280x720@30" value="3" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="distributeDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmDistribute" :loading="applyLoading">确认分发</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- P2P 分发弹窗 -->
    <el-dialog v-model="p2pDialogVisible" title="添加P2P流" width="560px">
      <div class="distribute-form">
        <el-form label-width="90px">
          <el-form-item :label="$t('common.deviceStr')">
            <el-select v-model="p2pDeviceIp" placeholder="选择设备" style="width: 100%;">
              <el-option
                v-for="device in deviceCloudOptions"
                :key="device.ip"
                :label="`${device.ip}${device.name ? ` (${device.name})` : ''}`"
                :value="device.ip"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('common.cloudMachineStr')">
            <el-select v-model="p2pCloudMachineId" placeholder="选择云机" style="width: 100%;" :disabled="!p2pDevice">
              <el-option
                v-for="machine in p2pAvailableCloudMachines"
                :key="machine.id || machine.name"
                :label="machine.name || machine.id"
                :value="machine.id || machine.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="P2P端口">
            <div style="display: flex; gap: 6px; align-items: center; width: 100%;">
              <el-input v-model="p2pListenPort" placeholder="例如 9001" type="number" />
              <el-button size="small" @click="decreaseP2PPort">-</el-button>
              <el-button size="small" @click="increaseP2PPort">+</el-button>
            </div>
            <div v-if="p2pPortConflict" style="font-size: 12px; color: #e6a23c; margin-top: 4px;">
              端口已被占用，请选择其他端口
            </div>
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="p2pStreamName" placeholder="可选，默认 p2p-端口" />
          </el-form-item>
          <el-form-item label="分辨率">
            <el-select v-model="p2pResolution" placeholder="分辨率" style="width: 100%;">
              <el-option label="自动" value="1" />
              <el-option label="1920x1080@30" value="2" />
              <el-option label="1280x720@30" value="3" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('common.listenAddr')">
            <div style="font-size: 12px; color: #666;">
              {{ p2pListenUrl || '请填写端口' }}
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="p2pDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmP2PDistribute" :loading="applyLoading">{{ $t('stream.startP2P') }}</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 摄像头推流弹窗 -->
    <el-dialog v-model="cameraDialogVisible" :title="$t('common.addCamera')" width="520px">
      <div class="distribute-form">
        <el-form label-width="90px">
          <el-form-item :label="$t('common.deviceStr')">
            <el-select v-model="camDeviceIp" placeholder="选择设备" style="width: 100%;">
              <el-option
                v-for="device in deviceCloudOptions"
                :key="device.ip"
                :label="`${device.ip}${device.name ? ` (${device.name})` : ''}`"
                :value="device.ip"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('common.cloudMachineStr')">
            <el-select v-model="camCloudMachineId" placeholder="选择云机" style="width: 100%;" :disabled="!camDevice">
              <el-option
                v-for="machine in camAvailableCloudMachines"
                :key="machine.id || machine.name"
                :label="machine.name || machine.id"
                :value="machine.id || machine.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="设备端口">
            <span style="line-height:32px; color:#606266; font-size:13px;">
              {{ camDetectedPort || '—' }}
              <span v-if="!camDetectedPort" style="color:#e6a23c; font-size:12px; margin-left:6px;">（请先选择云机）</span>
            </span>
          </el-form-item>
          <el-form-item label="摄像头">
            <div style="display:flex; gap:8px; width:100%;">
              <el-select
                v-model="camName"
                placeholder="选择摄像头"
                style="flex:1;"
                :loading="camListLoading"
                :disabled="camListLoading"
              >
                <el-option
                  v-for="cam in camDeviceList"
                  :key="cam"
                  :label="cam"
                  :value="cam"
                />
                <el-option v-if="camDeviceList.length === 0 && !camListLoading" label="无" value="" disabled />
              </el-select>
              <el-button size="small" :loading="camListLoading" @click="fetchCameraList">刷新</el-button>
            </div>
          </el-form-item>
          <el-form-item label="分辨率">
            <div style="display:flex; gap:8px;">
              <el-input v-model.number="camWidth" placeholder="宽" type="number" style="width:80px;" />
              <span style="line-height:32px;">x</span>
              <el-input v-model.number="camHeight" placeholder="高" type="number" style="width:80px;" />
            </div>
          </el-form-item>
          <el-form-item label="帧率/码率">
            <div style="display:flex; gap:8px; align-items:center;">
              <el-input v-model.number="camFps" placeholder="FPS" type="number" style="width:80px;" />
              <span style="line-height:32px; color:#999; font-size:12px;">fps</span>
              <el-input v-model.number="camBitrate" placeholder="码率" type="number" style="width:90px;" />
              <span style="line-height:32px; color:#999; font-size:12px;">kbps</span>
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="cameraDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmCameraStream" :loading="applyLoading">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onBeforeUnmount, watch, getCurrentInstance } from 'vue'

const _instance = getCurrentInstance()
const $t = (key, params) => _instance?.proxy?.$t(key, params) || key
const t = $t

import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, VideoCamera } from '@element-plus/icons-vue'
import axios from 'axios'
import { Call } from "@wailsio/runtime"
import { startProjection } from '../services/api.js'
import { 
  GetLocalIp, 
  GetActiveStreams, 
  StopPushSession,
  AddStreamName,
  DeleteStreamName
} from '../../bindings/edgeclient/app'

const props = defineProps({
  devices: { type: Array, default: () => [] },
  devicesStatusCache: { type: Map, default: () => new Map() },
  cloudMachineGroups: { type: Array, default: () => [] },
  fetchAndroidContainers: { type: Function, default: null },
  deviceCloudMachinesCache: { type: Object, default: () => new Map() }
})

const localIp = ref('')
const activeStreams = ref([])
const newStreamName = ref('')
const distributeDialogVisible = ref(false)
const currentStream = ref(null)
const streamDistributions = ref([])
const p2pDialogVisible = ref(false)
const activeStreamTab = ref('forward') // 当前激活的标签页
// 从 LocalStorage 加载
try {
  const saved = localStorage.getItem('streamDistributions')
  if (saved) {
    const parsed = JSON.parse(saved)
    streamDistributions.value = Array.isArray(parsed)
      ? parsed.map(item => ({ ...item, protocol: item?.protocol || 'httpflv' }))
      : []
  }
} catch (e) {
  console.error('Failed to load streamDistributions:', e)
}

// 监听并保存到 LocalStorage
watch(streamDistributions, (val) => {
  localStorage.setItem('streamDistributions', JSON.stringify(val))
}, { deep: true })

const refreshingDistributions = ref(false)
const distributeStreamName = ref('')
const distributeDeviceIp = ref('')
const distributeCloudMachineId = ref('')
const distributeProtocol = ref('httpflv')
const distributeResolution = ref('1')
const editDistributionIndex = ref(-1)
let timer = null
let autoSyncTriggerTimer = null

const p2pDeviceIp = ref('')
const p2pCloudMachineId = ref('')
const p2pListenPort = ref('')
const p2pStreamName = ref('')
const p2pResolution = ref('1')

const selectedDeviceIp = ref('')
const selectedCloudMachineId = ref('')
const sourceType = ref('')
const sourcePath = ref('')
const sourceResolution = ref('1')
const modifydevStatus = ref(null)
const queryLoading = ref(false)
const applyLoading = ref(false)
const autoSyncDone = ref(false)
const autoSyncing = ref(false)

const p2pListenUrl = computed(() => {
  const port = Number(p2pListenPort.value)
  if (!localIp.value || !port) return ''
  return `srt://${localIp.value}:${port}`
})

const getP2PListenUrl = (item) => {
  if (item?.listenUrl) return item.listenUrl
  if (item?.p2pListenPort && localIp.value) {
    return `srt://${localIp.value}:${item.p2pListenPort}`
  }
  return '-'
}

const formatInstanceName = (name) => {
  if (!name) return name
  const match = name.match(/^.+_\d+_(.+)$/)
  if (match && match.length > 1) {
    return match[1]
  }
  return name
}

const activeDistributions = computed(() => {
  return streamDistributions.value.filter(item => item.status === 'active')
})

const inactiveDistributions = computed(() => {
  return streamDistributions.value.filter(item => item.status !== 'active')
})

const p2pStreams = computed(() => {
  return streamDistributions.value.filter(item => item.protocol === 'p2p')
})

const p2pUsedPorts = computed(() => {
  const ports = new Set()
  streamDistributions.value.forEach(item => {
    if (item?.protocol !== 'p2p') return
    if (!item?.p2pListenPort) return
    if (item?.status === 'inactive') return
    ports.add(Number(item.p2pListenPort))
  })
  return ports
})

const p2pPortConflict = computed(() => {
  const port = Number(p2pListenPort.value)
  if (!port) return false
  return p2pUsedPorts.value.has(port)
})

const isStreamActive = (streamName) => {
  if (!streamName) return false
  return activeStreams.value.some(s => (s.streamName || s.StreamName) === streamName && (s.publisherIP || s.PublisherIP))
}

const hasActiveDistributionStream = () => {
  return streamDistributions.value.some(item => isStreamActive(item.streamName))
}

const onlineDevices = computed(() => {
  return props.devices.filter(d => props.devicesStatusCache.get(d.id) === 'online')
})

const hasDeviceCache = (deviceIp) => {
  return !!props.deviceCloudMachinesCache?.has?.(deviceIp)
}

const deviceCloudOptions = computed(() => {
  return onlineDevices.value
    .filter(device => hasDeviceCache(device.ip))
    .map(device => ({
      ...device,
      groupName: device.group || '默认分组',
      cloudMachines: getDeviceCloudMachines(device.ip)
    }))
})

const getDeviceByIp = (deviceIp) => {
  return deviceCloudOptions.value.find(d => d.ip === deviceIp) || props.devices.find(d => d.ip === deviceIp) || null
}

const getDeviceCloudMachines = (deviceIp) => {
  const cached = props.deviceCloudMachinesCache?.get?.(deviceIp)
  return Array.isArray(cached) ? cached : []
}

const ensureCloudMachinesLoaded = async (deviceIp) => {
  const device = getDeviceByIp(deviceIp)
  if (!device) return { device: null, cloudMachines: [] }
  let cloudMachines = getDeviceCloudMachines(deviceIp)
  if (!hasDeviceCache(deviceIp) && typeof props.fetchAndroidContainers === 'function') {
    await props.fetchAndroidContainers(device, true)
    cloudMachines = getDeviceCloudMachines(deviceIp)
  }
  return { device, cloudMachines }
}

const ensureDeviceCacheForDialogs = async () => {
  if (typeof props.fetchAndroidContainers !== 'function') return
  const targets = onlineDevices.value.filter(device => !hasDeviceCache(device.ip))
  if (targets.length === 0) return
  await Promise.allSettled(targets.map(device => props.fetchAndroidContainers(device, true)))
}

const syncDistributionParameters = async (item) => {
  try {
    if (!item || item.status !== 'active') {
      console.info('autoSync:skip:status', { id: item?.id, status: item?.status })
      return { ok: false, reason: 'status' }
    }
    if (!item.deviceIp || !item.cloudMachineId) {
      console.info('autoSync:skip:missing', { id: item?.id, deviceIp: item?.deviceIp, cloudMachineId: item?.cloudMachineId })
      return { ok: false, reason: 'missing' }
    }
    console.info('autoSync:start', { id: item.id, deviceIp: item.deviceIp, cloudMachineId: item.cloudMachineId, protocol: item.protocol })
    const ensured = await ensureCloudMachinesLoaded(item.deviceIp)
    const device = ensured.device
    const cloudMachine = ensured.cloudMachines.find(m => (m.id || m.name) === item.cloudMachineId)
    if (!device || !cloudMachine) {
      console.info('autoSync:skip:cloudMachine', { id: item.id, deviceIp: item.deviceIp, cloudMachineId: item.cloudMachineId })
      return { ok: false, reason: 'cloudMachine' }
    }
    const mappedPort = getMappedPort9082(cloudMachine)
    if (!mappedPort) {
      console.info('autoSync:skip:mappedPort', { id: item.id, deviceIp: item.deviceIp })
      return { ok: false, reason: 'mappedPort' }
    }
    const controlHost = getControlHost(device, cloudMachine)
    const baseUrl = `http://${controlHost}:${mappedPort}`
    if ((item.protocol || 'httpflv') === 'p2p') {
      const resolution = item.resolution || '1'
      const setUrl = `${baseUrl}/modifydev?cmd=4&type=camera&resolution=${resolution}`
      const res = await ProxyHttpGet(setUrl)
      try {
        const startUrl = `${baseUrl}/camera?cmd=start`
        await ProxyHttpGet(startUrl)
      } catch {}
      console.info('autoSync:done', { id: item.id, protocol: 'p2p', code: res?.code })
      return { ok: true }
    }
    const protocol = item.protocol || 'httpflv'
    const streamPath = buildStreamPath(protocol, item.streamName)
    const protocolType = buildProtocolType(protocol)
    const resolution = item.resolution || '1'
    const setUrl = `${baseUrl}/modifydev?cmd=4&type=${protocolType}&path=${encodeURIComponent(streamPath)}&resolution=${resolution}`
    const res = await ProxyHttpGet(setUrl)
    try {
      const startUrl = `${baseUrl}/camera?cmd=start`
      await ProxyHttpGet(startUrl)
    } catch {}
    console.info('autoSync:done', { id: item.id, protocol, code: res?.code })
    return { ok: true }
  } catch (e) {
    console.error('autoSync:error', { id: item?.id, message: e?.message || e })
    return { ok: false, reason: 'error' }
  }
}

const autoSyncActiveDistributions = async () => {
  if (autoSyncDone.value || autoSyncing.value) {
    console.info('autoSync:batch:skip', { done: autoSyncDone.value, syncing: autoSyncing.value })
    return
  }
  const targets = streamDistributions.value.filter(item => item.status === 'active' && isStreamActive(item.streamName))
  if (targets.length === 0) {
    console.info('autoSync:batch:empty')
    return
  }
  autoSyncing.value = true
  try {
    console.info('autoSync:batch:start', { count: targets.length })
    const chunkSize = 4
    for (let i = 0; i < targets.length; i += chunkSize) {
      const chunk = targets.slice(i, i + chunkSize)
      await Promise.allSettled(chunk.map(syncDistributionParameters))
    }
    autoSyncDone.value = true
    console.info('autoSync:batch:done', { count: targets.length })
  } finally {
    autoSyncing.value = false
  }
}

const distributeDevice = computed(() => {
  return deviceCloudOptions.value.find(d => d.ip === distributeDeviceIp.value) || null
})

const distributeAvailableCloudMachines = computed(() => {
  if (!distributeDevice.value?.ip) return []
  return getDeviceCloudMachines(distributeDevice.value.ip).filter(machine => {
    // 只显示运行中的云机
    return machine.status === 'running'
  })
})

const p2pDevice = computed(() => {
  return deviceCloudOptions.value.find(d => d.ip === p2pDeviceIp.value) || null
})

const p2pAvailableCloudMachines = computed(() => {
  if (!p2pDevice.value?.ip) return []
  return getDeviceCloudMachines(p2pDevice.value.ip).filter(machine => {
    // 只显示运行中的云机
    return machine.status === 'running'
  })
})

const streamOptions = computed(() => {
  return activeStreams.value.map(s => s.streamName).filter(Boolean)
})

const selectedDevice = computed(() => {
  return deviceCloudOptions.value.find(d => d.ip === selectedDeviceIp.value) || null
})

const availableCloudMachines = computed(() => {
  return selectedDevice.value?.cloudMachines || []
})

const selectedCloudMachine = computed(() => {
  return availableCloudMachines.value.find(m => (m.id || m.name) === selectedCloudMachineId.value) || null
})

watch(selectedDeviceIp, () => {
  selectedCloudMachineId.value = ''
  modifydevStatus.value = null
})

watch(distributeDeviceIp, () => {
  distributeCloudMachineId.value = ''
})

watch(p2pDeviceIp, () => {
  p2pCloudMachineId.value = ''
})

watch(distributeDialogVisible, (visible) => {
  if (visible) {
    ensureDeviceCacheForDialogs()
  }
})

watch(p2pDialogVisible, (visible) => {
  if (visible) {
    ensureDeviceCacheForDialogs()
  }
})

watch(activeStreams, () => {
  if (autoSyncDone.value || refreshingDistributions.value) return
  if (!streamDistributions.value.length) return
  if (!hasActiveDistributionStream()) return
  if (autoSyncTriggerTimer) clearTimeout(autoSyncTriggerTimer)
  autoSyncTriggerTimer = setTimeout(() => {
    checkDistributionsStatus()
  }, 400)
}, { deep: true })

// 批量查询分发状态
const checkDistributionsStatus = async () => {
  if (refreshingDistributions.value) return
  refreshingDistributions.value = true
  
  // 每次处理10个
  const chunkSize = 10
  const items = [...streamDistributions.value]
  console.info('distributionCheck:start', { count: items.length })
  
  try {
    for (let i = 0; i < items.length; i += chunkSize) {
      const chunk = items.slice(i, i + chunkSize)
      await Promise.all(chunk.map(checkSingleDistributionStatus))
    }
    await autoSyncActiveDistributions()
  } catch (e) {
    console.error('Batch check failed:', e)
  } finally {
    console.info('distributionCheck:done', { count: items.length })
    refreshingDistributions.value = false
  }
}

// 查询单个分发状态
const checkSingleDistributionStatus = async (item) => {
  try {
    // 找到对应的设备和云机对象以获取端口
    let device = deviceCloudOptions.value.find(d => d.ip === item.deviceIp)
    if (!device) {
      item.status = 'unknown'
      item.statusReason = _instance.proxy.$t('stream.deviceOfflineOrNotExist')
      return
    }

    if (item.protocol === 'p2p') {
      let cloudMachine = getDeviceCloudMachines(item.deviceIp).find(m => (m.id || m.name) === item.cloudMachineId)
      if (!cloudMachine) {
        const ensured = await ensureCloudMachinesLoaded(item.deviceIp)
        device = ensured.device || device
        cloudMachine = ensured.cloudMachines.find(m => (m.id || m.name) === item.cloudMachineId)
      }
      if (!cloudMachine) {
        item.status = 'unknown'
        item.statusReason = '云机信息未找到'
        return
      }
      const mappedPort = getMappedPort10006(cloudMachine)
      if (!mappedPort) {
        item.status = 'unknown'
        item.statusReason = '端口未映射'
        return
      }
      const p2pHost = getControlHost(device, cloudMachine) || item.deviceIp
      const res = await Call.ByName('main.App.GetP2PStatus', p2pHost, Number(mappedPort), item.streamName)
      if (res?.running) {
        item.status = 'active'
        item.statusReason = '运行中'
      } else {
        item.status = 'inactive'
        item.statusReason = res?.message || '已停止'
      }
      return
    }
    
    // 从设备云机列表中查找（这里需要设备在线才能获取到最新的云机列表）
    // 如果云机列表未加载，可能无法获取端口。
    // 我们尝试从 item 中获取 cloudMachineId 匹配
    let cloudMachine = getDeviceCloudMachines(item.deviceIp).find(m => (m.id || m.name) === item.cloudMachineId)
    if (!cloudMachine) {
      const ensured = await ensureCloudMachinesLoaded(item.deviceIp)
      device = ensured.device || device
      cloudMachine = ensured.cloudMachines.find(m => (m.id || m.name) === item.cloudMachineId)
    }
    
    // 如果找不到云机对象，无法获取动态端口，尝试使用默认规则或失败
    if (!cloudMachine) {
       // 尝试构建一个临时对象如果 indexNum 可知? 
       // 但 item 里没存 indexNum。如果找不到，只能标记未知。
       item.status = 'unknown'
       item.statusReason = '云机信息未找到'
       return
    }

    const mappedPort = getMappedPort9082(cloudMachine)
    if (!mappedPort) {
      item.status = 'unknown'
      item.statusReason = '端口未映射'
      return
    }

    const controlHost = getControlHost(device, cloudMachine)
    const baseUrl = `http://${controlHost}:${mappedPort}`
    const checkUrl = `${baseUrl}/modifydev?cmd=5`
    const res = await ProxyHttpGet(checkUrl)
    
    if (res?.code === 200) {
      const remotePath = res.path || ''
      const normalizedPath = remotePath.replace(/\\/g, '')
      const protocol = item.protocol || 'httpflv'
      const expectedKeys = []

      if (protocol === 'httpflv') {
        expectedKeys.push(`live/${item.streamName}.flv`, `live/${item.streamName}`)
      } else {
        expectedKeys.push(`live/${item.streamName}`)
      }

      const isMatch = expectedKeys.some(key => normalizedPath.includes(key))

      if (isMatch) {
        item.status = 'active'
        item.statusReason = $t('common.normal')
      } else {
        item.status = 'inactive'
        if (!normalizedPath) {
          item.statusReason = '未配置推流'
        } else {
          item.statusReason = `路径不匹配: ${normalizedPath}`
        }
      }
    } else {
      item.status = 'unknown'
      item.statusReason = `查询失败: ${res?.reason || '未知错误'}`
    }
  } catch (e) {
    item.status = 'unknown'
    item.statusReason = '请求异常'
  }
}

// 首次加载后延时自动刷新一次
watch(() => deviceCloudOptions.value.length, (len) => {
  if (len > 0 && streamDistributions.value.length > 0 && !refreshingDistributions.value) {
    // 稍微延迟确保设备列表完全就绪
    setTimeout(checkDistributionsStatus, 2000)
  }
}, { immediate: true })

onMounted(async () => {
  try {
    localIp.value = await GetLocalIp()
    // 不再默认创建test房间
    // refreshStreams()
    // 定时刷新
    timer = setInterval(refreshStreams, 5000)

    // 尝试在 mounted 时也触发一次（如果 watch 没触发）
    // 注意：deviceCloudOptions 是 computed，可能此时还没数据，所以上面的 watch 更可靠。
    // 但如果数据已经是现成的，watch immediate 应该会处理。
    // 这里加个双重保险
    if (deviceCloudOptions.value.length > 0 && streamDistributions.value.length > 0) {
      setTimeout(checkDistributionsStatus, 2000)
    }
  } catch (e) {
    console.error('Failed to init stream management:', e)
  }
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
  stopPreviewTimer()
})

const refreshStreams = async () => {
  try {
    // 尝试直接使用 Wails Runtime 调用，绕过可能存在的绑定转换问题
    let streams = await Call.ByID(3014324974)
    console.log('Raw Call.ByID result:', streams)

    // 如果直接调用返回空，尝试使用生成的绑定（虽然可能一样）
    if (!streams) {
       streams = await GetActiveStreams()
       console.log('GetActiveStreams result:', streams)
    }
    
    // 强制转换为响应式数组
    if (Array.isArray(streams)) {
      activeStreams.value = streams.map(s => ({
        streamName: s.streamName || s.StreamName,
        publisherIP: s.publisherIP || s.PublisherIP,
        videoBitrate: s.videoBitrate || s.VideoBitrate || 0,
        audioBitrate: s.audioBitrate || s.AudioBitrate || 0
      }))
    } else {
      activeStreams.value = []
    }
    console.log('activeStreams.value:', activeStreams.value)
  } catch (e) {
    console.error('GetActiveStreams error:', e)
  }
}

const addNewStream = async () => {
  if (!newStreamName.value) {
    ElMessage.warning('请输入房间号')
    return
  }
  try {
    // 调用 Call.ByID 或生成的绑定
    // ID: AddStreamName (需要重新生成 bindings 才能有 ID，这里假设先用生成的 JS 函数如果可用，或者用 runtime 调用)
    // 由于我们没有重新生成 bindings ID，这里先用 Call.ByName 如果知道名字，或者直接用生成的函数如果它能工作（但它可能没有 ID）。
    // Wait, generated bindings rely on IDs. Since I added methods in Go but haven't run `wails3 task dev` fully to regenerate everything (it runs in watch mode usually).
    // Let's assume wails3 task dev regenerated the JS. But if not, I might need to check IDs.
    // However, I can use Call.ByName if available? No, Wails 3 uses IDs.
    // Let's rely on the generated `AddStreamName` import. If it fails due to missing ID mapping, I'll need to trigger a full rebuild or check logs.
    // Actually, `wails3 task dev` should handle it.
    
    await AddStreamName(newStreamName.value)
    ElMessage.success(t('common.添加成功'))
    newStreamName.value = ''
    refreshStreams()
  } catch (e) {
    ElMessage.error('添加失败: ' + e)
  }
}

const handleDeleteStream = async (stream) => {
  try {
    await ElMessageBox.confirm(`确定要删除流 "${stream.streamName}" 吗？`, '提示', { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') })
    await DeleteStreamName(stream.streamName)
    ElMessage.success(t('common.删除成功'))
    refreshStreams()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败: ' + e)
    }
  }
}

const openDistributeDialog = (stream) => {
  currentStream.value = stream
  distributeStreamName.value = stream?.streamName || ''
  distributeDeviceIp.value = ''
  distributeCloudMachineId.value = ''
  distributeProtocol.value = 'httpflv'
  distributeResolution.value = '1'
  editDistributionIndex.value = -1
  distributeDialogVisible.value = true
}

const openP2PDialog = () => {
  p2pDeviceIp.value = ''
  p2pCloudMachineId.value = ''
  p2pListenPort.value = String(getNextAvailableP2PPort(9000))
  p2pStreamName.value = ''
  p2pResolution.value = '1'
  p2pDialogVisible.value = true
}

const getNextAvailableP2PPort = (startPort) => {
  let port = Number(startPort) || 9000
  while (p2pUsedPorts.value.has(port)) {
    port += 1
  }
  return port
}

const increaseP2PPort = () => {
  const current = Number(p2pListenPort.value) || 9000
  const next = getNextAvailableP2PPort(current + 1)
  p2pListenPort.value = String(next)
}

const decreaseP2PPort = () => {
  const current = Number(p2pListenPort.value) || 9000
  let port = current - 1
  if (port < 1) port = 1
  while (port > 0 && p2pUsedPorts.value.has(port)) {
    port -= 1
  }
  if (port <= 0) {
    port = getNextAvailableP2PPort(9000)
  }
  p2pListenPort.value = String(port)
}

const openEditDistribution = (row) => {
  const index = streamDistributions.value.findIndex(item => item.id === row.id)
  if (index === -1) return
  editDistributionIndex.value = index
  distributeStreamName.value = row.streamName
  distributeDeviceIp.value = row.deviceIp
  distributeCloudMachineId.value = row.cloudMachineId
  distributeProtocol.value = row.protocol || 'httpflv'
  distributeResolution.value = row.resolution
  distributeDialogVisible.value = true
}

const stopP2PDistribution = async (row) => {
  try {
    const mappedPort = getDistributionPort(row)
    if (!mappedPort) {
      ElMessage.error('未找到P2P端口映射')
      return
    }
    let device = deviceCloudOptions.value.find(d => d.ip === row.deviceIp) || null
    let cloudMachine = getDeviceCloudMachines(row.deviceIp).find(m => (m.id || m.name) === row.cloudMachineId)
    if (!cloudMachine) {
      const ensured = await ensureCloudMachinesLoaded(row.deviceIp)
      device = ensured.device || device
      cloudMachine = ensured.cloudMachines.find(m => (m.id || m.name) === row.cloudMachineId)
    }
    const p2pHost = getControlHost(device, cloudMachine) || row.deviceIp
    const res = await Call.ByName('main.App.StopP2P', p2pHost, Number(mappedPort), row.streamName)
    if (!res?.success) {
      throw new Error(res?.message || 'P2P 停止失败')
    }
    row.status = 'inactive'
    row.statusReason = '已停止'
  } catch (e) {
    ElMessage.error('停止失败: ' + (e.message || e))
  }
}

const startP2PDistribution = async (row) => {
  try {
    const listenPort = Number(row?.p2pListenPort)
    if (!listenPort) {
      ElMessage.error('缺少P2P端口')
      return
    }
    if (row?.status !== 'active' && p2pUsedPorts.value.has(listenPort)) {
      ElMessage.error('P2P端口冲突，请更换')
      return
    }
    let device = deviceCloudOptions.value.find(d => d.ip === row.deviceIp)
    if (!device) {
      ElMessage.error(_instance.proxy.$t('stream.deviceOfflineOrNotExist'))
      return
    }
    let cloudMachine = getDeviceCloudMachines(row.deviceIp).find(m => (m.id || m.name) === row.cloudMachineId)
    if (!cloudMachine) {
      const ensured = await ensureCloudMachinesLoaded(row.deviceIp)
      device = ensured.device || device
      cloudMachine = ensured.cloudMachines.find(m => (m.id || m.name) === row.cloudMachineId)
    }
    if (!cloudMachine) {
      ElMessage.error('云机信息未找到')
      return
    }
    const controlPort = getMappedPort9082(cloudMachine)
    const mappedPort = getMappedPort10006(cloudMachine)
    if (!controlPort) {
      ElMessage.error('未找到云机9082映射端口')
      return
    }
    if (!mappedPort) {
      ElMessage.error('未找到云机10006映射端口')
      return
    }

    const listenUrl = `srt://${localIp.value}:${listenPort}`
    const controlHost = getControlHost(device, cloudMachine)
    const baseUrl = `http://${controlHost}:${controlPort}`
    const setUrl = `${baseUrl}/modifydev?cmd=4&type=camera&path=${encodeURIComponent(listenUrl)}&resolution=${p2pResolution.value}`
    const setResData = await ProxyHttpGet(setUrl)
    if (setResData?.code !== 200) {
      const reason = setResData?.reason || ''
      if (reason.includes('path is empty')) {
        ElMessage.warning('老镜像不支持')
        const fallbackUrl = `${baseUrl}/modifydev?cmd=14&type=camera`
        const fallbackRes = await ProxyHttpGet(fallbackUrl)
        if (fallbackRes?.code !== 200) {
          throw new Error('出错了，可能是较老镜像不支持或者网络问题')
        }
      } else {
        throw new Error(setResData?.reason || '设置推流失败')
      }
    }
    try {
      const startUrl = `${baseUrl}/camera?cmd=stop`
      await ProxyHttpGet(startUrl)
    } catch {}

    const p2pHost = getControlHost(device, cloudMachine) || row.deviceIp
    const res = await Call.ByName('main.App.StartP2P', p2pHost, Number(mappedPort), row.streamName, listenPort)
    if (!res?.success) {
      throw new Error(res?.message || 'P2P 启动失败')
    }
    row.status = 'active'
    row.statusReason = `SRT监听: ${res?.listenUrl || getP2PListenUrl(row)}`
    row.mappedPort = mappedPort
  } catch (e) {
    const message = e?.message || e
    if (message === '出错了，可能是较老镜像不支持或者网络问题') {
      ElMessage.error(message)
    } else {
      ElMessage.error('启动失败: ' + message)
    }
  }
}

const openProjectionForDistribution = async (item) => {
  try {
    const machines = props.deviceCloudMachinesCache?.get?.(item.deviceIp)
    const container = Array.isArray(machines)
      ? machines.find(m => m.id === item.cloudMachineId || m.name === item.cloudMachineName)
      : null
    if (!container) {
      ElMessage.warning('未找到云机信息，请先刷新设备云机列表')
      return
    }
    await startProjection({ ip: item.deviceIp }, container)
    // ElMessage.success(t('common.投屏已启动'))
  } catch (e) {
    ElMessage.error('投屏失败: ' + (e?.message || e))
  }
}

const removeDistribution = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除这条分发记录吗？', '提示', { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') })
    if (row?.protocol === 'p2p') {
      await stopP2PDistribution(row)
    }
    const index = streamDistributions.value.findIndex(item => item.id === row.id)
    if (index !== -1) {
      streamDistributions.value.splice(index, 1)
      ElMessage.success(t('common.已删除'))
    }
  } catch (e) {}
}

const isMytNetwork = (machine) => {
  return (machine?.networkName || '').toString().toLowerCase() === 'myt'
}

const getControlHost = (device, machine) => {
  if (isMytNetwork(machine)) return machine?.ip || device?.ip || ''
  let ip = device?.ip || ''
  // OpenCecs 公网设备：deviceIp 含端口（如 219.139.239.165:19218），提取纯 IP
  if (ip && ip.includes(':')) ip = ip.split(':')[0]
  return ip
}

const extractPort = (container, portNumber) => {
  if (!container) return null
  const cacheKey = `_cachedPort_${portNumber}`
  if (container[cacheKey]) return container[cacheKey]
  let mappedPort = null
  const portKeyTcp = `${portNumber}/tcp`
  const portKeyUdp = `${portNumber}/udp`
  if (Array.isArray(container.Ports)) {
    const port = container.Ports.find(p => p.PrivatePort === portNumber && p.Type === 'tcp')
      || container.Ports.find(p => p.PrivatePort === portNumber && p.Type === 'udp')
    if (port && port.PublicPort) {
      mappedPort = port.PublicPort
    }
  }
  if (!mappedPort && container.NetworkSettings && container.NetworkSettings.Ports) {
    const portBinding = container.NetworkSettings.Ports[portKeyTcp] || container.NetworkSettings.Ports[portKeyUdp]
    if (portBinding && portBinding.length > 0) {
      mappedPort = portBinding[0].HostPort
    }
  }
  if (!mappedPort && container.PortBindings) {
    const portBinding = container.PortBindings[portKeyTcp] || container.PortBindings[portKeyUdp]
    if (portBinding && portBinding.length > 0) {
      mappedPort = portBinding[0].HostPort
    }
  }
  if (!mappedPort && container.portBindings) {
    const portBinding = container.portBindings[portKeyTcp] || container.portBindings[portKeyUdp]
    if (portBinding && portBinding.length > 0) {
      mappedPort = portBinding[0].HostPort
    }
  }

  // OpenCecs 公网设备：将 LAN HostPort 转换为公网端口
  if (mappedPort && window.openCecsPortMap) {
    const deviceIp = container.deviceIp || ''
    if (deviceIp && deviceIp.includes(':')) {
      let portMap = window.openCecsPortMap.get(deviceIp)
      if (!portMap) {
        const ipPrefix = deviceIp.split(':')[0] + ':'
        for (const [key, map] of window.openCecsPortMap) {
          if (key.startsWith(ipPrefix)) { portMap = map; break }
        }
      }
      if (portMap) {
        const publicPort = portMap.get(Number(mappedPort))
        if (publicPort) mappedPort = publicPort
      }
    }
  }

  container[cacheKey] = mappedPort
  return mappedPort
}

const getMappedPort9082 = (machine) => {
  if (isMytNetwork(machine)) return 9082
  const mappedPort = extractPort(machine, 9082)
  if (mappedPort) return mappedPort
  if (machine?.indexNum) return 10000 + Number(machine.indexNum)
  return null
}

const getMappedPort10006 = (machine) => {
  if (isMytNetwork(machine)) return 10006
  const mappedPort = extractPort(machine, 10006)
  if (mappedPort) return mappedPort
  return null
}

const getDistributionPort = (item) => {
  if (item?.mappedPort) return item.mappedPort
  const device = deviceCloudOptions.value.find(d => d.ip === item.deviceIp)
  if (!device) return null
  const cloudMachine = device.cloudMachines?.find(m => (m.id || m.name) === item.cloudMachineId)
  if (!cloudMachine) return null
  if (item?.protocol === 'p2p') return getMappedPort10006(cloudMachine)
  return getMappedPort9082(cloudMachine)
}

// 代理 HTTP GET 请求，解决跨域问题
// ID: 3591549565 (main.App.ProxyHttpGet)
const ProxyHttpGet = async (url) => {
  try {
    const result = await Call.ByID(3591549565, url)
    if (typeof result === 'string') {
      try {
        return JSON.parse(result)
      } catch (e) {
        return { code: 500, reason: 'Response Parse Error' }
      }
    }
    return result || { code: 500, reason: 'Empty Response' }
  } catch (e) {
    console.error('ProxyHttpGet error:', e)
    throw e
  }
}

const buildStreamPath = (protocol, streamName) => {
  if (protocol === 'httpflv') {
    return `http://${localIp.value}:8083/live/${streamName}.flv`
  }
  if (protocol === 'rtmp') {
    return `rtmp://${localIp.value}:1935/live/${streamName}`
  }
  return `http://${localIp.value}:8083/live/${streamName}.flv`
}

const buildProtocolType = (protocol) => {
  if (protocol === 'httpflv') return 'webrtc'
  if (protocol === 'rtmp') return 'rtmp'
  return 'webrtc'
}

const confirmDistribute = async () => {
  if (!distributeStreamName.value) {
    ElMessage.warning('请选择流')
    return
  }
  if (!distributeDeviceIp.value) {
    ElMessage.warning('请选择设备')
    return
  }
  if (!distributeCloudMachineId.value) {
    ElMessage.warning('请选择云机')
    return
  }
  const device = distributeDevice.value
  const cloudMachine = distributeAvailableCloudMachines.value.find(
    m => (m.id || m.name) === distributeCloudMachineId.value
  )
  if (!device || !cloudMachine) {
    ElMessage.warning('未找到对应云机')
    return
  }

  applyLoading.value = true
  try {
    const existingP2P = streamDistributions.value.find(item =>
      item.protocol === 'p2p' &&
      item.deviceIp === device.ip &&
      item.cloudMachineId === (cloudMachine.id || cloudMachine.name)
    )
    if (existingP2P) {
      await stopP2PDistribution(existingP2P)
    }

    const protocol = distributeProtocol.value || 'httpflv'
    const mappedPort = getMappedPort9082(cloudMachine)
    if (!mappedPort) {
      ElMessage.error('未找到云机9082映射端口，无法设置推流')
      return
    }

    const streamPath = buildStreamPath(protocol, distributeStreamName.value)
    const protocolType = buildProtocolType(protocol)
    const controlHost = getControlHost(device, cloudMachine)
    const baseUrl = `http://${controlHost}:${mappedPort}`
    
    // 1. 设置推流 (cmd=4) - 使用后端代理避免跨域
    const setUrl = `${baseUrl}/modifydev?cmd=4&type=${protocolType}&path=${encodeURIComponent(streamPath)}&resolution=${distributeResolution.value}`
    const setResData = await ProxyHttpGet(setUrl)
    
    if (setResData?.code !== 200) {
      throw new Error(setResData?.reason || '设置推流失败')
    }

    // 2. 触发播放启动（避免设备侧等待）
    try {
      const startUrl = `${baseUrl}/camera?cmd=start`
      await ProxyHttpGet(startUrl)
    } catch {}
    
    const record = {
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      streamName: distributeStreamName.value,
      deviceIp: device.ip,
      cloudMachineId: cloudMachine.id || cloudMachine.name,
      cloudMachineName: cloudMachine.name || cloudMachine.id,
      protocol,
      mappedPort,
      resolution: distributeResolution.value,
      status: 'pending',
      statusReason: '待校验'
    }

    // 检查是否已存在相同设备+云机的记录
    const existingIndex = streamDistributions.value.findIndex(item => 
      item.deviceIp === record.deviceIp && item.cloudMachineId === record.cloudMachineId
    )

    if (existingIndex >= 0) {
      // 如果存在，保留原ID但更新其他信息
      record.id = streamDistributions.value[existingIndex].id
      streamDistributions.value.splice(existingIndex, 1, record)
      ElMessage.success(t('common.已更新分发设置'))
    } else {
      streamDistributions.value.unshift(record)
      ElMessage.success(t('common.分发成功'))
    }
    setTimeout(() => {
      checkDistributionsStatus()
    }, 300)
    distributeDialogVisible.value = false
  } catch (e) {
    console.error(e)
    ElMessage.error('分发失败: ' + (e.message || e))
  } finally {
    applyLoading.value = false
  }
}

const confirmP2PDistribute = async () => {
  if (!p2pDeviceIp.value) {
    ElMessage.warning('请选择设备')
    return
  }
  if (!p2pCloudMachineId.value) {
    ElMessage.warning('请选择云机')
    return
  }
  const listenPort = Number(p2pListenPort.value)
  if (!listenPort || listenPort <= 0) {
    ElMessage.warning('请输入有效的P2P端口')
    return
  }
  if (p2pUsedPorts.value.has(listenPort)) {
    ElMessage.warning('P2P端口已被占用，请更换')
    return
  }

  const device = p2pDevice.value
  const cloudMachine = p2pAvailableCloudMachines.value.find(
    m => (m.id || m.name) === p2pCloudMachineId.value
  )
  if (!device || !cloudMachine) {
    ElMessage.warning('未找到对应云机')
    return
  }

  const controlPort = getMappedPort9082(cloudMachine)
  const mappedPort = getMappedPort10006(cloudMachine)
  if (!controlPort) {
    ElMessage.error('未找到云机9082映射端口，无法设置推流')
    return
  }
  if (!mappedPort) {
    ElMessage.error('未找到云机10006映射端口，无法设置推流')
    return
  }

  applyLoading.value = true
  try {
    const streamName = (p2pStreamName.value || '').trim() || `p2p-${listenPort}`

    const controlHost = getControlHost(device, cloudMachine)
    const baseUrl = `http://${controlHost}:${controlPort}`
    const setUrl = `${baseUrl}/modifydev?cmd=4&type=camera&resolution=1`
    const setResData = await ProxyHttpGet(setUrl)
    if (setResData?.code !== 200) {
      throw new Error(setResData?.reason || '设置推流失败')
    }
    try {
      const startUrl = `${baseUrl}/camera?cmd=start`
      await ProxyHttpGet(startUrl)
    } catch {}

    const p2pHost = getControlHost(device, cloudMachine) || device.ip
    const res = await Call.ByName('main.App.StartP2P', p2pHost, Number(mappedPort), streamName, listenPort)
    if (!res?.success) {
      throw new Error(res?.message || 'P2P 启动失败')
    }
    const effectiveListenUrl = res?.listenUrl || listenUrl

    const record = {
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      streamName,
      deviceIp: device.ip,
      cloudMachineId: cloudMachine.id || cloudMachine.name,
      cloudMachineName: cloudMachine.name || cloudMachine.id,
      protocol: 'p2p',
      mappedPort,
      p2pListenPort: listenPort,
      resolution: p2pResolution.value,
      status: 'pending',
      statusReason: `SRT监听: ${effectiveListenUrl}`
    }

    const existingIndex = streamDistributions.value.findIndex(item =>
      item.deviceIp === record.deviceIp && item.cloudMachineId === record.cloudMachineId
    )
    if (existingIndex >= 0) {
      record.id = streamDistributions.value[existingIndex].id
      streamDistributions.value.splice(existingIndex, 1, record)
      ElMessage.success(t('common.已更新P2P分发'))
    } else {
      streamDistributions.value.unshift(record)
      ElMessage.success(t('common.P2P分发成功'))
    }

    setTimeout(() => {
      checkDistributionsStatus()
    }, 300)
    p2pDialogVisible.value = false
  } catch (e) {
    console.error(e)
    ElMessage.error('P2P分发失败: ' + (e.message || e))
  } finally {
    applyLoading.value = false
  }
}

const handleStopPush = async (stream) => {
  try {
    await ElMessageBox.confirm('确定要断开此推流吗？', '提示', { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') })
    const res = await StopPushSession(stream.streamName)
    if (res.success) {
      ElMessage.success(t('common.已断开'))
      refreshStreams()
    } else {
      ElMessage.error(res.message)
    }
  } catch (e) {}
}

const copyToClipboard = (text) => {
  if (!text) {
    ElMessage.warning('没有可复制的内容')
    return
  }
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success(t('common.已复制到剪贴板'))
  }).catch(err => {
    console.error('复制失败:', err)
    ElMessage.error('复制失败')
  })
}

// ==============================
// 摄像头模式
// ==============================
const cameraDialogVisible = ref(false)
const camDeviceIp = ref('')
const camCloudMachineId = ref('')
const camName = ref('')
const camWidth = ref(1280)
const camHeight = ref(720)

// 预览
const camPreviewDataURL = ref('')     // base64 DataURL
const camPreviewLoading = ref(false)
const camPreviewActiveId = ref('')    // 当前预览对应的 cameraStream.id
const camPreviewInterval = 1          // 实时预览刷新间隔（秒）
let camPreviewTimer = null            // 定时器句柄
const camFps = ref(30)
const camBitrate = ref(2000)
const camDeviceList = ref([])    // 可用摄像头名称列表
const camListLoading = ref(false)

// 摄像头流列表，持久化到 LocalStorage
const cameraStreams = ref([])
try {
  const saved = localStorage.getItem('cameraStreams')
  if (saved) cameraStreams.value = JSON.parse(saved)
} catch (e) {}
watch(cameraStreams, val => {
  localStorage.setItem('cameraStreams', JSON.stringify(val))
}, { deep: true })

const camDevice = computed(() => deviceCloudOptions.value.find(d => d.ip === camDeviceIp.value) || null)
const camAvailableCloudMachines = computed(() => {
  if (!camDevice.value?.ip) return []
  return getDeviceCloudMachines(camDevice.value.ip).filter(m => m.status === 'running')
})

// 当前选中云机对应的 10006 映射端口（设备端口）
const camSelectedCloudMachine = computed(() => {
  if (!camCloudMachineId.value) return null
  return camAvailableCloudMachines.value.find(m => (m.id || m.name) === camCloudMachineId.value) || null
})
const camDetectedPort = computed(() => {
  if (!camSelectedCloudMachine.value) return null
  return getMappedPort10006(camSelectedCloudMachine.value) || null
})

watch(camDeviceIp, () => { camCloudMachineId.value = '' })

// 弹窗打开时刷新设备缓存并拉取摄像头列表
watch(cameraDialogVisible, async (visible) => {
  if (visible) {
    ensureDeviceCacheForDialogs()
    await fetchCameraList()
  }
})

const fetchCameraList = async () => {
  camListLoading.value = true
  try {
    const res = await Call.ByName('main.App.ListCameraDevices')
    console.log('[Camera] ListCameraDevices result:', res)
    if (res?.success && Array.isArray(res.cameras) && res.cameras.length > 0) {
      camDeviceList.value = res.cameras
      if (camName.value && !res.cameras.includes(camName.value)) {
        camName.value = ''
      }
    } else {
      camDeviceList.value = []
    }
  } catch (e) {
    console.error('[Camera] ListCameraDevices failed:', e)
    camDeviceList.value = []
  } finally {
    camListLoading.value = false
  }
}

const openCameraDialog = () => {
  camDeviceIp.value = ''
  camCloudMachineId.value = ''
  camName.value = ''
  camWidth.value = 1280
  camHeight.value = 720
  camFps.value = 30
  camBitrate.value = 2000
  cameraDialogVisible.value = true
}

const confirmCameraStream = () => {
  if (!camDeviceIp.value) { ElMessage.warning('请选择设备'); return }
  if (!camCloudMachineId.value) { ElMessage.warning('请选择云机'); return }
  const device = camDevice.value
  const cloudMachine = camSelectedCloudMachine.value
  if (!device || !cloudMachine) { ElMessage.warning('未找到对应云机'); return }

  const devicePort = camDetectedPort.value
  if (!devicePort) { ElMessage.warning('无法获取设备端口（云机未映射10006端口）'); return }

  const record = {
    id: `cam-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    deviceIp: device.ip,
    controlHost: getControlHost(device, cloudMachine),
    devicePort: Number(devicePort),
    cloudMachineId: cloudMachine.id || cloudMachine.name,
    cloudMachineName: cloudMachine.name || cloudMachine.id,
    camName: camName.value.trim(),
    width: camWidth.value || 1280,
    height: camHeight.value || 720,
    fps: camFps.value || 30,
    bitrate: camBitrate.value || 2000,
    status: 'stopped',
    loading: false
  }
  cameraStreams.value.unshift(record)
  cameraDialogVisible.value = false
  ElMessage.success(t('common.已添加，点击「启动」开始推流'))
}

const startCameraStream = async (item) => {
  item.loading = true
  try {
    const camNameArg = item.camName || ''
    const args = [
      '--camera',
      item.controlHost || item.deviceIp,
      String(item.devicePort || 30105),
      '--width', String(item.width),
      '--height', String(item.height),
      '--fps', String(item.fps),
      '--bitrate', String(item.bitrate),
      ...(camNameArg ? ['--cam-name', camNameArg] : [])
    ]
    const res = await Call.ByName('main.App.StartWindowsPusher', args)
    if (res?.success) {
      item.status = 'running'
      item.pid = res.pid
      ElMessage.success(t('common.摄像头推流已启动'))
      // 启动成功后抓取预览帧并开启实时刷新
      camPreviewActiveId.value = item.id
      fetchCameraPreview(item)
      startPreviewTimer(item)
    } else {
      ElMessage.error('启动失败: ' + (res?.message || '未知错误'))
    }
  } catch (e) {
    ElMessage.error('启动失败: ' + (e?.message || e))
  } finally {
    item.loading = false
  }
}

// 抓取摄像头预览帧（keepOld=true 时保留旧图，避免刷新时闪烁）
const fetchCameraPreview = async (item, keepOld = false) => {
  if (!keepOld) camPreviewDataURL.value = ''
  camPreviewLoading.value = true
  try {
    const res = await Call.ByName('main.App.CapturePreviewFrame',
      item.camName || '', 640, 360)
    if (res?.success && res.dataURL) {
      camPreviewDataURL.value = res.dataURL
    } else {
      console.warn('[Preview] 未能获取预览:', res?.message)
    }
  } catch (e) {
    console.error('[Preview] 预览失败:', e)
  } finally {
    camPreviewLoading.value = false
  }
}

// 启动实时预览定时器
const startPreviewTimer = (item) => {
  stopPreviewTimer()
  camPreviewTimer = setInterval(() => {
    // 只有该摄像头仍在运行时才刷新
    const current = cameraStreams.value.find(s => s.id === item.id)
    if (current && current.status === 'running') {
      fetchCameraPreview(current, true)
    } else {
      stopPreviewTimer()
    }
  }, camPreviewInterval * 1000)
}

// 停止实时预览定时器
const stopPreviewTimer = () => {
  if (camPreviewTimer) {
    clearInterval(camPreviewTimer)
    camPreviewTimer = null
  }
}

// 刷新预览（手动点击）
const refreshCameraPreview = () => {
  const item = cameraStreams.value.find(s => s.id === camPreviewActiveId.value)
  if (item) fetchCameraPreview(item, true)
}

const stopCameraStream = async (item) => {
  item.loading = true
  try {
    const res = await Call.ByName('main.App.StopWindowsPusher', item.pid || 0)
    if (res?.success || true) {
      item.status = 'stopped'
      item.pid = null
      ElMessage.success(t('common.已停止'))
      // 停止定时器并清除预览
      if (camPreviewActiveId.value === item.id) {
        stopPreviewTimer()
        camPreviewDataURL.value = ''
        camPreviewActiveId.value = ''
      }
    }
  } catch (e) {
    item.status = 'stopped'
    item.pid = null
    stopPreviewTimer()
    camPreviewDataURL.value = ''
    camPreviewActiveId.value = ''
  } finally {
    item.loading = false
  }
}

const removeCameraStream = async (item) => {
  try {
    await ElMessageBox.confirm('确定要删除这条摄像头推流吗？', '提示', { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') })
    if (item.status === 'running') await stopCameraStream(item)
    const idx = cameraStreams.value.findIndex(c => c.id === item.id)
    if (idx !== -1) cameraStreams.value.splice(idx, 1)
    // 若删除的是当前预览行，清除预览
    if (camPreviewActiveId.value === item.id) {
      stopPreviewTimer()
      camPreviewDataURL.value = ''
      camPreviewActiveId.value = ''
    }
    ElMessage.success(t('common.已删除'))
  } catch (e) {}
}

</script>

<style scoped>
.stream-management {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.status-card {
  margin-bottom: 10px;
  height: auto !important;
}
.server-info h3 {
  margin: 0 0 10px 0;
}
.main-content {
  display: flex;
  gap: 20px;
  flex: 1;
  min-height: 0;
}
.stream-list-card, .client-list-card {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
}

.device-cloud-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  min-height: 0;
}

.device-cloud-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.modifydev-form {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.modifydev-status {
  padding: 10px 12px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fafafa;
}

.modifydev-status-title {
  font-weight: 600;
  margin-bottom: 6px;
}

.modifydev-status-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.native-table-container {
  flex: 1;
  overflow: auto;
  width: 100%;
}

.native-table {
  width: 100%;
  border-collapse: collapse;
  color: #606266;
}

.native-table th, .native-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ebeef5; /* 使用更通用的边框颜色 */
}

.native-table th.cloud-machine-col,
.native-table td.cloud-machine-col {
  width: 160px;
  max-width: 160px;
}

.cloud-machine-name {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.inactive-status-scroll {
  max-width: 180px;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
}

.native-table th {
  background-color: #f5f7fa; /* 浅色表头 */
  color: #606266;
  font-weight: bold;
  position: sticky;
  top: 0;
}

.native-table tr:hover {
  background-color: #f5f7fa; /* 悬停效果 */
}

/* 如果确实是深色模式，可以通过类名区分，或者让用户自己适配。
   这里为了保险，先用通用样式，确保文字可见。 */

.status-card :deep(.el-card__body) {
  display: block;
  flex: initial;
  overflow: visible;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 使用说明样式 */
.guide-container {
  background: #fafbfc;
  border-radius: 8px;
}

.guide-section {
  margin-bottom: 4px;
}

.guide-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
  padding: 6px 10px;
  background: #ecf5ff;
  border-left: 4px solid #409eff;
  border-radius: 0 4px 4px 0;
}

.guide-icon {
  margin-right: 4px;
}

.guide-steps {
  margin: 0;
  padding-left: 20px;
  color: #606266;
}

.guide-steps li {
  margin-bottom: 8px;
}

.guide-tip {
  display: inline-block;
  margin-top: 4px;
  font-size: 12px;
  color: #e6a23c;
  background: #fdf6ec;
  border-radius: 4px;
  padding: 2px 8px;
}

.action-buttons .el-button {
  min-width: 80px;
  border-radius: 6px;
  transition: all 0.3s ease;
  font-weight: 500;
}

.action-buttons .el-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.action-buttons .el-button:active:not(:disabled) {
  transform: translateY(0);
}

.action-buttons .el-button i {
  margin-right: 4px;
}

.action-buttons .el-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.el-button+.el-button {
  margin-left: 0px;
}

</style>
