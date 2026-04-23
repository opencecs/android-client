<template>
    <div class="network-management-container">
        <div class="network-content">
            <el-tabs v-model="activeTab" class="network-tabs" @tab-change="handleTabChange">
                <el-tab-pane :label="$t('network.nodeManagement')" name="group">
                    <!-- 提示信息 -->
                    <el-alert
                      :title="$t('network.macvlanHint')"
                      type="warning"
                      :closable="false"
                      show-icon
                      style="margin: 15px;padding: 16px;"
                    />
                    
                    <div class="tab-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="$t('common.selectGroupPlaceholder')">
                                    <el-option :label="$t('common.all')" value="全部"></el-option>
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
                            <div class="panel-header" style="flex-direction: column; align-items: stretch;">
                                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
                                    <span>{{ $t('network.nodeManagementList') }}</span>
                                    <div style="display: flex; align-items: center;">
                                        <el-select v-model="selectedGroupId" :placeholder="$t('common.selectGroupPlaceholder')"
                                            style="width: 200px; margin-right: 10px;">
                                            <template #prefix>
                                                <span>{{ $t('common.group') }}</span>
                                            </template>
                                            <el-option v-for="group in groupList" :key="group.id" :label="group.alias"
                                                :value="group.id">
                                            </el-option>
                                        </el-select>
                                        <div v-if="selectedGroupId" style="margin-right: 10px;">
                                            <el-button type="primary" size="small"
                                                @click="handleEditGroupAlias(selectedGroupId)">
                                                <el-icon>
                                                    <Edit />
                                                </el-icon>{{ $t('network.editGroupAlias') }}
                                            </el-button>
                                            <el-button type="primary" size="small"
                                                @click="handleUpdateGroup(selectedGroupId)">
                                                <el-icon>
                                                    <RefreshRight />
                                                </el-icon>{{ $t('network.updateGroup') }}
                                            </el-button>
                                            <el-button type="danger" size="small"
                                                @click="handleDeleteGroup(selectedGroupId)">
                                                <el-icon>
                                                    <Delete />
                                                </el-icon>{{ $t('network.deleteGroup') }}
                                            </el-button>
                                        </div>
                                        <el-input v-model="searchName" style="width: 200px;margin-right: 10px;"
                                            :placeholder="$t('network.searchGroup')" clearable @clear="handleSearch"
                                            @keyup.enter="handleSearch">
                                            <template #prefix>
                                                <el-icon>
                                                    <Search />
                                                </el-icon>
                                            </template>
                                        </el-input>
                                        <el-button type="primary" @click="handleAddGroup"
                                            v-if="selectedDeviceIP">{{ $t('network.addGroup') }}</el-button>
                                    </div>
                                </div>
                                <div style="display: flex; align-items: center; gap: 10px;">
                                    <span style="color: #606266; font-size: 14px;">{{ $t('network.selectedNodes', { count: selectedVpcNodes.length }) }}</span>
                                    <el-button type="danger" @click="batchDeleteVpcNodes" size="small">
                                        {{ $t('network.batchDeleteNodes') }}
                                    </el-button>
                                    <el-button type="warning" @click="batchTestSpeed" size="small">
                                        {{ $t('network.batchSpeedTest') }}
                                    </el-button>
                                </div>
                            </div>
                            <el-table :data="filteredVpcList" style="width: 100%" v-loading="loading" stripe @selection-change="handleVpcNodeSelectionChange">
                                <el-table-column type="selection" width="40" align="center" />
                                <el-table-column prop="protocol" :label="$t('network.type')" align="center" width="120" />
                                <el-table-column prop="remarks" :label="$t('network.alias')" align="center" show-overflow-tooltip
                                    min-width="180" />
                                <el-table-column prop="remarks" :label="$t('network.address')" align="center" show-overflow-tooltip
                                    min-width="140">
                                    <template #default="scope">
                                        <span>{{ JSON.parse(scope.row.profile).server }}</span>
                                    </template>
                                </el-table-column>
                                <el-table-column prop="remarks" :label="$t('network.port')" align="center" width="100">
                                    <template #default="scope">
                                        <span>{{ JSON.parse(scope.row.profile).serverPort }}</span>
                                    </template>
                                </el-table-column>
                                <el-table-column prop="remarks" :label="$t('network.protocol')" align="center" width="120">
                                    <template #default="scope">
                                        <span>{{ JSON.parse(scope.row.profile).network }}</span>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('network.latency')" align="center" width="120">
                                    <template #default="scope">
                                        <span :style="scope.row.latency == '-1ms' ? 'color: red' : ''">{{
                                            scope.row.latency ? scope.row.latency : '-' }}</span>
                                    </template>
                                </el-table-column>
                                <!-- <el-table-column prop="groupAlias" label="订阅分组" align="center" width="120" /> -->

                                <!-- <el-table-column prop="protocol" label="协议" align="center" width="100" /> -->
                                <!-- <el-table-column prop="source" label="来源" align="center" width="80">
                                    <template #default="{ row }">
                                        <el-tag v-if="row.source === 1" type="success">订阅</el-tag>
                                        <el-tag v-else-if="row.source === 2" type="primary">单地址</el-tag>
                                        <el-tag v-else>{{ row.source }}</el-tag>
                                    </template>
                                </el-table-column> -->
                                <el-table-column :label="$t('common.operation')" width="300" align="center" fixed="right">
                                    <template #default="scope">
                                        <!-- <el-button type="warning" size="small" @click="openEditNodeDialog(scope.row)">{{ $t('network.editNode') }}</el-button> -->
                                        <el-button type="danger" size="small" @click="deleteContainerRule(scope.row)">{{ $t('network.deleteGroupNode') }}</el-button>
                                        <el-button type="primary" size="small" @click="testSpeed(scope.row)">{{ $t('network.speedTest') }}</el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                        </div>
                    </div>
                </el-tab-pane>
                <el-tab-pane :label="$t('network.nodeAllocation')" name="vpc">
                    <!-- 提示信息 -->
                    <el-alert
                      :title="$t('network.macvlanHint')"
                      type="warning"
                      :closable="false"
                      show-icon
                      style="margin: 15px;padding: 16px;"
                    />
                    <div class="vpc-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="$t('common.selectGroupPlaceholder')">
                                    <el-option :label="$t('common.all')" value="全部"></el-option>
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
                            <div class="panel-header">
                                <span>{{ $t('network.vpcNodes') }}</span>
                                <div>
                                    <el-button type="warning" v-if="selectedDeviceIP && selectedContainerRules.length > 0" @click="batchClearContainerVpc"
                                        style="margin-top: 10px; margin-right: 10px;">{{ $t('network.batchClearVpc') }}</el-button>
                                    <el-button type="primary" v-if="selectedDeviceIP" @click="setContainerVpc"
                                        style="margin-top: 10px;">{{ $t('network.assignVpc') }}</el-button>
                                </div>
                            </div>
                            <el-table :data="containerRuleList" v-loading="containerRuleLoading" stripe @selection-change="handleContainerRuleSelectionChange">
                                <el-table-column type="selection" width="55" align="center" />
                                <!-- <el-table-column prop="id" label="ID" align="center" width="80" /> -->
                                <el-table-column prop="containerName" :label="$t('network.machineName')" align="center" min-width="150">
                                    <template #default="{ row }">
                                        {{ extractNodeNumber(row.containerName) }}
                                    </template>
                                </el-table-column>
                                <!-- <el-table-column prop="targetPort" label="目标端口" align="center" width="100" /> -->
                                <el-table-column prop="containerIP" :label="$t('network.machineIP')" align="center" min-width="180" />
                                <!-- <el-table-column prop="sourceIp" label="源IP" align="center" min-width="150" /> -->
                                <el-table-column prop="groupName" :label="$t('network.sourceGroup')" align="center" min-width="180" />
                                <el-table-column prop="vpcRemarks" :label="$t('network.nodeName')" align="center" min-width="220" />
                                <el-table-column :label="$t('common.operation')" min-width="260" align="center" fixed="right">
                                    <template #default="scope">
                                        <!-- <el-button type="danger" @click="deleteContainerRule(scope.row)">{{ $t('network.deleteGroupNode') }}</el-button> -->
                                        <el-button type="warning" @click="clearContainerVpc(scope.row)"
                                            style="margin-top: 10px;">{{ $t('network.clearVpc') }}</el-button>
                                        <el-button type="primary" @click="enableContainerDnsWhitelist(scope.row)"
                                            style="margin-top: 10px;">{{ $t(scope.row.WhiteListDns != null && scope.row.WhiteListDns.length > 0 ? 'network.closeDns' : 'network.openDns')}}</el-button>
                                        <!-- <el-button type="info" @click="disableContainerDnsWhitelist(scope.row)"
                                            style="margin-top: 10px;">{{ $t('network.closeDnsWhitelist') }}</el-button> -->
                                    </template>
                                </el-table-column>
                                <!-- <el-table-column prop="status" label="状态" align="center" width="100">
                                    <template #default="{ row }">
                                        <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                                            {{ row.status === 1 ? '启用' : '禁用' }}
                                        </el-tag>
                                    </template>
                                </el-table-column> -->
                            </el-table>
                        </div>
                    </div>
                </el-tab-pane>

                 <!-- 域名过滤标签页 -->
                <el-tab-pane :label="$t('network.domainFilter')" name="domain-filter">
                    <div class="vpc-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="$t('common.selectGroupPlaceholder')">
                                    <el-option :label="$t('common.all')" value="全部"></el-option>
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
                        <div class="right-panel" style="padding: 16px; overflow-y: auto; display: flex; flex-direction: column; gap: 12px;">
                            <!-- 无设备提示 -->
                            <el-empty v-if="!selectedDeviceIP" :description="$t('network.selectDeviceFirst')" :image-size="60" style="margin: auto;" />
                            <template v-else>
                                <!-- 顶部操作按钮区 -->
                                <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
                                    <el-button type="primary" @click="openContainerDomainFilterDialog">{{ $t('network.setContainerDomainFilter') }}</el-button>
                                    <el-button type="warning" @click="openGlobalDomainFilterDialog">{{ $t('network.setGlobalDomainFilter') }}</el-button>
                                    <el-button @click="fetchGlobalDomainFilterList" :loading="globalDomainFilterLoading">{{ $t('network.queryGlobalDomainFilter') }}</el-button>
                                    <el-button
                                        v-if="domainQueryMode === 'container' && domainFilterList.length > 0"
                                        type="danger"
                                        plain
                                        @click="clearContainerDomainFilter"
                                    >{{ $t('network.clearFilter') }}</el-button>
                                    <el-button
                                        v-if="domainQueryMode === 'global' && globalDomainFilterList.length > 0"
                                        type="danger"
                                        plain
                                        @click="clearGlobalDomainFilter"
                                    >{{ $t('network.clearFilter') }}</el-button>
                                </div>

                                <!-- 分隔线 -->
                                <el-divider style="margin: 0;" />

                                <!-- 查询行：容器下拉 + 当前查询来源标签 -->
                                <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
                                    <span style="font-weight: 600; color: #303133; font-size: 14px; white-space: nowrap; flex-shrink: 0;">{{ $t('network.queryContainerDomainFilter') }}</span>
                                    <el-select
                                        v-model="domainFilterContainerID"
                                        :placeholder="$t('network.selectContainer')"
                                        style="flex: 1; min-width: 220px; max-width: 380px;"
                                        filterable
                                        clearable
                                        :loading="domainFilterVpcLoading"
                                        @change="onContainerDomainFilterChange"
                                        @clear="onContainerDomainFilterClear"
                                    >
                                        <el-option
                                            v-for="ct in domainFilterVpcContainers"
                                            :key="ct.containerName"
                                            :label="ct.containerName"
                                            :value="ct.containerName"
                                        >
                                            <span>{{ ct.containerName }}</span>
                                        </el-option>
                                        <template #empty>
                                            <div style="padding: 8px 12px; color: #909399; font-size: 13px; text-align: center;">
                                                {{ $t(domainFilterVpcLoading ? 'common.loading' : 'network.noVpcContainers') }}
                                            </div>
                                        </template>
                                    </el-select>
                                    <!-- 当前查询来源提示 -->
                                    <el-tag v-if="domainQueryMode === 'global'" type="warning" size="small">{{ $t('network.globalRules') }}</el-tag>
                                    <el-tag v-else-if="domainQueryMode === 'container'" type="primary" size="small">{{ domainFilterContainerID }}</el-tag>
                                </div>

                                <!-- 统一数据列表 -->
                                <el-table
                                    :data="unifiedDomainFilterList"
                                    v-loading="domainFilterLoading || globalDomainFilterLoading"
                                    stripe
                                    style="width: 100%;"
                                    :empty-text="$t(domainQueryMode ? 'network.noDomainRules' : 'network.selectContainerOrQuery')"
                                >
                                    <el-table-column type="index" :label="$t('network.index')" width="80" align="center" />
                                    <el-table-column :label="$t('network.filterDomain')" align="center">
                                        <template #default="{ row }">
                                            <span style="font-family: monospace;">{{ domainRuleValue(row.domain) }}</span>
                                        </template>
                                    </el-table-column>
                                </el-table>
                            </template>
                        </div>
                    </div>

                    <!-- ① {{ $t('network.setContainerDomainFilter') }}弹窗 -->
                    <el-dialog
                        v-model="showContainerDomainDialog"
                        :title="$t('network.setContainerDomainFilter')"
                        width="600px"
                        :close-on-click-modal="false"
                        @closed="resetContainerDomainDialog"
                    >
                        <el-form label-width="90px" style="padding-right: 10px;">
                            <!-- 选择容器 -->
                            <el-form-item :label="$t('network.selectContainer')" required>
                                <el-select
                                    v-model="containerDomainForm.containerID"
                                    :placeholder="$t('network.selectVpcContainer')"
                                    style="width: 100%;"
                                    filterable
                                    :loading="domainFilterVpcLoading"
                                >
                                    <el-option
                                        v-for="ct in domainFilterVpcContainers"
                                        :key="ct.containerName"
                                        :label="ct.containerName"
                                        :value="ct.containerName"
                                    >
                                        <span>{{ ct.containerName }}</span>
                                        <span style="color: #909399; font-size: 12px; margin-left: 8px;">{{ ct.containerIP }}</span>
                                    </el-option>
                                    <template #empty>
                                        <div style="padding: 8px 12px; color: #909399; font-size: 13px; text-align: center;">
                                            {{ $t(domainFilterVpcLoading ? 'common.loading' : 'network.noVpcContainers') }}
                                        </div>
                                    </template>
                                </el-select>
                            </el-form-item>

                            <!-- 域名规则列表 -->
                            <el-form-item :label="$t('network.domainRules')" required>
                                <div style="width: 100%;">
                                    <div
                                        v-for="(rule, idx) in containerDomainForm.rules"
                                        :key="idx"
                                        style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;"
                                    >
                                        <el-select v-model="rule.prefix" style="width: 180px;" size="default">
                                            <el-option label="domain:（子域名、IP）" value="domain:" />
                                            <el-option label="full:（完整域名、IP）" value="full:" />
                                            <el-option label="keyword:（关键字）" value="keyword:" />
                                            <!-- <el-option label="regexp:（正则）" value="regexp:" /> -->
                                        </el-select>
                                        <el-input
                                            v-model="rule.value"
                                            :placeholder="domainRulePlaceholder(rule.prefix)"
                                            style="flex: 1;"
                                            size="default"
                                        />
                                        <el-button
                                            :icon="Delete"
                                            circle
                                            size="small"
                                            type="danger"
                                            plain
                                            :disabled="containerDomainForm.rules.length === 1"
                                            @click="containerDomainForm.rules.splice(idx, 1)"
                                        />
                                    </div>
                                    <el-button size="small" @click="containerDomainForm.rules.push({ prefix: 'domain:', value: '' })">+ {{ $t('network.addRule') }}</el-button>
                                </div>
                            </el-form-item>

                            <!-- 规则说明 -->
                            <el-form-item>
                                <div style="background: #f5f7fa; border-radius: 4px; padding: 10px 12px; font-size: 12px; color: #606266; line-height: 1.9; width: 100%;">
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:</code> {{ $t('network.domainRuleDesc') }}</div>
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:</code> {{ $t('network.fullRuleDesc') }}</div>
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:</code> {{ $t('network.keywordRuleDesc') }}</div>
                                    <!-- <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">regexp:</code> 使用正则表达式匹配</div> -->
                                </div>
                            </el-form-item>
                        </el-form>
                        <template #footer>
                            <el-button @click="showContainerDomainDialog = false">{{ $t('common.cancel') }}</el-button>
                            <el-button
                                type="primary"
                                :loading="containerDomainSubmitLoading"
                                @click="submitContainerDomainFilter"
                            >{{ $t('common.confirm') }}</el-button>
                        </template>
                    </el-dialog>

                    <!-- ② {{ $t('network.setGlobalDomainFilter') }}弹窗 -->
                    <el-dialog
                        v-model="showGlobalDomainDialog"
                        :title="$t('network.setGlobalDomainFilter')"
                        width="600px"
                        :close-on-click-modal="false"
                        @closed="resetGlobalDomainDialog"
                    >
                        <el-form label-width="90px" style="padding-right: 10px;">
                            <!-- 域名规则列表 -->
                            <el-form-item :label="$t('network.domainRules')" required>
                                <div style="width: 100%;">
                                    <div
                                        v-for="(rule, idx) in globalDomainForm.rules"
                                        :key="idx"
                                        style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;"
                                    >
                                        <el-select v-model="rule.prefix" style="width: 180px;" size="default">
                                            <el-option label="domain:（子域名、IP）" value="domain:" />
                                            <el-option label="full:（完整域名、IP）" value="full:" />
                                            <el-option label="keyword:（关键字）" value="keyword:" />
                                            <!-- <el-option label="regexp:（正则）" value="regexp:" /> -->
                                        </el-select>
                                        <el-input
                                            v-model="rule.value"
                                            :placeholder="domainRulePlaceholder(rule.prefix)"
                                            style="flex: 1;"
                                            size="default"
                                        />
                                        <el-button
                                            :icon="Delete"
                                            circle
                                            size="small"
                                            type="danger"
                                            plain
                                            :disabled="globalDomainForm.rules.length === 1"
                                            @click="globalDomainForm.rules.splice(idx, 1)"
                                        />
                                    </div>
                                    <el-button size="small" @click="globalDomainForm.rules.push({ prefix: 'domain:', value: '' })">+ 添加规则</el-button>
                                </div>
                            </el-form-item>

                            <!-- 规则说明 -->
                            <el-form-item>
                                <div style="background: #f5f7fa; border-radius: 4px; padding: 10px 12px; font-size: 12px; color: #606266; line-height: 1.9; width: 100%;">
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:</code> {{ $t('network.domainRuleDesc') }}</div>
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:</code> {{ $t('network.fullRuleDesc') }}</div>
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:</code> {{ $t('network.keywordRuleDesc') }}</div>
                                    <!-- <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">regexp:</code> 使用正则表达式匹配</div> -->
                                </div>
                            </el-form-item>
                        </el-form>
                        <template #footer>
                            <el-button @click="showGlobalDomainDialog = false">{{ $t('common.cancel') }}</el-button>
                            <el-button
                                type="primary"
                                :loading="globalDomainSubmitLoading"
                                @click="submitGlobalDomainFilter"
                            >{{ $t('common.confirm') }}</el-button>
                        </template>
                    </el-dialog>
                </el-tab-pane>

                <!-- 域名直连标签页 -->
                <el-tab-pane :label="$t('network.domainDirect')" name="domain-direct">
                    <div class="vpc-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="$t('common.selectGroupPlaceholder')">
                                    <el-option :label="$t('common.all')" value="全部"></el-option>
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
                        <div class="right-panel" style="padding: 16px; overflow-y: auto; display: flex; flex-direction: column; gap: 12px;">
                            <!-- 无设备提示 -->
                            <el-empty v-if="!selectedDeviceIP" :description="$t('network.selectDeviceFirst')" :image-size="60" style="margin: auto;" />
                            <template v-else>
                                <!-- 顶部操作按钮区 -->
                                <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
                                    <el-button type="primary" @click="openDomainDirectDialog">{{ $t('network.setDomainDirect') }}</el-button>
                                    <el-button
                                        v-if="domainDirectList.length > 0"
                                        type="danger"
                                        plain
                                        @click="clearDomainDirect"
                                    >{{ $t('network.clearDirect') }}</el-button>
                                </div>

                                <!-- 分隔线 -->
                                <el-divider style="margin: 0;" />

                                <!-- 查询行：容器下拉 -->
                                <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
                                    <span style="font-weight: 600; color: #303133; font-size: 14px; white-space: nowrap; flex-shrink: 0;">{{ $t('network.queryContainerDomainDirect') }}</span>
                                    <el-select
                                        v-model="domainDirectContainerID"
                                        :placeholder="$t('network.selectContainer')"
                                        style="flex: 1; min-width: 220px; max-width: 380px;"
                                        filterable
                                        clearable
                                        :loading="domainDirectVpcLoading"
                                        @change="onDomainDirectContainerChange"
                                        @clear="onDomainDirectContainerClear"
                                    >
                                        <el-option
                                            v-for="ct in domainDirectVpcContainers"
                                            :key="ct.containerName"
                                            :label="ct.containerName"
                                            :value="ct.containerName"
                                        >
                                            <span>{{ ct.containerName }}</span>
                                        </el-option>
                                        <template #empty>
                                            <div style="padding: 8px 12px; color: #909399; font-size: 13px; text-align: center;">
                                                {{ $t(domainDirectVpcLoading ? 'common.loading' : 'network.noVpcContainers') }}
                                            </div>
                                        </template>
                                    </el-select>
                                    <el-tag v-if="domainDirectContainerID" type="primary" size="small">{{ domainDirectContainerID }}</el-tag>
                                </div>

                                <!-- 数据列表 -->
                                <el-table
                                    :data="domainDirectList"
                                    v-loading="domainDirectLoading"
                                    stripe
                                    style="width: 100%;"
                                    :empty-text="$t(domainDirectContainerID ? 'network.noDomainDirectRules' : 'network.selectContainerToQuery')"
                                >
                                    <el-table-column type="index" :label="$t('network.index')" width="80" align="center" />
                                    <el-table-column :label="$t('network.directDomain')" align="center">
                                        <template #default="{ row }">
                                            <span style="font-family: monospace;">{{ row.domain }}</span>
                                        </template>
                                    </el-table-column>
                                </el-table>
                            </template>
                        </div>
                    </div>

                    <!-- 设置域名直连弹窗 -->
                    <el-dialog
                        v-model="showDomainDirectDialog"
                        :title="$t('network.setDomainDirect')"
                        width="600px"
                        :close-on-click-modal="false"
                        @closed="resetDomainDirectDialog"
                    >
                        <el-form label-width="90px" style="padding-right: 10px;">
                            <!-- 选择容器 -->
                            <el-form-item :label="$t('network.selectContainer')" required>
                                <el-select
                                    v-model="domainDirectForm.containerID"
                                    :placeholder="$t('network.selectVpcContainer')"
                                    style="width: 100%;"
                                    filterable
                                    :loading="domainDirectVpcLoading"
                                >
                                    <el-option
                                        v-for="ct in domainDirectVpcContainers"
                                        :key="ct.containerName"
                                        :label="ct.containerName"
                                        :value="ct.containerName"
                                    >
                                        <span>{{ ct.containerName }}</span>
                                        <span style="color: #909399; font-size: 12px; margin-left: 8px;">{{ ct.containerIP }}</span>
                                    </el-option>
                                    <template #empty>
                                        <div style="padding: 8px 12px; color: #909399; font-size: 13px; text-align: center;">
                                            {{ $t(domainDirectVpcLoading ? 'common.loading' : 'network.noVpcContainers') }}
                                        </div>
                                    </template>
                                </el-select>
                            </el-form-item>

                            <!-- 域名列表 -->
                            <el-form-item :label="$t('network.directDomain')" required>
                                <div style="width: 100%;">
                                    <div
                                        v-for="(rule, idx) in domainDirectForm.rules"
                                        :key="idx"
                                        style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;"
                                    >
                                        <el-select v-model="rule.prefix" style="width: 180px;" size="default">
                                            <el-option label="domain:（子域名、IP）" value="domain:" />
                                            <el-option label="full:（完整域名、IP）" value="full:" />
                                            <el-option label="keyword:（关键字）" value="keyword:" />
                                        </el-select>
                                        <el-input
                                            v-model="rule.value"
                                            :placeholder="domainRulePlaceholder(rule.prefix)"
                                            style="flex: 1;"
                                            size="default"
                                        />
                                        <el-button
                                            :icon="Delete"
                                            circle
                                            size="small"
                                            type="danger"
                                            plain
                                            :disabled="domainDirectForm.rules.length === 1"
                                            @click="domainDirectForm.rules.splice(idx, 1)"
                                        />
                                    </div>
                                    <el-button size="small" @click="domainDirectForm.rules.push({ prefix: 'domain:', value: '' })">+ {{ $t('network.addDomain') }}</el-button>
                                </div>
                            </el-form-item>

                            <!-- 说明 -->
                            <el-form-item>
                                <div style="background: #f5f7fa; border-radius: 4px; padding: 10px 12px; font-size: 12px; color: #606266; line-height: 1.9; width: 100%;">
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:</code> {{ $t('network.domainRuleDesc') }}</div>
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:</code> {{ $t('network.fullRuleDesc') }}</div>
                                    <div><code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:</code> {{ $t('network.keywordRuleDesc') }}</div>
                                </div>
                            </el-form-item>
                        </el-form>
                        <template #footer>
                            <el-button @click="showDomainDirectDialog = false">{{ $t('common.cancel') }}</el-button>
                            <el-button
                                type="primary"
                                :loading="domainDirectSubmitLoading"
                                @click="submitDomainDirect"
                            >{{ $t('common.confirm') }}</el-button>
                        </template>
                    </el-dialog>
                </el-tab-pane>



                 <el-tab-pane :label="$t('network.privateNic')" name="private-nic">
                    <!-- 功能说明 -->
                    <div style="margin: 0 20px 20px 20px; padding: 12px 16px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                        <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 8px;">💡 {{ $t('network.privateNicDesc') }}</div>
                        <div style="font-size: 13px; line-height: 1.8; color: #606266;">
                            {{ $t('network.privateNicDescDetail') }}

                        </div>
                    </div>
                    <div class="vpc-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="$t('common.selectGroupPlaceholder')">
                                    <el-option :label="$t('common.all')" value="全部"></el-option>
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
                            <div class="panel-header">
                                <span>{{ $t('network.privateNicList') }}</span>
                                <el-button type="primary" v-if="selectedDeviceIP" @click="handleAddPrivateNic"
                                    style="margin-top: 10px;">{{ $t('network.createNic') }}</el-button>
                            </div>
                            <el-table :data="privateNicList" v-loading="privateNicLoading" stripe>
                                <el-table-column prop="name" :label="$t('network.nicName')" align="center" min-width="150" />
                                <!-- <el-table-column prop="ip" :label="$t('network.gateway')" align="center" min-width="150" />
                                <el-table-column prop="mask" label="子网掩码" align="center" min-width="150" /> -->
                                <el-table-column prop="cidr" label="CIDR" align="center" min-width="150" />
                                <el-table-column :label="$t('common.operation')" width="200" align="center" fixed="right">
                                    <template #default="scope">
                                        <el-button type="primary" size="small"
                                            @click="handleEditPrivateNic(scope.row)">{{ $t('common.edit') }}</el-button>
                                        <el-button type="danger" size="small"
                                            @click="handleDeletePrivateNic(scope.row)">{{ $t('common.delete') }}</el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                        </div>
                    </div>
                </el-tab-pane>
                <el-tab-pane :label="$t('network.publicNic')" name="public-nic">
                    <!-- 功能说明 -->
                    <div style="margin: 0 20px 20px 20px; padding: 12px 16px; background: #f0f9ff; border-left: 4px solid #67C23A; border-radius: 4px;">
                        <div style="font-weight: bold; color: #67C23A; font-size: 14px; margin-bottom: 8px;">💡 {{ $t('network.publicNicDesc') }}</div>
                        <div style="font-size: 13px; line-height: 1.8; color: #606266;">
                            {{ $t('network.publicNicDescDetail') }}

                        </div>
                    </div>
                    <div class="vpc-content">
                        <div class="left-panel">
                            <div class="panel-header"
                                style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.deviceList') }}</span>
                                <el-select v-model="localGroupFilter" size="small" style="width: 90px;"
                                    :placeholder="$t('common.selectGroupPlaceholder')">
                                    <el-option :label="$t('common.all')" value="全部"></el-option>
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
                        <div class="right-panel" style="padding: 20px; overflow-y: auto;">
                            <div class="panel-header" style="display: flex; justify-content: space-between; align-items: center;">
                                <span>{{ $t('network.publicNic') }}</span>
                            </div>
                            <div v-for="(nic, index) in publicNicList" :key="index" style="margin-bottom: 20px;">
                                <div style="font-weight: bold; margin-top: 15px; font-size: 16px;color: #000;">{{ nic.name }}</div>
                                <el-form :model="nic" label-width="280px" style="max-width: 500px;">
                                    <el-form-item label="网关">
                                        <el-input v-model="nic.gw" placeholder="例如: 10.10.0.1" disabled />
                                    </el-form-item>
                                    <el-form-item label="subnet">
                                        <el-input v-model="nic.subnet" placeholder="例如: 10.10.0.0/16" disabled />
                                    </el-form-item>
                                    
                                    <!-- 更新按钮说明 -->
                                    <div v-if="nic.type === 'macvlan'" style="margin-bottom: 15px; padding: 12px; background: #fff7e6; border-left: 4px solid #E6A23C; border-radius: 4px;">
                                        <div style="font-weight: bold; color: #E6A23C; font-size: 13px; margin-bottom: 8px;">⚠️ {{ $t('network.updateMacVlanTitle') }}</div>
                                        <div style="font-size: 12px; line-height: 1.8; color: #606266;">
                                            <div style="margin-bottom: 4px;"><strong>功能说明：</strong>用于在更换局域网网段后，将物理网卡的最新网络信息同步到MacVlan配置中。</div>
                                            <div style="margin-bottom: 4px;"><strong>执行过程：</strong></div>
                                            <div style="padding-left: 16px; margin-bottom: 4px;">1. 自动关闭该设备上所有正在运行的虚拟机和容器</div>
                                            <div style="padding-left: 16px; margin-bottom: 4px;">2. 同步物理网卡的网关和子网信息到MacVlan配置</div>
                                            <div style="padding-left: 16px; margin-bottom: 8px;">3. 完成后需要手动为每个虚拟机/容器重新设置MacVlan IP地址</div>
                                            <div style="color: #E6A23C;"><strong>⚠️ 注意：</strong>执行此操作会中断所有虚拟机/容器的运行，请谨慎操作！</div>
                                        </div>
                                    </div>
                                    
                                    <el-form-item v-if="nic.type === 'macvlan'">
                                        <el-button type="primary" @click="handleUpdatePublicNic(nic)" :loading="nic.loading">{{ $t('network.updateMacVlan') }}</el-button>
                                    </el-form-item>
                                </el-form>
                            </div>
                            <el-empty v-if="publicNicList.length === 0 && !publicNicLoading" :description="$t('network.noPublicNicData')" />
                        </div>
                    </div>
                </el-tab-pane>
                
               
                <!-- 使用说明标签页 -->
                <el-tab-pane :label="$t('network.usageGuide')" name="help">
                    <!-- 功能使用说明面板 -->
                    <div v-if="$i18n.locale === 'zh-CN'" style="margin: 20px; padding: 0; display: flex; flex-direction: column; gap: 16px;">

                        <!-- ① 节点管理说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #9333EA; border-radius: 4px;">
                            <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;">
                                <div style="font-weight: bold; color: #9333EA; font-size: 14px;">
                                    📖 节点管理说明
                                </div>
                            </div>
                            
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <!-- 1. 功能概述 -->
                                <el-collapse-item name="overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 功能概述</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">网络代理分组功能允许您为虚拟机/容器配置代理节点，实现网络加速和优化。通过配置代理分组，您可以：</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li>为不同的虚拟机/容器指定不同的网络代理节点</li>
                                            <li>实现网络流量的智能路由和负载均衡</li>
                                            <li>提升特定应用的网络访问速度</li>
                                            <li>支持多种主流代理协议，灵活配置</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <!-- 2. 配置方式 -->
                                <el-collapse-item name="config-methods" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">🔧 配置方式说明</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 6px;">方式一：订阅地址（推荐）</div>
                                            <p style="margin: 0 0 8px 0;">通过输入订阅链接，自动获取和更新多个代理节点。</p>
                                            <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-bottom: 8px;">
                                                <strong>优点：</strong>
                                                <ul style="margin: 4px 0; padding-left: 20px;">
                                                    <li>自动获取多个节点，无需手动配置</li>
                                                    <li>服务商更新节点后自动同步</li>
                                                    <li>配置简单，只需填入一个订阅链接</li>
                                                </ul>
                                            </div>
                                            <div style="background: #fff7e6; padding: 8px 12px; border-radius: 4px;">
                                                <strong>适用场景：</strong>从第三方服务商购买了代理服务
                                            </div>
                                        </div>
                                        
                                        <div>
                                            <div style="font-weight: bold; color: #67C23A; margin-bottom: 6px;">方式二：手动添加代理协议</div>
                                            <p style="margin: 0 0 8px 0;">手动输入具体的代理协议配置信息。</p>
                                            <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-bottom: 8px;">
                                                <strong>优点：</strong>
                                                <ul style="margin: 4px 0; padding-left: 20px;">
                                                    <li>完全自主控制，可使用自建节点</li>
                                                    <li>支持批量添加多个节点</li>
                                                    <li>支持8种主流代理协议</li>
                                                </ul>
                                            </div>
                                            <div style="background: #fff7e6; padding: 8px 12px; border-radius: 4px;">
                                                <strong>适用场景：</strong>拥有自建代理服务器或单独的节点配置
                                            </div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 3. 订阅地址配置步骤 -->
                                <el-collapse-item name="subscription-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 订阅地址配置步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; margin-bottom: 6px;">步骤1：获取订阅地址</div>
                                            <p style="margin: 0 0 6px 0;">从第三方代理服务商处获取订阅链接。常见的服务商包括：</p>
                                            <ul style="margin: 0 0 8px 0; padding-left: 20px;">
                                                <li>各大机场服务商的用户中心</li>
                                                <li>VPN服务提供商的订阅页面</li>
                                                <li>私有代理服务的订阅接口</li>
                                            </ul>
                                            <div style="background: #f0f9ff; padding: 8px 12px; border-radius: 4px; border-left: 3px solid #409EFF;">
                                                💡 <strong>提示：</strong>订阅链接通常是以 <code>http://</code> 或 <code>https://</code> 开头的URL地址
                                            </div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; margin-bottom: 6px;">步骤2：添加到客户端</div>
                                            <ol style="margin: 0; padding-left: 20px;">
                                                <li>点击"新增分组"按钮</li>
                                                <li>选择类型为"订阅地址"</li>
                                                <li>输入分组别名（如：香港节点、美国节点等）</li>
                                                <li>将订阅链接粘贴到"URL地址"输入框</li>
                                                <li>点击"确定"保存</li>
                                            </ol>
                                        </div>

                                        <div>
                                            <div style="font-weight: bold; margin-bottom: 6px;">步骤3：更新订阅</div>
                                            <p style="margin: 0;">系统会自动定期更新订阅内容，也可以在分组列表中手动点击"刷新"按钮立即更新。</p>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 4. 代理协议配置步骤 -->
                                <el-collapse-item name="protocol-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚙️ 代理协议配置步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0 0 12px 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;">
                                                <strong>选择配置类型：</strong>点击"新增分组"，选择类型为"代理协议"
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>输入分组别名：</strong>为这组节点起一个便于识别的名称
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>选择协议类型：</strong>从下拉菜单中选择要配置的协议类型（如vmess、vless、ss等）
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>填写协议信息：</strong>按照对应的格式填写代理节点配置
                                                <div style="background: #f5f7fa; padding: 8px; border-radius: 4px; margin-top: 4px;">
                                                    💡 点击输入框旁边的"支持的协议格式"可查看详细格式说明
                                                </div>
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>批量添加：</strong>可以使用中英文逗号分隔多个节点配置，一次性添加
                                            </li>
                                            <li>
                                                <strong>保存配置：</strong>点击"确定"完成添加
                                            </li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <!-- 5. 支持的协议格式 -->
                                <el-collapse-item name="protocols" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 支持的协议格式</span>
                                    </template>
                                    <div style="font-size: 12px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">1. VMess协议</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                vmess://base64(json配置)
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">标准的VMess协议格式，使用base64编码的JSON配置</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">2. VLESS协议</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                vless://uuid@server:port?type=ws&security=tls&sni=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">支持WebSocket、gRPC等传输方式</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">3. Shadowsocks协议</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                ss://base64(method:password)@server:port#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">遵循SIP002标准格式</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">4. Trojan协议</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                trojan://password@server:port?type=tcp&security=tls&sni=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">支持TCP、WebSocket等传输方式</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">5. SOCKS5代理</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                server/port/username/password/remarks/
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">用户名和密码可以为空，用斜杠分隔</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">6. HTTP代理</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                server/port/username/password/remarks/
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">格式与SOCKS5相同</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">7. WireGuard协议</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                wireguard://privateKey@server:port?publickey=xxx&address=xxx&mtu=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">现代化的VPN协议，性能优秀</div>
                                        </div>

                                        <div>
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">8. Hysteria2协议</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                hysteria2://auth@server:port?sni=xxx&insecure=1&obfs-password=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">基于QUIC的高性能代理协议</div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 6. 常见问题 -->
                                <el-collapse-item name="faq" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">❓ 常见问题</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q1: 配置保存失败怎么办？</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;">可能的原因：</p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>订阅地址格式不正确（需要是完整的URL）</li>
                                                    <li>协议格式不符合规范</li>
                                                    <li>网络连接问题</li>
                                                </ul>
                                                <p style="margin: 8px 0 0 0;"><strong>解决方法：</strong>检查输入格式，确保网络连接正常</p>
                                            </div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q2: 节点无法连接怎么办？</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;">可能的原因：</p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>节点服务器不可用或过期</li>
                                                    <li>配置参数错误</li>
                                                    <li>本地网络限制</li>
                                                </ul>
                                                <p style="margin: 8px 0 0 0;"><strong>解决方法：</strong>尝试更换其他节点，或联系服务商确认节点状态</p>
                                            </div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q3: 速度较慢怎么办？</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;"><strong>建议：</strong></p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>选择地理位置较近的节点</li>
                                                    <li>避开高峰时段</li>
                                                    <li>尝试不同的协议类型</li>
                                                    <li>联系服务商升级套餐</li>
                                                </ul>
                                            </div>
                                        </div>

                                        <div>
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q4: 订阅无法更新怎么办？</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;">可能的原因：</p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>订阅链接已失效或过期</li>
                                                    <li>服务商服务器维护中</li>
                                                    <li>网络连接问题</li>
                                                </ul>
                                                <p style="margin: 8px 0 0 0;"><strong>解决方法：</strong>联系服务商获取最新的订阅链接</p>
                                            </div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 7. 重要提示 -->
                                <el-collapse-item name="notice" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚠️ 重要提示</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="background: #fff7e6; padding: 12px; border-radius: 4px; border-left: 4px solid #E6A23C; margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 8px;">🚨 功能互斥说明</div>
                                            <p style="margin: 0 0 8px 0;">网络代理分组功能与MacVlan公有网卡功能<strong>互相排斥，不能同时使用</strong>。</p>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>使用MacVlan后，节点管理功能将被禁用</li>
                                                <li>配置节点管理后，建议不要再设置MacVlan</li>
                                                <li>切换功能前请先移除原有配置</li>
                                            </ul>
                                        </div>

                                        <div style="background: #f0f9ff; padding: 12px; border-radius: 4px; border-left: 4px solid #409EFF; margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 8px;">💡 使用建议</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>订阅地址更新频率建议设置为每天或每周</li>
                                                <li>定期检查节点可用性，及时更新失效节点</li>
                                                <li>建议配置多个分组，便于不同场景切换</li>
                                                <li>重要应用建议使用稳定性更好的节点</li>
                                            </ul>
                                        </div>

                                        <div style="background: #fef0f0; padding: 12px; border-radius: 4px; border-left: 4px solid #F56C6C;">
                                            <div style="font-weight: bold; color: #F56C6C; margin-bottom: 8px;">🔒 安全提醒</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>请从可信的服务商获取订阅地址</li>
                                                <li>不要使用来源不明的免费节点</li>
                                                <li>定期更换密码和敏感信息</li>
                                                <li>注意保护个人隐私和数据安全</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ② 节点分配说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                            <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 12px;">
                                🔗 节点分配说明
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="vpc-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 功能概述</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">节点分配功能用于将已配置好的代理节点指定给具体的云机（容器），实现精细化的网络代理管控。</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li>可以为每台云机独立指定代理节点（指定模式）</li>
                                            <li>也可以在某个分组内随机分配节点（随机模式）</li>
                                            <li>支持批量为多台云机清除已分配的VPC节点</li>
                                            <li>支持为容器开启/{{ $t('network.closeDnsWhitelist') }}</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="vpc-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 操作步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;"><strong>选择设备：</strong>在左侧设备列表中点击目标设备</li>
                                            <li style="margin-bottom: 8px;"><strong>查看已分配节点：</strong>右侧面板展示该设备下所有云机的VPC节点分配情况</li>
                                            <li style="margin-bottom: 8px;"><strong>{{ $t('network.assignVpc') }}：</strong>点击"{{ $t('network.assignVpc') }}"，按照三步向导完成分配
                                                <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-top: 6px;">
                                                    <div style="margin-bottom: 4px;">① <strong>选择云机：</strong>勾选需要配置的一台或多台云机</div>
                                                    <div style="margin-bottom: 4px;">② <strong>选择{{ $t('common.group') }}</strong>从已有节点分组中选择，并决定是"指定节点"还是"随机节点"</div>
                                                    <div>③ <strong>选择节点：</strong>（指定模式下）选择具体的代理节点后点击"确定"</div>
                                                </div>
                                            </li>
                                            <li style="margin-bottom: 8px;"><strong>清除VPC节点：</strong>可单独清除某台云机的节点，或勾选多台后批量清除</li>
                                            <li><strong>DNS白名单：</strong>点击"开启DNS"可为该云机启用DNS白名单，再次点击则关闭</li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="vpc-notice" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚠️ 注意事项</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="background: #fff7e6; padding: 10px 12px; border-radius: 4px; border-left: 3px solid #E6A23C;">
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>设置 MacVlan 后节点管理与节点分配功能将<strong>不再可用</strong></li>
                                                <li>分配前请确保已在"节点管理"中创建好代理分组和节点</li>
                                                <li>随机模式下每次重启云机可能会使用分组中不同的节点</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ③ 域名过滤说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #67C23A; border-radius: 4px;">
                            <div style="font-weight: bold; color: #67C23A; font-size: 14px; margin-bottom: 12px;">
                                🌐 域名过滤说明
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="domain-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 功能概述</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">域名过滤功能可以为云机容器或整个设备设置"域名屏蔽规则"，符合规则的域名请求将被代理<strong>直接丢弃拦截</strong>，不会转发也不会响应。</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li><strong>容器域名过滤：</strong>仅对指定的单个云机容器生效</li>
                                            <li><strong>全局域名过滤：</strong>对设备下所有云机容器生效</li>
                                            <li>规则<strong>同时支持域名和 IP 地址</strong>两种格式，可在域名过渡期间并行填写新旧域名及对应 IP</li>
                                            <li>支持三种规则匹配模式：子域名/IP 匹配、完整匹配、关键字匹配</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-rule-types" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 规则匹配类型说明</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">domain:</code>
                                                <span style="font-weight: bold;">子域名 / IP 匹配（推荐）</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">匹配该域名及其所有子域名，也支持直接填写 IP 地址。例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:example.com</code> 会同时匹配 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code> 和 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code>；填写 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:192.168.1.1</code> 则直接匹配该 IP</div>
                                        </div>
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">full:</code>
                                                <span style="font-weight: bold;">完整域名 / IP 精确匹配</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">仅精确匹配完整域名或完整 IP。例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:www.example.com</code> 仅匹配 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code>，不匹配 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code>；<code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:1.2.3.4</code> 仅匹配该 IP</div>
                                        </div>
                                        <div>
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">keyword:</code>
                                                <span style="font-weight: bold;">关键字匹配</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">匹配包含该关键字的任意域名或 IP。例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:google</code> 匹配所有含 "google" 的域名；<code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:192.168</code> 匹配所有含该字段的 IP</div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 操作步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 6px;">{{ $t('network.setContainerDomainFilter') }}：</div>
                                            <ol style="margin: 0; padding-left: 20px;">
                                                <li>在左侧选择目标设备</li>
                                                <li>点击"{{ $t('network.setContainerDomainFilter') }}"按钮</li>
                                                <li>在弹窗中选择目标容器</li>
                                                <li>添加域名规则（可添加多条），选择匹配类型并填写域名</li>
                                                <li>点击"确定"保存</li>
                                            </ol>
                                        </div>
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #67C23A; margin-bottom: 6px;">{{ $t('network.setGlobalDomainFilter') }}：</div>
                                            <ol style="margin: 0; padding-left: 20px;">
                                                <li>点击"{{ $t('network.setGlobalDomainFilter') }}"按钮</li>
                                                <li>添加域名规则后点击"确定"保存</li>
                                                <li>规则将对该设备下所有容器生效</li>
                                            </ol>
                                        </div>
                                        <div>
                                            <div style="font-weight: bold; color: #F56C6C; margin-bottom: 6px;">查询与清除：</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>在容器下拉框中选择容器可查看该容器的过滤规则</li>
                                                <li>点击"{{ $t('network.queryGlobalDomainFilter') }}"可查看全局规则</li>
                                                <li>点击"清除过滤"可删除当前查看的规则</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ④ 域名直连说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #E6A23C; border-radius: 4px;">
                            <div style="font-weight: bold; color: #E6A23C; font-size: 14px; margin-bottom: 12px;">
                                🔗 域名直连说明
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="domain-direct-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 功能概述</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">域名直连功能可以为指定容器设置"直连白名单"，符合规则的域名请求将<strong>绕过 VPC 代理</strong>，直接走本地网络连接。</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li><strong>仅对指定容器生效：</strong>每条规则绑定到单个云机容器</li>
                                            <li>规则<strong>同时支持域名和 IP 地址</strong>两种格式</li>
                                            <li>支持三种规则匹配模式：子域名/IP 匹配、完整匹配、关键字匹配</li>
                                            <li>适用于某些域名需要直接连接、不走代理的场景</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-direct-rule-types" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 规则匹配类型说明</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">domain:</code>
                                                <span style="font-weight: bold;">子域名 / IP 匹配（推荐）</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">匹配该域名及其所有子域名。例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:example.com</code> 会同时匹配 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code> 和 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code></div>
                                        </div>
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">full:</code>
                                                <span style="font-weight: bold;">完整域名 / IP 精确匹配</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">仅精确匹配完整域名或完整 IP。例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:www.example.com</code> 仅匹配 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code>，不匹配 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code></div>
                                        </div>
                                        <div>
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">keyword:</code>
                                                <span style="font-weight: bold;">关键字匹配</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">匹配包含该关键字的任意域名或 IP。例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:google</code> 匹配所有含 "google" 的域名</div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-direct-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 操作步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li>在左侧选择目标设备</li>
                                            <li>点击"设置域名直连"按钮</li>
                                            <li>在弹窗中选择目标容器</li>
                                            <li>添加直连规则（可添加多条），选择匹配类型并填写域名或 IP</li>
                                            <li>点击"确定"保存</li>
                                            <li>在容器下拉框中选择容器可查询当前直连规则</li>
                                            <li>点击"清除直连"可删除该容器的所有直连规则</li>
                                        </ol>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ⑤ 私有网卡说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                            <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 12px;">
                                🔌 私有网卡说明
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="private-nic-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 功能概述</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">私有网卡（mytBridge）会在设备内创建一个独立的虚拟网桥，为虚拟机/容器分配该网桥下的私有IP地址，从而实现：</p>
                                        <ul style="margin: 0 0 8px 0; padding-left: 20px;">
                                            <li>同一设备上不同云机之间的<strong>网络隔离</strong></li>
                                            <li>仍可使用"节点管理"中配置的IP代理功能</li>
                                            <li>不影响云机对外访问互联网</li>
                                        </ul>
                                        <div style="background: #f0f9ff; padding: 8px 12px; border-radius: 4px; border-left: 3px solid #409EFF;">
                                            💡 <strong>适用场景：</strong>需要多个云机使用不同代理IP，同时防止云机间相互访问
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="private-nic-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 操作步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;"><strong>选择设备：</strong>在左侧列表中点击目标设备</li>
                                            <li style="margin-bottom: 8px;"><strong>创建网卡：</strong>点击"创建网卡"按钮，填写以下信息：
                                                <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-top: 6px;">
                                                    <div style="margin-bottom: 4px;">• <strong>自定义名称：</strong>网卡名称前缀固定为 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">mytBridge_</code>，填写后缀部分</div>
                                                    <div>• <strong>CIDR：</strong>网段地址，例如 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">172.20.0.0/16</code>，决定虚拟IP的范围</div>
                                                </div>
                                            </li>
                                            <li style="margin-bottom: 8px;"><strong>编辑/删除：</strong>在列表中对已有私有网卡进行编辑或删除操作</li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="private-nic-cidr" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 CIDR填写说明</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">CIDR（无类别域间路由）格式为 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">IP地址/前缀长度</code>，例如：</p>
                                        <div style="background: #f5f5f5; padding: 8px 12px; border-radius: 4px; font-family: monospace; margin-bottom: 8px;">
                                            172.20.0.0/16 → 可分配 IP 范围：172.20.0.1 ~ 172.20.255.254<br>
                                            192.168.100.0/24 → 可分配 IP 范围：192.168.100.1 ~ 192.168.100.254
                                        </div>
                                        <div style="background: #fff7e6; padding: 8px 12px; border-radius: 4px; border-left: 3px solid #E6A23C;">
                                            ⚠️ <strong>注意：</strong>请避免与设备所在局域网网段冲突，建议使用 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">172.16.0.0/12</code> 或 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">10.0.0.0/8</code> 范围内的地址
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ⑤ 公有网卡说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #67C23A; border-radius: 4px;">
                            <div style="font-weight: bold; color: #67C23A; font-size: 14px; margin-bottom: 12px;">
                                🌍 公有网卡（MacVlan）说明
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="public-nic-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 功能概述</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">公有网卡（MacVlan）模式下，虚拟机/容器将直接使用设备所在局域网的网关和子网，与物理设备处于同一网段。</p>
                                        <div style="margin-bottom: 10px;">
                                            <div style="font-weight: bold; color: #67C23A; margin-bottom: 4px;">✅ 优点：</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>云机拥有局域网内真实IP，访问局域网资源更方便</li>
                                                <li>网络延迟更低，接近物理直连</li>
                                            </ul>
                                        </div>
                                        <div>
                                            <div style="font-weight: bold; color: #F56C6C; margin-bottom: 4px;">❌ 限制：</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>云机之间<strong>无法进行网络隔离</strong></li>
                                                <li>设置后<strong>无法使用节点管理的IP代理功能</strong></li>
                                                <li>与"节点管理""节点分配"功能互斥</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="public-nic-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 操作步骤</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;"><strong>选择设备：</strong>在左侧设备列表中点击目标设备</li>
                                            <li style="margin-bottom: 8px;"><strong>查看网卡：</strong>右侧面板展示该设备物理网卡的当前配置（网关、子网掩码等）</li>
                                            <li style="margin-bottom: 8px;"><strong>设置MacVlan IP：</strong>为虚拟机/容器分配一个与设备同网段的IP地址，点击"设置MacVlan IP"后保存</li>
                                            <li style="margin-bottom: 8px;"><strong>更新MacVlan配置：</strong>当设备更换局域网网段后，点击"更新MacVlan配置"将最新的网关/子网信息同步过来
                                                <div style="background: #fef0f0; padding: 8px 12px; border-radius: 4px; margin-top: 6px; border-left: 3px solid #F56C6C;">
                                                    ⚠️ 此操作会<strong>自动关闭所有正在运行的虚拟机和容器</strong>，执行后需要手动重新为每个容器设置MacVlan IP
                                                </div>
                                            </li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="public-nic-notice" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚠️ 重要提示</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="background: #fff7e6; padding: 12px; border-radius: 4px; border-left: 4px solid #E6A23C; margin-bottom: 10px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 6px;">🚨 功能互斥警告</div>
                                            <p style="margin: 0;">启用 MacVlan（公有网卡）后，<strong>节点管理、节点分配功能将被禁用</strong>。如需切换，请先在公有网卡中移除所有 MacVlan 配置后再使用节点管理功能。</p>
                                        </div>
                                        <div style="background: #f0f9ff; padding: 12px; border-radius: 4px; border-left: 4px solid #409EFF;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 6px;">💡 IP分配建议</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>为每台云机分配的IP必须与设备在同一网段，且不能与其他设备冲突</li>
                                                <li>建议从网段末尾开始分配，避免与 DHCP 自动分配的IP冲突</li>
                                                <li>更换局域网网段后务必及时执行"更新MacVlan配置"</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                    </div>
