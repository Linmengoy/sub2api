/**
 * Images API
 * 调用后端 /v1/images/generations 端点
 */

import { apiClient } from './client'

export interface ImageGenerationRequest {
  model?: string
  prompt: string
  n?: number
  size?: '1024x1024' | '1792x1024' | '1024x1792' | '256x256' | '512x512'
  quality?: 'standard' | 'hd'
  style?: 'vivid' | 'natural'
  response_format?: 'url' | 'b64_json'
}

export interface ImageGenerationResponse {
  created: number
  data: Array<{
    url?: string
    b64_json?: string
    revised_prompt?: string
  }>
}

export const imagesAPI = {
  async generate(params: ImageGenerationRequest): Promise<ImageGenerationResponse> {
    const response = await apiClient.post<ImageGenerationResponse>(
      '/images/generations',
      params
    )
    return response.data
  },

  async generateWithStream(
    params: ImageGenerationRequest,
    onProgress?: (data: any) => void
  ): Promise<ImageGenerationResponse> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      
      xhr.open('POST', '/api/v1/images/generations', true)
      
      const token = localStorage.getItem('auth_token')
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`)
      }
      xhr.setRequestHeader('Content-Type', 'application/json')
      xhr.setRequestHeader('Accept', 'text/event-stream')
      
      xhr.onprogress = (event) => {
        if (event.lengthComputable && onProgress) {
          const progress = (event.loaded / event.total) * 100
          onProgress({ type: 'progress', percentage: progress })
        }
      }
      
      xhr.onload = function() {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const data = JSON.parse(xhr.responseText)
            resolve(data)
          } catch (error) {
            reject(new Error('Failed to parse response'))
          }
        } else {
          try {
            const error = JSON.parse(xhr.responseText)
            reject(new Error(error.error?.message || 'Image generation failed'))
          } catch {
            reject(new Error(`Request failed with status ${xhr.status}`))
          }
        }
      }
      
      xhr.onerror = function() {
        reject(new Error('Network error'))
      }
      
      xhr.send(JSON.stringify({
        ...params,
        stream: false
      }))
    })
  },

  getImageUrl(b64Json: string): string {
    return `data:image/png;base64,${b64Json}`
  }
}
