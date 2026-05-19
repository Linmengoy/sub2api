<template>
  <div class="midjourney-container">
    <!-- 左侧导航 -->
    <div class="sidebar">
      <div class="logo">
        🎨 Midjourney AI
      </div>
      <nav class="nav-menu">
        <router-link 
          to="/generate" 
          class="nav-item"
          :class="{ active: $route.path === '/generate' }"
        >
          <span class="nav-icon">✨</span>
          {{ $t('image.navGenerate') }}
        </router-link>
        <router-link 
          to="/edit" 
          class="nav-item"
          :class="{ active: $route.path === '/edit' }"
        >
          <span class="nav-icon">🎨</span>
          {{ $t('image.navEdit') }}
        </router-link>
        <div class="nav-divider"></div>
        <div class="nav-item history">
          <span class="nav-icon">📜</span>
          {{ $t('image.history') }}
        </div>
      </nav>
      
      <div class="sidebar-footer">
        <div class="user-info">
          <div class="avatar">👤</div>
          <div class="user-details">
            <div class="user-name">User</div>
            <div class="user-credits">💎 Credits</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 顶部栏 -->
      <div class="top-bar">
        <div class="tab-switch">
          <button 
            class="tab"
            :class="{ active: mode === 'edit' }"
            @click="mode = 'edit'"
          >
            {{ $t('image.editMode') }}
          </button>
          <button 
            class="tab"
            :class="{ active: mode === 'vary' }"
            @click="mode = 'vary'"
          >
            {{ $t('image.varyMode') }}
          </button>
        </div>
        <div class="top-actions">
          <button class="icon-btn">🔔</button>
          <button class="icon-btn">⚙️</button>
        </div>
      </div>

      <!-- 编辑模式 -->
      <div v-if="mode === 'edit'" class="workspace">
        <div class="upload-section" v-if="!selectedImage">
          <div 
            class="upload-zone"
            :class="{ dragging: isDragging }"
            @dragover.prevent="isDragging = true"
            @dragleave="isDragging = false"
            @drop.prevent="handleDrop"
            @click="triggerFileInput"
          >
            <div class="upload-icon">📤</div>
            <div class="upload-text">{{ $t('image.uploadPrompt') }}</div>
            <div class="upload-hint">{{ $t('image.uploadHint') }}</div>
            <input 
              ref="fileInputRef"
              type="file" 
              accept="image/*"
              @change="handleFileSelect"
              style="display: none"
            />
          </div>
        </div>

        <div v-else class="editor-area">
          <!-- 图片预览 -->
          <div class="canvas-container">
            <div class="canvas-wrapper">
              <img :src="selectedImage" class="source-image" />
              <canvas 
                ref="maskCanvasRef"
                class="mask-canvas"
                @mousedown="startDrawing"
                @mousemove="draw"
                @mouseup="stopDrawing"
                @mouseleave="stopDrawing"
                @touchstart="handleTouchStart"
                @touchmove="handleTouchMove"
                @touchend="stopDrawing"
              />
            </div>
          </div>

          <!-- 右侧工具面板 -->
          <div class="tools-panel">
            <!-- 画笔工具 -->
            <div class="tool-section">
              <div class="section-title">{{ $t('image.brushTools') }}</div>
              <div class="tool-buttons">
                <button 
                  class="tool-btn"
                  :class="{ active: brushTool === 'mask' }"
                  @click="brushTool = 'mask'"
                >
                  🖌️ {{ $t('image.maskBrush') }}
                </button>
                <button 
                  class="tool-btn"
                  :class="{ active: brushTool === 'eraser' }"
                  @click="brushTool = 'eraser'"
                >
                  🧹 {{ $t('image.eraser') }}
                </button>
              </div>
              
              <div class="brush-size">
                <label>{{ $t('image.brushSize') }}</label>
                <input 
                  type="range" 
                  v-model.number="brushSize"
                  min="5"
                  max="100"
                />
                <span class="size-value">{{ brushSize }}px</span>
              </div>

              <button class="clear-btn" @click="clearMask">
                {{ $t('image.clearMask') }}
              </button>
            </div>

            <!-- 参数设置 -->
            <div class="tool-section">
              <div class="section-title">{{ $t('image.settings') }}</div>
              
              <div class="param-item">
                <label>{{ $t('image.model') }}</label>
                <select v-model="selectedModel">
                  <option value="dall-e-2">DALL-E 2</option>
                  <option value="dall-e-3">DALL-E 3</option>
                </select>
              </div>

              <div class="param-item">
                <label>{{ $t('image.size') }}</label>
                <select v-model="size">
                  <option value="1024x1024">1024×1024</option>
                  <option value="1792x1024">1792×1024</option>
                  <option value="1024x1792">1024×1792</option>
                </select>
              </div>

              <div class="param-item">
                <label>{{ $t('image.variations') }}</label>
                <select v-model="n">
                  <option :value="1">1</option>
                  <option :value="2">2</option>
                  <option :value="4">4</option>
                </select>
              </div>
            </div>

            <!-- 提示词 -->
            <div class="prompt-section">
              <div class="section-title">{{ $t('image.editPrompt') }}</div>
              <textarea 
                v-model="prompt"
                :placeholder="$t('image.editPromptPlaceholder')"
                rows="4"
              />
            </div>

            <!-- 生成按钮 -->
            <button 
              class="generate-btn"
              :disabled="!prompt || processing"
              @click="handleEdit"
            >
              <span v-if="processing">
                <LoadingSpinner size="sm" /> {{ $t('image.processing') }}
              </span>
              <span v-else>
                ✨ {{ $t('image.generateEdit') }}
              </span>
            </button>
          </div>
        </div>

        <!-- 结果展示 -->
        <div v-if="results.length > 0" class="results-section">
          <div class="results-header">
            <h3>{{ $t('image.results') }}</h3>
            <button class="clear-btn" @click="results = []">{{ $t('image.clearResults') }}</button>
          </div>
          <div class="image-grid">
            <div 
              v-for="(img, index) in results" 
              :key="index"
              class="result-card"
            >
              <img :src="img.url" class="result-image" />
              <div class="result-actions">
                <button class="action-btn" @click="downloadImage(img.url)">📥</button>
                <button class="action-btn" @click="useAsSource(img.url)">↩️</button>
                <button class="action-btn" @click="createVariation(img.url)">🔄</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 变体模式 -->
      <div v-else class="workspace">
        <div class="upload-section" v-if="!selectedImage">
          <div 
            class="upload-zone"
            :class="{ dragging: isDragging }"
            @dragover.prevent="isDragging = true"
            @dragleave="isDragging = false"
            @drop.prevent="handleDrop"
            @click="triggerFileInput"
          >
            <div class="upload-icon">📤</div>
            <div class="upload-text">{{ $t('image.uploadPrompt') }}</div>
            <div class="upload-hint">{{ $t('image.varyHint') }}</div>
            <input 
              ref="fileInputRef"
              type="file" 
              accept="image/*"
              @change="handleFileSelect"
              style="display: none"
            />
          </div>
        </div>

        <div v-else class="vary-area">
          <div class="source-preview">
            <img :src="selectedImage" class="source-preview-image" />
            <button class="change-btn" @click="clearSelection">{{ $t('image.changeImage') }}</button>
          </div>

          <div class="vary-tools">
            <div class="vary-options">
              <button 
                class="vary-btn primary"
                :disabled="processing"
                @click="handleVary('strong')"
              >
                🎯 {{ $t('image.varyStrong') }}
              </button>
              <button 
                class="vary-btn"
                :disabled="processing"
                @click="handleVary('subtle')"
              >
                🔍 {{ $t('image.varySubtle') }}
              </button>
            </div>

            <div class="vary-settings">
              <div class="param-item">
                <label>{{ $t('image.size') }}</label>
                <select v-model="size">
                  <option value="1024x1024">1024×1024</option>
                  <option value="1792x1024">1792×1024</option>
                  <option value="1024x1792">1024×1792</option>
                </select>
              </div>

              <div class="param-item">
                <label>{{ $t('image.variations') }}</label>
                <select v-model="n">
                  <option :value="1">1</option>
                  <option :value="2">2</option>
                  <option :value="4">4</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- 结果展示 -->
        <div v-if="results.length > 0" class="results-section">
          <div class="results-header">
            <h3>{{ $t('image.results') }}</h3>
            <button class="clear-btn" @click="results = []">{{ $t('image.clearResults') }}</button>
          </div>
          <div class="image-grid">
            <div 
              v-for="(img, index) in results" 
              :key="index"
              class="result-card"
            >
              <img :src="img.url" class="result-image" />
              <div class="result-actions">
                <button class="action-btn" @click="downloadImage(img.url)">📥</button>
                <button class="action-btn" @click="useAsSource(img.url)">↩️</button>
                <button class="action-btn" @click="createVariation(img.url)">🔄</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { imagesAPI } from '@/api/images'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const { t } = useI18n()

