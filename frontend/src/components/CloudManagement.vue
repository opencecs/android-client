<template>
  <el-row :gutter="12">
    <!-- 左侧栏：设备列表 -->
    <el-col :xs="8" :sm="7" :md="6" :lg="6" :xl="4">
      <el-card shadow="hover" class="device-card">
        <template #header>
          <div class="card-header" style="display: flex; justify-content: space-between; align-items: center;">
            <!-- <span>设备列表</span> -->
            <div style="display: flex; align-items: center;">
              <!-- 模式切换文字按钮 -->
              <div class="mode-switch-text" title="点击切换坑位/批量模式">
                <span 
                  :class="{ active: props.cloudManageMode === 'slot' }"
                  @click="handleCloudManageModeChange('slot')"
                >{{ t('cloudMachine.slotMode') }}</span>
                <span class="divider">|</span>
                <span 
                  :class="{ active: props.cloudManageMode === 'batch' }"
                  @click="handleCloudManageModeChange('batch')"
                >{{ t('cloudMachine.batchMode') }}</span>
              </div>
            </div>
            <!-- <el-button 
              type="success" 
              size="small" 
              @click="handleBatchCreateDevice"
            >
              <el-icon><Plus /></el-icon>
              <span>批量设备创建</span>
            </el-button> -->
          </div>
        </template>
        
        <!-- 搜索框（替换原来的模式切换） -->
        <div class="device-search-box">
          <el-input
            v-model="deviceSearchText"
            :placeholder="t('common.searchGroupOrIP')"
            clearable
            size="small"
            @clear="handleSearchClear"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        
        <!-- 坑位模式：设备IP列表 -->
        <div v-if="props.cloudManageMode === 'slot'" class="device-list-container">
          <div class="grouped-device-list">
            <template v-for="(groupDevices, groupName) in filteredGroupedDevicesByGroup" :key="groupName">
              <!-- 分组标题行 -->
              <div 
                class="group-separator"
                :class="{ 'drag-over': dragOverGroup === groupName }"
                @click="toggleDeviceGroup(groupName)"
                @dragover.prevent="handleDragOver($event, groupName)"
                @dragenter.prevent="handleDragEnter(groupName)"
                @drop="handleDropOnGroup($event, groupName)"
              >
                <el-icon v-if="isGroupCollapsed(groupName)" style="margin-right: 4px;"><ArrowRight /></el-icon>
                <el-icon v-else style="margin-right: 4px;"><ArrowDown /></el-icon>
                <span class="group-separator-span">{{ groupName }} ({{ groupDevices.length }})</span>
                <span v-if="dragOverGroup === groupName" class="drop-hint">{{ t('common.dropToGroup') }}</span>
                <el-button 
                  type="text"
                  size="small" 
                  :icon="Rank"
                  class="group-add-btn group-add-btn-icon"
                  @click.stop="showAddDeviceToGroupDialog(groupName)"
                >{{ t('cloudMachine.batchAddToGroup') }}</el-button>
              </div>
              <!-- 设备列表（可折叠） -->
              <div v-show="!isGroupCollapsed(groupName)" class="group-devices">
                <div 
                  v-for="device in groupDevices" 
                  :key="device.id"
                  class="device-menu-item"
                  :class="{ 'is-active': props.selectedCloudDevice?.id === device.id }"
                  draggable="true"
                  @click="handleCloudDeviceSelect(device)"
                  @dragstart="handleDragStart($event, device)"
                  @dragend="handleDragEnd"
                >
                  <div class="device-menu-item-content">
                    <!-- 设备型号首字母图标 -->
                    <div 
                      class="device-type-icon"
                      :style="{
                        backgroundColor: getDeviceTypeColor(device.name),
                        color: '#fff',
                        borderRadius: '4px',
                        width: '20px',
                        height: '20px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '12px',
                        fontWeight: 'bold',
                        marginRight: '8px'
                      }"
                    >
                      {{ getDeviceTypeName(device.name).charAt(0).toUpperCase() }}
                    </div>
                    <span class="device-ip">{{ device.ip }}</span>
                    <!-- 批量创建按钮 -->
                    <el-button 
                      size="small" 
                      type="link" 
                      :icon="Plus"
                      class="batch-create-btn device-create-btn small-plus-btn"
                      @click.stop="showCreateDialog(device, 'batch')"
                      :title="t('common.create')"
                    ></el-button>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
        
        <!-- 批量模式：分组树形结构 -->
        <div v-else class="batch-mode-container">
          <!-- 移除el-scrollbar，直接使用原生滚动 -->
          <div class="native-scroll-container">
            <el-tree 
              :data="filteredCloudMachineGroups" 
              :props="{
                label: (data) => {
                  if (data.cloudMachines) {
                    // 设备IP节点
                    return data.ip
                  } else if (data.screenshot) {
                    // 云机节点
                    return data.name
                  } else {
                    // 分组节点
                    return data.name
                  }
                },
                children: (data) => {
                  if (data.devices) {
                    return data.devices
                  } else if (data.cloudMachines) {
                    return data.cloudMachines
                  }
                  return []
                }
              }"
              show-checkbox
              default-expand-all
              node-key="id"
              :default-checked-keys="props.treeSelectedKeys"
              @check="handleTreeCheck"
              :allow-drop="handleDrop"
              @node-drop="handleNodeDrop"
              style="width: 100%;"
            >
              <template #default="{ node, data }">
                <span class="tree-node-label">
                  {{ 
                  data.cloudMachines ? data.ip : 
                  data.screenshot ? (() => {
                    // 将类似 "8569541a74175bfe052739c4321ea31b_2_T0002" 处理成 "T0002"
                    const nameParts = data.name.split('_');
                    return nameParts[nameParts.length - 1] || data.name;
                  })() : 
                  data.name 
                }}
                </span>
                <!-- 只有分组节点显示添加设备按钮 -->
                <el-button
                  v-if="!data.cloudMachines && !data.screenshot"
                  type="text"
                  size="small" 
                  :icon="Rank"
                  class="group-add-btn"
                  @click.stop="showAddDeviceToGroupDialog(data.name)"
                >
                {{ t('cloudMachine.batchAddToGroup') }}
              </el-button>
              </template>
            </el-tree>
          </div>
        </div>
      </el-card>
    </el-col>
    
    <!-- 右侧栏：云机管理主要内容 -->
    <el-col :xs="16" :sm="17" :md="18" :lg="18" :xl="20" class="cloud-management-right-column">
      <!-- 悬浮功能栏 -->
        <div class="floating-toolbar">
          <div class="device-info">
            <el-checkbox v-if="cloudManageMode === 'slot'" :indeterminate="isSlotCloudMachinesIndeterminate" v-model="isSlotCloudMachinesAllSelected" @change="handleSlotCloudMachinesSelectAll"></el-checkbox>
            <span class="device-ip-text">
              {{ props.cloudManageMode === 'slot' ? 
                (props.selectedCloudDevice ? t('common.cloudMachineColon', { ip: props.selectedCloudDevice.ip }) : t('common.cloudMachineManagement')) : 
                t('common.batchManagement') }}
            </span>
            <el-button 
              type="primary" 
              size="small" 
              @click="() => {
                if (props.cloudManageMode === 'batch') {
                  // 批量模式：刷新所有选中云机的截图
                  // 父组件会自动处理截图加载
                } else {
                  // 坑位模式：刷新云机列表和截图
                  props.selectedCloudDevice && localFetchAndroidContainers(props.selectedCloudDevice);
                }
              }"
              :disabled="props.cloudManageMode === 'batch' ? false : !props.selectedCloudDevice || props.loading"
            >
              <el-icon :class="{ 'is-rotating': props.loading }"><Refresh /></el-icon> {{ t('common.refreshCloudMachine') }}
            </el-button>
          </div>
        
        <!-- 批量操作按钮和布局切换 -->
        <el-space wrap class="batch-actions">
          <el-dropdown @command="handleBatchAction" trigger="click">
            <el-button size="small">
              {{ t('common.batchActions') }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="restart">{{ t('common.batchRestart') }}</el-dropdown-item>
                <el-dropdown-item command="reset">{{ t('common.batchReset') }}</el-dropdown-item>
                <el-dropdown-item command="projection">{{ t('common.batchProjection') }}</el-dropdown-item>
                <el-dropdown-item command="shutdown">{{ t('common.batchShutdown') }}</el-dropdown-item>
                <el-dropdown-item command="update-image">{{ t('common.batchUpdateImage') }}</el-dropdown-item>
                <el-dropdown-item command="switch-backup">{{ t('common.batchSwitchBackup') }}</el-dropdown-item>
                <el-dropdown-item command="delete" style="color: var(--el-color-danger)">{{ t('common.batchDelete') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-button 
            :type="props.isBatchProjectionControlling ? 'warning' : 'primary'" 
            @click="handleBatchAction('projection-control')" 
            size="small"
          >
            {{ props.isBatchProjectionControlling ? t('common.stopControl') : t('common.batchControl') }}
          </el-button>
          <el-button @click="handleBatchCleanupProjection" size="small">{{ t('common.batchCloseProjection') }}</el-button>
          <el-button @click="handleBatchAction('upload')" size="small">{{ t('common.batchUpload') }}</el-button>
          <el-button type="primary" @click="handleBatchAction('new')" size="small">{{ t('common.batchNewDevice') }}</el-button>
          
          <!-- 统一使用滑动条控制缩放（坑位模式和批量模式通用） -->
          <div style="display: flex; align-items: center; gap: 8px; min-width: 200px;">
            <span style="font-size: 12px; color: #606266; white-space: nowrap;">{{ t('common.zoom') }}:</span>
            <el-slider 
              v-model="screenshotScale" 
              :min="25" 
              :max="150" 
              :step="5"
              :show-tooltip="true"
              :format-tooltip="(val) => `${val}%`"
              style="flex: 1; min-width: 120px;"
              size="small"
              @change="saveScreenshotScale"
            />
            <span style="font-size: 12px; color: #909399; min-width: 36px;">{{ screenshotScale }}%</span>
          </div>
          
          <el-select v-model="layoutMode" :placeholder="t('common.layout')" size="small" style="width: 80px;">
            <el-option :label="t('common.grid')" value="grid"></el-option>
            <el-option :label="t('common.list')" value="list"></el-option>
          </el-select>
          <!-- 横竖屏切换按钮 -->
          <el-button 
            v-if="layoutMode === 'grid'" 
            size="small" 
            @click="cardOrientation = cardOrientation === 'vertical' ? 'horizontal' : 'vertical'"
            :title="cardOrientation === 'vertical' ? t('common.switchToHorizontal') : t('common.switchToVertical')"
          >
            {{ cardOrientation === 'vertical' ? t('common.vertical') : t('common.horizontal') }}
          </el-button>
        </el-space>
      </div>
      
      <!-- 云机列表容器 -->
      <div class="cloud-machine-container">
        <!-- 坑位模式：12个坑位操作 -->
        <div v-if="props.cloudManageMode === 'slot'" class="cloud-machine-slots">
          <!-- 移除IP显示文字 -->
          <el-checkbox-group 
            v-model="selectedSlotCloudMachines" 
            class="cloud-machine-grid" 
            :class="[`orientation-${cardOrientation}`]"
            :style="{ '--screenshot-scale': screenshotScale / 100 }"
            v-if="layoutMode === 'grid'"
          >
            <!-- 根据设备型号生成不同数量的坑位 -->
                    <el-card 
                      v-for="i in (props.selectedCloudDevice && props.selectedCloudDevice.name && props.selectedCloudDevice.name.toLowerCase().includes('p1') ? 24 : 12)" 
                      :key="i"
              shadow="hover"
              class="cloud-machine-card"
              :class="`card-${cardOrientation}`"
              @contextmenu="handleContextMenu($event, i)"
            >
              <!-- 切换备份时的覆盖层 -->
              <div 
                v-if="switchingBackupSlot === i" 
                class="switching-backup-overlay"
              >
                <el-icon class="is-loading"><Loading /></el-icon>
                <span>{{ t('common.switching') }}</span>
              </div>
              
              <template #header>
                <div class="cloud-machine-card-header-light">
                  <!-- 只显示坑位编号 -->
                  <el-checkbox :label="i"></el-checkbox>
                  <!-- <span>坑位 {{ i }}</span> -->
                  <span class="instance-name" :title="formatInstanceName(props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i)?.name)">
                    {{ formatInstanceName(props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i)?.name) }}
                  </span>
                  <!-- 根据实际容器数据显示状态 -->
                  <el-tag 
                            v-if="props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i)?.status === 'running'" 
                            size="small" 
                            type="success" 
                            class="status-tag-normal"
                          >
                            {{ t('common.running') }}
                          </el-tag>
                          <el-tag 
                            v-else-if="props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i)?.status" 
                            size="small" 
                            type="warning" 
                            class="status-tag-normal"
                          >
                            {{ t('common.backupCount', { count: getBackupCount(i) }) }}
                          </el-tag>
                          <el-tag 
                            v-else 
                            size="small" 
                            type="info" 
                            class="status-tag-normal"
                          >
                            {{ t('common.empty') }}
                          </el-tag>

                </div>
              </template>
              <!-- 根据实际容器数据显示内容 -->
              <div class="cloud-machine-screenshot-large">
                <template v-if="props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i)">
                  <!-- 有容器实例的坑位 -->
                  <template v-if="props.instances.find(inst => inst.indexNum === i)?.status === 'running'">
                    <!-- 运行中状态显示截图 -->
                    <template v-if="props.cloudMachinesByName.get(props.instances.find(inst => inst.indexNum === i)?.name)">
                      <!-- 使用新的截图组件（数据由后端缓存驱动） -->
                      <ScreenshotImage
                        :screenshot-data="(() => { const inst = props.instances.find(inst => inst.indexNum === i); return inst ? (props.screenshotCache?.get(`${props.selectedCloudDevice.ip}_${inst.name}`) || '') : '' })()"
                        :device-key="`${props.selectedCloudDevice.ip}_${i}`"
                        :rotate="cardOrientation === 'horizontal'"
                        @click="() => {
                          const instance = props.instances.find(inst => inst.indexNum === i);
                          if (instance) {
                            startProjection({ ip: instance.networkName == 'myt' ? instance.ip : props.selectedCloudDevice.ip }, instance);
                          }
                        }"
                      />
                    </template>
                    <template v-else>
                      <!-- 未找到云机记录，显示加载中 -->
                      <div class="screenshot-loading" @click="() => {
                          const instance = props.instances.find(inst => inst.indexNum === i);
                          if (instance) {
                            startProjection({ ip: instance.networkName == 'myt' ? instance.ip : props.selectedCloudDevice.ip }, instance);
                          }
                        }" style="cursor: pointer;">
                        <el-icon class="is-loading"><Loading /></el-icon>
                        <div style="text-align: center;">
                          <div>{{ t('common.booting') }}</div>
                        </div>
                      </div>
                    </template>
                  </template>
                  <template v-else>
                    <!-- 关机状态显示提示 -->
                    <div class="screenshot-offline">
                      <span>{{ t('common.shutdownClickBackup') }}</span>
                    </div>
                  </template>
                </template>
                <template v-else-if="props.selectedCloudDevice">
                  <!-- 未分配容器实例的坑位 -->
                  <div class="screenshot-empty">
                    <!-- <span>{{ selectedCloudDevice.ip + ' | 空坑位' }}</span> -->
                     <span>{{ t('common.emptySlot') }}</span>
                  </div>
                </template>
                <template v-else>
                  <!-- 未选择设备时，显示空坑位 -->
                  <div class="screenshot-empty">
                    <span>{{ t('common.emptySlot') }}</span>
                  </div>
                </template>
              </div>
              <div class="cloud-machine-info-small">
                <el-space size="small" class="cloud-machine-actions">
                    <!-- 创建按钮：始终可点击，未选设备时提示用户 -->
                    <el-button size="small" type="primary" class="device-create-btn" @click.stop="props.selectedCloudDevice ? showCreateDialog(props.selectedCloudDevice, 'slot', i) : ElMessage.warning(t('common.selectDeviceFirst'))">{{ t('common.create') }}</el-button>
                    
                    <!-- ADB按键，只在有容器实例且状态为运行中时显示 -->
                    <template v-if="props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i && inst.status === 'running')">
                      <el-button 
                        size="small" 
                        type="success" 
                        @click="handleADBConnect(i)"
                      >
                        ADB
                      </el-button>
                    </template>
                  
                  
                  
                  <template v-if="props.selectedCloudDevice && props.instances.find(inst => inst.indexNum === i)">
                    <!-- 有容器实例的坑位显示切换备份按钮 -->
                    <el-button 
                      size="small" 
                      type="primary" 
                      @click="showBackupList(i)"
                      :disabled="switchingBackupSlot === i"
                    >
                      {{ t('common.switchBackupBtn') }}
                    </el-button>
                  </template>

                </el-space>
              </div>
            </el-card>
          </el-checkbox-group>
          <!-- 列表模式显示12个坑位 -->
          <el-table 
            v-else 
            :data="groupedInstances" 
            stripe 
            size="small" 
            class="slot-table"
            :row-key="row => row.slotNum"
            @selection-change="handleSlotSelectionChange"
          >
            <el-table-column type="selection" :reserve-selection="true" width="55"></el-table-column>
            <el-table-column prop="slotNum" :label="t('common.slot')" width="70" align="center"s></el-table-column>
            <el-table-column :label="t('common.instanceNameLabel')" width="110" align="center">
              <template #default="scope">
                {{ formatInstanceName(scope.row.name) }}
              </template>
            </el-table-column>
            <el-table-column prop="ip" :label="t('common.ipAddress')" width="130" align="center"></el-table-column>
            <el-table-column :label="t('common.systemImage')" width="180" align="center">
              <template #default="scope">
                <div 
                  style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                  :title="getImageDisplayName(scope.row.image)"
                >
                  {{ getImageDisplayName(scope.row.image) }}
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.createTime')" width="160" align="center">
              <template #default="scope">
                {{ scope.row.created ? new Date(scope.row.created).toLocaleString('zh-CN') : scope.row.createTime }}
              </template>
            </el-table-column>
            <el-table-column prop="status" :label="t('common.statusLabel')" width="100" align="center">
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
            <el-table-column prop="modelName" :label="t('common.modelLabel')" width="100" align="center">
              <template #default="scope">
               {{ formatInstanceModel(scope.row.modelPath) || t('common.noneLabel') }}
              </template>
            </el-table-column>
            <el-table-column :label="t('common.operationLabel')" width="320" fixed="right" align="center">
              <template #default="scope">
                <el-space size="small">
                  <el-button 
                    v-if="scope.row.status === 'shutdown'"
                    size="small" 
                    type="primary"
                    @click="showCreateDialog(props.activeDevice, 'slot', scope.row.slotNum)"
                  >
                    {{ t('common.create') }}
                  </el-button>
                  
                  <!-- ADB按键，只在状态为运行中时显示 -->
                  <el-button 
                    v-if="scope.row.status === 'running'"
                    size="small" 
                    type="success"
                    @click="handleADBConnect(scope.row.slotNum)"
                  >
                    ADB
                  </el-button>
                  
                  <el-button 
                    v-if="scope.row.status !== 'shutdown'"
                    size="small" 
                    type="primary"
                    @click="showBackupList(scope.row.slotNum)"
                  >
                    切换
                  </el-button>
                  
                  <el-button 
                    v-if="scope.row.name"
                    size="small" 
                    type="warning"
                    @click="showCopyDialog(scope.row)"
                  >
                    {{ t('common.copy') }}
                  </el-button>

                  <el-button 
                    size="small" 
                    type="danger" 
                    @click="() => {
                      // 检查是否为空坑位（状态为 shutdown 且没有实例名称）
                      if (scope.row.status === 'shutdown' || !scope.row.name || scope.row.name === '') {
                        ElMessage.warning(t('common.slotNoMachineCannotDelete', { slot: scope.row.slotNum }));
                        return;
                      }
                      handleDeleteContainer(scope.row);
                    }"
                  >
                    {{ t('common.delete') }}
                  </el-button>
                </el-space>
              </template>
            </el-table-column>
          </el-table>
        </div>
        
        <!-- 批量模式：选中的云机列表 -->
        <div v-else class="batch-cloud-machines">
          <h3 style="margin-bottom: 16px; font-size: 14px; color: #606266;">
            {{ t('common.selectedCloudMachineList', { count: selectedCloudMachines.length }) }}
          </h3>
          
          <!-- 空状态提示 -->
          <div v-if="selectedCloudMachines.length === 0" style="text-align: center; padding: 60px 20px; color: #909399;">
            <el-icon :size="48" style="margin-bottom: 16px; color: #C0C4CC;">
              <InfoFilled />
            </el-icon>
            <p style="font-size: 14px; margin: 0 0 8px 0;">{{ t('common.noSelectedCloudMachine') }}</p>
            <p style="font-size: 12px; margin: 0; color: #C0C4CC;">{{ t('common.pleaseSelectCloudMachine') }}</p>
          </div>
          
          <!-- 云机列表 - 网格布局 -->
          <div 
            v-else-if="layoutMode === 'grid'" 
            class="cloud-machine-grid" 
            :class="[`orientation-${cardOrientation}`]"
            :style="{ '--screenshot-scale': screenshotScale / 100 }"
          >
            <el-card 
              v-for="machine in selectedCloudMachines" 
              :key="machine.id"
              shadow="hover"
              class="cloud-machine-card"
              :class="`card-${cardOrientation}`"
              @contextmenu="handleContextMenu($event, machine)"
            >
              <template #header>
                <div class="cloud-machine-card-header-light">
                  <span class="instance-name">{{ (() => {
                    const nameParts = machine.name.split('_');
                    return nameParts[nameParts.length - 1] || machine.name;
                  })() }}</span>
                  <el-tag size="small" type="success" class="status-tag-normal">{{ t('common.running') }}</el-tag>
                </div>
              </template>
              <div class="cloud-machine-screenshot-large">
                <!-- 使用新的截图组件（数据由后端缓存驱动，无需 screenshot URL） -->
                <ScreenshotImage
                  :screenshot-data="props.screenshotCache?.get(machine.id) || ''"
                  :device-key="machine.id"
                  :rotate="cardOrientation === 'horizontal'"
                  @click="() => { startProjection({ ip: machine.networkName == 'myt' ? machine.ip : machine.deviceIp }, machine) }"
                />
              </div>
              <div class="cloud-machine-info-small">
                <el-space size="small" class="cloud-machine-actions">
                  <el-button 
                    size="small" 
                    type="success" 
                    @click="handleADBConnect(machine)"
                  >
                    ADB
                  </el-button>
                </el-space>
              </div>
            </el-card>
          </div>
          
          <!-- 云机列表 - 列表布局 -->
          <div v-else>
            <div class="table-container">
              <el-table :data="selectedCloudMachines" stripe size="small"  class="cloud-machine-table">
                
                <el-table-column prop="ip" :label="t('common.ipAddress')" width="130" align="center"></el-table-column>
                <el-table-column :label="t('common.cloudMachineName')" width="110" align="center">
                  <template #default="scope">
                    {{ 
                      (() => {
                        const nameParts = scope.row.name.split('_');
                        return nameParts[nameParts.length - 1] || scope.row.name;
                      })() 
                    }}
                  </template>
                </el-table-column>
                <el-table-column :label="t('common.systemImage')" width="180" align="center">
                  <template #default="scope">
                    <div 
                      style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                      :title="getImageDisplayName(scope.row.image)"
                    >
                      {{ getImageDisplayName(scope.row.image) }}
                    </div>
                  </template>
                </el-table-column>
                <el-table-column prop="created" :label="t('common.createTime')" width="160" align="center"></el-table-column>
                <el-table-column prop="status" :label="t('common.statusLabel')" width="100" align="center">
                  <template #default="scope">
                    <el-tag size="small" :type="scope.row.status === 'running' ? 'success' : 'info'" class="status-tag-normal">
                      {{ scope.row.status === 'running' ? t('common.running') : t('common.shutdownStatus') }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="modelName" :label="t('common.modelLabel')" width="100" align="center">
                 <template #default="scope">
                   {{ formatInstanceModel(scope.row.modelPath) || t('common.noneLabel') }}
                 </template>
                </el-table-column>
                <el-table-column :label="t('common.operationLabel')" width="250" fixed="right" align="center">
                  <template #default="scope">
                    <el-space size="small">
                      <!-- ADB按键，只在状态为运行中时显示 -->
                      <el-button 
                        v-if="scope.row.status === 'running'"
                        size="small" 
                        type="success"
                        @click="handleADBConnect(scope.row)"
                      >
                        ADB
                      </el-button>
                      
                      <el-button size="small" type="primary" @click="() => { startProjection({ ip: scope.row.networkName == 'myt' ? scope.row.ip : scope.row.deviceIp }, scope.row) }">{{ t('common.openProjectionBtn') }}</el-button>

                      <el-button 
                    size="small" 
                    type="danger" 
                    @click="() => {
                      // 检查是否为空坑位（状态为 shutdown 且没有实例名称）
                      if (scope.row.status === 'shutdown' || !scope.row.name || scope.row.name === '') {
                        ElMessage.warning(t('common.slotNoMachineCannotDelete', { slot: scope.row.slotNum }));
                        return;
                      }
                      handleDeleteContainer(scope.row);
                    }"
                  >
                    {{ t('common.delete') }}
                  </el-button>
                    </el-space>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </div>
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
  
  <!-- 添加设备到分组对话框 -->
  <el-dialog
    v-model="addDeviceToGroupDialogVisible"
    :title="`${t('common.addDevicesToGroup')}: ${targetGroupForAddDevice}`"
    width="500px"
    append-to-body
  >
    <div class="add-device-to-group-content">
      <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <span style="color: #909399;">{{ t('common.selectDevicesToAdd') }}</span>
        <el-checkbox
          v-model="isAllDevicesToAddSelected"
          :indeterminate="isDevicesToAddIndeterminate"
          :disabled="availableDevicesToAdd.length === 0"
        >{{ t('common.selectAllLabel') }}</el-checkbox>
      </div>
      <el-checkbox-group v-model="selectedDevicesToAdd">
        <div class="device-checkbox-list" style="max-height: 300px; overflow-y: auto;">
          <el-checkbox 
            v-for="device in availableDevicesToAdd" 
            :key="device.id" 
            :label="device.id"
            style="display: flex; align-items: center; margin-bottom: 8px;"
          >
            <div 
              class="device-type-icon"
              :style="{
                backgroundColor: getDeviceTypeColor(device.name),
                color: '#fff',
                borderRadius: '4px',
                width: '20px',
                height: '20px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '12px',
                fontWeight: 'bold',
                marginRight: '8px'
              }"
            >
              {{ getDeviceTypeName(device.name).charAt(0).toUpperCase() }}
            </div>
            <span>{{ device.ip }} ({{ device.group || t('common.defaultGroup') }})</span>
          </el-checkbox>
        </div>
      </el-checkbox-group>
      <el-empty v-if="availableDevicesToAdd.length === 0" :description="t('common.allDevicesInGroup')" />
    </div>
    <template #footer>
      <el-button @click="addDeviceToGroupDialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="confirmAddDevicesToGroup" :disabled="selectedDevicesToAdd.length === 0">
        {{ t('common.addCount', { count: selectedDevicesToAdd.length }) }}
      </el-button>
    </template>
  </el-dialog>

  <!-- 复制云机对话框 -->
  <el-dialog
    v-model="copyDialogVisible"
    :title="t('common.copyCloudMachine')"
    width="460px"
    append-to-body
    @close="resetCopyForm"
  >
    <el-form :model="copyForm" label-width="90px" size="small">
      <el-form-item :label="t('common.cloudMachineName')">
        <el-input v-model="copyForm.name" disabled />
      </el-form-item>
      <el-form-item :label="t('common.targetSlot')">
        <el-input-number
          v-model="copyForm.indexNum"
          :min="1"
          :max="props.selectedCloudDevice && props.selectedCloudDevice.name && props.selectedCloudDevice.name.toLowerCase().includes('p1') ? 24 : 12"
          controls-position="right"
          style="width: 100%;"
        />
        <div style="font-size: 12px; color: #909399; margin-top: 4px;">{{ t('common.targetSlotHint') }}</div>
      </el-form-item>
      <el-form-item :label="t('common.copyCount')">
        <el-input-number
          v-model="copyForm.count"
          :min="1"
          :max="999"
          controls-position="right"
          style="width: 100%;"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="copyDialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="confirmCopyMachine">{{ t('common.confirmCopy') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>