<div v-else style="margin: 20px; padding: 0; display: flex; flex-direction: column; gap: 16px;">

                        <!-- ① Node Management Guide -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #9333EA; border-radius: 4px;">
                            <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;">
                                <div style="font-weight: bold; color: #9333EA; font-size: 14px;">
                                    📖 Node Management Guide
                                </div>
                            </div>
                            
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <!-- 1. Feature Overview -->
                                <el-collapse-item name="overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 Feature Overview</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">The proxy group feature allows you to configure proxy nodes for VMs/Containers to achieve network acceleration and optimization. You can:</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li>Assign different proxy nodes to different VMs/Containers</li>
                                            <li>Enable intelligent routing and load balancing of network traffic</li>
                                            <li>Improve network access speed for specific applications</li>
                                            <li>Support multiple mainstream proxy protocols with flexible configuration</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <!-- 2. 配置方式 -->
                                <el-collapse-item name="config-methods" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">🔧 Configuration Methods</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 6px;">Method 1: Subscription URL (Recommended)</div>
                                            <p style="margin: 0 0 8px 0;">Automatically fetch and update multiple proxy nodes by inputting a subscription link.</p>
                                            <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-bottom: 8px;">
                                                <strong>Advantages:</strong>
                                                <ul style="margin: 4px 0; padding-left: 20px;">
                                                    <li>Automatically fetches multiple nodes without manual configuration</li>
                                                    <li>Automatically syncs updates from the service provider</li>
                                                    <li>Simple configuration requiring only one subscription link</li>
                                                </ul>
                                            </div>
                                            <div style="background: #fff7e6; padding: 8px 12px; border-radius: 4px;">
                                                <strong>适用场景：</strong>从第三方服务商购买了代理服务
                                            </div>
                                        </div>
                                        
                                        <div>
                                            <div style="font-weight: bold; color: #67C23A; margin-bottom: 6px;">Method 2: Manually Add Proxy Protocol</div>
                                            <p style="margin: 0 0 8px 0;">Manually input specific proxy protocol configuration details.</p>
                                            <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-bottom: 8px;">
                                                <strong>Advantages:</strong>
                                                <ul style="margin: 4px 0; padding-left: 20px;">
                                                    <li>Full control, supporting self-hosted nodes</li>
                                                    <li>Supports adding multiple nodes in bulk</li>
                                                    <li>Supports 8 mainstream proxy protocols</li>
                                                </ul>
                                            </div>
                                            <div style="background: #fff7e6; padding: 8px 12px; border-radius: 4px;">
                                                <strong>适用场景：</strong>拥有自建代理服务器或单独的节点配置
                                            </div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 3. Subscription Config Steps -->
                                <el-collapse-item name="subscription-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 Subscription Config Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; margin-bottom: 6px;">Step 1: Get Subscription URL</div>
                                            <p style="margin: 0 0 6px 0;">Obtain a subscription link from third-party proxy providers. Common providers include:</p>
                                            <ul style="margin: 0 0 8px 0; padding-left: 20px;">
                                                <li>User centers of major proxy services</li>
                                                <li>Subscription pages of VPN providers</li>
                                                <li>Private proxy service subscription APIs</li>
                                            </ul>
                                            <div style="background: #f0f9ff; padding: 8px 12px; border-radius: 4px; border-left: 3px solid #409EFF;">
                                                💡 <strong>Tip:</strong>Subscription links typically start with <code>http://</code> 或 <code>https://</code> URL addresses
                                            </div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; margin-bottom: 6px;">Step 2: Add to Client</div>
                                            <ol style="margin: 0; padding-left: 20px;">
                                                <li>Click the "Add Group" button</li>
                                                <li>Select "Subscription URL" as the type</li>
                                                <li>Enter a group alias (e.g., HK Nodes, US Nodes)</li>
                                                <li>Paste the subscription URL into the input field</li>
                                                <li>Click "Confirm" to save</li>
                                            </ol>
                                        </div>

                                        <div>
                                            <div style="font-weight: bold; margin-bottom: 6px;">Step 3: Update Subscription</div>
                                            <p style="margin: 0;">The system regularly updates subscriptions automatically, or you can click "Refresh" in the group list to update immediately.</p>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 4. Proxy Protocol Config Steps -->
                                <el-collapse-item name="protocol-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚙️ Proxy Protocol Config Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0 0 12px 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;">
                                                <strong>Select Config Type:</strong>Click "Add Group", select "Proxy Protocol" type
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>Input Group Alias:</strong>Give the node group an easily recognizable name
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>Select Protocol Type:</strong>Choose the protocol from the dropdown (e.g., vmess, vless, ss, etc.)
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>Fill Protocol Info:</strong>Fill out the proxy node config according to the required format
                                                <div style="background: #f5f7fa; padding: 8px; border-radius: 4px; margin-top: 4px;">
                                                    💡 Click "Supported Protocol Formats" next to the input box to see detailed formats
                                                </div>
                                            </li>
                                            <li style="margin-bottom: 8px;">
                                                <strong>Bulk Add:</strong>You can use commas to separate multiple node configs for bulk insertion
                                            </li>
                                            <li>
                                                <strong>Save Configuration:</strong>Click "Confirm" to finish adding
                                            </li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <!-- 5. Supported Protocol Formats -->
                                <el-collapse-item name="protocols" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 Supported Protocol Formats</span>
                                    </template>
                                    <div style="font-size: 12px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">1. VMess Protocol</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                vmess://base64(json配置)
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Standard VMess protocol format, using base64 encoded JSON configs</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">2. VLESS Protocol</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                vless://uuid@server:port?type=ws&security=tls&sni=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Supports WebSocket, gRPC, and other transport methods</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">3. Shadowsocks Protocol</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                ss://base64(method:password)@server:port#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Complies with the SIP002 standard format</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">4. Trojan Protocol</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                trojan://password@server:port?type=tcp&security=tls&sni=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Supports TCP, WebSocket transport methods</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">5. SOCKS5 Proxy</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                server/port/username/password/remarks/
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Username and password can be empty, separated by slashes</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">6. HTTP Proxy</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                server/port/username/password/remarks/
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Format is the same as SOCKS5</div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">7. WireGuard Protocol</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                wireguard://privateKey@server:port?publickey=xxx&address=xxx&mtu=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">Modernized VPN protocol with excellent performance</div>
                                        </div>

                                        <div>
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 4px;">8. Hysteria2 Protocol</div>
                                            <div style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-family: monospace; overflow-x: auto;">
                                                hysteria2://auth@server:port?sni=xxx&insecure=1&obfs-password=xxx#remarks
                                            </div>
                                            <div style="color: #909399; margin-top: 4px;">High-performance proxy protocol based on QUIC</div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 6. Frequently Asked Questions (FAQ) -->
                                <el-collapse-item name="faq" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">❓ Frequently Asked Questions (FAQ)</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q1: Config saving failed?</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;">Possible causes:</p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>Incorrect subscription URL format (must be full URL)</li>
                                                    <li>Protocol format does not comply with specifications</li>
                                                    <li>Network connection problems</li>
                                                </ul>
                                                <p style="margin: 8px 0 0 0;"><strong>Solution:</strong>Check input format and ensure stable network connection</p>
                                            </div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q2: Node cannot connect?</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;">Possible causes:</p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>Node server is unavailable or expired</li>
                                                    <li>Configuration parameter error</li>
                                                    <li>Local network limitations</li>
                                                </ul>
                                                <p style="margin: 8px 0 0 0;"><strong>Solution:</strong>Try switching to another node, or contact service provider</p>
                                            </div>
                                        </div>

                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q3: Slow connection speed?</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;"><strong>Suggestions:</strong></p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>Select nodes that are geographically closer</li>
                                                    <li>Avoid peak hours</li>
                                                    <li>Try different protocol types</li>
                                                    <li>Contact provider to upgrade your plan</li>
                                                </ul>
                                            </div>
                                        </div>

                                        <div>
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 4px;">Q4: Subscription cannot update?</div>
                                            <div style="padding-left: 16px;">
                                                <p style="margin: 0 0 4px 0;">Possible causes:</p>
                                                <ul style="margin: 0; padding-left: 20px;">
                                                    <li>Subscription link has expired or is invalid</li>
                                                    <li>Service provider server under maintenance</li>
                                                    <li>Network connection problems</li>
                                                </ul>
                                                <p style="margin: 8px 0 0 0;"><strong>Solution:</strong>Contact the service provider for a new subscription link</p>
                                            </div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <!-- 7. Important Notices -->
                                <el-collapse-item name="notice" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚠️ Important Notices</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="background: #fff7e6; padding: 12px; border-radius: 4px; border-left: 4px solid #E6A23C; margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 8px;">🚨 Mutually Exclusive Features</div>
                                            <p style="margin: 0 0 8px 0;">Network Proxy Group feature and MacVlan Public NIC feature<strong>are mutually exclusive and cannot be used simultaneously</strong>。</p>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Enabling MacVlan will disable the Node Management features</li>
                                                <li>After configuring nodes, it is recommended NOT to setup MacVlan</li>
                                                <li>Remove existing configurations before switching features</li>
                                            </ul>
                                        </div>

                                        <div style="background: #f0f9ff; padding: 12px; border-radius: 4px; border-left: 4px solid #409EFF; margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 8px;">💡 Usage Suggestions</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Recommended to set subscription updates to daily or weekly</li>
                                                <li>Regularly check node availability and update dead nodes</li>
                                                <li>It is recommended to configure multiple groups for different scenarios</li>
                                                <li>Use more stable nodes for critical applications</li>
                                            </ul>
                                        </div>

                                        <div style="background: #fef0f0; padding: 12px; border-radius: 4px; border-left: 4px solid #F56C6C;">
                                            <div style="font-weight: bold; color: #F56C6C; margin-bottom: 8px;">🔒 Security Reminders</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Please obtain subscription URLs from trusted providers</li>
                                                <li>Do NOT use unknown free proxy nodes</li>
                                                <li>Periodically change passwords and sensitive data</li>
                                                <li>Always safeguard personal privacy and data security</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ② Node Allocation Guide -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                            <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 12px;">
                                🔗 Node Allocation Guide
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="vpc-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 Feature Overview</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">The Node Allocation feature assigns configured proxy nodes to specific VMs (containers) for fine-grained network control.</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li>You can independently assign a node to each VM (Assigned Mode)</li>
                                            <li>You can randomly assign a node from a group (Random Mode)</li>
                                            <li>Supports bulk clearing of assigned VPC nodes for multiple VMs</li>
                                            <li>支持为容器开启/{{ $t('network.closeDnsWhitelist') }}</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="vpc-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 Operation Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;"><strong>Select Device:</strong>Click the target device from the left device list</li>
                                            <li style="margin-bottom: 8px;"><strong>View Assigned Nodes:</strong>The right panel shows the VPC node allocation for all VMs on this device</li>
                                            <li style="margin-bottom: 8px;"><strong>{{ $t('network.assignVpc') }}：</strong>点击"{{ $t('network.assignVpc') }}"，按照三步向导完成分配
                                                <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-top: 6px;">
                                                    <div style="margin-bottom: 4px;">① <strong>Select VM:</strong>Check one or more VMs to configure</div>
                                                    <div style="margin-bottom: 4px;">② <strong>选择{{ $t('common.group') }}</strong>Choose from existing node groups and decide if it is "Assigned" or "Random"</div>
                                                    <div>③ <strong>Select Node:</strong>(In Assigned mode) select a specific node and click "Confirm"</div>
                                                </div>
                                            </li>
                                            <li style="margin-bottom: 8px;"><strong>Clear VPC Node:</strong>Clear specific VM nodes or do it in bulk on multiple selections</li>
                                            <li><strong>DNS Whitelist:</strong>Click "Enable DNS" to turn on DNS whitelist, click again to disable</li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="vpc-notice" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚠️ Cautions</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="background: #fff7e6; padding: 10px 12px; border-radius: 4px; border-left: 3px solid #E6A23C;">
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Enabling MacVlan will cause Node Management and Allocation to be<strong>UNAVAILABLE</strong></li>
                                                <li>Before allocating, ensure you have created proxy groups and nodes in Node Management</li>
                                                <li>In random mode, VMs may select a different node from the group each time they restart</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ③ Domain Filtering Guide -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #67C23A; border-radius: 4px;">
                            <div style="font-weight: bold; color: #67C23A; font-size: 14px; margin-bottom: 12px;">
                                🌐 Domain Filtering Guide
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="domain-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 Feature Overview</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">Domain Filtering sets domain blocking rules for specific containers or devices. Matching requests are<strong>DROPPED and INTERCEPTED directly</strong>，and will neither be forwarded nor responded to.</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li><strong>Container Domain Filtering:</strong>Applies ONLY to a specific VM container</li>
                                            <li><strong>Global Domain Filtering:</strong>Applies to ALL VM containers on the device</li>
                                            <li>Rules<strong>simultaneously support Domain Names and IPs</strong>, allowing concurrent inputs of new/old domains alongside IP addresses during transitions</li>
                                            <li>Supports 3 match modes: Subdomain/IP, Exact Match, and Keyword</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-rule-types" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 Match Rule Types Guide</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">domain:</code>
                                                <span style="font-weight: bold;">Subdomain / IP Match (Recommended)</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">Matches domains and all subdomains, including direct IPs.For example: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:example.com</code> will match both <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code> and <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code>; inputting <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:192.168.1.1</code> will directly block that IP</div>
                                        </div>
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">full:</code>
                                                <span style="font-weight: bold;">Exact Domain / IP Match</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">Matches strictly exact domain strings or exact IPs.For example: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:www.example.com</code> will ONLY match <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code>, and NOT match <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code>；<code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:1.2.3.4</code> will ONLY match该 IP</div>
                                        </div>
                                        <div>
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">keyword:</code>
                                                <span style="font-weight: bold;">Keyword Match</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">Matches any domain or IP containing the keyword.For example: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:google</code> will match any domain with "google";<code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:192.168</code> will match any IP containing that string</div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 Operation Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 6px;">{{ $t('network.setContainerDomainFilter') }}：</div>
                                            <ol style="margin: 0; padding-left: 20px;">
                                                <li>Select target device on the left</li>
                                                <li>点击"{{ $t('network.setContainerDomainFilter') }}"按钮</li>
                                                <li>Select target container in the pop-up</li>
                                                <li>Add domain rules (multiple allowed), choose match type and fill the domain</li>
                                                <li>Click "Confirm" to save</li>
                                            </ol>
                                        </div>
                                        <div style="margin-bottom: 12px;">
                                            <div style="font-weight: bold; color: #67C23A; margin-bottom: 6px;">{{ $t('network.setGlobalDomainFilter') }}：</div>
                                            <ol style="margin: 0; padding-left: 20px;">
                                                <li>点击"{{ $t('network.setGlobalDomainFilter') }}"按钮</li>
                                                <li>Add domain rules and click "Confirm" to save</li>
                                                <li>Rules will be applied to ALL containers on the device</li>
                                            </ol>
                                        </div>
                                        <div>
                                            <div style="font-weight: bold; color: #F56C6C; margin-bottom: 6px;">Query & Clear:</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Select a container in the dropdown to view its filter rules</li>
                                                <li>点击"{{ $t('network.queryGlobalDomainFilter') }}"可查看全局Rules</li>
                                                <li>Click "Clear Filter" to delete the currently viewed rules</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ④ Direct Domain Bypass Guide -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #E6A23C; border-radius: 4px;">
                            <div style="font-weight: bold; color: #E6A23C; font-size: 14px; margin-bottom: 12px;">
                                🔗 Direct Domain Bypass Guide
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="domain-direct-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 Feature Overview</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">Direct Domain assigns a "bypass whitelist" to specific containers. Matched domains will<strong>BYPASS VPC Proxies</strong>, and connect directly through local networks.</p>
                                        <ul style="margin: 0; padding-left: 20px;">
                                            <li><strong>Applying Only to Specific Containers:</strong>Each rule bounds to a single VM container</li>
                                            <li>Rules<strong>simultaneously support Domain Names and IPs</strong>两种格式</li>
                                            <li>Supports 3 match modes: Subdomain/IP, Exact Match, and Keyword</li>
                                            <li>Ideal for scenarios where certain domains require direct connections instead of proxies</li>
                                        </ul>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-direct-rule-types" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 Match Rule Types Guide</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">domain:</code>
                                                <span style="font-weight: bold;">Subdomain / IP Match (Recommended)</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">匹配该域名及其所有子域名。For example: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">domain:example.com</code> will match both <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code> and <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code></div>
                                        </div>
                                        <div style="margin-bottom: 10px;">
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">full:</code>
                                                <span style="font-weight: bold;">Exact Domain / IP Match</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">Matches strictly exact domain strings or exact IPs.For example: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">full:www.example.com</code> will ONLY match <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">www.example.com</code>, and NOT match <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">example.com</code></div>
                                        </div>
                                        <div>
                                            <div style="display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px;">
                                                <code style="background: #e8e8e8; padding: 1px 6px; border-radius: 3px; font-size: 12px;">keyword:</code>
                                                <span style="font-weight: bold;">Keyword Match</span>
                                            </div>
                                            <div style="padding-left: 16px; color: #909399;">Matches any domain or IP containing the keyword.For example: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">keyword:google</code> 匹配所有含 "google" 的域名</div>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="domain-direct-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 Operation Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li>Select target device on the left</li>
                                            <li>点击"设置域名直连"按钮</li>
                                            <li>Select target container in the pop-up</li>
                                            <li>Add direct rules (multiple allowed), choose match type and fill domain/IP</li>
                                            <li>Click "Confirm" to save</li>
                                            <li>Select a container in the dropdown to check active direct rules</li>
                                            <li>Click "Clear Bypass" to remove all existing direct rules for the container</li>
                                        </ol>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ⑤ Private NIC Guide -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #409EFF; border-radius: 4px;">
                            <div style="font-weight: bold; color: #409EFF; font-size: 14px; margin-bottom: 12px;">
                                🔌 Private NIC Guide
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="private-nic-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 Feature Overview</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">The Private NIC (mytBridge) creates an isolated virtual bridge on the device to allocate private IP addresses to VMs, enabling:</p>
                                        <ul style="margin: 0 0 8px 0; padding-left: 20px;">
                                            <li>Traffic between VMs on the same device being<strong>NETWORK ISOLATED</strong></li>
                                            <li>Proxy IP features from "Node Management" remain accessible</li>
                                            <li>Does not affect VM access to the external internet</li>
                                        </ul>
                                        <div style="background: #f0f9ff; padding: 8px 12px; border-radius: 4px; border-left: 3px solid #409EFF;">
                                            💡 <strong>适用场景：</strong>Needing distinct VIPs per VM while preventing VMs from accessing each other
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="private-nic-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 Operation Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;"><strong>Select Device:</strong>在左侧列表中点击目标设备</li>
                                            <li style="margin-bottom: 8px;"><strong>Create NIC:</strong>Click "Create NIC" and provide the following:
                                                <div style="background: #f5f7fa; padding: 8px 12px; border-radius: 4px; margin-top: 6px;">
                                                    <div style="margin-bottom: 4px;">• <strong>Custom Name:</strong>The prefix is permanently set to <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">mytBridge_</code>, only fill in the suffix portion</div>
                                                    <div>• <strong>CIDR：</strong>The subnet CIDR block, e.g., <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">172.20.0.0/16</code>, strictly defines the IP boundaries</div>
                                                </div>
                                            </li>
                                            <li style="margin-bottom: 8px;"><strong>Edit/Delete:</strong>In the list, you can edit or delete existing private NICs</li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="private-nic-cidr" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📝 CIDR Format Guildelines</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">CIDR（无类别域间路由）The format is <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">IP地址/前缀长度</code>，For example:</p>
                                        <div style="background: #f5f5f5; padding: 8px 12px; border-radius: 4px; font-family: monospace; margin-bottom: 8px;">
                                            172.20.0.0/16 → Allocatable IP ranges:172.20.0.1 ~ 172.20.255.254<br>
                                            192.168.100.0/24 → Allocatable IP ranges:192.168.100.1 ~ 192.168.100.254
                                        </div>
                                        <div style="background: #fff7e6; padding: 8px 12px; border-radius: 4px; border-left: 3px solid #E6A23C;">
                                            ⚠️ <strong>注意：</strong>Please avoid IPs overlapping with the device LAN. Recommended blocks: <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">172.16.0.0/12</code> 或 <code style="background:#e8e8e8;padding:1px 4px;border-radius:3px;">10.0.0.0/8</code> range
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                        <!-- ⑤ 公有网卡说明 -->
                        <div style="padding: 14px 18px; background: #f0f9ff; border-left: 4px solid #67C23A; border-radius: 4px;">
                            <div style="font-weight: bold; color: #67C23A; font-size: 14px; margin-bottom: 12px;">
                                🌍 Public NIC (MacVlan) Guide
                            </div>
                            <el-collapse v-model="activeHelpSections" style="border: none; background: transparent;">
                                <el-collapse-item name="public-nic-overview" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">💡 Feature Overview</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <p style="margin: 0 0 8px 0;">In Public NIC (MacVlan) mode, VMs utilize the gateway and subnet of the host device explicitly, positioning themselves on the EXACT local network.</p>
                                        <div style="margin-bottom: 10px;">
                                            <div style="font-weight: bold; color: #67C23A; margin-bottom: 4px;">✅ Advantages:</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>VM acquires a real IP on the LAN, facilitating easy local access</li>
                                                <li>Provides ultra-low network latencies akin to bare-metal linkage</li>
                                            </ul>
                                        </div>
                                        <div>
                                            <div style="font-weight: bold; color: #F56C6C; margin-bottom: 4px;">❌ Restrictions:</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Between VMs<strong>Network Isolation is IMPOSSIBLE</strong></li>
                                                <li>Once configured,<strong>Node Management IP Proxies become completely UNAVAILABLE</strong></li>
                                                <li>Mutually exclusive against both "Node Management" and "Node Allocation"</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="public-nic-steps" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">📋 Operation Steps</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <ol style="margin: 0; padding-left: 20px;">
                                            <li style="margin-bottom: 8px;"><strong>Select Device:</strong>Click the target device from the left device list</li>
                                            <li style="margin-bottom: 8px;"><strong>View NIC:</strong>The right panel illustrates current hardware network metadata (Gateways, Subnets)</li>
                                            <li style="margin-bottom: 8px;"><strong>Set MacVlan IP:</strong>Assign a sibling IP matching the host IP block. Afterwards, hit "Set MacVlan IP"</li>
                                            <li style="margin-bottom: 8px;"><strong>Update MacVlan Config:</strong>When the physical machine migrates subnets, hit "Update MacVlan Config" to dynamically bridge the subnets
                                                <div style="background: #fef0f0; padding: 8px 12px; border-radius: 4px; margin-top: 6px; border-left: 3px solid #F56C6C;">
                                                    ⚠️ This operation will<strong>AUTOMATICALLY SHUTDOWN ANY RUNNING VIRUTAL MACHINES OR CONTAINERS</strong>. Subsequently, it requires manual reallocation of IPs on each guest context
                                                </div>
                                            </li>
                                        </ol>
                                    </div>
                                </el-collapse-item>

                                <el-collapse-item name="public-nic-notice" style="margin-bottom: 8px;">
                                    <template #title>
                                        <span style="font-weight: 600; color: #303133; font-size: 13px;">⚠️ Important Notices</span>
                                    </template>
                                    <div style="font-size: 13px; line-height: 1.8; color: #606266; padding: 8px 12px;">
                                        <div style="background: #fff7e6; padding: 12px; border-radius: 4px; border-left: 4px solid #E6A23C; margin-bottom: 10px;">
                                            <div style="font-weight: bold; color: #E6A23C; margin-bottom: 6px;">🚨 Mutual Exclusion Warning</div>
                                            <p style="margin: 0;">Upon engaging MacVlan (Public NIC),<strong>Node allocation capabilities shall irrevocably be disabled</strong>. To switch modes, unconditionally flush any MacVlan setup prior to using proxies.</p>
                                        </div>
                                        <div style="background: #f0f9ff; padding: 12px; border-radius: 4px; border-left: 4px solid #409EFF;">
                                            <div style="font-weight: bold; color: #409EFF; margin-bottom: 6px;">💡 IP Allocation Suggestions</div>
                                            <ul style="margin: 0; padding-left: 20px;">
                                                <li>Static IPs distributed amongst the guests must exist within the gateway's broadcast boundaries, circumventing conflicts.</li>
                                                <li>Allocate tail-end IPs consecutively to preclude standard DHCP assignment pools</li>
                                                <li>Mandatory syncing of MacVlan configs is essential following router network topology substitutions</li>
                                            </ul>
                                        </div>
                                    </div>
                                </el-collapse-item>
                            </el-collapse>
                        </div>

                    </div>
                </el-tab-pane>
            </el-tabs>
        </div>

        <el-dialog v-model="dialogVisible" :title="$t('network.addGroup')" width="40%" :close-on-click-modal="false">
            <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
                <el-form-item label="类型" prop="type">
                    <el-radio-group v-model="formData.type">
                        <el-radio :label="1">订阅地址</el-radio>
                        <el-radio :label="2">代理协议</el-radio>
                        <!-- <el-radio :label="3">socks5节点</el-radio> -->
                    </el-radio-group>
                </el-form-item>
                <el-form-item label="分组别名" prop="alias">
                    <el-input v-model="formData.alias" placeholder="请输入分组名称" clearable />
                </el-form-item>
                <el-form-item label="URL地址" prop="url" v-if="formData.type == 1">
                    <el-input type="textarea" autosize v-model="formData.url" placeholder="请输入URL地址" clearable />
                </el-form-item>


                <el-form-item label="配置类型" prop="protocol" v-if="formData.type == 2">
                    <el-select v-model="formData.protocol" placeholder="请选择配置类型" @change="updateProtocolLabel">
                        <el-option v-for="option in protocolOptions" :key="option.value" :label="option.label"
                            :value="option.value" />
                    </el-select>
                </el-form-item>

                <el-form-item :label="selectedProtocolLabel" v-if="formData.protocol !== null">
                    <el-input type="textarea" autosize v-model="formData.addresses" placeholder="支持批量填入，批量填入请用中英文逗号分隔"
                        clearable />
                    <div class="protocol-tips" v-if="formData.type == 2">
                        <el-tooltip placement="right">
                            <template #content>
                                <div class="protocol-list">
                                    <div class="protocol-title">支持的协议格式：</div>
                                    <div class="protocol-item">vmess格式: base64(json) - 标准格式</div>
                                    <div class="protocol-item">vless格式: uuid@server:port?type=ws&security=tls&sni=xxx#remarks</div>
                                    <div class="protocol-item">ss格式: base64(method:password)@server:port#remarks (SIP002)</div>
                                    <div class="protocol-item">trojan格式: password@server:port?type=tcp&security=tls&sni=xxx#remarks</div>
                                    <div class="protocol-item">socks格式: server/port/user/password/remarks/   （用户名和密码可不填）</div>
                                    <div class="protocol-item">http格式: server/port/user/password/remarks/  （用户名和密码可不填）</div>
                                    <div class="protocol-item">wireguard格式: privateKey@server:port?publickey=xxx&address=xxx&mtu=xxx#remarks</div>
                                    <div class="protocol-item">hysteria2格式: auth@server:port?sni=xxx&insecure=1&obfs-password=xxx#remarks</div>
                                </div>
                            </template>
                            <el-button type="primary" link class="protocol-btn">
                                <el-icon>
                                    <QuestionFilled />
                                </el-icon>
                                <span style="margin-left: 4px;">支持的协议格式</span>
                            </el-button>
                        </el-tooltip>
                    </div>
                </el-form-item>

                <!-- <template v-if="formData.type === 3">
                    <el-form-item label="节点别名" prop="remarks">
                        <el-input v-model="formData.remarks" placeholder="请输入节点别名" clearable />
                    </el-form-item>
                    <el-form-item label="S5 IP" prop="socksIp">
                        <el-input v-model="formData.socksIp" placeholder="请输入s5ip" clearable />
                    </el-form-item>
                    <el-form-item label="S5 端口" prop="socksPort">
                        <el-input v-model="formData.socksPort" placeholder="请输入s5端口" clearable />
                    </el-form-item>
                    <el-form-item label="S5 用户名" prop="socksUser">
                        <el-input v-model="formData.socksUser" placeholder="请输入s5用户名" clearable />
                    </el-form-item>
                    <el-form-item label="S5 密码" prop="socksPassword">
                        <el-input v-model="formData.socksPassword" type="password" placeholder="请输入s5密码" clearable
                            show-password />
                    </el-form-item>
                </template> -->
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="editDialogVisible" title="编辑分组名称" width="40%" :close-on-click-modal="false">
            <el-form :model="editFormData" ref="editFormRef" label-width="100px">
                <!-- <el-form-item label="分组ID">
                    <el-input v-model="editFormData.id" disabled />
                </el-form-item> -->
                <el-form-item label="分组名称" prop="alias">
                    <el-input v-model="editFormData.alias" placeholder="请输入分组名称" clearable />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="editDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="handleEditSubmit" :loading="editLoading">确定</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="vpcDialogVisible" :title="$t('network.assignVpc')" width="60%" :close-on-click-modal="false">
            <div v-loading="vpcDialogLoading" class="vpc-dialog-content">
                <el-steps :active="stepActive" finish-status="success" style="margin-bottom: 20px; flex-shrink: 0;">
                    <el-step title="选择云机" />
                    <el-step title="选择分组" />
                    <el-step title="选择节点" />
                </el-steps>

                <div v-if="stepActive === 0" class="step-content">
                    <div style="margin-bottom: 15px; display: flex; gap: 10px;">
                        <el-input v-model="vpcContainerSearch" placeholder="按云机名称搜索" clearable style="width: 250px;" />
                        <el-input v-model="vpcSlotSearch" placeholder="按坑位号搜索" clearable style="width: 250px;" />
                    </div>

                    <el-table :data="filteredVpcContainerList" style="width: 100%" stripe height="100%" @selection-change="handleVpcContainerSelectionChange">
                        <el-table-column type="selection" width="55" align="center" />
                        <el-table-column prop="name" :label="$t('network.machineName')" align="center" min-width="150">
                            <template #default="{ row }">
                                {{ extractNodeNumber(row.name) }}
                            </template>
                        </el-table-column>
                        <el-table-column prop="indexNum" label="云机坑位" align="center" min-width="120" />
                        <el-table-column prop="dns" label="DNS" align="center" min-width="150" />
                        <el-table-column prop="status" label="状态" align="center" width="120">
                            <template #default="{ row }">
                                <el-tag :type="row.status === 'running' ? 'success' : 'info'">
                                    {{ row.status === 'running' ? '运行中' : '已停止' }}
                                </el-tag>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>

                <div v-if="stepActive === 1" class="step-content">
                    <el-button @click="stepActive = 0" style="margin-bottom: 10px;">← 返回选择云机</el-button>

                    <el-radio-group v-model="vpcSelectMode" style="margin-bottom: 15px;">
                        <el-radio label="specified">指定分组节点</el-radio>
                        <el-radio label="random">随机分组节点</el-radio>
                    </el-radio-group>

                    <el-table :data="vpcGroupList" style="width: 100%" stripe height="100%">
                        <el-table-column prop="alias" label="分组名称" align="center" min-width="180" />
                        <el-table-column :label="$t('common.operation')" width="120" align="center">
                            <template #default="{ row }">
                                <el-radio v-model="vpcSelectedGroupId" :label="row.id">
                                    <span></span>
                                </el-radio>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>

                <div v-if="stepActive === 2" class="step-content">
                    <el-button @click="stepActive = 1" style="margin-bottom: 10px;">← 返回选择分组</el-button>
                    <el-table :data="vpcNodeList" style="width: 100%" stripe height="100%">
                        <el-table-column prop="remarks" :label="$t('network.nodeName')" align="center" min-width="180" />
                        <el-table-column :label="$t('common.operation')" width="100" align="center">
                            <template #default="{ row }">
                                <el-radio v-if="vpcSelectMode === 'specified'" v-model="vpcSelectedNodeId"
                                    :label="row.id">
                                    <span></span>
                                </el-radio>
                                <span v-else style="color: #909399; font-size: 12px;">随机模式</span>
                            </template>
                        </el-table-column>
                    </el-table>
                </div>

                <div style="margin-top: 20px; text-align: right; flex-shrink: 0;">
                    <el-button v-if="stepActive > 0" @click="vpcDialogVisible = false">取消</el-button>
                    <el-button v-if="stepActive === 0" type="primary" @click="nextVpcStep"
                        :disabled="!canNextVpcStep">下一步</el-button>
                    <el-button v-if="stepActive === 1 && vpcSelectMode === 'specified'" type="primary"
                        @click="nextVpcStep" :disabled="!canNextVpcStep">下一步</el-button>
                    <el-button v-if="stepActive === 1 && vpcSelectMode === 'random'" type="primary"
                        @click="confirmSetVpc" :loading="vpcDialogLoading" :disabled="!canNextVpcStep">确定</el-button>
                    <el-button v-if="stepActive === 2" type="primary" @click="confirmSetVpc" :loading="vpcDialogLoading"
                        :disabled="!vpcSelectedNodeId">确定</el-button>
                </div>
            </div>
        </el-dialog>

        <el-dialog v-model="createPrivateNicDialogVisible" title="创建私有网卡" width="34%" :close-on-click-modal="false">
            <el-form :model="createPrivateNicForm" :rules="createPrivateNicRules" ref="createPrivateNicFormRef" label-width="100px">
                <el-form-item label="自定义名称" prop="customName">
                    <el-input v-model="createPrivateNicForm.customName" placeholder="请输入自定义名称" clearable>
                        <template #prepend>mytBridge_</template>
                    </el-input>
                </el-form-item>
                <el-form-item label="CIDR" prop="cidr">
                    <el-input v-model="createPrivateNicForm.cidr" placeholder="例如: 192.168.88.1" clearable />
                    <p style="font-size: 12px; color: red;">注意：禁止设置与路由器相同网段，否则出现无法恢复问题</p>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="createPrivateNicDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="submitCreatePrivateNic" :loading="createPrivateNicLoading">确定</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="editPrivateNicDialogVisible" title="编辑私有网卡" width="30%" :close-on-click-modal="false">
            <el-form :model="editPrivateNicForm" :rules="editPrivateNicRules" ref="editPrivateNicFormRef" label-width="100px">
                <el-form-item label="网卡名称">
                    <el-input v-model="editPrivateNicForm.name" disabled></el-input>
                </el-form-item>
                <el-form-item label="CIDR" prop="cidr">
                    <el-input v-model="editPrivateNicForm.cidr" placeholder="例如: 192.168.88.1" clearable />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="editPrivateNicDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="submitEditPrivateNic" :loading="editPrivateNicLoading">确定</el-button>
                </span>
            </template>
        </el-dialog>

        <!-- 编辑节点对话框 -->
        <el-dialog v-model="editNodeDialogVisible" :title="$t('network.editNode')" width="500px">
            <el-form :model="editNodeForm" label-width="80px">
                <el-form-item :label="$t('network.alias')">
                    <el-input v-model="editNodeForm.remarks" />
                </el-form-item>
                <el-form-item :label="$t('network.address')">
                    <el-input v-model="editNodeForm.server" />
                </el-form-item>
                <el-form-item :label="$t('network.port')">
                    <el-input v-model="editNodeForm.serverPort" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="editNodeDialogVisible = false">{{ $t('common.cancel') }}</el-button>
                <el-button type="primary" @click="submitEditNode" :loading="editNodeLoading">{{ $t('common.confirm') }}</el-button>
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


