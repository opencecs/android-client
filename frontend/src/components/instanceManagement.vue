<template>
    <div class="instance-management-container">

        <div class="instance-content">
            <div v-if="!isLoggedIn" class="login-required">
                <el-empty description="">
                    <template #description>
                        <div style="margin-bottom: 20px; font-size: 14px; color: #909399;">
                            {{ $t('instance.loginRequired') }}
                        </div>
                        <div style="display: flex; gap: 10px; justify-content: center;">
                            <el-button type="primary" @click="showSyncAuthDialog">
                                {{ $t('instance.login') }}
                            </el-button>
                            <el-button @click="openRegisterDialog">
                                {{ $t('instance.register') }}
                            </el-button>
                        </div>
                    </template>
                </el-empty>
            </div>
            <div v-else class="instance-list">
                <div class="search-bar">
                    <el-select v-model="searchState" :placeholder="$t('common.status')" clearable style="width: 120px;" @change="handleSearch">
                        <el-option :label="$t('common.all')" value=""></el-option>
                        <el-option :label="$t('instance.statusNormal')" :value="0"></el-option>
                        <el-option :label="$t('instance.statusExpiring')" :value="1"></el-option>
                        <el-option :label="$t('instance.statusExpired')" :value="2"></el-option>
                        <el-option :label="$t('instance.noInstance')" :value="-1"></el-option>
                    </el-select>
                    <el-input v-model="searchHostRabbet" size="medium" :placeholder="$t('instance.searchPlaceholder')" clearable @keyup.enter="handleSearch"
                        class="search-input">
                        <template #prefix>
                            <el-icon>
                                <Search />
                            </el-icon>
                        </template>
                    </el-input>
                    <el-button type="primary" @click="handleSearch" class="search-button">
                        {{ $t('instance.query') }}
                    </el-button>
                    <div>
                      <el-button type="primary" @click="handleBatchBuy" class="search-button">
                        {{ $t('instance.purchaseRenew') }}
                    </el-button>   
                    </div>
                    <div class="status-legend">
                        <span class="legend-item">
                            <span class="status-dot status-normal"></span>
                            {{ $t('instance.statusNormal') }}
                        </span>
                        <span class="legend-item">
                            <span class="status-dot status-warning"></span>
                            {{ $t('instance.statusExpiring') }}
                        </span>
                        <span class="legend-item">
                            <span class="status-dot status-expired"></span>
                            {{ $t('instance.statusExpired') }}
                        </span>
                    </div>
                </div>
                <el-table 
                    ref="instanceTableRef"
                    :data="pagedInstanceList" 
                    stripe 
                    @select="handleTableSelect" 
                    @select-all="handleTableSelectAll"
                    row-key="rabbet">
                    <el-table-column type="selection" width="55" align="center"></el-table-column>
                    <el-table-column :label="$t('instance.host')" width="180" align="center">
                        <template #default="scope">
                            <span class="host-name">{{ formName(scope.row.rabbet) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('instance.instanceIP')" align="center">
                        <template #default="scope">
                            <div class="child-container"
                                v-if="hasDisplayChildren(scope.row)">
                                <div v-for="(item, key) in getDisplayChildren(scope.row)" :key="key" class="child-item">
                                    <el-checkbox v-model="item.selected"
                                        @change="handleChildSelect(scope.row, key, item)" class="child-checkbox">
                                        <span :class="['child-name', item.isEmpty ? 'text-empty' : getStateTextClass(item.state)]">{{ $t('instance.instancePrefix') }}-{{ key }}</span>
                                    </el-checkbox>
                                </div>
                            </div>
                            <span v-else class="no-child">{{ $t('instance.noInstance') }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('common.operation')" width="150" align="center" fixed="right">
                        <template #default="scope">
                            <el-button type="primary" size="mini" @click="handleBatchOperate(scope.row)">
                                {{ $t('instance.details') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </el-table>
                <div class="pagination-container" v-if="displayTotal > 0">
                    <div class="pagination-wrapper">
                        <el-pagination background layout="prev, pager, next, jumper" :current-page="currentPage"
                            :page-size="pageSize" :total="displayTotal" @current-change="handleCurrentChange"
                            @size-change="handleSizeChange"></el-pagination>
                        <div class="total-text">{{ $t('instance.totalCount') }}{{ displayTotal }}</div>
                    </div>
                </div>
            </div>
        </div>

        <el-dialog v-model="packageDialogVisible" :title="$t('instance.selectPackage')" width="50%">
            <template #header>
                <div class="dialog-header">
                    <span class="dialog-title">{{ $t('instance.selectPackage') }}</span>
                </div>
            </template>
            <div class="selected-instances-section" v-if="selectedInstances.length > 0">
                <div class="section-title">{{ $t('instance.selectedInstances') }} ({{ selectedInstances.length }}{{ $t('instance.unit') }})</div>
                <div class="selected-tags">
                    <el-tag 
                        v-for="(item, index) in selectedInstances" 
                        :key="index" 
                        type="info" 
                        class="instance-tag" 
                        closable 
                        @close="handleDeselectInstance(index)">
                        {{ $t('instance.instancePrefix') }}-{{ item.instanceKey }}
                    </el-tag>
                </div>
            </div>
            <el-divider v-if="selectedInstances.length > 0" />
            <div v-if="loadingPackages" class="loading-container">
                <el-icon class="is-loading"><Loading /></el-icon>
                <span>{{ $t('instance.loadingPackages') }}</span>
            </div>
            <div v-else-if="packageList.length === 0" class="empty-container">
                <el-empty :description="$t('instance.noPackages')" />
            </div>
            <div v-else class="package-list">
                <el-radio-group v-model="selectedPackage">
                    <div v-for="pkg in packageList" :key="pkg.id" class="package-item" @click="selectedPackage = pkg.id">
                        <el-card class="package-card" :class="{ 'package-selected': selectedPackage === pkg.id }">
                            <div class="package-header">
                                <el-radio :label="pkg.id">{{ pkg.name }}</el-radio>
                            </div>
                            <div class="package-content">
                                <div class="package-price">¥{{ pkg.price }}元</div>
                            </div>
                        </el-card>
                    </div>
                </el-radio-group>
            </div>
            <div class="total-section" v-if="selectedPackage && selectedInstances.length > 0">
                <div class="total-row">
                    <span class="total-label">{{ $t('instance.subtotal') }}</span>
                    <span class="total-amount">¥{{ (parseFloat(packageList.find(p => p.id === selectedPackage)?.price || 0) * selectedInstances.length).toFixed(2) }}元</span>
                    <span class="total-instances">({{ selectedInstances.length }} {{ $t('instance.instances') }})</span>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="packageDialogVisible = false">{{ $t('common.cancel') }}</el-button>
                    <el-button type="primary" @click="handleConfirmBuy" :disabled="!selectedPackage || selectedInstances.length === 0">{{ $t('instance.confirmBuy') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-model="qrcodeDialogVisible" :title="$t('instance.scanToPay')" width="360px" :close-on-click-modal="false">
            <div class="qrcode-container">
                <img v-if="qrcodeImage" :src="'data:image/png;base64,' + qrcodeImage" alt="支付二维码" class="qrcode-image" />
                <div v-else class="qrcode-loading">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    <span>{{ $t('instance.loadingQR') }}</span>
                </div>
            </div>
            <div class="qrcode-tip">{{ $t('instance.alipayTip') }}</div>
        </el-dialog>

        <el-dialog v-model="detailDialogVisible" :title="$t('instance.detailTitle') + ' - ' + currentDetailHost" width="600px">
            <el-table :data="paginatedDetailList" stripe>
                <el-table-column prop="slot" :label="$t('instance.instanceSlot')" width="100" align="center">
                    <template #default="scope">
                        <span>{{ $t('instance.instancePrefix') }}-{{ scope.row.slot }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="extime" :label="$t('instance.validUntil')" align="center">
                    <template #default="scope">
                        <span :class="getStateTextClass(scope.row.extimeState)">{{ scope.row.extime }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="state" :label="$t('common.status')" width="120" align="center">
                    <template #default="scope">
                        <el-tag :type="getStateTagType(scope.row.state)">
                            {{ getStateText(scope.row.state) }}
                        </el-tag>
                    </template>
                </el-table-column>
            </el-table>
            <div class="detail-pagination-container" v-if="detailTotal > detailPageSize">
                <el-pagination
                    background
                    layout="prev, pager, next"
                    :current-page="detailCurrentPage"
                    :page-size="detailPageSize"
                    :total="detailTotal"
                    @current-change="handleDetailPageChange"
                    @size-change="handleDetailSizeChange">
                </el-pagination>
            </div>
        </el-dialog>

        <!-- 授权同步登录对话框 -->
        <el-dialog
            v-model="syncAuthDialogVisible"
            :title="$t('instance.authSync')"
            width="400px"
        >
            <el-form :model="syncAuthForm" label-width="80px">
                <el-form-item :label="$t('instance.username')" required>
                    <el-input
                        v-model="syncAuthForm.username"
                        :placeholder="$t('instance.enterUsernameOrPhone')"
                        autocomplete="off"
                    ></el-input>
                </el-form-item>
                <el-form-item :label="$t('instance.password')" required>
                    <el-input
                        v-model="syncAuthForm.password"
                        type="password"
                        :placeholder="$t('instance.enterPassword')"
                        show-password
                    ></el-input>
                </el-form-item>
                <el-form-item>
                    <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">
                        <el-checkbox v-model="syncAuthForm.saveCredentials">{{ $t('instance.rememberCredentials') }}</el-checkbox>
                        <el-link type="primary" :underline="false" @click="openForgotPasswordDialog" style="font-size: 13px;">{{ $t('instance.forgotPassword') }}</el-link>
                    </div>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleSyncAuthCancel">{{ $t('instance.cancel') }}</el-button>
                    <el-button type="primary" @click="handleSyncAuthSubmit" :loading="syncAuthLoading">
                        {{ syncAuthLoading ? $t('instance.loggingIn') : $t('instance.login') }}
                    </el-button>
                    <el-button type="success" @click="openRegisterDialog">{{ $t('instance.register') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <!-- 忘记密码对话框 -->
        <el-dialog
            v-model="forgotPasswordDialogVisible"
            title="忘记密码"
            width="400px"
            @close="handleForgotPasswordClose"
        >
            <el-form :model="forgotPasswordForm" label-width="0">
                <el-form-item>
                    <el-input
                        v-model="forgotPasswordForm.phone"
                        :placeholder="$t('instance.phoneNumber')"
                        autocomplete="off"
                        :prefix-icon="Cellphone"
                        clearable
                    ></el-input>
                    <div v-if="forgotPasswordErrors.phone" class="fp-error">
                        <el-icon style="margin-right:3px;"><WarningFilled /></el-icon>{{ forgotPasswordErrors.phone }}
                    </div>
                </el-form-item>
                <el-form-item>
                    <el-input
                        v-model="forgotPasswordForm.newPassword"
                        type="password"
                        :placeholder="$t('instance.newPassword')"
                        show-password
                        :prefix-icon="LockIcon"
                        clearable
                    ></el-input>
                    <div v-if="forgotPasswordErrors.newPassword" class="fp-error">
                        <el-icon style="margin-right:3px;"><WarningFilled /></el-icon>{{ forgotPasswordErrors.newPassword }}
                    </div>
                </el-form-item>
                <el-form-item>
                    <el-input
                        v-model="forgotPasswordForm.confirmPassword"
                        type="password"
                        :placeholder="$t('instance.confirmNewPassword')"
                        show-password
                        :prefix-icon="LockIcon"
                        clearable
                    ></el-input>
                    <div v-if="forgotPasswordErrors.confirmPassword" class="fp-error">
                        <el-icon style="margin-right:3px;"><WarningFilled /></el-icon>{{ forgotPasswordErrors.confirmPassword }}
                    </div>
                </el-form-item>
                <el-form-item>
                    <div style="display: flex; gap: 10px; width: 100%;">
                        <el-input
                            v-model="forgotPasswordForm.vcode"
                            :placeholder="$t('instance.phoneVerificationCode')"
                            style="flex: 1;"
                            clearable
                        ></el-input>
                        <el-button
                            type="primary"
                            @click="sendForgotPasswordVcode"
                            :loading="fpVcodeLoading"
                            :disabled="fpIsCountingDown"
                            style="white-space: nowrap;"
                        >
                            {{ fpVcodeButtonText }}
                        </el-button>
                    </div>
                </el-form-item>
            </el-form>
            <template #footer>
                <div class="fp-footer">
                    <el-button
                        type="primary"
                        style="width: 100%; margin-bottom: 10px;"
                        :loading="forgotPasswordLoading"
                        @click="handleForgotPasswordSubmit"
                    >
                        {{ forgotPasswordLoading ? $t('instance.resettingPassword') : $t('instance.resetPassword') }}
                    </el-button>
                    <div style="text-align: center; font-size: 13px; color: #666;">
                        {{ $t('instance.noAccount') }}<el-link type="primary" :underline="false" @click="openRegisterFromForgot">{{ $t('instance.registerNow') }}</el-link>
                    </div>
                </div>
            </template>
        </el-dialog>

        <!-- 注册对话框 -->
        <el-dialog
            v-model="registerDialogVisible"
            :title="$t('instance.userRegistration')"
            width="400px"
            @close="handleRegisterCancel"
        >
            <el-form :model="registerForm" label-width="100px">
                <el-form-item :label="$t('instance.phone')" required>
                    <el-input
                        v-model="registerForm.phone"
                        :placeholder="$t('instance.enterPhone')"
                        autocomplete="off"
                    ></el-input>
                </el-form-item>
                <el-form-item :label="$t('instance.loginPassword')" required>
                    <el-input
                        v-model="registerForm.password"
                        type="password"
                        :placeholder="$t('instance.enterPassword')"
                        show-password
                    ></el-input>
                </el-form-item>
                <el-form-item :label="$t('instance.confirmPassword')" required>
                    <el-input
                        v-model="registerForm.confirmPassword"
                        type="password"
                        :placeholder="$t('instance.enterPasswordAgain')"
                        show-password
                    ></el-input>
                </el-form-item>
                <el-form-item :label="$t('instance.verificationCode')" required>
                    <div style="display: flex; gap: 10px;">
                        <el-input
                            v-model="registerForm.vcode"
                            :placeholder="$t('instance.enterVerificationCode')"
                            style="flex: 1;"
                        ></el-input>
                        <el-button 
                            @click="sendVcode" 
                            :loading="sendVcodeLoading"
                            :disabled="isCountingDown"
                        >
                            {{ vcodeButtonText }}
                        </el-button>
                    </div>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleRegisterCancel">{{ $t('instance.cancel') }}</el-button>
                    <el-button type="primary" @click="handleRegisterSubmit" :loading="registerLoading">
                        {{ registerLoading ? $t('instance.registering') : $t('instance.register') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Loading, Cellphone, Lock as LockIcon, WarningFilled } from '@element-plus/icons-vue'
import { GetUserRabbetListWithToken, GetPackage, CreateOrder, QueryOrderStatus, GetPhoneVCode, Register } from '../../bindings/edgeclient/app'
import CryptoJS from 'crypto-js'

const props = defineProps({
    devices: {
        type: Array,
        default: () => []
    },
    token: {
        type: String,
        default: null
    }
})

const isLoggedIn = computed(() => {
    // 同时检查 props.token 和 localStorage，确保登录状态及时更新
    return !!props.token || !!localStorage.getItem('token')
})

const serverIP = computed(() => {
    return window.location.hostname || 'localhost'
})


const token = computed(() => props.token)

// 授权同步登录对话框
const syncAuthDialogVisible = ref(false)
const syncAuthForm = ref({
  username: '',
  password: '',
  saveCredentials: false
})
const syncAuthLoading = ref(false)

// 注册对话框
const registerDialogVisible = ref(false)
const registerForm = ref({
  phone: '',
  password: '',
  confirmPassword: '',
  vcode: '',
  vkey: ''
})
const registerLoading = ref(false)
const sendVcodeLoading = ref(false)
const vcodeCountdown = ref(0)
const vcodeTimer = ref(null)

// 忘记密码对话框
const forgotPasswordDialogVisible = ref(false)
const forgotPasswordForm = ref({
  phone: '',
  newPassword: '',
  confirmPassword: '',
  vcode: '',
  vkey: ''
})
const forgotPasswordErrors = ref({
  phone: '',
  newPassword: '',
  confirmPassword: ''
})
const forgotPasswordLoading = ref(false)
const fpVcodeLoading = ref(false)
const fpVcodeCountdown = ref(0)
const fpVcodeTimer = ref(null)

// 套餐相关数据
const packageDialogVisible = ref(false)
const packageList = ref([])
const selectedPackage = ref(null)
const loadingPackages = ref(false)
const selectedInstances = ref([])

// 二维码相关数据
const qrcodeDialogVisible = ref(false)
const qrcodeImage = ref('')

// 详情弹窗相关数据
const detailDialogVisible = ref(false)
const detailInstanceList = ref([])
const detailAllInstanceList = ref([])
const currentDetailHost = ref(null)
const detailCurrentPage = ref(1)
const detailPageSize = ref(12)
const detailTotal = ref(0)

const paginatedDetailList = computed(() => {
    const start = (detailCurrentPage.value - 1) * detailPageSize.value
    const end = start + detailPageSize.value
    return detailAllInstanceList.value.slice(start, end)
})

const handleDetailPageChange = (page) => {
    detailCurrentPage.value = page
}

const handleDetailSizeChange = (size) => {
    detailPageSize.value = size
    detailCurrentPage.value = 1
}

// 支付轮询相关数据
const paymentPollingTimer = ref(null)
const currentOrderId = ref('')
const currentOrderToken = ref('')

const instanceTableRef = ref(null)

// 实例列表
const instanceList = ref([])
const searchHostRabbet = ref('')
const searchState = ref('')

const isClientSideMode = computed(() => {
    const searchTerm = searchHostRabbet.value.trim()
    const isIP = searchTerm && isIPSearch(searchTerm)
    const hasStateFilter = searchState.value !== '' && searchState.value !== null && searchState.value !== undefined
    return isIP || hasStateFilter
})

// 过滤显示的子实例
const getDisplayChildren = (row) => {
    if (!row.child) return {}
    if (searchState.value === '' || searchState.value === null || searchState.value === undefined) {
        return row.child
    }
    const filtered = {}
    Object.entries(row.child).forEach(([key, item]) => {
        if (item.state === searchState.value) {
            filtered[key] = item
        }
    })
    return filtered
}

// 判断是否有显示的子实例
const hasDisplayChildren = (row) => {
    const children = getDisplayChildren(row)
    return children && Object.keys(children).length > 0
}

const filteredInstanceList = computed(() => {
    if (searchState.value === '' || searchState.value === null || searchState.value === undefined) {
        return instanceList.value
    }
    return instanceList.value.filter(row => hasDisplayChildren(row))
})

const pagedInstanceList = computed(() => {
    if (isClientSideMode.value) {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return filteredInstanceList.value.slice(start, end)
    }
    return filteredInstanceList.value
})

const displayTotal = computed(() => {
    if (isClientSideMode.value) {
        return filteredInstanceList.value.length
    }
    return total.value
})
// const filteredInstanceList = ref([])

// 分页相关数据
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 根据主机名模糊过滤实例列表
// const filterInstanceListByRabbet = () => {
//     if (!searchHostRabbet.value || !searchHostRabbet.value.trim()) {
//         filteredInstanceList.value = [...instanceList.value]
//     } else {
//         const keyword = searchHostRabbet.value.trim().toLowerCase()
        
//         filteredInstanceList.value = instanceList.value.filter(item => {
//             const match = item.rabbet && item.rabbet.toLowerCase().includes(keyword)
//             return match
//         })
//     }
// }

// 填充空实例位
const fillEmptySlots = (list) => {
    list.forEach(row => {
        const slotCount = row.rabbet && row.rabbet.charAt(0).toLowerCase() === 'p' ? 24 : 12
        
        if (!row.child) {
            row.child = {}
        }
        
        for (let i = 1; i <= slotCount; i++) {
            if (!row.child[i]) {
                row.child[i] = {
                    selected: false,
                    state: -1,
                    extime: '',
                    isEmpty: true
                }
            } else {
                row.child[i].selected = false
            }
        }
    })
}

// 根据状态获取样式类名
const getStateDotClass = (state) => {
    const stateClassMap = {
        0: 'status-normal',
        1: 'status-warning',
        2: 'status-expired'
    }
    return stateClassMap[state] || 'status-normal'
}

// 根据状态获取文字颜色类名
const getStateTextClass = (state) => {
    const stateTextClassMap = {
        0: 'text-normal',
        1: 'text-warning',
        2: 'text-expired'
    }
    return stateTextClassMap[state] || 'text-normal'
}

// 根据状态获取标签类型
const getStateTagType = (state) => {
    const stateTagTypeMap = {
        0: 'success',
        1: 'warning',
        2: 'danger'
    }
    return stateTagTypeMap[state] || 'info'
}

const getStateText = (state) => {
    const stateTextMap = {
        0: '正常',
        1: '即将到期',
        2: '已过期'
    }
    return stateTextMap[state] || '无实例'
}

// 页码变化
const handleCurrentChange = (page) => {
    currentPage.value = page
    console.log('页码变化:', page)
    fetchInstances()
}

// 每页条数变化
const handleSizeChange = (size) => {
    console.log('每页条数变化:', size)
    pageSize.value = size
    currentPage.value = 1
    fetchInstances()
}

// 获取选中实例
const getSelectedChildren = (row) => {
    if (!row.child) return []
    return Object.entries(row.child)
        .filter(([key, item]) => item.selected)
        .map(([key, item]) => ({ key, ...item }))
}


// 处理子项选中
const handleChildSelect = (row, key, item) => {
    console.log('选中实例:', row.rabbet, '实例-' + key, item.selected)
}

// 查看详情
const handleBatchOperate = (row) => {
    console.log('查看详情:', row)
    currentDetailHost.value = row.rabbet
    detailAllInstanceList.value = []
    
    if (row.child && Object.keys(row.child).length > 0) {
        Object.entries(row.child).forEach(([key, item]) => {
            detailAllInstanceList.value.push({
                slot: key,
                extime: item.extime || '无',
                state: item.state || 0,
                extimeState: item.extimeState || 0
            })
        })
    }
    
    detailTotal.value = detailAllInstanceList.value.length
    detailCurrentPage.value = 1
    detailDialogVisible.value = true
    console.log('查看详情:', row)
}

// 搜索
const handleSearch = () => {
    currentPage.value = 1
    selectedInstances.value = []
    selectedPackage.value = null
    
    if (instanceTableRef.value) {
        instanceTableRef.value.clearSelection()
    }
    
    fetchInstances()
}

// 获取所有选中的实例
const getAllSelectedInstances = () => {
    const selected = []
    
    // 如果是客户端模式，只从筛选后的列表中收集选中项
    // 实际上，即使用户之前选中了其他页的数据，或者其他筛选条件下的数据
    // 在点击购买时，通常只处理当前视图下“可见且选中”的数据，或者符合当前筛选条件的数据
    // 但如果用户希望跨筛选购买，那逻辑会很复杂。
    // 鉴于目前逻辑是：筛选改变会重置列表（fetchInstances），所以以前的选中状态会丢失。
    // 因此，我们只需要遍历当前有效的 instanceList 即可。
    // 但是，为了确保只购买符合当前筛选条件的实例，我们需要再次检查 state
    
    instanceList.value.forEach(row => {
        if (row.child) {
            Object.entries(row.child).forEach(([key, item]) => {
                // 必须选中，且如果存在状态筛选，必须符合状态
                // 注意：getDisplayChildren 逻辑是：如果无筛选，返回所有；有筛选，返回符合的。
                // 所以我们可以利用 hasDisplayChildren 或者再次判断
                
                // 简单判断：如果当前有筛选条件，那么必须符合筛选条件才算被“有效选中”
                const isMatchState = (searchState.value === '' || searchState.value === null || searchState.value === undefined) 
                                     || item.state === searchState.value
                
                if (item.selected && isMatchState) {
                    selected.push({
                        hostId: row.id,
                        hostRabbet: row.rabbet,
                        instanceKey: key,
                        ...item
                    })
                }
            })
        }
    })
    return selected
}

// 批量购买
const handleBatchBuy = async () => {
    const selected = getAllSelectedInstances()
    if (selected.length === 0) {
        ElMessage.warning('请先选择要购买/续费的实例')
        return
    }
    
    selectedInstances.value = selected
    packageDialogVisible.value = true
    selectedPackage.value = null
    
    await fetchPackages()
}

// 取消选中单个实例
const handleDeselectInstance = (index) => {
    selectedInstances.value.splice(index, 1)
    if (selectedInstances.value.length === 0) {
        packageDialogVisible.value = false
        ElMessage.warning('请先选择要购买/续费的实例')
    }
}

// 获取套餐列表
const fetchPackages = async () => {
    loadingPackages.value = true
    packageList.value = []
    
    try {
        const result = await GetPackage(token.value)
        console.log('GetPackage result:', result)
        
        // if (result.data && result.data.code == 200) {
            const data = result.data
            console.log('GetPackage data:', data)
            if (data.code == 200) {
                packageList.value = data.data.map((pkg) => ({
                    id: pkg.id,
                    name: pkg.name,
                    price: pkg.price || '0.00',
                    day: pkg.day || '0',
                }))
                
                if (packageList.value.length > 0) {
                    // ElMessage.success('获取套餐列表成功')
                } else {
                    ElMessage.warning('暂无套餐信息')
                }
            } else {
                ElMessage.error(data.msg || '获取套餐列表失败')
            }
        // }
        //  else {
        //     ElMessage.error(result.message || '获取套餐列表失败')
        // }
    } catch (error) {
        console.error('获取套餐列表失败:', error)
        ElMessage.error('获取套餐列表失败')
    } finally {
        loadingPackages.value = false
    }
}

// 确认购买
const handleConfirmBuy = async () => {
    if (!selectedPackage.value) {
        ElMessage.warning('请选择购买套餐')
        return
    }
    
    const pkg = packageList.value.find(p => p.id === selectedPackage.value)
    if (!pkg) {
        ElMessage.error('所选套餐不存在')
        return
    }
    
    const totalPrice = (parseFloat(pkg.price) * selectedInstances.value.length).toFixed(2)
    
    const instanceData = {}
    selectedInstances.value.forEach(item => {
        if (!instanceData[item.hostRabbet]) {
            instanceData[item.hostRabbet] = []
        }
        instanceData[item.hostRabbet].push(item.instanceKey)
    })
    
    const params = {
        rabbet: JSON.stringify(instanceData),
        package: pkg.id,
        money: totalPrice,
        paytype: 'zfb',
        token: token.value,
    }
    
    
    try {
        const result = await CreateOrder(params)
        console.log('CreateOrder result:', JSON.stringify(result))
        
        if (result.success) {
            // packageDialogVisible.value = false
            
            const data = result.data
            const qrcode = data && data.qrcode
            if (qrcode) {
                qrcodeImage.value = qrcode
                qrcodeDialogVisible.value = true
                const orderId = data.oid || ''
                // const orderToken = data.token || ''
                // if (orderId && orderToken) {
                    startPaymentPolling(orderId, token.value)
                // }
                // ElMessage.success('订单创建成功，请扫码支付')
            } else {
                ElMessage.success('订单创建成功')
            }
        } else {
            ElMessage.error(result.message || '创建订单失败')
            console.log('订单详情:', result)
        }
    } catch (error) {
        console.error('创建订单失败:', error)
        ElMessage.error('创建订单失败')
    }
}

// 查询支付状态
const queryPaymentStatus = async () => {
    try {
        const result = await QueryOrderStatus({
            id: currentOrderId.value,
            token: currentOrderToken.value
        })
        console.log('支付状态查询结果:', result)
        
        if (result.code == 200 && result.state == 2) {
            clearPaymentTimer()
            ElMessage.success('支付成功')
            qrcodeDialogVisible.value = false
            packageDialogVisible.value = false
            selectedInstances.value = []
            selectedPackage.value = null
            await fetchInstances()
            emit('handleSyncAuthorization')
        }
    } catch (error) {
        console.error('查询支付状态失败:', error)
    }
}


// 清除支付轮询定时器
const clearPaymentTimer = () => {
    if (paymentPollingTimer.value) {
        clearInterval(paymentPollingTimer.value)
        paymentPollingTimer.value = null
    }
}

// 启动支付轮询
const startPaymentPolling = (orderId, orderToken) => {
    currentOrderId.value = orderId
    currentOrderToken.value = orderToken
    
    clearPaymentTimer()
    
    paymentPollingTimer.value = setInterval(() => {
        queryPaymentStatus()
    }, 1000)
}

// 监听支付弹窗关闭
watch(qrcodeDialogVisible, (newVal) => {
    if (!newVal && paymentPollingTimer.value) {
        clearPaymentTimer()
        ElMessage.info('已取消支付')
    }
})

// 判断是否为IP地址搜索（支持完整IP或部分IP段模糊匹配）
const isIPSearch = (str) => {
    const trimmed = str.trim()
    if (!trimmed) return false
    
    // 检查是否包含数字和点号，且没有空格等非法字符
    const hasDigits = /\d/.test(trimmed)
    const hasDots = trimmed.includes('.')
    const noSpaces = !trimmed.includes(' ')
    const noLetters = !/[a-zA-Z]/.test(trimmed)
    
    // 数字和点的组合才认为是IP搜索
    return hasDigits && hasDots && noSpaces && noLetters
}

// 根据IP过滤实例列表
const filterInstancesByIP = (instances, ip) => {
    const searchIP = ip.trim()
    return instances.filter(item => {
        const device = props.devices.find(d => d.id == item.rabbet)
        console.log('Matching device for rabbet:', item.rabbet, 'device:', device)
        return device && device.ip && device.ip.includes(searchIP)
    })
}

// 获取实例列表
const fetchInstances = async () => {
    try {
        // 直接从 localStorage 获取最新的 token，避免 props 更新延迟
        const currentToken = localStorage.getItem('token') || props.token
        if (!currentToken) {
            ElMessage.warning('请登录后查看')
            return
        }
        const searchTerm = searchHostRabbet.value.trim()
        const isIPMode = searchTerm && isIPSearch(searchTerm)
        const hasStateFilter = searchState.value !== '' && searchState.value !== null && searchState.value !== undefined
        const fetchAll = isIPMode || hasStateFilter
        
        // 如果是IP搜索或状态过滤，获取所有数据进行客户端过滤
        const page = fetchAll ? 1 : currentPage.value
        const size = fetchAll ? 1000 : pageSize.value // IP搜索或状态过滤时获取更多数据
        
        const result = await GetUserRabbetListWithToken(
            currentToken,
            page,
            size,
            isIPMode ? '' : searchTerm // IP搜索时不传搜索参数
        )

        console.log('GetUserRabbetListWithToken result:', result)

        if (result.data && result.data.code == 0) {
            let dataList = result.data.data || []
            fillEmptySlots(dataList)

            // 将有IP的实例排在前面
            dataList.sort((a, b) => {
                const getIP = (id) => props.devices.find(d => d.id == id)?.ip
                const ipA = getIP(a.rabbet)
                const ipB = getIP(b.rabbet)
                
                if (ipA && !ipB) return -1
                if (!ipA && ipB) return 1
                return 0
            })
            
            // 如果是IP搜索，在客户端进行过滤
            if (isIPMode) {
                const filteredList = filterInstancesByIP(dataList, searchTerm)
                instanceList.value = filteredList
                total.value = filteredList.length
                if (filteredList.length === 0) {
                    ElMessage.warning('未找到匹配的实例')
                }
            } else {
                instanceList.value = dataList
                total.value = result.data.count || result.data.total || dataList.length
            }
        } else {
            instanceList.value = []
            total.value = 0
            ElMessage.error(result.message || '获取实例列表失败')
        }
    } catch (error) {
        instanceList.value = []
        total.value = 0
        ElMessage.error('获取实例列表失败')
    }
}

const handleTableSelect = (selection, row) => {
    if (!row.child) return
    
    const isSelected = selection.includes(row)
    
    // 只操作显示的子实例
    const displayChildren = getDisplayChildren(row)
    
    Object.keys(displayChildren).forEach(key => {
        if (row.child[key]) {
            row.child[key].selected = isSelected
        }
    })
}

const handleTableSelectAll = (selection) => {
    const selectedRabbets = new Set(selection.map(item => item.rabbet))
    
    const targetList = pagedInstanceList.value
    
    targetList.forEach(row => {
        if (row.child) {
            const displayChildren = getDisplayChildren(row)
            
            Object.keys(displayChildren).forEach(key => {
                if (row.child[key]) {
                    row.child[key].selected = selectedRabbets.has(row.rabbet)
                }
            })
        }
    })
}


const formName = (id) => {
   return props.devices.find(device => device.id == id)?.ip || id
}


// 定义事件
const emit = defineEmits([
  'handleSyncAuthorization',
  'update:token',
  'update-user-info'
]);

// ========== 登录/注册相关功能 ==========

// 显示授权同步对话框
const showSyncAuthDialog = () => {
  const syncAuthDeviceStr = localStorage.getItem('syncAuthCredentials')
  if (syncAuthDeviceStr) {
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

// 处理授权同步提交（登录）
const handleSyncAuthSubmit = async () => {
  if (!syncAuthForm.value.username) {
    ElMessage.warning('请输入用户名/手机号')
    return
  }
  if (!syncAuthForm.value.password) {
    ElMessage.warning('请输入密码')
    return
  }

  syncAuthLoading.value = true
  try {
    const params = {
      username: syncAuthForm.value.username,
      password: CryptoJS.MD5(syncAuthForm.value.password).toString(),
      _ts: new Date().getTime(),
    }

    const sortedKeys = Object.keys(params).sort()
    const paramStr = sortedKeys.map(k => `${params[k]}`).join('#') + '#' + '454&*&*fsdff'
    const sign = CryptoJS.MD5(paramStr).toString()
    const formData = new URLSearchParams()
    formData.append('type', 'login')

    const data = {
      uname: params.username,
      pwd: params.password,
      _ts: params._ts,
      _sign: sign
    }
    formData.append('data', JSON.stringify(data))

    const response = await fetch('https://moyunteng.com/api/sp_api.php', {
      method: 'POST',
      body: formData
    })
    const result = await response.json()
    console.log('登录响应:', result)
    
    if (result.code == 200) {
      ElMessage.success('登录成功')
      const newToken = result.data.token
      const uname = result.data.uname
      const uid = result.data.uid
      
      localStorage.setItem('token', newToken)
      localStorage.setItem('uname', uname)
      localStorage.setItem('uid', uid)
      
      // 通知父组件更新 token
      emit('update:token', newToken)
      emit('update-user-info', { token: newToken, uname, uid })

      if (syncAuthForm.value.saveCredentials) {
        localStorage.setItem('syncAuthCredentials', JSON.stringify({
          username: syncAuthForm.value.username,
          password: syncAuthForm.value.password,
          saveCredentials: syncAuthForm.value.saveCredentials
        }))
      } else {
        // 未勾选记住凭证，清除旧的保存凭证
        localStorage.removeItem('syncAuthCredentials')
      }
      
      syncAuthDialogVisible.value = false
      
      // 登录成功后刷新实例列表
      fetchInstances()
    } else {
      ElMessage.error(result.msg || '登录失败')
    }
  } catch (error) {
    console.error('登录失败:', error)
    ElMessage.error('登录失败: ' + error.message)
  } finally {
    syncAuthLoading.value = false
  }
}

// 处理授权同步取消
const handleSyncAuthCancel = () => {
  syncAuthDialogVisible.value = false
}

// 打开注册对话框
const openRegisterDialog = () => {
  syncAuthDialogVisible.value = false
  registerDialogVisible.value = true
  registerForm.value = {
    phone: '',
    password: '',
    confirmPassword: '',
    vcode: '',
    vkey: ''
  }
}

// 发送验证码
const sendVcode = async () => {
  if (!registerForm.value.phone) {
    ElMessage.warning('请输入手机号')
    return
  }

  const phoneReg = /^1[3-9]\d{9}$/
  if (!phoneReg.test(registerForm.value.phone)) {
    ElMessage.warning('请输入正确的手机号')
    return
  }

  sendVcodeLoading.value = true
  try {
    const result = await GetPhoneVCode(registerForm.value.phone, token.value || '')
    console.log('获取验证码结果:', result)
    // API返回code=0表示成功
    if (result.code == 0) {
      registerForm.value.vkey = result.data.vkey
      ElMessage.success('验证码已发送')
      
      // 开始倒计时
      vcodeCountdown.value = 60
      vcodeTimer.value = setInterval(() => {
        vcodeCountdown.value--
        if (vcodeCountdown.value <= 0) {
          clearInterval(vcodeTimer.value)
          vcodeTimer.value = null
        }
      }, 1000)
    } else {
      ElMessage.error(result.msg || '发送验证码失败')
    }
  } catch (error) {
    console.error('发送验证码失败:', error)
    ElMessage.error('发送验证码失败: ' + error.message)
  } finally {
    sendVcodeLoading.value = false
  }
}

// 处理注册提交
const handleRegisterSubmit = async () => {
  if (!registerForm.value.phone) {
    ElMessage.warning('请输入手机号')
    return
  }
  if (!registerForm.value.password) {
    ElMessage.warning('请输入密码')
    return
  }
  if (registerForm.value.password !== registerForm.value.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  if (!registerForm.value.vcode) {
    ElMessage.warning('请输入验证码')
    return
  }

  registerLoading.value = true
  try {
    const result = await Register({
      phone: registerForm.value.phone,
      password: registerForm.value.password,
      vcode: registerForm.value.vcode,
      vkey: registerForm.value.vkey
    })
    
    console.log('注册结果:', result)
    // API返回code=0表示成功
    if (result.code == 0) {
      ElMessage.success('注册成功，请登录')
      registerDialogVisible.value = false
      
      // 打开登录对话框并填充手机号
      syncAuthForm.value.username = registerForm.value.phone
      syncAuthForm.value.password = registerForm.value.password
      syncAuthDialogVisible.value = true
    } else {
      ElMessage.error(result.msg || '注册失败')
    }
  } catch (error) {
    console.error('注册失败:', error)
    ElMessage.error('注册失败: ' + error.message)
  } finally {
    registerLoading.value = false
  }
}

// 处理注册取消
const handleRegisterCancel = () => {
  registerDialogVisible.value = false
  if (vcodeTimer.value) {
    clearInterval(vcodeTimer.value)
    vcodeTimer.value = null
    vcodeCountdown.value = 0
  }
}

// 计算验证码按钮文字
const vcodeButtonText = computed(() => {
  if (vcodeCountdown.value > 0) {
    return `${vcodeCountdown.value}s后重新发送`
  }
  return '发送验证码'
})

// 是否正在倒计时
const isCountingDown = computed(() => vcodeCountdown.value > 0)

// 忘记密码：打开弹窗
const openForgotPasswordDialog = () => {
  forgotPasswordForm.value = { phone: '', newPassword: '', confirmPassword: '', vcode: '', vkey: '' }
  forgotPasswordErrors.value = { phone: '', newPassword: '', confirmPassword: '' }
  fpVcodeCountdown.value = 0
  if (fpVcodeTimer.value) {
    clearInterval(fpVcodeTimer.value)
    fpVcodeTimer.value = null
  }
  forgotPasswordDialogVisible.value = true
}

// 忘记密码：关闭弹窗
const handleForgotPasswordClose = () => {
  forgotPasswordDialogVisible.value = false
  if (fpVcodeTimer.value) {
    clearInterval(fpVcodeTimer.value)
    fpVcodeTimer.value = null
    fpVcodeCountdown.value = 0
  }
}

// 忘记密码：获取验证码
const sendForgotPasswordVcode = async () => {
  forgotPasswordErrors.value.phone = ''
  if (!forgotPasswordForm.value.phone) {
    forgotPasswordErrors.value.phone = '手机号码不能为空'
    return
  }
  const phoneReg = /^1[3-9]\d{9}$/
  if (!phoneReg.test(forgotPasswordForm.value.phone)) {
    forgotPasswordErrors.value.phone = '请输入正确的手机号'
    return
  }

  fpVcodeLoading.value = true
  try {
    const result = await GetPhoneVCode(forgotPasswordForm.value.phone, token.value || '')
    if (result.code == 0 || result.code == 200) {
      forgotPasswordForm.value.vkey = result.data && result.data.vkey ? result.data.vkey : ''
      ElMessage.success('验证码已发送')
      fpVcodeCountdown.value = 60
      fpVcodeTimer.value = setInterval(() => {
        fpVcodeCountdown.value--
        if (fpVcodeCountdown.value <= 0) {
          clearInterval(fpVcodeTimer.value)
          fpVcodeTimer.value = null
        }
      }, 1000)
    } else {
      ElMessage.error(result.msg || '发送验证码失败')
    }
  } catch (error) {
    console.error('发送验证码失败:', error)
    ElMessage.error('发送验证码失败: ' + error.message)
  } finally {
    fpVcodeLoading.value = false
  }
}

// 忘记密码：验证码按钮文字
const fpVcodeButtonText = computed(() => {
  if (fpVcodeCountdown.value > 0) {
    return `${fpVcodeCountdown.value}s后重试`
  }
  return '获取验证码'
})

// 忘记密码：是否倒计时中
const fpIsCountingDown = computed(() => fpVcodeCountdown.value > 0)

// 忘记密码：提交重置
const handleForgotPasswordSubmit = async () => {
  // 清空错误
  forgotPasswordErrors.value = { phone: '', newPassword: '', confirmPassword: '' }
  let hasError = false

  if (!forgotPasswordForm.value.phone) {
    forgotPasswordErrors.value.phone = '手机号码不能为空'
    hasError = true
  } else {
    const phoneReg = /^1[3-9]\d{9}$/
    if (!phoneReg.test(forgotPasswordForm.value.phone)) {
      forgotPasswordErrors.value.phone = '请输入正确的手机号'
      hasError = true
    }
  }
  if (!forgotPasswordForm.value.newPassword) {
    forgotPasswordErrors.value.newPassword = '新密码不能为空'
    hasError = true
  }
  if (!forgotPasswordForm.value.confirmPassword) {
    forgotPasswordErrors.value.confirmPassword = '确认新密码不能为空'
    hasError = true
  } else if (forgotPasswordForm.value.newPassword !== forgotPasswordForm.value.confirmPassword) {
    forgotPasswordErrors.value.confirmPassword = '两次输入的密码不一致'
    hasError = true
  }
  if (!forgotPasswordForm.value.vcode) {
    ElMessage.warning('请输入验证码')
    hasError = true
  }
  if (hasError) return

  forgotPasswordLoading.value = true
  try {
    const resp = await fetch('https://www.moyunteng.com/api/api.php', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        type: 'find_pwd',
        data: JSON.stringify({
          uname: forgotPasswordForm.value.phone,
          pwd: CryptoJS.MD5(forgotPasswordForm.value.newPassword).toString(),
          vcode: forgotPasswordForm.value.vcode,
          vkey: forgotPasswordForm.value.vkey
        })
      })
    })
    const result = await resp.json()
    if (result.code == 0 || result.code == 200) {
      ElMessage.success('密码重置成功，请重新登录')
      handleForgotPasswordClose()
      syncAuthDialogVisible.value = true
    } else {
      ElMessage.error(result.msg || '密码重置失败')
    }
  } catch (error) {
    console.error('密码重置失败:', error)
    ElMessage.error('密码重置失败: ' + error.message)
  } finally {
    forgotPasswordLoading.value = false
  }
}

// 忘记密码弹窗内跳转到注册
const openRegisterFromForgot = () => {
  handleForgotPasswordClose()
  openRegisterDialog()
}




// 暴露方法给父组件
defineExpose({
    fetchInstances
})
</script>

<style scoped>
.fp-error {
    display: flex;
    align-items: center;
    color: #f56c6c;
    font-size: 12px;
    margin-top: 4px;
    line-height: 1.4;
}

.fp-footer {
    width: 100%;
    display: flex;
    flex-direction: column;
}

.instance-management-container {
    /* padding: 20px; */
    height: 100%;
    box-sizing: border-box;
}


.instance-content {
    height: 100%;
}

.login-required {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
}

.instance-list {
    padding: 20px;
    height: calc(100% - 24px);
    overflow: hidden;
}

.search-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
}

.search-input {
    width: 300px;
}

.search-button {
    min-width: 80px;
}

.status-legend {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-left: 20px;
    padding-left: 20px;
    border-left: 1px solid #e4e7ed;
}

.legend-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #606266;
}

.status-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    display: inline-block;
}

.status-normal {
    background-color: #409EFF;
}

.status-warning {
    background-color: #E6A23C;
}

.status-expired {
    background-color: #F56C6C;
}

.text-normal {
    color: #409EFF;
}

.text-warning {
    color: #E6A23C;
}

.text-expired {
    color: #F56C6C;
}

.text-empty {
    color: #C0C4CC;
}

.detail-pagination-container {
    display: flex;
    justify-content: center;
    margin-top: 16px;
    padding: 12px 0;
}

.pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
    padding: 12px 0;
}

.host-name {
    font-weight: bold;
    color: #303133;
}

.child-container {
    display: flex;
    flex-wrap: wrap;
    gap: 1px;
    /* padding: 4px 0; */
    /* max-width: 800px; */
}

.child-item {
    display: flex;
    align-items: center;
    min-width: 60px;
}

.child-checkbox {
    display: flex;
    align-items: center;
    width: 100%;
}

.child-name {
    font-weight: 500;
    /* color: #606266; */
    /* margin-right: 8px; */
    min-width: 60px;
}

.child-state {
    color: #E6A23C;
    font-weight: 500;
    margin-right: 8px;
}

.child-time {
    color: #909399;
    font-size: 12px;
}

.child-status {
    margin-left: 8px;
}

.no-child {
    color: #909399;
    font-size: 13px;
}

.child-item .el-checkbox__label {
    padding-left: 4px !important;
}

.package-list {
    max-height: 400px;
    overflow-y: auto;
}

.package-item {
    margin-bottom: 12px;
    cursor: pointer;
}

.package-card {
    transition: all 0.3s;
    /* border: 2px solid transparent; */
}

.package-card:hover {
    border-color: #409EFF;
}

.package-selected {
    border-color: #409EFF;
    background-color: #ecf5ff;
}

.package-header {
    font-weight: bold;
    font-size: 16px;
    margin-bottom: 8px;
}

.package-content {
    padding-left: 24px;
}

.package-price {
    font-size: 18px;
    color: #F56C6C;
    font-weight: bold;
    margin-bottom: 4px;
}

.package-duration {
    font-size: 14px;
    color: #606266;
    margin-bottom: 8px;
}

.package-desc {
    font-size: 12px;
    color: #909399;
}

.loading-container {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 40px;
    color: #606266;
}

.empty-container {
    padding: 40px;
}

.selected-instances-section {
    margin-bottom: 16px;
}

.section-title {
    font-size: 14px;
    font-weight: bold;
    color: #606266;
    margin-bottom: 12px;
}

.selected-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.instance-tag {
    margin-right: 0;
}

.dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
}

.dialog-title {
    color: #303133;
    font-size: 18px;
    font-weight: bold;
}

.price-unit {
    font-size: 12px;
    color: #909399;
}

.total-section {
    background-color: #f5f7fa;
    border-radius: 4px;
    padding: 16px;
    margin-top: 16px;
}

.total-row {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
}

.total-label {
    font-size: 16px;
    font-weight: bold;
    color: #303133;
}

.total-amount {
    font-size: 24px;
    font-weight: bold;
    color: #F56C6C;
}

.total-instances {
    font-size: 14px;
    color: #909399;
}

.instance-tag {
    margin-right: 0;
    cursor: pointer;
}

.instance-tag:hover {
    background-color: #fcd3d3;
    border-color: #f56c6c;
    color: #f56c6c;
}

.qrcode-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 20px;
}

.qrcode-image {
    width: 250px;
    height: 250px;
}

.qrcode-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    color: #909399;
}

.qrcode-tip {
    text-align: center;
    font-size: 14px;
    color: #606266;
    margin-top: 16px;
    padding-bottom: 10px;
}

.el-radio-group {
    display: flex;
    justify-content: space-between;
}

:deep(.el-table .cell) {
  display: block;
}
.pagination-container {
  /* display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  padding: 0 20px; */
  position: absolute;
  bottom: -26px;
  right: 20px;
}
</style>