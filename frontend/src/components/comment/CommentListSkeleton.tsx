"use client"
import {Avatar, Skeleton} from "@radix-ui/themes"
import {motion} from "framer-motion"

const CommentListSkeleton = () => {
    return (
        <ul className="space-y-6">
            {[...Array(3)].map((_, index) => (
                <motion.li
                    key={index}
                    initial={{opacity: 0, y: 20}}
                    animate={{opacity: 1, y: 0}}
                    transition={{type: "spring", stiffness: 100, delay: index * 0.1}}
                    
                    
                    className="bg-card rounded-lg p-4 shadow-sm"
                >
                    <div className="flex items-center space-x-2 mb-2">
                        <Skeleton>
                            <Avatar size="2" radius="full" fallback={""}/>
                        </Skeleton>
                        <Skeleton className="h-4 w-24"/>
                        <Skeleton className="h-4 w-16"/>
                    </div>
                    <Skeleton className="h-4 w-full mb-2"/>
                    <Skeleton className="h-4 w-3/4 mb-2"/>
                    <Skeleton className="h-4 w-1/2 mb-4"/>
                    <div className="flex justify-between items-center">
                        <Skeleton className="h-4 w-20"/>
                        <Skeleton className="h-8 w-16 rounded-md"/>
                    </div>
                    <div className="mt-4 ml-8 border-l-2 border-muted pl-4">
                        <div className="flex items-center space-x-2 mb-2">
                            <Skeleton>
                                <Avatar size="1" radius="full" fallback={""}/>
                            </Skeleton>
                            <Skeleton className="h-3 w-20"/>
                        </div>
                        <Skeleton className="h-3 w-full mb-1"/>
                        <Skeleton className="h-3 w-2/3"/>
                    </div>
                </motion.li>
            ))}
        </ul>
    )
}

export default CommentListSkeleton

