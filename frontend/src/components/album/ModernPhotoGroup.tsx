"use client"

import React, { useState, useMemo } from 'react'
import { motion } from 'framer-motion'
import { Calendar } from 'lucide-react'
import MasonryGrid from './MasonryGrid'
import EnhancedModernAlbumView from './EnhancedModernAlbumView'
import { PhotoPreview } from './AlbumFlowClient'

interface ModernPhotoGroupProps {
  date: string
  photos: PhotoPreview[]
  lastPhotoRef?: (node: HTMLDivElement | null) => void
}

const ModernPhotoGroup: React.FC<ModernPhotoGroupProps> = ({
  date,
  photos,
}) => {
  const [selectedPhoto, setSelectedPhoto] = useState<PhotoPreview | null>(null)
  const [isViewerOpen, setIsViewerOpen] = useState(false)

  const currentPhotoIndex = useMemo(() => {
    if (!selectedPhoto) return -1
    return photos.findIndex(photo => photo.id === selectedPhoto.id)
  }, [selectedPhoto, photos])

  const handlePhotoClick = (photo: PhotoPreview) => {
    setSelectedPhoto(photo)
    setIsViewerOpen(true)
  }

  const handlePrevious = () => {
    if (currentPhotoIndex > 0) {
      setSelectedPhoto(photos[currentPhotoIndex - 1])
    }
  }

  const handleNext = () => {
    if (currentPhotoIndex < photos.length - 1) {
      setSelectedPhoto(photos[currentPhotoIndex + 1])
    }
  }

  const handleClose = () => {
    setIsViewerOpen(false)
    setSelectedPhoto(null)
  }

  const formatDateDisplay = (dateString: string) => {
    const date = new Date(dateString)
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)

    const isToday = date.toDateString() === today.toDateString()
    const isYesterday = date.toDateString() === yesterday.toDateString()

    if (isToday) return '今天'
    if (isYesterday) return '昨天'
    
    return date.toLocaleDateString('zh-CN', { 
      year: 'numeric', 
      month: 'long', 
      day: 'numeric' 
    })
  }

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ 
          type: "spring", 
          stiffness: 120, 
          damping: 18
        }}
        className="mb-12"
      >
        {/* 精致日期标题 */}
        <motion.div 
          className="flex items-center mb-6 group"
          whileHover={{ scale: 1.01 }}
          transition={{ type: "spring", stiffness: 400, damping: 10 }}
        >
          <div className="flex items-center space-x-2">
            <div className="p-1.5 bg-gray-100 dark:bg-gray-700 rounded-lg shadow-sm">
              <Calendar className="w-3.5 h-3.5 text-gray-600 dark:text-gray-300" />
            </div>
            <div>
              <h2 className="text-lg font-medium text-gray-800 dark:text-gray-200">
                {formatDateDisplay(date)}
              </h2>
              <p className="text-xs text-gray-400 dark:text-gray-500">
                {photos.length} 张
              </p>
            </div>
          </div>
          
          {/* 精致装饰线条 */}
          <div className="flex-1 ml-4">
            <div className="h-px bg-gray-200 dark:bg-gray-700" />
          </div>
        </motion.div>

        {/* 瀑布流网格 */}
        <MasonryGrid
          photos={photos}
          onPhotoClick={handlePhotoClick}
          className="transition-all duration-300"
        />
      </motion.div>

      {/* 增强版详情查看器 */}
      {selectedPhoto && (
        <EnhancedModernAlbumView
          photo={selectedPhoto}
          isOpen={isViewerOpen}
          onClose={handleClose}
          onPrevious={handlePrevious}
          onNext={handleNext}
          hasPrevious={currentPhotoIndex > 0}
          hasNext={currentPhotoIndex < photos.length - 1}
          totalCount={photos.length}
          currentIndex={currentPhotoIndex}
        />
      )}
    </>
  )
}

export default ModernPhotoGroup 