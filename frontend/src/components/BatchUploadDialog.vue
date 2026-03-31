<template>
  <el-dialog
    v-model="dialogVisible"
    :title="'批量上传文件到云机'"
    width="800px"
    :close-on-click-modal="false"
  >
    <div class="batch-upload-container">
      <!-- 操作坑位列表 -->
      <div class="slots-section" v-if="selectedMachines && selectedMachines.length > 0">
        <div class="section-title">操作坑位：</div>
        <div class="slots-list">
          <el-tag
            v-for="(machine, index) in selectedMachines"
            :key="getMachineKey(machine, index)"
            size="small"
            class="slot-tag"
          >
            {{ getMachineDisplayName(machine, index) }}
          </el-tag>
        </div>
      </div>

      <!-- 注意事项提示 -->
      <el-alert
        title="📌 上传说明"
        type="info"
        :closable="false"
        style="margin-bottom: 16px;"
      >
        <template #default>
          <div style="line-height: 1.6;">
            <p style="margin: 4px 0;">• 上传的文件默认保存在云机的 <code style="background: #f0f0f0; padding: 2px 6px; border-radius: 3px;">/sdcard/upload/</code> 目录下</p>
            <p style="margin: 4px 0;">• <strong>APK 文件</strong>会自动安装,请耐心等待安装完成</p>
            <p style="margin: 4px 0;">• <strong>APKX 文件</strong>不会自动安装,需要手动处理</p>
          </div>
        </template>
      </el-alert>

      <!-- APK 安装选项 -->
      <div v-if="hasApkFiles" style="margin-bottom: 16px; padding: 12px; background-color: #f5f7fa; border-radius: 4px;">
        <div style="margin-bottom: 8px; font-weight: bold; font-size: 13px;">⚙️ APK 安装选项</div>
        <el-checkbox-group v-model="apkInstallOptions" style="display: flex; flex-direction: column; gap: 6px;">
          <el-checkbox label="replace" style="margin: 0; height: auto;">
            <div style="line-height: 1.3; font-size: 12px;">
              <span>替换已存在的应用</span>
              <span style="color: #909399; font-size: 10px; margin-left: 4px;">(-r 覆盖安装)</span>
            </div>
          </el-checkbox>
          <el-checkbox label="test" style="margin: 0; height: auto;">
            <div style="line-height: 1.3; font-size: 12px;">
              <span>允许安装测试应用</span>
              <span style="color: #909399; font-size: 10px; margin-left: 4px;">(-t 测试模式)</span>
            </div>
          </el-checkbox>
          <el-checkbox label="grant" style="margin: 0; height: auto;">
            <div style="line-height: 1.3; font-size: 12px;">
              <span>自动授予所有权限</span>
              <span style="color: #909399; font-size: 10px; margin-left: 4px;">(-g 自动授权)</span>
            </div>
          </el-checkbox>
          <el-checkbox label="deleteAfterInstall" style="margin: 0; height: auto;">
            <div style="line-height: 1.3; font-size: 12px;">
              <span>安装成功后删除安装包</span>
              <span style="color: #909399; font-size: 10px; margin-left: 4px;">(失败时保留)</span>
            </div>
          </el-checkbox>
        </el-checkbox-group>
      </div>

      <!-- 文件选择区域 -->
      <div class="files-section" v-loading="filesLoading" element-loading-text="加载文件中...">
        <div class="dialog-header" style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
          <h3 style="margin: 0;">文件列表</h3>
          <div style="display: flex; gap: 8px; align-items: center;">
            <el-select v-model="sortType" size="small" style="width: 120px;" @change="handleSortChange">
              <el-option label="按时间排序" value="time"></el-option>
              <el-option label="按名称排序" value="name"></el-option>
            </el-select>
            <el-button type="primary" size="small" @click="openSharedDirectory">
              打开共享目录
            </el-button>
          </div>
        </div>

        <!-- 共享目录路径设置 -->
        <div style="margin-bottom: 12px; padding: 10px 12px; background: #f5f7fa; border-radius: 4px; border: 1px solid #e4e7ed;">
          <div style="font-size: 12px; color: #606266; margin-bottom: 8px; font-weight: 500;">📂 文件来源目录</div>
          <div style="display: flex; align-items: center; gap: 8px;">
            <el-input
              v-model="sharedDirPathInfo.path"
              placeholder="共享目录路径"
              size="small"
              :readonly="true"
              style="flex: 1;"
            />
            <el-button size="small" type="primary" :loading="sharedDirLoading" @click="handleSelectSharedDir">
              浏览
            </el-button>
          </div>
          <div style="display: flex; align-items: center; gap: 8px; margin-top: 6px;">
            <el-tag v-if="sharedDirPathInfo.isDefault" type="info" size="small">默认目录</el-tag>
            <el-tag v-else type="success" size="small">自定义目录</el-tag>
            <el-button
              v-if="!sharedDirPathInfo.isDefault"
              type="text"
              size="small"
              style="color: #909399; padding: 0;"
              :loading="sharedDirLoading"
              @click="handleResetSharedDirPath"
            >
              恢复默认
            </el-button>
          </div>
        </div>
        <div v-if="fileTree" class="file-tree">
          <!-- 递归渲染目录树 -->
          <div class="tree-node">
            <div class="node-checkbox">
              <el-checkbox
                :model-value="isDirectoryFullySelected(fileTree)"
                :indeterminate="isDirectoryPartiallySelected(fileTree)"
                @change="() => handleNodeSelectionChange(fileTree)"
              ></el-checkbox>
            </div>
            <span class="node-icon" v-if="fileTree.isDir && fileTree.children && fileTree.children.length > 0" @click.stop="toggleNodeExpanded(fileTree)">
              {{ fileTree.expanded ? '▼' : '▶' }}
            </span>
            <span class="node-icon" v-else-if="fileTree.isDir">
              📁
            </span>
            <span class="node-name" @click="toggleNodeExpanded(fileTree)">{{ fileTree.name }}</span>
          </div>
          <!-- 递归渲染子节点 -->
          <template v-if="fileTree.expanded && fileTree.isDir && fileTree.children && fileTree.children.length > 0">
            <div class="tree-children">
              <div v-for="child in fileTree.children" :key="child.path">
                <div class="tree-node">
                  <div class="node-checkbox" v-if="child.isDir">
                    <el-checkbox
                      :model-value="isDirectoryFullySelected(child)"
                      :indeterminate="isDirectoryPartiallySelected(child)"
                      @change="() => handleNodeSelectionChange(child)"
                    ></el-checkbox>
                  </div>
                  <div class="node-checkbox" v-else>
                    <el-checkbox
                      :model-value="selectedFiles.includes(child.path)"
                      @change="(val) => handleFileCheckChange(child.path, val)"
                    >{{ child.name }}</el-checkbox>
                  </div>
                  <span class="node-icon" v-if="child.isDir && child.children && child.children.length > 0" @click.stop="toggleNodeExpanded(child)">
                    {{ child.expanded ? '▼' : '▶' }}
                  </span>
                  <span class="node-icon" v-else-if="child.isDir">
                    📁
                  </span>
                  <span class="node-name" v-if="child.isDir" @click="toggleNodeExpanded(child)">{{ child.name }}</span>
                  <span class="node-size" v-if="!child.isDir">{{ (child.size / 1024).toFixed(2) }} KB</span>
                  <span class="node-date" v-if="!child.isDir">{{ new Date(child.modTime * 1000).toLocaleString() }}</span>
                </div>
                <!-- 递归渲染子目录 -->
                <template v-if="child.expanded && child.isDir && child.children && child.children.length > 0">
                  <div class="tree-children">
                    <div v-for="grandchild in child.children" :key="grandchild.path">
                      <div class="tree-node">
                        <div class="node-checkbox" v-if="grandchild.isDir">
                          <el-checkbox
                            :model-value="isDirectoryFullySelected(grandchild)"
                            :indeterminate="isDirectoryPartiallySelected(grandchild)"
                            @change="() => handleNodeSelectionChange(grandchild)"
                          ></el-checkbox>
                        </div>
                        <div class="node-checkbox" v-else>
                          <el-checkbox
                            :model-value="selectedFiles.includes(grandchild.path)"
                            @change="(val) => handleFileCheckChange(grandchild.path, val)"
                          >{{ grandchild.name }}</el-checkbox>
                        </div>
                        <span class="node-icon" v-if="grandchild.isDir && grandchild.children && grandchild.children.length > 0" @click.stop="toggleNodeExpanded(grandchild)">
                          {{ grandchild.expanded ? '▼' : '▶' }}
                        </span>
                        <span class="node-icon" v-else-if="grandchild.isDir">
                          📁
                        </span>
                        <span class="node-name" v-if="grandchild.isDir" @click="toggleNodeExpanded(grandchild)">{{ grandchild.name }}</span>
                        <span class="node-size" v-if="!grandchild.isDir">{{ (grandchild.size / 1024).toFixed(2) }} KB</span>
                        <span class="node-date" v-if="!grandchild.isDir">{{ new Date(grandchild.modTime * 1000).toLocaleString() }}</span>
                      </div>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </template>
        </div>
        <div v-else class="no-files">
          <el-empty description="共享目录中没有文件"></el-empty>
          <el-button type="primary" style="margin-top: 16px;" @click="openSharedDirectory">
            打开共享目录
          </el-button>
        </div>
      </div>

      <!-- 上传进度 -->
      <div class="upload-progress" v-if="uploading">
        <el-progress
          :percentage="uploadProgress"
          :status="uploadSuccess ? 'success' : ''"
          :stroke-width="10"
        ></el-progress>
        <div class="upload-status">{{ uploadStatus }}</div>
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleUpload" 
          :loading="uploading"
          :disabled="selectedFiles.length === 0"
        >
          {{ uploading ? '上传中...' : '上传' }}
        </el-button>
        <el-button 
          type="primary" 
          @click="handleRefresh" 
        >
         刷新共享文件 
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { ListSharedDirFiles, GetSharedDirPath, SetSharedDirPath, SelectDirectory } from '../../bindings/edgeclient/app.js';

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  selectedMachines: {
    type: Array,
    default: () => []
  },
  cloudManageMode: {
    type: String,
    default: 'slot'
  },
  selectedCloudDevice: {
    type: Object,
    default: null
  }
});

