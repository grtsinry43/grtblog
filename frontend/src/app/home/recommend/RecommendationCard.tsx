import type React from "react"
import {motion} from "framer-motion"
import {TagIcon, UserIcon, EyeIcon} from "lucide-react"
import Link from "next/link";

interface RecommendationItem {
    id: string
    title: string
    authorName: string
    shortUrl: string
    tags: string
    views: number
    cover?: string | null
}

const RecommendationCard: React.FC<{ item: RecommendationItem }> = ({item}) => (

    <Link href={`/posts/${item.shortUrl}`} passHref>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden transition-all duration-300
            ease-in-out hover:shadow-xl hover:scale-105">
            <motion.div
                whileHover={{y: -5}}
            >
                <div className="relative h-48 bg-gray-200 dark:bg-gray-700">
                    {item.cover ? (
                        <img src={item.cover} alt={item.title}
                             className="w-full h-full object-cover"/>
                    ) : (
                        <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
                            还没有封面捏
                        </div>
                    )}
                </div>
                <div className="p-6">
                    <h3 className="text-xl font-semibold mb-2 text-gray-800 dark:text-white line-clamp-2">{item.title}</h3>
                    <div className="flex items-center mb-2 text-gray-600 dark:text-gray-300">
                        <UserIcon className="w-4 h-4 mr-2"/>
                        <span className="text-sm">{item.authorName}</span>
                    </div>
                    <div className="flex items-center mb-2 text-gray-600 dark:text-gray-300">
                        <TagIcon className="w-4 h-4 mr-2"/>
                        <span className="text-sm">{item.tags.split(",")[0]}</span>
                    </div>
                    <div className="flex items-center text-gray-600 dark:text-gray-300">
                        <EyeIcon className="w-4 h-4 mr-2"/>
                        <span className="text-sm">{item.views} views</span>
                    </div>
                </div>
            </motion.div>
        </div>
    </Link>
)

export default RecommendationCard
