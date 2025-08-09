"use client"

import React, { useRef, useCallback, useEffect, useState } from 'react'
import { motion, PanInfo, useMotionValue, useTransform } from 'framer-motion'

interface MobileGestureHandlerProps {
  children: React.ReactNode
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  onDoubleClick?: () => void
  onPinchZoom?: (scale: number) => void
  enableZoom?: boolean
  enableSwipe?: boolean
  className?: string
}

interface TouchPoint {
  id: number
  x: number
  y: number
}

const MobileGestureHandler: React.FC<MobileGestureHandlerProps> = ({
  children,
  onSwipeLeft,
  onSwipeRight,
  onDoubleClick,
  onPinchZoom,
  enableZoom = true,
  enableSwipe = true,
  className = ''
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const lastTapRef = useRef<number>(0)
  const touchPointsRef = useRef<TouchPoint[]>([])
  const initialDistanceRef = useRef<number>(0)
  const initialScaleRef = useRef<number>(1)
  
  const x = useMotionValue(0)
  const scale = useMotionValue(1)
  const rotate = useTransform(x, [-200, 200], [-15, 15])
  const opacity = useTransform(x, [-300, 0, 300], [0.5, 1, 0.5])

  // 计算两点间距离
  const calculateDistance = useCallback((touch1: TouchPoint, touch2: TouchPoint) => {
    const dx = touch1.x - touch2.x
    const dy = touch1.y - touch2.y
    return Math.sqrt(dx * dx + dy * dy)
  }, [])

  // 处理触摸开始
  const handleTouchStart = useCallback((e: TouchEvent) => {
    e.preventDefault()
    
    const touches = Array.from(e.touches).map(touch => ({
      id: touch.identifier,
      x: touch.clientX,
      y: touch.clientY
    }))
    
    touchPointsRef.current = touches
    
    if (touches.length === 2 && enableZoom) {
      // 双指操作 - 准备缩放
      initialDistanceRef.current = calculateDistance(touches[0], touches[1])
      initialScaleRef.current = scale.get()
    }
    
    if (touches.length === 1) {
      // 单指操作 - 检测双击
      const now = Date.now()
      const timeDiff = now - lastTapRef.current
      
      if (timeDiff < 300 && timeDiff > 50) {
        // 双击
        onDoubleClick?.()
      }
      
      lastTapRef.current = now
    }
  }, [enableZoom, calculateDistance, scale, onDoubleClick])

  // 处理触摸移动
  const handleTouchMove = useCallback((e: TouchEvent) => {
    e.preventDefault()
    
    const touches = Array.from(e.touches).map(touch => ({
      id: touch.identifier,
      x: touch.clientX,
      y: touch.clientY
    }))
    
    if (touches.length === 2 && enableZoom && touchPointsRef.current.length === 2) {
      // 双指缩放
      const currentDistance = calculateDistance(touches[0], touches[1])
      const scaleRatio = currentDistance / initialDistanceRef.current
      const newScale = Math.max(0.5, Math.min(3, initialScaleRef.current * scaleRatio))
      
      scale.set(newScale)
      onPinchZoom?.(newScale)
    }
  }, [enableZoom, calculateDistance, scale, onPinchZoom])

  // 处理触摸结束
  const handleTouchEnd = useCallback(() => {
    touchPointsRef.current = []
    
    // 重置缩放动画
    if (scale.get() !== 1) {
      scale.set(1)
    }
  }, [scale])

  // 处理拖拽结束
  const handleDragEnd = useCallback((_: never, info: PanInfo) => {
    const threshold = 100
    const velocity = 500
    
    if (!enableSwipe) return
    
    if (Math.abs(info.offset.x) > threshold || Math.abs(info.velocity.x) > velocity) {
      if (info.offset.x > 0) {
        onSwipeRight?.()
      } else {
        onSwipeLeft?.()
      }
    }
    
    // 重置位置
    x.set(0)
  }, [enableSwipe, onSwipeLeft, onSwipeRight, x])

  // 绑定触摸事件
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    container.addEventListener('touchstart', handleTouchStart, { passive: false })
    container.addEventListener('touchmove', handleTouchMove, { passive: false })
    container.addEventListener('touchend', handleTouchEnd, { passive: false })

    return () => {
      container.removeEventListener('touchstart', handleTouchStart)
      container.removeEventListener('touchmove', handleTouchMove)
      container.removeEventListener('touchend', handleTouchEnd)
    }
  }, [handleTouchStart, handleTouchMove, handleTouchEnd])

  return (
    <motion.div
      ref={containerRef}
      className={`relative ${className}`}
      style={{
        x: enableSwipe ? x : 0,
        scale,
        rotate: enableSwipe ? rotate : 0,
        opacity: enableSwipe ? opacity : 1,
      }}
      drag={enableSwipe ? "x" : false}
      dragConstraints={{ left: 0, right: 0 }}
      dragElastic={0.2}
      onDragEnd={handleDragEnd}
      whileTap={{ scale: 0.95 }}
    >
      {children}
      
      {/* 手势指示器 */}
      <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 flex space-x-2 opacity-30 pointer-events-none">
        {enableSwipe && (
          <>
            <div className="w-2 h-2 bg-white rounded-full animate-pulse" />
            <div className="w-2 h-2 bg-white rounded-full animate-pulse" style={{ animationDelay: '0.2s' }} />
            <div className="w-2 h-2 bg-white rounded-full animate-pulse" style={{ animationDelay: '0.4s' }} />
          </>
        )}
      </div>
    </motion.div>
  )
}

// 移动端优化Hook
export const useMobileOptimization = () => {
  const [isMobile, setIsMobile] = useState(false)
  const [isLandscape, setIsLandscape] = useState(false)
  
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
      setIsLandscape(window.innerWidth > window.innerHeight)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    window.addEventListener('orientationchange', checkMobile)
    
    return () => {
      window.removeEventListener('resize', checkMobile)
      window.removeEventListener('orientationchange', checkMobile)
    }
  }, [])
  
  return { isMobile, isLandscape }
}

export default MobileGestureHandler 