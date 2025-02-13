"use client"

import type React from "react"
import {useEffect} from "react"
import {motion, useAnimation, type Variants, AnimatePresence} from "framer-motion"
import {CheckCircledIcon, ExclamationTriangleIcon, CrossCircledIcon, InfoCircledIcon} from "@radix-ui/react-icons"

export type NotificationType = "success" | "warning" | "error" | "info"

interface NotificationProps {
    message: string
    type: NotificationType
    onClose: () => void
}

const notificationVariants: Variants = {
    initial: {
        opacity: 0,
        y: -20,
        scale: 0.8,
    },
    animate: {
        opacity: 1,
        y: 0,
        scale: 1.05,
        transition: {
            duration: 0.2,
            ease: [0, 0.71, 0.2, 1.01],
        },
    },
    exit: {
        opacity: 0,
        y: -20,
        scale: 0.75,
        transition: {
            duration: 0.07,
            ease: "easeOut",
        },
    },
}

const typeConfig = {
    success: {icon: CheckCircledIcon, bgColor: "bg-green-500"},
    warning: {icon: ExclamationTriangleIcon, bgColor: "bg-yellow-500"},
    error: {icon: CrossCircledIcon, bgColor: "bg-red-500"},
    info: {icon: InfoCircledIcon, bgColor: "bg-blue-500"},
}

const Notification: React.FC<NotificationProps> = ({message, type, onClose}) => {
    const controls = useAnimation()
    const {icon: Icon, bgColor} = typeConfig[type]

    useEffect(() => {
        controls.start("animate")
        const timer = setTimeout(() => {
            controls.start("exit").then(onClose)
        }, 5000)

        return () => clearTimeout(timer)
    }, [controls, onClose])

    return (
        <AnimatePresence>
            <motion.div
                // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                // @ts-expect-error
                className="fixed top-16 left-1/2 z-50"
                initial="initial"
                animate={controls}
                exit="exit"
                variants={notificationVariants}
            >
                <motion.div
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-expect-error
                    className="bg-white dark:bg-black dark:bg-opacity-20 bg-opacity-20 backdrop-filter backdrop-blur-lg rounded-full shadow-lg flex items-center overflow-hidden"
                    initial={{width: 40, x: "-50%"}}
                    animate={{
                        width: "auto",
                        x: "-50%",
                        transition: {delay: 0.1, duration: 0.3, ease: "easeOut"},
                    }}
                >
                    <div className={`w-10 h-10 ${bgColor} rounded-full flex items-center justify-center flex-shrink-0`}>
                        <Icon className="text-white" height={20} width={20}/>
                    </div>
                    <motion.div
                        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                        // @ts-expect-error
                        className="px-4 py-2 overflow-hidden"
                        initial={{width: 0, opacity: 0}}
                        animate={{
                            width: "auto",
                            opacity: 1,
                            transition: {delay: 0.1, duration: 0.3, ease: "easeOut"},
                        }}
                    >
                        <motion.p
                            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                            // @ts-expect-error
                            className="text-sm text-gray-800 dark:text-gray-200 whitespace-nowrap"
                            initial={{x: 20, opacity: 0}}
                            animate={{
                                x: 0,
                                opacity: 1,
                                transition: {delay: 0.3, duration: 0.3, ease: "easeOut"},
                            }}
                        >
                            {message}
                        </motion.p>
                    </motion.div>
                </motion.div>
            </motion.div>
        </AnimatePresence>
    )
}

export default Notification

