import type React from "react"
import {motion} from "framer-motion"
import RecommendationCard from "./RecommendationCard"

interface RecommendationItem {
    id: string
    title: string
    authorName: string
    shortUrl: string
    tags: string
    views: number
    cover?: string | null
}

interface RecommendedSectionProps {
    recommendations: RecommendationItem[]
}

const RecommendedSection: React.FC<RecommendedSectionProps> = ({recommendations}) => {
    return (
        <motion.div
            initial={{opacity: 0, y: 20}}
            animate={{opacity: 1, y: 0}}
            transition={{duration: 0.5}}
        >
            {/* 移动端：横向滚动布局 */}
            <div className="block md:hidden overflow-hidden -mx-4">
                <div 
                    className="flex gap-3 overflow-x-auto pb-4 px-4 scrollbar-hide"
                    style={{
                        scrollbarWidth: 'none',
                        msOverflowStyle: 'none',
                        WebkitOverflowScrolling: 'touch',
                    }}
                >
                    {recommendations?.map((item, index) => (
                        <motion.div
                            key={item.id}
                            className="flex-none w-60"
                            initial={{opacity: 0, x: 20}}
                            animate={{opacity: 1, x: 0}}
                            transition={{
                                type: "spring",
                                stiffness: 100,
                                damping: 10,
                                delay: index * 0.1,
                            }}
                        >
                            <RecommendationCard item={item} isMobile={true}/>
                        </motion.div>
                    ))}
                    {/* 移动端添加最后的间距 */}
                    <div className="w-4 flex-none" />
                </div>
            </div>

            {/* 桌面端：网格布局 */}
            <div className="hidden md:grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-4 gap-4 lg:gap-5">
                {recommendations?.map((item, index) => (
                    <motion.div
                        key={item.id}
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        transition={{
                            type: "spring",
                            stiffness: 100,
                            damping: 10,
                            delay: index * 0.1,
                            bounce: 0.2,
                        }}
                    >
                        <RecommendationCard item={item} isMobile={false}/>
                    </motion.div>
                ))}
            </div>
        </motion.div>
    )
}

export default RecommendedSection

