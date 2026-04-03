<template>
    <div class="backup-management-container">
        <div class="backup-content">
            <el-tabs v-model="activeTab" class="backup-tabs" @tab-change="handleTabChange">
                <el-tab-pane :label="t('backup.backupModel')" name="model">
                    <!-- 提示信息 -->
                    <el-alert
                      :title="t('backup.tip')"
                      type="info"
                      :closable="false"
                      show-icon
                      style="margin: 15px;padding: 16px;"
                    />
                    <div class="tab-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ t('backup.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="t('backup.group')">
                                    <el-option :label="t('backup.all')" value="全部"></el-option>
                                    <el-option v-for="group in deviceGroups" :key="group" :label="group"
                                        :value="group"></el-option>
                                </el-select>
                            </div>
                            <div class="device-list">
                                <div v-for="device in deviceList" :key="device.ip"
                                    :class="['device-item', { active: selectedDeviceIP === device.ip }]"
                                    @click="selectDevice(device)">
                                    <span class="device-ip">{{ device.ip }}</span>
                                </div>
                            </div>
                        </div>
                        <div class="right-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ t('backup.backupModelList') }}</span>
                                <div style="display: flex; gap: 8px; align-items: center;">
                                    <el-button size="medium" @click="handleOpenBackupModelDir">{{ t('backup.openLocalBackupModel') }}</el-button>
                                    <el-input v-model="modelSearchText" :placeholder="t('backup.searchModelName')" size="medium"
                                        style="width: 200px;" clearable @clear="handleModelSearchClear">
                                        <template #prefix>
                                            <el-icon>
                                                <Search />
                                            </el-icon>
                                        </template>
                                    </el-input>
                                    <el-button type="primary" size="medium" @click="handleAddModel">
                                        <el-icon>
                                            <Plus />
                                        </el-icon> {{ t('backup.add') }}
                                    </el-button>
                                </div>
                            </div>
                            <el-table :data="paginatedBackupModelList" style="width: 100%" v-loading="loading" stripe>
                                <el-table-column type="index" :label="t('backup.index')" align="center" width="100" />
                                <el-table-column prop="name" :label="t('backup.modelName')" align="center" />
                                <el-table-column :label="t('backup.operation')" align="center">
                                    <template #default="scope">
                                        <el-button v-if="isModelLocalExists(scope.row.name)" size="mini" type="success"
                                            @click="handleImportMachine(scope.row)">
                                            {{ t('backup.import') }}
                                        </el-button>
                                        <el-button v-else size="mini" type="primary"
                                            @click="handleExportMachine(scope.row)">
                                            {{ t('backup.export') }}
                                        </el-button>
                                        <el-button size="mini" type="danger" @click="handleDeleteMachine(scope.row)">
                                            {{ t('backup.delete') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                            <div class="pagination-wrapper" v-if="filteredBackupModelList.length > 0">
                                <el-pagination v-model:current-page="currentPage" :page-size="pageSize"
                                    :total="filteredBackupModelList.length" layout="total, prev, pager, next, jumper"
                                    size="small" />
                            </div>
                        </div>
                    </div>
                </el-tab-pane>
                <el-tab-pane :label="t('backup.backupMachine')" name="machine">
                     <el-alert
                      :title="t('backup.tip')"
                      type="info"
                      :closable="false"
                      show-icon
                      style="margin: 15px;padding: 16px;"
                    />
                    <div class="machine-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ t('backup.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="t('backup.group')">
                                    <el-option :label="t('backup.all')" value="全部"></el-option>
                                    <el-option v-for="group in deviceGroups" :key="group" :label="group"
                                        :value="group"></el-option>
                                </el-select>
                            </div>
                            <div class="device-list">
                                <div v-for="device in deviceList" :key="device.ip"
                                    :class="['device-item', { active: selectedDeviceIP === device.ip }]"
                                    @click="selectDevice(device)">
                                    <span class="device-ip">{{ device.ip }}</span>
                                </div>
                            </div>
                        </div>
                        <div class="right-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ t('backup.backupMachineList') }}</span>
                                <div style="display: flex; gap: 8px; align-items: center;">
                                    <el-button size="medium" @click="handleOpenBackupMachineDir">{{ t('backup.openLocalBackupMachine') }}</el-button>
                                    <el-input v-model="machineSearchText" :placeholder="t('backup.searchMachineName')" size="medium"
                                        style="width: 200px;" clearable @clear="handleMachineSearchClear">
                                        <template #prefix>
                                            <el-icon>
                                                <Search />
                                            </el-icon>
                                        </template>
                                    </el-input>
                                    <el-button type="primary" size="medium" @click="handleAddBackupMachine">
                                        <el-icon>
                                            <Plus />
                                        </el-icon> {{ t('backup.add') }}
                                    </el-button>
                                </div>
                            </div>
                            <el-table :data="paginatedBackupMachineList" v-loading="backupMachineLoading" stripe
                                max-height="800" style="width: 100%;">
                                <el-table-column type="index" :label="t('backup.index')" align="center" width="80" />
                                <el-table-column prop="name" :label="t('backup.machineName')" align="center" show-overflow-tooltip />
                                <el-table-column prop="size" :label="t('backup.size')" align="center" width="120" />
                                <el-table-column :label="t('backup.operation')" align="center">
                                    <template #default="scope">
                                        <el-button v-if="isBackupMachineDownloaded(scope.row.name)" size="mini" type="success"
                                            @click="handleImportBackupMachine(scope.row)">
                                            {{ t('backup.import') }}
                                        </el-button>
                                        <el-button v-else size="mini" type="primary"
                                            :loading="isDownloading(scope.row.name)"
                                            :disabled="isDownloading(scope.row.name)"
                                            @click="handleDownloadMachine(scope.row)">
                                            {{ isDownloading(scope.row.name) ? t('backup.downloading') : t('backup.download') }}
                                        </el-button>
                                        <el-button size="mini" type="danger" 
                                            :disabled="isDownloading(scope.row.name)"
                                            @click="handleDeleteBackupMachine(scope.row)">
                                            {{ t('backup.delete') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                            <div class="pagination-wrapper" v-if="backupMachineList.length > 0">
                                <el-pagination v-model:current-page="machineCurrentPage" :page-size="machinePageSize"
                                    :total="backupMachineList.length" layout="total, prev, pager, next, jumper"
                                    size="small" />
                            </div>
                        </div>
                    </div>
                </el-tab-pane>

                <!-- 批量导入 -->
                <el-tab-pane :label="$t('backup.batchImport')" name="batch-import">
                    <BatchImportContent 
                        ref="batchImportRef"
                        :devices="devices" 
                        :device-firmware-info="deviceFirmwareInfo"
                        :devices-status-cache="devicesStatusCache"
                    />
                </el-tab-pane>

                <!-- 使用说明 -->
                <el-tab-pane :label="$t('backup.usageGuide')" name="help">
                    <div v-if="$i18n.locale === 'zh-CN'" style="height: 100%; overflow-y: auto; box-sizing: border-box;">
                    <div style="margin: 20px; padding: 0;">
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                            <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 12px;">
                                📖 备份管理使用说明
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">

                                <!-- 备份机型 -->
                                <el-collapse-item name="backup-model" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📦 备份机型</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p>用于将设备上的云机配置（机型模板）导出保存到本地，或将本地备份的机型导入到其他设备。</p>
                                        <ul style="margin: 6px 0; padding-left: 20px;">
                                            <li><strong>导出</strong>：在左侧选择设备后，右侧列表会显示该设备上的备份机型。点击"导出"按钮，将机型文件下载到本地保存。</li>
                                            <li><strong>{{ $t('backup.importBtn') }}</strong>：若本地已存在同名机型文件，操作列将显示"导入"按钮。点击后选择目标设备，即可将该机型批量导入到所选设备。</li>
                                            <li><strong>新增</strong>：点击右上角"新增"按钮，选择设备上的云机并命名，系统将对该云机的机型进行备份。</li>
                                            <li><strong>{{ $t('common.delete') }}</strong>：点击"删除"按钮，可删除设备上保存的对应备份机型记录。</li>
                                            <li>点击"打开本地备份机型目录"可快速查看本地已保存的机型文件。</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <!-- 备份云机 -->
                                <el-collapse-item name="backup-machine" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">☁️ 备份云机</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p>用于将设备上的完整云机镜像文件导出到本地，或将本地已下载的云机镜像恢复到指定设备和坑位。</p>
                                        <ul style="margin: 6px 0; padding-left: 20px;">
                                            <li><strong>下载</strong>：点击"下载"按钮，将设备上的云机备份文件下载到本机保存。下载过程中按钮会显示"下载中"状态，请勿重复点击。</li>
                                            <li><strong>导入</strong>：本地已下载的云机备份会显示"导入"按钮。点击后选择目标设备、填写坑位号及新云机名称，确认后将进行恢复操作。</li>
                                            <li><strong>新增</strong>：点击右上角"新增"按钮，选择设备上的云机，系统将对该云机进行整机备份（文件较大，请耐心等待）。</li>
                                            <li><strong>删除</strong>：删除操作将同时删除设备端备份记录及本地已下载的文件，请谨慎操作。</li>
                                            <li>点击"打开本地备份云机目录"可查看本地已下载的云机镜像文件。</li>
                                        </ul>
                                        <div style="margin-top: 8px; padding: 6px 10px; background: #fff7e6; border: 1px solid #ffd591; border-radius: 4px; color: #d46b08;">
                                            ⚠️ 注意：导入备份云机前请确保镜像文件已完整下载到本地，且目标设备处于在线状态。
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 批量导入 -->
                                <el-collapse-item name="batch-import-help" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">🚀 批量导入</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p>支持将本地已有的云机备份文件批量分发并导入到多台设备，适合大规模初始化场景。</p>
                                        <ul style="margin: 6px 0; padding-left: 20px;">
                                            <li>在"批量导入"标签页中，选择本地备份文件，再勾选目标设备，点击"开始导入"即可批量执行。</li>
                                            <li>导入进度会实时显示，成功/失败结果会分别统计汇报。</li>
                                            <li>批量导入期间请保持设备在线，避免网络中断导致导入失败。</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                            </el-collapse>
                        </div>
                    </div>
                    </div>
<div v-else style="height: 100%; overflow-y: auto; box-sizing: border-box;">
                    <div style="margin: 20px; padding: 0;">
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                            <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 12px;">
                                📖 Backup Management Usage Guide
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">

                                <!-- 备份机型 -->
                                <el-collapse-item name="backup-model" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📦 Backup Model</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p>Export VM configurations (model templates) from devices to local storage, or import locally backed-up models to other devices.</p>
                                        <ul style="margin: 6px 0; padding-left: 20px;">
                                            <li><strong>Export</strong>: After selecting a device on the left, the right panel shows its backup models. Click "Export" to download the model file to local storage.</li>
                                            <li><strong>{{ $t('backup.importBtn') }}</strong>: If a local model file with the same name exists, an "Import" button will appear. Click it and select target devices to batch-import the model.</li>
                                            <li><strong>Add</strong>: Click the "Add" button in the upper right, select a VM on the device and name it. The system will back up that VM's model.</li>
                                            <li><strong>{{ $t('common.delete') }}</strong>: Click the "Delete" button to remove the corresponding backup model record from the device.</li>
                                            <li>Click "Open Local Backup Model Directory" to quickly browse locally saved model files.</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <!-- 备份云机 -->
                                <el-collapse-item name="backup-machine" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">☁️ Backup VM</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p>Export complete VM image files from devices to local storage, or restore locally downloaded VM images to specified devices and slots.</p>
                                        <ul style="margin: 6px 0; padding-left: 20px;">
                                            <li><strong>Download</strong>: Click "Download" to save the VM backup file locally. The button will show "Downloading" status during the process — do not click repeatedly.</li>
                                            <li><strong>Import</strong>: Locally downloaded VM backups will show an "Import" button. Click it, select the target device, fill in the slot number and new VM name, then confirm to perform the restore.</li>
                                            <li><strong>Add</strong>: Click the "Add" button in the upper right, select a VM on the device. The system will perform a full VM backup (files are large — please be patient).</li>
                                            <li><strong>Delete</strong>: The delete operation removes both the device-side backup record and locally downloaded files — please proceed with caution.</li>
                                            <li>Click "Open Local Backup VM Directory" to view locally downloaded VM image files.</li>
                                        </ul>
                                        <div style="margin-top: 8px; padding: 6px 10px; background: #fff7e6; border: 1px solid #ffd591; border-radius: 4px; color: #d46b08;">
                                            ⚠️ Note: Before importing a backup VM, ensure the image file has been fully downloaded locally and the target device is online.
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 批量导入 -->
                                <el-collapse-item name="batch-import-help" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">🚀 Batch Import</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p>Supports batch distribution and import of locally available VM backup files to multiple devices — ideal for large-scale initialization scenarios.</p>
                                        <ul style="margin: 6px 0; padding-left: 20px;">
                                            <li>In the "Batch Import" tab, select local backup files, check target devices, then click "Start Import" to execute in batch.</li>
                                            <li>Import progress is displayed in real-time, with success/failure results reported separately.</li>
                                            <li>Keep devices online during batch import to avoid import failures caused by network interruptions.</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                            </el-collapse>
                        </div>
                    </div>
                    </div>
                </el-tab-pane>
            </el-tabs>
        </div>

        <el-dialog v-model="addDialogVisible" :title="t('backup.addBackupModel')" width="40%" :close-on-click-modal="false">
            <el-form label-width="100px">
                <el-form-item :label="t('backup.selectMachine')">
                    <el-select v-model="selectedContainer" :placeholder="t('backup.pleaseSelectMachine')" size="medium" style="width: 100%;"
                        @change="handleContainerChange">
                        <el-option v-for="container in containerList" :key="container.id" :label="container.name"
                            :value="container.id" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="t('backup.modelName')" v-if="selectedContainer">
                    <el-input v-model="modelName" :placeholder="t('backup.enterModelName')" size="medium" style="width: 100%;" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="addDialogVisible = false">{{ t('backup.cancel') }}</el-button>
                <el-button type="primary" @click="handleConfirmAdd" :disabled="!selectedContainer || !modelName.trim()">
                    {{ t('backup.confirm') }}
                </el-button>
            </template>
        </el-dialog>

        <el-dialog v-model="importDialogVisible" :title="t('backup.batchImportBackupModel')" width="60%" :close-on-click-modal="false">
            <div style="margin-bottom: 15px; padding: 10px; background: #f0f9ff; border: 1px solid #91d5ff; border-radius: 4px;">
                <div style="display: flex; align-items: center; gap: 8px; color: #0958d9;">
                    <el-icon><InfoFilled /></el-icon>
                    <span style="font-size: 14px;">
                        在线设备: <span style="font-weight: bold;">{{ onlineImportMachineList.length }}</span> 台 / 总共: {{ importMachineList.length }} 台
                    </span>
                </div>
            </div>
            <el-table ref="importMachineTableRef" :data="onlineImportMachineList" v-loading="loading" @selection-change="handleSelectionChange"
                max-height="400">
                <el-table-column type="selection" width="55" />
                <el-table-column prop="ip" :label="t('backup.deviceIP')" align="center" />
                <el-table-column :label="t('backup.deviceStatus')" align="center" width="100">
                    <template #default="scope">
                        <el-tag type="success" size="small">{{ $t('common.online') }}</el-tag>
                    </template>
                </el-table-column>
            </el-table>
            <template #footer>
                <el-button @click="cancelImportDialog">{{ t('backup.cancel') }}</el-button>
                <el-button type="primary" @click="handleBatchImport" :disabled="!selectedImportMachines.length">
                    {{ t('backup.batchImport') }} ({{ selectedImportMachines.length }})
                </el-button>
            </template>
        </el-dialog>

        <el-dialog v-model="addBackupMachineVisible" :title="t('backup.addBackupMachine')" width="40%" :close-on-click-modal="false">
            <el-form label-width="100px">
                <el-form-item :label="t('backup.selectMachine')">
                    <el-select v-model="selectedBackupMachineContainer" :placeholder="t('backup.pleaseSelectMachine')" size="medium" style="width: 100%;"
                        @change="handleBackupMachineContainerChange">
                        <el-option v-for="container in backupMachineContainerList" :key="container.id" :label="container.name"
                            :value="container.id" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="addBackupMachineVisible = false">{{ t('backup.cancel') }}</el-button>
                <el-button type="primary" @click="handleConfirmAddBackupMachine" :disabled="!selectedBackupMachineContainer" :loading="backupMachineCreating">
                    {{ t('backup.confirm') }}
                </el-button>
            </template>
        </el-dialog>

        <el-dialog v-model="importBackupDialogVisible" :title="t('backup.importBackupMachine') + (selectedDeviceIP ? t('backup.ensureImageDownloaded') : '')" width="40%" :close-on-click-modal="false">
            <template #header="{ titleId, titleClass }">
                <div :id="titleId" :class="['el-dialog__title', titleClass]">
                    {{ t('backup.importBackupMachine') }}
                    <span v-if="selectedDeviceIP" type="danger" size="small" style="margin-left: 8px;color: red;">
                        {{ t('backup.ensureImageDownloaded') }}
                    </span>
                </div>
            </template>
            
            <!-- 设备统计信息 -->
            <div style="margin-bottom: 15px; padding: 10px; background: #f0f9ff; border: 1px solid #91d5ff; border-radius: 4px;">
                <div style="display: flex; align-items: center; gap: 8px; color: #0958d9;">
                    <el-icon><InfoFilled /></el-icon>
                    <span style="font-size: 14px;">
                        在线设备: <span style="font-weight: bold;">{{ onlineDevicesForBackupImport.length }}</span> 台 / 总共: {{ props.devices.length }} 台
                    </span>
                </div>
            </div>

            <el-form label-width="100px">
                <el-form-item :label="t('backup.selectDevice')">
                    <el-select v-model="importBackupDeviceIP" :placeholder="t('backup.pleaseSelectDevice')" size="medium" style="width: 100%;"
                        @change="handleImportBackupDeviceChange">
                        <el-option v-for="device in onlineDevicesForBackupImport" :key="device.ip" 
                            :label="device.ip + ' (' + $t('common.online') + ')'"
                            :value="device.ip">
                            <div style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ device.ip }}</span>
                                <el-tag type="success" size="small">{{ $t('common.online') }}</el-tag>
                            </div>
                        </el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="t('backup.slotNumber')" v-if="importBackupDeviceIP">
                    <el-input-number v-model="importBackupSlot" :min="1" :max="importBackupSlotList.length" size="medium" style="width: 100%;" />
                    <div style="font-size: 12px; color: #909399; margin-top: 4px;">
                        {{ t('backup.availableSlots') }}: {{ importBackupSlotList.join(', ') || t('backup.noAvailableSlots') }}
                    </div>
                </el-form-item>
                <el-form-item :label="t('backup.machineName')">
                    <el-input v-model="importBackupMachineName" :placeholder="t('backup.enterMachineName')" size="medium" style="width: 100%;" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="importBackupDialogVisible = false">{{ t('backup.cancel') }}</el-button>
                <el-button type="primary" @click="handleConfirmImportBackupMachine" :disabled="!importBackupDeviceIP || !importBackupMachineName.trim()" :loading="importBackupLoading">
                    {{ t('backup.confirm') }}
                </el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>

// 返回设备的 host:port，若 ip 已含端口则直接使用，否则追加默认 8000
const getDeviceAddr = (ip) => {
  if (!ip) return ip
  const lastColon = ip.lastIndexOf(':')
  if (lastColon === -1) return ip + ':8000'
  return /^\d+$/.test(ip.slice(lastColon + 1)) ? ip : ip + ':8000'
}


import { ref, onMounted, computed, watch, getCurrentInstance, nextTick } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Refresh, Search, Plus, InfoFilled } from '@element-plus/icons-vue';
import { ExportBackupModel, CheckBackupModelExists, GetAllBackupModels, ImportBackupModel, DeleteBackupModel, DownloadBackupMachine, CheckBackupMachineFileExists, CheckBackupMachineFilesExistBatch, ImportBackupMachine, DeleteLocalBackupMachine, OpenBackupModelDir, OpenBackupMachineDir } from '../../bindings/edgeclient/app';
import BatchImportContent from './BatchImport.vue';

// 国际化支持
const { proxy } = getCurrentInstance()

// 批量导入组件引用
const batchImportRef = ref(null)
// 批量导入备份机型表格引用
const importMachineTableRef = ref(null)
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

const activeTab = ref('model');
const activeHelpSections = ref(['backup-model', 'backup-machine', 'batch-import-help']);
const loading = ref(false);
const selectedDeviceIP = ref('')
const selectedDeviceVersion = ref('v3')
const backupModelList = ref([]);
const backupMachineList = ref([]);
const backupMachineLoading = ref(false);
const machineSearchText = ref('');
const machineCurrentPage = ref(1);
const machinePageSize = ref(10);
const downloadedBackupMachineSet = ref(new Set());
const downloadingSet = ref(new Set()); // 正在下载的云机集合
const localGroupFilter = ref('全部')
const modelSearchText = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const addDialogVisible = ref(false)
const containerList = ref([])
const selectedContainer = ref(null)
const modelName = ref('')
const localBackupModelSet = ref(new Set())
const importDialogVisible = ref(false)
const importMachineList = ref([])
const selectedImportMachines = ref([])
const MachinesName = ref('')

// 计算属性：只显示在线设备用于导入
const onlineImportMachineList = computed(() => {
    return importMachineList.value.filter(device => {
        const status = props.devicesStatusCache.get(device.id)
        return status === 'online'
    })
})

// 计算属性：只显示在线设备用于备份云机导入
const onlineDevicesForBackupImport = computed(() => {
    return props.devices.filter(device => {
        const status = props.devicesStatusCache.get(device.id)
        return status === 'online'
    })
})
const addBackupMachineVisible = ref(false)
const backupMachineContainerList = ref([])
const selectedBackupMachineContainer = ref(null)
const backupMachineCreating = ref(false)
const importBackupDialogVisible = ref(false)
const importBackupMachineRow = ref(null)
const importBackupDeviceIP = ref('')
const importBackupDeviceName = ref('')
const importBackupSlot = ref(1)
const importBackupMachineName = ref('')
const importBackupSlotList = ref([])
const importBackupLoading = ref(false)

const props = defineProps({
    devices: {
        type: Array,
        default: () => []
    },
    deviceFirmwareInfo: {
        type: Map,
        default: () => new Map()
    },
    devicesStatusCache: {
        type: Map,
        default: () => new Map()
    }
})

const getAuthHeaders = (deviceIP) => {
    const savedPassword = localStorage.getItem('devicePasswords')
    const passwords = JSON.parse(savedPassword || '{}');
    const password = passwords[deviceIP] || null
    if (password) {
        const auth = btoa(`admin:${password}`)
        return {
            'Authorization': `Basic ${auth}`
        }
    }
    return {}
}

// 带超时的 fetch，默认 5 秒
const fetchWithTimeout = (url, options = {}, timeout = 5000) => {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeout)
    return fetch(url, { ...options, signal: controller.signal })
        .finally(() => clearTimeout(timer))
}