// 返回设备的 host:port，若 ip 已含端口则直接使用，否则追加默认 8000
const getDeviceAddr = (ip) => {
  if (!ip) return ip
  const lastColon = ip.lastIndexOf(':')
  if (lastColon === -1) return ip + ':8000'
  return /^\d+$/.test(ip.slice(lastColon + 1)) ? ip : ip + ':8000'
}


import { ref, reactive, computed, watch, nextTick, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { ElMessage, ElMessageBox, ElLoading, ElTree, ElSelect, ElOption, ElTable, ElTableColumn, ElCheckbox, ElCheckboxGroup, ElDialog, ElEmpty, ElDropdown, ElDropdownMenu, ElDropdownItem } from 'element-plus'
import { Plus, Refresh, VideoCamera, Warning, Loading, ArrowDown, ArrowRight, Rank, Search, InfoFilled } from '@element-plus/icons-vue'
import axios from 'axios'
import * as api from '../services/api.js'
import * as cloudMachineFunctions from '../services/cloudMachineFunctions.js'
import AddDeviceDialog from './AddDeviceDialog.vue'
import SeamlessImage from './SeamlessImage.vue'
import ScreenshotImage from './ScreenshotImage.vue'

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

// 导入 xterm.js 及其相关依赖
import 'xterm/css/xterm.css'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'

// 终端相关数据 - 确保终端独立性
let term = null, socket = null, fitAddon = null, heartbeatTimer = null
const terminalVisible = ref(false)

// 终端唯一标识符，用于区分不同终端实例
let terminalInstanceId = null

// 终端事件监听器数组，用于管理和清理事件监听器
const terminalEventListeners = []

// 云机管理模式切换按钮
// 使用props中的cloudManageMode、cloudMachineGroups和selectedCloudDevice
const selectedSlotCloudMachines = ref([]) // 选中的坑位云机ID
const isSlotCloudMachinesAllSelected = ref(false) // 是否全选坑位云机
const previousGroups = ref([]) // 用于日志防抖
const previousDeviceId = ref(null) // 用于日志防抖

// 布局模式切换
const layoutMode = ref('grid') // grid: 网格布局, list: 列表布局
const cardOrientation = ref('vertical') // vertical: 竖屏, horizontal: 横屏
const zoomLevel = ref('low') // high, medium, low
const screenshotScale = ref(80) // 截图缩放比例，25%-150%，默认80%

// 云机相关数据
const instances = ref([]) // 实例列表
const cloudMachines = ref([]) // 云机列表
const cloudMachinesByName = ref(new Map()) // 按名称索引的云机列表
const selectedCloudMachines = ref([]) // 选中的云机列表
let currentCheckedKeys = [] // 当前勾选的节点 key，用于分组数据更新时重算

// 设备相关数据
const devices = ref([]) // 设备列表
const devicesStatusCache = ref(new Map()) // 设备状态缓存
const loading = ref(false) // 加载状态
const statusCacheVersion = ref(0) // Map 变化触发器，用于强制 computed 重新求值
const deviceSearchText = ref('') // 搜索文本

// 分组折叠状态
const collapsedDeviceGroups = ref(new Set())

// 复制云机对话框
const copyDialogVisible = ref(false)
const copyLoading = ref(false)
const copyForm = ref({
  name: '',
  indexNum: 1,
  count: 1,
  version: 'v3'
})

// 搜索过滤后的分组设备列表
const filteredGroupedDevicesByGroup = computed(() => {
  // 引用 statusCacheVersion 强制在 Map 内容变化时重新求值
  const _v = statusCacheVersion.value
  let filteredDevices = props.devices.filter(d => props.devicesStatusCache.get(d.id) === 'online')

  // 如果有搜索文本，进行过滤
  if (deviceSearchText.value.trim()) {
    const searchText = deviceSearchText.value.trim().toLowerCase()
    filteredDevices = filteredDevices.filter(device => {
      const groupName = (device.group || '默认分组').toLowerCase()
      const deviceIP = device.ip.toLowerCase()
      return groupName.includes(searchText) || deviceIP.includes(searchText)
    })
  }

  const groups = {}
  const defaultGroup = '默认分组'
  
  filteredDevices.forEach(device => {
    const groupName = device.group || defaultGroup
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(device)
  })
  
  // 每个分组内的设备按IP地址排序
  Object.keys(groups).forEach(groupName => {
    groups[groupName].sort((a, b) => compareIPs(a.ip, b.ip))
  })
  
  return groups
})

// 清空搜索
const handleSearchClear = () => {
  deviceSearchText.value = ''
}

// 过滤后的树形结构（用于批量模式）
const filteredCloudMachineGroups = computed(() => {
  if (!deviceSearchText.value.trim()) {
    return props.cloudMachineGroups
  }

  const searchText = deviceSearchText.value.trim().toLowerCase()

  const filterGroup = (group) => {
    const filteredDevices = (group.devices || []).filter(device => {
      const deviceIP = device.ip.toLowerCase()
      if (deviceIP.includes(searchText)) {
        return true
      }

      if (device.cloudMachines) {
        const hasMatchingCloudMachines = device.cloudMachines.some(cm => {
          const cloudMachineName = (cm.name || '').toLowerCase()
          return cloudMachineName.includes(searchText)
        })
        return hasMatchingCloudMachines
      }

      return false
    })

    const filteredDeviceIPs = filteredDevices.map(d => d.ip)

    const devicesWithFilteredCloudMachines = filteredDevices.map(device => {
      if (!device.cloudMachines) {
        return device
      }

      if (deviceIPIncludesSearch(device.ip)) {
        return device
      }

      const filteredCloudMachines = device.cloudMachines.filter(cm => {
        const cloudMachineName = (cm.name || '').toLowerCase()
        return cloudMachineName.includes(searchText)
      })

      return {
        ...device,
        cloudMachines: filteredCloudMachines
      }
    })

    return devicesWithFilteredCloudMachines.length > 0 ? {
      ...group,
      devices: devicesWithFilteredCloudMachines
    } : null
  }

  const deviceIPIncludesSearch = (ip) => {
    return ip.toLowerCase().includes(searchText)
  }

  const filteredGroups = props.cloudMachineGroups
    .map(group => {
      const groupName = (group.name || '').toLowerCase()
      if (groupName.includes(searchText)) {
        return group
      }
      return filterGroup(group)
    })
    .filter(group => group !== null)

  return filteredGroups
})

// 按IP地址比较（用于正确排序）
const compareIPs = (ip1, ip2) => {
  const parts1 = ip1.split('.').map(Number)
  const parts2 = ip2.split('.').map(Number)
  for (let i = 0; i < 4; i++) {
    if (parts1[i] < parts2[i]) return -1
    if (parts1[i] > parts2[i]) return 1
  }
  return 0
}

// 按分组整理的设备列表（用于坑位模式设备列表）
const groupedDevicesByGroup = computed(() => {
  const _v2 = statusCacheVersion.value
  const onlineDevices = props.devices.filter(d => props.devicesStatusCache.get(d.id) === 'online')
  const groups = {}
  const defaultGroup = '默认分组'
  
  onlineDevices.forEach(device => {
    const groupName = device.group || defaultGroup
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(device)
  })
  
  // 每个分组内的设备按IP地址排序
  Object.keys(groups).forEach(groupName => {
    groups[groupName].sort((a, b) => compareIPs(a.ip, b.ip))
  })
  
  return groups
})

// 切换分组折叠状态
const toggleDeviceGroup = (groupName) => {
  if (collapsedDeviceGroups.value.has(groupName)) {
    collapsedDeviceGroups.value.delete(groupName)
  } else {
    collapsedDeviceGroups.value.add(groupName)
  }
}

// 检查分组是否折叠
const isGroupCollapsed = (groupName) => {
  return collapsedDeviceGroups.value.has(groupName)
}

// 拖拽相关状态
const draggedDevice = ref(null)
const dragOverGroup = ref(null)

// 开始拖拽
const handleDragStart = (event, device) => {
  console.log('[拖拽] 开始拖拽设备:', device.ip, 'ID:', device.id)
  draggedDevice.value = device
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', device.id)
}

// 拖拽进入分组
const handleDragEnter = (groupName) => {
  dragOverGroup.value = groupName
}

// 拖拽经过分组
const handleDragOver = (event, groupName) => {
  event.dataTransfer.dropEffect = 'move'
}

// 放置到分组
const handleDropOnGroup = (event, targetGroup) => {
  event.preventDefault()
  dragOverGroup.value = null
  
  console.log('[拖拽] 放置到分组:', targetGroup, '拖拽设备:', draggedDevice.value?.ip)
  
  if (draggedDevice.value) {
    const device = draggedDevice.value
    const oldGroup = device.group || '默认分组'
    
    console.log('[拖拽] 设备当前分组:', oldGroup, '目标分组:', targetGroup)
    
    if (oldGroup !== targetGroup) {
      console.log('[拖拽] 触发 moveDeviceToGroup 事件')
      emit('moveDeviceToGroup', device.id, targetGroup)
    } else {
      console.log('[拖拽] 设备已在目标分组中，无须移动')
    }
  } else {
    console.log('[拖拽] 错误: 没有拖拽设备数据')
  }
  
  draggedDevice.value = null
}

// 拖拽结束时清理状态
const handleDragEnd = () => {
  dragOverGroup.value = null
  draggedDevice.value = null
}

// 添加设备到分组对话框相关
const addDeviceToGroupDialogVisible = ref(false)
const targetGroupForAddDevice = ref('')
const selectedDevicesToAdd = ref([])

// 可添加到当前分组的设备列表
const availableDevicesToAdd = computed(() => {
  if (!targetGroupForAddDevice.value) return []
  return props.devices.filter(d => (d.group || '默认分组') !== targetGroupForAddDevice.value)
})

// 全选相关计算属性
const isAllDevicesToAddSelected = computed({
  get: () => {
    return availableDevicesToAdd.value.length > 0 && selectedDevicesToAdd.value.length === availableDevicesToAdd.value.length
  },
  set: (val) => {
    if (val) {
      selectedDevicesToAdd.value = availableDevicesToAdd.value.map(d => d.id)
    } else {
      selectedDevicesToAdd.value = []
    }
  }
})

const isDevicesToAddIndeterminate = computed(() => {
  return selectedDevicesToAdd.value.length > 0 && selectedDevicesToAdd.value.length < availableDevicesToAdd.value.length
})

// 显示添加设备到分组对话框
const showAddDeviceToGroupDialog = (groupName) => {
  targetGroupForAddDevice.value = groupName
  selectedDevicesToAdd.value = []
  addDeviceToGroupDialogVisible.value = true
}

// 确认添加设备到分组
const confirmAddDevicesToGroup = () => {
  selectedDevicesToAdd.value.forEach(deviceId => {
    emit('moveDeviceToGroup', deviceId, targetGroupForAddDevice.value)
  })
  ElMessage.success(`已将 ${selectedDevicesToAdd.value.length} 个设备添加到 "${targetGroupForAddDevice.value}"`)
  addDeviceToGroupDialogVisible.value = false
  selectedDevicesToAdd.value = []
}

// 树形结构相关 - 现在使用props接收来自App.vue的选中状态

// 截图管理相关 - 现在由父组件App.vue处理

// 备份相关
const switchingBackupSlot = ref(null) // 当前正在切换备份的坑位

// 添加设备对话框相关
const addDeviceDialogVisible = ref(false)
const existingDeviceIds = ref(new Set())

// 计算属性
const isSlotCloudMachinesIndeterminate = computed(() => {
  const maxSlots = props.selectedCloudDevice && props.selectedCloudDevice.name && props.selectedCloudDevice.name.toLowerCase().includes('p1') ? 24 : 12
  return selectedSlotCloudMachines.value.length > 0 && selectedSlotCloudMachines.value.length < maxSlots
})

const draggableEnabled = computed(() => {
  return props.cloudManageMode === 'batch'
})

const groupedInstances = computed(() => {
  if (!props.selectedCloudDevice) return []
  let result = []
  // 记录每个坑位已经处理了多少个实例
  const slotCountMap = new Map()
  
  // 根据设备型号确定显示的坑位数
  let maxSlots = 12
  if (props.selectedCloudDevice && props.selectedCloudDevice.name && props.selectedCloudDevice.name.toLowerCase().includes('p1')) {
    maxSlots = 24
  }
  
  for (let slotNum = 1; slotNum <= maxSlots; slotNum++) {
    const slotInstances = props.instances.filter(inst => inst.indexNum === slotNum)
    if (slotInstances.length > 0) {
      slotInstances.forEach((inst, index) => {
        result.push({
          slotNum,
          isFirstInSlot: index === 0, // 标记是否是该坑位的第一个实例
          instanceCount: slotInstances.length, // 该坑位的实例总数
          ...inst
        })
      })
      slotCountMap.set(slotNum, slotInstances.length)
    } else {
      result.push({
        slotNum,
        isFirstInSlot: true, // 只有一个空实例，所以是第一个
        instanceCount: 1,
        name: '',
        ip: '',
        image: '',
        createTime: '',
        status: 'shutdown',
        modelName: ''
      })
      slotCountMap.set(slotNum, 1)
    }
  }

  console.log('********groupedInstances:', result)
  return result
  
})

// 获取备份数量
const getBackupCount = (slotNum) => {
  if (!props.allInstances || !Array.isArray(props.allInstances)) return 0
  return props.allInstances.filter(inst => 
    inst.indexNum === slotNum && 
    inst.status !== 'running' &&
    inst.name
  ).length
}

// 格式化实例名称
const formatInstanceName = (name) => {
  if (!name) return name
  
  // 匹配格式：[任意字符]_数字_名称
  // 例如：p1e847b84af914895b56a14557d1813d_2_T00022222222 -> T00022222222
  // 例如：p1e847b84af914895b56a14557d1813d_4_sjz_cs -> sjz_cs
  const match = name.match(/^.+_\d+_(.+)$/)
  if (match && match.length > 1) {
    return match[1] // 只返回最后一个下划线后的名称部分
  }
  
  return name
}

// 获取镜像显示名称 - 完整版
const getImageDisplayName = (imageName) => {
  if (!imageName) return '未知镜像';
  const cleanedUrl = imageName.toLowerCase()
  
  // 1. 从镜像列表中查找匹配的镜像
  let matchedImage = null
  
  // 精确匹配：检查image.url是否与imageUrl完全匹配
  matchedImage = props.imageList.find(image => {
    return image.url && image.url.toLowerCase() === cleanedUrl
  })
  
  // 如果没有精确匹配，尝试模糊匹配：检查image.url是否是imageUrl的一部分
  if (!matchedImage) {
    matchedImage = props.imageList.find(image => {
      return image.url && cleanedUrl.includes(image.url.toLowerCase())
    })
  }
  
  // 如果没有找到，尝试反向匹配：检查imageUrl是否是image.url的一部分
  if (!matchedImage) {
    matchedImage = props.imageList.find(image => {
      return image.url && image.url.toLowerCase().includes(cleanedUrl)
    })
  }
  
  // 如果仍然没有找到，尝试匹配镜像名称
  if (!matchedImage) {
    // 从URL中提取镜像名称部分用于匹配
    const urlParts = cleanedUrl.split('/')
    const urlNamePart = urlParts[urlParts.length - 1]
    
    matchedImage = props.imageList.find(image => {
      return image.name && image.name.toLowerCase().includes(urlNamePart)
    })
  }
  
  if (matchedImage) {
    // 2. 如果找到匹配的镜像，返回其名称
    return matchedImage.name || matchedImage.url
  }
  
  // 3. 如果没有找到匹配的镜像，从URL中提取用户友好的名称
  
  // 处理本地文件路径
  if (cleanedUrl.includes('\\') || cleanedUrl.endsWith('.tar.gz')) {
    // 本地镜像路径，提取文件名
    const pathParts = cleanedUrl.split(/[\\/]/)
    let fileName = pathParts[pathParts.length - 1]
    // 移除文件扩展名
    fileName = fileName.replace('.tar.gz', '')
    fileName = fileName.replace('.tar', '')
    return fileName
  }
  
  // 处理Docker镜像URL
  // 例如：registry.magicloud.tech/magicloud/dobox-android13:Q1 -> dobox-android13:Q1
  // 例如：docker.io/library/nginx:latest -> nginx:latest
  const parts = cleanedUrl.split('/')
  if (parts.length > 0) {
    let imageName = parts[parts.length - 1]
    
    // 进一步优化：如果镜像名包含registry或magicloud等关键词，尝试提取更友好的名称
    const friendlyNameMatch = imageName.match(/([a-zA-Z0-9_-]+):([a-zA-Z0-9._-]+)$/)
    if (friendlyNameMatch) {
      return friendlyNameMatch[0] // 返回 镜像名:标签 格式
    }
    
    return imageName
  }
  
  // 4. 否则返回原始URL
  return imageUrl
};

// 格式化实例机型名称
const formatInstanceModel = (path) => {
  console.log('********formatInstanceModel:', path)
  if (!path) return ''
  
  // 匹配格式：[任意字符]_数字_名称
  // 例如：p1e847b84af914895b56a14557d1813d_2_T00022222222 -> T00022222222
  // 例如：p1e847b84af914895b56a14557d1813d_4_sjz_cs -> sjz_cs
  // 例如：/mmc/data/.../22111317PG -> 22111317PG
  
  // 尝试从路径中提取最后一部分作为机型
  const pathParts = path.split('/')
  const lastPart = pathParts[pathParts.length - 1]
  
  // 检查是否是有效的机型名称（非空且不包含路径分隔符）
  if (lastPart && !lastPart.includes('/')) {
    return lastPart
  }
  
  return path // 无法提取时返回原始名称
}

// 云机管理相关函数
const handleBatchCreateDevice = () => {
  // 批量设备创建，不传递特定设备，模式为 'multi-device-batch'
  emit('showCreateDialog', null, 'multi-device-batch')
}

const handleScanComplete = () => {
  // 扫描完成，Dialog内部会处理状态
}

const showAddDeviceDialog = () => {
  // 更新已存在设备ID列表
  existingDeviceIds.value = new Set(props.devices.map(d => d.id))
  // 显示对话框，对话框打开后会自动开始扫描
  addDeviceDialogVisible.value = true
}

const handleDeviceAdded = (device) => {
  // 直接通知父组件处理设备添加
  emit('device-added', device)
}

const handleBatchAddDevices = (devices) => {
  emit('handleBatchAddDevices', devices)
}

const handleCloudDeviceSelect = (device) => {
  console.log('选择云机设备:', device)
  // 通过事件通知父组件选择了哪个设备
  emit('selectedCloudDeviceChange', device)
  // 加载云机实例
  localFetchAndroidContainers(device)
}

const localFetchAndroidContainers = async (device) => {
  loading.value = true
  try {
    // 通过事件通知父组件获取Android容器
    emit('fetchAndroidContainers', device, true)
    // 父组件会自动处理截图加载，不需要在这里调用
  } catch (error) {
    ElMessage.error('获取云机实例失败')
    console.error('获取云机实例失败:', error)
  } finally {
    loading.value = false
  }
}

// 截图加载逻辑已移至App.vue中，由父组件统一处理

const showCreateDialog = (device, mode, slotNum = null) => {
  // 调用父组件方法显示创建云机对话框
  emit('showCreateDialog', device, mode, slotNum)
}

const handleContextMenu = (event, slotNum) => {
  // 阻止默认右键菜单
  event.preventDefault()
  // 调用父组件方法处理右键菜单
  emit('handleContextMenu', event, slotNum)
}

// 处理ADB连接 - 内嵌终端
const handleADBConnect = (slotNumOrInstance) => {
  try {
    let instance
    let device

    if (typeof slotNumOrInstance === 'object') {
      // 批量模式：直接传入了实例对象
      instance = slotNumOrInstance
      // 构造设备对象，只需要ip
      if (instance.deviceIp) {
        device = { ip: instance.deviceIp }
      }
    } else {
      // 坑位模式：传入了槽位号
      const slotNum = slotNumOrInstance
      // 查找对应槽位的容器实例
      instance = props.instances.find(inst => inst.indexNum === slotNum)
      if (!instance) {
        ElMessage.error(`未找到槽位 ${slotNum} 的容器实例`)
        return
      }
      device = props.activeDevice || props.selectedCloudDevice
    }
    
    if (!device) {
      ElMessage.error('未选择设备')
      return
    }
    
    // 调用连接终端函数
    connectTerminal(instance, device)
  } catch (error) {
    console.error('处理ADB连接失败:', error)
    ElMessage.error('处理ADB连接失败: ' + error.message)
  }
}

// 关闭终端 - 确保资源正确释放，不受其他操作影响
const closeTerminal = () => {
  // 只有当终端实例存在时才执行关闭操作
  if (term) {
    try {
      // 清理旧的终端实例
      term.dispose()
      term = null
    } catch (error) {
      console.error('终端 dispose 失败:', error)
    }
  }
  
  // 关闭 WebSocket 连接
  if (socket) {
    try {
      socket.close()
      socket = null
    } catch (error) {
      console.error('WebSocket 关闭失败:', error)
    }
  }
  
  // 清除心跳定时器
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
  
  // 清理事件监听器
  while (terminalEventListeners.length > 0) {
    const listener = terminalEventListeners.pop()
    if (listener instanceof Function) {
      document.removeEventListener('keydown', listener)
    }
  }
  
  // 隐藏终端元素
  const terminalElement = document.getElementById('terminal')
  if (terminalElement) {
    terminalElement.style.display = 'none'
  }
  
  // 隐藏遮罩层
  const overlay = document.getElementById('terminal-overlay')
  if (overlay) {
    overlay.style.display = 'none'
  }
  
  // 重置终端可见状态
  terminalVisible.value = false
  
  // 清除终端实例ID
  terminalInstanceId = null
}

// 连接终端 - 确保终端独立性，不受其他操作影响
const connectTerminal = (instance, targetDevice = null) => {
  // 生成唯一的终端实例ID，确保每次连接都是独立的
  const newInstanceId = Date.now().toString() + Math.random().toString(36).substr(2, 9)
  terminalInstanceId = newInstanceId
  
  try {
    console.log('连接终端:', instance, '实例ID:', newInstanceId)
    
    // 先关闭旧的终端实例
    closeTerminal()
    
    // 重新设置实例ID
    terminalInstanceId = newInstanceId
    
    // 显示终端
    terminalVisible.value = true
    
    // 获取终端元素
    const terminalElement = document.getElementById('terminal')
    if (!terminalElement) {
      ElMessage.error('未找到终端元素')
      terminalInstanceId = null
      return
    }
    
    // 清空终端元素内容，重新构建结构
    terminalElement.innerHTML = ''
    
    // 创建标题栏
    const header = document.createElement('div')
    header.className = 'terminal-header'
    
    // 创建标题文本
    const title = document.createElement('div')
    title.className = 'terminal-title'
    
    // 提取实例名称中的T000x部分
    let displayName = instance.name || instance.id
    const nameParts = displayName.split('_')
    if (nameParts.length > 0) {
      // 从名称中提取最后一部分作为显示名称，例如从"8569541a74175bfe052739c4321ea31b_1_T0001"中提取"T0001"
      displayName = nameParts[nameParts.length - 1]
    }
    
    title.innerText = `终端 - ${displayName}`
    
    // 创建关闭按钮
    const closeButton = document.createElement('button')
    closeButton.className = 'terminal-close-btn'
    closeButton.innerHTML = '关闭'
    closeButton.style.padding = '2px 8px'
    closeButton.style.backgroundColor = 'transparent'
    closeButton.style.color = '#ccc'
    closeButton.style.border = '1px solid #444'
    closeButton.style.borderRadius = '3px'
    closeButton.style.cursor = 'pointer'
    closeButton.style.fontSize = '12px'
    closeButton.style.lineHeight = '16px'
    closeButton.style.textAlign = 'center'
    closeButton.style.transition = 'all 0.2s ease'
    
    // 绑定关闭按钮点击事件，添加实例ID检查，确保只关闭当前实例
    closeButton.onclick = () => {
      if (terminalInstanceId === newInstanceId) {
        closeTerminal()
      }
    }
    
    // 添加悬停效果
    closeButton.onmouseenter = () => {
      closeButton.style.backgroundColor = '#ff4d4f'
      closeButton.style.color = 'white'
      closeButton.style.borderColor = '#ff4d4f'
    }
    
    closeButton.onmouseleave = () => {
      closeButton.style.backgroundColor = 'transparent'
      closeButton.style.color = '#ccc'
      closeButton.style.borderColor = '#444'
    }
    
    // 将标题和关闭按钮添加到标题栏
    header.appendChild(title)
    header.appendChild(closeButton)
    
    // 创建内容区域
    const content = document.createElement('div')
    content.className = 'terminal-content'
    content.id = 'terminal-content'
    
    // 将标题栏和内容区域添加到终端容器
    terminalElement.appendChild(header)
    terminalElement.appendChild(content)
    
    // 显示终端元素和遮罩层
    terminalElement.style.display = 'block'
    
    // 显示遮罩层
    const overlay = document.getElementById('terminal-overlay')
    if (overlay) {
      overlay.style.display = 'block'
    }
    
    // 初始化 Xterm，将其附加到内容区域，使用黑色主题，类似Windows命令提示符
    term = new Terminal({
      cursorBlink: true, 
      theme: {
        background: '#000000',
        foreground: '#cccccc',
        cursor: '#cccccc',
        selection: '#444444',
        selectionForeground: '#ffffff',
        black: '#000000',
        red: '#ff4747',
        green: '#00ff00',
        yellow: '#ffff00',
        blue: '#4799ff',
        magenta: '#ff00ff',
        cyan: '#00ffff',
        white: '#cccccc',
        brightBlack: '#808080',
        brightRed: '#ff4747',
        brightGreen: '#00ff00',
        brightYellow: '#ffff00',
        brightBlue: '#4799ff',
        brightMagenta: '#ff00ff',
        brightCyan: '#00ffff',
        brightWhite: '#ffffff'
      },
      // 启用复制粘贴功能
      allowProposedApi: true,
      disableStdin: false,
      rendererType: 'canvas', // 使用canvas渲染器，提供更好的复制支持
      // 启用鼠标支持
      mouseSelectionMode: true,
      allowTransparency: true
    })
    fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(content)
    fitAddon.fit()
    term.focus()
    
    // 确保复制功能正常工作
    term.onSelectionChange(() => {
      if (term.getSelection()) {
        // 当有选中内容时，将其复制到剪贴板
        navigator.clipboard.writeText(term.getSelection())
          .catch(err => {
            console.error('复制到剪贴板失败:', err)
          })
      }
    })
    
    // 添加键盘事件监听，支持复制粘贴
    // Electron环境中xterm.js不会自动处理粘贴，需要手动拦截
    term.attachCustomKeyEventHandler((event) => {
      // 只处理 keydown 事件，避免 keyup 时重复触发
      if (event.type !== 'keydown') return true
      
      // 支持 Ctrl+C 复制（仅在有选中文本时拦截，否则让xterm发送SIGINT）
      if (event.ctrlKey && event.key === 'c') {
        if (term.hasSelection()) {
          navigator.clipboard.writeText(term.getSelection())
            .catch(err => {
              console.error('复制到剪贴板失败:', err)
            })
          return false
        }
      }
      // 支持 Ctrl+V 粘贴
      else if (event.ctrlKey && event.key === 'v') {
        // 阻止浏览器默认的paste事件，防止xterm通过onData重复发送
        event.preventDefault()
        event.stopPropagation()
        
        navigator.clipboard.readText()
          .then(text => {
            if (text && socket && socket.readyState === WebSocket.OPEN) {
              // 清理不可见的Unicode字符（零宽空格、BOM、零宽连接符等）
              const cleanText = text
                .replace(/[\u200B\u200C\u200D\uFEFF\u00A0]/g, '')
                .replace(/\r\n/g, '\n')
                .replace(/\r/g, '\n')
              
              if (cleanText) {
                sendJson({ type: "stdin", data: cleanText })
              }
            }
          })
          .catch(err => {
            console.error('从剪贴板读取失败:', err)
          })
        // 返回 false 阻止 xterm 处理此按键
        return false
      }
      return true
    })
    
    // 拦截终端容器上的paste事件，防止xterm内部的paste处理导致重复发送
    content.addEventListener('paste', (event) => {
      event.preventDefault()
      event.stopPropagation()
    })
    
    // 不显示连接信息，保持终端界面简洁
    
    // 连接 WebSocket - 使用设备 IP + 端口（支持已含端口的外网映射地址）
    // 查找对应的设备
    const device = targetDevice || props.activeDevice || props.selectedCloudDevice
    if (!device) {
      term.write('\r\n\x1b[31m未找到设备信息\x1b[0m\r\n')
      return
    }
    
    const wsProtocol = location.protocol === 'https:' ? 'wss://' : 'ws://'
    let wsUrl = `${wsProtocol}${getDeviceAddr(device.ip)}/link/exec`
    
    const savedPassword = localStorage.getItem('devicePasswords')
    const passwords = JSON.parse(savedPassword || '{}');
    const password = passwords[device.ip] || null
    if (password) {
      wsUrl = `${wsProtocol}admin:${password}@${getDeviceAddr(device.ip)}/link/exec`
    }
    console.log('连接 WebSocket:', wsUrl)
    
    socket = new WebSocket(wsUrl)
    
    socket.onopen = () => {
      // 检查是否是当前实例，防止操作已关闭的终端
      if (terminalInstanceId !== newInstanceId) {
        return
      }
      
      console.log('WebSocket 连接成功')
      socket.binaryType = 'arraybuffer'
      
      // 发送登录指令
      sendJson({
        type: "login",
        container_id: instance.name || instance.id,
        shell: "/bin/sd"
      })
      
      sendResize()
      
      // 心跳
      heartbeatTimer = setInterval(() => {
        // 检查是否是当前实例
        if (terminalInstanceId === newInstanceId && socket.readyState === WebSocket.OPEN) {
          sendJson({ type: "heartbeat" })
        }
      }, 30000)
    }
    
    socket.onmessage = (event) => {
      // 检查是否是当前实例，防止操作已关闭的终端
      if (terminalInstanceId !== newInstanceId) {
        return
      }
      
      if (event.data instanceof ArrayBuffer) {
        // 将ArrayBuffer转换为字符串，检查是否包含"connected to container"
        const message = new TextDecoder().decode(event.data)
        if (!message.includes('connected to container')) {
          term.write(new Uint8Array(event.data))
        }
      } else {
        // 兼容旧的文本模式（以防万一）
        if (typeof event.data === 'string' && !event.data.includes('connected to container')) {
          term.write(event.data)
        }
      }
    }
    
    socket.onclose = (event) => {
      // 检查是否是当前实例，防止操作已关闭的终端
      if (terminalInstanceId !== newInstanceId) {
        return
      }
      
      console.log('WebSocket 连接关闭:', event)
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
      
      // 3秒后自动关闭终端，检查是否是当前实例
      setTimeout(() => {
        if (terminalInstanceId === newInstanceId) {
          closeTerminal()
        }
      }, 3000)
    }
    
    socket.onerror = (error) => {
      // 检查是否是当前实例，防止操作已关闭的终端
      if (terminalInstanceId !== newInstanceId) {
        return
      }
      
      console.error('WebSocket 连接错误:', error)
      
      // 3秒后自动关闭终端，检查是否是当前实例
      setTimeout(() => {
        if (terminalInstanceId === newInstanceId) {
          closeTerminal()
        }
      }, 3000)
    }
    
    // 输入监听（包括键盘输入和xterm原生粘贴）
    term.onData(data => {
      // 检查是否是当前实例，防止操作已关闭的终端
      if (terminalInstanceId !== newInstanceId || !socket || socket.readyState !== WebSocket.OPEN) {
        return
      }
      
      // 对于多字符输入（通常是粘贴），清理不可见的Unicode字符
      let cleanData = data
      if (data.length > 1) {
        cleanData = data
          .replace(/[\u200B\u200C\u200D\uFEFF\u00A0]/g, '')  // 移除零宽字符和非断行空格
          .replace(/\r\n/g, '\n')  // 统一换行符
          .replace(/\r/g, '\n')    // 处理旧式Mac换行
      }
      
      if (cleanData) {
        sendJson({ type: "stdin", data: cleanData })
      }
    })
    
    // 窗口调整 - 使用闭包保存当前实例ID，确保只影响当前终端
    const handleResize = () => {
      // 检查是否是当前实例，防止操作已关闭的终端
      if (terminalInstanceId === newInstanceId && fitAddon && term && terminalVisible.value) {
        fitAddon.fit()
        sendResize()
      }
    }
    
    window.addEventListener('resize', handleResize)
    
    // 注意：移除了 dispose 事件监听，因为 xterm.js v5.3.0 不支持 term.on() 方式
    // 窗口调整事件在 window 上监听，终端关闭时会在 closeTerminal 函数中清理
    
  } catch (error) {
    console.error('连接终端失败:', error)
    ElMessage.error('连接终端失败: ' + error.message)
    
    // 检查是否是当前实例，防止关闭其他终端
    if (terminalInstanceId === newInstanceId) {
      closeTerminal()
    }
  }
}

// 发送JSON数据到WebSocket
const sendJson = (obj) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(obj))
  }
}

