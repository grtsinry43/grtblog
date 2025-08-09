"use client";

import React, {useState, useCallback, useEffect, useRef} from 'react'
import Image from "next/image"
import {X} from "lucide-react"
import {createPortal} from "react-dom"
import {motion, AnimatePresence} from "framer-motion"
import useIsMobile from '@/hooks/useIsMobile'

interface ZoomedImageProps {
    src: string
    alt: string
    onClose: () => void
    // 触发元素的位置信息，用于动画起始位置
    triggerElement?: {
        x: number
        y: number
        width: number
        height: number
    }
}

const ZoomedImage: React.FC<ZoomedImageProps> = ({src, alt, onClose, triggerElement}) => {
    const [scale, setScale] = useState(1)
    const [position, setPosition] = useState({x: 0, y: 0})
    const [isDragging, setIsDragging] = useState(false)
    const [dragStart, setDragStart] = useState({x: 0, y: 0})
    const [imageLoaded, setImageLoaded] = useState(false)
    const [lastTap, setLastTap] = useState(0)
    const [initialDistance, setInitialDistance] = useState(0)
    const [initialScale, setInitialScale] = useState(1)
    
    const isMobile = useIsMobile()
    const imageRef = useRef<HTMLDivElement>(null)
    const containerRef = useRef<HTMLDivElement>(null)

    // 重置位置和缩放
    const resetTransform = useCallback(() => {
        setScale(1)
        setPosition({x: 0, y: 0})
    }, [])

    // 计算两点间距离
    const getDistance = useCallback((touches: React.TouchList) => {
        if (touches.length < 2) return 0
        const touch1 = touches[0]
        const touch2 = touches[1]
        return Math.sqrt(
            Math.pow(touch2.clientX - touch1.clientX, 2) +
            Math.pow(touch2.clientY - touch1.clientY, 2)
        )
    }, [])

    // 使用原生事件监听器处理滚轮，避免被动监听器问题
    useEffect(() => {
        if (isMobile) return

        const handleWheel = (e: WheelEvent) => {
            e.preventDefault()
            e.stopPropagation()
            
            const rect = imageRef.current?.getBoundingClientRect()
            if (!rect) return

            const deltaScale = -e.deltaY * 0.002
            const newScale = Math.max(0.5, Math.min(scale + deltaScale, 5))
            
            // 以鼠标位置为中心缩放
            const mouseX = e.clientX - rect.left - rect.width / 2
            const mouseY = e.clientY - rect.top - rect.height / 2
            
            const deltaX = mouseX * (newScale / scale - 1)
            const deltaY = mouseY * (newScale / scale - 1)
            
            setScale(newScale)
            setPosition(prev => ({
                x: prev.x - deltaX,
                y: prev.y - deltaY
            }))
        }

        const container = containerRef.current
        if (container) {
            container.addEventListener('wheel', handleWheel, { passive: false })
            return () => {
                container.removeEventListener('wheel', handleWheel)
            }
        }
    }, [scale, isMobile])

    // 鼠标事件处理（仅桌面端）
    const handleMouseDown = useCallback((e: React.MouseEvent) => {
        if (isMobile || e.button !== 0) return
        setIsDragging(true)
        setDragStart({x: e.clientX - position.x, y: e.clientY - position.y})
        e.preventDefault()
    }, [position, isMobile])

    const handleMouseMove = useCallback((e: React.MouseEvent) => {
        if (isMobile || !isDragging) return
        setPosition({
            x: e.clientX - dragStart.x,
            y: e.clientY - dragStart.y,
        })
    }, [isDragging, dragStart, isMobile])

    const handleMouseUp = useCallback(() => {
        if (!isMobile) {
            setIsDragging(false)
        }
    }, [isMobile])

    // 触摸事件处理（仅移动端）
    const handleTouchStart = useCallback((e: React.TouchEvent) => {
        if (!isMobile) return
        
        const now = Date.now()
        const touch = e.touches[0]
        
        if (e.touches.length === 1) {
            // 双击检测
            if (now - lastTap < 300) {
                e.preventDefault()
                // 双击放大/缩小
                const newScale = scale > 1.5 ? 1 : 2.5
                setScale(newScale)
                if (newScale === 1) {
                    setPosition({x: 0, y: 0})
                }
            } else {
                // 单击开始拖拽
                setIsDragging(true)
                setDragStart({
                    x: touch.clientX - position.x,
                    y: touch.clientY - position.y
                })
            }
            setLastTap(now)
        } else if (e.touches.length === 2) {
            // 双指缩放开始
            e.preventDefault()
            const distance = getDistance(e.touches)
            setInitialDistance(distance)
            setInitialScale(scale)
            setIsDragging(false)
        }
    }, [scale, position, lastTap, isMobile, getDistance])

    const handleTouchMove = useCallback((e: React.TouchEvent) => {
        if (!isMobile) return
        
        if (e.touches.length === 1 && isDragging) {
            e.preventDefault()
            const touch = e.touches[0]
            setPosition({
                x: touch.clientX - dragStart.x,
                y: touch.clientY - dragStart.y,
            })
        } else if (e.touches.length === 2) {
            // 双指缩放
            e.preventDefault()
            const distance = getDistance(e.touches)
            if (initialDistance > 0) {
                const scaleRatio = distance / initialDistance
                const newScale = Math.max(0.5, Math.min(initialScale * scaleRatio, 5))
                setScale(newScale)
            }
        }
    }, [isDragging, dragStart, isMobile, getDistance, initialDistance, initialScale])

    const handleTouchEnd = useCallback(() => {
        if (isMobile) {
            setIsDragging(false)
            setInitialDistance(0)
        }
    }, [isMobile])

    // ESC键关闭
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                onClose()
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        if (!isMobile) {
            document.addEventListener('mouseup', handleMouseUp)
        }
        
        return () => {
            document.removeEventListener('keydown', handleKeyDown)
            if (!isMobile) {
                document.removeEventListener('mouseup', handleMouseUp)
            }
        }
    }, [onClose, handleMouseUp, isMobile])

    // 点击背景关闭
    const handleBackdropClick = useCallback((e: React.MouseEvent) => {
        if (e.target === e.currentTarget) {
            onClose()
        }
    }, [onClose])

    // 防止页面滚动
    useEffect(() => {
        document.body.style.overflow = 'hidden'
        return () => {
            document.body.style.overflow = 'unset'
        }
    }, [])

    // 计算初始动画位置
    const getInitialPosition = () => {
        if (!triggerElement) {
            return {
                x: window.innerWidth / 2,
                y: window.innerHeight / 2,
                scale: 0.8
            }
        }

        const elementCenterX = triggerElement.x + triggerElement.width / 2
        const elementCenterY = triggerElement.y + triggerElement.height / 2

        return {
            x: elementCenterX,
            y: elementCenterY,
            scale: Math.min(triggerElement.width / 400, triggerElement.height / 300, 0.3)
        }
    }

    const initialPos = getInitialPosition()

    return createPortal(
        <AnimatePresence>
            <motion.div
                ref={containerRef}
                initial={{opacity: 0}}
                animate={{opacity: 1}}
                exit={{opacity: 0}}
                transition={{duration: 0.25, ease: "easeOut"}}
                className="fixed inset-0 bg-black/85 flex items-center justify-center z-50 backdrop-blur-sm"
                onClick={handleBackdropClick}
                onMouseDown={handleMouseDown}
                onMouseMove={handleMouseMove}
                onMouseUp={handleMouseUp}
                onTouchStart={handleTouchStart}
                onTouchMove={handleTouchMove}
                onTouchEnd={handleTouchEnd}
            >
                <motion.div
                    ref={imageRef}
                    initial={{
                        x: triggerElement ? initialPos.x - window.innerWidth / 2 : 0,
                        y: triggerElement ? initialPos.y - window.innerHeight / 2 : 0,
                        scale: initialPos.scale,
                        opacity: 0
                    }}
                    animate={{
                        x: position.x,
                        y: position.y,
                        scale: scale,
                        opacity: 1
                    }}
                    exit={{
                        x: triggerElement ? initialPos.x - window.innerWidth / 2 : 0,
                        y: triggerElement ? initialPos.y - window.innerHeight / 2 : 0,
                        scale: initialPos.scale,
                        opacity: 0
                    }}
                    transition={{
                        duration: 0.4,
                        ease: [0.25, 0.46, 0.45, 0.94],
                        opacity: { duration: 0.25 }
                    }}
                    className="relative max-w-[95vw] max-h-[95vh] w-auto h-auto"
                    style={{
                        cursor: isDragging ? "grabbing" : scale > 1 ? "grab" : "default",
                        touchAction: "none",
                    }}
                >
                    <motion.div
                        layout
                        className="relative rounded-lg overflow-hidden shadow-2xl bg-white dark:bg-gray-800"
                    >
                        <Image
                            src={src || "/placeholder.svg"}
                            alt={alt}
                            width={1200}
                            height={800}
                            className="max-w-none w-auto h-auto max-h-[95vh] object-contain"
                            style={{
                                opacity: imageLoaded ? 1 : 0,
                                transition: 'opacity 0.3s ease',
                            }}
                            onLoad={() => setImageLoaded(true)}
                            draggable={false}
                            unoptimized
                        />
                        
                        {/* Loading placeholder */}
                        {!imageLoaded && (
                            <div className="absolute inset-0 bg-gray-200 dark:bg-gray-700 animate-pulse rounded-lg flex items-center justify-center min-h-[200px] min-w-[200px]">
                                <div className="text-gray-400">加载中...</div>
                            </div>
                        )}
                    </motion.div>
                </motion.div>

                {/* 关闭按钮 */}
                <motion.button
                    initial={{opacity: 0, scale: 0.5}}
                    animate={{opacity: 1, scale: 1}}
                    exit={{opacity: 0, scale: 0.5}}
                    transition={{delay: 0.2, duration: 0.25, ease: "backOut"}}
                    onClick={onClose}
                    className="absolute top-4 right-4 bg-white/20 hover:bg-white/30 backdrop-blur-sm rounded-full p-3 transition-all duration-200 z-10 group"
                    aria-label="关闭图片查看器"
                >
                    <X className="w-5 h-5 text-white group-hover:scale-110 transition-transform duration-200"/>
                </motion.button>

                {/* 重置按钮 */}
                {(scale !== 1 || position.x !== 0 || position.y !== 0) && (
                    <motion.button
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        exit={{opacity: 0, y: 20}}
                        transition={{delay: 0.3, duration: 0.25}}
                        onClick={resetTransform}
                        className="absolute bottom-4 left-1/2 transform -translate-x-1/2 bg-white/20 hover:bg-white/30 backdrop-blur-sm rounded-full px-4 py-2 text-white text-sm transition-all duration-200 border border-white/10"
                    >
                        重置视图
                    </motion.button>
                )}

                {/* 移动端操作提示 */}
                {isMobile && (
                    <motion.div
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        exit={{opacity: 0, y: 20}}
                        transition={{delay: 0.5, duration: 0.25}}
                        className="absolute bottom-4 right-4 bg-white/20 backdrop-blur-sm rounded-lg px-3 py-2 text-white text-xs border border-white/10"
                    >
                        双击放大 • 双指缩放 • 拖拽移动
                    </motion.div>
                )}

                {/* 桌面端操作提示 */}
                {!isMobile && scale === 1 && (
                    <motion.div
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        exit={{opacity: 0, y: 20}}
                        transition={{delay: 0.5, duration: 0.25}}
                        className="absolute bottom-4 right-4 bg-white/20 backdrop-blur-sm rounded-lg px-3 py-2 text-white text-xs border border-white/10"
                    >
                        滚轮缩放 • 拖拽移动 • ESC 关闭
                    </motion.div>
                )}
            </motion.div>
        </AnimatePresence>,
        document.body,
    )
}

export default ZoomedImage