// 状态
const mode = ref<'edit' | 'vary'>('edit')
const selectedImage = ref<string | null>(null)
const isDragging = ref(false)
const brushTool = ref<'mask' | 'eraser'>('mask')
const brushSize = ref(30)
const selectedModel = ref('dall-e-2')
const size = ref('1024x1024')
const n = ref(1)
const prompt = ref('')
const processing = ref(false)
const results = ref<Array<{ url: string }>>([])

// Refs
const fileInputRef = ref<HTMLInputElement | null>(null)
const maskCanvasRef = ref<HTMLCanvasElement | null>(null)
const isDrawing = ref(false)
const lastPosition = ref<{ x: number; y: number } | null>(null)
const maskData = ref<string | null>(null)

// 处理文件选择
const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    loadImage(file)
  }
}

const handleDrop = (e: DragEvent) => {
  isDragging.value = false
  const file = e.dataTransfer?.files[0]
  if (file && file.type.startsWith('image/')) {
    loadImage(file)
  }
}

const loadImage = (file: File) => {
  const reader = new FileReader()
  reader.onload = async (e) => {
    selectedImage.value = e.target?.result as string
    await nextTick()
    initCanvas()
  }
  reader.readAsDataURL(file)
}

const clearSelection = () => {
  selectedImage.value = null
  maskData.value = null
  results.value = []
  prompt.value = ''
}

