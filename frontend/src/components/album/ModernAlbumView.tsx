"use client"

import React, { useState, useEffect, useCallback, useRef } from 'react'
import { motion, AnimatePresence, PanInfo } from 'framer-motion'
import { createPortal } from 'react-dom'
import { X, ChevronLeft, ChevronRight, Info, Download, Share2, MapPin, Calendar, Camera, Clock } from 'lucide-react'
import LazyImage from './LazyImage'
import { PhotoPreview } from './AlbumFlowClient'

interface ModernAlbumViewProps {
  photo: PhotoPreview
  isOpen: boolean
  onClose: () => void
  onPrevious: () => void
  onNext: () => void
  hasPrevious: boolean
  hasNext: boolean
  totalCount?: number
  currentIndex?: number
}

const ModernAlbumView: React.FC<ModernAlbumViewProps> = ({
  photo,
  isOpen,
  onClose,
  onPrevious,
  onNext,
  hasPrevious,
  hasNext,
  totalCount,
  currentIndex
}) => {
  const [showInfo, setShowInfo] = useState(false)
  const [, setIsImageLoaded] = useState(false)
  const [dragDirection, setDragDirection] = useState<'left' | 'right' | null>(null)
  const [showControls, setShowControls] = useState(true)
  const imageRef = useRef<HTMLDivElement>(null)
  const timerRef = useRef<NodeJS.Timeout | null>(null)

  // 自动隐藏控件
  useEffect(() => {
    const resetTimer = () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      setShowControls(true)
      timerRef.current = setTimeout(() => setShowControls(false), 3000)
    }

    if (isOpen) {
      resetTimer()
      const handleMouseMove = () => resetTimer()
      window.addEventListener('mousemove', handleMouseMove)
      return () => {
        window.removeEventListener('mousemove', handleMouseMove)
        if (timerRef.current) clearTimeout(timerRef.current)
      }
    }
  }, [isOpen])

  // 键盘导航
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return
      
      switch (e.key) {
        case 'Escape':
          onClose()
          break
        case 'ArrowLeft':
          if (hasPrevious) onPrevious()
          break
        case 'ArrowRight':
          if (hasNext) onNext()
          break
        case 'i':
        case 'I':
          setShowInfo(!showInfo)
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, hasPrevious, hasNext, onClose, onPrevious, onNext, showInfo])

  // 处理滑动手势
  const handleDragEnd = useCallback((_:never, info: PanInfo) => {
    const swipeThreshold = 100
    const swipeVelocityThreshold = 500

    if (Math.abs(info.offset.x) > swipeThreshold || Math.abs(info.velocity.x) > swipeVelocityThreshold) {
      if (info.offset.x > 0 && hasPrevious) {
        onPrevious()
      } else if (info.offset.x < 0 && hasNext) {
        onNext()
      }
    }
    setDragDirection(null)
  }, [hasPrevious, hasNext, onPrevious, onNext])

  const handleDrag = useCallback((_:never, info: PanInfo) => {
    if (Math.abs(info.offset.x) > 20) {
      setDragDirection(info.offset.x > 0 ? 'right' : 'left')
    }
  }, [])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return {
      date: date.toLocaleDateString('zh-CN', { 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      }),
      time: date.toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit' 
      })
    }
  }

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: '精彩照片分享',
          text: photo.description || `拍摄于 ${formatDate(photo.date).date}`,
          url: photo.url
        })
      } catch (error) {
        console.log('分享失败:', error)
      }
    } else {
      // 复制链接到剪贴板
      await navigator.clipboard.writeText(photo.url)
    }
  }

  const handleDownload = () => {
    const link = document.createElement('a')
    link.href = photo.url
    link.download = `photo_${photo.id}.jpg`
    link.click()
  }

  if (!isOpen) return null

  const { date, time } = formatDate(photo.date)

  return createPortal(
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.3 }}
        className="fixed inset-0 z-50 bg-black/95 backdrop-blur-sm"
        onClick={onClose}
      >
        {/* 顶部工具栏 */}
        <motion.div
          initial={{ y: -100, opacity: 0 }}
          animate={{ 
            y: showControls ? 0 : -100, 
            opacity: showControls ? 1 : 0 
          }}
          transition={{ duration: 0.3 }}
          className="absolute top-0 left-0 right-0 z-60 bg-gradient-to-b from-black/50 to-transparent p-4 sm:p-6"
        >
          <div className="flex items-center justify-between text-white">
            <div className="flex items-center space-x-4">
              <button
                onClick={onClose}
                className="p-2 hover:bg-white/10 rounded-full transition-colors"
              >
                <X size={24} />
              </button>
              {totalCount && currentIndex !== undefined && (
                <span className="text-sm opacity-75">
                  {currentIndex + 1} / {totalCount}
                </span>
              )}
            </div>
            
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setShowInfo(!showInfo)}
                className="p-2 hover:bg-white/10 rounded-full transition-colors"
              >
                <Info size={20} />
              </button>
              <button
                onClick={handleShare}
                className="p-2 hover:bg-white/10 rounded-full transition-colors"
              >
                <Share2 size={20} />
              </button>
              <button
                onClick={handleDownload}
                className="p-2 hover:bg-white/10 rounded-full transition-colors"
              >
                <Download size={20} />
              </button>
            </div>
          </div>
        </motion.div>

        {/* 主图区域 */}
        <div 
          className="flex items-center justify-center h-full p-4 sm:p-8"
          onClick={(e) => e.stopPropagation()}
        >
          <motion.div
            ref={imageRef}
            drag="x"
            dragConstraints={{ left: 0, right: 0 }}
            dragElastic={0.2}
            onDrag={handleDrag}
            onDragEnd={handleDragEnd}
            animate={{
              x: dragDirection === 'left' ? -20 : dragDirection === 'right' ? 20 : 0,
              scale: dragDirection ? 0.95 : 1
            }}
            transition={{ type: "spring", stiffness: 300, damping: 30 }}
            className="relative max-w-[90vw] max-h-[80vh] cursor-grab active:cursor-grabbing"
          >
            <div
              className="relative bg-center bg-cover rounded-xl overflow-hidden shadow-2xl"
              style={{ backgroundColor: photo.shade }}
            >
              <LazyImage
                src={photo.url}
                alt={`Photo from ${photo.date}`}
                placeholder={photo.shade}
                priority={true}
                onLoad={() => setIsImageLoaded(true)}
                className="w-full h-full object-contain"
              />
              
              {/* 拖拽提示 */}
              {dragDirection && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="absolute inset-0 flex items-center justify-center bg-black/20 rounded-xl"
                >
                  <div className="flex items-center space-x-2 text-white text-lg">
                    {dragDirection === 'left' && hasNext && (
                      <>
                        <span>下一张</span>
                        <ChevronRight size={24} />
                      </>
                    )}
                    {dragDirection === 'right' && hasPrevious && (
                      <>
                        <ChevronLeft size={24} />
                        <span>上一张</span>
                      </>
                    )}
                  </div>
                </motion.div>
              )}
            </div>
          </motion.div>
        </div>

        {/* 导航按钮 */}
        <AnimatePresence>
          {showControls && (
            <>
              {hasPrevious && (
                <motion.button
                  initial={{ x: -100, opacity: 0 }}
                  animate={{ x: 0, opacity: 1 }}
                  exit={{ x: -100, opacity: 0 }}
                  onClick={onPrevious}
                  className="absolute left-4 top-1/2 -translate-y-1/2 p-3 bg-black/30 hover:bg-black/50 text-white rounded-full backdrop-blur-sm transition-all"
                >
                  <ChevronLeft size={24} />
                </motion.button>
              )}
              
              {hasNext && (
                <motion.button
                  initial={{ x: 100, opacity: 0 }}
                  animate={{ x: 0, opacity: 1 }}
                  exit={{ x: 100, opacity: 0 }}
                  onClick={onNext}
                  className="absolute right-4 top-1/2 -translate-y-1/2 p-3 bg-black/30 hover:bg-black/50 text-white rounded-full backdrop-blur-sm transition-all"
                >
                  <ChevronRight size={24} />
                </motion.button>
              )}
            </>
          )}
        </AnimatePresence>

        {/* 信息面板 */}
        <AnimatePresence>
          {showInfo && (
            <motion.div
              initial={{ x: '100%' }}
              animate={{ x: 0 }}
              exit={{ x: '100%' }}
              transition={{ type: "spring", stiffness: 300, damping: 30 }}
              className="absolute right-0 top-0 bottom-0 w-80 bg-black/80 backdrop-blur-xl text-white overflow-y-auto"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="p-6 space-y-6">
                <div className="flex items-center justify-between">
                  <h3 className="text-xl font-semibold">照片信息</h3>
                  <button
                    onClick={() => setShowInfo(false)}
                    className="p-1 hover:bg-white/10 rounded"
                  >
                    <X size={20} />
                  </button>
                </div>

                <div className="space-y-4">
                  <div className="flex items-start space-x-3">
                    <Calendar size={20} className="text-blue-400 mt-0.5" />
                    <div>
                      <div className="font-medium">{date}</div>
                      <div className="text-sm text-gray-300 flex items-center mt-1">
                        <Clock size={16} className="mr-1" />
                        {time}
                      </div>
                    </div>
                  </div>

                  {photo.location && (
                    <div className="flex items-start space-x-3">
                      <MapPin size={20} className="text-green-400 mt-0.5" />
                      <div>
                        <div className="font-medium">拍摄地点</div>
                        <div className="text-sm text-gray-300">{photo.location}</div>
                      </div>
                    </div>
                  )}

                  {photo.device && (
                    <div className="flex items-start space-x-3">
                      <Camera size={20} className="text-purple-400 mt-0.5" />
                      <div>
                        <div className="font-medium">拍摄设备</div>
                        <div className="text-sm text-gray-300">{photo.device}</div>
                      </div>
                    </div>
                  )}

                  {photo.description && (
                    <div className="pt-4 border-t border-white/10">
                      <div className="font-medium mb-2">描述</div>
                      <div className="text-sm text-gray-300 leading-relaxed">
                        {photo.description}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </AnimatePresence>,
    document.body
  )
}

export default ModernAlbumView 