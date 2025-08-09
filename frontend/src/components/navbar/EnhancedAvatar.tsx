"use client"

import {useEffect, useState} from "react"
import {motion, AnimatePresence, LayoutGroup} from "framer-motion"
import {Avatar} from "@radix-ui/themes"
import {format} from "date-fns"
import Link from "next/link" // 引入 Link 组件

// 假设这些是你项目中的本地模块
// 如果路径不同，请相应修改
import emitter from "@/utils/eventBus"
import {getAuthorStatus} from "@/api/author-status"
import useIsMobile from "@/hooks/useIsMobile"

// --- 接口定义 ---
interface UserActivity {
    process?: string
    extend?: string
    media?: {
        artist: string
        thumbnail: string
        title: string
    }
    ok: number
    timestamp?: number
}

interface EnhancedAvatarProps {
    avatarSrc: string
}

// --- 共享头像组件 (集成 Link) ---
const SharedAvatar = ({
                          src,
                          isOnline,
                          showDot,
                      }: {
    src: string
    isOnline: boolean
    showDot: boolean
}) => (
    <motion.div
        layoutId="enhanced-avatar-box"
        className="relative shrink-0 w-10 h-10"
    >
        {/* 【链接修复】使用 Link 包裹 Avatar */}
        <Link href="/" aria-label="前往主页" className="block w-full h-full">
            <Avatar size="3" radius="full" src={src} fallback="A" className="w-full h-full"/>
        </Link>

        <AnimatePresence>
            {isOnline && showDot && (
                <motion.div
                    initial={{scale: 0}}
                    animate={{scale: 1}}
                    exit={{scale: 0}}
                    transition={{type: "spring", stiffness: 500, damping: 30}}
                    // 【主题修复】边框颜色现在可以适配深浅色背景
                    // 添加 pointer-events-none 防止遮挡点击
                    className="absolute bottom-[-2px] right-[-2px] w-3.5 h-3.5 bg-green-500 rounded-full border-2 border-white dark:border-black/80 pointer-events-none"
                >
                    <div className="absolute inset-0 bg-green-400 rounded-full animate-ping opacity-75"></div>
                </motion.div>
            )}
        </AnimatePresence>
    </motion.div>
)

