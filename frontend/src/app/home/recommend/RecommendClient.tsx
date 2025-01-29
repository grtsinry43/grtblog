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
                console.error("Error fetching recommendations:", error)
            } finally {
                setLoading(false)
            }
        }

        fetchRecommendations()
    }, [])

    return (
        <section className="py-12 bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-850">
            <div className="container mx-auto px-4">
                <h2 className="text-3xl font-bold mb-8 text-gray-800 dark:text-white"> 为你推荐 </h2>
                <div className="text-sm opacity-50 mb-4">
                    * 当你在阅读文章时，本站会根据你的阅读行为分析你的兴趣爱好，为你推荐更多你可能感兴趣的文章。
                    <br/>
                    {
                        isLogin ? "当前推荐是根据你所登录的账户生成的" : "当前推荐是基于会话数据的，如果你清除了浏览器缓存，推荐数据会被清除。\n" +
                            "你可以登录以获取更好的推荐体验。"
                    }
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
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                                {Array.from({length: 3}).map((_, index) => (
                                    <SkeletonCard key={index}/>
                                ))}
                            </div>
                        </motion.div>
                    ) : (
                        <motion.div key="content" initial={{opacity: 0}} animate={{opacity: 1}}
                                    transition={{duration: 0.5}}>
                            <RecommendedSection recommendations={recommendations}/>
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>
        </section>
    )
}

export default RecommendClient

