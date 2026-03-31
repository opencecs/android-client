import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {resolve} from 'path'
import {copyFileSync, mkdirSync, existsSync} from 'fs'
import os from 'os'

// https://vitejs.dev/config/
export default defineConfig(({mode}) => {
  const isProduction = mode === 'production'
  const isDevelopment = mode === 'development'
  const cpuCount = os.cpus().length
  
  console.log(`\n📦 构建模式: ${mode} (生产环境: ${isProduction}, CPU核心: ${cpuCount})\n`)
  
  return {
    base: './',  // 使用相对路径，而不是绝对路径
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src')
      }
    },
    server: {
      host: '0.0.0.0',
      port: 9245,
      strictPort: true,
      cors: true
    },
    publicDir: 'webplayer',
    build: {
      assetsDir: 'assets',
      copyPublicDir: true,
      emptyOutDir: true,
      
      // 启用生产环境压缩（esbuild，更快）
      minify: 'esbuild',

      // 开启多线程并行构建
      rollupOptions: {
        maxParallelFileOps: Math.max(cpuCount, 20),
        output: {
          // 禁用手动代码分割，让 Vite 自动处理（避免模块加载顺序问题）
          manualChunks: undefined,
          // 生产环境使用 hash 命名
          chunkFileNames: isProduction ? 'assets/js/[name]-[hash].js' : 'assets/[name].js',
          entryFileNames: isProduction ? 'assets/js/[name]-[hash].js' : 'assets/[name].js',
          assetFileNames: isProduction ? 'assets/[ext]/[name]-[hash].[ext]' : 'assets/[name].[ext]'
        },
        plugins: [{
          name: 'copy-play-html',
          generateBundle() {
            const srcPath = resolve(__dirname, 'webplayer/play.html')
            const destPath = resolve(__dirname, 'dist/play.html')
            
            if (!isDevelopment) {
              console.log('📄 Copying play.html to dist root...')
            }
            
            if (!existsSync(srcPath)) {
              console.error('❌ ERROR: Source play.html not found:', srcPath)
              return
            }
            
            copyFileSync(srcPath, destPath)
            
            if (!isDevelopment) {
              console.log('✅ play.html copied successfully')
            }
          }
        }]
      }
    },
    
    // 定义全局常量
    define: {
      __DEV__: !isProduction,
      __PROD__: isProduction
    }
  }
})
