"use client"

import {useEffect, useState} from "react"
import {motion, useAnimation} from "framer-motion"
import {Avatar} from "@radix-ui/themes"
import Link from "next/link"
import {format} from "date-fns"
import styles from "@/styles/NavBar.module.scss"
import emitter from "@/utils/eventBus"
import {getAuthorStatus} from "@/api/author-status"

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

export function EnhancedAvatar({avatarSrc}: EnhancedAvatarProps) {
    const controls = useAnimation()
    const [userActivity, setUserActivity] = useState<UserActivity>({ok: 0})
    const [isOnline, setIsOnline] = useState<boolean>(false)
    const [showInfo, setShowInfo] = useState<boolean>(false)

    useEffect(() => {
        const fetchStatus = async () => {
            const res = await getAuthorStatus()
            setUserActivity(res)
            setIsOnline(res.ok === 1)
        }

        fetchStatus()

        const timer = setInterval(fetchStatus, 1000 * 60 * 5)

        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        emitter.on("authorStatus", (content: UserActivity) => {
            setUserActivity(content)
            setIsOnline(content.ok === 1)
        })

        return () => {
            clearInterval(timer)
            emitter.off("authorStatus")
        }
    }, [])

    const handleMouseEnter = () => {
        controls.start({
            scale: 1.1,
            transition: {type: "spring", stiffness: 400, damping: 10},
        })
        setShowInfo(true)
    }

    const handleMouseLeave = () => {
        controls.start({
            scale: 1,
            transition: {type: "spring", stiffness: 400, damping: 15},
        })
        setShowInfo(false)
    }

    return (
        <div className="relative inline-block">
            {isOnline ? (
                <motion.div animate={controls} whileHover={{rotate: [0, -10, 10, -10, 0]}} transition={{duration: 0.5}}>
                    <Link href="/">
                        <div className="relative" onMouseEnter={handleMouseEnter} onMouseLeave={handleMouseLeave}>
                            <Avatar size="3" radius="large" src={avatarSrc} fallback="A" className={styles.avatar}/>
                            <motion.div
                                initial={{scale: 0}}
                                animate={{scale: 1}}
                                transition={{type: "spring", stiffness: 400, damping: 20}}
                            >
                                <div
                                    className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-white"></div>
                            </motion.div>
                        </div>
                    </Link>
                </motion.div>
            ) : (
                <Link href="/" onClick={() => {
                    emitter.emit("hideTitle");
                }}>
                    <div className="relative" onMouseEnter={handleMouseEnter} onMouseLeave={handleMouseLeave}>
                        <Avatar size="3" radius="large" src={avatarSrc} fallback="A" className={styles.avatar}/>
                    </div>
                </Link>
            )}
            {showInfo && isOnline && (
                <motion.div
                    initial={{opacity: 0, y: 10}}
                    animate={{opacity: 1, y: 0}}
                    exit={{opacity: 0, y: 10}}
                    transition={{type: "spring", stiffness: 400, damping: 20}}


                    className="absolute top-full left-1/2 transform -translate-x-1/2 mt-2 w-64 p-4 rounded-lg shadow-lg bg-white dark:bg-gray-900 backdrop-blur border border-gray-200 dark:border-gray-700 z-10"
                >
                    <h3 className="text-sm font-medium mb-2"> 当前在线，康康他在干什么👀</h3>
                    <p className="text-xs text-gray-700 dark:text-gray-300 mb-2">
                        grtsinry43 正在使用 <b>{userActivity.process}</b> {userActivity.extend}
                    </p>
                    {userActivity.media?.title && (
                        <div className="flex items-center mb-2">
                            <img
                                src={userActivity.media.thumbnail || "/placeholder.svg"}
                                alt={userActivity.media.title}
                                className="w-10 h-10 rounded mr-2"
                            />
                            <div>
                                <p className="text-xs font-medium">{userActivity.media.title}</p>
                                <p className="text-xs text-gray-500">{userActivity.media.artist}</p>
                            </div>
                        </div>
                    )}
                    {userActivity.timestamp && (
                        <p className="text-xs text-gray-400">
                            最后活跃于 {format(new Date(userActivity.timestamp * 1000), "yyyy-MM-dd HH:mm:ss")}
                        </p>
                    )}
                </motion.div>
            )}
        </div>
    )
}

