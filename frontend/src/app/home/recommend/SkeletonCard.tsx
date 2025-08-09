import type React from "react"

interface SkeletonCardProps {
    isMobile?: boolean
}

const SkeletonCard: React.FC<SkeletonCardProps> = ({ isMobile = false }) => (
    <div className={`relative overflow-hidden bg-white dark:bg-gray-900/50 backdrop-blur-sm border border-gray-200/60 dark:border-gray-700/60 w-full max-w-sm mx-auto ${
        isMobile ? "h-72 rounded-lg" : "h-80 rounded-xl"
    }`}>
        {/* 图片骨架 */}
        <div className={`bg-gray-200 dark:bg-gray-800 animate-pulse ${
            isMobile ? "h-36" : "h-40"
        }`} />
        
        {/* 内容骨架 */}
        <div className={isMobile ? "p-3 pb-3" : "p-4 pb-4"}>
            {/* 标题骨架 */}
            <div className={`bg-gray-200 dark:bg-gray-700 rounded-lg animate-pulse mb-3 ${
                isMobile ? "h-4" : "h-5"
            }`} />
            <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-lg w-3/4 mb-3 animate-pulse" />
            
            {/* 标签骨架 */}
            <div className="flex gap-1.5 mb-3">
                <div className="h-5 w-12 bg-gray-200 dark:bg-gray-700 rounded-md animate-pulse" />
                <div className="h-5 w-16 bg-gray-200 dark:bg-gray-700 rounded-md animate-pulse" />
            </div>
            
            {/* 底部信息骨架 */}
            <div className="mt-auto space-y-2">
                {/* 作者骨架 */}
                <div className="flex items-center">
                    <div className="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-700 animate-pulse mr-2" />
                    <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-20 animate-pulse" />
                </div>
                
                {/* 统计信息骨架 */}
                <div className="flex items-center justify-between">
                    <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-16 animate-pulse" />
                    <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-12 animate-pulse" />
                </div>
            </div>
        </div>
    </div>
)

export default SkeletonCard

