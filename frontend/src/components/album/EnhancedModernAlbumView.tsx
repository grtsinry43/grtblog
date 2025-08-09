"use client"

import React, {useState, useEffect, useCallback, useRef} from 'react'
import {motion, AnimatePresence} from 'framer-motion'
import {createPortal} from 'react-dom'
import {
    X,
    ChevronLeft,
    ChevronRight,
    Info,
    Download,
    Share2,
    MapPin,
    Calendar,
    Camera,
    Clock,
    ZoomIn,
    ZoomOut
} from 'lucide-react'
import LazyImage from './LazyImage'
import MobileGestureHandler, {useMobileOptimization} from './MobileGestureHandler'
import {PhotoPreview} from './AlbumFlowClient'

interface EnhancedModernAlbumViewProps {
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

const EnhancedModernAlbumView: React.FC<EnhancedModernAlbumViewProps> = ({
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
    const [showControls, setShowControls] = useState(true)
    const [zoomLevel, setZoomLevel] = useState(1)
    const [imagePosition, setImagePosition] = useState({x: 0, y: 0})

    const {isMobile} = useMobileOptimization()
    const imageRef = useRef<HTMLDivElement>(null)
    const timerRef = useRef<NodeJS.Timeout | null>(null)

    // 自动隐藏控件
    useEffect(() => {
        const resetTimer = () => {
            if (timerRef.current) clearTimeout(timerRef.current)
            setShowControls(true)
            if (!isMobile) {
                timerRef.current = setTimeout(() => setShowControls(false), 3000)
            }
        }

        if (isOpen) {
            resetTimer()
            if (!isMobile) {
                const handleMouseMove = () => resetTimer()
                window.addEventListener('mousemove', handleMouseMove)
                return () => {
                    window.removeEventListener('mousemove', handleMouseMove)
                    if (timerRef.current) clearTimeout(timerRef.current)
                }
            }
        }
    }, [isOpen, isMobile])

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
                case '+':
                case '=':
                    handleZoomIn()
                    break
                case '-':
                    handleZoomOut()
                    break
            }
        }

        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [isOpen, hasPrevious, hasNext, onClose, onPrevious, onNext, showInfo])

    // 重置状态当照片改变时
    useEffect(() => {
        setZoomLevel(1)
        setImagePosition({x: 0, y: 0})
        setIsImageLoaded(false)
    }, [photo.id])

    const handleZoomIn = () => {
        setZoomLevel(prev => Math.min(prev * 1.5, 4))
    }

    const handleZoomOut = () => {
        setZoomLevel(prev => Math.max(prev / 1.5, 0.5))
    }

    const handleDoubleClick = useCallback(() => {
        if (zoomLevel > 1) {
            setZoomLevel(1)
            setImagePosition({x: 0, y: 0})
        } else {
            setZoomLevel(2)
        }
    }, [zoomLevel])

    const handleSwipeLeft = useCallback(() => {
        if (hasNext) onNext()
    }, [hasNext, onNext])

    const handleSwipeRight = useCallback(() => {
        if (hasPrevious) onPrevious()
    }, [hasPrevious, onPrevious])

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
                    text: photo.description || ` 拍摄于 ${formatDate(photo.date).date}`,
                    url: photo.url
                })
            } catch (error) {
                console.log('分享失败:', error)
            }
        } else {
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

    const {date, time} = formatDate(photo.date)

    return createPortal(
        <AnimatePresence>
            <motion.div
                initial={{opacity: 0}}
                animate={{opacity: 1}}
                exit={{opacity: 0}}
                transition={{duration: 0.3}}
                className="fixed inset-0 z-50 bg-black/95 backdrop-blur-sm"
                onClick={onClose}
            >
                {/* 精致顶部工具栏 */}
                <motion.div
                    initial={{y: -50, opacity: 0}}
                    animate={{
                        y: showControls ? 0 : -50,
                        opacity: showControls ? 1 : 0
                    }}
                    transition={{duration: 0.3}}
                    className="absolute top-0 left-0 right-0 z-60 bg-gradient-to-b from-black/30 to-transparent p-3 sm:p-4"
                >
                    <div className="flex items-center justify-between text-white">
                        <div className="flex items-center space-x-3">
                            <button
                                onClick={onClose}
                                className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                            >
                                <X size={18}/>
                            </button>
                            {totalCount && currentIndex !== undefined && (
                                <span className="text-xs opacity-75">
                  {currentIndex + 1} / {totalCount}
                </span>
                            )}
                        </div>

                        <div className="flex items-center space-x-1">
                            {!isMobile && (
                                <>
                                    <button
                                        onClick={handleZoomOut}
                                        className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                                        disabled={zoomLevel <= 0.5}
                                    >
                                        <ZoomOut size={16}/>
                                    </button>
                                    <span
                                        className="text-xs px-1 min-w-[3rem] text-center">{Math.round(zoomLevel * 100)}%</span>
                                    <button
                                        onClick={handleZoomIn}
                                        className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                                        disabled={zoomLevel >= 4}
                                    >
                                        <ZoomIn size={16}/>
                                    </button>
                                    <div className="w-px h-4 bg-white/20 mx-1"/>
                                </>
                            )}
                            <button
                                onClick={() => setShowInfo(!showInfo)}
                                className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                            >
                                <Info size={16}/>
                            </button>
                            <button
                                onClick={handleShare}
                                className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                            >
                                <Share2 size={16}/>
                            </button>
                            <button
                                onClick={handleDownload}
                                className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                            >
                                <Download size={16}/>
                            </button>
                        </div>
                    </div>
                </motion.div>

                {/* 主图区域 */}
                <div
                    className="flex items-center justify-center h-full p-4 sm:p-8"
                    onClick={(e) => e.stopPropagation()}
                >
                    {isMobile ? (
                        <MobileGestureHandler
                            onSwipeLeft={handleSwipeLeft}
                            onSwipeRight={handleSwipeRight}
                            onDoubleClick={handleDoubleClick}
                            onPinchZoom={setZoomLevel}
                            className="w-full h-full flex items-center justify-center"
                        >
                            <motion.div
                                ref={imageRef}
                                style={{
                                    scale: zoomLevel,
                                    x: imagePosition.x,
                                    y: imagePosition.y
                                }}
                                transition={{type: "spring", stiffness: 300, damping: 30}}
                                className="relative max-w-[95vw] max-h-[85vh]"
                            >
                                <div
                                    className="relative bg-center bg-cover rounded-lg overflow-hidden shadow-xl"
                                    style={{backgroundColor: photo.shade}}
                                >
                                    <LazyImage
                                        src={photo.url}
                                        alt={`Photo from ${photo.date}`}
                                        placeholder={photo.shade}
                                        priority={true}
                                        onLoad={() => setIsImageLoaded(true)}
                                        className="w-full h-full object-contain"
                                    />
                                </div>
                            </motion.div>
                        </MobileGestureHandler>
                    ) : (
                        <motion.div
                            ref={imageRef}
                            drag={zoomLevel > 1}
                            dragConstraints={{
                                left: -100, right: 100, top: -100, bottom: 100
                            }}
                            dragElastic={0.1}
                            onDoubleClick={handleDoubleClick}
                            style={{
                                scale: zoomLevel,
                                cursor: zoomLevel > 1 ? 'grab' : 'zoom-in'
                            }}
                            transition={{type: "spring", stiffness: 300, damping: 30}}
                            className="relative max-w-[90vw] max-h-[80vh]"
                        >
                            <div
                                className="relative bg-center bg-cover rounded-lg overflow-hidden shadow-xl"
                                style={{backgroundColor: photo.shade}}
                            >
                                <LazyImage
                                    src={photo.url}
                                    alt={`Photo from ${photo.date}`}
                                    placeholder={photo.shade}
                                    priority={true}
                                    onLoad={() => setIsImageLoaded(true)}
                                    className="w-full h-full object-contain"
                                />
                            </div>
                        </motion.div>
                    )}
                </div>

                {/* 精致导航按钮 - 仅桌面端显示 */}
                <AnimatePresence>
                    {!isMobile && showControls && (
                        <>
                            {hasPrevious && (
                                <motion.button
                                    initial={{x: -50, opacity: 0}}
                                    animate={{x: 0, opacity: 1}}
                                    exit={{x: -50, opacity: 0}}
                                    onClick={onPrevious}
                                    className="absolute left-3 top-1/2 -translate-y-1/2 p-2.5 bg-black/20 hover:bg-black/40 text-white rounded-lg backdrop-blur-sm transition-all"
                                >
                                    <ChevronLeft size={20}/>
                                </motion.button>
                            )}

                            {hasNext && (
                                <motion.button
                                    initial={{x: 50, opacity: 0}}
                                    animate={{x: 0, opacity: 1}}
                                    exit={{x: 50, opacity: 0}}
                                    onClick={onNext}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 p-2.5 bg-black/20 hover:bg-black/40 text-white rounded-lg backdrop-blur-sm transition-all"
                                >
                                    <ChevronRight size={20}/>
                                </motion.button>
                            )}
                        </>
                    )}
                </AnimatePresence>

                {/* 精致信息面板 */}
                <AnimatePresence>
                    {showInfo && (
                        <motion.div
                            initial={{x: '100%'}}
                            animate={{x: 0}}
                            exit={{x: '100%'}}
                            transition={{type: "spring", stiffness: 300, damping: 30}}
                            className={`absolute right-0 top-0 bottom-0 bg-black/80 backdrop-blur-xl text-white overflow-y-auto ${
                                isMobile ? 'w-full' : 'w-72'
                            }`}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <div className="p-4 space-y-4">
                                <div className="flex items-center justify-between">
                                    <h3 className="text-lg font-medium"> 照片信息 </h3>
                                    <button
                                        onClick={() => setShowInfo(false)}
                                        className="p-1 hover:bg-white/10 rounded"
                                    >
                                        <X size={16}/>
                                    </button>
                                </div>

                                <div className="space-y-3">
                                    <div className="flex items-start space-x-2">
                                        <Calendar size={16} className="text-blue-400 mt-0.5"/>
                                        <div>
                                            <div className="text-sm font-medium">{date}</div>
                                            <div className="text-xs text-gray-300 flex items-center mt-0.5">
                                                <Clock size={12} className="mr-1"/>
                                                {time}
                                            </div>
                                        </div>
                                    </div>

                                    {photo.location && (
                                        <div className="flex items-start space-x-2">
                                            <MapPin size={16} className="text-green-400 mt-0.5"/>
                                            <div>
                                                <div className="text-sm font-medium"> 拍摄地点</div>
                                                <div className="text-xs text-gray-300">{photo.location}</div>
                                            </div>
                                        </div>
                                    )}

                                    {photo.device && (
                                        <div className="flex items-start space-x-2">
                                            <Camera size={16} className="text-purple-400 mt-0.5"/>
                                            <div>
                                                <div className="text-sm font-medium"> 拍摄设备</div>
                                                <div className="text-xs text-gray-300">{photo.device}</div>
                                            </div>
                                        </div>
                                    )}

                                    {photo.description && (
                                        <div className="pt-3 border-t border-white/10">
                                            <div className="text-sm font-medium mb-1"> 描述</div>
                                            <div className="text-xs text-gray-300 leading-relaxed">
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

export default EnhancedModernAlbumView 