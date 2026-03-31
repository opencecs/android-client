<template>
  <div class="screenshot-container" @click="handleClick">
    <template v-if="screenshotData">
      <!-- 有截图数据时直接显示 -->
      <img 
        :src="screenshotData" 
        alt="云机截图" 
        class="screenshot-img"
        :class="{ 'screenshot-img-rotated': rotate }"
      />
    </template>
    <template v-else>
      <!-- 暂无截图，显示等待状态 -->
      <div class="screenshot-empty">
        <el-icon><VideoCamera /></el-icon>
        <div style="text-align: center;">
          <div>等待截图...</div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { VideoCamera } from '@element-plus/icons-vue'

const props = defineProps({
  screenshotData: {
    type: String,
    default: '' // base64 DataURL，由父组件从后端缓存传入
  },
  deviceKey: {
    type: String,
    required: true // 用于标识设备，切换时父组件会更新 screenshotData
  },
  rotate: {
    type: Boolean,
    default: false // 是否旋转图片（横屏模式）
  }
})

const emit = defineEmits(['click'])

const handleClick = () => {
  emit('click')
}
</script>

<style scoped>
.screenshot-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.screenshot-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  display: block;
}

/* 横屏旋转模式 - 使用实测的正确值 */
.screenshot-img-rotated {
  position: absolute;
  top: 50%;
  left: 50%;
  
  /* 经过实测得到的最佳值 */
  width: 56%; /* 图片宽度为容器高度的56%，旋转后变成视觉高度 */
  height: 179%; /* 图片高度为容器高度的179%，旋转后变成视觉宽度 */
  
  transform: translate(-50%, -50%) rotate(-90deg);
  transform-origin: center center;
  object-fit: cover;
  max-width: none;
}

.screenshot-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: #909399;
}
</style>