// 发送窗口大小调整
const sendResize = () => {
  if (fitAddon && term && socket.readyState === WebSocket.OPEN) {
    const dims = fitAddon.proposeDimensions()
    if (dims) {
      sendJson({ type: "resize", cols: dims.cols, rows: dims.rows })
    }
  }
}

const showBackupList = (slotNum) => {
  // 调用父组件方法显示备份列表
  emit('showBackupList', slotNum)
}

// 批量清理投屏
const handleBatchCleanupProjection = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要关闭所有投屏窗口和投屏进程吗？',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    // 调用清理窗口
    await api.cleanupProjectionWindows()
    // 调用清理进程
    await api.cleanupProjectionProcesses()
    
    ElMessage.success('批量关闭投屏成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量关闭投屏失败:', error)
      ElMessage.error('批量关闭投屏失败')
    }
  }
}

const handleBatchAction = (action) => {
  console.log('[CloudManagement] handleBatchAction 被调用, action:', action, ', cardOrientation:', cardOrientation.value)
  // 调用父组件方法处理批量操作，根据模式传递不同的选中信息
  if (props.cloudManageMode === 'slot') {
    // 坑位模式：传递选中的坑位信息
    // 对于 projection-control 和 projection，额外传递卡片方向信息
    if (action === 'projection-control' || action === 'projection') {
      console.log('[CloudManagement] 传递 cardOrientation 给父组件:', cardOrientation.value)
      emit('handleBatchAction', action, selectedSlotCloudMachines.value, cardOrientation.value)
    } else {
      emit('handleBatchAction', action, selectedSlotCloudMachines.value)
    }
  } else {
    // 批量模式：传递选中的云机信息
    // 对于 projection-control 和 projection，额外传递卡片方向信息
    if (action === 'projection-control' || action === 'projection') {
      console.log('[CloudManagement] 传递 cardOrientation 给父组件:', cardOrientation.value)
      emit('handleBatchAction', action, props.selectedCloudMachines, cardOrientation.value)
    } else {
      emit('handleBatchAction', action, props.selectedCloudMachines)
    }
  }
}

