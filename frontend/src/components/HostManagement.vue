<template>
  <el-row :gutter="12" style="height: 100%; min-height: 600px;">
    <el-col :span="24" class="device-left-col" style="height: 100%;">
      <el-card shadow="hover" class="device-card" style="height: 100%;">
        <template #header>
          <div class="card-header" style="display: flex; justify-content: space-between; align-items: center;">
            <div style="flex: 1;"></div>
            <div style="display: flex; gap: 8px;">
              <el-button 
                v-if="localSelectedHostDevices.length > 0"
                type="danger" 
                size="small" 
                @click="$emit('handleBatchDeleteDevices')" 
                class="delete-button"
              >
                <el-icon><Delete /></el-icon> {{ t('common.deleteSelected') }}
              </el-button>
              <el-button 
                type="primary" 
                size="small" 
                @click="debouncedRefresh" 
                class="refresh-button"
                :disabled="loading"
              >
                <el-icon :class="{ 'is-rotating': loading }"><Refresh /></el-icon> {{ t('common.refresh') }}
              </el-button>
               <el-button 
                type="primary" 
                size="small" 
                v-if="!token"
                @click="$emit('showSyncAuthDialog')"
                class="sync-auth-button"
              >
                <el-icon><List /></el-icon> {{ t('common.syncAuth') }}
              </el-button>
               <el-button 
                type="success" 
                size="small" 
                v-else
                class="sync-auth-button"
                @click="$emit('handleSyncAuthorization')"
              >
                <el-icon><HelpFilled /></el-icon> {{ uname }} | {{ t('common.syncAuth') }}
              </el-button>
               <el-button 
                type="success" 
                size="small" 
                @click="handleBindHost"
                class="sync-auth-button"
              >
                <el-icon><HelpFilled /></el-icon> {{ t('common.bindHost') }}
              </el-button>
               <el-button 
                type="warning" 
                size="small" 
                @click="handleUnbindHost"
                class="sync-auth-button"
              >
               {{ t('common.unbindHost') }}
              </el-button>
              <el-button 
                type="success" 
                size="small" 
                @click="showAddDeviceDialog"
                class="add-device-button"
              >
                <el-icon><Plus /></el-icon>
                <span>{{ t('common.addDevice') }}</span>
              </el-button>
              <el-button 
                type="primary" 
                size="small" 
                @click="startBatchUpgrade" 
                :disabled="isBatchUpgrading"
                class="batch-upgrade-button"
              >
                <el-icon><Download /></el-icon> {{ t('common.batchUpgradeAPI') }}
              </el-button>
              <el-button 
                type="warning" 
                size="small" 
                @click="$emit('clearCache')"
                class="clear-cache-button"
              >
                <el-icon><Delete /></el-icon> {{ t('common.clearClientData') }}
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                @click="$emit('handleBatchDeleteHosts')"
                class="clear-cache-button"
              >
                <el-icon><Delete /></el-icon> {{ t('common.batchCleanDisk') }}
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                @click="$emit('logOut')"
                class="clear-cache-button"
                v-if="token"
              >
                <el-icon><CloseBold /></el-icon> {{ t('common.logout') }}
              </el-button>
            </div>
          </div>
        </template>
        <!-- 设备列表 -->
      <div class="device-list-header">
        <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
          <div class="device-list-tabs">
            <el-button 
              type="link" 
              size="small" 
              class="device-tab" 
              :class="{ active: localDeviceFilter === 'online' }"
              @click="localDeviceFilter = 'online'"
            >{{ t('common.onlineDevices') }}</el-button>
            <el-button 
              type="link" 
              size="small" 
              class="device-tab"
              :class="{ active: localDeviceFilter === 'offline' }"
              @click="localDeviceFilter = 'offline'"
            >{{ t('common.offlineDevices') }}</el-button>
          </div>
          <div class="device-group-filter" style="display: flex; align-items: center; gap: 8px;">
            <el-input
              v-model="searchIP"
              :placeholder="t('common.searchIP')"
              size="small"
              style="width: 150px;"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <span style="font-size: 12px; color: #909399;width: 30px;">{{ t('common.group') }}:</span>
            <el-select 
              v-model="localGroupFilter" 
              size="small" 
              style="width: 120px;"
              :placeholder="$t('common.selectGroupPlaceholder')"
            >
              <el-option :label="t('common.all')" :value="t('common.all')"></el-option>
              <el-option 
                v-for="group in deviceGroups" 
                :key="group" 
                :label="group" 
                :value="group"
              ></el-option>
            </el-select>
            <el-button 
              type="primary" 
              size="small"
              text
              @click="handleAddGroup"
              :title="t('common.addGroup')"
            >
              <el-icon><Plus /></el-icon>
            </el-button>
            <el-button 
              type="primary" 
              size="small"
              text
              @click="handleEditGroup"
              :disabled="!localGroupFilter || localGroupFilter === '全部' || localGroupFilter === '默认分组'"
              :title="t('common.editGroup')"
            >
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-button 
              type="primary" 
              size="small"
              text
              @click="handleDeleteGroup"
              :disabled="!localGroupFilter || localGroupFilter === '全部' || localGroupFilter === '默认分组'"
              :title="t('common.deleteGroup')"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
          
        <!-- 设备表格 -->
        <el-table 
          ref="deviceTableRef"
          :data="displayDevices" 
          stripe 
          size="small" 
          class="device-table"
          row-key="id"
          v-model:selection="localSelectedHostDevices"
          @selection-change="handleDirectSelectionChange"
        >
          <el-table-column v-if="!isViewingDeviceDetails" type="selection" width="55" :reserve-selection="true"></el-table-column>
          <el-table-column :label="t('common.device')" width="250" align="center">
            <template #default="scope">
              <div class="device-info-cell" style="display: flex; align-items: center; justify-content: center;">
                <div 
                  class="device-type-icon"
                  :style="{
                    backgroundColor: getDeviceTypeColor(scope.row.name),
                    color: '#fff',
                    borderRadius: '4px',
                    width: '24px',
                    height: '24px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '14px',
                    fontWeight: 'bold',
                    marginRight: '8px'
                  }"
                >
                  {{ scope.row.name.charAt(0).toUpperCase() }}
                </div>
                <div style="display: flex; flex-direction: column; align-items: flex-start;">
                  <span class="device-ip">{{ scope.row.ip }}</span>
                  <span 
                    v-if="devicesStatusCache.get(scope.row.id) === 'online' && deviceFirmwareInfo.get(scope.row.id)?.responseTime"
                    style="font-size: 11px; color: #67c23a;"
                  >
                    {{ t('common.deviceNetworkLatency') }}：{{ deviceFirmwareInfo.get(scope.row.id).responseTime }}ms
                  </span>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.hostFirmwareVersion')" min-width="120" align="center">
            <template #default="scope">
              <div>
                {{ formatSdkVersion(deviceFirmwareInfo.get(scope.row.id)?.sdkVersion) || t('common.unknown') }}
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.nvmeStorage')" min-width="180" align="center">
            <template #default="scope">
              <div>
                {{ formatStorage(deviceFirmwareInfo.get(scope.row.id)?.originalData?.mmcuse || 0, deviceFirmwareInfo.get(scope.row.id)?.originalData?.mmctotal || 0) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.bindStatus')" width="140" align="center">
            <template #default="scope">
              <el-tag 
                :type="getBindStatusType(scope.row.id)"
                size="small"
                style="max-width: 130px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: inline-block; vertical-align: middle;"
              >
                {{ getBindStatusText(scope.row.id) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.apiVersion')" width="200" align="center" v-if="localDeviceFilter === 'online'">
            <template #default="scope">
              <div>
                <div v-if="deviceVersionInfo.get(scope.row.id)" style="display: flex; align-items: center; flex-wrap: nowrap; gap: 8px;">
                  <div style="display: flex; align-items: center; gap: 8px;">
                    <span>{{ t('common.current') }}: {{ deviceVersionInfo.get(scope.row.id).currentVersion || t('common.unknown') }}</span>
                    <span>/</span>
                    <span>{{ t('common.latest') }}: {{ deviceVersionInfo.get(scope.row.id).latestVersion || t('common.unknown') }}</span>
                  </div>
                  <el-button
                    v-if="Number(deviceVersionInfo.get(scope.row.id).currentVersion) < Number(deviceVersionInfo.get(scope.row.id).latestVersion)"
                    size="small"
                    type="primary"
                    style="margin-left: auto;"
                    @click="handleUpgradeDevice(scope.row)"
                  >
                    {{ t('common.upgrade') }}
                  </el-button>
                </div>
                <div v-else>
                  <el-icon class="is-loading"><Loading /></el-icon> {{ t('common.loading') }}...
                </div>
              </div>
            </template>
          </el-table-column>
          <!-- <el-table-column label="云机数量" width="100" align="center">
            <template #default="scope">
              {{ scope.row.containers ? scope.row.containers.length : 0 }}
            </template>
          </el-table-column> -->
          <el-table-column :label="t('common.operation')" width="300" fixed="right" align="center" v-if="localDeviceFilter === 'online'">
            <template #default="scope">
              <el-space size="small">
                <el-button 
                  size="small" 
                  type="primary" 
                  @click="$emit('showDeviceDetails', scope.row)"
                >
                  {{ t('common.view') }}
                </el-button>
                <el-dropdown trigger="click" @command="(cmd) => handleDeviceGroupCommand(cmd, scope.row)">
                  <el-button size="small" class="group-dropdown-btn">
                    <span class="group-text">{{ t('common.group') }}: {{ scope.row.group || t('common.defaultGroup') }}</span>
                    <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item 
                        v-for="group in deviceGroups" 
                        :key="group" 
                        :command="group"
                        :disabled="scope.row.group === group"
                      >
                        {{ group }}
                      </el-dropdown-item>
                      <el-dropdown-item divided command="add-group">
                        <el-icon><Plus /></el-icon> {{ t('common.addNewGroup') }}
                      </el-dropdown-item>
                      <el-dropdown-item 
                        command="delete-group" 
                        :disabled="scope.row.group === '默认分组' || !scope.row.group"
                      >
                        <el-icon><Delete /></el-icon> {{ t('common.deleteCurrentGroup') }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <el-button 
                  size="small" 
                  type="text" 
                  :icon="Plus"
                  @click="$emit('showCreateDialog', scope.row, 'batch')"
                  :title="t('common.create')"
                ></el-button>
                <!-- <el-button 
                  size="small" 
                  type="danger" 
                  @click="$emit('handleDeleteDevice', scope.row)"
                >
                  {{ $t('common.delete') }}
                </el-button> -->
              </el-space>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-col>
    
    <el-col :span="12" class="device-right-col" style="height: 100%; position: relative; padding: 0 20px;">

      <!-- 收起按钮 - 覆盖整个左侧边缘 -->
      <div 
        class="collapse-button" 
        @click="$emit('collapseRightSidebar')" 
        style="position: absolute; left: -15px; top: 0; bottom: 0; width: 30px; z-index: 100; cursor: pointer; transition: all 0.3s ease; background-color: transparent;"
        :title="t('common.collapseRightSidebar')"
        @mouseenter="(e) => {
          e.currentTarget.style.backgroundColor = 'rgba(228, 231, 237, 0.5)';
          const btn = e.currentTarget.querySelector('.sidebar-collapse-btn');
          if (btn) {
            btn.style.backgroundColor = '#e4e7ed';
            btn.style.transform = 'scale(1.05)';
            btn.style.boxShadow = '-3px 0 10px rgba(0,0,0,0.15)';
          }
        }"
        @mouseleave="(e) => {
          e.currentTarget.style.backgroundColor = 'transparent';
          const btn = e.currentTarget.querySelector('.sidebar-collapse-btn');
          if (btn) {
            btn.style.backgroundColor = '#f0f0f0';
            btn.style.transform = 'scale(1)';
            btn.style.boxShadow = '-2px 0 6px rgba(0,0,0,0.1)';
          }
        }"
      >
        <div style="height: 100%; display: flex; align-items: center; justify-content: center;">
          <div 
            class="sidebar-collapse-btn"
            :title="t('common.collapseRightSidebar')"
            style="
              background-color: #f0f0f0; 
              color: #606266;
              border: 1px solid #dcdfe6;
              border-radius: 50%; 
              width: 40px; 
              height: 40px; 
              display: flex; 
              align-items: center; 
              justify-content: flex-end; 
              padding-right: 10px;
              box-shadow: -2px 0 6px rgba(0,0,0,0.1);
              transition: all 0.3s ease;
              cursor: pointer;
              margin-left: -5px;
            "
          >
            <!-- 向右箭头 -->
            <div style="
              width: 0;
              height: 0;
              border-top: 6px solid transparent;
              border-bottom: 6px solid transparent;
              border-left: 8px solid #606266;
              transition: all 0.3s ease;
            "></div>
          </div>
        </div>
      </div>
      <!-- 悬浮功能栏 -->
      <div class="floating-toolbar">
        <div class="device-info">
          <span class="device-ip-text">{{ activeDevice?.ip || t('common.noDeviceSelected') }}</span>
          <el-button 
            type="primary" 
            size="small" 
            @click="activeDevice && $emit('fetchAndroidContainers', activeDevice, true)"
            :disabled="!activeDevice || loading"
            class="refresh-button"
          >
            <el-icon :class="{ 'is-rotating': loading }"><Refresh /></el-icon> {{ t('common.refreshCloudMachine') }}
          </el-button>
          <span class="divider">|</span>
          <el-button 
            :type="currentRightTab === 'instance' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'instance' }"
            @click="currentRightTab = 'instance'"
          >{{ t('common.instance') }}</el-button>
          <el-button 
            :type="currentRightTab === 'image' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'image' }"
            @click="currentRightTab = 'image'"
          >{{ t('common.image') }}</el-button>
          <el-button 
            :type="currentRightTab === 'network' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'network' }"
            @click="currentRightTab = 'network'"
          >{{ t('common.network') }}</el-button>
          <el-button 
            :type="currentRightTab === 'host' ? 'primary' : 'text'" 
            size="small" 
            class="toolbar-button" 
            :class="{ active: currentRightTab === 'host' }"
            @click="currentRightTab = 'host'"
          >{{ t('common.host') }}</el-button>
        </div>
        
        <!-- 批量操作按钮 -->
        <el-space wrap class="batch-actions">
          <template v-if="currentRightTab === 'instance'">
            <el-button @click="$emit('handleBatchAction', 'restart')" size="small">{{ t('common.batchRestart') }}</el-button>
            <el-button @click="$emit('handleBatchAction', 'reset')" size="small">{{ t('common.batchReset') }}</el-button>
            <el-button @click="$emit('handleBatchAction', 'projection')" size="small">{{ t('common.batchProjection') }}</el-button>
            <el-button 
              :type="props.isBatchProjectionControlling ? 'warning' : 'primary'" 
              @click="$emit('handleBatchAction', 'projection-control')" 
              size="small"
            >
              {{ props.isBatchProjectionControlling ? t('common.stopControl') : t('common.batchControl') }}
            </el-button>
            <el-button @click="$emit('handleBatchAction', 'shutdown')" size="small">{{ t('common.batchShutdown') }}</el-button>
            <el-button type="danger" @click="$emit('handleBatchAction', 'delete')" size="small">{{ t('common.batchDelete') }}</el-button>
            <el-button type="primary" @click="$emit('handleBatchAction', 'new')" size="small">{{ t('common.batchNewDevice') }}</el-button>
          </template>
          <el-button 
            v-if="currentRightTab === 'image'"
            type="info" 
            size="small" 
            @click="$emit('refreshImageList')" 
            :disabled="!activeDevice || fetchingImages"
            class="refresh-button"
          >
            <el-icon :class="{ 'is-rotating': fetchingImages }"><Refresh /></el-icon> {{ t('common.refreshImages') }}
          </el-button>
          <el-button 
            v-if="currentRightTab === 'network'"
            type="info" 
            size="small" 
            @click="activeDevice && $emit('fetchDockerNetworks', activeDevice)" 
            :disabled="!activeDevice || dockerNetworksLoading"
            class="refresh-button"
          >
            <el-icon :class="{ 'is-rotating': dockerNetworksLoading }"><Refresh /></el-icon> {{ t('common.refreshNetworkList') }}
          </el-button>
          <!-- 🔧 移除刷新主机信息按钮，使用心跳机制自动更新 -->
        </el-space>
      </div>
      
      <!-- 右侧内容区域，根据标签页切换显示不同内容 -->
      <!-- 实例标签页 -->
      <div v-if="currentRightTab === 'instance'" class="table-container" style="overflow-x: auto;">
        <el-table :data="groupedInstances" stripe size="small" class="slot-table" :span-method="({ row, column, rowIndex, columnIndex }) => {
          // 只处理坑位列
          if (columnIndex === 0) {
            if (row.isFirstInSlot) {
              // 第一个实例，设置rowspan为实例总数
              return { rowspan: row.instanceCount, colspan: 1 };
            } else {
              // 非第一个实例，不显示
              return { rowspan: 0, colspan: 0 };
            }
          }
          return { rowspan: 1, colspan: 1 };
        }">
          <el-table-column :label="t('common.slot')" width="100">
            <template #default="scope">
              <div v-if="scope.row.isFirstInSlot" class="slot-cell-content">
                <span>{{ scope.row.slotNum }}</span>
                <el-button 
                  size="small" 
                  type="primary" 
                  @click="$emit('showCreateDialog', activeDevice, 'slot', scope.row.slotNum)"
                  class="create-button"
                  :disabled="!activeDevice"
                >
                  {{ t('common.create') }}
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.instanceNameLabel')" width="110">
            <template #default="scope">
              {{ formatInstanceName(scope.row.name) }}
            </template>
          </el-table-column>
          <el-table-column prop="ip" :label="t('common.ipAddress')" width="130"></el-table-column>
          <el-table-column :label="t('common.systemImage')" width="180">
            <template #default="scope">
              <div 
                style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                :title="getImageDisplayName(scope.row.image)"
              >
                {{ getImageDisplayName(scope.row.image) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.createTime')" width="160">
            <template #default="scope">
              {{ scope.row.created ? new Date(scope.row.created).toLocaleString('zh-CN') : scope.row.createTime }}
            </template>
          </el-table-column>
          <el-table-column prop="status" :label="t('common.statusLabel')" width="90">
            <template #default="scope">
              <el-tag 
                :type="scope.row.status === 'running' ? 'success' : scope.row.status === 'restarting' ? 'warning' : 'info'"
                size="small"
                class="status-tag-normal"
              >
                {{ scope.row.status === 'running' ? t('common.running') : (scope.row.status === 'shutdown' || scope.row.status === 'exited') ? t('common.shutdownStatus') : t('common.restartingStatus') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="modelName" :label="t('common.modelLabel')" width="100"></el-table-column>
          <el-table-column :label="t('common.operationLabel')" width="300" fixed="right">
            <template #default="scope">
              <el-space size="small">
                <!-- 开机/关机按钮 -->
                <el-button 
                  size="small" 
                  :type="scope.row.status === 'running' ? 'warning' : 'success'" 
                  @click="() => {
                    const action = scope.row.status === 'running' ? '关机' : '开机';
                    ElMessageBox.confirm(`确定要${action}实例 ${scope.row.name}吗？`, '操作确认', {
                      confirmButtonText: '确定',
                      cancelButtonText: '取消',
                      type: 'warning'
                    }).then(async () => {
                      try {
                        if (scope.row.status === 'running') {
                          await $emit('stopContainer', activeDevice, scope.row.name);
                        } else {
                          await $emit('startContainer', activeDevice, scope.row.name);
                        }
                        ElMessage.success(`${action}成功`);
                        await $emit('fetchAndroidContainers', activeDevice, true);
                      } catch (error) {
                        ElMessage.error(`${action}失败：${error.message}`);
                      }
                    }).catch(() => {
                      // 用户取消操作
                    });
                  }"
                >
                  {{ scope.row.status === 'running' ? t('common.shutdownStatus') : t('common.booting').replace('...', '') }}
                </el-button>
                
                <!-- 更多操作下拉菜单 -->
                <el-dropdown>
                  <el-button size="small">
                    {{ t('common.moreActions') }} <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="() => { $emit('startProjection', { ip: scope.row.deviceIp }, scope.row) }">{{ t('common.openProjection') }}</el-dropdown-item>

                      <el-dropdown-item @click="$emit('showUpdateImageDialog', scope.row)">{{ t('common.updateImage') }}</el-dropdown-item>
                      <el-dropdown-item divided @click="$emit('handleDeleteContainer', scope.row)">{{ t('common.delete') }}</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                

              </el-space>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <!-- 主机标签页 -->
      <div v-if="currentRightTab === 'host'" class="host-info-container">
        <el-card shadow="hover" class="host-info-card">
          <template #header>
            <div class="card-header">
              <span>{{ t('common.hostInfo') }}</span>
            </div>
          </template>
          
          <div v-if="activeDevice" class="host-info-content">
            <!-- 基本信息 -->
            <el-descriptions :title="t('common.basicInfo')" :column="2" border>
              <el-descriptions-item :label="t('common.hostName')">{{ activeDevice.name }}</el-descriptions-item>
              <el-descriptions-item :label="t('common.hostIP')">{{ activeDevice.ip }}</el-descriptions-item>
              <el-descriptions-item :label="t('common.deviceID')">{{ (activeDevice.version === 'v3' && v3DeviceInfo.originalData?.deviceId) ? v3DeviceInfo.originalData.deviceId : activeDevice.id }}</el-descriptions-item>
              <el-descriptions-item :label="t('common.deviceVersion')">{{ activeDevice.version }}</el-descriptions-item>
            </el-descriptions>
            
            <!-- V3设备额外信息 -->
            <div v-if="activeDevice.version === 'v3'" class="v3-info-section">
              <el-descriptions :title="t('common.v3DeviceInfo')" :column="2" border>
                <el-descriptions-item :label="t('common.hostFirmwareVersion')">{{ v3DeviceInfo.sdkVersion || t('common.loading') }}</el-descriptions-item>
                <el-descriptions-item :label="t('common.deviceModel')">{{ v3DeviceInfo.deviceModel || t('common.loading') }}</el-descriptions-item>
                <el-descriptions-item :label="t('common.cpuTemp')">{{ v3DeviceInfo.originalData?.cputemp || t('common.loading') }}°C</el-descriptions-item>
                <el-descriptions-item :label="t('common.cpuLoad')">{{ v3DeviceInfo.originalData?.cpuload || t('common.loading') }}</el-descriptions-item>
                <el-descriptions-item :label="t('common.memoryUsage')">{{ v3DeviceInfo.originalData?.memuse || 0 }}/{{ v3DeviceInfo.originalData?.memtotal || 0 }} MB</el-descriptions-item>
                <el-descriptions-item :label="t('common.storageUsage')">{{ v3DeviceInfo.originalData?.mmcuse || 0 }}/{{ v3DeviceInfo.originalData?.mmctotal || 0 }} MB</el-descriptions-item>
                <el-descriptions-item :label="t('common.ipAddress')">{{ v3DeviceInfo.originalData?.ip || t('common.loading') }}</el-descriptions-item>
                <el-descriptions-item :label="t('common.macAddress')">{{ v3DeviceInfo.originalData?.hwaddr || t('common.loading') }}</el-descriptions-item>
              </el-descriptions>
              
              <!-- 密码管理功能 -->
              <div class="password-section" style="margin-top: 20px;">
                <el-divider>{{ t('common.passwordManagement') }}</el-divider>
                <el-space>
                  <el-button 
                    type="primary" 
                    size="small" 
                    @click="$emit('showPasswordDialog')"
                    :disabled="!activeDevice"
                  >
                    {{ t('common.setPassword') }}
                  </el-button>
                  <el-button 
                    type="warning" 
                    size="small" 
                    @click="$emit('handleClosePassword')"
                    :disabled="!activeDevice"
                  >
                    {{ t('common.closePassword') }}
                  </el-button>
                  <span class="password-hint" style="color: #606266; font-size: 12px;">
                    {{ t('common.passwordHint') }}
                  </span>
                </el-space>
              </div>
              
              <!-- SDK升级功能 -->
              <div v-if="showUpgradeButton && !isViewingDeviceDetails" class="upgrade-section">
                <el-divider>{{ t('common.sdkUpgrade') }}</el-divider>
                <el-space>
                  <span>{{ t('common.currentSDKVersion') }}: {{ v3LatestInfo.originalData?.currentVersion }}</span>
                  <span>{{ t('common.latestSDKVersion') }}: {{ v3LatestInfo.originalData?.latestVersion }}</span>
                  <el-button 
                    type="primary" 
                    @click="$emit('upgradeSDK')"
                    :loading="upgrading"
                    :disabled="upgrading"
                  >
                    <template v-if="upgrading && upgradeProgress > 0">
                      {{ t('common.upgrading') }} ({{ upgradeProgress.toFixed(1) }}%)
                    </template>
                    <template v-else>
                      {{ t('common.upgradeSDK') }}
                    </template>
                  </el-button>
                </el-space>
              </div>

          </div>
          </div>
          
          <div v-else class="no-device-selected">
            <el-empty :description="t('common.selectDeviceFirst')" :image-size="100"></el-empty>
          </div>
        </el-card>
      </div>
      
      <!-- 网络标签页 -->
      <div v-if="currentRightTab === 'network'" class="network-info-container">
        <el-card shadow="hover" class="network-info-card">
          <template #header>
            <div class="card-header">
              <span>{{ t('common.dockerNetworkList') }}</span>
              <el-space>
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="$emit('showAddMacvlanDialog')" 
                  :disabled="!activeDevice"
                >
                  <el-icon><Plus /></el-icon> {{ t('common.addMacvlanNetwork') }}
                </el-button>
              </el-space>
            </div>
          </template>
          
          <div v-if="activeDevice" class="network-info-content">
            <!-- 加载状态 -->
            <div v-if="dockerNetworksLoading" class="network-loading">
              <el-skeleton :rows="5" animated></el-skeleton>
            </div>
            
            <!-- 错误信息 -->
            <div v-else-if="dockerNetworksError" class="network-error">
              <el-alert
                :title="t('common.loadFailed')"
                :description="dockerNetworksError"
                type="error"
                show-icon
              ></el-alert>
            </div>
            
            <!-- 网络列表 -->
          <div v-else-if="dockerNetworks.length > 0" class="table-container">
            <el-table :data="dockerNetworks" stripe size="small" class="network-table">
                <el-table-column prop="Name" :label="t('common.networkName')" width="180"></el-table-column>
                <el-table-column prop="ID" :label="t('common.networkID')" width="120"></el-table-column>
                <el-table-column :label="t('common.subnet')" width="150">
                  <template #default="scope">
                    <span v-if="scope.row.IPAM?.Config?.[0]?.Subnet">{{ scope.row.IPAM.Config[0].Subnet }}</span>
                    <span v-else>无</span>
                  </template>
                </el-table-column>
                <el-table-column prop="Driver" :label="t('common.driver')" width="100"></el-table-column>
                <el-table-column prop="Scope" :label="t('common.scope')" width="100"></el-table-column>
                <el-table-column :label="t('common.gateway')" width="150">
                  <template #default="scope">
                    <span v-if="scope.row.IPAM?.Config?.[0]?.Gateway">{{ scope.row.IPAM.Config[0].Gateway }}</span>
                    <span v-else>无</span>
                  </template>
                </el-table-column>
                <el-table-column :label="t('common.ipRange')" width="150">
                  <template #default="scope">
                    <span v-if="scope.row.IPAM?.Config?.[0]?.IPRange">{{ scope.row.IPAM.Config[0].IPRange }}</span>
                    <span v-else>无</span>
                  </template>
                </el-table-column>
                <el-table-column prop="Containers" :label="t('common.containerCount')" width="120">
                  <template #default="scope">
                    <span>{{ Object.keys(scope.row.Containers || {}).length }}</span>
                  </template>
                </el-table-column>
                <el-table-column prop="Internal" :label="t('common.private')" width="80">
                  <template #default="scope">
                    <el-tag v-if="scope.row.Internal" type="danger" size="small">{{ t('common.yes') }}</el-tag>
                    <el-tag v-else type="success" size="small">{{ t('common.no') }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column :label="t('common.operationLabel')" width="150" fixed="right">
                  <template #default="scope">
                    <el-button
                      type="primary"
                      size="small"
                      @click="$emit('handleEditNetwork', scope.row)"
                      :disabled="Object.keys(scope.row.Containers || {}).length > 0"
                    >
                      {{ t('common.edit') }}
                    </el-button>
                    <el-button
                      type="danger"
                      size="small"
                      @click="$emit('handleDeleteNetwork', scope.row)"
                      :disabled="Object.keys(scope.row.Containers || {}).length > 0"
                    >
                      {{ t('common.delete') }}
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
            
            <!-- 空状态 -->
            <div v-else class="network-empty">
              <el-empty :description="t('common.noNetworkInfo')" :image-size="100"></el-empty>
            </div>
          </div>
          
          <div v-else class="no-device-selected">
            <el-empty :description="$t('common.selectDeviceFirst')" :image-size="100"></el-empty>
          </div>
        </el-card>
      </div>
      
      <!-- 镜像标签页 -->
      <div v-if="currentRightTab === 'image'" class="image-info-container">
        <!-- 未选择设备 -->
        <div v-if="!activeDevice" class="no-device" style="padding: 50px 20px;">
          <el-empty :description="$t('common.selectDeviceFirst')" :image-size="100"></el-empty>
        </div>
        
        <!-- 镜像分类标签和刷新按钮 -->
        <div v-else class="image-category-tabs">
          <div>
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
              <el-tabs v-model="localSelectedImageCategory" style="flex: 1;">
                    <el-tab-pane :label="$t('common.onlineImages')" name="online">
                      <!-- 在线镜像列表，按当前选择的设备筛选 -->
                      <div class="online-images-container">
                        <div v-if="fetchingImages" class="image-loading">
                          <el-skeleton :rows="5" animated></el-skeleton>
                        </div>
                        <div v-else-if="filteredImageList.length === 0" class="no-images">
                          <el-empty :description="$t('common.noMatchingImages')" :image-size="100"></el-empty>
                        </div>
                        <div v-else class="online-images-list">
                          <!-- 筛选条件 -->
                          <div class="image-filters" style="margin-bottom: 15px; padding: 0 15px;">
                            <el-row :gutter="20">
                              <el-col :span="8">
                                <el-input
                                  v-model="imageFilters.name"
                                  :placeholder="$t('common.filterImageName')"
                                  clearable
                                  size="small"
                                >
                                  <template #prefix>
                                    <el-icon><Search /></el-icon>
                                  </template>
                                </el-input>
                              </el-col>
                              <el-col :span="8">
                                <el-input
                                  v-model="imageFilters.url"
                                  :placeholder="$t('common.filterImageURL')"
                                  clearable
                                  size="small"
                                >
                                  <template #prefix>
                                    <el-icon><Link /></el-icon>
                                  </template>
                                </el-input>
                              </el-col>
                              <el-col :span="8">
                                <el-checkbox v-model="imageFilters.includeCompatible" size="small" style="margin-right: 15px;">
                                  {{ $t('common.compatibleImages') }}
                                </el-checkbox>
                                <el-button
                                  type="info"
                                  size="small"
                                  @click="resetImageFilters"
                                  style="margin-right: 10px;"
                                >
                                  {{ $t('image.resetFilter') }}
                                </el-button>
                                <el-button 
                                  type="primary" 
                                  size="small" 
                                  @click="$emit('refreshOnlineImages')" 
                                  :disabled="!activeDevice || fetchingImages"
                                  class="refresh-button"
                                >
                                  <el-icon :class="{ 'is-rotating': fetchingImages }"><Refresh /></el-icon> $t('common.refreshOnlineImages')
                                </el-button>
                              </el-col>
                            </el-row>
                          </div>

                          <!-- 镜像列表 -->
                          <div class="online-images-grid">
                            <el-card
                              v-for="image in filteredImageList"
                              :key="image.id"
                              shadow="hover"
                              class="image-card"
                            >
                              <template #header>
                                <div class="image-card-header">
                                  <span class="image-name">{{ image.name }}</span>
                                  <el-tag
                                    v-if="image.version"
                                    type="info"
                                    size="small"
                                    class="version-tag"
                                  >
                                    {{ image.version }}
                                  </el-tag>
                                </div>
                              </template>
                              <div class="image-card-content">
                                <div class="image-meta">
                                  <div class="image-meta-item">
                                    <el-icon><Calendar /></el-icon>
                                    <span>{{ new Date(image.createTime).toLocaleDateString('zh-CN') }}</span>
                                  </div>
                                  <div class="image-meta-item">
                                    <el-icon><DataAnalysis /></el-icon>
                                    <span>{{ image.size || $t('common.unknown') }}</span>
                                  </div>
                                  <div class="image-meta-item">
                                    <el-icon><StarFilled /></el-icon>
                                    <span>{{ image.starRating || 0 }}</span>
                                  </div>
                                </div>
                                <div class="image-description">
                                  {{ image.description || $t('common.noDescription') }}
                                </div>
                                <div class="image-actions">
                                  <el-button
                                    type="primary"
                                    size="small"
                                    @click="$emit('downloadImage', activeDevice, image)"
                                    :disabled="isDownloadingImage"
                                  >
                                    <el-icon><Download /></el-icon> {{ $t('common.downloadImage') }}
                                  </el-button>
                                  <el-button
                                    type="success"
                                    size="small"
                                    @click="$emit('showCreateDialog', activeDevice, 'batch', 0, image)"
                                  >
                                    <el-icon><Plus /></el-icon> {{ $t('image.batchCreate') }}
                                  </el-button>
                                </div>
                              </div>
                            </el-card>
                          </div>
                        </div>
                      </div>
                    </el-tab-pane>
                    <el-tab-pane :label="$t('common.localImages')" name="local">
                      <!-- 本地镜像列表 -->
                      <div class="local-images-container">
                        <div style="display: flex; justify-content: flex-end; align-items: center; margin: 15px 0;">
                          <el-button 
                            type="primary" 
                            size="small" 
                            @click="$emit('refreshLocalImages')" 
                            :disabled="!activeDevice || isLoadingLocalImages"
                            class="refresh-button"
                          >
                            <el-icon :class="{ 'is-rotating': isLoadingLocalImages }"><Refresh /></el-icon> $t('image.refreshLocalImages')
                          </el-button>
                        </div>
                        <div v-if="isLoadingLocalImages" class="image-loading">
                          <el-skeleton :rows="5" animated></el-skeleton>
                        </div>
                        <div v-else-if="localCachedImages.length === 0" class="no-images">
                          <el-empty :description="$t('image.noLocalCacheImages')" :image-size="100"></el-empty>
                        </div>
                        <div v-else class="local-image-items">
                          <div class="image-table-wrapper">
                            <el-table :data="localCachedImages" size="small" style="width: 100%" border>
                            <el-table-column prop="name" :label="$t('image.imageName')" width="30%"></el-table-column>
                            <el-table-column prop="url" :label="$t('image.imagePath')" width="40%"></el-table-column>
                            <el-table-column prop="size" :label="$t('image.size')" width="10%"></el-table-column>
                            <el-table-column prop="createTime" :label="$t('image.createTime')" width="15%"></el-table-column>
                            <el-table-column :label="$t('common.operation')" width="15%" fixed="right">
                              <template #default="scope">
                                <el-button 
                                  type="primary" 
                                  size="small" 
                                  @click="$emit('showCreateDialog', activeDevice, 'batch', 0, scope.row)"
                                >
                                  <el-icon><Plus /></el-icon> {{ $t('image.createFromLocal') }}
                                </el-button>
                              </template>
                            </el-table-column>
                            </el-table>
                          </div>
                        </div>
                      </div>
                    </el-tab-pane>
                    <el-tab-pane :label="$t('image.deviceImages')" name="box">
                      <!-- 盒子镜像列表（设备上存在的镜像） -->
                      <div class="box-images-container">
                        <div style="display: flex; justify-content: flex-end; align-items: center; margin: 15px 0;">
                          <el-button 
                            type="primary" 
                            size="small" 
                            @click="$emit('refreshBoxImages')" 
                            :disabled="!activeDevice || isLoadingBoxImages"
                            class="refresh-button"
                          >
                            <el-icon :class="{ 'is-rotating': isLoadingBoxImages }"><Refresh /></el-icon> $t('image.refreshDeviceImages')
                          </el-button>
                        </div>
                        <div v-if="isLoadingBoxImages" class="image-loading">
                          <el-skeleton :rows="5" animated></el-skeleton>
                        </div>
                        <div v-else-if="boxImages.length === 0" class="no-images">
                          <el-empty :description="$t('image.noDeviceImages')" :image-size="100"></el-empty>
                        </div>
                        <div v-else class="box-image-items">
                          <div class="image-table-wrapper">
                            <el-table :data="boxImages" size="small" style="width: 100%" border>
                            <el-table-column prop="name" :label="$t('image.imageName')" width="30%"></el-table-column>
                            <el-table-column prop="url" :label="$t('image.imageURL')" width="40%"></el-table-column>
                            <el-table-column prop="size" :label="$t('image.size')" width="10%"></el-table-column>
                            <el-table-column prop="createTime" :label="$t('image.createTime')" width="15%"></el-table-column>
                            <el-table-column :label="$t('common.operation')" width="15%" fixed="right">
                              <template #default="scope">
                                <el-button 
                                  type="primary" 
                                  size="small" 
                                  @click="$emit('showCreateDialog', activeDevice, 'batch', 0, scope.row)"
                                >
                                  <el-icon><Plus /></el-icon> {{ $t('image.createFromDevice') }}
                                </el-button>
                              </template>
                            </el-table-column>
                            </el-table>
                          </div>
                        </div>
                      </div>
                    </el-tab-pane>
                  </el-tabs>
                </div>
              </div>
            </div>
            
            <!-- 下载进度条 -->
            <div v-if="isDownloadingImage" class="download-progress-container" style="padding: 0 20px 20px;">
              <el-progress 
                :percentage="downloadProgress" 
                :stroke-width="20" 
                :show-text="true"
                status="success"
              ></el-progress>
              <div class="download-progress-text">
                {{ $t('image.downloadingImage') }}{{ currentDownloadImage?.name }}
              </div>
            </div>
            
            <!-- 上传进度已整合到任务列表 -->
      </div>
    </el-col>
  </el-row>
  
  <!-- 添加设备对话框 -->
  <AddDeviceDialog
    v-model:visible="addDeviceDialogVisible"
    :existing-device-ids="existingDeviceIds"
    @device-added="handleDeviceAdded"
    @devices-added="handleBatchAddDevices"
    @scan-complete="handleScanComplete"
  />

  <!-- 主机解绑对话框 -->
  <el-dialog
    v-model="unbindDialogVisible"
    :title="t('common.unbindHost')"
    width="600px"
    center
    destroy-on-close
    :show-close="true"
  >
    <div style="padding: 0 10px;">
      <div style="color: #F56C6C; margin-bottom: 20px; line-height: 1.5; font-weight: bold;">
        {{ t('common.unbindWarning') }}
      </div>
      
      <el-form :model="unbindForm" label-width="80px" label-position="left">
        <el-form-item :label="t('common.phoneNumber')">
          <el-input v-model="unbindForm.phone" disabled></el-input>
        </el-form-item>
        <el-form-item :label="t('common.verificationCode')">
          <div style="display: flex; width: 100%; gap: 10px;">
            <el-input v-model="unbindForm.code" :placeholder="t('common.verificationCode')"></el-input>
            <el-button type="primary" @click="handleSendCode" :disabled="isCountingDown" style="width: 120px;">
              {{ codeButtonText }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item>
           <el-button type="primary" @click="submitUnbind" :loading="unbindLoading">{{ t('common.submit') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
  </el-dialog>


  <!-- 升级进度对话框 -->
  <!-- <el-dialog
    v-model="deviceUpgradeProgressDialogVisible"
    :title="$t('image.deviceUpgrade')"
    width="400px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    center
  >
    <div v-if="deviceUpgradingDevice">
      <div class="upgrade-info">
        <p>{{ $t('image.device') }}{{ deviceUpgradingDevice.ip }}</p>
        <p>{{ $t('image.status') }}{{ deviceUpgradeStatus }}</p>
      </div>
      
      <div class="progress-container">
        <el-progress
          :percentage="deviceUpgradeProgress"
          :stroke-width="20"
          :show-text="true"
          status="success"
        ></el-progress> -->
        <!-- <div class="progress-details">
          <span>{{ formatFileSize(deviceUpgradeCurrentSize) }} / {{ formatFileSize(deviceUpgradeTotalSize) }}</span>
        </div> -->
      <!-- </div>
    </div>
  </el-dialog> -->
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, reactive, getCurrentInstance, nextTick } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Delete, Refresh, List, HelpFilled, Plus, Download, Calendar, DataAnalysis, StarFilled, ArrowDown, CloseBold, Loading, Search, Link, Edit } from '@element-plus/icons-vue';
import { Events } from '@wailsio/runtime';
import AddDeviceDialog from './AddDeviceDialog.vue';
import { SyncAuthorization, UpgradeSDK, BatchUpgradeDevices, GetPhoneVCode, UnbindHost } from '../../bindings/edgeclient/app';
import { getDevicePassword, startProjection, startProjectionBatchControl } from '../services/api.js';

// 国际化支持
const { proxy } = getCurrentInstance()
const t = (key, params) => {
  const _ = proxy.$i18n.locale
  let text = proxy.$i18n.t(key)
  if (params) {
    Object.keys(params).forEach(param => {
      text = text.replace(`{${param}}`, params[param])
    })
  }
  return text
}

// 导入防抖函数
const debounce = (fn, delay) => {
  let timer = null;
  return (...args) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      fn.apply(this, args);
    }, delay);
  };
};

// 设备扫描相关变量
const addDeviceDialogVisible = ref(false)
const existingDeviceIds = ref(new Set())

// 定义props
const props = defineProps({
  devices: {
    type: Array,
    required: true
  },
  activeDevice: {
    type: Object,
    default: null
  },
  instances: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  isBatchUpgrading: {
    type: Boolean,
    default: false
  },
  token: {
    type: String,
    default: ''
  },
  uname: {
    type: String,
    default: ''
  },
  filteredDevices: {
    type: Array,
    required: true
  },
  deviceFirmwareInfo: {
    type: Map,
    required: true
  },
  deviceVersionInfo: {
    type: Map,
    required: true
  },
  selectedHostDevices: {
    type: Array,
    default: () => []
  },
  isViewingDeviceDetails: {
    type: Boolean,
    default: false
  },
  groupedInstances: {
    type: Array,
    required: true
  },
  currentRightTab: {
    type: String,
    default: 'instance'
  },
  v3DeviceInfo: {
    type: Object,
    default: () => {}
  },
  v3LatestInfo: {
    type: Object,
    default: () => {}
  },
  showUpgradeButton: {
    type: Boolean,
    default: false
  },
  upgrading: {
    type: Boolean,
    default: false
  },
  upgradeProgress: {
    type: Number,
    default: 0
  },
  dockerNetworks: {
    type: Array,
    required: true
  },
  dockerNetworksLoading: {
    type: Boolean,
    default: false
  },
  dockerNetworksError: {
    type: String,
    default: ''
  },
  fetchingImages: {
    type: Boolean,
    default: false
  },
  selectedImageCategory: {
    type: String,
    default: 'online'
  },
  filteredImageList: {
    type: Array,
    required: true
  },
  imageFilters: {
    type: Object,
    default: () => ({ name: '', url: '', includeCompatible: false })
  },
  isLoadingLocalImages: {
    type: Boolean,
    default: false
  },
  localCachedImages: {
    type: Array,
    default: () => []
  },
  isLoadingBoxImages: {
    type: Boolean,
    default: false
  },
  boxImages: {
    type: Array,
    default: () => []
  },
  isDownloadingImage: {
    type: Boolean,
    default: false
  },
  downloadProgress: {
    type: Number,
    default: 0
  },
  currentDownloadImage: {
    type: Object,
    default: null
  },
  deviceFilter: {
    type: String,
    default: 'online'
  },
  devicesStatusCache: {
    type: Map,
    required: true
  },
  deviceBindStatus: {
    type: Map,
    required: true
  },
  deviceGroups: {
    type: Array,
    default: () => ['默认分组']
  },
  deviceGroupFilter: {
    type: String,
    default: '全部'
  },
  filteredDevicesByGroup: {
    type: Array,
    required: true
  },
  deviceGroupsTree: {
    type: Array,
    default: () => []
  },
  activeTab: {
    type: String,
    default: ''
  },
  isBatchProjectionControlling: {
    type: Boolean,
    default: false
  }
});

// 定义事件
const emit = defineEmits([
  'discoverAndLoadDevices',
  'handleBatchDeleteDevices',
  'showSyncAuthDialog',
  'refchDevices',
  'handleSyncAuthorization',
  'showAddDeviceDialog',
  'clearCache',
  'handleBatchDeleteHosts',
  'logOut',
  'handleHostDeviceSelectionChange',
  'showDeviceDetails',
  'showCreateDialog',
  'handleDeleteDevice',
  'collapseRightSidebar',
  'fetchAndroidContainers',
  'handleBatchAction',
  'refreshImageList',
  'fetchDockerNetworks',
  'fetchV3DeviceInfo',
  'showPasswordDialog',
  'handleClosePassword',
  'upgradeSDK',
  'handleEditNetwork',
  'handleDeleteNetwork',
  'showAddMacvlanDialog',
  'switchImageCategory',
  'refreshOnlineImages',
  'resetImageFilters',
  'downloadImage',
  'startContainer',
  'stopContainer',
  'startProjection',
  'showUpdateImageDialog',
  'handleDeleteContainer',
  'refreshLocalImages',
  'refreshBoxImages',
  'handleAddDevice',
  'handleBindsTest',
  'showAuthDialog',
  'handleGetDeviceVersion',
  'addDeviceGroup',
  'renameDeviceGroup',
  'deleteDeviceGroup',
  'moveDeviceToGroup',
  'handleBatchAddDevices',
  'update:deviceGroupFilter'
]);

// 内部状态
const currentRightTab = ref('instance');
// 本地存储选中的设备，避免直接修改props
const localSelectedHostDevices = ref([]);
// 本地存储选中的镜像分类，避免直接修改props
const localSelectedImageCategory = ref('online');
// 本地存储设备过滤条件，避免直接修改props
const localDeviceFilter = ref(props.deviceFilter);

// IP搜索
const searchIP = ref('');

// 计算属性：过滤后的设备列表（包含IP搜索）
const displayDevices = computed(() => {
  if (!searchIP.value) {
    return props.filteredDevicesByGroup;
  }
  const query = searchIP.value.trim().toLowerCase();
  return props.filteredDevicesByGroup.filter(device => 
    device.ip && device.ip.toLowerCase().includes(query)
  );
});

// 批量升级相关状态
const isBatchUpgrading = ref(false)
const deviceUpgradeProgress = ref(0) // 升级进度百分比

// 表格引用
const deviceTableRef = ref(null);

// 升级进度相关状态
const deviceUpgradeProgressDialogVisible = ref(false) // 升级进度对话框可见性
const deviceUpgradingDevice = ref(null) // 当前升级设备
const deviceUpgradeCurrentSize = ref(0) // 当前升级大小
const deviceUpgradeTotalSize = ref(0) // 升级总大小
const deviceUpgradeStatus = ref('') // 升级状态

// 监听props.deviceFilter的变化，更新本地变量
watch(
  () => props.deviceFilter,
  (newValue) => {
    localDeviceFilter.value = newValue;
  }
);

// 监听本地设备过滤条件的变化，通知父组件
watch(
  () => localDeviceFilter.value,
  (newValue) => {
    emit('update:deviceFilter', newValue);
  }
);

// 本地存储分组过滤条件，避免直接修改props
const localGroupFilter = ref(props.deviceGroupFilter);

// 监听props.deviceGroupFilter的变化，更新本地变量
watch(
  () => props.deviceGroupFilter,
  (newValue) => {
    localGroupFilter.value = newValue;
  }
);

// 监听本地分组过滤条件的变化，通知父组件
watch(
  () => localGroupFilter.value,
  (newValue) => {
    emit('update:deviceGroupFilter', newValue);
  }
);

// 处理设备分组操作
const handleDeviceGroupCommand = (command, device) => {
  if (command === 'add-group') {
    // 弹出对话框让用户输入新分组名称
    ElMessageBox.prompt(t('common.enterNewGroupName'), t('common.addGroup'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputPattern: /\S+/,
      inputErrorMessage: t('common.groupNameCannotBeEmpty')
    }).then(({ value }) => {
      emit('addDeviceGroup', value)
      // 自动将当前设备移动到新分组
      emit('moveDeviceToGroup', device.id, value)
    }).catch(() => {})
  } else if (command === 'delete-group') {
    // 删除当前分组
    const currentGroup = device.group || '默认分组'
    ElMessageBox.confirm(
      `${t('common.deleteGroupMessage', {group: currentGroup})}`,
      t('common.deleteGroupConfirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    ).then(() => {
      emit('deleteDeviceGroup', currentGroup)
    }).catch(() => {})
  } else {
    // 将设备移动到指定分组
    emit('moveDeviceToGroup', device.id, command)
  }
}

// 添加新分组
const handleAddGroup = () => {
  ElMessageBox.prompt(t('common.enterNewGroupName'), t('common.addGroup'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    inputPattern: /\S+/,
    inputErrorMessage: t('common.groupNameCannotBeEmpty')
  }).then(({ value }) => {
    emit('addDeviceGroup', value)
  }).catch(() => {})
}

// 编辑分组名称
const handleEditGroup = () => {
  if (!localGroupFilter.value || localGroupFilter.value === '全部' || localGroupFilter.value === '默认分组') {
    return
  }
  
  ElMessageBox.prompt(t('common.enterNewGroupName'), t('common.editGroup'), {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputValue: localGroupFilter.value,
    inputPattern: /\S+/,
    inputErrorMessage: '分组名称不能为空'
  }).then(({ value }) => {
    if (value === localGroupFilter.value) {
      return
    }
    emit('renameDeviceGroup', localGroupFilter.value, value)
    localGroupFilter.value = '全部'
  }).catch(() => {})
}

// 删除分组
const handleDeleteGroup = () => {
  if (!localGroupFilter.value || localGroupFilter.value === '全部' || localGroupFilter.value === '默认分组') {
    return
  }
  
  ElMessageBox.confirm(
    `${t('common.deleteGroupMessage', {group: localGroupFilter.value})}`,
    t('common.deleteGroupConfirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }
  ).then(() => {
    emit('deleteDeviceGroup', localGroupFilter.value)
    localGroupFilter.value = '全部'
  }).catch(() => {})
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 处理升级进度事件
const handleUpgradeProgress = (event) => {
  try {
    console.log('收到升级进度事件:', event)
    
    const lines = event.split('\n')
    
    for (const line of lines) {
      const trimmedLine = line.trim()
      if (!trimmedLine) continue
      
      if (trimmedLine.startsWith('data:')) {
        const dataStr = trimmedLine.slice(5).trim()
        if (dataStr) {
          const progressData = JSON.parse(dataStr)
          
          if (progressData.progress !== undefined) {
            deviceUpgradeProgress.value = Math.min(100, Math.round(progressData.progress * 100))
            deviceUpgradeCurrentSize.value = progressData.current
            deviceUpgradeTotalSize.value = progressData.total
            deviceUpgradeStatus.value = `正在升级: ${deviceUpgradeProgress.value}% (${formatFileSize(progressData.current)}/${formatFileSize(progressData.total)})`
          }
        }
      } else if (trimmedLine.startsWith('event: complete')) {
        deviceUpgradeProgress.value = 100
        deviceUpgradeStatus.value = '升级完成'
        
        ElMessage.success(t('common.deviceUpgradeSuccess', { ip: deviceUpgradingDevice.value.ip }))
        
        setTimeout(() => {
          emit('handleGetDeviceVersion', deviceUpgradingDevice.value, true)
        }, 1000)
        
        // setTimeout(() => {
        //   deviceUpgradeProgressDialogVisible.value = false
        // }, 2000)
      }
    }
  } catch (error) {
    console.error('处理升级进度事件失败:', error)
  }
}

// 初始化本地状态
onMounted(() => {
  localSelectedHostDevices.value = [...props.selectedHostDevices];
  localSelectedImageCategory.value = props.selectedImageCategory;
  
  // 监听升级进度事件
  Events.On('upgrade-progress', (data) => {
    handleUpgradeProgress(data)
  })
});

// 组件卸载时移除事件监听
onBeforeUnmount(() => {
  Events.Off('upgrade-progress')
});



// 直接处理设备选择变化事件
const handleDirectSelectionChange = (selection) => {
  // 更新localSelectedHostDevices
  localSelectedHostDevices.value = selection
  
  // 同时触发watch监听器
  emit('handleHostDeviceSelectionChange', selection)
}

// 创建防抖版本的刷新函数，延迟300ms
const debouncedRefresh = debounce(() => {
  emit('refchDevices');
}, 300);
// 监听props.selectedHostDevices的变化，更新本地变量
watch(
  () => props.selectedHostDevices,
  (newValue) => {
    // 只有当props的值与本地不一致时才更新，避免死循环
    if (JSON.stringify(newValue) !== JSON.stringify(localSelectedHostDevices.value)) {
       localSelectedHostDevices.value = [...newValue];
       // 如果为空，同时清除表格选中状态
       if (newValue.length === 0 && deviceTableRef.value) {
         deviceTableRef.value.clearSelection();
       }
    }
  },
  { deep: true }
);

// 监听displayDevices变化（定时刷新会导致数据对象引用更新），恢复表格选中状态
watch(
  () => displayDevices.value,
  (newDevices) => {
    if (!deviceTableRef.value || localSelectedHostDevices.value.length === 0) return;
    const selectedIds = new Set(localSelectedHostDevices.value.map(d => d.id));
    nextTick(() => {
      newDevices.forEach(row => {
        if (selectedIds.has(row.id)) {
          deviceTableRef.value.toggleRowSelection(row, true);
        }
      });
    });
  }
);

// 监听本地选中设备的变化，通知父组件
watch(
  () => localSelectedHostDevices.value,
  (newValue) => {
    emit('handleHostDeviceSelectionChange', newValue);
  },
  { deep: true }
);

// 监听本地选中镜像分类的变化，通知父组件
watch(
  () => localSelectedImageCategory.value,
  (newValue) => {
    emit('switchImageCategory', newValue);
  }
);

// 辅助方法
const getDeviceTypeName = (deviceName) => {
  if (!deviceName) return '';
  const name = deviceName.toLowerCase();
  // 提取设备类型，如从 'q1_v3' 中提取 'q1'
  const parts = name.split('_');
  return parts[0] || '';
};

const getDeviceTypeColor = (deviceName) => {
  const deviceType = getDeviceTypeName(deviceName);
  const colorMap = {
    'q1': '#409EFF', // 蓝色
    'p1': '#67C23A', // 绿色
    'p0': '#E6A23C', // 黄色
    'v3': '#F56C6C', // 红色
    'm48': '#E6A23C', // 黄色
    'c1': '#F56C6C', // 红色
    'a1': '#909399' // 灰色
  };
  return colorMap[deviceType] || '#909399'; // 默认灰色
};

const formatInstanceName = (name) => {
  if (!name) return '';
  const parts = name.split('_');
  if (parts.length > 2) {
    return parts[parts.length - 1];
  }
  return name;
};

const formatStorage = (used, total) => {
  if (!total || total === 0) return '0/0 MB';
  if (total >= 1000) {
    const usedGB = (used / 1024).toFixed(1);
    const totalGB = (total / 1024).toFixed(1);
    return `${usedGB} GB/${totalGB} GB`;
  }
  return `${used || 0} MB/${total || 0} MB`;
};

const formatSdkVersion = (version) => {
  if (!version) return '未知';
  const match = version.match(/v(\d+\.\d+\.\d+)/);
  if (match) {
    return `v${match[1]}`;
  }
  return version;
};

const getImageDisplayName = (imageName) => {
  if (!imageName) return '未知镜像';
  // 从镜像名称中提取有意义的部分
  const parts = imageName.split('/');
  return parts[parts.length - 1] || imageName;
};

// 获取绑定状态文本
const getBindStatusText = (deviceId) => {
  if (!props.token) {
    return t('common.notLoggedIn');
  }
  const status = props.deviceBindStatus.get(deviceId) || 0;
  switch (status) {
    case 0:
      return t('common.unbound');
    case 1:
      return t('common.bound');
    case 2:
      return t('common.boundByOthers');
    default:
      return t('common.unknown');
  }
};

// 获取绑定状态类型（用于el-tag的type属性）
const getBindStatusType = (deviceId) => {
  const status = props.deviceBindStatus.get(deviceId) || 0;
  switch (status) {
    case 0:
      return 'info';
    case 1:
      return 'success';
    case 2:
      return 'warning';
    default:
      return 'danger';
  }
};

const resetImageFilters = () => {
  emit('resetImageFilters');
};

// 解绑主机相关
const unbindDialogVisible = ref(false);
const unbindForm = reactive({
  phone: '',
  code: '',
  vkey: '' // 新增 vkey 字段
});
const unbindLoading = ref(false);
const codeCountdown = ref(0);
const codeButtonText = computed(() => {
  if (codeCountdown.value > 0) {
    return `${codeCountdown.value}s${t('common.resendCode')}`
  }
  return t('common.sendCode')
});
const isCountingDown = computed(() => codeCountdown.value > 0);
let countdownTimer = null;

const handleUnbindHost = () => {
  // 1. 检查是否已登录
  if (!props.token) {
    // 未登录，弹出登录框 (参考绑定主机)
    emit('showSyncAuthDialog');
    return;
  }
  
  // 2. 检查是否勾选了设备
  if (localSelectedHostDevices.value.length === 0) {
    ElMessage.warning('请先选择要解绑的设备');
    return;
  }
  
  // 3. 检查是否有非本账号绑定的设备 (status === 2)
  // 假设 deviceBindStatus: 1=已绑定(自己), 2=被绑定(他人), 0=未绑定
  const notOwnDevices = localSelectedHostDevices.value.filter(device => {
    const status = props.deviceBindStatus.get(device.id);
    return status === 2;
  });
  
  if (notOwnDevices.length > 0) {
    ElMessage.warning('非本账号绑定，无法进行解绑操作');
    return;
  }
  
  // 4. 检查是否有已绑定的设备 (status === 1)
  const boundDevices = localSelectedHostDevices.value.filter(device => {
    const status = props.deviceBindStatus.get(device.id);
    return status === 1;
  });
  
  if (boundDevices.length === 0) {
    ElMessage.warning('请选择已绑定的设备');
    return;
  }
  
  // 初始化表单
  unbindForm.phone = props.uname || '';
  unbindForm.code = '';
  unbindForm.vkey = ''; // 重置 vkey
  unbindDialogVisible.value = true;
};

const handleSendCode = async () => {
  if (isCountingDown.value) return;
  if (!unbindForm.phone) {
    ElMessage.warning('请输入手机号码');
    return;
  }
  
  try {
    if (!props.token) {
      ElMessage.warning('请先登录');
      return;
    }
    
    console.log('正在发送验证码:', unbindForm.phone);
    const result = await GetPhoneVCode(unbindForm.phone, props.token);
    console.log('发送验证码结果:', result);
    
    if (result && result.code == 200) {
      ElMessage.success('验证码已发送');
      // 保存 vkey
      if (result.data && result.data.vkey) {
        unbindForm.vkey = result.data.vkey;
        console.log('获取到的vkey:', unbindForm.vkey);
      }
      
      // 开始倒计时
      startCountdown();
    } else {
      ElMessage.error(result.msg || result.message || '发送验证码失败');
    }
  } catch (error) {
    console.error('发送验证码失败:', error);
    ElMessage.error('发送验证码失败: ' + (error.message || '未知错误'));
  }
};

const startCountdown = () => {
  codeCountdown.value = 60;
  
  countdownTimer = setInterval(() => {
    codeCountdown.value--;
    if (codeCountdown.value <= 0) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
  }, 1000);
};

const submitUnbind = async () => {
  if (!unbindForm.code) {
    ElMessage.warning('请输入验证码');
    return;
  }
  
  if (!unbindForm.vkey) {
     // 如果没有vkey，可能是没有先获取验证码
     ElMessage.warning('请先获取验证码');
     return;
  }
  
  unbindLoading.value = true;
  // 传递选中的设备ID列表和验证码
  const deviceIds = localSelectedHostDevices.value
    .filter(d => props.deviceBindStatus.get(d.id) === 1)
    .map(d => d.id);
    
  try {
    console.log('正在解绑设备:', deviceIds);
    const result = await UnbindHost(props.token, deviceIds, unbindForm.code, unbindForm.vkey);
    console.log('解绑结果:', result);
    
    if (result && result.code == 200) {
      ElMessage.success('解绑成功');
      
      // 更新本地状态
      deviceIds.forEach(id => {
         props.deviceBindStatus.set(id, 0); // 标记为未绑定
      });
      
      unbindDialogVisible.value = false;
      localSelectedHostDevices.value = []; // 清空选择
      
      // 清除表格选中状态
      if (deviceTableRef.value) {
        deviceTableRef.value.clearSelection();
      }
    } else {
      ElMessage.error(result.msg || result.message || '解绑失败');
    }
  } catch (error) {
    console.error('解绑失败:', error);
    ElMessage.error('解绑失败: ' + (error.message || '未知错误'));
  } finally {
    unbindLoading.value = false;
  }
};

// 监听对话框关闭，清除定时器
watch(unbindDialogVisible, (val) => {
  if (!val && countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
    codeCountdown.value = 0;
  }
});

// 处理绑定主机操作
const handleBindHost = () => {
  // 1. 检查是否已登录
  if (!props.token) {
    // 未登录，弹出登录框
    emit('showSyncAuthDialog');
    return;
  }
  
  // 2. 检查是否勾选了设备
  if (localSelectedHostDevices.value.length === 0) {
    ElMessage.warning('请先选择要绑定的设备');
    return;
  }
  
  // 3. 检查是否有已绑定的设备
  const boundDevices = localSelectedHostDevices.value.filter(device => {
    const status = props.deviceBindStatus.get(device.id);
    return status === 1; // 1表示已绑定
  });
  
  if (boundDevices.length > 0) {
    // 有已绑定的设备，给出提示
    ElMessage.warning('已绑定的主机无法重复绑定');
    return;
  }
  
  // 4. 执行绑定操作
  emit('handleBindsTest');
};
const showSyncAuthDialog = () => {
  const syncAuthDeviceStr = localStorage.getItem('syncAuthCredentials')
  console.log('syncAuthDeviceStr:', syncAuthDeviceStr)
  if(syncAuthDeviceStr) {
    try {
      const syncAuthDevice = JSON.parse(syncAuthDeviceStr)
      syncAuthForm.value = {
        username: syncAuthDevice.username || '',
        password: syncAuthDevice.password || '',
        saveCredentials: syncAuthDevice.saveCredentials || false
      }
    } catch (error) {
      console.error('Failed to parse syncAuthCredentials:', error)
      syncAuthForm.value = {
        username: '',
        password: '',
        saveCredentials: false
      }
    }
  } else {
      syncAuthForm.value = {
      username: '',
      password: '',
      saveCredentials: false
    }
  }
  syncAuthDialogVisible.value = true
}
const handleSyncAuthorization = async () => {
  const currentToken = token.value 
  if (!currentToken) {
    ElMessage.warning('请先获取授权token');
    return;
  }
  
  const devices = filteredDevices.value;
  if (devices.length === 0) {
    ElMessage.warning('请先添加设备');
    return;
  }
  
  const deviceIPs = devices.map(device => device.ip);
  console.log('deviceIPs:', deviceIPs)
  
  try {
    let successCount = 0;
    for (const deviceIP of deviceIPs) {
      try {
        const result = await SyncAuthorization(currentToken, deviceIP);
        if (result.success) {
          successCount++;
        } else {
          console.error(`设备 ${deviceIP} 同步授权失败:`, result.message);
        }
      } catch (deviceError) {
        console.error(`设备 ${deviceIP} 同步授权出错:`, deviceError);
      }
    }
    
    if (successCount === deviceIPs.length) {
      ElMessage.success(t('common.syncAuthCompletedForAllDevices', { count: successCount }));
    } else if (successCount > 0) {
      ElMessage.warning(t('common.syncAuthPartialSuccess', { count: successCount }));
    } else {
      ElMessage.error(t('common.syncAuthAllFailed'));
    }
  } catch (error) {
    console.error('同步授权失败:', error);
    ElMessage.error('同步授权失败');
  }
}
const showAddDeviceDialog = () => {
  // 更新已存在设备ID列表
  existingDeviceIds.value = new Set(props.devices.map(d => d.id))
  // 显示对话框，对话框打开后会自动开始扫描
  addDeviceDialogVisible.value = true
}

const handleDeviceAdded = (device) => {
  // 通知父组件添加设备
  emit('handleAddDevice', device)
}

const handleBatchAddDevices = (devices) => {
  emit('handleBatchAddDevices', devices)
}

const handleScanComplete = () => {
  // 扫描完成
}

const startBatchUpgrade = async () => {
  try {
    // 使用用户在界面上选择的设备，而不是重新计算所有设备
    let devicesToUpgrade = [...localSelectedHostDevices.value]
    
    // 过滤出需要升级的设备
    devicesToUpgrade = devicesToUpgrade.filter(device => {
      // 只处理在线设备
      if (props.devicesStatusCache.get(device.id) === 'online' && device.version === 'v3') {
        // 检查是否需要升级
        const versionInfo = props.deviceVersionInfo.get(device.id)
        if (!versionInfo) return false
        const current = parseFloat(versionInfo.currentVersion)
        const latest = parseFloat(versionInfo.latestVersion)
        return !isNaN(current) && !isNaN(latest) && current < latest
      }
      return false
    })
    
    if (devicesToUpgrade.length === 0) {
      ElMessage.info('没有需要升级的设备')
      return
    }
    
    await ElMessageBox.confirm(`确定要批量升级 ${devicesToUpgrade.length} 个设备吗？升级过程可能需要较长时间。`, '批量升级确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 🔧 构建批量升级请求参数
    const upgradeRequests = devicesToUpgrade.map(device => {
      const versionInfo = props.deviceVersionInfo.get(device.id)
      const password = device.password || getDevicePassword(device.ip) || ''
      
      return {
        deviceIP: device.ip,
        latestVersion: String(versionInfo.latestVersion), // 转换为字符串
        password: password
      }
    })
    
    console.log('[批量升级] 准备升级设备:', upgradeRequests)
    
    // 显示加载状态
    isBatchUpgrading.value = true
    ElMessage({ message: '正在批量升级设备...', type: 'info' })
    
    // 🔧 调用后端批量升级接口
    const result = await BatchUpgradeDevices(upgradeRequests)
    
    console.log('[批量升级] 升级结果:', result)
    
    if (result.success) {
      ElMessage.success(result.message)
      
      // 显示详细结果
      if (result.results && result.results.length > 0) {
        let failedDevices = result.results.filter(r => !r.success)
        if (failedDevices.length > 0) {
          console.log('[批量升级] 失败设备详情:', failedDevices)
          
          // 🔧 处理认证失败的设备
          let authFailedDevices = failedDevices.filter(r => r.errorType === 'auth_required')
          if (authFailedDevices.length > 0) {
            ElMessage.warning(`${authFailedDevices.length} 个设备认证失败，请检查设备密码`)
          }
        }
      }
      
      // 🔧 不要手动调用handleGetDeviceVersion,让后端心跳自动更新
      // 后端已经重置了LastAPICheckTime,TCP Ping会在1-2秒内自动刷新
      // console.log('[批量升级] 后端会在1-2秒内自动刷新设备状态')
      
    } else {
      ElMessage.error(result.message || '批量升级失败')
    }
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('启动批量升级失败:', error)
      ElMessage.error(`启动批量升级失败: ${error.message}`)
    }
  } finally {
    isBatchUpgrading.value = false
  }
}


const clearCache = async () => {
  try {
    await ElMessageBox.confirm('确定要清理所有客户端数据并重新加载吗？这将清除所有设备信息、密码和镜像缓存。', '清理客户端数据确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 显示加载状态
    ElMessage({ message: '正在清理客户端数据...', type: 'info' })
    
    // 1. 清理localStorage缓存
    localStorage.removeItem('deviceCache')
    localStorage.removeItem('devicePasswords')
    localStorage.removeItem('mytos_image_list')
    localStorage.removeItem('mytos_image_list_last_update')
    
    // 2. 清理内存缓存（Map类型）
    deviceCloudMachinesCache.value.clear()
    cloudMachineLoadingState.value.clear()
    devicesLastUpdateTime.value.clear()
    devicesStatusCache.value.clear()
    deviceVersionInfo.value.clear()
    onlineImagesByModel.value.clear()
    imageDownloadStatus.value.clear()
    imageUploadStatus.value.clear()
    
    // 3. 清理内存缓存（ref类型）
    devices.value = []
    activeDevice.value = null
    instances.value = []
    allInstances.value = []
    cloudMachines.value = []
    selectedCloudMachines.value = []
    phoneModels.value = []
    imageList.value = []
    filteredImageList.value = []
    sharedFiles.value = []
    selectedFiles.value = []
    dockerNetworks.value = []
    localCachedImages.value = []
    
    // 4. 清理api.js中的内存缓存
    containersMemoryCache.clear()
    
    // 5. 重新加载所有数据
    await refreshData()
    
    ElMessage({ message: '客户端数据清理完成，数据已重新加载', type: 'success' })
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清理客户端数据失败:', error)
      ElMessage.error('清理客户端数据失败: ' + error.message)
    }
  }
}
const handleHostDeviceSelectionChange = async (selection) => {
  selectedHostDevices.value = selection
  console.log('设备选择变化:', selection)
  
  // 如果只选择了一个设备，设置为当前活动设备
  if (selection.length === 1) {
    await handleDeviceSelect(selection[0])
  }
}
const handleUpgradeDevice = async (device, isBatchUpgrade = false) => {
  try {
    // 对于批量升级，不显示单个设备的确认对话框
    if (!isBatchUpgrade) {
      await ElMessageBox.confirm(`确定要升级设备 ${device.ip} 吗？升级过程可能需要几分钟时间。`, '升级确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    }

    console.log('开始升级设备:', device.ip)
    
    // 获取升级相关信息
    const versionInfo = props.deviceVersionInfo.get(device.id)
    if (!versionInfo || !versionInfo.latestVersion) {
      ElMessage.error(`设备 ${device.ip} 升级失败: 未获取到最新版本信息`)
      return false
    }
    
    console.log('设备版本信息:', versionInfo)
    
    // 先检查是否需要认证，获取保存的密码
    let password = device.password
    if (!password) {
      password = getDevicePassword(device.ip)
      console.log('从本地存储获取密码:', password ? '已获取到密码' : '未找到密码')
    }
    
    // 🔧 统一使用批量升级接口(单设备也走批量逻辑)
    const upgradeRequests = [{
      deviceIP: device.ip,
      latestVersion: String(versionInfo.latestVersion), // 转换为字符串
      password: password || ''
    }]
    
    console.log('[单设备升级] 准备升级:', upgradeRequests)
    
    // 调用批量升级接口
    const result = await BatchUpgradeDevices(upgradeRequests)

    console.log('[单设备升级] 升级结果:', result)
    
    if (result.success) {
      ElMessage.success(`设备 ${device.ip} 升级成功`)
      
      // 🔧 不要手动刷新版本,让后端心跳自动更新(避免版本闪烁)
      console.log('[单设备升级] 后端会在1-2秒内自动刷新设备状态')
      return true
      
    } else {
      // 检查单个设备的升级结果
      const deviceResult = result.results && result.results[0]
      
      // 检查是否是认证错误
      if (deviceResult && deviceResult.errorType === 'auth_required') {
        console.log('升级设备认证失败，需要显示认证对话框')
        // 检查是否保存了密码
        const savedPassword = getDevicePassword(device.ip)
        if (!savedPassword) {
          // 没有保存的密码，显示认证对话框
          emit('showAuthDialog', device)
          ElMessage.warning('设备需要认证，请输入设备密码')
        } else {
          // 有保存的密码但认证失败
          ElMessage.error('设备密码错误，请重新输入')
          emit('showAuthDialog', device)
        }
        return false
      }
      
      ElMessage.error(`设备 ${device.ip} 升级失败: ${deviceResult?.message || result.message}`)
      return false
    }
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('升级设备失败:', error)
      ElMessage.error(`设备 ${device.ip} 升级失败: ${error.message}`)
      //deviceUpgradeProgressDialogVisible.value = false
      return false
    }
    // 用户取消了升级
    //deviceUpgradeProgressDialogVisible.value = false
    return 'cancel'
  }
}

const showCreateDialog = async (device, mode, slot = 0, localImage = null) => {
  createDevice.value = device
  createMode.value = mode
  currentSlot.value = slot
  
  // 使用上一次的选择初始化表单
  if (mode === 'batch') {
    createForm.value = {
      name: 'T000',
      modelName: 'random', // 默认随机机型
      count: 1,
      startSlot: 1,
      imageSelect: lastImageSelection.value.imageSelect || '',
      customImageUrl: lastImageSelection.value.customImageUrl || '',
      imageCategory: lastImageSelection.value.imageCategory || 'online',
      localImageUrl: lastImageSelection.value.localImageUrl || '',
      imageSource: lastImageSelection.value.imageSource || 'pc',
      cacheToLocal: false,
      networkMode: 'bridge',
      ipaddr: '',
      resolution: 'default',
      customResolution: {
        width: '720',
        height: '1280',
        dpi: '320'
      },
      sandboxSize: 28,
      dns: '223.5.5.5',
      customDns: '',
      countryCode: 'CN', // 默认为中国
      // S5代理设置
      s5Type: '0',
      s5IP: '',
      s5Port: '',
      s5User: '',
      s5Password: '',
      enableMagisk: false,
      enableGMS: false
    }
  } else {
    createForm.value = {
      name: 'T000',
      modelName: 'random', // 默认随机机型
      count: 1,
      startSlot: slot,
      imageSelect: lastImageSelection.value.imageSelect || '',
      customImageUrl: lastImageSelection.value.customImageUrl || '',
      imageCategory: lastImageSelection.value.imageCategory || 'online',
      localImageUrl: lastImageSelection.value.localImageUrl || '',
      imageSource: lastImageSelection.value.imageSource || 'pc',
      cacheToLocal: false,
      networkMode: 'bridge',
      ipaddr: '',
      resolution: 'default',
      customResolution: {
        width: '720',
        height: '1280',
        dpi: '320'
      },
      sandboxSize: 28,
      dns: '223.5.5.5',
      customDns: '',
      countryCode: 'CN', // 默认为中国
      // S5代理设置
      s5Type: '0',
      s5IP: '',
      s5Port: '',
      s5User: '',
      s5Password: '',
      enableMagisk: false,
      enableGMS: false
    }
  }
  
  // 获取设备类型
  const deviceType = device.name || 'C1'
  
  // 先显示创建对话框
  createDialogVisible.value = true
  
  // 加载本地镜像列表，确保用户可以在对话框中选择本地镜像
  await fetchLocalCachedImages()
  
  // 如果提供了本地镜像，直接设置表单
  if (localImage) {
    createForm.value.imageCategory = 'local'
    createForm.value.localImageUrl = localImage.url
    createForm.value.imageSource = 'local'
    console.log('从本地镜像创建云机，镜像URL:', localImage.url)
    
    // 如果是V3设备，仍然需要获取型号列表
    if (device.version === 'v3') {
      await getV3PhoneModels(device.ip)
    }
    
    return
  }
  
  // 如果是V0-V2设备，显示SDK加载蒙版并执行加载流程
  if (device.version !== 'v3') {
    try {
      // 显示SDK加载蒙版
      sdkLoadingVisible.value = true
      sdkLoadingMessage.value = '加载MYT SDK中'
      
      // 调用后端创建V0-V2设备的SDK
      await createV0V2Device(device)
      
      // 更新提示信息
      sdkLoadingMessage.value = '加载镜像列表中'
      
      // 获取镜像列表
      await fetchImageList(deviceType)
      
      // 设置镜像选择：始终默认选第一个镜像
      if (filteredImageList.value.length > 0) {
        createForm.value.imageSelect = filteredImageList.value[0].url
        createForm.value.imageCategory = 'online'
        createForm.value.localImageUrl = ''
        createForm.value.imageSource = 'pc'
      }
      
      // 隐藏SDK加载蒙版
      sdkLoadingVisible.value = false
    } catch (error) {
      console.error('加载MYT SDK失败:', error)
      ElMessage.error(`加载MYT SDK失败：${error.message}`)
      // 隐藏SDK加载蒙版
      sdkLoadingVisible.value = false
      // 关闭创建对话框
      createDialogVisible.value = false
    }
  } else {
    // V3设备正常流程
    // 如果是V3设备，获取型号列表
    await getV3PhoneModels(device.ip)
    
    // 获取机型国家列表
    await getCountryList(device.ip)
    
    // 获取镜像列表
    await fetchImageList(deviceType)
    
    // 设置镜像选择：始终默认选第一个镜像
    if (filteredImageList.value.length > 0) {
      createForm.value.imageSelect = filteredImageList.value[0].url
      createForm.value.imageCategory = 'online'
      createForm.value.localImageUrl = ''
      createForm.value.imageSource = 'pc'
    }
  }
}
const handleDeleteDevice = async (device) => {
  try {
    await ElMessageBox.confirm(`确定要删除设备 ${device.ip} 吗？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger'
    })
    
    // 从设备列表中移除
    const index = devices.value.findIndex(d => d.ip === device.ip)
    if (index !== -1) {
      devices.value.splice(index, 1)
      // 保存到本地存储
      saveDevicesToLocalStorage()
      // 清除设备相关的缓存
      deviceCloudMachinesCache.value.delete(device.ip)
      // 如果删除的是当前激活的设备，清空激活状态
      if (activeDevice.value && activeDevice.value.ip === device.ip) {
        activeDevice.value = null
        instances.value = []
        allInstances.value = []
        cloudMachines.value = []
      }
      ElMessage.success('设备删除成功')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除设备失败:', error)
      ElMessage.error('删除设备失败')
    }
  }
}
const collapseRightSidebar = () => {
  console.log('关闭设备详情弹窗');
  deviceDetailsDialogVisible.value = false;
  // 退出查看详情模式，显示勾选框
  isViewingDeviceDetails.value = false;
}
const fetchAndroidContainers = async (device, isUserInitiated = false) => {
  if (!device) return
  
  console.log('fetchAndroidContainers called with device:', device.ip, 'version:', device.version, 'isUserInitiated:', isUserInitiated)
  
  // 只有用户主动刷新或设备未加载过时才显示加载状态
  if (isUserInitiated || !deviceCloudMachinesCache.value.has(device.ip)) {
    cloudMachineLoadingState.value.set(device.ip, true)
  }
  
  try {
    // 使用认证重试逻辑
    await authRetry(device, async (password = null) => {
      // 设置10秒超时
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 10000);
      
      let containers = null;
      let isV3Success = false;
      
      try {
        // 尝试获取容器列表
        containers = await getContainers(device, password);
        clearTimeout(timeoutId);
        
        console.log('getContainers returned for device', device.ip, containers)
        
        // 检查是否认证失败
        if (containers.code === 61 && containers.message === 'Authentication Failed') {
          console.log('认证失败，需要重新认证');
          throw new Error('Authentication Failed');
        }
        
        // 根据设备版本处理不同的容器数据格式
        let processedContainers = []
        let allRawContainers = []
        
        if (device.version === 'v3') {
          // V3 API: 正确处理API响应格式 {code:0, message:"OK", data:{count:18, list:[...]}}
          let rawContainers = []
          if (containers.code === 0) {
            // API调用成功，不管云机数量多少，都将设备标记为在线
            isV3Success = true;
            // 检查是否有list字段
            if (containers.data && containers.data.list) {
              rawContainers = containers.data.list
            } else if (containers.list) {
              // 兼容旧格式
              rawContainers = containers.list
            } else {
              // 云机数量为0，list字段不存在，使用空数组
              rawContainers = []
              console.log('V3 API返回成功，但云机数量为0')
            }
          } else {
            console.log('V3 API返回错误，尝试回退到Docker API')
            // V3 API失败，抛出错误触发回退逻辑
            throw new Error('V3 API返回错误')
          }
          
          // 保存所有原始容器数据，用于备份列表
          allRawContainers = rawContainers.map(container => ({
            ...container,
            status: container.status === 'running' ? 'running' : 'shutdown',
            deviceIp: device.ip
          }))
          
          // 转换容器数据并按坑位分组，每个坑位只保留一个容器，优先保留running状态的容器
          const containersBySlot = new Map()
          
          rawContainers.forEach(container => {
            const slotNum = container.indexNum
            if (slotNum) {
              const existingContainer = containersBySlot.get(slotNum)
              const processedContainer = {
                ...container,
                // 转换V3 API状态为前端期望的状态格式
                status: container.status === 'running' ? 'running' : 'shutdown',
                deviceIp: device.ip,
                networkMode: container.networkMode || container.network || container.NetworkMode || container?.HostConfig?.NetworkMode
              }
              
              if (!existingContainer) {
                // 该坑位还没有容器，直接添加
                containersBySlot.set(slotNum, processedContainer)
              } else if (processedContainer.status === 'running' && existingContainer.status !== 'running') {
              // 新容器是running状态，而现有容器不是，替换
              containersBySlot.set(slotNum, processedContainer)
            }
          }
        })
        
        // 将Map转换为数组
        processedContainers = Array.from(containersBySlot.values())
        console.log('V3 API processedContainers:', processedContainers)
        
        // 转换为云机格式，添加screenshot等属性
        const deviceCloudMachines = processedContainers.map(inst => {
          // 提取端口并生成截图URL，只在创建云机对象时执行一次，后续直接从screenshot属性获取
          const screenshotUrl = getCloudMachineScreenshotUrl(device, inst);
          return {
            id: inst.name,
            name: inst.name,
            status: inst.status,
            screenshot: screenshotUrl, // 缓存截图URL，包含端口信息
            screenshotData: null, // 存储截图URL
            screenshotError: false, // 存储截图加载状态
            hasLoadedOnce: false, // 标记是否已成功加载过至少一次
            ip: inst.ip,
            modelName: inst.modelName,
            deviceIp: device.ip,
            indexNum: inst.indexNum,
            networkMode: inst.networkMode || inst.network || inst.NetworkMode || inst?.HostConfig?.NetworkMode
          };
        });
        
        // ⚠️ 不再手动设置设备在线状态，由心跳检测系统统一管理
        // devicesStatusCache.value.set(device.id, 'online');
      devicesLastUpdateTime.value.set(device.id, Date.now());
      
      // 缓存设备云机列表
      deviceCloudMachinesCache.value.set(device.ip, deviceCloudMachines);
      
      // 如果是当前选中设备，更新全局云机列表
      if (activeDevice.value && activeDevice.value.ip === device.ip) {
        instances.value = processedContainers;
        allInstances.value = allRawContainers; // 保存所有容器，用于备份列表
        updateCloudMachines();
      }
      
      console.log('Device', device.ip, 'has', deviceCloudMachines.length, 'cloud machines in cache')
      return true;
      }
    } catch (v3Error) {
      console.log(`V3 API查询失败，尝试回退到Docker API: ${v3Error.message}`);
      // 清除之前的超时定时器，避免冲突
      clearTimeout(timeoutId);
      
      // 如果是认证错误，直接抛出，让authRetry函数处理
      if (v3Error.message === 'Authentication Failed') {
        console.log('V3 API认证失败，直接抛出错误让authRetry处理');
        throw v3Error;
      }
    }
    
    // V0-V2 Docker API或V3回退: 需要转换数据格式
    try {
      // 重新设置超时，回退到Docker API
      const dockerController = new AbortController();
      const dockerTimeoutId = setTimeout(() => dockerController.abort(), 10000);
      
      // 调用getContainers函数，会自动根据设备版本处理
      const dockerContainers = await getContainers(device, password);
      clearTimeout(dockerTimeoutId);
      
      console.log('Docker API返回数据 for device', device.ip, dockerContainers)
      
      // 处理Docker API返回的容器数据
      const processedContainers = (dockerContainers || []).filter(container => {
        // 过滤掉系统插件容器
        const image = container.Image || container.image;
        const name = container.Name || container.Names?.[0];
        const isSystemContainer = image?.includes('myt_sdk') || 
                                 image?.includes('myt_vpc_plugin') ||
                                 name?.includes('myt_sdk') ||
                                 name?.includes('myt_vpc_plugin');
        return !isSystemContainer;
      }).filter(container => {
        // 解析容器坑位编号
        const slotNum = parseContainerSlot(container, device);
        // 只有当坑位编号有效时才添加到列表中
        return slotNum !== null;
      }).map((container, index) => {
        // 解析容器坑位编号
        const slotNum = parseContainerSlot(container, device) || (index + 1);
        
        // 转换Docker API数据格式为我们需要的格式
        const processedContainer = {
          name: container.Names ? container.Names[0]?.replace('/', '') || `container-${index}` : `container-${index}`, // 安全处理容器名称
          status: container.Status?.startsWith('Up') ? 'running' : 'shutdown',
          indexNum: slotNum, // 使用解析得到的坑位编号
          ip: `${device.ip.split('.')[0]}.${device.ip.split('.')[1]}.3.${slotNum + 100}`, // 根据坑位编号生成IP地址
          deviceIp: device.ip, // 添加设备IP属性，记录容器所属设备
          networkMode: container.HostConfig?.NetworkMode || container.NetworkMode || container.networkMode || container.network,
          image: container.Image,
          createTime: new Date(container.Created * 1000).toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
          }),
          modelName: 'Q1', // 默认机型
          // 保留原始端口信息，用于提取9082端口映射
          Ports: container.Ports,
          NetworkSettings: container.NetworkSettings,
          // 兼容不同Docker API版本的端口绑定
          PortBindings: container.PortBindings,
          portBindings: container.portBindings,
          // 保留原始container对象，以便后续使用
          rawContainer: container
        }
        return processedContainer
      })
      
      // 对于V0-V2，所有容器都在processedContainers中
      const allRawContainers = processedContainers;
      console.log('V0-V2 API processedContainers:', processedContainers)
      
      // 转换为云机格式，添加screenshot等属性
      const deviceCloudMachines = processedContainers.map(inst => {
        // 提取端口并生成截图URL，只在创建云机对象时执行一次，后续直接从screenshot属性获取
        const screenshotUrl = getCloudMachineScreenshotUrl(device, inst);
        console.log('Generated cloud machine with screenshot URL:', screenshotUrl, 'for container:', inst.name, 'slot:', inst.indexNum)
        return {
          id: inst.name,
          name: inst.name,
          status: inst.status,
          screenshot: screenshotUrl, // 缓存截图URL，包含端口信息
          screenshotData: null, // 存储截图URL
          screenshotError: false, // 存储截图加载状态
          hasLoadedOnce: false, // 标记是否已成功加载过至少一次
          ip: inst.ip,
          modelName: inst.modelName,
          deviceIp: device.ip,
          indexNum: inst.indexNum,
          networkMode: inst.networkMode || inst.network || inst.NetworkMode || inst?.HostConfig?.NetworkMode
        };
      });
      
      // 缓存设备云机列表
      deviceCloudMachinesCache.value.set(device.ip, deviceCloudMachines);
      

      // 如果是当前选中设备，更新全局云机列表
      if (activeDevice.value && activeDevice.value.ip === device.ip) {
        instances.value = processedContainers;
        allInstances.value = allRawContainers; // 保存所有容器，用于备份列表
        updateCloudMachines();
      }
      
      console.log('Device', device.ip, 'has', deviceCloudMachines.length, 'cloud machines in cache')
      return true;
    } catch (dockerError) {
      console.error('Docker API查询也失败:', dockerError);
      
      // 如果是认证错误，直接抛出，让authRetry函数处理
      if (dockerError.message === 'Authentication Failed') {
        console.log('Docker API认证失败，直接抛出错误让authRetry处理');
        throw dockerError;
      }
      
      // 所有API都失败时，返回空数组
      deviceCloudMachinesCache.value.set(device.ip, []);
      // ⚠️ 不手动标记设备为离线，由心跳检测系统统一管理
      // devicesStatusCache.value.set(device.id, 'offline');
      return false;
    }

    });
  } catch (error) {
    if (error.name === 'AbortError') {
      console.error('获取安卓云机列表超时:', device.ip)
      // 超时错误，标记设备为离线
      deviceCloudMachinesCache.value.set(device.ip, []);
      devicesStatusCache.value.set(device.id, 'offline');
    } else if (error.message === 'Authentication Failed') {
      console.error('获取安卓云机列表失败 - 认证错误:', device.ip)
      console.error('Error details:', error.stack)
      // 认证错误，不标记设备为离线，让认证对话框有机会显示
      // 只清空缓存，不标记离线
      deviceCloudMachinesCache.value.set(device.ip, []);
    } else {
      console.error('获取安卓云机列表失败:', device.ip, error)
      console.error('Error details:', error.stack)
      // 其他错误，清空缓存
      deviceCloudMachinesCache.value.set(device.ip, []);
      // ⚠️ 不手动标记设备为离线，由心跳检测系统统一管理
      // devicesStatusCache.value.set(device.id, 'offline');
    }
  } finally {
    cloudMachineLoadingState.value.set(device.ip, false);
  }
}
const handleBatchAction = async (action, selectedItems, cardOrientation = null) => {
  // 根据云机管理模式检查是否有选中的云机
  let hasSelectedMachines = false
  if (cloudManageMode.value === 'slot') {
    hasSelectedMachines = selectedSlotCloudMachines.value.length > 0
  } else {
    hasSelectedMachines = selectedCloudMachines.value.length > 0
  }
  
  if (!hasSelectedMachines) {
    console.error('没有选中的云机')
    return
  }
  
  loading.value = true
  try {
    switch (action) {
      case 'restart':
        // 实现批量重启功能
        console.log(`执行批量重启操作`)
        
        let containersToRestart = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          containersToRestart = instances.value.filter(inst => 
            selectedSlotCloudMachines.value.includes(inst.indexNum) && 
            inst.status === 'running' // 只选择运行中的容器
          )
        } else {
          // 批量模式：使用已选中的云机
          containersToRestart = selectedCloudMachines.value.filter(machine => 
            machine.status === 'running' // 只选择运行中的容器
          )
        }
        
        if (containersToRestart.length === 0) {
          ElMessage.warning('没有选中运行中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(
            `确定要重启选中的 ${containersToRestart.length} 个运行中的云机吗？重启后容器将会停止并重新启动。`, 
            '批量重启云机', 
            {
              confirmButtonText: '确定',
              cancelButtonText: '取消',
              type: 'warning'
            }
          )
          
          // 为每个容器添加设备信息
          const restartTargets = containersToRestart.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('restart', restartTargets)
          executeTask(taskId)
          
          ElMessage.success('批量重启任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量重启操作')
            ElMessage.info('已取消批量重启操作')
          } else {
            console.error('批量重启失败:', error)
            ElMessage.error(`批量重启失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'reset':
        // 实现批量重置功能
        console.log(`执行批量重置操作`)
        
        let resetContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          resetContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // 批量模式：使用已选中的云机
          resetContainersToOperate = selectedCloudMachines.value
        }
        
        if (resetContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要重置选中的 ${resetContainersToOperate.length} 个容器吗？重置后容器将会被恢复到初始状态。`, '批量重置容器', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          })
          
          // 检查设备版本是否支持重置功能
          let allDevicesSupported = true
          const unsupportedDevices = new Set()
          
          if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
            if (selectedCloudDevice.value.version !== 'v3') {
              allDevicesSupported = false
              unsupportedDevices.add(selectedCloudDevice.value.ip)
            }
          } else {
            // 批量模式下检查所有选中容器的设备版本
            resetContainersToOperate.forEach(container => {
              if ((container.deviceVersion || 'v3') !== 'v3') {
                allDevicesSupported = false
                unsupportedDevices.add(container.deviceIp)
              }
            })
          }
          
          if (!allDevicesSupported) {
            ElMessage.warning(`以下设备不支持重置容器功能: ${Array.from(unsupportedDevices).join(', ')}`)
            break
          }
          
          // 为每个容器添加设备信息
          const resetTargets = resetContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('reset', resetTargets)
          executeTask(taskId)
          
          ElMessage.success('批量重置任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量重置操作')
            ElMessage.info('已取消批量重置操作')
          } else {
            console.error('批量重置失败:', error)
            ElMessage.error(`批量重置失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'projection':
        // 实现批量投屏功能
        console.log(`执行批量投屏操作`)
        
        let projectionContainersToOperate = []
        let allContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          allContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // 批量模式：使用已选中的云机
          allContainersToOperate = selectedCloudMachines.value
        }
        
        // 只保留运行中的云机
        projectionContainersToOperate = allContainersToOperate.filter(container => container.status === 'running')
        
        if (allContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        if (projectionContainersToOperate.length === 0) {
          ElMessage.warning('选中的云机都没有处于运行状态')
          break
        }
        
        // 如果有部分云机未运行，给出提示
        if (projectionContainersToOperate.length < allContainersToOperate.length) {
          const notRunningCount = allContainersToOperate.length - projectionContainersToOperate.length
          ElMessage.info(`有 ${notRunningCount} 个选中的云机未处于运行状态，将只对运行中的 ${projectionContainersToOperate.length} 个云机打开投屏`)
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要对选中的 ${projectionContainersToOperate.length} 个云机打开投屏吗？`, '批量投屏', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'info'
          })
          
          // 计算 orient 参数（如果有 cardOrientation）
          // horizontal: 横屏 = 1, vertical: 竖屏 = 0
          const customOrient = cardOrientation ? (cardOrientation === 'horizontal' ? 1 : 0) : null
          console.log('[批量投屏] cardOrientation:', cardOrientation, ', customOrient:', customOrient)
          
          // 执行批量投屏操作，不使用任务队列
          let successCount = 0
          let failCount = 0
          
          for (const container of projectionContainersToOperate) {
            try {
              if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
                // 坑位模式
                await startProjection({ ip: selectedCloudDevice.value.ip }, container, customOrient)
              } else {
                // 批量模式
                await startProjection({ ip: container.deviceIp }, container, customOrient)
              }
              successCount++
            } catch (error) {
              console.error(`对云机 ${container.name || container.ID} 打开投屏失败:`, error)
              failCount++
            }
          }
          
          // 显示操作结果
          if (successCount > 0) {
            ElMessage.success(`成功对 ${successCount} 个云机打开投屏`)
          }
          if (failCount > 0) {
            ElMessage.warning(`对 ${failCount} 个云机打开投屏失败`)
          }
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量投屏操作')
            ElMessage.info('已取消批量投屏操作')
          } else {
            console.error('批量投屏失败:', error)
            ElMessage.error(`批量投屏失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'projection-control':
        // ????????
        console.log(`????????`)
        console.log('[批量控制] 接收到的 cardOrientation 参数:', cardOrientation)

        let controlContainersToOperate = []
        let allControlContainersToOperate = []

        // ?????????????????
        if (cloudManageMode.value === 'slot') {
          // ???????????????????
          allControlContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // ?????????????
          allControlContainersToOperate = selectedCloudMachines.value
        }

        // ?????????
        controlContainersToOperate = allControlContainersToOperate.filter(container => container.status === 'running')

        if (allControlContainersToOperate.length === 0) {
          ElMessage.warning('???????')
          break
        }

        if (controlContainersToOperate.length === 0) {
          ElMessage.warning('??????????????')
          break
        }

        if (controlContainersToOperate.length < allControlContainersToOperate.length) {
          const notRunningCount = allControlContainersToOperate.length - controlContainersToOperate.length
          ElMessage.info(`? ${notRunningCount} ????????????????????? ${controlContainersToOperate.length} ?????????`)
        }

        try {
          await ElMessageBox.confirm(`??????? ${controlContainersToOperate.length} ???????????`, '????', {
            confirmButtonText: '??',
            cancelButtonText: '??',
            type: 'info'
          })

          const device = (cloudManageMode.value === 'slot' && selectedCloudDevice.value)
            ? { ip: selectedCloudDevice.value.ip }
            : null

          // 如果 cardOrientation 有值，根据它计算 orient 参数
          // horizontal: 横屏 = 1, vertical: 竖屏 = 0
          // 如果没有 cardOrientation（例如从其他地方调用），则传 null 让 API 自动判断
          const orient = cardOrientation ? (cardOrientation === 'horizontal' ? 1 : 0) : null
          console.log('[批量控制] 计算出的 orient:', orient, ', cardOrientation:', cardOrientation)
          
          await startProjectionBatchControl(device, controlContainersToOperate, `????`, orient)
        } catch (error) {
          if (error === 'cancel') {
            console.log('???????????')
            ElMessage.info('?????????')
          } else {
            console.error('??????:', error)
            ElMessage.error(`??????: ${error.message || '????'}`)
          }
        }
        break
      case 'shutdown':
        // 实现批量关机功能
        console.log(`执行批量关机操作`)
        
        let shutdownContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          shutdownContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // 批量模式：使用已选中的云机
          shutdownContainersToOperate = selectedCloudMachines.value
        }
        
        if (shutdownContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要关闭选中的 ${shutdownContainersToOperate.length} 个容器吗？关闭后容器将会停止运行。`, '批量关机容器', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          })
          
          // 为每个容器添加设备信息
          const shutdownTargets = shutdownContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('shutdown', shutdownTargets)
          executeTask(taskId)
          
          ElMessage.success('批量关机任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量关机操作')
            ElMessage.info('已取消批量关机操作')
          } else {
            console.error('批量关机失败:', error)
            ElMessage.error(`批量关机失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'delete':
        // 实现批量删除功能
        console.log(`执行批量删除操作`)
        
        let deleteContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          deleteContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // 批量模式：使用已选中的云机
          deleteContainersToOperate = selectedCloudMachines.value
        }
        
        if (deleteContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        try {
          // 显示确认对话框
          await ElMessageBox.confirm(`确定要删除选中的 ${deleteContainersToOperate.length} 个云机吗？删除后数据将无法恢复。`, '批量删除云机', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'danger'
          })
          
          // 为每个容器添加设备信息
          const deleteTargets = deleteContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 添加到任务队列
          const taskId = addTaskToQueue('delete', deleteTargets)
          executeTask(taskId)
          
          ElMessage.success('批量删除任务已添加到队列')
        } catch (error) {
          if (error === 'cancel') {
            // 用户取消了操作
            console.log('用户取消了批量删除操作')
            ElMessage.info('已取消批量删除操作')
          } else {
            console.error('批量删除失败:', error)
            ElMessage.error(`批量删除失败: ${error.message || '未知错误'}`)
          }
        }
        break
      case 'switchModel':
        // 实现批量切换机型功能
        console.log(`执行批量切换机型操作`)
        
        let switchModelContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          switchModelContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // 批量模式：使用已选中的云机
          switchModelContainersToOperate = selectedCloudMachines.value
        }
        
        if (switchModelContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        // 检查是否所有选中的云机都处于运行状态
        const runningContainers = switchModelContainersToOperate.filter(container => container.status === 'running')
        
        if (runningContainers.length === 0) {
          ElMessage.warning('没有选中已运行的云机，批量切换机型操作仅支持已运行的云机')
          break
        }
        
        if (runningContainers.length < switchModelContainersToOperate.length) {
          ElMessage.warning(`部分选中的云机未运行，仅对 ${runningContainers.length} 个已运行的云机执行操作`)
          // 更新要操作的容器列表，只包含已运行的云机
          switchModelContainersToOperate = runningContainers
        }
        
        // 检查设备版本是否支持切换机型功能
        let newIsV3Device = false
        let newTargetDevice = null
        
        if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
          newIsV3Device = selectedCloudDevice.value.version === 'v3'
          newTargetDevice = selectedCloudDevice.value
        } else if (cloudManageMode.value === 'batch') {
          // 批量模式下，确保所有选中的云机都在同一个设备上
          const deviceIps = new Set(switchModelContainersToOperate.map(machine => machine.deviceIp))
          if (deviceIps.size !== 1) {
            ElMessage.error('批量切换机型功能只支持同一设备上的云机')
            break
          }
          
          const deviceIp = Array.from(deviceIps)[0]
          const versions = new Set(switchModelContainersToOperate.map(machine => machine.deviceVersion))
          newIsV3Device = versions.has('v3') && versions.size === 1
          
          if (newIsV3Device) {
            newTargetDevice = { ip: deviceIp, version: 'v3' }
          }
        }
        
        if (!newIsV3Device || !newTargetDevice) {
          ElMessage.error('只有V3版本设备支持批量切换机型功能')
          break
        }
        
        try {
          // 显示加载状态提示
          const loadingMsg = ElMessage({
            message: '加载机型列表中...',
            type: 'info',
            duration: 0
          })
          
          // 获取可用的手机型号列表
          await getV3PhoneModels(newTargetDevice.ip)
          
          // 关闭加载提示
          setTimeout(() => {
            ElMessage.closeAll()
          }, 100)
          
          if (phoneModels.value.length === 0) {
            ElMessage.warning('未获取到可用的机型列表')
            break
          }
          
          // 为每个容器添加设备信息
          batchSwitchModelTargets.value = newContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 设置操作类型为'new'，表示这是通过批量新机按钮触发的操作
          batchSwitchModelOperationType.value = 'new'
          
          // 初始化机型槽
          initModelSlots()
          
          // 打开批量切换机型对话框
          batchSwitchModelDialogVisible.value = true
        } catch (error) {
          console.error('批量切换机型失败:', error)
          ElMessage.error(`批量切换机型失败: ${error.message || '未知错误'}`)
        }
        break
      case 'new':
        // 实现批量切换机型功能（批量新机）
        console.log(`执行批量切换机型操作（批量新机）`)
        
        let newContainersToOperate = []
        
        // 根据云机管理模式获取需要操作的容器
        if (cloudManageMode.value === 'slot') {
          // 坑位模式：根据选中的坑位号获取容器实例
          newContainersToOperate = instances.value.filter(inst => selectedSlotCloudMachines.value.includes(inst.indexNum))
        } else {
          // 批量模式：使用已选中的云机
          newContainersToOperate = selectedCloudMachines.value
        }
        
        if (newContainersToOperate.length === 0) {
          ElMessage.warning('没有选中的云机')
          break
        }
        
        // 检查是否所有选中的云机都处于运行状态
        const runningNewContainers = newContainersToOperate.filter(container => container.status === 'running')
        
        if (runningNewContainers.length === 0) {
          ElMessage.warning('没有选中已运行的云机，批量新机操作仅支持已运行的云机')
          break
        }
        
        if (runningNewContainers.length < newContainersToOperate.length) {
          ElMessage.warning(`部分选中的云机未运行，仅对 ${runningNewContainers.length} 个已运行的云机执行操作`)
          // 更新要操作的容器列表，只包含已运行的云机
          newContainersToOperate = runningNewContainers
        }
        
        // 检查设备版本是否支持切换机型功能
        let isV3Device = false
        let targetDevice = null
        
        if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
          isV3Device = selectedCloudDevice.value.version === 'v3'
          targetDevice = selectedCloudDevice.value
        } else if (cloudManageMode.value === 'batch') {
          // 批量模式下，确保所有选中的云机都在同一个设备上
          const deviceIps = new Set(newContainersToOperate.map(machine => machine.deviceIp))
          if (deviceIps.size !== 1) {
            ElMessage.error('批量切换机型功能只支持同一设备上的云机')
            break
          }
          
          const deviceIp = Array.from(deviceIps)[0]
          const versions = new Set(newContainersToOperate.map(machine => machine.deviceVersion))
          isV3Device = versions.has('v3') && versions.size === 1
          
          if (isV3Device) {
            targetDevice = { ip: deviceIp, version: 'v3' }
          }
        }
        
        if (!isV3Device || !targetDevice) {
          ElMessage.error('只有V3版本设备支持批量切换机型功能')
          break
        }
        
        try {
          // 显示加载状态提示
          const loadingMsg = ElMessage({
            message: '加载机型列表中...',
            type: 'info',
            duration: 0
          })
          
          // 获取可用的手机型号列表
          await getV3PhoneModels(targetDevice.ip)
          
          // 关闭加载提示
          setTimeout(() => {
            ElMessage.closeAll()
          }, 100)
          
          if (phoneModels.value.length === 0) {
            ElMessage.warning('未获取到可用的机型列表')
            break
          }
          
          // 为每个容器添加设备信息
          batchSwitchModelTargets.value = newContainersToOperate.map(container => {
            if (cloudManageMode.value === 'slot' && selectedCloudDevice.value) {
              return {
                ...container,
                deviceIp: selectedCloudDevice.value.ip,
                deviceVersion: selectedCloudDevice.value.version
              }
            } else {
              return container
            }
          })
          
          // 初始化机型槽
          initModelSlots()
          
          // 打开批量切换机型对话框
          batchSwitchModelDialogVisible.value = true
        } catch (error) {
          console.error('批量切换机型失败:', error)
          ElMessage.error(`批量切换机型失败: ${error.message || '未知错误'}`)
        }
        break
      default:
        console.error('未知的批量操作:', action)
        ElMessage.error('未知的批量操作')
    }
  } catch (error) {
    console.error('执行批量操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    loading.value = false
  }
}
const refreshImageList = async () => {
  if (activeDevice.value) {
    const deviceType = activeDevice.value.name || 'C1'
    await fetchImageList(deviceType)
    
    // 检查在线镜像的上传状态
    await checkOnlineImagesUploadStatus()
  } else {
    await fetchImageList('')
  }
}

const showPasswordDialog = () => {
  if (!activeDevice.value) {
    ElMessage.error('请先选择设备')
    return
  }
  
  passwordForm.value.password = ''
  passwordDialogVisible.value = true
}
const handleClosePassword = async () => {
  if (!activeDevice.value) return
  
  try {
    const confirmResult = await ElMessageBox.confirm('确定要关闭设备密码吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    if (confirmResult) {
      // 检查设备是否需要认证
      if (activeDevice.value.version === 'v3') {
        try {
          // 先尝试无密码访问，看是否需要认证
          const result = await getContainers(activeDevice.value, null);
          // 如果认证失败，说明设备已经设置了密码，需要先认证
          if (result && result.code === 61 && result.message === 'Authentication Failed') {
            // 使用authRetry处理认证
            await authRetry(activeDevice.value, async (currentPassword) => {
              passwordLoading.value = true
              const result = await closeDevicePassword(activeDevice.value, currentPassword)
              if (result.success) {
                ElMessage.success('密码关闭成功')
                // 移除本地保存的密码
                removeDevicePassword(activeDevice.value.ip)
              } else {
                ElMessage.error(`密码关闭失败: ${result.message}`)
              }
            })
            return
          }
        } catch (error) {
          // 无密码访问失败，需要认证
          await authRetry(activeDevice.value, async (currentPassword) => {
            passwordLoading.value = true
            const result = await closeDevicePassword(activeDevice.value, currentPassword)
            if (result.success) {
              ElMessage.success('密码关闭成功')
              // 移除本地保存的密码
              removeDevicePassword(activeDevice.value.ip)
            } else {
              ElMessage.error(`密码关闭失败: ${result.message}`)
            }
          })
          return
        }
      }
      
      // 不需要认证或已经认证成功，直接关闭密码
      passwordLoading.value = true
      // 获取当前保存的密码
      const currentPassword = getDevicePassword(activeDevice.value.ip);
      const result = await closeDevicePassword(activeDevice.value, currentPassword)
      if (result.success) {
        ElMessage.success('密码关闭成功')
        // 移除本地保存的密码
        removeDevicePassword(activeDevice.value.ip)
      } else {
        ElMessage.error(`密码关闭失败: ${result.message}`)
      }
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('关闭密码失败:', error)
      ElMessage.error(`关闭密码失败: ${error.message}`)
    }
  } finally {
    passwordLoading.value = false
  }
}
const upgradeSDK = async () => {
  if (!activeDevice.value || activeDevice.value.version !== 'v3') return
  
  try {
    upgrading.value = true
    ElMessage.info('开始升级SDK...')
    
    // 先检查本地存储的密码
    const savedPassword = getDevicePassword(activeDevice.value.ip)
    console.log('升级SDK - 本地存储的密码:', savedPassword)
    
    // 使用authRetry处理认证
    await authRetry(activeDevice.value, async (password) => {
      try {
        console.log('升级SDK - 收到的密码:', password)
        console.log('升级SDK - 设备IP:', activeDevice.value.ip)
        
        // 使用Wails IPC调用后端UpgradeSDK函数
        const result = await UpgradeSDK(
          activeDevice.value.ip,
          password || ''
        )
        
        if (result.success) {
          ElMessage.success('SDK升级请求成功，正在进行升级...')
          // 重新获取设备信息
          fetchV3DeviceInfo(activeDevice.value)
        } else {
          ElMessage.error(`SDK升级失败: ${result.message}`)
        }
      } catch (error) {
        console.error('SDK升级失败:', error)
        ElMessage.error('SDK升级失败，请检查设备连接')
      } finally {
        upgrading.value = false
      }
    })
  } catch (error) {
    console.error('SDK升级失败:', error)
    ElMessage.error('SDK升级失败，请检查设备连接')
    upgrading.value = false
  }
}
const handleEditNetwork = (network) => {
  if (!activeDevice.value) return
  
  // 填充表单数据
  editNetworkForm.value = {
    networkName: network.Name,
    networkID: network.ID || network.Id, // 同时支持ID和Id字段（处理大小写问题）
    subnet: network.IPAM?.Config?.[0]?.Subnet || '',
    gateway: network.IPAM?.Config?.[0]?.Gateway || '',
    ipRange: network.IPAM?.Config?.[0]?.IPRange || '',
    isPrivate: network.Internal || false
  }
  
  currentEditingNetwork.value = network
  editNetworkDialogVisible.value = true
}
const handleDeleteNetwork = (network) => {
  if (!activeDevice.value) return
  
  // 确认删除
  ElMessageBox.confirm(
    `确定要删除网络 ${network.Name} 吗？`,
    '删除网络',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      // 调用后端API删除网络，同时支持ID和Id字段（处理大小写问题）
      const networkId = network.ID || network.Id
      const result = await deleteDockerNetwork(activeDevice.value, networkId)
      
      if (result.success) {
        ElMessage.success('网络删除成功')
        
        // 刷新网络列表
        await fetchDockerNetworks(activeDevice.value)
      } else {
        ElMessage.error(`删除网络失败: ${result.message}`)
      }
    } catch (error) {
      console.error('删除网络失败:', error)
      ElMessage.error(`删除网络失败: ${error.message}`)
    }
  }).catch(() => {
    // 取消删除
  })
}
const showAddMacvlanDialog = () => {
  // 重置表单
  addMacvlanForm.value = {
    networkName: '',
    parentInterface: '',
    subnet: '',
    gateway: '',
    ipRange: '',
    isPrivate: false
  }
  addMacvlanDialogVisible.value = true
}
const switchImageCategory = async (category) => {
  selectedImageCategory.value = category
  console.log('切换镜像分类:', category)
  
  if (category === 'local') {
    await fetchLocalCachedImages()
    
    // 如果是V3设备，确保已经获取了型号列表
    if (activeDevice.value && activeDevice.value.version === 'v3') {
      await getV3PhoneModels(activeDevice.value.ip)
    }
  } else if (category === 'online') {
    // 在线镜像分类，确保已经获取了在线镜像列表
    if (activeDevice.value) {
      const deviceType = activeDevice.value.name || 'C1'
      await fetchImageList(deviceType)
    } else {
      await fetchImageList('')
    }
    
    // 确保categorizeOnlineImages已经执行完毕，onlineImagesByModel有数据
    await nextTick()
    
    // 确保默认选中在线镜像Q1
    if (onlineImagesByModel.value.has('Q1')) {
      // 如果有Q1型号，直接选中Q1
      currentOnlineImageModel.value = 'Q1'
      console.log('默认选中在线镜像型号: Q1')
    } else if (onlineImagesByModel.value.size > 0) {
      // 如果没有Q1型号，选中第一个型号
      const firstModel = Array.from(onlineImagesByModel.value.keys())[0]
      currentOnlineImageModel.value = firstModel
      console.log('默认选中第一个在线镜像型号:', firstModel)
    } else if (imageList.value.length > 0) {
      // 如果按型号分类没有数据，但imageList有数据，说明可能没有按型号分类
      console.log('在线镜像按型号分类为空，但imageList有数据:', imageList.value.length)
      // 手动创建Q1分类
      currentOnlineImageModel.value = 'Q1'
      // 确保onlineImagesByModel中有Q1分类
      if (!onlineImagesByModel.value.has('Q1')) {
        onlineImagesByModel.value.set('Q1', imageList.value)
      }
    }
  }
}
const refreshOnlineImages = async () => {
  // 确保选中在线镜像标签页
  selectedImageCategory.value = 'online'
  
  if (activeDevice.value) {
    const deviceType = activeDevice.value.name || 'C1'
    await fetchImageList(deviceType)
    
    // 检查在线镜像的上传状态
    await checkOnlineImagesUploadStatus()
    
    // 确保显示在线镜像内容
    await switchImageCategory('online')
  } else {
    await fetchImageList('')
    
    // 确保显示在线镜像内容
    await switchImageCategory('online')
  }
}

const showUpdateImageDialog = async (container) => {
  if (!activeDevice.value) {
    ElMessage.error('没有选中设备')
    return
  }
  
  updateImageContainer.value = container
  
  // 重置表单
  updateImageForm.value = {
    imageSelect: '',
    customImageUrl: '',
    modelName: container.modelName || '',
    enableMagisk: false,
    enableGMS: false,
    dns: container.dns || '',
    customDns: '',
    resolution: 'default',
    customResolution: {
      width: '720',
      height: '1280',
      dpi: '320'
    },
    longitud: '',
    latitude: ''
  }
  
  // 获取设备类型
  const deviceType = activeDevice.value.name || 'C1'
  
  // 获取镜像列表
  await fetchImageList(deviceType)
  
  // 获取当前容器使用的镜像
  const currentImageUrl = container.image || ''
  const cleanedUrl = currentImageUrl.toLowerCase()
  
  // 查找当前镜像是否在列表中 - 优化匹配逻辑
  let currentImage = null
  
  // 1. 精确匹配：检查image.url是否与currentImageUrl完全匹配
  currentImage = filteredImageList.value.find(image => {
    return image.url && image.url.toLowerCase() === cleanedUrl
  })
  
  // 2. 如果没有精确匹配，尝试模糊匹配：检查image.url是否是currentImageUrl的一部分
  if (!currentImage) {
    currentImage = filteredImageList.value.find(image => {
      return image.url && cleanedUrl.includes(image.url.toLowerCase())
    })
  }
  
  // 3. 如果仍然没有找到，尝试在完整镜像列表中查找
  if (!currentImage) {
    currentImage = imageList.value.find(image => {
      return image.url && image.url.toLowerCase() === cleanedUrl
    })
  }
  
  // 4. 最后尝试在完整镜像列表中进行模糊匹配
  if (!currentImage) {
    currentImage = imageList.value.find(image => {
      return image.url && cleanedUrl.includes(image.url.toLowerCase())
    })
  }
  
  if (currentImage) {
    // 如果在列表中，选择它
    updateImageForm.value.imageSelect = currentImage.url
  } else if (currentImageUrl) {
    // 如果不在列表中但有镜像URL，选择自定义并填写
    updateImageForm.value.imageSelect = 'custom'
    updateImageForm.value.customImageUrl = currentImageUrl
  } else {
    // 如果没有镜像URL，设置默认镜像为过滤后的第一个镜像
    if (filteredImageList.value.length > 0) {
      updateImageForm.value.imageSelect = filteredImageList.value[0].url
    }
  }
  
  // 如果是V3设备，获取型号列表
  if (activeDevice.value.version === 'v3') {
    await getV3PhoneModels(activeDevice.value.ip)
  }
  
  updateImageDialogVisible.value = true
}
const refreshLocalImages = async () => {
  await fetchLocalCachedImages()
}
const refreshBoxImages = async () => {
  await fetchBoxImages()
}
</script>

<style scoped>
/* 组件样式 */
.device-left-col {
  display: flex;
  flex-direction: column;
}

.device-right-col {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.device-card {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.group-dropdown-btn {
  max-width: 120px;
}

.group-dropdown-btn .group-text {
  display: inline-block;
  max-width: 85px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.device-table {
  flex: 1;
}

.floating-toolbar {
  background-color: #fff;
  border-bottom: 1px solid #dcdfe6;
  padding: 10px 15px;
  margin-bottom: 15px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.device-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.device-ip-text {
  font-weight: bold;
  font-size: 14px;
}

.divider {
  color: #dcdfe6;
}

.toolbar-button {
  margin-right: 5px;
}

.toolbar-button.active {
  font-weight: bold;
}

.batch-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.table-container {
  flex: 1;
  overflow: auto;
}

.slot-table {
  width: 100%;
}

.slot-cell-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
}

.status-tag-normal {
  min-width: 60px;
  text-align: center;
}

.host-info-container {
  flex: 1;
  overflow: auto;
}

.host-info-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.host-info-content {
  flex: 1;
  overflow: auto;
}

.no-device-selected {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.network-info-container {
  flex: 1;
  overflow: auto;
}

.network-info-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.network-info-content {
  flex: 1;
  overflow: auto;
}

.network-loading, .image-loading {
  padding: 20px;
}

.network-empty, .no-images {
  padding: 50px 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.image-info-container {
  flex: 1;
  overflow: auto;
}

.image-category-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.online-images-container, .local-images-container, .box-images-container {
  flex: 1;
  overflow: auto;
}

.online-images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  padding: 0 20px 20px;
}

.image-card {
  height: fit-content;
}

.image-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.image-name {
  font-weight: bold;
  font-size: 14px;
}

.version-tag {
  margin-left: 10px;
}

.image-card-content {
  padding: 10px 0;
}

.image-meta {
  display: flex;
  gap: 15px;
  margin-bottom: 10px;
}

.image-meta-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #606266;
}

.image-description {
  margin-bottom: 15px;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}

.image-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.download-progress-container {
  margin-top: 20px;
}

.download-progress-text {
  text-align: center;
  margin-top: 10px;
  font-size: 14px;
  color: #606266;
}

.image-table-wrapper {
  padding: 0 20px 20px;
  overflow: auto;
}

.collapse-button {
  cursor: pointer;
}
:deep(.el-table .cell) {
  display: block;
}
</style>
