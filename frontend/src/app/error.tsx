"use client"

import {useState, useEffect} from "react"
import {Button} from "@/components/ui/button"
import {Card} from "@/components/ui/card"
import {RefreshCw, XCircle} from "lucide-react"
import {cn} from "@/lib/utils"
import {ScrollArea} from "@radix-ui/themes";

export default function Error({
                                  error,
                                  reset,
                              }: {
    error: Error & { digest?: string }
    reset: () => void
}) {
    const [isRefreshing, setIsRefreshing] = useState(false)
    const [isExpanded, setIsExpanded] = useState(false)
    const [isVisible, setIsVisible] = useState(false)

    useEffect(() => {
        console.error(error)

        const timer = setTimeout(() => setIsVisible(true), 100)
        return () => clearTimeout(timer)
    }, [error])

    const handleRefresh = () => {
        setIsRefreshing(true)
        setTimeout(() => {
            reset();
            setIsRefreshing(false)
        }, 800)
    }

    const toggleExpand = () => {
        setIsExpanded(!isExpanded)
    }

    const errorMessage = error.message || "未知错误"
    const errorDigest = error.digest ? `(${error.digest})` : ""

    return (
        <div className="flex min-h-[70vh] items-center justify-center p-4">
            <Card
                className={cn(
                    "overflow-hidden max-w-md border-none transition-all duration-500 ease-in-out",
                    "bg-white dark:bg-gray-900",
                    "shadow-[0_10px_40px_-15px_rgba(0,0,0,0.2)] dark:shadow-[0_10px_40px_-15px_rgba(255,255,255,0.1)]",
                    isVisible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8",
                    isExpanded ? "max-w-3xl" : "max-w-md",
                )}
            >
                <div className="bg-gradient-to-r from-red-500 via-pink-500 to-purple-500 h-1"/>

                <div className="p-5">
                    <div className="flex items-start">
                        <div
                            className={cn(
                                "flex-shrink-0 mr-4 transition-all duration-500",
                                "bg-red-50 dark:bg-red-900/20 rounded-full p-2",
                                isExpanded ? "scale-90" : "scale-100",
                            )}
                        >
                            <XCircle className="h-6 w-6 text-red-500 dark:text-red-400"/>
                        </div>

                        <div className="flex-1">
                            <h2 className={cn("font-bold transition-all duration-300", isExpanded ? "text-xl mb-3" : "text-lg mb-2")}>
                                Oops!
                            </h2>

                            <div
                                className={cn("space-y-2 transition-all duration-500", isExpanded ? "opacity-100" : "opacity-90")}>
                                <p className="text-sm text-gray-700 dark:text-gray-300">看起来客户端组件渲染出现了一些问题
                                    (｡•́︿•̀｡)</p>

                                <div
                                    className={cn(
                                        "overflow-hidden transition-all duration-500 ease-in-out",
                                        isExpanded ? "max-h-80 opacity-100" : "max-h-0 opacity-0",
                                    )}
                                >
                                    <ScrollArea scrollbars={"horizontal"} style={{maxWidth: "680px"}}>
                                        <div
                                            className="w-full mt-2 rounded bg-gray-100 dark:bg-gray-800 p-2 text-xs font-mono max-h-40 overflow-hidden">

                                            <div className="overflow-x-auto" style={{maxHeight: "160px"}}>
                                                <p className="text-red-500 dark:text-red-400 font-medium whitespace-nowrap">{errorMessage}</p>
                                                {errorDigest && (
                                                    <p className="text-gray-500 dark:text-gray-400 mt-1 whitespace-nowrap">错误ID: {errorDigest}</p>
                                                )}
                                                {error.stack && (
                                                    <pre
                                                        className="text-gray-600 dark:text-gray-400 mt-2 text-[10px] leading-tight whitespace-pre">
                          {error.stack.split("\n").slice(0, 5).join("\n")}
                        </pre>
                                                )}
                                            </div>
                                        </div>
                                    </ScrollArea>
                                    <p className="text-xs text-gray-600 dark:text-gray-400 mt-2">
                                        这可能是由于网络连接问题或者代码执行错误导致的。我们建议您尝试刷新页面，<br/>
                                        如果问题依旧存在，请检查您的网络连接或联系网站管理员获取帮助。<br/>
                                        如果方便的话，您可以复制上面的错误信息并发送给站长，以便更好地帮助您解决问题。
                                    </p>
                                </div>

                                <p className="text-xs text-gray-500 dark:text-gray-400">你可以尝试刷新页面，如果问题依旧，请联系站长</p>

                                <p className="text-xs text-gray-400 dark:text-gray-500 italic">很是抱歉.｡･ﾟﾟ･(＞_＜)･ﾟﾟ･｡.</p>
                            </div>
                        </div>
                    </div>

                    <div className="mt-4 flex items-center justify-between">
                        <Button variant="ghost" size="sm" onClick={toggleExpand}
                                className="text-xs h-8 px-2 text-gray-500">
                            {isExpanded ? "收起详情" : "查看详情"}
                        </Button>

                        <Button
                            onClick={handleRefresh}
                            variant="default"
                            className={cn(
                                "h-8 px-3 text-xs bg-gradient-to-r from-red-500 to-pink-500 hover:from-red-600 hover:to-pink-600 border-none",
                                "transition-all duration-300 ease-in-out",
                                isRefreshing && "animate-pulse",
                            )}
                            size="sm"
                            disabled={isRefreshing}
                        >
                            <RefreshCw className={cn("mr-1 h-3 w-3", isRefreshing && "animate-spin")}/>
                            刷新页面
                        </Button>
                    </div>
                </div>
            </Card>
        </div>
    )
}