import { ref, computed, reactive, watch, getCurrentInstance } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Delete, QuestionFilled, Edit, RefreshRight, Refresh } from '@element-plus/icons-vue'

const props = defineProps({
    devices: {
        type: Array,
        default: () => []
    },
    token: {
        type: String,
        default: null
    },
    deviceFirmwareInfo: {
        type: Map,
        default: () => new Map()
    },
    deviceVersionInfo: {
        type: Map,
        default: () => new Map()
    },
    devicesStatusCache: {
        type: Map,
        default: () => new Map()
    }
})

const { proxy } = getCurrentInstance()
const t = proxy.$t
const activeTab = ref('group')
const selectedDeviceIP = ref('')
const selectedDeviceVersion = ref('v3')
const groupList = ref([])
const vpcList = ref([])
const loading = ref(false)
const searchName = ref('')
const selectedGroupId = ref('')
const selectedVpcNodes = ref([])
const containerRuleList = ref([])
const selectedContainerRules = ref([])
const containerRuleLoading = ref(false)
const sdkVersionCache = new Map()
const localGroupFilter = ref('全部')

// 功能说明面板折叠状态
const activeHelpSections = ref(['overview'])

// 域名过滤
const domainFilterContainerID = ref('')
const domainFilterList = ref([])
const domainFilterLoading = ref(false)
const domainFilterVpcContainers = ref([])
const domainFilterVpcLoading = ref(false)