const handleSlotCloudMachinesSelectAll = (val) => {
  // 处理坑位云机全选
  if (val && props.selectedCloudDevice) {
    selectedSlotCloudMachines.value = Array.from({ length: (props.selectedCloudDevice.name?.toLowerCase().includes('p1') ? 24 : 12) }, (_, i) => i + 1)
  } else {
    selectedSlotCloudMachines.value = []
  }
}

const handleSlotSelectionChange = (selection) => {
  console.log('坑位云机选择变化:', selection)
  // 同步列表选择到 selectedSlotCloudMachines
  selectedSlotCloudMachines.value = selection.map(row => row.slotNum)
  // 处理列表模式下的选择变化
  // emit('handleSlotSelectionChange', selection)
}

const handleDeleteContainer = (container) => {
  // 调用父组件方法删除容器
  emit('handleDeleteContainer', container)
}

// 显示复制云机对话框
const showCopyDialog = (row) => {
  // 优先通过 row.androidType 判断版本，V2 -> v2，否则 v3
  let version = 'v3'
  if (row.androidType && row.androidType.toLowerCase() === 'v2') {
    version = 'v2'
  } else if (row.androidType && row.androidType.toLowerCase() === 'v3') {
    version = 'v3'
  } else {
    // 兜底：从设备名称中判断
    const device = props.selectedCloudDevice
    version = device?.version || (device?.name?.includes('_v2') ? 'v2' : 'v3')
  }
  copyForm.value = {
    name: row.name,
    indexNum: row.slotNum,
    count: 1,
    version
  }
  copyDialogVisible.value = true
}

