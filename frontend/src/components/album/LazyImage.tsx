"use client"

import React, { useState, useRef, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

interface LazyImageProps {
  src: string
  alt: string
  className?: string
  placeholder?: string
  aspectRatio?: number
  onLoad?: () => void
  onClick?: () => void
  priority?: boolean
  sizes?: string
}

// 图片缓存
const imageCache = new Set<string>()

// 预加载图片
const preloadImage = (src: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (imageCache.has(src)) {
      resolve()
      return
    }
    
    const img = new Image()
    img.onload = () => {
      imageCache.add(src)
      resolve()
    }
    img.onerror = reject
    img.src = src
  })
}

const LazyImage: React.FC<LazyImageProps> = ({
  src,
  alt,
  className = '',
  placeholder,
  aspectRatio = 4/3,
  onLoad,
  onClick,
  priority = false,
  sizes
}) => {
  const [isLoaded, setIsLoaded] = useState(false)
  const [isInView, setIsInView] = useState(priority)
  const [error, setError] = useState(false)
  const imgRef = useRef<HTMLImageElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  // Intersection Observer 用于懒加载
  useEffect(() => {
    if (priority) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true)
          observer.disconnect()
        }
      },
      {
        rootMargin: '200px 0px', // 提前200px开始加载
        threshold: 0.1
      }
    )

    if (containerRef.current) {
      observer.observe(containerRef.current)
    }

    return () => observer.disconnect()
  }, [priority])

  // 预加载图片
  useEffect(() => {
    if (!isInView) return

    preloadImage(src)
      .then(() => {
        setIsLoaded(true)
        onLoad?.()
      })
      .catch(() => {
        setError(true)
      })
  }, [isInView, src, onLoad])

  const handleImageLoad = useCallback(() => {
    setIsLoaded(true)
    onLoad?.()
  }, [onLoad])

  return (
    <div
      ref={containerRef}
      className={`relative overflow-hidden cursor-pointer ${className}`}
      style={{ aspectRatio }}
      onClick={onClick}
    >
      {/* 占位背景 */}
      {placeholder && (
        <div
          className="absolute inset-0 bg-center bg-cover"
          style={{
            backgroundColor: placeholder,
            filter: 'blur(20px)',
            transform: 'scale(1.1)' // 避免模糊边缘
          }}
        />
      )}

      {/* 加载状态 */}
      <AnimatePresence>
        {!isLoaded && isInView && !error && (
          <motion.div
            initial={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.3 }}
            className="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-800"
          >
            <div className="relative">
              <div className="w-8 h-8 border-3 border-gray-300 border-t-blue-500 rounded-full animate-spin" />
              <div className="absolute inset-0 w-8 h-8 border-3 border-transparent border-t-blue-400 rounded-full animate-spin opacity-50" 
                   style={{ animationDelay: '0.1s', animationDirection: 'reverse' }} />
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* 错误状态 */}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-800 text-gray-400">
          <div className="text-center">
            <div className="text-2xl mb-2">📷</div>
            <div className="text-sm">加载失败</div>
          </div>
        </div>
      )}

      {/* 实际图片 */}
      {isInView && !error && (
        <motion.img
          ref={imgRef}
          src={src}
          alt={alt}
          className="absolute inset-0 w-full h-full object-cover transition-transform duration-700 hover:scale-105"
          style={{ 
            opacity: isLoaded ? 1 : 0,
            transition: 'opacity 0.5s ease-in-out'
          }}
          onLoad={handleImageLoad}
          sizes={sizes}
          loading={priority ? 'eager' : 'lazy'}
        />
      )}

      {/* 悬停遮罩 */}
      <div className="absolute inset-0 bg-black/0 hover:bg-black/10 transition-colors duration-300" />
    </div>
  )
}

export default LazyImage 