// Canvas 初始化
const initCanvas = () => {
  const canvas = maskCanvasRef.value
  if (!canvas) return
  
  const img = new Image()
  img.onload = () => {
    canvas.width = img.width
    canvas.height = img.height
    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
    }
  }
  img.src = selectedImage.value!
}

// 绘制功能
const startDrawing = (e: MouseEvent) => {
  if (!selectedImage.value) return
  isDrawing.value = true
  const canvas = maskCanvasRef.value
  if (!canvas) return
  
  const rect = canvas.getBoundingClientRect()
  const scaleX = canvas.width / rect.width
  const scaleY = canvas.height / rect.height
  
  lastPosition.value = {
    x: (e.clientX - rect.left) * scaleX,
    y: (e.clientY - rect.top) * scaleY
  }
}

const draw = (e: MouseEvent) => {
  if (!isDrawing.value || !lastPosition.value) return
  
  const canvas = maskCanvasRef.value
  if (!canvas) return
  
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  const rect = canvas.getBoundingClientRect()
  const scaleX = canvas.width / rect.width
  const scaleY = canvas.height / rect.height
  
  const x = (e.clientX - rect.left) * scaleX
  const y = (e.clientY - rect.top) * scaleY
  
  ctx.beginPath()
  ctx.moveTo(lastPosition.value.x, lastPosition.value.y)
  ctx.lineTo(x, y)
  ctx.strokeStyle = brushTool.value === 'mask' ? 'rgba(0, 0, 0, 1)' : 'rgba(255, 255, 255, 1)'
  ctx.lineWidth = brushSize.value
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.stroke()
  
  lastPosition.value = { x, y }
}

const stopDrawing = () => {
  isDrawing.value = false
  lastPosition.value = null
  
  const canvas = maskCanvasRef.value
  if (canvas) {
    maskData.value = canvas.toDataURL('image/png')
  }
}

const handleTouchStart = (e: TouchEvent) => {
  e.preventDefault()
  const touch = e.touches[0]
  const mouseEvent = new MouseEvent('mousedown', {
    clientX: touch.clientX,
    clientY: touch.clientY
  })
  startDrawing(mouseEvent as any)
}

const handleTouchMove = (e: TouchEvent) => {
  e.preventDefault()
  const touch = e.touches[0]
  const mouseEvent = new MouseEvent('mousemove', {
    clientX: touch.clientX,
    clientY: touch.clientY
  })
  draw(mouseEvent as any)
}

const clearMask = () => {
  const canvas = maskCanvasRef.value
  if (!canvas) return
  
  const ctx = canvas.getContext('2d')
  if (ctx) {
    ctx.clearRect(0, 0, canvas.width, canvas.height)
  }
  maskData.value = null
}

