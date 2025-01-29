import type React from "react"

const SkeletonCard: React.FC = () => (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
        <div className="h-48 bg-gray-300 dark:bg-gray-700 animate-pulse" />
        <div className="p-6">
            <div className="h-6 bg-gray-300 dark:bg-gray-700 rounded w-3/4 mb-2 animate-pulse" />
            <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2 mb-2 animate-pulse" />
            <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/4 animate-pulse" />
        </div>
    </div>
)

export default SkeletonCard