const compareIPs = (ip1, ip2) => {
    const parts1 = ip1.split('.').map(Number)
    const parts2 = ip2.split('.').map(Number)
    for (let i = 0; i < 4; i++) {
        if (parts1[i] < parts2[i]) return -1
        if (parts1[i] > parts2[i]) return 1
    }
    return 0
}

const deviceGroups = computed(() => {
    const groups = new Set(['默认分组'])
    props.devices.forEach(device => {
        if (device.group) {
            groups.add(device.group)
        }
    })
    return Array.from(groups).sort()
})

const filteredDevicesByGroup = computed(() => {
    let devices = props.devices

    if (localGroupFilter.value !== '全部') {
        devices = devices.filter(device => device.group === localGroupFilter.value)
    }

    return devices.slice().sort((a, b) => compareIPs(a.ip, b.ip))
})

const deviceList = computed(() => {
    // 只显示在线的设备
    return filteredDevicesByGroup.value.filter(device => {
        const status = props.devicesStatusCache.get(device.id)
        return status === 'online'
    })
})

const filteredBackupModelList = computed(() => {
    if (!modelSearchText.value.trim()) {
        return backupModelList.value
    }
    const searchText = modelSearchText.value.trim().toLowerCase()
    return backupModelList.value.filter(item =>
        item.name && item.name.toLowerCase().includes(searchText)
    )
})

