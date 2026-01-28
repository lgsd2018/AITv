import request from '../utils/request'
import type { AIServiceConfig, CreateAIConfigRequest, UpdateAIConfigRequest, TestConnectionRequest } from '../types/ai'

export const aiAPI = {
  // AI Configs
  list(serviceType?: string) {
    return request.get<AIServiceConfig[]>('/ai-configs', {
      params: { service_type: serviceType }
    })
  },

  get(id: number) {
    return request.get<AIServiceConfig>(`/ai-configs/${id}`)
  },

  create(data: CreateAIConfigRequest) {
    return request.post<AIServiceConfig>('/ai-configs', data)
  },

  update(id: number, data: UpdateAIConfigRequest) {
    return request.put<AIServiceConfig>(`/ai-configs/${id}`, data)
  },

  delete(id: number) {
    return request.delete(`/ai-configs/${id}`)
  },

  testConnection(data: TestConnectionRequest) {
    return request.post<{ message: string }>('/ai-configs/test', data)
  },

  // Other AI methods
  reversePrompt(imageUrl: string) {
    return request.post<{ prompt: string }>('/ai/reverse-prompt', {
      image_url: imageUrl
    })
  }
}
