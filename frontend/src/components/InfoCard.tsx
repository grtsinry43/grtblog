"use client"

import {useEffect, useState} from "react"
import {useWebsiteInfo} from "@/app/website-info-provider"
import useIsMobile from "@/hooks/useIsMobile"
import {motion} from "framer-motion"
import {Clock, Heart} from "lucide-react"

const InfoCard = () => {
    const websiteInfo = useWebsiteInfo()
    const createTime = new Date(websiteInfo.WEBSITE_CREATE_TIME)
    const [timeUnits, setTimeUnits] = useState({days: 0, hours: 0, minutes: 0, seconds: 0})
    const isMobile = useIsMobile()

    useEffect(() => {
        const interval = setInterval(() => {
            const now = new Date()
            const diff = now.getTime() - createTime.getTime()
            const days = Math.floor(diff / (1000 * 60 * 60 * 24))
            const hours = Math.floor((diff / (1000 * 60 * 60)) % 24)
            const minutes = Math.floor((diff / (1000 * 60)) % 60)
            const seconds = Math.floor((diff / 1000) % 60)
            setTimeUnits({days, hours, minutes, seconds})
        }, 1000)

        return () => clearInterval(interval)
    }, [createTime])

    if (isMobile) {
        return null
    }

    return (
        <motion.div
            initial={{opacity: 0, y: 20}}
            animate={{opacity: 1, y: 0}}
            transition={{duration: 0.5}}
            className="sticky mt-80 ml-4 mb-4"
        >
            <div
                className="w-64 bg-white dark:bg-background rounded-lg border border-gray-100 dark:border-gray-700 overflow-hidden shadow-sm">
                <div className="p-4">
                    {/* 标题 */}
                    <div className="flex items-center justify-center mb-3">
                        <Clock className="w-4 h-4 text-primary mr-1.5"/>
                        <h3 className="text-sm font-medium text-gray-700 dark:text-gray-200">网站运行时间</h3>
                    </div>

                    {/* 时间数字 */}
                    <div className="grid grid-cols-4 gap-1 mb-3">
                        <div className="flex flex-col items-center">
                            <div className="text-xl font-bold text-primary">{timeUnits.days}</div>
                            <div className="text-xs text-gray-500 dark:text-gray-400">天</div>
                        </div>
                        <div className="flex flex-col items-center">
                            <div className="text-xl font-bold text-primary">{timeUnits.hours}</div>
                            <div className="text-xs text-gray-500 dark:text-gray-400">时</div>
                        </div>
                        <div className="flex flex-col items-center">
                            <div className="text-xl font-bold text-primary">{timeUnits.minutes}</div>
                            <div className="text-xs text-gray-500 dark:text-gray-400">分</div>
                        </div>
                        <div className="flex flex-col items-center">
                            <div className="text-xl font-bold text-primary">{timeUnits.seconds}</div>
                            <div className="text-xs text-gray-500 dark:text-gray-400">秒</div>
                        </div>
                    </div>

                    {/* 分隔线 */}
                    <div className="h-px bg-gray-100 dark:bg-gray-700 my-3"></div>

                    {/* 底部文字 */}
                    <div className="text-center">
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">在风雨飘摇之中</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400 flex items-center justify-center mb-1">
                            感谢陪伴与支持
                            <motion.div
                                animate={{scale: [1, 1.2, 1]}}
                                transition={{duration: 1.5, repeat: Number.POSITIVE_INFINITY}}
                                className="ml-1"
                            >
                                <Heart className="w-3 h-3 text-red-500 fill-red-500"/>
                            </motion.div>
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">愿我们不负热爱，继续前行</p>
                    </div>
                </div>
            </div>
        </motion.div>
    )
}

export default InfoCard