// 容器域名过滤弹窗
const showContainerDomainDialog = ref(false)
const containerDomainSubmitLoading = ref(false)
const containerDomainForm = reactive({
    containerID: '',
    rules: [{ prefix: 'domain:', value: '' }]
})

// 全局域名过滤弹窗
const showGlobalDomainDialog = ref(false)
const globalDomainSubmitLoading = ref(false)

// 全局域名过滤查询
const globalDomainFilterList = ref([])
const globalDomainFilterLoading = ref(false)
const globalDomainFilterVisible = ref(false)

// 统一查询模式：'container' | 'global' | ''
const domainQueryMode = ref('')
const unifiedDomainFilterList = computed(() =>
    domainQueryMode.value === 'global' ? globalDomainFilterList.value : domainFilterList.value
)

const onContainerDomainFilterChange = async (val) => {
    if (!val) return
    domainQueryMode.value = 'container'
    await fetchDomainFilterList()
}
const onContainerDomainFilterClear = () => {
    domainQueryMode.value = ''
    domainFilterList.value = []
}
const globalDomainForm = reactive({
    rules: [{ prefix: 'domain:', value: '' }]
})

// 兼容旧变量（fetchDomainFilterList 用到）
const showDomainFilterDialog = ref(false)
const domainFilterDialogInput = ref('')
const domainFilterSubmitLoading = ref(false)
const domainFilterDialogDomains = computed(() =>
    domainFilterDialogInput.value
        .split('\n')
        .map(s => s.trim())
        .filter(s => s.length > 0)
)