const paginatedBackupModelList = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return filteredBackupModelList.value.slice(start, end)
})

const handleModelSearchClear = () => {
    currentPage.value = 1
}

const filteredBackupMachineList = computed(() => {
    if (!machineSearchText.value.trim()) {
        return backupMachineList.value
    }
    const searchText = machineSearchText.value.trim().toLowerCase()
    return backupMachineList.value.filter(item =>
        item.name && item.name.toLowerCase().includes(searchText)
    )
})

const paginatedBackupMachineList = computed(() => {
    const start = (machineCurrentPage.value - 1) * machinePageSize.value
    const end = start + machinePageSize.value
    return filteredBackupMachineList.value.slice(start, end)
})

const handleMachineSearchClear = () => {
    machineCurrentPage.value = 1
}

const handleAddModel = async () => {
    if (!selectedDeviceIP.value) {
        ElMessage.warning('请先选择设备')
        return
    }
    addDialogVisible.value = true
    selectedContainer.value = null
    modelName.value = ''
    await fetchContainers()
}

const handleAddBackupMachine = async () => {
    if (!selectedDeviceIP.value) {
        ElMessage.warning('请先选择设备')
        return
    }
    addBackupMachineVisible.value = true
    selectedBackupMachineContainer.value = null
    await fetchBackupMachineContainers()
}

