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
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {recommendations.map((item, index) => (
                    <motion.div
                        key={item.id}
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        transition={{
                            type: "spring",
                            stiffness: 100,
                            damping: 10,
                            delay: index * 0.1,
                            bounce: 0.3,
                        }}
                    >
                        <RecommendationCard item={item}/>
                    </motion.div>
                ))}
            </div>
        </motion.div>
    )
}

export default RecommendedSection

