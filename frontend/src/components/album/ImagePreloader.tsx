"use client"

import { useEffect, useCallback, useRef, useState } from 'react'
import { PhotoPreview } from './AlbumFlowClient'

interface ImagePreloaderProps {
  photos: PhotoPreview[]
  priority?: number
  onProgress?: (loaded: number, total: number) => void
}

// 全局图片缓存
class ImageCache {
  private cache = new Map<string, HTMLImageElement>()
  private loading = new Set<string>()
  private maxSize = 100 // 最大缓存数量
  
  async preload(src: string): Promise<HTMLImageElement> {
    // 如果已经缓存，直接返回
    if (this.cache.has(src)) {
      return this.cache.get(src)!
    }
    
    // 如果正在加载，等待加载完成
    if (this.loading.has(src)) {
      return new Promise((resolve, reject) => {
        const checkCache = () => {
          if (this.cache.has(src)) {
            resolve(this.cache.get(src)!)
          } else if (!this.loading.has(src)) {
            reject(new Error('加载失败'))
          } else {
            setTimeout(checkCache, 100)
          }
        }
        checkCache()
      })
    }
    
    // 开始加载
    this.loading.add(src)
    
    return new Promise((resolve, reject) => {
      const img = new Image()
      
      img.onload = () => {
        this.loading.delete(src)
        
        // 缓存管理 - LRU策略
        if (this.cache.size >= this.maxSize) {
          const firstKey = this.cache.keys().next().value
          if (firstKey) {
            this.cache.delete(firstKey)
          }
        }
        
        this.cache.set(src, img)
        resolve(img)
      }
      
      img.onerror = () => {
        this.loading.delete(src)
        reject(new Error(`加载图片失败: ${src}`))
      }
      
      // 设置缓存策略
      img.crossOrigin = 'anonymous'
      img.src = src
    })
  }
  
  has(src: string): boolean {
    return this.cache.has(src)
  }
  
  get(src: string): HTMLImageElement | undefined {
    return this.cache.get(src)
  }
  
  clear(): void {
    this.cache.clear()
    this.loading.clear()
  }
  
  getStats() {
    return {
      cached: this.cache.size,
      loading: this.loading.size,
      maxSize: this.maxSize
    }
  }
}

// 全局缓存实例
const imageCache = new ImageCache()

// 预加载Hook
export const useImagePreloader = () => {
  const preloadImages = useCallback(async (
    urls: string[], 
    onProgress?: (loaded: number, total: number) => void
  ) => {
    let loaded = 0
    const total = urls.length
    
    const results = await Promise.allSettled(
      urls.map(async (url) => {
        try {
          await imageCache.preload(url)
          loaded++
          onProgress?.(loaded, total)
          return url
        } catch (error) {
          loaded++
          onProgress?.(loaded, total)
          throw error
        }
      })
    )
    
    return results
  }, [])
  
  return {
    preloadImages,
    cache: imageCache,
    isImageCached: (src: string) => imageCache.has(src)
  }
}

// 智能预加载组件
const ImagePreloader: React.FC<ImagePreloaderProps> = ({ 
  photos, 
  priority = 5,
  onProgress 
}) => {
  const { preloadImages } = useImagePreloader()
  const preloadedRef = useRef(new Set<string>())
  
  useEffect(() => {
    if (!photos.length) return
    
    // 按优先级预加载图片
    const preloadBatch = async () => {
      // 首先预加载前几张图片（高优先级）
      const priorityUrls = photos
        .slice(0, priority)
        .map(photo => photo.url)
        .filter(url => !preloadedRef.current.has(url))
      
      if (priorityUrls.length > 0) {
        await preloadImages(priorityUrls, onProgress)
        priorityUrls.forEach(url => preloadedRef.current.add(url))
      }
      
      // 然后在空闲时间预加载剩余图片
      if ('requestIdleCallback' in window) {
        const preloadRest = () => {
          const remainingUrls = photos
            .slice(priority)
            .map(photo => photo.url)
            .filter(url => !preloadedRef.current.has(url))
            .slice(0, 10) // 每次最多预加载10张
          
          if (remainingUrls.length > 0) {
            preloadImages(remainingUrls).then(() => {
              remainingUrls.forEach(url => preloadedRef.current.add(url))
              
              // 如果还有更多图片，继续预加载
              if (preloadedRef.current.size < photos.length) {
                requestIdleCallback(preloadRest)
              }
            })
          }
        }
        
        requestIdleCallback(preloadRest)
      }
    }
    
    preloadBatch()
  }, [photos, priority, preloadImages, onProgress])
  
  return null // 这是一个无UI组件
}

// 预加载状态Hook
export const usePreloadProgress = () => {
  const [progress, setProgress] = useState({ loaded: 0, total: 0 })
  
  const updateProgress = useCallback((loaded: number, total: number) => {
    setProgress({ loaded, total })
  }, [])
  
  const percentage = progress.total > 0 ? (progress.loaded / progress.total) * 100 : 0
  
  return {
    progress,
    percentage,
    updateProgress,
    isComplete: progress.loaded === progress.total && progress.total > 0
  }
}

export default ImagePreloader 