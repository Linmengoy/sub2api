<template>
  <div class="image-generator">
    <div class="generator-container">
      <div class="generator-header">
        <h1 class="page-title">🎨 {{ $t('image.title') }}</h1>
        <p class="page-description">{{ $t('image.description') }}</p>
      </div>

      <div class="generator-layout">
        <div class="controls-section">
          <div class="prompt-card">
            <h3>{{ $t('image.prompt') }}</h3>
            <TextArea
              v-model="prompt"
              :placeholder="$t('image.enterPrompt')"
              :rows="5"
              :maxlength="4000"
              show-count
            />
          </div>

          <div class="model-selector">
            <h3>{{ $t('image.model') }}</h3>
            <div class="model-grid">
              <div
                v-for="model in models"
                :key="model.id"
                :class="['model-card', { active: selectedModel === model.id }]"
                @click="selectedModel = model.id"
              >
                <div class="model-icon">{{ model.icon }}</div>
                <div class="model-name">{{ model.name }}</div>
                <div class="model-price">{{ model.price }}</div>
              </div>
            </div>
          </div>

          <div class="parameters-grid">
            <div class="parameter-card">
              <label>{{ $t('image.size') }}</label>
              <Select v-model="size" :options="sizeOptions" />
            </div>

            <div class="parameter-card">
              <label>{{ $t('image.quality') }}</label>
              <Select v-model="quality" :options="qualityOptions" />
            </div>

            <div class="parameter-card">
              <label>{{ $t('image.style') }}</label>
              <Select v-model="style" :options="styleOptions" />
            </div>

            <div class="parameter-card">
              <label>{{ $t('image.number') }}</label>
              <Select v-model="n" :options="numberOptions" />
            </div>
          </div>

          <Button
            type="primary"
            size="lg"
            :loading="generating"
            :disabled="!prompt.trim() || generating"
            class="generate-button"
            @click="handleGenerate"
          >
            {{ generating ? $t('image.generating') : $t('image.generate') }}
          </Button>

          <div v-if="generating" class="generation-progress">
            <LoadingSpinner size="md" />
            <span>{{ $t('image.pleaseWait') }}</span>
          </div>
        </div>

        <div class="preview-section">
          <div v-if="!imageUrl && !generating" class="placeholder">
            <div class="placeholder-icon">🖼️</div>
            <p>{{ $t('image.placeholder') }}</p>
          </div>

          <div v-if="generating && !imageUrl" class="generating-placeholder">
            <LoadingSpinner size="lg" />
            <p>{{ $t('image.generating') }}</p>
          </div>

          <div v-if="imageUrl" class="result-container">
            <img :src="imageUrl" :alt="$t('image.generatedImage')" class="generated-image" />
            <div class="image-actions">
              <Button @click="handleDownload" icon="download">
                📥 {{ $t('image.download') }}
              </Button>
              <Button @click="handleShare" icon="share">
                🔗 {{ $t('image.share') }}
              </Button>
              <Button type="primary" @click="handleRegenerate" icon="refresh">
                🔄 {{ $t('image.regenerate') }}
              </Button>
            </div>
            <div v-if="revisedPrompt" class="revised-prompt">
              <h4>{{ $t('image.revisedPrompt') }}</h4>
              <p>{{ revisedPrompt }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="templates-section">
        <h3>💡 {{ $t('image.templates') }}</h3>
        <div class="template-list">
          <Tag
            v-for="template in templates"
            :key="template"
            class="template-tag"
            @click="prompt = template"
          >
            {{ template.substring(0, 30) }}{{ template.length > 30 ? '...' : '' }}
          </Tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Button from '@/components/common/Button.vue'
import TextArea from '@/components/common/TextArea.vue'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Tag from '@/components/common/Tag.vue'
import { imagesAPI } from '@/api/images'

const { t } = useI18n()

const prompt = ref('')
const selectedModel = ref('dall-e-3')
const size = ref('1024x1024')
const quality = ref('standard')
const style = ref('vivid')
const n = ref(1)
const generating = ref(false)
const imageUrl = ref('')
const revisedPrompt = ref('')

const models = [
  { id: 'dall-e-3', name: 'DALL-E 3', icon: '🎨', price: t('image.prices.dalle3') },
  { id: 'dall-e-2', name: 'DALL-E 2', icon: '🖌️', price: t('image.prices.dalle2') },
  { id: 'gpt-image-2', name: 'GPT Image', icon: '🤖', price: t('image.prices.gptImage') },
]

const sizeOptions = [
  { value: '1024x1024', label: '1024×1024 (1:1)' },
  { value: '1792x1024', label: '1792×1024 (16:9)' },
  { value: '1024x1792', label: '1024×1792 (9:16)' },
  { value: '512x512', label: '512×512 (1:1)' },
  { value: '256x256', label: '256×256 (1:1)' },
]

const qualityOptions = [
  { value: 'standard', label: t('image.qualityStandard') },
  { value: 'hd', label: t('image.qualityHD') },
]

const styleOptions = [
  { value: 'vivid', label: t('image.styleVivid') },
  { value: 'natural', label: t('image.styleNatural') },
]

const numberOptions = [
  { value: 1, label: '1' },
  { value: 2, label: '2' },
  { value: 3, label: '3' },
  { value: 4, label: '4' },
]

const templates = [
  'A beautiful sunset over the ocean with vibrant orange and purple colors',
  'Futuristic city with flying cars and neon lights',
  'Cute cat wearing a space suit floating in space',
  'Abstract art with geometric shapes in bright colors',
  'A cozy coffee shop interior with warm lighting',
  'Mountain landscape covered in snow at sunrise',
  'Underwater coral reef with colorful fish',
  'Steampunk airship flying through clouds',
]

async function handleGenerate() {
  if (!prompt.value.trim() || generating.value) return

  generating.value = true
  imageUrl.value = ''
  revisedPrompt.value = ''

  try {
    const response = await imagesAPI.generate({
      model: selectedModel.value,
      prompt: prompt.value,
      n: n.value,
      size: size.value as any,
      quality: quality.value,
      style: style.value,
      response_format: 'url',
    })

    if (response.data && response.data.length > 0) {
      const imageData = response.data[0]
      imageUrl.value = imageData.url || ''
      revisedPrompt.value = imageData.revised_prompt || ''
    }
  } catch (error: any) {
    console.error('Image generation failed:', error)
    alert(error.message || t('image.generationFailed'))
  } finally {
    generating.value = false
  }
}

function handleDownload() {
  if (!imageUrl.value) return

  const link = document.createElement('a')
  link.href = imageUrl.value
  link.download = `generated-image-${Date.now()}.png`
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function handleShare() {
  if (!imageUrl.value) return

  if (navigator.share) {
    navigator.share({
      title: t('image.generatedImage'),
      text: prompt.value,
      url: imageUrl.value,
    })
  } else {
    navigator.clipboard.writeText(imageUrl.value)
    alert(t('image.linkCopied'))
  }
}

function handleRegenerate() {
  handleGenerate()
}
</script>

<style scoped>
.image-generator {
  min-height: 100vh;
  background: linear-gradient(135deg, #f0fdfa 0%, #e2e8f0 100%);
  padding: 2rem;
}

.generator-container {
  max-width: 1400px;
  margin: 0 auto;
}

.generator-header {
  text-align: center;
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: #0f766e;
  margin-bottom: 0.5rem;
}

.page-description {
  font-size: 1.125rem;
  color: #64748b;
}

.generator-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  margin-bottom: 2rem;
}

.controls-section {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.prompt-card,
.model-selector,
.parameters-grid {
  background: white;
  border-radius: 1rem;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 1rem;
}

.model-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.model-card {
  padding: 1rem;
  border: 2px solid #e2e8f0;
  border-radius: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.model-card:hover {
  border-color: #14b8a6;
  transform: translateY(-2px);
}

.model-card.active {
  border-color: #14b8a6;
  background: #f0fdfa;
}

.model-icon {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.model-name {
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 0.25rem;
}

.model-price {
  font-size: 0.875rem;
  color: #64748b;
}

.parameters-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.parameter-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.parameter-card label {
  font-weight: 600;
  color: #475569;
  font-size: 0.875rem;
}

.generate-button {
  width: 100%;
  height: 3rem;
  font-size: 1.125rem;
}

.generation-progress {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  color: #64748b;
  font-size: 0.875rem;
}

.preview-section {
  background: white;
  border-radius: 1rem;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  min-height: 500px;
}

.placeholder,
.generating-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}

.placeholder-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.placeholder p,
.generating-placeholder p {
  font-size: 1.125rem;
}

.result-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.generated-image {
  width: 100%;
  border-radius: 0.75rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
}

.image-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.revised-prompt {
  background: #f8fafc;
  border-radius: 0.5rem;
  padding: 1rem;
  border-left: 4px solid #14b8a6;
}

.revised-prompt h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 0.5rem;
}

.revised-prompt p {
  color: #475569;
  font-size: 0.875rem;
}

.templates-section {
  background: white;
  border-radius: 1rem;
  padding: 1.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.template-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.template-tag {
  cursor: pointer;
  padding: 0.5rem 1rem;
  background: #f1f5f9;
  border-radius: 9999px;
  font-size: 0.875rem;
  color: #475569;
  transition: all 0.2s;
}

.template-tag:hover {
  background: #e2e8f0;
  color: #0f766e;
}

@media (max-width: 1024px) {
  .generator-layout {
    grid-template-columns: 1fr;
  }

  .model-grid {
    grid-template-columns: 1fr;
  }

  .parameters-grid {
    grid-template-columns: 1fr;
  }
}
</style>