const emit = defineEmits(['update:visible', 'upload', 'openSharedDirectory']);

const dialogVisible = ref(false);
const fileTree = ref(null);
const selectedFiles = ref([]);
const filesLoading = ref(false);
const uploading = ref(false);
const uploadProgress = ref(0);
const uploadStatus = ref('');
const uploadSuccess = ref(false);
const sortType = ref('time');
const apkInstallOptions = ref(['replace', 'test', 'grant', 'deleteAfterInstall']); // 默认选中替换、测试、授权、安装后删除

// 共享目录路径
const sharedDirPathInfo = ref({ path: '', isDefault: true, defaultPath: '' });
const sharedDirLoading = ref(false);

const loadSharedDirPath = async () => {
  try {
    const result = await GetSharedDirPath();
    if (result) {
      sharedDirPathInfo.value = {
        path: result.path || '',
        isDefault: result.isDefault !== false,
        defaultPath: result.defaultPath || result.path || '',
      };
    }
  } catch (e) {
    console.error('获取共享目录路径失败', e);
  }
};

const handleSelectSharedDir = async () => {
  sharedDirLoading.value = true;
  try {
    const result = await SelectDirectory();
    if (result && result.success && result.path) {
      // 保存路径
      const saveResult = await SetSharedDirPath(result.path);
      if (saveResult && saveResult.success) {
        sharedDirPathInfo.value.path = result.path;
        sharedDirPathInfo.value.isDefault = false;
        ElMessage.success('共享目录路径已保存');
        // 重新加载文件列表
        await loadSharedFiles();
      } else {
        ElMessage.error(saveResult?.message || '保存路径失败');
      }
    } else if (result && !result.success && result.message !== '用户取消选择') {
      ElMessage.error(result.message || '选择目录失败');
    }
  } catch (e) {
    ElMessage.error('选择目录失败');
  } finally {
    sharedDirLoading.value = false;
  }
};

