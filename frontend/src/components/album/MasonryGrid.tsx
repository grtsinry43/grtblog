"use client"

import React, {useState, useEffect, useRef, useMemo} from 'react'
import {motion, AnimatePresence} from 'framer-motion'
import LazyImage from './LazyImage'
import {PhotoPreview} from './AlbumFlowClient'

interface MasonryGridProps {
    photos: PhotoPreview[]
    onPhotoClick: (photo: PhotoPreview) => void
    columns?: number
    gap?: number
    className?: string
}

interface PhotoWithDimensions extends PhotoPreview {
    width?: number
    height?: number
    aspectRatio?: number
}

const MasonryGrid: React.FC<MasonryGridProps> = ({
                                                     photos,
                                                     onPhotoClick,
                                                     columns = 3,
                                                     gap = 12,
                                                     className = ''
                                                 }) => {
    const [photosWithDimensions, setPhotosWithDimensions] = useState<PhotoWithDimensions[]>([])
    const [, setLoadedPhotos] = useState<Set<string>>(new Set())
    const [columnHeights, setColumnHeights] = useState<number[]>([])
    const containerRef = useRef<HTMLDivElement>(null)

    // 根据屏幕大小调整列数
    const responsiveColumns = useMemo(() => {
        if (typeof window === 'undefined') return columns

        const width = window.innerWidth
        if (width < 640) return 1      // sm
        if (width < 768) return 2      // md
        if (width < 1024) return 2     // lg
        if (width < 1280) return 3     // xl
        return Math.min(columns, 4)    // 2xl+
    }, [columns])

    // 初始化列高度数组
    useEffect(() => {
        setColumnHeights(new Array(responsiveColumns).fill(0))
    }, [responsiveColumns])

    // 获取图片尺寸信息
    useEffect(() => {
        const loadImageDimensions = async () => {
            const photosWithDims = await Promise.all(
                photos.map(async (photo) => {
                    // 检查类型安全的 width 和 height 属性
                    const photoWithDims = photo as PhotoWithDimensions
                    if (photoWithDims.width && photoWithDims.height) {
                        return {
                            ...photo,
                            width: photoWithDims.width,
                            height: photoWithDims.height,
                            aspectRatio: photoWithDims.width / photoWithDims.height
                        }
                    }

                    // 否则加载图片获取尺寸
                    return new Promise<PhotoWithDimensions>((resolve) => {
                        const img = new Image()
                        img.onload = () => {
                            resolve({
                                ...photo,
                                width: img.naturalWidth,
                                height: img.naturalHeight,
                                aspectRatio: img.naturalWidth / img.naturalHeight
                            })
                        }
                        img.onerror = () => {
                            // 错误时使用默认比例
                            resolve({
                                ...photo,
                                aspectRatio: 4 / 3
                            })
                        }
                        img.src = photo.url
                    })
                })
            )

            setPhotosWithDimensions(photosWithDims)
        }

        if (photos.length > 0) {
            loadImageDimensions()
        }
    }, [photos])

    // 计算每张图片的位置
    const calculateLayout = useMemo(() => {
        if (!photosWithDimensions.length || !columnHeights.length) return []

        const layout: Array<{
            photo: PhotoWithDimensions
            x: number
            y: number
            width: number
            height: number
            column: number
        }> = []

        const heights = [...columnHeights]
        const columnWidth = containerRef.current
            ? (containerRef.current.offsetWidth - gap * (responsiveColumns - 1)) / responsiveColumns
            : 300

        photosWithDimensions.forEach((photo) => {
            // 找到最短的列
            const shortestColumnIndex = heights.indexOf(Math.min(...heights))

            // 计算图片高度
            const aspectRatio = photo.aspectRatio || 4 / 3
            const imageHeight = columnWidth / aspectRatio

            // 添加到布局
            layout.push({
                photo,
                x: shortestColumnIndex * (columnWidth + gap),
                y: heights[shortestColumnIndex],
                width: columnWidth,
                height: imageHeight,
                column: shortestColumnIndex
            })

            // 更新列高度
            heights[shortestColumnIndex] += imageHeight + gap
        })

        return layout
    }, [photosWithDimensions, columnHeights, gap, responsiveColumns])

    // 容器总高度
    const containerHeight = useMemo(() => {
        if (!calculateLayout.length) return 0
        return Math.max(...calculateLayout.map(item => item.y + item.height))
    }, [calculateLayout])

    const handlePhotoLoad = (photoId: string) => {
        setLoadedPhotos(prev => new Set([...prev, photoId]))
    }

    return (
        <div
            ref={containerRef}
            className={`relative ${className}`}
            style={{height: containerHeight}}
        >
            <AnimatePresence>
                {calculateLayout.map((item, index) => (
                    <motion.div
                        key={item.photo.id}
                        initial={{
                            opacity: 0,
                            scale: 0.9,
                            y: 20
                        }}
                        animate={{
                            opacity: 1,
                            scale: 1,
                            y: 0,
                            x: item.x,
                            position: 'absolute',
                            top: item.y,
                            width: item.width,
                            height: item.height
                        }}
                        exit={{
                            opacity: 0,
                            scale: 0.9,
                            transition: {duration: 0.2}
                        }}
                        transition={{
                            type: "spring",
                            stiffness: 300,
                            damping: 25,
                            delay: index * 0.03 // 错开动画时间
                        }}
                        whileHover={{
                            scale: 1.02,
                            zIndex: 10,
                            transition: {duration: 0.2}
                        }}
                        className="cursor-pointer group"
                    >
                        <LazyImage
                            src={item.photo.url}
                            alt={`Photo from ${item.photo.date}`}
                            placeholder={item.photo.shade}
                            aspectRatio={item.photo.aspectRatio}
                            onLoad={() => handlePhotoLoad(item.photo.id)}
                            onClick={() => onPhotoClick(item.photo)}
                            className="w-full h-full rounded-lg shadow-sm hover:shadow-md transition-all duration-300"
                            sizes={`(max-width: 640px) 100vw, (max-width: 768px) 50vw, (max-width: 1024px) 33vw, 25vw`}
                        />

                        {/* 精致图片信息悬浮层 */}
                        <motion.div
                            initial={{opacity: 0}}
                            whileHover={{opacity: 1}}
                            transition={{duration: 0.2}}
                            className="absolute inset-0 bg-gradient-to-t from-black/40 via-transparent to-transparent rounded-lg flex flex-col justify-end p-3 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                        >
                            <div className="text-white">
                                {item.photo.location && (
                                    <div className="text-xs opacity-90 mb-1">
                                        📍 {item.photo.location}
                                    </div>
                                )}
                                <div className="text-xs opacity-75">
                                    {new Date(item.photo.date).toLocaleDateString()}
                                </div>
                            </div>
                        </motion.div>
                    </motion.div>
                ))}
            </AnimatePresence>
        </div>
    )
}

export default MasonryGrid 