// 生成功能
const handleEdit = async () => {
  if (!selectedImage.value || !prompt.value) return
  
  processing.value = true
  try {
    const response = await imagesAPI.edit({
      model: selectedModel.value,
      image: selectedImage.value,
      mask: maskData.value || undefined,
      prompt: prompt.value,
      n: n.value,
      size: size.value as any
    })
    
    results.value = response.data.map(item => ({
      url: item.url || (item.b64_json ? `data:image/png;base64,${item.b64_json}` : '')
    })).filter(img => img.url)
  } catch (error) {
    console.error('Edit failed:', error)
    alert(t('image.editFailed'))
  } finally {
    processing.value = false
  }
}

const handleVary = async (type: 'strong' | 'subtle') => {
  if (!selectedImage.value) return
  
  processing.value = true
  try {
    const response = await imagesAPI.createVariation({
      model: selectedModel.value,
      image: selectedImage.value,
      n: n.value,
      size: size.value as any
    })
    
    results.value = response.data.map(item => ({
      url: item.url || (item.b64_json ? `data:image/png;base64,${item.b64_json}` : '')
    })).filter(img => img.url)
  } catch (error) {
    console.error('Variation failed:', error)
    alert(t('image.varyFailed'))
  } finally {
    processing.value = false
  }
}

const createVariation = (url: string) => {
  selectedImage.value = url
  mode.value = 'vary'
  results.value = []
}

const useAsSource = (url: string) => {
  selectedImage.value = url
  mode.value = 'edit'
  results.value = []
  initCanvas()
}

