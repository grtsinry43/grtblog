import {CalendarIcon, AlertCircleIcon, ClockIcon} from "lucide-react"

interface ArticleOutdatedProps {
    publishDate?: string
    lastUpdated?: string
    outdatedDays?: number
    customMessage?: string
}

export default function ArticleOutdatedNotice({
                                                  publishDate = "2025-01-01",
                                                  lastUpdated,
                                                  outdatedDays = 365, // 默认一年后显示过时提醒
                                                  customMessage,
                                              }: ArticleOutdatedProps) {
    // 计算文章是否过时
    const lastDate = lastUpdated ? new Date(lastUpdated) : new Date(publishDate)
    const daysSinceUpdate = Math.floor((new Date().getTime() - lastDate.getTime()) / (1000 * 3600 * 24))
    const isOutdated = daysSinceUpdate > outdatedDays

    // 如果文章未过时，不显示组件
    if (!isOutdated) return null

    // 格式化日期
    const formatDate = (dateString: string) => {
        const date = new Date(dateString)
        return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`
    }

    const publishDateFormatted = formatDate(publishDate)
    const lastUpdatedFormatted = lastUpdated ? formatDate(lastUpdated) : null

    return (
        <div className="my-6 text-xs">
            <div
                className="border-l-2 border-amber-400 dark:border-amber-500 bg-amber-50/80 dark:bg-amber-900/20 rounded-r px-3 py-2.5">
                <div className="flex items-start gap-2">
                    <AlertCircleIcon className="h-3.5 w-3.5 text-amber-500 dark:text-amber-400 mt-0.5 flex-shrink-0"/>
                    <div className="text-gray-700 dark:text-gray-200">
                        <p className="leading-tight font-medium">{customMessage || "本文内容可能已经过时，请注意信息的时效性。"}</p>

                        <div
                            className="flex flex-wrap gap-x-4 gap-y-1 mt-2 text-[10px] text-gray-500 dark:text-gray-400">
                            <div className="flex items-center gap-1">
                                <CalendarIcon className="h-3 w-3"/>
                                <span>发布于 {publishDateFormatted}</span>
                            </div>

                            {lastUpdated && (
                                <div className="flex items-center gap-1">
                                    <ClockIcon className="h-3 w-3"/>
                                    <span>最后更新 {lastUpdatedFormatted}</span>
                                </div>
                            )}

                            <div className="flex items-center gap-1">
                                <span>已有 {daysSinceUpdate} 天未更新</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