// 重置复制表单
const resetCopyForm = () => {
  copyForm.value = { name: '', indexNum: 1, count: 1, version: 'v3' }
}

// 确认复制云机 —— 提交给父组件任务队列处理
const confirmCopyMachine = () => {
  const device = props.selectedCloudDevice
  if (!device) {
    ElMessage.error('未选择设备')
    return
  }
  const { name, indexNum, count, version } = copyForm.value
  if (!name) {
    ElMessage.error('云机名称不能为空')
    return
  }
  emit('startCopyTask', { device, name, indexNum, count, version })
  copyDialogVisible.value = false
}

const startProjection = async (device, instance) => {
  try {
    // 根据 cardOrientation 计算 orient 值
    // vertical(竖屏) -> orient = 0 (portrait)
    // horizontal(横屏) -> orient = 1 (landscape)
    const customOrient = cardOrientation.value === 'horizontal' ? 1 : 0
    console.log('[CloudManagement] startProjection 被调用')
    console.log('[CloudManagement] cardOrientation.value:', cardOrientation.value)
    console.log('[CloudManagement] customOrient:', customOrient)
    console.log('[CloudManagement] device:', device)
    console.log('[CloudManagement] instance:', instance)
    
    // 调用 API 开始投屏，传递自定义方向
    await api.startProjection(device, instance, customOrient)
  } catch (error) {
    console.error('启动投屏失败:', error)
  }
}