// 域名规则辅助函数
const DOMAIN_PREFIXES = ['domain:', 'full:', 'keyword:', 'regexp:']
const domainRulePrefix = (domain) => {
    if (!domain) return 'domain:'
    const hit = DOMAIN_PREFIXES.find(p => domain.startsWith(p))
    return hit || 'domain:'
}
const domainRuleValue = (domain) => {
    if (!domain) return ''
    const hit = DOMAIN_PREFIXES.find(p => domain.startsWith(p))
    return hit ? domain.slice(hit.length) : domain
}
const domainRulePlaceholder = (prefix) => {
    if (prefix === 'full:') return '例如：www.example.com'
    if (prefix === 'keyword:') return '例如：ads'
    if (prefix === 'regexp:') return '例如：^ad\\d+\\..*'
    return '例如：example.com'
}

const openContainerDomainFilterDialog = () => {
    containerDomainForm.containerID = ''
    containerDomainForm.rules = [{ prefix: 'domain:', value: '' }]
    showContainerDomainDialog.value = true
}

const clearContainerDomainFilter = async () => {
    if (!domainFilterContainerID.value) {
        ElMessage.warning('请先选择容器')
        return
    }
    try {
        await ElMessageBox.confirm(
            `确定清除容器 "${domainFilterContainerID.value}" 的所有域名过滤规则吗？`,
            '确认清除',
            { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
    } catch {
        return
    }
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainFilter?containerID=${encodeURIComponent(domainFilterContainerID.value.trim())}`,
            {
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('清除成功')
                domainFilterList.value = []
            } else {
                ElMessage.error(data.message || '清除失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('清除域名过滤失败:', error)
        ElMessage.error('清除失败，请检查网络连接')
    }
}

const clearGlobalDomainFilter = async () => {
    try {
        await ElMessageBox.confirm(
            '确定清除全局域名过滤规则吗？',
            '确认清除',
            { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
    } catch {
        return
    }
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainFilter/global`,
            {
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('清除成功')
                globalDomainFilterList.value = []
            } else {
                ElMessage.error(data.message || '清除失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('清除全局域名过滤失败:', error)
        ElMessage.error('清除失败，请检查网络连接')
    }
}
const resetContainerDomainDialog = () => {
    containerDomainForm.containerID = ''
    containerDomainForm.rules = [{ prefix: 'domain:', value: '' }]
}
const openGlobalDomainFilterDialog = () => {
    globalDomainForm.rules = [{ prefix: 'domain:', value: '' }]
    showGlobalDomainDialog.value = true
}
const resetGlobalDomainDialog = () => {
    globalDomainForm.rules = [{ prefix: 'domain:', value: '' }]
}

// 域名直连
const domainDirectContainerID = ref('')
const domainDirectList = ref([])
const domainDirectLoading = ref(false)
const domainDirectVpcContainers = ref([])
const domainDirectVpcLoading = ref(false)

// 域名直连弹窗
const showDomainDirectDialog = ref(false)
const domainDirectSubmitLoading = ref(false)
const domainDirectForm = reactive({
    containerID: '',
    rules: [{ prefix: 'domain:', value: '' }]
})

const onDomainDirectContainerChange = async (val) => {
    if (!val) return
    await fetchDomainDirectList()
}
const onDomainDirectContainerClear = () => {
    domainDirectList.value = []
}

const openDomainDirectDialog = () => {
    domainDirectForm.containerID = ''
    domainDirectForm.rules = [{ prefix: 'domain:', value: '' }]
    showDomainDirectDialog.value = true
}
const resetDomainDirectDialog = () => {
    domainDirectForm.containerID = ''
    domainDirectForm.rules = [{ prefix: 'domain:', value: '' }]
}

const clearDomainDirect = async () => {
    if (!domainDirectContainerID.value) {
        ElMessage.warning('请先选择容器')
        return
    }
    try {
        await ElMessageBox.confirm(
            `确定清除容器 "${domainDirectContainerID.value}" 的所有域名直连规则吗？`,
            '确认清除',
            { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
    } catch {
        return
    }
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainDirect?containerID=${encodeURIComponent(domainDirectContainerID.value.trim())}`,
            {
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('清除成功')
                domainDirectList.value = []
            } else {
                ElMessage.error(data.message || '清除失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('清除域名直连失败:', error)
        ElMessage.error('清除失败，请检查网络连接')
    }
}

// Private NIC variables
const privateNicList = ref([])
const privateNicLoading = ref(false)
const createPrivateNicDialogVisible = ref(false)
const createPrivateNicLoading = ref(false)
const createPrivateNicFormRef = ref(null)
const createPrivateNicForm = reactive({
    customName: '',
    cidr: ''
})
const createPrivateNicRules = {
    customName: [
        { required: true, message: '请输入自定义名称', trigger: 'blur' },
        { max: 4, message: '长度不能超过4个字符', trigger: 'blur' },
        { pattern: /^[a-zA-Z0-9]+$/, message: '只允许输入数字或英文', trigger: 'blur' }
    ],
    cidr: [
        { required: true, message: '请输入CIDR', trigger: 'blur' }
    ]
}

const editPrivateNicDialogVisible = ref(false)
const editPrivateNicLoading = ref(false)
const editPrivateNicFormRef = ref(null)
const editPrivateNicForm = reactive({
    name: '',
    cidr: ''
})
const editPrivateNicRules = {
    cidr: [
        { required: true, message: '请输入CIDR', trigger: 'blur' }
    ]
}

// Public NIC variables
const publicNicList = ref([])
const publicNicLoading = ref(false)
const hasMacVlan = computed(() => publicNicList.value.some(nic => nic.name === 'MacVlan'))




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

const groupMap = computed(() => {
    const map = {}
    groupList.value.forEach(group => {
        map[group.id] = group.alias
    })
    return map
})

const filteredVpcList = computed(() => {
    if (!selectedGroupId.value) {
        return []  // 未选择分组时不显示数据
    }
    return vpcList.value.filter(vpc => vpc.groupId === selectedGroupId.value)
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

const getDeviceSDKVersion = async (deviceIP, deviceVersion) => {
    const cacheKey = `${deviceIP}`
    if (sdkVersionCache.has(cacheKey)) {
        return sdkVersionCache.get(cacheKey)
    }

    let sdkVersion = '0'

    const deviceId = props.devices.find(d => d.ip === deviceIP)?.id

    const versionInfo = props.deviceVersionInfo.get(deviceId)
    if (versionInfo?.currentVersion) {
        sdkVersion = String(versionInfo.currentVersion)
        sdkVersionCache.set(cacheKey, sdkVersion)
        return sdkVersion
    }

    if (deviceVersion === 'v3') {
        try {
            const savedPassword = localStorage.getItem(`device_password_${deviceIP}`)
            let headers = {}
            if (savedPassword) {
                const auth = btoa(`admin:${savedPassword}`)
                headers = {
                    'Authorization': `Basic ${auth}`
                }
            }

            const controller = new AbortController()
            const timeoutId = setTimeout(() => controller.abort(), 5000)

            const response = await fetch(
                `http://${getDeviceAddr(deviceIP)}/api/v1/device/version`,
                {
                    signal: controller.signal,
                    headers: headers
                }
            )
            clearTimeout(timeoutId)

            if (response.ok) {
                const data = await response.json()
                if (data.code === 0 && data.data) {
                    sdkVersion = String(data.data.currentVersion || data.data.version || '0')
                    sdkVersionCache.set(cacheKey, sdkVersion)
                }
            }
        } catch (error) {
            if (error.name !== 'AbortError') {
                console.error('获取设备API版本失败:', error)
            }
        }
    }

    return sdkVersion
}

const checkAndWarnSDKVersion = async (deviceIP, deviceVersion) => {
    const sdkVersion = await getDeviceSDKVersion(deviceIP, deviceVersion)
    const versionNum = parseInt(sdkVersion, 10) || 0

    if (versionNum < 41) {
        ElMessage.warning({
            message: `请升级SDK，当前SDK版本: ${sdkVersion}，网络管理要求SDK版本>=41`,
            duration: 0,
            showClose: true,
            type: 'warning'
        })
        return false
    }
    return true
}

const dialogVisible = ref(false)
const formRef = ref(null)
const submitLoading = ref(false)
const editDialogVisible = ref(false)
const editFormRef = ref(null)
const editFormData = reactive({
    id: '',
    alias: ''
})
const editLoading = ref(false)
const vpcDialogVisible = ref(false)
const vpcDialogLoading = ref(false)
const selectedContainer = ref(null)
const selectedVpcNode = ref('')
const vpcContainerList = ref([])
const vpcContainerSearch = ref('')
const vpcSlotSearch = ref('')
const vpcSelectedContainers = ref([])
const vpcGroupList = ref([])
const vpcSelectedGroupId = ref('')
const vpcNodeList = ref([])
const vpcSelectedNodeId = ref('')
const vpcSelectMode = ref('specified')
const stepActive = ref(0)

const canNextVpcStep = computed(() => {
    if (stepActive.value === 0) return vpcSelectedContainers.value.length > 0
    if (stepActive.value === 1) return !!vpcSelectedGroupId.value
    if (stepActive.value === 2) return !!vpcSelectedNodeId.value
    return false
})

const filteredVpcContainerList = computed(() => {
    let list = vpcContainerList.value
    // 过滤掉已分配节点的云机
    if (containerRuleList.value.length > 0) {
        const assignedNames = new Set(containerRuleList.value.map(r => r.containerName))
        list = list.filter(c => !assignedNames.has(c.name))
    }
    if (vpcContainerSearch.value) {
        const search = vpcContainerSearch.value.toLowerCase()
        list = list.filter(c => c.name && c.name.toLowerCase().includes(search))
    }
    if (vpcSlotSearch.value) {
        const slotSearch = vpcSlotSearch.value.trim()
        list = list.filter(c => c.indexNum !== undefined && String(c.indexNum).includes(slotSearch))
    }
    return list
})

const extractNodeNumber = (remarks) => {
    if (!remarks) return ''
    const parts = remarks.split('_')
    return parts.length > 0 ? parts[parts.length - 1] : remarks
}

const formData = reactive({
    alias: '',
    url: '',
    type: 1,
    protocol: null,
    addresses: []
    // remarks: '',
    // socksIp: '',
    // socksPort: '',
    // socksUser: '',
    // socksPassword: ''
})

// 监听类型切换，清空之前的内容
watch(() => formData.type, (newType, oldType) => {
    if (newType !== oldType) {
        formData.url = ''
        formData.addresses = []
        formData.protocol = null
        selectedProtocolLabel.value = ''
    }
})

const protocolOptions = [
    { label: 'vmess', value: '1' },
    { label: 'vless', value: '2' },
    { label: 'ss', value: '3' },
    { label: 'trojan', value: '4' },
    { label: 'socks', value: '5' },
    { label: 'http', value: '6' },
    { label: 'wireguard', value: '7' },
    { label: 'hysteria2', value: '8' }
]

const selectedProtocolLabel = ref('') // 当前选中的协议label

const formRules = {
    // type: [
    //     { required: true, message: '请选择类型', trigger: 'change' }
    // ],
    alias: [
        { required: true, message: '请输入分组名称', trigger: 'blur' },
        { min: 2, max: 50, message: '分组名称长度为2-50个字符', trigger: 'blur' }
    ],
    url: [
        { required: true, message: '请输入URL地址', trigger: 'blur' }
    ],
    protocol: [
        { required: true, message: '请选择配置类型', trigger: 'change' }
    ]
    // remarks: [
    //     { required: true, message: '请输入节点别名', trigger: 'blur' },
    //     { min: 2, max: 50, message: '节点别名长度为2-50个字符', trigger: 'blur' }
    // ],
    // socksIp: [
    //     { required: true, message: '请输入s5ip', trigger: 'blur' }
    // ],
    // socksPort: [
    //     { required: true, message: '请输入s5端口', trigger: 'blur' }
    // ],
    // socksUser: [
    //     { required: false, message: '请输入s5用户名', trigger: 'blur' }
    // ],
    // socksPassword: [
    //     { required: false, message: '请输入s5密码', trigger: 'blur' }
    // ]
}

const handleEditGroupAlias = (groupId) => {
    const group = groupList.value.find(g => g.id === groupId)
    if (group) {
        editFormData.id = group.id
        editFormData.alias = group.alias
        editDialogVisible.value = true
    }
}


const handleUpdateGroup = async (groupId) => {
    try {
        await ElMessageBox.confirm(
            '更新分组会重新从订阅源拉取节点，您手动编辑的节点修改（别名、地址、端口等）将被覆盖，可能导致网络失效。确定要继续吗？',
            '更新分组确认',
            {
                confirmButtonText: '确定更新',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )
    } catch {
        return
    }

    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/group/update`,
            {
                method: 'POST',
                headers: getAuthHeaders(selectedDeviceIP.value),
                body: JSON.stringify({
                    id: groupId,
                })
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('更新成功')
                fetchGroupList(false)
            } else {
                ElMessage.error(data.message || '更新失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('更新失败:', error)
        ElMessage.error('更新失败，请检查网络连接')
    }
}

const handleEditSubmit = async () => {
    if (!editFormData.alias) {
        ElMessage.warning('请填写分组名称')
        return
    }

    editLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/group/alias`,
            {
                method: 'POST',
                headers: {
                    // 'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({
                    id: editFormData.id,
                    newAlias: editFormData.alias
                })
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('修改分组名称成功')
                editDialogVisible.value = false
                editFormData.id = ''
                editFormData.alias = ''
                fetchGroupList()
            } else {
                ElMessage.error(data.message || '修改失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('修改分组名称失败:', error)
        ElMessage.error('修改分组名称失败，请检查网络连接')
    } finally {
        editLoading.value = false
    }
}

const handleAddGroup = () => {
    formData.alias = ''
    formData.url = ''
    formData.type = 1
    formData.remarks = ''
    formData.socksIp = ''
    formData.socksPort = ''
    formData.socksUser = ''
    formData.socksPassword = ''
    dialogVisible.value = true
    if (formRef.value) {
        formRef.value.resetFields()
    }
}

const handleDeleteGroup = async (groupId = null) => {
    const targetId = groupId || selectedGroupId.value
    if (!targetId) return

    const groupName = groupMap.value[targetId]

    try {
        await ElMessageBox.confirm(
            `确定要 ${t('network.deleteGroup')} "${groupName}" 吗？删除后该分组下的所有节点也将被删除。`,
            '删除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/group/?id=${targetId}`,
            {
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success(`${t('network.deleteGroup')}成功`)
                selectedGroupId.value = ''
                fetchGroupList()
            } else {
                ElMessage.error(data.message || `${t('network.deleteGroup')}失败`)
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error(`${t('network.deleteGroup')}失败:`, error)
            ElMessage.error(`${t('network.deleteGroup')}失败，请检查网络连接`)
        }
    }
}

// 编辑节点
const editNodeDialogVisible = ref(false)
const editNodeLoading = ref(false)
const editNodeForm = ref({
    vpcID: '',
    remarks: '',
    server: '',
    serverPort: '',
    profile: null
})

const openEditNodeDialog = (row) => {
    const profile = JSON.parse(row.profile)
    editNodeForm.value = {
        vpcID: row.id,
        remarks: row.remarks || '',
        server: profile.server || '',
        serverPort: profile.serverPort || '',
        profile: profile
    }
    editNodeDialogVisible.value = true
}

const submitEditNode = async () => {
    editNodeLoading.value = true
    try {
        const updatedProfile = { ...editNodeForm.value.profile }
        updatedProfile.server = editNodeForm.value.server
        updatedProfile.serverPort = editNodeForm.value.serverPort
        updatedProfile.remarks = editNodeForm.value.remarks

        const body = {
            vpcID: editNodeForm.value.vpcID,
            remarks: editNodeForm.value.remarks,
            profile: updatedProfile
        }

        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc`,
            {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify(body)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('编辑成功')
                editNodeDialogVisible.value = false
                fetchGroupList(false)
            } else {
                ElMessage.error(data.message || '编辑失败')
            }
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('编辑节点失败:', error)
        ElMessage.error('编辑失败，请检查网络连接')
    } finally {
        editNodeLoading.value = false
    }
}

const deleteContainerRule = async (row) => {
    console.log('deleteContainerRule', row)
    try {
        await ElMessageBox.confirm(
            `确定要删除内分组"${row.containerName}"节点吗？`,
            '删除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc?vpcID=${row.id}`,
            {
                method: 'DELETE',
                headers: {
                    // 'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                }
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('删除成功')
                // fetchContainerRule()
                fetchGroupList(false)
            } else {
                ElMessage.error(data.message || '删除失败')
            }
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error('删除失败:', error)
            ElMessage.error('删除失败，请检查网络连接')
        }
    }
}


const testSpeed = async (row) => {
    console.log('testSpeed', row)
    const profile = JSON.parse(row.profile)
    row.latency = '测速中...'

    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/test?address=${profile.server}:${profile.serverPort}`,
            {
                method: 'GET',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                if (data.data.msg !== '') {
                    row.latency = '-1ms'
                    ElMessage.error(data.data.msg)
                } else {
                    row.latency = data.data.latency
                    ElMessage.success('测速成功')
                }
            } else {
                row.latency = '-1ms'
                ElMessage.error(data.message || '测速失败')
            }
        } else if (response.status === 404) {
            row.latency = '-1ms'
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            row.latency = '-1ms'
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        row.latency = '-1ms'
        if (error !== 'cancel') {
            console.error('测速失败:', error)
            ElMessage.error('测速失败，请检查网络连接')
        }
    }
}


const handleSubmit = async () => {
    if (!formRef.value) return

    await formRef.value.validate(async (valid) => {
        if (valid) {
            // 处理地址输入
            let addresses = []
            let addressText = ''

            // type=1（URL订阅）使用 formData.url，type=2（单地址）使用 formData.addresses
            if (formData.type === 1) {
                if (!formData.url || !formData.url.trim()) {
                    ElMessage.error('请输入URL地址')
                    return
                }
                // addresses = [formData.url.trim()]
            } else {
                if (formData.addresses) {
                    if (Array.isArray(formData.addresses)) {
                        addressText = formData.addresses[0] || ''
                    } else {
                        addressText = formData.addresses || ''
                    }
                }

                if (!addressText.trim()) {
                    ElMessage.error('请输入地址信息')
                    return
                }

                addresses = addressText.split(/[，,]/).map(a => a.trim()).filter(a => a)
                if (addresses.length === 0) {
                    ElMessage.error('请输入有效的地址')
                    return
                }
            }

            // 自动添加协议前缀（除socks和http外）
            if (formData.type === 2 && formData.protocol && formData.protocol !== '5' && formData.protocol !== '6') {
                const protocolPrefixes = {
                    '1': 'vmess://',
                    '2': 'vless://',
                    '3': 'ss://',
                    '4': 'trojan://',
                    '7': 'wireguard://',
                    '8': 'hysteria2://'
                }
                const prefix = protocolPrefixes[formData.protocol]
                if (prefix) {
                    addresses = addresses.map(address => {
                        const trimmed = address.trim().toLowerCase()
                        if (!trimmed.startsWith(prefix)) {
                            return prefix + address.trim()
                        }
                        return address.trim()
                    })
                }
            }

            // 仅对socks协议进行格式验证
            if (formData.type === 2 && (formData.protocol === '5' || formData.protocol === '6')) {
                const validation = validateBatchAddresses(addresses, formData.protocol)
                if (!validation.valid) {
                    ElMessage.error(validation.message)
                    return
                }

                const protocolPrefix = formData.protocol === '5' ? 'socks' : 'http'
                
                // 将socks/http地址转换为标准格式
                addresses = addresses.map(address => {
                    const trimmedAddress = address.trim()

                    if (!trimmedAddress.endsWith('/')) {
                        return trimmedAddress
                    }

                    const parts = trimmedAddress.split('/').filter(p => p)
                    if (parts.length < 2) {
                        return trimmedAddress
                    }

                    const ip = parts[0]
                    const port = parts[1]
                    const username = parts[2] || ''
                    const password = parts[3] || ''
                    const remarks = parts[4] || `${protocolPrefix}_${Math.random().toString(36).substring(2, 8)}`

                    if (username && password) {
                        return `${protocolPrefix}://${username}:${password}@${ip}:${port}#${remarks}`
                    } else {
                        return `${protocolPrefix}://${ip}:${port}#${remarks}`
                    }
                })
            }

            submitLoading.value = true
            try {
                let submitData = {
                    alias: formData.alias,
                    url: formData.url,
                    source: formData.type,
                    addresses: addresses
                }

                let submitUrl = `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/group`

                const response = await fetch(
                    submitUrl,
                    {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            ...getAuthHeaders(selectedDeviceIP.value)
                        },
                        body: JSON.stringify(submitData)
                    }
                )

                if (response.ok) {
                    const data = await response.json()
                    if (data.code === 0) {
                        ElMessage.success(`${t('network.addGroup')}成功`)
                        dialogVisible.value = false
                        formData.alias = ''
                        formData.url = ''
                        formData.addresses = []
                        formData.protocol = null
                        selectedProtocolLabel.value = ''
                        fetchGroupList()
                    } else {
                        ElMessage.error(data.message || `${t('network.addGroup')}失败`)
                    }
                } else if (response.status === 404) {
                    await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
                } else {
                    ElMessage.error('接口请求失败')
                }
            } catch (error) {
                console.error(`${t('network.addGroup')}失败:`, error)
                ElMessage.error(`${t('network.addGroup')}失败，请检查网络连接`)
            } finally {
                submitLoading.value = false
            }
        }
    })
}