const handleResetSharedDirPath = async () => {
  sharedDirLoading.value = true;
  try {
    const result = await SetSharedDirPath('');
    if (result && result.success) {
      ElMessage.success('已恢复默认目录');
      sharedDirPathInfo.value = {
        path: result.path || sharedDirPathInfo.value.defaultPath,
        isDefault: true,
        defaultPath: sharedDirPathInfo.value.defaultPath,
      };
      await loadSharedFiles();
    } else {
      ElMessage.error(result?.message || '恢复默认失败');
    }
  } catch (e) {
    ElMessage.error('恢复默认失败');
  } finally {
    sharedDirLoading.value = false;
  }
};

// 检测是否有 APK 文件被选中
const hasApkFiles = computed(() => {
  return selectedFiles.value.some(filePath => 
    filePath.toLowerCase().endsWith('.apk')
  );
});

watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
  if (newVal) {
    loadSharedDirPath();
    if (props.selectedMachines && props.selectedMachines.length > 0) {
      loadSharedFiles();
    }
  }
});

watch(dialogVisible, (newVal) => {
  emit('update:visible', newVal);
});

const getMachineKey = (machine, index) => {
  if (props.cloudManageMode === 'slot') {
    // 坑位模式下，machine 可能是坑位号或容器对象
    const slotNum = typeof machine === 'object' ? machine.indexNum : machine
    return `slot-${slotNum}`;
  }
  // 批量模式下，machine 是云机对象
  return machine.id || machine.name || `machine-${index}`;
};

