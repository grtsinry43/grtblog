import type React from "react"
import {motion} from "framer-motion"
import {TagIcon, UserIcon, EyeIcon} from "lucide-react"
import Link from "next/link";
import {clsx} from "clsx";
import {article_font} from "@/app/fonts/font";

interface RecommendationItem {
    id: string
    title: string
    authorName: string
    shortUrl: string
    tags: string
    views: number
    cover?: string | null
}

interface RecommendationCardProps {
    item: RecommendationItem
    isMobile?: boolean
}

const RecommendationCard: React.FC<RecommendationCardProps> = ({item, isMobile = false}) => {
    
    // 根据访问量生成热度等级
    const getHeatLevel = (views: number) => {
        if (views > 1000) return { level: 'hot', color: 'text-red-500', bg: 'bg-red-50 dark:bg-red-950/30' }
        if (views > 500) return { level: 'trending', color: 'text-orange-500', bg: 'bg-orange-50 dark:bg-orange-950/30' }
        if (views > 100) return { level: 'popular', color: 'text-blue-500', bg: 'bg-blue-50 dark:bg-blue-950/30' }
        return { level: 'new', color: 'text-green-500', bg: 'bg-green-50 dark:bg-green-950/30' }
    }

    const heatLevel = getHeatLevel(item.views)
    
    return (
        <Link href={`/posts/${item.shortUrl}`} passHref className="block">
            <motion.div
                whileHover={{
                    y: -4,
                    transition: { duration: 0.2, ease: "easeOut" }
                }}
                className="group h-full"
            >
                <div className={clsx(
                    "relative overflow-hidden bg-white dark:bg-gray-900/50 backdrop-blur-sm",
                    "border border-gray-200/60 dark:border-gray-700/60",
                    "transition-all duration-300 ease-out",
                    "hover:border-gray-300/80 dark:hover:border-gray-600/80",
                    "hover:shadow-lg hover:shadow-gray-200/50 dark:hover:shadow-gray-900/50",
                    "w-full max-w-sm mx-auto",
                    isMobile ? "h-72 rounded-lg" : "h-80 rounded-xl"
                )}>
                    
                    {/* 顶部装饰条 */}
                    <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                    
                    {/* 图片区域 */}
                    <div className="relative overflow-hidden">
                        <div className={clsx(
                            "relative bg-gray-50 dark:bg-gray-800",
                            isMobile ? "h-36" : "h-40"
                        )}>
                            {item.cover ? (
                                <>
                                    <img 
                                        src={item.cover} 
                                        alt={item.title}
                                        className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                                    />
                                    {/* 渐变遮罩 */}
                                    <div className="absolute inset-0 bg-gradient-to-t from-black/20 via-transparent to-transparent" />
                                </>
                            ) : (
                                <div className="flex items-center justify-center h-full bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-800 dark:to-gray-700">
                                    <div className="text-center text-gray-400 dark:text-gray-500">
                                        <div className="w-10 h-10 mx-auto mb-2 rounded-lg bg-gray-300/50 dark:bg-gray-600/50 flex items-center justify-center">
                                            <TagIcon className="w-5 h-5" />
                                        </div>
                                        <span className="text-xs font-medium opacity-75">精选内容</span>
                                    </div>
                                </div>
                            )}
                            
                            {/* 热度标签 */}
                            <div className={clsx(
                                "absolute top-3 right-3 px-2.5 py-1 rounded-md text-xs font-medium",
                                "bg-white/90 dark:bg-gray-900/90 backdrop-blur-sm",
                                "border border-gray-200/50 dark:border-gray-700/50",
                                heatLevel.color
                            )}>
                                {item.views}
                            </div>
                        </div>
                    </div>
                    
                    {/* 内容区域 */}
                    <div className={clsx(
                        "p-4 h-full flex flex-col",
                        isMobile ? "pb-3" : "pb-4"
                    )}>
                        {/* 标题 */}
                        <h3 className={clsx(
                            article_font.className,
                            "font-semibold text-gray-900 dark:text-gray-100 line-clamp-2 mb-3",
                            "transition-colors duration-200 group-hover:text-blue-600 dark:group-hover:text-blue-400",
                            isMobile ? "text-sm leading-tight" : "text-base leading-snug"
                        )}>
                            {item.title}
                        </h3>
                        
                        {/* 标签区域 */}
                        <div className="flex flex-wrap gap-1.5 mb-3">
                            {item.tags.split(",").slice(0, 2).map((tag, index) => (
                                <span
                                    key={index}
                                    className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 border border-gray-200/50 dark:border-gray-700/50"
                                >
                                    {tag.trim()}
                                </span>
                            ))}
                        </div>
                        
                        {/* 底部信息 */}
                        <div className="mt-auto space-y-2">
                            {/* 作者 */}
                            <div className="flex items-center text-gray-600 dark:text-gray-400 text-xs">
                                <div className="w-6 h-6 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mr-2 flex-shrink-0">
                                    <UserIcon className="w-3 h-3 text-white"/>
                                </div>
                                <span className="truncate font-medium">{item.authorName}</span>
                            </div>
                            
                            {/* 统计信息 */}
                            <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-500">
                                <div className="flex items-center">
                                    <EyeIcon className="w-3 h-3 mr-1"/>
                                    <span>{item.views} 阅读</span>
                                </div>
                                <div className={clsx(
                                    "px-2 py-0.5 rounded-md text-xs font-medium",
                                    heatLevel.bg,
                                    heatLevel.color
                                )}>
                                    {heatLevel.level === 'hot' && '热门'}
                                    {heatLevel.level === 'trending' && '趋势'}
                                    {heatLevel.level === 'popular' && '推荐'}
                                    {heatLevel.level === 'new' && '最新'}
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    {/* 底部装饰 */}
                    <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-gray-200 dark:via-gray-700 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                    
                    {/* 左侧装饰线 */}
                    <div className="absolute left-0 top-0 bottom-0 w-px bg-gradient-to-b from-blue-500 via-purple-500 to-pink-500 opacity-0 group-hover:opacity-60 transition-opacity duration-300" />
                </div>
            </motion.div>
        </Link>
    )
}

export default RecommendationCard
