"use client"

import type React from "react"
import { useEffect } from "react"
import { createPortal } from "react-dom"
import { motion, useAnimation, type Variants, AnimatePresence } from "framer-motion"
import { CheckCircledIcon, ExclamationTriangleIcon, CrossCircledIcon, InfoCircledIcon } from "@radix-ui/react-icons"

export type NotificationType = "success" | "warning" | "error" | "info"

interface NotificationProps {
    message: string
    type: NotificationType
    onClose: () => void
}

const containerVariants: Variants = {
    initial: {
        opacity: 0,
        y: -60,
        scale: 0.9
    },
    animate: {
        opacity: 1,
        y: 0,
        scale: 1,
        transition: {
            type: "spring",
            duration: 0.6,
            bounce: 0.4,
            stiffness: 300,
            damping: 20
        }
    },
    exit: {
        opacity: 0,
        y: -40,
        scale: 0.95,
        transition: {
            duration: 0.3,
            ease: [0.4, 0, 0.2, 1]
        }
    }
}

const contentVariants: Variants = {
    initial: { 
        opacity: 0,
        scale: 0.95
    },
    animate: {
        opacity: 1,
        scale: 1,
        transition: {
            delay: 0.1,
            duration: 0.4,
            type: "spring",
            bounce: 0.3
        }
    }
}

const iconVariants: Variants = {
    initial: { 
        scale: 0,
        rotate: -90
    },
    animate: {
        scale: 1,
        rotate: 0,
        transition: {
            delay: 0.2,
            type: "spring",
            duration: 0.5,
            bounce: 0.6
        }
    }
}

const textVariants: Variants = {
    initial: { 
        opacity: 0,
        x: 10
    },
    animate: {
        opacity: 1,
        x: 0,
        transition: {
            delay: 0.3,
            duration: 0.3,
            ease: "easeOut"
        }
    }
}

const closeButtonVariants: Variants = {
    initial: { 
        opacity: 0,
        scale: 0
    },
    animate: {
        opacity: 1,
        scale: 1,
        transition: {
            delay: 0.4,
            type: "spring",
            duration: 0.4,
            bounce: 0.5
        }
    }
}

const typeConfig = {
    success: {
        icon: CheckCircledIcon,
        iconColor: "text-green-600 dark:text-green-400",
        bgColor: "bg-green-50 dark:bg-green-950/30",
        borderColor: "border-green-200 dark:border-green-800"
    },
    warning: {
        icon: ExclamationTriangleIcon,
        iconColor: "text-amber-600 dark:text-amber-400",
        bgColor: "bg-amber-50 dark:bg-amber-950/30",
        borderColor: "border-amber-200 dark:border-amber-800"
    },
    error: {
        icon: CrossCircledIcon,
        iconColor: "text-red-600 dark:text-red-400",
        bgColor: "bg-red-50 dark:bg-red-950/30",
        borderColor: "border-red-200 dark:border-red-800"
    },
    info: {
        icon: InfoCircledIcon,
        iconColor: "text-blue-600 dark:text-blue-400",
        bgColor: "bg-blue-50 dark:bg-blue-950/30",
        borderColor: "border-blue-200 dark:border-blue-800"
    }
}

const Notification: React.FC<NotificationProps> = ({ message, type, onClose }) => {
    const controls = useAnimation()
    const { icon: Icon, iconColor, bgColor, borderColor } = typeConfig[type]

    useEffect(() => {
        controls.start("animate")
        
        const timer = setTimeout(() => {
            controls.start("exit").then(onClose)
        }, 4000)

        return () => clearTimeout(timer)
    }, [controls, onClose])

    const notificationContent = (
        <AnimatePresence>
            <motion.div
                className="fixed inset-0 pointer-events-none z-[9999] flex items-start justify-center pt-20"
                initial="initial"
                animate={controls}
                exit="exit"
                variants={containerVariants}
            >
                <motion.div
                    className={`
                        relative rounded-xl border ${borderColor} ${bgColor}
                        backdrop-blur-sm shadow-lg shadow-black/5 dark:shadow-black/20
                        pointer-events-auto min-w-80 max-w-md overflow-hidden
                        hover:shadow-xl hover:shadow-black/10 dark:hover:shadow-black/30
                        transition-shadow duration-300
                    `}
                    variants={contentVariants}
                    whileHover={{ 
                        scale: 1.02,
                        transition: { type: "spring", bounce: 0.4, duration: 0.3 }
                    }}
                    whileTap={{ scale: 0.98 }}
                >
                    <div className="flex items-center gap-3 px-4 py-3">
                        {/* Icon */}
                        <motion.div 
                            className="flex-shrink-0"
                            variants={iconVariants}
                        >
                            <Icon className={`${iconColor} w-5 h-5`} width={20} height={20} />
                        </motion.div>

                        {/* Message */}
                        <motion.div 
                            className="flex-1 min-w-0"
                            variants={textVariants}
                        >
                            <p className="text-sm text-gray-800 dark:text-gray-200 leading-relaxed">
                                {message}
                            </p>
                        </motion.div>

                        {/* Close button */}
                        <motion.button
                            onClick={onClose}
                            className="
                                flex-shrink-0 w-5 h-5 rounded-md
                                flex items-center justify-center
                                text-gray-400 hover:text-gray-600 dark:hover:text-gray-300
                                hover:bg-gray-100 dark:hover:bg-gray-700/50
                                transition-colors duration-150
                            "
                            variants={closeButtonVariants}
                            whileHover={{ 
                                scale: 1.1,
                                rotate: 90,
                                transition: { type: "spring", bounce: 0.6, duration: 0.3 }
                            }}
                            whileTap={{ 
                                scale: 0.9,
                                transition: { duration: 0.1 }
                            }}
                        >
                            <svg
                                className="w-3 h-3"
                                fill="none"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </motion.button>
                    </div>

                    {/* Subtle shine effect */}
                    <motion.div
                        className="absolute inset-0 rounded-xl bg-gradient-to-r from-transparent via-white/10 to-transparent opacity-0"
                        initial={{ x: "-100%" }}
                        animate={{ 
                            x: "100%",
                            opacity: [0, 1, 0],
                            transition: {
                                delay: 0.5,
                                duration: 1.5,
                                ease: "easeInOut"
                            }
                        }}
                    />
                </motion.div>
            </motion.div>
        </AnimatePresence>
    )

    // 使用 Portal 确保在最顶层渲染
    if (typeof window !== "undefined") {
        return createPortal(notificationContent, document.body)
    }

    return null
}

export default Notification