const getMachineDisplayName = (machine, index) => {
  if (props.cloudManageMode === 'slot') {
    // 坑位模式下，machine 可能是坑位号或容器对象
    const slotNum = typeof machine === 'object' ? machine.indexNum : machine
    return `坑位 ${slotNum}`;
  }
  // 批量模式下，machine 是云机对象
  return machine.name || `云机 ${index + 1}`;
};

const sortFileTree = (node) => {
  if (node && node.children && node.children.length > 0) {
    if (sortType.value === 'time') {
      node.children.sort((a, b) => (b.modTime || 0) - (a.modTime || 0));
    } else if (sortType.value === 'name') {
      node.children.sort((a, b) => {
        const nameA = (a.name || '').toLowerCase();
        const nameB = (b.name || '').toLowerCase();
        return nameA.localeCompare(nameB);
      });
    }
    node.children.forEach(child => {
      if (child.isDir) {
        sortFileTree(child);
      }
    });
  }
};

const handleSortChange = () => {
  if (fileTree.value) {
    sortFileTree(fileTree.value);
  }
};

const loadSharedFiles = async () => {
  filesLoading.value = true;
  try {
    const result = await ListSharedDirFiles();
    if (result.success) {
      fileTree.value = result.tree;
      if (fileTree.value) {
        sortFileTree(fileTree.value);
      }
      selectedFiles.value = [];
      if (fileTree.value && fileTree.value.isDir) {
        fileTree.value.expanded = true;
      }
    } else {
      ElMessage.error(result.message || '加载共享文件失败');
      fileTree.value = null;
    }
  } catch (error) {
    console.error('加载共享文件失败:', error);
    ElMessage.error('加载共享文件失败');
    fileTree.value = null;
  } finally {
    filesLoading.value = false;
  }
};

const handleRefresh = async () => {
  filesLoading.value = true;
  try {
    const result = await ListSharedDirFiles();
    if (result.success) {
      fileTree.value = result.tree;
      if (fileTree.value) {
        sortFileTree(fileTree.value);
      }
      selectedFiles.value = [];
      if (fileTree.value && fileTree.value.isDir) {
        fileTree.value.expanded = true;
      }
      ElMessage.success('刷新共享文件成功');
    } else {
      ElMessage.error(result.message || '加载共享文件失败');
      fileTree.value = null;
    }
  } catch (error) {
    console.error('加载共享文件失败:', error);
    ElMessage.error('加载共享文件失败');
    fileTree.value = null;
  } finally {
    filesLoading.value = false;
  }
}

const toggleNodeExpanded = (node) => {
  if (node.isDir) {
    node.expanded = !node.expanded;
  }
};

const handleFileCheckChange = (path, checked) => {
  if (checked) {
    if (!selectedFiles.value.includes(path)) {
      selectedFiles.value = [...selectedFiles.value, path];
    }
  } else {
    selectedFiles.value = selectedFiles.value.filter(p => p !== path);
  }
};

const collectFilePaths = (node) => {
  const filePaths = [];
  const collect = (n) => {
    if (n.children) {
      n.children.forEach(child => {
        if (!child.isDir) {
          filePaths.push(child.path);
        } else {
          collect(child);
        }
      });
    }
  };
  collect(node);
  return filePaths;
};

const isDirectoryFullySelected = (node) => {
  if (!node || !node.isDir) return false;
  if (!node.children || node.children.length === 0) return false;
  
  const filePaths = collectFilePaths(node);
  if (filePaths.length === 0) return false;
  
  return filePaths.every(filePath => selectedFiles.value.includes(filePath));
};

