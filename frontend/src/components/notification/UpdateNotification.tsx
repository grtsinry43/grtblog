"use client"

import {useEffect, useState, useRef} from "react"
import {Button} from "@/components/ui/button"
import {X, RefreshCw, ExternalLink} from "lucide-react"
import {cn} from "@/lib/utils"
import channel from "@/utils/channel"

interface UpdateNotification {
    content: string
    publishAt: string
    link?: string // Optional link to navigate to
}

const UpdateNotification = () => {
    const [notification, setNotification] = useState<UpdateNotification>({
        content: "",
        publishAt: "",
    })
    const [show, setShow] = useState(false)
    const [elapsedTime, setElapsedTime] = useState("")
    const [isMobile, setIsMobile] = useState(false)
    const [progress, setProgress] = useState(100)
    const progressTimerRef = useRef<NodeJS.Timeout | null>(null)
    const hideTimerRef = useRef<NodeJS.Timeout | null>(null)
    const isHovering = useRef(false)

    // Check if device is mobile
    useEffect(() => {
        const checkMobile = () => {
            setIsMobile(window.innerWidth < 768)
        }

        checkMobile()
        window.addEventListener("resize", checkMobile)

        return () => window.removeEventListener("resize", checkMobile)
    }, [])

    // Start countdown timers
    const startTimers = () => {
        // Clear any existing timers
        if (progressTimerRef.current) clearInterval(progressTimerRef.current)
        if (hideTimerRef.current) clearTimeout(hideTimerRef.current)

        // Reset progress
        setProgress(100)

        // Start progress countdown
        progressTimerRef.current = setInterval(() => {
            if (isHovering.current) return // Pause countdown while hovering

            setProgress((prev) => {
                if (prev <= 0) {
                    if (progressTimerRef.current) clearInterval(progressTimerRef.current)
                    return 0
                }
                return prev - 1
            })
        }, 100)

        // Auto hide after 10 seconds
        hideTimerRef.current = setTimeout(() => {
            if (!isHovering.current) setShow(false)
        }, 10000)
    }

    // Listen for update notifications
    useEffect(() => {
        channel.port1.onmessage = (event) => {
            console.log(event)
            const res = event.data
            if (res) {
                setNotification(res)
                setShow(true)
                startTimers()
            }
        }

        return () => {
            if (progressTimerRef.current) clearInterval(progressTimerRef.current)
            if (hideTimerRef.current) clearTimeout(hideTimerRef.current)
        }
    }, [])

    // Calculate elapsed time
    useEffect(() => {
        if (!notification.publishAt) return

        const calculateElapsedTime = () => {
            const timeParsed = new Date(notification.publishAt)
            const now = new Date()
            const timeDifference = Math.abs(now.getTime() - timeParsed.getTime())

            const days = Math.floor(timeDifference / (1000 * 60 * 60 * 24))
            const hours = Math.floor((timeDifference % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
            const minutes = Math.floor((timeDifference % (1000 * 60 * 60)) / (1000 * 60))
            const seconds = Math.floor((timeDifference % (1000 * 60)) / 1000)

            let elapsedTime = ""
            if (days > 0) elapsedTime += days + " 天 "
            if (hours > 0) elapsedTime += hours + " 小时 "
            if (minutes > 0) elapsedTime += minutes + " 分钟 "
            if (seconds > 0) elapsedTime += seconds + " 秒"
            elapsedTime += "前"

            setElapsedTime(elapsedTime)
        }

        calculateElapsedTime()
        const interval = setInterval(calculateElapsedTime, 1000)

        return () => clearInterval(interval)
    }, [notification.publishAt])

    // Handle mouse hover
    const handleMouseEnter = () => {
        isHovering.current = true

        // Reset progress to 100% when mouse enters
        setProgress(100)

        // Clear any existing timers
        if (hideTimerRef.current) clearTimeout(hideTimerRef.current)
        if (progressTimerRef.current) clearInterval(progressTimerRef.current)
    }

    const handleMouseLeave = () => {
        isHovering.current = false

        // Restart the timers when mouse leaves
        startTimers()
    }

    // Handle go check button click
    const handleGoCheck = () => {
        if (notification.link) {
            window.open(notification.link, "_blank")
        }
        // Close notification after clicking
        setShow(false)
    }

    if (!notification.content) return null

    return (
        <div
            className={cn(
                "fixed z-[9999] transition-all duration-500 ease-in-out",
                isMobile ? "bottom-0 left-0 right-0 m-4 w-auto" : "top-5 right-5 w-[320px]",
                show
                    ? isMobile
                        ? "translate-y-0 opacity-100"
                        : "translate-x-0 opacity-100"
                    : isMobile
                        ? "translate-y-[150%] opacity-0"
                        : "translate-x-[120%] opacity-0",
            )}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            <div
                className="relative overflow-hidden rounded-lg border bg-card/80 text-card-foreground shadow-md backdrop-blur-md">
                {/* Enhanced progress bar */}
                <div
                    className="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-transparent via-muted/30 to-transparent overflow-hidden">
                    <div
                        className="absolute inset-y-0 left-0 bg-primary opacity-30 h-full rounded-r-full"
                        style={{
                            width: `${progress}%`,
                            transition: "width 100ms linear",
                            boxShadow: "0 0 8px rgba(var(--primary), 0.5)",
                        }}
                    >
                        {/* Animated particles */}
                        <div className="absolute inset-0 overflow-hidden">
                            <div className="absolute top-0 right-1/4 w-1 h-full bg-white/20 animate-pulse-fast"></div>
                            <div className="absolute top-0 right-2/3 w-0.5 h-full bg-white/30 animate-pulse-slow"></div>
                            <div
                                className="absolute top-0 right-1/2 w-0.5 h-full bg-white/20 animate-pulse-medium"></div>
                        </div>
                    </div>
                </div>

                {/* Progress indicator */}
                {/*<div className="absolute top-1 right-2 text-[10px] text-muted-foreground/70 font-mono">*/}
                {/*    {Math.ceil(progress / 10)}s*/}
                {/*</div>*/}

                <div className="p-4 pt-5">
                    <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-1.5 font-medium text-sm">
                            <RefreshCw className="h-3.5 w-3.5 text-primary animate-spin-slow"/>
                            <span>更新通知</span>
                        </div>
                        <div className="text-xs text-muted-foreground">{elapsedTime}</div>
                    </div>

                    <div className="text-sm mb-3">{notification.content}</div>

                    <div className="flex justify-end gap-2">
                        <Button variant="ghost" size="sm" className="h-7 px-2 text-xs" onClick={() => setShow(false)}>
                            <X className="h-3 w-3 mr-1"/>
                            关闭
                        </Button>

                        <Button variant="default" size="sm" className="h-7 px-2 text-xs group" onClick={handleGoCheck}>
                            <ExternalLink
                                className="h-3 w-3 mr-1 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5"/>
                            去围观
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default UpdateNotification

