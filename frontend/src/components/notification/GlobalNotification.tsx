"use client"

import {useEffect, useState} from "react"
import {Checkbox} from "@/components/ui/checkbox"
import {Button} from "@/components/ui/button"
import {X, Bell} from "lucide-react"
import {getGlobalNotification, type Notification} from "@/api/nofitication"
import {cn} from "@/lib/utils"

const GlobalNotification = () => {
    const [notification, setNotification] = useState<Notification>({
        id: "",
        content: "",
        publishAt: "",
        expireAt: "",
        allowClose: false,
    })
    const [show, setShow] = useState(false)
    const [elapsedTime, setElapsedTime] = useState("")
    const [isMobile, setIsMobile] = useState(false)

    // Check if device is mobile
    useEffect(() => {
        const checkMobile = () => {
            setIsMobile(window.innerWidth < 768)
        }

        checkMobile()
        window.addEventListener("resize", checkMobile)

        return () => window.removeEventListener("resize", checkMobile)
    }, [])

    // Fetch notification
    useEffect(() => {
        getGlobalNotification().then((res) => {
            if (res) {
                setNotification(res)
                // Read from local storage, format is {notificationId: true}
                const closeNotifications = JSON.parse(localStorage.getItem("closeNotifications") ?? "{}")
                if (closeNotifications[res.id]) {
                    setShow(false)
                } else {
                    setTimeout(() => setShow(true), 1000)
                }
            }
        })
    }, [])

    // Calculate elapsed time
    useEffect(() => {
        const calculateElapsedTime = () => {
            if (!notification.publishAt) return

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

    const handleDontShowAgain = (checked: boolean) => {
        // Save to local storage, format is {notificationId: true}
        const closeNotifications = JSON.parse(localStorage.getItem("closeNotifications") ?? "{}")
        closeNotifications[notification.id] = checked
        localStorage.setItem("closeNotifications", JSON.stringify(closeNotifications))
    }

    if (!notification.id) return null

    return (
        <div
            className={cn(
                "fixed z-[9999] transition-all duration-500 ease-in-out shadow-lg",
                isMobile ? "top-0 left-0 right-0 m-4 w-auto" : "top-5 right-5 w-[350px]",
                show ? (isMobile ? "translate-y-0" : "translate-x-0") : isMobile ? "translate-y-[-200%]" : "translate-x-[120%]",
            )}
        >
            <div className="relative overflow-hidden rounded-lg border bg-card/80 text-card-foreground backdrop-blur-2xl">
                {/* Animated accent border */}
                <div
                    className="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-primary to-primary/50 animate-pulse"/>

                <div className="p-5" key={notification.id ?? "notification-null"}>
                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2 font-semibold">
                            <Bell className="h-4 w-4 text-primary animate-[pulse_3s_ease-in-out_infinite]"/>
                            <span>全站通知</span>
                        </div>
                        <div className="text-xs text-muted-foreground">{elapsedTime}</div>
                    </div>

                    <div className="mb-4 text-sm">{notification.content}</div>

                    <div className="flex items-center justify-between mt-2">
                        {notification.allowClose && (
                            <div className="flex items-center space-x-2">
                                <Checkbox id="dont-show-again" onCheckedChange={handleDontShowAgain}/>
                                <label htmlFor="dont-show-again"
                                       className="text-xs text-muted-foreground cursor-pointer">
                                    不再提示
                                </label>
                            </div>
                        )}

                        <div className="flex gap-2">
                            <Button variant="outline" size="sm" className="h-8 px-3 text-xs"
                                    onClick={() => setShow(false)}>
                                <X className="h-3.5 w-3.5 mr-1"/>
                                关闭
                            </Button>

                            {/*<Button variant="default" size="sm" className="h-8 px-3 text-xs group">*/}
                            {/*    <ExternalLink*/}
                            {/*        className="h-3.5 w-3.5 mr-1 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5"/>*/}
                            {/*    查看详情*/}
                            {/*</Button>*/}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GlobalNotification

