import axios from 'axios'
import type { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import router from '@/router'
import { useUserStore } from '@/store'
import { apiUrlPrefix, isObject, Message } from '@/utils'

interface CustomFetchOptions {
  prefix?: string
  loading?: string
  loadFn?: () => void
  cancelResInterceptor?: boolean
  successMessage?: boolean
}

type CustomAxiosRequestConfig = AxiosRequestConfig & {
  fetchOptions?: CustomFetchOptions
}

type CustomInternalAxiosRequestConfig = InternalAxiosRequestConfig & {
  fetchOptions?: CustomFetchOptions
}

const instance = axios.create({
  baseURL: apiUrlPrefix,
  timeout: 600000,
  headers: {
    'Content-Type': 'application/json',
  },
  fetchOptions: {
    prefix: '',
    successMessage: false,
  },
} as CustomAxiosRequestConfig)

const whitePathList = ['/login']

const controllerStore = new Map<string, AbortController>()

const generateReqKey = (config: AxiosRequestConfig): string => {
  const { method, url, params, data } = config
  return `${method}-${url}-${JSON.stringify(params)}-${JSON.stringify(data)}`
}

const removeController = (config: AxiosRequestConfig): void => {
  const key = generateReqKey(config)
  if (controllerStore.has(key)) {
    controllerStore.delete(key)
  }
}

// TEMPORARY FOR DEMO: отключаем AbortController
instance.interceptors.request.use((config: CustomInternalAxiosRequestConfig) => {
  // addController(config)  // ← Закомментировано для демо
  const { headers, url } = config
  if (!whitePathList.includes(url || '')) {
    const userStore = useUserStore()
    if (headers) {
      headers.Authorization = `Bearer ${userStore.login.token}`
    }
  }
  return config
})

instance.interceptors.request.use((config: CustomInternalAxiosRequestConfig) => {
  const { fetchOptions } = config
  if (isObject<CustomFetchOptions>(fetchOptions)) {
    config.url = `${fetchOptions.prefix || ''}` + config.url
  }
  return config
})

instance.interceptors.request.use((config: CustomInternalAxiosRequestConfig) => {
  const { fetchOptions } = config
  if (isObject<CustomFetchOptions>(fetchOptions) && fetchOptions.loading) {
    const { close } = Message.loading(fetchOptions.loading)
    fetchOptions.loadFn = close
  }
  return config
})

instance.interceptors.response.use(
  response => {
    removeController(response.config)

    const { fetchOptions } = response.config as CustomAxiosRequestConfig
    const { code, data, msg, success } = response.data || {}

    if (fetchOptions?.cancelResInterceptor) {
      return Promise.resolve(data)
    }

    if (fetchOptions?.loadFn) {
      fetchOptions.loadFn()
    }

    if (response.data.type === 'application/octet-stream') {
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      const disposition = response.headers['content-disposition']
      const filename = disposition
        ? decodeURIComponent(disposition.split('filename=')[1])
        : 'downloaded_file'
      link.setAttribute('download', filename)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      return Promise.resolve(data)
    }

    if (code === 0 && success) {
      if (fetchOptions?.successMessage) {
        Message.success(msg)
      }
      return Promise.resolve(data)
    }
    if (code === undefined && success === undefined) {
      return Promise.resolve(response.data)
    }
    if (!success && msg) {
      Message.error(msg, {
        close: true,
      })
    }
    if (code === 401) {
      const userStore = useUserStore()
      userStore.resetLoginInfo()
      router.push('/login')
    }
    return Promise.reject(response.data)
  },
  error => {
    if (!axios.isCancel(error)) {
      removeController(error.config || {})
    }
    return Promise.reject(error)
  }
)

const clearPendingRequests = (): void => {
  controllerStore.forEach((controller, key) => {
    controller.abort()
    controllerStore.delete(key)
  })
}

export { instance, clearPendingRequests }