"use client"

import React, { useEffect, useRef, useState, useCallback, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { ArrowUp, Loader2 } from 'lucide-react'
import ModernPhotoGroup from './ModernPhotoGroup'
import ImagePreloader, { usePreloadProgress } from './ImagePreloader'
import LoadingProgressBar from './LoadingProgressBar'
import { fetchPhotosByPage } from '@/api/photos'
import { PhotoPreview } from './AlbumFlowClient'

interface ModernAlbumFlowClientProps {
  initialImages: PhotoPreview[]
}

const ModernAlbumFlowClient: React.FC<ModernAlbumFlowClientProps> = ({ initialImages }) => {
  const [photoGroups, setPhotoGroups] = useState<Record<string, PhotoPreview[]>>({})
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [showScrollToTop, setShowScrollToTop] = useState(false)
  const [allPhotos, setAllPhotos] = useState<PhotoPreview[]>([])
  
  const { progress, percentage, updateProgress } = usePreloadProgress()
  const observer = useRef<IntersectionObserver | null>(null)
  const loadTriggerRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  // 性能优化：使用 memo 计算排序后的分组
  const sortedPhotoGroups = useMemo(() => {
    return Object.entries(photoGroups).sort(([a], [b]) => 
      new Date(b).getTime() - new Date(a).getTime()
    )
  }, [photoGroups])

  // 初始化照片分组
  useEffect(() => {
    if (initialImages.length > 0) {
      groupPhotosByDate(initialImages)
      setAllPhotos(initialImages)
    }
  }, [initialImages])

  // 监听滚动，显示/隐藏回到顶部按钮
  useEffect(() => {
    const handleScroll = () => {
      setShowScrollToTop(window.scrollY > 600)
    }

    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  // 设置无限滚动观察器
  useEffect(() => {
    if (!hasMore || loading) return

    if (observer.current) observer.current.disconnect()

    observer.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          loadMorePhotos()
        }
      },
      {
        rootMargin: '400px', // 提前400px触发加载
        threshold: 0.1
      }
    )

    if (loadTriggerRef.current) {
      observer.current.observe(loadTriggerRef.current)
    }

    return () => {
      if (observer.current) observer.current.disconnect()
    }
  }, [hasMore, loading])

  const groupPhotosByDate = useCallback((photos: PhotoPreview[]) => {
    setPhotoGroups(prevGroups => {
      const newGroups = { ...prevGroups }
      
      photos.forEach(photo => {
        const date = new Date(photo.date).toLocaleDateString('zh-CN')
        if (!newGroups[date]) {
          newGroups[date] = []
        }
        
        // 避免重复添加
        if (!newGroups[date].some(p => p.id === photo.id)) {
          newGroups[date].push(photo)
        }
      })
      
      // 对每个日期分组内的照片按时间排序
      Object.keys(newGroups).forEach(date => {
        newGroups[date].sort((a, b) => 
          new Date(b.date).getTime() - new Date(a.date).getTime()
        )
      })
      
      return newGroups
    })
  }, [])

  const loadMorePhotos = useCallback(async () => {
    if (loading || !hasMore) return

    setLoading(true)
    
    try {
      const newPhotos = await fetchPhotosByPage(page + 1, 12) // 增加每页数量以提升性能
      
      if (newPhotos.length === 0) {
        setHasMore(false)
      } else {
        groupPhotosByDate(newPhotos)
        setPage(prev => prev + 1)
        setAllPhotos(prev => [...prev, ...newPhotos])
      }
    } catch (error) {
      console.error('加载照片失败:', error)
    } finally {
      setLoading(false)
    }
  }, [loading, hasMore, page, groupPhotosByDate])

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    })
  }

  const totalPhotoCount = useMemo(() => {
    return Object.values(photoGroups).reduce((sum, photos) => sum + photos.length, 0)
  }, [photoGroups])

  return (
    <div ref={containerRef} className="relative">
      {/* 精致加载进度条 */}
      <LoadingProgressBar
        progress={percentage}
        isVisible={loading && progress.total > 0}
      />

      {/* 图片预加载器 */}
      <ImagePreloader
        photos={allPhotos}
        priority={8}
        onProgress={updateProgress}
      />

      {/* 精致统计信息 */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="mb-6 text-center"
      >
        <div className="inline-flex items-center space-x-3 px-4 py-2 bg-gray-50/80 dark:bg-gray-800/50 backdrop-blur-sm rounded-xl shadow-sm text-xs">
          <span className="text-gray-600 dark:text-gray-400">
            <span className="font-medium text-gray-800 dark:text-gray-200">{totalPhotoCount}</span> 张照片
          </span>
          <div className="w-px h-3 bg-gray-300 dark:bg-gray-600" />
          <span className="text-gray-600 dark:text-gray-400">
            <span className="font-medium text-gray-800 dark:text-gray-200">{sortedPhotoGroups.length}</span> 天
          </span>
          {progress.total > 0 && (
            <>
              <div className="w-px h-3 bg-gray-300 dark:bg-gray-600" />
              <span className="text-gray-600 dark:text-gray-400">
                缓存 <span className="font-medium text-gray-800 dark:text-gray-200">{progress.loaded}</span>/{progress.total}
              </span>
            </>
          )}
        </div>
      </motion.div>

      {/* 照片分组 */}
      <div className="space-y-12">
        <AnimatePresence>
          {sortedPhotoGroups.map(([date, photos], groupIndex) => (
            <ModernPhotoGroup
              key={date}
              date={date}
              photos={photos}
              lastPhotoRef={
                groupIndex === sortedPhotoGroups.length - 1 
                  ? (node) => { loadTriggerRef.current = node }
                  : undefined
              }
            />
          ))}
        </AnimatePresence>
      </div>

      {/* 精致加载状态 */}
      <AnimatePresence>
        {loading && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="flex items-center justify-center py-12"
          >
            <div className="flex items-center space-x-2 px-4 py-2 bg-gray-50/80 dark:bg-gray-800/50 backdrop-blur-sm rounded-lg shadow-sm">
              <Loader2 className="w-3.5 h-3.5 animate-spin text-gray-500" />
              <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                加载中
              </span>
              {progress.total > 0 && (
                <span className="text-xs text-gray-400">
                  {Math.round(percentage)}%
                </span>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* 加载触发器 */}
      <div ref={loadTriggerRef} className="h-20 flex items-center justify-center">
        {!hasMore && !loading && totalPhotoCount > 0 && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center py-6"
          >
            <div className="inline-flex flex-col items-center space-y-2 px-6 py-4 bg-gray-50/60 dark:bg-gray-800/30 rounded-2xl">
              <div className="text-lg">📸</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">
                已浏览全部 {totalPhotoCount} 张照片
              </div>
              <div className="text-xs text-gray-400 dark:text-gray-500">
                本站照片未经授权禁止转载
              </div>
            </div>
          </motion.div>
        )}
      </div>

      {/* 精致回到顶部按钮 */}
      <AnimatePresence>
        {showScrollToTop && (
          <motion.button
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={scrollToTop}
            className="fixed bottom-6 right-6 z-40 p-3 bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm text-gray-600 dark:text-gray-300 rounded-xl shadow-lg border border-gray-200/50 dark:border-gray-700/50 hover:shadow-xl transition-all duration-300"
          >
            <ArrowUp className="w-4 h-4" />
          </motion.button>
        )}
      </AnimatePresence>
    </div>
  )
}

export default ModernAlbumFlowClient 