// --- 主组件 ---
export function EnhancedAvatar({avatarSrc}: EnhancedAvatarProps) {
    // --- State 和 Hooks ---
    const [userActivity, setUserActivity] = useState<UserActivity>({ok: 0})
    const [isOnline, setIsOnline] = useState<boolean>(false)
    const [showInfo, setShowInfo] = useState<boolean>(false)
    const isMobile = useIsMobile()

    // --- 数据获取和事件处理 ---
    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const res = await getAuthorStatus()
                setUserActivity(res)
                setIsOnline(res.ok === 1)
            } catch (error) {
                console.error("Failed to fetch author status:", error)
                setIsOnline(false)
            }
        }

        fetchStatus()
        // 增加轮询频率以便更快地响应状态变化
        const timer = setInterval(fetchStatus, 1000 * 30)

        const handleStatusUpdate = (content: UserActivity) => {
            setUserActivity(content)
            setIsOnline(content.ok === 1)
        }

        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        emitter.on("authorStatus", handleStatusUpdate)

        return () => {
            clearInterval(timer)
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-expect-error
            emitter.off("authorStatus", handleStatusUpdate)
        }
    }, [])

    // 如果卡片展开时变为离线，则自动关闭卡片
    useEffect(() => {
        if (showInfo && !isOnline) {
            setShowInfo(false)
        }
    }, [isOnline, showInfo])

    // 移动端交互：点击外部关闭卡片
    useEffect(() => {
        if (isMobile && showInfo) {
            const handleClickOutside = (e: MouseEvent) => {
                if (!(e.target as HTMLElement).closest("[data-enhanced-avatar-container]")) {
                    setShowInfo(false)
                }
            }
            document.addEventListener("click", handleClickOutside)
            return () => document.removeEventListener("click", handleClickOutside)
        }
    }, [isMobile, showInfo])

    const handleMouseEnter = () => !isMobile && isOnline && setShowInfo(true)
    const handleMouseLeave = () => !isMobile && setShowInfo(false)

    const handleClick = (e: React.MouseEvent) => {
        // 在移动端点击是为了展开卡片，而不是跳转页面，所以要阻止默认行为
        if (isMobile && isOnline) {
            // 检查点击目标是否在头像链接内部，如果是，则阻止默认跳转
            const target = e.target as HTMLElement
            if (target.closest('a[href="/"]')) {
                e.preventDefault()
            }
            setShowInfo(!showInfo)
        }
        // 在 PC 端，点击事件会自然地由 Link 组件处理，无需干预
    }

    // --- 动画参数 ---
    const transition = {type: "spring", stiffness: 500, damping: 30}

    // --- 状态逻辑 ---
    const isExpandedCapsule = isOnline && !showInfo

    return (
        <LayoutGroup>
            {/* 1. 布局占位符 */}
            <motion.div
                onMouseEnter={handleMouseEnter}
                onMouseLeave={handleMouseLeave}
                onClick={handleClick}
                data-enhanced-avatar-container
                className="relative h-12 flex items-center cursor-pointer"
                animate={{width: isExpandedCapsule ? 170 : 48}}
                transition={transition}
            >
                {/* 2. 视觉层 */}
                <AnimatePresence initial={false} mode="wait">
                    {showInfo && isOnline ? (
                        // --- 状态三：详情卡片（在线）---
                        <motion.div
                            key="card"
                            className="absolute top-1/2 -translate-y-1/2 left-0 z-50 w-[calc(100vw-32px)] max-w-sm bg-white/90 dark:bg-gray-950/95 backdrop-blur-xl border border-gray-200/80 dark:border-white/10 shadow-2xl rounded-2xl"
                            initial={{opacity: 0, scale: 0.85}}
                            animate={{opacity: 1, scale: 1, transition}}
                            exit={{opacity: 0, scale: 0.85, transition: {duration: 0.2}}}
                        >
                            <div className="p-4 space-y-4">
                                <div className="flex items-center gap-3">
                                    <SharedAvatar src={avatarSrc} isOnline={isOnline} showDot={true}/>
                                    <div className="flex-1">
                                        <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100">
                                            当前在线
                                        </h3>
                                        <p className="text-xs text-gray-500 dark:text-gray-400"> 康康他在干什么 👀 </p>
                                    </div>
                                    {isMobile && (
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                setShowInfo(false)
                                            }}
                                            className="w-7 h-7 rounded-full bg-gray-500/10 dark:bg-white/10 flex items-center justify-center text-gray-600 dark:text-gray-400 hover:bg-gray-500/20 dark:hover:bg-white/20 transition-colors"
                                        >
                                            ✕
                                        </button>
                                    )}
                                </div>

                                <div>
                                    <p className="text-sm text-gray-700 dark:text-gray-300">
                                        <span
                                            className="font-medium text-gray-900 dark:text-white">grtsinry43</span>{" "}
                                        正在使用 {" "}
                                        <span className="font-semibold text-blue-600 dark:text-blue-400">
                      {userActivity.process || "未知应用"}
                    </span>
                                    </p>
                                    {userActivity.extend && (
                                        <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                                            {userActivity.extend}
                                        </p>
                                    )}
                                </div>

                                {userActivity.media?.title && (
                                    <div
                                        className="flex items-center gap-3 p-3 bg-gray-100/50 dark:bg-white/5 rounded-xl">
                                        <img
                                            src={userActivity.media.thumbnail || "/placeholder.svg"}
                                            alt={userActivity.media.title}
                                            className="w-12 h-12 rounded-lg object-cover shadow-md"
                                        />
                                        <div className="flex-1 min-w-0">
                                            <p className="text-sm font-semibold text-gray-800 dark:text-gray-100 truncate">
                                                {userActivity.media.title}
                                            </p>
                                            <p className="text-xs text-gray-600 dark:text-gray-400 truncate">
                                                {userActivity.media.artist}
                                            </p>
                                        </div>
                                    </div>
                                )}

                                {userActivity.timestamp && (
                                    <div
                                        className="text-xs text-gray-400 dark:text-gray-500 pt-3 border-t border-gray-200/60 dark:border-white/10">
                                        最后活跃于 {" "}
                                        {format(new Date(userActivity.timestamp * 1000), "yyyy-MM-dd HH:mm:ss")}
                                    </div>
                                )}
                            </div>
                        </motion.div>
                    ) : (
                        // --- 状态一 & 二：圆形头像 或 胶囊 ---
                        <motion.div
                            key="capsule"
                            className="absolute inset-0 flex items-center bg-white/90 dark:bg-gray-950/20 backdrop-blur-md border border-gray-200/80 dark:border-white/5 shadow-lg rounded-full p-1"
                            initial={{opacity: 0}}
                            animate={{opacity: 1}}
                            exit={{opacity: 0, transition: {duration: 0.1}}}
                        >
                            <SharedAvatar src={avatarSrc} isOnline={isOnline} showDot={!isExpandedCapsule}/>
                            {isExpandedCapsule && (
                                <div className="pl-2 pr-3 min-w-0 flex-1">
                                    <div className="flex items-center gap-1.5 mb-0.5">
                                        <span className="text-green-500 dark:text-green-400 text-xs font-bold">●</span>
                                        <span className="text-xs font-semibold text-gray-800 dark:text-gray-200">
                      在线
                    </span>
                                    </div>
                                    <p className="text-xs text-gray-600 dark:text-gray-400 truncate">
                                        {userActivity.process || "摸鱼中..."}
                                    </p>
                                </div>
                            )}
                        </motion.div>
                    )}
                </AnimatePresence>
            </motion.div>
        </LayoutGroup>
    )
}