const fetchBackupMachineContainers = async () => {
    try {
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/android?name=&running=false`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code == 0) {
                backupMachineContainerList.value = (data.data.list || []).filter(item => item.androidType !== 'V2')
            } else {
                ElMessage.error(data.message || '获取云机列表失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        ElMessage.error('获取云机列表失败，请检查网络连接')
    }
}

const handleBackupMachineContainerChange = (value) => {
    selectedBackupMachineContainer.value = value
}

const handleConfirmAddBackupMachine = async () => {
    if (!selectedBackupMachineContainer.value) {
        ElMessage.warning('请选择云机')
        return
    }

    const container = backupMachineContainerList.value.find(c => c.id === selectedBackupMachineContainer.value)
    if (!container) {
        ElMessage.warning('未找到选择的云机')
        return
    }

    backupMachineCreating.value = true
    try {
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/android/export`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({
                  name: container.name
                })
            },
            30000  // 备份操作给 30 秒
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code == 0) {
                ElMessage.success('备份成功')
                addBackupMachineVisible.value = false
                await fetchBackupMachines()
            } else {
                ElMessage.warning(data.message || '备份失败')
            }
        } else {
            ElMessage.error('备份失败')
        }
    } catch (error) {
        ElMessage.error('备份失败，请检查网络连接')
    } finally {
        backupMachineCreating.value = false
    }
}

