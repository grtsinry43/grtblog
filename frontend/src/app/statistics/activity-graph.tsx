"use client"

import React, {useState, useEffect} from "react"
import {motion, AnimatePresence} from "framer-motion"
import {Calendar, GitCommit, Star, TrendingUp, BookOpen, Tag, Clock, Sparkles} from "lucide-react"
import {format, parseISO} from "date-fns"
import {zhCN} from "date-fns/locale"
import {useTheme} from "next-themes"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs"
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from "@/components/ui/tooltip"
import {ScrollArea} from "@/components/ui/scroll-area"
import Link from "next/link"
import FloatingMenu from "@/components/menu/FloatingMenu";

interface BlogPost {
    type?: "article" | "statusUpdate"
    title: string
    shortUrl: string
    category: string
    createdAt: string
}

interface YearData {
    articleCount: number
    statusUpdateCount: number
    articles: BlogPost[]
    statusUpdates: BlogPost[]
}

interface BlogData {
    [year: string]: YearData
}

interface ContributionDay {
    date: string
    count: number
    level: 0 | 1 | 2 | 3 | 4 // 0 = no contributions, 4 = most contributions
    posts: BlogPost[]
}

interface WeekData {
    days: (ContributionDay | null)[]
    weekIndex: number
}

interface ContributionGraphProps {
    data?: BlogData
    year?: string
    title?: string
    quote?: QuoteProps
}

// Add the QuoteProps interface after the ContributionGraphProps interface
interface QuoteProps {
    quote?: string
    author?: string
    authorTitle?: string
    authorAvatar?: string
}

type ColorScheme = "green" | "blue" | "purple" | "pink" | "orange"

