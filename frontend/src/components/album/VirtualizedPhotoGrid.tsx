"use client"

import React, { useState, useEffect, useRef, useMemo } from 'react'
import { FixedSizeGrid as Grid } from 'react-window'
import { motion } from 'framer-motion'
import LazyImage from './LazyImage'
import { PhotoPreview } from './AlbumFlowClient'

interface VirtualizedPhotoGridProps {
  photos: PhotoPreview[]
  onPhotoClick: (photo: PhotoPreview) => void
  containerHeight?: number
  itemSize?: number
  gap?: number
}

interface GridItemProps {
  columnIndex: number
  rowIndex: number
  style: React.CSSProperties
  data: {
    photos: PhotoPreview[]
    columnsCount: number
    onPhotoClick: (photo: PhotoPreview) => void
    itemSize: number
    gap: number
  }
}

const GridItem: React.FC<GridItemProps> = ({ columnIndex, rowIndex, style, data }) => {
  const { photos, columnsCount, onPhotoClick, gap } = data
  const photoIndex = rowIndex * columnsCount + columnIndex
  const photo = photos[photoIndex]

  if (!photo) return null

  return (
    <div
      style={{
        ...style,
        padding: gap / 2,
      }}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.8 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{
          duration: 0.3,
          delay: photoIndex * 0.02, // 错开动画
        }}
        whileHover={{ scale: 1.05 }}
        className="h-full"
      >
        <LazyImage
          src={photo.url}
          alt={`Photo from ${photo.date}`}
          placeholder={photo.shade}
          aspectRatio={1} // 正方形网格
          onClick={() => onPhotoClick(photo)}
          className="w-full h-full rounded-lg shadow-md hover:shadow-xl transition-shadow duration-300"
          sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
        />
      </motion.div>
    </div>
  )
}

const VirtualizedPhotoGrid: React.FC<VirtualizedPhotoGridProps> = ({
  photos,
  onPhotoClick,
  containerHeight = 600,
  itemSize = 200,
  gap = 16
}) => {
  const [containerWidth, setContainerWidth] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)

  // 计算列数
  const columnsCount = useMemo(() => {
    if (containerWidth === 0) return 1
    return Math.floor(containerWidth / (itemSize + gap))
  }, [containerWidth, itemSize, gap])

  // 计算行数
  const rowsCount = useMemo(() => {
    return Math.ceil(photos.length / columnsCount)
  }, [photos.length, columnsCount])

  // 监听容器尺寸变化
  useEffect(() => {
    const updateSize = () => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.offsetWidth)
      }
    }

    updateSize()
    
    const resizeObserver = new ResizeObserver(updateSize)
    if (containerRef.current) {
      resizeObserver.observe(containerRef.current)
    }

    return () => resizeObserver.disconnect()
  }, [])

  // 准备传递给 Grid 的数据
  const itemData = useMemo(() => ({
    photos,
    columnsCount,
    onPhotoClick,
    itemSize,
    gap
  }), [photos, columnsCount, onPhotoClick, itemSize, gap])

  if (containerWidth === 0) {
    return (
      <div ref={containerRef} className="w-full h-64 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
      </div>
    )
  }

  return (
    <div ref={containerRef} className="w-full">
      <Grid
        columnCount={columnsCount}
        columnWidth={itemSize + gap}
        height={containerHeight}
        rowCount={rowsCount}
        rowHeight={itemSize + gap}
        width={containerWidth}
        itemData={itemData}
        overscanRowCount={2} // 预渲染行数
        overscanColumnCount={1} // 预渲染列数
      >
        {GridItem}
      </Grid>
    </div>
  )
}

export default VirtualizedPhotoGrid 