const isDirectoryPartiallySelected = (node) => {
  if (!node || !node.isDir) return false;
  if (!node.children || node.children.length === 0) return false;
  
  const filePaths = collectFilePaths(node);
  if (filePaths.length === 0) return false;
  
  const allSelected = filePaths.every(filePath => selectedFiles.value.includes(filePath));
  if (allSelected) return false;
  
  const someSelected = filePaths.some(filePath => selectedFiles.value.includes(filePath));
  return someSelected;
};

const handleNodeSelectionChange = (node) => {
  if (!node || !node.isDir || !node.children) return;
  
  const filePaths = collectFilePaths(node);
  const currentState = isDirectoryFullySelected(node);
  const shouldCheck = !currentState;
  
  if (shouldCheck) {
    selectedFiles.value = [...new Set([...selectedFiles.value, ...filePaths])];
  } else {
    selectedFiles.value = selectedFiles.value.filter(filePath => !filePaths.includes(filePath));
  }
};

const openSharedDirectory = () => {
  emit('openSharedDirectory');
};

const handleUpload = async () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请选择要上传的文件');
    return;
  }
  
  if (!props.selectedMachines || props.selectedMachines.length === 0) {
    ElMessage.warning('没有选中的云机');
    return;
  }
  
  uploading.value = true;
  uploadProgress.value = 0;
  uploadStatus.value = '准备上传...';
  uploadSuccess.value = false;
  
  try {
    // 构建 APK 安装选项
    const apkOptions = {
      replace: apkInstallOptions.value.includes('replace'),
      test: apkInstallOptions.value.includes('test'),
      grant: apkInstallOptions.value.includes('grant'),
      deleteAfterInstall: apkInstallOptions.value.includes('deleteAfterInstall')
    };
    
    emit('upload', {
      files: selectedFiles.value,
      machines: props.selectedMachines,
      cloudManageMode: props.cloudManageMode,
      selectedCloudDevice: props.selectedCloudDevice,
      apkOptions, // 传递 APK 安装选项
      onProgress: (progress, status) => {
        uploadProgress.value = progress;
        uploadStatus.value = status;
      }
    });
    
    uploadSuccess.value = true;
    uploadStatus.value = '上传完成';
    ElMessage.success('文件上传任务已开始');
  } catch (error) {
    console.error('上传失败:', error);
    ElMessage.error(`上传失败: ${error.message || '未知错误'}`);
    uploadSuccess.value = false;
    uploadStatus.value = '上传失败';
  } finally {
    uploading.value = false;
  }
};

const handleClose = () => {
  if (uploading.value) {
    ElMessage.warning('上传中，无法关闭');
    return;
  }
  dialogVisible.value = false;
  selectedFiles.value = [];
  fileTree.value = null;
};

defineExpose({
  setUploadProgress: (progress, status) => {
    uploadProgress.value = progress;
    uploadStatus.value = status;
  },
  setUploadSuccess: (success) => {
    uploadSuccess.value = success;
    uploading.value = false;
    if (success) {
      uploadStatus.value = '上传完成';
    }
  },
  reset: () => {
    selectedFiles.value = [];
    fileTree.value = null;
    uploading.value = false;
    uploadProgress.value = 0;
    uploadStatus.value = '';
    uploadSuccess.value = false;
  }
});
</script>

<style scoped>
.batch-upload-container {
  overflow: visible;
}

.slots-section {
  margin-bottom: 20px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.section-title {
  font-weight: 500;
  margin-bottom: 8px;
  color: #606266;
}

.slots-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.slot-tag {
  margin-right: 0;
}

.files-section {
  min-height: 200px;
}

.file-tree {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 12px;
  max-height: 360px;
  overflow-y: auto;
}

.tree-node {
  display: flex;
  align-items: center;
  padding: 4px 0;
}

.tree-children {
  padding-left: 20px;
}

.node-checkbox {
  margin-right: 8px;
}

.node-icon {
  margin-right: 8px;
  cursor: pointer;
  user-select: none;
}

.node-name {
  cursor: pointer;
  flex: 1;
}

.node-size {
  width: 80px;
  text-align: right;
  color: #909399;
  font-size: 12px;
}

.node-date {
  width: 150px;
  text-align: right;
  color: #909399;
  font-size: 12px;
  margin-left: 12px;
}

.no-files {
  text-align: center;
  padding: 40px 0;
}

.upload-progress {
  margin-top: 20px;
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.upload-status {
  margin-top: 8px;
  text-align: center;
  color: #606266;
  font-size: 14px;
}
</style>
