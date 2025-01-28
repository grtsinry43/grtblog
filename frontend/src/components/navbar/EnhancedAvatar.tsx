"use client"

import React, {useEffect} from "react"
import {motion, useAnimation} from "framer-motion"
import {Avatar} from "@radix-ui/themes"
import {Popover, PopoverContent, PopoverTrigger} from "@radix-ui/react-popover"
import {format} from "date-fns"
import styles from "@/styles/NavBar.module.scss";
import emitter from "@/utils/eventBus";
import {getAuthorStatus} from "@/api/author-status";

interface UserActivity {
    process?: string
    extend?: string,
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
    const [userActivity, setUserActivity] = React.useState<UserActivity>({
        ok: 0,
    });
    const [isOnline, setIsOnline] = React.useState<boolean>(false);

    useEffect(() => {
        // 挂载时获取一次用户状态
        getAuthorStatus().then((res) => {
            setUserActivity(res);
            if (res.ok === 1) {
                setIsOnline(true);
            }
        })
    }, []);

    useEffect(() => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        emitter.on("authorStatus", (content: UserActivity) => {
            setUserActivity(content);
            if (content.ok === 1) {
                setIsOnline(true);
            } else {
                setIsOnline(false);
            }
        });

        // 每 5 分钟获取一次用户状态（强制刷新）
        const timer = setInterval(() => {
            getAuthorStatus().then((res) => {
                setUserActivity(res);
                if (res.ok === 1) {
                    setIsOnline(true);
                } else {
                    setIsOnline(false);
                }
            })
        }, 1000 * 60 * 5);

        return () => {
            clearInterval(timer);
        }

    }, []);

    const handleMouseEnter = () => {
        if (isOnline) {
            controls.start({scale: 1.1, transition: {type: "spring", stiffness: 300, damping: 10}})
        }
    }

    const handleMouseLeave = () => {
        if (isOnline) {
            controls.start({scale: 1, transition: {type: "spring", stiffness: 300, damping: 10}})
        }
    }

    return (
        <Popover>
            <PopoverTrigger asChild>
                <motion.div
                    animate={controls}
                >
                    <div
                        className="relative inline-block"
                        onMouseEnter={handleMouseEnter}
                        onMouseLeave={handleMouseLeave}>
                        <Avatar size="3" radius="large" src={avatarSrc}
                                fallback="A"
                                className={styles.avatar}/>
                        {isOnline && (
                            <motion.div
                                initial={{scale: 0}}
                                animate={{scale: 1}}
                                transition={{type: "spring", stiffness: 300, damping: 20}}
                            >
                                <div
                                    className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-white"></div>
                            </motion.div>
                        )}
                    </div>
                </motion.div>
            </PopoverTrigger>
            <PopoverContent
                className="w-80 p-4 rounded-lg shadow-lg bg-white dark:bg-black border border-gray-200 dark:border-gray-900"
                sideOffset={5}>
                <motion.div
                    initial={{opacity: 0, y: 10}}
                    animate={{opacity: 1, y: 0}}
                    transition={{type: "spring", stiffness: 300, damping: 20}}
                >
                    <h3 className="text-lg font mb-2">{
                        isOnline ? "当前在线，康康他在干什么👀" : "没有在线捏"
                    }</h3>
                    {
                        isOnline && (
                            <p className="text-sm text-gray-700 dark:text-gray-300 mb-2">
                                grtsinry43 正在使用 <b>{userActivity.process}</b> {userActivity.extend} </p>
                        )
                    }
                    {
                        userActivity.media?.title && (
                            <div className="flex items-center mb-2">
                                <img
                                    src={userActivity.media.thumbnail || "/placeholder.svg"}
                                    alt={userActivity.media.title}
                                    className="w-12 h-12 rounded mr-3"
                                />
                                <div>
                                    <p className="text-sm font-medium">{userActivity.media.title}</p>
                                    <p className="text-xs text-gray-500">{userActivity.media.artist}</p>
                                </div>
                            </div>
                        )
                    }
                    {
                        userActivity.timestamp && (
                            <p className="text-xs text-gray-400">
                                最后活跃于 {format(new Date(userActivity?.timestamp * 1000), "yyyy-MM-dd HH:mm:ss")}
                            </p>
                        )
                    }
                </motion.div>
            </PopoverContent>
        </Popover>
    )
}