const selectDevice = async (device) => {
    selectedDeviceIP.value = device.ip
    selectedDeviceVersion.value = device.version || 'v3'
    selectedGroupId.value = ''
    await fetchGroupList()
    if (activeTab.value === 'vpc') {
        await fetchContainerRule()
    } else if (activeTab.value === 'private-nic') {
        await fetchPrivateNics()
    } else if (activeTab.value === 'public-nic') {
        await fetchPublicNics()
    } else if (activeTab.value === 'domain-filter') {
        domainFilterContainerID.value = ''
        domainFilterList.value = []
        await fetchDomainFilterVpcContainers()
    } else if (activeTab.value === 'domain-direct') {
        domainDirectContainerID.value = ''
        domainDirectList.value = []
        await fetchDomainDirectVpcContainers()
    }
}

const fetchGroupList = async (autoSelectFirst = true) => {
    if (!selectedDeviceIP.value) {
        groupList.value = []
        vpcList.value = []
        return
    }

    loading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/group?alias=${searchName.value}`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code == 0) {
                groupList.value = data.data.list || []

                if (autoSelectFirst && groupList.value.length > 0) {
                    selectedGroupId.value = groupList.value[0].id
                }

                const allVpcs = []
                groupList.value.forEach(group => {
                    if (group.vpcs && group.vpcs.list) {
                        group.vpcs.list.forEach(vpc => {
                            allVpcs.push({
                                ...vpc,
                                groupAlias: group.alias,
                                latency: null
                            })
                        })
                    }
                })
                vpcList.value = allVpcs
            } else {
                ElMessage.error(data.message || '获取分组列表失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('获取分组列表失败:', error)
        ElMessage.error('获取分组列表失败，请检查网络连接')
    } finally {
        loading.value = false
    }
}

const fetchContainerRule = async () => {
    if (!selectedDeviceIP.value) {
        containerRuleList.value = []
        return
    }

    containerRuleList.value = []
    containerRuleLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/containerRule`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                containerRuleList.value = data.data.list || []
            } else {
                ElMessage.error(data.message || '获取列表失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('获取列表失败:', error)
        ElMessage.error('获取列表失败，请检查网络连接')
    } finally {
        containerRuleLoading.value = false
    }
}