const fetchContainers = async () => {
    try {
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/android?name=&running=false`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code == 0) {
                containerList.value = (data.data.list || []).filter(item => item.androidType !== 'V2')
            } else {
                ElMessage.error(data.message || '获取云机列表失败')
            }
        } else {
            ElMessage.error('获取云机列表失败')
        }
    } catch (error) {
        ElMessage.error('获取云机列表失败，请检查网络连接')
    }
}

const handleContainerChange = (value) => {
    selectedContainer.value = value
}

const handleConfirmAdd = async () => {
    if (!selectedContainer.value || !modelName.value.trim()) {
        return
    }
    const container = containerList.value.find(c => c.id === selectedContainer.value)
    if (!container) {
        ElMessage.error('未找到选择的云机')
        return
    }

    try {
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/android/backup/model`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({
                    name: container.name,
                    suffix: modelName.value.trim()
                })
            }
        )
        const data = await response.json()
        if (data.code === 0) {
            ElMessage.success('新增备份机型成功')
            addDialogVisible.value = false
            await fetchBackupModels()
        } else {
            ElMessage.error(data.message || '新增备份机型失败')
        }
    } catch (error) {
        ElMessage.error('新增备份机型失败，请检查网络连接')
    }
}

const selectDevice = async (device) => {
    selectedDeviceIP.value = device.ip
    selectedDeviceVersion.value = device.version || 'v3'
    currentPage.value = 1
    modelSearchText.value = ''
    // 三个请求互相独立，全部并行
    const tasks = [fetchBackupModels(), fetchLocalBackupModels()]
    if (activeTab.value === 'machine') {
        tasks.push(fetchBackupMachines())
    }
    await Promise.all(tasks)
}