// 根据勾选的 key 从 groups 数据中提取云机列表（复用逻辑）
const recalcSelectedMachines = (checkedKeys, groups) => {
  const selectedMachines = []
  const machineIds = new Set()
  const findSelectedItems = (nodes) => {
    nodes.forEach(node => {
      if (node.screenshot !== undefined && !node.cloudMachines && !node.devices) {
        // 云机节点
        if (checkedKeys.includes(node.id) && !machineIds.has(node.id)) {
          selectedMachines.push(node)
          machineIds.add(node.id)
        }
      } else if (node.cloudMachines) {
        // 设备节点
        if (checkedKeys.includes(node.id)) {
          node.cloudMachines.forEach(machine => {
            if (!machineIds.has(machine.id)) {
              selectedMachines.push(machine)
              machineIds.add(machine.id)
            }
          })
        } else {
          findSelectedItems(node.cloudMachines)
        }
      } else if (node.devices) {
        // 分组节点
        if (checkedKeys.includes(node.id)) {
          node.devices.forEach(device => {
            device.cloudMachines.forEach(machine => {
              if (!machineIds.has(machine.id)) {
                selectedMachines.push(machine)
                machineIds.add(machine.id)
              }
            })
          })
        } else {
          findSelectedItems(node.devices)
        }
      }
    })
  }
  findSelectedItems(groups)
  return selectedMachines
}

