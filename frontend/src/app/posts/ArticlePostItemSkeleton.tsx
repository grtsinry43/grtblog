import {HashtagIcon, TagIcon} from "@heroicons/react/24/outline"
import {Calendar, Eye, ThumbsUp} from "lucide-react"
import {AiOutlineComment} from "react-icons/ai"
import {motion} from "framer-motion"

const ArticlePageItemSkeleton = () => {
    return (
        <motion.div 
            className="relative py-8"
            initial={{opacity: 0}}
            animate={{opacity: 1}}
            transition={{duration: 0.3}}
        >
            {/* 标题骨架 - 使用更优雅的shimmer效果 */}
            <div className="mb-3">
                <motion.div 
                    className="h-6 bg-gradient-to-r from-gray-100 via-gray-200 to-gray-100 dark:from-gray-800 dark:via-gray-700 dark:to-gray-800 rounded-lg relative overflow-hidden"
                    style={{width: '75%'}}
                    animate={{
                        backgroundPosition: ['0% 50%', '100% 50%', '0% 50%']
                    }}
                    transition={{
                        duration: 2,
                        ease: "linear",
                        repeat: Infinity
                    }}
                >
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 dark:via-white/5 to-transparent transform -skew-x-12 animate-pulse" />
                </motion.div>
            </div>

            {/* 摘要骨架 */}
            <div className="mb-4 space-y-3">
                <motion.div 
                    className="h-4 bg-gradient-to-r from-gray-100 via-gray-200 to-gray-100 dark:from-gray-800 dark:via-gray-700 dark:to-gray-800 rounded-md relative overflow-hidden"
                    animate={{
                        backgroundPosition: ['0% 50%', '100% 50%', '0% 50%']
                    }}
                    transition={{
                        duration: 2,
                        ease: "linear",
                        repeat: Infinity,
                        delay: 0.1
                    }}
                >
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 dark:via-white/5 to-transparent transform -skew-x-12" />
                </motion.div>
                <motion.div 
                    className="h-4 bg-gradient-to-r from-gray-100 via-gray-200 to-gray-100 dark:from-gray-800 dark:via-gray-700 dark:to-gray-800 rounded-md relative overflow-hidden"
                    style={{width: '85%'}}
                    animate={{
                        backgroundPosition: ['0% 50%', '100% 50%', '0% 50%']
                    }}
                    transition={{
                        duration: 2,
                        ease: "linear",
                        repeat: Infinity,
                        delay: 0.2
                    }}
                >
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 dark:via-white/5 to-transparent transform -skew-x-12" />
                </motion.div>
            </div>

            {/* 元数据骨架 */}
            <div className="flex flex-wrap items-center gap-4">
                {/* 分类骨架 */}
                <motion.div 
                    className="flex items-center"
                    initial={{opacity: 0, y: 10}}
                    animate={{opacity: 1, y: 0}}
                    transition={{delay: 0.3}}
                >
                    <HashtagIcon className="w-3 h-3 mr-1 text-gray-400 dark:text-gray-500" />
                    <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full w-16 animate-pulse"></div>
                </motion.div>

                {/* 时间骨架 */}
                <motion.div 
                    className="flex items-center"
                    initial={{opacity: 0, y: 10}}
                    animate={{opacity: 1, y: 0}}
                    transition={{delay: 0.4}}
                >
                    <Calendar className="w-3 h-3 mr-1 text-gray-400 dark:text-gray-500" />
                    <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full w-20 animate-pulse"></div>
                </motion.div>

                {/* 标签骨架 */}
                <motion.div 
                    className="flex items-center"
                    initial={{opacity: 0, y: 10}}
                    animate={{opacity: 1, y: 0}}
                    transition={{delay: 0.5}}
                >
                    <TagIcon className="w-3 h-3 mr-1 text-gray-400 dark:text-gray-500" />
                    <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full w-12 animate-pulse"></div>
                </motion.div>

                {/* 统计信息骨架 */}
                <motion.div 
                    className="flex items-center gap-3 ml-auto"
                    initial={{opacity: 0, y: 10}}
                    animate={{opacity: 1, y: 0}}
                    transition={{delay: 0.6}}
                >
                    <div className="flex items-center">
                        <Eye className="w-3 h-3 mr-1 text-gray-400 dark:text-gray-500" />
                        <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full w-6 animate-pulse"></div>
                    </div>
                    <div className="flex items-center">
                        <AiOutlineComment className="w-3 h-3 mr-1 text-gray-400 dark:text-gray-500" />
                        <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full w-6 animate-pulse"></div>
                    </div>
                    <div className="flex items-center">
                        <ThumbsUp className="w-3 h-3 mr-1 text-gray-400 dark:text-gray-500" />
                        <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full w-6 animate-pulse"></div>
                    </div>
                </motion.div>
            </div>

            {/* 骨架屏底部装饰线 - 更subtle */}
            <motion.div 
                className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-gray-200/20 dark:via-gray-700/10 to-transparent"
                initial={{scaleX: 0}}
                animate={{scaleX: 1}}
                transition={{delay: 0.7, duration: 0.5}}
            />
        </motion.div>
    )
}

export default ArticlePageItemSkeleton

