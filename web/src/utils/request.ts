import type { AxiosError, AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'
import { ElMessage } from 'element-plus'

type RetryableRequestConfig = AxiosRequestConfig & {
  retry?: number
  retryDelay?: number
  __retryCount?: number
  __requestId?: string
  __startTime?: number
}

interface CustomAxiosInstance extends Omit<AxiosInstance, 'get' | 'post' | 'put' | 'patch' | 'delete'> {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 600000, // 10分钟超时，匹配后端AI生成接口
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

const buildRequestId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`

const isPlainObject = (value: unknown): value is Record<string, unknown> => {
  if (!value || typeof value !== 'object') return false
  return Object.prototype.toString.call(value) === '[object Object]'
}

const sanitizeData = (data: unknown): unknown => {
  if (Array.isArray(data)) {
    return data.map(item => sanitizeData(item))
  }
  if (!isPlainObject(data)) return data
  const result: Record<string, unknown> = {}
  Object.keys(data).forEach((key) => {
    if (['api_key', 'authorization', 'password', 'token'].includes(key.toLowerCase())) {
      result[key] = '***'
    } else {
      result[key] = sanitizeData((data as Record<string, unknown>)[key])
    }
  })
  return result
}

const sanitizeHeaders = (headers: unknown): unknown => {
  if (!isPlainObject(headers)) return headers
  const result: Record<string, unknown> = {}
  Object.keys(headers).forEach((key) => {
    if (['authorization', 'cookie', 'set-cookie'].includes(key.toLowerCase())) {
      result[key] = '***'
    } else {
      result[key] = (headers as Record<string, unknown>)[key]
    }
  })
  return result
}

const normalizeErrorMessage = (message?: string) => {
  const text = message || ''
  const normalized = String(text).toLowerCase()
  if (
    normalized.includes('no such host') ||
    normalized.includes('dial tcp') ||
    normalized.includes('ark.cn-beijing.volces.com')
  ) {
    return 'AI服务地址无法解析，请在 设置 > AI服务配置 中切换可用服务或检查网络/DNS'
  }
  if (
    normalized.includes('authenticationerror') ||
    normalized.includes('unauthorized') ||
    normalized.includes('api key') ||
    normalized.includes('ak/sk')
  ) {
    return 'AI服务鉴权失败，请在 设置 > AI服务配置 中检查 API Key/AK/SK 是否有效'
  }
  if (
    normalized.includes('no active config found') ||
    normalized.includes('no image ai config found')
  ) {
    return '未找到可用的AI服务配置，请先在 设置 > AI服务配置 中启用配置'
  }
  return message || '请求失败'
}

const logRequest = (config: RetryableRequestConfig) => {
  const payload = {
    request_id: config.__requestId,
    url: config.baseURL ? `${config.baseURL}${config.url || ''}` : config.url,
    method: config.method,
    headers: sanitizeHeaders(config.headers),
    params: sanitizeData(config.params),
    data: sanitizeData(config.data)
  }
  console.info('API Request', payload)
}

const logResponse = (config: RetryableRequestConfig, response: any) => {
  const duration = config.__startTime ? Date.now() - config.__startTime : undefined
  const payload = {
    request_id: config.__requestId,
    url: config.baseURL ? `${config.baseURL}${config.url || ''}` : config.url,
    method: config.method,
    status: response?.status,
    duration,
    headers: sanitizeHeaders(response?.headers),
    data: sanitizeData(response?.data)
  }
  console.info('API Response', payload)
}

const logError = (config: RetryableRequestConfig | undefined, error: AxiosError<any>) => {
  const duration = config?.__startTime ? Date.now() - config.__startTime : undefined
  const payload = {
    request_id: config?.__requestId,
    url: config?.baseURL ? `${config?.baseURL}${config?.url || ''}` : config?.url,
    method: config?.method,
    status: error.response?.status,
    duration,
    response_headers: sanitizeHeaders(error.response?.headers),
    response_data: sanitizeData(error.response?.data),
    message: error.message
  }
  console.error('API Error', payload)
}

// 开源版本 - 无需认证token
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const retryConfig = config as RetryableRequestConfig
    retryConfig.__startTime = Date.now()
    if (!retryConfig.__requestId) {
      retryConfig.__requestId = buildRequestId()
    }
    if (!retryConfig.headers) {
      retryConfig.headers = {}
    }
    const headersAny = retryConfig.headers as any
    if (typeof headersAny?.set === 'function') {
      headersAny.set('X-Request-Id', retryConfig.__requestId)
    } else {
      headersAny['X-Request-Id'] = retryConfig.__requestId
    }
    const method = (retryConfig.method || 'get').toLowerCase()
    if (method === 'get' && (retryConfig.url || '').includes('/ai-configs')) {
      if (retryConfig.retry == null) {
        retryConfig.retry = 2
      }
      if (retryConfig.retryDelay == null) {
        retryConfig.retryDelay = 300
      }
    }
    // 优化提示词接口允许短暂重试
    if (method === 'post' && (retryConfig.url || '').includes('/ai/optimize-prompt')) {
      if (retryConfig.retry == null) {
        retryConfig.retry = 2
      }
      if (retryConfig.retryDelay == null) {
        retryConfig.retryDelay = 300
      }
    }
    logRequest(retryConfig)
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const retryConfig = response.config as RetryableRequestConfig
    logResponse(retryConfig, response)
    const res = response.data
    if (res.success) {
      return res.data
    } else {
      // 不在这里显示错误提示，让业务代码自行处理
      return Promise.reject(new Error(normalizeErrorMessage(res.error?.message)))
    }
  },
  (error: AxiosError<any>) => {
    const retryConfig = error.config as RetryableRequestConfig | undefined
    logError(retryConfig, error)
    if (retryConfig) {
      const method = (retryConfig.method || 'get').toLowerCase()
      const status = error.response?.status
      const url = retryConfig.url || ''
      const isIdempotent = ['get', 'head', 'options'].includes(method)
      // 非幂等接口仅对白名单开启重试
      const allowRetryPost = method === 'post' && url.includes('/ai/optimize-prompt')
      const shouldRetry = (isIdempotent || allowRetryPost) && (status == null || status >= 500)
      const maxRetries = retryConfig.retry ?? 0
      const retryCount = retryConfig.__retryCount ?? 0
      if (shouldRetry && retryCount < maxRetries) {
        retryConfig.__retryCount = retryCount + 1
        const delay = (retryConfig.retryDelay ?? 300) * (retryConfig.__retryCount)
        return new Promise((resolve) => {
          setTimeout(() => resolve(request.request(retryConfig)), delay)
        })
      }
    }
    const responseMessage = error.response?.data?.error?.message || error.response?.data?.message || error.message
    return Promise.reject(new Error(normalizeErrorMessage(responseMessage)))
  }
)

export default request
