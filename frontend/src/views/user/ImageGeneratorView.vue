<template>
  <div class="midjourney-generator">
    <!-- 左侧导航 -->
    <div class="sidebar">
      <div class="logo">
        🎨 Midjourney AI
      </div>
      <nav class="nav-menu">
        <div class="nav-item active">
          <span class="nav-icon">✨</span>
          {{ $t('image.navGenerate') }}
        </div>
        <router-link 
          to="/edit" 
          class="nav-item"
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
        <div class="top-title">
          <h1>✨ {{ $t('image.navGenerate') }}</h1>
          <span class="subtitle">{{ $t('image.description') }}</span>
        </div>
        <div class="top-actions">
          <button class="icon-btn">🔔</button>
          <button class="icon-btn">⚙️</button>
        </div>
      </div>

      <!-- 工作区 -->
      <div class="workspace">
        <div class="workspace-layout">
          <!-- 左侧：图片展示区 -->
          <div class="preview-area">
            <!-- 加载状态 -->
            <div v-if="generating" class="loading-overlay">
              <div class="loading-content">
                <LoadingSpinner size="lg" />
                <div class="loading-text">{{ $t('image.generating') }}</div>
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: progress + '%' }"></div>
                </div>
                <div class="job-id">Job ID: {{ currentJobId }}</div>
              </div>
            </div>

            <!-- 占位符 -->
            <div v-if="!currentImage && !generating" class="placeholder">
              <div class="placeholder-icon">🖼️</div>
              <div class="placeholder-text">{{ $t('image.placeholder') }}</div>
              <div class="placeholder-hint">{{ $t('image.enterPromptHint') }}</div>
            </div>

            <!-- 当前生成的图片 -->
            <div v-if="currentImage" class="image-container">
              <img :src="currentImage" class="generated-image" />
              
              <!-- 图片操作按钮 -->
              <div class="image-actions-overlay">
                <button class="action-btn" @click="handleU(1)" :disabled="generating">U1</button>
                <button class="action-btn" @click="handleU(2)" :disabled="generating">U2</button>
                <button class="action-btn" @click="handleU(3)" :disabled="generating">U3</button>
                <button class="action-btn" @click="handleU(4)" :disabled="generating">U4</button>
              </div>
              
              <div class="image-info">
                <div class="info-item">📏 {{ currentSize }}</div>
                <div class="info-item">🎨 {{ selectedModel }}</div>
              </div>
            </div>

            <!-- 历史图片网格 -->
            <div v-if="imageHistory.length > 0" class="history-section">
              <div class="history-header">
                <h3>{{ $t('image.recentJobs') }}</h3>
                <button class="clear-history" @click="clearHistory">
                  {{ $t('image.clearAll') }}
                </button>
              </div>
              <div class="history-grid">
                <div 
                  v-for="(img, index) in imageHistory" 
                  :key="index"
                  class="history-card"
                  @click="selectHistoryImage(img)"
                >
                  <img :src="img.url" class="history-image" />
                  <div class="history-overlay">
                    <button class="use-btn" @click.stop="useAsSource(img)">
                      ↩️ {{ $t('image.useAsSource') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 右侧：参数控制面板 -->
          <div class="control-panel">
            <!-- 提示词输入 -->
            <div class="prompt-section">
              <div class="section-header">
                <span class="section-title">{{ $t('image.prompt') }}</span>
                <button class="enhance-btn" @click="enhancePrompt">
                  ✨ {{ $t('image.enhance') }}
                </button>
              </div>
              <textarea 
                v-model="prompt"
                :placeholder="$t('image.enterPrompt')"
                rows="5"
                class="prompt-input"
              />
              <div class="prompt-actions">
                <button class="action-btn secondary" @click="clearPrompt">
                  🗑️ {{ $t('image.clear') }}
                </button>
                <button class="action-btn secondary" @click="randomPrompt">
                  🎲 {{ $t('image.random') }}
                </button>
              </div>
            </div>

            <!-- 参数网格 -->
            <div class="params-grid">
              <!-- 模型选择 -->
              <div class="param-section">
                <label class="param-label">{{ $t('image.model') }}</label>
                <div class="model-selector">
                  <button 
                    v-for="model in modelOptions" 
                    :key="model.id"
                    :class="['model-btn', { active: selectedModel === model.id }]"
                    @click="selectedModel = model.id"
                  >
                    <span class="model-icon">{{ model.icon }}</span>
                    <span class="model-name">{{ model.name }}</span>
                  </button>
                </div>
              </div>

              <!-- 宽高比 -->
              <div class="param-section">
                <label class="param-label">{{ $t('image.aspectRatio') }}</label>
                <div class="aspect-selector">
                  <button 
                    v-for="ratio in aspectRatios" 
                    :key="ratio.value"
                    :class="['aspect-btn', { active: selectedAspect === ratio.value }]"
                    @click="selectedAspect = ratio.value"
                  >
                    <div class="aspect-icon" :style="ratio.style"></div>
                    <span class="aspect-label">{{ ratio.label }}</span>
                  </button>
                </div>
              </div>

              <!-- 版本选择 -->
              <div class="param-section">
                <label class="param-label">{{ $t('image.version') }}</label>
                <select v-model="selectedVersion" class="version-select">
                  <option value="v6">v6 (最新)</option>
                  <option value="v5.2">v5.2</option>
                  <option value="v5.1">v5.1</option>
                  <option value="niji">niji</option>
                </select>
              </div>

              <!-- 质量参数 -->
              <div class="param-section">
                <label class="param-label">{{ $t('image.quality') }}</label>
                <div class="quality-selector">
                  <button 
                    :class="['quality-btn', { active: quality === '0.25' }]"
                    @click="quality = '0.25'"
                  >
                    {{ $t('image.fast') }}
                  </button>
                  <button 
                    :class="['quality-btn', { active: quality === '1' }]"
                    @click="quality = '1'"
                  >
                    {{ $t('image.standard') }}
                  </button>
                  <button 
                    :class="['quality-btn', { active: quality === '2' }]"
                    @click="quality = '2'"
                  >
                    {{ $t('image.detailed') }}
                  </button>
                </div>
              </div>

              <!-- 风格 -->
              <div class="param-section">
                <label class="param-label">{{ $t('image.style') }}</label>
                <div class="style-selector">
                  <button 
                    v-for="style in styleOptions" 
                    :key="style.id"
                    :class="['style-btn', { active: selectedStyle === style.id }]"
                    @click="selectedStyle = style.id"
                  >
                    {{ style.icon }} {{ style.name }}
                  </button>
                </div>
              </div>

              <!-- 数量 -->
              <div class="param-section">
                <label class="param-label">{{ $t('image.number') }}</label>
                <div class="number-selector">
                  <button 
                    v-for="n in [1, 2, 4]" 
                    :key="n"
                    :class="['num-btn', { active: imageCount === n }]"
                    @click="imageCount = n"
                  >
                    {{ n }}
                  </button>
                </div>
              </div>
            </div>

            <!-- 生成按钮 -->
            <button 
              class="generate-btn"
              :disabled="!prompt.trim() || generating"
              @click="handleGenerate"
            >
              <span v-if="generating">
                <LoadingSpinner size="sm" /> {{ $t('image.generating') }}
              </span>
              <span v-else>
                🚀 {{ $t('image.imagine') }}
              </span>
            </button>

            <!-- 快捷提示词 -->
            <div class="quick-prompts">
              <div class="quick-header">{{ $t('image.quickPrompts') }}</div>
              <div class="quick-list">
                <Tag 
                  v-for="q in quickPrompts" 
                  :key="q"
                  class="quick-tag"
                  @click="prompt = q"
                >
                  {{ q.substring(0, 25) }}{{ q.length > 25 ? '...' : '' }}
                </Tag>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { imagesAPI } from '@/api/images'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Tag from '@/components/common/Tag.vue'

const { t } = useI18n()

// 状态
const prompt = ref('')
const selectedModel = ref('dall-e-3')
const selectedAspect = ref('1:1')
const selectedVersion = ref('v6')
const quality = ref('1')
const selectedStyle = ref('auto')
const imageCount = ref(4)
const generating = ref(false)
const progress = ref(0)
const currentJobId = ref('')
const currentImage = ref('')
const currentSize = ref('1024x1024')
const imageHistory = ref<Array<{ url: string; prompt: string }>>([])

// 配置选项
const modelOptions = [
  { id: 'dall-e-3', name: 'DALL-E 3', icon: '🎨' },
  { id: 'dall-e-2', name: 'DALL-E 2', icon: '🖌️' },
  { id: 'gpt-image-2', name: 'GPT Image', icon: '🤖' },
]

const aspectRatios = [
  { value: '1:1', label: '1:1', style: { width: '40px', height: '40px' } },
  { value: '16:9', label: '16:9', style: { width: '53px', height: '30px' } },
  { value: '9:16', label: '9:16', style: { width: '30px', height: '53px' } },
  { value: '4:3', label: '4:3', style: { width: '53px', height: '40px' } },
  { value: '3:4', label: '3:4', style: { width: '40px', height: '53px' } },
]

const styleOptions = [
  { id: 'auto', name: '自动', icon: '🎯' },
  { id: 'vivid', name: '生动', icon: '🌈' },
  { id: 'natural', name: '自然', icon: '🌿' },
]

const quickPrompts = [
  'A beautiful sunset over the ocean',
  'Futuristic city with flying cars',
  'Cute cat wearing a space suit',
  'Abstract geometric art pattern',
  'Cozy coffee shop interior',
  'Mountain landscape at sunrise',
]

// 计算尺寸
const computedSize = computed(() => {
  const aspectMap: Record<string, string> = {
    '1:1': '1024x1024',
    '16:9': '1792x1024',
    '9:16': '1024x1792',
    '4:3': '1024x768',
    '3:4': '768x1024',
  }
  return aspectMap[selectedAspect.value] || '1024x1024'
})

// 生成图片
const handleGenerate = async () => {
  if (!prompt.value.trim() || generating.value) return

  generating.value = true
  progress.value = 0
  currentJobId.value = `job_${Date.now()}`
  
  try {
    // 模拟进度
    const progressInterval = setInterval(() => {
      if (progress.value < 90) {
        progress.value += Math.random() * 15
      }
    }, 500)

    const response = await imagesAPI.generate({
      model: selectedModel.value,
      prompt: prompt.value,
      n: imageCount.value,
      size: computedSize.value as any,
      quality: quality.value as any,
      style: selectedStyle.value as any,
    })

    clearInterval(progressInterval)
    progress.value = 100

    if (response.data && response.data.length > 0) {
      currentImage.value = response.data[0].url || ''
      currentSize.value = computedSize.value
      
      // 添加到历史
      imageHistory.value.unshift({
        url: currentImage.value,
        prompt: prompt.value,
      })
      
      // 限制历史数量
      if (imageHistory.value.length > 12) {
        imageHistory.value.pop()
      }
    }
  } catch (error) {
    console.error('Generation failed:', error)
    alert(t('image.generationFailed'))
  } finally {
    setTimeout(() => {
      generating.value = false
      progress.value = 0
    }, 500)
  }
}

// U 按钮处理
const handleU = (index: number) => {
  console.log(`Upscale image ${index}`)
  // 实现放大功能
}

// 增强提示词
const enhancePrompt = () => {
  console.log('Enhancing prompt...')
  // 调用增强提示词API
}

// 清除提示词
const clearPrompt = () => {
  prompt.value = ''
}

// 随机提示词
const randomPrompt = () => {
  const random = quickPrompts[Math.floor(Math.random() * quickPrompts.length)]
  prompt.value = random
}

// 选择历史图片
const selectHistoryImage = (img: { url: string; prompt: string }) => {
  currentImage.value = img.url
}

// 使用图片作为源
const useAsSource = (img: { url: string; prompt: string }) => {
  prompt.value = img.prompt
  currentImage.value = img.url
}

// 清除历史
const clearHistory = () => {
  imageHistory.value = []
}
</script>

<style scoped>
.midjourney-generator {
  display: flex;
  min-height: 100vh;
  background: #1a1a1a;
  color: #fff;
}

/* 侧边栏 */
.sidebar {
  width: 260px;
  background: #0f0f0f;
  border-right: 1px solid #2a2a2a;
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
  background: #2a2a2a;
  color: #fff;
}

.nav-item.active {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
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
  background: #1a1a1a;
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
  height: 72px;
  background: #0f0f0f;
  border-bottom: 1px solid #2a2a2a;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.top-title h1 {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 4px;
}

.top-title .subtitle {
  font-size: 0.75rem;
  color: #a0a0a0;
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
  background: #2a2a2a;
  color: #fff;
}

/* 工作区 */
.workspace {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.workspace-layout {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 24px;
  height: calc(100vh - 120px);
}

/* 预览区 */
.preview-area {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(26, 26, 26, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  border-radius: 12px;
}

.loading-content {
  text-align: center;
}

.loading-text {
  margin: 16px 0;
  font-size: 1.125rem;
  color: #a0a0a0;
}

.progress-bar {
  width: 300px;
  height: 4px;
  background: #2a2a2a;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366f1, #8b5cf6);
  transition: width 0.3s ease;
}

.job-id {
  font-size: 0.75rem;
  color: #666;
  font-family: monospace;
}

.placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
}

.placeholder-icon {
  font-size: 6rem;
  margin-bottom: 16px;
}

.placeholder-text {
  font-size: 1.125rem;
  margin-bottom: 8px;
}

.placeholder-hint {
  font-size: 0.875rem;
  color: #666;
}

.image-container {
  position: relative;
  margin-bottom: 24px;
}

.generated-image {
  width: 100%;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.image-actions-overlay {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.image-container:hover .image-actions-overlay {
  opacity: 1;
}

.image-actions-overlay .action-btn {
  padding: 8px 16px;
  background: rgba(0, 0, 0, 0.8);
  border: 1px solid #3a3a3a;
  border-radius: 6px;
  color: #fff;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.image-actions-overlay .action-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: transparent;
}

.image-actions-overlay .action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.image-info {
  display: flex;
  gap: 16px;
  margin-top: 12px;
}

.info-item {
  font-size: 0.875rem;
  color: #a0a0a0;
}

/* 历史区域 */
.history-section {
  margin-top: auto;
  padding-top: 24px;
  border-top: 1px solid #2a2a2a;
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.history-header h3 {
  font-size: 1rem;
  font-weight: 600;
}

.clear-history {
  padding: 4px 12px;
  background: transparent;
  border: 1px solid #3a3a3a;
  border-radius: 4px;
  color: #ef4444;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-history:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.history-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.history-card {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
}

.history-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s;
}

.history-card:hover .history-image {
  transform: scale(1.05);
}

.history-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.history-card:hover .history-overlay {
  opacity: 1;
}

.use-btn {
  padding: 6px 12px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 0.75rem;
  cursor: pointer;
}

/* 控制面板 */
.control-panel {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
}

.prompt-section {
  background: #0f0f0f;
  border-radius: 8px;
  padding: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #a0a0a0;
}

.enhance-btn {
  padding: 4px 12px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.enhance-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(99, 102, 241, 0.3);
}

.prompt-input {
  width: 100%;
  padding: 12px;
  background: #1a1a1a;
  border: 1px solid #2a2a2a;
  border-radius: 6px;
  color: #fff;
  font-size: 0.875rem;
  resize: none;
  margin-bottom: 12px;
}

.prompt-input:focus {
  outline: none;
  border-color: #6366f1;
}

.prompt-actions {
  display: flex;
  gap: 8px;
}

.prompt-actions .action-btn {
  flex: 1;
  padding: 6px 12px;
  background: #2a2a2a;
  border: 1px solid #3a3a3a;
  border-radius: 4px;
  color: #a0a0a0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.prompt-actions .action-btn:hover {
  background: #3a3a3a;
  color: #fff;
}

/* 参数网格 */
.params-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.param-section {
  background: #0f0f0f;
  border-radius: 8px;
  padding: 16px;
}

.param-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  color: #a0a0a0;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* 模型选择 */
.model-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.model-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px 8px;
  background: #1a1a1a;
  border: 2px solid #2a2a2a;
  border-radius: 8px;
  color: #a0a0a0;
  cursor: pointer;
  transition: all 0.2s;
}

.model-btn:hover {
  border-color: #3a3a3a;
  color: #fff;
}

.model-btn.active {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
  color: #fff;
}

.model-icon {
  font-size: 1.5rem;
}

.model-name {
  font-size: 0.75rem;
  font-weight: 600;
}

/* 宽高比 */
.aspect-selector {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}

.aspect-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 4px;
  background: #1a1a1a;
  border: 2px solid #2a2a2a;
  border-radius: 6px;
  color: #a0a0a0;
  cursor: pointer;
  transition: all 0.2s;
}

.aspect-btn:hover {
  border-color: #3a3a3a;
  color: #fff;
}

.aspect-btn.active {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
  color: #fff;
}

.aspect-icon {
  background: #666;
  border-radius: 4px;
}

.aspect-label {
  font-size: 0.625rem;
  font-weight: 600;
}

/* 版本选择 */
.version-select {
  width: 100%;
  padding: 10px 12px;
  background: #1a1a1a;
  border: 1px solid #2a2a2a;
  border-radius: 6px;
  color: #fff;
  font-size: 0.875rem;
  cursor: pointer;
}

.version-select:focus {
  outline: none;
  border-color: #6366f1;
}

/* 质量选择 */
.quality-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.quality-btn {
  padding: 10px 12px;
  background: #1a1a1a;
  border: 2px solid #2a2a2a;
  border-radius: 6px;
  color: #a0a0a0;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.quality-btn:hover {
  border-color: #3a3a3a;
  color: #fff;
}

.quality-btn.active {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
  color: #fff;
}

/* 风格选择 */
.style-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.style-btn {
  padding: 10px 12px;
  background: #1a1a1a;
  border: 2px solid #2a2a2a;
  border-radius: 6px;
  color: #a0a0a0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.style-btn:hover {
  border-color: #3a3a3a;
  color: #fff;
}

.style-btn.active {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
  color: #fff;
}

/* 数量选择 */
.number-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.num-btn {
  padding: 10px 12px;
  background: #1a1a1a;
  border: 2px solid #2a2a2a;
  border-radius: 6px;
  color: #a0a0a0;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.num-btn:hover {
  border-color: #3a3a3a;
  color: #fff;
}

.num-btn.active {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
  color: #fff;
}

/* 生成按钮 */
.generate-btn {
  width: 100%;
  padding: 16px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.generate-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(99, 102, 241, 0.4);
}

.generate-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 快捷提示词 */
.quick-prompts {
  background: #0f0f0f;
  border-radius: 8px;
  padding: 16px;
}

.quick-header {
  font-size: 0.75rem;
  font-weight: 600;
  color: #a0a0a0;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.quick-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.quick-tag {
  padding: 6px 12px;
  background: #1a1a1a;
  border: 1px solid #2a2a2a;
  border-radius: 16px;
  color: #a0a0a0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-tag:hover {
  background: #2a2a2a;
  border-color: #3a3a3a;
  color: #fff;
}

@media (max-width: 1200px) {
  .workspace-layout {
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