const fetchLocalBackupModels = async () => {
    try {
        const result = await GetAllBackupModels()
        if (result.success) {
            localBackupModelSet.value = new Set(result.list || [])
        }
    } catch (error) {
        console.error('获取本地备份机型列表失败:', error)
    }
}

const isModelLocalExists = (modelName) => {
    return localBackupModelSet.value.has(modelName)
}

const handleSelectionChange = (selection) => {
    selectedImportMachines.value = selection
}
// if (activeTab.value === 'vpc') {
//     await fetchContainerRule()
// }
// }


const fetchBackupModels = async () => {
    if (!selectedDeviceIP.value) {
        backupModelList.value = []
        return
    }
    // 当前选中设备已离线，跳过请求
    const selectedDevice = props.devices.find(d => d.ip === selectedDeviceIP.value)
    if (!selectedDevice || props.devicesStatusCache.get(selectedDevice.id) !== 'online') {
        backupModelList.value = []
        return
    }
    loading.value = true
    try {
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/android/backup/model`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code == 0) {
                backupModelList.value = data.data.list || []
            } else {
                ElMessage.error(data.message || '获取备份机型列表失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        ElMessage.error('获取备份机型列表失败，请检查网络连接')
    } finally {
        loading.value = false
    }
}

const fetchBackupMachines = async () => {
    if (!selectedDeviceIP.value) {
        backupMachineList.value = []
        return
    }
    // 当前选中设备已离线，跳过请求
    const selectedDevice = props.devices.find(d => d.ip === selectedDeviceIP.value)
    if (!selectedDevice || props.devicesStatusCache.get(selectedDevice.id) !== 'online') {
        backupMachineList.value = []
        return
    }
    backupMachineLoading.value = true
    try {
        const nameParam = machineSearchText.value.trim() ? `&name=${encodeURIComponent(machineSearchText.value.trim())}` : ''
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/backup?name=${nameParam}`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code == 0) {
                backupMachineList.value = data.data.list || []
                await checkDownloadedFiles()
            } else {
                backupMachineList.value = []
                ElMessage.error(data.message || '获取备份云机列表失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        ElMessage.error('获取备份云机列表失败，请检查网络连接')
    } finally {
        backupMachineLoading.value = false
    }
}

const checkDownloadedFiles = async () => {
    if (backupMachineList.value.length === 0) return
    try {
        const names = backupMachineList.value.map(m => m.name)
        const result = await CheckBackupMachineFilesExistBatch(names)
        if (result?.success) {
            const newSet = new Set()
            for (const [name, exists] of Object.entries(result.result || {})) {
                if (exists) newSet.add(name)
            }
            downloadedBackupMachineSet.value = newSet
        }
    } catch (err) {
        console.error('批量检查文件失败:', err)
    }
}

const isBackupMachineDownloaded = (machineName) => {
    return downloadedBackupMachineSet.value.has(machineName)
}

// 检查是否正在下载
const isDownloading = (machineName) => {
    return downloadingSet.value.has(machineName)
}

const handleOpenBackupModelDir = async () => {
    try {
        await OpenBackupModelDir()
    } catch (error) {
        ElMessage.error('打开目录失败')
    }
}

const handleOpenBackupMachineDir = async () => {
    try {
        await OpenBackupMachineDir()
    } catch (error) {
        ElMessage.error('打开目录失败')
    }
}

const handleTabChange = async (tab) => {
    // if (selectedDevice.value) {
        loading.value = true;
        try {
            if (tab === 'model') {
                await fetchBackupModels();
                backupMachineList.value = [];
            } else {
                await fetchBackupMachines();
                backupModelList.value = [];
            }
        } catch (error) {
            ElMessage.error('获取备份数据失败');
        } finally {
            loading.value = false;
        }
    // }
};


const handleRestore = (row) => {
    console.log('恢复云机:', row);
};

const handleExportMachine = async (row) => {
    if (!selectedDeviceIP.value) {
        ElMessage.warning('请先选择设备')
        return
    }

    try {
        const result = await ExportBackupModel(selectedDeviceIP.value, row.name)
        if (result.success) {
            ElMessage.success(`导出成功，保存在: ${result.filePath}`)
            await fetchLocalBackupModels()
        } else {
            ElMessage.error(result.message || '导出失败')
        }
    } catch (error) {
        ElMessage.error(`导出失败: ${error.message}`)
    }
};

const handleImportMachine = async (row) => {
    if (!selectedDeviceIP.value) {
        ElMessage.warning('请先选择设备')
        return
    }

    importDialogVisible.value = true
    selectedImportMachines.value = []
    MachinesName.value = row.name

    // loading.value = true

    importMachineList.value = props.devices || []

    // 弹窗打开后清除表格视觉选中状态
    nextTick(() => {
        importMachineTableRef.value?.clearSelection()
    })
};

const cancelImportDialog = () => {
    importDialogVisible.value = false
    selectedImportMachines.value = []
    importMachineTableRef.value?.clearSelection()
}

const handleBatchImport = async () => {
    if (!selectedImportMachines.value.length) {
        ElMessage.warning('请选择要导入的设备')
        return
    }

    loading.value = true
    const results = await Promise.all(
        selectedImportMachines.value.map(machine =>
            ImportBackupModel(machine.ip, MachinesName.value).catch(() => ({ success: false }))
        )
    )
    loading.value = false
    importDialogVisible.value = false
    selectedImportMachines.value = []
    importMachineList.value = []
    importMachineTableRef.value?.clearSelection()

    const successCount = results.filter(r => r?.success).length
    const failCount = results.length - successCount

    if (failCount === 0) {
        ElMessage.success(`成功导入 ${successCount} 个云机`)
        await fetchLocalBackupModels()
    } else {
        ElMessage.warning(`成功 ${successCount} 个，失败 ${failCount} 个`)
    }
};


const handleDeleteBackupMachine = async (row) => {
    try {
        await ElMessageBox.confirm(`确定要删除备份云机 "${row.name}" 吗?`, '确认删除', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        });

        backupMachineLoading.value = true;
        const response = await fetchWithTimeout(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/backup?name=${encodeURIComponent(row.name)}`,
            { 
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        );

        if (response.ok) {
            const data = await response.json();
            if (data.code == 0) {
                ElMessage.success('删除成功');
                downloadedBackupMachineSet.value.delete(row.name);
                await DeleteLocalBackupMachine(row.name);
                await fetchBackupMachines();
            } else {
                ElMessage.warning(data.message || '删除失败');
            }
        } else {
            ElMessage.error('删除失败');
        }
    } catch (error) {
        console.error('删除备份云机失败:', error);
        if (error === 'cancel' || error?.message === 'cancel') {
            ElMessage.info('已取消删除');
        } else {
            ElMessage.error('删除失败: ' + (error?.message || '未知错误'));
        }
    } finally {
        backupMachineLoading.value = false;
    }
};

const handleDownloadMachine = async (row) => {
    // 防止重复下载
    if (downloadingSet.value.has(row.name)) {
        ElMessage.warning('该云机正在下载中，请勿重复操作');
        return;
    }

    try {
        // 标记为下载中
        downloadingSet.value.add(row.name);
        
        ElMessage.info('开始下载备份云机...');
        backupMachineLoading.value = true;

        const result = await DownloadBackupMachine(selectedDeviceIP.value, row.name);

        if (result.success) {
            ElMessage.success('下载成功');
            downloadedBackupMachineSet.value.add(row.name);
        } else {
            ElMessage.warning(result.message || '下载失败');
        }
        await fetchBackupMachines();
    } catch (error) {
        console.error('下载备份云机失败:', error);
        ElMessage.error('下载失败: ' + (error?.message || '未知错误'));
    } finally {
        // 移除下载中标记
        downloadingSet.value.delete(row.name);
        backupMachineLoading.value = false;
    }
};

const handleImportBackupMachine = (row) => {
    importBackupMachineRow.value = row
    importBackupDeviceIP.value = ''
    importBackupDeviceName.value = row.name
    importBackupSlot.value = 1
    importBackupMachineName.value = ''
    importBackupSlotList.value = []
    importBackupDialogVisible.value = true
};

const handleImportBackupDeviceChange = async (deviceIP) => {
    const device = props.devices.find(d => d.ip === deviceIP)
    console.log('handleImportBackupDeviceChange', device)
    if (device) {
        // importBackupDeviceName.value = device.name || device.ip
        const isP1V3 = device.version == 'v3' && device.name?.toLowerCase()?.includes('p1')
        const maxSlot = isP1V3 ? 24 : 12
        importBackupSlotList.value = Array.from({ length: maxSlot }, (_, i) => i + 1)
        importBackupSlot.value = 1
    }
};

const handleConfirmImportBackupMachine = async () => {
    if (!importBackupDeviceIP.value || !importBackupMachineName.value.trim()) {
        ElMessage.warning('请填写完整信息')
        return
    }

    const device = props.devices.find(d => d.ip === importBackupDeviceIP.value)
    const isP1V3 = device?.version == 'v3' && device?.name?.toLowerCase()?.includes('p1')
    const maxSlot = isP1V3 ? 24 : 12
    if (importBackupSlot.value < 1 || importBackupSlot.value > maxSlot) {
        ElMessage.warning(`坑位号必须在 1-${maxSlot} 之间`)
        return
    }

    try {
        importBackupLoading.value = true

        const result = await ImportBackupMachine(
            importBackupDeviceIP.value,
            importBackupDeviceName.value,
            importBackupMachineName.value.trim(),
            importBackupSlot.value
        )

        if (result.success) {
            ElMessage.success('导入成功')
        } else {
            ElMessage.warning(result.message || '导入失败')
        }
    } catch (error) {
        console.error('导入备份云机失败:', error)
        ElMessage.error('导入失败: ' + (error?.message || '未知错误'))
    } finally {
        importBackupLoading.value = false
        importBackupDialogVisible.value = false
    }
};

const handleDeleteMachine = async (row) => {
    try {
        await ElMessageBox.confirm(`确定要删除备份云机 "${row.name}" 吗?`, '确认删除', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        });

        loading.value = true;
        const result = await DeleteBackupModel(selectedDeviceIP.value, row.name);

        if (result.success) {
            ElMessage.success('删除成功');
        } else {
            ElMessage.warning(result.message || '删除失败');
        }

        await Promise.all([fetchBackupModels(), fetchLocalBackupModels()]);
    } catch (error) {
        console.error('删除备份机型失败:', error);
        if (error === 'cancel' || error?.message === 'cancel') {
            ElMessage.info('已取消删除');
        } else {
            ElMessage.error('删除失败: ' + (error?.message || '未知错误'));
        }
    } finally {
        loading.value = false;
    }
};

const fetchBackups = async () => {
    // 若当前选中设备已离线，清空选中，重新选第一台在线设备
    if (selectedDeviceIP.value) {
        const selectedDevice = props.devices.find(d => d.ip === selectedDeviceIP.value)
        if (!selectedDevice || props.devicesStatusCache.get(selectedDevice.id) !== 'online') {
            selectedDeviceIP.value = ''
            backupModelList.value = []
            backupMachineList.value = []
        }
    }
    // 没有选中设备时，自动选中第一台在线设备
    if (!selectedDeviceIP.value && deviceList.value.length > 0) {
        await selectDevice(deviceList.value[0])
        return
    }
    // 三个请求互相独立，全部并行
    const tasks = [fetchBackupModels(), fetchLocalBackupModels()]
    if (activeTab.value === 'machine') {
        tasks.push(fetchBackupMachines())
    }
    await Promise.all(tasks)

    // 同时刷新批量导入列表
    if (batchImportRef.value) {
        batchImportRef.value.refreshBackupFiles()
    }
};

// 监听在线设备列表变化
watch(deviceList, (newList) => {
    // 若当前选中设备不在在线列表中，清空选中状态
    if (selectedDeviceIP.value && !newList.find(d => d.ip === selectedDeviceIP.value)) {
        selectedDeviceIP.value = ''
        backupModelList.value = []
        backupMachineList.value = []
    }
    // 若没有选中设备，自动选中第一台在线设备
    if (!selectedDeviceIP.value && newList.length > 0) {
        selectDevice(newList[0])
    }
})

defineExpose({
    fetchBackups
})

</script>

<style scoped>
.backup-management-container {
    height: 100%;
    box-sizing: border-box;
}

.backup-content {
    height: 100%;
    background: white;
    border-radius: 4px;
}

:deep(.el-table .cell) {
    display: block;
}

.backup-tabs {
    height: 100%;
}

:deep(.el-tabs__content) {
    height: calc(100% - 55px);
}

:deep(.el-tab-pane) {
    height: 100%;
    overflow-y: auto;
}

.tab-content {
    display: flex;
    height: 100%;
    gap: 16px;
    /* padding: 16px; */
    box-sizing: border-box;
}

.left-panel {
    width: 180px;
    border: 1px solid #ebeef5;
    border-radius: 4px;
    display: flex;
    flex-direction: column;
    background: #fafafa;
    flex-shrink: 0;
}

.panel-header {
    padding: 12px 16px;
    border-bottom: 1px solid #ebeef5;
    font-weight: 600;
    color: #303133;
    background: #f5f7fa;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.device-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
}

.device-item {
    padding: 14px 12px;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
    transition: all 0.2s;
}

.device-item:hover {
    background: #ecf5ff;
}

.device-item.active {
    background: #409eff;
    color: white;
}

.device-ip {
    font-size: 14px;
    color: #606266;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.device-item.active .device-ip {
    color: white;
}

.right-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    border: 1px solid #ebeef5;
    border-radius: 4px;
    overflow: hidden;
}

.right-panel .panel-header {
    border-bottom: 1px solid #ebeef5;
}

.pagination-wrapper {
    padding: 10px 16px;
    border-top: 1px solid #ebeef5;
    display: flex;
    justify-content: flex-end;
}

.machine-content {
    display: flex;
    /* justify-content: center; */
    /* align-items: center; */
    gap: 16px;
    height: 100%;
    /* padding: 40px; */
}
</style>