const handleTabChange = async (tab) => {
    if (tab === 'vpc' && selectedDeviceIP.value) {
        await fetchContainerRule()
    } else if (tab === 'private-nic' && selectedDeviceIP.value) {
        await fetchPrivateNics()
    } else if (tab === 'public-nic' && selectedDeviceIP.value) {
        await fetchPublicNics()
    } else if (tab === 'domain-filter') {
        domainFilterList.value = []
        domainFilterContainerID.value = ''
        if (selectedDeviceIP.value) {
            await fetchDomainFilterVpcContainers()
        }
    } else if (tab === 'domain-direct') {
        domainDirectList.value = []
        domainDirectContainerID.value = ''
        if (selectedDeviceIP.value) {
            await fetchDomainDirectVpcContainers()
        }
    }
}

const fetchDomainFilterVpcContainers = async () => {
    if (!selectedDeviceIP.value) return
    domainFilterVpcLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/containerRule`,
            { headers: getAuthHeaders(selectedDeviceIP.value) }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                domainFilterVpcContainers.value = data.data?.list || []
            } else {
                domainFilterVpcContainers.value = []
            }
        } else {
            domainFilterVpcContainers.value = []
        }
    } catch (e) {
        console.error('获取VPC容器列表失败:', e)
        domainFilterVpcContainers.value = []
    } finally {
        domainFilterVpcLoading.value = false
    }
}

const fetchDomainDirectVpcContainers = async () => {
    if (!selectedDeviceIP.value) return
    domainDirectVpcLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/containerRule`,
            { headers: getAuthHeaders(selectedDeviceIP.value) }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                domainDirectVpcContainers.value = data.data?.list || []
            } else {
                domainDirectVpcContainers.value = []
            }
        } else {
            domainDirectVpcContainers.value = []
        }
    } catch (e) {
        console.error('获取VPC容器列表失败:', e)
        domainDirectVpcContainers.value = []
    } finally {
        domainDirectVpcLoading.value = false
    }
}

const fetchDomainDirectList = async () => {
    if (!selectedDeviceIP.value || !domainDirectContainerID.value) {
        domainDirectList.value = []
        return
    }
    domainDirectLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainDirect?containerID=${encodeURIComponent(domainDirectContainerID.value)}`,
            { headers: getAuthHeaders(selectedDeviceIP.value) }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                domainDirectList.value = (data.data?.domains || []).map(d => ({ domain: d }))
            } else {
                ElMessage.error(data.message || '查询失败')
                domainDirectList.value = []
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
            domainDirectList.value = []
        } else {
            ElMessage.error('接口请求失败')
            domainDirectList.value = []
        }
    } catch (error) {
        console.error('查询域名直连列表失败:', error)
        ElMessage.error('查询失败，请检查网络连接')
        domainDirectList.value = []
    } finally {
        domainDirectLoading.value = false
    }
}

const submitDomainDirect = async () => {
    if (!domainDirectForm.containerID) {
        ElMessage.warning('请选择容器')
        return
    }
    const domains = domainDirectForm.rules
        .filter(r => r.value.trim())
        .map(r => r.prefix + r.value.trim())
    if (!domains.length) {
        ElMessage.warning('请至少填写一条域名')
        return
    }
    domainDirectSubmitLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainDirect`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({
                    containerID: domainDirectForm.containerID,
                    domains: domains
                })
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('设置成功')
                showDomainDirectDialog.value = false
                if (domainDirectContainerID.value === domainDirectForm.containerID) {
                    await fetchDomainDirectList()
                }
            } else {
                ElMessage.error(data.message || '设置失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('设置域名直连失败:', error)
        ElMessage.error('设置失败，请检查网络连接')
    } finally {
        domainDirectSubmitLoading.value = false
    }
}

const fetchDomainFilterList = async () => {
    if (!selectedDeviceIP.value || !domainFilterContainerID.value) {
        domainFilterList.value = []
        return
    }
    domainFilterLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainFilter?containerID=${encodeURIComponent(domainFilterContainerID.value)}`,
            { headers: getAuthHeaders(selectedDeviceIP.value) }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                domainFilterList.value = (data.data?.domains || []).map(d => ({ domain: d }))
            } else {
                ElMessage.error(data.message || '查询失败')
                domainFilterList.value = []
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
            domainFilterList.value = []
        } else {
            ElMessage.error('接口请求失败')
            domainFilterList.value = []
        }
    } catch (error) {
        console.error('查询域名过滤列表失败:', error)
        ElMessage.error('查询失败，请检查网络连接')
        domainFilterList.value = []
    } finally {
        domainFilterLoading.value = false
    }
}

const submitContainerDomainFilter = async () => {
    if (!containerDomainForm.containerID) {
        ElMessage.warning('请选择容器')
        return
    }
    const domains = containerDomainForm.rules
        .filter(r => r.value.trim())
        .map(r => r.prefix + r.value.trim())
    if (!domains.length) {
        ElMessage.warning('请至少填写一条域名规则')
        return
    }
    containerDomainSubmitLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainFilter`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({
                    containerID: containerDomainForm.containerID,
                    domains: domains
                })
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('设置成功')
                showContainerDomainDialog.value = false
                // 如果查询面板已选中同一容器，刷新列表
                if (domainFilterContainerID.value === containerDomainForm.containerID) {
                    await fetchDomainFilterList()
                }
            } else {
                ElMessage.error(data.message || '设置失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error(`${t('network.setContainerDomainFilter')}失败:`, error)
        ElMessage.error('设置失败，请检查网络连接')
    } finally {
        containerDomainSubmitLoading.value = false
    }
}

const fetchGlobalDomainFilterList = async () => {
    globalDomainFilterLoading.value = true
    // 切换到全局模式，清空容器选择
    domainQueryMode.value = 'global'
    domainFilterContainerID.value = ''
    domainFilterList.value = []
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainFilter/global`,
            { headers: getAuthHeaders(selectedDeviceIP.value) }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                globalDomainFilterList.value = (data.data?.domains || []).map(d => ({ domain: d }))
            } else {
                ElMessage.error(data.message || '查询失败')
                globalDomainFilterList.value = []
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
            globalDomainFilterList.value = []
        } else {
            ElMessage.error('接口请求失败')
            globalDomainFilterList.value = []
        }
    } catch (error) {
        console.error(`${t('network.queryGlobalDomainFilter')}失败:`, error)
        ElMessage.error('查询失败，请检查网络连接')
        globalDomainFilterList.value = []
    } finally {
        globalDomainFilterLoading.value = false
    }
}

const submitGlobalDomainFilter = async () => {
    const domains = globalDomainForm.rules
        .filter(r => r.value.trim())
        .map(r => r.prefix + r.value.trim())
    if (!domains.length) {
        ElMessage.warning('请至少填写一条域名规则')
        return
    }
    globalDomainSubmitLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/domainFilter/global`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({ domains: domains })
            }
        )
        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('全局域名过滤设置成功')
                showGlobalDomainDialog.value = false
            } else {
                ElMessage.error(data.message || '设置失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error(`${t('network.setGlobalDomainFilter')}失败:`, error)
        ElMessage.error('设置失败，请检查网络连接')
    } finally {
        globalDomainSubmitLoading.value = false
    }
}



const handleSearch = () => {
    fetchGroupList()
}

const handleSizeChange = (val) => {
    fetchGroupList()
}

const handlePageChange = (val) => {
    fetchGroupList()
}

const setContainerVpc = async () => {
    if (!selectedDeviceIP.value) {
        ElMessage.warning('请先在左侧选择设备')
        return
    }

    vpcSelectedContainers.value = []
    vpcSelectedGroupId.value = ''
    vpcSelectedNodeId.value = ''
    vpcContainerList.value = []
    vpcGroupList.value = []
    vpcNodeList.value = []
    vpcContainerSearch.value = ''
    vpcSlotSearch.value = ''
    vpcSelectMode.value = 'specified'
    stepActive.value = 0
    vpcDialogVisible.value = true

    await loadVpcContainers()
}

// 处理选择云机
const handleVpcContainerSelectionChange = (selection) => {
    vpcSelectedContainers.value = selection
}

const nextVpcStep = async () => {
    stepActive.value++

    if (stepActive.value === 1) {
        await loadVpcGroups()
    } else if (stepActive.value === 2) {
        await loadVpcNodes()
    }
}

const loadVpcContainers = async () => {
    if (!selectedDeviceIP.value) return

    vpcDialogLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/android?name=&running=false`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                const list = data.data?.list && data.data?.list.filter(c => c.networkName !== 'myt') || []
                vpcContainerList.value = list.map(c => ({
                    id: c.id,
                    name: c.name,
                    ip: c.ip || '',
                    status: c.status === 'running' ? 'running' : 'shutdown',
                    indexNum: c.indexNum || '',
                    dns: c.dns || ''
                }))
            } else {
                ElMessage.error(data.message || '获取云机列表失败')
            }
        } else if (response.status === 401) {
            console.error('获取云机列表认证失败 (401):', {
                deviceIP: selectedDeviceIP.value,
                hasPassword: !!getAuthHeaders(selectedDeviceIP.value).Authorization
            })
            ElMessage.error('认证失败，请确认设备已授权')
        } else {
            console.error('获取云机列表请求失败:', response.status)
            ElMessage.error(`接口请求失败 (${response.status})`)
        }
    } catch (error) {
        console.error('获取云机列表失败:', error)
        ElMessage.error('获取云机列表失败，请检查网络连接')
    } finally {
        vpcDialogLoading.value = false
    }
}

const loadVpcGroups = async () => {
    console.log('loadVpcGroups', selectedDeviceIP.value);

    if (!selectedDeviceIP.value) return

    vpcDialogLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/group`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                vpcGroupList.value = data.data?.list || []
            } else {
                ElMessage.error(data.message || '获取分组列表失败')
            }
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('获取分组列表失败:', error)
        ElMessage.error('获取分组列表失败，请检查网络连接')
    } finally {
        vpcDialogLoading.value = false
    }
}

const loadVpcNodes = async () => {
    if (!selectedDeviceIP.value || !vpcSelectedGroupId.value) return

    const selectedGroup = vpcGroupList.value.find(g => g.id === vpcSelectedGroupId.value)
    if (!selectedGroup?.vpcs?.list) {
        vpcNodeList.value = []
        return
    }

    vpcNodeList.value = selectedGroup.vpcs.list.map(vpc => ({
        ...vpc,
        groupAlias: selectedGroup.alias,
        latency: null
    }))
}

const handleContainerRuleSelectionChange = (selection) => {
    selectedContainerRules.value = selection
}

const handleVpcNodeSelectionChange = (selection) => {
    selectedVpcNodes.value = selection
}

const batchDeleteVpcNodes = async () => {
    if (selectedVpcNodes.value.length === 0) return

    try {
        await ElMessageBox.confirm(
            `确定要删除选中的 ${selectedVpcNodes.value.length} 个分组节点吗？`,
            '批量删除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const deletePromises = selectedVpcNodes.value.map(async (row) => {
            try {
                const response = await fetch(
                    `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc?vpcID=${row.id}`,
                    {
                        method: 'DELETE',
                        headers: {
                            ...getAuthHeaders(selectedDeviceIP.value)
                        }
                    }
                )

                if (response.ok) {
                    const data = await response.json()
                    return { success: data.code === 0, message: data.message, row }
                } else {
                    return { success: false, message: '接口请求失败', row }
                }
            } catch (error) {
                return { success: false, message: '网络错误', row }
            }
        })

        const results = await Promise.all(deletePromises)
        const successCount = results.filter(r => r.success).length
        const failCount = results.length - successCount

        if (failCount === 0) {
            ElMessage.success(`批量删除成功，共删除 ${successCount} 个节点`)
        } else {
            ElMessage.warning(`删除完成：成功 ${successCount} 个，失败 ${failCount} 个`)
        }

        fetchGroupList(false)
    } catch (error) {
        if (error !== 'cancel') {
            console.error('批量删除失败:', error)
            ElMessage.error('批量删除失败，请检查网络连接')
        }
    }
}

const batchTestSpeed = async () => {
    if (selectedVpcNodes.value.length === 0) return

    ElMessage.info(`开始 ${t('network.batchSpeedTest')}，共 ${selectedVpcNodes.value.length} 个节点`)

    const testPromises = selectedVpcNodes.value.map(async (row) => {
        try {
            const profile = JSON.parse(row.profile)
            row.latency = '测速中...'

            const response = await fetch(
                `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/test?address=${profile.server}:${profile.serverPort}`,
                {
                    method: 'GET',
                    headers: getAuthHeaders(selectedDeviceIP.value)
                }
            )

            if (response.ok) {
                const data = await response.json()
                if (data.code === 0) {
                    if (data.data.msg !== '') {
                        row.latency = '-1ms'
                        return { success: false, row, message: data.data.msg }
                    } else {
                        row.latency = data.data.latency
                        return { success: true, row, latency: data.data.latency }
                    }
                } else {
                    row.latency = '-1ms'
                    return { success: false, row, message: data.message || '测速失败' }
                }
            } else {
                row.latency = '-1ms'
                return { success: false, row, message: '接口请求失败' }
            }
        } catch (error) {
            row.latency = '-1ms'
            return { success: false, row, message: '网络错误' }
        }
    })

    const results = await Promise.all(testPromises)
    const successCount = results.filter(r => r.success).length
    const failCount = results.length - successCount

    if (failCount === 0) {
        ElMessage.success(`${t('network.batchSpeedTest')}完成，共测试 ${successCount} 个节点`)
    } else {
        ElMessage.warning(`测速完成：成功 ${successCount} 个，失败 ${failCount} 个`)
    }
}

const batchClearContainerVpc = async () => {
    if (selectedContainerRules.value.length === 0) return

    try {
        await ElMessageBox.confirm(
            `确定要清除选中的 ${selectedContainerRules.value.length} 个云机的VPC节点吗？`,
            '批量清除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const deviceIP = selectedDeviceIP.value
        if (!deviceIP) {
            ElMessage.warning('无法确定设备IP')
            return
        }

        const names = selectedContainerRules.value.map(row => row.containerName)

        const response = await fetch(
            `http://${getDeviceAddr(deviceIP)}/mytVpc/delRule/batch`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(deviceIP)
                },
                body: JSON.stringify({
                    name: names
                })
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('批量清除成功')
                fetchContainerRule()
            } else {
                ElMessage.error(data.message || '批量清除失败')
            }
        } else if (response.status === 401) {
            console.error('批量清除VPC节点认证失败 (401):', {
                deviceIP: deviceIP,
                hasPassword: !!getAuthHeaders(deviceIP).Authorization
            })
            ElMessage.error('认证失败，请确认设备已授权')
        } else {
            console.error('批量清除VPC节点请求失败:', response.status)
            ElMessage.error(`接口请求失败 (${response.status})`)
        }
        
    } catch (error) {
        if (error !== 'cancel') {
             console.error('Batch clear failed', error)
             ElMessage.error('批量清除失败，请检查网络连接')
        }
    }
}

const clearContainerVpc = async (row) => {
    const deviceIP = row.deviceIp || selectedDeviceIP.value
    if (!deviceIP) {
        ElMessage.warning('无法确定设备IP')
        return
    }

    try {
        await ElMessageBox.confirm(
            `确定要清除云机"${row.containerName}"的VPC节点吗？`,
            '清除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const response = await fetch(
            `http://${getDeviceAddr(deviceIP)}/mytVpc/delRule`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(deviceIP)
                },
                body: JSON.stringify({
                    name: row.containerName
                })
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('清除VPC节点成功')
                fetchContainerRule()
            } else {
                ElMessage.error(data.message || '清除失败')
            }
        } else if (response.status === 401) {
            console.error('清除VPC节点认证失败 (401):', {
                deviceIP: deviceIP,
                hasPassword: !!getAuthHeaders(deviceIP).Authorization
            })
            ElMessage.error('认证失败，请确认设备已授权')
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(deviceIP, selectedDeviceVersion.value)
        } else {
            console.error('清除VPC节点请求失败:', response.status)
            ElMessage.error(`接口请求失败 (${response.status})`)
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error('清除失败:', error)
            ElMessage.error('清除失败，请检查网络连接')
        }
    }
}


