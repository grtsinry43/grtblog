"use client"

import React, {useState, useEffect} from "react"
import {motion, AnimatePresence} from "framer-motion"
import SkeletonCard from "./SkeletonCard";
import RecommendedSection from "./RecommendedSection"
import {getRecommend} from "@/api/recommend"
import {useAppSelector} from "@/redux/hooks";

interface RecommendationItem {
    id: string
    title: string
    authorName: string
    tags: string
    shortUrl: string
    views: number
    cover?: string | null
}

const RecommendClient = () => {
    const [recommendations, setRecommendations] = useState<RecommendationItem[]>([])
    const [loading, setLoading] = useState(true)

    const {isLogin} = useAppSelector(state => state.user)

    useEffect(() => {
        const fetchRecommendations = async () => {
            try {
                const response = await getRecommend()
                setRecommendations(response)
            } catch (error) {
                console.log(error)
                setRecommendations([])
            } finally {
                setLoading(false)
            }
        }

        fetchRecommendations()
    }, [])

    return (
        <section className="py-8 md:py-12"
                 style={{
                     background: "linear-gradient(to bottom right, rgba(var(--primary),0.03), rgba(var(--background),0.2))"
                 }}>
            <div className="container mx-auto px-4">
                {/* 限制最大宽度，让内容更集中 */}
                <div className="max-w-7xl mx-auto">
                    <div className="mb-6 md:mb-8">
                        <h2 className="text-2xl md:text-3xl font-bold mb-3 md:mb-4 text-gray-800 dark:text-white">
                            🎯 为你推荐
                        </h2>
                        <div className="text-xs md:text-sm opacity-60 text-gray-600 dark:text-gray-400">
                            <p className="mb-1">
                                * 根据你的阅读行为智能推荐感兴趣的文章
                            </p>
                            <p className="text-xs">
                                {isLogin 
                                    ? "基于你的登录账户生成个性化推荐" 
                                    : "基于会话数据推荐，登录获取更精准推荐体验"
                                }
                            </p>
                        </div>
                    </div>
                    
                    <AnimatePresence mode="wait">
                        {loading ? (
                            <motion.div
                                key="skeleton"
                                initial={{opacity: 0}}
                                animate={{opacity: 1}}
                                exit={{opacity: 0}}
                                transition={{duration: 0.5}}
                            >
                                {/* 移动端骨架屏 */}
                                <div className="block md:hidden overflow-hidden -mx-4">
                                    <div className="flex gap-3 overflow-x-auto pb-4 px-4">
                                        {Array.from({length: 3}).map((_, index) => (
                                            <div key={index} className="flex-none w-60">
                                                <SkeletonCard isMobile={true}/>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                                
                                {/* 桌面端骨架屏 */}
                                <div className="hidden md:grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 lg:gap-5">
                                    {Array.from({length: 4}).map((_, index) => (
                                        <SkeletonCard key={index} isMobile={false}/>
                                    ))}
                                </div>
                            </motion.div>
                        ) : (
                            <motion.div 
                                key="content" 
                                initial={{opacity: 0}} 
                                animate={{opacity: 1}}
                                transition={{duration: 0.5}}
                            >
                                <RecommendedSection recommendations={recommendations}/>
                            </motion.div>
                        )}
                    </AnimatePresence>
                    
                    {!loading && recommendations.length === 0 && (
                        <motion.div
                            initial={{opacity: 0, y: 20}}
                            animate={{opacity: 1, y: 0}}
                            className="text-center py-12"
                        >
                            <div className="text-gray-400 dark:text-gray-500">
                                <div className="text-4xl mb-4">📝</div>
                                <p className="text-lg mb-2">暂时没有推荐内容</p>
                                <p className="text-sm opacity-70">多阅读一些文章，我们会为你推荐更多精彩内容</p>
                            </div>
                        </motion.div>
                    )}
                </div>
            </div>
        </section>
    )
}

export default RecommendClient