// 树形结构勾选事件处理
const handleTreeCheck = (data, checkedInfo) => {
  // 获取选中的节点ID数组
  const checkedKeys = checkedInfo.checkedKeys
  currentCheckedKeys = [...checkedKeys] // 保存当前勾选 key
  
  // 筛选出所有选中的云机节点和设备节点
  const selectedMachines = []
  const selectedDevices = []
  const machineIds = new Set() // 用于去重
  
  // 递归遍历所有节点，检查是否被选中
  const findSelectedItems = (nodes) => {
    nodes.forEach(node => {
      if (node.screenshot) {
        // 云机节点，检查是否被选中
        if (checkedKeys.includes(node.id) && !machineIds.has(node.id)) {
          selectedMachines.push(node)
          machineIds.add(node.id)
        }
      } else if (node.cloudMachines) {
        // 设备节点，检查是否被选中
        if (checkedKeys.includes(node.id)) {
          // 设备被选中，添加到选中设备列表
          selectedDevices.push(node)
          // 添加其下所有云机（去重）
          node.cloudMachines.forEach(machine => {
            if (!machineIds.has(machine.id)) {
              selectedMachines.push(machine)
              machineIds.add(machine.id)
            }
          })
        } else {
          // 递归检查云机节点
          findSelectedItems(node.cloudMachines)
        }
      } else if (node.devices) {
        // 分组节点，检查是否被选中
        if (checkedKeys.includes(node.id)) {
          // 分组被选中，添加其下所有设备和云机（去重）
          node.devices.forEach(device => {
            selectedDevices.push(device)
            device.cloudMachines.forEach(machine => {
              if (!machineIds.has(machine.id)) {
                selectedMachines.push(machine)
                machineIds.add(machine.id)
              }
            })
          })
        } else {
          // 递归检查设备节点
          findSelectedItems(node.devices)
        }
      }
    })
  }
  
  // 开始遍历所有分组
  findSelectedItems(props.cloudMachineGroups)
  
  // 更新选中的云机列表，确保每次都重新赋值
  selectedCloudMachines.value = [...selectedMachines]
  console.log('selectedCloudMachines:', selectedCloudMachines.value)
  
  // 通知父组件处理截图加载队列
  emit('handleTreeCheck', { 
    selectedMachines: [...selectedMachines], 
    selectedDevices: [...selectedDevices],
    treeSelectedKeys: [...checkedKeys]
  })
}

// 截图队列管理现在由父组件App.vue处理

const handleDrop = (draggingNode, dropNode, dropType) => {
  // 处理树形结构拖放
  return true
}

const handleNodeDrop = (draggingNode, dropNode, dropType, ev) => {
  // 处理树形结构拖放完成
  emit('handleNodeDrop', draggingNode, dropNode, dropType, ev)
}

const updateSelectedCloudMachines = () => {
  // 更新选中的云机列表
  selectedCloudMachines.value = []
  // 遍历树形结构选中的键，获取对应的云机
  // 这里需要根据实际数据结构实现
}

// 从设备名称中提取设备型号
const getDeviceTypeName = (deviceName) => {
  if (!deviceName) return 'unknown'
  // 从设备名称中提取型号，例如q1_v2 -> q1, p1_v3 -> p1
  const parts = deviceName.split('_')
  return parts[0] || 'unknown'
}

// 获取设备类型颜色
const getDeviceTypeColor = (deviceName) => {
  const deviceType = getDeviceTypeName(deviceName)
  const colorMap = {
    'q1': '#409EFF', // 蓝色
    'p1': '#67C23A', // 绿色
    'm48': '#E6A23C', // 黄色
    'c1': '#F56C6C', // 红色
    'a1': '#909399' // 灰色
  }
  return colorMap[deviceType] || '#909399' // 默认灰色
}

// 这些函数已经在前面定义过了，不需要重复定义

// 事件定义
const emit = defineEmits([
  'showAddDeviceDialog',
  'showCreateDialog',
  'handleContextMenu',
  'handleADBConnect',
  'showBackupList',
  'handleBatchAction',
  // 'handleSlotSelectionChange',
  'handleDeleteContainer',
  'startProjection',
  'handleNodeDrop',
  'handleTreeCheck',
  'cloudManageModeChange',
  'selectedCloudDeviceChange',
  'fetchAndroidContainers',
  'device-added',
  'handleBatchAddDevices',
  'moveDeviceToGroup',
  'startCopyTask'
])

// 属性接收
const props = defineProps({
  activeDevice: {
    type: Object,
    default: null
  },
  cloudManageMode: {
    type: String,
    default: 'slot'
  },
  cloudMachineGroups: {
    type: Array,
    default: () => []
  },
  selectedCloudDevice: {
    type: Object,
    default: null
  },
  instances: {
    type: Array,
    default: () => []
  },
  cloudMachines: {
    type: Array,
    default: () => []
  },
  cloudMachinesByName: {
    type: Map,
    default: () => new Map()
  },
  selectedCloudMachines: {
    type: Array,
    default: () => []
  },
  treeSelectedKeys: {
    type: Array,
    default: () => []
  },
  devices: {
    type: Array,
    default: () => []
  },
  devicesStatusCache: {
    type: Map,
    default: () => new Map()
  },
  loading: {
    type: Boolean,
    default: false
  },
  allInstances: {
    type: Array,
    default: () => []
  },
  imageList: {
    type: Array,
    default: () => []
  },
  isBatchProjectionControlling: {
    type: Boolean,
    default: false
  },
  screenshotCache: {
    type: Map,
    default: () => new Map() // Map<"ip_containerName", base64DataURL>
  }
})

// 监听属性变化
watch(() => props.instances, (newInstances) => {
  instances.value = newInstances
}, { deep: true })

// 监听云机分组数据变化
watch(() => props.cloudMachineGroups, (newGroups) => {
  if (JSON.stringify(newGroups) !== JSON.stringify(previousGroups.value)) {
    console.log('云机分组数据变化:', newGroups)
    previousGroups.value = [...newGroups]
    // 分组数据更新后，用当前勾选 key 从新数据中重新提取 selectedCloudMachines
    // 避免截图区域仍显示旧对象引用（status 等字段不更新）
    if (currentCheckedKeys.length > 0) {
      selectedCloudMachines.value = recalcSelectedMachines(currentCheckedKeys, newGroups)
    }
  }
}, { deep: true })

// 监听选中云机设备变化
watch(() => props.selectedCloudDevice, (newDevice) => {
  if (newDevice?.id !== previousDeviceId.value) {
    console.log('选中云机设备变化:', newDevice)
    previousDeviceId.value = newDevice?.id
  }
}, { deep: true })

// 处理云机管理模式变化
const handleCloudManageModeChange = (mode) => {
  console.log('切换云机管理模式:', mode)
  emit('cloudManageModeChange', mode)
}

// 监听云机管理模式变化
watch(() => props.cloudManageMode, (newMode, oldMode) => {
  if (oldMode !== newMode) {
    console.log('云机管理模式切换:', oldMode, '->', newMode)
  }
  // 切换模式时清理选中状态
  if (newMode === 'slot') {
    selectedSlotCloudMachines.value = []
    isSlotCloudMachinesAllSelected.value = false
    isSlotCloudMachinesIndeterminate.value = false
  } else {
    selectedCloudMachines.value = []
  }
}, { deep: true })

watch(() => props.cloudMachines, (newCloudMachines) => {
  cloudMachines.value = newCloudMachines
}, { deep: true })

watch(() => props.cloudMachinesByName, (newCloudMachinesByName) => {
  cloudMachinesByName.value = newCloudMachinesByName
}, { deep: true })

watch(() => props.selectedCloudMachines, (newSelectedCloudMachines) => {
  selectedCloudMachines.value = newSelectedCloudMachines
}, { deep: true })

watch(() => props.devices, (newDevices) => {
  devices.value = newDevices
}, { deep: true, immediate: true })

watch(() => props.devicesStatusCache, (newDevicesStatusCache) => {
  devicesStatusCache.value = newDevicesStatusCache
  statusCacheVersion.value++ // 递增触发器，强制 computed 重新求值
}, { deep: true, immediate: true })

// 已注释：不再需要同步父组件的 scanningDevices，因为使用事件来控制
// watch(() => props.scanningDevices, (newScanningDevices) => {
//   scanningDevices.value = newScanningDevices
// })

watch(() => props.loading, (newLoading) => {
  loading.value = newLoading
})

// 截图缩放保存和加载函数
const saveScreenshotScale = () => {
  localStorage.setItem('cloudManagement_screenshotScale', screenshotScale.value.toString())
  console.log('已保存截图缩放比例:', screenshotScale.value + '%')
}

const loadScreenshotScale = () => {
  const savedScale = localStorage.getItem('cloudManagement_screenshotScale')
  if (savedScale) {
    screenshotScale.value = parseInt(savedScale)
    console.log('已加载截图缩放比例:', screenshotScale.value + '%')
  }
}

// 生命周期钩子
onMounted(() => {
  // 组件挂载时的初始化
  loadScreenshotScale() // 加载保存的截图缩放比例
})

onBeforeUnmount(() => {
  // 组件卸载前的清理工作
})
</script>

<style scoped>
/* 云机网格样式 */
.cloud-machine-grid {
  --screenshot-scale: 1; /* 默认缩放比例 */
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(calc(150px * var(--screenshot-scale)), 1fr));
  gap: calc(10px * var(--screenshot-scale));
  margin-top: 20px;
}

/* 竖屏方向网格 - 默认 */
.cloud-machine-grid.orientation-vertical {
  grid-template-columns: repeat(auto-fill, minmax(calc(150px * var(--screenshot-scale)), 1fr));
}

/* 横屏方向网格 - 适配 16:9 宽屏比例 */
.cloud-machine-grid.orientation-horizontal {
  grid-template-columns: repeat(auto-fill, minmax(calc(200px * var(--screenshot-scale)), 1fr));
}


/* 云机卡片 - 基础样式 */
.cloud-machine-card {
  margin-bottom: 0;
  height: auto;
  max-width: calc(200px * var(--screenshot-scale, 1)); /* 限制最大宽度，防止被grid拉伸 */
  border-radius: calc(8px * var(--screenshot-scale));
  overflow: hidden;
  transition: all 0.3s ease;
  position: relative;
  padding: 0;
  /* 移除卡片的默认内边距 */
  --el-card-padding: 0;
}

/* 云机卡片内容容器 - 统一设置，移除所有高度限制 */
.cloud-machine-card .el-card__body {
  padding: 0 !important;
  height: auto !important;
  min-height: 0 !important; /* 强制移除最小高度限制 */
  max-height: none !important;
}

/* 横屏卡片的 body - 再次确保不受高度限制 */
.cloud-machine-card.card-horizontal .el-card__body {
  min-height: 0 !important;
  height: auto !important;
}

/* 云机卡片内容 */
.cloud-machine-card .el-card__body > * {
  width: 100%;
  height: auto;
}

/* 云机信息和按钮部分 */
.cloud-machine-info-small {
  padding: calc(10px * var(--screenshot-scale, 1));
  background: #f8f9fa;
  border-top: max(1px, calc(1px * var(--screenshot-scale, 1))) solid #e9ecef;
}

/* 云机操作按钮 */
.cloud-machine-actions {
  display: flex;
  justify-content: center;
  gap: calc(8px * var(--screenshot-scale, 1));
  width: 100%;
}

/* 强制覆盖 el-space 的固定间距 */
.cloud-machine-info-small :deep(.el-space) {
  gap: calc(8px * var(--screenshot-scale, 1)) !important;
}