const enableContainerDnsWhitelist = async (row) => {
    const deviceIP = row.deviceIp || selectedDeviceIP.value
    if (!deviceIP) {
        ElMessage.warning('无法确定设备IP')
        return
    }

    try {
        await ElMessageBox.confirm(
            `${(row.WhiteListDns != null && row.WhiteListDns.length > 0) ? '确定要关闭' : '确定要开启'}DNS"${row.containerName}"的白名单吗？`,
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const isEnable = !(row.WhiteListDns != null && row.WhiteListDns.length > 0)
        let whiteListDns = []

        if (isEnable) {
            const containerName = row.containerName
            let containerDns = ''

            if (vpcContainerList.value.length > 0) {
                const matchedContainer = vpcContainerList.value.find(c => c.name === containerName)
                containerDns = matchedContainer?.dns || ''
            }

            if (!containerDns) {
                try {
                    const resp = await fetch(`http://${getDeviceAddr(deviceIP)}/android?name=&running=false`, {
                        headers: getAuthHeaders(deviceIP)
                    })
                    if (resp.ok) {
                        const data = await resp.json()
                        if (data.code === 0) {
                            const list = data.data?.list || []
                            const matched = list.find(c => c.name === containerName)
                            containerDns = matched?.dns || ''
                        }
                    } else if (resp.status === 401) {
                        console.error('获取云机DNS认证失败 (401):', {
                            deviceIP: deviceIP,
                            hasPassword: !!getAuthHeaders(deviceIP).Authorization
                        })
                    }
                } catch (e) {
                    console.error('获取云机DNS失败:', e)
                }
            }

            whiteListDns = containerDns ? [containerDns] : []
        }

        const response = await fetch(
            `http://${getDeviceAddr(deviceIP)}/mytVpc/whiteListDns`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(deviceIP)
                },
                body: JSON.stringify({
                    ruleID: row.id,
                    enable: isEnable,
                    whiteListDns: whiteListDns
                })
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('操作成功')
                fetchContainerRule()
            } else {
                ElMessage.error(data.message || '操作失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(deviceIP, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error('操作失败:', error)
            ElMessage.error('操作失败，请检查网络连接')
        }
    }

} 

const handleVpcRowClick = (row) => {
    selectedVpcNode.value = row.id
}

const confirmSetVpc = async () => {
    if (vpcSelectedContainers.value.length === 0) {
        ElMessage.warning('请选择云机')
        return
    }

    if (!selectedDeviceIP.value) {
        ElMessage.error('设备IP无效')
        return
    }

    if (vpcSelectMode.value === 'random') {
        if (!vpcSelectedGroupId.value) {
            ElMessage.warning('请选择分组')
            return
        }

        const selectedGroup = vpcGroupList.value.find(g => g.id === vpcSelectedGroupId.value)
        if (!selectedGroup?.vpcs?.list || selectedGroup.vpcs.list.length === 0) {
            ElMessage.error('该分组下没有可用的节点')
            return
        }

        const allNodes = selectedGroup.vpcs.list
        // 批量分配时，每台云机随机分配不同的节点
        vpcDialogLoading.value = true
        let successCount = 0
        let failCount = 0
        try {
            for (const container of vpcSelectedContainers.value) {
                const randomIndex = Math.floor(Math.random() * allNodes.length)
                const randomVpcId = allNodes[randomIndex].id

                const response = await fetch(
                    `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/addRule/batch`,
                    {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            ...getAuthHeaders(selectedDeviceIP.value)
                        },
                        body: JSON.stringify({
                            names: [container.name],
                            vpcID: randomVpcId,
                        })
                    }
                )
                if (response.ok) {
                    const data = await response.json()
                    if (data.code === 0) {
                        successCount++
                    } else {
                        failCount++
                    }
                } else {
                    failCount++
                }
            }

            if (failCount === 0) {
                ElMessage.success(`设置VPC节点成功，共 ${successCount} 台`)
            } else {
                ElMessage.warning(`设置完成，成功 ${successCount} 台，失败 ${failCount} 台`)
            }
            vpcDialogVisible.value = false
            fetchContainerRule()
        } catch (error) {
            console.error('设置VPC节点失败:', error)
            ElMessage.error('设置VPC节点失败，请检查网络连接')
        } finally {
            vpcDialogLoading.value = false
        }
        return
    }

    // 指定节点模式
    let finalVpcId = vpcSelectedNodeId.value
    if (!vpcSelectedNodeId.value) {
        ElMessage.warning('请选择节点')
        return
    }

    vpcDialogLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytVpc/addRule/batch `,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...getAuthHeaders(selectedDeviceIP.value)
                },
                body: JSON.stringify({
                    names: vpcSelectedContainers.value.map(c => c.name),
                    vpcID: finalVpcId,
                    // whiteListDns: vpcSelectedContainerDns.value ? [vpcSelectedContainerDns.value] : []
                })
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('设置VPC节点成功')
                vpcDialogVisible.value = false
                fetchContainerRule()
            } else {
                ElMessage.error(data.message || '设置失败')
            }
        } else if (response.status === 401) {
            console.error('设置VPC节点认证失败 (401):', {
                deviceIP: selectedDeviceIP.value,
                hasPassword: !!getAuthHeaders(selectedDeviceIP.value).Authorization,
                savedPasswords: localStorage.getItem('devicePasswords')
            })
            ElMessage.error('认证失败，请确认设备已授权')
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            console.error('设置VPC节点请求失败:', response.status, response.statusText)
            ElMessage.error(`接口请求失败 (${response.status})`)
        }
    } catch (error) {
        console.error('设置VPC节点失败:', error)
        ElMessage.error('设置VPC节点失败，请检查网络连接')
    } finally {
        vpcDialogLoading.value = false
    }
}

const updateProtocolLabel = (value) => {
    const option = protocolOptions.find(opt => opt.value === value)
    selectedProtocolLabel.value = option ? option.label : ''
}

const validateAddressFormat = (address, protocol) => {
    const trimmedAddress = address.trim()

    // socks 和 http 使用相同的格式验证
    if (protocol === '5' || protocol === '6') {
        if (!trimmedAddress.endsWith('/')) {
            return { 
                valid: false, 
                message: '地址必须以 / 结尾' 
            }
        }
        
        const parts = trimmedAddress.split('/').filter(p => p)
        
        if (parts.length < 2) {
            return { 
                valid: false, 
                message: '格式错误，请按照: IP/端口/用户名/密码/节点别名 格式输入' 
            }
        }
    }
    
    return { valid: true, message: '格式正确' }
}

const validateProtocolMatch = (address, expectedProtocol) => {
    const trimmedAddress = address.trim().toLowerCase()

    const protocolPrefixes = {
        '1': 'vmess://',
        '2': 'vless://',
        '3': 'ss://',
        '4': 'trojan://',
        '5': 'socks://',
        '6': 'http://',
        '7': 'wireguard://',
        '8': 'hysteria2://'
    }

    const expectedPrefix = protocolPrefixes[expectedProtocol]
    if (!expectedPrefix) {
        return { valid: true, message: '' }
    }

    // socks 和 http 不受协议前缀限制（使用 IP/端口/用户名/密码/ 格式）
    if (expectedProtocol === '5' || expectedProtocol === '6') {
        return { valid: true, message: '' }
    }

    if (!trimmedAddress.startsWith(expectedPrefix)) {
        return {
            valid: false,
            message: `地址必须以 ${expectedPrefix} 开头`
        }
    }

    return { valid: true, message: '' }
}

const validateBatchAddresses = (addresses, protocol) => {
    const errors = []
    let index = 0

    const protocolExamples = {
        '1': 'vmess://base64(json)',
        '2': 'vless://uuid@server:port?type=ws&security=tls',
        '3': 'ss://base64(method:password)@server:port',
        '4': 'trojan://password@server:port?type=tcp&security=tls',
        '5': 'IP/端口/用户名/密码/节点别名',
        '6': 'IP/端口/用户名/密码/节点别名',
        '7': 'wireguard://privateKey@server:port',
        '8': 'hysteria2://auth@server:port'
    }

    addresses.forEach(address => {
        index++
        const formatResult = validateAddressFormat(address, protocol)
        if (!formatResult.valid) {
            errors.push(`第${index}行: ${formatResult.message}`)
        }

        // 验证协议类型一致性（socks除外）
        const protocolResult = validateProtocolMatch(address, protocol)
        if (!protocolResult.valid) {
            errors.push(`第${index}行: ${protocolResult.message}`)
        }
    })

    if (errors.length > 0) {
        const example = protocolExamples[protocol] || ''
        return {
            valid: false,
            message: `验证失败:\n${errors.join('\n')}\n\n正确格式: ${example}`
        }
    }

    return { valid: true, message: '所有地址格式正确' }
}

const fetchPrivateNics = async () => {
    if (!selectedDeviceIP.value) {
        privateNicList.value = []
        return
    }

    privateNicLoading.value = true
    try {
        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytBridge`,
            {
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                privateNicList.value = data.data.list || []
            } else {
                ElMessage.error(data.message || '获取私有网卡列表失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        console.error('获取私有网卡列表失败:', error)
        ElMessage.error('获取私有网卡列表失败，请检查网络连接')
    } finally {
        privateNicLoading.value = false
    }
}

const handleAddPrivateNic = () => {
    createPrivateNicForm.customName = ''
    createPrivateNicForm.cidr = ''
    createPrivateNicDialogVisible.value = true
    if (createPrivateNicFormRef.value) {
        createPrivateNicFormRef.value.resetFields()
    }
}

const submitCreatePrivateNic = async () => {
    if (!createPrivateNicFormRef.value) return

    await createPrivateNicFormRef.value.validate(async (valid) => {
        if (valid) {
            createPrivateNicLoading.value = true
            try {
                const response = await fetch(
                    `http://${getDeviceAddr(selectedDeviceIP.value)}/mytBridge`,
                    {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            ...getAuthHeaders(selectedDeviceIP.value)
                        },
                        body: JSON.stringify({
                            customName: createPrivateNicForm.customName,
                            cidr: `${createPrivateNicForm.cidr}/24`
                        })
                    }
                )

                if (response.ok) {
                    const data = await response.json()
                    if (data.code === 0) {
                        ElMessage.success('创建私有网卡成功')
                        createPrivateNicDialogVisible.value = false
                        fetchPrivateNics()
                    } else {
                        ElMessage.error(data.message || '创建私有网卡失败')
                    }
                } else if (response.status === 404) {
                    await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
                } else {
                    ElMessage.error('接口请求失败')
                }
            } catch (error) {
                console.error('创建私有网卡失败:', error)
                ElMessage.error('创建私有网卡失败，请检查网络连接')
            } finally {
                createPrivateNicLoading.value = false
            }
        }
    })
}

const handleEditPrivateNic = (row) => {
    editPrivateNicForm.name = row.name
    // editPrivateNicForm.displayName = row.name.replace('myt_bridge_', '')
    editPrivateNicForm.cidr = row.cidr.split('/')[0]
    editPrivateNicDialogVisible.value = true
    if (editPrivateNicFormRef.value) {
        editPrivateNicFormRef.value.clearValidate()
    }
}

const submitEditPrivateNic = async () => {
    if (!editPrivateNicFormRef.value) return

    await editPrivateNicFormRef.value.validate(async (valid) => {
        if (valid) {
            editPrivateNicLoading.value = true
            try {
                const response = await fetch(
                    `http://${getDeviceAddr(selectedDeviceIP.value)}/mytBridge`,
                    {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            ...getAuthHeaders(selectedDeviceIP.value)
                        },
                        body: JSON.stringify({
                            name: editPrivateNicForm.name,
                            newCidr: `${editPrivateNicForm.cidr}/24`
                        })
                    }
                )

                if (response.ok) {
                    const data = await response.json()
                    if (data.code === 0) {
                        ElMessage.success('更新私有网卡成功')
                        editPrivateNicDialogVisible.value = false
                        fetchPrivateNics()
                    } else {
                        ElMessage.error(data.message || '更新私有网卡失败')
                    }
                } else if (response.status === 404) {
                    await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
                } else {
                    ElMessage.error('接口请求失败')
                }
            } catch (error) {
                console.error('更新私有网卡失败:', error)
                ElMessage.error('更新私有网卡失败，请检查网络连接')
            } finally {
                editPrivateNicLoading.value = false
            }
        }
    })
}

const handleDeletePrivateNic = async (row) => {
    try {
        await ElMessageBox.confirm(
            `确定要删除私有网卡"${row.name}"吗？`,
            '删除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/mytBridge?name=${row.name}`,
            {
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('删除私有网卡成功')
                fetchPrivateNics()
            } else {
                ElMessage.error(data.message || '删除私有网卡失败')
            }
        } else if (response.status === 404) {
            await checkAndWarnSDKVersion(selectedDeviceIP.value, selectedDeviceVersion.value)
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error('删除私有网卡失败:', error)
            ElMessage.error('删除私有网卡失败，请检查网络连接')
        }
    }
}

const handleAddPublicNic = () => {
    publicNicList.value.push({
        name: 'MacVlan',
        gw: '',
        subnet: '',
        isolate: false,
        loading: false,
        isNew: true
    })
}

const handleUpdatePublicNic = async (nic) => {
    if (!selectedDeviceIP.value) {
        ElMessage.error('请先选择设备')
        return
    }

    try {
        await ElMessageBox.confirm(
            '更新macvlan会关闭所有安卓，请慎重更新',
            '更新确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        nic.loading = true
        try {
            const response = await fetch(
                `http://${getDeviceAddr(selectedDeviceIP.value)}/macvlan`,
                {
                    method: nic.isNew ? 'POST' : 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        ...getAuthHeaders(selectedDeviceIP.value)
                    },
                    body: JSON.stringify({
                        gw: nic.gw,
                        subnet: nic.subnet,
                        // private: nic.isolate
                    })
                }
            )

            if (response.ok) {
                const data = await response.json()
                if (data.code === 0) {
                    ElMessage.success('保存成功')
                    nic.isNew = false
                    fetchPublicNics()
                } else {
                    ElMessage.error(data.message || '保存失败')
                }
            } else {
                ElMessage.error('接口请求失败')
            }
        } catch (error) {
            console.error('保存失败:', error)
            ElMessage.error('保存失败，请检查网络连接')
        } finally {
            nic.loading = false
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error('操作取消或失败:', error)
        }
    }
}

const handleDeletePublicNic = async (nic) => {
    try {
        await ElMessageBox.confirm(
            '确定要删除该公有网卡吗？',
            '删除确认',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }
        )

        const response = await fetch(
            `http://${getDeviceAddr(selectedDeviceIP.value)}/macvlan`,
            {
                method: 'DELETE',
                headers: getAuthHeaders(selectedDeviceIP.value)
            }
        )

        if (response.ok) {
            const data = await response.json()
            if (data.code === 0) {
                ElMessage.success('删除成功')
                fetchPublicNics()
            } else {
                ElMessage.error(data.message || '删除失败')
            }
        } else {
            ElMessage.error('接口请求失败')
        }
    } catch (error) {
        if (error !== 'cancel') {
            console.error('删除失败:', error)
            ElMessage.error('删除失败，请检查网络连接')
        }
    }
}

const fetchPublicNics = async () => {
    if (!selectedDeviceIP.value) {
        publicNicList.value = []
        return
    }

    publicNicLoading.value = true
    try {
        const list = []
        
        // 1. 获取物理网卡信息 (via /macvlan interface as per requirement)
        let macVlanExists = false
        try {
            const response = await fetch(`http://${getDeviceAddr(selectedDeviceIP.value)}/macvlan`, {
                headers: getAuthHeaders(selectedDeviceIP.value)
            })
            if (response.ok) {
                const data = await response.json()
                if (data.code === 0 && data.data) {
                    const info = data.data
                    
                    // Helper to parse info string
                    const parseInfo = (infoStr) => {
                         let res = { gw: '', subnet: '', isolate: false }
                         let data = infoStr
                         try {
                             if (typeof infoStr === 'string' && (infoStr.startsWith('{') || infoStr.startsWith('['))) {
                                 data = JSON.parse(infoStr)
                             }
                             if (typeof data === 'object' && data !== null) {
                                 if (data.IPAM && Array.isArray(data.IPAM.Config) && data.IPAM.Config.length > 0) {
                                     res.gw = data.IPAM.Config[0].Gateway || ''
                                     res.subnet = data.IPAM.Config[0].Subnet || ''
                                     if (data.Options && data.Options.macvlan_mode === 'private') {
                                         res.isolate = true
                                     }
                                 } else {
                                     res.gw = data.gw || data.gateway || ''
                                     res.subnet = data.subnet || ''
                                     res.isolate = !!data.isolate
                                 }
                             }
                         } catch (e) {
                             console.error('Parse error', e)
                         }
                         return res
                    }

                    // Display Physical Info (netWork_eth0 or macVlan as fallback if eth0 missing but user calls it physical)
                    // Priority: netWork_eth0 -> macVlan (labeled as Physical)
                    let physicalFound = false
                    if (info.netWork_eth0) {
                        const parsed = parseInfo(info.netWork_eth0)
                        list.push({
                            name: '物理网卡信息',
                            gw: parsed.gw,
                            subnet: parsed.subnet,
                            isolate: parsed.isolate,
                            loading: false,
                            isNew: false,
                            type: 'physical'
                        })
                        physicalFound = true
                    } 
                    
                    // If netWork_eth0 is missing, but macVlan exists, and user says /macvlan returns physical info,
                    // we use macVlan data as "Physical Network Card Info".
                    if (!physicalFound && info.macVlan) {
                         const parsed = parseInfo(info.macVlan)
                         list.push({
                            name: '物理网卡信息',
                            gw: parsed.gw,
                            subnet: parsed.subnet,
                            isolate: parsed.isolate,
                            loading: false,
                            isNew: false,
                            type: 'physical'
                        })
                    }

                    // Check if MacVlan exists (to determine if we need to CREATE or UPDATE the macvlan card)
                    if (info.macVlan) {
                        macVlanExists = true
                    }
                }
            }
        } catch (e) {
            console.error('Fetch macvlan failed', e)
        }

        // 2. 获取 MacVlan 信息 (via /server/network as per requirement)
        try {
            const netResp = await fetch(`http://${getDeviceAddr(selectedDeviceIP.value)}/server/network`, {
                headers: getAuthHeaders(selectedDeviceIP.value)
            })
            if (netResp.ok) {
                const netData = await netResp.json()
                if (netData.code === 0 && netData.data && netData.data.info) {
                    const info = netData.data.info
                    list.push({
                        name: 'MacVlan',
                        gw: info.gateway,
                        subnet: info.networkCIDR,
                        isolate: false, // Default to false as switch is removed
                        loading: false,
                        isNew: !macVlanExists, // Determine if it's new based on /macvlan response
                        type: 'macvlan'
                    })
                }
            }
        } catch (e) {
            console.error('Fetch server network failed', e)
        }

        publicNicList.value = list
        console.log('Public NICs:', publicNicList)

    } catch (error) {
        console.error('获取公有网卡信息失败:', error)
        ElMessage.error('获取公有网卡信息失败，请检查网络连接')
    } finally {
        publicNicLoading.value = false
    }
}

const fetchNetworks = async () => {
    fetchGroupList(false)
    if (activeTab.value === 'vpc') {
        fetchContainerRule()
    } else if (activeTab.value === 'private-nic') {
        fetchPrivateNics()
    } else if (activeTab.value === 'public-nic') {
        // fetchPublicNics()
    }
}

defineExpose({
    fetchNetworks
})
</script>

<style scoped>
.network-management-container {
    height: 100%;
    box-sizing: border-box;
}

.network-content {
    height: 100%;
    background: white;
    border-radius: 4px;
}

.network-tabs {
    height: 100%;
}

:deep(.el-tabs__content) {
    height: calc(100% - 55px);
}

:deep(.el-tab-pane) {
    height: 100%;
    overflow-y: auto; /* 允许内容超出时滚动 */
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
    padding: 10px 12px;
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
    padding: 16px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid #ebeef5;
    background: #fafafa;
}

.vpc-content {
    display: flex;
    /* justify-content: center; */
    /* align-items: center; */
    gap: 16px;
    height: 100%;
    /* padding: 40px; */
}

:deep(.el-table .cell) {
    display: block;
}

.protocol-tips {
    margin-top: 8px;
}

.protocol-btn {
    font-size: 13px;
    padding: 0;
}

.protocol-list {
    padding: 8px 0;
}

.protocol-title {
    font-weight: bold;
    margin-bottom: 8px;
    color: #fff;
}

.protocol-item {
    padding: 2px 0;
    font-family: monospace;
    color: #67c23a;
}

.group-option-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.vpc-dialog-content {
    display: flex;
    flex-direction: column;
    height: 500px;
    overflow: hidden;
}

.step-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
}

:deep(.el-table__body-wrapper) {
   max-height: calc(100vh - 400px) !important; 
}
</style>