const downloadImage = (url: string) => {
  const link = document.createElement('a')
  link.href = url
  link.download = `generated-image-${Date.now()}.png`
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

onMounted(() => {
  // 初始化
})
</script>

<style scoped>
.midjourney-container {
  display: flex;
  min-height: 100vh;
  background: #1a1a1a;
  color: #fff;
}

/* 左侧导航 */
.sidebar {
  width: 260px;
  background: #0f172a;
  border-right: 1px solid #1e293b;
  display: flex;
  flex-direction: column;
}

.logo {
  padding: 24px 20px;
  font-size: 1.25rem;
  font-weight: 700;
  color: #fff;
  border-bottom: 1px solid #2a2a2a;
}

.nav-menu {
  flex: 1;
  padding: 20px 12px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  color: #a0a0a0;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.nav-item:hover {
  background: #1e293b;
  color: #fff;
}

.nav-item.active {
  background: linear-gradient(135deg, #14b8a6, #0d9488);
  color: #fff;
}

.nav-icon {
  font-size: 1.25rem;
}

.nav-divider {
  height: 1px;
  background: #2a2a2a;
  margin: 16px 0;
}

.history {
  color: #a0a0a0;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #2a2a2a;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #1e293b;
  border-radius: 8px;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #2a2a2a;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.user-name {
  font-weight: 600;
  font-size: 0.875rem;
}

.user-credits {
  font-size: 0.75rem;
  color: #a0a0a0;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.top-bar {
  height: 64px;
  background: #0f172a;
  border-bottom: 1px solid #1e293b;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.tab-switch {
  display: flex;
  gap: 8px;
}

.tab {
  padding: 8px 20px;
  background: transparent;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
  font-weight: 500;
}

.tab:hover {
  border-color: #475569;
  color: #fff;
}

.tab.active {
  background: #334155;
  border-color: #334155;
  color: #fff;
}

.top-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  width: 40px;
  height: 40px;
  border: none;
  background: transparent;
  color: #a0a0a0;
  cursor: pointer;
  border-radius: 8px;
  font-size: 1.125rem;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: #1e293b;
  color: #fff;
}

/* 工作区 */
.workspace {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* 上传区域 */
.upload-section {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.upload-zone {
  width: 100%;
  max-width: 600px;
  height: 400px;
  border: 2px dashed #475569;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  background: #1e293b;
}

.upload-zone:hover,
.upload-zone.dragging {
  border-color: #14b8a6;
  background: #0f172a;
}

.upload-icon {
  font-size: 4rem;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 8px;
  color: #fff;
}

.upload-hint {
  font-size: 0.875rem;
  color: #a0a0a0;
}

/* 编辑区域 */
.editor-area {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.canvas-container {
  background: #1e293b;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 500px;
}

.canvas-wrapper {
  position: relative;
  max-width: 100%;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

.source-image {
  display: block;
  max-width: 100%;
  max-height: 600px;
  width: auto;
  height: auto;
}

.mask-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  cursor: crosshair;
}

/* 工具面板 */
.tools-panel {
  background: #1e293b;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.tool-section {
  padding-bottom: 20px;
  border-bottom: 1px solid #334155;
}

.tool-section:last-of-type {
  border-bottom: none;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #a0a0a0;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tool-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 16px;
}

.tool-btn {
  padding: 10px 16px;
  background: #334155;
  border: 1px solid #475569;
  border-radius: 8px;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.tool-btn:hover {
  background: #475569;
  color: #fff;
}

.tool-btn.active {
  background: linear-gradient(135deg, #14b8a6, #0d9488);
  border-color: transparent;
  color: #fff;
}

.brush-size {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.brush-size label {
  font-size: 0.75rem;
  color: #a0a0a0;
}

.brush-size input[type="range"] {
  width: 100%;
  accent-color: #6366f1;
}

.size-value {
  font-size: 0.75rem;
  color: #a0a0a0;
  text-align: right;
}

.clear-btn {
  width: 100%;
  padding: 8px 16px;
  background: transparent;
  border: 1px solid #3a3a3a;
  border-radius: 6px;
  color: #ef4444;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
  margin-top: 12px;
}

.clear-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.param-item {
  margin-bottom: 16px;
}

.param-item label {
  display: block;
  font-size: 0.75rem;
  color: #a0a0a0;
  margin-bottom: 8px;
}

.param-item select {
  width: 100%;
  padding: 10px 12px;
  background: #334155;
  border: 1px solid #475569;
  border-radius: 8px;
  color: #fff;
  font-size: 0.875rem;
  cursor: pointer;
}

.param-item select:focus {
  outline: none;
  border-color: #14b8a6;
}

.prompt-section {
  margin-bottom: 20px;
}

.prompt-section textarea {
  width: 100%;
  padding: 12px;
  background: #334155;
  border: 1px solid #475569;
  border-radius: 8px;
  color: #fff;
  font-size: 0.875rem;
  resize: none;
}

.prompt-section textarea:focus {
  outline: none;
  border-color: #14b8a6;
}

.generate-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #14b8a6, #0d9488);
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.generate-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(20, 184, 166, 0.4);
}

.generate-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 变体模式 */
.vary-area {
  max-width: 800px;
  margin: 0 auto;
}

.source-preview {
  text-align: center;
  margin-bottom: 32px;
}

.source-preview-image {
  max-width: 100%;
  max-height: 500px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  margin-bottom: 16px;
}

.change-btn {
  padding: 8px 16px;
  background: #334155;
  border: 1px solid #475569;
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.change-btn:hover {
  background: #475569;
  color: #fff;
}

.vary-tools {
  background: #1e293b;
  border-radius: 12px;
  padding: 24px;
}

.vary-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 24px;
}

.vary-btn {
  padding: 16px 24px;
  background: #334155;
  border: 2px solid #475569;
  border-radius: 10px;
  color: #fff;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 600;
  transition: all 0.2s;
}

.vary-btn:hover:not(:disabled) {
  background: #475569;
  border-color: #64748b;
  transform: translateY(-2px);
}

.vary-btn.primary {
  background: linear-gradient(135deg, #14b8a6, #0d9488);
  border-color: transparent;
}

.vary-btn.primary:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(20, 184, 166, 0.4);
}

.vary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.vary-settings {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

/* 结果展示 */
.results-section {
  margin-top: 32px;
  max-width: 1400px;
  margin-left: auto;
  margin-right: auto;
}

.results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.results-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.result-card {
  background: #1e293b;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s;
}

.result-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.result-image {
  width: 100%;
  aspect-ratio: 1;
  object-fit: cover;
}

.result-actions {
  padding: 12px;
  display: flex;
  gap: 8px;
}

.action-btn {
  flex: 1;
  padding: 8px;
  background: #334155;
  border: none;
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #475569;
  color: #fff;
}

@media (max-width: 1024px) {
  .editor-area {
    grid-template-columns: 1fr;
  }
  
  .sidebar {
    position: fixed;
    left: -260px;
    top: 0;
    bottom: 0;
    z-index: 100;
    transition: left 0.3s;
  }
}
</style>