.cloud-machine-info-small :deep(.el-space__item) {
  margin-right: 0 !important;
}

/* 按钮样式 - 根据缩放调整字体和padding，但保持合理的高度 */
.cloud-machine-info-small :deep(.el-button--small) {
  font-size: max(8px, calc(12px * var(--screenshot-scale, 1)));
  padding: calc(4px * var(--screenshot-scale, 1)) calc(8px * var(--screenshot-scale, 1));
  line-height: 1.5;
}

/* 在极小缩放时调整按钮样式 */
@media (min-width: 1px) {
  .cloud-machine-grid :deep(.el-button--small) {
    white-space: nowrap;
  }
}

/* 标签样式 - 跟随缩放但保持协调 */
.cloud-machine-card-header-light :deep(.el-tag) {
  font-size: max(7px, calc(10px * var(--screenshot-scale, 1)));
  padding: 0 calc(4px * var(--screenshot-scale, 1));
  line-height: 1.5;
  min-width: unset !important; /* 移除固定最小宽度 */
}

/* 状态标签在网格中也要跟随缩放 */
.cloud-machine-grid .status-tag-normal {
  font-size: max(7px, calc(10px * var(--screenshot-scale, 1))) !important;
  padding: 0 calc(4px * var(--screenshot-scale, 1)) !important;
}

/* 复选框样式 - 整体缩放更协调 */
.cloud-machine-card-header-light :deep(.el-checkbox) {
  /* 使用transform scale会让复选框整体缩放，包括label */
  font-size: max(7px, calc(12px * var(--screenshot-scale, 1)));
  flex-shrink: 0; /* 防止复选框被压缩 */
  line-height: 1.5; /* 与header保持一致 */
  margin: 0; /* 移除默认margin */
  height: auto; /* 让高度自适应 */
  display: flex; /* 使用flex布局 */
  align-items: center; /* 垂直居中 */
}

.cloud-machine-card-header-light :deep(.el-checkbox__label) {
  font-size: max(7px, calc(10px * var(--screenshot-scale, 1)));
  padding-left: calc(2px * var(--screenshot-scale, 1));
 /*line-height: 1.5;  与header保持一致 */
}

.cloud-machine-card-header-light :deep(.el-checkbox__inner) {
  width: max(10px, calc(14px * var(--screenshot-scale, 1)));
  height: max(10px, calc(14px * var(--screenshot-scale, 1)));
}

.cloud-machine-card-header-light :deep(.el-checkbox__inner::after) {
  width: max(2px, calc(3px * var(--screenshot-scale, 1)));
  height: max(5px, calc(7px * var(--screenshot-scale, 1)));
  left: calc(4px * var(--screenshot-scale, 1));
}

/* 云机截图样式 */
.cloud-machine-screenshot-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  display: block;
}

/* 云机截图容器 - 基础样式，移除 min-height 让 aspect-ratio 生效 */
.cloud-machine-screenshot-large {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  padding: 0;
  position: relative;
  overflow: hidden;
}

/* 云机截图容器中的图片 */
.cloud-machine-screenshot-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  position: absolute;
  top: 0;
  left: 0;
}

.cloud-machine-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.cloud-machine-card-header-light {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: max(7px, calc(11px * var(--screenshot-scale, 1)));
  font-weight: 600;
  padding: calc(4px * var(--screenshot-scale, 1)) calc(8px * var(--screenshot-scale, 1));
  background: #f8f9fa;
  color: #495057;
  margin: 0;
  border-bottom: max(1px, calc(1px * var(--screenshot-scale, 1))) solid #e9ecef;
  line-height: 1.5;
}

.instance-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin: 0 calc(4px * var(--screenshot-scale, 1));
  text-align: center;
  font-size: max(7px, calc(11px * var(--screenshot-scale, 1)));
  line-height: 1.5; /* 与header保持一致 */
  display: inline-block; /* 确保不会产生额外高度 */
  vertical-align: middle; /* 垂直居中对齐 */
}

/* 仅在云机卡片范围内移除默认padding */
.cloud-machine-card .el-card__header {
  padding: 0 !important;
  border-bottom: none !important;
}

/* 增大截图区域，适配180x320截图，恢复层次感背景 */
.cloud-machine-screenshot-large {
  width: 100%;
  /* 默认竖屏比例 9:16 (180:320 或 360:640)，横屏时会被 card-horizontal 覆盖为 16:9 */
  aspect-ratio: 9 / 16;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

/* 横屏卡片样式 - 标准 16:9 宽屏比例 */
.cloud-machine-card.card-horizontal .cloud-machine-screenshot-large {
  aspect-ratio: 16 / 9 !important; /* 横屏：宽:高 = 16:9 */
  height: auto !important;
  min-height: unset !important;
}

.screenshot-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f0f2f5;
  color: #909399;
  font-size: max(8px, calc(14px * var(--screenshot-scale, 1)));
}

.screenshot-offline {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f0f2f5;
  color: #909399;
  font-size: max(8px, calc(14px * var(--screenshot-scale, 1)));
}

.screenshot-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f0f2f5;
  color: #909399;
  font-size: max(8px, calc(14px * var(--screenshot-scale, 1)));
}

.screenshot-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f56c6c;
  color: white;
  font-size: max(8px, calc(14px * var(--screenshot-scale, 1)));
}

.switching-backup-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: white;
  z-index: 10;
}

.floating-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  margin-bottom: 16px;
}

.batch-actions {
  display: flex;
  gap: 3px !important;
}

.status-tag-normal {
  font-size: 12px;
  padding: 0 8px;
}

.small-plus-btn {
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.device-create-btn {
  margin-right: auto;
}

/* .cloud-machine-table {
  width: 100%;
} */

.cloud-machine-screenshot-small {
  width: 100px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background-color: #f0f2f5;
}

.cloud-machine-screenshot-img-small {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.screenshot-loading-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f0f2f5;
  color: #909399;
}

.screenshot-error-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f56c6c;
  color: white;
}

.screenshot-empty-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #f0f2f5;
  color: #909399;
}

.device-ip-text {
  margin: 0 12px;
  font-weight: bold;
}

.native-scroll-container {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

.table-container {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

/* 设备菜单样式 */
.device-menu {
  overflow-y: auto;
}

.device-menu-item {
  padding: 8px 16px;
  border-bottom: 1px solid #f0f0f0;
  cursor: grab;
}

.device-menu-item:active {
  cursor: grabbing;
}

.device-menu-item.is-active {
  background-color: #ecf5ff;
  color: #409EFF;
}

/* 分组设备列表样式 */
.grouped-device-list {
  padding: 0;
}

.group-separator {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
  font-size: 13px;
  font-weight: 500;
  color: #606266;
  cursor: pointer;
  transition: background-color 0.2s;
  /* display: inline-block; */
  max-width: 260px;
  /* overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle; */
}

.group-separator-span {
  display: inline-block;
  /* max-width: 260px; */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.group-separator:hover {
  background-color: #e4e7ed;
}

.group-separator.drag-over {
  background-color: #409EFF !important;
  border-color: #409EFF;
  color: #fff;
}

.drop-hint {
  margin-left: auto;
  font-size: 11px;
  color: #409EFF;
  font-weight: normal;
}

.group-separator.drag-over .drop-hint {
  color: #fff;
}

.group-add-btn {
  margin-left: auto;
  opacity: 0.6;
  transition: opacity 0.2s, color 0.2s;
}

.group-add-btn:hover {
  opacity: 1;
}

.group-separator.drag-over .group-add-btn {
  opacity: 0.3;
  color: #fff;
}

.tree-node-label {
  /* display: flex; */
  /* align-items: center; */
  display: inline-block;
  min-width: 56px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.group-devices {
  transition: height 0.2s ease;
}

.device-menu-item-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.device-ip {
  flex: 1;
  margin: 0 8px;
}

.batch-create-btn {
  margin-left: auto;
}

/* 模式切换文字按钮样式 */
.mode-switch-text {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #606266;
  background-color: transparent;
  border-radius: 6px;
  padding: 4px 8px;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s ease;
}

.mode-switch-text:hover {
  background-color: #f5f7fa;
}

.mode-switch-text span {
  padding: 2px 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.mode-switch-text span.active {
  background-color: #409eff;
  color: #fff;
  font-weight: 500;
}

.mode-switch-text .divider {
  padding: 0 4px;
  color: #c0c4cc;
  font-weight: 300;
}

/* 搜索框样式 */
.device-search-box {
  margin-bottom: 12px;
  padding: 0 4px;
}

.device-search-box .el-input {
  width: 100%;
}

/* 设备标签样式 */
.device-list-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.device-tab {
  padding: 4px 12px;
  border-radius: 4px;
  transition: all 0.3s;
}

.device-tab.active {
  background-color: #ecf5ff;
  color: #409eff;
  font-weight: bold;
}

/* 树形结构样式 */
.tree-node-label {
  font-size: 14px;
}

/* 批量模式容器 */
.batch-mode-container {
  padding: 12px;
  background-color: #fafafa;
  border-radius: 4px;
}

/* 4K屏幕适配 - 放大树形选择框 */
@media screen and (min-width: 2560px) {
  /* 搜索框放大 */
  .device-search-box .el-input {
    font-size: 18px;
  }
  :deep(.device-search-box .el-input__wrapper) {
    padding: 6px 14px;
    min-height: 44px;
  }
  :deep(.device-search-box .el-input__inner) {
    font-size: 18px;
    height: 44px;
    line-height: 44px;
  }
  :deep(.device-search-box .el-input__prefix .el-icon) {
    font-size: 20px;
  }

  /* 坑位模式：分组标题行放大 */
  .group-separator {
    padding: 12px 16px;
    font-size: 18px;
    max-width: none;
  }
  .group-separator .el-icon {
    font-size: 18px;
  }

  /* 坑位模式：设备列表项放大 */
  .device-menu-item {
    padding: 12px 20px;
  }
  .device-ip {
    font-size: 17px;
  }
  .device-type-icon {
    width: 30px !important;
    height: 30px !important;
    font-size: 16px !important;
    border-radius: 6px !important;
  }

  /* 批量模式：el-tree 整体放大 */
  :deep(.batch-mode-container .el-tree-node__content) {
    height: 44px;
    line-height: 44px;
    font-size: 17px;
  }
  :deep(.batch-mode-container .el-tree-node__expand-icon) {
    font-size: 18px;
    padding: 8px;
  }
  :deep(.batch-mode-container .el-checkbox__inner) {
    transform: scale(1.5);
    transform-origin: center center;
  }
  :deep(.batch-mode-container .el-checkbox) {
    margin-right: 14px;
  }
  .tree-node-label {
    font-size: 17px;
  }
  .group-add-btn {
    font-size: 16px;
  }

  /* 模式切换按钮放大 */
  .mode-switch-text {
    font-size: 18px;
    padding: 6px 12px;
  }
  .mode-switch-text span {
    padding: 4px 12px;
  }
}
:deep(.el-table .cell) {
  display: block;
}
:deep(.el-checkbox__label) {
  display: flex;
  align-items: center;
}

/* 最终覆盖规则 - 确保没有 min-height 限制 */
.cloud-machine-card .el-card__body,
.cloud-machine-card.card-vertical .el-card__body,
.cloud-machine-card.card-horizontal .el-card__body {
  min-height: 0 !important;
  height: auto !important;
}

/* 缩放功能样式 - 已被滑动条动态缩放替代，这些旧样式不再需要 */

.cloud-machine-container {
  height: calc(100vh - 220px);
}

</style>