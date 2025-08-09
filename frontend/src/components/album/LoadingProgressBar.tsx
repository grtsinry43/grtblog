"use client"

import React from 'react'
import { motion } from 'framer-motion'

interface LoadingProgressBarProps {
  progress: number
  isVisible: boolean
  className?: string
}

const LoadingProgressBar: React.FC<LoadingProgressBarProps> = ({
  progress,
  isVisible,
  className = ''
}) => {
  if (!isVisible) return null

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className={`fixed top-0 left-0 right-0 z-50 ${className}`}
    >
      <div className="h-0.5 bg-gray-100 dark:bg-gray-800">
        <motion.div
          className="h-full bg-blue-500"
          initial={{ width: 0 }}
          animate={{ width: `${Math.min(100, Math.max(0, progress))}%` }}
          transition={{ duration: 0.3, ease: "easeOut" }}
        />
      </div>
    </motion.div>
  )
}

export default LoadingProgressBar 