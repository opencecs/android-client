<template>
  <div class="seamless-image-container" @click="handleClick">
    <img
      v-for="img in activeImages"
      :key="img.key"
      :src="img.src"
      :alt="alt"
      :class="[imgClass, 'seamless-img']"
      :style="{ 
        opacity: img.loaded ? 1 : 0, 
        zIndex: img.zIndex,
        transition: 'opacity 0.1s ease-out' 
      }"
      @load="handleImageLoad(img)"
      @error="handleError(img, $event)"
    />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  src: {
    type: String,
    default: ''
  },
  alt: {
    type: String,
    default: ''
  },
  imgClass: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['click', 'error'])

const activeImages = ref([])
let zIndexCounter = 10

watch(() => props.src, (newSrc) => {
  if (!newSrc) return

  // 为新图片创建唯一key
  const newKey = `img-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  
  // 添加新图片
  // 初始 loaded 为 false，确保不可见（透明）
  // zIndex 递增，确保新图片覆盖在旧图片之上
  activeImages.value.push({
    key: newKey,
    src: newSrc,
    loaded: false,
    zIndex: zIndexCounter++
  })

  // 安全清理：防止数组无限增长
  // 如果积累了太多未加载的图片（极端情况），或者清理逻辑未触发
  if (activeImages.value.length > 10) {
    // 强制保留最后 2 张
    activeImages.value = activeImages.value.slice(-2)
  }
}, { immediate: true })

const handleImageLoad = (img) => {
  // 图片加载完成（包括解码），设置为可见
  // 由于有 transition，opacity 会平滑过渡到 1
  img.loaded = true
  
  // 延迟清理旧图片
  // 确保新图片完全显示后再移除底下的旧图片
  setTimeout(() => {
    // 移除所有 zIndex 小于当前图片的图片（即旧图片）
    // 这样可以确保当前显示的图片始终存在
    activeImages.value = activeImages.value.filter(item => item.zIndex >= img.zIndex)
  }, 150) // 略大于 transition 时间，确保过渡完成
}

const handleClick = (event) => {
  emit('click', event)
}

const handleError = (img, event) => {
  // 加载失败也标记为 loaded，避免一直不可见
  img.loaded = true
  emit('error', event)
  
  // 同样清理旧图片
  setTimeout(() => {
    activeImages.value = activeImages.value.filter(item => item.zIndex >= img.zIndex)
  }, 150)
}
</script>

<style scoped>
.seamless-image-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* 确保容器有背景色，避免图片加载前的空白 */
  background-color: #f0f2f5; 
}

.seamless-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  display: block;
  /* 确保图片不会响应鼠标事件，除非它是最上面的 */
  /* pointer-events: none; */
}
</style>