// Update the function parameters to include the quote prop
export default function BlogActivityGraph({
                                              data: providedData,
                                              year = new Date().getFullYear().toString(),
                                              title = "更新频率",
                                              quote,
                                          }: ContributionGraphProps) {
    const {resolvedTheme} = useTheme()
    const [colorScheme, setColorScheme] = useState<ColorScheme>("green")
    const [weeks, setWeeks] = useState<WeekData[]>([])
    const [totalContributions, setTotalContributions] = useState(0)
    const [maxContributions, setMaxContributions] = useState(0)
    const [longestStreak, setLongestStreak] = useState(0)
    const [currentStreak, setCurrentStreak] = useState(0)
    const [isLoading, setIsLoading] = useState(true)
    const [selectedYear, setSelectedYear] = useState(year)
    const [availableYears, setAvailableYears] = useState<string[]>([])
    const [selectedPost, setSelectedPost] = useState<BlogPost | null>(null)
    const [categoryStats, setCategoryStats] = useState<{ [key: string]: number }>({})

    const handleColorSchemeChange = (value: string) => {
        setColorScheme(value as ColorScheme);
    };

    // 生成月份（缩写）
    const months = ["1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"]

    // 计算月份位置
    const [monthLabels, setMonthLabels] = useState<{ name: string; index: number }[]>([])

    // 颜色方案
    const colorSchemes = {
        green: {
            light: ["#e6f7ec", "#bae7cc", "#5cc98c", "#2da160", "#216e46"],
            dark: ["#0e4429", "#006d32", "#26a641", "#39d353", "#56d364"],
        },
        blue: {
            light: ["#e6f1ff", "#b6d9ff", "#79b8ff", "#2188ff", "#0366d6"],
            dark: ["#0d419d", "#0969da", "#39d0ff", "#7ce3ff", "#a5d6ff"],
        },
        purple: {
            light: ["#f5f0ff", "#e2d9f3", "#d2bef9", "#a371f7", "#8957e5"],
            dark: ["#4c2889", "#6e40c9", "#8957e5", "#bf8cff", "#d2a8ff"],
        },
        pink: {
            light: ["#ffeef8", "#fedbeb", "#f9b3dd", "#ec6cb9", "#bf3989"],
            dark: ["#78184a", "#ae4c82", "#d76aaa", "#ff8ed2", "#ffb3d8"],
        },
        orange: {
            light: ["#fff8f2", "#ffebda", "#ffcb9a", "#f7934e", "#e16f24"],
            dark: ["#762c00", "#c65d21", "#f0883e", "#ffbe7d", "#ffd8b5"],
        },
    }

    const isDark = resolvedTheme === "dark"
    const selectedColors = isDark ? colorSchemes[colorScheme].dark : colorSchemes[colorScheme].light

    useEffect(() => {
        if (providedData) {
            // 提取可用年份
            const years = Object.keys(providedData).sort((a, b) => Number.parseInt(b) - Number.parseInt(a))
            setAvailableYears(years)

            if (years.length > 0 && !years.includes(selectedYear)) {
                setSelectedYear(years[0])
            } else {
                processData(providedData, selectedYear)
            }
        }
    }, [providedData, selectedYear])

    useEffect(() => {
        if (providedData && selectedYear) {
            processData(providedData, selectedYear)
        }
    }, [selectedYear, resolvedTheme])

    const processData = (blogData: BlogData, year: string) => {
        setIsLoading(true)

        if (!blogData[year]) {
            setWeeks([])
            setTotalContributions(0)
            setIsLoading(false)
            return
        }

        const yearData = blogData[year]
        for (const postItem in yearData.articles) {
            yearData.articles[postItem].type = "article"
        }
        for (const postItem in yearData.statusUpdates) {
            yearData.statusUpdates[postItem].type = "statusUpdate"
        }
        const allPosts = [...yearData.articles, ...yearData.statusUpdates]

        // 按日期分组帖子
        const postsByDate: { [date: string]: BlogPost[] } = {}
        const categories: { [category: string]: number } = {}

        allPosts.forEach((post) => {
            const date =
                post.createdAt.split(" ")[0] === post.createdAt ? post.createdAt.split("T")[0] : post.createdAt.split(" ")[0]
            if (!postsByDate[date]) {
                postsByDate[date] = []
            }
            postsByDate[date].push(post)

            // 统计分类
            if (!categories[post.category]) {
                categories[post.category] = 0
            }
            categories[post.category]++
        })

        setCategoryStats(categories)

        // 创建贡献数据
        const contributionData: ContributionDay[] = []
        let totalCount = 0
        let maxCount = 0

        // 获取年份的开始日期
        const startDate = new Date(`${year}-01-01`)

        // 获取结束日期（如果是当前年份，则为今天；否则为年末）
        const today = new Date()
        const isCurrentYear = Number.parseInt(year) === today.getFullYear()
        const endDate = isCurrentYear ? today : new Date(`${year}-12-31`)

        // 为每一天创建数据
        const currentDate = new Date(startDate)
        while (currentDate <= endDate) {
            const dateStr = currentDate.toISOString().split("T")[0]
            const posts = postsByDate[dateStr] || []
            const count = posts.length

            totalCount += count
            maxCount = Math.max(maxCount, count)

            // 先不设置级别，后面会根据最大值动态调整
            contributionData.push({
                date: dateStr,
                count,
                level: 0, // 临时值
                posts,
            })

            // 前进一天
            currentDate.setDate(currentDate.getDate() + 1)
        }

        // 根据最大贡献数动态设置级别
        contributionData.forEach((day) => {
            if (day.count === 0) {
                day.level = 0
            } else {
                // 计算相对级别：1 到 4 之间
                // 如果最大值是 1，那么 1 就是最深色 (4 级)
                // 如果最大值是 5，那么 1 是 1 级，2 是 2 级，以此类推
                const relativeLevel = Math.ceil((day.count / maxCount) * 4)
                day.level = relativeLevel as 0 | 1 | 2 | 3 | 4
            }
        })

        setTotalContributions(totalCount)
        setMaxContributions(maxCount)
        organizeDataIntoWeeks(contributionData)
        calculateStats(contributionData)

        setTimeout(() => {
            setIsLoading(false)
        }, 500)
    }

    // 计算统计数据
    const calculateStats = (contributionData: ContributionDay[]) => {
        if (!contributionData.length) return

        // 按日期排序
        const sortedData = [...contributionData].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())

        // 计算最长连续更新天数
        let currentStreak = 0
        let maxStreak = 0
        let tempStreak = 0

        // 从最近的日期开始计算当前连续天数
        for (let i = sortedData.length - 1; i >= 0; i--) {
            if (sortedData[i].count > 0) {
                currentStreak++
            } else {
                break
            }
        }

        // 计算最长连续天数
        for (let i = 0; i < sortedData.length; i++) {
            if (sortedData[i].count > 0) {
                tempStreak++
                maxStreak = Math.max(maxStreak, tempStreak)
            } else {
                tempStreak = 0
            }
        }

        setLongestStreak(maxStreak)
        setCurrentStreak(currentStreak)
    }

    // 将数据组织成周格式
    const organizeDataIntoWeeks = (contributionData: ContributionDay[]) => {
        if (!contributionData.length) return

        // 按日期排序
        const sortedData = [...contributionData].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())

        const firstDate = new Date(sortedData[0].date)
        // 调整到上一个周日
        const startDate = new Date(firstDate)
        startDate.setDate(firstDate.getDate() - firstDate.getDay())

        const weeksData: WeekData[] = []
        const monthPositions: { name: string; index: number }[] = []

        const currentDate = new Date(startDate)
        let weekIndex = 0
        let currentWeek: (ContributionDay | null)[] = []

        // 记录第一个月
        let currentMonth = currentDate.getMonth()
        monthPositions.push({
            name: months[currentMonth],
            index: 0,
        })

        // 获取最后一天的日期
        const lastDate = new Date(sortedData[sortedData.length - 1].date)
        // 调整到下一个周六
        const endDate = new Date(lastDate)
        const daysToAdd = 6 - endDate.getDay()
        endDate.setDate(endDate.getDate() + daysToAdd)

        // 生成周数据直到结束日期
        while (currentDate <= endDate) {
            currentWeek = []

            for (let d = 0; d < 7; d++) {
                // 检查月份变化
                if (d === 0 && currentDate.getMonth() !== currentMonth) {
                    currentMonth = currentDate.getMonth()
                    monthPositions.push({
                        name: months[currentMonth],
                        index: weekIndex,
                    })
                }

                const dateStr = currentDate.toISOString().split("T")[0]
                const dayData = sortedData.find((d) => d.date === dateStr)

                if (dayData) {
                    currentWeek.push(dayData)
                } else {
                    // 如果没有数据，填充 null
                    currentWeek.push(null)
                }

                // 前进一天
                currentDate.setDate(currentDate.getDate() + 1)
            }

            weeksData.push({
                days: currentWeek,
                weekIndex: weekIndex,
            })

            weekIndex++
        }

        setWeeks(weeksData)
        setMonthLabels(monthPositions)
    }

    // 格式化日期
    const formatDate = (dateString: string) => {
        try {
            const date = parseISO(dateString)
            return format(date, "yyyy年MM月dd日 EEEE", {locale: zhCN})
        } catch (e) {
            console.error("Date formatting error:", e)
            return dateString
        }
    }

    // 格式化时间
    const formatTime = (dateString: string) => {
        try {
            const date = parseISO(dateString)
            return format(date, "HH:mm", {locale: zhCN})
        } catch (e) {
            console.error("Time formatting error:", e)
            return ""
        }
    }

    // 加载动画变体
    const containerVariants = {
        hidden: {opacity: 0},
        visible: {
            opacity: 1,
            transition: {
                staggerChildren: 0.01,
                delayChildren: 0.2,
            },
        },
    }

    const cellVariants = {
        hidden: {scale: 0.5, opacity: 0},
        visible: {
            scale: 1,
            opacity: 1,
            transition: {type: "spring", stiffness: 300, damping: 20},
        },
    }

    const cardVariants = {
        hidden: {opacity: 0, y: 20},
        visible: {
            opacity: 1,
            y: 0,
            transition: {type: "spring", stiffness: 300, damping: 30},
        },
        exit: {
            opacity: 0,
            y: 20,
            transition: {duration: 0.2},
        },
    }

    // 计算贡献级别的颜色
    const getLevelColor = (level: number) => {
        if (level === 0) return resolvedTheme === "dark" ? "#1f2937" : "#f3f4f6"
        return selectedColors[level - 1]
    }

    // 计算贡献级别的阴影
    const getLevelShadow = (level: number) => {
        if (level === 0) return "none"
        return resolvedTheme === "dark" ? `0 0 8px ${selectedColors[level - 1]}50` : "none"
    }

    // 动态生成类名，避免字符串拼接问题
    const getColorClass = (prefix: string) => {
        const colorMap = {
            green: {
                bg: "bg-green-500",
                bgHover: "hover:bg-green-600",
                bgActive: "data-[state=active]:bg-green-500/20",
                text: "text-green-400",
                textHover: "hover:text-green-500",
                textActive: "data-[state=active]:text-green-400",
                gradientFrom: "from-green-500/20",
                gradientTo: "to-green-500/5",
            },
            blue: {
                bg: "bg-blue-500",
                bgHover: "hover:bg-blue-600",
                bgActive: "data-[state=active]:bg-blue-500/20",
                text: "text-blue-400",
                textHover: "hover:text-blue-500",
                textActive: "data-[state=active]:text-blue-400",
                gradientFrom: "from-blue-500/20",
                gradientTo: "to-blue-500/5",
            },
            purple: {
                bg: "bg-purple-500",
                bgHover: "hover:bg-purple-600",
                bgActive: "data-[state=active]:bg-purple-500/20",
                text: "text-purple-400",
                textHover: "hover:text-purple-500",
                textActive: "data-[state=active]:text-purple-400",
                gradientFrom: "from-purple-500/20",
                gradientTo: "to-purple-500/5",
            },
            pink: {
                bg: "bg-pink-500",
                bgHover: "hover:bg-pink-600",
                bgActive: "data-[state=active]:bg-pink-500/20",
                text: "text-pink-400",
                textHover: "hover:text-pink-500",
                textActive: "data-[state=active]:text-pink-400",
                gradientFrom: "from-pink-500/20",
                gradientTo: "to-pink-500/5",
            },
            orange: {
                bg: "bg-orange-500",
                bgHover: "hover:bg-orange-600",
                bgActive: "data-[state=active]:bg-orange-500/20",
                text: "text-orange-400",
                textHover: "hover:text-orange-500",
                textActive: "data-[state=active]:text-orange-400",
                gradientFrom: "from-orange-500/20",
                gradientTo: "to-orange-500/5",
            },
        }

        return colorMap[colorScheme as keyof typeof colorMap][prefix as keyof (typeof colorMap)[typeof colorScheme]]
    }

    // Add the AuthorQuote component after the getColorClass function and before the return statement
    const AuthorQuote = ({quote, author, authorTitle, authorAvatar}: QuoteProps) => {
        if (!quote) return null

        return (
            <div
                className="mb-6 rounded-xl p-5 border relative overflow-hidden bg-white/80 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/30 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-md transition-all duration-300 hover:shadow-lg">
                <div className={`absolute top-0 left-0 w-1 h-full ${getColorClass("bg")}`}></div>

                <div className="flex flex-col md:flex-row gap-4 items-start md:items-center">
                    {authorAvatar && (
                        <div className="flex-shrink-0">
                            <div
                                className="w-16 h-16 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-md">
                                <img
                                    src={authorAvatar || "/placeholder.svg"}
                                    alt={author || "作者"}
                                    className="w-full h-full object-cover"
                                />
                            </div>
                        </div>
                    )}

                    <div className="flex-1">
                        <blockquote
                            className="text-lg md:text-xl italic font-medium text-gray-700 dark:text-gray-300 mb-2">
                            <span className={`text-3xl ${getColorClass("text")} mr-1`}>&#34;</span>
                            {quote}
                            <span className={`text-3xl ${getColorClass("text")} ml-1`}>&#34;</span>
                        </blockquote>

                        {(author || authorTitle) && (
                            <div className="flex items-center">
                                <div className={`w-8 h-px ${getColorClass("bg")} mr-3 opacity-50`}></div>
                                <div>
                                    {author && <p className="font-semibold">{author}</p>}
                                    {authorTitle &&
                                        <p className="text-sm text-gray-500 dark:text-gray-400">{authorTitle}</p>}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div
            className="w-full max-w-4xl rounded-xl p-6 border shadow-xl relative overflow-hidden transition-colors duration-300 bg-gradient-to-br from-white to-gray-50 border-gray-200 text-gray-900 dark:bg-gradient-to-br dark:from-gray-900 dark:to-gray-950 dark:border-gray-800 dark:text-white">
            {/* 装饰元素 */}
            <div
                className={`absolute top-0 right-0 w-64 h-64 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 bg-gradient-to-br ${getColorClass("gradientFrom")} ${getColorClass("gradientTo")}`}
            ></div>
            <div
                className={`absolute bottom-0 left-0 w-64 h-64 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2 bg-gradient-to-tr ${getColorClass("gradientFrom")} ${getColorClass("gradientTo")}`}
            ></div>

            {/* 作者引言 */}
            {quote && <AuthorQuote {...quote} />}

            {/* 标题和统计信息 */}
            <div className="relative z-10">
                <div className="flex justify-between items-center mb-6">
                    <div>
                        <div className="flex items-center gap-2 mb-1">
                            <h3 className="text-md font-bold flex items-center gap-2">
                                <Calendar className={`w-5 h-5 ${getColorClass("text")}`}/>
                                {title}
                            </h3>
                        </div>
                        <p className="text-sm text-gray-500 dark:text-gray-400"> 每一天的更新动态捏 </p>
                    </div>
                    <div
                        className="px-4 py-2 rounded-lg border bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/50 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-sm hover:shadow-md transition-shadow duration-300">
            <span
                className={`text-2xl font-bold bg-gradient-to-r ${getColorClass("text")} to-gray-900 dark:to-white bg-clip-text`}
            >
              {totalContributions}
            </span>
                        <span className="text-sm ml-2 text-gray-500 dark:text-gray-400"> 更新 </span>
                    </div>
                </div>

                {/* 年份选择器 */}
                {availableYears.length > 0 && (
                    <div className="mb-4">
                        <Tabs value={selectedYear} onValueChange={setSelectedYear} className="w-full">
                            <TabsList
                                className="bg-white/70 border-gray-200/70 backdrop-blur-sm border dark:bg-gray-800/50 dark:border-gray-700/50 p-1 shadow-inner">
                                {availableYears.map((year) => (
                                    <TabsTrigger
                                        key={year}
                                        value={year}
                                        className={`${getColorClass("bgActive")} ${getColorClass("textActive")} transition-all duration-300 data-[state=active]:shadow-md`}
                                    >
                                        {year} 年
                                    </TabsTrigger>
                                ))}
                            </TabsList>
                        </Tabs>
                    </div>
                )}

                {/* 统计卡片 */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                    <div
                        className="rounded-lg p-3 border flex items-center bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/30 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-sm hover:shadow-md transition-all duration-300 transform hover:-translate-y-1">
                        <div
                            className={`p-2 rounded-full mr-3 ${getColorClass("bg")} bg-opacity-20 dark:bg-opacity-30`}>
                            <TrendingUp className={`w-5 h-5 ${getColorClass("text")}`}/>
                        </div>
                        <div>
                            <p className="text-xs text-gray-500 dark:text-gray-400"> 最长连续更新 </p>
                            <p className="text-lg font-bold">{longestStreak} 天 </p>
                        </div>
                    </div>

                    <div
                        className="rounded-lg p-3 border flex items-center bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/30 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-sm hover:shadow-md transition-all duration-300 transform hover:-translate-y-1">
                        <div
                            className={`p-2 rounded-full mr-3 ${getColorClass("bg")} bg-opacity-20 dark:bg-opacity-30`}>
                            <GitCommit className={`w-5 h-5 ${getColorClass("text")}`}/>
                        </div>
                        <div>
                            <p className="text-xs text-gray-500 dark:text-gray-400"> 当前连续更新 </p>
                            <p className="text-lg font-bold">{currentStreak} 天 </p>
                        </div>
                    </div>

                    <div
                        className="rounded-lg p-3 border flex items-center bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/30 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-sm hover:shadow-md transition-all duration-300 transform hover:-translate-y-1">
                        <div
                            className={`p-2 rounded-full mr-3 ${getColorClass("bg")} bg-opacity-20 dark:bg-opacity-30`}>
                            <Star className={`w-5 h-5 ${getColorClass("text")}`}/>
                        </div>
                        <div>
                            <p className="text-xs text-gray-500 dark:text-gray-400"> 单日最多更新 </p>
                            <p className="text-lg font-bold">{maxContributions} 篇 </p>
                        </div>
                    </div>
                </div>

                {/* 分类统计 */}
                {Object.keys(categoryStats).length > 0 && (
                    <div
                        className="rounded-lg p-4 border mb-6 bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/30 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-sm hover:shadow-md transition-all duration-300">
                        <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
                            <Tag className={`w-4 h-4 ${getColorClass("text")}`}/>
                            分类统计
                        </h4>
                        <div className="flex flex-wrap gap-2">
                            {Object.entries(categoryStats).map(([category, count]) => (
                                <Badge
                                    key={category}
                                    variant={isDark ? "outline" : "secondary"}
                                    className={`bg-gray-100 dark:border-gray-700 dark:bg-gray-800/50 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors duration-200 cursor-default`}
                                >
                                    {category}: {count} 篇
                                </Badge>
                            ))}
                        </div>
                    </div>
                )}
            </div>

            {/* 热力图 */}
            <div
                className="rounded-xl p-4 border relative z-10 bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/20 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-md">
                {isLoading ? (
                    <div className="flex flex-col items-center justify-center py-12">
                        <div
                            className={`w-10 h-10 border-4 rounded-full animate-spin mb-4 border-gray-300 border-t-${colorScheme}-500 dark:border-gray-600 dark:border-t-${colorScheme}-400`}
                        ></div>
                        <p className="text-gray-500 dark:text-gray-400"> 加载中...</p>
                    </div>
                ) : (
                    <ScrollArea className="pb-4">
                        <div className="flex mt-2">
                            <div className="w-full">
                                <div className="flex mb-1 relative">
                                    {monthLabels.map((month, index) => (
                                        <span
                                            key={index}
                                            className="text-xs absolute text-gray-500 dark:text-gray-400"
                                            style={{left: `${(month.index / Math.max(weeks.length - 1, 1)) * 100}%`}}
                                        >
                      {month.name}
                    </span>
                                    ))}
                                </div>

                                <motion.div
                                    className="grid grid-rows-7 grid-flow-col gap-1 mt-6"
                                    variants={containerVariants}
                                    initial="hidden"
                                    animate="visible"
                                    style={{width: "fit-content"}}
                                >
                                    {weeks.map((week, weekIndex) =>
                                        week.days.map((day, dayIndex) => (
                                            <TooltipProvider key={`${weekIndex}-${dayIndex}`}>
                                                <Tooltip>
                                                    <TooltipTrigger asChild>
                                                        <motion.div
                                                            variants={cellVariants}
                                                            className={`w-3 h-3 rounded-sm transition-all duration-300 hover:scale-125 hover:z-10 cursor-pointer`}
                                                            style={{
                                                                backgroundColor:
                                                                    day && day.level > 0 ? getLevelColor(day.level) : isDark ? "#1f2937" : "#f3f4f6",
                                                                boxShadow: day && day.level > 0 ? getLevelShadow(day.level) : "none",
                                                                border:
                                                                    day && day.level === 0
                                                                        ? isDark
                                                                            ? "1px solid rgba(55, 65, 81, 0.5)"
                                                                            : "1px solid rgba(229, 231, 235, 1)"
                                                                        : "none",
                                                            }}
                                                            onClick={() => day && day.posts.length > 0 && setSelectedPost(day.posts[0])}
                                                        />
                                                    </TooltipTrigger>
                                                    {day && (
                                                        <TooltipContent
                                                            side="top"
                                                            className={`p-3 rounded-lg shadow-lg z-50 max-w-[280px] backdrop-blur-sm bg-white/90 border border-gray-200 text-gray-900 dark:bg-gray-800/90 dark:border dark:border-gray-700 dark:text-white`}
                                                            style={{
                                                                boxShadow:
                                                                    resolvedTheme === "dark"
                                                                        ? "0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.3)"
                                                                        : "0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.04)",
                                                            }}
                                                        >
                                                            <div className="text-sm">
                                                                <div
                                                                    className="font-bold mb-1">{formatDate(day.date)}</div>
                                                                <div
                                                                    className={`font-medium ${getColorClass("text")}`}>{day.count} 篇更新
                                                                </div>
                                                                {day.posts.length > 0 && (
                                                                    <div
                                                                        className="mt-2 max-h-[200px] overflow-y-auto pr-1 custom-scrollbar">
                                                                        {day.posts.map((post, i) => (
                                                                            <div
                                                                                key={i}
                                                                                className={`mb-2 p-2 rounded-md text-xs transition-colors bg-gray-100/80 hover:bg-gray-200/80 dark:bg-gray-700/50 dark:hover:bg-gray-700 transform hover:scale-102 transition-transform duration-200`}
                                                                                onClick={() => {
                                                                                    setSelectedPost(post)
                                                                                }}
                                                                            >
                                                                                <div
                                                                                    className="font-medium truncate">{post.title}</div>
                                                                                <div
                                                                                    className="flex justify-between mt-1">
                                                                                    <span
                                                                                        className="text-gray-500 dark:text-gray-400">{post.category}</span>
                                                                                    <span
                                                                                        className="text-gray-500 dark:text-gray-400">
                                            {formatTime(post.createdAt)}
                                          </span>
                                                                                </div>
                                                                            </div>
                                                                        ))}
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </TooltipContent>
                                                    )}
                                                </Tooltip>
                                            </TooltipProvider>
                                        )),
                                    )}
                                </motion.div>
                            </div>
                        </div>
                    </ScrollArea>
                )}

                {/* 图例 */}
                <div className="flex justify-end items-center mt-4">
                    <span className="text-xs mr-2 text-gray-400 dark:text-gray-500"> 更少 </span>
                    <div className="flex gap-1">
                        <div
                            className="w-3 h-3 rounded-sm bg-gray-100 border border-gray-200 dark:bg-gray-800 dark:border dark:border-gray-700/50"/>
                        {selectedColors.map((color, i) => (
                            <div key={i} className="w-3 h-3 rounded-sm" style={{backgroundColor: color}}/>
                        ))}
                    </div>
                    <span className="text-xs ml-2 text-gray-400 dark:text-gray-500"> 更多 </span>
                </div>
            </div>

            {/* 选中的文章详情 */}
            <AnimatePresence>
                {selectedPost && (
                    <motion.div
                        className={`mt-6 rounded-xl p-4 border relative z-10 bg-white/70 backdrop-blur-sm border-gray-200/70 dark:bg-gray-800/30 dark:backdrop-blur-sm dark:border-gray-700/50 shadow-lg ${getColorClass("gradientFrom")} bg-gradient-to-br to-transparent`}
                        variants={cardVariants}
                        initial="hidden"
                        animate="visible"
                        exit="exit"
                    >
                        <div className="flex justify-between items-start mb-2">
                            <h4 className="text-lg font-bold">{selectedPost.title}</h4>
                            <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => setSelectedPost(null)}
                                className="text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white rounded-full"
                            >
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    width="18"
                                    height="18"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                >
                                    <line x1="18" y1="6" x2="6" y2="18"></line>
                                    <line x1="6" y1="6" x2="18" y2="18"></line>
                                </svg>
                            </Button>
                        </div>

                        <div className="flex flex-wrap gap-2 mb-3">
                            <Badge
                                variant={isDark ? "outline" : "secondary"}
                                className={`bg-gray-100 dark:border-gray-700 dark:bg-gray-800/50 ${getColorClass("text")}`}
                            >
                                <Tag className="w-3 h-3 mr-1"/>
                                {selectedPost.category}
                            </Badge>
                            <Badge
                                variant={isDark ? "outline" : "secondary"}
                                className="bg-gray-100 dark:border-gray-700 dark:bg-gray-800/50"
                            >
                                <Clock className="w-3 h-3 mr-1"/>
                                {formatTime(selectedPost.createdAt)}
                            </Badge>
                        </div>

                        <div className="flex items-center gap-2 mt-4">
                            <Link
                                href={
                                    selectedPost.type === "article"
                                        ? `/posts/${selectedPost.shortUrl}`
                                        : `/moments${format(parseISO(selectedPost.createdAt), "/yyyy/MM/dd" + selectedPost.shortUrl)}`
                                }
                            >
                                <Button
                                    size="sm"
                                    className={`${getColorClass("bg")} ${getColorClass("bgHover")} text-white shadow-md hover:shadow-lg transition-all duration-300`}
                                >
                                    <BookOpen className="w-4 h-4 mr-2"/>
                                    阅读全文
                                </Button>
                            </Link>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* 底部信息 */}
            <div
                className="mt-4 text-xs text-center text-gray-400 dark:text-gray-500 flex items-center justify-center gap-1">
                <Sparkles className={`w-3 h-3 ${getColorClass("text")}`}/>
                数据更新于 {new Date().toLocaleDateString("zh-CN")} · 点击方块查看详情
            </div>
            <FloatingMenu items={[
                {
                    type: 'select',
                    value: colorScheme,
                    label: '颜色方案',
                    icon: <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                               xmlns="http://www.w3.org/2000/svg">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                              d="M5 13l4 4L19 7"/>
                    </svg>,
                    onChange: (e: string) => {
                        handleColorSchemeChange(e)
                    },
                    options: [
                        {label: "绿色", value: "green"},
                        {label: "蓝色", value: "blue"},
                        {label: "紫色", value: "purple"},
                        {label: "粉色", value: "pink"},
                        {label: "橙色", value: "orange"},
                    ],
                    placeholder: "选择颜色方案"
                }
            ]}/>
        </div>
    )
}

